package scene

import (
	"testing"
)

func TestVNode_Hierarchy(t *testing.T) {
	root := NewVNode("root")
	child := NewVNode("child")
	root.AddChild(child)

	if len(root.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(root.Children))
	}

	if child.Parent != root {
		t.Error("child parent not set correctly")
	}

	if root.Count() != 2 {
		t.Errorf("expected count 2, got %d", root.Count())
	}
}

func TestVNode_Dirty(t *testing.T) {
	root := NewVNode("root")
	child := NewVNode("child")
	root.AddChild(child)

	child.MarkDirty(123)

	if !child.IsDirty() {
		t.Error("child should be dirty")
	}

	if !root.IsDirty() {
		t.Error("root should be dirty (propagation)")
	}

	if len(child.DirtySignals) != 1 || child.DirtySignals[0] != 123 {
		t.Errorf("expected dirty signal 123, got %v", child.DirtySignals)
	}

	root.ClearDirty()

	if root.IsDirty() || child.IsDirty() {
		t.Error("nodes should not be dirty after ClearDirty")
	}
}

func TestVNode_FindNode(t *testing.T) {
	root := NewVNode("root")
	child := NewVNode("child")
	root.AddChild(child)

	found := root.FindNode(child.ID)
	if found != child {
		t.Error("failed to find child node by ID")
	}
}

func TestVNode_PoolReset(t *testing.T) {
	node := NewVNode("comp")
	node.MarkDirty(123)
	node.AddSignal(456)
	
	id := node.ID
	ReleaseVNode(node)
	
	// Get new node from pool (might be the same one)
	newNode := NewVNode("newcomp")
	
	if newNode.IsDirty() {
		t.Error("newly pooled node should not be dirty")
	}
	if len(newNode.DirtySignals) != 0 {
		t.Errorf("newly pooled node should have 0 dirty signals, got %v", newNode.DirtySignals)
	}
	if len(newNode.Signals) != 0 {
		t.Errorf("newly pooled node should have 0 signals, got %v", newNode.Signals)
	}
	if newNode.ID == id {
		// If it's the same node, it should have a new ID
		t.Error("newly pooled node should have a new ID")
	}
}
