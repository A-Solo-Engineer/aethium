package platform

import (
	"fmt"
	"github.com/A-Solo-Engineer/aethium/canvas"
)

type DesktopBackend struct {
	Width, Height int
}

func NewDesktopBackend(w, h int) *DesktopBackend {
	return &DesktopBackend{Width: w, Height: h}
}

func (b *DesktopBackend) Render(dl *canvas.DrawList) error {
	fmt.Printf("--- Desktop Render Frame (%d commands) ---\n", dl.Count())
	for _, cmd := range dl.Cmds {
		switch cmd.Kind {
		case canvas.CmdFillRect:
			fmt.Printf("FillRect: at (%.1f, %.1f) size %.1f x %.1f color #%08X\n", cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Color)
		case canvas.CmdDrawText:
			fmt.Printf("DrawText: '%s' at (%.1f, %.1f) color #%08X\n", cmd.Text, cmd.X, cmd.Y, cmd.Color)
		}
	}
	return nil
}
