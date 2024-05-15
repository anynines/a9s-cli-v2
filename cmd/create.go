package cmd

import (
	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/a9s-cli-v2/pg"
	"github.com/spf13/cobra"
)

var cmdCreate = &cobra.Command{
	Use:   "create",
	Short: "Create data service resources such as data service instances, service bindings, backups and restore jobs.",
	Long:  `Create data service resources including data service instances, service bindings backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select the data service resource type you would like to instantiate.")

		cmd.Help()
	},
}

var cmdCreatePG = &cobra.Command{
	Use:   "pg",
	Short: "Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.",
	Long:  `Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a PostgreSQL resource such as (service) instance.")
		cmd.Help()
	},
}

var cmdPGInstance = &cobra.Command{
	Use:   "instance",
	Short: "Create a PostgreSQL service instance.",
	Long:  `Create a PostgreSQL service instance`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.CreatePGServiceInstance()

		if !(demo.DoNotApply) {
			instance := demo.A8sPGServiceInstance
			demo.WaitForServiceInstanceToBecomeReady(instance.Namespace, instance.Name, instance.Replicas)
		}
	},
}

var cmdPGBackup = &cobra.Command{
	Use:   "backup",
	Short: "Create a PostgreSQL backup of a PostgreSQL service instance.",
	Long:  `Create a PostgreSQL backup of a PostgreSQL service instance`,
	Run: func(cmd *cobra.Command, args []string) {
		// DoNotApply is processed in the Create func
		demo.CreatePGServiceInstanceBackup()
	},
}

var cmdPGRestore = &cobra.Command{
	Use:   "restore",
	Short: "Create a PostgreSQL restore of a PostgreSQL backup.",
	Long:  `Create a PostgreSQL restore of a PostgreSQL backup of a PostgreSQL service instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		// DoNotApply is processed in the Create func
		demo.CreatePGServiceInstanceRestore()
	},
}

var cmdCreateCluster = &cobra.Command{
	Use:   "cluster",
	Short: "Create a local development Kubernetes cluster with a given stack.",
	Long: `Guides through the creation of a local development Kubernetes cluster, 
	helps to install all necessary prerequisites and finally configures and installs
	the chosen stack. Select a sub-command to create corresponding stack.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a demo sub-command.")
		cmd.Help()
	},
}

var cmdCreateClusterA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Create a local development cluster including a8s Data Services such as a8s Postgres.",
	Long:  `Helps with the creation of a local Kubernetes cluster, installing the a8s Data Service operator(s) including necessary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateA8sDemoEnvironment()
	},
}

var cmdCreatePGBinding = &cobra.Command{
	Use:   "servicebinding",
	Short: "Create a PostgreSQL service binding = Postgres user/pass + Kubernets Secret.",
	Long:  `Create a PostgreSQL service binding and thus a Kubernetes Secret containing username/password credentials unique to the service binding.`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.A8sPGServiceBinding.ServiceInstanceKind = pg.A8sPGServiceInstanceKind
		demo.CreatePGServiceBinding()
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

	// create pg instance
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of thePGservice instance.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.Name, "name", "example-pg", "name of the PG service instance.")
	cmdPGInstance.PersistentFlags().StringVarP(&demo.A8sPGServiceInstance.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
	cmdPGInstance.PersistentFlags().IntVar(&demo.A8sPGServiceInstance.Replicas, "replicas", 1, "number of Pods (replicas) the service instance's statefulset will have.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.VolumeSize, "volume-size", "1Gi", "Volume size of the persistent volume claim(s)d of the service instance's statefulset.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.Version, "service-version", "14", "Postgres version. The given version must be supported by the automation.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.RequestsCPU, "requests-cpu", "100m", "Resources -> requests -> cpu of the service instance's statefulset.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.LimitsMemory, "limits-memory", "100Mi", "Resources -> limits -> memory  of the service instance's statefulset.")
	cmdPGInstance.PersistentFlags().BoolVar(&demo.DoNotApply, "no-apply", false, "If this flag is set, the service instance YAML spec is not applied (kubectl apply -f).")

	// cmdPG.PersistentFlags().StringVarP(&demo.OutputFormat, "output", "o", "", "Output format. Options: \"yaml\".")

	cmdCreatePG.AddCommand(cmdPGInstance)

	cmdPGBackup.PersistentFlags().StringVar(&demo.A8sPGBackup.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG backup.")
	cmdPGBackup.PersistentFlags().StringVar(&demo.A8sPGBackup.Name, "name", "example-pg-1", "name of the PG backup. Not the name of the service instance.")
	cmdPGBackup.PersistentFlags().StringVarP(&demo.A8sPGBackup.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to be backed up.")
	cmdPGBackup.PersistentFlags().StringVarP(&demo.A8sPGBackup.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
	cmdCreatePG.AddCommand(cmdPGBackup)

	// Should the restore act on the backup resource or should there be a separate object for it?
	cmdPGRestore.PersistentFlags().StringVar(&demo.A8sPGRestore.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG backup.")
	cmdPGRestore.PersistentFlags().StringVar(&demo.A8sPGRestore.Name, "name", "example-pg-1", "name of the PG restore. Not the name of the service instance or the backup.")
	cmdPGRestore.PersistentFlags().StringVarP(&demo.A8sPGRestore.BackupName, "backup", "b", "example-pg-backup", "name of the PG backup to be restored.")
	cmdPGRestore.PersistentFlags().StringVarP(&demo.A8sPGRestore.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to be restored.")
	cmdPGRestore.PersistentFlags().StringVarP(&demo.A8sPGRestore.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
	cmdCreatePG.AddCommand(cmdPGRestore)

	cmdCreate.AddCommand(cmdCreatePG)

	// create pg binding
	// cmdCreatePGBinding.
	cmdCreatePGBinding.PersistentFlags().StringVar(&demo.A8sPGServiceBinding.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG service binding.")
	cmdCreatePGBinding.PersistentFlags().StringVar(&demo.A8sPGServiceBinding.Name, "name", "example-pg-1", "name of the PG service binding. NOT the name of the PG service instance.")
	cmdCreatePGBinding.PersistentFlags().StringVarP(&demo.A8sPGServiceBinding.Namespace, "namespace", "n", "default", "namespace of the PG service instance. NOT the app's namespace.")
	cmdCreatePGBinding.PersistentFlags().StringVarP(&demo.A8sPGServiceBinding.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to bind to.")
	cmdCreatePG.AddCommand(cmdCreatePGBinding)

	// create demo a8s
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "eu-central-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureProvider, "backup-provider", "AWS", "specify the infrastructure provider as supported by the a8s Backup Manager.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.DeploymentVersion, "deployment-version", "v0.3.0", "specify the version corresponding to the a8s-deployment git version tag. Use \"latest\" to get the untagged version.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.ClusterNrOfNodes, "cluster-nr-of-nodes", "3", "specify number of Kubernetes nodes.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.ClusterMemory, "cluster-memory", "4gb", "specify memory of the Kubernetes cluster.")
	cmdCreateClusterA8s.PersistentFlags().BoolVar(&demo.NoPreCheck, "no-precheck", false, "skip the verification of prerequisites.")

	// create demo
	cmdCreateCluster.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "minikube", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" an \"kind\"")
	cmdCreateCluster.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")

	cmdCreateCluster.AddCommand(cmdCreateClusterA8s)
	cmdCreate.AddCommand(cmdCreateCluster)
	rootCmd.PersistentFlags().BoolVarP(&demo.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")
	rootCmd.PersistentFlags().BoolVarP(&makeup.Verbose, "verbose", "v", false, "enable verbose output?")
	rootCmd.AddCommand(cmdCreate)
}
