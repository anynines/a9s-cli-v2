package pg

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const A8sPGServiceInstanceAPIGroup = "postgresql.anynines.com"
const A8sPGServiceInstanceAPIGroupLabel = "a8s.a9s/dsi-group=" + A8sPGServiceInstanceAPIGroup
const A8sPGBackupAPIGroup = "backups.anynines.com"
const A8sPGBackupKind = "Backup"
const A8sPGBackupKindPlural = "backups"
const A8sPGRestoreKind = "Restore"
const A8sPGServiceInstanceKind = "PostgreSQL"
const A8sPGServiceInstanceKindPlural = "postgresqls"
const A8sPGServiceInstanceKindLabel = "a8s.a9s/dsi-kind=" + A8sPGServiceInstanceKind
const A8sPGLabelPrimary = "a8s.a9s/replication-role=master"
const A8sPGServiceInstanceNameLabelKey = "a8s.a9s/dsi-name"

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

/*
Similar to kubectl watch ...
Needed: Namespace and name of the resource to be observed.
*/
func WaitForPGBackupResourceToBecomeReady(namespace, name string, resource string) error {

	//TODO Get API Group and API Version from a constant or the backup object
	// e.g. by making them separate fields in the BackupObject and make APIVersion an function
	gvr := schema.GroupVersionResource{Group: "backups.anynines.com", Version: "v1beta3", Resource: resource}

	desiredConditionsMap := make(map[string]interface{})
	desiredConditionsMap["type"] = "Complete"
	desiredConditionsMap["status"] = "True"

	err := k8s.WaitForKubernetesResource(namespace, gvr, desiredConditionsMap)

	return err
}

func WaitForPGBackupToBecomeReady(namespace, name string) {
	err := WaitForPGBackupResourceToBecomeReady(namespace, name, "backups")

	if err != nil {
		makeup.PrintFail("The backup has not been successful. Does the service instance exist?")
	} else {
		makeup.PrintCheckmark(fmt.Sprintf("The backup with the name %s in namespace %s has been successful.", name, namespace))
	}
}

func WaitForPGRestoreToBecomeReady(namespace, name string) {
	err := WaitForPGBackupResourceToBecomeReady(namespace, name, "restores")

	if err != nil {
		makeup.PrintFail("The restore has not been completed. Does the service instance and backup exist?")
	} else {
		makeup.PrintCheckmark(fmt.Sprintf("The restore with the name %s in namespace %s has been successful.", name, namespace))
	}
}

func DoesServiceInstanceExist(namespace, name string) bool {
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

	cmd, output, err := k8s.Kubectl(unattendedMode, commandElements...)

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

func DoesBackupExist(namespace, backupName string) bool {
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

	cmd, output, err := k8s.Kubectl(unattendedMode, commandElements...)

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
