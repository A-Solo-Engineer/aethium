package canvas

import (
	"sync"
)

type CmdKind uint8

const (
	CmdFillRect CmdKind = iota + 1
	CmdStrokeRect
	CmdDrawText
	CmdClip
	CmdTransform
)

type Color uint32 // ARGB, 0xAARRGGBB

type DrawCmd struct {
	Kind       CmdKind
	X, Y, W, H float32
	Color      Color
	Text       string
	Transform  [6]float32
}

type DrawList struct {
	Cmds []DrawCmd
}

var drawListPool = sync.Pool{
	New: func() any {
		return &DrawList{}
	},
}

func NewDrawList() *DrawList {
	dl := drawListPool.Get().(*DrawList)
	dl.Cmds = dl.Cmds[:0]
	return dl
}

func (dl *DrawList) Release() {
	if dl == nil {
		return
	}
	dl.Cmds = dl.Cmds[:0]
	drawListPool.Put(dl)
}

func (dl *DrawList) Append(cmd DrawCmd) {
	if dl == nil {
		return
	}
	dl.Cmds = append(dl.Cmds, cmd)
}

func (dl *DrawList) AppendSlice(cmds []DrawCmd) {
	if dl == nil || len(cmds) == 0 {
		return
	}
	dl.Cmds = append(dl.Cmds, cmds...)
}

func Merge(dst, src *DrawList) {
	if dst == nil || src == nil {
		return
	}
	dst.Cmds = append(dst.Cmds, src.Cmds...)
	src.Release()
}

type Rect struct {
	X, Y, W, H float32
}

func FillRect(dl *DrawList, r Rect, c Color) {
	dl.Append(DrawCmd{Kind: CmdFillRect, X: r.X, Y: r.Y, W: r.W, H: r.H, Color: c})
}

func StrokeRect(dl *DrawList, r Rect, c Color) {
	dl.Append(DrawCmd{Kind: CmdStrokeRect, X: r.X, Y: r.Y, W: r.W, H: r.H, Color: c})
}

func DrawText(dl *DrawList, x, y float32, text string, c Color) {
	if len(text) > 4096 {
		text = text[:4096]
	}
	dl.Append(DrawCmd{Kind: CmdDrawText, X: x, Y: y, Text: text, Color: c})
}

func Clip(dl *DrawList, r Rect) {
	dl.Append(DrawCmd{Kind: CmdClip, X: r.X, Y: r.Y, W: r.W, H: r.H})
}

func Transform(dl *DrawList, matrix [6]float32) {
	dl.Append(DrawCmd{Kind: CmdTransform, Transform: matrix})
}

type Graphics interface {
	FillRect(x, y, w, h float32, color Color)
	StrokeRect(x, y, w, h float32, color Color)
	DrawText(x, y float32, text string, color Color)
	SetClip(x, y, w, h float32)
	SetTransform(matrix [6]float32)
}

// Render executes all draw commands in the list using the provided graphics interface.
func (dl *DrawList) Render(g Graphics) {
	if dl == nil || g == nil {
		return
	}
	for _, cmd := range dl.Cmds {
		switch cmd.Kind {
		case CmdFillRect:
			g.FillRect(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Color)
		case CmdStrokeRect:
			g.StrokeRect(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Color)
		case CmdDrawText:
			g.DrawText(cmd.X, cmd.Y, cmd.Text, cmd.Color)
		case CmdClip:
			g.SetClip(cmd.X, cmd.Y, cmd.W, cmd.H)
		case CmdTransform:
			g.SetTransform(cmd.Transform)
		}
	}
}

// Reset clears the draw list without releasing to pool
func (dl *DrawList) Reset() {
	if dl == nil {
		return
	}
	dl.Cmds = dl.Cmds[:0]
}

// Cap ensures the draw list has at least the specified capacity
func (dl *DrawList) Cap(capacity int) {
	if dl == nil {
		return
	}
	if cap(dl.Cmds) < capacity {
		newCmds := make([]DrawCmd, 0, capacity)
		newCmds = append(newCmds, dl.Cmds...)
		dl.Cmds = newCmds
	}
}

// Count returns the number of commands in the draw list
func (dl *DrawList) Count() int {
	if dl == nil {
		return 0
	}
	return len(dl.Cmds)
}
