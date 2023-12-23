package pg

import (
	"fmt"
	"strconv"

	"github.com/anynines/a9s-cli-v2/makeup"
	"gopkg.in/yaml.v2"
)

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
	instanceMap["apiVersion"] = "postgresql.anynines.com/" + instance.ApiVersion
	instanceMap["kind"] = instance.Kind

	metadata := make(map[string]interface{})
	instanceMap["metadata"] = metadata
	metadata["name"] = instance.Name

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
