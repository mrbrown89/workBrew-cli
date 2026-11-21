package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	var workspaceURL string

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure Workbrew CLI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runSetup(cmd, workspaceURL)
		},
	}

	setupCmd.Flags().StringVar(&workspaceURL, "url", "", "Workbrew workspace URL (https)")
	// MarkFlagRequired only fails if the flag doesn't exist, which it does.
	_ = setupCmd.MarkFlagRequired("url")

	return setupCmd
}

// validateWorkspaceURL checks that raw is an https URL with a host and
// returns it with any trailing slashes removed.
func validateWorkspaceURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid --url: %w", err)
	}
	if u.Scheme != "https" {
		return "", fmt.Errorf("invalid --url %q: must use https", raw)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("invalid --url %q: missing host", raw)
	}

	return strings.TrimRight(raw, "/"), nil
}

// promptForToken asks for the API token on stderr and reads it from the
// command's stdin, which must be a terminal.
func promptForToken(cmd *cobra.Command) (string, error) {
	fd, ok := terminalFd(cmd.InOrStdin())
	if !ok {
		return "", errors.New("stdin is not a terminal; run setup interactively to enter the API token")
	}

	errOut := newPrinter(cmd.ErrOrStderr())
	errOut.Print("Workbrew API token: ")
	token, err := readPassword(cmd.Context(), fd)
	errOut.Println()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(token), nil
}

func runSetup(cmd *cobra.Command, rawURL string) error {
	validURL, err := validateWorkspaceURL(rawURL)
	if err != nil {
		return err
	}

	token, err := promptForToken(cmd)
	if err != nil {
		return fmt.Errorf("reading API token: %w", err)
	}
	if token == "" {
		return errors.New("API token cannot be empty")
	}

	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := saveConfig(Config{URL: validURL}); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if err := saveAPIToken(token); err != nil {
		return fmt.Errorf("saving API token to keychain: %w", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	out := newPrinter(cmd.OutOrStdout())
	out.Println("Configuration saved")
	out.Println("API token saved to keychain")
	out.Println("Config path:", configPath)
	return out.Err()
}
