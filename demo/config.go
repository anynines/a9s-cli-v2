package demo

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/fatih/color"
	"github.com/sethvargo/go-password/password"
	"gopkg.in/yaml.v2"
)

/*
TODO Correct misleading naming. This function does not establish a config file path
but sets a package variable. This appears to be a left-over from earlier versions
and needs refactoring.
*/
func EstablishConfigFilePath() {
	makeup.PrintVerbose("Setting a config file path in order to persist settings...")

	homeDir, err := os.UserHomeDir()

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't obtain your home directory. Aborting...")

	}

	configFilePath = filepath.Join(homeDir, configFileName)

	makeup.PrintVerbose("Settings will be stored at " + configFilePath)

}

/*
Proposes and creates a meaningful working directory on first time use.
TODO Reduce code complexity in this function.
*/
func EstablishWorkingDir() {
	makeup.PrintH1("Setting up a Working Directory")
	makeup.PrintVerbose("We will need a working directory for the demo. Let's find one..")

	reader := bufio.NewReader(os.Stdin)

	for {
		homeDir, err := os.UserHomeDir()

		if err != nil {
			makeup.ExitDueToFatalError(err, "Couldn't obtain your home directory.")
		}

		cwd := filepath.Join(homeDir, defaultWorkDir)

		fmt.Println("The current working directory is: ", cwd)
		fmt.Print("Do you want to use this directory as a working directory? (y/n): ")
		choice, _ := reader.ReadString('\n')

		if strings.HasPrefix(choice, "y") {
			fmt.Println("Yes")
			DemoConfig.WorkingDir = cwd
			break
		} else if strings.HasPrefix(choice, "n") {
			DemoConfig.WorkingDir = promptPath()
			saveConfig()
			break
		} else {
			fmt.Println("Invalid choice. Please try again.")
		}
	}

	saveConfig()
}

/*
Execute this at the beginning of every command that requires a config to be present.
*/
func EnsureConfigIsLoaded() {
	EstablishConfigFilePath()

	if !LoadConfig() {
		makeup.ExitDueToFatalError(nil, "There is no config, yet. Please create a demo environment before attempting to create a service instance.")
	}
}

func EstablishConfig() {
	EstablishConfigFilePath()

	if !LoadConfig() {
		EstablishWorkingDir()
	}
}

// https://dev.to/sagartrimukhe/generate-yaml-files-in-golang-29h1
func saveConfig() {

	//TODO Make configurable / prompt from user
	if DemoConfig.DemoSpace == "" {
		DemoConfig.DemoSpace = defaultDemoSpace
	}

	yamlData, err := yaml.Marshal(&DemoConfig)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't save config file. Aborting...")
	}

	err = os.WriteFile(configFilePath, yamlData, 0600)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't save config file. Aborting...")
	}
}

func LoadConfig() bool {
	yamlFile, err := os.ReadFile(configFilePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			makeup.PrintVerbose("No config file found.")
			return false
		}

		makeup.ExitDueToFatalError(err, "Couldn't open config file.")
	}

	err = yaml.Unmarshal(yamlFile, &DemoConfig)

	if err != nil {
		makeup.PrintFail("Coudln't parse config file.")
	}

	makeup.PrintVerbose("Using the following working directory: " + DemoConfig.WorkingDir)

	return true
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

		// As a safety measure, do not write a relative path to the config file.
		// Otherwise, the CLI execution depends on being in the correct working dir.
		if !filepath.IsAbs(path) {
			currentPath, err := os.Getwd()
			if err != nil {
				makeup.ExitDueToFatalError(err, "Could not get the execution directory.")
			}

			path = filepath.Join(currentPath, path)
		}

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

/*
Generates an encryption password file for backups if it doesnt exist.
Does nothing if the file already exists.

TODO Make this an optional parameter so that users can set this value
*/
func establishEncryptionPasswordFile() {
	makeup.PrintVerbose("In order to encrypt backups we need an encryption password.")
	makeup.Print("Checking if encryption password file for backups already exists...")

	filePath := BackupConfigEncryptionPasswordFilePath()

	_, err := os.Stat(filePath)
	if err == nil {
		makeup.Print("There's already an encryption password file. Skipping password generation...")
		return
	}

	if !os.IsNotExist(err) {
		makeup.ExitDueToFatalError(err, "Could not check the encryption password file.")
	}

	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	backupPassword, err := password.Generate(64, 10, 10, false, false)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't generate encryption password for backup config.")
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
		makeup.ExitDueToFatalError(err, "Couldn't create file to store  encryption password for backup config to filepath: "+filePath)
	}

	defer f.Close()

	f.WriteString(content)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't write password to file to store encryption password for backup config to filepath: "+filePath)
	}

	f.Sync()
}

func CheckIfFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	} else {
		return false
	}
}

func A8sDeploymentLocalPath() string {
	return filepath.Join(DemoConfig.WorkingDir, demoA8sDeploymentLocalDir)
}

func A8sDeploymentExamplesPath() string {
	return filepath.Join(A8sDeploymentLocalPath(), "examples")
}

func UserManifestsPath() string {
	fp := filepath.Join(DemoConfig.WorkingDir, "usermanifests")

	err := os.MkdirAll(fp, os.ModePerm)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't create user manifests folder at: "+fp)
	}

	return fp
}

func BackupConfigBasePath() string {
	return filepath.Join(A8sDeploymentLocalPath(), "deploy", "a8s", "backup-config")
}

func BackupConfigAccessKeyIdFilePath() string {
	return filepath.Join(BackupConfigBasePath(), "access-key-id")
}

func BackupConfigSecretAccessKeyFilePath() string {
	return filepath.Join(BackupConfigBasePath(), "secret-access-key")
}

func BackupConfigEncryptionPasswordFilePath() string {
	return filepath.Join(BackupConfigBasePath(), "encryption-password")
}

/*
TODO A function called read... should not write a file > Should be separate functions
Checks if there's a file.
If not it prompts to read the file content from STDIN.
Skips if the file is already present
*/
func ReadStringFromFileOrConsole(filePath, contentType string, showContent bool) string {

	//TODO This is not a good way to handle the existence of the file.
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		makeup.Print("There's already an " + contentType + " file...")
		return ""
	}

	// Enter access key id as the access-key-id-file doesnt exist, yet.
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your " + contentType + ": ")

	content, err := reader.ReadString('\n')

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't read  "+contentType+"  from STDIN.")
	}

	if showContent {
		makeup.Print(contentType + " : " + content)
	}

	return content
}

/*
Checks if there's an access key id file.
If not it prompts to read the access key id from STDIN.
Skips if the access key id file is already present
*/
func establishAccessKeyId() {
	makeup.PrintH2("In order to store backups on an object store such as S3, we need an ACCESS KEY ID.")

	filePath := BackupConfigAccessKeyIdFilePath()

	if BackupStoreAccessKey == "" {
		makeup.ExitDueToFatalError(nil, "The backup-store-accesskey can't be empty!")
	}

	saveStringToFile(filePath, BackupStoreAccessKey)
}

func establishSecretAccessKey() {
	makeup.PrintH2("In order to store backups on an object store such as S3, we need a SECRET KEY.")

	filePath := BackupConfigSecretAccessKeyFilePath()

	if BackupStoreSecretKey == "" {
		makeup.ExitDueToFatalError(nil, "The backup-store-secretkey can't be empty!")
	}

	saveStringToFile(filePath, BackupStoreSecretKey)
}

func backupStoreConfigFilePath() string {
	return filepath.Join(BackupConfigBasePath(), "backup-store-config.yaml")
}

func establishBackupStoreConfigYaml() {
	makeup.PrintH2("Checking the backup-store-config.yaml file...")

	filePath := backupStoreConfigFilePath()

	if CheckIfFileExists(filePath) {
		makeup.PrintCheckmark(fmt.Sprintf("There's already a backup-store-config.yaml file at %s. Old configuration data will be overwritten if there were changes to them.", filePath))
	} else {
		makeup.Print("Writing a backup-store-config.yaml with configurations to " + filePath)
	}

	var actualProvider string

	// For minio the backup_agent will be configured using an S3 compatible storage client
	if strings.ToLower(BackupInfrastructureProvider) == "minio" {
		actualProvider = "AWS"

		// The endpoint is not set as a default cmd param as with AWS as a provider, this endpoint would cause problems
		BackupInfrastructureEndpoint = "http://minio.minio-dev.svc.cluster.local:9000"
		BackupInfrastructurePathStyle = true
	} else {
		actualProvider = BackupInfrastructureProvider
	}

	// TODO Make backup store configurable
	blobStoreConfig := BlobStore{
		Config: BlobStoreConfig{
			CloudConfig: BlobStoreCloudConfiguration{
				Provider:  actualProvider,
				Container: BackupInfrastructureBucket,
				Region:    BackupInfrastructureRegion,
				Endpoint:  BackupInfrastructureEndpoint,
				PathStyle: BackupInfrastructurePathStyle,
			},
		},
	}

	//TODO Refactor using WriteYAMLToFile
	yamlData, err := yaml.Marshal(&blobStoreConfig)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't generate backup-store-config.yaml file. Aborting...")
	}

	err = os.WriteFile(filePath, yamlData, 0644)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't save backup-store-config.yaml file. Aborting...")
	}

}

func GetConfig() Config {
	return DemoConfig
}

func EstablishBackupStoreCredentials() {
	establishEncryptionPasswordFile()
	establishAccessKeyId()
	establishSecretAccessKey()
	establishBackupStoreConfigYaml()

	//TODO deploy/a8s/backup-config/backup-store-config.yaml.template
}

/*
Writes the provided YAML string to a YAML file at the given path.
*/
func WriteYAMLToFile(instanceYAML string, manifestPath string) {

	err := os.WriteFile(manifestPath, []byte(instanceYAML), 0600)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't save YAML file at: "+manifestPath)
	}

	makeup.PrintInfo("The YAML manifest is located at: " + manifestPath)

	makeup.Print("The YAML manifest contains: ")
	err = makeup.PrintYAMLFile(manifestPath)

	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't read manifest from "+manifestPath)
	}
}

/*
Returns a filepath located in the user manifests path.
*/
func GetUserManifestPath(filename string) string {
	manifestsPath := UserManifestsPath()

	manifestPath := filepath.Join(manifestsPath, filename)

	return manifestPath
}
