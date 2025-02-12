package klutch

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
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
	PortFlag                 int    = 8080
	KeycloakCaPathFlag       string = ""
	IdTokenModeFlag          bool   = false
	BackendIssuerUrlFlag     string = ""
	BackendClientIdFlag      string = ""
	BackendClientSecretFlag  string = ""
	LoadKonnectorImageFlag   string = ""
	LoadBackendImageFlag     string = ""
	OIDCClusterClientIDFlag  string = ""
	OIDCClusterIssuerURLFlag string = ""
	KindClusterOnlyFlag      bool   = false
)

// ControlPlaneClusterInfo contains information about the created Control Plane Cluster.
type ControlPlaneClusterInfo struct {
	Host        string `yaml:"host"`
	IngressPort string `yaml:"ingressPort"`
}

type KlutchManager struct {
	// cpK8s is the Kubernetes client for the Klutch Control Plane Cluster.
	cpK8s *k8s.KubeClient

	// appK8s is the Kubernetes client for the Klutch App Cluster.
	appK8s *k8s.KubeClient
}

func NewKlutchManager() *KlutchManager {
	return &KlutchManager{
		cpK8s:  k8s.NewKubeClient(contextControlPlane),
		appK8s: k8s.NewKubeClient(contextApp),
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

	if KindClusterOnlyFlag {
		makeup.PrintSuccess("Done")
		os.Exit(0)
	}

	k.DeployIngressNginx()
	k.WaitForIngressNginx()

	k.DeployDex(hostIP, port, BackendClientIdFlag, BackendClientSecretFlag, BackendIssuerUrlFlag)
	k.WaitForDex()

	k.DeployBindBackend(hostIP)
	k.WaitForBindBackend()

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser(demo.UnattendedMode)
	a8s := demo.NewA8sDemoManager(contextControlPlane)
	a8s.DeployA8sStack()
}

func (k *KlutchManager) deployAppCluster() {

	DeployAppCluster(appClusterName)
	WaitForKindCluster(k.appK8s)

	if LoadKonnectorImageFlag != "" {
		cmd := exec.Command("kind", "load", "docker-image", LoadKonnectorImageFlag, "-n", "klutch-app")
		out, err := cmd.CombinedOutput()
		if err != nil {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Failed to load docker image: %v", string(out)))
			os.Exit(1)
		}
	}

	switchContext(contextApp) // in case the app cluster already existed, switch to it.
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
