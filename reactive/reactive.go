package reactive

import (
	"context"
	"sync"
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

type Subscriber struct {
	id SubscriberID
	cb func()
}

type Computed[T comparable] struct {
	id      SignalID
	deps    func() []SignalID
	compute func() T
	cached  T
	mu      sync.RWMutex
}

var (
	currentTracker DependencyTracker
	trackerMu      sync.RWMutex

	subscriberMu sync.Mutex
	subSeq       SubscriberID
	subscribers  = map[SubscriberID]func(){}

	signalSeq   SignalID
	signalSeqMu sync.Mutex
)

func SetDependencyTracker(t DependencyTracker) {
	trackerMu.Lock()
	currentTracker = t
	trackerMu.Unlock()
}

func CurrentTracker() DependencyTracker {
	trackerMu.RLock()
	defer trackerMu.RUnlock()
	return currentTracker
}

func newSignalID() SignalID {
	signalSeqMu.Lock()
	signalSeq++
	id := signalSeq
	signalSeqMu.Unlock()
	return id
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
	if s.value != v {
		s.value = v
		notifyLocked(s.id)
	}
	s.mu.Unlock()
}

func (s *Signal[T]) Peek() T {
	s.mu.RLock()
	value := s.value
	s.mu.RUnlock()
	return value
}

func Subscribe(cb func()) (SubscriberID, func()) {
	subscriberMu.Lock()
	subSeq++
	id := subscriberSeq
	subscribers[id] = cb
	subscriberMu.Unlock()

	return id, func() {
		Unsubscribe(id)
	}
}

func Unsubscribe(id SubscriberID) {
	subscriberMu.Lock()
	delete(subscribers, id)
	subscriberMu.Unlock()
}

func Notify(id SignalID) {
	subscriberMu.Lock()
	subs := make([]func(), 0, len(subscribers))
	for _, cb := range subscribers {
		subs = append(subs, cb)
	}
	subscriberMu.Unlock()

	for _, cb := range subs {
		cb()
	}
}

func notifyLocked(id SignalID) {
	for _, cb := range subscribers {
		cb()
	}
}

func NewComputed[T comparable](deps func() []SignalID, compute func() T) *Computed[T] {
	c := &Computed[T]{
		id:      newSignalID(),
		deps:    deps,
		compute: compute,
	}
	c.cached = compute()
	return c
}

func (c *Computed[T]) Get() T {
	if tracker := CurrentTracker(); tracker != nil {
		tracker.Track(c.id)
	}
	c.mu.RLock()
	value := c.cached
	c.mu.RUnlock()
	return value
}

func (c *Computed[T]) SignalID() SignalID {
	return c.id
}

func NewEffect(fn func(ctx EffectContext)) EffectID {
	go fn(EffectContext{Ctx: context.Background()})
	return EffectID(0)
}

func DisposeEffect(id EffectID) {
	// no-op placeholder for effect disposal
}
