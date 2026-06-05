//go:build !tinygo || !js

package webview

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/A-Solo-Engineer/aethium/platform"
)

//go:embed assets
var assets embed.FS

type Loader struct {
	Backend platform.Backend
	addr    string
}

func NewLoader(backend platform.Backend) *Loader {
	return &Loader{Backend: backend, addr: "127.0.0.1:5173"}
}

func (l *Loader) Run() error {
	// Serve assets
	http.HandleFunc("/", l.serveIndex)
	http.HandleFunc("/app.wasm", l.serveWasm)
	http.HandleFunc("/aethium.js", l.serveJS)

	// Start server
	fmt.Printf("Serving Aethium app at http://%s\n", l.addr)
	return http.ListenAndServe(l.addr, nil)
}

func (l *Loader) serveIndex(w http.ResponseWriter, r *http.Request) {
	data, err := assets.ReadFile("assets/index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func (l *Loader) serveWasm(w http.ResponseWriter, r *http.Request) {
	data, err := assets.ReadFile("assets/app.wasm")
	if err != nil {
		http.Error(w, "app.wasm not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/wasm")
	w.Write(data)
}

func (l *Loader) serveJS(w http.ResponseWriter, r *http.Request) {
	data, err := assets.ReadFile("assets/aethium.js")
	if err != nil {
		http.Error(w, "aethium.js not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(data)
}

func EmbedAssets(wasmPath, jsPath string) error {
	// Embed assets from files
	// This is a placeholder for the actual embedding logic
	return nil
}

func GetEmbeddedAssets() (fs.FS, error) {
	return assets, nil
}

func GetAssetPath(name string) (string, error) {
	return filepath.Join("assets", name), nil
}

