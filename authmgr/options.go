package authmgr

import (
	"fmt"
	"html/template"
	"os"
	"path"
)

// Option sets option variables.
type Option func(*Manager) error

// OptTemplateDir sets the template directory for templates (and loads the
// templates).
func OptTemplateDir(dir string) Option {
	return func(m *Manager) error {
		// TODO: actually use this info.
		fi, err := os.Stat(dir)
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return fmt.Errorf("not a directory: %s", dir)
		}
		for _, filename := range []string{tmCallback, tmIndex} {
			fi, err = os.Stat(filename)
			if err != nil {
				return err
			}
			if !fi.Mode().IsRegular() {
				return fmt.Errorf("not a regular file: %s", filename)
			}

		}
		tmpl, err = template.ParseFiles(path.Join(dir, tmCallback), path.Join(dir, tmIndex))
		if err != nil {
			return err
		}
		m.opts.templateDir = dir
		return nil
	}
}

// OptListenerAddr sets the template directory for templates.
func OptListenerAddr(addr string) Option {
	return func(m *Manager) error {
		m.opts.listenerAddr = addr
		return nil
	}
}

// OptTryWebAuth sets the flag to attempt to present user with browser for
// authentication.  Otherwise - console login is used.
func OptTryWebAuth(b bool, rootPath string, redirectURLBase string) Option {
	return func(m *Manager) error {
		m.opts.webRootPath = rootPath
		m.opts.tryWebAuth = b
		m.opts.redirectURLBase = redirectURLBase
		return nil
	}
}

// OptAppName sets the application name.
func OptAppName(vendor, name string) Option {
	return func(m *Manager) error {
		m.opts.vendor = vendor
		m.opts.appname = name
		return nil
	}
}

// OptUseIndexPage sets the optional index page display prior to redirecting
// to Oauth provider scopes permission page.
func OptUseIndexPage(b bool) Option {
	return func(m *Manager) error {
		m.opts.useIndexPage = b
		return nil
	}
}

// OptResetAuth resets the token and forces reauthentication.
func OptResetAuth() Option {
	return func(m *Manager) error {
		return m.removeToken()
	}
}
