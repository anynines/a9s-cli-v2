package aws

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/k8s"
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
		} else if len(parts) == 2 && !strings.Contains(parts[1], "/") {
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
	if output, err := makeup.Command(args[0], args[1:]...).Env("HELM_EXPERIMENTAL_OCI=1").Ctx(ctx).WithPrompt().Run(); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to deploy tenant operator.\n"+string(output))
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
	ref := parseImageRef(trimmed)
	if ref.awsCommand == "" {
		return
	}
	region, _ := getFirstNonEmptyString(ref.awsRegion,
		strings.TrimSpace(cfg.TenantOperatorRegion),
		strings.TrimSpace(cfg.Region),
		"eu-central-1")

	makeup.PrintInfo(fmt.Sprintf("Logging into Helm OCI registry %s via ECR...", ref.registry))
	pwOutput, err := makeup.Command("aws", ref.awsCommand, "get-login-password", "--region", region).Ctx(ctx).NoPrompt().Run()
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to obtain ECR login password for %s:\n%s", ref.registry, pwOutput))
		return
	}

	if loginOutput, err := makeup.Command("helm", "registry", "login", ref.registry, "--username", "AWS", "--password-stdin").Ctx(ctx).Stdin([]byte(pwOutput)).NoPrompt().Run(); err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Helm registry login to %s failed: %v\nstderr: %s", ref.registry, err, strings.TrimSpace(string(loginOutput))))
	}
	makeup.PrintInfo(fmt.Sprintf("Helm registry login to %s succeeded.", ref.registry))
}

func getFirstNonEmptyString(arg string, args ...string) (string, error) {
	args = append([]string{arg}, args...)
	i := slices.IndexFunc(args, func(s string) bool {
		return s != ""
	})
	if i == -1 {
		return "", fmt.Errorf("All input strings were empty")
	}
	return args[i], nil
}

// waitForTenantOperator waits for the tenant operator deployment to become ready.
func waitForTenantOperator(ctx context.Context, namespace string) {
	makeup.PrintInfo("Waiting for tenant operator deployment to become ready...")
	k8sClient := k8s.NewKubeClient("")
	output, err := k8sClient.RolloutStatus("deployment", "a9s-tenants-operator", namespace, "--timeout=2m")
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Tenant operator did not become ready.\nstderr: %s", strings.TrimSpace(output)))
	}
	if makeup.Verbose {
		makeup.Print(output)
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
	errOut, err := runCmd(ctx, "helm", args...)
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
	if ref.awsCommand == "" {
		if _, err := execLookPath("docker"); err != nil {
			return fmt.Errorf("Unable to verify tenant operator image %q because Docker is not installed. Install Docker or use an ECR image so the CLI can verify it.", ref.original)
		}

		errOut, err := runCmd(ctx, "docker", "manifest", "inspect", ref.original)
		if err == nil {
			return nil
		}

		if strings.TrimSpace(errOut) == "" {
			errOut = err.Error()
		}
		return fmt.Errorf("Tenant operator image %q not found or not accessible via Docker.\nstderr: %s", ref.original, strings.TrimSpace(errOut))
	}

	region, err := getFirstNonEmptyString(ref.awsRegion,
		strings.TrimSpace(cfg.TenantOperatorRegion),
		strings.TrimSpace(cfg.Region),
		ecrRegionFromHost(ref.registry))
	if err != nil {
		return fmt.Errorf("Unable to determine AWS region for tenant operator image %q. Set --tenant-operator-region or AWS_REGION.", ref.original)
	}

	imageId := fmt.Sprintf("imageTag=%s", ref.tag)
	if ref.digest != "" {
		imageId = fmt.Sprintf("imageDigest=%s", ref.digest)
	}

	args := []string{ref.awsCommand, "describe-images",
		"--region", region,
		"--repository-name", ref.repository,
		"--image-ids", imageId}
	errOut, err := runCmd(ctx, "aws", args...)
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

type imageRef struct {
	original   string
	registry   string
	repository string
	tag        string
	digest     string
	awsCommand string
	awsRegion  string
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
	out.repository = base
	if !looksLikeRegistryHost(parts[0]) {
		return out
	}
	out.registry = parts[0]
	out.repository = strings.Join(parts[1:], "/")
	// Determine AWS ECR command based on registry type
	if !strings.Contains(out.registry, ".ecr.") {
		return out
	}
	if !strings.HasPrefix(out.registry, "public.") {
		out.awsCommand = "ecr"
		return out
	}
	out.awsCommand = "ecr-public"
	// since ecr-public is a global service the region must be set to us-east-1
	out.awsRegion = "us-east-1"
	parts = strings.Split(out.repository, "/")
	if len(parts) < 2 {
		err := fmt.Errorf("expected at least 2 substrings separated by a /, got %d", len(parts))
		awsLogger.Fatalf(err, "failed to trim registry prefix for public ecr image %q", ref)
	}
	out.repository = strings.Join(parts[1:], "/")
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
