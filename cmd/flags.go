package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

func markFlagsOneRequiredButMutuallyExclusive(cmd *cobra.Command, p1 *string, name1 string, p2 *string, name2 string) {
	cmd.MarkFlagsOneRequired(name1, name2)
	cmd.MarkFlagsMutuallyExclusive(name1, name2)
	prependToRun(cmd, func(funcCmd *cobra.Command, args []string) error {
		if p1 != nil {
			if strings.TrimSpace(*p1) == "" {
				// the flags are mutually exclusive so if flag1 is set then
				// flag2 can't be populated so we fail
				return fmt.Errorf("required flag %s must be set to a non-empty value", name1)
			}
			return nil
		}
		// the flags are required so if flag1 is unset then flag2 must be
		// populated
		if strings.TrimSpace(*p2) == "" {
			return fmt.Errorf("required flag %s must be set to a non-empty value", name2)
		}
		return nil
	})
}

func initRequiredStringFlag(cmd *cobra.Command, p *string, name string, value, usage string) {
	initRequiredStringFlagP(cmd, p, name, "", value, usage)
}

func initRequiredStringFlagP(cmd *cobra.Command, p *string, name, shorthand, value, usage string) {
	if shorthand == "" {
		cmd.Flags().StringVar(p, name, value, usage)
	} else {
		cmd.Flags().StringVarP(p, name, shorthand, value, usage)
	}
	if err := cmd.MarkFlagRequired(name); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to mark flag "+name+" as required")
	}
	prependToRun(cmd, func(funcCmd *cobra.Command, args []string) error {
		if *p == "" {
			return fmt.Errorf("required flag %s must not be empty", name)
		}
		return nil
	})
}

func initRequiredStringFlagWithDependency[T any](otherFlagValue *T, otherFlagName string, otherFlagExpectedValue T, cmd *cobra.Command, p *string, name string, value, usage string) {
	cmd.Flags().StringVar(p, name, value, usage)
	prependToRun(cmd, func(*cobra.Command, []string) error {
		if otherFlagValue != nil {
			if cmp.Equal(*otherFlagValue, otherFlagExpectedValue) {
				if p == nil {
					return fmt.Errorf("flag %s is required for %s=%v but is missing", name, otherFlagName, otherFlagExpectedValue)
				}
				if *p == "" {
					return fmt.Errorf("flag %s is required for %s=%v but is empty", name, otherFlagName, otherFlagExpectedValue)
				}
			}
		}
		return nil
	})
}

func prependToRun(cmd *cobra.Command, newRunFunc func(funcCmd *cobra.Command, args []string) error) {
	oldCmd := *cmd
	if oldCmd.RunE != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			if err := newRunFunc(cmd, args); err != nil {
				return err
			}
			return oldCmd.RunE(cmd, args)
		}
		return
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := newRunFunc(cmd, args); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		oldCmd.Run(cmd, args)
	}
}
