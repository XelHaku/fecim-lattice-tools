# Ferroelectric CIM Lattice Tools

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

FeCIM Lattice Tools demonstrates ferroelectric compute-in-memory (FeCIM) technology based on Dr. external research group's HfO₂-ZrO₂ superlattice research at external research institution. Unlike traditional binary memory (0/1), FeCIM supports **30 discrete analog states per cell** (~4.9 bits/cell) as demonstrated in Dr. Tour's COSM 2025 presentation [1]. Similar multi-level capabilities (32-140 states) have been independently verified in peer-reviewed literature [2][3].

> **DISCLAIMER**: Ferroelectric CIM is at **TRL 4** (lab validation) per Dr. Tour's own statement at COSM 2025 [1]. The **30 states** and **87% MNIST** claims are from Dr. Tour's presentation; similar results appear in peer-reviewed literature [2][4]. Energy efficiency vs NAND is **25-100×** (Samsung Nature 2025 [5]). Dr. Tour's "10M× vs NAND" claim has been **removed** from this project as no peer-reviewed data supports it.

---

## Quick Start

```bash
# Clone and run
git clone https://github.com/your-org/fecim-lattice-tools.git
cd fecim-lattice-tools
./launch.sh
```

Or build manually:

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
```

---

## Installation

### Prerequisites

- **Go 1.24+** — [Download](https://go.dev/dl/)
- **C compiler** (gcc/clang) for CGO
- **OpenGL libraries**

### Optional Dependencies

- **Docker** — For Module 6 EDA tools (OpenLane/OpenROAD/KLayout)
- **Graphviz** — For Yosys circuit schematic visualization (`sudo apt install graphviz`)

### Linux (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev
# Optional: for Module 6 Yosys schematic visualization
sudo apt-get install -y graphviz
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
3. Run: `go build -o fecim-lattice-tools.exe ./cmd/fecim-lattice-tools`

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
| **3. MNIST** | Application | Neural network digit recognition (87-98% accuracy) |
| **4. Circuits** | System | DAC/ADC/TIA peripheral design |
| **5. Comparison** | Business | Technology benchmarks, technical briefing |
| **6. EDA Suite** | Tooling | Chip design and fabrication export |

---

## Why Ferroelectric CIM Matters

> *Compute-in-memory can reduce energy consumption by 50-80% for memory-bound workloads, which account for up to 80% of execution time in modern datacenters.*
> — Peer-reviewed CIM literature [7][8]

### The Memory Wall Problem

Traditional computing moves data constantly between memory and processor — this data movement consumes most of the energy in modern systems. FeCIM eliminates this by computing directly where data is stored.

| Aspect | Traditional | FeCIM |
|--------|-------------|-------|
| Memory states | 2 (binary) | **30 levels** (4.9 bits/cell) |
| Compute location | Separate CPU/GPU | **In the memory itself** |
| Data movement | Constant bottleneck | **Zero** |
| Energy vs NAND | 1× | **25-100× lower** [5] |
| CMOS compatible | N/A | **Yes** (standard fab) |

*Energy comparison from Samsung FeFET research [5]. Higher improvements (up to 70,000×) reported for AI inference vs GPUs [6].*

### Key Specifications

| Metric | Value | Notes |
|--------|-------|-------|
| Discrete levels | 30 | Dr. Tour COSM 2025 [1]; peer-reviewed: 32-140 [2][3] |
| Bits per cell | 5-7+ | log₂(32)=5 to log₂(140)≈7 |
| MNIST accuracy | 87-98% | Dr. Tour: 87% [1]; Peer-reviewed: 96.6-98.24% [4][9] |
| Endurance | 10⁹-10¹² | IEEE IRPS 2022 [10]; V:HfO₂ 2024 [11] |
| 3D Integration | 22nm BEOL | CEA-Leti December 2024 [12] |
| Cryogenic | 5K-300K | +25% memory window at 14K [13] |
| Automotive | Grade 0 | AEC-Q100 qualified [14] |
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

> *"We're at 87% validation here."* — Dr. Tour (Note: Conference claim, unverified; Peer-reviewed record: 98.24% [9])

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
├── cmd/fecim-lattice-tools/    # Unified GUI entry point
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

## References

[1] Dr. external research group, "Ferroelectric CIM: Ultra-Low-Power AI Computing," COSM 2025 Technology Summit, November 2024. [Transcript](docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md) - Primary source for 30 states, 87% MNIST, TRL 4 status

[2] M. Jerry et al., "Ferroelectric FET analog synapse for acceleration of deep neural network training," IEEE IEDM 2017. DOI: 10.1109/IEDM.2017.8268338 (32 states demonstrated)

[3] C.-M. Song et al., "Ferroelectric 2D SnS2 Analog Synaptic FET," Advanced Science, 2024. DOI: 10.1002/advs.202308588 (140 levels demonstrated)

[4] "First in-memory computing crossbar using multi-level FeFET," Nature Communications, 2023. DOI: 10.1038/s41467-023-42110-y (96.6% accuracy, 7 VT states)

[5] "Ferroelectric transistors for low-power NAND flash memory," Nature, 2025. DOI: 10.1038/s41586-025-09793-3 (94-96% energy reduction = 25-100× improvement)

[6] "Analog in-memory computing attention mechanism for large language models," Nature Computational Science, 2025. DOI: 10.1038/s43588-025-00854-1 (70,000× energy efficiency vs GPU)

[7] "Benchmarking energy consumption and latency for neuromorphic computing," APL Machine Learning, 2023. DOI: 10.1063/5.0219604

[8] "Two-dimensional fully ferroelectric-gated hybrid computing-in-memory hardware," Science Advances, 2024. DOI: 10.1126/sciadv.adp0174 (0.24 fJ per operation)

[9] "HZO ferroelectric tunnel junction reservoir computing," ScienceDirect, 2025. DOI: 10.1016/j.jallcom.2025.034309 (98.24% MNIST accuracy)

[10] IEEE IRPS 2022 - FeFET endurance characteristics (10⁹ cycles demonstrated)

[11] "Vanadium-doped HfO₂ ferroelectric," Nano Letters, 2024. DOI: 10.1021/acs.nanolett.4c05671 (10¹² cycles extrapolated)

[12] CEA-Leti, "Embedded FeRAM Platform at 22nm FD-SOI," December 2024 (3D BEOL integration)

[13] "Cryogenic FeFET operation," Frontiers in Nanotechnology, 2024. DOI: 10.3389/fnano.2024.1371386 (5K-300K characterization)

[14] Fraunhofer IPMS, "FeFET AEC-Q100 Grade 0 Qualification," VLSI 2024 (automotive -40°C to 150°C)

### Additional Resources
- [HfO2-based ferroelectric: fundamentals and applications](https://www.nature.com/articles/s41578-022-00431-2) - Nature Reviews Materials
- [Enhancing ferroelectric stability in HfO2/ZrO2 superlattices](https://www.nature.com/articles/s41467-025-61758-2) - Nature Communications 2025
- [Rice Innovation: IronLattice Grant](https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants) - Verified IronLattice funding

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
