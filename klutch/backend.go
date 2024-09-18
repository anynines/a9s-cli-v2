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
	// Host name by which the backend is accessed from outside the docker network. Should be "127.0.0.1".
	ExternalHost string
	// External ingress port accessed from outside the docker network. Should be the flag passed by the user.
	ExternalIngressPort string
	// Internal ingress port, corresponds to the one defined in the applied nginx-ingress manifests. Should be "80".
	InternalIngressPort string
	CookieEncryptionKey string
	CookieSigningKey    string
	// Docker internal k8s api port. Should be 6443 by default.
	K8sApiPort      string
	K8sApiCaCertB64 string
}

// Deploys dex and the klutch-bind backend.
func (k *KlutchManager) DeployBindBackend(port string) {
	makeup.PrintH1("Deploying the klutch-bind backend...")

	makeup.PrintH2("Applying the klutch-bind backend CRDs...")

	k.mgmtK8s.KubectlApplyF(backendCRDManifestsURL, true)

	makeup.PrintCheckmark("klutch-bind backend CRDs applied.")
	makeup.PrintH2("Applying the klutch-bind backend manifests...")

	// We need the provider cluster's CA certificate
	clusterCert := getClusterCert(k.mgmtK8s)
	encodedCert := base64.StdEncoding.EncodeToString(clusterCert)

	cookieSigningKey := generateRandom32BytesBase64()
	cookieEncryptionKey := generateRandom32BytesBase64()

	templateVars := &backendTemplateVars{
		ExternalHost:        "127.0.0.1",
		ExternalIngressPort: port,
		InternalIngressPort: "80",
		CookieEncryptionKey: cookieEncryptionKey,
		CookieSigningKey:    cookieSigningKey,
		K8sApiPort:          "6443",
		K8sApiCaCertB64:     encodedCert,
	}

	manifests, err := renderTemplate(bindBackendManifestsTemplate, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not render the klutch-bind backend manifests.")
	}

	makeup.PrintH2("Creating a kind cluster with following config: ")
	makeup.PrintYAML(manifests.Bytes(), false)
	makeup.WaitForUser(demo.UnattendedMode)

	k.mgmtK8s.KubectlApplyStdin(manifests)

	makeup.Print("klutch-bind backend applied.")
}

// Waits for the klutch-bind backend deployment to be ready.
// Note: the manifests contain an init-container which waits for dex to be ready,
// because the backend requires dex to be up and running in order to start.
// This avoids delays/complications due to crash loop backoffs.
func (k *KlutchManager) WaitForBindBackend() {
	makeup.PrintH1("Waiting for the klutch-bind backend to become ready...")

	k.mgmtK8s.KubectlWaitForRollout("deployment", "anynines-backend", "default")

	makeup.PrintCheckmark("The klutch-bind backend appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}
