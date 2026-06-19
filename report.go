package main

import (
	"encoding/json"
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

var reportOutdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "Show outdated formulae and casks",
	Run: func(cmd *cobra.Command, args []string) {
		runOutdatedReport()
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
	fmt.Println()

	fmt.Printf("%-18s %-30s %-14s %-18s\n", "Serial Number", "Assigned User", "macOS", "Seen")
	fmt.Printf("%-18s %-30s %-14s %-18s\n", "-------------", "-------------", "-----", "---------")

	for _, device := range devices {
		fmt.Printf(
			"%-18s %-30s %-14s %-18s\n",
			device.SerialNumber,
			device.AssignedUser,
			stripMacOSPrefix(device.OSVersion),
			daysAgo(device.LastSeenAt),
		)
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(devices, "", "  ")
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
		return
	}

	fmt.Printf("\nTotal Devices: %d\n", len(devices))
}

func runOutdatedReport() {
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

	formulae, err := getFormulae(config, token)
	if err != nil {
		fmt.Println("Could not get formulae:", err)
		return
	}

	casks, err := getCasks(config, token)
	if err != nil {
		fmt.Println("Could not get casks:", err)
		return
	}

	fmt.Println("Workbrew Outdated Apps")
	fmt.Println("----------------------")
	fmt.Println()

	fmt.Printf("%-8s %-35s %-18s %-8s\n", "Type", "Name", "Version", "Devices")
	fmt.Printf("%-8s %-35s %-18s %-8s\n", "----", "----", "-------", "-------")

	count := 0

	for _, formula := range formulae {
		if !formula.Outdated {
			continue
		}

		fmt.Printf(
			"%-8s %-35s %-18s %-8d\n",
			"Formula",
			formula.Name,
			formula.HomebrewCoreVersion,
			len(formula.Devices),
		)

		count++
	}

	for _, cask := range casks {
		if !cask.Outdated {
			continue
		}

		fmt.Printf(
			"%-8s %-35s %-18s %-8d\n",
			"Cask",
			cask.Name,
			cask.HomebrewCaskVersion,
			len(cask.Devices),
		)

		count++
	}

	fmt.Printf("\nTotal Outdated Packages: %d\n", count)
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
	reportCmd.AddCommand(reportOutdatedCmd)

}
