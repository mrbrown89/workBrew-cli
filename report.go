package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show Workbrew reporting information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Report command coming soon")
	},
}
