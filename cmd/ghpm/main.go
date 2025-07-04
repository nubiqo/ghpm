package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/huzaifanur/ghpm/internal/config"
	"github.com/huzaifanur/ghpm/internal/ui"
	"github.com/huzaifanur/ghpm/pkg/logger"
)

func main() {
	logger := logger.New()
	logger.Infow("Starting GHPM application")

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalw("Failed to load config", "error", err)
	}

	fyneApp := app.NewWithID("com.ghpm.app")
	mainUI := ui.NewUI(fyneApp, cfg)
	mainUI.Show()
}

// xasxa
