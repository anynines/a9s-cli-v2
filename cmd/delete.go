package cmd

import (
	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var cmdDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete data service resources such as data service instances, service bindings, backups and restore jobs.",
	Long:  `Delete data service resources including data service instances, service bindings backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		//ExecuteA8sPGDemo()

		makeup.PrintWarning(" " + "Please select the data service resource type you would like to delete.")

		cmd.Help()
	},
}

var cmdDeleteDemo = &cobra.Command{
	Use:   "demo",
	Short: "Delete demo resources.",
	Long:  `Delete demo resources of the corresponding resources. Use sub-commands to chose the demo resource to free.`,
	Run: func(cmd *cobra.Command, args []string) {

		makeup.PrintWarning(" " + "Use a sub-command to choose the demo resource to be freed.")

		cmd.Help()
	},
}

var cmdDeleteDemoA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Delete the a8s Data Service demo Kubernetes cluster.",
	Long: `Delete the a8s Data Service demo Kubernetes cluster in order to free corresponding 
	resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterManager := demo.BuildKubernetesClusterManager()

		clusterManager.DeleteKubernetesCluster()
	},
}

var cmdDeletePG = &cobra.Command{
	Use:   "pg",
	Short: "Delete PostgreSQL resources such as service instances, service bindings, backups and restore jobs.",
	Long:  `Delete PostgreSQL resources such as service instances, service bindings, backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// ExecuteA8sPGDemo()
		makeup.PrintWarning(" " + "Please select a PostgreSQL resource such as (service) instance.")
		cmd.Help()
	},
}

var cmdDeletePGInstance = &cobra.Command{
	Use:   "instance",
	Short: "Delete a PostgreSQL service instance.",
	Long:  `Delete a PostgreSQL service instance`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.DeletePGServiceInstance()

		//TODO Make configurable
		// demo.WaitForServiceInstanceToBecomeReady("default", "sample-pg-cluster", 3)
	},
}

func init() {

	cmdDeletePGInstance.PersistentFlags().StringVar(&demo.DeleteA8sPGInstanceName, "name", "a8s-pg-instance", "name of the pg service instance to be deleted.")
	cmdDeletePG.AddCommand(cmdDeletePGInstance)
	cmdDelete.AddCommand(cmdDeletePG)
	cmdDelete.AddCommand(cmdDeleteDemo)

	cmdDeleteDemo.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "minikube", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" an \"kind\"")
	cmdDeleteDemo.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")
	cmdDeleteDemo.PersistentFlags().BoolVarP(&demo.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")

	cmdDeleteDemo.AddCommand(cmdDeleteDemoA8s)
	rootCmd.AddCommand(cmdDelete)
}
