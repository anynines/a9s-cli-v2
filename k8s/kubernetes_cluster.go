package k8s

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/makeup"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

/*
Verifies if there's a Kubernetes cluster.
Does not verify whether it is the intended Kubernetes cluster.
*/
func CheckIfAnyKubernetesIsRunning() bool {
	_, err := makeup.NewCommand("kubectl", "api-versions").NoPrompt().Run()
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
	// check if the kubeconfig flag is set, if so, use that value
	if flag.Lookup("kubeconfig") != nil {
		makeup.PrintVerbose("Kubernetes configuration is set by config flag.")
		return flag.Lookup("kubeconfig").Value.String()
	}

	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		makeup.PrintVerbose("Kubernetes configuration is set by the $KUBECONFIG env variable.")
		return kubeconfig
	}

	home := homedir.HomeDir()
	if home == "" {
		makeup.ExitDueToFatalError(nil, "Unable to find Kubernetes configuration.")
	}

	makeup.PrintVerbose("Kubernetes configuration not set by flag or $KUBECONFIG, falling back to $HOME/.kube/config.")
	return filepath.Join(home, ".kube", "config")
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
