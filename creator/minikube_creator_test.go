package creator

/*
Execute with: go test ./...
*/

import (
	"testing"
	"time"
)

func TestMinikubeExists(t *testing.T) {
	// spec := buildStandardClusterSpec()
	// c := MinikubeCreator{}

	// // c.Create(spec, true)

	// if !c.Exists(spec.Name) {
	// 	t.Fatal("Cluster with name " + spec.Name + " should exist but doesn't.")
	// }
}

func TestMinikubeRunning(t *testing.T) {

}

/*
TODO Find a more elegant solution to perform end to end tests.
*/
func TestMinikubeCreate(t *testing.T) {
	if !testing.Short() {
		spec := BuildStandardClusterSpec()
		spec.Name = "a8s-testing-minikube-creator"

		c := MinikubeCreator{}

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

func TestMinikubeDelete(t *testing.T) {

}
