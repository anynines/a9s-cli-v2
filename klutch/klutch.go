package klutch

import (
	"context"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	PortFlag     int  = 8080
	LoopbackMode bool = false

	//go:embed templates/proxySidecarPatch.tmpl
	proxySidecarPatch string
)

// ControlPlaneClusterInfo contains information about the created Control Plane Cluster.
type ControlPlaneClusterInfo struct {
	Host                string `yaml:"host"`
	BackendExposurePort string `yaml:"backendExposurePort"`
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
	klutch.deployKindControlPlaneCluster()
	klutch.deployAppCluster()
	printSummary()
}

func (k *KlutchManager) deployKindControlPlaneCluster() {
	hostIP, err := determineHostLocalIP()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't obtain the host's local IP address. Aborting...")
	}

	makeup.PrintInfo(fmt.Sprintf("Using IP address `%s` for the Control Plane Cluster.", hostIP))

	port := strconv.Itoa(PortFlag)
	DeployControlPlaneKindCluster(controlPlaneClusterName, hostIP, port)
	WaitForKindCluster(k.cpK8s)
	writeControlPlaneClusterInfoToFile(demo.DemoConfig.WorkingDir, hostIP, port)

	k.DeployEnvoyGateway()
	k.WaitForEnvoyGateway()

	k.DeployDex(hostIP, port)
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
func writeControlPlaneClusterInfoToFile(workDir string, hostIP string, backendExposurePort string) {
	info := &ControlPlaneClusterInfo{
		Host:                hostIP,
		BackendExposurePort: backendExposurePort,
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

func (k *KlutchManager) addLoopbackProxyToDeployment(k8sClient *k8s.KubeClient, deploymentNamespace, deploymentName, proxyName, port string) {
	makeup.PrintWait(`Loopback Mode is active, checking for proxy "` + proxyName + `" for port ` + port + ` to Host Loopback Device to Deployment "` +
		deploymentName + `" in namespace "` + deploymentNamespace + `"`)
	templateVars := struct{ Name, Port, Operation, ContainerIndex string }{proxyName, port, "add", "-"}
	deployment, err := k8sClient.GetKubernetesClientSet().AppsV1().Deployments(deploymentNamespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		makeup.ExitDueToFatalError(err, `Failed to retrieve Deployment "`+deploymentNamespace+"/"+deploymentName+`" to check for existing proxy`)
	}

	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == proxyName {
			makeup.PrintWait("Proxy container already exists, switching patch operation to 'replace'")
			templateVars.Operation = "replace"
			templateVars.ContainerIndex = fmt.Sprintf("%d", i)
			break
		}
	}

	deploymentPatch, err := renderTemplate(proxySidecarPatch, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, `Failed to render the template for patching the Deployment "`+deploymentNamespace+"/"+deploymentName+`"`)
	}
	_, output, err := k8sClient.Kubectl(demo.UnattendedMode, "patch", "deployment", "-n", deploymentNamespace, deploymentName,
		"--type=json", "-p="+deploymentPatch.String(),
	)
	if err != nil {
		makeup.ExitDueToFatalError(err, `Failed to apply prox patch for Deployment "`+deploymentNamespace+"/"+deploymentName+`": `+string(output))
	}
}
