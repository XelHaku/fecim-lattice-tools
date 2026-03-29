# Open-Source Tools Summary for FeCIM Development

**A comprehensive catalog of 65 open-source projects analyzed for Ferroelectric Compute-in-Memory research.**

*Generated: February 2026*

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Category Overview](#category-overview)
3. [Tool Catalog by Category](#tool-catalog-by-category)
4. [Top 15 Tools for FeCIM](#top-15-tools-for-fecim)
5. [Integration Workflows by Module](#integration-workflows-by-module)
6. [Quick Reference](#quick-reference)

---

## Executive Summary

This document summarizes **65 open-source projects** cloned to `opensource/`, organized into **10 functional categories**. Each tool was evaluated for its relevance to FeCIM (Ferroelectric Compute-in-Memory) development across six project modules:

| Module | Focus |
|--------|-------|
| **M1** | Hysteresis & P-E curves |
| **M2** | Crossbar array simulation |
| **M3** | MNIST neural network inference |
| **M4** | Peripheral circuits (ADC/DAC) |
| **M5** | Lattice visualization |
| **M6** | EDA / chip tapeout |

> [!IMPORTANT]
> The 15 highest-priority tools are highlighted in [Top 15 Tools for FeCIM](#top-15-tools-for-fecim). These are the tools most directly applicable to the project's current development needs.

---

## Category Overview

| # | Category | Tool Count | Primary Use | Key Languages |
|---|----------|-----------|-------------|---------------|
| 1 | Ferroelectric Simulation | 8 | Device physics, P-E curves | C++, Python, Fortran |
| 2 | Hysteresis Modeling | 7 | Preisach/Jiles-Atherton models | Python, Julia |
| 3 | Circuit Simulation | 8 | SPICE, compact models | C++, Python |
| 4 | Circuit Analysis Libraries | 5 | Crossbar solvers, symbolic analysis | Python |
| 5 | Memristor/RRAM Tools | 6 | Device models, crossbar simulation | Python, Verilog-A |
| 6 | NN Hardware Mapping | 8 | Quantization, hardware-aware training | Python (PyTorch/TF) |
| 7 | EDA Tools | 11 | RTL-to-GDSII, PDKs | C++, TCL, Verilog |
| 8 | Physics Visualization | 8 | 3D rendering, P-E plotting | Python |
| 9 | Scientific Computing | 8 | FEM, DFT, materials databases | Python, C++, Fortran |
| 10 | Data Acquisition | 5 | Lab instrument control | Python |

---

## Tool Catalog by Category

### 1. Ferroelectric Simulation

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **FerroX** | BSD-3 | 🔴 HIGH | GPU-accelerated phase-field simulator for HfO₂, TDGL-based 3D simulation |
| **FERRET** | LGPL | 🔴 HIGH | MOOSE-based multiphysics for domain dynamics, Landau-Devonshire theory |
| **PFECAP** | GPL-2 | 🔴 HIGH | Verilog-A Preisach model for SPICE integration (1T1C cells) |
| **Ferro** | CC BY-NC-SA | 🟡 MEDIUM | Python library for experimental P-E loop analysis, Preisach fitting |
| **pymatgen** | MIT | 🟡 MEDIUM | Materials analysis, DFT-based polarization, dopant screening |
| **Q-POP-Thermo** | MIT | 🟡 MEDIUM | Landau-Ginzburg-Devonshire phase diagrams |
| **negativec** | GPL-3 | 🟡 MEDIUM | Negative capacitance FET physics simulation |
| **feram** | GPL-2 | 🟢 LOW | Effective Hamiltonian for ABO₃ perovskites (not HfO₂) |

**Key takeaway**: FerroX is the primary tool for detailed device simulation. PFECAP bridges to SPICE. Ferro handles experimental data analysis.

---

### 2. Hysteresis Modeling

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **python-preisach** | MIT | 🔴 HIGH | Educational Preisach model implementation |
| **Preisachmodel** | MIT | 🔴 HIGH | Full forward + inverse Preisach with data fitting |
| **ferro_scripts** | MIT | 🔴 HIGH | Garrity/Tour P-E curves, YAML-configurable |
| **pyhist** | MIT | 🟡 MEDIUM | Generic hysteresis modeling |
| **JAmodel** | MIT | 🟡 MEDIUM | Jiles-Atherton magnetic hysteresis (adaptable) |
| **pyjam** | MIT | 🟡 MEDIUM | Jiles-Atherton with parameter optimization |
| **CIM.jl** | MIT | 🟡 MEDIUM | Julia implementation of CIM models |

**Key takeaway**: Preisach models are the foundation for M1. `ferro_scripts` provides the fastest path to P-E curve generation.

---

### 3. Circuit Simulation

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **ngspice** | BSD-3 | 🔴 HIGH | Industry-standard SPICE with OSDI/Verilog-A support |
| **PySpice** | GPL-3 | 🔴 HIGH | Python wrapper for ngspice, parametric sweeps |
| **OpenVAF** | GPL-3 | 🔴 HIGH | Verilog-A → OSDI compiler for compact models |
| **Xyce** | GPL-3 | 🟡 MEDIUM | Parallel SPICE for large circuits (millions of nodes) |
| **QUCS-S** | GPL-2 | 🟡 MEDIUM | GUI schematic editor for ngspice/Xyce |
| **KiCad** | GPL-3 | 🟡 MEDIUM | Professional EDA with built-in ngspice |
| **Ahkab** | GPL-2 | 🟢 LOW | Pure-Python SPICE (educational) |
| **CircuitJS** | GPL-2 | 🟢 LOW | Browser-based real-time circuit simulator |
| **GnuCap** | GPL-3 | 🟢 LOW | Interactive circuit exploration |

**Key takeaway**: ngspice + PySpice + OpenVAF form the core circuit simulation stack. Use Xyce for full-crossbar SPICE simulations.

---

### 4. Circuit Analysis Libraries

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **badcrossbar** | MIT | 🔴 HIGH | Exact nodal analysis for passive crossbar IR drop & sneak paths |
| **CrossSim** | BSD-3 | 🔴 HIGH | GPU-accelerated MVM simulator with comprehensive non-ideality models |
| **MemTorch** | GPL-3 | 🔴 HIGH | PyTorch-integrated memristive crossbar simulation |
| **Lcapy** | GPL-3 | 🟡 MEDIUM | Symbolic circuit analysis (transfer functions, noise) |
| **SymCircuit** | MIT | 🟢 LOW | Basic symbolic circuit solver |

**Key takeaway**: badcrossbar for validation, CrossSim for production simulation, MemTorch for NN integration.

---

### 5. Memristor/RRAM Tools

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **MemTorch** | GPL-3 | 🔴 HIGH | PyTorch extension for memristive DNNs, crossbar mapping |
| **CrossSim** | BSD-3 | 🔴 HIGH | Sandia's reference CIM simulator with device models (SONOS, PCM, RRAM) |
| **MNSIM 2.0** | MIT | 🟡 MEDIUM | Fast design space exploration for memristor systems |
| **badcrossbar** | MIT | 🔴 HIGH | IR drop and sneak path analysis |
| **VTEAM** | MIT | 🟡 MEDIUM | Voltage-threshold memristor compact model (Verilog-A) |
| **Yakopcic RRAM** | MIT | 🟡 MEDIUM | Physics-based RRAM model, adaptable for FeFETs |

**Key takeaway**: CrossSim is the reference implementation. MemTorch adds PyTorch integration. VTEAM/Yakopcic models are adaptable for ferroelectric switching.

---

### 6. NN Hardware Mapping

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **Brevitas** | BSD-3 | 🔴 HIGH | Arbitrary-bit QAT (native 5-bit = 32 levels ≈ FeCIM 30 levels) |
| **IBM AIHWKIT** | Apache-2 | 🔴 HIGH | Hardware-aware training with analog device non-idealities |
| **CrossSim** | BSD-3 | 🔴 HIGH | GPU-accelerated crossbar mapping with Keras/PyTorch layers |
| **HAWQ** | BSD-3 | 🟡 MEDIUM | Hessian-aware mixed-precision quantization |
| **QKeras** | Apache-2 | 🟡 MEDIUM | Keras quantization with energy estimation |
| **NNCF** | BSD-3 | 🟡 MEDIUM | Intel's compression framework (quantization + pruning) |
| **TF Model Opt** | Apache-2 | 🟡 MEDIUM | Google's post-training quantization (8-bit focus) |
| **DNN+NeuroSim** | MIT | 🟡 MEDIUM | NN + device co-simulation framework |

**Key takeaway**: Brevitas for 5-bit quantization → IBM AIHWKIT for hardware-aware training → CrossSim for crossbar validation.

---

### 7. EDA Tools

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **OpenLane** | Apache-2 | 🔴 HIGH | Automated RTL-to-GDSII flow (600+ tapeouts) |
| **Yosys** | ISC | 🔴 HIGH | RTL synthesis (Verilog → gate-level netlist) |
| **OpenROAD** | BSD-3 | 🔴 HIGH | Place & route engine |
| **SKY130 PDK** | Apache-2 | 🔴 HIGH | 130nm process design kit (Google-sponsored) |
| **Caravel** | Apache-2 | 🔴 HIGH | Harness for free chip tapeout via Efabless |
| **IIC-OSIC-TOOLS** | Various | 🔴 HIGH | All-in-one Docker container with all EDA tools |
| **Magic** | Open Source | 🟡 MEDIUM | Layout editor, DRC, parasitic extraction |
| **Netgen** | Open Source | 🟡 MEDIUM | Layout vs. Schematic (LVS) verification |
| **KLayout** | GPL | 🟡 MEDIUM | GDSII viewer with Python scripting |
| **GF180 PDK** | Apache-2 | 🟡 MEDIUM | 180nm PDK (better for analog) |
| **IHP130 PDK** | Open Source | 🟢 LOW | 130nm BiCMOS (RF analog) |

**Key takeaway**: OpenLane orchestrates the entire RTL-to-GDSII flow. SKY130 is the primary target process. IIC-OSIC-TOOLS provides a single Docker container with everything.

---

### 8. Physics Visualization

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **Matplotlib** | PSF | 🟡 MEDIUM | Publication-quality 2D plots, P-E curves |
| **hysteresis** | Apache-2 | 🔴 HIGH | Specialized P-E loop analysis (FORC diagrams, feature extraction) |
| **Plotly** | MIT | 🟡 MEDIUM | Interactive web-based 2D/3D plots |
| **VPython** | BSD | 🟡 MEDIUM | Real-time 3D physics animation (Jupyter) |
| **PyVista** | MIT | 🟡 MEDIUM | VTK-based 3D scientific visualization |
| **Mayavi** | BSD | 🟡 MEDIUM | VTK-based vector/scalar field visualization |
| **K3D-Jupyter** | MIT | 🟡 MEDIUM | WebGL 3D visualization in Jupyter |
| **Manim** | MIT | 🟢 LOW | Mathematical animation engine |

**Key takeaway**: `hysteresis` package is the only tool specifically designed for ferroelectric loop analysis. PyVista for 3D domain structures.

---

### 9. Scientific Computing

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **FEniCSx** | LGPL | 🟡 MEDIUM | Modern FEM framework (phase-field, electromechanics) |
| **MOOSE** | LGPL | 🟡 MEDIUM | Multiphysics FEM, hosts FERRET module |
| **deal.II** | LGPL | 🟢 LOW | C++ FEM library with 1000+ tutorials |
| **SfePy** | BSD-3 | 🟡 MEDIUM | Python FEM with native piezoelectric support |
| **GPAW** | GPL-3 | 🟢 LOW | Open-source DFT calculator |
| **Quantum ESPRESSO** | GPL-2 | 🟢 LOW | DFT suite with phonon/Berry phase support |
| **ABINIT** | GPL-2 | 🟢 LOW | DFT with many-body physics |
| **Materials Project** | CC/MIT | 🟡 MEDIUM | Database of 150K+ computed materials properties |

**Key takeaway**: MOOSE/FERRET for phase-field simulations. Materials Project for property lookups. SfePy uniquely supports piezoelectric coupling.

---

### 10. Data Acquisition

| Tool | License | FeCIM Relevance | Description |
|------|---------|-----------------|-------------|
| **PyVISA** | MIT | 🟡 MEDIUM | Foundation protocol for lab instrument communication |
| **PyMeasure** | MIT | 🟡 MEDIUM | High-level instrument drivers (Keithley, Agilent, SR830) |
| **QCoDeS** | MIT | 🟡 MEDIUM | Enterprise-grade lab framework with SQLite logging |
| **python-ivi** | MIT | 🟡 MEDIUM | IVI-standard instrument substitution |
| **nidaqmx** | MIT | 🟡 MEDIUM | NI DAQ card multi-channel synchronization |

**Key takeaway**: PyMeasure is the recommended starting point for lab automation. QCoDeS for advanced multi-instrument setups.

---

## Top 15 Tools for FeCIM

These are the most impactful tools across all categories, ranked by direct applicability:

| Rank | Tool | Category | Why It Matters |
|------|------|----------|----------------|
| 1 | **CrossSim** | Crossbar Sim | Reference CIM simulator with GPU, validated against SPICE |
| 2 | **Brevitas** | NN Mapping | Native 5-bit quantization matching FeCIM's 30 levels |
| 3 | **ngspice** | Circuit Sim | Industry-standard SPICE for detailed circuit validation |
| 4 | **badcrossbar** | Circuit Analysis | Exact IR drop and sneak path analysis for validation |
| 5 | **IBM AIHWKIT** | NN Mapping | Most realistic hardware-aware training (noise, drift, variation) |
| 6 | **FerroX** | Ferro Sim | GPU-accelerated 3D HfO₂ device simulation |
| 7 | **OpenLane** | EDA | Automated RTL-to-GDSII with 600+ successful tapeouts |
| 8 | **PFECAP** | Ferro Sim | Verilog-A ferroelectric model for SPICE circuits |
| 9 | **MemTorch** | Memristor | PyTorch-integrated crossbar simulation with non-idealities |
| 10 | **PySpice** | Circuit Sim | Python automation of SPICE parametric sweeps |
| 11 | **SKY130 PDK** | EDA | Target 130nm process for chip tapeout |
| 12 | **Preisachmodel** | Hysteresis | Forward + inverse Preisach with data fitting |
| 13 | **Lcapy** | Circuit Analysis | Symbolic transfer function derivation for peripherals |
| 14 | **hysteresis** | Visualization | Specialized P-E loop analysis (FORC, energy, Ec/Pr extraction) |
| 15 | **IIC-OSIC-TOOLS** | EDA | Single Docker container with entire EDA toolchain |

---

## Integration Workflows by Module

### Module 1: Hysteresis

```
ferro_scripts → Generate P-E curves
       ↓
Ferro / Preisachmodel → Fit Preisach distribution
       ↓
hysteresis package → Extract Ec, Pr, loop area
       ↓
PFECAP → SPICE-compatible model
       ↓
module1-hysteresis (Go) → Interactive GUI
```

### Module 2: Crossbar Array

```
badcrossbar → Quick IR drop validation (exact)
       ↓
CrossSim → Full MVM with non-idealities (GPU)
       ↓
ngspice + PFECAP → Circuit-level spot checks
       ↓
module2-crossbar (Go) → Behavioral simulator
```

### Module 3: MNIST Neural Network

```
Brevitas → 5-bit quantization-aware training
       ↓
CrossSim → Map weights to crossbar tiles
       ↓
IBM AIHWKIT / MemTorch → Validate under device variations
       ↓
module3-mnist (Go) → Interactive inference demo
```

### Module 4: Peripheral Circuits

```
KiCad / QUCS-S → Schematic capture (ADC/DAC)
       ↓
PySpice / ngspice → Parametric simulation
       ↓
Lcapy → Symbolic transfer function analysis
       ↓
module4-circuits (Go) → Educational visualization
```

### Module 5: Lattice Visualization

```
pymatgen → Crystal structure data
       ↓
VPython / PyVista → 3D superlattice rendering
       ↓
Matplotlib → Publication-quality static plots
       ↓
module5-lattice (Go + Fyne) → Interactive 3D viewer
```

### Module 6: EDA / Chip Design

```
Yosys → Synthesize digital control Verilog
       ↓
OpenROAD → Place & route
       ↓
Magic → Custom analog cell layout + DRC
       ↓
Netgen → LVS verification
       ↓
OpenLane → Automated RTL-to-GDSII
       ↓
Caravel → Free tapeout via Efabless
```

---

## Quick Reference

### Installation Quick Start

```bash
# Core simulation stack (Python)
pip install badcrossbar memtorch brevitas aihwkit pyspice lcapy

# Visualization
pip install matplotlib plotly pyvista hysteresis vpython

# Lab equipment
pip install pyvisa pymeasure qcodes

# EDA (Docker - recommended)
docker pull hpretl/iic-osic-tools:latest
```

### Decision Matrix: "Which Tool Should I Use?"

| I want to... | Use this tool |
|--------------|---------------|
| Generate P-E curves quickly | `ferro_scripts` |
| Fit experimental hysteresis data | `Ferro` or `Preisachmodel` |
| Verify crossbar IR drop | `badcrossbar` |
| Simulate full crossbar MVM with GPU | `CrossSim` |
| Train 5-bit quantized NN | `Brevitas` |
| Train NN with device noise/drift | `IBM AIHWKIT` |
| Run SPICE simulation | `ngspice` (CLI) or `PySpice` (Python) |
| Automate parametric sweeps | `PySpice` |
| Create ferroelectric SPICE model | `PFECAP` + `OpenVAF` |
| Design a chip RTL-to-GDSII | `OpenLane` |
| Visualize 3D domain structures | `PyVista` or `VPython` |
| Control lab instruments | `PyMeasure` |
| Get all EDA tools in one container | `IIC-OSIC-TOOLS` Docker |

### Key Resources

- Detailed analysis: [4-research/opensource-tools/](4-research/opensource-tools/)
  - [ferroelectric-simulation-tools.md](4-research/opensource-tools/ferroelectric-simulation-tools.md)
  - [hysteresis-modeling-tools.md](4-research/opensource-tools/hysteresis-modeling-tools.md)
  - [circuit-simulation-tools.md](4-research/opensource-tools/circuit-simulation-tools.md)
  - [circuit-analysis-libraries.md](4-research/opensource-tools/circuit-analysis-libraries.md)
  - [memristor-rram-tools.md](4-research/opensource-tools/memristor-rram-tools.md)
  - [nn-hardware-mapping-tools.md](4-research/opensource-tools/nn-hardware-mapping-tools.md)
  - [eda-tools.md](4-research/opensource-tools/eda-tools.md)
  - [physics-visualization-tools.md](4-research/opensource-tools/physics-visualization-tools.md)
  - [scientific-computing-tools.md](4-research/opensource-tools/scientific-computing-tools.md)
  - [data-acquisition-tools.md](4-research/opensource-tools/data-acquisition-tools.md)
  - [tool-comparison-matrix.md](4-research/opensource-tools/tool-comparison-matrix.md)
  - [opensource-crossbar.md](4-research/opensource-tools/opensource-crossbar.md)
- Clone script: [opensource/clone-all.sh](../opensource/clone-all.sh)
