package runtime

import (
	"errors"
	"sync"

	"github.com/A-Solo-Engineer/aethium/canvas"
	"github.com/A-Solo-Engineer/aethium/platform"
	"github.com/A-Solo-Engineer/aethium/reactive"
	"github.com/A-Solo-Engineer/aethium/scene"
)

type Component interface {
	Init(ctx InitContext) error
	Mount(ctx MountContext) error
	Update(ctx UpdateContext) error
	Unmount(ctx UnmountContext) error
}

type InitContext struct {
	Props   any
	Runtime *Runtime
}

type MountContext struct {
	Index int
	Node  *scene.VNode
}

type UpdateContext struct {
	Index int
	Dirty []reactive.SignalID
	Node  *scene.VNode
}

type UnmountContext struct {
	Index int
	Node  *scene.VNode
}

type Runtime struct {
	backend           platform.Backend
	uiQueueMu         sync.Mutex
	uiQueue           []func()
	root              *scene.VNode
	frameMu           sync.Mutex
	frame             int
	depMu             sync.Mutex
	currentNode       *scene.VNode
	nodeSignals       map[scene.NodeID]map[reactive.SignalID]struct{}
	signalNodes       map[reactive.SignalID]map[scene.NodeID]*scene.VNode
	signalUnsubscribe func()
}

func NewRuntime(backend platform.Backend) *Runtime {
	r := &Runtime{
		backend:     backend,
		uiQueue:     make([]func(), 0, 64),
		nodeSignals: make(map[scene.NodeID]map[reactive.SignalID]struct{}),
		signalNodes: make(map[reactive.SignalID]map[scene.NodeID]*scene.VNode),
	}
	// Note: Dependency tracking is handled per-tick by pushing/popping
	// the runtime instance onto the reactive tracker stack.
	scene.SetCurrentTracker(r)
	_, unsubscribe := reactive.SubscribeAll(r.handleSignalChange)
	r.signalUnsubscribe = unsubscribe
	return r
}

func (r *Runtime) ScheduleOnUI(fn func()) {
	r.uiQueueMu.Lock()
	r.uiQueue = append(r.uiQueue, fn)
	r.uiQueueMu.Unlock()
}

func (r *Runtime) drainUIQueue() {
	r.uiQueueMu.Lock()
	if len(r.uiQueue) == 0 {
		r.uiQueueMu.Unlock()
		return
	}
	work := r.uiQueue
	r.uiQueue = make([]func(), 0, len(work))
	r.uiQueueMu.Unlock()

	for _, fn := range work {
		fn()
	}
}

func (r *Runtime) Attach(root Component) error {
	if root == nil {
		return errors.New("root component is nil")
	}
	r.root = scene.NewVNode(root)

	if err := root.Init(InitContext{Runtime: r}); err != nil {
		return err
	}
	if err := root.Mount(MountContext{Index: 0, Node: r.root}); err != nil {
		return err
	}

	return nil
}

func (r *Runtime) Tick() error {
	r.frameMu.Lock()
	r.frame++
	r.frameMu.Unlock()

	r.drainUIQueue()

	if r.root == nil {
		return nil
	}

	// Build frame with tracking enabled
	reactive.PushDependencyTracker(r)
	defer reactive.PopDependencyTracker()

	dl := canvas.NewDrawList()
	defer dl.Release()

	if err := r.buildFrame(r.root, dl, 0); err != nil {
		return err
	}

	// Render
	if r.backend != nil {
		return r.backend.Render(dl)
	}

	return nil
}

func (r *Runtime) buildFrame(node *scene.VNode, dl *canvas.DrawList, index int) error {
	if node == nil {
		return nil
	}
	r.depMu.Lock()
	prevNode := r.currentNode
	r.currentNode = node
	r.clearNodeDependencies(node)
	r.depMu.Unlock()

	defer func() {
		r.depMu.Lock()
		r.currentNode = prevNode
		r.depMu.Unlock()
	}()

	// Call component Update if dirty
	if node.IsDirty() {
		if comp, ok := node.Component.(Component); ok {
			ctx := UpdateContext{
				Index: index,
				Dirty: node.DirtySignals,
				Node:  node,
			}
			if err := comp.Update(ctx); err != nil {
				return err
			}
		}
		node.Dirty = false
		node.DirtySignals = node.DirtySignals[:0]
	}

	// Render this node
	if comp, ok := node.Component.(Component); ok {
		if viewFn, ok := comp.(interface{ View() []canvas.DrawCmd }); ok {
			cmds := viewFn.View()
			for _, cmd := range cmds {
				dl.Append(cmd)
			}
		}
	}

	// Recursively build children
	for i, child := range node.Children {
		if err := r.buildFrame(child, dl, index+i); err != nil {
			return err
		}
	}

	r.depMu.Lock()
	// No need to set r.currentNode = nil here, defer handles it
	r.depMu.Unlock()

	return nil
}

func (r *Runtime) Detach(root scene.NodeID) error {
	if r.root == nil || r.root.ID != root {
		return nil
	}
	err := r.unmountNode(r.root)
	r.root = nil
	return err
}

func (r *Runtime) unmountNode(node *scene.VNode) error {
	if node == nil {
		return nil
	}
	var firstErr error
	// Unmount children first (post-order)
	for _, child := range node.Children {
		if err := r.unmountNode(child); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if comp, ok := node.Component.(Component); ok {
		if err := comp.Unmount(UnmountContext{Index: 0, Node: node}); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Remove dependency mappings for this node
	r.clearNodeDependencies(node)

	// Return node to pool
	scene.ReleaseVNode(node)

	return firstErr
}

func (r *Runtime) Track(id reactive.SignalID) {
	r.depMu.Lock()
	defer r.depMu.Unlock()

	if r.currentNode == nil {
		return
	}

	if _, ok := r.nodeSignals[r.currentNode.ID]; !ok {
		r.nodeSignals[r.currentNode.ID] = make(map[reactive.SignalID]struct{})
	}
	r.nodeSignals[r.currentNode.ID][id] = struct{}{}

	if _, ok := r.signalNodes[id]; !ok {
		r.signalNodes[id] = make(map[scene.NodeID]*scene.VNode)
	}
	r.signalNodes[id][r.currentNode.ID] = r.currentNode
}

func (r *Runtime) handleSignalChange(id reactive.SignalID) {
	// Notify must schedule on the UI to avoid data races with Tick().
	r.ScheduleOnUI(func() {
		r.depMu.Lock()
		nodes, ok := r.signalNodes[id]
		if !ok {
			r.depMu.Unlock()
			return
		}

		// Mark nodes dirty while holding the lock to avoid races with Track
		for _, node := range nodes {
			node.MarkDirty(id)
		}
		r.depMu.Unlock()
	})
}

func (r *Runtime) clearNodeDependencies(node *scene.VNode) {
	if node == nil {
		return
	}

	if signalIDs, ok := r.nodeSignals[node.ID]; ok {
		for signalID := range signalIDs {
			if nodes, found := r.signalNodes[signalID]; found {
				delete(nodes, node.ID)
				if len(nodes) == 0 {
					delete(r.signalNodes, signalID)
				}
			}
		}
		delete(r.nodeSignals, node.ID)
	}
}

func (r *Runtime) GetFrame() int {
	r.frameMu.Lock()
	defer r.frameMu.Unlock()
	return r.frame
}

func (r *Runtime) GetRoot() *scene.VNode {
	return r.root
}

func (r *Runtime) SetBackend(backend platform.Backend) {
	r.backend = backend
}
