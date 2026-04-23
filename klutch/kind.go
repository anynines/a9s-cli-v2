package klutch

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
)

//go:embed templates/kindConfig.tmpl
var cmcKindConfigTemplate string

type clusterConfigTemplateVars struct {
	Name            string
	BackendHostPort string
	HostLanIP       string
}

// DeployControlPlaneKindCluster creates a new kind cluster configured to act as a local Control Plane Cluster.
// It enables the ingress feature for the provided port and configures the k8s API Server to listen on the provided IP.
func DeployControlPlaneKindCluster(clusterName string, hostIP string, ingressPort string) {
	exists, err := clusterExists(clusterName)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("unexpected error while checking if cluster %s exists:", clusterName))
	}

	if exists {
		makeup.PrintWarning(fmt.Sprintf("Cluster %s already exists. Skipping creation. If the existing cluster is not correctly configured, Klutch will not work. In that case, delete the cluster and start again.", clusterName))
		makeup.WaitForUser()
		return
	}

	templateVars := clusterConfigTemplateVars{
		Name:            clusterName,
		BackendHostPort: ingressPort,
		HostLanIP:       hostIP,
	}

	renderedTemplate, err := renderTemplate(cmcKindConfigTemplate, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to render template with parameters %+v.", templateVars))
	}

	makeup.PrintH2("Creating a kind cluster with following config: ")

	if out, err := makeup.Command("kind", "create", "cluster", "--config", "-").Stdin(renderedTemplate.Bytes()).WithPrompt().Run(); err != nil {
		makeup.ExitDueToFatalError(err, "An error occurred while executing the command 'kind create cluster':\n"+string(out))
	}
}

func WaitForKindCluster(k8s *k8s.KubeClient) {
	makeup.PrintH1("Waiting for the Kind cluster to become ready...")

	k8s.KubectlWaitForResourceCondition("ready", "node", "", "", "")

	k8s.KubectlWaitForSystemToBecomeReady("kube-system", []string{
		"k8s-app=kube-dns",
		"k8s-app=kube-proxy",
		"k8s-app=kindnet",
		"component=kube-controller-manager",
	})

	makeup.PrintCheckmark("Kind cluster appears to be ready.")
}

// DeployAppCluster deploys a simple kind cluster with the given name if it doesn't already exists.
func DeployAppCluster(clusterName string) {
	makeup.PrintH1("Deploying an App Cluster with Kind ...")
	exists, err := clusterExists(clusterName)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while checking if cluster %s exists", clusterName))
	}

	if exists {
		makeup.PrintWarning(fmt.Sprintf("Cluster %s already exists. Skipping creation.", clusterName))
		makeup.WaitForUser()
		return
	}

	if _, err := makeup.Command("kind", "create", "cluster", "--name", clusterName).WithPrompt().Run(); err != nil {
		makeup.ExitDueToFatalError(err, "Could not start the command.")
	}
}

// TODO: Similar code exists in kind_creator.go, but awkward to use in this specific case.
func clusterExists(clusterName string) (bool, error) {
	output, err := makeup.Command("kind", "get", "clusters").NoPrompt().Run()
	if err != nil {
		return false, err
	}

	strOutput := string(output)
	split := strings.Split(strOutput, "\n")
	for _, entry := range split {
		if entry == clusterName {
			return true, nil
		}
	}

	return false, nil
}
