package main

/*
Next: Resolve error:

error: accumulating resources: accumulation err='accumulating resources from
'../backup-config': '/Users/jfischer/Dropbox/workspace/a8s-pg/demo/deploy/a8s/backup-config'
must resolve to a file': recursed accumulation of path '/Users/jfischer/Dropbox/workspace/a8s-pg/demo/deploy/a8s/backup-config': loading KV pairs: file sources: [./backup-store-config.yaml]: evalsymlink failure on '/Users/jfischer/Dropbox/workspace/a8s-pg/demo/deploy/a8s/backup-config/backup-store-config.yaml' : lstat /Users/jfischer/Dropbox/workspace/a8s-pg/demo/deploy/a8s/backup-config/backup-store-config.yaml: no such file or directory


TODO:
- Create S3 bucket with configs
- waitForA8sToBecomeReady

*/

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	//TODO Use this instead: https://github.com/charmbracelet/lipgloss

	"github.com/fatih/color"
	"github.com/sethvargo/go-password/password"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	WorkingDir string `yaml:"WorkingDir"`
}

// Settings
// TODO make configurable / cli param
const kind_demo_cluster_name = "a8s-demo"
const configFileName = ".a8s"
const demoGitRepo = "git@github.com:anynines/a8s-deployment.git"
const demoNamespace = "default"
const certManagerNamespace = "cert-manager"
const certManagerManifestUrl = "https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml"
const default_waiting_time_in_s = 10

var configFilePath string
var cfg Config

func isCommandAvailable(name string) bool {
	//	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	cmd := exec.Command("command", "-v", name)
	if err := cmd.Run(); err != nil {
		color.Red("Couldn't find " + name + " command.")
		return false
	}

	color.Green("Found " + name + ".")
	return true
}

func checkIfDockerIsRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	if err != nil {
		color.Red("Docker is not running")
		return false
	}
	color.Green("Docker is running")
	return true
}

func checkIfKindClusterExists() bool {
	cmd := exec.Command("kind", "get", "clusters")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		color.Red("Couldn't capture output of 'kind get clusters' command.")
		log.Fatal(err)
		return false
	}

	strOutput := string(output)

	fmt.Println("\nKind clusters:")
	fmt.Println(strOutput)

	if strings.Contains(strOutput, kind_demo_cluster_name) {
		color.Green("There is a suitable Kind cluster with the name " + kind_demo_cluster_name + " running.")
		return true
	}

	// Check if the output contains the string "No kind clusters found."
	if strings.Contains(strOutput, "No kind clusters found.") {
		color.Red("There are no kind clusters. A cluster with the name: " + kind_demo_cluster_name + " is needed.")
		return false
	}

	color.Red("There appears to be kind clusters but none with the name: " + kind_demo_cluster_name + ".")
	return false
}

func checkPrerequisites() {
	allGood := true

	color.Blue("Checking Prerequisites...")

	if !isCommandAvailable("docker") {
		allGood = false
	}

	if !checkIfDockerIsRunning() {
		allGood = false
	}

	if !isCommandAvailable("kind") {
		allGood = false
	}

	if !isCommandAvailable("cmctl") {
		color.Red("The cert-manager CLI isn't installed. Please visit: https://cert-manager.io/docs/reference/cmctl/#installation")
		allGood = false
	}

	if !checkIfKindClusterExists() {
		createKindCluster()

		fmt.Println()
		color.Blue("Rerunning prerequisite check ...")
		checkPrerequisites()
	}

	checkSelectedCluster()

	if !allGood {
		color.Red("Sadly, mandatory prerequisited haven't been met. Aborting...")
		os.Exit(1)
	} else {
		color.Green("Is all good man! Let's proceed...")
	}
}

func printWelcomeScreen() {
	fmt.Println("Welcome to the a8s Demos")
	color.Blue("Currently the a8s PostgreSQL or short a8s-pg demo is selected.")

}

func establishConfigFilePath() {
	color.Blue("Setting a config file path in order to persist settings...")

	homeDir, err := os.UserHomeDir()

	if err != nil {
		exitDueToFatalError(err, "Couldn't obtain your home directory. Aborting...")

	}

	configFilePath = filepath.Join(homeDir, configFileName)

	color.Blue("Settings will be stored at " + configFilePath)

}

func establishWorkingDir() {
	color.Magenta("We will need a working directory for the demo. Let's find one..")

	reader := bufio.NewReader(os.Stdin)

	for {
		cwd, err := os.Getwd()

		if err != nil {
			exitDueToFatalError(err, "Couldn't obtain your current working directory.")
		}

		fmt.Println("The current working directory is: ", cwd)
		fmt.Print("Can we the current directory as a working directory? (y/n): ")
		choice, _ := reader.ReadString('\n')

		if strings.HasPrefix(choice, "y") {
			fmt.Println("Yes")
			cfg.WorkingDir = cwd
			break
		} else if strings.HasPrefix(choice, "n") {
			cfg.WorkingDir = promptPath()
			saveConfig()
			break
		} else {
			fmt.Println("Invalid choice. Please try again.")
		}
	}

	saveConfig()
}

// https://dev.to/sagartrimukhe/generate-yaml-files-in-golang-29h1
func saveConfig() {
	yamlData, err := yaml.Marshal(&cfg)

	if err != nil {
		exitDueToFatalError(err, "Couldn't save config file. Aborting...")
	}

	err = os.WriteFile(configFilePath, yamlData, 0600)

	if err != nil {
		exitDueToFatalError(err, "Couldn't save config file. Aborting...")
	}
}

func loadConfig() bool {
	yamlFile, err := os.ReadFile(configFilePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			color.Blue("No config file found.")
			return false
		}

		exitDueToFatalError(err, "Couldn't open config file.")
	}

	err = yaml.Unmarshal(yamlFile, &cfg)

	if err != nil {
		color.Red("Coudln't parse config file.")
	}

	color.Blue("Using the following working directory: " + cfg.WorkingDir)

	return true
}

func exitDueToFatalError(err error, msg string) {
	color.Red(msg)
	fmt.Print(err)
	os.Exit(1)
}

func promptPath() string {
	reader := bufio.NewReader(os.Stdin)

	for {

		color.Yellow("No, ok. Then please enter to the working directory of your choice.")

		fmt.Print("Enter a path: ")

		// Create a new scanner to read user input
		scanner := bufio.NewScanner(os.Stdin)

		scanner.Scan()
		scanner.Err()

		// Retrieve the entered path
		path := scanner.Text()

		fmt.Print("Awesome. We got " + path + " as a working directory. Is this ok? (y/n)")
		choice, _ := reader.ReadString('\n')

		if strings.HasPrefix(choice, "y") {
			return path
		} else if strings.HasPrefix(choice, "n") {
			fmt.Println("Ok, no problem. Please try again.")
		} else {
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func checkoutGitRepository(repositoryURL, localDirectory string) error {
	// Check if the local directory already exists
	if _, err := os.Stat(localDirectory); !os.IsNotExist(err) {
		return fmt.Errorf("local directory already exists")
	}

	// Run the git clone command to checkout the repository
	cmd := exec.Command("git", "clone", repositoryURL, localDirectory)

	color.Blue("Executing: " + cmd.String())

	output, err := cmd.CombinedOutput()

	if err != nil {
		color.Red("Failed to checkout the git repository: %v", err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return err
	} else {
		fmt.Println(string(output))
		return nil
	}
}

func checkoutDeploymentGitRepository() {
	color.Blue("Checking out git repository with demo manifests...")
	checkoutGitRepository(demoGitRepo, cfg.WorkingDir)
}

func createKindCluster() {
	color.Blue("Let's create a Kubernetes cluster named " + kind_demo_cluster_name + " using Kind...")

	// kind create cluster --name a8s-ds --config kind-cluster-3nodes.yaml
	cmd := exec.Command("kind", "create", "cluster", "--name", kind_demo_cluster_name)

	color.Blue("Executing: " + cmd.String())

	output, err := cmd.CombinedOutput()

	if err != nil {
		color.Red("Failed to execute the command: %v", err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return
	} else {
		fmt.Println(string(output))
		return
	}
}

func checkSelectedCluster() {
	color.Blue("Checking whether the " + kind_demo_cluster_name + " cluster is selected...")
	cmd := exec.Command("kubectl", "config", "current-context")

	output, err := cmd.CombinedOutput()

	if err != nil {
		exitDueToFatalError(err, "Can't retrieve the currently selected cluster using the command: "+cmd.String())
	}

	current_context := strings.TrimSpace(string(output))

	color.Blue("The currently selected Kubernetes context is: " + current_context)

	desired_context_name := "kind-" + kind_demo_cluster_name

	if strings.HasPrefix(current_context, desired_context_name) {
		color.Green("It seems that the right context is selected: " + desired_context_name)
	} else {
		color.Red("The expected context is " + desired_context_name + " but the current context is: " + current_context + ". Please select the desired context! Try executing: ")
		fmt.Println("kubectl config use-context " + desired_context_name)
		os.Exit(1)
	}
}

func getKubernetesConfigPath() string {
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		color.Blue("Kubernetes configuration is set by the $KUBECONFIG env variable.")
	} else if home := homedir.HomeDir(); home != "" {
		color.Blue("Kubernetes configuration is set by $HOME/.kube/config.")
		kubeconfig = *flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		color.Blue("Kubernetes configuration is set by config flag.")
		kubeconfig = *flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// Set the bool variable based on the flags passed in by the user
	flag.Parse()

	return kubeconfig
}

func countPodsInDemoNamespace() int {
	return countPodsInNamespace(demoNamespace)
}

func getKubernetesClientSet() *kubernetes.Clientset {
	kubeconfig := getKubernetesConfigPath()
	color.Blue("Kubernetes config located at: " + kubeconfig)

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/main.go
func countPodsInNamespace(namespace string) int {

	color.Blue("Checking whether there are pods in the cluster...")

	clientset := getKubernetesClientSet()

	//for {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	return len(pods.Items)
}

func kubectlApplyF(yamlFilepath string) {

	cmd := exec.Command("kubectl", "apply", "-f", yamlFilepath)

	output, err := cmd.CombinedOutput()

	color.Blue(cmd.String())

	if err != nil {
		exitDueToFatalError(err, "Can't kubectl apply with command: "+cmd.String())
	}

	fmt.Println(string(output))
}

func kubectlApplyKustomize(kustomizeFilepath string) {

	cmd := exec.Command("kubectl", "apply", "--kustomize", kustomizeFilepath)

	output, err := cmd.CombinedOutput()

	color.Blue(cmd.String())

	fmt.Println(string(output))

	if err != nil {
		exitDueToFatalError(err, "Can't kubectl kustomize with using the command: "+cmd.String())
	}

}

func applyA8sManifests() {
	color.Magenta("Applying the a8s Data Service manifests...")
	kustomizePath := filepath.Join(cfg.WorkingDir, "deploy", "a8s", "manifests")
	kubectlApplyKustomize(kustomizePath)
	color.Magenta("Done applying a8s manifests.")
}

func waitForCertManagerToBecomeReady() {
	color.Blue("Waiting for the cert-manager API to become ready.")
	cmd := exec.Command("cmctl", "check", "api")

	for {
		output, err := cmd.CombinedOutput()

		color.Blue(cmd.String())

		if err != nil {
			exitDueToFatalError(err, "Can't verify the cert-manager's API: "+cmd.String())
		}

		strOutput := string(output)

		fmt.Println(strOutput)

		if strings.TrimSpace(strOutput) == "The cert-manager API is ready" {
			color.Green("The cert-manager is ready")
			break
		} else {
			color.Yellow("Continuing to wait for the cert-manager API...")
		}

		time.Sleep(30 * time.Second)
	}

	// var namespace = flag.String("namespace", certManagerNamespace)
	// var selector = flag.String("selector", "app=cert-manager", "pod selector")
	// var timeout = flag.Int("timeout", default_waiting_time_in_s, "timeout in seconds")
	// flag.Parse()

	// kubeconfig := getKubernetesConfigPath()

	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	panic(err)
	// }
	// clientSet, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err)
	// }

	// // Block up to timeout seconds for listed pods in namespace/selector to enter running state
	// err = k8sutils.WaitForPodBySelectorRunning(clientSet, *namespace, *selector, *timeout)
	// if err != nil {
	// 	log.Errorf("\nThe pod never entered running phase\n")
	// 	os.Exit(1)
	// }
	// fmt.Printf("\nAll pods in namespace=\"%s\" with selector=\"%s\" are running!\n", *namespace, *selector)
}

// for {
// 	count := countPodsInNamespace(certManagerNamespace)

// 	// There should be 3 pods
// 	if count == 3 {
// 		// Check if they are ready

// 		pods, err := clientset.CoreV1().Pods(certManagerNamespace).List(context.TODO(), metav1.ListOptions{})

// 		if err != nil {
// 			panic(err.Error())
// 		}

// 		for _, pod := range pods.Items {
// 			pod.Status.Conditions
// 			fmt.Println(pod.Name + ": " + pod.Status.String())
// 			fmt.Println("\n")
// 		}
// 	}

// 	time.Sleep(default_waiting_time_in_s * time.Second)
// }

func applyCertManagerManifests() {
	count := countPodsInNamespace(certManagerNamespace)

	if count > 0 {
		color.Blue("Found %d pods in the %s namespace", count, certManagerNamespace)
	}

	kubectlApplyF(certManagerManifestUrl)

	waitForCertManagerToBecomeReady()
}

func checkIfFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	} else {
		return false
	}
}

func backupConfigAccessKeyIdFilePath() string {
	return filepath.Join(cfg.WorkingDir, "deploy", "a8s", "backup-config", "access-key-id")
}

func backupConfigSecretAccessKeyFilePath() string {
	return filepath.Join(cfg.WorkingDir, "deploy", "a8s", "backup-config", "secret-access-key")
}

func backupConfigEncryptionPasswordFilePath() string {
	return filepath.Join(cfg.WorkingDir, "deploy", "a8s", "backup-config", "encryption-password")
}

/*
		Generates an encryption password file for backups if it doesnt exist.
	  Does nothing if the file already exists.
*/
func establishEncryptionPasswordFile() {
	color.Blue("Checking if encryption password file for backups already exists...")

	filePath := backupConfigEncryptionPasswordFilePath()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		color.Magenta("There's already an encryption password file. Skipping password generation...")
		return
	}

	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	backupPassword, err := password.Generate(64, 10, 10, false, false)

	if err != nil {
		exitDueToFatalError(err, "Couldn't generate encryption password for backup config.")
	}

	// Store password in file
	f, err := os.Create(filePath)

	if err != nil {
		exitDueToFatalError(err, "Couldn't create file to store  encryption password for backup config to filepath: "+filePath)
	}

	defer f.Close()

	f.WriteString(backupPassword)

	if err != nil {
		exitDueToFatalError(err, "Couldn't write password to file to store encryption password for backup config to filepath: "+filePath)
	}

	f.Sync()
}

func establishBackupStoreCredentials() {
	establishEncryptionPasswordFile()
	// accessKeyId
	// secretAccessKey
}

func main() {
	printWelcomeScreen()

	establishConfigFilePath()

	if !loadConfig() {
		establishWorkingDir()
	}

	checkPrerequisites()

	checkoutDeploymentGitRepository()

	if countPodsInDemoNamespace() == 0 {
		color.Green("Kubernetes cluster has no pods in " + demoNamespace + " namespace.")
	}

	establishBackupStoreCredentials()

	applyCertManagerManifests()

	applyA8sManifests()

}
