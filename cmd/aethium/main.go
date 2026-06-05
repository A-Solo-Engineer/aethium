package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
		if err := scaffoldNewApp(*module, *template, *dir); err != nil {
			fmt.Printf("Error scaffolding app: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Scaffolding new app: module=%s template=%s dir=%s\n", *module, *template, *dir)
	case "build":
		buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
		target := buildCmd.String("target", "desktop", "Target: wasm, desktop, or all")
		release := buildCmd.Bool("release", true, "Release build")
		output := buildCmd.String("output", "dist/", "Output directory")
		tags := buildCmd.String("tags", "", "Build tags")
		buildCmd.Parse(os.Args[2:])
		if err := buildApp(*target, *release, *output, *tags); err != nil {
			fmt.Printf("Error building app: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Building for %s target...\n", *target)
		fmt.Printf("Release: %v, Output: %s\n", *release, *output)
	case "dev":
		devCmd := flag.NewFlagSet("dev", flag.ExitOnError)
		target := devCmd.String("target", "wasm", "Target: wasm or desktop")
		addr := devCmd.String("addr", "127.0.0.1:5173", "Bind address")
		open := devCmd.Bool("open", false, "Open browser")
		devCmd.Parse(os.Args[2:])
		if err := startDevServer(*target, *addr, *open); err != nil {
			fmt.Printf("Error starting dev server: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Starting dev server for %s target at %s\n", *target, *addr)
	case "bundle":
		bundleCmd := flag.NewFlagSet("bundle", flag.ExitOnError)
		target := bundleCmd.String("target", "desktop", "Target: desktop or wasm")
		format := bundleCmd.String("format", "native", "Format: native or wasm")
		output := bundleCmd.String("output", "dist/bundle", "Output file")
		bundleCmd.Parse(os.Args[2:])
		if err := createBundle(*target, *format, *output); err != nil {
			fmt.Printf("Error creating bundle: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Creating bundle for %s target in %s format\n", *target, *format)
		fmt.Printf("Output: %s\n", *output)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func scaffoldNewApp(module, template, dir string) error {
	fmt.Printf("Scaffolding new app: module=%s template=%s dir=%s\n", module, template, dir)

	// Create basic directory structure
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create go.mod
	goModContent := fmt.Sprintf("module %s\n\ngo 1.22\n\nrequire github.com/A-Solo-Engineer/aethium v0.0.1\n", module)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create main.go
	mainContent := fmt.Sprintf(`package main

import (
	"fmt"
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
`, module)
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Copy example app
	exampleDir := filepath.Join(dir, "app")
	if err := os.MkdirAll(exampleDir, 0755); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	// Copy counter.go
	counterContent := `package app

import (
	"fmt"
	"time"

	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/reactive"
	"github.com/A-Solo-Engineer/aethium/runtime"
)

type Counter struct {
	count *reactive.Signal[int]
}

func NewCounter() *Counter {
	return &Counter{
		count: reactive.NewSignal(0),
	}
}

func (c *Counter) Init(ctx runtime.InitContext) error {
	fmt.Println("Counter initialized")
	return nil
}

func (c *Counter) Mount(ctx runtime.MountContext) error {
	fmt.Println("Counter mounted")
	return nil
}

func (c *Counter) Update(ctx runtime.UpdateContext) error {
	// Update logic would go here
	return nil
}

func (c *Counter) Unmount(ctx runtime.UnmountContext) error {
	fmt.Println("Counter unmounted")
	return nil
}

func (c *Counter) View() []canvas.DrawCmd {
	count := c.count.Get()
	return []canvas.DrawCmd{
		{Kind: canvas.CmdFillRect, X: 10, Y: 10, W: 100, H: 50, Color: 0xFF0000FF},
		{Kind: canvas.CmdDrawText, X: 120, Y: 35, Text: fmt.Sprintf("Count: %d", count), Color: 0xFFFFFFFF},
	}
}

func Run() {
	fmt.Println("Aethium hello example starting...")

	// Create runtime
	runtime := runtime.NewRuntime(nil)

	// Create counter component
	counter := NewCounter()

	// Attach to runtime
	if err := runtime.Attach(counter); err != nil {
		fmt.Printf("Error attaching counter: %v\n", err)
		return
	}

	fmt.Println("Starting render loop...")
	frame := 0
	for {
		// Update counter every 60 frames
		if frame%60 == 0 {
			count := counter.count.Get()
			newCount := count + 1
			counter.count.Set(newCount)
			fmt.Printf("Frame %d: Count = %d\n", frame, newCount)
		}

		// Render frame
		if err := runtime.Tick(); err != nil {
			fmt.Printf("Error rendering frame: %v\n", err)
			break
		}

		frame++
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}
}

func RunDev() {
	fmt.Println("Starting Aethium dev server...")

	// In dev mode, we'd start the dev server
	// For now, just run the app
	Run()
}
`
	if err := os.WriteFile(filepath.Join(exampleDir, "counter.go"), []byte(counterContent), 0644); err != nil {
		return fmt.Errorf("failed to create counter.go: %w", err)
	}

	return nil
}

func buildApp(target string, release bool, output, tags string) error {
	fmt.Printf("Building for %s target...\n", target)

	// Create output directory
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	switch target {
	case "wasm":
		cmd := exec.Command("tinygo", "build", "-o", filepath.Join(output, "app.wasm"), "-target=wasi", "-tags=tinygo", ".")
		cmd.Dir = "."
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("wasm build failed: %v, output: %s", err, string(output))
		}
	case "desktop":
		buildTags := ""
		if tags != "" {
			buildTags = "-tags=" + tags
		}
		cmd := exec.Command("go", "build", "-o", filepath.Join(output, "app.exe"), buildTags, ".")
		cmd.Dir = "."
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("desktop build failed: %v, output: %s", err, string(output))
		}
	case "all":
		if err := buildApp("wasm", release, output, tags); err != nil {
			return err
		}
		if err := buildApp("desktop", release, output, tags); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown target: %s", target)
	}

	return nil
}

func startDevServer(target, addr string, open bool) error {
	fmt.Printf("Starting dev server for %s target at %s\n", target, addr)

	// Start dev server
	// For now, just serve the current directory
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)
			http.ServeFile(w, r, r.URL.Path[1:])
		}),
	}

	if open {
		// Open browser
		url := "http://" + addr
		cmd := exec.Command("cmd", "/c", "start", url)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to open browser: %v\n", err)
		}
	}

	fmt.Printf("Dev server running at http://%s\n", addr)
	return server.ListenAndServe()
}

func createBundle(target, format, output string) error {
	fmt.Printf("Creating bundle for %s target in %s format\n", target, format)

	// Create bundle directory
	bundleDir := filepath.Dir(output)
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}

	// For now, just copy the built app
	switch target {
	case "desktop":
		src := filepath.Join("dist", "app.exe")
		if err := copyFile(src, output); err != nil {
			return fmt.Errorf("failed to copy desktop app: %w", err)
		}
	case "wasm":
		src := filepath.Join("dist", "app.wasm")
		if err := copyFile(src, output); err != nil {
			return fmt.Errorf("failed to copy wasm app: %w", err)
		}
	default:
		return fmt.Errorf("unknown target: %s", target)
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
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
