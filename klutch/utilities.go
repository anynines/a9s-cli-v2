package klutch

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// switchContext changes the current kubeconfig context.
func switchContext(context string) {
	cmd := exec.Command("kubectl", "config", "use-context", context)
	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not switch context: %s", string(output)))
	}
}

// generateRandom32BytesBase64 returns a random base64 string of length 32.
func generateRandom32BytesBase64() string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Unknown error generating random values")
	}
	str := base64.StdEncoding.EncodeToString(randomBytes)
	return str
}

// getClusterCert extracts the kube API's CA certificate.
func getClusterCert(k8s *k8s.KubeClient) []byte {
	cmd := k8s.KubectlWithContextCommand("get", "configmap", "--namespace", "kube-system", "kube-root-ca.crt", "-o", "jsonpath={.data.ca\\.crt}")
	output, err := cmd.Output()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Error getting cluster cert")
	}
	return output
}

// getClusterExternalPort extracts the given context's kubernetes API port from the kubeconfig file.
// If the port is absent or can't be used, returns 80 or 443 depending on the found URL scheme.
// TODO: unit test. abstract the config loading so it can be mocked.
func getClusterExternalPort(kubeContext string) string {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Error loading kubeconfig file %s", kubeconfig))
	}

	ctx, exists := config.Contexts[kubeContext]
	if !exists {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("context %s not found in kubeconfig", kubeContext))
	}

	cluster, exists := config.Clusters[ctx.Cluster]
	if !exists {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("cluster %s not found in kubeconfig", ctx.Cluster))
	}

	url, err := url.Parse(cluster.Server)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("cluster url %s cannot be parsed", cluster.Server))
	}

	port := url.Port()
	if port == "" {
		if url.Scheme == "https" {
			return "443"
		} else if url.Scheme == "http" {
			return "80"
		} else {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("cannot determine port: unknown url scheme %s", url.Scheme))
		}
	}

	return port
}

// renderTemplate renders the given go-template and returns a buffer containing the result.
func renderTemplate(rawTemplate string, vars interface{}) (*bytes.Buffer, error) {
	parsedTemplate, err := template.New("").Parse(rawTemplate)
	if err != nil {
		return nil, fmt.Errorf("unexpected error occured while parsing template: %v", err)
	}

	w := &bytes.Buffer{}
	err = parsedTemplate.Execute(w, vars)
	if err != nil {
		return nil, fmt.Errorf("unexpected error occured while rendering template: %v", err)
	}

	return w, nil
}

// printCommandOutput reads and prints line by line.
func printCommandOutput(r io.ReadCloser) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println("\r" + scanner.Text())
	}
}
