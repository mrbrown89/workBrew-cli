package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const setupHint = "run `workbrew-cli setup --url <workspace-url>`"

type Config struct {
	URL string `yaml:"url"`
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("finding user config directory: %w", err)
	}

	return filepath.Join(configDir, "workbrew-cli", "config.yaml"), nil
}

func ensureConfigDir() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	return os.MkdirAll(filepath.Dir(configPath), 0o700)
}

func saveConfig(config Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, content, 0o600)
}

func loadConfig() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("no config found at %s; %s", configPath, setupHint)
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", configPath, err)
	}

	if config.URL == "" {
		return Config{}, fmt.Errorf("config %s has no workspace URL; %s", configPath, setupHint)
	}

	return config, nil
}
