package app

import (
	"testing"
	"github.com/A-Solo-Engineer/aethium/runtime"
)

type mockBackend struct{}
func (m *mockBackend) Render(dl *any) error { return nil }
func (m *mockBackend) FillRect(x, y, w, h float32, color uint32) {}
func (m *mockBackend) StrokeRect(x, y, w, h float32, color uint32) {}
func (m *mockBackend) DrawText(x, y float32, text string, color uint32) {}
func (m *mockBackend) SetClip(x, y, w, h float32) {}
func (m *mockBackend) SetTransform(matrix [6]float32) {}

func TestTodoApp_AddTodo(t *testing.T) {
	rt := runtime.NewRuntime(nil)
	app := NewTodoApp()
	
	if err := rt.Attach(app); err != nil {
		t.Fatalf("Failed to attach app: %v", err)
	}
	
	initialCount := len(app.todos.Get())
	if initialCount != 2 {
		t.Errorf("Expected 2 initial todos, got %d", initialCount)
	}
	
	app.AddTodo("Test Task")
	
	newCount := len(app.todos.Get())
	if newCount != 3 {
		t.Errorf("Expected 3 todos after addition, got %d", newCount)
	}
	
	lastTodo := app.todos.Get()[2]
	if lastTodo.Text.Get() != "Test Task" {
		t.Errorf("Expected 'Test Task', got '%s'", lastTodo.Text.Get())
	}
}
