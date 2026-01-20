# IronLattice-vis Playbook

**MISSION:** Create world-class visualization demos to help Dr. external research group pitch IronLattice to investors, engineers, and foundry partners.

**PRIMARY REFERENCE:** `ironlattice-transcript.md` (Dr. Tour's Nov 2024 presentation)
**TASK TRACKING:** `TODO.md` (authoritative task list)
**PAPERS:** `opensource/papers/08_Documentation/PAPERS_NEEDED.md`

---

## CURRENT STATUS (2026-01-19)

```
THE IRONLATTICE STORY - 8 DEMOS

Demo 1        Demo 2        Demo 3        Demo 4
"How the      "How we       "What we      "How it fits
memory        compute       can build     in a real
cell works"   in memory"    with it"      chip"
    ↓             ↓             ↓             ↓
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│ P-E     │   │Crossbar │   │  MNIST  │   │Peripheral│
│Hysteresis│   │   MVM   │   │  (sim)  │   │ Circuits │
└─────────┘   └─────────┘   └─────────┘   └─────────┘
  ✅ FYNE      ✅ FYNE       ✅ FYNE       ✅ CLI

Demo 5        Demo 6        Demo 7        Demo 8
"1000×        "Scalable     "Real-world   "Why IL
cooler"       3D stack"     challenges"   wins"
    ↓             ↓             ↓             ↓
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│ Thermal │   │Multi-   │   │ Non-    │   │Comparison│
│   Map   │   │ Layer   │   │idealities│   │  Chart   │
└─────────┘   └─────────┘   └─────────┘   └─────────┘
  ✅ CLI       🔲 TODO       ✅ in Demo2   🔲 PRIORITY
```

---

## QUICK START - Run All GUIs

```bash
# Demo 1: Ferroelectric Hysteresis (P-E curve, 30 levels)
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis

# Demo 2: Crossbar MVM (IR drop, sneak paths, heatmaps)
cd demo2-crossbar && go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui

# Demo 3: MNIST Neural Network (draw digits, simulation)
# NOTE: IronLattice hardware = 87%, theoretical max = 88%
cd demo3-mnist && go build -o mnist-gui ./cmd/mnist-gui && ./mnist-gui

# Demo 4: Peripheral Circuits (CLI - linearity, timing, power)
cd demo4-circuits && go run ./cmd/circuits --all

# Demo 5: Thermal Simulation (CLI - heat maps)
cd demo5-thermal && go run ./cmd/thermal --realtime

# Run all tests
go test ./...
```

**Build Dependencies (Fyne GUI):**
```bash
# Ubuntu/Debian
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel
```

---

## TARGET AUDIENCES

| Audience | What They Care About | Key Demos |
|----------|---------------------|-----------|
| **Investors** | ROI, $711B market, 80-90% energy savings | Demo 3, 8 |
| **Engineers** | Physics accuracy, real-world issues | Demo 1, 2, 5, 7 |
| **Foundries** | CMOS compatibility, process flow | Demo 4 |
| **Strategic Partners** | Competitive advantage | Demo 8 |

---

## Demo 1: Hysteresis (Memory Cell Physics) ✅ FYNE GUI

**Story:** "This is how the memory cell works"

**Implemented Features:**
- Mayergoyz Preisach model with full hysteron distribution
- 30 discrete levels clearly shown (LevelIndicator)
- Temperature dependence modeling
- Thread-safe simulation engine
- Real-time P-E hysteresis curve with fade trail
- Material selector (Default HZO, Optimized, IronLattice)
- Waveform selector (Sine, Triangle, Square, Manual)

**Run:**
```bash
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis
```

**GUI Package:** `demo1-hysteresis/pkg/gui/`
- Custom widgets: `PEPlot`, `LevelIndicator`

**Tests:** 25 passing

---

## Demo 2: Crossbar MVM (Compute-in-Memory) ✅ FYNE GUI

**Story:** "This is how we compute in memory"

**Implemented Features:**
- IR drop analysis with wire resistance modeling
- Sneak path current analysis with visualization
- Interactive heatmap with click-to-select cells
- Three tabbed views: Conductance, IR Drop, Sneak Paths
- 30 discrete conductance levels
- Custom "IronLattice" colormap

**Run:**
```bash
cd demo2-crossbar && go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui
```

**GUI Package:** `demo2-crossbar/pkg/gui/`
- Custom widgets: `CrossbarHeatmap`, `VectorBarChart`, `DiscreteLevel30Indicator`

**Tests:** 14 passing

---

## Demo 3: MNIST (Neural Network) ✅ FYNE GUI

**Story:** "This is what we can build with it"

> ⚠️ **HARDWARE vs SIMULATION:** IronLattice hardware achieved **87%** with **88% theoretical maximum** (per Dr. Tour). Our simulation uses idealized conditions and may exceed real hardware capabilities.

**Implemented Features:**
- Simulation accuracy varies (idealized, may exceed hardware)
- Interactive 28x28 digit drawing canvas
- Real-time inference as you draw
- Layer activation visualization (input → hidden → output)
- Confusion matrix with clickable cells
- Per-class metrics (precision, recall, F1)
- 30 discrete weight levels

**Run:**
```bash
cd demo3-mnist && go build -o mnist-gui ./cmd/mnist-gui && ./mnist-gui
```

**GUI Package:** `demo3-mnist/pkg/gui/`
- Custom widgets: `DigitCanvas`, `LayerActivationView`, `ConfusionMatrix`, `MetricsPanel`

**Tests:** 9 passing

---

## Demo 4: Peripheral Circuits (System Integration) ✅ CLI

**Story:** "This is how it fits in a real chip"

**Implemented Features:**
- DAC: Digital → Write voltage (5-bit, 30 levels)
- ADC: Analog → Digital level (5-bit)
- TIA: Transimpedance Amplifier
- Charge Pump: 1V → ±1.5V
- INL/DNL linearity analysis
- Timing diagrams
- Power breakdown

**Run:**
```bash
cd demo4-circuits && go run ./cmd/circuits --all
cd demo4-circuits && go run ./cmd/circuits --linearity
cd demo4-circuits && go run ./cmd/circuits --timing
cd demo4-circuits && go run ./cmd/circuits --power
```

**Tests:** 9 passing

---

## Demo 5: Thermal Simulation ✅ CLI

**Story:** "1000× cooler than competition"

**Implemented Features:**
- 2D heat map visualization
- Real-time heat diffusion
- Multi-layer heat coupling
- Hotspot identification
- Thermal throttling warnings
- IronLattice's low-power advantage

**Run:**
```bash
cd demo5-thermal && go run ./cmd/thermal --realtime
```

**Tests:** 17 passing

---

## Demo 8: Technology Comparison 🔲 PRIORITY

**Story:** "Why IronLattice wins vs everyone else"

**Purpose:** The slide Dr. Tour shows investors

```
┌──────────────────────────────────────────────────────────────────┐
│                    COMPUTE PERFORMANCE COMPARISON                 │
├────────────────┬─────────────┬─────────────┬─────────────────────┤
│    Metric      │  DRAM+CPU   │    GPU      │    IronLattice      │
├────────────────┼─────────────┼─────────────┼─────────────────────┤
│ Energy vs NAND │     1×      │    0.1×     │   0.0000001× (10M×) │
│ Speed vs NAND  │     1×      │    100×     │   1,000,000×        │
│ Data Movement  │   O(n²)     │   O(n²)     │        0            │
│ Memory Refresh │   Required  │   Required  │       None          │
│ CMOS Compatible│     Yes     │    Yes      │       Yes           │
│ 30 Analog States│    No      │    No       │       Yes           │
└────────────────┴─────────────┴─────────────┴─────────────────────┘
```

**To Implement:**
- [ ] Side-by-side animated comparison
- [ ] Energy meter visualization (10M× difference)
- [ ] Data center savings calculator
- [ ] Competitive matrix from Dr. Tour's slides

---

## IRONLATTICE SPECS (From Dr. Tour)

| Spec | IronLattice Hardware | Our Simulation | Verification |
|------|---------------------|----------------|--------------|
| Analog states | **30 levels** | ✅ 30 levels | VERIFIED |
| MNIST accuracy | **87%** (88% max) | Variable | ⚠️ SIM ONLY |
| Energy vs NAND | 10M× lower | N/A | UNVERIFIED |
| Energy vs DRAM | 1000× lower | N/A | UNVERIFIED |
| Speed vs NAND | 1M× faster | N/A | UNVERIFIED |
| Data center savings | **80-90%** | N/A | UNVERIFIED |
| CMOS compatible | Standard fab | ✅ Modeled | VERIFIED |
| TRL | **4 (lab only)** | — | VERIFIED |

> ⚠️ Energy claims are from Dr. Tour's presentation and have not been independently verified. IronLattice is at TRL 4 (lab validation), not production.

---

## DR. TOUR'S PHASED MARKET ENTRY

```
PHASE 1               PHASE 2               PHASE 3
┌─────────────┐      ┌─────────────┐      ┌─────────────────┐
│  Replace    │  →   │  Replace    │  →   │  Full Compute-  │
│  NAND Flash │      │  DRAM       │      │  in-Memory      │
└─────────────┘      └─────────────┘      └─────────────────┘
  Easy entry           No refresh           80-90% energy
  No SW changes        1000× lower E        savings
```

---

## PAPER LIBRARY STATUS

**VALID (40+ papers):** See `papers/downloaded/` and `opensource/papers/`

**CORRUPTED (need IEEE access):**
- `Mayergoyz_IEEE_1986.pdf` - Preisach model (CRITICAL)
- `IEEE_CIM_Survey_2023.pdf` - CIM overview
- `Tour_In2Se3_ChemRxiv.pdf` - 2D ferroelectrics

**Full list:** `opensource/papers/08_Documentation/PAPERS_NEEDED.md`

---

## ALL TESTS

```bash
go test ./...   # 110+ tests passing
```

| Package | Tests |
|---------|-------|
| ferroelectric | 20 |
| simulation | 5 |
| crossbar | 14 |
| training (mnist) | 9 |
| peripherals | 9 |
| thermal | 17 |
| multilayer | 17 |
| nonidealities | 20 |
| comparison | 19 |

---

## FILE STRUCTURE

```
ironlattice-vis/
├── demo1-hysteresis/     ✅ P-E curve + Fyne GUI
├── demo2-crossbar/       ✅ Crossbar MVM + Fyne GUI
├── demo3-mnist/          ✅ MNIST (simulation) + Fyne GUI
├── demo4-circuits/       ✅ Peripherals (CLI)
├── demo5-thermal/        ✅ Thermal sim (CLI)
├── demo6-multilayer/     🔲 3D multi-layer
├── demo7-nonidealities/  ✅ (integrated in demo2)
├── demo8-comparison/     🔲 Technology comparison
├── papers/               Scientific papers
├── opensource/papers/    Additional papers + PAPERS_NEEDED.md
├── PLAYBOOK.md           This file (project handbook)
├── TODO.md               Strategic task list
└── ironlattice-transcript.md  Dr. Tour's presentation
```

---

## NEXT PRIORITIES

1. **Demo 2, 3, 4: Upgrade to Live Presentation Slides** (see below)
2. **Demo 8: Technology Comparison** - Investor pitch slide
3. **Demo 5 Fyne GUI** - Thermal heat map
4. **Web deployment** - For remote investor presentations
5. **Pitch video** - 2-min demo reel

---

## LIVE PRESENTATION SLIDE PATTERN (Learn from Demo 1)

Demo 1 established the gold standard for "live presentation slides" — interactive visualizations that tell a story while the audience watches. **Apply this pattern to Demos 2, 3, and 4.**

### What Makes Demo 1 Special

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Demo 1: The Complete "Live Slide" Pattern                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────┐ ┌────────────────────┐ ┌────┐ ┌───────────┬───────────────────┐ │
│  │ VISUAL  │ │   MAIN PLOT        │ │LEVEL│ │ CONTROLS  │ EDUCATIONAL       │ │
│  │ ANCHOR  │ │   (the science)    │ │ BAR │ │           │ PANEL             │ │
│  │         │ │                    │ │     │ │ Material  │                   │ │
│  │ ┌─────┐ │ │  P-E hysteresis    │ │ 30  │ │ Waveform  │ "What You're      │ │
│  │ │ 24  │ │ │  with fade trail   │ │ ▓▓  │ │ Frequency │  Seeing"          │ │
│  │ └─────┘ │ │                    │ │ ▓▓  │ │           │                   │ │
│  │         │ │  Ec/Pr markers     │ │ ░░  │ │ [Pause]   │ Context-sensitive │ │
│  │ Memory  │ │                    │ │ ░░  │ │ [Reset]   │ explanations      │ │
│  │  Cell   │ └────────────────────┘ │     │ ├───────────┼───────────────────┤ │
│  └─────────┘                        └────┘ │ Status    │ MEMORY LOG        │ │
│                                            │ E: 0.85   │                   │ │
│                                            │ P: 25.3   │ >> WRITE(28)      │ │
│  ● Status bar: mode + current action       │ [WRITE]   │    Got: 27 [OK]   │ │
│                                            └───────────┴───────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────┘
```

### The 6 Essential Components

| Component | Purpose | Demo 1 Example |
|-----------|---------|----------------|
| 1. **Visual Anchor** | Hero element that grabs attention | Memory cell with level number |
| 2. **Main Plot** | The science in action | P-E hysteresis with fade trail |
| 3. **Level Indicator** | Show the 30-level advantage | Vertical bar with 30 segments |
| 4. **Controls** | Let presenter/audience interact | Material, waveform, frequency |
| 5. **Educational Panel** | Explain what's happening NOW | "WRITE: E>Ec sets state" |
| 6. **Operation Log** | Show live operations | ">> WRITE(28) ... Got: 27 [OK]" |

### Demo Modes (from Demo 1)

Each demo should have multiple **waveform/demo modes**:

| Mode Type | Purpose | Example |
|-----------|---------|---------|
| **Auto Mode** | Self-running for passive viewing | Sine wave traces full loop |
| **Demo Mode** | Step-by-step educational | Write/Read Demo shows 4 phases |
| **Manual Mode** | Full control for Q&A | Slider control |
| **Random Mode** | Show versatility | Random Walk picks random levels |

---

## DEMO 2 UPGRADE PLAN: Crossbar MVM as Live Slide

**Story:** "This is how we compute in memory"

### Current State
- ✅ Heat map visualization
- ✅ IR drop / sneak path views
- ✅ Basic controls

### Upgrade to Live Slide

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Demo 2: Crossbar MVM - "How We Compute in Memory"                           │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────┐ ┌────┐ ┌───────────┬───────────────────┐│
│  │     CROSSBAR HEAT MAP           │ │ 30 │ │ CONTROLS  │ WHAT YOU'RE       ││
│  │     (the visual anchor)         │ │ L  │ │           │ SEEING            ││
│  │                                 │ │ E  │ │ Array:8×8 │                   ││
│  │   V₀  V₁  V₂  V₃                │ │ V  │ │ Demo: MVM │ MVM DEMO          ││
│  │   ↓   ↓   ↓   ↓                 │ │ E  │ │           │                   ││
│  │  ┌───┬───┬───┬───┐              │ │ L  │ │ [Run MVM] │ 1. Input voltages ││
│  │  │▓▓▓│░░░│▓░▓│░▓░│→ I₀          │ │ S  │ │ [IR Drop] │    applied to     ││
│  │  ├───┼───┼───┼───┤              │ │    │ │ [Sneak]   │    columns        ││
│  │  │░▓░│▓▓▓│░░░│▓▓░│→ I₁          │ │ ▓▓ │ ├───────────┤                   ││
│  │  └───┴───┴───┴───┘              │ │ ▓░ │ │ INPUT V   │ 2. Current flows  ││
│  │                                 │ │ ░░ │ │ [1.0][0.5]│    through ALL    ││
│  │  Selected: (2,1) Level 18       │ │    │ │ [0.8][0.3]│    cells at once  ││
│  │  G = 45.2 µS                    │ └────┘ ├───────────┤                   ││
│  └─────────────────────────────────┘        │ OUTPUT I  │ 3. Row currents   ││
│                                             │ I₀: 2.34  │    = dot product! ││
│  ┌─────────────────────────────────────────┐│ I₁: 1.87  │                   ││
│  │  OPERATION LOG                          │├───────────┼───────────────────┤│
│  │  >> MVM: V=[1.0,0.5,0.8,0.3]           ││ PHYSICS   │ EFFICIENCY        ││
│  │     Computing I = G × V ...            ││ I = G × V │                   ││
│  │     Output: I=[2.34, 1.87] ✓           ││ (Ohm's    │ GPU: ~1000 fJ/op  ││
│  │  >> IR Drop analysis...                ││  Law!)    │ IL:  ~10 fJ/op    ││
│  │     Max drop: 3.2% at (7,7)            ││           │ = 100× better!    ││
│  └─────────────────────────────────────────┘└───────────┴───────────────────┘│
│  ● MVM Demo | Computing: I = G × V (4×4 = 16 multiplications in parallel)    │
└──────────────────────────────────────────────────────────────────────────────┘
```

### New Demo Modes for Demo 2

| Mode | What It Shows |
|------|---------------|
| **MVM Demo** | Animate voltage application → current collection |
| **IR Drop Demo** | Pulse travels across array, show voltage attenuation |
| **Sneak Path Demo** | Highlight unwanted current paths in red |
| **Write Weight Demo** | Program a cell, show conductance change |
| **Comparison Demo** | Side-by-side: ideal vs non-ideal results |

### Key Additions Needed

1. **Input Vector Display** - Show V values being applied
2. **Output Vector Display** - Show I results with animation
3. **Educational Panel** - "1. Voltages applied to columns..."
4. **Operation Log** - ">> MVM: V=[...] → I=[...]"
5. **Physics Callout** - "I = G × V (Ohm's Law!)"
6. **Efficiency Comparison** - "GPU: 1000 fJ, IL: 10 fJ"

---

## DEMO 3 UPGRADE PLAN: MNIST as Live Slide

**Story:** "This is what we can build with it"

### Current State
- ✅ Digit drawing canvas
- ✅ Classification output
- ✅ Confusion matrix

### Upgrade to Live Slide

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Demo 3: MNIST Neural Network - "What We Can Build"                          │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────┐ ┌──────────────────────┐ ┌───────────────────────────────────┐│
│  │  DRAW     │ │  LAYER ACTIVATIONS   │ │ WHAT YOU'RE SEEING               ││
│  │  HERE     │ │                      │ │                                   ││
│  │           │ │  Input    Hidden Out │ │ INFERENCE DEMO                    ││
│  │  ┌─────┐  │ │  784  →  128  →  10  │ │                                   ││
│  │  │     │  │ │  ┌──┐    ┌──┐   ┌──┐ │ │ 1. Your drawing becomes          ││
│  │  │  7  │  │ │  │▓▓│    │░▓│   │  │ │ │    784 pixel values              ││
│  │  │     │  │ │  │▓░│    │▓▓│   │▓▓│ │ │                                   ││
│  │  └─────┘  │ │  │░░│    │░▓│   │░░│ │ │ 2. MVM #1: 784×128 = 100K        ││
│  │           │ │  └──┘    └──┘   └──┘ │ │    multiplications (instant!)    ││
│  │  28×28    │ │                      │ │                                   ││
│  └───────────┘ └──────────────────────┘ │ 3. MVM #2: 128×10 = 1,280        ││
│                                         │    multiplications (instant!)    ││
│  ┌─────────────────────────────────────┐│                                   ││
│  │  PREDICTION CONFIDENCE              ││ 4. Highest score wins:           ││
│  │  0: ░░░░░░░░░░  2%                 ││    "That's a 7!"                 ││
│  │  1: ░░░░░░░░░░  1%                 │├───────────────────────────────────┤│
│  │  ...                               ││ HARDWARE ACCURACY                 ││
│  │  7: ████████████████████  95% ←WIN ││                                   ││
│  │  8: ░░░░░░░░░░  2%                 ││ IronLattice: 87%                  ││
│  │  9: ░░░░░░░░░░  0%                 ││ Theoretical: 88%                  ││
│  └─────────────────────────────────────┘│ (Simulation may exceed this)     ││
│                                         └───────────────────────────────────┘│
│  ● Inference Demo | Layer 2 complete: argmax([...]) = 7 (95% confidence)     │
└──────────────────────────────────────────────────────────────────────────────┘
```

### New Demo Modes for Demo 3

| Mode | What It Shows |
|------|---------------|
| **Draw & Classify** | Real-time inference as you draw |
| **Inference Demo** | Step-by-step: input → layer 1 → layer 2 → output |
| **Training Demo** | Show weight updates during learning |
| **Confusion Demo** | Click a cell to see example mistakes |
| **Batch Test** | Run 100 random samples, show accuracy |

### Key Additions Needed

1. **Layer Activation View** - Show activations flowing through network
2. **Confidence Bars** - Animated bars for each digit (0-9)
3. **Educational Panel** - "1. Your drawing becomes 784 pixels..."
4. **Hardware Accuracy Callout** - "IronLattice: 87% (88% max)"
5. **MVM Counter** - "100,352 multiplications in 1 cycle!"
6. **Operation Log** - ">> Classify: input → hidden → output"

---

## DEMO 4 UPGRADE PLAN: Peripheral Circuits as Live Slide

**Story:** "This is how it fits in a real chip"

### Current State
- ✅ CLI with DAC/ADC/TIA analysis
- ❌ No GUI

### Upgrade to Live Slide

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Demo 4: Peripheral Circuits - "How It Fits in a Real Chip"                  │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────────────────────────────┐ ┌───────────────────────────────┐ │
│  │  SIGNAL FLOW DIAGRAM                  │ │ WHAT YOU'RE SEEING            │ │
│  │                                       │ │                               │ │
│  │   Digital    DAC      Array      ADC  │ │ WRITE CYCLE DEMO              │ │
│  │   Input      │        │          │    │ │                               │ │
│  │   [01101]    │        │          │    │ │ 1. Digital value [01101]      │ │
│  │      │       ▼        ▼          ▼    │ │    = Level 13 of 30           │ │
│  │      │    ┌─────┐  ┌─────┐  ┌─────┐   │ │                               │ │
│  │      └───▶│ DAC │─▶│CELL │─▶│ ADC │   │ │ 2. DAC converts to            │ │
│  │           │5-bit│  │FeFET│  │5-bit│   │ │    +1.25V write pulse         │ │
│  │           └─────┘  └─────┘  └─────┘   │ │                               │ │
│  │              │         │        │     │ │ 3. Cell programs to           │ │
│  │            1.25V    Level 13  [01101] │ │    conductance level 13       │ │
│  │              │         │        │     │ │                               │ │
│  │           ┌─────┐      │     ┌─────┐  │ │ 4. ADC reads back [01101]     │ │
│  │           │Charge│     │     │ TIA │  │ │    = Level 13 ✓               │ │
│  │           │ Pump │     │     └─────┘  │ │                               │ │
│  │           │1V→±3V│     │              │ ├───────────────────────────────┤ │
│  │           └─────┘      │              │ │ TIMING DIAGRAM                │ │
│  └───────────────────────────────────────┘ │                               │ │
│                                            │ DAC:  ████░░░░░░░░            │ │
│  ┌───────────────────────────────────────┐ │ CELL: ░░░░████░░░░            │ │
│  │  LINEARITY ANALYSIS (INL/DNL)         │ │ ADC:  ░░░░░░░░████            │ │
│  │                                       │ │       0   50  100 ns          │ │
│  │  Level:  0  5  10  15  20  25  30     │ ├───────────────────────────────┤ │
│  │  INL:   ░░░▓▓▓░░░▓▓░░░░▓▓▓░░░▓░░     │ │ POWER BREAKDOWN               │ │
│  │  DNL:   ▓░░░▓▓░░░░▓░░▓▓░░░▓░░░▓▓     │ │                               │ │
│  │                                       │ │ DAC:        12 µW             │ │
│  │  Max INL: 0.3 LSB  Max DNL: 0.4 LSB   │ │ ADC:        45 µW             │ │
│  │  (Excellent linearity!)               │ │ TIA:         8 µW             │ │
│  └───────────────────────────────────────┘ │ Charge Pump: 25 µW            │ │
│                                            │ TOTAL:      90 µW             │ │
│  ● Write Cycle Demo | Programming level 13 → Read back: 13 ✓                │ │
└──────────────────────────────────────────────────────────────────────────────┘
```

### New Demo Modes for Demo 4

| Mode | What It Shows |
|------|---------------|
| **Write Cycle Demo** | Digital → DAC → Cell → ADC → Digital |
| **Read Cycle Demo** | Show non-destructive read operation |
| **Linearity Demo** | Sweep all 30 levels, show INL/DNL |
| **Timing Demo** | Animated timing diagram |
| **Power Demo** | Show power consumption breakdown |

### Key Additions Needed

1. **Signal Flow Diagram** - Animated data path through peripherals
2. **Timing Diagram** - Show sequence of DAC → Cell → ADC
3. **INL/DNL Plot** - Linearity analysis visualization
4. **Power Breakdown** - Pie chart or bar showing each component
5. **Educational Panel** - "1. Digital value [01101] = Level 13..."
6. **Voltage/Current Meters** - Real-time values

---

## IMPLEMENTATION CHECKLIST

For each demo upgrade, implement in this order:

### Phase 1: Layout & Components
- [ ] Create ASCII mockup of final layout
- [ ] Implement educational panel widget
- [ ] Implement operation log widget
- [ ] Add demo mode dropdown

### Phase 2: Demo Modes
- [ ] Implement auto-running demo mode
- [ ] Implement step-by-step educational mode
- [ ] Implement manual/interactive mode
- [ ] Add frequency/speed slider

### Phase 3: Polish
- [ ] Add status bar with current action
- [ ] Add "What You're Seeing" context panel
- [ ] Add key metrics/callouts
- [ ] Test presenter workflow

### Pattern: Educational Panel Text

```go
// Context-sensitive educational text based on current mode/phase
func GetEducationalText(mode string, phase int) string {
    switch mode {
    case "MVM Demo":
        switch phase {
        case 1: return "1. Input voltages applied to columns"
        case 2: return "2. Current flows through ALL cells at once"
        case 3: return "3. Row currents = dot product!"
        }
    case "Write/Read Demo":
        switch phase {
        case 1: return "WRITE: E>Ec sets state"
        case 2: return "HOLD: E=0, P persists!"
        case 3: return "READ: E<Ec, no change"
        }
    }
    return ""
}
```

### Pattern: Operation Log

```go
// Scrolling log of operations for audience to follow
type OperationLog struct {
    entries []string
    maxLines int
}

func (l *OperationLog) Add(entry string) {
    l.entries = append(l.entries, fmt.Sprintf(">> %s", entry))
    if len(l.entries) > l.maxLines {
        l.entries = l.entries[1:]
    }
}
```

---

## DR. TOUR QUOTES

> "It's got **30 discrete states**. So it's not 0-1-0-1."

> "We're at **87% validation** here... theoretical is 88%."

> "**Compute in memory** where the same device does the memory and the computation."

> "This could lower the requirements in a data center by **80 to 90%**."

> "Works on a **standard CMOS line** and can translate just like that."

> "There's **no exotic materials** in here. There's no graphene."

---

## WEEBIT NANO PRECEDENT

Dr. Tour's previous spinout (2015):
> "This company Weebit—this is another memory that came out of my lab... it's selling now on the market with three big customers."

**IronLattice is his next one.**

---

*Last updated: 2026-01-19*
