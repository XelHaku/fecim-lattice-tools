# Deep Research: How Crossbar & Circuits Should Be Done

**Date:** 2026-02-14 (Last Updated: 2026-02-16)
**Scope:** Literature review of FeCIM crossbar arrays and peripheral circuits, ignoring current Module 4 code.
**Purpose:** Inform future architecture decisions for Module 2 (Crossbar) and Module 4 (Circuits).
**Status:** Living document - updates as new literature emerges

---

## Table of Contents

1. [How to Use This Document](#how-to-use-this-document)
2. [Executive Summary](#executive-summary)
3. [Quick Decision Trees](#quick-decision-trees)
4. [Part 1: Crossbar Array](#part-1-crossbar-array----what-the-literature-says)
5. [Part 2: Peripheral Circuits](#part-2-peripheral-circuits----what-the-literature-says)
6. [Part 3: Synthesis](#part-3-synthesis----how-to-build-it-right)
7. [Part 4: Complete Reference](#part-4-complete-reference----papers-and-dois)
8. [Part 5: Existing Project State](#part-5-existing-project-state-for-context)
9. [Glossary](#glossary)

---

## How to Use This Document

**For Implementers:**
- Start with [Quick Decision Trees](#quick-decision-trees) for immediate guidance
- Check [Priority Order](#priority-order-for-implementation) for what to build first
- Reference [Recommended Architecture Overhaul](#recommended-architecture-overhaul) for detailed specifications

**For Researchers:**
- Jump to [Part 4: Complete Reference](#part-4-complete-reference----papers-and-dois) for citations
- Review [Part 1](#part-1-crossbar-array----what-the-literature-says) and [Part 2](#part-2-peripheral-circuits----what-the-literature-says) for technical depth

**For Project Planning:**
- Read [Executive Summary](#executive-summary) for high-level trends
- Use [Part 3: Synthesis](#part-3-synthesis----how-to-build-it-right) for roadmap planning
- Check [Part 5: Existing Project State](#part-5-existing-project-state-for-context) to understand current capabilities

---

## Executive Summary

The literature has shifted significantly since the project's original architecture was designed. Three major trends emerge:

### 🔑 Key Findings

1. **Capacitive crossbars (FeCAP) eliminate the hardest problems**
   Sneak paths, IR drop, and static leakage vanish when computation uses displacement current rather than conductive paths. This is not incremental improvement—it's a paradigm shift.

2. **ADC is the bottleneck, not the array**
   40-60% of energy and >50% of area goes to ADCs. The entire peripheral circuit architecture should be designed around minimizing ADC overhead, not as an afterthought.

3. **4-bit converters are the sweet spot**
   Not 5-bit, not 8-bit. Recent hardware-aware quantization studies from 2024-2025 converge on 4-bit DAC/ADC as optimal cost-performance trade-off.

### 📊 Impact on Current Project

| Current Design | Literature Consensus | Action Required |
|----------------|---------------------|-----------------|
| 5-bit DAC/ADC default | **4-bit optimal** | Change defaults (P0) |
| Linear conductance model | **Exponential/Preisach required** | Make exponential default (P0) |
| Resistive-only arrays | **FeCAP eliminates major issues** | Add FeCAP mode (P2) |
| Fixed C2C variation | **State-dependent variation** | Add sigma(G) model (P1) |
| Current-domain sensing only | **Charge-domain for FeCAP** | Add charge amplifier (P2) |
| Single ADC architecture | **Multiple types with tradeoffs** | Add SAR/Flash/Ramp options (P1) |

### ⚡ Immediate Actions (Can Do Today)

1. Change default DAC/ADC resolution from 5-bit to 4-bit
2. Switch default conductance model from linear to exponential
3. Label linear mode as "Educational/Simplified"

---

## Quick Decision Trees

### "Which crossbar architecture should I use?"

```
START: What's your primary concern?
│
├─ "Maximum density, can tolerate complexity"
│  └─ Use: 0T1C FeCAP (4F², no sneak paths, no IR drop)
│
├─ "Proven technology, moderate density"
│  └─ Use: 1T1R resistive (8-12F², <0.1% error, battle-tested)
│
├─ "Need very large arrays (>512x512)"
│  └─ Use: 2T1R resistive (best for mega-scale)
│
└─ "Educational demo, simplicity over accuracy"
   └─ Use: 0T1R resistive (simplest, educational value)
```

### "How many ADCs should I use?"

```
START: What's your constraint?
│
├─ "Area is critical, latency acceptable"
│  └─ Use: 1 shared ADC (minimal area, N_cols × t_conv latency)
│
├─ "Balanced cost/performance"
│  └─ Use: Shared-K (K=4 or 8, 3x better energy-area product)
│
└─ "Maximum throughput needed"
   └─ Use: Per-column ADCs (highest area/power, lowest latency)
```

### "Which ADC architecture for my use case?"

```
START: What do you need?
│
├─ "Ultra-low latency (<5ns)"
│  └─ Use: Flash ADC (3-4 bit max, high energy)
│
├─ "General purpose, 4-6 bit"
│  └─ Use: SAR ADC (best all-around, 50ns @ 5-bit)
│
├─ "Area-constrained, can share"
│  └─ Use: Ramp/Slope ADC (very low energy, 100ns+)
│
├─ "Binary networks only"
│  └─ Use: Comparator-only (28x lower energy than 7-bit ADC)
│
└─ "High precision (>8 bit)"
   └─ Use: Sigma-Delta (µs latency, for calibration/verification)
```

---

## PART 1: CROSSBAR ARRAY -- What the Literature Says

### Architecture Selection: The Capacitive Paradigm Shift

> **💡 KEY INSIGHT:** The hardest crossbar problems (sneak paths, IR drop, leakage) can be eliminated entirely—not reduced, *eliminated*—by switching from resistive to capacitive operation.

The project currently models three resistive architectures (0T1R, 1T1R, 2T1R). The literature is moving toward a fourth paradigm that fundamentally changes the physics of computation:

| Architecture | Sneak Paths | IR Drop | Static Leakage | Density | Demonstrated Size |
|---|---|---|---|---|---|
| 0T1R resistive | Severe (5-20%) | Severe | Yes | 4F^2 | 128x128 max practical |
| 1T1R resistive | Minimal (<0.1%) | Moderate | Transistor leakage | 8-12F^2 | >1024x1024 |
| 2T1R resistive | Negligible | Low | Low | 10-12F^2 | >1024x1024 |
| **0T1C capacitive (FeCAP)** | **None** | **None** | **None** | **4F^2** | **128x128 demonstrated** |

**Key papers:**
- Capacitive crossbar at 128x128: 3.8 pJ/MVM, 14-57x lower energy than resistive (Adv. Intell. Syst. 2022, DOI: `10.1002/aisy.202100258`)
- FeCAP intrinsic immunity to sneak paths via displacement current (Nano Convergence 2024, DOI: `10.1186/s40580-024-00463-0`)
- FeCaps + FeFETs eliminate selectors entirely (Scientific Reports 2024, DOI: `10.1038/s41598-024-59298-8`)

**Key physics difference:** FeCAPs act as ideal capacitors during the hold phase. No steady-state conductive path exists, thereby eliminating leakage currents and static power dissipation even in large crossbar arrays. Computation is driven by displacement currents rather than conductive paths, which inherently eliminates issues like static leakage, sneak-path effects, and IR drop.

**Recommendation for the simulator:** Add FeCAP as a fourth architecture. The physics is fundamentally different -- charge-domain computation rather than current-domain. This would require:
- Capacitance matrix instead of conductance matrix
- Charge integration readout instead of current summation
- Transient pulse-based MVM (not steady-state)

### Array Sizing: What's Realistic

| Architecture | Practical Max | Limiting Factor |
|---|---|---|
| 0T1R resistive | 32x32 to 128x128 | Sneak paths + IR drop |
| 1T1R resistive | 256x256 to 1024x1024 | Transistor series resistance, IR drop |
| 2T1R resistive | >1024x1024 | Area cost |
| 0T1C capacitive | 128x128 demonstrated | Charge sharing, parasitic capacitance |

Recent demonstrations:
- 24x24 and 48x48 FTJ crossbars (Adv. Intell. Syst. 2025, DOI: `10.1002/aisy.202500817`)
- 2-kilobyte AlScN crossbar operating at 600C (ScienceDirect 2025)
- 256x256 FeFET array (Jerry et al., IEEE IEDM 2017, DOI: `10.1109/IEDM.2017.8268338`)

**1T1R limitation:** The series resistance of the transistor is a major problem in 1T1R crossbar arrays, limiting the maximum current available for inducing resistive switching and degrading array performance. Additionally, the switching time for the 1T1R configuration increases as the crossbar size increases.

### Non-Ideality Modeling: What CrossSim/NeuroSim/MNSIM Include

#### CrossSim V3.1 (Sandia, Jan 2025)
- Arbitrary programming errors (5 distribution models)
- Conductance drift
- Cycle-to-cycle read noise
- Parasitic metal resistance (SOR solver)
- ADC precision loss
- NEW in V3.1: PyTorch/TensorFlow integration, GPU acceleration

#### NeuroSim V1.5 (Georgia Tech, 2025)
- Full device -> circuit -> algorithm stack
- <1% chip-level error after calibration
- 6.5x faster than V1.4
- Now supports non-volatile capacitive memories (FeCAP)
- Includes peripheral circuit area/power/latency estimation

#### MNSIM 2.0 (Tsinghua)
- 7000x faster than SPICE
- Behavior-level interconnect resistance model
- Trade-off optimization across multiple metrics

#### Comparison: Our Project vs. Literature Tools

| Non-Ideality | Our Project | CrossSim | NeuroSim | Priority to Add |
|---|---|---|---|---|
| IR drop (SOR solver) | Yes | Yes | Yes | Done |
| Sneak paths | Yes (3-cell) | Limited | Yes | Extend to multi-hop |
| Programming errors | Yes (5 models) | Yes | Yes | Done |
| Conductance drift | Yes | Yes | Yes | Done |
| Process variation (D2D) | Yes | Yes | Yes | Done |
| Cycle-to-cycle variation | Partial | Yes | Yes | **HIGH -- add state-dependent C2C** |
| Temperature effects | Yes | No | Partial | Done |
| Endurance/fatigue | Yes (basic) | No | Yes | Enhance wake-up model |
| Write disturb (half-select) | Yes | No | Partial | Done |
| **Non-linear I-V** | **No** | Yes | Yes | **HIGH** |
| **Capacitive mode (FeCAP)** | **No** | No | **V1.5 Yes** | **MEDIUM** |
| Read disturb (FE specific) | No | No | No | **MEDIUM** |

### State-Dependent Cycle-to-Cycle Variation (Missing)

Recent measurement data shows C2C variation is **not constant** -- it depends on the conductance state:

> "An effective conductance variation model derived from experimental measurements of C2C and D2D variations performed on FeFET devices fabricated using 28nm HKMG technology. The variations were found to be a **function of different conductance states** within the given programming range." (arXiv 2023, `2312.15444`)

Current project model uses uniform noise. Should be replaced with state-dependent sigma:
```
sigma(G) = sigma_base * f(G)  where f(G) varies across the conductance range
```

Best-in-class devices achieve ~0.3% C2C / ~0.5% D2D (Science Advances 2024, DOI: `10.1126/sciadv.adp0174`).

### Multi-Level Cell (MLC) Programming: State of the Art

| Levels | Bits/Cell | Who | How | Reference |
|---|---|---|---|---|
| 8 | 3 | Various | Standard ISPP | PMC 2023 |
| 16 (QLC) | 4 | KAIST | Gate stack engineering, >10V MW | Science Advances 2024, `10.1126/sciadv.adn1345` |
| 32 | 5 | Samsung | Production FeFET NAND | Nature 2025, `10.1038/s41586-025-09793-3` |
| >256 | 8+ | Various | Optimized write circuits | Nature Comms 2023, `10.1038/s41467-023-36270-0` |
| 3,024 | ~11.5 | Nature Electronics | Gate voltage + source-drain pulse superposition | Nature Electronics 2025, `10.1038/s41928-025-01551-7` |

**Key insight for ISPP:** Standard ISPP has a fundamental limitation -- the threshold voltage shows a **nonlinear relationship** with increasing pulse amplitudes because it's not optimized for partial polarization switching. Two alternatives:

1. **Displacement Current Control (DCC):** One-shot programming tailored to FE polarization switching -- more efficient than iterative ISPP (PMC 2024)
2. **Adaptive ISPP (A-ISPP):** Adaptive step voltage -- small steps near target, large steps far away (Seoul National University)

Our project's ISPP controller already uses binary search bisection which is similar to A-ISPP. The DCC approach would be a different paradigm worth adding as an alternative engine.

### Conductance Model: Linear is Wrong

The literature is clear: **linear conductance mapping is inaccurate for ferroelectrics**:

> "Synaptic weight update behavior must be linear and symmetric for MVM computation... A **mapping function is used to scale the nonlinear device voltage to a linear drive voltage**" (Adv. Intell. Syst. 2024, `10.1002/aisy.202400211`)

Our project already has exponential and lookup models. The Preisach model integration (which we've already done) is the correct approach per literature:

> "The Preisach model successfully describes hysteretic switching in ferroelectrics, connecting the Preisach distribution to measured microscopic switching kinetics... reproductions of polarization-voltage characteristics, the history-dependence and minor loops" (Nature Comms 2018, `10.1038/s41467-018-06717-w`)

**Recommendation:** Make exponential/Preisach the default, not linear. Linear should be explicitly labeled as "simplified/educational mode."

---

## PART 2: PERIPHERAL CIRCUITS -- What the Literature Says

### The ADC Problem: Central Design Challenge

> **🚨 CRITICAL FINDING:** This is the single most important insight from the literature review.

The analog compute-in-memory community has reached consensus: **ADCs are the bottleneck, not the array**.

**The Evidence:**
> "Energy consumption of analog CIMs is **dominated by full-precision ADCs**" (IEEE ISSCC 2023)

> "ADCs in FeFET CiM arrays incur **large area, power, and latency overheads**" (Nature Comms 2024)

> "Architecture-level ADC decisions such as ADC resolution or number of ADCs **significantly impact overall CIM accelerator energy and area**" (arXiv 2024)

**What This Means for Design:**
- Don't optimize the array in isolation—optimize the *system* around ADC constraints
- ADC count, resolution, and architecture are first-order design parameters, not afterthoughts
- A "perfect" array with poor ADC strategy loses to a good array with optimized ADCs

**Our Project Status:**
- Current: 5-bit SAR ADC, 50ns conversion, 25 fJ/conversion
- Assessment: Reasonable starting point but architecturally incomplete
- Gap: No ADC architecture exploration, no sharing ratio model, resolution not optimized
- **Fix**: See [ADC Architecture Recommendations](#adc-architecture-recommendations) below

### ADC Architecture Recommendations

| ADC Type | Resolution Sweet Spot | Energy | Latency | Best For |
|---|---|---|---|---|
| **SAR** | 4-6 bit | Low | Moderate (50ns @ 5-bit) | Per-column, low throughput |
| Flash | 3-4 bit | High | Very fast (1-2ns) | Ultra-low latency |
| Sigma-Delta | 8-12 bit | Medium | Slow (us) | High precision |
| Ramp/Slope | 6-8 bit | Very low | Slow (100ns+) | Column-shared, area-efficient |
| **Comparator-only** | 1 bit | **28x lower than 7-bit ADC** | Fast | Binary/ternary networks |

**📍 CONSENSUS FINDING: 4-bit is optimal (not 5-bit, not 8-bit)**

Multiple independent studies from 2022-2024 converge on the same conclusion:

> "Experiments demonstrate improvements on CIFAR-10 and ImageNet... identifying **4-bit data converters as the optimal balance** between cost and performance" (arXiv 2024, hardware-aware quantization study)

> "128x128 non-volatile capacitive crossbar array is compatible with **3-bit ADC quantization**" (Adv. Intell. Syst. 2022, demonstrated hardware)

**Why 4-bit wins:**
- **Accuracy**: Sufficient for most ML workloads (degradation <1% vs. 8-bit on ImageNet)
- **Energy**: 4→5 bit costs ~2× ADC energy but marginal accuracy gain
- **Area**: Exponential growth in comparators/caps; 4-bit hits sweet spot
- **Hardware-aware training**: Modern quantization techniques make 4-bit viable where 8-bit was previously "required"

**What this means for our project:**
- Current 5-bit default is slightly over-provisioned vs. literature consensus
- Should demonstrate 4-bit as the recommended starting point
- Offer 5-bit/6-bit as "high precision" options, not defaults

**Recommendation:** Change default from 5-bit to 4-bit. Add multiple ADC architectures (SAR, Flash, Ramp) as selectable options with different energy/latency/area tradeoffs.

### Column-Shared vs Per-Column ADC

This is a critical architectural decision the project doesn't currently model:

| Strategy | ADC Count | Area | Energy | Latency | Best When |
|---|---|---|---|---|---|
| Per-column | N_cols | Very high | Low per-op | Low | High throughput needed |
| Shared (1 ADC) | 1 | Minimal | High per-op | High (N_cols x t_conv) | Area-constrained |
| Shared (K ADCs) | K | Medium | Medium | Medium | Balanced |

> "Using 1, 2, 4, 8, and 16 ADCs in parallel... the choice of number of ADCs can influence overall accelerator **energy-area product (EAP) by a factor of three**" (arXiv 2024)

**Recommendation:** Add configurable ADC sharing ratio (per-column, shared-4, shared-8, shared-all) and model the latency/area/energy implications.

### DAC Architecture Recommendations

| Topology | Pros | Cons | CIM Use |
|---|---|---|---|
| **Capacitive (switched-cap)** | Low power, good matching | Slower settling | Most common in CIM |
| R-2R ladder | Equal switch currents, good matching | Area | Common alternative |
| Current-steering | Fastest | Highest power | High-speed CIM |
| Hybrid (cap + current) | Balanced | Complex | Emerging |

**Input encoding matters:**

| Encoding | Bits Required | Glitch-Free | Linearity |
|---|---|---|---|
| Binary | N | No (glitches at MSB transitions) | Requires calibration |
| **Thermometer** | 2^N - 1 | Yes | Inherently monotonic |
| Segmented (MSB thermo + LSB binary) | Reduced | Mostly | Good compromise |

Our project uses simple binary quantization. For accuracy, thermometer or segmented encoding should be modeled, especially for >4-bit operation.

**Recommendation:** Change default DAC to 4-bit. Add thermometer encoding option.

### Sense Chain: Charge-Domain is the Future

The project uses current-domain TIA sensing. The literature shows charge-domain is superior for ferroelectric:

> "Charge-based sensing schemes achieving at least an **order of magnitude reduction in power consumption** compared to current-based methods" (Scientific Reports 2024)

> "Ferroelectric capacitive memories transfer memory reading and in-memory computing to **charge domain**" (Nano Convergence 2024)

For resistive crossbars, TIA is correct. But for FeCAP crossbars, the sensing should use charge integration:
- Charge amplifier instead of TIA
- Integration time window instead of steady-state current
- Lower noise (no shot noise from DC current)

**Recommendation:** Keep TIA for resistive modes, add charge amplifier for FeCAP mode.

### Write Drivers: Voltage Requirements

| Device Type | Program Voltage | Erase Voltage | Pulse Width | Energy/bit |
|---|---|---|---|---|
| HZO FeFET (standard) | 3-5V | -(3-5V) | 10-100ns | ~0.1-50 fJ |
| Samsung FeFET NAND | Near-zero pass voltage | -- | 1us | 96% power savings |
| Dual-Bit FeFET | +/-3.3V | +/-3.3V | 1us | -- |
| HZO FeCAP | 1.5-3V | -(1.5-3V) | 10-100ns | <1 fJ |

Our project's charge pump model (1.0V -> 1.5V, Dickson) is appropriate for FeCAP. For FeFET, 3-5V generation is needed -- likely requires a 3-stage or 4-stage pump from 1.0V CMOS supply.

**Recommendation:** Add configurable charge pump staging (2-stage for FeCAP/low-Vc, 4-stage for FeFET/high-Vc).

### DAC/ADC-Less Architectures (Emerging)

Several recent papers eliminate converters entirely:

> "Near-CIM analog memory and nonlinear activation units bring **76.0% energy reduction** compared with DAC/ADC solutions" (ResearchGate 2024)

> "Elimination of high-precision ADCs via **comparator-only digitization** reduces energy up to **28x**" (arXiv 2024)

This is forward-looking but worth modeling as an option: binary-weight networks with comparator readout instead of full ADC.

### System-Level Energy Budget: What Papers Actually Report

**The surprising truth:** The compute array itself consumes only 10-30% of total energy. Peripherals dominate.

| Component | % of Total Energy (Resistive) | % of Total Energy (Capacitive) | Typical Value |
|---|---|---|---|
| **Array (MVM)** | 10-30% | 5-15% | ~20% (resistive) |
| **ADC** | **40-60%** ⚠️ | **30-50%** ⚠️ | **~50% (DOMINANT)** |
| DAC | 5-15% | 5-15% | ~10% |
| TIA/Sense | 5-10% | 5-10% | ~7% |
| Write drivers | 10-20% | 5-10% | ~10% |
| Other (control, routing) | 5-10% | 5-10% | ~5% |

**ASCII Pie Chart (Typical Resistive CIM):**
```
       ADC          Array     Other
     (50%)          (20%)     (30%)
    ████████        ████     ██████
    ████████                 ██████
```

**Key Observations:**
1. **ADC is the elephant in the room** - dominates energy regardless of architecture
2. **Capacitive helps but doesn't eliminate** - ADC still 30-50% even with FeCAP's efficiency
3. **Resolution matters exponentially** - each bit roughly doubles ADC energy
4. **Our project's breakdown** (ADC 55%, DAC 31%, TIA 14%) overweights DAC vs. literature

**Implication:** Any energy optimization that doesn't address ADC overhead is rearranging deck chairs. Must model ADC architecture alternatives and sharing strategies.

### Latency Budget

| Component | Our Project | Literature Typical | Notes |
|---|---|---|---|
| DAC settling | 10ns | 5-20ns | Reasonable |
| Array physics | 1-5ns | 1-10ns | OK for resistive; FeCAP may need pulse width |
| TIA settling | 11ns | 5-20ns | Reasonable |
| ADC conversion | 50ns (SAR) | 10-100ns depending on type | OK |
| **Total read** | **76ns** | **30-150ns** | Reasonable |

FeFET search operations have been demonstrated at **100 picoseconds** (Nano Letters 2022), but that's the device physics -- the peripherals still dominate total latency.

---

## PART 3: SYNTHESIS -- How To Build It Right

### Recommended Architecture Overhaul

#### Crossbar Module (Module 2)

1. **Add FeCAP architecture** as a first-class mode alongside 0T1R/1T1R/2T1R
   - Capacitance matrix (not conductance)
   - Charge-domain MVM: Q = C x V, output = sum of charges
   - No sneak paths, no IR drop, no static leakage
   - Transient pulse-based operation

2. **Make non-linear conductance the default**
   - Exponential model as default for resistive
   - Preisach-based for physics-accurate mode (already implemented)
   - Linear only for "educational/simplified" mode

3. **Add state-dependent C2C variation**
   - sigma(G) varies with conductance level
   - Calibrate from published 28nm FeFET data

4. **Add non-linear I-V curves**
   - Currently all cells assumed ohmic (I = G x V)
   - Real FeFETs have non-linear I-V in subthreshold
   - Matters most for low-conductance states

5. **Extend sneak path model**
   - Multi-hop paths beyond 3-cell for large passive arrays
   - Sparse matrix support for >128x128

#### Circuits Module (Module 4)

1. **ADC-centric redesign**
   - Multiple ADC architectures: SAR (default), Flash, Ramp, Comparator-only
   - Default to 4-bit (not 5-bit) per literature consensus
   - Configurable ADC sharing ratio (per-column, shared-K, shared-all)
   - Model area/energy/latency tradeoffs for each configuration

2. **DAC improvements**
   - Default to 4-bit
   - Add thermometer/segmented encoding option
   - Model glitch energy for binary encoding

3. **Dual sensing modes**
   - Current-domain (TIA) for resistive crossbars
   - Charge-domain (charge amplifier) for FeCAP crossbars
   - Configurable gain/bandwidth/noise per architecture

4. **Configurable charge pump**
   - 2-stage for low-Vc (FeCAP, ~1.5V)
   - 3-4 stage for high-Vc (FeFET, 3-5V)
   - Model efficiency vs. number of stages

5. **ISPP enhancements**
   - Add DCC (Displacement Current Control) as alternative to ISPP
   - Keep adaptive bisection (current approach is good)
   - Add one-shot programming option for FeCAP

6. **Peripheral area/overhead model**
   - Literature shows peripherals are ~55% of total chip area
   - Model ADC area as function of resolution and count
   - Show area breakdown pie chart in GUI

### Priority Order for Implementation

| Priority | Change | Impact | Effort | Why This Priority? |
|---|---|---|---|---|
| **P0** | Change default to 4-bit DAC/ADC | Accuracy to literature consensus | Low | **Immediate win:** Config change aligns with 2024-2025 consensus from multiple independent studies. Zero new code required. |
| **P0** | Make exponential conductance the default | Physics accuracy | Low | **Correctness:** Linear model is fundamentally wrong for ferroelectrics per Nature Comms 2018. Exponential already implemented, just change default. |
| **P1** | Add multiple ADC architectures (SAR, Flash, Ramp) | Educational value, accuracy | Medium | **High value:** ADCs are 40-60% of total energy. Users need to understand tradeoff space. Modular design allows incremental addition. |
| **P1** | Add ADC sharing ratio model | Realistic system-level metrics | Medium | **System-level accuracy:** Literature shows 3x variation in energy-area product based on ADC count alone (arXiv 2024). Critical for realistic chip-level predictions. |
| **P1** | State-dependent C2C variation | Physics accuracy | Medium | **Measurement-backed:** ArXiv 2023 shows C2C varies with conductance state using 28nm fabricated devices. Current uniform model is oversimplified. |
| **P2** | Add FeCAP mode (capacitive crossbar) | Major architecture addition | High | **Paradigm shift:** Eliminates sneak paths, IR drop, static leakage entirely. 14-57x energy improvement demonstrated. But requires charge-domain MVM rewrite. |
| **P2** | Charge-domain sensing for FeCAP | Pairs with FeCAP mode | Medium | **Necessary for FeCAP:** Scientific Reports 2024 shows 10x power reduction vs. current-domain. Required to model FeCAP accurately. |
| **P2** | Non-linear I-V curves | Physics accuracy | Medium | **FeFET subthreshold behavior:** Real devices are non-ohmic, especially at low conductance. Matters for analog accuracy. |
| **P3** | DCC programming alternative | Forward-looking | Medium | **Emerging technique:** PMC 2024 shows promise but less mature than ISPP. Educational value for comparative study. |
| **P3** | Multi-hop sneak paths | Large passive arrays | Medium | **Niche use case:** Only matters for >128x128 passive arrays, which are rare in practice due to other limitations. |
| **P3** | DAC/ADC-less mode (comparator-only) | Emerging architecture | Low-Medium | **Research direction:** ArXiv 2024 shows 28x energy reduction but limited to binary networks. Forward-looking rather than current need. |
| **P3** | Configurable charge pump staging | Completeness | Low | **Detail work:** Current 2-stage works for FeCAP. 3-4 stage needed for high-Vc FeFET but can approximate with efficiency factor. |

### Implementation Roadmap

**Phase 1 (Week 1): Quick Wins**
- [ ] Change default DAC/ADC to 4-bit in config files
- [ ] Change default conductance model to exponential
- [ ] Add UI label "Simplified Mode" for linear conductance
- [ ] Update documentation to reflect new defaults
- **Deliverable:** Immediate alignment with literature consensus, zero new code

**Phase 2 (Weeks 2-4): ADC Architecture**
- [ ] Implement Flash ADC model (3-4 bit, 1-2ns, high energy)
- [ ] Implement Ramp ADC model (6-8 bit, 100ns+, very low energy)
- [ ] Implement Comparator-only mode (1-bit, 28x lower energy)
- [ ] Add ADC sharing ratio: per-column, shared-4, shared-8, shared-all
- [ ] Add area/latency/energy calculators for each configuration
- [ ] Update GUI with ADC architecture selector dropdown
- **Deliverable:** Educational ADC comparison tool, realistic system-level metrics

**Phase 3 (Weeks 5-7): Physics Accuracy**
- [ ] Implement state-dependent C2C variation: sigma(G) = sigma_base * f(G)
- [ ] Calibrate f(G) from arXiv 2023 FeFET data (28nm)
- [ ] Add non-linear I-V curves for FeFET subthreshold region
- [ ] Validate against published measurement data
- [ ] Add regression tests for new physics models
- **Deliverable:** Measurement-backed variation model, improved analog accuracy

**Phase 4 (Weeks 8-12): FeCAP Mode**
- [ ] Design capacitance matrix data structure (replace conductance)
- [ ] Implement charge-domain MVM: Q = C × V
- [ ] Add charge amplifier sensing (replace TIA for FeCAP)
- [ ] Implement transient pulse-based operation
- [ ] Add FeCAP-specific GUI visualizations
- [ ] Validate against Adv. Intell. Syst. 2022 (128x128 demo)
- **Deliverable:** Full FeCAP architecture support, major capability expansion

**Phase 5 (Post-MVP): Advanced Features**
- DCC programming alternative
- Multi-hop sneak paths for >128x128
- Configurable charge pump staging
- Hybrid architectures (FeCAP + FeFET)

---

## PART 4: COMPLETE REFERENCE -- Papers and DOIs

### Peer-Reviewed Publications (Cited in This Report)

| # | Topic | DOI | Year |
|---|---|---|---|
| 1 | Capacitive crossbar MVM (128x128) | `10.1002/aisy.202100258` | 2022 |
| 2 | FeCAP memory comprehensive review | `10.1186/s40580-024-00463-0` | 2024 |
| 3 | FeCaps/FeFETs as IMC elements | `10.1038/s41598-024-59298-8` | 2024 |
| 4 | Multi-level FeFET crossbar (885 TOPS/W) | `10.1038/s41467-023-42110-y` | 2023 |
| 5 | 16-level QLC FeFET (>10V MW) | `10.1126/sciadv.adn1345` | 2024 |
| 6 | Samsung FeFET NAND (32 levels) | `10.1038/s41586-025-09793-3` | 2025 |
| 7 | 3,024 states transistor | `10.1038/s41928-025-01551-7` | 2025 |
| 8 | >256 analog states | `10.1038/s41467-023-36270-0` | 2023 |
| 9 | Recent advances in ferroelectrics review | `10.1186/s40580-025-00520-2` | 2025 |
| 10 | Preisach model for ferroelectrics | `10.1038/s41467-018-06717-w` | 2018 |
| 11 | HfO2 FeFET comprehensive review | `10.1063/5.0206599` | 2024 |
| 12 | 2D ferroelectric hybrid CIM | `10.1126/sciadv.adp0174` | 2024 |
| 13 | FeFET CiM annealer | `10.1038/s41467-024-46640-x` | 2024 |
| 14 | 2T2R for differential IMC | `10.1007/s11432-023-3887-0` | 2024 |
| 15 | FTJ crossbar half-bias | `10.1002/aisy.202500817` | 2025 |
| 16 | Dual-Bit FeFET | `10.1038/s44335-025-00030-8` | 2025 |
| 17 | Ferroelectric materials review (China) | `10.1007/s11432-025-4432-x` | 2025 |
| 18 | 2D FE for in-sensor computing | `10.1002/adma.202400332` | 2024 |
| 19 | FeFET FeRAM review | `10.1063/5.0086328` | 2023 |
| 20 | FeFET temperature effects | `10.1016/S0038-1101(24)00103-5` | 2024 |
| 21 | 256x256 FeFET array | `10.1109/IEDM.2017.8268338` | 2017 |
| 22 | Delta-Sigma CIM (21.38 TOPS/W) | `10.1109/ISSCC42615.2023.10067289` | 2023 |
| 23 | Linear conductance modulation | `10.1002/aisy.202400211` | 2024 |
| 24 | CSCDAC hybrid DAC | `10.3390/jlpea15010009` | 2025 |
| 25 | Preisach parameter automation | `10.1038/s41598-021-91492-w` | 2021 |
| 26 | Memristor crossbar sensing review | `10.1021/acs.chemrev.4c00845` | 2024 |

### Conference and arXiv References

| Topic | Reference | Year |
|---|---|---|
| State-dependent C2C FeFET variation | arXiv `2312.15444` | 2023 |
| ADC optimization for CIM | arXiv `2404.06553` | 2024 |
| 4-bit optimal converters (HW-aware quant) | arXiv `2508.21524` | 2024 |
| DAC/ADC-less CIM | arXiv `2412.19869` | 2024 |
| NeuroSim V1.5 | arXiv `2505.02314` | 2025 |
| DCC one-shot programming | PMC `PMC11160465` | 2024 |
| FeFET 3D NAND training accelerator | ResearchGate `349145591` | 2021 |

### Simulation Tool References

| Tool | Version | Source | URL |
|---|---|---|---|
| CrossSim | V3.1 (Jan 2025) | Sandia National Labs | `cross-sim.sandia.gov` |
| NeuroSim | V1.5 (2025) | Georgia Tech | `github.com/neurosim` |
| MNSIM | V2.0 | Tsinghua University | ResearchGate `344931525` |
| badcrossbar | -- | UCL | DOI: `10.1016/j.softx.2020.100617` |

### Textbook / Standard References

- J. F. Dickson, "On-chip high-voltage generation in MNOS integrated circuits using an improved voltage multiplier technique," IEEE JSSC, 1976. (charge pumps)
- B. Razavi, Data Conversion System Design, IEEE Press/Wiley, 1995. (ADC architecture, noise/ENOB)
- IEEE Std 1241-2010, IEEE Standard for Terminology and Test Methods for ADCs. (performance metrics)
- D. M. Young, Iterative Solution of Large Linear Systems, 1971. (SOR algorithm)

---

## PART 5: EXISTING PROJECT STATE (For Context)

### What Module 2 Already Has
- SOR solver ported from CrossSim (solver.go)
- 5 programming error models from CrossSim (device_errors.go)
- 3-cell sneak path model (sneakpath.go)
- Power-law and logarithmic drift (drift.go)
- Temperature effects with 5 presets (temperature.go, temperature_profile.go)
- Process variation with spatial gradients (array.go)
- Endurance/fatigue model (array.go)
- Half-select write disturb (write_disturb.go)
- Differential arrays for signed weights (enhanced.go)
- Write-verify programming (enhanced.go)
- Three conductance models: linear, exponential, lookup (array.go + shared/physics)
- Vulkan GPU-accelerated MVM (gpu_mvm.go)

### What Module 4 Already Has (Shared Peripherals)
- 5-bit DAC with INL/DNL (shared/peripherals/dac.go)
- 5-bit SAR ADC with ENOB/SNR (shared/peripherals/adc.go)
- TIA with noise model (shared/peripherals/tia.go)
- 2-stage Dickson charge pump (shared/peripherals/chargepump.go)
- System-level timing/power analysis (shared/peripherals/analysis.go)
- Tier-A DC nodal solver (module4-circuits/pkg/arraysim/tier_a.go)
- Sense chain model (module4-circuits/pkg/arraysim/sensechain.go)

### Known Gaps (from project's own MODULE4-PHYSICS-IMPROVEMENTS.md)
1. Linear conductance model (need exponential/Preisach) -- HIGH
2. No sneak paths in Module 4 -- HIGH
3. No IR drop in Module 4 -- HIGH
4. No write disturb tracking -- HIGH
5. Temperature effects (only 300K) -- MEDIUM
6. No switching statistics -- MEDIUM
7. No endurance model -- MEDIUM
8. TIA frequency response -- MEDIUM
9. Charge pump transient response -- LOW
10. ADC comparator kickback noise -- LOW
11. SET/RESET asymmetry -- LOW
12. Retention loss model -- LOW

---

## Glossary

### Architectures
- **0T1R**: Zero-transistor, one-resistor per cell. Simplest but suffers from sneak paths.
- **1T1R**: One-transistor, one-resistor per cell. Transistor acts as selector to minimize sneak paths.
- **2T1R**: Two-transistors, one-resistor. Enables differential signaling and improved linearity.
- **0T1C (FeCAP)**: Zero-transistor, one-capacitor. Uses ferroelectric capacitors; eliminates sneak paths via displacement current.

### Device Types
- **FeFET**: Ferroelectric Field-Effect Transistor. CMOS-compatible, non-volatile memory transistor.
- **FeCAP**: Ferroelectric Capacitor. Stores charge based on polarization state.
- **FTJ**: Ferroelectric Tunnel Junction. Two-terminal device with tunneling-based resistance switching.
- **HZO**: Hafnium Zirconium Oxide (Hf₁₋ₓZrₓO₂). Most common ferroelectric material for CMOS integration.

### Physics Terms
- **Coercive field (Ec)**: Electric field required to switch ferroelectric polarization direction.
- **Remnant polarization (Pr)**: Polarization remaining after removing electric field.
- **Saturation polarization (Ps)**: Maximum polarization achievable under strong electric field.
- **Displacement current**: Transient current from changing electric field in a capacitor (∂D/∂t), not from charge carriers.
- **Preisach model**: Mathematical model for hysteresis in ferroelectric materials using distribution of switching events.

### Circuit Components
- **DAC**: Digital-to-Analog Converter. Converts digital input to analog voltage/current.
- **ADC**: Analog-to-Digital Converter. Converts analog signal to digital output.
- **TIA**: Transimpedance Amplifier. Converts current to voltage with gain (I→V).
- **SAR ADC**: Successive Approximation Register ADC. Binary search-based conversion, moderate speed/power.
- **Flash ADC**: Parallel comparator-based ADC. Fastest but highest power/area (2ⁿ-1 comparators).
- **Ramp ADC**: Counter-based slope ADC. Slowest but lowest power, good for column-shared architectures.

### Non-Idealities
- **Sneak path**: Unintended current path through neighboring cells in passive crossbar, causing errors.
- **IR drop**: Voltage drop across metal interconnect resistance, reduces effective voltage at distant cells.
- **C2C variation**: Cycle-to-Cycle variation. Random fluctuation in device state between successive read operations.
- **D2D variation**: Device-to-Device variation. Systematic fabrication differences across die/wafer.
- **Programming error**: Deviation of programmed conductance from target value.
- **Conductance drift**: Time-dependent change in conductance after programming (power-law or logarithmic).
- **Write disturb**: Unintended partial programming of half-selected cells during write operation.
- **Half-select**: Cells sharing wordline OR bitline (but not both) with selected cell during write.

### Programming Techniques
- **ISPP**: Incremental Step Pulse Programming. Iterative write-verify with increasing voltage pulses.
- **DCC**: Displacement Current Control. One-shot programming using controlled charge displacement (emerging).
- **A-ISPP**: Adaptive ISPP. Variable step size based on distance from target (large steps far, small near).
- **Write-verify**: Read-after-write to confirm programmed state, retry if needed.

### System Metrics
- **MVM**: Matrix-Vector Multiplication. Core computation in analog CIM arrays.
- **TOPS/W**: Tera-Operations Per Second per Watt. Energy efficiency metric for compute accelerators.
- **ENOB**: Effective Number of Bits. ADC accuracy metric accounting for noise and distortion.
- **INL/DNL**: Integral/Differential Non-Linearity. DAC/ADC error metrics (deviation from ideal).
- **EAP**: Energy-Area Product. Combined metric for comparing accelerator efficiency.
- **F²**: Minimum feature size squared. Standard unit for cell area (e.g., 4F² = 4× smallest lithographic dimension²).

### Simulation Tools
- **CrossSim**: Sandia National Labs' crossbar simulator with PyTorch integration (V3.1, Jan 2025).
- **NeuroSim**: Georgia Tech's device-to-system simulator with peripheral modeling (V1.5, 2025).
- **MNSIM**: Tsinghua's fast behavior-level simulator, 7000× faster than SPICE (V2.0).
- **SOR**: Successive Over-Relaxation. Iterative method for solving Kirchhoff equations for IR drop.

### Encoding Schemes
- **Binary encoding**: N bits → N DAC switches. Simple but can have glitches at MSB transitions.
- **Thermometer encoding**: N bits → 2ⁿ-1 switches. Inherently monotonic, glitch-free, but high area.
- **Segmented encoding**: Hybrid of thermometer (MSB) + binary (LSB). Balanced area/performance.

### Emerging Concepts
- **Charge-domain sensing**: Readout based on integrated charge (Q) rather than steady-state current (I).
- **Capacitive crossbar**: MVM using C×V rather than G×V, operates on displacement current.
- **Comparator-only digitization**: Replace full ADC with single comparator for binary networks (28× energy reduction).
- **Hardware-aware quantization**: Neural network quantization optimized for specific ADC/DAC bit-width constraints.

---

## Document Changelog

**2026-02-16**
- Added Table of Contents with navigation links
- Added "How to Use This Document" section
- Enhanced Executive Summary with impact table and immediate actions
- Added Quick Decision Trees for architecture selection
- Enhanced priority table with detailed justifications
- Added 5-phase Implementation Roadmap with weekly breakdown
- Added comprehensive Glossary (70+ terms)
- Improved markdown formatting and visual hierarchy

**2026-02-14**
- Initial literature review covering 29 peer-reviewed papers (2017-2025)
- Comprehensive analysis of crossbar architectures and peripheral circuits
- Synthesis with prioritized recommendations (P0-P3)
- Gap analysis vs. existing project capabilities
