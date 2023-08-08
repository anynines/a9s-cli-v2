package main

/*
Next:

Ask for details for backup store config instead of using defaults.


TODO:

- Create S3 bucket with configs
- waitForA8sToBecomeReady

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

	// demo.CheckoutDeploymentGitRepository()

	// if demo.CountPodsInDemoNamespace() == 0 {
	// 	PrintCheckmark("Kubernetes cluster has no pods in " + demo.GetConfig().DemoSpace + " namespace.")
	// }

	// demo.EstablishBackupStoreCredentials()

	// demo.ApplyCertManagerManifests()

	// demo.ApplyA8sManifests()
}
