# Aethium — Vision

**Aethium** is an ultra-lightweight application framework in Go for building web and desktop applications with a fraction of the memory, startup time, and binary size of Electron/Node.js stacks.

---

## Mission

Aethium exists to prove that a full application shell—routing, reactive UI, asset loading, and desktop packaging—can ship without dragging a browser engine, a JavaScript runtime, or hundreds of megabytes of transitive dependencies.

**Anti-bloat philosophy (concrete):**

| Metric | Aethium target (exit criteria) | Baseline to beat | Delta |
|--------|-------------------------------|------------------|-------|
| Idle RAM (hello-world, desktop) | ≤ 40 MB resident | Electron ~150–250 MB typical | **≥ 75% reduction** |
| Desktop binary (hello-world, packaged) | ≤ 5 MB | Electron app ~120 MB (runtime + app) | **~96% smaller** |
| Browser/Wasm bundle (hello-world) | ≤ 500 KB gzipped | React + Webpack SPA ~250–400 KB gzipped (app only; excludes devtools) | **Comparable or smaller total deliverable** |
| Cold startup (desktop, mid-range) | ≤ 150 ms to first frame | Electron ~800 ms–2 s | **≥ 5× faster** |
| Cold startup (browser/Wasm, mid-range) | ≤ 300 ms to interactive canvas | Typical SPA hydration 1–3 s | **≥ 3× faster** |

“Mid-range hardware” for measurement: 4-core CPU (e.g. Intel i5-8250U class), 8 GB RAM, SSD, 1080p display, Windows 11 or macOS 13+, Chrome 120+ for Wasm path.

Success is not “feels fast”—it is **documented pass/fail** against the Resource Efficiency Budget in `BUILD_SYSTEM.md`. Stage 2 implementations that miss a ceiling require a written architectural exception **before merge**.

---

## Why Go

Aethium is not choosing Go for slogan-level speed. It is choosing Go because three structural properties align with a small, concurrent UI runtime:

1. **Compilation model** — Go produces a single static binary (native desktop) or a self-contained Wasm module without a separate bytecode VM, JIT warmup, or `node_modules` resolution graph. The linker dead-strips unused packages; the deployable artifact is predictable in size. That directly supports the ≤ 5 MB desktop and ≤ 500 KB gzipped Wasm exit criteria in ways JavaScript ecosystems cannot match without heroic bundling and still shipping a full JS engine.

2. **Goroutine scheduler** — UI work must stay on the browser/OS event thread, but network, decode, and file I/O must not block it. Go’s M:N scheduler plus channels give a first-class pattern: bounded worker pools, backpressure, and cancellation (`context.Context`) without callback pyramids. Rust can do this; JavaScript defers to the single main thread unless Workers are introduced (second runtime, serialization tax).

3. **`sync.Pool` and value-oriented APIs** — Reactive UI churn allocates many short-lived nodes and draw records. Go’s `sync.Pool` is designed exactly for amortizing allocation across frames without manual free lists. Combined with signal-granular updates (see `STATE_MANAGEMENT.md`), the GC sees fewer young-gen spikes than immutable tree diffing or per-frame DOM wrapper allocation.

Go’s GC is non-generational in the same way as Rust’s ownership, but for a UI framework targeting **tens of MB RAM**, not microsecond tail latency in a game engine, Go’s trade-off is acceptable—and far cheaper in engineering time than maintaining an auditable unsafe UI core in Rust or C++.

---

## Why Not X

### Rust

**Rejected as foundation.** Rust offers maximal performance and zero-cost abstractions, but Aethium optimizes for **time-to-ship and contributor velocity** on a small team. A cross-platform UI framework in Rust implies `unsafe` FFI to every platform, async runtime choice fragmentation, and compile times that punish the hot-reload budget in `BUILD_SYSTEM.md`. Rust remains acceptable for **optional** native acceleration crates later; it is not the core language.

### Zig

**Rejected as foundation.** Zig’s comptime and C interop are compelling for systems tooling, but the ecosystem lacks mature, cross-platform UI/Wasm packaging, devtools, and hiring depth. Aethium would spend Stage 2 building fundamentals (package manager ergonomics, Wasm story) instead of the framework.

### C++

**Rejected as foundation.** C++ minimizes runtime size but maximizes memory-unsafety and build complexity (CMake, ABI, STL across platforms). Electron’s bloat is not caused by C++—it is caused by shipping Chromium. Aethium achieves small binaries via Go/TinyGo + immediate-mode rendering without taking C++’s safety and CI cost.

### Flutter

**Rejected as foundation.** Flutter ships its own rendering engine and Dart VM—excellent product, wrong fit for **Go-native** teams and the stated Wasm/desktop unified story. Binary sizes for minimal Flutter desktop apps are typically **15–30 MB+**, failing the ≤ 5 MB desktop exit criterion without heroic stripping. Aethium targets developers who want Go source, Go modules, and Go tooling end-to-end.

---

## License Decision

**Chosen license: GNU Affero General Public License v3.0 (AGPL-3.0).**

**Non-negotiable requirement:** Any party deploying a modified version of the Aethium core engine as a network service must publish their modifications.

### Evaluation

| License | Network service modification disclosure | Verdict |
|---------|----------------------------------------|---------|
| MIT | No obligation to share modifications when software runs on a server users interact with over a network. Permits proprietary SaaS wrappers around a forked engine. | **Rejected** |
| Apache 2.0 | Same as MIT for network use. Patent grant is valuable but does not close the “cloud capture” loophole. Section 4 redistribution obligations apply to **distribution of copies**, not to **making modified software available for remote interaction** in the AGPL sense. | **Rejected** |
| AGPL-3.0 | Section 13 requires offering corresponding source to users who interact with the modified program over a network. Forks used as multi-tenant UI platforms cannot keep engine changes proprietary. | **Selected** |

### Why MIT and Apache 2.0 fail (precisely)

- **MIT** only requires preserving copyright notice in copies. A company may fork Aethium, improve the Wasm renderer, deploy it as `ui.example.com` for thousands of tenants, and **never** release those improvements—users receive interaction, not a binary “copy” in the traditional sense. That is the **Oracle/cloud-capture scenario**: monetize a better engine while the commons stagnates.

- **Apache 2.0** adds patent retaliation and NOTICE file requirements but **does not** trigger source distribution when the program is run as a service. Hosted “Aethium-as-a-platform” with closed-source patches remains compliant.

### How AGPL-3.0 prevents cloud capture

AGPL-3.0 treats **network interaction** as a distribution trigger for the modified version’s source. If a vendor ships a managed desktop/browser IDE built on a modified Aethium core, they must provide source for their engine changes to users of that service—closing the loophole that allowed AWS-era open-core capture of MongoDB/Elasticsearch patterns before SSPL/AGPL adoption.

**Trade-off accepted:** AGPL is incompatible with some corporate policies and “link in proprietary app” models without source release. Aethium prioritizes **commons sustainability** over maximum corporate adoption.

---

## Non-Goals

Aethium will **never**:

1. **Embed or ship Chromium/Electron** — No WebView-as-the-primary-renderer for the browser target; desktop WebView is a host shell only (see `ARCHITECTURE.md`).

2. **Provide a JavaScript/TypeScript plugin runtime** — No user-extensible JS scripting layer; extensions are Go modules compiled into the app.

3. **Implement a full design tool or WYSIWYG builder** — Layout is code-first; visual builders are out of scope.

4. **Ship a batteries-included component library rivaling Material/Chakra** — Core ships primitives (canvas, text, input hit regions); rich widget kits are community or app responsibility.

5. **Support legacy browsers (IE, pre-es6module)** — Wasm + WebGL2 baseline only.

6. **Guarantee bit-identical rendering across GPU drivers** — Immediate-mode canvas targets consistency “best effort”; print/PDF perfection is a non-goal.

7. **Replace the Go standard library or invent a package manager** — `go mod` remains the dependency backbone.

---

## Document map

| Document | Purpose |
|----------|---------|
| `ARCHITECTURE.md` | Rendering, Wasm toolchain, concurrency, lifecycle, desktop strategy |
| `STATE_MANAGEMENT.md` | Signals, pools, derived state, locking |
| `BUILD_SYSTEM.md` | CLI contract, budgets, dev server |
| `CONTRIBUTING.md` | Governance, PRs, AGPL contribution policy |

**Stage gate:** No framework or application code until the maintainer types: **`APPROVED — proceed to Stage 2`**.
