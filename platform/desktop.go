package platform

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/A-Solo-Engineer/aethium/canvas"
)

type DesktopBackend struct {
	Width, Height int
	serverOnce    sync.Once
	mu            sync.Mutex
	lastFrame     []byte
	clients       map[chan []byte]struct{}
}

func NewDesktopBackend(w, h int) *DesktopBackend {
	return &DesktopBackend{
		Width:   w,
		Height:  h,
		clients: make(map[chan []byte]struct{}),
	}
}

func (b *DesktopBackend) Render(dl *canvas.DrawList) error {
	b.serverOnce.Do(func() {
		go b.startServer()
	})

	// Convert DrawList to JSON for the web client
	cmds := make([][]any, len(dl.Cmds))
	for i, cmd := range dl.Cmds {
		cmdData := []any{
			uint8(cmd.Kind),
			cmd.X,
			cmd.Y,
			cmd.W,
			cmd.H,
			uint32(cmd.Color),
			cmd.Text,
		}
		if cmd.Kind == canvas.CmdTransform {
			transform := make([]any, 6)
			for j, v := range cmd.Transform {
				transform[j] = v
			}
			cmdData = append(cmdData, transform)
		}
		cmds[i] = cmdData
	}

	data, err := json.Marshal(cmds)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.lastFrame = data
	for client := range b.clients {
		select {
		case client <- data:
		default:
			// Client slow, skip frame
		}
	}
	b.mu.Unlock()

	return nil
}

func (b *DesktopBackend) startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", b.handleIndex)
	mux.HandleFunc("/events", b.handleEvents)

	addr := "127.0.0.1:5173"
	fmt.Printf("\n--- Aethium Desktop Server ---")
	fmt.Printf("\nStarting GUI at http://%s\n\n", addr)

	go func() {
		time.Sleep(500 * time.Millisecond)
		b.openBrowser("http://" + addr)
	}()

	http.ListenAndServe(addr, mux)
}

func (b *DesktopBackend) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Aethium Desktop</title>
    <style>
        body { margin: 0; background: #000; overflow: hidden; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: sans-serif; }
        canvas { background: #fff; box-shadow: 0 0 20px rgba(0,0,0,0.5); }
    </style>
</head>
<body>
    <canvas id="canvas" width="%d" height="%d"></canvas>
    <script>
        const canvas = document.getElementById('canvas');
        const ctx = canvas.getContext('2d');

        const Aethium = {
            parseColor: function(color) {
                const a = ((color >> 24) & 0xFF) / 255;
                const r = (color >> 16) & 0xFF;
                const g = (color >> 8) & 0xFF;
                const b = color & 0xFF;
                return "rgba(" + r + "," + g + "," + b + "," + a + ")";
            },
            renderFrame: function(cmds) {
                ctx.clearRect(0, 0, canvas.width, canvas.height);
                for (const cmd of cmds) {
                    const kind = cmd[0];
                    switch (kind) {
                        case 1: // FillRect
                            ctx.fillStyle = this.parseColor(cmd[5]);
                            ctx.fillRect(cmd[1], cmd[2], cmd[3], cmd[4]);
                            break;
                        case 2: // StrokeRect
                            ctx.strokeStyle = this.parseColor(cmd[5]);
                            ctx.strokeRect(cmd[1], cmd[2], cmd[3], cmd[4]);
                            break;
                        case 3: // DrawText
                            ctx.fillStyle = this.parseColor(cmd[5]);
                            ctx.font = '16px sans-serif';
                            ctx.fillText(cmd[6], cmd[1], cmd[2]);
                            break;
                        case 4: // Clip
                            ctx.beginPath();
                            ctx.rect(cmd[1], cmd[2], cmd[3], cmd[4]);
                            ctx.clip();
                            break;
                    }
                }
            }
        };

        const ev = new EventSource('/events');
        ev.onmessage = (e) => {
            const cmds = JSON.parse(e.data);
            requestAnimationFrame(() => Aethium.renderFrame(cmds));
        };
    </script>
</body>
</html>`, b.Width, b.Height)
}

func (b *DesktopBackend) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 1)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	if b.lastFrame != nil {
		ch <- b.lastFrame
	}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
	}()

	for {
		select {
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

func (b *DesktopBackend) openBrowser(url string) {
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
