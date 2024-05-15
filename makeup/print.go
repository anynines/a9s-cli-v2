package makeup

import (
	"bufio"
	"fmt"
	"os"

	"github.com/NilPointer-Software/emoji"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var Verbose bool

/*
The makeup package contains helper methods to format output and print messages to the user.
*/

func PrintWelcomeScreen(unattendedMode bool, title, subtitle string) {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	fmt.Println()

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

	PrintH2(subtitle)

	WaitForUser(unattendedMode)
}

func PrintCommandBox(s string) {
	fmt.Println(CommandBox(s))
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

func PrintSuccess(s string) {
	PrintCheckmark(s)
}

func PrintFlexedBiceps(s string) {
	fmt.Println(ListFlexedBiceps(s))
}

func Print(s string) {
	fmt.Println(RegularText(s))
}

func PrintBright(s string) {
	fmt.Println(Bright(s))
}

/*
Prints only if the Verbose flag is set.
*/
func PrintVerbose(s string) {
	if Verbose {
		fmt.Println(H2(s))
	}
}

func PrintSuccessSummary(s string) {
	fmt.Println(ListParty(s))
}

func PrintInfo(s string) {
	PrintEmoji(" "+s, emoji.Information)
}

func WaitForUser(unattendedMode bool) {
	if !unattendedMode {
		msg := "Press <ENTER> key to continue or <CTRL>+C to abort."
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

		PrintEmoji("...", emoji.Emoji(emoji.ManRunning.Tone(emoji.Default)))
	}
}

func ExitDueToFatalError(err error, msg string) {
	PrintFail(msg)
	fmt.Print(err)
	os.Exit(1)
}
