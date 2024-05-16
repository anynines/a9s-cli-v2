package k8s

import (
	"strings"
	"testing"
	"time"

	"github.com/anynines/a9s-cli-v2/creator"
)

const KubectlTestClusterName = "a8s-test-kubectl"

var KubectlTestCreator creator.KubernetesCreator

/*
Note: A relatively weak test as it is quite superficial.
But its tests whether kubectl is there and wether a trivial command succeeds.

Run with: go test ./... -run TestKubectl -v -count=1
*/
func TestKubectl(t *testing.T) {
	cmd, output, err := Kubectl(false, "--help")

	if err != nil {
		t.Fatalf("Couldn't execute \"kubectl --help\": %v", err)
	}

	expectedCmd := "kubectl --help"

	if !strings.Contains(cmd.String(), expectedCmd) {
		t.Fatalf("Kubectl did not issue the right kubectl command. Expected: %s but got %s", expectedCmd, cmd.String())
	}

	expectedOutput := "kubectl controls the Kubernetes cluster manager."
	outputString := string(output)
	if !strings.Contains(outputString, expectedOutput) {
		t.Fatalf("kubectl --help should contain \"%s\" but was: %s", outputString, expectedCmd)
	}
}

func TestFindFirstPodByLabel(t *testing.T) {
	nonExistingLabel := "non-existing-label=true"
	name, err := FindFirstPodByLabel("default", nonExistingLabel)
	if err != nil {
		if err != ErrNotFound {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	} else {
		t.Errorf("Shouldn't find a non-existing pod with label %s but found pod with name %s.", nonExistingLabel, name)
	}

	// Poor style
	time.Sleep(20 * time.Second)

	// Create pod with label
	knownLabel := "test-label=ihslsd"
	Kubectl(true, "run", "randomxxkdj", "--image=busybox", "--labels", knownLabel, "--", "sleep", "600")

	_, err = FindFirstPodByLabel("default", knownLabel)
	if err != nil {
		if err == ErrNotFound {
			t.Errorf("Should find a pod with label %s but didn't find any pod with that label.", knownLabel)
		} else {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	}
}

func BuildStandardClusterSpec() creator.KubernetesClusterSpec {
	clusterName := KubectlTestClusterName

	return creator.KubernetesClusterSpec{
		Name:                 clusterName,
		NodeMemory:           "2gb",
		NrOfNodes:            1,
		InfrastructureRegion: "eu-central-1",
	}
}

func getTestCreator() creator.KubernetesCreator {
	if KubectlTestCreator == nil {
		KubectlTestCreator = creator.MinikubeCreator{} // creator.KindCreator{LocalWorkDir: os.TempDir()}
	}

	return KubectlTestCreator
}

func CreateTestCluster() {
	spec := BuildStandardClusterSpec()
	c := getTestCreator()

	c.Create(spec, true)
}

func DeleteTestCluster() {
	c := getTestCreator()

	c.Delete(KubectlTestClusterName, true)
}

/*
Some of these tests require a Kubernetes cluster and thus we need to
perform some setup work before running the actual tests.

See: https://pkg.go.dev/testing#hdr-Main
*/
func TestMain(m *testing.M) {

	CreateTestCluster()
	defer DeleteTestCluster()

	m.Run()

}
