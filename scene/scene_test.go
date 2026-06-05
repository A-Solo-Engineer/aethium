package scene

import (
	"testing"
)

func TestVNodeCreation(t *testing.T) {
	comp := "test-component"
	vnode := NewVNode(comp)

	if vnode.Component != comp {
		t.Errorf("Expected component %s, got %s", comp, vnode.Component)
	}
	if vnode.ID == 0 {
		t.Error("Expected non-zero VNode ID")
	}
	if len(vnode.Children) != 0 {
		t.Errorf("Expected no children, got %d", len(vnode.Children))
	}
	if vnode.Parent != nil {
		t.Error("Expected no parent")
	}
}

func TestVNodeParentChild(t *testing.T) {
	parent := NewVNode("parent")
	child1 := NewVNode("child1")
	child2 := NewVNode("child2")

	parent.AddChild(child1)
	parent.AddChild(child2)

	if len(parent.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(parent.Children))
	}

	if parent.Children[0] != child1 || parent.Children[1] != child2 {
		t.Error("Children not added in correct order")
	}

	if child1.Parent != parent || child2.Parent != parent {
		t.Error("Children should reference parent")
	}
}

func TestVNodeRemoveChild(t *testing.T) {
	parent := NewVNode("parent")
	child1 := NewVNode("child1")
	child2 := NewVNode("child2")

	parent.AddChild(child1)
	parent.AddChild(child2)

	parent.RemoveChild(child1)

	if len(parent.Children) != 1 {
		t.Errorf("Expected 1 child after removal, got %d", len(parent.Children))
	}

	if parent.Children[0] != child2 {
		t.Error("Wrong child removed")
	}

	if child1.Parent != nil {
		t.Error("Removed child should not reference parent")
	}
}

func TestVNodeDirty(t *testing.T) {
	vnode := NewVNode("test")

	if vnode.IsDirty() {
		t.Error("New VNode should not be dirty")
	}

	vnode.MarkDirty()

	if !vnode.IsDirty() {
		t.Error("VNode should be dirty after MarkDirty")
	}

	vnode.Dirty = false

	if vnode.IsDirty() {
		t.Error("VNode should not be dirty after Dirty = false")
	}
}

func TestVNodeRelease(t *testing.T) {
	vnode := NewVNode("test")
	vnode.MarkDirty()

	ReleaseVNode(vnode)

	if vnode.Component != nil {
		t.Error("Released VNode should have nil component")
	}
	if vnode.Dirty {
		t.Error("Released VNode should not be dirty")
	}
	if len(vnode.Children) != 0 {
		t.Errorf("Released VNode should have no children, got %d", len(vnode.Children))
	}
	if vnode.Parent != nil {
		t.Error("Released VNode should have no parent")
	}
}

func TestVNodeFindChild(t *testing.T) {
	parent := NewVNode("parent")
	child1 := NewVNode("child1")
	child2 := NewVNode("child2")

	parent.AddChild(child1)
	parent.AddChild(child2)

	found := parent.FindChild(func(child *VNode) bool {
		return child.Component == "child1"
	})

	if found != child1 {
		t.Error("Should find child1")
	}

	notFound := parent.FindChild(func(child *VNode) bool {
		return child.Component == "nonexistent"
	})

	if notFound != nil {
		t.Error("Should not find nonexistent child")
	}
}

func TestVNodeGetDescendants(t *testing.T) {
	root := NewVNode("root")
	child1 := NewVNode("child1")
	child2 := NewVNode("child2")
	grandchild1 := NewVNode("grandchild1")
	grandchild2 := NewVNode("grandchild2")

	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(grandchild1)
	child1.AddChild(grandchild2)

	descendants := root.GetDescendants()

	if len(descendants) != 4 {
		t.Errorf("Expected 4 descendants, got %d", len(descendants))
	}

	// Check that all descendants are included
	found := make(map[string]bool)
	for _, desc := range descendants {
		found[desc.Component.(string)] = true
	}

	expected := []string{"child1", "child2", "grandchild1", "grandchild2"}
	for _, comp := range expected {
		if !found[comp] {
			t.Errorf("Missing descendant: %s", comp)
		}
	}
}
