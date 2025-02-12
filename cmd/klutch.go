package cmd

import (
	"os"

	"github.com/anynines/a9s-cli-v2/klutch"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var klutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "Klutch related commands",
	Long:  "Commands for deploying and interacting with Klutch",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		cmd.Help()
	},
}

var deployKlutchCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Klutch Control Plane Cluster",
	Long: `Deploys a Kind cluster which serves as the Klutch Control Plane Cluster.
It includes components such as crossplane, the klutch-bind backend and the a8s-framework.
Additionally deploys an App Cluster with Kind which can be used to bind to the Control Plane Cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		var port = klutch.PortFlag
		if port < 1 || port > 65535 {
			makeup.PrintFail("Invalid port number. Must be between 1 and 65535.")
			os.Exit(1)
		}

		klutch.DeployKlutchClusters()
	},
}

var bindKlutchCmd = &cobra.Command{
	Use:   "bind",
	Short: "Bind from an App Cluster to the Control Plane Cluster",
	Long:  `Starts the binding process to the Control Plane Cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		klutch.Bind()
	},
}

var deleteKlutchCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the klutch clusters",
	Long:  `Deletes the Control Plane and App kind clusters.`,
	Run: func(cmd *cobra.Command, args []string) {
		klutch.DeleteClusters()
	},
}

func init() {
	deployKlutchCmd.Flags().IntVar(&klutch.PortFlag, "port", 8080, "Port the Control Plane Cluster backend should listen on")
	deployKlutchCmd.Flags().StringVar(&klutch.KeycloakCaPathFlag, "keycloak-ca-path", "./keycloak-ca.crt", "Path to the keycloak CA file")

	deployKlutchCmd.Flags().StringVar(&klutch.BackendClientIdFlag, "backend-client-id", "backend", "OIDC client ID for the backend")
	deployKlutchCmd.Flags().StringVar(&klutch.BackendClientSecretFlag, "backend-client-secret", "", "OIDC client secret for the backend")
	deployKlutchCmd.Flags().StringVar(&klutch.BackendIssuerUrlFlag, "backend-issuer-url", "", "OIDC issuer URL for the backend")

	deployKlutchCmd.Flags().BoolVar(&klutch.IdTokenModeFlag, "enable-id-token-mode", false, "Whether to deploy the control plane cluster in Id Token or not.")

	deployKlutchCmd.Flags().StringVar(&klutch.LoadKonnectorImageFlag, "load-konnector-image", "", "konnector image to load into the App Cluster")
	deployKlutchCmd.Flags().StringVar(&klutch.LoadBackendImageFlag, "load-backend-image", "", "Backend image to load into the Control Plane Cluster")

	deployKlutchCmd.Flags().StringVar(&klutch.OIDCClusterClientIDFlag, "cluster-client-id", "", "OIDC Client ID for the control plane cluster")
	deployKlutchCmd.Flags().StringVar(&klutch.OIDCClusterIssuerURLFlag, "cluster-issuer", "", "OIDC issuer URL for the control plane cluster")

	// deployKlutchCmd.Flags().BoolVar(&klutch.KindClusterOnlyFlag, "deploy-kind-only", false, "Only deploy the control plane kind cluster with nothing on it")

	klutchCmd.AddCommand(deployKlutchCmd)

	klutchCmd.AddCommand(bindKlutchCmd)

	klutchCmd.AddCommand(deleteKlutchCmd)

	rootCmd.AddCommand(klutchCmd)
}
