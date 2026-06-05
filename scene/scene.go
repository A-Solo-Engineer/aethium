package scene

import (
	"sync"

	"github.com/A-Solo-Engineer/aethium/reactive"
)

type NodeID uint64

type DependencyTracker interface {
	Track(reactive.SignalID)
}

type VNode struct {
	ID           NodeID
	Component    any
	Children     []*VNode
	Signals      []reactive.SignalID
	Dirty        bool
	DirtySignals []reactive.SignalID
	Parent       *VNode
}

var vnodePool = sync.Pool{
	New: func() any { return &VNode{} },
}

type MountContext struct {
	Index int
}

type UpdateContext struct {
	Index int
	Dirty []reactive.SignalID
}

type Context struct {
	currentTracker DependencyTracker
	trackerMu      sync.RWMutex

	nodeSeq   NodeID
	nodeSeqMu sync.Mutex
}

func NewContext() *Context {
	return &Context{}
}

var (
	defaultContext = NewContext()
)

func (c *Context) SetCurrentTracker(t DependencyTracker) {
	c.trackerMu.Lock()
	c.currentTracker = t
	c.trackerMu.Unlock()
}

func (c *Context) CurrentTracker() DependencyTracker {
	c.trackerMu.RLock()
	defer c.trackerMu.RUnlock()
	return c.currentTracker
}

func (c *Context) newNodeID() NodeID {
	c.nodeSeqMu.Lock()
	c.nodeSeq++
	id := c.nodeSeq
	c.nodeSeqMu.Unlock()
	return id
}

func (c *Context) NewVNode(component any) *VNode {
	n := vnodePool.Get().(*VNode)
	n.Reset()
	n.ID = c.newNodeID()
	n.Component = component
	return n
}

func SetCurrentTracker(t DependencyTracker) {
	defaultContext.SetCurrentTracker(t)
}

func CurrentTracker() DependencyTracker {
	return defaultContext.CurrentTracker()
}

func NewVNode(component any) *VNode {
	return defaultContext.NewVNode(component)
}

func ReleaseVNode(n *VNode) {
	if n == nil {
		return
	}
	n.Reset()
	vnodePool.Put(n)
}

func (n *VNode) Reset() {
	// Clear slices but keep capacity for small sizes to avoid reallocations.
	// For large slices, we nil them out to let GC reclaim memory.
	if cap(n.Children) > 32 {
		n.Children = nil
	} else {
		n.Children = n.Children[:0]
	}
	if cap(n.Signals) > 64 {
		n.Signals = nil
	} else {
		n.Signals = n.Signals[:0]
	}
	if cap(n.DirtySignals) > 64 {
		n.DirtySignals = nil
	} else {
		n.DirtySignals = n.DirtySignals[:0]
	}
	n.Dirty = false
	n.Parent = nil
}

func (n *VNode) AddChild(child *VNode) {
	n.Children = append(n.Children, child)
	child.Parent = n
}

func (n *VNode) AddSignal(id reactive.SignalID) {
	n.Signals = append(n.Signals, id)
}

func (n *VNode) MarkDirty(sid ...reactive.SignalID) {
	n.Dirty = true
	if len(sid) > 0 {
		n.DirtySignals = append(n.DirtySignals, sid...)
	}
	if n.Parent != nil {
		n.Parent.MarkDirty()
	}
}

func (n *VNode) IsDirty() bool {
	return n.Dirty
}

func (n *VNode) ClearDirty() {
	n.Dirty = false
	n.DirtySignals = n.DirtySignals[:0]
	for _, child := range n.Children {
		child.ClearDirty()
	}
}

func (n *VNode) Count() int {
	count := 1
	for _, child := range n.Children {
		count += child.Count()
	}
	return count
}

func (n *VNode) FindNode(id NodeID) *VNode {
	if n.ID == id {
		return n
	}
	for _, child := range n.Children {
		if found := child.FindNode(id); found != nil {
			return found
		}
	}
	return nil
}
