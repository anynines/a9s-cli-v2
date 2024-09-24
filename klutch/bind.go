package klutch

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	prereq "github.com/anynines/a9s-cli-v2/prerequisites"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
)

//go:embed manifests/examples/postgresql-example.yaml
var postgresqlExample string

//go:embed manifests/examples/servicebinding-example.yaml
var servicebindingExample string

//go:embed manifests/examples/backup-example.yaml
var backupExample string

//go:embed manifests/examples/restore-example.yaml
var restoreExample string

const (
	konnectorImage = "public.ecr.aws/w5n9a2g2/anynines/konnector:v1.3.0"
)

type NamespacedName = types.NamespacedName

// Helper type to extract the resource from an APIServiceExportRequest manifest
// TODO: can use actual type once Klutch is open source
type serviceExportRequest struct {
	Spec struct {
		Resources []struct {
			Group    string `yaml:"group"`
			Resource string `yaml:"resource"`
		} `yaml:"resources"`
	} `yaml:"spec"`
}

func Bind() {
	makeup.PrintWelcomeScreen(
		demo.UnattendedMode,
		demoTitle,
		"Let's bind an API from the App Cluster to the Control Plane Cluster...")

	demo.EstablishConfig()

	checkBindPrerequisites()

	km := NewKlutchManager()
	resource := km.bindResource()

	printBindSummary()
	printResourceSuggestion(resource)
}

func (k *KlutchManager) bindResource() string {
	controlPlaneInfo := getControlPlaneClusterInfoFromFile(demo.DemoConfig.WorkingDir)

	// Only proceed if the backend is ready.
	checkBackendEndpoint(controlPlaneInfo)

	secret, exportRequestYaml := k.startInteractiveBind(controlPlaneInfo)
	k.finishInteractiveBinding(secret, *exportRequestYaml)

	resource, group, err := determineBoundResource(exportRequestYaml)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to determine which resource was bound")
	}

	k.appK8s.WaitForCRDCreationAndReady(resource + "." + group)

	return resource
}

// Starts the interactive binding process by calling `kubectl bind` against the App cluster,
// with the URL targetting the Control Plane Cluster backend.
// Automatically opens the URL presented by bind's output, waits for the user to authorize in the browser,
// and captures the resulting APIServiceExportRequest yaml file and other needed information for the next phase.
// Returns the captured secret name and namespace, and a buffer containing the yaml file.
func (k *KlutchManager) startInteractiveBind(info ControlPlaneInfo) (NamespacedName, *bytes.Buffer) {
	url := getExportUrl(info)

	cmd := k.appK8s.KubectlWithContextCommand(
		"bind",
		url,
		"--konnector-image",
		konnectorImage,
	)

	// Print the non- dry-run command to avoid confusion.
	// We use --dry-run for automation purposes.
	makeup.PrintCommandBox(cmd.String())
	makeup.PrintBright("This process will open a browser window for you. Authenticate with \"admin@example.com\" and \"password\", then select the API you wish to bind.")
	makeup.WaitForUser(demo.UnattendedMode)

	cmd.Args = append(cmd.Args, "--dry-run", "-o", "yaml")

	// Stdout will print the yaml manifest that needs to be applied in the next phase.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	// Stderr outputs the following information to be extracted:
	// - the authorization URL that needs to be opened in a browser.
	// - after authorization has been executed in the browser, the name of the secret and its namespace.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	secretChan := make(chan NamespacedName, 1)
	stderrErrChan := make(chan error, 1)

	yamlChan := make(chan *bytes.Buffer, 1)
	stdoutErrChan := make(chan error, 1)

	go scanInteractiveBindStderr(stderr, info, secretChan, stderrErrChan)
	go scanInteractiveBindStdout(stdout, yamlChan, stdoutErrChan)

	if err := cmd.Start(); err != nil {
		makeup.ExitDueToFatalError(err, "Could not start the bind command.")
	}

	if err := cmd.Wait(); err != nil {
		makeup.ExitDueToFatalError(err, "Error occured while executing the bind command.")
	}

	select {
	case err := <-stderrErrChan:
		makeup.ExitDueToFatalError(err, "Error occured while executing the bind command.")
	case err := <-stdoutErrChan:
		makeup.ExitDueToFatalError(err, "Error occured while executing the bind command.")
	default:
		break
	}

	var secret NamespacedName
	var exportRequestYaml *bytes.Buffer

	// At this point, the channels should have data written to them. If they don't, something went wrong.
	select {
	case secret = <-secretChan:
	default:
		makeup.ExitDueToFatalError(err, "Error occured while executing the bind command.")
	}

	select {
	case exportRequestYaml = <-yamlChan:
	default:
		makeup.ExitDueToFatalError(err, "Error occured while executing the bind command.")
	}

	return secret, exportRequestYaml
}

// Scans the bind command's stderr for the URL to open (and opens it) and the secret printed afterward.
func scanInteractiveBindStderr(stderr io.ReadCloser, info ControlPlaneInfo, secretChan chan NamespacedName, errChan chan error) {
	urlFound := false
	secret := NamespacedName{}

	// The following patterns can break if the backend/konnector change their output format. Review them if changing image versions.
	urlPattern := regexp.MustCompile(fmt.Sprintf(`http://%s:%s/authorize\?.*`, info.Host, info.IngressPort))
	// We are looking for <namespace>/<name> where namespace and name are k8s compliant.
	secretPattern := regexp.MustCompile(`secret ([a-z0-9][-a-z0-9]*[a-z0-9]?)/([a-z0-9][-a-z0-9]*[a-z0-9]?) for host`)

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		// First part: find the URL we need to open, and open it.
		if !urlFound {
			url := urlPattern.FindString(line)
			if url != "" {
				urlFound = true
				err := openURL(url)
				if err != nil {
					// Instead of returning an error, be lenient because the user can still open
					// the URL by hand if openURL didn't work (e.g. xdg-open not present)
					makeup.PrintWarning("Could not open the URL automatically. Please open above URL in a browser to proceed.")
					continue
				}
			}

			// don't bother scanning for the secret yet.
			continue
		}

		// Second part: after the user has authorized and selected an API to bind, find the secret that was created.
		matches := secretPattern.FindStringSubmatch(line)
		if len(matches) == 3 {
			secret.Namespace = matches[1]
			secret.Name = matches[2]
			secretChan <- secret
			close(secretChan)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		errChan <- fmt.Errorf("error reading from the command stderr: %v", err)
	}

	errChan <- fmt.Errorf("did not find expected secret in the command output")

	close(errChan)
}

// Scans the bind command's stdout for the yaml manifest we want to apply.
// Assumes that the yaml starts with "apiVersion:" and ends when the stdout stream ends.
func scanInteractiveBindStdout(stdout io.ReadCloser, yamlChan chan *bytes.Buffer, errChan chan error) {
	yamlBytes := &bytes.Buffer{}
	yamlFound := false

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if !yamlFound && isYAMLStart(line) {
			yamlFound = true
		}

		if yamlFound {
			yamlBytes.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from stdout:", err)
		errChan <- err
		close(errChan)
		return
	}

	// We rely on the stdout closing to exit the above loop.
	yamlChan <- yamlBytes
	close(yamlChan)
}

// Finishes the interactive binding process by calling `kubectl bind apiservice` with the extracted secret name and namespace and APIServiceBinding yaml.
func (k *KlutchManager) finishInteractiveBinding(secret NamespacedName, exportRequestYaml bytes.Buffer) {
	secretName := secret.Name
	secretNamespace := secret.Namespace

	yamlTempFile, err := os.CreateTemp(os.TempDir(), "bind-request-yaml.yaml")
	if err != nil {
		makeup.ExitDueToFatalError(err, "Error occured while setting up the bind command.")
	}

	_, err = yamlTempFile.Write(exportRequestYaml.Bytes())
	if err != nil {
		makeup.ExitDueToFatalError(err, "Error occured while setting up the bind command.")
	}

	cmd := k.appK8s.KubectlWithContextCommand(
		"bind",
		"apiservice",
		"--remote-kubeconfig-namespace",
		secretNamespace,
		"--remote-kubeconfig-name",
		secretName,
		"--konnector-image",
		konnectorImage,
		"-f",
		yamlTempFile.Name(),
	)

	// We have to write "Yes" to "Yes/No" questions via stdin.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	// Stdout will print the [Yes/No] prompts we want to answer.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	// Print stderr to show progress to the user.
	go printCommandOutput(stderr)

	go func() {
		scanner := bufio.NewScanner(stdout)
		yesNo := regexp.MustCompile(`\[No,Yes\]`)
		writer := bufio.NewWriter(stdin)

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // Print to show progress to the user.

			if prompt := yesNo.FindString(line); prompt != "" {
				fmt.Fprintf(writer, "Yes\n")
				err := writer.Flush()
				if err != nil {
					makeup.ExitDueToFatalError(err, "error while flushing writer")
				}
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		makeup.ExitDueToFatalError(err, "Could not execute the bind command.")
	}

	if err := cmd.Wait(); err != nil {
		makeup.ExitDueToFatalError(err, "Could not execute the bind command.")
	}

	err = os.Remove(yamlTempFile.Name())
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not remove file %s", yamlTempFile.Name()))
	}
}

// determineBoundResource parses the APIServiceExportRequest yaml and returns the resource name and group that was bound.
func determineBoundResource(exportRequestYaml *bytes.Buffer) (string, string, error) {
	var data serviceExportRequest

	err := yaml.Unmarshal(exportRequestYaml.Bytes(), &data)
	if err != nil {
		return "", "", err
	}

	if len(data.Spec.Resources) > 0 {
		// During interactive binding, only one resource can be bound at a time,
		// so we take the first element.
		resource := data.Spec.Resources[0]
		return resource.Resource, resource.Group, nil
	} else {
		return "", "", fmt.Errorf("no resource found in manifest")
	}
}

// Prints an example yaml manifest for the resource which was bound during the process.
func printResourceSuggestion(resource string) {
	var file string

	switch strings.ToLower(resource) {
	case "postgresqlinstances":
		file = postgresqlExample
	case "servicebindings":
		file = servicebindingExample
	case "backups":
		file = backupExample
	case "restores":
		file = restoreExample
	}

	makeup.PrintCheckmark(fmt.Sprintf("You've bound the %s resource. You can now apply instances of this resource, for example with the following yaml:", resource))
	makeup.PrintYAML(bytes.NewBufferString(file).Bytes(), false)
}

func printBindSummary() {
	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Called the kubectl bind plugin to start the interactive binding process")
	makeup.PrintCheckmark("Authorized the Control Plane Cluster to manage the selected API on your App Cluster.")
}

func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Run()
}

// Loads the Control Plane Cluster information from the workspace. If it doesn't exists, prints a suggestion to the user to run the
// deploy command first.
func getControlPlaneClusterInfoFromFile(workDir string) ControlPlaneClusterInfo {
	path := filepath.Join(workDir, controlPlaneClusterInfoFilePath, controlPlaneClusterInfoFilePath)
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			makeup.ExitDueToFatalError(err, "The Control Plane Cluster info file does not exist. Have you created a Control Plane Cluster with `deploy`?")
		}

		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while reading from Control Plane Cluster info to file %s", path))
	}

	var controlPlaneClusterInfo ControlPlaneClusterInfo
	err = yaml.Unmarshal(bytes, &controlPlaneClusterInfo)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while reading from Control Plane Cluster info to file %s", path))
	}

	if controlPlaneClusterInfo.IngressPort == "" || controlPlaneClusterInfo.Host == "" {
		makeup.ExitDueToFatalError(err, "The Control Plane Cluster info file is incomplete. Please try deploying the Control Plane Cluster again with `deploy`")
	}

	return controlPlaneClusterInfo
}

// Checks if a given string represents the start of a the APIServiceExportRequest yaml file.
func isYAMLStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "apiVersion:")
}

// Checks if prerequisites of the bind command are met.
func checkBindPrerequisites() {
	makeup.PrintH1("Checking Prerequisites...")

	commonTools := prereq.GetCommonRequiredTools()

	requiredTools := []prereq.RequiredTool{
		commonTools[prereq.ToolKubectl], commonTools[prereq.ToolBind],
	}

	prereq.CheckRequiredTools(requiredTools)

	prereq.CheckDockerRunning()
}

// checkBackendEndpoint checks if the backend is reachable.
func checkBackendEndpoint(info ControlPlaneClusterInfo) {
	url := getExportUrl(info)

	resp, err := http.Get(url)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Got unexpected error trying to reach the backend. Please verify or wait for the Control Plane Cluster to be fully ready.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		makeup.ExitDueToFatalError(nil, "The backend does not appear to be ready. Please verify or wait for the Control Plane Cluster to be fully ready.")
	}
}

// getExportUrl returns the export URL of the backend.
func getExportUrl(info ControlPlaneClusterInfo) string {
	url := (&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", info.Host, info.IngressPort),
		Path:   "export",
	})

	return url.String()
}
