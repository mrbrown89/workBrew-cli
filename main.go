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

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(reportCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
