package klutch

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	prereq "github.com/anynines/a9s-cli-v2/prerequisites"
	"gopkg.in/yaml.v2"
)

const (
	demoTitle                       = "Klutch Demo"
	applyControlPlaneTitle          = "Applying Klutch Control Plane to the current Kubernetes cluster"
	controlPlaneClusterInfoFilePath = "klutch"
	controlPlaneClusterInfoFileName = "control-plane-cluster-info.yaml"
	controlPlaneClusterName         = "klutch-control-plane"
	appClusterName                  = "klutch-app"
	contextControlPlane             = "kind-" + controlPlaneClusterName
	contextApp                      = "kind-" + appClusterName
)

var (
	PortFlag int = 8080
)

// ControlPlaneClusterInfo contains information about the created Control Plane Cluster.
type ControlPlaneClusterInfo struct {
	Host        string `yaml:"host"`
	IngressPort string `yaml:"ingressPort"`
}

type KlutchManager struct {
	// cpContext is the kube context for the Klutch Control Plane Cluster.
	cpContext string
	// cpK8s is the Kubernetes client for the Klutch Control Plane Cluster.
	cpK8s *k8s.KubeClient

	// appContext is the kube context for the Klutch App Cluster.
	appContext string
	// appK8s is the Kubernetes client for the Klutch App Cluster.
	appK8s *k8s.KubeClient
}

func NewKlutchManager() *KlutchManager {
	return &KlutchManager{
		cpContext:  contextControlPlane,
		cpK8s:      k8s.NewKubeClient(contextControlPlane),
		appContext: contextApp,
		appK8s:     k8s.NewKubeClient(contextApp),
	}
}

func NewKlutchManagerWithContexts(controlPlaneContext, appContext string) *KlutchManager {
	return &KlutchManager{
		cpContext:  controlPlaneContext,
		cpK8s:      k8s.NewKubeClient(controlPlaneContext),
		appContext: appContext,
		appK8s:     k8s.NewKubeClient(appContext),
	}
}

// DeployKlutchClusters deploys the Control Plane Cluster with all Klutch components, and a
// app cluster.
func DeployKlutchClusters() {
	makeup.PrintWelcomeScreen(
		demo.UnattendedMode,
		demoTitle,
		"Let's deploy a Klutch setup with Kind...")

	// Makes sure the "WorkingDir" variable is set. This allows us to re-use existing code to deploy
	// the a8s stack, and write information about the created cluster to a file in the user's
	// configured working dir.
	demo.EstablishConfig()

	checkDeployPrerequisites()

	klutch := NewKlutchManager()
	klutch.deployControlPlaneCluster()
	klutch.deployAppCluster()
	printSummary()
}

// ApplyKlutchControlPlane installs the Klutch control plane components into the current kube context.
func ApplyKlutchControlPlane(host string, ingressPort int, acmCertificateARN string) {
	makeup.PrintWelcomeScreen(
		demo.UnattendedMode,
		applyControlPlaneTitle,
		"Let's install the Klutch control plane into your current Kubernetes cluster...")

	demo.EstablishConfig()

	checkControlPlaneInstallPrerequisites()

	if ingressPort < 1 || ingressPort > 65535 {
		makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
	}

	manager := NewKlutchManagerWithContexts("", "")
	manager.applyControlPlaneToContext(host, strconv.Itoa(ingressPort), acmCertificateARN)
	printControlPlaneSummary(demo.DemoConfig.WorkingDir)
}

func (k *KlutchManager) deployControlPlaneCluster() {
	hostIP, err := determineHostLocalIP()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't obtain the host's local IP address. Aborting...")
	}

	makeup.PrintInfo(fmt.Sprintf("Using IP address `%s` for the Control Plane Cluster.", hostIP))

	port := strconv.Itoa(PortFlag)
	DeployControlPlaneKindCluster(controlPlaneClusterName, hostIP, port)
	WaitForKindCluster(k.cpK8s)
	writeControlPlaneClusterInfoToFile(demo.DemoConfig.WorkingDir, hostIP, port)

	ingressClass := "nginx"

	k.DeployIngressNginx()
	k.WaitForIngressNginx()

	scheme := determineIngressScheme(ingressClass, false)

	k.DeployDex(hostIP, port, ingressClass, scheme, "")
	k.WaitForDex()

	k.DeployBindBackend(hostIP, port, ingressClass, scheme)
	k.WaitForBindBackend()

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser(demo.UnattendedMode)
	a8s := demo.NewA8sDemoManager(k.cpContext)
	a8s.DeployA8sStack()
}

func (k *KlutchManager) deployAppCluster() {

	DeployAppCluster(appClusterName)
	WaitForKindCluster(k.appK8s)

	switchContext(k.appContext) // in case the app cluster already existed, switch to it.
}

func printSummary() {
	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Deployed a Klutch Control Plane Cluster with Kind.")
	makeup.PrintCheckmark("Deployed Dex Idp and the anynines klutch-bind backend.")
	makeup.PrintCheckmark("Deployed Crossplane and the Kubernetes provider.")
	makeup.PrintCheckmark("Deployed the Klutch Crossplane configuration package.")
	makeup.PrintCheckmark("Deployed Klutch API Service Export Templates to make the Klutch Crossplane APIs available to App Clusters.")
	makeup.PrintCheckmark("Deployed the a8s Stack.")
	makeup.PrintCheckmark("Deployed an App Cluster.")
	makeup.PrintSuccessSummary("You are now ready to bind APIs from the App Cluster using the `a9s klutch bind` command.")
}

func (k *KlutchManager) applyControlPlaneToContext(host string, ingressPort string, acmCertificateARN string) {
	ingressClass := detectIngressClass(k.cpK8s)
	tlsEnabled := acmCertificateARN != "" && ingressClass == "alb"

	// ALB ingress is created without TLS in our manifests; default to port 80 when no cert is provided
	// to avoid calling a closed HTTPS listener. When a cert is provided, force port 443.
	if tlsEnabled && ingressPort != "443" {
		makeup.PrintInfo("ACM certificate provided; enabling TLS on ALB and using port 443.")
		ingressPort = "443"
	} else if ingressClass == "alb" && ingressPort == "443" {
		makeup.PrintInfo("Detected ALB ingress and no TLS configuration. Switching ingress port to 80.")
		ingressPort = "80"
	}

	scheme := determineIngressScheme(ingressClass, tlsEnabled)

	// Bootstrap host: if none was provided, use kubeconfig server host to allow ingress creation,
	// but we will replace it with the ingress LB hostname/IP before finishing.
	initialHost := host
	if initialHost == "" {
		initialHost = getClusterExternalHost("")
		makeup.PrintInfo(fmt.Sprintf("No host provided. Using provisional host `%s` to bootstrap deployment; it will be replaced once ingress has an address.", initialHost))
	}
	host = initialHost

	if ingressClass == "nginx" {
		k.DeployIngressNginx()
		k.WaitForIngressNginx()
	} else {
		makeup.PrintInfo(fmt.Sprintf("Ingress class `%s` detected. Skipping ingress-nginx installation.", ingressClass))
	}

	k.DeployDex(initialHost, ingressPort, ingressClass, scheme, acmCertificateARN)
	k.WaitForDex()

	k.DeployBindBackend(initialHost, ingressPort, ingressClass, scheme)

	// If we're using ALB, wait for the ALB hostname and re-apply Dex + backend manifests
	// with the resolved host so OIDC URLs are correct.
	if ingressClass == "alb" {
		resolvedHost := waitForIngressHost(k.cpK8s, "dex-ingress", "default")
		if resolvedHost == "" {
			makeup.ExitDueToFatalError(nil, "Could not determine ingress hostname/IP. Aborting instead of using the Kubernetes API server host.")
		}
		if resolvedHost != initialHost {
			makeup.PrintInfo(fmt.Sprintf("Detected ALB hostname `%s`. Re-applying Dex and backend manifests with this host.", resolvedHost))
			host = resolvedHost
			k.DeployDex(host, ingressPort, ingressClass, scheme, acmCertificateARN)
			k.DeployBindBackend(host, ingressPort, ingressClass, scheme)
		} else {
			host = initialHost
		}
	}

	k.WaitForBindBackend()

	writeControlPlaneClusterInfoToFile(demo.DemoConfig.WorkingDir, host, ingressPort)

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser(demo.UnattendedMode)
	a8s := demo.NewA8sDemoManager(k.cpContext)
	a8s.DeployA8sStack()
}

func detectIngressClass(k8sClient *k8s.KubeClient) string {
	cmd := k8sClient.KubectlWithContextCommand("get", "ingressclass", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.Output()
	if err != nil {
		makeup.PrintWarning(fmt.Sprintf("Could not detect ingress classes (defaulting to nginx): %v", err))
		return "nginx"
	}

	classes := strings.Fields(string(output))
	hasAlb := false
	hasNginx := false
	for _, c := range classes {
		if c == "alb" {
			hasAlb = true
		}
		if c == "nginx" {
			hasNginx = true
		}
	}

	if hasAlb {
		makeup.PrintInfo("Detected ingress class `alb`. Will use it and skip deploying ingress-nginx.")
		return "alb"
	}

	if hasNginx {
		makeup.PrintInfo("Detected ingress class `nginx`. Will use it.")
		return "nginx"
	}

	makeup.PrintWarning("No ingress class detected. Defaulting to `nginx` and deploying ingress-nginx.")
	return "nginx"
}

// determineIngressScheme returns the URL scheme to use for ingress endpoints.
func determineIngressScheme(_ string, tlsEnabled bool) string {
	if tlsEnabled {
		return "https"
	}
	return "http"
}

// waitForIngressHost waits for an ingress to report a load balancer hostname/IP and returns it.
func waitForIngressHost(k8sClient *k8s.KubeClient, name, namespace string) string {
	timeout := time.After(15 * time.Minute)
	tick := time.Tick(15 * time.Second)
	for {
		select {
		case <-timeout:
			descCmd := k8sClient.KubectlWithContextCommand("describe", "ingress", name, "-n", namespace)
			if out, err := descCmd.CombinedOutput(); err == nil {
				makeup.PrintWarning(fmt.Sprintf("Ingress %s/%s description:\n%s", namespace, name, string(out)))
			}
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Timed out waiting for ingress %s/%s to have a hostname/IP", namespace, name))
		case <-tick:
			cmd := k8sClient.KubectlWithContextCommand("get", "ingress", name, "-n", namespace, "-o", "jsonpath={.status.loadBalancer.ingress[0].hostname}")
			output, err := cmd.Output()
			if err != nil {
				makeup.PrintWarning(fmt.Sprintf("Error while checking ingress %s/%s hostname: %v", namespace, name, err))
				continue
			}
			host := strings.TrimSpace(string(output))
			if host != "" {
				return host
			}

			// try IP as fallback
			cmdIP := k8sClient.KubectlWithContextCommand("get", "ingress", name, "-n", namespace, "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
			ipOutput, err := cmdIP.Output()
			if err == nil {
				ip := strings.TrimSpace(string(ipOutput))
				if ip != "" {
					return ip
				}
			}
			makeup.PrintInfo(fmt.Sprintf("Ingress %s/%s has no hostname/IP yet. Waiting...", namespace, name))
		}
	}
}

func printControlPlaneSummary(workDir string) {
	filePath := filepath.Join(workDir, controlPlaneClusterInfoFilePath, controlPlaneClusterInfoFileName)

	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Installed the Klutch control plane components into the current Kubernetes cluster.")
	makeup.PrintCheckmark("Deployed Dex Idp and the anynines klutch-bind backend.")
	makeup.PrintCheckmark("Deployed Crossplane and the Kubernetes provider.")
	makeup.PrintCheckmark("Deployed the Klutch Crossplane configuration package.")
	makeup.PrintCheckmark("Deployed Klutch API Service Export Templates to make the Klutch Crossplane APIs available to App Clusters.")
	makeup.PrintCheckmark("Deployed the a8s Stack.")
	makeup.PrintCheckmark(fmt.Sprintf("Wrote Control Plane Cluster information to %s", filePath))
	makeup.PrintSuccessSummary("You are now ready to bind APIs from an App Cluster using the `a9s klutch bind` command.")
}

// Writes information about the Control Plane Cluster to a file, to give other commands such as `bind` the information they need.
func writeControlPlaneClusterInfoToFile(workDir string, hostIP string, ingressPort string) {
	info := &ControlPlaneClusterInfo{
		Host:        hostIP,
		IngressPort: ingressPort,
	}

	data, err := yaml.Marshal(info)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Unexpected error while writing Control Plane Cluster info to file.")
	}

	path := filepath.Join(workDir, controlPlaneClusterInfoFilePath)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while writing Control Plane Cluster info to file. Could not create path %s", path))
	}

	file := filepath.Join(path, controlPlaneClusterInfoFileName)
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while writing Control Plane Cluster info to file. Could not write to file %s", file))
	}

	makeup.PrintInfo(fmt.Sprintf("Wrote Control Plane Cluster information to %s", file))
}

// Checks if prerequisites of the deploy command are met.
func checkDeployPrerequisites() {
	makeup.PrintH1("Checking Prerequisites...")

	commonTools := prereq.GetCommonRequiredTools()

	requiredTools := []prereq.RequiredTool{
		commonTools[prereq.ToolGit],
		commonTools[prereq.ToolDocker],
		commonTools[prereq.ToolKind],
		commonTools[prereq.ToolKubectl],
		commonTools[prereq.ToolHelm],
	}

	prereq.CheckRequiredTools(requiredTools)

	prereq.CheckDockerRunning()
}

// Checks if prerequisites of the control plane install command are met.
func checkControlPlaneInstallPrerequisites() {
	makeup.PrintH1("Checking Prerequisites...")

	commonTools := prereq.GetCommonRequiredTools()

	requiredTools := []prereq.RequiredTool{
		commonTools[prereq.ToolGit],
		commonTools[prereq.ToolKubectl],
		commonTools[prereq.ToolHelm],
	}

	prereq.CheckRequiredTools(requiredTools)
}
