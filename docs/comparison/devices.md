# Compute-in-Memory (CIM) Device Technologies

> **Note:** This document summarizes reported values from literature. These values are **not verified** by this project. See `docs/comparison/HONESTY_AUDIT.md` for the current verification scope.

Comprehensive comparison of analog memory technologies for compute-in-memory applications, with deep dive into FeFET devices.

---

## Quick Navigation

| Section | Description |
|---------|-------------|
| [Comparison Table](#1-device-technologies-comparison) | Side-by-side comparison of all CIM technologies |
| [FeFET](#fefet-ferroelectric-field-effect-transistor) | Ferroelectric FET detailed analysis |
| [ReRAM](#reram-resistive-ram) | Resistive switching memory |
| [PCM](#pcm-phase-change-memory) | Phase-change memory |
| [MRAM](#mram-magnetoresistive-ram) | Magnetic tunnel junction memory |
| [FeFET Deep Dive](#3-fefet-deep-dive-hzo-material) | HZO material physics, Preisach model |

---

## 1. Device Technologies Comparison

| Technology | Mechanism | States | Speed | Endurance | Energy | CMOS Compatible | Maturity (reported) |
|------------|-----------|--------|-------|-----------|--------|-----------------|----------|
| **FeFET** | Ferroelectric polarization → Vth | multi-level (reported) [REPORTED] | 10-100 ns | 10⁹-10¹² [REPORTED] | Ultra-low | ✅ Yes | Reported TRL 4-6 |
| **ReRAM** | Conductive filament | 2-16 | 1-100 ns | 10⁶-10¹⁰ | Low | ✅ Yes | Reported TRL 6-8 |
| **PCM** | Crystalline ↔ Amorphous | 4-16 | 50-500 ns | 10⁸-10⁹ | Medium | ✅ Yes | Reported TRL 9 (Intel Optane) |
| **MRAM** | MTJ spin switching | 2-4 | 2-20 ns | >10¹⁵ | Low-Medium | ⚠️ Partial | Reported TRL 8-9 |

### Key Metrics Explained

- **States**: Number of stable analog levels per cell (higher = more bits/cell)
- **Speed**: Write/read latency (lower = faster inference)
- **Endurance**: Number of write cycles before failure (higher = longer lifetime)
- **Energy**: Energy per operation (lower = more efficient)
- **CMOS Compatible**: Can be integrated with standard CMOS fabrication

---

## 2. Device Technology Details

---

### FeFET (Ferroelectric Field-Effect Transistor)

#### Physical Mechanism
Voltage induces ferroelectric domain switching in HfO₂-based gate oxide, modulating channel threshold voltage (Vth).

```
ASCII Cross-Section:

     Gate (Control)
          |
    ┌─────┴─────┐
    │   HfZrO   │  ← Ferroelectric layer (10nm)
    ├───────────┤
    │    SiO₂   │  ← Interfacial layer (1-2nm)
    ├───────────┤
    │  Si (n+)  │  ← Channel
    └───────────┘
      ↑       ↑
    Source  Drain

Polarization States:
P→ (right): ⊕⊕⊕⊕ → Negative Vth shift → More current
P← (left):  ⊖⊖⊖⊖ → Positive Vth shift → Less current
```

#### Current-Voltage Characteristics
- **Non-linear**: Current depends on Vth shift from polarization
- **Transfer curve**: IDS vs VGS shows hysteresis
- **Multi-level**: Intermediate polarization states create multiple Vth levels

#### Multi-Level Storage
- **Research**: multi-level levels (reported in literature) demonstrated in lab conditions [REPORTED - Oh 2017, Song 2024]
- **Simulation baseline**: 30-level baseline (unverified) [PLAUSIBLE - within demonstrated range]
- **Mechanism**: Partial polarization via domain distribution

#### Endurance
- **Demonstrated**: 10⁹-10¹² cycles [REPORTED - IEEE IRPS 2022, Nano Letters 2024 (V:HfO₂), Science 2024]
- **V-doped HfO₂**: 10¹² cycles achieved with vanadium doping
- **Failure mechanism**: Wake-up, fatigue (domain pinning)

#### Retention
- **Target**: 10 years (extrapolated from accelerated aging)
- **Mechanism**: Depolarization field causes gradual retention loss
- **Mitigation**: Interfacial layer engineering

#### Variation Sources
- **Device-to-Device (D2D)**:
  - Ferroelectric domain distribution
  - Conventional MOSFET variation (RDF, WFV, LER)
- **Cycle-to-Cycle (C2C)**:
  - Stochastic domain switching (~1% after 10⁵ pulses)
  - Temperature-dependent variation

#### Advantages
✅ CMOS-compatible (same fab tools)
✅ High multi-level capability (multi-level states reported in literature; 30-level baseline is a demo baseline, configurable)
✅ Ultra-low energy (~10 fJ/bit)
✅ Non-volatile (retains data without power)
✅ High endurance (10⁹-10¹² cycles demonstrated) [REPORTED]

#### Challenges
❌ Limited commercial availability (reported TRL 4-6)
❌ Retention needs improvement (10 years target)
❌ Cycle-to-cycle variation (stochastic switching)

---

### ReRAM (Resistive RAM)

#### Physical Mechanism
Oxygen vacancy migration forms/breaks conductive filaments through a metal oxide.

```
ASCII Cross-Section:

    Top Electrode (Pt, TiN)
          |
    ┌─────┴─────┐
    │   HfO₂    │  ← Switching layer (5-20nm)
    │           │
    │    ●●●    │  ← Conductive filament (oxygen vacancies)
    │   ● ●     │
    │  ●   ●    │
    └───────────┘
          |
   Bottom Electrode (TiN, Pt)

States:
SET   (form filament):   LRS (Low Resistance State)
RESET (rupture filament): HRS (High Resistance State)
```

#### Operating Principle
- **SET**: Positive voltage drives oxygen ions away, forming vacancy filament (LRS)
- **RESET**: Negative voltage drives oxygen ions back, rupturing filament (HRS)
- **Analog states**: Partial filament formation creates intermediate resistances

#### Multi-Level Storage
- **Practical**: 4-16 levels achievable
- **Challenge**: Stochastic filament formation causes C2C variation
- **Mitigation**: Compliance current limiting, verify-after-write schemes

#### Endurance
- **Typical**: 10⁶-10⁸ cycles
- **High-endurance variants**: 10¹⁰ cycles (with advanced materials)
- **Failure**: Filament wear-out, stuck states

#### I-V Characteristics
- **Forming**: Initial high voltage to create first filament
- **SET**: ~1-2V, positive polarity
- **RESET**: ~1-2V, negative polarity
- **Read**: Low voltage (~0.1V) to avoid disturb

#### Advantages
✅ Simple structure (metal-insulator-metal)
✅ Fast switching (1-100 ns)
✅ CMOS back-end compatible
✅ Scalable to sub-10nm

#### Challenges
❌ Cycle-to-cycle variation (stochastic filament)
❌ Limited endurance (10⁶-10¹⁰ vs 10¹² for FeFET)
❌ Requires forming voltage
❌ Read disturb issues

---

### PCM (Phase-Change Memory)

#### Physical Mechanism
Joule heating changes Ge₂Sb₂Te₅ (GST) chalcogenide between crystalline and amorphous phases.

```
ASCII Cross-Section:

    Top Electrode
          |
    ┌─────┴─────┐
    │    GST    │  ← Phase-change layer (20-50nm)
    │  ░░▓▓░░   │  ░ = amorphous (high R)
    │  ░▓▓▓▓░   │  ▓ = crystalline (low R)
    └─────┬─────┘
          |
    Bottom Electrode (heater)

Phase Transition:
Amorphous → Crystalline: Slow heating (200-400°C), gradual cooling
Crystalline → Amorphous: Fast heating (600°C), rapid quench
```

#### Operating Principle
- **RESET (Amorphous)**: High current pulse → melt (>600°C) → rapid quench
- **SET (Crystalline)**: Medium current pulse → anneal (200-400°C) → slow cool
- **Analog levels**: Partial crystallization via pulse width/amplitude control

#### Multi-Level Storage
- **Practical**: 4-16 levels
- **Mechanism**: Mix of crystalline/amorphous regions
- **Challenge**: Resistance drift (crystalline grain growth over time)

#### Endurance
- **Typical**: 10⁸-10⁹ cycles
- **Failure**: Material degradation, void formation
- **Commercial**: Intel Optane (discontinued 2022)

#### I-V Characteristics
- **Threshold switching**: Material switches to low-R state above Vth
- **RESET**: Short, high-current pulse (melt-quench)
- **SET**: Long, medium-current pulse (anneal)
- **Read**: Low voltage below Vth

#### Resistance Drift
- **Mechanism**: Crystalline grain growth → resistance increases over time
- **Impact**: Analog states shift, reducing multi-level reliability
- **Mitigation**: Periodic refresh, drift-compensated encoding

#### Advantages
✅ Mature technology (Intel Optane, reported TRL 9)
✅ Fast read (20-50 ns)
✅ High endurance (10⁸-10⁹)
✅ CMOS-compatible

#### Challenges
❌ High RESET energy (~1-10 pJ/bit)
❌ Resistance drift (analog states shift)
❌ Limited scalability (needs heater current)
❌ Slow write (50-500 ns)

---

### MRAM (Magnetoresistive RAM)

#### Physical Mechanism
Spin-polarized current switches free layer magnetization in a magnetic tunnel junction (MTJ).

```
ASCII Cross-Section:

    Top Electrode
          |
    ┌─────┴─────┐
    │  CoFeB    │  ← Free layer (magnetization can flip)
    ├───────────┤     ↑↑↑ or ↓↓↓
    │   MgO     │  ← Tunnel barrier (1nm)
    ├───────────┤
    │  CoFeB    │  ← Pinned layer (fixed magnetization)
    └─────┬─────┘     ↑↑↑ (always up)
          |
   Bottom Electrode

States:
Parallel (P):     ↑↑↑ | ↑↑↑ → Low resistance (TMR)
Antiparallel (AP): ↓↓↓ | ↑↑↑ → High resistance
```

#### Operating Principle
- **Write**: Spin-polarized current (STT) switches free layer magnetization
  - P → AP: Electrons spin-polarized "down" push free layer down
  - AP → P: Electrons spin-polarized "up" pull free layer up
- **Read**: Measure resistance (TMR effect)

#### Multi-Level Storage
- **Current**: Primarily binary (2 states)
- **Research**: 4-level MRAM via multi-domain states
- **Challenge**: Difficult to create stable intermediate states

#### Tunnel Magnetoresistance (TMR)
- **Definition**: TMR = (R_AP - R_P) / R_P
- **Modern MTJ**: TMR > 200% (CoFeB/MgO)
- **Impact**: Higher TMR → better read margin

#### Endurance
- **Highest of all CIM technologies**: >10¹⁵ cycles
- **No wear-out**: Magnetic switching is reversible

#### I-V Characteristics
- **Write**: Current-driven (1-100 µA)
- **Read**: Voltage-driven (50-200 mV)
- **Switching time**: 2-20 ns (very fast)

#### Advantages
✅ Highest endurance (>10¹⁵ cycles)
✅ Fastest switching (2-20 ns)
✅ Non-volatile
✅ Radiation-hard (space applications)

#### Challenges
❌ Limited multi-level capability (2-4 states)
❌ Partial CMOS compatibility (requires magnetic materials)
❌ High write energy vs FeFET
❌ Stochastic switching at small dimensions

---

## 3. FeFET Deep Dive (HZO Material)

### 3.1 Material: Hf₀.₅Zr₀.₅O₂ (HZO)

#### Crystal Structure
- **Phase**: Orthorhombic (Pca2₁)
- **Why ferroelectric?**: Non-centrosymmetric structure → spontaneous polarization
- **Discovery**: Böscke et al., 2011 (unexpected ferroelectricity in HfO₂)

#### Stabilization Mechanisms
1. **Zr Doping**: Hf₀.₅Zr₀.₅O₂ ratio stabilizes orthorhombic phase
2. **Thin Film Confinement**: <10nm thickness favors orthorhombic over monoclinic
3. **Thermal Annealing**: 400-600°C post-deposition anneal induces phase transformation
4. **Capping Layer**: TiN or Pt electrodes provide interfacial stress

#### Key Parameters (from CLAUDE.md)

| Parameter | Value | Source |
|-----------|-------|--------|
| **Pr** (Remanent Polarization) | 15-34 µC/cm² (RT), 75 µC/cm² (4K) | Nature Commun. 2025, Adv. Elec. Mat. 2024 [REPORTED] |
| **Ec** (Coercive Field) | 0.6-1.5 MV/cm | Nature Commun. 2025, Nano Letters 2024 [REPORTED] |
| **Endurance** | 10⁹-10¹² demonstrated | IEEE IRPS 2022, Nano Letters 2024 (V:HfO₂) [REPORTED] |
| **Retention** | 10 years (extrapolated) | Literature |
| **Thickness** | 5-10 nm | Optimal for ferroelectricity |
| **States** | 30 discrete (demo baseline (configurable)), multi-level (reported) (others) | COSM 2025 [PLAUSIBLE], Oh 2017, Song 2024 [REPORTED] |

---

### 3.2 Polarization → Threshold Voltage Relationship

#### Physical Model
Polarization modulates the effective gate charge, shifting the threshold voltage.

**Equation:**
```
ΔVth = -(P × t_FE) / (C_ox × q)

Where:
  P      = Polarization (C/cm²)
  t_FE   = Ferroelectric layer thickness (cm)
  C_ox   = Gate oxide capacitance per unit area (F/cm²)
  q      = Elementary charge (1.6 × 10⁻¹⁹ C)
```

#### Sign Convention
- **P → channel** (positive polarization): Electrons induced at interface → Vth shifts **left** (lower Vth, more current)
- **P ← channel** (negative polarization): Electrons depleted at interface → Vth shifts **right** (higher Vth, less current)

#### Example Calculation
For Pr = 25 µC/cm², t_FE = 10 nm, C_ox = 10 µF/cm²:
```
ΔVth = -(25 × 10⁻⁶ × 10 × 10⁻⁷) / (10 × 10⁻⁶) ≈ -0.25 V
```
→ Threshold voltage shifts by ±0.25 V between full polarization states.

---

### 3.3 30-Level Mechanism

#### How 30 Discrete States Are Achieved

Traditional ferroelectrics: 2 states (binary)
**HZO FeFET**: 30+ states via partial polarization

**Mechanisms:**
1. **Domain Distribution**: HZO film contains ~10⁴-10⁶ ferroelectric domains per cell
2. **Independent Switching**: Each domain switches independently (stochastic)
3. **Partial Polarization**: Intermediate voltage → partial domain switching → intermediate Vth
4. **Superlattice Stabilization**: HfO₂-ZrO₂ superlattice creates energy barriers stabilizing intermediate states

#### Domain Switching Dynamics
- **Single domain**: Switches abruptly at Ec (coercive field)
- **Multi-domain ensemble**: Gradual switching over voltage range (distribution of Ec values)
- **Result**: Continuous analog tuning → discretized into 30 levels for digital compatibility (demo baseline; demo baseline (configurable))

#### 30 States = 4.9 bits/cell
```
log₂(30) ≈ 4.9 bits/cell

Compare to:
  NAND Flash: 3 bits/cell (TLC), 4 bits/cell (QLC)
  ReRAM: 2-4 bits/cell
  FeFET: 4.9 bits/cell (30-level demo baseline; demo baseline (configurable))
```

---

### 3.4 Preisach Model

The **Preisach model** is used to simulate hysteresis and multi-level behavior in FeFETs.

#### Mathematical Formulation

**Elementary Hysteron:**
- Each domain is modeled as a bistable hysteron with switching thresholds (α, β)
- α = up-switching threshold (positive field)
- β = down-switching threshold (negative field)

**Distribution Function:**
```
ρ(α, β) = probability density of hysterons with thresholds (α, β)

Total Polarization:
P(V) = ∬ ρ(α,β) × γ(α,β,V) dα dβ

Where:
  γ(α,β,V) = ±1 (hysteron state: +1 if up, -1 if down)
```

#### What the Preisach Model Captures

✅ **Hysteresis**: Memory of previous voltage history
✅ **Fatigue**: Endurance degradation (shift in ρ distribution)
✅ **Minor Loops**: Partial switching (voltage doesn't reach ±Ec)
✅ **Multi-level states**: Intermediate polarization from partial domain switching

#### Preisach Implementation (from module1-hysteresis)

```go
// module1-hysteresis/pkg/ferroelectric/preisach.go

type PreisachModel struct {
    Material    *HZOMaterial
    States      [][]int      // Hysteron memory array
    AlphaGrid   []float64    // Up-switching thresholds
    BetaGrid    []float64    // Down-switching thresholds
    Distribution [][]float64  // ρ(α,β) probability density
}

func (p *PreisachModel) Update(E float64) float64 {
    // Update hysteron states based on field E
    // Return total polarization P
}

func (p *PreisachModel) DiscreteStates(N int) []float64 {
    // Extract N discrete polarization levels
    // Used to get 30 discrete states (demo baseline; demo baseline (configurable))
}
```

---

### 3.5 Variation Sources

#### Device-to-Device (D2D) Variation
Affects array uniformity across different cells.

| Source | Mechanism | Impact |
|--------|-----------|--------|
| **Domain Distribution** | Non-uniform domain size/density | ±5-10% Pr variation |
| **RDF (Random Dopant Fluctuation)** | Random Hf/Zr distribution | ±2% Vth variation |
| **WFV (Work Function Variation)** | TiN grain boundary variation | ±3% Vth variation |
| **LER (Line Edge Roughness)** | Gate edge roughness | ±1% Vth variation |

**Total D2D Variation**: ~±10-15% (combines in quadrature)

#### Cycle-to-Cycle (C2C) Variation
Affects write repeatability for the same cell.

| Source | Mechanism | Impact |
|--------|-----------|--------|
| **Stochastic Switching** | Domain nucleation randomness | ±1% after 10⁵ pulses |
| **Temperature Fluctuation** | kT energy assists switching | ±0.5% at 85°C |

**Mitigation:**
- Verify-after-write schemes
- Error correction codes (ECC)
- Averaging over multiple cells

---

### 3.6 Drift Mechanisms

Unlike PCM (which has severe drift), FeFET drift is **milder** but still present.

#### Depolarization Field
- **Mechanism**: Internal field opposes polarization → gradual relaxation
- **Time scale**: Years (much slower than PCM)
- **Impact**: Slow retention loss

#### Charge Trapping
- **Mechanism**: Charges trap at HZO/SiO₂ interface → screen polarization
- **Time scale**: Hours to days
- **Mitigation**: Periodic refresh, interfacial layer optimization

#### Comparison to PCM Drift

| Technology | Drift Mechanism | Time Scale | Severity |
|------------|-----------------|------------|----------|
| **FeFET** | Depolarization, charge trapping | Days-years | Mild |
| **PCM** | Crystalline grain growth | Hours-months | Severe |

---

## 4. ASCII Device Cross-Sections

### FeFET Device Structure (Detailed)

```
┌───────────────────────────────────────────────────────────┐
│                   Gate Electrode (TiN)                     │
│                   (Control Voltage VG)                     │
└─────────────────────┬─────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │   Hf₀.₅Zr₀.₅O₂ (HZO)     │  ← Ferroelectric layer (10nm)
        │   ⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕⊕  │     Polarization domains
        ├───────────────────────────┤
        │        SiO₂              │  ← Interfacial oxide (1-2nm)
        ├───────────────────────────┤
        │    Silicon Channel       │  ← n+ or p+ doped
        └───────┬──────────┬────────┘
                │          │
         ┌──────┴──┐    ┌──┴──────┐
         │ Source  │    │  Drain  │  ← Metal contacts
         │  (S)    │    │   (D)   │
         └─────────┘    └─────────┘

Operation:
  VG > 0: P↑ (up) → Vth↓ → IDS↑ (high current state)
  VG < 0: P↓ (down) → Vth↑ → IDS↓ (low current state)
```

### ReRAM Device Structure

```
┌─────────────────────────────────────┐
│   Top Electrode (Pt, TiN)           │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┐
    │      HfO₂          │  ← Switching oxide (5-20nm)
    │                    │
    │       ●●●●●        │  ← Conductive filament
    │      ●     ●       │     (oxygen vacancies)
    │     ●       ●      │
    │      ●     ●       │
    │       ●●●●●        │
    └──────────┬──────────┘
               │
┌──────────────┴──────────────────────┐
│   Bottom Electrode (TiN, Pt)        │
└─────────────────────────────────────┘

SET:   V+ → O²⁻ migrates up → V••_O forms filament → LRS
RESET: V- → O²⁻ migrates down → filament ruptures → HRS
```

### PCM Device Structure

```
┌─────────────────────────────────────┐
│   Top Electrode (TiN)               │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┐
    │   Ge₂Sb₂Te₅ (GST)  │  ← Phase-change layer (20-50nm)
    │                    │
    │   ░░░▓▓▓▓▓░░░     │  ░ = Amorphous (high R)
    │   ░░▓▓▓▓▓▓░░      │  ▓ = Crystalline (low R)
    │   ░▓▓▓▓▓▓▓▓░      │
    └──────────┬──────────┘
               │
    ┌──────────┴──────────┐
    │   Heater Element    │  ← Resistive heater (W, TiN)
    └──────────┬──────────┘
               │
┌──────────────┴──────────────────────┐
│   Bottom Electrode (TiN)            │
└─────────────────────────────────────┘

RESET: High current → Joule heat → melt (600°C) → quench → Amorphous
SET:   Medium current → anneal (300°C) → slow cool → Crystalline
```

### MRAM (STT-MTJ) Device Structure

```
┌─────────────────────────────────────┐
│   Top Electrode (Ta)                │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┐
    │   CoFeB (Free)      │  ← Free layer (magnetization can flip)
    │   ↑↑↑↑↑↑↑↑↑↑↑↑↑    │     ↑↑↑ or ↓↓↓
    ├─────────────────────┤
    │   MgO Barrier       │  ← Tunnel barrier (0.8-1.2nm)
    ├─────────────────────┤
    │   CoFeB (Pinned)    │  ← Pinned layer (fixed magnetization)
    │   ↑↑↑↑↑↑↑↑↑↑↑↑↑    │     ↑↑↑ (always)
    ├─────────────────────┤
    │   Synthetic AF      │  ← Reference layer stabilization
    └──────────┬──────────┘
               │
┌──────────────┴──────────────────────┐
│   Bottom Electrode (Ta)             │
└─────────────────────────────────────┘

P (Parallel):     ↑↑↑ | ↑↑↑ → Low R
AP (Antiparallel): ↓↓↓ | ↑↑↑ → High R (TMR > 200%)
```

---

## 5. Technology Selection Guide

### When to Use Each Technology

| Application | Best Choice | Reason |
|-------------|-------------|--------|
| **Multi-bit CIM** | FeFET | 30+ levels, ultra-low energy |
| **Binary CIM** | MRAM | Highest endurance, fastest |
| **Embedded NVM** | ReRAM | CMOS back-end, mature |
| **Storage-class memory** | PCM | Mature (Intel Optane) |
| **Edge AI inference** | FeFET | Energy-efficient, high density |
| **Harsh environment** | MRAM | Radiation-hard, no wear-out |

### Workload Considerations

| Workload | Write Frequency | Best Technology |
|----------|-----------------|-----------------|
| **Neural Network Inference** | Low (weights fixed) | FeFET (multi-level) |
| **In-Memory Training** | High (weight updates) | MRAM (high endurance) |
| **Reconfigurable Logic** | Medium | ReRAM (fast, CMOS-compatible) |

---

## 6. References

### FeFET (HZO)
1. Böscke et al., "Ferroelectricity in hafnium oxide thin films", *Appl. Phys. Lett.* 2011
2. Dr. external research group, COSM 2025 (30 discrete states demo baseline (configurable); unverified)
3. [Nature Commun. 2025](https://doi.org/10.1038/s41467-025-61758-2) - HfO₂/ZrO₂ superlattice (Pr = 15-34 µC/cm², Ec = 1.0-1.5 MV/cm)
4. [IEEE IRPS 2022](https://doi.org/10.1109/IRPS48227.2022.9764533) - High endurance HZO (>10¹¹ cycles)
5. [PMC11197553](https://pmc.ncbi.nlm.nih.gov/articles/PMC11197553/) - HfO₂ ferroelectric endurance review (>5×10¹² cycles)
6. [PMC9740545](https://pmc.ncbi.nlm.nih.gov/articles/PMC9740545/) - Polarization switching kinetics in HZO
7. [Nature Reviews Materials](https://doi.org/10.1038/s41578-022-00431-2) - HfO₂ fundamentals review

### ReRAM
5. Wong et al., "Metal-Oxide RRAM", *Proc. IEEE* 2012
6. Ielmini, "Resistive switching memories", *Nature Electron.* 2018

### PCM
7. Raoux et al., "Phase-change random access memory", *IBM JRD* 2008
8. Intel Optane (discontinued 2022)

### MRAM
9. Ikeda et al., "A perpendicular-anisotropy CoFeB–MgO MTJ", *Nature Mater.* 2010
10. Worledge et al., "Spin torque switching of perp. Ta/CoFeB/MgO", *Appl. Phys. Lett.* 2011

### Preisach Model
11. Mayergoyz, "Mathematical models of hysteresis", *IEEE Trans. Mag.* 1986
12. Ni et al., "Preisach model for HfO₂ FeFETs", *IEEE EDL* 2019

---

## 7. Appendix: Parameter Definitions

| Parameter | Definition | Units |
|-----------|------------|-------|
| **Pr** | Remanent polarization (at E=0) | µC/cm² |
| **Ec** | Coercive field (field to switch) | MV/cm |
| **TMR** | Tunnel magnetoresistance ratio | % |
| **HRS/LRS** | High/Low resistance state | Ω |
| **R_AP/R_P** | Antiparallel/Parallel resistance | Ω |
| **ENOB** | Effective number of bits (ADC/DAC) | bits |
| **Vth** | Threshold voltage | V |
| **IDS** | Drain-source current | A |

---

**Document Status**: ✅ Complete
**Last Updated**: 2026-01-28
**Maintainer**: FeCIM Lattice Tools Project
**Verification**: See `HONESTY_AUDIT.md` for full claim verification details
**Related Docs**:
- `CLAUDE.md` (physics constants)
- `docs/development/scriptReference.md` (implementation)
- `module1-hysteresis/` (Preisach model code)
