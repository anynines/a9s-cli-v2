package cmd

import (
	"context"
	"fmt"
	"strings"

	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var getKlutchTenantRegion string
var getKlutchTenantSecretName string
var getKlutchTenantPrefix string

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources managed by the CLI.",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		cmd.Help()
	},
}

var getKlutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "Get Klutch resources.",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		cmd.Help()
	},
}

var getKlutchTenantsCmd = &cobra.Command{
	Use:   "tenants",
	Short: "List Klutch tenants (Cognito credentials in Secrets Manager).",
	Run: func(cmd *cobra.Command, args []string) {
		region := strings.TrimSpace(getKlutchTenantRegion)
		if region == "" {
			region = klutchaws.ControlPlaneDefaultRegion()
		}
		tenants, err := klutchaws.ListTenantSecrets(context.Background(), region, getKlutchTenantPrefix)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to list Klutch tenants.")
		}
		if len(tenants) == 0 {
			makeup.PrintInfo("No Klutch tenants found.")
			return
		}
		makeup.PrintH1("Klutch Tenants (Secrets Manager)")
		for _, t := range tenants {
			makeup.Print(fmt.Sprintf("- %s", t))
		}
	},
}

var getKlutchTenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Get Klutch tenant credentials.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tenantName := strings.TrimSpace(args[0])
		if tenantName == "" {
			makeup.ExitDueToFatalError(nil, "tenant name is required.")
		}
		region := strings.TrimSpace(getKlutchTenantRegion)
		if region == "" {
			region = klutchaws.ControlPlaneDefaultRegion()
		}
		secretName := klutchaws.TenantSecretName(tenantName, getKlutchTenantSecretName)

		conn, err := klutchaws.GetTenantCredentials(context.Background(), region, secretName)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to retrieve Klutch tenant credentials.")
		}

		makeup.PrintH1(fmt.Sprintf("Klutch Tenant: %s", tenantName))
		makeup.PrintInfo(fmt.Sprintf("Secret:       %s", secretName))
		makeup.PrintInfo(fmt.Sprintf("Issuer:       %s", conn.IssuerURL))
		makeup.PrintInfo(fmt.Sprintf("Client ID:    %s", conn.ClientID))
		makeup.PrintInfo(fmt.Sprintf("Client Secret:%s", conn.ClientSecret))
		makeup.PrintInfo(fmt.Sprintf("Scope:        %s", conn.Scope))
		makeup.PrintSuccessSummary("Distribute these credentials securely to the workload cluster owner.")
	},
}

func init() {
	getKlutchTenantsCmd.Flags().StringVar(&getKlutchTenantRegion, "region", "", "AWS region for Cognito/Secrets Manager (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	getKlutchTenantsCmd.Flags().StringVar(&getKlutchTenantPrefix, "prefix", "klutch/", "Secret name prefix to filter tenants.")
	getKlutchTenantCmd.Flags().StringVar(&getKlutchTenantRegion, "region", "", "AWS region for Cognito/Secrets Manager (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	getKlutchTenantCmd.Flags().StringVar(&getKlutchTenantSecretName, "secret-name", "", "Secrets Manager name that holds the tenant credentials (defaults to klutch/<tenant>/oidc-client).")

	getKlutchCmd.AddCommand(getKlutchTenantsCmd)
	getKlutchCmd.AddCommand(getKlutchTenantCmd)
	getCmd.AddCommand(getKlutchCmd)
	rootCmd.AddCommand(getCmd)
}
