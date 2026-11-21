package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var format outputFormat
	var analyticsDevice string

	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Show Workbrew reporting information",
		Args:  cobra.ArbitraryArgs,
		RunE:  runGroup,
	}
	addOutputFlag(reportCmd, &format)

	analyticsCmd := &cobra.Command{
		Use:   "analytics",
		Short: "Show Workbrew analytics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runAnalyticsReport(cmd, format, analyticsDevice)
		},
	}
	analyticsCmd.Flags().StringVar(
		&analyticsDevice,
		"device",
		"",
		"Show only this device (exact serial number, case-insensitive)",
	)

	reportCmd.AddCommand(
		&cobra.Command{
			Use:   "summary",
			Short: "List Workbrew devices (alias for `devices list`)",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runDevicesList(cmd, format)
			},
		},
		&cobra.Command{
			Use:   "outdated",
			Short: "Show outdated formulae and casks",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runOutdatedReport(cmd, format)
			},
		},
		newVulnerabilitiesCmd(&format),
		analyticsCmd,
		&cobra.Command{
			Use:   "brewfiles",
			Short: "Show Workbrew Brewfiles",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runBrewfilesReport(cmd, format)
			},
		},
		&cobra.Command{
			Use:   "brew-commands",
			Short: "Show Workbrew Brew Commands",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runBrewCommandsReport(cmd, format)
			},
		},
	)

	return reportCmd
}

func runOutdatedReport(cmd *cobra.Command, format outputFormat) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	formulae, err := client.Formulae(cmd.Context())
	if err != nil {
		return fmt.Errorf("getting formulae: %w", err)
	}

	casks, err := client.Casks(cmd.Context())
	if err != nil {
		return fmt.Errorf("getting casks: %w", err)
	}

	outdatedFormulae := make([]workbrew.Formula, 0)
	for _, formula := range formulae {
		if formula.Outdated {
			outdatedFormulae = append(outdatedFormulae, formula)
		}
	}

	outdatedCasks := make([]workbrew.Cask, 0)
	for _, cask := range casks {
		if cask.Outdated {
			outdatedCasks = append(outdatedCasks, cask)
		}
	}

	out := newPrinter(cmd.OutOrStdout())
	if format == formatJSON {
		return printJSON(out, map[string]any{
			"formulae": outdatedFormulae,
			"casks":    outdatedCasks,
		})
	}

	out.Heading("Workbrew Outdated Apps")

	table := newTable(out)
	table.Row("Type", "Name", "Version", "Devices")
	table.Row("----", "----", "-------", "-------")
	for _, formula := range outdatedFormulae {
		table.Row("Formula", formula.Name, formula.HomebrewCoreVersion, len(formula.Devices))
	}
	for _, cask := range outdatedCasks {
		table.Row("Cask", cask.Name, cask.HomebrewCaskVersion, len(cask.Devices))
	}
	table.Flush()

	out.Printf("\nTotal Outdated Packages: %d\n", len(outdatedFormulae)+len(outdatedCasks))
	return out.Err()
}

// filterAnalytics keeps the items for device (exact serial, case-insensitive).
// An empty device keeps everything.
func filterAnalytics(analytics []workbrew.AnalyticsItem, device string) []workbrew.AnalyticsItem {
	if device == "" {
		return analytics
	}

	filtered := make([]workbrew.AnalyticsItem, 0)
	for _, item := range analytics {
		if strings.EqualFold(item.Device, device) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// sortAnalytics orders by count (highest first), then command, then device.
func sortAnalytics(analytics []workbrew.AnalyticsItem) {
	slices.SortStableFunc(analytics, func(a, b workbrew.AnalyticsItem) int {
		return cmp.Or(
			cmp.Compare(b.Count, a.Count),
			cmp.Compare(a.Command, b.Command),
			cmp.Compare(a.Device, b.Device),
		)
	})
}

func runAnalyticsReport(cmd *cobra.Command, format outputFormat, device string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	analytics, err := client.Analytics(cmd.Context())
	if err != nil {
		return fmt.Errorf("getting analytics: %w", err)
	}

	analytics = filterAnalytics(analytics, strings.TrimSpace(device))
	sortAnalytics(analytics)

	out := newPrinter(cmd.OutOrStdout())
	if format == formatJSON {
		return printJSON(out, analytics)
	}

	out.Heading("Workbrew Analytics")

	table := newTable(out)
	table.Row("Device", "Count", "Last Run", "Command")
	table.Row("------", "-----", "--------", "-------")
	for _, item := range analytics {
		table.Row(item.Device, item.Count, item.LastRun, item.Command)
	}
	table.Flush()

	out.Printf("\nTotal Analytics Items: %d\n", len(analytics))
	return out.Err()
}

func runBrewfilesReport(cmd *cobra.Command, format outputFormat) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	brewfiles, err := client.Brewfiles(cmd.Context())
	if err != nil {
		return fmt.Errorf("getting brewfiles: %w", err)
	}

	out := newPrinter(cmd.OutOrStdout())
	if format == formatJSON {
		return printJSON(out, brewfiles)
	}

	out.Heading("Workbrew Brewfiles")

	table := newTable(out)
	table.Row("Label", "Slug", "Runs", "Started", "Finished", "Devices")
	table.Row("-----", "----", "----", "-------", "--------", "-------")
	for _, brewfile := range brewfiles {
		table.Row(
			brewfile.Label,
			brewfile.Slug,
			brewfile.RunCount,
			brewfile.StartedAt,
			brewfile.FinishedAt,
			len(brewfile.Devices),
		)
	}
	table.Flush()

	out.Printf("\nTotal Brewfiles: %d\n", len(brewfiles))
	return out.Err()
}

// sortBrewCommands orders by run count (highest first), then command.
func sortBrewCommands(brewCommands []workbrew.BrewCommand) {
	slices.SortStableFunc(brewCommands, func(a, b workbrew.BrewCommand) int {
		return cmp.Or(
			cmp.Compare(b.RunCount, a.RunCount),
			cmp.Compare(a.Command, b.Command),
		)
	})
}

func runBrewCommandsReport(cmd *cobra.Command, format outputFormat) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	brewCommands, err := client.BrewCommands(cmd.Context())
	if err != nil {
		return fmt.Errorf("getting brew commands: %w", err)
	}

	sortBrewCommands(brewCommands)

	out := newPrinter(cmd.OutOrStdout())
	if format == formatJSON {
		return printJSON(out, brewCommands)
	}

	out.Heading("Workbrew Brew Commands")

	table := newTable(out)
	table.Row("Runs", "Devices", "Updated By", "Finished", "Command")
	table.Row("----", "-------", "----------", "--------", "-------")
	for _, brewCommand := range brewCommands {
		table.Row(
			brewCommand.RunCount,
			len(brewCommand.Devices),
			brewCommand.LastUpdatedByUser,
			brewCommand.FinishedAt,
			brewCommand.Command,
		)
	}
	table.Flush()

	out.Printf("\nTotal Brew Commands: %d\n", len(brewCommands))
	return out.Err()
}
