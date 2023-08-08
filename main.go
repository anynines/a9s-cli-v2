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

var (

	// General.

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	debug     bool
)

// type mainMenuModel struct {
// 	choices  []string // items on the list
// 	cursor   int      // which item our cursor is pointing at
// 	selected int      // wich todo items are selected
// }

// func initialModel() mainMenuModel {
// 	return mainMenuModel{
// 		choices: []string{"Autopilot Mode", "Interactive Mode", "Yet another mode"},

// 		selected: 0,
// 	}
// }

// func (m mainMenuModel) Init() tea.Cmd {
// 	// Just return `nil`, which means "no I/O right now, please."
// 	return nil
// }

// func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {

// 	// Is it a key press?
// 	case tea.KeyMsg:

// 		// Cool, what was the actual key pressed?
// 		switch msg.String() {

// 		// These keys should exit the program.
// 		case "ctrl+c", "q":
// 			return m, tea.Quit

// 		// The "up" and "k" keys move the cursor up
// 		case "up", "k":
// 			if m.cursor > 0 {
// 				m.cursor--
// 			}

// 		// The "down" and "j" keys move the cursor down
// 		case "down", "j":
// 			if m.cursor < len(m.choices)-1 {
// 				m.cursor++
// 			}
// 		case "n":
// 			if m.selected >= 0 {
// 				return m, tea.Quit
// 			}

// 		// The "enter" key and the spacebar (a literal space) toggle
// 		// the selected state for the item that the cursor is pointing at.
// 		case "enter", " ":
// 			//_, ok := m.selected[m.cursor]

// 			m.selected = m.cursor
// 		}
// 	}

// 	// Return the updated model to the Bubble Tea runtime for processing.
// 	// Note that we're not returning a command.
// 	return m, nil
// }

func listCheckmark(s string) string {

	checkMark := lipgloss.NewStyle().SetString("✓").
		Foreground(special).
		PaddingRight(1).
		String()

	return checkMark + lipgloss.NewStyle().
		// Strikethrough(true).
		//Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
		Render(s)
}

// func (m mainMenuModel) View() string {
// 	// The header
// 	s := "How do you want to run the demo?\n\n"

// 	if debug {
// 		s += fmt.Sprintf("Cursor: %d\n", m.cursor)
// 		s += fmt.Sprintf("Selected: %d\n\n", m.selected)
// 	}

// 	// Iterate over our choices
// 	for i, choice := range m.choices {

// 		// Is the cursor pointing at this choice?
// 		cursor := " " // no cursor
// 		if m.cursor == i {
// 			cursor = ">" // cursor!
// 		}

// 		// Is this choice selected?
// 		checked := " " // not selected

// 		if m.selected == i {
// 			checked = "x" // selected!
// 		}

// 		// Render the row
// 		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
// 	}

// 	// s += listCheckmark("F and F")

// 	// The footer
// 	s += "\nPress q to quit.\n"
// 	s += "\nPress n to proceed with your selection.\n"

// 	// Send the UI for rendering
// 	return s
// }

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

	// p := tea.NewProgram(initialModel())
	// if _, err := p.Run(); err != nil {
	// 	demo.ExitDueToFatalError(err, "Upsi!")
	// }

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
