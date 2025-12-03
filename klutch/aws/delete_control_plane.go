package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

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

	awsLogger.Section("Klutch Control Plane Deletion (AWS)")
	if opts.DryRun {
		awsLogger.Infof("Dry-run enabled: no changes will be made. Showing planned actions and resources.")
	}
	awsLogger.Printf("Region:                           %s", cfg.Region)
	awsLogger.Printf("EKS Cluster Name:                 %s", cfg.ClusterName)
	awsLogger.Printf("EKS Nodegroup Name:               %s", cfg.NodegroupName)
	awsLogger.Printf("ALB Controller IAM Policy Name:   %s", cfg.ALBControllerPolicyName)
	awsLogger.Printf("Control Plane Security Group Name:%s", cfg.ControlPlaneSGName)
	awsLogger.Printf("Hosted Zone Name:                 %s", defaultString(opts.HostedZoneName, "<not set>"))
	awsLogger.Printf("Include DNS Records:              %t", opts.IncludeDNSRecords)
	awsLogger.Printf("Include Hosted Zone:              %t", opts.IncludeHostedZone)
	awsLogger.Printf("Include SSL Certificate:          %t", opts.IncludeSSLCertificate)
	awsLogger.Printf("ACM Certificate ARN:              %s", defaultString(opts.ACMCertificateARN, "<auto-discover>"))
	awsLogger.Printf("Dry Run:                          %t", opts.DryRun)
	awsLogger.Printf("Force DNS:                        %t", opts.ForceDNS)

	if !opts.DryRun {
		if !makeup.ConfirmYes("This will delete the Klutch Control Plane on AWS. Type 'yes' to continue: ") {
			makeup.PrintInfo("Deletion aborted.")
			return
		}
	}

	for _, cmd := range []string{"aws", "kubectl", "eksctl", "helm", "jq"} {
		if _, err := execLookPath(cmd); err != nil {
			awsLogger.Fatalf(err, "Required command %q is not installed or not in PATH", cmd)
		}
	}

	accountID, errOut, err := runCmd(ctx, "aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	if err != nil || accountID == "" || accountID == "None" || accountID == "null" {
		awsLogger.Fatalf(err, "Unable to determine AWS Account ID. Run 'aws configure'. stderr: %s", errOut)
	}
	awsLogger.Infof("AWS Account ID: %s", accountID)

	clusterExists, clusterReachable := discoverCluster(ctx, cfg, opts)

	if clusterReachable {
		kubernetesCleanup(ctx, opts)
		iamCleanup(ctx, cfg, opts)
	} else {
		awsLogger.Warningf("Skipping Kubernetes/IAM cleanup (cluster unreachable or missing).")
	}

	if clusterExists {
		deleteNodegroupsAndCluster(ctx, cfg, opts)
	} else {
		awsLogger.Infof("Cluster does not exist. Skipping nodegroup/cluster deletion.")
	}

	vpcID := findKlutchVPC(ctx)
	if vpcID == "" {
		awsLogger.Infof("No Klutch VPC found.")
		dnsAndACMCleanup(ctx, nil, opts)
		return
	}

	deleteVPCDependencies(ctx, vpcID, opts)
	deleteVPC(ctx, vpcID, opts)
}

func discoverCluster(ctx context.Context, cfg Config, opts DeleteOptions) (bool, bool) {
	awsLogger.Section("Discover EKS Cluster")
	clusterExists := false
	clusterReachable := false
	status := "UNKNOWN"

	out, errOut, err := runCmd(ctx, "aws", "eks", "describe-cluster",
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
				if _, errOut, err := runCmd(ctx, "aws", "eks", "update-kubeconfig",
					"--name", cfg.ClusterName,
					"--region", cfg.Region); err != nil {
					awsLogger.Warningf("Could not update kubeconfig: %v\nstderr: %s", err, errOut)
				} else if _, _, err := runCmd(ctx, "kubectl", "version", "--request-timeout=5s"); err == nil {
					clusterReachable = true
				}
			}
		}
	} else if err != nil && !strings.Contains(errOut, "ResourceNotFoundException") {
		awsLogger.Warningf("describe-cluster failed: %v\nstderr: %s", err, errOut)
	}

	return clusterExists, clusterReachable
}

func kubernetesCleanup(ctx context.Context, opts DeleteOptions) {
	awsLogger.Section("Kubernetes Cleanup")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete storageclass gp3 (if present).")
	} else {
		if _, errOut, err := runCmd(ctx, "kubectl", "delete", "storageclass", "gp3", "--ignore-not-found"); err != nil {
			awsLogger.Warningf("Failed to delete storageclass gp3: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Deleted storageclass gp3 (if present).")
		}
	}

	if opts.DryRun {
		awsLogger.Infof("Dry-run: would uninstall AWS LB Controller Helm release (if present).")
	} else {
		if _, errOut, err := runCmd(ctx, "helm", "uninstall", "aws-load-balancer-controller", "-n", "kube-system"); err != nil {
			awsLogger.Warningf("Failed to uninstall AWS LB Controller: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Uninstalled AWS LB Controller (if present).")
		}
	}
}

func iamCleanup(ctx context.Context, cfg Config, opts DeleteOptions) {
	awsLogger.Section("IAM Cleanup")
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete IAM service account aws-load-balancer-controller.")
	} else {
		if _, errOut, err := runCmd(ctx, "eksctl", "delete", "iamserviceaccount",
			"--cluster", cfg.ClusterName,
			"--region", cfg.Region,
			"--namespace", "kube-system",
			"--name", "aws-load-balancer-controller",
			"--wait"); err != nil {
			awsLogger.Warningf("Failed to delete IAM service account: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Deleted IAM service account (if present).")
		}
	}

	if opts.DryRun {
		awsLogger.Infof("Dry-run: would disassociate IAM OIDC provider.")
	} else {
		if _, errOut, err := runCmd(ctx, "eksctl", "utils", "disassociate-iam-oidc-provider",
			"--cluster", cfg.ClusterName,
			"--region", cfg.Region,
			"--approve"); err != nil {
			awsLogger.Warningf("Failed to disassociate OIDC provider: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Disassociated OIDC provider (if present).")
		}
	}

	policyArn, _, err := runCmd(ctx, "aws", "iam", "list-policies", "--scope", "Local",
		"--query", fmt.Sprintf("Policies[?PolicyName=='%s'].Arn | [0]", cfg.ALBControllerPolicyName),
		"--output", "text")
	if err != nil || policyArn == "" || policyArn == "None" {
		awsLogger.Infof("No ALB IAM policy found.")
		return
	}

	versions, _, _ := runCmd(ctx, "aws", "iam", "list-policy-versions",
		"--policy-arn", policyArn,
		"--query", "Versions[?IsDefaultVersion==`false`].VersionId",
		"--output", "text")
	for _, v := range strings.Fields(versions) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete policy version %s for %s", v, cfg.ALBControllerPolicyName)
			continue
		}
		if _, errOut, err := runCmd(ctx, "aws", "iam", "delete-policy-version",
			"--policy-arn", policyArn, "--version-id", v); err != nil {
			awsLogger.Warningf("Failed to delete policy version %s: %v\nstderr: %s", v, err, errOut)
		}
	}
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would delete ALB IAM policy %s", cfg.ALBControllerPolicyName)
	} else {
		if _, errOut, err := runCmd(ctx, "aws", "iam", "delete-policy", "--policy-arn", policyArn); err != nil {
			awsLogger.Warningf("Failed to delete ALB IAM policy: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Deleted ALB IAM policy.")
		}
	}
}

func deleteNodegroupsAndCluster(ctx context.Context, cfg Config, opts DeleteOptions) {
	awsLogger.Section("Delete Nodegroups and Cluster")

	ngs, _, _ := runCmd(ctx, "aws", "eks", "list-nodegroups",
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
		if _, errOut, err := runCmd(ctx, "aws", "eks", "delete-nodegroup",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", ng,
			"--region", cfg.Region); err != nil {
			awsLogger.Warningf("Failed to request deletion for nodegroup %s: %v\nstderr: %s", ng, err, errOut)
		}
		if _, errOut, err := runCmd(ctx, "aws", "eks", "wait", "nodegroup-deleted",
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
		if _, errOut, err := runCmd(ctx, "aws", "eks", "delete-cluster",
			"--name", cfg.ClusterName,
			"--region", cfg.Region); err != nil {
			awsLogger.Warningf("Failed to request cluster deletion: %v\nstderr: %s", err, errOut)
		}
		if _, errOut, err := runCmd(ctx, "aws", "eks", "wait", "cluster-deleted",
			"--name", cfg.ClusterName,
			"--region", cfg.Region); err != nil {
			awsLogger.Warningf("Wait for cluster deletion failed: %v\nstderr: %s", err, errOut)
		} else {
			awsLogger.Successf("Cluster deleted.")
		}
	}
}

func findKlutchVPC(ctx context.Context) string {
	awsLogger.Section("Discover Klutch VPC")
	vpcID, errOut, err := runCmd(ctx, "aws", "ec2", "describe-vpcs",
		"--filters", "Name=tag:Klutch,Values=ControlPlane",
		"--query", "Vpcs[0].VpcId", "--output", "text")
	if err != nil || vpcID == "" || vpcID == "None" {
		awsLogger.Infof("No Klutch VPC found (stderr: %s)", errOut)
		return ""
	}
	awsLogger.Infof("Klutch VPC: %s", vpcID)
	return vpcID
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
	eps, _, _ := runCmd(ctx, "aws", "ec2", "describe-vpc-endpoints",
		"--filters", "Name=vpc-id,Values="+vpcID,
		"--query", "VpcEndpoints[].VpcEndpointId",
		"--output", "text")
	for _, ep := range strings.Fields(eps) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete VPC endpoint %s", ep)
			continue
		}
		_, errOut, err := runCmd(ctx, "aws", "ec2", "delete-vpc-endpoints", "--vpc-endpoint-ids", ep)
		if err != nil {
			awsLogger.Warningf("Failed to delete VPC endpoint %s: %v\nstderr: %s", ep, err, errOut)
		}
	}
}

func deleteLoadBalancers(ctx context.Context, vpcID string, opts DeleteOptions) []string {
	awsLogger.Infof("Deleting Load Balancers in VPC...")
	var lbTargets []string
	lbs, _, _ := runCmd(ctx, "aws", "elbv2", "describe-load-balancers",
		"--query", "LoadBalancers[?VpcId==`"+vpcID+"`].LoadBalancerArn",
		"--output", "text")
	for _, lb := range strings.Fields(lbs) {
		awsLogger.Infof("  LoadBalancer: %s", lb)

		lbDNS, _, _ := runCmd(ctx, "aws", "elbv2", "describe-load-balancers",
			"--load-balancer-arns", lb,
			"--query", "LoadBalancers[0].DNSName",
			"--output", "text")
		if lbDNS != "" && lbDNS != "None" {
			lbTargets = append(lbTargets, lbDNS)
		}

		tgs, _, _ := runCmd(ctx, "aws", "elbv2", "describe-target-groups",
			"--load-balancer-arn", lb,
			"--query", "TargetGroups[].TargetGroupArn",
			"--output", "text")

		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete load balancer %s and target groups %s", lb, strings.TrimSpace(tgs))
		} else {
			if _, errOut, err := runCmd(ctx, "aws", "elbv2", "delete-load-balancer", "--load-balancer-arn", lb); err != nil {
				awsLogger.Warningf("Failed to delete load balancer %s: %v\nstderr: %s", lb, err, errOut)
			}

			time.Sleep(20 * time.Second) // allow ENIs to detach

			for _, tg := range strings.Fields(tgs) {
				if _, errOut, err := runCmd(ctx, "aws", "elbv2", "delete-target-group", "--target-group-arn", tg); err != nil {
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
	for i := 0; i < 10; i++ {
		enis, _, _ := runCmd(ctx, "aws", "ec2", "describe-network-interfaces",
			"--filters", "Name=vpc-id,Values="+vpcID,
			"--query", "NetworkInterfaces[?starts_with(Description, 'ELB ')].NetworkInterfaceId",
			"--output", "text")
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
	natEIPs, _, _ := runCmd(ctx, "aws", "ec2", "describe-nat-gateways",
		"--filter", "Name=vpc-id,Values="+vpcID,
		"--query", "NatGateways[].NatGatewayAddresses[].AllocationId",
		"--output", "text")

	ngws, _, _ := runCmd(ctx, "aws", "ec2", "describe-nat-gateways",
		"--filter", "Name=vpc-id,Values="+vpcID,
		"--query", "NatGateways[].NatGatewayId",
		"--output", "text")

	for _, ng := range strings.Fields(ngws) {
		awsLogger.Infof("Requesting deletion of NAT Gateway %s...", ng)
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete NAT Gateway %s", ng)
			continue
		}
		if _, errOut, err := runCmd(ctx, "aws", "ec2", "delete-nat-gateway", "--nat-gateway-id", ng); err != nil {
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
			state, _, _ := runCmd(ctx, "aws", "ec2", "describe-nat-gateways",
				"--nat-gateway-ids", ng,
				"--query", "NatGateways[0].State",
				"--output", "text")
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
	igwID, _, _ := runCmd(ctx, "aws", "ec2", "describe-internet-gateways",
		"--filters", "Name=attachment.vpc-id,Values="+vpcID,
		"--query", "InternetGateways[0].InternetGatewayId",
		"--output", "text")
	if igwID == "" || igwID == "None" {
		return
	}
	if opts.DryRun {
		awsLogger.Infof("Dry-run: would detach and delete IGW %s", igwID)
		return
	}
	if _, errOut, err := runCmd(ctx, "aws", "ec2", "detach-internet-gateway",
		"--internet-gateway-id", igwID,
		"--vpc-id", vpcID); err != nil {
		awsLogger.Warningf("Failed to detach IGW %s: %v\nstderr: %s", igwID, err, errOut)
	}
	if _, errOut, err := runCmd(ctx, "aws", "ec2", "delete-internet-gateway",
		"--internet-gateway-id", igwID); err != nil {
		awsLogger.Warningf("Failed to delete IGW %s: %v\nstderr: %s", igwID, err, errOut)
	}
}

func deleteRouteTables(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting Route Tables...")
	nonMainRTs, _, _ := runCmd(ctx, "aws", "ec2", "describe-route-tables",
		"--filters", "Name=vpc-id,Values="+vpcID,
		"--query", "RouteTables[?!(Associations[?Main==`true`])].RouteTableId",
		"--output", "text")

	for _, rt := range strings.Fields(nonMainRTs) {
		awsLogger.Infof("Processing non-main route table %s...", rt)
		assocs, _, _ := runCmd(ctx, "aws", "ec2", "describe-route-tables",
			"--route-table-ids", rt,
			"--query", "RouteTables[0].Associations[].RouteTableAssociationId",
			"--output", "text")
		for _, assoc := range strings.Fields(assocs) {
			if opts.DryRun {
				awsLogger.Infof("Dry-run: would disassociate route table %s association %s", rt, assoc)
				continue
			}
			if _, errOut, err := runCmd(ctx, "aws", "ec2", "disassociate-route-table", "--association-id", assoc); err != nil {
				awsLogger.Warningf("Failed to disassociate route table %s: %v\nstderr: %s", assoc, err, errOut)
			}
		}
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete route table %s", rt)
		} else {
			if _, errOut, err := runCmd(ctx, "aws", "ec2", "delete-route-table", "--route-table-id", rt); err != nil {
				awsLogger.Warningf("Failed to delete route table %s: %v\nstderr: %s", rt, err, errOut)
			}
		}
	}
}

func deleteENIs(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting ENIs...")
	enis, _, _ := runCmd(ctx, "aws", "ec2", "describe-network-interfaces",
		"--filters", "Name=vpc-id,Values="+vpcID,
		"--query", "NetworkInterfaces[].NetworkInterfaceId",
		"--output", "text")
	for _, eni := range strings.Fields(enis) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete ENI %s", eni)
		} else {
			if _, errOut, err := runCmd(ctx, "aws", "ec2", "delete-network-interface", "--network-interface-id", eni); err != nil {
				awsLogger.Warningf("Failed to delete ENI %s: %v\nstderr: %s", eni, err, errOut)
			}
		}
	}
}

func deleteSubnets(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting Subnets...")
	subnets, _, _ := runCmd(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID,
		"--query", "Subnets[].SubnetId",
		"--output", "text")
	for _, sn := range strings.Fields(subnets) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete subnet %s", sn)
		} else {
			if _, errOut, err := runCmd(ctx, "aws", "ec2", "delete-subnet", "--subnet-id", sn); err != nil {
				awsLogger.Warningf("Failed to delete subnet %s: %v\nstderr: %s", sn, err, errOut)
			}
		}
	}
}

func deleteSecurityGroups(ctx context.Context, vpcID string, opts DeleteOptions) {
	awsLogger.Infof("Deleting non-default Security Groups in VPC...")
	sgs, _, _ := runCmd(ctx, "aws", "ec2", "describe-security-groups",
		"--filters", "Name=vpc-id,Values="+vpcID,
		"--query", "SecurityGroups[?GroupName!=`default`].GroupId",
		"--output", "text")
	if strings.TrimSpace(sgs) == "" {
		awsLogger.Infof("No non-default Security Groups found to delete.")
		return
	}
	for _, sg := range strings.Fields(sgs) {
		if opts.DryRun {
			awsLogger.Infof("Dry-run: would delete security group %s", sg)
		} else {
			if _, errOut, err := runCmd(ctx, "aws", "ec2", "delete-security-group", "--group-id", sg); err != nil {
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
	if _, errOut, err := runCmd(ctx, "aws", "ec2", "associate-dhcp-options",
		"--dhcp-options-id", "default",
		"--vpc-id", vpcID); err != nil {
		awsLogger.Warningf("Failed to reset DHCP options: %v\nstderr: %s", err, errOut)
	}
}

func releaseEIPs(ctx context.Context, natEIPs []string, opts DeleteOptions) {
	if len(natEIPs) == 0 {
		awsLogger.Infof("No NAT EIP AllocationIds found to release.")
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
		if _, errOut, err := runCmd(ctx, "aws", "ec2", "release-address", "--allocation-id", alloc); err != nil {
			awsLogger.Warningf("Failed to release EIP %s: %v\nstderr: %s", alloc, err, errOut)
		}
	}
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
	out, errOut, err := runCmd(ctx, "aws", "route53", "list-hosted-zones-by-name",
		"--dns-name", normalized,
		"--query", query,
		"--output", "text")
	if err != nil {
		awsLogger.Warningf("Failed to list hosted zones for %s: %v\nstderr: %s", hostedZoneName, err, errOut)
		return ""
	}
	if out == "" || out == "None" || out == "null" {
		awsLogger.Warningf("Hosted zone %s not found via Route53 list-hosted-zones-by-name.", hostedZoneName)
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(out), "/hostedzone/")
}

func listHostedZoneRecords(ctx context.Context, zoneID string) ([]map[string]interface{}, error) {
	out, errOut, err := runCmd(ctx, "aws", "route53", "list-resource-record-sets",
		"--hosted-zone-id", zoneID,
		"--query", "ResourceRecordSets[?Type!=`NS` && Type!=`SOA`]",
		"--output", "json")
	if err != nil {
		return nil, fmt.Errorf("listing resource record sets: %v (stderr: %s)", err, errOut)
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

	if _, errOut, err := runCmd(ctx, "aws", "route53", "change-resource-record-sets",
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
	if _, errOut, err := runCmd(ctx, "aws", "route53", "delete-hosted-zone", "--id", zoneID); err != nil {
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

	listOut, errOut, err := runCmd(ctx, "aws", args...)
	if err != nil {
		awsLogger.Warningf("Could not list ACM certificates: %v\nstderr: %s", err, errOut)
		return ""
	}

	for _, arn := range strings.Fields(listOut) {
		tagArgs := []string{"acm", "list-tags-for-certificate", "--certificate-arn", arn, "--output", "json"}
		if opts.Region != "" {
			tagArgs = append(tagArgs, "--region", opts.Region)
		}
		tagsOut, tagErrOut, tagErr := runCmd(ctx, "aws", tagArgs...)
		if tagErr != nil {
			awsLogger.Warningf("Could not list tags for ACM certificate %s: %v\nstderr: %s", arn, tagErr, tagErrOut)
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
		for _, t := range resp.Tags {
			if t.Key == klutchTagKey && t.Value == klutchTagValue {
				return arn
			}
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
	if _, errOut, err := runCmd(ctx, "aws", args...); err != nil {
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
	if _, errOut, err := runCmd(ctx, "aws", "ec2", "delete-vpc", "--vpc-id", vpcID); err != nil {
		awsLogger.Warningf("Failed to delete VPC %s due to dependencies. Running diagnostics...\nstderr: %s", vpcID, errOut)
		runDiagnostics(ctx, vpcID)
	} else {
		awsLogger.Successf("VPC %s deleted successfully.", vpcID)
	}
}

func runDiagnostics(ctx context.Context, vpcID string) {
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
		if _, errOut, err := runCmd(ctx, "aws", d.args...); err != nil {
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

// execLookPath is separated for testability.
var execLookPath = func(file string) (string, error) {
	return exec.LookPath(file)
}
