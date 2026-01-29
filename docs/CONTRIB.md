# Contributing Guide - FeCIM Lattice Tools

> **Source of Truth:** `go.mod`, `launch.sh`, `CLAUDE.md`, `docs/development/WORKFLOWS.md`

## Quick Start

```bash
# Clone and build
git clone https://github.com/your-org/fecim-lattice-tools.git
cd fecim-lattice-tools
go mod download
./launch.sh
```

## Development Environment

### System Requirements

| Requirement | Version | Check Command |
|-------------|---------|---------------|
| Go | 1.24+ | `go version` |
| GCC | Any | `gcc --version` |
| OpenGL libs | - | See below |

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install -y \
  gcc libgl1-mesa-dev libx11-dev libxinerama-dev \
  libxrandr-dev libxcursor-dev libxi-dev libxext-dev libxfixes-dev
# Optional: for Module 6 Yosys schematic visualization
sudo apt-get install -y graphviz
```

**Linux (Fedora/RHEL):**
```bash
sudo dnf install -y gcc mesa-libGL-devel libX11-devel libXcursor-devel \
  libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
```

**macOS:**
```bash
xcode-select --install
```

**Windows:**
1. Install [MSYS2](https://www.msys2.org/) or [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
2. Ensure `gcc` is in your PATH

### Optional Dependencies

| Tool | Purpose | Installation |
|------|---------|--------------|
| `ffmpeg` | Video recording | `sudo apt install ffmpeg` |
| `graphviz` | Yosys schematic visualization | `sudo apt install graphviz` |
| `docker` | OpenLane/OpenROAD (Module 6 EDA) | [Docker Install](https://docs.docker.com/get-docker/) |
| `yosys` | Verilog synthesis (Module 6) | `sudo apt install yosys` |
| `openroad` | Physical design (Module 6) | Via Docker recommended |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FYNE_DEBUG_RESIZE` | `0` | Debug layout issues (set to `1`) |
| `FYNE_NO_GL` | `0` | Disable GPU rendering (software fallback) |
| `FYNE_THEME` | `dark` | UI theme (`dark` or `light`) |
| `FYNE_SCALE` | auto | DPI scaling (e.g., `1.5` for HiDPI) |
| `GDK_BACKEND` | auto | Force X11: `GDK_BACKEND=x11` |

**Note:** No `.env` file required. Use CLI flags instead:
```bash
./fecim-lattice-tools --verbosity 2  # 0=off, 1=info, 2=debug, 3=trace
```

## Available Scripts

### Core Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `launch.sh` | Build and run GUI | `./launch.sh [--verbosity LEVEL]` |
| `scripts/build-all.sh` | Build all standalone modules | `./scripts/build-all.sh` |
| `commit-push.sh` | Scheduled git commits | `./commit-push.sh -12` (in 12h) |

### Module-Specific Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `module3-mnist/scripts/train_all_sizes.sh` | Train MNIST networks (h64, h128, h256) | `./train_all_sizes.sh` |
| `module3-mnist/scripts/benchmark.sh` | Compare simulation vs literature | `./benchmark.sh` |
| `module6-eda/examples/01-basic-8x8/run.sh` | EDA 8x8 crossbar example | `./run.sh` |
| `module6-eda/examples/02-mnist-layer/run.sh` | EDA MNIST layer example | `./run.sh` |
| `module1-hysteresis/shaders/compile.sh` | Compile Vulkan shaders | `./compile.sh` |

### Build Commands

```bash
# Standard build
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Release build (optimized, smaller binary)
go build -ldflags="-s -w" -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Debug build with race detector
go build -race -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Cross-compile
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o fecim-linux ./cmd/fecim-lattice-tools
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o fecim-mac ./cmd/fecim-lattice-tools
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o fecim.exe ./cmd/fecim-lattice-tools
```

## Testing

### Run All Tests

```bash
go test ./...                    # All 117 tests
go test -v ./...                 # Verbose output
go test -race ./...              # With race detector
go test -cover ./...             # With coverage
```

### Run Module-Specific Tests

```bash
go test ./module2-crossbar/pkg/crossbar -v   # Crossbar tests
go test ./module3-mnist/pkg/core -v          # MNIST core tests
go test -run TestQuantization ./...          # Specific test pattern
```

### Generate Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Counts by Package

| Package | Tests |
|---------|-------|
| module1-hysteresis/pkg/ferroelectric | 7 |
| module1-hysteresis/pkg/simulation | 4 |
| module2-crossbar/pkg/crossbar | 29 |
| module3-mnist/pkg/core | 33 |
| module3-mnist/pkg/training | 3 |
| module4-circuits/pkg/peripherals | 9 |
| module5-comparison/pkg/comparison | 19 |
| module6-eda/pkg/export | 3 |
| shared/* | 12 |
| **Total** | **117** |

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature
```

### 2. Make Changes

Follow the module structure:
```
moduleN-name/
├── cmd/           # CLI entry points
├── pkg/
│   ├── core/      # Core logic
│   ├── gui/       # Fyne GUI (embedded.go required)
│   └── physics/   # Physics simulation
└── README.md
```

### 3. Run Tests

```bash
go test -race -v ./... && echo "All tests passed"
```

### 4. Commit

```bash
git commit -m "feat: add feature description"
```

**Commit Types:** `feat`, `fix`, `docs`, `test`, `refactor`, `style`, `chore`, `perf`

### 5. Create Pull Request

Ensure:
- [ ] All tests pass (`go test ./...`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Code formatted (`go fmt ./...`)
- [ ] Code vetted (`go vet ./...`)
- [ ] Documentation updated if needed

## Critical Rules

### Thread Safety

**ALL UI updates from goroutines MUST use `fyne.Do()`:**

```go
// WRONG - Race condition
go func() {
    label.SetText("Updated")
}()

// CORRECT - Thread-safe
go func() {
    fyne.Do(func() {
        label.SetText("Updated")
    })
}()
```

### 30-Level Quantization

Always use the standard quantization function:
```go
import "fecim-lattice-tools/module2-crossbar/pkg/crossbar"

level := crossbar.QuantizeTo30Levels(value)
```

### Embedded App Interface

New modules must implement:
```go
type EmbeddedApp struct { ... }

func (e *EmbeddedApp) BuildContent(app fyne.App, win fyne.Window) fyne.CanvasObject
func (e *EmbeddedApp) Start()
func (e *EmbeddedApp) Stop()
```

### Do NOT Modify

- Binaries - Never commit compiled binaries

## Dependencies

### Direct Dependencies (from go.mod)

| Package | Version | Purpose |
|---------|---------|---------|
| fyne.io/fyne/v2 | v2.7.2 | GUI framework |
| charmbracelet/bubbletea | v1.2.4 | TUI (module1 only) |
| charmbracelet/lipgloss | v1.0.0 | TUI styling |
| go-gl/glfw | v3.3 | OpenGL bindings |
| vulkan-go/vulkan | v0.0.0 | Vulkan rendering |

**Go Version:** 1.24+ (toolchain go1.24.12)

### Update Dependencies

```bash
go get -u ./...      # Update all
go mod tidy          # Clean up
```

## Project Structure

```
fecim-lattice-tools/
├── cmd/fecim-lattice-tools/  # Unified GUI entry point
├── module1-hysteresis/       # P-E curve, Preisach model
├── module2-crossbar/         # MVM, non-idealities (4 tabs)
├── module3-mnist/            # Neural network demo
├── module4-circuits/         # DAC/ADC/TIA peripherals
├── module5-comparison/       # Technology comparison (technical briefing)
├── module6-eda/              # EDA design suite
├── shared/                   # Theme, widgets, logging
├── docs/
│   ├── development/          # Dev docs (WORKFLOWS.md, TESTING.md)
│   └── cim/                  # Physics docs (HONESTY_AUDIT.md)
└── scripts/                  # Build scripts
```

## Getting Help

| Topic | Resource |
|-------|----------|
| Function lookup | `docs/development/scriptReference.md` |
| Error resolution | `docs/development/scriptReference.md#error-resolution-guide` |
| Fyne patterns | `docs/development/GUI/FYNE_NOTES.md` |
| Physics accuracy | `docs/cim/HONESTY_AUDIT.md` |
| Testing | `docs/development/TESTING.md` |
| Project rules | `CLAUDE.md` |

---

**Last Synced:** 2026-01-26 | **Source:** go.mod (Go 1.24.12), shell scripts, CLAUDE.md
