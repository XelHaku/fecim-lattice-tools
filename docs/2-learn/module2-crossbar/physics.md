# Module 2: Crossbar - Physics Reference

**Navigation:** [← Module 2 Index](./README.md) | [ELI5](./eli5.md) | [Features](./features.md) | [Architecture](./architecture.md)

---

## Evidence Status

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

---

## Prerequisites

- Ohm's law (V = IR)
- Matrix-vector multiplication
- Basic circuit networks (Kirchhoff's laws)

**Recommended:** Read [ELI5](./eli5.md) first for intuition.

---

## Core Model

A crossbar array computes **y = G · v** in hardware, where:
- **G** is the conductance matrix (programmable weights)
- **v** is the input voltage vector
- **y** is the output current vector

Non-idealities (IR drop, sneak paths, variation, drift) perturb this ideal behavior.

---

## Key Equations

### 1. Ideal Matrix-Vector Multiply

```
I_i = Σ(j=1 to cols) G_ij × V_j

where:
  I_i = output current from row i
  G_ij = conductance of cell at (i,j)
  V_j = input voltage applied to column j
```

**Physical Implementation:**
- Apply voltages V_j to columns (word lines)
- Measure currents I_i from rows (bit lines)
- Each cell contributes: i_ij = G_ij × V_j
- Row current is automatic sum (Kirchhoff's current law)

### 2. Conductance Quantization

```
Discrete level: L ∈ {0, 1, 2, ..., 29}

G(L) = G_min + (L / 29) × (G_max - G_min)

where:
  G_min = 10 µS (OFF state, minimum conductance)
  G_max = 100 µS (ON state, maximum conductance)
  Conductance ratio = G_max / G_min = 10
```

**Normalized form (used in code):**
```
g_normalized = L / 29  (range: [0, 1])
G_physical = G_min + g_normalized × (G_max - G_min)
```

### 3. Conductance Models

Three models available via `Config.ConductanceModel`:

#### Linear Model (default, simple)
```
G(g_norm) = G_min + g_norm × (G_max - G_min)
```

**Pros:** Simple, fast
**Cons:** Doesn't match FeFET physics at extremes

#### Exponential Model (physics-accurate)
```
G(g_norm) = G_min × exp(ln(G_max / G_min) × g_norm)
            = G_min × (G_max / G_min)^g_norm
```

**Pros:** Matches exponential VT-dependent FeFET behavior
**Cons:** Slightly more computation

#### Lookup Table (measurement-calibrated)
```
G(L) = ConductanceTable[L]  // 30-element array

where ConductanceTable is measured data from real devices
```

**Pros:** Highest accuracy (from measurements)
**Cons:** Requires calibration data

---

## Non-Ideality Physics

### 1. IR Drop (Voltage Drop on Wires)

**Cause:** Metal interconnects have resistance

**Model:**
```
V_eff[i][j] = V_WL[i] - V_BL[j] - IR_drop_row[j] - IR_drop_col[i]

where:
  IR_drop_row[j] = j × R_wordline × I_row[i]
  IR_drop_col[i] = i × R_bitline × I_col[j]

  R_wordline = resistance per cell pitch (default: 2.5 Ω)
  R_bitline = resistance per cell pitch (default: 2.5 Ω)
```

**Physical constants:**
- Copper resistivity: ρ = 1.68×10⁻⁸ Ω·m
- Wire cross-section (32nm node): ~0.05 µm²
- Contact resistance: ~50 Ω per via

**Temperature dependence:**
```
R(T) = R₀ × (1 + TCR × (T - T_ref))

where:
  TCR = 0.00393 /K (copper temperature coefficient)
  T_ref = 300 K
```

**Architecture scaling:**
```
if Architecture == "0T1R":
    R_wordline *= 1.5  // More sneak current → higher total current
    R_bitline *= 1.5
```

**Impact:** Cells farther from drivers see reduced voltage → accuracy degradation

### 2. Sneak Paths (Parasitic Currents)

**Cause:** In passive (0T1R) crossbars, current can flow through unselected cells

**Three-cell sneak path example:**
```
Selected cell at (r_s, c_s):
Path: WL[r_s] → Cell[r_s][c_j] → BL[c_j] →
      Cell[r_i][c_j] → WL[r_i] → Cell[r_i][c_s] → BL[c_s]

Series conductance:
G_sneak = 1 / (1/G₁ + 1/G₂ + 1/G₃)

Total sneak current:
I_sneak = V_in × G_sneak
```

**Full calculation (for arrays ≤ 32×32):**
```
For each row i:
  For each intermediate column j:
    For each source row k ≠ i:
      g1 = Cell[k][input_col]
      g2 = Cell[k][j]
      g3 = Cell[i][j]
      I_sneak += V × (1 / (1/g1 + 1/g2 + 1/g3))
```

**Simplified model (for arrays > 32×32):**
```
I_sneak ≈ 0.01 × I_ideal  (1% sneak factor, upper bound)
```

**Architecture impact:**
```
if Architecture == "1T1R":
    isolation_factor = 0.001  // Transistor provides ~1000× isolation
    I_sneak *= isolation_factor
```

**Why 1T1R helps:** Transistor blocks reverse current paths

### 3. Device Variation

**Sources:**

#### a) Random Variation (per-cell)
```
G_actual = G_nominal × (1 + σ × randn())

where:
  σ = NoiseLevel (default: 0.05 = 5%)
  randn() = Gaussian random variable, μ=0, σ=1
```

#### b) Systematic Gradients (wafer-scale)
```
gradient_x = 1 + GradientX × (col - center_col)
gradient_y = 1 + GradientY × (row - center_row)

where:
  GradientX, GradientY = spatial gradient (%/cell, default: 0.1%)
```

#### c) Edge Effects (boundary cells)
```
if cell is on array edge:
    edge_factor = 1 - EdgeEffect  (default: 5% degradation)
```

**Combined effect:**
```
G_final = G_nominal × noise_factor × gradient_x × gradient_y × edge_factor
```

### 4. Conductance Drift

**Physical cause:** Ionic migration, charge trapping, depolarization over time

**Power-law drift model:**
```
G(t) = G₀ × (t / t₀)^(-ν)

where:
  ν = drift coefficient
  t₀ = reference time (1 second)
```

**Drift coefficients (literature comparison):**

| Technology | ν | Source |
|------------|---|--------|
| FeFET (assumed) | 0.001 | Conservative estimate |
| FeFET (literature-derived) | 0.0005 | From retention data |
| RRAM | 0.05 | Literature |
| PCM | 0.1 | Literature |
| Flash | 0.02 | Literature |

**Important:** FeCIM drift is **estimated** from retention requirements (>10 years), not directly measured.

**Thermal activation (Arrhenius):**
```
rate(T) = rate_ref × exp(-Ea / (k_B × T))

where:
  Ea ≈ 0.5 eV (ferroelectric switching activation energy)
  k_B = 8.617×10⁻⁵ eV/K (Boltzmann constant)
```

### 5. Temperature Effects

#### Wire Resistance
```
R(T) = R₀ × (1 + 0.00393 × (T - 300))  // Copper TCR
```

#### Conductance Window
```
Cryogenic (< 100K):
  Window expansion: 1.5× (Pr increases from ~25 to ~75 µC/cm² at 4K)

Room temperature (300K):
  No adjustment (baseline)

High temperature (> 300K):
  Window narrowing: 0.9 per 100K (conservative model)
```

#### Drift Rate
```
drift_rate(T) ∝ exp(-Ea / k_B T)

Higher temperature → faster drift
```

---

## Parameters and Units

### Crossbar Array

| Symbol | Meaning | Units | Typical Value |
|--------|---------|-------|---------------|
| **G** | Conductance | S (Siemens) | 10-100 µS |
| **V** | Input voltage | V | 0-1 V |
| **I** | Output current | A | µA range |
| **R_wire** | Wire resistance | Ω | 2.5 Ω/pitch |
| **G_ij** | Cell (i,j) conductance | S or unitless [0,1] | Normalized in MVM |
| **V_j** | Input voltage column j | V | Applied voltage |
| **I_i** | Row i summed output | A | Measured current |
| **V_drop** | Voltage drop on wire | V | IR drop |

### Non-Idealities

| Symbol | Meaning | Units | Typical Value |
|--------|---------|-------|---------------|
| **σ** | Device variation | % | 5% |
| **ν** | Drift exponent | unitless | 0.0005-0.001 |
| **TCR** | Temperature coefficient | /K | 0.00393 |
| **E_a** | Activation energy | eV | 0.5 |
| **G_ratio** | G_max / G_min | unitless | 10 |

### Energy

| Parameter | Value | Units |
|-----------|-------|-------|
| Cell read energy | 0.01 | fJ |
| ADC energy (6-bit) | 0.5 | pJ |
| DAC energy | 0.1 | pJ/input |
| GPU MAC energy | 10 | pJ |

---

## Assumptions and Limits

- **Linear conductance model** for ideal MVM (before non-idealities)
- **Non-idealities are simplified:** Not device-calibrated, using literature-inspired parameters
- **Noise is additive:** Modeled as Gaussian perturbations
- **Baseline MVM uses normalized conductance:** Physical units applied in analysis paths
- **Drift coefficients are estimated:** Not directly measured for FeCIM
- **Temperature model is pragmatic:** Not full thermodynamic simulation

---

## Where It Lives in Code

### Core Array
- `module2-crossbar/pkg/crossbar/array.go` - Array structure, MVM, quantization
- `module2-crossbar/pkg/crossbar/nonidealities.go` - Enhanced MVM with all effects

### Non-Idealities
- `module2-crossbar/pkg/crossbar/irdrop.go` - IR drop analysis
- `module2-crossbar/pkg/crossbar/sneakpath.go` - Sneak path calculations
- `module2-crossbar/pkg/crossbar/drift.go` - Drift simulation
- `module2-crossbar/pkg/crossbar/temperature.go` - Temperature effects

### Enhanced Operations
- `module2-crossbar/pkg/crossbar/enhanced.go` - Differential arrays, write-verify

### Tests
- `module2-crossbar/pkg/crossbar/*_test.go` - Physics validation tests

---

## Validation

### Test Coverage

```bash
# Physics correctness tests
go test -run TestArrayMVM ./module2-crossbar/pkg/crossbar
go test -run TestQuantization ./module2-crossbar/pkg/crossbar

# Non-ideality tests
go test -run TestIRDrop ./module2-crossbar/pkg/crossbar
go test -run TestSneakPaths ./module2-crossbar/pkg/crossbar
go test -run TestVariation ./module2-crossbar/pkg/crossbar

# Temperature tests
go test -run TestTemperature ./module2-crossbar/pkg/crossbar
```

### Golden Values

Reference calculations verified against:
- Hand calculations for simple cases
- CrossSim (Sandia) for sneak paths
- BadCrossbar for IR drop
- Literature values for drift coefficients

---

## Sources and References

### Implementation
- **MVM:** Standard crossbar array textbook equations
- **IR Drop:** Analog circuit analysis (Sedra/Smith)
- **Sneak Paths:** RRAM crossbar literature (adapted for FeFET)
- **Drift:** Comparison with RRAM, PCM, Flash literature

### Literature (see `research.md` for full citations)
- Retention data: Fraunhofer IPMS 2024
- Temperature effects: CEA-Leti 2024
- Cryogenic operation: Advanced Electronics Materials 2024

### Code References
- `docs/development/SCRIPT_REFERENCE.md#demo-2-crossbar-module2-crossbar`
- `docs/crossbar/reference/ARCHITECTURE.md` (comprehensive architecture doc)
- `docs/crossbar/reference/PHYSICS.md` (detailed physics audit)

---

## Next Steps

- **For features** → [features.md](./features.md)
- **For code structure** → [architecture.md](./architecture.md)
- **For integration** → [tools.md](./tools.md)
- **For voltage schemes** → [voltage-rules.md](./voltage-rules.md)

---

**Last Updated:** 2026-02-16
