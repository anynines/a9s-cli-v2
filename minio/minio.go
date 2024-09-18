package minio

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
)

const MinioNamespace = "minio-dev"
const MinioSystemName = "Minio"

// It tells the Go compiler to embed all files and directories under the manifests/ directory into the binary.
//
//go:embed manifest/*
var manifestsFS embed.FS

type MinioManager struct {
	K8s            *k8s.KubeClient
	UnattendedMode bool
}

func NewMinioManager(kubeContext string, unattendedMode bool) *MinioManager {
	return &MinioManager{
		K8s:            k8s.NewKubeClient(kubeContext),
		UnattendedMode: unattendedMode,
	}
}

func (m *MinioManager) ApplyMinioManifests(workingDir string) {
	makeup.PrintH1("Applying Minio manifest...")

	minioManifestPath := filepath.Join(workingDir, "minio")
	m.K8s.KubectlApplyF(minioManifestPath, m.UnattendedMode)

	makeup.PrintCheckmark("Done applying Minio manifest.")
}

func (m *MinioManager) WaitForMinioToBecomeReady() {
	expectedPods := []k8s.PodExpectationState{
		{Name: "minio", Running: false},
	}

	m.K8s.WaitForSystemToBecomeReady(MinioNamespace, MinioSystemName, expectedPods)
	makeup.WaitForUser(m.UnattendedMode)
}

func SetupMinioRepository(workingDir string) {
	makeup.PrintH1("Setting up MinIO repository...")

	minioRepoPath := filepath.Join(workingDir, "minio")
	makeup.Print("Local repository path: " + minioRepoPath)

	if err := os.MkdirAll(minioRepoPath, 0755); err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to create minio directory: %v", err))
		return
	}

	// It walks through the "manifest" directory in the embedded filesystem 'manifestsFS'.
	err := fs.WalkDir(manifestsFS, "manifest", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, err := manifestsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		outPath := filepath.Join(minioRepoPath, filepath.Base(path))
		if err := os.WriteFile(outPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", outPath, err)
		}

		return nil
	})

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("failed to set up MinIO repository: %v", err))
	}

	makeup.PrintCheckmark("MinIO repository set up successfully")
}
