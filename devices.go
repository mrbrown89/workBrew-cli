package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Show Workbrew devices",
}

var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Workbrew devices",
	Run: func(cmd *cobra.Command, args []string) {
		runDevicesList()
	},
}

var devicesGetCmd = &cobra.Command{
	Use:   "get <serial>",
	Short: "Show details for a Workbrew device",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runDevicesGet(args[0])
	},
}

var devicesOutputFormat string

func runDevicesList() {
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

	if devicesOutputFormat == "json" {
		output, err := json.MarshalIndent(devices, "", "  ")
		if err != nil {
			fmt.Println("Could not create JSON output:", err)
			return
		}

		fmt.Println(string(output))
		return
	}

	fmt.Println("Workbrew Devices")
	fmt.Println("----------------")
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

func runDevicesGet(serial string) {
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

	for _, device := range devices {
		if device.SerialNumber == serial {
			if devicesOutputFormat == "json" {
				output, err := json.MarshalIndent(device, "", "  ")
				if err != nil {
					fmt.Println("Could not create JSON output:", err)
					return
				}

				fmt.Println(string(output))
				return
			}

			fmt.Println("Workbrew Device")
			fmt.Println("---------------")
			fmt.Println()
			fmt.Println("Serial Number:", device.SerialNumber)
			fmt.Println("Assigned User:", device.AssignedUser)
			fmt.Println("Device Type:  ", device.DeviceType)
			fmt.Println("macOS:        ", stripMacOSPrefix(device.OSVersion))
			fmt.Println("Last Seen:    ", daysAgo(device.LastSeenAt))
			return
		}
	}

	fmt.Println("No device found with serial:", serial)
}

func init() {
	devicesCmd.PersistentFlags().StringVarP(
		&devicesOutputFormat,
		"output",
		"o",
		"table",
		"Output format: table or json",
	)

	devicesCmd.AddCommand(devicesListCmd)
	devicesCmd.AddCommand(devicesGetCmd)
}
