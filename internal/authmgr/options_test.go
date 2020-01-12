package authmgr

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shibukawa/configdir"
	"golang.org/x/oauth2"
)

func TestOptTemplateDir(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OptTemplateDir(tt.args.dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OptTemplateDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptListenerAddr(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name    string
		args    args
		before  *Manager
		after   *Manager
		wantErr bool
	}{
		{"listener set", args{"new"}, &Manager{}, &Manager{listenerAddr: ""}, false},
		{"empty listener", args{""}, &Manager{}, &Manager{listenerAddr: listenerHost + ":" + listenerPort}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptListenerAddr(tt.args.addr)
			m := &Manager{}
			err := got(m)
			if (err != nil) != tt.wantErr {
				t.Errorf("OptListenerAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.args.addr, m.listenerAddr); diff != "" {
				t.Errorf("OptListenerAddr() fail, (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestOptTryWebAuth(t *testing.T) {
	type args struct {
		b           bool
		redirectURL string
	}
	tests := []struct {
		name    string
		args    args
		before  *Manager
		after   *Manager
		wantErr bool
	}{
		{"t, set",
			args{true, "blah"},
			&Manager{tryWebAuth: false, redirectURL: ""},
			&Manager{tryWebAuth: true, redirectURL: "blah"},
			false,
		},
		{"t, unset",
			args{true, ""},
			&Manager{tryWebAuth: false, redirectURL: ""},
			&Manager{tryWebAuth: true, redirectURL: ""},
			false,
		},
		{"f, set",
			args{false, "lol"},
			&Manager{tryWebAuth: false, redirectURL: ""},
			&Manager{tryWebAuth: false, redirectURL: "lol"},
			false,
		},
		{"f, unset",
			args{false, ""},
			&Manager{tryWebAuth: false, redirectURL: ""},
			&Manager{tryWebAuth: false, redirectURL: ""},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptTryWebAuth(tt.args.b, tt.args.redirectURL)
			m := tt.before
			err := got(m)
			if diff := cmp.Diff(tt.after, m, cmp.AllowUnexported(Manager{})); diff != "" {
				t.Errorf("OptTryWebAuth() fail, (-want,+got):\n%s", diff)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OptTryWebAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestOptAppName(t *testing.T) {
	type args struct {
		vendor string
		name   string
	}
	tests := []struct {
		name    string
		args    args
		before  *Manager
		after   *Manager
		wantErr bool
	}{
		{"1", args{"vendor", "appname"}, &Manager{}, &Manager{vendor: "vendor", appname: "appname", configDir: configdir.New("vendor", "appname")}, false},
		{"empty", args{"", ""}, &Manager{config: &oauth2.Config{ClientID: "blah"}}, &Manager{config: &oauth2.Config{ClientID: "blah"}, vendor: defVendor, appname: defAppPrefix + "5bf1fd927dfb8679496a2e6cf00cbe50c1c87145", configDir: configdir.New(defVendor, defAppPrefix+"5bf1fd927dfb8679496a2e6cf00cbe50c1c87145")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptAppName(tt.args.vendor, tt.args.name)
			m := tt.before
			err := got(m)
			if diff := cmp.Diff(tt.after, m, cmp.AllowUnexported(Manager{})); diff != "" {
				t.Errorf("OptAppName() fail, (-want,+got):\n%s", diff)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OptAppName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestOptUseIndexPage(t *testing.T) {
	type args struct {
		b bool
	}
	tests := []struct {
		name    string
		args    args
		before  *Manager
		after   *Manager
		wantErr bool
	}{
		{"t", args{true}, &Manager{}, &Manager{useIndexPage: true}, false},
		{"f", args{false}, &Manager{useIndexPage: true}, &Manager{useIndexPage: false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptUseIndexPage(tt.args.b)
			m := tt.before
			err := got(m)
			if diff := cmp.Diff(tt.after, m, cmp.AllowUnexported(Manager{})); diff != "" {
				t.Errorf("OptUseIndexPage() fail, (-want,+got):\n%s", diff)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OptUseIndexPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}