package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// severity is the CVSS v3 qualitative rating of a score.
type severity string

const (
	severityCritical severity = "Critical"
	severityHigh     severity = "High"
	severityMedium   severity = "Medium"
	severityLow      severity = "Low"
	severityNone     severity = "None"
	severityUnscored severity = "Unscored"
)

// parseSeverity parses the --severity flag (case-insensitive). An empty
// value means no filter.
func parseSeverity(value string) (severity, error) {
	switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
	case "":
		return "", nil
	case "critical":
		return severityCritical, nil
	case "high":
		return severityHigh, nil
	case "medium":
		return severityMedium, nil
	case "low":
		return severityLow, nil
	default:
		return "", fmt.Errorf("invalid severity %q; use critical, high, medium, or low", normalized)
	}
}

func newVulnerabilitiesCmd(format *outputFormat) *cobra.Command {
	var severity, device string

	vulnerabilitiesCmd := &cobra.Command{
		Use:   "vulnerabilities",
		Short: "Show vulnerable packages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runVulnerabilitiesReport(cmd, *format, severity, device)
		},
	}

	vulnerabilitiesCmd.Flags().StringVar(
		&severity,
		"severity",
		"",
		"Filter vulnerabilities by severity: critical, high, medium, or low",
	)
	vulnerabilitiesCmd.Flags().StringVar(
		&device,
		"device",
		"",
		"Show only this device (exact serial number, case-insensitive)",
	)

	return vulnerabilitiesCmd
}

// filterVulnerabilities keeps the CVEs matching severity on packages affecting device (exact serial, case-insensitive).
// Empty filters match everything. Packages left with no CVEs are dropped,
// and with a device filter AffectedDevices is narrowed to that device.
func filterVulnerabilities(items []workbrew.VulnerabilityReport, wantSeverity severity, device string) []workbrew.VulnerabilityReport {
	if wantSeverity == "" && device == "" {
		return items
	}

	filtered := make([]workbrew.VulnerabilityReport, 0)

	for _, item := range items {
		if device != "" {
			match := slices.IndexFunc(item.AffectedDevices, func(affected string) bool {
				return strings.EqualFold(affected, device)
			})
			if match < 0 {
				continue
			}

			item.AffectedDevices = []string{item.AffectedDevices[match]}
		}

		if wantSeverity != "" {
			filteredCVEs := make([]workbrew.CVE, 0)
			for _, cve := range item.Vulnerabilities {
				if vulnerabilitySeverity(cve.CVSSScore) == wantSeverity {
					filteredCVEs = append(filteredCVEs, cve)
				}
			}

			item.Vulnerabilities = filteredCVEs
		}

		if len(item.Vulnerabilities) > 0 {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func runVulnerabilitiesReport(cmd *cobra.Command, format outputFormat, severityFlag, device string) error {
	wantSeverity, err := parseSeverity(severityFlag)
	if err != nil {
		return err
	}
	device = strings.TrimSpace(device)

	client, err := loadClient()
	if err != nil {
		return err
	}

	// The table shows device names, so fetch devices alongside the report.
	// A failed device lookup only costs the names, so it is not returned
	// through the group (which would cancel the report request).
	var (
		vulnerabilities []workbrew.VulnerabilityReport
		devices         []workbrew.Device
		devicesErr      error
	)

	group, ctx := errgroup.WithContext(cmd.Context())
	group.Go(func() (err error) {
		if vulnerabilities, err = client.Vulnerabilities(ctx); err != nil {
			return fmt.Errorf("getting vulnerabilities: %w", err)
		}
		return nil
	})
	if format == formatTable {
		group.Go(func() error {
			devices, devicesErr = client.Devices(ctx)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}

	vulnerabilities = filterVulnerabilities(vulnerabilities, wantSeverity, device)

	out := newPrinter(cmd.OutOrStdout())
	if format == formatJSON {
		return printJSON(out, vulnerabilities)
	}

	deviceNames := make(map[string]string)
	if devicesErr != nil {
		if err := cmd.Context().Err(); err != nil {
			return err
		}
		newPrinter(cmd.ErrOrStderr()).Printf("warning: could not look up device names, showing serial numbers: %v\n", devicesErr)
	}
	for _, device := range devices {
		if device.AssignedUser != "" {
			deviceNames[strings.ToLower(device.SerialNumber)] = device.AssignedUser
		}
	}

	out.Heading("Workbrew Vulnerabilities")

	table := newTable(out)
	if wantSeverity == "" && device == "" {
		table.Row("Package", "Type", "CVEs", "Max CVSS", "Devices")
		table.Row("-------", "----", "----", "--------", "-------")

		for _, item := range vulnerabilities {
			var highestScore *float64
			for _, cve := range item.Vulnerabilities {
				if cve.CVSSScore != nil && (highestScore == nil || *cve.CVSSScore > *highestScore) {
					highestScore = cve.CVSSScore
				}
			}

			table.Row(
				vulnerabilityPackageName(item),
				item.PackageType,
				len(item.Vulnerabilities),
				vulnerabilityCVSS(highestScore),
				vulnerabilityDeviceNames(item.AffectedDevices, deviceNames),
			)
		}
	} else {
		table.Row("Package", "Type", "CVE", "CVSS", "Severity", "Devices")
		table.Row("-------", "----", "---", "----", "--------", "-------")

		for _, item := range vulnerabilities {
			for _, cve := range item.Vulnerabilities {
				table.Row(
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
	table.Flush()

	printVulnerabilitySummary(out, vulnerabilities)
	return out.Err()
}

func vulnerabilityPackageName(item workbrew.VulnerabilityReport) string {
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

// vulnerabilitySeverity rates a CVSS v3 score. Missing scores and scores
// outside 0-10 are Unscored.
func vulnerabilitySeverity(score *float64) severity {
	switch {
	case score == nil || *score < 0.0 || *score > 10.0:
		return severityUnscored
	case *score >= 9.0:
		return severityCritical
	case *score >= 7.0:
		return severityHigh
	case *score >= 4.0:
		return severityMedium
	case *score > 0.0:
		return severityLow
	default:
		return severityNone
	}
}

func printVulnerabilitySummary(p *printer, vulnerabilities []workbrew.VulnerabilityReport) {
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
			case severityCritical:
				critical++
			case severityHigh:
				high++
			case severityMedium:
				medium++
			case severityLow:
				low++
			default:
				noneOrUnscored++
			}
		}
	}

	p.Printf("\nTotal Vulnerability Findings: %d\n", totalFindings)
	p.Printf("Total Unique CVEs: %d\n", len(uniqueCVEs))
	p.Printf("Total Affected Packages: %d\n", len(affectedPackages))
	p.Printf("Total Unique Affected Devices: %d\n", len(affectedDevices))
	p.Printf("Critical Vulnerabilities: %d\n", critical)
	p.Printf("High Vulnerabilities: %d\n", high)
	p.Printf("Medium Vulnerabilities: %d\n", medium)
	p.Printf("Low Vulnerabilities: %d\n", low)
	p.Printf("None/Unscored Vulnerabilities: %d\n", noneOrUnscored)
}
