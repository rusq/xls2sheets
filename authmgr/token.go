package authmgr

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/shibukawa/configdir"
	"golang.org/x/oauth2"
)

const (
	vendor      = "rusq"
	application = "sheets-refresh"

	tokFile = application + "-token.bin"
)

// template filenames
const (
	tmLogin    = "login.html"
	tmCallback = "callback.html"
	tmIndex    = "index.html"
)

const oauthStateString = "01224277302367423221"

// config directories
var configDirs = configdir.New(vendor, application)

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

	cache := configDirs.QueryCacheFolder()
	if err := cache.MkdirAll(); err != nil {
		log.Fatalf("unable to create cache directory structure")
	}
	return filepath.Join(cache.Path, path)
}

func (Manager) path(filename string) string {
	configDirs.LocalPath, _ = filepath.Abs(".")
	folder := configDirs.QueryFolderContainsFile(filename)
	if folder != nil {
		return filepath.Join(folder.Path, filename)
	}
	// check the existance in cache folder
	cache := configDirs.QueryCacheFolder()
	if cache.Exists(tokFile) {
		return filepath.Join(cache.Path, tokFile)
	}
	return ""
}

// RemoveToken finds and removes tokenFile from cache folder.  If the token
// file is not present it does nothing.
func RemoveToken() error {
	tokenPath := getTokenPath(tokFile)
	if tokenPath == "" {
		return nil
	}
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// getTokenPath returns the path for the tokenFile.  If file not found
// returns an empty string.
func getTokenPath(tokenFile string) string {
	// check the file locally and in user/system configuration folders
	configDirs.LocalPath, _ = filepath.Abs(".")
	folder := configDirs.QueryFolderContainsFile(tokenFile)
	if folder != nil {
		return filepath.Join(folder.Path, tokenFile)
	}
	// check the existance in cache folder
	cache := configDirs.QueryCacheFolder()
	if cache.Exists(tokFile) {
		return filepath.Join(cache.Path, tokFile)
	}
	return ""
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
		fullPath = m.createPath(tokFile)
	}

	log.Printf("Saving token file to: %s", fullPath)
	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
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
		return nil, fmt.Errorf("Unable to read authorization code: %v", err)
	}

	tok, err := m.Config().Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
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
	clean := make(chan struct{}, 1)
	go func() {
		errC <- srv.ListenAndServe()
		close(clean)
	}()
	fmt.Printf("callback server listening on %s\n", m.listenerAddr)

	fmt.Println("Please follow the Instructions in your browser to authorize sheets-refresh.")
	if err := openbrowser("http://" + m.listenerAddr); err != nil {
		fmt.Printf("If your browser does not open automatically, please click here to authenticate google sheets:\n%s\n", m.listenerAddr)
	}

	select {
	case token := <-tokenChan:
		srv.Close()
		<-clean
		return token, nil
	case err := <-errC:
		if err != nil {
			return nil, err
		}
	}
	return nil, errors.New("this should not happen")
}

func (m *Manager) authHandler(tokenChan chan<- *oauth2.Token) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, index)
	})
	mux.HandleFunc("/login", m.loginHandler)
	mux.HandleFunc("/callback", m.tokenCallbackHandler(tokenChan))

	return mux
}

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	authURL := m.Config().AuthCodeURL(oauthStateString)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (m *Manager) tokenCallbackHandler(tokenChan chan<- *oauth2.Token) http.HandlerFunc {
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
		w.WriteHeader(200)
		w.Write([]byte(success))
		select {
		case tokenChan <- token:
		default:
			http.Error(w, "failed to return the token from callback", http.StatusInternalServerError)
		}
	}
}

func openbrowser(url string) (err error) {

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return
}
