# IronLattice Research Findings - January 2026

This document compiles research findings on IronLattice ferroelectric CIM technology, Dr. external research group's work, and related curriculum topics.

---

## 1. IronLattice Company Status

### Funding & Recognition
- **One Small Step Grant (2025):** $50,000 from Rice Innovation, Cycle 4
- **Status:** Lab-stage spin-out from external research institution Tour Lab
- **Source:** [Rice News](https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants)

### Key Personnel
| Name | Role | Background |
|------|------|------------|
| **Dr. external research group** | Principal Investigator | T.T. and W.F. Chao Professor of Chemistry, external research institution |
| **Dr. Jaeho Shin** | Device Engineer, Lead | 10+ years semiconductor fabrication, superlattice inventor |
| **Tawfik Jarjour** | Commercialization Lead | Rice alumnus, semiconductor industry experience |

### Performance Claims (COSM 2025)
| Metric | vs. NAND Flash | vs. DRAM |
|--------|---------------|----------|
| Energy | **10⁶× lower** | **10³× lower** |
| Speed | **10⁶× faster** | Comparable |
| Operating Voltage | **90% reduction** | Lower |
| Data Center Energy | **80-90% reduction** | - |

### Technology Specifications
- **Structure:** HfO₂/ZrO₂ superlattice (atomically precise layers)
- **Analog States:** ~30 discrete levels per cell
- **Endurance Target:** >10¹² cycles
- **MNIST Accuracy:** 87% (near theoretical max for device precision)
- **CMOS Compatibility:** Native BEOL integration

---

## 2. Dr. external research group Publications - Ferroelectric/Neuromorphic

### In₂Se₃ for Neuromorphic Computing (2024-2025)
**Title:** "In2Se3 Synthesized by the FWF Method for Neuromorphic Computing"

| Attribute | Detail |
|-----------|--------|
| **Authors** | Jaeho Shin, Jingon Jang, Chi Hun Choi, Jaegyu Kim, Lucas Eddy, Phelecia Scotland, Lane W. Martin, Yimo Han, James M. Tour |
| **Published** | Advanced Electronic Materials, November 2024 |
| **Preprint** | [ChemRxiv](https://chemrxiv.org/engage/chemrxiv/article-details/659ef4cee9ebbb4db9de84cb) |
| **Full Paper** | [Wiley](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aelm.202400603) |

**Key Innovations:**
1. **Flash-Within-Flash (FWF) Synthesis:** Novel Joule heating method for non-conductive precursors
2. **Material:** α-In₂Se₃ crystals with R3m structure, 2D ferroelectric semiconductor
3. **Device:** Ferroelectric semiconductor FET (FS-FET) as artificial synapse
4. **Synaptic Behaviors:** EPSC/IPSC, PPF, STDP demonstrated
5. **MNIST Accuracy:** ~87% in single-layer neural network

**Characterization:**
- XRD: α-In₂Se₃ with R3m crystalline structure
- Raman peaks: 105, 181, 193 cm⁻¹

### Related Tour Lab Publications
| Paper | Year | Topic |
|-------|------|-------|
| Stoichiometric Engineering of Indium Selenide via FWF | 2024 | Flash synthesis |
| Flash Graphene synthesis | 2020+ | research sample foundation |

---

## 3. Ferroelectric CIM Benchmark Data

### MNIST Accuracy Comparison (2024-2025)

| Technology | Array Size | Accuracy | Year | Source |
|------------|-----------|----------|------|--------|
| Ferroelectric memristor (RC) | 24×24 | **98.78%** | 2025 | [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963) |
| RRAM crossbar | 784×128×10 | 98.1% | 2024 | ResearchGate |
| FeFET (1F-1R) | 8×8 | ~97% | 2023 | IEEE |
| Multi-level FeFET | 32×32 | 96.6% | 2023 | [Nature Comms](https://www.nature.com/articles/s41467-023-42110-y) |
| 2D SnS2 FeFET | - | 94% | 2024 | [Wiley](https://advanced.onlinelibrary.wiley.com/doi/10.1002/advs.202308588) |
| FTJ crossbar | - | 92% | 2024 | [SemiEngineering](https://semiengineering.com/ferroelectric-tunnel-junctions-in-crossbar-array-analog-in-memory-compute-accelerators/) |
| **IronLattice (claimed)** | - | **87%** | 2025 | COSM 2025 |

### Key Performance Metrics for Synaptic Devices

| Metric | Requirement | Best Achieved |
|--------|-------------|---------------|
| Conductance states | ≥32 (5-bit) | >128 (vdW FeFET) |
| Gmax/Gmin ratio | >10 | >10⁵ (2D SnS2) |
| Endurance | >10⁶ cycles | 10⁷ cycles (2D SnS2) |
| Linearity (αp/αd) | Close to 1 | 0.45/0.73 (ITO FeFET) |
| Symmetry | Close to 1 | 0.89 (ITO FeFET) |

---

## 4. Superlattice FeFET Technology

### Structure: FE/DE/FE Superlattice
The superlattice approach (ferroelectric/dielectric/ferroelectric layers) provides critical advantages:

| Benefit | Mechanism |
|---------|-----------|
| **Phase Stabilization** | Lattice mismatch locks orthorhombic Pca2₁ phase |
| **Wake-up Elimination** | Interfaces inhibit oxygen vacancy diffusion |
| **Improved Linearity** | Moderate domain switching via layer coupling |
| **Higher Endurance** | Interface defect trapping (>10⁹ cycles) |
| **Lower Eᶜ** | FE-AFE competition reduces switching barrier |

### Key Papers

| Title | Authors | Year | Source |
|-------|---------|------|--------|
| BEOL-Compatible Superlattice FEFET Analog Synapse | Aabrar et al. | 2022 | [IEEE](https://ieeexplore.ieee.org/document/9691825/) |
| A Thousand State Superlattice FeFET Analog Weight Cell | Aabrar et al. | 2022 | IEEE VLSI |
| Self-Rectifying FTJ Synapse (HZH SL) | - | 2024 | [RSC](https://pubs.rsc.org/en/content/articlelanding/2024/mh/d4mh00519h) |
| HfO₂-ZrO₂ Superlattice Improved Endurance | Shin et al. | 2021 | [ResearchGate](https://www.researchgate.net/publication/357111874_HfO2-ZrO2_Superlattice_Ferroelectric_Capacitor_with_Improved_Endurance_Performance_and_Higher_Fatigue_Recovery_Capability) |
| Adaptive Control Epitaxial HfO₂/ZrO₂ | - | 2025 | [Nature Comms](https://www.nature.com/articles/s41467-025-61758-2) |

---

## 5. Simulation Tools & Frameworks

### FerroX - GPU Phase-Field Simulator
**Source:** [GitHub](https://github.com/AMReX-Microelectronics/FerroX) | [arXiv](https://arxiv.org/abs/2210.15668)

| Feature | Description |
|---------|-------------|
| **Framework** | AMReX (massively parallel) |
| **Equations** | TDGL + Poisson + Charge transport |
| **GPU Support** | CUDA, 15× speedup vs CPU |
| **Materials** | HfZrO₂, configurable parameters |
| **Applications** | MFIM, MFIS stacks, domain dynamics |

**Key Capabilities:**
- Time-dependent Ginzburg-Landau (TDGL) equation solver
- 180° ferroelectric domain formation modeling
- Polarization switching characteristics
- Boundary conditions for various interfaces

### IBM AIHWKit - Analog Hardware Simulation
**Source:** [GitHub](https://github.com/IBM/aihwkit) | [Docs](https://aihwkit.readthedocs.io/)

| Feature | Description |
|---------|-------------|
| **Purpose** | Simulate inference/training on analog CIM |
| **Language** | Python with C++/CUDA backend |
| **Layers** | FC, Conv1D/2D/3D, LSTM |
| **Devices** | PCM, ReRAM, ECRAM, Capacitors |

**Training Methods:**
- Hardware-Aware Training (HWA)
- Mixed-precision training
- Tiki-taka optimizer
- Noise injection for robustness

**Noise Models:**
- Output-referred noise
- Device fluctuations
- ADC/DAC discretization
- Weight update stochasticity

### Preisach Model Implementation
**Purpose:** Fast hysteresis simulation for circuit-level analysis

| Implementation | Source |
|----------------|--------|
| Verilog-A model | [GitHub](https://github.com/DavidTobar456/pfecapRevision) |
| Spectre integration | Bartic et al., J. Appl. Phys. 89(6), 2001 |
| SPICE compatible | Uses tanh approximation |

**Parameters:**
- Up-switch field (U), Down-switch field (D)
- Distribution functions for hysteron ensemble
- 5 independent parameters for weight function

---

## 6. Curriculum Topics Summary

### Physics Foundation
1. **HfO₂ Crystallography:** Monoclinic → Orthorhombic (Pca2₁) phase transition
2. **Ferroelectricity:** Spontaneous polarization, switchable dipoles
3. **Domain Dynamics:** Nucleation-limited switching (NLS), domain wall motion
4. **Hysteresis:** P-E curves, coercive field (Eᶜ), remanent polarization (Pᵣ)

### Device Engineering
1. **FeFET Architecture:** Gate stack design, threshold voltage modulation
2. **Superlattice Benefits:** Strain engineering, defect control
3. **BEOL Integration:** Low-temperature processing (<450°C)
4. **Analog States:** Partial polarization switching for multi-level storage

### CIM Architecture
1. **Crossbar Arrays:** Ohm's law (multiplication), Kirchhoff's law (summation)
2. **Matrix-Vector Multiplication:** O(1) complexity, parallel computation
3. **Non-idealities:** Sneak paths, IR drop, device variation
4. **ADC/DAC:** Precision vs. energy tradeoff

### AI/ML Integration
1. **Weight Mapping:** Conductance representation, differential pairs
2. **Noise-Aware Training:** Inject hardware noise during training
3. **Quantization:** INT4/INT8 for reduced ADC requirements
4. **STDP:** Spike-timing dependent plasticity for on-chip learning

### Simulation Methods
1. **Phase-Field (TDGL):** Domain evolution, microstructure
2. **Preisach Model:** Fast hysteresis for circuit simulation
3. **Landau-Khalatnikov:** Switching dynamics
4. **Monte Carlo:** Device variability analysis

---

## 7. Demo Implementation Targets

### Demo 1: Hysteresis Visualizer
- **Status:** Framework in place
- **Physics:** Preisach model for P-E curve
- **Visualization:** Real-time domain coloring, interactive voltage control
- **Goal:** Demonstrate 30 discrete analog states

### Demo 2: Crossbar MVM
- **Status:** Planned
- **Architecture:** 8×8 or 16×16 array
- **Operations:** Vector input, weight storage, current summation
- **Goal:** Show O(1) matrix multiplication

### Demo 3: MNIST Inference
- **Status:** Planned
- **Network:** 784×128×10 (single hidden layer)
- **Target Accuracy:** 87% (matching IronLattice claims)
- **Features:** Noise injection, weight quantization

---

## 8. Key External Resources

### Conference Presentations
- COSM 2025: Dr. Tour on IronLattice breakthrough
- IEEE IEDM 2025: FeFET session (Dec 2025)
- Stanford EE Seminar (Feb 2025): Ferroelectric AI hardware

### Research Groups
| Group | Institution | Focus |
|-------|-------------|-------|
| Tour Lab | external research institution | IronLattice, flash synthesis |
| Salahuddin Group | UC Berkeley | FerroX, NC-FETs |
| Datta Group | Georgia Tech | FeFET reliability |
| IBM Research | Zurich | AIHWKit, PCM arrays |

### Online Resources
- [AIHWKit Documentation](https://aihwkit.readthedocs.io/)
- [FerroX GitHub](https://github.com/AMReX-Microelectronics/FerroX)
- [Vulkan Tutorial - Compute](https://vulkan-tutorial.com/Compute_Shader)
- [go-vk Bindings](https://github.com/bbredesen/go-vk)

---

## References

1. Rice Innovation One Small Step Grants: https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants
2. In2Se3 FWF Neuromorphic: https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aelm.202400603
3. FerroX Phase-Field: https://arxiv.org/abs/2210.15668
4. IBM AIHWKit: https://github.com/IBM/aihwkit
5. Multi-level FeFET Crossbar: https://www.nature.com/articles/s41467-023-42110-y
6. BEOL Superlattice FeFET: https://ieeexplore.ieee.org/document/9691825/
7. Self-Rectifying FTJ: https://pubs.rsc.org/en/content/articlelanding/2024/mh/d4mh00519h
8. 2D SnS2 FeFET: https://advanced.onlinelibrary.wiley.com/doi/10.1002/advs.202308588
9. Ferroelectric Memristor RC: https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963
10. Level1Techs Forum Discussion: https://forum.level1techs.com/t/new-chip-tech-claims-1million-faster-than-nand/243943

---

*Last updated: January 2026*
