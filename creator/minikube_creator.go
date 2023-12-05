package creator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"

	"github.com/anynines/a9s-cli-v2/makeup"
)

type MinikubeCreator struct {
}

/*
Represents a single entry of the simplified list as returned by
running the command 'minikube profile list -o json'.
*/
type minikubeClusterStatusItem struct {
	Name   string
	Status string
}

func (si *minikubeClusterStatusItem) String() string {
	return fmt.Sprintf("%s:\t\t%s\n", si.Name, si.Status)
}

type minikubeClusterStatus struct {
	Valid []minikubeClusterStatusItem
}

func (s *minikubeClusterStatus) String() string {
	ret := ""

	for _, status := range s.Valid {
		ret += status.String()
	}
	return ret
}

func (c MinikubeCreator) Exists(name string) bool {
	ret := false

	// Output example: https://gist.github.com/fischerjulian/ae095c2848c5c9cd668a5c25bbd83a94s
	cmd := exec.Command("minikube", "profile", "list", "-o", "json")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't capture output of 'minikube profile list -o json' command.")
	}

	var clusterStatus minikubeClusterStatus

	json.Unmarshal(output, &clusterStatus)
	// fmt.Printf("Status: %+v", clusterStatus.Valid)

	makeup.PrintListFromMultilineString("Minikube Clusters:", clusterStatus.String())

	// Compare by name
	contains := slices.ContainsFunc(clusterStatus.Valid, func(n minikubeClusterStatusItem) bool {
		return n.Name == name
	})

	if contains {
		ret = true
		makeup.PrintCheckmark("There is a suitable Minikube cluster with the name " + name + " running.")
	} else {
		ret = false
		makeup.PrintWarning(" There are no Minikube clusters. A cluster with the name: " + name + " is needed.")
	}

	return ret
}

func (c MinikubeCreator) Running(name string) bool {
	ret := false
	// Output example: https://gist.github.com/fischerjulian/ae095c2848c5c9cd668a5c25bbd83a94s
	cmd := exec.Command("minikube", "profile", "list", "-o", "json")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't capture output of 'minikube profile list -o json' command.")
	}

	// strOutput := string(output)
	// PrintListFromMultilineString("Minikube Clusters", strOutput)

	var clusterStatus minikubeClusterStatus

	desired_a8sDemoClusterStatus := minikubeClusterStatusItem{
		Name:   name,
		Status: "Running",
	}

	json.Unmarshal(output, &clusterStatus)
	// fmt.Printf("Status: %+v", clusterStatus.Valid)

	makeup.PrintListFromMultilineString("Minikube Clusters:", clusterStatus.String())

	if slices.Contains(clusterStatus.Valid, desired_a8sDemoClusterStatus) {
		ret = true
		makeup.PrintCheckmark("There is a suitable Minikube cluster with the name " + name + " running.")
	} else {
		ret = false
		makeup.PrintWarning(" There is no minikube cluster with the name: " + name + ".")
	}

	return ret
}

func (c MinikubeCreator) Create(spec KubernetesClusterSpec, unattended bool) {

	makeup.PrintFlexedBiceps("Let's create a Kubernetes cluster named " + spec.Name + " using minikube...")

	cmd := exec.Command("minikube", "start", "--nodes", strconv.Itoa(spec.NrOfNodes), "--memory", spec.NodeMemory, "--profile", spec.Name)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(unattended)

	output, err := cmd.CombinedOutput()

	if err != nil {
		makeup.PrintFail("Failed to execute the command: " + err.Error())
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
func (c MinikubeCreator) Delete(name string, unattended bool) {
	makeup.PrintWarning("Deleting the Demo Kubernetes Cluster " + name + "...")

	cmd := exec.Command("minikube", "delete", "--profile", name)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(unattended)

	output, err := cmd.CombinedOutput()

	if err != nil {
		makeup.PrintFail("Failed to execute the command: " + err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return
	} else {
		fmt.Println(string(output))
		return
	}
}
