package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const CliName = "a9s"

var rootCmd = &cobra.Command{
	Use:   CliName,
	Short: CliName + " is a tool to help you with using modules of the a9s Platform",
	Long:  `A tool to make the use of a9s Platform modules more enjoyable.`,
	Run: func(cmd *cobra.Command, args []string) {

		cmd.Help()
		// // Do Stuff Here
		// demo.PrintWelcomeScreen()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
