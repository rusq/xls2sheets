// Package authmgr provides simple interface for Google oauth2 authentication
// for console applications.
package authmgr

import (
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
