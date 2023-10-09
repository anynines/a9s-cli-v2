package cmd

import (
	"fmt"

	"github.com/fischerjulian/a8s-demo/demo"
	"github.com/spf13/cobra"
)

var cmdDemo = &cobra.Command{
	Use:   "demo",
	Short: "Create a local demo environment for " + TheA8sPGProductName + " using a Kind Kubernetes cluster and installs",
	Long: `The demo assistent guides through the creation of Kind based Kubernetes cluster, 
	helps to install all necessary prerequisites within the cluster and finally configures and installs
	the operator:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n")
		demo.PrintWarning(" Please select a product by using a sub-command.\n")
		fmt.Printf("Example: a9s demo a8s-pg\n")
	},
}

var cmdDemoA8sPG = &cobra.Command{
	Use:   "a8s-pg",
	Short: "Handling a9s Platform demos.",
	Long: `The demo assistent guides through the creation of a9s Platform demos, 
	helps to install all necessary prerequisites and finally configures and installs
	the chosen product.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteA8sPGDemo()
	},
}

func init() {
	cmdDemo.AddCommand(cmdDemoA8sPG)
	rootCmd.AddCommand(cmdDemo)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runDemoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runDemoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
