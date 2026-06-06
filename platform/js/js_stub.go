//go:build tinygo || js

package js

import (
	"syscall/js"

	"github.com/A-Solo-Engineer/aethium/canvas"
)

type WebBackend struct {
	canvasID string
}

func NewWebBackend(canvasID string) *WebBackend {
	return &WebBackend{canvasID: canvasID}
}

func (b *WebBackend) Render(dl *canvas.DrawList) error {
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

// Implement canvas.Graphics for the WebBackend if we ever want to call it directly
func (b *WebBackend) FillRect(x, y, w, h float32, color canvas.Color) {
	js.Global().Get("Aethium").Call("fillRect", x, y, w, h, uint32(color))
}

func (b *WebBackend) StrokeRect(x, y, w, h float32, color canvas.Color) {
	js.Global().Get("Aethium").Call("strokeRect", x, y, w, h, uint32(color))
}

func (b *WebBackend) DrawText(x, y float32, text string, color canvas.Color) {
	js.Global().Get("Aethium").Call("drawText", x, y, text, uint32(color))
}

func (b *WebBackend) SetClip(x, y, w, h float32) {
	js.Global().Get("Aethium").Call("setClip", x, y, w, h)
}

func (b *WebBackend) SetTransform(matrix [6]float32) {
	m := make([]any, 6)
	for i, v := range matrix {
		m[i] = v
	}
	js.Global().Get("Aethium").Call("setTransform", m)
}
