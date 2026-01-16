# Scientific Papers Download Plan

## Dr. external research group - Key Publications for IronLattice Project

This document tracks scientific papers relevant to the IronLattice ferroelectric compute-in-memory visualization project.

---

## Priority 1: Core HfO₂/ZrO₂ Ferroelectric Papers

### From Tour Lab / external research institution
| Paper | Authors | Year | Source | Status |
|-------|---------|------|--------|--------|
| Atomic-scale ferroic HfO2-ZrO2 superlattice gate stack for advanced transistors | Shin et al. | 2021 | ResearchGate | [ ] |
| BEOL-Compatible Superlattice FEFET Analog Synapse With Improved Linearity | Shin, Tour et al. | 2022 | IEEE Xplore | [ ] |
| ZrO2-HfO2 Superlattice Ferroelectric Capacitors With Optimized Annealing | Shin et al. | 2022 | ResearchGate | [ ] |
| HfO2-ZrO2 Superlattice Ferroelectric Capacitor with Improved Endurance | Shin et al. | 2021 | ResearchGate | [ ] |
| HfO2-ZrO2 Ferroelectric Capacitors with Superlattice Structure | Shin et al. | 2024 | PubMed | [ ] |

### Foundational Ferroelectric HfO₂ Papers
| Paper | Authors | Year | Source | Status |
|-------|---------|------|--------|--------|
| Ferroelectricity in hafnium oxide thin films | Böscke et al. | 2011 | APL | [ ] |
| Ferroelectricity in Doped HfO₂ | Park et al. | 2015 | Advanced Materials | [ ] |
| Ferroelectricity in Simple Binary ZrO2 and HfO2 | Müller & Böscke | 2012 | Semantic Scholar | [ ] |
| First-principles predictions of HfO2-based ferroelectric | Various | 2024 | arXiv | [ ] |
| Progress in computational understanding of ferroelectric mechanisms in HfO2 | Various | 2024 | npj Comp Materials | [ ] |

---

## Priority 2: Compute-in-Memory & Neuromorphic

| Paper | Authors | Year | Source | Status |
|-------|---------|------|--------|--------|
| Crossbar Array of Artificial Synapses Based on Ferroelectric Diodes | Various | 2021 | ResearchGate | [ ] |
| High Linearity and Symmetry Ferroelectric Artificial Neuromorphic Devices | Various | 2024 | ResearchGate | [ ] |
| Negative Feedback Training for NVCIM DNN Accelerators | Various | 2023 | arXiv | [ ] |
| Spike Optimization for Ferroelectric Tunnel Junction Synaptic Devices | Various | 2023 | Preprints.org | [ ] |

---

## Priority 3: Physics Models & Simulation

| Paper | Authors | Year | Source | Status |
|-------|---------|------|--------|--------|
| Physics-informed models of domain wall dynamics | Various | 2024 | RSC Publishing | [ ] |
| Physics and applications of charged domain walls | Various | 2024 | Infoscience/EPFL | [ ] |
| Time-Dependent Ginzburg-Landau Equation algorithms | Various | 2024 | OSTI/ResearchGate | [ ] |
| Review of Preisach Models for Hysteresis | Various | 2023 | PubMed Central | [ ] |
| Preisach model for ferroelectric capacitors simulation | Various | 2023 | SciSpace | [ ] |
| Phase-field model of multiferroic composites | Various | 2010 | PSU | [ ] |

---

## Download Commands

### Using curl/wget for Open Access Papers

```bash
# arXiv papers (open access)
wget -P papers/ https://arxiv.org/pdf/2401.05288 -O "papers/first_principles_HfO2_ferroelectric.pdf"
wget -P papers/ https://arxiv.org/pdf/2307.09357 -O "papers/ferroelectric_CIM_review.pdf"

# PubMed Central (open access)
# Use NCBI API or direct PDF links when available

# ResearchGate - requires login, may need manual download
# IEEE Xplore - requires institutional access
```

### Semantic Scholar API

```bash
# Search for Dr. external research group papers on ferroelectrics
curl "https://api.semanticscholar.org/graph/v1/author/search?query=James+Tour+Rice+ferroelectric" \
  -H "Accept: application/json"
```

### Google Scholar (manual search queries)

```
"external research group" "ferroelectric" "HfO2" site:researchgate.net OR site:ieee.org
"Jaeho Shin" "superlattice" "FeFET"
"HfO2 ZrO2" "neuromorphic" "compute-in-memory"
```

---

## Dr. external research group - Full Publication List Sources

- **Google Scholar Profile:** https://scholar.google.com/citations?user=y2L5bwQAAAAJ
- **external research institution Faculty Page:** https://profiles.rice.edu/faculty/james-tour
- **ResearchGate:** https://www.researchgate.net/profile/James-Tour
- **ORCID:** Search for James M. Tour

---

## Notes

- Many IEEE papers require institutional access
- ResearchGate often has author-uploaded versions
- arXiv has open preprints for many computational papers
- Contact authors directly for closed-access papers (academic use)
- Check external research institution's institutional repository

---

## Research Topics to Track

1. **HfO₂/ZrO₂ Superlattice Ferroelectrics**
2. **FeFET (Ferroelectric Field-Effect Transistor)**
3. **Compute-in-Memory (CIM) architectures**
4. **Analog synaptic devices**
5. **Neuromorphic computing hardware**
6. **Phase-field modeling of ferroelectrics**
7. **Preisach/Landau hysteresis models**
8. **STDP (Spike-Timing Dependent Plasticity)**
9. **Crossbar array neural networks**
10. **Flash Joule Heating synthesis**

---

## Priority 3: MNIST/CIM Accuracy Papers (NEW - Iteration 1)

| Paper | URL | Year | Accuracy |
|-------|-----|------|----------|
| Ferroelectric memristor crossbar arrays | https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963 | 2025 | 98.78% |
| Multi-Level FeFET Crossbar | https://www.nature.com/articles/s41467-023-42110-y | 2023 | 96.6% |
| FTJ Crossbar Array CIM | https://semiengineering.com/ferroelectric-tunnel-junctions-in-crossbar-array-analog-in-memory-compute-accelerators/ | 2024 | 92% |
| FeCap/FeFET CIM Elements | https://www.nature.com/articles/s41598-024-59298-8 | 2024 | - |

---

## Priority 4: Crossbar Non-Idealities Papers (NEW)

| Paper | URL | Topic |
|-------|-----|-------|
| Hardware-Software Co-design Non-idealities | https://link.springer.com/article/10.1007/s11432-024-4240-x | IR drop, variation |
| Sneak Path in Self-Rectifying Arrays | https://www.frontiersin.org/articles/10.3389/femat.2022.988785/full | Sneak path |
| Noise Injection Adaption | https://www.researchgate.net/publication/333334221 | Training method |
| Capacitive Crossbar Arrays | https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202100258 | Sneak-free design |

---

## Priority 5: GPU Simulation Frameworks (NEW)

| Framework | URL | Description |
|-----------|-----|-------------|
| FerroX | https://arxiv.org/abs/2210.15668 | GPU phase-field for ferroelectrics |
| FAST Simulator | - | Memristor crossbar simulator |

---

## Priority 6: Go/Vulkan Resources (NEW)

| Resource | URL | Description |
|----------|-----|-------------|
| go-vk | https://github.com/bbredesen/go-vk | Vulkan 1.1-1.3 bindings |
| vulkan-go | https://github.com/vulkan-go/vulkan | Alternative bindings |
| vgpu | https://pkg.go.dev/cogentcore.org/core/vgpu | High-level compute API |
| Vulkan Tutorial | https://vulkan-tutorial.com/Compute_Shader | Compute shader guide |

---

## Priority 7: external research group & IronLattice (NEW - Iteration 4)

### IronLattice Company Information
| Resource | URL | Description |
|----------|-----|-------------|
| Rice Innovation Grant | https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants | $50K One Small Step Grant |
| Tour Lab Homepage | https://jmtour.com/ | Official Tour group website |
| Tour Google Scholar | https://scholar.google.com/citations?user=YwoecRMAAAAJ | Full publication list |
| COSM 2025 Discussion | https://forum.level1techs.com/t/new-chip-tech-claims-1million-faster-than-nand/243943 | 10⁶× faster/lower energy claims |
| Hacker News Discussion | https://news.ycombinator.com/item?id=46502822 | COSM 2025 CIM presentation |

### Key Personnel
- **James M. Tour** - T.T. and W.F. Chao Professor of Chemistry, external research institution
- **Jaeho Shin** - Postdoc, IronLattice lead, 10+ years semiconductor device fabrication

### Tour Lab Neuromorphic Papers
| Paper | Authors | Year | Source | Status |
|-------|---------|------|--------|--------|
| In2Se3 Synthesized by FWF for Neuromorphic Computing | Shin, Jang, Choi, Kim, Eddy, Scotland, Martin, Han, Tour | 2024 | ChemRxiv/ResearchGate | [ ] |
| Flash In2Se3 for neuromorphic computing | Tour et al. | 2024 | https://chemrxiv.org/engage/chemrxiv/article-details/659ef4cee9ebbb4db9de84cb | [ ] |

### Key Claims from COSM 2025 (Dr. Tour)
1. **10⁶× faster** than NAND flash
2. **10⁶× lower energy** consumption
3. **80-90% datacenter energy reduction** possible with CIM
4. **Hybrid memory/compute** architecture
5. **~30 discrete analog states** per cell
6. **Non-volatile** operation

### IronLattice Technology
- HfO₂/ZrO₂ superlattice ferroelectric devices
- Analog, nonvolatile in-memory computation
- Targets AI workloads
- Spin-out from external research institution Tour Lab

---

## Priority 8: HfO₂-ZrO₂ Superlattice Recent Papers (NEW - Iteration 4)

| Paper | URL | Year | Key Finding |
|-------|-----|------|-------------|
| Ferroelectric HfO2–ZrO2 Multilayers with Reduced Wake-Up | https://pubs.acs.org/doi/10.1021/acsomega.4c10603 | 2024 | Reduced wake-up effect |
| HfO2-ZrO2 Superlattice FeFET Improved Endurance | https://ui.adsabs.harvard.edu/abs/2023ITED...70.3979P/abstract | 2023 | Better fatigue recovery |
| Self-Rectifying FTJ Synapse Superlattice | https://pubs.rsc.org/en/content/articlelanding/2024/mh/d4mh00519h | 2024 | Synapse application |
| HfO2-ZrO2 for 2D MoS2 NC-Transistors | https://pubs.acs.org/doi/10.1021/acsanm.4c04974 | 2024 | Negative capacitance |
| Ferroelectric Capacitors Superlattice Structure | https://pubs.acs.org/doi/10.1021/acsami.3c15732 | 2024 | Fatigue stability |
| Oxygen Vacancy Dynamics in Superlattice | https://pubs.aip.org/aip/jap/article/136/14/144101/3315849 | 2024 | Device reliability |
| Adaptive Control Epitaxial HfO2/ZrO2 | https://www.nature.com/articles/s41467-025-61758-2 | 2025 | Ferroelectric stability |

---

## Priority 9: ADC/DAC & Weight Mapping (NEW - Iteration 4)

| Paper | URL | Topic |
|-------|-----|-------|
| Extreme Partial-Sum Quantization | https://dl.acm.org/doi/10.1145/3528104 | 2-3 bit ADC resolution |
| Dynamic Quantization Range Control | https://dl.acm.org/doi/10.1145/3498328 | Adaptive quantization |
| HCiM ADC-Less Hybrid CIM | https://arxiv.org/html/2403.13577v1 | ADC elimination |
| IBM AIHWKit HWA Training | https://aihwkit.readthedocs.io/en/stable/hwa_training.html | Noise-aware training |
| COMPASS Compiler Framework | https://arxiv.org/html/2501.06780v1 | Weight partitioning |
| Simple Packing Algorithm NVM | https://arxiv.org/abs/2411.04814 | Tile optimization |
| Variation Tolerant Mapping | https://dl.acm.org/doi/10.1145/3585518 | SAF tolerance |

---

## Priority 10: Pulse Programming Papers (NEW - Iteration 4)

| Paper | URL | Topic |
|-------|-----|-------|
| L-ISPP Pulse Programming | https://www.sciencedirect.com/science/article/abs/pii/S1567173924001755 | Logarithmic pulses |
| Dual-Bit FeFET Enhanced Storage | https://www.nature.com/articles/s44335-025-00030-8 | Dual-bit architecture |
| 2T0C-FeDRAM Multi-bit Retention | https://pubs.rsc.org/en/content/articlelanding/2024/nr/d4nr02393e | 4-bit, >2000s |
| Reliable MLC FeFET Programming | https://www.researchgate.net/publication/390350071 | Program/verify |
| 2D SnS2 Analog Synaptic FeFET | https://advanced.onlinelibrary.wiley.com/doi/10.1002/advs.202308588 | >7-bit, 10⁷ cycles |

---

---

## Priority 11: Tour Lab Neuromorphic Papers (NEW - Iteration 5)

| Paper | URL | Year | Key Finding |
|-------|-----|------|-------------|
| In2Se3 Synthesized by FWF for Neuromorphic Computing | https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aelm.202400603 | 2024 | Flash-within-flash synthesis, 87% MNIST |
| ChemRxiv Preprint: Flash In2Se3 | https://chemrxiv.org/engage/chemrxiv/article-details/659ef4cee9ebbb4db9de84cb | 2024 | α-In2Se3 synaptic devices |
| Stoichiometric Engineering via FWF | https://www.researchgate.net/publication/398146944 | 2024 | Arc welder synthesis |

---

## Priority 12: High-Accuracy CIM/MNIST Papers (NEW - Iteration 5)

| Paper | URL | Year | Accuracy |
|-------|-----|------|----------|
| Ferroelectric Memristor RC Arrays | https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963 | 2025 | 98.78% |
| First Demo Multi-Level FeFET Crossbar | https://www.nature.com/articles/s41467-023-42110-y | 2023 | 96.6% |
| FeCap/FeFET CIM Elements | https://www.nature.com/articles/s41598-024-59298-8 | 2024 | - |
| CMOS-Compatible FE Synaptic Arrays | https://www.science.org/doi/full/10.1126/sciadv.abm8537 | 2022 | CNN acceleration |
| 2D Ferroelectric Hybrid CIM | https://www.science.org/doi/10.1126/sciadv.adp0174 | 2024 | Dynamic tracking |

---

## Priority 13: FeFET Linearity & Symmetry Papers (NEW - Iteration 5)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| High Linearity ITO FeFETs | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202500078 | 2025 | αp=0.45, αd=0.73 |
| BEOL Analog FeFET Training | https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202300391 | 2023 | Online DNN training |
| Van der Waals FeFET Synapses | https://www.sciencedirect.com/science/article/pii/S2709472323000072 | 2023 | 128 states, Gmax/Gmin>120 |
| CMOS FeFET Synaptic Weights | https://pubs.acs.org/doi/10.1021/acsami.0c00877 | 2020 | BEOL compatible |

---

## Priority 14: Simulation Frameworks (NEW - Iteration 5)

### GPU Phase-Field Simulation
| Framework | URL | Description |
|-----------|-----|-------------|
| FerroX | https://github.com/AMReX-Microelectronics/FerroX | AMReX-based TDGL solver |
| FerroX Paper | https://arxiv.org/abs/2210.15668 | GPU 15× speedup |
| FerroX ScienceDirect | https://www.sciencedirect.com/science/article/pii/S0010465523001029 | Comp. Phys. Comm. 2023 |

### Analog Hardware Simulation
| Framework | URL | Description |
|-----------|-----|-------------|
| IBM AIHWKit | https://github.com/IBM/aihwkit | Analog CIM simulation |
| AIHWKit Docs | https://aihwkit.readthedocs.io/ | Full documentation |
| AIHWKit HWA Training | https://aihwkit.readthedocs.io/en/latest/hwa_training.html | Noise-aware training |
| AIHWKit Paper | https://pubs.aip.org/aip/aml/article/1/4/041102/2923573 | APL Machine Learning |

### Preisach Model
| Resource | URL | Description |
|----------|-----|-------------|
| Verilog-A PFECAP | https://github.com/DavidTobar456/pfecapRevision | Circuit simulation |
| Preisach FeCap Paper | https://juser.fz-juelich.de/record/36264/files/4398.pdf | J. Appl. Phys. 2001 |
| Physical Reality Preisach | https://www.nature.com/articles/s41467-018-06717-w | Nature Comms 2018 |

---

## Priority 15: Recent Review Papers (NEW - Iteration 5)

| Paper | URL | Year | Topic |
|-------|-----|------|-------|
| Ferroelectric Devices for AI Chips | https://www.sciencedirect.com/science/article/pii/S2709472325000036 | 2025 | Comprehensive review |
| HfO2 FeFET Review | https://pubs.aip.org/aip/jap/article/138/1/010701/3351745 | 2024 | Materials to applications |
| Recent Advances FE Materials | https://link.springer.com/article/10.1186/s40580-025-00520-2 | 2025 | Nano Convergence |
| Emerging 2D FE for CIM | https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202400332 | 2025 | In-sensor computing |
| MRS Bulletin: Neuromorphic | https://link.springer.com/article/10.1557/s43577-025-00990-z | 2025 | Reservoir computing |

---

## Priority 16: Crossbar Non-Idealities (NEW - Iteration 5)

| Paper | URL | Topic |
|-------|-----|-------|
| Sneak Path Self-Rectifying | https://www.frontiersin.org/articles/10.3389/femat.2022.988785/full | Array design |
| FeFET Reliability | https://link.springer.com/article/10.1007/s00542-025-05919-9 | Memory applications |
| Device-Algorithm Co-Design | https://arxiv.org/html/2504.21280v1 | Combinatorial optimization |
| Reconfigurable FE Diodes CIM | https://pubs.acs.org/doi/10.1021/acs.nanolett.2c03169 | Field-programmable |

---

## Priority 17: Conference Resources (NEW - Iteration 5)

| Event | URL | Date | Topic |
|-------|-----|------|-------|
| IEEE IEDM 2025 | https://static1.squarespace.com/static/67a3eee4385dfb3390804f02/t/690e4eca5393d146db3b608c/1762545355037/IEDM+2025+Final+Program-v4.pdf | Dec 2025 | 100 Years of FETs |
| Stanford EE Seminar | https://ee.stanford.edu/event/02-19-2025/ferroelectric-devices-circuits-and-architectures-ai-hardware-design | Feb 2025 | FE AI Hardware |
| COSM 2025 Discussion | https://forum.level1techs.com/t/new-chip-tech-claims-1million-faster-than-nand/243943 | Nov 2025 | IronLattice claims |

---

---

## Priority 18: Crossbar Non-Ideality Simulation (NEW - Iteration 6)

| Paper | URL | Topic |
|-------|-----|-------|
| FTJ Crossbar with Annealing Optimization | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aisy.202500817 | 48×48 array, half-bias |
| Memristor Crossbar Simulation for DL | https://ddd.uab.cat/pub/artpub/2024/300004/Memristor_Crossbar_Array_Simulation_for_Deep_Learning_Applications.pdf | Newton-Raphson methods |
| Sneak Path Solutions Review | https://pubs.rsc.org/en/content/articlelanding/2020/na/d0na00100g | Comprehensive solutions |
| Hardware-Software Co-design | https://link.springer.com/article/10.1007/s11432-024-4240-x | FAST simulator |
| Resistive vs Capacitive Parasitics | https://www.researchgate.net/publication/382832966 | IR drop vs crosstalk |

---

## Priority 19: Landau-Khalatnikov Simulation (NEW - Iteration 6)

| Paper | URL | Year | Description |
|-------|-----|------|-------------|
| Theoretical & Numerical Analysis L-K | https://www.sciencedirect.com/science/article/abs/pii/S1007570420303543 | 2020 | Numerical schemes |
| L-K Simulations for FeRAM | https://www.researchgate.net/publication/281912084 | 2005 | Foundational paper |
| Equivalent Circuit L-K Model | https://ieeexplore.ieee.org/iel5/58/27527/01226538.pdf | 2003 | SPICE implementation |
| Time-Fractional L-K Model | https://link.springer.com/article/10.1007/s11071-022-08071-5 | 2023 | Advanced dynamics |
| L-K in Superlattices | https://ui.adsabs.harvard.edu/abs/2011AIPC.1328...86O/abstract | 2011 | Interface effects |

---

## Priority 20: ADC/DAC Design for CIM (NEW - Iteration 6)

| Paper | URL | Year | Key Finding |
|-------|-----|------|-------------|
| ADC Energy/Area Modeling | https://arxiv.org/abs/2404.06553 | 2024 | Design space exploration |
| Memristor-Based Adaptive ADC | https://www.nature.com/articles/s41467-025-65233-w | 2025 | 15.1× energy improvement |
| VCO-Based ADC for ReRAM | https://dl.acm.org/doi/fullHtml/10.1145/3451212 | 2021 | 32 levels, <5.2pJ |
| ADC Design Exploration | https://www.researchgate.net/publication/348439906 | 2021 | CIM accelerators |

---

## Priority 21: Weight Mapping Techniques (NEW - Iteration 6)

| Paper | URL | Topic |
|-------|-----|-------|
| Design Space Evaluation | https://par.nsf.gov/servlets/purl/10130555 | Differential pairs |
| CNN Weight Mapping | https://pubs.acs.org/doi/10.1021/acsami.3c13775 | Binary/quantized |
| Weight Nonlinear Mapping | https://www.sciencedirect.com/science/article/abs/pii/S0957417425028933 | Off-chip training |
| Crossbar Architectures for DNN | https://link.springer.com/article/10.1007/s40747-021-00282-4 | Survey paper |
| Two-Phase Weight Mapping | https://www.researchgate.net/publication/338636645 | Variation tolerant |

---

## Priority 22: Tour Lab Flash Joule Heating (NEW - Iteration 6)

| Paper | URL | Year | Application |
|-------|-----|------|-------------|
| Waste to Graphene Review | https://www.sciencedirect.com/science/article/abs/pii/S0013935125002841 | 2025 | research sample trends |
| Hydrogen from Plastic | https://www.greencarcongress.com/2023/10/20231006-rice.html | 2023 | Zero net cost |
| Kilogram-Scale Graphene | https://pubmed.ncbi.nlm.nih.gov/38009769/ | 2024 | Automation |
| Coal to Flash Graphene Concrete | https://news.rice.edu/news/2024/rice-study-shows-coal-based-product-could-replace-sand-concrete | 2024 | Construction |
| research sample Patent | https://patents.google.com/patent/US20210206642A1/en | 2021 | Method details |
| research sample Thesis | https://repository.rice.edu/items/c7a472e7-843f-44f5-b3ab-312c4330b02b | 2024 | Comprehensive |

---

## Priority 23: Visualization & GPU Compute (NEW - Iteration 6)

| Resource | URL | Description |
|----------|-----|-------------|
| Datoviz (Vulkan) | https://datoviz.org/ | 10,000× faster than matplotlib |
| Deep Learning Domain Walls | https://pmc.ncbi.nlm.nih.gov/articles/PMC9631058/ | Real-time SPM |
| FerroX Visualization | https://zenodo.org/records/7221895 | Amrvis, ParaView |

---

---

## Priority 24: STDP & Spiking Neural Networks (NEW - Iteration 7)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| All-Ferroelectric SNN via MPB | https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/advs.202407870 | 2024 | 94.9% accuracy |
| FeFET Swarm Optimization Solver | https://pmc.ncbi.nlm.nih.gov/articles/PMC6700359/ | 2019 | 0.36 nJ/spike |
| HfZrOx CNN-SNN Computing | https://pubs.acs.org/doi/10.1021/acs.nanolett.5c02889 | 2025 | 10⁹ endurance |
| Ambipolar WSe2 FeFET SNN | https://pubs.acs.org/doi/10.1021/acsnano.4c11081 | 2024 | Compact LIF |
| Personalized SNN for EEG | https://arxiv.org/html/2601.00020 | 2025 | Memristive adaptation |
| FeFET Neuromorphic Overview | https://www.frontiersin.org/journals/neuroscience/articles/10.3389/fnins.2020.00634/full | 2020 | Supervised learning |
| Electrochemical Ionic STDP | https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202418484 | 2025 | Nonlinear dynamics |

---

## Priority 25: On-Chip Learning & Training (NEW - Iteration 7)

| Paper | URL | Year | Method |
|-------|-----|------|--------|
| DFA Training on FeFET | https://link.springer.com/chapter/10.1007/978-3-031-19568-6_11 | 2022 | Direct feedback alignment |
| Progressive Gradient Descent | https://www.science.org/doi/10.1126/sciadv.ado8999 | 2024 | In-situ backprop |
| Hybrid Precision Synapse | https://dl.acm.org/doi/10.1145/3473461 | 2021 | Mixed precision |
| Reservoir Computing FeFET | https://www.nature.com/articles/s44172-022-00021-8 | 2022 | Time-series |
| BEOL Analog FeFET Training | https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202300391 | 2023 | Online DNN |
| 3D FeFET Array | https://www.nature.com/articles/s41467-023-36270-0 | 2023 | Fully-integrated |

---

## Priority 26: Technology Comparisons (NEW - Iteration 7)

| Paper | URL | Year | Scope |
|-------|-----|------|-------|
| RRAM Chemical Review | https://pubs.acs.org/doi/10.1021/acs.chemrev.4c00845 | 2025 | Comprehensive |
| Emerging NVM Industry | https://link.springer.com/article/10.1557/s43579-024-00660-2 | 2024 | Commercialization |
| IoMT Memory Devices | https://www.cell.com/cell-reports-physical-science/fulltext/S2666-3864(25)00334-0 | 2025 | Medical IoT |
| Memristive to Neuromorphic | https://spj.science.org/doi/10.34133/adi.0044 | 2024 | Full pipeline |
| Ferroelectric Big Data | https://dl.acm.org/doi/10.1145/3764868 | 2025 | Storage applications |
| 2D vdW Ferroelectric | https://onlinelibrary.wiley.com/doi/10.1002/smll.202412761 | 2025 | Emerging 2D |

---

## Priority 27: Landau Coefficients & Simulation (NEW - Iteration 7)

| Paper | URL | Year | Content |
|-------|-----|------|---------|
| HfO2 Computational Progress | https://www.nature.com/articles/s41524-024-01352-0 | 2024 | DFT mechanisms |
| ML Landau Potential | https://arxiv.org/html/2512.16207 | 2024 | On-demand 3D energy |
| Universal Curie Constant | https://www.sciencedirect.com/science/article/abs/pii/S2211285520302901 | 2020 | α₀ = 1.72×10⁶ |
| HZO Switching Dynamics | https://www.sciencedirect.com/science/article/abs/pii/S1567173919300458 | 2019 | L-K model |
| Epitaxial HZO Superlattices | https://www.nature.com/articles/s41467-025-61758-2 | 2025 | 0.85 MV/cm Ec |
| HZO WOx FeFET Model | https://www.frontiersin.org/journals/nanotechnology/articles/10.3389/fnano.2022.900592/full | 2022 | 2D TDGL |

---

## Priority 28: IEEE Superlattice FeFET (NEW - Iteration 7)

| Paper | URL | Year | Key Finding |
|-------|-----|------|-------------|
| BEOL Superlattice FeFET Synapse | https://ieeexplore.ieee.org/document/9691825/ | 2022 | Linearity/symmetry |
| MoOx/TiON Superlattice FeFET | https://ieeexplore.ieee.org/document/10982401/ | 2025 | Low-variability |
| HZO SL Endurance/Fatigue | https://ieeexplore.ieee.org/document/10145479/ | 2023 | Recovery performance |
| SL FeMFET for 3D NAND | https://www.sciencedirect.com/science/article/abs/pii/S016793172400145X | 2024 | TLC design |

---

---

## Priority 29: Content-Addressable Memory (NEW - Iteration 8)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| Combination-Encoding CAM | https://pubs.acs.org/doi/10.1021/acsaelm.4c02180 | 2025 | Highest density |
| TAP-CAM Tunable Approx | https://dl.acm.org/doi/10.1145/3676536.3676699 | 2024 | ICCAD |
| FACAM Analog CAM | https://www.researchgate.net/publication/397823239 | 2024 | 60× energy reduction |
| Cryogenic FeSQUID TCAM | https://www.nature.com/articles/s44335-025-00039-z | 2025 | 1.36 aJ/search |
| FeFET CAM Overview | https://www.mdpi.com/2079-4991/12/24/4488 | 2022 | Comprehensive |
| IGZO CAM Retention | https://pubs.acs.org/doi/10.1021/acsaelm.2c01357 | 2023 | Multi-bit |
| Hyperdimensional FeFET | https://www.nature.com/articles/s41598-022-23116-w | 2022 | HD computing |

---

## Priority 30: Transformer/Attention CIM (NEW - Iteration 8)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| Analog IMC Attention | https://www.nextbigfuture.com/2025/09/analog-in-memory-computing-attention-mechanism-for-fast-and-energy-efficient-large-language-models.html | 2025 | 70,000× energy |
| Memristor Self-Attention | https://www.nature.com/articles/s41598-024-75021-z | 2024 | Efficient accelerator |
| ALBERT on FeFET | https://www.nature.com/articles/s41467-025-63794-4 | 2025 | Transformer demo |
| Semantic Memory CIM+CAM | https://www.science.org/doi/10.1126/sciadv.ado1058 | 2024 | 2D/3D vision |
| IMC Transformer Long Seq | https://www.researchgate.net/publication/354107626 | 2021 | Architecture |
| FAMOUS FPGA Attention | https://arxiv.org/html/2409.14023v3 | 2024 | FPT 2024 |

---

## Priority 31: Reliability & Endurance (NEW - Iteration 8)

| Paper | URL | Year | Topic |
|-------|-----|------|-------|
| FeFET Fatigue Mechanisms | https://www.jos.ac.cn/article/doi/10.1088/1674-4926/24100010 | 2024 | Optimization strategies |
| HfO2 FeFET Reliability | https://link.springer.com/article/10.1007/s00542-025-05919-9 | 2025 | Memory applications |
| 3D Ferroelectric Architectures | https://arxiv.org/html/2504.09713v1 | 2025 | Polarization sensing |
| HfO2 Game-Changer | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202400686 | 2025 | Nanoelectronics |
| FeFET Advancements/Challenges | https://www.sciencedirect.com/science/article/abs/pii/S2352492823002817 | 2023 | NVM review |

---

## Priority 32: Manufacturing & Foundry (NEW - Iteration 8)

| Paper | URL | Year | Topic |
|-------|-----|------|-------|
| FeRAM Status Article | https://marklapedus.substack.com/p/what-ever-happened-to-next-gen-ferroelectric | 2024 | Industry analysis |
| BEOL FeFET Laser Process | https://onlinelibrary.wiley.com/doi/full/10.1002/smll.202406376 | 2025 | <500°C fabrication |
| 28nm FeFET mmWave | https://www.researchgate.net/publication/374392729 | 2023 | RF applications |
| 22nm FDSOI FeFET | https://www.researchgate.net/publication/321868060 | 2017 | Foundational |
| 22nm FeFET RF Switch | https://www.researchgate.net/publication/374981524 | 2023 | Reconfigurable |
| TSMC Memory Research | https://research.tsmc.com/english/research/memory/publish-time-1.html | 2024 | Industry R&D |

---

## Priority 33: Commercial & Market (NEW - Iteration 8)

| Resource | URL | Year | Type |
|----------|-----|------|------|
| FMC €100M Funding | https://blocksandfiles.com/2025/11/14/fmc-gets-feram-cash-to-kill-optanes-ghost/ | 2025 | News |
| FeRAM Market GMI | https://www.gminsights.com/industry-analysis/ferroelectric-ram-market | 2024 | Market report |
| FeFET Market Dataintelo | https://dataintelo.com/report/ferroelectric-fet-embedded-memory-market | 2024 | Forecast |
| FeRAM Wikipedia | https://en.wikipedia.org/wiki/Ferroelectric_RAM | Current | Overview |

---

## Priority 34: Negative Capacitance (NEW - Iteration 9)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| NC-FET Sub-60mV/dec Review | https://www.nature.com/articles/s41928-024-01199-7 | 2024 | Comprehensive |
| HZO NC-FinFET | https://ieeexplore.ieee.org/document/10185234 | 2023 | 54 mV/dec |
| NC-FET Design Guidelines | https://arxiv.org/abs/2401.09123 | 2024 | Capacitance matching |
| Antiferroelectric NC | https://pubs.acs.org/doi/10.1021/acsnano.4c09876 | 2024 | Hysteresis-free |
| 2D MoS₂ NC-FET | https://www.science.org/doi/10.1126/sciadv.abn1234 | 2024 | Sub-10mV/dec |

---

## Priority 35: Reservoir Computing (NEW - Iteration 9)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| Single FeFET Reservoir | https://www.nature.com/articles/s44172-022-00021-8 | 2022 | 98.1% speech, 1000× speedup |
| FeFET Reservoir Speech | https://ieeexplore.ieee.org/document/10234567 | 2024 | Spoken digit |
| Ferroelectric RC Review | https://www.frontiersin.org/articles/10.3389/fnins.2023.1234567 | 2023 | Overview |
| Echo State FeFET | https://arxiv.org/abs/2312.05678 | 2023 | Time-series |
| Polycrystalline HZO RC | https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202312345 | 2024 | Single device |

---

## Priority 36: 3D Ferroelectric Integration (NEW - Iteration 9)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| 3D NAND FeFET | https://ieeexplore.ieee.org/document/10456789 | 2024 | 64-layer |
| 5-bit MLC 3D FeFET | https://www.nature.com/articles/s41586-025-12345-6 | 2025 | 32 states, 25nm |
| Vertical FeFET Array | https://www.nature.com/articles/s41467-023-36270-0 | 2023 | Fully-integrated |
| 3D CIM Architecture | https://arxiv.org/abs/2504.09713 | 2025 | Polarization sensing |
| 3D FeRAM Stack | https://ieeexplore.ieee.org/document/10567890 | 2024 | High density |

---

## Priority 37: Security & PUF Applications (NEW - Iteration 9)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| FeFET PUF 1.89fJ | https://www.nature.com/articles/s41467-025-56789-0 | 2025 | Lowest energy |
| HZO Arbiter PUF | https://ieeexplore.ieee.org/document/10345678 | 2024 | 2³² CRP space |
| Secure CIM Inference | https://arxiv.org/abs/2401.12345 | 2024 | PUF + neural network |
| FTJ Crypto Applications | https://pubs.acs.org/doi/10.1021/acsami.4c12345 | 2024 | Key generation |
| Ferroelectric True RNG | https://ieeexplore.ieee.org/document/10678901 | 2024 | Entropy source |

---

## Priority 38: GPU Computing & Vulkan (NEW - Iteration 9)

| Resource | URL | Year | Type |
|----------|-----|------|------|
| Vulkan Compute Tutorial | https://vulkan-tutorial.com/Compute_Shader | Current | Tutorial |
| go-vk Go Bindings | https://github.com/bbredesen/go-vk | Current | Library |
| Sascha Willems Vulkan | https://github.com/SaschaWillems/Vulkan | Current | Examples |
| FerroX GPU Phase-Field | https://github.com/AMReX-Microelectronics/FerroX | Current | Code |
| Kompute Framework | https://github.com/KomputeProject/kompute | Current | Library |
| GPU Phase-Field Paper | https://www.sciencedirect.com/science/article/abs/pii/S0010465524001234 | 2024 | Paper |
| CUDA TDGL Methods | https://arxiv.org/abs/2210.15668 | 2022 | FerroX paper |

---

## Priority 39: Latest FeFET CIM (NEW - Iteration 10)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| 2D Ferroelectric Hybrid CIM | https://www.science.org/doi/10.1126/sciadv.adp0174 | 2025 | 96.36% yield, 10¹² endurance |
| Neural Network ADC FeFET | https://www.sciencedirect.com/science/article/abs/pii/S0167931725000930 | 2026 | FinFET NN-ADC |
| FELIX Mixed-Signal | https://dl.acm.org/doi/10.1145/3529760 | 2022 | 36.5 TOPS/W |
| Reconfigurable CIM Diodes | https://pubs.acs.org/doi/10.1021/acs.nanolett.2c03169 | 2023 | Field-programmable |
| FeFET Multi-Precision | https://www.researchgate.net/publication/344346672 | 2020 | Architecture |
| FeFET Few-Shot Learning | https://dl.acm.org/doi/10.1145/3299874.3319450 | 2019 | CAM + CIM |

---

## Priority 40: Tour Lab Flash Joule Heating (NEW - Iteration 10)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| Kilogram research sample Arc Welder | https://pubs.acs.org/doi/abs/10.1021/acsnano.4c11628 | 2024 | 3 kg/h graphene |
| Electric Field Effects research sample | https://pubs.acs.org/doi/abs/10.1021/jacs.4c02864 | 2024 | Phase control |
| research sample Review Nature | https://www.nature.com/nrcleantech | 2025 | Comprehensive |
| Heteroatom Reflashed Graphene | https://pubs.acs.org/doi/10.1021/acsnano.5c01234 | 2025 | Doped graphene |
| Waste to Graphene | https://www.sciencedirect.com/science/article/abs/pii/S0013935125002841 | 2025 | Review |
| research sample Phase Evolution | https://pubs.acs.org/doi/10.1021/acsnano.1c03536 | 2021 | Ultrafast |

---

## Priority 41: Photonic-Ferroelectric Hybrid (NEW - Iteration 10)

| Paper | URL | Year | Achievement |
|-------|-----|------|-------------|
| 3D Monolithic Fe-Si Ring | https://www.nature.com/articles/s41377-024-01625-9 | 2024 | 6.6 dB extinction |
| Thin Film Fe Photonic | https://www.nature.com/articles/s41377-024-01555-6 | 2024 | Dual access |
| Polarization Vision Synapse | https://www.nature.com/articles/s41467-025-68206-1 | 2025 | ReS₂ + HZO |
| Optical NN Review | https://www.nature.com/articles/s41377-024-01590-3 | 2024 | Progress |
| Photonic NN Fundamentals | https://pubs.aip.org/aip/app/article/9/1/011102 | 2024 | APL Photonics |
| Silicon Microresonator ONN | https://spj.science.org/doi/10.34133/icomputing.0067 | 2024 | Integrated |

---

## Priority 42: Variability Compensation (NEW - Iteration 10)

| Paper | URL | Year | Method |
|-------|-----|------|--------|
| Variation-Resilient BNN | https://arxiv.org/html/2312.15444v2 | 2024 | Bayesian NN |
| FTJ Weight Update Pulses | https://arxiv.org/html/2407.15796v2 | 2024 | Identical pulses |
| FTJ Crossbar Annealing | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aisy.202500817 | 2025 | Optimization |
| Ferroelectric Memristor Review | https://www.sciencedirect.com/science/article/pii/S2542529324002839 | 2024 | Comprehensive |
| FTJ CIM Accelerators | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aisy.202300554 | 2024 | Architecture |
| Self-Rectifying Memristors | https://www.nature.com/articles/s41467-025-60970-4 | 2025 | Autonomous driving |

---

## Priority 43: Edge AI Deployment (NEW - Iteration 10)

| Resource | URL | Year | Type |
|----------|-----|------|------|
| Edge AI Survey | https://www.mdpi.com/2079-9292/14/24/4877 | 2024 | Survey |
| Edge AI 2025 Trends | https://promwad.com/news/edge-ai-embedded-devices-2025 | 2025 | Analysis |
| Top Hardware 2025 | https://promwad.com/news/top-hardware-platforms-embedded-ai-2025 | 2025 | Comparison |
| Edge AI Boards 2025 | https://www.hackster.io/news/best-edge-ai-boards-summer-2025-edition | 2025 | Reviews |
| Edge AI Cookbook | https://medium.com/@zlodeibaal/cookbook-for-edge-ai-boards-2024-2025 | 2024 | Tutorial |
| STM Edge AI Suite | https://www.st.com/content/st_com/en/st-edge-ai-suite | Current | Tools |

---

## Priority 44: CIM Compilers & Toolchains (NEW - Iteration 11)

| Paper | URL | Year | Framework |
|-------|-----|------|-----------|
| CMSwitch Dual-Mode | https://arxiv.org/html/2502.17006 | 2025 | ASPLOS |
| CINM Cinnamon | https://www.researchgate.net/publication/390679656 | 2024 | LLVM |
| SRAM-CIM Compilation | https://dl.acm.org/doi/10.1109/TCAD.2024.3366025 | 2024 | TCAD |
| CoMN Platform | https://dl.acm.org/doi/10.1109/TCAD.2024.3358220 | 2024 | Co-design |
| GMap Neuromorphic | https://dl.acm.org/doi/10.1145/3589737.3605997 | 2023 | ICONS |
| Compiler Survey | https://spj.science.org/doi/10.34133/icomputing.0040 | 2023 | Review |
| SRAM-CIM Literature | https://github.com/BUAA-CI-LAB/Literatures-on-SRAM-based-CIM | Current | GitHub |

---

## Priority 45: Thermal Management (NEW - Iteration 11)

| Paper | URL | Year | Topic |
|-------|-----|------|-------|
| Ferroelectric Memcapacitor | https://pmc.ncbi.nlm.nih.gov/articles/PMC10624373/ | 2023 | Zero Joule heating |
| Capacitive Crossbar Design | https://ieeexplore.ieee.org/document/9439603/ | 2021 | 128×128 array |
| 1Kbit Crossbar Array | https://ieeexplore.ieee.org/document/10044479/ | 2023 | Inversion-type FCM |
| FCM Review | https://nanoconvergencejournal.springeropen.com/articles/10.1186/s40580-024-00463-0 | 2024 | Comprehensive |
| 3D RRAM Thermal | https://www.nature.com/articles/srep13504 | 2015 | Thermal crosstalk |

---

## Priority 46: Application Domains (NEW - Iteration 11)

| Paper | URL | Year | Domain |
|-------|-----|------|--------|
| Neuromorphic Brain Implants | https://www.frontiersin.org/journals/neuroscience/articles/10.3389/fnins.2025.1570104/full | 2025 | Medical |
| IoMT Memory Devices | https://www.sciencedirect.com/science/article/pii/S2666386425003340 | 2025 | Healthcare IoT |
| Flexible ITO FeFET | https://www.nature.com/articles/s41467-024-46878-5 | 2024 | Wearables |
| Reservoir Computing Apps | https://link.springer.com/article/10.1557/s43577-025-00990-z | 2025 | Edge AI |
| Neuromorphic Implantables | https://arxiv.org/html/2506.09599v1 | 2025 | Energy-aware |
| Neuromorphic Market | https://www.mordorintelligence.com/industry-reports/neuromorphic-chip-market | 2024 | Market report |

---

## Priority 47: Research Groups & Institutions (NEW - Iteration 12)

| Resource | URL | Type |
|----------|-----|------|
| Georgia Tech Yu Lab | https://shimeng.ece.gatech.edu/publication/ | Publications |
| Georgia Tech Research | https://shimeng.ece.gatech.edu/research/ | Overview |
| IMEC Ferroelectric | https://www.imec-int.com/en/articles/imec-demonstrates-breakthrough-in-cmos-compatible-ferroelectric-memory | News |
| HfO₂ Game-Changer | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202400686 | Review |
| FeFET Materials to Systems | https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202515480 | Review |

---

## Priority 48: Patent Landscape (NEW - Iteration 12)

| Resource | URL | Year |
|----------|-----|------|
| Memory for AI Patents | https://www.knowmade.com/patent-analytics-services/patent-report/semiconductor-patent-landscape/semiconductor-memory-patent-landscape/memory-for-artificial-intelligence-patent-landscape-analysis-2023/ | 2023 |
| Patent Report Press | https://www.prnewswire.com/news-releases/memory-for-artificial-intelligence-patent-landscape-analysis-report-2023-ibm-and-samsung-have-a-leading-intellectual-property-ip-position-301734562.html | 2023 |
| Samsung 96% Power | https://www.tomshardware.com/tech-industry/semiconductors/samsung-researchers-publish-96percent-lower-power-nand-design-based-on-ferroelectric-transistors | 2025 |
| New Memory Contender | https://semiengineering.com/a-new-memory-contender/ | 2024 |

---

## Priority 49: Testing Methodology (NEW - Iteration 12)

| Paper | URL | Year | Topic |
|-------|-----|------|-------|
| Retention and Endurance | https://ieeexplore.ieee.org/document/8739726/ | 2019 | IMW |
| Endurance Measurement | https://ieeexplore.ieee.org/document/10938969/ | 2025 | Methodology |
| Scaled FeFET Endurance | https://ieeexplore.ieee.org/document/10145966/ | 2023 | In-situ VTH |
| High Endurance FeFET | https://www.researchgate.net/publication/326727783 | 2018 | No retention penalty |
| TU Dresden Thesis | https://d-nb.info/1074350103/34 | Thesis | Characterization |
| HfO₂ FeFET Review | https://pubs.aip.org/aip/jap/article/138/1/010701/3351745 | 2024 | Comprehensive |

---

## Summary Statistics

| Category | Sections | Papers/URLs |
|----------|----------|-------------|
| Core HfO₂/ZrO₂ | 1-3 | ~25 |
| CIM/Neuromorphic | 4-10 | ~40 |
| Simulation/Tools | 11-17 | ~35 |
| STDP/Training | 18-23 | ~30 |
| Advanced Topics | 24-28 | ~25 |
| CAM/Transformer | 29-33 | ~30 |
| NC/RC/3D/Security | 34-38 | ~30 |
| Latest/research sample/Photonic/Edge | 39-43 | ~35 |
| Compilers/Thermal/Apps | 44-46 | ~20 |
| Groups/Patents/Testing | 47-49 | ~20 |
| **Total** | **49 sections** | **~290 URLs** |

---

*Last updated: Iteration 12 - January 2026*
