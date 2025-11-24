package klutch

import (
	_ "embed"
	"encoding/base64"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
)

const (
	backendCRDManifestsURL = "https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/v1.4.1/crds.yaml"
)

//go:embed templates/backend.tmpl
var bindBackendManifestsTemplate string

type backendTemplateVars struct {
	CookieEncryptionKey string
	CookieSigningKey    string
	ExternalAddress     string
	IngressPort         string
	K8sApiCaCertB64     string
	IngressClass        string
	Scheme              string
}

// Deploys dex and the klutch-bind backend.
func (k *KlutchManager) DeployBindBackend(ingressPort string, ingressClass string, scheme string) {
	makeup.PrintH1("Deploying the klutch-bind backend...")

	makeup.PrintH2("Applying the klutch-bind backend CRDs...")

	k.cpK8s.KubectlApplyF(backendCRDManifestsURL, true)

	makeup.PrintCheckmark("klutch-bind backend CRDs applied.")
	makeup.PrintH2("Applying the klutch-bind backend manifests...")

	// We need the provider cluster's CA certificate
	clusterCert := getClusterCert(k.cpK8s)
	encodedCert := base64.StdEncoding.EncodeToString(clusterCert)

	rbg := RandomByteGenerator{}
	cookieSigningKey := rbg.GenerateRandom32BytesBase64()
	cookieEncryptionKey := rbg.GenerateRandom32BytesBase64()
	externalAddress := getClusterExternalAddress(k.cpContext)

	templateVars := &backendTemplateVars{
		CookieEncryptionKey: cookieEncryptionKey,
		CookieSigningKey:    cookieSigningKey,
		ExternalAddress:     externalAddress,
		IngressPort:         ingressPort,
		K8sApiCaCertB64:     encodedCert,
		IngressClass:        ingressClass,
		Scheme:              scheme,
	}

	manifests, err := renderTemplate(bindBackendManifestsTemplate, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not render the klutch-bind backend manifests.")
	}

	makeup.PrintH2("Creating a kind cluster with following config: ")
	makeup.PrintYAML(manifests.Bytes(), false)
	makeup.WaitForUser(demo.UnattendedMode)

	k.cpK8s.KubectlApplyStdin(manifests)

	makeup.Print("klutch-bind backend applied.")
}

// Waits for the klutch-bind backend deployment to be ready.
// Note: the manifests contain an init-container which waits for dex to be ready,
// because the backend requires dex to be up and running in order to start.
// This avoids delays/complications due to crash loop backoffs.
func (k *KlutchManager) WaitForBindBackend() {
	makeup.PrintH1("Waiting for the klutch-bind backend to become ready...")

	k.cpK8s.KubectlWaitForRollout("deployment", "anynines-backend", "default")

	makeup.PrintCheckmark("The klutch-bind backend appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}
