package cmd

import (
	"context"
	"strings"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/klutch"
	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var (
	sharedKlutchControlPlaneHost           string
	sharedKlutchControlPlanePort           int
	sharedKlutchControlPlaneACMCertARN     string
	sharedKlutchControlPlaneHostedZone     string
	sharedKlutchOIDCProvider               string
	sharedKlutchOIDCIssuerURL              string
	sharedKlutchOIDCClientID               string
	sharedKlutchOIDCClientSecret           string
	sharedKlutchOIDCCallbackURL            string
	applyKlutchClusterName                 string
	applyKlutchRegion                      string
	sharedKlutchTenantOperatorImage        string
	sharedKlutchTenantOperatorChart        string
	sharedKlutchTenantOperatorChartVersion string
	sharedKlutchTenantOperatorRoleARN      string
	sharedKlutchTenantOperatorRegion       string
	sharedKlutchTenantOperatorBindURL      string
	sharedKlutchTenantOperatorBindRequest  string
	sharedKlutchBackendImageRef            string
	sharedKlutchBackendImageURL            string
	sharedKlutchBackendImageTag            string
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply components to the current Kubernetes cluster.",
	Long:  `Applies components such as the Klutch control plane to the current Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var applyKlutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "Apply Klutch components to the current Kubernetes cluster.",
	Long:  `Applies Klutch components such as the control plane to the current Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var cmdApplyKlutchControlPlane = &cobra.Command{
	Use: "control-plane",
	Aliases: []string{
		"klutch-control-plane",
	},
	Short: "Install the Klutch control plane onto the current Kubernetes cluster.",
	Long:  "Installs the Klutch control plane and its dependencies into the currently selected kube context.",
	Run: func(cmd *cobra.Command, args []string) {
		if sharedKlutchControlPlanePort < 1 || sharedKlutchControlPlanePort > 65535 {
			makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
		}

		if sharedKlutchControlPlaneHostedZone == "" {
			makeup.ExitDueToFatalError(nil, "The --hosted-zone-name flag is required until self-signed certificates are supported.")
		}

		imgURL, imgTag := resolveBackendImageRef(sharedKlutchBackendImageRef, sharedKlutchBackendImageURL, sharedKlutchBackendImageTag)
		klutch.SetBindBackendImage(imgURL, imgTag)

		klutch.SetControlPlaneOIDCOptions(klutch.OIDCOptions{
			Provider:     klutch.OIDCProvider(sharedKlutchOIDCProvider),
			IssuerURL:    sharedKlutchOIDCIssuerURL,
			ClientID:     sharedKlutchOIDCClientID,
			ClientSecret: sharedKlutchOIDCClientSecret,
			CallbackURL:  sharedKlutchOIDCCallbackURL,
		})

		if strings.EqualFold(strings.TrimSpace(demo.KubernetesTool), "aws") {
			cfgOpts := klutchaws.CreateOptions{
				ClusterName:                strings.TrimSpace(applyKlutchClusterName),
				Region:                     strings.TrimSpace(applyKlutchRegion),
				TenantOperatorImage:        strings.TrimSpace(sharedKlutchTenantOperatorImage),
				TenantOperatorChart:        strings.TrimSpace(sharedKlutchTenantOperatorChart),
				TenantOperatorChartVersion: strings.TrimSpace(sharedKlutchTenantOperatorChartVersion),
				TenantOperatorRoleARN:      strings.TrimSpace(sharedKlutchTenantOperatorRoleARN),
				TenantOperatorRegion:       strings.TrimSpace(sharedKlutchTenantOperatorRegion),
				TenantOperatorBindURL:      strings.TrimSpace(sharedKlutchTenantOperatorBindURL),
				TenantOperatorBindRequest:  strings.TrimSpace(sharedKlutchTenantOperatorBindRequest),
				HostedZoneName:             strings.TrimSpace(sharedKlutchControlPlaneHostedZone),
			}
			klutchaws.ApplyControlPlaneAddons(context.Background(), cfgOpts)
		}

		klutch.ApplyKlutchControlPlane(sharedKlutchControlPlaneHost, sharedKlutchControlPlanePort, sharedKlutchControlPlaneACMCertARN, sharedKlutchControlPlaneHostedZone, applyKlutchClusterName)
	},
}

func init() {

	initFlagsApplyKlutchControlPlane(cmdApplyKlutchControlPlane)
	applyKlutchCmd.AddCommand(cmdApplyKlutchControlPlane)
	applyCmd.AddCommand(applyKlutchCmd)
	rootCmd.AddCommand(applyCmd)
}

func initFlagsApplyKlutchControlPlane(cmd *cobra.Command) {
	initSharedFlagsKlutchControlPlaneStack(cmd)

	cmd.Flags().StringVarP(&demo.KubernetesTool, "provider", "p", "", "provider for the Kubernetes cluster. Valid options are \"minikube\", \"kind\", and \"aws\" (for Klutch).")
	cmd.Flags().StringVarP(&demo.DemoClusterName, "cluster-name", "", "", "Existing AWS control-plane cluster name (defaults to klutch-control-plane).")
	cmd.Flags().StringVar(&applyKlutchRegion, "region", "", "AWS region for the EKS cluster (defaults to eu-central-1).")
}

func initSharedFlagsKlutchControlPlaneStack(cmd *cobra.Command) {
	cmd.Flags().StringVar(&sharedKlutchControlPlaneHost, "host", "", "Host (IP or DNS name) to reach the ingress. Defaults to the Kubernetes API server host of the current kube context.")
	cmd.Flags().StringVar(&sharedKlutchControlPlaneACMCertARN, "acm-certificate-arn", "", "ACM certificate ARN to enable HTTPS on the ALB ingress for the control plane.")
	cmd.Flags().IntVar(&sharedKlutchControlPlanePort, "ingress-port", 443, "Port the ingress should listen on.")
	initRequiredStringFlag(cmd, &sharedKlutchControlPlaneHostedZone, "hosted-zone-name", "", "Route53 hosted zone name (FQDN). Required until self-signed certificates are supported. If provided and no ACM ARN is supplied, the CLI will request an ACM cert and create DNS validation records automatically.")
	cmd.Flags().StringVar(&sharedKlutchOIDCProvider, "oidc-provider", "", "OIDC provider to use for the Klutch control plane. Defaults to cognito when --provider=aws, otherwise dex.")
	initRequiredStringFlagWithDependency(cmd, &sharedKlutchOIDCIssuerURL, "oidc-issuer-url", "", "OIDC issuer URL (required for oidc-provider=cognito).", &sharedKlutchOIDCProvider, "oidc-provider", "cognito")
	initRequiredStringFlagWithDependency(cmd, &sharedKlutchOIDCClientID, "oidc-client-id", "", "OIDC client ID (required for oidc-provider=cognito).", &sharedKlutchOIDCProvider, "oidc-provider", "cognito")
	initRequiredStringFlagWithDependency(cmd, &sharedKlutchOIDCClientSecret, "oidc-client-secret", "", "OIDC client secret (required for oidc-provider=cognito).", &sharedKlutchOIDCProvider, "oidc-provider", "cognito")
	cmd.Flags().StringVar(&sharedKlutchOIDCCallbackURL, "oidc-callback-url", "", "OIDC callback URL to configure on the backend. Defaults to https://<host>/callback when not provided.")
	cmd.Flags().StringVar(&sharedKlutchTenantOperatorImage, "tenant-operator-image", "", "Tenant operator container image (override default).")
	cmd.Flags().StringVar(&sharedKlutchTenantOperatorChart, "tenant-operator-chart", "", "Tenant operator Helm chart (OCI URL, override default).")
	cmd.Flags().StringVar(&sharedKlutchTenantOperatorChartVersion, "tenant-operator-chart-version", "", "Tenant operator Helm chart version (for OCI charts).")
	cmd.Flags().StringVar(&sharedKlutchTenantOperatorRoleARN, "tenant-operator-role-arn", "", "IAM role ARN for the tenant operator service account (IRSA).")
	cmd.Flags().StringVar(&sharedKlutchTenantOperatorRegion, "tenant-operator-region", "", "Region for tenant operator AWS calls (defaults to control-plane region).")
	cmd.Flags().StringVar(&sharedKlutchTenantOperatorBindURL, "tenant-operator-bind-url", "", "Bind URL to pass to the tenant operator config (override default).")
	cmd.Flags().StringVar(&sharedKlutchTenantOperatorBindRequest, "tenant-operator-bind-request", "", "Bind request JSON to pass to the tenant operator config (override default).")
	cmd.Flags().StringVar(&sharedKlutchBackendImageRef, "klutch-bind-backend-img", "", "Override the klutch-bind backend image as <repo>:<tag>.")
	cmd.Flags().StringVar(&sharedKlutchBackendImageURL, "klutch-bind-backend-img-url", "", "Override the klutch-bind backend image URL (repository).")
	cmd.Flags().StringVar(&sharedKlutchBackendImageTag, "klutch-bind-backend-img-tag", "", "Override the klutch-bind backend image tag.")
}
