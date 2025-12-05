package klutch

import (
	"fmt"
	"strings"
)

type OIDCProvider string

const (
	OIDCProviderDex     OIDCProvider = "dex"
	OIDCProviderCognito OIDCProvider = "cognito"
)

// OIDCOptions contains external OIDC configuration. When Provider is left empty,
// the defaults are applied based on the selected Kubernetes provider (Dex for
// local/kind; Cognito for AWS).
type OIDCOptions struct {
	Provider     OIDCProvider
	IssuerURL    string
	ClientID     string
	ClientSecret string
	CallbackURL  string
}

var controlPlaneOIDCOptions OIDCOptions

// SetControlPlaneOIDCOptions stores OIDC options provided via CLI flags before applying the control plane.
func SetControlPlaneOIDCOptions(opts OIDCOptions) {
	controlPlaneOIDCOptions = opts.normalize()
}

// effectiveOIDCOptions applies defaults and normalization based on the Kubernetes provider and callback host.
func effectiveOIDCOptions(provider string, scheme string, callbackHost string) OIDCOptions {
	resolved := controlPlaneOIDCOptions.normalize()
	if resolved.Provider == "" {
		resolved.Provider = defaultOIDCProvider(provider)
	}

	if resolved.CallbackURL == "" && callbackHost != "" {
		resolved.CallbackURL = fmt.Sprintf("%s://%s/callback", scheme, callbackHost)
	}

	return resolved
}

func defaultOIDCProvider(kubernetesProvider string) OIDCProvider {
	switch strings.ToLower(strings.TrimSpace(kubernetesProvider)) {
	case "aws":
		return OIDCProviderCognito
	default:
		return OIDCProviderDex
	}
}

func (o OIDCOptions) normalize() OIDCOptions {
	o.Provider = OIDCProvider(strings.ToLower(strings.TrimSpace(string(o.Provider))))
	o.IssuerURL = strings.TrimSpace(o.IssuerURL)
	o.ClientID = strings.TrimSpace(o.ClientID)
	o.ClientSecret = strings.TrimSpace(o.ClientSecret)
	o.CallbackURL = strings.TrimSpace(o.CallbackURL)
	return o
}

func (o OIDCOptions) validate() error {
	switch o.Provider {
	case OIDCProviderDex:
		return nil
	case OIDCProviderCognito:
		if o.IssuerURL == "" {
			return fmt.Errorf("oidc issuer url is required for provider %q", o.Provider)
		}
		if o.ClientID == "" {
			return fmt.Errorf("oidc client id is required for provider %q", o.Provider)
		}
		if o.ClientSecret == "" {
			return fmt.Errorf("oidc client secret is required for provider %q", o.Provider)
		}
		return nil
	case "":
		return fmt.Errorf("oidc provider must be set")
	default:
		return fmt.Errorf("unsupported oidc provider %q", o.Provider)
	}
}
