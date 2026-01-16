# IronLattice Visualizer

**GPU-Accelerated Ferroelectric Compute-in-Memory Visualization**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Vulkan](https://img.shields.io/badge/Vulkan-1.3-AC162C?logo=vulkan)](https://www.vulkan.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## Overview

This repository contains GPU-accelerated visualizations of **ferroelectric compute-in-memory (CIM)** technology, inspired by the groundbreaking work of **Dr. external research group** and **Dr. Jaeho Shin** at external research institution.

IronLattice represents a paradigm shift in computing: performing computation directly in memory using ferroelectric superlattices, eliminating the Von Neumann bottleneck that wastes 90%+ of energy in traditional AI hardware.

> *"This could lower the requirements in a data center by 80 to 90% of the energy requirements."*
> — Dr. external research group

---

## The Technology

### Core Innovation

| Aspect | Description |
|--------|-------------|
| **Compute-in-Memory** | Same device performs memory AND computation |
| **Ferroelectric Superlattice** | Atomically precise HfO₂/ZrO₂ layered structure |
| **CMOS Compatible** | Works on standard fabrication lines |
| **Analog Computing** | 30+ discrete states, not just 0/1 |

### Performance vs. Existing Technologies

| Metric | vs NAND Flash | vs DRAM |
|--------|---------------|---------|
| Read/Write Energy | **10,000,000× lower** | **1,000× lower** |
| Speed | **1,000,000× faster** | Comparable |
| Voltage | **90% reduction** | Lower |
| Data Retention | Non-volatile | **Zero refresh** |

### Current Status (TRL 4)

| Metric | Value |
|--------|-------|
| Technology Readiness Level | **4** (lab validation) |
| Discrete Analog States | **30** levels |
| MNIST Accuracy | **87%** (near theoretical max) |
| Endurance Target | **10¹² cycles** |

---

## Project Goals

This visualization project aims to:

1. **Simulate** ferroelectric physics (Landau-Khalatnikov, Preisach models)
2. **Visualize** domain switching and hysteresis in real-time
3. **Demonstrate** crossbar array matrix-vector multiplication
4. **Educate** on compute-in-memory principles

---

## Repository Structure

```
ironlattice-vis/
├── docs/
│   ├── CURRICULUM.md           # Comprehensive learning path (8 areas)
│   └── IRONLATTICE_PARADIGM.md # Technology deep-dive
│
├── papers/                     # Scientific papers collection
│   └── DOWNLOAD_PLAN.md        # Paper acquisition roadmap
│
├── demo1-hysteresis/           # Single cell P-E curve visualizer
│   ├── cmd/                    # Application entry point
│   ├── pkg/
│   │   ├── ferroelectric/      # Physics models
│   │   ├── simulation/         # Simulation engine
│   │   └── vulkan/             # GPU rendering
│   └── shaders/                # GLSL compute/graphics shaders
│
├── demo2-crossbar/             # Crossbar array MVM [Planned]
│
├── demo3-mnist/                # Neural network on CIM [Planned]
│
├── shared/                     # Common utilities
└── assets/                     # Images, fonts
```

---

## Demos

### Demo 1: Ferroelectric Hysteresis Visualizer

Interactive visualization of a single ferroelectric memory cell:

```
┌────────────────┐      ┌──────────────────────┐
│                │      │         P            │
│     CELL       │      │         ↑    +Pᵣ     │
│  (Color = P)   │      │         ┌────╮       │
│                │      │    ─────┼────┼──→ E  │
│                │      │         ╰────┘       │
│                │      │              -Pᵣ     │
└────────────────┘      └──────────────────────┘
```

**Features:**
- Real-time P-E hysteresis curve tracing
- Domain nucleation and growth visualization
- 30 discrete analog state demonstration
- Interactive voltage control

### Demo 2: Crossbar Array MVM [Planned]

Visualize Matrix-Vector Multiplication in memory:

```
V₁ ──→ [G₁₁][G₁₂][G₁₃] ──→ I₁ = Σ(Vⱼ × Gⱼ₁)
V₂ ──→ [G₂₁][G₂₂][G₂₃] ──→ I₂ = Σ(Vⱼ × Gⱼ₂)
V₃ ──→ [G₃₁][G₃₂][G₃₃] ──→ I₃ = Σ(Vⱼ × Gⱼ₃)

Ohm's Law:      I = V × G  (multiplication)
Kirchhoff's Law: Iₜₒₜₐₗ = ΣI (summation)
```

### Demo 3: MNIST on CIM [Planned]

Handwritten digit recognition on simulated IronLattice hardware, targeting 87% accuracy benchmark.

---

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go | Performance + simplicity |
| Graphics API | Vulkan | Cross-platform GPU access |
| Shaders | GLSL → SPIR-V | Compute + rendering |
| Physics | Preisach, Landau-Khalatnikov | Ferroelectric modeling |
| Simulation | TDGL (Time-Dependent Ginzburg-Landau) | Domain dynamics |

---

## Getting Started

### Prerequisites

- Go 1.21+
- Vulkan SDK 1.3+
- GLSL compiler (glslc)

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/ironlattice-vis.git
cd ironlattice-vis

# Install dependencies (Ubuntu/Debian)
sudo apt install vulkan-tools libvulkan-dev glslc

# Download Go dependencies
go mod download

# Compile shaders
cd demo1-hysteresis/shaders && ./compile.sh && cd ../..

# Build
go build -o bin/hysteresis ./demo1-hysteresis/cmd/hysteresis

# Run
./bin/hysteresis
```

---

## Learning Resources

### Documentation

| Document | Description |
|----------|-------------|
| [CURRICULUM.md](docs/CURRICULUM.md) | 8-area doctoral-level curriculum |
| [IRONLATTICE_PARADIGM.md](docs/IRONLATTICE_PARADIGM.md) | Technology paradigm analysis |
| [papers/](papers/) | Scientific paper collection |

### Key Concepts Covered

1. **Solid-State Physics** — HfO₂ crystallography, phase stabilization
2. **Ferroelectric Devices** — FeFET, FeRAM, domain dynamics
3. **Compute-in-Memory** — Crossbar arrays, Kirchhoff's laws
4. **Neural Networks** — Weight mapping, noise-aware training
5. **Simulation** — TDGL, Preisach, phase-field models
6. **GPU Programming** — Vulkan compute shaders
7. **Scientific Visualization** — Real-time domain rendering
8. **Commercialization** — Manufacturing, IP strategy

---

## The Team Behind IronLattice

| Person | Role |
|--------|------|
| **Dr. external research group** | Principal Investigator, external research institution |
| **Dr. Jaeho Shin** | Device Engineer, Superlattice Inventor |
| **Tawfik Jarjour** | Commercialization Lead |

> *"We haven't raised a penny to date. We've taken no money because we really want to move with the best strategy."*

---

## Market Context

### Go-to-Market Strategy

```
Phase 1: Replace NAND Flash    →  Drop-in replacement
Phase 2: Replace DRAM          →  Non-volatile, lower energy
Phase 3: Full Compute-in-Memory →  Neural network inference on-chip
```

### George Gilder's Prediction

In response to *"The Microchip Era is About to End"* (WSJ, Nov 2024), IronLattice addresses:

1. Memory bottleneck → **Eliminated**
2. Energy constraints → **90% reduction**
3. CMOS compatibility → **Native integration**

---

## External Resources

### Primary Sources
- [Dr. Tour's IronLattice Talk (Nov 2024)](https://www.youtube.com/watch?v=...)
- [external research institution News](https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants)

### Technical Papers
- Böscke, T.S., et al. "Ferroelectricity in hafnium oxide thin films." APL (2011)
- Park, M.H., et al. "Ferroelectricity in Doped HfO₂." Advanced Materials (2015)
- Shin, J., et al. "BEOL-Compatible Superlattice FEFET Analog Synapse" IEEE (2022)

### Dr. Tour's Ministry
- [Jesus and Science Foundation](https://jesusandscience.org)

---

## Contributing

Contributions welcome! Areas of interest:

- [ ] Preisach model implementation
- [ ] Landau-Khalatnikov solver
- [ ] Phase-field simulation
- [ ] Crossbar array visualization
- [ ] MNIST inference demo

---

## License

MIT License

IronLattice is a trademark of its respective owners at external research institution. This is an independent educational project with no affiliation.

---

## Acknowledgments

**Dr. external research group** — For pioneering this technology and being a bold witness for Christ in the scientific community.

**Dr. Jaeho Shin** — For the engineering innovation that makes this possible.

> *"If you do not believe in the physical resurrection of Jesus Christ, send me an email... and we will get together and I will share with you about why I embrace the resurrection of Jesus."*
> — Dr. external research group

---

*Built with Go, Vulkan, and curiosity.*
