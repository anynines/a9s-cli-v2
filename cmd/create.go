package cmd

import (
	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var cmdCreate = &cobra.Command{
	Use:   "create",
	Short: "Create data service resources such as data service instances, service bindings, backups and restore jobs.",
	Long:  `Create data service resources including data service instances, service bindings backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		//ExecuteA8sPGDemo()

		makeup.PrintWarning(" " + "Please select the data service resource type you would like to instantiate.")

		cmd.Help()
	},
}

var cmdPG = &cobra.Command{
	Use:   "pg",
	Short: "Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.",
	Long:  `Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		// ExecuteA8sPGDemo()
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
			instance := demo.A8sServiceInstance
			demo.WaitForServiceInstanceToBecomeReady(instance.Namespace, instance.Name, instance.Replicas)
		}
	},
}

var cmdCreateDemo = &cobra.Command{
	Use:   "demo",
	Short: "Create an a9s Platform demo environment.",
	Long: `The demo assistent guides through the creation of a9s Platform demos, 
	helps to install all necessary prerequisites and finally configures and installs
	the chosen product. Select a sub-command to create corresponding demo environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a demo sub-command.")
		cmd.Help()
	},
}

var cmdCreateDemoA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Create a demo environment for the pod based a8s Data Services such as a8s Postgres.",
	Long:  `The demo assistent helps with the creation of a local Kubernetes cluster, installing the a8s Data Service operator(s) including necessary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateA8sDemoEnvironment()
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

	// create pg instance
	cmdPG.PersistentFlags().StringVar(&demo.A8sServiceInstance.ApiVersion, "api-version", "v1beta3", "api version of the pg service instance.")
	cmdPG.PersistentFlags().StringVar(&demo.A8sServiceInstance.Name, "name", "a8s-pg-instance", "name of the pg service instance.")
	cmdPG.PersistentFlags().StringVar(&demo.A8sServiceInstance.Namespace, "namespace", "default", "namespace of the pg service instance.")
	cmdPG.PersistentFlags().IntVar(&demo.A8sServiceInstance.Replicas, "replicas", 1, "number of Pods (replicas) the service instance's statefulset will have.")
	cmdPG.PersistentFlags().StringVar(&demo.A8sServiceInstance.VolumeSize, "volume-size", "1Gi", "Volume size of the persistent volume claim(s)d of the service instance's statefulset.")
	cmdPG.PersistentFlags().StringVar(&demo.A8sServiceInstance.Version, "service-version", "14", "Postgres version. The given version must be supported by the automation.")
	cmdPG.PersistentFlags().StringVar(&demo.A8sServiceInstance.RequestsCPU, "requests-cpu", "100m", "Resources -> requests -> cpu of the service instance's statefulset.")
	cmdPG.PersistentFlags().StringVar(&demo.A8sServiceInstance.LimitsMemory, "limits-memory", "100Mi", "Resources -> limits -> memory  of the service instance's statefulset.")
	cmdPG.PersistentFlags().BoolVar(&demo.DoNotApply, "no-apply", false, "If this flag is set, the service instance YAML spec is not applied (kubectl apply -f).")

	// cmdPG.PersistentFlags().StringVarP(&demo.OutputFormat, "output", "o", "", "Output format. Options: \"yaml\".")

	cmdPG.AddCommand(cmdPGInstance)

	cmdCreate.AddCommand(cmdPG)

	// create demo a8s
	cmdCreateDemoA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "us-east-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdCreateDemoA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmdCreateDemoA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-provider", "AWS", "specify the infrastructure provider as supported by the a8s Backup Manager.")
	cmdCreateDemoA8s.PersistentFlags().StringVar(&demo.DeploymentVersion, "deployment-version", "v0.3.0", "specify the version corresponding to the a8s-deployment git version tag. Use \"latest\" to get the untagged version.")
	cmdCreateDemoA8s.PersistentFlags().StringVar(&demo.ClusterNrOfNodes, "cluster-nr-of-nodes", "3", "specify number of Kubernetes nodes.")
	cmdCreateDemoA8s.PersistentFlags().StringVar(&demo.ClusterMemory, "cluster-memory", "4gb", "specify memory of the Kubernetes cluster.")
	cmdCreateDemoA8s.PersistentFlags().BoolVar(&demo.NoPreCheck, "no-precheck", false, "skip the verification of prerequisites.")

	// create demo
	cmdCreateDemo.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "minikube", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" an \"kind\"")
	cmdCreateDemo.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")
	cmdCreateDemo.PersistentFlags().BoolVarP(&demo.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")

	cmdCreateDemo.AddCommand(cmdCreateDemoA8s)
	cmdCreate.AddCommand(cmdCreateDemo)
	rootCmd.AddCommand(cmdCreate)
}
