package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var commandYes bool

var commandCmd = &cobra.Command{
	Use:   "command",
	Short: "Create Workbrew brew commands",
}

var commandUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Create a brew update command",
	Run: func(cmd *cobra.Command, args []string) {
		runCreateBrewCommand("update")
	},
}

var commandUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Create a brew upgrade command",
	Run: func(cmd *cobra.Command, args []string) {
		runCreateBrewCommand("upgrade")
	},
}

func confirmBrewCommand(arguments string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This will create a Workbrew command:")
	fmt.Println()
	fmt.Printf("brew %s\n", arguments)
	fmt.Println()
	fmt.Println("Target: Workspace default scope")
	fmt.Print("Continue? [y/N]: ")

	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	return input == "y" || input == "yes"
}

func runCreateBrewCommand(arguments string) {
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

	if !commandYes && !confirmBrewCommand(arguments) {
		fmt.Println("Cancelled")
		return
	}

	result, err := createBrewCommand(config, token, arguments)
	if err != nil {
		fmt.Println("Could not create brew command:", err)
		return
	}

	fmt.Println(result.Message)
}

func init() {
	commandCmd.PersistentFlags().BoolVarP(
		&commandYes,
		"yes",
		"y",
		false,
		"Skip confirmation prompt",
	)

	commandCmd.AddCommand(commandUpdateCmd)
	commandCmd.AddCommand(commandUpgradeCmd)
}
