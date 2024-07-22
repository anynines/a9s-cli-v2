package demo

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/anynines/a9s-cli-v2/makeup"
)

func CheckPrerequisites() {
	allGood := true

	makeup.PrintH1("Checking Prerequisites...")

	SelectClusterProvider()
	CheckCommandAvailability()

	// !NoPreCheck > Perform a pre-check
	if !NoPreCheck && !allGood {
		makeup.PrintFailSummary("Sadly, mandatory prerequisites haven't been met. Aborting...")
		os.Exit(1)
	}
}

func IsCommandAvailable(cmdName string) bool {
	path, err := exec.LookPath(cmdName)
	if err != nil {
		requiredCmds := RequiredCommands()

		msg := "Couldn't find " + cmdName + " command: " + err.Error() + "."

		if requiredCmds[cmdName][runtime.GOOS] != "" {
			msg += " Try running: " + requiredCmds[cmdName][runtime.GOOS]
		}

		makeup.PrintFail(msg)

		return false
	}

	makeup.PrintCheckmark("Found " + cmdName + " at path " + path + ".")

	return true
}

func CheckCommandAvailability() {

	allGood := true

	requiredCmds := RequiredCommands()

	// cmdDetails
	for cmdName := range requiredCmds {

		if !IsCommandAvailable(cmdName) {
			allGood = false
		}
	}

	if !allGood {
		makeup.PrintFailSummary("Sadly, mandatory commands are missing. Aborting...")
		os.Exit(1)
	} else {
		makeup.PrintSuccessSummary("All necessary commands are present.")
	}
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
	if KubernetesTool != "" {

		// The user has provided a provider param
		if IsCommandAvailable(KubernetesTool) {
			makeup.PrintCheckmark("Using " + KubernetesTool)
			return
		} else {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("The selected Kubernetes provider %s can't be found. Please verify that it is installed and in the search PATH.", KubernetesTool))
		}
	} else {
		// No param was set.

		// Minikube has priority
		if IsCommandAvailable("minikube") {
			KubernetesTool = "minikube"
			makeup.PrintCheckmark("Using " + KubernetesTool)
			return
		} else if IsCommandAvailable("kind") {
			KubernetesTool = "kind"
			makeup.PrintCheckmark("Using " + KubernetesTool)
			return
		} else {
			makeup.ExitDueToFatalError(nil, "No suitable Kubernetes provider was found. Please install either Minikube or Kind.")
		}
	}
}
