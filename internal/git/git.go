package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/huzaifanur/ghpm/pkg/logger"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (g *Manager) SwitchProfile(profile ProfileInterface) error {
	log := logger.New()
	defer log.Close()

	if err := g.setGitConfig(profile.GetGitUsername(), profile.GetGitEmail()); err != nil {
		return fmt.Errorf("failed to set git config: %w", err)
	}

	if profile.HasSSHKeys() {
		if err := profile.WriteSSHKeysToSystem(); err != nil {
			return fmt.Errorf("failed to write SSH keys: %w", err)
		}

		if err := g.TestSSHConnection(); err != nil {
			log.Warnw("SSH test failed after switching profile", "error", err)
		}
	}

	log.Infow("Switched to profile",
		"name", profile.GetName(),
		"username", profile.GetGitUsername(),
		"email", profile.GetGitEmail())
	return nil
}

func (g *Manager) setGitConfig(username, email string) error {
	if err := validateGitInput(username, email); err != nil {
		return fmt.Errorf("invalid git configuration: %w", err)
	}

	cmd := exec.Command("git", "config", "--global", "user.name", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git username: %w", err)
	}

	cmd = exec.Command("git", "config", "--global", "user.email", email)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git email: %w", err)
	}

	return nil
}

func (g *Manager) GetCurrentGitConfig() (username, email string, err error) {
	cmd := exec.Command("git", "config", "--global", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git username: %w", err)
	}
	username = strings.TrimSpace(string(output))

	cmd = exec.Command("git", "config", "--global", "user.email")
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git email: %w", err)
	}
	email = strings.TrimSpace(string(output))

	return username, email, nil
}

func (g *Manager) TestSSHConnection() error {
	privateKeyPath, _, err := detectSSHKeyPaths()
	if err != nil {
		return fmt.Errorf("no SSH keys found: %w", err)
	}

	if _, err := os.Stat(privateKeyPath); err != nil {
		return fmt.Errorf("SSH private key not found at %s", privateKeyPath)
	}

	if err := g.checkSSHKeyPermissions(); err != nil {
		return fmt.Errorf("SSH key permissions error: %w", err)
	}

	cmd := exec.Command("ssh", "-T", "-o", "StrictHostKeyChecking=no", "-o", "BatchMode=yes", "-o", "ConnectTimeout=10", "git@github.com")
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	if containsSuccessMessage(outputStr) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("SSH command failed: %v, output: %s", err, outputStr)
	}

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

func (g *Manager) checkSSHKeyPermissions() error {
	sshDir := os.ExpandEnv("$HOME/.ssh")

	privateKeyPath, publicKeyPath, err := detectSSHKeyPaths()
	if err != nil {
		return fmt.Errorf("no SSH keys found: %w", err)
	}

	if info, err := os.Stat(sshDir); err == nil {
		if info.Mode().Perm() != 0700 {
			if err := os.Chmod(sshDir, 0700); err != nil {
				return fmt.Errorf("failed to fix SSH directory permissions: %w", err)
			}
		}
	}

	if info, err := os.Stat(privateKeyPath); err == nil {
		if info.Mode().Perm() != 0600 {
			if err := os.Chmod(privateKeyPath, 0600); err != nil {
				return fmt.Errorf("failed to fix private key permissions: %w", err)
			}
		}
	}

	if info, err := os.Stat(publicKeyPath); err == nil {
		if info.Mode().Perm() != 0644 {
			if err := os.Chmod(publicKeyPath, 0644); err != nil {
				return fmt.Errorf("failed to fix public key permissions: %w", err)
			}
		}
	}

	return nil
}

func containsSuccessMessage(output string) bool {
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

func (g *Manager) ValidateSSHKey(keyContent string, isPrivate bool) error {
	if strings.TrimSpace(keyContent) == "" {
		return fmt.Errorf("SSH key content is empty")
	}

	if isPrivate {
		if !strings.Contains(keyContent, "BEGIN") || !strings.Contains(keyContent, "PRIVATE KEY") {
			return fmt.Errorf("invalid private key format")
		}
	} else {
		if !strings.HasPrefix(strings.TrimSpace(keyContent), "ssh-") {
			return fmt.Errorf("invalid public key format")
		}
	}

	return nil
}

func (g *Manager) GetSSHKeyFingerprint() (string, error) {
	_, publicKeyPath, err := detectSSHKeyPaths()
	if err != nil {
		return "", fmt.Errorf("no SSH keys found: %w", err)
	}

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

func detectSSHKeyPaths() (string, string, error) {
	sshDir := os.ExpandEnv("$HOME/.ssh")

	keyTypes := []string{
		"id_ed25519",
		"id_ecdsa",
		"id_rsa",
		"id_dsa",
	}

	for _, keyType := range keyTypes {
		privateKeyPath := filepath.Join(sshDir, keyType)
		publicKeyPath := filepath.Join(sshDir, keyType+".pub")

		if _, err := os.Stat(privateKeyPath); err == nil {
			if _, err := os.Stat(publicKeyPath); err == nil {
				return privateKeyPath, publicKeyPath, nil
			}
		}
	}

	return "", "", fmt.Errorf("no SSH key pair found")
}

func validateGitInput(username, email string) error {
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9\s._-]+$`)
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("invalid username: contains unsafe characters")
	}

	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	if len(username) > 100 {
		return fmt.Errorf("username too long (max 100 characters)")
	}
	if len(email) > 254 {
		return fmt.Errorf("email too long (max 254 characters)")
	}

	return nil
}
