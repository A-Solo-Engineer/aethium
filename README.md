# Aethium

A minimal, high-performance UI framework for Go with immediate-mode rendering, supporting both browser (Wasm) and desktop targets.

---

## Framework Overview

| Feature | Browser (Wasm) | Desktop (Native) |
|---------|----------------|------------------|
| **Toolchain** | TinyGo 0.34.0 | Go 1.22+ |
| **Binary Size** | ≤ 500 KB (gzipped) | ≤ 5 MB |
| **Rendering** | WebGL2 Canvas | System WebView (WebView2/WebKit) |
| **Threading** | Single-threaded | Main thread (UI) + Background Workers |

---

## Core Principles

- **Single Codebase**: Write once in Go, run on the web and desktop with identical visuals.
- **Immediate-Mode**: No DOM, no widget tree. Pure draw command stream for maximum performance.
- **Fine-Grained Reactivity**: Signal-based updates ensure only the necessary parts of the UI re-render.
- **Zero Dependencies**: The core framework is built entirely on the Go standard library and JS built-ins.

---

## Quick Start

```bash
# Install the CLI
go install github.com/A-Solo-Engineer/aethium/cmd/aethium@latest

# Create a new project
aethium new --module example.com/myapp

# Build for your target
aethium build --target desktop  # For Windows/macOS/Linux
aethium build --target wasm     # For the Web
```

---

## Documentation

| Guide | Description |
|-------|-------------|
| [Architecture](docs/ARCHITECTURE.md) | Technical design, rendering strategy, and system diagrams. |
| [State Management](docs/STATE_MANAGEMENT.md) | Signals, computed values, effects, and memory pooling. |
| [Usage Guide](docs/USAGE.md) | API reference, component lifecycle, and examples. |
| [Build System](docs/BUILD_SYSTEM.md) | TinyGo optimizations and platform-specific build flags. |

---

## Project Structure

```mermaid
graph LR
    A[App Code] --> B[Reactive Context]
    A --> C[Scene Graph]
    B --> D[Runtime]
    C --> D
    D --> E[DrawList]
    E --> F[Platform Backend]
```

---

## License

AGPL-3.0
