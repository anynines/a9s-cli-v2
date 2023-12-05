package demo

import (
	"strconv"

	"github.com/anynines/a9s-cli-v2/creator"
	"github.com/anynines/a9s-cli-v2/makeup"
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
