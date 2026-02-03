# Research Notes: Open Source Ferroelectric Tools

> **Note:** Internal notes only. Any numeric claims are reported from sources and not independently verified.

## 1. CrossSim
**Repository:** https://github.com/sandialabs/cross-sim
**Description:** A GPU-accelerated simulator for analog in-memory computing, modeling device/circuit non-idealities for crossbar arrays.
**Key Features:**
- Neural network interface (PyTorch, Keras)
- Hardware-aware training
- Device models: RRAM, PCM, FeFET, SRAM
- Non-idealities: Programming errors, conductance drift, read noise

**Key Papers:**
- "The Impact of Analog-to-Digital Converter Architecture and Variability on Analog Neural Network Accuracy" (2023)
- "An Accurate, Error-Tolerant, and Energy-Efficient Neural Network Inference Engine Based on SONOS Analog Memory" (2022)
- "On the accuracy of analog neural network inference accelerators" (2022)
- "CrossSim: GPU-Accelerated Simulation of Analog Neural Networks" (SAND2021-12318C)
- "CrossSim: a hardware/software co-design tool for analog in-memory computing" (SAND2024-05171C)

## 2. WaCPro
**Repository:** https://github.com/DUTh-FET/WaCPro
**Description:** Open-source GUI application for waveform generation and crossbar programming.
**Key Features:**
- GUI-based design
- Waveform types: Step, Ramp, Half-Sine, Square
- Visualization: Preview pulses, heatmaps
- Export formats: .txt, .csv, .mat

**Key Papers:**
- "WaCPro: An Open-Source Application for Waveform and Crossbar Programming in Nanotechnology Research" (IEEE Open Journal of Nanotechnology)

## 3. PEtra
**Repository:** https://github.com/IONICS-Lab/PEtra (Note: Search found connection to ETH Zurich paper, need to verify repo URL)
**Description:** Open-source PE loop tracer for polymer piezoelectrics.
**Key Features:**
- Ultra-low current sensitivity (2 pA)
- Adjustable gain (10^3 to 10^7 V/A)
- Frequency range: 0.1 Hz to 200 Hz

**Key Papers:**
- "PEtra: A Flexible and Open-Source PE Loop Tracer for Polymer Thin-Film Transducers" (arXiv, 2024)

## 4. ferro_scripts
**Repository:** https://github.com/WMD-group/ferro_scripts
**Description:** Python script for generating ferroelectric hysteresis loops.
**Key Features:**
- Physics-based model (Garrity et al. 2014)
- YAML configuration for material parameters
- Material presets: BaTiO3, CrCA

**Key Papers:**
- "Hyperferroelectrics: Proper Ferroelectrics with Persistent Polarization" (Phys. Rev. Lett. 112, 127601, 2014)

## 5. Preisachmodel
**Repository:** https://github.com/fddf22/Preisachmodel
**Description:** Forward and numerically inverted Preisach model implementation.
**Key Features:**
- Forward model: Input field -> Output polarization
- Inverse model: Numerical inversion for control
- Based on Mayergoyz's theory

**Key Papers/Books:**
- "Mathematical Models of Hysteresis" (IEEE Trans. Mag, 1986, Mayergoyz)
- "The Science of Hysteresis" (Mayergoyz & Bertotti, 2005)

## 6. FerroX
**Repository:** https://github.com/AMReX-Microelectronics/FerroX (Verified)
**Description:** A GPU-accelerated, 3D Phase-Field Simulation Framework for Modeling Ferroelectric Devices.
**Key Features:**
- Solves Time-Dependent Ginzburg-Landau (TDGL) equation
- AMReX-based (massively parallel, GPU support)
- 15x speedup on GPUs
- Modeling ferroelectric domain-wall induced negative capacitance

**Key Papers:**
- "FerroX: A GPU-accelerated, 3D Phase-Field Simulation Framework for Modeling Ferroelectric Devices" (arXiv:2210.15668, 2022)

## 7. IBM AIHWKit
**Repository:** https://github.com/IBM/aihwkit
**Description:** Python library for simulating analog in-memory computing (AIMC) for neural networks.
**Key Features:**
- PyTorch integration
- Models noise and non-idealities of AIMC chips
- Analog AI Cloud Composer
- AIHWKIT-Lightning for scalable training

**Key Papers:**
- "Using the IBM Analog In-Memory Hardware Acceleration Kit for Neural Network Training and Inference" (arXiv:2307.09357, APL Machine Learning, 2023)
- "IBM Analog Hardware Acceleration Kit for ICML 2021" (Demo paper)

## 8. Dr. external research group / Jaeho Shin
**Research Focus:** 2D ferroelectric semiconductors for neuromorphic computing.
**Key Papers:**
- "In2Se3 Synthesized by the FWF Method for Neuromorphic Computing" (ChemRxiv 2024, Adv. Electronic Materials 2025)
  - Describes gram-scale synthesis of alpha-In2Se3 using Flash-Within-Flash (FWF) Joule heating.
  - Demonstrates synaptic devices with ~87% accuracy on MNIST.
- "Stoichiometric Engineering of Indium Selenide Compounds Realized by Flash-within-Flash with an Arc Welder" (ACS Nano, 2025)
  - Describes alpha-In2Se3 FET devices with ferroelectric switching and hysteresis.

## 9. Competitors (2D Ferroelectrics / Neuromorphic)
**Key Groups & Papers:**
- **CUHK (Chunsheng Chen):** "Emerging 2D Ferroelectric Devices for In-Sensor and In-Memory Computing" (Adv. Mater. 2025)
- **Shenzhen University (Yongbiao Zhai):** "Reconfigurable 2D-ferroelectric platform for neuromorphic computing" (Applied Physics Reviews 2023)

## 10. HZO / SiO CMOS Integration
**Key Findings:**
- ZrO2 interfacial layers generally improve ferroelectricity compared to SiO2.
- SiO2 naturally forms and can be engineered (thinned) to reduce operating voltage.

**Key Papers/Topics to Find:**
- "Impact of Interfacial Layers on Ferroelectric Hf0.5Zr0.5O2" (General topic/search)
- "Wake-up free ferroelectric HZO with ZrO2 interfacial layer"

## 11. CIM Algorithms & Electronics
**Key Algorithms:**
- **Quantization:** Critical for CIM. Mixed-precision and binary neural networks (BNN) are common.
- **Hardware-Aware Training:** Techniques like A-TRICE to handle device variation and noise.
- **Mapping Frameworks:** "Weight Pool" (CIMPool) for compression.

**Key Electronics/Circuits:**
- **ADC/DAC:** Major bottleneck. Research on "ADC-less" designs or IMADC (In-Memory ADC).
- **Architectures:** Crossbar arrays (ReRAM, FeFET, PCM). Hybrid architectures.

**Key Papers to Find:**
- "A Survey of Computing-in-Memory Processor: From Circuit to Application" (IEEE Open Journal of SSC, 2023)
- "Compute-in-Memory based Neural Network Accelerators for Safety-Critical Systems: Worst-Case Scenarios and Protections" (A-TRICE) (arXiv:2312.01633)
- "Analog, In-memory Compute Architectures for Artificial Intelligence" (Review)

## 12. Massive Paper Hunt (2024-2025 Findings)
### Tour/Shin (Rice)
- **"Ultrafast and scalable materials synthesis with flash-within-flash Joule heating"** (Nature Chem 2024). Key for material quality.
- **"Electric Field Effects in Flash Joule Heating Synthesis"** (JACS 2024). Relevant to ferroelectric formation?

### Competitors & CIM
- **KAIST (Shinhyun Choi):** "Ultra-low Power Memory" (Nature 2024). Phase change, but relevant to neuromorphic.
- **Tsinghua (Peng Yao):** "28 nm RRAM-Based 81.1 TOPS/mm2/bit CIM Macro" (2024).
- **Tsinghua (Huaqiang Wu):** "Face Classification using Electronic Synapses" (Nature Comms 2025).

### HZO Reliability (IMEC)
- **"La Doped HZO-Based 3D-Trench Metal-Ferroelectric-Metal Capacitors"** (IEEE EDL / IRPS 2024). High endurance (>10^12).
- **Non-destructive readout Mechanism** (IEDM/IMEC 2024). >10^11 read cycles.

## 13. Advanced Materials & 3D Architectures (2024-2025)
### AlScN (High-Pr Ferroelectric)
- **"Demonstration of highly scaled AlScN ferroelectric diode memory"** (2025). High density (>100 Mbit/mm2), 5nm scaling.
- **"Ferroelectric Aluminum Scandium Nitride Transistors..."** (2024). Synaptic functions, 93.8% accuracy.
- **Key Advantage:** High Remanent Polarization (Pr) suitable for FeDiodes and FTJs.

### Vertical FeFET (3D NAND-like)
- **"Ferroelectric transistors for low-power NAND flash memory"** (Nature 2025). FeNAND as alternative to Charge Trap.
- **"A Full Spectrum of 3D Ferroelectric Memory Architectures"** (April 2025). Discusses VC-FeNAND.
- **Key Concept:** Using the vertical channel structure of 3D NAND but replacing the charge trap layer with ferroelectric (HZO) for lower voltage operation.

### Negative Capacitance (NCFET)
- **"NCFET-based SRAM computing-in-memory"** (Expected 2025). Lower energy than CMOS CiM.
- **Impact:** Steep subthreshold slope for low-power logic/compute.

## 14. Global Competitors (US/EU Focus)
### NaMLab (Germany) - The HZO Powerhouse
- **Focus:** HfO2-based ferroelectrics, FeFETs, and 3D integration.
- **Key 2025 Work:** "Ferroelectricity in hafnium oxide: CMOS compatible FeFETs" (Aug 2025).
- **Recent Findings:** Interface charge effects, HZO thickness scaling in FTJs.

### UC Berkeley (Salahuddin) - Negative Capacitance
- **Focus:** NCFETs, overcoming Boltzmann limit.
- **Key 2025 Work:** "Enabling Floating Body Effect... for Spiking Neuron" (EDTM 2025).

### Stanford (Wong/Pop) - 2D & Memory
- **Focus:** 2D material memory, Phase Change Memory, FeFET.
- **Key 2024 Work:** "Flexible Ferroelectric Memory" (IFETC 2024).
- **Key 2025 Work:** "Future of Memory: Massive, Diverse..." (Discusses FeFET integration).

## 15. CMOS Integration Resources
- **"Ferroelectric Non-Volatile Memory Guide"**: Practical integration challenges.
- **"Integration of FeRAM Devices into a Standard CMOS Process"**: Detailed process impact analysis.
- **"Analog, In-memory Compute Architectures for AI"**: Critical for the "CMOS Analog Design" request.

## 16. Patents & Industrial IP (2024-2025)
### Samsung Electronics
- **Patent US20240015983A1** (Jan 2024): "Three-dimensional ferroelectric memory device".
    - *Key Tech*: FeFETs in 3D NAND structure for low voltage.
- **Patent US20240206188A1** (Feb 2024): Collab with KAIST on "Ferroelectric capacitors...".

### external research group (Rice) Patents
- **Finding**: Patents focus on "Flash Joule Heating" (research sample) platform, not specific "Ferroelectric" devices.
- **Implication**: The *method* is patented (US Patent for research sample synthesis), applied to *all* materials (including potential ferroelectrics).

## 17. Deep Academic Sources (Dissertations)
- **Mattia Segatto (2024)**: "Modeling and Simulation of Ferroelectric-based Devices for Neuromorphic Computing".
    - *Scope*: FTJs, Antiferroelectric models, Neuromorphic focus.
- **Stanford Thesis (HZO)**: "Hafnium oxide based ferroelectric materials for memory applications".
- **Stanford Thesis (HZO)**: "Hafnium oxide based ferroelectric materials for memory applications".
- **NaMLab Report (March 2025)**: "NaMLab Two-Year Report 2022.2023". Critical source for latest HZO data.

## 18. Compute-in-Memory Deep Dive (2024-2025)
### SRAM-CIM (Digital & Analog)
- **Trends**: High precision (FP16/BF16) and transformer acceleration.
- **Key 2025 Papers**:
    - "A 22-nm 109.3-to-249.5-TFLOPS/W Outlier-Aware Floating-Point SRAM CiM" (JSSC 2025).
    - "A 28nm 64kb Bit-Rotated Hybrid-CIM Macro" (ISSCC 2025).
- **Architecture**: Move towards "ADC-less" or hybrid analog-digital to save area.

### RRAM/MRAM/FeFET CiM
- **RRAM**: "A Novel High-Density and Stackable 3D Vertical RRAM-Based CIM Macro" (IEDM 2025).
- **FeFET**: "FeFET-based Ternary Neural Networks for robust CiM" (Nov 2025).
- **ADC-less**: "HCiM: ADC-Less Hybrid Analog-Digital Compute in Memory Accelerator" (2024). Eliminates power-hungry ADCs via algorithm-hardware co-design.

## 19. Critical Research Gaps & Weaknesses
### Weakness 1: Tour's research sample Process Specificity
- **Issue**: We have general research sample patents, but **no specific recipe** (voltage, pulse width, precursors) for synthesizing *ferroelectric* phases. The connection is theoretical ("Relevant to ferroelectric formation?").
- **Action Needed**: Deep search for "Flash Joule Heating of Metal Oxides/Chalcogenides" specifically looking for phase transformation data.

### Weakness 2: Lack of Negative Results
- **Issue**: Research is biased towards "record-breaking" results. Missing data on **yield**, **variability** (device-to-device), and **failure modes** (e.g., AlScN dielectric breakdown vs coercive field).
- **Action Needed**: Search for "reliability issues", "variability", and "failure analysis" of AlScN and 2D ferroelectrics.

### Weakness 3: Comparative Benchmarks (The "So What?")
- **Issue**: We have isolated numbers (81 TOPS/W, 10^12 cycles) but no **unified comparison table**. How does NaMLab's FeFET *really* compare to Tsinghua's RRAM-CIM in a system-level simulation?
- **Action Needed**: Aggregate metrics into a "Figures of Merit" comparison table.

### Weakness 4: Toolchain Maturity
- **Issue**: Tools like CrossSim and IBM AIHWKit are mature for RRAM/PCM, but support for **FeFET (especially 3D/Vertical)** and **AlScN** is likely non-existent or requires custom modeling.
- **Action Needed**: Verify if these tools allow custom device model injection (Verilog-A/Python) for novel materials.

## 20. Verification & Gap Filling (Executed)
### A. Flash Joule Heating (research sample) Recipe
- **Parameters Found**: 3000 Kelvin, 0.1s - 1s pulse duration, ~80 Volts (discharging).
- **Precursors**: Organometallics embedded in Carbon/Graphene, or Metal Salts (Fe, Mn, Ni) dissolved in water.
- **Status**: General parameters found, but *specific* ferroelectric phase recipe is still a research gap (requires experimentation).

### B. AlScN Reliability Data
- **Endurance**: >10^10 cycles achieved with *partial polarization switching*.
- **Failure Mode**: Abrupt failure due to conductive filament formation (filamentary breakdown).
- **Control**: HfO2 interlayer can improve breakdown field.

### C. Figures of Merit (FOM) Comparison

| Metric | AlScN (New) | FeFET (HZO) | RRAM-CIM (Tsinghua) | SRAM-CIM (TSMC) |
| :--- | :--- | :--- | :--- | :--- |
| **Endurance** | >10^10 (Partial) | 10^12 (Opt.) | 10^9 - 10^11 | >10^16 (Infinite) |
| **Speed** | <10 ns (Fast) | ~10 ns | ~100 ns | <1 ns |
| **Efficiency** | High (Low leakage) | High | 81 TOPS/W | ~4000 TOPS/W |
| **Voltage** | High (~10-20V) | Low (<5V) | Med (~1-3V) | Low of V_dd |
| **Density** | High (5nm compatible) | High (3D NAND) | High | Low (6T-8T Large) |
| **Non-Volatile** | Yes | Yes | Yes | No |

### D. Tool Support for Custom Models
- **CrossSim**: Supports custom Python functions for device errors and lookup tables. Defined in `inference_config.py`.
- **IBM AIHWKit**: Robust support for "Custom Resistive Devices" via `PulsedDevice` class. Allows defining custom update behaviors in Python.
- **Conclusion**: We CAN build custom AlScN/FeFET models in these tools without waiting for official support.

## 21. FTJ Device Parameters (Extracted from Verified Paper)
**Source:** "Asymmetric Resonant Ferroelectric Tunnel Junctions for Simultaneous High Tunnel Electroresistance and Low Resistance-Area Product" (arXiv:2504.11137)
**Device Structure:** TiN / HZO (3nm) / TiO2 (Quantum Well) / HZO (1nm) / TiN
**Key Parameters:**
- **HZO Thickness:** 4 nm total (3nm + 1nm split)
- **Quantum Well (TiO2):** ~2 nm (optimal for resonance)
- **Polarization (Ps):** 5.5 μC/cm²
- **Remanent Polarization (Pr):** 5.0 μC/cm²
- **Coercive Field (Ec):** 1 MV/cm
- **Performance:**
    - **TER Ratio:** ~6.15 × 10⁴ %
    - **RA Product:** 47.1 Ω·cm² @ 0.175V
- **Material Properties (for Simulation):**
    - **HZO:** Affinity = 2.8 eV, Dielectric Constant = 25, Eff. Mass = 0.14m₀
    - **TiO2:** Affinity = 4.0 eV, Dielectric Constant = 48, Eff. Mass = 0.5m₀
    - **TiN:** Work Function = 4.3 eV
