package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/klutch"
	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/a9s-cli-v2/pg"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var createClusterKlutchDryRun bool
var createKlutchControlPlane = klutchaws.CreateControlPlaneCluster
var createKlutchWorkload = klutchaws.CreateWorkloadCluster
var createClusterKlutchControlPlaneSkipApply bool
var createKlutchTenantName string
var createKlutchTenantRegion string
var createKlutchTenantStoreSecret bool = true
var createKlutchTenantSecretName string
var createKlutchTenantForce bool
var createKlutchTenantBindRequestFile string
var createKlutchWorkloadAutobindControlPlaneName string

var createKlutchWorkloadTenantName string
var createKlutchWorkloadTenantSecretName string
var createKlutchWorkloadTenantRegion string
var createKlutchWorkloadBindRequestFile string
var createKlutchNodeType string
var createKlutchNodes int
var createKlutchRegion string

// controlPlaneCognitoPoolFromCluster tries to read the control-plane Cognito issuer
// from the in-cluster oidc-config secret and derives the user pool ID (and region).
func controlPlaneCognitoPoolFromCluster() (poolID string, region string) {
	kc := k8s.NewKubeClient("")
	clientset := kc.GetKubernetesClientSet()
	sec, err := clientset.CoreV1().Secrets("default").Get(context.Background(), "oidc-config", metav1.GetOptions{})
	if err != nil {
		return "", ""
	}
	issuer := ""
	if b, ok := sec.Data["oidc-issuer-url"]; ok {
		issuer = string(b)
	}
	if issuer == "" {
		if s, ok := sec.StringData["oidc-issuer-url"]; ok {
			issuer = s
		}
	}
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return "", ""
	}
	u, err := url.Parse(issuer)
	if err != nil {
		return "", ""
	}
	pathSegs := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathSegs) > 0 {
		poolID = pathSegs[len(pathSegs)-1]
	}
	hostParts := strings.Split(u.Host, ".")
	if len(hostParts) >= 2 {
		region = hostParts[1]
	}
	return poolID, region
}

type controlPlaneOIDCCreds struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	TokenURL     string
}

// controlPlaneOIDCFromCluster reads the control-plane oidc-config secret and discovers the token endpoint.
func controlPlaneOIDCFromCluster() controlPlaneOIDCCreds {
	kc := k8s.NewKubeClient("")
	clientset := kc.GetKubernetesClientSet()
	sec, err := clientset.CoreV1().Secrets("default").Get(context.Background(), "oidc-config", metav1.GetOptions{})
	if err != nil {
		return controlPlaneOIDCCreds{}
	}

	getVal := func(key string) string {
		if b, ok := sec.Data[key]; ok {
			return strings.TrimSpace(string(b))
		}
		if sec.StringData != nil {
			if s, ok := sec.StringData[key]; ok {
				return strings.TrimSpace(s)
			}
		}
		return ""
	}

	issuer := getVal("oidc-issuer-url")
	clientID := getVal("oidc-issuer-client-id")
	clientSecret := getVal("oidc-issuer-client-secret")
	tokenURL := discoverTokenEndpoint(issuer)
	return controlPlaneOIDCCreds{
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
	}
}

// discoverTokenEndpoint performs OIDC discovery to find the token endpoint.
func discoverTokenEndpoint(issuer string) string {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return ""
	}
	url := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var doc map[string]interface{}
	if err := json.Unmarshal(body, &doc); err != nil {
		return ""
	}
	if t, ok := doc["token_endpoint"].(string); ok {
		return strings.TrimSpace(t)
	}
	return ""
}

var cmdCreate = &cobra.Command{
	Use:   "create",
	Short: "Create data service resources such as data service instances, service bindings, backups and restore jobs.",
	Long:  `Create data service resources including data service instances, service bindings backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select the data service resource type you would like to instantiate.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var cmdCreatePG = &cobra.Command{
	Use:   "pg",
	Short: "Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.",
	Long:  `Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a PostgreSQL resource such as (service) instance.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var cmdPGInstance = &cobra.Command{
	Use:   "instance",
	Short: "Create a PostgreSQL service instance.",
	Long:  `Create a PostgreSQL service instance`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		a8s.CreatePGServiceInstance()

		if !(demo.DoNotApply) {
			instance := demo.A8sPGServiceInstance
			a8s.WaitForServiceInstanceToBecomeReady(instance.Namespace, instance.Name, instance.Replicas)
		}
	},
}

var cmdPGBackup = &cobra.Command{
	Use:   "backup",
	Short: "Create a PostgreSQL backup of a PostgreSQL service instance.",
	Long:  `Create a PostgreSQL backup of a PostgreSQL service instance`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")

		// DoNotApply is processed in the Create func
		a8s.CreatePGServiceInstanceBackup()
	},
}

var cmdPGRestore = &cobra.Command{
	Use:   "restore",
	Short: "Create a PostgreSQL restore of a PostgreSQL backup.",
	Long:  `Create a PostgreSQL restore of a PostgreSQL backup of a PostgreSQL service instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")

		// DoNotApply is processed in the Create func
		a8s.CreatePGServiceInstanceRestore()
	},
}

var cmdCreateCluster = &cobra.Command{
	Use:   "cluster",
	Short: "Create a local development Kubernetes cluster with a given stack.",
	Long: `Guides through the creation of a local development Kubernetes cluster,
	helps to install all necessary prerequisites and finally configures and installs
	the chosen stack. Select a sub-command to create corresponding stack.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var cmdCreateStack = &cobra.Command{
	Use:   "stack",
	Short: "Applies the specified stack to the currently selected Kubernetes cluster.",
	Long: `Guides through the installation of the given anynines stack to the currently selected Kubernetes cluster,
	helps to install all necessary prerequisites and finally configures and installs
	the chosen stack. Select a sub-command to create corresponding stack.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var cmdCreateClusterA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Create a local development cluster including a8s Data Services such as a8s Postgres.",
	Long:  `Helps with the creation of a local Kubernetes cluster, installing the a8s Data Service operator(s) including necessary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		a8s.CreateA8sStack(true)
	},
}

var cmdCreateClusterKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Create Klutch clusters.",
	Long:  `Create Klutch clusters on the selected provider. Currently only AWS is supported for the control plane.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var cmdCreateKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Create Klutch resources.",
	Long:  `Create Klutch resources such as tenants (Cognito app clients).`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		if err := cmd.Help(); err != nil {
			makeup.ExitDueToFatalError(err, "")
		}
	},
}

var cmdCreateClusterKlutchControlPlane = &cobra.Command{
	Use:   "control-plane",
	Short: "Create the Klutch control plane cluster (and install it).",
	Long: `Creates the Klutch control plane cluster on the selected provider and installs the Klutch control plane components.
Use --no-apply to only provision the cluster. Currently only AWS is supported.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := klutchaws.CreateOptions{DryRun: createClusterKlutchDryRun}
		if cmd.Flags().Changed("cluster-name") {
			options.ClusterName = strings.TrimSpace(demo.DemoClusterName)
		}
		options.Region = strings.TrimSpace(createKlutchRegion)
		options.NodeInstanceTypes = strings.TrimSpace(createKlutchNodeType)
		options.NodeCount = createKlutchNodes
		options.TenantOperatorImage = strings.TrimSpace(sharedKlutchTenantOperatorImage)
		options.TenantOperatorChart = strings.TrimSpace(sharedKlutchTenantOperatorChart)
		options.TenantOperatorChartVersion = strings.TrimSpace(sharedKlutchTenantOperatorChartVersion)
		options.TenantOperatorRoleARN = strings.TrimSpace(sharedKlutchTenantOperatorRoleARN)
		options.TenantOperatorRegion = strings.TrimSpace(sharedKlutchTenantOperatorRegion)
		options.TenantOperatorBindURL = strings.TrimSpace(sharedKlutchTenantOperatorBindURL)
		options.TenantOperatorBindRequest = strings.TrimSpace(sharedKlutchTenantOperatorBindRequest)
		options.HostedZoneName = strings.TrimSpace(sharedKlutchControlPlaneHostedZone)

		if createClusterKlutchDryRun {
			makeup.PrintInfo("Skipping Klutch control plane install because --dry-run was provided.")
			return
		}

		if createClusterKlutchControlPlaneSkipApply {
			makeup.PrintInfo("Skipping Klutch control plane install because --no-apply was provided.")
			return
		}

		if sharedKlutchControlPlaneHostedZone == "" {
			makeup.ExitDueToFatalError(nil, "The --hosted-zone-name flag is required to install the Klutch control plane. Use --no-apply to provision only.")
		}

		if sharedKlutchControlPlanePort < 1 || sharedKlutchControlPlanePort > 65535 {
			makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
		}

		if err := runKlutchClusterCreation(demo.KubernetesTool, options); err != nil {
			makeup.ExitDueToFatalError(nil, err.Error())
		}

		imgURL, imgTag := resolveBackendImageRef(sharedKlutchBackendImageRef, sharedKlutchBackendImageURL, sharedKlutchBackendImageTag)
		klutch.SetBindBackendImage(imgURL, imgTag)

		klutch.SetControlPlaneOIDCOptions(klutch.OIDCOptions{
			Provider:     klutch.OIDCProvider(sharedKlutchOIDCProvider),
			IssuerURL:    sharedKlutchOIDCIssuerURL,
			ClientID:     sharedKlutchOIDCClientID,
			ClientSecret: sharedKlutchOIDCClientSecret,
			CallbackURL:  sharedKlutchOIDCCallbackURL,
		})

		klutch.ApplyKlutchControlPlane(sharedKlutchControlPlaneHost, sharedKlutchControlPlanePort, sharedKlutchControlPlaneACMCertARN, sharedKlutchControlPlaneHostedZone, options.ClusterName)
	},
}

var cmdCreateKlutchTenant = &cobra.Command{
	Use:   "tenant",
	Short: "Create a Klutch tenant (Cognito app client) for workload bindings.",
	Long:  `Creates a Klutch tenant by applying a Tenant CR to the control-plane cluster (reconciled by the tenant operator).`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(createKlutchTenantName) == "" {
			makeup.ExitDueToFatalError(nil, "The --tenant-name flag is required.")
		}

		region := strings.TrimSpace(createKlutchTenantRegion)
		if region == "" {
			region = klutchaws.ControlPlaneDefaultRegion()
		}

		// Build bind request APIs (default or from file).
		var apis []klutch.GroupResource
		if strings.TrimSpace(createKlutchTenantBindRequestFile) != "" {
			data, err := os.ReadFile(createKlutchTenantBindRequestFile)
			if err != nil {
				makeup.ExitDueToFatalError(err, "Failed to read bind request file.")
			}
			var payload struct {
				ClusterID string                 `json:"clusterID"`
				Apis      []klutch.GroupResource `json:"apis"`
			}
			if err := json.Unmarshal(data, &payload); err != nil {
				makeup.ExitDueToFatalError(err, "Bind request file is not valid JSON.")
			}
			apis = payload.Apis
		} else {
			if br, err := klutch.DefaultBindRequestJSON(createKlutchTenantName); err == nil {
				var payload struct {
					ClusterID string                 `json:"clusterID"`
					Apis      []klutch.GroupResource `json:"apis"`
				}
				_ = json.Unmarshal(br, &payload)
				apis = payload.Apis
			}
		}
		if len(apis) == 0 {
			makeup.ExitDueToFatalError(nil, "No APIs specified for the tenant bind request.")
		}

		ns := "a9s-tenants-operator-system"
		tenantUUID := uuid.New().String()
		tenantCR := map[string]interface{}{
			"apiVersion": "klutch.anynines.com/v1alpha1",
			"kind":       "Tenant",
			"metadata": map[string]interface{}{
				"name":      createKlutchTenantName,
				"namespace": ns,
			},
			"spec": map[string]interface{}{
				"displayName": createKlutchTenantName,
				"tenantUUID":  tenantUUID,
				"provider":    "cognito",
				"region":      region,
				"apis":        apis,
			},
		}

		yamlBytes, err := yaml.Marshal(tenantCR)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to render Tenant manifest.")
		}

		// Use the kubectl client to apply the manifest
		k8sClient := k8s.NewKubeClient("")
		if _, err := k8sClient.ApplyWithPrompt(yamlBytes, fmt.Sprintf("Tenant %s", createKlutchTenantName)); err != nil {
			makeup.ExitDueToFatalError(err, "Failed to apply Tenant CR")
		}

		makeup.PrintSuccessSummary(fmt.Sprintf("Tenant %s created via tenant operator. Wait for reconciliation to provision Cognito client and secret.", createKlutchTenantName))
	},
}

var cmdCreateClusterKlutchWorkload = &cobra.Command{
	Use:   "workload",
	Short: "Create a Klutch workload cluster (EKS).",
	Long:  `Creates a Klutch workload cluster on the selected provider. Currently only AWS is supported.`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := klutchaws.CreateOptions{
			DryRun:               createClusterKlutchDryRun,
			Region:               strings.TrimSpace(createKlutchRegion),
			NodeInstanceTypes:    strings.TrimSpace(createKlutchNodeType),
			NodeCount:            createKlutchNodes,
			ControlPlaneToBindTo: createKlutchWorkloadAutobindControlPlaneName,
		}
		var tenantConn *klutchaws.OIDCConnection
		var tenantBindRequest []byte
		var tenantSecretName string
		if cmd.Flags().Changed("cluster-name") && strings.TrimSpace(demo.DemoClusterName) != "" {
			opts.ClusterName = strings.TrimSpace(demo.DemoClusterName)
		} else if envName := strings.TrimSpace(os.Getenv("WORKLOAD_CLUSTER_NAME")); envName != "" {
			opts.ClusterName = envName
		} else {
			opts.ClusterName = klutchaws.RandomWorkloadClusterName()
			makeup.PrintInfo(fmt.Sprintf("Generated workload cluster name: %s", opts.ClusterName))
		}

		// Pre-validate tenant and bind request to fail fast before provisioning.
		if strings.TrimSpace(createKlutchWorkloadTenantName) != "" || strings.TrimSpace(createKlutchWorkloadTenantSecretName) != "" {
			region := strings.TrimSpace(createKlutchWorkloadTenantRegion)
			if region == "" {
				region = klutchaws.ControlPlaneDefaultRegion()
			}
			tenantSecretName = klutchaws.TenantSecretName(createKlutchWorkloadTenantName, createKlutchWorkloadTenantSecretName)
			conn, err := klutchaws.GetTenantCredentials(context.Background(), region, tenantSecretName)
			if err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to load tenant secret %s in %s", tenantSecretName, region))
			}
			if strings.TrimSpace(conn.BindURL) == "" {
				makeup.ExitDueToFatalError(nil, "Tenant secret is missing bind_url. Provide a tenant with bind_url or recreate the tenant.")
			}
			if strings.TrimSpace(conn.TokenURL) == "" || strings.HasPrefix(strings.TrimSpace(conn.TokenURL), "https://.") {
				makeup.ExitDueToFatalError(nil, "Tenant secret has an invalid token_url. Recreate the tenant and ensure Cognito domain provisioning succeeded.")
			}
			if strings.TrimSpace(conn.ClientID) == "" || strings.TrimSpace(conn.ClientSecret) == "" || strings.TrimSpace(conn.Scope) == "" {
				makeup.ExitDueToFatalError(nil, "Tenant secret is missing required OIDC fields (client_id/client_secret/scope). Recreate the tenant.")
			}
			tenantConn = &conn

			if strings.TrimSpace(createKlutchWorkloadBindRequestFile) != "" {
				data, err := os.ReadFile(createKlutchWorkloadBindRequestFile)
				if err != nil {
					makeup.ExitDueToFatalError(err, "Failed to read bind request file override.")
				}
				tenantBindRequest = data
			} else if strings.TrimSpace(conn.BindRequest) != "" {
				tenantBindRequest = []byte(conn.BindRequest)
			}
			if err := klutch.ValidateBindRequest(tenantBindRequest); err != nil {
				makeup.ExitDueToFatalError(err, "Invalid or missing bind request (tenant secret or override).")
			}
		}

		if err := runKlutchClusterCreationWith(demo.KubernetesTool, opts, createKlutchWorkload); err != nil {
			makeup.ExitDueToFatalError(nil, err.Error())
		}

		if tenantConn != nil {
			bindOpts := klutch.NonInteractiveBindOptions{
				ControlPlaneURL:         strings.TrimSpace(tenantConn.BindURL),
				OIDCClientID:            tenantConn.ClientID,
				OIDCClientSecret:        tenantConn.ClientSecret,
				OIDCTokenURL:            tenantConn.TokenURL,
				OIDCScope:               tenantConn.Scope,
				KonnectorImage:          "",
				WriteKubeconfigTo:       "",
				WorkloadKubeconfigPath:  "",
				WorkloadContext:         "",
				BindRequestData:         tenantBindRequest,
				ControlPlaneClusterName: createKlutchWorkloadAutobindControlPlaneName,
			}

			makeup.PrintInfo(fmt.Sprintf("Auto-binding workload cluster %s to control plane cluster %s using credentials from tenant secret %s...", opts.ClusterName, bindOpts.ControlPlaneClusterName, tenantSecretName))
			if err := klutch.NonInteractiveBind(context.Background(), bindOpts); err != nil {
				makeup.ExitDueToFatalError(err, "Failed to bind workload cluster.")
			}
			makeup.PrintSuccessSummary("Workload cluster bound to control plane using tenant secret.")
		}
	},
}

func runKlutchClusterCreation(provider string, opts klutchaws.CreateOptions) error {
	return runKlutchClusterCreationWith(provider, opts, createKlutchControlPlane)
}

func runKlutchClusterCreationWith(provider string, opts klutchaws.CreateOptions, creator func(context.Context, klutchaws.CreateOptions)) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return fmt.Errorf("Please select a provider via -p. Supported provider for Klutch cluster creation is \"aws\".")
	}

	if provider != "aws" {
		return fmt.Errorf("The Klutch cluster creation currently only supports the \"aws\" provider.")
	}

	creator(context.Background(), opts)
	return nil
}

var cmdCreateStackA8s = &cobra.Command{
	Use:   "a8s",
	Short: "Applies the a8s stack to the currently selected Kubernetes cluster.",
	Long:  `Applies the a8s stack to the currently selected Kubernetes cluster, installing the a8s Data Service operator(s) including necessary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		a8s.CreateA8sStack(false)
	},
}

var cmdCreatePGBinding = &cobra.Command{
	Use:   "servicebinding",
	Short: "Create a PostgreSQL service binding = Postgres user/pass + Kubernets Secret.",
	Long:  `Create a PostgreSQL service binding and thus a Kubernetes Secret containing username/password credentials unique to the service binding.`,
	Run: func(cmd *cobra.Command, args []string) {
		a8s := demo.NewA8sDemoManager("")
		demo.A8sPGServiceBinding.ServiceInstanceKind = pg.A8sPGServiceInstanceKind
		a8s.CreatePGServiceBinding()
	},
}

func init() {

	/*
		The required struct to generate a yaml file should already be present in the operator.
		This also creates a tight depedency to the operator itself including
		api versions and the corresponding data schema comprising configurable attributes.
		This means that the CLI version needs to be kept in sync with the operator versions.
		Assuming that more and more services will be supported, it may require to
		modify the CLI from various teams.

		Hence, over time the codebase must be split into sub modules and some types of changes must happen
		fully automtically or otherwise the release process becomes a nightmare and may lead to
		a large delay between operator and CLI releases.

	*/

	rootCmd.PersistentFlags().BoolVarP(&makeup.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")
	rootCmd.PersistentFlags().BoolVarP(&makeup.Verbose, "verbose", "v", false, "enable verbose output?")
	rootCmd.PersistentFlags().BoolVar(&makeup.ShowCommands, "show-commands", false, "output shell commands when they are executed")

	initFlagsCreate(cmdCreate)
	rootCmd.AddCommand(cmdCreate)
}

func initFlagsCreate(cmd *cobra.Command) {
	initFlagsCreatePG(cmdCreatePG)
	cmd.AddCommand(cmdCreatePG)

	initFlagsCreateCluster(cmdCreateCluster)
	cmd.AddCommand(cmdCreateCluster)

	initFlagsCreateKlutchTenant(cmdCreateKlutchTenant)
	cmdCreateKlutch.AddCommand(cmdCreateKlutchTenant)
	cmd.AddCommand(cmdCreateKlutch)

	initFlagsCreateStack(cmdCreateStack)
	cmd.AddCommand(cmdCreateStack)
}

func initFlagsCreateStack(cmdCreateStack *cobra.Command) {
	cmdCreateStack.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")

	initFlagsCreateStackA8s(cmdCreateStackA8s)
	cmdCreateStack.AddCommand(cmdCreateStackA8s)
}

func initFlagsCreateCluster(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" and \"kind\" for local demos, as well as \"aws\" for Klutch.")
	cmd.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")

	initFlagsCreateClusterA8s(cmdCreateClusterA8s)
	cmd.AddCommand(cmdCreateClusterA8s)

	initFlagsCreateClusterKlutch(cmdCreateClusterKlutch)
	cmd.AddCommand(cmdCreateClusterKlutch)
}

func initFlagsCreateClusterKlutch(cmd *cobra.Command) {
	initRequiredPersistentStringFlagP(cmd, &demo.KubernetesTool, "provider", "p", "aws", "provider for creating the Kubernetes cluster. Currently the only valid option for Klutch is \"aws\".")
	cmd.PersistentFlags().StringVar(&createKlutchRegion, "region", "", "AWS region for the EKS cluster (defaults to eu-central-1).")
	cmd.PersistentFlags().StringVar(&createKlutchNodeType, "eks-node-type", "t3a.xlarge", "Instance type for EKS nodegroups.")
	cmd.PersistentFlags().IntVar(&createKlutchNodes, "eks-nodes", 3, "Number of worker nodes (sets min/max/desired to this value).")
	cmd.PersistentFlags().BoolVar(&createClusterKlutchDryRun, "dry-run", false, "Show planned AWS resources and commands for Klutch without creating them.")

	initFlagsCreateClusterKlutchControlPlane(cmdCreateClusterKlutchControlPlane)
	cmd.AddCommand(cmdCreateClusterKlutchControlPlane)

	initFlagsCreateClusterKlutchWorkload(cmdCreateClusterKlutchWorkload)
	cmd.AddCommand(cmdCreateClusterKlutchWorkload)
}

func initFlagsCreatePG(cmd *cobra.Command) {
	// create pg instance
	initFlagsPgInstance(cmdPGInstance)
	cmd.AddCommand(cmdPGInstance)

	initFlagsPgBackup(cmdPGBackup)
	cmd.AddCommand(cmdPGBackup)

	initFlagsPgRestore(cmdPGRestore)
	cmd.AddCommand(cmdPGRestore)

	initFlagsCreatePgBinding(cmdCreatePGBinding)
	cmd.AddCommand(cmdCreatePGBinding)
}

func initFlagsCreateClusterKlutchWorkload(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "", "name of the demo Kubernetes cluster.")
	cmd.Flags().StringVar(&createKlutchWorkloadTenantName, "tenant-name", "", "Tenant name to auto-bind with.")
	cmd.Flags().StringVar(&createKlutchWorkloadTenantSecretName, "tenant-secret-name", "", "Explicit tenant secret name (defaults to klutch/<tenant>/oidc-client).")
	cmd.Flags().StringVar(&createKlutchWorkloadTenantRegion, "tenant-region", "", "AWS region for the tenant secret (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	cmd.Flags().StringVar(&createKlutchWorkloadBindRequestFile, "bind-request-file", "", "Optional bind request JSON to override the tenant's stored bind request.")
	cmd.Flags().StringVar(&createKlutchWorkloadAutobindControlPlaneName, "control-plane-cluster", "", "Control plane cluster name for CA lookup (defaults to klutch-control-plane).")
}

func initFlagsCreateKlutchTenant(cmd *cobra.Command) {
	initRequiredStringFlag(cmd, &createKlutchTenantName, "tenant-name", "", "Name/prefix for the tenant (used to name the Cognito app client).")
	cmd.Flags().StringVar(&createKlutchTenantRegion, "region", "", "AWS region for Cognito (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	cmd.Flags().BoolVar(&createKlutchTenantStoreSecret, "store-secret", true, "Store the tenant credentials in AWS Secrets Manager.")
	cmd.Flags().StringVar(&createKlutchTenantSecretName, "secret-name", "", "Secrets Manager name to store the tenant credentials (defaults to klutch/<tenant>/oidc-client).")
	cmd.Flags().BoolVar(&createKlutchTenantForce, "force", false, "Overwrite an existing tenant secret if it already exists.")
	cmd.Flags().StringVar(&createKlutchTenantBindRequestFile, "bind-request-file", "", "Path to bind request JSON to store with the tenant (defaults to all exported services).")
}

func initFlagsCreateClusterKlutchControlPlane(cmd *cobra.Command) {
	// flags shared with 'create cluster klutch workload' and 'apply klutch control-plane' command
	cmd.Flags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "klutch-control-plane", "name of the demo Kubernetes cluster.")
	cmd.Flags().BoolVar(&createClusterKlutchControlPlaneSkipApply, "no-apply", false, "Create the Klutch control plane cluster without installing the Klutch control plane components.")

	// init flags shared with 'apply klutch control-plane' command
	initSharedFlagsKlutchControlPlaneStack(cmd)
}

func initFlagsCreateStackA8s(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "eu-central-1", "specify the infrastructure region to store backups such as \"eu-central-1\".")
	cmd.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmd.PersistentFlags().StringVar(&demo.BackupInfrastructureProvider, "backup-provider", "minio", "specify the infrastructure provider as supported by the a8s Backup Manager. Valid options are: minio and AWS.")
	cmd.PersistentFlags().StringVar(&demo.BackupInfrastructureEndpoint, "backup-store-endpoint", "", "the endpoint of the S3 compatible backup object storage. When minio is selected, the default is set to http://minio.minio-dev.svc.cluster.local:9000.")
	cmd.PersistentFlags().BoolVar(&demo.BackupInfrastructurePathStyle, "backup-store-pathstyle", false, "influences the URI schema used to talk to the S3 compatible backup object store. Default is false but is set to true when minio is selected as backup-provider.")
	cmd.PersistentFlags().StringVar(&demo.BackupStoreAccessKey, "backup-store-accesskey", "a8s-user", "the access key id for the backup store.")
	cmd.PersistentFlags().StringVar(&demo.BackupStoreSecretKey, "backup-store-secretkey", "a8s-password", "the secret key for the backup store.")
	cmd.PersistentFlags().StringVar(&demo.DeploymentVersion, "deployment-version", demo.DefaultDeploymentVersion, "specify the version corresponding to the a8s-deployment git version tag. Use \"latest\" to get the untagged version.")
	cmd.PersistentFlags().BoolVar(&demo.NoPreCheck, "no-precheck", false, "skip the verification of prerequisites.")
}

func initFlagsCreateClusterA8s(cmd *cobra.Command) {
	// create cluster a8s
	initFlagsCreateStackA8s(cmd)
	cmd.PersistentFlags().StringVar(&demo.ClusterNrOfNodes, "cluster-nr-of-nodes", "3", "specify number of Kubernetes nodes.")
	cmd.PersistentFlags().StringVar(&demo.ClusterMemory, "cluster-memory", "4gb", "specify memory of the Kubernetes cluster.")
}

func initFlagsCreatePgBinding(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceBinding.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG service binding.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceBinding.Name, "name", "example-pg-1", "name of the PG service binding. NOT the name of the PG service instance.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGServiceBinding.Namespace, "namespace", "n", "default", "namespace of the PG service instance. NOT the app's namespace.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGServiceBinding.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to bind to.")
}

func initFlagsPgRestore(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&demo.A8sPGRestore.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG backup.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGRestore.Name, "name", "example-pg-1", "name of the PG restore. Not the name of the service instance or the backup.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGRestore.BackupName, "backup", "b", "example-pg-backup", "name of the PG backup to be restored.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGRestore.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to be restored.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGRestore.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
}

func initFlagsPgBackup(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&demo.A8sPGBackup.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG backup.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGBackup.Name, "name", "example-pg-1", "name of the PG backup. Not the name of the service instance.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGBackup.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to be backed up.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGBackup.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
}

func initFlagsPgInstance(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of thePGservice instance.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.Name, "name", "example-pg", "name of the PG service instance.")
	cmd.PersistentFlags().StringVarP(&demo.A8sPGServiceInstance.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
	cmd.PersistentFlags().IntVar(&demo.A8sPGServiceInstance.Replicas, "replicas", 1, "number of Pods (replicas) the service instance's statefulset will have.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.VolumeSize, "volume-size", "1Gi", "Volume size of the persistent volume claim(s)d of the service instance's statefulset.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.Version, "service-version", "14", "Postgres version. The given version must be supported by the automation.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.RequestsCPU, "requests-cpu", "100m", "Resources -> requests -> cpu of the service instance's statefulset.")
	cmd.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.LimitsMemory, "limits-memory", "100Mi", "Resources -> limits -> memory  of the service instance's statefulset.")
	cmd.PersistentFlags().BoolVar(&demo.DoNotApply, "no-apply", false, "If this flag is set, the service instance YAML spec is not applied (kubectl apply -f).")
}
