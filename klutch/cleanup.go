package klutch

import (
	"fmt"
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
	manager.deleteKubectlResource("ingress", "dex-ingress", "default")
	manager.deleteKubectlResource("ingress", "anynines-backend", "default")
	manager.deleteKubectlResource("service", "dex", "default")
	manager.deleteKubectlResource("service", "anynines-backend", "default")
	manager.deleteKubectlResource("deployment", "dex", "default")
	manager.deleteKubectlResource("deployment", "anynines-backend", "default")
	manager.deleteKubectlResource("secret", "oidc-config", "default")
	manager.deleteKubectlResource("secret", "cookie-config", "default")
	manager.deleteKubectlResource("secret", "k8sca", "default")
	manager.deleteKubectlResource("configmap", "dex-config", "default")

	// Ingress-nginx (only deletes namespace; harmless if it doesn't exist)
	manager.deleteKubectlResource("ns", "ingress-nginx", "")

	// Crossplane helm release
	if output, err := makeup.Command("helm", "uninstall", "crossplane", "-n", "crossplane-system").WithPrompt().Run(); err != nil {
		makeup.PrintWarning(fmt.Sprintf("Helm uninstall crossplane failed (ignored): %v %s", err, string(output)))
	} else {
		makeup.PrintCheckmark("Uninstalled crossplane helm release.")
	}

	// Delete crossplane namespace (best-effort)
	manager.deleteKubectlResource("ns", "crossplane-system", "")

	// Tenant operator helm release
	if output, err := makeup.Command("helm", "uninstall", "a9s-tenants-operator", "-n", "a9s-tenants-operator-system").WithPrompt().Run(); err != nil {
		makeup.PrintWarning(fmt.Sprintf("Helm uninstall tenant operator failed (ignored): %v %s", err, string(output)))
	} else {
		makeup.PrintCheckmark("Uninstalled tenant operator helm release.")
	}

	// Delete tenant operator namespace (best-effort)
	manager.deleteKubectlResource("ns", "a9s-tenants-operator-system", "")

	makeup.PrintSuccessSummary("Klutch control plane resources removed from the current cluster.")
}

func (m *KlutchManager) deleteKubectlResource(resourceType, name, namespace string) {

	output, err := m.cpK8s.Delete(resourceType, name, namespace, "", true)
	if err != nil {
		makeup.PrintWarning(fmt.Sprintf("Could not delete resource (%v): %s", err, string(output)))
		return
	}
	makeup.PrintCheckmark(fmt.Sprintf("Deleted resource %s/%s.%s via kubectl", namespace, name, resourceType))
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
