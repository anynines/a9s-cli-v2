package k8s

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/makeup"
)

func (k *KubeClient) ApplyCertManagerManifests(waitForUser bool) {
	makeup.PrintH1("Ensuring running cert-manager")
	output, err := k.Get("deployment", "cert-manager", "cert-manager", "", true)
	if err != nil {
		makeup.ExitDueToFatalError(err, "could not get cert-manager deployment")
	}
	if output != "" {
		makeup.PrintCheckmark("Deployment cert-manager already exists, skipping")
		return
	}
	count := k.CountPodsInNamespace(CertManagerNamespace)

	if count > 0 {
		makeup.Print(fmt.Sprintf("Found %d pods in the %s namespace", count, CertManagerNamespace))
	}

	// Fetch and apply cert-manager manifests
	if _, err := k.ApplyFromUrl(CertManagerManifestUrl, CertManagerManifestUrl); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply cert-manager manifests")
	}
	k.WaitForCertManagerToBecomeReady()
}

func (k *KubeClient) WaitForCertManagerToBecomeReady() {
	makeup.PrintH1("Waiting for the cert-manager components to become ready...")

	deployments := []string{"cert-manager", "cert-manager-webhook", "cert-manager-cainjector"}
	namespace := "cert-manager"

	for _, deployment := range deployments {
		k.KubectlWaitForRollout("deployment", deployment, namespace)
	}

	makeup.PrintCheckmark("The cert-manager API is ready.")
}
