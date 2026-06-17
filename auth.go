package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long: `Manage authentication and API credentials.

Examples:
  workbrew auth status
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := loadAPIToken()
		if err != nil {
			fmt.Println("No API token found")
			return
		}

		fmt.Println("API token found in keychain")
		fmt.Printf("Token length: %d characters\n", len(token))
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
