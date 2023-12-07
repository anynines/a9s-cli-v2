package demo

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/anynines/a9s-cli-v2/makeup"
)

func IsCommandAvailable(cmdName string) bool {
	//	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	//	cmd := exec.Command("command", "-v", cmdName)
	// if err := cmd.Run(); err != nil {
	path, err := exec.LookPath(cmdName)
	if err != nil {
		requiredCmds := RequiredCommands()

		msg := "Couldn't find " + cmdName + " command: " + err.Error() + "."

		if requiredCmds[cmdName][runtime.GOOS] != "" {
			msg += " Try running: " + requiredCmds[cmdName][runtime.GOOS]
		}

		makeup.PrintFail(msg)

		return false
	}

	makeup.PrintCheckmark("Found " + cmdName + " at path " + path + ".")

	return true
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
		makeup.PrintFailSummary("Sadly, mandatory commands are missing. Aborting...")
		os.Exit(1)
	} else {
		makeup.PrintSuccessSummary("All necessary commands are present.")
	}
}
