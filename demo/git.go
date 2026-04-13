package demo

import (
	"fmt"
	"os"
	"os/exec"
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

func CheckoutGitRepository(repositoryURL, localDirectory string, tag string) error {
	// Check if the local directory already exists and remove it to ensure we have the correct version
	if _, err := os.Stat(localDirectory); !os.IsNotExist(err) {
		makeup.PrintInfo("Removing existing a8s-deployment directory to ensure correct version is checked out...")
		err := os.RemoveAll(localDirectory)
		if err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	var cmd *exec.Cmd

	err := os.MkdirAll(localDirectory, os.ModePerm)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't create local directory to clone demo-app repository at "+localDirectory+".")
	}

	// Run the git clone command to checkout the repository
	if tag == "latest" {
		cmd = exec.Command("git", "clone", repositoryURL, localDirectory)
	} else {
		cmd = exec.Command("git", "clone", "--branch", tag, repositoryURL, localDirectory)
	}

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(UnattendedMode)

	output, err := cmd.CombinedOutput()

	if err != nil {
		makeup.PrintFail("Failed to checkout the git repository: " + err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return err
	} else {

		fmt.Println(string(output))

		return nil
	}
}
