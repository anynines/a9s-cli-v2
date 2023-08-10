package demo

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
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/sethvargo/go-password/password"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Settings
// TODO make configurable / cli param
const kindDemoClusterName = "a8s-demo"
const configFileName = ".a8s"
const demoGitRepo = "git@github.com:anynines/a8s-deployment.git"
const certManagerNamespace = "cert-manager"
const certManagerManifestUrl = "https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml"
const defaultDemoSpace = "default"

// const default_waiting_time_in_s = 10

type Config struct {
	WorkingDir string `yaml:"WorkingDir"`
	DemoSpace  string `yaml:"DemoSpace"`
}

type BlobStore struct {
	Config BlobStoreConfig `yaml:"config"`
}

type BlobStoreConfig struct {
	CloudConfig BlobStoreCloudConfiguration `yaml:"cloud_configuration"`
}

type BlobStoreCloudConfiguration struct {
	Provider  string `yaml:"provider"`
	Container string `yaml:"container"`
	Region    string `yaml:"region"`
}

var configFilePath string
var cfg Config

func IsCommandAvailable(cmdName string) bool {
	//	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	cmd := exec.Command("command", "-v", cmdName)
	if err := cmd.Run(); err != nil {
		requiredCmds := RequiredCommands()

		msg := "Couldn't find " + cmdName + " command."

		if requiredCmds[cmdName][runtime.GOOS] != "" {
			msg += " Try running: " + requiredCmds[cmdName][runtime.GOOS]
		}

		PrintFail(msg)

		return false
	}

	PrintCheckmark("Found " + cmdName + ".")

	return true
}

func checkIfDockerIsRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	if err != nil {
		PrintFail("Docker is not running")
		return false
	}
	PrintCheckmark("Docker is running")
	return true
}

func checkIfKubernetesIsRunning() bool {
	cmd := exec.Command("kubectl", "api-versions")
	err := cmd.Run()
	if err != nil {
		PrintFail("Kubernetes is not running")
		return false
	}
	PrintCheckmark("Kubernetes is running")
	return true
}

func CheckIfKindClusterExists() bool {
	cmd := exec.Command("kind", "get", "clusters")

	// Capture the command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		PrintFail("Couldn't capture output of 'kind get clusters' command.")
		log.Fatal(err)
		return false
	}

	strOutput := string(output)

	PrintListFromMultilineString("Kind Clusters", strOutput)

	// fmt.Println("\nKind clusters:")
	// fmt.Println(strOutput)

	if strings.Contains(strOutput, kindDemoClusterName) {
		PrintCheckmark("There is a suitable Kind cluster with the name " + kindDemoClusterName + " running.")
		return true
	}

	// Check if the output contains the string "No kind clusters found."
	if strings.Contains(strOutput, "No kind clusters found.") {
		PrintFail("There are no kind clusters. A cluster with the name: " + kindDemoClusterName + " is needed.")
		return false
	}

	PrintFail("There appears to be kind clusters but none with the name: " + kindDemoClusterName + ".")
	return false
}

func CheckCommandAvailability() {

	allGood := true

	requiredCmds := RequiredCommands()

	// cmdDetails
	for cmdName, _ := range requiredCmds {

		if !IsCommandAvailable(cmdName) {
			allGood = false
		}
	}

	if !allGood {
		PrintFailSummary("Sadly, mandatory commands are missing. Aborting...")
		os.Exit(1)
	} else {
		PrintCheckmark("Is all good man! Let's proceed...")
	}
}

func CheckPrerequisites() {
	allGood := true

	PrintH1("Checking Prerequisites...")

	CheckCommandAvailability()

	if !checkIfDockerIsRunning() {
		allGood = false
	}

	if !checkIfKubernetesIsRunning() {
		allGood = false
	}

	if !CheckIfKindClusterExists() {
		CreateKindCluster()

		fmt.Println()
		PrintH2("Rerunning prerequisite check ...")
		CheckPrerequisites()
		allGood = true
	}

	CheckSelectedCluster()

	if !allGood {
		PrintFailSummary("Sadly, mandatory prerequisited haven't been met. Aborting...")
		os.Exit(1)
	} else {
		PrintCheckmark("Is all good man! Let's proceed...")
	}
}
func PrintWelcomeScreen() {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	fmt.Println()

	title := "Welcome to the a8s Data Service demos"

	PrintH2("Currently the a8s PostgreSQL or short a8s-pg demo is selected.")

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f8f8f8")).
		Background(lipgloss.Color("#505d78")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(0).
		Width(physicalWidth - 2).
		Align(lipgloss.Center).
		//AlignVertical(lipgloss.Top).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5a6987")).
		BorderBackground(lipgloss.Color("e4833e"))
	fmt.Println(style.Render(title))
}

func EstablishConfigFilePath() {
	PrintH2("Setting a config file path in order to persist settings...")

	homeDir, err := os.UserHomeDir()

	if err != nil {
		ExitDueToFatalError(err, "Couldn't obtain your home directory. Aborting...")

	}

	configFilePath = filepath.Join(homeDir, configFileName)

	PrintH2("Settings will be stored at " + configFilePath)

}

func EstablishWorkingDir() {
	PrintH1("Setting up a Working Directory")
	PrintH2("We will need a working directory for the demo. Let's find one..")

	reader := bufio.NewReader(os.Stdin)

	for {
		cwd, err := os.Getwd()

		if err != nil {
			ExitDueToFatalError(err, "Couldn't obtain your current working directory.")
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

	//TODO Make configurable / prompt from user
	if cfg.DemoSpace == "" {
		cfg.DemoSpace = defaultDemoSpace
	}

	yamlData, err := yaml.Marshal(&cfg)

	if err != nil {
		ExitDueToFatalError(err, "Couldn't save config file. Aborting...")
	}

	err = os.WriteFile(configFilePath, yamlData, 0600)

	if err != nil {
		ExitDueToFatalError(err, "Couldn't save config file. Aborting...")
	}
}

func LoadConfig() bool {
	yamlFile, err := os.ReadFile(configFilePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			PrintH2("No config file found.")
			return false
		}

		ExitDueToFatalError(err, "Couldn't open config file.")
	}

	err = yaml.Unmarshal(yamlFile, &cfg)

	if err != nil {
		PrintFail("Coudln't parse config file.")
	}

	PrintH2("Using the following working directory: " + cfg.WorkingDir)

	return true
}

func ExitDueToFatalError(err error, msg string) {
	PrintFail(msg)
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

		fmt.Print("Awesome. We got " + path + " as a working directory. Is this ok? (y/n): ")
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

func CheckoutGitRepository(repositoryURL, localDirectory string) error {
	// Check if the local directory already exists
	if _, err := os.Stat(localDirectory); !os.IsNotExist(err) {
		return fmt.Errorf("local directory already exists")
	}

	// Run the git clone command to checkout the repository
	cmd := exec.Command("git", "clone", repositoryURL, localDirectory)

	PrintH2("Executing: " + cmd.String())

	output, err := cmd.CombinedOutput()

	if err != nil {
		PrintFail("Failed to checkout the git repository: " + err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return err
	} else {
		fmt.Println(string(output))
		return nil
	}
}

func CheckoutDeploymentGitRepository() {
	PrintH2("Checking out git repository with demo manifests...")
	CheckoutGitRepository(demoGitRepo, cfg.WorkingDir)
}

func CreateKindCluster() {
	PrintH2("Let's create a Kubernetes cluster named " + kindDemoClusterName + " using Kind...")

	// kind create cluster --name a8s-ds --config kind-cluster-3nodes.yaml
	cmd := exec.Command("kind", "create", "cluster", "--name", kindDemoClusterName)

	PrintH2("Executing: " + cmd.String())

	output, err := cmd.CombinedOutput()

	if err != nil {
		PrintFail("Failed to execute the command: " + err.Error())
		fmt.Println(string(output))
		os.Exit(1)
		return
	} else {
		fmt.Println(string(output))
		return
	}
}

func CheckSelectedCluster() {
	PrintH2("Checking whether the " + kindDemoClusterName + " cluster is selected...")
	cmd := exec.Command("kubectl", "config", "current-context")

	output, err := cmd.CombinedOutput()

	if err != nil {
		ExitDueToFatalError(err, "Can't retrieve the currently selected cluster using the command: "+cmd.String())
	}

	current_context := strings.TrimSpace(string(output))

	PrintH2("The currently selected Kubernetes context is: " + current_context)

	desired_context_name := "kind-" + kindDemoClusterName

	if strings.HasPrefix(current_context, desired_context_name) {
		PrintCheckmark("It seems that the right context is selected: " + desired_context_name)
	} else {
		PrintFail("The expected context is " + desired_context_name + " but the current context is: " + current_context + ". Please select the desired context! Try executing: ")
		fmt.Println("kubectl config use-context " + desired_context_name)
		os.Exit(1)
	}
}

func GetKubernetesConfigPath() string {
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		PrintH2("Kubernetes configuration is set by the $KUBECONFIG env variable.")
	} else if home := homedir.HomeDir(); home != "" {
		PrintH2("Kubernetes configuration is set by $HOME/.kube/config.")
		kubeconfig = *flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		PrintH2("Kubernetes configuration is set by config flag.")
		kubeconfig = *flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// Set the bool variable based on the flags passed in by the user
	flag.Parse()

	return kubeconfig
}

func CountPodsInDemoNamespace() int {
	return countPodsInNamespace(cfg.DemoSpace)
}

func GetKubernetesClientSet() *kubernetes.Clientset {
	kubeconfig := GetKubernetesConfigPath()
	PrintH2("Kubernetes config located at: " + kubeconfig)

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

	PrintH2("Checking whether there are pods in the cluster...")

	clientset := GetKubernetesClientSet()

	//for {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	return len(pods.Items)
}

func KubectlApplyF(yamlFilepath string) {

	cmd := exec.Command("kubectl", "apply", "-f", yamlFilepath)

	output, err := cmd.CombinedOutput()

	PrintH2(cmd.String())

	if err != nil {
		ExitDueToFatalError(err, "Can't kubectl apply with command: "+cmd.String())
	}

	fmt.Println(string(output))
}

func KubectlApplyKustomize(kustomizeFilepath string) {

	cmd := exec.Command("kubectl", "apply", "--kustomize", kustomizeFilepath)

	output, err := cmd.CombinedOutput()

	PrintH2(cmd.String())

	fmt.Println(string(output))

	if err != nil {
		ExitDueToFatalError(err, "Can't kubectl kustomize with using the command: "+cmd.String())
	}

}

func ApplyA8sManifests() {
	PrintH1("Applying the a8s Data Service manifests...")
	kustomizePath := filepath.Join(cfg.WorkingDir, "deploy", "a8s", "manifests")
	KubectlApplyKustomize(kustomizePath)
	PrintCheckmark("Done applying a8s manifests.")
}

func WaitForCertManagerToBecomeReady() {
	PrintH2("Waiting for the cert-manager API to become ready.")
	cmd := exec.Command("cmctl", "check", "api")

	for {
		output, err := cmd.CombinedOutput()

		PrintH2(cmd.String())

		//TODO Crash loop detection / timeout
		if err != nil {
			PrintFail("Continuing to wait for the cert-manager API...")
		}

		strOutput := string(output)

		fmt.Println(strOutput)

		if strings.TrimSpace(strOutput) == "The cert-manager API is ready" {
			PrintCheckmark("The cert-manager is ready")
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

func ApplyCertManagerManifests() {
	count := countPodsInNamespace(certManagerNamespace)

	if count > 0 {
		PrintH2(fmt.Sprintf("Found %d pods in the %s namespace", count, certManagerNamespace))
	}

	KubectlApplyF(certManagerManifestUrl)

	WaitForCertManagerToBecomeReady()
}

func CheckIfFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	} else {
		return false
	}
}

func BackupConfigAccessKeyIdFilePath() string {
	return filepath.Join(cfg.WorkingDir, "deploy", "a8s", "backup-config", "access-key-id")
}

func BackupConfigSecretAccessKeyFilePath() string {
	return filepath.Join(cfg.WorkingDir, "deploy", "a8s", "backup-config", "secret-access-key")
}

func BackupConfigEncryptionPasswordFilePath() string {
	return filepath.Join(cfg.WorkingDir, "deploy", "a8s", "backup-config", "encryption-password")
}

/*
Generates an encryption password file for backups if it doesnt exist.
Does nothing if the file already exists.
*/
func EstablishEncryptionPasswordFile() {
	PrintH2("In order to encrypt backups we need an encryption password.")
	PrintH2("Checking if encryption password file for backups already exists...")

	filePath := BackupConfigEncryptionPasswordFilePath()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		PrintH2("There's already an encryption password file. Skipping password generation...")
		return
	}

	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	backupPassword, err := password.Generate(64, 10, 10, false, false)

	if err != nil {
		ExitDueToFatalError(err, "Couldn't generate encryption password for backup config.")
	}

	saveStringToFile(filePath, backupPassword)
}

/*
Writes content to a file.
Doesn't check if file exists.
Replaces its content if it does exist.
*/
func saveStringToFile(filePath, content string) {
	// Store password in file
	f, err := os.Create(filePath)

	if err != nil {
		ExitDueToFatalError(err, "Couldn't create file to store  encryption password for backup config to filepath: "+filePath)
	}

	defer f.Close()

	f.WriteString(content)

	if err != nil {
		ExitDueToFatalError(err, "Couldn't write password to file to store encryption password for backup config to filepath: "+filePath)
	}

	f.Sync()
}

/*
Checks if there's a file.
If not it prompts to read the file content from STDIN.
Skips if the file is already present
*/
func ReadStringFromFileOrConsole(filePath, contentType string, showContent bool) {

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		PrintH2("There's already an " + contentType + " file...")
		return
	}

	// Enter access key id as the access-key-id-file doesnt exist, yet.
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your " + contentType + ": ")

	accessKeyId, err := reader.ReadString('\n')

	if err != nil {
		ExitDueToFatalError(err, "Can't read  "+contentType+"  from STDIN.")
	}

	if showContent {
		PrintH2(contentType + " : " + accessKeyId)
	}

	// Write file
	saveStringToFile(filePath, accessKeyId)
}

/*
Checks if there's an access key id file.
If not it prompts to read the access key id from STDIN.
Skips if the access key id file is already present
*/
func EstablishAccessKeyId() {
	PrintH2("In order to store backups on an object store such as S3, we need an ACCESS KEY ID.")

	filePath := BackupConfigAccessKeyIdFilePath()

	ReadStringFromFileOrConsole(filePath, "ACCESS KEY ID", true)
}

func establishSecretAccessKey() {
	PrintH2("In order to store backups on an object store such as S3, we need a SECRET KEY.")

	filePath := BackupConfigSecretAccessKeyFilePath()

	ReadStringFromFileOrConsole(filePath, "SECRET KEY", false)
}

func backupStoreConfigFilePath() string {
	return filepath.Join(cfg.WorkingDir, "deploy", "a8s", "backup-config", "backup-store-config.yaml")
}

func establishBackupStoreConfigYaml() {
	PrintH2("Checking the backup-store-config.yaml file...")

	filePath := backupStoreConfigFilePath()

	if CheckIfFileExists(filePath) {
		PrintCheckmark(fmt.Sprintf("There's already a backup-store-config.yaml file at %s. Trusting that the file is ok.", filePath))
	} else {
		PrintH2("Writing a backup-store-config.yaml with defaults to " + filePath)
		blobStoreConfig := BlobStore{
			Config: BlobStoreConfig{
				CloudConfig: BlobStoreCloudConfiguration{
					Provider:  "AWS",
					Container: "a8s-backups",
					Region:    "eu-central-0",
				},
			},
		}

		yamlData, err := yaml.Marshal(&blobStoreConfig)

		if err != nil {
			ExitDueToFatalError(err, "Couldn't generate backup-store-config.yaml file. Aborting...")
		}

		err = os.WriteFile(filePath, yamlData, 0644)

		if err != nil {
			ExitDueToFatalError(err, "Couldn't save backup-store-config.yaml file. Aborting...")
		}
	}
}

func GetConfig() Config {
	return cfg
}

func EstablishBackupStoreCredentials() {
	EstablishEncryptionPasswordFile()
	EstablishAccessKeyId()
	establishSecretAccessKey()

	establishBackupStoreConfigYaml()

	//TODO deploy/a8s/backup-config/backup-store-config.yaml.template
}
