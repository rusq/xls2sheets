package authmgr

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		{"listener set", args{"new"}, &Manager{}, &Manager{opts: options{listenerAddr: ""}}, false},
		{"empty listener", args{""}, &Manager{}, &Manager{opts: options{listenerAddr: listenerHost + ":" + listenerPort}}, false},
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
			if diff := cmp.Diff(tt.args.addr, m.opts.listenerAddr); diff != "" {
				t.Errorf("OptListenerAddr() fail, (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestOptTryWebAuth(t *testing.T) {
	type args struct {
		b               bool
		rootPath        string
		redirectURLbase string
	}
	tests := []struct {
		name    string
		args    args
		before  *Manager
		after   *Manager
		wantErr bool
	}{
		{"t, set",
			args{true, "", "blah"},
			&Manager{opts: options{tryWebAuth: false, webRootPath: "", redirectURLBase: ""}},
			&Manager{opts: options{tryWebAuth: true, webRootPath: "", redirectURLBase: "blah"}},
			false,
		},
		{"t, unset",
			args{true, "/kek", ""},
			&Manager{opts: options{tryWebAuth: false, webRootPath: "", redirectURLBase: ""}},
			&Manager{opts: options{tryWebAuth: true, webRootPath: "/kek", redirectURLBase: ""}},
			false,
		},
		{"f, set",
			args{false, "", "lol"},
			&Manager{opts: options{tryWebAuth: false, webRootPath: "", redirectURLBase: ""}},
			&Manager{opts: options{tryWebAuth: false, webRootPath: "", redirectURLBase: "lol"}},
			false,
		},
		{"f, unset",
			args{false, "", ""},
			&Manager{opts: options{tryWebAuth: false, webRootPath: "", redirectURLBase: ""}},
			&Manager{opts: options{tryWebAuth: false, webRootPath: "", redirectURLBase: ""}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptTryWebAuth(tt.args.b, tt.args.rootPath, tt.args.redirectURLbase)
			m := tt.before
			err := got(m)
			if diff := cmp.Diff(tt.after, m, cmp.AllowUnexported(Manager{}, options{})); diff != "" {
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
		{"1", args{"vendor", "appname"}, &Manager{}, &Manager{opts: options{vendor: "vendor", appname: "appname"}}, false},
		{"empty", args{"", ""}, &Manager{}, &Manager{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := OptAppName(tt.args.vendor, tt.args.name)
			m := tt.before
			err := opt(m)
			if diff := cmp.Diff(tt.after, m, cmp.AllowUnexported(Manager{}, options{})); diff != "" {
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
		{"t", args{true}, &Manager{}, &Manager{opts: options{useIndexPage: true}}, false},
		{"f", args{false}, &Manager{opts: options{useIndexPage: true}}, &Manager{opts: options{useIndexPage: false}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptUseIndexPage(tt.args.b)
			m := tt.before
			err := got(m)
			if diff := cmp.Diff(tt.after, m, cmp.AllowUnexported(Manager{}, options{})); diff != "" {
				t.Errorf("OptUseIndexPage() fail, (-want,+got):\n%s", diff)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("OptUseIndexPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
