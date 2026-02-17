# Research Paper Topics - Quick Reference

A one-page guide to each research topic, with key papers and application areas.

---

## 01. Ferroelectric Materials (42 papers)

**What:** Core physics of ferroelectric materials, especially HfO₂-ZrO₂ superlattices

**Key Concepts:**
- HfO₂ ferroelectricity in thin films
- ZrO₂ doping and superlattice effects
- Domain switching and hysteresis (Preisach model)
- Negative capacitance physics
- In₂Se₃ 2D ferroelectrics
- Wakeup and fatigue mechanisms

**Why It Matters:**
- Fundamental understanding of FeCIM's switching behavior
- Enables accurate hysteresis simulation
- Guides material engineering (Al:HfO₂, Si:HfO₂, codoping)

**Typical Use:**
- Module 1 hysteresis modeling
- Physics validation and simulation tuning
- Material selection for new devices

**Key Authors:** Salahuddin, Datta, Tung, Zhou, Siannas

**Most Cited:**
- Salahuddin-Datta 2007 (Negative Capacitance)
- Chatterjee 2018 (Design and Characterization)
- Jerry 2017 (FeFET as Analog Synapse)

---

## 02. Training Algorithms (11 papers)

**What:** Neural network training with low-precision and quantized weights

**Key Concepts:**
- Quantization-aware training (QAT)
- Binarized neural networks (BNN)
- Ternary neural networks (TNN)
- Post-training quantization (PTQ)
- Analog training in crossbars

**Why It Matters:**
- FeCIM cells store only ~4.9 bits per level
- Requires quantization strategies to maintain accuracy
- Bridges deep learning to analog hardware

**Typical Use:**
- Module 3 MNIST with realistic bit-widths
- Algorithm optimization before hardware deployment
- Accuracy prediction for analog crossbars

**Key Concepts:**
- TIKI-TAKA: Analog training at low precision
- FeFET BNN with device variation resilience
- Variation-aware training

**Most Cited:**
- QAT surveys (2023)
- Variation-resilient FeFET BNN (2024)

---

## 03. Simulation Tools (11 papers + OPENSOURCE_TOOLS.md)

**What:** Open-source and commercial tools for simulating CIM and ferroelectric systems

**Available Tools:**
1. **CrossSim** (Sandia) - Realistic CIM simulation with sneak paths
2. **FerroX** - GPU phase-field ferroelectric simulation
3. **NeuroSim** - Hardware-DNN benchmark
4. **BadCrossbar** - Python crossbar simulator
5. **FERRET** - Ferroelectric domain simulation
6. **IBM AIHWKit** - Quantum-classical training
7. **COMPASS** - Crossbar compiler framework (2025)

**Why It Matters:**
- Validates design before fabrication
- Estimates power, latency, accuracy
- Connects high-level models to transistor-level detail

**Typical Use:**
- Module 2 crossbar validation
- Pre-silicon verification
- Benchmarking against other CIM techs

**Comparison Table:** See `OPENSOURCE_TOOLS.md`

**Most Cited:**
- CrossSim (SAND2021-12318C)
- FerroX GPU simulation
- NeuroSim integrated benchmark

---

## 04. CIM Architectures (32 papers)

**What:** Complete compute-in-memory system designs with ferroelectric devices

**Key Concepts:**
- Crossbar array design and layout
- Sneak path mitigation (self-rectifying arrays, current-sensing)
- ADC design and precision trade-offs
- Multi-level cell (MLC) FeFET programming
- Temperature resilience
- ADC-less hybrid designs (HCiM)
- Memory cell design (1T1C, 1D1R)

**Why It Matters:**
- Specifies actual system performance
- Shows feasibility vs NAND, RRAM, other CIM
- Reveals practical constraints (ADC power, precision)

**Typical Use:**
- Module 2 crossbar design
- Module 5 technology comparison
- Hardware-software codesign

**Key Challenges:**
- Sneak path current (10-100× signal)
- ADC power consumption (dominates total)
- Write pulse width (microseconds)
- Non-linearity of analog weights

**Most Cited:**
- Sneak path analysis papers (2022+)
- Temperature-resilient FeFET CIM (2024)
- ADC precision vs accuracy trade-offs (2024)

---

## 05. Neuromorphic Computing (7 papers)

**What:** Brain-inspired computing with spiking neurons and synaptic plasticity

**Key Concepts:**
- FeFET as artificial synapse
- Spike-timing-dependent plasticity (STDP)
- Synaptic weight update mechanisms
- 2D ferroelectric materials for neuromorphics
- Neuromorphic spintronics

**Why It Matters:**
- 100-10,000× more energy-efficient than ANNs
- FeFET's non-volatile switching mimics biological learning
- Enables edge AI and continuous learning

**Typical Use:**
- Future Module 3 extension (SNN demo)
- Edge learning applications
- Ultra-low-power IoT systems

**Emerging Areas:**
- All-ferroelectric SNNs
- Hybrid ANN-SNN systems

---

## 06. Photonic Computing (5 papers)

**What:** Neural networks using photons instead of electrons

**Key Concepts:**
- Optical phase shifters
- Photonic interferometers for matrix multiplication
- Wavelength-division multiplexing
- Integration with ferroelectric devices

**Why It Matters:**
- 1000× higher bandwidth than electrical
- Lower latency for certain workloads
- Emerging technology with high potential

**Typical Use:**
- Future high-speed inference
- Hybrid photonic-electronic systems
- Research/prototyping phase

**Status:** Early research (not yet in FeCIM)

---

## 07. Memory Architectures (3 papers)

**What:** Integration of memory and compute subsystems

**Key Concepts:**
- HBM (High Bandwidth Memory) for GPU/AI
- 3D memory stacking
- Memory-side acceleration (near-memory compute)
- Power and thermal issues

**Why It Matters:**
- FeCIM could serve as on-die memory
- Reduces data movement (energy bottleneck)

**Status:** Informational for future architectures

---

## 08. Industry Reports (5 papers)

**What:** Market surveys, technology roadmaps, industry standards

**Key Roadmaps:**
- ITRS/IRDS 2025 (semiconductor roadmap)
- AI accelerator trends
- Memory technology outlook

**Why It Matters:**
- Positions FeCIM vs industry trends
- Identifies market windows and timing
- Provides standardization context

**Typical Use:**
- Module 5 technology comparison
- Market sizing and opportunities
- Competitive landscape analysis

---

## 09. Reviews & Surveys (6 papers)

**What:** Comprehensive literature reviews on related topics

**Coverage:**
- Emerging memory technologies
- Non-volatile memory for machine learning
- In-memory computing surveys
- Flash vs NVM comparison

**Why It Matters:**
- Situates FeCIM in broader context
- Identifies research gaps
- Validates competitive claims

---

## 10. CIM Compilers & Mapping (2 papers)

**What:** Software tools to map neural networks to CIM hardware

**Tools:**
- **COMPASS** - Crossbar compiler framework (2025)
- **CIM Explorer** - BNN/TNN mapping for RRAM

**Why It Matters:**
- Automates network-to-hardware mapping
- Optimizes for crossbar constraints (ADC precision, weight quantization)

**Typical Use:**
- Module 3 MNIST compilation
- Future larger network deployment

---

## 11. Reservoir Computing (3 papers)

**What:** Recurrent neural networks with fixed random weights

**Key Concepts:**
- All-ferroelectric reservoir computing
- Analog reservoir with ferroelectric weights
- Echo-state networks

**Why It Matters:**
- Alternative to traditional DNNs
- Potentially simpler training
- Good for temporal/sequential tasks

**Status:** Exploratory for FeCIM

---

## 12. Spiking Neural Networks (7 papers)

**What:** Neural networks based on spiking neurons and event-driven computation

**Key Concepts:**
- Spike-timing-dependent plasticity (STDP)
- Neuromorphic hardware
- Energy efficiency (100-10,000× vs ANNs)
- Computer vision with SNNs
- Architecture search for SNNs

**Why It Matters:**
- FeFET naturally implements synaptic plasticity
- Extreme energy efficiency aligns with FeCIM goals
- Emerging application area

**Typical Use:**
- Future SNN-based Module 3
- Edge AI with ultra-low power
- Event-based sensors (DVS)

**Most Cited:**
- All-ferroelectric SNN (Advanced Science 2024)
- FeFET STDP implementation (IEEE EDL 2024)

---

## 13. In-Memory Training (6 papers)

**What:** On-chip neural network training using analog crossbars

**Key Concepts:**
- Backpropagation in analog hardware
- Gradient computation
- Weight update mechanisms
- Limited conductance levels
- Training accuracy with analog noise

**Why It Matters:**
- Most CIM is inference-only
- On-chip training enables edge learning
- Federated learning without data movement
- Competitive differentiator vs competitors

**Typical Use:**
- Future Module 3 training mode
- Edge AI with continuous learning
- Privacy-preserving ML

**Challenges:**
- Analog noise in gradient computation
- Limited weight precision
- Backward pass through crossbars

---

## 14. Transformer/LLM Accelerators (4 papers)

**What:** Hardware acceleration specifically for transformers and large language models

**Key Concepts:**
- Attention mechanism acceleration
- Memory bandwidth bottleneck
- In-memory attention computation
- Transformer-specific CIM design

**Why It Matters:**
- Transformers dominate modern AI
- LLMs are hottest application
- Attention is memory-bound (perfect for CIM)

**Typical Use:**
- Future Module 3 transformer demo
- LLM inference acceleration
- Showing relevance to hot AI trends

**Most Cited:**
- CIMFormer (2024)
- Analog-digital hybrid attention (2024)

---

## 15. 3D Stacking Architectures (6 papers)

**What:** Vertical integration of FeFET cells in 3D arrays (like 3D NAND)

**Key Concepts:**
- Vertical FeFET stacking (512-layer prototype, Samsung Nature 2025)
- String arrays (like NAND flash)
- Layer-to-layer interconnects
- Density scaling (Gb/mm²)

**Why It Matters:**
- 1000× density increase vs 2D crossbars
- Competes with 3D NAND Flash
- Essential for NAND replacement claims

**Typical Use:**
- Module 5 technology comparison
- Future high-density CIM arrays
- Production roadmapping

**Key Paper:**
- Samsung Nature 2025: 512-layer FeFET prototype

---

## 16. Photonic-Ferroelectric Hybrids (5 papers)

**What:** Combining ferroelectric optical modulators with photonic neural networks

**Key Concepts:**
- Non-volatile optical phase shifters
- Ferroelectric-gated modulators
- Photonic synapses
- Hybrid photonic-electronic systems

**Why It Matters:**
- 1000× bandwidth of electrical
- Novel differentiator vs pure electrical CIM
- Research frontier with high potential

**Status:** Early research (not yet in FeCIM)

**Most Cited:**
- Hybrid photonic attention accelerator (2025)
- Optical neural networks progress (2024)

---

## 17. Security & Cryptography (2 papers)

**What:** Using ferroelectric devices for hardware security

**Key Concepts:**
- Physical unclonable functions (PUFs)
- Ferroelectric charge domain computing
- Lightweight cryptography

**Why It Matters:**
- FeCIM's non-volatile switching is inherently random
- PUFs provide device authentication
- Protects AI models from extraction

**Status:** Future application area

---

## 18. ALD Process Control (5 papers)

**What:** Atomic layer deposition (ALD) techniques for HZO ferroelectrics

**Key Concepts:**
- HZO deposition recipes
- Annealing profiles (280°C to 600°C)
- Wake-free operation
- Multilayer ferroelectricity
- Heterogeneous doping

**Why It Matters:**
- Determines device performance
- Specifies manufacturing compatibility
- Enables process transfer to foundries

**Typical Use:**
- Module 6 EDA process corners
- Fab collaboration and specs
- Device yield and variability

**Key Specs:**
- Deposition temp: <300°C (BEOL compatible)
- Annealing: 280°C-500°C (process window)
- Co-doping improves wake-free behavior

---

## 19. Variability & Yield (4 papers)

**What:** Device-to-device variation, temperature effects, reliability

**Key Concepts:**
- Temperature-dependent switching
- Device mismatch
- Retention time variation
- Yield and production challenges

**Why It Matters:**
- Production quality assurance
- Reliability for automotive/aerospace
- Statistical design methodologies

**Typical Use:**
- Module 5 manufacturing comparison
- Yield prediction
- Reliability specs for customers

---

## 20. Manufacturing Integration (README only)

**What:** Integration of FeFET into existing CMOS production lines

**Key Topics:**
- BEOL/FEOL compatibility
- Thermal budget (<500°C)
- Process corners (slow/typical/fast)
- AEC-Q100 automotive qualification
- Wafer-scale manufacturing (200mm/300mm)
- Fraunhofer IPMS status

**Why It Matters:**
- Determines time-to-market
- Required for production claims
- Foundry collaboration basis

**Status:** Strategic documentation (papers in 18, 19)

---

## 21. 3D Stacking (README only)

**What:** Vertical integration roadmap and technical challenges

**Key Topics:**
- Layer count roadmap (64 → 256 → 512 → 1024)
- Monolithic 3D integration
- Via and interconnect design
- Thermal management
- Density comparison vs 3D NAND (51.2 Gb/mm²)

**Why It Matters:**
- Required for NAND replacement claims
- Differentiates from 2D-only competitors
- Production scaling path

**Status:** Strategic planning (papers in 15)

---

## 22. Automotive & Harsh Environment (README only)

**What:** AEC-Q100 qualification and extended temperature operation

**Key Topics:**
- Temperature range: -40°C to 150°C
- HTOL: 1000h @ 150°C
- AEC-Q100 Grade 0 testing
- $15B automotive memory market
- Fraunhofer IPMS qualification progress

**Why It Matters:**
- Opens $15B market
- Automotive is most demanding segment
- Grade 0 is gold standard

**Status:** Strategic documentation (papers in industry reports)

---

## 23. Cryogenic Operation (1 paper)

**What:** FeFET operation at ultra-low temperatures (4K) for quantum computing

**Key Topics:**
- 4K operation and polarization enhancement
- Quantum computing integration (qubit control, QEC)
- Quantum-classical interface design
- $2B+ market opportunity
- Deep-cryo operation (1.5K)
- FeSQUID: Ferroelectric SQUID for quantum applications

**Why It Matters:**
- Blue ocean market (few competitors)
- Quantum computing is $1B+ by 2030
- FeCIM could be quantum memory interface

**Status:** Cutting-edge research (fesquid_cryo_cam_2025.pdf)

**Key Paper:**
- FeSQUID Cryo CAM 2025: Ferroelectric SQUID for cryogenic applications

---

## Quick Reference by Application

### For AI Inference
- **Topics 02, 03, 04:** Training, simulation, CIM architectures
- **Key papers:** ADC precision (04), CrossSim (03), QAT survey (02)

### For Edge Learning
- **Topics 02, 13, 12:** Training algorithms, in-memory training, SNNs
- **Key papers:** On-chip backprop (13), STDP (12), TIKI-TAKA (02)

### For Energy Efficiency
- **Topics 04, 12, 13:** CIM architectures, SNNs, in-memory training
- **Key papers:** SNN surveys (12), analog training (13), energy analysis (04)

### For Manufacturing Readiness
- **Topics 18, 19, 20, 22:** ALD control, variability, integration, automotive
- **Key papers:** Process control (18), temperature modeling (19)

### For Market Opportunities
- **Topics 15, 21, 22, 23:** 3D stacking, automotive, cryogenic
- **Key papers:** Samsung Nature 2025 (15), Fraunhofer (22)

### For Fundamental Understanding
- **Topics 01, 03, 04:** Ferroelectric materials, simulation, CIM
- **Key papers:** Preisach (01), CrossSim (03), sneak paths (04)

---

---

**Total Papers: 167+**

*Last updated: 2026-02-02*
