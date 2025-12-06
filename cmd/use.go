package cmd

import (
	"fmt"
	"os/exec"
	"strings"

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

		contexts, err := listKubectlContexts()
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
		if err := exec.Command("kubectl", "config", "use-context", target).Run(); err != nil {
			return fmt.Errorf("failed to switch kubectl context to %s: %w", target, err)
		}
		makeup.PrintSuccessSummary(fmt.Sprintf("kubectl context switched to %s", target))
		return nil
	},
}

func listKubectlContexts() ([]string, error) {
	out, err := exec.Command("kubectl", "config", "get-contexts", "-o", "name").Output()
	if err != nil {
		return nil, fmt.Errorf("kubectl config get-contexts failed: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var contexts []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			contexts = append(contexts, strings.TrimSpace(l))
		}
	}
	return contexts, nil
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

func init() {
	useKlutchCmd.Flags().StringVarP(&useKlutchClusterName, "cluster-name", "c", "", "Klutch cluster name to switch to.")

	useCmd.AddCommand(useKlutchCmd)
	rootCmd.AddCommand(useCmd)
}
