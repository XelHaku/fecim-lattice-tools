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

FeCIM Lattice Tools is a unified educational visualization suite for ferroelectric compute-in-memory research and teaching. The application launches multiple modules from a single entry point, each focused on one part of the simulation stack.

### Entry Point: cmd/fecim-lattice-tools

The main application (`cmd/fecim-lattice-tools/main.go`) implements a **unified launcher** pattern:

```
┌─────────────────────────────────────────────┐
│  FeCIM Lattice Tools (main.go)              │
│  ────────────────────────────────────────   │
│  Fyne App                                   │
│  ├─ Window (1400x900 default, resizable)   │
│  └─ ContentStack (8 views)                  │
│     ├─ View 0: Home/Launcher               │
│     ├─ View 1: Module 1 (Hysteresis)       │
│     ├─ View 2: Module 2 (Crossbar)         │
│     ├─ View 3: Module 3 (MNIST)            │
│     ├─ View 4: Module 4 (Circuits)         │
│     ├─ View 5: Module 5 (Comparison)       │
│     ├─ View 6: Module 6 (EDA)              │
│     └─ View 7: Module 7 (Documentation)    │
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

### The 7-Module Story

Each module demonstrates a layer in the FeCIM stack:

| Module | Purpose | Physics | GUI Framework | Status |
|--------|---------|---------|---------------|--------|
| **1. Hysteresis** | Memory cell physics | Preisach (GUI) + Landau‑Khalatnikov (headless) | Real-time P-E curve + ISPP demo | Stable |
| **2. Crossbar** | Matrix-vector multiply | Ohm's law, IR drop, sneak paths, drift | Heatmap with MVM operations | Stable |
| **3. MNIST** | Neural network application | Network inference with quantized weights | Digit drawing + recognition | Stable |
| **4. Circuits** | Peripheral electronics | DAC/ADC/TIA circuit modeling | Schematic diagrams, waveforms | Stable |
| **5. Comparison** | Technology benchmarks | Energy, speed, density metrics | Interactive comparison charts | Stable |
| **6. EDA** | Chip design tools | Placement, routing (experimental) | Layout visualization | WIP |
| **7. Documentation** | In-app docs viewer | N/A | Search, navigation, glossary | Stable |

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
                    │  ├─ pkg/core/             │
                    │  │  ├─ network.go         │
                    │  │  ├─ network_inference.go│
                    │  │  └─ quantize.go        │
                    │  ├─ pkg/mnist/            │
                    │  │  └─ loader.go          │
                    │  └─ pkg/gui/              │
                    │     └─ embedded.go        │
                    └──────────────────────────┘
                            │
           ┌────────────────┴────────────────┐
           │                                 │
    ┌──────▼──────────────┐          ┌──────▼──────────────┐
    │ module4-circuits    │          │ module5-comparison  │
    │ ├─ gui/            │          │ ├─ data/           │
    │ │  └─ embedded.go  │          │ ├─ metrics.go      │
    │ └─ uses shared/    │          │ └─ gui/            │
    │    peripherals     │          │    └─ embedded.go  │
    └───────────────────┘          └─────────────────────┘
            │
            │
    ┌───────▼──────────────┐
    │ module6-eda          │
    │ ├─ placement/        │
    │ ├─ netlist/          │
    │ └─ gui/              │
    │    └─ embedded.go    │
    └──────────────────────┘
            │
    ┌───────▼──────────────┐
    │ module7-docs         │
    │ └─ gui/              │
    │    ├─ embedded.go    │
    │    ├─ navigation.go  │
    │    ├─ search.go      │
    │    ├─ layout.go      │
    │    ├─ persistence.go │
    │    └─ glossary_int...│
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
    // BuildContent returns the module UI for the current host app/window.
    // Implementations must be idempotent for the same embedding context and
    // may rebuild when the host context changes.
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
| Documentation | `EmbeddedDocsApp` | `module7-docs/pkg/gui/embedded.go` |

Module 7 (Documentation) is curriculum-first: it indexes `docs/documentation` by default, orders
module folders first with the research-papers index next, exposes module quick-access
shortcuts for ELI5/PHYSICS/FEATURES/OPENSOURCE-TOOLS pages, and includes quick links to the
curriculum overview, module index, and research index.

### 2. Application Composition

The main app creates all modules and manages their lifecycle:

```go
// cmd/fecim-lattice-tools/main.go (simplified)

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

// Bind each module to the current host app/window
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
- **Stable embedding**: `BuildContent(...)` reuses existing content for the same host window and rebuilds only when embedding context changes
- **Lazy execution**: Only active modules consume CPU
- **Clean transitions**: Proper cleanup when switching tabs

Most embedded modules implement this through `shared/widgets/EmbeddedAppBase.BuildOrReuseContent(...)`,
which centralizes host-context tracking and keeps `BuildContent(...)` deterministic across repeated calls.

---

## Physics Layer

### 1. 30-Level Quantization

This demo uses **30 configurable analog states per cell** as the default educational simulator profile:

```go
// module2-crossbar/pkg/crossbar/array.go

const FeCIMLevels = 30  // Simulation baseline: "30 discrete states"

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

Module 1 implements ferroelectric hysteresis simulation with two complementary models and comprehensive material support.
The **GUI path** renders the Preisach model in real time; the **headless diagnostics path**
(`cmd/fecim-lattice-tools --mode hysteresis`) exercises the L‑K solver in `shared/physics/` and the
shared WriteController state machine in `module1-hysteresis/pkg/controller/`.

#### Preisach Model Architecture

**Classical Preisach Model** (`preisach.go`):
- S-shaped switching with history tracking
- Switching function: `P = Ps * tanh((E - Ec_eff) / delta)`
- LIFO stack of turning points for minor loop closure

**Advanced Mayergoyz Model** (`preisach_advanced.go`):
- Full hysteron grid on Preisach plane (α > β region)
- 2D Gaussian distribution `μ(α, β)` across 50×50 hysteron grid
- Temperature-corrected coercive field: `Ec(T) = Ec0 * (1 - T/Tc)^0.5`
- Domain switching via KAI model: `P(t) = Ps * (1 - exp(-(t/τ)^n))`

```go
// module1-hysteresis/pkg/ferroelectric/preisach_advanced.go

type MayergoyzPreisach struct {
    material    *HZOMaterial
    hysterons   [][]int8           // 50×50 grid of bistable units (±1)
    distribution [][]float64       // 2D Gaussian weighting
    gridSize    int                // Fixed at 50 (physics resolution)
    levels      int                // User-selectable 2-256 (default 30)
}

// Update applies electric field and returns new polarization
func (p *MayergoyzPreisach) Update(E float64) float64 {
    // For each hysteron: switch if |E| exceeds threshold
    // Weight contributions by distribution μ(α, β)
    // Apply temperature and fatigue corrections
    // Return total polarization
}

// DiscreteStates maps polarization to conductance levels
func (p *MayergoyzPreisach) DiscreteStates(N int) []DiscreteState {
    // Inverse tanh mapping: E ≈ Ec + δ * arctanh(P/Ps)
    // Conductance: G = 1 µS + (normalizedP * 99 µS)
}
```

#### Landau‑Khalatnikov Solver (Dynamic Equation)

For time‑domain physics (and equation‑level verification), the project includes a
first‑order **Landau‑Khalatnikov (L‑K)** solver in `shared/physics/landau.go`.

Core equation (as implemented):

```
ρ_eff * dP/dt = E_eff - (2αP + 4βP^3 + 6γP^5)
E_eff = E_applied - K_dep * P
α(T,σ) = (T - Tc) / (2 ε0 C) - 2 Q12 σ
ρ_eff = ρ + (R_series * A / d)
```

Key parameter mapping (from `HZOMaterial` → `LKSolver`):

- `BetaLandau`, `GammaLandau`, `RhoViscosity`
- `K_dep` (depolarization slope)
- `StressGPa` + `Q12` (electrostriction)
- `SeriesResistanceOhm`, `Thickness`, `Area`
- `CurieTemp`, `CurieConst`
- `Tau0NLS`, `EaNLS` (Merz‑law incubation)

Usage:

- **Headless diagnostics**: `cmd/fecim-lattice-tools/mode.go` runs an L‑K sweep and a **multi‑target ISPP sequence**
  (`pos-1`, `pos-2`, `neg-1`) to exercise both branches without resetting between every step.
- **ISPP controller**: `module1-hysteresis/pkg/controller` drives write/verify sequencing (shared by GUI + headless).
- **ISPP conservative bounds**: after the first post‑cross undershoot, the upper bound is tightened once by a
  factor that scales with `|P_target|/Ps` (0.2–0.6 of the remaining bracket) to reduce overshoot on the next midpoint.
- **ISPP branch crossing**: binary-search midpoints stay low-biased while `currentP * targetP < 0` using
  `bias = 0.1 + 0.2 * |P_target|/Ps` (clamped ~0.1-0.3 of the bracket) to reduce overshoot resets, then
  revert to midpoint 0.5 after the branch is crossed.
- **Logs**: `lk-solver` (equation terms) + `ispp` (write/verify loop) provide headless validation evidence.

**Authoritative validation:** The headless hysteresis path (`--mode hysteresis`) is the **source of truth**
for physics and ISPP correctness. GUI runs are illustrative; acceptance requires a headless log with
`lk-solver` + `ispp` evidence.

#### Material System

Eight built-in materials with reported in literature parameters:

| Material | Pr (µC/cm²) | Ec (MV/cm) | Endurance | Use Case |
|----------|-------------|------------|-----------|----------|
| **DefaultHZO** | 25 | 1.2 | 10¹⁰ | Si-doped baseline |
| **FeCIMMaterial** | 30 | 1.0 | 10⁹ | Educational simulator baseline (default) |
| **FeCIMMaterialTarget** | 30 | 1.0 | 10¹² | Aspirational target |
| **LiteratureSuperlattice** | 45 | 0.8 | 10¹⁰ | Cheema 2020 (NC benefit) |
| **CryogenicHZO** | 75 | 1.5 | 10¹⁰ | 4K operation |
| **HZOStandard32** | 20 | 1.2 | 10⁸ | 32-state benchmark |
| **HZOFJT140** | 25 | 1.0 | 10⁸ | 140-state FTJ |
| **AlScN** | 120 | 5.0 | 10⁸ | High-Pr alternative |

```go
// module1-hysteresis/pkg/ferroelectric/material.go

type HZOMaterial struct {
    Name            string
    Pr              float64 // Remanent polarization (C/m²)
    Ps              float64 // Saturation polarization (C/m²)
    Ec              float64 // Coercive field (V/m)
    Thickness       float64 // Film thickness (m)
    Area            float64 // Active area (m²)
    CurieTemp       float64 // Curie temperature (K)
    CurieConst      float64 // Curie constant (K)

    // Landau-Khalatnikov parameters
    BetaLandau      float64 // J m^5 / C^4
    GammaLandau     float64 // J m^9 / C^6
    RhoViscosity    float64 // Ohm·m
    K_dep           float64 // V·m/C

    // Coupling / parasitics
    Q12             float64 // m^4 / C^2
    StressGPa       float64 // GPa
    SeriesResistanceOhm float64 // Ohms

    // NLS (Merz law)
    Tau0NLS         float64 // s
    EaNLS           float64 // V/m

    EnduranceCycles float64 // Cycles to 10% Pr loss
    RetentionTime   float64 // Retention time at RT (s)
    ImrintField     float64 // Imprint field (V/m)
}

// Physics methods
func (m *HZOMaterial) CoerciveVoltage() float64           // Ec * Thickness
func (m *HZOMaterial) SwitchingTime(T float64) float64    // Arrhenius: τ0 * exp(Ea/kB*T)
func (m *HZOMaterial) CoerciveFieldAtTemp(T float64)      // Tc-scaled Ec
func (m *HZOMaterial) EnduranceAtCycles(N float64)        // Stretched exponential
func (m *HZOMaterial) RetentionAtTime(t, T float64)       // Temperature-accelerated loss
```

#### Modeled Physics Phenomena

| Phenomenon | Implementation | Location |
|------------|----------------|----------|
| **Hysteresis loops** | Asymmetric ascending/descending branches | `preisach.go` |
| **Minor loops** | LIFO history stack for turning points | `preisach.go` |
| **Domain dynamics** | KAI model time-dependent switching | `preisach_advanced.go` |
| **Temperature effects** | Tc-scaled Ec, Pr, τ | `material.go` |
| **Wake-up** | Distribution enhancement (0.8 → 1.0) | `preisach_advanced.go` |
| **Fatigue** | Stretched exponential: `Pr(N) = Pr0 * exp(-(N/N0)^β)` | `preisach_advanced.go` |
| **Imprinting** | Preferential switching bias | `material.go` |

#### Simulation Engine

The simulation engine (`simulation/engine.go`) provides time-stepping:

```go
type Engine struct {
    preisach    *MayergoyzPreisach
    dt          float64              // Time step (default 1 ns)
    maxHistory  int                  // Circular buffer size (1000)
    mu          sync.RWMutex         // Thread safety
}

// Waveform types
const (
    WaveformManual   // Direct voltage control
    WaveformSine     // V(t) = A * sin(2πft)
    WaveformTriangle // Linear ramp -A to +A
    WaveformSquare   // ±A pulse train
)

// RunRealtime executes simulation with UI callback
func (e *Engine) RunRealtime(callback func(State), targetFPS int)
```

#### GUI Visualization Components

Custom Fyne widgets for hysteresis visualization:

| Widget | File | Purpose |
|--------|------|---------|
| **PEPlot** | `widgets/peplot.go` | Real-time P-E curve with Ec/Pr markers |
| **CellVisualizer** | `widgets/cell.go` | Blue→White→Red polarization gradient |
| **LevelIndicator** | `widgets/level.go` | 30-level clickable bar |
| **ModeIndicator** | `widgets/mode.go` | WRITE (red) / READ (green) status |

#### Write/Read Demo State Machine

The GUI includes a 7-phase demo for FeCIM operations:

```
Phase 0: SATURATE   → Apply ±Emax to set initial state
Phase 1: SETTLE     → Apply intermediate field for target level
Phase 2: HOLD       → Maintain E for charge relaxation
Phase 3: READ       → Return to E=0, capture level
Phase 4: DISPLAY    → Hold for visual confirmation
Phase 5: RETENTION  → Track drift at E=0 over time
Phase 6: VERIFY     → Re-read to confirm write success
```

#### Calibration System

Multi-temperature calibration with persistent storage:

- **Key temperatures**: 233K (-40°C), 273K, 300K (RT), 373K, 423K (150°C)
- **Algorithm**: Binary search with oscillation detection
- **Storage**: `data/hysteresis_calibration.json` (schema v2)
- **Migration**: Handles legacy v1 single-temp format

#### Performance Characteristics

| Metric | Value |
|--------|-------|
| **Hysteron grid** | 50×50 = 2,500 bistable units |
| **Update complexity** | O(2,500) per time step |
| **Memory usage** | ~500 KB (grid + distribution) |
| **Frame rate** | 60 FPS with 30-100 physics steps/frame |
| **Calibration** | O(30 × 10) binary searches |

**Physics accuracy**:
- Remanent polarization: 15-75 µC/cm² (material-dependent, reported in literature)
- Coercive field: 0.6-5.0 MV/cm (reported in literature)
- Endurance: 10⁸-10¹² cycles (material-dependent)
- Bits per cell: 4.91 bits for 30 levels, configurable 1-8 bits

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

### Example: Module 4 (Peripheral Circuits) Data Flow

```
Mode Selection (READ / WRITE / COMPUTE)
    │
    ├─ Load material + calibration (physics.yaml)
    │   ├─ Coercive voltage → read/write ranges
    │   └─ DAC range mode (read vs write)
    │
    ├─ Configure word lines + DAC preset
    │   ├─ READ: single WL active, read-range DAC
    │   ├─ WRITE: single WL active, write-range DAC
    │   └─ COMPUTE: all WLs active, input vector DAC
    │
    ├─ Signal chain execution (DeviceState)
    │   ├─ DAC → Array (G × V) → Row current
    │   ├─ TIA (gain-scaled for MVM) → Voltage
    │   └─ ADC → Digital levels + saturation flags
    │
    ├─ Passive 0T1R write handling
    │   └─ DAC-only column drive (all WL=0V, selected BL=−V_write; full column disturb)
    │
    └─ UI update
        ├─ Array canvas refresh
        ├─ Per‑row current/voltage/level labels
        └─ Status + mode guidance
```

**Interfaces used**:
- `shared/peripherals` (DAC/ADC/TIA/ChargePump + analysis)
- `shared/physics` (HZOMaterial, conductance mapping)
- `config/physics.yaml` (FieldMinRatio/FieldMaxRatio calibration)
- Timing/power baselines from `shared/peripherals.AnalyzeTiming/AnalyzePower` (read ~76ns, write ~203ns).

### Example: Module 1 (Hysteresis) Data Flow

```
Auto-Mode / Manual Input
    │
    ├─ Generate E-field waveform
    │   ├─ Sine, triangle, or demo sequence
    │   └─ Apply ramp limits (Ec-based bounds)
    │
    ├─ Physics path (per time step)
    │   ├─ Preisach loop (GUI visualization)
    │   │   └─ Apply E-field to hysteron stack → P
    │   ├─ Landau-Khalatnikov diagnostics (headless mode)
    │   │   └─ LKSolver.Step(E, dt) → P (logs equation terms)
    │   └─ Quantize P → discrete level (0..N-1)
    │
    ├─ Write/Read Demo (when enabled)
    │   ├─ RESET to opposite saturation (±2×Ec)
    │   ├─ ISPP loop (Apply → Wait → Verify → Adjust)
    │   └─ Overshoot → deep reset → binary-search restart
    │
    └─ Render loop (every ~50ms)
        ├─ P‑E curve plot
        ├─ 30-level indicator
        ├─ Waveform display
        ├─ Material properties
        └─ Statistics (Ec, Pr, ISPP success rate)
```

**Headless mode shortcut (authoritative):** `--mode hysteresis` bypasses the GUI and runs
`FeCIMMaterial()` → `LKSolver.Step()` → `WriteController.WriteTargetWithReset()` with log output only.
This is the acceptance gate for physics fidelity; the GUI path is not used for equation validation.

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
// cmd/fecim-lattice-tools/main.go

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
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
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
./fecim-lattice-tools -verbosity=2

# Resize debugging (if FYNE_DEBUG_RESIZE=1)
FYNE_DEBUG_RESIZE=1 ./fecim-lattice-tools
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

- **Educational simulator baseline**: 30 configurable analog states per cell
- **Literature-backed FeCIM references**: DOI-linked source material in the research index
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

**Last Updated**: 2026-01-29
**Architecture Version**: 3.1 (Unified Launcher + Docs)
**Modules**: 7 (Hysteresis, Crossbar, MNIST, Circuits, Comparison, EDA, Documentation)
**Total Tests**: See CI (`go test ./...`)
