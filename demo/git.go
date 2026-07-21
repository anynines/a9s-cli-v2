package demo

import (
	"os"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/makeup"
)

func CheckoutDeploymentGitRepository() {
	makeup.PrintH1("Checking out git repository with a8s Data Service manifests...")
	makeup.Print("Remote Repository is at: " + demoGitRepo)
	makeup.Print("Local working dir: " + DemoConfig.WorkingDir)

	demoA8sDeploymentLocalFilepath := filepath.Join(DemoConfig.WorkingDir, demoA8sDeploymentLocalDir)

	CheckoutGitRepository(demoGitRepo, demoA8sDeploymentLocalFilepath, DeploymentVersion)
}

func CheckoutGitRepository(repositoryURL, localDirectory string, tag string) {
	// Check if the local directory already exists and remove it to ensure we have the correct version
	if repoExists(localDirectory) {
		makeup.PrintInfo("Removing existing a8s-deployment directory to ensure correct version is checked out...")
		err := os.RemoveAll(localDirectory)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to remove existing directory")
		}
	}
	// clone fresh
	if err := os.MkdirAll(localDirectory, os.ModePerm); err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't create local directory to clone repository at "+localDirectory+".")
	}

	args := []string{"clone"}
	if tag != "latest" {
		args = append(args, "--branch", tag)
	}
	args = append(args, repositoryURL, localDirectory)

	output, err := makeup.NewCommand("git", args...).WithPrompt().Run()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to clone the git repository:\n"+string(output))
	}
	makeup.PrintInfo("Successfully initialized git repository")
}

// repoExists returns true when localDirectory contains a git repository.
func repoExists(localDirectory string) bool {
	_, err := os.Stat(localDirectory)
	if err == nil {
		info, err := os.Stat(filepath.Join(localDirectory, ".git"))
		if err == nil {
			return info.IsDir()
		}
		if os.IsNotExist(err) {
			return false
		}
		makeup.ExitDueToFatalError(err, "Failed check for .git directory")
	}

	if os.IsNotExist(err) {
		return false
	}

	makeup.PrintInfo("Error while checking whether a8s-deployment is already cloned, removing existing a8s-deployment directory to ensure correct version is checked out...")

	err = os.RemoveAll(localDirectory)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to remove existing directory")
	}

	return false
}
