# Module 1: Hysteresis - Physics Reference

**Navigation:** [← Module 1 Index](./README.md) | [ELI5](./eli5.md) | [Features](./features.md) | [Materials](./materials.md)

---

## Evidence Status

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Core Model Overview](#core-model-overview)
3. [Key Equations](#key-equations)
4. [Parameters and Units](#parameters-and-units)
5. [Preisach Model Details](#preisach-model-details)
6. [Landau-Khalatnikov Model](#landau-khalatnikov-model)
7. [Implementation Details](#implementation-details)
8. [Write/Read Operations](#writeread-operations)
9. [Minor Loops](#minor-loops)
10. [Temperature Dependence](#temperature-dependence)
11. [What's Real vs Simplified](#whats-real-vs-simplified)

---

## Prerequisites

- Electric field and polarization basics
- Units and dimensional analysis
- Simple integrals and summations

**Recommended:** Read [ELI5 introduction](./eli5.md) first if you're new to ferroelectrics.

---

## Core Model Overview

Hysteresis is modeled as **history-dependent polarization versus electric field**. The relationship P(E) forms a loop because the path going up differs from the path coming down—this creates memory.

### Two Physics Models Available

| Model | Type | Use Case | Default |
|-------|------|----------|---------|
| **Preisach** | Phenomenological (hysteron-based) | General hysteresis, analog states | ✅ Yes |
| **Landau-Khalatnikov (L-K)** | Thermodynamic (mean-field) | Phase transitions, temperature effects | No |

The Preisach model represents the material as a weighted sum of bistable elements (hysterons). Each hysteron switches at different thresholds and contributes to total polarization.

---

## Key Equations

### Preisach Model

```
P(E) = ∬ μ(α,β) γ_{α,β}(E) dα dβ

where:
  μ(α,β) = Preisach density/weighting function
  γ_{α,β}(E) ∈ {-1, +1} = hysteron state
  α = up-switching threshold
  β = down-switching threshold
```

**Constraints:**
- α > β (valid hysterons only)
- ∬ μ(α,β) dα dβ = 1 (normalized)

**Current Implementation:** Uses tanh-based Everett function approximation, not FORC-calibrated distribution.

### Landau-Khalatnikov Model

```
F = α₀(T-Tc)P² + βP⁴ + γP⁶ - EP

dP/dt = -(1/η) dF/dP

where:
  F = Landau free energy
  Tc = Curie temperature
  η = viscosity coefficient
  α₀, β, γ = Landau coefficients
```

### Key Derived Parameters

```
Pr = P(E = 0) after saturation
Ec ≈ field where P crosses 0 on the major loop
```

---

## Parameters and Units

| Symbol | Meaning | Units | Typical Value |
|--------|---------|-------|---------------|
| **E** | Electric field | V/m (or MV/cm) | 0.5 - 6 MV/cm |
| **P** | Polarization | C/m² (or µC/cm²) | 15 - 150 µC/cm² |
| **Ec** | Coercive field | V/m | 1.0×10⁸ V/m (1 MV/cm) |
| **Pr** | Remanent polarization | C/m² | 0.25 C/m² (25 µC/cm²) |
| **Ps** | Saturation polarization | C/m² | 0.30 C/m² (30 µC/cm²) |
| **α** | Hysteron up-switch | V/m | α > Ec |
| **β** | Hysteron down-switch | V/m | β < -Ec |
| **μ(α,β)** | Preisach density | scaled to yield P | Gaussian distribution |
| **γ_{α,β}** | Hysteron state | unitless | ±1 |
| **τ** | Switching time | s | 0.3 - 100 ns |
| **Tc** | Curie temperature | K | 700 - 1300 K |

---

## Preisach Model Details

### The Hysteron Concept

A **hysteron** is the simplest element with hysteresis: a two-state switch with different ON and OFF thresholds.

```
         State
           ↑
         +1 ├────────────────╮
            │                │
            │    α (ON)      │
          0 ├────────────────┼───────→ E
            │    β (OFF)     │
            │                │
         -1 ├────────────────╯

Memory zone: β < E < α
```

**Key property:** Between β and α, the state persists—this is the hysteresis memory.

### Material as Hysteron Ensemble

Real ferroelectric = millions of tiny domains, each acting like a hysteron:

```
Macroscopic loop = Σ individual hysterons

P_total = Σ μ_i × state_i
```

### Current Implementation (Code-Based)

The module uses a **Preisach-stack formulation with tanh-based Everett function**:

- Main implementation: `module1-hysteresis/pkg/ferroelectric/preisach.go`
- Stack engine: `shared/physics/preisach.go`
- Everett kernel: `TanhEverett.Calculate(alpha, beta)`

```go
Pirrev := p.stack.Update(E)               // irreversible Preisach contribution
P := Pirrev + p.reversiblePolarization(E) // add reversible dielectric branch
```

### Everett Shape Parameter

Loop squareness is controlled by `Delta` in `TanhEverett`:

- **Smaller Delta** → sharper switching, squarer loop
- **Larger Delta** → softer transitions, slanted loop

`Delta` is auto-tuned (`tuneDeltaForPr`) to match material Pr/Ps ratio.

### Hysteron Distribution

Hysterons are distributed with Gaussian spread around ±Ec:

```
α ~ N(+Ec, σ_α)
β ~ N(-Ec, σ_β)

where:
  σ_α = alpha_sigma_ratio × Ec  (typically 0.20)
  σ_β = beta_sigma_ratio × Ec   (typically 0.20)
```

Wider distribution = more gradual switching = better for analog levels.

---

## Landau-Khalatnikov Model

### Free Energy Expansion

```
F(P,T) = α₀(T-Tc)P² + βP⁴ + γP⁶ - EP + ½ε₀ε_hf E²

Equilibrium: dF/dP = 0
Dynamics: dP/dt = -(1/η) dF/dP
```

### Phase Transition

```
For T < Tc: Two stable minima at ±Pr
For T > Tc: Single minimum at P = 0 (paraelectric)

Pr(T) ∝ √(Tc - T)  (near Tc)
```

### Parameters Used

- **Tc:** Curie temperature (723 K for HZO)
- **ε_hf:** High-frequency permittivity (30 for HZO)
- **ε_lf:** Low-frequency permittivity (40 for HZO)
- **thickness:** Film thickness (for depolarization field)

**Note:** Current temperature implementation uses linear scaling, not full Curie-law collapse.

---

## Implementation Details

### How 30 Levels Are Discretized

Continuous polarization P is mapped to discrete levels:

```go
discreteLevel = int(math.Round((normalizedP + 1) / 2 * 29))

where:
  normalizedP = P / Ps  (range: -1 to +1)
```

| Normalized P | Level | Physical State |
|--------------|-------|----------------|
| -1.0 (−Ps) | 0 | Fully negative |
| -0.5 | 7 | Mid-negative |
| 0.0 | 15 | Neutral |
| +0.5 | 22 | Mid-positive |
| +1.0 (+Ps) | 29 | Fully positive |

**Formula:** `Level = round((P/Ps + 1) × 14.5)`

### WRITE/READ Mode Detection

The UI determines current mode by comparing |E| to Ec:

```go
if math.Abs(eField) > material.Ec {
    mode = "WRITE"  // |E| > Ec: can switch polarization
} else {
    mode = "READ"   // |E| < Ec: non-destructive sensing
}
```

This provides visual context, but actual memory operations use the phase machine (see next section).

### Switching Time (τ)

**Defined but not used in real-time visualization.**

The material includes τ = 10 ns and a KAI (Kolmogorov-Avrami-Ishibashi) model:

```go
// KAI model: progress = 1 - exp(-(t/τ)^n)
// n = 2.0 (Avrami exponent for 2D domain growth)
```

But the interactive loop uses **instantaneous switching** for real-time performance (60 FPS). This is physically reasonable at 1 Hz cycling where τ = 10 ns is negligible.

---

## Write/Read Operations

### Actual Demo Control Flow

The current Write/Read demo is **controller-driven**, not a single-threshold comparator.

```
WRITE(target) → HOLD(settle) → READ(sense) → DISPLAY(result) → next target
```

### WRITE Phase (ISPP-style)

During WRITE, the controller (`module1-hysteresis/pkg/controller/writer.go`) applies bounded pulses, verifies readback level, and iterates until convergence:

- Pulse amplitude/duration stepped under controller policy
- Binary search with overshoot recovery
- Verification gates progress to next phase
- Guard-band correction for boundary states

**Key algorithms:**
- Binary search with field bounds [VMin, VMax]
- Overshoot detection and bounds adjustment
- Guard-band pulses (max 2) for fine-tuning
- ACCEPT ±1 logic after 8+ overshoots

### READ Phase

READ is a dedicated phase using sensing conditions designed to avoid state disturbance:

- Readback evaluated against target level
- Result surfaced in UI log/telemetry
- Field strength kept below Ec

### Why This Matters

1. **Matches implementation reality** — behavior from phase/controller logic, not idealized thresholds
2. **Closer to practical memory** — write-verify loops like real ISPP
3. **Physics still visible** — UI shows |E| vs Ec context for diagnostics

---

## Minor Loops

### Partial Traversal

If you reverse direction before completing a full cycle, you get a **minor loop**:

```
Full major loop:              Minor loop:
      ╭───────╮                  ╭───╮
     ╱         ╲                ╱ ╭←╯ Turned back
    │           │              │  │   early!
    │     ●     │              │  ↓
     ╲         ╱                ╲
      ╰───────╯                  ╰─
```

**The Preisach model handles this** by tracking turning points (field extrema where direction reversed).

**Stack implementation:**
- Stores history as (field, direction) turning points
- Updates hysteron states based on current field vs history
- Automatically creates correct minor loop shapes

**Why it matters:** Real memory operations do partial writes. Physics must correctly predict intermediate states.

---

## Temperature Dependence

### Current Implementation

Linear temperature scaling around 300 K using material coefficients:

```go
Ec(T) = Ec_300K + TempCoeffEc × (T - 300K)
Ps(T) = Ps_300K + TempCoeffPr × (T - 300K)
```

Then safety clamps applied (minimum Ec and Ps).

**Coefficients:**
- `TempCoeffEc`: typically -1.5×10⁵ to -2.5×10⁵ V/m/K
- `TempCoeffPr`: typically -3×10⁻⁵ to -5×10⁻⁵ C/m²/K

### Physical Expectation

Near Curie temperature, should follow:

```
Pr(T) ∝ √(Tc - T)    (Landau theory)
Ec(T) ∝ (Tc - T)     (roughly)
```

**Current status:** Pragmatic linear model, not full Curie-law collapse in Preisach implementation.

---

## What's Real vs Simplified

| Aspect | Implementation | Status |
|--------|----------------|--------|
| P from E | Preisach model (hysteron sum) | ✅ Model-based |
| Hysteresis | Emergent from hysteron memory | ✅ Model-based |
| Loop shape | Everett kernel + stack memory | ✅ Emergent |
| 30 levels | Linear discretization of P | ✅ Simple & correct |
| Minor loops | Implicit via hysteron states | ✅ Works correctly |
| Write/Read | Phase machine + controller | ✅ Implemented |
| τ switching | Defined but not in viz | ⚠️ Quasistatic approx |
| Temperature | Ec(T), Ps(T) scaling | ✅ Model-based |
| FORC calibration | Not used (tanh Everett) | ⚠️ Simplified |

---

## Assumptions and Limits

- **Quasi-static loops:** No high-frequency domain dynamics
- **Uniform material:** No spatial gradients or microstructure
- **Idealized hysterons:** Not tied to specific domain wall physics
- **Simulation defaults:** Parameters are baselines unless cited; not measured device data
- **Everett approximation:** Not FORC-calibrated from experimental data

---

## Where It Lives In Code

### Core Physics
- `module1-hysteresis/pkg/ferroelectric/preisach.go` - Tanh-based Preisach
- `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go` - Mayergoyz Preisach
- `shared/physics/preisach.go` - Stack engine
- `shared/physics/landau_khalatnikov.go` - L-K solver

### Materials
- `module1-hysteresis/pkg/ferroelectric/material.go` - Material parameter sets
- `shared/physics/material.go` - Shared material definitions

### GUI/Simulation
- `module1-hysteresis/pkg/gui/gui.go` - GUI initialization
- `module1-hysteresis/pkg/gui/physics_engine.go` - Engine switching logic
- `module1-hysteresis/pkg/gui/simulation.go` - Physics update loop

### Controller
- `module1-hysteresis/pkg/controller/writer.go` - ISPP write controller
- `module1-hysteresis/pkg/controller/state_machine.go` - Phase machine

---

## Demo Waveform Modes

| Mode | Physics Demonstrated |
|------|---------------------|
| **Manual** | Direct E-field control, explore hysteresis |
| **Sine Wave** | Full loop traversal, continuous cycling |
| **Triangle Wave** | Linear ramps showing Ec threshold |
| **Square Wave** | Fast switching dynamics |
| **Random Walk** | Multi-level storage (30-level baseline) |
| **Write/Read Demo** | Complete memory cycle with ISPP |

---

## Sources and References

### Documentation
- [hysteresis.physics.md](../../../hysteresis/hysteresis.physics.md) - Extended physics deep-dive
- [materials.md](./materials.md) - Material parameter tables
- [ELI5](./eli5.md) - Beginner-friendly explanation

### Code
- `docs/development/SCRIPT_REFERENCE.md` - Function lookups
- Test files: `module1-hysteresis/pkg/controller/*_test.go`

### Literature
- `docs/4-research/papers/by-topic/01-ferroelectric-materials/`
- `docs/4-research/video-transcripts/COSM_2025_AI_Hardware_Breakthrough/` (conference material)
- HZO P-E loop examples: DOI citations in material definitions

---

**Next Steps:**
- For practical usage → [Features](./features.md)
- For material parameters → [Materials](./materials.md)
- For CLI reference → [Run Modes](./run-modes.md)
- For integration → [Tools](./tools.md)

---

**Last Updated:** 2026-02-16
