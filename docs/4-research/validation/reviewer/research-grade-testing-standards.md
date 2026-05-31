# Research-Grade Ferroelectric Memory Testing Standards

**Document Purpose**: Standards and metrics for validating HZO-based ferroelectric compute-in-memory (FeCIM) simulators against research-grade expectations.

**Last Updated**: 2025-02-13
**Research Period**: 2024-2025 literature

---

## Executive Summary

This document compiles research-grade testing standards for ferroelectric memory devices (FeFET, FTJ, FeCIM) based on recent peer-reviewed literature (2024-2025). These standards represent what a physics reviewer would expect in a credible ferroelectric CIM characterization paper.

**Key Finding**: A research-grade simulator should validate against these protocols to demonstrate physics fidelity, not just functional correctness.

---

## 1. Standard Ferroelectric Characterization Tests

### 1.1 P-E Hysteresis Loop Measurements

**PUND (Positive-Up-Negative-Down) Protocol** - Gold standard for isolating ferroelectric switching current

**Measurement Setup**:
- Five triangular pulses: ±Vmax with identical rise/fall times
- Pre-switch pulse aligns ferroelectric domains
- Separates switching from non-switching currents (leakage, capacitive)
- Removes dielectric and leakage contributions to extract remanent polarization

**Variants**:
- **DHM (Dynamic Hysteresis Measurement)**: Shows all current contributions
- **DLCC (Dynamic Leakage Current Compensation)**: Corrects for leakage via software
- **PUND**: Isolates ferroelectric contribution only

**Standard Practice**: PUND with bipolar trapezoidal cycles, 3 μs pulse duration for wake-up and endurance tests

**Nano-Scale Extension**: Nano-PUND combines AFM spatial resolution with current probing immunity to electrostatic artifacts

**Sources**:
- [Nano-positive up negative down in binary oxide ferroelectrics](https://pubs.aip.org/aip/apm/article/12/2/021112/3265432/Nano-positive-up-negative-down-in-binary-oxide)
- [Novel Electrical Characterization Method for Antiferroelectrics](https://arxiv.org/html/2501.05358v1)
- [An Original Positive-Up-Negative-Down Protocol](https://pubs.acs.org/doi/10.1021/acs.nanolett.5c00851)

---

### 1.2 Switching Dynamics

**Nucleation-Limited Switching (NLS) Model** - Explains polycrystalline HZO switching dynamics

**Key Characteristics**:
- Lorentzian distribution of logarithmic domain-switching times
- Sub-nanosecond switching: 925 ps demonstrated in HZO
- Time-resolved characterization via synchrotron X-ray diffraction (2-200 kHz range)
- Coupling between polarization and structural motion limited by domain wall velocity

**Measurement Techniques**:
- Transient current integration via pulse switching method
- Domain imaging via Piezoresponse Force Microscopy (PFM)
- Operando synchrotron-based time-resolved X-ray diffraction

**What to Validate**:
- Switching time distribution shape (Lorentzian vs Gaussian)
- Field-dependent switching kinetics
- Temperature dependence of switching speed

**Sources**:
- [Nucleation-Limited Switching Dynamics Model](https://ieeexplore.ieee.org/document/9640228/)
- [Time-Domain Perspective on Structural Response](https://pmc.ncbi.nlm.nih.gov/articles/PMC11311524/)
- [Ferroelectric Polarization-Switching Dynamics](https://pubs.acs.org/doi/10.1021/acsami.8b11681)

---

### 1.3 Retention Testing Protocols

**Standard Requirement**: >10 years retention at operating temperature

**Extrapolation Methodology**:
- Arrhenius equation for thermal acceleration
- High-temperature bake (85°C, 125°C, 200°C)
- Extract activation energy (Ea) from retention loss vs temperature
- Extrapolate to room temperature or 85°C for 10-year prediction

**Example Performance Targets**:
- 10^4 s measured retention → extrapolate to 10 years at 85°C
- Activation energy ~1.05 eV → 280 years MTTF at room temperature
- 82% retention after 10 years at 200°C (aggressive spec)

**Retention Loss Mechanisms**:
- Charge trapping at interfaces
- Oxygen vacancy redistribution
- Depolarization field buildup
- Domain wall pinning/depinning

**What Reviewers Expect**:
- Measured retention data at multiple temperatures
- Arrhenius plot with fitted activation energy
- Extrapolated 10-year retention estimate with error bars
- Discussion of failure mechanism (charge vs structural)

**Sources**:
- [Retention Improvement of HZO-Based Capacitors](https://pubs.acs.org/doi/10.1021/acsomega.2c06237)
- [Arrhenius equation for reliability (JEDEC)](https://www.jedec.org/standards-documents/dictionary/terms/arrhenius-equation-reliability)
- [Evaluation of Data Retention (NASA)](https://nepp.nasa.gov/DocUploads/76E2BAEB-6580-4537-AE532AB80A1A4191/Evaluation_Data_Retention_Alex_Sharma.pdf)

---

### 1.4 Endurance/Fatigue Cycling

**Performance Benchmarks**:
- AFE FTJs: >10^9 program/erase cycles
- FeFETs: >10^11 cycles with stable retention >2×10^4 s
- FeRAM commercial requirement: 10^15 cycles

**Fatigue Mechanisms**:
- Charge trapping and trap generation (fundamental cause)
- Interface trap generation under stress cycling
- Grain boundary and interface defect effects
- Wake-up effect (initial improvement before degradation)

**Test Protocol**:
- Cycling with bipolar pulses at operational voltage
- Periodic retention checks (e.g., every 10^6 cycles)
- Memory window tracking vs cycle count
- Identify onset of catastrophic failure

**Advanced Characterization**:
- Interface engineering to reduce initial traps
- Measure interface trap generation rate under cycling
- Separate bulk vs interface contributions to fatigue

**Sources**:
- [Recent advances in ferroelectric materials](https://pmc.ncbi.nlm.nih.gov/articles/PMC12592630/)
- [Retention and Endurance of FeFET Memory Cells](https://www.researchgate.net/publication/333922597_Retention_and_Endurance_of_FeFET_Memory_Cells)
- [High Endurance FeFET Without Retention Penalty](https://www.researchgate.net/publication/326727783_High_Endurance_Ferroelectric_Hafnium_Oxide-Based_FeFET_Memory_Without_Retention_Penalty)

---

### 1.5 Wake-Up and Imprint Characterization

**Wake-Up Effect**: Initial improvement in ferroelectric response during early cycling

**Characterization Methods**:
- Track Pr/Ec vs cycle number (first 10^3-10^6 cycles)
- Quantify activation energy for wake-up reversal (oxygen vacancy redistribution)
- Temperature acceleration studies (bake testing)

**Imprint Effect**: Hysteresis loop shift along E-field axis, degrades retention

**Static Imprint**: Fixed loop shift at zero field
**Dynamic Imprint**: Field-dependent loop shift under AC operation

**Advanced Technique (2025)**:
- **FORC (First-Order Reversal Curve) Analysis**: Systematically characterizes dynamic imprint
- Measures coercive field distribution changes
- Separates domain switching contributions
- Combined with transient I-V and P-V measurements

**What to Report**:
- Wake-up saturation cycle count
- Imprint magnitude (ΔEc) vs time/cycles
- Temperature dependence of imprint formation
- Recovery mechanisms and timescales

**Sources**:
- [Dynamic Imprint and Recovery Mechanisms (2025)](https://www.mdpi.com/2079-9292/14/23/4593)
- [Wake-Up Reversal Behavior in La:HZO](https://ieeexplore.ieee.org/document/10079178/)
- [Huge Reduction of Wake-Up Effect](https://pubs.acs.org/doi/10.1021/acsaelm.9b00367)

---

### 1.6 Temperature Dependence

**Critical Parameters to Characterize**:
- Ec(T): Coercive field vs temperature
- Pr(T): Remanent polarization vs temperature
- Ps(T): Saturation polarization vs temperature
- Retention(T): Activation energy extraction

**Temperature Ranges**:
- Cryogenic (40 K) → Room temperature (300 K) → Elevated (85°C-200°C)
- Polar phase stability at low T
- Domain depinning at elevated T

**Recent Findings (2024-2025)**:
- {111} texture transfer temperature dependence (200-300°C deposition)
- Polar phase stabilization at low T, depinning at high T
- Crystallographic orientation effects on coercive field

**Sources**:
- [New Insights at Cryogenic Temperatures (2026)](https://advanced.onlinelibrary.wiley.com/doi/10.1002/apxr.202500127)
- [Temperature-Dependent {111}-Texture Transfer](https://pubs.acs.org/doi/10.1021/acsami.4c17978)

---

### 1.7 Device-to-Device Variability (Statistical Testing)

**Critical Metrics**:
- Coercive voltage (Vc): σ/μ ~ 0.019-0.023 (Gaussian distribution)
- Cycle-to-cycle variation: ~0.3% (state-of-art 2D ferroelectrics)
- Device-to-device variation: ~0.5%
- Memory window (MW) variability: σ/μ depends on gate length scaling

**Variability Sources**:
- Random ferroelectric-dielectric (FE-DE) phase distribution
- Grain size and grain boundary density
- Interface trap distribution
- Channel percolation paths
- Nucleation-limited domain growth stochasticity

**Test Protocol**:
- Minimum 20-50 devices for statistical significance
- Monte Carlo simulations fitted to measured distributions
- Report mean, standard deviation, and distribution shape
- Scaling study: variability vs device dimensions

**Mitigation Strategies (Research-Grade)**:
- Microwave annealing
- HfO2 interfacial layer insertion
- Minimize monoclinic phase fraction
- Control grain size uniformity

**Sources**:
- [Random and Systematic Variation in Nanoscale FeFETs](https://www.frontiersin.org/journals/nanotechnology/articles/10.3389/fnano.2021.826232/full)
- [Variability Analysis for FeFET Memory](https://www.researchgate.net/publication/338661174_Variability_Analysis_for_Ferroelectric_FET_Nonvolatile_Memories_Considering_Random_Ferroelectric-Dielectric_Phase_Distribution)
- [Two-dimensional ferroelectric-gated hybrid CIM](https://www.science.org/doi/10.1126/sciadv.adp0174)

---

## 2. CIM-Specific Test Protocols

### 2.1 Multi-Level Cell (MLC) Programming Accuracy

**ISPP (Incremental Step Pulse Programming)** - Standard method for MLC programming

**Protocol**:
1. Apply initial programming voltage (V_start)
2. Verify: Measure current at fixed read voltage (V_read)
3. If target not reached: V_prog += ΔV (typically 0.05 V step)
4. Repeat until target conductance achieved or max iterations

**Key Metrics**:
- Number of distinct states achieved (e.g., 7 states = ~3 bits/cell, 32 states = 5 bits)
- Programming accuracy: distance from target level
- Overshoot probability and magnitude
- Convergence speed: iterations to target

**Advanced Variants**:
- **L-ISPP (Logarithmic ISPP)**: Reduces energy, improves linearity
- **Program/Verify with adaptive V_start**: Uses previous pulse result to set next starting voltage

**What Limits MLC Resolution**:
- Overshoot due to stochastic switching
- Read noise and variability
- State drift over time
- Device-to-device variation

**Demonstrated Performance (2024-2025)**:
- 1k FeFET crossbar: 7 distinct VT states (~3 bits/cell)
- HZO FTJ: 32 conductance states (5 bits), 1.6% cycle-to-cycle variation, nonlinearity <1

**Sources**:
- [Reliable multi-level cell programming in FeFET arrays](https://www.researchgate.net/publication/390350071_Reliable_multi-level_cell_programming_in_FeFET_arrays_for_in-memory_computing)
- [First demonstration of in-memory computing crossbar](https://www.nature.com/articles/s41467-023-42110-y)
- [Pulse program for improving learning accuracy](https://www.sciencedirect.com/science/article/abs/pii/S1567173924001755)

---

### 2.2 Read Disturb Characterization

**Problem**: FeFETs share write/read paths → read operations can alter stored state

**Test Protocol**:
1. Program cell to target state
2. Perform N read operations (e.g., 10^6, 10^9)
3. Measure state degradation vs read count
4. Characterize at different read voltages

**Non-Destructive Read (NDR) Requirement**:
- Commercial FeRAM target: >10^11 NDR endurance
- Recent FeCAP demonstration: >10^11 NDR with 8.7 capacitive memory window

**What to Report**:
- Read endurance before detectable state change
- Read voltage margin (max V_read that maintains >10^11 cycles)
- Temperature dependence of read disturb

**Sources**:
- [A non-destructive readout mechanism](https://www.imec-int.com/en/articles/non-destructive-readout-mechanism-ferroelectric-capacitors-0)
- [Quasi-Nondestructive Read Out](https://www.osti.gov/servlets/purl/2421736)

---

### 2.3 Write Disturb / Half-Select Disturb

**Problem**: In crossbar arrays, unselected cells experience partial write voltage

**Inhibit Schemes**:
- **VDD/2 scheme**: Unselected cells see VGB = 0 or VDD/2
- **VDD/3 scheme**: Unselected cells see VGB = ±VDD/3

**Critical Finding**: For FeFETs where |VW1| > |VW0|, only writing logical '1' causes disturb

**Test Protocol**:
1. Program entire array to known pattern (e.g., checkerboard)
2. Write single cell repeatedly (worst-case aggressor)
3. Measure victim cell state degradation vs write cycles
4. Test both VDD/2 and VDD/3 schemes

**What to Report**:
- Maximum write cycles before victim cells flip
- Optimal inhibit voltage for minimum disturb
- Array size limit based on worst-case disturb

**Sources**:
- [Write Disturb in Ferroelectric FETs](https://ieeexplore.ieee.org/document/8474379/)
- [C-AND: Mixed Writing Scheme for Disturb Reduction](https://www.researchgate.net/publication/360834027_C-AND_Mixed_Writing_Scheme_for_Disturb_Reduction_in_1T_Ferroelectric_FET_Memory)

---

### 2.4 Analog Compute Accuracy (MVM Error Metrics)

**Matrix-Vector Multiplication (MVM) Fidelity**

**Key Error Sources**:
1. Device-to-device variation in weight values
2. Cycle-to-cycle (pulse-to-pulse) variation
3. Read noise (ADC quantization)
4. IR drop in crossbar (see Section 3 below)
5. Sneak path currents

**Accuracy Metrics**:
- **MAC Error**: |MAC_ideal - MAC_actual| / MAC_ideal
- **Inference Accuracy**: MNIST, CIFAR-10 classification accuracy
- **Bit Error Rate (BER)**: Probability of incorrect state read

**Demonstrated Performance (2024-2025)**:
- 96.6% MNIST accuracy (multi-level FeFET crossbar)
- 91.5% CIFAR-10 accuracy (image classification)
- 4% BER in inference-only with 1% degradation from floating-point

**Hardware-Software Co-Design**:
- Variation-aware training (Bayesian Neural Networks)
- Error-resilient architectures
- Peripheral circuit optimization (sense amp design)

**Sources**:
- [First demonstration of in-memory computing crossbar](https://pmc.ncbi.nlm.nih.gov/articles/PMC10564859/)
- [Ferroelectric capacitors for ML workloads](https://www.nature.com/articles/s41598-024-59298-8)
- [Two-dimensional ferroelectric-gated hybrid CIM](https://www.science.org/doi/10.1126/sciadv.adp0174)

---

### 2.5 Linearity of Conductance States

**Critical for On-Chip Training**: Neural networks require linear, symmetric weight updates

**Linearity Metrics**:
- **Nonlinearity Factor (NL)**: Deviation from ideal linear fit
  - State-of-art: |NLp| ~ 0.01-0.016 (potentiation), |NLd| ~ 0.01-0.013 (depression)
- **R² Goodness of Fit**: Should approach 0.999
- **Symmetry**: |NLp - NLd| < 0.01 (asymmetry metric)

**Performance Targets for Efficient Training**:
- ≥32 conductance states (≥5-bit precision)
- Gmax/Gmin > 10 (dynamic range)
- Linear and symmetric weight updates
- Low operation voltage ≤5 V

**Demonstrated Performance**:
- HZO FTJ: 32 states, 1.6% variation, NL < 1
- 2D vdW ferroelectrics: NLp = 0.016, NLd = 0.013, asymmetry = 0.003
- FeFET with ITO channel: R² = 0.999, NL = 0.01/-0.01

**Sources**:
- [Van der Waals ferroelectric transistors](https://www.sciencedirect.com/science/article/pii/S2709472323000072)
- [High Linearity and Symmetry Ferroelectric Devices (2025)](https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202500078)
- [High-precision and linear weight updates](https://www.nature.com/articles/s41467-022-28303-x)

---

### 2.6 State Retention After Compute Operations

**Problem**: CIM operations may degrade stored weights over time

**Test Protocol**:
1. Program array to weight matrix
2. Perform N compute operations (MVM cycles)
3. Measure weight drift vs operation count and time
4. Separate thermal drift from operation-induced drift

**Key Metrics**:
- Weight drift rate (%/hour, %/cycle)
- Retention time for ±1 LSB accuracy
- Temperature dependence (85°C aging test)

**Challenges**:
- Stable polarization states difficult to maintain at room T (incipient ferroelectricity)
- Trap-assisted relaxation
- Domain wall motion under repeated field cycling

**Sources**:
- [Ferroelectricity-Defects Synergistic Artificial Synapses](https://pubs.acs.org/doi/10.1021/acsami.4c01489)
- [Physical modeling of HZO-based FeFETs](https://www.frontiersin.org/journals/nanotechnology/articles/10.3389/fnano.2022.900592/full)

---

### 2.7 Bit Error Rate at Different Levels

**Definition**: Probability of reading incorrect state

**BER Test Protocol**:
1. Program large array (≥1k cells) to each target level
2. Immediately read all cells → measure fresh BER
3. Age at temperature (85°C, 125°C) → measure aged BER
4. Repeat for all MLC levels

**Expected BER Profile**:
- Mid-range levels: higher BER (closer to decision boundaries)
- Extreme levels (min/max G): lower BER (well-separated)
- BER increases with temperature and retention time

**Hardware-Level Mitigation**:
- Error correction codes (ECC)
- Wider programming margins
- Adaptive read thresholds

**Demonstrated Performance**:
- 4% BER with 1% accuracy loss (CIFAR-10 inference)
- High on/off ratio (>10^7) enables low BER at extremes

**Sources**:
- [Ferroelectric capacitors for ML workloads](https://www.nature.com/articles/s41598-024-59298-8)
- [Variation-Resilient FeFET-Based CIM](https://arxiv.org/html/2312.15444v1)

---

## 3. Reviewer-Expected Metrics

**What a ferroelectric CIM paper MUST include:**

### 3.1 Programming Window / Memory Window

**Definition**: Separation between programmed states

**For FeFET**: Memory Window (MW) = VTH,high - VTH,low
**For FTJ**: MW = (Ghigh - Glow) / Gavg or Rhigh/Rlow ratio

**Targets**:
- Large MW (>1 V for FeFET, >10× for conductance ratio)
- Stable MW over endurance cycling
- Minimal MW degradation with gate length scaling

**Variability Consideration**: Report MW distribution across devices (σ/μ)

---

### 3.2 On/Off Ratio

**Definition**: Gmax / Gmin or Imax / Imin

**Targets**:
- State-of-art 2D ferroelectrics: >10^7
- Minimum for CIM: >10 (for MLC operation)
- Required for reliable MLC: >100

**Impact**:
- High on/off ratio → low BER
- Determines number of reliably distinguishable states
- Affects ADC requirements for current sensing

---

### 3.3 Endurance (Cycles to Failure)

**Definitions**:
- **Hard Failure**: Complete loss of switchability
- **Soft Failure**: MW degradation below usable threshold (e.g., 50% initial MW)

**Industry Targets**:
- Embedded NVM: 10^5 cycles
- Stand-alone NVM: 10^6-10^9 cycles
- FeRAM commercial: 10^15 cycles

**Report Format**:
- Endurance curve: MW vs cycle count (log scale)
- Failure criterion clearly defined
- Cycling conditions (voltage, frequency, temperature)

---

### 3.4 Retention (Extrapolated at 85°C or 125°C)

**Standard Test Temperature**: 85°C (JEDEC automotive spec)
**Aggressive Spec**: 125°C or higher

**Reporting Requirements**:
- Measured retention at multiple T (room T, 85°C, 125°C)
- Arrhenius plot with linear fit
- Extracted activation energy (Ea)
- **Extrapolated 10-year retention** with confidence interval

**Example Claim**: "Extrapolated retention >10 years at 85°C with Ea = 1.05 eV (R² = 0.98)"

---

### 3.5 Switching Energy

**Measurement**: Energy per bit flip (program operation)

**Reported Units**: fJ/bit (femtojoules) or pJ/bit

**State-of-Art Performance**:
- Silicon-channel FeFET: ~100 fJ/bit
- FeRAM: 50 fJ/bit
- FeFET edge applications: fJ/bit with 50 ns write latency

**Calculation**:
```
E_switch = ∫ V(t) × I(t) dt
```
Integrate voltage × current over write pulse duration

**What Matters**:
- Total energy (not just CV² capacitive component)
- Include resistive losses in array
- Report at worst-case conditions (maximum array size)

**Sources**:
- [Hafnium oxide-based FeFETs](https://pubs.aip.org/aip/jap/article/138/1/010701/3351745/Hafnium-oxide-based-ferroelectric-field-effect)
- [Ferroelectric field effect transistors for electronics](https://pubs.aip.org/aip/apr/article/10/1/011310/2881283/Ferroelectric-field-effect-transistors-for)

---

### 3.6 Read/Write Latency

**Definitions**:
- **Write Latency**: Time to program a cell to target state
- **Read Latency**: Time to sense and convert stored state to digital output

**Reported Performance**:
- FeFET: <10 ns switching, 50-100 ns write latency
- SK hynix PCRAM (reference): 385 ns read / 100 ns write
- FeRAM commercial: 10 ns read/write access time

**Measurement Considerations**:
- Exclude peripheral overhead (decoder, ADC) for device-level characterization
- Include peripheral for system-level benchmarking
- Report at different MLC levels (higher levels may require longer verify)

**Sources**:
- [What Ever Happened To Ferroelectric Memories?](https://marklapedus.substack.com/p/what-ever-happened-to-next-gen-ferroelectric)
- [Ferroelectric RAM performs calculations](https://techxplore.com/news/2025-04-ferroelectric-device-memory.html)

---

## 4. Standard Test Patterns for Crossbar Arrays

### 4.1 Checkerboard Pattern

**Purpose**: Detects maximum current leakage, worst-case data retention faults

**Pattern**: Alternating 0/1 in 2D grid
```
0 1 0 1
1 0 1 0
0 1 0 1
1 0 1 0
```

**What It Tests**:
- Crosstalk between adjacent cells
- Maximum sneak path currents
- Leakage-induced data retention faults
- Shorts between neighboring cells

**Expected Use**: Pre-deployment array validation, periodic health checks

**Sources**:
- [State of Memory Pattern Assessment (NASA)](https://nepp.nasa.gov/docs/papers/2021-Murphy-Memory-Pattern-Assessment-Report-CL-22-1768.pdf)
- [Basics of Memory Testing in VLSI](https://vlsiuniverse.com/basics-of-memory-testing-in-vlsi-memory/)

---

### 4.2 All-Ones / All-Zeros

**Purpose**: Detect stuck-at faults, uniform stress test

**All-Ones (1-fill)**: Every cell programmed to high conductance
**All-Zeros (0-fill)**: Every cell programmed to low conductance

**What It Tests**:
- Cells stuck at opposite state (stuck-at-0 or stuck-at-1)
- Uniform write stress across entire array
- Total array current draw (power validation)
- IR drop uniformity

**Use Case**: Manufacturing test, burn-in stress test

---

### 4.3 Walking Ones/Zeros

**Walking Ones**: Single '1' cell moves through all-zero background
**Walking Zeros**: Single '0' cell moves through all-one background

**Purpose**: Detect single-cell addressing faults

**What It Tests**:
- Decoder functionality (each address uniquely selectable)
- Isolated cell write/read without disturb
- Positional faults (e.g., address bit inversions)

**Implementation**:
```
Cycle 1: 10000000
Cycle 2: 01000000
Cycle 3: 00100000
...
```

---

### 4.4 Worst-Case Sneak Path Pattern

**Problem**: In passive crossbar (no selector), current sneaks through unselected paths

**Worst-Case Condition**: Maximum number of low-resistance parallel sneak paths

**Example Pattern** (reading cell [0,0] in 4×4 array):
```
Target cell at [0,0] is HIGH resistance (storing '0')
All other cells LOW resistance (storing '1')
→ Maximum sneak current through alternate paths
```

**IR Drop Validation**:
- Checkerboard pattern triggers maximum IR drop
- Worst-case voltage drop at cell farthest from driver
- IR drop limits maximum array size

**Quantification**:
- Model sneak current vs array size (N×N)
- Measure IR drop from driver to worst-case cell
- Define maximum N before read margin collapses

**Sources**:
- [Sneak-Path Testing of Crossbar Memories](https://www.researchgate.net/publication/258793659_Sneak-Path_Testing_of_Crossbar-Based_Nonvolatile_Random_Access_Memories)
- [Sneak Path Characterization in Memristor Crossbar](https://www.academia.edu/99538016/Sneak_Path_Characterization_in_Memristor_Crossbar_Circuits)
- [Post-Moore Memory Technology: SPC Phenomena](https://www.mdpi.com/2072-666X/12/1/50)

---

### 4.5 March Test Algorithms

**March Tests**: Systematic read/write sequences to detect fault models

**Common Fault Models**:
- **Stuck-At Fault (SAF)**: Cell cannot change from 0 or 1
- **Transition Fault (TF)**: Cell fails to transition during write (0→1 or 1→0)
- **Coupling Fault (CF)**: Aggressor cell write affects victim cell
  - State coupling, idempotent coupling, inversion coupling
  - Transition coupling, read destructive coupling

**Standard March Algorithms**:
- **March C-**: O(10N) complexity, detects SAF, TF, coupling faults
- **March AZ1**: O(13N) complexity, improved transition coupling detection
- **March PCM**: Optimized for phase-change memory (applicable to ferroelectric)

**March Element Notation**:
- ⇑: Ascending address order
- ⇓: Descending address order
- w0: Write 0
- w1: Write 1
- r0: Read 0 (expected)
- r1: Read 1 (expected)

**Example March C- Algorithm**:
```
{⇓(w0); ⇑(r0,w1); ⇑(r1,w0); ⇓(r0,w1); ⇓(r1,w0); ⇑(r0)}
```

**Application to Ferroelectric CIM**:
- Adapted for multi-level cells (beyond binary 0/1)
- Validates MLC transition faults
- Detects ferroelectric-specific issues (incomplete switching, imprint)

**Sources**:
- [A review paper on memory fault models (2024)](https://wseas.com/journals/eeacs/2024/a34eeacs-010(2024).pdf)
- [RRAM Defect Modeling and Failure Analysis](https://ieeexplore.ieee.org/iel7/12/6980145/06725492.pdf)
- [March-based RAM diagnosis algorithms](https://www.researchgate.net/publication/261238099_March-based_RAM_diagnosis_algorithms_for_stuck-at_and_coupling_faults)

---

## 5. Advanced Characterization Techniques

### 5.1 Domain Wall Pinning/Depinning (2025 Research)

**Mechanism**: Point defects (oxygen vacancies) temporarily pin domain wall segments

**Characterization Methods**:
- **Ab initio simulations**: Predict pinning energy landscape
- **Elastic energy loss spectroscopy**: Measure activation energies for VO jumps
- **ML-controlled automated PFM**: High-throughput switching event analysis (1500+ events)

**Key Findings (2025)**:
- 1% defect concentration pins significant wall fractions
- Activation energy controls clamping, fatigue, and mobility
- Domain wall velocity reaches steady state when pinning/depinning balances

**What to Report**:
- Activation energy for depinning (eV)
- Defect concentration effects on wall mobility
- Field dependence of wall velocity

**Sources**:
- [Control of ferroelectric domain wall dynamics (2025)](https://pubs.aip.org/aip/jap/article/137/15/154103/3344377/Control-of-ferroelectric-domain-wall-dynamics-by)
- [Oxygen vacancies in ferroelectrics (2025)](https://journals.aps.org/prb/abstract/10.1103/PhysRevB.111.054112)
- [Exploring Domain Wall Pinning via AFM (2025)](https://arxiv.org/abs/2505.24062)

---

### 5.2 Thickness Scaling Effects

**Critical Finding**: Aggressive scaling (<10 nm) amplifies grain boundary and interface defects

**Characterization Protocol**:
- Fabricate thickness series: 3 nm, 5 nm, 8 nm, 10 nm, 15 nm
- Measure Pr, Ec, leakage vs thickness
- Characterize grain size distribution (SEM, TEM)
- XRD for phase composition vs thickness

**Key Observations**:
- Grain size decreases with thickness → more grain boundaries
- Trap density increases (in-plane and out-of-plane)
- FTJ memory window fails to improve with scaling (polarization degrades)
- Device-to-device variability worsens with scaling

**Optimal Thickness Window**: 8-12 nm for HZO (balances polarization, leakage, variability)

**Sources**:
- [Thickness scaling of HZO films (2025)](https://www.sciencedirect.com/science/article/abs/pii/S0925838825038113)
- [Effect of HZO Thickness Scaling in FTJ](https://pubs.acs.org/doi/10.1021/acsaelm.5c00469)
- [Mesoscopic-scale grain formation](https://link.springer.com/article/10.1186/s40580-022-00342-6)

---

### 5.3 Cycle-to-Cycle and Pulse-to-Pulse Variation

**Measurement Setup**:
- Program single cell to target level repeatedly (100+ cycles)
- Measure conductance after each program pulse
- Extract statistical distribution (mean, σ, shape)

**State-of-Art Performance**:
- 2D ferroelectrics: 0.3% cycle-to-cycle, 0.5% device-to-device
- HZO FTJ: 1.6% cycle-to-cycle (32-state operation)
- HZO artificial synapse: 2.75% (128-state conductance)

**Reporting Format**:
- σ/μ ratio for each MLC level
- Distribution shape (Gaussian, Lorentzian)
- Correlation: Does variation increase with target level?

**Sources**:
- [Switching Dynamics of Ferroelectric HZO](https://strand.nd.edu/sites/default/files/publications/2024-10/2018%20Cristobal%20Alessandri%20EDL.pdf)
- [Random and Systematic Variation in Nanoscale FeFETs](https://www.frontiersin.org/journals/nanotechnology/articles/10.3389/fnano.2021.826232/full)

---

## 6. Neural Network Benchmark Validation

### 6.1 MNIST Handwritten Digit Recognition

**Standard Benchmark**: 60k training, 10k test images (28×28 grayscale)

**Network Architectures**:
- **MLP5**: Simple 5-layer perceptron
- **LeNet**: Classic CNN (conv-pool-conv-pool-fc-fc)

**Performance Targets**:
- **Ideal (floating-point)**: 98-99%
- **Ferroelectric hardware**: >95% (within 1-4% of ideal)

**Reported Performance (2024-2025)**:
- FeFET crossbar: 96.6%
- Ferroelectric domain-wall synapse: 98.7%
- FeFET with variation mitigation: >97%

**Training Methodology**:
- Custom PyTorch layers incorporating crossbar MAC
- 100 epochs, SGD with momentum 0.9, weight decay 10^-4
- Report both training and inference accuracy

---

### 6.2 CIFAR-10 Image Classification

**Benchmark**: 50k training, 10k test (32×32 RGB, 10 classes)

**Network Architectures**:
- **ResNet-20**: Residual network (deeper, harder)
- **VGG-8**: Visual Geometry Group (wider)
- **AlexNet**: Classical deep CNN

**Performance Targets**:
- **Ideal (floating-point)**: 85-92% (depends on architecture)
- **Ferroelectric hardware**: Within 3-6% of ideal

**Reported Performance**:
- FeFET crossbar: 91.5%
- ResNet-20 on FeFET: 81.7%
- VGG-8 with 6-7 bit ENOB: within 1% of floating-point

**Challenge Factors**:
- Deeper networks amplify error accumulation
- Weight precision requirements higher than MNIST
- Sensitivity to device variation increases

---

### 6.3 Hardware-Software Co-Design

**Error Mitigation Techniques**:
1. **Variation-Aware Training**: Bayesian Neural Networks model device uncertainty
2. **Peripheral Optimization**: Sense amp design, ADC bit depth
3. **Error Correction**: Algorithm-level compensation for known device characteristics

**Effective Number of Bits (ENOB)**:
- Measure actual state distinguishability accounting for noise
- ENOB = log2(Gmax / noise_floor)
- Target: 6-7 ENOB for deep networks (CIFAR-10, ImageNet)

**Reporting Requirements**:
- Network architecture (layers, sizes, activations)
- Training recipe (optimizer, learning rate, epochs)
- Weight/activation quantization (bit precision)
- Hardware-specific constraints (ADC resolution, non-idealities)
- Comparison: Ideal vs hardware-aware training vs hardware inference

**Sources**:
- [Ferroelectric Domain-Wall Synapse](https://pubs.acs.org/doi/10.1021/acs.nanolett.5c02081)
- [Ferroelectric capacitors for ML workloads](https://www.nature.com/articles/s41598-024-59298-8)
- [Variation-Resilient FeFET CIM](https://arxiv.org/html/2312.15444v1)

---

## 7. Array Size Scalability Validation

### 7.1 Experimentally Demonstrated Array Sizes (2024-2025)

**Recent Demonstrations**:
- **24 × 24 FTJ crossbar** (576 devices, 2025)
- **32 × 32 inversion-type FeCAP** (1024 devices)
- **48 × 48 FTJ crossbar** (2304 devices, half-bias verified)
- **1-kbit crossbar** (inversion-type FCM on SOI)

**Scaling Validation**:
- Varied cell size (10×10 µm² to 70×70 µm²) to prove scalability
- Half-bias operation confirmed (unselected cells stable)
- Device uniformity maintained across array

**Array Size Limits**:
- IR drop becomes dominant in large passive crossbars
- Sneak current scales poorly with N² (N×N array)
- Requires 1T1C (one transistor, one capacitor) for large arrays

**What to Report**:
- Largest fabricated array size (N×N)
- Yield (% functional cells)
- Uniformity metrics across array (spatial distribution of Pr, Ec)
- Comparison: small array vs large array performance

**Sources**:
- [Ferroelectric memristor crossbar (2025)](https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963)
- [Ferroelectric capacitive memories arrays](https://link.springer.com/article/10.1186/s40580-024-00463-0)
- [FTJ Crossbar Array with Annealing Optimization](https://advanced.onlinelibrary.wiley.com/doi/10.1002/aisy.202500817)

---

### 7.2 IR Drop and Sneak Path Modeling

**IR Drop Problem**: Voltage drop across word/bit lines limits array size

**Worst-Case IR Drop**:
- Occurs at cell farthest from driver
- Depends on: array size, metal resistance, write current, parallel write count

**Modeling Approach**:
- Equivalent resistance network simulation
- Checkerboard pattern for maximum current draw
- Extract voltage at each cell location

**Sneak Path Current**:
- Increases with array size and number of low-resistance cells
- Formulas exist for sneak path count vs array size and resistance distribution
- Limits passive crossbar size to ~64×64 without selectors

**Validation Requirement**:
- Measure read current vs cell position in array
- Compare center cells vs edge cells
- Report maximum array size before read margin collapse

**Sources**:
- [Overcoming Challenges of Crossbar Architectures](https://users.cs.utah.edu/~rajeev/pubs/hpca15.pdf)
- [Sneak Path Current Modeling for CIM](https://arxiv.org/html/2511.21796v2)

---

## 8. Simulator Validation Checklist

**A research-grade FeCIM simulator should demonstrate:**

### Level 1: Basic Physics Validation
- [ ] P-E hysteresis matches experimental HZO data (Pr, Ps, Ec)
- [ ] PUND extraction implemented (separates switching from leakage)
- [ ] Retention extrapolation via Arrhenius (10-year @ 85°C)
- [ ] Endurance cycling with fatigue modeling (>10^6 cycles)
- [ ] Temperature dependence of Ec, Pr, Ps

### Level 2: Device-Level CIM Validation
- [ ] MLC programming with ISPP algorithm
- [ ] Device-to-device variability (Gaussian distributions, σ/μ < 0.05)
- [ ] Cycle-to-cycle variation (<5% for practical CIM)
- [ ] Read/write disturb quantification
- [ ] State retention after compute operations

### Level 3: Array-Level Validation
- [ ] Crossbar array modeling (at least 32×32)
- [ ] IR drop vs array size and resistance state
- [ ] Sneak path current modeling
- [ ] Checkerboard pattern worst-case validation
- [ ] March test fault coverage

### Level 4: System-Level Validation
- [ ] MVM accuracy with non-idealities
- [ ] MNIST benchmark (>95% accuracy)
- [ ] CIFAR-10 benchmark (within 5% of ideal)
- [ ] Weight update linearity metrics (NL < 0.1)
- [ ] BER vs MLC level

### Level 5: Advanced Characterization
- [ ] Domain wall dynamics (nucleation-limited switching model)
- [ ] Wake-up and imprint characterization
- [ ] Thickness scaling effects (grain size, interface traps)
- [ ] FORC analysis for dynamic imprint
- [ ] Switching energy calculation (fJ/bit)

---

## 9. Methodology for Simulator Claims

### 9.1 What Constitutes a Valid Claim

**Valid Simulator Claim**:
> "Simulator models ISPP convergence using Landau-Khalatnikov physics with parameters extracted from [Ref], achieving <2% error vs experimental HZO switching traces."

**Invalid Simulator Claim**:
> "Simulator achieves 98% MNIST accuracy" (without hardware validation or non-ideality modeling)

**Key Distinction**:
- **Physics fidelity**: Validated against experimental device measurements
- **Functional accuracy**: Validated against system-level benchmarks with modeled non-idealities

### 9.2 Required Experimental Validation

**For Each Modeled Phenomenon, Provide**:
1. **Reference Data**: Published experimental measurements (journal paper, conference, dataset)
2. **Parameter Extraction**: How model parameters were fitted to data
3. **Error Quantification**: RMSE, R², or other goodness-of-fit metric
4. **Validity Range**: Conditions under which model is accurate (voltage range, temperature, etc.)

**Example Documentation**:
```
Landau-Khalatnikov Switching Model:
- Reference: [Park et al., APL 2020] - HZO switching dynamics
- Parameters: α = 1.2e8, β = -3.4e9, γ = 1.1e10 (SI units)
- Fitted to: Fig 3(b) P-E loops at 1 kHz, 300 K
- RMSE: 1.8 µC/cm² (2.1% relative error)
- Valid for: E = 1-4 MV/cm, f = 100 Hz - 10 kHz, T = 250-350 K
```

### 9.3 Uncertainty Propagation

**When Claiming System-Level Accuracy**:
- Propagate device-level variation through system model
- Monte Carlo simulation with measured σ/μ distributions
- Report confidence intervals, not just mean accuracy

**Example**:
> "MNIST accuracy: 96.2% ± 0.8% (95% CI, N=100 Monte Carlo runs with σ/μ = 0.03 device variation)"

---

## 10. Comparison to Commercial Standards

### 10.1 JEDEC Standards

**Relevant JEDEC Specs**:
- **JESD22**: Environmental test methods (temperature cycling, bake, HAST)
- **JESD57**: "Worst-case" conditions for memory testing
- **Retention**: 10 years at 85°C (commercial NVM requirement)

**Note**: HZO ferroelectrics not yet JEDEC-standardized (emerging technology)

**Sources**:
- [JEDEC Arrhenius equation for reliability](https://www.jedec.org/standards-documents/dictionary/terms/arrhenius-equation-reliability)

### 10.2 FeRAM Industry Targets (Reference Benchmark)

**Texas Instruments FeRAM Commercial Specs**:
- 10 ns read/write access time
- 10^15 endurance cycles
- 6F² cell size
- >10 year retention
- 50 fJ/bit energy

**HZO FeFET Comparison (Research Phase)**:
- <10 ns switching demonstrated
- 10^11 endurance (FeFET), 10^9 (FTJ) demonstrated
- 10 year retention extrapolated (not long-term validated)
- ~100 fJ/bit demonstrated
- Not yet commercialized

**Sources**:
- [Framework for Embedded NVM Development](https://www.mdpi.com/2079-9292/14/4/818)

---

## 11. Open Research Questions (2025)

**Areas Where Standards Are Still Evolving**:

1. **MLC Retention**: Long-term stability of intermediate states under-studied
2. **CIM-Specific Endurance**: Does repeated MAC operation degrade faster than binary cycling?
3. **Thermal Budget**: Integration with back-end-of-line (BEOL) CMOS processes
4. **3D Scaling**: Vertical integration effects on ferroelectric properties
5. **Selector Device Co-Integration**: 1S1C (one selector, one capacitor) standardization
6. **Mixed-Signal Peripheral Circuits**: ADC/DAC design for optimal MLC sensing

---

## 12. Summary: Research-Grade Validation Pyramid

```
                    Level 5: Advanced Physics
                   (Domain walls, FORC, thickness scaling)
                  /                                        \
          Level 4: System Validation                        \
        (MNIST/CIFAR, MVM accuracy, BER)                     \
       /                                    \                 \
  Level 3: Array-Level                      \                 \
 (IR drop, sneak paths, march tests)         \                 \
/                              \               \                \
Level 2: Device CIM             \               \                \
(MLC, disturb, linearity, drift) \               \                \
                                  \               \                \
                        Level 1: Basic Physics     \                \
                    (P-E loops, retention, endurance, temp)          \
```

**Minimum for "Research-Grade" Claim**: Levels 1-3 validated
**Publication-Quality**: Levels 1-4 validated with statistical rigor
**State-of-Art**: Level 5 includes novel physics insights

---

## References Summary

### Characterization Methods
- [Nano-PUND in binary oxide ferroelectrics](https://pubs.aip.org/aip/apm/article/12/2/021112/3265432/Nano-positive-up-negative-down-in-binary-oxide)
- [Nucleation-Limited Switching Model](https://ieeexplore.ieee.org/document/9640228/)
- [FORC Analysis for Dynamic Imprint (2025)](https://www.mdpi.com/2079-9292/14/23/4593)

### Device Performance
- [First demonstration of MLC FeFET crossbar](https://www.nature.com/articles/s41467-023-42110-y)
- [Two-dimensional ferroelectric-gated hybrid CIM](https://www.science.org/doi/10.1126/sciadv.adp0174)
- [High Linearity Ferroelectric Devices (2025)](https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202500078)

### Array Scaling
- [Ferroelectric memristor crossbar (2025)](https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963)
- [Sneak Path Current Modeling](https://arxiv.org/html/2511.21796v2)
- [Write Disturb in FeFETs](https://ieeexplore.ieee.org/document/8474379/)

### System Validation
- [Ferroelectric capacitors for ML workloads](https://www.nature.com/articles/s41598-024-59298-8)
- [Variation-Resilient FeFET CIM](https://arxiv.org/html/2312.15444v1)

### Reliability
- [Retention Improvement of HZO Capacitors](https://pubs.acs.org/doi/10.1021/acsomega.2c06237)
- [Recent advances in ferroelectric materials (2024)](https://pmc.ncbi.nlm.nih.gov/articles/PMC12592630/)

---

## Document Maintenance

**Version**: 1.0
**Date**: 2025-02-13
**Maintainer**: FeCIM Lattice Tools validation team
**Review Cycle**: Quarterly (track new publications)

**Contributing**: To add new standards or update metrics, submit references from peer-reviewed journals (IEEE, Nature, ACS, Wiley) published within last 2 years.
