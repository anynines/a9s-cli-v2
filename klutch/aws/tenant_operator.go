package aws

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
)

// deployTenantOperator installs the tenant-operator via Helm (OCI chart).
func deployTenantOperator(ctx context.Context, cfg Config, accountID string) {
	if strings.TrimSpace(cfg.TenantOperatorChart) == "" {
		makeup.PrintWarning("No tenant operator chart specified; skipping tenant operator deployment.")
		return
	}

	chart := strings.TrimSpace(cfg.TenantOperatorChart)
	version := strings.TrimSpace(cfg.TenantOperatorChartVersion)
	// If the chart ref includes a tag (oci://...:x.y.z), split it into chart+version to avoid invalid reference errors.
	if strings.HasPrefix(chart, "oci://") {
		if parts := strings.Split(chart, ":"); len(parts) > 2 {
			version = parts[len(parts)-1]
			chart = strings.Join(parts[:len(parts)-1], ":")
		} else if len(parts) == 2 && strings.Contains(parts[1], "/") == false {
			// oci://repo:tag (unlikely), handle gracefully
			version = parts[1]
			chart = parts[0]
		}
	}

	imageRepo := cfg.TenantOperatorImage
	if idx := strings.LastIndex(imageRepo, ":"); idx > 0 {
		// split tag if present
	}

	ns := "a9s-tenants-operator-system"
	release := "a9s-tenants-operator"

	ensureHelmRegistryLogin(ctx, cfg)
	if err := verifyTenantOperatorArtifacts(ctx, cfg, chart, version); err != nil {
		makeup.ExitDueToFatalError(err, err.Error())
	}

	args := []string{
		"helm", "upgrade", "--install", release, chart,
		"--namespace", ns, "--create-namespace",
	}
	if version != "" {
		args = append(args, "--version", version)
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

	// Build a values file for config to avoid --set parsing issues with JSON (commas/quotes).
	var valuesBuf bytes.Buffer
	valuesBuf.WriteString("config:\n")
	region := strings.TrimSpace(cfg.TenantOperatorRegion)
	if region == "" {
		region = strings.TrimSpace(cfg.Region)
	}
	if region != "" {
		fmt.Fprintf(&valuesBuf, "  region: \"%s\"\n", region)
		// Also pass via --set to ensure region is applied even if the values file is ignored.
		args = append(args, "--set", fmt.Sprintf("config.region=%s", region))
	}
	if strings.TrimSpace(cfg.TenantOperatorBindURL) != "" {
		fmt.Fprintf(&valuesBuf, "  bindURL: \"%s\"\n", cfg.TenantOperatorBindURL)
	}
	if strings.TrimSpace(cfg.TenantOperatorBindRequest) != "" {
		valuesBuf.WriteString("  bindRequest: |\n")
		for _, line := range strings.Split(cfg.TenantOperatorBindRequest, "\n") {
			fmt.Fprintf(&valuesBuf, "    %s\n", line)
		}
	}
	tmpValues := fmt.Sprintf("/tmp/tenant-operator-values-%d.yaml", time.Now().UnixNano())
	if err := os.WriteFile(tmpValues, valuesBuf.Bytes(), 0600); err == nil {
		args = append(args, "-f", tmpValues)
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

func verifyTenantOperatorArtifacts(ctx context.Context, cfg Config, chart, version string) error {
	if err := verifyTenantOperatorChart(ctx, chart, version); err != nil {
		return err
	}
	if err := verifyTenantOperatorImage(ctx, cfg); err != nil {
		return err
	}
	return nil
}

func verifyTenantOperatorChart(ctx context.Context, chart, version string) error {
	if strings.TrimSpace(chart) == "" {
		return nil
	}
	args := []string{"show", "chart", chart}
	if strings.TrimSpace(version) != "" {
		args = append(args, "--version", version)
	}
	_, errOut, err := runCmd(ctx, "helm", args...)
	if err == nil {
		return nil
	}
	if isOCIChartDescriptorError(errOut) {
		if version == "" {
			return fmt.Errorf("Tenant operator chart %q is not a Helm chart artifact in the OCI registry (manifest has only one descriptor). Ensure the chart is pushed to the registry (not a container image).", chart)
		}
		return fmt.Errorf("Tenant operator chart %q (version %s) is not a Helm chart artifact in the OCI registry (manifest has only one descriptor). Ensure the chart is pushed to the registry (not a container image).", chart, version)
	}
	if version == "" {
		return fmt.Errorf("Tenant operator chart %q not found or not accessible via Helm.\nstderr: %s", chart, strings.TrimSpace(errOut))
	}
	return fmt.Errorf("Tenant operator chart %q (version %s) not found or not accessible via Helm.\nstderr: %s", chart, version, strings.TrimSpace(errOut))
}

func verifyTenantOperatorImage(ctx context.Context, cfg Config) error {
	ref := parseImageRef(cfg.TenantOperatorImage)
	if ref.original == "" {
		return nil
	}
	if ref.repository == "" {
		return fmt.Errorf("Tenant operator image reference %q is invalid (missing repository).", ref.original)
	}
	if ref.registry != "" && strings.Contains(ref.registry, ".ecr.") {
		region := strings.TrimSpace(cfg.TenantOperatorRegion)
		if region == "" {
			region = strings.TrimSpace(cfg.Region)
		}
		if region == "" {
			region = ecrRegionFromHost(ref.registry)
		}
		if region == "" {
			return fmt.Errorf("Unable to determine AWS region for tenant operator image %q. Set --tenant-operator-region or AWS_REGION.", ref.original)
		}
		args := []string{"ecr", "describe-images", "--region", region, "--repository-name", ref.repository, "--image-ids"}
		if ref.digest != "" {
			args = append(args, fmt.Sprintf("imageDigest=%s", ref.digest))
		} else {
			args = append(args, fmt.Sprintf("imageTag=%s", ref.tag))
		}
		_, errOut, err := runCmd(ctx, "aws", args...)
		if err == nil {
			return nil
		}
		if strings.Contains(errOut, "RepositoryNotFoundException") {
			return fmt.Errorf("Tenant operator image repository %q not found in ECR registry %s. Create the repository or update --tenant-operator-image.", ref.repository, ref.registry)
		}
		if strings.Contains(errOut, "ImageNotFoundException") {
			return fmt.Errorf("Tenant operator image %q not found in ECR (tag/digest missing). Push the image or update --tenant-operator-image.", ref.original)
		}
		return fmt.Errorf("Failed to verify tenant operator image %q in ECR.\nstderr: %s", ref.original, strings.TrimSpace(errOut))
	}

	if _, err := execLookPath("docker"); err != nil {
		return fmt.Errorf("Unable to verify tenant operator image %q because Docker is not installed. Install Docker or use an ECR image so the CLI can verify it.", ref.original)
	}
	_, errOut, err := runCmd(ctx, "docker", "manifest", "inspect", ref.original)
	if err != nil {
		if strings.TrimSpace(errOut) == "" {
			errOut = err.Error()
		}
		return fmt.Errorf("Tenant operator image %q not found or not accessible via Docker.\nstderr: %s", ref.original, strings.TrimSpace(errOut))
	}
	return nil
}

type imageRef struct {
	original   string
	registry   string
	repository string
	tag        string
	digest     string
}

func parseImageRef(ref string) imageRef {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return imageRef{}
	}
	out := imageRef{original: trimmed}

	base := trimmed
	if at := strings.Index(base, "@"); at != -1 {
		out.digest = base[at+1:]
		base = base[:at]
	}

	tag := ""
	if colon := strings.LastIndex(base, ":"); colon != -1 {
		slash := strings.LastIndex(base, "/")
		if colon > slash {
			tag = base[colon+1:]
			base = base[:colon]
		}
	}
	if tag == "" && out.digest == "" {
		tag = "latest"
	}
	out.tag = tag

	parts := strings.Split(base, "/")
	if len(parts) == 1 {
		out.repository = base
		return out
	}
	if looksLikeRegistryHost(parts[0]) {
		out.registry = parts[0]
		out.repository = strings.Join(parts[1:], "/")
		return out
	}
	out.repository = base
	return out
}

func looksLikeRegistryHost(host string) bool {
	return strings.Contains(host, ".") || strings.Contains(host, ":") || host == "localhost"
}

func ecrRegionFromHost(host string) string {
	parts := strings.SplitN(host, ".ecr.", 2)
	if len(parts) != 2 {
		return ""
	}
	region := strings.TrimSuffix(parts[1], ".amazonaws.com")
	region = strings.TrimSuffix(region, ".amazonaws.com.cn")
	return region
}

func isOCIChartDescriptorError(errOut string) bool {
	return strings.Contains(errOut, "manifest does not contain minimum number of descriptors")
}
