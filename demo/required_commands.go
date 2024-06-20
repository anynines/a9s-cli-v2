package demo

/*
List of CLI and other binaries required for using the a9s CLI.
*/

//var requiredCommands map[string]map[string]string

func RequiredCommands() map[string]map[string]string {
	cmds := make(map[string]map[string]string)

	cmds["docker"] = make(map[string]string)
	cmds["docker"]["darwin"] = "brew install docker"

	cmds["minikube"] = make(map[string]string)
	cmds["minikube"]["darwin"] = "brew install minikube"

	cmds["cmctl"] = make(map[string]string)
	cmds["cmctl"]["darwin"] = "brew install cmctl"

	return cmds
}
