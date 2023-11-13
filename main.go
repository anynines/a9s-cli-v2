package main

import (
	"os"

	"github.com/anynines/a9s-cli-v2/cmd"
)

var debug bool

func main() {

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	cmd.Execute()
}
