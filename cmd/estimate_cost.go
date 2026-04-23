package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anynines/a9s-cli-v2/cost/klutchaws"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var (
	estimateProvider      string
	estimateRegion        string
	estimateInstanceType  string
	estimateDesiredNodes  int
	estimateMinNodes      int
	estimateMaxNodes      int
	estimateNodeDiskGiB   int
	estimateNatGateways   int
	estimateOutput        string
	estimatePricingRegion string
)

var estimateCostCmd = &cobra.Command{
	Use:   "estimate-cost",
	Short: "Estimate infrastructure costs.",
	Long:  `Estimate infrastructure costs for supported stacks and providers.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" Please choose a resource to estimate.")
		cmd.Help()
	},
}

var estimateCostClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Estimate cluster-related costs.",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" Please choose a cluster type to estimate.")
		cmd.Help()
	},
}

var estimateCostClusterKlutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "Estimate the Klutch control plane cluster costs.",
	Long: `Pulls region-specific AWS list prices to calculate hourly and monthly costs for
the Klutch control plane BOM (EKS control plane, worker nodes, root volumes, NAT gateways, KMS key).
Outputs either a lipgloss-styled table (default) or JSON.`,
	Run: func(cmd *cobra.Command, args []string) {
		provider := strings.ToLower(strings.TrimSpace(estimateProvider))
		if provider == "" {
			makeup.ExitDueToFatalError(nil, "Please select a provider via -p. Supported provider for Klutch cost estimation is \"aws\".")
		}
		if provider != "aws" {
			makeup.ExitDueToFatalError(nil, "The Klutch cost estimation currently only supports the \"aws\" provider.")
		}

		cfg := klutchaws.EstimateConfig{
			Region:        estimateRegion,
			InstanceType:  estimateInstanceType,
			DesiredNodes:  estimateDesiredNodes,
			MinNodes:      estimateMinNodes,
			MaxNodes:      estimateMaxNodes,
			NodeDiskGiB:   estimateNodeDiskGiB,
			NatGateways:   estimateNatGateways,
			PricingRegion: estimatePricingRegion,
		}

		report, err := klutchaws.Estimate(context.Background(), cfg)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to estimate Klutch control plane costs.")
		}

		switch strings.ToLower(strings.TrimSpace(estimateOutput)) {
		case "", "table":
			fmt.Println(makeup.RenderCostTable(*report))
			if len(report.Assumptions) > 0 {
				fmt.Println(makeup.H2("Assumptions"))
				for _, a := range report.Assumptions {
					makeup.Print(" - " + a)
				}
			}
			if len(report.NotIncluded) > 0 {
				fmt.Println(makeup.H2("Not included"))
				for _, n := range report.NotIncluded {
					makeup.Print(" - " + n)
				}
			}
		case "json":
			b, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				makeup.ExitDueToFatalError(err, "Failed to marshal estimation to JSON.")
			}
			fmt.Println(string(b))
		default:
			makeup.ExitDueToFatalError(nil, "Unsupported output format. Use \"table\" (default) or \"json\".")
		}
	},
}

func init() {
	defaultRegion := "eu-central-1"
	defaultInstanceType := "t3a.xlarge"

	initRequiredStringFlagP(estimateCostClusterKlutchCmd, &estimateProvider, "provider", "p", "", "Provider to use (only \"aws\" supported).")
	estimateCostClusterKlutchCmd.PersistentFlags().StringVar(&estimateRegion, "region", defaultRegion, "AWS region to price.")
	estimateCostClusterKlutchCmd.PersistentFlags().StringVar(&estimateInstanceType, "instance-type", defaultInstanceType, "EC2 instance type for worker nodes.")
	estimateCostClusterKlutchCmd.PersistentFlags().IntVar(&estimateDesiredNodes, "desired-nodes", 3, "Desired number of worker nodes.")
	estimateCostClusterKlutchCmd.PersistentFlags().IntVar(&estimateMinNodes, "min-nodes", 3, "Minimum number of worker nodes.")
	estimateCostClusterKlutchCmd.PersistentFlags().IntVar(&estimateMaxNodes, "max-nodes", 5, "Maximum number of worker nodes.")
	estimateCostClusterKlutchCmd.PersistentFlags().IntVar(&estimateNodeDiskGiB, "node-disk-gib", 80, "Root volume size per node in GiB (gp3).")
	estimateCostClusterKlutchCmd.PersistentFlags().IntVar(&estimateNatGateways, "nat-gateways", 3, "Number of NAT gateways to include in the estimate.")
	estimateCostClusterKlutchCmd.PersistentFlags().StringVarP(&estimateOutput, "output", "o", "table", "Output format: \"table\" (default) or \"json\".")
	estimateCostClusterKlutchCmd.PersistentFlags().StringVar(&estimatePricingRegion, "pricing-region", "us-east-1", "AWS region used for the Pricing API (defaults to us-east-1).")

	estimateCostClusterCmd.AddCommand(estimateCostClusterKlutchCmd)
	estimateCostCmd.AddCommand(estimateCostClusterCmd)
	rootCmd.AddCommand(estimateCostCmd)
}
