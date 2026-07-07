package klutch

import (
	_ "embed"

	"github.com/anynines/a9s-cli-v2/makeup"
)

const (
	ingressManifestsUrl = "https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.1/deploy/static/provider/kind/deploy.yaml"
)

//go:embed manifests/nginx-ingress-config.yaml
var ingressConfigMap string

// DeployIngressNginx applies the ingress-nginx manifests and an additional configMap to configure it.
// The config increases the request header size limit to cope with bind's header sizes becoming very large.
func (k *KlutchManager) DeployIngressNginx() {
	makeup.PrintH1("Applying ingress-nginx manifests...")

	// Fetch and apply ingress-nginx manifests
	if _, err := k.cpK8s.ApplyFromUrl(ingressManifestsUrl, ingressManifestsUrl); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply ingress-nginx manifests")
	}

	// Apply configmap
	if _, err := k.cpK8s.ApplyWithPrompt([]byte(ingressConfigMap), "ingress-nginx configmap"); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply ingress-nginx configmap")
	}

	makeup.Print("Done applying ingress-nginx manifests.")
}

const (
	gatewayManifestsUrl = "https://github.com/envoyproxy/gateway/releases/download/v1.7.1/install.yaml"
)

// TODO(jlu) fix this after the merge conflicts have been resolved
// // DeployEnvoyGateway applies the Envoy Gateway manifests and an additional configMap to configure it.
// // The config increases the request header size limit to cope with bind's header sizes becoming very large.
// func (k *KlutchManager) DeployEnvoyGateway() {
// 	makeup.PrintH1("Applying Envoy Gateway manifests...")

// 	if _, _, err := k.cpK8s.Kubectl(demo.UnattendedMode, "apply", "-f", gatewayManifestsUrl, "--server-side", "--force-conflicts"); err != nil {
// 		makeup.ExitDueToFatalError(err, "could not apply Envoy Gateway Manifests")
// 	}

// 	makeup.Print("Done applying Envoy Gateway manifests.")
// }

func (k *KlutchManager) WaitForIngressNginx() {
	makeup.PrintH1("Waiting for Ingress Nginx to become ready...")

	k.cpK8s.KubectlWaitForRollout("deployment", "ingress-nginx-controller", "ingress-nginx")
	k.cpK8s.KubectlWaitForResourceCondition("complete", "job", "ingress-nginx-admission-create", "ingress-nginx", "")
	k.cpK8s.KubectlWaitForResourceCondition("complete", "job", "ingress-nginx-admission-patch", "ingress-nginx", "")

	// TODO(jlu) fix this after the merge conflicts have been resolved
	// k.cpK8s.KubectlWaitForRollout("deployment", "envoy-gateway", "envoy-gateway-system")

	// makeup.PrintCheckmark("Envoy Gateway appears to be ready.")
}
