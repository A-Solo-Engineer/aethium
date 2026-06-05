package reactive

import (
	"context"
	"sync"
	"sync/atomic"
)

type SignalID uint64
type SubscriberID uint64
type EffectID uint64

type DependencyTracker interface {
	Track(SignalID)
}

type EffectRuntime interface {
	ScheduleOnUI(fn func())
}

type EffectContext struct {
	Ctx     context.Context
	Runtime EffectRuntime
}

type Signal[T comparable] struct {
	ctx   *Context
	mu    sync.RWMutex
	id    SignalID
	value T
}

type Computed[T comparable] struct {
	ctx         *Context
	id          SignalID
	compute     func() T
	cached      T
	mu          sync.RWMutex
	dirty       bool
	unsubscribe func()
}

type Effect struct {
	ctx    *Context
	id     EffectID
	fn     func(ctx EffectContext)
	mu     sync.RWMutex
	active bool
	unsub  func()
}

type Context struct {
	trackerStack   []DependencyTracker
	trackerStackMu sync.Mutex

	subscriberMu sync.Mutex
	subscribers  map[SubscriberID]func()

	signalSubscriberMu sync.Mutex
	signalSubscribers  map[SignalID]map[SubscriberID]func(SignalID)
	subscriberSignals  map[SubscriberID][]SignalID

	effectMu sync.Mutex
	effects  map[EffectID]*Effect

	signalSeq atomic.Uint64
	subSeq    atomic.Uint64
	effectSeq atomic.Uint64
}

func NewContext() *Context {
	return &Context{
		subscribers:       make(map[SubscriberID]func()),
		signalSubscribers: make(map[SignalID]map[SubscriberID]func(SignalID)),
		subscriberSignals: make(map[SubscriberID][]SignalID),
		effects:           make(map[EffectID]*Effect),
	}
}

var (
	defaultContext = NewContext()
)

func (c *Context) PushDependencyTracker(t DependencyTracker) {
	c.trackerStackMu.Lock()
	c.trackerStack = append(c.trackerStack, t)
	c.trackerStackMu.Unlock()
}

func (c *Context) PopDependencyTracker() {
	c.trackerStackMu.Lock()
	if len(c.trackerStack) > 0 {
		c.trackerStack = c.trackerStack[:len(c.trackerStack)-1]
	}
	c.trackerStackMu.Unlock()
}

func (c *Context) CurrentTracker() DependencyTracker {
	c.trackerStackMu.Lock()
	defer c.trackerStackMu.Unlock()
	if len(c.trackerStack) == 0 {
		return nil
	}
	return c.trackerStack[len(c.trackerStack)-1]
}

func (c *Context) newSignalID() SignalID {
	return SignalID(c.signalSeq.Add(1))
}

func (c *Context) newSubscriberID() SubscriberID {
	return SubscriberID(c.subSeq.Add(1))
}

func (c *Context) newEffectID() EffectID {
	return EffectID(c.effectSeq.Add(1))
}

func NewSignal[T comparable](initial T) *Signal[T] {
	return NewSignalWithContext(defaultContext, initial)
}

func NewSignalWithContext[T comparable](c *Context, initial T) *Signal[T] {
	return &Signal[T]{ctx: c, id: c.newSignalID(), value: initial}
}

func PushDependencyTracker(t DependencyTracker) {
	defaultContext.PushDependencyTracker(t)
}

func PopDependencyTracker() {
	defaultContext.PopDependencyTracker()
}

func CurrentTracker() DependencyTracker {
	return defaultContext.CurrentTracker()
}

func (s *Signal[T]) ID() SignalID {
	return s.id
}

func (s *Signal[T]) Get() T {
	s.mu.RLock()
	value := s.value
	s.mu.RUnlock()

	if tracker := s.ctx.CurrentTracker(); tracker != nil {
		tracker.Track(s.id)
	}

	return value
}

func (s *Signal[T]) Set(v T) {
	s.mu.Lock()
	changed := s.value != v
	if changed {
		s.value = v
	}
	s.mu.Unlock()

	if changed {
		s.ctx.Notify(s.id)
	}
}

func (s *Signal[T]) Peek() T {
	s.mu.RLock()
	value := s.value
	s.mu.RUnlock()
	return value
}

func (c *Context) Subscribe(cb func()) (SubscriberID, func()) {
	id := c.newSubscriberID()
	c.subscriberMu.Lock()
	c.subscribers[id] = cb
	c.subscriberMu.Unlock()

	return id, func() {
		c.Unsubscribe(id)
	}
}

func (c *Context) SubscribeAll(cb func(SignalID)) (SubscriberID, func()) {
	id := c.newSubscriberID()
	c.signalSubscriberMu.Lock()
	if c.signalSubscribers == nil {
		c.signalSubscribers = map[SignalID]map[SubscriberID]func(SignalID){}
	}
	if c.signalSubscribers[0] == nil {
		c.signalSubscribers[0] = map[SubscriberID]func(SignalID){}
	}
	c.signalSubscribers[0][id] = cb

	if c.subscriberSignals == nil {
		c.subscriberSignals = map[SubscriberID][]SignalID{}
	}
	c.subscriberSignals[id] = append(c.subscriberSignals[id], 0)
	c.signalSubscriberMu.Unlock()

	return id, func() {
		c.UnsubscribeAll(id)
	}
}

func (c *Context) Unsubscribe(id SubscriberID) {
	c.subscriberMu.Lock()
	delete(c.subscribers, id)
	c.subscriberMu.Unlock()

	c.UnsubscribeAll(id)
}

func (c *Context) UnsubscribeAll(id SubscriberID) {
	c.signalSubscriberMu.Lock()
	defer c.signalSubscriberMu.Unlock()

	signals, ok := c.subscriberSignals[id]
	if !ok {
		return
	}

	for _, sid := range signals {
		if listeners, ok := c.signalSubscribers[sid]; ok {
			delete(listeners, id)
			if len(listeners) == 0 {
				delete(c.signalSubscribers, sid)
			}
		}
	}
	delete(c.subscriberSignals, id)
}

func (c *Context) SubscribeSignal(sid SignalID, cb func(SignalID)) (SubscriberID, func()) {
	id := c.newSubscriberID()
	c.signalSubscriberMu.Lock()
	if c.signalSubscribers == nil {
		c.signalSubscribers = map[SignalID]map[SubscriberID]func(SignalID){}
	}
	if c.signalSubscribers[sid] == nil {
		c.signalSubscribers[sid] = map[SubscriberID]func(SignalID){}
	}
	c.signalSubscribers[sid][id] = cb

	if c.subscriberSignals == nil {
		c.subscriberSignals = map[SubscriberID][]SignalID{}
	}
	c.subscriberSignals[id] = append(c.subscriberSignals[id], sid)
	c.signalSubscriberMu.Unlock()

	return id, func() {
		c.UnsubscribeAll(id)
	}
}

func (c *Context) Notify(id SignalID) {
	c.subscriberMu.Lock()
	subs := make([]func(), 0, len(c.subscribers))
	for _, cb := range c.subscribers {
		subs = append(subs, cb)
	}
	c.subscriberMu.Unlock()

	c.signalSubscriberMu.Lock()
	signalSubs := make([]func(SignalID), 0)
	// Notify specific signal listeners
	if listeners, ok := c.signalSubscribers[id]; ok {
		for _, cb := range listeners {
			signalSubs = append(signalSubs, cb)
		}
	}
	// Notify "all" listeners (ID 0)
	if listeners, ok := c.signalSubscribers[0]; ok {
		for _, cb := range listeners {
			signalSubs = append(signalSubs, cb)
		}
	}
	c.signalSubscriberMu.Unlock()

	for _, cb := range subs {
		cb()
	}
	for _, cb := range signalSubs {
		cb(id)
	}
}

func Subscribe(cb func()) (SubscriberID, func()) {
	return defaultContext.Subscribe(cb)
}

func SubscribeAll(cb func(SignalID)) (SubscriberID, func()) {
	return defaultContext.SubscribeAll(cb)
}

func Unsubscribe(id SubscriberID) {
	defaultContext.Unsubscribe(id)
}

func UnsubscribeAll(id SubscriberID) {
	defaultContext.UnsubscribeAll(id)
}

func SubscribeSignal(sid SignalID, cb func(SignalID)) (SubscriberID, func()) {
	return defaultContext.SubscribeSignal(sid, cb)
}

func Notify(id SignalID) {
	defaultContext.Notify(id)
}

// notifyLocked removed: it read subscribers without holding the lock
// and was unused. Effects should be scheduled via an EffectRuntime.

type recordingTracker struct {
	mu   sync.Mutex
	deps map[SignalID]struct{}
}

func (r *recordingTracker) Track(id SignalID) {
	r.mu.Lock()
	if r.deps == nil {
		r.deps = make(map[SignalID]struct{})
	}
	r.deps[id] = struct{}{}
	r.mu.Unlock()
}

func NewComputed[T comparable](compute func() T) *Computed[T] {
	return NewComputedWithContext(defaultContext, compute)
}

func NewComputedWithContext[T comparable](c *Context, compute func() T) *Computed[T] {
	comp := &Computed[T]{
		ctx:     c,
		id:      c.newSignalID(),
		compute: compute,
		dirty:   true,
	}

	comp.update()
	return comp
}

func (c *Computed[T]) update() {
	tracker := &recordingTracker{}

	c.ctx.PushDependencyTracker(tracker)
	newVal := c.compute()
	c.ctx.PopDependencyTracker()

	c.mu.Lock()
	c.cached = newVal
	c.dirty = false
	if c.unsubscribe != nil {
		c.unsubscribe()
	}

	unsubscribes := make([]func(), 0, len(tracker.deps))
	for d := range tracker.deps {
		_, unsub := c.ctx.SubscribeSignal(d, func(sid SignalID) {
			c.Invalidate()
			c.ctx.Notify(c.id)
		})
		unsubscribes = append(unsubscribes, unsub)
	}
	c.unsubscribe = func() {
		for _, unsub := range unsubscribes {
			unsub()
		}
	}
	c.mu.Unlock()
}

func (c *Computed[T]) Dispose() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.unsubscribe != nil {
		c.unsubscribe()
	}
}

func (c *Computed[T]) Get() T {
	if tracker := c.ctx.CurrentTracker(); tracker != nil {
		tracker.Track(c.id)
	}

	c.mu.Lock()
	if c.dirty {
		c.mu.Unlock()
		c.update()
		c.mu.Lock()
	}
	value := c.cached
	c.mu.Unlock()
	return value
}

func (c *Computed[T]) SignalID() SignalID {
	return c.id
}

func (c *Computed[T]) Invalidate() {
	c.mu.Lock()
	c.dirty = true
	c.mu.Unlock()
}

// NewEffect removed: effects must be created with a runtime to ensure
// execution is routed through ScheduleOnUI and avoid data races.

func (c *Context) NewEffectWithRuntime(rt EffectRuntime, fn func(ctx EffectContext)) EffectID {
	id := c.newEffectID()
	c.effectMu.Lock()
	effect := &Effect{
		ctx:    c,
		id:     id,
		fn:     fn,
		active: true,
	}
	c.effects[id] = effect
	c.effectMu.Unlock()

	var run func()
	run = func() {
		rt.ScheduleOnUI(func() {
			c.effectMu.Lock()
			e, ok := c.effects[id]
			c.effectMu.Unlock()
			if !ok || !e.IsActive() {
				return
			}

			tracker := &recordingTracker{}
			c.PushDependencyTracker(tracker)

			e.fn(EffectContext{Ctx: context.Background(), Runtime: rt})

			c.PopDependencyTracker()

			e.mu.Lock()
			if e.unsub != nil {
				e.unsub()
			}
			unsubscribes := make([]func(), 0, len(tracker.deps))
			for d := range tracker.deps {
				_, unsub := c.SubscribeSignal(d, func(sid SignalID) {
					run()
				})
				unsubscribes = append(unsubscribes, unsub)
			}
			e.unsub = func() {
				for _, unsub := range unsubscribes {
					unsub()
				}
			}
			e.mu.Unlock()
		})
	}

	run()
	return id
}

func NewEffectWithRuntime(rt EffectRuntime, fn func(ctx EffectContext)) EffectID {
	return defaultContext.NewEffectWithRuntime(rt, fn)
}

func (c *Context) DisposeEffect(id EffectID) {
	c.effectMu.Lock()
	if effect, ok := c.effects[id]; ok {
		effect.mu.Lock()
		effect.active = false
		if effect.unsub != nil {
			effect.unsub()
		}
		effect.mu.Unlock()
		delete(c.effects, id)
	}
	c.effectMu.Unlock()
}

func DisposeEffect(id EffectID) {
	defaultContext.DisposeEffect(id)
}

func (e *Effect) IsActive() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.active
}
