//go:build !tinygo

package webview

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", l.serveIndex)
	mux.HandleFunc("/app.wasm", l.serveWasm)
	mux.HandleFunc("/aethium.js", l.serveJS)

	// Start server
	fmt.Printf("Serving Aethium app at http://%s\n", l.addr)
	server := &http.Server{
		Addr:    l.addr,
		Handler: mux,
	}
	return server.ListenAndServe()
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
	// Copy the provided WASM and JS files into the assets directory
	// so they can be embedded in the next build.
	if wasmPath != "" {
		data, err := os.ReadFile(wasmPath)
		if err != nil {
			return fmt.Errorf("failed to read WASM: %w", err)
		}
		if err := os.WriteFile("platform/webview/assets/app.wasm", data, 0644); err != nil {
			return fmt.Errorf("failed to write WASM asset: %w", err)
		}
	}
	if jsPath != "" {
		data, err := os.ReadFile(jsPath)
		if err != nil {
			return fmt.Errorf("failed to read JS: %w", err)
		}
		if err := os.WriteFile("platform/webview/assets/aethium.js", data, 0644); err != nil {
			return fmt.Errorf("failed to write JS asset: %w", err)
		}
	}
	return nil
}

func GetEmbeddedAssets() (fs.FS, error) {
	return assets, nil
}

func GetAssetPath(name string) (string, error) {
	return filepath.Join("assets", name), nil
}
