package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var outputFormat string
var analyticsDevice string
var vulnerabilitiesSeverity string
var vulnerabilitiesDevice string

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
	Short: "Show vulnerable packages",
	Run: func(cmd *cobra.Command, args []string) {
		runVulnerabilitiesReport()
	},
}

var reportAnalyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Show Workbrew analytics",
	Run: func(cmd *cobra.Command, args []string) {
		runAnalyticsReport()
	},
}

var reportBrewfilesCmd = &cobra.Command{
	Use:   "brewfiles",
	Short: "Show Workbrew Brewfiles",
	Run: func(cmd *cobra.Command, args []string) {
		runBrewfilesReport()
	},
}

var reportBrewCommandsCmd = &cobra.Command{
	Use:   "brew-commands",
	Short: "Show Workbrew Brew Commands",
	Run: func(cmd *cobra.Command, args []string) {
		runBrewCommandsReport()
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
	severityFilter := strings.ToLower(strings.TrimSpace(vulnerabilitiesSeverity))
	if severityFilter != "" && severityFilter != "critical" && severityFilter != "high" && severityFilter != "medium" && severityFilter != "low" {
		fmt.Println("Invalid severity. Use critical, high, medium, or low.")
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

	vulnerabilities, err := getVulnerabilities(config, token)
	if err != nil {
		fmt.Println("Could not get vulnerabilities:", err)
		return
	}

	if severityFilter != "" || vulnerabilitiesDevice != "" {
		filteredVulnerabilities := make([]VulnerabilityReport, 0)

		for _, item := range vulnerabilities {
			if vulnerabilitiesDevice != "" {
				matchedDevice := ""

				for _, device := range item.AffectedDevices {
					if strings.EqualFold(device, vulnerabilitiesDevice) {
						matchedDevice = device
						break
					}
				}

				if matchedDevice == "" {
					continue
				}

				item.AffectedDevices = []string{matchedDevice}
			}

			if severityFilter != "" {
				filteredCVEs := make([]CVE, 0)

				for _, cve := range item.Vulnerabilities {
					if strings.ToLower(vulnerabilitySeverity(cve.CVSSScore)) == severityFilter {
						filteredCVEs = append(filteredCVEs, cve)
					}
				}

				item.Vulnerabilities = filteredCVEs
			}

			if len(item.Vulnerabilities) > 0 {
				filteredVulnerabilities = append(filteredVulnerabilities, item)
			}
		}

		vulnerabilities = filteredVulnerabilities
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

	deviceNames := make(map[string]string)
	devices, err := getDevices(config, token)
	if err == nil {
		for _, device := range devices {
			if device.AssignedUser != "" {
				deviceNames[strings.ToLower(device.SerialNumber)] = device.AssignedUser
			}
		}
	}

	fmt.Println("Workbrew Vulnerabilities")
	fmt.Println("------------------------")
	fmt.Println()

	if severityFilter == "" && vulnerabilitiesDevice == "" {
		fmt.Printf("%-30s %-18s %-10s %-10s %s\n", "Package", "Type", "CVEs", "Max CVSS", "Devices")
		fmt.Printf("%-30s %-18s %-10s %-10s %s\n", "-------", "----", "----", "--------", "-------")

		for _, item := range vulnerabilities {
			maxCVSS := "Unscored"
			var highestScore *float64

			for _, cve := range item.Vulnerabilities {
				if cve.CVSSScore != nil && (highestScore == nil || *cve.CVSSScore > *highestScore) {
					highestScore = cve.CVSSScore
				}
			}

			if highestScore != nil {
				maxCVSS = fmt.Sprintf("%.1f", *highestScore)
			}

			fmt.Printf(
				"%-30s %-18s %-10d %-10s %s\n",
				vulnerabilityPackageName(item),
				item.PackageType,
				len(item.Vulnerabilities),
				maxCVSS,
				vulnerabilityDeviceNames(item.AffectedDevices, deviceNames),
			)
		}
	} else {
		fmt.Printf("%-30s %-18s %-20s %-10s %-10s %s\n", "Package", "Type", "CVE", "CVSS", "Severity", "Devices")
		fmt.Printf("%-30s %-18s %-20s %-10s %-10s %s\n", "-------", "----", "---", "----", "--------", "-------")

		for _, item := range vulnerabilities {
			for _, cve := range item.Vulnerabilities {
				fmt.Printf(
					"%-30s %-18s %-20s %-10s %-10s %s\n",
					vulnerabilityPackageName(item),
					item.PackageType,
					cve.CleanID,
					vulnerabilityCVSS(cve.CVSSScore),
					vulnerabilitySeverity(cve.CVSSScore),
					vulnerabilityDeviceNames(item.AffectedDevices, deviceNames),
				)
			}
		}
	}

	printVulnerabilitySummary(vulnerabilities)
}

func vulnerabilityPackageName(item VulnerabilityReport) string {
	if item.DisplayName != "" {
		return item.DisplayName
	}

	return item.Package
}

func vulnerabilityDeviceNames(affectedDevices []string, deviceNames map[string]string) string {
	resolvedDevices := make([]string, 0, len(affectedDevices))

	for _, serialNumber := range affectedDevices {
		deviceName := deviceNames[strings.ToLower(serialNumber)]
		if deviceName == "" {
			deviceName = serialNumber
		}

		resolvedDevices = append(resolvedDevices, deviceName)
	}

	return strings.Join(resolvedDevices, ", ")
}

func vulnerabilityCVSS(score *float64) string {
	if score == nil {
		return "Unscored"
	}

	return fmt.Sprintf("%.1f", *score)
}

func vulnerabilitySeverity(score *float64) string {
	if score == nil || *score < 0.0 || *score > 10.0 {
		return "Unscored"
	}

	if *score >= 9.0 {
		return "Critical"
	}

	if *score >= 7.0 {
		return "High"
	}

	if *score >= 4.0 {
		return "Medium"
	}

	if *score > 0.0 {
		return "Low"
	}

	return "None"
}

func printVulnerabilitySummary(vulnerabilities []VulnerabilityReport) {
	uniqueCVEs := make(map[string]bool)
	affectedPackages := make(map[string]bool)
	affectedDevices := make(map[string]bool)
	totalFindings := 0
	critical := 0
	high := 0
	medium := 0
	low := 0
	noneOrUnscored := 0

	for _, item := range vulnerabilities {
		packageName := item.Package
		if packageName == "" {
			packageName = item.DisplayName
		}

		affectedPackages[item.PackageType+"\x00"+packageName] = true

		for _, device := range item.AffectedDevices {
			affectedDevices[device] = true
		}

		for _, cve := range item.Vulnerabilities {
			totalFindings++

			if cve.CleanID != "" {
				uniqueCVEs[cve.CleanID] = true
			}

			switch vulnerabilitySeverity(cve.CVSSScore) {
			case "Critical":
				critical++
			case "High":
				high++
			case "Medium":
				medium++
			case "Low":
				low++
			default:
				noneOrUnscored++
			}
		}
	}

	fmt.Printf("\nTotal Vulnerability Findings: %d\n", totalFindings)
	fmt.Printf("Total Unique CVEs: %d\n", len(uniqueCVEs))
	fmt.Printf("Total Affected Packages: %d\n", len(affectedPackages))
	fmt.Printf("Total Unique Affected Devices: %d\n", len(affectedDevices))
	fmt.Printf("Critical Vulnerabilities: %d\n", critical)
	fmt.Printf("High Vulnerabilities: %d\n", high)
	fmt.Printf("Medium Vulnerabilities: %d\n", medium)
	fmt.Printf("Low Vulnerabilities: %d\n", low)
	fmt.Printf("None/Unscored Vulnerabilities: %d\n", noneOrUnscored)
}

func runAnalyticsReport() {
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

	analytics, err := getAnalytics(config, token)
	if err != nil {
		fmt.Println("Could not get analytics:", err)
		return
	}

	if analyticsDevice != "" {
		var filteredAnalytics []AnalyticsItem
		search := strings.ToLower(analyticsDevice)

		for _, item := range analytics {
			if strings.Contains(strings.ToLower(item.Device), search) {
				filteredAnalytics = append(filteredAnalytics, item)
			}
		}

		analytics = filteredAnalytics
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(analytics, "", "  ")
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
		return
	}

	sort.Slice(analytics, func(i, j int) bool {
		return analytics[i].Count > analytics[j].Count
	})

	fmt.Println("Workbrew Analytics")
	fmt.Println("------------------")
	fmt.Println()

	fmt.Printf("%-18s %-7s %-25s %s\n", "Device", "Count", "Last Run", "Command")
	fmt.Printf("%-18s %-7s %-25s %s\n", "------", "-----", "--------", "-------")

	for _, item := range analytics {
		fmt.Printf(
			"%-18s %-7d %-25s %s\n",
			item.Device,
			item.Count,
			item.LastRun,
			item.Command,
		)
	}

	fmt.Printf("\nTotal Analytics Items: %d\n", len(analytics))
}

func runBrewfilesReport() {
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

	brewfiles, err := getBrewfiles(config, token)
	if err != nil {
		fmt.Println("Could not get brewfiles:", err)
		return
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(brewfiles, "", "  ")
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
		return
	}

	fmt.Println("Workbrew Brewfiles")
	fmt.Println("------------------")
	fmt.Println()

	fmt.Printf("%-20s %-10s %-8s %-15s %-15s %-10s\n", "Label", "Slug", "Runs", "Started", "Finished", "Devices")
	fmt.Printf("%-20s %-10s %-8s %-15s %-15s %-10s\n", "-----", "----", "----", "-------", "--------", "-------")

	for _, brewfile := range brewfiles {
		fmt.Printf(
			"%-20s %-10s %-8d %-15s %-15s %-10d\n",
			brewfile.Label,
			brewfile.Slug,
			brewfile.RunCount,
			brewfile.StartedAt,
			brewfile.FinishedAt,
			len(brewfile.Devices),
		)
	}

	fmt.Printf("\nTotal Brewfiles: %d\n", len(brewfiles))
}

func runBrewCommandsReport() {
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

	brewCommands, err := getBrewCommands(config, token)
	if err != nil {
		fmt.Println("Could not get brew commands:", err)
		return
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(brewCommands, "", "  ")
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
		return
	}

	sort.Slice(brewCommands, func(i, j int) bool {
		return brewCommands[i].RunCount > brewCommands[j].RunCount
	})

	fmt.Println("Workbrew Brew Commands")
	fmt.Println("----------------------")
	fmt.Println()

	fmt.Printf(
		"%-6s %-8s %-12s %-25s %s\n",
		"Runs",
		"Devices",
		"Updated By",
		"Finished",
		"Command",
	)

	fmt.Printf(
		"%-6s %-8s %-12s %-25s %s\n",
		"----",
		"-------",
		"----------",
		"--------",
		"-------",
	)

	for _, cmd := range brewCommands {
		fmt.Printf(
			"%-6d %-8d %-12s %-25s %s\n",
			cmd.RunCount,
			len(cmd.Devices),
			cmd.LastUpdatedByUser,
			cmd.FinishedAt,
			cmd.Command,
		)
	}

	fmt.Printf("\nTotal Brew Commands: %d\n", len(brewCommands))
}

func init() {
	reportCmd.PersistentFlags().StringVarP(
		&outputFormat,
		"output",
		"o",
		"table",
		"Output format: table or json",
	)

	reportAnalyticsCmd.Flags().StringVar(
		&analyticsDevice,
		"device",
		"",
		"Filter analytics by device serial number",
	)

	reportVulnerabilitiesCmd.Flags().StringVar(
		&vulnerabilitiesSeverity,
		"severity",
		"",
		"Filter vulnerabilities by severity: critical, high, medium, or low",
	)

	reportVulnerabilitiesCmd.Flags().StringVar(
		&vulnerabilitiesDevice,
		"device",
		"",
		"Filter vulnerabilities by device serial number",
	)

	reportCmd.AddCommand(reportSummaryCmd)
	reportCmd.AddCommand(reportOutdatedCmd)
	reportCmd.AddCommand(reportVulnerabilitiesCmd)
	reportCmd.AddCommand(reportAnalyticsCmd)
	reportCmd.AddCommand(reportBrewfilesCmd)
	reportCmd.AddCommand(reportBrewCommandsCmd)
}
