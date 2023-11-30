package demo

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Valid options: "kind"
var DemoClusterName string
var UnattendedMode bool // Ask yes-no questions or assume "yes"

// TODO Remove
func CheckIfKindClusterExists(kindDemoClusterName string) bool {
	cmd := exec.Command("kind", "get", "clusters")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		PrintFail("Couldn't capture output of 'kind get clusters' command.")
		log.Fatal(err)
		return false
	}

	strOutput := string(output)

	PrintListFromMultilineString("Kind Clusters", strOutput)

	// fmt.Println("\nKind clusters:")
	// fmt.Println(strOutput)

	if strings.Contains(strOutput, kindDemoClusterName) {
		PrintCheckmark("There is a suitable Kind cluster with the name " + kindDemoClusterName + " running.")
		return true
	}

	// Check if the output contains the string "No kind clusters found."
	if strings.Contains(strOutput, "No kind clusters found.") {
		PrintWarning(" There are no kind clusters. A cluster with the name: " + kindDemoClusterName + " is needed.")
		return false
	}

	PrintFail("There appears to be kind clusters but none with the name: " + kindDemoClusterName + ".")
	return false
}

// TODO Remove
func CreateKindCluster(kindDemoClusterName string) {
	PrintFlexedBiceps("Let's create a Kubernetes cluster named " + kindDemoClusterName + " using Kind...")

	// kind create cluster --name a8s-ds --config kind-cluster-3nodes.yaml
	cmd := exec.Command("kind", "create", "cluster", "--name", kindDemoClusterName)

	PrintCommandBox(cmd.String())
	WaitForUser()

	output, err := cmd.CombinedOutput()

	if err != nil {
		PrintFail("Failed to execute the command: " + err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return
	} else {
		fmt.Println(string(output))
		return
	}
}

/*
Represents a single entry of the simplified list as returned by
running the command 'minikube profile list -o json'.
*/
type MinikubeClusterStatusItem struct {
	Name   string
	Status string
}

func (si *MinikubeClusterStatusItem) String() string {
	return fmt.Sprintf("%s:\t\t%s\n", si.Name, si.Status)
}

type MinikubeClusterStatus struct {
	Valid []MinikubeClusterStatusItem
}

func (s *MinikubeClusterStatus) String() string {
	ret := ""

	for _, status := range s.Valid {
		ret += status.String()
	}
	return ret
}

func CheckIfMinkubeClusterExists(demoClusterName string) bool {
	ret := false
	// Output example: https://gist.github.com/fischerjulian/ae095c2848c5c9cd668a5c25bbd83a94s
	cmd := exec.Command("minikube", "profile", "list", "-o", "json")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		ExitDueToFatalError(err, "Couldn't capture output of 'minikube profile list -o json' command.")
	}

	// strOutput := string(output)
	// PrintListFromMultilineString("Minikube Clusters", strOutput)

	var clusterStatus MinikubeClusterStatus

	desired_a8sDemoClusterStatus := MinikubeClusterStatusItem{
		Name:   demoClusterName,
		Status: "Running",
	}

	json.Unmarshal(output, &clusterStatus)
	// fmt.Printf("Status: %+v", clusterStatus.Valid)

	PrintListFromMultilineString("Minikube Clusters:", clusterStatus.String())

	if slices.Contains(clusterStatus.Valid, desired_a8sDemoClusterStatus) {
		ret = true
		PrintCheckmark("There is a suitable Minikube cluster with the name " + demoClusterName + " running.")
	} else {
		ret = false
		PrintWarning(" There are no Minikube clusters. A cluster with the name: " + demoClusterName + " is needed.")
	}

	return ret
}

// TODO Remove code duplication with kind
func CreateMinkubeCluster(demoClusterName string) {
	PrintFlexedBiceps("Let's create a Kubernetes cluster named " + demoClusterName + " using minikube...")

	//TODO make configurable
	memory := "8gb"
	nr_of_nodes := "4"

	cmd := exec.Command("minikube", "start", "--nodes", nr_of_nodes, "--memory", memory, "--profile", demoClusterName)

	PrintCommandBox(cmd.String())
	WaitForUser()

	output, err := cmd.CombinedOutput()

	if err != nil {
		PrintFail("Failed to execute the command: " + err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return
	} else {
		fmt.Println(string(output))
		return
	}
}

/*
Deletes the given demo Kubernetes cluster.
*/
func DeleteKubernetesCluster() {
	PrintWarning("Deleting the Demo Kubernetes Cluster " + DemoClusterName + "...")

	cmd := exec.Command("minikube", "delete", "--profile", DemoClusterName)

	PrintCommandBox(cmd.String())
	WaitForUser()

	output, err := cmd.CombinedOutput()

	if err != nil {
		PrintFail("Failed to execute the command: " + err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return
	} else {
		fmt.Println(string(output))
		return
	}
}
