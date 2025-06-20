package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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
	if p.SSHPrivateKey != "" || p.SSHPublicKey != "" {
		if err := g.switchSSHKeys(p); err != nil {
			return fmt.Errorf("failed to switch SSH keys: %w", err)
		}
	}

	log.Printf("Switched to profile: %s", p.Name)
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
	username = string(output)
	if len(username) > 0 && username[len(username)-1] == '\n' {
		username = username[:len(username)-1]
	}

	// Get email
	cmd = exec.Command("git", "config", "--global", "user.email")
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git email: %w", err)
	}
	email = string(output)
	if len(email) > 0 && email[len(email)-1] == '\n' {
		email = email[:len(email)-1]
	}

	return username, email, nil
}

// switchSSHKeys switches the SSH keys for the profile
func (g *GitManager) switchSSHKeys(p *Profile) error {
	sshDir := os.ExpandEnv("$HOME/.ssh")

	// Ensure SSH directory exists
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	// Backup existing keys
	if err := g.backupSSHKeys(p); err != nil {
		log.Printf("Warning: failed to backup SSH keys: %v", err)
	}

	// Copy new keys
	privateKey, publicKey := p.GetSSHKeyPaths()

	if privateKey != "" {
		destPrivate := filepath.Join(sshDir, "id_rsa")
		if err := g.copyFile(privateKey, destPrivate, 0600); err != nil {
			return fmt.Errorf("failed to copy private key: %w", err)
		}
	}

	if publicKey != "" {
		destPublic := filepath.Join(sshDir, "id_rsa.pub")
		if err := g.copyFile(publicKey, destPublic, 0644); err != nil {
			return fmt.Errorf("failed to copy public key: %w", err)
		}
	}

	return nil
}

// backupSSHKeys backs up the current SSH keys
func (g *GitManager) backupSSHKeys(p *Profile) error {
	sshDir := os.ExpandEnv("$HOME/.ssh")
	backupDir := p.GetBackupDir()

	// Create backup directory
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup id_rsa if exists
	privateKeyPath := filepath.Join(sshDir, "id_rsa")
	if _, err := os.Stat(privateKeyPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		backupPath := filepath.Join(backupDir, fmt.Sprintf("id_rsa_%s", timestamp))
		if err := g.copyFile(privateKeyPath, backupPath, 0600); err != nil {
			return fmt.Errorf("failed to backup private key: %w", err)
		}
	}

	// Backup id_rsa.pub if exists
	publicKeyPath := filepath.Join(sshDir, "id_rsa.pub")
	if _, err := os.Stat(publicKeyPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		backupPath := filepath.Join(backupDir, fmt.Sprintf("id_rsa.pub_%s", timestamp))
		if err := g.copyFile(publicKeyPath, backupPath, 0644); err != nil {
			return fmt.Errorf("failed to backup public key: %w", err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst with the specified permissions
func (g *GitManager) copyFile(src, dst string, perm os.FileMode) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// TestSSHConnection tests the SSH connection to GitHub
func (g *GitManager) TestSSHConnection() error {
	cmd := exec.Command("ssh", "-T", "git@github.com")
	output, err := cmd.CombinedOutput()

	// SSH to GitHub returns exit code 1 even on successful authentication
	// Check the output for successful authentication message
	outputStr := string(output)
	if err != nil && !containsSuccessMessage(outputStr) {
		return fmt.Errorf("SSH test failed: %s", outputStr)
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

	for _, msg := range successMessages {
		if len(output) > 0 && len(msg) > 0 && contains(output, msg) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsAt(s, substr, 0)
}

// containsAt checks if a string contains a substring starting at any position
func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
