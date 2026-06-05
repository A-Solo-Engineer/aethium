package app

import (
	"fmt"
	"time"

	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/reactive"
	"github.com/A-Solo-Engineer/aethium/runtime"
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
	fmt.Println("Counter initialized")
	return nil
}

func (c *Counter) Mount(ctx runtime.MountContext) error {
	fmt.Println("Counter mounted")
	return nil
}

func (c *Counter) Update(ctx runtime.UpdateContext) error {
	// Update logic would go here
	return nil
}

func (c *Counter) Unmount(ctx runtime.UnmountContext) error {
	fmt.Println("Counter unmounted")
	return nil
}

func (c *Counter) View() []canvas.DrawCmd {
	count := c.count.Get()
	return []canvas.DrawCmd{
		{Kind: canvas.CmdFillRect, X: 10, Y: 10, W: 100, H: 50, Color: 0xFF0000FF},
		{Kind: canvas.CmdDrawText, X: 120, Y: 35, Text: fmt.Sprintf("Count: %d", count), Color: 0xFFFFFFFF},
	}
}

func Run() {
	fmt.Println("Aethium hello example starting...")

	// Create runtime
	runtime := runtime.NewRuntime(nil)

	// Create counter component
	counter := NewCounter()

	// Attach to runtime
	if err := runtime.Attach(counter); err != nil {
		fmt.Printf("Error attaching counter: %v\n", err)
		return
	}

	fmt.Println("Starting render loop...")
	frame := 0
	for {
		// Update counter every 60 frames
		if frame%60 == 0 {
			count := counter.count.Get()
			newCount := count + 1
			counter.count.Set(newCount)
			fmt.Printf("Frame %d: Count = %d\n", frame, newCount)
		}

		// Render frame
		if err := runtime.Tick(); err != nil {
			fmt.Printf("Error rendering frame: %v\n", err)
			break
		}

		frame++
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}
}

func RunDev() {
	fmt.Println("Starting Aethium dev server...")

	// In dev mode, we'd start the dev server
	// For now, just run the app
	Run()
}

