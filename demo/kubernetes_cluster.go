package demo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func CheckIfKindClusterExists() bool {
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

func CreateKindCluster() {
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
