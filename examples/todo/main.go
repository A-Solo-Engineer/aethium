package main

import (
	"fmt"
	"time"

	"github.com/A-Solo-Engineer/aethium/examples/todo/app"
	"github.com/A-Solo-Engineer/aethium/platform"
	"github.com/A-Solo-Engineer/aethium/runtime"
)

func main() {
	fmt.Println("Aethium Todo Example")

	// Create desktop backend
	backend := platform.NewDesktopBackend(400, 600)

	// Create runtime
	rt := runtime.NewRuntime(backend)

	// Create and attach Todo app
	todoApp := app.NewTodoApp()
	if err := rt.Attach(todoApp); err != nil {
		fmt.Printf("Error attaching app: %v\n", err)
		return
	}

	fmt.Println("Starting render loop...")
	fmt.Println("Press Ctrl+C to stop.")

	// Simulate some interaction for testing
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("Simulating: Adding a new task...")
		todoApp.AddTodo("Verification complete")
	}()

	for {
		if err := rt.Tick(); err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}
		time.Sleep(16 * time.Millisecond)
	}
}
