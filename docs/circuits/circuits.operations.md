# Crossbar Array Operations: Passive (0T1R) vs 1T1R Architectures

**Read, Write, and Compute Operations for Ferroelectric Compute-in-Memory Arrays**

*Last Updated: January 2026*

---

## Executive Summary

This document details the operational rules for FeCIM crossbar arrays in two architectures:

| Architecture | Cell Structure | Density | Sneak Path Error | Best For |
|--------------|----------------|---------|------------------|----------|
| **Passive (0T1R)** | FeFET only | 4F² | 5-20% | High density, edge inference |
| **1T1R** | Transistor + FeFET | 8-12F² | ~0% | High accuracy, training |

**Key Papers Referenced:**
- *Crossbar_Sneak_Path_Analysis_arXiv* — Sneak path modeling
- *sneak_path_self_rectifying_arrays_2022* — Self-rectifying mitigation
- *multilevel_fefet_crossbar_2023* — Multi-level FeFET programming
- *FeFET_Crossbar_MNIST_Hardware_arXiv* — Hardware demonstration
- *ferroelectric_CIM_review_2023* — Comprehensive FeCIM review
- *Multi_Level_FeFET_Programming_arXiv* — Programming schemes
- *Temperature_Resilient_FeFET_CIM_2024* — Thermal robustness
- *Analog_CIM_Energy_Efficiency_arXiv* — Energy analysis

---

## 1. Architecture Overview

### 1.1 Passive Array (0T1R)

```
           BL₀    BL₁    BL₂    BL₃     ← Bit Lines (Columns)
            │      │      │      │
WL₀ ────────●──────●──────●──────●────→ SL₀
            │      │      │      │
WL₁ ────────●──────●──────●──────●────→ SL₁
            │      │      │      │
WL₂ ────────●──────●──────●──────●────→ SL₂
            │      │      │      │
WL₃ ────────●──────●──────●──────●────→ SL₃
            ↑                     ↑
        Word Lines           Sense Lines
         (Rows)               (Outputs)

● = FeFET cell (direct WL-to-BL connection)
```

**Characteristics:**
- **Density:** 4F² per cell (F = minimum feature size) — highest possible
- **Fabrication:** Simple, fewer masks, lower cost
- **Sneak Paths:** Current flows through unintended paths (5-20% error)
- **Read Disturb:** Unselected cells see partial voltage, risk of state change
- **Write Disturb:** Half-select problem affects adjacent cells

**Paper Reference:** *sneak_path_self_rectifying_arrays_2022* demonstrates 10⁴:1 rectification ratio in self-rectifying FeFET, enabling viable passive arrays.

### 1.2 1T1R Array (Transistor-Gated)

```
           BL₀    BL₁    BL₂    BL₃     ← Bit Lines (Columns)
            │      │      │      │
       ┌─T─┤      │      │      │
WL₀ ───○───●──────●──────●──────●────→ SL₀
       │   │      │      │      │
       ├─T─┤      │      │      │
WL₁ ───○───●──────●──────●──────●────→ SL₁
       │   │      │      │      │
       ├─T─┤      │      │      │
WL₂ ───○───●──────●──────●──────●────→ SL₂
       │   │      │      │      │
       └─T─┤      │      │      │
WL₃ ───○───●──────●──────●──────●────→ SL₃
       ↑
   Transistor Gate
   (Row Selector)

○ = Access transistor (controlled by WL)
● = FeFET storage element
```

**Characteristics:**
- **Density:** 8-12F² per cell (transistor adds area)
- **Fabrication:** Standard CMOS + FeFET integration
- **Sneak Paths:** Eliminated (~1000:1 on/off ratio isolation)
- **Read Disturb:** None (unselected rows fully isolated)
- **Write Disturb:** None (transistor gates write current)

**Paper Reference:** *FeFET_Crossbar_MNIST_Hardware_arXiv* demonstrates 87% MNIST accuracy with 1T1R FeFET array (128×64), showing practical elimination of sneak-path errors.

---

## 2. WRITE Operation

### 2.1 Write Physics

FeFET write relies on ferroelectric polarization switching:

```
Write Voltage Requirements (HZO-based FeFET, 10nm film):
- Coercive field Ec: 0.6-1.5 MV/cm
- Coercive voltage Vc = Ec × thickness ≈ 0.6-1.5V
- Programming voltage Vprog: 1.2-1.5V (safe margin above Vc)
- Programming pulse width: 50-100ns (typical)

Source: Nature Communications 2025, Nano Letters 2024
```

**30-Level Programming:**

Per *multilevel_fefet_crossbar_2023*, multi-level states are achieved via:
1. **Incremental Step Pulse Programming (ISPP):** Series of increasing voltage pulses
2. **Write-Verify:** Program, read, adjust until target level reached
3. **Partial Polarization:** Intermediate voltages create partial domain switching

```
Programming Pulse Sequence (Write-Verify):

Level Target: 15 (of 0-29)

Step 1: Apply V_start = 1.0V × (15/29) = 0.52V
Step 2: Read back → measured level = 12 (too low)
Step 3: Apply V_adjust = 0.52V + 0.05V = 0.57V
Step 4: Read back → measured level = 15 ✓ (target reached)

Typical convergence: 3-5 iterations
```

### 2.2 Write in Passive (0T1R) Array

**Problem: Half-Select Disturb**

When writing to cell (1,1), adjacent cells experience partial voltage:

```
Write to cell (1,1): WL₁=+1.5V, BL₁=0V

         BL₀=0.75V  BL₁=0V   BL₂=0.75V
              │        │         │
WL₀=0V  ──────●────────●─────────●────
              │   +0.75V │  +0.75V│
              │        │         │
WL₁=+1.5V ────●────────●─────────●────
              │ +0.75V │ +1.5V   │ +0.75V
              │    ✗   │   ✓     │   ✗
              │        │ TARGET  │
WL₂=0V  ──────●────────●─────────●────
              │   +0.75V │  +0.75V│
                   ✗          ✗

✓ = Full write voltage (correct)
✗ = Half-select voltage (potential disturb)
```

**Mitigation Strategy (V/2 Scheme):**

Per *Crossbar_Sneak_Path_Analysis_arXiv*:
- Unselected WLs biased to V/2 (0.75V)
- Unselected BLs biased to V/2 (0.75V)
- Target cell sees full ±V (1.5V - 0V = 1.5V)
- Half-selected cells see ±V/2 (below Vc threshold)

```
Voltage Distribution (V/2 Scheme):

Cell at (1,1): 1.5V - 0V = +1.5V     → WRITES
Cell at (0,1): 0.75V - 0V = +0.75V   → No disturb (below Vc)
Cell at (1,0): 1.5V - 0.75V = +0.75V → No disturb (below Vc)
Cell at (0,0): 0.75V - 0.75V = 0V    → No disturb
```

**Write Accuracy (Passive):**
- Without V/2: 10-30% of adjacent cells may disturb
- With V/2: <1% disturb rate (depends on Vc margin)
- Self-rectifying FeFET: Additional 10-100× margin

### 2.3 Write in 1T1R Array

**Transistor Gating:**

```
Write to cell (1,1):

WL Control:
- WL₁ (selected row): HIGH (transistor ON)
- WL₀, WL₂, WL₃: LOW (transistors OFF)

Voltage Path:
BL₁ → Transistor(ON) → FeFET(1,1) → SL₁(grounded)

         BL₀    BL₁=+1.5V  BL₂
          │        │        │
    [OFF] │   [ON] │  [OFF] │
WL₀ ───○──●────○───●────○───●────
       ↓       ↓       ↓
WL₁ ───○──●────○───●────○───●────
      OFF     ON ←  OFF
              │
              ↓
         FeFET gets
         full 1.5V
              │
WL₂ ───○──●────○───●────○───●────
      OFF      OFF     OFF
```

**Advantages:**
1. **No half-select:** Unselected rows completely isolated
2. **Full voltage:** Target cell receives exact Vprog
3. **No adjacent disturb:** Current cannot flow through OFF transistors
4. **Deterministic:** Write succeeds in single pulse (no retry needed)

**Write Accuracy (1T1R):**
- Disturb rate: 0% (transistor provides ~1000:1 isolation)
- Write variation: 3-5% (device-to-device only)

**Paper Reference:** *Multi_Level_FeFET_Programming_arXiv* shows 1T1R enables 32-140 discrete levels with <5% variation.

### 2.4 Write Timing Comparison

| Parameter | Passive (0T1R) | 1T1R |
|-----------|----------------|------|
| Single pulse write | 50-100ns | 50-100ns |
| Write-verify iterations | 3-5 | 1-2 |
| Total write time | 200-500ns | 50-150ns |
| Energy per write | 10-50 fJ | 10-30 fJ |
| Endurance | 10¹⁰ cycles | 10¹⁰ cycles |

**Source:** *Analog_CIM_Energy_Efficiency_arXiv*, IEEE IRPS 2022

---

## 3. READ Operation

### 3.1 Read Physics

FeFET read measures drain current modulated by polarization state:

```
Read Operation:
1. Apply small read voltage Vread (must be << Vc to avoid disturb)
2. Measure drain current I_d = G × V_read
3. Convert current to level via TIA + ADC

Safe Read Voltage:
- Vread ≤ 0.5V (well below Vc of 0.6-1.5V)
- Typical: Vread = 0.1-0.3V for non-destructive read

Current Range (30 levels):
- Level 0 (erased): I_min ≈ 1 µA
- Level 29 (programmed): I_max ≈ 100 µA
- Current ratio: 100:1 (excellent sensing margin)
```

### 3.2 Read in Passive (0T1R) Array

**Problem: Sneak Path Currents**

When reading cell (1,1), parallel paths add parasitic current:

```
Read target: cell (1,1)
Apply: Vread = 0.3V on BL₁, measure current on SL₁

DESIRED PATH:
BL₁ → FeFET(1,1) → SL₁
Current = G(1,1) × 0.3V = 30µA (target level 15)

SNEAK PATH EXAMPLE:
BL₁ → FeFET(0,1) → WL₀ → FeFET(0,0) → BL₀ → ... → SL₁
Parallel current through 3+ cells

         BL₀        BL₁=+0.3V    BL₂
          │            │          │
WL₀ ──────●────────────●──────────●────
          │     ↗      │          │
          │   Sneak    │          │
          │   Path     │          │
WL₁ ──────●────────────●──────────●────→ SL₁ (measure here)
          │            │ ↓        │
          │         Target        │
          │         Current       │
WL₂ ──────●────────────●──────────●────
```

**Three-Cell Sneak Path Model:**

Per *Crossbar_Sneak_Path_Analysis_arXiv*:

```
Sneak conductance (worst case):
G_sneak = Σ[ 1 / (1/G_a + 1/G_b + 1/G_c) ]

For each parallel path through cells a, b, c

Example calculation:
- G_target = 50 µS (level 15)
- G_sneak ≈ 5 µS (from ~100 parallel paths in 128×128 array)

Sneak error = G_sneak / G_target = 5/50 = 10%
```

**Measured vs. Actual:**
- Measured current: I_meas = I_target + I_sneak
- Typical error: 5-20% depending on array size and weight distribution

**Mitigation Strategies:**

1. **Floating rows/columns:** Bias unselected lines to eliminate some paths
2. **Self-rectifying FeFET:** Built-in diode reduces reverse current 10-1000×
3. **V/2 biasing:** Similar to write, reduces sneak magnitude
4. **Calibration:** Measure and subtract estimated sneak current

**Read Accuracy (Passive):**

| Array Size | Sneak Error | Effective Bits |
|------------|-------------|----------------|
| 16×16 | 2-5% | 4.5 bits |
| 64×64 | 5-10% | 4.0 bits |
| 128×128 | 10-15% | 3.5 bits |
| 256×256 | 15-25% | 3.0 bits |

**Source:** *memory_tech_crossbar_dnn_accuracy_2024*

### 3.3 Read in 1T1R Array

**Transistor Isolation:**

```
Read target: cell (1,1)

WL Control:
- WL₁ = HIGH → Transistor ON → Row 1 connected
- WL₀, WL₂, WL₃ = LOW → Transistors OFF → Rows isolated

         BL₀    BL₁=+0.3V  BL₂
          │        │        │
    [OFF] │  [OFF] │  [OFF] │
WL₀ ───○──●────○───●────○───●────  (isolated)
              │
    [OFF]     │ON  │  [OFF]
WL₁ ───○──●────○───●────○───●────→ SL₁
              │    ↓
              │  Only this path
              │  conducts
              │
    [OFF] │  [OFF] │  [OFF] │
WL₂ ───○──●────○───●────○───●────  (isolated)
```

**Clean Signal Path:**
- Current path: BL₁ → T₁₁(ON) → FeFET(1,1) → SL₁
- No parallel paths possible
- Measured = Actual (within device noise)

**Read Accuracy (1T1R):**

| Array Size | Sneak Error | Effective Bits |
|------------|-------------|----------------|
| 16×16 | ~0% | 4.9 bits |
| 64×64 | ~0% | 4.9 bits |
| 128×128 | ~0% | 4.9 bits |
| 256×256 | ~0% | 4.9 bits |

**Paper Reference:** *FeFET_Crossbar_MNIST_Hardware_arXiv* achieves full 30-level resolution (4.9 bits) with 1T1R architecture.

### 3.4 Read Timing Comparison

| Parameter | Passive (0T1R) | 1T1R |
|-----------|----------------|------|
| TIA settling | 10ns | 10ns |
| ADC conversion | 50ns | 50ns |
| Total read time | ~60ns | ~60ns |
| Read accuracy | ±2-3 levels | ±0.5 levels |
| Read energy | 1-5 fJ | 2-8 fJ |

---

## 4. COMPUTE Operation (Matrix-Vector Multiplication)

### 4.1 Compute Physics

MVM leverages Ohm's Law and Kirchhoff's Current Law:

```
Input: Voltage vector V = [V₀, V₁, V₂, ..., Vₙ] on columns
Weights: Conductance matrix G stored in cells
Output: Current vector I = [I₀, I₁, I₂, ..., Iₘ] on rows

Physics:
I_row_i = Σⱼ (G_ij × V_j)    [Ohm's Law per cell]
                             [KCL sums currents on row]

This IS matrix multiplication: I = G × V
Computed in O(1) time (single analog cycle)!
```

**Timing:**
- DAC settling: 5-10ns
- Array propagation: 5ns
- TIA/ADC: 50-60ns
- **Total MVM: ~60-70ns** (regardless of array size!)

**Paper Reference:** *Analog_CIM_Energy_Efficiency_arXiv* reports 10-1000× energy efficiency over digital MACs.

### 4.2 Compute in Passive (0T1R) Array

**All Rows Active (Intentional):**

Unlike read/write, MVM requires ALL cells to participate:

```
Compute: y = W × x

Input voltages x on ALL columns:
BL₀ = V₀, BL₁ = V₁, BL₂ = V₂, BL₃ = V₃

All rows collect summed current:
SL₀ collects: I₀ = G₀₀×V₀ + G₀₁×V₁ + G₀₂×V₂ + G₀₃×V₃
SL₁ collects: I₁ = G₁₀×V₀ + G₁₁×V₁ + G₁₂×V₂ + G₁₃×V₃
...

         V₀=0.5V  V₁=0.8V  V₂=0.2V  V₃=0.9V
            │        │        │        │
WL₀ ────────●────────●────────●────────●────→ I₀
            │        │        │        │
WL₁ ────────●────────●────────●────────●────→ I₁
            │        │        │        │
WL₂ ────────●────────●────────●────────●────→ I₂
            │        │        │        │
WL₃ ────────●────────●────────●────────●────→ I₃
```

**Sneak Path Error in MVM:**

In compute mode, sneak paths STILL affect output:
- Current intended for SL₀ leaks to SL₁, etc.
- Cross-row interference adds 5-20% error to each output
- Error compounds through neural network layers

**MVM Accuracy (Passive):**

Per *memory_tech_crossbar_dnn_accuracy_2024*:

| Metric | Ideal | Passive (0T1R) |
|--------|-------|----------------|
| Single layer error | 0% | 5-15% |
| 2-layer network | 0% | 8-25% |
| 4-layer network | 0% | 15-40% |
| MNIST accuracy | 98.5% | 85-92% |
| CIFAR-10 accuracy | 92% | 75-85% |

**Mitigation for Passive MVM:**
1. **Error-aware training:** Train neural network expecting ~10% noise
2. **Smaller arrays:** Tile into 64×64 sub-arrays
3. **Calibration:** Measure error matrix, subtract in software
4. **Self-rectifying:** Reduces cross-talk significantly

### 4.3 Compute in 1T1R Array

**Critical Difference: ALL Transistors ON**

For MVM, 1T1R turns ON all access transistors:

```
Compute mode: All WL = HIGH

         V₀       V₁       V₂       V₃
          │        │        │        │
    [ON]  │  [ON]  │  [ON]  │  [ON]  │
WL₀ ───○──●────○───●────○───●────○───●────→ I₀
       ↓       ↓       ↓       ↓
ALL TRANSISTORS CONDUCTING
       ↓       ↓       ↓       ↓
WL₁ ───○──●────○───●────○───●────○───●────→ I₁
       ON      ON      ON      ON
```

**Why Enable All Transistors?**

MVM requires full matrix participation:
- Each output I_i needs contributions from ALL columns
- Disabling any transistor would skip that weight
- COMPUTE = Full matrix active (unlike READ = single cell)

**But Still Better Than Passive:**

Even with all transistors ON, 1T1R provides benefits:
1. **Uniform current distribution:** No resistance mismatch
2. **Lower IR drop:** Transistor ON-resistance helps equalize paths
3. **No sneak path multiplication:** Clean Ohm's Law operation

**MVM Accuracy (1T1R):**

| Metric | Ideal | 1T1R |
|--------|-------|------|
| Single layer error | 0% | 1-3% |
| 2-layer network | 0% | 2-5% |
| 4-layer network | 0% | 4-8% |
| MNIST accuracy | 98.5% | 96-98% |
| CIFAR-10 accuracy | 92% | 88-91% |

**Paper Reference:** *FeFET_Crossbar_MNIST_Hardware_arXiv* achieves 87% MNIST in first-generation hardware; *multilevel_fefet_crossbar_2023* projects 96%+ with optimized 1T1R.

### 4.4 Compute Timing Comparison

| Parameter | Passive (0T1R) | 1T1R |
|-----------|----------------|------|
| DAC settling | 10ns | 10ns |
| Array compute | 5ns | 5ns |
| TIA settling | 10ns | 10ns |
| ADC conversion | 50ns | 50ns |
| **Total MVM** | **~75ns** | **~75ns** |
| Throughput (64×64) | 54 GOPS | 54 GOPS |
| Throughput (256×256) | 870 GOPS | 870 GOPS |

**Note:** Timing is identical; the difference is accuracy.

---

## 5. Operation Mode Summary

### 5.1 Quick Reference Table

| Operation | Passive (0T1R) | 1T1R |
|-----------|----------------|------|
| **WRITE** | V/2 scheme required, 10-30% adjacent disturb risk | Transistor gates single row, 0% disturb |
| **READ** | Sneak paths add 5-20% error | Transistor isolates, ~0% error |
| **COMPUTE** | All rows active, 5-15% sneak error | All transistors ON, 1-3% error |

### 5.2 Transistor State by Operation

| Operation | Selected Row | Unselected Rows |
|-----------|--------------|-----------------|
| WRITE (1T1R) | ON | OFF |
| READ (1T1R) | ON | OFF |
| COMPUTE (1T1R) | ON | ON (all rows!) |

### 5.3 When to Use Each Architecture

**Choose Passive (0T1R) when:**
- Density is critical (edge AI, mobile)
- Inference only (no training)
- Error tolerance exists (≥10% margin in application)
- Self-rectifying FeFET available

**Choose 1T1R when:**
- Accuracy is critical (medical, automotive)
- Training in memory required
- Multi-layer networks (error compounds)
- High bit-precision needed (≥4 bits)

---

## 6. Voltage Specifications

### 6.1 Operating Voltages

| Operation | Target Voltage | Unselected Voltage | Duration |
|-----------|---------------|--------------------| ---------|
| WRITE (positive) | +1.2 to +1.5V | V/2 or floating | 50-100ns |
| WRITE (negative) | -1.2 to -1.5V | V/2 or floating | 50-100ns |
| READ | +0.1 to +0.5V | 0V or floating | 10-20ns |
| COMPUTE (input) | 0 to +1.0V | N/A (all active) | 5-10ns |

### 6.2 Voltage Zones

```
Voltage Safety Zones for HZO FeFET (10nm):

         -1.5V              0V              +1.5V
           │                 │                 │
  ─────────┼─────────────────┼─────────────────┼─────────
           │    SAFE READ    │    SAFE READ    │
           │   (no disturb)  │   (no disturb)  │
     ◄─────┼────────────────►│◄────────────────┼─────►
    ERASE  │   -0.5V to +0.5V │              PROGRAM
   WRITE   │                 │               WRITE
           │                 │                 │
  ─────────┴─────────────────┴─────────────────┴─────────
            ← Vc ≈ 0.6-1.5V →

Safe read: |V| < Vc (typically < 0.5V)
Program: V > Vc + margin (typically 1.2-1.5V)
Erase: V < -Vc - margin (typically -1.2 to -1.5V)
```

### 6.3 Charge Pump Requirements

CMOS supply (1.0V) requires voltage boost for write:

```
Charge Pump Specifications:
- Input: 1.0V (standard CMOS supply)
- Output: ±1.5V (FeFET write voltage)
- Type: 2-stage Dickson pump
- Efficiency: 70%
- Rise time: 40ns
- Ripple: <50mV

Source: Our implementation (module4-circuits/pkg/peripherals/chargepump.go)
```

---

## 7. Paper References (Complete List)

### 7.1 Architecture & Sneak Paths

| Paper | Key Finding | Year |
|-------|-------------|------|
| *Crossbar_Sneak_Path_Analysis_arXiv* | Three-cell sneak model validated | 2023 |
| *sneak_path_self_rectifying_arrays_2022* | 10⁴:1 rectification enables 0T1R | 2022 |
| *memory_tech_crossbar_dnn_accuracy_2024* | Architecture comparison metrics | 2024 |
| *cim_landscape_overview_2024* | Industry landscape and trends | 2024 |

### 7.2 FeFET Programming

| Paper | Key Finding | Year |
|-------|-------------|------|
| *multilevel_fefet_crossbar_2023* | 32 levels with write-verify | 2023 |
| *Multi_Level_FeFET_Programming_arXiv* | 140 levels demonstrated | 2024 |
| *FeFET_Crossbar_Impact_arXiv* | Non-ideality analysis | 2023 |
| *fecap_fefet_cim_elements_2024* | FeCap vs FeFET comparison | 2024 |

### 7.3 Hardware Demonstrations

| Paper | Key Finding | Year |
|-------|-------------|------|
| *FeFET_Crossbar_MNIST_Hardware_arXiv* | 87% MNIST, 128×64 array | 2024 |
| *ferroelectric_CIM_review_2023* | Comprehensive FeCIM survey | 2023 |
| *Temperature_Resilient_FeFET_CIM_2024* | -40°C to +125°C operation | 2024 |
| *3D_FeFET_Architectures_2025* | 3D stacked arrays demonstrated | 2025 |

### 7.4 Energy & Performance

| Paper | Key Finding | Year |
|-------|-------------|------|
| *Analog_CIM_Energy_Efficiency_arXiv* | 10-1000× vs GPU for inference | 2023 |
| *HCiM_ADC_Less_2024* | ADC elimination technique | 2024 |
| *pruning_adc_efficiency_crossbar_2024* | Pruning reduces ADC needs | 2024 |

### 7.5 Primary Literature Sources

**Material Properties:**
- Nature Communications 2025: HZO Pr = 15-34 µC/cm²
- Advanced Electronic Materials 2024: Cryogenic Pr = 75 µC/cm²
- Nano Letters 2024: Ec = 0.6-1.5 MV/cm, V:HfO₂ endurance 10¹²

**Endurance:**
- IEEE IRPS 2022: 10⁹ cycles demonstrated
- Science 2024: 10¹² cycles (V:HfO₂)

**MNIST Benchmarks:**
- Nature Communications 2023: 96.6% accuracy
- ScienceDirect 2025: 98.24% (FTJ reservoir)

---

## 8. Implementation in FeCIM Visualizer

### 8.1 Module 4 Code Locations

```
module4-circuits/
├── pkg/gui/
│   ├── app.go                   # Main app with architecture state
│   ├── tab_operations.go        # Unified operations view (WRITE/READ/COMPUTE)
│   │   ├── createArchitectureToggle()  # 1T1R/0T1R toggle
│   │   ├── drawSharedArray()           # Array visualization with transistors
│   │   └── updateModeHelp()            # Architecture-aware help text
│   ├── helpers.go               # Drawing utilities (gradients, glow effects)
│   └── drawing.go               # Timing diagram utilities
│   ├── tab_operations_write.go  # WRITE mode panel
│   │   ├── tab_operations_read.go   # READ mode panel
│   │   └── tab_operations_compute.go # COMPUTE mode panel
│   ├── helpers.go               # Drawing utilities (gradients, glow effects)
│   └── drawing.go               # Timing diagram utilities
└── pkg/peripherals/
    ├── dac.go                   # 5-bit DAC model (32 levels, use 30)
    ├── adc.go                   # 5-bit SAR ADC model (32 levels, use 30)
    ├── tia.go                   # 10kΩ TIA for current sensing
    ├── chargepump.go            # ±1.5V voltage boost circuit
    ├── analysis.go              # Performance analysis utilities
    └── peripherals_test.go      # Unit tests
```

### 8.2 Architecture Toggle (Implemented)

**Location:** OPERATIONS view header, next to mode selector (WRITE/READ/COMPUTE)

**Toggle Buttons:**
- `[PASSIVE]` - 0T1R passive crossbar (default)
- `[1T1R GATE]` - Transistor-gated array

**Visual Indication:**
- Passive (0T1R): No transistor symbols shown
- 1T1R: Transistor symbols on left of each row
  - **GREEN with glow** = transistor ON (conducting)
  - **GRAY** = transistor OFF (isolated)
  - "WL" (Word Line) label above transistors

**Mode-Specific Behavior:**
- WRITE/READ: Selected row transistor ON, others OFF
- COMPUTE: ALL transistors ON for full MVM

**Help Text Updates:**
The mode help text dynamically updates based on architecture:
- 1T1R WRITE: "Transistor gates ONLY selected row"
- 1T1R READ: "Transistor isolates selected row"
- 1T1R COMPUTE: "ALL transistors ON for full MVM"
- Passive modes show sneak path error warnings (~5-20%)

---

## 9. Implementation Status & TODOs

### 9.1 Feature Status

| Feature | Status | Location | Notes |
|---------|--------|----------|-------|
| Architecture toggle (1T1R/0T1R) | ✅ Complete | `tab_operations.go:987-1040` | Fully functional with visual feedback |
| MOSFET transistor drawing | ✅ Complete | `tab_operations.go:405-491` | Green glow=ON, gray=OFF |
| Sneak path visualization | ✅ Complete | `tab_operations.go:306-333` | Faded red lines in passive mode |
| Write pulse waveform | ✅ Complete | `tab_operations_write.go:179-284` | Shows Ec threshold |
| Read zone diagram | ✅ Complete | `tab_operations_read.go:153-244` | Color-coded safety zones |
| Compute MVM animation | ✅ Complete | `tab_operations_compute.go:340-390` | Step-by-step visualization |
| TIA current conversion | ✅ Complete | `peripherals/tia.go` | 10kΩ gain, 100MHz BW |
| Charge pump modeling | ✅ Complete | `peripherals/chargepump.go` | 2-stage Dickson, ±1.5V |
| Timing diagrams | ✅ Complete | `tab_reference_timing.go` | WRITE/READ/COMPUTE waveforms |
| INL/DNL modeling | ✅ Complete | `peripherals/dac.go`, `adc.go` | Nonlinearity simulation |

### 9.2 Known Discrepancies

| Issue | Severity | Details | Resolution |
|-------|----------|---------|------------|
| GUI DAC/ADC bits default | ⚠️ Minor | `app.go:29-30` defaults to 8-bit, but peripherals use 5-bit | No functional impact; peripherals use correct 5-bit |
| Bit selector range | ℹ️ Info | UI allows 4-12 bit selection | Educational feature; 5-bit optimal for 30 levels |

### 9.3 TODOs (Future Enhancements)

| Priority | Feature | Description |
|----------|---------|-------------|
| LOW | Export functionality | UX-002: Export diagrams/data (noted in GUI.module4.md) |
| LOW | Temperature model | Add temperature-dependent INL/DNL |
| LOW | Process corners | Fast/slow/typical corner analysis |
| LOW | Write-verify animation | Show iterative programming cycle |
| LOW | Sneak path quantification | Display actual sneak current percentage |

### 9.4 Code Quality Notes

- **No TODO/FIXME comments** in codebase — all features complete
- **Unit tests** present in `peripherals/peripherals_test.go`
- **Thread safety** via `sync.RWMutex` in app state
- **Fyne.Do()** used correctly for UI updates from goroutines

---

## Appendix A: Quick Reference Card

```
┌─────────────────────────────────────────────────────────────────┐
│                    FeCIM OPERATION QUICK REFERENCE              │
├─────────────────────────────────────────────────────────────────┤
│ WRITE (1.2-1.5V, 50-100ns)                                      │
│   Passive: V/2 bias unselected, watch for disturb               │
│   1T1R:    Single row ON, full isolation                        │
├─────────────────────────────────────────────────────────────────┤
│ READ (0.1-0.5V, ~60ns total)                                    │
│   Passive: Expect 5-20% sneak path error                        │
│   1T1R:    Single row ON, clean signal                          │
├─────────────────────────────────────────────────────────────────┤
│ COMPUTE (0-1V inputs, ~75ns total)                              │
│   Passive: ALL rows active, 5-15% error                         │
│   1T1R:    ALL transistors ON, 1-3% error                       │
├─────────────────────────────────────────────────────────────────┤
│ TRANSISTOR STATE (1T1R only):                                   │
│   WRITE: Selected=ON, Unselected=OFF                            │
│   READ:  Selected=ON, Unselected=OFF                            │
│   COMPUTE: ALL=ON                                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Related Documentation

- **[circuits.research.md](circuits.research.md)** — Peripheral circuits meta-study
- **[circuits.ELI5.md](circuits.ELI5.md)** — Simple explanations + module spec
- **[circuits.opensource.md](circuits.opensource.md)** — Open-source tools
- **[../crossbar/crossbar.research.md](../crossbar/crossbar.research.md)** — Crossbar array research
- **[../development/GUI/GUI.module4.md](../development/GUI/GUI.module4.md)** — Module 4 GUI reference

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
**Source:** Meta-study of 40+ research papers on crossbar architectures and CIM operations
