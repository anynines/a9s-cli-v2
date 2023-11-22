package main

import (
	"os"

	"github.com/anynines/a9s-cli-v2/cmd"
	"github.com/anynines/a9s-cli-v2/demo"
)

var Debug bool

func main() {

	//TODO Make configurable > Command line option
	// Valid options: "kind"
	demo.KubernetesTool = "minikube"

	if os.Getenv("DEBUG") != "" {
		Debug = true
	}

	cmd.Execute()
}
