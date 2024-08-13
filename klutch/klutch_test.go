package klutch

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/anynines/a9s-cli-v2/demo"
	"gopkg.in/yaml.v2"
)

// Smoke test. Assumes the Deploy command has been run and the management cluster is up.
// Tests whether the individual components of the management cluster function to some degree.
// Run this test by using `A9S_CLI_KLUTCH_TEST=true go test -run TestDeploy ./klutch`
func TestDeploy(t *testing.T) {
	if os.Getenv("A9S_CLI_KLUTCH_TEST") != "true" {
		t.Skip("Skipping klutch smoke test; A9S_CLI_KLUTCH_TEST not set to true")
	}

	demo.EstablishConfig() // Makes sure the WorkingDir variable is set. TODO: this needs better handling
	configPath := filepath.Join(demo.DemoConfig.WorkingDir, mgmtClusterInfoFilePath, mgmtClusterInfoFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config file to be readable, got error: %v", err)
	}

	info := ManagementClusterInfo{}
	err = yaml.Unmarshal(data, &info)
	if err != nil {
		t.Fatalf("expected config data to be readable, got error: %v", err)
	}

	checkKindClusterRunning(t, mgmtClusterName)
	checkKindClusterRunning(t, consumerClusterName)
	checkBackendRunning(t, info)
	checkCrossplaneA8sRunning(t)
}

func checkKindClusterRunning(t *testing.T, clusterName string) {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", clusterName), "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected no error listing containers, but got error %v : %s", err, string(output))
	}

	if len(output) == 0 {
		t.Fatalf("expected the %s control plane container to be running, but none is", clusterName)
	}
}

func checkBackendRunning(t *testing.T, info ManagementClusterInfo) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s", info.Host, info.IngressPort))
	if err != nil {
		t.Fatalf("expected backend to be reachable, got error %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", resp.StatusCode)
	}
}

// Check crossplane and a8s by applying a postgresqlinstance claim.
func checkCrossplaneA8sRunning(t *testing.T) {
	claim := `apiVersion: anynines.com/v1
kind: PostgresqlInstance
metadata:
  name: klutch-test-pg
  namespace: default
spec:
  service: "a9s-postgresql13"
  plan: "postgresql-single-nano"
  expose: "Internal"
  compositionRef:
    name: a8s-postgresql
`

	in := bytes.NewBufferString(claim)
	cmd := exec.Command("kubectl", "apply", "--context", contextMgmt, "-f", "-")
	cmd.Stdin = in
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected no error while applying claim, but got %v : %s", err, string(output))
	}

	cmdWait := exec.Command("kubectl", "wait", "--context", contextMgmt, "--for=condition=ready", "postgresqlinstances", "klutch-test-pg", "--timeout=120s")
	output, err = cmdWait.CombinedOutput()
	if err != nil {
		t.Fatalf("expected no error while waiting for claim, but got %v : %s", err, string(output))
	}

	cmdWaitPg := exec.Command("kubectl", "wait", "--context", contextMgmt, "--for=condition=ready", "pod", "--selector", "a8s.a9s/dsi-name=klutch-test-pg", "--timeout=120s")
	output, err = cmdWaitPg.CombinedOutput()
	if err != nil {
		t.Fatalf("expected no error while waiting for pg instance, but got %v : %s", err, string(output))
	}

	in = bytes.NewBufferString(claim)
	cmdDelete := exec.Command("kubectl", "delete", "--timeout=120s", "--wait=false", "--context", contextMgmt, "-f", "-")
	cmdDelete.Stdin = in
	output, err = cmdDelete.CombinedOutput()
	if err != nil {
		t.Fatalf("expected no error while deleting claim, but got %v : %s", err, string(output))
	}
}
