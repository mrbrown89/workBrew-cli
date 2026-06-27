package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.3.1"

func getVersion() string {
	return version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(getVersion())
	},
}
