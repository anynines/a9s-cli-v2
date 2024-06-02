package demo

import (
	"os"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
)

func CheckK8sCluster(createClusterIfNotExists bool) {
	makeup.PrintH1("Checking Kubernetes Cluster...")

	allGood := true

	clusterSpec := BuildKubernetesClusterSpec()
	clusterManager := BuildKubernetesClusterManager()

	if createClusterIfNotExists {
		clusterManager.CreateKubernetesClusterIfNotExists(clusterSpec)
	}

	// At this point there should be a Kubernetescluster
	if !k8s.CheckIfAnyKubernetesIsRunning() {
		allGood = false
	}

	// Only if there's a suitable cluster, the cluster may also be selected.
	// In any other case, the demo cluster has to be created, first.
	if !clusterManager.IsClusterSelectedAsCurrentContext(DemoClusterName) {
		allGood = false
	}

	// !NoPreCheck > Perform a pre-check
	if !NoPreCheck && !allGood {
		makeup.PrintFailSummary("Sadly, mandatory prerequisites haven't been met. Aborting...")
		os.Exit(1)
	}
}

/*
Builds a cluster manager without params by using shared package variables.
*/
func BuildKubernetesClusterManager() k8s.ClusterManager {

	clusterManager := k8s.BuildClusterManager(
		DemoConfig.WorkingDir,
		DemoClusterName,
		KubernetesTool,
		UnattendedMode,
	)

	return clusterManager
}
