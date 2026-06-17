package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "workbrew",
		Short: "A CLI for querying Workbrew",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Workbrew CLI")
		},
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(getVersion())
		},
	}

	var reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Show Workbrew reporting information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Report command coming soon")
		},
	}

	rootCmd.AddCommand(versionCmd)
        rootCmd.AddCommand(reportCmd)

	rootCmd.Execute()
}
