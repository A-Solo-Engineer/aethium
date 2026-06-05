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

type Computed[T comparable] struct {
    id          SignalID
    deps        func() []SignalID
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
}

var (
    currentTracker DependencyTracker
    trackerMu      sync.RWMutex

    subscriberMu sync.Mutex
    subSeq       SubscriberID
    subscribers  = map[SubscriberID]func(){}

    signalSubscriberMu sync.Mutex
    signalSubscribers  = map[SignalID]map[SubscriberID]func(SignalID){}

    effectMu sync.Mutex
    effectSeq EffectID
    effects  = map[EffectID]*Effect{}

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
    subscriberMu.Lock()
    subSeq++
    id := subSeq
    subscribers[id] = cb
    subscriberMu.Unlock()

    return id, func() {
        Unsubscribe(id)
    }
}

func SubscribeAll(cb func(SignalID)) (SubscriberID, func()) {
    signalSubscriberMu.Lock()
    subSeq++
    id := subSeq
    if signalSubscribers == nil {
        signalSubscribers = map[SignalID]map[SubscriberID]func(SignalID){}
    }
    if signalSubscribers[0] == nil {
        signalSubscribers[0] = map[SubscriberID]func(SignalID){}
    }
    signalSubscribers[0][id] = cb
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
    for signalID, listeners := range signalSubscribers {
        delete(listeners, id)
        if len(listeners) == 0 {
            delete(signalSubscribers, signalID)
        }
    }
    signalSubscriberMu.Unlock()
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
    for _, listeners := range signalSubscribers {
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

func NewComputed[T comparable](deps func() []SignalID, compute func() T) *Computed[T] {
    c := &Computed[T]{
        id:      newSignalID(),
        deps:    deps,
        compute: compute,
        cached:  compute(),
        dirty:   false,
    }

    // Subscribe to all signal changes and invalidate when a dependency changes.
    _, unsubscribe := SubscribeAll(func(sid SignalID) {
        for _, d := range c.deps() {
            if d == sid {
                c.Invalidate()
                Notify(c.id)
                return
            }
        }
    })
    c.unsubscribe = unsubscribe

    return c
}

func (c *Computed[T]) Dispose() {
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
        newVal := c.compute()
        if newVal != c.cached {
            c.cached = newVal
        }
        c.dirty = false
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
    effectMu.Lock()
    effectSeq++
    id := effectSeq
    effect := &Effect{
        id:     id,
        fn:     fn,
        active: true,
    }
    effects[id] = effect
    effectMu.Unlock()

    // Route effect execution through the runtime's UI scheduler to avoid
    // data races with Computed.Get() and CurrentTracker access.
    rt.ScheduleOnUI(func() {
        effectMu.Lock()
        e, ok := effects[id]
        effectMu.Unlock()
        if !ok || !e.IsActive() {
            return
        }
        e.fn(EffectContext{Ctx: context.Background(), Runtime: rt})
    })

    return id
}

func DisposeEffect(id EffectID) {
    effectMu.Lock()
    if effect, ok := effects[id]; ok {
        effect.mu.Lock()
        effect.active = false
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
