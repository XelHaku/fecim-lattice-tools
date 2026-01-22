# Ferroelectric CIM Visualizer

**6 World-Class Modules for Ferroelectric Compute-in-Memory**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Fyne](https://img.shields.io/badge/Fyne-2.4-blue?logo=go)](https://fyne.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Modules](https://img.shields.io/badge/Modules-6%2F6-brightgreen.svg)]()

---

> **DISCLAIMER**: Ferroelectric CIM is at **TRL 4** (lab validation only). Hardware achieved **87% MNIST** (88% theoretical max). Energy claims (10M× vs NAND) are from Dr. Tour's presentation and have not been independently verified. See [HONESTY_AUDIT.md](docs/opensource/papers/08_Documentation/HONESTY_AUDIT.md).

---

## The FeCIM Story: 6 Core Modules

```
THE FeCIM NARRATIVE
═══════════════════════════════════════════════════════════════════════

"How does the          "How do we          "What can we
 memory cell work?"     compute with it?"   build with it?"
      ↓                      ↓                    ↓
┌──────────┐          ┌──────────┐          ┌──────────┐
│ MODULE 1 │    →     │ MODULE 2 │    →     │ MODULE 3 │
│Hysteresis│          │ Crossbar │          │  MNIST   │
│          │          │   +      │          │  87%     │
│30 levels │          │Non-Ideal │          │FLAGSHIP  │
└──────────┘          └──────────┘          └──────────┘
  PHYSICS              COMPUTE             APPLICATION

"How does it fit       "Why does FeCIM        "How do we
 in a real chip?"       beat everything?"      build chips?"
      ↓                      ↓                      ↓
┌──────────┐          ┌──────────┐          ┌──────────┐
│ MODULE 4 │    →     │ MODULE 5 │    →     │ MODULE 6 │
│ Circuits │          │Comparison│          │  Design  │
│  System  │          │ Investor │          │  Suite   │
│   CMOS   │          │  Pitch   │          │   EDA    │
└──────────┘          └──────────┘          └──────────┘
  SYSTEM               BUSINESS              TOOLING
```

```
Module 1: "How the memory cell works"           ✅ Fyne GUI - P-E Hysteresis
Module 2: "How we compute + handle challenges"  ✅ Fyne GUI - Crossbar MVM + Non-Idealities
Module 3: "What we can build with it"           ✅ Fyne GUI - MNIST 87% (FP vs CIM)
Module 4: "How it fits in a real chip"          ✅ Fyne GUI - Peripheral Circuits
Module 5: "Why FeCIM beats everything"          ✅ Fyne GUI - Technology Comparison
Module 6: "How do we build chips?"              ✅ Fyne GUI - FeCIM Design Suite (EDA)
```

---

## Quick Start

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Run the unified visualizer (ALL 6 MODULES)
./launch.sh

# Run all tests
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

## Module Details

### Module 1: Ferroelectric Hysteresis ✅

**Purpose:** Understand single cell physics — "The Memory Cell"

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

### Module 2: Crossbar MVM + Non-Idealities ✅ (4 TABS)

**Purpose:** Understand compute-in-memory AND real-world challenges — "The Crossbar Computer"

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

**Tab 1: Ideal MVM**
- Interactive heatmap with click-to-program cells
- Matrix-vector multiplication visualization
- Perfect operation baseline

**Tab 2: IR Drop Analysis**
- Wire resistance modeling
- Voltage gradient heatmap
- Worst-case corner identification
- Mitigation: 2x/4x wider metal lines

**Tab 3: Sneak Path Currents**
- Parasitic current visualization
- Target cell vs interference
- SNR degradation
- Mitigation: Selector devices (100:1, 1000:1)

**Tab 4: Drift & Variation**
- Conductance drift over time
- Read disturb effects
- FeCIM vs ReRAM vs PCM comparison
- 10-year retention prediction

---

### Module 3: MNIST Neural Network ✅ (FLAGSHIP)

**Purpose:** See real AI application — "The AI Brain"

> **Note:** Ferroelectric CIM hardware achieved **87%** with **88% theoretical max** (Dr. Tour). Our simulation includes dual-mode FP vs CIM comparison.

```
┌─────────┐    ┌─────────┐    ┌─────────┐
│ 28 × 28 │    │ 784×128 │    │ 128×10  │
│  INPUT  │ ─→ │ Layer 1 │ ─→ │ Layer 2 │ ─→ Prediction
│  DIGIT  │    │ Crossbar│    │ Crossbar│
└─────────┘    └─────────┘    └─────────┘
```

**Features:**
- Interactive 28×28 digit drawing canvas
- **Dual-mode inference:** Full Precision vs CIM comparison
- Hardware controls: Quantization levels, noise, ADC/DAC bits
- Failure mode presets (Ideal, Quant Cliff, Noisy, Broken ADC)
- Weight visualization heatmap (30 discrete colors)
- Guided Tour mode (7 steps)
- Energy efficiency display (10,000× savings)

---

### Module 4: Peripheral Circuits ✅

**Purpose:** Understand full chip system — "The Chip System"

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
- CMOS compatibility checklist

---

### Module 5: Technology Comparison ✅ (INVESTOR PITCH)

**Purpose:** The business case — "Why FeCIM Wins"

```
Energy per MAC (fJ)
═══════════════════════════════════════════
CPU+DRAM  ████████████████████████████ 1000
GPU+HBM   ██████████                    100
FeCIM     █                              10

Competitive Matrix (Only FeCIM has ✅ everywhere)
┌──────────┬──────┬──────┬──────┬──────┐
│ Feature  │FeCIM │ NAND │ReRAM │ PCM  │
├──────────┼──────┼──────┼──────┼──────┤
│ Energy   │  ✅  │  ❌  │  🟡  │  🟡  │
│ Speed    │  ✅  │  ❌  │  ✅  │  ❌  │
│ Endure   │  ✅  │  ❌  │  ❌  │  🟡  │
│ 30 lvls  │  ✅  │  ✅  │  ❌  │  ✅  │
│ CIM      │  ✅  │  ❌  │  🟡  │  🟡  │
└──────────┴──────┴──────┴──────┴──────┘
```

**Features:**
- Energy per MAC bar chart comparison
- Competitive technology matrix (FeCIM vs NAND vs ReRAM vs PCM vs MRAM)
- **Data center savings calculator** (input GPUs, see annual savings)
- Market opportunity ($403B by 2030)
- TRL progression roadmap (currently TRL 4)
- Verified vs claimed specifications with sources

---

### Module 6: FeCIM Design Suite ✅ (EDA TOOLING)

**Purpose:** Bridge from simulation to silicon — "The Chip Builder"

```
Neural Network Weights          Physical Crossbar Array

┌─────────────────┐             ┌─────────────────┐
│  0.5  -0.3  0.8 │   Compile   │ G₁₅  G₈   G₂₂  │
│ -0.2   0.6  0.1 │ ─────────→  │ G₁₁  G₁₈  G₅   │
│  0.9  -0.7  0.4 │             │ G₂₆  G₃   G₁₄  │
└─────────────────┘             └─────────────────┘
   Floating Point                 30-Level Cells
```

**Features:**
- **Compiler Tab:** Load weights, configure array, compile to cells
- **Layout Tab:** Visual crossbar grid (color-coded by conductance)
- **Export Tab:** JSON, CSV, and SPICE netlist generation
- Quantization to 30 FeCIM levels
- PSNR and utilization statistics
- CLI tool for automated/headless compilation

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| GUI | Fyne 2.4 |
| Physics | Preisach/Mayergoyz model |
| Neural Network | Crossbar MVM simulation |
| Non-Idealities | IR drop, sneak paths, drift |

---

## Repository Structure

```
multilayer-ferroelectric-cim-visualizer/
├── cmd/
│   └── fecim-visualizer/  ✅ Unified GUI (ALL 6 MODULES)
├── demo1-hysteresis/      ✅ Single cell P-E curve
├── demo2-crossbar/        ✅ Crossbar MVM + Non-Idealities (4 tabs)
├── demo3-mnist/           ✅ MNIST classifier (FP vs CIM)
├── demo4-circuits/        ✅ Peripheral circuits
├── demo6-eda/             ✅ FeCIM Design Suite (EDA tooling)
├── demo8-comparison/      ✅ Technology comparison (Module 5)
├── shared/                Shared packages (theme, logging)
├── docs/
│   └── archive/           Archived demos
└── go.mod
```

### Archived Modules

The following modules were consolidated or archived during restructuring:
- **Thermal**: Content merged into comparison
- **3D Stack**: Archived — too futuristic for current TRL
- **Non-Idealities**: Merged into Module 2 as tabs

See `docs/archive/removed-demos/README.md` for details.

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

*6 world-class demos. The future of computing is here.*

*Built with Go, Fyne, and curiosity.*
