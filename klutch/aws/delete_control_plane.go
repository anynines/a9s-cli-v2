package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
)

// DeleteOptions configures the teardown of the Klutch control plane on AWS.
type DeleteOptions struct {
	Region                  string
	ClusterName             string
	NodegroupName           string
	ALBControllerPolicyName string
	ControlPlaneSGName      string
	HostedZoneName          string
	IncludeDNSRecords       bool
	IncludeHostedZone       bool
	IncludeSSLCertificate   bool
	ACMCertificateARN       string
	DryRun                  bool
	ForceDNS                bool
	CleanupOrphans          bool
	SkipPrompt              bool
	ScheduleKmsDeletion     bool
}

// DeleteControlPlaneCluster tears down the EKS control plane and AWS resources that were created by CreateControlPlaneCluster.
// It mirrors the flow of 70-delete-eks-control-plane-cluster.sh with safe defaults (DNS/ACM are opt-in).
func DeleteControlPlaneCluster(ctx context.Context, opts DeleteOptions) {
	cfg := defaultConfig()
	if opts.Region != "" {
		cfg.Region = opts.Region
	}
	if opts.ClusterName != "" {
		cfg.ClusterName = opts.ClusterName
	}
	cfg.NodegroupName = fmt.Sprintf("%s-nodegroup", opts.ClusterName)
	cfg.ClusterIamRoleName += "-" + cfg.ClusterName
	cfg.NodeRoleName += "-" + cfg.ClusterName
	cfg.ALBControllerPolicyName += "-" + cfg.ClusterName
	cfg.ControlPlaneSGName = fmt.Sprintf("%s-sg", cfg.ClusterName)
	cfg.ResourceNamePrefix = cfg.ClusterName

	deleteCluster(ctx, cfg, opts)
}

// DeleteWorkloadCluster deletes a Klutch workload EKS cluster and its AWS resources.
func DeleteWorkloadCluster(ctx context.Context, opts DeleteOptions) {
	cfg := workloadConfig(opts.ClusterName)
	if opts.Region != "" {
		cfg.Region = opts.Region
	}
	if opts.ClusterName != "" {
		cfg.ClusterName = opts.ClusterName
	}

	deleteCluster(ctx, cfg, opts)
}

func deleteCluster(ctx context.Context, cfg Config, opts DeleteOptions) {
	restore := setKlutchContext(cfg)
	defer restore()

	if opts.NodegroupName != "" {
		cfg.NodegroupName = opts.NodegroupName
	}
	if opts.ALBControllerPolicyName != "" {
		cfg.ALBControllerPolicyName = opts.ALBControllerPolicyName
	}
	if opts.ControlPlaneSGName != "" {
		cfg.ControlPlaneSGName = opts.ControlPlaneSGName
	}
	opts.Region = cfg.Region
	if opts.IncludeHostedZone && !opts.IncludeDNSRecords {
		opts.IncludeDNSRecords = true
	}

	awsLogger.Section(fmt.Sprintf("Klutch %s Deletion (AWS)", cfg.ClusterRole))
	if opts.DryRun {
		awsLogger.Infof("Dry-run enabled: no changes will be made. Showing planned actions and resources.")
	}
	awsLogger.Printf("Region:                           %s", cfg.Region)
	awsLogger.Printf("EKS Cluster Name:                 %s", cfg.ClusterName)
	awsLogger.Printf("EKS Nodegroup Name:               %s", cfg.NodegroupName)
	awsLogger.Printf("ALB Controller IAM Policy Name:   %s", cfg.ALBControllerPolicyName)
	awsLogger.Printf("Cluster Security Group Name:      %s", cfg.ControlPlaneSGName)
	awsLogger.Printf("Hosted Zone Name:                 %s", defaultString(opts.HostedZoneName, "<not set>"))
	awsLogger.Printf("Include DNS Records:              %t", opts.IncludeDNSRecords)
	awsLogger.Printf("Include Hosted Zone:              %t", opts.IncludeHostedZone)
	awsLogger.Printf("Include SSL Certificate:          %t", opts.IncludeSSLCertificate)
	awsLogger.Printf("ACM Certificate ARN:              %s", defaultString(opts.ACMCertificateARN, "<auto-discover>"))
	awsLogger.Printf("Dry Run:                          %t", opts.DryRun)
	awsLogger.Printf("Force DNS:                        %t", opts.ForceDNS)

	if !opts.DryRun {
		if opts.SkipPrompt {
			awsLogger.Infof("Skipping delete confirmation because --yes and --really were provided.")
		} else {
			prompt := fmt.Sprintf("This will delete the Klutch %s on AWS. Type 'yes' to continue: ", strings.ToLower(cfg.ClusterRole))
			if !makeup.ConfirmYes(prompt) {
				makeup.PrintInfo("Deletion aborted.")
				return
			}
		}
	}

	for _, cmd := range []string{"aws", "kubectl", "eksctl", "helm", "jq"} {
		if _, err := execLookPath(cmd); err != nil {
			awsLogger.Fatalf(err, "Required command %q is not installed or not in PATH", cmd)
		}
	}

	out, err := runCmd(ctx, "aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	if err != nil || out == "" || out == "None" || out == "null" {
		awsLogger.Fatalf(err, "Unable to determine AWS Account ID. Run 'aws configure'. stderr: %s", out)
	}
	awsLogger.Infof("AWS Account ID: %s", out)

	clusterExists, clusterReachable := discoverCluster(ctx, cfg, opts)

	if clusterReachable {
		kubernetesCleanup(ctx, cfg, opts)
	}
	iamCleanup(ctx, cfg, opts, out, clusterReachable)

	if clusterExists {
		deleteNodegroupsAndCluster(ctx, cfg, opts)
	} else {
		awsLogger.Infof("Cluster does not exist. Skipping nodegroup/cluster deletion.")
	}

	klutchKmsKeyCleanup(cfg, ctx, opts)

	// Always remove the ClusterName tags all Hosted Zones
	// to free the zones for adoption by a new cluster.
	removeClusterNameTagFromAllHostedZones(ctx, opts)

	vpcID := findKlutchVPC(cfg, ctx, opts.Region)

	if vpcID == "" {
		awsLogger.Infof("No Klutch VPC found.")
		dnsAndACMCleanup(ctx, nil, opts)
		if opts.CleanupOrphans {
			cleanupTaggedEIPs(cfg, ctx, opts)
		}
		return
	}

	deleteVPCDependencies(ctx, vpcID, opts)
	deleteVPC(ctx, vpcID, opts)
	if opts.CleanupOrphans {
		cleanupTaggedEIPs(cfg, ctx, opts)
	}
	klutchKmsKeyCleanup(cfg, ctx, opts)
}

func discoverCluster(ctx context.Context, cfg Config, opts DeleteOptions) (bool, bool) {
	awsLogger.Section("Discover EKS Cluster")
	clusterExists := false
	clusterReachable := false
	status := "UNKNOWN"

	out, err := runCmd(ctx, "aws", "eks", "describe-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "cluster.status",
		"--output", "text")
	if err == nil && strings.TrimSpace(out) != "" {
		clusterExists = true
		status = strings.TrimSpace(out)
		awsLogger.Infof("Cluster status: %s", status)
		if status == "ACTIVE" {
			if opts.DryRun {
				clusterReachable = true
			} else {
				if errOut, err := runCmdWithPrompt(ctx, "aws", "eks", "update-kubeconfig",
					"--name", cfg.ClusterName,
					"--region", cfg.Region); err != nil {
					awsLogger.Warningf("Could not update kubeconfig: %v\nstderr: %s", err, errOut)
				} else if _, err := k8s.Version(false, "5s"); err == nil {
					clusterReachable = true
				}
			}
		}
	} else if err != nil && !strings.Contains(out, "ResourceNotFoundException") {
		awsLogger.Warningf("describe-cluster failed: %v\nstderr: %s", err, out)
	}

	return clusterExists, clusterReachable
}

func kubernetesCleanup(ctx context.Context, cfg Config, opts DeleteOptions) {
	awsLogger.Section("Kubernetes Cleanup")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete storageclass gp3 (if present).")
	} else {
		k8sClient := k8s.NewKubeClient("")
		if errOut, err := k8sClient.Delete("storageclass", "gp3", "", "Remove StorageClass GP3", true); err != nil {
			awsLogger.Warningf("Failed to delete storageclass gp3: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Deleted storageclass gp3 (if present).")
		}
	}

	if opts.DryRun {
		awsLogger.Infof("Dry-run: would uninstall AWS LB Controller Helm release (if present).")
	} else {
		if errOut, err := runCmdWithPrompt(ctx, "helm", "uninstall", cfg.AlbServiceAccountName, "-n", "kube-system"); err != nil {
			awsLogger.Warningf("Failed to uninstall AWS LB Controller: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Uninstalled AWS LB Controller (if present).")
		}
	}
}

func iamCleanup(ctx context.Context, cfg Config, opts DeleteOptions, accountID string, clusterReachable bool) {
	awsLogger.Section("IAM Cleanup")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete IAM service account aws-load-balancer-controller.")
	} else if clusterReachable {
		if errOut, err := runCmdWithPrompt(ctx, "eksctl", "delete", "iamserviceaccount",
			"--cluster", cfg.ClusterName,
			"--region", cfg.Region,
			"--namespace", "kube-system",
			"--name", cfg.AlbServiceAccountName,
			"--wait"); err != nil {
			awsLogger.Warningf("Failed to delete IAM service account: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Deleted IAM service account (if present).")
		}
	} else {
		awsLogger.Warningf("Skipping IAM service account deletion via eksctl (cluster unreachable).")
	}

	if cfg.ClusterRole == clusterRoleControlPlane {
		// Delete all Tenant CRs (best-effort) before tearing down operator/IAM.
		deleteTenants(ctx, clusterReachable)
		// Attempt to delete tenant operator IAM role (stale roles break IRSA on recreate).
		deleteTenantOperatorRole(ctx, cfg, accountID, opts)
	}
	deleteOIDCProvider(ctx, cfg, accountID, opts, clusterReachable)

	policyArn, err := runCmd(ctx, "aws", "iam", "list-policies", "--scope", "Local",
		"--query", fmt.Sprintf("Policies[?PolicyName=='%s'].Arn | [0]", cfg.ALBControllerPolicyName),
		"--output", "text")
	if err != nil || policyArn == "" || policyArn == "None" {
		awsLogger.Infof("No ALB IAM policy found.")
		return
	}

	versions, _ := runCmd(ctx, "aws", "iam", "list-policy-versions",
		"--policy-arn", policyArn,
		"--query", "Versions[?IsDefaultVersion==`false`].VersionId",
		"--output", "text")
	for _, v := range strings.Fields(versions) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete policy version %s for %s", v, cfg.ALBControllerPolicyName)
			continue
		}
		if errOut, err := runCmdWithPrompt(ctx, "aws", "iam", "delete-policy-version",
			"--policy-arn", policyArn, "--version-id", v); err != nil {
			awsLogger.Warningf("Failed to delete policy version %s: %v\nstderr: %s", v, err, errOut)
		}
	}
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete ALB IAM policy %s", cfg.ALBControllerPolicyName)
	} else {
		if errOut, err := runCmdWithPrompt(ctx, "aws", "iam", "delete-policy", "--policy-arn", policyArn); err != nil {
			awsLogger.Warningf("Failed to delete ALB IAM policy: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Deleted ALB IAM policy.")
		}
	}
}

func deleteNodegroupsAndCluster(ctx context.Context, cfg Config, opts DeleteOptions) {
	awsLogger.Section("Delete Nodegroups and Cluster")

	ngs, _ := runCmd(ctx, "aws", "eks", "list-nodegroups",
		"--cluster-name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "nodegroups[]",
		"--output", "text")

	for _, ng := range strings.Fields(ngs) {
		awsLogger.Infof("Deleting nodegroup: %s", ng)
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete nodegroup %s and wait for deletion.", ng)
			continue
		}
		if errOut, err := runCmdWithPrompt(ctx, "aws", "eks", "delete-nodegroup",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", ng,
			"--region", cfg.Region); err != nil {
			awsLogger.Warningf("Failed to request deletion for nodegroup %s: %v\nstderr: %s", ng, err, errOut)
		}
		if errOut, err := runCmd(ctx, "aws", "eks", "wait", "nodegroup-deleted",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", ng,
			"--region", cfg.Region); err != nil {
			awsLogger.Warningf("Wait for nodegroup %s deletion failed: %v\nstderr: %s", ng, err, errOut)
		}
	}

	awsLogger.Infof("Deleting cluster...")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete cluster %s and wait for deletion.", cfg.ClusterName)
	} else {
		if errOut, err := runCmdWithPrompt(ctx, "aws", "eks", "delete-cluster",
			"--name", cfg.ClusterName,
			"--region", cfg.Region); err != nil {
			awsLogger.Warningf("Failed to request cluster deletion: %v\nstderr: %s", err, errOut)
		}
		if errOut, err := runCmd(ctx, "aws", "eks", "wait", "cluster-deleted",
			"--name", cfg.ClusterName,
			"--region", cfg.Region); err != nil {
			awsLogger.Warningf("Wait for cluster deletion failed: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Cluster deleted.")
		}
	}
}
func klutchKmsKeyCleanup(cfg Config, ctx context.Context, opts DeleteOptions) {
	awsLogger.Section("Disable Klutch KMS Keys")

	args := []string{"resourcegroupstaggingapi", "get-resources",
		"--tag-filters", fmt.Sprintf("Key=%s,Values=%s", klutchTagKey, klutchTagValue),
		"--resource-type-filters", "kms",
		"--query", "ResourceTagMappingList[].ResourceARN",
		"--output", "text"}
	args = appendRegion(args, opts.Region)

	out, errOut, err := runCmd(ctx, "aws", args...)
	if err != nil {
		awsLogger.Warningf("Failed to list Klutch-tagged KMS keys: %v\nstderr: %s", err, errOut)
		return
	}

	keyARNs := strings.Fields(out)
	if len(keyARNs) == 0 {
		awsLogger.Infof("No Klutch-tagged KMS keys found.")
		return
	}

	awsLogger.Infof("Found %d Klutch-tagged KMS key(s).", len(keyARNs))
	if opts.ScheduleKmsDeletion {
		awsLogger.Infof(`"--schedule-kms-deletion" flag is active - will attempt to schedule the keys to be deleted in 7 days.`)
	}

	keysDisablingFailed := []string{}
	keysDeletionSchedulingFailed := []string{}
	keysRetaggingFailed := []string{}
	for _, arn := range keyARNs {
		awsLogger.Infof("Disabling KMS key %s...", arn)
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would disable KMS key %s.", arn)
			awsLogger.Infof("Dry-run: would retag KMS key %s.", arn)
			if opts.ScheduleKmsDeletion {
				awsLogger.Infof("Dry-run: would schedule deletion for KMS key %s.", arn)
			}
			continue
		}
		keysDisablingFailed = disableKmsKey(ctx, arn, opts, keysDisablingFailed)
		keysRetaggingFailed = retagKmsKey(cfg, ctx, arn, opts, keysRetaggingFailed)
		if opts.ScheduleKmsDeletion {
			keysDeletionSchedulingFailed = scheduleDeletionForKmsKey(ctx, arn, opts, keysDeletionSchedulingFailed)
		}
	}
	errMessages := []string{}
	if len(keysDisablingFailed) > 0 {
		errMessages = append(errMessages, logKmsCleanupFailure("Unable to disable all keys: failed to disable key", keysDisablingFailed))
	}
	if len(keysDeletionSchedulingFailed) > 0 {
		errMessages = append(errMessages, logKmsCleanupFailure("Unable to delete all keys: failed to delete disabled key", keysDeletionSchedulingFailed))
	}
	if len(keysRetaggingFailed) > 0 {
		errMessages = append(errMessages, logKmsCleanupFailure("Unable to retag all keys: failed to retag disabled key", keysRetaggingFailed))
	}
	if len(errMessages) > 0 {
		awsLogger.Fatalf(nil, strings.Join(errMessages, "\n"))
	}
}

func scheduleDeletionForKmsKey(ctx context.Context, arn string, opts DeleteOptions, keysDeletionSchedulingFailed []string) []string {
	schedArgs := []string{"kms", "schedule-key-deletion", "--key-id", arn, "--pending-window-in-days", "7"}
	schedArgs = appendRegion(schedArgs, opts.Region)
	if _, errOut, err := runCmd(ctx, "aws", schedArgs...); err != nil {
		awsLogger.Warningf("Failed to schedule deletion for KMS key %s: %v\nstderr: %s", arn, err, errOut)
		return append(keysDeletionSchedulingFailed, arn)
	}
	awsLogger.Successf("Scheduled KMS key %s for deletion in 7 days.", arn)
	return keysDeletionSchedulingFailed
}

func retagKmsKey(cfg Config, ctx context.Context, arn string, opts DeleteOptions, keysRetaggingFailed []string) []string {
	retagArgs := appendRegion([]string{"kms", "tag-resource",
		"--key-id", arn, "--tags",
		"TagKey=Name,TagValue=" + resourceName(cfg, "kms-key-retired")}, opts.Region)
	if _, errOut, err := runCmd(ctx, "aws", retagArgs...); err != nil {
		awsLogger.Warningf("Failed to retag KMS key %s as retired: %v\nstderr: %s", arn, err, errOut)
		return append(keysRetaggingFailed, arn)
	}
	awsLogger.Successf("Retagged KMS key %s as retired.", arn)
	return keysRetaggingFailed
}

func disableKmsKey(ctx context.Context, arn string, opts DeleteOptions, keysDisablingFailed []string) []string {
	disableArgs := appendRegion([]string{"kms", "disable-key",
		"--key-id", arn},
		opts.Region)
	if _, errOut, err := runCmd(ctx, "aws", disableArgs...); err != nil {
		awsLogger.Warningf("Failed to disable KMS key %s: %v\nstderr: %s", arn, err, errOut)
		return append(keysDisablingFailed, arn)
	}
	awsLogger.Successf("Disabled KMS key %s.", arn)
	return keysDisablingFailed
}

func logKmsCleanupFailure(message string, keys []string) string {
	if len(keys) == 1 {
		return fmt.Sprintf("%s %s\n", message, keys[0])
	}
	return fmt.Sprintf("%ss:\n- %s", message, strings.Join(keys, "\n- "))
}

func findKlutchVPC(cfg Config, ctx context.Context, region string) string {
	awsLogger.Section("Discover Klutch VPC")
	args := []string{"ec2", "describe-vpcs",
		"--filters", fmt.Sprintf("Name=tag:%s,Values=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Name=tag:Name,Values=%s", resourceName(cfg, "vpc")),
		"--query", "Vpcs[0].VpcId", "--output", "text"}
	args = appendRegion(args, region)
	out, err := runCmd(ctx, "aws", args...)
	if err != nil || out == "" || out == "None" {
		awsLogger.Infof("No Klutch VPC found (stderr: %s)", out)
		return ""
	}
	awsLogger.Infof("Klutch VPC: %s", out)
	return out
}

func deleteVPCDependencies(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Section("Delete VPC Dependencies")

	deleteVPCEndpoints(ctx, vpcID, opts)
	lbTargets := deleteLoadBalancers(ctx, vpcID, opts)
	waitForELBENIs(ctx, vpcID, opts)
	natEIPs := deleteNATGateways(ctx, vpcID, opts)
	deleteInternetGateway(ctx, vpcID, opts)
	deleteRouteTables(ctx, vpcID, opts)
	deleteENIs(ctx, vpcID, opts)
	deleteSubnets(ctx, vpcID, opts)
	deleteSecurityGroups(ctx, vpcID, opts)
	resetDHCPOptions(ctx, vpcID, opts)

	dnsAndACMCleanup(ctx, lbTargets, opts)
	releaseEIPs(ctx, natEIPs, opts)
}

func deleteVPCEndpoints(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting VPC Endpoints...")
	args := []string{"ec2", "describe-vpc-endpoints",
		"--filters", "Name=vpc-id,Values=" + vpcID,
		"--query", "VpcEndpoints[].VpcEndpointId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	eps, _ := runCmd(ctx, "aws", args...)
	for _, ep := range strings.Fields(eps) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete VPC endpoint %s", ep)
			continue
		}
		args = []string{"ec2", "delete-vpc-endpoints", "--vpc-endpoint-ids", ep}
		args = appendRegion(args, opts.Region)
		errOut, err := runCmdWithPrompt(ctx, "aws", args...)
		if err != nil {
			awsLogger.Warningf("Failed to delete VPC endpoint %s: %v\nstderr: %s", ep, err, errOut)
		}
	}
}

func deleteLoadBalancers(ctx context.Context, vpcID string, opts DeleteOptions) []string {
	awsLogger.Infof("Deleting Load Balancers in VPC...")
	var lbTargets []string
	args := []string{"elbv2", "describe-load-balancers",
		"--query", "LoadBalancers[?VpcId==`" + vpcID + "`].LoadBalancerArn",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	lbs, _ := runCmd(ctx, "aws", args...)
	for _, lb := range strings.Fields(lbs) {
		awsLogger.Infof("  LoadBalancer: %s", lb)

		args = []string{"elbv2", "describe-load-balancers",
			"--load-balancer-arns", lb,
			"--query", "LoadBalancers[0].DNSName",
			"--output", "text"}
		args = appendRegion(args, opts.Region)
		lbDNS, _ := runCmd(ctx, "aws", args...)
		if lbDNS != "" && lbDNS != "None" {
			lbTargets = append(lbTargets, lbDNS)
		}

		args = []string{"elbv2", "describe-target-groups",
			"--load-balancer-arn", lb,
			"--query", "TargetGroups[].TargetGroupArn",
			"--output", "text"}
		args = appendRegion(args, opts.Region)
		tgs, _ := runCmd(ctx, "aws", args...)

		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete load balancer %s and target groups %s", lb, strings.TrimSpace(tgs))
		} else {
			args = []string{"elbv2", "delete-load-balancer", "--load-balancer-arn", lb}
			args = appendRegion(args, opts.Region)
			if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
				awsLogger.Warningf("Failed to delete load balancer %s: %v\nstderr: %s", lb, err, errOut)
			}

			time.Sleep(20 * time.Second) // allow ENIs to detach

			for _, tg := range strings.Fields(tgs) {
				args = []string{"elbv2", "delete-target-group", "--target-group-arn", tg}
				args = appendRegion(args, opts.Region)
				if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
					awsLogger.Warningf("Failed to delete target group %s: %v\nstderr: %s", tg, err, errOut)
				}
			}
		}
	}
	return lbTargets
}

func waitForELBENIs(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Waiting for ELB network interfaces to be released...")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would poll for ELB ENI cleanup.")
		return
	}
	for range 10 {
		args := []string{"ec2", "describe-network-interfaces",
			"--filters", "Name=vpc-id,Values=" + vpcID,
			"--query", "NetworkInterfaces[?starts_with(Description, 'ELB ')].NetworkInterfaceId",
			"--output", "text"}
		args = appendRegion(args, opts.Region)
		enis, _ := runCmd(ctx, "aws", args...)
		if strings.TrimSpace(enis) == "" {
			awsLogger.Infof("No ELB ENIs remaining.")
			return
		}
		awsLogger.Infof("Still present ELB ENIs: %s – waiting 15s...", enis)
		time.Sleep(15 * time.Second)
	}
}

func deleteNATGateways(ctx context.Context, vpcID string, opts DeleteOptions) []string {
	awsLogger.Infof("Deleting NAT Gateways...")
	args := []string{"ec2", "describe-nat-gateways",
		"--filter", "Name=vpc-id,Values=" + vpcID,
		"--query", "NatGateways[].NatGatewayAddresses[].AllocationId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	natEIPs, _ := runCmd(ctx, "aws", args...)

	args = []string{"ec2", "describe-nat-gateways",
		"--filter", "Name=vpc-id,Values=" + vpcID,
		"--query", "NatGateways[].NatGatewayId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	ngws, _ := runCmd(ctx, "aws", args...)

	for _, ng := range strings.Fields(ngws) {
		awsLogger.Infof("Requesting deletion of NAT Gateway %s...", ng)
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete NAT Gateway %s", ng)
			continue
		}
		args = []string{"ec2", "delete-nat-gateway", "--nat-gateway-id", ng}
		args = appendRegion(args, opts.Region)
		if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
			awsLogger.Warningf("Failed to delete NAT Gateway %s: %v\nstderr: %s", ng, err, errOut)
		}
	}

	for _, ng := range strings.Fields(ngws) {
		if ng == "" {
			continue
		}
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would wait for NAT Gateway %s deletion.", ng)
			continue
		}
		for {
			args = []string{"ec2", "describe-nat-gateways",
				"--nat-gateway-ids", ng,
				"--query", "NatGateways[0].State",
				"--output", "text"}
			args = appendRegion(args, opts.Region)
			state, _ := runCmd(ctx, "aws", args...)
			if state == "deleted" || state == "nat-gateway-not-found" || state == "None" || state == "null" {
				awsLogger.Infof("NAT Gateway %s is deleted.", ng)
				break
			}
			awsLogger.Infof("NAT Gateway %s still in state '%s', waiting 15s...", ng, state)
			time.Sleep(15 * time.Second)
		}
	}
	return strings.Fields(natEIPs)
}

func deleteInternetGateway(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Detaching and deleting Internet Gateway...")
	args := []string{"ec2", "describe-internet-gateways",
		"--filters", "Name=attachment.vpc-id,Values=" + vpcID,
		"--query", "InternetGateways[0].InternetGatewayId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	igwID, _ := runCmd(ctx, "aws", args...)
	if igwID == "" || igwID == "None" {
		return
	}
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would detach and delete IGW %s", igwID)
		return
	}
	args = []string{"ec2", "detach-internet-gateway",
		"--internet-gateway-id", igwID,
		"--vpc-id", vpcID}
	args = appendRegion(args, opts.Region)
	if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
		awsLogger.Warningf("Failed to detach IGW %s: %v\nstderr: %s", igwID, err, errOut)
	}
	args = []string{"ec2", "delete-internet-gateway",
		"--internet-gateway-id", igwID}
	args = appendRegion(args, opts.Region)
	if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
		awsLogger.Warningf("Failed to delete IGW %s: %v\nstderr: %s", igwID, err, errOut)
	}
}

func deleteRouteTables(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting Route Tables...")
	args := []string{"ec2", "describe-route-tables",
		"--filters", "Name=vpc-id,Values=" + vpcID,
		"--query", "RouteTables[?!(Associations[?Main==`true`])].RouteTableId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	nonMainRTs, _ := runCmd(ctx, "aws", args...)

	for _, rt := range strings.Fields(nonMainRTs) {
		awsLogger.Infof("Processing non-main route table %s...", rt)
		args = []string{"ec2", "describe-route-tables",
			"--route-table-ids", rt,
			"--query", "RouteTables[0].Associations[].RouteTableAssociationId",
			"--output", "text"}
		args = appendRegion(args, opts.Region)
		assocs, _ := runCmd(ctx, "aws", args...)
		for assoc := range strings.FieldsSeq(assocs) {
			if opts.DryRun {
				awsLogger.Infof("Dry-run: would disassociate route table %s association %s", rt, assoc)
				continue
			}
			args = []string{"ec2", "disassociate-route-table", "--association-id", assoc}
			args = appendRegion(args, opts.Region)
			if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
				awsLogger.Warningf("Failed to disassociate route table %s: %v\nstderr: %s", assoc, err, errOut)
			}
		}
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete route table %s", rt)
		} else {
			args = []string{"ec2", "delete-route-table", "--route-table-id", rt}
			args = appendRegion(args, opts.Region)
			if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
				awsLogger.Warningf("Failed to delete route table %s: %v\nstderr: %s", rt, err, errOut)
			}
		}
	}
}

func deleteENIs(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting ENIs...")
	args := []string{"ec2", "describe-network-interfaces",
		"--filters", "Name=vpc-id,Values=" + vpcID,
		"--query", "NetworkInterfaces[].NetworkInterfaceId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	enis, _ := runCmd(ctx, "aws", args...)
	for _, eni := range strings.Fields(enis) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete ENI %s", eni)
		} else {
			args = []string{"ec2", "delete-network-interface", "--network-interface-id", eni}
			args = appendRegion(args, opts.Region)
			if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
				awsLogger.Warningf("Failed to delete ENI %s: %v\nstderr: %s", eni, err, errOut)
			}
		}
	}
}

func deleteTenantOperatorRole(ctx context.Context, cfg Config, accountID string, opts DeleteOptions) {
	roleName := resourceName(cfg, "tenant-operator")
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName)

	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete tenant operator IAM role %s (inline policies first).", roleArn)
		return
	}

	_, _ = runCmdWithPrompt(ctx, "aws", "iam", "delete-role-policy",
		"--role-name", roleName,
		"--policy-name", "TenantOperatorInline")

	if errOut, err := runCmdWithPrompt(ctx, "aws", "iam", "delete-role", "--role-name", roleName); err != nil {
		if strings.Contains(errOut, "NoSuchEntity") {
			awsLogger.Infof("Tenant operator IAM role %s not found (nothing to delete).", roleArn)
			return
		}
		awsLogger.Warningf("Failed to delete tenant operator IAM role %s: %v\nstderr: %s", roleArn, err, errOut)
		return
	}
	awsLogger.Successf("Deleted tenant operator IAM role %s.", roleArn)
}

func deleteOIDCProvider(ctx context.Context, cfg Config, accountID string, opts DeleteOptions, clusterReachable bool) {
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete IAM OIDC provider for cluster %s.", cfg.ClusterName)
		return
	}

	issuer := ""
	if clusterReachable {
		out, err := runCmd(ctx, "aws", "eks", "describe-cluster",
			"--name", cfg.ClusterName,
			"--region", cfg.Region,
			"--query", "cluster.identity.oidc.issuer",
			"--output", "text")
		if err != nil || strings.TrimSpace(out) == "" {
			awsLogger.Warningf("Failed to discover OIDC issuer for cluster %s; skipping OIDC provider deletion.\nstderr: %s", cfg.ClusterName, out)
			return
		}
		issuer = strings.TrimSpace(out)
	} else {
		awsLogger.Warningf("Cluster unreachable; skipping OIDC provider discovery/deletion.")
		return
	}

	providerHost := strings.TrimPrefix(issuer, "https://")
	providerArn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountID, providerHost)

	if errOut, err := runCmdWithPrompt(ctx, "aws", "iam", "delete-open-id-connect-provider",
		"--open-id-connect-provider-arn", providerArn); err != nil {
		if strings.Contains(errOut, "NoSuchEntity") {
			awsLogger.Infof("IAM OIDC provider %s not found (nothing to delete).", providerArn)
			return
		}
		awsLogger.Warningf("Failed to delete IAM OIDC provider %s: %v\nstderr: %s", providerArn, err, errOut)
		return
	}
	awsLogger.Successf("Deleted IAM OIDC provider %s.", providerArn)
}

func deleteTenants(ctx context.Context, clusterReachable bool) {
	if !clusterReachable {
		awsLogger.Warningf("Cluster unreachable; skipping Tenant CR deletion.")
		return
	}
	awsLogger.Section("Tenant Cleanup")
	k8sClient := k8s.NewKubeClient("")
	out, err := k8sClient.Get("tenants.klutch.anynines.com", "-A", "", "jsonpath={range .items[*]}{.metadata.namespace}/{.metadata.name}{\"\\n\"}{end}", true)
	if err != nil {
		awsLogger.Warningf("Failed to list tenants: %v\nstderr: %s", err, out)
		return
	}
	tenants := strings.Fields(out)
	if len(tenants) == 0 {
		awsLogger.Infof("No Tenant CRs found.")
		return
	}
	for _, t := range tenants {
		parts := strings.SplitN(t, "/", 2)
		ns, name := parts[0], parts[1]
		awsLogger.Infof("Deleting Tenant %s/%s...", ns, name)
		k8sClient := k8s.NewKubeClient("")
		if errOut, err := k8sClient.Delete("tenant", name, ns, "", true); err != nil {
			awsLogger.Warningf("Failed to delete Tenant %s/%s: %v\nstderr: %s", ns, name, err, errOut)
		}
	}
}

func deleteSubnets(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting Subnets...")
	args := []string{"ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values=" + vpcID,
		"--query", "Subnets[].SubnetId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	subnets, _ := runCmd(ctx, "aws", args...)
	for _, sn := range strings.Fields(subnets) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete subnet %s", sn)
		} else {
			args = []string{"ec2", "delete-subnet", "--subnet-id", sn}
			args = appendRegion(args, opts.Region)
			if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
				awsLogger.Warningf("Failed to delete subnet %s: %v\nstderr: %s", sn, err, errOut)
			}
		}
	}
}

func deleteSecurityGroups(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting non-default Security Groups in VPC...")
	args := []string{"ec2", "describe-security-groups",
		"--filters", "Name=vpc-id,Values=" + vpcID,
		"--query", "SecurityGroups[?GroupName!=`default`].GroupId",
		"--output", "text"}
	args = appendRegion(args, opts.Region)
	sgs, _ := runCmd(ctx, "aws", args...)
	if strings.TrimSpace(sgs) == "" {
		awsLogger.Infof("No non-default Security Groups found to delete.")
		return
	}
	for _, sg := range strings.Fields(sgs) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete security group %s", sg)
		} else {
			args = []string{"ec2", "delete-security-group", "--group-id", sg}
			args = appendRegion(args, opts.Region)
			if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
				awsLogger.Warningf("Failed to delete security group %s: %v\nstderr: %s", sg, err, errOut)
			}
		}
	}
}

func resetDHCPOptions(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Resetting DHCP Options to default...")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would associate default DHCP options with VPC %s", vpcID)
		return
	}
	args := []string{"ec2", "associate-dhcp-options",
		"--dhcp-options-id", "default",
		"--vpc-id", vpcID}
	args = appendRegion(args, opts.Region)
	if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
		awsLogger.Warningf("Failed to reset DHCP options: %v\nstderr: %s", err, errOut)
	}
}

func releaseEIPs(ctx context.Context, natEIPs []string, opts DeleteOptions) {
	if len(natEIPs) == 0 {
		awsLogger.Infof("No EIP AllocationIds found to release.")
		return
	}
	for _, alloc := range natEIPs {
		if alloc == "" || alloc == "None" || alloc == "null" {
			continue
		}
		awsLogger.Infof("Releasing EIP AllocationId %s...", alloc)
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would release EIP %s", alloc)
			continue
		}
		args := []string{"ec2", "release-address", "--allocation-id", alloc}
		if opts.Region != "" {
			args = append(args, "--region", opts.Region)
		}
		for attempt := 1; attempt <= 2; attempt++ {
			if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
				if attempt == 1 && isAuthError(errOut) {
					awsLogger.Warningf("Release failed for %s due to AWS authentication. Retrying once after validating credentials...", alloc)
					if idErrOut, idErr := runCmd(ctx, "aws", "sts", "get-caller-identity", "--output", "text"); idErr != nil {
						awsLogger.Warningf("AWS credentials appear invalid. Refresh them (e.g., aws sso login) and rerun:\n  aws ec2 release-address --allocation-id %s --region %s\nSTS stderr: %s", alloc, defaultString(opts.Region, "<region>"), idErrOut)
						break
					}
					time.Sleep(2 * time.Second)
					continue
				}
				awsLogger.Warningf("Failed to release EIP %s: %v\nstderr: %s", alloc, err, errOut)
			} else {
				awsLogger.Successf("Released EIP %s.", alloc)
			}
			break
		}
	}
}

func cleanupTaggedEIPs(cfg Config, ctx context.Context, opts DeleteOptions) {
	awsLogger.Section("Orphaned EIP Cleanup")

	filters := []string{fmt.Sprintf("Name=tag:%s,Values=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Name=tag:Name,Values=%s,%s,%s",
			resourceName(cfg, "nat-eip", "a"),
			resourceName(cfg, "nat-eip", "b"),
			resourceName(cfg, "nat-eip", "c"),
		),
	}
	if cn := strings.TrimSpace(opts.ClusterName); cn != "" {
		filters = append(filters, fmt.Sprintf("Name=tag:%s,Values=%s", clusterNameTagKey, cn))
	}

	args := []string{"ec2", "describe-addresses", "--filters"}
	args = append(args, filters...)
	args = append(args, "--query", "Addresses[].AllocationId", "--output", "text")
	if opts.Region != "" {
		args = append(args, "--region", opts.Region)
	}

	out, err := runCmd(ctx, "aws", args...)
	if err != nil {
		awsLogger.Warningf("Failed to list Klutch-tagged EIPs: %v\nstderr: %s", err, out)
		return
	}
	allocs := strings.Fields(out)
	if len(allocs) == 0 {
		awsLogger.Infof("No Klutch-tagged EIPs remaining.")
		return
	}

	awsLogger.Infof("Found %d Klutch-tagged EIP(s) to release post-delete.", len(allocs))
	releaseEIPs(ctx, allocs, opts)
}

func removeClusterNameTagFromAllHostedZones(ctx context.Context, opts DeleteOptions) {
	zoneID := findHostedZoneIdByClusterNameTag(ctx, opts)
	for zoneID != "" {
		removeClusterNameTagFromHostedZone(ctx, zoneID, opts)
		zoneID = findHostedZoneIdByClusterNameTag(ctx, opts)
	}
}

// findHostedZoneIdByClusterNameTag finds the Route53 hosted zone tagged with
// ClusterName=clusterName using the Resource Groups Tagging API, which supports
// server-side tag filtering (unlike aws route53 list-hosted-zones).
// Returns the plain zone ID (without the /hostedzone/ prefix), or empty string if not found.
func findHostedZoneIdByClusterNameTag(ctx context.Context, opts DeleteOptions) string {
	if opts.DryRun {
		return ""
	}

	out, err := runCmd(ctx, "aws", "resourcegroupstaggingapi", "get-resources",
		"--resource-type-filters", "route53:hostedzone",
		"--tag-filters", fmt.Sprintf("Key=ClusterName,Values=%s", opts.ClusterName),
		"--query", "ResourceTagMappingList[0].ResourceARN",
		// since Route53 is a global service the region must be set to us-east-1
		"--region", "us-east-1",
		"--output", "text")
	if err != nil {
		awsLogger.Warningf("Failed to search for hosted zone by opts.ClusterName tag: %v\nstderr: %s", err, out)
		return ""
	}
	outNormalized := strings.ToLower(strings.TrimSpace(out))
	if out == "" || outNormalized == "none" || outNormalized == "null" {
		awsLogger.Infof("Found no hosted zones tagged with ClusterName=%s.", opts.ClusterName)
		return ""
	}
	// ARN format: arn:aws:route53:::hostedzone/ZONE_ID
	id := out[strings.LastIndex(out, "/")+1:]
	awsLogger.Infof("Found hosted zone id=%s tagged with ClusterName=%s.", id, opts.ClusterName)
	return id
}

// removeClusterNameTagFromHostedZone removes the ClusterName tag from a Route53 hosted zone
// so that a new cluster can adopt it.
func removeClusterNameTagFromHostedZone(ctx context.Context, zoneID string, opts DeleteOptions) {
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would remove ClusterName tag from hosted zone %s.", zoneID)
		return
	}
	if out, err := runCmd(ctx, "aws", "route53", "change-tags-for-resource",
		"--resource-type", "hostedzone",
		"--resource-id", zoneID,
		"--remove-tag-keys", "ClusterName"); err != nil {
		awsLogger.Warningf("Failed to remove ClusterName tag from hosted zone %s: %v\nstderr: %s", zoneID, err, out)
		return
	}
	awsLogger.Successf("Removed ClusterName tag from hosted zone %s.", zoneID)
}

func findHostedZoneIDByName(ctx context.Context, hostedZoneName string) string {
	if strings.TrimSpace(hostedZoneName) == "" {
		return ""
	}
	normalized := strings.ToLower(hostedZoneName)
	if !strings.HasSuffix(normalized, ".") {
		normalized += "."
	}

	query := fmt.Sprintf("HostedZones[?Name==`%s`].Id | [0]", normalized)
	out, err := runCmd(ctx, "aws", "route53", "list-hosted-zones-by-name",
		"--dns-name", normalized,
		"--query", query,
		"--output", "text")
	if err != nil {
		awsLogger.Warningf("Failed to list hosted zones for %s: %v\nstderr: %s", hostedZoneName, err, out)
		return ""
	}
	if out == "" || out == "None" || out == "null" {
		awsLogger.Warningf("Hosted zone %s not found via Route53 list-hosted-zones-by-name.", hostedZoneName)
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(out), "/hostedzone/")
}

func listHostedZoneRecords(ctx context.Context, zoneID string) ([]map[string]interface{}, error) {
	out, err := runCmd(ctx, "aws", "route53", "list-resource-record-sets",
		"--hosted-zone-id", zoneID,
		"--query", "ResourceRecordSets[?Type!=`NS` && Type!=`SOA`]",
		"--output", "json")
	if err != nil {
		return nil, fmt.Errorf("listing resource record sets: %v (stderr: %s)", err, out)
	}
	if strings.TrimSpace(out) == "" || strings.TrimSpace(out) == "null" {
		return nil, nil
	}

	var records []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &records); err != nil {
		return nil, fmt.Errorf("parsing resource record sets: %w", err)
	}
	return records, nil
}

func deleteDNSRecords(ctx context.Context, zoneID, hostedZoneName string, opts DeleteOptions) {
	records, err := listHostedZoneRecords(ctx, zoneID)
	if err != nil {
		awsLogger.Warningf("Could not list DNS records for hosted zone %s: %v", hostedZoneName, err)
		return
	}
	if len(records) == 0 {
		awsLogger.Infof("No non-default DNS records found in hosted zone %s.", hostedZoneName)
		return
	}

	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete %d DNS record(s) from hosted zone %s.", len(records), hostedZoneName)
		return
	}

	var changes []map[string]interface{}
	for _, r := range records {
		changes = append(changes, map[string]interface{}{
			"Action":            "DELETE",
			"ResourceRecordSet": r,
		})
	}

	payload, err := json.Marshal(map[string]interface{}{
		"Changes": changes,
	})
	if err != nil {
		awsLogger.Warningf("Failed to marshal DNS deletion batch for %s: %v", hostedZoneName, err)
		return
	}

	if errOut, err := runCmdWithPrompt(ctx, "aws", "route53", "change-resource-record-sets",
		"--hosted-zone-id", zoneID,
		"--change-batch", string(payload)); err != nil {
		awsLogger.Warningf("Failed to delete DNS records in hosted zone %s: %v\nstderr: %s", hostedZoneName, err, errOut)
	} else {
		awsLogger.Successf("Deleted %d DNS record(s) from hosted zone %s.", len(records), hostedZoneName)
	}
}

func deleteHostedZone(ctx context.Context, zoneID, hostedZoneName string, opts DeleteOptions) {
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete hosted zone %s (ID %s).", hostedZoneName, zoneID)
		return
	}
	if errOut, err := runCmdWithPrompt(ctx, "aws", "route53", "delete-hosted-zone", "--id", zoneID); err != nil {
		awsLogger.Warningf("Failed to delete hosted zone %s: %v\nstderr: %s", hostedZoneName, err, errOut)
	} else {
		awsLogger.Successf("Deleted hosted zone %s.", hostedZoneName)
	}
}

func discoverKlutchCertificateARN(ctx context.Context, opts DeleteOptions) string {
	awsLogger.Infof("Discovering Klutch ACM certificate (tag %s=%s)...", klutchTagKey, klutchTagValue)
	args := []string{"acm", "list-certificates", "--query", "CertificateSummaryList[].CertificateArn", "--output", "text"}
	if opts.Region != "" {
		args = append(args, "--region", opts.Region)
	}

	out, err := runCmd(ctx, "aws", args...)
	if err != nil {
		awsLogger.Warningf("Could not list ACM certificates: %v\nstderr: %s", err, out)
		return ""
	}

	for arn := range strings.FieldsSeq(out) {
		tagArgs := []string{"acm", "list-tags-for-certificate", "--certificate-arn", arn, "--output", "json"}
		if opts.Region != "" {
			tagArgs = append(tagArgs, "--region", opts.Region)
		}
		tagsOut, tagErr := runCmd(ctx, "aws", tagArgs...)
		if tagErr != nil {
			awsLogger.Warningf("Could not list tags for ACM certificate %s: %v\nstderr: %s", arn, tagErr, tagsOut)
			continue
		}

		var resp struct {
			Tags []struct {
				Key   string `json:"Key"`
				Value string `json:"Value"`
			} `json:"Tags"`
		}
		if err := json.Unmarshal([]byte(tagsOut), &resp); err != nil {
			awsLogger.Warningf("Could not parse tags for ACM certificate %s: %v", arn, err)
			continue
		}
		foundKlutchTag := false
		resourceNameMatches := false
		for _, t := range resp.Tags {
			if t.Key == klutchTagKey && t.Value == klutchTagValue {
				foundKlutchTag = true
			}
			if t.Key == "HostedZoneName" && t.Value == opts.HostedZoneName {
				resourceNameMatches = true
			}
		}
		if foundKlutchTag && resourceNameMatches {
			return arn
		}
	}

	return ""
}

func deleteACMCertificate(ctx context.Context, opts DeleteOptions) {
	arn := strings.TrimSpace(opts.ACMCertificateARN)
	if arn == "" {
		arn = discoverKlutchCertificateARN(ctx, opts)
		if arn == "" {
			awsLogger.Warningf("No ACM certificate ARN provided and no Klutch-tagged certificate found. Skipping ACM deletion.")
			return
		}
		awsLogger.Infof("Using discovered Klutch ACM certificate ARN: %s", arn)
	}

	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete ACM certificate %s.", arn)
		return
	}

	args := []string{"acm", "delete-certificate", "--certificate-arn", arn}
	if opts.Region != "" {
		args = append(args, "--region", opts.Region)
	}
	if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
		awsLogger.Warningf("Failed to delete ACM certificate %s: %v\nstderr: %s", arn, err, errOut)
	} else {
		awsLogger.Successf("Deleted ACM certificate %s.", arn)
	}
}

func dnsAndACMCleanup(ctx context.Context, lbTargets []string, opts DeleteOptions) {
	_ = lbTargets
	if !(opts.IncludeDNSRecords || opts.IncludeHostedZone || opts.IncludeSSLCertificate) {
		return
	}

	awsLogger.Section("DNS and ACM Cleanup")

	if opts.IncludeHostedZone && !opts.IncludeDNSRecords {
		opts.IncludeDNSRecords = true
	}

	var zoneID string
	hostedZoneName := strings.TrimSpace(opts.HostedZoneName)
	if (opts.IncludeDNSRecords || opts.IncludeHostedZone) && hostedZoneName == "" {
		awsLogger.Warningf("Hosted zone name not provided; skipping DNS/hosted zone cleanup.")
		return
	} else if hostedZoneName != "" && (opts.IncludeDNSRecords || opts.IncludeHostedZone) {
		zoneID = findHostedZoneIDByName(ctx, hostedZoneName)
		if zoneID == "" {
			awsLogger.Warningf("Could not find hosted zone %s; skipping DNS/hosted zone cleanup. Ensure you provided --hosted-zone-name.", hostedZoneName)
		}
	}

	if zoneID != "" && opts.IncludeDNSRecords {
		deleteDNSRecords(ctx, zoneID, hostedZoneName, opts)
	}

	if zoneID != "" && opts.IncludeHostedZone {
		deleteHostedZone(ctx, zoneID, hostedZoneName, opts)
	}

	if opts.IncludeSSLCertificate {
		deleteACMCertificate(ctx, opts)
	}
}

func deleteVPC(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Section("Delete VPC")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete VPC %s", vpcID)
		return
	}
	args := []string{"ec2", "delete-vpc", "--vpc-id", vpcID}
	args = appendRegion(args, opts.Region)
	if errOut, err := runCmdWithPrompt(ctx, "aws", args...); err != nil {
		awsLogger.Warningf("Failed to delete VPC %s due to dependencies. Running diagnostics...\nstderr: %s", vpcID, errOut)
		runDiagnostics(ctx, vpcID, opts.Region)
	} else {
		awsLogger.Successf("VPC %s deleted successfully.", vpcID)
	}
}

func runDiagnostics(ctx context.Context, vpcID, region string) {
	type diag struct {
		title string
		args  []string
	}
	diagnostics := []diag{
		{"Remaining Route Tables", []string{"ec2", "describe-route-tables", "--filters", "Name=vpc-id,Values=" + vpcID, "--output", "table"}},
		{"Remaining Subnets", []string{"ec2", "describe-subnets", "--filters", "Name=vpc-id,Values=" + vpcID, "--output", "table"}},
		{"Remaining Network Interfaces", []string{"ec2", "describe-network-interfaces", "--filters", "Name=vpc-id,Values=" + vpcID, "--output", "table"}},
		{"Remaining VPC Endpoints", []string{"ec2", "describe-vpc-endpoints", "--filters", "Name=vpc-id,Values=" + vpcID, "--output", "table"}},
		{"Remaining Transit Gateway VPC Attachments", []string{"ec2", "describe-transit-gateway-vpc-attachments", "--filters", "Name=vpc-id,Values=" + vpcID, "--output", "table"}},
		{"Remaining VPC Peering Connections", []string{"ec2", "describe-vpc-peering-connections", "--filters", "Name=requester-vpc-info.vpc-id,Values=" + vpcID, "Name=accepter-vpc-info.vpc-id,Values=" + vpcID, "--output", "table"}},
		{"Remaining VPN Gateways", []string{"ec2", "describe-vpn-gateways", "--filters", "Name=attachment.vpc-id,Values=" + vpcID, "--output", "table"}},
		{"Remaining NAT Gateways", []string{"ec2", "describe-nat-gateways", "--filter", "Name=vpc-id,Values=" + vpcID, "--output", "table"}},
		{"DHCP Options associated with VPC", []string{"ec2", "describe-vpcs", "--vpc-ids", vpcID, "--query", "Vpcs[].DhcpOptionsId", "--output", "text"}},
	}
	for _, d := range diagnostics {
		awsLogger.Infof("--- Diagnostics: %s ---", d.title)
		args := appendRegion(d.args, region)
		if errOut, err := runCmd(ctx, "aws", args...); err != nil {
			awsLogger.Warningf("Diagnostic %s failed: %v\nstderr: %s", d.title, err, errOut)
		}
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func appendRegion(args []string, region string) []string {
	if strings.TrimSpace(region) == "" {
		return args
	}
	return append(args, "--region", region)
}

// execLookPath is separated for testability.
var execLookPath = func(file string) (string, error) {
	return exec.LookPath(file)
}
