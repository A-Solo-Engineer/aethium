package canvas

import (
	"testing"
)

func TestDrawListAppend(t *testing.T) {
	dl := NewDrawList()
	defer dl.Release()

	// Test appending a command
	cmd := DrawCmd{Kind: CmdFillRect, X: 10, Y: 20, W: 30, H: 40, Color: 0xFF0000FF}
	dl.Append(cmd)

	if len(dl.Cmds) != 1 {
		t.Errorf("Expected 1 command, got %d", len(dl.Cmds))
	}

	if dl.Cmds[0] != cmd {
		t.Errorf("Command mismatch: expected %+v, got %+v", cmd, dl.Cmds[0])
	}
}

func TestDrawListMerge(t *testing.T) {
	dl1 := NewDrawList()
	defer dl1.Release()
	dl2 := NewDrawList()
	defer dl2.Release()

	// Add commands to both lists
	dl1.Append(DrawCmd{Kind: CmdFillRect, X: 10, Y: 20, W: 30, H: 40, Color: 0xFF0000FF})
	dl2.Append(DrawCmd{Kind: CmdStrokeRect, X: 50, Y: 60, W: 70, H: 80, Color: 0x00FF00FF})

	originalLen1 := len(dl1.Cmds)
	originalLen2 := len(dl2.Cmds)

	Merge(dl1, dl2)

	if len(dl1.Cmds) != originalLen1+originalLen2 {
		t.Errorf("Expected %d commands, got %d", originalLen1+originalLen2, len(dl1.Cmds))
	}

	if len(dl2.Cmds) != 0 {
		t.Errorf("Expected dl2 to be empty after merge, got %d commands", len(dl2.Cmds))
	}
}

func TestDrawListRelease(t *testing.T) {
	dl := NewDrawList()
	dl.Append(DrawCmd{Kind: CmdFillRect, X: 10, Y: 20, W: 30, H: 40, Color: 0xFF0000FF})

	// Release should clear the commands
	dl.Release()

	if len(dl.Cmds) != 0 {
		t.Errorf("Expected 0 commands after release, got %d", len(dl.Cmds))
	}
}

func TestFillRect(t *testing.T) {
	dl := NewDrawList()
	defer dl.Release()

	rect := Rect{X: 10, Y: 20, W: 30, H: 40}
	color := Color(0xFF0000FF)

	FillRect(dl, rect, color)

	if len(dl.Cmds) != 1 {
		t.Errorf("Expected 1 command, got %d", len(dl.Cmds))
	}

	cmd := dl.Cmds[0]
	if cmd.Kind != CmdFillRect {
		t.Errorf("Expected CmdFillRect, got %d", cmd.Kind)
	}
	if cmd.X != rect.X || cmd.Y != rect.Y || cmd.W != rect.W || cmd.H != rect.H {
		t.Errorf("Rect coordinates mismatch: expected %+v, got %+v", rect, Rect{X: cmd.X, Y: cmd.Y, W: cmd.W, H: cmd.H})
	}
	if cmd.Color != color {
		t.Errorf("Color mismatch: expected %d, got %d", color, cmd.Color)
	}
}

func TestDrawText(t *testing.T) {
	dl := NewDrawList()
	defer dl.Release()

	x, y := 10.0, 20.0
	text := "Hello, World!"
	color := Color(0xFFFFFFFF)

	DrawText(dl, x, y, text, color)

	if len(dl.Cmds) != 1 {
		t.Errorf("Expected 1 command, got %d", len(dl.Cmds))
	}

	cmd := dl.Cmds[0]
	if cmd.Kind != CmdDrawText {
		t.Errorf("Expected CmdDrawText, got %d", cmd.Kind)
	}
	if cmd.X != x || cmd.Y != y {
		t.Errorf("Text position mismatch: expected (%f, %f), got (%f, %f)", x, y, cmd.X, cmd.Y)
	}
	if cmd.Text != text {
		t.Errorf("Text mismatch: expected %s, got %s", text, cmd.Text)
	}
	if cmd.Color != color {
		t.Errorf("Color mismatch: expected %d, got %d", color, cmd.Color)
	}
}

func TestDrawTextTruncation(t *testing.T) {
	dl := NewDrawList()
	defer dl.Release()

	// Create a very long text
	longText := ""
	for i := 0; i < 100; i++ {
		longText += "A"
	}

	DrawText(dl, 10, 20, longText, 0xFFFFFFFF)

	if len(dl.Cmds) != 1 {
		t.Errorf("Expected 1 command, got %d", len(dl.Cmds))
	}

	cmd := dl.Cmds[0]
	if len(cmd.Text) > 4096 {
		t.Errorf("Text should be truncated to max 4096 characters, got %d", len(cmd.Text))
	}
}
