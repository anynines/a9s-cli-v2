package prerequisites

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/anynines/a9s-cli-v2/makeup"
)

const (
	GOOSDarwin = "darwin"
	GOOSLinux  = "linux"
)

// A RequiredTool represents a tool that must be present on the system.
// It includes information on how to install it for specific operating systems.
type RequiredTool struct {
	// CommandName is the command the tool is invoked with on the system.
	CommandName string
	// HelpCommands contains commands that help install the tool on a specific operating system.
	// e.g. HelpCommands["darwin"]
	HelpCommands map[string]string
	// HelpURLs contains URLs with more information on how to install the tool on a specific
	// operating system. e.g. HelpCommands["linux"]
	HelpURLs map[string]string
}

const (
	ToolGit        = "git"
	ToolDocker     = "docker"
	ToolKind       = "kind"
	ToolMinikube   = "minikube"
	ToolCmctl      = "cmctl"
	ToolKubectl    = "kubectl"
	ToolCrossplane = "crossplane"
	ToolHelm       = "helm"
	ToolBind       = "kubectl-bind"
)

// GetCommonRequiredTools returns a set of tools commonly required by different CLI commands.
// Each command can pick tools that it requires from this set.
func GetCommonRequiredTools() map[string]RequiredTool {
	return map[string]RequiredTool{
		ToolGit: {
			CommandName: ToolGit,
			HelpCommands: map[string]string{
				GOOSDarwin: "brew install git",
			},
			HelpURLs: map[string]string{
				GOOSLinux: "https://github.com/git-guides/install-git#install-git-on-linux",
			},
		},
		ToolDocker: {
			CommandName: ToolDocker,
			HelpCommands: map[string]string{
				GOOSDarwin: "brew install docker",
			},
			HelpURLs: map[string]string{
				GOOSLinux: "https://docs.docker.com/engine/install/",
			},
		},
		ToolKind: {
			CommandName: ToolKind,
			HelpCommands: map[string]string{
				GOOSDarwin: "brew install kind",
			},
			HelpURLs: map[string]string{
				GOOSLinux: "https://kind.sigs.k8s.io/docs/user/quick-start#installation",
			},
		},
		ToolMinikube: {
			CommandName: ToolMinikube,
			HelpCommands: map[string]string{
				GOOSDarwin: "brew install minikube",
			},
			HelpURLs: map[string]string{
				GOOSLinux: "https://minikube.sigs.k8s.io/docs/start",
			},
		},
		ToolCmctl: {
			CommandName: ToolCmctl,
			HelpCommands: map[string]string{
				GOOSDarwin: "brew install cmctl",
			},
			HelpURLs: map[string]string{
				GOOSLinux: "https://cert-manager.io/docs/reference/cmctl/#manual-installation",
			},
		},
		ToolKubectl: {
			CommandName: ToolKubectl,
			HelpCommands: map[string]string{
				GOOSDarwin: "brew install kubectl",
			},
			HelpURLs: map[string]string{
				GOOSLinux: "https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/",
			},
		},
		ToolCrossplane: {
			CommandName: ToolCrossplane,
			HelpURLs: map[string]string{
				GOOSDarwin: "https://docs.crossplane.io/latest/cli/",
				GOOSLinux:  "https://docs.crossplane.io/latest/cli/",
			},
		},
		ToolHelm: {
			CommandName: ToolHelm,
			HelpCommands: map[string]string{
				GOOSDarwin: "brew install helm",
			},
			HelpURLs: map[string]string{
				GOOSLinux: "https://helm.sh/docs/intro/install/",
			},
		},
		ToolBind: {
			CommandName: ToolBind,
			HelpURLs: map[string]string{
				GOOSDarwin: "https://docs.k8s.anynines.com/docs/develop/platform-operator/central-management-cluster-setup/#binding-a-consumer-cluster-interactive",
				GOOSLinux:  "https://docs.k8s.anynines.com/docs/develop/platform-operator/central-management-cluster-setup/#binding-a-consumer-cluster-interactive",
			},
		},
	}
}

// IsToolAvailable checks if a given tool is present on the system.
// If the tool is missing, info about the tool is printed.
// Returns true if the tool is present.
func IsToolAvailable(tool RequiredTool) bool {
	path, err := exec.LookPath(tool.CommandName)
	if err != nil {
		msg := "Couldn't find " + tool.CommandName + " command: " + err.Error() + "."

		if tool.HelpCommands[runtime.GOOS] != "" {
			msg += " Try running: " + tool.HelpCommands[runtime.GOOS]
		}

		if tool.HelpURLs[runtime.GOOS] != "" {
			msg += " See: " + tool.HelpURLs[runtime.GOOS]
		}

		makeup.PrintFail(msg)
		return false
	}

	makeup.PrintCheckmark("Found " + tool.CommandName + " at path " + path + ".")
	return true
}

// CheckRequiredTools checks if each given tool is present on the system.
// If a tool is missing, info about the tool is printed. At the end, if any tool is missing,
// the program exits with an error.
func CheckRequiredTools(requiredTools []RequiredTool) {
	allToolsPresent := true
	for _, tool := range requiredTools {
		if !IsToolAvailable(tool) {
			allToolsPresent = false
		}
	}

	if !allToolsPresent {
		makeup.PrintFailSummary("Sadly, mandatory commands are missing. Aborting...")
		os.Exit(1)
	}

	makeup.PrintSuccessSummary("All necessary commands are present.")
}

// CheckDockerRunning checks if docker is running. If it is not running, exit the program with an
// error.
func CheckDockerRunning() {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	if err != nil {
		makeup.PrintFailSummary("Docker must be running. Aborting...")
		os.Exit(1)
	}
}
