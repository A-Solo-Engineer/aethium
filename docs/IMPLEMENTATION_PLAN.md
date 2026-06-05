# Aethium — Implementation Plan (Stage 2)

**Status:** Draft — awaiting **`APPROVED — proceed to implementation`** before any source code.

**Prerequisites satisfied:** Stage 1 docs approved (`APPROVED — proceed to Stage 2`).

**Scope:** Core rendering abstraction (Wasm + desktop), reactive state engine, component lifecycle runtime, `aethium` CLI, dev server with hot-reload. Delivers `examples/hello` passing all Resource Efficiency Budget exit criteria in `BUILD_SYSTEM.md`.

**Module root:** `github.com/aethium-dev/aethium`  
**Go version:** `1.22` (toolchain ≥ 1.22.7)  
**TinyGo:** `0.34.0` (browser/app Wasm only)

---

## 1. Repository layout

```
github.com/aethium-dev/aethium/
├── go.mod
├── LICENSE                          # AGPL-3.0
├── CHANGELOG.md
├── aethium.toml                     # example config (documented; used by examples)
│
├── canvas/                          # DrawList IR (TinyGo + native)
├── reactive/                        # Signals, Computed, Effect (TinyGo + native)
├── scene/                           # VNode tree, dirty propagation (TinyGo + native)
├── runtime/                         # Lifecycle, UI queue, Tick (TinyGo + native)
├── platform/                        # Backend interface + shared host types
│   ├── backend.go
│   ├── event.go
│   ├── js/                          # //go:build tinygo && js
│   └── webview/                     # //go:build !tinygo || !js  (desktop loader)
│
├── cmd/aethium/                     # CLI entry (native Go only)
│   └── main.go
├── internal/
│   ├── build/                       # TinyGo/go orchestration, size gates
│   ├── devserver/                   # HTTP, WS/SSE reload
│   ├── scaffold/                    # aethium new templates
│   └── embedassets/                 # index.html, aethium.js templates
│
├── examples/hello/                  # Budget reference app
│   ├── main.go
│   └── app/
│       └── counter.go
│
└── docs/                            # Stage 1 + this plan + rfcs/ exceptions/
```

### Build tags

| Tag | Meaning |
|-----|---------|
| `tinygo` | File compiles under TinyGo 0.34.0; no unsupported stdlib |
| `js` | Wasm browser host (`platform/js`) |
| `desktop` | Native WebView host (`platform/webview`) — default for `go build` |

**Rule:** `canvas`, `reactive`, `scene`, `runtime`, `platform` (root types only) compile with `-tags tinygo,js` and without tags on desktop.

**Rule:** `cmd/aethium`, `internal/*`, `platform/webview` never compile into `app.wasm`.

---

## 2. Dependency graph (packages)

```
examples/hello
    └── runtime, reactive, canvas, scene, platform/js | platform/webview

cmd/aethium
    └── internal/build, internal/devserver, internal/scaffold, internal/embedassets

runtime
    └── scene, reactive, canvas, platform

scene
    └── reactive, canvas

reactive
    └── (stdlib only)

canvas
    └── (stdlib only)

platform/js
    └── platform, runtime (init hook only via RegisterBackend)

platform/webview
    └── platform, github.com/webview/webview
```

**Forbidden cycles:** `reactive` must not import `runtime` or `scene`. `runtime` imports `reactive`; `scene` imports `reactive` only. Build tracking (`CurrentBuild`) lives in `scene` and is called from `reactive` via a narrow callback interface registered at `runtime` init to avoid import cycle:

```go
// scene/buildctx.go — implemented by runtime, registered once
type DependencyTracker interface {
    Track(SignalID)
}
```

---

## 3. Public API — `canvas`

Package: `github.com/aethium-dev/aethium/canvas`

Immediate-mode draw IR. No platform imports.

### Types

```go
type CmdKind uint8

const (
    CmdFillRect CmdKind = iota + 1
    CmdStrokeRect
    CmdDrawText
    CmdClip
    CmdTransform
)

type Color uint32 // ARGB, 0xAARRGGBB

type DrawCmd struct {
    Kind     CmdKind
    X, Y, W, H float32
    Color    Color
    Text     string // CmdDrawText only; max 4KB per cmd enforced in builder
    Transform [6]float32 // 2x3 affine; CmdTransform only
}

type DrawList struct {
    Cmds []DrawCmd
}

type Rect struct {
    X, Y, W, H float32
}
```

### Functions

```go
// NewDrawList returns a pooled-empty list; caller must Release when done.
func NewDrawList() *DrawList

// Release returns backing slice to pool per STATE_MANAGEMENT.md reset contract.
func (dl *DrawList) Release()

// Append adds a command; may reallocate until Cap sufficient.
func (dl *DrawList) Append(cmd DrawCmd)

// Merge appends src into dst and Release src.
func Merge(dst, src *DrawList)

// Builder helpers (allocate one DrawCmd each)
func FillRect(dl *DrawList, r Rect, c Color)
func StrokeRect(dl *DrawList, r Rect, c Color)
func DrawText(dl *DrawList, x, y float32, text string, c Color)
func Clip(dl *DrawList, r Rect)
```

### Budget allocation (`canvas` + minimal bitmap font)

| Artifact | Desktop | Wasm gzip |
|----------|---------|-----------|
| `canvas` package | ≤ 400 KB | ≤ 60 KB |
| `canvas/font` (embedded 8×8 bitmap, ASCII 32–126) | ≤ 600 KB | ≤ 60 KB |
| **Subtotal** | **≤ 1.0 MB** | **≤ 120 KB** |

Font: single `internal/fontdata` blob, no TTF shaping in Stage 2.

---

## 4. Public API — `reactive`

Package: `github.com/aethium-dev/aethium/reactive`

### Types

```go
type SignalID uint64

type SubscriberID uint64

// Signal holds comparable T only.
type Signal[T comparable] struct { /* unexported */ }

type Subscriber struct { /* unexported */ }

type EffectID uint64

type EffectContext struct {
    Ctx context.Context
    Runtime *EffectRuntime // narrow: ScheduleOnUI only
}

type EffectRuntime interface {
    ScheduleOnUI(fn func())
}

type Computed[T comparable] struct { /* unexported */ }
```

### Functions — signals

```go
func NewSignal[T comparable](initial T) *Signal[T]

func (s *Signal[T]) ID() SignalID

func (s *Signal[T]) Get() T

func (s *Signal[T]) Set(v T)

func (s *Signal[T]) Peek() T // no dependency tracking; UI-thread only

func Subscribe(cb func()) (SubscriberID, func()) // unsubscribe on return call

func Unsubscribe(id SubscriberID)
```

`Get()` calls `scene.CurrentTracker().Track(id)` when tracker non-nil.

### Functions — computed & effects

```go
func NewComputed[T comparable](
    deps func() []SignalID,
    compute func() T,
) *Computed[T]

func (c *Computed[T]) Get() T

func (c *Computed[T]) SignalID() SignalID // exposes as signal dependency

func NewEffect(fn func(ctx EffectContext)) EffectID

func DisposeEffect(id EffectID)

// Notify marks subscribers dirty; called internally from Set.
func Notify(id SignalID)
```

### Budget allocation (`reactive`)

| Artifact | Desktop | Wasm gzip |
|----------|---------|-----------|
| `reactive` | ≤ 350 KB | ≤ 80 KB |

---

## 5. Public API — `scene`

Package: `github.com/aethium-dev/aethium/scene`

VNode tree and dirty propagation. Does **not** run lifecycle hooks (that is `runtime`).

### Types

```go
type NodeID uint64

type VNode struct {
    ID        NodeID
    Parent    NodeID
    Children  []NodeID
    Component componentRef // opaque interface{} + type hash in runtime
    SignalIDs []SignalID
}

type DirtySet struct { /* unexported */ }

// View builds draw commands for one component instance.
type View func(ctx ViewContext) *canvas.DrawList

type ViewContext struct {
    NodeID NodeID
    Props  any
}
```

### Functions

```go
func (n *VNode) Reset() // pool contract

func AcquireVNode() *VNode
func ReleaseVNode(n *VNode)

func NewDirtySet() *DirtySet
func (d *DirtySet) Mark(ids ...NodeID)
func (d *DirtySet) TakeAll() []NodeID
func (d *DirtySet) Release()

func MarkDirtyForSignal(signalID SignalID, reg *SignalRegistry)

// SignalRegistry maps signal -> vnode subscribers; owned by runtime, passed in
type SignalRegistry struct { /* unexported */ }

func (r *SignalRegistry) Register(node NodeID, signals ...SignalID)
func (r *SignalRegistry) Unregister(node NodeID)

func SetDependencyTracker(t DependencyTracker)
func CurrentTracker() DependencyTracker
```

### Budget allocation (`scene`)

| Artifact | Desktop | Wasm gzip |
|----------|---------|-----------|
| `scene` | ≤ 300 KB | ≤ 70 KB |

---

## 6. Public API — `runtime`

Package: `github.com/aethium-dev/aethium/runtime`

Lifecycle orchestration, UI queue, frame tick.

### Types — lifecycle (per ARCHITECTURE.md)

```go
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
    Dirty []scene.NodeID
}

type UnmountContext struct {
    MountContext
}

type FrameInfo struct {
    Frame uint64
    Delta time.Duration // seconds as float64 internally if TinyGo lacks time — use float64 seconds
    Width, Height float32
}
```

### Types — runtime

```go
type Runtime struct { /* unexported */ }

type RuntimeConfig struct {
    Backend platform.Backend
    WorkerChanCap int // default 64
}
```

### Functions

```go
func New(cfg RuntimeConfig) (*Runtime, error)

func (r *Runtime) Attach(root Component, props any) (scene.NodeID, error)

func (r *Runtime) Detach(id scene.NodeID) error

func (r *Runtime) Tick(frame FrameInfo) error

// ScheduleOnUI enqueues fn for next Tick drain; safe from workers.
func (r *Runtime) ScheduleOnUI(fn func())

// PumpEvents drains host mailbox; called from platform before Tick.
func (r *Runtime) PumpEvents()

// RunWorker starts fn in goroutine; panics if called from Wasm build (build tag stub).
func (r *Runtime) RunWorker(fn func())

// RegisterView associates a component type with its View function.
func RegisterView[T any](view scene.View)

// CurrentBuild returns build context for signal tracking.
func CurrentBuild() *BuildContext
```

```go
type BuildContext struct { /* unexported */ }
func (b *BuildContext) Track(id reactive.SignalID)
```

### Internal tick pipeline (not public, documented for implementers)

1. `PumpEvents` → drain `uiMailbox` → `ScheduleOnUI` closures  
2. Flush pending signal notifications → `scene.MarkDirtyForSignal`  
3. `DirtySet.TakeAll` → `Update` on affected nodes (depth-first)  
4. Each `View` → `canvas.DrawList` → merge to frame list  
5. `platform.Backend.Submit(frameList)`  
6. Run `reactive` effect batch (post-submit)

### Budget allocation (`runtime` + `scene` + `reactive` combined row)

| Artifact | Desktop | Wasm gzip |
|----------|---------|-----------|
| `runtime` | ≤ 850 KB | ≤ 100 KB |
| **Combined runtime+reactive+scene ceiling** | **≤ 1.5 MB** | **≤ 250 KB** |

Implementation must verify combined totals in CI, not only per-package estimates.

---

## 7. Public API — `platform`

Package: `github.com/aethium-dev/aethium/platform`

### Types

```go
type InputKind int

const (
    InputPointerDown InputKind = iota + 1
    InputPointerUp
    InputPointerMove
    InputKeyDown
    InputKeyUp
)

type InputEvent struct {
    Kind      InputKind
    X, Y      float32
    Button    int
    Key       string
    Modifiers uint32
}

type Backend interface {
    Init() error
    Run(loop func(frame FrameInfo) error) error // blocks; calls loop each frame
    Submit(dl *canvas.DrawList) error
    SetInputHandler(h func(InputEvent))
    Shutdown() error
}

type FrameInfo = runtime.FrameInfo // alias to avoid duplicate type
```

### `platform/js` (TinyGo Wasm)

```go
// RegisterBackend installs syscall/js host. Called from app init.
func RegisterBackend() error

// Export for aethium.js — name must match glue
func AethiumPump() // called each rAF from JS
```

### `platform/webview` (native desktop)

```go
type WebViewConfig struct {
    Title   string
    Width   int
    Height  int
    DataDir string // wasm+html root or embed FS
}

func Run(cfg WebViewConfig, rt *runtime.Runtime) error
```

### Budget allocation

| Artifact | Desktop | Wasm gzip |
|----------|---------|-----------|
| `platform/js` + `aethium.js` glue | N/A | ≤ 80 KB |
| `platform/webview` | ≤ 800 KB | N/A |

---

## 8. Public API — `cmd/aethium` (CLI contract)

Implements `BUILD_SYSTEM.md` exactly. Native Go only.

```go
// main dispatches:
//   aethium new    --module PATH [--template minimal] [--dir .]
//   aethium build  --target desktop|wasm|all [--release true] [--output dist/] [-tags] [--tinygo auto]
//   aethium dev    --target wasm|desktop [--addr 127.0.0.1:5173] [--open] [--tinygo-opt 0]
//   aethium bundle --target desktop|wasm [--format native|wasm] [--output dist/bundle]
```

### `internal/build` (not public)

```go
func BuildWasm(cfg WasmConfig) (artifacts WasmArtifacts, err error)
func BuildDesktop(cfg DesktopConfig) (artifacts DesktopArtifacts, err error)
func CheckBudgets(artifacts Paths) (report BudgetReport, err error)
```

Post-build `CheckBudgets` enforces ceilings unless `--skip-budget-check` and `docs/exceptions/*.md` present.

### Budget allocation (`cmd/aethium` + embed)

| Artifact | Desktop | Wasm gzip |
|----------|---------|-----------|
| CLI binary (includes templates, dev server for `aethium dev` binary size) | ≤ 1.2 MB | N/A |

Note: `aethium dev` is a subcommand of the same binary; dev server code lives in `internal/devserver` and counts toward CLI budget when building `cmd/aethium`.

---

## 9. Application author API (`examples/hello`)

Apps import public packages only:

```go
import (
    "github.com/aethium-dev/aethium/runtime"
    "github.com/aethium-dev/aethium/reactive"
    "github.com/aethium-dev/aethium/canvas"
    "github.com/aethium-dev/aethium/platform/js" // wasm main only
)
```

### Minimal hello pattern

```go
func main() {
    js.RegisterBackend()
    rt, _ := runtime.New(runtime.RuntimeConfig{ /* backend from platform */ })
    count := reactive.NewSignal(0)
    runtime.RegisterView[CounterProps](func(ctx scene.ViewContext) *canvas.DrawList { /* ... */ })
    rt.Attach(&Counter{}, CounterProps{Signal: count})
    /* platform Run loop */
}
```

### User app budget (`examples/hello`)

| Artifact | Desktop | Wasm gzip |
|----------|---------|-----------|
| App + embed | ≤ 500 KB | ≤ 50 KB |

---

## 10. Resource Efficiency Budget — full allocation table

**Exit criteria** (from `BUILD_SYSTEM.md`) — implementation must pass all:

| Criterion | Ceiling |
|-----------|---------|
| Desktop binary (bundled hello) | ≤ 5 MB |
| Wasm gzip total (wasm+js+html) | ≤ 500 KB |
| Cold startup desktop p50 | ≤ 150 ms |
| Cold startup browser/Wasm p50 | ≤ 300 ms |
| Idle RAM desktop | ≤ 40 MB |

### Per-module allocations (sum ≤ ceilings)

| Module / artifact | Desktop (max) | Wasm gzip (max) | Verification |
|-------------------|---------------|-----------------|--------------|
| `canvas` + font | 1.0 MB | 120 KB | `go tool nm -size` / wasm gzip split |
| `reactive` | (in combined) | (in combined) | bench + size |
| `scene` | (in combined) | (in combined) | bench + size |
| `runtime` | 1.5 MB combined w/ reactive+scene | 250 KB combined | single `tinygo build` link map |
| `platform/webview` | 0.8 MB | — | desktop link only |
| `platform/js` + glue | — | 80 KB | gzip `aethium.js` + platform/js archive |
| `cmd/aethium` | 1.2 MB | — | release binary size |
| `examples/hello` app | 0.5 MB | 50 KB | remainder after framework |
| **Shipping total** | **≤ 5.0 MB** | **≤ 500 KB** | `internal/build.CheckBudgets` |

**Headroom rule:** Framework subtotals target **90%** of ceiling (4.5 MB / 450 KB) so hello app has margin without exception docs.

### CI enforcement

```text
go test ./...
tinygo build -target=wasm -tags='tinygo,js' -o /tmp/app.wasm ./examples/hello
aethium build --target all --output /tmp/dist && aethium-internal-budget-check /tmp/dist
```

`internal/build.CheckBudgets` is invoked from `cmd/aethium build` by default.

---

## 11. Sequenced build order

Implementation proceeds in this order. **Do not start a phase until the previous phase has tests green and size estimate logged.**

| Phase | Deliverable | Why first |
|-------|-------------|-----------|
| **P0** | `go.mod`, LICENSE, CI skeleton, `internal/build` budget stub | Establishes module path and gates |
| **P1** | `canvas` + pools + unit tests | Leaf package; DrawList IR required by all else |
| **P2** | `reactive` + `Notify` + bench `BenchmarkSignalNotify` | State layer without runtime cycle |
| **P3** | `scene` + dirty set + `SignalRegistry` + pool tests | Tree structure before lifecycle |
| **P4** | `runtime` lifecycle + `Tick` + `ScheduleOnUI` (native tests with mock `Backend`) | Integrates P1–P3; bench `BenchmarkTick` |
| **P5** | `platform` mock + `platform/js` + `aethium.js` + Wasm hello (static canvas clear) | Proves TinyGo path + ≤300 ms startup work |
| **P6** | `platform/webview` + desktop hello | Proves WebView host + ≤150 ms startup work |
| **P7** | `cmd/aethium` (`new`, `build`) + `examples/hello` full counter UI | End-to-end production build + budget pass |
| **P8** | `internal/devserver` + `aethium dev` hot-reload | Depends on stable Wasm emit from P7 |
| **P9** | `aethium bundle` + size report in CI + CHANGELOG | Shipping path complete |

### Phase exit checklist (each phase)

- [ ] Unit tests (`go test`) pass; `-race` on native packages  
- [ ] TinyGo compile of touched packages (if applicable)  
- [ ] No new deps outside allowlist (`webview` only per `BUILD_SYSTEM.md`)  
- [ ] AGPL header on new files  
- [ ] Running size estimate documented in PR/commit message  

---

## 12. Testing & benchmarks (Stage 2)

| Package | Required tests |
|---------|----------------|
| `canvas` | Pool reset, merge release, cmd bounds |
| `reactive` | Signal dedup Set, computed lazy stale, effect once/frame |
| `scene` | Dirty propagation, unregister on unmount |
| `runtime` | Init→Mount→Update→Unmount order, no blocking on UI thread |
| `platform/js` | Smoke: Init+one Submit (build tag gated) |

| Benchmark | Target (informal) |
|-----------|-------------------|
| `BenchmarkSignalNotify` | Baseline recorded P0.4; regress ≤5% |
| `BenchmarkTick` | < 2 ms/op hello tree, desktop native mock |
| `BenchmarkDrawListAppend` | Document baseline |

---

## 13. `aethium.toml` schema (Stage 2)

```toml
[project]
name = "hello"
module = "example.com/hello"

[wasm]
target = "wasm"
tinygo_version = "0.34.0"
opt = "2"           # release
tags = ["tinygo,js"]

[desktop]
tags = ["desktop"]
webview_title = "Hello"

[dev]
addr = "127.0.0.1:5173"
```

Parsed by `internal/build`; unknown keys error.

---

## 14. Dev server implementation notes (P8)

| Component | Behavior |
|-----------|----------|
| File watcher | `*.go` → TinyGo rebuild `app.wasm` only |
| WS/SSE | Push `{type:"reload", wasm:"/app.wasm?v=hash"}` |
| `aethium.js` | Injected script; `AethiumPump` + module swap |
| Latency | Log rebuild ms; warn if p95 > 800 ms rebuild or > 1000 ms e2e |

Desktop dev: exec rebuild of native loader + WebView restart; **≤3 s p95** documented exception only.

---

## 15. Stage gates (summary)

| Gate | Requirement |
|------|-------------|
| Stage 1 → 2 plan | **`APPROVED — proceed to Stage 2`** ✅ |
| Stage 2 plan → code | **`APPROVED — proceed to implementation`** (pending) |
| Stage 2 → 3 docs | All budgets pass or `docs/exceptions/` approved; Stage 3 reconciliation |

**No `.go` implementation files until `APPROVED — proceed to implementation`.**

---

## 16. Document cross-reference

| Decision | Source doc |
|----------|------------|
| Immediate-mode canvas | `ARCHITECTURE.md` |
| Signals, pools | `STATE_MANAGEMENT.md` |
| CLI flags, budgets | `BUILD_SYSTEM.md` |
| AGPL, metrics | `VISION.md` |
| Solo governance, PR rules | `CONTRIBUTING.md` |
