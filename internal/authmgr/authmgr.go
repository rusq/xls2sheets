// Package authmgr provides simple interface for Google oauth2 authentication
// for console applications.
package authmgr

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	maxCredFileSz = 32768 // 32KiB for credentials file is more than enough

	listenerHost = "localhost"
	listenerPort = "6061" //  to avoid collision with godoc etc.
)
const (
	defVendor    = "rusq"
	defApp       = "authmgr"
	defAppPrefix = "auth-"
)

// Manager is authorization manager
type Manager struct {
	token  *oauth2.Token
	config *oauth2.Config

	reqFunc tokenReqFunc

	cacheDir string

	opts options
}

type options struct {
	webRootPath     string
	redirectURLBase string
	templateDir     string
	listenerAddr    string
	tryWebAuth      bool
	useIndexPage    bool

	vendor  string
	appname string
}

type tokenReqFunc func() (*oauth2.Token, error)

// applyOpts applies specified options
func applyOpts(m *Manager, opts ...Option) (*Manager, error) {
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}
	m.setBrowserAuth(m.opts.tryWebAuth, m.opts.listenerAddr, m.opts.redirectURLBase)

	return m, nil
}

// New creates a new instance of Manager from oauth.Config
func New(config *oauth2.Config, opts ...Option) (*Manager, error) {
	m, err := applyOpts(&Manager{config: config}, opts...)
	if err != nil {
		return nil, err
	}
	if m.opts.vendor == "" {
		m.opts.vendor = defVendor
	}
	if m.opts.appname == "" {
		m.opts.appname = defAppPrefix + m.clientIDhash()
	}
	ucd, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	m.cacheDir = filepath.Join(ucd, m.opts.vendor, m.opts.appname)
	// ensure dir exists
	if err := os.MkdirAll(m.cacheDir, 0700); err != nil {
		return nil, err
	}

	return m, nil
}

// NewFromGoogleCreds creates manager from a credentials file
func NewFromGoogleCreds(filename string, scopes []string, opts ...Option) (*Manager, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if fi.Size() == 0 || fi.Size() > maxCredFileSz {
		return nil, fmt.Errorf("suspicious file size: %d", fi.Size())
	}

	// check permissions if not on windows.
	if runtime.GOOS != "windows" {
		// check if the permissions on the file are set correctly
		permissions := fi.Mode().Perm()
		if !(permissions == 0600 || permissions == 0400) {
			return nil, fmt.Errorf("credentials file is to permissive (%o), "+
				"to fix - run:\n\tchmod 600 %s", permissions, filename)
		}
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read the client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse the client secret file: %s", err)
	}
	return New(config, opts...)
}

// NewFromEnv creates manager from environment variables.
func NewFromEnv(idKey, secretKey string, scopes []string, opts ...Option) (*Manager, error) {
	id := os.Getenv(idKey)
	secret := os.Getenv(secretKey)
	if id == "" || secret == "" {
		return nil, fmt.Errorf("environment variables %q and/or %q are not set", idKey, secretKey)
	}
	config := &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
	return New(config, opts...)
}

// setBrowserAuth sets the required variables for web auth.
func (m *Manager) setBrowserAuth(enabled bool, listenerAddr, redirectURLBase string) {
	if !enabled {
		// terminal prompt
		m.reqFunc = m.cliTokenRequest
		return
	}
	// browser token request
	m.reqFunc = m.browserTokenRequest

	// set request parameters
	if m.opts.webRootPath == "" {
		m.opts.webRootPath = basepath
	} else {
		if m.opts.webRootPath[len(m.opts.webRootPath)-1] != '/' {
			m.opts.webRootPath = m.opts.webRootPath + "/"
		}
	}
	if listenerAddr == "" {
		m.opts.listenerAddr = fmt.Sprintf("%s:%s", listenerHost, listenerPort)
	}
	if redirectURLBase == "" {
		m.config.RedirectURL = fmt.Sprintf("http://%s%s", m.opts.listenerAddr, m.callbackPath())
	} else {
		m.config.RedirectURL = path.Join(redirectURLBase, m.callbackPath())
	}
}

func (m *Manager) clientIDhash() string {
	h := sha1.New()
	_, err := io.WriteString(h, m.config.ClientID)
	if err != nil {
		panic("clientIDhash: " + err.Error())
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Client returns authenticated client.
func (m *Manager) Client() (*http.Client, error) {
	tok, err := m.Token()
	if err != nil {
		return nil, err
	}
	return m.Config().Client(context.Background(), tok), nil
}

// Token return oauth2 token.
func (m *Manager) Token() (*oauth2.Token, error) {
	if m.token != nil {
		return m.token, nil
	}
	// try to load from disk
	token, err := m.loadToken(filepath.Join(m.cacheDir, m.tokenName()))
	if err != nil {
		// try to auth
		token, err = m.reqFunc()
		if err != nil {
			return nil, err
		}
		if err := m.saveToken(token); err != nil {
			return nil, err
		}
	}
	m.token = token
	return token, nil
}

// Config returns oauth2 config.
func (m *Manager) Config() *oauth2.Config {
	return m.config
}

// OpenBrowser attempts to open browser.
func OpenBrowser(url string) (err error) {
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
