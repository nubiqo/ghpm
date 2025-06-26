package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	// Set up logging with single line format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Create app
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	// Set app metadata
	metadata := myApp.Metadata()
	metadata.Name = "GitHub Profile Manager"
	metadata.ID = "com.ghpm.app"

	// Initialize config directory and load configuration
	if err := initializeConfigDirectory(); err != nil {
		log.Fatalf("Failed to initialize config directory: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("Failed to load config: %v - creating new config directory", err)
		cfg = &Config{profiles: make(map[string]*Profile)}
	}

	log.Printf("Loaded %d profiles from config", len(cfg.GetProfiles()))

	// Create and show main window
	ui := NewUI(myApp, cfg)
	ui.Show()

	myApp.Run()
}

// initializeConfigDirectory ensures the .ghpm directory exists with proper permissions
func initializeConfigDirectory() error {
	configDir := GetConfigDir()

	// Create config directory with proper permissions
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Ensure SSH directory exists with proper permissions for SSH operations
	sshDir := os.ExpandEnv("$HOME/.ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}

	log.Printf("Initialized config directory: %s", configDir)
	return nil
}

func init() {
	// Set app metadata early
	os.Setenv("FYNE_APP_ID", "com.ghpm.app")
}
