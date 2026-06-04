package platform

import "github.com/aethium-dev/aethium/canvas"

type Backend interface {
	Render(dl *canvas.DrawList) error
}
