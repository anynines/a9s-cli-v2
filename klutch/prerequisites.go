package klutch

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/anynines/a9s-cli-v2/makeup"
)

// Returns the list of commands required by the klutch demo, and helper text for installing them.
// TODO: rethink this construct / generalize so each command in the CLI is consistent.
// Documentation should already contain instructions and links to the required tools,
// therefore also adding them here creates a drift risk.
func requiredCommands() map[string]map[string]string {
	cmds := make(map[string]map[string]string)

	cmds["git"] = make(map[string]string)
	cmds["git"]["darwin"] = "brew install git"

	cmds["docker"] = make(map[string]string)
	cmds["docker"]["darwin"] = "brew install docker"

	cmds["kind"] = make(map[string]string)
	cmds["kind"]["darwin"] = "brew install kind"

	cmds["kubectl"] = make(map[string]string)
	cmds["kubectl"]["darwin"] = "brew install kubectl"

	cmds["crossplane"] = make(map[string]string)
	cmds["crossplane"]["darwin"] = "curl -sL \"https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh\" | sh"
	cmds["crossplane"]["linux"] = "curl -sL \"https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh\" | sh"

	cmds["helm"] = make(map[string]string)
	cmds["helm"]["darwin"] = "brew install helm"

	cmds["kubectl-bind"] = make(map[string]string)
	cmds["kubectl-bind"]["darwin"] = "https://docs.k8s.anynines.com/docs/develop/platform-operator/central-management-cluster-setup/#binding-a-consumer-cluster-interactive"
	cmds["crossplane"]["linux"] = "https://docs.k8s.anynines.com/docs/develop/platform-operator/central-management-cluster-setup/#binding-a-consumer-cluster-interactive"

	cmds["cmctl"] = make(map[string]string)
	cmds["cmctl"]["darwin"] = "brew install cmctl"

	return cmds
}

// Checks if all commands defined by RequiredCommands are present.
// TODO: nearly identical to a8s demo code.
func checkKlutchDemoPrerequisites() {
	makeup.PrintH1("Checking Prerequisites...")

	requiredCmds := requiredCommands()
	allCommandsPresent := true
	for cmdName := range requiredCmds {
		path, err := exec.LookPath(cmdName)
		if err != nil {
			msg := "Couldn't find " + cmdName + " command: " + err.Error() + "."

			if requiredCmds[cmdName][runtime.GOOS] != "" {
				msg += " To install it, try: " + requiredCmds[cmdName][runtime.GOOS]
			}

			makeup.PrintFail(msg)
			allCommandsPresent = false
		}

		makeup.PrintCheckmark("Found " + cmdName + " at path " + path + ".")
	}

	if !allCommandsPresent {
		makeup.PrintFailSummary("Sadly, mandatory commands are missing. Aborting...")
		os.Exit(1)
	}

	if !isDockerRunning() {
		makeup.PrintFailSummary("Docker must be running for the klutch demo. Aborting...")
		os.Exit(1)
	}
}

func isDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}
