package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/klutch"
	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/a9s-cli-v2/pg"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createKlutchDryRun bool
var createKlutchControlPlane = klutchaws.CreateControlPlaneCluster
var createKlutchWorkload = klutchaws.CreateWorkloadCluster
var createKlutchSkipApply bool
var createKlutchApplyHost string
var createKlutchApplyIngressPort int
var createKlutchApplyACMCertificateARN string
var createKlutchApplyHostedZone string
var createKlutchOIDCProvider string
var createKlutchOIDCIssuerURL string
var createKlutchOIDCClientID string
var createKlutchOIDCClientSecret string
var createKlutchOIDCCallbackURL string
var createKlutchTenantName string
var createKlutchTenantRegion string
var createKlutchTenantUserPoolID string
var createKlutchTenantStoreSecret bool = true
var createKlutchTenantSecretName string
var createKlutchTenantForce bool

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
		a8s := demo.NewA8sDemoManager("")
		a8s.CreatePGServiceInstance()

		if !(demo.DoNotApply) {
			instance := demo.A8sPGServiceInstance
			a8s.WaitForServiceInstanceToBecomeReady(instance.Namespace, instance.Name, instance.Replicas)
		}
	},
}

var cmdPGBackup = &cobra.Command{
	Use:   "backup",
	Short: "Create a PostgreSQL backup of a PostgreSQL service instance.",
	Long:  `Create a PostgreSQL backup of a PostgreSQL service instance`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")

		// DoNotApply is processed in the Create func
		a8s.CreatePGServiceInstanceBackup()
	},
}

var cmdPGRestore = &cobra.Command{
	Use:   "restore",
	Short: "Create a PostgreSQL restore of a PostgreSQL backup.",
	Long:  `Create a PostgreSQL restore of a PostgreSQL backup of a PostgreSQL service instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")

		// DoNotApply is processed in the Create func
		a8s.CreatePGServiceInstanceRestore()
	},
}

var cmdCreateCluster = &cobra.Command{
	Use:   "cluster",
	Short: "Create a local development Kubernetes cluster with a given stack.",
	Long: `Guides through the creation of a local development Kubernetes cluster, 
	helps to install all necessary prerequisites and finally configures and installs
	the chosen stack. Select a sub-command to create corresponding stack.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		cmd.Help()
	},
}

var cmdCreateStack = &cobra.Command{
	Use:   "stack",
	Short: "Applies the specified stack to the currently selected Kubernetes cluster.",
	Long: `Guides through the installation of the given anynines stack to the currently selected Kubernetes cluster, 
	helps to install all necessary prerequisites and finally configures and installs
	the chosen stack. Select a sub-command to create corresponding stack.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		cmd.Help()
	},
}

var cmdCreateClusterA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Create a local development cluster including a8s Data Services such as a8s Postgres.",
	Long:  `Helps with the creation of a local Kubernetes cluster, installing the a8s Data Service operator(s) including necessary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		a8s.CreateA8sStack(true)
	},
}

var cmdCreateClusterKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Create Klutch clusters.",
	Long:  `Create Klutch clusters on the selected provider. Currently only AWS is supported for the control plane.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		cmd.Help()
	},
}

var cmdCreateKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Create Klutch resources.",
	Long:  `Create Klutch resources such as tenants (Cognito app clients).`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		cmd.Help()
	},
}

var cmdCreateClusterKlutchControlPlane = &cobra.Command{
	Use:   "control-plane",
	Short: "Create the Klutch control plane cluster (and install it).",
	Long: `Creates the Klutch control plane cluster on the selected provider and installs the Klutch control plane components. 
Use --no-apply to only provision the cluster. Currently only AWS is supported.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := klutchaws.CreateOptions{DryRun: createKlutchDryRun}
		if cmd.Flags().Changed("cluster-name") {
			options.ClusterName = strings.TrimSpace(demo.DemoClusterName)
		}

		if err := runKlutchClusterCreation(demo.KubernetesTool, options); err != nil {
			makeup.ExitDueToFatalError(nil, err.Error())
		}

		if createKlutchDryRun {
			makeup.PrintInfo("Skipping Klutch control plane install because --dry-run was provided.")
			return
		}

		if createKlutchSkipApply {
			makeup.PrintInfo("Skipping Klutch control plane install because --no-apply was provided.")
			return
		}

		if createKlutchApplyHostedZone == "" {
			makeup.ExitDueToFatalError(nil, "The --hosted-zone-name flag is required to install the Klutch control plane. Use --no-apply to provision only.")
		}

		if createKlutchApplyIngressPort < 1 || createKlutchApplyIngressPort > 65535 {
			makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
		}

		klutch.SetControlPlaneOIDCOptions(klutch.OIDCOptions{
			Provider:     klutch.OIDCProvider(createKlutchOIDCProvider),
			IssuerURL:    createKlutchOIDCIssuerURL,
			ClientID:     createKlutchOIDCClientID,
			ClientSecret: createKlutchOIDCClientSecret,
			CallbackURL:  createKlutchOIDCCallbackURL,
		})

		klutch.ApplyKlutchControlPlane(createKlutchApplyHost, createKlutchApplyIngressPort, createKlutchApplyACMCertificateARN, createKlutchApplyHostedZone)
	},
}

var cmdCreateClusterKlutchTenant = &cobra.Command{
	Use:   "tenant",
	Short: "Create a Klutch tenant (Cognito app client) for workload bindings.",
	Long:  `Creates or reuses a Cognito app client scoped for Klutch bindings and prints issuer/client credentials for the workload owner.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(createKlutchTenantName) == "" {
			makeup.ExitDueToFatalError(nil, "The --tenant-name flag is required.")
		}

		region := strings.TrimSpace(createKlutchTenantRegion)
		if region == "" {
			region = klutchaws.ControlPlaneDefaultRegion()
		}

		secretName := klutchaws.TenantSecretName(createKlutchTenantName, createKlutchTenantSecretName)

		if klutchaws.TenantSecretExists(context.Background(), region, secretName) && !createKlutchTenantForce {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("A tenant secret already exists at %s in %s. Use --force to overwrite.", secretName, region))
		}

		makeup.PrintInfo("Creating or reusing Cognito app client for Klutch tenant...")
		tenantUUID := uuid.New().String()

		conn, err := klutchaws.EnsureCognitoOIDC(context.Background(), region, createKlutchTenantName, createKlutchTenantUserPoolID, tenantUUID)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to create or discover the Cognito app client for the Klutch tenant.")
		}
		conn.TenantUUID = tenantUUID
		conn.TenantName = createKlutchTenantName

		makeup.PrintH1("Klutch Tenant Credentials")
		makeup.PrintInfo(fmt.Sprintf("Issuer:        %s", conn.IssuerURL))
		makeup.PrintInfo(fmt.Sprintf("Client ID:     %s", conn.ClientID))
		makeup.PrintInfo(fmt.Sprintf("Client Secret: %s", conn.ClientSecret))
		makeup.PrintInfo(fmt.Sprintf("Scope:         %s", conn.Scope))
		makeup.PrintInfo(fmt.Sprintf("Tenant UUID:   %s", conn.TenantUUID))

		if createKlutchTenantStoreSecret {
			makeup.PrintInfo("Storing tenant credentials in AWS Secrets Manager...")
			if err := klutchaws.StoreCognitoCredentialsSecret(context.Background(), region, secretName, conn); err != nil {
				makeup.ExitDueToFatalError(err, "Failed to store Cognito credentials in AWS Secrets Manager.")
			}
			makeup.PrintSuccess(fmt.Sprintf("Stored credentials in Secrets Manager secret: %s", secretName))
		}

		if createKlutchTenantStoreSecret {
			makeup.PrintSuccessSummary("Credentials stored in AWS Secrets Manager; share access (not raw secrets) with the workload owner.")
			makeup.PrintInfo("Workload owners can create their cluster with `a9s create cluster klutch workload --provider aws --cluster-name <name>`.")
			makeup.PrintInfo("Keep the issuer/client ID/secret for this tenant; they are required by the non-interactive bind flow (`a9s klutch bind` using a token issued by this OIDC app).")
		} else {
			makeup.PrintSuccessSummary("Store these values securely (e.g., AWS Secrets Manager) and share with the workload cluster owner.")
			makeup.PrintInfo("Workload owners can create their cluster with `a9s create cluster klutch workload --provider aws --cluster-name <name>`.")
			makeup.PrintInfo("Keep the issuer/client ID/secret for this tenant; they are required by the non-interactive bind flow (`a9s klutch bind` using a token issued by this OIDC app).")
		}
	},
}

var cmdCreateClusterKlutchWorkload = &cobra.Command{
	Use:   "workload",
	Short: "Create a Klutch workload cluster (EKS).",
	Long:  `Creates a Klutch workload cluster on the selected provider. Currently only AWS is supported.`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := klutchaws.CreateOptions{DryRun: createKlutchDryRun}

		if cmd.Flags().Changed("cluster-name") {
			opts.ClusterName = strings.TrimSpace(demo.DemoClusterName)
		} else if envName := strings.TrimSpace(os.Getenv("WORKLOAD_CLUSTER_NAME")); envName != "" {
			opts.ClusterName = envName
		} else {
			opts.ClusterName = klutchaws.RandomWorkloadClusterName()
			makeup.PrintInfo(fmt.Sprintf("Generated workload cluster name: %s", opts.ClusterName))
		}

		if err := runKlutchClusterCreationWith(demo.KubernetesTool, opts, createKlutchWorkload); err != nil {
			makeup.ExitDueToFatalError(nil, err.Error())
		}
	},
}

func runKlutchClusterCreation(provider string, opts klutchaws.CreateOptions) error {
	return runKlutchClusterCreationWith(provider, opts, createKlutchControlPlane)
}

func runKlutchClusterCreationWith(provider string, opts klutchaws.CreateOptions, creator func(context.Context, klutchaws.CreateOptions)) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return fmt.Errorf("Please select a provider via -p. Supported provider for Klutch cluster creation is \"aws\".")
	}

	if provider != "aws" {
		return fmt.Errorf("The Klutch cluster creation currently only supports the \"aws\" provider.")
	}

	creator(context.Background(), opts)
	return nil
}

var cmdCreateStackA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Applies the a8s stack to the currently selected Kubernetes cluster.",
	Long:  `Applies the a8s stack to the currently selected Kubernetes cluster, installing the a8s Data Service operator(s) including necessary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		a8s.CreateA8sStack(false)
	},
}

var cmdCreatePGBinding = &cobra.Command{
	Use:   "servicebinding",
	Short: "Create a PostgreSQL service binding = Postgres user/pass + Kubernets Secret.",
	Long:  `Create a PostgreSQL service binding and thus a Kubernetes Secret containing username/password credentials unique to the service binding.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		demo.A8sPGServiceBinding.ServiceInstanceKind = pg.A8sPGServiceInstanceKind
		a8s.CreatePGServiceBinding()
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

	// create cluster a8s
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "us-east-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureProvider, "backup-provider", "minio", "specify the infrastructure provider as supported by the a8s Backup Manager. Valid options are: minio and AWS.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureEndpoint, "backup-store-endpoint", "", "the endpoint of the S3 compatible backup object storage. When minio is selected, the default is set to http://minio.minio-dev.svc.cluster.local:9000.")
	cmdCreateClusterA8s.PersistentFlags().BoolVar(&demo.BackupInfrastructurePathStyle, "backup-store-pathstyle", false, "influences the URI schema used to talk to the S3 compatible backup object store. Default is false but is set to true when minio is selected as backup-provider.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupStoreAccessKey, "backup-store-accesskey", "a8s-user", "the access key id for the backup store.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupStoreSecretKey, "backup-store-secretkey", "a8s-password", "the secret key for the backup store.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.DeploymentVersion, "deployment-version", "v1.2.0", "specify the version corresponding to the a8s-deployment git version tag. Use \"latest\" to get the untagged version.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.ClusterNrOfNodes, "cluster-nr-of-nodes", "3", "specify number of Kubernetes nodes.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.ClusterMemory, "cluster-memory", "4gb", "specify memory of the Kubernetes cluster.")
	cmdCreateClusterA8s.PersistentFlags().BoolVar(&demo.NoPreCheck, "no-precheck", false, "skip the verification of prerequisites.")

	//TODO Remove duplicate code in default values between cluster and stack.
	// create stack a8s
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "eu-central-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureProvider, "backup-provider", "minio", "specify the infrastructure provider as supported by the a8s Backup Manager. Valid options are: minio and AWS.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureEndpoint, "backup-store-endpoint", "", "the endpoint of the S3 compatible backup object storage. When minio is selected, the default is set to http://minio.minio-dev.svc.cluster.local:9000.")
	cmdCreateStackA8s.PersistentFlags().BoolVar(&demo.BackupInfrastructurePathStyle, "backup-store-pathstyle", false, "influences the URI schema used to talk to the S3 compatible backup object store. Default is false but is set to true when minio is selected as backup-provider.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupStoreAccessKey, "backup-store-accesskey", "a8s-user", "the access key id for the backup store.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupStoreSecretKey, "backup-store-secretkey", "a8s-password", "the secret key for the backup store.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.DeploymentVersion, "deployment-version", "v1.2.0", "specify the version corresponding to the a8s-deployment git version tag. Use \"latest\" to get the untagged version.")
	cmdCreateStackA8s.PersistentFlags().BoolVar(&demo.NoPreCheck, "no-precheck", false, "skip the verification of prerequisites.")

	// create demo
	cmdCreateCluster.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" and \"kind\" for local demos, as well as \"aws\" for Klutch.")
	cmdCreateCluster.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")
	cmdCreateClusterKlutchControlPlane.Flags().BoolVar(&createKlutchDryRun, "dry-run", false, "Show planned AWS resources and commands for Klutch without creating them.")
	cmdCreateClusterKlutchWorkload.Flags().BoolVar(&createKlutchDryRun, "dry-run", false, "Show planned AWS resources and commands for Klutch without creating them.")
	cmdCreateClusterKlutchControlPlane.Flags().BoolVar(&createKlutchSkipApply, "no-apply", false, "Create the Klutch control plane cluster without installing the Klutch control plane components.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchApplyHost, "host", "", "Host (IP or DNS name) to reach the ingress when applying the control plane. Defaults to the Kubernetes API server host of the current kube context.")
	cmdCreateClusterKlutchControlPlane.Flags().IntVar(&createKlutchApplyIngressPort, "ingress-port", 443, "Port the ingress should listen on when applying the control plane.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchApplyACMCertificateARN, "acm-certificate-arn", "", "ACM certificate ARN to enable HTTPS on the ALB ingress when applying the control plane.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchApplyHostedZone, "hosted-zone-name", "", "Route53 hosted zone name (FQDN). Required unless --no-apply is set. If provided and no ACM ARN is supplied, the CLI will request an ACM cert and create DNS validation records automatically.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchOIDCProvider, "oidc-provider", "", "OIDC provider to use for the Klutch control plane. Defaults to cognito when --provider=aws, otherwise dex.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchOIDCIssuerURL, "oidc-issuer-url", "", "OIDC issuer URL (required for oidc-provider=cognito).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchOIDCClientID, "oidc-client-id", "", "OIDC client ID (required for oidc-provider=cognito).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchOIDCClientSecret, "oidc-client-secret", "", "OIDC client secret (required for oidc-provider=cognito).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchOIDCCallbackURL, "oidc-callback-url", "", "OIDC callback URL to configure on the backend. Defaults to https://<host>/callback when not provided.")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantName, "tenant-name", "", "Name/prefix for the tenant (used to name the Cognito app client).")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantRegion, "region", "", "AWS region for Cognito (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantUserPoolID, "user-pool-id", "", "Existing Cognito user pool ID to reuse. If omitted, a pool named <tenant>-klutch is created or reused.")
	cmdCreateClusterKlutchTenant.Flags().BoolVar(&createKlutchTenantStoreSecret, "store-secret", true, "Store the tenant credentials in AWS Secrets Manager.")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantSecretName, "secret-name", "", "Secrets Manager name to store the tenant credentials (defaults to klutch/<tenant>/oidc-client).")
	cmdCreateClusterKlutchTenant.Flags().BoolVar(&createKlutchTenantForce, "force", false, "Overwrite an existing tenant secret if it already exists.")

	cmdCreateCluster.AddCommand(cmdCreateClusterA8s)
	cmdCreateClusterKlutch.AddCommand(cmdCreateClusterKlutchControlPlane)
	cmdCreateClusterKlutch.AddCommand(cmdCreateClusterKlutchWorkload)
	cmdCreateClusterKlutch.AddCommand(cmdCreateClusterKlutchTenant)
	cmdCreateCluster.AddCommand(cmdCreateClusterKlutch)
	cmdCreateKlutch.AddCommand(cmdCreateClusterKlutchTenant)
	cmdCreateStack.AddCommand(cmdCreateStackA8s)
	cmdCreateStack.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")

	cmdCreate.AddCommand(cmdCreateCluster)
	cmdCreate.AddCommand(cmdCreateKlutch)
	cmdCreate.AddCommand(cmdCreateStack)
	rootCmd.PersistentFlags().BoolVarP(&demo.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")
	rootCmd.PersistentFlags().BoolVarP(&makeup.Verbose, "verbose", "v", false, "enable verbose output?")
	rootCmd.AddCommand(cmdCreate)
}
