package klutch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/anynines/klutchio/bind/deploy/crd"
	"github.com/anynines/klutchio/bind/deploy/konnector"
	bindv1alpha1 "github.com/anynines/klutchio/bind/pkg/apis/bind/v1alpha1"
	bindclient "github.com/anynines/klutchio/bind/pkg/client/clientset/versioned"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	corev1 "k8s.io/api/core/v1"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const DefaultControlPlaneClusterName = "klutch-control-plane"

const (
	nonInteractiveBindMaxAttempts   = 5
	nonInteractiveBindRetryInterval = 10 * time.Second
)

// NonInteractiveBindOptions controls the non-interactive workload bind flow.
type NonInteractiveBindOptions struct {
	ControlPlaneURL         string
	BindRequestPath         string
	BindRequestData         []byte
	OIDCClientID            string
	OIDCClientSecret        string
	OIDCTokenURL            string
	OIDCScope               string
	KonnectorImage          string
	WriteKubeconfigTo       string
	WorkloadKubeconfigPath  string
	WorkloadContext         string
	ControlPlaneClusterName string
}

// bindRequest mirrors the backend BindRequest payload.
type bindRequest struct {
	ClusterID string                 `json:"clusterID"`
	Apis      []metav1.GroupResource `json:"apis"`
}

type backendBindError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e backendBindError) Error() string {
	return fmt.Sprintf("backend returned %s: %s", e.Status, strings.TrimSpace(e.Body))
}

// ValidateBindRequest ensures the payload is valid JSON and has required fields.
func ValidateBindRequest(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("bind request is required (from tenant secret or --bind-request-file)")
	}
	_, err := parseBindRequest(data)
	return err
}

// NonInteractiveBind executes the helper workflow inline: requests a binding via the backend,
// applies the returned manifests to the workload cluster, and creates export requests on the control plane.
func NonInteractiveBind(ctx context.Context, opts NonInteractiveBindOptions) error {
	if strings.TrimSpace(opts.ControlPlaneURL) == "" {
		return fmt.Errorf("control-plane URL is required")
	}
	if len(opts.BindRequestData) == 0 && strings.TrimSpace(opts.BindRequestPath) == "" {
		return fmt.Errorf("bind request is required (from tenant secret or --bind-request-file)")
	}

	clientID := firstNonEmpty(opts.OIDCClientID, os.Getenv("OIDC_CLIENT_ID"))
	clientSecret := firstNonEmpty(opts.OIDCClientSecret, os.Getenv("OIDC_CLIENT_SECRET"))
	tokenURL := firstNonEmpty(opts.OIDCTokenURL, os.Getenv("OIDC_TOKEN_URL"))
	if clientID == "" || clientSecret == "" || tokenURL == "" {
		return fmt.Errorf("OIDC client ID, client secret, and token URL are required (flags or OIDC_CLIENT_ID/OIDC_CLIENT_SECRET/OIDC_TOKEN_URL)")
	}

	scope := opts.OIDCScope
	if strings.TrimSpace(scope) == "" {
		return fmt.Errorf("OIDC scope is required")
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
		return fmt.Errorf("bind request is required (from tenant secret or --bind-request-file)")
	}
	if _, err := parseBindRequest(reqBytes); err != nil {
		return err
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

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tkn))
	body, err := callNonInteractiveBindWithRetry(ctx, httpClient, targetURL, reqBytes, tkn.AccessToken)
	if err != nil {
		return err
	}
	var bindingResp bindv1alpha1.BindingResponse
	if err := json.Unmarshal(body, &bindingResp); err != nil {
		return fmt.Errorf("failed to parse backend response: %w", err)
	}

	clusterName := strings.TrimSpace(opts.ControlPlaneClusterName)
	if clusterName == "" {
		clusterName = DefaultControlPlaneClusterName
	}

	regionHint := strings.TrimSpace(os.Getenv("CONTROL_PLANE_CLUSTER_REGION"))
	if patched, err := ensureKubeconfigCA(bindingResp.Kubeconfig, clusterName, regionHint); err == nil {
		bindingResp.Kubeconfig = patched
	} else {
		makeup.PrintWarning(fmt.Sprintf("Could not ensure control-plane CA in kubeconfig: %v", err))
	}

	if opts.WriteKubeconfigTo != "" {
		if err := os.WriteFile(opts.WriteKubeconfigTo, bindingResp.Kubeconfig, 0600); err != nil {
			return fmt.Errorf("failed to write kubeconfig to %s: %w", opts.WriteKubeconfigTo, err)
		}
		makeup.PrintCheckmark(fmt.Sprintf("Wrote control-plane kubeconfig to %s", opts.WriteKubeconfigTo))
	}

	crdManifests, manifests, exportRequests, err := buildManifests(bindingResp, firstNonEmpty(opts.KonnectorImage, konnectorImage))
	if err != nil {
		return err
	}

	k8sClient := k8s.NewKubeClient(opts.WorkloadContext)
	if strings.TrimSpace(crdManifests) != "" {
		if _, err := k8sClient.ApplyWithPrompt([]byte(crdManifests), "Binding CRDs"); err != nil {
			return err
		}
		if err := waitForAPIServiceBinding(ctx, opts.WorkloadKubeconfigPath, opts.WorkloadContext); err != nil {
			return err
		}
	}
	if _, err := k8sClient.ApplyWithPrompt([]byte(manifests), "Binding Resources"); err != nil {
		return err
	}
	makeup.PrintCheckmark("Applied klutch-bind resources to the workload cluster.")

	if err := createExportRequests(ctx, bindingResp.Kubeconfig, exportRequests); err != nil {
		return err
	}
	makeup.PrintCheckmark("Submitted export requests to the control plane.")

	return nil
}

func callNonInteractiveBindWithRetry(ctx context.Context, httpClient *http.Client, targetURL string, reqBytes []byte, accessToken string) ([]byte, error) {
	var lastErr error

	for attempt := 1; attempt <= nonInteractiveBindMaxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(reqBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to call backend: %w", err)
		} else {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("failed to read backend response: %w", readErr)
			} else if resp.StatusCode == http.StatusOK {
				return body, nil
			} else {
				lastErr = backendBindError{
					StatusCode: resp.StatusCode,
					Status:     resp.Status,
					Body:       string(body),
				}
			}
		}

		if !shouldRetryNonInteractiveBind(lastErr) || attempt == nonInteractiveBindMaxAttempts {
			return nil, lastErr
		}

		makeup.PrintWarning(fmt.Sprintf("Bind backend request failed (attempt %d/%d): %v. Retrying in %s...", attempt, nonInteractiveBindMaxAttempts, lastErr, nonInteractiveBindRetryInterval))
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(nonInteractiveBindRetryInterval):
		}
	}

	return nil, lastErr
}

func shouldRetryNonInteractiveBind(err error) bool {
	var bindErr backendBindError
	if errors.As(err, &bindErr) {
		return bindErr.StatusCode >= http.StatusInternalServerError || bindErr.StatusCode == http.StatusTooManyRequests
	}
	return true
}

// ensureKubeconfigCA injects control-plane CA data to avoid TLS errors.
func ensureKubeconfigCA(kubeconfig []byte, clusterName, regionHint string) ([]byte, error) {
	cfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	changed := false

	// Determine region from env or first cluster server.
	region := strings.TrimSpace(regionHint)
	if region == "" {
		for _, c := range cfg.Clusters {
			if c != nil {
				region = controlPlaneRegionFromURL(c.Server)
				if region != "" {
					break
				}
			}
		}
	}
	if env := strings.TrimSpace(os.Getenv("CONTROL_PLANE_CLUSTER_REGION")); env != "" {
		region = env
	}
	if region == "" {
		return nil, fmt.Errorf("could not infer control-plane region")
	}

	// Fetch CA for the explicit control-plane cluster name, fallback to kubeconfig name.
	var ca []byte
	var fetchErr error
	if cn := strings.TrimSpace(clusterName); cn != "" {
		ca, fetchErr = fetchClusterCA(cn, region)
	}
	if fetchErr != nil || len(ca) == 0 {
		return nil, fetchErr
	}
	if len(ca) == 0 {
		return nil, fmt.Errorf("empty CA data for control-plane cluster")
	}

	for name, c := range cfg.Clusters {
		if c == nil || c.InsecureSkipTLSVerify {
			continue
		}
		if !sameCA(ca, c.CertificateAuthorityData) {
			c.CertificateAuthorityData = ca
			cfg.Clusters[name] = c
			changed = true
		}
	}
	if !changed {
		return kubeconfig, nil
	}
	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return nil, fmt.Errorf("write kubeconfig: %w", err)
	}
	return out, nil
}

func controlPlaneRegionFromURL(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	parts := strings.Split(host, ".")
	for i, p := range parts {
		if p == "eks" && i > 0 {
			return parts[i-1]
		}
	}
	if len(parts) >= 3 {
		return parts[len(parts)-3]
	}
	return ""
}

func fetchClusterCA(clusterName, region string) ([]byte, error) {
	out, err := makeup.NewCommand("aws", "eks", "describe-cluster",
		"--name", clusterName,
		"--region", region,
		"--query", "cluster.certificateAuthority.data",
		"--output", "text").NoPrompt().Run()
	if err != nil {
		return nil, fmt.Errorf("fetch CA for cluster %s in %s: %w", clusterName, region, err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	pem, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode CA for cluster %s: %w", clusterName, err)
	}
	return pem, nil
}

func sameCA(expected, existing []byte) bool {
	e := bytes.TrimSpace(expected)
	c := bytes.TrimSpace(existing)
	if len(e) == 0 && len(c) == 0 {
		return true
	}
	return sha256.Sum256(e) == sha256.Sum256(c)
}

func parseBindRequest(data []byte) (bindRequest, error) {
	var br bindRequest
	if len(data) == 0 {
		return br, fmt.Errorf("bind request is required (from tenant secret or --bind-request-file)")
	}
	if err := json.Unmarshal(data, &br); err != nil {
		return br, fmt.Errorf("bind request is invalid JSON: %w", err)
	}
	if strings.TrimSpace(br.ClusterID) == "" {
		return br, fmt.Errorf("bind request is missing clusterID")
	}
	if len(br.Apis) == 0 {
		return br, fmt.Errorf("bind request has no apis")
	}
	return br, nil
}

// buildManifests assembles the YAML to apply on the workload cluster and returns the export requests for the control plane.
// CRDs are returned separately to allow applying them before the remaining resources.
func buildManifests(resp bindv1alpha1.BindingResponse, konnectorImg string) (string, string, []bindv1alpha1.APIServiceExportRequest, error) {
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
		return "", "", nil, fmt.Errorf("failed to load CRDs: %w", err)
	}

	konnectorManifests, err := konnector.Bytes(konnectorImg)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to load konnector manifests: %w", err)
	}

	apiBindings, exportReqs, err := parseBindingResponse(resp)
	if err != nil {
		return "", "", nil, err
	}

	var docs []string
	for _, obj := range []metav1.Object{&ns, &kfgSecret} {
		yml, err := yamlStr(obj)
		if err != nil {
			return "", "", nil, err
		}
		docs = append(docs, yml)
	}

	for _, b := range apiBindings {
		yml, err := yamlStr(&b)
		if err != nil {
			return "", "", nil, err
		}
		docs = append(docs, yml)
	}

	for _, km := range konnectorManifests {
		docs = append(docs, string(km))
	}

	return strings.Join(crdDocs(crds), "\n---\n"), strings.Join(docs, "\n---\n"), exportReqs, nil
}

func crdDocs(crds []apixv1.CustomResourceDefinition) []string {
	var docs []string
	for _, cr := range crds {
		yml, err := yamlStr(&cr)
		if err == nil {
			docs = append(docs, yml)
		}
	}
	return docs
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

func waitForAPIServiceBinding(ctx context.Context, kubeconfigPath, kubeContext string) error {
	k8sClient := k8s.NewKubeClient("")
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("APIServiceBinding CRD not registered in time")
		case <-tick:
			out, err := k8sClient.ApiResources("klutch.anynines.com", kubeconfigPath, kubeContext, "name")
			if err == nil && strings.Contains(string(out), "apiservicebindings") {
				return nil
			}
		}
	}
}

func createExportRequests(ctx context.Context, kubeconfig []byte, requests []bindv1alpha1.APIServiceExportRequest) error {
	kfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to parse returned kubeconfig: %w", err)
	}

	currentCtx := kfg.CurrentContext
	if currentCtx == "" || kfg.Contexts[currentCtx] == nil {
		return fmt.Errorf("returned kubeconfig missing current context:\n%s", string(kubeconfig))
	}
	if kfg.Contexts[currentCtx].Namespace == "" {
		return fmt.Errorf("returned kubeconfig missing current namespace:\n%s", string(kubeconfig))
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
		if _, err := client.KlutchBindV1alpha1().APIServiceExportRequests(ns).Create(ctx, &r, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create export request %s in %s: %w", r.GenerateName, ns, err)
		}
	}
	return nil
}

func ensureControlPlaneCRDs(ctx context.Context, restCfg *rest.Config) error {
	// CRDs must already exist on the control-plane; the returned kubeconfig typically cannot create them.
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
