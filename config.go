package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	URL string `yaml:"url"`
}

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

func saveConfig(config Config) error {
	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(
		getConfigPath(),
		content,
		0600,
	)
}

func loadConfig() (Config, error) {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return Config{}, err
	}

	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
