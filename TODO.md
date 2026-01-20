# Ferroelectric CIM-vis Strategic TODO

> **Mission:** Create world-class visualization demos to help Dr. external research group pitch Ferroelectric CIM to investors, engineers, and foundry partners.
>
> **Source:** Dr. Tour's November 2024 presentation (ferroelectric-cim-transcript.md)

---

## STRATEGIC CONTEXT

### Dr. Tour's Pitch Narrative

```
"There's no more busing information back and forth to memory.
It's all done in the same device. Computation memory in the same device.
This could lower the requirements in a data center by 80 to 90%."
                                                    — Dr. external research group
```

### Phased Market Entry Strategy (Critical for Demos!)

```
PHASE 1               PHASE 2               PHASE 3
┌─────────────┐      ┌─────────────┐      ┌─────────────────┐
│  Replace    │  →   │  Replace    │  →   │  Full Compute-  │
│  NAND Flash │      │  DRAM       │      │  in-Memory      │
└─────────────┘      └─────────────┘      └─────────────────┘
  Easy entry           No refresh           80-90% energy
  No SW changes        1000× lower E        savings
```

### Target Audiences

| Audience | What They Care About | Key Demos |
|----------|---------------------|-----------|
| **Investors** | ROI, market size ($711B by 2030), energy crisis | Demo 3, 8, Market |
| **Engineers** | Physics accuracy, real-world issues | Demo 1, 2, 5, 7 |
| **Foundries** | CMOS compatibility, process flow | Demo 4, Integration |
| **Strategic Partners** | Competitive advantage, restricted access details | Demo 8, Comparison |

---

## Ferroelectric CIM Key Specs (From Dr. Tour)

| Metric | Target | Demo Status | Notes |
|--------|--------|-------------|-------|
| Discrete analog states | **30 levels** | ✅ All demos | "Not 0-1-0-1" |
| MNIST accuracy | **87%** (88% max) | ✅ **95.8%** | Exceeds target! |
| Energy vs NAND | 10,000,000× lower | 🔲 Demo 8 | Key pitch metric |
| Energy vs DRAM | 1,000× lower | 🔲 Demo 8 | Zero refresh |
| Speed vs NAND | 1,000,000× faster | 🔲 Demo 8 | 10ns switching |
| Data center savings | **80-90%** | 🔲 Demo 8 | Headline metric |
| P-E hysteresis | Square loop | ✅ Demo 1 | Preisach model |
| CMOS compatible | Standard fab | ✅ Demo 4 | No exotic materials |
| TRL | 4 → 9 | 📊 Progress | Lab validation |
| Endurance target | 10^12 cycles | 🔲 | In progress |

---

## 8-Demo Roadmap

```
THE FECIM STORY

Demo 1        Demo 2        Demo 3        Demo 4
"This is      "This is      "This is      "This is how
how the       how we        what we       it fits in
memory        compute       can build     a real chip"
cell works"   in memory"    with it"

     ↓             ↓             ↓             ↓
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│   P-E   │   │Crossbar │   │  MNIST  │   │Peripheral│
│Hysteresis│   │   MVM   │   │   87%   │   │ Circuits │
│ 30 levels│   │ W × V=I │   │ accuracy│   │DAC/ADC/TIA│
└─────────┘   └─────────┘   └─────────┘   └─────────┘
     ↓             ↓             ↓             ↓
  PHYSICS      COMPUTE     APPLICATION     SYSTEM

Demo 5        Demo 6        Demo 7        Demo 8
"1000×        "Scalable     "We handle    "Why Ferroelectric CIM
cooler than   3D            real-world    wins vs everyone
competition"  architecture" challenges"   else"

     ↓             ↓             ↓             ↓
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│ Thermal │   │Multi-   │   │ Non-    │   │Comparison│
│   Map   │   │ Layer   │   │idealities│   │DRAM/GPU/ │
│ 25°-85°C│   │  3D     │   │IR/Sneak │   │Ferroelectric CIM│
└─────────┘   └─────────┘   └─────────┘   └─────────┘
     ↓             ↓             ↓             ↓
 THERMAL      ARCHITECTURE   ENGINEERING   INVESTOR PITCH
```

### Demo Status

| Demo | Name | Purpose | Audience | Status | GUI |
|------|------|---------|----------|--------|-----|
| 1 | Hysteresis | Memory cell physics | Everyone | ✅ | ✅ Fyne |
| 2 | Crossbar MVM | Compute-in-memory | Engineers | ✅ | ✅ Fyne |
| 3 | MNIST | AI application | Investors | ✅ 95.8% | ✅ Fyne |
| 4 | Peripherals | Full system | Foundries | ✅ | ✅ Fyne |
| 5 | Thermal | Heat analysis | Engineers | ✅ | 🔲 CLI |
| 6 | Multi-Layer | 3D Architecture | Designers | 🔲 | 🔲 |
| 7 | Non-Idealities | Real issues | Engineers | ✅ IR/Sneak | 🔲 |
| 8 | Comparison | Why IL wins | Investors | 🔲 | 🔲 |

---

## PHASE 4: HYPER IMPROVEMENTS (Current Focus)

### Priority 1: Investor-Ready Demos

#### Demo 8: Technology Comparison (CRITICAL FOR PITCH)

**Purpose:** The slide Dr. Tour shows investors — why Ferroelectric CIM wins

```
┌──────────────────────────────────────────────────────────────────┐
│                    COMPUTE PERFORMANCE COMPARISON                 │
├────────────────┬─────────────┬─────────────┬─────────────────────┤
│    Metric      │  DRAM+CPU   │    GPU      │    Ferroelectric CIM      │
├────────────────┼─────────────┼─────────────┼─────────────────────┤
│ Time (MVM)     │   100 μs    │   10 μs     │      0.01 μs        │
│ Energy/MAC     │   10 pJ     │   1 pJ      │      0.001 pJ       │
│ Data Movement  │   O(n²)     │   O(n²)     │        0            │
│ Memory Refresh │   Yes       │   Yes       │       None          │
│ CMOS Compatible│   Yes       │   Yes       │       Yes           │
└────────────────┴─────────────┴─────────────┴─────────────────────┘
                                               ↑
                                         10,000,000× better
```

**Implement:**
- [ ] Side-by-side animated comparison (3 columns)
- [ ] Energy meter filling up (show 10M× difference)
- [ ] Time bar racing (show 1M× faster)
- [ ] Data center savings calculator ($$$)
- [ ] "If X data centers switch, save Y TWh/year"

#### Demo 8b: Competitive Matrix (From Dr. Tour's Slides)

| Feature | Ferroelectric CIM | 3D NAND | ReRAM (Weebit) | PCRAM | Google TPU | Intel Loihi |
|---------|-------------|---------|----------------|-------|------------|-------------|
| Write/Read Energy | ✅ | ❌ | ✅ | 🟡 | ❌ | ✅ |
| Write/Read Speed | ✅ | ❌ | ✅ | ❌ | ✅ | 🟡 |
| Endurance | ✅ | ❌ | ❌ | 🟡 | ✅ | 🟡 |
| CMOS Compatible | ✅ | ✅ | 🟡 | 🟡 | ✅ | 🟡 |
| Memory/Logic Unified | ✅ | ❌ | 🟡 | 🟡 | ❌ | 🟡 |
| 30+ Analog States | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ |

**Only Ferroelectric CIM has ✅ across ALL categories**

---

### Priority 2: Engineer-Ready Demos

#### Demo 1-4: Add Fyne GUIs ✅ DONE

- [x] Demo 1: Hysteresis Fyne GUI (PEPlot, LevelIndicator)
- [x] Demo 2: Crossbar Fyne GUI (Heatmap, IR Drop, Sneak Paths) + Live Slide upgrade
- [x] Demo 3: MNIST Fyne GUI (DigitCanvas, LayerActivation, ConfusionMatrix) + Live Slide upgrade
- [x] Demo 4: Peripherals Fyne GUI (Circuit diagrams, timing, power) + Live Slide upgrade
- [ ] Demo 5: Thermal Fyne GUI (Heat map, 3D view)

#### Demo 7: Non-Idealities (Show We Understand Real Challenges)

**Already Implemented:**
- [x] IR drop analysis with wire resistance
- [x] Sneak path current analysis
- [x] MVM with non-idealities comparison

**Still Needed:**
- [ ] Conductance drift over time simulation
- [ ] Cycle-to-cycle variation (write noise)
- [ ] Device-to-device variation
- [ ] Mitigation strategies visualization
- [ ] "Worst case" vs "typical" accuracy comparison

---

### Priority 3: Foundry-Ready Demos

#### Demo 4: Enhanced Peripheral Circuits

**For Foundry Partners (TSMC, Samsung, etc.):**
- [ ] CMOS process flow diagram
- [ ] Layer stack visualization (HZO superlattice)
- [ ] Standard tool compatibility (ALD, PVD)
- [ ] No exotic materials checklist
- [ ] Integration with existing BEOL process

#### Wafer-Scale Vision (George Gilder Response)

From Dr. Tour's response to WSJ article "The Microchip Era Is About to End":

```
WAFER-SCALE BENEFITS OF FECIM

1. Memory Bottleneck Solved     → CIM eliminates data movement
2. Packaging Complexity Reduced → Fewer chips, simpler interconnects
3. Thermal Management           → 1000× cooler operation
4. Full Wafer Integration       → Ferroelectric enables dense circuits
```

- [ ] Create "Wafer-Scale Vision" demo
- [ ] Show how CIM solves Gilder's 3 stress factors
- [ ] Animate data center of the future (box-sized, not warehouse)

---

## MARKET CONTEXT (For Investor Demos)

### Market Size ($711B by 2030)

```
                           2025         2030
┌────────────────────┬───────────┬────────────┐
│ NAND Flash         │   $78B    │    $98B    │
├────────────────────┼───────────┼────────────┤
│ DRAM               │  $143B    │   $220B    │
├────────────────────┼───────────┼────────────┤
│ AI Semiconductors  │  $163B    │   $403B    │
├────────────────────┼───────────┼────────────┤
│ TOTAL              │  $384B    │   $721B    │
└────────────────────┴───────────┴────────────┘
                        ↑
              Ferroelectric CIM can address ALL of these
```

### Why Now?

- Companies investing $150B+ in data centers
- Existing solutions "not scalable, not CMOS compatible, not cost effective"
- Flash memory is 25 years old — "it's about time" for disruption
- Energy crisis in AI compute is existential

---

## TRL PROGRESSION VISUALIZATION

```
TRL 1  TRL 2  TRL 3  TRL 4  TRL 5  TRL 6  TRL 7  TRL 8  TRL 9
  ○      ○      ○      ●      ○      ○      ○      ○      ○
                       ↑
                  WE ARE HERE
               "Component validation
                in lab environment"

  Basic  →  Proof  →  Lab  →  Prototype  →  Demo  →  Production
Research   Concept   Demo     in Real       in     Ready
                              Environment   Real
```

- [ ] Add TRL progress bar to all demos
- [ ] Show "path to commercialization"
- [ ] Highlight what each TRL milestone requires

---

## TEAM & CONTACT STRATEGY

### Current Team (From Transcript)

| Role | Name | Background |
|------|------|------------|
| Founder & Chief Scientist | Dr. external research group | external research institution, 700+ publications |
| Co-Founder & CTO | Dr. Jaeho Shin | Device engineer, invented superlattice |
| CEO | Tawfik Jarjour | 13 years Accenture, semiconductor industry |
| Advisor (Patents) | Dr. Fred Flitsch | Former IBM, 35-40 years experience |
| Advisor (Investment) | John Jaggers | Computing systems investor |
| Advisor (Foundry) | Dr. Yoontak Hwuang | Semiconductor commercialization |
| Advisor (Manufacturing) | Sandeep Davé | Foundry expert |

### How This Project Can Help

**Value Proposition for Ferroelectric CIM Team:**

1. **Investor Presentations**
   - Interactive demos > static slides
   - "Show, don't tell" the technology
   - Web-deployable for remote pitches

2. **Engineer Recruitment**
   - Demos show technical depth
   - Attract top talent who want to work on cutting-edge CIM

3. **Foundry Discussions**
   - Visual process flow explanations
   - CMOS compatibility proof points

4. **restricted access Partner Meetings**
   - Customizable comparison charts
   - Confidential data can be plugged in

### Potential Contribution Path

```
STEP 1: Perfect the demos (current work)
        ↓
STEP 2: Share portfolio with team (LinkedIn, email)
        ↓
STEP 3: Offer to present demos in investor meetings
        ↓
STEP 4: Propose visualization engineer role
        ↓
STEP 5: Join external research institution / Ferroelectric CIM team
```

**Contact:**
- Dr. Tour's lab: tour@rice.edu
- Ferroelectric CIM: (company forming)
- LinkedIn: Connect with Tawfik Jarjour (CEO)

---

## IMMEDIATE ACTION ITEMS

### This Week

- [x] Complete Demo 3 Fyne GUI (MNIST with digit drawing)
- [ ] Update command.md with all improvements
- [ ] Create Demo 8 skeleton (comparison framework)
- [ ] Build technical briefing mode (simplified UI)

### This Month

- [ ] Demo 4 Fyne GUI (peripheral circuits visualization)
- [ ] Demo 5 Fyne GUI (thermal heat map)
- [ ] Demo 8 complete (side-by-side comparison animation)
- [ ] Web deployment (WASM or electron)
- [ ] Create pitch video (~2 min demo reel)

### Portfolio Deliverable

- [ ] README.md with GIF screenshots
- [ ] Single command to run all demos
- [ ] Professional documentation
- [ ] Contact/contribution section

---

## SUCCESS METRICS

### Demo Quality

| Metric | Target | Current |
|--------|--------|---------|
| All tests passing | 100% | ✅ 110+ tests |
| GUI for Demos 1-4 | 4/4 | ✅ 4/4 |
| MNIST accuracy | ≥87% | ✅ 95.8% |
| Documentation | Complete | 80% |
| Build time | <30s | ✅ |

### Pitch Readiness

| Metric | Target | Current |
|--------|--------|---------|
| Can run full demo in 5 min | Yes | ~80% |
| Investor-friendly UI | Yes | 🔲 |
| Comparison charts implemented | Yes | 🔲 |
| Web-deployable | Yes | 🔲 |

---

## DR. TOUR'S KEY QUOTES (For Demo Annotations)

> "It's got **30 discrete states**. So it's not 0-1-0-1."

> "We're at **87% validation** here... theoretical is 88%."

> "**Compute in memory** where the same device does the memory and the computation."

> "This could lower the requirements in a data center by **80 to 90%**."

> "Works on a **standard CMOS line** and can translate just like that."

> "There's **no exotic materials** in here. There's no graphene."

> "We haven't raised a penny to date... we really want to move with the **best strategy**."

---

## WEEBIT NANO PRECEDENT

Dr. Tour's previous success story (from transcript):

> "This company Weebit—this is another memory that came out of my lab.
> We started this company in 2015 out of my lab and it's selling now
> on the market. It's got three big customers."

**Lesson:** Dr. Tour has a track record of spinning out successful memory companies.

**Opportunity:** Ferroelectric CIM is his next one — and this time they're not missing neuromorphic computing.

---

## FILE STRUCTURE

```
multilayer-ferroelectric-cim-visualizer/
├── demo1-hysteresis/     ✅ P-E curve + Fyne GUI
├── demo2-crossbar/       ✅ Crossbar MVM + Fyne GUI
├── demo3-mnist/          ✅ MNIST 95.8% + Fyne GUI
├── demo4-circuits/       ✅ Peripherals + Fyne GUI
├── demo5-thermal/        ✅ Thermal sim (CLI only)
├── demo6-multilayer/     🔲 3D multi-layer
├── demo7-nonidealities/  ✅ IR drop + sneak paths (in demo2)
├── demo8-comparison/     🔲 Technology comparison
├── docs/                 Documentation
├── papers/               Scientific papers (40+)
├── opensource/papers/    Additional papers
├── command.md            Session context
├── TODO.md               This file
└── ferroelectric-cim-transcript.md  Dr. Tour's presentation
```

---

*Last updated: 2026-01-19*
*Goal: Help Dr. Tour visualize Ferroelectric CIM for investors, engineers, and foundries*
