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
	"github.com/kr/pretty"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	m1u "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

const CertManagerNamespace = "cert-manager"
const CertManagerManifestUrl = "https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml"

// TODO Make configurable
const kubectlWaitTimeoutOption = "--timeout=120s"

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
	makeup.PrintH1("Waiting for " + systemName + " to become ready...")

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
				makeup.PrintWait("The " + expectedPodPrefix.Name + " is not ready (yet).")
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

/*
Uses kubectl wait to wait for each expected pod to become ready.
Pods are identified by label and namespace.
*/
func KubectlWaitForSystemToBecomeReady(namespace string, expectedPodsByLabels []string) {
	for _, podLabel := range expectedPodsByLabels {
		KubectlWaitForPod(namespace, podLabel)
	}
}

func KubectlWaitForPod(namespace, podLabel string) {

	// kubectl wait --for=condition=Ready pod -l "app.kubernetes.io/name=backup-manager" -n a8s-system
	// Outcome 1: error: timed out waiting for the condition on pods/a8s-backup-controller-manager-788fcd578d-kzb4f
	// Outcome 2: pod/postgresql-controller-manager-7f8c7758d-28lc2 condition met
	cmd := exec.Command("kubectl", "wait", "--for=condition=Ready", "pod", "-l", podLabel, "-n", namespace, kubectlWaitTimeoutOption)

	output, err := cmd.CombinedOutput()

	strOutput := string(output)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Pod with label %s in namespace %s has not become ready on time", podLabel, namespace))
	}

	if !strings.Contains(strOutput, "condition met") {
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Pod with label %s in namespace %s has not become ready but conditions haven't been met. Got: %s", podLabel, namespace, strOutput))
	}
}

/*
TODO This method did not work when the backup-manager went into a CrashLoopBackOff. There is likely a bug here.
*/
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
				makeup.Print("The Pod " + pod.Name + " in pending but should be running.")
				return false
			case v1.PodSucceeded:
				makeup.Print("The Pod " + pod.Name + " has succeeded but should be running.")
				return false
			case v1.PodUnknown:
				makeup.Print("The Pod " + pod.Name + " has an unknown status but should be running.")
				return false
			default:
				return false
			}
		}
	}
	return false
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
Wait for a Kubernetes resource to reach either a desired or failed state.

Namespace
Name: name of the object to wait for, e.g. name of the backup
The desiredConditionsMap contains the conditions to indicate success while
the failedConditionsMap contains the conditions to indicate failure.

Example:

gvr := schema.GroupVersionResource{Group: "backups.anynines.com", Version: "v1beta3", Resource: "backups"}

desiredConditionsMap := make(map[string]interface{})
desiredConditionsMap["reason"] = "Complete"
desiredConditionsMap["status"] = "True"

failedConditionsMap := make(map[string]interface{})
failedConditionsMap["reason"] = "PermanentlyFailed"
failedConditionsMap["status"] = "True"

TODO
  - Refactor using WaitForKubernetesResourceWithFunction
  - Rename WaitForKubernetesResourceWithFunction to WaitForKubernetesResource
*/
func WaitForKubernetesResource(namespace, name string, gvr schema.GroupVersionResource, desiredConditionsMap map[string]interface{}, failedConditionsMap map[string]interface{}) error {

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	}

	dynamicClient := GetDynamicKubernetesClient()

	// Watch for changes in Backup resources.
	watchInterface, err := dynamicClient.Resource(gvr).Namespace(namespace).Watch(context.TODO(), listOptions)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't create dynamic WatchInterface to watch Kubernetes resource %v.", gvr))
	}

	makeup.PrintVerbose(fmt.Sprintf("Watching for resources: %v", gvr))

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

			if makeup.Verbose {
				fmt.Print("Event object:")
				pretty.Print(backup)
			}

			// Check the status.conditions for the desired status.
			status, exists, err := m1u.NestedFieldCopy(backup.Object, "status", "conditions")
			if err != nil && !exists {
				if makeup.Verbose {
					makeup.PrintWait("There is not status, yet.")
				}
				continue
			}

			/*
				Conditions is a list of condition maps.
				One of the condition maps in conditions has the "Status" => "True".
				This is the current condition.

				Conditions change over time so that this section is part of a loop and
				whill be executed when conditions change.

				We are waiting for the circumstance when there's a condition map with
				"Reason" => "Complete" and "Status" => "True".

				TODO There are also other cases which represent a final state, for
					example when a backup has permanently failed. They should also be captured
					to indicate the user that the backup/restore has failed instead of
					keeping the loop running while blocking the cli.
			*/
			conditions, ok := status.([]interface{})
			if !ok {
				if makeup.Verbose {
					makeup.PrintWait(".")
				}
				continue
			}

			if makeup.Verbose {
				makeup.PrintWait("Status is now available. Checking conditions...")
			}

			conditionsAreMet := false
			failedConditionsAreMet := false

			for _, condition := range conditions {
				makeup.PrintVerbose(fmt.Sprintf("Investigating condition %v of conditions\n", condition))
				condMap, ok := condition.(map[string]interface{})
				if !ok {
					makeup.PrintWarning("Condition is not a map")
					continue
				}

				makeup.PrintVerbose(fmt.Sprintf("%v\n", condMap))

				// There are several condition fields only the condition with Status => true matters
				// hence: if one of the condition maps has Status => true and has the desired "reason",
				// 	we are ready to proceed.
				if ConditionsAreMet(condMap, desiredConditionsMap) {
					conditionsAreMet = true
					break
				}

				if failedConditionsMap != nil && ConditionsAreMet(condMap, failedConditionsMap) {
					failedConditionsAreMet = true
					break
				}
			}

			//TODO The conditionsAreMet variable is not necessary but increases readability. Does it?
			// No it doesn't. Code here could also be put in the above if clause (if ConditionsAreMet ...)
			if conditionsAreMet {
				//makeup.PrintCheckmark("Operation complete for resource: " + backup.GetName())
				return nil
				//
			} else {
				if makeup.Verbose {
					makeup.PrintWait("Desired conditions are not met, yet...")
				}
			}

			if failedConditionsAreMet {
				errorMessage := fmt.Sprintf("waiting for Kubernetes resource %v in namespace %s has failed. Resource reached failed state", gvr, namespace)
				if makeup.Verbose {
					makeup.PrintWarning(errorMessage)
				}
				return errors.New(errorMessage)
			}

			continue
		}
	}
	return errors.New("expected conditions have not been met")
}

/*
name refers to the metadata.name value of the object of interest.

waitLonger is a function describing what to wait for covering both success and failure scenarios.
It returns true if waiting shall go on and false if the awaited event has happened.g
*/
func WaitForKubernetesResourceWithFunction(namespace, name string, gvr schema.GroupVersionResource, waitLonger func(object *m1u.Unstructured) bool) error {

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	}

	dynamicClient := GetDynamicKubernetesClient()

	// Watch for changes in Backup resources.
	watchInterface, err := dynamicClient.Resource(gvr).Namespace(namespace).Watch(context.TODO(), listOptions)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't create dynamic WatchInterface to watch Kubernetes resource %v.", gvr))
	}

	makeup.PrintVerbose(fmt.Sprintf("Watching for resources: %v", gvr))

	var goOn bool

	for event := range watchInterface.ResultChan() {
		switch event.Type {
		case watch.Error:
			makeup.ExitDueToFatalError(err, "A watch.Error occurred watching the resource.")
		case watch.Added, watch.Modified:
			object, ok := event.Object.(*m1u.Unstructured)
			if !ok {
				makeup.ExitDueToFatalError(nil, "Could not cast to Unstructured")
			}

			if makeup.Verbose {
				fmt.Print("Event object:")
				pretty.Print(object)
			}

			goOn = waitLonger(object)

			if !goOn {
				return nil
			}
		}
	}
	return errors.New("expected conditions have not been met")
}

/*
Verifies whether the key-value pairs of expectedConditionsMap are contained in
actualConditionsMap.

The actualConditionsMap is a single record of a conditions array similar to:

	map[lastTransitionTime:2024-01-03T07:04:28Z message:Restore object has been created reason:Initialized status:False type:PermanentlyFailed]

The ConditionsAreMet function has to be applied against all condition entries, each being a condition map.
*/
func ConditionsAreMet(actualConditionsMap, expectedConditionsMap map[string]interface{}) bool {
	for key, expectedValue := range expectedConditionsMap {
		makeup.PrintVerbose(fmt.Sprintf("\nConditionsAreMet? Checking whether %v / %v is in %v\n", key, expectedValue, actualConditionsMap))
		actualValue, exists := actualConditionsMap[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	return true
}

/*
Wait for the given service account in the given namespace to become ready.
Blocks during wait.
*/
func WaitForServiceAccount(unattendedMode bool, namespace, serviceAccountName string) {

	for nrAttempts := 0; nrAttempts <= 600; nrAttempts++ {

		// Wait x s for the the serviceAccountToShowUp
		_, output, err := Kubectl(unattendedMode, "get", "serviceaccount", "-n", namespace, serviceAccountName)

		if err == nil {

			// Found the service account
			return
		}

		if strings.Contains(string(output), "serviceaccounts \""+serviceAccountName+"\" not found") {

			// Did not find the service account
			makeup.Print(fmt.Sprintf("The service account %s does not exist (yet) in namespace %s.", serviceAccountName, namespace))
		} else {

			// Some other error occured
			makeup.ExitDueToFatalError(err, "Can't get service account "+serviceAccountName+" in namespace "+namespace)
		}

		time.Sleep(2 * time.Second)
	}
	makeup.ExitDueToFatalError(nil, fmt.Sprintf("Timeout. Can't get service account "+serviceAccountName+" in namespace "+namespace))
}

func CreateNamespace(unattendedMode bool, namespace string) {
	_, output, err := Kubectl(unattendedMode, "create", "namespace", namespace)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Couldn't create namespace %s. Output was: %s", namespace, string(output)))
	}
}
