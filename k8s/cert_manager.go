package k8s

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
)

func (k *KubeClient) ApplyCertManagerManifests(waitForUser bool) {
	makeup.PrintH1("Installing the cert-manager")
	count := k.CountPodsInNamespace(CertManagerNamespace)

	if count > 0 {
		makeup.Print(fmt.Sprintf("Found %d pods in the %s namespace", count, CertManagerNamespace))
	}

	k.KubectlApplyF(CertManagerManifestUrl, waitForUser)

	k.WaitForCertManagerToBecomeReady()
}

func (k *KubeClient) WaitForCertManagerToBecomeReady() {
	makeup.PrintH1("Waiting for the cert-manager API to become ready.")

	allReady := true
	deployments := []string{"cert-manager", "cert-manager-webhook", "cert-manager-cainjector"}

	for _, deployment := range deployments {
		cmd := exec.Command("kubectl", "rollout", "status", fmt.Sprintf("deployment/%s", deployment), "-n", "cert-manager", "--timeout=180s")

		if k.KubeContext != "" {
			cmd.Args = append(cmd.Args, "--context", k.KubeContext)
		}

		output, err := cmd.CombinedOutput()

		makeup.Print(cmd.String())

		strOutput := string(output)

		fmt.Println(strOutput)

		if err != nil {
			allReady = false
			makeup.PrintWait(fmt.Sprintf("Continuing to wait for the %s deployment...", deployment))
		}

		if allReady {
			makeup.PrintCheckmark("All cert-manager components are ready")
			return
		} else {
			makeup.PrintWait("Continuing to wait for the cert-manager API...")
		}

		time.Sleep(30 * time.Second)
	}

	makeup.PrintFailSummary("The cert-manager did not become ready within reasonable time.")
}
