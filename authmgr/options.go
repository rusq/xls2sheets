package authmgr

// Option sets option variables.
type Option func(*Manager)

// OptTemplateDir sets the template directory for templates.
func OptTemplateDir(dir string) Option {
	return func(m *Manager) {
		// TODO: actually use this info.
		m.templateDir = dir
	}
}

// OptListenerAddr sets the template directory for templates.
func OptListenerAddr(addr string) Option {
	return func(m *Manager) {
		m.listenerAddr = addr
	}
}

// OptTryWebAuth sets the flag to attempt to present user with browser
// for authentication.  Otherwise - console login is used.
func OptTryWebAuth(b bool, redirectURL string) Option {
	return func(m *Manager) {
		m.tryWebAuth = b
		m.redirectURL = redirectURL
	}
}

func OptAppName(vendor, name string) Option {
	return func(m *Manager) {
		m.vendor = vendor
		m.appname = name
	}
}

func OptUseIndexPage(b bool) Option {
	return func(m *Manager) {
		m.useIndexPage = b
	}
}
