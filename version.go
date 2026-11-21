package main

import (
	"github.com/spf13/cobra"
)

const version = "0.4.0"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			out := newPrinter(cmd.OutOrStdout())
			out.Println(version)
			return out.Err()
		},
	}
}
