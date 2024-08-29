package klutch

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os/exec"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	configPackageName  = "w5n9a2g2-anynines-dataservices"
	configPackageImage = "public.ecr.aws/w5n9a2g2/anynines/dataservices:v1.3.0"
	helmChartUrl       = "https://charts.crossplane.io/stable/crossplane-1.16.0.tgz"
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
		"--kube-context", contextMgmt,
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

	k.mgmtK8s.KubectlWaitForSystemToBecomeReady("crossplane-system", []string{
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

	k.mgmtK8s.KubectlApplyStdin(in)

	makeup.Print("Kubernetes Crossplane provider applied.")
}

func (k *KlutchManager) WaitForProviderKubernetes() {
	makeup.PrintH1("Waiting for the Kubernetes Crossplane provider to become ready...")

	k.mgmtK8s.KubectlWaitForResourceCondition("healthy", "providers", "provider-kubernetes", "crossplane-system")
	k.mgmtK8s.KubectlWaitForResourceConditionWithSelector("healthy", "providerrevision", "pkg.crossplane.io/package=provider-kubernetes", "crossplane-system")
	k.mgmtK8s.KubectlWaitForRolloutWithSelector("deployment", "klutch-provider=provider-kubernetes", "crossplane-system")
	k.mgmtK8s.KubectlWaitForResourceConditionWithSelector("ready", "pod", "pkg.crossplane.io/provider=provider-kubernetes", "crossplane-system")
	k.mgmtK8s.KubectlWaitForResourceCondition("established", "crd", "configurations.pkg.crossplane.io", "crossplane-system")

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

	k.mgmtK8s.KubectlApplyStdin(in)

	makeup.PrintCheckmark("Kubernetes Crossplane provider config applied.")
}

// Deploys the a8s APIs (CRDs and Compositions).
// Currently uses the `crossplane xpkg` method because the manifests are closed source. Once they are open source,
// the manifests can be referred directly via URL. This would avoid the dependency on the crossplane CLI, which is poorly featured.
func (k *KlutchManager) DeployKlutchCrossplaneConfigPkg() {
	makeup.PrintH1("Deploying the Klutch Crossplane configuration package...")

	// In order to be idempotent, we want the configuration package to be re-applied if it already exist.
	// `crossplane xpkg install configuration` fails if it already exists, and `crossplane xpkg update` fails if it doesn't exist.
	// So we need to first determine if the xpkg is already installed or not.
	client := k.mgmtK8s.GetDynamicKubernetesClient()

	gvr := schema.GroupVersionResource{Group: "pkg.crossplane.io", Version: "v1", Resource: "configurations"}
	_, err := client.Resource(gvr).Get(context.TODO(), configPackageName, metav1.GetOptions{})

	verb := "install"
	if err == nil {
		makeup.PrintWarning("Klutch Crossplane configuration package already exists. Updating it...")
		verb = "update"
	} else {
		if !errors.IsNotFound(err) {
			makeup.ExitDueToFatalError(err, "An unexpected error occured while checking if the Klutch Crossplane configuration package exist")
		}
	}

	switchContext(contextMgmt) // Crossplane CLI does not allow specifying the context via flag.
	cmd := exec.Command("crossplane", "xpkg", verb, "configuration", configPackageImage)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(demo.UnattendedMode)

	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not execute crossplane CLI: %v", string(output)))
	}

	makeup.Print("Klutch Crossplane configuration package applied.")
}

func (k *KlutchManager) WaitForKlutchCrossplaneConfigPkg() {
	makeup.PrintH1("Waiting for the Klutch Crossplane configuration package to become ready...")

	k.mgmtK8s.KubectlWaitForResourceCondition("healthy", "configuration", configPackageName, "crossplane-system")

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

	k.mgmtK8s.KubectlApplyStdin(in)

	makeup.PrintCheckmark("Klutch APIServiceExportTemplates applied.")
}
