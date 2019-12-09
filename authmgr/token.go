package authmgr

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
)

// template and webserver paths
const (
	tmCallback = "callback.html"
	tmIndex    = "index.html"

	basepath     = "/"
	loginPath    = basepath + "login"
	callbackPath = basepath + "callback"
)

var oauthStateString = randString(16)

//
// PATH functions
//

// createPath creates the path to token and returns the full path to
// tokenFile including tokenfilename.  I.e. on mac:
//    /Users/Youruser/Library/Caches/rusq/sheets-refresh/token.json
func (m *Manager) createPath(path string) string {
	tokenPath := m.path(path)
	if tokenPath != "" {
		// do nothing if the path exists
		return tokenPath
	}

	cache := m.configDir.QueryCacheFolder()
	if err := cache.MkdirAll(); err != nil {
		log.Fatalf("unable to create cache directory structure")
	}
	return filepath.Join(cache.Path, path)
}

func (m *Manager) path(filename string) string {
	m.configDir.LocalPath, _ = filepath.Abs(".")
	folder := m.configDir.QueryFolderContainsFile(filename)
	if folder != nil {
		return filepath.Join(folder.Path, filename)
	}
	// check the existance in cache folder
	cache := m.configDir.QueryCacheFolder()
	if cache.Exists(m.tokenName()) {
		return filepath.Join(cache.Path, m.tokenName())
	}
	return ""
}

// RemoveToken finds and removes tokenFile from cache folder.  If the token
// file is not present it does nothing.
func (m *Manager) RemoveToken() error {
	tokenPath := m.path(m.tokenName())
	if tokenPath == "" {
		return nil
	}
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

//
// Token request and manipulation
//

// loadToken creates a new token manager from token file.
func (m *Manager) loadToken(filename string) (*oauth2.Token, error) {
	// get the token from local storage
	tokenPath := m.path(filename)
	if tokenPath == "" {
		return nil, fmt.Errorf("not found: %s", filename)
	}

	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	if err := gob.NewDecoder(f).Decode(token); err != nil {
		return nil, err
	}

	return token, nil
}

// saveToken the token to file.
func (m *Manager) saveToken(token *oauth2.Token) error {
	var fullPath = m.tokenFile
	if fullPath == "" {
		fullPath = m.createPath(m.tokenName())
	}

	log.Printf("Saving token file to: %s", fullPath)
	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	m.tokenFile = fullPath
	return gob.NewEncoder(f).Encode(token)
}

func (m *Manager) cliTokenRequest() (*oauth2.Token, error) {
	authURL := m.Config().AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n"+
		"Enter authorization code: ", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	tok, err := m.Config().Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}

	return tok, nil
}

func (m *Manager) browserTokenRequest() (*oauth2.Token, error) {
	tokenChan := make(chan *oauth2.Token)
	srv := http.Server{
		Addr:    m.listenerAddr,
		Handler: m.authHandler(tokenChan),
	}

	errC := make(chan error, 1)
	isShutdown := make(chan struct{}, 1)
	go func() {
		errC <- srv.ListenAndServe()
		close(isShutdown)
	}()
	log.Printf("callback server listening on %s\n", m.listenerAddr)

	fmt.Printf("Please follow the Instructions in your browser to authorize %s\n"+
		"or press [Ctrl]+[C] to cancel...\n", m.appname)
	if err := openBrowser("http://" + m.listenerAddr); err != nil {
		fmt.Printf("If your browser does not open automatically, please open"+
			" this link to authenticate google sheets:\n%s\n", m.listenerAddr)
	}

	var token *oauth2.Token
	select {
	case err := <-errC:
		if err != nil {
			return nil, err
		}
	case token = <-tokenChan:
	}
	// once the token is received, shutdown the server.
	srv.Close()
	<-isShutdown
	return token, nil
}

// authHandler registers authentication handling routes.
func (m *Manager) authHandler(tokenChan chan<- *oauth2.Token) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", m.rootHandler)
	mux.HandleFunc(loginPath, m.loginHandler)
	mux.HandleFunc(callbackPath, m.createCallbackHandler(tokenChan))

	return mux
}

type appInfo struct {
	AppName string
}

func (m *Manager) rootHandler(w http.ResponseWriter, r *http.Request) {
	if !m.useIndexPage {
		http.Redirect(w, r, loginPath, http.StatusTemporaryRedirect)
		return
	}
	if err := tmpl.ExecuteTemplate(w, tmIndex, appInfo{m.appname}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	authURL := m.Config().AuthCodeURL(oauthStateString)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (m *Manager) createCallbackHandler(tokenChan chan<- *oauth2.Token) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		if state != oauthStateString {
			log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// code exchange
		code := r.FormValue("code")
		token, err := m.Config().Exchange(context.Background(), code)
		if err != nil {
			fmt.Printf("Code exchange failed with '%s'\n", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// success page, rendering just before shutting down the whole thing.
		if err := tmpl.ExecuteTemplate(w, tmCallback, appInfo{m.appname}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		select {
		case tokenChan <- token:
		default:
			http.Error(w, "failed to return the token from callback", http.StatusInternalServerError)
		}
	}
}

// openBrowser attempts to open browser.
func openBrowser(url string) (err error) {
	switch runtime.GOOS {
	default:
		err = fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return
}

func (m *Manager) tokenName() string {
	return "auth-token.bin"
}
