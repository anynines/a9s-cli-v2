package cmd

import (
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var cmdPG = &cobra.Command{
	Use:   "pg",
	Short: "Interacting with Postgres.",
	Long:  `Commands for interacting with Postgres.`,
	Run: func(cmd *cobra.Command, args []string) {

		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")

		cmd.Help()
	},
}

var cmdPGApply = &cobra.Command{
	Use:   "apply",
	Short: "Apply an SQL file to the given Postgres instance",
	Long:  "Applying an SQL file will upload the file, use psql to apply it and then delete the file within the target pod.",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// cmdPGApply --sql dump.sql

func init() {
	rootCmd.AddCommand(cmdCreatePG)
}
