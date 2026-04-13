package k8s

/*
Functions interacting with the kubectl command.
*/

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var ErrNotFound = errors.New("resource was not found")

/*
Uploads the given file to the given container in the given pod to the given
remote target folder.

Example kubectl command: kubectl cp demo_data.sql default/clustered-0:/home/postgres -c postgres
*/
func (k *KubeClient) KubectlUploadFileToPod(namespace, podName, containerName, fileToUpload, remoteTargetFolder string) error {
	opts := KubectlOpts{
		Command:        "cp",
		Kind:           fileToUpload,
		Name:           namespace + "/" + podName + ":" + remoteTargetFolder,
		AdditionalArgs: []string{"-c", containerName},
	}

	_, _, err := runKubeCtlCommand(opts.withContextFrom(k))

	return err
}

/*
Deletes a file in the remote pod/container.
Executes command similar to: kubectl exec solo-0 -n default -c postgres -- rm /tmp/demo.sql
*/
func (k *KubeClient) KubectlDeleteFileFromPod(namespace, podName, containerName, remoteFilename string) error {
	opts := KubectlOpts{
		Command:   "exec",
		Kind:      podName,
		Namespace: namespace,
		AdditionalArgs: []string{
			"-c", containerName,
			"--",
			"rm", remoteFilename,
		},
	}

	_, _, err := runKubeCtlCommand(opts.withContextFrom(k))

	return err
}

/*
Executes similar to: kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
*/
func (k *KubeClient) FindFirstPodByLabel(namespace, label string) (string, error) {

	// kubectl get pods -n default -l 'a8s.a9s/replication-role=master,a8s.a9s/dsi-group=postgresql.anynines.com,a8s.a9s/dsi-kind=Postgresql,a8s.a9s/dsi-name=clustered' -o=jsonpath='{.items[*].metadata.name}'
	// output := "clustered-0 clustered-1 clustered-2 solo-0"

	opts := KubectlOpts{
		Command:      "get",
		Kind:         "pods",
		Namespace:    namespace,
		Selector:     label,
		OutputFormat: "jsonpath={.items[*].metadata.name}",
	}

	cmd, output, err := runKubeCtlCommand(opts.withContextFrom(k))

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't kubectl using the command: "+cmd)
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

var commandsWithoutSubcommands = []string{
	"api-resources",
	"cluster-info",
	"version",
}

// execCommand is the function used to create external commands.
// Replaced in unit tests to prevent actual kubectl execution.
// var execCommand = exec.Command
var execCommand = func(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// execCommandContext is the context-aware variant of execCommand.
// Replaced in unit tests to prevent actual kubectl execution.
var execCommandContext = exec.CommandContext

// kubectlWithContextCommand returns a kubectl command using the specified kubeconfig context. If the context is empty, it is not used.
func (k *KubeClient) kubectlWithContextCommand(args ...string) *exec.Cmd {
	if k.KubeContext != "" {
		args = append(args, "--context", k.KubeContext)
	}

	return execCommand("kubectl", args...)
}

func (k *KubeClient) Create(kind string, name string, namespace string) ([]byte, error) {
	opts := KubectlOpts{
		Command:   "create",
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}

	_, output, err := runKubeCtlCommand(opts.withContextFrom(k))
	return output, err
}

// showManifestInPager displays content in a pager ($PAGER, defaulting to "less -R").
func showManifestInPager(content []byte) {
	pager := os.Getenv("PAGER")
	if pager == "" || pager == "less" {
		pager = "less -R"
	}
	parts := strings.Fields(pager)
	if _, err := makeup.Command(parts[0], parts[1:]...).Stdin(content).Quiet().Interactive().Run(); err != nil {
		fmt.Println(string(content))
	}
}

// promptUserForExecution prompts the user to either view the manifest or execute the command.
// Returns true if the user wants to execute, false to abort.
func promptUserForExecution(manifestBytes []byte, command string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n  " + makeup.Bright("What would you like to do?"))
	fmt.Println("    [v]iew:      View manifest")
	fmt.Println("    [e]xecute:   Execute kubectl " + command)
	fmt.Println("    [c]ancel:    Cancel/abort")
	fmt.Print("\n Your choice (v/e/c) or press Enter to execute: ")

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		choice := strings.ToLower(strings.TrimSpace(input))

		switch choice {
		case "v", "view":
			tokens, p := makeup.GetColoredYAML(manifestBytes)
			showManifestInPager([]byte(p.PrintTokens(tokens)))
		case "", "e", "execute":
			return true
		case "c", "cancel", "abort", "n", "no":
			return false
		default:
			makeup.PrintWarning(fmt.Sprintf("Invalid choice '%s'. Please enter 'v' to view, 'e' to execute, or 'c' to cancel.", choice))
		}
		// After returning from the pager, erase the current line and reprint
		// the prompt so the user has a clean input line.
		fmt.Print("\r\033[2K Your choice (v/e/c) or press Enter to execute: ")
	}
}

// ApplyWithPrompt is the unified wrapper for all kubectl apply operations.
// It handles manifest display, user prompts, and the actual apply operation.
// This is the ONLY function that should be used to apply manifests - all other code
// should call this directly instead of using exec.Command or other wrappers.
// Parameters:
//   - manifestBytes: the YAML manifest content to apply
//   - description: a brief description of what is being applied (for user feedback)
func (k *KubeClient) ApplyWithPrompt(manifestBytes []byte, description string) (string, error) {
	return k.kubectlCommandWithPrompt("apply", manifestBytes, description)
}

// Delete is the unified wrapper for all kubectl delete operations.
// It handles manifest display, user prompts, and the actual delete operation.
// This is the ONLY function that should be used to delete manifests - all other code
// should call this directly instead of using exec.Command or other wrappers.
// Parameters:
//   - resource: resource identifier (e.g. Kind) of the resource to delete
//   - name: name of the resource to delete
//   - namespace: namespace of the resource to delete
//   - description: a brief description of what is being deleted (for user feedback)
//   - ignoreNotFound: indicates whether to set the "--ignore-not-found" flag when retrieving the resource
func (k *KubeClient) Delete(resource, name, namespace, description string, ignoreNotFound bool) (string, error) {
	manifest, err := k.Get(resource, name, namespace, "yaml", ignoreNotFound)
	if err != nil {
		return "", fmt.Errorf("could not check resource to delete: %w", err)
	}
	if len(manifest) == 0 {
		message := ""
		if namespace != "" {
			message = fmt.Sprintf(" in namespace %s is already gone.", namespace)
		}
		makeup.PrintCheckmark(fmt.Sprintf("Resource %s/%s%s is already gone.", resource, name, message))
		return "", nil
	}
	return k.kubectlCommandWithPrompt("delete", []byte(manifest), description)
}

func (k *KubeClient) Get(resource string, name string, namespace, format string, ignoreNotFound bool) (string, error) {
	opts := KubectlOpts{
		Command:        "get",
		Kind:           resource,
		Name:           name,
		Namespace:      namespace,
		OutputFormat:   format,
		IgnoreNotFound: ignoreNotFound,
	}

	_, outBytes, err := runKubeCtlCommand(opts.withContextFrom(k))
	return string(outBytes), err
}

func (k *KubeClient) Describe(resource string, name string, namespace string) ([]byte, error) {
	opts := KubectlOpts{
		Command:   "describe",
		Kind:      resource,
		Name:      name,
		Namespace: namespace,
	}
	_, output, err := runKubeCtlCommand(opts.withContextFrom(k))
	return output, err
}

func (k *KubeClient) Bind(url, konnectorImage string) (io.ReadCloser, io.ReadCloser, func() error, func() error) {
	cmd := k.kubectlWithContextCommand(
		"bind",
		url,
		"--konnector-image",
		konnectorImage,
		"--dry-run",
		"-o", "yaml",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	return stdout, stderr, cmd.Start, cmd.Wait
}
func (k *KubeClient) BindApiService(secretNamespace, secretName, konnectorImage, yamlTempFile string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, func() error, func() error) {
	cmd := k.kubectlWithContextCommand(
		"bind",
		"apiservice",
		"--remote-kubeconfig-namespace",
		secretNamespace,
		"--remote-kubeconfig-name",
		secretName,
		"--konnector-image",
		konnectorImage,
		"-f",
		yamlTempFile,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not set up the bind command.")
	}

	return stdin, stdout, stderr, cmd.Start, cmd.Wait
}

func (k *KubeClient) kubectlCommandWithPrompt(command string, manifestBytes []byte, description string) (string, error) {
	if description != "" {
		makeup.PrintInfo("Preparing to " + command + " " + description)
	}

	// If verbose mode is enabled, always print the manifest
	if makeup.Verbose {
		fmt.Println("\n" + makeup.H2("Manifest to "+command+":"))
		makeup.PrintYAML(manifestBytes, true)
	}

	// In unattended mode (--yes flag), execute immediately without prompting
	if makeup.UnattendedMode {
		makeup.PrintInfo("Executing '" + command + "' (unattended mode)...")
	} else {
		// Interactive mode: prompt user unless verbose already showed the manifest
		shouldExecute := promptUserForExecution(manifestBytes, command)
		if !shouldExecute {
			return "", fmt.Errorf("user cancelled kubectl %s", command)
		}
	}

	// Execute the actual kubectl command

	opts := KubectlOpts{
		Command:  command,
		StdIn:    manifestBytes,
		Filename: "-",
	}

	_, output, err := runKubeCtlCommand(opts.withContextFrom(k))
	if err != nil {
		makeup.PrintFail(fmt.Sprintf("failed to "+command+" manifest:\n%s\n%s", string(output), string(manifestBytes)))
		return string(output), err
	}

	if makeup.Verbose {
		fmt.Println(string(output))
	}

	makeup.PrintSuccess("Executed " + command + " " + description + " successfully.")
	return string(output), nil
}

// ApplyFromFile is a convenience wrapper that reads a file and applies it.
// This handles the -f flag behavior - reading from a file path.
// Parameters:
//   - yamlFilepath: path to the YAML manifest file (can be a URL or local file path)
//   - unattendedMode: if true, skip all prompts and apply immediately (respects --yes flag)
func (k *KubeClient) ApplyFromFile(yamlFilepath string) (string, error) {
	// Read the file content
	info, err := os.Stat(yamlFilepath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		entries, err := os.ReadDir(yamlFilepath)
		if err != nil {
			return "", err
		}
		output := ""
		for _, entry := range entries {
			entryOutput, err := k.ApplyFromFile(filepath.Join(yamlFilepath, entry.Name()))
			output += entryOutput
			if err != nil {
				return entryOutput, err
			}
		}
		return output, nil
	}
	manifestBytes, err := os.ReadFile(yamlFilepath)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest file %s: %w", yamlFilepath, err)
	}

	// Use the unified apply wrapper
	return k.ApplyWithPrompt(manifestBytes, yamlFilepath)
}

// ApplyKustomize is a convenience wrapper that renders kustomize output and applies it.
// This handles the -k flag behavior - rendering a kustomize directory.
// Parameters:
//   - kustomizeFilepath: path to the directory containing kustomization.yaml
//   - unattendedMode: if true, skip all prompts and apply immediately (respects --yes flag)
func (k *KubeClient) ApplyKustomize(kustomizeFilepath string) (string, error) {
	renderOpts := KubectlOpts{
		Command: "kustomize",
		Kind:    kustomizeFilepath,
	}

	// Generate the kustomize output first
	_, output, err := runKubeCtlCommand(renderOpts.withContextFrom(k))
	if err != nil {
		return "", fmt.Errorf("failed to render kustomize from %s: %w", kustomizeFilepath, err)
	}

	// Use the unified apply wrapper
	return k.ApplyWithPrompt(output, fmt.Sprintf("kustomize: %s", kustomizeFilepath))
}

// ApplyFromUrl is a convenience wrapper that fetches manifests from a URL and applies them.
// Parameters:
//   - manifestUrl: the URL to the manifest content to apply
//   - description: a brief description of what is being applied (for user feedback)
//   - unattendedMode: if true, skip all prompts and apply immediately (respects --yes flag)
func (k *KubeClient) ApplyFromUrl(manifestUrl, description string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(manifestUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", manifestUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: HTTP %d", manifestUrl, resp.StatusCode)
	}
	manifest, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body %s: %w", manifestUrl, err)
	}
	return k.ApplyWithPrompt(manifest, description)
}

func (k *KubeClient) DeleteFromFile(yamlFilepath string) {
	manifestBytes, err := os.ReadFile(yamlFilepath)
	if err != nil {
		makeup.ExitDueToFatalError(err, "failed to read manifest file "+yamlFilepath)
	}
	k.DeleteFromManifest(string(manifestBytes))
}

func (k *KubeClient) DeleteFromManifest(manifest string) {
	obj, err := yaml.Parse(string(manifest))
	if err != nil {
		makeup.ExitDueToFatalError(err, "failed to parse manifest:\n"+string(manifest))
	}

	// Use the unified delete wrapper
	if _, err := k.Delete(obj.GetKind(), obj.GetName(), obj.GetNamespace(), "", false); err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("failed to delete resource %s/%s in namespace %s (manifest: %s):\n ", obj.GetKind(), obj.GetName(), obj.GetNamespace(), manifest))
	}
}

func (k *KubeClient) Exec(serviceInstanceName, namespace, RemoteUploadContainerName, command string, args ...string) (string, error) {
	opts := KubectlOpts{
		Command:   "exec",
		Kind:      serviceInstanceName,
		Namespace: namespace,
		AdditionalArgs: append([]string{
			"-c",
			RemoteUploadContainerName,
			"--",
			command},
			args...),
	}
	_, output, err := runKubeCtlCommand(opts.withContextFrom(k))
	return string(output), err
}

func Contexts(filter string) ([]string, error) {
	opts := KubectlOpts{
		Command:      "config",
		Kind:         "get-contexts",
		OutputFormat: "name",
	}
	_, out, err := runKubeCtlCommand(opts)
	if err != nil {
		makeup.PrintFail("Failed to list contexts:\n" + string(out))
		return nil, err
	}
	return trimAndFilter(out, filter), nil
}
func Clusters(filter string) ([]string, error) {
	opts := KubectlOpts{
		Command: "config",
		Kind:    "get-clusters",
	}
	_, out, err := runKubeCtlCommand(opts)
	if err != nil {
		makeup.PrintFail("Failed to list clusters:\n" + string(out))
		return nil, err
	}
	return trimAndFilter(out, filter), nil
}

func trimAndFilter(input []byte, filter string) []string {
	lines := strings.Split(strings.TrimSpace(string(input)), "\n")
	var res []string
	for _, l := range lines {
		if s := strings.TrimSpace(l); s != "" &&
			strings.Contains(strings.ToLower(s), filter) {
			res = append(res, s)
		}
	}
	return res
}

func ClusterInfo(ctx context.Context, kubeContext string) (string, error) {
	opts := KubectlOpts{
		Command: "cluster-info",
		Context: kubeContext,
	}
	_, out, err := runKubeCtlCommand(opts)
	return string(out), err
}

func SwitchContext(target string) ([]byte, error) {
	opts := KubectlOpts{
		Command: "config",
		Kind:    "use-context",
		Name:    target,
	}
	_, out, err := runKubeCtlCommand(opts)
	return out, err
}

func CurrentContext() (string, error) {
	opts := KubectlOpts{
		Command: "config",
		Kind:    "current-context",
	}
	_, out, err := runKubeCtlCommand(opts)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (k *KubeClient) RolloutStatus(resourceType, name, namespace, timeout string) (string, error) {
	if resourceType != "" {
		name = resourceType + "/" + name
	}
	opts := KubectlOpts{
		Command:   "rollout",
		Kind:      "status",
		Name:      name,
		Namespace: namespace,
		Timeout:   timeout,
	}

	_, out, err := runKubeCtlCommand(opts.withContextFrom(k))
	return string(out), err
}

func (k *KubeClient) ApiResources(apiGroup, kubeconfigPath, kubeContext, format string) (string, error) {
	if apiGroup != "" {
		apiGroup = "--api-group=" + apiGroup
	}
	opts := KubectlOpts{
		Command:        "api-resources",
		Kind:           apiGroup,
		OutputFormat:   format,
		KubeConfigPath: kubeconfigPath,
		Context:        kubeContext,
	}

	_, out, err := runKubeCtlCommand(opts.withContextFrom(k))
	return string(out), err
}

func Version(client bool, requestTimeout string) (string, error) {
	opts := KubectlOpts{
		Command: "version",
	}
	if client {
		opts.AdditionalArgs = append(opts.AdditionalArgs, "--client")
	}
	if requestTimeout != "" {
		opts.AdditionalArgs = append(opts.AdditionalArgs, "--request-timeout="+requestTimeout)
	}

	_, out, err := runKubeCtlCommand(opts)
	return string(out), err
}

func (k *KubeClient) Run(namespace, name, image, labels, command string, args ...string) (string, error) {
	if !strings.HasPrefix(image, "--image=") {
		image = "--image=" + image
	}
	opts := KubectlOpts{
		Command:        "run",
		Kind:           name,
		Name:           image,
		Namespace:      namespace,
		Selector:       labels,
		AdditionalArgs: []string{},
	}

	opts.AdditionalArgs = append(opts.AdditionalArgs, "--", command)
	opts.AdditionalArgs = append(opts.AdditionalArgs, args...)

	_, out, err := runKubeCtlCommand(opts.withContextFrom(k))
	return string(out), err
}

type KubectlOpts struct {
	Command        string
	Kind           string
	Name           string
	Namespace      string
	OutputFormat   string
	StdIn          []byte
	Filename       string
	Selector       string
	Context        string
	KubeConfigPath string
	Timeout        string
	AdditionalArgs []string
	IgnoreNotFound bool
}

func (opts KubectlOpts) withContextFrom(k *KubeClient) KubectlOpts {
	if opts.Context == "" && k.KubeContext != "" {
		opts.Context = k.KubeContext
	}
	return opts
}

func runKubeCtlCommand(opts KubectlOpts) (string, []byte, error) {
	args := []string{
		opts.Command,
	}
	if opts.Kind == "" {
		if opts.Filename == "" && !slices.Contains(commandsWithoutSubcommands, opts.Command) {
			return "kubectl " + args[0], nil, fmt.Errorf("resource kind may not be empty")
		}
	} else {
		args = append(args, opts.Kind)
	}

	args = addIfNotEmpty(args, "", opts.Name)
	args = addIfNotEmpty(args, "-n", opts.Namespace)
	args = addIfNotEmpty(args, "-o", opts.OutputFormat)
	args = addIfNotEmpty(args, "-f", opts.Filename)
	args = addIfNotEmpty(args, "--context", opts.Context)
	args = addIfNotEmpty(args, "-l", opts.Selector)
	args = addIfNotEmpty(args, "--kubeconfig", opts.KubeConfigPath)
	args = addIfNotEmpty(args, "--timeout=", opts.Timeout)

	if opts.IgnoreNotFound {
		args = append(args, "--ignore-not-found")
	}

	if opts.AdditionalArgs != nil {
		args = append(args, opts.AdditionalArgs...)
	}

	output, err := makeup.Command("kubectl", args...).Stdin(opts.StdIn).Quiet().Run()

	if makeup.Verbose || err != nil {
		fmt.Println(string(output))
	}

	return "kubectl " + strings.Join(args, " "), output, err
}

func addIfNotEmpty(args []string, prefix, value string) []string {
	if value == "" {
		return args
	}
	if prefix != "" && !strings.HasSuffix(prefix, "=") {
		return append(args, prefix, value)
	}
	if strings.HasSuffix(prefix, "=") && !strings.HasPrefix(value, prefix) {
		value = prefix + value
	}

	return append(args, value)
}
