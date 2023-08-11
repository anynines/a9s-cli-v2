package main

/*
Next:

Ask for details for backup store config instead of using defaults.


TODO:


- Use Cases:
	- Pre-Create
		- Create S3 bucket with configs
	- Create
		- waitForA8sToBecomeReady
	- Delete
		- Remove cluster
		- Remove everything (incl. config files)
*/

import (
	"os"

	"github.com/fischerjulian/a8s-demo/demo"
)

var debug bool

func main() {

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	demo.PrintWelcomeScreen()

	demo.EstablishConfigFilePath()

	if !demo.LoadConfig() {
		demo.EstablishWorkingDir()
	}

	demo.CheckPrerequisites()

	demo.WaitForUser()

	demo.CheckoutDeploymentGitRepository()

	if demo.CountPodsInDemoNamespace() == 0 {
		demo.PrintCheckmark("Kubernetes cluster has no pods in " + demo.GetConfig().DemoSpace + " namespace.")
	}

	demo.EstablishBackupStoreCredentials()

	demo.ApplyCertManagerManifests()

	demo.ApplyA8sManifests()

	demo.PrintDemoSummary()
}
