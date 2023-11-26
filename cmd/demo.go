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

var cmdDemoDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete the demo Kubernetes cluster.",
	Long: `Delete the demo Kubernetes cluster in order to free corresponding 
	resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.DeleteKubernetesCluster()
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

	// Depricated: Focus on minikube support
	//cmdDemoA8sPG.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "minikube", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" an \"kind\"")

	cmdDemoA8sPG.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "us-east-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdDemoA8sPG.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmdDemoA8sPG.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-provider", "AWS", "specify the infrastructure provider as supported by the a8s Backup Manager.")
	cmdDemo.AddCommand(cmdDemoA8sPG)
	cmdDemo.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")
	cmdDemo.PersistentFlags().BoolVarP(&demo.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")
	cmdDemo.AddCommand(cmdDemoPwd)
	cmdDemo.AddCommand(cmdDemoDelete)
	rootCmd.AddCommand(cmdDemo)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runDemoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runDemoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
