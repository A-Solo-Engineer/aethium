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

	// Check if tinygo is installed
	if _, err := exec.LookPath("tinygo"); err != nil {
		return fmt.Errorf("tinygo not found in PATH: %w. Please install TinyGo to build for WASM", err)
	}

	// TinyGo build command
	outputPath := filepath.Join(cfg.OutputDir, "app.wasm")
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	cmd := exec.Command("tinygo", "build", "-o", outputPath, "-target=wasm", "-tags=tinygo", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tinygo build failed: %w", err)
	}

	fmt.Printf("Wasm build complete: %s\n", outputPath)
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

func Bundle(cfg BuildConfig) error {
	fmt.Printf("Creating bundle for %s target...\n", cfg.Target)

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build first
	if err := Build(cfg); err != nil {
		return err
	}

	// Copy assets based on target
	if cfg.Target == "wasm" || cfg.Target == "all" {
		assets := []string{"index.html", "aethium.js"}
		for _, asset := range assets {
			src := filepath.Join("platform", "webview", "assets", asset)
			dst := filepath.Join(cfg.OutputDir, asset)
			if err := copyFile(src, dst); err != nil {
				fmt.Printf("Warning: failed to copy asset %s: %v\n", asset, err)
			}
		}
	}

	fmt.Println("Bundle created successfully")
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
