package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// unknownSubcommandError is returned when a command group (devices, report,
// command, auth) is given a subcommand it doesn't have.
type unknownSubcommandError struct {
	cmd  *cobra.Command
	name string
}

func (e *unknownSubcommandError) Error() string {
	var msg strings.Builder
	fmt.Fprintf(&msg, "unknown command %q for %q", e.name, e.cmd.CommandPath())

	// Same layout as cobra's own suggestions for the root command.
	if e.cmd.DisableSuggestions {
		return msg.String()
	}
	if suggestions := e.cmd.SuggestionsFor(e.name); len(suggestions) > 0 {
		msg.WriteString("\n\nDid you mean this?\n")
		for _, suggestion := range suggestions {
			fmt.Fprintf(&msg, "\t%s\n", suggestion)
		}
	}

	return msg.String()
}

// runGroup is the RunE for command groups: with no arguments it shows the
// group's help, otherwise the first argument is an unknown subcommand.
// Cobra would otherwise show help and exit 0 for both.
func runGroup(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if len(args) == 0 {
		return cmd.Help()
	}

	// Cobra only applies its default edit distance when it builds suggestions
	// for the root; SuggestionsFor on its own would use 0 (prefix matches only).
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}

	return &unknownSubcommandError{cmd: cmd, name: args[0]}
}
