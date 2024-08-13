package klutch

import (
	"bytes"
	_ "embed"

	"github.com/anynines/a9s-cli-v2/makeup"
)

const (
	ingressManifestsUrl = "https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml"
)

//go:embed manifests/nginx-ingress-config.yaml
var ingressConfigMap string

// DeployIngressNginx applies the ingress-nginx manifests and an additional configMap to configure it.
// The config increases the request header size limit to cope with bind's header sizes becoming very large.
func (k *KlutchManager) DeployIngressNginx() {
	makeup.PrintH1("Applying ingress-nginx manifests...")

	k.mgmtK8s.KubectlApplyF(ingressManifestsUrl, true)

	// Apply configmap
	in := bytes.NewBufferString(ingressConfigMap)
	k.mgmtK8s.KubectlApplyStdin(in)

	makeup.Print("Done applying ingress-nginx manifests.")
}

func (k *KlutchManager) WaitForIngressNginx() {
	makeup.PrintH1("Waiting for ingress-nginx to become ready...")

	k.mgmtK8s.KubectlWaitForRollout("deployment", "ingress-nginx-controller", "ingress-nginx")
	k.mgmtK8s.KubectlWaitForResourceCondition("complete", "job", "ingress-nginx-admission-create", "ingress-nginx")
	k.mgmtK8s.KubectlWaitForResourceCondition("complete", "job", "ingress-nginx-admission-patch", "ingress-nginx")

	makeup.PrintCheckmark("ingress-nginx appears to be ready.")
}
