package k8s

/*
Functions interacting with the kubectl command.
*/

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/anynines/a9s-cli-v2/makeup"
)

var ErrNotFound = errors.New("resource was not found")

// KubectlWithContextCommand returns a kubectl command using the specified kubeconfig context. If the context is empty, it is not used.
func (k *KubeClient) KubectlWithContextCommand(args ...string) *exec.Cmd {
	if k.KubeContext != "" {
		args = append(args, "--context", k.KubeContext)
	}

	return exec.Command("kubectl", args...)
}

/*
Variadic function to use kubectl.
If kubeContext is not empty, it is added as the --context parameter.

Returns: cmd, output, err
*/
func (k *KubeClient) Kubectl(unattendedMode bool, kubectlArg ...string) (*exec.Cmd, []byte, error) {
	cmd := k.KubectlWithContextCommand(kubectlArg...)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(unattendedMode)

	output, err := cmd.CombinedOutput()

	if makeup.Verbose {
		fmt.Println(string(output))
	}

	return cmd, output, err
}

/*
Examples:
delete postgresql {name}
apply -f {path}
delete -f {path}
*/
func (k *KubeClient) KubectlAct(verb, flag, filepath string, waitForUser bool) {
	// cmd := exec.Command("kubectl", verb, flag, filepath)

	cmd, _, err := k.Kubectl(waitForUser, verb, flag, filepath)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl "+verb+" with using the command: "+cmd.String())
	}
}

func (k *KubeClient) KubectlApplyF(yamlFilepath string, waitForUser bool) {
	k.KubectlAct("apply", "-f", yamlFilepath, waitForUser)
}

func (k *KubeClient) KubectlDeleteF(yamlFilepath string, waitForUser bool) {
	k.KubectlAct("delete", "-f", yamlFilepath, waitForUser)
}

func (k *KubeClient) KubectlApplyKustomize(kustomizeFilepath string, waitForUser bool) {
	k.KubectlAct("apply", "--kustomize", kustomizeFilepath, waitForUser)
}

// KubectlApplyStdin Calls "kubectl apply -f -" with the given bytes buffer as input.
func (k *KubeClient) KubectlApplyStdin(in *bytes.Buffer) {
	cmd := k.KubectlWithContextCommand("apply", "-f", "-")
	cmd.Stdin = in

	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not apply resources: %v ", string(output)))
	}
}

/*
Uploads the given file to the given container in the given pod to the given
remote target folder.

Example kubectl command: kubectl cp demo_data.sql default/clustered-0:/home/postgres -c postgres
*/
func (k *KubeClient) KubectlUploadFileToPod(unattendedMode bool, namespace, podName, containerName, fileToUpload, remoteTargetFolder string) error {
	commandElements := make([]string, 0)

	commandElements = append(commandElements, "cp")
	commandElements = append(commandElements, fileToUpload)
	commandElements = append(commandElements, namespace+"/"+podName+":"+remoteTargetFolder)
	commandElements = append(commandElements, "-c")
	commandElements = append(commandElements, containerName)

	_, _, err := k.Kubectl(unattendedMode, commandElements...)

	return err
}

/*
Deletes a file in the remote pod/container.
Executes command similar to: kubectl exec solo-0 -n default -c postgres -- rm /tmp/demo.sql
*/
func (k *KubeClient) KubectlDeleteFileFromPod(unattendedMode bool, namespace, podName, containerName, remoteFilename string) error {
	commandElements := make([]string, 0)

	commandElements = append(commandElements, "exec")
	commandElements = append(commandElements, podName)
	commandElements = append(commandElements, "-n")
	commandElements = append(commandElements, namespace)
	commandElements = append(commandElements, "-c")
	commandElements = append(commandElements, containerName)
	commandElements = append(commandElements, "--")
	commandElements = append(commandElements, "rm")
	commandElements = append(commandElements, remoteFilename)

	_, _, err := k.Kubectl(unattendedMode, commandElements...)

	return err
}

/*
Executes similar to: kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
*/
func (k *KubeClient) FindFirstPodByLabel(namespace, label string) (string, error) {

	// Ignore the Don't Execute flag
	unattendedMode := true

	// kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
	// output := "clustered-0 clustered-1 clustered-2 solo-0"

	commandElements := make([]string, 0)
	commandElements = append(commandElements, "get")
	commandElements = append(commandElements, "pods")

	// Namespace
	commandElements = append(commandElements, "-n")
	commandElements = append(commandElements, namespace)

	// Labels
	commandElements = append(commandElements, "-l")
	commandElements = append(commandElements, label)

	// Output jsonpath
	// In a shell you want to quote 'jsonpath ...'
	// Here you don't otherwise you'll get '' added to the result string
	commandElements = append(commandElements, "-o=jsonpath={.items[*].metadata.name}")

	cmd, output, err := k.Kubectl(unattendedMode, commandElements...)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl using the command: "+cmd.String())
	}

	outputString := string(output)
	if outputString == "" {
		return "", ErrNotFound
	}

	podNames := strings.Fields(outputString)

	if len(podNames) > 0 {
		podName := podNames[0]
		return podName, nil
	} else {
		return "", ErrNotFound
	}
}
