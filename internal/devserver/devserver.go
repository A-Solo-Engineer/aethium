package devserver

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type Server struct {
	Addr      string
	Target    string
	Open      bool
	BuildPath string
	BuildCmd  *exec.Cmd
}

func New(addr, target string, open bool) *Server {
	return &Server{
		Addr:      addr,
		Target:    target,
		Open:      open,
		BuildPath: ".",
	}
}

func (s *Server) Start() error {
	if s.Addr == "" {
		s.Addr = "127.0.0.1:5173"
	}

	// Start build process in background
	go func() {
		fmt.Println("Starting initial build...")
		// Use 'go build' instead of 'go run' to generate app.wasm if target is wasm
		if s.Target == "wasm" {
			s.BuildCmd = exec.Command("tinygo", "build", "-o", "app.wasm", "-target=wasm", ".")
		} else {
			s.BuildCmd = exec.Command("go", "build", "-o", "aethium-app", ".")
		}
		s.BuildCmd.Dir = s.BuildPath
		s.BuildCmd.Stdout = os.Stdout
		s.BuildCmd.Stderr = os.Stderr
		if err := s.BuildCmd.Run(); err != nil {
			fmt.Printf("Initial build failed: %v\n", err)
		} else {
			fmt.Println("Initial build complete")
		}
	}()

	// Register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serveIndex)
	mux.HandleFunc("/app.wasm", s.serveWasm)
	mux.HandleFunc("/aethium.js", s.serveJS)

	fmt.Printf("Serving Aethium app at http://%s\n", s.Addr)
	if s.Open {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser("http://" + s.Addr)
		}()
	}

	server := &http.Server{
		Addr:    s.Addr,
		Handler: mux,
	}

	return server.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.BuildCmd != nil && s.BuildCmd.Process != nil {
		return s.BuildCmd.Process.Kill()
	}
	return nil
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("platform", "webview", "assets", "index.html")
	data, err := os.ReadFile(path)
	if err != nil {
		// Fallback to minimal index
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><script src="/aethium.js"></script></head><body><canvas id="canvas" width="800" height="600"></canvas><script>Aethium.initRuntime('canvas');</script></body></html>`))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func (s *Server) serveWasm(w http.ResponseWriter, r *http.Request) {
	// Check in build path first, then current dir
	wasmPath := filepath.Join(s.BuildPath, "app.wasm")
	if _, err := os.Stat(wasmPath); err != nil {
		wasmPath = "app.wasm"
	}
	data, err := os.ReadFile(wasmPath)
	if err != nil {
		http.Error(w, "app.wasm not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/wasm")
	w.Write(data)
}

func (s *Server) serveJS(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("platform", "webview", "assets", "aethium.js")
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "aethium.js not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(data)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
