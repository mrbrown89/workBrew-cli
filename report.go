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

	token, err := loadAPIToken()
	if err != nil {
		fmt.Println("No API token found. Run setup first.")
		return
	}

	devices, err := getDevices(config, token)
	if err != nil {
		fmt.Println("Could not get devices:", err)
		return
	}

	fmt.Println("Workbrew Device Summary")
	fmt.Println("-----------------------")
	fmt.Printf("Total Devices: %d\n", len(devices))
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
