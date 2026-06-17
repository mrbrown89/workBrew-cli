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

	fmt.Println("Setup not implemented yet")
	fmt.Println("Workspace URL:", workspaceURL)

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
