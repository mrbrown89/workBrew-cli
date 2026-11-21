package main

import (
	"errors"
	"fmt"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
	"github.com/zalando/go-keyring"
)

// loadToken reads the API token from the keychain. A missing token gets a
// "run setup" message; any other keychain error is passed up as it is.
func loadToken() (string, error) {
	token, err := loadAPIToken()
	if errors.Is(err, keyring.ErrNotFound) {
		return "", fmt.Errorf("no API token found in keychain; %s", setupHint)
	}
	if err != nil {
		return "", fmt.Errorf("reading API token from keychain: %w", err)
	}

	return token, nil
}

// loadClient loads the config and API token once and returns a client for
// the configured workspace. Only commands that call the API use it.
func loadClient() (*workbrew.Client, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	token, err := loadToken()
	if err != nil {
		return nil, err
	}

	return workbrew.New(config.URL, token)
}
