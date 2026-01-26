# FeCIM Lattice Tools - Architecture Guide

A comprehensive reference to the design, structure, and patterns of the FeCIM Lattice Tools project.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Module Dependency Graph](#module-dependency-graph)
3. [Core Abstractions](#core-abstractions)
4. [Physics Layer](#physics-layer)
5. [GUI Layer](#gui-layer)
6. [Data Flow](#data-flow)
7. [Threading Model](#threading-model)

---

## High-Level Architecture

FeCIM Lattice Tools is a **unified educational visualization suite** for Ferroelectric Compute-in-Memory (FeCIM) based on Dr. external research group's superlattice research. The entire application is orchestrated through a single entry point that launches six independent modules as tabs, each demonstrating a different aspect of FeCIM technology.

### Entry Point: cmd/fecim-visualizer

The main application (`cmd/fecim-visualizer/main.go`) implements a **unified launcher** pattern:

```
┌─────────────────────────────────────────────┐
│  FeCIM Lattice Tools (main.go)              │
│  ────────────────────────────────────────   │
│  Fyne App                                   │
│  ├─ Window (1400x900 default, resizable)   │
│  └─ ContentStack (7 views)                  │
│     ├─ View 0: Home/Launcher               │
│     ├─ View 1: Module 1 (Hysteresis)       │
│     ├─ View 2: Module 2 (Crossbar)         │
│     ├─ View 3: Module 3 (MNIST)            │
│     ├─ View 4: Module 4 (Circuits)         │
│     ├─ View 5: Module 5 (Comparison)       │
│     └─ View 6: Module 6 (EDA)              │
│                                             │
│  Features:                                  │
│  ├─ Persistent window state                │
│  ├─ View navigation (Home button, cards)   │
│  ├─ Screenshot capture                     │
│  ├─ FFmpeg video recording                 │
│  └─ Module lifecycle (Start/Stop)          │
└─────────────────────────────────────────────┘
```

The launcher dynamically loads demo cards and manages lifecycle—modules only run when their tab is active.

### The 6-Module Story

Each module demonstrates a layer in the FeCIM stack:

| Module | Purpose | Physics | GUI Framework | Status |
|--------|---------|---------|---------------|--------|
| **1. Hysteresis** | Memory cell physics | Preisach model, 30-level quantization | Real-time P-E curve plotting | Stable |
| **2. Crossbar** | Matrix-vector multiply | Ohm's law, IR drop, sneak paths, drift | Heatmap with MVM operations | Stable |
| **3. MNIST** | Neural network application | Network inference with quantized weights | Digit drawing + recognition | Stable |
| **4. Circuits** | Peripheral electronics | DAC/ADC/TIA circuit modeling | Schematic diagrams, waveforms | Stable |
| **5. Comparison** | Technology benchmarks | Energy, speed, density metrics | Interactive comparison charts | Stable |
| **6. EDA** | Chip design tools | Placement, routing (experimental) | Layout visualization | WIP |

---

## Module Dependency Graph

```
┌──────────────────────────────────────────────────────────────┐
│                        shared/                               │
│    ┌─────────────────┬──────────────────┬───────────────┐   │
│    │   theme         │    widgets       │   logging     │   │
│    │   (colors)      │   (adaptive UI)  │   (verbosity) │   │
│    └────────┬────────┴────────┬─────────┴───────┬───────┘   │
└─────────────┼────────────────┼──────────────────┼────────────┘
              │                │                  │
    ┌─────────▼────────┐       │         ┌────────▼─────────────────┐
    │  module1-hysteresis   │       │         │  module2-crossbar     │
    │  ├─ ferroelectric/    │       │         │  ├─ crossbar/        │
    │  │  ├─ material.go    │       │         │  │  ├─ array.go      │
    │  │  ├─ preisach.go    │       │         │  │  ├─ irdrop.go     │
    │  │  └─ render.go      │       │         │  │  ├─ sneak*.go     │
    │  ├─ simulation/       │       │         │  │  ├─ drift.go       │
    │  └─ gui/             │       │         │  │  └─ enhanced.go    │
    │     └─ embedded.go    │       │         │  └─ gui/            │
    └─────────────────────┘       │         │     └─ embedded.go    │
                                  │         └──────────────────────┘
                                  │
                    ┌─────────────▼──────────────┐
                    │  module3-mnist             │
                    │  ├─ network/              │
                    │  │  ├─ weights.go         │
                    │  │  └─ inference.go       │
                    │  └─ gui/                  │
                    │     └─ embedded.go        │
                    └──────────────────────────┘
                            │
           ┌────────────────┴────────────────┐
           │                                 │
    ┌──────▼──────────────┐          ┌──────▼──────────────┐
    │ module4-circuits    │          │ module5-comparison  │
    │ ├─ dac/            │          │ ├─ data/           │
    │ ├─ adc/            │          │ ├─ metrics.go      │
    │ ├─ tia/            │          │ └─ gui/            │
    │ └─ gui/            │          │    └─ embedded.go  │
    │    └─ embedded.go  │          └─────────────────────┘
    └───────────────────┘
            │
            │
    ┌───────▼──────────────┐
    │ module6-eda          │
    │ ├─ placement/        │
    │ ├─ netlist/          │
    │ └─ gui/              │
    │    └─ embedded.go    │
    └──────────────────────┘
```

**Key principle**: Each module is **independent**. Modules can use shared infrastructure (theme, logging, adaptive layouts) but don't depend on each other. This enables:

- Standalone execution or embedded tabs
- Isolated development and testing
- Easy refactoring without cascade effects
- Clean separation of concerns

---

## Core Abstractions

### 1. EmbeddedApp Interface

All modules implement the same interface for seamless integration:

```go
// Pseudo-interface (Go doesn't formalize this, but all modules follow it)
type EmbeddedApp interface {
    // BuildContent creates the UI once (called at app startup)
    BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject

    // Start begins the module's main loop (called when tab becomes active)
    Start()

    // Stop ends the module's main loop (called when navigating away)
    Stop()
}
```

Each module has a concrete implementation:

| Module | Type | Location |
|--------|------|----------|
| Hysteresis | `EmbeddedApp` | `module1-hysteresis/pkg/gui/embedded.go` |
| Crossbar | `EmbeddedCrossbarApp` | `module2-crossbar/pkg/gui/embedded.go` |
| MNIST | `EmbeddedDualModeApp` | `module3-mnist/pkg/gui/embedded.go` |
| Circuits | `EmbeddedCircuitsApp` | `module4-circuits/pkg/gui/embedded.go` |
| Comparison | `EmbeddedComparisonApp` | `module5-comparison/pkg/gui/embedded.go` |
| EDA | `EmbeddedEDAApp` | `module6-eda/pkg/gui/embedded.go` |

### 2. Application Composition

The main app creates all modules and manages their lifecycle:

```go
// cmd/fecim-visualizer/main.go (simplified)

type DemoApp struct {
    demo1 *demo1gui.EmbeddedApp           // Hysteresis
    demo2 *demo2gui.EmbeddedCrossbarApp   // Crossbar
    demo3 *demo3gui.EmbeddedDualModeApp   // MNIST
    demo4 *demo4gui.EmbeddedCircuitsApp   // Circuits
    demo5 *demo5gui.EmbeddedComparisonApp // Comparison
    demo6 *demo6gui.EmbeddedEDAApp        // EDA
}

// Usage
demos := &DemoApp{
    demo1: demo1gui.NewEmbeddedApp(),
    demo2: demo2gui.NewEmbeddedCrossbarApp(),
    // ...
}

// Build all UIs once
views := []fyne.CanvasObject{
    launcherContent,
    container.NewMax(demos.demo1.BuildContent(fyneApp, window)),
    container.NewMax(demos.demo2.BuildContent(fyneApp, window)),
    // ...
}

// Create stack
contentStack := container.NewStack(views...)

// Lifecycle management
onViewChange = func(index int) {
    // Stop old demo
    demos.demo1.Stop() // if was running
    // Start new demo
    demos.demo1.Start() // if now active
}
```

This pattern ensures:
- **Isolation**: Each module manages its own state
- **Lazy execution**: Only active modules consume CPU
- **Clean transitions**: Proper cleanup when switching tabs

---

## Physics Layer

### 1. 30-Level Quantization

The core of FeCIM is the **30 discrete analog states**—based on Dr. Tour's research (COSM 2025 transcript):

```go
// module2-crossbar/pkg/crossbar/array.go

const FeCIMLevels = 30  // Dr. Tour: "30 discrete states"

// QuantizeTo30Levels quantizes any value to one of 30 levels
func QuantizeTo30Levels(value float64) float64 {
    value = math.Max(0, math.Min(1, value))           // Clamp to [0, 1]
    level := math.Round(value * float64(FeCIMLevels-1)) // Round to 0-29
    return level / float64(FeCIMLevels-1)              // Return normalized
}

// Example: 0.5 → 15/29 ≈ 0.517 (level 15)
```

This is used everywhere weights are programmed.

### 2. Hysteresis Model (Module 1)

The Preisach model simulates ferroelectric polarization switching:

```go
// module1-hysteresis/pkg/ferroelectric/preisach.go

type MayergoyzPreisach struct {
    material *HZOMaterial
    hysterons map[string]*Hysteron  // Elementary hysteresis operators
    levels   int                     // Discretization levels (30 for FeCIM)
}

// Material parameters from Nature Communications 2025
type HZOMaterial struct {
    Pr float64  // Remanent polarization (15-34 µC/cm²)
    Ec float64  // Coercive field (1.0-1.5 MV/cm)
    // ...
}

// SimulateStep applies electric field and returns new polarization
func (p *MayergoyzPreisach) SimulateStep(eField float64) float64 {
    // Update each hysteron based on field
    // Sum contributions
    // Return total polarization
}
```

**Physics accuracy**:
- Remanent polarization: 15-34 µC/cm² (verified from literature)
- Coercive field: 1.0-1.5 MV/cm (verified)
- Endurance: 10¹² cycles (target, shown in literature)

### 3. Crossbar Array Simulation (Module 2)

The Array class simulates matrix-vector multiplication with non-idealities:

```go
// module2-crossbar/pkg/crossbar/array.go

type Array struct {
    config *Config
    cells  [][]Cell  // [row][col]
    // ...
}

type Cell struct {
    Conductance float64     // Programmed weight (0-1, quantized to 30 levels)
    NoiseFactor float64     // Device variation
    SwitchingCount int64    // Endurance tracking
}

// ComputeMVM performs matrix-vector multiply with non-idealities
func (a *Array) ComputeMVM(input []float64) []float64 {
    output := make([]float64, a.config.Rows)

    // Ideal accumulation
    for i := 0; i < a.config.Rows; i++ {
        for j := 0; j < a.config.Cols; j++ {
            output[i] += input[j] * a.cells[i][j].Conductance
        }
    }

    // Apply non-idealities
    output = a.applyIRDrop(output)       // Voltage drop across wires
    output = a.applySneakPaths(output)   // Leakage current
    output = a.applyDrift(output)        // Device parameter drift

    return output
}
```

**Non-idealities implemented**:

| Effect | File | Impact |
|--------|------|--------|
| **IR Drop** | `irdrop.go` | Voltage loss in interconnects, scales with current |
| **Sneak Paths** | `sneakpath.go` | Unintended current paths through unselected cells |
| **Drift** | `drift.go` | Gradual conductance change with time/cycles |

---

## GUI Layer

### 1. Fyne Framework

All modules use **Fyne**, a modern Go GUI framework with these characteristics:

- **Cross-platform**: Windows, macOS, Linux, iOS, Android
- **Canvas-based**: Raw drawing with pixel-level control
- **Widget-based**: Pre-built components (buttons, sliders, containers)
- **Responsive**: Scales to different screen sizes
- **Thread-safe**: All UI updates via `fyne.Do()`

### 2. Thread Safety Pattern

Every goroutine that updates the UI must use `fyne.Do()`:

```go
// ✓ CORRECT: Update from goroutine
go func() {
    result := computeHeavyWork()
    fyne.Do(func() {
        label.SetText(result)  // Safe: runs on main thread
    })
}()

// ✗ WRONG: Direct update from goroutine
go func() {
    result := computeHeavyWork()
    label.SetText(result)  // CRASH: not on main thread
}()
```

All modules follow this pattern religiously.

### 3. Shared Theme

The `shared/theme` package provides consistent branding across all modules:

```go
// shared/theme/theme.go

var (
    ColorPrimary = color.RGBA{0, 212, 255, 255}   // Cyan accent
    ColorBackground = color.RGBA{0, 50, 100, 255} // Dark blue
    ColorText = color.RGBA{240, 244, 248, 255}    // Off-white
    ColorSuccess = color.RGBA{87, 204, 153, 255}  // Green
    ColorError = color.RGBA{255, 107, 107, 255}   // Red
    // ...
)

// Implements fyne.Theme
type FeCIMTheme struct{}

func (t *FeCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
    // Maps Fyne theme requests to FeCIM colors
    // Result: Consistent dark blue + cyan theme across all tabs
}
```

**Contrast ratios** (WCAG AA compliance):
- ColorBackground vs ColorText: ~15:1 (excellent)
- ColorBackground vs ColorPrimary: ~4.5:1 (good for accents)

### 4. Adaptive Layout for Responsive UI

The `shared/widgets/AdaptiveLayout` class enables responsive design:

```go
// shared/widgets/adaptive_layout.go

type AdaptiveLayout struct {
    zones []fyne.CanvasObject  // Content to reparent
    tabLabels []string          // Mobile tab names
    currentBreakpoint Breakpoint // SM/MD/LG/XL
}

// Desktop layout (XL/LG: > 768px)
// ┌──────────────────────────┐
// │ Canvas      │ Controls   │
// │             │ Info       │
// └──────────────┴────────────┘

// Mobile layout (MD/SM: < 768px)
// ┌────────────────────┐
// │  [Draw] [Results]  │ <- Tabs
// │                    │
// │    Current zone    │
// └────────────────────┘

// Usage in Module 3 (MNIST)
adaptive := NewAdaptiveLayout(
    []fyne.CanvasObject{drawingZone, resultsZone, weightsZone},
    []string{"Draw", "Results", "Weights"},
)
adaptive.SetDesktopLayout(func(zones []fyne.CanvasObject) fyne.CanvasObject {
    return container.NewHSplit(
        container.NewVBox(zones[0], zones[2]),  // Canvas + Weights
        zones[1],                                // Results
    )
})
```

**Breakpoints**:
- **SM** (< 480px): Single column, mobile mode
- **MD** (480-768px): Mobile mode, limited splits
- **LG** (768-1024px): Dual column
- **XL** (> 1024px): Full desktop layout

### 5. Custom Widgets

Each module implements custom widgets for specialized visualization:

| Module | Widget | Purpose |
|--------|--------|---------|
| **1** | `PEPlot` | Real-time P-E curve plotting |
| **1** | `LevelWidget` | 30-level bar indicator |
| **2** | `CrossbarHeatmap` | Conductance visualization |
| **3** | `DigitCanvas` | Mouse drawing surface |
| **4** | Various | Circuit schematic elements |
| **5** | ChartWidget | Interactive comparison bars |

Custom widgets extend `widget.BaseWidget`:

```go
type CrossbarHeatmap struct {
    widget.BaseWidget
    array *crossbar.Array
    selectedRow, selectedCol int
}

func (h *CrossbarHeatmap) CreateRenderer() fyne.WidgetRenderer {
    return &crossbarHeatmapRenderer{
        heatmap: h,
        objects: []fyne.CanvasObject{}, // canvas objects to render
    }
}

func (r *crossbarHeatmapRenderer) Layout(size fyne.Size) {
    // Position all canvas objects (rectangles, text, etc.)
}

func (r *crossbarHeatmapRenderer) Refresh() {
    // Recalculate colors based on current state
}
```

---

## Data Flow

### Example: Module 2 (Crossbar) Data Flow

```
User Input
    │
    ├─ Program Weight (slider)
    │   ├─ Quantize to 30 levels
    │   ├─ Update Array[row][col]
    │   └─ Queue UI refresh
    │
    ├─ Run MVM (button)
    │   ├─ Array.ComputeMVM(input)
    │   │   ├─ Ideal multiply
    │   │   ├─ Apply IR drop
    │   │   ├─ Apply sneak paths
    │   │   └─ Apply drift
    │   │
    │   ├─ Collect output
    │   ├─ Collect metrics (IR drop %, sneak path %, etc.)
    │   └─ Update displays
    │
    └─ Visualization
        ├─ Heatmap refresh (conductance colors)
        ├─ Output graph (result values)
        ├─ Metrics panel (IR drop, sneak paths)
        └─ Animation (optional: step-by-step visualization)
```

**Rate limiting**:
- Simulation can run at full speed (100+ updates/sec)
- UI updates throttled to ~60 FPS (16.7 ms between refreshes)
- Large heatmaps (64x64) batch canvas updates

### Example: Module 1 (Hysteresis) Data Flow

```
Auto-Mode / Manual Input
    │
    ├─ Generate E-field waveform
    │   ├─ Sine, sawtooth, or custom
    │   └─ Quantize to 8-bit DAC steps
    │
    ├─ For each time step:
    │   ├─ Update Preisach model
    │   │   └─ Apply E-field to all hysterons
    │   ├─ Read polarization
    │   ├─ Buffer point (E, P)
    │   └─ Check for 30-level transitions
    │
    └─ Render loop (every ~50ms):
        ├─ P-E curve plot
        ├─ 30-level histogram
        ├─ Waveform display
        ├─ Material properties
        └─ Statistics (Ec, Pr, etc.)
```

---

## Threading Model

### Main Thread (Fyne Event Loop)

The Fyne runtime manages a single main thread that:
1. Processes user input (mouse, keyboard, window events)
2. Executes all `fyne.Do()` callbacks
3. Renders frames to screen

### Background Goroutines (Per Module)

Each module typically runs 1-2 goroutines:

```go
// Module 1: Hysteresis
type App struct {
    running bool  // Controlled by Start()/Stop()
}

func (a *App) Start() {
    a.running = true
    go a.simulationLoop()  // Runs until a.running = false
}

func (a *App) simulationLoop() {
    ticker := time.NewTicker(50 * time.Millisecond)  // 20 Hz
    defer ticker.Stop()

    for {
        if !a.running {
            return  // Clean exit when Stop() called
        }

        select {
        case <-ticker.C:
            // Do physics simulation
            result := a.preisach.SimulateStep(eField)

            // Update UI from main thread
            fyne.Do(func() {
                a.updatePEPlot(result)
                a.updateLevelWidget(result)
            })
        }
    }
}

func (a *App) Stop() {
    a.running = false  // Loop will exit naturally
}
```

**Design principles**:
1. Physics simulation runs on background goroutine (non-blocking)
2. UI updates batched and dispatched via `fyne.Do()`
3. Clean shutdown via flag + `Stop()` method
4. No blocking operations on main thread

### Recording Goroutine (FFmpeg)

The main app can optionally record video:

```go
// cmd/fecim-visualizer/main.go

type RecordingState struct {
    isRecording bool
    stopChan    chan struct{}  // Signal to stop
    stdin       io.WriteCloser // Pipe to FFmpeg
}

func (rs *RecordingState) captureLoop(width, height int) {
    ticker := time.NewTicker(50 * time.Millisecond)  // 20 FPS
    defer ticker.Stop()

    for {
        select {
        case <-rs.stopChan:
            return  // Stop signal received
        case <-ticker.C:
            // Capture frame on main thread
            var img image.Image
            done := make(chan struct{})
            fyne.Do(func() {
                img = window.Canvas().Capture()
                close(done)
            })
            <-done

            // Convert and write to FFmpeg (background)
            rs.stdin.Write(frameBytes)
        }
    }
}
```

---

## Concurrency Patterns

### Pattern 1: Simulation Loop

Used by Module 1 (Hysteresis), Module 3 (MNIST), Module 5 (Comparison):

```go
func (a *App) simulationLoop() {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        if !a.running {
            return
        }

        select {
        case <-ticker.C:
            // Compute (background)
            state := a.compute()

            // Update UI (main thread)
            fyne.Do(func() {
                a.refresh(state)
            })
        }
    }
}
```

**Advantages**:
- Predictable timing
- Non-blocking UI
- Easy to start/stop

### Pattern 2: On-Demand Computation

Used by Module 2 (Crossbar), Module 4 (Circuits):

```go
// Button handler
func onRunMVM() {
    go func() {
        // Compute on background thread
        result := computeIntensive()

        // Update UI
        fyne.Do(func() {
            displayResult(result)
        })
    }()
}
```

**Advantages**:
- Responds immediately to user action
- No wasted cycles when module idle
- Simple to implement

---

## Key Design Decisions

### 1. Why Unified Launcher?

Originally, each module was standalone. The unified launcher was introduced to:
- Tell a coherent story (physics → compute → application → system → business → design)
- Share common infrastructure (theme, logging, adaptive layout)
- Reduce startup time (load all modules once)
- Enable smooth transitions between concepts

### 2. Why EmbeddedApp Interface?

Modules could use different patterns (tabbed apps, separate windows, etc.), but the interface enforces:
- Consistent API: `BuildContent()`, `Start()`, `Stop()`
- Lifecycle management: modules only consume resources when active
- Testability: each module can be tested in isolation
- Reusability: modules could be embedded in other applications

### 3. Why Preisach Model for Hysteresis?

The Preisach model (vs. simpler Tanh-based models) provides:
- Physical accuracy: matches real ferroelectric material behavior
- Memory: hysteron interaction history affects current output
- 30-level validation: can verify discrete state transitions
- Extensibility: supports advanced analysis (loss, remnants, etc.)

### 4. Why Canvas-Based Rendering?

Fyne provides both high-level widgets and low-level canvas. Modules use canvas for:
- **High-level**: Lists, text, buttons (fast to code)
- **Low-level**: Heatmaps, plots, custom shapes (custom widgets)

This hybrid approach balances development speed and visual polish.

### 5. Why No State Serialization?

The application doesn't save/restore simulation state between sessions because:
- Educational purpose: each session is a fresh learning experience
- Complexity: non-trivial to serialize Preisach hysteresis model state
- Simplicity: avoids versioning issues
- However: window size and last-viewed tab ARE saved for UX continuity

---

## Dependencies and Versions

### Go Modules (go.mod)

```
fyne v2              - GUI framework
image/* (stdlib)     - Image processing
math (stdlib)        - Numerics
sync (stdlib)        - Concurrency primitives
os, io, fmt (stdlib) - Standard I/O
```

### No External Physics Libraries

The physics is implemented from scratch:
- Preisach model: custom implementation
- Crossbar array: custom simulation
- Neural network: forward inference only (weights from disk)

This ensures:
- No licensing conflicts
- Full control over accuracy
- Transparent simulation
- Educational clarity

---

## Build & Testing

### Build

```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer
./fecim-visualizer
```

### Tests

```bash
go test ./...                            # All tests
go test ./module1-hysteresis/pkg/...    # Module 1 only
go test -race ./...                      # Race condition detection
```

**Test coverage**:
- Physics models: unit tests for quantization, Preisach, MVM
- Non-idealities: IR drop, sneak paths, drift accuracy
- GUI: custom widget rendering (canvas object count, positions)

### Debugging

```bash
# Verbose logging (3 levels: 1=info, 2=debug, 3=trace)
./fecim-visualizer -verbosity=2

# Resize debugging (if FYNE_DEBUG_RESIZE=1)
FYNE_DEBUG_RESIZE=1 ./fecim-visualizer
```

---

## Performance Considerations

### Simulation Performance

- **Hysteresis**: 100+ steps/second (limited by UI refresh rate)
- **Crossbar MVM**: 1000+ operations/second
- **MNIST inference**: 30+ images/second
- **Bottleneck**: UI rendering (heatmaps, plots)

### Memory Usage

- **Crossbar Array (64×64)**: ~64KB (cells + conductances)
- **Preisach Model**: ~100KB (hysterons for 30 levels)
- **MNIST Weights**: ~5MB (trained model)
- **Total**: ~10-20MB per module instance

### Optimization Techniques

1. **Canvas batching**: Group updates before refresh
2. **Heatmap caching**: Only redraw changed cells
3. **Plot decimation**: Skip frames if UI can't keep up
4. **Quantization early**: 30-level quantization happens at program time, not render time

---

## Known Limitations & TODOs

### Current Limitations

1. **Module 6 (EDA)**: Work in progress—placement algorithm incomplete
2. **Recording**: Requires FFmpeg installed on system
3. **No persistence**: Simulation state not saved between sessions
4. **Fyne Wayland**: Minor window resize oscillation on some Wayland compositors (mitigated)

### Future Improvements

1. Multi-cell arrays in Module 1 (array simulation)
2. Advanced MVM analysis (sneak path current tracing visualization)
3. Custom neural network training in Module 3
4. More chip peripheral types in Module 4
5. Behavioral model export for circuit simulators (SPICE)
6. EDA routing algorithm completion

---

## References

- **COSM 2025**: Dr. external research group on FeCIM: [ironlattice-transcript.md](../videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md)
- **Nature Communications 2025**: HfO₂-ZrO₂ superlattice material parameters
- **IEEE IRPS 2022**: Endurance and reliability data
- **Preisach Model**: Classical hysteresis modeling (research papers in module1 docs)
- **Fyne Documentation**: https://fyne.io/

---

## Contributing

When adding new features to the architecture:

1. **Respect module independence**: No cross-module imports in physics/gui code
2. **Use EmbeddedApp pattern**: New modules must implement `BuildContent()`, `Start()`, `Stop()`
3. **Follow theme**: Use `shared/theme` colors, not hardcoded colors
4. **Thread safety**: Always use `fyne.Do()` for UI updates from goroutines
5. **Document physics**: Physics models should have references to literature or first-principles
6. **Test concurrency**: Run `go test -race` to catch data races

---

**Last Updated**: 2026-01-25
**Architecture Version**: 3.0 (Unified Launcher)
**Modules**: 6 (Hysteresis, Crossbar, MNIST, Circuits, Comparison, EDA)
