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

var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test Workbrew API authentication",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Println("Could not load config. Run setup first.")
			return
		}

		token, err := loadAPIToken()
		if err != nil {
			fmt.Println("No API token found. Run setup first.")
			return
		}

		var devices []Device

		err = workbrewGetJSON(config, token, "devices.json", &devices)

		if err != nil {
			fmt.Println("Authentication test failed:", err)
			return
		}

		fmt.Println("Authentication successful")
		fmt.Println("Workspace URL:", config.URL)
	},
}

func init() {

	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authTestCmd)

}
