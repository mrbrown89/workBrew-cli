package main

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newDevicesCmd() *cobra.Command {
	var format outputFormat
	var showApps bool

	devicesCmd := &cobra.Command{
		Use:   "devices",
		Short: "Show Workbrew devices",
		Args:  cobra.ArbitraryArgs,
		RunE:  runGroup,
	}
	addOutputFlag(devicesCmd, &format)

	devicesGetCmd := &cobra.Command{
		Use:   "get <serial-or-user>",
		Short: "Show details for a Workbrew device",
		Long: `Show details for a Workbrew device.

The search value may be a serial number or an assigned user. An exact
(case-insensitive) match on either is preferred; otherwise a partial match
is used. If more than one device matches, the matches are listed on stderr
and the command exits with status 1.

Examples:
  workbrew-cli devices get ABC1234567
  workbrew-cli devices get alex
  workbrew-cli devices get "Alex Example"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runDevicesGet(cmd, args[0], format, showApps)
		},
	}
	devicesGetCmd.Flags().BoolVar(&showApps, "apps", false, "Show installed formulae and casks")

	devicesCmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List Workbrew devices",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				return runDevicesList(cmd, format)
			},
		},
		devicesGetCmd,
	)

	return devicesCmd
}

func appStatus(outdated bool) string {
	if outdated {
		return " [OUTDATED]"
	}

	return ""
}

// runDevicesList backs both `devices list` and `report summary`.
func runDevicesList(cmd *cobra.Command, format outputFormat) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	devices, err := client.Devices(cmd.Context())
	if err != nil {
		return fmt.Errorf("getting devices: %w", err)
	}

	out := newPrinter(cmd.OutOrStdout())
	if format == formatJSON {
		return printJSON(out, devices)
	}

	out.Heading("Workbrew Devices")
	writeDeviceTable(out, devices, time.Now())

	out.Printf("\nTotal Devices: %d\n", len(devices))
	return out.Err()
}

func writeDeviceTable(p *printer, devices []workbrew.Device, now time.Time) {
	table := newTable(p)
	table.Row("Serial Number", "Assigned User", "macOS", "Seen")
	table.Row("-------------", "-------------", "-----", "----")

	for _, device := range devices {
		table.Row(
			device.SerialNumber,
			device.AssignedUser,
			stripMacOSPrefix(device.OSVersion),
			daysAgo(device.LastSeenAt, now),
		)
	}

	table.Flush()
}

// matchDevices returns the devices whose serial number or assigned user
// equals query (case-insensitive). If there are none, it falls back to
// devices where either field contains query.
func matchDevices(devices []workbrew.Device, query string) []workbrew.Device {
	var exact, partial []workbrew.Device
	search := strings.ToLower(query)

	for _, device := range devices {
		switch {
		case strings.EqualFold(device.SerialNumber, query), strings.EqualFold(device.AssignedUser, query):
			exact = append(exact, device)
		case strings.Contains(strings.ToLower(device.SerialNumber), search),
			strings.Contains(strings.ToLower(device.AssignedUser), search):
			partial = append(partial, device)
		}
	}

	if len(exact) > 0 {
		return exact
	}

	return partial
}

// deviceDetails is the JSON shape of `devices get`. Formulae and Casks are
// only set (and then never null) when --apps is given.
type deviceDetails struct {
	workbrew.Device
	FormulaCount  int                `json:"formula_count"`
	CaskCount     int                `json:"cask_count"`
	OutdatedCount int                `json:"outdated_count"`
	Formulae      []workbrew.Formula `json:"formulae,omitzero"`
	Casks         []workbrew.Cask    `json:"casks,omitzero"`
}

// compareApps orders outdated apps first, then by name.
func compareApps(aOutdated bool, aName string, bOutdated bool, bName string) int {
	if aOutdated != bOutdated {
		if aOutdated {
			return -1
		}
		return 1
	}

	return cmp.Compare(aName, bName)
}

func buildDeviceDetails(
	device workbrew.Device,
	formulae []workbrew.Formula,
	casks []workbrew.Cask,
	withApps bool,
) deviceDetails {
	apps := appsForDevice(device.SerialNumber, formulae, casks)

	details := deviceDetails{
		Device:        device,
		FormulaCount:  len(apps.formulae),
		CaskCount:     len(apps.casks),
		OutdatedCount: apps.outdated,
	}

	if withApps {
		details.Formulae = apps.formulae
		slices.SortStableFunc(details.Formulae, func(a, b workbrew.Formula) int {
			return compareApps(a.Outdated, a.Name, b.Outdated, b.Name)
		})

		details.Casks = apps.casks
		slices.SortStableFunc(details.Casks, func(a, b workbrew.Cask) int {
			return compareApps(a.Outdated, a.Name, b.Outdated, b.Name)
		})
	}

	return details
}

func runDevicesGet(cmd *cobra.Command, query string, format outputFormat, showApps bool) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("device search value cannot be empty")
	}

	client, err := loadClient()
	if err != nil {
		return err
	}

	var (
		devices  []workbrew.Device
		formulae []workbrew.Formula
		casks    []workbrew.Cask
	)

	group, ctx := errgroup.WithContext(cmd.Context())
	group.Go(func() (err error) {
		if devices, err = client.Devices(ctx); err != nil {
			return fmt.Errorf("getting devices: %w", err)
		}
		return nil
	})
	group.Go(func() (err error) {
		if formulae, err = client.Formulae(ctx); err != nil {
			return fmt.Errorf("getting formulae: %w", err)
		}
		return nil
	})
	group.Go(func() (err error) {
		if casks, err = client.Casks(ctx); err != nil {
			return fmt.Errorf("getting casks: %w", err)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		return err
	}

	now := time.Now()

	matches := matchDevices(devices, query)
	switch len(matches) {
	case 0:
		return fmt.Errorf("no device found matching %q", query)
	case 1:
		// Show it below.
	default:
		writeDeviceTable(newPrinter(cmd.ErrOrStderr()), matches, now)
		return fmt.Errorf("%q matches %d devices; use a full serial number or user name", query, len(matches))
	}

	details := buildDeviceDetails(matches[0], formulae, casks, showApps)

	out := newPrinter(cmd.OutOrStdout())
	if format == formatJSON {
		return printJSON(out, details)
	}

	out.Heading("Workbrew Device")

	table := newTable(out)
	table.Row("Serial Number:", details.SerialNumber)
	table.Row("Assigned User:", details.AssignedUser)
	table.Row("Device Type:", details.DeviceType)
	table.Row("macOS:", stripMacOSPrefix(details.OSVersion))
	table.Row("Last Seen:", daysAgo(details.LastSeenAt, now))
	table.Row("", "") // blank line that keeps the columns aligned
	table.Row("Formulae:", details.FormulaCount)
	table.Row("Casks:", details.CaskCount)
	table.Row("Outdated:", details.OutdatedCount)
	table.Flush()

	if !showApps {
		return out.Err()
	}

	out.Println()
	out.Heading("Formulae")
	for _, formula := range details.Formulae {
		out.Println(formula.Name + appStatus(formula.Outdated))
	}

	out.Println()
	out.Heading("Casks")
	for _, cask := range details.Casks {
		out.Println(cask.Name + appStatus(cask.Outdated))
	}

	return out.Err()
}
