package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type BuildConfig struct {
	Target    string
	Release   bool
	OutputDir string
	Tags      string
}

func Build(cfg BuildConfig) error {
	fmt.Printf("Building for %s target...\n", cfg.Target)

	switch cfg.Target {
	case "wasm":
		return buildWasm(cfg)
	case "desktop":
		return buildDesktop(cfg)
	case "all":
		if err := buildWasm(cfg); err != nil {
			return err
		}
		return buildDesktop(cfg)
	default:
		return fmt.Errorf("unknown target: %s", cfg.Target)
	}
}

func buildWasm(cfg BuildConfig) error {
	fmt.Println("Building for Wasm (TinyGo)...")

	// TinyGo build command
	cmd := exec.Command("tinygo", "build", "-o", filepath.Join(cfg.OutputDir, "app.wasm"), "-target=wasm", "-tags=tinygo", ".")
	cmd.Dir = cfg.OutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tinygo build failed: %w", err)
	}

	fmt.Println("Wasm build complete")
	return nil
}

func buildDesktop(cfg BuildConfig) error {
	fmt.Println("Building for desktop (Go)...")

	// Go build command
	buildArgs := []string{"build"}
	if cfg.Release {
		buildArgs = append(buildArgs, "-ldflags", "-s -w")
	}
	buildArgs = append(buildArgs, "-o", filepath.Join(cfg.OutputDir, "aethium-app"))
	buildArgs = append(buildArgs, ".")

	cmd := exec.Command("go", buildArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	fmt.Println("Desktop build complete")
	return nil
}
