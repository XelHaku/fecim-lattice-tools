# CIM Equation Quick Reference

> **Note:** This document contains reported values and illustrative calculations. It is not a verification source. See `docs/comparison/HONESTY_AUDIT.md`.


**Complete collection of key equations for Compute-in-Memory (CIM) systems.**

Last Updated: January 2026

---

## Table of Contents

1. [Basic CIM Operations](#1-basic-cim-operations)
2. [IR Drop](#2-ir-drop)
3. [Sneak Path](#3-sneak-path)
4. [Quantization](#4-quantization)
5. [Noise Models](#5-noise-models)
6. [Energy](#6-energy)
7. [Accuracy Formulas](#7-accuracy-formulas)
8. [Device Parameters (HfO₂-ZrO₂ FeFET)](#8-device-parameters-hfo%E2%82%82-zro%E2%82%82-fefet)

---

## 1. Basic CIM Operations

### 1.1 Single Cell (Ohm's Law)

**Equation:**
```
I = G × V
```

**Variables:**
- `I`: Output current (A)
- `G`: Conductance (S), typically 1-100 µS
- `V`: Input voltage (V), typically 0.1-0.5V for read operations

**Source:** Fundamental physics (Ohm's Law)

---

### 1.2 Row Output (Kirchhoff's Current Law)

**Equation:**
```
I_i = Σⱼ (G_ij × V_j)
```

**Variables:**
- `I_i`: Current collected from row i
- `G_ij`: Conductance at crosspoint (i,j)
- `V_j`: Voltage applied to column j
- Sum over all columns j

**Source:** Kirchhoff's Current Law (parallel currents sum)

**Physical Meaning:** Each crosspoint multiplies G×V, currents sum on the row wire. This IS matrix-vector multiplication happening in physics.

---

### 1.3 Matrix Form

**Equation:**
```
I = G × V
```

**Variables:**
- `I`: Output current vector (m×1)
- `G`: Conductance matrix (m×n)
- `V`: Input voltage vector (n×1)

**Source:** Linear algebra representation of physical MVM

**Example (2×2):**
```
[I₀]   [G₀₀  G₀₁]   [V₀]
[I₁] = [G₁₀  G₁₁] × [V₁]
```

---

### 1.4 Differential Pair (Signed Weights)

**Equation:**
```
W = G⁺ - G⁻
```

**Variables:**
- `W`: Effective weight (can be positive or negative)
- `G⁺`: Conductance of positive cell
- `G⁻`: Conductance of negative cell

**Source:** Standard CIM architecture for bipolar weights

**Implementation:** Two physical crossbars or interleaved columns.

**Example:**
- W = +0.5 → G⁺ = 0.75, G⁻ = 0.25
- W = -0.5 → G⁺ = 0.25, G⁻ = 0.75

---

## 2. IR Drop

### 2.1 Voltage Distribution

**Equation:**
```
V_ij = V_applied - Σ(I × R_wire × distance)
```

**Variables:**
- `V_ij`: Actual voltage at crosspoint (i,j)
- `V_applied`: Voltage at wire entry point
- `I`: Current flowing through wire segment
- `R_wire`: Wire resistance per unit length, typically **2.5Ω per cell pitch** (45nm node)
- `distance`: Cumulative wire length from source

**Source:** Kirchhoff's Voltage Law

**Iterative Solution:** Voltage depends on current, current depends on voltage → solve with relaxation method.

---

### 2.2 Effective Conductance Compensation

**Equation:**
```
G_effective = G_stored × (V_applied / V_actual)
```

**Variables:**
- `G_effective`: Apparent conductance seen by readout circuit
- `G_stored`: Programmed conductance value
- `V_actual`: Actual voltage at cell (reduced by IR drop)

**Source:** IEEE papers on IR drop mitigation

**Example:**
- Stored: G = 50 µS
- V_applied = 0.5V, V_actual = 0.45V (10% IR drop)
- G_effective = 50 × (0.5/0.45) = 55.6 µS (11% error)

---

### 2.3 IR Drop Impact (Empirical)

**Source:** IEEE TCAD 2022 (from crossbar.research.md)

| Array Size | Max IR Drop | Accuracy Loss |
|------------|-------------|---------------|
| 64×64      | 5%          | 0.3%          |
| 128×128    | 12%         | 1.2%          |
| 256×256    | 22%         | 3.5%          |
| 512×512    | 35%         | 8%+           |

**Takeaway:** IR drop limits practical arrays to ~256×256 without compensation.

---

## 3. Sneak Path

### 3.1 Three-Cell Series Conductance

**Equation:**
```
G_sneak = 1 / (1/G₁ + 1/G₂ + 1/G₃)
```

**Variables:**
- `G_sneak`: Effective conductance of sneak path (three cells in series)
- `G₁, G₂, G₃`: Conductances of cells in sneak path

**Source:** Series resistance model

**Physical Meaning:** Current flows through three unintended cells to reach target row.

**Example:**
- Three cells: G₁ = G₂ = G₃ = 50 µS
- G_sneak = 1/(1/50 + 1/50 + 1/50) = 16.7 µS (33% of single cell)

---

### 3.2 Sneak Current

**Equation:**
```
I_sneak = V × G_sneak
```

**Variables:**
- `I_sneak`: Parasitic current from sneak path
- `V`: Applied read voltage
- `G_sneak`: Sneak path conductance (from 3.1)

**Source:** Ohm's Law applied to sneak path

**Total current error:**
```
I_total = I_signal + I_sneak
```

Sneak current adds to signal, causing **overestimation** of conductance.

---

### 3.3 Signal-to-Noise Ratio

**Equation:**
```
SNR = I_signal / I_sneak
```

**Variables:**
- `SNR`: Signal-to-noise ratio
- `I_signal`: Intended current from target cell
- `I_sneak`: Total parasitic sneak current

**Target:** SNR > 10 for reliable read (10:1 ratio)

**Source:** ReRAM/memristor literature

**Mitigation strategies:**
| Architecture | Sneak Path Error | SNR        |
|--------------|------------------|------------|
| Passive (0T1R) | 5-20%          | 5-20       |
| 1T1R         | ~0%              | ∞          |
| 1S1R (selector) | 0.5-2%       | 50-200     |
| Self-rectifying | 1-5%         | 20-100     |

---

### 3.4 Nonlinearity Parameter

**Equation:**
```
k = I(V_op) / I(V_op/2)
```

**Variables:**
- `k`: Nonlinearity coefficient
- `I(V_op)`: Current at operating voltage V_op
- `I(V_op/2)`: Current at half operating voltage

**Requirement:** k > 10 for large arrays (suppresses sneak paths)

**Source:** Device nonlinearity studies

**Ideal device:** k = 2 (linear, Ohm's Law)
**Self-rectifying FeFET:** k = 10-100 (asymmetric I-V curve)

---

## 4. Quantization

### 4.1 Linear N-Level Quantization

**Equation:**
```
G_level = G_min + (level/(N-1)) × (G_max - G_min)
```

**Variables:**
- `G_level`: Conductance for quantization level
- `level`: Integer level index (0, 1, 2, ..., N-1)
- `N`: Number of discrete levels (30 for demo baseline; demo baseline (configurable))
- `G_min`: Minimum conductance (typically 1 µS)
- `G_max`: Maximum conductance (typically 100 µS)

**Source:** Standard quantization

**Example (30 levels; demo baseline):**
- Level 0: G = 1 µS
- Level 15: G = 51.5 µS (midpoint)
- Level 29: G = 100 µS

---

### 4.2 Bits per Cell

**Equation:**
```
bits = log₂(N)
```

**Variables:**
- `bits`: Information capacity per cell
- `N`: Number of discrete distinguishable levels

**Examples:**
- 2 levels (binary) = 1 bit/cell
- 16 levels = 4 bits/cell
- 30 levels = **4.9 bits/cell** (FeCIM demo baseline; demo baseline (configurable))
- 32 levels = 5 bits/cell

**Source:** Shannon information theory

---

### 4.3 Quantization Error

**Equation:**
```
error = |W_original - W_quantized|

Mean_error ≈ (G_max - G_min) / (2 × N)
```

**Variables:**
- `W_original`: Desired continuous weight value
- `W_quantized`: Nearest quantized level
- `Mean_error`: Average quantization error

**Source:** Quantization theory

**Example (30 levels; demo baseline, G_range = 99 µS):**
```
Mean_error = 99 / (2 × 30) = 1.65 µS
```

**Typical accuracy impact:** 0.5-1% for neural networks with 30 levels (demo baseline).

---

## 5. Noise Models

### 5.1 Thermal Noise (Johnson-Nyquist)

**Equation:**
```
σ_thermal = √(4 × k_B × T × R × BW)
```

**Constants & Variables:**
- `k_B` = 1.38 × 10⁻²³ J/K (Boltzmann constant)
- `T` = 300 K (room temperature, 27°C)
- `R` = Resistance (Ω)
- `BW` = Bandwidth (Hz), typically 1 MHz - 1 GHz
- `σ_thermal` = RMS noise voltage (V)

**Source:** Fundamental physics (thermal fluctuations)

**Example (R = 10kΩ, BW = 1 MHz):**
```
σ_thermal = √(4 × 1.38e-23 × 300 × 10000 × 1e6) = 12.9 µV
```

---

### 5.2 Shot Noise

**Equation:**
```
σ_shot = √(2 × q × I × BW)
```

**Constants & Variables:**
- `q` = 1.6 × 10⁻¹⁹ C (elementary charge)
- `I` = DC current (A)
- `BW` = Bandwidth (Hz)
- `σ_shot` = RMS noise current (A)

**Source:** Fundamental physics (discrete charge carriers)

**Example (I = 1 µA, BW = 1 MHz):**
```
σ_shot = √(2 × 1.6e-19 × 1e-6 × 1e6) = 17.9 pA
```

**Dominant when:** I is small (sub-µA currents)

---

### 5.3 Combined Read Noise

**Equation:**
```
σ_read = √(σ_thermal² + σ_shot²)
```

**Source:** Independent noise sources add in quadrature

**Typical values for FeCIM read:**
- Thermal noise: 10-50 µV (sense amplifier)
- Shot noise: 10-100 pA (low read currents)
- Combined SNR: 30-40 dB

---

### 5.4 FeFET Drift Model

**Equation:**
```
G(t) = G₀ × (1 + ν × log(t/t₀))
```

**Variables:**
- `G(t)`: Conductance at time t after programming
- `G₀`: Initial conductance immediately after programming
- `ν`: Drift coefficient (technology-dependent)
- `t`: Time elapsed since programming (seconds)
- `t₀`: Reference time, typically 1 second

**Source:** Drift literature (empirical model)

**Note:** The 0.001 value for FeFET is assumed for simulation purposes. No reported in literature source exists for FeFET-specific drift coefficients. The value is qualitatively expected to be lower than RRAM/PCM based on retention characteristics.

**Drift coefficients (from IEEE TCAD 2022 for RRAM/PCM, FeFET assumed):**
| Memory Type | ν (drift coefficient) | 10-Year Retention | Source |
|-------------|----------------------|-------------------|--------|
| PCM         | 0.05-0.1             | 90-95%            | Literature |
| RRAM        | 0.01-0.05            | 95-99%            | Literature |
| **FeFET**   | **0.001-0.01**       | **99%+**          | **Assumed (no peer review)** |
| Flash       | 0.01-0.02            | 99%+              | Literature |

**Example (FeFET, ν = 0.001):**
- After 1 second: G = G₀ × 1.000 (0% change)
- After 1 hour: G = G₀ × 1.004 (0.4% change)
- After 1 year: G = G₀ × 1.009 (0.9% change)
- After 10 years: G = G₀ × 1.01 (1% change)

**Takeaway:** FeFET drift is negligible compared to other error sources.

---

## 6. Energy

### 6.1 CIM MAC Energy

**Equation:**
```
E_MAC = C × V² + G × V × t
```

**Variables:**
- `E_MAC`: Energy per multiply-accumulate operation (Joules)
- `C`: Parasitic capacitance (~fF = 10⁻¹⁵ F)
- `V`: Operating voltage (~0.5V)
- `G`: Conductance (~µS = 10⁻⁶ S)
- `t`: Operation time (~ns = 10⁻⁹ s)

**Source:** CIM energy analysis papers

**Typical values:**
- First term (capacitance charging): 10-50 fJ
- Second term (resistive dissipation): 5-300 fJ
- **Total: 17 fJ - 350 fJ per MAC**

**Example (C = 40 fF, V = 0.5V, G = 50 µS, t = 10 ns):**
```
E_MAC = (40e-15 × 0.5²) + (50e-6 × 0.5 × 10e-9)
      = 10 fJ + 0.25 fJ = 10.25 fJ
```

---

### 6.2 Digital MAC Energy

**Equation:**
```
E_digital = E_multiply + E_add + E_data_movement
```

**Typical values (32-bit operations at 7nm node):**
- `E_multiply`: ~3.7 pJ (integer multiply)
- `E_add`: ~0.9 pJ (integer add)
- `E_data_movement`: ~200 pJ per DRAM access

**Source:** Sze et al., "Efficient Processing of Deep Neural Networks"

**Total per MAC:**
- Compute-only: 4.6 pJ
- **With data movement: ~200-400 pJ** (memory access dominates!)

---

### 6.3 Energy Ratio (CIM Advantage)

**Equation:**
```
CIM_advantage = E_digital / E_CIM
```

**Typical values:**
- MAC alone: 10-100× (4.6 pJ / 50 fJ = 92×)
- **Including data movement:** illustrative order-of-magnitude example (verify with sources)

**Source:** IBM, Mythic AI, various CIM papers

**Why such huge advantage?**
1. CIM eliminates memory access (zero data movement)
2. Analog computation (no switching energy for gates)
3. Parallel operations (all MACs in 1 cycle)

---

### 6.4 CIM vs GPU Energy Comparison (Empirical)

**Source:** IEEE JSSC 2022, Nature Electronics 2023

| Operation | GPU+HBM | FeCIM | Advantage |
|-----------|---------|-------|-----------|
| Single MAC | 100 pJ | 10-50 fJ | 2000-10000× |
| 64×64 MVM | 640 nJ | 0.6-3 nJ | 200-1000× |
| MNIST inference (full) | 50 mJ | 5 µJ | illustrative |

**Note:** Advantage decreases with smaller batch sizes (peripheral overhead).

---

## 7. Accuracy Formulas

### 7.1 Combined Non-Ideality Impact (Empirical)

**Heuristic formula:**
```
acc_loss ≈ f(q_error, ir_drop, sneak_ratio, noise_σ)
```

**Where:**
- `q_error`: Quantization error (0.5-1% for 30 levels; demo baseline)
- `ir_drop`: IR drop percentage (1-3% for 128×128)
- `sneak_ratio`: Sneak current fraction (0-5% for passive, 0% for 1T1R)
- `noise_σ`: Read noise std deviation (0.5-2%)

**Typical total accuracy loss:** 0.5-1.5% for well-designed systems (128×128, 1T1R, 6-bit ADC)

**Source:** Placeholder benchmarks (update with primary sources)

---

### 7.2 ADC SNR

**Equation:**
```
SNR_ADC = 6.02 × N + 1.76  (dB)
```

**Variables:**
- `N`: ADC bit width
- `SNR_ADC`: Signal-to-noise ratio in decibels

**Source:** ADC theory (quantization noise dominated)

**Examples:**
- 4-bit ADC: SNR = 25.8 dB
- 6-bit ADC: SNR = 37.9 dB
- 8-bit ADC: SNR = 49.9 dB

---

### 7.3 Effective Number of Bits (ENOB)

**Equation:**
```
ENOB = (SNR_measured - 1.76) / 6.02
```

**Variables:**
- `SNR_measured`: Measured SNR including all non-idealities (dB)
- `ENOB`: Effective bit resolution

**Typical degradation:** 0.5-1.5 bits less than nominal ADC bits

**Source:** ADC characterization

**Example (6-bit ADC with SNR = 34 dB):**
```
ENOB = (34 - 1.76) / 6.02 = 5.35 bits
```
Effective resolution ≈ 5 bits despite 6-bit ADC.

---

### 7.4 MNIST Accuracy vs. ADC Bits (Illustrative)

**Source:** IEEE TCAD 2022 (from crossbar.research.md)

| ADC Bits | Relative Power | MNIST Accuracy |
|----------|----------------|----------------|
| 4-bit    | 1×             | 85-90%         |
| 6-bit    | 4×             | 90-95%         |
| 8-bit    | 16×            | 95-97%         |
| 10-bit   | 64×            | 97-98%         |

**Takeaway:** 6-bit ADC is sweet spot for most workloads (good accuracy, reasonable power).

---

## 8. Device Parameters (HfO₂-ZrO₂ FeFET)

### 8.1 Ferroelectric Parameters

| Parameter | Symbol | Value | Source |
|-----------|--------|-------|--------|
| **Remanent polarization** | Pr | 15-34 µC/cm² | Nature Commun. 2025 (PMC12254504) |
| **Coercive field** | Ec | 1.0-1.5 MV/cm | Nature Commun. 2025 (PMC12254504) |
| Film thickness | t_FE | 10-30 nm | HfO₂-ZrO₂ superlattice papers |
| Superlattice period | Λ | 2-5 nm | (HfO₂/ZrO₂ bilayer repeat) |
| Wake-up cycles | - | < 100 | ACS Omega 2024 (superlattice reduces wake-up) |

**Source references:**
- PMC12254504: "Enhancing ferroelectric stability: wide-range of adaptive control in epitaxial HfO₂/ZrO₂ superlattices"
- PubMed 38166401: "HfO₂-ZrO₂ Ferroelectric Capacitors with Superlattice Structure"

---

### 8.2 Device Operation Parameters

| Parameter | Symbol | Value | Source |
|-----------|--------|-------|--------|
| **Conductance range** | G | 1-100 µS | Design parameter (tunable) |
| **Analog states** | N | **30** | **Dr. Tour COSM 2025** |
| Bits per cell | - | 4.9 | log₂(30) |
| **Endurance** | - | **10⁹-10¹²** cycles | Dr. Tour / IEEE IRPS 2022 |
| Retention | - | 10+ years | FeFET literature |
| Read voltage | V_read | 0.1-0.5 V | Typical operation |
| Write voltage | V_write | 1.5-3.0 V | Programming pulses |
| Write time | t_write | 10-100 ns | Fast switching |
| Read time | t_read | 10 ns | Sense amplifier speed |

---

### 8.3 Noise and Variation Parameters

| Parameter | Symbol | Value | Source |
|-----------|--------|-------|--------|
| **Drift coefficient** | ν | **0.001** | **Assumed (no reported in literature source)** |
| Device-to-device variation | σ_D2D | ~3% | HZO studies |
| Cycle-to-cycle variation | σ_C2C | ~1% | FeFET papers |
| Read noise (thermal) | σ_thermal | 0.5% σ/µ | Johnson noise in sense amps |
| Temperature coefficient | - | ±1% per 10°C | Temperature sensitivity |

**Statistical requirement:** 3σ separation between adjacent levels for reliable read.

**Example calculation (30 levels; demo baseline):**
- Total variation: σ_total = √(σ_D2D² + σ_C2C² + σ_read²) = √(3² + 1² + 0.5²) ≈ 3.2%
- Level spacing: 100% / 29 = 3.45% (just meets 3σ ≈ 9.6% separation)

---

### 8.4 Energy Parameters

| Parameter | Value | Source |
|-----------|-------|--------|
| **Read energy** | 0.01-0.1 fJ | FeFET advantage over RRAM |
| **Write energy** | 10-100 fJ | Low polarization switching energy |
| Capacitance | 10-100 fF | Parasitic capacitance (cell + wire) |
| Switching current | 1-10 µA | Displacement current (no Joule heating) |

**Comparison to other technologies:**
| Technology | Read (fJ) | Write (fJ) | Endurance | Multi-level |
|------------|-----------|------------|-----------|-------------|
| FeFET      | 0.01-0.1  | 10-100     | 10¹⁰-10¹² | 4-5 bits    |
| RRAM       | 0.1-1     | 100-1000   | 10⁶-10⁹   | 2-4 bits    |
| PCM        | 1-10      | 1000-10000 | 10⁶-10⁹   | 2-4 bits    |
| SRAM       | 1-10      | 10-100     | 10¹⁶      | 1 bit       |

---

## 9. Key Equations Summary Table

| Category | Most Important Equation | Use Case |
|----------|-------------------------|----------|
| **MVM** | `I = G × V` | Core compute operation |
| **IR Drop** | `V_ij = V_applied - Σ(I × R_wire × dist)` | Voltage degradation modeling |
| **Sneak Path** | `G_sneak = 1/(1/G₁ + 1/G₂ + 1/G₃)` | Parasitic current estimation |
| **Quantization** | `G_level = G_min + (level/29) × (G_max - G_min)` | 30-level programming (demo baseline) |
| **Thermal Noise** | `σ = √(4 × k_B × T × R × BW)` | Read noise floor |
| **Drift** | `G(t) = G₀ × (1 + ν × log(t/t₀))` | Long-term retention |
| **Energy** | `E_MAC = C × V² + G × V × t` | Power estimation |
| **ADC SNR** | `SNR = 6.02 × N + 1.76` | ADC resolution needed |

---

## 10. Design Rules of Thumb

### Quantization
- **30 levels (baseline)**: Minimum for high-accuracy neural networks (demo assumption)
- **16 levels (4-bit)**: Acceptable for many workloads
- **< 8 levels**: Severe accuracy degradation

### Array Size
- **≤ 64×64**: Ideal (minimal IR drop, no compensation needed)
- **128×128**: Good (1-2% accuracy loss from IR drop)
- **256×256**: Practical limit (requires compensation circuits)
- **> 512×512**: Severe IR drop (not recommended without advanced mitigation)

### ADC Bits
- **4-bit**: Lower accuracy (illustrative)
- **6-bit**: **Sweet spot** (90-95% accuracy, 4× power vs 4-bit)
- **8-bit**: Overkill for most CIM workloads (16× power)

### Architecture Choice
- **1T1R**: Industry standard (zero sneak paths, 50% area penalty)
- **1S1R**: High density compromise (1-2% sneak error)
- **Self-rectifying 0T1R**: FeFET advantage (1-5% sneak, highest density)

### Noise Budget
- **Total variation < 3%**: Required for 30-level demo baseline operation
- **Dominant sources**: Device variation (3%), thermal noise (0.5%), drift (< 1%)

### Energy Targets
- **< 100 fJ/MAC**: Competitive with best RRAM/PCM
- **< 50 fJ/MAC**: FeFET advantage (displacement current, low capacitance)
- **< 10 fJ/MAC**: Aggressive target (requires ultra-low voltage, small cells)

---

## References

### Primary Sources
1. **Dr. external research group COSM 2025** - First public disclosure of 30-level FeCIM (docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md)
2. **Nature Communications 2025 (PMC12254504)** - "Enhancing ferroelectric stability: wide-range of adaptive control in epitaxial HfO₂/ZrO₂ superlattices"
3. **IEEE TCAD 2022** - IR drop and sneak path benchmarks (crossbar.research.md)
4. **Sze et al.** - "Efficient Processing of Deep Neural Networks" (digital MAC energy reference)

### Literature Meta-Study
- **Complete analysis:** docs/crossbar/crossbar.research.md
- **40+ papers** synthesized covering CIM architectures, FeFET crossbars, non-idealities

### Project Documentation
- **Physics deep-dive:** docs/crossbar/crossbar.physics.md
- **Testing guide:** docs/development/TESTING.md
- **Fyne GUI reference:** docs/development/FYNE_NOTES.md

---

**Maintained by:** FeCIM Lattice Tools Project
**For:** AI agents, researchers, and developers working with CIM systems
**Last verification:** January 2026 against reported in literature literature
