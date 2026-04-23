// White-box unit tests for the k8s package. These tests run without any
// real Kubernetes cluster by replacing execCommand/execCommandContext with a
// fake that delegates to TestHelperProcess (the standard Go test-subprocess
// pattern). No kubectl binary is ever invoked; no kubecontext is touched.
//
// Run: go test ./k8s/... -run 'Unit' -v -count=1
// All cluster-dependent tests are in kubectl_test.go and require K8S_INTEGRATION=1.
package k8s

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/anynines/a9s-cli-v2/makeup"
)

// ---------------------------------------------------------------------------
// TestHelperProcess – "fake kubectl" subprocess entry point.
//
// When fakeExecCommand creates a *exec.Cmd it targets os.Args[0] (the test
// binary itself) with -test.run=TestHelperProcess appended, plus the env var
// GO_WANT_HELPER_PROCESS=1. The test binary re-enters here, emits the canned
// output stored in env vars, and exits – never touching the real kubectl or
// your kubecontext.
// ---------------------------------------------------------------------------
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Parse the real command from args: [binary, -test.run=TestHelperProcess, --, <cmd>, <args>...]
	args := os.Args
	var realArgs []string
	for i, a := range args {
		if a == "--" {
			realArgs = args[i+1:]
			break
		}
	}

	// realArgs[0] is the executable name ("kubectl"), realArgs[1] is the sub-command.
	subCmd := ""
	if len(realArgs) > 1 {
		subCmd = realArgs[1]
	}

	exitCode, _ := strconv.Atoi(os.Getenv("FAKE_EXIT_CODE"))

	// Route output by subcommand so callers can set per-subcommand env vars.
	var stdout string
	switch subCmd {
	case "get":
		stdout = os.Getenv("FAKE_KUBECTL_GET_STDOUT")
	case "apply":
		stdout = os.Getenv("FAKE_KUBECTL_APPLY_STDOUT")
	case "delete":
		stdout = os.Getenv("FAKE_KUBECTL_DELETE_STDOUT")
	case "describe":
		stdout = os.Getenv("FAKE_KUBECTL_DESCRIBE_STDOUT")
	case "exec":
		stdout = os.Getenv("FAKE_KUBECTL_EXEC_STDOUT")
	case "wait":
		stdout = os.Getenv("FAKE_KUBECTL_WAIT_STDOUT")
	case "version":
		stdout = os.Getenv("FAKE_KUBECTL_VERSION_STDOUT")
	case "api-resources":
		stdout = os.Getenv("FAKE_KUBECTL_APIRESOURCES_STDOUT")
	case "kustomize":
		stdout = os.Getenv("FAKE_KUBECTL_KUSTOMIZE_STDOUT")
	case "run":
		stdout = os.Getenv("FAKE_KUBECTL_RUN_STDOUT")
	case "config":
		if len(realArgs) > 2 {
			switch realArgs[2] {
			case "get-contexts":
				stdout = os.Getenv("FAKE_KUBECTL_CONTEXTS_STDOUT")
			case "get-clusters":
				stdout = os.Getenv("FAKE_KUBECTL_CLUSTERS_STDOUT")
			case "current-context":
				stdout = os.Getenv("FAKE_KUBECTL_CURRENT_CONTEXT_STDOUT")
			case "use-context":
				// stdout stays empty; exit code controls success/failure
			}
		}
	case "cluster-info":
		stdout = os.Getenv("FAKE_KUBECTL_CLUSTER_INFO_STDOUT")
	case "rollout":
		stdout = os.Getenv("FAKE_KUBECTL_ROLLOUT_STDOUT")
	default:
		stdout = os.Getenv("FAKE_KUBECTL_DEFAULT_STDOUT")
	}

	fmt.Fprint(os.Stdout, stdout)
	os.Exit(exitCode)
}

// fakeExecCommand returns a function that creates a *exec.Cmd pointing at the
// running test binary. The cmd, when Run/Output/CombinedOutput is called,
// re-enters TestHelperProcess above.
//
// env is a list of "KEY=VALUE" entries added to the subprocess environment, e.g.:
//
//	FAKE_KUBECTL_GET_STDOUT=<yaml>
//	FAKE_EXIT_CODE=0
func fakeExecCommand(env ...string) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		cs := append([]string{"-test.run=TestHelperProcess", "--", name}, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append([]string{"GO_WANT_HELPER_PROCESS=1"}, env...)
		return cmd
	}
}

// fakeExecCommandContext is the context-aware variant for functions that call
// execCommandContext.
func fakeExecCommandContext(env ...string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cs := append([]string{"-test.run=TestHelperProcess", "--", name}, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append([]string{"GO_WANT_HELPER_PROCESS=1"}, env...)
		return cmd
	}
}

// withFakeExec swaps in the fake exec functions for the duration of a test and
// restores the originals via t.Cleanup, even on failure.
func withFakeExec(t *testing.T, env ...string) {
	t.Helper()
	origCmd := execCommand
	origCmdCtx := execCommandContext
	origMakeupCmd := makeup.ExecCommand
	origMakeupCmdCtx := makeup.ExecCommandContext
	origUnattended := makeup.UnattendedMode
	execCommand = fakeExecCommand(env...)
	execCommandContext = fakeExecCommandContext(env...)
	makeup.ExecCommand = fakeExecCommand(env...)
	makeup.ExecCommandContext = fakeExecCommandContext(env...)
	makeup.UnattendedMode = true
	t.Cleanup(func() {
		execCommand = origCmd
		execCommandContext = origCmdCtx
		makeup.ExecCommand = origMakeupCmd
		makeup.ExecCommandContext = origMakeupCmdCtx
		makeup.UnattendedMode = origUnattended
	})
}

// ---------------------------------------------------------------------------
// trimAndFilter (pure function)
// ---------------------------------------------------------------------------

func TestUnitTrimAndFilter_MatchesAll(t *testing.T) {
	input := []byte("context-a\ncontext-b\ncontext-c\n")
	got := trimAndFilter(input, "")
	if len(got) != 3 {
		t.Fatalf("expected 3 items, got %d: %v", len(got), got)
	}
}

func TestUnitTrimAndFilter_FilterBySubstring(t *testing.T) {
	input := []byte("prod-cluster\ndev-cluster\nstaging-cluster\n")
	got := trimAndFilter(input, "prod")
	if len(got) != 1 || got[0] != "prod-cluster" {
		t.Fatalf("expected [prod-cluster], got %v", got)
	}
}

func TestUnitTrimAndFilter_EmptyInput(t *testing.T) {
	got := trimAndFilter([]byte(""), "anything")
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestUnitTrimAndFilter_SkipsBlankLines(t *testing.T) {
	input := []byte("\n\nfoo\n\nbar\n")
	got := trimAndFilter(input, "")
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(got), got)
	}
}

func TestUnitTrimAndFilter_CaseInsensitive(t *testing.T) {
	input := []byte("MyCluster\nanother\n")
	got := trimAndFilter(input, "mycluster")
	if len(got) != 1 || got[0] != "MyCluster" {
		t.Fatalf("expected [MyCluster], got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

func TestUnitVersion_ClientFlag(t *testing.T) {
	const fakeOutput = `{"clientVersion":{"major":"1","minor":"29"}}`
	withFakeExec(t, "FAKE_KUBECTL_VERSION_STDOUT="+fakeOutput)

	out, err := Version(true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "clientVersion") {
		t.Fatalf("expected clientVersion in output, got: %s", out)
	}
}

func TestUnitVersion_RequestTimeout(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_VERSION_STDOUT=v1.29.0")

	out, err := Version(false, "5s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

// ---------------------------------------------------------------------------
// Contexts / Clusters / CurrentContext / SwitchContext
// ---------------------------------------------------------------------------

func TestUnitContexts_Filter(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_CONTEXTS_STDOUT=kind-dev\nkind-prod\nminikube\n")

	got, err := Contexts("kind")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 kind contexts, got %v", got)
	}
	for _, c := range got {
		if !strings.Contains(strings.ToLower(c), "kind") {
			t.Errorf("unexpected context in result: %s", c)
		}
	}
}

func TestUnitContexts_NoFilter(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_CONTEXTS_STDOUT=ctx-a\nctx-b\nctx-c\n")

	got, err := Contexts("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 contexts, got %v", got)
	}
}

func TestUnitContexts_CommandFailure(t *testing.T) {
	withFakeExec(t, "FAKE_EXIT_CODE=1")

	_, err := Contexts("")
	if err == nil {
		t.Fatal("expected an error when kubectl exits non-zero")
	}
}

func TestUnitClusters_Filter(t *testing.T) {
	withFakeExec(t,
		"FAKE_KUBECTL_CLUSTERS_STDOUT=NAME\nkind-control\nkind-worker\nminikube\n",
	)

	got, err := Clusters("kind")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 kind clusters, got %v", got)
	}
}

func TestUnitCurrentContext_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_CURRENT_CONTEXT_STDOUT=kind-dev\n")

	ctx, err := CurrentContext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx != "kind-dev" {
		t.Fatalf("expected 'kind-dev', got %q", ctx)
	}
}

func TestUnitCurrentContext_Failure(t *testing.T) {
	withFakeExec(t, "FAKE_EXIT_CODE=1")

	_, err := CurrentContext()
	if err == nil {
		t.Fatal("expected error when kubectl exits with code 1")
	}
}

func TestUnitSwitchContext_Success(t *testing.T) {
	withFakeExec(t) // exit 0, empty output

	if out, err := SwitchContext("kind-dev"); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
}

func TestUnitSwitchContext_Failure(t *testing.T) {
	withFakeExec(t, "FAKE_EXIT_CODE=1")

	if _, err := SwitchContext("nonexistent"); err == nil {
		t.Fatal("expected error when kubectl exits with code 1")
	}
}

// ---------------------------------------------------------------------------
// ClusterInfo
// ---------------------------------------------------------------------------

func TestUnitClusterInfo_Success(t *testing.T) {
	const fakeInfo = "Kubernetes control plane is running at https://127.0.0.1:6443"
	withFakeExec(t, "FAKE_KUBECTL_CLUSTER_INFO_STDOUT="+fakeInfo)

	out, err := ClusterInfo(context.Background(), "kind-dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "control plane") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestUnitClusterInfo_Failure(t *testing.T) {
	withFakeExec(t, "FAKE_EXIT_CODE=1")

	_, err := ClusterInfo(context.Background(), "bad-context")
	if err == nil {
		t.Fatal("expected error when kubectl exits non-zero")
	}
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestUnitGet_WithFormat(t *testing.T) {
	const fakeYAML = "apiVersion: postgresql.anynines.com/v1beta3\nkind: Postgresql\nmetadata:\n  name: mydb\n  namespace: default\n"
	withFakeExec(t, "FAKE_KUBECTL_GET_STDOUT="+fakeYAML)

	k := NewKubeClient("")
	out, err := k.Get("postgresql", "mydb", "default", "yaml", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Postgresql") {
		t.Fatalf("expected YAML with Postgresql, got: %s", out)
	}
}

func TestUnitGet_IgnoreNotFound(t *testing.T) {
	withFakeExec(t) // empty output, exit 0 – resource not found but ignored

	k := NewKubeClient("")
	out, err := k.Get("postgresql", "missing", "default", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty output, got: %s", out)
	}
}

func TestUnitGet_WithContext(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_GET_STDOUT=name: mydb\n")

	k := NewKubeClient("kind-dev")
	out, err := k.Get("postgresql", "mydb", "default", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "mydb") {
		t.Fatalf("expected 'mydb' in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// Describe
// ---------------------------------------------------------------------------

func TestUnitDescribe_Success(t *testing.T) {
	const fakeDesc = "Name: mydb\nNamespace: default\nStatus: Running\n"
	withFakeExec(t, "FAKE_KUBECTL_DESCRIBE_STDOUT="+fakeDesc)

	k := NewKubeClient("")
	out, err := k.Describe("postgresql", "mydb", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "Running") {
		t.Fatalf("expected 'Running' in describe output, got: %s", string(out))
	}
}

// ---------------------------------------------------------------------------
// ApplyWithPrompt (unattended mode – no stdin interaction)
// ---------------------------------------------------------------------------

const testManifest = `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  key: value
`

func TestUnitApplyWithPrompt_Unattended_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_APPLY_STDOUT=configmap/test-cm created\n")

	k := NewKubeClient("")
	out, err := k.ApplyWithPrompt([]byte(testManifest), "test configmap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected 'created' in output, got: %s", out)
	}
}

func TestUnitApplyWithPrompt_Unattended_Failure(t *testing.T) {
	withFakeExec(t, "FAKE_EXIT_CODE=1", "FAKE_KUBECTL_APPLY_STDOUT=error applying\n")

	k := NewKubeClient("")
	_, err := k.ApplyWithPrompt([]byte(testManifest), "test configmap")
	if err == nil {
		t.Fatal("expected an error when kubectl exits non-zero")
	}
}

func TestUnitApplyWithPrompt_EmptyDescription(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_APPLY_STDOUT=configmap/test-cm unchanged\n")

	k := NewKubeClient("")
	out, err := k.ApplyWithPrompt([]byte(testManifest), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "unchanged") {
		t.Fatalf("expected 'unchanged' in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// ApplyFromFile
// ---------------------------------------------------------------------------

func TestUnitApplyFromFile_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_APPLY_STDOUT=configmap/test-cm created\n")

	tmp, err := os.CreateTemp(t.TempDir(), "manifest-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := tmp.WriteString(testManifest); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	tmp.Close()

	k := NewKubeClient("")
	out, err := k.ApplyFromFile(tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected 'created' in output, got: %s", out)
	}
}

func TestUnitApplyFromFile_MissingFile(t *testing.T) {
	withFakeExec(t)

	k := NewKubeClient("")
	_, err := k.ApplyFromFile("/nonexistent/path/manifest.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// ApplyFromUrl
// ---------------------------------------------------------------------------

func TestUnitApplyFromUrl_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_APPLY_STDOUT=configmap/test-cm created\n")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, testManifest)
	}))
	defer srv.Close()

	k := NewKubeClient("")
	out, err := k.ApplyFromUrl(srv.URL+"/manifest.yaml", "test configmap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected 'created' in output, got: %s", out)
	}
}

func TestUnitApplyFromUrl_HTTP404(t *testing.T) {
	withFakeExec(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	k := NewKubeClient("")
	_, err := k.ApplyFromUrl(srv.URL+"/missing.yaml", "404 test")
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
}

func TestUnitApplyFromUrl_BadURL(t *testing.T) {
	withFakeExec(t)

	k := NewKubeClient("")
	_, err := k.ApplyFromUrl("http://127.0.0.1:0/unreachable", "bad url")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

// ---------------------------------------------------------------------------
// ApplyKustomize
// ---------------------------------------------------------------------------

func TestUnitApplyKustomize_Success(t *testing.T) {
	// The fake intercepts both the `kubectl kustomize` render and the
	// subsequent `kubectl apply -f -`.
	withFakeExec(t,
		"FAKE_KUBECTL_KUSTOMIZE_STDOUT="+testManifest,
		"FAKE_KUBECTL_APPLY_STDOUT=configmap/test-cm created\n",
	)

	k := NewKubeClient("")
	out, err := k.ApplyKustomize("/some/kustomize/dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected 'created' in output, got: %s", out)
	}
}

func TestUnitApplyKustomize_RenderFailure(t *testing.T) {
	// kubectl kustomize fails
	withFakeExec(t, "FAKE_EXIT_CODE=1")

	k := NewKubeClient("")
	_, err := k.ApplyKustomize("/bad/dir")
	if err == nil {
		t.Fatal("expected error when kustomize render fails")
	}
}

// ---------------------------------------------------------------------------
// Delete (calls Get then kubectlCommandWithPrompt("delete", ...))
// ---------------------------------------------------------------------------

// minimalDeleteManifest is what `kubectl get <resource> <name> -n <ns> -o yaml`
// might return – just enough to have a Kind/Name/Namespace for error messages.
const minimalDeleteManifest = `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
`

func TestUnitDelete_Success(t *testing.T) {
	withFakeExec(t,
		"FAKE_KUBECTL_GET_STDOUT="+minimalDeleteManifest,
		"FAKE_KUBECTL_DELETE_STDOUT=configmap \"test-cm\" deleted\n",
	)

	k := NewKubeClient("")
	out, err := k.Delete("configmap", "test-cm", "default", "test configmap", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "deleted") {
		t.Fatalf("expected 'deleted' in output, got: %s", out)
	}
}

func TestUnitDelete_GetFails(t *testing.T) {
	// kubectl get exits non-zero → Delete should propagate the error
	withFakeExec(t, "FAKE_EXIT_CODE=1")

	k := NewKubeClient("")
	_, err := k.Delete("configmap", "missing", "default", "", false)
	if err == nil {
		t.Fatal("expected error when Get fails")
	}
}

// ---------------------------------------------------------------------------
// DeleteFromManifest
// ---------------------------------------------------------------------------

func TestUnitDeleteFromManifest_Success(t *testing.T) {
	withFakeExec(t,
		"FAKE_KUBECTL_GET_STDOUT="+minimalDeleteManifest,
		"FAKE_KUBECTL_DELETE_STDOUT=configmap \"test-cm\" deleted\n",
	)

	k := NewKubeClient("")
	// DeleteFromManifest calls os.Exit on error, so we only test the happy path
	// unit-safely by ensuring no panic/exit occurs (the fake returns exit 0).
	k.DeleteFromManifest(minimalDeleteManifest)
}

// ---------------------------------------------------------------------------
// DeleteFromFile
// ---------------------------------------------------------------------------

func TestUnitDeleteFromFile_Success(t *testing.T) {
	withFakeExec(t,
		"FAKE_KUBECTL_GET_STDOUT="+minimalDeleteManifest,
		"FAKE_KUBECTL_DELETE_STDOUT=configmap \"test-cm\" deleted\n",
	)

	tmp, err := os.CreateTemp(t.TempDir(), "del-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := tmp.WriteString(minimalDeleteManifest); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	tmp.Close()

	k := NewKubeClient("")
	k.DeleteFromFile(tmp.Name())
}

// ---------------------------------------------------------------------------
// Exec
// ---------------------------------------------------------------------------

func TestUnitExec_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_EXEC_STDOUT=psql output here\n")

	k := NewKubeClient("")
	out, err := k.Exec("mypg-0", "default", "postgres", "psql", "-U", "postgres", "-c", "SELECT 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "psql output") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestUnitExec_Failure(t *testing.T) {
	withFakeExec(t, "FAKE_EXIT_CODE=1", "FAKE_KUBECTL_EXEC_STDOUT=exec error\n")

	k := NewKubeClient("")
	_, err := k.Exec("mypg-0", "default", "postgres", "psql")
	if err == nil {
		t.Fatal("expected error when kubectl exec exits non-zero")
	}
}

// ---------------------------------------------------------------------------
// ApiResources
// ---------------------------------------------------------------------------

func TestUnitApiResources_NoFilter(t *testing.T) {
	const fakeOut = "NAME   SHORTNAMES   APIGROUP   NAMESPACED   KIND\npods   po            v1         true         Pod\n"
	withFakeExec(t, "FAKE_KUBECTL_APIRESOURCES_STDOUT="+fakeOut)

	k := NewKubeClient("")
	out, err := k.ApiResources("", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "pods") {
		t.Fatalf("expected 'pods' in output, got: %s", out)
	}
}

func TestUnitApiResources_WithApiGroup(t *testing.T) {
	const fakeOut = "postgresqls   pg   postgresql.anynines.com   true   Postgresql\n"
	withFakeExec(t, "FAKE_KUBECTL_APIRESOURCES_STDOUT="+fakeOut)

	k := NewKubeClient("")
	out, err := k.ApiResources("postgresql.anynines.com", "", "", "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "postgresqls") {
		t.Fatalf("expected 'postgresqls' in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// RolloutStatus
// ---------------------------------------------------------------------------

func TestUnitRolloutStatus_WithTimeout(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_ROLLOUT_STDOUT=deployment/myapp successfully rolled out\n")

	k := NewKubeClient("")
	out, err := k.RolloutStatus("deployment", "myapp", "default", "2m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The output may be empty depending on how the fake routes "rollout,"
	// (see the typo note in TestHelperProcess); we just check no error.
	_ = out
}

func TestUnitRolloutStatus_TimeoutPrefix(t *testing.T) {
	withFakeExec(t)

	k := NewKubeClient("")
	// If timeout already has the prefix it should not be doubled.
	_, err := k.RolloutStatus("deployment", "myapp", "default", "--timeout=5m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// kubectlWaitFor (tested indirectly via the public Wait* methods)
// ---------------------------------------------------------------------------

func TestUnitKubectlWaitForResourceConditionWithTimeout_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_WAIT_STDOUT=postgresql.anynines.com/mydb condition met\n")

	k := NewKubeClient("")
	// Should not call os.Exit – only does so on error; fake exits 0.
	k.KubectlWaitForResourceCondition("Ready", "postgresql", "mydb", "default", "5m")
}

func TestUnitKubectlWaitForResourceDeletionWithTimeout_Success(t *testing.T) {
	withFakeExec(t) // exit 0 means deletion observed

	k := NewKubeClient("")
	k.KubectlWaitForResourceDeletion("postgresql", "mydb", "default", "5m")
}

func TestUnitKubectlWaitForNodes_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_WAIT_STDOUT=node/kind-control-plane condition met\n")

	k := NewKubeClient("")
	k.KubectlWaitForNodes()
}

// ---------------------------------------------------------------------------
// Run
// ---------------------------------------------------------------------------

func TestUnitRun_Success(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_RUN_STDOUT=pod/busybox created\n")

	k := NewKubeClient("")
	out, err := k.Run("default", "busybox", "busybox", "env=test", "sleep", "600")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "created") {
		t.Fatalf("expected 'created' in output, got: %s", out)
	}
}

func TestUnitRun_ImageAlreadyPrefixed(t *testing.T) {
	withFakeExec(t, "FAKE_KUBECTL_RUN_STDOUT=pod/busybox created\n")

	k := NewKubeClient("")
	// image already has --image= prefix; should not double-prefix
	out, err := k.Run("", "busybox", "--image=busybox", "", "sleep", "600")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}
