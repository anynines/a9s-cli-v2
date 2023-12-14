package demo

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anynines/a9s-cli-v2/creator"
	"github.com/anynines/a9s-cli-v2/makeup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Valid options: "kind"
var DemoClusterName string
var UnattendedMode bool // Ask yes-no questions or assume "yes"
var ClusterNrOfNodes string
var ClusterMemory string

var kubernetesCreator creator.KubernetesCreator

func BuildKubernetesClusterSpec() creator.KubernetesClusterSpec {
	nrOfNodes, err := strconv.Atoi(ClusterNrOfNodes)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't determine the number of Kubernetes nodes from the param: "+ClusterNrOfNodes)
	}

	spec := creator.KubernetesClusterSpec{
		Name:                 DemoClusterName,
		NrOfNodes:            nrOfNodes,
		NodeMemory:           ClusterMemory,
		InfrastructureRegion: BackupInfrastructureRegion,
	}

	return spec
}

/*
Singleton-like getter to obtain the K8sCreator.
Instantiates a create if non exists, yet.

Adding a new KubernetesCreator requires to
- modify this method an add the corresponding creator type
- implement the creator type by implementing the creator.KubernetesCreator interface
- implement a unit test for the creator type
*/
func GetKubernetesCreator() creator.KubernetesCreator {

	if kubernetesCreator == nil {
		switch KubernetesTool {
		case "kind":
			kubernetesCreator = creator.KindCreator{LocalWorkDir: DemoConfig.WorkingDir}
		case "minikube":
			kubernetesCreator = creator.MinikubeCreator{}
		default:
			makeup.ExitDueToFatalError(nil, "Invalid Kubernetes providers selected: "+KubernetesTool)
		}
	}

	return kubernetesCreator
}

/*
Deletes the given demo Kubernetes cluster.
*/
func DeleteKubernetesCluster() {
	makeup.PrintWarning("Deleting the Demo Kubernetes Cluster using " + KubernetesTool + " named " + DemoClusterName + "...")

	kCreator := GetKubernetesCreator()

	if kCreator.Exists(DemoClusterName) {
		kCreator.Delete(DemoClusterName, UnattendedMode)
	} else {
		makeup.PrintInfo("There was no cluster using " + KubernetesTool + " named " + DemoClusterName + ". There's nothing to be done.")
	}
	makeup.PrintCheckmark("The Demo Kubernetes Cluster has been deleted.")
}

func checkIfDockerIsRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	if err != nil {
		makeup.PrintFail("Docker is not running.")
		makeup.PrintInfo("Please start the Docker daemon. In case you are using Docker Desktop, start Docker Desktop.")
		return false
	}
	makeup.PrintCheckmark("Docker is running.")
	return true
}

func checkIfKubernetesIsRunning() bool {
	cmd := exec.Command("kubectl", "api-versions")
	err := cmd.Run()
	if err != nil {
		makeup.PrintFail("Kubernetes is not running.")
		makeup.PrintInfo("Please try to restart it or recreate it (delete and re-run the creation).")
		makeup.PrintInfo("Try deleting the Kubernetes cluster with: \"a9s demo delete\". Then recreate it.")
		return false
	}
	makeup.PrintCheckmark("Kubernetes is running.")
	return true
}

func CheckSelectedCluster() {
	makeup.Print("Checking whether the " + DemoClusterName + " cluster is selected...")
	cmd := exec.Command("kubectl", "config", "current-context")

	output, err := cmd.CombinedOutput()

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't retrieve the currently selected cluster using the command: "+cmd.String())
	}

	current_context := strings.TrimSpace(string(output))

	makeup.Print("The currently selected Kubernetes context is: " + current_context)

	desired_context_name := GetKubernetesCreator().GetContext(DemoClusterName)

	if strings.HasPrefix(current_context, desired_context_name) {
		makeup.PrintCheckmark("It seems that the right context is selected: " + desired_context_name)
	} else {
		makeup.PrintFail("The expected context is " + desired_context_name + " but the current context is: " + current_context + ". Please select the desired context! Try executing: ")
		fmt.Println("kubectl config use-context " + desired_context_name)
		os.Exit(1)
	}
}

func GetKubernetesConfigPath() string {
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		makeup.Print("Kubernetes configuration is set by the $KUBECONFIG env variable.")
	} else if home := homedir.HomeDir(); home != "" {
		makeup.Print("Kubernetes configuration is set by $HOME/.kube/config.")
		flag.CommandLine = flag.NewFlagSet("kubeconfig", flag.ExitOnError)
		kubeconfig = *flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		makeup.Print("Kubernetes configuration is set by config flag.")
		kubeconfig = *flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// Set the bool variable based on the flags passed in by the user
	flag.Parse()

	return kubeconfig
}

func GetKubernetesClientSet() *kubernetes.Clientset {
	kubeconfig := GetKubernetesConfigPath()
	makeup.Print("Kubernetes config located at: " + kubeconfig)

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
