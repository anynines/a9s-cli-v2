package k8s_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anynines/a9s-cli-v2/creator"
	"github.com/anynines/a9s-cli-v2/k8s"
)

// integrationEnabled returns true when the K8S_INTEGRATION=1 environment
// variable is set, gating tests that spin up real Kubernetes clusters.
func integrationEnabled() bool {
	return os.Getenv("K8S_INTEGRATION") == "1"
}

const KubectlTestClusterName = "a8s-test-kubectl"

var KubectlTestCreator creator.KubernetesCreator

/*
Note: A relatively weak test as it is quite superficial.
But its tests whether kubectl is there and wether a trivial command succeeds.

Run with: go test ./... -run TestKubectl -v -count=1
*/
func TestKubectl(t *testing.T) {
	output, err := k8s.Version(true, "")

	if err != nil {
		t.Fatalf("Couldn't execute \"kubectl --help\": %v", err)
	}

	expectedOutput := ""
	outputString := string(output)
	if !strings.Contains(outputString, expectedOutput) {
		t.Fatalf("kubectl --help should contain \"%s\" but was: %s", outputString, expectedOutput)
	}
}

func TestKubectlWithContext(t *testing.T) {
	if !integrationEnabled() {
		t.Skip("skipping integration test; set K8S_INTEGRATION=1 to run")
	}
	k := k8s.NewKubeClient("a8s-test-kubectl")
	output, err := k.Get("pod", "", "", "", true)

	if err != nil {
		t.Fatalf("Couldn't execute \"kubectl kubectl get pod --context a8s-test-kubectl\": %v : %v", err, string(output))
	}

	expectedOutput := "No resources found in default namespace."
	outputString := string(output)
	if !strings.Contains(outputString, expectedOutput) {
		t.Fatalf("kubectl get pod --context a8s-test-kubectl should contain \"%s\" but was: %s", outputString, expectedOutput)
	}
}

func TestFindFirstPodByLabel(t *testing.T) {
	if !integrationEnabled() {
		t.Skip("skipping integration test; set K8S_INTEGRATION=1 to run")
	}
	k := k8s.NewKubeClient("")
	nonExistingLabel := "non-existing-label=true"
	name, err := k.FindFirstPodByLabel("default", nonExistingLabel)
	if err != nil {
		if err != k8s.ErrNotFound {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	} else {
		t.Errorf("Shouldn't find a non-existing pod with label %s but found pod with name %s.", nonExistingLabel, name)
	}

	// Poor style
	time.Sleep(20 * time.Second)

	// Create pod with label
	knownLabel := "test-label=ihslsd"
	_, err = k.Run("", "randomxxkdj", "busybox", knownLabel, "sleep", "600")
	if err != nil {
		t.Errorf("Unexpected error while running the sleep command: %s", err.Error())
	}

	_, err = k.FindFirstPodByLabel("default", knownLabel)
	if err != nil {
		if err == k8s.ErrNotFound {
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
	if !integrationEnabled() {
		// Fast path: no cluster needed; unit tests run standalone.
		os.Exit(m.Run())
	}

	CreateTestCluster()
	defer DeleteTestCluster()

	m.Run()
}
