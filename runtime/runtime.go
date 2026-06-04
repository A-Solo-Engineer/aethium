package runtime

import (
	"errors"
	"github.com/aethium-dev/aethium/canvas"
	"github.com/aethium-dev/aethium/platform"
	"github.com/aethium-dev/aethium/reactive"
	"github.com/aethium-dev/aethium/scene"
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
	InitContext
	Node *scene.VNode
}

type UpdateContext struct {
	MountContext
	Dirty []reactive.SignalID
}

type UnmountContext struct {
	MountContext
}

type Runtime struct {
	backend platform.Backend
	uiQueue chan func()
	root    *scene.VNode
}

func NewRuntime(backend platform.Backend) *Runtime {
	r := &Runtime{
		backend: backend,
		uiQueue: make(chan func(), 64),
	}
	reactive.SetDependencyTracker(r)
	scene.SetCurrentTracker(r)
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
	return nil
}

func (r *Runtime) Tick() error {
	r.drainUIQueue()
	if r.backend != nil {
		dl := canvas.NewDrawList()
		defer dl.Release()
		return r.backend.Render(dl)
	}
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
	// placeholder; actual dirty-tracking logic will be implemented in Stage 2
}
