//go:build tinygo && js

package js

import (
	"github.com/aethium-dev/aethium/canvas"
	"github.com/aethium-dev/aethium/platform"
)

type Backend struct{}

func NewBackend() *Backend {
	return &Backend{}
}

func (b *Backend) Render(dl *canvas.DrawList) error {
	// Browser rendering is implemented in the JS host bridge.
	return nil
}

func RegisterBackend(backend platform.Backend) {
	// Entry point for runtime registration.
}
