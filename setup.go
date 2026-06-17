package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var workspaceURL string

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure Workbrew CLI",
	Run: func(cmd *cobra.Command, args []string) {
		runSetup()
	},
}

func runSetup() {
	if err := ensureConfigDir(); err != nil {
		fmt.Println("Could not create config directory:", err)
		return
	}

	config := Config{
		URL: workspaceURL,
	}

	if err := saveConfig(config); err != nil {
		fmt.Println("Could not save config:", err)
		return
	}

	fmt.Println("Configuration saved")
	fmt.Println("Config path:", getConfigPath())
}

func init() {

	setupCmd.Flags().StringVar(
		&workspaceURL,
		"url",
		"",
		"Workbrew workspace URL",
	)
	setupCmd.MarkFlagRequired("url")
}
