package demo

/*
List of CLI and other binaries required for using the a9s CLI.
*/

//var requiredCommands map[string]map[string]string

func RequiredCommands() map[string]map[string]string {
	cmds := make(map[string]map[string]string)

	cmds["git"] = make(map[string]string)
	cmds["git"]["darwin"] = "brew install git"

	cmds["docker"] = make(map[string]string)
	cmds["docker"]["darwin"] = "brew install docker"

	cmds["cmctl"] = make(map[string]string)
	cmds["cmctl"]["darwin"] = "brew install cmctl"

	return cmds
}
