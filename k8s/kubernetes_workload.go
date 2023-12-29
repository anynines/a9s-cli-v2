package k8s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	m1u "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

const certManagerNamespace = "cert-manager"
const certManagerManifestUrl = "https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml"

/*
Represents the state of a Pod which is expected to be running at some point.
The attribute "Running" is meant to be updated by a control loop.
*/
type PodExpectationState struct {
	Name    string
	Running bool
}

/*
Wait for a set of Pods known by name to enter the status "Running".
*/
func WaitForSystemToBecomeReady(namespace, systemName string, expectedPods []PodExpectationState) {
	makeup.PrintH1("Waiting for the " + systemName + " to become ready...")

	allGood := true

	makeup.Print(fmt.Sprintf("Checking the existence of the following %d Pods: ", len(expectedPods)))

out:
	for {
		// We start optimistically that all pods are running
		allGood = true
		for _, expectedPodPrefix := range expectedPods {
			makeup.Print("Checking the " + expectedPodPrefix.Name + "...")
			if checkIfPodHasStatusRunningInNamespace(expectedPodPrefix.Name, namespace) {
				makeup.PrintCheckmark("The " + expectedPodPrefix.Name + " appears to be running.")
				expectedPodPrefix.Running = true
			} else {
				// Sadly, at least one pod isn't running so we need another loop iteration
				makeup.PrintFail("The " + expectedPodPrefix.Name + " is not ready (yet).")
				allGood = false
			}

		}
		if allGood {
			makeup.PrintSuccessSummary("The " + systemName + " appears to be ready. All expected pods are running.")
			break out
		} else {
			makeup.PrintWait("The " + systemName + " is not ready (yet), let's try again in 5s ...")
			time.Sleep(5 * time.Second)
		}
	}
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
				makeup.PrintCheckmark("The Pod " + pod.Name + " is running as expected.")
				return true
			case v1.PodFailed:
				makeup.PrintFail("The Pod " + pod.Name + "h has failed but should be running.")
				// makeup.PrintFail("The " + A8sSystemName + " has not been installed successfully.")
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

/*
Examples:
delete postgresql {name}
apply -f {path}
delete -f {path}
apply --kustomize {path}
*/
func KubectlAct(verb, flag, filepath string, waitForUser bool) {
	cmd := exec.Command("kubectl", verb, flag, filepath)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(waitForUser)

	output, err := cmd.CombinedOutput()

	fmt.Println(string(output))

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl "+verb+" with using the command: "+cmd.String())
	}
}

func KubectlApplyF(yamlFilepath string, waitForUser bool) {
	KubectlAct("apply", "-f", yamlFilepath, waitForUser)
}

func KubectlDeleteF(yamlFilepath string, waitForUser bool) {
	KubectlAct("delete", "-f", yamlFilepath, waitForUser)
}

func KubectlApplyKustomize(kustomizeFilepath string, waitForUser bool) {
	KubectlAct("apply", "--kustomize", kustomizeFilepath, waitForUser)
}

// https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/main.go
func CountPodsInNamespace(namespace string) int {

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

func ApplyCertManagerManifests(waitForUser bool) {
	makeup.PrintH1("Installing the cert-manager")
	count := CountPodsInNamespace(certManagerNamespace)

	if count > 0 {
		makeup.Print(fmt.Sprintf("Found %d pods in the %s namespace", count, certManagerNamespace))
	}

	KubectlApplyF(certManagerManifestUrl, waitForUser)

	WaitForCertManagerToBecomeReady()
}

/*
See: https://github.com/kubernetes/client-go > dynamic.
A dynamic client can perform generic operations on arbitrary Kubernetes API objects.
*/
func GetDynamicKubernetesClient() *dynamic.DynamicClient {
	// use the current context in kubeconfig
	config := GetKubernetesConfig()

	// create the clientset
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't create dynamic Kubernetes client.")
	}

	return dynamicClient
}

/*
Example:

gvr := schema.GroupVersionResource{Group: "backups.anynines.com", Version: "v1beta3", Resource: "backups"}

desiredConditionsMap := make(map[string]interface{})
desiredConditionsMap["reason"] = "Complete"
desiredConditionsMap["status"] = "True"
*/
func WaitForKubernetesResource(namespace string, gvr schema.GroupVersionResource, desiredConditionsMap map[string]interface{}) error {

	dynamicClient := GetDynamicKubernetesClient()

	// Watch for changes in Backup resources.
	watchInterface, err := dynamicClient.Resource(gvr).Namespace(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't create dynamic WatchInterface to watch Kubernetes resource %v.", gvr))
	}

	makeup.Print(fmt.Sprintf("Watching for resources: %v", gvr))

	for event := range watchInterface.ResultChan() {
		switch event.Type {
		case watch.Error:
			makeup.PrintFail(fmt.Sprintf("It was all for nothing! %v", event))
			os.Exit(1)
		case watch.Added, watch.Modified:
			backup, ok := event.Object.(*m1u.Unstructured)
			if !ok {
				makeup.ExitDueToFatalError(nil, "Could not cast to Unstructured")
			}

			// pretty.Print(backup)

			// Check the status.conditions for the desired status.
			status, exists, err := m1u.NestedFieldCopy(backup.Object, "status", "conditions")
			if err != nil && !exists {
				makeup.PrintWait("There is not status, yet.")
				continue
			}

			conditions, ok := status.([]interface{})
			if !ok {
				makeup.PrintWait(".")
				continue
			}

			makeup.PrintWait("Status is now available. Checking conditions...")

			for _, condition := range conditions {
				condMap, ok := condition.(map[string]interface{})
				if !ok {
					makeup.PrintWarning("Condition is not a map")
					continue
				}

				// fmt.Printf("%v\n", condMap)

				//TODO Check for conditions in desiredConditionsMap
				if ConditionsAreMet(condMap, desiredConditionsMap) {
					makeup.PrintCheckmark("Backup complete: " + backup.GetName())
					return nil
					//
				} else {
					makeup.PrintWait("Desired conditions are not met, yet...")
				}
			}

			continue
		}
	}
	return errors.New("expected conditions have not been met")
}

/*
Verifies whether the key-value pairs of expectedConditionsMap are contained in
actualConditionsMap.
*/
func ConditionsAreMet(actualConditionsMap, expectedConditionsMap map[string]interface{}) bool {
	for key, expectedValue := range expectedConditionsMap {
		actualValue, exists := actualConditionsMap[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	return true
}
