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
var KubernetesTool string

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

func CheckIfMinkubeClusterExists(kindDemoClusterName string) bool {

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
		Name:   kindDemoClusterName,
		Status: "Running",
	}

	json.Unmarshal(output, &clusterStatus)
	// fmt.Printf("Status: %+v", clusterStatus.Valid)

	PrintListFromMultilineString("Minikube Clusters:", clusterStatus.String())

	slices.Contains(clusterStatus.Valid, desired_a8sDemoClusterStatus)

	os.Exit(1)
	return false
}

// TODO Remove code duplication with kind
func CreateMinkubeCluster(kindDemoClusterName string) {
	PrintFlexedBiceps("Let's create a Kubernetes cluster named " + kindDemoClusterName + " using minikube...")

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
