package cmd

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var TheA8sPGProductName = "a8s Postgres"

var cmdCluster = &cobra.Command{
	Use:   "cluster",
	Short: "Commands related to a9s Platform demos.",
	Long:  `Commands related to a9s Platform demos, e.g. printing the local workding directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n")
		makeup.PrintWarning(" Please select a demo sub-command.\n")
		fmt.Printf(" Examples: \n")
		fmt.Printf("\ta9s cluster pwd\t\tPrint the configured working directory.\n")
		fmt.Printf("\n\n")
	},
}

var cmdClusterPwd = &cobra.Command{
	Use:   "pwd",
	Short: "Print the configured working directory for demos from the ~/.a8s config file.",
	Long:  `Print the configured working directory for demos from the ~/.a8s config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.EstablishConfig()

		fmt.Printf("%s", demo.DemoConfig.WorkingDir)
	},
}

// TODO Move. This is not the right place for business logic.
func CreateA8sDemoEnvironment() {
	makeup.PrintWelcomeScreen(demo.UnattendedMode)

	demo.EstablishConfig()

	demo.CheckPrerequisites()

	makeup.WaitForUser(demo.UnattendedMode)

	demo.CheckoutDeploymentGitRepository()

	demo.CheckoutDemoAppGitRepository()

	if demo.CountPodsInDemoNamespace() == 0 {
		makeup.PrintCheckmark("Kubernetes cluster has no pods in " + demo.GetConfig().DemoSpace + " namespace.")
	}

	demo.EstablishBackupStoreCredentials()

	k8s.ApplyCertManagerManifests(demo.UnattendedMode)

	demo.ApplyA8sManifests()

	demo.WaitForA8sSystemToBecomeReady()

	demo.PrintDemoSummary()
}

func init() {
	cmdCluster.AddCommand(cmdClusterPwd)
	rootCmd.AddCommand(cmdCluster)
}
