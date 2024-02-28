package main

import (
	"os"

	"github.com/anynines/a9s-cli-v2/cmd"
)

var Debug bool

func main() {
	if os.Getenv("DEBUG") != "" {
		Debug = true
	}

	cmd.Execute()
}
