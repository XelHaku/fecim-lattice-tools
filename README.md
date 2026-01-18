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

## Implementation Status

| Demo | Physics | Graphics | Overall |
|------|---------|----------|---------|
| **Demo 1: Hysteresis** | Complete | Complete | **Vulkan visualization working** |
| **Demo 2: Crossbar MVM** | Complete | Complete | **Terminal visualization working** |
| **Demo 3: MNIST** | Complete | Complete | **Interactive classification working** |

**All demos run independently!**

```bash
# Demo 1: Vulkan P-E hysteresis visualization
cd demo1-hysteresis && go build -o hysteresis ./cmd/hysteresis && ./hysteresis

# Demo 2: Terminal crossbar MVM visualization
cd demo2-crossbar && go build -o inference ./cmd/inference && ./inference --show-mvm

# Demo 3: Interactive MNIST digit classification
cd demo3-mnist && go build -o mnist ./cmd/mnist && ./mnist --interactive
```

---

## Repository Structure

```
ironlattice-vis/
├── docs/                        # Comprehensive documentation (3.7 MB)
│   ├── CURRICULUM.md            # 8-area doctoral curriculum
│   ├── CURRICULUM_DETAILED.md   # Expanded learning path
│   ├── IRONLATTICE_PARADIGM.md  # Technology deep-dive
│   ├── PROJECT_ROADMAP.md       # Implementation timeline
│   ├── VULKAN_DEMO_GUIDE.md     # Graphics implementation guide
│   ├── HZO_PARAMETERS.md        # Material constants
│   ├── RESEARCH_LOG.md          # Research journal
│   └── RESEARCH_FINDINGS_*.md   # Weekly research summaries
│
├── papers/                      # Scientific papers collection
│   ├── downloaded/              # 19 PDFs (arXiv, Nature, IEEE, etc.)
│   ├── DOWNLOAD_PLAN.md         # Paper acquisition roadmap
│   ├── paper_metadata.json      # Paper index
│   └── paper_downloader.py      # Automated fetcher
│
├── demo1-hysteresis/            # Single cell P-E curve visualizer
│   ├── cmd/hysteresis/          # Application entry point
│   ├── pkg/
│   │   ├── ferroelectric/       # Preisach model, material params
│   │   ├── simulation/          # Time-stepping engine
│   │   └── render/              # Graphics pipeline (WIP)
│   ├── shaders/                 # GLSL compute/graphics shaders
│   ├── PHYSICS.md               # Physics documentation
│   └── README.md                # Demo-specific docs
│
├── demo2-crossbar/              # Crossbar array MVM visualizer
│   ├── cmd/inference/           # Application entry point
│   ├── pkg/
│   │   ├── crossbar/            # Array modeling
│   │   ├── network/             # Neural network layers
│   │   └── data/                # MNIST loading
│   ├── shaders/                 # MVM compute shaders
│   ├── PHYSICS.md               # Physics documentation
│   └── README.md                # Demo-specific docs
│
├── demo3-mnist/                 # MNIST neural network classifier
│   ├── cmd/mnist/               # Application entry point
│   ├── pkg/
│   │   ├── mnist/               # MNIST data loading
│   │   └── training/            # Neural network on crossbar
│   └── data/                    # Pretrained weights
│
└── go.mod                       # Go module definition
```

---

## Demos

### Demo 1: Ferroelectric Hysteresis Visualizer

**Status:** Complete with Vulkan visualization

Interactive visualization of a single ferroelectric memory cell with real-time P-E hysteresis curve:

```
┌────────────────┐      ┌──────────────────────┐      ┌───────────┐
│                │      │         P            │      │ ████ 30   │
│     CELL       │      │         ↑    +Pᵣ     │      │ ████ 29   │
│  (Color = P)   │      │         ┌────╮       │      │ ▓▓▓▓ ...  │
│                │      │    ─────┼────┼──→ E  │      │ ░░░░ 1    │
│                │      │         ╰────┘       │      │      0    │
│                │      │              -Pᵣ     │      │ 30 LEVELS │
└────────────────┘      └──────────────────────┘      └───────────┘
```

**Features:**
- Preisach hysteresis model with history tracking
- HZO material parameters from literature
- Real-time Vulkan GPU rendering
- 30 discrete analog state visualization
- Keyboard controls for E-field (UP/DOWN arrows)
- Multiple waveforms (sine, triangle, square)

**Build and Run:**
```bash
cd demo1-hysteresis
./shaders/compile.sh   # Compile SPIR-V shaders
go build -o hysteresis ./cmd/hysteresis
./hysteresis           # Vulkan window opens
```

**Controls:**
- **UP/DOWN arrows** - Adjust electric field
- **ESC** - Exit

### Demo 2: Crossbar Array MVM

**Status:** Complete with terminal visualization

Visualize Matrix-Vector Multiplication in memory using colorful terminal display:

```
    Input Vector (Voltages)
    ↓   ↓   ↓   ↓   ↓
V₁ ──→ [G₁₁][G₁₂][G₁₃] ──→ I₁ = Σ(Vⱼ × Gⱼ₁)    Output
V₂ ──→ [G₂₁][G₂₂][G₂₃] ──→ I₂ = Σ(Vⱼ × Gⱼ₂)    Currents
V₃ ──→ [G₃₁][G₃₂][G₃₃] ──→ I₃ = Σ(Vⱼ × Gⱼ₃)    (Outputs)

Ohm's Law:      I = V × G  (multiplication)
Kirchhoff's Law: Iₜₒₜₐₗ = ΣI (summation)
```

**Features:**
- Crossbar array with conductance visualization (block characters)
- Real-time MVM computation display
- Input/output vector visualization
- 30-level conductance states
- DAC/ADC quantization modeling
- Device noise simulation

**Build and Run:**
```bash
cd demo2-crossbar
go build -o inference ./cmd/inference
./inference --show-mvm      # Show MVM operation
./inference --show-array    # Show full crossbar state
```

**Options:**
- `--show-array` - Display crossbar conductance matrix
- `--show-mvm` - Visualize matrix-vector multiplication
- `--no-color` - Disable colored output

### Demo 3: MNIST Neural Network Classifier

**Status:** Complete with interactive mode

Neural network digit classification running on ferroelectric crossbar arrays:

```
    ┌─────────────────────────────────────────────┐
    │  Input: 28x28 = 784 pixels                  │
    │         ↓                                   │
    │  Layer 1: 784 → 128 (Crossbar Array #1)     │
    │         ↓ ReLU                              │
    │  Layer 2: 128 → 10 (Crossbar Array #2)      │
    │         ↓ Softmax                           │
    │  Output: 10 classes (digits 0-9)            │
    │  Target: 87% accuracy (IronLattice spec)    │
    └─────────────────────────────────────────────┘
```

**Features:**
- Interactive digit drawing (ASCII art input)
- Sample digit generation for testing
- Full softmax probability visualization
- Training mode with MNIST or synthetic data
- Weight quantization to 30 discrete levels
- Weight save/load for pretrained models

**Build and Run:**
```bash
cd demo3-mnist
go build -o mnist ./cmd/mnist
./mnist --interactive          # Interactive mode (default)
./mnist --train --epochs 10    # Training mode
./mnist --evaluate             # Evaluation mode
```

**Interactive Commands:**
- `sample N` - Classify sample digit N (0-9)
- `draw` - Enter custom digit drawing mode
- `test` - Run on random test samples
- `quit` - Exit

**Training Options:**
- `--train` - Train the network
- `--epochs N` - Number of training epochs
- `--hidden N` - Hidden layer size (default: 128)
- `--noise F` - Device noise level 0-1 (default: 0.02)
- `--save FILE` - Save trained weights
- `--load FILE` - Load pretrained weights

---

## Tech Stack

| Component | Technology | Purpose | Status |
|-----------|------------|---------|--------|
| Language | Go 1.21+ | Performance + simplicity | **Ready** |
| Graphics API | Vulkan 1.3 | Cross-platform GPU access | **Working** |
| Shaders | GLSL → SPIR-V | Compute + rendering | **Working** |
| Physics | Preisach model | Ferroelectric hysteresis | **Complete** |
| Neural Network | Crossbar MVM | MNIST classification | **Complete** |
| Simulation | TDGL | Domain dynamics | Planned |

### Dependencies

```go
github.com/bbredesen/go-vk  // Vulkan bindings
github.com/go-gl/glfw       // Window management
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- Vulkan SDK 1.3+ (for Demo 1 graphics)
- GLSL compiler `glslc` (for shader compilation)

### Quick Start

```bash
# Clone repository
git clone https://github.com/yourusername/ironlattice-vis.git
cd ironlattice-vis

# Install Go dependencies
go mod tidy
```

### Full Installation (Ubuntu/Debian)

```bash
# Install system dependencies
sudo apt install vulkan-tools libvulkan-dev glslc

# Compile shaders for Demo 1
cd demo1-hysteresis/shaders && ./compile.sh && cd ../..
```

### Running the Demos

**Demo 1: Hysteresis Visualization (Vulkan)**
```bash
cd demo1-hysteresis
go build -o hysteresis ./cmd/hysteresis
./hysteresis
```

**Demo 2: Crossbar MVM (Terminal)**
```bash
cd demo2-crossbar
go build -o inference ./cmd/inference
./inference --show-mvm
```

**Demo 3: MNIST Classifier (Interactive)**
```bash
cd demo3-mnist
go build -o mnist ./cmd/mnist
./mnist --interactive
```

### Headless Mode (No Graphics)

For systems without Vulkan, Demo 1 supports headless mode:
```bash
go run demo1-hysteresis/cmd/hysteresis/main.go --headless
```

---

## Learning Resources

### Documentation

| Document | Description |
|----------|-------------|
| [CURRICULUM.md](docs/CURRICULUM.md) | 8-area doctoral-level curriculum |
| [CURRICULUM_DETAILED.md](docs/CURRICULUM_DETAILED.md) | Expanded learning path |
| [IRONLATTICE_PARADIGM.md](docs/IRONLATTICE_PARADIGM.md) | Technology paradigm analysis |
| [PROJECT_ROADMAP.md](docs/PROJECT_ROADMAP.md) | Implementation timeline |
| [VULKAN_DEMO_GUIDE.md](docs/VULKAN_DEMO_GUIDE.md) | Graphics implementation guide |
| [HZO_PARAMETERS.md](docs/HZO_PARAMETERS.md) | Material constants reference |
| [papers/](papers/) | 19 scientific papers (arXiv, Nature, IEEE) |

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
- Dr. Tour's IronLattice Talk (Nov 2024) — Search "external research group IronLattice" on YouTube
- [external research institution News](https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants)

### Technical Papers
- Böscke, T.S., et al. "Ferroelectricity in hafnium oxide thin films." APL (2011)
- Park, M.H., et al. "Ferroelectricity in Doped HfO₂." Advanced Materials (2015)
- Shin, J., et al. "BEOL-Compatible Superlattice FEFET Analog Synapse" IEEE (2022)

### Dr. Tour's Ministry
- [Jesus and Science Foundation](https://jesusandscience.org)

---

## Contributing

Contributions welcome! Current priorities:

- [x] Preisach model implementation
- [x] Vulkan graphics pipeline for demo 1
- [x] 30-level discrete state visualization
- [x] MVM crossbar array simulation (demo 2)
- [x] Terminal visualization for crossbar
- [x] MNIST neural network on crossbar (demo 3)
- [x] Interactive digit classification
- [ ] Landau-Khalatnikov solver
- [ ] Phase-field simulation
- [ ] GPU-accelerated training

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
