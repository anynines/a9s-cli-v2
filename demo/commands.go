package demo

import (
	"fmt"
	"os"

	"github.com/anynines/a9s-cli-v2/makeup"
	prereq "github.com/anynines/a9s-cli-v2/prerequisites"
)

func CheckPrerequisites() {
	allGood := true

	makeup.PrintH1("Checking Prerequisites...")

	SelectClusterProvider()
	checkCommandAvailability()

	// !NoPreCheck > Perform a pre-check
	if !NoPreCheck && !allGood {
		makeup.PrintFailSummary("Sadly, mandatory prerequisites haven't been met. Aborting...")
		os.Exit(1)
	}
}

// checkCommandAvailability checks if the required tools are present on the system.
func checkCommandAvailability() {
	commonTools := prereq.GetCommonRequiredTools()

	requiredTools := []prereq.RequiredTool{
		commonTools[prereq.ToolGit],
		commonTools[prereq.ToolDocker],
		commonTools[prereq.ToolKubectl],
	}

	prereq.CheckRequiredTools(requiredTools)
}

/*
Try to use the provided KubernetesTool if one was explicitly given as a param.
If no param is provided:
If only minikube is present, use minikube.
If only kind is present, use kind.
If multiple kubernetes tools are present, use minikube.
If no suitable kubernetes tool is present, abort with an error message.
*/
func SelectClusterProvider() {
	var tool prereq.RequiredTool
	commonTools := prereq.GetCommonRequiredTools()

	if KubernetesTool != "" {
		switch KubernetesTool {
		case "minikube":
			tool = commonTools[prereq.ToolMinikube]
		case "kind":
			tool = commonTools[prereq.ToolKind]
		default:
			makeup.ExitDueToFatalError(nil, "The selected Kubernetes provider %s is not valid. Valid values are `minikube` or `kind`.")
		}

		// The user has provided a provider param
		if prereq.IsToolAvailable(tool) {
			makeup.PrintCheckmark("Using " + KubernetesTool)
			return
		} else {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("The selected Kubernetes provider %s can't be found. Please verify that it is installed and in the search PATH.", KubernetesTool))
		}
	} else {
		// No param was set.

		// Minikube has priority
		if prereq.IsToolAvailable(commonTools[prereq.ToolMinikube]) {
			KubernetesTool = prereq.ToolMinikube
			makeup.PrintCheckmark("Using " + KubernetesTool)
			return
		} else if prereq.IsToolAvailable(commonTools[prereq.ToolKind]) {
			KubernetesTool = prereq.ToolKind
			makeup.PrintCheckmark("Using " + KubernetesTool)
			return
		} else {
			makeup.ExitDueToFatalError(nil, "No suitable Kubernetes provider was found. Please install either Minikube or Kind.")
		}
	}
}
