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
	"runtime"

	"github.com/shibukawa/configdir"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	maxCredFileSz = 131072 // 128KB for credentials file is more than enough

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

	tokenFile string
	configDir configdir.ConfigDir

	// options
	redirectURL  string
	templateDir  string
	listenerAddr string
	tryWebAuth   bool
	useIndexPage bool

	vendor  string
	appname string
}

type tokenReqFunc func() (*oauth2.Token, error)

// apply applies specified options
func (m *Manager) apply(opts ...Option) (*Manager, error) {
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}
	m.setBrowserAuth(m.tryWebAuth, m.listenerAddr, m.redirectURL)
	m.setAppName()

	return m, nil
}

// New creates a new instance of Manager from oauth.Config
func New(config *oauth2.Config, opts ...Option) (*Manager, error) {

	m := &Manager{config: config}
	if _, err := m.apply(opts...); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) setAppName() {
	if m.vendor == "" {
		m.vendor = defVendor
	}
	if m.appname == "" {
		m.appname = defAppPrefix + m.clientIDhash()
	}
	m.configDir = configdir.New(m.vendor, m.appname)
}

// NewFromGoogleCreds creates manager from a credentials file
func NewFromGoogleCreds(filename string, scopes []string, opts ...Option) (*Manager, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if fi.Size() == 0 || fi.Size() > maxCredFileSz { //1 MB
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
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %s", err)
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
func (m *Manager) setBrowserAuth(enabled bool, listenerAddr, redirectURL string) {
	if !enabled {
		// terminal prompt
		m.reqFunc = m.cliTokenRequest
		return
	}
	// browser token request
	m.reqFunc = m.browserTokenRequest

	if listenerAddr == "" {
		m.listenerAddr = fmt.Sprintf("%s:%s", listenerHost, listenerPort)
	}
	if redirectURL == "" {
		m.config.RedirectURL = fmt.Sprintf("http://%s%s", m.listenerAddr, callbackPath)
	} else {
		m.config.RedirectURL = redirectURL
	}
}

func (m *Manager) clientIDhash() string {
	h := sha1.New()
	io.WriteString(h, m.config.ClientID)
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
	token, err := m.loadToken(m.tokenName())
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
