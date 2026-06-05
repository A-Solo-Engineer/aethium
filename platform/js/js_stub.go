//go:build tinygo

package js

import (
	"syscall/js"

	"github.com/A-Solo-Engineer/aethium/canvas"
)

type Backend struct {
	canvasID string
}

func NewBackend(canvasID string) *Backend {
	return &Backend{canvasID: canvasID}
}

func (b *Backend) Render(dl *canvas.DrawList) error {
	if dl == nil {
		return nil
	}

	// Convert DrawCmds to JS-friendly format (flat array for performance)
	cmds := make([]any, len(dl.Cmds))
	for i, cmd := range dl.Cmds {
		cmdData := []any{
			uint8(cmd.Kind),
			cmd.X,
			cmd.Y,
			cmd.W,
			cmd.H,
			uint32(cmd.Color),
			cmd.Text,
		}
		if cmd.Kind == canvas.CmdTransform {
			transform := make([]any, 6)
			for j, v := range cmd.Transform {
				transform[j] = v
			}
			cmdData = append(cmdData, transform)
		}
		cmds[i] = cmdData
	}

	js.Global().Get("Aethium").Call("renderFrame", cmds)
	return nil
}

var globalBackend *Backend

func RegisterBackend(backend *Backend) {
	globalBackend = backend
}

// Exported functions for JS host bridge
func InitRuntime(canvasID string) {
	backend := NewBackend(canvasID)
	RegisterBackend(backend)
	js.Global().Get("Aethium").Call("initRuntime", canvasID)
}

func RenderFrame(dl *canvas.DrawList) {
	if globalBackend != nil {
		globalBackend.Render(dl)
	}
}

func ScheduleOnUI(fn func()) {
	// Schedule a function to run on the UI thread
	// Implemented via queueMicrotask and requestAnimationFrame
	js.Global().Get("queueMicrotask").Invoke(js.FuncOf(func(this js.Value, args []js.Value) any {
		fn()
		return nil
	}))
}

func PumpEvents() {
	// Called by the JS event loop to drain the UI queue
	// This is a placeholder; actual implementation in JS host bridge
}
