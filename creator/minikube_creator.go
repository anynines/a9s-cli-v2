package creator

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

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

	// Capture the command output
	// Output example: https://gist.github.com/fischerjulian/ae095c2848c5c9cd668a5c25bbd83a94s
	output, err := makeup.NewCommand("minikube", "profile", "list", "-o", "json").NoPrompt().Run()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't capture output of 'minikube profile list -o json' command:\n"+string(output))
	}

	var clusterStatus minikubeClusterStatus

	err = json.Unmarshal(output, &clusterStatus)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't unmarshal the output of 'minikube profile list -o json' command.")
	}

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

	// Capture the command output
	// Output example: https://gist.github.com/fischerjulian/ae095c2848c5c9cd668a5c25bbd83a94s
	output, err := makeup.NewCommand("minikube", "profile", "list", "-o", "json").NoPrompt().Run()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't capture output of 'minikube profile list -o json' command:\n"+string(output))
	}

	// strOutput := string(output)
	// PrintListFromMultilineString("Minikube Clusters", strOutput)

	var clusterStatus minikubeClusterStatus

	if err := json.Unmarshal(output, &clusterStatus); err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't parse output of 'minikube profile list -o json'.")
	}
	// fmt.Printf("Status: %+v", clusterStatus.Valid)

	makeup.PrintListFromMultilineString("Minikube Clusters:", clusterStatus.String())

	if slices.ContainsFunc(clusterStatus.Valid, func(item minikubeClusterStatusItem) bool {
		return item.Name == name && isRunningStatus(item.Status)
	}) {
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

	output, err := makeup.NewCommand("minikube", "start", "--nodes", strconv.Itoa(spec.NrOfNodes), "--memory", spec.NodeMemory, "--profile", spec.Name).WithPrompt().Run()

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to execute the command 'minikube start --nodes %s --memory %s --profile %s':\n%s",
			strconv.Itoa(spec.NrOfNodes), spec.NodeMemory, spec.Name, output))
	}
}

/*
Deletes the given demo Kubernetes cluster.
*/
func (c MinikubeCreator) Delete(name string, unattended bool) {
	makeup.PrintWarning("Deleting the Demo Kubernetes Cluster " + name + "...")

	output, err := makeup.NewCommand("minikube", "delete", "--profile", name).WithPrompt().Run()

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to execute the command 'minikube delete --profile %s':\n%s", name, output))
	}
}

func (c MinikubeCreator) GetContext(name string) string {
	return name
}

func isRunningStatus(status string) bool {
	return strings.EqualFold(status, "running") || strings.EqualFold(status, "ok")
}
