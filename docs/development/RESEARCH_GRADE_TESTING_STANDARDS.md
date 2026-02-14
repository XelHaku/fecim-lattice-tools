# Research-Grade Ferroelectric Memory Testing Standards

**Generated:** 2026-02-13
**Purpose:** Define testing standards for research-grade ferroelectric compute-in-memory (FeCIM) simulators based on current literature (2024-2026)

---

## Executive Summary

This document synthesizes testing methodologies from recent ferroelectric memory research (HZO-based FeFET, FTJ, and FeCIM devices) to establish what constitutes "research-grade" validation for a physics-based simulator. A simulator claiming research-grade accuracy should validate against these standard characterization protocols.

**Key Principle:** Research-grade testing separates true ferroelectric behavior from artifacts, validates across statistical populations, and reports standardized metrics that enable peer comparison.

---

## 1. Standard Ferroelectric Characterization Tests

### 1.1 P-E Hysteresis Loop Measurements

#### 1.1.1 DHM (Dynamic Hysteresis Mode)
- **Method:** Single bipolar triangular pulse at specified frequency
- **Advantages:** Fast, easy for repeated cycling
- **Limitations:** Does NOT account for leakage current; difficult to estimate true switchable polarization
- **Typical Frequency:** 1 kHz standard, with frequency dependence characterization
- **Key Parameters Extracted:** Ps (saturation), Pr (remnant), Ec (coercive field), hysteresis loop area

**Reference:** [NPL Report CMMT(A) 152](https://physlab.lums.edu.pk/images/e/eb/Reframay4.pdf), [Advanced Materials Lab - Ferroelectric Hysteresis](https://advancematerialslab.com/ferroelectric-hysteresis-loop-artifacts/)

#### 1.1.2 PUND (Positive-Up Negative-Down)
- **Method:** Sequence of five consecutive voltage pulses (1 kHz standard) with different polarities
- **Advantages:**
  - Eliminates leakage current contributions
  - Removes dielectric displacement effects
  - Accounts for parasitic capacitance
  - Extracts true remnant polarization (2Pr)
- **Gold Standard:** PUND is preferred for publication-quality ferroelectric verification

**Reference:** [Advanced Materials Lab - PUND Measurements](https://advancematerialslab.com/how-to-perform-pund-measurement/)

#### 1.1.3 Validation Requirements
A research-grade simulator SHOULD:
- [ ] Support both DHM and PUND measurement simulation
- [ ] Report Ps, Pr, Ec from hysteresis loops
- [ ] Demonstrate frequency dependence of hysteresis (coercive field increases with frequency due to polarization reversal inertia)
- [ ] Show narrowing of hysteresis loop with temperature (energy barrier reduction)
- [ ] Separate switching current from leakage/capacitive components (PUND validation)

---

### 1.2 Switching Dynamics

#### 1.2.1 Nucleation-Limited Switching (NLS) Model
- **Physics:** Polycrystalline ferroelectric films switch via nucleation-limited dynamics
- **Key Metrics:**
  - Incubation time of nucleation
  - Intra-grain and inter-grain switching statistics
  - No restrictions on waveform or switching time scales
- **State-of-the-Art Performance (2024):**
  - Sub-nanosecond switching: **925 ps** characteristic time (single crossbar)
  - Record fast switching: **360 ps** (0.1 μm² crossbar array)

**Reference:** [IEEE Xplore - Nucleation-Limited Switching](https://ieeexplore.ieee.org/document/9640228/), [NIST - Ultrafast Measurements](https://engineering.purdue.edu/~yep/Papers/APL_Fast%20Pulse_NIST_2019.pdf)

#### 1.2.2 Measurement Techniques
- Ultrafast pulse generator
- Variable gain transimpedance current amplifier
- Ultrafast high-definition oscilloscope
- Impedance-matched pickoff tee circuits (minimize signal reflection)
- Real-time monitoring of transient polarization switching current

**Validation Requirements:**
- [ ] Simulator should model nucleation-limited switching behavior
- [ ] Report switching time as a function of pulse amplitude
- [ ] Capture incubation time before nucleation
- [ ] Support nanosecond to sub-nanosecond switching regimes

---

### 1.3 Retention Testing

#### 1.3.1 Standard Test Conditions
- **Industry Standard:** 10 years retention @ 85°C
- **Challenge:** Cannot measure 10 years directly → requires extrapolation methods
- **Extrapolation Techniques:**
  - High-temperature accelerated aging (up to 160°C)
  - Linear extrapolation from retention vs. time plots
  - Arrhenius modeling of temperature dependence

#### 1.3.2 State-of-the-Art Performance (2024-2026)
- HZO retention: >93% polarization after projected 10 years
- 2D ferroelectric-gated units: >10 years retention (>10^5 seconds measured, extrapolated)
- FTJ: >10^4 seconds retention, extrapolated >10 years @ elevated temperature

**Reference:** [Materials Horizons - Lorentz-tail engineering](https://pubs.rsc.org/en/content/articlelanding/2026/mh/d5mh01981h), [Phys. Rev. Applied - Retention Model](https://link.aps.org/doi/10.1103/PhysRevApplied.18.064084)

#### 1.3.3 Validation Requirements
- [ ] Simulate retention decay over time (exponential, Lorentz-tail, or Arrhenius models)
- [ ] Support temperature-dependent retention characterization
- [ ] Provide extrapolation tools for 10-year prediction from short-term data
- [ ] Report retention at standard condition: 10 years @ 85°C

---

### 1.4 Endurance/Fatigue Cycling

#### 1.4.1 Metrics
- **Endurance:** Number of write/erase cycles before failure
- **Fatigue:** Cycle-by-cycle decrease in remnant polarization after certain cycles
- **Industry Targets:**
  - FeRAM production: 10^12 to 10^15 cycles
  - HZO research (2024): 10^9 cycles demonstrated
  - FeFET challenge: Limited to 10^4–10^5 cycles (gate-stack degradation)

#### 1.4.2 Failure Mechanisms
- Wake-up effect: Rise in switching polarization during initial cycling
- Fatigue: Polarization degradation after extended cycling
- Large coercive fields cause endurance problems
- Temperature dependence: Endurance varies across -40°C to 40°C range

**Reference:** [MDPI - Characterization of HZO Films](https://www.mdpi.com/2079-4991/14/22/1801), [IEEE Xplore - Retention and Endurance](https://ieeexplore.ieee.org/document/8739726/)

#### 1.4.3 Validation Requirements
- [ ] Simulate fatigue (Pr degradation vs. cycle count)
- [ ] Model wake-up effect (initial Pr increase during first ~10^3–10^5 cycles)
- [ ] Report endurance limit (cycles to X% polarization loss, e.g., 50%)
- [ ] Characterize endurance temperature dependence

---

### 1.5 Wake-Up Characterization

#### 1.5.1 Physics
- **Mechanism:** Oxygen vacancy (VO) and O2− ion redistribution under electric field
- **Effect:** More uniform internal field → new domain involvement via phase transformation
- **Observation:** Polarization increases during initial cycling stage
- **Recent Progress (2024):** Wake-up-free HZO via co-plasma ALD fabrication

**Reference:** [ACS Applied Materials - Phase-Exchange-Driven Wake-Up](https://pubs.acs.org/doi/abs/10.1021/acsami.0c03570), [Nature Communications - Polar and Antipolar Phase Transition](https://www.nature.com/articles/s41467-022-28236-5)

#### 1.5.2 Validation Requirements
- [ ] Model polarization evolution during first 10^3 to 10^6 cycles
- [ ] Show wake-up mechanism (defect redistribution, field uniformization)
- [ ] Report "wake-up-free" materials (if applicable)
- [ ] Capture relationship between wake-up and switching field strength

---

### 1.6 Imprint Testing

#### 1.6.1 Definition
- **Imprint:** Gradual shift of hysteresis curve with repeated unipolar stress
- **Manifestation:** Preferred polarization state develops; coercive fields become asymmetric
- **Mechanism:** Related to charge injection, defect migration

#### 1.6.2 Testing Protocol
- Apply repeated unipolar pulses
- Periodically measure P-E loop
- Quantify horizontal shift of loop center
- Report imprint voltage vs. stress time/cycles

**Reference:** [MDPI - Wake-Up and Imprint Effects](https://www.mdpi.com/2079-9042/13/6/1021)

#### 1.6.3 Validation Requirements
- [ ] Simulate imprint accumulation under unipolar stress
- [ ] Report imprint voltage shift vs. stress cycles
- [ ] Model charge injection effects if relevant

---

### 1.7 Temperature Dependence of Ferroelectric Parameters

#### 1.7.1 Standard Characterization
- **Temperature Range:** Typically -40°C to 125°C (industrial), up to 160°C (research)
- **Key Observations:**
  - Ps, Pr, Vc (coercive voltage) decrease with increasing temperature
  - Ec decreases linearly with temperature (all frequencies)
  - Hysteresis loop narrows with temperature (energy barrier reduction)

**Reference:** [ScienceDirect - Temperature Dependence PZT](https://www.sciencedirect.com/science/article/pii/S0272884219315068), [ACS Applied Materials - Temperature-Dependent Ferroelectric Properties](https://pubs.acs.org/doi/10.1021/acsami.4c03002)

#### 1.7.2 Validation Requirements
- [ ] Ps(T), Pr(T), Ec(T) temperature dependence
- [ ] Linear Ec decrease with temperature
- [ ] Hysteresis loop area reduction with temperature
- [ ] Support wide temperature range (-40°C to 125°C minimum)

---

### 1.8 Device-to-Device Variability

#### 1.8.1 Statistical Testing Requirements
- **Device-to-Device (D2D) Variation:**
  - Increases drastically with device scaling
  - Dominates over cycle-to-cycle variation for scaled devices
  - Requires Monte Carlo or statistical ensemble modeling

- **Cycle-to-Cycle (C2C) Variation:**
  - Almost insensitive to device size
  - Function of conductance state (not fixed dispersion)
  - Measurable via PUND over multiple cycles

**Reference:** [IEEE Xplore - Fundamental Understanding of D2D Variation](https://ieeexplore.ieee.org/document/8776497/), [ScienceDirect - Implementation of D2D and C2C Variability](https://www.sciencedirect.com/science/article/abs/pii/S0038110122000934)

#### 1.8.2 State-of-the-Art (2024)
- 2D ferroelectric gates: ~0.3% C2C variation, ~0.5% D2D variation (ultralow)
- Machine learning-assisted variability analysis from experimental metrology

#### 1.8.3 Validation Requirements
- [ ] Report D2D variation (σ/μ) as function of device size
- [ ] Report C2C variation over cycling
- [ ] Provide statistical distributions (Gaussian, lognormal, etc.)
- [ ] Support Monte Carlo simulation for array-level variation
- [ ] Variation should be state-dependent (different for each conductance level)

---

## 2. CIM-Specific Test Protocols

### 2.1 Multi-Level Cell (MLC) Programming Accuracy

#### 2.1.1 Key Metrics
- **Number of Levels:** State-of-the-art FeFET: 7 distinct VT states (~3 bits/cell), up to 16 levels demonstrated
- **Programming Window:** Voltage range for reliable state separation (10 V demonstrated in 2024)
- **State Separation:** Minimum detectable difference between adjacent levels
- **Program/Verify Scheme:** Iterative pulse-verify to precisely control threshold voltage

**Reference:** [ResearchGate - Reliable MLC Programming in FeFET](https://www.researchgate.net/publication/390350071_Reliable_multi-level_cell_programming_in_FeFET_arrays_for_in-memory_computing), [Science Advances - Unlocking Large Memory Windows](https://www.science.org/doi/10.1126/sciadv.adn1345)

#### 2.1.2 Validation Requirements
- [ ] Support configurable number of analog levels (minimum 7-16 for research-grade)
- [ ] Implement program/verify write scheme
- [ ] Report programming window (voltage range for level separation)
- [ ] Demonstrate level stability over retention/cycling

---

### 2.2 Conductance Linearity

#### 2.2.1 Resistance Allocation Schemes
- **ISO-ΔR:** Resistance linearly spaced (R1, R2, R3, ... Rn)
- **ISO-ΔI:** Programming current (inverse resistance) linearly spaced → Adopted for MLC schemes based on reset current control

#### 2.2.2 Validation Requirements
- [ ] Report conductance linearity metric (R² or max deviation from linear fit)
- [ ] Support both potentiation (LTP) and depression (LTD) linearity
- [ ] Characterize conductance drift over time (affects MVM accuracy)

**Reference:** [ResearchGate - MLC Programming and Conductance](https://www.researchgate.net/publication/390350071_Reliable_multi-level_cell_programming_in_FeFET_arrays_for_in-memory_computing)

---

### 2.3 Read Disturb Characterization

#### 2.3.1 Test Protocol
- Write target state to cell
- Perform repeated read operations (10^3 to 10^9 reads)
- Monitor state degradation
- Report read disturb as shift in conductance/polarization

#### 2.3.2 State-of-the-Art Solutions
- Separated write/read paths (asymmetric double-gate FeFET) → disturb-free read
- Optimal pass voltage in NAND arrays → negligible read disturb

**Reference:** [IEEE Xplore - Write Disturb in FeFETs](https://ieeexplore.ieee.org/document/8474379/), [ACS Applied Materials - Exploring Disturb Characteristics](https://pubs.acs.org/doi/10.1021/acsami.4c03785)

#### 2.3.3 Validation Requirements
- [ ] Simulate read disturb (state shift after repeated reads)
- [ ] Report read disturb vs. read voltage
- [ ] Report read disturb vs. number of read operations

---

### 2.4 Write Disturb / Half-Select Disturb

#### 2.4.1 Array-Level Testing
- **VDD/2 Scheme:** Unselected cells experience 0 or VDD/2 only
- **VDD/3 Scheme:** Unselected cells experience both positive and negative voltages
- **Worst-Case Characterization:** Half-selected cells during write operations

#### 2.4.2 Inhibition Schemes
- VW/2 and VW/3 inhibition bias schemes tested experimentally
- C-AND (crossed-AND) architecture with asymmetric switching voltage compensation

**Reference:** [arXiv - C-AND Mixed Writing Scheme](https://arxiv.org/pdf/2205.12061), [IEEE Xplore - Write Disturb in 1T-FeFET](https://ieeexplore.ieee.org/document/8474379/)

#### 2.4.3 Validation Requirements
- [ ] Simulate half-select disturb in crossbar arrays
- [ ] Test VDD/2 and VDD/3 write schemes
- [ ] Report worst-case write disturb (checkerboard pattern)
- [ ] Validate inhibit voltage effectiveness

---

### 2.5 Analog Compute Accuracy (MVM Error Metrics)

#### 2.5.1 Key Metrics
- **Bit Error Rate (BER):** State-of-the-art FeFET: 4% BER with only 1% accuracy degradation from floating-point (CIFAR-10)
- **Relative Error:** Error vs. input vector magnitude for different bit-widths
- **Inference Accuracy:** Classification accuracy on standard benchmarks (MNIST, CIFAR-10)

**Reference:** [IEEE Xplore - Multi-Level Operation for CIM](https://ieeexplore.ieee.org/document/10145940/), [Nature - First Demonstration MLC FeFET](https://www.nature.com/articles/s41467-023-42110-y)

#### 2.5.2 Error Sources in Analog MVM
- Conductance polarity dependence
- IR drop in bitlines/wordlines
- Sneak path currents
- ADC quantization
- Conductance drift during inference

#### 2.5.3 Validation Requirements
- [ ] Report BER for MLC programming
- [ ] Measure inference accuracy on standard benchmarks
- [ ] Compare to floating-point baseline
- [ ] Quantify MVM error (L2 norm, relative error) vs. ideal computation
- [ ] Test with calibration procedures (analog/digital)

---

### 2.6 State Retention After Compute Operations

#### 2.6.1 CIM-Specific Retention
- **Challenge:** Memristor state drift induced by recurring sub-threshold voltage stress during inference
- **Crossbar-Level Testing:** First Mb-level long-term retention characterization (2024)
- **Relaxed Requirements for Training:** Gradient accumulation materials need only very short retention

**Reference:** [IEEE Xplore - Crossbar-Level Retention](https://ieeexplore.ieee.org/document/9462921/), [Nature Communications - Fast and Robust Training](https://www.nature.com/articles/s41467-024-51221-z)

#### 2.6.2 Validation Requirements
- [ ] Simulate state drift during repeated MVM operations
- [ ] Report retention after 10^6 to 10^9 compute cycles
- [ ] Distinguish inference retention requirements (long) vs. training (short)
- [ ] Test sub-threshold stress effects

---

### 2.7 Bit Error Rate at Different Conductance Levels

#### 2.7.1 State-Dependent BER
- FTJ devices show continuous analog tuning beyond binary states
- Conductance variation is a function of different states (not fixed)
- Raw BER: <5.9×10^-2 demonstrated for ferroelectric-silicon memory

**Reference:** [Wiley - Ferroelectric Devices](https://advanced.onlinelibrary.wiley.com/doi/10.1002/aisy.202500817)

#### 2.7.2 Validation Requirements
- [ ] Report BER as function of programmed conductance level
- [ ] Show higher/lower BER for extreme vs. mid-range levels
- [ ] Test BER degradation with retention/cycling

---

## 3. Metrics Reviewers Expect in Ferroelectric CIM Papers

### 3.1 Core Ferroelectric Metrics

| Metric | Symbol | Typical Units | Expected Reporting |
|--------|--------|---------------|-------------------|
| Saturation Polarization | Ps | μC/cm² | From P-E loop (DHM/PUND) |
| Remnant Polarization | Pr | μC/cm² | From P-E loop, PUND preferred |
| Coercive Field | Ec | MV/cm or kV/cm | Both positive and negative branches |
| Programming Window | ΔV | Volts | Voltage range for stable states |
| Memory Window | ΔVT | Volts | Threshold voltage separation (FeFET) |
| On/Off Ratio | Ion/Ioff | Dimensionless | Current ratio between states |

**Reference:** Multiple sources including [MDPI](https://www.mdpi.com/2079-4991/14/22/1801), [Science Advances](https://www.science.org/doi/10.1126/sciadv.adn1345)

---

### 3.2 Reliability Metrics

| Metric | Standard Target | Measurement Method | Expected Reporting |
|--------|----------------|-------------------|-------------------|
| Endurance | 10^9 cycles (HZO research)<br>10^12 cycles (FeRAM production) | Cycling P-E loops, monitor Pr degradation | Cycles to 50% polarization loss |
| Retention | 10 years @ 85°C | High-temp acceleration + extrapolation | Extrapolated polarization @ 10 years |
| Switching Energy | fJ to pJ range | Integrate IV during switching | Energy per write operation |
| Switching Time | <1 ns (state-of-art) | Ultrafast current measurement | Characteristic switching time (τ) |

**Reference:** [IEEE Xplore - Retention and Endurance](https://ieeexplore.ieee.org/document/8739726/), [TechXplore - FeRAM Performance](https://techxplore.com/news/2025-04-ferroelectric-device-memory.html)

---

### 3.3 CIM-Specific Metrics

| Metric | Typical Value | Measurement | Purpose |
|--------|--------------|-------------|---------|
| MLC Levels | 7–16 states | Program/verify, read distribution | Storage density |
| Conductance Linearity | R² > 0.95 | LTP/LTD curve fitting | Synaptic learning accuracy |
| BER | <5% (FeFET state-of-art) | Monte Carlo over variations | Inference reliability |
| MVM Accuracy | 1–10% degradation from FP | MNIST/CIFAR-10 inference | Application performance |
| Read/Write Latency | ns to μs | Oscilloscope measurement | Throughput |
| Read/Write Energy | fJ to pJ per operation | Power × time measurement | Energy efficiency |

**Reference:** [Nature Communications - MLC FeFET Crossbar](https://www.nature.com/articles/s41467-023-42110-y), [Nature - Achieving High Precision in AIMC](https://www.nature.com/articles/s44335-025-00044-2)

---

## 4. Standard Test Patterns for Crossbar Arrays

### 4.1 Checkerboard Pattern
- **Purpose:** Alternate high/low states to test cell-to-cell coupling, worst-case capacitive coupling
- **Pattern:** `[1,0,1,0; 0,1,0,1; 1,0,1,0; ...]`
- **Validation:** Read all cells, verify no flipping from intended state

---

### 4.2 All-Ones / All-Zeros
- **All-Ones:** Worst-case sneak path current (maximum leakage)
- **All-Zeros:** Best-case sneak path current (minimum leakage)
- **Purpose:** Characterize sneak path extremes, IR drop under maximum current

**Reference:** [arXiv - Sneak Path Current Modeling](https://arxiv.org/html/2511.21796)

---

### 4.3 Walking Ones / Walking Zeros
- **Walking Ones:** Write single '1' bit, shift left through array (e.g., 00000001 → 00000010 → 00000100...)
- **Walking Zeros:** Inverse (single '0' bit walking through '1's)
- **Purpose:** Detect address decoder faults, stuck-at faults, cell isolation failures

**Reference:** [EECS Michigan - Memory Testing](http://www.eecs.umich.edu/courses/eecs579.f01/579.L15.pdf), [Verification Academy - Walking Ones Test](https://verificationacademy.com/verification-methodology-reference/uvm/docs_1.1b/html/files/reg/sequences/uvm_mem_walk_seq-svh.html)

---

### 4.4 March Test Algorithms

#### 4.4.1 March C (11n operations)
- Writes/reads words of 0s
- Writes/reads words of 1s
- Both ascending and descending address order
- **Fault Coverage:** Comprehensive stuck-at, transition, coupling faults

#### 4.4.2 March C- (10n operations)
- Modified March C with redundant Read 0 operation removed
- Same fault coverage, reduced test time

**Reference:** [ResearchGate - Simple Diagnostic Method](https://www.researchgate.net/publication/261855956_A_simple_diagnostic_method_for_memory_testing)

#### 4.4.3 Limitations for Memristor Arrays
- March algorithms require serial programming/reading
- Impractical for large neuromorphic arrays (10^6+ cells)
- Alternative: Statistical sampling or application-specific testing

**Reference:** [RSC - Sneak Path Solutions](https://pubs.rsc.org/en/content/articlelanding/2020/na/d0na00100g)

---

### 4.5 Worst-Case Sneak Path Pattern
- **Objective:** Maximize sneak path current to test sensing margins
- **Pattern:** Depends on array topology (1T1C, passive crossbar, etc.)
- **Validation:** Analytical models show <10.9% error vs. SPICE, 10–4784× faster

**Reference:** [arXiv - Sneak Path Modeling](https://arxiv.org/html/2511.21796v2)

---

## 5. What a Simulator SHOULD Validate (Checklist)

### 5.1 Core Ferroelectric Physics
- [ ] **P-E Hysteresis:** Both DHM and PUND methods, report Ps/Pr/Ec
- [ ] **Switching Dynamics:** Nucleation-limited switching, sub-ns capable
- [ ] **Retention:** Extrapolation to 10 years @ 85°C
- [ ] **Endurance:** Cycling to 10^9+ cycles, model fatigue
- [ ] **Wake-Up:** Initial polarization rise (first 10^3–10^6 cycles)
- [ ] **Imprint:** Horizontal loop shift under unipolar stress
- [ ] **Temperature Dependence:** Ps(T), Pr(T), Ec(T) from -40°C to 125°C

### 5.2 Statistical Characterization
- [ ] **Device-to-Device Variation:** Scaling-dependent, Monte Carlo capable
- [ ] **Cycle-to-Cycle Variation:** State-dependent, reported over many cycles
- [ ] **Array-Level Statistics:** Variation propagation through MVM operations

### 5.3 CIM-Specific Validation
- [ ] **MLC Programming:** 7+ levels, program/verify scheme
- [ ] **Conductance Linearity:** LTP/LTD, R² > 0.95
- [ ] **Read Disturb:** State shift after 10^6+ reads
- [ ] **Write Disturb:** Half-select in VDD/2, VDD/3 schemes
- [ ] **BER vs. Level:** State-dependent error rates
- [ ] **MVM Accuracy:** MNIST/CIFAR-10 inference within 1–10% of floating-point
- [ ] **State Retention After Compute:** Drift under sub-threshold stress

### 5.4 Array-Level Testing
- [ ] **Checkerboard Pattern:** Cell coupling validation
- [ ] **All-Ones/Zeros:** Sneak path extremes, IR drop
- [ ] **Walking Ones/Zeros:** Decoder and cell isolation faults
- [ ] **March C/C-:** Comprehensive fault coverage (if applicable)
- [ ] **Worst-Case Sneak Path:** Sensing margin validation

---

## 6. Comparison to Industry Standards (JEDEC/IEEE)

### 6.1 Current State
- **No dedicated JEDEC standard for ferroelectric memory testing** (as of 2024 search)
- **JESD22 Series:** General semiconductor stress testing (not FeRAM-specific)
- **IEEE Standards:** No unified ferroelectric memory testing standard identified

### 6.2 Implication for Simulators
Research-grade simulators should follow **community best practices from literature** rather than a formal standard. Key reference papers (2024-2026) include:
- Multi-level FeFET CIM operations (IEEE, Nature Communications)
- HZO characterization methodologies (MDPI, ACS)
- Retention/endurance models (Phys. Rev. Applied, IEEE)

**Reference:** [JEDEC Standards Search](https://www.jedec.org/standards-documents), [Wikipedia - JEDEC Memory Standards](https://en.wikipedia.org/wiki/JEDEC_memory_standards)

---

## 7. Recommended Testing Workflow for Simulator Validation

### Stage 1: Core Physics Validation
1. **P-E Loop:** Compare DHM and PUND against experimental data
2. **Switching Dynamics:** Validate nucleation-limited model, switching time vs. voltage
3. **Temperature Sweep:** Verify Ec(T), Pr(T), Ps(T) trends

### Stage 2: Reliability Characterization
1. **Retention:** Accelerated aging simulation, extrapolate to 10 years @ 85°C
2. **Endurance:** Cycle to 10^6+ cycles, track Pr degradation
3. **Wake-Up/Imprint:** Demonstrate mechanisms in first 10^5 cycles

### Stage 3: CIM Operation Validation
1. **MLC Programming:** Program 7–16 levels, verify stability
2. **MVM Accuracy:** Run MNIST inference, compare to floating-point
3. **BER Characterization:** Report BER vs. conductance level
4. **Disturb Testing:** Read/write disturb under array conditions

### Stage 4: Array-Level Testing
1. **Standard Patterns:** Checkerboard, all-ones/zeros, walking patterns
2. **Sneak Path:** Measure worst-case sneak current, validate against analytical models
3. **Statistical Testing:** Monte Carlo with D2D/C2C variation on 1000+ virtual devices

---

## 8. Sources and References

### Standard Ferroelectric Characterization
- [NPL Report CMMT(A) 152 - Ferroelectric Hysteresis Measurement](https://physlab.lums.edu.pk/images/e/eb/Reframay4.pdf)
- [Advanced Materials Lab - PUND Measurements](https://advancematerialslab.com/how-to-perform-pund-measurement/)
- [Advanced Materials Lab - Ferroelectric Hysteresis Loop Artifacts](https://advancematerialslab.com/ferroelectric-hysteresis-loop-artifacts/)
- [MDPI - Characterization of HZO Films (2024)](https://www.mdpi.com/2079-4991/14/22/1801)
- [Materials Horizons - Lorentz-tail Engineering 10-Year Retention](https://pubs.rsc.org/en/content/articlelanding/2026/mh/d5mh01981h)

### Switching Dynamics
- [IEEE Xplore - Nucleation-Limited Switching Dynamics Model](https://ieeexplore.ieee.org/document/9640228/)
- [Purdue/NIST - Ultrafast Measurements of Polarization Switching](https://engineering.purdue.edu/~yep/Papers/APL_Fast%20Pulse_NIST_2019.pdf)
- [NIST - Dynamics Studies of Polarization Switching](https://www.nist.gov/publications/dynamics-studies-polarization-switching-ferroelectric-hafnium-zirconium-oxide)

### Reliability (Retention, Endurance, Wake-Up)
- [Phys. Rev. Applied - Retention Model and Express Test](https://link.aps.org/doi/10.1103/PhysRevApplied.18.064084)
- [IEEE Xplore - Retention and Endurance of FeFET Memory Cells](https://ieeexplore.ieee.org/document/8739726/)
- [ACS Applied Materials - Phase-Exchange-Driven Wake-Up and Fatigue](https://pubs.acs.org/doi/abs/10.1021/acsami.0c03570)
- [Nature Communications - Reversible Transition Polar/Antipolar Phases](https://www.nature.com/articles/s41467-022-28236-5)
- [MDPI - Wake-Up and Imprint Effects with Cycling](https://www.mdpi.com/2079-9042/13/6/1021)

### CIM-Specific Testing
- [ResearchGate - Reliable MLC Programming in FeFET Arrays](https://www.researchgate.net/publication/390350071_Reliable_multi-level_cell_programming_in_FeFET_arrays_for_in-memory_computing)
- [IEEE Xplore - Multi-Level Operation for CIM](https://ieeexplore.ieee.org/document/10145940/)
- [Nature Communications - First Demonstration MLC FeFET Crossbar](https://www.nature.com/articles/s41467-023-42110-y)
- [Science Advances - Unlocking Large Memory Windows 16-Level](https://www.science.org/doi/10.1126/sciadv.adn1345)
- [Science Advances - 2D Ferroelectric-Gated Hybrid CIM](https://www.science.org/doi/10.1126/sciadv.adp0174)

### Read/Write Disturb
- [IEEE Xplore - Write Disturb in FeFETs](https://ieeexplore.ieee.org/document/8474379/)
- [arXiv - C-AND Mixed Writing Scheme](https://arxiv.org/pdf/2205.12061)
- [ACS Applied Materials - Exploring Disturb Characteristics 2D/3D](https://pubs.acs.org/doi/10.1021/acsami.4c03785)

### Variability and Statistical Testing
- [IEEE Xplore - Fundamental Understanding D2D Variation](https://ieeexplore.ieee.org/document/8776497/)
- [ScienceDirect - Implementation D2D and C2C Variability](https://www.sciencedirect.com/science/article/abs/pii/S0038110122000934)
- [IEEE Xplore - Machine Learning Assisted Statistical Variation](https://ieeexplore.ieee.org/document/9830392/)

### Analog Compute Accuracy and MVM
- [Nature - Demonstration 4-Quadrant Analog MVM](https://www.nature.com/articles/s44335-024-00010-4)
- [Nature - Achieving High Precision in AIMC](https://www.nature.com/articles/s44335-025-00044-2)
- [ResearchGate - Homogeneous Processing Fabric MVM](https://www.researchgate.net/publication/363859413_A_Homogeneous_Processing_Fabric_for_Matrix-Vector_Multiplication_and_Associative_Search_Using_Ferroelectric_Time-Domain_Compute-in-Memory)

### Crossbar Array Testing and Sneak Paths
- [arXiv - Sneak Path Current Modeling](https://arxiv.org/html/2511.21796)
- [RSC - Research Progress on Sneak Path Solutions](https://pubs.rsc.org/en/content/articlelanding/2020/na/d0na00100g)
- [IEEE Xplore - Sneak-Path Constraints in Memristor Crossbars](https://ieeexplore.ieee.org/document/6620207/)

### State Retention After Compute
- [IEEE Xplore - Crossbar-Level Retention Characterization](https://ieeexplore.ieee.org/document/9462921/)
- [Nature Communications - Fast and Robust Analog Training](https://www.nature.com/articles/s41467-024-51221-z)

### Memory Test Patterns
- [EECS Michigan - Memory Testing Lecture](http://www.eecs.umich.edu/courses/eecs579.f01/579.L15.pdf)
- [Verification Academy - UVM Walking Ones Test Sequences](https://verificationacademy.com/verification-methodology-reference/uvm/docs_1.1b/html/files/reg/sequences/uvm_mem_walk_seq-svh.html)
- [ResearchGate - Simple Diagnostic Method for Memory Testing](https://www.researchgate.net/publication/261855956_A_simple_diagnostic_method_for_memory_testing)

### Performance Benchmarks
- [TechXplore - FeRAM Performs Calculations](https://techxplore.com/news/2025-04-ferroelectric-device-memory.html)
- [PMC - Ferroelectric Devices for Content-Addressable Memory](https://pmc.ncbi.nlm.nih.gov/articles/PMC9785747/)
- [Infineon - F-RAM Ferroelectric RAM](https://www.infineon.com/products/memories/f-ram-ferroelectric-ram)

### Temperature Dependence
- [ScienceDirect - Temperature Dependence PZT Thin Films](https://www.sciencedirect.com/science/article/pii/S0272884219315068)
- [ACS Applied Materials - Temperature-Dependent Ferroelectric Properties](https://pubs.acs.org/doi/10.1021/acsami.4c03002)

### Standards Organizations
- [JEDEC Standards Search](https://www.jedec.org/standards-documents)
- [Wikipedia - JEDEC Memory Standards](https://en.wikipedia.org/wiki/JEDEC_memory_standards)

---

## 9. Conclusion

A **research-grade ferroelectric CIM simulator** should validate against the testing protocols documented in recent experimental literature (2024-2026). While no formal JEDEC/IEEE standard exists specifically for ferroelectric memory, the scientific community has converged on best practices covering:

1. **Ferroelectric characterization:** PUND-validated P-E loops, switching dynamics, retention, endurance, wake-up
2. **Statistical rigor:** D2D and C2C variation, Monte Carlo array simulations
3. **CIM-specific metrics:** MLC programming (7+ levels), BER, MVM accuracy, conductance linearity
4. **Array-level validation:** Standard test patterns (checkerboard, walking, march), sneak path modeling

**A simulator that passes these tests can credibly claim research-grade accuracy** and enable meaningful comparison to experimental ferroelectric CIM devices.

---

**Document Status:** Research synthesis complete (2026-02-13)
**Next Steps:** Map simulator capabilities against this checklist, identify validation gaps, prioritize test implementation
