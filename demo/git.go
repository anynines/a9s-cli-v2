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
	// Check if the local directory already exists
	/*
		If the target directory already exists and is non-empty, git clone will fail with: "already exists and is not an empty directory."
		However, assuming that a non-existing directory is healthy would be naive as the directory may be incomplete,
		e.g. due to a cancellation of a previous run.
	*/

	if _, err := os.Stat(localDirectory); !os.IsNotExist(err) {
		makeup.PrintInfo("The a8s-deployment directory already exists. Please verify that the directory is up to date and contents are healthy. If you are unsure, delete it. It'll will be cloned from the remote repository, again.")
		return
	}

	args := []string{}

	err := os.MkdirAll(localDirectory, os.ModePerm)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't create local directory to clone demo-app repository at "+localDirectory+".")
	}

	// Run the git clone command to checkout the repository
	if tag == "latest" {
		args = append(args, "clone", repositoryURL, localDirectory)
	} else {
		args = append(args, "clone", "--branch", tag, repositoryURL, localDirectory)
	}

	output, err := makeup.Command("git", args...).WithPrompt().Run()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to checkout the git repository:\n"+string(output))
	}
}
