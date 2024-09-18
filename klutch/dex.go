package klutch

import (
	_ "embed"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
)

//go:embed templates/dex.tmpl
var dexManifestsTemplate string

type dexTemplateVars struct {
	// Used to determine the node name of the management cluster
	MgmtClusterName string
	// Host name by which dex is accessed from outside the docker network. Should be "127.0.0.1".
	ExternalHost string
	// External ingress port accessed from outside the docker network. Should be the flag passed by the user.
	ExternalIngressPort string
	// Internal ingress port, corresponds to the one defined in the applied nginx-ingress manifests. Should be "80".
	InternalIngressPort string
	DexClientSecret     string
}

func (k *KlutchManager) DeployDex(ingressPort string) {
	makeup.PrintH1("Deploying Dex Idp...")

	dexClientSecret := generateRandom32BytesBase64()

	templateVars := &dexTemplateVars{
		// Host:                hostIP,
		MgmtClusterName:     mgmtClusterName,
		ExternalHost:        "127.0.0.1",
		InternalIngressPort: "80",
		ExternalIngressPort: ingressPort,
		DexClientSecret:     dexClientSecret,
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
