//go:build tinygo && js

package js

import (
	"syscall/js"

	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/platform"
)

type Backend struct {
	canvasID string
}

func NewBackend(canvasID string) *Backend {
	return &Backend{canvasID: canvasID}
}

func (b *Backend) Render(dl *canvas.DrawList) error {
	// Browser rendering is implemented in the JS host bridge.
	// This is a placeholder for the Go-to-JS bridge.
	return nil
}

func RegisterBackend(backend platform.Backend) {
	// Entry point for runtime registration.
	// The JS host bridge will call this during initialization.
}

// Exported functions for JS host bridge
func InitRuntime(canvasID string) {
	// Initialize the runtime with the canvas element ID
	// This will be called from the generated aethium.js
}

func RenderFrame(dl *canvas.DrawList) {
	// Called from JS to render a frame
	// The actual rendering happens in the JS host bridge
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
