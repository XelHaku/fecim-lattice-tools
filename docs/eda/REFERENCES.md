# Scientific References: FeCIM Design Suite

This tool implements simulation models based on published research. It is **not affiliated with or endorsed by** the Tour Group at external research institution or any other research institution.

## Disclaimer

- All models are based on **published literature**, not validated against actual hardware
- Performance claims (energy, speed, endurance) are **from cited papers**, not independently verified
- This is **educational/research software**, not a production tool

---

## 1. Core Device Physics

*Published research that our simulation models are based on.*

* **30-State FeFET Device:** *"Flash In2Se3 for Neuromorphic Computing"* (Shin, Tour, et al., 2025). [View](https://www.researchgate.net/publication/388360521_Flash_In2Se3_for_neuromorphic_computing)
  * *How we use it:* Our 30-level quantization model is based on this demonstrated capability.

* **Flash Joule Heating Synthesis:** *"Stoichiometric Engineering... by Flash-within-Flash"* (2025).
  * *How we use it:* Referenced for material properties; we do not implement this manufacturing process.

* **HZO Ferroelectrics:** Park et al., Adv. Mater. 2015; Cheema et al., Nature 2020.
  * *How we use it:* Hysteresis model parameters (Pr, Ec, Ps) are from these publications.

---

## 2. Open Source EDA Ecosystem

*Tools we generate files for (not tools we provide).*

* **OpenLane:** *"OpenLANE: The Open-Source Digital ASIC Implementation Flow"* (WOSET, 2020). [View](https://woset-workshop.github.io/PDFs/2020/a21.pdf)
  * *How we use it:* We generate Verilog/DEF files compatible with OpenLane format.

* **OpenROAD:** *"Empowering innovation: OpenROAD and the future of open-source EDA"* (EE World, 2024).
  * *How we use it:* Our DEF files use OpenROAD-compatible syntax.

* **GDSFactory:** *"GDSFactory: Build Better Hardware with Better Software"* (IEEE, 2024).
  * *How we use it:* Referenced for future GDSII generation (not currently implemented).

---

## 3. Compute-in-Memory Research

*Academic tools that inspired our approach (we are not affiliated with these projects).*

* **NeuroSim:** *"NeuroSim V1.5: Improved Software Backbone for Benchmarking CIM Accelerators"* (Georgia Tech, 2025).
  * *How we use it:* Referenced for CIM energy modeling methodology.

* **CiMLoop:** *"CiMLoop: A Flexible, Accurate, and Fast Compute-In-Memory Modeling Tool"* (MIT, 2024). [View](https://arxiv.org/pdf/2405.07259)
  * *How we use it:* Referenced for architectural exploration approach.

* **CINM Compiler:** *"CINM (Cinnamon): A Compilation Infrastructure for Heterogeneous CIM"* (ACM ASPLOS, 2024).
  * *How we use it:* Referenced for weight-to-hardware mapping methodology.

---

## 4. Performance Claims (From Literature)

**Important:** The following claims are from published papers, not independently verified by this project.

| Claim | Source | Our Status |
|-------|--------|------------|
| 30 discrete states | Tour Lab 2024/2025 | Simulated, not validated |
| 87% MNIST accuracy | Tour COSM presentation | Target, not achieved |
| 10^9 endurance cycles | Tour Lab (demonstrated) | Used in models |
| 10^12 endurance cycles | Tour Lab (target) | Not demonstrated |
| 10ns switching | Various HZO papers | Used in models |

---

## 5. Market Context (Opinion Pieces)

*These are opinion articles, not peer-reviewed research.*

* *"The Microchip Era Is About to End"* by George Gilder (WSJ, 2025) - [View](https://www.wsj.com/articles/the-microchip-era-is-about-to-end-wafer-scale-integration-computing-ai-3a9d554a)
  * *Note:* This is an opinion piece about wafer-scale integration trends, not a validation of FeCIM technology.

---

## 6. Latest Research (2024-2026)

*Recently published papers extending the state of the art.*

### 6.1 Ferroelectric CIM Architectures

* **Device-Algorithm Co-Design of FeCIM Annealer:** Qian Y, et al., *DAC 2025*. [View](https://dl.acm.org/doi/10.1109/DAC63849.2025.11133307)
  * *Finding:* 1503-1716× energy reduction vs. state-of-the-art for 3000-node Max-Cut; 98% success rate.

* **C-Nash: Nash Equilibrium Solver:** *DAC 2024*. [View](https://dl.acm.org/doi/10.1145/3649329.3655988)
  * *Finding:* First FeCIM architecture handling mixed strategy Nash Equilibrium.

* **Ferroelectric CIM Annealer:** *Nature Communications 2024*. [View](https://www.nature.com/articles/s41467-024-46640-x)
  * *Finding:* 75% chip size saving with lossless QUBO matrix compression.

* **In-Memory Ferroelectric Differentiator:** *Nature Communications 2025*. [View](https://www.nature.com/articles/s41467-025-58359-4)
  * *Finding:* 40×40 FeCap array for differential computation.

### 6.2 Transformer/LLM on CIM

* **Analog IMC Attention Mechanism:** Leroux et al., *Nature Computational Science 2025*. [View](https://www.nature.com/articles/s43588-025-00854-1)
  * *Finding:* **70,000× energy reduction** and **100× speed-up** vs. GPUs for attention; 1.5B parameter model.

* **Efficient Memristor Accelerator for Transformers:** *Scientific Reports 2024*. [View](https://www.nature.com/articles/s41598-024-75021-z)
  * *Finding:* Self-attention acceleration on memristor crossbar.

### 6.3 HfO₂-ZrO₂ Superlattice Advances

* **Enhanced Ferroelectric Stability in Epitaxial Superlattices:** *Nature Communications 2025*. [View](https://www.nature.com/articles/s41467-025-61758-2)
  * *Finding:* Stable ferroelectricity up to 100nm; >10⁹ cycles; coercive field ~0.85 MV/cm.

* **Tailored Coercive Field and Temperature Reliability:** Lehninger, *Adv. Physics Research 2023*. [View](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/apxr.202200108)
  * *Finding:* Enhanced polarization and improved high-temperature retention.

* **Ultrathin Ferroic HfO₂-ZrO₂ Gate Stack:** *Nature 2022*. [View](https://www.nature.com/articles/s41586-022-04425-6)
  * *Finding:* Scaled to ~20 Å; direct integration onto Si transistors.

* **Roadmap on Ferroelectric Hafnia/Zirconia:** *APL Materials 2023*. [View](https://pubs.aip.org/aip/apm/article/11/8/089201/2908480/Roadmap-on-ferroelectric-hafnia-and-zirconia-based)
  * *Finding:* Comprehensive review of HfO₂/ZrO₂ challenges and opportunities.

### 6.4 FeFET Compact Modeling

* **Temperature and Variability-Aware FeFET Model:** *Solid-State Electronics 2024*. [View](https://www.sciencedirect.com/science/article/abs/pii/S0038110124001035)
  * *Finding:* Preisach-based Verilog-A model with temperature and variability effects; validated for MLC and MAC.

* **Multidomain FeFET Compact Model:** *IEEE TED 2021*. [View](https://ieeexplore.ieee.org/document/9336322/)
  * *Finding:* Nucleation-growth dynamics; charge trapping; channel percolation.

### 6.5 Simulation Tools (2024-2025 Updates)

* **NeuroSim V1.5/V2.1:** Georgia Tech. [View](https://www.frontiersin.org/journals/artificial-intelligence/articles/10.3389/frai.2021.659060/full)
  * *Update:* PyTorch interface; chip-level error <1% after calibration.

* **CrossSim:** Sandia National Labs. [View](https://cross-sim.sandia.gov/)
  * *Finding:* GPU-accelerated crossbar simulator; device/circuit non-ideality modeling.

* **3D_NeuroSim V1.0:** [View](https://github.com/neurosim/3D_NeuroSim_V1.0)
  * *Finding:* Monolithic and heterogeneous 3D integration support.

### 6.6 Industry & Fabrication

* **Ferroelectric Memory Company (FMC) €100M Funding:** *Bloomberg 2025*. [View](https://www.bloomberg.com/news/articles/2025-11-13/memory-chip-startup-raises-100-million-for-energy-saving-tech)
  * *Finding:* Dresden startup commercializing DRAM+ based on ferroelectric HfO₂.

* **IHP Open Source PDK:** [View](https://github.com/IHP-GmbH/IHP-Open-PDK)
  * *Update:* April 2025 tape-out (T586); OpenROAD flow supported; RRAM module available in SG13S.

* **OpenLane 2.0 Release:** Efabless, April 2024. [View](https://www.globenewswire.com/news-release/2024/04/18/2865597/0/en/Efabless-Announces-the-Release-of-the-OpenLane-2-Development-Platform-Transforming-Custom-Silicon-Design-Flows.html)
  * *Finding:* Python-based infrastructure; fully customizable flows.

### 6.7 2D Ferroelectric Materials

* **In₂Se₃ for Neuromorphic Computing:** Shin, Tour et al., *Adv. Electronic Materials 2025*. [View](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aelm.202400603)
  * *Finding:* Flash-within-flash synthesis; gram-scale α-In₂Se₃; robust synaptic behavior.

* **Fully Ferroelectric-Gated 2D CIM:** *Science Advances 2024*. [View](https://www.science.org/doi/10.1126/sciadv.adp0174)
  * *Finding:* 96.36% wafer-scale yield; >10¹² endurance; 99.8% tracking accuracy.

---

## 7. Tour Group Publications (external research institution)

*Papers from Dr. external research group's research group at external research institution directly relevant to FeCIM technology.*

### 7.1 2D Ferroelectric Neuromorphic Devices

* **In₂Se₃ Synthesized by FWF Method for Neuromorphic Computing**
  * *Authors:* Jaeho Shin, Jingon Jang, Chi Hun Choi, Jaegyu Kim, Lucas Eddy, Phelecia Scotland, Lane W. Martin, Yimo Han, James M. Tour
  * *Journal:* Advanced Electronic Materials, 2025, Vol. 11, Issue 5, Article 2400603
  * *Published:* [View](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aelm.202400603)
  * *Preprint:* [ChemRxiv](https://chemrxiv.org/engage/chemrxiv/article-details/659ef4cee9ebbb4db9de84cb)
  * *Key Findings:*
    - Gram-scale synthesis of α-In₂Se₃ crystals using flash-within-flash (FWF) Joule heating
    - 2D ferroelectric semiconductor FET as artificial synaptic device platform
    - Modulates polarization via gate electrical pulses for synaptic behavior
    - Non-volatile memory with ON/OFF states due to ferroelectric characteristics
    - Robust reliability under repeated electrical pulses
  * *How we use it:* Foundation for 30-level quantization and synaptic weight programming models.

### 7.2 Flash Joule Heating Technology

* **Flash-within-Flash Joule Heating (FWF) for Sustainable Synthesis**
  * *Authors:* Tour Group, external research institution
  * *Journal:* Nature Chemistry, August 2024
  * *Source:* [Rice News](https://news.rice.edu/news/2024/new-twist-synthesis-technique-developed-rice-promises-sustainable-manufacturing)
  * *Key Findings:*
    - Gram-scale production of compounds in seconds
    - 50% reduction in energy, water, and greenhouse gas emissions
    - Ideal for semiconductor materials: MoSe₂, WSe₂, α-In₂Se₃
    - Enables synthesis regardless of precursor conductance
  * *How we use it:* Referenced for understanding material synthesis pathways.

### 7.3 IronLattice Startup (Tour Lab Spinout)

* **IronLattice: Superlattice-Based Ferroelectric CIM**
  * *Source:* [Rice Innovation One Small Step Grant, 2025](https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants)
  * *Status:* external research institution spinout, $50,000 grant recipient
  * *Technology:*
    - Superlattice-based ferroelectric device
    - Analog, nonvolatile in-memory computation for AI workloads
    - Claims: "millions × lower energy and million × faster than NAND"
  * *How we use it:* Validates commercial potential of FeCIM technology.

### 7.4 COSM 2025 Presentation

* **"Breakthrough: Radical New AI Hardware Design That Nvidia Can't Ignore"**
  * *Speaker:* external research group
  * *Event:* [COSM 2025, Scottsdale, Arizona](https://cosm.tech/speaker/jim-tour/)
  * *Topics:*
    - Compute-in-memory (CiM) technology overview
    - Elimination of data-transfer bottleneck
    - Energy and processing time savings
    - IronLattice technology demonstration
  * *Video:* Available on COSM YouTube channel

### 7.5 Collaborations

* **Yimo Han Lab (external research institution)**
  * *Focus:* 2D ferroelectric materials, electron microscopy
  * *Publications:* [Han Lab](https://hanlab.blogs.rice.edu/publications/)
  * *Collaboration:* Co-author on In₂Se₃ neuromorphic computing paper
  * *Related Work:* Van der Waals ferroelectric materials mapping ([Nature Communications](https://www.nature.com/articles/s41467-023-42110-y))

### 7.6 Related Superlattice FeFET Research

*Note: These papers are from collaborating institutions, not Tour Group directly.*

* **BEOL-Compatible Superlattice FeFET Analog Synapse**
  * *Authors:* Aabrar et al. (Georgia Tech, Shimeng Yu group)
  * *Journal:* IEEE TED, 2022. [View](https://ieeexplore.ieee.org/document/9691825/)
  * *Finding:* Superlattice FE/DE/FE gate-stack for improved linearity/symmetry of weight updates.

* **A Thousand State Superlattice FeFET Analog Weight Cell**
  * *Authors:* Georgia Tech
  * *Event:* IEEE VLSI Symposium 2022
  * *Finding:* Up to 1,000 conductance states demonstrated.

---

## What We Don't Claim

1. **No affiliation** with external research institution, Tour Lab, or any cited institution
2. **No endorsement** from any researcher or company mentioned
3. **No hardware validation** - all models are simulation-only
4. **No production readiness** - this is educational/research software
5. **No ownership** of FeCIM technology - we implement published concepts
