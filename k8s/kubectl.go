package k8s

/*
Functions interacting with the kubectl command.
*/

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/anynines/a9s-cli-v2/makeup"
)

var ErrNotFound = errors.New("Resource was not found")

/*
Variadic function to use kubectl.
*/
func Kubectl(waitForUser bool, kubectlArg ...string) (*exec.Cmd, []byte, error) {

	cmd := exec.Command("kubectl", kubectlArg...)

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(waitForUser)

	output, err := cmd.CombinedOutput()

	//TODO Use makeup
	fmt.Println(string(output))

	return cmd, output, err
}

/*
Examples:
delete postgresql {name}
apply -f {path}
delete -f {path}
*/
func KubectlAct(verb, flag, filepath string, waitForUser bool) {
	// cmd := exec.Command("kubectl", verb, flag, filepath)

	cmd, _, err := Kubectl(waitForUser, verb, flag, filepath)

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

// /*
// Uploads the given file to the tmp folder within the target pod.
// */
func KubectlUploadFileToTmp() error {

	return nil
}

/*
Uploads the given file to the given container in the given pod to the given
remote target folder.
*/
func KubectlUploadFileToPod(namespace, podName, containerName, fileToUpload, remoteTargetFolder string) error {
	return nil
}

/*
Deletes a file from the remote tmp directory.
*/
func KubectlDeleteTmpFile() error {
	return nil
}

/*
Deletes a file in the remote pod/container.
*/
func KubectlDeleteFileInPod(namespace, podName, containerName, remoteFilename string) error {
	return nil
}

/*
Executes the kubectl exec command.
*/
func KubectlExec() error {
	return nil
}

/*
Executes the kubectl cp command.

Note that the tar command must be present within the target pod or
kubectl cp will fail.
*/
func KubectlCp() error {

	return nil
}

/*
Executes similar to: kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
*/
func FindFirstPodByLabel(namespace, label string) (string, error) {

	// Ignore the Don't Execute flag
	waitForUser := false

	// kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
	// output := "clustered-0 clustered-1 clustered-2 solo-0"
	cmd, output, err := Kubectl(waitForUser, label)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl using the command: "+cmd.String())
	}

	outputString := string(output)
	if outputString == "" {
		return "", ErrNotFound
	}

	podNames := strings.Fields(outputString)

	return podNames[0], nil
}
