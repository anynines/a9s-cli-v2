package k8s

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/anynines/a9s-cli-v2/creator"
	"github.com/anynines/a9s-cli-v2/makeup"
)

type ClusterManager struct {
	ClusterName    string                    // name of the cluster
	Creator        creator.KubernetesCreator // e.g. creator.KindCreator
	CreatorName    string                    // "kind", "minikube"
	UnattendedMode bool                      // false > Ask user for feedback inbetween steps
	WorkDir        string                    // general working directory, e.g. where to store kind manifests
}

/*
ClusterManager factory function.

Build a ClusterManager instance.
*/
func BuildClusterManager(workDir, name, creatorName string, unattendedMode bool) ClusterManager {

	manager := ClusterManager{
		ClusterName:    name,
		WorkDir:        workDir,
		CreatorName:    creatorName,
		UnattendedMode: unattendedMode,
	}

	switch creatorName {
	case "kind":
		manager.Creator = creator.KindCreator{LocalWorkDir: workDir}
	case "minikube":
		manager.Creator = creator.MinikubeCreator{}
	default:
		makeup.ExitDueToFatalError(nil, "Invalid Kubernetes provider selected: "+creatorName)
	}

	return manager
}

func (m *ClusterManager) CreateKubernetesClusterIfNotExists(spec creator.KubernetesClusterSpec) {
	if !m.Creator.Exists(m.ClusterName) {
		m.Creator.Create(spec, m.UnattendedMode)
	}
}

func (m *ClusterManager) DeleteKubernetesCluster() {
	makeup.PrintWarning("Deleting the Kubernetes Cluster using " + m.CreatorName + " named " + m.ClusterName + "...")

	kCreator := m.Creator

	if kCreator.Exists(m.ClusterName) {
		kCreator.Delete(m.ClusterName, m.UnattendedMode)
	} else {
		makeup.PrintInfo("There was no cluster using " + m.CreatorName + " named " + m.ClusterName + ". There's nothing to be done.")
	}
	makeup.PrintCheckmark("The Kubernetes Cluster " + m.ClusterName + " has been deleted.")
}

/*
Checks whether given cluster is also set as the current context.
This ensures that subsequent actions are applied to the correct Kubernetes cluster.
*/
func (m *ClusterManager) IsClusterSelectedAsCurrentContext(clusterName string) bool {
	makeup.Print("Checking whether the " + clusterName + " cluster is selected...")
	cmd := exec.Command("kubectl", "config", "current-context")

	output, err := cmd.CombinedOutput()

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't retrieve the currently selected cluster using the command: "+cmd.String())
	}

	current_context := strings.TrimSpace(string(output))

	makeup.Print("The currently selected Kubernetes context is: " + current_context)

	desiredContextName := m.Creator.GetContext(clusterName)

	if strings.HasPrefix(current_context, desiredContextName) {
		makeup.PrintCheckmark("It seems that the right context is selected: " + desiredContextName)
		return true
	} else {
		makeup.PrintFail("The expected context is " + desiredContextName + " but the current context is: " + current_context + ". Please select the desired context! Try executing: ")
		fmt.Println("kubectl config use-context " + desiredContextName)
		return false
	}
}
