// Package authmgr provides simple interface for Google oauth2 authentication
// for console applications.
package authmgr

import (
	"reflect"
	"testing"

	"golang.org/x/oauth2"
)

func TestManager_clientIDhash(t *testing.T) {
	type fields struct {
		config *oauth2.Config
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
				config: tt.fields.config,
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
