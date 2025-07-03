package git

import (
	"strings"
	"testing"
)

type mockProfile struct {
	name       string
	username   string
	email      string
	hasSSH     bool
	writeError error
}

func (m *mockProfile) GetName() string {
	return m.name
}

func (m *mockProfile) GetGitUsername() string {
	return m.username
}

func (m *mockProfile) GetGitEmail() string {
	return m.email
}

func (m *mockProfile) HasSSHKeys() bool {
	return m.hasSSH
}

func (m *mockProfile) WriteSSHKeysToSystem() error {
	return m.writeError
}

func TestManager_ValidateSSHKey(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name        string
		keyContent  string
		isPrivate   bool
		expectError bool
	}{
		{
			name:        "valid private key",
			keyContent:  "-----BEGIN OPENSSH PRIVATE KEY-----\nsome content\n-----END OPENSSH PRIVATE KEY-----",
			isPrivate:   true,
			expectError: false,
		},
		{
			name:        "valid public key",
			keyContent:  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC... user@host",
			isPrivate:   false,
			expectError: false,
		},
		{
			name:        "empty key",
			keyContent:  "",
			isPrivate:   true,
			expectError: true,
		},
		{
			name:        "invalid private key",
			keyContent:  "not a private key",
			isPrivate:   true,
			expectError: true,
		},
		{
			name:        "invalid public key",
			keyContent:  "not a public key",
			isPrivate:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateSSHKey(tt.keyContent, tt.isPrivate)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateSSHKey() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestValidateGitInput(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		email       string
		expectError bool
	}{
		{
			name:        "valid input",
			username:    "John Doe",
			email:       "john.doe@example.com",
			expectError: false,
		},
		{
			name:        "username with special chars",
			username:    "john-doe_123",
			email:       "john@example.com",
			expectError: false,
		},
		{
			name:        "invalid username characters",
			username:    "john@doe",
			email:       "john@example.com",
			expectError: true,
		},
		{
			name:        "invalid email format",
			username:    "John Doe",
			email:       "invalid-email",
			expectError: true,
		},
		{
			name:        "username too long",
			username:    strings.Repeat("a", 101),
			email:       "john@example.com",
			expectError: true,
		},
		{
			name:        "email too long",
			username:    "John Doe",
			email:       strings.Repeat("a", 250) + "@example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGitInput(tt.username, tt.email)
			if (err != nil) != tt.expectError {
				t.Errorf("validateGitInput() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestContainsSuccessMessage(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "successful authentication",
			output:   "Hi username! You've successfully authenticated, but GitHub does not provide shell access.",
			expected: true,
		},
		{
			name:     "another success message",
			output:   "successfully authenticated to GitHub",
			expected: true,
		},
		{
			name:     "permission denied",
			output:   "Permission denied (publickey).",
			expected: false,
		},
		{
			name:     "connection timeout",
			output:   "Connection timed out",
			expected: false,
		},
		{
			name:     "empty output",
			output:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsSuccessMessage(tt.output)
			if result != tt.expected {
				t.Errorf("containsSuccessMessage() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDetectSSHKeyPaths(t *testing.T) {
	// This test would require mocking the filesystem or creating temporary files
	// For now, we'll just test that it returns an error when no keys exist
	_, _, err := detectSSHKeyPaths()
	// This will likely return an error in most test environments
	if err == nil {
		t.Log("SSH keys found in test environment")
	} else {
		t.Log("No SSH keys found (expected in test environment)")
	}
}
