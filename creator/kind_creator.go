package creator

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anynines/a9s-cli-v2/makeup"
	"gopkg.in/yaml.v2"
)

type KindCreator struct {
	LocalWorkDir string
}

type KindClusterConfig struct {
	Kind       string              `yaml:"kind"`
	ApiVersion string              `yaml:"apiVersion"`
	Nodes      []map[string]string `yaml:"nodes"`
}

func (c KindCreator) buildKindClusterConfig(nrOfNodes int) KindClusterConfig {
	controlPlaneNode := make(map[string]string)
	controlPlaneNode["role"] = "control-plane"

	workerNode := make(map[string]string)
	workerNode["role"] = "worker"

	nodes := make([]map[string]string, 1)
	nodes[0] = controlPlaneNode

	for i := 1; i < nrOfNodes; i++ {
		nodes = append(nodes, workerNode)
	}

	config := KindClusterConfig{
		Kind:       "Cluster",
		ApiVersion: "kind.x-k8s.io/v1alpha4",
		Nodes:      nodes,
	}

	makeup.Print(fmt.Sprintf("Kind number of nodes: %d", nrOfNodes))
	makeup.Print(fmt.Sprintf("Kind Cluster Config:\n\t%v", config))

	return config
}

func (c KindCreator) getClusterConfigFilepath(spec KubernetesClusterSpec) string {
	designatedFilename := "kind-" + spec.Name + "config.yaml"

	designatedDir := filepath.Join(c.LocalWorkDir, "kind")

	err := os.MkdirAll(designatedDir, os.ModePerm)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couln't create workdir for kind at: "+designatedDir)
	}

	designatedFilepath := filepath.Join(designatedDir, designatedFilename)
	return designatedFilepath
}

func (c KindCreator) generateClusterConfigFile(spec KubernetesClusterSpec, config KindClusterConfig, designatedFilename string) {
	var designatedFilepath string

	yaml, err := yaml.Marshal(&config)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to generate Kind cluster config file. Error while marshaling.")
	}

	designatedFilepath = c.getClusterConfigFilepath(spec)

	f, err := os.Create(designatedFilepath)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Cannot generate Kind cluster config file. Can't create config file at: "+designatedFilepath)
	}

	defer f.Close()

	_, err = f.WriteString(string(yaml))

	if err != nil {
		makeup.ExitDueToFatalError(err, "Cannot generate Kind cluster config file. Can't write to config file.")
	}
}

func (c KindCreator) Create(spec KubernetesClusterSpec, unattendedMode bool) {
	makeup.PrintFlexedBiceps("Let's create a Kubernetes cluster named " + spec.Name + " using Kind...")

	//TODO Create YAML file
	config := c.buildKindClusterConfig(spec.NrOfNodes)
	filepath := c.getClusterConfigFilepath(spec)
	c.generateClusterConfigFile(spec, config, filepath)

	// kind create cluster --name a8s-ds --config kind-cluster-3nodes.yaml
	cmd := exec.Command("kind", "create", "cluster", "--name", spec.Name, "--config", filepath)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(unattendedMode)

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

func (c KindCreator) Exists(clustername string) bool {
	cmd := exec.Command("kind", "get", "clusters")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.PrintFail("Couldn't capture output of 'kind get clusters' command.")
		log.Fatal(err)
		return false
	}

	strOutput := string(output)

	makeup.PrintListFromMultilineString("Kind Clusters", strOutput)

	split := strings.Split(strOutput, "\n")
	for _, entry := range split {
		if entry == clustername {
			makeup.PrintCheckmark("There is a suitable Kind cluster with the name " + clustername + " running.")
			return true
		}
	}

	// Check if the output contains the string "No kind clusters found."
	if strings.Contains(strOutput, "No kind clusters found.") {
		makeup.PrintWarning(" There is no kind cluster with the name: " + clustername + ".")
		return false
	}

	makeup.PrintInfo("There appears to be kind clusters but none with the name: " + clustername + ".")
	return false
}

func (c KindCreator) Running(clustername string) bool {
	// Can't run if doesn't exist
	if !c.Exists(clustername) {
		return false
	}

	cmd := exec.Command("kubectl", "cluster-info", "--context", "kind-"+clustername)

	output, err := cmd.CombinedOutput()

	if err != nil {
		makeup.PrintFail("Failed to execute the command: " + err.Error())
		fmt.Println(string(output))
		return false
	} else {
		fmt.Println(string(output))
		return true
	}
}

func (c KindCreator) Delete(name string, unattendedMode bool) {
	makeup.PrintWarning("Deleting the Demo Kubernetes Cluster " + name + "...")

	cmd := exec.Command("kind", "delete", "cluster", "-n", name)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(unattendedMode)

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
Returns the context name for the given Kind cluster.
*/
func (c KindCreator) GetContext(kindClusterName string) string {
	return "kind-" + kindClusterName
}
