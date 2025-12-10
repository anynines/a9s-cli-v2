package aws

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/anynines/a9s-cli-v2/makeup"
)

// deployTenantOperator installs the tenant-operator via Helm (OCI chart).
func deployTenantOperator(ctx context.Context, cfg Config, accountID string) {
	if strings.TrimSpace(cfg.TenantOperatorChart) == "" {
		makeup.PrintWarning("No tenant operator chart specified; skipping tenant operator deployment.")
		return
	}

	imageRepo := cfg.TenantOperatorImage
	if idx := strings.LastIndex(imageRepo, ":"); idx > 0 {
		// split tag if present
	}

	ns := "a9s-tenants-operator-system"
	release := "a9s-tenants-operator"

	args := []string{
		"helm", "upgrade", "--install", release, cfg.TenantOperatorChart,
		"--namespace", ns, "--create-namespace",
	}
	if cfg.TenantOperatorImage != "" {
		tag := "latest"
		repo := cfg.TenantOperatorImage
		if parts := strings.Split(cfg.TenantOperatorImage, ":"); len(parts) == 2 {
			repo = parts[0]
			tag = parts[1]
		}
		args = append(args, "--set", fmt.Sprintf("image.repository=%s", repo))
		args = append(args, "--set", fmt.Sprintf("image.tag=%s", tag))
	}
	if cfg.TenantOperatorRoleARN != "" {
		escaped := strings.ReplaceAll(cfg.TenantOperatorRoleARN, "/", "\\/")
		args = append(args, "--set", fmt.Sprintf(`serviceAccount.annotations.eks\.amazonaws\.com/role-arn=%s`, escaped))
	}
	if cfg.TenantOperatorRegion != "" {
		args = append(args, "--set", fmt.Sprintf("config.region=%s", cfg.TenantOperatorRegion))
	} else {
		args = append(args, "--set", fmt.Sprintf("config.region=%s", cfg.Region))
	}
	if cfg.TenantOperatorBindURL != "" {
		args = append(args, "--set", fmt.Sprintf("config.bindURL=%s", cfg.TenantOperatorBindURL))
	}
	if cfg.TenantOperatorBindRequest != "" {
		args = append(args, "--set", fmt.Sprintf("config.bindRequest=%s", cfg.TenantOperatorBindRequest))
	}

	makeup.PrintInfo("Deploying tenant operator via Helm...")
	cmd := execCommandContext(ctx, args[0], args[1:]...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to deploy tenant operator.\nstderr: %s", errBuf.String()))
	}
	if makeup.Verbose {
		makeup.Print(outBuf.String())
	}
	makeup.PrintSuccess("Tenant operator deployed.")
}
