<!-- Category: Physics | Module: module1-hysteresis | Reading time: ~12 min -->
# Module 1 Physics: Ferroelectric Hysteresis

> Formal physics treatment of the two models used in Module 1:
> Preisach (phenomenological) and Landau-Khalatnikov (thermodynamic).

## Prerequisites

- Electric field and polarization basics
- Units and dimensional analysis (V/m, C/m^2)
- Simple integrals, summations, and differential equations

---

## 1. Landau-Khalatnikov (L-K) Equation

The L-K model treats the ferroelectric as a thermodynamic system whose free
energy F has a double-well shape in polarization P.

### Free Energy

```
F(P, T, E) = alpha_0 (T - Tc) P^2  +  beta P^4  +  gamma P^6  -  E P
```

| Symbol | Meaning | Units |
|--------|---------|-------|
| F | Landau free energy density | J/m^3 |
| P | Polarization | C/m^2 |
| T | Temperature | K |
| Tc | Curie temperature | K |
| alpha_0 | Landau coefficient (1st order) | J m / C^2 K |
| beta | Landau coefficient (2nd order) | J m^5 / C^4 |
| gamma | Landau coefficient (3rd order) | J m^9 / C^6 |
| E | Applied electric field | V/m |

### Dynamics

The polarization evolves toward the free-energy minimum:

```
dP/dt = -(1/eta) * dF/dP
```

where eta is the viscosity (damping) coefficient with units of (V s / m).
The equilibrium condition dF/dP = 0 gives the static P-E curve.

### Phase Transition

```
T < Tc  -->  two stable minima at +/- Pr (ferroelectric)
T > Tc  -->  single minimum at P = 0 (paraelectric)

Near Tc:  Pr(T) ~ sqrt(Tc - T)
```

### L-K Parameters (HZO Default)

| Parameter | Value | Notes |
|-----------|-------|-------|
| Tc | 723 K (450 C) | Curie temperature |
| epsilon_hf | 30 | High-frequency permittivity |
| epsilon_lf | 40 | Low-frequency permittivity |
| thickness | 10 nm | For depolarization field |

---

## 2. Preisach Model

The Preisach model represents the material as a weighted collection of
elementary bistable switches called hysterons.

### Hysteron

A single hysteron has two states (+1 and -1) and two switching thresholds:

```
       State
         ^
       +1 |---------.
          |         |
          |   alpha |   (switches ON here)
        0 |---------+----------> E
          |   beta  |   (switches OFF here)
          |         |
       -1 |---------'

  Memory zone: beta < E < alpha
  (hysteron retains its current state)
```

Key property: alpha > beta always. Between beta and alpha the hysteron holds
whatever state it was last pushed into. That is hysteresis at the single-
element level.

### Preisach Integral

The total polarization is the weighted sum over all hysterons:

```
P(E) = integral integral  mu(alpha, beta) * gamma_{alpha,beta}(E)  d alpha  d beta
```

| Symbol | Meaning |
|--------|---------|
| mu(alpha, beta) | Preisach density (weighting function) |
| gamma_{alpha,beta}(E) | Hysteron state: +1 or -1 |
| alpha | Up-switching threshold for this hysteron |
| beta | Down-switching threshold for this hysteron |

Constraint: alpha > beta (valid hysterons only in the upper-left triangle of
the Preisach plane).

### Preisach Plane

```
  alpha ^
        |  . . . . . . .
        | . . . . . . .    <-- each dot is one hysteron
        |. . . . . . .         with thresholds (alpha, beta)
        |. . . . . . .
        |. . . . . .
        +-----------------> beta

  Only the triangle alpha > beta is populated.
  Hysterons near (Ec, -Ec) dominate the major loop.
```

### Tanh-Based Everett Function (Implementation)

Instead of storing millions of explicit hysterons, the simulator uses a
continuous Everett kernel based on the product of two tanh functions:

```
E(alpha, beta) = [1 + tanh((alpha - Ec)/Delta)] * [1 - tanh((beta + Ec)/Delta)] * Ps/4
```

This product form is the mathematically correct integral of a sech^2 Preisach
density. It is always non-negative (unlike the older difference form that
required clamping).

The shape parameter Delta controls loop squareness:
- Smaller Delta --> sharper switching, squarer loop
- Larger Delta --> softer transition, more slanted loop

Delta is auto-tuned to match the material's Pr/Ps ratio.

### Hysteron Distribution

Hysterons are distributed with Gaussian spread around +/- Ec:

```
alpha ~ N(+Ec, sigma_alpha)       sigma_alpha = 0.20 * Ec
beta  ~ N(-Ec, sigma_beta)        sigma_beta  = 0.20 * Ec
```

Wider distribution = more gradual switching = better for analog levels.

### Minor Loops and Turning-Point Stack

When the field reverses before completing a full cycle, a minor loop forms.
The Preisach stack tracks every turning point (field extremum where direction
reversed). This history determines which hysterons are ON and which are OFF,
producing correct minor loop shapes automatically.

---

## 3. ISPP Algorithm

Incremental Step Pulse Programming converges a cell to a target discrete level
using a binary-search strategy:

```
1. Set bounds [V_min, V_max] around estimated target voltage
2. Apply pulse at midpoint voltage
3. Read back actual level
4. If level == target: SUCCESS
5. If level < target: raise V_min to midpoint (need more field)
6. If level > target: lower V_max to midpoint (need less field)
7. If overshoot detected: widen bounds, recover
8. Repeat until converged or max iterations
```

Key algorithmic details:
- Guard-band correction: up to 2 fine-tuning pulses for boundary states
- Overshoot recovery: bounds widened minimally, not reset to full range
- ACCEPT +/-1 logic: after 8+ overshoots, accept level within 1 of target
  (physics-limited convergence for sharp-switching materials)
- Overshoot limit (30): triggers success, not failure (physics limitation)

---

## 4. Material Parameter Table

All parameters are simulation baselines from literature calibration. They are
not measured data from a specific device.

| Material | Ec (MV/cm) | Pr (uC/cm^2) | Ps (uC/cm^2) | Tau (ns) | Tc (C) |
|----------|-----------|--------------|--------------|---------|--------|
| Standard HZO | 1.2 | 25 | 30 | 1 | 450 |
| FeCIM HZO | 1.0 | 30 | 35 | 10 | 450 |
| Superlattice | 0.85 | 50 | 55 | 0.36 | 500 |
| Cryogenic HZO (4K) | 1.5 | 75 | 80 | 1 | 450 |
| AlScN | 5.0 | 120 | 130 | 10 | 1000 |
| HZO-32 | 1.0 | 20 | 25 | 10 | 450 |
| FTJ-140 | 1.2 | 25 | 30 | 20 | 450 |

### Selection Guide

| Goal | Best material |
|------|---------------|
| Lowest write voltage | Superlattice (Ec = 0.85 MV/cm) |
| Highest polarization | AlScN (Pr = 120 uC/cm^2) |
| Best endurance | Standard HZO or Superlattice (10^10 cycles) |
| Cryogenic operation | Cryogenic HZO (enhanced Pr at 4K) |
| Most analog states | FTJ-140 (up to 140 states reported) |

---

## 5. Temperature Dependence

Current implementation uses linear scaling around 300 K:

```
Ec(T) = Ec(300K) + TempCoeffEc * (T - 300)
Ps(T) = Ps(300K) + TempCoeffPr * (T - 300)
```

Typical coefficients:
- TempCoeffEc: -1.5e5 to -2.5e5 V/m/K (Ec decreases with temperature)
- TempCoeffPr: -3e-5 to -5e-5 C/m^2/K (Pr decreases with temperature)

Full Landau theory predicts Pr(T) ~ sqrt(Tc - T) near Tc, but the current
Preisach implementation uses the pragmatic linear model.

---

## 6. What Is Real vs Simplified

| Aspect | Status | Notes |
|--------|--------|-------|
| P-E hysteresis loop | Model-based | Emergent from Preisach hysteron sum |
| Minor loops | Correct | Implicit via turning-point stack |
| 30-level discretization | Simple and correct | Linear mapping of P to levels |
| ISPP convergence | Implemented | Binary search with overshoot recovery |
| Switching time (tau) | Defined but not in viz | Quasistatic approximation at 1 Hz |
| Temperature dependence | Model-based | Linear scaling, not Curie-law collapse |
| FORC calibration | Not used | Tanh Everett approximation instead |
| Spatial domain structure | Not modeled | Uniform material assumption |

---

## Where It Lives in Code

| Component | File |
|-----------|------|
| Preisach model | `module1-hysteresis/pkg/ferroelectric/preisach.go` |
| Preisach stack engine | `shared/physics/preisach.go` |
| Everett kernel | `shared/physics/tanh_everett.go` |
| L-K solver | `shared/physics/landau_khalatnikov.go` |
| ISPP write controller | `module1-hysteresis/pkg/controller/writer.go` |
| Material definitions | `module1-hysteresis/pkg/ferroelectric/material.go` |
| GUI / simulation loop | `module1-hysteresis/pkg/gui/simulation.go` |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
