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
	dl.Render(b)
	return nil
}

func (b *DesktopBackend) FillRect(x, y, w, h float32, color canvas.Color) {
	fmt.Printf("FillRect: at (%.1f, %.1f) size %.1f x %.1f color #%08X\n", x, y, w, h, color)
}

func (b *DesktopBackend) StrokeRect(x, y, w, h float32, color canvas.Color) {
	fmt.Printf("StrokeRect: at (%.1f, %.1f) size %.1f x %.1f color #%08X\n", x, y, w, h, color)
}

func (b *DesktopBackend) DrawText(x, y float32, text string, color canvas.Color) {
	fmt.Printf("DrawText: '%s' at (%.1f, %.1f) color #%08X\n", text, x, y, color)
}

func (b *DesktopBackend) SetClip(x, y, w, h float32) {
	fmt.Printf("SetClip: at (%.1f, %.1f) size %.1f x %.1f\n", x, y, w, h)
}

func (b *DesktopBackend) SetTransform(matrix [6]float32) {
	fmt.Printf("SetTransform: %v\n", matrix)
}
