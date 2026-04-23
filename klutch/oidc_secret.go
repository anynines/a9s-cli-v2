package klutch

import (
	"fmt"

	"github.com/anynines/a9s-cli-v2/makeup"
)

// applyOIDCSecret ensures the oidc-config secret exists with the provided external OIDC values.
func (k *KlutchManager) applyOIDCSecret(opts OIDCOptions) {
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: oidc-config
  namespace: default
type: Opaque
stringData:
  oidc-issuer-client-id: "%s"
  oidc-issuer-client-secret: "%s"
  oidc-issuer-url: "%s"
  oidc-callback-url: "%s"
`, opts.ClientID, opts.ClientSecret, opts.IssuerURL, opts.CallbackURL)

	// Note: Manifest display and waiting are handled by KubectlApplyWithPrompt
	if _, err := k.cpK8s.ApplyWithPrompt([]byte(manifest), "external OIDC configuration secret"); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply external OIDC configuration secret")
	}
}
