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
	demoTitle               = "Klutch Demo"
	mgmtClusterInfoFilePath = "klutch"
	mgmtClusterInfoFileName = "management-cluster-info.yaml"
	mgmtClusterName         = "klutch-management"
	consumerClusterName     = "klutch-consumer"
	contextMgmt             = "kind-" + mgmtClusterName
	contextConsumer         = "kind-" + consumerClusterName
)

var (
	PortFlag int = 8080
)

// ManagementClusterInfo contains information about the created management cluster.
type ManagementClusterInfo struct {
	Host        string `yaml:"host"`
	IngressPort string `yaml:"ingressPort"`
}

type KlutchManager struct {
	mgmtK8s     *k8s.KubeClient
	consumerK8s *k8s.KubeClient
}

func NewKlutchManager() *KlutchManager {
	return &KlutchManager{
		mgmtK8s:     k8s.NewKubeClient(contextMgmt),
		consumerK8s: k8s.NewKubeClient(contextConsumer),
	}
}

// DeployKlutchClusters deploys the central management cluster with all Klutch components, and a
// consumer cluster.
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
	klutch.deployCentralManagementCluster()
	klutch.deployConsumerCluster()
	printSummary()
}

func (k *KlutchManager) deployCentralManagementCluster() {
	hostIP, err := determineHostLocalIP()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't obtain the host's local IP address. Aborting...")
	}

	makeup.PrintInfo(fmt.Sprintf("Using IP address `%s` for the central management cluster.", hostIP))

	port := strconv.Itoa(PortFlag)
	DeployManagementKindCluster(mgmtClusterName, hostIP, port)
	WaitForKindCluster(k.mgmtK8s)
	writeManagementClusterInfoToFile(demo.DemoConfig.WorkingDir, hostIP, port)

	k.DeployIngressNginx()
	k.WaitForIngressNginx()

	k.DeployDex(hostIP, port)
	k.WaitForDex()

	k.DeployBindBackend(hostIP)
	k.WaitForBindBackend()

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser(demo.UnattendedMode)
	a8s := demo.NewA8sDemoManager(contextMgmt)
	a8s.DeployA8sStack()
}

func (k *KlutchManager) deployConsumerCluster() {

	DeployConsumerCluster(consumerClusterName)
	WaitForKindCluster(k.consumerK8s)

	switchContext(contextConsumer) // in case the consumer cluster already existed, switch to it.
}

func printSummary() {
	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Deployed a Klutch management Kind cluster.")
	makeup.PrintCheckmark("Deployed Dex Idp and the anynines klutch-bind backend.")
	makeup.PrintCheckmark("Deployed Crossplane and the Kubernetes provider.")
	makeup.PrintCheckmark("Deployed the Klutch Crossplane configuration package.")
	makeup.PrintCheckmark("Deployed Klutch API Service Export Templates to make the Klutch Crossplane APIs available to consumer clusters.")
	makeup.PrintCheckmark("Deployed the a8s Stack.")
	makeup.PrintCheckmark("Deployed a consumer cluster.")
	makeup.PrintSuccessSummary("You are now ready to bind APIs from the consumer cluster using the `a9s klutch bind` command.")
}

// Writes information about the management cluster to a file, to give other commands such as `bind` the information they need.
func writeManagementClusterInfoToFile(workDir string, hostIP string, ingressPort string) {
	info := &ManagementClusterInfo{
		Host:        hostIP,
		IngressPort: ingressPort,
	}

	data, err := yaml.Marshal(info)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Unexpected error while writing managment cluster info to file.")
	}

	path := filepath.Join(workDir, mgmtClusterInfoFilePath)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while writing managment cluster info to file. Could not create path %s", path))
	}

	file := filepath.Join(path, mgmtClusterInfoFileName)
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while writing managment cluster info to file. Could not write to file %s", file))
	}

	makeup.PrintInfo(fmt.Sprintf("Wrote management cluster information to %s", file))
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
		commonTools[prereq.ToolCrossplane],
		commonTools[prereq.ToolHelm],
	}

	prereq.CheckRequiredTools(requiredTools)

	prereq.CheckDockerRunning()
}
