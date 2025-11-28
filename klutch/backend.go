package klutch

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"net"
	"time"

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
func (k *KlutchManager) WaitForBindBackend(host string, port string) {
	makeup.PrintH1("Waiting for the klutch-bind backend to become ready...")

	waitForHostReachable(host, port, 10*time.Minute)

	k.cpK8s.KubectlWaitForRollout("deployment", "anynines-backend", "default")

	makeup.PrintCheckmark("The klutch-bind backend appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}

// waitForHostReachable waits until the host resolves and a TCP connection to host:port succeeds.
func waitForHostReachable(host string, port string, timeout time.Duration) {
	if host == "" || port == "" {
		return
	}

	deadline := time.Now().Add(timeout)
	for {
		_, err := net.LookupHost(host)
		if err != nil {
			if time.Now().After(deadline) {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("DNS for %s did not resolve within %s", host, timeout))
			}
			makeup.PrintInfo(fmt.Sprintf("Host %s not resolvable yet (%v). Waiting...", host, err))
			time.Sleep(5 * time.Second)
			continue
		}

		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 5*time.Second)
		if err == nil {
			_ = conn.Close()
			return
		}

		if time.Now().After(deadline) {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not reach %s:%s within %s", host, port, timeout))
		}

		makeup.PrintInfo(fmt.Sprintf("Host %s:%s not reachable yet (%v). Waiting...", host, port, err))
		time.Sleep(5 * time.Second)
	}
}
