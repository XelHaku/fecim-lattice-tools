# Ferroelectric CIM Visualizer

**GPU-Accelerated Ferroelectric Compute-in-Memory Demos**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Fyne](https://img.shields.io/badge/Fyne-2.4-blue?logo=go)](https://fyne.io)
[![Vulkan](https://img.shields.io/badge/Vulkan-1.3-AC162C?logo=vulkan)](https://www.vulkan.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Demos](https://img.shields.io/badge/Demos-8%2F8-brightgreen.svg)]()

---

> **DISCLAIMER**: Ferroelectric CIM is at **TRL 4** (lab validation only). Hardware achieved **87% MNIST** (88% theoretical max). Energy claims (10M× vs NAND) are from Dr. Tour's presentation and have not been independently verified. See [HONESTY_AUDIT.md](docs/opensource/papers/08_Documentation/HONESTY_AUDIT.md).

---

## The Story: 8 Demos (ALL COMPLETE)

```
Demo 1: "This is how the memory cell works"      ✅ Fyne GUI - P-E Hysteresis
Demo 2: "This is how we compute in memory"       ✅ Fyne GUI - Crossbar MVM
Demo 3: "This is what we can build with it"      ✅ Fyne GUI - MNIST 87%
Demo 4: "This is how it fits in a real chip"     ✅ Fyne GUI - Peripheral Circuits
Demo 5: "This is how we manage heat"             ✅ Fyne GUI - Thermal Analysis
Demo 6: "This is how we scale to 3D"             ✅ Fyne GUI - 3D Multilayer Stack
Demo 7: "This is what can go wrong"              ✅ Fyne GUI - Non-Idealities
Demo 8: "This is why it beats everything else"   ✅ Fyne GUI - Technology Comparison
```

---

## Quick Start

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Build and run the unified visualizer (ALL 8 DEMOS)
go build ./cmd/fecim-visualizer && ./fecim-visualizer

# Or run individual demos:
# Demo 1: Ferroelectric Hysteresis (P-E curve, 30 levels)
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis

# Demo 2: Crossbar MVM (IR drop, sneak paths, heatmaps)
cd demo2-crossbar && go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui

# Demo 3: MNIST Neural Network (draw digits, watch inference)
cd demo3-mnist && go build -o mnist-gui ./cmd/mnist-gui && ./mnist-gui

# Demo 4: Peripheral Circuits (DAC, ADC, timing)
cd demo4-circuits && go build -o circuits-gui ./cmd/circuits-gui && ./circuits-gui

# Demo 5: Thermal Simulation
cd demo5-thermal && go build -o thermal ./cmd/thermal && ./thermal

# Demo 6: 3D Multilayer Stack
cd demo6-multilayer && go build -o multilayer-gui ./cmd/multilayer-gui && ./multilayer-gui

# Demo 7: Non-Idealities Analysis
cd demo7-nonidealities && go build -o nonidealities-gui ./cmd/nonidealities-gui && ./nonidealities-gui

# Demo 8: Technology Comparison
cd demo8-comparison && go build -o comparison-gui ./cmd/comparison-gui && ./comparison-gui

# Run all tests (157 tests)
go test ./...
```

---

## Why Ferroelectric CIM Matters

> *"This could lower data center energy by 80 to 90%."*
> — Dr. external research group, external research institution

| What | Traditional | Ferroelectric CIM |
|------|-------------|-------------|
| Memory states | 2 (0/1) | **30 levels** |
| Compute location | Separate CPU/GPU | **In the memory** |
| Data movement | Constant | **Zero** |
| Energy vs NAND | 1× | **10,000,000×** lower* |

*\*Claimed, not independently verified*

---

## Demo Details

### Demo 1: Ferroelectric Hysteresis ✅

**Purpose:** Understand single cell physics

```
      P                    ┌───────────┐
      ↑     ╭────╮         │ ████ 30   │
   +Pr├─────╯    │         │ ████ 29   │
      │          │         │ ▓▓▓▓ ...  │
   ───┼──────────┼───→ E   │ ░░░░ 1    │
   -Pr├──────────╯         │      0    │
      ↓                    │ 30 LEVELS │
                           └───────────┘
```

**Features:**
- Real-time P-E hysteresis curve with fade trail
- 30 discrete levels visualized
- Mayergoyz Preisach model
- Material selector (Default HZO, Optimized, Ferroelectric CIM)
- Waveform modes (Sine, Triangle, Square, Manual)

---

### Demo 2: Crossbar Array MVM ✅

**Purpose:** Understand compute-in-memory

```
     V₀   V₁   V₂   V₃  (input voltages)
      │    │    │    │
 ─────●────●────●────●───→ I₀
      │    │    │    │
 ─────●────●────●────●───→ I₁  (output currents)
      │    │    │    │
 ─────●────●────●────●───→ I₂

 ●=conductance (30 levels, color coded)
```

**Features:**
- Interactive heatmap with click-to-select cells
- IR drop analysis with wire resistance modeling
- Sneak path current visualization
- Three tabbed views: Conductance, IR Drop, Sneak Paths

---

### Demo 3: MNIST Neural Network ✅

**Purpose:** See real AI application

> **Note:** Ferroelectric CIM hardware achieved **87%** with **88% theoretical max** (Dr. Tour). Our simulation uses idealized conditions and may exceed real hardware.

```
┌─────────┐    ┌─────────┐    ┌─────────┐
│ 28 × 28 │    │ 784×128 │    │ 128×10  │
│  INPUT  │ ─→ │ Layer 1 │ ─→ │ Layer 2 │ ─→ Prediction
│  DIGIT  │    │ Crossbar│    │ Crossbar│
└─────────┘    └─────────┘    └─────────┘
```

**Features:**
- Interactive 28×28 digit drawing canvas
- Real-time inference as you draw
- Layer activation visualization
- Confusion matrix with clickable cells
- Per-class metrics (precision, recall, F1)

---

### Demo 4: Peripheral Circuits ✅

**Purpose:** Understand full chip system

```
WRITE PATH                 READ PATH

Digital: [22]             Digital: [22]
    │                          ↑
    ▼                          │
┌───────┐                  ┌───────┐
│  DAC  │                  │  ADC  │
│ 5-bit │                  │ 5-bit │
└───┬───┘                  └───┬───┘
    │                          ↑
    ▼                          │
┌─────────────────────────────────────┐
│            CROSSBAR ARRAY           │
└─────────────────────────────────────┘
```

**Features:**
- DAC/ADC conversion visualization
- Charge pump operation
- TIA (Transimpedance Amplifier)
- INL/DNL linearity analysis
- Timing diagrams
- Power breakdown

---

### Demo 5: Thermal Simulation ✅

**Purpose:** Engineering analysis

```
Top View (Heat Map)        Side View

░░░▒▒▓▓████▓▓▒▒░░░        ███ Layer 3
░░▒▒▓██████████▓▒▒░░       ↕ heat
░▒▓████████████████▓▒░     ███ Layer 2
░░▒▒▓██████████▓▒▒░░       ↕ heat
░░░▒▒▓▓████▓▓▒▒░░░         ███ Layer 1
                           ░░░ Heat Sink
25°C ░▒▓█ 85°C
```

**Features:**
- 2D heat map visualization
- Real-time heat diffusion
- Hotspot identification
- Thermal throttling warnings

---

### Demo 6: Multi-Layer 3D Stack ✅

**Purpose:** Visualize 3D stacking for massive parallelism

```
                    ┌─────────────────┐
    Layer 3 ────→   │ 64×10 (Output)  │
                    └───────┬─────────┘
                            │ vias
                    ┌───────┴─────────┐
    Layer 2 ────→   │ 128×64 (Hidden) │
                    └───────┬─────────┘
                            │ vias
                    ┌───────┴─────────┐
    Layer 1 ────→   │ 784×128 (Input) │
                    └─────────────────┘
```

**Features:**
- 3D isometric visualization of stacked layers
- Layer selector and highlight
- Via network visualization
- Layer specifications display
- Energy comparison with traditional compute

---

### Demo 7: Non-Idealities ✅

**Purpose:** Real-world engineering challenges

```
Tab 1: IR Drop Analysis
┌──────────────────────┐
│ ■■■□□□□  V_applied   │
│ ■■□□□□□  V_actual    │
│ Worst corner: (15,15)│
└──────────────────────┘

Tab 2: Sneak Path Analysis
┌──────────────────────┐
│ Target: X            │
│ Parasitic paths: ─── │
│ SNR: 24.5 dB        │
└──────────────────────┘

Tab 3: Conductance Drift
┌──────────────────────┐
│ FeCIM: 0.001%/decade │
│ RRAM: 0.1%/decade   │
│ PCM: 1%/decade      │
└──────────────────────┘
```

**Features:**
- IR drop voltage gradient heatmap
- Sneak path current visualization
- Conductance drift time-series plot
- Technology comparison (FeCIM vs RRAM vs PCM)
- Mitigation strategies

---

### Demo 8: Technology Comparison ✅

**Purpose:** The bottom line — why FeCIM wins

```
Energy per MAC (fJ)
═══════════════════════════════════════════
CPU+DRAM  ████████████████████████████ 1000
GPU+HBM   ██████████                    100
FeCIM     █                              10

Von Neumann              Compute-in-Memory
┌──────┐   ←→   ┌──────┐     ┌────────────┐
│ CPU  │  data  │ Mem  │     │ Mem=Compute│
└──────┘        └──────┘     │  No move   │
  Bottleneck!                └────────────┘
```

**Features:**
- Energy per MAC bar chart comparison
- Architecture diagrams (Von Neumann vs CIM)
- Data center power calculator
- Verified vs claimed specifications
- Source citations and disclaimers

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| GUI | Fyne 2.4 |
| GPU | Vulkan 1.3 |
| Shaders | GLSL → SPIR-V |
| Physics | Preisach/Mayergoyz model |
| Neural Network | Crossbar MVM simulation |
| Tests | 157 passing |

---

## Repository Structure

```
multilayer-ferroelectric-cim-visualizer/
├── cmd/
│   └── fecim-visualizer/  ✅ Unified GUI (ALL 8 DEMOS)
├── demo1-hysteresis/      ✅ Single cell P-E curve (Fyne GUI)
├── demo2-crossbar/        ✅ Crossbar MVM (Fyne GUI)
├── demo3-mnist/           ✅ MNIST classifier (Fyne GUI)
├── demo4-circuits/        ✅ Peripheral circuits (Fyne GUI)
├── demo5-thermal/         ✅ Thermal simulation (Fyne GUI)
├── demo6-multilayer/      ✅ 3D multi-layer (Fyne GUI)
├── demo7-nonidealities/   ✅ Non-idealities (Fyne GUI)
├── demo8-comparison/      ✅ Technology comparison (Fyne GUI)
├── shared/                Shared packages (theme, logging)
├── docs/                  Documentation
└── go.mod
```

---

## The Team Behind Ferroelectric CIM

| Person | Role |
|--------|------|
| **Dr. external research group** | Principal Investigator, external research institution |
| **Dr. Jaeho Shin** | Device Engineer, Superlattice Inventor |
| **Tawfik Jarjour** | Commercialization Lead |

---

## Key Quotes from Dr. Tour

> *"It's got 30 discrete states. So it's not 0-1-0-1."*

> *"We're at 87% validation here... theoretical is 88%."*

> *"Compute in memory where the same device does the memory and the computation."*

> *"This could lower the requirements in a data center by 80 to 90%."*

> *"Works on a standard CMOS line and can translate just like that."*

---

## License

MIT License

Ferroelectric CIM is a trademark of its respective owners at external research institution. This is an independent educational visualization project with no affiliation.

---

*8/8 demos complete. The future of computing is here.*

*Built with Go, Fyne, and curiosity.*
