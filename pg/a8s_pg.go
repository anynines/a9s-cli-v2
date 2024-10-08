package pg

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"gopkg.in/yaml.v2"
	m1u "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const DefaultPGAPIVersion = "v1beta3"

// Backup
const A8sPGBackupAPIGroup = "backups.anynines.com"
const A8sPGBackupKind = "Backup"
const A8sPGBackupKindPlural = "backups"

// Restore
const A8sPGRestoreKind = "Restore"

// Instance
const A8sPGServiceInstanceAPIGroup = "postgresql.anynines.com"
const A8sPGServiceInstanceAPIGroupLabel = "a8s.a9s/dsi-group=" + A8sPGServiceInstanceAPIGroup
const A8sPGServiceInstanceKind = "Postgresql"
const A8sPGServiceInstanceKindPlural = "postgresqls"
const A8sPGServiceInstanceKindLabel = "a8s.a9s/dsi-kind=" + A8sPGServiceInstanceKind

const A8sPGServiceInstanceNameLabelKey = "a8s.a9s/dsi-name"
const A8sPGLabelPrimary = "a8s.a9s/replication-role=master"

// Service Binding
const A8sPGServiceBindingKind = "ServiceBinding"
const A8sPGServiceBindingAPIGroup = "servicebindings.anynines.com"

type ServiceInstance struct {
	Kind         string
	ApiVersion   string
	Name         string
	Namespace    string
	Replicas     int
	VolumeSize   string
	Version      string
	RequestsCPU  string
	LimitsMemory string
}

type Backup struct {
	ApiVersion          string
	Name                string
	Namespace           string
	ServiceInstanceName string
}

type Restore struct {
	ApiVersion          string
	Name                string
	BackupName          string
	Namespace           string
	ServiceInstanceName string
}

type ServiceBinding struct {
	ApiVersion          string
	Name                string
	Namespace           string
	ServiceInstanceName string
	ServiceInstanceKind string
}

type PgManager struct {
	K8s *k8s.KubeClient
}

// NewPgManager returns a PgManager for the given kube context. If the context is empty,
// it is ignored.
func NewPgManager(kubeContext string) *PgManager {
	return &PgManager{
		K8s: k8s.NewKubeClient(kubeContext),
	}
}

func ServiceInstanceToYAML(instance ServiceInstance) string {
	instanceMap := make(map[string]interface{})
	instanceMap["apiVersion"] = A8sPGServiceInstanceAPIGroup + "/" + instance.ApiVersion
	instanceMap["kind"] = instance.Kind

	metadata := make(map[string]interface{})
	instanceMap["metadata"] = metadata
	metadata["name"] = instance.Name
	metadata["namespace"] = instance.Namespace

	spec := make(map[string]interface{})
	instanceMap["spec"] = spec
	spec["replicas"] = instance.Replicas
	spec["volumeSize"] = instance.VolumeSize
	spec["version"], _ = strconv.Atoi(instance.Version)

	resources := make(map[string]interface{})
	spec["resources"] = resources

	requests := make(map[string]interface{})
	resources["requests"] = requests
	requests["cpu"] = instance.RequestsCPU

	limits := make(map[string]interface{})
	resources["limits"] = limits
	limits["memory"] = instance.LimitsMemory

	yamlBytes, err := yaml.Marshal(instanceMap)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't generate YAML for service instance: %v", instance))
	}

	return string(yamlBytes)
}

/*
Creates a backup YAML manifest for the given service instance name.

Returns a string.
*/
func BackupToYAML(backup Backup) string {
	backupMap := make(map[string]interface{})
	backupMap["apiVersion"] = A8sPGBackupAPIGroup + "/" + backup.ApiVersion
	backupMap["kind"] = A8sPGBackupKind

	metadata := make(map[string]interface{})
	backupMap["metadata"] = metadata
	metadata["name"] = backup.Name
	metadata["namespace"] = backup.Namespace

	spec := make(map[string]interface{})
	backupMap["spec"] = spec

	serviceInstanceMap := make(map[string]interface{})
	spec["serviceInstance"] = serviceInstanceMap

	serviceInstanceMap["apiGroup"] = A8sPGServiceInstanceAPIGroup
	serviceInstanceMap["kind"] = A8sPGServiceInstanceKind
	serviceInstanceMap["name"] = backup.ServiceInstanceName

	yamlBytes, err := yaml.Marshal(backupMap)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't generate YAML for a8s Postgres backup with name: %s for service instance: %s", backup.Name, backup.ServiceInstanceName))
	}

	return string(yamlBytes)
}

/*
Creates a restore YAML manifest to restore a backup.
*/
func RestoreToYAML(restore Restore) string {
	restoreMap := make(map[string]interface{})
	restoreMap["apiVersion"] = A8sPGBackupAPIGroup + "/" + restore.ApiVersion
	restoreMap["kind"] = A8sPGRestoreKind

	metadata := make(map[string]interface{})
	restoreMap["metadata"] = metadata
	metadata["name"] = restore.Name
	metadata["namespace"] = restore.Namespace

	spec := make(map[string]interface{})
	restoreMap["spec"] = spec

	serviceInstanceMap := make(map[string]interface{})
	spec["serviceInstance"] = serviceInstanceMap

	serviceInstanceMap["apiGroup"] = A8sPGServiceInstanceAPIGroup
	serviceInstanceMap["kind"] = A8sPGServiceInstanceKind
	serviceInstanceMap["name"] = restore.ServiceInstanceName

	spec["backupName"] = restore.BackupName

	yamlBytes, err := yaml.Marshal(restoreMap)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't generate YAML for a8s Postgres restore with name: %s for backup %s of service instance: %s", restore.Name, restore.BackupName, restore.ServiceInstanceName))
	}

	return string(yamlBytes)
}

func ServiceBindingToYAML(binding ServiceBinding) string {
	bindingMap := make(map[string]interface{})
	bindingMap["apiVersion"] = A8sPGServiceBindingAPIGroup + "/" + binding.ApiVersion
	bindingMap["kind"] = A8sPGServiceBindingKind

	metadata := make(map[string]interface{})
	bindingMap["metadata"] = metadata
	metadata["name"] = binding.Name
	metadata["namespace"] = binding.Namespace

	spec := make(map[string]interface{})
	bindingMap["spec"] = spec

	instanceMap := make(map[string]interface{})
	spec["instance"] = instanceMap

	// Assumption: apiVersion of a service binding and its service instance is always the same
	instanceMap["apiVersion"] = A8sPGServiceInstanceAPIGroup + "/" + binding.ApiVersion
	instanceMap["kind"] = binding.ServiceInstanceKind
	instanceMap["name"] = binding.ServiceInstanceName
	instanceMap["namespace"] = binding.Namespace

	yamlBytes, err := yaml.Marshal(bindingMap)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't generate YAML for service binding: %v", binding))
	}

	return string(yamlBytes)
}

/*
Similar to kubectl watch ...
Needed: Namespace and name of the resource to be observed.
*/
func (pgm *PgManager) WaitForPGBackupResourceToBecomeReady(namespace, name string, resource string) error {
	makeup.PrintWait(fmt.Sprintf("Waiting for %s to become ready...", resource))

	//TODO Get API Group and API Version from a constant or the backup object
	// e.g. by making them separate fields in the BackupObject and make APIVersion an function
	gvr := schema.GroupVersionResource{Group: "backups.anynines.com", Version: DefaultPGAPIVersion, Resource: resource}

	desiredConditionsMap := make(map[string]interface{})
	desiredConditionsMap["type"] = "Complete"
	desiredConditionsMap["status"] = "True"

	failedConditionsMap := make(map[string]interface{})
	failedConditionsMap["type"] = "PermanentlyFailed"
	failedConditionsMap["status"] = "True"

	err := pgm.K8s.WaitForKubernetesResource(namespace, name, gvr, desiredConditionsMap, failedConditionsMap)

	return err
}

func (pgm *PgManager) WaitForPGBackupToBecomeReady(namespace, name string) {
	err := pgm.WaitForPGBackupResourceToBecomeReady(namespace, name, "backups")

	if err != nil {
		makeup.PrintFail("The backup has not been successful: " + err.Error() + ". Does the service instance exist?")
	} else {
		makeup.PrintCheckmark(fmt.Sprintf("The backup with the name %s in namespace %s has been successful.", name, namespace))
	}
}

func (pgm *PgManager) WaitForPGRestoreToBecomeReady(namespace, name string) {
	err := pgm.WaitForPGBackupResourceToBecomeReady(namespace, name, "restores")

	if err != nil {
		makeup.PrintFail("The restore has not been completed. Does the service instance and backup exist?")
	} else {
		makeup.PrintCheckmark(fmt.Sprintf("The restore with the name %s in namespace %s has been successful.", name, namespace))
	}
}

func (pgm *PgManager) DoesServiceInstanceExist(namespace, name string) bool {
	// Ignore the Don't Execute flag
	unattendedMode := true

	commandElements := make([]string, 0)
	commandElements = append(commandElements, "get")
	commandElements = append(commandElements, A8sPGServiceInstanceKindPlural)

	// Namespace
	commandElements = append(commandElements, "-n")
	commandElements = append(commandElements, namespace)

	// Output jsonpath
	commandElements = append(commandElements, "-o=jsonpath={.items[*].metadata.name}")

	cmd, output, err := pgm.K8s.Kubectl(unattendedMode, commandElements...)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl using the command: "+cmd.String())
	}

	outputString := string(output)

	if outputString == "" {
		return false
	}

	instanceNames := strings.Fields(outputString)

	return slices.Contains(instanceNames, name)
}

func (pgm *PgManager) DoesBackupExist(namespace, backupName string) bool {
	// Ignore the Don't Execute flag
	unattendedMode := true

	commandElements := make([]string, 0)
	commandElements = append(commandElements, "get")
	commandElements = append(commandElements, A8sPGBackupKindPlural)

	// Namespace
	commandElements = append(commandElements, "-n")
	commandElements = append(commandElements, namespace)

	// Output jsonpath
	commandElements = append(commandElements, "-o=jsonpath={.items[*].metadata.name}")

	cmd, output, err := pgm.K8s.Kubectl(unattendedMode, commandElements...)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl using the command: "+cmd.String())
	}

	outputString := string(output)

	if outputString == "" {
		return false
	}

	instanceNames := strings.Fields(outputString)

	return slices.Contains(instanceNames, backupName)
}

func (pgm *PgManager) WaitForPGServiceBindingToBecomeReady(binding ServiceBinding) error {
	makeup.PrintWait("Waiting for service binding to become ready...")
	gvr := schema.GroupVersionResource{Group: A8sPGServiceBindingAPIGroup, Version: DefaultPGAPIVersion, Resource: "servicebindings"}

	desiredConditionsMap := make(map[string]interface{})
	desiredConditionsMap["type"] = "Complete"
	desiredConditionsMap["status"] = "True"

	failedConditionsMap := make(map[string]interface{})
	failedConditionsMap["type"] = "PermanentlyFailed"
	failedConditionsMap["status"] = "True"

	err := pgm.K8s.WaitForKubernetesResourceWithFunction(binding.Namespace, binding.Name, gvr, func(object *m1u.Unstructured) bool {
		statusImplmentedInterface, exists, err := m1u.NestedFieldCopy(object.Object, "status", "implemented")

		makeup.PrintVerbose(fmt.Sprintf("Status implemented interface: %v", statusImplmentedInterface))

		if err != nil && !exists {
			// Wait longer
			makeup.PrintVerbose("There is not status, yet.")
			return true
		}

		if statusImplmentedInterface == nil {
			return true
		}

		statusImplemented := statusImplmentedInterface.(bool)
		return !statusImplemented
	})

	return err
}
