# IronLattice Research Findings - Part 2 (January 2026)

Additional research findings on STDP, SNNs, technology comparisons, on-chip learning, and simulation parameters.

---

## 1. STDP (Spike-Timing Dependent Plasticity)

### Hardware Implementation with FeFET

STDP is a biologically-inspired learning rule where synaptic weight changes depend on the relative timing of pre- and post-synaptic spikes:

```
Δw = A_+ × exp(-Δt/τ_+)  if Δt > 0 (pre before post)
Δw = -A_- × exp(Δt/τ_-)   if Δt < 0 (post before pre)
```

### FeFET STDP Demonstrations

| Device | Achievement | Year | Source |
|--------|-------------|------|--------|
| 2D SnS2 FeFET | EPSC/IPSC, PPF, STDP | 2024 | [Wiley](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/advs.202308588) |
| Si:HfO2 FeFET (28nm) | Gradual modulation via pulse width | 2025 | [Nano Convergence](https://nanoconvergencejournal.springeropen.com/articles/10.1186/s40580-025-00520-2) |
| Electrochemical ionic | Non-volatile STDP | 2025 | [Wiley](https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202418484) |

### Key Mechanisms

1. **Partial Polarization Switching:** Gradual domain nucleation enables analog weight updates
2. **Pulse Engineering:** Both amplitude and width modulation control Δw
3. **Three-Terminal Design:** FeFET separates read/write paths for non-destructive readout

---

## 2. Spiking Neural Networks on Ferroelectric Hardware

### All-Ferroelectric SNN Architecture

| Component | Implementation | Performance |
|-----------|----------------|-------------|
| Synapse | FeFET/FTJ crossbar | Multi-level conductance |
| Neuron | DG-MPBTFT LIF | No capacitor needed |
| System Accuracy | MNIST classification | **94.9%** |

**Source:** [All-Ferroelectric SNN via MPB Neurons](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/advs.202407870)

### Energy Efficiency

| Metric | FeFET SNN | CMOS Baseline |
|--------|-----------|---------------|
| Energy per spike | **0.36 nJ** | ~1 nJ |
| Power | 1/3 of CMOS | Baseline |
| Process | 45nm FinFET | 45nm |

**Source:** [Swarm Optimization FeFET SNN](https://pmc.ncbi.nlm.nih.gov/articles/PMC6700359/)

### HfZrOx-Based SNN Components

| Component | Material | Function |
|-----------|----------|----------|
| Synapse | Ferroelectric HfZrOx | Non-volatile weights |
| Neuron | Antiferroelectric HfZrOx | Integrate-and-fire |
| Endurance | Both | **10⁹ cycles** |

**Source:** [HfZrOx CNN-SNN Computing](https://pubs.acs.org/doi/10.1021/acs.nanolett.5c02889)

---

## 3. Technology Comparison: Ferroelectric vs. Competitors

### Energy Consumption Benchmark

| Technology | SET/RESET Energy | Voltage | Pulse Duration |
|------------|------------------|---------|----------------|
| **FeFET/FTJ** | **10-100 fJ** | 0.1-4 V | 10-100 ns |
| ReRAM | 10-100 fJ | 0.5-3 V | 10-100 ns |
| MRAM (STT) | 100-1000 fJ | 0.5-1.5 V | 1-10 ns |
| PCM | 1000-10000 fJ | 1-3 V | 100-10000 ns |

### Performance Comparison

| Metric | FeRAM/FeFET | ReRAM | PCM | STT-MRAM |
|--------|-------------|-------|-----|----------|
| **Read/Write Speed** | 20-80 ns | 10-100 ns | 50-1000 ns | 1-10 ns |
| **Write Energy** | ~1 fJ | ~10 fJ | ~100 fJ | ~100 fJ |
| **Endurance** | 10¹⁰+ | 10⁶-10¹² | 10⁸-10⁹ | 10¹⁵ |
| **Retention** | 10 years | 10 years | 10 years | 20 years |
| **Scalability** | <10 nm | <10 nm | ~20 nm | ~20 nm |
| **CMOS Compat.** | **Native** | Good | Good | Moderate |

### Intel VLSI 2024 Results (Hafnia-Based)

| Metric | Value |
|--------|-------|
| Write voltage (bit line) | ≤1.0 V |
| Write voltage (plate line) | ≤1.3 V |
| Read voltage | ≤0.6 V |
| Retention | 10 years (elevated T) |
| Write/Read energy | **<100 fJ** (lowest reported) |

**Source:** [MRS Communications](https://link.springer.com/article/10.1557/s43579-024-00660-2)

### Market Outlook (2025-2030)

| Technology | Status | CAGR |
|------------|--------|------|
| Neuromorphic Memory | Emerging | **35%+** |
| ReRAM | Commercialized | 25% |
| STT-MRAM | Production | 20% |
| PCM (3D XPoint) | Intel/Micron | 15% |
| **FeFET** | Pre-production | Highest potential |

---

## 4. On-Chip Learning Techniques

### Direct Feedback Alignment (DFA) on FeFET

Unlike backpropagation which requires long-range weight transport, DFA uses random feedback:

```
δ_l = B_l × δ_output    (B_l is random, fixed matrix)
```

**Advantages:**
- No weight transport problem
- Pipelined layer updates
- Suitable for FeFET implementation

**Source:** [DFA Training with FeFET](https://link.springer.com/chapter/10.1007/978-3-031-19568-6_11)

### Progressive Gradient Descent

**Key Innovation:** Updates each layer progressively without storing full gradient chain

```
For layer l:
  1. Forward pass through layer l
  2. Compute local gradient
  3. Update weights immediately
  4. Propagate to next layer
```

**Result:** Identical learning to conventional backprop but with lower memory requirements.

**Source:** [Science Advances (2024)](https://www.science.org/doi/10.1126/sciadv.ado8999)

### Hybrid Precision Synapse (HPS) for Training

| Precision | Component | Purpose |
|-----------|-----------|---------|
| High (FP32) | Accumulator | Gradient accumulation |
| Low (INT4) | FeFET | Weight storage |
| Mixed | System | Balance accuracy/efficiency |

**Key Components:**
- FeFET-based RNG for stochastic gradient descent
- Ultra-low-power FeFET ADC
- Pipelined DNN training architecture

**Source:** [ACM JETC](https://dl.acm.org/doi/10.1145/3473461)

### BEOL FeFET for Online Training

**Characteristics:**
- Back-end-of-line compatible
- Flexible substrate compatible
- Accurate online training demonstrated

**Achievement:** Enables DNN accelerators with on-chip adaptation.

**Source:** [Wiley AISY](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202300391)

---

## 5. Landau Coefficients for HfO₂/ZrO₂ Simulation

### Universal Curie Constant (Experimental)

| Parameter | Value | Unit |
|-----------|-------|------|
| Curie Constant C | (5.8 ± 0.46) × 10⁻⁷ | K·C·V⁻¹·m⁻¹ |
| **α₀** | **(1.72 ± 0.14) × 10⁶** | V·m·K⁻¹·C⁻¹ |

**Note:** Universal across dopant types and concentrations.

**Source:** [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S2211285520302901)

### Landau-Devonshire Free Energy

```
F = α(T-T_c)P² + βP⁴ + γP⁶ - EP + κ|∇P|²
```

| Coefficient | Typical Value | Notes |
|-------------|---------------|-------|
| α₀ | 1.72 × 10⁶ V·m·K⁻¹·C⁻¹ | Temperature coefficient |
| T_c | ~450-600 K | Curie temperature |
| β | Negative | First-order transition |
| γ | Positive | Stability requirement |
| κ | ~10⁻¹⁰ J·m³/C² | Domain wall width |

### Phase-Field Parameters for HZO

| Parameter | Symbol | Value Range |
|-----------|--------|-------------|
| Spontaneous polarization | P_s | 15-30 μC/cm² |
| Coercive field | E_c | 0.8-2.0 MV/cm |
| Dielectric constant | ε_r | 25-35 |
| Domain wall width | δ | 1-3 nm |
| Kinetic coefficient | L | 10⁻⁶-10⁻⁴ m³/(J·s) |

### Computational Approaches

1. **First-Principles:** DFT calculations for coefficient extraction
2. **Machine Learning:** On-demand Landau potential construction
3. **Fitting:** Experimental P-E loops to extract parameters

**Source:** [arXiv (ML Landau)](https://arxiv.org/html/2512.16207)

---

## 6. Tour Lab / external research institution Publications Summary

### Jaeho Shin Profile

| Attribute | Detail |
|-----------|--------|
| **Affiliation** | external research institution |
| **Experience** | 8+ years semiconductor device fabrication |
| **Focus Areas** | 2D electronics, Flash Joule Heating, Neuromorphic devices |

### Key Publications (2024-2025)

| Paper | Journal | Year | Topic |
|-------|---------|------|-------|
| In2Se3 by FWF for Neuromorphic | Adv. Electron. Mater. | 2025 | 87% MNIST |
| Flash-within-flash synthesis | Nature Chemistry | 2024 | Gram-scale materials |
| Kilogram research sample with arc welder | ACS Nano | 2024 | Scalable synthesis |

### Research Group Connections

| Lab | PI | Connection |
|-----|----| -----------|
| Tour Lab | James M. Tour | IronLattice PI |
| Han Lab | Yimo Han | 2D materials characterization |

**Source:** [Han Lab Publications](https://hanlab.blogs.rice.edu/publications/)

---

## 7. Additional Papers for Download Plan

### STDP & SNN Papers

| Paper | URL | Year |
|-------|-----|------|
| All-Ferroelectric SNN MPB | https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/advs.202407870 | 2024 |
| FeFET Swarm Optimization | https://pmc.ncbi.nlm.nih.gov/articles/PMC6700359/ | 2019 |
| HfZrOx CNN-SNN | https://pubs.acs.org/doi/10.1021/acs.nanolett.5c02889 | 2025 |
| Ambipolar WSe2 FeFET SNN | https://pubs.acs.org/doi/10.1021/acsnano.4c11081 | 2024 |
| Personalized SNN for EEG | https://arxiv.org/html/2601.00020 | 2025 |

### On-Chip Learning Papers

| Paper | URL | Year |
|-------|-----|------|
| DFA Training on FeFET | https://link.springer.com/chapter/10.1007/978-3-031-19568-6_11 | 2022 |
| Progressive Gradient Descent | https://www.science.org/doi/10.1126/sciadv.ado8999 | 2024 |
| Hybrid Precision Synapse | https://dl.acm.org/doi/10.1145/3473461 | 2021 |
| Reservoir Computing FeFET | https://www.nature.com/articles/s44172-022-00021-8 | 2022 |

### Technology Comparison Papers

| Paper | URL | Year |
|-------|-----|------|
| RRAM Review (Chem Rev) | https://pubs.acs.org/doi/10.1021/acs.chemrev.4c00845 | 2025 |
| Emerging NVM Industry | https://link.springer.com/article/10.1557/s43579-024-00660-2 | 2024 |
| IoMT Memory Devices | https://www.cell.com/cell-reports-physical-science/fulltext/S2666-3864(25)00334-0 | 2025 |
| Memristive to Neuromorphic Chips | https://spj.science.org/doi/10.34133/adi.0044 | 2024 |

### Simulation & Modeling Papers

| Paper | URL | Year |
|-------|-----|------|
| HfO2 Computational Progress | https://www.nature.com/articles/s41524-024-01352-0 | 2024 |
| ML Landau Potential | https://arxiv.org/html/2512.16207 | 2024 |
| Universal Curie Constant | https://www.sciencedirect.com/science/article/abs/pii/S2211285520302901 | 2020 |
| HZO Switching Dynamics | https://www.sciencedirect.com/science/article/abs/pii/S1567173919300458 | 2019 |

---

## 8. Key Takeaways for Demo Implementation

### Demo 1: Hysteresis Visualizer

**Landau Parameters to Use:**
```go
params := &LandauParams{
    Alpha0: 1.72e6,      // V·m·K⁻¹·C⁻¹
    Tc:     500,         // K (adjust for material)
    T:      300,         // K (room temperature)
    Beta:   -2.5e9,      // First-order transition
    Gamma:  1.5e11,      // Stability
    Ps:     25e-2,       // C/m² (25 μC/cm²)
    Ec:     1.0e8,       // V/m (1 MV/cm)
}
```

### Demo 2: Crossbar MVM

**Non-Ideality Parameters:**
```go
nonIdealities := &CrossbarNonIdealities{
    WireResistance:  1.0,    // Ω per segment
    SneakRatio:      1000,   // On/Off ratio
    D2DVariation:    0.05,   // 5% device variation
    ReadNoise:       0.02,   // 2% read noise
}
```

### Demo 3: MNIST Inference

**Training Configuration:**
```go
trainingConfig := &TrainingConfig{
    WeightBits:      4,
    NoiseInjection:  0.05,   // 5% Gaussian
    LearningRate:    0.01,
    UseHWATraining:  true,   // Hardware-aware
    TargetAccuracy:  0.87,   // Match IronLattice claims
}
```

---

## References

1. 2D SnS2 FeFET STDP: https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/advs.202308588
2. All-Ferroelectric SNN: https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/advs.202407870
3. RRAM Chemistry Review: https://pubs.acs.org/doi/10.1021/acs.chemrev.4c00845
4. NVM Industry Progress: https://link.springer.com/article/10.1557/s43579-024-00660-2
5. DFA FeFET Training: https://link.springer.com/chapter/10.1007/978-3-031-19568-6_11
6. Progressive Backprop: https://www.science.org/doi/10.1126/sciadv.ado8999
7. HfO2 Computational Progress: https://www.nature.com/articles/s41524-024-01352-0
8. Universal Curie Constant: https://www.sciencedirect.com/science/article/abs/pii/S2211285520302901

---

*Last updated: January 2026 - Part 2*
