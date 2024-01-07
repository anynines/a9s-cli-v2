package k8s

/*
Functions interacting with the kubectl command.
*/

import (
	"fmt"
	"os/exec"

	"github.com/anynines/a9s-cli-v2/makeup"
)

/*
Examples:
delete postgresql {name}
apply -f {path}
delete -f {path}
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

// /*
// Uploads the given file to the tmp folder within the target pod.
// */
func KubectlUploadFileToTmp() error {

	return nil
}

// /*
// Deletes a file from the remote tmp directory.
// */
// func KubectlDeleteTmpFile() error {
// 	return nil
// }

// /*
// Deletes a file in the remote pod/container.
// */
// func KubectlDeleteFile() error {
// 	return nil
// }

// /*
// Executes the kubectl exec command.
// */
// func KubectlExec() error {
// 	return nil
// }

// /*
// Executes the kubectl cp command.

// Note that the tar command must be present within the target pod or
// kubectl cp will fail.
// */
// func KubectlCp() error {

// 	return nil
// }
