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

var createKlutchDryRun bool
var createKlutchControlPlane = klutchaws.CreateControlPlaneCluster
var createKlutchWorkload = klutchaws.CreateWorkloadCluster
var createKlutchSkipApply bool
var createKlutchApplyHost string
var createKlutchApplyIngressPort int
var createKlutchApplyACMCertificateARN string
var createKlutchApplyHostedZone string
var createKlutchOIDCProvider string
var createKlutchOIDCIssuerURL string
var createKlutchOIDCClientID string
var createKlutchOIDCClientSecret string
var createKlutchOIDCCallbackURL string
var createKlutchTenantName string
var createKlutchTenantRegion string
var createKlutchTenantUserPoolID string
var createKlutchTenantStoreSecret bool = true
var createKlutchTenantSecretName string
var createKlutchTenantForce bool
var createKlutchTenantTokenURL string
var createKlutchTenantBindURL string
var createKlutchTenantBindRequestFile string
var createKlutchWorkloadAutobindControlPlaneName string

var createKlutchWorkloadTenantName string
var createKlutchWorkloadTenantSecretName string
var createKlutchWorkloadTenantRegion string
var createKlutchWorkloadBindRequestFile string
var createKlutchNodeType string
var createKlutchNodes int
var createKlutchTenantOperatorImage string
var createKlutchTenantOperatorChart string
var createKlutchTenantOperatorChartVersion string
var createKlutchTenantOperatorRoleARN string
var createKlutchTenantOperatorRegion string
var createKlutchTenantOperatorBindURL string
var createKlutchTenantOperatorBindRequest string
var createKlutchBackendImageRef string
var createKlutchBackendImageURL string
var createKlutchBackendImageTag string

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

		cmd.Help()
	},
}

var cmdCreatePG = &cobra.Command{
	Use:   "pg",
	Short: "Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.",
	Long:  `Create PostgreSQL resources such as service instances, service bindings, backups and restore jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please select a PostgreSQL resource such as (service) instance.")
		cmd.Help()
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
		cmd.Help()
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
		cmd.Help()
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
		cmd.Help()
	},
}

var cmdCreateKlutch = &cobra.Command{
	Use:   "klutch",
	Short: "Create Klutch resources.",
	Long:  `Create Klutch resources such as tenants (Cognito app clients).`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		cmd.Help()
	},
}

var cmdCreateClusterKlutchControlPlane = &cobra.Command{
	Use:   "control-plane",
	Short: "Create the Klutch control plane cluster (and install it).",
	Long: `Creates the Klutch control plane cluster on the selected provider and installs the Klutch control plane components.
Use --no-apply to only provision the cluster. Currently only AWS is supported.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := klutchaws.CreateOptions{DryRun: createKlutchDryRun}
		if cmd.Flags().Changed("cluster-name") {
			options.ClusterName = strings.TrimSpace(demo.DemoClusterName)
		}
		options.NodeInstanceTypes = strings.TrimSpace(createKlutchNodeType)
		options.NodeCount = createKlutchNodes
		options.TenantOperatorImage = strings.TrimSpace(createKlutchTenantOperatorImage)
		options.TenantOperatorChart = strings.TrimSpace(createKlutchTenantOperatorChart)
		options.TenantOperatorChartVersion = strings.TrimSpace(createKlutchTenantOperatorChartVersion)
		options.TenantOperatorRoleARN = strings.TrimSpace(createKlutchTenantOperatorRoleARN)
		options.TenantOperatorRegion = strings.TrimSpace(createKlutchTenantOperatorRegion)
		options.TenantOperatorBindURL = strings.TrimSpace(createKlutchTenantOperatorBindURL)
		options.TenantOperatorBindRequest = strings.TrimSpace(createKlutchTenantOperatorBindRequest)
		options.HostedZoneName = strings.TrimSpace(createKlutchApplyHostedZone)

		if createKlutchDryRun {
			makeup.PrintInfo("Skipping Klutch control plane install because --dry-run was provided.")
			return
		}

		if createKlutchSkipApply {
			makeup.PrintInfo("Skipping Klutch control plane install because --no-apply was provided.")
			return
		}

		if createKlutchApplyHostedZone == "" {
			makeup.ExitDueToFatalError(nil, "The --hosted-zone-name flag is required to install the Klutch control plane. Use --no-apply to provision only.")
		}

		if createKlutchApplyIngressPort < 1 || createKlutchApplyIngressPort > 65535 {
			makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
		}

		if err := runKlutchClusterCreation(demo.KubernetesTool, options); err != nil {
			makeup.ExitDueToFatalError(nil, err.Error())
		}

		imgURL, imgTag := resolveBackendImageRef(createKlutchBackendImageRef, createKlutchBackendImageURL, createKlutchBackendImageTag)
		klutch.SetBindBackendImage(imgURL, imgTag)

		klutch.SetControlPlaneOIDCOptions(klutch.OIDCOptions{
			Provider:     klutch.OIDCProvider(createKlutchOIDCProvider),
			IssuerURL:    createKlutchOIDCIssuerURL,
			ClientID:     createKlutchOIDCClientID,
			ClientSecret: createKlutchOIDCClientSecret,
			CallbackURL:  createKlutchOIDCCallbackURL,
		})

		klutch.ApplyKlutchControlPlane(createKlutchApplyHost, createKlutchApplyIngressPort, createKlutchApplyACMCertificateARN, createKlutchApplyHostedZone, options.ClusterName)
	},
}

var cmdCreateClusterKlutchTenant = &cobra.Command{
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
		if createKlutchTenantUserPoolID != "" {
			makeup.PrintWarning("Ignoring --user-pool-id; tenant operator manages user pools.")
		}
		if strings.TrimSpace(createKlutchTenantBindURL) != "" {
			makeup.PrintWarning("Ignoring --bind-url; tenant operator config map provides bind URL.")
		}
		if strings.TrimSpace(createKlutchTenantTokenURL) != "" {
			makeup.PrintWarning("Ignoring --token-url; tenant operator discovers token URL.")
		}

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
			DryRun:               createKlutchDryRun,
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

	// create pg instance
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of thePGservice instance.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.Name, "name", "example-pg", "name of the PG service instance.")
	cmdPGInstance.PersistentFlags().StringVarP(&demo.A8sPGServiceInstance.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
	cmdPGInstance.PersistentFlags().IntVar(&demo.A8sPGServiceInstance.Replicas, "replicas", 1, "number of Pods (replicas) the service instance's statefulset will have.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.VolumeSize, "volume-size", "1Gi", "Volume size of the persistent volume claim(s)d of the service instance's statefulset.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.Version, "service-version", "14", "Postgres version. The given version must be supported by the automation.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.RequestsCPU, "requests-cpu", "100m", "Resources -> requests -> cpu of the service instance's statefulset.")
	cmdPGInstance.PersistentFlags().StringVar(&demo.A8sPGServiceInstance.LimitsMemory, "limits-memory", "100Mi", "Resources -> limits -> memory  of the service instance's statefulset.")
	cmdPGInstance.PersistentFlags().BoolVar(&demo.DoNotApply, "no-apply", false, "If this flag is set, the service instance YAML spec is not applied (kubectl apply -f).")

	// cmdPG.PersistentFlags().StringVarP(&demo.OutputFormat, "output", "o", "", "Output format. Options: \"yaml\".")

	cmdCreatePG.AddCommand(cmdPGInstance)

	cmdPGBackup.PersistentFlags().StringVar(&demo.A8sPGBackup.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG backup.")
	cmdPGBackup.PersistentFlags().StringVar(&demo.A8sPGBackup.Name, "name", "example-pg-1", "name of the PG backup. Not the name of the service instance.")
	cmdPGBackup.PersistentFlags().StringVarP(&demo.A8sPGBackup.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to be backed up.")
	cmdPGBackup.PersistentFlags().StringVarP(&demo.A8sPGBackup.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
	cmdCreatePG.AddCommand(cmdPGBackup)

	// Should the restore act on the backup resource or should there be a separate object for it?
	cmdPGRestore.PersistentFlags().StringVar(&demo.A8sPGRestore.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG backup.")
	cmdPGRestore.PersistentFlags().StringVar(&demo.A8sPGRestore.Name, "name", "example-pg-1", "name of the PG restore. Not the name of the service instance or the backup.")
	cmdPGRestore.PersistentFlags().StringVarP(&demo.A8sPGRestore.BackupName, "backup", "b", "example-pg-backup", "name of the PG backup to be restored.")
	cmdPGRestore.PersistentFlags().StringVarP(&demo.A8sPGRestore.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to be restored.")
	cmdPGRestore.PersistentFlags().StringVarP(&demo.A8sPGRestore.Namespace, "namespace", "n", "default", "namespace of the PG service instance.")
	cmdCreatePG.AddCommand(cmdPGRestore)

	cmdCreate.AddCommand(cmdCreatePG)

	// create pg binding
	// cmdCreatePGBinding.
	cmdCreatePGBinding.PersistentFlags().StringVar(&demo.A8sPGServiceBinding.ApiVersion, "api-version", pg.DefaultPGAPIVersion, "api version of the PG service binding.")
	cmdCreatePGBinding.PersistentFlags().StringVar(&demo.A8sPGServiceBinding.Name, "name", "example-pg-1", "name of the PG service binding. NOT the name of the PG service instance.")
	cmdCreatePGBinding.PersistentFlags().StringVarP(&demo.A8sPGServiceBinding.Namespace, "namespace", "n", "default", "namespace of the PG service instance. NOT the app's namespace.")
	cmdCreatePGBinding.PersistentFlags().StringVarP(&demo.A8sPGServiceBinding.ServiceInstanceName, "service-instance", "i", "example-pg", "name of the PG service instance to bind to.")
	cmdCreatePG.AddCommand(cmdCreatePGBinding)

	// create cluster a8s
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "us-east-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureProvider, "backup-provider", "minio", "specify the infrastructure provider as supported by the a8s Backup Manager. Valid options are: minio and AWS.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureEndpoint, "backup-store-endpoint", "", "the endpoint of the S3 compatible backup object storage. When minio is selected, the default is set to http://minio.minio-dev.svc.cluster.local:9000.")
	cmdCreateClusterA8s.PersistentFlags().BoolVar(&demo.BackupInfrastructurePathStyle, "backup-store-pathstyle", false, "influences the URI schema used to talk to the S3 compatible backup object store. Default is false but is set to true when minio is selected as backup-provider.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupStoreAccessKey, "backup-store-accesskey", "a8s-user", "the access key id for the backup store.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.BackupStoreSecretKey, "backup-store-secretkey", "a8s-password", "the secret key for the backup store.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.DeploymentVersion, "deployment-version", demo.DefaultDeploymentVersion, "specify the version corresponding to the a8s-deployment git version tag. Use \"latest\" to get the untagged version.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.ClusterNrOfNodes, "cluster-nr-of-nodes", "3", "specify number of Kubernetes nodes.")
	cmdCreateClusterA8s.PersistentFlags().StringVar(&demo.ClusterMemory, "cluster-memory", "4gb", "specify memory of the Kubernetes cluster.")
	cmdCreateClusterA8s.PersistentFlags().BoolVar(&demo.NoPreCheck, "no-precheck", false, "skip the verification of prerequisites.")

	//TODO Remove duplicate code in default values between cluster and stack.
	// create stack a8s
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureRegion, "backup-region", "eu-central-1", "specify the infrastructure region to store backups such as \"us-east-1\".")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureBucket, "backup-bucket", "a8s-backups", "specify the infrastructure object store bucket name.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureProvider, "backup-provider", "minio", "specify the infrastructure provider as supported by the a8s Backup Manager. Valid options are: minio and AWS.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupInfrastructureEndpoint, "backup-store-endpoint", "", "the endpoint of the S3 compatible backup object storage. When minio is selected, the default is set to http://minio.minio-dev.svc.cluster.local:9000.")
	cmdCreateStackA8s.PersistentFlags().BoolVar(&demo.BackupInfrastructurePathStyle, "backup-store-pathstyle", false, "influences the URI schema used to talk to the S3 compatible backup object store. Default is false but is set to true when minio is selected as backup-provider.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupStoreAccessKey, "backup-store-accesskey", "a8s-user", "the access key id for the backup store.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.BackupStoreSecretKey, "backup-store-secretkey", "a8s-password", "the secret key for the backup store.")
	cmdCreateStackA8s.PersistentFlags().StringVar(&demo.DeploymentVersion, "deployment-version", demo.DefaultDeploymentVersion, "specify the version corresponding to the a8s-deployment git version tag. Use \"latest\" to get the untagged version.")
	cmdCreateStackA8s.PersistentFlags().BoolVar(&demo.NoPreCheck, "no-precheck", false, "skip the verification of prerequisites.")

	// create demo
	cmdCreateClusterKlutchControlPlane.Flags().BoolVar(&createKlutchSkipApply, "no-apply", false, "Create the Klutch control plane cluster without installing the Klutch control plane components.")
	initRequiredStringFlagWithDependency(&createKlutchSkipApply, "no-apply", false, cmdCreateClusterKlutchControlPlane, &createKlutchApplyHostedZone, "hosted-zone-name", "", "Route53 hosted zone name (FQDN). Required unless --no-apply is set. If provided and no ACM ARN is supplied, the CLI will request an ACM cert and create DNS validation records automatically.")
	cmdCreateCluster.PersistentFlags().StringVarP(&demo.KubernetesTool, "provider", "p", "", "provider for creating the Kubernetes cluster. Valid options are \"minikube\" and \"kind\" for local demos, as well as \"aws\" for Klutch.")
	cmdCreateCluster.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")
	cmdCreateClusterKlutchControlPlane.Flags().BoolVar(&createKlutchDryRun, "dry-run", false, "Show planned AWS resources and commands for Klutch without creating them.")
	cmdCreateClusterKlutchWorkload.Flags().BoolVar(&createKlutchDryRun, "dry-run", false, "Show planned AWS resources and commands for Klutch without creating them.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchApplyHost, "host", "", "Host (IP or DNS name) to reach the ingress when applying the control plane. Defaults to the Kubernetes API server host of the current kube context.")
	cmdCreateClusterKlutchControlPlane.Flags().IntVar(&createKlutchApplyIngressPort, "ingress-port", 443, "Port the ingress should listen on when applying the control plane.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchApplyACMCertificateARN, "acm-certificate-arn", "", "ACM certificate ARN to enable HTTPS on the ALB ingress when applying the control plane.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchOIDCProvider, "oidc-provider", "", "OIDC provider to use for the Klutch control plane. Defaults to cognito when --provider=aws, otherwise dex.")
	initRequiredStringFlagWithDependency(&createKlutchOIDCProvider, "oidc-provider", "cognito", cmdCreateClusterKlutchControlPlane, &createKlutchOIDCIssuerURL, "oidc-issuer-url", "", "OIDC issuer URL (required for oidc-provider=cognito).")
	initRequiredStringFlagWithDependency(&createKlutchOIDCProvider, "oidc-provider", "cognito", cmdCreateClusterKlutchControlPlane, &createKlutchOIDCClientID, "oidc-client-id", "", "OIDC client ID (required for oidc-provider=cognito).")
	initRequiredStringFlagWithDependency(&createKlutchOIDCProvider, "oidc-provider", "cognito", cmdCreateClusterKlutchControlPlane, &createKlutchOIDCClientSecret, "oidc-client-secret", "", "OIDC client secret (required for oidc-provider=cognito).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchOIDCCallbackURL, "oidc-callback-url", "", "OIDC callback URL to configure on the backend. Defaults to https://<host>/callback when not provided.")
	initRequiredStringFlagP(cmdCreateClusterKlutchControlPlane, &demo.KubernetesTool, "provider", "p", "aws", "provider for creating the Kubernetes cluster. Currently the only valid option for Klutch is \"aws\".")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchTenantOperatorImage, "tenant-operator-image", "", "Tenant operator container image (defaults to built-in ECR dev image).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchTenantOperatorChart, "tenant-operator-chart", "", "Tenant operator Helm chart (OCI URL).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchTenantOperatorRoleARN, "tenant-operator-role-arn", "", "IAM role ARN for the tenant operator service account (IRSA).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchTenantOperatorRegion, "tenant-operator-region", "", "Region for tenant operator AWS calls (defaults to control-plane region).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchTenantOperatorBindURL, "tenant-operator-bind-url", "", "Bind URL to pass to the tenant operator config (override default derived bind URL).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchTenantOperatorBindRequest, "tenant-operator-bind-request", "", "Bind request JSON to pass to the tenant operator config (defaults to all exported services).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchNodeType, "eks-node-type", "t3a.xlarge", "Instance type for EKS nodegroups.")
	cmdCreateClusterKlutchControlPlane.Flags().IntVar(&createKlutchNodes, "eks-nodes", 3, "Number of worker nodes (sets min/max/desired to this value).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchBackendImageRef, "klutch-bind-backend-img", "", "Override the klutch-bind backend image as <repo>:<tag>.")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchBackendImageURL, "klutch-bind-backend-img-url", "", "Override the klutch-bind backend image URL (repository).")
	cmdCreateClusterKlutchControlPlane.Flags().StringVar(&createKlutchBackendImageTag, "klutch-bind-backend-img-tag", "", "Override the klutch-bind backend image tag.")
	initRequiredStringFlag(cmdCreateClusterKlutchTenant, &createKlutchTenantName, "tenant-name", "", "Name/prefix for the tenant (used to name the Cognito app client).")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantRegion, "region", "", "AWS region for Cognito (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantUserPoolID, "user-pool-id", "", "Existing Cognito user pool ID to reuse. If omitted, a pool named <tenant>-klutch is created or reused.")
	cmdCreateClusterKlutchTenant.Flags().BoolVar(&createKlutchTenantStoreSecret, "store-secret", true, "Store the tenant credentials in AWS Secrets Manager.")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantSecretName, "secret-name", "", "Secrets Manager name to store the tenant credentials (defaults to klutch/<tenant>/oidc-client).")
	cmdCreateClusterKlutchTenant.Flags().BoolVar(&createKlutchTenantForce, "force", false, "Overwrite an existing tenant secret if it already exists.")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantTokenURL, "token-url", "", "Override the OAuth2 token URL (defaults to Cognito hosted domain token endpoint).")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantBindURL, "bind-url", "", "Control-plane bind URL to store with the tenant (e.g. https://klutch-bind.example.com/bind-noninteractive).")
	cmdCreateClusterKlutchTenant.Flags().StringVar(&createKlutchTenantBindRequestFile, "bind-request-file", "", "Path to bind request JSON to store with the tenant (defaults to all exported services).")
	initRequiredStringFlagP(cmdCreateClusterKlutchWorkload, &demo.KubernetesTool, "provider", "p", "aws", "provider for creating the Kubernetes cluster. Currently the only valid option for Klutch is \"aws\".")
	cmdCreateClusterKlutchWorkload.Flags().StringVar(&createKlutchWorkloadTenantName, "tenant-name", "", "Tenant name to auto-bind with.")
	cmdCreateClusterKlutchWorkload.Flags().StringVar(&createKlutchWorkloadTenantSecretName, "tenant-secret-name", "", "Explicit tenant secret name (defaults to klutch/<tenant>/oidc-client).")
	cmdCreateClusterKlutchWorkload.Flags().StringVar(&createKlutchWorkloadTenantRegion, "tenant-region", "", "AWS region for the tenant secret (defaults to CONTROL_PLANE_CLUSTER_REGION or eu-central-1).")
	cmdCreateClusterKlutchWorkload.Flags().StringVar(&createKlutchWorkloadBindRequestFile, "bind-request-file", "", "Optional bind request JSON to override the tenant's stored bind request.")
	cmdCreateClusterKlutchWorkload.Flags().StringVar(&createKlutchNodeType, "eks-node-type", "t3a.xlarge", "Instance type for EKS nodegroups.")
	cmdCreateClusterKlutchWorkload.Flags().StringVar(&createKlutchWorkloadAutobindControlPlaneName, "control-plane-cluster", "", "Control plane cluster name for CA lookup (defaults to klutch-control-plane).")
	cmdCreateClusterKlutchWorkload.Flags().IntVar(&createKlutchNodes, "eks-nodes", 3, "Number of worker nodes (sets min/max/desired to this value).")

	cmdCreateCluster.AddCommand(cmdCreateClusterA8s)
	cmdCreateClusterKlutch.AddCommand(cmdCreateClusterKlutchControlPlane)
	cmdCreateClusterKlutch.AddCommand(cmdCreateClusterKlutchWorkload)
	cmdCreateClusterKlutch.AddCommand(cmdCreateClusterKlutchTenant)
	cmdCreateCluster.AddCommand(cmdCreateClusterKlutch)
	cmdCreateKlutch.AddCommand(cmdCreateClusterKlutchTenant)
	cmdCreateStack.AddCommand(cmdCreateStackA8s)
	cmdCreateStack.PersistentFlags().StringVarP(&demo.DemoClusterName, "cluster-name", "c", "a8s-demo", "name of the demo Kubernetes cluster.")

	cmdCreate.AddCommand(cmdCreateCluster)
	cmdCreate.AddCommand(cmdCreateKlutch)
	cmdCreate.AddCommand(cmdCreateStack)
	rootCmd.PersistentFlags().BoolVarP(&makeup.UnattendedMode, "yes", "y", false, "skip yes-no questions by answering with \"yes\".")
	rootCmd.PersistentFlags().BoolVarP(&makeup.Verbose, "verbose", "v", false, "enable verbose output?")
	rootCmd.PersistentFlags().BoolVar(&makeup.ShowCommands, "show-commands", false, "output shell commands when they are executed")
	rootCmd.AddCommand(cmdCreate)
}
