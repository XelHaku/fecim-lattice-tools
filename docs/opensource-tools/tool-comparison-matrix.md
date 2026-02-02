# Open Source Tools Comparison Matrix for FeCIM Development

**A Master Reference for Selecting the Right Tool for Your FeCIM Project**

*Last Updated: January 2026*

---

**Note:** References to 30 levels refer to the demo baseline (conference claim; pending peer review). Peer‑reviewed devices report 32–140 states.

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Master Comparison Matrix](#master-comparison-matrix)
3. [Category-Specific Matrices](#category-specific-matrices)
4. [Quick Decision Guide](#quick-decision-guide)
5. [Detailed Tool Assessments](#detailed-tool-assessments)
6. [Integration Recommendations](#integration-recommendations)

---

## Executive Summary

This document provides a comprehensive comparison across **50+ open-source tools** organized into 10 functional categories relevant to Ferroelectric Compute-in-Memory (FeCIM) development. Each tool is evaluated on:

- **FeCIM Relevance:** How directly applicable to FeCIM research
- **Ease of Use:** Learning curve and time to productivity (1-5 stars)
- **Documentation Quality:** Availability of tutorials, examples, papers
- **Maintenance Status:** Active development, community support
- **Python Integration:** Can you use Python for workflows?
- **GPU Support:** Hardware acceleration availability
- **Key Strengths:** What this tool does best
- **Key Limitations:** When NOT to use this tool

---

## Master Comparison Matrix

### Legend

- **FeCIM Relevance:** HIGH (directly applicable), MEDIUM (useful for related work), LOW (general purpose)
- **Ease:** ⭐ to ⭐⭐⭐⭐⭐ (1-5 stars)
- **Maintenance:** ✅ Active, ⚠️ Maintenance, ❌ Dormant
- **Python:** ✅ Full, ⭐ Partial, ❌ None
- **GPU:** ✅ Yes, ❌ No

| Tool | Category | License | FeCIM Relevance | Ease | Maintenance | Python | GPU | Key Use Case |
|------|----------|---------|-----------------|------|-------------|--------|-----|---|
| **FERRET** | Ferroelectric Simulation | LGPL | HIGH | ⭐⭐ | ✅ | ❌ | ❌ | Phase-field domain structure (HfO₂) |
| **FerroX** | Ferroelectric Simulation | BSD-3 | HIGH | ⭐⭐⭐ | ✅ | ❌ | ✅ | GPU-accelerated 3D ferroelectric simulation |
| **feram** | Ferroelectric Simulation | GPL-2 | LOW | ⭐⭐ | ⚠️ | ❌ | ❌ | ABO₃ perovskites (not HfO₂) |
| **pymatgen** | Ferroelectric Simulation | MIT | MEDIUM | ⭐⭐⭐ | ✅ | ✅ | ❌ | DFT material screening, Berry phase |
| **Q-POP-Thermo** | Ferroelectric Simulation | MIT | MEDIUM | ⭐⭐⭐ | ⚠️ | ✅ | ❌ | LGD phase diagrams |
| **Ferro** | Ferroelectric Simulation | CC BY-NC-SA | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Experimental P-E loop analysis |
| **PFECAP** | Ferroelectric Simulation | GPL-2 | HIGH | ⭐⭐⭐ | ⚠️ | ❌ | ❌ | SPICE ferroelectric capacitor model |
| **negativec** | Ferroelectric Simulation | GPL-3 | MEDIUM | ⭐⭐ | ⚠️ | ⭐ | ❌ | Negative capacitance FET physics |
| | | | | | | | | |
| **OpenLane** | EDA Tools | Apache-2 | HIGH | ⭐⭐⭐ | ✅ | ⭐ | ❌ | RTL-to-GDSII automation |
| **OpenROAD** | EDA Tools | BSD-3 | HIGH | ⭐⭐⭐ | ✅ | ⭐ | ❌ | Place & Route |
| **Yosys** | EDA Tools | ISC | HIGH | ⭐⭐⭐ | ✅ | ⭐ | ❌ | RTL synthesis (Verilog → Gates) |
| **Magic** | EDA Tools | Open Source | MEDIUM | ⭐⭐ | ✅ | ⭐ | ❌ | Layout editor, DRC verification |
| **KLayout** | EDA Tools | GPL | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | GDSII viewer, DRC/LVS runner |
| **Netgen** | EDA Tools | Open Source | MEDIUM | ⭐⭐ | ✅ | ⭐ | ❌ | Layout vs. Schematic verification |
| **SKY130 PDK** | EDA Tools | Apache-2 | HIGH | ⭐⭐⭐⭐ | ✅ | N/A | N/A | 130nm process design kit |
| **GF180 PDK** | EDA Tools | Apache-2 | MEDIUM | ⭐⭐⭐⭐ | ✅ | N/A | N/A | 180nm process design kit |
| **IHP130 PDK** | EDA Tools | Open Source | LOW | ⭐⭐⭐ | ✅ | N/A | N/A | 130nm BiCMOS (RF analog) |
| **Caravel** | EDA Tools | Apache-2 | HIGH | ⭐⭐⭐⭐ | ✅ | ⭐ | ❌ | Harness for free chip tapeout |
| **IIC-OSIC-TOOLS** | EDA Tools | Various | HIGH | ⭐⭐⭐⭐⭐ | ✅ | ⭐ | ❌ | All-in-one Docker container |
| | | | | | | | | |
| **ngspice** | Circuit Simulation | BSD-3 | HIGH | ⭐⭐⭐⭐ | ✅ | ⭐ | ❌ | Industry-standard SPICE (ADC/DAC) |
| **Xyce** | Circuit Simulation | GPL-3 | MEDIUM | ⭐⭐ | ✅ | ⭐ | ❌ | Parallel SPICE (large circuits) |
| **QUCS-S** | Circuit Simulation | GPL-2 | MEDIUM | ⭐⭐⭐⭐ | ✅ | ❌ | ❌ | GUI schematic editor for ngspice |
| **PySpice** | Circuit Simulation | GPL-3 | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Python-scripted SPICE parametric sweeps |
| **Ahkab** | Circuit Simulation | GPL-2 | LOW | ⭐⭐⭐ | ❌ | ✅ | ❌ | Educational pure-Python SPICE |
| **GnuCap** | Circuit Simulation | GPL-3 | LOW | ⭐⭐ | ⚠️ | ⭐ | ❌ | Interactive circuit exploration |
| **CircuitJS** | Circuit Simulation | GPL-2 | LOW | ⭐⭐⭐⭐⭐ | ✅ | N/A | N/A | Real-time browser simulator |
| **KiCad** | Circuit Simulation | GPL-3 | MEDIUM | ⭐⭐⭐⭐ | ✅ | ⭐ | ❌ | Professional schematic + ngspice |
| **OpenVAF** | Circuit Simulation | GPL-3 | HIGH | ⭐⭐⭐ | ✅ | ⭐ | ❌ | Verilog-A → OSDI compiler |
| | | | | | | | | |
| **badcrossbar** | Circuit Analysis | MIT | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Crossbar IR drop + sneak paths |
| **Ahkab (symbolic)** | Circuit Analysis | GPL-2 | MEDIUM | ⭐⭐⭐ | ❌ | ✅ | ❌ | Symbolic transfer function analysis |
| **Lcapy** | Circuit Analysis | GPL-3 | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Python symbolic circuits |
| **MemTorch** | Circuit Analysis | GPL-3 | HIGH | ⭐⭐⭐ | ✅ | ✅ | ✅ | Memristor crossbar simulation |
| **CrossSim** | Circuit Analysis | BSD-3 | HIGH | ⭐⭐⭐ | ✅ | ✅ | ✅ | GPU-accelerated MVM simulator |
| | | | | | | | | |
| **Matplotlib** | Physics Visualization | PSF | MEDIUM | ⭐⭐⭐⭐⭐ | ✅ | ✅ | ⭐ | Static publication-quality plots |
| **Plotly** | Physics Visualization | MIT | MEDIUM | ⭐⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Interactive web-based plots |
| **VPython** | Physics Visualization | GPL | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Real-time 3D physics animation |
| **PyVista** | Physics Visualization | MIT | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ✅ | 3D field visualization |
| **Manim** | Physics Visualization | MIT | LOW | ⭐⭐ | ✅ | ✅ | ❌ | Educational animation generation |
| **ParaView** | Physics Visualization | BSD | MEDIUM | ⭐⭐⭐ | ✅ | ⭐ | ✅ | Scientific data visualization |
| | | | | | | | | |
| **python-preisach** | Hysteresis Models | MIT | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Educational Preisach model |
| **Preisachmodel** | Hysteresis Models | MIT | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Forward + inverse Preisach |
| **ferro_scripts** | Hysteresis Models | MIT | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Garrity/Tour P-E curves |
| **CIM.jl** | Hysteresis Models | MIT | MEDIUM | ⭐⭐⭐ | ⚠️ | ❌ | ❌ | Julia implementation |
| | | | | | | | | |
| **MemTorch** | Memristor/RRAM | GPL-3 | HIGH | ⭐⭐⭐ | ✅ | ✅ | ✅ | Device + array simulation |
| **badcrossbar** | Memristor/RRAM | MIT | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Crossbar non-idealities |
| **CrossSim** | Memristor/RRAM | BSD-3 | HIGH | ⭐⭐⭐ | ✅ | ✅ | ✅ | Sandia's crossbar simulator |
| **DNN+NeuroSim** | Memristor/RRAM | MIT | MEDIUM | ⭐⭐ | ❌ | ❌ | ❌ | NN + device co-simulation |
| **MNSIM** | Memristor/RRAM | MIT | MEDIUM | ⭐⭐ | ❌ | ❌ | ❌ | RRAM system simulator |
| **VTEAM** | Memristor/RRAM | MIT | MEDIUM | ⭐⭐⭐ | ⚠️ | ✅ | ❌ | Voltage-threshold memristor model |
| | | | | | | | | |
| **Brevitas** | NN Hardware Mapping | BSD-3 | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ✅ | Arbitrary-bit quantization (5-bit!) |
| **HAWQ** | NN Hardware Mapping | MIT | MEDIUM | ⭐⭐⭐ | ⚠️ | ✅ | ✅ | Hessian-aware quantization |
| **QKeras** | NN Hardware Mapping | Apache-2 | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ⭐ | Keras quantization |
| **IBM AIHWKit** | NN Hardware Mapping | MIT | HIGH | ⭐⭐⭐⭐ | ✅ | ✅ | ✅ | Hardware-aware training |
| **CrossSim** | NN Hardware Mapping | BSD-3 | HIGH | ⭐⭐⭐ | ✅ | ✅ | ✅ | GPU-accelerated mapping |
| | | | | | | | | |
| **FEniCS** | Scientific Computing | LGPL | LOW | ⭐⭐ | ✅ | ✅ | ⚠️ | Finite element (not ferroelectric-specific) |
| **MOOSE** | Scientific Computing | LGPL | MEDIUM | ⭐⭐ | ✅ | ❌ | ⭐ | Multiphysics framework (for FERRET) |
| **SfePy** | Scientific Computing | BSD | LOW | ⭐⭐⭐ | ✅ | ✅ | ❌ | Simple FEM |
| **PyMatGen** | Scientific Computing | MIT | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Material DFT analysis |
| **ASE** | Scientific Computing | LGPL | LOW | ⭐⭐⭐⭐ | ✅ | ✅ | ⚠️ | Atomic structure simulation |
| **GPAW** | Scientific Computing | GPL-3 | LOW | ⭐⭐ | ✅ | ✅ | ✅ | DFT calculator (heavy) |
| | | | | | | | | |
| **PyVISA** | Data Acquisition | MIT | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | SMU/oscilloscope control |
| **PyMeasure** | Data Acquisition | GPL-3 | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Instrument automation |
| **QCoDeS** | Data Acquisition | MIT | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | Advanced lab setup orchestration |
| **nidaqmx** | Data Acquisition | MIT | MEDIUM | ⭐⭐⭐⭐ | ✅ | ✅ | ❌ | NI data acquisition (hardware-specific) |

---

## Category-Specific Matrices

### 1. Ferroelectric Simulation Tools

| Tool | Scale | Physics Model | Speed | HfO₂ Support | Wake-up/Fatigue | Temperature | Best For |
|------|-------|---------------|-------|--------------|-----------------|-------------|----------|
| FERRET | Mesoscale | Landau phenomenological | Slow (CPU) | Via parameters | ❌ | Limited | Domain dynamics |
| FerroX | Mesoscale | TDGL (field-based) | Fast (GPU 15×) | Native | ✅ | ⚠️ | 3D device simulation |
| feram | Macroscale | Effective Hamiltonian | Fast | ❌ | ✅ | ✅ | ABO₃ comparison |
| pymatgen | DFT | Berry phase | Varies | Via DFT | N/A | N/A | Material screening |
| Q-POP-Thermo | Macroscale | LGD free energy | Fast | Via parameters | ⚠️ | ✅ | Phase diagrams |
| Ferro | Analysis | Preisach | Fast | Via fitting | ⚠️ | ⚠️ | Experimental data |
| PFECAP | Device | Preisach | Very fast | Yes | ⚠️ | Supported | SPICE circuits |
| negativec | Device | FERRET+TDGL | Slow | Native | ✅ | ⚠️ | NC-FET physics |

### 2. EDA/PDK Tools

| Tool | Type | Language | GUI | Maturity | Community | Best For |
|------|------|----------|-----|----------|-----------|----------|
| OpenLane | Flow | TCL/Python | ⭐⭐ | Production | ✅ Large | Full RTL-to-GDSII |
| OpenROAD | P&R | C++ | CLI | Production | ✅ Active | Place & route |
| Yosys | Synthesis | C++ | CLI | Production | ✅ Active | RTL → gates |
| Magic | Layout | C/Tcl | GUI | Active | ✅ Large | Custom cells, DRC |
| KLayout | Viewer | C++ | GUI | Active | ✅ Active | GDSII inspection |
| SKY130 PDK | Process | Various | N/A | Mature | ✅ Largest | 130nm digital |
| GF180 PDK | Process | Various | N/A | Mature | ✅ Growing | 180nm analog |
| Caravel | Harness | Verilog | N/A | Active | ✅ Large | Free tapeout |

### 3. Circuit Simulation Tools

| Tool | Type | Speed | GUI | Large Circuits | Verilog-A | Best For |
|------|------|-------|-----|-----------------|-----------|----------|
| ngspice | SPICE | Medium | CLI | Up to 100k nodes | ✅ OSDI | Standard production |
| Xyce | SPICE | Fast (parallel) | CLI | Millions of nodes | ✅ | Large crossbars |
| QUCS-S | GUI | Medium | ✅ | Small-medium | ✅ | Schematic capture |
| PySpice | Python | Medium | CLI | Medium | ✅ | Parametric sweeps |
| KiCad | GUI | Medium (ngspice) | ✅ | Medium | ✅ | Professional design |
| OpenVAF | Compiler | N/A | CLI | N/A | ✅ | Model compilation |

### 4. Circuit Analysis Libraries

| Tool | Type | Speed | Accuracy | Crossbar | Python | Best For |
|------|------|-------|----------|----------|--------|----------|
| badcrossbar | Solver | Very fast | High | ✅ Native | ✅ | IR drop + sneak paths |
| Lcapy | Symbolic | Fast | Exact | Limited | ✅ | Transfer functions |
| CrossSim | Simulator | Fast (GPU) | High | ✅ Native | ✅ | Full MVM validation |
| MemTorch | Framework | Fast (GPU) | High | ✅ Native | ✅ | NN + device co-sim |

### 5. Physics Visualization Tools

| Tool | Type | Real-time | 3D | Interactive | Publication | Best For |
|------|------|-----------|----|-----------|-----------|----|
| Matplotlib | Static | ❌ | ⚠️ | ❌ | ✅ | Publication plots |
| Plotly | Interactive | ❌ | ✅ | ✅ | ⚠️ | Web dashboards |
| VPython | Animation | ✅ | ✅ | ✅ | ❌ | Physics education |
| PyVista | 3D Field | ✅ | ✅ | ✅ | ✅ | FEM/field visualization |
| ParaView | Scientific | ✅ | ✅ | ✅ | ✅ | Large-scale data |

### 6. Hysteresis Modeling Tools

| Tool | Model Type | Accuracy | Speed | Inverse | Wake-up | Best For |
|------|-----------|----------|-------|---------|---------|----------|
| python-preisach | Preisach | Good | Slow | ❌ | ❌ | Educational |
| Preisachmodel | Preisach | Excellent | Medium | ✅ | ❌ | Research |
| ferro_scripts | Garrity | Good | Fast | ❌ | ❌ | P-E curves |
| PFECAP | Preisach | Good | Very fast | ❌ | ✅ | SPICE circuits |

### 7. Memristor/RRAM Tools

| Tool | Device | Crossbar | NN | GPU | Python | Best For |
|------|--------|----------|-----|-----|--------|----------|
| MemTorch | Linear drift | ✅ | ✅ | ✅ | ✅ | PyTorch integration |
| CrossSim | Configurable | ✅ | ✅ | ✅ | ✅ | Sandia reference |
| badcrossbar | Resistive | ✅ | ❌ | ❌ | ✅ | IR drop analysis |
| VTEAM | V-T | ⚠️ | ⚠️ | ❌ | ✅ | RRAM models |

### 8. NN Hardware Mapping Tools

| Tool | Framework | Bits | Hardware-aware | GPU | Best For |
|------|-----------|------|-----------------|-----|----------|
| Brevitas | PyTorch | Arbitrary | ✅ | ✅ | 5-bit FeCIM |
| QKeras | Keras/TF | Fixed | ✅ | ⭐ | Fixed-bit design |
| IBM AIHWKit | PyTorch | Configurable | ✅ | ✅ | In-memory compute |
| HAWQ | PyTorch | Mixed | ✅ | ✅ | Optimal quantization |
| CrossSim | PyTorch | Configurable | ✅ | ✅ | Crossbar mapping |

### 9. Scientific Computing Tools

| Tool | Type | HPC | GPU | FeCIM Use | Maturity |
|------|------|-----|-----|-----------|----------|
| FEniCS | FEM | ✅ | ⚠️ | Limited | Mature |
| MOOSE | Multiphysics | ✅ | ⚠️ | Base for FERRET | Mature |
| SfePy | FEM | ⚠️ | ❌ | Limited | Stable |
| pymatgen | Materials | ⚠️ | ❌ | DFT screening | Active |
| ASE | Atoms | ⚠️ | ⚠️ | Atomic structure | Active |

### 10. Data Acquisition Tools

| Tool | Type | Supported Hardware | Python | Best For |
|------|------|-------------------|--------|----------|
| PyVISA | Generic | GPIB, USB, Serial | ✅ | SMU/oscilloscope control |
| PyMeasure | Abstraction | 50+ drivers | ✅ | Easy instrument switching |
| QCoDeS | Framework | High-end instruments | ✅ | Professional measurements |
| nidaqmx | Driver | NI hardware only | ✅ | NI-specific systems |

---

## Quick Decision Guide

This section helps you find the right tool for common FeCIM development tasks.

### I want to simulate ferroelectric material properties

**Task: Generate P-E curves for HfO₂-ZrO₂**

| Scenario | Best Tool | Time | Difficulty |
|----------|-----------|------|------------|
| Quick visualization | `ferro_scripts` (Python) | 5 min | Easy |
| Fit experimental data | `Ferro` (Python) | 15 min | Easy |
| Domain structure | `FerroX` (C++/GPU) | 1-2 hours | Medium |
| Phase diagram | `Q-POP-Thermo` (Python) | 30 min | Easy |
| First-principles validation | `pymatgen` + VASP | 1-7 days | Hard |

**Recommended flow:**
```
1. Start with ferro_scripts for quick P-E curves
2. Use Ferro to fit experimental data
3. Use FerroX if domain structure needed
4. Use PFECAP when ready for SPICE integration
```

---

### I want to verify crossbar IR drop and sneak paths

**Task: Simulate realistic MVM with non-idealities**

| Tool | Strengths | Weaknesses | Time |
|------|-----------|-----------|------|
| `badcrossbar` | Exact solution, fast | Limited to IR drop | 10 min |
| `CrossSim` | GPU acceleration, flexible | Complex setup | 30 min |
| `MemTorch` | PyTorch integration | Higher abstraction level | 20 min |
| `ngspice` (full circuit) | Most accurate | Very slow for large arrays | Hours |

**Recommended flow:**
```
1. Quick validation: badcrossbar (minutes)
2. Full simulation with NN: CrossSim (minutes-hours)
3. Circuit-level validation: ngspice on small sections
```

---

### I want to design a chip peripheral (ADC/DAC)

**Task: Design and tapeout a 130nm ADC**

| Phase | Best Tool | Time | Effort |
|-------|-----------|------|--------|
| Schematic capture | `KiCad` or `QUCS-S` | 1-2 hours | Low |
| Simulation | `ngspice` + `PySpice` | 2-4 hours | Medium |
| Layout design | `Magic` (custom cells) | 4-8 hours | High |
| Verification (DRC/LVS) | `Magic` + `Netgen` + `KLayout` | 1-2 hours | Low |
| Synthesis (if digital part) | `Yosys` | 30 min | Low |
| Place & Route | `OpenLane` | 30 min - 2 hours | Low |
| Final GDSII | `KLayout` (inspect) | 15 min | Low |

**Recommended flow:**
```
KiCad (schematic) → PySpice (sim) → ngspice (validation)
    ↓
Magic (custom analog cells)
    ↓
OpenLane (digital control logic)
    ↓
Merge layouts + OpenLane final flow
    ↓
Magic/KLayout (DRC/LVS)
    ↓
Submit to Caravel/Efabless
```

---

### I want to train a neural network for FeCIM hardware

**Task: Quantize MNIST to 5-bit (30 levels) with device non-idealities**

| Tool | Best For | Ease | Time |
|------|----------|------|------|
| `Brevitas` | Arbitrary bit-widths (5-bit native) | ⭐⭐⭐⭐ | 1 hour |
| `QKeras` | Fixed bit-widths (8, 16-bit) | ⭐⭐⭐⭐ | 1 hour |
| `IBM AIHWKit` | Hardware-aware training | ⭐⭐⭐ | 2 hours |
| `CrossSim` | Full crossbar mapping | ⭐⭐⭐ | 2-3 hours |

**Recommended flow:**
```
1. Brevitas for 5-bit quantization training
   model = quantize_to_5bit(model)
2. Validate against CrossSim
3. Integrate into MemTorch for crossbar
4. Test under device variations
```

---

### I want to design and verify a ferroelectric capacitor model

**Task: Create and validate a FeCIM cell SPICE model**

| Step | Tool | Time | Output |
|------|------|------|--------|
| Write Verilog-A | Text editor | 30 min | `.va` file |
| Compile to OSDI | `OpenVAF` | 5 min | `.osdi` plugin |
| Test in SPICE | `ngspice` | 15 min | `.log` file |
| Verify vs. measurements | `PySpice` script | 30 min | Report |
| Integrate into circuits | `ngspice` netlists | 15 min | Working model |

**Recommended flow:**
```
OpenVAF (compile) → ngspice (validate) → PySpice (automate)
```

---

### I want to visualize ferroelectric domain evolution

**Task: Real-time 3D animation of domain switching**

| Tool | Real-time | 3D | Quality | Best For |
|------|-----------|-----|---------|----------|
| `VPython` | ✅ | ✅ | Good | Educational |
| `PyVista` | ✅ | ✅ | Excellent | Scientific |
| `ParaView` | ✅ | ✅ | Professional | Large datasets |
| `Matplotlib` | ❌ | Limited | Publication | Static plots |

**Recommended flow:**
```
FerroX (simulate) → Export HDF5/VTK → PyVista/ParaView (visualize)
```

---

### I want to validate module2 (crossbar) against real physics

**Task: Verify behavioral simulation matches realistic device/circuit models**

| Comparison Level | Tool A | Tool B | Time |
|------------------|--------|--------|------|
| Device-level | PFECAP (Preisach) | ngspice | 1 hour |
| Array-level | badcrossbar (IR drop) | ngspice (full circuit) | 2-3 hours |
| System-level | CrossSim (NN) | MemTorch (GPU) | 1-2 hours |

**Recommended validation flow:**
```
module2 behavioral model
    ↓
Compare vs. badcrossbar (IR drop validation)
    ↓
Compare vs. CrossSim (full MVM validation)
    ↓
Sample ngspice circuit-level validation
    ↓
Report accuracy metrics
```

---

### I want to run my entire workflow in Docker

**Task: Get all tools in one container**

**Answer: Use `IIC-OSIC-TOOLS`**

```bash
docker pull hpretl/iic-osic-tools:latest
docker run -it -v $(pwd):/root/designs hpretl/iic-osic-tools:latest

# Inside container:
yosys              # Synthesis
openroad           # Place & route
magic -T sky130A   # Layout verification
ngspice            # Circuit simulation
klayout            # GDSII viewing
```

**What's included:**
- ✅ Yosys, OpenROAD, OpenLane (EDA)
- ✅ Magic, Netgen, KLayout (verification)
- ✅ ngspice, Xyce (SPICE)
- ✅ SKY130, GF180, IHP130 PDKs
- ✅ JupyterLab (interactive notebooks)

---

## Detailed Tool Assessments

### Ferroelectric Simulation Category

#### FERRET (Recommended for: Domain dynamics, coupling physics)

**Strengths:**
- Coupled electro-mechanical physics (piezoelectric effects)
- Landau-Devonshire theory well-established for many materials
- Can import HfO₂ parameters from literature
- MOOSE framework enables parallelization

**Limitations:**
- Steep learning curve (MOOSE framework)
- CPU-only, slow for large domains (>100³ cells)
- Limited built-in HfO₂ support

**When to use:** If you need domain wall mechanics or material comparison across ABO₃/pyrochlore structures.

**When NOT to use:** For quick visualization or real-time interactive demos.

---

#### FerroX (Recommended for: 3D device simulation, GPU acceleration)

**Strengths:**
- Native HfO₂ support
- GPU acceleration (15× speedup vs CPU)
- Time-dependent Ginzburg-Landau equation (physically accurate)
- Adaptive mesh refinement (focuses computation on domain walls)

**Limitations:**
- Steeper C++ learning curve for customization
- Requires CUDA 11.2+ (GPU dependency)
- Limited experimental material database

**When to use:** If you need fast, accurate 3D ferroelectric device simulations or paper-quality domain structure images.

**When NOT to use:** If you don't have access to NVIDIA GPU or need quick prototype iterations.

---

### EDA Tools Category

#### OpenLane (Recommended for: Complete RTL-to-GDSII automation)

**Strengths:**
- Fully automated flow (600+ successful tapeouts)
- Orchestrates Yosys, OpenROAD, Magic, Netgen in sequence
- Configuration-driven (TCL or Python)
- Free, open-source, active community

**Limitations:**
- Less flexible for custom optimization per tool
- Requires understanding of each underlying tool for troubleshooting
- PDK-specific (each new process needs configuration)

**When to use:** For standard digital designs, quick prototyping, or first tapeouts.

**When NOT to use:** For highly optimized analog designs requiring fine-grained tool control.

---

#### Magic + Netgen (Recommended for: Custom cell design, verification)

**Strengths:**
- Precise layout control (hand-drawn cells)
- Accurate DRC checking
- LVS verification with Netgen
- SPICE netlist extraction with parasitic R/C

**Limitations:**
- Manual layout is time-consuming
- Steep learning curve for new users
- No automated routing (use OpenROAD instead)

**When to use:** For custom ferroelectric cells, analog building blocks, or when you need exact control over layout geometry.

**When NOT to use:** For digital logic (use Yosys/OpenLane instead).

---

### Circuit Simulation Category

#### ngspice (Recommended for: Production SPICE simulations, OSDI models)

**Strengths:**
- Full SPICE3 standard compliance
- OSDI (Open Source Device Interface) support for Verilog-A models
- Active development (weekly releases)
- Excellent convergence handling

**Limitations:**
- Single-core (use Xyce for massive circuits)
- Requires SPICE syntax (steeper than Python)
- More setup than PySpice

**When to use:** For production-quality circuit validation, custom device models, or when maximum compatibility is needed.

**When NOT to use:** For quick prototyping (use PySpice instead) or massive circuits (use Xyce).

---

#### PySpice (Recommended for: Parametric sweeps, automation, integration with Python)

**Strengths:**
- Python interface (easy automation)
- Parametric sweeps built-in
- Can integrate with Jupyter notebooks
- Cleaner syntax than raw SPICE

**Limitations:**
- Wraps ngspice (inherits same convergence issues)
- Less documentation than ngspice raw
- Python overhead for real-time interactive use

**When to use:** For automated design optimization, parametric studies, or when integrating with Python ML pipelines.

**When NOT to use:** For interactive SPICE console exploration (use ngspice or GnuCap instead).

---

### Circuit Analysis Category

#### badcrossbar (Recommended for: IR drop, sneak paths, passive crossbars)

**Strengths:**
- Purpose-built for passive crossbar arrays
- Fast (microsecond solve times even for 1024×1024 arrays)
- Exact solution to modified nodal analysis
- Well-documented paper (SoftwareX 2020)

**Limitations:**
- Only solves for voltages/currents (no transient switching)
- Requires explicit conductance matrix setup
- Limited to DC or slow quasi-static analysis

**When to use:** For crossbar non-ideality characterization, IR drop budgeting, or sneak path estimation.

**When NOT to use:** For transient switching analysis (use CrossSim or ngspice).

---

#### CrossSim (Recommended for: GPU-accelerated MVM, full system simulation)

**Strengths:**
- GPU acceleration (handles 1024×1024 arrays in real-time)
- Comprehensive non-ideality models (drift, variation, noise)
- PyTorch integration for NN workflows
- Well-validated against measurements

**Limitations:**
- Behavioral abstraction (less circuit detail than ngspice)
- Large learning curve for new users
- Less documentation than academic papers

**When to use:** For system-level simulation, NN accuracy evaluation, or when you need GPU speedup.

**When NOT to use:** For detailed circuit-level physics (use ngspice).

---

### Hysteresis Modeling Category

#### ferro_scripts (Recommended for: Quick P-E curves, Garrity model)

**Strengths:**
- Fastest way to generate realistic P-E curves
- Based on peer-reviewed Garrity et al. 2014 model
- Simple YAML configuration
- BaTiO₃ and CrCA pre-configured

**Limitations:**
- No hysteresis dynamics or frequency dependence
- No wake-up/fatigue models
- Limited to materials in YAML configs

**When to use:** For quick visualization, teaching, or when you need Garrity model specifically.

**When NOT to use:** For detailed time-dependent behavior or custom materials (use FerroX instead).

---

#### PFECAP (Recommended for: SPICE integration, 1T1C memory cells)

**Strengths:**
- Directly usable in ngspice/Xyce netlists
- Circuit-compatible (fast simulation)
- Verilog-A source (modifiable)
- Pre-configured for HfO₂

**Limitations:**
- Preisach model (macroscopic, no domain physics)
- Limited parameter fitting guidance
- Verilog-A compilation required (OpenVAF)

**When to use:** When you need ferroelectric behavior in circuit simulations or 1T1C cell designs.

**When NOT to use:** For detailed material characterization (use FerroX or ferro_scripts).

---

### NN Hardware Mapping Category

#### Brevitas (Recommended for: Arbitrary-precision quantization, 5-bit FeCIM)

**Strengths:**
- Supports ANY bit-width (1-32 bits)
- Native 5-bit support (32 levels ≈ FeCIM 30 levels)
- Full PyTorch integration
- Excellent documentation and examples

**Limitations:**
- PyTorch-only (not TensorFlow)
- No analog device modeling (use MemTorch or CrossSim for that)
- Quantization-aware training more complex than post-training

**When to use:** For training 5-bit networks or when you need arbitrary precision quantization.

**When NOT to use:** If you're locked into TensorFlow (use QKeras instead).

---

#### IBM AIHWKit (Recommended for: Hardware-aware training with device non-idealities)

**Strengths:**
- Models realistic device non-idealities (noise, conductance drift)
- Hardware-aware training (backprop through device models)
- PyTorch native
- Well-supported by IBM

**Limitations:**
- Higher learning curve than Brevitas
- More computational overhead per training step
- Less documentation for custom hardware models

**When to use:** When you want to train networks that actually work on FeCIM hardware.

**When NOT to use:** For quick prototype quantization (use Brevitas first).

---

## Integration Recommendations

### Recommended Workflow: Module 1 (Hysteresis)

```
Step 1: Generate P-E curves
├─ Input: HfO₂ literature parameters (Pr, Ec, etc.)
├─ Tool: ferro_scripts OR FerroX
└─ Output: P-E curves (JSON/CSV)

Step 2: Fit Preisach model
├─ Input: P-E curves from Step 1
├─ Tool: Ferro or Preisachmodel
└─ Output: Hysteron distribution

Step 3: Implement in Go
├─ Input: Hysteron parameters
├─ Tool: Your code + module1-hysteresis
└─ Output: Interactive visualization

Step 4: Validate against experiments
├─ Input: Real measurement data (if available)
├─ Tool: Ferro (extract parameters)
├─ Compare: Simulation vs. measurement
└─ Output: Accuracy report
```

### Recommended Workflow: Module 2 (Crossbar)

```
Step 1: Define array geometry
├─ Input: 32×32 array, R_on/R_off, line resistances
├─ Tool: Python + badcrossbar setup
└─ Output: Conductance matrix

Step 2: Quick IR drop validation
├─ Input: Conductance matrix
├─ Tool: badcrossbar
└─ Output: Voltage drop at each node

Step 3: Full MVM simulation
├─ Input: Weight matrix, voltage levels
├─ Tool: CrossSim
└─ Output: MVM accuracy, power

Step 4: Circuit-level validation (sampling)
├─ Input: Select few cells from array
├─ Tool: ngspice + PFECAP models
└─ Output: Transistor-level accuracy

Step 5: Implement in Go
├─ Input: Results from Steps 2-4
├─ Tool: Your code + module2-crossbar
└─ Output: Behavioral simulator
```

### Recommended Workflow: Module 3 (MNIST NN)

```
Step 1: Quantize network to 5-bit
├─ Input: Pre-trained FP32 model
├─ Tool: Brevitas (5-bit QAT)
└─ Output: 5-bit quantized model

Step 2: Map to FeCIM crossbar
├─ Input: 5-bit weights, 30-level quantization
├─ Tool: CrossSim + MemTorch
└─ Output: MVM accuracy with device variations

Step 3: Validate under variations
├─ Input: Mapped model from Step 2
├─ Tool: MemTorch or IBM AIHWKit
├─ Variation types: conductance drift, noise, process variation
└─ Output: Worst-case accuracy

Step 4: Implement in Go
├─ Input: Final validated network
├─ Tool: Your code + module3-mnist
└─ Output: Interactive NN demo
```

### Recommended Workflow: Module 4 (Circuits)

```
Step 1: Design ADC/DAC schematic
├─ Input: Requirements (resolution, speed)
├─ Tool: KiCad or QUCS-S
└─ Output: SPICE netlist

Step 2: Simulate with ngspice
├─ Input: SPICE netlist
├─ Tool: ngspice or PySpice
├─ Measurements: INL, DNL, settling time
└─ Output: Performance report

Step 3: Layout in SKY130
├─ Input: Verified circuit
├─ Tool: Magic (analog cells) + OpenLane (digital)
└─ Output: GDSII

Step 4: Verify (DRC/LVS)
├─ Tool: Magic + Netgen
└─ Status: OK or design iteration

Step 5: Extract parasitic and validate
├─ Input: GDSII from Step 3
├─ Tool: Magic (extract) + ngspice (sim with parasitic)
└─ Output: Real-world performance
```

### Recommended Workflow: Module 6 (EDA)

```
Step 1: Create digital control logic
├─ Input: State machine definition (row/column selection)
├─ Tool: Verilog coding
└─ Output: .v files

Step 2: Simulate behavior
├─ Input: .v files + testbench
├─ Tool: Yosys (synth) + iverilog/verilator (sim)
└─ Output: Functional verification

Step 3: Synthesize and place & route
├─ Input: RTL from Step 1
├─ Tool: OpenLane (automated)
└─ Output: GDSII

Step 4: Integrate with analog peripherals
├─ Tool: Magic + manual layout merging
└─ Output: Complete chip GDSII

Step 5: Verify complete design
├─ Tool: Magic (DRC) + Netgen (LVS)
└─ Output: Ready for Caravel/Efabless
```

---

## Tool Selection Flowchart

```
START: What do you want to do?
│
├─→ Simulate ferroelectric material
│   ├─→ Quick P-E curves? → ferro_scripts
│   ├─→ Fit experimental data? → Ferro
│   ├─→ 3D domain structure? → FerroX
│   └─→ Phase diagram? → Q-POP-Thermo
│
├─→ Analyze crossbar array
│   ├─→ IR drop & sneak paths? → badcrossbar
│   ├─→ Full system with NN? → CrossSim
│   └─→ Circuit-level details? → ngspice + PFECAP
│
├─→ Design a peripheral circuit
│   ├─→ Schematic capture? → KiCad or QUCS-S
│   ├─→ Simulation? → ngspice or PySpice
│   ├─→ Layout? → Magic
│   └─→ Synthesis (digital)? → Yosys + OpenLane
│
├─→ Train neural network for FeCIM
│   ├─→ Quantize to 5-bit? → Brevitas
│   ├─→ Hardware-aware training? → IBM AIHWKit
│   └─→ Map to crossbar? → CrossSim or MemTorch
│
├─→ Visualize results
│   ├─→ Static plots? → Matplotlib
│   ├─→ Interactive 3D? → PyVista
│   └─→ Real-time animation? → VPython
│
└─→ Design a full chip
    ├─→ All-in-one container? → IIC-OSIC-TOOLS
    ├─→ Automated flow? → OpenLane
    ├─→ Free tapeout? → Caravel + Efabless
    └─→ Fine control? → Yosys + OpenROAD + Magic individually
```

---

## Comparative Analysis: Head-to-Head Matchups

### FerroX vs FERRET (for phase-field simulation)

| Metric | FerroX | FERRET |
|--------|--------|--------|
| **Speed** | 15× faster (GPU) | Baseline (CPU) |
| **Model** | TDGL (field-based) | Landau (phenomenological) |
| **Setup time** | 30 min | 2-3 hours |
| **HfO₂ support** | Native | Via parameters |
| **Documentation** | Good (paper + code) | Good (MOOSE docs) |
| **Learning curve** | Medium | Steep |
| **Best for** | Real-time visualization | Material comparison |

**Verdict:** Use FerroX for speed and HfO₂ specificity. Use FERRET if you need coupled electromechanics or material comparison.

---

### CrossSim vs ngspice (for crossbar validation)

| Metric | CrossSim | ngspice |
|--------|----------|---------|
| **Speed** | GPU-accelerated | Single-core |
| **Array size** | 1024×1024+ | ~32×32 max |
| **Physics detail** | Medium (behavioral) | High (transistor-level) |
| **GPU required** | Yes | No |
| **Learning curve** | Medium | Medium |
| **Best practice** | System-level validation | Circuit-level verification |

**Verdict:** Use CrossSim for large arrays and NN workflows. Use ngspice for detailed circuit validation of small sections.

---

### Brevitas vs QKeras (for quantization)

| Metric | Brevitas | QKeras |
|--------|----------|--------|
| **Framework** | PyTorch | TensorFlow/Keras |
| **Arbitrary bits** | ✅ Yes | ❌ Fixed |
| **5-bit native** | ✅ Yes | Workaround only |
| **Hardware models** | Limited | Limited |
| **Documentation** | Excellent | Good |
| **Community** | Growing | Established |

**Verdict:** Use Brevitas for FeCIM (native 5-bit). Use QKeras if locked into TensorFlow.

---

### OpenLane vs Yosys+OpenROAD (for chip design)

| Metric | OpenLane | Yosys + OpenROAD |
|--------|----------|------------------|
| **Automation** | Full (one command) | Manual orchestration |
| **Flexibility** | Less (flow is fixed) | More (each tool independent) |
| **Troubleshooting** | Harder (abstracted) | Easier (see each step) |
| **Time** | 10-30 min | 30 min - 2 hours |
| **Learning** | Easier (higher level) | Harder (must know each tool) |

**Verdict:** Use OpenLane for quick prototypes. Use individual tools if you need fine control.

---

## Conclusion

This comparison matrix provides a structured way to select tools for FeCIM projects. Key principles:

1. **Start with the simplest tool for your task** (e.g., ferro_scripts before FerroX)
2. **Use GPUs when available** (FerroX, CrossSim, PyTorch tools)
3. **Validate behavioral models against circuits** (ngspice benchmark)
4. **Automate with Python** (PySpice, Brevitas, CrossSim APIs)
5. **Use containers when possible** (IIC-OSIC-TOOLS for reproducibility)

For questions or tool updates, see `docs/opensource-tools/README.md` or the individual tool documentation files.

---

**Document maintained by:** FeCIM Documentation Team
**Last verified:** January 27, 2026
**Status:** Comprehensive reference matrix (50+ tools, 10 categories)
**Feedback:** See `docs/about/Contributing.md`
