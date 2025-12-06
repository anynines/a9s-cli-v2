package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var getKlutchTenantRegion string
var getKlutchTenantSecretName string
var getKlutchTenantPrefix string

// getKlutchClustersCmd lists Klutch-related kubectl contexts/clusters.
var getKlutchClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "List Klutch clusters (control-plane and workload) from kubectl contexts/clusters.",
	RunE: func(cmd *cobra.Command, args []string) error {
		contexts, err := listKubectlNames("config", "get-contexts", "-o", "name")
		if err != nil {
			return fmt.Errorf("failed to list kubectl contexts: %w", err)
		}
		clusters, err := listKubectlNames("config", "get-clusters")
		if err != nil {
			return fmt.Errorf("failed to list kubectl clusters: %w", err)
		}

		seen := map[string]struct{}{}
		var matches []string
		for _, n := range append(contexts, clusters...) {
			if strings.Contains(strings.ToLower(n), "klutch") {
				if _, ok := seen[n]; !ok {
					seen[n] = struct{}{}
					matches = append(matches, n)
				}
			}
		}

		if len(matches) == 0 {
			makeup.PrintInfo("No Klutch-related kubectl contexts or clusters found.")
			return nil
		}

		makeup.PrintH1("Klutch clusters detected in kubectl config")
		for _, m := range matches {
			makeup.Print(fmt.Sprintf("- %s", m))
		}
		return nil
	},
}

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
	getKlutchCmd.AddCommand(getKlutchClustersCmd)
	getCmd.AddCommand(getKlutchCmd)
	rootCmd.AddCommand(getCmd)
}

// listKubectlNames runs a kubectl subcommand and returns non-empty lines from stdout.
func listKubectlNames(args ...string) ([]string, error) {
	out, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var res []string
	for _, l := range lines {
		if s := strings.TrimSpace(l); s != "" {
			res = append(res, s)
		}
	}
	return res, nil
}
