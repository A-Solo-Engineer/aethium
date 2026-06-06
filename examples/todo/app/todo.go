package app

import (
	"fmt"
	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/reactive"
	"github.com/A-Solo-Engineer/aethium/runtime"
)

type TodoItem struct {
	ID        int
	Text      *reactive.Signal[string]
	Completed *reactive.Signal[bool]
}

type TodoApp struct {
	todos *reactive.Signal[[]*TodoItem]
	input *reactive.Signal[string]
	ctx   runtime.InitContext
}

func NewTodoApp() *TodoApp {
	return &TodoApp{}
}

func (a *TodoApp) Init(ctx runtime.InitContext) error {
	a.ctx = ctx
	a.todos = reactive.NewSignalWithContext(ctx.Runtime.Reactive(), []*TodoItem{})
	a.input = reactive.NewSignalWithContext(ctx.Runtime.Reactive(), "")
	
	// Add some initial items
	a.AddTodo("Learn Aethium")
	a.AddTodo("Build a Todo App")
	return nil
}

func (a *TodoApp) AddTodo(text string) {
	if text == "" {
		return
	}
	
	rctx := a.ctx.Runtime.Reactive()
	item := &TodoItem{
		ID:        len(a.todos.Peek()),
		Text:      reactive.NewSignalWithContext(rctx, text),
		Completed: reactive.NewSignalWithContext(rctx, false),
	}
	
	current := a.todos.Peek()
	a.todos.Set(append(current, item))
}

func (a *TodoApp) Mount(ctx runtime.MountContext) error {
	return nil
}

func (a *TodoApp) Update(ctx runtime.UpdateContext) error {
	return nil
}

func (a *TodoApp) Unmount(ctx runtime.UnmountContext) error {
	return nil
}

func (a *TodoApp) View() []canvas.DrawCmd {
	cmds := []canvas.DrawCmd{
		{Kind: canvas.CmdFillRect, X: 0, Y: 0, W: 400, H: 600, Color: 0xFFF0F0F0}, // Background
		{Kind: canvas.CmdDrawText, X: 20, Y: 40, Text: "Aethium Todo List", Color: 0xFF333333},
	}

	y := float32(80)
	todos := a.todos.Get()
	
	for _, item := range todos {
		text := item.Text.Get()
		completed := item.Completed.Get()
		
		color := canvas.Color(0xFFFFFFFF)
		textColor := canvas.Color(0xFF000000)
		if completed {
			color = 0xFFD0D0D0
			textColor = 0xFF888888
		}
		
		cmds = append(cmds, canvas.DrawCmd{
			Kind: canvas.CmdFillRect, X: 20, Y: y, W: 360, H: 40, Color: color,
		})
		
		status := "[ ]"
		if completed {
			status = "[x]"
		}
		
		cmds = append(cmds, canvas.DrawCmd{
			Kind: canvas.CmdDrawText, X: 30, Y: y + 25, Text: fmt.Sprintf("%s %s", status, text), Color: textColor,
		})
		
		y += 50
	}
	
	return cmds
}
