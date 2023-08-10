package demo

import (
	"bufio"
	"fmt"
	"os"

	"github.com/NilPointer-Software/emoji"
	"github.com/charmbracelet/lipgloss"
)

var (

	// General.

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#e4833e", Dark: "#e4833e"}
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

func ListBaseStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		PaddingRight(1).
		PaddingLeft(1)
}

func ListEmoji(s string, theEmoji emoji.Emoji) string {
	symbolStr := ListBaseStyle().SetString(fmt.Sprintf("%v", theEmoji)).
		// Foreground(special).
		String()

	return symbolStr + lipgloss.NewStyle().Render(s)
}

func ListCheckmark(s string) string {
	return ListEmoji(s, emoji.CheckMarkButton)
}

func ListWait(s string) string {
	return ListEmoji(s, emoji.HourglassNotDone)
}

func ListFail(s string) string {
	return ListEmoji(s, emoji.CrossMark)
}

func ListWarning(s string) string {
	return ListEmoji(s, emoji.Warning)
}

func ListFlexedBiceps(s string) string {
	return ListEmoji(s, emoji.Emoji(emoji.FlexedBiceps.Tone(emoji.Default)))
}

func ListFailSummary(s string) string {
	// TODO DRY for style
	checkMark := lipgloss.NewStyle().SetString(fmt.Sprintf("%v", emoji.PoliceCarLight)).
		// Foreground(special).
		PaddingRight(1).
		PaddingLeft(1).
		Underline(true).
		Blink(true).
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
		Foreground(lipgloss.AdaptiveColor{Light: "#5a6987", Dark: "#505d78"}).
		Render(s)
}

func RegularText(s string) string {
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

func PrintWait(s string) {
	fmt.Println(ListWait(s))
}

func PrintWarning(s string) {
	fmt.Println(ListWarning(s))
}

func PrintFailSummary(s string) {
	fmt.Println(ListFailSummary(s))
}

func PrintCheckmark(s string) {
	fmt.Println(ListCheckmark(s))
}
func PrintFlexedBiceps(s string) {
	fmt.Println(ListFlexedBiceps(s))
}

func Print(s string) {
	fmt.Println(RegularText(s))
}

func WaitForUser() {

	msg := "Press ENTER key to continue"
	style := lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		MarginLeft(1).
		Foreground(highlight).
		Underline(true).
		Render(msg)

	fmt.Println(style)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
