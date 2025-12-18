package klutch

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/makeup"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// DeleteControlPlaneInstall removes Klutch control plane artifacts from the current kube context.
// It is intended to clean up resources created by "a9s apply klutch-control-plane".
func DeleteControlPlaneInstall() {
	makeup.PrintH1("Deleting Klutch control plane resources from the current Kubernetes cluster...")
	makeup.PrintWarning("This will delete the klutch backend, Dex related ingresses, services, secrets, the tenant operator release/namespace, and the crossplane release/namespace from the CURRENT Kubernetes context. This action cannot be undone.")

	ctxName, server := currentContextAndServer()
	if ctxName != "" || server != "" {
		makeup.PrintInfo(fmt.Sprintf("Current kube context: %s (server: %s)", ctxName, server))
	}

	if !makeup.ConfirmYes("Type 'yes' to proceed: ") {
		makeup.PrintInfo("Deletion aborted.")
		return
	}
	makeup.PrintInfo("Deletion accepted. Starting deletion...")

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

	// Tenant operator helm release
	cmd = exec.Command("helm", "uninstall", "a9s-tenants-operator", "-n", "a9s-tenants-operator-system")
	if output, err := cmd.CombinedOutput(); err != nil {
		makeup.PrintWarning(fmt.Sprintf("Helm uninstall tenant operator failed (ignored): %v %s", err, string(output)))
	} else {
		makeup.PrintCheckmark("Uninstalled tenant operator helm release.")
	}

	// Delete tenant operator namespace (best-effort)
	deleteKubectlResource(manager, "delete", "ns", "a9s-tenants-operator-system", "--ignore-not-found")

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

func currentContextAndServer() (string, string) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", ""
	}

	ctxName := config.CurrentContext
	ctx, ok := config.Contexts[ctxName]
	if !ok {
		return ctxName, ""
	}

	cluster, ok := config.Clusters[ctx.Cluster]
	if !ok {
		return ctxName, ""
	}

	return ctxName, cluster.Server
}
