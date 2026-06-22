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

var reportVulnerabilitiesCmd = &cobra.Command{
	Use:   "vulnerabilities",
	Short: "Show vulnerable formulae",
	Run: func(cmd *cobra.Command, args []string) {
		runVulnerabilitiesReport()
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

	if outputFormat == "json" {
		output, err := json.MarshalIndent(devices, "", "  ")
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
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

	if outputFormat == "json" {
		var outdatedFormulae []Formula
		var outdatedCasks []Cask

		for _, formula := range formulae {
			if formula.Outdated {
				outdatedFormulae = append(outdatedFormulae, formula)
			}
		}

		for _, cask := range casks {
			if cask.Outdated {
				outdatedCasks = append(outdatedCasks, cask)
			}
		}

		output, err := json.MarshalIndent(
			map[string]any{
				"formulae": outdatedFormulae,
				"casks":    outdatedCasks,
			},
			"",
			"  ",
		)
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
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

func runVulnerabilitiesReport() {
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

	vulnerabilities, err := getVulnerabilities(config, token)
	if err != nil {
		fmt.Println("Could not get vulnerabilities:", err)
		return
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(vulnerabilities, "", "  ")
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
		return
	}

	fmt.Println("Workbrew Vulnerabilities")
	fmt.Println("------------------------")
	fmt.Println()

	fmt.Printf("%-30s %-10s %-8s %-8s\n", "Formula", "CVEs", "Max CVSS", "Devices")
	fmt.Printf("%-30s %-10s %-8s %-8s\n", "-------", "----", "--------", "-------")

	totalCVEs := 0

	for _, item := range vulnerabilities {
		maxCVSS := 0.0

		for _, cve := range item.Vulnerabilities {
			totalCVEs++

			if cve.CVSSScore > maxCVSS {
				maxCVSS = cve.CVSSScore
			}
		}

		fmt.Printf(
			"%-30s %-10d %-8.1f %-8d\n",
			item.Formula,
			len(item.Vulnerabilities),
			maxCVSS,
			len(item.OutdatedDevices),
		)
	}

	fmt.Printf("\nTotal Vulnerable Formulae: %d\n", len(vulnerabilities))
	fmt.Printf("Total CVEs: %d\n", totalCVEs)
}

func init() {
	reportCmd.PersistentFlags().StringVarP(
		&outputFormat,
		"output",
		"o",
		"table",
		"Output format: table or json",
	)

	reportCmd.AddCommand(reportSummaryCmd)
	reportCmd.AddCommand(reportOutdatedCmd)
	reportCmd.AddCommand(reportVulnerabilitiesCmd)
}
