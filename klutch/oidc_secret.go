package klutch

import (
	"bytes"
	"fmt"

	"github.com/anynines/a9s-cli-v2/demo"
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

	makeup.PrintH2("Applying external OIDC configuration secret...")
	makeup.PrintYAML([]byte(manifest), false)
	makeup.WaitForUser(demo.UnattendedMode)

	k.cpK8s.KubectlApplyStdin(bytes.NewBufferString(manifest))

	makeup.Print("Applied external OIDC configuration.")
}
