package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Profile represents a GitHub profile configuration
type Profile struct {
	Name          string `json:"name"`
	GitUsername   string `json:"git_username"`
	GitEmail      string `json:"git_email"`
	SSHPrivateKey string `json:"ssh_private_key"` // Content of the private key
	SSHPublicKey  string `json:"ssh_public_key"`  // Content of the public key
	IsActive      bool   `json:"is_active"`
	CreatedFrom   string `json:"created_from"` // "system", "manual", "import"
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

	// Validate profile name (should be filesystem safe)
	if strings.ContainsAny(p.Name, "/\\:*?\"<>|") {
		return fmt.Errorf("profile name contains invalid characters")
	}

	return nil
}

// HasSSHKeys returns true if the profile has SSH keys
func (p *Profile) HasSSHKeys() bool {
	return p.SSHPrivateKey != "" && p.SSHPublicKey != ""
}

// LoadSSHKeysFromFiles loads SSH key content from file paths
func (p *Profile) LoadSSHKeysFromFiles(privateKeyPath, publicKeyPath string) error {
	if privateKeyPath != "" {
		privateKey, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}
		p.SSHPrivateKey = string(privateKey)
	}

	if publicKeyPath != "" {
		publicKey, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read public key: %w", err)
		}
		p.SSHPublicKey = string(publicKey)
	}

	return nil
}

// WriteSSHKeysToSystem writes the SSH keys to the system SSH directory
func (p *Profile) WriteSSHKeysToSystem() error {
	if !p.HasSSHKeys() {
		return nil // No keys to write
	}

	sshDir := os.ExpandEnv("$HOME/.ssh")
	dumpDir := filepath.Join(sshDir, "dump")

	// Ensure SSH directory exists with correct permissions
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	// Create dump directory to preserve existing files
	if err := os.MkdirAll(dumpDir, 0700); err != nil {
		return fmt.Errorf("failed to create dump directory: %w", err)
	}

	// Move existing SSH files to dump folder before writing new ones
	if err := p.moveExistingSSHFiles(sshDir, dumpDir); err != nil {
		return fmt.Errorf("failed to backup existing SSH files: %w", err)
	}

	// Write private key
	privateKeyPath := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(privateKeyPath, []byte(p.SSHPrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key
	publicKeyPath := filepath.Join(sshDir, "id_rsa.pub")
	if err := os.WriteFile(publicKeyPath, []byte(p.SSHPublicKey), 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

// moveExistingSSHFiles moves existing SSH files to dump folder
func (p *Profile) moveExistingSSHFiles(sshDir, dumpDir string) error {
	// Generate timestamp for unique dump folder
	timestamp := time.Now().Format("20060102_150405")
	timestampDir := filepath.Join(dumpDir, timestamp)

	// Read all files in SSH directory
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil // If we can't read the directory, assume nothing to move
	}

	// Check if any files exist that need to be moved (exclude dump directory itself)
	hasFiles := false
	for _, file := range files {
		if !file.IsDir() || (file.IsDir() && file.Name() != "dump") {
			hasFiles = true
			break
		}
	}

	if !hasFiles {
		return nil // Nothing to move
	}

	// Create timestamped dump directory
	if err := os.MkdirAll(timestampDir, 0700); err != nil {
		return fmt.Errorf("failed to create timestamped dump directory: %w", err)
	}

	// Move each existing file and directory (except dump folder)
	for _, file := range files {
		// Skip the dump directory itself to avoid recursive issues
		if file.IsDir() && file.Name() == "dump" {
			continue
		}

		srcPath := filepath.Join(sshDir, file.Name())
		dstPath := filepath.Join(timestampDir, file.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s to dump: %w", file.Name(), err)
		}
	}

	return nil
}

// String returns a string representation of the profile
func (p *Profile) String() string {
	active := ""
	if p.IsActive {
		active = " [ACTIVE]"
	}

	sshStatus := ""
	if p.HasSSHKeys() {
		sshStatus = " [SSH]"
	}

	return fmt.Sprintf("%s <%s>%s%s", p.GitUsername, p.GitEmail, sshStatus, active)
}

// CreateFromSystem creates a profile from current system configuration
func CreateFromSystem(name string) (*Profile, error) {
	profile := &Profile{
		Name:        name,
		CreatedFrom: "system",
	}

	// Get current git config
	gitManager := NewGitManager()
	username, email, err := gitManager.GetCurrentGitConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get git config: %w", err)
	}

	profile.GitUsername = username
	profile.GitEmail = email

	// Try to read existing SSH keys
	sshDir := os.ExpandEnv("$HOME/.ssh")
	privateKeyPath := filepath.Join(sshDir, "id_rsa")
	publicKeyPath := filepath.Join(sshDir, "id_rsa.pub")

	// Read private key if exists
	if data, err := os.ReadFile(privateKeyPath); err == nil {
		profile.SSHPrivateKey = string(data)
	}

	// Read public key if exists
	if data, err := os.ReadFile(publicKeyPath); err == nil {
		profile.SSHPublicKey = string(data)
	}

	return profile, nil
}

// Clone creates a copy of the profile with a new name
func (p *Profile) Clone(newName string) *Profile {
	return &Profile{
		Name:          newName,
		GitUsername:   p.GitUsername,
		GitEmail:      p.GitEmail,
		SSHPrivateKey: p.SSHPrivateKey,
		SSHPublicKey:  p.SSHPublicKey,
		IsActive:      false, // New profile is never active
		CreatedFrom:   "clone",
	}
}
