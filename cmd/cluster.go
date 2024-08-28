package cmd

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var TheA8sPGProductName = "a8s Postgres"

var cmdCluster = &cobra.Command{
	Use:   "cluster",
	Short: "Commands related to Kubernetes clusters.",
	Long:  `Commands related to Kubernetes clusters, e.g. printing the local workding directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n")
		makeup.PrintWarning(" Please select a cluster sub-command.\n")
		fmt.Printf(" Examples: \n")
		fmt.Printf("\ta9s cluster pwd\t\tPrint the configured working directory.\n")
		fmt.Printf("\n\n")
	},
}

var cmdClusterPwd = &cobra.Command{
	Use:   "pwd",
	Short: "Print the configured working directory from the ~/.a9s config file.",
	Long:  `Print the configured working directory from the ~/.a9s config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.EstablishConfig()

		fmt.Printf("%s", demo.DemoConfig.WorkingDir)
	},
}

func init() {
	cmdCluster.AddCommand(cmdClusterPwd)
	rootCmd.AddCommand(cmdCluster)
}
