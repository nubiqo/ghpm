package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Profiles []Profile `json:"profiles"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return os.ExpandEnv("$HOME/.github-profile-manager.json")
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{Profiles: []Profile{}}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to file
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddProfile adds a new profile to the configuration
func (c *Config) AddProfile(p Profile) error {
	// Check for duplicate names
	for _, existing := range c.Profiles {
		if existing.Name == p.Name {
			return fmt.Errorf("profile with name '%s' already exists", p.Name)
		}
	}

	c.Profiles = append(c.Profiles, p)
	return c.Save()
}

// UpdateProfile updates an existing profile
func (c *Config) UpdateProfile(name string, p Profile) error {
	for i, existing := range c.Profiles {
		if existing.Name == name {
			// If name changed, check for duplicates
			if name != p.Name {
				for j, other := range c.Profiles {
					if i != j && other.Name == p.Name {
						return fmt.Errorf("profile with name '%s' already exists", p.Name)
					}
				}
			}
			c.Profiles[i] = p
			return c.Save()
		}
	}
	return fmt.Errorf("profile '%s' not found", name)
}

// DeleteProfile removes a profile from the configuration
func (c *Config) DeleteProfile(name string) error {
	for i, p := range c.Profiles {
		if p.Name == name {
			// Don't delete active profile
			if p.IsActive {
				return fmt.Errorf("cannot delete active profile")
			}
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return c.Save()
		}
	}
	return fmt.Errorf("profile '%s' not found", name)
}

// GetProfile returns a profile by name
func (c *Config) GetProfile(name string) (*Profile, error) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i], nil
		}
	}
	return nil, fmt.Errorf("profile '%s' not found", name)
}

// GetActiveProfile returns the currently active profile
func (c *Config) GetActiveProfile() *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].IsActive {
			return &c.Profiles[i]
		}
	}
	return nil
}

// SetActiveProfile sets the active profile
func (c *Config) SetActiveProfile(name string) error {
	found := false
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			c.Profiles[i].IsActive = true
			found = true
		} else {
			c.Profiles[i].IsActive = false
		}
	}

	if !found {
		return fmt.Errorf("profile '%s' not found", name)
	}

	return c.Save()
}
