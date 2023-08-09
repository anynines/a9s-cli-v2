package demo

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (

	// General.

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	list = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(subtle).
		MarginRight(2).
		MarginLeft(1).
		MarginTop(1).
		Height(8).
		Width(30)

	listHeader = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2).
			MarginLeft(1).
			Render

	listItem = lipgloss.NewStyle().PaddingLeft(2).Render
)

func PrintListFromMultilineString(header, multiLineString string) {

	// lines := strings.Builder{}

	// strings.ea

	myList := lipgloss.JoinVertical(lipgloss.Left,
		listHeader(header),
		multiLineString,
	)

	fmt.Println(list.Render(myList))
}

func ListCheckmark(s string) string {

	checkMark := lipgloss.NewStyle().SetString("\u2705").
		Foreground(special).
		PaddingRight(1).
		PaddingLeft(1).
		String()

	return checkMark + lipgloss.NewStyle().
		// Strikethrough(true).
		//Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
		Render(s)
}

func ListFail(s string) string {
	checkMark := lipgloss.NewStyle().SetString("\u274C").
		// Foreground(special).
		PaddingRight(1).
		PaddingLeft(1).
		String()

	return checkMark + lipgloss.NewStyle().
		// Strikethrough(true).
		//Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
		Render(s)
}

func H1(s string) string {
	return lipgloss.NewStyle().
		Underline(true).
		PaddingLeft(1).
		PaddingTop(1).
		PaddingBottom(1).
		Bold(true).
		Render(s)
}

func H2(s string) string {
	return lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
		Render(s)
}

func PrintH1(s string) {
	fmt.Println(H1(s))
}

func PrintH2(s string) {
	fmt.Println(H2(s))
}

func PrintFail(s string) {
	fmt.Println(ListFail(s))
}

func PrintCheckmark(s string) {
	fmt.Println(ListCheckmark(s))
}
