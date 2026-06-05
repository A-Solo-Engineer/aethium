# Aethium Usage Guide

## Project Structure

```
myapp/
├── go.mod
├── main.go
├── app/
│   ├── view.go
│   └── state.go
└── aethium.toml
```

## Creating a New Project

```bash
aethium new --module example.com/myapp --dir myapp
```

This creates:
- `go.mod` with your module path
- `main.go` entry point
- `app/view.go` for components
- `app/state.go` for signals
- `aethium.toml` configuration

## Writing Components

```go
package app

import (
    "github.com/A-Solo-Engineer/aethium/canvas"
    "github.com/A-Solo-Engineer/aethium/runtime"
    "github.com/A-Solo-Engineer/aethium/reactive"
)

type Counter struct {
    count *reactive.Signal[int]
}

func NewCounter() *Counter {
    return &Counter{
        count: reactive.NewSignal(0),
    }
}

func (c *Counter) Init(ctx runtime.InitContext) error {
    return nil
}

func (c *Counter) Mount(ctx runtime.MountContext) error {
    return nil
}

func (c *Counter) Update(ctx runtime.UpdateContext) error {
    return nil
}

func (c *Counter) Unmount(ctx runtime.UnmountContext) error {
    return nil
}

func (c *Counter) View() []canvas.DrawCmd {
    count := c.count.Get()
    return []canvas.DrawCmd{
        canvas.FillRect(canvas.Rect{X: 10, Y: 10, W: 100, H: 50}, 0xFF0000FF),
        canvas.DrawText(120, 35, fmt.Sprintf("Count: %d", count), 0xFFFFFFFF),
    }
}
```

## Using Signals

```go
// Create a signal
count := reactive.NewSignal(0)

// Read value
value := count.Get()

// Write value
count.Set(42)

// Subscribe to changes
id, unsubscribe := reactive.Subscribe(func() {
    fmt.Println("Signal changed!")
})
defer unsubscribe()
```

## Using Computed Values

```go
total := reactive.NewComputed(
    func() []reactive.SignalID {
        return []reactive.SignalID{signal1.ID(), signal2.ID()}
    },
    func() int {
        return signal1.Get() + signal2.Get()
    },
)
```

## Using Effects

```go
effectID := reactive.NewEffect(func(ctx reactive.EffectContext) {
    fmt.Println("Effect running")
})
```

## Building

### Desktop

```bash
aethium build --target desktop --output dist/
```

Produces: `dist/myapp.exe`

### Browser (Wasm)

```bash
aethium build --target wasm --output dist/
```

Produces:
- `dist/app.wasm`
- `dist/index.html`
- `dist/aethium.js`

## Development Server

```bash
aethium dev --target wasm --open
```

Hot-reload enabled for Wasm builds.

## Configuration

`aethium.toml`:

```toml
[build]
target = "wasm"
tinygo_opt = 0

[dev]
addr = "127.0.0.1:5173"
open = true
```

## Platform-Specific Notes

### Browser

- Uses TinyGo 0.34.0
- WebGL2 canvas
- Single-threaded execution
- No DOM manipulation

### Desktop

- Uses Go 1.22+
- System WebView (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux)
- Same rendering pipeline as browser
