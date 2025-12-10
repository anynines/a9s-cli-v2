package aws

import (
	"bytes"
	"context"
	"fmt"
	"os"
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

	ensureHelmRegistryLogin(ctx, cfg)

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
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	// Ensure OCI support is on for Helm (required for ECR-hosted charts).
	cmd.Env = append(os.Environ(), "HELM_EXPERIMENTAL_OCI=1")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to deploy tenant operator.\nstderr: %s", errBuf.String()))
	}
	if makeup.Verbose {
		makeup.Print(outBuf.String())
	}
	waitForTenantOperator(ctx, ns)
}

// ensureHelmRegistryLogin performs a helm registry login for ECR-backed OCI charts to avoid 403 token expiry.
func ensureHelmRegistryLogin(ctx context.Context, cfg Config) {
	chart := strings.TrimSpace(cfg.TenantOperatorChart)
	if !strings.HasPrefix(chart, "oci://") {
		return
	}
	trimmed := strings.TrimPrefix(chart, "oci://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 {
		return
	}
	host := parts[0]
	if !strings.Contains(host, ".ecr.") {
		return
	}
	region := strings.TrimSpace(cfg.TenantOperatorRegion)
	if region == "" {
		region = strings.TrimSpace(cfg.Region)
	}
	if region == "" {
		region = "eu-central-1"
	}

	makeup.PrintInfo(fmt.Sprintf("Logging into Helm OCI registry %s via ECR...", host))
	passCmd := exec.CommandContext(ctx, "aws", "ecr", "get-login-password", "--region", region)
	var pwBuf bytes.Buffer
	passCmd.Stdout = &pwBuf
	if err := passCmd.Run(); err != nil {
		makeup.PrintWarning(fmt.Sprintf("Failed to obtain ECR login password for %s: %v", host, err))
		return
	}

	loginCmd := exec.CommandContext(ctx, "helm", "registry", "login", host, "--username", "AWS", "--password-stdin")
	loginCmd.Stdin = bytes.NewReader(pwBuf.Bytes())
	var outBuf, errBuf bytes.Buffer
	loginCmd.Stdout = &outBuf
	loginCmd.Stderr = &errBuf
	if err := loginCmd.Run(); err != nil {
		makeup.PrintWarning(fmt.Sprintf("Helm registry login to %s failed: %v\nstderr: %s", host, err, strings.TrimSpace(errBuf.String())))
		return
	}
	if makeup.Verbose {
		makeup.Print(outBuf.String())
	}
	makeup.PrintInfo(fmt.Sprintf("Helm registry login to %s succeeded.", host))
}

// waitForTenantOperator waits for the tenant operator deployment to become ready.
func waitForTenantOperator(ctx context.Context, namespace string) {
	makeup.PrintInfo("Waiting for tenant operator deployment to become ready...")
	cmd := exec.CommandContext(ctx, "kubectl", "-n", namespace, "rollout", "status", "deployment/a9s-tenants-operator", "--timeout=2m")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Tenant operator did not become ready.\nstderr: %s", strings.TrimSpace(errBuf.String())))
	}
	if makeup.Verbose {
		makeup.Print(outBuf.String())
	}
	makeup.PrintSuccess("Tenant operator is ready.")
}
