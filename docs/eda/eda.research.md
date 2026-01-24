# EDA Research Collection for FeCIM Project

**A Reference Collection of 350+ Papers and Resources**

*Last Updated: January 2026 (Updated with 2024-2025 breakthrough papers)*

---

## Overview

This document collects and references 310+ papers, tools, and resources gathered during the FeCIM Visualizer project. It organizes findings by topic and provides links to original sources for those interested in FeCIM/CIM research.

**Important Disclaimer:** This is a literature collection, not original research. The papers and tools listed here are the work of their respective authors and institutions. We have not validated claims made in these papers.

### Key Findings

1. **Open-source EDA is production-ready for digital CMOS** but requires custom integration for FeFET/CIM designs
2. **The FeFET modeling gap is closing** with OpenVAF and Verilog-A compact models enabling SPICE simulation
3. **Architecture-level tools (NeuroSim, CiMLoop) are mature** and can guide design decisions before circuit implementation
4. **IHP's open PDK with RRAM support** provides the closest path to fabricating CIM arrays in the open ecosystem
5. **30-level quantization** used in this project aligns with state-of-the-art MLC FeFET demonstrations

---

## 1. Paper Corpus Overview

### 1.1 Distribution by Topic

| Category | Papers | Key Sources |
|----------|--------|-------------|
| FeFET/HfO₂ Materials | 45+ | Nature, Adv. Materials, J. Applied Physics |
| CIM Architecture | 40+ | ISSCC, VLSI, IEEE JSSC |
| Simulation Tools | 35+ | arXiv, GitHub, Academic |
| EDA/RTL-to-GDSII | 30+ | OpenROAD, OpenLane, WOSET |
| Neuromorphic/SNN | 30+ | Frontiers, Nature Communications |
| ADC/DAC Design | 25+ | ICCAD, DAC, IEEE |
| Non-Idealities | 25+ | Science China, ResearchGate |
| 3D Integration | 20+ | Nature, Science Advances |
| Transformers/LLM | 15+ | arXiv, Nature Computational Science |
| Photonic CIM | 15+ | MIT, Lightmatter, IEEE |
| Security/PUF | 10+ | IEEE, ACS |

### 1.2 Publication Timeline

```
2020-2021: Foundation papers (NeuroSim validation, OpenLane)
2022-2023: CIM accelerator demonstrations, FeFET modeling advances
2024:      LLM/Transformer CIM, 3D integration, mature tools
2025:      Production deployments, IHP shuttles, analog attention
2026:      Full-stack systems, commercial viability
```

---

## 2. Critical Findings by Research Area

### 2.1 FeFET Device Physics

**Consensus Points:**
- HfO₂-ZrO₂ superlattices achieve >10¹² cycle endurance (vs. 10⁴-10⁵ for standard HfO₂)
- Coercive field Ec ≈ 1 MV/cm; Remanent polarization Pr ≈ 25 µC/cm²
- 30 discrete analog states demonstrated in MLC FeFET (aligned with project's quantization)
- Temperature effects require Verilog-A models with history tracking

**Key Papers:**
| Paper | Finding | Relevance |
|-------|---------|-----------|
| HfO₂-ZrO₂ Superlattice (Shin, Tour) | 10¹² endurance | Direct project foundation |
| University of Oulu Thesis (2025) | Cadence Verilog-A FeCap model | Simulation methodology |
| Sub-3nm FeFET FinFET (2024) | 85.2% power reduction | Future node scaling |

**Gap Identified:** No open PDK includes native FeFET devices. Custom Verilog-A models required.

### 2.2 CIM Architecture and Simulation

**Consensus Points:**
- Matrix-Vector Multiply (MVM) in O(1) via Kirchhoff's law is the core operation
- IR-drop and sneak paths are dominant non-idealities (up to 20% accuracy loss)
- ADC resolution is the primary energy bottleneck (6-8 bits typical)
- Architecture exploration tools (CiMLoop) are 1000× faster than SPICE

**Tool Comparison:**

| Tool | Level | Speed | Accuracy | Open Source |
|------|-------|-------|----------|-------------|
| NeuroSim | Circuit macro | Medium | <5% vs silicon | Yes |
| CiMLoop | System | Fast | Statistical | Yes |
| CrossSim | Array | Fast (GPU) | Good | Yes |
| FAST | Array | Fast | Good | Partial |
| ngspice | Transistor | Slow | High | Yes |
| HSPICE | Transistor | Slow | High | No |

**CrossSim** (Sandia National Labs): GPU-accelerated Python crossbar simulator for neural networks. Models device/circuit non-idealities including programming errors, conductance drift, and ADC precision loss. [View](https://cross-sim.sandia.gov/)

**Recommended Workflow:**
```
1. CiMLoop → Architecture exploration (minutes)
2. NeuroSim → Energy/area estimation (hours)
3. ngspice + Verilog-A → Cell verification (hours)
4. Full array → Only for critical paths
```

### 2.3 Open-Source EDA Readiness

**Production-Ready:**
- Yosys (synthesis) - Tape-out proven
- OpenROAD (P&R) - Used by Infineon, Intel
- Magic VLSI (DRC/LVS) - SKY130 sign-off tool
- GDSFactory (layout scripting) - v9.31 with KLayout backend
- ngspice + OpenVAF - Verilog-A FeFET simulation

**Gaps for FeCIM:**
| Gap | Status | Workaround |
|-----|--------|------------|
| FeFET in PDK | Not available | Custom Verilog-A + ngspice OSDI |
| Crossbar compiler | None open-source | Module 6 attempts this (educational) |
| Large array simulation | O(N²) SPICE | NeuroSim/CiMLoop for estimation |
| Automated CIM layout | Manual/scripted | GDSFactory + KLayout Python |

### 2.4 ADC/DAC Design for CIM

**Key Findings:**
- 4-6 bit ADC sufficient for most neural networks (vs 8-bit overkill)
- ADC consumes 30-60% of total CIM energy
- Current-mode SAR ADCs eliminate TIA (3mW @ 100 MSps)
- ADC-less architectures emerging (HCiM achieves 28× energy reduction)

**Design Space:**

| ADC Type | Resolution | Energy | Speed | CIM Suitability |
|----------|------------|--------|-------|-----------------|
| Flash | 4-6 bit | High | Fast | Good for inference |
| SAR | 6-8 bit | Medium | Medium | General purpose |
| VCO-based | 5-6 bit | Low | Medium | Energy-efficient |
| ADC-less | N/A | Lowest | Varies | Emerging |

**Implication for Project:** Demo 4 (Circuits) should model 5-6 bit ADC as baseline.

### 2.5 Non-Ideality Compensation

**Ranked by Impact:**
1. **IR-Drop** (5-20% accuracy loss) - Mitigated by CAFM scheme, activation modulation
2. **Device Variation** (3-10% loss) - Mitigated by noise-aware training (AIHWKIT)
3. **Sneak Paths** (2-8% loss) - Mitigated by 1T1R/self-rectifying cells
4. **Drift** (1-5% loss) - Mitigated by periodic refresh

**Best Practices from Literature:**
```go
// Pseudo-code for non-ideality-aware inference
result := idealMVM(weights, inputs)
result = applyIRDropCompensation(result, arraySize)
result = applyVariationNoise(result, sigma=0.05)
return quantizeADC(result, bits=6)
```

### 2.6 Transformer/LLM on CIM

**Breakthrough Finding (2025):**
- Analog IMC attention mechanism achieves **70,000× energy reduction** vs digital
- **100× speed-up** compared to GPU for attention computation
- Key insight: Attention's MatMul is naturally suited to crossbar MVM

**Architectural Implications:**
| Component | CIM Suitability | Challenge |
|-----------|-----------------|-----------|
| QKV projection | Excellent | Weight precision |
| Attention MatMul | Excellent | Dynamic range |
| FFN layers | Excellent | Standard MVM |
| Softmax | Poor | Non-linear |
| LayerNorm | Poor | Division required |

**Implication for Project:** Future Demo could show attention mechanism on FeCIM.

### 2.7 Fabrication Pathways

**Open Shuttle Options (2026):**

| Shuttle | Process | FeCIM Suitability | Cost |
|---------|---------|-------------------|------|
| Tiny Tapeout IHP | 130nm BiCMOS | High (RRAM support) | ~$100 |
| Tiny Tapeout SKY130 | 130nm CMOS | Medium (HV modules) | ~$100 |
| IHP Direct | 130nm SG13S | Highest (memristor PDK) | Research |
| GF180MCU | 180nm | Medium (10V HV) | Via Efabless |

**Recommended Path:**
1. Simulate with ngspice + custom FeFET model
2. Layout with GDSFactory/KLayout
3. Tape out peripheral circuits on IHP (CMOS)
4. FeFET array: Research fab partnership or post-processing

---

## 3. Synthesis: State-of-the-Art Benchmarks

### 3.1 MNIST Accuracy on CIM Hardware

| Implementation | Accuracy | Array Size | Technology |
|----------------|----------|------------|------------|
| Ferroelectric memristor RC (2025) | 98.78% | 32×32 | HfO₂ |
| Multi-level FeFET crossbar (2023) | 96.6% | 64×64 | 28nm |
| FTJ crossbar (2024) | 92% | 128×128 | HfO₂ |
| Tour Lab In₂Se₃ (2024) | 87% | Research | Flash synthesis |

**Note:** Our Demo 3 simulation uses 30-level quantization but has **not been validated against real hardware**. The Tour Lab result is from actual physical devices.

### 3.2 Energy Efficiency Comparison

| Technology | TOPS/W | Source |
|------------|--------|--------|
| GPU (A100) | 0.3 | NVIDIA |
| TPU v4 | 1.5 | Google |
| Digital ASIC | 5-10 | Various |
| SRAM CIM | 10-50 | ISSCC 2024 |
| RRAM CIM | 50-150 | Nature 2024 |
| FeFET CIM | 100-500 | Projected |
| Analog Attention | 1000+ | Nature Comp Sci 2025 |

### 3.3 Endurance vs. Retention Trade-off

| Device | Endurance (cycles) | Retention | MLC States |
|--------|-------------------|-----------|------------|
| NAND Flash | 10³-10⁴ | 10 years | 4 (TLC) |
| RRAM | 10⁶-10¹² | 10 years | 2-4 |
| PCM | 10⁸-10⁹ | 10 years | 2-4 |
| FeFET (standard) | 10⁴-10⁵ | 10 years | 4-8 |
| FeFET (superlattice) | 10¹⁰-10¹² | 10 years | 30 |
| STT-MRAM | 10¹⁵ | 10 years | 2 |

---

## 4. Recommendations for FeCIM Project

### 4.1 Immediate Actions

1. **Integrate CiMLoop export in Module 6**
   - Export YAML configuration for architecture exploration
   - Validate energy/area estimates before circuit design

2. **Enhance SPICE export with realistic models**
   - Include Verilog-A FeFET model (Preisach-based)
   - Add IR-drop network for large arrays
   - Reference: University of Oulu thesis methodology

3. **Add ADC-aware quantization to Demo 3**
   - Implement 5-6 bit ADC noise model
   - Show accuracy vs. ADC resolution trade-off

### 4.2 Medium-Term Enhancements

1. **Implement IR-drop visualization in Demo 2**
   - Show voltage drop across array
   - Demonstrate CAFM compensation scheme

2. **Add attention mechanism demo**
   - Visualize QKV projection on crossbar
   - Show energy comparison vs. digital

3. **GDSFactory layout generation**
   - Procedurally generate crossbar arrays
   - Export as OpenLane macro

### 4.3 Research Directions

1. **On-chip learning simulation**
   - Implement hardware backprop (per Science Advances 2024)
   - Show weight update dynamics

2. **3D integration modeling**
   - Simulate multi-layer crossbar stacks
   - Model thermal effects

3. **Security/PUF applications**
   - Leverage FeFET stochasticity for PUF
   - Demonstrate hardware key generation

---

## 5. Tool Integration Roadmap

### 5.1 Current State (Module 6)

```
Neural Network Weights
        ↓
Module 6 Compiler (30-level quantization)
        ↓
Cell Assignments (row, col, G, V_prog)
        ↓
Export: JSON / CSV / SPICE
```

### 5.2 Enhanced Flow (Proposed)

```
Neural Network (PyTorch/ONNX)
        ↓
CiMLoop (Architecture exploration)
        ↓
Module 6 Compiler (Enhanced)
   ├── YAML → CiMLoop validation
   ├── SPICE → ngspice + OpenVAF (.osdi)
   ├── Python → GDSFactory layout
   └── JSON/CSV → Documentation
        ↓
GDSFactory
        ↓
OpenLane (as macro)
        ↓
IHP Shuttle → Silicon
```

### 5.3 Required Tool Versions

| Tool | Minimum Version | Purpose |
|------|-----------------|---------|
| ngspice | 43+ | OSDI support |
| OpenVAF | 24.6.0+ | OSDI 0.4 API |
| GDSFactory | 9.31+ | KLayout backend |
| OpenLane | 2.x | Python-based flow |
| CiMLoop | Latest | MIT repo |
| NeuroSim | V2.1 | PyTorch interface |

---

## 6. Key Research Groups to Follow

| Institution | Focus | Key Contacts |
|-------------|-------|--------------|
| **external research institution (Tour Lab)** | HfO₂-ZrO₂ superlattice, FeCIM company | Dr. external research group, Dr. Jaeho Shin |
| **Georgia Tech** | NeuroSim, CIM benchmarking | Prof. Shimeng Yu |
| **MIT** | CiMLoop, photonic computing | Prof. Murmann |
| **IBM Research** | PCM AIMC, AIHWKIT | Zurich/Almaden labs |
| **ETH Zurich** | Open-source RISC-V, IHP shuttles | PULP team |
| **Stanford** | Memristor arrays | Prof. H.-S. Philip Wong |
| **IMEC** | FeFET process development | Leuven, Belgium |

---

## 7. Bibliography (Categorized)

### 7.1 Core FeFET/HfO₂ Papers

1. Shin et al., "Atomic-scale ferroic HfO2-ZrO2 superlattice gate stack," 2021
2. Shin, Tour et al., "BEOL-Compatible Superlattice FEFET Analog Synapse," IEEE 2022
3. "HfO₂-based FeFETs: From materials to applications," J. Applied Physics, 2025
4. "Ferroelectric Hafnium Oxide: A Potential Game-Changer," Adv. Electronic Materials, 2025
5. "Progress in computational understanding of HfO₂ ferroelectrics," npj Comp. Materials, 2024

### 7.2 CIM Architecture

6. "Memory Is All You Need: CIM Architectures for LLM Inference," arXiv 2024
7. "A compute-in-memory chip based on RRAM," Nature 2022
8. "Architecture and Programming of Analog IMC Accelerators," IBM/IPDPS 2024
9. "Fast and robust analog in-memory DNN training," Nature Communications 2024
10. "CINM (Cinnamon): Compilation Infrastructure for CIM," ACM ASPLOS 2024

### 7.3 Simulation Tools

11. "NeuroSim V1.5: Improved Software Backbone," arXiv 2025
12. "CiMLoop: A Flexible, Accurate, and Fast CIM Modeling Tool," MIT 2024
13. "FerroX: GPU-accelerated phase-field simulator," Comp. Phys. Comm. 2023
14. "IBM AIHWKIT: Analog AI Hardware Kit," APL Machine Learning 2023
15. "COMPASS Compiler Framework for CIM," arXiv 2025

### 7.4 Open-Source EDA

16. "OpenLANE: The Open-Source Digital ASIC Implementation Flow," WOSET 2020
17. "Empowering innovation: OpenROAD and the future of open-source EDA," 2024
18. "GDSFactory Documentation v9.31.0," 2025
19. "Basilisk: End-to-End Open-Source RISC-V SoC in IHP," arXiv 2024
20. "IHP Open Source PDK Documentation," 2025

### 7.5 ADC/DAC and Readout

21. "HCiM: ADC-Less Hybrid Analog-Digital CIM Accelerator," ASP-DAC 2025
22. "Current-Mode SAR ADC for Memristor Readout," MEMRISYS 2024
23. "Memristor-based adaptive ADC for CIM," Nature Communications 2025
24. "Readout Circuit Design for RRAM Array-Based CIM," MDPI Electronics 2024
25. "Review of SRAM-based Compute-in-Memory Circuits," arXiv 2024

### 7.6 Non-Idealities

26. "Optimizing hardware-software co-design for memristor crossbars," Science China 2025
27. "Hardware implementation of memristor-based ANNs," Nature Communications 2024
28. "Sneak path solutions review," RSC Nanoscale Advances 2020
29. "Hardware-Aware Quantization for Accurate Memristor NNs," ICCAD 2024
30. "Model quantization for computing-in-memory: a survey," Science China 2025

### 7.7 Transformers/LLM on CIM

31. "Analog IMC attention mechanism for fast LLMs," Nature Computational Science 2025
32. "Efficient memristor accelerator for transformer self-attention," Scientific Reports 2024
33. "ALBERT on FeFET," Nature Communications 2025
34. "Analog and Digital Hybrid Attention Accelerator," arXiv/IEEE 2024
35. "HARDSEA: Hybrid Analog-ReRAM Digital-SRAM," IEEE TVLSI 2024

### 7.8 Neuromorphic/SNN

36. "Memristor-Based Spiking Neuromorphic Systems," MDPI Nanomaterials 2025
37. "Fully memristive SNN for energy-efficient graph learning," Science 2025
38. "The road to commercial success for neuromorphic technologies," Nature Communications 2025
39. "Roadmap to Neuromorphic Computing with Emerging Technologies," arXiv 2024
40. "Enabling Efficient Processing of SNNs with On-Chip Learning," arXiv 2025

### 7.9 3D Integration

41. "M3D-LIME: Monolithic 3D integration of RRAM hybrid memory," Nature Communications 2023
42. "Eq-CIM: Monolithic 3D IGZO-RRAM-SRAM architecture," Science China 2025
43. "M3D-MP4: Multi-Layer CNT-CMOS/RRAM Mixed-Precision CIM," IEEE 2025
44. "3D Stacked IGZO 2T0C DRAM for CIM," Science Advances 2025
45. "Monolithic 3D integration for energy-efficient computing," ScienceDirect 2024

### 7.10 Photonic CIM

46. "MIT Photonic Processor for Ultrafast AI," Nature Photonics 2024
47. "MAFT-ONN: Photonic Processor for 6G," MIT News 2025
48. "Lightmatter Photonic AI Processor," 2025
49. "Photonics for Neuromorphic Computing: Fundamentals," Advanced Materials 2025
50. "Optical computing accelerators: Principle and perspective," Frontiers of Physics 2025

---

## 8. Glossary

| Term | Definition |
|------|------------|
| **FeFET** | Ferroelectric Field-Effect Transistor |
| **CIM** | Compute-in-Memory |
| **MVM** | Matrix-Vector Multiply |
| **HZO** | Hafnium Zirconium Oxide (HfₓZr₁₋ₓO₂) |
| **OSDI** | Open Source Device Interface |
| **PDK** | Process Design Kit |
| **GDSII** | Graphic Data System II (layout format) |
| **LVS** | Layout Versus Schematic |
| **DRC** | Design Rule Check |
| **SAR** | Successive Approximation Register (ADC type) |
| **TIA** | Transimpedance Amplifier |
| **CAFM** | Crossbar Aware Forward Mapping |
| **BEOL** | Back-End-Of-Line (metal layers) |

---

## 9. Appendix: Paper Collection Locations

| Topic | Location in Project |
|-------|---------------------|
| Simulation Tools PDFs | `docs/papers/by-topic/03-simulation-tools/` |
| Paper URLs | `docs/papers/DOWNLOAD_PLAN.md` |
| Papers Needing Manual Download | `docs/papers/by-topic/PAPERS_NEEDED.md` |
| EDA Tool Documentation | `docs/eda/eda.opensource.md` |
| Research Findings | `docs/project/05-research/findings/` |

---

## 10. Latest Research Update (2024-2026)

*This section captures breakthrough research published since the initial collection.*

### 10.1 Transformer/LLM Acceleration on CIM (Major Breakthrough)

**The Most Significant Finding (2025):**

A paper in *Nature Computational Science* (September 2025) demonstrated analog in-memory computing attention achieving:
- **70,000× energy reduction** compared to GPUs
- **100× speed-up** for attention computation
- Successfully trained a **1.5 billion parameter** model

**Key Technical Details:**
- Uses gain-cell memories (CMOS-compatible, easy to write)
- Sliding window attention to bound physical memory size
- Initialization algorithm achieving GPT-2-level performance without retraining

| Reference | Finding |
|-----------|---------|
| [Leroux et al., Nature Comp Sci 2025](https://www.nature.com/articles/s43588-025-00854-1) | 70,000× energy, 100× speed |

### 10.2 Ferroelectric CIM Annealer Architectures

**DAC 2025 Breakthrough:**
Device-Algorithm Co-Design of FeCIM In-Situ Annealer:
- **1503-1716× energy reduction** vs. state-of-the-art annealers
- **98% success rate** for 3000-node Max-Cut problems
- VMV (vector-matrix-vector) acceleration on FeFET arrays

**Nature Communications 2024:**
- 75% chip size saving via lossless QUBO matrix compression
- FeFET three-terminal structure enables matrix compression

### 10.3 HfO₂-ZrO₂ Superlattice Stability Advances

**Key 2025 Findings:**

| Paper | Advancement |
|-------|-------------|
| Nature Comms 2025 | Stable ferroelectricity up to 100nm thickness |
| Nature Comms 2025 | >10⁹ switching cycles (fatigue resistance) |
| Nature Comms 2025 | Low coercive field ~0.85 MV/cm |
| Nature 2022 | Scaled to ~20 Å on Si transistors |

**Implications for Project:**
- Validates use of HfO₂-ZrO₂ superlattice in simulations
- Supports 10¹²+ cycle target for FeCIM devices
- Lower Ec enables lower programming voltages

### 10.4 FeFET Compact Modeling Advances

**2024 Verilog-A Model Improvements:**

| Feature | Status | Paper |
|---------|--------|-------|
| Temperature effects | Implemented | Solid-State Electronics 2024 |
| Device variability | Implemented | Solid-State Electronics 2024 |
| Preisach hysteresis | Standard | Multiple papers |
| MLC/MAC validation | Validated | Solid-State Electronics 2024 |

**Model Availability:**
- OpenVAF-compatible Verilog-A models now available
- Cadence Spectre validation for MVM operations
- BSIM-based FeFET compact models published

### 10.5 Industry Developments

**Ferroelectric Memory Company (FMC):**
- **€100M funding** (November 2025)
- DRAM+ technology based on ferroelectric HfO₂
- Dresden, Germany (spun out from TU Dresden/GlobalFoundries)
- Targets energy-efficient memory for AI

**IHP Open Source PDK Updates:**
- April 2025 tape-out (Testfield T586) completed
- RRAM module available in SG13S (upon request)
- OpenROAD flow fully supported
- Magic VLSI support planned for 2025

**OpenLane 2.0:**
- Released April 2024
- Python-based infrastructure (not just Tcl)
- Fully customizable flows
- Critical for integrating custom CIM macros

### 10.6 2D Ferroelectric Material Synthesis

**Tour Group In₂Se₃ (2025):**
- Flash-within-flash Joule heating (FWF) method
- **Gram-scale** α-In₂Se₃ crystal synthesis
- Robust synaptic behavior demonstrated
- Path to scalable neuromorphic devices

**2D MoS₂ CIM (2024):**
- 96.36% wafer-scale yield
- >10¹² endurance cycles
- >10 years retention
- 263× power efficiency vs. GPU for dynamic tracking

### 10.7 Simulation Tool Updates

| Tool | Update | Key Capability |
|------|--------|----------------|
| **NeuroSim V2.1** | PyTorch interface | <1% chip-level error |
| **CrossSim** | Sandia Labs | GPU-accelerated crossbar sim |
| **3D_NeuroSim** | New release | Monolithic 3D integration |
| **FAST** | Updated | Hardware-software co-design |
| **ngspice 43/44** | OSDI 0.4 | Verilog-A FeFET models |
| **OpenVAF** | Reloaded | Modern compact models |

### 10.8 Updated Research Group Contacts

| Institution | Focus | Contact/Lab |
|-------------|-------|-------------|
| **external research institution** | In₂Se₃ FeFET, Flash synthesis | Tour Lab |
| **Georgia Tech** | NeuroSim, CIM benchmarking | Prof. Shimeng Yu |
| **MIT** | CiMLoop, Analog attention | Prof. Murmann |
| **FZ Jülich** | Analog IMC attention | Leroux team |
| **TU Dresden/FMC** | Ferroelectric DRAM+ | FMC startup |
| **IHP Leibniz** | Open PDK, RRAM | OpenPDK team |

---

## 11. Recommended Reading Order

For newcomers to FeCIM research:

1. **Start Here:** Tour presentation on 30-state FeFET
2. **Device Physics:** Roadmap on ferroelectric hafnia (APL Materials 2023)
3. **CIM Architecture:** CiMLoop paper (MIT 2024)
4. **Breakthrough:** Analog IMC attention (Nature Comp Sci 2025)
5. **Simulation:** NeuroSim validation paper
6. **EDA Tools:** This project's `eda.opensource.md`

---

*This document references 350+ papers and resources collected during the project. For the complete paper collection and detailed references, see the documents listed in Appendix 9. All papers remain the intellectual property of their respective authors and institutions.*
