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
	applyKlutchControlPlaneHost           string
	applyKlutchControlPlanePort           int
	applyKlutchControlPlaneACMCertARN     string
	applyKlutchControlPlaneHostedZone     string
	applyKlutchOIDCProvider               string
	applyKlutchOIDCIssuerURL              string
	applyKlutchOIDCClientID               string
	applyKlutchOIDCClientSecret           string
	applyKlutchOIDCCallbackURL            string
	applyKlutchClusterName                string
	applyKlutchTenantOperatorImage        string
	applyKlutchTenantOperatorChart        string
	applyKlutchTenantOperatorChartVersion string
	applyKlutchTenantOperatorRoleARN      string
	applyKlutchTenantOperatorRegion       string
	applyKlutchTenantOperatorBindURL      string
	applyKlutchTenantOperatorBindRequest  string
	applyKlutchBackendImageRef            string
	applyKlutchBackendImageURL            string
	applyKlutchBackendImageTag            string
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply components to the current Kubernetes cluster.",
	Long:  `Applies components such as the Klutch control plane to the current Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		cmd.Help()
	},
}

var applyKlutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "Apply Klutch components to the current Kubernetes cluster.",
	Long:  `Applies Klutch components such as the control plane to the current Kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		cmd.Help()
	},
}

var applyKlutchControlPlaneCmd = &cobra.Command{
	Use: "control-plane",
	Aliases: []string{
		"klutch-control-plane",
	},
	Short: "Install the Klutch control plane onto the current Kubernetes cluster.",
	Long:  "Installs the Klutch control plane and its dependencies into the currently selected kube context.",
	Run: func(cmd *cobra.Command, args []string) {
		if applyKlutchControlPlanePort < 1 || applyKlutchControlPlanePort > 65535 {
			makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
		}

		if applyKlutchControlPlaneHostedZone == "" {
			makeup.ExitDueToFatalError(nil, "The --hosted-zone-name flag is required until self-signed certificates are supported.")
		}

		imgURL, imgTag := resolveBackendImageRef(applyKlutchBackendImageRef, applyKlutchBackendImageURL, applyKlutchBackendImageTag)
		klutch.SetBindBackendImage(imgURL, imgTag)

		klutch.SetControlPlaneOIDCOptions(klutch.OIDCOptions{
			Provider:     klutch.OIDCProvider(applyKlutchOIDCProvider),
			IssuerURL:    applyKlutchOIDCIssuerURL,
			ClientID:     applyKlutchOIDCClientID,
			ClientSecret: applyKlutchOIDCClientSecret,
			CallbackURL:  applyKlutchOIDCCallbackURL,
		})

		if strings.EqualFold(strings.TrimSpace(demo.KubernetesTool), "aws") {
			cfgOpts := klutchaws.CreateOptions{
				ClusterName:                strings.TrimSpace(applyKlutchClusterName),
				TenantOperatorImage:        strings.TrimSpace(applyKlutchTenantOperatorImage),
				TenantOperatorChart:        strings.TrimSpace(applyKlutchTenantOperatorChart),
				TenantOperatorChartVersion: strings.TrimSpace(applyKlutchTenantOperatorChartVersion),
				TenantOperatorRoleARN:      strings.TrimSpace(applyKlutchTenantOperatorRoleARN),
				TenantOperatorRegion:       strings.TrimSpace(applyKlutchTenantOperatorRegion),
				TenantOperatorBindURL:      strings.TrimSpace(applyKlutchTenantOperatorBindURL),
				TenantOperatorBindRequest:  strings.TrimSpace(applyKlutchTenantOperatorBindRequest),
				HostedZoneName:             strings.TrimSpace(applyKlutchControlPlaneHostedZone),
			}
			klutchaws.ApplyControlPlaneAddons(context.Background(), cfgOpts)
		}

		klutch.ApplyKlutchControlPlane(applyKlutchControlPlaneHost, applyKlutchControlPlanePort, applyKlutchControlPlaneACMCertARN, applyKlutchControlPlaneHostedZone, applyKlutchClusterName)
	},
}

func init() {
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchControlPlaneHost, "host", "", "Host (IP or DNS name) to reach the ingress. Defaults to the Kubernetes API server host of the current kube context.")
	applyKlutchControlPlaneCmd.Flags().IntVar(&applyKlutchControlPlanePort, "ingress-port", 443, "Port the ingress should listen on.")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchControlPlaneACMCertARN, "acm-certificate-arn", "", "ACM certificate ARN to enable HTTPS on the ALB ingress for the control plane.")
	initRequiredStringFlag(applyKlutchControlPlaneCmd, &applyKlutchControlPlaneHostedZone, "hosted-zone-name", "", "Route53 hosted zone name (FQDN). Required until self-signed certificates are supported. If provided and no ACM ARN is supplied, the CLI will request an ACM cert and create DNS validation records automatically.")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchOIDCProvider, "oidc-provider", "", "OIDC provider to use for the Klutch control plane. Defaults to cognito when --provider=aws, otherwise dex.")
	initRequiredStringFlagWithDependency(&applyKlutchOIDCProvider, "oidc-provider", "cognito", applyKlutchControlPlaneCmd, &applyKlutchOIDCIssuerURL, "oidc-issuer-url", "", "OIDC issuer URL (required for oidc-provider=cognito).")
	initRequiredStringFlagWithDependency(&applyKlutchOIDCProvider, "oidc-provider", "cognito", applyKlutchControlPlaneCmd, &applyKlutchOIDCClientID, "oidc-client-id", "", "OIDC client ID (required for oidc-provider=cognito).")
	initRequiredStringFlagWithDependency(&applyKlutchOIDCProvider, "oidc-provider", "cognito", applyKlutchControlPlaneCmd, &applyKlutchOIDCClientSecret, "oidc-client-secret", "", "OIDC client secret (required for oidc-provider=cognito).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchOIDCCallbackURL, "oidc-callback-url", "", "OIDC callback URL to configure on the backend. Defaults to https://<host>/callback when not provided.")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchClusterName, "cluster-name", "", "Existing AWS control-plane cluster name (defaults to klutch-control-plane).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchTenantOperatorImage, "tenant-operator-image", "", "Tenant operator container image (override default).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchTenantOperatorChart, "tenant-operator-chart", "", "Tenant operator Helm chart (OCI URL, override default).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchTenantOperatorChartVersion, "tenant-operator-chart-version", "", "Tenant operator Helm chart version (for OCI charts).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchTenantOperatorRoleARN, "tenant-operator-role-arn", "", "IAM role ARN for the tenant operator service account (IRSA).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchTenantOperatorRegion, "tenant-operator-region", "", "Region for tenant operator AWS calls (defaults to control-plane region).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchTenantOperatorBindURL, "tenant-operator-bind-url", "", "Bind URL to pass to the tenant operator config (override default).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchTenantOperatorBindRequest, "tenant-operator-bind-request", "", "Bind request JSON to pass to the tenant operator config (override default).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchBackendImageRef, "klutch-bind-backend-img", "", "Override the klutch-bind backend image as <repo>:<tag>.")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchBackendImageURL, "klutch-bind-backend-img-url", "", "Override the klutch-bind backend image URL (repository).")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchBackendImageTag, "klutch-bind-backend-img-tag", "", "Override the klutch-bind backend image tag.")
	applyKlutchControlPlaneCmd.Flags().StringVarP(&demo.KubernetesTool, "provider", "p", "", "provider for the Kubernetes cluster. Valid options are \"minikube\", \"kind\", and \"aws\" (for Klutch).")

	applyKlutchCmd.AddCommand(applyKlutchControlPlaneCmd)
	applyCmd.AddCommand(applyKlutchCmd)
	rootCmd.AddCommand(applyCmd)
}
