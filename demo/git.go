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

func CheckoutDemoAppGitRepository() {
	makeup.PrintH1("Checking out git repository with demo application manifests...")
	makeup.Print("Remote Repository is at: " + demoAppGitRepo)
	makeup.Print("Local working dir: " + DemoConfig.WorkingDir)

	demoAppLocalFilepath := filepath.Join(DemoConfig.WorkingDir, demoAppLocalDir)

	//TODO Introduce releases for the demo app
	CheckoutGitRepository(demoAppGitRepo, demoAppLocalFilepath, "latest")
}

func CheckoutGitRepository(repositoryURL, localDirectory string, tag string) error {
	// Check if the local directory already exists
	/*
		If the target directory already exists and is non-empty, git clone will fail with: "already exists and is not an empty directory."
		However, assuming that a non-existing directory is healthy would be naive as the directory may be incomplete,
		e.g. due to a cancellation of a previous run.
	*/

	// if _, err := os.Stat(localDirectory); !os.IsNotExist(err) {
	// 	makeup.ExitDueToFatalError(err, "Can't checkout git repo "+repositoryURL+": The target directory: "+localDirectory+" does not exist.")
	// 	//return fmt.Errorf("local directory already exists")
	// }

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
