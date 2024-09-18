package klutch

import (
	_ "embed"
	"fmt"
	"os/exec"
	"strings"

	"github.com/anynines/a9s-cli-v2/demo"
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

// DeployManagementKindCluster creates a new kind cluster configured to act as a local central management cluster.
// It enables the ingress feature for the provided port and configures the k8s API Server to listen on the provided IP.
func DeployManagementKindCluster(clusterName string, ingressPort string) {
	exists, err := clusterExists(clusterName)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("unexpected error while checking if cluster %s exists:", clusterName))
	}

	if exists {
		makeup.PrintWarning(fmt.Sprintf("Cluster %s already exists. Skipping creation. If the existing cluster is not correctly configured, Klutch will not work. In that case, delete the cluster and start again.", clusterName))
		makeup.WaitForUser(demo.UnattendedMode)
		return
	}

	templateVars := clusterConfigTemplateVars{
		Name:            clusterName,
		BackendHostPort: ingressPort,
	}

	renderedTemplate, err := renderTemplate(cmcKindConfigTemplate, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to render template with parameters %+v.", templateVars))
	}

	makeup.PrintH2("Creating a kind cluster with following config: ")
	makeup.PrintYAML(renderedTemplate.Bytes(), false)
	makeup.WaitForUser(demo.UnattendedMode)

	cmd := exec.Command("kind", "create", "cluster", "--config", "-")
	cmd.Stdin = renderedTemplate

	// Print stderr to show progress to the user.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the command.")
	}

	go printCommandOutput(stderr)

	if err := cmd.Start(); err != nil {
		makeup.ExitDueToFatalError(err, "Could not start the command.")
	}

	if err := cmd.Wait(); err != nil {
		makeup.ExitDueToFatalError(err, "Error occured while executing the command.")
	}
}

func WaitForKindCluster(k8s *k8s.KubeClient) {
	makeup.PrintH1("Waiting for the Kind cluster to become ready...")

	k8s.KubectlWaitForNodes()

	k8s.KubectlWaitForSystemToBecomeReady("kube-system", []string{
		"k8s-app=kube-dns",
		"k8s-app=kube-proxy",
		"k8s-app=kindnet",
		"component=kube-controller-manager",
	})

	makeup.PrintCheckmark("Kind cluster appears to be ready.")
}

// DeployConsumerCluster deploys a simple kind cluster with the given name if it doesn't already exists.
func DeployConsumerCluster(clusterName string) {
	makeup.PrintH1("Deploying a Consumer Kind cluster...")
	exists, err := clusterExists(clusterName)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while checking if cluster %s exists", clusterName))
	}

	if exists {
		makeup.PrintWarning(fmt.Sprintf("Cluster %s already exists. Skipping creation.", clusterName))
		makeup.WaitForUser(demo.UnattendedMode)
		return
	}

	cmd := exec.Command("kind", "create", "cluster", "--name", clusterName)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(demo.UnattendedMode)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the command.")
	}

	go printCommandOutput(stderr)

	if err := cmd.Start(); err != nil {
		makeup.ExitDueToFatalError(err, "Could not start the command.")
	}

	if err := cmd.Wait(); err != nil {
		makeup.ExitDueToFatalError(err, "Error occured while executing the command.")
	}
}

// TODO: Similar code exists in kind_creator.go, but awkward to use in this specific case.
func clusterExists(clusterName string) (bool, error) {
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.CombinedOutput()
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
