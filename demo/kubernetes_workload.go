package demo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
Wait for a set of Pods known by name to enter the status "Running".
*/
func WaitForSystemToBecomeReady(systemName string, expectedPods []PodExpectationState) {
	makeup.PrintH1("Waiting for the " + systemName + " to become ready...")

	allGood := true

	//TODO Make configurable or move to beginning of file for better maintainability
	systemNamespace := "a8s-system"

out:
	for {
		// We start optimistically that all pods are running
		allGood = true
		for _, expectedPodPrefix := range expectedPods {
			makeup.Print("Checking the " + expectedPodPrefix.Name + "...")
			if checkIfPodHasStatusRunningInNamespace(expectedPodPrefix.Name, systemNamespace) {
				makeup.PrintCheckmark("The " + expectedPodPrefix.Name + " appears to be running.")
				expectedPodPrefix.Running = true
			} else {
				// Sadly, at least one pod isn't running so we need another loop iteration
				makeup.PrintFail("The " + expectedPodPrefix.Name + " is not ready (yet).")
				allGood = false
			}

			if allGood {
				makeup.PrintSuccessSummary("The " + systemName + " appears to be ready. All expected pods are running.")
				break out
			} else {
				makeup.PrintWait("The " + systemNamespace + " is not ready (yet), let's try again in 5s ...")
				time.Sleep(5 * time.Second)
			}
		}
	}
	makeup.WaitForUser(UnattendedMode)
}

func checkIfPodHasStatusRunningInNamespace(podNameStartsWith string, namespace string) bool {
	clientset := GetKubernetesClientSet()

	//for {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, podNameStartsWith) {
			makeup.Print("Found pod with prefix " + podNameStartsWith)

			// if debug {
			// 	//pod.Status.Phase
			// 	makeup.Print("Pod has status: " + pod.Status.String())
			// }

			switch phase := pod.Status.Phase; phase {
			case v1.PodRunning:
				makeup.PrintCheckmark("The Pod " + pod.Name + "h is running as expected.")
				return true
			case v1.PodFailed:
				makeup.PrintFail("The Pod " + pod.Name + "h has failed but should be running.")
				makeup.PrintFail("The " + A8sSystemName + " has not been installed successfully.")
				os.Exit(1)

			case v1.PodPending:
				makeup.Print("The Pod " + pod.Name + "h in pending but should be running.")
				return false
			case v1.PodSucceeded:
				makeup.Print("The Pod " + pod.Name + "h has succeeded but should be running.")
				return false
			case v1.PodUnknown:
				makeup.Print("The Pod " + pod.Name + "h has an unknown status but should be running.")
				return false
			default:
				return false
			}
		}
	}
	return false
}

func CountPodsInDemoNamespace() int {
	return countPodsInNamespace(DemoConfig.DemoSpace)
}
func KubectlApplyF(yamlFilepath string) {

	cmd := exec.Command("kubectl", "apply", "-f", yamlFilepath)

	output, err := cmd.CombinedOutput()

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(UnattendedMode)

	if err != nil {
		makeup.PrintWarning(string(output))
		makeup.ExitDueToFatalError(err, "Can't kubectl apply with command: "+cmd.String())
	}

	fmt.Println(string(output))
}

func KubectlApplyKustomize(kustomizeFilepath string) {

	cmd := exec.Command("kubectl", "apply", "--kustomize", kustomizeFilepath)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(UnattendedMode)

	output, err := cmd.CombinedOutput()

	fmt.Println(string(output))

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl kustomize with using the command: "+cmd.String())
	}

}

// https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/main.go
func countPodsInNamespace(namespace string) int {

	makeup.PrintH2("Checking whether there are pods in the cluster...")

	clientset := GetKubernetesClientSet()

	//for {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	return len(pods.Items)
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

func ApplyCertManagerManifests() {
	makeup.PrintH1("Installing the cert-manager")
	count := countPodsInNamespace(certManagerNamespace)

	if count > 0 {
		makeup.Print(fmt.Sprintf("Found %d pods in the %s namespace", count, certManagerNamespace))
	}

	KubectlApplyF(certManagerManifestUrl)

	WaitForCertManagerToBecomeReady()
}
