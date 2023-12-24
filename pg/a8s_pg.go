package pg

import (
	"fmt"
	"strconv"

	"github.com/anynines/a9s-cli-v2/makeup"
	"gopkg.in/yaml.v2"
)

const A8sPGServiceInstanceAPIGroup = "postgresql.anynines.com"
const A8sPGBackupAPIGroup = "backups.anynines.com"
const A8sPGBackupKind = "Backup"
const A8sPGServiceInstanceKind = "PostgreSQL"

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
func BackupToYAML(namespace, apiVersion, backupName, serviceInstanceName string) string {
	backupMap := make(map[string]interface{})
	backupMap["apiVersion"] = A8sPGBackupAPIGroup + "/" + apiVersion
	backupMap["kind"] = A8sPGBackupKind

	metadata := make(map[string]interface{})
	backupMap["metadata"] = metadata
	metadata["name"] = backupName
	metadata["namespace"] = namespace

	spec := make(map[string]interface{})
	backupMap["spec"] = spec

	serviceInstanceMap := make(map[string]interface{})
	spec["serviceInstance"] = serviceInstanceMap

	serviceInstanceMap["apiGroup"] = A8sPGServiceInstanceAPIGroup
	serviceInstanceMap["kind"] = A8sPGServiceInstanceKind
	serviceInstanceMap["name"] = serviceInstanceName

	yamlBytes, err := yaml.Marshal(backupMap)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't generate YAML for a8s Postgres backup with name: %s for service instance: %s", backupName, serviceInstanceName))
	}

	return string(yamlBytes)
}
