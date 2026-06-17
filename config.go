package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Could not determine user config directory")
		return ""
	}

	return filepath.Join(configDir, "workbrew-cli", "config.yaml")
}

func ensureConfigDir() error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	return os.MkdirAll(configDir, 0700)
}

type Config struct {
	URL string
}

func saveConfig(config Config) error {
	content := fmt.Sprintf("url: %s\n", config.URL)

	return os.WriteFile(
		getConfigPath(),
		[]byte(content),
		0600,
	)
}
