package canvas

import "sync"

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
	Kind      CmdKind
	X, Y, W, H float32
	Color     Color
	Text      string
	Transform [6]float32
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
