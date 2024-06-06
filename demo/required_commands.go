package demo

func RequiredCommands() map[string]map[string]string {
	cmds := make(map[string]map[string]string)

	cmds["docker"] = make(map[string]string)
	cmds["docker"]["darwin"] = "brew install docker"

	cmds["minikube"] = make(map[string]string)
	cmds["minikube"]["darwin"] = "brew install minikube"

	cmds["cmctl"] = make(map[string]string)
	cmds["cmctl"]["darwin"] = "brew install cmctl"

	cmds["minio-mc"] = make(map[string]string)
	cmds["cmctl"]["darwin"] = "brew install minio-mc"

	return cmds
}
