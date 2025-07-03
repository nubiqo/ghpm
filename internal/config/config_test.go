package config

import (
	"os"
	"testing"

	"github.com/huzaifanur/ghpm/internal/profile"
)

// getTestConfigDir is a variable that can be overridden for testing
var getTestConfigDir = GetConfigDir

func newTestConfig(t *testing.T) *Config {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "ghpm-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return NewConfig(tempDir)
}

func TestConfig_AddProfile(t *testing.T) {
	config := newTestConfig(t)

	p := &profile.Profile{
		Name:        "test-profile",
		GitUsername: "testuser",
		GitEmail:    "test@example.com",
		CreatedFrom: "manual",
	}

	err := config.AddProfile(p)
	if err != nil {
		t.Errorf("AddProfile() error = %v", err)
	}

	retrieved, err := config.GetProfile("test-profile")
	if err != nil {
		t.Errorf("GetProfile() error = %v", err)
	}

	if retrieved.Name != p.Name {
		t.Errorf("GetProfile() name = %v, want %v", retrieved.Name, p.Name)
	}
}

func TestConfig_AddProfile_Duplicate(t *testing.T) {
	config := newTestConfig(t)

	p := &profile.Profile{
		Name:        "test-profile",
		GitUsername: "testuser",
		GitEmail:    "test@example.com",
		CreatedFrom: "manual",
	}

	err := config.AddProfile(p)
	if err != nil {
		t.Errorf("AddProfile() error = %v", err)
	}

	err = config.AddProfile(p)
	if err == nil {
		t.Error("AddProfile() expected error for duplicate profile")
	}
}

func TestConfig_GetProfiles(t *testing.T) {
	config := newTestConfig(t)

	p1 := &profile.Profile{
		Name:        "profile1",
		GitUsername: "user1",
		GitEmail:    "user1@example.com",
		CreatedFrom: "manual",
	}

	p2 := &profile.Profile{
		Name:        "profile2",
		GitUsername: "user2",
		GitEmail:    "user2@example.com",
		CreatedFrom: "manual",
	}

	config.AddProfile(p1)
	config.AddProfile(p2)

	profiles := config.GetProfiles()
	if len(profiles) != 2 {
		t.Errorf("GetProfiles() count = %v, want %v", len(profiles), 2)
	}
}

func TestConfig_SetActiveProfile(t *testing.T) {
	config := newTestConfig(t)

	p1 := &profile.Profile{
		Name:        "profile1",
		GitUsername: "user1",
		GitEmail:    "user1@example.com",
		CreatedFrom: "manual",
		IsActive:    true,
	}

	p2 := &profile.Profile{
		Name:        "profile2",
		GitUsername: "user2",
		GitEmail:    "user2@example.com",
		CreatedFrom: "manual",
	}

	config.AddProfile(p1)
	config.AddProfile(p2)

	err := config.SetActiveProfile("profile2")
	if err != nil {
		t.Errorf("SetActiveProfile() error = %v", err)
	}

	active := config.GetActiveProfile()
	if active == nil || active.Name != "profile2" {
		t.Errorf("GetActiveProfile() name = %v, want %v", active.Name, "profile2")
	}

	retrieved1, _ := config.GetProfile("profile1")
	if retrieved1.IsActive {
		t.Error("Profile1 should not be active after switching")
	}
}

func TestConfig_DeleteProfile(t *testing.T) {
	config := newTestConfig(t)

	p := &profile.Profile{
		Name:        "test-profile",
		GitUsername: "testuser",
		GitEmail:    "test@example.com",
		CreatedFrom: "manual",
	}

	config.AddProfile(p)

	err := config.DeleteProfile("test-profile")
	if err != nil {
		t.Errorf("DeleteProfile() error = %v", err)
	}

	_, err = config.GetProfile("test-profile")
	if err == nil {
		t.Error("GetProfile() should return error for deleted profile")
	}
}

func TestConfig_DeleteActiveProfile(t *testing.T) {
	config := newTestConfig(t)

	p := &profile.Profile{
		Name:        "test-profile",
		GitUsername: "testuser",
		GitEmail:    "test@example.com",
		CreatedFrom: "manual",
		IsActive:    true,
	}

	config.AddProfile(p)

	err := config.DeleteProfile("test-profile")
	if err == nil {
		t.Error("DeleteProfile() should return error for active profile")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal name",
			input:    "normal-profile",
			expected: "normal-profile",
		},
		{
			name:     "name with invalid characters",
			input:    "profile/with\\bad:chars",
			expected: "profile_with_bad_chars",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "profile",
		},
		{
			name:     "name with dots and spaces",
			input:    " ..profile.. ",
			expected: "profile",
		},
		{
			name:     "very long name",
			input:    string(make([]byte, 150)),
			expected: string(make([]byte, 100)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if len(result) > 100 {
				t.Errorf("sanitizeFilename() result too long: %d", len(result))
			}
			if tt.name != "very long name" && result != tt.expected {
				t.Errorf("sanitizeFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_ImportProfile_InvalidFile(t *testing.T) {
	config := newTestConfig(t)

	// Test non-existent file
	_, err := config.ImportProfile("non-existent-file.json")
	if err == nil {
		t.Error("ImportProfile() should return error for non-existent file")
	}

	// Test invalid extension
	tmpFile, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = config.ImportProfile(tmpFile.Name())
	if err == nil {
		t.Error("ImportProfile() should return error for non-json file")
	}
}

func TestConfig_ExportProfile_InvalidDirectory(t *testing.T) {
	config := newTestConfig(t)

	p := &profile.Profile{
		Name:        "test-profile",
		GitUsername: "testuser",
		GitEmail:    "test@example.com",
		CreatedFrom: "manual",
	}

	config.AddProfile(p)

	err := config.ExportProfile("test-profile", "/non-existent-directory")
	if err == nil {
		t.Error("ExportProfile() should return error for non-existent directory")
	}
}
