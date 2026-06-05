package runtime_test

import (
	"testing"

	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/reactive"
	"github.com/A-Solo-Engineer/aethium/runtime"
)

type noopBackend struct{}

func (n *noopBackend) Render(dl *canvas.DrawList) error { return nil }

type benchComponent struct {
	count *reactive.Signal[int]
}

func (c *benchComponent) Init(ctx runtime.InitContext) error       { return nil }
func (c *benchComponent) Mount(ctx runtime.MountContext) error     { return nil }
func (c *benchComponent) Update(ctx runtime.UpdateContext) error   { return nil }
func (c *benchComponent) Unmount(ctx runtime.UnmountContext) error { return nil }

func (c *benchComponent) View() []canvas.DrawCmd {
	_ = c.count.Get() // registers dependency
	return []canvas.DrawCmd{
		{Kind: canvas.CmdFillRect, X: 0, Y: 0, W: 100, H: 50, Color: 0xFF0000FF},
	}
}

func TestRuntime_Attach(t *testing.T) {
	rt := runtime.NewRuntime(&noopBackend{})
	comp := &benchComponent{count: reactive.NewSignal(0)}

	if err := rt.Attach(comp); err != nil {
		t.Fatalf("Attach failed: %v", err)
	}

	if rt.GetRoot() == nil {
		t.Error("root node should not be nil after Attach")
	}
}

func TestRuntime_Tick(t *testing.T) {
	backend := &mockBackend{}
	rt := runtime.NewRuntime(backend)
	sig := reactive.NewSignal(0)
	comp := &benchComponent{count: sig}

	rt.Attach(comp)

	// Initial tick
	if err := rt.Tick(); err != nil {
		t.Fatalf("Tick failed: %v", err)
	}

	if backend.renderCount != 1 {
		t.Errorf("expected 1 render call, got %d", backend.renderCount)
	}

	// Update signal and tick
	sig.Set(1)
	// handleSignalChange schedules on UI, so we need to wait or drain
	// In tests, we can just call Tick which drains the queue
	if err := rt.Tick(); err != nil {
		t.Fatalf("Tick failed: %v", err)
	}

	if backend.renderCount != 2 {
		t.Errorf("expected 2 render calls, got %d", backend.renderCount)
	}
}

type mockBackend struct {
	renderCount int
}

func (m *mockBackend) Render(dl *canvas.DrawList) error {
	m.renderCount++
	return nil
}

func BenchmarkTick(b *testing.B) {
	b.ReportAllocs()

	rt := runtime.NewRuntime(&noopBackend{})
	comp := &benchComponent{count: reactive.NewSignal(0)}

	if err := rt.Attach(comp); err != nil {
		b.Fatalf("Attach failed: %v", err)
	}

	// Warm up: one tick to register dependencies
	if err := rt.Tick(); err != nil {
		b.Fatalf("warmup Tick failed: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Mark signal dirty before each iteration
		comp.count.Set(0)
		if err := rt.Tick(); err != nil {
			b.Fatalf("Tick failed: %v", err)
		}
	}
}
