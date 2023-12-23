package pg

//  go test ./... -run TestServiceInstanceToYAML -v

import (
	"testing"
)

func TestServiceInstanceToYAML(t *testing.T) {

	instance := ServiceInstance{
		ApiVersion:   "postgresql.anynines.com/v1beta3",
		Name:         "sample-pg-cluster",
		Kind:         "Postgresql",
		Namespace:    "default",
		Replicas:     3,
		VolumeSize:   "1Gi",
		Version:      "14",
		RequestsCPU:  "100m",
		LimitsMemory: "100Mi",
	}

	yamlString := ServiceInstanceToYAML(instance)

	t.Logf("YAML String:\n%s", yamlString)

	//TODO Think of meaningful test
	// a) is it valid yaml?
	// b) marshal <> unmarshal > objects should be equal
	// c) is the right content available?
	//		test string with regexp >
	// d) is the spec valid given its apiversion & kind (Postgresql)
	//			test using kubectl validation ?! kubectl must know whether the given spec of that kind is valid
	// 			this doesn't say that the spec is correct but ... better than nothing
}
