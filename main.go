package main

/*
Next:

Ask for details for backup store config instead of using defaults.


TODO:


- Use Cases:
	- Pre-Create
		- Create S3 bucket with configs
	- Create
		- waitForA8sToBecomeReady
	- Delete
		- Remove cluster
		- Remove everything (incl. config files)
*/

import (
	"os"

	"github.com/fischerjulian/a8s-demo/cmd"
)

var debug bool

func main() {

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	cmd.Execute()
}
