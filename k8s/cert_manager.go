package k8s

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
)

func ApplyCertManagerManifests(waitForUser bool) {
	makeup.PrintH1("Installing the cert-manager")
	count := CountPodsInNamespace(CertManagerNamespace)

	if count > 0 {
		makeup.Print(fmt.Sprintf("Found %d pods in the %s namespace", count, CertManagerNamespace))
	}

	KubectlApplyF(CertManagerManifestUrl, waitForUser)

	WaitForCertManagerToBecomeReady()
}

func WaitForCertManagerToBecomeReady() {
	makeup.PrintH1("Waiting for the cert-manager API to become ready.")
	crashLoopBackoffCount := 10

	for i := 1; i <= crashLoopBackoffCount; i++ {
		cmd := exec.Command("cmctl", "check", "api")
		output, err := cmd.CombinedOutput()

		makeup.Print(cmd.String())

		//TODO Crash loop detection / timeout
		if err != nil {
			makeup.PrintWait("Continuing to wait for the cert-manager API...")
		}

		strOutput := string(output)

		fmt.Println(strOutput)

		if strings.TrimSpace(strOutput) == "The cert-manager API is ready" {
			makeup.PrintCheckmark("The cert-manager is ready")
			return
		} else {
			makeup.PrintWait("Continuing to wait for the cert-manager API...")
		}

		time.Sleep(30 * time.Second)
	}

	makeup.PrintFailSummary("The cert-manager did not become ready within reasonable time.")
}
