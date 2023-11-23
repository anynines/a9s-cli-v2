package cmd

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/spf13/cobra"
)

var TheA8sPGProductName = "a8s Postgres"

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
		demo.PrintWarning(" Please select a demo sub-command.\n")
		fmt.Printf(" Examples: \n")
		fmt.Printf("\ta9s demo pwd\t\tPrint the configured working directory.\n")
		fmt.Printf("\ta9s demo a8s-pg\t\tExecute the a8s-pg product demo.")
		fmt.Printf("\n\n")
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

var cmdDemoPwd = &cobra.Command{
	Use:   "pwd",
	Short: "Print the configured working directory for demos.",
	Long:  `Print the configured working directory for demos`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.EstablishConfig()

		fmt.Printf("\n%s\n\n", demo.DemoConfig.WorkingDir)
	},
}

func ExecuteA8sPGDemo() {
	demo.PrintWelcomeScreen()

	demo.EstablishConfig()

	demo.CheckPrerequisites()

	demo.WaitForUser()

	demo.CheckoutDeploymentGitRepository()

	if demo.CountPodsInDemoNamespace() == 0 {
		demo.PrintCheckmark("Kubernetes cluster has no pods in " + demo.GetConfig().DemoSpace + " namespace.")
	}

	demo.EstablishBackupStoreCredentials()

	demo.ApplyCertManagerManifests()

	demo.ApplyA8sManifests()

	demo.WaitForSystemToBecomeReady()

	demo.PrintDemoSummary()
}

func init() {
	cmdDemo.AddCommand(cmdDemoA8sPG)
	cmdDemo.AddCommand(cmdDemoPwd)
	rootCmd.AddCommand(cmdDemo)
	rootCmd.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "minikube", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" an \"kind\"")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runDemoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runDemoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
