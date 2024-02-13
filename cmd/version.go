package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var BuildTimestamp string
var CliVersion string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of the a9s cli",
	Long:  `All software has versions. This is a9s cli's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("a9s cli version: %s\n", CliVersion)
		fmt.Printf("Built at %s\n", BuildTimestamp)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
