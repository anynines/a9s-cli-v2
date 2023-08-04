package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	//TODO Use this instead: https://github.com/charmbracelet/lipgloss
	"github.com/fatih/color"
	"sigs.k8s.io/yaml"
)

type Config struct {
	WorkingDir string `yaml:"WorkingDir"`
}

// Settings
// TODO make configurable / cli param
const kind_demo_cluster_name = "a8s-demo"
const configFileName = ".a8s"

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

	color.Red("There appear to be kind clusters but none with the name: " + kind_demo_cluster_name + ".")
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

	if !checkIfKindClusterExists() {
		allGood = false
	}

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

	err = os.WriteFile(configFilePath, yamlData, 0644)

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

func main() {
	printWelcomeScreen()

	establishConfigFilePath()

	if !loadConfig() {
		establishWorkingDir()
	}

	checkPrerequisites()
}
