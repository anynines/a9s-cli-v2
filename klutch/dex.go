package klutch

import (
	"context"
	_ "embed"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//go:embed templates/dex.tmpl
var dexManifestsTemplate string

type dexTemplateVars struct {
	Host            string
	IngressPort     string
	DexClientSecret string
}

func (k *KlutchManager) DeployDex(hostIP string, ingressPort string) {
	makeup.PrintH1("Deploying Dex Idp...")

	client := k.cpK8s.GetKubernetesClientSet()
	bg := RandomByteGenerator{}
	dexClientSecret := k.getOIDCIssuerClientSecret(client, bg)

	templateVars := &dexTemplateVars{
		Host:            hostIP,
		IngressPort:     ingressPort,
		DexClientSecret: dexClientSecret,
	}

	manifests, err := renderTemplate(dexManifestsTemplate, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not render the dex manifests.")
	}

	makeup.PrintH2("Applying the following manifests: ")
	makeup.PrintYAML(manifests.Bytes(), false)
	makeup.WaitForUser(demo.UnattendedMode)

	k.cpK8s.KubectlApplyStdin(manifests)

	makeup.Print("Done applying the dex manifests.")
}

// Waits for the dex deployment to be ready.
func (k *KlutchManager) WaitForDex() {
	makeup.PrintH1("Waiting for dex to become ready...")

	k.cpK8s.KubectlWaitForRollout("deployment", "dex", "default")

	makeup.PrintCheckmark("Dex appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}

// getOIDCIssuerClientSecret checks if the dex oidc-config secret exists and returns the oidc-issuer-client-secret if set.
// If the secret or key don't exist, or the value in the secret is empty, a randomly generated value is returned.
// This is done so an existing secret isn't overwritten by a randomly generated value when re-applying the manifests.
func (k *KlutchManager) getOIDCIssuerClientSecret(client kubernetes.Interface, bg ByteGenerator) string {
	secretName := "oidc-config" // See dex manifests
	secretKey := "oidc-issuer-client-secret"

	secret, err := client.CoreV1().Secrets("default").Get(context.TODO(), secretName, v1.GetOptions{})

	if err != nil && !errors.IsNotFound(err) {
		makeup.ExitDueToFatalError(err, "Could not get oidc-config secret.")
	}

	if err != nil {
		// Secret doesn't exist
		return bg.GenerateRandom32BytesBase64()
	}

	clientSecret, ok := secret.Data[secretKey]
	if !ok || string(clientSecret) == "" {
		return bg.GenerateRandom32BytesBase64()
	}

	return string(clientSecret)
}
