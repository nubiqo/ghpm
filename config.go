package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the application configuration
type Config struct {
	profiles map[string]*Profile
}

// GetConfigDir returns the path to the config directory
func GetConfigDir() string {
	return os.ExpandEnv("$HOME/.ghpm")
}

// LoadConfig loads all profiles from the config directory
func LoadConfig() (*Config, error) {
	configDir := GetConfigDir()

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	config := &Config{
		profiles: make(map[string]*Profile),
	}

	// Read all profile files
	files, err := os.ReadDir(configDir)
	if err != nil {
		return config, nil // Return empty config if directory can't be read
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			profilePath := filepath.Join(configDir, file.Name())
			profile, err := loadProfileFromFile(profilePath)
			if err != nil {
				continue // Skip invalid profiles
			}
			config.profiles[profile.Name] = profile
		}
	}

	return config, nil
}

// loadProfileFromFile loads a single profile from a JSON file
func loadProfileFromFile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}

// GetProfiles returns all profiles as a slice
func (c *Config) GetProfiles() []*Profile {
	profiles := make([]*Profile, 0, len(c.profiles))
	for _, profile := range c.profiles {
		profiles = append(profiles, profile)
	}
	return profiles
}

// AddProfile adds a new profile to the configuration
func (c *Config) AddProfile(p *Profile) error {
	// Check for duplicate names
	if _, exists := c.profiles[p.Name]; exists {
		return fmt.Errorf("profile with name '%s' already exists", p.Name)
	}

	// Save profile to file
	if err := c.saveProfileToFile(p); err != nil {
		return err
	}

	c.profiles[p.Name] = p
	return nil
}

// UpdateProfile updates an existing profile
func (c *Config) UpdateProfile(oldName string, p *Profile) error {
	// If name changed, check for duplicates and remove old file
	if oldName != p.Name {
		if _, exists := c.profiles[p.Name]; exists {
			return fmt.Errorf("profile with name '%s' already exists", p.Name)
		}

		// Remove old file
		oldPath := filepath.Join(GetConfigDir(), oldName+".json")
		os.Remove(oldPath)

		// Remove from map
		delete(c.profiles, oldName)
	}

	// Save updated profile
	if err := c.saveProfileToFile(p); err != nil {
		return err
	}

	c.profiles[p.Name] = p
	return nil
}

// DeleteProfile removes a profile from the configuration
func (c *Config) DeleteProfile(name string) error {
	profile, exists := c.profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Don't delete active profile
	if profile.IsActive {
		return fmt.Errorf("cannot delete active profile")
	}

	// Remove file
	profilePath := filepath.Join(GetConfigDir(), name+".json")
	if err := os.Remove(profilePath); err != nil {
		return fmt.Errorf("failed to remove profile file: %w", err)
	}

	delete(c.profiles, name)
	return nil
}

// GetProfile returns a profile by name
func (c *Config) GetProfile(name string) (*Profile, error) {
	profile, exists := c.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}
	return profile, nil
}

// GetActiveProfile returns the currently active profile
func (c *Config) GetActiveProfile() *Profile {
	for _, profile := range c.profiles {
		if profile.IsActive {
			return profile
		}
	}
	return nil
}

// SetActiveProfile sets the active profile
func (c *Config) SetActiveProfile(name string) error {
	targetProfile, exists := c.profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Set all profiles to inactive
	for _, profile := range c.profiles {
		if profile.IsActive {
			profile.IsActive = false
			c.saveProfileToFile(profile)
		}
	}

	// Set target profile to active
	targetProfile.IsActive = true
	return c.saveProfileToFile(targetProfile)
}

// saveProfileToFile saves a profile to its JSON file
func (c *Config) saveProfileToFile(p *Profile) error {
	configDir := GetConfigDir()
	profilePath := filepath.Join(configDir, p.Name+".json")

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// ExportProfile exports a profile to a specified directory
func (c *Config) ExportProfile(name, exportDir string) error {
	profile, exists := c.profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	exportPath := filepath.Join(exportDir, name+".json")

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(exportPath, data, 0600); err != nil {
		return fmt.Errorf("failed to export profile: %w", err)
	}

	return nil
}

// ImportProfile imports a profile from a JSON file
func (c *Config) ImportProfile(filePath string) (*Profile, error) {
	profile, err := loadProfileFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to import profile: %w", err)
	}

	// Check for duplicate names
	if _, exists := c.profiles[profile.Name]; exists {
		return nil, fmt.Errorf("profile with name '%s' already exists", profile.Name)
	}

	// Add the imported profile
	if err := c.AddProfile(profile); err != nil {
		return nil, err
	}

	return profile, nil
}
