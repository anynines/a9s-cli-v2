package cmd

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var TheA8sPGProductName = "a8s Postgres"

var cmdDemo = &cobra.Command{
	Use:   "demo",
	Short: "Commands related to a9s Platform demos.",
	Long:  `Commands related to a9s Platform demos, e.g. printing the local workding directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n")
		makeup.PrintWarning(" Please select a demo sub-command.\n")
		fmt.Printf(" Examples: \n")
		fmt.Printf("\ta9s demo pwd\t\tPrint the configured working directory.\n")
		fmt.Printf("\ta9s demo a8s-pg\t\tExecute the a8s-pg product demo.")
		fmt.Printf("\n\n")
	},
}

// var cmdDemoA8sPG = &cobra.Command{
// 	Use:   "a8s-pg",
// 	Short: "Handling a9s Platform demos.",
// 	Long: `The demo assistent guides through the creation of a9s Platform demos,
// 	helps to install all necessary prerequisites and finally configures and installs
// 	the chosen product.`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		CreateA8sDemoEnvironment()
// 	},
// }

var cmdDemoPwd = &cobra.Command{
	Use:   "pwd",
	Short: "Print the configured working directory for demos from the ~/.a8s config file.",
	Long:  `Print the configured working directory for demos from the ~/.a8s config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.EstablishConfig()

		fmt.Printf("%s", demo.DemoConfig.WorkingDir)
	},
}

// var cmdDemoDelete = &cobra.Command{
// 	Use:   "delete",
// 	Short: "Delete the demo Kubernetes cluster.",
// 	Long: `Delete the demo Kubernetes cluster in order to free corresponding
// 	resources.`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		demo.DeleteKubernetesCluster()
// 	},
// }

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
	cmdDemo.AddCommand(cmdDemoPwd)
	rootCmd.AddCommand(cmdDemo)
}
