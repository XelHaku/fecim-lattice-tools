# Crossbar Array MVM Visualization - Demo Guide

**FeCIM Lattice Tools - Module 2: Crossbar Demo**

> Interactive visualization of Matrix-Vector Multiplication (MVM) in ferroelectric crossbar arrays.

---

## Overview

Demo 2 provides an interactive visualization of how Ferroelectric CIM performs analog neural network inference using physical Ohm's law and Kirchhoff's current law, achieving massive parallelism.

### What This Demo Shows

1. **Matrix-Vector Multiplication (MVM)** — Parallel analog computation using conductance × voltage = current
2. **30 Discrete Conductance Levels** — Each cell stores ~4.9 bits (30 states) of synaptic weight
3. **Non-Idealities Modeling** — IR drop, sneak paths, device variation, ADC quantization
4. **Real-time Crossbar Visualization** — Interactive heatmap with cell-level inspection

---

## Quick Start

```bash
# Navigate to project root
cd <local-path>

# Build unified app
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Run and select "Crossbar" tab
./fecim-lattice-tools

# OR: Launch directly
./launch.sh
```

### Standalone Demo

```bash
# Navigate to module directory
cd module2-crossbar

# Build the GUI application
go build -o crossbar-gui ./cmd/crossbar-gui

# Run Fyne GUI mode (recommended)
./crossbar-gui

# Enhanced mode with all features
./crossbar-gui -enhanced
```

---

## Run Modes

### 1. Embedded Mode (Recommended)

Launch from the unified visualizer:

```bash
./fecim-lattice-tools
# Select "Crossbar" tab from the tabbed interface
```

**Features:**
- Integrated with other modules
- Consistent theme and layout
- Shared logging and configuration

### 2. Standalone GUI Mode

```bash
./crossbar-gui
```

Cross-platform native GUI featuring:
- **Interactive heatmap visualization** with click-to-select cells
- **Three tabbed views**: Conductance, IR Drop, Sneak Paths
- **Real-time control panel** with sliders:
  - Array size (8×8 to 128×128)
  - Noise level (0-20%)
  - ADC resolution (4-10 bits)
- **Custom "Ferroelectric CIM" colormap** matching 30 discrete levels
- **30-level discrete indicator widget**
- **Vector bar charts** for input/output visualization
- **One-click MVM, IR Drop, and Sneak Path analysis**
- **RMSE comparison charts** (ideal vs actual)
- **Live statistics panel**

### 3. Enhanced Mode

```bash
./crossbar-gui -enhanced
```

Includes all standard features plus:
- **Color legends** - Shows what colors mean (Level 0-29, %, etc.)
- **Metrics panel** - Real-time accuracy, energy, and performance
- **Before/After toggle** - Side-by-side ideal vs actual comparison
- **Accuracy waterfall** - Step-by-step degradation visualization
- **Enhanced MVM animation** - Smooth, informative 3-phase animation
- **Comparison badge** - "FeCIM vs GPU: 10,000× better" widget
- **Data export** - CSV (weights) and JSON (analysis) export

**GUI Controls:**
| Control | Function |
|---------|----------|
| Array Size Slider | Resize crossbar (8×8 to 128×128) |
| Noise Slider | Device-to-device variation (0-20%) |
| ADC Bits Slider | ADC resolution (4-10 bits) |
| Colormap Dropdown | ferroelectric-cim, viridis, plasma, coolwarm |
| Run MVM | Execute matrix-vector multiplication |
| Analyze IR Drop | Show voltage drop heatmap |
| Analyze Sneak Paths | Show sneak current map |
| Reset Array | Reprogram random weights |

**Heatmap Interaction:**
- Click any cell to see its conductance level (0-29)
- Right-click to clear selection
- Tabs switch between Conductance, IR Drop, and Sneak Path views
- Yellow border highlights selected/worst-case cells

---

## Physics Model

### Matrix-Vector Multiplication (MVM)

The crossbar array performs parallel MVM using Ohm's law:

```
Input Vector (Voltages)
    V₀  V₁  V₂  V₃  V₄  V₅  V₆  V₇
    ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓
   ┌───┬───┬───┬───┬───┬───┬───┬───┐
I₀ │G₀₀│G₀₁│G₀₂│G₀₃│G₀₄│G₀₅│G₀₆│G₀₇│→ I₀ = Σⱼ Gᵢⱼ × Vⱼ
I₁ │G₁₀│G₁₁│G₁₂│G₁₃│G₁₄│G₁₅│G₁₆│G₁₇│→ I₁ = Σⱼ Gᵢⱼ × Vⱼ
I₂ │G₂₀│G₂₁│G₂₂│G₂₃│G₂₄│G₂₅│G₂₆│G₂₇│→ I₂ = Σⱼ Gᵢⱼ × Vⱼ
   └───┴───┴───┴───┴───┴───┴───┴───┘

Output Current = Weight Matrix × Input Vector
     I        =        G      ×      V
```

**Formula:**
```
I[i] = Σⱼ G[i,j] × V[j]
```

Where:
- `G[i,j]` = Conductance of cell (i,j) — represents synaptic weight
- `V[j]` = Input voltage on column j
- `I[i]` = Output current on row i

### Non-Idealities Modeled

| Effect | Description | Impact | Mitigation |
|--------|-------------|--------|------------|
| **IR Drop** | Voltage attenuation along wires due to resistance | Cells far from driver see lower voltage | Add driver amplifiers, limit array size |
| **Sneak Paths** | Parasitic currents through unselected cells | Corrupts output currents | Use 1T1R (transistor) or selector devices |
| **Device Variation** | Cell-to-cell conductance spread (σ/μ) | Reduces effective precision | Scheme C programming, calibration |
| **ADC Quantization** | Limited output bit precision | Quantization noise | Higher resolution ADC (6+ bits) |

---

## Demo Features

### 1. Conductance Matrix View

Shows the programmed weights as a heatmap:

```
        WEIGHT HEATMAP
┌─────────────────────┐
│ ▓▓▓ ░░░ ▓░░ ░▓░ ▓▓▓ │  Row 0
│ ░▓░ ▓░░ ▓▓▓ ░░░ ░▓░ │  Row 1
│ ▓░░ ░░░ ░▓░ ▓▓▓ ░░░ │  Row 2
└─────────────────────┘

▓ = High conductance (level 20-29)
▒ = Medium conductance (level 10-19)
░ = Low conductance (level 0-9)
```

**Click on cell** to see:
- FeCIM level (0-29)
- Conductance (µS)
- Resistance (kΩ)
- Normalized value [0,1]
- Bit representation

### 2. IR Drop Analysis

Visualizes voltage distribution across the array:

```
        VOLTAGE MAP
┌─────────────────────┐
│ 1.0  0.98 0.96 0.94 │  Voltage decreases
│ 0.98 0.96 0.94 0.92 │  from top-left
│ 0.96 0.94 0.92 0.90 │  to bottom-right
│ 0.94 0.92 0.90 0.88 │  (wire resistance)
└─────────────────────┘
```

**Shows:**
- Effective voltage at each cell
- Voltage drop percentage
- Word line and bit line voltages
- Distance from drivers
- Mitigation strategies

### 3. Sneak Path Visualization

Highlights parasitic current paths:

```
     TARGET CELL: [2,2]
┌─────────────────────┐
│  -    -    -    -   │
│  -    -    -    -   │
│  ↓    ↓    ●    ↓   │  ● = target
│  -    -    -    -   │  ↓ = sneak path
└─────────────────────┘
     Same column affected
```

**Displays:**
- Sneak current magnitude
- Signal-to-noise ratio (SNR)
- Path classification (ROW/COL/DIAG)
- Mitigation options

### 4. MVM Animation

Watch the computation happen in real-time:

**Phase 1:** Input voltages applied to columns (cyan highlight)
**Phase 2:** Current flows through cells (wave animation)
**Phase 3:** Output currents collected from rows (orange highlight)

**Duration:** ~1.1 seconds total (smooth 60 FPS)

### 5. Accuracy Analysis (Enhanced Mode)

Waterfall chart showing degradation sources:

```
Ideal: 90.0% ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ↓ -0.8%
Quantization: 89.2% ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ↓ -1.7%
IR Drop: 87.5% ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ↓ -0.7%
Variation: 86.8% ▓▓▓▓▓▓▓▓▓▓▓▓▓▓
  ↓ -1.6%
Final: 85.2% ▓▓▓▓▓▓▓▓▓▓▓▓▓

Target (Dr. Tour): 87% ←
```

---

## Architecture

```
module2-crossbar/
├── cmd/
│   └── crossbar-gui/
│       └── main.go           # Entry point
├── pkg/
│   ├── crossbar/             # Array modeling
│   │   ├── array.go          # Crossbar structure & MVM
│   │   ├── cell.go           # FeFET/FTJ cell model
│   │   ├── wire.go           # Wire resistance model
│   │   ├── nonidealities.go  # IR drop, sneak paths
│   │   └── enhanced.go       # Integrated simulation
│   ├── gui/
│   │   ├── app.go            # Main application
│   │   ├── app_enhanced.go   # Enhanced layout
│   │   ├── embedded.go       # Embeddable interface
│   │   ├── heatmap.go        # CrossbarHeatmap widget
│   │   ├── controls.go       # ControlPanel, StatsPanel
│   │   ├── vectors.go        # VectorBarChart widgets
│   │   ├── widgets.go        # Enhanced widgets
│   │   └── tooltips.go       # Tooltip system
│   └── compute/              # Future: Vulkan compute
│       ├── mvm.go            # MVM kernel
│       └── nonideal.go       # Non-ideality injection
└── shaders/                  # SPIR-V shaders (future)
    ├── mvm.comp              # MVM compute shader
    ├── crossbar.vert         # Grid vertex shader
    └── crossbar.frag         # Cell color shader
```

---

## Tests

```bash
# Run all tests
cd module2-crossbar
go test ./...

# Run crossbar package tests
go test ./pkg/crossbar -v

# Run with verbose non-idealities tests
go test ./pkg/crossbar -v -run TestNonidealities
```

Test coverage:
- MVM correctness verification
- IR drop calculation
- Sneak path current analysis
- 30-level conductance quantization
- Non-ideality impact on accuracy

---

## Benchmarks (from Literature)

| Architecture | MNIST Accuracy | Source |
|--------------|----------------|--------|
| 24×24 FE Memristor (sim) | 98.78% | ScienceDirect 2025 |
| Multi-Level FeFET 28nm (sim) | 96.6% | Nature Comms 2023 |
| FTJ Crossbar (sim) | 92% | SemiEngineering 2024 |
| **Ferroelectric CIM Hardware** | **87%** | Dr. Tour COSM 2025 |

**Important Note:** Dr. Tour's hardware claims are unverified (conference presentation only). Our simulation may achieve higher accuracy because it doesn't capture all hardware non-idealities.

---

## Troubleshooting

### GUI fails to start

**Linux:** Install required dependencies:
```bash
# Debian/Ubuntu
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel \
    libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel
```

### Array computation is slow

For large arrays (>64×64):
- Use GPU acceleration (future Vulkan compute shader)
- Reduce visualization update frequency
- Consider chunked MVM for memory

---

## Demo Flow (For Video/Presentation)

1. **Open** - Window shows crossbar heatmap
2. **Point out** color legend ("See the 30 levels")
3. **Click** "Run MVM"
4. **Watch** animation (1.1 seconds, smooth)
5. **Show** metrics panel updating
6. **Navigate** to "IR Drop" tab
7. **Click** worst-case cell (corner)
8. **Show** tooltip with analysis
9. **Navigate** to "Sneak Paths" tab
10. **Point at** cross pattern
11. **Navigate** to "Accuracy Analysis" tab (enhanced mode)
12. **Point at** waterfall chart

**Total time:** 2 minutes for full demo

---

## Related Documentation

- **[Physics Deep Dive](crossbar.physics.md)** - Complete technical reference
- **[ELI5 Explanation](crossbar.ELI5.md)** - Simple analogies
- **[Research Papers](crossbar.research.md)** - Academic citations
- **[Implementation Notes](../CLAUDE.md)** - Developer guide

---

## License

Part of the FeCIM Lattice Tools project.

---

**Source:** Based on Dr. external research group's HfO₂-ZrO₂ superlattice research (COSM 2025)
**Implementation:** FeCIM Lattice Tools visualization suite
**Documentation:** See docs/crossbar/ for complete references
