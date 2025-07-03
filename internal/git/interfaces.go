package git

// ProfileInterface defines what git operations need from a profile
type ProfileInterface interface {
	GetName() string
	GetGitUsername() string
	GetGitEmail() string
	HasSSHKeys() bool
	WriteSSHKeysToSystem() error
}
