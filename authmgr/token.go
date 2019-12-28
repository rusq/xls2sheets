package authmgr

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/oauth2"
)

// template and webserver paths
const (
	tmCallback = "callback.html"
	tmIndex    = "index.html"

	basepath  = "/"
	pLogin    = "login"
	pCallback = "callback"
)

type appInfoPage struct {
	AppName   string
	LoginPath string
}

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

// removeToken finds and removes tokenFile from cache folder.  If the token
// file is not present it does nothing.
func (m *Manager) removeToken() error {
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

// saveToken saves the token to file.
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

// cliTokenRequest does the auth exchange using current terminal.
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

// browserTokenRequest requests the token through the web.
func (m *Manager) browserTokenRequest() (*oauth2.Token, error) {
	tokenChan := make(chan *oauth2.Token)

	srv := http.Server{
		Addr:    m.opts.listenerAddr,
		Handler: m.Handlers(tokenChan),
	}

	errC := make(chan error, 1)
	isShutdown := make(chan struct{}, 1)
	go func() {
		errC <- srv.ListenAndServe()
		close(isShutdown)
	}()
	log.Printf("callback server listening on %s\n", m.opts.listenerAddr)

	fmt.Printf("Please follow the Instructions in your browser to authorize %s\n"+
		"or press [Ctrl]+[C] to cancel...\n", m.opts.appname)
	if err := OpenBrowser("http://" + m.opts.listenerAddr + basepath); err != nil {
		fmt.Printf("If your browser does not open automatically, please open"+
			" this link to authenticate google sheets:\n%s\n", m.opts.listenerAddr)
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

// Handlers registers authentication handling routes.
func (m *Manager) Handlers(tokenChan chan<- *oauth2.Token) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(m.opts.webRootPath, m.rootHandler)
	mux.HandleFunc(m.loginPath(), m.loginHandler)
	mux.HandleFunc(m.callbackPath(), m.createCallbackHandler(tokenChan))

	return mux
}

func (m *Manager) callbackPath() string {
	return path.Join(m.opts.webRootPath, pCallback) + "/"
}

func (m *Manager) loginPath() string {
	return path.Join(m.opts.webRootPath, pLogin) + "/"
}

func (m *Manager) rootHandler(w http.ResponseWriter, r *http.Request) {
	if !m.opts.useIndexPage {
		http.Redirect(w, r, pLogin, http.StatusTemporaryRedirect)
		return
	}
	if err := tmpl.ExecuteTemplate(w, tmIndex, appInfoPage{m.opts.appname, m.loginPath()}); err != nil {
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
			http.Redirect(w, r, basepath, http.StatusTemporaryRedirect)
			return
		}

		// code exchange
		code := r.FormValue("code")
		token, err := m.Config().Exchange(context.Background(), code)
		if err != nil {
			fmt.Printf("Code exchange failed with '%s'\n", err)
			http.Redirect(w, r, basepath, http.StatusTemporaryRedirect)
			return
		}

		// success page, rendering just before shutting down the whole thing.
		if err := tmpl.ExecuteTemplate(w, tmCallback, appInfoPage{AppName: m.opts.appname}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		select {
		case tokenChan <- token:
		default:
			http.Error(w, "failed to return the token from callback", http.StatusInternalServerError)
		}
	}
}

// tokenName returns the token name.
func (m *Manager) tokenName() string {
	return "auth-token.bin"
}
