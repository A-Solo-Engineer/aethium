# Aethium

A minimal, high-performance UI framework for Go with immediate-mode rendering, supporting both browser (Wasm) and desktop targets.

## Features

- **Immediate-mode rendering** via a draw command stream
- **TinyGo for browser** (≤ 500 KB gzipped)
- **Go 1.22+ for desktop** (≤ 5 MB binary)
- **Single codebase** for web and desktop
- **Zero dependencies** in core framework
- **AGPL-3.0** licensed

## Quick Start

```bash
# Install the CLI
go install github.com/A-Solo-Engineer/aethium/cmd/aethium@latest

# Create a new project
aethium new --module example.com/myapp

# Build for desktop
aethium build --target desktop --output dist/

# Build for browser
aethium build --target wasm --output dist/
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Build System](docs/BUILD_SYSTEM.md)
- [Implementation Plan](docs/IMPLEMENTATION_PLAN.md)
- [State Management](docs/STATE_MANAGEMENT.md)
- [Usage Guide](docs/USAGE.md)
- [Support](docs/SUPPORT.md)

## License

AGPL-3.0
