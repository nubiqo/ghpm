package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/huzaifanur/ghpm/internal/profile"
)

type Config struct {
	profiles sync.Map
	configDir string
}

func NewConfig(configDir ...string) *Config {
	c := &Config{}
	if len(configDir) > 0 && configDir[0] != "" {
		c.configDir = configDir[0]
	} else {
		c.configDir = GetConfigDir()
	}
	return c
}

var getConfigDirFunc = func() string {
	return os.ExpandEnv("$HOME/.ghpm")
}

func GetConfigDir() string {
	return getConfigDirFunc()
}

func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[/\\:*?"<>|]`)
	safe := re.ReplaceAllString(name, "_")

	safe = strings.Trim(safe, ". \t\n\r")

	if safe == "" {
		safe = "profile"
	}

	if len(safe) > 100 {
		safe = safe[:100]
	}

	return safe
}

func LoadConfig() (*Config, error) {
	config := NewConfig() // Use NewConfig to set configDir

	if err := os.MkdirAll(config.configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	files, err := os.ReadDir(config.configDir)
	if err != nil {
		return config, nil
	}

	var activeProfileFound bool
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			profilePath := filepath.Join(config.configDir, file.Name())
			p, err := loadProfileFromFile(profilePath)
			if err != nil {
				continue
			}

			if p.IsActive {
				if activeProfileFound {
					p.IsActive = false // Deactivate if another active profile already found
					config.saveProfileToFile(p) // Save the deactivated profile
				} else {
					activeProfileFound = true
				}
			}
			config.profiles.Store(p.Name, p)
		}
	}

	return config, nil
}

func loadProfileFromFile(path string) (*profile.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var p profile.Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &p, nil
}

func (c *Config) GetProfiles() []*profile.Profile {
	var profiles []*profile.Profile
	c.profiles.Range(func(key, value any) bool {
		profiles = append(profiles, value.(*profile.Profile))
		return true
	})
	return profiles
}

func (c *Config) AddProfile(p *profile.Profile) error {
	if _, exists := c.profiles.Load(p.Name); exists {
		return fmt.Errorf("profile with name '%s' already exists", p.Name)
	}

	if err := c.saveProfileToFile(p); err != nil {
		return err
	}

	c.profiles.Store(p.Name, p)
	return nil
}

func (c *Config) UpdateProfile(oldName string, p *profile.Profile) error {
	if oldName != p.Name {
		if _, exists := c.profiles.Load(p.Name); exists {
			return fmt.Errorf("profile with name '%s' already exists", p.Name)
		}

		if err := c.saveProfileToFile(p); err != nil {
			return err
		}

		oldPath := filepath.Join(c.configDir, oldName+".json")
		os.Remove(oldPath)

		c.profiles.Delete(oldName)
		c.profiles.Store(p.Name, p)
	} else {
		if err := c.saveProfileToFile(p); err != nil {
			return err
		}
		c.profiles.Store(p.Name, p)
	}

	return nil
}

func (c *Config) DeleteProfile(name string) error {
	value, exists := c.profiles.Load(name)
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	p := value.(*profile.Profile)
	if p.IsActive {
		return fmt.Errorf("cannot delete active profile")
	}

	profilePath := filepath.Join(c.configDir, name+".json")
	if err := os.Remove(profilePath); err != nil {
		return fmt.Errorf("failed to remove profile file: %w", err)
	}

	c.profiles.Delete(name)
	return nil
}

func (c *Config) GetProfile(name string) (*profile.Profile, error) {
	value, exists := c.profiles.Load(name)
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}
	return value.(*profile.Profile), nil
}

func (c *Config) GetActiveProfile() *profile.Profile {
	var activeProfile *profile.Profile
	c.profiles.Range(func(key, value any) bool {
		p := value.(*profile.Profile)
		if p.IsActive {
			activeProfile = p
			return false
		}
		return true
	})
	return activeProfile
}

func (c *Config) SetActiveProfile(name string) error {
	value, exists := c.profiles.Load(name)
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	targetProfile := value.(*profile.Profile)

	c.profiles.Range(func(key, value any) bool {
		p := value.(*profile.Profile)
		if p.IsActive {
			p.IsActive = false
			if err := c.saveProfileToFile(p); err != nil {
				fmt.Printf("Warning: failed to save profile %s: %v\n", p.Name, err)
			}
		}
		return true
	})

	targetProfile.IsActive = true
	return c.saveProfileToFile(targetProfile)
}

func (c *Config) saveProfileToFile(p *profile.Profile) error {
	profilePath := filepath.Join(c.configDir, p.Name+".json")

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

func (c *Config) ExportProfile(name, exportDir string) error {
	value, exists := c.profiles.Load(name)
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	p := value.(*profile.Profile)

	if exportDir == "" {
		return fmt.Errorf("export directory cannot be empty")
	}

	if info, err := os.Stat(exportDir); err != nil {
		return fmt.Errorf("export directory does not exist: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("export path is not a directory")
	}

	safeName := sanitizeFilename(name)
	exportPath := filepath.Join(exportDir, safeName+".json")

	if _, err := os.Stat(exportPath); err == nil {
		return fmt.Errorf("export file already exists: %s", exportPath)
	}

	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid profile data: %w", err)
	}

	if p.HasSSHKeys() {
		if strings.TrimSpace(p.SSHPrivateKey) == "" {
			return fmt.Errorf("profile marked as having SSH keys but private key is empty")
		}
		if strings.TrimSpace(p.SSHPublicKey) == "" {
			return fmt.Errorf("profile marked as having SSH keys but public key is empty")
		}
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	tmpFile, err := os.CreateTemp(exportDir, safeName+".json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary export file: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write export data: %w", err)
	}

	if err := tmpFile.Chmod(0600); err != nil {
		return fmt.Errorf("failed to set export file permissions: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync export file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close export file: %w", err)
	}

	if err := os.Rename(tmpPath, exportPath); err != nil {
		return fmt.Errorf("failed to finalize export: %w", err)
	}

	return nil
}

func (c *Config) ImportProfile(filePath string) (*profile.Profile, error) {
	if filePath == "" {
		return nil, fmt.Errorf("import file path cannot be empty")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("import file does not exist: %w", err)
	}

	if info.Size() > 1024*1024 {
		return nil, fmt.Errorf("import file too large (max 1MB)")
	}

	if info.Size() < 10 {
		return nil, fmt.Errorf("import file too small to be valid")
	}

	if !strings.HasSuffix(strings.ToLower(filePath), ".json") {
		return nil, fmt.Errorf("import file must have .json extension")
	}

	p, err := loadProfileFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to import profile: %w", err)
	}

	if err := c.validateImportedProfile(p); err != nil {
		return nil, fmt.Errorf("invalid imported profile: %w", err)
	}

	if _, exists := c.profiles.Load(p.Name); exists {
		return nil, fmt.Errorf("profile with name '%s' already exists", p.Name)
	}

	p.IsActive = false
	p.CreatedFrom = "import"

	if err := c.AddProfile(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *Config) validateImportedProfile(p *profile.Profile) error {
	if err := p.Validate(); err != nil {
		return err
	}

	if len(p.Name) > 100 {
		return fmt.Errorf("profile name too long (max 100 characters)")
	}

	if len(p.GitUsername) > 100 {
		return fmt.Errorf("git username too long (max 100 characters)")
	}

	if len(p.GitEmail) > 254 {
		return fmt.Errorf("git email too long (max 254 characters)")
	}

	if p.SSHPrivateKey != "" || p.SSHPublicKey != "" {
		if p.SSHPrivateKey == "" || p.SSHPublicKey == "" {
			return fmt.Errorf("if SSH keys are provided, both private and public keys must be present")
		}

		if len(p.SSHPrivateKey) > 16384 {
			return fmt.Errorf("private SSH key too large (max 16KB)")
		}

		if len(p.SSHPublicKey) > 4096 {
			return fmt.Errorf("public SSH key too large (max 4KB)")
		}
	}

	validSources := []string{"system", "manual", "import", "clone"}
	validSource := false
	for _, source := range validSources {
		if p.CreatedFrom == source {
			validSource = true
			break
		}
	}
	if !validSource {
		p.CreatedFrom = "import"
	}

	return nil
}
