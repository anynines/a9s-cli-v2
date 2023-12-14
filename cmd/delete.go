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
		demo.DeleteKubernetesCluster()
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

	/*
		The required struct to generate a yaml file should already be present in the operator.
		This also creates a tight depedency to the operator itself including
		api versions and the corresponding data schema comprising configurable attributes.
		This means that the CLI version needs to be kept in sync with the operator versions.
		Assuming that more and more services will be supported, it may require to
		modify the CLI from various teams.

		Hence, over time the codebase must be split into sub modules and some types of changes must happen
		fully automtically or otherwise the release process becomes a nightmare and may lead to
		a large delay between operator and CLI releases.

	*/

	// apiVersion
	// name
	// namespace
	// replicas
	// volume size
	// version
	// resource requests cpu
	// resource limits memory

	// expose
	// affinity

	// cmdPGInstance.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "us-east-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdDeletePG.AddCommand(cmdDeletePGInstance)
	cmdDelete.AddCommand(cmdDeletePG)
	cmdDelete.AddCommand(cmdDeleteDemo)

	cmdDeleteDemo.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "minikube", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" an \"kind\"")
	cmdDeleteDemo.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")
	cmdDeleteDemo.PersistentFlags().BoolVarP(&demo.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")

	cmdDeleteDemo.AddCommand(cmdDeleteDemoA8s)
	rootCmd.AddCommand(cmdDelete)
}
