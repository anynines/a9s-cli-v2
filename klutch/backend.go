package klutch

import (
	_ "embed"
	"encoding/base64"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
)

const (
	backendCRDManifestsURL = "https://anynines-artifacts.s3.eu-central-1.amazonaws.com/central-management/v1.3.0/crds.yaml"
)

//go:embed templates/backend.tmpl
var bindBackendManifestsTemplate string

type backendTemplateVars struct {
	Host                string
	CookieEncryptionKey string
	CookieSigningKey    string
	K8sApiPort          string
	K8sApiCaCertB64     string
}

// Deploys dex and the klutch-bind backend.
func (k *KlutchManager) DeployBindBackend(hostIP string) {
	makeup.PrintH1("Deploying the klutch-bind backend...")

	makeup.PrintH2("Applying the klutch-bind backend CRDs...")

	k.cpK8s.KubectlApplyF(backendCRDManifestsURL, true)

	makeup.PrintCheckmark("klutch-bind backend CRDs applied.")
	makeup.PrintH2("Applying the klutch-bind backend manifests...")

	// We need the provider cluster's CA certificate
	clusterCert := getClusterCert(k.cpK8s)
	encodedCert := base64.StdEncoding.EncodeToString(clusterCert)

	clusterPort := getClusterExternalPort(contextControlPlane)

	cookieSigningKey := generateRandom32BytesBase64()
	cookieEncryptionKey := generateRandom32BytesBase64()

	templateVars := &backendTemplateVars{
		Host:                hostIP,
		CookieEncryptionKey: cookieEncryptionKey,
		CookieSigningKey:    cookieSigningKey,
		K8sApiPort:          clusterPort,
		K8sApiCaCertB64:     encodedCert,
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
