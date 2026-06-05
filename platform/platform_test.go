package platform

import (
	"testing"

	"github.com/A-Solo-Engineer/aethium/canvas"
)

// Mock backend for testing
type MockBackend struct {
	RenderCount  int
	LastDrawList *canvas.DrawList
}

func (m *MockBackend) Render(dl *canvas.DrawList) error {
	m.RenderCount++
	m.LastDrawList = dl
	return nil
}

func TestBackendInterface(t *testing.T) {
	var backend Backend = &MockBackend{}

	// Test that the interface is satisfied
	if backend == nil {
		t.Error("Backend should not be nil")
	}
}

func TestMockBackendRender(t *testing.T) {
	backend := &MockBackend{}
	dl := canvas.NewDrawList()
	defer dl.Release()

	// Add a command
	dl.Append(canvas.DrawCmd{Kind: canvas.CmdFillRect, X: 10, Y: 20, W: 30, H: 40, Color: 0xFF0000FF})

	// Render
	err := backend.Render(dl)

	if err != nil {
		t.Errorf("Render should not return error: %v", err)
	}

	if backend.RenderCount != 1 {
		t.Errorf("Expected 1 render call, got %d", backend.RenderCount)
	}

	if backend.LastDrawList != dl {
		t.Error("LastDrawList should reference the draw list that was rendered")
	}
}

func TestMockBackendMultipleRenders(t *testing.T) {
	backend := &MockBackend{}

	// Render multiple times
	for i := 0; i < 5; i++ {
		dl := canvas.NewDrawList()
		dl.Append(canvas.DrawCmd{Kind: canvas.CmdFillRect, X: float32(i), Y: float32(i), W: 10, H: 10, Color: 0xFF0000FF})
		backend.Render(dl)
		dl.Release()
	}

	if backend.RenderCount != 5 {
		t.Errorf("Expected 5 render calls, got %d", backend.RenderCount)
	}
}

func TestMockBackendNilDrawList(t *testing.T) {
	backend := &MockBackend{}

	// Render with nil draw list
	err := backend.Render(nil)

	if err != nil {
		t.Errorf("Render should not return error for nil draw list: %v", err)
	}

	if backend.RenderCount != 1 {
		t.Errorf("Expected 1 render call, got %d", backend.RenderCount)
	}

	if backend.LastDrawList != nil {
		t.Error("LastDrawList should be nil when rendering nil draw list")
	}
}
