package k8s

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/makeup"
)

func (k *KubeClient) ApplyCertManagerManifests(waitForUser bool) {
	makeup.PrintH1("Installing the cert-manager")
	count := k.CountPodsInNamespace(CertManagerNamespace)

	if count > 0 {
		makeup.Print(fmt.Sprintf("Found %d pods in the %s namespace", count, CertManagerNamespace))
	}

	k.KubectlApplyF(CertManagerManifestUrl, waitForUser)

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
