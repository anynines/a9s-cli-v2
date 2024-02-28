package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Helps to distinguish individual builds
var BuildTimestamp string

// The CLI version relates to the changelog and readme
var CliVersion string

// Helps to trace a particular build to its version in the git repo
var LastCommit string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of the a9s cli",
	Long:  `All software has versions. This is a9s cli's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("a9s cli version: %s\n", CliVersion)
		fmt.Printf("Built at %s\n", BuildTimestamp)
		fmt.Printf("Based on last git commit: %s\n", LastCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
