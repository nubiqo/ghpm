package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitManager handles git and SSH operations
type GitManager struct{}

// NewGitManager creates a new GitManager instance
func NewGitManager() *GitManager {
	return &GitManager{}
}

// SwitchProfile switches to the specified profile
func (g *GitManager) SwitchProfile(p *Profile) error {
	// Set git config
	if err := g.setGitConfig(p.GitUsername, p.GitEmail); err != nil {
		return fmt.Errorf("failed to set git config: %w", err)
	}

	// Handle SSH keys if provided
	if p.HasSSHKeys() {
		if err := p.WriteSSHKeysToSystem(); err != nil {
			return fmt.Errorf("failed to write SSH keys: %w", err)
		}

		// Test SSH connection after writing keys
		if err := g.TestSSHConnection(); err != nil {
			log.Printf("Warning: SSH test failed after switching profile: %v", err)
		}
	}

	log.Printf("Switched to profile: %s (%s <%s>)", p.Name, p.GitUsername, p.GitEmail)
	return nil
}

// setGitConfig sets the global git username and email
func (g *GitManager) setGitConfig(username, email string) error {
	// Set username
	cmd := exec.Command("git", "config", "--global", "user.name", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git username: %w", err)
	}

	// Set email
	cmd = exec.Command("git", "config", "--global", "user.email", email)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git email: %w", err)
	}

	return nil
}

// GetCurrentGitConfig returns the current git configuration
func (g *GitManager) GetCurrentGitConfig() (username, email string, err error) {
	// Get username
	cmd := exec.Command("git", "config", "--global", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git username: %w", err)
	}
	username = strings.TrimSpace(string(output))

	// Get email
	cmd = exec.Command("git", "config", "--global", "user.email")
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git email: %w", err)
	}
	email = strings.TrimSpace(string(output))

	return username, email, nil
}

// TestSSHConnection tests the SSH connection to GitHub
func (g *GitManager) TestSSHConnection() error {
	// First check if SSH keys exist
	sshDir := os.ExpandEnv("$HOME/.ssh")
	privateKeyPath := filepath.Join(sshDir, "id_rsa")

	if _, err := os.Stat(privateKeyPath); err != nil {
		return fmt.Errorf("SSH private key not found at %s", privateKeyPath)
	}

	// Check private key permissions
	if err := g.checkSSHKeyPermissions(); err != nil {
		return fmt.Errorf("SSH key permissions error: %w", err)
	}

	// Test SSH connection
	cmd := exec.Command("ssh", "-T", "-o", "StrictHostKeyChecking=no", "-o", "BatchMode=yes", "-o", "ConnectTimeout=10", "git@github.com")
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	// SSH to GitHub returns exit code 1 even on successful authentication
	// Check the output for successful authentication message
	if containsSuccessMessage(outputStr) {
		return nil
	}

	// If we have an error and no success message, include error info
	if err != nil {
		return fmt.Errorf("SSH command failed: %v, output: %s", err, outputStr)
	}

	// Check for common SSH errors
	if strings.Contains(outputStr, "Permission denied") {
		return fmt.Errorf("SSH authentication failed - check your SSH key")
	}

	if strings.Contains(outputStr, "Could not resolve hostname") {
		return fmt.Errorf("network error - could not connect to GitHub")
	}

	if strings.Contains(outputStr, "Connection timed out") {
		return fmt.Errorf("connection timeout - check your network connection")
	}

	return fmt.Errorf("SSH test failed: %s", outputStr)
}

// checkSSHKeyPermissions ensures SSH keys have correct permissions
func (g *GitManager) checkSSHKeyPermissions() error {
	sshDir := os.ExpandEnv("$HOME/.ssh")
	privateKeyPath := filepath.Join(sshDir, "id_rsa")
	publicKeyPath := filepath.Join(sshDir, "id_rsa.pub")

	// Check SSH directory permissions (should be 700)
	if info, err := os.Stat(sshDir); err == nil {
		if info.Mode().Perm() != 0700 {
			if err := os.Chmod(sshDir, 0700); err != nil {
				return fmt.Errorf("failed to fix SSH directory permissions: %w", err)
			}
		}
	}

	// Check private key permissions (should be 600)
	if info, err := os.Stat(privateKeyPath); err == nil {
		if info.Mode().Perm() != 0600 {
			if err := os.Chmod(privateKeyPath, 0600); err != nil {
				return fmt.Errorf("failed to fix private key permissions: %w", err)
			}
		}
	}

	// Check public key permissions (should be 644)
	if info, err := os.Stat(publicKeyPath); err == nil {
		if info.Mode().Perm() != 0644 {
			if err := os.Chmod(publicKeyPath, 0644); err != nil {
				return fmt.Errorf("failed to fix public key permissions: %w", err)
			}
		}
	}

	return nil
}

// containsSuccessMessage checks if the SSH output contains a success message
func containsSuccessMessage(output string) bool {
	// GitHub's successful authentication messages
	successMessages := []string{
		"successfully authenticated",
		"You've successfully authenticated",
		"Hi ",
	}

	outputLower := strings.ToLower(output)
	for _, msg := range successMessages {
		if strings.Contains(outputLower, strings.ToLower(msg)) {
			return true
		}
	}

	return false
}

// ValidateSSHKey validates SSH key format and content
func (g *GitManager) ValidateSSHKey(keyContent string, isPrivate bool) error {
	if strings.TrimSpace(keyContent) == "" {
		return fmt.Errorf("SSH key content is empty")
	}

	if isPrivate {
		// Check for private key markers
		if !strings.Contains(keyContent, "BEGIN") || !strings.Contains(keyContent, "PRIVATE KEY") {
			return fmt.Errorf("invalid private key format")
		}
	} else {
		// Check for public key format
		if !strings.HasPrefix(strings.TrimSpace(keyContent), "ssh-") {
			return fmt.Errorf("invalid public key format")
		}
	}

	return nil
}

// GetSSHKeyFingerprint gets the fingerprint of the current SSH key
func (g *GitManager) GetSSHKeyFingerprint() (string, error) {
	sshDir := os.ExpandEnv("$HOME/.ssh")
	publicKeyPath := filepath.Join(sshDir, "id_rsa.pub")

	if _, err := os.Stat(publicKeyPath); err != nil {
		return "", fmt.Errorf("public key not found")
	}

	cmd := exec.Command("ssh-keygen", "-lf", publicKeyPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get fingerprint: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
