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
		return err
	}

	// Render
	if r.backend != nil {
		return r.backend.Render(dl)
	}

	return nil
}

func (r *Runtime) buildFrame(node *scene.VNode, dl *canvas.DrawList, index int) error {
	r.depMu.Lock()
	r.currentNode = node
	r.clearNodeDependencies(node)
	r.depMu.Unlock()

	// Call component Update if dirty
	if node.IsDirty() {
		if comp, ok := node.Component.(Component); ok {
			ctx := UpdateContext{
				Index: index,
				Node:  node,
			}
			if err := comp.Update(ctx); err != nil {
				return err
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
			return err
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
	r.root = nil
	return nil
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
	r.depMu.Lock()
	nodes, ok := r.signalNodes[id]
	r.depMu.Unlock()
	if !ok {
		return
	}

	for _, node := range nodes {
		node.MarkDirty()
	}
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

