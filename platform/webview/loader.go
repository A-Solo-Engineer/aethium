//go:build !tinygo || !js

package webview

import "github.com/aethium-dev/aethium/platform"

type Loader struct {
	Backend platform.Backend
}

func NewLoader(backend platform.Backend) *Loader {
	return &Loader{Backend: backend}
}

func (l *Loader) Run() error {
	// Desktop WebView host is implemented in the next stage.
	return nil
}
