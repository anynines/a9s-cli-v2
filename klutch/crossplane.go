package klutch

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
)

const (
	configPackageName        = "anynines-dataservices"
	configPackageManifestUrl = "https://raw.githubusercontent.com/anynines/klutchio/main/crossplane-api/deploy/config-pkg-anynines.yaml"
	helmChartUrl             = "https://charts.crossplane.io/stable/crossplane-1.16.0.tgz"
)

//go:embed manifests/provider-kubernetes.yaml
var providerKubernetesManifests string

//go:embed manifests/export-templates.yaml
var exportTemplatesManifests string

//go:embed manifests/config-in-cluster.yaml
var configInClusterManifests string

func (k *KlutchManager) DeployCrossplaneComponents() {
	k.DeployCrossplaneHelmChart()
	k.WaitForCrossplaneHelmChart()

	k.DeployProviderKubernetes()
	k.WaitForProviderKubernetes()

	k.DeployProviderKubernetesConfig()
	// Note: there doesn't seem to be anything to wait on for the ProviderConfig.

	k.DeployKlutchCrossplaneConfigPkg()
	k.WaitForKlutchCrossplaneConfigPkg()

	k.DeployKlutchExportTemplates()
}

func (k *KlutchManager) DeployCrossplaneHelmChart() {
	makeup.PrintH1("Deploying the Crossplane Helm chart...")

	cmd := exec.Command("helm",
		"upgrade", "-i",
		"crossplane",
		"--kube-context", contextControlPlane,
		"--namespace", "crossplane-system", "--create-namespace",
		helmChartUrl,
		"--set", `args={"--enable-ssa-claims"}`,
	)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(demo.UnattendedMode)

	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("error while deploying helm chart: %s", string(output)))
	}

	makeup.Print("Crossplane Helm chart applied.")
}

func (k *KlutchManager) WaitForCrossplaneHelmChart() {
	makeup.PrintH1("Waiting for the Crossplane components to become ready...")

	k.cpK8s.KubectlWaitForSystemToBecomeReady("crossplane-system", []string{
		"app=crossplane",
		"app=crossplane-rbac-manager",
	})

	makeup.PrintCheckmark("The Crossplane components appear to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}

func (k *KlutchManager) DeployProviderKubernetes() {
	makeup.PrintH1("Deploying the Kubernetes Crossplane provider...")

	in := bytes.NewBufferString(providerKubernetesManifests)

	makeup.PrintH2("Applying the following manifests: ")
	makeup.PrintYAML(in.Bytes(), false)

	k.cpK8s.KubectlApplyStdin(in)

	makeup.Print("Kubernetes Crossplane provider applied.")
}

func (k *KlutchManager) WaitForProviderKubernetes() {
	makeup.PrintH1("Waiting for the Kubernetes Crossplane provider to become ready...")

	k.cpK8s.KubectlWaitForResourceCondition("healthy", "providers", "provider-kubernetes", "crossplane-system")
	k.cpK8s.KubectlWaitForResourceConditionWithSelector("healthy", "providerrevision", "pkg.crossplane.io/package=provider-kubernetes", "crossplane-system")
	k.cpK8s.KubectlWaitForRolloutWithSelector("deployment", "klutch-provider=provider-kubernetes", "crossplane-system")
	k.cpK8s.KubectlWaitForResourceConditionWithSelector("ready", "pod", "pkg.crossplane.io/provider=provider-kubernetes", "crossplane-system")
	k.cpK8s.KubectlWaitForResourceCondition("established", "crd", "configurations.pkg.crossplane.io", "crossplane-system")

	makeup.PrintCheckmark("The Kubernetes Crossplane provider appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}

// Deploys the Kubernetes Provider Config
func (k *KlutchManager) DeployProviderKubernetesConfig() {
	makeup.PrintH1("Deploying the Kubernetes Crossplane provider config...")

	in := bytes.NewBufferString(configInClusterManifests)
	makeup.PrintH2("Applying the following manifests: ")
	makeup.PrintYAML(in.Bytes(), false)
	makeup.WaitForUser(demo.UnattendedMode)

	k.cpK8s.KubectlApplyStdin(in)

	makeup.PrintCheckmark("Kubernetes Crossplane provider config applied.")
}

// Deploys the a8s APIs (CRDs and Compositions).
func (k *KlutchManager) DeployKlutchCrossplaneConfigPkg() {
	makeup.PrintH1("Deploying the Klutch Crossplane configuration package...")

	k.cpK8s.KubectlApplyF(configPackageManifestUrl, true)

	makeup.PrintCheckmark("Klutch Crossplane configuration package applied.")
}

func (k *KlutchManager) WaitForKlutchCrossplaneConfigPkg() {
	makeup.PrintH1("Waiting for the Klutch Crossplane configuration package to become ready...")

	k.cpK8s.KubectlWaitForResourceCondition("healthy", "configuration", configPackageName, "crossplane-system")

	makeup.PrintCheckmark("The Klutch Crossplane configuration package appears to be ready.")
	makeup.WaitForUser(demo.UnattendedMode)
}

// Applies APIServiceExportTemplates for the a8s crossplane APIs.
func (k *KlutchManager) DeployKlutchExportTemplates() {
	makeup.PrintH1("Deploying the Klutch APIServiceExportTemplates...")

	in := bytes.NewBufferString(exportTemplatesManifests)

	makeup.PrintH2("Applying the following manifests: ")
	makeup.PrintYAML(in.Bytes(), false)
	makeup.WaitForUser(demo.UnattendedMode)

	k.cpK8s.KubectlApplyStdin(in)

	makeup.PrintCheckmark("Klutch APIServiceExportTemplates applied.")
}
