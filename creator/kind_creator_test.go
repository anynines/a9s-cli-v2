package creator

/*
Execute with: go test ./...
*/

import (
	"os"
	"testing"
	"time"
)

func BuildStandardClusterSpec() KubernetesClusterSpec {
	clusterName := "a8s-test-creator"

	return KubernetesClusterSpec{
		Name:                 clusterName,
		NodeMemory:           "2gb",
		NrOfNodes:            1,
		InfrastructureRegion: "eu-central-1",
	}
}

func TestKindExists(t *testing.T) {
	// spec := buildStandardClusterSpec()
	// c := MinikubeCreator{}

	// // c.Create(spec, true)

	// if !c.Exists(spec.Name) {
	// 	t.Fatal("Cluster with name " + spec.Name + " should exist but doesn't.")
	// }
}

func TestKindRunning(t *testing.T) {

}

/*
TODO Find a more elegant solution to perform end to end tests.
RUN: go test ./... -run TestKindCreate -v -count=1
Note: -cound=1 disables test caching
*/
func TestKindCreate(t *testing.T) {
	if !testing.Short() {
		spec := BuildStandardClusterSpec()
		spec.Name = "a8s-test-kind-creator"

		c := KindCreator{}

		c.LocalWorkDir = os.TempDir()

		c.Create(spec, true)

		time.Sleep(30 * time.Second)

		if !c.Exists(spec.Name) {
			t.Fatal("Cluster with name " + spec.Name + " should exist but doesn't.")
		}

		if !c.Running(spec.Name) {
			t.Fatal("Cluster with name " + spec.Name + " should be running but isn't.")
		}

		c.Delete(spec.Name, true)

		if c.Exists(spec.Name) {
			t.Fatal("Cluster with name " + spec.Name + " shouldn't exist but does.")
		}

		if c.Running(spec.Name) {
			t.Fatal("Cluster with name " + spec.Name + " shouldn't be running but is.")
		}
	} else {
		t.Skip("Skipping creation in short mode")
	}
}
