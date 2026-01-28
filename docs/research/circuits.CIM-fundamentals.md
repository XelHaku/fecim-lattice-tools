# Compute-in-Memory (CIM) Fundamentals for FeFET Arrays

> **Module 4 Reference Document** - Physics basis for peripheral circuit simulation

## Overview

Ferroelectric FET (FeFET) based Compute-in-Memory performs matrix-vector multiplication (MVM) in O(1) time using analog physics:
- **Multiplication**: Ohm's Law (I = G × V)
- **Accumulation**: Kirchhoff's Current Law (currents sum at column nodes)

This document covers the physics of READ, WRITE, and COMPUTE operations.

---

## 1. Read Operation Physics

### 1.1 Sensing Mechanism

FeFET read is **non-destructive** - unlike FeRAM which requires rewrite after read. The polarization state modulates the channel threshold voltage (Vth):

| Polarization State | Vth Shift | Channel State | Conductance |
|-------------------|-----------|---------------|-------------|
| Positive (+Pr) | Low Vth | Enhanced | High G (Gmax) |
| Negative (-Pr) | High Vth | Depleted | Low G (Gmin) |

**Key Equation:**
```
ΔVth = -2 × Pr × t_fe / (ε₀ × ε_fe)
```
Where:
- Pr = remnant polarization (15-34 µC/cm² for HZO)
- t_fe = ferroelectric thickness (10nm typical)
- ε_fe = ferroelectric permittivity (~25 for HZO)

### 1.2 Safe Read Voltage

Read voltage must stay **below coercive voltage** to avoid disturbing the polarization state:

| Parameter | Value | Derivation |
|-----------|-------|------------|
| Coercive Field (Ec) | 0.6-1.5 MV/cm | Material property |
| Film Thickness | 10 nm | Fabrication choice |
| Coercive Voltage (Vc) | 0.6-1.5 V | Vc = Ec × thickness |
| Safe Read Max | 0.3-0.75 V | 50% of Vc (field_min_ratio) |

**Read Disturb Concern**: Repetitive reads at high voltage can cause gradual depolarization. Module 4 uses `field_min_ratio = 0.5` from physics.yaml.

### 1.3 TIA-Based Current Sensing

The Transimpedance Amplifier (TIA) converts cell current to voltage:

```
V_out = I_cell × R_tia + V_offset
```

Module 4 defaults:
- R_tia = 10 kΩ (gain)
- I_max = 100 µA (saturation)
- V_out_max = 1.0 V

---

## 2. Write Operation Physics

### 2.1 Polarization Switching

FeFET write uses electric field to switch ferroelectric domains:

**Switching Mechanism**: 3D anisotropic nucleation and growth along orthorhombic phase axes. The process is **nucleation-limited** with strong temperature dependence.

**Key Parameters:**

| Parameter | HZO (10nm) | FTJ | Source |
|-----------|------------|-----|--------|
| Ec | 1.0 MV/cm | 0.8 MV/cm | Nature Commun. 2025 |
| Vc (= Ec × t) | 1.0 V | 0.8 V | Calculated |
| Write Min | 1.0 V | 0.8 V | At Vc |
| Write Max | 2.5 V | 2.0 V | 2.5× Vc |
| Switching Time | <10 ns | <1 ns | IEEE 2024 |

### 2.2 Multi-Level Programming

Analog states are achieved through **partial polarization switching**:

1. **Voltage Amplitude Method**: Lower V → fewer domains switch → intermediate state
2. **Pulse Width Method**: Shorter pulse → incomplete switching
3. **ISPP (Incremental Step Pulse Program)**: Best linearity, 90 states demonstrated

**Module 4 Implementation:**
- 30 discrete levels (FeCIM standard)
- Linear voltage mapping: Level 0 → Vc, Level 29 → 2.5×Vc
- Each level corresponds to a conductance state

### 2.3 Write Voltage Derivation

From physics.yaml calibration:
```go
Vc = material.CoerciveField * material.Thickness  // Ec × t
WriteMin = Vc                                      // Just above switching
WriteMax = Vc * calibration.FieldMaxRatio         // 2.5 × Vc default
```

---

## 3. Compute (MVM) Operation

### 3.1 Matrix-Vector Multiplication Principle

```
         V₀   V₁   V₂   V₃  (DAC voltages = input vector)
          ↓    ↓    ↓    ↓
    WL₀ →[G₀₀][G₀₁][G₀₂][G₀₃]→ I₀ = Σ(Gᵢⱼ × Vⱼ)
    WL₁ →[G₁₀][G₁₁][G₁₂][G₁₃]→ I₁ = Σ(Gᵢⱼ × Vⱼ)
    WL₂ →[G₂₀][G₂₁][G₂₂][G₂₃]→ I₂ = Σ(Gᵢⱼ × Vⱼ)
          ↓    ↓    ↓    ↓
                          TIA → ADC → Output vector
```

**Physics Basis:**
1. **Ohm's Law**: I = G × V (multiplication at each crosspoint)
2. **KCL**: Currents sum along bitlines (accumulation)
3. **Result**: Output current = dot product of conductance row × voltage column

### 3.2 Current Calculation

For row r with all word lines active:
```go
I_r = Σ(G[r][c] × V[c]) for c = 0 to cols-1
```

Where:
- G[r][c] = conductance of cell at (r,c) in Siemens
- V[c] = DAC voltage on column c in Volts
- I_r = total current into row r in Amperes

### 3.3 Conductance-to-Level Mapping

Current implementation (linear):
```go
G = Gmin + (Gmax - Gmin) × (level / (numLevels - 1))
```

**Gap Identified**: Should use nonlinear FeFET model:
```go
G = Gmin + (Gmax - Gmin) × sigmoid(k × (P - P_threshold))
```

---

## 4. Architecture Differences

### 4.1 Passive Crossbar (0T1R)

**Structure**: No transistor, direct crosspoint connection

**Advantages:**
- Highest density (no transistor area)
- Simplest fabrication
- Capacitive ferroelectric: No DC sneak paths

**Disadvantages:**
- **Sneak Path Currents**: Unselected cells create parallel current paths
- Sneak error: 5-20% of signal current in worst case
- All word lines always active

**Module 4 Behavior**: When passive mode selected:
- All WL checkboxes locked ON
- UI shows "Passive: All rows always on"
- Sneak path warning displayed

### 4.2 1T1R Architecture

**Structure**: One transistor per cell for row selection

**Advantages:**
- Eliminates sneak paths
- True single-cell selection
- Better read accuracy

**Disadvantages:**
- Lower density (transistor area)
- Series resistance limits current
- More complex fabrication

### 4.3 2T1R Architecture

**Structure**: Additional transistor for column selection

**Advantages:**
- Best isolation
- Reduced wire IR drop
- Precise cell control

**Disadvantages:**
- Lowest density
- Highest complexity

---

## 5. Key Performance Parameters

### 5.1 Conductance Range

| Parameter | Value | Source |
|-----------|-------|--------|
| Gmin | 1 µS | physics.yaml default |
| Gmax | 100 µS | physics.yaml default |
| On/Off Ratio | 100:1 | Gmax/Gmin |
| Advanced Devices | up to 10⁵:1 | Grain engineering 2025 |

### 5.2 Current and Voltage Ranges

| Operation | Voltage Range | Current Range |
|-----------|---------------|---------------|
| Read | 0 - 0.5V | 0.5 - 50 µA |
| Write | 1.0 - 2.5V | N/A (switching) |
| Compute | 0 - 1.0V | 0 - 100 µA (saturates) |

### 5.3 ADC Requirements

For 30-level FeCIM:
- **Minimum bits**: 5-bit (32 levels)
- **Effective bits**: 4.5-bit with INL/DNL
- **Power trade-off**: Higher resolution = higher power

---

## 6. Non-Idealities

### 6.1 IR Drop

Wire resistance causes voltage drop from source to far cells:
```
V_effective = V_applied - I × R_wire
```

Impact: 2-5% voltage reduction at far corner of 64×64 array

### 6.2 Sneak Paths (Passive Only)

Parallel current paths through unselected cells:
```
I_sneak = V_read × Σ(G_unselected_parallel_paths)
```

Impact: 5-20% error in worst case

### 6.3 Device Variation

- **Cycle-to-cycle**: Gaussian σ/μ ≈ 5%
- **Device-to-device**: σ/μ ≈ 10%
- **Temperature drift**: ΔG/G ≈ 0.1%/°C

### 6.4 Endurance

| Material | Endurance | Source |
|----------|-----------|--------|
| HZO | 10⁹ cycles | IEEE IRPS 2022 |
| V:HfO₂ | 10¹² cycles | Nano Letters 2024 |
| Superlattice | 10¹⁰ + recovery | Nature Commun. 2025 |

---

## 7. Module 4 Implementation Status

### Currently Implemented
- [x] DAC voltage generation (read/write presets)
- [x] TIA current-to-voltage conversion
- [x] ADC voltage-to-level quantization
- [x] Basic MVM (I = G × V summation)
- [x] Material-derived voltage ranges
- [x] Architecture selection (0T1R, 1T1R, 2T1R)
- [x] Passive mode WL enforcement

### Gaps to Address (see MODULE4-PHYSICS-IMPROVEMENTS.md)
- [ ] Nonlinear conductance model
- [ ] IR drop integration from Module 2
- [ ] Sneak path calculation for passive mode
- [ ] TIA noise in simulation path
- [ ] ADC INL/DNL in simulation path

---

## References

1. Nature Communications 2025 - HfO₂/ZrO₂ superlattice ferroelectrics
2. IEEE IRPS 2022 - FeFET endurance characterization
3. Science 2025 - Ultralow coercive field rhombohedral phase
4. Advanced Functional Materials 2025 - 3D anisotropic switching
5. Dr. external research group, COSM 2025 - FeCIM demonstration
6. Module 2 crossbar/nonidealities.go - IR drop and sneak path models
