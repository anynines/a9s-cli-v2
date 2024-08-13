package k8s

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/creator"
	"github.com/anynines/a9s-cli-v2/makeup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var Creator creator.KubernetesCreator

func CheckIfDockerIsRunning() bool {
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

/*
Verifies if there's a Kubernetes cluster.
Does not verify whether it is the intended Kubernetes cluster.
*/
func CheckIfAnyKubernetesIsRunning() bool {
	cmd := exec.Command("kubectl", "api-versions")
	err := cmd.Run()
	if err != nil {
		makeup.PrintFail("Kubernetes is not running.")
		makeup.PrintInfo("Please try to restart it or recreate it (delete and re-run the creation).")
		makeup.PrintInfo("Try deleting the Kubernetes cluster with: \"a9s delete cluster a8s\". Then recreate it.")
		return false
	}
	makeup.PrintCheckmark("Kubernetes is running.")
	return true
}

func GetKubernetesConfigPath() string {
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		makeup.PrintVerbose("Kubernetes configuration is set by the $KUBECONFIG env variable.")
	} else if home := homedir.HomeDir(); home != "" {
		makeup.PrintVerbose("Kubernetes configuration is set by $HOME/.kube/config.")
		flag.CommandLine = flag.NewFlagSet("kubeconfig", flag.ExitOnError)
		kubeconfig = *flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		makeup.PrintVerbose("Kubernetes configuration is set by config flag.")
		kubeconfig = *flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// Set the bool variable based on the flags passed in by the user
	flag.Parse()

	return kubeconfig
}

func (k *KubeClient) GetKubernetesConfig() *rest.Config {
	kubeconfigPath := GetKubernetesConfigPath()
	makeup.PrintVerbose("Kubernetes config located at: " + kubeconfigPath)

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: k.KubeContext}).ClientConfig()

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't create Kubernetes config.")
	}

	return config
}

func (k *KubeClient) GetKubernetesClientSet() *kubernetes.Clientset {
	config := k.GetKubernetesConfig()

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't create Kubernetes ClientSet.")
	}

	return clientset
}
