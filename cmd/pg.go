package cmd

import (
	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/a9s-cli-v2/pg"
	"github.com/spf13/cobra"
)

var NoDeleteSQLFile bool
var UnattendedMode bool
var ApplySQLStatement string

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
		pgm := pg.NewPgManager("")

		if ApplySQLStatement != "" {
			pgm.ApplySQLStatementToPGServiceInstance(UnattendedMode, demo.A8sPGServiceInstance.Namespace, demo.A8sPGServiceInstance.Name, ApplySQLStatement)
		} else if pg.SQLFilename != "" {
			pgm.ApplySQLFileToPGServiceInstance(UnattendedMode, demo.A8sPGServiceInstance.Namespace, demo.A8sPGServiceInstance.Name, pg.SQLFilename, NoDeleteSQLFile)
		} else {
			makeup.ExitDueToFatalError(nil, "Please supply either --sql with an SQL statement or --file with a path to a sql file.")
		}
	},
}

func init() {
	cmdPG.AddCommand(cmdPGApply)

	//TODO Make service-instance mandatory param without default value
	//TODO Make file mandatory param without default value
	cmdPG.PersistentFlags().StringVarP(&demo.A8sPGServiceInstance.Namespace, "namespace", "n", "default", "namespace of the pg service instance.")
	cmdPG.PersistentFlags().BoolVar(&NoDeleteSQLFile, "no-delete", false, "if set the uploaded SQL file won't be deleted after applying it.")

	cmdPG.PersistentFlags().StringVarP(&demo.A8sPGServiceInstance.Name, "service-instance", "i", "", "name of the pg service instance.")
	cmdPG.MarkFlagRequired("service-instance")

	cmdPGApply.PersistentFlags().StringVarP(&pg.SQLFilename, "file", "f", "", "path to the SQL file to be applied to the given pg service instance.")
	cmdPGApply.PersistentFlags().StringVar(&ApplySQLStatement, "sql", "", "applies the given SQL statement to the given pg service instance.")

	// It's either --file or --sql but never both
	cmdPGApply.MarkFlagsMutuallyExclusive("file", "sql")

	cmdPG.PersistentFlags().BoolVarP(&UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")

	rootCmd.AddCommand(cmdPG)
}
