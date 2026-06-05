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
	mu    sync.RWMutex
	id    SignalID
	value T
}

type Computed[T comparable] struct {
	id          SignalID
	compute     func() T
	cached      T
	mu          sync.RWMutex
	dirty       bool
	unsubscribe func()
}

type Effect struct {
	id     EffectID
	fn     func(ctx EffectContext)
	mu     sync.RWMutex
	active bool
	unsub  func()
}

var (
	trackerStack   []DependencyTracker
	trackerStackMu sync.Mutex

	subscriberMu sync.Mutex
	subscribers  = map[SubscriberID]func(){}

	signalSubscriberMu sync.Mutex
	signalSubscribers  = map[SignalID]map[SubscriberID]func(SignalID){}
	subscriberSignals  = map[SubscriberID][]SignalID{}

	effectMu sync.Mutex
	effects  = map[EffectID]*Effect{}

	signalSeq atomic.Uint64
	subSeq    atomic.Uint64
	effectSeq atomic.Uint64
)

func newSignalID() SignalID {
	return SignalID(signalSeq.Add(1))
}

func newSubscriberID() SubscriberID {
	return SubscriberID(subSeq.Add(1))
}

func newEffectID() EffectID {
	return EffectID(effectSeq.Add(1))
}

func PushDependencyTracker(t DependencyTracker) {
	trackerStackMu.Lock()
	trackerStack = append(trackerStack, t)
	trackerStackMu.Unlock()
}

func PopDependencyTracker() {
	trackerStackMu.Lock()
	if len(trackerStack) > 0 {
		trackerStack = trackerStack[:len(trackerStack)-1]
	}
	trackerStackMu.Unlock()
}

func CurrentTracker() DependencyTracker {
	trackerStackMu.Lock()
	defer trackerStackMu.Unlock()
	if len(trackerStack) == 0 {
		return nil
	}
	return trackerStack[len(trackerStack)-1]
}

func NewSignal[T comparable](initial T) *Signal[T] {
	return &Signal[T]{id: newSignalID(), value: initial}
}

func (s *Signal[T]) ID() SignalID {
	return s.id
}

func (s *Signal[T]) Get() T {
	s.mu.RLock()
	value := s.value
	s.mu.RUnlock()

	if tracker := CurrentTracker(); tracker != nil {
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
		Notify(s.id)
	}
}

func (s *Signal[T]) Peek() T {
	s.mu.RLock()
	value := s.value
	s.mu.RUnlock()
	return value
}

func Subscribe(cb func()) (SubscriberID, func()) {
	id := newSubscriberID()
	subscriberMu.Lock()
	subscribers[id] = cb
	subscriberMu.Unlock()

	return id, func() {
		Unsubscribe(id)
	}
}

func SubscribeAll(cb func(SignalID)) (SubscriberID, func()) {
	id := newSubscriberID()
	signalSubscriberMu.Lock()
	if signalSubscribers == nil {
		signalSubscribers = map[SignalID]map[SubscriberID]func(SignalID){}
	}
	if signalSubscribers[0] == nil {
		signalSubscribers[0] = map[SubscriberID]func(SignalID){}
	}
	signalSubscribers[0][id] = cb

	if subscriberSignals == nil {
		subscriberSignals = map[SubscriberID][]SignalID{}
	}
	subscriberSignals[id] = append(subscriberSignals[id], 0)
	signalSubscriberMu.Unlock()

	return id, func() {
		UnsubscribeAll(id)
	}
}

func Unsubscribe(id SubscriberID) {
	subscriberMu.Lock()
	delete(subscribers, id)
	subscriberMu.Unlock()

	UnsubscribeAll(id)
}

func UnsubscribeAll(id SubscriberID) {
	signalSubscriberMu.Lock()
	defer signalSubscriberMu.Unlock()

	signals, ok := subscriberSignals[id]
	if !ok {
		return
	}

	for _, sid := range signals {
		if listeners, ok := signalSubscribers[sid]; ok {
			delete(listeners, id)
			if len(listeners) == 0 {
				delete(signalSubscribers, sid)
			}
		}
	}
	delete(subscriberSignals, id)
}

func SubscribeSignal(sid SignalID, cb func(SignalID)) (SubscriberID, func()) {
	id := newSubscriberID()
	signalSubscriberMu.Lock()
	if signalSubscribers == nil {
		signalSubscribers = map[SignalID]map[SubscriberID]func(SignalID){}
	}
	if signalSubscribers[sid] == nil {
		signalSubscribers[sid] = map[SubscriberID]func(SignalID){}
	}
	signalSubscribers[sid][id] = cb

	if subscriberSignals == nil {
		subscriberSignals = map[SubscriberID][]SignalID{}
	}
	subscriberSignals[id] = append(subscriberSignals[id], sid)
	signalSubscriberMu.Unlock()

	return id, func() {
		UnsubscribeAll(id)
	}
}

func Notify(id SignalID) {
	subscriberMu.Lock()
	subs := make([]func(), 0, len(subscribers))
	for _, cb := range subscribers {
		subs = append(subs, cb)
	}
	subscriberMu.Unlock()

	signalSubscriberMu.Lock()
	signalSubs := make([]func(SignalID), 0)
	// Notify specific signal listeners
	if listeners, ok := signalSubscribers[id]; ok {
		for _, cb := range listeners {
			signalSubs = append(signalSubs, cb)
		}
	}
	// Notify "all" listeners (ID 0)
	if listeners, ok := signalSubscribers[0]; ok {
		for _, cb := range listeners {
			signalSubs = append(signalSubs, cb)
		}
	}
	signalSubscriberMu.Unlock()

	for _, cb := range subs {
		cb()
	}
	for _, cb := range signalSubs {
		cb(id)
	}
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
	c := &Computed[T]{
		id:      newSignalID(),
		compute: compute,
		dirty:   true,
	}

	c.update()
	return c
}

func (c *Computed[T]) update() {
	tracker := &recordingTracker{}

	PushDependencyTracker(tracker)
	newVal := c.compute()
	PopDependencyTracker()

	c.mu.Lock()
	c.cached = newVal
	c.dirty = false
	if c.unsubscribe != nil {
		c.unsubscribe()
	}

	unsubscribes := make([]func(), 0, len(tracker.deps))
	for d := range tracker.deps {
		_, unsub := SubscribeSignal(d, func(sid SignalID) {
			c.Invalidate()
			Notify(c.id)
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
	if tracker := CurrentTracker(); tracker != nil {
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

func NewEffectWithRuntime(rt EffectRuntime, fn func(ctx EffectContext)) EffectID {
	id := newEffectID()
	effectMu.Lock()
	effect := &Effect{
		id:     id,
		fn:     fn,
		active: true,
	}
	effects[id] = effect
	effectMu.Unlock()

	var run func()
	run = func() {
		rt.ScheduleOnUI(func() {
			effectMu.Lock()
			e, ok := effects[id]
			effectMu.Unlock()
			if !ok || !e.IsActive() {
				return
			}

			tracker := &recordingTracker{}
			PushDependencyTracker(tracker)

			e.fn(EffectContext{Ctx: context.Background(), Runtime: rt})

			PopDependencyTracker()

			e.mu.Lock()
			if e.unsub != nil {
				e.unsub()
			}
			unsubscribes := make([]func(), 0, len(tracker.deps))
			for d := range tracker.deps {
				_, unsub := SubscribeSignal(d, func(sid SignalID) {
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

func DisposeEffect(id EffectID) {
	effectMu.Lock()
	if effect, ok := effects[id]; ok {
		effect.mu.Lock()
		effect.active = false
		if effect.unsub != nil {
			effect.unsub()
		}
		effect.mu.Unlock()
		delete(effects, id)
	}
	effectMu.Unlock()
}

func (e *Effect) IsActive() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.active
}
