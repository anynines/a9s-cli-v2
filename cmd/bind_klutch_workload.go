package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/anynines/a9s-cli-v2/klutch"
	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
	"github.com/anynines/a9s-cli-v2/makeup"
	bindcmd "github.com/anynines/klutchio/bind/pkg/kubectl/bind/cmd"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	bindKlutchWorkloadControlPlane        string
	bindKlutchWorkloadControlPlaneCluster string
	bindKlutchWorkloadKubeconfig          string
	bindKlutchWorkloadContext             string
	bindKlutchWorkloadOutput              string
	bindKlutchWorkloadDryRun              bool
	bindKlutchWorkloadSkipKonnector       bool
	bindKlutchWorkloadKonnectorImage      string
	bindKlutchWorkloadExtraArgs           []string
	bindKlutchWorkloadInteractive         bool
	bindKlutchWorkloadRequestFile         string
	bindKlutchWorkloadOIDCClientID        string
	bindKlutchWorkloadOIDCClientSecret    string
	bindKlutchWorkloadOIDCTokenURL        string
	bindKlutchWorkloadOIDCScope           string
	bindKlutchWorkloadWriteKubeconfig     string
	bindKlutchWorkloadTenantName          string
	bindKlutchWorkloadTenantSecretName    string
	bindKlutchWorkloadTenantRegion        string
)

var bindCmd = &cobra.Command{
	Use:   "bind",
	Short: "Bind resources to control planes.",
	Long:  "Bind clusters and APIs to Klutch control planes.",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		_ = cmd.Help()
	},
}

var cmdBindKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Bind Klutch resources.",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		_ = cmd.Help()
	},
}

var cmdBindKlutchWorkload = &cobra.Command{
	Use:          "workload",
	Short:        "Bind a workload cluster to a Klutch control plane.",
	Long:         "Runs the non-interactive helper workflow by default to connect a workload cluster to a Klutch control plane endpoint, or falls back to the interactive kube-bind flow.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if bindKlutchWorkloadInteractive {
			if strings.TrimSpace(bindKlutchWorkloadControlPlane) == "" {
				return fmt.Errorf("the --control-plane flag is required for interactive bind")
			}
			makeup.PrintInfo(fmt.Sprintf("Binding workload cluster to control plane %s (interactive kube-bind)...", bindKlutchWorkloadControlPlane))

			pluginCmd, err := bindcmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
			if err != nil {
				return err
			}

			pluginArgs := []string{bindKlutchWorkloadControlPlane}
			if bindKlutchWorkloadKubeconfig != "" {
				pluginArgs = append(pluginArgs, "--kubeconfig", bindKlutchWorkloadKubeconfig)
			}
			if bindKlutchWorkloadContext != "" {
				pluginArgs = append(pluginArgs, "--context", bindKlutchWorkloadContext)
			}
			if bindKlutchWorkloadOutput != "" {
				pluginArgs = append(pluginArgs, "-o", bindKlutchWorkloadOutput)
			}
			if bindKlutchWorkloadDryRun {
				pluginArgs = append(pluginArgs, "--dry-run")
			}
			if bindKlutchWorkloadSkipKonnector {
				pluginArgs = append(pluginArgs, "--skip-konnector")
			}
			if bindKlutchWorkloadKonnectorImage != "" {
				pluginArgs = append(pluginArgs, "--konnector-image", bindKlutchWorkloadKonnectorImage)
			}
			if len(bindKlutchWorkloadExtraArgs) > 0 {
				pluginArgs = append(pluginArgs, bindKlutchWorkloadExtraArgs...)
			}

			pluginCmd.SetArgs(pluginArgs)
			return pluginCmd.Execute()
		}

		opts := klutch.NonInteractiveBindOptions{
			ControlPlaneURL:         strings.TrimSpace(bindKlutchWorkloadControlPlane),
			BindRequestPath:         bindKlutchWorkloadRequestFile,
			OIDCClientID:            bindKlutchWorkloadOIDCClientID,
			OIDCClientSecret:        bindKlutchWorkloadOIDCClientSecret,
			OIDCTokenURL:            bindKlutchWorkloadOIDCTokenURL,
			OIDCScope:               bindKlutchWorkloadOIDCScope,
			KonnectorImage:          bindKlutchWorkloadKonnectorImage,
			WriteKubeconfigTo:       bindKlutchWorkloadWriteKubeconfig,
			WorkloadKubeconfigPath:  bindKlutchWorkloadKubeconfig,
			WorkloadContext:         bindKlutchWorkloadContext,
			ControlPlaneClusterName: strings.TrimSpace(bindKlutchWorkloadControlPlaneCluster),
		}

		if strings.TrimSpace(bindKlutchWorkloadTenantName) != "" || strings.TrimSpace(bindKlutchWorkloadTenantSecretName) != "" {
			region := strings.TrimSpace(bindKlutchWorkloadTenantRegion)
			if region == "" {
				region = klutchaws.ControlPlaneDefaultRegion()
			}
			secretName := klutchaws.TenantSecretName(bindKlutchWorkloadTenantName, bindKlutchWorkloadTenantSecretName)
			conn, err := klutchaws.GetTenantCredentials(cmd.Context(), region, secretName)
			if err != nil {
				return fmt.Errorf("failed to load tenant secret %s in %s: %w", secretName, region, err)
			}
			if strings.TrimSpace(conn.BindURL) == "" {
				return fmt.Errorf("tenant secret %s is missing bind_url; recreate the tenant or set --control-plane explicitly", secretName)
			}
			if opts.ControlPlaneURL == "" && strings.TrimSpace(conn.BindURL) != "" {
				opts.ControlPlaneURL = strings.TrimSpace(conn.BindURL)
			}
			if opts.OIDCClientID == "" {
				opts.OIDCClientID = conn.ClientID
			}
			if opts.OIDCClientSecret == "" {
				opts.OIDCClientSecret = conn.ClientSecret
			}
			if opts.OIDCTokenURL == "" {
				opts.OIDCTokenURL = conn.TokenURL
			}
			if strings.TrimSpace(opts.OIDCScope) == "" {
				opts.OIDCScope = conn.Scope
			}
			if len(opts.BindRequestData) == 0 && strings.TrimSpace(conn.BindRequest) == "" {
				return fmt.Errorf("tenant secret %s is missing bind_request; regenerate the tenant or supply --bind-request-file", secretName)
			}
			if len(opts.BindRequestData) == 0 && strings.TrimSpace(conn.BindRequest) != "" {
				opts.BindRequestData = []byte(conn.BindRequest)
			}
			if opts.OIDCClientID == "" || opts.OIDCClientSecret == "" || opts.OIDCTokenURL == "" {
				return fmt.Errorf("tenant secret %s is missing required OIDC fields (client_id/client_secret/token_url)", secretName)
			}
			makeup.PrintInfo(fmt.Sprintf("Using tenant OIDC client %s (issuer %s)", opts.OIDCClientID, conn.IssuerURL))
		}

		if opts.ControlPlaneURL == "" {
			return fmt.Errorf("control-plane URL is required (flag --control-plane or bind_url in tenant secret)")
		}
		if len(opts.BindRequestData) == 0 && strings.TrimSpace(opts.BindRequestPath) == "" {
			return fmt.Errorf("bind request is required (from tenant secret or --bind-request-file)")
		}
		if len(opts.BindRequestData) > 0 {
			if err := klutch.ValidateBindRequest(opts.BindRequestData); err != nil {
				return err
			}
		}

		makeup.PrintInfo(fmt.Sprintf("Binding workload cluster to control plane %s (non-interactive)...", opts.ControlPlaneURL))
		return klutch.NonInteractiveBind(cmd.Context(), opts)
	},
}

func init() {
	initFlagsBindKlutchWorkload(cmdBindKlutchWorkload)
	cmdBindKlutch.AddCommand(cmdBindKlutchWorkload)
	bindCmd.AddCommand(cmdBindKlutch)
	rootCmd.AddCommand(bindCmd)
}

func initFlagsBindKlutchWorkload(cmd *cobra.Command) {
	initRequiredStringFlagWithDependency(cmd, &bindKlutchWorkloadControlPlane, "control-plane", "", "Klutch control plane bind endpoint (e.g. https://klutch-bind.example.com/exports).", &bindKlutchWorkloadInteractive, "interactive-bind", true)
	cmd.Flags().StringVar(&bindKlutchWorkloadControlPlaneCluster, "control-plane-cluster", "", "Control plane cluster name for CA lookup (defaults to klutch-control-plane).")
	cmd.Flags().StringVar(&bindKlutchWorkloadKubeconfig, "kubeconfig", "", "Path to the workload cluster kubeconfig.")
	cmd.Flags().StringVar(&bindKlutchWorkloadContext, "context", "", "Workload cluster kubeconfig context to use.")
	cmd.Flags().StringVarP(&bindKlutchWorkloadOutput, "output", "o", "", "Output format passed to kube-bind (e.g. yaml).")
	cmd.Flags().BoolVar(&bindKlutchWorkloadDryRun, "dry-run", false, "Pass through --dry-run to kube-bind.")
	cmd.Flags().BoolVar(&bindKlutchWorkloadSkipKonnector, "skip-konnector", false, "Skip deployment of the konnector (pass-through to kube-bind).")
	cmd.Flags().StringVar(&bindKlutchWorkloadKonnectorImage, "konnector-image", "", "Override the konnector image (pass-through to kube-bind).")
	cmd.Flags().StringSliceVar(&bindKlutchWorkloadExtraArgs, "bind-arg", []string{}, "Additional arguments to pass through to kube-bind.")
	cmd.Flags().BoolVar(&bindKlutchWorkloadInteractive, "interactive-bind", false, "Use the interactive kube-bind flow (opens browser). Default is non-interactive helper flow.")
	cmd.Flags().StringVar(&bindKlutchWorkloadRequestFile, "bind-request-file", "", "Path to JSON bind request (clusterID and apis) for non-interactive flow; defaults to the value stored in the tenant secret.")
	cmd.Flags().StringVar(&bindKlutchWorkloadOIDCClientID, "oidc-client-id", "", "OIDC client ID for non-interactive flow (defaults to OIDC_CLIENT_ID).")
	cmd.Flags().StringVar(&bindKlutchWorkloadOIDCClientSecret, "oidc-client-secret", "", "OIDC client secret for non-interactive flow (defaults to OIDC_CLIENT_SECRET).")
	cmd.Flags().StringVar(&bindKlutchWorkloadOIDCTokenURL, "oidc-token-url", "", "OIDC token URL for non-interactive flow (defaults to OIDC_TOKEN_URL).")
	cmd.Flags().StringVar(&bindKlutchWorkloadOIDCScope, "oidc-scope", "", "OIDC scopes for non-interactive flow (defaults to tenant scope).")
	cmd.Flags().StringVar(&bindKlutchWorkloadWriteKubeconfig, "write-kubeconfig", "", "Optional path to write control-plane kubeconfig returned by backend.")
	cmd.Flags().StringVar(&bindKlutchWorkloadTenantName, "tenant-name", "", "Klutch tenant name whose secret holds OIDC credentials.")
	cmd.Flags().StringVar(&bindKlutchWorkloadTenantSecretName, "tenant-secret-name", "", "Explicit Secrets Manager name for the tenant credentials (defaults to klutch/<tenant>/oidc-client).")
	cmd.Flags().StringVar(&bindKlutchWorkloadTenantRegion, "tenant-region", "", "AWS region for the tenant secret (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
}
