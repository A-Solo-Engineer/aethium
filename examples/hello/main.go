package main

import (
	"os"

	"github.com/A-Solo-Engineer/aethium/examples/hello/app"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		app.RunDev()
	} else {
		app.Run()
	}
}

