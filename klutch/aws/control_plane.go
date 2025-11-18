package aws

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
		log.Fatalf("❌ %s %v failed: %v\nstderr: %s", name, args, err, errOut)
	}
	return out
}

func CreateControlPlaneCluster(ctx context.Context) {
	log.SetFlags(0)

	cfg := defaultConfig()

	log.Println("✅ Starting 10-install-eks-control-plane-cluster (Go version)")

	for _, cmd := range []string{"aws", "kubectl", "eksctl", "helm"} {
		if _, err := exec.LookPath(cmd); err != nil {
			log.Fatalf("❌ ERROR: Required command %q is not installed or not in PATH", cmd)
		}
	}
	log.Println("✅ All required commands (aws, kubectl, eksctl, helm) are available.")

	log.Println("===== CONFIG =====")
	log.Printf("Region:                           %s", cfg.Region)
	log.Printf("Cluster Name:                     %s", cfg.ClusterName)
	log.Printf("Nodegroup Name:                   %s", cfg.NodegroupName)
	log.Printf("Node Instance Types:              %s", cfg.NodeInstanceTypes)
	log.Printf("Nodegroup Scaling:                %s", cfg.NodeScalingConfig)
	log.Printf("Cluster Role Name:                %s", cfg.ClusterRoleName)
	log.Printf("Node Role Name:                   %s", cfg.NodeRoleName)
	log.Printf("VPC CIDR:                         %s", cfg.BaseCIDR)
	log.Printf("Public Subnets:                   %s, %s, %s", cfg.PubACIDR, cfg.PubBCIDR, cfg.PubCCIDR)
	log.Printf("Private Subnets:                  %s, %s, %s", cfg.PrivACIDR, cfg.PrivBCIDR, cfg.PrivCCIDR)
	log.Printf("ALB Controller Version:           %s", cfg.ALBControllerVersion)
	log.Printf("ALB Controller Policy Name:       %s", cfg.ALBControllerPolicyName)
	log.Printf("Control Plane Security Group:     %s", cfg.ControlPlaneSGName)
	log.Println("===================")

	log.Println("Detecting AWS Account ID...")
	accountID, errOut, err := runCmd(ctx, "aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	if err != nil || accountID == "" || accountID == "None" || accountID == "null" {
		log.Fatalf("❌ ERROR: Unable to determine AWS Account ID. Run 'aws configure'. stderr: %s", errOut)
	}
	log.Printf("ACCOUNT_ID: %s", accountID)

	log.Printf("Checking if cluster '%s' already exists...", cfg.ClusterName)
	clusterStatus := "NONE"
	if out, errOut, err := runCmd(ctx, "aws", "eks", "describe-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--query", "cluster.status",
		"--output", "text"); err == nil {
		clusterStatus = out
	} else if !strings.Contains(errOut, "ResourceNotFoundException") {
		log.Fatalf("❌ ERROR: aws eks describe-cluster failed: %v\nstderr: %s", err, errOut)
	}
	clusterExists := false
	switch clusterStatus {
	case "NONE":
		log.Println("Cluster does not exist. It will be created.")
	case "ACTIVE":
		log.Println("Cluster already exists and is ACTIVE. Reusing it.")
		clusterExists = true
	case "CREATING":
		log.Println("Cluster exists and is in CREATING state. Waiting until ACTIVE...")
		mustRun(ctx, "aws", "eks", "wait", "cluster-active",
			"--name", cfg.ClusterName, "--region", cfg.Region)
		log.Println("Cluster is now ACTIVE.")
		clusterExists = true
	default:
		log.Fatalf("❌ ERROR: Cluster exists but is in bad state: %s. Delete or fix manually.", clusterStatus)
	}

	log.Println("Checking for existing Klutch VPC...")
	vpcID := ""
	if out, _, err := runCmd(ctx, "aws", "ec2", "describe-vpcs",
		"--filters", "Name=tag:Klutch,Values=ControlPlane",
		"--query", "Vpcs[0].VpcId",
		"--output", "text"); err == nil && out != "None" && out != "null" {
		vpcID = out
		log.Printf("Reusing existing Klutch VPC: %s", vpcID)
	} else {
		log.Println("No existing Klutch VPC found. A new VPC will be created.")
	}

	if vpcID == "" {
		log.Println("Creating VPC...")
		vpcID = mustRun(ctx, "aws", "ec2", "create-vpc",
			"--cidr-block", cfg.BaseCIDR,
			"--query", "Vpc.VpcId", "--output", "text")
		mustRun(ctx, "aws", "ec2", "create-tags",
			"--resources", vpcID,
			"--tags", "Key=Klutch,Value=ControlPlane")
		log.Println("Created VPC:", vpcID)
	}

	ensureClusterRole(ctx, cfg.ClusterRoleName)
	ensureNodeRole(ctx, cfg.NodeRoleName)

	keyArn := ensureKMSKey(ctx, cfg.Region, accountID, cfg.ClusterRoleName)

	ensureNetworking(ctx, cfg, vpcID)

	if !clusterExists {
		createCluster(ctx, cfg, vpcID, keyArn, accountID)
	} else {
		log.Println("EKS cluster already exists. Skipping creation.")
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

	log.Println("🎉 EKS control plane cluster for Klutch is ready.")
	log.Printf("   Cluster:   %s", cfg.ClusterName)
	log.Printf("   Region:    %s", cfg.Region)
	log.Printf("   VPC:       %s", vpcID)
	log.Printf("   KMS Key:   %s", keyArn)
}

func ensureClusterRole(ctx context.Context, roleName string) {
	log.Printf("Ensuring IAM role '%s' exists...", roleName)
	if _, _, err := runCmd(ctx, "aws", "iam", "get-role", "--role-name", roleName); err == nil {
		log.Printf("EKS cluster role '%s' already exists.", roleName)
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
		log.Fatalf("writing trust policy failed: %v", err)
	}
	mustRun(ctx, "aws", "iam", "create-role",
		"--role-name", roleName,
		"--assume-role-policy-document", "file://"+tmp)
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy")
	log.Printf("Created EKS cluster role '%s'.", roleName)
}

func ensureNodeRole(ctx context.Context, roleName string) {
	log.Printf("Ensuring IAM role '%s' exists...", roleName)
	if _, _, err := runCmd(ctx, "aws", "iam", "get-role", "--role-name", roleName); err == nil {
		log.Printf("EKS node role '%s' already exists.", roleName)
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
		log.Fatalf("writing node trust policy failed: %v", err)
	}
	mustRun(ctx, "aws", "iam", "create-role",
		"--role-name", roleName,
		"--assume-role-policy-document", "file://"+tmp)
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy")
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy")
	mustRun(ctx, "aws", "iam", "attach-role-policy",
		"--role-name", roleName,
		"--policy-arn", "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly")
	log.Printf("Created EKS node role '%s'.", roleName)
}

func ensureKMSKey(ctx context.Context, region, accountID, clusterRole string) string {
	keyID := os.Getenv("KEY_ID")
	if keyID == "" {
		log.Println("Creating new KMS key for EKS secret encryption...")
		keyID = mustRun(ctx, "aws", "kms", "create-key",
			"--description", "Klutch Control Plane – EKS Encryption",
			"--query", "KeyMetadata.KeyId",
			"--output", "text")
	} else {
		log.Println("Using existing KMS KEY_ID:", keyID)
	}
	keyArn := fmt.Sprintf("arn:aws:kms:%s:%s:key/%s", region, accountID, keyID)
	log.Println("KMS KEY_ID: ", keyID)
	log.Println("KMS KEY_ARN:", keyArn)

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
		log.Fatalf("writing kms role policy failed: %v", err)
	}
	mustRun(ctx, "aws", "iam", "put-role-policy",
		"--role-name", clusterRole,
		"--policy-name", "EKSClusterKMSAccess",
		"--policy-document", "file://"+tmp)
	return keyArn
}

func ensureNetworking(ctx context.Context, cfg Config, vpcID string) {
	log.Println("Ensuring networking (VPC attributes, IGW, subnets, routes, NAT, SG)...")

	mustRun(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-support", "{\"Value\":true}")
	mustRun(ctx, "aws", "ec2", "modify-vpc-attribute",
		"--vpc-id", vpcID,
		"--enable-dns-hostnames", "{\"Value\":true}")

	igwID := ensureIGW(ctx, vpcID)

	pubA := ensureSubnet(ctx, vpcID, cfg.PubACIDR, cfg.Region+"a")
	pubB := ensureSubnet(ctx, vpcID, cfg.PubBCIDR, cfg.Region+"b")
	pubC := ensureSubnet(ctx, vpcID, cfg.PubCCIDR, cfg.Region+"c")
	privA := ensureSubnet(ctx, vpcID, cfg.PrivACIDR, cfg.Region+"a")
	privB := ensureSubnet(ctx, vpcID, cfg.PrivBCIDR, cfg.Region+"b")
	privC := ensureSubnet(ctx, vpcID, cfg.PrivCCIDR, cfg.Region+"c")

	log.Printf("PUBLIC SUBNETS:  %s, %s, %s", pubA, pubB, pubC)
	log.Printf("PRIVATE SUBNETS: %s, %s, %s", privA, privB, privC)

	publicRT := ensurePublicRouteTable(ctx, vpcID, igwID, []string{pubA, pubB, pubC})

	ensureElasticIPQuota(ctx)
	natA, natB, natC := ensureNATs(ctx, vpcID, pubA, pubB, pubC)

	privRTA := ensurePrivateRT(ctx, vpcID, privA, natA)
	privRTB := ensurePrivateRT(ctx, vpcID, privB, natB)
	privRTC := ensurePrivateRT(ctx, vpcID, privC, natC)
	_ = publicRT
	log.Printf("PRIVATE ROUTE TABLES: %s, %s, %s", privRTA, privRTB, privRTC)

	ensureSecurityGroup(ctx, vpcID, cfg.ControlPlaneSGName)
}

func ensureIGW(ctx context.Context, vpcID string) string {
	log.Printf("Checking for existing Internet Gateway attached to VPC %s...", vpcID)
	igwID, _, _ := runCmd(ctx, "aws", "ec2", "describe-internet-gateways",
		"--filters", "Name=attachment.vpc-id,Values="+vpcID,
		"--query", "InternetGateways[0].InternetGatewayId",
		"--output", "text")
	if igwID == "" || igwID == "None" || igwID == "null" {
		log.Println("No Internet Gateway found. Creating and attaching a new one...")
		igwID = mustRun(ctx, "aws", "ec2", "create-internet-gateway",
			"--query", "InternetGateway.InternetGatewayId",
			"--output", "text")
		mustRun(ctx, "aws", "ec2", "attach-internet-gateway",
			"--internet-gateway-id", igwID,
			"--vpc-id", vpcID)
	} else {
		log.Println("Reusing existing Internet Gateway:", igwID)
	}
	log.Println("IGW_ID =", igwID)
	return igwID
}

func ensureSubnet(ctx context.Context, vpcID, cidr, az string) string {
	out, _, _ := runCmd(ctx, "aws", "ec2", "describe-subnets",
		"--filters", "Name=vpc-id,Values="+vpcID, "Name=cidr-block,Values="+cidr,
		"--query", "Subnets[0].SubnetId",
		"--output", "text")
	if out == "" || out == "None" || out == "null" {
		log.Printf("Creating subnet %s in AZ %s...", cidr, az)
		out = mustRun(ctx, "aws", "ec2", "create-subnet",
			"--vpc-id", vpcID,
			"--cidr-block", cidr,
			"--availability-zone", az,
			"--query", "Subnet.SubnetId",
			"--output", "text")
	} else {
		log.Printf("Reusing subnet %s: %s", cidr, out)
	}
	return out
}

func ensurePublicRouteTable(ctx context.Context, vpcID, igwID string, pubSubnets []string) string {
	log.Println("Checking for existing public route table...")
	rtID, _, _ := runCmd(ctx, "aws", "ec2", "describe-route-tables",
		"--filters", "Name=vpc-id,Values="+vpcID,
		"--query", "RouteTables[?Routes[?GatewayId=='"+igwID+"' && DestinationCidrBlock=='0.0.0.0/0']].RouteTableId | [0]",
		"--output", "text")
	if rtID == "" || rtID == "None" || rtID == "null" {
		log.Println("Creating public route table...")
		rtID = mustRun(ctx, "aws", "ec2", "create-route-table",
			"--vpc-id", vpcID,
			"--query", "RouteTable.RouteTableId",
			"--output", "text")
		mustRun(ctx, "aws", "ec2", "create-route",
			"--route-table-id", rtID,
			"--destination-cidr-block", "0.0.0.0/0",
			"--gateway-id", igwID)
	} else {
		log.Println("Reusing public route table:", rtID)
	}
	for _, sn := range pubSubnets {
		_, _, _ = runCmd(ctx, "aws", "ec2", "associate-route-table",
			"--route-table-id", rtID,
			"--subnet-id", sn)
	}
	return rtID
}

func ensureElasticIPQuota(ctx context.Context) {
	log.Println("Checking Elastic IP quota...")
	out, errOut, err := runCmd(ctx, "aws", "ec2", "describe-addresses",
		"--query", "Addresses[].AllocationId",
		"--output", "text")
	currentCount := 0
	if err == nil && strings.TrimSpace(out) != "" {
		parts := strings.Fields(out)
		currentCount = len(parts)
	} else if err != nil && !strings.Contains(errOut, "AuthFailure") {
		log.Printf("WARN: describe-addresses failed: %v, stderr: %s", err, errOut)
	}
	required := 3
	quotaRaw, errOutQ, errQ := runCmd(ctx, "aws", "service-quotas", "get-service-quota",
		"--service-code", "ec2",
		"--quota-code", "L-0263D0A3",
		"--query", "Quota.Value",
		"--output", "text")
	if errQ != nil {
		log.Printf("WARN: service-quotas not available or failed: %v, stderr: %s", errQ, errOutQ)
		log.Println("Skipping Elastic IP quota enforcement.")
		return
	}
	log.Printf("Current EIPs allocated: %d", currentCount)
	log.Printf("Required EIPs for install: %d", required)
	log.Printf("Elastic IP quota: %s", quotaRaw)
	if quotaRaw != "Unknown" && quotaRaw != "" {
		qFloat, _ := strconv.ParseFloat(quotaRaw, 64)
		qInt := int(qFloat)
		if currentCount+required > qInt {
			log.Fatalf("❌ ERROR: Not enough Elastic IP quota. Have %d, quota %d, need %d.",
				currentCount, qInt, currentCount+required)
		}
	}
	log.Println("✅ Elastic IP quota is sufficient.")
}

func ensureNATs(ctx context.Context, vpcID, pubA, pubB, pubC string) (string, string, string) {
	log.Println("Ensuring NAT Gateways exist...")
	var newNATs []string
	ensure := func(subnet string) string {
		natID, _, _ := runCmd(ctx, "aws", "ec2", "describe-nat-gateways",
			"--filter",
			"Name=vpc-id,Values="+vpcID,
			"Name=subnet-id,Values="+subnet,
			"Name=state,Values=available",
			"--query", "NatGateways[0].NatGatewayId",
			"--output", "text")
		if natID == "" || natID == "None" || natID == "null" {
			log.Printf("Creating NAT Gateway in subnet %s...", subnet)
			alloc := mustRun(ctx, "aws", "ec2", "allocate-address",
				"--domain", "vpc",
				"--query", "AllocationId",
				"--output", "text")
			natID = mustRun(ctx, "aws", "ec2", "create-nat-gateway",
				"--subnet-id", subnet,
				"--allocation-id", alloc,
				"--query", "NatGateway.NatGatewayId",
				"--output", "text")
			newNATs = append(newNATs, natID)
		} else {
			log.Println("Reusing NAT Gateway:", natID)
		}
		return natID
	}
	natA := ensure(pubA)
	natB := ensure(pubB)
	natC := ensure(pubC)

	if len(newNATs) > 0 {
		args := []string{"ec2", "wait", "nat-gateway-available", "--nat-gateway-ids"}
		args = append(args, newNATs...)
		mustRun(ctx, "aws", args...)
		log.Println("New NAT Gateways are available.")
	}
	log.Printf("NAT Gateways: %s, %s, %s", natA, natB, natC)
	return natA, natB, natC
}

func ensurePrivateRT(ctx context.Context, vpcID, privSubnet, natID string) string {
	rtID, _, _ := runCmd(ctx, "aws", "ec2", "describe-route-tables",
		"--filters",
		"Name=vpc-id,Values="+vpcID,
		"Name=association.subnet-id,Values="+privSubnet,
		"--query", "RouteTables[?Routes[?NatGatewayId=='"+natID+"' && DestinationCidrBlock=='0.0.0.0/0']].RouteTableId | [0]",
		"--output", "text")
	if rtID == "" || rtID == "None" || rtID == "null" {
		log.Printf("Creating private route table for subnet %s...", privSubnet)
		rtID = mustRun(ctx, "aws", "ec2", "create-route-table",
			"--vpc-id", vpcID,
			"--query", "RouteTable.RouteTableId",
			"--output", "text")
		mustRun(ctx, "aws", "ec2", "create-route",
			"--route-table-id", rtID,
			"--destination-cidr-block", "0.0.0.0/0",
			"--nat-gateway-id", natID)
		mustRun(ctx, "aws", "ec2", "associate-route-table",
			"--route-table-id", rtID,
			"--subnet-id", privSubnet)
	} else {
		log.Printf("Reusing private route table: %s", rtID)
	}
	return rtID
}

func ensureSecurityGroup(ctx context.Context, vpcID, sgName string) string {
	log.Println("Ensuring security group exists...")
	sgID, _, _ := runCmd(ctx, "aws", "ec2", "describe-security-groups",
		"--filters",
		"Name=vpc-id,Values="+vpcID,
		"Name=group-name,Values="+sgName,
		"--query", "SecurityGroups[0].GroupId",
		"--output", "text")
	if sgID == "" || sgID == "None" || sgID == "null" {
		sgID = mustRun(ctx, "aws", "ec2", "create-security-group",
			"--group-name", sgName,
			"--description", "Klutch Control Plane and node communication",
			"--vpc-id", vpcID,
			"--query", "GroupId",
			"--output", "text")
		log.Println("Created security group:", sgID)
	} else {
		log.Println("Reusing security group:", sgID)
	}
	_, _, _ = runCmd(ctx, "aws", "ec2", "authorize-security-group-ingress",
		"--group-id", sgID,
		"--protocol", "-1",
		"--source-group", sgID)
	_, _, _ = runCmd(ctx, "aws", "ec2", "authorize-security-group-egress",
		"--group-id", sgID,
		"--protocol", "-1",
		"--cidr", "0.0.0.0/0")

	log.Println("CONTROL_PLANE_SG_ID =", sgID)
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
	log.Printf("Creating EKS cluster '%s'...", cfg.ClusterName)
	mustRun(ctx, "aws", "eks", "create-cluster",
		"--name", cfg.ClusterName,
		"--region", cfg.Region,
		"--role-arn", fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, cfg.ClusterRoleName),
		"--resources-vpc-config", fmt.Sprintf("subnetIds=%s,securityGroupIds=%s", subnets, sgID),
		"--encryption-config", fmt.Sprintf("[{\"resources\":[\"secrets\"],\"provider\":{\"keyArn\":\"%s\"}}]", keyArn),
	)
	log.Printf("Waiting for EKS cluster '%s' to become ACTIVE...", cfg.ClusterName)
	mustRun(ctx, "aws", "eks", "wait", "cluster-active",
		"--name", cfg.ClusterName,
		"--region", cfg.Region)
	log.Println("✅ Cluster is ACTIVE.")
}

func ensureDefaultEBSEncryption(ctx context.Context) {
	log.Println("Ensuring AWS account default EBS encryption is enabled...")
	out, _, err := runCmd(ctx, "aws", "ec2", "get-ebs-encryption-by-default",
		"--query", "EbsEncryptionByDefault", "--output", "text")
	if err != nil {
		log.Printf("WARN: get-ebs-encryption-by-default failed, skipping. Out=%s", out)
		return
	}
	if out != "true" {
		log.Println("Default EBS encryption is currently disabled. Enabling now...")
		mustRun(ctx, "aws", "ec2", "enable-ebs-encryption-by-default")
		log.Println("Default EBS encryption has been enabled.")
	} else {
		log.Println("Default EBS encryption is already enabled.")
	}
}

func ensureNodegroup(ctx context.Context, cfg Config, vpcID, accountID string) {
	log.Printf("Checking nodegroup '%s'...", cfg.NodegroupName)
	status, errOut, err := runCmd(ctx, "aws", "eks", "describe-nodegroup",
		"--cluster-name", cfg.ClusterName,
		"--nodegroup-name", cfg.NodegroupName,
		"--region", cfg.Region,
		"--query", "nodegroup.status",
		"--output", "text")
	if err != nil && !strings.Contains(errOut, "ResourceNotFoundException") {
		log.Fatalf("❌ ERROR: describe-nodegroup failed: %v\nstderr: %s", err, errOut)
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
		log.Printf("Creating nodegroup '%s'...", cfg.NodegroupName)
		mustRun(ctx, "aws", "eks", "create-nodegroup",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--scaling-config", cfg.NodeScalingConfig,
			"--instance-types", cfg.NodeInstanceTypes,
			"--subnets", privA, privB, privC,
			"--ami-type", cfg.NodeAMIType,
			"--node-role", fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, cfg.NodeRoleName),
			"--region", cfg.Region)
		log.Printf("Waiting for nodegroup '%s' to become ACTIVE...", cfg.NodegroupName)
		mustRun(ctx, "aws", "eks", "wait", "nodegroup-active",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		log.Println("✅ Nodegroup is ACTIVE.")
	}

	switch status {
	case "NONE":
		log.Println("Nodegroup does not exist. Creating...")
		create()
	case "ACTIVE":
		log.Println("Nodegroup already exists and is ACTIVE. Reusing it.")
	case "CREATING":
		log.Println("Nodegroup is in CREATING. Waiting until ACTIVE...")
		mustRun(ctx, "aws", "eks", "wait", "nodegroup-active",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		log.Println("✅ Nodegroup is ACTIVE.")
	case "DELETING":
		log.Fatal("❌ ERROR: Nodegroup is currently DELETING. Wait for it to finish and re-run the installer.")
	default:
		log.Printf("Nodegroup is in bad state: %s. Deleting and recreating...", status)
		mustRun(ctx, "aws", "eks", "delete-nodegroup",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		log.Printf("Waiting for nodegroup '%s' to be deleted...", cfg.NodegroupName)
		mustRun(ctx, "aws", "eks", "wait", "nodegroup-deleted",
			"--cluster-name", cfg.ClusterName,
			"--nodegroup-name", cfg.NodegroupName,
			"--region", cfg.Region)
		create()
	}
}

func waitForNodesReady(ctx context.Context) {
	log.Println("Waiting for at least one Ready node...")
	for {
		out, _, err := runCmd(ctx, "kubectl", "get", "nodes")
		if err == nil && strings.Contains(out, " Ready") {
			log.Println("✅ Nodes are Ready:")
			fmt.Println(out)
			return
		}
		log.Println("No Ready nodes yet, sleeping 10s...")
		time.Sleep(10 * time.Second)
	}
}

func ensureGp3StorageClass(ctx context.Context) {
	log.Println("Creating gp3 StorageClass and setting it as default...")
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
		log.Fatalf("❌ kubectl apply gp3 SC failed: %v\nstderr: %s", err, errBuf.String())
	}
	log.Println("✅ gp3 StorageClass installed and set as default.")
}

func ensureALBController(ctx context.Context, cfg Config, vpcID, accountID string) {
	log.Println("Installing AWS Load Balancer Controller...")

	mustRun(ctx, "eksctl", "utils", "associate-iam-oidc-provider",
		"--region", cfg.Region,
		"--cluster", cfg.ClusterName,
		"--approve")

	policyURL := fmt.Sprintf("https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/%s/docs/install/iam_policy.json", cfg.ALBControllerVersion)
	log.Println("Using AWS Load Balancer Controller version:", cfg.ALBControllerVersion)
	log.Println("IAM policy URL:", policyURL)

	_, errOut, err := runCmd(ctx, "curl", "-sSfL", "-o", "aws-load-balancer-controller-policy.json", policyURL)
	if err != nil {
		log.Fatalf("curl failed: %v\nstderr: %s", err, errOut)
	}

	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", accountID, cfg.ALBControllerPolicyName)
	if _, _, err := runCmd(ctx, "aws", "iam", "get-policy", "--policy-arn", policyArn); err != nil {
		log.Printf("Creating IAM policy %s", cfg.ALBControllerPolicyName)
		mustRun(ctx, "aws", "iam", "create-policy",
			"--policy-name", cfg.ALBControllerPolicyName,
			"--policy-document", "file://aws-load-balancer-controller-policy.json")
	} else {
		log.Printf("IAM policy %s already exists.", cfg.ALBControllerPolicyName)
	}

	log.Println("Creating IAM service account for AWS Load Balancer Controller...")
	mustRun(ctx, "eksctl", "create", "iamserviceaccount",
		"--cluster", cfg.ClusterName,
		"--namespace", "kube-system",
		"--name", "aws-load-balancer-controller",
		"--attach-policy-arn", policyArn,
		"--region", cfg.Region,
		"--approve",
		"--override-existing-serviceaccounts")

	log.Println("Installing AWS Load Balancer Controller via Helm...")
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

	log.Println("Waiting for aws-load-balancer-controller deployment rollout...")
	if _, errOut, err := runCmd(ctx, "kubectl", "rollout", "status",
		"deployment/aws-load-balancer-controller", "-n", "kube-system"); err != nil {
		log.Fatalf("❌ ALB controller rollout failed: %v\nstderr: %s", err, errOut)
	}
	log.Println("✅ aws-load-balancer-controller deployment is ready.")
}
