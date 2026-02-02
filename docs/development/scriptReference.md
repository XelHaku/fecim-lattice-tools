# FeCIM Visualizer Script Reference

Quick reference for file structure and key functions. Use this for fast lookups.

## Quick Navigation (For AI Agents)

| I need to... | Go to section |
|--------------|---------------|
| Find a function | [Quick Function Lookups](#quick-function-lookups) |
| Add a new demo | [Common Patterns: Adding New Demo](#adding-a-new-demo-module) |
| Fix a UI crash | [Error Resolution Guide](#error-resolution-guide) |
| Understand imports | [Import Patterns](#import-patterns) |
| Check thread safety | [Thread Safety Guide](#thread-safety-guide) |
| Find file dependencies | [Module Dependencies](#module-dependencies) |
| Modify physics | [Decision Trees](#decision-trees) |

## Decision Trees

### "I need to modify..." Decision Tree

| Modify What | File Location | Key Type/Function |
|-------------|---------------|-------------------|
| Crossbar physics | `module2-crossbar/pkg/crossbar/array.go` | `Config`, `Array`, `MVM()` |
| Hysteresis model | `module1-hysteresis/pkg/ferroelectric/preisach.go` | `PreisachModel`, `Update()` |
| MNIST inference | `module3-mnist/pkg/core/network.go` | `DualModeNetwork`, `Infer()` |
| Circuit peripherals | `module4-circuits/pkg/peripherals/*.go` | `DAC`, `ADC`, `TIA` |
| Theme/colors | `shared/theme/theme.go` | `ColorPrimary`, `ColorBackground` |
| Non-idealities | `module2-crossbar/pkg/crossbar/nonidealities.go` | `AnalyzeIRDrop()`, `AnalyzeSneakPaths()` |
| Quantization levels | `module2-crossbar/pkg/crossbar/array.go` | `FeCIMLevels`, `QuantizeTo30Levels()` |

### "I need to add..." Decision Tree

| Add What | Template File | Required Interface |
|----------|---------------|-------------------|
| New demo module | Copy `module4-circuits/pkg/gui/embedded.go` | `NewEmbedded*App()`, `BuildContent()`, `Start()`, `Stop()` |
| New tab to existing demo | See `module2-crossbar/pkg/gui/tabs/*.go` | Tab struct with `CreateContent()` method |
| New physics test | `*_test.go` in same package | `Test*` function with `*testing.T` |
| New peripheral | `module4-circuits/pkg/peripherals/` | Struct with `Default*()` constructor |

### "I need to debug..." Decision Tree

| Problem | First Check | Solution Pattern |
|---------|-------------|------------------|
| UI not updating | Missing `fyne.Do()` wrapper | Wrap in `fyne.Do(func() { ... })` |
| Nil pointer in GUI | Widget not initialized | Check `BuildContent()` called before `Start()` |
| Wrong quantization | Check `FeCIMLevels` constant | Use `crossbar.QuantizeTo30Levels()` |
| Import error | Check module path | Use `fecim-lattice-tools/module*` |
| Test fails on CI | GUI test without display | Skip with `t.Skip("Requires display")` |
| Goroutine panic | Race condition | Add mutex or use channels |

## Error Resolution Guide

### Common Errors and Fixes

| Error Message | Cause | Fix |
|---------------|-------|-----|
| `panic: runtime error: invalid memory address` | UI update from goroutine | Wrap in `fyne.Do(func() { ... })` |
| `undefined: crossbar.NewArray` | Wrong import path | Import `fecim-lattice-tools/module2-crossbar/pkg/crossbar` |
| `type *XxxApp has no field or method BuildContent` | Missing embedded interface | Implement: `BuildContent()`, `Start()`, `Stop()` |
| `cannot use x (type float64) as type int` | Quantization type mismatch | Use `int(crossbar.QuantizeTo30Levels(x) * 29)` for level index |
| `fyne: no OpenGL context` | GUI test without display | Add `t.Skip("Requires display")` |
| `weights not loaded` | Missing weight file | Call `LoadWeights()` before `Infer()` |
| `conductance out of range` | Value > 1.0 or < 0.0 | Normalize to [0, 1] before `ProgramWeight()` |
| `slice bounds out of range` | Array dimension mismatch | Check `array.Rows()` and `array.Cols()` match input |
| `context deadline exceeded` | Slow operation blocking | Move to goroutine with callback |
| `duplicate declaration` | Same name in package | Rename or check imports |

### Thread Safety Errors

| Symptom | Check | Fix |
|---------|-------|-----|
| Random crashes on UI update | Update from goroutine? | Use `fyne.Do()` |
| Race condition warnings | Shared state access? | Use `sync.Mutex` or channel |
| Frozen UI | Blocking on main thread? | Move to goroutine with `fyne.Do()` callback |
| Data corruption | Concurrent map access? | Use `sync.Map` or mutex |

### Build Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `go: module not found` | Module not in go.mod | Run `go mod tidy` |
| `cgo: pkg-config not found` | Missing system deps | Install `libgl1-mesa-dev` (Linux) |
| `vulkan headers not found` | Missing Vulkan SDK | Install `vulkan-sdk` or use non-Vulkan mode |
| `undefined: widget.NewXxx` | Old Fyne version | Run `go get fyne.io/fyne/v2@latest` |

## Thread Safety Guide

### Functions Requiring fyne.Do() Wrapper

Any UI update from a goroutine MUST use `fyne.Do()`:

| Component | Safe Pattern |
|-----------|--------------|
| Label text | `fyne.Do(func() { label.SetText("new") })` |
| Container add | `fyne.Do(func() { container.Add(widget) })` |
| Refresh widget | `fyne.Do(func() { widget.Refresh() })` |
| Progress bar | `fyne.Do(func() { progress.SetValue(0.5) })` |
| Status updates | `fyne.Do(func() { app.statusLabel.SetText("msg") })` |
| Heatmap redraw | `fyne.Do(func() { heatmap.Refresh() })` |

### Thread-Safe Code Pattern

```go
// WRONG - will crash randomly
go func() {
    label.SetText("Updated")  // NO: direct UI update from goroutine
}()

// CORRECT - thread-safe
go func() {
    result := heavyComputation()
    fyne.Do(func() {
        label.SetText(result)  // YES: wrapped in fyne.Do()
    })
}()
```

### Concurrent Data Access Pattern

```go
// Use mutex for shared state
type SafeState struct {
    mu    sync.Mutex
    value float64
}

func (s *SafeState) Update(v float64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.value = v
}
```

## Import Patterns

### Standard Module Imports

```go
import (
    // Core crossbar (MVM, quantization)
    "fecim-lattice-tools/module2-crossbar/pkg/crossbar"

    // Hysteresis model (Preisach)
    "fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"

    // MNIST network (DualModeNetwork)
    "fecim-lattice-tools/module3-mnist/pkg/core"

    // Circuit peripherals (DAC, ADC, TIA)
    "fecim-lattice-tools/module4-circuits/pkg/peripherals"

    // Comparison metrics
    "fecim-lattice-tools/module5-comparison/pkg/comparison"

    // EDA compiler
    "fecim-lattice-tools/module6-eda/pkg/compiler"

    // Shared theme
    "fecim-lattice-tools/shared/theme"

    // Fyne GUI
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/canvas"
)
```

## Module Dependencies

### Dependency Graph

```
cmd/fecim-lattice-tools
    ├── shared/theme
    ├── module1-hysteresis/pkg/gui
    ├── module2-crossbar/pkg/gui
    ├── module3-mnist/pkg/gui
    ├── module4-circuits/pkg/gui
    ├── module5-comparison/pkg/gui
    ├── module6-eda/pkg/gui
    └── module7-docs/pkg/gui

module3-mnist/pkg/core
    └── module2-crossbar/pkg/crossbar  (for quantization)

module4-circuits/pkg/peripherals
    └── (standalone - no internal deps)

module5-comparison/pkg/comparison
    └── (standalone - no internal deps)
```

### Safe to Modify (no dependents)
- `shared/theme/` - Only imported by GUI packages
- `module*-*/pkg/gui/` - Only imported by main app
- `docs/` - No code imports

### Modify with Care (has dependents)
- `module2-crossbar/pkg/crossbar/` - Imported by module3-mnist
- Constants like `FeCIMLevels` affect multiple modules

## Common Patterns

### Adding a New Demo Module

1. Create directory structure:
```
module7-newdemo/
    cmd/newdemo-gui/main.go
    pkg/newdemo/logic.go
    pkg/gui/
        app.go
        embedded.go  # Required for unified app
```

2. Implement embedded interface in `pkg/gui/embedded.go`:
```go
package gui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
)

type EmbeddedNewDemoApp struct {
    // internal state
    statusLabel *widget.Label
}

func NewEmbeddedNewDemoApp() *EmbeddedNewDemoApp {
    return &EmbeddedNewDemoApp{}
}

func (app *EmbeddedNewDemoApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
    app.statusLabel = widget.NewLabel("Ready")
    // Create and return UI content
    return widget.NewLabel("New Demo Content")
}

func (app *EmbeddedNewDemoApp) Start() {
    // Called when tab selected - start animations, load data
}

func (app *EmbeddedNewDemoApp) Stop() {
    // Called when tab deselected - stop animations, cleanup
}
```

3. Register in `cmd/fecim-lattice-tools/main.go`:
```go
import newdemo "fecim-lattice-tools/module7-newdemo/pkg/gui"

// In main(), add to demos slice:
newDemoApp := newdemo.NewEmbeddedNewDemoApp()
// Add tab: container.NewTabItem("7.NewDemo", newDemoApp.BuildContent(app, window))
```

### Adding a Physics Parameter

1. Define constant in appropriate package with citation:
```go
// module2-crossbar/pkg/crossbar/array.go
const (
    FeCIMLevels = 30  // Conference claim (COSM 2025), pending peer review
    NewParam    = 42  // [Citation required] - add DOI or source
)
```

2. Add physics test in `*_test.go`:
```go
func TestNewParamPhysics(t *testing.T) {
    // Test physical validity
    if NewParam < 0 {
        t.Error("NewParam must be non-negative")
    }
}
```

3. Document in `CLAUDE.md` Physics Constants table.

### Updating Quantization Logic

**Location:** `module2-crossbar/pkg/crossbar/array.go`

Current implementation:
```go
// QuantizeTo30Levels maps [0,1] to one of 30 discrete levels (demo baseline; conference claim)
func QuantizeTo30Levels(value float64) float64 {
    if value < 0 {
        value = 0
    } else if value > 1 {
        value = 1
    }
    level := int(value * float64(FeCIMLevels-1) + 0.5)
    return float64(level) / float64(FeCIMLevels-1)
}
```

To modify quantization:
1. Change `FeCIMLevels` constant (requires citation for new value)
2. Update tests in `module2-crossbar/pkg/crossbar/physics_test.go`
3. Run `go test ./module2-crossbar/...` to verify
4. Update CLAUDE.md Physics Constants table

### Running Inference with Custom Parameters

```go
import "fecim-lattice-tools/module3-mnist/pkg/core"

// Create network with custom config
config := core.DefaultNetworkConfig()
config.NumLevels = 30
config.NoiseLevel = 0.02
config.ADCBits = 6

network := core.NewDualModeNetwork(config)
network.LoadWeights("path/to/weights.gob")

// Run inference
result := network.Infer(inputImage)  // inputImage is []float64 of length 784
fmt.Printf("FP prediction: %d, CIM prediction: %d\n", result.FPPrediction, result.CIMPrediction)
```

### Creating a Custom Heatmap Widget

```go
import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

// Get conductance matrix from array
array := crossbar.NewArray(crossbar.Config{Rows: 8, Cols: 8})
matrix := array.GetConductanceMatrix()

// Create heatmap visualization
for i := 0; i < 8; i++ {
    for j := 0; j < 8; j++ {
        value := matrix[i][j]
        // Map value [0,1] to color
        rect := canvas.NewRectangle(theme.ColorForValue(value))
        // Add to container at position (i, j)
    }
}
```

## Directory Structure

```
fecim-lattice-tools/
├── cmd/
│   ├── fecim-lattice-tools/          # Unified GUI application entry point
│   │   ├── main.go                # Main entry, creates tabbed app with 5 demos
│   │   └── launcher.go            # Home tab with demo cards
│   └── launcher/                  # Legacy launcher
│
├── shared/                        # Shared utilities across all demos
│   ├── theme/theme.go             # FeCIM color theme (ColorPrimary, ColorBackground, etc.)
│   └── logging/logging.go         # Shared logger with file + stdout output
│
├── module1-hysteresis/            # Demo 1: P-E Hysteresis Curve (Memory Cell Physics)
│   ├── cmd/hysteresis/main.go     # Standalone entry point
│   ├── pkg/ferroelectric/
│   │   ├── preisach.go            # Preisach hysteresis model (core)
│   │   ├── preisach_advanced.go   # Advanced Preisach extensions
│   │   └── material.go            # HZO material definitions
│   ├── pkg/gui/
│   │   ├── embedded.go            # EmbeddedApp for unified GUI
│   │   ├── gui.go                 # Main GUI implementation
│   │   ├── labench.go             # Lab bench UI components
│   │   └── overlay.go             # Lab bench overlay rendering
│   ├── pkg/render/
│   │   ├── plot.go                # P-E curve plotting
│   │   ├── render.go              # Rendering engine
│   │   └── vulkan.go              # Vulkan GPU renderer
│   ├── pkg/simulation/
│   │   └── engine.go              # Simulation engine
│   ├── pkg/tui/
│   │   └── tui.go                 # Terminal UI mode
│   └── shaders/                   # SPIR-V compute/vertex/fragment shaders
│
├── module2-crossbar/              # Demo 2: Crossbar + Non-Idealities (4 tabs)
│   ├── cmd/
│   │   ├── crossbar-gui/main.go   # Standalone GUI entry point
│   │   └── inference/main.go      # Inference CLI
│   ├── pkg/crossbar/
│   │   ├── array.go               # Core crossbar array implementation
│   │   ├── nonidealities.go       # IR drop, sneak path, drift analysis
│   │   ├── drift.go               # Conductance drift over time
│   │   ├── irdrop.go              # IR drop voltage analysis
│   │   ├── sneakpath.go           # Sneak path current analysis
│   │   └── reference.go           # Reference implementations
│   ├── pkg/gui/
│   │   ├── app.go                 # TabbedCrossbarApp (with 4 sub-tabs)
│   │   ├── embedded.go            # EmbeddedTabbedCrossbarApp
│   │   ├── heatmap.go             # Conductance heatmap widget
│   │   ├── controls.go            # Control panel
│   │   ├── vectors.go             # MVM visualization
│   │   ├── liveslide.go           # Live slide components
│   │   └── tabs/
│   │       ├── ideal_tab.go       # Ideal MVM tab
│   │       ├── irdrop_tab.go      # IR drop analysis tab
│   │       ├── sneak_tab.go       # Sneak path analysis tab
│   │       └── drift_tab.go       # Drift analysis tab
│   ├── pkg/visualization/
│   │   ├── heatmap.go             # Heatmap generation
│   │   └── terminal.go            # Terminal visualization
│   ├── pkg/network/
│   │   └── network.go             # Network interface
│   ├── pkg/weights/
│   │   ├── weights.go             # Weight management
│   │   └── serialization.go       # Weight I/O
│   └── shaders/                   # MVM compute shaders
│
├── module3-mnist/                 # Demo 3: AI Brain (Dual Mode: FP vs CIM)
│   ├── cmd/
│   │   ├── mnist/main.go          # Training/inference CLI
│   │   └── mnist-gui/main.go      # GUI entry point
│   ├── pkg/core/
│   │   ├── network.go             # DualModeNetwork (FP vs CIM inference)
│   │   └── quantize.go            # Quantization to 30 levels
│   ├── pkg/gui/
│   │   ├── app.go                 # MNISTApp (basic)
│   │   ├── dualmode.go            # DualModeApp (FP vs CIM side-by-side)
│   │   ├── embedded.go            # EmbeddedMNISTApp + EmbeddedDualModeApp
│   │   ├── canvas.go              # Drawing canvas widget
│   │   ├── activations.go         # Layer activation visualization
│   │   ├── metrics.go             # Accuracy metrics display
│   │   ├── dialogs.go             # Dialog components
│   │   ├── liveslide.go           # Live slide components
│   │   └── tour.go                # Dr. Tour reference info
│   ├── pkg/mnist/
│   │   └── loader.go              # MNIST dataset loading
│   ├── pkg/training/
│   │   └── network.go             # Alternative training network
│   ├── data/                      # MNIST dataset (gzipped)
│   ├── scripts/
│   │   ├── train_all_sizes.sh
│   │   └── benchmark.sh
│   ├── train_and_save.go
│   ├── train_full_precision.go
│   └── train_mnist_proper.go
│
├── module4-circuits/              # Demo 4: Chip System (Peripheral Circuits)
│   ├── cmd/circuits-gui/main.go   # GUI entry point
│   ├── pkg/peripherals/
│   │   ├── dac.go                 # DAC converter model
│   │   ├── adc.go                 # ADC converter model
│   │   ├── tia.go                 # Transimpedance amplifier
│   │   ├── chargepump.go          # Programming voltage pump
│   │   └── analysis.go            # Circuit analysis (timing, power)
│   └── pkg/gui/
│       ├── app.go                 # CircuitsApp
│       ├── embedded.go            # EmbeddedCircuitsApp
│       ├── signalflow.go          # Signal flow diagram
│       └── liveslide.go           # Live slide components
│
├── module5-comparison/            # Demo 5: Business Case (Technical Briefing)
│   ├── cmd/comparison-gui/main.go # GUI entry point
│   ├── pkg/comparison/
│   │   ├── architecture.go        # Architecture specs + comparison
│   │   └── render.go              # Comparison charts rendering
│   └── pkg/gui/
│       ├── app.go                 # ComparisonApp
│       ├── embedded.go            # EmbeddedComparisonApp
│       ├── liveslide.go           # Live slide components
│       └── widgets.go             # Custom comparison widgets
│
├── module6-eda/                   # EDA Tools & Crossbar Compiler
│   ├── pkg/compiler/
│   │   ├── compiler.go            # Main crossbar compiler
│   │   └── types.go               # CompileConfig, CrossbarMapping
│   ├── pkg/export/
│   │   ├── csv.go                 # CSV export
│   │   ├── json.go                # JSON export
│   │   └── spice.go               # SPICE netlist generation
│   └── pkg/gui/
│       ├── app.go                 # EDA GUI app
│       └── tabs/
│           ├── compiler_tab.go    # Compiler interface
│           ├── export_tab.go      # Export format selection
│           ├── layout_tab.go      # Physical layout visualization
│           └── state.go           # State management
│
├── module7-docs/                  # Documentation Viewer
│   └── pkg/gui/
│       ├── embedded.go            # EmbeddedDocsApp
│       ├── navigation.go          # BreadcrumbWidget, TableOfContentsWidget, QuickAccessPanel
│       ├── search.go              # SearchIndex, SearchDialog
│       ├── layout.go              # Responsive LayoutManager
│       ├── persistence.go         # DocsHistory (recent/favorites)
│       └── glossary_integration.go # GlossaryPillsWidget, RelatedDocsWidget
│
├── docs/                          # Documentation
│   ├── eda/                       # EDA documentation
│   ├── opensource/                # Open science resources
│   ├── papers/                    # Research papers (organized by topic)
│   ├── project/                   # Project planning
│   └── archive/removed-demos/     # Archived demos (thermal, 3D stack, etc.)
│
├── scripts/
│   └── build-all.sh               # Build all demos script
│
└── logs/                          # Runtime logs (datetime-stamped)
```

## Key Constants

```go
// module2-crossbar/pkg/crossbar/array.go
const FeCIMLevels = 30  // Conference claim (COSM 2025), pending peer review

// shared/theme/theme.go
ColorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan
ColorSecondary  = color.RGBA{255, 107, 107, 255} // Coral
ColorBackground = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
ColorAccent     = color.RGBA{78, 205, 196, 255}  // Teal
ColorWarning    = color.RGBA{255, 230, 109, 255} // Yellow
ColorPositive   = color.RGBA{255, 100, 100, 255} // Positive polarization
ColorNegative   = color.RGBA{100, 150, 255, 255} // Negative polarization

// module6-eda/pkg/compiler/types.go
DefaultConfig = CompileConfig{
    ArrayRows: 128,
    ArrayCols: 128,
    Levels:    30,
    GMin:      10.0 µS,
    GMax:      100.0 µS,
}
```

## Core Types & Functions

### Unified Application (cmd/fecim-lattice-tools/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| main.go | `DemoApp` | Holds all 5 embedded demo instances |
| main.go | `main()` | Creates tabbed Fyne app, manages demo start/stop |
| main.go | `feCIMTheme` | Implements fyne.Theme for FeCIM branding |
| launcher.go | `DemoInfo` | Demo metadata (name, description, icon) |
| launcher.go | `DemoCard` | Clickable demo card widget |
| launcher.go | `GetDemos()` | Returns DemoInfo slice for all demos |
| launcher.go | `CreateLauncherContent()` | Creates home tab with clickable demo cards |

### Demo 1: Hysteresis (module1-hysteresis/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/ferroelectric/preisach.go | `PreisachModel` | Preisach hysteresis model with memory |
| pkg/ferroelectric/preisach.go | `NewPreisachModel()` | Constructor with material parameters |
| pkg/ferroelectric/preisach.go | `Update(E float64)` | Apply electric field, return polarization |
| pkg/ferroelectric/preisach.go | `GetHysteresisLoop()` | Generate full P-E curve data |
| pkg/ferroelectric/preisach.go | `DiscreteStates(N int)` | Get N discrete polarization states |
| pkg/ferroelectric/preisach_advanced.go | `AdvancedPreisach` | Extended model features |
| pkg/ferroelectric/material.go | `HZOMaterial` | HZO material properties |
| pkg/ferroelectric/material.go | `DefaultHZO()` | Default HZO material |
| pkg/ferroelectric/material.go | `FeCIMMaterial()` | Optimized FeCIM material |
| pkg/ferroelectric/material.go | `CoerciveVoltage()` | Coercive voltage calculation |
| pkg/ferroelectric/material.go | `SwitchingEnergy()` | Energy per switch |
| pkg/ferroelectric/material.go | `DiscreteLevel()` | Map to discrete level |
| pkg/gui/embedded.go | `EmbeddedApp` | Embedded version for unified GUI |
| pkg/gui/embedded.go | `BuildContent()` | Build UI for tab |
| pkg/gui/embedded.go | `Start()` / `Stop()` | Lifecycle methods |
| pkg/gui/overlay.go | `RenderText()` | Render lab bench status overlay |
| pkg/render/vulkan.go | `VulkanRenderer` | GPU-accelerated rendering |
| pkg/simulation/engine.go | `Engine` | Simulation engine |

### Demo 2: Crossbar (module2-crossbar/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/crossbar/array.go | `Config` | Array configuration (rows, cols, noise, ADC/DAC bits) |
| pkg/crossbar/array.go | `Cell` | Individual crossbar cell |
| pkg/crossbar/array.go | `Array` | Crossbar array with cells matrix |
| pkg/crossbar/array.go | `NewArray(cfg)` | Create new crossbar array |
| pkg/crossbar/array.go | `ProgramWeight()` | Program weight to cell (quantizes to 30-level demo baseline) |
| pkg/crossbar/array.go | `ProgramWeightMatrix()` | Program entire weight matrix |
| pkg/crossbar/array.go | `QuantizeTo30Levels()` | Quantize value to FeCIM 30-level demo baseline |
| pkg/crossbar/array.go | `GetLevel()` | Get quantization level (0-29) |
| pkg/crossbar/array.go | `MVM(input)` | Matrix-vector multiply: y = W × x |
| pkg/crossbar/array.go | `VMM(input)` | Vector-matrix multiply: y = x × W |
| pkg/crossbar/array.go | `quantizeDAC()` | DAC quantization |
| pkg/crossbar/array.go | `quantizeADC()` | ADC quantization |
| pkg/crossbar/array.go | `GetStats()` | Array statistics |
| pkg/crossbar/array.go | `GetConductanceMatrix()` | Get raw conductances |
| pkg/crossbar/nonidealities.go | `WireParams` | Wire resistance parameters |
| pkg/crossbar/nonidealities.go | `IRDropAnalysis` | IR drop analysis results |
| pkg/crossbar/nonidealities.go | `SneakPathAnalysis` | Sneak path analysis results |
| pkg/crossbar/nonidealities.go | `DefaultWireParams()` | Default wire parameters |
| pkg/crossbar/nonidealities.go | `AnalyzeIRDrop()` | Compute IR drop across array |
| pkg/crossbar/nonidealities.go | `AnalyzeSneakPaths()` | Compute sneak path currents |
| pkg/crossbar/nonidealities.go | `MVMWithIRDrop()` | MVM with IR drop effects |
| pkg/crossbar/nonidealities.go | `ComputeError()` | Compute error vs ideal |
| pkg/crossbar/drift.go | `DriftSimulator` | Conductance drift simulator |
| pkg/crossbar/drift.go | `DriftSnapshot` | Drift state at time t |
| pkg/crossbar/drift.go | `NewDriftSimulator()` | Create drift simulator |
| pkg/crossbar/irdrop.go | `IRDropSimulator` | IR drop simulator |
| pkg/crossbar/irdrop.go | `IRDropMitigation` | Mitigation strategies |
| pkg/crossbar/sneakpath.go | `SneakPathAnalyzer` | Sneak path analyzer |
| pkg/crossbar/sneakpath.go | `SneakMitigation` | Mitigation strategies |
| pkg/crossbar/reference.go | `CPUReference` | Reference CPU implementation |
| pkg/crossbar/reference.go | `MVMWithNoise()` | MVM with noise injection |
| pkg/crossbar/reference.go | `MVMWithQuantization()` | MVM with quantization |
| pkg/gui/app.go | `TabbedCrossbarApp` | Main app with 4 sub-tabs |
| pkg/gui/app.go | `NewTabbedCrossbarApp()` | Constructor |
| pkg/gui/app.go | `runMVM()` | Execute animated MVM operation |
| pkg/gui/app.go | `analyzeIRDrop()` | Run IR drop analysis |
| pkg/gui/app.go | `analyzeSneakPaths()` | Run sneak path analysis |
| pkg/gui/embedded.go | `EmbeddedTabbedCrossbarApp` | Embeddable version for unified GUI |
| pkg/gui/heatmap.go | `CrossbarHeatmap` | Interactive conductance heatmap widget |
| pkg/gui/tabs/ideal_tab.go | `IdealTab` | Ideal MVM tab |
| pkg/gui/tabs/irdrop_tab.go | `IRDropTab` | IR drop analysis tab |
| pkg/gui/tabs/sneak_tab.go | `SneakPathTab` | Sneak path analysis tab |
| pkg/gui/tabs/drift_tab.go | `DriftTab` | Drift analysis tab |

### Demo 3: MNIST (module3-mnist/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/core/network.go | `NetworkConfig` | Network configuration |
| pkg/core/network.go | `DualModeNetwork` | FP vs CIM dual-mode network |
| pkg/core/network.go | `InferenceResult` | Inference output with metrics |
| pkg/core/network.go | `DefaultNetworkConfig()` | Default config |
| pkg/core/network.go | `NewDualModeNetwork()` | Constructor |
| pkg/core/network.go | `LoadWeights()` | Load pre-trained weights |
| pkg/core/network.go | `SetNumLevels()` | Set quantization levels |
| pkg/core/network.go | `SetNoiseLevel()` | Set noise injection level |
| pkg/core/network.go | `SetADCBits()` / `SetDACBits()` | Set ADC/DAC resolution |
| pkg/core/network.go | `Infer()` | Run inference (FP + CIM) |
| pkg/core/network.go | `InferFPOnly()` | FP inference only |
| pkg/core/network.go | `InferCIMOnly()` | CIM inference only |
| pkg/core/network.go | `forwardFP()` | Full-precision forward pass |
| pkg/core/network.go | `forwardCIM()` | Quantized CIM forward pass |
| pkg/core/network.go | `RequantizeWeights()` | Re-quantize weights |
| pkg/core/network.go | `GetQuantizationStats()` | Get quantization statistics |
| pkg/core/quantize.go | `QuantizationStats` | Quantization statistics |
| pkg/core/quantize.go | `RandomSource` | Random number generator |
| pkg/core/quantize.go | `QuantizeWeights()` | Quantize weight matrix |
| pkg/core/quantize.go | `QuantizeBias()` | Quantize bias vector |
| pkg/core/quantize.go | `ComputeQuantizationStats()` | Compute stats |
| pkg/core/quantize.go | `AddGaussianNoise()` | Add noise to values |
| pkg/mnist/loader.go | `LoadMNIST()` | Load MNIST dataset from gzipped files |
| pkg/gui/app.go | `MNISTApp` | Basic MNIST app |
| pkg/gui/dualmode.go | `DualModeApp` | FP vs CIM dual-mode app |
| pkg/gui/embedded.go | `EmbeddedMNISTApp` | Basic embedded version |
| pkg/gui/embedded.go | `EmbeddedDualModeApp` | Dual-mode embedded version |
| pkg/gui/canvas.go | `DrawingCanvas` | Custom drawing widget for digit input |
| pkg/gui/activations.go | Various | Layer activation visualization |
| pkg/gui/metrics.go | Various | Accuracy and performance metrics |
| pkg/training/network.go | `Network` | Alternative training network |

### Demo 4: Circuits (module4-circuits/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/peripherals/dac.go | `DAC` | Digital-to-analog converter model |
| pkg/peripherals/dac.go | `DefaultDAC()` | Default DAC (30-level demo baseline) |
| pkg/peripherals/dac.go | `Levels()` | Number of output levels |
| pkg/peripherals/dac.go | `Convert()` | Digital to analog conversion |
| pkg/peripherals/dac.go | `ConvertWithNonlinearity()` | With INL/DNL |
| pkg/peripherals/dac.go | `EnergyPerConversion()` | Energy per conversion |
| pkg/peripherals/adc.go | `ADC` | Analog-to-digital converter model |
| pkg/peripherals/adc.go | `ADCType` | ADC architecture type |
| pkg/peripherals/adc.go | `INLDNLAnalysis` | INL/DNL analysis results |
| pkg/peripherals/adc.go | `DefaultADC()` | Default ADC |
| pkg/peripherals/adc.go | `Convert()` | Analog to digital conversion |
| pkg/peripherals/adc.go | `ENOB()` | Effective number of bits |
| pkg/peripherals/adc.go | `AnalyzeINLDNL()` | INL/DNL analysis |
| pkg/peripherals/tia.go | `TIA` | Transimpedance amplifier model |
| pkg/peripherals/tia.go | `DefaultTIA()` | Default TIA |
| pkg/peripherals/tia.go | `Convert()` | Current to voltage conversion |
| pkg/peripherals/tia.go | `SNR()` | Signal-to-noise ratio |
| pkg/peripherals/tia.go | `DynamicRange()` | Dynamic range |
| pkg/peripherals/tia.go | `SettlingTime()` | Settling time |
| pkg/peripherals/chargepump.go | `ChargePump` | Programming voltage generator |
| pkg/peripherals/chargepump.go | `DefaultChargePump()` | Default charge pump |
| pkg/peripherals/chargepump.go | `NegativePump()` | Negative voltage pump |
| pkg/peripherals/chargepump.go | `ActualOutputVoltage()` | Output with losses |
| pkg/peripherals/chargepump.go | `EnergyPerOperation()` | Energy per operation |
| pkg/peripherals/chargepump.go | `SupportsLevel()` | Check if level supported |
| pkg/peripherals/analysis.go | `TimingAnalysis` | Timing analysis results |
| pkg/peripherals/analysis.go | `PowerBreakdown` | Power breakdown |
| pkg/peripherals/analysis.go | `TransferFunction` | Transfer function data |
| pkg/peripherals/analysis.go | `AnalyzeTiming()` | Full timing analysis |
| pkg/peripherals/analysis.go | `AnalyzePower()` | Power consumption analysis |
| pkg/peripherals/analysis.go | `ComputeTransferFunction()` | Transfer function |
| pkg/gui/app.go | `CircuitsApp` | Main circuits app |
| pkg/gui/embedded.go | `EmbeddedCircuitsApp` | Embedded version |
| pkg/gui/signalflow.go | `SignalFlowDiagram` | Signal flow visualization |

### Demo 5: Comparison (module5-comparison/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/comparison/architecture.go | `Architecture` | Memory technology specs |
| pkg/comparison/architecture.go | `InferenceResult` | Inference benchmark result |
| pkg/comparison/architecture.go | `Workload` | Benchmark workload definition |
| pkg/comparison/architecture.go | `DataCenterMetrics` | Data center scale metrics |
| pkg/comparison/architecture.go | `ComparisonResult` | Full comparison result |
| pkg/comparison/architecture.go | `FeCIMAdvantage` | FeCIM advantages summary |
| pkg/comparison/architecture.go | `TraditionalCPU()` | CPU architecture |
| pkg/comparison/architecture.go | `GPUAccelerator()` | GPU architecture |
| pkg/comparison/architecture.go | `FeCIMChip()` | FeCIM architecture |
| pkg/comparison/architecture.go | `CustomArchitecture()` | Custom architecture builder |
| pkg/comparison/architecture.go | `CalculateEfficiency()` | Compute efficiency metrics |
| pkg/comparison/architecture.go | `RunInference()` | Run inference benchmark |
| pkg/comparison/architecture.go | `MNISTWorkload()` | MNIST benchmark |
| pkg/comparison/architecture.go | `ResNet50Workload()` | ResNet-50 benchmark |
| pkg/comparison/architecture.go | `BERTBaseWorkload()` | BERT-Base benchmark |
| pkg/comparison/architecture.go | `GPT2Workload()` | GPT-2 benchmark |
| pkg/comparison/architecture.go | `LLMWorkload()` | Large LLM benchmark |
| pkg/comparison/architecture.go | `ScaleToDataCenter()` | Scale to data center |
| pkg/comparison/architecture.go | `CompareArchitectures()` | Compare all architectures |
| pkg/comparison/architecture.go | `CalculateAdvantages()` | Calculate FeCIM advantages |
| pkg/comparison/render.go | `Renderer` | Comparison chart renderer |
| pkg/comparison/render.go | `RenderArchitectureSpecs()` | Render specs table |
| pkg/comparison/render.go | `RenderBarChart()` | Render bar chart |
| pkg/comparison/render.go | `RenderAdvantages()` | Render advantages |
| pkg/gui/app.go | `ComparisonApp` | Main comparison app |
| pkg/gui/embedded.go | `EmbeddedComparisonApp` | Embedded version |
| pkg/gui/widgets.go | Various | Custom comparison widgets |

### Module 6: EDA Tools (module6-eda/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/compiler/types.go | `CompileConfig` | Compiler configuration |
| pkg/compiler/types.go | `CellAssignment` | Cell mapping assignment |
| pkg/compiler/types.go | `CrossbarMapping` | Full crossbar mapping |
| pkg/compiler/types.go | `Stats` | Compilation statistics |
| pkg/compiler/types.go | `DefaultConfig()` | Default compiler config |
| pkg/compiler/compiler.go | `Compile()` | Compile network to crossbar |
| pkg/export/csv.go | `ExportCSV()` | Export mapping to CSV |
| pkg/export/json.go | `ExportJSON()` | Export mapping to JSON |
| pkg/export/spice.go | `ExportSPICE()` | Export as SPICE netlist |
| pkg/export/spice.go | `GenerateSPICE()` | Generate SPICE code |
| pkg/gui/app.go | `CreateMainWindow()` | Create EDA GUI window |
| pkg/gui/tabs/compiler_tab.go | Compiler tab | Compiler interface |
| pkg/gui/tabs/export_tab.go | Export tab | Export format selection |
| pkg/gui/tabs/layout_tab.go | Layout tab | Physical layout visualization |
| pkg/gui/tabs/state.go | State management | Shared GUI state |

### Module 7: Documentation (module7-docs/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/gui/embedded.go | `EmbeddedDocsApp` | Main documentation viewer app |
| pkg/gui/embedded.go | `NewEmbeddedDocsApp()` | Constructor |
| pkg/gui/embedded.go | `BuildContent()` | Build UI with tree, search, content |
| pkg/gui/embedded.go | `Start()` / `Stop()` | Lifecycle methods |
| pkg/gui/embedded.go | `loadDocument()` | Load and render markdown file |
| pkg/gui/search.go | `SearchIndex` | Full-text search index with TF-IDF |
| pkg/gui/search.go | `NewSearchIndex()` | Constructor, builds index |
| pkg/gui/search.go | `Query()` | Fuzzy search with TF-IDF ranking |
| pkg/gui/search.go | `SearchDialog` | Modal search dialog with keyboard nav |
| pkg/gui/navigation.go | `BreadcrumbWidget` | Hierarchical path navigation |
| pkg/gui/navigation.go | `TableOfContentsWidget` | Auto-generated ToC from headings |
| pkg/gui/navigation.go | `QuickAccessPanel` | Recent & favorite documents |
| pkg/gui/layout.go | `LayoutManager` | Responsive layout (Mobile/Tablet/Desktop/Wide) |
| pkg/gui/persistence.go | `DocsHistory` | Recent/favorites persistence to JSON |
| pkg/gui/persistence.go | `AddRecent()` | Add to LRU recent list |
| pkg/gui/persistence.go | `ToggleFavorite()` | Add/remove from favorites |
| pkg/gui/glossary_integration.go | `GlossaryPillsWidget` | Display detected glossary terms |
| pkg/gui/glossary_integration.go | `DetectGlossaryTerms()` | Scan content for glossary terms |
| pkg/gui/glossary_integration.go | `DocumentMetadataWidget` | Category, reading time, terms |
| pkg/gui/glossary_integration.go | `RelatedDocsWidget` | Related document suggestions |

## Shared Utilities

### Theme (shared/theme/theme.go)

```go
type FeCIMTheme struct{}

func (t *FeCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color
func (t *FeCIMTheme) Font(style fyne.TextStyle) fyne.Resource
func (t *FeCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource
func (t *FeCIMTheme) Size(name fyne.ThemeSizeName) float32

// Color constants
ColorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan
ColorSecondary  = color.RGBA{255, 107, 107, 255} // Coral red
ColorBackground = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
ColorAccent     = color.RGBA{78, 205, 196, 255}  // Teal
ColorWarning    = color.RGBA{255, 230, 109, 255} // Yellow
```

### Logging (shared/logging/logging.go)

```go
type Logger struct { ... }

func NewLogger(name string) *Logger  // Creates timestamped log file in logs/
func (l *Logger) Printf(format string, v ...any)
func (l *Logger) Close()
func getLogsDir() string  // Returns logs directory path
```

## Build & Run

```bash
# Build unified visualizer
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Run unified app
./fecim-lattice-tools

# Or use launch script
./launch.sh

# Build individual modules
go build -o module1-hysteresis/hysteresis ./module1-hysteresis/cmd/hysteresis
go build -o module2-crossbar/crossbar-gui ./module2-crossbar/cmd/crossbar-gui
go build -o module3-mnist/mnist-gui ./module3-mnist/cmd/mnist-gui
go build -o module4-circuits/circuits-gui ./module4-circuits/cmd/circuits-gui
go build -o module5-comparison/comparison-gui ./module5-comparison/cmd/comparison-gui

# Run tests
go test ./...
go test ./module2-crossbar/pkg/crossbar
go test -v -run TestPreisach
```

## Embedded App Pattern

Each demo follows this pattern for embedding in the unified GUI:

```go
// pkg/gui/embedded.go
type EmbeddedXxxApp struct {
    // internal state
}

func NewEmbeddedXxxApp() *EmbeddedXxxApp { ... }
func (app *EmbeddedXxxApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject { ... }
func (app *EmbeddedXxxApp) Start() { ... }  // Called when tab selected
func (app *EmbeddedXxxApp) Stop() { ... }   // Called when tab deselected
```

**Implementations:**
- `EmbeddedApp` - module1 (hysteresis)
- `EmbeddedTabbedCrossbarApp` - module2 (with 4 sub-tabs: Ideal, IR Drop, Sneak, Drift)
- `EmbeddedMNISTApp` - module3 (basic)
- `EmbeddedDualModeApp` - module3 (FP vs CIM side-by-side)
- `EmbeddedCircuitsApp` - module4
- `EmbeddedComparisonApp` - module5
- `EmbeddedEDAApp` - module6
- `EmbeddedDocsApp` - module7 (documentation viewer)

## Quick Function Lookups

| Need | File | Function |
|------|------|----------|
| Quantize to 30 levels | module2-crossbar/pkg/crossbar/array.go | `QuantizeTo30Levels()` |
| Create crossbar array | module2-crossbar/pkg/crossbar/array.go | `NewArray()` |
| Run MVM | module2-crossbar/pkg/crossbar/array.go | `Array.MVM()` |
| Run VMM | module2-crossbar/pkg/crossbar/array.go | `Array.VMM()` |
| Create Preisach model | module1-hysteresis/pkg/ferroelectric/preisach.go | `NewPreisachModel()` |
| Get P-E loop data | module1-hysteresis/pkg/ferroelectric/preisach.go | `GetHysteresisLoop()` |
| Get material properties | module1-hysteresis/pkg/ferroelectric/material.go | `DefaultHZO()`, `FeCIMMaterial()` |
| IR drop analysis | module2-crossbar/pkg/crossbar/nonidealities.go | `AnalyzeIRDrop()` |
| Sneak path analysis | module2-crossbar/pkg/crossbar/nonidealities.go | `AnalyzeSneakPaths()` |
| Drift simulation | module2-crossbar/pkg/crossbar/drift.go | `NewDriftSimulator()` |
| Create dual-mode network | module3-mnist/pkg/core/network.go | `NewDualModeNetwork()` |
| Run FP/CIM inference | module3-mnist/pkg/core/network.go | `Infer()`, `InferFPOnly()`, `InferCIMOnly()` |
| DAC model | module4-circuits/pkg/peripherals/dac.go | `DefaultDAC()` |
| ADC model | module4-circuits/pkg/peripherals/adc.go | `DefaultADC()` |
| TIA model | module4-circuits/pkg/peripherals/tia.go | `DefaultTIA()` |
| Compare architectures | module5-comparison/pkg/comparison/architecture.go | `CompareArchitectures()` |
| Compile to crossbar | module6-eda/pkg/compiler/compiler.go | `Compile()` |
| Export SPICE netlist | module6-eda/pkg/export/spice.go | `ExportSPICE()` |
| FeCIM theme colors | shared/theme/theme.go | `ColorPrimary`, `ColorBackground` |
| Create logger | shared/logging/logging.go | `NewLogger()` |
| Search documentation | module7-docs/pkg/gui/search.go | `SearchIndex.Query()` |
| Show search dialog | module7-docs/pkg/gui/search.go | `SearchDialog.Show()` |
| Detect glossary terms | module7-docs/pkg/gui/glossary_integration.go | `DetectGlossaryTerms()` |
| Find related docs | module7-docs/pkg/gui/glossary_integration.go | `FindRelated()` |

## Archived Demos

The following demos were archived (see `docs/archive/removed-demos/`):
- **demo5-thermal** - Merged into comparison module
- **demo6-multilayer** (3D Stack) - Archived (too futuristic for TRL 4)
- **demo7-nonidealities** - Merged into module2-crossbar as tabs

## Dependencies

- **fyne.io/fyne/v2 v2.7.2** - Cross-platform GUI framework
- **github.com/charmbracelet/bubbletea v1.2.4** - Terminal UI (module1 TUI mode)
- **github.com/go-gl/glfw/v3.3/glfw** - Window management
- **github.com/vulkan-go/vulkan** - GPU rendering (module1 Vulkan mode)
- Go toolchain: **go1.24.12**
