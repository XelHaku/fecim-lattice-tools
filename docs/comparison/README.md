# CIM Research Paper Summary

Comprehensive catalog of research sources for Compute-in-Memory (CIM) technology, covering ferroelectric devices, crossbar arrays, neural network inference, and the complete FeCIM design suite.

> **Note:** This is a literature catalog. Values and claims are reported from sources and are **not verified** by this project. See `docs/comparison/HONESTY_AUDIT.md` for the current verification scope.

## Overview

This document organizes 300+ research sources and URLs across 50 priority categories, supporting documentation for:
- **Ferroelectric device physics** and HfO₂/ZrO₂ superlattices
- **Crossbar array architectures** and non-idealities
- **Matrix-vector multiplication** (MVM) and CIM inference
- **Multi-level programming** and analog weight storage
- **Neural network quantization** and MNIST demonstrations
- **EDA tools integration** (OpenLane, OpenROAD, NeuroSim, CrossSim)
- **Emerging applications**: SNNs, transformers/LLMs, CAM, cryogenic, automotive
- **Manufacturing pathways**: 3D stacking, BEOL integration, process control

---

## Priority 1: Foundational Ferroelectric Physics

### HfO₂/ZrO₂ Superlattice Advances

| Title | Source | Year | Key Finding |
|-------|--------|------|------------|
| Atomic-scale ferroic HfO2-ZrO2 superlattice gate stack | external research institution / ResearchGate | 2021 | Enabled ~20 Å scaling on Si transistors |
| Enhancing ferroelectric stability in epitaxial HfO2/ZrO2 superlattices | Nature Communications / PMC | 2025 | >10⁹ cycles, coercive field ~0.85 MV/cm |
| HfO2-ZrO2 Superlattice with Improved Endurance | ResearchGate / PubMed | 2021-2023 | 10¹² cycle capability demonstrated |
| Ferroelectric HfO2–ZrO2 Multilayers with Reduced Wake-Up | ACS Omega | 2024 | Manufacturing process control improvements |
| Ferroelectric materials, devices, and chips | Springer / Sci China Info Sci | 2025 | Comprehensive review covering applications |
| HfO2-ZrO2 for 2D MoS2 NC-Transistors | ACS Applied Nano Materials | 2024 | Negative capacitance integration |

### Ferroelectric Fundamentals

| Title | Source | Year |
|-------|--------|------|
| Ferroelectricity in hafnium oxide thin films | Applied Physics Letters (APL) | 2011 |
| Ferroelectricity in Doped HfO₂ | Advanced Materials | 2015 |
| Ferroelectricity in Simple Binary ZrO2 and HfO2 | Semantic Scholar | 2012 |
| First-principles predictions of HfO2-based ferroelectric | arXiv:2401.05288 | 2024 |
| Progress in computational understanding of ferroelectric mechanisms | npj Computational Materials | 2024 |

---

## Priority 2: Crossbar Physics and Non-Idealities

### Matrix-Vector Multiplication Fundamentals

| Paper | Key Concept | Finding |
|-------|------------|---------|
| In-Memory Computing Deep Learning | MVM throughput scaling | O(n²) scaling; latency 10-100 ns typical |
| Memristor CIM Survey | Energy efficiency | 10-1000× better than GPU for inference |
| Memory Tech Crossbar DNN Accuracy | Technology comparison | FeFET most robust across variations |

### IR Drop Modeling

| Paper | Source | Topic |
|-------|--------|-------|
| Fast IR-Drop Model of Memristor Crossbars | IEEE TCAD 2024 | 10.9% modeling error, 4784× faster than SPICE |
| Accurate Prediction Under I-V Nonlinearity | IEEE TCAD 2022 | Combined IR drop + nonlinearity model |
| Resistive vs Capacitive Parasitics | ResearchGate 2025 | Performance trade-off analysis |

**Literature Findings:**
- 64×64 array: ~5% max IR drop, 0.3% accuracy loss
- 128×128 array: ~12% max IR drop, 1.2% accuracy loss
- 256×256 array: ~22% max IR drop, 3.5% accuracy loss
- Mitigation: voltage compensation, array tiling, thicker metal

### Sneak Path Current

| Paper | Source | Key Contribution |
|-------|--------|------------------|
| Sneak Path Current Modeling Framework | Frontiers Materials / arXiv 2025 | Closed-form framework for 1nm nodes |
| Sneak Path in Self-Rectifying Arrays | Frontiers Materials 2022 | 1T1R vs 1S1R vs passive comparison |
| Research progress on sneak path solutions | RSC 2020 | Comprehensive mitigation strategies |

**Architecture Performance:**
| Array Type | Sneak Error | Mitigation |
|-----------|-------------|-----------|
| Passive (0T1R) | 5-20% | Self-rectifying cells |
| 1T1R | ~0% | Transistor isolates |
| 1S1R (selector) | 0.5-2% | Selector blocks |
| 1D1R (diode) | 1-5% | Diode asymmetry |

### Device Variation and Drift

| Topic | Literature Findings |
|-------|-------------------|
| Cycle-to-Cycle Variation | RRAM: 5-15%, PCM: 8-20%, FeFET: 3-8%, Flash: 2-5% |
| Device-to-Device Variation | RRAM: 10-20%, PCM: 15-30%, FeFET: 5-15%, Flash: 5-10% |
| Conductance Drift Coefficient ν | PCM: 0.05-0.1, RRAM: 0.01-0.05, FeFET: 0.001-0.01 |
| 10-Year Retention | PCM: 90-95%, RRAM: 95-99%, FeFET: 99%+, Flash: 99%+ |

---

## Priority 3: CIM Inference Architectures

### Foundational CIM Papers

| Paper | Authors/Source | Year | Key Metric |
|-------|---|------|----------|
| Crossbar Array of Artificial Synapses Based on Ferroelectric Diodes | Various | 2021 | Synaptic implementation |
| High Linearity and Symmetry Ferroelectric Artificial Neuromorphic Devices | ResearchGate | 2024 | Analog weight precision |
| Negative Feedback Training for NVCIM DNN Accelerators | arXiv | 2023 | Training on CIM arrays |
| Spike Optimization for Ferroelectric Tunnel Junction Synaptic Devices | Preprints.org | 2023 | SNN integration |

### Multi-Level FeFET Programming

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| First Demo Multi-Level FeFET Crossbar | Nature Communications | 2023 | 96.6% MNIST accuracy |
| Dual-Bit FeFET Enhanced Storage | Nature | 2025 | Dual-bit architecture for 8+ states |
| Reliable MLC FeFET Programming | ResearchGate | 2025 | Program/verify methodology |
| 2D SnS2 Analog Synaptic FeFET | Advanced Science | 2024 | >7-bit, 10⁷ cycles |

### FeFET Linearity and Symmetry

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| High Linearity ITO FeFETs | Advanced Electronic Materials | 2025 | αp=0.45, αd=0.73 |
| BEOL Analog FeFET Training | Advanced Intelligent Systems | 2023 | Online DNN training capability |
| Van der Waals FeFET Synapses | Nano Energy | 2023 | 128 states, Gmax/Gmin>120 |
| CMOS FeFET Synaptic Weights | ACS Applied Materials & Interfaces | 2020 | BEOL compatible |

---

## Priority 4: MNIST and Neural Network Inference

### Quantization-Aware Training

| Paper | Key Finding | ADC Impact |
|-------|------------|-----------|
| Quantization Survey (2023) | QAT recovers 90%+ accuracy loss | 4-bit ADC: baseline |
| Low-Bit Quantization (2022) | 4-bit weights sufficient with QAT | 6-bit ADC: 4× power |
| Extreme Partial-Sum Quantization | 2-3 bit ADC resolution possible | 8-bit ADC: 16× power |
| Dynamic Quantization Range Control | Adaptive precision | 10-bit ADC: 64× power |

### MNIST Demonstrations

| Paper | Hardware | Accuracy | Endurance |
|-------|----------|----------|-----------|
| Flash In2Se3 for Neuromorphic Computing | In₂Se₃ FeFET | Reported (unverified) | Gram-scale synthesis |
| FeFET Crossbar MNIST Hardware Demo | 128×64 FeFET array | Reported (unverified) | 10⁹+ cycles (reported) |
| Ferroelectric Memristor RC Arrays | Ferroelectric devices | 98.78% | 2025 publication |
| Multi-Level FeFET Crossbar | Multi-level FeFET | 96.6% | Write-verify capable |
| FeCap/FeFET CIM Elements | Comparison study | Variable | Different topologies |
| CMOS-Compatible FE Synaptic Arrays | CMOS-compatible | CNN acceleration | 2022 paper |
| 2D Ferroelectric Hybrid CIM | 2D van der Waals | 96.36% yield | 10¹² endurance |

---

## Priority 5: ADC/DAC Design

### ADC Power Problem

CIM Power Budget (Typical):
- Array (MVM): 10-30%
- **ADC: 50-80%** (DOMINANT)
- DAC: 5-15%
- Digital logic: 5-10%

### ADC-Less Architectures

| Paper | Source | Year | Approach |
|-------|--------|------|----------|
| HCiM ADC-Less Hybrid CIM | arXiv:2403.13577 | 2024 | Eliminate ADC entirely |
| Pruning ADC Efficiency Crossbar | Various | 2024 | Neural network optimization |
| Memristor-Based Adaptive ADC | Nature 2025 | 2025 | 15.1× energy improvement |
| VCO-Based ADC for ReRAM | ACM 2021 | 2021 | 32 levels, <5.2pJ |

### Weight Mapping

| Paper | Source | Topic |
|-------|--------|-------|
| Extreme Partial-Sum Quantization | ACM 2021 | 2-3 bit ADC resolution |
| CNN Weight Mapping | ACS Applied Materials | Binary/quantized mapping |
| Weight Nonlinear Mapping | Science Direct | Off-chip training |
| Crossbar Architectures for DNN | Springer | Survey of topologies |
| Two-Phase Weight Mapping | ResearchGate | Variation tolerant |

---

## Priority 6: Simulation and EDA Tools

### GPU Phase-Field Simulation

| Tool | Repository | Year | Key Feature |
|------|-----------|------|------------|
| FerroX | GitHub AMReX-Microelectronics | Current | 15× GPU speedup, TDGL solver |
| FerroX Paper | arXiv:2210.15668 | 2022 | Computational methodology |

### Analog Hardware Simulation

| Framework | URL | Description |
|-----------|-----|-------------|
| IBM AIHWKit | github.com/IBM/aihwkit | Analog CIM simulation, noise-aware training |
| AIHWKit Paper | APL Machine Learning | 2023 publication |
| CrossSim 3.1 | github.com/sandialabs/cross-sim | GPU-accelerated, PyTorch integration |
| NeuroSim 1.5/2.1 | GeorgiaTech / GitHub | Device→circuit→chip hierarchy, <1% calibration error |
| 3D_NeuroSim V1.0 | GitHub | Monolithic and heterogeneous 3D support |
| CINM (Cinnamon) | ACM ASPLOS 2024 | LLVM-based compilation infrastructure |
| CiMLoop | MIT arXiv:2405.07259 | Flexible, accurate CIM modeling |

### Preisach Model Implementation

| Resource | URL | Type |
|----------|-----|------|
| Verilog-A PFECAP | github.com/DavidTobar456/pfecapRevision | Circuit simulation |
| Preisach FeCap Paper | J. Applied Physics 2001 | Foundational |
| Physical Reality Preisach | Nature Communications 2018 | Theoretical validation |

### EDA Tools and PDKs

| Tool | Source | Year | Status |
|------|--------|------|--------|
| OpenLane 2.0 | Efabless | 2024 | Python-based, fully customizable |
| OpenROAD | Free/open-source | Current | DRC/LVS compatible |
| GDSFactory | IEEE 2024 | 2024 | GDSII generation framework |
| IHP Open Source PDK | IHP-GmbH | 2025 | April 2025 tape-out, RRAM module |
| SkyWater SKY130 | Open-source | Current | 130nm node, widely available |

---

## Priority 7: On-Chip Training and Learning

### Hardware Backpropagation

| Paper | Source | Year | Method |
|-------|--------|------|--------|
| DFA Training on FeFET | Springer 2022 | 2022 | Direct feedback alignment |
| Progressive Gradient Descent | Science Advances | 2024 | In-situ backprop, 40×40 FeCap array |
| Hybrid Precision Synapse | ACM 2021 | 2021 | Mixed precision training |
| Reservoir Computing FeFET | Nature 2022 | 2022 | Time-series, 98.1% speech, 1000× speedup |
| BEOL Analog FeFET Training | Advanced Intelligent Systems | 2023 | Online DNN training |
| 3D FeFET Array | Nature Communications | 2023 | Fully-integrated learning capability |

---

## Priority 8: Neuromorphic Computing

### Spiking Neural Networks (SNNs)

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| All-Ferroelectric SNN via MPB | Advanced Science | 2024 | 94.9% accuracy |
| FeFET Swarm Optimization Solver | PMC 2019 | 2019 | 0.36 nJ/spike |
| HfZrOx CNN-SNN Computing | ACS Nano Letters | 2025 | 10⁹ endurance |
| Ambipolar WSe2 FeFET SNN | ACS Nano | 2024 | Compact LIF neurons |
| Personalized SNN for EEG | arXiv:2601.00020 | 2025 | Memristive adaptation |
| FeFET Neuromorphic Overview | Frontiers Neuroscience | 2020 | Supervised learning foundations |
| Electrochemical Ionic STDP | Advanced Materials | 2025 | Nonlinear STDP dynamics |

### Reservoir Computing

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| Single FeFET Reservoir | Nature | 2022 | 98.1% speech, 1000× speedup |
| FeFET Reservoir Speech | IEEE 2024 | 2024 | Spoken digit recognition |
| Ferroelectric RC Review | Frontiers Neuroscience | 2023 | Overview and applications |
| Echo State FeFET | arXiv | 2023 | Time-series processing |
| Polycrystalline HZO RC | Advanced Materials | 2024 | Single device implementation |

---

## Priority 9: Advanced CIM Applications

### Content-Addressable Memory (CAM)

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| Combination-Encoding CAM | ACS Applied Electronics Materials | 2025 | Highest density |
| TAP-CAM Tunable Approximate | ACM ICCAD | 2024 | Tunable precision |
| FACAM Analog CAM | ResearchGate | 2024 | 60× energy reduction |
| Cryogenic FeSQUID TCAM | Nature | 2025 | 1.36 aJ/search |
| FeFET CAM Overview | MDPI Nanomaterials | 2022 | Comprehensive review |
| IGZO CAM Retention | ACS Applied Electronics Materials | 2023 | Multi-bit retention |
| Hyperdimensional FeFET | Nature Scientific Reports | 2022 | HD computing applications |

### Transformer and LLM Acceleration

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| Analog IMC Attention Mechanism | Nature Computational Science | 2025 | **70,000× energy reduction**, 100× speed |
| Memristor Self-Attention | Nature Scientific Reports | 2024 | Efficient accelerator design |
| ALBERT on FeFET | Nature Communications | 2025 | Transformer demo on FeFET |
| Semantic Memory CIM+CAM | Science Advances | 2024 | 2D/3D vision processing |
| IMC Transformer Long Sequences | ResearchGate | 2021 | Attention architecture |
| FAMOUS FPGA Attention | arXiv:2409.14023 | 2024 | FPT 2024 presentation |

### Combinatorial Optimization

| Paper | Source | Year | Application |
|-------|--------|------|-------------|
| Device-Algorithm Co-Design FeCIM Annealer | DAC 2025 | 2025 | 1503-1716× energy vs SOTA, Max-Cut |
| C-Nash Nash Equilibrium Solver | DAC 2024 | 2024 | Mixed strategy Nash |
| Ferroelectric CIM Annealer | Nature Communications | 2024 | 75% chip size saving |

### In-Memory Computation

| Paper | Source | Year | Finding |
|-------|--------|------|---------|
| In-Memory Ferroelectric Differentiator | Nature Communications | 2025 | 40×40 FeCap array for differentiation |
| Reconfigurable FE Diodes CIM | ACS Nano Letters | 2023 | Field-programmable architecture |

---

## Priority 10: Manufacturing and Integration

### BEOL Integration

| Paper | Source | Year | Process |
|-------|--------|------|---------|
| BEOL FeFET Laser Process | Small 2025 | 2025 | <500°C fabrication compatible |
| BEOL Superlattice FeFET Synapse | IEEE TED | 2022 | Linearity and symmetry |
| MoOx/TiON Superlattice FeFET | IEEE 2025 | 2025 | Low-variability design |
| HZO SL Endurance/Fatigue | IEEE 2023 | 2023 | Recovery performance |
| SL FeMFET for 3D NAND | Science Direct | 2024 | TLC design for storage |

### 3D Stacking Architecture

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| 3D NAND FeFET | IEEE IEDM 2024 | 2024 | 64-layer demonstration |
| 5-bit MLC 3D FeFET | Nature | 2025 | 32 states at 25nm |
| Vertical FeFET Array | Nature Communications | 2023 | Fully-integrated 3D |
| 3D CIM Architecture | arXiv:2504.09713 | 2025 | Polarization sensing |
| 3D FeRAM Stack | IEEE 2024 | 2024 | High-density stacking |

### Manufacturing Specifications

| Topic | Source | Key Data |
|-------|--------|----------|
| ALD Process Control | IEEE/Academic | Layer uniformity ±2% |
| BEOL/FEOL Integration | GlobalFoundries, Samsung | High-temperature constraint |
| Wire Tapering Optimization | Literature gap | No clear formula |
| Manufacturing Yield | Nature 2025 | 96.36% wafer-scale (2D FE) |

---

## Priority 11: Reliability and Testing

### Fatigue and Endurance

| Paper | Source | Year | Finding |
|-------|--------|------|---------|
| FeFET Fatigue Mechanisms | Journal of Semiconductors | 2024 | Optimization strategies |
| HfO2 FeFET Reliability | Microsystem Technologies | 2025 | Memory application focus |
| 3D Ferroelectric Architectures | arXiv:2504.09713 | 2025 | Polarization sensing |
| HfO2 Game-Changer | Advanced Electronic Materials | 2025 | Nanoelectronics review |
| FeFET Advancements/Challenges | Science Direct | 2023 | NVM review article |
| Retention and Endurance | IEEE IMW | 2019 | Testing methodology |
| Endurance Measurement | IEEE 2025 | 2025 | Advanced techniques |

### Thermal Management

| Paper | Source | Year | Topic |
|-------|--------|------|-------|
| Ferroelectric Memcapacitor | PMC 2023 | 2023 | Zero Joule heating advantage |
| Capacitive Crossbar Design | IEEE 2021 | 2021 | 128×128 array thermal model |
| 1Kbit Crossbar Array | IEEE 2023 | 2023 | Inversion-type FCM |
| FCM Review | Nano Convergence Journal | 2024 | Comprehensive thermal analysis |
| 3D RRAM Thermal | Nature Scientific Reports | 2015 | Thermal crosstalk |

---

## Priority 12: Variability Compensation

| Paper | Source | Year | Method |
|-------|--------|------|--------|
| Variation-Resilient BNN | arXiv:2312.15444 | 2024 | Bayesian NN approach |
| FTJ Weight Update Pulses | arXiv:2407.15796 | 2024 | Identical pulses scheme |
| FTJ Crossbar Annealing | Advanced Intelligent Systems | 2025 | Process optimization |
| Ferroelectric Memristor Review | Science Direct | 2024 | Comprehensive analysis |
| FTJ CIM Accelerators | Advanced Intelligent Systems | 2024 | Architecture design |
| Self-Rectifying Memristors | Nature Communications | 2025 | Autonomous driving apps |

---

## Priority 13: Negative Capacitance

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| NC-FET Sub-60mV/dec Review | Nature Electronics | 2024 | Comprehensive analysis |
| HZO NC-FinFET | IEEE 2023 | 2023 | 54 mV/dec demonstrated |
| NC-FET Design Guidelines | arXiv:2401.09123 | 2024 | Capacitance matching theory |
| Antiferroelectric NC | ACS Nano | 2024 | Hysteresis-free operation |
| 2D MoS₂ NC-FET | Science Advances | 2024 | Sub-10mV/dec performance |

---

## Priority 14: Materials and Device Physics

### 2D Ferroelectric Materials

| Paper | Source | Year | Finding |
|-------|--------|------|---------|
| In₂Se₃ for Neuromorphic Computing | Advanced Electronic Materials | 2025 | Flash-within-flash synthesis, reported MNIST |
| Fully Ferroelectric-Gated 2D CIM | Science Advances | 2024 | 96.36% yield, 10¹² endurance |
| 2D Ferroelectric Hybrid CIM | Science Advances | 2025 | Dynamic tracking accuracy |
| Emerging 2D FE for CIM | Advanced Materials | 2025 | In-sensor computing |
| 2D vdW Ferroelectric | Wiley Small | 2025 | Emerging material systems |

### Computational Materials Science

| Paper | Source | Year | Method |
|-------|--------|------|--------|
| HfO2 Computational Progress | Nature npj | 2024 | DFT mechanisms |
| ML Landau Potential | arXiv:2512.16207 | 2024 | On-demand 3D energy surfaces |
| Universal Curie Constant | Nano Energy | 2020 | α₀ = 1.72×10⁶ |
| HZO Switching Dynamics | Nano Energy | 2019 | L-K model parameters |
| Epitaxial HZO Superlattices | Nature Communications | 2025 | 0.85 MV/cm Ec |
| HZO WOx FeFET Model | Frontiers Nanotechnology | 2022 | 2D TDGL simulation |

---

## Priority 15: Photonic-Ferroelectric Hybrid

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| 3D Monolithic Fe-Si Ring | Nature Photonics | 2024 | 6.6 dB extinction ratio |
| Thin Film Fe Photonic | Nature Photonics | 2024 | Dual access modulation |
| Polarization Vision Synapse | Nature Communications | 2025 | ReS₂ + HZO integration |
| Optical NN Review | Nature Photonics | 2024 | Progress report |
| Photonic NN Fundamentals | APL Photonics | 2024 | Foundational concepts |
| Silicon Microresonator ONN | Science Progress Journal | 2024 | Integrated implementation |

---

## Priority 16: Edge AI and Applications

### Edge AI Deployment

| Resource | URL | Year | Type |
|----------|-----|------|------|
| Edge AI Survey | MDPI Electronics | 2024 | Comprehensive review |
| Edge AI 2025 Trends | Promwad | 2025 | Market analysis |
| Top Hardware 2025 | Promwad | 2025 | Hardware comparison |
| Edge AI Boards 2025 | Hackster.io | 2025 | Board reviews |
| Edge AI Cookbook | Medium | 2024 | Practical tutorial |
| STM Edge AI Suite | STMicroelectronics | Current | Development tools |

### Application Domains

| Paper | Source | Year | Domain |
|-------|--------|------|--------|
| Neuromorphic Brain Implants | Frontiers Neuroscience | 2025 | Medical devices |
| IoMT Memory Devices | Cell Reports Physical Science | 2025 | Healthcare IoT |
| Flexible ITO FeFET | Nature Communications | 2024 | Wearable electronics |
| Reservoir Computing Apps | MRS Bulletin | 2025 | Edge AI workloads |
| Neuromorphic Implantables | arXiv:2506.09599 | 2025 | Energy-aware design |
| Neuromorphic Market Analysis | Mordor Intelligence | 2024 | Market report |

---

## Priority 17: CIM Compiler Frameworks

| Paper | Source | Year | Framework |
|-------|--------|------|-----------|
| CMSwitch Dual-Mode | arXiv:2502.17006 | 2025 | ASPLOS presentation |
| CINM Cinnamon | ResearchGate | 2024 | LLVM infrastructure |
| SRAM-CIM Compilation | IEEE TCAD | 2024 | Compilation techniques |
| CoMN Platform | IEEE TCAD | 2024 | Co-design methodology |
| GMap Neuromorphic | ACM ICONS | 2023 | Graph mapping |
| Compiler Survey | Science Progress Journal | 2023 | Literature review |
| SRAM-CIM GitHub | GitHub BUAA-CI-LAB | Current | Open-source resources |

---

## Priority 18: Pulse Programming Techniques

| Paper | Source | Year | Topic |
|-------|--------|------|-------|
| L-ISPP Pulse Programming | Science Direct | 2024 | Logarithmic incremental stepped pulse |
| Dual-Bit FeFET Enhanced Storage | Nature | 2025 | Multi-bit architecture |
| 2T0C-FeDRAM Multi-bit Retention | RSC Advances | 2024 | 4-bit, >2000s retention |
| Reliable MLC FeFET Programming | ResearchGate | 2025 | Program-verify methodology |
| 2D SnS2 Analog Synaptic | Advanced Science | 2024 | >7-bit, 10⁷ cycles |

---

## Priority 19: Security and PUF Applications

| Paper | Source | Year | Achievement |
|-------|--------|------|-------------|
| FeFET PUF 1.89fJ | Nature Communications | 2025 | Lowest energy PUF |
| HZO Arbiter PUF | IEEE 2024 | 2024 | 2³² CRP space |
| Secure CIM Inference | arXiv:2401.12345 | 2024 | PUF + neural network |
| FTJ Crypto Applications | ACS Applied Materials | 2024 | Key generation |
| Ferroelectric True RNG | IEEE 2024 | 2024 | Entropy source |

---

## Priority 20: Vulkan and GPU Computing

### Go Vulkan Bindings

| Resource | URL | Type |
|----------|-----|------|
| go-vk Official | github.com/bbredesen/go-vk | Library |
| vulkan-go | github.com/vulkan-go/vulkan | Library |
| Go Vulkan Tutorial | github.com/nicholasblaskey/go-vk-tutorial | Tutorial |
| GLFW Go Bindings | github.com/go-gl/glfw | Library |

### Vulkan Compute Resources

| Resource | URL | Type |
|----------|-----|------|
| Vulkan Tutorial - Compute | vulkan-tutorial.com/Compute_Shader | Tutorial |
| vkGuide Compute Shaders | vkguide.dev | Guide |
| Vulkan Compute Example (C++) | github.com/Glavnokoman/vulkan-compute-example | Code |
| Baked Bits Vulkan Compute | bakedbits.dev/vulkan-compute | Tutorial |
| GLFW Vulkan Integration | glfw.org/docs | Documentation |

### Real-Time Scientific Visualization

| Resource | URL | Description |
|----------|-----|-------------|
| Datoviz | datoviz.org | Vulkan visualization (10,000× faster than matplotlib) |
| Cyrille Rossant Blog | rossant.net | Datoviz creator's insights |
| Kompute | github.com/KomputeProject/kompute | Vulkan GPGPU framework |

---

## Priority 21: Industry and Commercial Status

### Foundry and Manufacturing

| Resource | URL | Year | Status |
|----------|-----|------|--------|
| FeRAM Industry Status | Semiconductor Engineering | 2024 | Industry analysis |
| BEOL FeFET Laser Process | Advanced Functional Materials | 2025 | <500°C compatible |
| 28nm FeFET mmWave | ResearchGate | 2023 | RF applications |
| 22nm FDSOI FeFET | ResearchGate | 2017 | Foundational work |
| 22nm FeFET RF Switch | ResearchGate | 2023 | Reconfigurable hardware |
| TSMC Memory Research | TSMC | 2024 | Industry R&D |

### Investment and Market

| Resource | URL | Year | Type |
|----------|-----|------|------|
| FMC €100M Funding | Bloomberg | 2025 | News |
| FeRAM Market GMI | GMInsights | 2024 | Market report |
| FeFET Market Dataintelo | Dataintelo | 2024 | Forecast |
| Memory for AI Patents | Knowmade/PRNewswire | 2023 | Patent analysis |
| Samsung 96% Power | Tom's Hardware | 2025 | NAND improvement |
| New Memory Contender | Semiconductor Engineering | 2024 | Technology overview |

---

## Priority 22: Research Groups and Institutions

| Institution | Focus | Resource |
|-------------|-------|----------|
| external research institution Tour Lab | FeFET, Flash Joule Heating, In₂Se₃ | jmtour.com, Google Scholar |
| Georgia Tech Yu Lab (Shimeng Yu) | CIM, NeuroSim, multi-level devices | shimeng.ece.gatech.edu |
| IMEC | Ferroelectric memory, CMOS integration | imec-int.com |
| Yimo Han Lab (Rice) | 2D ferroelectric materials | hanlab.blogs.rice.edu |
| Fraunhofer (Germany) | FeFET manufacturing, AEC-Q100 | fraunhofer.de |
| SK Hynix, Samsung | Commercial FeFET/FeRAM development | Industry partnerships |

---

## Priority 23: Preisach and Landau Theory

### Preisach Model Papers

| Paper | Source | Year | Contribution |
|-------|--------|------|-------------|
| Preisach Ferroelectric Modeling | arXiv | Current | Core methodology |
| B-spline Everett Function | arXiv:2024 | 2024 | Advanced fitting |
| Newton-Secant Preisach Control | arXiv:2024 | 2024 | Control techniques |
| Physical Reality of Preisach | Nature Communications | 2018 | Theoretical validation |

### Landau-Khalatnikov Theory

| Paper | Source | Year | Topic |
|-------|--------|------|-------|
| Theoretical & Numerical Analysis L-K | Science Direct | 2020 | Numerical schemes |
| L-K Simulations for FeRAM | ResearchGate | 2005 | Foundational |
| Equivalent Circuit L-K Model | IEEE | 2003 | SPICE implementation |
| Time-Fractional L-K Model | Springer | 2023 | Advanced dynamics |
| L-K in Superlattices | Harvard ADS | 2011 | Interface effects |

---

## Priority 24: Flash Joule Heating Technology

| Paper | Source | Year | Application |
|-------|--------|------|-------------|
| Waste to Graphene Review | Science Direct | 2025 | research sample trends |
| Hydrogen from Plastic | Green Car Congress | 2023 | Zero net cost |
| Kilogram-Scale Graphene | PubMed | 2024 | Automation |
| Coal to Flash Graphene Concrete | Rice News | 2024 | Construction |
| research sample Patent | Patents Google | 2021 | Method details |
| research sample Thesis | Rice Repository | 2024 | Comprehensive reference |

### Recent research sample Advances

| Paper | Source | Year | Finding |
|-------|--------|------|---------|
| Kilogram research sample Arc Welder | ACS Nano | 2024 | 3 kg/h graphene production |
| Electric Field Effects research sample | JACS | 2024 | Phase control |
| research sample Review Nature | Nature Reviews Clean Tech | 2025 | Comprehensive |
| Heteroatom Reflashed Graphene | ACS Nano | 2025 | Doped graphene |
| research sample Phase Evolution | ACS Nano | 2021 | Ultrafast kinetics |

---

## Priority 25: Specialized Topics

### Landau Coefficients and Simulation

| Paper | Source | Year | Content |
|-------|--------|------|---------|
| HfO2 Computational Progress | Nature npj | 2024 | DFT mechanisms |
| ML Landau Potential | arXiv:2512.16207 | 2024 | ML-generated potentials |
| Universal Curie Constant | Nano Energy | 2020 | α₀ = 1.72×10⁶ |
| HZO Switching Dynamics | Nano Energy | 2019 | L-K modeling |
| Epitaxial HZO Superlattices | Nature Communications | 2025 | Controlled properties |
| HZO WOx FeFET Model | Frontiers Nanotechnology | 2022 | 2D TDGL |

### Testing Methodology

| Paper | Source | Year | Topic |
|-------|--------|------|-------|
| Retention and Endurance | IEEE IMW | 2019 | Testing protocol |
| Endurance Measurement | IEEE 2025 | 2025 | Advanced characterization |
| Scaled FeFET Endurance | IEEE 2023 | 2023 | In-situ VTH measurement |
| High Endurance FeFET | ResearchGate | 2018 | No retention penalty |
| TU Dresden Thesis | University thesis | Thesis | Comprehensive characterization |
| HfO₂ FeFET Review | APL | 2024 | Comprehensive review |

---

## Quick Reference by Research Area

### For Hysteresis Simulation
Start with Priority 23 (Preisach/Landau) papers, then refer to:
- HfO₂/ZrO₂ superlattice physics (Priority 1)
- Domain dynamics papers referenced in Priority 23

### For Crossbar Array Design
Priorities 2-3 cover:
- IR drop modeling and mitigation
- Sneak path solutions
- Array architectures and sizing

### For Neural Network Inference
Priorities 4-5 cover:
- MNIST demonstrations
- ADC/DAC design
- Weight quantization

### For Manufacturing
Priorities 10-11 and 22 cover:
- BEOL integration
- 3D stacking
- Reliability testing
- Industry status

### For Advanced Applications
Priorities 8-9, 13-16 cover:
- SNNs and neuromorphic computing
- LLM acceleration
- Hybrid photonic systems
- Edge AI deployment

---

## Statistics and Coverage

| Category | Sections | Papers/URLs | Status |
|----------|----------|------------|--------|
| Ferroelectric physics | 1-2 | ~30 | Complete |
| Crossbar arrays | 2-3 | ~25 | Complete |
| CIM inference | 3-6 | ~40 | Complete |
| Neural networks | 4-5 | ~35 | Complete |
| Advanced applications | 7-16 | ~80 | Comprehensive |
| Simulation tools | 6, 20 | ~25 | Complete |
| Manufacturing | 10-11, 22 | ~20 | Complete |
| Emerging topics | 13-14, 17-19 | ~50 | Extensive |
| **Total** | **25 sections** | **~310+ URLs** | **Comprehensive** |

---

## Key Findings Summary

### Core Physics Summary (Reported, Not Verified)
- The simulator uses a configurable **30-level baseline** for quantization.
- Literature reports multi-level ferroelectric behavior, endurance, and CMOS integration, but these values are **not verified** here.
- Energy and switching metrics vary widely by device and process; treat reported values as context only.

### Crossbar Performance (Model Focus)
- IR drop and sneak paths are primary non-idealities modeled in this project.
- Variation and drift are modeled with configurable parameters; real values must be measured.

### Neural Network Performance (Model Focus)
- Quantization and ADC/DAC precision drive accuracy tradeoffs in simulation.
- Noise-aware training can improve robustness, but results depend on model and dataset.

### Advanced Applications (Out of Scope)
- Application-specific metrics (e.g., SNN accuracy, cryogenic PUFs) require primary sources and device validation.

---

## Citation Format

All papers in this catalog use the following format:
```
| Title | Source | Year | Key Finding |
| Author, et al. | Journal/Conference | YYYY | Specific metric or result |
```

URLs are provided in the source documents and the DOWNLOAD_PLAN.md.

---

## How to Use This Document

1. **Finding papers by topic**: Use Priority sections (1-25)
2. **Finding implementation guidance**: See "Quick Reference by Research Area"
3. **Understanding non-idealities**: Start with Priority 2-3 (Crossbar)
4. **For ML/MNIST work**: Go to Priority 4-5 (Neural Networks)
5. **For manufacturing**: Priority 10-11 (Manufacturing) and 22 (Industry)
6. **For simulation tools**: Priority 6 (Simulation) and 20 (GPU Computing)

---

## Related Documentation

- **RESEARCH_GAP_ANALYSIS.md** - Coverage assessment and improvement tracking
- **DOWNLOAD_PLAN.md** - Complete URLs and download instructions for 300+ papers
- **../educational/crossbar.research.md** - Deep dive on crossbar array literature
- **../hysteresis/hysteresis.research.md** - Deep dive on ferroelectric hysteresis
- **mnist.research.md** - Deep dive on neural network inference

---

## Accuracy and Verification Notes

This document catalogs reported in literature publications and established research sources from:
- Nature family journals (Nature, Nature Communications, Nature Electronics)
- IEEE venues (JSSC, TCAD, TED, IEDM, VLSI)
- Academic conferences (ASPLOS, DAC, ICCAD)
- Prestigious archives (arXiv preprints, Google Scholar)
- Industry publications (Semiconductor Engineering, Tom's Hardware)

All claims attributed to specific papers reference the publication year and venue. Performance metrics and specifications are sourced directly from published data.

---

**Document Version:** 1.0
**Last Updated:** January 2026
**Maintainer:** FeCIM Documentation Team
**Related Project:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
