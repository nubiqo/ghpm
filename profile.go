package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Profile represents a GitHub profile configuration
type Profile struct {
	Name          string `json:"name"`
	GitUsername   string `json:"git_username"`
	GitEmail      string `json:"git_email"`
	SSHPrivateKey string `json:"ssh_private_key"`
	SSHPublicKey  string `json:"ssh_public_key"`
	IsActive      bool   `json:"is_active"`
}

// Validate checks if the profile has valid data
func (p *Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if p.GitUsername == "" {
		return fmt.Errorf("git username cannot be empty")
	}
	if p.GitEmail == "" {
		return fmt.Errorf("git email cannot be empty")
	}

	// Check SSH keys if provided
	if p.SSHPrivateKey != "" {
		expandedPath := os.ExpandEnv(p.SSHPrivateKey)
		if _, err := os.Stat(expandedPath); err != nil {
			return fmt.Errorf("SSH private key not found: %s", p.SSHPrivateKey)
		}
	}

	if p.SSHPublicKey != "" {
		expandedPath := os.ExpandEnv(p.SSHPublicKey)
		if _, err := os.Stat(expandedPath); err != nil {
			return fmt.Errorf("SSH public key not found: %s", p.SSHPublicKey)
		}
	}

	return nil
}

// GetSSHKeyPaths returns the expanded SSH key paths
func (p *Profile) GetSSHKeyPaths() (privateKey, publicKey string) {
	if p.SSHPrivateKey != "" {
		privateKey = os.ExpandEnv(p.SSHPrivateKey)
	}
	if p.SSHPublicKey != "" {
		publicKey = os.ExpandEnv(p.SSHPublicKey)
	}
	return
}

// GetBackupDir returns the backup directory for this profile
func (p *Profile) GetBackupDir() string {
	configDir := os.ExpandEnv("$HOME/.config/github-profile-manager")
	return filepath.Join(configDir, "backups", p.Name)
}

// String returns a string representation of the profile
func (p *Profile) String() string {
	active := ""
	if p.IsActive {
		active = " [ACTIVE]"
	}
	return fmt.Sprintf("%s <%s>%s", p.GitUsername, p.GitEmail, active)
}
