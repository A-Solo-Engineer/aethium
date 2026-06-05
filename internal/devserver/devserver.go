package devserver

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
		Addr:    addr,
		Target:  target,
		Open:    open,
		BuildPath: ".",
	}
}

func (s *Server) Start() error {
	if s.Addr == "" {
		s.Addr = "127.0.0.1:5173"
	}

	// Start build process in background
	s.BuildCmd = exec.Command("go", "run", ".")
	s.BuildCmd.Dir = s.BuildPath
	s.BuildCmd.Stdout = os.Stdout
	s.BuildCmd.Stderr = os.Stderr

	if err := s.BuildCmd.Start(); err != nil {
		return fmt.Errorf("failed to start build process: %w", err)
	}

	// Wait for build to complete
	if err := s.BuildCmd.Wait(); err != nil {
		return fmt.Errorf("build process failed: %w", err)
	}

	// Start HTTP server
	http.HandleFunc("/", s.serveIndex)
	http.HandleFunc("/app.wasm", s.serveWasm)
	http.HandleFunc("/aethium.js", s.serveJS)

	fmt.Printf("Serving Aethium app at http://%s\n", s.Addr)
	if s.Open {
		go func() {
			time.Sleep(100 * time.Millisecond)
			openBrowser("http://" + s.Addr)
		}()
	}

	return http.ListenAndServe(s.Addr, nil)
}

func (s *Server) Stop() error {
	if s.BuildCmd != nil && s.BuildCmd.Process != nil {
		return s.BuildCmd.Process.Kill()
	}
	return nil
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Aethium App</title>
    <script src="/aethium.js"></script>
</head>
<body>
    <canvas id="canvas" width="800" height="600"></canvas>
    <script>
        // Initialize the runtime
        Aethium.initRuntime('canvas');
    </script>
</body>
</html>`))
}

func (s *Server) serveWasm(w http.ResponseWriter, r *http.Request) {
	wasmPath := filepath.Join(s.BuildPath, "app.wasm")
	data, err := os.ReadFile(wasmPath)
	if err != nil {
		http.Error(w, "app.wasm not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/wasm")
	w.Write(data)
}

func (s *Server) serveJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(`// Aethium JS Host Bridge
// This is a placeholder for the actual JS bridge

const Aethium = {
    initRuntime: function(canvasID) {
        console.log('Initializing runtime with canvas:', canvasID);
        // Initialize the runtime here
    },

    renderFrame: function(dl) {
        console.log('Rendering frame with', dl.Count(), 'commands');
        // Render the frame here
    },

    scheduleOnUI: function(fn) {
        queueMicrotask(fn);
    },

    pumpEvents: function() {
        // Pump events from the UI queue
    }
};

// Start the event loop
setInterval(function() {
    Aethium.pumpEvents();
}, 16); // ~60 FPS
`))
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch os.Getenv("GOOS") {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

