# Hysteresis Explained Like I'm 5

**Understanding Ferroelectric Memory Through Simple Analogies**

---

**Note:** References to “30 levels” refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

**Implementation note:** Current Preisach behavior is computed with a **tanh Everett approximation** (`Delta`-tuned), **not** with a FORC-calibrated Preisach distribution extracted from measured FORC data.

## Part 1: What is Hysteresis? (The Rubber Band)

### The Simplest Explanation

Imagine you have a rubber band. When you stretch it and let go:
- It doesn't snap back to EXACTLY where it started
- It "remembers" being stretched for a moment
- The path going UP is different from the path coming DOWN

**That's hysteresis!** The output (position) depends not just on the input (pull), but on what happened BEFORE.

```
Stretching:              Releasing:

   ↑                        ↓
   │    ●───→               │    ●
   │   ╱                    │     ╲
   │  ╱                     │      ╲
   │ ╱                      │       ╲
   ●                        │        ●
                            ↓

   Different paths!
```

---

## Part 2: Why Does It Matter for Computers?

### Regular Memory (Like a Whiteboard)

Write something → Erase → Gone forever

```
┌──────────────┐     Power off     ┌──────────────┐
│  Hello!      │  ──────────────→  │              │
└──────────────┘                   └──────────────┘
     (data)                          (empty!)
```

### Ferroelectric Memory (Like Carving in Clay)

Write something → Turn off power → Still there!

```
┌──────────────┐     Power off     ┌──────────────┐
│  ⌒ Hello! ⌒  │  ──────────────→  │  ⌒ Hello! ⌒  │
└──────────────┘                   └──────────────┘
   (carved in)                      (still there!)
```

**The hysteresis loop is what makes this work!**

---

## Part 3: The P-E Loop (The Magic Graph)

### What Are P and E?

| Letter | Stands For | Simple Meaning |
|--------|------------|----------------|
| **E** | Electric Field | The "push" you apply (like pressing a button) |
| **P** | Polarization | The "response" of the material (like the button position) |

### The Loop Explained

```
                           ③ MAXIMUM (all tiny magnets aligned)
                              ╭───────╮
         P                   ╱         ╲
         ↑                  │           │
         │     ②           │           │     ④
     +Pr ├────────●         │           │        ← MEMORY STATE
         │       ╱           │           │         (P when E=0)
         │      ╱             ╲         ╱
         │     ●───────────────●───────●
         ├─────┼───────────────┼───────┼────→ E
         │     │      ①        │       │
     -Pr ├────────────────────●        │        ← OTHER MEMORY STATE
         │      ╲             ╱         │
         │       ╲           │           │
         │        ╲         ╱
         │         ╰───────╯
                       ⑤

              -Ec    0    +Ec
                     ↑
              SWITCHING POINT
        (need this much push to flip)
```

### Walking Around the Loop (Like a Hike)

| Step | What's Happening | Hiking Analogy |
|------|------------------|----------------|
| ① | Start at +Pr, no push applied | Standing at the top of a hill |
| ② | Push harder (+E increases) | Climbing higher |
| ③ | Maximum! Can't go higher | At the peak |
| ④ | Stop pushing, come back to +Pr | Slide back to stable ledge |
| ⑤ | Push the other way (-E) | Go down the other side |
| ⑥ | Cross through -Ec | Pass the "sticky point" |
| ⑦ | Reach -Pr | Stable on the other side! |

**The magic:** When you stop pushing (E = 0), you stay at either +Pr or -Pr. That's MEMORY!

---

## Part 4: Why 30 Levels, Not Just 2?

### Binary Memory (Regular)

Like a light switch: ON or OFF

```
      │
  ON  ●
      │
      │
      │
  OFF ●
      │
```

**1 bit of information**

### Ferroelectric Memory (FeCIM)

Like a dimmer switch with 30 positions!

```
      │
  30  ●
  29  ●
  28  ●
   ⋮  ⋮
  16  ●
  15  ●
   ⋮  ⋮
   2  ●
   1  ●
   0  ●
      │
```

**~5 bits of information (log₂(30) ≈ 4.9)**

### Why This Matters

| Memory Type | Bits per Cell | Cells for 1 MB |
|-------------|---------------|----------------|
| Binary | 1 bit | 8,388,608 cells |
| FeCIM (30-level baseline) | ~5 bits | ~1,677,722 cells |

**Same storage, 5x fewer cells!**

---

## Part 5: The Hysteron Concept (The Stubborn Magnets)

### What's a Hysteron?

Think of a tiny magnet that:
- Flips UP at one voltage (say +1.2V)
- Flips DOWN at a DIFFERENT voltage (say -0.8V)
- **Stays put** in between!

```
              UP at +1.2V
                  │
    ──────────────┼──────────────  Voltage
                  │         │
                  │    DOWN at -0.8V
                  │         │
    [────GAP────]  ← In this gap, it REMEMBERS!
```

### Material = Millions of Hysterons

Each with slightly different flip voltages:

```
Hysteron 1: flips at +1.0V / -0.9V
Hysteron 2: flips at +1.3V / -0.7V
Hysteron 3: flips at +0.9V / -1.1V
    ⋮
Hysteron 450: flips at +1.1V / -0.8V

Add them all up → Smooth hysteresis loop!
```

**The loop shape EMERGES from the distribution of hysterons. It's not drawn — it's physics!**

---

## Part 5.5: Why the Loop EMERGES (Step by Step)

When you slowly push and pull on these stubborn switches, something interesting happens:

```
                    PUSH HARD →

        "Okay, I flipped!"
              ╭───────╮
             ╱    3    ╲
            │           │
       2   │           │   4
           ●           ●
          ╱             ╲
    1 ───●───────────────●─── 5
          ╲             ╱
           ●           ●
       8   │           │   6
            │           │
             ╲    7    ╱
              ╰───────╯

                    ← PULL HARD
```

**The loop EMERGES because each hysteron flips at different voltages:**
1. Push a little → the "easy" hysterons start to flip (low threshold)
2. Push harder → more hysterons flip (medium threshold)
3. Push really hard → even the "stubborn" ones flip (high threshold)
4. Stop pushing → all hysterons STAY where they are (memory!)
5. Pull back → they DON'T flip immediately (different threshold going down!)
6. Keep pulling → now they start flipping the other way
7. Pull really hard → ALL flipped the other way
8. Stop → they stay again!

**The key insight:** Each hysteron has a GAP between its "flip up" and "flip down" voltage. This gap creates hysteresis!

```
One hysteron example:
         Flip UP at +1.2V
              │
    ──────────┼──────────────  E
              │         │
              │    Flip DOWN at -0.8V
              │         │
    [───GAP───]  ← In this gap, it REMEMBERS its state!
```

---

## Part 6: Write vs Read (The Sticky Threshold)

### The Key Insight

There's a magic voltage called **Ec** (coercive field):
- Push HARDER than Ec → Things change (WRITE)
- Push SOFTER than Ec → Things stay the same (READ)

```
         |←───── READ ZONE ─────→|←── WRITE ──→|
         |     (safe sensing)    |  (changes!) |
         |                       |             |
    ─────┼───────────────────────┼─────────────┼────→ Voltage
         0                      Ec           Emax
```

### In Practice

| Operation | Voltage | What Happens |
|-----------|---------|--------------|
| **WRITE** | > Ec | Polarization changes, new data stored |
| **READ** | < Ec | Polarization unchanged, just sense it |
| **HOLD** | = 0 | Polarization stays (memory!) |

---

## Part 7: The Perfect Hysteresis Module for FeCIM

### What Should It Do?

A production-ready hysteresis simulation module for FeCIM would need these capabilities:

---

### 7.1 Core Hysteresis Model Requirements

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **Preisach model** | Physics-accurate hysteron-based simulation | ✅ Implemented |
| **Emergent loops** | Loop shape comes from math, not drawn | ✅ Implemented |
| **Minor loops** | Partial cycles close correctly | ✅ Implemented |
| **History tracking** | Remember previous turning points | ✅ Implemented |
| **30-level quantization** | Discrete states for FeCIM | ✅ Implemented |

**What "production-ready" looks like:**
```python
model = FeCIMHysteresis(
    material="HZO_superlattice",
    Pr=25e-6,           # C/cm²
    Ps=30e-6,           # C/cm²
    Ec=1.0e6,           # V/cm
    levels=30,
    model_type="mayergoyz_preisach"
)

# Apply field, get polarization with full history tracking
P = model.update(E)

# Get discrete level
level = model.get_level()  # 0-29
```

---

### 7.2 Temperature and Environmental Effects

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **Temperature-dependent Ec** | Linear scaling around 300K via `TempCoeffEc` | ✅ Implemented |
| **Temperature-dependent Pr/Ps** | Linear scaling via `TempCoeffPr` | ✅ Implemented |
| **Curie-law collapse above Tc** | Ec,Pr → 0 above Tc | ❌ Not implemented in code |
| **Real-time T adjustment** | GUI slider for temperature | ⚠️ Partial/limited GUI support |

**What "production-ready" looks like:**
```python
model.set_temperature(350)  # Kelvin
# Ec and Pr automatically adjust

# High-temperature behavior in current simulator
model.set_temperature(800)
# Applies linear coefficient scaling + safety clamps
# (does NOT force Ec = 0, Pr = 0 by Curie-law collapse)
```

---

### 7.3 Switching Dynamics

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **KAI model** | Kolmogorov-Avrami-Ishibashi switching | ✅ Implemented |
| **Switching time τ** | ~1-10 ns for HZO | ✅ Parameter defined |
| **Time-resolved simulation** | P(t) during switching | ✅ SimulateDomainSwitching() |
| **Time-resolved visualization** | See switching happen | ⚠️ Implemented but not in GUI |

**What "production-ready" looks like:**
```python
# Simulate switching dynamics
times, P_values, switched_count = model.simulate_switching(
    E_applied=2*Ec,
    duration=100e-9,  # 100 ns
    steps=1000
)

# KAI model: P(t) = Ps × (1 - exp(-(t/τ)^n))
# n = 2.0 for 2D domain growth
```

---

### 7.4 Reliability Effects (Wake-up and Fatigue)

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **Wake-up modeling** | Pr increases first ~100 cycles | ✅ Basic |
| **Fatigue modeling** | Pr decreases after many cycles | ✅ Basic |
| **Cycle counting** | Track total cycles | ✅ Implemented |
| **Live degradation display** | Show Pr vs. cycle count | ⚠️ Model ready, GUI missing |

**What "production-ready" looks like:**
```python
# After cycling
for _ in range(1000):
    model.full_cycle()

cycles, degradation, wakeup = model.get_fatigue_state()
# cycles = 1000
# degradation = 0.00001% (very low for HZO superlattice)
# wakeup = 1.0 (fully woken up)

# Pr after N cycles
Pr_aged = model.material.EnduranceAtCycles(1e10)
# Returns ~90% of original Pr
```

---

### 7.5 Multi-Level Programming

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **30 discrete levels (baseline)** | Linear in polarization | ✅ Implemented |
| **Level-to-voltage mapping** | V_prog for each level | ✅ DiscreteStates() |
| **Level-to-conductance** | G for crossbar integration | ✅ Implemented |
| **Programming sequence** | Write-verify cycles | ⚠️ Not visualized |

**What "production-ready" looks like:**
```python
states = model.get_discrete_states(30)

for state in states:
    print(f"Level {state.level}:")
    print(f"  Polarization: {state.polarization} C/m²")
    print(f"  Normalized P: {state.normalized_p}")  # -1 to +1
    print(f"  Program V: {state.voltage} V")
    print(f"  Conductance: {state.conductance} S")
```

---

### 7.6 Preisach Plane Visualization

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **Hysteron grid** | Show α-β plane | ❌ Not exposed by current API |
| **State coloring** | +1 vs -1 hysterons | ❌ No `GetPreisachPlane()` in current code |
| **Distribution weights** | Show μ(α,β) | ❌ No public getter in module API |
| **Interactive display** | Watch hysterons flip | ❌ Depends on missing plane/state API |

**What "production-ready" looks like:**
```python
# Not yet available in current module API:
# alphas, betas, states = model.get_preisach_plane()
# A dedicated export/debug API is required first.
```

---

### 7.7 Export and Integration

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **P-E loop export** | Save loop data (CSV/JSON) | ⚠️ Not exposed as a stable module API yet |
| **Crossbar integration** | Feed conductances to Module 2 | ✅ Via 30-level baseline |
| **Controller telemetry export** | Emit WRD phase/controller state for analysis | ⚠️ Available in diagnostics paths, not packaged as one-click export API |

**What "production-ready" looks like:**
```python
# Export loop + controller telemetry in one call
model.export_trace("hysteresis_trace.csv")

# Include: E, P, normalized_P, level,
# WRD phase, controller state, target/readback level
```

---

### 7.8 Visualization Requirements

| Requirement | Description | Current State |
|-------------|-------------|---------------|
| **Real-time P-E plot** | 60 FPS loop tracing | ✅ Implemented |
| **30-level bar** | Show current discrete state | ✅ Implemented |
| **WRITE/READ indicator** | Mode based on |E| vs Ec | ✅ Implemented |
| **Temperature control** | Slider to adjust T | ⚠️ Missing in GUI |
| **Preisach plane** | Hysteron state heatmap | ⚠️ Model ready, GUI missing |
| **Fatigue tracker** | Cycle count and degradation | ⚠️ Model ready, GUI missing |

---

### 7.9 The "Dream" Hysteresis Module

If someone built the ultimate ferroelectric hysteresis simulation:

```
$ fecim-hysteresis interactive

┌───────────────────────────────────────────────────────────────────────────┐
│  FeCIM Hysteresis Simulator v2.0                                           │
├───────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────────────┐  ┌─────────────────────┐  ┌────────────────────┐    │
│  │   P-E LOOP      │  │   PREISACH PLANE    │  │   PARAMETERS       │    │
│  │                 │  │                     │  │                    │    │
│  │      ╭───╮      │  │   α ↑               │  │  Material: HZO     │    │
│  │     ╱     ╲     │  │     ███░░░░░        │  │  Pr: 25 µC/cm²     │    │
│  │    │   ●   │    │  │     ██░░░░░░        │  │  Ec: 1.2 MV/cm     │    │
│  │     ╲     ╱     │  │     █░░░░░░░ → β    │  │  T: 300 K ────●    │    │
│  │      ╰───╯      │  │                     │  │  Cycles: 1.2M      │    │
│  │                 │  │  ██ = switched UP   │  │  Wakeup: 100%      │    │
│  │  Level: 24/30   │  │  ░░ = switched DOWN │  │  Fatigue: 0.01%    │    │
│  │  Mode: [READ]   │  │                     │  │                    │    │
│  └─────────────────┘  └─────────────────────┘  └────────────────────┘    │
│                                                                           │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │  SWITCHING DYNAMICS (KAI Model)                                      │  │
│  │                                                                      │  │
│  │  P ┼ ─────────────────────────●●●●●●●●●●●●●●●●●                      │  │
│  │    │                      ●●●●                                       │  │
│  │    │                   ●●●                                           │  │
│  │    │               ●●●●                                              │  │
│  │    │         ●●●●●●                                                  │  │
│  │    ├──●●●●●───────────────────────────────────────────→ t            │  │
│  │    0                                               τ=10ns            │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                           │
│  [Manual] [Sine] [Triangle] [Square] [Random] [Write/Read Demo]          │
│                                                                           │
│  Export: [P-E Loop] [Verilog-A] [NeuroSim] [SPICE]                       │
└───────────────────────────────────────────────────────────────────────────┘
```

---

### 7.10 Gap Analysis: Current vs. Perfect

| Feature | Current | Perfect | Gap |
|---------|---------|---------|-----|
| Preisach model | ✅ Complete | ✅ | None |
| Temperature GUI | ⚠️ Model only | Slider + display | Small |
| Preisach plane viz | ⚠️ Model only | Heatmap in GUI | Medium |
| Switching dynamics viz | ⚠️ Model only | Animation | Medium |
| Fatigue tracking GUI | ⚠️ Model only | Live counter | Small |
| SPICE export | ❌ | Verilog-A params | Medium |
| NeuroSim export | ❌ | Config file | Medium |

### 7.11 Development Effort Estimate

| Enhancement | Effort | Priority |
|-------------|--------|----------|
| Temperature slider in GUI | 2 hours | High |
| Preisach plane heatmap | 4 hours | Medium |
| Switching dynamics plot | 4 hours | Medium |
| Fatigue counter display | 1 hour | Low |
| Verilog-A export | 8 hours | High |
| NeuroSim export | 4 hours | Medium |
| **Total** | ~23 hours | |

---

## Part 8: Glossary (Big Words Made Simple)

| Term | Simple Definition |
|------|-------------------|
| **Hysteresis** | Output depends on history, not just current input |
| **Polarization (P)** | How much tiny magnets are aligned inside |
| **Electric Field (E)** | The "push" from applied voltage |
| **Coercive Field (Ec)** | The push needed to flip the magnets |
| **Remanent (Pr)** | Polarization when you stop pushing (the memory!) |
| **Saturation (Ps)** | Maximum possible polarization |
| **Hysteron** | One tiny bistable switch with two flip points |
| **Preisach Model** | Math model using many hysterons |
| **Minor Loop** | Going partway around the loop and back |
| **KAI Model** | How fast the magnets flip over time |
| **Wake-up** | Polarization increasing during first cycles |
| **Fatigue** | Polarization decreasing after many cycles |
| **Curie Temperature** | Temperature where ferroelectricity disappears |

---

## Part 9: Learning Resources

### Beginner

- **This demo!** Run `./launch.sh` (or the unified app command)
- **(this file)** - You're reading it!
- **YouTube:** "Ferroelectric memory explained"

### Intermediate

- **../hysteresis/hysteresis.physics.md** for deep physics
- **../hysteresis/hysteresis.research.md** for paper references
- **hysteresis.opensource.md** for tools

### Advanced

- `module1-hysteresis/pkg/ferroelectric/preisach.go` source code
- **Mayergoyz "Mathematical Models of Hysteresis" (1986)**
- **Park et al. "Ferroelectricity in Doped Hafnium Oxide" (2015)**

---

## Part 10: Summary

### The Bottom Line

**Hysteresis is what makes ferroelectric memory work.**

1. **The loop shape** comes from millions of tiny switches (hysterons)
2. **The memory** comes from the gap between "flip up" and "flip down" voltages
3. **30-level baseline** gives us ~5x more storage than binary
4. **Write vs Read** is orchestrated by a WRD phase-machine/controller (with `|E|` vs `Ec` as physics context)

### What Demo 1 Shows

| Feature | What It Demonstrates |
|---------|---------------------|
| P-E Loop | How polarization responds to field |
| 30 Levels (baseline) | Multi-level storage capability |
| WRITE/READ | Threshold-based memory operations |
| Minor Loops | History-dependent behavior |
| Material Compare | Different HZO variants |

### The Key Insight

> **Hysteresis isn't a bug — it's the feature that enables memory!**

The fact that the path up is different from the path down means the system REMEMBERS which way you pushed it. That's the foundation of all non-volatile ferroelectric memory.

### Summary for Kids

| Concept | Simple Version |
|---------|---------------|
| Ferroelectric | A material with stubborn magnets inside |
| Hysteresis | The magnets remember which way you pushed them |
| 30 Levels (baseline) | Like a 30-floor parking garage for data |
| Non-volatile | Remembers even when unplugged (like a carved rock) |
| Compute-in-Memory | Do math where the data lives (no commute!) |

### Technical Note: What's Actually Running

For the curious, here's what the demo actually computes:

| What you see | What's really happening |
|--------------|------------------------|
| The loop shape | Preisach stack + tanh Everett kernel (history dependent) |
| The smooth curve | Controlled by Everett width parameter `Delta` |
| The 30 levels (baseline) | Simple formula: `Level = round((P/Ps + 1) × 14.5)` |
| Memory effect | Turning-point/stack memory in Preisach operator |
| WRITE/READ behavior | Driven by WRD phase-machine + controller; UI also shows `|E|` vs `Ec` context |
| Memory Log | Tracks phase transitions in Write/Read Demo mode |

The physics is real — the loop is **emergent**, not drawn!

---

### Quick Reference Card

```
┌─────────────────────────────────────────────────────────────┐
│              HYSTERESIS QUICK REFERENCE                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  KEY EQUATION:                                              │
│     P = Σ μ(α,β) × γ(α,β)   [Preisach model]               │
│                                                             │
│  KEY VALUES (HZO):                                          │
│     Pr = 25 µC/cm²   (remanent polarization)               │
│     Ps = 30 µC/cm²   (saturation polarization)             │
│     Ec = 1.2 MV/cm   (coercive field)                      │
│     τ  = 1-10 ns     (switching time)                      │
│     Tc = 723 K       (Curie temperature)                   │
│                                                             │
│  OPERATIONS:                                                │
│     WRITE: WRD controller applies pulse/verify cycles      │
│     READ:  WRD read phase uses bounded sensing pulses       │
│     HOLD:  E = 0      → P persists (MEMORY!)               │
│                                                             │
│  LEVELS:                                                    │
│     Level = round((P/Ps + 1) × 14.5)   [0 to 29]          │
│     Bits/cell = log₂(30) ≈ 4.9                             │
│                                                             │
│  TEMPERATURE (current simulator):                           │
│     Ec(T) = Ec_300K + TempCoeffEc·(T-300K)                │
│     Ps(T) = Ps_300K + TempCoeffPr·(T-300K)                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Part 11: Running the Demo

### Quick Start

```bash
# From project root
./launch.sh
# Or build and run
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
# Then select "Hysteresis" tab
```

### Things to Try

#### 1. Sine Wave Mode (Default)
- Watch the loop form automatically
- See how P "lags behind" E — that's the memory!

#### 2. Random Walk Mode
- Select "Random Walk" from the waveform dropdown
- Watch it pick random levels and ramp to them
- This shows "store this, store that" — real memory operation!

#### 3. Write/Read Demo Mode (Best for understanding!)
- Select "Write/Read Demo" from the dropdown
- Watch the 4-phase cycle:
  - **WRITE**: Voltage goes HIGH (past Ec) → level changes
  - **HOLD**: Voltage returns to ZERO → level STAYS! (memory!)
  - **READ**: Small voltage pulse (below Ec) → level unchanged
  - **DISPLAY**: Shows what was written vs what was read
- The Memory Log shows each operation!

#### 4. Manual Mode
- Select "Manual" and drag the slider yourself
- Stop halfway — see how the level "remembers" where you stopped
- Try different materials — some have "stickier" magnets than others

#### 5. Frequency Slider
- Speed up or slow down ANY mode with the frequency slider
- Slow = easier to see what's happening
- Fast = more dramatic!

---

*"The memory is in the loop."*

---

*This document is part of the FeCIM Visualizer project. For deep physics, see [../hysteresis/hysteresis.physics.md](../hysteresis/hysteresis.physics.md). For research details, see [../hysteresis/hysteresis.research.md](../hysteresis/hysteresis.research.md). For open-source tools, see [hysteresis.opensource.md](hysteresis.opensource.md). For demo-specific documentation, see [../hysteresis/hysteresis.demo.md](../hysteresis/hysteresis.demo.md).*
