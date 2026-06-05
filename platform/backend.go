package platform

import "github.com/A-Solo-Engineer/aethium/canvas"

type Backend interface {
	Render(dl *canvas.DrawList) error
}
