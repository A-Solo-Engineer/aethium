package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "new":
		newCmd := flag.NewFlagSet("new", flag.ExitOnError)
		module := newCmd.String("module", "", "Go module path for the new app")
		template := newCmd.String("template", "minimal", "Template to scaffold")
		dir := newCmd.String("dir", ".", "Output directory")
		newCmd.Parse(os.Args[2:])
		fmt.Printf("scaffold new app: module=%s template=%s dir=%s\n", *module, *template, *dir)
	case "build":
		fmt.Println("build command placeholder")
	case "dev":
		fmt.Println("dev server placeholder")
	case "bundle":
		fmt.Println("bundle command placeholder")
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		usage()
	}
}

func usage() {
	fmt.Println("Usage: aethium <command> [flags]")
	fmt.Println("Commands: new, build, dev, bundle")
}
