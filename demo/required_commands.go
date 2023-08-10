package demo

func RequiredCommands() map[string]map[string]string {
	cmds := make(map[string]map[string]string)

	cmds["docker"] = make(map[string]string)
	cmds["docker"]["darwin"] = "brew install docker"

	cmds["kind"] = make(map[string]string)
	cmds["kind"]["darwin"] = "brew install kind"

	cmds["cmctl"] = make(map[string]string)
	cmds["cmctl"]["darwin"] = "brew install cmctl"

	return cmds
}
