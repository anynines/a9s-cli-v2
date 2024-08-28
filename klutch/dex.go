package klutch

import (
	_ "embed"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
)

//go:embed templates/dex.tmpl
var dexManifestsTemplate string

type dexTemplateVars struct {
	Host            string
	IngressPort     string
	DexClientSecret string
}

func (k *KlutchManager) DeployDex(hostIP string, ingressPort string) {
	makeup.PrintH1("Deploying Dex Idp...")

	dexClientSecret := generateRandom32BytesBase64()

	templateVars := &dexTemplateVars{
		Host:            hostIP,
		IngressPort:     ingressPort,
		DexClientSecret: dexClientSecret,
	}

	manifests, err := renderTemplate(dexManifestsTemplate, templateVars)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not render the dex manifests.")
	}

	makeup.PrintH2("Applying the following manifests: ")
	makeup.PrintYAML(manifests.Bytes(), false)
	makeup.WaitForUser(demo.UnattendedMode)

	k.mgmtK8s.KubectlApplyStdin(manifests)

	makeup.Print("Done applying the dex manifests.")
}

// Waits for the dex deployment to be ready.
func (k *KlutchManager) WaitForDex() {
	makeup.PrintH1("Waiting for dex to become ready...")

	k.mgmtK8s.KubectlWaitForRollout("deployment", "dex", "default")

	makeup.PrintCheckmark("Dex appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}
