package aws

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/util/json"
)

type Config struct {
	Region                          string
	ClusterName                     string
	NodegroupName                   string
	NodeInstanceTypes               string
	NodeScalingConfig               string
	NodeAMIType                     string
	ClusterRoleName                 string
	NodeRoleName                    string
	BaseCIDR                        string
	PubACIDR, PubBCIDR, PubCCIDR    string
	PrivACIDR, PrivBCIDR, PrivCCIDR string
	ALBControllerVersion            string
	ALBControllerPolicyName         string
	ControlPlaneSGName              string
	KlutchTagValue                  string
	ResourceNamePrefix              string
	ClusterRole                     string
	AlbServiceAccountName           string

	TenantOperatorImage        string
	TenantOperatorChart        string
	TenantOperatorChartVersion string
	TenantOperatorRoleARN      string
	TenantOperatorRegion       string
	TenantOperatorBindURL      string
	TenantOperatorBindRequest  string
	HostedZoneName             string
}

// CreateOptions configures Klutch cluster creation.
type CreateOptions struct {
	// DryRun prints the planned resources and commands without creating them.
	DryRun bool
	// ClusterName overrides the default name if set (used by workload clusters).
	ClusterName string
	// ControlPlaneToBindTo contains the name of the Control Plane to bind to when creating a Workload Cluster
	ControlPlaneToBindTo string
	// Node overrides for control-plane/workload clusters.
	NodeInstanceTypes string
	NodeCount         int
	// TenantOperator overrides (control-plane only).
	TenantOperatorImage        string
	TenantOperatorChart        string
	TenantOperatorChartVersion string
	TenantOperatorRoleARN      string
	TenantOperatorRegion       string
	TenantOperatorBindURL      string
	TenantOperatorBindRequest  string
	HostedZoneName             string
}

type styledLogger struct{}

func newStyledLogger() *styledLogger {
	return &styledLogger{}
}

func (l *styledLogger) Infof(format string, args ...interface{}) {
	makeup.PrintInfo(fmt.Sprintf(format, args...))
}

func (l *styledLogger) Successf(format string, args ...interface{}) {
	makeup.PrintSuccess(fmt.Sprintf(format, args...))
}

func (l *styledLogger) Warningf(format string, args ...interface{}) {
	makeup.PrintWarning(fmt.Sprintf(format, args...))
}

func (l *styledLogger) Printf(format string, args ...interface{}) {
	makeup.Print(fmt.Sprintf(format, args...))
}

func (l *styledLogger) Section(title string) {
	makeup.PrintH1(title)
}

func (l *styledLogger) Summaryf(format string, args ...interface{}) {
	makeup.PrintSuccessSummary(fmt.Sprintf(format, args...))
}

func (l *styledLogger) Fatalf(err error, format string, args ...interface{}) {
	makeup.ExitDueToFatalError(err, fmt.Sprintf(format, args...))
}

var awsLogger = newStyledLogger()

var (
	klutchTagKey                      = "Klutch"
	klutchTagValue                    = "ControlPlane"
	klutchNamePrefix                  = "klutch-control-plane"
	klutchRoleLabel                   = "control plane"
	clusterNameTagKey                 = "eks.cluster/name"
	clusterIDTagKey                   = "eks.cluster/id"
	currentClusterName                string
	currentClusterArn                 string
	defaultTenantOperatorImage        = "public.ecr.aws/h6x7g6i7/anynines/cli-resources/tenants-operator:0.1.0"
	defaultTenantOperatorChartImage   = "oci://public.ecr.aws/h6x7g6i7/anynines/cli-resources/tenants-operator-chart"
	defaultTenantOperatorChartVersion = "0.1.0"
	// date of pinning: 16.04.26;
	// reason for pinning: serviceManaged field in DescribeAddresses API
	// response is required
	awsCliMinimumVersion = "v2.24.20"
)

func setKlutchContext(cfg Config) func() {
	prevTag := klutchTagValue
	prevPrefix := klutchNamePrefix
	prevRole := klutchRoleLabel

	if cfg.KlutchTagValue != "" {
		klutchTagValue = cfg.KlutchTagValue
	}
	if cfg.ResourceNamePrefix != "" {
		klutchNamePrefix = cfg.ResourceNamePrefix
	}
	if cfg.ClusterRole != "" {
		klutchRoleLabel = strings.ToLower(cfg.ClusterRole)
	}

	return func() {
		klutchTagValue = prevTag
		klutchNamePrefix = prevPrefix
		klutchRoleLabel = prevRole
	}
}

// RandomWorkloadClusterName generates a Klutch workload cluster name with a random hex suffix.
func RandomWorkloadClusterName() string {
	const suffixBytes = 4
	buf := make([]byte, suffixBytes)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to generate random workload cluster name: %w", err))
	}
	return fmt.Sprintf("klutch-workload-cluster-%s", hex.EncodeToString(buf))
}

func resourceName(cfg Config, parts ...string) string {
	return fmt.Sprintf("%s-%s", cfg.ResourceNamePrefix, strings.Join(parts, "-"))
}

func setClusterTagContext(name, arn string) {
	currentClusterName = name
	currentClusterArn = arn
}

func clusterTagPairsKV() []string {
	if currentClusterName == "" || currentClusterArn == "" {
		return nil
	}
	return []string{
		fmt.Sprintf("Key=%s,Value=%s", clusterNameTagKey, currentClusterName),
		fmt.Sprintf("Key=%s,Value=%s", clusterIDTagKey, currentClusterArn),
	}
}

func clusterTagPairsKMS() []string {
	if currentClusterName == "" || currentClusterArn == "" {
		return nil
	}
	return []string{
		fmt.Sprintf("TagKey=%s,TagValue=%s", clusterNameTagKey, currentClusterName),
		fmt.Sprintf("TagKey=%s,TagValue=%s", clusterIDTagKey, currentClusterArn),
	}
}

func tagEC2Resource(ctx context.Context, resourceID, name string) {
	tags := append([]string{
		fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Key=Name,Value=%s", name),
	}, clusterTagPairsKV()...)
	args := []string{
		"ec2", "create-tags",
		"--resources", resourceID,
		"--tags",
	}
	args = append(args, tags...)
	mustRun(ctx, "aws", args...)
}

func getenv(key, def string) string {
	return def
}

func defaultConfig() Config {
	return Config{
		Region:                     "eu-central-1",
		ClusterName:                "klutch-control-plane",
		NodegroupName:              "klutch-control-plane-nodegroup",
		NodeInstanceTypes:          "t3a.xlarge",
		NodeScalingConfig:          "minSize=3,maxSize=3,desiredSize=3",
		NodeAMIType:                "AL2023_x86_64_STANDARD",
		ClusterRoleName:            "EKSClusterRole",
		NodeRoleName:               "EKSNodeInstanceRole",
		BaseCIDR:                   "10.0.0.0/16",
		PubACIDR:                   "10.0.1.0/24",
		PubBCIDR:                   "10.0.2.0/24",
		PubCCIDR:                   "10.0.3.0/24",
		PrivACIDR:                  "10.0.101.0/24",
		PrivBCIDR:                  "10.0.102.0/24",
		PrivCCIDR:                  "10.0.103.0/24",
		ALBControllerVersion:       "v2.7.1",
		ALBControllerPolicyName:    "AWSLoadBalancerControllerIAMPolicy",
		ControlPlaneSGName:         "klutch-control-plane-sg",
		KlutchTagValue:             "ControlPlane",
		ResourceNamePrefix:         "klutch-control-plane",
		ClusterRole:                "Control Plane",
		TenantOperatorImage:        defaultTenantOperatorImage,
		TenantOperatorChart:        defaultTenantOperatorChartImage,
		TenantOperatorChartVersion: defaultTenantOperatorChartVersion,
		HostedZoneName:             "",
		AlbServiceAccountName:      "aws-load-balancer-controller",
	}
}

// ControlPlaneDefaultRegion returns the default region used for Klutch control plane resources.
func ControlPlaneDefaultRegion() string {
	return defaultConfig().Region
}

func workloadConfig(clusterName string) Config {
	cfg := defaultConfig()
	cfg.KlutchTagValue = "Workload"
	cfg.ClusterRole = "Workload"

	if clusterName == "" {
		clusterName = RandomWorkloadClusterName()
	}
	cfg.ClusterName = clusterName

	cfg.ClusterRoleName += "-" + cfg.ClusterName
	cfg.NodeRoleName += "-" + cfg.ClusterName
	cfg.ALBControllerPolicyName += "-" + cfg.ClusterName
	cfg.NodegroupName = fmt.Sprintf("%s-nodegroup", clusterName)
	cfg.ControlPlaneSGName = fmt.Sprintf("%s-sg", clusterName)
	cfg.ResourceNamePrefix = clusterName

	return cfg
}

// runCmd is assignable for tests; default implementation is defaultRunCmd.
var runCmd = defaultRunCmd

func defaultRunCmd(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stdout := strings.TrimSpace(outBuf.String())
	stderr := strings.TrimSpace(errBuf.String())

	if err != nil && isAWSCLI(name) && isAuthError(stderr) {
		awsLogger.Fatalf(err, "AWS authentication failed while running %s %s. Refresh credentials (e.g., aws sso login) and rerun.\nstderr: %s", name, strings.Join(args, " "), stderr)
	}

	return stdout, stderr, err
}

func isAWSCLI(name string) bool {
	base := filepath.Base(name)
	return base == "aws" || base == "aws.exe"
}

func isAuthError(errOut string) bool {
	lower := strings.ToLower(errOut)
	return strings.Contains(lower, "authfailure") ||
		strings.Contains(lower, "validate the provided access credentials") ||
		strings.Contains(lower, "expiredtoken") ||
		strings.Contains(lower, "token has expired") ||
		strings.Contains(lower, "invalidclienttokenid") ||
		strings.Contains(lower, "signaturedoesnotmatch") ||
		strings.Contains(lower, "request has expired") ||
		strings.Contains(lower, "security token included in the request is invalid") ||
		strings.Contains(lower, "could not find credentials") ||
		strings.Contains(lower, "no credential providers")
}

func mustRun(ctx context.Context, name string, args ...string) string {
	out, errOut, err := runCmd(ctx, name, args...)
	if err != nil {
		awsLogger.Fatalf(err, "❌ %s %v failed: %v\nstderr: %s", name, args, err, errOut)
	}
	return out
}

func CreateControlPlaneCluster(ctx context.Context, opts CreateOptions) {
	cfg := defaultConfig()
	if opts.ClusterName != "" {
		cfg.ClusterName = opts.ClusterName
	}
	cfg.NodegroupName = fmt.Sprintf("%s-nodegroup", cfg.ClusterName)
	cfg.ClusterRoleName += "-" + cfg.ClusterName
	cfg.NodeRoleName += "-" + cfg.ClusterName
	cfg.ALBControllerPolicyName += "-" + cfg.ClusterName
	cfg.ControlPlaneSGName = fmt.Sprintf("%s-sg", cfg.ClusterName)
	cfg.ResourceNamePrefix = cfg.ClusterName
	if opts.NodeInstanceTypes != "" {
		cfg.NodeInstanceTypes = opts.NodeInstanceTypes
	}
	if opts.NodeCount > 0 {
		cfg.NodeScalingConfig = fmt.Sprintf("minSize=%d,maxSize=%d,desiredSize=%d", opts.NodeCount, opts.NodeCount, opts.NodeCount)
	}
	if opts.TenantOperatorImage != "" {
		cfg.TenantOperatorImage = opts.TenantOperatorImage
	}
	if opts.TenantOperatorChart != "" {
		cfg.TenantOperatorChart = opts.TenantOperatorChart
	}
	if opts.TenantOperatorChartVersion != "" {
		cfg.TenantOperatorChartVersion = opts.TenantOperatorChartVersion
	}
	if opts.TenantOperatorRoleARN != "" {
		cfg.TenantOperatorRoleARN = opts.TenantOperatorRoleARN
	}
	if opts.TenantOperatorRegion != "" {
		cfg.TenantOperatorRegion = opts.TenantOperatorRegion
	}
	if opts.TenantOperatorBindURL != "" {
		cfg.TenantOperatorBindURL = opts.TenantOperatorBindURL
	}
	if opts.TenantOperatorBindRequest != "" {
		cfg.TenantOperatorBindRequest = opts.TenantOperatorBindRequest
	}
	if opts.HostedZoneName != "" {
		cfg.HostedZoneName = opts.HostedZoneName
	}
	provisionCluster(ctx, cfg, opts)
}

func CreateWorkloadCluster(ctx context.Context, opts CreateOptions) {
	cfg := workloadConfig(opts.ClusterName)
	if opts.NodeInstanceTypes != "" {
		cfg.NodeInstanceTypes = opts.NodeInstanceTypes
	}
	if opts.NodeCount > 0 {
		cfg.NodeScalingConfig = fmt.Sprintf("minSize=%d,maxSize=%d,desiredSize=%d", opts.NodeCount, opts.NodeCount, opts.NodeCount)
	}
	provisionCluster(ctx, cfg, opts)
}

func provisionCluster(ctx context.Context, cfg Config, opts CreateOptions) {
	restore := setKlutchContext(cfg)
	defer restore()

	awsLogger.Successf("Starting Klutch %s EKS cluster provisioning (Go version)", strings.ToLower(cfg.ClusterRole))

	requiredCmds := []string{"aws", "kubectl", "eksctl", "helm"}
	if opts.DryRun {
		awsLogger.Infof("Dry-run enabled: no changes will be made. Required tools for execution: %s.", strings.Join(requiredCmds, ", "))
	} else {
		for _, cmd := range requiredCmds {
			if _, err := execLookPath(cmd); err != nil {
				awsLogger.Fatalf(err, "Required command %q is not installed or not in PATH", cmd)
			}
		}
		checkAWSCLIVersion(ctx)
		awsLogger.Successf("All required commands (%s) are available.", strings.Join(requiredCmds, ", "))
	}

	// check cluster name length early to avoid cluster failing during provisioning due to this issue
	if fullAlbServiceAccountName := fmt.Sprintf("eksctl-%s-addon-iamserviceaccount-kube-system-%s", cfg.ClusterName, cfg.AlbServiceAccountName); len(fullAlbServiceAccountName) > 128 {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Cluster name %s is too long, cluster name may not exceed %d characters.", cfg.ClusterName, 128-len("eksctl--addon-iamserviceaccount-kube-system-"+cfg.AlbServiceAccountName)))
	}

	awsLogger.Section("Configuration")
	awsLogger.Printf("Klutch Role:                      %s", cfg.ClusterRole)
	awsLogger.Printf("Klutch Tag Value:                 %s", klutchTagValue)
	awsLogger.Printf("Resource Name Prefix:             %s", klutchNamePrefix)
	awsLogger.Printf("Region:                           %s", cfg.Region)
	awsLogger.Printf("Cluster Name:                     %s", cfg.ClusterName)
	if cfg.ClusterRole != "Control Plane" {
		autoBindControlPlane := opts.ControlPlaneToBindTo
		if autoBindControlPlane == "" {
			autoBindControlPlane = "<none given, using default value 'klutch-control-plane'>"
		}
		awsLogger.Printf("Control Plane Cluster to bind to: %s", autoBindControlPlane)
	}
	awsLogger.Printf("Nodegroup Name:                   %s", cfg.NodegroupName)
	awsLogger.Printf("Node Instance Types:              %s", cfg.NodeInstanceTypes)
	awsLogger.Printf("Nodegroup Scaling:                %s", cfg.NodeScalingConfig)
	awsLogger.Printf("Cluster Role Name:                %s", cfg.ClusterRoleName)
	awsLogger.Printf("Node Role Name:                   %s", cfg.NodeRoleName)
	awsLogger.Printf("VPC CIDR:                         %s", cfg.BaseCIDR)
	awsLogger.Printf("Public Subnets:                   %s, %s, %s", cfg.PubACIDR, cfg.PubBCIDR, cfg.PubCCIDR)
	awsLogger.Printf("Private Subnets:                  %s, %s, %s", cfg.PrivACIDR, cfg.PrivBCIDR, cfg.PrivCCIDR)
	awsLogger.Printf("ALB Controller Version:           %s", cfg.ALBControllerVersion)
	awsLogger.Printf("ALB Controller Policy Name:       %s", cfg.ALBControllerPolicyName)
	awsLogger.Printf("Cluster Security Group:           %s", cfg.ControlPlaneSGName)
	if cfg.TenantOperatorImage != "" {
		awsLogger.Printf("Tenant Operator Image:            %s", cfg.TenantOperatorImage)
	}
	if cfg.TenantOperatorChart != "" {
		awsLogger.Printf("Tenant Operator Chart:            %s", cfg.TenantOperatorChart)
	}

	if opts.DryRun {
		printCreatePlan(cfg)
		return
	}

	awsLogger.Infof("Detecting AWS Account ID...")
	accountID, errOut, err := runCmd(ctx, "aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	if err != nil || accountID == "" || accountID == "None" || accountID == "null" {
		awsLogger.Fatalf(err, "Unable to determine AWS Account ID. Run 'aws configure'. stderr: %s", errOut)
	}
	awsLogger.Infof("ACCOUNT_ID: %s", accountID)
	clusterArn := fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", cfg.Region, accountID, cfg.ClusterName)
	setClusterTagContext(cfg.ClusterName, clusterArn)

	awsLogger.Infof("Checking if cluster '%s' already exists...", cfg.ClusterName)
	clusterStatus := "NONE"
	if out, errOut, err := runCmd(ctx, "aws", "eks", "describe-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "cluster.status",
		"--output", "text"); err == nil {
		clusterStatus = out
	} else if !strings.Contains(errOut, "ResourceNotFoundException") {
		awsLogger.Fatalf(err, "aws eks describe-cluster failed\nstderr: %s", errOut)
	}
	clusterExists := false
	switch clusterStatus {
	case "NONE":
		awsLogger.Infof("Cluster does not exist. It will be created.")
	case "ACTIVE":
		awsLogger.Successf("Cluster already exists and is ACTIVE. Reusing it.")
		clusterExists = true
	case "CREATING":
		awsLogger.Warningf("Cluster exists and is in CREATING state. Waiting until ACTIVE...")
		mustRun(ctx, "aws", "eks", "wait", "cluster-active",
			"--name", cfg.ClusterName, "--region", cfg.Region)
		awsLogger.Successf("Cluster is now ACTIVE.")
		clusterExists = true
	default:
		awsLogger.Fatalf(nil, "Cluster exists but is in bad state: %s. Delete or fix manually.", clusterStatus)
	}

	awsLogger.Infof("Checking for existing Klutch %s VPC...", klutchRoleLabel)
	vpcID := ""
	if out, _, err := runCmd(ctx, "aws", "ec2", "describe-vpcs",
		"--filters", fmt.Sprintf("Name=tag:%s,Values=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Name=tag:Name,Values=%s", resourceName(cfg, "vpc")),
		"--query", "Vpcs[0].VpcId",
		"--output", "text"); err == nil && out != "None" && out != "null" {
		vpcID = out
		awsLogger.Successf("Reusing existing Klutch VPC: %s", vpcID)
	} else {
		awsLogger.Infof("No existing Klutch VPC found. A new VPC will be created.")
	}

	if vpcID == "" {
		awsLogger.Infof("Creating VPC...")
		vpcID = mustRun(ctx, "aws", "ec2", "create-vpc",
			"--cidr-block", cfg.BaseCIDR,
			"--query", "Vpc.VpcId", "--output", "text")
		tagEC2Resource(ctx, vpcID, resourceName(cfg, "vpc"))
		awsLogger.Successf("Created VPC: %s", vpcID)
	}

	ensureClusterRole(ctx, cfg.ClusterRoleName)
	ensureNodeRole(ctx, cfg.NodeRoleName)

	keyArn := ensureKMSKey(cfg, ctx, cfg.Region, accountID, cfg.ClusterRoleName)

	ensureNetworking(ctx, cfg, vpcID)

	if !clusterExists {
		createEKSCluster(ctx, cfg, vpcID, keyArn, accountID, clusterArn)
	} else {
		awsLogger.Infof("EKS cluster already exists. Skipping creation.")
	}
	mustRun(ctx, "aws", "eks", "describe-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "cluster.status")
	tagKMSKeyForCluster(ctx, keyArn, cfg.Region, accountID, cfg.ClusterName)

	ensureDefaultEBSEncryption(ctx)

	ensureNodegroup(ctx, cfg, vpcID, accountID)

	mustRun(ctx, "aws", "eks", "update-kubeconfig",
		"--region", cfg.Region,
		"--name", cfg.ClusterName)
	waitForNodesReady(ctx)

	ensureGp3StorageClass(ctx)

	ensureALBController(ctx, cfg, vpcID, accountID)

	if strings.EqualFold(cfg.ClusterRole, "Control Plane") {
		populateTenantOperatorDefaults(ctx, &cfg)
		if cfg.TenantOperatorRoleARN == "" {
			cfg.TenantOperatorRoleARN = ensureTenantOperatorRole(ctx, cfg, accountID)
		} else {
			awsLogger.Infof("Using provided tenant operator IAM role: %s", cfg.TenantOperatorRoleARN)
		}
		deployTenantOperator(ctx, cfg, accountID)
	} else {
		awsLogger.Infof("Skipping tenant operator IAM role and deployment for %s cluster.", cfg.ClusterRole)
	}

	awsLogger.Summaryf("Klutch %s EKS cluster is ready.", klutchRoleLabel)
	awsLogger.Printf("   Cluster:   %s", cfg.ClusterName)
	awsLogger.Printf("   Region:    %s", cfg.Region)
	awsLogger.Printf("   VPC:       %s", vpcID)
	awsLogger.Printf("   KMS Key:   %s", keyArn)
}

func checkAWSCLIVersion(ctx context.Context) {
	out, errOut, err := runCmd(ctx, "aws", "--version")
	if err != nil {
		awsLogger.Fatalf(err, "Could not check for aws CLI version:\n %s", errOut)
	}
	pattern := regexp.MustCompile(`aws-cli/([0-9]+\.[0-9]+\.[0-9]+)`)
	match := pattern.FindStringSubmatch(out)
	if len(match) < 2 {
		awsLogger.Fatalf(err, "Could not extract aws CLI version:\n %s", errOut)
	}
	version := "v" + match[1]
	if cmp := semver.Compare(version, awsCliMinimumVersion); cmp < 0 {
		awsLogger.Fatalf(nil, "AWS CLI version %s is not supported, minimum required version is %s", version, awsCliMinimumVersion)
	}
}

func printCreatePlan(cfg Config) {
	awsLogger.Section("Dry-Run Plan")
	roleLabel := strings.ToLower(cfg.ClusterRole)
	awsLogger.Infof("Listing the AWS resources that would be created or verified and the commands that would be executed for the Klutch %s cluster.", roleLabel)

	accountPlaceholder := "<account-id>"
	clusterArnPlaceholder := fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", cfg.Region, accountPlaceholder, cfg.ClusterName)
	vpcID := "<vpc-id>"
	igwID := "<internet-gateway-id>"
	sgID := "<control-plane-sg>"
	publicSubnets := []string{"<public-subnet-a>", "<public-subnet-b>", "<public-subnet-c>"}
	privateSubnets := []string{"<private-subnet-a>", "<private-subnet-b>", "<private-subnet-c>"}
	natGatewayIDs := []string{"<nat-gateway-a>", "<nat-gateway-b>", "<nat-gateway-c>"}
	clusterTags := fmt.Sprintf("%s=%s,Name=%s,%s=%s,%s=%s", klutchTagKey, klutchTagValue, cfg.ClusterName, clusterNameTagKey, cfg.ClusterName, clusterIDTagKey, clusterArnPlaceholder)
	nodeTags := fmt.Sprintf("%s=%s,Name=%s,%s=%s,%s=%s", klutchTagKey, klutchTagValue, cfg.NodegroupName, clusterNameTagKey, cfg.ClusterName, clusterIDTagKey, clusterArnPlaceholder)

	type planItem struct {
		Title    string
		Purpose  string
		Commands []string
	}

	plan := []planItem{
		{
			Title:   "AWS identity and cluster lookup",
			Purpose: fmt.Sprintf("Identify the AWS account and reuse the %s cluster when it already exists.", roleLabel),
			Commands: []string{
				"aws sts get-caller-identity --query Account --output text",
				fmt.Sprintf("aws eks describe-cluster --name %s --region %s --query cluster.status --output text", cfg.ClusterName, cfg.Region),
			},
		},
		{
			Title:   "IAM roles for cluster and nodes",
			Purpose: "Grant the EKS service and worker nodes the permissions required to manage AWS resources and pull container images.",
			Commands: []string{
				fmt.Sprintf("aws iam get-role --role-name %s", cfg.ClusterRoleName),
				fmt.Sprintf("aws iam create-role --role-name %s --assume-role-policy-document file:///tmp/eks-cluster-trust.json --tags Key=%s,Value=%s Key=Name,Value=%s", cfg.ClusterRoleName, klutchTagKey, klutchTagValue, cfg.ClusterRoleName),
				fmt.Sprintf("aws iam attach-role-policy --role-name %s --policy-arn arn:aws:iam::aws:policy/AmazonEKSClusterPolicy", cfg.ClusterRoleName),
				fmt.Sprintf("aws iam get-role --role-name %s", cfg.NodeRoleName),
				fmt.Sprintf("aws iam create-role --role-name %s --assume-role-policy-document file:///tmp/eks-node-trust.json --tags Key=%s,Value=%s Key=Name,Value=%s", cfg.NodeRoleName, klutchTagKey, klutchTagValue, cfg.NodeRoleName),
				fmt.Sprintf("aws iam attach-role-policy --role-name %s --policy-arn arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy", cfg.NodeRoleName),
				fmt.Sprintf("aws iam attach-role-policy --role-name %s --policy-arn arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy", cfg.NodeRoleName),
				fmt.Sprintf("aws iam attach-role-policy --role-name %s --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly", cfg.NodeRoleName),
			},
		},
		{
			Title:   "KMS key for secret encryption",
			Purpose: "Create and tag a KMS key so EKS can encrypt Kubernetes secrets at rest.",
			Commands: []string{
				fmt.Sprintf("aws kms create-key --description \"Encrypts secret data stored by the Klutch %s EKS cluster\" --query KeyMetadata.KeyId --output text --tags TagKey=Klutch,TagValue=%s TagKey=Name,TagValue=%s", roleLabel, klutchTagValue, resourceName(cfg, "kms-key")),
				fmt.Sprintf("aws iam put-role-policy --role-name %s --policy-document file:///tmp/eks-kms-policy.json", cfg.ClusterRoleName),
				fmt.Sprintf("aws kms tag-resource --key-id <kms-arn> --tags TagKey=%s,TagValue=%s TagKey=%s,TagValue=%s", clusterNameTagKey, cfg.ClusterName, clusterIDTagKey, clusterArnPlaceholder),
			},
		},
		{
			Title:   "Networking (VPC, subnets, routing, NAT, security group)",
			Purpose: "Provision a dedicated VPC with DNS support, public/private subnets, internet and NAT gateways, routing tables, and the cluster security group.",
			Commands: []string{
				fmt.Sprintf("aws ec2 describe-vpcs --filters Name=tag:%s,Values=%s --query Vpcs[0].VpcId --output text", klutchTagKey, klutchTagValue),
				fmt.Sprintf("aws ec2 create-vpc --cidr-block %s --query Vpc.VpcId --output text", cfg.BaseCIDR),
				fmt.Sprintf("aws ec2 modify-vpc-attribute --vpc-id %s --enable-dns-support '{\"Value\":true}'", vpcID),
				fmt.Sprintf("aws ec2 modify-vpc-attribute --vpc-id %s --enable-dns-hostnames '{\"Value\":true}'", vpcID),
				fmt.Sprintf("aws ec2 create-internet-gateway --query InternetGateway.InternetGatewayId --output text; aws ec2 attach-internet-gateway --internet-gateway-id %s --vpc-id %s", igwID, vpcID),
				fmt.Sprintf("aws ec2 create-subnet --vpc-id %s --cidr-block %s --availability-zone %sa (repeat for %s / %s public CIDRs)", vpcID, cfg.PubACIDR, cfg.Region, cfg.PubBCIDR, cfg.PubCCIDR),
				fmt.Sprintf("aws ec2 create-subnet --vpc-id %s --cidr-block %s --availability-zone %sa (repeat for %s / %s private CIDRs)", vpcID, cfg.PrivACIDR, cfg.Region, cfg.PrivBCIDR, cfg.PrivCCIDR),
				fmt.Sprintf("aws ec2 create-route-table --vpc-id %s; aws ec2 create-route --route-table-id <public-rt> --destination-cidr-block 0.0.0.0/0 --gateway-id %s; aws ec2 associate-route-table --route-table-id <public-rt> --subnet-id %s (for all public subnets)", vpcID, igwID, strings.Join(publicSubnets, "/")),
				"aws ec2 describe-addresses --query Addresses[].AllocationId --output text; aws service-quotas get-service-quota --service-code ec2 --quota-code L-0263D0A3 --query Quota.Value --output text",
				fmt.Sprintf("aws ec2 allocate-address --domain vpc; aws ec2 create-nat-gateway --subnet-id %s --allocation-id <alloc-id> --query NatGateway.NatGatewayId --output text (for each public subnet); aws ec2 wait nat-gateway-available --nat-gateway-ids %s", publicSubnets[0], strings.Join(natGatewayIDs, " ")),
				fmt.Sprintf("aws ec2 create-route-table --vpc-id %s; aws ec2 create-route --route-table-id <private-rt> --destination-cidr-block 0.0.0.0/0 --nat-gateway-id %s; aws ec2 associate-route-table --route-table-id <private-rt> --subnet-id %s (for each private subnet)", vpcID, natGatewayIDs[0], strings.Join(privateSubnets, "/")),
				fmt.Sprintf("aws ec2 create-security-group --group-name %s --description \"Restricts traffic for Klutch %s components and worker nodes\" --vpc-id %s", cfg.ControlPlaneSGName, roleLabel, vpcID),
				fmt.Sprintf("aws ec2 authorize-security-group-ingress --group-id %s --protocol -1 --source-group %s; aws ec2 authorize-security-group-egress --group-id %s --protocol -1 --cidr 0.0.0.0/0", sgID, sgID, sgID),
			},
		},
		{
			Title:   "EKS cluster",
			Purpose: "Create the EKS cluster with secret encryption, private networking, and Klutch tags.",
			Commands: []string{
				fmt.Sprintf("aws eks create-cluster --name %s --region %s --role-arn arn:aws:iam::%s:role/%s --resources-vpc-config subnetIds=%s,securityGroupIds=%s --encryption-config [{\"resources\":[\"secrets\"],\"provider\":{\"keyArn\":\"<kms-arn>\"}}] --tags %s", cfg.ClusterName, cfg.Region, accountPlaceholder, cfg.ClusterRoleName, strings.Join(privateSubnets, ","), sgID, clusterTags),
				fmt.Sprintf("aws eks wait cluster-active --name %s --region %s", cfg.ClusterName, cfg.Region),
				fmt.Sprintf("aws eks describe-cluster --name %s --region %s --query cluster.status", cfg.ClusterName, cfg.Region),
			},
		},
		{
			Title:   "Managed nodegroup",
			Purpose: "Provision worker nodes for the cluster workloads with the requested capacity and instance type.",
			Commands: []string{
				fmt.Sprintf("aws eks describe-nodegroup --cluster-name %s --nodegroup-name %s --region %s --query nodegroup.status --output text", cfg.ClusterName, cfg.NodegroupName, cfg.Region),
				fmt.Sprintf("aws eks create-nodegroup --cluster-name %s --nodegroup-name %s --scaling-config %s --instance-types %s --subnets %s --ami-type %s --node-role arn:aws:iam::%s:role/%s --region %s --tags %s", cfg.ClusterName, cfg.NodegroupName, cfg.NodeScalingConfig, cfg.NodeInstanceTypes, strings.Join(privateSubnets, " "), cfg.NodeAMIType, accountPlaceholder, cfg.NodeRoleName, cfg.Region, nodeTags),
				fmt.Sprintf("aws eks wait nodegroup-active --cluster-name %s --nodegroup-name %s --region %s", cfg.ClusterName, cfg.NodegroupName, cfg.Region),
			},
		},
		{
			Title:   "Kubeconfig and node readiness",
			Purpose: "Point kubectl to the cluster and wait until nodes report Ready.",
			Commands: []string{
				fmt.Sprintf("aws eks update-kubeconfig --region %s --name %s", cfg.Region, cfg.ClusterName),
				"kubectl get nodes (polled until at least one node is Ready)",
			},
		},
		{
			Title:   "Account default EBS encryption",
			Purpose: "Ensure new EBS volumes are encrypted by default for this AWS account.",
			Commands: []string{
				"aws ec2 get-ebs-encryption-by-default",
				"aws ec2 enable-ebs-encryption-by-default",
			},
		},
		{
			Title:   "Default gp3 StorageClass",
			Purpose: "Install and mark a gp3-backed Kubernetes StorageClass as the cluster default.",
			Commands: []string{
				"kubectl apply -f <gp3-storageclass-manifest>",
			},
		},
		{
			Title:   "AWS Load Balancer Controller",
			Purpose: "Install the ALB controller used for Klutch ingress and service routing.",
			Commands: []string{
				fmt.Sprintf("eksctl utils associate-iam-oidc-provider --region %s --cluster %s --approve", cfg.Region, cfg.ClusterName),
				"curl -sSfL -o aws-load-balancer-controller-policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/main/docs/install/iam_policy.json",
				fmt.Sprintf("aws iam create-policy --policy-name %s --policy-document file://aws-load-balancer-controller-policy.json --description \"Allows the Klutch %s to run the AWS Load Balancer Controller safely\" --tags Key=%s,Value=%s Key=Name,Value=%s", cfg.ALBControllerPolicyName, roleLabel, klutchTagKey, klutchTagValue, cfg.ALBControllerPolicyName),
				fmt.Sprintf("aws iam create-policy-version --policy-arn arn:aws:iam::%s:policy/%s --policy-document file://aws-load-balancer-controller-policy.json --set-as-default", accountPlaceholder, cfg.ALBControllerPolicyName),
				fmt.Sprintf("eksctl create iamserviceaccount --cluster %s --namespace kube-system --name aws-load-balancer-controller --attach-policy-arn arn:aws:iam::%s:policy/%s --region %s --approve --override-existing-serviceaccounts", cfg.ClusterName, accountPlaceholder, cfg.ALBControllerPolicyName, cfg.Region),
				"helm repo add eks https://aws.github.io/eks-charts",
				"helm repo update",
				fmt.Sprintf("helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller -n kube-system --set clusterName=%s --set region=%s --set vpcId=%s --set serviceAccount.create=false --set serviceAccount.name=aws-load-balancer-controller", cfg.ClusterName, cfg.Region, vpcID),
				fmt.Sprintf("aws iam attach-role-policy --role-name <role-derived-from-sa> --policy-arn arn:aws:iam::%s:policy/%s", accountPlaceholder, cfg.ALBControllerPolicyName),
				"kubectl rollout status deployment/aws-load-balancer-controller -n kube-system",
			},
		},
	}

	for _, item := range plan {
		awsLogger.Printf("- %s: %s", item.Title, item.Purpose)
		for _, cmd := range item.Commands {
			awsLogger.Printf("    %s", cmd)
		}
	}

	awsLogger.Infof("Dry-run complete. Run without --dry-run to execute the commands above.")
}

func ensureClusterRole(ctx context.Context, roleName string) {
	awsLogger.Infof("Ensuring IAM role '%s' exists...", roleName)
	if _, _, err := runCmd(ctx, "aws", "iam", "get-role", "--role-name", roleName); err == nil {
		awsLogger.Successf("EKS cluster role '%s' already exists.", roleName)
		return
	}
	trust := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`
	tmp := "/tmp/eks-cluster-trust.json"
	if err := os.WriteFile(tmp, []byte(trust), 0600); err != nil {
		awsLogger.Fatalf(err, "writing trust policy failed")
	}
	args := []string{
		"iam", "create-role",
		"--role-name", roleName,
		"--assume-role-policy-document", "file://" + tmp,
		"--description", fmt.Sprintf("Allows the Klutch %s EKS cluster to manage AWS resources on its behalf", klutchRoleLabel),
		"--tags",
	}
	args = append(args, append([]string{
		fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Key=Name,Value=%s", roleName),
	}, clusterTagPairsKV()...)...)
	mustRun(ctx, "aws", args...)
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy")
	awsLogger.Successf("Created EKS cluster role '%s'.", roleName)
}

func ensureNodeRole(ctx context.Context, roleName string) {
	awsLogger.Infof("Ensuring IAM role '%s' exists...", roleName)
	if _, _, err := runCmd(ctx, "aws", "iam", "get-role", "--role-name", roleName); err == nil {
		awsLogger.Successf("EKS node role '%s' already exists.", roleName)
		return
	}
	trust := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`
	tmp := "/tmp/eks-node-trust.json"
	if err := os.WriteFile(tmp, []byte(trust), 0600); err != nil {
		awsLogger.Fatalf(err, "writing node trust policy failed")
	}
	args := []string{
		"iam", "create-role",
		"--role-name", roleName,
		"--assume-role-policy-document", "file://" + tmp,
		"--description", fmt.Sprintf("Provides Klutch %s worker nodes the permissions required to integrate with AWS", klutchRoleLabel),
		"--tags",
	}
	args = append(args, append([]string{
		fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Key=Name,Value=%s", roleName),
	}, clusterTagPairsKV()...)...)
	mustRun(ctx, "aws", args...)
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy")
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy")
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly")
	awsLogger.Successf("Created EKS node role '%s'.", roleName)
}

func ensureTenantOperatorRole(ctx context.Context, cfg Config, accountID string) string {
	roleName := resourceName(cfg, "tenant-operator")
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName)

	awsLogger.Section("Tenant Operator IAM role")
	awsLogger.Infof("Ensuring IAM role for tenant operator exists (role=%s)...", roleName)

	// Ensure OIDC provider associated for IRSA.
	mustRun(ctx, "eksctl", "utils", "associate-iam-oidc-provider",
		"--region", cfg.Region,
		"--cluster", cfg.ClusterName,
		"--approve")

	issuer, errOut, err := runCmd(ctx, "aws", "eks", "describe-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "cluster.identity.oidc.issuer",
		"--output", "text")
	if err != nil || strings.TrimSpace(issuer) == "" {
		awsLogger.Fatalf(err, "Failed to discover OIDC issuer for cluster %s\nstderr: %s", cfg.ClusterName, errOut)
	}
	issuer = strings.TrimSpace(issuer)
	providerHost := strings.TrimPrefix(issuer, "https://")
	providerArn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountID, providerHost)

	trust := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "%s"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "%s:sub": "system:serviceaccount:a9s-tenants-operator-system:a9s-tenants-operator",
          "%s:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}`, providerArn, providerHost, providerHost)
	trustFile := "/tmp/tenant-operator-trust.json"
	if err := os.WriteFile(trustFile, []byte(trust), 0600); err != nil {
		awsLogger.Fatalf(err, "writing tenant operator trust policy failed")
	}

	roleExists := false
	if _, _, err := runCmd(ctx, "aws", "iam", "get-role", "--role-name", roleName); err == nil {
		roleExists = true
		// Update trust to match current cluster issuer/subject.
		if _, errOut, err := runCmd(ctx, "aws", "iam", "update-assume-role-policy",
			"--role-name", roleName,
			"--policy-document", "file://"+trustFile); err != nil {
			awsLogger.Warningf("Failed to update tenant operator role trust policy: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Updated tenant operator IAM role trust policy: %s", roleArn)
		}
	} else {
		args := []string{
			"iam", "create-role",
			"--role-name", roleName,
			"--assume-role-policy-document", "file://" + trustFile,
			"--description", fmt.Sprintf("Allows the Klutch %s tenant operator to manage Cognito and Secrets Manager for tenants", klutchRoleLabel),
			"--tags",
		}
		args = append(args, append([]string{
			fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
			fmt.Sprintf("Key=Name,Value=%s", roleName),
		}, clusterTagPairsKV()...)...)
		mustRun(ctx, "aws", args...)
	}

	secretArn := fmt.Sprintf("arn:aws:secretsmanager:%s:%s:secret:klutch/*", cfg.Region, accountID)
	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cognito-idp:CreateUserPool",
        "cognito-idp:ListUserPools",
        "cognito-idp:CreateUserPoolClient",
        "cognito-idp:DescribeUserPool",
        "cognito-idp:DescribeUserPoolClient",
        "cognito-idp:ListUserPoolClients",
        "cognito-idp:UpdateUserPoolClient",
        "cognito-idp:DeleteUserPoolClient",
        "cognito-idp:CreateUserPoolDomain",
        "cognito-idp:DescribeUserPoolDomain",
        "cognito-idp:DeleteUserPoolDomain",
        "cognito-idp:CreateResourceServer",
        "cognito-idp:DescribeResourceServer",
        "cognito-idp:UpdateResourceServer",
        "cognito-idp:DeleteResourceServer",
        "cognito-idp:TagResource",
        "secretsmanager:CreateSecret",
        "secretsmanager:PutSecretValue",
        "secretsmanager:UpdateSecret",
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret",
        "secretsmanager:TagResource"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:CreateSecret",
        "secretsmanager:PutSecretValue",
        "secretsmanager:UpdateSecret",
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret",
        "secretsmanager:TagResource"
      ],
      "Resource": "%s"
    }
  ]
}`, secretArn)
	policyFile := "/tmp/tenant-operator-policy.json"
	if err := os.WriteFile(policyFile, []byte(policy), 0600); err != nil {
		awsLogger.Fatalf(err, "writing tenant operator policy failed")
	}

	mustRun(ctx, "aws", "iam", "put-role-policy",
		"--role-name", roleName,
		"--policy-name", "TenantOperatorInline",
		"--policy-document", "file://"+policyFile)

	if roleExists {
		awsLogger.Successf("Updated tenant operator IAM role: %s", roleArn)
	} else {
		awsLogger.Successf("Created tenant operator IAM role: %s", roleArn)
	}
	return roleArn
}

func resolveKubectlContextForCluster(ctx context.Context, cfg Config) string {
	out, err := k8s.Contexts(cfg.ClusterName)
	if err != nil || len(out) == 0 {
		makeup.PrintWarning("Could retrieve contexts, falling back to guess " + cfg.ClusterName)
		return cfg.ClusterName
	}

	context := ""
	id, err := getAccountID(ctx)
	if err != nil || len(id) == 0 {
		makeup.PrintWarning("Could not retrieve AWS Account ID for Kubecontext searching, falling back on RegEx")
		id = `[0-9]+`
	}
	eksctlIoPattern := regexp.MustCompile(fmt.Sprintf(`[^@]+@[^@]+@%s\.%s\.eksctl\.io`, cfg.ClusterName, cfg.Region))
	arnPattern := regexp.MustCompile(fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", cfg.Region, id, cfg.ClusterName))
	for _, line := range out {
		if line == cfg.ClusterName ||
			eksctlIoPattern.Match([]byte(line)) ||
			arnPattern.Match([]byte(line)) {
			context = line
		}
	}

	if context != "" {
		makeup.PrintSuccess("Found context " + context)
		return context
	}

	makeup.PrintWarning("Could not find exact match for " + cfg.ClusterName + " via considered string comparisons, falling back to guess " + out[0])
	return out[0]
}

// ApplyControlPlaneAddons installs AWS-side addons (tenant operator) onto an existing control-plane cluster.
// It assumes the EKS cluster already exists and that ALB/etc. are in place.
func ApplyControlPlaneAddons(ctx context.Context, opts CreateOptions) {
	cfg := defaultConfig()
	if strings.TrimSpace(opts.ClusterName) != "" {
		cfg.ClusterName = strings.TrimSpace(opts.ClusterName)
	}
	if opts.TenantOperatorImage != "" {
		cfg.TenantOperatorImage = opts.TenantOperatorImage
	}
	if opts.TenantOperatorChart != "" {
		cfg.TenantOperatorChart = opts.TenantOperatorChart
	}
	if opts.TenantOperatorChartVersion != "" {
		cfg.TenantOperatorChartVersion = opts.TenantOperatorChartVersion
	}
	if opts.TenantOperatorRoleARN != "" {
		cfg.TenantOperatorRoleARN = opts.TenantOperatorRoleARN
	}
	if opts.TenantOperatorRegion != "" {
		cfg.TenantOperatorRegion = opts.TenantOperatorRegion
	}
	if opts.TenantOperatorBindURL != "" {
		cfg.TenantOperatorBindURL = opts.TenantOperatorBindURL
	}
	if opts.TenantOperatorBindRequest != "" {
		cfg.TenantOperatorBindRequest = opts.TenantOperatorBindRequest
	}
	if opts.HostedZoneName != "" {
		cfg.HostedZoneName = opts.HostedZoneName
	}

	awsLogger.Successf("Applying Klutch control-plane addons to existing cluster %s", cfg.ClusterName)

	accountID, errOut, err := runCmd(ctx, "aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	if err != nil || strings.TrimSpace(accountID) == "" {
		awsLogger.Fatalf(err, "Unable to determine AWS Account ID. stderr: %s", errOut)
	}
	clusterArn := fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", cfg.Region, accountID, cfg.ClusterName)
	setClusterTagContext(cfg.ClusterName, clusterArn)

	// Ensure kubeconfig points to the cluster.
	mustRun(ctx, "aws", "eks", "update-kubeconfig", "--region", cfg.Region, "--name", cfg.ClusterName)
	// Switch kubectl context to the control-plane cluster so subsequent apply steps hit the right cluster.
	cpCtx := resolveKubectlContextForCluster(ctx, cfg)
	if strings.TrimSpace(cpCtx) != "" {
		if out, err := k8s.SwitchContext(cpCtx); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to switch kubectl context to %s (continuing):\n: %s", cpCtx, out))
		} else {
			awsLogger.Infof("Using kubectl context %s for control-plane apply.", cpCtx)
		}
	}

	// Ensure OIDC provider associated for IRSA.
	mustRun(ctx, "eksctl", "utils", "associate-iam-oidc-provider",
		"--region", cfg.Region,
		"--cluster", cfg.ClusterName,
		"--approve")

	// Ensure IAM role for tenant operator and deploy it.
	cfg.TenantOperatorRoleARN = ensureTenantOperatorRole(ctx, cfg, accountID)

	// Derive defaults for bind URL/request if missing.
	populateTenantOperatorDefaults(ctx, &cfg)

	deployTenantOperator(ctx, cfg, accountID)

	awsLogger.Successf("Klutch control-plane addons applied to cluster %s.", cfg.ClusterName)
}

func ensureKMSKey(cfg Config, ctx context.Context, region, accountID, clusterRole string) string {
	keyID := os.Getenv("KEY_ID")
	if keyID == "" {
		awsLogger.Infof("Creating new KMS key for EKS secret encryption...")
		tags := append([]string{
			fmt.Sprintf("TagKey=%s,TagValue=%s", klutchTagKey, klutchTagValue),
			fmt.Sprintf("TagKey=Name,TagValue=%s", resourceName(cfg, "kms-key")),
		}, clusterTagPairsKMS()...)
		args := []string{
			"kms", "create-key",
			"--description", fmt.Sprintf("Encrypts secret data stored by the Klutch %s EKS cluster", klutchRoleLabel),
			"--query", "KeyMetadata.KeyId",
			"--output", "text",
			"--tags",
		}
		args = append(args, tags...)
		keyID = mustRun(ctx, "aws", args...)
	} else {
		awsLogger.Infof("Using existing KMS KEY_ID: %s", keyID)
	}
	keyArn := fmt.Sprintf("arn:aws:kms:%s:%s:key/%s", region, accountID, keyID)
	awsLogger.Infof("KMS KEY_ID: %s", keyID)
	awsLogger.Infof("KMS KEY_ARN: %s", keyArn)

	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowEKSToUseKMSKeyForSecretsEncryption",
      "Effect": "Allow",
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:GenerateDataKey*",
        "kms:DescribeKey"
      ],
      "Resource": "%s"
    }
  ]
}`, keyArn)
	tmp := "/tmp/eks-kms-policy.json"
	if err := os.WriteFile(tmp, []byte(policy), 0600); err != nil {
		awsLogger.Fatalf(err, "writing kms role policy failed")
	}
	mustRun(ctx, "aws", "iam", "put-role-policy",
		"--role-name", clusterRole,
		"--policy-name", "EKSClusterKMSAccess",
		"--policy-document", "file://"+tmp)
	return keyArn
}

func tagKMSKeyForCluster(ctx context.Context, keyArn, region, accountID, clusterName string) {
	clusterArn := fmt.Sprintf("arn:aws:eks:%s:%s:cluster/%s", region, accountID, clusterName)
	if _, errOut, err := runCmd(ctx, "aws", "kms", "tag-resource",
		"--key-id", keyArn,
		"--tags",
		fmt.Sprintf("TagKey=%s,TagValue=%s", "eks.cluster/name", clusterName),
		fmt.Sprintf("TagKey=%s,TagValue=%s", "eks.cluster/id", clusterArn)); err != nil {
		awsLogger.Warningf("Failed to tag KMS key %s with cluster info: %v\nstderr: %s", keyArn, err, errOut)
	} else {
		awsLogger.Successf("Tagged KMS key with cluster info (name=%s, id=%s).", clusterName, clusterArn)
	}
}

func ensureNetworking(ctx context.Context, cfg Config, vpcID string) {
	awsLogger.Section("Networking")
	awsLogger.Infof("Ensuring networking components (VPC attributes, IGW, subnets, routes, NAT, SG)...")

	mustRun(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-support", "{\"Value\":true}")
	mustRun(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-hostnames", "{\"Value\":true}")

	igwID := ensureIGW(cfg, ctx, vpcID)

	pubA := ensureSubnet(ctx, vpcID, cfg.PubACIDR, cfg.Region+"a", resourceName(cfg, "public-subnet-a"))
	pubB := ensureSubnet(ctx, vpcID, cfg.PubBCIDR, cfg.Region+"b", resourceName(cfg, "public-subnet-b"))
	pubC := ensureSubnet(ctx, vpcID, cfg.PubCCIDR, cfg.Region+"c", resourceName(cfg, "public-subnet-c"))
	privA := ensureSubnet(ctx, vpcID, cfg.PrivACIDR, cfg.Region+"a", resourceName(cfg, "private-subnet-a"))
	privB := ensureSubnet(ctx, vpcID, cfg.PrivBCIDR, cfg.Region+"b", resourceName(cfg, "private-subnet-b"))
	privC := ensureSubnet(ctx, vpcID, cfg.PrivCCIDR, cfg.Region+"c", resourceName(cfg, "private-subnet-c"))

	awsLogger.Printf("PUBLIC SUBNETS:  %s, %s, %s", pubA, pubB, pubC)
	awsLogger.Printf("PRIVATE SUBNETS: %s, %s, %s", privA, privB, privC)

	publicRT := ensurePublicRouteTable(cfg, ctx, vpcID, igwID, []string{pubA, pubB, pubC})

	natA, natB, natC := ensureNATs(cfg, ctx, vpcID, pubA, pubB, pubC)

	privRTA := ensurePrivateRT(ctx, vpcID, privA, natA, resourceName(cfg, "private-route-table-a"))
	privRTB := ensurePrivateRT(ctx, vpcID, privB, natB, resourceName(cfg, "private-route-table-b"))
	privRTC := ensurePrivateRT(ctx, vpcID, privC, natC, resourceName(cfg, "private-route-table-c"))
	_ = publicRT
	awsLogger.Printf("PRIVATE ROUTE TABLES: %s, %s, %s", privRTA, privRTB, privRTC)

	ensureSecurityGroup(cfg, ctx, vpcID, cfg.ControlPlaneSGName)
}

// deriveBindURLFromCluster tries to read the control-plane info ConfigMap to build a bind URL.
func deriveBindURLFromCluster(ctx context.Context) string {
	k8sClient := k8s.NewKubeClient("")
	hostByte, _ := k8sClient.Get("configmap", "klutch-control-plane-info", "-A", "jsonpath={.data.host}", false)
	portByte, _ := k8sClient.Get("configmap", "klutch-control-plane-info", "-A", "jsonpath={.data.ingressPort}", false)
	host := strings.TrimSpace(string(hostByte))
	port := strings.TrimSpace(string(portByte))
	if host == "" {
		return ""
	}
	scheme := "https"
	if port == "80" {
		scheme = "http"
	}
	switch port {
	case "", "443", "80":
		return fmt.Sprintf("%s://%s/bind-noninteractive", scheme, host)
	default:
		return fmt.Sprintf("%s://%s:%s/bind-noninteractive", scheme, host, port)
	}
}

// defaultBindRequestJSON returns the default bind request JSON for a tenant clusterID.
func defaultBindRequestJSON(clusterID string) string {
	req := struct {
		ClusterID string `json:"clusterID"`
		Apis      []struct {
			Group    string `json:"group"`
			Resource string `json:"resource"`
		} `json:"apis"`
	}{
		ClusterID: clusterID,
		Apis: []struct {
			Group    string `json:"group"`
			Resource string `json:"resource"`
		}{
			{Group: "anynines.com", Resource: "postgresqlinstances"},
			{Group: "anynines.com", Resource: "servicebindings"},
			{Group: "anynines.com", Resource: "backups"},
			{Group: "anynines.com", Resource: "restores"},
		},
	}
	b, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	return string(b)
}

// populateTenantOperatorDefaults fills bind URL/request and region if missing.
func populateTenantOperatorDefaults(ctx context.Context, cfg *Config) {
	if strings.TrimSpace(cfg.TenantOperatorBindURL) == "" {
		if url := deriveBindURLFromCluster(ctx); strings.TrimSpace(url) != "" {
			cfg.TenantOperatorBindURL = url
		} else if hz := strings.Trim(strings.TrimSpace(cfg.HostedZoneName), "."); hz != "" {
			host := fmt.Sprintf("klutch-bind.%s", hz)
			cfg.TenantOperatorBindURL = fmt.Sprintf("https://%s/bind-noninteractive", host)
		}
	}
	if strings.TrimSpace(cfg.TenantOperatorBindRequest) == "" {
		cfg.TenantOperatorBindRequest = defaultBindRequestJSON(cfg.ClusterName)
	}
	if strings.TrimSpace(cfg.TenantOperatorRegion) == "" {
		cfg.TenantOperatorRegion = cfg.Region
	}
}

func ensureIGW(cfg Config, ctx context.Context, vpcID string) string {
	awsLogger.Infof("Checking for existing Internet Gateway attached to VPC %s...", vpcID)
	igwID, _, _ := runCmd(ctx, "aws", "ec2", "describe-internet-gateways",
		"--filters", "Name=attachment.vpc-id,Values="+vpcID,
		"--query", "InternetGateways[0].InternetGatewayId",
		"--output", "text")
	if igwID == "" || igwID == "None" || igwID == "null" {
		awsLogger.Infof("No Internet Gateway found. Creating and attaching a new one...")
		igwID = mustRun(ctx, "aws", "ec2", "create-internet-gateway",
			"--query", "InternetGateway.InternetGatewayId",
			"--output", "text")
		tagEC2Resource(ctx, igwID, resourceName(cfg, "internet-gateway"))
		mustRun(ctx, "aws", "ec2", "attach-internet-gateway",
			"--internet-gateway-id", igwID,
			"--vpc-id", vpcID)
	} else {
		awsLogger.Successf("Reusing existing Internet Gateway: %s", igwID)
	}
	awsLogger.Infof("IGW_ID = %s", igwID)
	return igwID
}

func ensureSubnet(ctx context.Context, vpcID, cidr, az, name string) string {
	out, _, _ := runCmd(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+cidr,
		"--query", "Subnets[0].SubnetId",
		"--output", "text")
	if out == "" || out == "None" || out == "null" {
		awsLogger.Infof("Creating subnet %s in AZ %s...", cidr, az)
		out = mustRun(ctx, "aws", "ec2", "create-subnet",
			"--vpc-id", vpcID,
			"--cidr-block", cidr,
			"--availability-zone", az,
			"--query", "Subnet.SubnetId",
			"--output", "text")
		tagEC2Resource(ctx, out, name)
	} else {
		awsLogger.Successf("Reusing subnet %s: %s", cidr, out)
	}
	return out
}

func ensurePublicRouteTable(cfg Config, ctx context.Context, vpcID, igwID string, pubSubnets []string) string {
	awsLogger.Infof("Checking for existing public route table...")
	rtID, _, _ := runCmd(ctx, "aws", "ec2", "describe-route-tables",
		"--filters", "Name=vpc-id,Values="+vpcID,
		"--query", "RouteTables[?Routes[?GatewayId=='"+igwID+"' && DestinationCidrBlock=='0.0.0.0/0']].RouteTableId | [0]",
		"--output", "text")
	if rtID == "" || rtID == "None" || rtID == "null" {
		awsLogger.Infof("Creating public route table...")
		rtID = mustRun(ctx, "aws", "ec2", "create-route-table",
			"--vpc-id", vpcID,
			"--query", "RouteTable.RouteTableId",
			"--output", "text")
		tagEC2Resource(ctx, rtID, resourceName(cfg, "public-route-table"))
		mustRun(ctx, "aws", "ec2", "create-route",
			"--route-table-id", rtID,
			"--destination-cidr-block", "0.0.0.0/0",
			"--gateway-id", igwID)
	} else {
		awsLogger.Successf("Reusing public route table: %s", rtID)
	}
	for _, sn := range pubSubnets {
		ensureRouteTableAssociation(ctx, rtID, sn)
	}
	return rtID
}

func ensureElasticIPQuota(ctx context.Context, required int) {
	awsLogger.Infof("Checking Elastic IP quota (need %d new EIPs)...", required)
	out, errOut, err := runCmd(ctx, "aws", "ec2", "describe-addresses",
		"--query", "Addresses[?!ServiceManaged].AllocationId",
		"--output", "text")
	currentCount := 0
	if err == nil && strings.TrimSpace(out) != "" {
		parts := strings.Fields(out)
		currentCount = len(parts)
	} else if err != nil && !strings.Contains(errOut, "AuthFailure") {
		awsLogger.Warningf("describe-addresses failed: %v, stderr: %s", err, errOut)
	}
	quotaRaw, errOutQ, errQ := runCmd(ctx, "aws", "service-quotas", "get-service-quota",
		"--service-code", "ec2",
		"--quota-code", "L-0263D0A3",
		"--query", "Quota.Value",
		"--output", "text")
	if errQ != nil {
		awsLogger.Warningf("service-quotas not available or failed: %v, stderr: %s", errQ, errOutQ)
		awsLogger.Warningf("Skipping Elastic IP quota enforcement.")
		return
	}
	awsLogger.Printf("Current EIPs allocated: %d", currentCount)
	awsLogger.Printf("Required EIPs for install: %d", required)
	awsLogger.Printf("Elastic IP quota: %s", quotaRaw)
	if quotaRaw != "Unknown" && quotaRaw != "" {
		qFloat, _ := strconv.ParseFloat(quotaRaw, 64)
		qInt := int(qFloat)
		if currentCount+required > qInt {
			awsLogger.Fatalf(nil, "Not enough Elastic IP quota. Have %d, quota %d, need %d.",
				currentCount, qInt, currentCount+required)
		}
	}
	awsLogger.Successf("Elastic IP quota is sufficient.")
}

func ensureNATs(cfg Config, ctx context.Context, vpcID, pubA, pubB, pubC string) (string, string, string) {
	awsLogger.Infof("Ensuring NAT Gateways exist...")
	createdNewNats := false
	natSubnets := map[string]string{"a": pubA, "b": pubB, "c": pubC}
	natIDs := map[string]string{}
	// First, check which NATs need to be created
	for zone, subnet := range natSubnets {
		natID, _, _ := runCmd(ctx, "aws", "ec2", "describe-nat-gateways",
			"--filter",
			"Name=vpc-id,Values="+vpcID,
			"Name=subnet-id,Values="+subnet,
			"Name=state,Values=available",
			"--query", "NatGateways[0].NatGatewayId",
			"--output", "text")
		if !(natID == "" || natID == "None" || natID == "null") {
			natIDs[zone] = natID
		}
	}
	missingNATs := len(natSubnets) - len(natIDs)
	if missingNATs > 0 {
		ensureElasticIPQuota(ctx, missingNATs)
	}
	for zone, subnet := range natSubnets {
		if natIDs[zone] != "" {
			awsLogger.Successf("Reusing NAT Gateway: %s", natIDs[zone])
			tagExistingNatResources(cfg, ctx, natIDs[zone], zone)
			continue
		}
		awsLogger.Infof("Creating NAT Gateway in subnet %s...", subnet)
		alloc := mustRun(ctx, "aws", "ec2", "allocate-address",
			"--domain", "vpc",
			"--query", "AllocationId",
			"--output", "text")
		tagEC2Resource(ctx, alloc, resourceName(cfg, "nat-eip", zone))
		natID := mustRun(ctx, "aws", "ec2", "create-nat-gateway",
			"--subnet-id", subnet,
			"--allocation-id", alloc,
			"--query", "NatGateway.NatGatewayId",
			"--output", "text")
		tagEC2Resource(ctx, natID, resourceName(cfg, "nat-gateway", zone))
		natIDs[zone] = natID
		createdNewNats = true
	}

	if createdNewNats {
		args := []string{"ec2", "wait", "nat-gateway-available", "--nat-gateway-ids"}
		args = append(args, slices.Collect(maps.Values(natIDs))...)
		mustRun(ctx, "aws", args...)
		awsLogger.Successf("New NAT Gateways are available.")
	}
	awsLogger.Printf("NAT Gateways: %s, %s, %s", natIDs["a"], natIDs["b"], natIDs["c"])
	return natIDs["a"], natIDs["b"], natIDs["c"]
}

func tagExistingNatResources(cfg Config, ctx context.Context, natID, label string) {
	tagEC2Resource(ctx, natID, resourceName(cfg, "nat-gateway", label))
	allocs, errOut, err := runCmd(ctx, "aws", "ec2", "describe-nat-gateways",
		"--nat-gateway-ids", natID,
		"--query", "NatGateways[].NatGatewayAddresses[].AllocationId",
		"--output", "text")
	if err != nil {
		awsLogger.Warningf("Failed to describe NAT Gateway %s for tagging: %v\nstderr: %s", natID, err, errOut)
		return
	}
	allocs = strings.TrimSpace(allocs)
	if allocs == "" || allocs == "None" || allocs == "null" {
		return
	}
	for _, alloc := range strings.Fields(allocs) {
		tagEC2Resource(ctx, alloc, resourceName(cfg, "nat-eip", label))
	}
}

func ensurePrivateRT(ctx context.Context, vpcID, privSubnet, natID, name string) string {
	rtID, _, _ := runCmd(ctx, "aws", "ec2", "describe-route-tables",
		"--filters",
		"Name=vpc-id,Values="+vpcID,
		"Name=association.subnet-id,Values="+privSubnet,
		"--query", "RouteTables[?Routes[?NatGatewayId=='"+natID+"' && DestinationCidrBlock=='0.0.0.0/0']].RouteTableId | [0]",
		"--output", "text")
	if rtID == "" || rtID == "None" || rtID == "null" {
		awsLogger.Infof("Creating private route table for subnet %s...", privSubnet)
		rtID = mustRun(ctx, "aws", "ec2", "create-route-table",
			"--vpc-id", vpcID,
			"--query", "RouteTable.RouteTableId",
			"--output", "text")
		tagEC2Resource(ctx, rtID, name)
		mustRun(ctx, "aws", "ec2", "create-route",
			"--route-table-id", rtID,
			"--destination-cidr-block", "0.0.0.0/0",
			"--nat-gateway-id", natID)
	} else {
		awsLogger.Successf("Reusing private route table: %s", rtID)
	}
	ensureRouteTableAssociation(ctx, rtID, privSubnet)
	return rtID
}

func ensureSecurityGroup(cfg Config, ctx context.Context, vpcID, sgName string) string {
	awsLogger.Infof("Ensuring security group exists...")
	sgID, _, _ := runCmd(ctx, "aws", "ec2", "describe-security-groups",
		"--filters",
		"Name=vpc-id,Values="+vpcID,
		"Name=group-name,Values="+sgName,
		"--query", "SecurityGroups[0].GroupId",
		"--output", "text")
	if sgID == "" || sgID == "None" || sgID == "null" {
		sgID = mustRun(ctx, "aws", "ec2", "create-security-group",
			"--group-name", sgName,
			"--description", fmt.Sprintf("Restricts traffic for Klutch %s components and worker nodes", klutchRoleLabel),
			"--vpc-id", vpcID,
			"--query", "GroupId",
			"--output", "text")
		tagEC2Resource(ctx, sgID, resourceName(cfg, "security-group"))
		awsLogger.Successf("Created security group: %s", sgID)
	} else {
		awsLogger.Successf("Reusing security group: %s", sgID)
	}
	_, _, _ = runCmd(ctx, "aws", "ec2", "authorize-security-group-ingress",
		"--group-id", sgID,
		"--protocol", "-1",
		"--source-group", sgID)
	_, _, _ = runCmd(ctx, "aws", "ec2", "authorize-security-group-egress",
		"--group-id", sgID,
		"--protocol", "-1",
		"--cidr", "0.0.0.0/0")

	awsLogger.Printf("CLUSTER_SG_ID = %s", sgID)
	return sgID
}

func createEKSCluster(ctx context.Context, cfg Config, vpcID, keyArn, accountID, clusterArn string) {
	privA := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+cfg.PrivACIDR,
		"--query", "Subnets[0].SubnetId", "--output", "text")
	privB := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+cfg.PrivBCIDR,
		"--query", "Subnets[0].SubnetId", "--output", "text")
	privC := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+cfg.PrivCCIDR,
		"--query", "Subnets[0].SubnetId", "--output", "text")
	sgID := mustRun(ctx, "aws", "ec2", "describe-security-groups",
		"--filters",
		"Name=vpc-id,Values="+vpcID,
		"Name=group-name,Values="+cfg.ControlPlaneSGName,
		"--query", "SecurityGroups[0].GroupId",
		"--output", "text")

	subnets := strings.Join([]string{privA, privB, privC}, ",")
	awsLogger.Infof("Creating EKS cluster '%s'...", cfg.ClusterName)
	mustRun(ctx, "aws", "eks", "create-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--role-arn", fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, cfg.ClusterRoleName),
		"--resources-vpc-config", fmt.Sprintf("subnetIds=%s,securityGroupIds=%s", subnets, sgID),
		"--encryption-config", fmt.Sprintf("[{\"resources\":[\"secrets\"],\"provider\":{\"keyArn\":\"%s\"}}]", keyArn),
		"--tags", fmt.Sprintf("%s=%s,Name=%s,%s=%s,%s=%s", klutchTagKey, klutchTagValue, cfg.ClusterName, clusterNameTagKey, cfg.ClusterName, clusterIDTagKey, clusterArn),
	)
	awsLogger.Infof("Waiting for EKS cluster '%s' to become ACTIVE...", cfg.ClusterName)
	mustRun(ctx, "aws", "eks", "wait", "cluster-active",
		"--name", cfg.ClusterName,
		"--region", cfg.Region)
	awsLogger.Successf("Cluster is ACTIVE.")
}

func ensureDefaultEBSEncryption(ctx context.Context) {
	awsLogger.Infof("Ensuring AWS account default EBS encryption is enabled...")
	out, _, err := runCmd(ctx, "aws", "ec2", "get-ebs-encryption-by-default",
		"--query", "EbsEncryptionByDefault", "--output", "text")
	if err != nil {
		awsLogger.Warningf("get-ebs-encryption-by-default failed, skipping. Out=%s", out)
		return
	}
	if out != "true" {
		awsLogger.Warningf("Default EBS encryption is currently disabled. Enabling now...")
		mustRun(ctx, "aws", "ec2", "enable-ebs-encryption-by-default")
		awsLogger.Successf("Default EBS encryption has been enabled.")
	} else {
		awsLogger.Successf("Default EBS encryption is already enabled.")
	}
}

func ensureNodegroup(ctx context.Context, cfg Config, vpcID, accountID string) {
	awsLogger.Infof("Checking nodegroup '%s'...", cfg.NodegroupName)
	status, errOut, err := runCmd(ctx, "aws", "eks", "describe-nodegroup",
		"--cluster-name", cfg.ClusterName,
		"--nodegroup-name", cfg.NodegroupName,
		"--region", cfg.Region,
		"--query", "nodegroup.status",
		"--output", "text")
	if err != nil && !strings.Contains(errOut, "ResourceNotFoundException") {
		awsLogger.Fatalf(err, "describe-nodegroup failed\nstderr: %s", errOut)
	}
	if status == "" {
		status = "NONE"
	}

	privA := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+getenv("PRIV_A_CIDR", "10.0.101.0/24"),
		"--query", "Subnets[0].SubnetId", "--output", "text")
	privB := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+getenv("PRIV_B_CIDR", "10.0.102.0/24"),
		"--query", "Subnets[0].SubnetId", "--output", "text")
	privC := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+getenv("PRIV_C_CIDR", "10.0.103.0/24"),
		"--query", "Subnets[0].SubnetId", "--output", "text")

	create := func() {
		awsLogger.Infof("Creating nodegroup '%s'...", cfg.NodegroupName)
		mustRun(ctx, "aws", "eks", "create-nodegroup",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--scaling-config", cfg.NodeScalingConfig,
			"--instance-types", cfg.NodeInstanceTypes,
			"--subnets", privA, privB, privC,
			"--ami-type", cfg.NodeAMIType,
			"--node-role", fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, cfg.NodeRoleName),
			"--region", cfg.Region,
			"--tags", fmt.Sprintf("%s=%s,Name=%s,%s=%s,%s=%s", klutchTagKey, klutchTagValue, cfg.NodegroupName, clusterNameTagKey, cfg.ClusterName, clusterIDTagKey, currentClusterArn))
		awsLogger.Infof("Waiting for nodegroup '%s' to become ACTIVE...", cfg.NodegroupName)
		mustRun(ctx, "aws", "eks", "wait", "nodegroup-active",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		awsLogger.Successf("Nodegroup is ACTIVE.")
	}

	switch status {
	case "NONE":
		awsLogger.Infof("Nodegroup does not exist. Creating...")
		create()
	case "ACTIVE":
		awsLogger.Successf("Nodegroup already exists and is ACTIVE. Reusing it.")
	case "CREATING":
		awsLogger.Warningf("Nodegroup is in CREATING. Waiting until ACTIVE...")
		mustRun(ctx, "aws", "eks", "wait", "nodegroup-active",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		awsLogger.Successf("Nodegroup is ACTIVE.")
	case "DELETING":
		awsLogger.Fatalf(nil, "Nodegroup is currently DELETING. Wait for it to finish and re-run the installer.")
	default:
		awsLogger.Warningf("Nodegroup is in bad state: %s. Deleting and recreating...", status)
		mustRun(ctx, "aws", "eks", "delete-nodegroup",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		awsLogger.Infof("Waiting for nodegroup '%s' to be deleted...", cfg.NodegroupName)
		mustRun(ctx, "aws", "eks", "wait", "nodegroup-deleted",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		create()
	}
}

func waitForNodesReady(ctx context.Context) {
	awsLogger.Section("Cluster Nodes")
	awsLogger.Infof("Waiting for at least one Ready node...")
	k8sClient := k8s.NewKubeClient("")
	for {
		out, err := k8sClient.Get("nodes", "", "", "", true)
		makeup.PrintInfo(fmt.Sprintf("out='%s', err='%s'", out, err))
		if err == nil && strings.Contains(out, " Ready") {
			awsLogger.Successf("Nodes are Ready:")
			makeup.Print(out)
			return
		}
		awsLogger.Warningf("No Ready nodes yet, sleeping 10s...")
		time.Sleep(10 * time.Second)
	}
}

func ensureGp3StorageClass(ctx context.Context) {
	awsLogger.Infof("Creating gp3 StorageClass and setting it as default...")
	yaml := `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: ebs.csi.aws.com
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
parameters:
  type: gp3
  encrypted: "true"
`

	// Use the kubectl client to apply the manifest
	k8sClient := k8s.NewKubeClient("")
	if _, err := k8sClient.ApplyWithPrompt([]byte(yaml), "gp3 StorageClass"); err != nil {
		awsLogger.Fatalf(err, "❌ Failed to apply gp3 StorageClass: %v", err)
	}
	awsLogger.Successf("gp3 StorageClass installed and set as default.")
}

func ensureVpcDnsEnabled(ctx context.Context, vpcID string) {
	awsLogger.Infof("Ensuring VPC %s has DNS support and hostnames enabled...", vpcID)

	// Enable DNS support
	if _, errOut, err := runCmd(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-support", "{\"Value\":true}"); err != nil {
		awsLogger.Warningf("Failed to enable DNS support on VPC %s (continued): %v\nstderr: %s", vpcID, err, errOut)
	} else {
		awsLogger.Successf("Enabled DNS support on VPC %s.", vpcID)
	}

	// Enable DNS hostnames
	if _, errOut, err := runCmd(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-hostnames", "{\"Value\":true}"); err != nil {
		awsLogger.Warningf("Failed to enable DNS hostnames on VPC %s (continued): %v\nstderr: %s", vpcID, err, errOut)
	} else {
		awsLogger.Successf("Enabled DNS hostnames on VPC %s.", vpcID)
	}
}

func ensureALBController(ctx context.Context, cfg Config, vpcID, accountID string) {
	awsLogger.Section("AWS Load Balancer Controller")
	awsLogger.Infof("Installing AWS Load Balancer Controller...")

	ensureVpcDnsEnabled(ctx, vpcID)

	mustRun(ctx, "eksctl", "utils", "associate-iam-oidc-provider",
		"--region", cfg.Region,
		"--cluster", cfg.ClusterName,
		"--approve")

	// Always use the latest policy from main to ensure required permissions (e.g., DescribeRouteTables) are present.
	policyURL := "https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/main/docs/install/iam_policy.json"
	awsLogger.Printf("Using AWS Load Balancer Controller version: %s", cfg.ALBControllerVersion)
	awsLogger.Printf("IAM policy URL: %s", policyURL)

	_, errOut, err := runCmd(ctx, "curl", "-sSfL", "-o", "aws-load-balancer-controller-policy.json", policyURL)
	if err != nil {
		awsLogger.Fatalf(err, "curl failed\nstderr: %s", errOut)
	}

	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", accountID, cfg.ALBControllerPolicyName)
	if _, _, err := runCmd(ctx, "aws", "iam", "get-policy", "--policy-arn", policyArn); err != nil {
		awsLogger.Infof("Creating IAM policy %s", cfg.ALBControllerPolicyName)
		args := []string{
			"iam", "create-policy",
			"--policy-name", cfg.ALBControllerPolicyName,
			"--policy-document", "file://aws-load-balancer-controller-policy.json",
			"--description", fmt.Sprintf("Allows the Klutch %s to run the AWS Load Balancer Controller safely", klutchRoleLabel),
			"--tags",
		}
		args = append(args, append([]string{
			fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
			fmt.Sprintf("Key=Name,Value=%s", cfg.ALBControllerPolicyName),
		}, clusterTagPairsKV()...)...)
		mustRun(ctx, "aws", args...)
	} else {
		awsLogger.Successf("IAM policy %s already exists.", cfg.ALBControllerPolicyName)
		awsLogger.Infof("Updating IAM policy %s to latest version...", cfg.ALBControllerPolicyName)
		ensurePolicyVersion(ctx, policyArn, "file://aws-load-balancer-controller-policy.json")
	}

	awsLogger.Infof("Creating IAM service account for AWS Load Balancer Controller...")
	mustRun(ctx, "eksctl", "create", "iamserviceaccount",
		"--cluster", cfg.ClusterName,
		"--namespace", "kube-system",
		"--name", cfg.AlbServiceAccountName,
		"--attach-policy-arn", policyArn,
		"--region", cfg.Region,
		"--approve",
		"--override-existing-serviceaccounts")

	awsLogger.Infof("Installing AWS Load Balancer Controller via Helm...")
	_, stdErr, err := runCmd(ctx, "helm", "repo", "add", "eks", "https://aws.github.io/eks-charts")
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not add EKS helm repo:\n%s", stdErr))
	}
	mustRun(ctx, "helm", "repo", "update")

	args := []string{
		"upgrade", "--install", cfg.AlbServiceAccountName, "eks/aws-load-balancer-controller",
		"-n", "kube-system",
		"--set", "clusterName=" + cfg.ClusterName,
		"--set", "region=" + cfg.Region,
		"--set", "vpcId=" + vpcID,
		"--set", "serviceAccount.create=false",
		"--set", "serviceAccount.name=" + cfg.AlbServiceAccountName,
	}
	mustRun(ctx, "helm", args...)

	// Re-attach the managed policy to the role derived from the service account annotation
	// to ensure the controller has the updated permissions.
	roleName := getALBControllerRoleName(ctx, cfg)
	if roleName != "" {
		awsLogger.Infof("Attaching managed policy %s to role %s to ensure ALB controller permissions are present...", cfg.ALBControllerPolicyName, roleName)
		if _, errOut, err := runCmd(ctx, "aws", "iam", "attach-role-policy",
			"--role-name", roleName,
			"--policy-arn", policyArn); err != nil {
			awsLogger.Warningf("Failed to attach policy to role %s (continued): %v\nstderr: %s", roleName, err, errOut)
		} else {
			awsLogger.Successf("Attached managed policy to role %s.", roleName)
		}
	}

	k8sClient := k8s.NewKubeClient("")

	awsLogger.Infof("Waiting for aws-load-balancer-controller deployment rollout...")
	if errOut, err := k8sClient.RolloutStatus("deployment", cfg.AlbServiceAccountName, "kube-system", ""); err != nil {
		awsLogger.Fatalf(err, "❌ ALB controller rollout failed\nstderr: %s", errOut)
	}
	awsLogger.Successf("aws-load-balancer-controller deployment is ready.")
}

func getALBControllerRoleName(ctx context.Context, cfg Config) string {
	type sa struct {
		Metadata struct {
			Annotations map[string]string `json:"annotations"`
		} `json:"metadata"`
	}

	k8sClient := k8s.NewKubeClient("")
	out, err := k8sClient.Get("sa", cfg.AlbServiceAccountName, "kube-system", "json", false)
	if err != nil {
		awsLogger.Warningf("Could not fetch service account to derive role name: %v\nstderr: %s", err, out)
		return ""
	}

	var s sa
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		awsLogger.Warningf("Could not parse service account json to derive role name: %v", err)
		return ""
	}

	roleArn := s.Metadata.Annotations["eks.amazonaws.com/role-arn"]
	if roleArn == "" {
		awsLogger.Warningf("Service account is missing eks.amazonaws.com/role-arn annotation; cannot attach inline policy automatically.")
		return ""
	}

	parts := strings.Split(roleArn, "/")
	roleName := parts[len(parts)-1]
	return roleName
}

// ensureRouteTableAssociation makes sure the given subnet is associated with the desired route table.
// If already associated, it is left untouched. If associated elsewhere, it is replaced.
func ensureRouteTableAssociation(ctx context.Context, rtID, subnetID string) {
	type assoc struct {
		RouteTableId  string `json:"RouteTableId"`
		AssociationId string `json:"AssociationId"`
		Main          bool   `json:"Main"`
	}
	type rt struct {
		Associations []assoc `json:"Associations"`
	}
	type describe struct {
		RouteTables []rt `json:"RouteTables"`
	}

	out, _, err := runCmd(ctx, "aws", "ec2", "describe-route-tables",
		"--filters", "Name=association.subnet-id,Values="+subnetID,
		"--query", "RouteTables[*].Associations[]",
		"--output", "json")
	if err == nil {
		var desc []assoc
		if err := json.Unmarshal([]byte(out), &desc); err == nil && len(desc) > 0 {
			for _, a := range desc {
				if a.RouteTableId == rtID {
					awsLogger.Successf("Subnet %s already associated with route table %s. Skipping.", subnetID, rtID)
					return
				}
				// Replace existing association
				if a.Main {
					awsLogger.Infof("Replacing main route table association %s for subnet %s with %s...", a.AssociationId, subnetID, rtID)
					_, errOut, err := runCmd(ctx, "aws", "ec2", "replace-route-table-association",
						"--association-id", a.AssociationId,
						"--route-table-id", rtID)
					if err != nil {
						awsLogger.Warningf("Failed to replace route table association for subnet %s: %v\nstderr: %s", subnetID, err, errOut)
					}
					return
				}
				awsLogger.Infof("Disassociating subnet %s from route table %s and associating with %s...", subnetID, a.RouteTableId, rtID)
				_, errOut, err := runCmd(ctx, "aws", "ec2", "disassociate-route-table",
					"--association-id", a.AssociationId)
				if err != nil {
					awsLogger.Warningf("Failed to disassociate subnet %s: %v\nstderr: %s", subnetID, err, errOut)
				}
				break
			}
		}
	}

	awsLogger.Infof("Associating subnet %s with route table %s...", subnetID, rtID)
	_, errOut, err := runCmd(ctx, "aws", "ec2", "associate-route-table",
		"--route-table-id", rtID,
		"--subnet-id", subnetID)
	if err != nil {
		awsLogger.Warningf("Failed to associate subnet %s with route table %s: %v\nstderr: %s", subnetID, rtID, err, errOut)
	} else {
		awsLogger.Successf("Associated subnet %s with route table %s.", subnetID, rtID)
	}
}

// ensurePolicyVersion sets a new default policy version from the given document,
// pruning an old non-default version if necessary to stay within the 5-version limit.
func ensurePolicyVersion(ctx context.Context, policyArn, policyDocument string) {
	versionsOut, _, err := runCmd(ctx, "aws", "iam", "list-policy-versions", "--policy-arn", policyArn)
	if err == nil {
		type version struct {
			VersionId        string `json:"VersionId"`
			IsDefaultVersion bool   `json:"IsDefaultVersion"`
		}
		type list struct {
			Versions []version `json:"Versions"`
		}
		var lv list
		if err := json.Unmarshal([]byte(versionsOut), &lv); err == nil {
			if len(lv.Versions) >= 5 {
				for _, v := range lv.Versions {
					if !v.IsDefaultVersion {
						awsLogger.Infof("Deleting old policy version %s to make room for an updated ALB controller policy...", v.VersionId)
						_, errOut, err := runCmd(ctx, "aws", "iam", "delete-policy-version", "--policy-arn", policyArn, "--version-id", v.VersionId)
						if err != nil {
							awsLogger.Warningf("Failed to delete policy version %s (continued): %v\nstderr: %s", v.VersionId, err, errOut)
						}
						break
					}
				}
			}
		}
	}

	if _, errOut, err := runCmd(ctx, "aws", "iam", "create-policy-version",
		"--policy-arn", policyArn,
		"--policy-document", policyDocument,
		"--set-as-default"); err != nil {
		awsLogger.Warningf("Failed to update policy %s to latest version (continued): %v\nstderr: %s", policyArn, err, errOut)
	} else {
		awsLogger.Successf("Updated IAM policy %s to latest version.", policyArn)
	}
}
