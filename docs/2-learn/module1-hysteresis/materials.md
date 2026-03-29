# Ferroelectric Materials Physics Reference

This document explains all physical parameters in `config/physics.yaml` for ferroelectric compute-in-memory (FeCIM) simulation.

> **Note:** Parameters here are simulation defaults and literature-inspired ranges. They are not validated hardware specs.

## Table of Contents

1. [Fundamental Constants](#1-fundamental-constants)
2. [Polarization Parameters](#2-polarization-parameters)
3. [Electric Field Parameters](#3-electric-field-parameters)
4. [Dielectric Properties](#4-dielectric-properties)
5. [Film Geometry](#5-film-geometry)
6. [Switching Dynamics](#6-switching-dynamics)
7. [Temperature Properties](#7-temperature-properties)
8. [Reliability Parameters](#8-reliability-parameters)
9. [Crossbar Array Parameters](#9-crossbar-array-parameters)
10. [Energy Parameters](#10-energy-parameters)
11. [Preisach Model Parameters](#11-preisach-model-parameters)
12. [Material Comparison](#12-material-comparison)

---

## Model Parameter Usage

Parameters are used by different hysteresis models. Legend:
- **[P]** = Preisach Model (hysteron-based, phenomenological)
- **[L]** = Landau Model (thermodynamic, mean-field theory)
- **[B]** = Both models

### Quick Reference Matrix

| Parameter | Preisach | Landau | Description |
|-----------|:--------:|:------:|-------------|
| `pr_c_m2` | **[P]** | **[L]** | Remanent polarization |
| `ps_c_m2` | **[P]** | **[L]** | Saturation polarization |
| `ec_v_m` | **[P]** | **[L]** | Coercive field |
| `epsilon_hf` | - | **[L]** | High-frequency permittivity |
| `epsilon_lf` | - | **[L]** | Low-frequency permittivity |
| `thickness_m` | - | **[L]** | Film thickness (depolarization) |
| `tau_s` | **[P]** | - | Switching time constant |
| `tau0_s` | **[P]** | - | Attempt frequency inverse |
| `activation_energy_ev` | **[P]** | - | Switching activation energy |
| `kai_exponent` | **[P]** | - | KAI domain growth exponent |
| `curie_temp_k` | - | **[L]** | Curie temperature (Tc) |
| `temp_coeff_ec` | **[P]** | **[L]** | dEc/dT |
| `temp_coeff_pr` | **[P]** | **[L]** | dPr/dT |
| `alpha_sigma_ratio` | **[P]** | - | Hysteron distribution width |
| `beta_sigma_ratio` | **[P]** | - | Hysteron distribution width |
| `correlation` | **[P]** | - | α-β correlation |
| `fatigue_rate` | **[P]** | - | Degradation per cycle |
| `wakeup_cycles` | **[P]** | - | Wake-up effect |

### Preisach Model Overview
The **Preisach model** represents ferroelectric hysteresis as a weighted sum of elementary bistable units (hysterons). Each hysteron has two switching fields (α, β) and switches between ±1 states.

```
P(E) = ∫∫ μ(α,β) γ(α,β)[E] dα dβ
```

**Key parameters:** Pr, Ps, Ec, τ, τ₀, Ea, KAI exponent, hysteron distribution

### Landau Model Overview
The **Landau-Devonshire model** uses thermodynamic free energy expansion to describe ferroelectric phase transitions:

```
F = α₀(T-Tc)P² + βP⁴ + γP⁶ - EP
```

Where α₀, β, γ are Landau coefficients derived from material properties.

**Key parameters:** Pr, Ps, Ec, ε, Tc, thickness (for depolarization field)

---

## 1. Fundamental Constants

### `fecim_levels` (30)
**Physical Meaning:** Number of discrete analog conductance states per memory cell.

Each ferroelectric cell can be programmed to one of 30 distinct polarization states, corresponding to 30 different threshold voltages (VT) in a FeFET or 30 different tunneling resistances in an FTJ.

**Information Density:**
```
bits_per_cell = log₂(30) = 4.91 bits/cell
```

Compare to:
- NAND Flash MLC: 2 bits/cell (4 levels)
- NAND Flash TLC: 3 bits/cell (8 levels)
- FeCIM: 4.91 bits/cell (30-level baseline; simulation baseline)

### `boltzmann_ev` (8.617×10⁻⁵ eV/K)
**Physical Meaning:** Boltzmann constant in electron-volt units.

Used in temperature-dependent calculations:
```
kT at 300K = 8.617×10⁻⁵ × 300 = 0.0259 eV ≈ 26 meV
```

This thermal energy determines:
- Switching probability at given field
- Retention stability (states must be >> kT apart)
- Temperature dependence of all properties

### `epsilon_0` (8.854×10⁻¹² F/m)
**Physical Meaning:** Vacuum permittivity (electric constant).

Fundamental constant relating electric field to charge displacement:
```
D = ε₀εᵣE + P
```
Where D is displacement, E is electric field, P is polarization.

---

## 2. Polarization Parameters

### `pr_c_m2` - Remanent Polarization (Pr) **[P][L]**
**Unit:** C/m² (Coulombs per square meter)
**Typical Range:** 0.15 - 1.50 C/m² (15 - 150 µC/cm²)
**Used by:** Preisach (hysteron saturation), Landau (equilibrium state)

**Physical Meaning:** The polarization that remains after the external electric field is removed.

```
         P (Polarization)
         ↑
    Ps ──┼────────●
         │       ╱│
    Pr ──┼──────●─┼── ← Remanent polarization (field = 0)
         │     ╱  │
         │    ╱   │
   ─────┼───╱────┼────→ E (Electric Field)
        │  ╱     │
   -Pr ─┼─●──────┼──
        │╱       │
   -Ps ─●────────┼──
```

**Why It Matters:**
- **Memory:** Pr determines the signal difference between "0" and "1" states
- **Analog levels:** Higher Pr allows more distinguishable intermediate states (e.g., partial-switching FeFET demonstrations in Jerry et al., 2017; Pešić et al., *Adv. Funct. Mater.* 2016)
- **Read margin:** ΔQ = 2 × Pr × Area determines charge difference

**Material Values:**
| Material | Pr (µC/cm²) | Notes |
|----------|-------------|-------|
| Standard HZO | 25 | Baseline |
| FeCIM HZO | 30 | Estimated |
| Literature Superlattice | 22 | IEEE 10787441 2024; revised from Cheema 2020 |
| Cryogenic HZO (4K) | 75 | Enhanced at cryo |
| AlScN | 120 | Very high |

### `ps_c_m2` - Saturation Polarization (Ps) **[P][L]**
**Unit:** C/m²
**Typical Range:** 0.25 - 1.50 C/m²
**Used by:** Preisach (maximum hysteron sum), Landau (P⁶ term coefficient)

**Physical Meaning:** Maximum polarization when all ferroelectric domains are aligned.

```
Ps = Pr + reversible component
```

The ratio Pr/Ps indicates:
- **Pr/Ps ≈ 1:** Square hysteresis loop (ideal for binary memory)
- **Pr/Ps < 1:** Slanted loop (better for analog/multi-level)

**Physical Origin:**
Ps is determined by the crystal structure:
```
Ps = (ionic charge × displacement) / unit cell volume
```

For HfO₂ in orthorhombic Pca2₁ phase:
- Oxygen ions displace ~0.5 Å from centrosymmetric positions
- Creates spontaneous dipole moment per unit cell

---

## 3. Electric Field Parameters

### `ec_v_m` - Coercive Field (Ec) **[P][L]**
**Unit:** V/m (Volts per meter)
**Typical Range:** 0.5×10⁸ - 6×10⁸ V/m (0.5 - 6 MV/cm)
**Used by:** Preisach (hysteron switching thresholds), Landau (derived from α,β coefficients)

**Physical Meaning:** The electric field required to switch polarization direction (reduce P to zero).

```
         P
         ↑
         │    ╱───●
         │   ╱    │
         │  ╱     │
   ──────┼─●──────┼────→ E
         │ ↑      │
         │ Ec     │
         │        │
```

**Why It Matters:**
- **Write voltage:** V_write = Ec × thickness
- **Read disturb:** V_read must be << Ec × thickness
- **Power consumption:** Higher Ec = higher write energy
- **State granularity:** Very high Ec (like AlScN) limits analog levels

**Coercive Voltage Calculation:**
```
V_coercive = Ec × t

Example (HZO, 10nm):
V_coercive = 1.0×10⁸ V/m × 10×10⁻⁹ m = 1.0 V
```

**Material Values:**
| Material | Ec (MV/cm) | V_coercive (10nm) |
|----------|------------|-------------------|
| Standard HZO | 1.2 | 1.2 V |
| Literature Superlattice | 0.85 | 0.85 V |
| Cryogenic HZO | 1.5 | 1.5 V |
| AlScN | 5.0 | 10 V (20nm film) |

### `memory_window_v` - Memory Window
**Unit:** Volts
**Typical Range:** 1 - 15 V

**Physical Meaning:** Voltage difference between the two stable polarization states as measured in a FeFET threshold voltage shift.

```
Memory Window = ΔVT = VT(+Pr) - VT(-Pr)
```

Larger memory window enables:
- More distinguishable analog levels
- Better noise immunity
- Longer retention

---

## 4. Dielectric Properties

### `epsilon_hf` - High-Frequency Permittivity **[L]**
**Unit:** Dimensionless (relative to ε₀)
**Typical Range:** 10 - 50
**Used by:** Landau (background permittivity in free energy)

**Physical Meaning:** Dielectric constant at frequencies above ferroelectric domain switching (typically > 1 GHz).

At high frequencies, only electronic and ionic polarization respond:
```
ε_hf = ε_electronic + ε_ionic
```

### `epsilon_lf` - Low-Frequency Permittivity **[L]**
**Unit:** Dimensionless
**Typical Range:** 20 - 100
**Used by:** Landau (total permittivity including domain contribution)

**Physical Meaning:** Dielectric constant at low frequencies where domain walls can move.

```
ε_lf = ε_hf + ε_domain_wall
```

The difference (ε_lf - ε_hf) represents domain wall contribution.

**Capacitance Calculation:**
```
C = ε₀ × ε_r × A / t

Example (HZO, ε=30, A=100nm², t=10nm):
C = 8.854×10⁻¹² × 30 × 100×10⁻¹⁸ / 10×10⁻⁹
C = 2.66×10⁻¹⁷ F = 26.6 aF
```

### `loss_tangent` (tan δ)
**Unit:** Dimensionless
**Typical Range:** 0.005 - 0.03

**Physical Meaning:** Ratio of energy dissipated to energy stored per cycle.

```
tan δ = ε'' / ε' = Power_dissipated / (ω × Energy_stored)
```

**Why It Matters:**
- **Heat generation:** Higher tan δ = more heat during operation
- **Signal integrity:** Loss reduces signal amplitude
- **Efficiency:** Low tan δ desired for energy-efficient CIM

---

## 5. Film Geometry

### `thickness_m` - Film Thickness **[L]**
**Unit:** meters
**Typical Range:** 4 - 20 nm
**Used by:** Landau (depolarization field Ed = -P/(ε₀εᵣ) for thin films)

**Physical Meaning:** Thickness of the ferroelectric layer.

**Critical Relationships:**
```
V_coercive = Ec × thickness       (Write voltage)
C ∝ 1/thickness                   (Capacitance)
Tunneling ∝ exp(-thickness)       (For FTJ)
```

**Thickness Trade-offs:**
| Thinner | Thicker |
|---------|---------|
| Lower write voltage | Higher write voltage |
| Higher capacitance | Lower capacitance |
| Better tunneling (FTJ) | More stable |
| Depolarization field issues | Less surface effects |

**FTJ Special Case:**
For ferroelectric tunnel junctions, ultrathin films (4-5 nm) enable quantum tunneling:
```
J_tunnel ∝ exp(-2κt)
where κ = √(2m*Φ)/ℏ
```

### `area_m2` - Cell Area
**Unit:** m²
**Typical Range:** 10⁻¹⁵ - 10⁻¹⁰ m² (1 nm² - 0.1 µm²)

**Physical Meaning:** Active area of one memory cell.

**Scaling Relations:**
```
Charge = Pr × Area
Capacitance ∝ Area
Current = J × Area
```

### `cell_pitch_nm` - Cell Pitch
**Unit:** nanometers
**Typical Range:** 20 - 100 nm

**Physical Meaning:** Center-to-center spacing between adjacent cells.

```
cell_pitch = √Area + isolation_gap
```

**Density Calculation:**
```
Cells/mm² = 10¹² / (pitch_nm)²

Example (45nm pitch):
Cells/mm² = 10¹² / 45² = 494 million cells/mm²
```

---

## 6. Switching Dynamics

### `tau_s` - Switching Time Constant **[P]**
**Unit:** seconds
**Typical Range:** 0.3 - 100 ns
**Used by:** Preisach (time-dependent hysteron switching, KAI model)

**Physical Meaning:** Characteristic time for polarization reversal.

**Nucleation-Limited Switching (NLS) Model:**
```
P(t) = Ps × [1 - exp(-(t/τ)^n)]
```

Where n is the dimensionality of domain growth.

### `tau0_s` - Attempt Time **[P]**
**Unit:** seconds
**Typical Value:** ~10⁻¹³ s (100 fs)
**Used by:** Preisach (Arrhenius prefactor for thermally-activated switching)

**Physical Meaning:** Inverse of attempt frequency for domain nucleation.

From transition state theory:
```
τ₀ = h / (kT) ≈ 10⁻¹³ s at room temperature
```

### `activation_energy_ev` - Activation Energy (Ea) **[P]**
**Unit:** electron-volts
**Typical Range:** 0.3 - 1.2 eV
**Used by:** Preisach (thermally-activated switching: τ = τ₀·exp(Ea/kT))

**Physical Meaning:** Energy barrier for domain nucleation/switching.

**Arrhenius Relationship:**
```
τ = τ₀ × exp(Ea / kT)

Example (Ea = 0.7 eV, T = 300K):
τ = 10⁻¹³ × exp(0.7 / 0.0259) = 10⁻¹³ × 10¹¹·⁷ = 5×10⁻² s
```

**Field-Dependent Activation:**
```
Ea(E) = Ea₀ × (1 - E/E₀)^α

where α ≈ 1.5-2 (Merz's law)
```

### `kai_exponent` - KAI Model Exponent **[P]**
**Unit:** Dimensionless
**Typical Range:** 1.5 - 3.0
**Used by:** Preisach (domain growth dimensionality in P(t) = Ps·[1-exp(-(t/τ)ⁿ)])

**Physical Meaning:** Dimensionality of domain growth in the Kolmogorov-Avrami-Ishibashi (KAI) model.

```
P(t) = Ps × [1 - exp(-(t/τ)^n)]

n = 1: 1D needle-like growth
n = 2: 2D circular domain expansion (typical for thin films)
n = 3: 3D spherical growth
```

**For HZO thin films:** n ≈ 2.0-2.5 (2D growth dominates)

---

## 7. Temperature Properties

### `curie_temp_k` - Curie Temperature (Tc) **[L]**
**Unit:** Kelvin
**Typical Range:** 700 - 1300 K
**Used by:** Landau (phase transition: α = α₀(T-Tc) in free energy)

**Physical Meaning:** Temperature above which ferroelectricity disappears (phase transition to paraelectric).

**Landau Theory:**
```
P² ∝ (Tc - T)    for T < Tc
P = 0             for T > Tc
```

**Material Values:**
| Material | Tc (K) | Tc (°C) |
|----------|--------|---------|
| HZO | 723 | 450 |
| Literature Superlattice | 773 | 500 |
| AlScN | 1273 | 1000 |

Higher Tc provides:
- Better thermal stability
- BEOL (back-end-of-line) compatibility
- Wider operating temperature range

### `temp_coeff_ec` - Temperature Coefficient of Ec **[P][L]**
**Unit:** V/m/K
**Typical Value:** -1.5×10⁵ to -2.5×10⁵ V/m/K
**Used by:** Preisach (T-dependent hysteron thresholds), Landau (T-dependent Ec from α(T))

**Physical Meaning:** How coercive field changes with temperature.

```
Ec(T) = Ec(T₀) + (dEc/dT) × (T - T₀)
```

**Negative coefficient:** Ec decreases as temperature increases (easier switching).

### `temp_coeff_pr` - Temperature Coefficient of Pr **[P][L]**
**Unit:** C/m²/K
**Typical Value:** -3×10⁻⁵ to -5×10⁻⁵ C/m²/K
**Used by:** Preisach (T-dependent hysteron saturation), Landau (Pr(T) = Pr₀·(1-T/Tc)^β)

**Physical Meaning:** How remanent polarization changes with temperature.

```
Pr(T) = Pr(T₀) × (1 - T/Tc)^β
```

Where β ≈ 0.5 (mean-field theory).

---

## 8. Reliability Parameters

### `endurance_cycles` - Endurance Limit
**Unit:** Cycles
**Typical Range:** 10⁶ - 10¹²

**Physical Meaning:** Number of write cycles before significant degradation.

**Degradation Mechanism:**
Repeated switching creates:
- Oxygen vacancy accumulation at interfaces
- Charge trapping in defects
- Domain pinning sites

**Stretched Exponential Model:**
```
Pr(N) = Pr₀ × exp[-(N/N₀)^β]

where:
  N = cycle count
  N₀ = endurance limit
  β ≈ 0.3 (typical)
```

**Material Comparison:**
| Material | Endurance | Notes |
|----------|-----------|-------|
| NAND Flash | 10³-10⁴ | Very limited |
| Standard HZO | 10¹⁰ | Verified |
| FeCIM (demonstrated) | 10⁹ | Conservative |
| FeCIM (target) | 10¹² | Aspirational |
| V:HfO₂ | 10¹² | Vanadium-doped |

### `retention_time_s` - Retention Time
**Unit:** seconds
**Typical Range:** 10⁷ - 10¹⁰ s (months to centuries)

**Physical Meaning:** Time data remains valid without refresh.

**Retention Loss Mechanisms:**
1. **Depolarization field:** Internal field opposing polarization
2. **Thermal fluctuation:** Random domain switching
3. **Imprint:** Preferred state from defect alignment

**Temperature Acceleration:**
```
t_retention(T) = t_retention(T_ref) × exp[Ea/k × (1/T - 1/T_ref)]
```

**Common Targets:**
- Consumer: 10 years at 85°C
- Automotive: 10 years at 150°C
- Cryogenic: >100 years at 4K

### `imprint_field_v_m` - Imprint Field
**Unit:** V/m
**Typical Range:** 0.1×10⁶ - 2×10⁶ V/m

**Physical Meaning:** Built-in field that shifts the hysteresis loop.

```
         P
         ↑
         │     ╱──● ← Shifted loop
         │    ╱   │
   ──────┼───╱────┼────→ E
         │  ╱  ↑  │
         │ ╱   │  │
         │╱    Imprint shift
```

**Causes:**
- Charge injection into interfaces
- Defect dipole alignment
- Asymmetric electrodes

**Effect:** One polarization state becomes preferred, reducing memory window.

---

## 9. Crossbar Array Parameters

### Conductance Parameters

**`conductance_min_s` (Gmin) and `conductance_max_s` (Gmax)**
**Unit:** Siemens (S = 1/Ω)
**Typical Range:** 1 µS - 100 µS

**Physical Meaning:** The conductance range mapped to weight values in neural network computation.

```
Weight mapping:
w = 0 → G = Gmin = 1 µS
w = 1 → G = Gmax = 100 µS

Analog levels:
G_level = Gmin + (level/29) × (Gmax - Gmin)
```

**`conductance_ratio` (Gmax/Gmin)**
**Typical Value:** 100

Higher ratio enables:
- Better weight resolution
- Larger dynamic range
- More distinguishable analog levels

### `ter_ratio` - Tunneling Electroresistance Ratio
**Unit:** Percentage
**Typical Range:** 100% - 10,000%

**Physical Meaning:** Resistance ratio between two polarization states in FTJ.

```
TER = (R_high - R_low) / R_low × 100%

Example: TER = 911% means R_high/R_low ≈ 10
```

### Non-Ideality Parameters

**`device_variation`** (σ = 5%)
Device-to-device variation in conductance due to:
- Film thickness variation
- Grain size distribution
- Interface roughness

**`read_noise`** (σ = 2%)
Cycle-to-cycle variation in read current due to:
- Thermal noise
- 1/f noise
- Random telegraph noise

**`write_noise`** (σ = 3%)
Variation in programmed state due to:
- Stochastic domain nucleation
- Pulse timing jitter

### Line Resistance

**`word_line_resistance_ohm` and `bit_line_resistance_ohm`**
**Unit:** Ohms per segment
**Typical Value:** 10 Ω

**Physical Meaning:** Metal line resistance causing IR drop.

**IR Drop Effect:**
```
V_cell(i,j) = V_applied - I × R_line × (i + j)
```

Cells farther from drivers see reduced voltage, causing:
- Non-uniform writing
- Reduced read margin
- Array size limitations

### Sneak Path Parameters

**`sneak_path_enabled`**
In resistive crossbar arrays, current can flow through unselected cells:

```
    Selected cell
         ↓
    ●────●────●
    │    │    │
    ●────⊗────●  ← Sneak current paths
    │    │    │
    ●────●────●
```

**`sneak_conductance_ratio`** (0.01)
Ratio of leakage current to intended current.

**`fcm_sneak_free`** (true)
Ferroelectric capacitor-based arrays avoid DC sneak paths because capacitors block DC current.

---

## 10. Energy Parameters

### Per-Operation Energy

**`read_energy_j`** (~1 fJ)
Energy to sense the state of one cell.

**`write_energy_j`** (~10 fJ)
Energy to program one cell:
```
E_write ≈ C × V² = 2 × Pr × Ec × Volume
```

**`mac_energy_j`** (~5 fJ)
Energy for one multiply-accumulate operation:
```
E_MAC = E_read + E_accumulate
```

### Energy Comparison

| Operation | FeCIM | NAND | DRAM | SRAM |
|-----------|-------|------|------|------|
| Write | 10 fJ | 500 fJ | 5 pJ | 50 fJ |
| Read | 1 fJ | 10 fJ | 5 pJ | 5 fJ |
| Advantage | **1×** | 50× worse | 1000× worse | 5× worse |

### CIM Efficiency Metrics

**`cim_area_efficiency_gops_mm2`** (160 GOPS/mm²)
Giga-operations per second per square millimeter.

**`cim_energy_efficiency_tops_w`** (25.5 TOPS/W)
Tera-operations per second per Watt.

Compare to:
- GPU: ~1 TOPS/W
- TPU: ~5 TOPS/W
- FeCIM: 25.5 TOPS/W (5× better than TPU)

---

## 11. Preisach Model Parameters

The Preisach model represents ferroelectric hysteresis as a collection of elementary bistable units (hysterons).

### `grid_size` (30) **[P]**
**Physical Meaning:** Number of hysterons along each axis of the Preisach plane.

Total hysterons = grid_size² = 900

Higher grid size = smoother hysteresis loop simulation.

### Distribution Parameters **[P]**

**`alpha_sigma_ratio`** (0.20) and **`beta_sigma_ratio`** (0.20)
**Physical Meaning:** Width of the hysteron distribution relative to Ec.
**Used by:** Preisach only (defines Gaussian distribution of switching fields)

```
σ_α = 0.20 × Ec
σ_β = 0.20 × Ec
```

Wider distribution = more gradual switching = better analog behavior.

**`correlation`** (0.5)
Correlation between switching-up (α) and switching-down (β) fields.

```
ρ = 0: Independent switching fields
ρ = 1: Perfectly correlated (α = β)
```

### Fatigue Model **[P]**

**`fatigue_rate`** (10⁻¹⁰)
Degradation of hysteron amplitude per cycle.
**Used by:** Preisach only (reduces hysteron weights with cycling)

**`wakeup_cycles`** (100)
Number of cycles for wake-up effect (initial Pr increase).

**`initial_wakeup`** (0.8)
Initial Pr as fraction of final Pr before wake-up completes.

Wake-up occurs because:
- Field cycling redistributes defects
- Domain structure stabilizes
- Interface charges equilibrate

---

## 12. Material Comparison

### CMOS-Compatible Materials Summary

| Parameter | Standard HZO | FeCIM HZO | Superlattice | Cryogenic | HZO-32 | FTJ-140 | AlScN |
|-----------|-------------|-----------|--------------|-----------|--------|---------|-------|
| **Pr (µC/cm²)** | 25 | 30 | 50 | 75 | 20 | 25 | 120 |
| **Ec (MV/cm)** | 1.2 | 1.0 | 0.85 | 1.5 | 1.0 | 1.2 | 5.0 |
| **Thickness (nm)** | 10 | 10 | 10 | 10 | 10 | 4.5 | 20 |
| **τ (ns)** | 1 | 10 | 0.36 | 1 | 10 | 20 | 10 |
| **Endurance** | 10¹⁰ | 10⁹ | 10¹⁰ | 10⁹ | 10⁸ | 10⁷ | 10⁹ |
| **Analog States** | ~30 | 30 | 30+ | 30 | 32 | 140 | 8-16 |
| **Tc (°C)** | 450 | 450 | 500 | 450 | 450 | 450 | 1000 |

### Selection Guide

**For maximum analog states:** HZO FTJ (140 states)
**For best endurance:** Standard HZO or Superlattice (10¹⁰)
**For lowest voltage:** Literature Superlattice (0.85 MV/cm)
**For quantum computing:** Cryogenic HZO (enhanced Pr at 4K)
**For high temperature:** AlScN (Tc > 1000°C)
**For FeCIM simulation:** FeCIM HZO (demonstrated values)

---

## References

1. Park et al., Adv. Mater. 27, 1811 (2015) - HZO ferroelectricity discovery
2. Cheema et al., Nature 580, 478 (2020) - Superlattice enhancement
3. Nature Commun. 2025, doi:10.1038/s41467-025-61758-2 - Epitaxial stability
4. ACS Nano 2024, doi:10.1021/acsnano.4c01992 - 4nm HZO endurance
5. ACS Appl. Nano Mater. 2024, doi:10.1021/acsanm.4c04974 - 50µC/cm² Pr
6. Purdue Thesis 2024 - Sub-ns switching (360ps record)
7. Adv. Electron. Mater. 2024, doi:10.1002/aelm.202300879 - Cryogenic HZO
8. Oh et al., IEEE EDL 38(6), 732 (2017) - 32 analog states
9. Song et al., Adv. Science 2024, doi:10.1002/advs.202308588 - 140 states FTJ
10. Nature Commun. 2025, doi:10.1038/s41467-025-62904-6 - AlScN
