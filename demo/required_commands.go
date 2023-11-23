package demo

func RequiredCommands() map[string]map[string]string {
	cmds := make(map[string]map[string]string)

	cmds["docker"] = make(map[string]string)
	cmds["docker"]["darwin"] = "brew install docker"

	//TODO Depending of which k8s is selected, only the select cmd should be required
	cmds["kind"] = make(map[string]string)
	cmds["kind"]["darwin"] = "brew install kind"

	cmds["minikube"] = make(map[string]string)
	cmds["minikube"]["darwin"] = "brew install minikube"

	cmds["cmctl"] = make(map[string]string)
	cmds["cmctl"]["darwin"] = "brew install cmctl"

	return cmds
}
