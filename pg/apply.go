package pg

/*
a8s PG specific apply functions
*/
import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
)

const RemoteUploadDir = "/tmp"

// Name of the container within the target pod to upload the file to.
const RemoteUploadContainerName = "postgres"
const a8sPGDefaultDatabaseName = "a9s_apps_default_db"
const a8sPGDefaultDatabaseUser = "postgres"

// Filename when used with option --file or -f
var SQLFilename = ""

/*
Identifies the current primary of the given PG service instance.

Executes: kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
*/
func FindPrimaryPodOfServiceInstance(namespace, serviceInstanceName string) string {

	instanceLabel := fmt.Sprintf("%s=%s", A8sPGServiceInstanceNameLabelKey, serviceInstanceName)

	label := fmt.Sprintf("%s,%s,%s", A8sPGLabelPrimary, A8sPGServiceInstanceAPIGroupLabel, instanceLabel)

	podName, err := k8s.FindFirstPodByLabel(namespace, label)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't find primary pod of service instance: %s/%s", namespace, serviceInstanceName))
	}

	return podName
}

/*
Executes a SQL file within a Pod using PSQL.
Both the sql file and the psql command must be present in the target pod.

Example:
kubectl exec -n default clustered-0 -c postgres -- psql -U postgres -d a9s_apps_default_db -f demo_data.sql
*/
func ExecuteSQLFileWithinPod(unattendedMode bool, namespace, serviceInstanceName, remoteSQLFilePath string) error {

	commandElements := make([]string, 0)
	commandElements = append(commandElements, "exec")
	commandElements = append(commandElements, "-n")
	commandElements = append(commandElements, namespace)
	commandElements = append(commandElements, serviceInstanceName)
	commandElements = append(commandElements, "-c")
	commandElements = append(commandElements, RemoteUploadContainerName)
	commandElements = append(commandElements, "--")
	commandElements = append(commandElements, "psql")
	commandElements = append(commandElements, "-U")
	commandElements = append(commandElements, a8sPGDefaultDatabaseUser)
	commandElements = append(commandElements, "-d")
	commandElements = append(commandElements, a8sPGDefaultDatabaseName)
	commandElements = append(commandElements, "-f")
	commandElements = append(commandElements, remoteSQLFilePath)

	k8s.Kubectl(unattendedMode, commandElements...)

	return nil
}

// TODO Remove code duplication with ExecuteSQLFileWithinPod
func ExecuteSQLStatementWithinPod(unattendedMode bool, namespace, serviceInstanceName, sqlStatement string) (*exec.Cmd, []byte, error) {

	commandElements := make([]string, 0)
	commandElements = append(commandElements, "exec")
	commandElements = append(commandElements, "-n")
	commandElements = append(commandElements, namespace)
	commandElements = append(commandElements, serviceInstanceName)
	commandElements = append(commandElements, "-c")
	commandElements = append(commandElements, RemoteUploadContainerName)
	commandElements = append(commandElements, "--")
	commandElements = append(commandElements, "psql")
	commandElements = append(commandElements, "-U")
	commandElements = append(commandElements, a8sPGDefaultDatabaseUser)
	commandElements = append(commandElements, "-d")
	commandElements = append(commandElements, a8sPGDefaultDatabaseName)
	commandElements = append(commandElements, "-c")
	commandElements = append(commandElements, sqlStatement)

	return k8s.Kubectl(unattendedMode, commandElements...)
}

/*
Upload and execute psql for the given file on the primary of the given service instance.

If noDelete is set to true, the uploaded file won't be deleted after a successful apply.
Not that, if errors occur, the process will be stopped and hence, an uploaded file will remain
in the pod and not be deleted, in case its execution failed.

Example kubectl command: kubectl exec -n default clustered-0 -c postgres -- psql -U postgres -d a9s_apps_default_db -f demo_data.sql
*/
func ApplySQLFileToPGServiceInstance(unattendedMode bool, namespace, serviceInstanceName, sqlFileToUpload string, noDelete bool) {
	// Determine primary pod name
	podName := FindPrimaryPodOfServiceInstance(namespace, serviceInstanceName)

	if podName != "" {
		makeup.PrintVerbose(fmt.Sprintf("Found primary pod %s of service instance %s", podName, serviceInstanceName))
	} else {
		makeup.ExitDueToFatalError(nil, "Can't find primary pod of service instance "+serviceInstanceName)
	}

	// Upload
	err := k8s.KubectlUploadFileToPod(unattendedMode, namespace, podName, RemoteUploadContainerName, sqlFileToUpload, RemoteUploadDir)

	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Can't upload file %s to service instance %s", sqlFileToUpload, serviceInstanceName))
	}

	// Handle filepaths for sqlFileToUpload > Preserve the filename
	sqlFilename := filepath.Base(sqlFileToUpload)
	remoteSQLFilePath := filepath.Join(RemoteUploadDir, sqlFilename)

	// Apply SQL file using psql
	err = ExecuteSQLFileWithinPod(unattendedMode, namespace, podName, remoteSQLFilePath)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't execute provided SQL file to pod "+podName)
	}

	if !noDelete {

		// Delete file
		err = k8s.KubectlDeleteFileFromPod(unattendedMode, namespace, podName, RemoteUploadContainerName, remoteSQLFilePath)

		if err != nil {
			makeup.ExitDueToFatalError(err, "Couldn't delete uploaded SQL file from pod "+podName)
		}
	}

	makeup.PrintCheckmark("Successfully applied SQL file to pod " + podName + "")
}

// TODO Reduce code duplication with ApplySQLFileToPGServiceInstance
func ApplySQLStatementToPGServiceInstance(unattendedMode bool, namespace, serviceInstanceName, sqlStatement string) {

	// Determine primary pod name
	podName := FindPrimaryPodOfServiceInstance(namespace, serviceInstanceName)

	if podName != "" {
		makeup.PrintVerbose(fmt.Sprintf("Found primary pod %s of service instance %s", podName, serviceInstanceName))
	} else {
		makeup.ExitDueToFatalError(nil, "Can't find primary pod of service instance "+serviceInstanceName)
	}

	// Apply SQL file using psql
	_, output, err := ExecuteSQLStatementWithinPod(unattendedMode, namespace, podName, sqlStatement)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't execute provided SQL file to pod "+podName)
	}

	makeup.PrintBright(string(output))

	makeup.PrintCheckmark("Successfully applied SQL file to pod " + podName + "")
}
