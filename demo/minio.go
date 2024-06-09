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
func ApplyMinioManifests() {
	makeup.PrintH1("Applying Minio manifests...")

	minioManifestPath := filepath.Join(DemoConfig.WorkingDir, DemoAppLocalDir, "minio")
	k8s.KubectlApplyKustomize(minioManifestPath, UnattendedMode)

	makeup.PrintCheckmark("Done applying Minio manifests.")
}

func WaitForMinioToBecomeReady() {
	expectedPods := []k8s.PodExpectationState{
		{Name: "minio", Running: false},
	}

	k8s.WaitForSystemToBecomeReady(MinioNamespace, MinioSystemName, expectedPods)
	makeup.WaitForUser(UnattendedMode)
}
