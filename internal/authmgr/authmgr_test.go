// Package authmgr provides simple interface for Google oauth2 authentication
// for console applications.
package authmgr

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/shibukawa/configdir"
	"golang.org/x/oauth2"
)

func TestManager_clientIDhash(t *testing.T) {
	type fields struct {
		token        *oauth2.Token
		config       *oauth2.Config
		reqFunc      tokenReqFunc
		tokenFile    string
		configDir    configdir.ConfigDir
		redirectURL  string
		templateDir  string
		listenerAddr string
		tryWebAuth   bool
		useIndexPage bool
		vendor       string
		appname      string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"test", fields{config: &oauth2.Config{ClientID: "test"}}, "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				token:        tt.fields.token,
				config:       tt.fields.config,
				reqFunc:      tt.fields.reqFunc,
				tokenFile:    tt.fields.tokenFile,
				configDir:    tt.fields.configDir,
				redirectURL:  tt.fields.redirectURL,
				templateDir:  tt.fields.templateDir,
				listenerAddr: tt.fields.listenerAddr,
				tryWebAuth:   tt.fields.tryWebAuth,
				useIndexPage: tt.fields.useIndexPage,
				vendor:       tt.fields.vendor,
				appname:      tt.fields.appname,
			}
			if got := m.clientIDhash(); got != tt.want {
				t.Errorf("Manager.clientIDhash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFromGoogleCreds(t *testing.T) {
	type args struct {
		filename string
		scopes   []string
		opts     []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Manager
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromGoogleCreds(tt.args.filename, tt.args.scopes, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromGoogleCreds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromGoogleCreds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFromEnv(t *testing.T) {
	type args struct {
		idKey     string
		secretKey string
		scopes    []string
		opts      []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Manager
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromEnv(tt.args.idKey, tt.args.secretKey, tt.args.scopes, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_setBrowserAuth(t *testing.T) {
	type fields struct {
		token        *oauth2.Token
		config       *oauth2.Config
		reqFunc      tokenReqFunc
		tokenFile    string
		configDir    configdir.ConfigDir
		redirectURL  string
		templateDir  string
		listenerAddr string
		tryWebAuth   bool
		useIndexPage bool
		vendor       string
		appname      string
	}
	type args struct {
		enabled      bool
		listenerAddr string
		redirectURL  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				token:        tt.fields.token,
				config:       tt.fields.config,
				reqFunc:      tt.fields.reqFunc,
				tokenFile:    tt.fields.tokenFile,
				configDir:    tt.fields.configDir,
				redirectURL:  tt.fields.redirectURL,
				templateDir:  tt.fields.templateDir,
				listenerAddr: tt.fields.listenerAddr,
				tryWebAuth:   tt.fields.tryWebAuth,
				useIndexPage: tt.fields.useIndexPage,
				vendor:       tt.fields.vendor,
				appname:      tt.fields.appname,
			}
			m.setBrowserAuth(tt.args.enabled, tt.args.listenerAddr, tt.args.redirectURL)
		})
	}
}

func TestManager_Client(t *testing.T) {
	type fields struct {
		token        *oauth2.Token
		config       *oauth2.Config
		reqFunc      tokenReqFunc
		tokenFile    string
		configDir    configdir.ConfigDir
		redirectURL  string
		templateDir  string
		listenerAddr string
		tryWebAuth   bool
		useIndexPage bool
		vendor       string
		appname      string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *http.Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				token:        tt.fields.token,
				config:       tt.fields.config,
				reqFunc:      tt.fields.reqFunc,
				tokenFile:    tt.fields.tokenFile,
				configDir:    tt.fields.configDir,
				redirectURL:  tt.fields.redirectURL,
				templateDir:  tt.fields.templateDir,
				listenerAddr: tt.fields.listenerAddr,
				tryWebAuth:   tt.fields.tryWebAuth,
				useIndexPage: tt.fields.useIndexPage,
				vendor:       tt.fields.vendor,
				appname:      tt.fields.appname,
			}
			got, err := m.Client()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Client() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Manager.Client() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_Token(t *testing.T) {
	type fields struct {
		token        *oauth2.Token
		config       *oauth2.Config
		reqFunc      tokenReqFunc
		tokenFile    string
		configDir    configdir.ConfigDir
		redirectURL  string
		templateDir  string
		listenerAddr string
		tryWebAuth   bool
		useIndexPage bool
		vendor       string
		appname      string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *oauth2.Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				token:        tt.fields.token,
				config:       tt.fields.config,
				reqFunc:      tt.fields.reqFunc,
				tokenFile:    tt.fields.tokenFile,
				configDir:    tt.fields.configDir,
				redirectURL:  tt.fields.redirectURL,
				templateDir:  tt.fields.templateDir,
				listenerAddr: tt.fields.listenerAddr,
				tryWebAuth:   tt.fields.tryWebAuth,
				useIndexPage: tt.fields.useIndexPage,
				vendor:       tt.fields.vendor,
				appname:      tt.fields.appname,
			}
			got, err := m.Token()
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Manager.Token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_Config(t *testing.T) {
	type fields struct {
		token        *oauth2.Token
		config       *oauth2.Config
		reqFunc      tokenReqFunc
		tokenFile    string
		configDir    configdir.ConfigDir
		redirectURL  string
		templateDir  string
		listenerAddr string
		tryWebAuth   bool
		useIndexPage bool
		vendor       string
		appname      string
	}
	tests := []struct {
		name   string
		fields fields
		want   *oauth2.Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				token:        tt.fields.token,
				config:       tt.fields.config,
				reqFunc:      tt.fields.reqFunc,
				tokenFile:    tt.fields.tokenFile,
				configDir:    tt.fields.configDir,
				redirectURL:  tt.fields.redirectURL,
				templateDir:  tt.fields.templateDir,
				listenerAddr: tt.fields.listenerAddr,
				tryWebAuth:   tt.fields.tryWebAuth,
				useIndexPage: tt.fields.useIndexPage,
				vendor:       tt.fields.vendor,
				appname:      tt.fields.appname,
			}
			if got := m.Config(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Manager.Config() = %v, want %v", got, tt.want)
			}
		})
	}
}
