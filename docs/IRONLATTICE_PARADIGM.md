# The IronLattice Paradigm: Ferroelectric Superlattice Architectures for Neuromorphic Compute-In-Memory

---

## 1. Introduction: The Thermodynamic Crisis of Artificial Intelligence

The global computational infrastructure stands at a precipice defined by a thermodynamic crisis. As Artificial Intelligence (AI) models transition from experimental curiosities to the backbone of the global economy, the energy required to train and deploy these systems has escalated exponentially. The current trajectory of AI development, driven by Large Language Models (LLMs) and generative networks scaling into the trillions of parameters, is fundamentally at odds with the physical limitations of the incumbent hardware architecture.

This report provides an exhaustive technical analysis of "IronLattice," a breakthrough ferroelectric Compute-In-Memory (CiM) technology developed by Dr. external research group and his team at external research institution. By shifting the computational paradigm from digital data shuttling to analog in-situ processing, this technology promises to reduce the energy consumption of AI inference by over 90%, addressing the single most critical bottleneck in modern computing.

### 1.1 The Energy Implications of the Von Neumann Bottleneck

For over seventy years, the digital world has operated on the von Neumann architecture, a design philosophy that physically separates the processing unit (CPU/GPU) from the memory unit (DRAM/SRAM). In this model, data must be fetched from memory, transported across a bus to the processor, computed upon, and then written back to memory. For general-purpose computing, this flexibility is invaluable. However, for the specific matrix-heavy workloads of Deep Neural Networks (DNNs), it is energetically catastrophic.

In modern AI workloads, the "memory wall" has become the dominant factor in power consumption. Research indicates that accessing a single piece of data from off-chip DRAM consumes between 100 to 1,000 times more energy than the floating-point operation performed on that data. Consequently, in large-scale inference tasks, over 90% of the total energy budget is expended not on "thinking" (calculation), but on "moving" (data transport). This inefficiency creates a hard ceiling for the deployment of AI at the edge—on drones, satellites, and mobile devices—where power budgets are measured in milliwatts rather than kilowatts.

### 1.2 The IronLattice Solution

Dr. Tour's solution, commercialized through the startup IronLattice, attacks this problem at the device physics level. By utilizing ferroelectric superlattices based on Hafnium Oxide (HfO₂) and Zirconium Oxide (ZrO₂), the Tour Group has developed a hardware architecture where memory cells double as computational units. This Compute-In-Memory approach eliminates the data bus entirely for the weight matrices of neural networks. The computation is performed in the analog domain, utilizing the intrinsic physical properties of the material to perform multiplication and accumulation instantly and in parallel.

This report dissects the underlying physics of these ferroelectric superlattices, benchmarks them against industry incumbents like NAND Flash and Google's TPU, and outlines the commercial and theological context of Dr. Tour's work.

---

## 2. The Architecture of Efficiency: Deconstructing Compute-In-Memory

To appreciate the magnitude of the IronLattice innovation, one must understand the mechanics of Compute-In-Memory (CiM) and how it diverges from the digital logic that has defined the silicon era.

### 2.1 Principles of Analog Compute-In-Memory

Traditional digital accelerators (like GPUs or TPUs) simulate neural networks. They store weights as binary numbers (0s and 1s) and use logic gates to perform binary multiplication. In contrast, the IronLattice architecture physically emulates the neural network.

#### 2.1.1 The Crossbar Array Topology

The core structure of the IronLattice device is the crossbar array. In this grid, word lines (rows) and bit lines (columns) intersect. At each intersection sits a ferroelectric memory device.

- **Weights as Conductance:** The "weight" of a neural connection is stored not as a digital number, but as the physical conductance (G) of the ferroelectric device. By tuning the polarization state of the ferroelectric material, the conductance can be set to a specific analog value.

- **Inputs as Voltage:** The input data (activations) are applied as voltage pulses (V) along the rows.

- **Multiplication via Ohm's Law:** As the voltage passes through the device, the current (I) that emerges is the product of the voltage and the conductance (I = V × G). This is a physical multiplication, occurring instantly without logic gates.

- **Accumulation via Kirchhoff's Law:** The currents from all devices in a column merge onto the bit line. According to Kirchhoff's Current Law, these currents sum algebraically (I_total = ΣI).

This architecture allows an entire matrix-vector multiplication—the fundamental operation of AI—to occur in a single clock cycle, regardless of the matrix size. The energy cost of moving weights is zero because the weights remain stationary in the device structure.

### 2.2 The 90% Energy Reduction Claim

The claim of a 90% reduction in energy usage is derived from a comparative analysis of data movement costs versus switching costs.

- **Baseline (Von Neumann):** In a standard GPU, fetching a 32-bit weight from DRAM requires ~640 picojoules (pJ). The Multiply-Accumulate (MAC) operation itself takes only ~3-5 pJ. Thus, data movement accounts for >99% of the energy in memory-bound workloads.

- **IronLattice (CiM):** In the CiM architecture, the "fetch" energy is eliminated. The energy cost is limited to the dynamic power dissipated by the current flowing through the array and the peripheral analog-to-digital converters (ADCs). Because ferroelectric switching is purely field-driven (utilizing displacement current rather than conduction current during programming), the write energy is also exceptionally low compared to current-driven technologies like MRAM or ReRAM.

It is important to nuance this claim by noting that while hardware innovations like IronLattice target a 90% reduction via architectural efficiency, concurrent research in algorithmic efficiency also targets similar reductions. The synergy of hardware CiM (IronLattice) with algorithmic compression offers a potential combined efficiency gain of over 99% compared to unoptimized, GPU-based baselines.

---

## 3. Material Physics: The Ferroelectric HfO₂-ZrO₂ Superlattice

The critical enabler of this architecture is not the circuit design, but the material science. Traditional ferroelectrics (like Perovskites/PZT) are incompatible with modern silicon manufacturing. Dr. Tour's team utilizes Hafnium Zirconium Oxide (HZO) in a superlattice configuration, a sophisticated nanostructure that solves the historic stability problems of ferroelectric hafnia.

### 3.1 Ferroelectricity in Hafnium Oxide

Ferroelectricity is the property of a material to possess a spontaneous electric polarization that can be reversed by an electric field. This switchable polarization is the mechanism for storing information. Hafnium Oxide (HfO₂) is the standard gate dielectric in all modern transistors (CMOS compatible). However, pure HfO₂ crystallizes in a stable monoclinic phase, which is not ferroelectric. To induce ferroelectricity, it must be forced into a metastable orthorhombic (Pca2₁) phase.

### 3.2 The Superlattice Innovation

The "IronLattice" technology leverages a superlattice structure—alternating atomic-scale layers of HfO₂ and ZrO₂ (e.g., 2nm HfO₂ / 2nm ZrO₂). This layering provides several critical advantages over solid-solution alloys:

#### 3.2.1 Strain Engineering and Phase Stabilization

In a solid solution (random mix of Hf and Zr atoms), the crystal phase is difficult to control precisely. In a superlattice, the interfaces between the HfO₂ and ZrO₂ layers introduce lattice mismatch strains. This mechanical confinement "locks" the crystal lattice into the desired ferroelectric orthorhombic phase, preventing it from relaxing back into the non-functional monoclinic phase. This ensures that the device retains its memory properties even when scaled down to nanometer dimensions.

#### 3.2.2 Eliminating the "Wake-up" Effect

A major drawback of traditional ferroelectric HfO₂ is the "wake-up" effect, where the device must be cycled thousands of times before it achieves its full polarization. This is due to the slow redistribution of oxygen vacancies. The IronLattice superlattice structure, with its frequent interfaces, inhibits the long-range diffusion of defects. Research indicates that these superlattice devices exhibit robust ferroelectricity from the very first cycle, eliminating the need for wake-up cycling and simplifying the manufacturing test process.

#### 3.2.3 Endurance and Fatigue Recovery

"Fatigue" refers to the loss of switchable polarization after millions of cycles. In standard devices, charge trapping at grain boundaries causes this failure. The superlattice structure has been shown to reduce defect density and charge trapping significantly. Jaeho Shin and collaborators have demonstrated that these structures can endure over 10⁹ switching cycles—orders of magnitude higher than Flash memory—and possess unique "fatigue recovery" capabilities where a rest period or a specific voltage pulse can heal the device.

### 3.3 Manufacturing: Flash-Within-Flash Synthesis

While the superlattice structure implies precision deposition (like Atomic Layer Deposition, ALD), the Tour group is famous for "Flash Joule Heating" (research sample). Snippets indicate that the group has developed a "Flash-within-Flash" technique to synthesize high-purity chalcogenides and potentially oxide precursors rapidly. While the primary superlattice devices for chips are likely made via ALD for precision, the research sample method may be utilized for creating bulk source materials or potentially for novel 2D ferroelectric layers (like Indium Selenide) that could be integrated into future iterations of the IronLattice technology.

---

## 4. IronLattice: The Entity and the Innovation

The transition from academic breakthrough to commercial product is facilitated by the startup company IronLattice, spun out of external research institution.

### 4.1 Company Profile

| Attribute | Details |
|-----------|---------|
| **Name** | IronLattice |
| **Origin** | external research institution, Tour Lab (Department of Chemistry/Nanoengineering) |
| **Core Technology** | Neuromorphic AI computing devices based on ferroelectric superlattice structures |
| **Leadership** | Led by Jaeho Shin, a postdoctoral researcher with over a decade of expertise in semiconductor device fabrication and neuromorphic architectures. Advised by Tawfik Jarjour, a Rice alumnus with extensive industry experience in semiconductor manufacturing. |
| **Funding Status** | Recipient of the external research institution "One Small Step" Grant (Cycle 4, 2025), securing non-dilutive capital to advance the Technology Readiness Level (TRL) from lab prototype to commercial viability. |

### 4.2 The "IronLattice" Device Class

The term "IronLattice" is likely a double entendre: referencing the "ferro" (iron-like) magnetic hysteresis behavior of the material and the rigid crystal lattice of the oxide superstructures. The specific device is described as a superlattice-based ferroelectric device enabling analog, non-volatile in-memory computation.

Unlike purely digital memory companies, IronLattice is explicitly targeting the neuromorphic and analog compute space. This suggests the commercial product will be an IP block (Intellectual Property) or a standalone accelerator chip designed to be integrated with standard CMOS logic, providing a dedicated "AI co-processor" functionality that handles the heavy lifting of matrix math at a fraction of the power of the main CPU.

---

## 5. Competitive Landscape and Comparative Analysis

The AI hardware landscape is crowded and fiercely competitive. To understand IronLattice's value proposition, we must benchmark it against both traditional memories and emerging AI accelerators.

### 5.1 Comparative Matrix

| Feature | IronLattice (Fe-Superlattice) | NAND Flash | DRAM | Google TPU (Digital) | Intel Loihi 2 | Mythic.AI |
|---------|------------------------------|------------|------|---------------------|---------------|-----------|
| **Physics Principle** | Ferroelectric Polarization | Charge Trapping (Floating Gate) | Capacitor Charge | Digital Logic (Systolic Array) | Spiking Neural Network (Digital) | Flash-based Analog Compute |
| **Write Speed** | < 100 ns | Slow (~100 μs) | Fast (~10 ns) | N/A | Fast | Slow |
| **Write Voltage** | Low (~1-3 V) | High (~10-20 V) | Low | Low | Low | High |
| **Endurance** | High (>10⁹) | Low (10³-10⁵) | Infinite | N/A | Infinite | Low (Flash limits) |
| **Read/Compute Power** | Ultra-Low (Field Driven) | Low | High (Refresh + Bus) | High (Data Movement) | Low (Event Driven) | Low |
| **CMOS Compatibility** | Native (BEOL) | Native | Native | Native | Native | Native (Legacy nodes) |
| **Volatility** | Non-Volatile | Non-Volatile | Volatile | N/A | Volatile (SRAM) | Non-Volatile |

### 5.2 Detailed Analysis of Competitors

#### 5.2.1 NAND Flash (Storage)

NAND Flash is the incumbent non-volatile memory. While dense and cheap, it relies on trapping electrons through a thick oxide, which requires high voltages (>10V) and causes physical damage to the insulator over time (wear-out).

**IronLattice Advantage:** IronLattice uses ferroelectric switching, which is purely electrostatic. It moves atoms slightly within the unit cell rather than forcing electrons through an insulator. This results in switching energies orders of magnitude lower than Flash and endurance ratings millions of times higher. For AI training, where weights are updated frequently, Flash is unusable; IronLattice is viable.

#### 5.2.2 DRAM (Working Memory)

DRAM is fast but volatile; it loses data when power is cut. It also requires constant "refresh" cycles, consuming significant static power.

**IronLattice Advantage:** IronLattice is non-volatile. An edge device (like a smart sensor) can power down completely to save battery, wake up instantly, and resume processing without reloading data from disk. DRAM cannot do this.

#### 5.2.3 Google TPU (Digital Accelerator)

The TPU is the state-of-the-art for digital AI. It is highly optimized but still adheres to the von Neumann separation of HBM (High Bandwidth Memory) and compute cores.

**IronLattice Advantage:** The TPU burns kilowatts of power moving data back and forth. IronLattice performs the math inside the memory array, theoretically offering 10-100x better TOPs/Watt (Tera-Operations per Second per Watt) efficiency for inference workloads.

#### 5.2.4 Intel Loihi 2 (Neuromorphic)

Loihi is a digital implementation of a Spiking Neural Network (SNN). It simulates neurons using SRAM and logic gates.

**IronLattice Advantage:** Loihi requires multiple transistors (6-12) to store a single bit of SRAM, and many more to simulate a neuron. IronLattice can store a multi-bit weight in a single ferroelectric device. This density advantage allows IronLattice to pack larger models into a smaller physical footprint.

#### 5.2.5 Mythic.AI (Analog Flash)

Mythic.AI also pursues analog compute-in-memory but uses older Flash memory technology.

**IronLattice Advantage:** Mythic is bound by the limitations of Flash—high write voltage and low endurance. This limits Mythic primarily to inference-only applications where weights are static. IronLattice's high endurance (10⁹+ cycles) opens the door for on-chip training and continuous learning, where the AI model updates itself in the field—a capability Mythic struggles to support.

---

## 6. Strategic Applications in Edge and Cloud Computing

The unique combination of non-volatility, analog precision, and extreme energy efficiency positions IronLattice to disrupt several key markets.

### 6.1 Edge Computing and IoT: The Primary Frontier

This is the most immediate application. Billions of IoT devices (cameras, wearables, industrial sensors) collect vast amounts of data but lack the power to process it.

**Use Case:** A battery-powered security camera that uses an IronLattice chip to identify faces locally. Because the chip is non-volatile, it sleeps at near-zero power and wakes up only when an event is detected.

**Energy Impact:** The 90% energy reduction allows such devices to run for months on a coin cell battery rather than days.

### 6.2 Data Center Inference

While training (creating the model) is computationally intensive, inference (using the model) accounts for the majority of data center cycles.

**Use Case:** Large Language Model (LLM) inference. Serving millions of ChatGPT queries requires massive memory bandwidth. IronLattice cards could replace GPU inference farms, utilizing the "in-memory" architecture to deliver answers with significantly lower latency and electricity costs.

### 6.3 Neuromorphic and "Spiking" Systems

The analog nature of ferroelectric conductance mimics the biological synapse's plasticity.

**Use Case:** Autonomous robotics. Robots need to adapt to new environments (one-shot learning) rather than relying on pre-trained static models. IronLattice allows for "Hebbian learning" (neurons that fire together, wire together) to be implemented directly in hardware, enabling robots to learn movement patterns in real-time.

### 6.4 Radiation-Hardened Environments (Space)

Ferroelectric materials are inherently resistant to ionizing radiation, unlike charge-based Flash or DRAM which suffer from bit-flips in space.

**Use Case:** Satellite constellations (like Starlink) processing imagery in orbit. IronLattice provides the density of Flash with the radiation tolerance required for long-term orbital missions.

---

## 7. Avenues for Computational Physics Contribution

For a researcher with a background in GPU acceleration and computational physics, the nascent field of ferroelectric CiM offers fertile ground for high-impact contributions. The interaction between the microscopic quantum mechanical effects of the superlattice and the macroscopic circuit behavior creates complex simulation challenges.

### 7.1 Simulation of Ferroelectric Device Behavior

The switching of polarization in HZO superlattices is not instant; it involves the nucleation and growth of domains.

**Contribution:** Developing Phase-Field Models utilizing the Time-Dependent Ginzburg-Landau (TDGL) equations. These simulations model how polarization domains flip under an electric field and how grain boundaries in the superlattice impede this motion. Implementing these solvers on GPUs (using CUDA) is essential because simulating a realistic 3D superlattice with millions of mesh points is computationally prohibitive on CPUs.

### 7.2 GPU-Accelerated Architectural Modeling

Standard circuit simulators (like SPICE) cannot handle the scale of a full AI chip (millions of devices).

**Contribution:** Creating behavioral models (e.g., in Verilog-A or Python/PyTorch) that abstract the complex physics of the IronLattice device into a "synaptic" transfer function. You could build a GPU-accelerated emulator that runs a neural network on a "virtual" IronLattice chip, injecting realistic noise, hysteresis, and variability derived from your physics models. This "hardware-software co-design" is critical for verifying that a neural network will actually work on the physical chip before it is manufactured.

### 7.3 Visualization of Device Physics

**Contribution:** The concept of a "superlattice" and "strain engineering" is abstract. High-fidelity 3D visualization of the atomic lattice, showing the oxygen vacancy migration paths and the ferroelectric dipole flipping mechanism, would be invaluable for both debugging the device physics and communicating the innovation to investors and partners.

### 7.4 Benchmarking and "Noise-Aware" Training

**Contribution:** Analog hardware is noisy. You can contribute by developing training algorithms that are "noise-aware." By training a neural network on a GPU that simulates the specific noise profile of the IronLattice device, you can produce AI models that are robust to the hardware's imperfections, maximizing the accuracy of the final system.

---

## 8. The Intersection of Faith, Science, and Origin

Dr. external research group is a singular figure in the scientific community, maintaining a dual identity as a world-class synthetic chemist and an outspoken evangelical Christian apologist. He explicitly rejects the notion that faith and science are non-overlapping magisteria, instead arguing that rigorous scientific inquiry affirms the tenets of his faith.

### 8.1 The "Jesus and Science" Foundation

Dr. Tour's ministry work is organized under the Jesus and Science Foundation (jesusandscience.org). This 501(c)(3) non-profit organization has a stated mission to "introduce the world to salvation in Jesus Christ" and "remove barriers to faith by examining natural evidence."

**Activities:** The foundation funds the production of his podcast ("The Science & Faith Podcast"), his YouTube channel content, and supports humanitarian efforts (providing meals). It serves as the platform for his "Resurrection Challenge" and his critiques of abiogenesis.

### 8.2 The Abiogenesis Critique

Dr. Tour is perhaps the most vocal scientific critic of current Origin of Life (abiogenesis) research. Leveraging his expertise in synthetic organic chemistry, he argues that the creation of a living cell from non-living chemicals via random processes is chemically impossible.

- **The Chemical Hurdle:** He details the immense difficulty of synthesizing basic life building blocks (sugars, amino acids, lipids) in a prebiotic environment. He highlights the "homochirality problem"—life requires exclusively left-handed amino acids, but natural chemistry produces 50/50 mixtures (racemic), which destroy protein function.

- **The Assembly Hurdle:** Even if the components existed, Tour argues that the information required to assemble them into a functional cell—the "Interactome"—cannot arise without an intelligent agent. He compares the probability of a cell self-assembling to a factory spontaneously forming out of a junkyard.

### 8.3 Evidence for the Resurrection

Tour applies a forensic/historical methodology to the resurrection of Jesus. He engages with skeptics by challenging them to examine the historical reliability of the New Testament documents, the empty tomb, and the behavior of the apostles. He asserts that the physical resurrection of Jesus is the most logical deduction from the available historical evidence, framing this belief not as a leap of faith but as a rational conclusion.

### 8.4 Integration in Practice

Tour does not shy away from mixing these worlds. He holds Bible studies with students and peers, offers to meet personally via Zoom with any non-believer to discuss the resurrection, and uses his scientific platform to advocate for a worldview that accommodates both rigorous material science (like IronLattice) and metaphysical truth.

---

## 9. Conclusion

The "IronLattice" technology represents a pivotal moment in the evolution of computing. By synthesizing the durability of superlattice materials with the efficiency of analog physics, Dr. external research group and his team at external research institution have engineered a potential solution to the AI energy crisis. This technology does not merely iterate on the past; it dismantles the Von Neumann bottleneck that has constrained computing for seventy years.

### Key Insights

- **Physics as Compute:** IronLattice marks a transition from simulating math with logic gates to performing math with material physics. This shift is essential for reducing the energy cost of AI by the targeted 90%.

- **Material Mastery:** The innovation is fundamentally one of material science—specifically, the ability to control the phase and strain of HfO₂-ZrO₂ superlattices to achieve stability and endurance that raw materials lack.

- **Commercial Realism:** With the formation of the IronLattice startup, seed funding, and leadership by experienced device physicists like Jaeho Shin, the technology is moving rapidly from academic curiosity to commercial prototype.

- **A Holistic Worldview:** The development of this technology occurs alongside Dr. Tour's vigorous defense of theistic faith, illustrating a career defined by the pursuit of truth in both the physical properties of atoms and the metaphysical questions of existence.

For the computational physicist, IronLattice offers a rich, unmapped territory. The need to simulate, visualize, and optimize these quantum-mechanical devices for macroscopic AI workloads is urgent. As the industry pivots toward "Green AI," the contributions of those who can bridge the gap between the Ginzburg-Landau equations and the PyTorch command line will be instrumental in defining the next era of human computation.

---

## Works Cited

1. Practical changes could reduce AI energy demand by up to 90% | UCL News, https://www.ucl.ac.uk/news/2025/jul/practical-changes-could-reduce-ai-energy-demand-90
2. Life-Cycle Emissions of AI Hardware: A Cradle-To-Grave Approach and Generational Trends - arXiv, https://arxiv.org/html/2502.01671v1
3. At COSM, Sharing Information Is Key to Solving Tech Problems | Science and Culture Today, https://scienceandculture.com/2025/11/at-cosm-sharing-information-is-key-to-solving-tech-problems/
4. New AI chip for in-memory computing - Electronica, https://electronica.de/en/industry-portal/detail/new-ai-chip-for-in-memory-computing.html
5. Brain-Inspired Algorithms May Slash AI Energy Costs | Mirage News, https://www.miragenews.com/brain-inspired-algorithms-may-slash-ai-energy-1590801/
6. Rice Innovation awards fourth cycle of One Small Step Grants, https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants
7. Why In-Memory Computation Is So Important For Edge AI - Semiconductor Engineering, https://semiengineering.com/why-in-memory-computation-is-so-important-for-edge-ai/
8. (PDF) HfO2-ZrO2 Superlattice Ferroelectric Capacitor with Improved Endurance Performance and Higher Fatigue Recovery Capability - ResearchGate, https://www.researchgate.net/publication/357111874_HfO2-ZrO2_Superlattice_Ferroelectric_Capacitor_with_Improved_Endurance_Performance_and_Higher_Fatigue_Recovery_Capability
9. HfO2-ZrO2 Ferroelectric Capacitors with Superlattice Structure: Improving Fatigue Stability, Fatigue Recovery, and Switching Speed - PubMed, https://pubmed.ncbi.nlm.nih.gov/38166401/
10. Enhancing ferroelectric stability: wide-range of adaptive control in epitaxial HfO2/ZrO2 superlattices - PubMed Central, https://pmc.ncbi.nlm.nih.gov/articles/PMC12254504/
11. Ferroelectric HfO2–ZrO2 Multilayers with Reduced Wake-Up | ACS Omega, https://pubs.acs.org/doi/10.1021/acsomega.4c10603
12. HfO 2 –ZrO 2 Ferroelectric Capacitors with Superlattice Structure: Improving Fatigue Stability, Fatigue Recovery, and Switching Speed | Request PDF - ResearchGate, https://www.researchgate.net/publication/377132836_HfO_2_-ZrO_2_Ferroelectric_Capacitors_with_Superlattice_Structure_Improving_Fatigue_Stability_Fatigue_Recovery_and_Switching_Speed
13. Stoichiometric Engineering of Indium Selenide Compounds Realized by Flash-within-Flash with an Arc Welder | Request PDF - ResearchGate, https://www.researchgate.net/publication/398146944_Stoichiometric_Engineering_of_Indium_Selenide_Compounds_Realized_by_Flash-within-Flash_with_an_Arc_Welder
14. The Relentless Genius of external research group | Office of Research | external research institution, https://research.rice.edu/news/relentless-genius-james-tour
15. Rice Research Review | Fall 2025 by external research institution - Issuu, https://issuu.com/riceuniversity/docs/rice_research_review_fall_2025
16. French Team Develops First Hybrid Memory Technology Enabling On-Chip AI Learning and Inference | TechPowerUp Forums, https://www.techpowerup.com/forums/threads/french-team-develops-first-hybrid-memory-technology-enabling-on-chip-ai-learning-and-inference.341384/
17. The future of satellite and mobile networks | MWC Doha 2025 full summit #mwc25, https://www.youtube.com/watch?v=eP39CdkuJ98
18. Ferroelectric Devices, Circuits and Architectures for AI Hardware Design | Stanford Electrical Engineering, https://ee.stanford.edu/event/02-19-2025/ferroelectric-devices-circuits-and-architectures-ai-hardware-design
19. Robust ferroelectricity in HfO2/ZrO2 superlattices and its microscopic... - ResearchGate, https://www.researchgate.net/figure/Robust-ferroelectricity-in-HfO2-ZrO2-superlattices-and-its-microscopic-origin-a-XRD_fig4_393617731
20. AI hardware reimagined for lower energy use - Cornell Chronicle, https://news.cornell.edu/stories/2025/09/ai-hardware-reimagined-lower-energy-use
21. About Us - Jesus and Science, https://jesusandscience.org/about/
22. The Science & Faith Podcast, https://podcasts.apple.com/us/podcast/the-science-faith-podcast/id160942632
23. Jesus And Science Foundation - Full Filing - Nonprofit Explorer - ProPublica, https://projects.propublica.org/nonprofits/organizations/851876814/202521259349302437/full
24. Evolution/Creation - Dr. external research group, https://jmtour.com/evolution-creation/
25. January 29, 2025 - Jesus and Science, https://jesusandscience.org/2025/01/january-29-2025/
