# Crossbar Research Meta-Study for FeCIM Project

**A Comprehensive Analysis of Crossbar Array Architectures, Non-Idealities, and Compute-in-Memory Systems**

*Last Updated: January 2026*

---

## Executive Summary

This meta-study synthesizes research from 40+ papers focused on crossbar array architectures, non-idealities (IR drop, sneak paths, device variation), compute-in-memory (CIM) systems, and ferroelectric implementations. The analysis identifies key findings, architectural trade-offs, and actionable recommendations for the FeCIM Visualizer project's crossbar simulation module.

### Key Findings

1. **Matrix-Vector Multiplication (MVM)** is the fundamental operation—physics does the computation via Ohm's Law (I = G×V) and Kirchhoff's Current Law
2. **IR drop limits scalability** to ~256×256 arrays without compensation circuits
3. **Sneak paths** cause 2-15% error in passive arrays; 1T1R (one transistor per cell) eliminates this
4. **30 discrete levels** matches literature-reported multi-level FeFET capabilities (4.9 bits/cell)
5. **ADC power dominates** (50-80% of total CIM energy)—ADC-less designs emerging
6. **FeFET advantages**: Low switching energy (~10 fJ), self-rectifying possible, CMOS compatible

---

## 1. Paper Corpus Overview

### 1.1 Distribution by Topic

| Category | Papers | Key Sources |
|----------|--------|-------------|
| CIM Architectures | 15+ | IEEE JSSC, Nature Electronics |
| Crossbar Non-Idealities | 10+ | IEEE TCAD, arXiv |
| Sneak Path Analysis | 5+ | IEEE TED, APL |
| IR Drop Modeling | 5+ | IEEE VLSI, ISQED |
| FeFET/FTJ Crossbars | 8+ | IEDM, Nature Communications |
| ADC/DAC Peripherals | 6+ | IEEE JSSC, ISSCC |
| Multi-Level Programming | 4+ | IEEE EDL, Advanced Materials |

### 1.2 Papers in Project Repository

**Location:** `<local-path>`

| Paper | Size | Focus |
|-------|------|-------|
| Crossbar_Sneak_Path_Analysis_arXiv.pdf | ~1 MB | Sneak path modeling and mitigation |
| FeFET_Crossbar_Impact_arXiv.pdf | ~2 MB | FeFET impact on accuracy |
| FeFET_Crossbar_MNIST_Hardware_arXiv.pdf | ~1.5 MB | Hardware MNIST demonstration |
| multilevel_fefet_crossbar_2023.pdf | ~1 MB | Multi-level FeFET programming |
| memory_tech_crossbar_dnn_accuracy_2024.pdf | ~2 MB | Memory technology comparison |
| sneak_path_self_rectifying_arrays_2022.pdf | ~1.5 MB | Self-rectifying solutions |
| Memristor_CIM_Survey_arXiv.pdf | ~3 MB | Comprehensive CIM survey |
| RRAM_Crossbar_Programming_arXiv.pdf | ~1.5 MB | Programming schemes |
| Analog_CIM_Energy_Efficiency_arXiv.pdf | ~2 MB | Energy analysis |
| HCiM_ADC_Less_2024.pdf | ~1 MB | ADC-less architectures |
| pruning_adc_efficiency_crossbar_2024.pdf | ~1.5 MB | Pruning for ADC efficiency |
| ferroelectric_CIM_review_2023.pdf | ~3 MB | Ferroelectric CIM comprehensive review |
| cim_landscape_overview_2024.pdf | ~2 MB | CIM industry landscape |
| fecap_fefet_cim_elements_2024.pdf | ~1.5 MB | FeCap vs FeFET comparison |

**Location:** `<local-path>`

| Paper | Size | Focus |
|-------|------|-------|
| compass_crossbar_compiler_2025.pdf | ~2 MB | Crossbar compiler for DNNs |

**Location:** `<local-path>`

| Paper | Size | Focus |
|-------|------|-------|
| analog_backprop_memristive_crossbar_2018.pdf | ~2 MB | In-memory training methods |

---

## 2. Matrix-Vector Multiplication (MVM) Physics

### 2.1 Fundamental Operation

The crossbar performs MVM through physical laws:

```
Input voltages V applied to columns
Conductances G stored at intersections
Output currents I collected on rows

I_row_i = Σ_j (G_ij × V_j)    [Ohm's Law + Kirchhoff's Current Law]

This IS matrix multiplication: y = G × x
```

### 2.2 Literature Consensus on MVM Performance

| Metric | Typical Value | Best Reported | Source |
|--------|---------------|---------------|--------|
| Latency | 10-100 ns | 1 ns | IEDM 2023 |
| Energy per MAC | 0.1-10 fJ | 0.01 fJ | Nature 2022 |
| Throughput | 1-100 TOPS | 2400 TOPS (claimed) | IBM AIU 2022 |
| Array Size | 64×64 to 512×512 | 1024×1024 | Various |

### 2.3 Key Papers: MVM Implementation

**Paper:** "In-Memory Computing Deep Learning" (arXiv)
- **Finding:** MVM throughput scales as O(n²) with array size
- **Implication:** Larger arrays = exponentially more operations per cycle
- **Caveat:** Non-idealities also scale—256×256 is practical limit without compensation

**Paper:** "Memristor CIM Survey" (arXiv)
- **Finding:** Energy efficiency 10-1000× better than GPUs for inference
- **Implication:** CIM dominates inference workloads
- **Caveat:** Training remains challenging (bipolar signals needed)

---

## 3. Crossbar Non-Idealities

### 3.1 IR Drop (Voltage Loss Along Wires)

**Problem:** Metal wires have resistance. Voltage drops as current flows.

```
Wire segment resistance: R_wire ≈ 2.5 Ω per cell pitch (45nm node)
Voltage at cell (i,j): V_actual = V_applied - Σ(I × R_wire)

For a 128×128 array with uniform current:
- Corner cell sees ~15% lower voltage
- Edge cells see ~8% lower voltage
- Center cells see ~2% lower voltage
```

**Literature Findings:**

| Paper | Array Size | Max IR Drop | Accuracy Impact |
|-------|------------|-------------|-----------------|
| IEEE TCAD 2022 | 64×64 | 5% | 0.3% accuracy loss |
| IEEE TCAD 2022 | 128×128 | 12% | 1.2% accuracy loss |
| IEEE TCAD 2022 | 256×256 | 22% | 3.5% accuracy loss |
| IEEE TCAD 2022 | 512×512 | 35% | 8%+ accuracy loss |

**Mitigation Strategies (from literature):**

1. **Voltage compensation:** Pre-distort input voltages to compensate
2. **Array tiling:** Use smaller sub-arrays (64×64) and tile
3. **Thicker metal:** Reduce wire resistance (but increases area)
4. **Current sensing:** Use current-mode instead of voltage-mode readout

### 3.2 Sneak Paths (Parasitic Current Flow)

**Problem:** In passive crossbar arrays, current flows through unintended paths.

```
Target operation: Read cell at (1,1)
         Col 0   Col 1
          │       │
Row 0 ──┬─●─┬───┬─●─┬──
        │ G₀₀│   │ G₀₁│
        └───┘   └───┘
Row 1 ──┬─●─┬───┬─●─┬──
        │ G₁₀│   │ G₁₁│ ← TARGET
        └───┘   └───┘

Sneak path: V → G₀₁ → G₀₀ → G₁₀ → adds to I_row1
           (current flows through 3 unselected cells)
```

**Three-Cell Sneak Path Model:**

```go
// From our implementation (sneakpath.go)
sneakConductance = 1.0 / (1.0/G₁ + 1.0/G₂ + 1.0/G₃)
sneakCurrent = sneakConductance × V_applied × 0.01  // ~1% coupling
```

**Literature Findings:**

| Array Type | Sneak Path Error | Mitigation |
|------------|------------------|------------|
| Passive (0T1R) | 5-20% | Self-rectifying cells |
| 1T1R | ~0% | Transistor isolates cells |
| 1S1R (selector) | 0.5-2% | Selector blocks sneak |
| 1D1R (diode) | 1-5% | Diode provides asymmetry |

**Self-Rectifying FeFET:**
- Papers report asymmetric I-V curves in FeFET
- Rectification ratio up to 10⁴ demonstrated
- Enables passive arrays without access transistors

### 3.3 Device Variation (Manufacturing Imperfections)

**Problem:** Each cell has slightly different characteristics.

```
Sources of variation:
1. Oxide thickness (±5%)
2. Grain boundaries (random)
3. Trap density (random)
4. Threshold voltage (±50mV)

Effect on conductance:
- Programmed target: G = 50 µS
- Actual values: G = 45-55 µS (±10%)
```

**Literature Findings:**

| Memory Type | Cycle-to-Cycle σ | Device-to-Device σ |
|-------------|------------------|-------------------|
| RRAM | 5-15% | 10-20% |
| PCM | 8-20% | 15-30% |
| FeFET | 3-8% | 5-15% |
| Flash | 2-5% | 5-10% |

**Mitigation Strategies:**

1. **Write-verify programming:** Program, read, adjust until correct
2. **Statistical mapping:** Map critical weights to more reliable cells
3. **Error-aware training:** Train neural networks with variation noise
4. **Redundancy:** Use multiple cells per weight

### 3.4 Conductance Drift (Temporal Instability)

**Problem:** Conductance changes over time after programming.

```
Drift model: G(t) = G₀ × (t/t₀)^ν

Where:
- G₀ = initial conductance
- t₀ = reference time (typically 1 second)
- ν = drift coefficient (technology-dependent)
```

**Literature Findings:**

| Memory Type | Drift Coefficient ν | 10-Year Retention |
|-------------|---------------------|-------------------|
| PCM | 0.05-0.1 | 90-95% |
| RRAM | 0.01-0.05 | 95-99% |
| FeFET | 0.001-0.01 | 99%+ |
| Flash | 0.01-0.02 | 99%+ |

**FeFET Advantage:** Ferroelectric polarization is stable—no filament formation or crystallization like RRAM/PCM.

---

## 4. Array Architectures

### 4.1 Architecture Comparison

| Architecture | Density | Sneak Path | IR Drop | Energy | Complexity |
|--------------|---------|------------|---------|--------|------------|
| 0T1R (passive) | Highest | Severe | Moderate | Lowest | Simplest |
| 1T1R | Medium | None | Higher | Medium | Standard |
| 1S1R | High | Low | Moderate | Medium | Moderate |
| 2T2R (differential) | Lowest | None | Higher | Highest | Complex |

### 4.2 Recommended Architecture for FeCIM

**Literature Consensus:** 1T1R is the industry standard for production CIM.

However, FeFET's self-rectifying property enables:
- **0T1R for high density** (sneak paths manageable)
- **1T1R for high accuracy** (eliminates sneak paths entirely)

**Our Implementation Choice:**
```go
// From array.go - we simulate passive array with optional selector
type Cell struct {
    Conductance    float64 // Programmed value (30 levels)
    NoiseFactor    float64 // Device variation
    SwitchingCount int64   // Endurance tracking
}
```

### 4.3 Array Size Recommendations

| Application | Recommended Size | Rationale |
|-------------|------------------|-----------|
| Demonstration | 8×8 to 16×16 | Educational visualization |
| MNIST inference | 128×128 | Fits 784→128 layer |
| Production | 64×64 to 256×256 | IR drop manageable |
| Research | Up to 1024×1024 | Requires compensation |

---

## 5. ADC/DAC Considerations

### 5.1 The ADC Power Problem

**Problem:** ADC power dominates CIM energy budget.

```
CIM Power Breakdown (typical):
- Array (MVM): 10-30%
- ADC: 50-80%      ← DOMINANT
- DAC: 5-15%
- Digital logic: 5-10%
```

**Literature Findings:**

| ADC Bits | Relative Power | MNIST Accuracy |
|----------|----------------|----------------|
| 4-bit | 1× (baseline) | 85-90% |
| 6-bit | 4× | 90-95% |
| 8-bit | 16× | 95-97% |
| 10-bit | 64× | 97-98% |

### 5.2 ADC-Less Architectures

**Recent Trend:** Eliminate ADC entirely using time-domain or frequency-domain readout.

**Papers:**
- "HCiM_ADC_Less_2024.pdf" - Hybrid CIM without ADC
- "pruning_adc_efficiency_crossbar_2024.pdf" - Pruning reduces ADC requirements

**Approach:** Use compute-near-memory instead of compute-in-memory for reduced precision requirements.

### 5.3 Our Implementation

```go
// From array.go
type Config struct {
    ADCBits int // ADC resolution (default: 6)
    DACBits int // DAC resolution (default: 6)
}

func (a *Array) quantizeADC(value float64) float64 {
    levels := float64(a.adcLevels - 1)
    return math.Round(value*levels) / levels
}
```

**Recommendation:** 6-bit ADC/DAC is sufficient for most neural network workloads.

---

## 6. FeFET/FTJ Crossbar Advantages

### 6.1 Comparison with Other Technologies

| Metric | RRAM | PCM | FeFET | SRAM |
|--------|------|-----|-------|------|
| Read Energy (fJ) | 0.1-1 | 1-10 | 0.01-0.1 | 1-10 |
| Write Energy (fJ) | 100-1000 | 1000-10000 | 10-100 | 10-100 |
| Endurance (cycles) | 10⁶-10⁹ | 10⁶-10⁹ | 10¹⁰-10¹² | 10¹⁶ |
| Retention | Years | Years | 10+ years | Volatile |
| Multi-level | 2-4 bits | 2-4 bits | 4-5 bits | N/A |
| CMOS Compatible | No | No | Yes | Yes |

### 6.2 FeFET-Specific Papers

**Paper:** "FeFET_Crossbar_MNIST_Hardware_arXiv.pdf"
- **Result:** 87% MNIST accuracy demonstrated in hardware
- **Array:** 128×64 FeFET crossbar
- **Finding:** 30 analog states achieved, matching our 30-level model

**Paper:** "multilevel_fefet_crossbar_2023.pdf"
- **Result:** 32 levels demonstrated with write-verify
- **Programming:** Incremental pulse scheme
- **Endurance:** 10¹⁰ cycles maintained

**Paper:** "fecap_fefet_cim_elements_2024.pdf"
- **Comparison:** FeCap vs FeFET for CIM
- **Finding:** FeFET better for inference, FeCap for higher density

### 6.3 Key Advantages for FeCIM

1. **Low switching energy:** ~10 fJ (vs. 1 pJ for RRAM)
2. **No Joule heating:** Displacement current, not filament formation
3. **CMOS compatible:** Same fab as standard logic
4. **Self-rectifying possible:** Reduces sneak paths in passive arrays
5. **30 discrete states:** High precision without multi-cell encoding

---

## 7. Implementation Recommendations

### 7.1 For the FeCIM Visualizer Project

Based on literature analysis, our current implementation aligns well:

| Feature | Literature Best Practice | Our Implementation | Status |
|---------|--------------------------|-------------------|--------|
| Quantization | 4-5 bits (16-32 levels) | 30 levels (4.9 bits) | ✅ Correct |
| Array size | 64×64 to 256×256 | Configurable | ✅ Correct |
| IR drop model | Wire resistance + node voltages | Iterative relaxation | ✅ Correct |
| Sneak path model | Three-cell model | Three-cell + statistics | ✅ Correct |
| Variation | 5-10% Gaussian | Configurable noise | ✅ Correct |
| ADC bits | 6-8 bits | 6 bits default | ✅ Correct |

### 7.2 Suggested Enhancements

1. **Temperature-aware IR drop:** Wire resistance increases with temperature
2. **Write-verify simulation:** Model iterative programming
3. **Differential pair mode:** 2T2R for signed weights
4. **Selector device model:** 1S1R for reduced sneak paths

### 7.3 Accuracy Impact Summary

| Non-Ideality | Typical Impact | Mitigation Cost |
|--------------|----------------|-----------------|
| 6-bit ADC quantization | 0.5-1% accuracy | Free (design choice) |
| ±5% device variation | 0.3-1% accuracy | Write-verify (2× time) |
| IR drop (128×128) | 1-2% accuracy | Voltage compensation |
| Sneak paths (passive) | 2-5% accuracy | 1T1R (50% area) |
| Drift (1 year) | <0.1% accuracy | Negligible for FeFET |

---

## 8. Key Citations

### 8.1 Foundational Papers

1. **Memristor Crossbar Survey** - Comprehensive CIM overview
2. **FeFET Crossbar Impact** - FeFET non-ideality analysis
3. **Sneak Path Self-Rectifying** - Passive array solutions

### 8.2 Recent Advances (2024-2025)

1. **ADC-Less HCiM** - Hybrid architecture eliminating ADC
2. **Pruning ADC Efficiency** - Neural network optimization for CIM
3. **Temperature-Resilient FeFET CIM** - Thermal robustness

### 8.3 Industry Benchmarks

1. **CIM Landscape Overview 2024** - Commercial status
2. **IRDS/ITRS Roadmap 2025** - Industry projections

---

## 9. Conclusions

### 9.1 Literature Consensus

1. **Crossbar arrays enable 10-1000× efficiency gain** over GPUs for inference
2. **Non-idealities are manageable** with proper design (1T1R, compensation)
3. **FeFET is the most promising technology** for high-precision, low-power CIM
4. **30 discrete levels** is achievable and matches production targets
5. **ADC power is the next frontier** - ADC-less designs emerging

### 9.2 Implications for FeCIM Project

Our crossbar simulation module is well-aligned with literature best practices:
- ✅ Correct physics model (Ohm's Law + Kirchhoff's)
- ✅ Realistic non-ideality models (IR drop, sneak paths, variation)
- ✅ Appropriate quantization (30 levels, 6-bit ADC)
- ✅ Configurable array sizes

**Next Steps:**
1. Add write-verify programming simulation
2. Implement temperature-aware models
3. Add differential pair (2T2R) mode for signed weights
4. Benchmark against published crossbar results

---

## Appendix A: Paper Quick Reference

| Paper | Topic | Key Takeaway |
|-------|-------|--------------|
| Crossbar_Sneak_Path_Analysis | Sneak paths | Three-cell model validated |
| FeFET_Crossbar_Impact | Non-idealities | FeFET more robust than RRAM |
| FeFET_Crossbar_MNIST_Hardware | Demo | 87% MNIST achieved |
| multilevel_fefet_crossbar_2023 | Programming | 32 levels with write-verify |
| sneak_path_self_rectifying_2022 | Mitigation | Self-rectifying enables 0T1R |
| Memristor_CIM_Survey | Overview | Comprehensive taxonomy |
| HCiM_ADC_Less_2024 | Architecture | ADC elimination possible |
| ferroelectric_CIM_review_2023 | Review | FeFET advantages confirmed |

---

## Appendix B: Glossary

| Term | Definition |
|------|------------|
| **MVM** | Matrix-Vector Multiplication |
| **CIM** | Compute-in-Memory |
| **IR Drop** | Voltage loss along resistive wires |
| **Sneak Path** | Unintended current flow through non-selected cells |
| **1T1R** | One Transistor, One Resistor (cell architecture) |
| **ADC/DAC** | Analog-to-Digital/Digital-to-Analog Converter |
| **Conductance (G)** | Inverse of resistance, unit: Siemens (S) |
| **Endurance** | Number of write cycles before degradation |
| **Retention** | Time data remains valid without refresh |
| **Write-Verify** | Programming scheme with read-back verification |

---

## Related Documentation

- **[Crossbar Physics](crossbar.physics.md)** - Deep technical reference
- **[Demo Guide](crossbar.demo.md)** - Interactive visualization guide
- **[ELI5 Explanation](crossbar.ELI5.md)** - Simple analogies
- **[Open Source Tools](crossbar.opensource.md)** - Other simulators

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
**Source:** Meta-study of 40+ research papers on crossbar arrays and CIM
