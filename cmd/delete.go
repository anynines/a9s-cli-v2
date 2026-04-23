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
	"github.com/spf13/cobra"
)

var Namespace, ServiceInstanceName string
var deleteKlutchDryRun bool
var deleteKlutchCleanupDNSACM bool
var deleteKlutchDeleteDNSZone bool
var deleteKlutchDeleteACMCertificate bool
var deleteKlutchHostedZoneName string
var deleteKlutchACMCertificateARN string
var deleteKlutchCleanupOrphans bool
var deleteKlutchReally bool
var deleteKlutchScheduleKmsDeletion bool
var deleteKlutchTenantRegion string
var deleteKlutchTenantSecretName string

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
	Use:   "cluster",
	Short: "Delete resources.",
	Long:  `Delete given resources. Use sub-commands to chose the resource to delete.`,
	Run: func(cmd *cobra.Command, args []string) {

		makeup.PrintWarning(" " + "Use a sub-command to choose the demo resource to be deleted.")

		cmd.Help()
	},
}

var cmdDeleteDemoA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Delete the given a8s Data Service Kubernetes cluster.",
	Long: `Delete the given a8s Data Service Kubernetes cluster in order to free corresponding
	resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.SelectClusterProvider()
		clusterManager := demo.BuildKubernetesClusterManager()
		clusterManager.DeleteKubernetesCluster()
	},
}

var cmdDeletePG = &cobra.Command{
	Use:   "pg",
	Short: "Delete PostgreSQL resources such as service instances, service bindings, backups and restore jobs.",
	Long:  `Delete PostgreSQL resources such as service instances, service bindings, backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a PostgreSQL resource such as (service) instance.")
		cmd.Help()
	},
}

var cmdDeletePGInstance = &cobra.Command{
	Use:   "instance",
	Short: "Delete a PostgreSQL service instance.",
	Long:  `Delete a PostgreSQL service instance`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		a8s.DeletePGServiceInstance(Namespace, ServiceInstanceName)

		//TODO Make configurable
		// demo.WaitForServiceInstanceToBecomeReady("default", "sample-pg-cluster", 3)
	},
}

var cmdDeletePGBinding = &cobra.Command{
	Use:   "servicebinding",
	Short: "Delete a PostgreSQL service binding.",
	Long:  "Delete a PostgreSQL service binding.",
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		a8s.DeletePGServiceBinding()
	},
}

var cmdDeleteKlutchControlPlane = &cobra.Command{
	Use:   "control-plane",
	Short: "Delete Klutch control plane resources from the current Kubernetes cluster.",
	Long:  `Deletes Klutch control plane resources (Dex, backend, ingress, Crossplane release) from the current kube context.`,
	Run: func(cmd *cobra.Command, args []string) {
		klutch.DeleteControlPlaneInstall()
	},
}

var cmdDeleteKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Delete Klutch components from the current Kubernetes cluster.",
	Long:  `Deletes Klutch components such as the control plane from the current kube context.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		cmd.Help()
	},
}

func runKlutchClusterDeletion(provider string, opts klutchaws.DeleteOptions, deleter func(context.Context, klutchaws.DeleteOptions)) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return fmt.Errorf("Please select a provider via -p. Supported provider for Klutch cluster deletion is \"aws\".")
	}

	if provider != "aws" {
		return fmt.Errorf("The Klutch cluster deletion currently only supports the \"aws\" provider.")
	}

	deleter(context.Background(), opts)
	return nil
}

var cmdDeleteClusterKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Delete Klutch clusters on AWS.",
	Long:  "Use the control-plane or workload subcommand to delete the corresponding Klutch EKS cluster and tagged AWS infrastructure.",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select either the control-plane or workload subcommand.")
		cmd.Help()
	},
}

var cmdDeleteClusterKlutchControlPlane = &cobra.Command{
	Use:   "control-plane",
	Short: "Delete the Klutch control plane cluster (AWS).",
	Long:  `Deletes the Klutch control plane EKS cluster and tagged AWS infrastructure (VPC, subnets, NAT, ALB, IAM). Optional flags can also remove Klutch Route53 DNS/hosted zone and ACM certificate.`,
	Run: func(cmd *cobra.Command, args []string) {
		if (deleteKlutchCleanupDNSACM || deleteKlutchDeleteDNSZone) && strings.TrimSpace(deleteKlutchHostedZoneName) == "" {
			makeup.ExitDueToFatalError(nil, "Hosted zone name is required when using --cleanup-dns-acm or --delete-dns-zone.")
		}

		opts := klutchaws.DeleteOptions{
			DryRun:                deleteKlutchDryRun,
			IncludeDNSRecords:     deleteKlutchCleanupDNSACM || deleteKlutchDeleteDNSZone,
			IncludeHostedZone:     deleteKlutchCleanupDNSACM || deleteKlutchDeleteDNSZone,
			IncludeSSLCertificate: deleteKlutchCleanupDNSACM || deleteKlutchDeleteACMCertificate,
			HostedZoneName:        deleteKlutchHostedZoneName,
			ACMCertificateARN:     deleteKlutchACMCertificateARN,
			CleanupOrphans:        deleteKlutchCleanupOrphans,
			SkipPrompt:            makeup.UnattendedMode && deleteKlutchReally,
			ScheduleKmsDeletion:   deleteKlutchScheduleKmsDeletion,
		}

		if cmd.Flags().Changed("cluster-name") {
			opts.ClusterName = strings.TrimSpace(demo.DemoClusterName)
		}

		if err := runKlutchClusterDeletion(demo.KubernetesTool, opts, klutchaws.DeleteControlPlaneCluster); err != nil {
			makeup.ExitDueToFatalError(nil, err.Error())
		}
	},
}

var cmdDeleteClusterKlutchWorkload = &cobra.Command{
	Use:   "workload",
	Short: "Delete the Klutch workload cluster (AWS).",
	Long:  `Deletes the Klutch workload EKS cluster and tagged AWS infrastructure (VPC, subnets, NAT, ALB, IAM).`,
	Run: func(cmd *cobra.Command, args []string) {
		if (deleteKlutchCleanupDNSACM || deleteKlutchDeleteDNSZone) && strings.TrimSpace(deleteKlutchHostedZoneName) == "" {
			makeup.ExitDueToFatalError(nil, "Hosted zone name is required when using --cleanup-dns-acm or --delete-dns-zone.")
		}

		opts := klutchaws.DeleteOptions{
			DryRun:                deleteKlutchDryRun,
			IncludeDNSRecords:     deleteKlutchCleanupDNSACM || deleteKlutchDeleteDNSZone,
			IncludeHostedZone:     deleteKlutchCleanupDNSACM || deleteKlutchDeleteDNSZone,
			IncludeSSLCertificate: deleteKlutchCleanupDNSACM || deleteKlutchDeleteACMCertificate,
			HostedZoneName:        deleteKlutchHostedZoneName,
			ACMCertificateARN:     deleteKlutchACMCertificateARN,
			SkipPrompt:            makeup.UnattendedMode && deleteKlutchReally,
		}

		if cmd.Flags().Changed("cluster-name") {
			opts.ClusterName = strings.TrimSpace(demo.DemoClusterName)
		} else if envName := strings.TrimSpace(os.Getenv("WORKLOAD_CLUSTER_NAME")); envName != "" {
			opts.ClusterName = envName
		}

		if strings.TrimSpace(opts.ClusterName) == "" {
			makeup.ExitDueToFatalError(nil, "Please provide --cluster-name or set WORKLOAD_CLUSTER_NAME to delete a Klutch workload cluster.")
		}

		if err := runKlutchClusterDeletion(demo.KubernetesTool, opts, klutchaws.DeleteWorkloadCluster); err != nil {
			makeup.ExitDueToFatalError(nil, err.Error())
		}
	},
}

var cmdDeleteKlutchTenant = &cobra.Command{
	Use:   "tenant",
	Short: "Delete a Klutch tenant (Cognito credentials secret).",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tenantName := strings.TrimSpace(args[0])
		if tenantName == "" {
			makeup.ExitDueToFatalError(nil, "Please provide a tenant name.")
		}
		region := strings.TrimSpace(deleteKlutchTenantRegion)
		if region == "" {
			region = klutchaws.ControlPlaneDefaultRegion()
		}
		secretName := klutchaws.TenantSecretName(tenantName, deleteKlutchTenantSecretName)

		if !makeup.ConfirmYes(fmt.Sprintf("This will delete Secrets Manager secret %s in %s. Type 'yes' to continue: ", secretName, region)) {
			makeup.PrintInfo("Deletion aborted.")
			return
		}

		if err := klutchaws.DeleteTenantSecret(context.Background(), region, secretName); err != nil {
			makeup.ExitDueToFatalError(err, "Failed to delete tenant secret.")
		}
		makeup.PrintSuccessSummary(fmt.Sprintf("Deleted tenant secret %s in %s.", secretName, region))
	},
}

func init() {

	cmdDeletePGInstance.PersistentFlags().StringVar(&ServiceInstanceName, "name", "a8s-pg-instance", "name of the pg service instance to be deleted.")
	cmdDeletePGInstance.PersistentFlags().StringVarP(&Namespace, "namespace", "n", "default", "namespace of the pg service instance to be deleted.")
	cmdDeletePG.AddCommand(cmdDeletePGInstance)
	cmdDelete.AddCommand(cmdDeletePG)
	cmdDelete.AddCommand(cmdDeleteDemo)

	// Service Bindings
	cmdDeletePG.PersistentFlags().StringVar(&demo.A8sPGServiceBinding.Name, "name", "example-pg-1", "name of the PG service binding. NOT the name of the PG service instance.")
	cmdDeletePGBinding.Flags().StringVarP(&demo.A8sPGServiceBinding.Namespace, "namespace", "n", "default", "namespace of the PG service binding.")

	cmdDeletePG.AddCommand(cmdDeletePGBinding)

	cmdDeleteDemo.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "", "provider for the Kubernetes cluster. Valid options are \"minikube\", \"kind\", and \"aws\" (for Klutch).")
	cmdDeleteDemo.PersistentFlags().BoolVar(&deleteKlutchDryRun, "dry-run", false, "Show planned AWS deletions for Klutch without making changes.")
	addKlutchControlPlaneFlags(cmdDeleteClusterKlutchControlPlane)
	addKlutchWorkloadFlags(cmdDeleteClusterKlutchWorkload)
	cmdDeleteKlutchTenant.Flags().StringVar(&deleteKlutchTenantRegion, "region", "", "AWS region for Cognito/Secrets Manager (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	cmdDeleteKlutchTenant.Flags().StringVar(&deleteKlutchTenantSecretName, "secret-name", "", "Secrets Manager name that holds the tenant credentials (defaults to klutch/<tenant>/oidc-client).")
	cmdDeleteDemo.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")
	cmdDeleteDemo.PersistentFlags().BoolVarP(&makeup.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")

	cmdDeleteDemo.AddCommand(cmdDeleteDemoA8s)
	cmdDeleteDemo.AddCommand(cmdDeleteClusterKlutch)
	cmdDeleteClusterKlutch.AddCommand(cmdDeleteClusterKlutchControlPlane)
	cmdDeleteClusterKlutch.AddCommand(cmdDeleteClusterKlutchWorkload)
	cmdDeleteKlutch.AddCommand(cmdDeleteKlutchControlPlane)
	cmdDelete.AddCommand(cmdDeleteKlutch)
	cmdDelete.AddCommand(cmdDeleteKlutchTenant)
	rootCmd.AddCommand(cmdDelete)
}

func addKlutchControlPlaneFlags(cmd *cobra.Command) {
	initRequiredStringFlagP(cmd, &demo.KubernetesTool, "provider", "p", "aws", "provider for deleting the Kubernetes cluster. Currently the only valid option for Klutch is \"aws\".")
	cmd.Flags().BoolVar(&deleteKlutchCleanupDNSACM, "cleanup-dns-acm", false, "Delete Klutch Route53 DNS records/hosted zone and ACM certificate (opt-in; destructive).")
	cmd.Flags().BoolVar(&deleteKlutchDeleteDNSZone, "delete-dns-zone", false, "Delete Klutch Route53 hosted zone (and its records).")
	cmd.Flags().BoolVar(&deleteKlutchDeleteACMCertificate, "delete-acm-certificate", false, "Delete Klutch ACM certificate.")
	initRequiredStringFlagWithDependency(&deleteKlutchCleanupDNSACM, "cleanup-dns-acm", true, cmd, &deleteKlutchHostedZoneName, "hosted-zone-name", "", "Hosted zone name to clean up when using DNS deletion flags.")
	setStringFlagDependency(&deleteKlutchDeleteDNSZone, "delete-dns-zone", true, cmd, &deleteKlutchHostedZoneName, "hosted-zone-name")
	cmd.Flags().StringVar(&deleteKlutchACMCertificateARN, "acm-certificate-arn", "", "ACM certificate ARN to delete (falls back to discovering a tagged Klutch certificate).")
	cmd.Flags().BoolVar(&deleteKlutchCleanupOrphans, "cleanup-orphans", false, "Attempt to remove Klutch-tagged orphaned AWS resources (e.g., EIPs) after cluster deletion.")
	cmd.Flags().BoolVar(&deleteKlutchReally, "really", false, "Confirm destructive Klutch cluster deletion (requires --yes to skip the prompt).")
	cmd.Flags().BoolVar(&deleteKlutchScheduleKmsDeletion, "schedule-kms-deletion", false, "Attempt to schedule any Klutch-tagged KMS keys for deletion after 7 days.")
}

func addKlutchWorkloadFlags(cmd *cobra.Command) {
	initRequiredStringFlagP(cmd, &demo.KubernetesTool, "provider", "p", "aws", "provider for deleting the Kubernetes cluster. Currently the only valid option for Klutch is \"aws\".")
	initRequiredStringFlagP(cmd, &demo.DemoClusterName, "cluster-name", "c", "", "name of the Workload cluster to delete.")
	cmd.Flags().BoolVar(&deleteKlutchCleanupDNSACM, "cleanup-dns-acm", false, "Delete Klutch Route53 DNS records/hosted zone and ACM certificate (opt-in; destructive).")
	cmd.Flags().BoolVar(&deleteKlutchDeleteDNSZone, "delete-dns-zone", false, "Delete Klutch Route53 hosted zone (and its records).")
	cmd.Flags().BoolVar(&deleteKlutchDeleteACMCertificate, "delete-acm-certificate", false, "Delete Klutch ACM certificate.")
	initRequiredStringFlagWithDependency(&deleteKlutchCleanupDNSACM, "cleanup-dns-acm", true, cmd, &deleteKlutchHostedZoneName, "hosted-zone-name", "", "Hosted zone name to clean up when using DNS deletion flags.")
	setStringFlagDependency(&deleteKlutchDeleteDNSZone, "delete-dns-zone", true, cmd, &deleteKlutchHostedZoneName, "hosted-zone-name")
	cmd.Flags().StringVar(&deleteKlutchACMCertificateARN, "acm-certificate-arn", "", "ACM certificate ARN to delete (falls back to discovering a tagged Klutch certificate).")
	cmd.Flags().BoolVar(&deleteKlutchCleanupOrphans, "cleanup-orphans", false, "Attempt to remove Klutch-tagged orphaned AWS resources (e.g., EIPs) after cluster deletion.")
	cmd.Flags().BoolVar(&deleteKlutchReally, "really", false, "Confirm destructive Klutch cluster deletion (requires --yes to skip the prompt).")
}
