# Aethium — Build System

Contract for the Aethium CLI (`aethium`), outputs, dev server, dependencies, and **Resource Efficiency Budget** (hard exit criteria for Stage 2).

Aligned with `ARCHITECTURE.md`: TinyGo for browser, Go 1.22+ for desktop, WebView host, immediate-mode canvas.

---

## CLI Tool Design

Binary name: **`aethium`**. Module: `github.com/aethium-dev/aethium/cmd/aethium`.

### `aethium new`

Scaffold a project.

| Flag | Default | Description |
|------|---------|-------------|
| `--module` | (required) | Go module path, e.g. `example.com/myapp` |
| `--template` | `minimal` | `minimal` only in Stage 2 |
| `--dir` | `.` | Output directory |

**Output:**

```
<dir>/
  go.mod
  main.go
  app/view.go
  aethium.toml
```

### `aethium build`

Production compile.

| Flag | Default | Description |
|------|---------|-------------|
| `--target` | `desktop` | `wasm`, `desktop`, `all` |
| `--release` | `true` | Optimizations on |
| `--output` | `dist/` | Artifact directory |
| `-tags` | `` | Passed to Go/TinyGo |
| `--tinygo` | auto | `auto` uses TinyGo for `wasm`, Go for `desktop` |

**Output (by target):**

| Target | Artifacts under `--output` |
|--------|---------------------------|
| `wasm` | `app.wasm`, `app.wasm.map` (if debug), `index.html`, `aethium.js` (glue, < 15 KB gzipped) |
| `desktop` | `myapp.exe` / `myapp` (OS binary), `resources/` (embedded wasm + html if not single-file) |
| `all` | Both trees above |

Exit code `0` only if post-build size checks pass (see Budget) or `--skip-budget-check` with maintainer exception file present.

### `aethium dev`

Development server with hot reload.

| Flag | Default | Description |
|------|---------|-------------|
| `--target` | `wasm` | `wasm` or `desktop` |
| `--addr` | `127.0.0.1:5173` | Bind address |
| `--open` | `false` | Open browser |
| `--tinygo-opt` | `0` | TinyGo opt level in dev (0 = fast compile) |

**Output:** Running process; logs compile time and reload latency.

### `aethium bundle`

Single-file distributable.

| Flag | Default | Description |
|------|---------|-------------|
| `--target` | `desktop` | `desktop` or `wasm` |
| `--format` | `native` | `native` (exe), `wasm` (standalone html+inline wasm) |
| `--output` | `dist/bundle` | File path |

**Output:**

| Format | File |
|--------|------|
| `native` | Single executable with `embed`ded `index.html` + `app.wasm` |
| `wasm` | Single `index.html` with base64 or inline wasm (for itch.io-style hosting) |

---

## Output Targets

### (a) Browser / Wasm

- `index.html` — canvas + loader
- `aethium.js` — host bridge (< 15 KB gzipped)
- `app.wasm` — TinyGo build
- Optional: hashed assets from `assets/` via `embed`

Deploy to any static host; no server required for default apps.

### (b) Desktop binary

- Native executable linking WebView loader
- Embedded or adjacent `resources/` (wasm + html)
- No Electron, no bundled Chromium

### (c) Self-contained single-file executable

- `aethium bundle --target desktop --format native -o MyApp.exe`
- All runtime assets embedded via `go:embed`
- **Must** still satisfy desktop binary ≤ 5 MB budget (compressed on disk measurement)

---

## Dev Server — Hot Reload (Wasm)

### What recompiles

| Change | Action |
|--------|--------|
| `*.go` under module | Incremental **TinyGo** rebuild of `app.wasm` only |
| `assets/*` | Copy + refresh; no Go rebuild |
| `index.html`, `aethium.toml` | Inject reload; no wasm rebuild unless config changes compiler flags |

### What is injected

1. WebSocket or SSE channel from `aethium dev` to browser.
2. On successful build, server sends `{ "type": "reload", "wasm": "/app.wasm?v=<hash>" }`.
3. Host JS swaps Wasm module via `WebAssembly.instantiateStreaming` (fallback: array buffer) **without** full page reload when possible.

### Reload latency target

| Metric | Target |
|--------|--------|
| TinyGo incremental rebuild (hello-world, dev opt 0) | ≤ **800 ms** p95 on mid-range hardware |
| Browser swap to painted frame | ≤ **200 ms** after wasm ready |
| **End-to-end** save → visible update | ≤ **1000 ms** p95 |
| Desktop `aethium dev` full reload (Go binary relink) | ≤ **3 s** p95 — **documented exception:** full native binary relink required; not comparable to Wasm hot path |

---

## Dependency Philosophy

**Framework core (`github.com/aethium-dev/aethium/...`):**

- **Minimal allowlist** — stdlib-first; third-party modules require RFC + size audit.
- **Allowed without RFC (Stage 2):**
  - `github.com/webview/webview` (desktop host only; not linked in TinyGo wasm build)
- **Forbidden in core:** heavy ORMs, `gin`/`echo`, `cobra` (CLI uses stdlib `flag`), `viper`, anything that pulls `google.golang.org/grpc` transitively.
- **User apps:** unrestricted `go.mod`; framework does not police app dependencies.

**Rationale:** Supply-chain and binary-size risk concentrate in the engine; AGPL core must be auditable in < 1 day.

**TinyGo:** Core packages must compile under TinyGo 0.34.0; CI runs `tinygo build` on every core PR.

---

## Resource Efficiency Budget

**These are exit criteria for Stage 2**, not aspirations. Failure requires documented architectural exception **approved before merge** (issue label `budget-exception`, maintainer sign-off).

Reference implementation for measurement: `examples/hello` (counter button, one label, ~80 lines Go, one font baked as minimal bitmap in embed).

### Measurement methodology

| Metric | How measured |
|--------|----------------|
| Desktop binary | Release `aethium bundle -o hello`; file size on disk |
| Wasm bundle | `gzip -c dist/wasm/app.wasm + aethium.js + index.html` total |
| Cold startup desktop | Process start → first `DrawCmd` presented; 50 runs, p50 |
| Cold startup browser/Wasm | Navigation start → interactive canvas; Chrome DevTools Performance panel; no CPU throttle, no network throttle; 50 runs, p50 |
| Idle RAM desktop | After 5 s idle, resident set size (Windows Task Manager “Memory", macOS `footprint`) |

### Targets vs baselines

| Criterion | Aethium exit (max) | Baseline | Documented delta |
|-----------|---------------------|----------|------------------|
| Desktop binary | **≤ 5 MB** | Electron packaged hello ~**120 MB** | **~115 MB smaller (~96%)** |
| Browser/Wasm bundle | **≤ 500 KB** gzipped | Create React App production ~**180–250 KB** JS + ~**50 KB** runtime chunk ≈ **230–300 KB** gzipped (JS only); full page with html ~**280–350 KB** | Aethium target **≤ 500 KB** total wasm+glue+html; comparable band with **no JS engine** |
| Cold startup (desktop) | **≤ 150 ms** p50 | Electron hello ~**800 ms–2000 ms** | **≥ 5× faster** |
| Cold startup (browser/Wasm) | **≤ 300 ms** to interactive canvas | Typical SPA hydration **≥ 1–3 s** | **≥ 3× faster** |
| Idle RAM (desktop hello) | **≤ 40 MB** | Electron ~**150–250 MB** | **≥ 60% reduction** |

### Per-package budget allocations (Stage 2 planning)

Derived from ceilings; implementation plan must not exceed without exception.

| Package | Desktop binary budget | Wasm gzip budget |
|---------|----------------------|------------------|
| `cmd/aethium` + embed | ≤ 1.2 MB | N/A |
| `platform/webview` | ≤ 0.8 MB | N/A |
| `platform/js` + glue | N/A | ≤ 80 KB |
| `runtime` + `reactive` + `scene` | ≤ 1.5 MB | ≤ 250 KB |
| `canvas` + fonts (minimal) | ≤ 1.0 MB | ≤ 120 KB |
| User app (`examples/hello`) | ≤ 0.5 MB | ≤ 50 KB |

### Stage 2 gate linkage

Before Stage 2 coding begins, maintainer approves architecture docs. **During Stage 2**, any PR merging without meeting budgets must include:

1. `docs/exceptions/<issue>.md` — root cause, mitigation plan, re-target date
2. Updated table above with revised numbers

**Stage 3** replaces projections with measured **Benchmark Report** (pass/fail).

---

## Benchmark Report (Stage 3 placeholder)

_Implemented after Stage 2. Template:_

```markdown
## Benchmark Report (measured YYYY-MM-DD)

| Criterion | Target | Measured | Pass |
|-----------|--------|----------|------|
| Desktop binary | ≤ 5 MB | TBD | TBD |
| Wasm gzip total | ≤ 500 KB | TBD | TBD |
| Cold startup desktop p50 | ≤ 150 ms | TBD | TBD |
```

---

## CLI ↔ architecture consistency

| CLI behavior | Architecture |
|--------------|--------------|
| `wasm` → TinyGo | `ARCHITECTURE.md` Wasm toolchain |
| `desktop` → Go + WebView | Desktop target strategy |
| Draw pipeline | Immediate-mode `canvas.DrawList` |
| Hot reload UI queue | `runtime.ScheduleOnUI` |

**Stage gate:** No code until **`APPROVED — proceed to Stage 2`**. Stage 2 begins with `docs/IMPLEMENTATION_PLAN.md`, not source files.
