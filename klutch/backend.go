package klutch

import (
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
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
	BackendImageURL     string
	BackendImageTag     string
	ExternalCASecret    string
	ExternalAddress     string
	IngressPort         string
	K8sApiCaCertB64     string
	IngressClass        string
	Scheme              string
	ACMCertificateARN   string
	EnableTLS           bool
	ServiceType         string
	WaitForDex          bool
}

const (
	defaultBackendImageURL = "public.ecr.aws/w5n9a2g2/anynines/kubebind-backend"
	defaultBackendImageTag = "dev5"
	externalCASecretName   = "klutch-bind-external-ca"
)

var (
	bindBackendImageURL = defaultBackendImageURL
	bindBackendImageTag = defaultBackendImageTag
)

// SetBindBackendImage overrides the backend image URL/tag when provided (non-empty).
func SetBindBackendImage(imageURL, imageTag string) {
	if imageURL = strings.TrimSpace(imageURL); imageURL != "" {
		bindBackendImageURL = imageURL
	}
	if imageTag = strings.TrimSpace(imageTag); imageTag != "" {
		bindBackendImageTag = imageTag
	}
}

// createExternalCASecret fetches the ACM certificate chain and creates a secret for the backend to mount.
func createExternalCASecret(acmArn string, k *k8s.KubeClient) (string, error) {
	arn := strings.TrimSpace(acmArn)
	if arn == "" {
		return "", fmt.Errorf("empty ACM ARN")
	}
	// Fetch cert + chain; concatenating covers intermediates.
	cmd := exec.Command("aws", "acm", "get-certificate", "--certificate-arn", arn, "--query", "Certificate")
	certOut, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch ACM certificate: %w", err)
	}
	cmd = exec.Command("aws", "acm", "get-certificate", "--certificate-arn", arn, "--query", "CertificateChain")
	chainOut, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch ACM certificate chain: %w", err)
	}
	// Remove surrounding quotes from JSON string output.
	clean := func(b []byte) string {
		s := strings.TrimSpace(string(b))
		s = strings.TrimPrefix(s, "\"")
		s = strings.TrimSuffix(s, "\"")
		s = strings.ReplaceAll(s, `\n`, "\n")
		return s
	}
	pem := strings.TrimSpace(clean(certOut) + "\n" + clean(chainOut))
	if pem == "" {
		return "", fmt.Errorf("empty ACM certificate data")
	}

	manifest := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: default
type: Opaque
stringData:
  ca.crt: |
%s
`, externalCASecretName, indent(pem, "    "))

	if _, err = k.ApplyWithPrompt([]byte(manifest), "external CA secret"); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply external CA secret")
	}
	return externalCASecretName, nil
}

func indent(text, pad string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

// Deploys dex (if required by OIDC provider) and the klutch-bind backend.
func (k *KlutchManager) DeployBindBackend(host string, ingressPort string, ingressClass string, scheme string, acmCertificateARN string, enableTLS bool, waitForDex bool) {
	makeup.PrintH1("Deploying the klutch-bind backend...")

	makeup.PrintH2("Applying the klutch-bind backend CRDs...")

	if _, err := k.cpK8s.ApplyFromUrl(backendCRDManifestsURL, backendCRDManifestsURL); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply klutch-bind backend CRDs")
	}

	makeup.PrintCheckmark("klutch-bind backend CRDs applied.")
	makeup.PrintH2("Applying the klutch-bind backend manifests...")

	// We need the provider cluster's CA certificate
	clusterCert := getClusterCert(k.cpK8s)
	encodedCert := base64.StdEncoding.EncodeToString(clusterCert)

	rbg := RandomByteGenerator{}
	cookieSigningKey := rbg.GenerateRandom32BytesBase64()
	cookieEncryptionKey := rbg.GenerateRandom32BytesBase64()
	// Use the control-plane Kubernetes API address for generated kubeconfigs (not the backend ingress).
	apiAddress := getClusterExternalAddress(k.cpContext)
	externalAddress := apiAddress

	// Only mount an external CA when the backend targets a different endpoint than the control-plane API.
	externalCASecret := ""
	if strings.TrimSpace(acmCertificateARN) != "" && externalAddress != apiAddress {
		if secret, err := createExternalCASecret(acmCertificateARN, k.cpK8s); err != nil {
			makeup.PrintWarning(fmt.Sprintf("Could not create external CA secret from ACM certificate: %v", err))
		} else {
			externalCASecret = secret
			makeup.PrintInfo(fmt.Sprintf("Created external CA secret %s for backend TLS validation.", secret))
		}
	}

	templateVars := &backendTemplateVars{
		CookieEncryptionKey: cookieEncryptionKey,
		CookieSigningKey:    cookieSigningKey,
		BackendImageURL:     bindBackendImageURL,
		BackendImageTag:     bindBackendImageTag,
		ExternalCASecret:    externalCASecret,
		ExternalAddress:     externalAddress,
		IngressPort:         ingressPort,
		K8sApiCaCertB64:     encodedCert,
		IngressClass:        ingressClass,
		Scheme:              scheme,
		ACMCertificateARN:   acmCertificateARN,
		EnableTLS:           enableTLS,
		WaitForDex:          waitForDex,
	}

	if ingressClass == "alb" {
		templateVars.ServiceType = "NodePort"
	} else {
		templateVars.ServiceType = "ClusterIP"
	}

	manifests, err := renderTemplate(bindBackendManifestsTemplate, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not render the klutch-bind backend manifests.")
	}

	// Note: Manifest display and waiting are handled by KubectlApplyWithPrompt
	if _, err = k.cpK8s.ApplyWithPrompt(manifests.Bytes(), "klutch-bind backend"); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply klutch-bind backend manifests")
	}

	makeup.Print("klutch-bind backend applied.")
}

// Waits for the klutch-bind backend deployment to be ready.
// Note: the manifests contain an init-container which waits for dex to be ready,
// because the backend requires dex to be up and running in order to start.
// This avoids delays/complications due to crash loop backoffs.
func (k *KlutchManager) WaitForBindBackend(host string, port string, scheme string) {
	makeup.PrintH1("Waiting for the klutch-bind backend to become ready...")

	waitForHostReachable(host, port, 10*time.Minute)

	k.cpK8s.KubectlWaitForRollout("deployment", "anynines-backend", "default")

	verifyBindEndpoint(host, port, scheme, 10*time.Minute)

	makeup.PrintCheckmark("The klutch-bind backend appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}

// verifyBindEndpoint checks that /export returns a non-empty payload.
func verifyBindEndpoint(host, port, scheme string, timeout time.Duration) {
	if host == "" || port == "" {
		return
	}

	base := fmt.Sprintf("%s://%s", scheme, host)
	if !(scheme == "https" && port == "443") && !(scheme == "http" && port == "80") {
		base = fmt.Sprintf("%s:%s", base, port)
	}
	url := fmt.Sprintf("%s/export", strings.TrimRight(base, "/"))

	deadline := time.Now().Add(timeout)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: insecureTLSConfig(),
		},
	}

	for {
		resp, err := client.Get(url)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && len(body) > 0 {
				return
			}
		}

		if time.Now().After(deadline) {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("klutch-bind endpoint %s did not return a valid response within %s", url, timeout))
		}

		makeup.PrintInfo(fmt.Sprintf("klutch-bind endpoint %s not ready yet. Waiting...", url))
		time.Sleep(5 * time.Second)
	}
}

func insecureTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
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
