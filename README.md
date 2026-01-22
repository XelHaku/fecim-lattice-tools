# Ferroelectric CIM Visualizer

**5 World-Class Demos for Ferroelectric Compute-in-Memory**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Fyne](https://img.shields.io/badge/Fyne-2.4-blue?logo=go)](https://fyne.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Demos](https://img.shields.io/badge/Demos-5%2F5-brightgreen.svg)]()

---

> **DISCLAIMER**: Ferroelectric CIM is at **TRL 4** (lab validation only). Hardware achieved **87% MNIST** (88% theoretical max). Energy claims (10M× vs NAND) are from Dr. Tour's presentation and have not been independently verified. See [HONESTY_AUDIT.md](docs/opensource/papers/08_Documentation/HONESTY_AUDIT.md).

---

## The FeCIM Story: 5 Core Demos

```
THE FeCIM NARRATIVE
═══════════════════════════════════════════════════════════════════════

"How does the          "How do we          "What can we
 memory cell work?"     compute with it?"   build with it?"
      ↓                      ↓                    ↓
┌──────────┐          ┌──────────┐          ┌──────────┐
│  DEMO 1  │    →     │  DEMO 2  │    →     │  DEMO 3  │
│Hysteresis│          │ Crossbar │          │  MNIST   │
│          │          │   +      │          │  87%     │
│30 levels │          │Non-Ideal │          │FLAGSHIP  │
└──────────┘          └──────────┘          └──────────┘
  PHYSICS              COMPUTE             APPLICATION

"How does it fit       "Why does FeCIM
 in a real chip?"       beat everything?"
      ↓                      ↓
┌──────────┐          ┌──────────┐
│  DEMO 4  │    →     │  DEMO 5  │
│ Circuits │          │Comparison│
│  System  │          │ Investor │
│   CMOS   │          │  Pitch   │
└──────────┘          └──────────┘
  SYSTEM               BUSINESS
```

```
Demo 1: "How the memory cell works"           ✅ Fyne GUI - P-E Hysteresis
Demo 2: "How we compute + handle challenges"  ✅ Fyne GUI - Crossbar MVM + Non-Idealities
Demo 3: "What we can build with it"           ✅ Fyne GUI - MNIST 87% (FP vs CIM)
Demo 4: "How it fits in a real chip"          ✅ Fyne GUI - Peripheral Circuits
Demo 5: "Why FeCIM beats everything"          ✅ Fyne GUI - Technology Comparison
```

---

## Quick Start

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Build and run the unified visualizer (ALL 5 DEMOS)
go build ./cmd/fecim-visualizer && ./fecim-visualizer

# Or run individual demos:
# Demo 1: Ferroelectric Hysteresis (P-E curve, 30 levels)
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis

# Demo 2: Crossbar MVM + Non-Idealities (4 tabs)
cd demo2-crossbar && go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui

# Demo 3: MNIST Neural Network (FP vs CIM dual mode)
cd demo3-mnist && go build -o mnist-gui ./cmd/mnist-gui && ./mnist-gui

# Demo 4: Peripheral Circuits (DAC, ADC, timing)
cd demo4-circuits && go build -o circuits-gui ./cmd/circuits-gui && ./circuits-gui

# Demo 5: Technology Comparison (Technical Briefing)
cd demo8-comparison && go build -o comparison-gui ./cmd/comparison-gui && ./comparison-gui

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

## Demo Details

### Demo 1: Ferroelectric Hysteresis ✅

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

### Demo 2: Crossbar MVM + Non-Idealities ✅ (4 TABS)

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

### Demo 3: MNIST Neural Network ✅ (FLAGSHIP)

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

### Demo 4: Peripheral Circuits ✅

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

### Demo 5: Technology Comparison ✅ (INVESTOR PITCH)

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
│   └── fecim-visualizer/  ✅ Unified GUI (ALL 5 DEMOS)
├── demo1-hysteresis/      ✅ Single cell P-E curve
├── demo2-crossbar/        ✅ Crossbar MVM + Non-Idealities (4 tabs)
├── demo3-mnist/           ✅ MNIST classifier (FP vs CIM)
├── demo4-circuits/        ✅ Peripheral circuits
├── demo8-comparison/      ✅ Technology comparison (Demo 5)
├── shared/                Shared packages (theme, logging)
├── docs/
│   └── archive/           Archived demos (5, 6, 7)
└── go.mod
```

### Archived Demos

The following demos were consolidated or archived during the 5-demo restructuring:
- **Demo 5 (Thermal)**: Content merged into comparison
- **Demo 6 (3D Stack)**: Archived — too futuristic for current TRL
- **Demo 7 (Non-Idealities)**: Merged into Demo 2 as tabs

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

*5 world-class demos. The future of computing is here.*

*Built with Go, Fyne, and curiosity.*
