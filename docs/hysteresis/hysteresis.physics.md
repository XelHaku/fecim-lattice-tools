# Ferroelectric Physics: Deep Technical Reference

> **Citation audit note:** Any factual/numeric claim in this file that does not include a verified DOI citation should be treated as **[CITATION NEEDED - placeholder value]**.

Start here if you've never studied ferroelectrics before.

**Note:** References to “30 levels” refer to the demo baseline (configurable). Literature reports multi-level states (not verified here). Numeric values below are simulation defaults or illustrative unless a peer-reviewed citation is provided (**[CITATION NEEDED - placeholder value]**).

**Implementation note:** Current Preisach behavior is computed with a **tanh Everett approximation** (`Delta`-tuned), **not** with a FORC-calibrated Preisach distribution extracted from measured FORC data.

---

## Part 1: What Are We Even Talking About?

### Atoms and Charges

Everything is made of atoms. Atoms have:
- **Protons (+)** in the center (positive charge)
- **Electrons (-)** orbiting around (negative charge)

When positive and negative charges are separated, we call this a **dipole**:

```
     Before                After applying force

     ⊕⊖                      ⊕─────⊖
   (neutral)              (dipole - charges separated)
```

### What is Polarization?

**Polarization (P)** = how much the charges are separated, on average, in a material.

```
Unpolarized crystal:       Polarized crystal:
┌────────────────┐         ┌────────────────┐
│ ⊕⊖  ⊕⊖  ⊕⊖  │         │ ⊕→⊖ ⊕→⊖ ⊕→⊖ │
│ ⊕⊖  ⊕⊖  ⊕⊖  │         │ ⊕→⊖ ⊕→⊖ ⊕→⊖ │  →→→ Net P
│ ⊕⊖  ⊕⊖  ⊕⊖  │         │ ⊕→⊖ ⊕→⊖ ⊕→⊖ │
└────────────────┘         └────────────────┘
   P = 0                      P > 0 (pointing right)
```

**Units of P:** microcoulombs per square centimeter (μC/cm²)
- 25 μC/cm² means: 25 microcoulombs of charge separation per cm² of material

### What is an Electric Field?

An **Electric Field (E)** is the "push" felt by charges in a region.

```
                  Electric Field E →
         ─────────────────────────────→
         ─────────────────────────────→
         ─────────────────────────────→

Positive charges feel pushed RIGHT →
Negative charges feel pushed LEFT ←
```

**Units of E:** megavolts per centimeter (MV/cm)
- 1 MV/cm = 1,000,000 volts across 1 centimeter of material

**Relationship:** If you apply 1V across a 10nm film:
```
E = Voltage / Thickness = 1V / 10nm = 1V / 10⁻⁶cm = 1 MV/cm
```

---

## Part 2: What Makes Ferroelectrics Special?

### Normal Materials (Dielectrics)

In most materials:
1. Apply electric field → charges separate (polarize)
2. Remove electric field → charges return to original position
3. **No memory!**

```
Normal material response:

P (polarization)
↑
│       /
│      /
│     /
│    /
│   /
├──/─────→ E (field)
│
Same path up and down!
```

### Ferroelectric Materials: They REMEMBER!

In ferroelectric materials:
1. Apply field → charges separate AND crystal structure shifts
2. Remove field → **charges stay separated!**
3. **MEMORY!** (like a light switch that stays on)

```
Crystal structure shift (simplified):

    Before field           After field (stays!)
    ┌───┬───┬───┐         ┌───┬───┬───┐
    │   │ ⊕ │   │         │   │   │   │
    │ ⊕ │   │ ⊕ │ ──E→→   │ ⊕ │ ⊕ │ ⊕ │ ← Center atom
    │   │ ⊕ │   │         │   │   │   │    moved UP
    └───┴───┴───┘         └───┴───┴───┘

    P = 0                  P > 0 (permanent!)
```

The center atom literally moves to a new stable position in the crystal lattice!

---

## Part 3: Hysteresis - The Loop

### What is Hysteresis?

**Hysteresis** = Greek for "lagging behind"

The output (P) doesn't just depend on the input (E)—it depends on the **history** of what happened before.

**Real-world examples of hysteresis:**
- Thermostat: Turns on at 68°F, off at 72°F (not same point!)
- Mechanical switch: Clicks on, clicks off at different positions
- Rubber band: Stretching path ≠ releasing path

### The P-E Hysteresis Loop

```
                  ③ Saturated (all dipoles aligned)
                     ╭───────╮
         P          ╱         ╲
         ↑         │           │
         │    ②   │           │   ④
     Ps ─┼────────●           ●───────  ← SATURATION
         │       ╱             ╲         (maximum possible P)
     Pr ─┼─────●               │        ← REMANENT
         │     │               │         (P when E=0, THE MEMORY!)
         ├─────┼───────────────┼────→ E
         │     │    ①          │
    -Pr ─┼─────┼───────────────●
         │     ╲               ╱
    -Ps ─┼─────────●           ●───────
         │          ╲         ╱
         │           ╰───────╯
                         ⑤

              -Ec    0    +Ec
                     ↑
              COERCIVE FIELD
        (field needed to SWITCH direction)
```

### Walking Through the Loop

| Step | What happens |
|------|--------------|
| ① | Start: P = +Pr (positive remanent state, no field applied) |
| ② | Apply +E: P increases toward saturation |
| ③ | At high +E: Saturated, P = Ps (all dipoles aligned up) |
| ④ | Reduce E to 0: P drops to Pr (STILL POSITIVE! Memory!) |
| ⑤ | Apply -E: P decreases, crosses through 0 at -Ec |
| ⑥ | At high -E: Saturated negative, P = -Ps |
| ⑦ | Reduce E to 0: P = -Pr (negative remanent state) |
| ⑧ | Apply +E: Must reach +Ec to flip back positive |

### Key Parameters Explained

| Parameter | Name | Meaning | Example (simulation default; [CITATION NEEDED - placeholder value]) |
|-----------|------|---------|-----------|
| **Ps** | Saturation Polarization | Maximum possible separation of charges | 25 μC/cm² |
| **Pr** | Remanent Polarization | Polarization remaining at zero field (THE MEMORY) | ~20 μC/cm² |
| **Ec** | Coercive Field | Field required to flip the polarization | 1.0 MV/cm |

**In plain terms:**
- **Ps = 25 μC/cm²** → "When all dipoles align, we get this much charge separation"
- **Ec = 1.0 MV/cm** → "Need 1 million volts per centimeter to flip the switch"
  - For 10nm film: 1.0 MV/cm × 10nm = 1.0V needed to switch!

---

## Part 4: Why 30 States, Not Just 2?

### Binary Memory (Traditional)

Normal flash memory: ON or OFF (1 or 0)
```
     P
     ↑
 +Pr ├────● State 1 ("ON")
     │
   0 ├────
     │
 -Pr ├────● State 0 ("OFF")
```

### Analog Memory (Ferroelectric CIM)

Ferroelectrics can be set to IN-BETWEEN values!

```
     P
     ↑
 +Ps ├────● State 30
     ├────● State 29
     ├────● State 28
     ├    ⋮
     ├────● State 16
   0 ├────● State 15
     ├    ⋮
     ├────● State 2
     ├────● State 1
 -Ps ├────● State 0
```

**How?** By stopping at different points on the hysteresis curve using precisely controlled voltage pulses.

**Why useful?**
- Each cell stores 5 bits instead of 1 bit (log₂(30) ≈ 5)
- For AI: Can represent neural network weights directly (analog compute)

---

## Part 5: The Hysteron Concept

### What is a Hysteron?

A **hysteron** is the simplest possible element with hysteresis: a switch that turns ON and OFF at **different** thresholds.

```
Think of it like a sticky light switch:

         Output (on/off)
           ↑
         1 ├────────────────╮
           │                │
           │    α (ON)      │
         0 ├────────────────┼───────→ Input
           │    β (OFF)     │
           │                │
        -1 ├────────────────╯

α = 2V to turn ON
β = 1V to turn OFF

So: ON at 2V, OFF at 1V, NOT THE SAME!
```

### Material = Many Hysterons

Real ferroelectric = millions of tiny domains, each acting like a hysteron with slightly different (α, β):

```
   One big loop = sum of many small hysterons

   ╭──╮   =   [╭╮] + [╭╮] + [╭╮] + ... millions
  ╱    ╲       α₁β₁   α₂β₂   α₃β₃
 │      │
  ╲    ╱
   ╰──╯
```

This is the **Preisach model**: The macroscopic hysteresis loop emerges from the sum of microscopic hysterons.

---

## Part 6: Write vs Read Operations

### The Actual Demo Control Flow (Controller + Phase Machine)

The current Write/Read demo is **controller-driven**, not a single-threshold comparator.
`|E|` vs `Ec` still matters physically, but operation sequencing is orchestrated by a
phase machine in `simulation.go` plus the write controller in
`module1-hysteresis/pkg/controller`.

```
WRITE(target) → HOLD(settle) → READ(sense) → DISPLAY(result) → next target
```

### WRITE Phase (ISPP-style pulse/verify loop)

During WRITE, the controller applies bounded pulses, verifies readback level, and
iterates until target level is reached (or retry limits trigger failure).

- Pulse amplitude/duration are stepped under controller policy
- Verification gates progress to the next phase
- Convergence and recalibration happen between targets when needed

### READ Phase (bounded sensing in sequence)

READ is a dedicated phase in the WRD sequence. It uses sensing conditions intended
to avoid state disturbance while obtaining readback level for compare/logging.

- Readback is evaluated against target/programmed level
- Result is surfaced in UI log/telemetry before advancing

### Why This Matters

1. **Matches implementation reality** — behavior comes from phase/controller logic
2. **Closer to practical memory operation** — write-verify loops, not one-shot thresholding
3. **Physics still visible** — UI `|E|` vs `Ec` indicator provides field-context diagnostics

---

## Part 7: Minor Loops

### What if We Don't Complete the Full Cycle?

If you only go partway around the loop and reverse, you get a **minor loop**:

```
Full major loop:              Minor loop:
      ╭───────╮                  ╭───╮
     ╱         ╲                ╱ ╭←╯
    │           │              │  │ Turned back
    │     ●     │              │  ↓ early!
     ╲         ╱                ╲
      ╰───────╯                  ╰─
```

**The Preisach model handles this** by tracking "turning points" (where you reversed direction).

**Why it matters:** In real memory operation, you might do partial writes, and the physics must correctly predict what happens.

---

## Summary Table

| Term | Plain English | Unit |
|------|---------------|------|
| **Polarization (P)** | How much positive/negative charges are separated | μC/cm² |
| **Electric Field (E)** | The "push" on charges from applied voltage | MV/cm |
| **Saturation (Ps)** | Maximum possible polarization | μC/cm² |
| **Remanent (Pr)** | Polarization that remains when E=0 (the memory!) | μC/cm² |
| **Coercive Field (Ec)** | Field needed to flip the polarization direction | MV/cm |
| **Hysteresis** | Output depends on history, path up ≠ path down | - |
| **Hysteron** | One tiny switch element with ON/OFF thresholds | - |
| **Preisach Model** | Many hysterons with distributed thresholds = one loop | - |

---

## What Demo 1 Visualizes

With this understanding, Demo 1 shows:

1. **The P-E Loop** - As you drag voltage left/right, watch P trace the hysteresis curve
2. **The 30 States** - See which analog level you're at based on P value
3. **Minor Loops** - Reverse direction partway and see the inner loops form
4. **Material Comparison** - Different Ec, Ps values → different loop shapes
5. **WRITE vs READ** - WRD phase-machine executes WRITE/HOLD/READ/DISPLAY while UI also shows `|E|` vs `Ec` context
6. **Memory Operations Log** - Watch actual WRITE/READ cycles in the Write/Read Demo mode

### The Write/Read Demo Mode

The demo includes a special mode that demonstrates actual memory operations:

```
Phase 1: WRITE     Phase 2: HOLD      Phase 3: READ      Phase 4: DISPLAY
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ E > Ec      │    │ E = 0       │    │ E < Ec      │    │ E = 0       │
│ P changes!  │ →  │ P persists! │ →  │ P unchanged │ →  │ Show result │
│ WRITE mode  │    │ MEMORY!     │    │ READ mode   │    │ Wrote=Read? │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

**Key insight:** The HOLD phase proves non-volatility — P stays at the written value even when E returns to zero!

---

## Part 8: How Demo 1 Actually Implements the Physics

This section documents what the code does — based on code review.

### Core Model: Preisach stack + Tanh Everett (current code)

The module currently uses a **Preisach-stack formulation with a tanh-based Everett function**, not an explicit per-hysteron `[]Hysteron` mesh in this file.

- Main implementation: `module1-hysteresis/pkg/ferroelectric/preisach.go`
- Stack engine: `shared/physics/preisach.go`
- Everett kernel: `TanhEverett.Calculate(alpha, beta)` in `module1-hysteresis/pkg/ferroelectric/preisach.go`

`PreisachModel.Update(E)` computes:

```go
Pirrev := p.stack.Update(E)              // irreversible Preisach contribution
P := Pirrev + p.reversiblePolarization(E) // reversible dielectric branch
```

So the P-E loop remains history-dependent and Preisach-like, but the current representation is **stack/turning-point based**, with Everett-weighted areas, rather than the explicit `Hysteron{Alpha,Beta,State}` loop shown in older drafts.

### Where Hysteresis Comes From

**The hysteresis is EMERGENT, not forced.** Here's why:

1. Each hysteron has Alpha > Beta (e.g., α = +1.1 Ec, β = -0.9 Ec)
2. When E increases past Alpha → hysteron switches to +1
3. When E decreases past Beta → hysteron switches to -1
4. **Between Beta and Alpha: the state PERSISTS** — this is the memory

```
For one hysteron with α = 1.2, β = -1.0:

          E increasing →
State: -1 ─────────────┬───────── +1
                       │α=1.2
                       │
          ← E decreasing
State: +1 ─────────────────────┬─ -1
                              │β=-1.0

The gap between α and β is where hysteresis lives!
```

### Everett Shape Parameter (Why the Loop is Square-ish)

In the current implementation, loop squareness is controlled primarily by the Everett width parameter `Delta` in `TanhEverett`.

- Smaller `Delta` → sharper switching and squarer loop
- Larger `Delta` → softer/slanted transitions

`Delta` is tuned (`tuneDeltaForPr`) to match material remanent polarization (`Pr`) relative to irreversible saturation (`Ps`).

### How 30 Levels Are Discretized

The continuous polarization P is mapped to discrete levels in the simulation loop:

```go
a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * 29))
```

Where `normalizedP = P / Ps` ranges from -1 to +1.

### WRITE/READ Mode Detection

The UI determines the current mode by comparing |E| to Ec:

```go
if math.Abs(eField) > a.material.Ec {
    a.modeLabel.SetText("Mode: [WRITE] |E|>Ec")
} else {
    a.modeLabel.SetText("Mode: (READ) |E|<Ec")
}
```

| Normalized P | Level |
|--------------|-------|
| -1.0 (−Ps)   | 0     |
| 0.0          | 15    |
| +1.0 (+Ps)   | 29    |

**Formula:** `Level = round((P/Ps + 1) × 14.5)` = 0 to 29

### Does τ (Switching Time) Affect the Visualization?

**No — the real-time loop uses instantaneous switching.**

The simulation targets real-time frame rates (often ~60 FPS on typical hardware) and calls:
```go
a.polarization = a.preisach.Update(a.electricField)
```

This `Update()` switches hysterons instantaneously when E crosses their thresholds.

The τ = 10 ns switching time IS defined in the material and there IS a `SimulateDomainSwitching()` function using KAI (Kolmogorov-Avrami-Ishibashi) dynamics:

```go
// KAI model: progress = 1 - exp(-(t/τ)^n)
// n = 2.0 (Avrami exponent for 2D domain growth)
```

But this function is **not called** during the interactive visualization loop. This is physically reasonable: at 1 Hz cycling, τ = 10 ns is negligible (the system is always in equilibrium).

### Temperature Dependence

Current code applies **linear temperature scaling around 300 K** using material coefficients (`TempCoeffEc`, `TempCoeffPr`) in `PreisachModel.SetTemperature`:

- `Ec(T) = Ec_300K + TempCoeffEc*(T-300K)`
- `Ps(T) = Ps_300K + TempCoeffPr*(T-300K)`

Then safety clamps are applied (minimum `Ec` and `Ps`).

This is a pragmatic simulator model; it is **not** currently a Curie-law collapse model (`Ec→0` above `Tc`) in the implementation.

---

## Summary: What's Real vs. Simplified

| Aspect | Implementation | Status |
|--------|---------------|--------|
| P from E | Preisach model (hysteron sum) | ✅ Model-based |
| Hysteresis | Emergent from hysteron memory | ✅ Model-based |
| Loop shape | From Everett kernel shape (`Delta`) + Preisach stack memory | ✅ Emergent, not forced |
| 30 levels (baseline) | Linear discretization of P | ✅ Simple & correct |
| Minor loops | Implicit via hysteron states | ✅ Works correctly |
| Write vs Read | WRD phase-machine + controller sequencing (with `|E|`/`Ec` context indicator) | ✅ Implemented |
| τ switching | Defined but not used in viz | ⚠️ Quasistatic approx |
| Temperature | Ec(T) scaling implemented | ✅ Model-based |

---

## Demo Waveform Modes

| Mode | Physics Demonstrated |
|------|---------------------|
| Manual | Direct E-field control, hysteresis exploration |
| Sine Wave | Full hysteresis loop traversal |
| Triangle Wave | Linear ramps showing Ec threshold |
| Square Wave | Fast switching dynamics |
| Random Walk | Multi-level storage (30-level baseline) |
| Write/Read Demo | Complete memory operation cycle |

---

*This document is part of the FeCIM Visualizer project. For beginner explanations, see [../hysteresis/hysteresis.ELI5.md](../hysteresis/hysteresis.ELI5.md). For research references, see [../hysteresis/hysteresis.research.md](../hysteresis/hysteresis.research.md). For open-source tools, see [hysteresis.opensource.md](hysteresis.opensource.md).*
