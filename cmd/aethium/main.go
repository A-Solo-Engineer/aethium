package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/A-Solo-Engineer/aethium/internal/build"
	"github.com/A-Solo-Engineer/aethium/internal/devserver"
	"github.com/A-Solo-Engineer/aethium/internal/scaffold"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "new":
		newCmd := flag.NewFlagSet("new", flag.ExitOnError)
		module := newCmd.String("module", "", "Go module path for the new app")
		template := newCmd.String("template", "minimal", "Template to scaffold")
		dir := newCmd.String("dir", ".", "Output directory")
		newCmd.Parse(os.Args[2:])
		if *module == "" {
			fmt.Println("Error: --module is required")
			newCmd.Usage()
			os.Exit(1)
		}
		if err := scaffold.ScaffoldNew(*module, *dir, *template); err != nil {
			fmt.Printf("Error scaffolding project: %v\n", err)
			os.Exit(1)
		}
	case "build":
		buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
		target := buildCmd.String("target", "desktop", "Target: wasm, desktop, or all")
		release := buildCmd.Bool("release", true, "Release build")
		output := buildCmd.String("output", "dist/", "Output directory")
		tags := buildCmd.String("tags", "", "Build tags")
		buildCmd.Parse(os.Args[2:])
		cfg := build.BuildConfig{
			Target:    *target,
			Release:   *release,
			OutputDir: *output,
			Tags:      *tags,
		}
		if err := build.Build(cfg); err != nil {
			fmt.Printf("Build failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Build complete")
	case "dev":
		devCmd := flag.NewFlagSet("dev", flag.ExitOnError)
		target := devCmd.String("target", "wasm", "Target: wasm or desktop")
		addr := devCmd.String("addr", "127.0.0.1:5173", "Bind address")
		open := devCmd.Bool("open", false, "Open browser")
		devCmd.Parse(os.Args[2:])
		srv := devserver.New(*addr, *target, *open)
		if err := srv.Start(); err != nil {
			fmt.Printf("Dev server failed: %v\n", err)
			os.Exit(1)
		}
	case "bundle":
		bundleCmd := flag.NewFlagSet("bundle", flag.ExitOnError)
		target := bundleCmd.String("target", "desktop", "Target: desktop or wasm")
		output := bundleCmd.String("output", "dist/bundle", "Output directory")
		bundleCmd.Parse(os.Args[2:])
		cfg := build.BuildConfig{
			Target:    *target,
			OutputDir: *output,
			Release:   true,
		}
		if err := build.Bundle(cfg); err != nil {
			fmt.Printf("Bundle failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Bundle created successfully")
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: aethium <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  new       Scaffold a new project")
	fmt.Println("  build     Build the application")
	fmt.Println("  dev       Start development server")
	fmt.Println("  bundle    Create a distributable bundle")
	fmt.Println()
	fmt.Println("Run 'aethium <command> -h' for command-specific help")
}
