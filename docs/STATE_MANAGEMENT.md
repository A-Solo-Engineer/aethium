# Aethium — State Management

Binding decisions for reactivity, memory, derived state, side effects, and concurrency safety. Must agree with `ARCHITECTURE.md` (`VNode`, `Runtime`, signal-driven `Update`).

---

## Reactivity Model Decision

**Chosen: (a) Signal-based reactivity (SolidJS-style fine-grained subscriptions).**

Each reactive value is a `Signal[T]` with explicit subscriber edges. When a signal changes, only components and computations registered on that signal are marked dirty—no full-tree diff and no virtual DOM.

### Rejected: (b) Observable structs with dirty-checking

| Reason | Detail |
|--------|--------|
| Go type system | Struct field tagging (`obs:"name"`) requires reflection or code generation for every field; easy to miss fields. |
| CPU cost | Periodic deep equality on large structs scales O(n) per frame—hostile to 60 fps on mid-range hardware. |
| GC | Reflect-based comparison allocates interface wrappers and slice scratch space. |

### Rejected: (c) Immutable state trees with diffing (Redux/Elm)

| Reason | Detail |
|--------|--------|
| GC pressure | Every keystroke clones maps/slices; young-gen churn fights `sync.Pool` benefits. |
| Boilerplate | Go lacks native spread/update operators; ergonomics push users toward code gen or `any`. |
| Over-invalidation risk | Structural sharing libraries are immature in Go; without them, diffing reintroduces O(n) scans. |

### Why signals fit Go

- **Value semantics + pointers:** `Signal[T]` holds `atomic.Value` or mutex-guarded `T` with subscriber slice stored as `[]*Subscriber` reused via pool.
- **Explicit graphs:** Dependencies are registered at read time (`signal.Get()` in `Build`), matching Go’s explicit-error style.
- **Granular GC:** Fewer short-lived allocations than immutable patches; pooled `VNode` and `DrawCmd` buffers complement signals.

---

## Data Flow Diagram

```
[User input]
    |
    v
Host event (click/key) ---> platform.InputEvent
    |
    v
runtime.UI queue <--- ScheduleOnUI
    |
    v
Handler func()  e.g. button.OnClick
    |
    v
signal.Set(counter, counter+1)     // Signal[int]
    |
    v
reactive.Notify(signalID)          // marks Subscriber nodes
    |
    v
scene.MarkDirty(vnodeID...)        // maps SignalID -> []*VNode
    |
    v
Runtime.Tick -> Update(UpdateContext{Dirty: ...})
    |
    v
component.Build(ctx) -> []canvas.DrawCmd   // reads signals via .Get()
    |
    v
canvas.MergeDrawList -> platform.Submit
    |
    v
GPU / canvas repaint (immediate-mode)
```

### Go types at each step

| Step | Type |
|------|------|
| Input | `platform.InputEvent` |
| Handler | `func()` scheduled on UI thread |
| Mutation | `reactive.Signal[T].Set(T)` |
| Notification | `reactive.SignalID`, `reactive.Subscriber` |
| Dirty tree | `scene.NodeID`, `*scene.VNode` |
| Re-render | `runtime.UpdateContext` |
| Output | `[]canvas.DrawCmd` |

---

## Memory Strategy

### What gets pooled

| Pool | Type | When Get | Reset contract |
|------|------|----------|----------------|
| `vnodePool` | `*scene.VNode` | `Init` / tree expand | `Reset()` clears `Children`, `Signals`, `Component` ref; zero `ID` |
| `drawCmdPool` | `[]canvas.DrawCmd` | each `Build` | `cmds = cmds[:0]`; do not retain pointers to sub-slices |
| `subscriberPool` | `*reactive.Subscriber` | `signal.Subscribe` | `Reset()` clears `SignalIDs`, `Callback` |
| `dirtyListPool` | `[]scene.NodeID` | `MarkDirty` burst | `ids = ids[:0]` |

Global pools (package-level):

```go
var vnodePool = sync.Pool{
    New: func() any { return &scene.VNode{} },
}
```

### Interaction with signals

Signals **do not** live in pools—they are long-lived app state. `VNode` is ephemeral tree structure; signals persist across mount/unmount unless explicitly disposed in `Unmount`.

On `Unmount`, `signal.Unsubscribe(subscriberID)` runs **before** `vnodePool.Put` so no stale edges reference freed nodes.

### Concrete example (pseudocode)

```go
type Signal[T any] struct {
    id    SignalID
    mu    sync.Mutex
    value T
    subs  []*Subscriber // small; not pooled
}

func (s *Signal[T]) Get() T {
    if cur := runtime.CurrentBuild(); cur != nil {
        cur.Track(s.id) // registers dependency for this frame/build
    }
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.value
}

func (s *Signal[T]) Set(v T) {
    s.mu.Lock()
    if s.value == v { // requires comparable T or reflect for generic - API uses comparable constraint
        s.mu.Unlock()
        return
    }
    s.value = v
    s.mu.Unlock()
    reactive.Notify(s.id)
}

func acquireVNode() *scene.VNode {
    n := vnodePool.Get().(*scene.VNode)
    n.Reset()
    return n
}

func releaseVNode(n *scene.VNode) {
    n.Reset()
    vnodePool.Put(n)
}
```

**Comparable constraint:** Public API documents `Signal[T]` requires `T comparable` for cheap equality short-circuit; non-comparable types use `Signal[*T]` with pointer identity.

---

## Derived State & Side Effects

### Derived (computed) values

**Construct:** `reactive.Computed[T](deps func() []SignalID, compute func() T)`

- Internally registers on all dependency signals.
- Maintains `cached T` and `version uint64`.
- On dependency notify, marks computed **stale** but does **not** immediately run `compute`.
- On `Get()` during `Build`, if stale, runs `compute` once (lazy), updates cache, does not notify dependents unless output changed (comparable `T`).

**No redundant re-renders:** Components depend on `Computed` signal IDs, not raw deps. Intermediate signals can update without dirtying UI if computed output is unchanged.

### Side effects

**Construct:** `reactive.Effect(func(ctx EffectContext))`

- **Not** part of the render graph. Effects run **after** `Runtime.Tick` paint, on UI thread, coalesced per frame max once per Effect instance.
- Allowed: network (`http.Client`), file I/O initiation, `ScheduleOnUI` follow-ups.
- Forbidden inside `Build`: any blocking I/O or `Effect` registration.

**Isolation:** `EffectContext` carries `context.Context` cancelled on `Unmount`. Workers started from effects return via channels; effects resume only through `ScheduleOnUI`.

```
Signal change --> Update/Build (sync, UI thread)
              --> queue effects
              --> after SubmitDrawList --> run Effect batch
```

---

## Concurrency Safety

### Goroutine-safe by design

| Operation | Mechanism |
|-----------|-----------|
| `Signal.Set` / `Get` (no build context) | `sync.Mutex` per signal |
| Worker → UI delivery | `chan` + `ScheduleOnUI` only |
| `http` / file read in worker | Immutable or channel-passed results |

### Not safe without discipline

| Operation | Rule |
|-----------|------|
| `Build`, `Mount`, canvas submit | UI goroutine only |
| `Computed.Get` off UI thread | **Forbidden** — may touch build tracking |
| Reading parent `VNode` from worker | **Forbidden** |

### Locking strategy

1. **Per-signal mutex** — low contention; signals are coarse (user/session), not per-pixel.
2. **`Runtime` global `sync.RWMutex`** — protects tree structure mutations (`Attach`/`Detach`); reads during `Tick` hold `RLock`.
3. **No lock ordering across signals** — prevents deadlock; `Effect` must not call `Set` on signal A while holding lock on B; use `ScheduleOnUI` to defer cross-signal writes.
4. **Worker results** — passed as immutable copies or `[]byte` owned by receiver; no shared pointers into `VNode`.

### Documented anti-pattern

Calling `signal.Set` from a worker without `ScheduleOnUI` is a **data race** on subscribers; CI race detector (`-race`) must pass on runtime tests.

---

## Cross-reference

| Item | Location |
|------|----------|
| Lifecycle hooks using signals | `ARCHITECTURE.md` |
| TinyGo / build | `BUILD_SYSTEM.md` |
| Benchmarks for signal flush | `CONTRIBUTING.md` |

**Stage gate:** No implementation until **`APPROVED — proceed to Stage 2`**.
