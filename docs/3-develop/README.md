# Developer Guide

**Everything you need to build, extend, and contribute to FeCIM Lattice Tools.**

---

## 🏁 Quick Start for Developers

```bash
# Clone
git clone https://github.com/[your-repo]/fecim-lattice-tools.git
cd fecim-lattice-tools

# Install dev dependencies (Linux)
sudo apt-get install -y git

# Build
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Run tests
go test ./...

# Run with race detection
go test -race ./...
```

Default UI shell: `gogpu/ui`. The canonical app builds with `CGO_ENABLED=0` and routes UI state through `shared/viewmodel`.

Legacy Fyne shell: `cmd/fecim-lattice-tools-fyne` with `-tags legacy_fyne`. Use it only for temporary parity checks or legacy GUI maintenance.

---

## 📖 Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [api-reference.md](api-reference.md) | All public APIs (7 packages) | All developers |
| [code-quality.md](code-quality.md) | Standards and linting | Contributors |
| [memory-optimization.md](memory-optimization.md) | Memory profiling and tuning | Performance |
| [accessibility.md](accessibility.md) | UI accessibility standards | GUI developers |
| [repo-health.md](repo-health.md) | CI status and health metrics | Maintainers |
| [architecture/](architecture/) | System architecture docs | All developers |
| [automation/](automation/) | Build and CI automation | DevOps |
| [gui/](gui/) | GUI development guides | GUI developers |
| [testing/](testing/) | Testing methodology | All developers |
| [**FeCIM Skills**](../../tools/fecim-skills/README.md) | Agent skills for Claude Code, Codex, opencode | All developers |

---

## 🏗️ Architecture Overview

### Repository Structure

```
fecim-lattice-tools/
├── cmd/
│   ├── fecim-lattice-tools/    # Default zero-CGO gogpu/ui entry point
│   ├── fecim-lattice-tools-fyne/ # Tagged legacy Fyne entry point
│   └── latex-svg/              # LaTeX-to-SVG utility
├── module1-hysteresis/         # P-E curves, Preisach model
│   └── pkg/
│       ├── ferroelectric/      # Physics engine
│       ├── controller/         # ISPP write controller
│       └── gui/                # Tagged legacy Fyne adapter
├── module2-crossbar/           # MVM and non-idealities
│   └── pkg/crossbar/           # Crossbar array simulation
├── module3-mnist/              # Neural network inference
│   └── pkg/core/               # Dual-mode network
├── module4-circuits/           # Peripheral circuits
│   └── pkg/
│       ├── arraysim/           # Array simulation
│       └── gui/                # Tagged legacy Fyne adapter
├── module5-comparison/         # Technology comparison
├── module6-eda/                # EDA tools
├── shared/
│   ├── physics/                # Core physics models
│   ├── peripherals/            # DAC, ADC, TIA models
│   ├── viewmodel/              # UI-neutral state bridge
│   ├── io/                     # File I/O utilities
│   ├── widgets/                # Tagged legacy Fyne components
│   ├── theme/                  # Tagged legacy Fyne styling
│   └── logging/                # Structured logging
└── data/
    └── calibrations/           # Material calibration data
```

### Key Design Patterns

**UI Boundary**

Every module exposes UI-neutral state through `shared/viewmodel`. The default `gogpu/ui` shell renders those state snapshots and dispatches actions back to the viewmodel layer:

```go
type ModuleViewModel interface {
    Snapshot() ModuleState
    Dispatch(action ModuleAction) error
}
```

Tagged legacy Fyne adapters may still use `BuildContent`, `Start`, and `Stop`, but that API is not the canonical module contract for new UI work.

**Thread-Safe UI State**

Default UI work should move data through viewmodel snapshots instead of mutating widgets from simulation goroutines. Tagged legacy Fyne adapters still need `fyne.Do()` when a goroutine touches a Fyne widget:

```go
// Legacy Fyne adapter only
go func() {
    result := heavyComputation()
    fyne.Do(func() {
        label.SetText(result)
    })
}()

// WRONG - will cause race condition
go func() {
    result := heavyComputation()
    label.SetText(result) // never call UI directly from goroutine
}()
```

**30-Level Quantization**

The default simulation baseline uses 30 discrete conductance levels:

```go
// Quantize a normalized value [0,1] to 30 levels
quantized := crossbar.QuantizeTo30Levels(value)
```

This is configurable per material. See [api-reference.md#quantization-functions](api-reference.md#quantization-functions).

---

## 🔧 Core APIs

### Package Overview

| Package | Import Path | Purpose |
|---------|------------|---------|
| `shared/physics` | `fecim-lattice-tools/shared/physics` | Material models, L-K solver, WriteController |
| `shared/peripherals` | `fecim-lattice-tools/shared/peripherals` | DAC, ADC, TIA, charge pump |
| `shared/io` | `fecim-lattice-tools/shared/io` | JSON and file utilities |
| `shared/viewmodel` | `fecim-lattice-tools/shared/viewmodel` | UI-neutral state and actions |
| `shared/widgets` | `fecim-lattice-tools/shared/widgets` | Tagged legacy Fyne components |
| `ferroelectric` | `fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric` | Hysteresis, Preisach model |
| `crossbar` | `fecim-lattice-tools/module2-crossbar/pkg/crossbar` | Crossbar array simulation |
| `core` (MNIST) | `fecim-lattice-tools/module3-mnist/pkg/core` | Neural network inference |

Full API documentation: [api-reference.md](api-reference.md)

### Quick Examples

**Material + Write Control:**
```go
mat := physics.FeCIMMaterial()
solver := physics.NewLKSolver()
solver.ConfigureFromMaterial(mat)
wc := physics.NewWriteController(solver, mat)
attempts, ok, _ := wc.WriteTarget(60e-6) // target 60 µS
```

**Crossbar MVM:**
```go
arr, _ := crossbar.NewArray(&crossbar.Config{Rows: 4, Cols: 4})
defer arr.Destroy()
arr.ProgramWeightMatrix(weights)
output, _ := arr.MVM(inputVector)
```

**DAC/ADC Round-Trip:**
```go
dac := peripherals.DefaultDAC()
adc := peripherals.DefaultADC()
voltage := dac.Convert(12)        // level 12 → voltage
code := adc.Convert(voltage)      // voltage → digital code
```

---

## 🧪 Testing

### Running Tests

```bash
# All tests
go test ./...

# Single package
go test ./module2-crossbar/pkg/crossbar

# With race detector
go test -race ./...

# With coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./module2-crossbar/pkg/crossbar
```

### Test Structure

| Test Location | What It Tests |
|--------------|---------------|
| `module1-hysteresis/pkg/controller/*_test.go` | ISPP write controller (9 files) |
| `cmd/fecim-lattice-tools/mode_engine_matrix_test.go` | Headless ISPP (9 materials × 2 engines) |
| `cmd/fecim-lattice-tools/mode_preisach_target_progression_test.go` | Preisach state targeting |
| `module4-circuits/pkg/arraysim/*_test.go` | Array simulation regression |
| `shared/physics/*_test.go` | Physics model unit tests |
| `shared/peripherals/*_test.go` | Peripheral model unit tests |

### Updating Golden Files

Physics regression golden files must be regenerated when physics changes:

```bash
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...
```

Full testing guide: [testing/](testing/)

---

## 📏 Code Standards

### Required Before Each Commit

```bash
# Format
go fmt ./...

# Vet
go vet ./...

# Test
go test ./...

# Race check
go test -race ./...
```

### Key Rules

1. **UI boundary:** Keep simulation and viewmodel packages free of Fyne and `gogpu/ui` imports
2. **30-level quantization:** Use `crossbar.QuantizeTo30Levels(value)` for canonical form
3. **Default shell:** Implement new UI behavior in `internal/gogpuapp` using `shared/viewmodel`
4. **Legacy Fyne:** Use `fyne.Do(func() { ... })` only in `-tags legacy_fyne` adapters
5. **No binaries committed:** Never commit compiled binaries

### Commit Format

```
type: description

Types: feat, fix, docs, refactor, test, chore
Examples:
  feat: add ISPP method selector to Module 4
  fix: correct Preisach minor loop calculation
  docs: add package doc to shared/physics
  test: add ensemble ISPP convergence test
```

Full standards: [code-quality.md](code-quality.md)

---

## 🖥️ UI Development

### Default gogpu/ui Rules

The canonical shell lives under `internal/gogpuapp` and renders state from `shared/viewmodel`. New UI work should keep business logic and simulation state out of shell packages.

### Legacy Fyne Rules

Tagged legacy Fyne adapters require all widget operations from goroutines to go through `fyne.Do()`:

```go
// The safe pattern for long-running goroutines:
go func() {
    // Do computation here (outside fyne.Do)
    result := computeHeavyThing()

    // Update UI here (inside fyne.Do)
    fyne.Do(func() {
        myLabel.SetText(result)
        myProgressBar.SetValue(1.0)
    })
}()
```

### Module Viewmodel Pattern

Each module's canonical UI state belongs in `shared/viewmodel/<module>/`. The standard shape is:

```go
type State struct {
    // renderable fields for the shell
}

type Model struct {
    // module-specific state and services
}

func (m *Model) Snapshot() State {
    // return immutable render state
}

func (m *Model) Dispatch(action Action) error {
    // update model state from shell events
}
```

GUI development guide: [gui/](gui/)

Fyne-specific notes for tagged legacy adapters: [gui/FYNE_NOTES.md](gui/FYNE_NOTES.md)

---

## 🔌 Extending the Simulator

### Add a New Material

1. Add parameters to `shared/physics/material.go`
2. Create constructor function following existing pattern
3. Add to `AllMaterials()` list
4. Regenerate calibration: `go run ./cmd/fecim-lattice-tools --calibrate --material your_material`
5. Update golden test data if physics changed

### Add a New Module

1. Create `moduleN-name/` directory with standard structure
2. Add a UI-neutral viewmodel under `shared/viewmodel/<module>/`
3. Register the default shell view in `internal/gogpuapp`
4. Add package docs in `pkg/*/doc.go` files
5. Write tests covering core functionality
6. Document in `docs/2-learn/moduleN-name/`

### Add a New Non-Ideality

1. Implement effect in `module2-crossbar/pkg/crossbar/`
2. Add to `MVMOptions` struct
3. Update `MVMWithNonIdealities()` pipeline
4. Add regression test with golden data
5. Document in [api-reference.md](api-reference.md)

---

## 🐛 Debugging

### Enable Verbose Logging

```bash
./fecim-lattice-tools --logger --verbosity debug
```

Log levels: `off` | `info` | `debug` | `trace`

Log files: `logs/` directory with datetime stamps.

### Race Detection

```bash
go build -race -o fecim-race ./cmd/fecim-lattice-tools
./fecim-race
# Race detector reports go to stderr
```

### Memory Profiling

```go
// Temporarily add to main.go
import _ "net/http/pprof"
import "net/http"

func init() {
    go http.ListenAndServe("localhost:6060", nil)
}
```

Then:
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

### Common Errors

| Error | Cause | Fix |
|-------|-------|-----|
| Default app build unexpectedly needs CGO | A default package imported legacy Fyne | Move the import behind `-tags legacy_fyne` |
| Stale UI state | Viewmodel snapshot not refreshed after action | Update the model action and focused viewmodel test |
| Legacy Fyne tab panic | Widget update outside `fyne.Do()` | Wrap the legacy adapter mutation in `fyne.Do(func() {...})` |

---

## 📊 Performance

### Known Hot Paths

1. `crossbar.MVM()` - Matrix-vector multiply (N×M ops per call)
2. `preisach.Calculate()` - Hysteresis computation (per simulation step)
3. GUI refresh loops - Target ~50ms frame time (20 FPS)

### Benchmarking

```bash
# Crossbar MVM benchmark
go test -bench=BenchmarkMVM -benchmem ./module2-crossbar/pkg/crossbar

# Physics benchmark
go test -bench=. -benchmem ./shared/physics

# Profile a benchmark
go test -cpuprofile=cpu.prof -bench=BenchmarkMVM ./module2-crossbar/pkg/crossbar
go tool pprof cpu.prof
```

Memory optimization guide: [memory-optimization.md](memory-optimization.md)

---

## 🔒 Security Notes

This application has no network connectivity, no authentication, and stores no sensitive data. It is a fully offline tool.

File permissions for output directories:
```bash
mkdir -p screenshots recordings output logs
chmod 755 screenshots recordings output logs
```

---

## 📦 Dependencies

Key dependencies (see `go.mod`):

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/gogpu/ui` | current module pin | Default zero-CGO UI shell |
| `fyne.io/fyne/v2` | 2.7.2 | Tagged legacy Fyne parity shell |
| `golang.org/x/image` | latest | Image processing |

Dependency management:
```bash
# Update all dependencies
go get -u ./...
go mod tidy

# Legacy Fyne dependency changes should stay behind -tags legacy_fyne
```

---

## 🤝 Contributing

### Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Make changes with tests
4. Run the full check suite: `go fmt ./... && go vet ./... && go test -race ./...`
5. Commit with conventional format: `feat: add your feature`
6. Open a pull request

### What Needs Help

- Additional ferroelectric material parameters
- Physics model improvements
- UI accessibility enhancements
- Research paper indexing
- Documentation improvements

Full contribution guide: [../../CONTRIBUTING.md](../../CONTRIBUTING.md)

---

## 🔗 Quick Links

**Development:**
- [API Reference](api-reference.md) - Complete package APIs
- [Architecture](architecture/) - System design
- [Testing](testing/) - Testing guide
- [GUI Notes](gui/FYNE_NOTES.md) - tagged legacy Fyne tips

**Standards:**
- [Code Quality](code-quality.md) - Style guide
- [Accessibility](accessibility.md) - UI standards
- [Memory](memory-optimization.md) - Performance guide

**Operations:**
- [Runbook](../1-getting-started/runbook.md) - Build and ops
- [Repo Health](repo-health.md) - CI status
- [Automation](automation/) - Build scripts

---

**Last Updated:** 2026-02-16
**Go Version:** 1.25+
**Default UI:** `gogpu/ui`
