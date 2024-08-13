package demo

import (
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
)

const MinioNamespace = "minio-dev"
const MinioSystemName = "Minio"

/*
Minio related automation.
*/

/*
Applies minio related manifests.
*/
func (m *A8sDemoManager) ApplyMinioManifests() {
	makeup.PrintH1("Applying Minio manifests...")

	// TODO Make CLI parameter
	minioNamespace := "minio-dev"

	m.K8s.CreateNamespaceIfNotExists(UnattendedMode, minioNamespace)

	m.K8s.WaitForServiceAccount(UnattendedMode, minioNamespace, "default")
	// TODO: Above commands are probably needed because the minio manifests in a8s-demo repo are
	// for a simple pod, and not a deployment.

	minioManifestPath := filepath.Join(DemoConfig.WorkingDir, DemoAppLocalDir, "minio")
	m.K8s.KubectlApplyKustomize(minioManifestPath, UnattendedMode)

	makeup.PrintCheckmark("Done applying Minio manifests.")
}

func (m *A8sDemoManager) WaitForMinioToBecomeReady() {
	expectedPods := []k8s.PodExpectationState{
		{Name: "minio", Running: false},
	}

	m.K8s.WaitForSystemToBecomeReady(MinioNamespace, MinioSystemName, expectedPods)
	makeup.WaitForUser(UnattendedMode)
}
