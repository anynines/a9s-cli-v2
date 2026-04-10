package klutch

import (
	_ "embed"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
)

const (
	gatewayManifestsUrl = "https://github.com/envoyproxy/gateway/releases/download/v1.7.1/install.yaml"
)

// DeployEnvoyGateway applies the Envoy Gateway manifests and an additional configMap to configure it.
// The config increases the request header size limit to cope with bind's header sizes becoming very large.
func (k *KlutchManager) DeployEnvoyGateway() {
	makeup.PrintH1("Applying Envoy Gateway manifests...")

	if _, _, err := k.cpK8s.Kubectl(demo.UnattendedMode, "apply", "-f", gatewayManifestsUrl, "--server-side", "--force-conflicts"); err != nil {
		makeup.ExitDueToFatalError(err, "could not apply Envoy Gateway Manifests")
	}

	makeup.Print("Done applying Envoy Gateway manifests.")
}

func (k *KlutchManager) WaitForEnvoyGateway() {
	makeup.PrintH1("Waiting for Envoy Gateway to become ready...")

	k.cpK8s.KubectlWaitForRollout("deployment", "envoy-gateway", "envoy-gateway-system")

	makeup.PrintCheckmark("Envoy Gateway appears to be ready.")
}
