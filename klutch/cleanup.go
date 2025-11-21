package klutch

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/anynines/a9s-cli-v2/makeup"
)

// DeleteControlPlaneInstall removes Klutch control plane artifacts from the current kube context.
// It is intended to clean up resources created by "a9s apply klutch-control-plane".
func DeleteControlPlaneInstall() {
	makeup.PrintH1("Deleting Klutch control plane resources from the current Kubernetes cluster...")
	makeup.PrintWarning("This will delete the klutch backend, Dex related ingresses, services, secrets, and the crossplane release/namespace from the CURRENT Kubernetes context. This action cannot be undone.")

	makeup.PrintBright("Type YES to proceed: ")
	var confirm string
	_, err := fmt.Fscan(os.Stdin, &confirm)
	if err != nil || confirm != "YES" {
		makeup.PrintInfo("Deletion aborted.")
		return
	}

	manager := NewKlutchManagerWithContexts("", "")

	// Namespaced Klutch resources
	deleteKubectlResource(manager, "delete", "ingress", "dex-ingress", "-n", "default", "--ignore-not-found")
	deleteKubectlResource(manager, "delete", "ingress", "anynines-backend", "-n", "default", "--ignore-not-found")
	deleteKubectlResource(manager, "delete", "service", "dex", "anynines-backend", "-n", "default", "--ignore-not-found")
	deleteKubectlResource(manager, "delete", "deployment", "dex", "anynines-backend", "-n", "default", "--ignore-not-found")
	deleteKubectlResource(manager, "delete", "secret", "oidc-config", "cookie-config", "k8sca", "-n", "default", "--ignore-not-found")
	deleteKubectlResource(manager, "delete", "configmap", "dex-config", "-n", "default", "--ignore-not-found")

	// Ingress-nginx (only deletes namespace; harmless if it doesn't exist)
	deleteKubectlResource(manager, "delete", "ns", "ingress-nginx", "--ignore-not-found")

	// Crossplane helm release
	cmd := exec.Command("helm", "uninstall", "crossplane", "-n", "crossplane-system")
	if output, err := cmd.CombinedOutput(); err != nil {
		makeup.PrintWarning(fmt.Sprintf("Helm uninstall crossplane failed (ignored): %v %s", err, string(output)))
	} else {
		makeup.PrintCheckmark("Uninstalled crossplane helm release.")
	}

	// Delete crossplane namespace (best-effort)
	deleteKubectlResource(manager, "delete", "ns", "crossplane-system", "--ignore-not-found")

	makeup.PrintSuccessSummary("Klutch control plane resources removed from the current cluster.")
}

func deleteKubectlResource(manager *KlutchManager, args ...string) {
	cmd := manager.cpK8s.KubectlWithContextCommand(args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.PrintWarning(fmt.Sprintf("Could not delete resource (%v): %s", err, string(output)))
		return
	}
	makeup.PrintCheckmark(fmt.Sprintf("Deleted resource via kubectl %v", args))
}
