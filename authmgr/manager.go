// Package authmgr provides simple interface for Google oauth2 authentication
// for console applications.
package authmgr

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	maxCredFileSz = 1048576 // 1MB for credentials file is more than enough

	listenerHost = "localhost"
	listenerPort = "6061" //  to avoid collision with godoc etc.
)

type Manager struct {
	token     *oauth2.Token
	config    *oauth2.Config
	tokenFile string

	reqFunc tokenReqFunc
	// options
	redirectURL  string
	templateDir  string
	listenerAddr string
	tryWebAuth   bool

	tokenChan chan *oauth2.Token // for web request

	vendor  string
	appName string
}

type tokenReqFunc func() (*oauth2.Token, error)

// apply applies specified options
func (m *Manager) apply(opts ...Option) *Manager {
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// New creates a new instance of Manager from oauth.Config
func New(config *oauth2.Config, opts ...Option) *Manager {

	m := &Manager{config: config}
	m.apply(opts...)

	// populate some parameters
	if m.listenerAddr == "" {
		m.listenerAddr = fmt.Sprintf("%s:%s", listenerHost, listenerPort)
	}

	m.setWebAuth(m.tryWebAuth, m.redirectURL)
	return m
}

func (m *Manager) setWebAuth(enabled bool, redirectURL string) {
	if enabled {
		m.reqFunc = m.browserTokenRequest
		if redirectURL == "" {
			m.config.RedirectURL = fmt.Sprintf("http://%s/callback", m.listenerAddr)
		} else {
			m.config.RedirectURL = redirectURL
		}
	} else {
		m.reqFunc = m.cliTokenRequest
	}
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
	// check if the permissions on the file are set correctly
	permissions := fi.Mode().Perm()
	if !(permissions == 0600 || permissions == 0400) {
		return nil, fmt.Errorf("credentials file is to permissive (%o), "+
			"to fix - run:\n\tchmod 600 %s", permissions, filename)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %s", err)
	}
	return New(config, opts...), nil
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
	return New(config, opts...), nil
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
	// try from disk
	token, err := m.loadToken(m.path(tokFile))
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
