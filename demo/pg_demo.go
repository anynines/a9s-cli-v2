package demo

/*
	TODO Make a8s package installing the a8s demo (incl. future a8s data services)
*/

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/anynines/a9s-cli-v2/creator"
	"github.com/anynines/a9s-cli-v2/k8s"
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

const defaultDemoSpace = "default"
const A8sSystemName = "a8s Postgres Control Plane"
const A8sSystemNamespace = "a8s-system"

var BackupInfrastructureProvider string // e.g. AWS
var BackupInfrastructureRegion string   // e.g. us-east-1
var BackupInfrastructureBucket string   // e.g. a8s-backups

var A8sPGServiceInstance pg.ServiceInstance
var DeleteA8sPGInstanceName string

var A8sPGBackup pg.Backup
var A8sPGRestore pg.Restore

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

// TODO Why not merge some of these settings into DemoConfig?
// TODO Maybe DemoConfig or the entire demo package should be renamed a8s-demo or become less a8s specific
var DemoClusterName string
var UnattendedMode bool // Ask yes-no questions or assume "yes"
var ClusterNrOfNodes string
var ClusterMemory string

func BuildKubernetesClusterSpec() creator.KubernetesClusterSpec {
	nrOfNodes, err := strconv.Atoi(ClusterNrOfNodes)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't determine the number of Kubernetes nodes from the param: "+ClusterNrOfNodes)
	}

	spec := creator.KubernetesClusterSpec{
		Name:                 DemoClusterName,
		NrOfNodes:            nrOfNodes,
		NodeMemory:           ClusterMemory,
		InfrastructureRegion: BackupInfrastructureRegion,
	}

	return spec
}

/*
Builds a cluster manager without params by using shared package variables.
*/
func BuildKubernetesClusterManager() k8s.ClusterManager {

	clusterManager := k8s.BuildClusterManager(
		DemoConfig.WorkingDir,
		DemoClusterName,
		KubernetesTool,
		UnattendedMode,
	)

	return clusterManager
}

func CheckPrerequisites() {
	allGood := true

	makeup.PrintH1("Checking Prerequisites...")

	CheckCommandAvailability()

	//TODO Remove section
	// Docker is not a pre-requisite.
	// if !kubernetes.CheckIfDockerIsRunning() {
	// 	allGood = false
	// }

	// TODO Remove section
	// We don't need to check if Kubernetes is running as we are about to
	// create a Kubernetes cluster

	//k8sCreator := kubernetes.GetKubernetesCreator(KubernetesTool, DemoConfig.WorkingDir)

	clusterSpec := BuildKubernetesClusterSpec()

	clusterManager := BuildKubernetesClusterManager()

	clusterManager.CreateKubernetesClusterIfNotExists(clusterSpec)

	// At this point there should be a Kubernetescluster
	if !k8s.CheckIfAnyKubernetesIsRunning() {
		allGood = false
	}

	// Only if there's a suitable cluster, the cluster may also be selected.
	// In any other case, the demo cluster has to be created, first.
	if !clusterManager.IsClusterSelectedAsCurrentContext(DemoClusterName) {
		allGood = false
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
	k8s.KubectlApplyKustomize(kustomizePath, UnattendedMode)
	makeup.PrintCheckmark("Done applying a8s manifests.")
}

func WaitForA8sSystemToBecomeReady() {
	expectedPods := []k8s.PodExpectationState{
		{Name: "a8s-backup-controller-manager", Running: false},
		{Name: "postgresql-controller-manager", Running: false},
		{Name: "service-binding-controller-manager", Running: false},
	}

	k8s.WaitForSystemToBecomeReady(A8sSystemNamespace, A8sSystemName, expectedPods)
	makeup.WaitForUser(UnattendedMode)
}

func WaitForServiceInstanceToBecomeReady(namespace, serviceInstanceName string, nrOfInstances int) {
	expectedPods := make([]k8s.PodExpectationState, nrOfInstances)

	for i := 0; i < nrOfInstances; i++ {
		expectedPods[i] = k8s.PodExpectationState{
			Name:    fmt.Sprintf("%s-%d", serviceInstanceName, i),
			Running: false,
		}
	}

	k8s.WaitForSystemToBecomeReady(namespace, serviceInstanceName, expectedPods)
	makeup.WaitForUser(UnattendedMode)
}

// TODO Move to pg package
func CreatePGServiceInstance() {
	makeup.PrintH1("Creating a a8s Postgres Service Instance...")

	EnsureConfigIsLoaded()

	// TODO Find a more elegant way/place for setting the Kind attribute
	A8sPGServiceInstance.Kind = "Postgresql"
	instanceYAML := pg.ServiceInstanceToYAML(A8sPGServiceInstance)

	instanceManifestPath := getServiceInstanceManifestPath(A8sPGServiceInstance.Name)

	WriteYAMLToFile(instanceYAML, instanceManifestPath)

	if !DoNotApply {
		k8s.KubectlApplyF(instanceManifestPath, UnattendedMode)
	}
}

/*
Returns the filepath to the service instance manifest.
*/
func getServiceInstanceManifestPath(serviceInstanceName string) string {
	return GetUserManifestPath("a8s-pg-instance-" + serviceInstanceName + ".yaml")
}

// TODO Move to pg package
// Refactor to DRY with Create ... > CRUDPGServiceInstance
func DeletePGServiceInstance(namespace, serviceInstanceName string) {
	makeup.PrintH1("Deleting a a8s Postgres Service Instance...")

	EnsureConfigIsLoaded()

	if !pg.DoesServiceInstanceExist(namespace, serviceInstanceName) {
		makeup.PrintWarning(fmt.Sprintf("Can't delete service instance. Service instance %s doesn't exist in namespace %s!", serviceInstanceName, namespace))
		os.Exit(0)
	}

	// TODO Make "postgresqls" a constant
	_, _, err := k8s.Kubectl(UnattendedMode, "delete", "postgresqls", serviceInstanceName, "-n", namespace)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't delete service instance.")
	} else {
		makeup.PrintCheckmark(fmt.Sprintf("Service instance %s successfully deleted from namespace %s.", serviceInstanceName, namespace))
	}
}

func CountPodsInDemoNamespace() int {
	return k8s.CountPodsInNamespace(DemoConfig.DemoSpace)
}

func getBackupManifestPath(backupName string) string {
	makeup.PrintVerbose("Generating manifest for backup: " + backupName + " ...")
	return GetUserManifestPath("a8s-pg-backup-" + backupName + ".yaml")
}

func getRestoreManifestPath(backupName string) string {
	makeup.PrintVerbose("Generating manifest for backup restore: " + backupName + " ...")
	return GetUserManifestPath("a8s-pg-restore-" + backupName + ".yaml")
}

// TODO Move to pg package
func CreatePGServiceInstanceBackup() {
	EnsureConfigIsLoaded()

	if !pg.DoesServiceInstanceExist(A8sPGBackup.Namespace, A8sPGBackup.ServiceInstanceName) {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create backup for non-existing service instance %s in namespace %s", A8sPGBackup.ServiceInstanceName, A8sPGBackup.Namespace))
	}

	makeup.PrintH1("Creating an a8s Postgres service instance backup...")

	yaml := pg.BackupToYAML(A8sPGBackup)

	WriteYAMLToFile(yaml, getBackupManifestPath(A8sPGBackup.Name))

	if !DoNotApply {
		k8s.KubectlApplyF(getBackupManifestPath(A8sPGBackup.Name), UnattendedMode)
	}

	pg.WaitForPGBackupToBecomeReady(A8sPGBackup.Namespace, A8sPGBackup.Name)
}

// TODO Reduce code duplicity with CreatePGServiceInstanceBackup
func CreatePGServiceInstanceRestore() {
	EnsureConfigIsLoaded()

	if !pg.DoesServiceInstanceExist(A8sPGRestore.Namespace, A8sPGRestore.ServiceInstanceName) {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create restore for non-existing service instance %s in namespace %s", A8sPGBackup.ServiceInstanceName, A8sPGBackup.Namespace))
	}

	makeup.PrintH1("Creating an a8s Postgres Service Instance Backup Restore...")

	yaml := pg.RestoreToYAML(A8sPGRestore)

	WriteYAMLToFile(yaml, getRestoreManifestPath(A8sPGRestore.Name))

	if !DoNotApply {
		k8s.KubectlApplyF(getRestoreManifestPath(A8sPGRestore.Name), UnattendedMode)
	}

	pg.WaitForPGRestoreToBecomeReady(A8sPGRestore.Namespace, A8sPGRestore.Name)
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
