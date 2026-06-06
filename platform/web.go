//go:build !tinygo && !js

package platform

import (
	"errors"
	"github.com/A-Solo-Engineer/aethium/canvas"
)

type WebBackend struct {
	canvasID string
}

func NewWebBackend(canvasID string) *WebBackend {
	return &WebBackend{canvasID: canvasID}
}

func (b *WebBackend) Render(dl *canvas.DrawList) error {
	return errors.New("WebBackend is only available when compiled for WASM (TinyGo/JS)")
}
