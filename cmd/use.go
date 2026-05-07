package cmd

import (
	"fmt"
	"strings"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
)

var (
	useKlutchClusterName string
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch local kubectl context",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		_ = cmd.Help()
	},
}

var useKlutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "Switch kubectl context to a Klutch cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(useKlutchClusterName) == "" {
			return fmt.Errorf("--cluster-name is required")
		}

		contexts, err := k8s.Contexts("")
		if err != nil {
			return err
		}
		if len(contexts) == 0 {
			return fmt.Errorf("no kubectl contexts found")
		}

		target := selectContext(contexts, useKlutchClusterName)
		if target == "" {
			return fmt.Errorf("no kubectl context found matching cluster name %q", useKlutchClusterName)
		}

		makeup.PrintInfo(fmt.Sprintf("Switching kubectl context to %s ...", target))
		if out, err := k8s.SwitchContext(target); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("failed to switch kubectl context to %s: %s", target, out))
		}
		makeup.PrintSuccessSummary(fmt.Sprintf("kubectl context switched to %s", target))
		return nil
	},
}

// selectContext picks the best-matching context for a cluster name.
// Prefers exact match; otherwise returns a single substring match; returns empty string on ambiguity or no match.
func selectContext(contexts []string, clusterName string) string {
	clean := strings.TrimSpace(clusterName)
	exact := ""
	subMatches := []string{}
	for _, c := range contexts {
		if c == clean {
			exact = c
			break
		}
		if strings.Contains(c, clean) {
			subMatches = append(subMatches, c)
		}
	}
	if exact != "" {
		return exact
	}
	if len(subMatches) == 1 {
		return subMatches[0]
	}
	// ambiguous or none
	return ""
}

var useClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Switch kubectl context to a cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a subcommand from the list below.")
		_ = cmd.Help()
	},
}
var useClusterKlutchCmd = &cobra.Command{
	Use:   "klutch",
	Short: "Switch kubectl context to a Klutch cluster.",
	RunE:  useKlutchCmd.RunE,
}

func init() {
	initRequiredStringFlagP(useClusterKlutchCmd, &useKlutchClusterName, "cluster-name", "c", "", "Klutch cluster name to switch to.")
	useClusterCmd.AddCommand(useClusterKlutchCmd)
	useCmd.AddCommand(useClusterCmd)

	initRequiredStringFlagP(useKlutchCmd, &useKlutchClusterName, "cluster-name", "c", "", "Klutch cluster name to switch to.")
	useCmd.AddCommand(useKlutchCmd)

	rootCmd.AddCommand(useCmd)
}
