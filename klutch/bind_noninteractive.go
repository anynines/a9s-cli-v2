package klutch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/klutchio/bind/deploy/crd"
	"github.com/anynines/klutchio/bind/deploy/konnector"
	bindv1alpha1 "github.com/anynines/klutchio/bind/pkg/apis/bind/v1alpha1"
	bindclient "github.com/anynines/klutchio/bind/pkg/client/clientset/versioned"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

// NonInteractiveBindOptions controls the non-interactive workload bind flow.
type NonInteractiveBindOptions struct {
	ControlPlaneURL    string
	BindRequestPath    string
	BindRequestData    []byte
	OIDCClientID       string
	OIDCClientSecret   string
	OIDCTokenURL       string
	OIDCScope          string
	KonnectorImage     string
	WriteKubeconfigTo  string
	WorkloadKubeconfig string
	WorkloadContext    string
}

// bindRequest mirrors the backend BindRequest payload.
type bindRequest struct {
	ClusterID string                 `json:"clusterID"`
	Apis      []metav1.GroupResource `json:"apis"`
}

// NonInteractiveBind executes the helper workflow inline: requests a binding via the backend,
// applies the returned manifests to the workload cluster, and creates export requests on the control plane.
func NonInteractiveBind(ctx context.Context, opts NonInteractiveBindOptions) error {
	if strings.TrimSpace(opts.ControlPlaneURL) == "" {
		return fmt.Errorf("control-plane URL is required")
	}
	if strings.TrimSpace(opts.BindRequestPath) == "" {
		return fmt.Errorf("bind request file (--bind-request-file) is required")
	}

	clientID := firstNonEmpty(opts.OIDCClientID, os.Getenv("OIDC_CLIENT_ID"))
	clientSecret := firstNonEmpty(opts.OIDCClientSecret, os.Getenv("OIDC_CLIENT_SECRET"))
	tokenURL := firstNonEmpty(opts.OIDCTokenURL, os.Getenv("OIDC_TOKEN_URL"))
	if clientID == "" || clientSecret == "" || tokenURL == "" {
		return fmt.Errorf("OIDC client ID, client secret, and token URL are required (flags or OIDC_CLIENT_ID/OIDC_CLIENT_SECRET/OIDC_TOKEN_URL)")
	}

	scope := opts.OIDCScope
	if strings.TrimSpace(scope) == "" {
		scope = "openid profile email offline_access"
	}

	reqBytes := opts.BindRequestData
	if len(reqBytes) == 0 && strings.TrimSpace(opts.BindRequestPath) != "" {
		var err error
		reqBytes, err = os.ReadFile(opts.BindRequestPath)
		if err != nil {
			return fmt.Errorf("failed to read bind request file: %w", err)
		}
	}
	if len(reqBytes) == 0 {
		return fmt.Errorf("bind request is required (from secret or --bind-request-file)")
	}
	var br bindRequest
	if err := json.Unmarshal(reqBytes, &br); err != nil {
		return fmt.Errorf("bind request is invalid: %w", err)
	}

	targetURL := opts.ControlPlaneURL
	if !strings.Contains(targetURL, "bind-noninteractive") {
		targetURL = strings.TrimRight(targetURL, "/") + "/bind-noninteractive"
	}

	makeup.PrintInfo(fmt.Sprintf("Requesting non-interactive bind from %s ...", targetURL))
	tokenCfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       strings.Fields(scope),
	}
	tkn, err := tokenCfg.Token(ctx)
	if err != nil {
		return fmt.Errorf("failed to obtain token from %s: %w", tokenURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tkn.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tkn))
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call backend: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read backend response: %w", err)
	}
	var bindingResp bindv1alpha1.BindingResponse
	if err := json.Unmarshal(body, &bindingResp); err != nil {
		return fmt.Errorf("failed to parse backend response: %w", err)
	}

	if opts.WriteKubeconfigTo != "" {
		if err := os.WriteFile(opts.WriteKubeconfigTo, bindingResp.Kubeconfig, 0600); err != nil {
			return fmt.Errorf("failed to write kubeconfig to %s: %w", opts.WriteKubeconfigTo, err)
		}
		makeup.PrintCheckmark(fmt.Sprintf("Wrote control-plane kubeconfig to %s", opts.WriteKubeconfigTo))
	}

	manifests, exportRequests, err := buildManifests(bindingResp, firstNonEmpty(opts.KonnectorImage, konnectorImage))
	if err != nil {
		return err
	}

	if err := applyManifests(ctx, manifests, opts.WorkloadKubeconfig, opts.WorkloadContext); err != nil {
		return err
	}
	makeup.PrintCheckmark("Applied klutch-bind resources to the workload cluster.")

	if err := createExportRequests(ctx, bindingResp.Kubeconfig, exportRequests); err != nil {
		return err
	}
	makeup.PrintCheckmark("Submitted export requests to the control plane.")

	return nil
}

// buildManifests assembles the YAML to apply on the workload cluster and returns the export requests for the control plane.
func buildManifests(resp bindv1alpha1.BindingResponse, konnectorImg string) (string, []bindv1alpha1.APIServiceExportRequest, error) {
	ns := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: corev1.SchemeGroupVersion.Version},
		ObjectMeta: metav1.ObjectMeta{
			Name: "klutch-bind",
		},
	}

	kfgSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "klutch-bind",
			Name:      "kubeconfig",
		},
		Data: map[string][]byte{"kubeconfig": resp.Kubeconfig},
	}

	crds, err := crd.CRDs()
	if err != nil {
		return "", nil, fmt.Errorf("failed to load CRDs: %w", err)
	}

	konnectorManifests, err := konnector.Bytes(konnectorImg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to load konnector manifests: %w", err)
	}

	apiBindings, exportReqs, err := parseBindingResponse(resp)
	if err != nil {
		return "", nil, err
	}

	var docs []string
	for _, obj := range []metav1.Object{&ns, &kfgSecret} {
		yml, err := yamlStr(obj)
		if err != nil {
			return "", nil, err
		}
		docs = append(docs, yml)
	}

	for _, cr := range crds {
		yml, err := yamlStr(&cr)
		if err != nil {
			return "", nil, err
		}
		docs = append(docs, yml)
	}

	for _, b := range apiBindings {
		yml, err := yamlStr(&b)
		if err != nil {
			return "", nil, err
		}
		docs = append(docs, yml)
	}

	for _, km := range konnectorManifests {
		docs = append(docs, string(km))
	}

	return strings.Join(docs, "\n---\n"), exportReqs, nil
}

func parseBindingResponse(resp bindv1alpha1.BindingResponse) ([]bindv1alpha1.APIServiceBinding, []bindv1alpha1.APIServiceExportRequest, error) {
	var bindings []bindv1alpha1.APIServiceBinding
	var exportReqs []bindv1alpha1.APIServiceExportRequest

	for _, raw := range resp.Requests {
		var meta metav1.TypeMeta
		if err := json.Unmarshal(raw.Raw, &meta); err != nil {
			return nil, nil, fmt.Errorf("failed to parse binding response: %w", err)
		}

		var apiReq bindv1alpha1.APIServiceExportRequestResponse
		if err := json.Unmarshal(raw.Raw, &apiReq); err != nil {
			return nil, nil, fmt.Errorf("failed to parse APIServiceExportRequestResponse: %w", err)
		}

		for _, res := range apiReq.Spec.Resources {
			var claims []bindv1alpha1.AcceptablePermissionClaim
			for _, c := range res.PermissionClaims {
				claims = append(claims, bindv1alpha1.AcceptablePermissionClaim{
					PermissionClaim: c,
					State:           bindv1alpha1.ClaimAccepted,
				})
			}
			bindings = append(bindings, bindv1alpha1.APIServiceBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "APIServiceBinding",
					APIVersion: bindv1alpha1.SchemeGroupVersion.Identifier(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "klutch-bind",
					Name:      res.Resource + "." + res.Group,
				},
				Spec: bindv1alpha1.APIServiceBindingSpec{
					KubeconfigSecretRef: bindv1alpha1.ClusterSecretKeyRef{
						LocalSecretKeyRef: bindv1alpha1.LocalSecretKeyRef{Name: "kubeconfig", Key: "kubeconfig"},
						Namespace:         "klutch-bind",
					},
					PermissionClaims: claims,
				},
			})
		}

		var req bindv1alpha1.APIServiceExportRequest
		if err := json.Unmarshal(raw.Raw, &req); err != nil {
			return nil, nil, fmt.Errorf("failed to parse export request: %w", err)
		}
		req.ObjectMeta.GenerateName = req.ObjectMeta.Name
		req.ObjectMeta.Name = ""
		exportReqs = append(exportReqs, req)
	}

	return bindings, exportReqs, nil
}

func applyManifests(ctx context.Context, manifest string, kubeconfigPath, kubeContext string) error {
	cmdArgs := []string{"apply", "-f", "-"}
	if strings.TrimSpace(kubeconfigPath) != "" {
		cmdArgs = append(cmdArgs, "--kubeconfig", kubeconfigPath)
	}
	if strings.TrimSpace(kubeContext) != "" {
		cmdArgs = append(cmdArgs, "--context", kubeContext)
	}

	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	cmd.Stdin = bytes.NewBufferString(manifest)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply workload manifests: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	if makeup.Verbose {
		fmt.Println(string(out))
	}
	return nil
}

func createExportRequests(ctx context.Context, kubeconfig []byte, requests []bindv1alpha1.APIServiceExportRequest) error {
	kfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to parse returned kubeconfig: %w", err)
	}

	currentCtx := kfg.CurrentContext
	if currentCtx == "" || kfg.Contexts[currentCtx] == nil || kfg.Contexts[currentCtx].Namespace == "" {
		return fmt.Errorf("returned kubeconfig missing current context/namespace")
	}
	ns := kfg.Contexts[currentCtx].Namespace

	restCfg, err := clientcmd.NewDefaultClientConfig(*kfg, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build client config from returned kubeconfig: %w", err)
	}

	client, err := bindclient.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to build bind client: %w", err)
	}

	for _, r := range requests {
		r := r
		_, err := client.KlutchBindV1alpha1().APIServiceExportRequests(ns).Create(ctx, &r, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create export request %s in %s: %w", r.GenerateName, ns, err)
		}
	}
	return nil
}

func yamlStr(obj metav1.Object) (string, error) {
	k, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(k), nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
