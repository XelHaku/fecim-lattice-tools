<!-- Category: Physics | Module: module2-crossbar | Reading time: ~10 min -->
# Module 2 Physics: Crossbar Array Computation

> Formal physics and equations for the crossbar simulator --
> ideal MVM, conductance models, and non-ideality physics.

## Prerequisites

- Ohm's law (V = IR, I = GV)
- Matrix-vector multiplication
- Basic circuit analysis (Kirchhoff's laws)

---

## 1. Ideal Matrix-Vector Multiply (MVM)

```
I_i = sum_{j=1}^{cols}  G_ij * V_j

where:
  I_i   = output current from row i
  G_ij  = conductance of cell at position (i, j)
  V_j   = input voltage applied to column j
```

Each cell contributes i_ij = G_ij * V_j. Kirchhoff's current law sums the
contributions along each row wire automatically.

---

## 2. Conductance Quantization

```
Discrete level: L in {0, 1, 2, ..., 29}

G(L) = G_min + (L / 29) * (G_max - G_min)

  G_min = 10 uS  (OFF state)
  G_max = 100 uS (ON state)
  Ratio = G_max / G_min = 10
```

---

## 3. Conductance Models

Three models are available:

| Model | Equation | Use Case |
|-------|----------|----------|
| Linear | G = G_min + g_norm * (G_max - G_min) | Simple, fast prototyping |
| Exponential | G = G_min * (G_max/G_min)^g_norm | Physics-accurate FeFET behavior |
| Lookup Table | G = Table[level] | Highest accuracy (from measurements) |

The exponential model matches the exponential relationship between threshold
voltage and channel conductance in FeFETs.

---

## 4. Non-Ideality Physics

### 4.1 IR Drop (Wire Resistance)

Metal interconnects have resistance. Voltage drops cumulatively along the wire.

```
V_eff[i][j] = V_WL[i] - V_BL[j] - IR_drop_row[j] - IR_drop_col[i]

IR_drop_row[j] = j * R_wordline * I_row[i]
IR_drop_col[i] = i * R_bitline  * I_col[j]

Defaults:
  R_wordline = 2.5 ohm/pitch
  R_bitline  = 2.5 ohm/pitch
```

Temperature dependence of wire resistance:

```
R(T) = R_0 * (1 + TCR * (T - T_ref))
  TCR = 0.00393 /K  (copper)
  T_ref = 300 K
```

Architecture scaling: 0T1R arrays see 1.5x higher effective wire resistance
(more total current from sneak paths).

### 4.2 Sneak Paths (Passive Crossbar)

In 0T1R arrays, current can flow through unselected cells via three-cell
paths:

```
Selected cell at (r_s, c_s):
Sneak path: WL[r_s] -> Cell[r_s][j] -> BL[j] ->
            Cell[k][j] -> WL[k] -> Cell[k][c_s] -> BL[c_s]

Series conductance of path:
  G_sneak = 1 / (1/G1 + 1/G2 + 1/G3)

Total sneak current: I_sneak = V * G_sneak
```

For arrays up to 32x32: full enumeration of all three-cell paths.
For larger arrays: simplified model (1% sneak factor upper bound).

Architecture impact:

| Architecture | Sneak Isolation |
|--------------|----------------|
| 0T1R (passive) | 1x (full sneak) |
| 1T1R (transistor) | 1000x isolation |

### 4.3 Device Variation

Three sources combined multiplicatively:

```
G_actual = G_nominal * noise * gradient_x * gradient_y * edge_factor

noise      = 1 + sigma * randn()         (per-cell random, sigma = 5%)
gradient_x = 1 + GradientX * (col - center)  (systematic, 0.1%/cell)
gradient_y = 1 + GradientY * (row - center)
edge_factor = 1 - EdgeEffect if on boundary  (5% degradation)
```

### 4.4 Conductance Drift

Power-law model for time-dependent conductance change:

```
G(t) = G_0 * (t / t_0)^(-nu)

  nu = drift coefficient
  t_0 = reference time (1 second)
```

| Technology | nu | Source |
|-----------|-----|--------|
| FeFET (conservative) | 0.001 | Estimated from retention data |
| FeFET (literature) | 0.0005 | Derived from retention data |
| RRAM | 0.05 | Literature |
| PCM | 0.1 | Literature |

Thermal activation: drift rate follows Arrhenius with Ea ~ 0.5 eV.

### 4.5 Temperature Effects

| Effect | Model |
|--------|-------|
| Wire resistance | R(T) = R_0 * (1 + 0.00393 * (T - 300)) |
| Conductance window (cryo) | 1.5x expansion below 100 K |
| Conductance window (hot) | 10% narrowing per 100 K above 300 K |
| Drift rate | Arrhenius: rate ~ exp(-Ea / kT) |

---

## 5. Parameters and Units

| Symbol | Meaning | Units | Typical Value |
|--------|---------|-------|---------------|
| G | Conductance | S (Siemens) | 10-100 uS |
| V | Input voltage | V | 0-1 V |
| I | Output current | A | uA range |
| R_wire | Wire resistance per pitch | ohm | 2.5 ohm |
| sigma | Device variation | fraction | 0.05 (5%) |
| nu | Drift exponent | unitless | 0.0005-0.001 |
| TCR | Temperature coefficient of resistance | 1/K | 0.00393 |
| Ea | Activation energy | eV | 0.5 |

---

## 6. Energy Model

```
Array energy  = rows * cols * 0.01 fJ        (cell read)
ADC energy    = rows * 0.5 pJ * 2^(bits-6)   (output conversion)
DAC energy    = cols * 0.1 pJ                 (input conversion)
Total         = Array + ADC + DAC

GPU reference = MAC_operations * 10 pJ/MAC
Efficiency    = GPU_energy / FeCIM_energy
```

Typical 128x128 array with 6-bit ADC: ~0.67 pJ total vs ~163 pJ GPU equivalent.

---

## 7. Assumptions and Limits

- Linear conductance model for ideal MVM (before non-idealities)
- Non-idealities use literature-inspired parameters, not device-calibrated
- Noise modeled as additive Gaussian perturbations
- Drift coefficients estimated from retention data, not directly measured
- Sneak path simplified model for arrays > 32x32

---

## Where It Lives in Code

| Component | File |
|-----------|------|
| Array, MVM, quantization | `module2-crossbar/pkg/crossbar/array.go` |
| Non-idealities (IR, sneak) | `module2-crossbar/pkg/crossbar/nonidealities.go` |
| IR drop analysis | `module2-crossbar/pkg/crossbar/irdrop.go` |
| Sneak path calculation | `module2-crossbar/pkg/crossbar/sneakpath.go` |
| Drift simulation | `module2-crossbar/pkg/crossbar/drift.go` |
| Temperature effects | `module2-crossbar/pkg/crossbar/temperature.go` |
| Differential arrays | `module2-crossbar/pkg/crossbar/enhanced.go` |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
