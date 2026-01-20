# The Ferroelectric CIM Protocol: Solid-State Physics, Neuromorphic Architecture, and the Compute-in-Memory Revolution via Ferroelectric Superlattices

---

## Executive Summary

The semiconductor industry stands at a technological precipice defined by two existential barriers: the end of Dennard scaling and the Von Neumann "Memory Wall." While transistor density has followed Moore's Law, the energy efficiency of data movement has stagnated, making the transfer of information between memory (DRAM) and processor (CPU/GPU) the dominant bottleneck for Artificial Intelligence (AI) workloads. In this context, the emerging Ferroelectric CIM technology, incubated in Dr. external research group's laboratory at external research institution and led by Dr. Jaeho Shin, proposes a radical paradigm shift: Compute-in-Memory (CIM) based on ferroelectric oxide superlattices.

This technical report, designed as an exhaustive doctoral-level compendium, breaks down the fundamental physics, device engineering, AI algorithms, and simulation strategies necessary to understand and build Ferroelectric CIM technology. Unlike conventional solid alloys, Ferroelectric CIM uses an atomically precise HfO₂/ZrO₂ superlattice structure to stabilize the polar orthorhombic phase (Pca2₁), enabling analog synaptic devices with unprecedented linearity, endurance, and energy efficiency.

---

## AREA 1: SOLID-STATE PHYSICS IN HAFNIUM OXIDES

The foundation of Ferroelectric CIM technology is not digital electronics, but advanced materials physics. To understand how a memory can perform calculations, we must first understand how hafnium dioxide (HfO₂), a standard dielectric material, transforms into an intelligent material capable of remembering its electrical history.

### 1.1 Crystallography and Phase Stabilization

Historically, HfO₂ has been used in the CMOS industry solely as an amorphous or monoclinic "high-k" dielectric to prevent leakage currents in transistors. Under equilibrium conditions at ambient temperature and pressure, HfO₂ crystallizes in a monoclinic structure (space group P2₁/c), which is centrosymmetric and, by definition, cannot exhibit ferroelectricity. Ferroelectricity requires a breaking of inversion symmetry that allows the existence of a switchable permanent electric dipole.

#### 1.1.1 The Elusive Orthorhombic Phase (Pca2₁)

The fundamental discovery enabling Ferroelectric CIM is the kinetic stabilization of a metastable phase: the orthorhombic phase with space group Pca2₁. In this structure, oxygen atoms are displaced from their high-symmetry positions, creating two distinct oxygen sublattices: oxygen coordinated in three (3C) and four (4C) positions. The collective displacement of 3C oxygen anions along the c-axis breaks centrosymmetry and generates spontaneous polarization (Pₛ).

The thermodynamics of this phase are precarious. The free energy difference between the monoclinic phase (non-polar, stable) and the orthorhombic phase (polar, metastable) is small, but the activation barrier for transformation is high. To stabilize the Pca2₁ phase necessary for Ferroelectric CIM, entropy and strain engineering strategies are employed:

- **Grain Size Effect:** The orthorhombic phase has a lower surface energy than the monoclinic phase. Therefore, in nanocrystalline grains (<10 nm) or ultrathin films, the Pca2₁ phase becomes energetically favorable due to the surface contribution to the total Gibbs free energy.

- **Zirconium (Zr) Doping:** ZrO₂ is isostructural with HfO₂ but tends to crystallize in tetragonal phase at lower temperatures. The introduction of Zr reduces the crystallization temperature and increases lattice symmetry, facilitating the capture of the orthorhombic phase during rapid cooling after annealing.

- **Mechanical Confinement (Capping):** The use of rigid metal electrodes like Titanium Nitride (TiN) during the crystallization process exerts mechanical stress that inhibits the volume expansion and shearing necessary to transition to the monoclinic phase, thereby "freezing" the structure in the desired ferroelectric phase.

### 1.2 Ferroelectricity Mechanisms and Domains

Ferroelectricity in Ferroelectric CIM manifests through the ability to invert the polarization vector by applying an external electric field greater than the coercive field (Eᶜ). However, at the microscopic level, this process is not uniform.

#### 1.2.1 Domain Dynamics and Domain Walls

The material divides into domains—volumetric regions where all electric dipoles point in the same direction. These domains are separated by domain walls (DWs). In fluorite-structure oxides like HZO (Hafnium-Zirconium-Oxide), domain walls are extremely thin, on the order of one or two unit cells, contrasting with the more diffuse walls of traditional perovskites.

The switching process, which constitutes writing a bit or updating a synaptic weight, occurs in stages:

1. **Nucleation:** Small embryos of domains with opposite polarization form at preferential sites, typically defects or interfaces where the local energy barrier is lower. This process is stochastic and is described by the Nucleation-Limited Switching (NLS) model, implying that switching speed depends on nucleation probability rather than wall propagation speed.

2. **Growth:** Nuclei rapidly expand in the field direction (longitudinal growth) and more slowly laterally, moving domain walls through the crystal until they coalesce.

Understanding this dynamics is crucial for analog computing. While a digital memory seeks complete and rapid switching, Ferroelectric CIM exploits partial switching. By controlling the time or amplitude of the voltage pulse, domain wall propagation can be stopped at intermediate points, creating a mixture of "up" and "down" domains. The macroscopic average of this mixed polarization defines the analog conductance state (the "weight") of the device.

### 1.3 The Ferroelectric Hysteresis Curve

The fingerprint of any ferroelectric device is its Polarization-Field (P-E) hysteresis curve.

#### 1.3.1 Critical Parameters

- **Remanent Polarization (Pᵣ):** The stored charge when voltage is zero. For HZO, typical values are around 20-30 μC/cm². A high Pᵣ is vital for maximizing the signal-to-noise ratio in reading and for strongly modulating the channel conductance in a FeFET transistor.

- **Coercive Field (Eᶜ):** The field required to nullify the polarization. In HZO, Eᶜ is relatively high (~1-2 MV/cm) compared to older materials like PZT. This is advantageous for scalability, as it allows thinner devices without losing state stability, and provides excellent immunity to electromagnetic perturbations.

The area enclosed within the hysteresis loop represents the energy dissipated per cycle. In high-frequency applications like CIM, minimizing this dissipation through materials engineering is key to energy efficiency.

### 1.4 Physics of HfO₂/ZrO₂ Superlattices

Here lies Ferroelectric CIM's central innovation and Dr. Jaeho Shin's work. While the general industry uses solid solutions (HZO, a random mixture of Hf and Zr atoms), Ferroelectric CIM employs Superlattices.

#### 1.4.1 Superlattice vs. Solid Solution

A superlattice consists of alternating discrete layers of HfO₂ and ZrO₂ (for example, 2nm Hf / 2nm Zr). This ordered structure offers profound physical advantages:

- **Strain Engineering:** By epitaxially growing alternating layers, the lattice mismatch between ZrO₂ (which tends to be tetragonal/antiferroelectric) and HfO₂ generates coherent strain. ZrO₂ acts as a structural template that forces HfO₂ to adopt and maintain the ferroelectric orthorhombic phase, stabilizing ferroelectricity in a much greater thickness range (up to 100 nm) than in solid solution films.

- **Ferroelectric-Antiferroelectric Competition:** ZrO₂ is naturally antiferroelectric (AFE), meaning its adjacent dipoles tend to cancel. By intercalating it with ferroelectric HfO₂, a "frustrated" or flattened energy landscape is created. This energetic competition reduces the switching barrier, improving the linearity of analog response and reducing the operative coercive field, which is critical for low operating voltages compatible with modern logic.

- **Defect Control:** The interfaces between layers act as sinks for oxygen vacancies, preventing them from migrating and agglomerating at the electrodes, which is the main cause of fatigue and dielectric breakdown. This results in superior endurance, exceeding 10¹⁰ cycles.

---

## AREA 2: SEMICONDUCTOR DEVICES AND MEMORIES

The material physics must be encapsulated in a functional device. Compatibility with the CMOS (Complementary Metal-Oxide-Semiconductor) process is the economic imperative that separates Ferroelectric CIM from niche technologies.

### 2.1 CMOS Integration and Scalability

Ferroelectric CIM benefits from hafnium oxide already being an omnipresent material in modern chip foundries, used as a gate dielectric in advanced logic transistors. Unlike traditional ferroelectrics containing lead or bismuth (lethal contaminants for silicon), HZO is "CMOS-friendly."

The fabrication of an Ferroelectric CIM superlattice is typically performed via Atomic Layer Deposition (ALD). This process allows angstrom-level thickness control, sequentially depositing Hafnium and Zirconium precursors. Integration can be "Front-End-of-Line" (FEOL), building memory directly alongside logic transistors, or "Back-End-of-Line" (BEOL), depositing memory in the upper metallic interconnection layers, allowing memory to be stacked on logic for extreme densities.

### 2.2 Device Architecture: FeFET

There are several ways to build a ferroelectric memory, but for Artificial Intelligence in-memory (CIM) applications, the Ferroelectric Field-Effect Transistor (FeFET) is the superior architecture chosen by Ferroelectric CIM.

#### 2.2.1 FeFET vs. FeRAM vs. FTJ

| Type | Description | Limitations |
|------|-------------|-------------|
| **FeRAM (1T-1C)** | Uses a separate ferroelectric capacitor | Destructive read (data must be erased to read), requires constant rewriting, limiting speed and endurance. Not ideal for analog CIM. |
| **FTJ (Ferroelectric Tunnel Junction)** | Two-terminal device where polarization modulates a quantum tunnel barrier | Although offering high density, read currents are extremely low, making fast and precise analog summation difficult in large neural networks. |
| **FeFET (1T)** | The ferroelectric is integrated directly as the gate insulator of a transistor | The material's polarization alters the transistor's threshold voltage (Vₜₕ), modulating the conductivity of the underlying semiconductor channel. |

**CIM Advantage:** A FeFET behaves as a programmable resistance with gain. A small polarization charge controls a large drain current, allowing non-destructive, high-signal reading. Additionally, the transistor channel provides the linearity necessary for analog multiplication (I = G × V).

#### 2.2.2 The Depolarization Field Challenge

A critical challenge in FeFETs is the depolarization field (Eₐₑₚ). When no voltage is applied, polarization charges in the ferroelectric induce opposite charges in the semiconductor, creating an internal electric field that fights to depolarize the material, threatening data retention. Ferroelectric CIM superlattices, through engineering of intercalated dielectric layers, help mitigate this effect by stabilizing domains through electrostatic coupling between layers.

### 2.3 Non-Volatile Memory Comparison

To position Ferroelectric CIM in the market, comparing its fundamental metrics with competing technologies is crucial.

| Technology | Physical Mechanism | Endurance (Cycles) | Write Speed | Write Energy | Analog CIM Suitability |
|------------|-------------------|-------------------|-------------|--------------|----------------------|
| Flash NAND | Charge Trap | 10⁴ - 10⁵ | Slow (ms) | High (High Voltage) | Low (Non-linear, high variability) |
| ReRAM | Conductive Filament | 10⁶ - 10⁹ | Fast (ns) | Medium | Medium (Noise and stochasticity problems) |
| PCM | Phase Change (Heat) | 10⁸ | Medium | Very High (melting) | Medium (Resistance drift) |
| MRAM | Magnetic Spin | >10¹⁵ | Very Fast | Low | Low (Very small resistance window, ~200%) |
| **Ferroelectric CIM (FeFET)** | Superlattice Polarization | >10¹⁰ - 10¹² | Fast (ns) | Very Low (Electric Field) | High (Superior linearity, analog states) |

**Strategic Insight:** While MRAM wins in infinite endurance, it lacks the dynamic range (Rₒff/Rₒₙ) necessary to store multiple bits per cell (analog weights). ReRAM has dynamic range, but its filamentary nature is noisy and difficult to control for linear updates. Ferroelectric CIM (superlattice FeFET) occupies the "sweet spot": high endurance, low consumption (field switching, not current), and excellent analog control thanks to the superlattice's domain dynamics.

---

## AREA 3: COMPUTE-IN-MEMORY (CIM) AND SYSTEMS ARCHITECTURE

### 3.1 The Von Neumann Bottleneck

Classical computer architecture separates the processing unit (CPU/GPU) from memory. To perform an operation, data must travel through a limited bus. In the AI era, where models like GPT-4 have trillions of parameters, the energy cost of moving data exceeds by orders of magnitude the cost of computing them. Moving 64 bits of data from external DRAM memory consumes approximately 1000 times more energy than performing an addition operation with those same bits. This phenomenon is known as the "Memory Wall."

### 3.2 The Crossbar Matrix and Kirchhoff's Laws

Ferroelectric CIM solves this problem by eliminating data movement. Computation occurs within memory using a crossbar array architecture.

#### 3.2.1 Analog Matrix-Vector Multiplication (MVM)

The fundamental AI operation is multiplying a weight matrix (W) by an input vector (X): Y = W × X. In Ferroelectric CIM, this is performed instantaneously by leveraging the laws of physics:

- **Ohm's Law (Multiplication):** Each FeFET memory cell stores a weight as a conductance Gᵢⱼ. The input is applied as a voltage Vᵢ on the word line. The resulting current through the cell is Iᵢⱼ = Vᵢ × Gᵢⱼ.

- **Kirchhoff's Current Law (Summation):** The currents from all cells in a column are automatically summed at the bit line: Iⱼ = Σᵢ(Vᵢ × Gᵢⱼ).

This allows performing a complete matrix multiplication in a single clock step (O(1)), regardless of matrix size, offering massive parallelism and unmatched energy efficiency.

### 3.3 Analog Computing Challenges

Despite its efficiency, analog computing introduces noise. Precision is not infinite as in digital (32/64 bits). Ferroelectric CIM must manage:

- **Read/Write Noise:** Thermal and electronic variations.
- **Converter Consumption (ADC/DAC):** To communicate with the rest of the digital system, analog signals must be converted. High-precision Analog-to-Digital Converters (ADC) are costly in area and energy. Therefore, Ferroelectric CIM benefits from operating with low precision (INT4, INT8) where ADC requirements are lower.

---

## AREA 4: NEURAL NETWORKS AND ARTIFICIAL INTELLIGENCE

### 4.1 Mapping Algorithms to Hardware

Ferroelectric CIM hardware is designed to accelerate Deep Neural Networks (DNNs). The trained synaptic weights of a model (for example, a Transformer or CNN) are transferred to the conductance states of the FeFETs. Since conductance is always positive, device pairs (differential conductance G⁺ - G⁻) are typically used to represent positive and negative weights.

### 4.2 Linearity and Symmetry in Training

For a chip to not only execute (inference) but also learn (online training), weight updates must be predictable.

- **Linearity:** An identical voltage pulse should cause the same conductance change (ΔG) regardless of the device's current state.
- **Symmetry:** The ease of increasing the weight (Potentiation) should equal the ease of decreasing it (Depression).

Most emerging technologies fail here. Ferroelectric CIM superlattices, however, have demonstrated superior linearity and symmetry (see Figure S53). The multilayer structure moderates domain switching, avoiding abrupt changes and allowing gradual and controlled weight updating, essential for the Backpropagation algorithm to converge correctly.

### 4.3 Noise-Aware Training

Since analog hardware is intrinsically noisy, Ferroelectric CIM employs Noise-Aware Training techniques. During the model training phase (in software), deliberate Gaussian noise is injected into weights and activations, simulating the physical imperfections of the chip (cycle-to-cycle variability, read noise). This forces the neural network to find robust and flat solutions in the optimization landscape, so that when the model is loaded onto the physical Ferroelectric CIM chip, its performance does not degrade despite the device's real noise.

### 4.4 Neuromorphic Computing and SNNs

Beyond conventional Deep Learning, Ferroelectric CIM enables Spiking Neural Networks (SNNs). These networks more faithfully mimic the biological brain, communicating through voltage spikes dispersed over time.

**STDP Plasticity:** Ferroelectric CIM devices can implement Spike-Timing Dependent Plasticity (STDP), a biological learning rule where the synaptic connection strengthens if the input precedes the output. The intrinsic temporal dynamics of ferroelectric polarization and pulse interaction in the FeFET allow STDP implementation naturally without complex external circuitry.

---

## AREA 5: ADVANCED MODELING AND SIMULATION

Chip design is not done by trial and error, but through rigorous multiphysics simulation.

### 5.1 Landau-Khalatnikov Equations and TDGL

To model the fundamental physics of polarization switching in time and space, the Time-Dependent Ginzburg-Landau (TDGL) equation is used.

Where F is the free energy functional that includes Landau energy terms (double-well potential), gradient energy (cost of creating domain walls), elastic energy (superlattice strain), and electrostatic energy. Solving this partial differential equation (PDE) allows visualization of how domains nucleate and grow under the influence of superlattice structure and applied electric fields.

### 5.2 Preisach Model for Hysteresis

For faster circuit-level simulations (at the system level), the Preisach Model is used. This mathematical model decomposes the material's complex hysteresis into an infinite superposition of simple rectangular switching operators (hysterons), each with its own on and off thresholds.

**Utility:** Allows precise prediction of minor hysteresis loops and device response to arbitrary pulse sequences, crucial for designing analog write algorithms.

### 5.3 Phase-Field Modeling

Phase-field modeling is the definitive tool for visualizing microstructure. It allows simulating the 3D evolution of ferroelectric domains, showing how they interact with grain boundaries and superlattice interfaces. Phase-field simulations confirm that the superlattice structure induces smaller and denser domains, which favors analog switching linearity.

---

## AREA 6: GPU PROGRAMMING AND VULKAN COMPUTE

Simulating millions of unit cells with coupled TDGL equations requires massive computing power. This is where high-performance GPU programming comes in.

### 6.1 Why Vulkan for Scientific Simulation

Although CUDA is the academic standard, Vulkan is the strategic choice for a modern and marketable simulation tool:

- **Hardware Independence:** Vulkan runs on NVIDIA, AMD, Intel, and mobile GPUs. This democratizes the simulation tool.
- **Explicit Memory Control:** Vulkan allows manual management of memory allocation and synchronization, which is vital for optimizing bandwidth in finite difference (FDM) simulations that are memory-bound.
- **Graphics Interoperability:** Being a graphics API, simulation results (calculated in Compute Shaders) reside in GPU memory and can be visualized immediately without costly transfers back to the CPU.

### 6.2 Compute Shader Architecture

The simulation core is the Compute Shader.

- **Finite Difference Method (FDM):** The material's continuous space is discretized into a grid. A parallel shader calculates the next polarization state (Pₜ₊₁) for each cell based on the current state of its neighbors (Pₜ).
- **Synchronization:** Memory barriers are used to ensure all threads have finished reading time step t before writing step t+1, avoiding race conditions.
- **Shared Memory:** To optimize performance, neighbor data is loaded into the work group's shared memory ("local cache") of the GPU, drastically reducing global memory access latency.

---

## AREA 7: SCIENTIFIC VISUALIZATION

The ability to "see" the invisible is a powerful communication tool.

### 7.1 Vector Field Visualization

Phase-field simulations produce vector data (polarization Pₓ, Pᵧ, Pᵤ) at each point in space.

- **Isosurface Technique:** Surfaces where polarization is zero (P=0) are extracted to visualize Domain Walls in 3D.
- **Color Maps:** Divergent color maps (e.g., Red-Blue) are used to represent domain orientation (Up/Down) in cross-sections, allowing identification of complex structures like ferroelectric vortices or skyrmion topological textures that may arise in superlattices.

### 7.2 Real-Time Rendering

Using Vulkan's graphics pipeline, these simulations are rendered in real time. This allows the user (the device engineer) to interact with the simulation: change the applied voltage with a slider and instantly see how domains respond and how the P-E hysteresis curve is traced on screen, providing physical intuition that static equations cannot give.

---

## AREA 8: PRACTICAL APPLICATIONS

The ultimate test of any technology is its real-world utility. Ferroelectric CIM targets multiple high-value application domains.

### 8.1 Edge AI Inference

The primary near-term market for ferroelectric CIM:

**Target Applications:**
- Image classification (CNNs for object detection)
- Natural language processing (Transformer attention layers)
- Anomaly detection (industrial sensors, healthcare monitoring)

**Key Advantages:**
- **Ultra-low power:** <1W for inference vs 100W+ for GPU
- **Low latency:** In-memory compute eliminates data transfer delays
- **Small footprint:** No external DRAM required
- **Always-on:** Non-volatile weights enable instant wake-up

### 8.2 Neuromorphic Computing

Spiking Neural Networks (SNNs) represent the next frontier:

**Biological Advantages:**
- Event-driven processing (compute only when spikes occur)
- Temporal coding (information in spike timing)
- Massive parallelism (like biological brains)

**Ferroelectric CIM Fit:**
- FeFET polarization dynamics naturally mimic membrane potential
- STDP learning rules implementable in hardware
- Energy per spike: ~2 fJ (approaching biological efficiency)

**Applications:**
- Sensory processing (vision, audio)
- Robotics control (real-time adaptation)
- Brain-machine interfaces

### 8.3 Autonomous Systems

Self-driving vehicles and drones require:
- Real-time object detection
- Sensor fusion (LIDAR, camera, radar)
- Low power consumption
- Radiation tolerance

**FeFET Benefits:**
- Parallel processing in crossbar
- Instant-on (non-volatile)
- Robust to temperature extremes

### 8.4 IoT and Wearables

Edge computing with severe constraints:
- Battery life critical (weeks to years)
- Form factor limited
- Always-on processing required

**Solutions:**
- Near-zero standby power
- Wake-on-pattern detection
- On-device learning without cloud connectivity

### 8.5 Data Center Acceleration

While power-constrained, data centers benefit from:
- Reduced data movement (processing near storage)
- Accelerated specific workloads (inference, recommendation systems)
- Lower cooling requirements

### 8.6 Security Applications

FeFET PUF (Physical Unclonable Functions):
- **Hardware fingerprinting:** Unique device signatures from intrinsic variation
- **Secure key generation:** Non-clonable cryptographic keys
- **Anti-counterfeiting:** Authenticate genuine chips

The same cycle-to-cycle variation that challenges precision computing becomes an asset for security.

---

## AREA 9: MANUFACTURING AND COMMERCIALIZATION

Theory must materialize into a product. The path from university laboratory to market is known as the "Valley of Death."

### 9.1 Manufacturing Processes and Metrology

Ferroelectric CIM manufacturing is based on standard industry tools, reducing adoption risk.

- **ALD (Atomic Layer Deposition):** This is the critical process. It requires high-purity organometallic precursors and optimized purge times to create sharp HfO₂/ZrO₂ interfaces without diffusive intermixing.
- **Rapid Thermal Annealing (RTA):** The thermal treatment to crystallize the Pca2₁ phase must be compatible with the BEOL thermal budget (<450°C) to not damage the underlying copper transistors. Superlattice use helps reduce this required crystallization temperature.
- **Characterization:** Techniques such as Grazing Incidence X-Ray Diffraction (GIXRD) are used to confirm the crystalline phase and Transmission Electron Microscopy (TEM) to inspect superlattice interface quality.

### 9.2 Commercial Strategy and Intellectual Property

Ferroelectric CIM, as a spin-off from external research group and Jaeho Shin's laboratory, possesses a strategic advantage: intellectual property (IP) over the specific superlattice architecture for neuromorphic applications.

- **Patents:** Key patents protect the exact composition, layer thicknesses, and operation methods to achieve high linearity.
- **The Valley of Death:** The transition from TRL 3 (Lab proof of concept) to TRL 7 (Prototype in operational environment) requires intensive capital. Ferroelectric CIM has secured initial funding ("One Small Step Grant"), but long-term success will depend on alliances with major manufacturers (Foundries like TSMC or GlobalFoundries) that need to integrate high-density non-volatile memory in their advanced logic nodes for Edge AI applications.

---

## CONCLUSION: MASTERING FECIM

This curriculum spans from the movement of an oxygen atom in a crystal lattice to the execution of massive language models on a chip. Mastering these nine areas not only provides the knowledge to build Ferroelectric CIM but grants a comprehensive vision of the future of computing.

Ferroelectric superlattice technology represents a rare convergence of new physics, manufacturing compatibility, and urgent market need. By replacing data movement with materials physics, Ferroelectric CIM has the potential to redefine computational efficiency for the age of artificial intelligence. Armed with this deep knowledge, you are positioned not only to observe this revolution but to lead it.

---

## Appendix: Technical Data Tables

### Table 1: HfO₂ Crystalline Phases

| Phase | Crystal System | Space Group | Stability (Bulk) | Electrical Property |
|-------|----------------|-------------|------------------|---------------------|
| Monoclinic (m) | Monoclinic | P2₁/c | Stable (RT) | Dielectric (Non-polar) |
| Tetragonal (t) | Tetragonal | P4₂/nmc | High Temp / AFE | Antiferroelectric |
| Cubic (c) | Cubic | Fm3̄m | Very High Temp | Paraelectric |
| Orthorhombic (o) | Orthorhombic | Pca2₁ | Metastable | Ferroelectric (Polar) |

### Table 2: Vulkan Compute Pipeline Architecture for TDGL

| Stage | Vulkan Object | Function in Ferroelectric CIM Simulation |
|-------|---------------|-----------------------------------|
| Memory | VkDeviceMemory | Stores the polarization grid (Pᵢⱼₖ) in VRAM. |
| Buffer | VkBuffer (Storage) | Interface for shader to read/write cell states. |
| Shader | VkShaderModule (SPIR-V) | GLSL kernel that solves the FDM differential equation. |
| Execution | vkCmdDispatch | Launches millions of threads in parallel (one per cell). |
| Synchronization | VkPipelineBarrier | Ensures temporal coherence between integration steps t and t+1. |

---

## Works Cited

1. The interplay of ferroelectricity and magneto-transport in non-magnetic moiré superlattices, https://pmc.ncbi.nlm.nih.gov/articles/PMC12217394/
2. Rice Innovation awards fourth cycle of One Small Step Grants, https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants
3. Enhancing ferroelectric stability: wide-range of adaptive control in ..., https://pmc.ncbi.nlm.nih.gov/articles/PMC12254504/
4. (PDF) Atomic-scale ferroic HfO2-ZrO2 superlattice gate stack for advanced transistors, https://www.researchgate.net/publication/350926295_Atomic-scale_ferroic_HfO2-ZrO2_superlattice_gate_stack_for_advanced_transistors
5. First-principles predictions of HfO2-based ferroelectric ... - arXiv, https://arxiv.org/pdf/2401.05288
6. Progress in computational understanding of ferroelectric mechanisms in HfO2, https://liutheory.westlake.edu.cn/pdf/s41524-024-01352-0.pdf
7. Ferroelectricity in Simple Binary ZrO2 and HfO2. - Semantic Scholar, https://www.semanticscholar.org/paper/Ferroelectricity-in-Simple-Binary-ZrO2-and-HfO2.-M%C3%BCller-B%C3%B6scke/36804402a5834490932a15052b1334e1c853c3d8
8. Demonstration of ferroelectricity in PLD grown HfO 2 -ZrO 2 nanolaminates - AIMS Press, https://www.aimspress.com/article/doi/10.3934/matersci.2023018?viewType=HTML
9. Physics-informed models of domain wall dynamics as a route for autonomous domain wall design via reinforcement learning - RSC Publishing, https://pubs.rsc.org/en/content/articlehtml/2024/dd/d3dd00126a
10. Physics and applications of charged domain walls - Infoscience, https://infoscience.epfl.ch/server/api/core/bitstreams/31a85b0f-9b73-4b9d-98d4-a694f7485aac/content
11. BEOL-Compatible Superlattice FEFET Analog Synapse With Improved Linearity and Symmetry of Weight Update - IEEE Xplore, https://ieeexplore.ieee.org/document/9691825/
12. ZrO2-HfO2 Superlattice Ferroelectric Capacitors With Optimized Annealing to Achieve Extremely High Polarization Stability | Request PDF - ResearchGate, https://www.researchgate.net/publication/362258936_ZrO2-HfO2_Superlattice_Ferroelectric_Capacitors_With_Optimized_Annealing_to_Achieve_Extremely_High_Polarization_Stability
13. Why is nonvolatile ferroelectric memory field-effect transistor still elusive? - ResearchGate, https://www.researchgate.net/publication/3254357_Why_is_nonvolatile_ferroelectric_memory_field-effect_transistor_still_elusive
14. (PDF) Crossbar Array of Artificial Synapses Based on Ferroelectric Diodes - ResearchGate, https://www.researchgate.net/publication/355360598_Crossbar_Array_of_Artificial_Synapses_Based_on_Ferroelectric_Diodes
15. (PDF) High Linearity and Symmetry Ferroelectric Artificial Neuromorphic Devices Based on Ultrathin Indium‐Tin‐Oxide Channels - ResearchGate, https://www.researchgate.net/publication/390837342_High_Linearity_and_Symmetry_Ferroelectric_Artificial_Neuromorphic_Devices_Based_on_Ultrathin_Indium-Tin-Oxide_Channels
16. arXiv:2307.09357v1 [cs.ET] 18 Jul 2023, https://arxiv.org/pdf/2307.09357
17. Improving Linearity and Symmetry of Synaptic Update Characteristics and Retentivity of Synaptic States of the Domain-Wall Device - IEEE Xplore, https://ieeexplore.ieee.org/iel8/8782713/10829839/10787236.pdf
18. The Relentless Genius of external research group | Office of Research | external research institution, https://research.rice.edu/news/relentless-genius-james-tour
19. Binarized Sensing Layer - Emergent Mind, https://www.emergentmind.com/topics/binarized-sensing-layer
20. Negative Feedback Training: A Novel Concept to Improve Robustness of NVCIM DNN Accelerators - arXiv, https://arxiv.org/html/2305.14561v3
21. Spike Optimization to Improve Properties of Ferroelectric Tunnel Junction Synaptic Devices for Neuromorphic Computing System Applications - Preprints.org, https://www.preprints.org/manuscript/202309.0008
22. Coupled Time-Dependent Ginzburg-Landau Equation for Superconductivity and Elastic - OSTI.GOV, https://www.osti.gov/servlets/purl/2341973
23. An Efficient Numerical Algorithm for Solving Coupled Time-Dependent Ginzburg-Landau Equation for Superconductivity and Elasticity - ResearchGate, https://www.researchgate.net/publication/395927943_An_Efficient_Numerical_Algorithm_for_Solving_Coupled_Time-Dependent_Ginzburg-Landau_Equation_for_Superconductivity_and_Elasticity
24. Review of Play and Preisach Models for Hysteresis in Magnetic Materials - PubMed Central, https://pmc.ncbi.nlm.nih.gov/articles/PMC10051722/
25. Preisach model for the simulation of ferroelectric capacitors - SciSpace, https://scispace.com/pdf/preisach-model-for-the-simulation-of-ferroelectric-234qtsdv2o.pdf
26. Phase-field model of multiferroic composites: Domain structures of ferroelectric particles embedded in a ferromagnetic matrix - Computational Materials Science Group, http://www.mmm.psu.edu/PWu2010_PM_MultiferroicComposites.pdf
27. Pyramidal charged domain walls in ferroelectric BiFeO 3 - arXiv, https://arxiv.org/html/2501.01190v1
28. [P] Vulkan as an alternative to CUDA in scientific simulation software - computational magnetism : r/MachineLearning - Reddit, https://www.reddit.com/r/MachineLearning/comments/ilcw2f/p_vulkan_as_an_alternative_to_cuda_in_scientific/
29. Real-time Particle-based Snow Simulation with Vulkan - GitHub, https://github.com/giaosame/RealTimeParticleBasedSnowSimulation
30. Basic steps for setting up a bare minimum compute shader for beginners like myself - Reddit, https://www.reddit.com/r/vulkan/comments/1aun2fc/basic_steps_for_setting_up_a_bare_minimum_compute/
31. Compute chapter for Vulkan-Tutorial - - Sascha Willems, https://www.saschawillems.de/blog/2023/02/08/compute-chapter-for-vulkan-tutorial/
32. Vulkan Examples - PowerVR Developer Documentation, https://docs.imgtec.com/sdk-documentation/html/examples/vulkan-examples.html
33. Thermally Stable Ferroelectric Memory > Patents - KAIST MII LAB, https://mii.kaist.ac.kr/bbs/board.php?bo_table=sub3_3&wr_id=37&page=4
34. A-SITE-AND/OR B-SITE-MODIFIED PBZRTIO3 MATERIALS AND (PB, SR, CA, BA, MG) (ZR, TI,NB, TA)O3 FILMS HAVING UTILITY IN FERROELECTRIC RANDOM ACCESS MEMORIES AND HIGH PERFORMANCE THIN FILM MICROACTUATORS - NASA Technical Reports Server (NTRS), https://ntrs.nasa.gov/citations/20080007013
35. Leading a Carbon Nanotube Innovator Out of the Valley of Death - EE Times, https://www.eetimes.com/leading-a-carbon-nanotube-innovator-out-of-the-valley-of-death/
36. Page Reply to BIS-2021-0011 March 28, 2021 To: Semiconductor Manufacturing Supply Chain Re - Regulations.gov, https://downloads.regulations.gov/BIS-2021-0011-0006/attachment_1.pdf
