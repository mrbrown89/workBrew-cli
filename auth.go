package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Args:  cobra.ArbitraryArgs,
		Long: `Manage authentication and API credentials.

Examples:
  workbrew-cli auth status
  workbrew-cli auth test
`,
		RunE: runGroup,
	}

	authCmd.AddCommand(
		&cobra.Command{
			Use:   "status",
			Short: "Check authentication status",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runAuthStatus(cmd)
			},
		},
		&cobra.Command{
			Use:   "test",
			Short: "Test Workbrew API authentication",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runAuthTest(cmd)
			},
		},
	)

	return authCmd
}

func runAuthStatus(cmd *cobra.Command) error {
	token, err := loadToken()
	if err != nil {
		return err
	}

	out := newPrinter(cmd.OutOrStdout())
	out.Println("API token found in keychain")
	out.Printf("Token length: %d characters\n", len(token))
	return out.Err()
}

func runAuthTest(cmd *cobra.Command) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	if _, err := client.Devices(cmd.Context()); err != nil {
		return fmt.Errorf("authentication test failed: %w", err)
	}

	out := newPrinter(cmd.OutOrStdout())
	out.Println("Authentication successful")
	out.Println("Workspace URL:", client.BaseURL())
	return out.Err()
}
