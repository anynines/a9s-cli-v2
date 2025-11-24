package cmd

import (
	"github.com/anynines/a9s-cli-v2/klutch"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var (
	applyKlutchControlPlaneHost       string
	applyKlutchControlPlanePort       int
	applyKlutchControlPlaneACMCertARN string
	applyKlutchControlPlaneHostedZone string
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

var applyKlutchControlPlaneCmd = &cobra.Command{
	Use:   "klutch-control-plane",
	Short: "Install the Klutch control plane onto the current Kubernetes cluster.",
	Long:  "Installs the Klutch control plane and its dependencies into the currently selected kube context.",
	Run: func(cmd *cobra.Command, args []string) {
		if applyKlutchControlPlanePort < 1 || applyKlutchControlPlanePort > 65535 {
			makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
		}

		klutch.ApplyKlutchControlPlane(applyKlutchControlPlaneHost, applyKlutchControlPlanePort, applyKlutchControlPlaneACMCertARN, applyKlutchControlPlaneHostedZone)
	},
}

func init() {
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchControlPlaneHost, "host", "", "Host (IP or DNS name) to reach the ingress. Defaults to the Kubernetes API server host of the current kube context.")
	applyKlutchControlPlaneCmd.Flags().IntVar(&applyKlutchControlPlanePort, "ingress-port", 443, "Port the ingress should listen on.")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchControlPlaneACMCertARN, "acm-certificate-arn", "", "ACM certificate ARN to enable HTTPS on the ALB ingress for Dex.")
	applyKlutchControlPlaneCmd.Flags().StringVar(&applyKlutchControlPlaneHostedZone, "hosted-zone-name", "", "Route53 hosted zone name (FQDN). If provided and no ACM ARN is supplied, the CLI will request an ACM cert and create DNS validation records automatically.")

	applyCmd.AddCommand(applyKlutchControlPlaneCmd)
	rootCmd.AddCommand(applyCmd)
}
