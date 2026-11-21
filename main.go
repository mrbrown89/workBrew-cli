package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
)

// exitInterrupted is the conventional exit status for SIGINT (128 + 2).
const exitInterrupted = 130

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// After the first Ctrl-C cancels ctx, restore the default handler so a
	// second Ctrl-C kills the process even if something is stuck.
	go func() {
		<-ctx.Done()
		stop()
	}()

	rootCmd := newRootCmd()
	if cmd, err := rootCmd.ExecuteContextC(ctx); err != nil {
		if ctx.Err() != nil {
			newPrinter(rootCmd.ErrOrStderr()).Println("interrupted")
			return exitInterrupted
		}
		printError(rootCmd.ErrOrStderr(), cmd, err)
		return 1
	}

	return 0
}

// printError writes err to w. Unknown subcommands don't print usage, so they
// get a pointer to --help instead: for the root, cobra fails while finding the
// command, before its usage handling; for groups, runGroup returns an
// unknownSubcommandError. "Did you mean" suggestions are already part of err.
// Other unknown-command errors (from cobra.NoArgs) print usage.
func printError(w io.Writer, cmd *cobra.Command, err error) {
	// Write errors are ignored: this is already the error path.
	p := newPrinter(w)
	p.Println("Error:", err)

	var unknown *unknownSubcommandError
	switch {
	case errors.As(err, &unknown):
		p.Printf("Run '%s --help' for usage.\n", unknown.cmd.CommandPath())
	case cmd != nil && !cmd.HasParent() && strings.HasPrefix(err.Error(), "unknown command "):
		p.Printf("Run '%s --help' for usage.\n", cmd.CommandPath())
	}
}

// newRootCmd builds a fresh command tree. Each call has its own flag state,
// so tests can run commands independently.
func newRootCmd() *cobra.Command {
	// No Run: running the bare command shows help.
	rootCmd := &cobra.Command{
		Use:           "workbrew-cli",
		Short:         "A CLI for querying Workbrew",
		Version:       version,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		newVersionCmd(),
		newReportCmd(),
		newSetupCmd(),
		newAuthCmd(),
		newDevicesCmd(),
		newCommandCmd(),
	)

	return rootCmd
}
