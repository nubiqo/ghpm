package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/huzaifanur/ghpm/internal/git"
)

type Profile struct {
	Name          string `json:"name"`
	GitUsername   string `json:"git_username"`
	GitEmail      string `json:"git_email"`
	SSHPrivateKey string `json:"ssh_private_key"`
	SSHPublicKey  string `json:"ssh_public_key"`
	IsActive      bool   `json:"is_active"`
	CreatedFrom   string `json:"created_from"`
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

	if strings.ContainsAny(p.Name, "/\\:*?\"<>|") {
		return fmt.Errorf("profile name contains invalid characters")
	}

	if err := validateGitInput(p.GitUsername, p.GitEmail); err != nil {
		return fmt.Errorf("invalid git configuration: %w", err)
	}

	if p.SSHPrivateKey == "" || p.SSHPublicKey == "" {
		return fmt.Errorf("SSH private and public keys are required")
	}

	return nil
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

func (p *Profile) HasSSHKeys() bool {
	return p.SSHPrivateKey != "" && p.SSHPublicKey != ""
}

func (p *Profile) detectKeyType() string {
	if p.SSHPublicKey == "" {
		return "id_rsa"
	}

	publicKey := strings.TrimSpace(p.SSHPublicKey)

	if strings.HasPrefix(publicKey, "ssh-ed25519") {
		return "id_ed25519"
	} else if strings.HasPrefix(publicKey, "ssh-ecdsa") {
		return "id_ecdsa"
	} else if strings.HasPrefix(publicKey, "ssh-rsa") {
		return "id_rsa"
	} else if strings.HasPrefix(publicKey, "ssh-dss") {
		return "id_dsa"
	}

	return "id_rsa"
}

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

func (p *Profile) WriteSSHKeysToSystem() error {
	if !p.HasSSHKeys() {
		return nil
	}

	sshDir := os.ExpandEnv("$HOME/.ssh")

	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	if info, err := os.Stat(sshDir); err == nil {
		if info.Mode().Perm() != 0700 {
			if err := os.Chmod(sshDir, 0700); err != nil {
				return fmt.Errorf("failed to set SSH directory permissions: %w", err)
			}
		}
	}

	keyType := p.detectKeyType()

	privateKeyPath := filepath.Join(sshDir, keyType)
	if err := p.atomicWriteFile(privateKeyPath, []byte(p.SSHPrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	publicKeyPath := filepath.Join(sshDir, keyType+".pub")
	if err := p.atomicWriteFile(publicKeyPath, []byte(p.SSHPublicKey), 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

func (p *Profile) atomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	tmpFile, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err := tmpFile.Chmod(perm); err != nil {
		return fmt.Errorf("failed to set permissions on temporary file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err := os.Rename(tmpPath, filename); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

func (p *Profile) String() string {
	active := ""
	if p.IsActive {
		active = " [ACTIVE]"
	}

	return fmt.Sprintf("%s <%s>%s", p.GitUsername, p.GitEmail, active)
}

func CreateFromSystem(name string) (*Profile, error) {
	profile := &Profile{
		Name:        name,
		CreatedFrom: "system",
	}

	gitManager := git.NewManager()
	username, email, err := gitManager.GetCurrentGitConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get git config: %w", err)
	}

	profile.GitUsername = username
	profile.GitEmail = email

	privateKeyPath, publicKeyPath, err := detectSSHKeyPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to detect SSH keys: %w", err)
	}

	if data, err := os.ReadFile(privateKeyPath); err == nil {
		profile.SSHPrivateKey = string(data)
	}

	if data, err := os.ReadFile(publicKeyPath); err == nil {
		profile.SSHPublicKey = string(data)
	}

	return profile, nil
}

func (p *Profile) Clone(newName string) *Profile {
	return &Profile{
		Name:          newName,
		GitUsername:   p.GitUsername,
		GitEmail:      p.GitEmail,
		SSHPrivateKey: p.SSHPrivateKey,
		SSHPublicKey:  p.SSHPublicKey,
		IsActive:      false,
		CreatedFrom:   "clone",
	}
}

func (p *Profile) GetName() string {
	return p.Name
}

func (p *Profile) GetGitUsername() string {
	return p.GitUsername
}

func (p *Profile) GetGitEmail() string {
	return p.GitEmail
}
