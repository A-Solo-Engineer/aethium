package runtime

import (
	"errors"
	"fmt"
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
	uiQueue           chan func()
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
		uiQueue:     make(chan func(), 64),
		nodeSignals: make(map[scene.NodeID]map[reactive.SignalID]struct{}),
		signalNodes: make(map[reactive.SignalID]map[scene.NodeID]*scene.VNode),
	}
	reactive.SetDependencyTracker(r)
	scene.SetCurrentTracker(r)
	_, unsubscribe := reactive.SubscribeAll(r.handleSignalChange)
	r.signalUnsubscribe = unsubscribe
	return r
}

func (r *Runtime) ScheduleOnUI(fn func()) {
	select {
	case r.uiQueue <- fn:
	default:
		go func() {
			r.uiQueue <- fn
		}()
	}
}

func (r *Runtime) drainUIQueue() {
	for {
		select {
		case fn := <-r.uiQueue:
			fn()
		default:
			return
		}
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
	if r == nil {
		return errors.New("runtime is nil")
	}

	r.frameMu.Lock()
	r.frame++
	r.frameMu.Unlock()

	r.drainUIQueue()

	if r.root == nil {
		return nil
	}

	// Build frame
	dl := canvas.NewDrawList()
	defer dl.Release()

	if err := r.buildFrame(r.root, dl, 0); err != nil {
		return fmt.Errorf("buildFrame failed: %w", err)
	}

	// Render
	if r.backend != nil {
		if err := r.backend.Render(dl); err != nil {
			return fmt.Errorf("backend.Render failed: %w", err)
		}
	}

	return nil
}

func (r *Runtime) buildFrame(node *scene.VNode, dl *canvas.DrawList, index int) error {
	if node == nil {
		return errors.New("node is nil")
	}

	r.depMu.Lock()
	r.currentNode = node
	r.clearNodeDependencies(node)
	// depMu must be unlocked before calling component methods;
	// View() calls signal.Get() which calls Track() which acquires depMu.
	r.depMu.Unlock()

	// Call component Update if dirty
	if node.IsDirty() {
		if comp, ok := node.Component.(Component); ok {
			ctx := UpdateContext{
				Index: index,
				Node:  node,
				Dirty: r.getNodeDirtySignals(node),
			}
			if err := comp.Update(ctx); err != nil {
				return fmt.Errorf("component.Update failed: %w", err)
			}
		}
		node.Dirty = false
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
			return fmt.Errorf("buildFrame child failed: %w", err)
		}
	}

	r.depMu.Lock()
	r.currentNode = nil
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
		r.depMu.Unlock()
		if !ok {
			return
		}

		for _, node := range nodes {
			node.MarkDirty()
		}
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

func (r *Runtime) getNodeDirtySignals(node *scene.VNode) []reactive.SignalID {
	r.depMu.Lock()
	defer r.depMu.Unlock()

	if signals, ok := r.nodeSignals[node.ID]; ok {
		dirtySignals := make([]reactive.SignalID, 0, len(signals))
		for signalID := range signals {
			dirtySignals = append(dirtySignals, signalID)
		}
		return dirtySignals
	}
	return nil
}

