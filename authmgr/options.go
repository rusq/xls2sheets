package authmgr

type Option func(*Manager)

// OptTemplateDir sets the template directory for templates.
func OptTemplateDir(dir string) Option {
	return func(m *Manager) {
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
	}
}
