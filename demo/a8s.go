package demo

/*
	TODO Make a8s package installing the a8s demo (incl. future a8s data services)
*/

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/a9s-cli-v2/pg"
)

//TODO Separate generic, non-pg methods into a separate file

// Settings
// TODO There's clutter of package variables here. Reorganize these variables in to more meaningful
// structures.
const configFileName = ".a9s"
const demoGitRepo = "https://github.com/anynines/a8s-deployment.git" // "git@github.com:anynines/a8s-deployment.git"
const demoAppGitRepo = "https://github.com/anynines/a8s-demo.git"
const DemoAppLocalDir = "a8s-demo"
const demoA8sDeploymentLocalDir = "a8s-deployment"
const defaultWorkDir = "a9s" // $home/WorkDir as the default proposal for a work dir.

const defaultDemoSpace = "a8s-demo"
const A8sSystemName = "a8s Postgres Control Plane"
const A8sSystemNamespace = "a8s-system"

// TODO This is a poor man's struct!
var BackupInfrastructureProvider string // e.g. AWS
var BackupInfrastructureRegion string   // e.g. us-east-1
var BackupInfrastructureBucket string   // e.g. a8s-backups
var BackupInfrastructureEndpoint string // e.g. https://localhost:9000 for local minio
var BackupInfrastructurePathStyle bool  // e.g. false // Must be true for minio

var BackupStoreAccessKey string
var BackupStoreSecretKey string

var A8sPGServiceInstance pg.ServiceInstance
var DeleteA8sPGInstanceName string

var A8sPGBackup pg.Backup
var A8sPGRestore pg.Restore

var A8sPGServiceBinding pg.ServiceBinding

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
	Endpoint  string `yaml:"endpoint,omitempty"`
	PathStyle bool   `yaml:"path_style,omitempty"`
}

var configFilePath string
var DemoConfig Config

// TODO Why not merge some of these settings into DemoConfig?
// TODO Maybe DemoConfig or the entire demo package should be renamed a8s-demo or become less a8s specific
var DemoClusterName string
var UnattendedMode bool // Ask yes-no questions or assume "yes"
var ClusterNrOfNodes string
var ClusterMemory string

type A8sDemoManager struct {
	KubeContext string
	K8s         *k8s.KubeClient
	Pg          *pg.PgManager
}

// NewA8sDemoManager returns a A8sDemoManager for the given kube context. If the context is empty,
// it is ignored.
func NewA8sDemoManager(kubeContext string) *A8sDemoManager {
	return &A8sDemoManager{
		KubeContext: kubeContext,
		K8s:         k8s.NewKubeClient(kubeContext),
		Pg:          pg.NewPgManager(kubeContext),
	}
}

func (m *A8sDemoManager) CreateA8sStack(createClusterIfNotExists bool) {
	title := ""

	//TODO Tidy up
	if createClusterIfNotExists {
		title = "anynines Cluster Management"
	} else {
		title = "anynines Stack Management"
	}

	makeup.PrintWelcomeScreen(
		UnattendedMode,
		title,
		"Let's set up a Kubernetes stack together...")

	EstablishConfig()

	CheckPrerequisites()

	makeup.WaitForUser(UnattendedMode)

	//TODO It's odd that a check method also creates a k8s cluster
	CheckK8sCluster(createClusterIfNotExists)

	if m.CountPodsInDemoNamespace() == 0 {
		makeup.PrintCheckmark("Kubernetes cluster has no pods in " + GetConfig().DemoSpace + " namespace.")
	}

	m.DeployA8sStack()

	PrintDemoSummary()
}

// DeployA8sStack assumes that prerequisites are met. It was moved out of above function to allow re-use by other components.
func (m *A8sDemoManager) DeployA8sStack() {
	CheckoutDeploymentGitRepository()

	CheckoutDemoAppGitRepository()

	// TODO Refactor - See backlog "Refactor `EstablishBackupStoreCredentials`"
	EstablishBackupStoreCredentials()

	//TODO find a more elegant way to deal with minio
	if strings.ToLower(BackupInfrastructureProvider) == "minio" {
		m.ApplyMinioManifests()
		m.WaitForMinioToBecomeReady()
	}

	m.K8s.ApplyCertManagerManifests(UnattendedMode)

	m.ApplyA8sManifests()

	m.WaitForA8sSystemToBecomeReady()
}

/*
Applies the manifests of the a8s-deployment repository and thus installs a8s PG.
*/
func (m *A8sDemoManager) ApplyA8sManifests() {
	makeup.PrintH1("Applying the a8s Data Service manifests...")

	kustomizePath := filepath.Join(DemoConfig.WorkingDir, demoA8sDeploymentLocalDir, "deploy", "a8s", "manifests")

	m.K8s.KubectlApplyKustomize(kustomizePath, UnattendedMode)

	makeup.PrintCheckmark("Done applying a8s manifests.")
}

func (m *A8sDemoManager) WaitForA8sSystemToBecomeReady() {
	makeup.PrintH1("Waiting for the a8s System to become ready...")
	expectedPodsByLabels := []string{
		"app.kubernetes.io/name=backup-manager",                     // a8s-backup-controller-manager
		"app.kubernetes.io/name=postgresql-controller-manager",      // postgresql-controller-manager
		"app.kubernetes.io/name=service-binding-controller-manager", // service-binding-controller-manager
	}

	m.K8s.KubectlWaitForSystemToBecomeReady(A8sSystemNamespace, expectedPodsByLabels)

	makeup.PrintCheckmark("The a8s System appears to be ready.")
	makeup.WaitForUser(UnattendedMode)
}

func (m *A8sDemoManager) WaitForServiceInstanceToBecomeReady(namespace, serviceInstanceName string, nrOfInstances int) {
	expectedPods := make([]k8s.PodExpectationState, nrOfInstances)

	for i := 0; i < nrOfInstances; i++ {
		expectedPods[i] = k8s.PodExpectationState{
			Name:    fmt.Sprintf("%s-%d", serviceInstanceName, i),
			Running: false,
		}
	}

	m.K8s.WaitForSystemToBecomeReady(namespace, serviceInstanceName, expectedPods)

	makeup.WaitForUser(UnattendedMode)
}

func (m *A8sDemoManager) CreatePGServiceInstance() {
	makeup.PrintH1("Creating a a8s Postgres Service Instance...")

	EnsureConfigIsLoaded()

	// TODO Find a more elegant way/place for setting the Kind attribute
	A8sPGServiceInstance.Kind = "Postgresql"
	instanceYAML := pg.ServiceInstanceToYAML(A8sPGServiceInstance)

	instanceManifestPath := getServiceInstanceManifestPath(A8sPGServiceInstance.Name)

	WriteYAMLToFile(instanceYAML, instanceManifestPath)

	if !DoNotApply {
		m.K8s.KubectlApplyF(instanceManifestPath, UnattendedMode)
	}
}

/*
Returns the filepath to the service instance manifest.
*/
func getServiceInstanceManifestPath(serviceInstanceName string) string {
	return GetUserManifestPath("a8s-pg-instance-" + serviceInstanceName + ".yaml")
}

func (m *A8sDemoManager) DeletePGServiceInstance(namespace, serviceInstanceName string) {
	makeup.PrintH1("Deleting a a8s Postgres Service Instance...")

	EnsureConfigIsLoaded()

	if !m.Pg.DoesServiceInstanceExist(namespace, serviceInstanceName) {
		makeup.PrintWarning(fmt.Sprintf("Can't delete service instance. Service instance %s doesn't exist in namespace %s!", serviceInstanceName, namespace))
		os.Exit(0)
	}

	// TODO Make "postgresqls" a constant
	_, _, err := m.K8s.Kubectl(UnattendedMode, "delete", "postgresqls", serviceInstanceName, "-n", namespace)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't delete service instance.")
	} else {
		makeup.PrintCheckmark(fmt.Sprintf("Service instance %s successfully deleted from namespace %s.", serviceInstanceName, namespace))
	}
}

func (m *A8sDemoManager) CountPodsInDemoNamespace() int {
	return m.K8s.CountPodsInNamespace(DemoConfig.DemoSpace)
}

func getBackupManifestPath(backupName string) string {
	makeup.PrintVerbose("Generating manifest for backup: " + backupName + " ...")
	return GetUserManifestPath("a8s-pg-backup-" + backupName + ".yaml")
}

func getRestoreManifestPath(backupName string) string {
	makeup.PrintVerbose("Generating manifest for backup restore: " + backupName + " ...")

	//TODO their could be collisions for names used in mulitple namespace
	return GetUserManifestPath("a8s-pg-restore-" + backupName + ".yaml")
}

func (m *A8sDemoManager) CreatePGServiceInstanceBackup() {
	EnsureConfigIsLoaded()

	if !m.Pg.DoesServiceInstanceExist(A8sPGBackup.Namespace, A8sPGBackup.ServiceInstanceName) {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create backup for non-existing service instance %s in namespace %s", A8sPGBackup.ServiceInstanceName, A8sPGBackup.Namespace))
	}

	makeup.PrintH1("Creating an a8s Postgres service instance backup...")

	yaml := pg.BackupToYAML(A8sPGBackup)

	WriteYAMLToFile(yaml, getBackupManifestPath(A8sPGBackup.Name))

	if !DoNotApply {
		m.K8s.KubectlApplyF(getBackupManifestPath(A8sPGBackup.Name), UnattendedMode)
	}

	m.Pg.WaitForPGBackupToBecomeReady(A8sPGBackup.Namespace, A8sPGBackup.Name)
}

// TODO Reduce code duplicity with CreatePGServiceInstanceBackup
func (m *A8sDemoManager) CreatePGServiceInstanceRestore() {
	EnsureConfigIsLoaded()

	if !m.Pg.DoesServiceInstanceExist(A8sPGRestore.Namespace, A8sPGRestore.ServiceInstanceName) {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create restore for non-existing service instance %s in namespace %s", A8sPGRestore.ServiceInstanceName, A8sPGRestore.Namespace))
	}

	if !m.Pg.DoesBackupExist(A8sPGRestore.Namespace, A8sPGRestore.BackupName) {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create restore for non-existing backup %s in namespace %s", A8sPGRestore.BackupName, A8sPGRestore.Namespace))
	}

	makeup.PrintH1("Creating an a8s Postgres Service Instance Backup Restore...")

	yaml := pg.RestoreToYAML(A8sPGRestore)

	WriteYAMLToFile(yaml, getRestoreManifestPath(A8sPGRestore.Name))

	if !DoNotApply {
		m.K8s.KubectlApplyF(getRestoreManifestPath(A8sPGRestore.Name), UnattendedMode)
	}

	m.Pg.WaitForPGRestoreToBecomeReady(A8sPGRestore.Namespace, A8sPGRestore.Name)
}

func getServiceBindingManifestPath(binding pg.ServiceBinding) string {
	makeup.PrintVerbose("Generating manifest for service binding: " + binding.Name + " ...")

	// TODO their could be collisions for names used in mulitple namespace
	// ADD binding.Namespace + "-" +
	return GetUserManifestPath("a8s-pg-service-binding-" + binding.Name + ".yaml")
}

func (m *A8sDemoManager) CreatePGServiceBinding() {
	makeup.PrintH1("Creating a a8s Postgres Service Binding...")
	EnsureConfigIsLoaded()

	if !m.Pg.DoesServiceInstanceExist(A8sPGServiceBinding.Namespace, A8sPGServiceBinding.ServiceInstanceName) {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create service binding for non-existing service instance %s in namespace %s", A8sPGServiceBinding.ServiceInstanceName, A8sPGServiceBinding.Namespace))
	}

	yaml := pg.ServiceBindingToYAML(A8sPGServiceBinding)

	WriteYAMLToFile(yaml, getServiceBindingManifestPath(A8sPGServiceBinding))

	if !DoNotApply {
		m.K8s.KubectlApplyF(getServiceBindingManifestPath(A8sPGServiceBinding), UnattendedMode)
	}

	err := m.Pg.WaitForPGServiceBindingToBecomeReady(A8sPGServiceBinding)

	if err != nil {
		makeup.ExitDueToFatalError(err, "A problem occurred creating the service binding.")
	} else {
		makeup.PrintCheckmark("The service binding has been created successfully.")
	}
}

func (m *A8sDemoManager) DeletePGServiceBinding() {
	makeup.PrintH1("Deleting a a8s Postgres Service Binding...")
	EnsureConfigIsLoaded()

	_, _, err := m.K8s.Kubectl(UnattendedMode, "delete", "servicebinding", A8sPGServiceBinding.Name, "-n", A8sPGServiceBinding.Namespace)

	if err != nil {
		makeup.ExitDueToFatalError(err, "A problem occurred deleting the service binding.")
	} else {
		makeup.PrintCheckmark("The service binding has been deleted successfully.")
	}
}

func PrintDemoSummary() {
	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Created a Kubernetes Cluster using " + KubernetesTool + " named: " + DemoClusterName + ".")
	makeup.PrintCheckmark("Installed cert-manager on the Kubernetes cluster.")
	if strings.ToLower(BackupInfrastructureProvider) == "minio" {
		makeup.PrintCheckmark("Installed the MinIO object store.")
	}
	makeup.PrintCheckmark("Created a configuration for the backup object store.")
	makeup.PrintCheckmark("Installing the a8s Postgres control plane.\n")
	makeup.PrintSuccessSummary("You are now ready to create a8s Postgres service instances.")
}
