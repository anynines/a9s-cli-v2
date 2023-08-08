package main

/*
Next:

Ask for details for backup store config instead of using defaults.


TODO:

- Create S3 bucket with configs
- waitForA8sToBecomeReady

*/

import (
	"fmt"
	"os"

	//TODO Use this instead: https://github.com/charmbracelet/lipgloss

	"github.com/fatih/color"
	"golang.org/x/term"

	// tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
)

var debug bool

func printWelcomeScreen() {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	fmt.Println()

	title := "Welcome to the a8s Data Service demos"

	color.Blue("Currently the a8s PostgreSQL or short a8s-pg demo is selected.")

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

func main() {

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	printWelcomeScreen()

	// demo.EstablishConfigFilePath()

	// if !demo.LoadConfig() {
	// 	demo.EstablishWorkingDir()
	// }

	// demo.CheckPrerequisites()

	// demo.CheckoutDeploymentGitRepository()

	// if demo.CountPodsInDemoNamespace() == 0 {
	// 	color.Green("Kubernetes cluster has no pods in " + demo.GetConfig().DemoSpace + " namespace.")
	// }

	// demo.EstablishBackupStoreCredentials()

	// demo.ApplyCertManagerManifests()

	// demo.ApplyA8sManifests()
}
