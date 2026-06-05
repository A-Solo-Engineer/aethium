//go:build !tinygo

package desktop

import (
	"fmt"

	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/platform"
)

type DesktopBackend struct {
	windowWidth  int
	windowHeight int
}

func NewDesktopBackend(width, height int) *DesktopBackend {
	return &DesktopBackend{
		windowWidth:  width,
		windowHeight: height,
	}
}

func (d *DesktopBackend) Render(dl *canvas.DrawList) error {
	// In a real implementation, this would render to a native window
	// For now, just log the draw commands
	fmt.Printf("Desktop Backend: Rendering %d draw commands to %dx%d window\n", len(dl.Cmds), d.windowWidth, d.windowHeight)
	return nil
}

// Additional desktop-specific methods would go here
func (d *DesktopBackend) SetSize(width, height int) {
	d.windowWidth = width
	d.windowHeight = height
}

func (d *DesktopBackend) GetSize() (int, int) {
	return d.windowWidth, d.windowHeight
}