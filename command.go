package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCommandCmd() *cobra.Command {
	var yes bool

	commandCmd := &cobra.Command{
		Use:   "command",
		Short: "Create Workbrew brew commands",
		Args:  cobra.ArbitraryArgs,
		RunE:  runGroup,
	}

	commandCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	for _, arguments := range []string{"update", "upgrade"} {
		commandCmd.AddCommand(&cobra.Command{
			Use:   arguments,
			Short: fmt.Sprintf("Create a brew %s command", arguments),
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runCreateBrewCommand(cmd, arguments, yes)
			},
		})
	}

	return commandCmd
}

// confirmBrewCommand shows what will be created on stderr and reads the
// answer from the command's stdin.
func confirmBrewCommand(cmd *cobra.Command, arguments string) (bool, error) {
	errOut := newPrinter(cmd.ErrOrStderr())
	errOut.Println("This will create a Workbrew command:")
	errOut.Println()
	errOut.Printf("brew %s\n", arguments)
	errOut.Println()
	errOut.Println("Target: Workspace default scope")
	errOut.Print("Continue? [y/N]: ")

	input, err := readLine(cmd.Context(), cmd.InOrStdin())
	if err != nil {
		return false, fmt.Errorf("reading confirmation: %w", err)
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}

func runCreateBrewCommand(cmd *cobra.Command, arguments string, yes bool) error {
	if _, isTerminal := terminalFd(cmd.InOrStdin()); !yes && !isTerminal {
		return errors.New("stdin is not a terminal; pass --yes to create the command without a confirmation prompt")
	}

	client, err := loadClient()
	if err != nil {
		return err
	}

	if !yes {
		confirmed, err := confirmBrewCommand(cmd, arguments)
		if err != nil {
			return err
		}
		if !confirmed {
			newPrinter(cmd.ErrOrStderr()).Println("Cancelled")
			return nil
		}
	}

	result, err := client.CreateBrewCommand(cmd.Context(), arguments)
	if err != nil {
		return fmt.Errorf("creating brew command: %w", err)
	}

	out := newPrinter(cmd.OutOrStdout())
	out.Println(result.Message)
	return out.Err()
}
