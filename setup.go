package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var workspaceURL string

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure Workbrew CLI",
	Run: func(cmd *cobra.Command, args []string) {
		runSetup()
	},
}

func promptForToken() (string, error) {
	fmt.Print("Workbrew API token: ")

	tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(tokenBytes)), nil
}

func runSetup() {
	if err := ensureConfigDir(); err != nil {
		fmt.Println("Could not create config directory:", err)
		return
	}

	token, err := promptForToken()
	if err != nil {
		fmt.Println("Could not read API token:", err)
		return
	}

	if token == "" {
		fmt.Println("API token cannot be empty")
		return
	}

	config := Config{
		URL: workspaceURL,
	}

	if err := saveConfig(config); err != nil {
		fmt.Println("Could not save config:", err)
		return
	}

	if err := saveAPIToken(token); err != nil {
		fmt.Println("Could not save API token:", err)
		return
	}

	fmt.Println("Configuration saved")
	fmt.Println("API token saved to keychain")
	fmt.Println("Config path:", getConfigPath())
}

func init() {
	setupCmd.Flags().StringVar(
		&workspaceURL,
		"url",
		"",
		"Workbrew workspace URL",
	)

	if err := setupCmd.MarkFlagRequired("url"); err != nil {
		log.Fatal(err)
	}
}
