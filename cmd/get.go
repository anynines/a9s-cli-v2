package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/anynines/a9s-cli-v2/k8s"
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
		return printKlutchClusters()
	},
}

// a9s get clusters klutch
var getClustersKlutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "List Klutch clusters (control-plane and workload) from kubectl contexts/clusters.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return printKlutchClusters()
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
		pattern := regexp.MustCompile("klutch/(.+)/oidc-client")
		for _, t := range tenants {
			tenantName := ""
			matches := pattern.FindStringSubmatch(t)
			if len(matches) < 2 {
				tenantName = t
			} else {
				tenantName = matches[1]
			}
			makeup.Print("- Name:    " + tenantName)
			makeup.Print("  Secret:  " + t)
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
	// synonym: a9s get clusters klutch
	getClustersCmd := &cobra.Command{
		Use:   "clusters",
		Short: "List clusters",
		Run: func(cmd *cobra.Command, args []string) {
			makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
			_ = cmd.Help()
		},
	}
	getClustersCmd.AddCommand(getClustersKlutchCmd)
	getCmd.AddCommand(getClustersCmd)
	rootCmd.AddCommand(getCmd)
}

func printKlutchClusters() error {
	makeup.PrintInfo("Detecting Klutch clusters...")

	contexts, err := k8s.Contexts("klutch")
	if err != nil {
		return fmt.Errorf("failed to list kubectl contexts: %w", err)
	}
	clusters, err := k8s.Clusters("klutch")
	if err != nil {
		return fmt.Errorf("failed to list kubectl clusters: %w", err)
	}

	matches := append(contexts, clusters...)
	if len(matches) == 0 {
		makeup.PrintInfo("No Klutch-related kubectl contexts or clusters found.")
		return nil
	}

	var rows []clusterRow
	for _, m := range matches {
		clusterName := deriveClusterName(m)
		status := "inactive"
		// Quick existence check with low timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if _, err := k8s.ClusterInfo(ctx, m); err == nil {
			status = "active"
		}
		cancel()
		rows = append(rows, clusterRow{
			Cluster: clusterName,
			Status:  status,
		})
	}

	makeup.PrintH1("Klutch clusters detected in kubectl config")
	makeup.Print(renderClusterTable(rows))
	makeup.PrintInfo("Tip: switch kubectl context with `a9s use klutch --cluster-name <name>` or `a9s use cluster klutch -c <name>`.")
	return nil
}

// deriveClusterName tries to extract the bare cluster name from a context/cluster string.
func deriveClusterName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}

type clusterRow struct {
	Cluster string
	Status  string
}

func renderClusterTable(rows []clusterRow) string {
	if len(rows) == 0 {
		return ""
	}
	headers := []string{"Cluster", "Status"}
	data := make([][]string, 0, len(rows))
	for _, r := range rows {
		icon := "⏳"
		if strings.EqualFold(r.Status, "active") {
			icon = "✅"
		}
		data = append(data, []string{r.Cluster, fmt.Sprintf("%s %s", icon, r.Status)})
	}
	// compute widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range data {
		for i, col := range row {
			if len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e4833e")).Padding(0, 1)
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D9DCCF")).Padding(0, 1)
	divider := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color("#505d7a"))

	renderRow := func(cols []string, style lipgloss.Style) string {
		rendered := make([]string, len(cols))
		for i, c := range cols {
			rendered[i] = lipgloss.NewStyle().Width(widths[i]).Render(c)
		}
		return style.Render(lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
	}

	lines := []string{
		renderRow(headers, headerStyle),
		divider.Render(strings.Repeat(" ", sum(widths)+len(widths))),
	}
	for _, row := range data {
		lines = append(lines, renderRow(row, rowStyle))
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#505d7a")).
		Padding(0, 1)
	return box.Render(strings.Join(lines, "\n"))
}

func sum(ints []int) int {
	total := 0
	for _, v := range ints {
		total += v
	}
	return total
}
