package aws

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
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

const (
	klutchTagKey     = "Klutch"
	klutchTagValue   = "ControlPlane"
	klutchNamePrefix = "klutch-control-plane"
)

func resourceName(parts ...string) string {
	return fmt.Sprintf("%s-%s", klutchNamePrefix, strings.Join(parts, "-"))
}

func tagEC2Resource(ctx context.Context, resourceID, name string) {
	mustRun(ctx, "aws", "ec2", "create-tags",
		"--resources", resourceID,
		"--tags",
		fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Key=Name,Value=%s", name))
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func defaultConfig() Config {
	return Config{
		Region:                  getenv("CONTROL_PLANE_CLUSTER_REGION", "eu-central-1"),
		ClusterName:             getenv("CONTROL_PLANE_CLUSTER_NAME", "klutch-control-plane"),
		NodegroupName:           getenv("CONTROL_PLANE_CLUSTER_NODEGROUP_NAME", "klutch-control-plane-nodegroup"),
		NodeInstanceTypes:       getenv("CONTROL_PLANE_CLUSTER_NODEGROUP_INSTANCE_TYPES", "t3a.xlarge"),
		NodeScalingConfig:       getenv("CONTROL_PLANE_CLUSTER_NODEGROUP_SCALING_CONFIG", "minSize=3,maxSize=5,desiredSize=3"),
		NodeAMIType:             getenv("CONTROL_PLANE_CLUSTER_NODEGROUP_AMI_TYPE", "AL2023_x86_64_STANDARD"),
		ClusterRoleName:         getenv("EKS_CLUSTER_ROLE_NAME", "EKSClusterRole"),
		NodeRoleName:            getenv("EKS_NODE_ROLE_NAME", "EKSNodeInstanceRole"),
		BaseCIDR:                getenv("BASE_CIDR", "10.0.0.0/16"),
		PubACIDR:                getenv("PUB_A_CIDR", "10.0.1.0/24"),
		PubBCIDR:                getenv("PUB_B_CIDR", "10.0.2.0/24"),
		PubCCIDR:                getenv("PUB_C_CIDR", "10.0.3.0/24"),
		PrivACIDR:               getenv("PRIV_A_CIDR", "10.0.101.0/24"),
		PrivBCIDR:               getenv("PRIV_B_CIDR", "10.0.102.0/24"),
		PrivCCIDR:               getenv("PRIV_C_CIDR", "10.0.103.0/24"),
		ALBControllerVersion:    getenv("ALB_CONTROLLER_VERSION", "v2.7.1"),
		ALBControllerPolicyName: getenv("ALB_CONTROLLER_POLICY_NAME", "AWSLoadBalancerControllerIAMPolicy"),
		ControlPlaneSGName:      getenv("CONTROL_PLANE_SG_NAME", "klutch-control-plane-sg"),
	}
}

func runCmd(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

func mustRun(ctx context.Context, name string, args ...string) string {
	out, errOut, err := runCmd(ctx, name, args...)
	if err != nil {
		awsLogger.Fatalf(err, "❌ %s %v failed: %v\nstderr: %s", name, args, err, errOut)
	}
	return out
}

func CreateControlPlaneCluster(ctx context.Context) {
	cfg := defaultConfig()

	awsLogger.Successf("✅ Starting 10-install-eks-control-plane-cluster (Go version)")

	for _, cmd := range []string{"aws", "kubectl", "eksctl", "helm"} {
		if _, err := exec.LookPath(cmd); err != nil {
			awsLogger.Fatalf(err, "❌ ERROR: Required command %q is not installed or not in PATH", cmd)
		}
	}
	awsLogger.Successf("✅ All required commands (aws, kubectl, eksctl, helm) are available.")

	awsLogger.Section("Configuration")
	awsLogger.Printf("Region:                           %s", cfg.Region)
	awsLogger.Printf("Cluster Name:                     %s", cfg.ClusterName)
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
	awsLogger.Printf("Control Plane Security Group:     %s", cfg.ControlPlaneSGName)

	awsLogger.Infof("Detecting AWS Account ID...")
	accountID, errOut, err := runCmd(ctx, "aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	if err != nil || accountID == "" || accountID == "None" || accountID == "null" {
		awsLogger.Fatalf(err, "❌ ERROR: Unable to determine AWS Account ID. Run 'aws configure'. stderr: %s", errOut)
	}
	awsLogger.Infof("ACCOUNT_ID: %s", accountID)

	awsLogger.Infof("Checking if cluster '%s' already exists...", cfg.ClusterName)
	clusterStatus := "NONE"
	if out, errOut, err := runCmd(ctx, "aws", "eks", "describe-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "cluster.status",
		"--output", "text"); err == nil {
		clusterStatus = out
	} else if !strings.Contains(errOut, "ResourceNotFoundException") {
		awsLogger.Fatalf(err, "❌ ERROR: aws eks describe-cluster failed\nstderr: %s", errOut)
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
		awsLogger.Fatalf(nil, "❌ ERROR: Cluster exists but is in bad state: %s. Delete or fix manually.", clusterStatus)
	}

	awsLogger.Infof("Checking for existing Klutch VPC...")
	vpcID := ""
	if out, _, err := runCmd(ctx, "aws", "ec2", "describe-vpcs",
		"--filters", "Name=tag:Klutch,Values=ControlPlane",
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
		tagEC2Resource(ctx, vpcID, resourceName("vpc"))
		awsLogger.Successf("Created VPC: %s", vpcID)
	}

	ensureClusterRole(ctx, cfg.ClusterRoleName)
	ensureNodeRole(ctx, cfg.NodeRoleName)

	keyArn := ensureKMSKey(ctx, cfg.Region, accountID, cfg.ClusterRoleName)

	ensureNetworking(ctx, cfg, vpcID)

	if !clusterExists {
		createCluster(ctx, cfg, vpcID, keyArn, accountID)
	} else {
		awsLogger.Infof("EKS cluster already exists. Skipping creation.")
	}
	mustRun(ctx, "aws", "eks", "describe-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "cluster.status")

	ensureDefaultEBSEncryption(ctx)

	ensureNodegroup(ctx, cfg, vpcID, accountID)

	mustRun(ctx, "aws", "eks", "update-kubeconfig",
		"--region", cfg.Region,
		"--name", cfg.ClusterName)
	waitForNodesReady(ctx)

	ensureGp3StorageClass(ctx)

	ensureALBController(ctx, cfg, vpcID, accountID)

	awsLogger.Summaryf("🎉 EKS control plane cluster for Klutch is ready.")
	awsLogger.Printf("   Cluster:   %s", cfg.ClusterName)
	awsLogger.Printf("   Region:    %s", cfg.Region)
	awsLogger.Printf("   VPC:       %s", vpcID)
	awsLogger.Printf("   KMS Key:   %s", keyArn)
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
	mustRun(ctx, "aws", "iam", "create-role",
		"--role-name", roleName,
		"--assume-role-policy-document", "file://"+tmp,
		"--description", "Allows the Klutch control plane EKS cluster to manage AWS resources on its behalf",
		"--tags",
		fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Key=Name,Value=%s", roleName))
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
	mustRun(ctx, "aws", "iam", "create-role",
		"--role-name", roleName,
		"--assume-role-policy-document", "file://"+tmp,
		"--description", "Provides Klutch control plane worker nodes the permissions required to integrate with AWS",
		"--tags",
		fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
		fmt.Sprintf("Key=Name,Value=%s", roleName))
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

func ensureKMSKey(ctx context.Context, region, accountID, clusterRole string) string {
	keyID := os.Getenv("KEY_ID")
	if keyID == "" {
		awsLogger.Infof("Creating new KMS key for EKS secret encryption...")
		keyID = mustRun(ctx, "aws", "kms", "create-key",
			"--description", "Encrypts secret data stored by the Klutch control plane EKS cluster",
			"--query", "KeyMetadata.KeyId",
			"--output", "text",
			"--tags",
			fmt.Sprintf("TagKey=%s,TagValue=%s", klutchTagKey, klutchTagValue),
			fmt.Sprintf("TagKey=Name,TagValue=%s", resourceName("kms-key")))
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

func ensureNetworking(ctx context.Context, cfg Config, vpcID string) {
	awsLogger.Section("Networking")
	awsLogger.Infof("Ensuring networking components (VPC attributes, IGW, subnets, routes, NAT, SG)...")

	mustRun(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-support", "{\"Value\":true}")
	mustRun(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-hostnames", "{\"Value\":true}")

	igwID := ensureIGW(ctx, vpcID)

	pubA := ensureSubnet(ctx, vpcID, cfg.PubACIDR, cfg.Region+"a", resourceName("public-subnet-a"))
	pubB := ensureSubnet(ctx, vpcID, cfg.PubBCIDR, cfg.Region+"b", resourceName("public-subnet-b"))
	pubC := ensureSubnet(ctx, vpcID, cfg.PubCCIDR, cfg.Region+"c", resourceName("public-subnet-c"))
	privA := ensureSubnet(ctx, vpcID, cfg.PrivACIDR, cfg.Region+"a", resourceName("private-subnet-a"))
	privB := ensureSubnet(ctx, vpcID, cfg.PrivBCIDR, cfg.Region+"b", resourceName("private-subnet-b"))
	privC := ensureSubnet(ctx, vpcID, cfg.PrivCCIDR, cfg.Region+"c", resourceName("private-subnet-c"))

	awsLogger.Printf("PUBLIC SUBNETS:  %s, %s, %s", pubA, pubB, pubC)
	awsLogger.Printf("PRIVATE SUBNETS: %s, %s, %s", privA, privB, privC)

	publicRT := ensurePublicRouteTable(ctx, vpcID, igwID, []string{pubA, pubB, pubC})

	ensureElasticIPQuota(ctx)
	natA, natB, natC := ensureNATs(ctx, vpcID, pubA, pubB, pubC)

	privRTA := ensurePrivateRT(ctx, vpcID, privA, natA, resourceName("private-route-table-a"))
	privRTB := ensurePrivateRT(ctx, vpcID, privB, natB, resourceName("private-route-table-b"))
	privRTC := ensurePrivateRT(ctx, vpcID, privC, natC, resourceName("private-route-table-c"))
	_ = publicRT
	awsLogger.Printf("PRIVATE ROUTE TABLES: %s, %s, %s", privRTA, privRTB, privRTC)

	ensureSecurityGroup(ctx, vpcID, cfg.ControlPlaneSGName)
}

func ensureIGW(ctx context.Context, vpcID string) string {
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
		tagEC2Resource(ctx, igwID, resourceName("internet-gateway"))
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

func ensurePublicRouteTable(ctx context.Context, vpcID, igwID string, pubSubnets []string) string {
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
		tagEC2Resource(ctx, rtID, resourceName("public-route-table"))
		mustRun(ctx, "aws", "ec2", "create-route",
			"--route-table-id", rtID,
			"--destination-cidr-block", "0.0.0.0/0",
			"--gateway-id", igwID)
	} else {
		awsLogger.Successf("Reusing public route table: %s", rtID)
	}
	for _, sn := range pubSubnets {
		_, _, _ = runCmd(ctx, "aws", "ec2", "associate-route-table",
			"--route-table-id", rtID,
			"--subnet-id", sn)
	}
	return rtID
}

func ensureElasticIPQuota(ctx context.Context) {
	awsLogger.Infof("Checking Elastic IP quota...")
	out, errOut, err := runCmd(ctx, "aws", "ec2", "describe-addresses",
		"--query", "Addresses[].AllocationId",
		"--output", "text")
	currentCount := 0
	if err == nil && strings.TrimSpace(out) != "" {
		parts := strings.Fields(out)
		currentCount = len(parts)
	} else if err != nil && !strings.Contains(errOut, "AuthFailure") {
		awsLogger.Warningf("describe-addresses failed: %v, stderr: %s", err, errOut)
	}
	required := 3
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
			awsLogger.Fatalf(nil, "❌ ERROR: Not enough Elastic IP quota. Have %d, quota %d, need %d.",
				currentCount, qInt, currentCount+required)
		}
	}
	awsLogger.Successf("✅ Elastic IP quota is sufficient.")
}

func ensureNATs(ctx context.Context, vpcID, pubA, pubB, pubC string) (string, string, string) {
	awsLogger.Infof("Ensuring NAT Gateways exist...")
	var newNATs []string
	ensure := func(subnet, label string) string {
		natID, _, _ := runCmd(ctx, "aws", "ec2", "describe-nat-gateways",
			"--filter",
			"Name=vpc-id,Values="+vpcID,
			"Name=subnet-id,Values="+subnet,
			"Name=state,Values=available",
			"--query", "NatGateways[0].NatGatewayId",
			"--output", "text")
		if natID == "" || natID == "None" || natID == "null" {
			awsLogger.Infof("Creating NAT Gateway in subnet %s...", subnet)
			alloc := mustRun(ctx, "aws", "ec2", "allocate-address",
				"--domain", "vpc",
				"--query", "AllocationId",
				"--output", "text")
			tagEC2Resource(ctx, alloc, resourceName("nat-eip", label))
			natID = mustRun(ctx, "aws", "ec2", "create-nat-gateway",
				"--subnet-id", subnet,
				"--allocation-id", alloc,
				"--query", "NatGateway.NatGatewayId",
				"--output", "text")
			tagEC2Resource(ctx, natID, resourceName("nat-gateway", label))
			newNATs = append(newNATs, natID)
		} else {
			awsLogger.Successf("Reusing NAT Gateway: %s", natID)
		}
		return natID
	}
	natA := ensure(pubA, "a")
	natB := ensure(pubB, "b")
	natC := ensure(pubC, "c")

	if len(newNATs) > 0 {
		args := []string{"ec2", "wait", "nat-gateway-available", "--nat-gateway-ids"}
		args = append(args, newNATs...)
		mustRun(ctx, "aws", args...)
		awsLogger.Successf("New NAT Gateways are available.")
	}
	awsLogger.Printf("NAT Gateways: %s, %s, %s", natA, natB, natC)
	return natA, natB, natC
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
		mustRun(ctx, "aws", "ec2", "associate-route-table",
			"--route-table-id", rtID,
			"--subnet-id", privSubnet)
	} else {
		awsLogger.Successf("Reusing private route table: %s", rtID)
	}
	return rtID
}

func ensureSecurityGroup(ctx context.Context, vpcID, sgName string) string {
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
			"--description", "Restricts traffic for Klutch control plane components and worker nodes",
			"--vpc-id", vpcID,
			"--query", "GroupId",
			"--output", "text")
		tagEC2Resource(ctx, sgID, resourceName("security-group"))
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

	awsLogger.Printf("CONTROL_PLANE_SG_ID = %s", sgID)
	return sgID
}

func createCluster(ctx context.Context, cfg Config, vpcID, keyArn, accountID string) {
	privA := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+getenv("PRIV_A_CIDR", "10.0.101.0/24"),
		"--query", "Subnets[0].SubnetId", "--output", "text")
	privB := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+getenv("PRIV_B_CIDR", "10.0.102.0/24"),
		"--query", "Subnets[0].SubnetId", "--output", "text")
	privC := mustRun(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+getenv("PRIV_C_CIDR", "10.0.103.0/24"),
		"--query", "Subnets[0].SubnetId", "--output", "text")
	sgID := mustRun(ctx, "aws", "ec2", "describe-security-groups",
		"--filters",
		"Name=vpc-id,Values="+vpcID,
		"Name=group-name,Values="+getenv("CONTROL_PLANE_SG_NAME", "klutch-control-plane-sg"),
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
		"--tags", fmt.Sprintf("%s=%s,Name=%s", klutchTagKey, klutchTagValue, cfg.ClusterName),
	)
	awsLogger.Infof("Waiting for EKS cluster '%s' to become ACTIVE...", cfg.ClusterName)
	mustRun(ctx, "aws", "eks", "wait", "cluster-active",
		"--name", cfg.ClusterName,
		"--region", cfg.Region)
	awsLogger.Successf("✅ Cluster is ACTIVE.")
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
		awsLogger.Fatalf(err, "❌ ERROR: describe-nodegroup failed\nstderr: %s", errOut)
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
			"--tags", fmt.Sprintf("%s=%s,Name=%s", klutchTagKey, klutchTagValue, cfg.NodegroupName))
		awsLogger.Infof("Waiting for nodegroup '%s' to become ACTIVE...", cfg.NodegroupName)
		mustRun(ctx, "aws", "eks", "wait", "nodegroup-active",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		awsLogger.Successf("✅ Nodegroup is ACTIVE.")
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
		awsLogger.Successf("✅ Nodegroup is ACTIVE.")
	case "DELETING":
		awsLogger.Fatalf(nil, "❌ ERROR: Nodegroup is currently DELETING. Wait for it to finish and re-run the installer.")
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
	for {
		out, _, err := runCmd(ctx, "kubectl", "get", "nodes")
		if err == nil && strings.Contains(out, " Ready") {
			awsLogger.Successf("✅ Nodes are Ready:")
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
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(yaml)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		awsLogger.Fatalf(err, "❌ kubectl apply gp3 SC failed\nstderr: %s", errBuf.String())
	}
	awsLogger.Successf("✅ gp3 StorageClass installed and set as default.")
}

func ensureALBController(ctx context.Context, cfg Config, vpcID, accountID string) {
	awsLogger.Section("AWS Load Balancer Controller")
	awsLogger.Infof("Installing AWS Load Balancer Controller...")

	mustRun(ctx, "eksctl", "utils", "associate-iam-oidc-provider",
		"--region", cfg.Region,
		"--cluster", cfg.ClusterName,
		"--approve")

	policyURL := fmt.Sprintf("https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/%s/docs/install/iam_policy.json", cfg.ALBControllerVersion)
	awsLogger.Printf("Using AWS Load Balancer Controller version: %s", cfg.ALBControllerVersion)
	awsLogger.Printf("IAM policy URL: %s", policyURL)

	_, errOut, err := runCmd(ctx, "curl", "-sSfL", "-o", "aws-load-balancer-controller-policy.json", policyURL)
	if err != nil {
		awsLogger.Fatalf(err, "curl failed\nstderr: %s", errOut)
	}

	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", accountID, cfg.ALBControllerPolicyName)
	if _, _, err := runCmd(ctx, "aws", "iam", "get-policy", "--policy-arn", policyArn); err != nil {
		awsLogger.Infof("Creating IAM policy %s", cfg.ALBControllerPolicyName)
		mustRun(ctx, "aws", "iam", "create-policy",
			"--policy-name", cfg.ALBControllerPolicyName,
			"--policy-document", "file://aws-load-balancer-controller-policy.json",
			"--description", "Allows the Klutch control plane to run the AWS Load Balancer Controller safely",
			"--tags",
			fmt.Sprintf("Key=%s,Value=%s", klutchTagKey, klutchTagValue),
			fmt.Sprintf("Key=Name,Value=%s", cfg.ALBControllerPolicyName))
	} else {
		awsLogger.Successf("IAM policy %s already exists.", cfg.ALBControllerPolicyName)
	}

	awsLogger.Infof("Creating IAM service account for AWS Load Balancer Controller...")
	mustRun(ctx, "eksctl", "create", "iamserviceaccount",
		"--cluster", cfg.ClusterName,
		"--namespace", "kube-system",
		"--name", "aws-load-balancer-controller",
		"--attach-policy-arn", policyArn,
		"--region", cfg.Region,
		"--approve",
		"--override-existing-serviceaccounts")

	awsLogger.Infof("Installing AWS Load Balancer Controller via Helm...")
	_, _, _ = runCmd(ctx, "helm", "repo", "add", "eks", "https://aws.github.io/eks-charts")
	mustRun(ctx, "helm", "repo", "update")

	args := []string{
		"upgrade", "--install", "aws-load-balancer-controller", "eks/aws-load-balancer-controller",
		"-n", "kube-system",
		"--set", "clusterName=" + cfg.ClusterName,
		"--set", "region=" + cfg.Region,
		"--set", "vpcId=" + vpcID,
		"--set", "serviceAccount.create=false",
		"--set", "serviceAccount.name=aws-load-balancer-controller",
	}
	mustRun(ctx, "helm", args...)

	// Attach the policy inline to the role derived from the service account annotation
	// to ensure all required permissions are present even if a managed policy with the same name exists.
	roleName := getALBControllerRoleName(ctx)
	if roleName != "" {
		awsLogger.Infof("Attaching inline IAM policy to role %s to ensure ALB controller permissions are present...", roleName)
		if len(roleName) > 64 {
			awsLogger.Warningf("Role name %s exceeds 64 characters; skipping inline policy attachment.", roleName)
		} else {
			if _, errOut, err := runCmd(ctx, "aws", "iam", "put-role-policy",
				"--role-name", roleName,
				"--policy-name", cfg.ALBControllerPolicyName+"-inline",
				"--policy-document", "file://aws-load-balancer-controller-policy.json"); err != nil {
				awsLogger.Warningf("Failed to attach inline policy to role %s (continued): %v\nstderr: %s", roleName, err, errOut)
			} else {
				awsLogger.Successf("✅ Attached inline policy to role %s.", roleName)
			}
		}
	}

	awsLogger.Infof("Waiting for aws-load-balancer-controller deployment rollout...")
	if _, errOut, err := runCmd(ctx, "kubectl", "rollout", "status",
		"deployment/aws-load-balancer-controller", "-n", "kube-system"); err != nil {
		awsLogger.Fatalf(err, "❌ ALB controller rollout failed\nstderr: %s", errOut)
	}
	awsLogger.Successf("✅ aws-load-balancer-controller deployment is ready.")
}

func getALBControllerRoleName(ctx context.Context) string {
	type sa struct {
		Metadata struct {
			Annotations map[string]string `json:"annotations"`
		} `json:"metadata"`
	}

	out, errOut, err := runCmd(ctx, "kubectl", "get", "sa", "aws-load-balancer-controller", "-n", "kube-system", "-o", "json")
	if err != nil {
		awsLogger.Warningf("Could not fetch service account to derive role name: %v\nstderr: %s", err, errOut)
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
