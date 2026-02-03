# Research Report: Module 2 vs Module 4 Physics Comparison

> **Note:** Internal analysis note. Values are reported/illustrative and not validated by this codebase.

**Date:** 2026-01-27
**Research Goal:** Compare the physics and electronics foundations of Module 2 (Crossbar) and Module 4 (Circuits), focusing on Kirchhoff's laws for read, write, and compute operations.

---

## Executive Summary

Module 2 (Crossbar) and Module 4 (Circuits) are **complementary, not redundant** — they model different abstraction layers of the same FeCIM system using identical underlying physics.

| Aspect | Module 2 (Crossbar) | Module 4 (Circuits) |
|--------|---------------------|---------------------|
| **Focus** | Array-level analog physics | Circuit-level peripherals |
| **Domain** | Conductance matrices, MVM | DAC/ADC/TIA, voltage generation |
| **Abstraction** | High (normalized 0-1) | Low (real voltages, currents) |
| **Shared Physics** | Ohm's Law, KCL, KVL | Ohm's Law, KCL, KVL |

---

## Core Physics: Identical Foundation

Both modules are built on **two fundamental laws**:

### 1. Ohm's Law (Analog Multiplication)

```
I = G × V
```

- **Module 2**: `array.go:146` — Conductance × Input voltage → Cell current
- **Module 4**: TIA, DAC calculations — Voltage/current relationships

### 2. Kirchhoff's Current Law (Analog Accumulation)

```
I_out = Σ(G_ij × V_j)
```

- **Module 2**: `array.go:137-149` — Currents sum automatically at bitline nodes
- **Module 4**: TIA input node — Sum of all column currents

This physics enables **O(1) time matrix-vector multiplication** without active adder circuits.

---

## Operation Comparison

### READ Operation

| Stage | Module 2 | Module 4 |
|-------|----------|----------|
| Cell sensing | `I = G × V_read` | — |
| Current sum | KCL at bitline | TIA input node |
| Amplification | — | TIA: V = I × 10kΩ |
| Digitization | — | ADC: 5-bit SAR, 50ns |
| **Latency** | ~1 ns (physics) | 61 ns (circuits) |

### WRITE Operation

| Stage | Module 2 | Module 4 |
|-------|----------|----------|
| Level selection | `SetLevel(row, col, level)` | DAC: 5-bit → voltage |
| Voltage generation | Abstracted | Charge pump: ±1.5V |
| Cell programming | Instantaneous | 100ns pulse |
| **Energy** | Not modeled | 2.18 pJ/cell |

### COMPUTE (MVM) Operation

| Stage | Module 2 | Module 4 |
|-------|----------|----------|
| Input encoding | Normalized vector | DAC array (4 parallel) |
| Multiplication | I = G × V (all cells) | — |
| Accumulation | KCL (automatic) | Wire physics |
| Output sensing | Sum currents | TIA array → ADC array |
| **Latency** | ~10 ns | 75 ns total |
| **Energy** | Not modeled | 2.4 pJ (16 MACs) |

---

## Detailed Physics Analysis

### Module 2: Crossbar Array Physics

#### Fundamental Equations

1. **Ohm's Law:** `I = G × V` — Maps weight×input to current
2. **Kirchhoff's Current Law:** `I_row = Σ(G_ij × V_j)` — Automatic summation via physics
3. **Matrix-Vector Multiplication:** `y = W × x` — Realized in O(1) analog operation
4. **IR Drop:** `V_drop = I × R_wire` — Resistive voltage loss along interconnects

#### Physical Constants

| Parameter | Value | Notes |
|-----------|-------|-------|
| Quantization levels | 30 states (demo baseline; simulation baseline) | ~4.9 bits/cell |
| Conductance range | 10-100 µS | Normalized 0.0-1.0 |
| Wire resistance | 2.5 Ω/segment | Typical metal interconnect |
| Temperature coefficient | 0.00393 K⁻¹ | For drift modeling |

#### Non-Idealities Modeled

| Non-Ideality | Magnitude | Implementation |
|--------------|-----------|----------------|
| IR Drop | 5-15% (64×64 array) | `irdrop.go:71-134` |
| Sneak Paths (0T1R) | 10-100% of signal | `nonidealities.go` |
| Sneak Paths (1T1R) | 0.1% of signal | 1000× improvement |
| Device Variation | 3-10% | Gaussian model |
| Conductance Drift | 0.001 coefficient | FeFET advantage |

#### Architecture Comparison

|                      | 0T1R (Passive) | 1T1R (Active) |
|----------------------|----------------|---------------|
| Density              | 4F²            | 6-8F²         |
| Sneak isolation      | 1:1            | 1000:1        |
| IR drop multiplier   | 1.5×           | 1.0×          |
| Max practical size   | 128×128        | >1024×1024    |

### Module 4: Peripheral Circuit Physics

#### DAC (Digital-to-Analog Converter)

- **Resolution:** 5-bit (32 levels)
- **Output range:** ±1.5V
- **Settling time:** 10 ns
- **Energy:** 0.1 pJ/conversion
- **INL/DNL:** 0.5/0.25 LSB
- **Level resolution:** 96.77 mV/level

#### ADC (Analog-to-Digital Converter)

- **Architecture:** Successive Approximation (SAR)
- **Resolution:** 5-bit (32 codes)
- **Conversion time:** 50 ns
- **Energy:** 0.5 pJ/conversion
- **ENOB:** 4.77 bits (accounting for INL/DNL)

#### TIA (Transimpedance Amplifier)

- **Gain:** 10 kΩ
- **Bandwidth:** 100 MHz
- **Settling time:** 11 ns
- **Input noise:** 100 µV RMS
- **SNR:** 74 dB at mid-level currents
- **Offset:** 5 mV typical

#### Charge Pump

- **Output:** ±1.5V from 1V supply
- **Architecture:** 2-stage Dickson
- **Efficiency:** 70%
- **Rise time:** 40 ns
- **Energy:** 2.14 pJ/write (dominant cost)

---

## Key Findings

### Finding 1: Complementary Abstraction Layers

- Module 2: Models **what** happens (weights stored as conductances, MVM via physics)
- Module 4: Models **how** it happens (DAC generates voltages, TIA/ADC sense currents)
- **No code overlap** — different concerns, same physics
- **Shared interfaces:** 0 (conceptual integration only)

### Finding 2: Consistent 30-Level Quantization

Both modules use the 30-level demo baseline:

- **Module 2**: Conductance 10-100 µS (normalized 0.0-1.0)
- **Module 4**: Voltage -1.5V to +1.5V (96.8 mV/level)
- **Precision**: ~4.9 bits/cell
- **ADC margin**: 32 codes for 30 levels (2 extra for headroom)

### Finding 3: Non-Idealities Split Across Modules

| Non-Ideality | Module 2 | Module 4 |
|--------------|----------|----------|
| IR drop | ✅ Modeled (3-10%) | — |
| Sneak paths | ✅ Modeled (0T1R vs 1T1R) | — |
| Conductance drift | ✅ Modeled | — |
| DAC INL/DNL | — | ✅ Modeled |
| ADC ENOB | — | ✅ 4.77 effective bits |
| TIA noise | — | ✅ 100 µV RMS |

### Finding 4: Energy Efficiency — 67× Better Than GPU

| Metric | FeCIM (Module 2+4) | GPU |
|--------|-------------------|-----|
| 16-MAC energy | 2.4 pJ | 160 pJ |
| Per-MAC | 0.15 pJ | 10 pJ |
| Bottleneck | ADC (82%) | DRAM access |

**Mechanism:** Eliminates DRAM access (9.75 pJ/MAC) by storing weights in-situ as conductances.

### Finding 5: IR Drop Behavior Differs by Operation

| Operation | IR Drop Impact | Reason |
|-----------|----------------|--------|
| WRITE | <0.1% | Single cell conducts |
| READ | 3-10% | Many cells conduct |
| COMPUTE | 3-10% | All cells active |

### Finding 6: Latency Dominated by Peripherals

| Component | Time | Percentage |
|-----------|------|------------|
| Crossbar physics | 1-5 ns | 7% |
| TIA settling | 11 ns | 15% |
| ADC conversion | 50 ns | 67% |
| DAC settling | 10 ns | 13% |
| **Total** | 75 ns | 100% |

---

## Signal Flow: Complete Picture

### WRITE Path

```
                    WRITE PATH
    ┌──────────────────────────────────────────┐
    │  Digital Level (0-29)                     │
    │      ↓                                    │
    │  [MODULE 4 DAC] 5-bit → Voltage          │
    │      ↓                                    │
    │  [MODULE 4 Charge Pump] ±1.5V drive      │
    │      ↓                                    │
    │  [MODULE 2 Cell] Polarization switching   │
    │      ↓                                    │
    │  Conductance State (10-100 µS)           │
    └──────────────────────────────────────────┘
```

### READ/COMPUTE Path

```
                   READ/COMPUTE PATH
    ┌──────────────────────────────────────────┐
    │  Input Vector (digital)                   │
    │      ↓                                    │
    │  [MODULE 4 DAC Array] → Input voltages   │
    │      ↓                                    │
    │  [MODULE 2 Crossbar] I = G × V (all)     │
    │      ↓                                    │
    │  [KCL Physics] I_out = Σ currents        │
    │      ↓                                    │
    │  [MODULE 4 TIA] V = I × 10kΩ             │
    │      ↓                                    │
    │  [MODULE 4 ADC] 5-bit digitization       │
    │      ↓                                    │
    │  Output Vector (digital)                  │
    └──────────────────────────────────────────┘
```

### Timing Diagram (64×64 MVM)

```
Time (ns):  0    10   20   30   40   50   60   70   75
            │    │    │    │    │    │    │    │    │
DAC:        ████████████
Crossbar:        ██████
TIA:                  ████████████
ADC:                              ██████████████████████
            └────────────────────────────────────────┘
                        Total: 75 ns
```

---

## Error Budget Analysis

| Source | Module | Magnitude | Mitigation |
|--------|--------|-----------|------------|
| Cell quantization | 2 | 3.3% (1/30) | Fundamental limit |
| IR drop | 2 | 3-10% | Wider metal, tiling |
| Sneak (0T1R) | 2 | 50-100% | Use 1T1R architecture |
| Sneak (1T1R) | 2 | 0.1% | Acceptable |
| Device variation | 2 | 1-5% | Screening, calibration |
| ADC quantization | 4 | 3.1% (1/32) | Matches cell precision |
| TIA noise | 4 | 0.01% | Negligible |

**Dominant errors:** IR drop, Sneak paths, Quantization

---

## Limitations Identified

| Gap | Impact | Recommendation |
|-----|--------|----------------|
| No Module 2↔4 interface | Cannot simulate end-to-end | Create unified pipeline |
| Linear conductance model | Real FeFETs are exponential | Add nonlinear model |
| No ISPP write-verify | Multi-level accuracy limited | Implement write-verify loop |
| Temperature fixed at 300K | Cannot show cryogenic benefits | Add T-dependent parameters |
| No endurance modeling | Assumes infinite cycles | Add fatigue model |
| Sneak paths not in Module 4 | Passive mode incomplete | Add sneak path model |

---

## Recommendations

### For Large Arrays (>64×64)

1. **Mandatory**: Use 1T1R architecture (eliminates sneak)
2. **Metal optimization**: Wider lines or Cu interconnect (reduce IR drop)
3. **Tiling**: Break into 32×32 sub-arrays (shorten current paths)

### For High Throughput

1. **ADC upgrade**: Flash or pipeline (10 ns vs 50 ns) → 5× speedup
2. **TIA bandwidth**: 500 MHz target (settle < 3 ns)
3. **Parallelization**: 256+ read channels

### For Tool Integration

1. **Create unified READ path simulation**: Crossbar → TIA → ADC
2. **Add validation suite**: Test current ranges match peripheral specs
3. **Document interface**: Define physical units handoff (µA range)

---

## Conclusion

Module 2 and Module 4 implement **the same physics (Ohm's Law + Kirchhoff's Laws)** at different abstraction levels. They are designed to work together:

1. **Module 4** generates write voltages and senses read currents
2. **Module 2** models the crossbar array physics (MVM, non-idealities)
3. Together they enable **O(1) analog compute** that is **67× more energy-efficient** than GPUs

The modules share no code but share all physics — making them complementary educational tools for understanding FeCIM systems from circuit-level to array-level.

---

## References

### Peer-Reviewed Literature

| Topic | Paper | Journal | Year | DOI |
|-------|-------|---------|------|-----|
| MVM Physics | First In-Memory Computing Crossbar Using Multi-Level FeFET | Nature Communications | 2023 | [10.1038/s41467-023-42110-y](https://doi.org/10.1038/s41467-023-42110-y) |
| Energy Efficiency | Samsung FeFET for Low-Power NAND Flash | Nature | 2025 | [10.1038/s41586-025-09793-3](https://doi.org/10.1038/s41586-025-09793-3) |
| CIM Applications | Recent Advances in Ferroelectric CIM | Nano Convergence | 2025 | [10.1186/s40580-025-00520-2](https://doi.org/10.1186/s40580-025-00520-2) |
| ADC Power Dominance | ADC-less Hybrid CIM Architecture | IEEE JSSC | 2024 | See `04-cim-architectures/` |
| NeuroSim Validation | NeuroSim Simulator Validation and Benchmark | Frontiers in AI | 2021 | [10.3389/frai.2021.659060](https://doi.org/10.3389/frai.2021.659060) |
| IR Drop Modeling | badcrossbar: Python Tool for Crossbar Arrays | SoftwareX | 2020 | [10.1016/j.softx.2020.100617](https://doi.org/10.1016/j.softx.2020.100617) |

### Simulation Tool References

| Tool | Purpose | Citation |
|------|---------|----------|
| **CrossSim** | GPU-accelerated CIM simulation | SAND2021-12318C (Sandia) |
| **badcrossbar** | Exact IR drop calculation | DOI: 10.1016/j.softx.2020.100617 |
| **NeuroSim** | Device-circuit-architecture co-simulation | DOI: 10.3389/frai.2021.659060 |

### Code Locations

| Module | Component | File Path |
|--------|-----------|-----------|
| Module 2 | MVM | `module2-crossbar/pkg/crossbar/array.go:123-161` |
| Module 2 | IR Drop | `module2-crossbar/pkg/nonidealities/irdrop.go` |
| Module 2 | Sneak Paths | `module2-crossbar/pkg/nonidealities/sneak.go` |
| Module 4 | DAC | `module4-circuits/pkg/dac/dac.go` |
| Module 4 | ADC | `module4-circuits/pkg/adc/adc.go` |
| Module 4 | TIA | `module4-circuits/pkg/tia/tia.go` |

### Related Documents

- [circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md) — CIM operation physics
- [crossbar-arrays.md](crossbar-arrays.md) — Crossbar architecture synthesis
- [cim-circuits.md](cim-circuits.md) — Peripheral circuit specifications
- [MODULE4-PHYSICS-IMPROVEMENTS.md](MODULE4-PHYSICS-IMPROVEMENTS.md) — Gap analysis
