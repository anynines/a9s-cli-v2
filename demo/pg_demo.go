package demo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/a9s-cli-v2/pg"
)

//TODO Separate generic, non-pg methods into a separate file

// Settings
// TODO make configurable / cli param
const configFileName = ".a8s"
const demoGitRepo = "https://github.com/anynines/a8s-deployment.git" // "git@github.com:anynines/a8s-deployment.git"
const demoAppGitRepo = "https://github.com/anynines/a8s-demo.git"
const demoAppLocalDir = "a8s-demo"
const demoA8sDeploymentLocalDir = "a8s-deployment"

const certManagerNamespace = "cert-manager"
const certManagerManifestUrl = "https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml"
const defaultDemoSpace = "default"
const A8sSystemName = "a8s Postgres Control Plane"
const A8sSystemNamespace = "a8s-system"

var BackupInfrastructureProvider string // e.g. AWS
var BackupInfrastructureRegion string   // e.g. us-east-1
var BackupInfrastructureBucket string   // e.g. a8s-backups

var A8sPGServiceInstance pg.ServiceInstance
var DeleteA8sPGInstanceName string

var DeploymentVersion string // e.g. v0.3.0
var NoPreCheck bool          // e.g. false -> Perform prechecks
var DoNotApply bool          // e.g. yes --> do not execute kubectl apply -f ...
var KubernetesTool string    // e.g. "minikube" or "kind"

// const default_waiting_time_in_s = 10

type Config struct {
	WorkingDir string `yaml:"WorkingDir"`
	DemoSpace  string `yaml:"DemoSpace"`
}

type BlobStore struct {
	Config BlobStoreConfig `yaml:"config"`
}

type BlobStoreConfig struct {
	CloudConfig BlobStoreCloudConfiguration `yaml:"cloud_configuration"`
}

type BlobStoreCloudConfiguration struct {
	Provider  string `yaml:"provider"`
	Container string `yaml:"container"`
	Region    string `yaml:"region"`
}

var configFilePath string
var DemoConfig Config

func CheckPrerequisites() {
	allGood := true

	makeup.PrintH1("Checking Prerequisites...")

	CheckCommandAvailability()

	if !checkIfDockerIsRunning() {
		allGood = false
	}

	if !checkIfKubernetesIsRunning() {
		allGood = false
	}

	k8sCreator := GetKubernetesCreator()

	if !k8sCreator.Exists(DemoClusterName) {

		spec := BuildKubernetesClusterSpec()

		k8sCreator.Create(spec, UnattendedMode)

		fmt.Println()
		makeup.PrintH2("Rerunning prerequisite check ...")
		CheckPrerequisites()
		allGood = true
	} else {

		// Only if there's a suitable cluster, the cluster may also be selected.
		// In any other case, the demo cluster has to be created, first.
		CheckSelectedCluster()
	}

	// !NoPreCheck > Perform a pre-check
	if !NoPreCheck && !allGood {
		makeup.PrintFailSummary("Sadly, mandatory prerequisited haven't been met. Aborting...")
		os.Exit(1)
	}
}

/*
Applies the manifests of the a8s-deployment repository and thus installs a8s PG.
*/
func ApplyA8sManifests() {
	makeup.PrintH1("Applying the a8s Data Service manifests...")
	kustomizePath := filepath.Join(DemoConfig.WorkingDir, demoA8sDeploymentLocalDir, "deploy", "a8s", "manifests")
	KubectlApplyKustomize(kustomizePath)
	makeup.PrintCheckmark("Done applying a8s manifests.")
}

/*
Represents the state of a Pod which is expected to be running at some point.
The attribute "Running" is meant to be updated by a control loop.
*/
type PodExpectationState struct {
	Name    string
	Running bool
}

func WaitForA8sSystemToBecomeReady() {
	expectedPods := []PodExpectationState{
		{Name: "a8s-backup-controller-manager", Running: false},
		{Name: "postgresql-controller-manager", Running: false},
		{Name: "service-binding-controller-manager", Running: false},
	}

	WaitForSystemToBecomeReady(A8sSystemNamespace, A8sSystemName, expectedPods)
}

func WaitForServiceInstanceToBecomeReady(namespace, serviceInstanceName string, nrOfInstances int) {
	expectedPods := make([]PodExpectationState, 3)

	for i := 0; i < nrOfInstances; i++ {
		expectedPods[i] = PodExpectationState{
			Name:    fmt.Sprintf("%s-%d", serviceInstanceName, i),
			Running: false,
		}
	}

	WaitForSystemToBecomeReady(namespace, serviceInstanceName, expectedPods)
}

func CreatePGServiceInstance() {
	makeup.PrintH1("Creating a a8s Postgres Service Instance...")

	EstablishConfigFilePath()

	if !LoadConfig() {
		makeup.ExitDueToFatalError(nil, "There is no config, yet. Please create a demo environment before attempting to create a service instance.")
	}

	// TODO Find a more elegant way/place for setting the Kind attribute
	A8sPGServiceInstance.Kind = "Postgresql"
	instanceYAML := pg.ServiceInstanceToYAML(A8sPGServiceInstance)

	manifestsPath := UserManifestsPath()

	instanceManifestPath := filepath.Join(manifestsPath, A8sPGServiceInstance.Name+"-instance.yaml")

	err := os.WriteFile(instanceManifestPath, []byte(instanceYAML), 0600)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't save YAML file at: "+instanceManifestPath)
	}

	makeup.PrintInfo("The YAML manifest of the service instance is located at: " + instanceManifestPath)

	makeup.Print("The YAML manifest contains: ")
	err = makeup.PrintYAMLFile(instanceManifestPath)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't read service instance manifest from "+instanceManifestPath)
	}

	if !DoNotApply {
		KubectlApplyF(instanceManifestPath)
	}
}

// Refactor to DRY with Create ... > CRUDPGServiceInstance
func DeletePGServiceInstance() {
	makeup.PrintH1("Deleting a a8s Postgres Service Instance...")

	EstablishConfigFilePath()

	if !LoadConfig() {
		makeup.ExitDueToFatalError(nil, "There is no config, yet. Please create a demo environment before attempting to create a service instance.")
	}

	makeup.Print("Using default values for deleting the instance.")

	// // Stage 1: apply static manifest
	// // TODO stage 2: create struct, generate manifest based on parameters

	// exampleManifestPath := filepath.Join(A8sDeploymentExamplesPath(), "postgresql-instance.yaml")

	// makeup.PrintInfo("The YAML manifest of the service instance is located at: " + exampleManifestPath)

	// makeup.Print("The YAML manifest contains: ")
	// err := makeup.PrintYAMLFile(exampleManifestPath)

	// if err != nil {
	// 	makeup.ExitDueToFatalError(err, "Can't read service instance manifest from "+exampleManifestPath)
	// }

	KubectlAct("delete", "postgresqls", DeleteA8sPGInstanceName)
}

func PrintDemoSummary() {
	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Created a Kubernetes Cluster using " + KubernetesTool + " named: " + DemoClusterName + ".")
	makeup.PrintCheckmark("Installed cert-manager on the Kubernetes cluster.")
	makeup.PrintCheckmark("Created a configuration for the backup object store.")
	makeup.PrintCheckmark("Installing the a8s Postgres control plane.\n")

	//TODO Check whether Pods- from the a8s-system are ready
	//makeup.PrintCheckmark("Installed the a8s Postgres control plane.\n")
	makeup.PrintSuccessSummary("You are now ready to create a8s Postgres service instances.")
}
