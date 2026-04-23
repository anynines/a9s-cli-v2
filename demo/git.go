package demo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	if !repoExists(localDirectory) {
		// No existing repo – clone fresh.
		if err := os.MkdirAll(localDirectory, os.ModePerm); err != nil {
			makeup.ExitDueToFatalError(err, "Couldn't create local directory to clone repository at "+localDirectory+".")
		}

		args := []string{"clone"}
		if tag != "latest" {
			args = append(args, "--branch", tag)
		}
		args = append(args, repositoryURL, localDirectory)

		output, err := makeup.Command("git", args...).WithPrompt().Run()
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to clone the git repository:\n"+string(output))
		}

		return
	}

	makeup.PrintInfo("Found existing repository at " + localDirectory + ", checking for local changes...")

	// Fail fast if the working tree is dirty to avoid silently discarding user changes.
	output, err := makeup.Command("git", "-C", localDirectory, "status", "--porcelain").Run()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to check git status of "+localDirectory)
	}
	if strings.TrimSpace(string(output)) != "" {
		makeup.ExitDueToFatalError(
			fmt.Errorf("uncommitted changes detected"),
			"The repository at "+localDirectory+" has local modifications. "+
				"Please commit, stash, or discard your changes before proceeding:\n"+string(output),
		)
	}

	makeup.PrintInfo("Fetching latest refs from remote...")
	if output, err := makeup.Command("git", "-C", localDirectory, "fetch", "--tags", "--force", "origin").WithPrompt().Run(); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to fetch from remote:\n"+string(output))
	}

	// Resolve the desired ref: for "latest" use the default remote branch HEAD.
	ref := tag
	if ref == "latest" {
		ref = "origin/HEAD"
	}

	makeup.PrintInfo("Checking out " + ref + "...")
	if output, err := makeup.Command("git", "-C", localDirectory, "checkout", ref).WithPrompt().Run(); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to checkout "+ref+":\n"+string(output))
	}

	// If we are on a branch (not detached HEAD), pull to fast-forward.
	branchOut, err := makeup.Command("git", "-C", localDirectory, "symbolic-ref", "--short", "HEAD").Run()
	if err == nil && strings.TrimSpace(string(branchOut)) != "" {
		makeup.PrintInfo("Pulling latest changes for branch " + strings.TrimSpace(string(branchOut)) + "...")
		if output, err := makeup.Command("git", "-C", localDirectory, "pull", "--ff-only").WithPrompt().Run(); err != nil {
			makeup.ExitDueToFatalError(err, "Failed to pull latest changes:\n"+string(output))
		}
	}
}

// repoExists returns true when localDirectory contains a git repository.
func repoExists(localDirectory string) bool {
	_, err := os.Stat(localDirectory)
	if err == nil {
		info, err := os.Stat(filepath.Join(localDirectory, ".git"))
		if !os.IsNotExist(err) {
			makeup.ExitDueToFatalError(err, "Failed to check for .git directory")
		}

		return err == nil && info.IsDir()
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
