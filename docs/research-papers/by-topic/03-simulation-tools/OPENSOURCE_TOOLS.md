# Opensource Simulation Tools for FeCIM

**Last Updated:** 2026-01-29
**Status:** Comprehensive documentation of opensource tools used in this project

This document catalogs the opensource simulation tools available for ferroelectric and crossbar array research, along with their associated publications.

---

## 1. CrossSim (Sandia National Laboratories)

### Overview
GPU-accelerated, Python-based crossbar simulator for analog in-memory computing. Designed to model how analog hardware effects in resistive crossbars impact algorithm accuracy.

**Website:** https://cross-sim.sandia.gov/
**GitHub:** https://github.com/sandialabs/cross-sim
**Latest Version:** V3.1 (January 24, 2025)

### Key Features
- Numpy-like API for drop-in replacement
- Device non-idealities: programming errors, conductance drift, read noise, ADC precision loss
- Parasitic metal resistance modeling via internal circuit simulator
- PyTorch and TensorFlow-Keras interfaces (V3.1)
- CUDA GPU acceleration
- Tested on ResNet50/ImageNet (~3× TensorFlow speed)

### Publications

#### Official Papers
1. **CrossSim: GPU-Accelerated Simulation of Analog Neural Networks**
   - Conference: SAND2021-12318C, October 2021
   - Authors: Xiao, Bennett, Feinberg, Marinella, Agarwal
   - [OSTI Link](https://www.osti.gov/biblio/1890592)

2. **CrossSim Inference Manual v2.0**
   - Technical Report, Sandia National Laboratories, May 2022
   - [OSTI Link](https://www.osti.gov/biblio/1869509)

3. **On the accuracy of analog neural network inference accelerators**
   - Authors: T. P. Xiao et al.
   - Journal: IEEE Circuits and Systems Magazine, vol. 22, no. 4, pp. 26-48, 2022

#### Citing CrossSim
```bibtex
@article{CrossSim,
  author = {Ben Feinberg and T. Patrick Xiao and Curtis J. Brinker and Christopher H. Bennett and Matthew J. Marinella and Sapan Agarwal},
  title = {CrossSim: accuracy simulation of analog in-memory computing},
  url = {https://github.com/sandialabs/cross-sim},
}
```

### Tutorial Presentations
- **ISCA 2024:** "CrossSim: A Hardware/Software Co-Design Tool for Analog In-Memory Computing"
- **NICE 2024:** April 26, 2024 tutorial session

---

## 2. badcrossbar (University College London)

### Overview
Nodal analysis solver for passive crossbar arrays with line resistance. Computes currents in all branches and voltages at all nodes, with visualization capabilities.

**GitHub:** https://github.com/joksas/badcrossbar
**Documentation:** https://badcrossbar.readthedocs.io/

### Key Features
- Python-based with Numpy
- Arbitrary voltage inputs, device resistances, interconnect resistances
- Visualization of currents and voltages
- Support for infinite resistance (insulating) devices
- PDF/vector image output

### Publications

#### Official Paper
**badcrossbar: A Python tool for computing and plotting currents and voltages in passive crossbar arrays**
- Authors: Joksas, Dovydas and Mehonic, Adnan
- Journal: SoftwareX, Volume 12, Article 100617
- Year: 2020
- DOI: [10.1016/j.softx.2020.100617](https://doi.org/10.1016/j.softx.2020.100617)

```bibtex
@article{JoksasMehonic2020,
  author = {Joksas, Dovydas and Mehonic, Adnan},
  title = {{badcrossbar}: A {Python} tool for computing and plotting currents and voltages in passive crossbar arrays},
  journal = {SoftwareX},
  volume = {12},
  pages = {100617},
  year = {2020},
  doi = {10.1016/j.softx.2020.100617},
}
```

#### Related Publications
1. **Memristive crossbars as hardware accelerators: modelling, design and new uses** (PhD Thesis)
   - Author: Dovydas Joksas
   - Institution: University College London, 2022
   - [UCL Discovery](https://discovery.ucl.ac.uk/id/eprint/10152211/)

2. **Nonideality-Aware Training for Accurate and Robust Low-Power Memristive Neural Networks**
   - Authors: Joksas, D. et al.
   - Journal: Advanced Science, 2022
   - DOI: [10.1002/advs.202105784](https://doi.org/10.1002/advs.202105784)

---

## 3. FERRET (MOOSE-Based)

### Overview
Open-source MOOSE application for parallel mesoscale simulations of ferroic and related electronic materials using finite-element methods.

**Website:** https://mangerij.github.io/ferret/
**GitHub:** https://github.com/mangerij/ferret

### Key Features
- Real-space finite-element-method (FEM) simulations
- Time-dependent Ginzburg-Landau approach
- Supports: ferroelectrics, ferromagnets, antiferromagnets, multiferroics, piezoelectrics, thermoelectrics
- Parallel computing via MOOSE framework
- ParaView visualization

### Key Developers
- John Mangeri (Technical University of Denmark)
- Serge Nakhmanson (University of Connecticut)
- Olle Heinonen (Argonne National Laboratory)

### Publications (20+ papers, 2015-2024)

#### Highlighted Papers

1. **Influence of Elastic and Surface Strains on the Optical Properties of Semiconducting Core-Shell Nanoparticles**
   - Authors: Mangeri, J., Heinonen, O., Karpeyev, D., Nakhmanson, S.
   - Journal: Phys. Rev. Appl. 4, 014001 (2015)
   - DOI: [10.1103/PhysRevApplied.4.014001](https://doi.org/10.1103/PhysRevApplied.4.014001)

2. **Topological phase transformations and intrinsic size effects in ferroelectric nanoparticles**
   - Authors: Mangeri, J., Espinal, Y., Jokisaari, A., et al.
   - Journal: Nanoscale 9, 1616-1624 (2017)
   - DOI: [10.1039/C6NR09111C](https://doi.org/10.1039/C6NR09111C)

3. **Hopfions emerge in ferroelectrics**
   - Authors: Nahas, Y., Prokhorenko, S., Louis, L., et al.
   - Journal: Nature Communications 11, 2823 (2020)
   - DOI: [10.1038/s41467-020-16258-w](https://doi.org/10.1038/s41467-020-16258-w)

4. **Controllable skyrmion chirality in ferroelectrics**
   - Authors: Tikhonov, Y., Kondovych, S., Mangeri, J., et al.
   - Journal: Scientific Reports 10, 8657 (2020)
   - DOI: [10.1038/s41598-020-65291-8](https://doi.org/10.1038/s41598-020-65291-8)

5. **Manipulating chiral spin transport with ferroelectric polarization**
   - Journal: Nature Materials 23, 898-904 (2024)
   - DOI: [10.1038/s41563-024-01898-4](https://doi.org/10.1038/s41563-024-01898-4)

6. **Towards modeling thermoelectric properties of anisotropic polycrystalline materials**
   - Authors: Basaula, D., Daeipour, M., Kuna, L., et al.
   - Journal: Acta Materialia 228, 117743 (2022)
   - DOI: [10.1016/j.actamat.2022.117743](https://doi.org/10.1016/j.actamat.2022.117743)

#### Full Publication List (Reverse Chronological)
- Extrinsic dielectric response due to domain wall motion in ferroelectric BaTiO₃ (Comp. Mater. Today, 2024)
- Ferroelectric Texture of Individual Barium Titanate Nanocrystals (ACS Nano, 2024)
- Manipulating chiral spin transport with ferroelectric polarization (Nature Materials, 2024)
- Predicting thermoelectric figure of merit in complex materials (Acta Materialia, 2024)
- Modeling structure–properties relations in compositionally disordered relaxor dielectrics (J. Appl. Phys., 2023)
- Coupled magnetostructural continuum model for multiferroic BiFeO₃ (Phys. Rev. B, 2023)
- Towards modeling thermoelectric properties of anisotropic polycrystalline materials (Acta Materialia, 2022)
- Surface charge mediated polar response in ferroelectric nanoparticles (Appl. Phys. Lett., 2021)
- Hopfions emerge in ferroelectrics (Nature Comm., 2020)
- Mesoscale modeling of light transmission modulation in ceramics (Acta Materialia, 2020)
- Controllable skyrmion chirality in ferroelectrics (Sci. Rep., 2020)
- Size, shape and crystallographic orientation dependence of field-induced behavior (J. Appl. Phys., 2019)
- Harnessing ferroelectric domains for negative capacitance (Communications Physics, 2019)
- Mesoscale modeling of polycrystalline light transmission (Acta Materialia, 2019)
- Electromechanical control of polarization vortex ordering (Appl. Phys. Lett., 2018)
- Metastable vortex-like polarization textures in ferroelectric nanoparticles (J. Appl. Phys., 2018)
- Domain alignment within PbTiO₃/SrTiO₃ superlattice nanostructure (Nanoscale, 2018)
- Topological phase transformations in ferroelectric nanoparticles (Nanoscale, 2017)
- Stress-induced shift of band gap in ZnO nanowires (Phys. Rev. Appl., 2017)
- Influence of Elastic and Surface Strains on Optical Properties (Phys. Rev. Appl., 2015)

---

## 4. FerroX (Lawrence Berkeley Lab / AMReX)

### Overview
Massively parallel, 3D phase-field simulation framework for modeling ferroelectric materials based scalable logic devices. Self-consistently solves TDGL, Poisson's equation, and semiconductor charge equations.

**GitHub:** https://github.com/AMReX-Microelectronics/FerroX
**Contact:** Zhi (Jackie) Yao (jackie_zhiyao@lbl.gov), Andy Nonaka (ajnonaka@lbl.gov)

### Key Features
- Exascale Computing Project (ECP) framework (AMReX)
- GPU acceleration (CUDA, HIP, SYCL)
- MPI + OpenMP parallelization
- MFIM and MFISM device simulation
- Negative capacitance (NC) effect modeling
- HZO polycrystalline multi-phase simulations

### Publications

#### Primary Paper
**FerroX: A GPU-accelerated, 3D Phase-Field Simulation Framework for Modeling Ferroelectric Devices**
- Authors: P. Kumar, A. Nonaka, R. Jambunathan, G. Pahwa, S. Salahuddin, Z. Yao
- Journal: Computer Physics Communications, Volume 290, Article 108757
- Year: 2023
- DOI: [10.1016/j.cpc.2023.108757](https://doi.org/10.1016/j.cpc.2023.108757)
- arXiv: [2210.15668](https://arxiv.org/abs/2210.15668)

#### HZO Polycrystalline Paper
**3D ferroelectric phase field simulations of polycrystalline multi-phase hafnia and zirconia based ultra-thin films**
- Authors: P. Kumar, M. Hoffmann, A. Nonaka, S. Salahuddin, Z. Yao
- Journal: Advanced Electronic Materials
- Year: 2024
- DOI: [10.1002/aelm.202400085](https://doi.org/10.1002/aelm.202400085)
- arXiv: [2402.05331](https://arxiv.org/abs/2402.05331)

---

## 5. Additional Tools

### NeuroSim (Georgia Tech)
- **Purpose:** CIM hardware accelerator simulation with device/circuit non-idealities
- **Paper:** "NeuroSim Simulator for Compute-in-Memory Hardware Accelerator"
- **Journal:** Frontiers in Artificial Intelligence, 2021
- **DOI:** [10.3389/frai.2021.659060](https://doi.org/10.3389/frai.2021.659060)
- **GitHub:** https://github.com/neurosim/DNN_NeuroSim_V1.4
- **Used by:** Intel, Samsung, TSMC, SK Hynix

### XbarSim (New - 2024)
- **Purpose:** Decomposition-based memristive crossbar simulator
- **Paper:** "XbarSim: A Decomposition-Based Memristive Crossbar Simulator"
- **arXiv:** [2410.19993](https://arxiv.org/abs/2410.19993)
- **Note:** Addresses limitations in CrossSim and badcrossbar

---

## Tool Comparison

| Feature | CrossSim | badcrossbar | FERRET | FerroX |
|---------|----------|-------------|--------|--------|
| **Primary Use** | CIM accuracy | Crossbar IV | Mesoscale FE | Device simulation |
| **Method** | Circuit sim | Nodal analysis | FEM (MOOSE) | Phase-field (AMReX) |
| **GPU Support** | CUDA | No | Limited | CUDA/HIP/SYCL |
| **Neural Networks** | Yes (DNN) | No | No | No |
| **Ferroelectric** | Any device | Any device | Core focus | HfO₂/HZO focus |
| **Publications** | ~5 | 1 (+ thesis) | 20+ | 2 |
| **Active Dev** | Very active | Maintained | Active | Active |

---

## Integration with This Project

### Module 2 (Crossbar) Validation
The FeCIM project validates crossbar algorithms against both CrossSim and badcrossbar:
- Location: `module2-crossbar/pkg/crossbar/`
- Tests: `validation_test.go`
- Purpose: Ensure our Go implementation matches established Python tools

### Recommended Workflow
1. **Device modeling:** Use FerroX for HZO phase-field simulations
2. **Mesoscale physics:** Use FERRET for domain dynamics
3. **Array-level accuracy:** Use CrossSim for DNN inference validation
4. **Quick IV checks:** Use badcrossbar for parasitic resistance analysis

---

## References

All DOIs and links verified as of 2026-01-29.
