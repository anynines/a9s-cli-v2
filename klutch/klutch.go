package klutch

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	prereq "github.com/anynines/a9s-cli-v2/prerequisites"
	"gopkg.in/yaml.v2"
)

const (
	demoTitle                       = "Klutch Demo"
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
func ApplyKlutchControlPlane(host string, ingressPort int) {
	makeup.PrintWelcomeScreen(
		demo.UnattendedMode,
		demoTitle,
		"Let's install the Klutch control plane into your current Kubernetes cluster...")

	demo.EstablishConfig()

	checkControlPlaneInstallPrerequisites()

	if host == "" {
		derivedHost := getClusterExternalHost("")
		makeup.PrintInfo(fmt.Sprintf("No host provided via --host. Using cluster server host `%s`.", derivedHost))
		host = derivedHost
	}

	if ingressPort < 1 || ingressPort > 65535 {
		makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
	}

	manager := NewKlutchManagerWithContexts("", "")
	manager.applyControlPlaneToContext(host, strconv.Itoa(ingressPort))
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

	k.DeployIngressNginx()
	k.WaitForIngressNginx()

	k.DeployDex(hostIP, port)
	k.WaitForDex()

	k.DeployBindBackend(hostIP)
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

func (k *KlutchManager) applyControlPlaneToContext(host string, ingressPort string) {
	writeControlPlaneClusterInfoToFile(demo.DemoConfig.WorkingDir, host, ingressPort)

	k.DeployIngressNginx()
	k.WaitForIngressNginx()

	k.DeployDex(host, ingressPort)
	k.WaitForDex()

	k.DeployBindBackend(host)
	k.WaitForBindBackend()

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser(demo.UnattendedMode)
	a8s := demo.NewA8sDemoManager(k.cpContext)
	a8s.DeployA8sStack()
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
