package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type ScaffoldConfig struct {
	Module   string
	Dir      string
	Template string
}

func ScaffoldNew(module, dir, template string) error {
	cfg := ScaffoldConfig{
		Module:   module,
		Dir:      dir,
		Template: template,
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	switch template {
	case "minimal":
		return scaffoldMinimal(cfg)
	default:
		return fmt.Errorf("unknown template: %s", template)
	}
}

func scaffoldMinimal(cfg ScaffoldConfig) error {
	// Create go.mod
	goMod := `module {{.Module}}

go 1.22

require github.com/A-Solo-Engineer/aethium v0.0.0
`
	if err := writeTemplate(filepath.Join(cfg.Dir, "go.mod"), goMod, cfg); err != nil {
		return err
	}

	// Create main.go
	mainGo := `package main

import "github.com/A-Solo-Engineer/aethium/examples/hello/app"

func main() {
	app.Run()
}
`
	if err := writeTemplate(filepath.Join(cfg.Dir, "main.go"), mainGo, cfg); err != nil {
		return err
	}

	// Create app/view.go
	viewGo := `package app

import (
	"fmt"

	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/reactive"
	"github.com/A-Solo-Engineer/aethium/runtime"
)

type Counter struct {
	count *reactive.Signal[int]
}

func NewCounter() *Counter {
	return &Counter{count: reactive.NewSignal(0)}
}

func (c *Counter) Init(ctx runtime.InitContext) error  { return nil }
func (c *Counter) Mount(ctx runtime.MountContext) error { return nil }
func (c *Counter) Update(ctx runtime.UpdateContext) error { return nil }
func (c *Counter) Unmount(ctx runtime.UnmountContext) error { return nil }

func (c *Counter) View() []canvas.DrawCmd {
	count := c.count.Get()
	dl := canvas.NewDrawList()
	defer dl.Release()
	canvas.FillRect(dl, canvas.Rect{X: 10, Y: 10, W: 100, H: 50}, 0xFF0000FF)
	canvas.DrawText(dl, 120, 35, fmt.Sprintf("Count: %d", count), 0xFFFFFFFF)
	cmds := make([]canvas.DrawCmd, len(dl.Cmds))
	copy(cmds, dl.Cmds)
	return cmds
}
`
	if err := writeTemplate(filepath.Join(cfg.Dir, "app", "view.go"), viewGo, cfg); err != nil {
		return err
	}

	// Create app/state.go
	stateGo := `package app

import (
	"fmt"

	"github.com/A-Solo-Engineer/aethium/runtime"
)

func Run() {
	fmt.Println("Aethium app starting...")
	rt := runtime.NewRuntime(nil)
	counter := NewCounter()
	if err := rt.Attach(counter); err != nil {
		fmt.Printf("attach error: %v\n", err)
		return
	}
	for {
		if err := rt.Tick(); err != nil {
			fmt.Printf("tick error: %v\n", err)
			return
		}
	}
}
`
	if err := writeTemplate(filepath.Join(cfg.Dir, "app", "state.go"), stateGo, cfg); err != nil {
		return err
	}

	// Create aethium.toml
	toml := `[build]
target = "wasm"
tinygo_opt = 0

[dev]
addr = "127.0.0.1:5173"
open = true
`
	if err := writeTemplate(filepath.Join(cfg.Dir, "aethium.toml"), toml, cfg); err != nil {
		return err
	}

	fmt.Printf("Successfully scaffolded minimal app in %s\n", cfg.Dir)
	return nil
}

func writeTemplate(path, tmpl string, cfg ScaffoldConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	t, err := template.New("scaffold").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := t.Execute(file, cfg); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

