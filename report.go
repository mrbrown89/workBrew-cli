package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var outputFormat string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show Workbrew reporting information",
}

var reportSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show a summary report",
	Run: func(cmd *cobra.Command, args []string) {
		runSummaryReport()
	},
}

func runSummaryReport() {
	if outputFormat != "table" && outputFormat != "json" {
		fmt.Println("Invalid output format. Use table or json.")
		return
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Could not load config. Run setup first.")
		return
	}

	fmt.Println("Summary report not implemented yet")
	fmt.Println("Workspace URL:", config.URL)
	fmt.Println("Output format:", outputFormat)
}

func init() {
	reportSummaryCmd.Flags().StringVarP(
		&outputFormat,
		"output",
		"o",
		"table",
		"Output format: table or json",
	)

	reportCmd.AddCommand(reportSummaryCmd)
}
