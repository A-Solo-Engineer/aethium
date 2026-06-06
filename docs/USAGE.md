# Aethium Usage Guide

## API Cheat Sheet

| Task | Function / Method | Context |
|------|-------------------|---------|
| **Create Signal** | `reactive.NewSignal(val)` | Default |
| **Create Signal** | `rt.Reactive().NewSignal(val)` | Instance |
| **Create Computed** | `reactive.NewComputed(fn)` | Default |
| **Register Effect** | `reactive.NewEffectWithRuntime(rt, fn)` | Instance Required |
| **Schedule UI Task**| `rt.ScheduleOnUI(fn)` | Instance |
| **Attach App** | `rt.Attach(component)` | Instance |

---

## Writing Components

Aethium components implement the `runtime.Component` interface.

```go
type Counter struct {
    count *reactive.Signal[int]
}

func (c *Counter) Init(ctx runtime.InitContext) error {
    // Initialize signals using the runtime's context
    c.count = reactive.NewSignalWithContext(ctx.Runtime.Reactive(), 0)
    return nil
}

func (c *Counter) View() []canvas.DrawCmd {
    count := c.count.Get() // Automatic dependency tracking
    return []canvas.DrawCmd{
        {Kind: canvas.CmdFillRect, X: 10, Y: 10, W: 100, H: 50, Color: 0xFF0000FF},
        {Kind: canvas.CmdDrawText, X: 120, Y: 35, Text: fmt.Sprintf("Count: %d", count), Color: 0xFFFFFFFF},
    }
}
```

---

## Complex State: The Todo Example

Aethium supports slices and complex structures in signals. Use `WithEquality` to optimize performance.

```go
type TodoApp struct {
    todos *reactive.Signal[[]*TodoItem]
}

func (a *TodoApp) Init(ctx runtime.InitContext) error {
    rctx := ctx.Runtime.Reactive()
    a.todos = reactive.NewSignalWithContext(rctx, []*TodoItem{})
    
    // Optional: Custom equality to prevent redundant re-renders
    a.todos.WithEquality(func(a, b []*TodoItem) bool {
        return len(a) == len(b) 
    })
    return nil
}
```

---

## Platform-Specific Deployment

### 1. Browser (Wasm)
Aethium uses TinyGo to produce ultra-small binaries.

```bash
# Build for web
aethium build --target wasm
```

### 2. Desktop (WebView)
Uses the system's native WebView for rendering, keeping binary size minimal.

```bash
# Build for desktop
aethium build --target desktop
```

---

## Advanced: Custom Rendering

You can implement the `canvas.Graphics` interface to create your own rendering backend.

```go
type MyRenderer struct {}

func (r *MyRenderer) FillRect(x, y, w, h float32, color canvas.Color) {
    // Your custom rendering logic here
}

// ... implement other methods
```
