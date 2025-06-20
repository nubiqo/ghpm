package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

func main() {
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Create app
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	// Initialize config
	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("Failed to load config: %v - creating new config", err)
		cfg = &Config{Profiles: []Profile{}}
		if err := cfg.Save(); err != nil {
			log.Fatalf("Failed to create config: %v", err)
		}
	}

	// Create and show main window
	ui := NewUI(myApp, cfg)
	ui.Show()

	myApp.Run()
}

func init() {
	// Ensure config directory exists
	configDir := os.ExpandEnv("$HOME/.config/github-profile-manager")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}
}
