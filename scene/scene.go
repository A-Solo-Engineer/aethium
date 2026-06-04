package scene

import "github.com/aethium-dev/aethium/reactive"

type NodeID uint64

type DependencyTracker interface {
	Track(reactive.SignalID)
}

type VNode struct {
	ID        NodeID
	Component any
	Children  []*VNode
	Signals   []reactive.SignalID
}

var currentTracker DependencyTracker

func SetCurrentTracker(t DependencyTracker) {
	currentTracker = t
}

func CurrentTracker() DependencyTracker {
	return currentTracker
}

func NewVNode(component any) *VNode {
	return &VNode{Component: component}
}
