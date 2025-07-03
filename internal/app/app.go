package app

import (
	"os"

	"github.com/huzaifanur/ghpm/pkg/logger"
)

func InitializeConfigDirectory() error {
	log := logger.New()
	defer log.Close()

	configDir := getConfigDir()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	sshDir := os.ExpandEnv("$HOME/.ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}

	log.Infow("Initialized config directory", "path", configDir)
	return nil
}

func getConfigDir() string {
	return os.ExpandEnv("$HOME/.ghpm")
}

func GetConfigDir() string {
	return getConfigDir()
}
