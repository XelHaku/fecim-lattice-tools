# Ferroelectric CIM Visualizer

**Educational visualization suite for Ferroelectric Compute-in-Memory (FeCIM) technology**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Fyne](https://img.shields.io/badge/Fyne-2.7.2-blue?logo=go)](https://fyne.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)]()
[![Modules](https://img.shields.io/badge/Modules-6-brightgreen.svg)]()

> **"Compute in memory where the same device does the memory and the computation."**
> — Dr. external research group, external research institution

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Modules](#modules)
  - [Module 1: Hysteresis](#module-1-ferroelectric-hysteresis-)
  - [Module 2: Crossbar + Non-Idealities](#module-2-crossbar-mvm--non-idealities--4-tabs)
  - [Module 3: MNIST Neural Network](#module-3-mnist-neural-network--flagship)
  - [Module 4: Peripheral Circuits](#module-4-peripheral-circuits-)
  - [Module 5: Technology Comparison](#module-5-technology-comparison--investor-pitch)
  - [Module 6: Design Suite (EDA)](#module-6-fecim-design-suite--chip-design-tool)
- [Why FeCIM Matters](#why-ferroelectric-cim-matters)
- [Technical Stack](#technical-stack)
- [Repository Structure](#repository-structure)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

FeCIM Visualizer demonstrates Dr. external research group's research on HfO₂-ZrO₂ superlattice-based memory devices. Unlike traditional binary memory (0/1), FeCIM supports **30 analog states per cell**, enabling ~4.9 bits/cell storage and efficient neural network inference.

> **DISCLAIMER**: Ferroelectric CIM is at **TRL 4** (lab validation only). Hardware achieved **87% MNIST** accuracy (88% theoretical max). Energy claims (10M× vs NAND) are from Dr. Tour's presentation and have not been independently verified. See [HONESTY_AUDIT.md](docs/opensource/papers/08_Documentation/HONESTY_AUDIT.md).

---

## Quick Start

```bash
# Clone and run
git clone https://github.com/XelHaku/multilayer-ferroelectric-cim-visualizer.git
cd multilayer-ferroelectric-cim-visualizer
./launch.sh
```

Or build manually:

```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer
```

---

## Installation

### Prerequisites

- **Go 1.24+** — [Download](https://go.dev/dl/)
- **C compiler** (gcc/clang) for CGO
- **OpenGL libraries**

### Linux (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev
go mod download
./launch.sh
```

### Linux (Fedora/RHEL)

```bash
sudo dnf install -y gcc mesa-libGL-devel libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
go mod download
./launch.sh
```

### macOS

```bash
xcode-select --install  # Install command line tools
go mod download
./launch.sh
```

### Windows

1. Install [MSYS2](https://www.msys2.org/) or [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
2. Ensure `gcc` is in your PATH
3. Run: `go build -o fecim-visualizer.exe ./cmd/fecim-visualizer`

### Running Tests

```bash
go test ./...                              # All tests
go test -v ./module2-crossbar/pkg/crossbar # Crossbar tests only
```

---

## Modules

The visualizer includes 6 interconnected modules that tell the FeCIM story:

```
PHYSICS → COMPUTE → APPLICATION → SYSTEM → BUSINESS → TOOLING

┌────────────┐    ┌────────────┐    ┌────────────┐
│  Module 1  │───▶│  Module 2  │───▶│  Module 3  │
│ Hysteresis │    │  Crossbar  │    │   MNIST    │
│  30 levels │    │  + Noise   │    │    87%     │
└────────────┘    └────────────┘    └────────────┘
      │                                    │
      ▼                                    ▼
┌────────────┐    ┌────────────┐    ┌────────────┐
│  Module 4  │◀───│  Module 5  │◀───│  Module 6  │
│  Circuits  │    │ Comparison │    │    EDA     │
│    CMOS    │    │  Business  │    │   Suite    │
└────────────┘    └────────────┘    └────────────┘
```

| Module | Focus | Description |
|--------|-------|-------------|
| **1. Hysteresis** | Physics | P-E curve, Preisach model, 30 discrete levels |
| **2. Crossbar** | Compute | MVM operations + non-idealities (4 tabs) |
| **3. MNIST** | Application | Neural network digit recognition (87% accuracy) |
| **4. Circuits** | System | DAC/ADC/TIA peripheral design |
| **5. Comparison** | Business | Technology benchmarks, technical briefing |
| **6. EDA Suite** | Tooling | Chip design and fabrication export |

---

## Why Ferroelectric CIM Matters

> *"This could lower data center energy by 80 to 90%."*
> — Dr. external research group, external research institution

### The Memory Wall Problem

Traditional computing moves data constantly between memory and processor — this data movement consumes most of the energy in modern systems. FeCIM eliminates this by computing directly where data is stored.

| Aspect | Traditional | FeCIM |
|--------|-------------|-------|
| Memory states | 2 (binary) | **30 levels** (4.9 bits/cell) |
| Compute location | Separate CPU/GPU | **In the memory itself** |
| Data movement | Constant bottleneck | **Zero** |
| Energy vs NAND | 1× | **10,000,000× lower*** |
| CMOS compatible | N/A | **Yes** (standard fab) |

*\*Energy claims from Dr. Tour's presentation; not independently verified*

### Key Specifications

| Metric | Value | Notes |
|--------|-------|-------|
| Discrete levels | 30 | Per memory cell |
| Bits per cell | 4.91 | log₂(30) |
| MNIST accuracy | 87% | Hardware validated (88% theoretical max) |
| Technology Readiness | TRL 4 | Lab validation complete |

---

## Module Details

### Module 1: Ferroelectric Hysteresis ✅

> *"It's got 30 discrete states. So it's not 0-1-0-1."* — Dr. Tour

Visualizes single-cell ferroelectric physics using the Mayergoyz Preisach model.

```
Polarization (P)              30 Discrete Levels
      ↑     ╭────╮            ┌───────────┐
   +Pr├─────╯    │            │ ████ 30   │
      │          │            │ ████ 29   │
   ───┼──────────┼───→ E      │ ▓▓▓▓ ...  │
   -Pr├──────────╯            │ ░░░░ 1    │
      ↓                       └───────────┘
```

**Features:**
- Real-time P-E hysteresis curve with fade trail
- 30 discrete levels visualization
- Material presets (Default HZO, Optimized, FeCIM)
- Waveform modes: Sine, Triangle, Square, Manual

---

### Module 2: Crossbar MVM + Non-Idealities ✅ (4 Tabs)

Matrix-vector multiplication (MVM) via Kirchhoff's current law, plus real-world challenges.

```
     V₀   V₁   V₂   V₃  (input)        I_out[i] = Σ G[i,j] × V_in[j]
      │    │    │    │
 ─────●────●────●────●───→ I₀          ● = conductance (30 levels)
      │    │    │    │
 ─────●────●────●────●───→ I₁          Analog multiply-accumulate
      │    │    │    │                 in O(1) time
 ─────●────●────●────●───→ I₂
```

| Tab | Focus | Key Features |
|-----|-------|--------------|
| **Ideal MVM** | Baseline | Interactive cell programming, MVM visualization |
| **IR Drop** | Wire resistance | Voltage gradient heatmap, metal width mitigation |
| **Sneak Paths** | Parasitic currents | SNR analysis, selector device modeling |
| **Drift** | Temporal variation | 10-year retention, FeCIM vs ReRAM vs PCM |

---

### Module 3: MNIST Neural Network ✅ (Flagship Demo)

> *"We're at 87% validation here... theoretical is 88%."* — Dr. Tour

Interactive digit recognition comparing full-precision vs CIM inference.

```
┌─────────┐    ┌─────────┐    ┌─────────┐
│ 28 × 28 │───▶│ 784×128 │───▶│ 128×10  │───▶ Prediction
│  Input  │    │ Layer 1 │    │ Layer 2 │     (0-9)
└─────────┘    └─────────┘    └─────────┘
   Drawing       Crossbar       Crossbar
   Canvas         Array          Array
```

**Features:**
- Interactive 28×28 drawing canvas
- **Dual-mode:** Full Precision vs CIM side-by-side
- Adjustable: quantization levels, noise, ADC/DAC bits
- Failure mode presets (Ideal, Quant Cliff, Noisy, Broken ADC)
- Weight visualization with 30-level color coding
- Guided Tour mode (7 steps)

---

### Module 4: Peripheral Circuits ✅

> *"Works on a standard CMOS line and can translate just like that."* — Dr. Tour

Complete chip system with analog/digital interfaces.

```
WRITE PATH                    READ PATH
Digital [22] ──▶ DAC ──┐  ┌── ADC ──▶ Digital [22]
                       ▼  ▲
              ┌────────────────────┐
              │   CROSSBAR ARRAY   │
              │    (30 levels)     │
              └────────────────────┘
```

**Features:**
- DAC/ADC conversion visualization
- Charge pump and TIA (Transimpedance Amplifier)
- INL/DNL linearity analysis
- Timing diagrams and power breakdown
- CMOS compatibility checklist

---

### Module 5: Technology Comparison ✅ (Technical Briefing)

The business case for FeCIM vs competing technologies.

```
Energy per MAC (fJ)                    Competitive Matrix
                                       ┌──────────┬──────┬──────┬──────┐
CPU+DRAM  ████████████████████ 1000    │ Feature  │FeCIM │ReRAM │ PCM  │
GPU+HBM   ████████              100    ├──────────┼──────┼──────┼──────┤
FeCIM     █                      10    │ Energy   │  ✅  │  🟡  │  🟡  │
                                       │ Speed    │  ✅  │  ✅  │  ❌  │
                                       │ Endurance│  ✅  │  ❌  │  🟡  │
                                       │ 30 levels│  ✅  │  ❌  │  ✅  │
                                       └──────────┴──────┴──────┴──────┘
```

**Features:**
- Energy per MAC comparison charts
- Technology matrix (FeCIM vs NAND vs ReRAM vs PCM vs MRAM)
- **Data center savings calculator** (GPU count → annual savings)
- Market opportunity ($403B by 2030)
- TRL progression roadmap
- Verified vs claimed specifications with sources

---

### Module 6: FeCIM Design Suite ✅ (EDA Tool)

Design FeCIM chips for fabrication with OpenLane/OpenROAD integration.

```
Specification ──▶ Physical Layout ──▶ Fabrication Files
┌────────────┐    ┌──────────────┐    ┌────────────────┐
│ Mode: Store│    │ 4×4 FeFET    │    │ .v  (Verilog)  │
│ Size: 256² │───▶│ Array Grid   │───▶│ .def (Layout)  │
│ Tech: SKY130    │ WL/BL Routes │    │ .sp  (SPICE)   │
└────────────┘    └──────────────┘    └────────────────┘
```

**Design Modes:**

| Mode | Application | Use Case |
|------|-------------|----------|
| **Storage** | NAND replacement | High-density storage (4.9 bits/cell) |
| **Memory** | DRAM replacement | Fast zero-refresh memory |
| **Compute** | AI accelerator | Analog MVM for neural networks |

**Example:**
```bash
go run ./cmd/eda-cli -mode storage -rows 4 -cols 4 -name hello_storage
```

**Tabs:** Configure → Layout → HDL → Explorer → Simulate → Export → Learn

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24+ |
| GUI Framework | Fyne 2.7.2 |
| Physics Model | Preisach/Mayergoyz |
| Compute | Crossbar MVM simulation |
| Non-Idealities | IR drop, sneak paths, drift |

---

## Repository Structure

```
fecim-lattice-tools/
├── cmd/fecim-visualizer/    # Unified GUI entry point
├── module1-hysteresis/      # P-E curve physics
├── module2-crossbar/        # MVM + non-idealities
├── module3-mnist/           # Neural network demo
├── module4-circuits/        # Peripheral circuits
├── module5-comparison/      # Technology benchmarks
├── module6-eda/             # Design suite
├── shared/                  # Common theme, logging
├── docs/                    # Documentation, archive
└── go.mod
```

---

## Contributing

Contributions are welcome. Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Follow existing code patterns (see `CLAUDE.md` for conventions)
4. Run tests (`go test ./...`)
5. Submit a pull request

---

## Research Team

| Person | Role |
|--------|------|
| **Dr. external research group** | Principal Investigator, external research institution |
| **Dr. Jaeho Shin** | Device Engineer, Superlattice Inventor |
| **Tawfik Jarjour** | Commercialization Lead |

---

## License

MIT License

This is an independent educational visualization project. Ferroelectric CIM research originates from external research institution. No official affiliation.

---

<p align="center">
<i>Built with Go, Fyne, and curiosity.</i>
</p>
