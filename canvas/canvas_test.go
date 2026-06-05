package canvas

import (
	"testing"
)

func TestDrawList_Append(t *testing.T) {
	dl := NewDrawList()
	defer dl.Release()

	cmd := DrawCmd{Kind: CmdFillRect, X: 0, Y: 0, W: 100, H: 100, Color: 0xFF0000FF}
	dl.Append(cmd)

	if dl.Count() != 1 {
		t.Errorf("expected 1 command, got %d", dl.Count())
	}

	if dl.Cmds[0].Kind != CmdFillRect {
		t.Errorf("expected CmdFillRect, got %v", dl.Cmds[0].Kind)
	}
}

func TestDrawList_AppendSlice(t *testing.T) {
	dl := NewDrawList()
	defer dl.Release()

	cmds := []DrawCmd{
		{Kind: CmdFillRect, X: 0, Y: 0, W: 100, H: 100},
		{Kind: CmdDrawText, X: 10, Y: 10, Text: "Hello"},
	}
	dl.AppendSlice(cmds)

	if dl.Count() != 2 {
		t.Errorf("expected 2 commands, got %d", dl.Count())
	}
}

func TestDrawList_Reset(t *testing.T) {
	dl := NewDrawList()
	defer dl.Release()

	dl.Append(DrawCmd{Kind: CmdFillRect})
	dl.Reset()

	if dl.Count() != 0 {
		t.Errorf("expected 0 commands after reset, got %d", dl.Count())
	}
}

func TestMerge(t *testing.T) {
	dl1 := NewDrawList()
	defer dl1.Release()
	dl1.Append(DrawCmd{Kind: CmdFillRect})

	dl2 := NewDrawList()
	dl2.Append(DrawCmd{Kind: CmdDrawText})

	Merge(dl1, dl2)

	if dl1.Count() != 2 {
		t.Errorf("expected 2 commands after merge, got %d", dl1.Count())
	}
}
