# Open Source Tools for FeCIM Development

**Comprehensive index and navigation guide for ferroelectric compute-in-memory (FeCIM) tools and resources**

---

## Overview

This documentation catalogs open-source tools relevant to ferroelectric compute-in-memory systems. Whether you're simulating hysteresis loops, designing crossbar arrays, training neural networks, or characterizing devices in the lab, this guide connects you to the right tools.

> **Note:** Counts are approximate and tool claims are reported from sources, not verified by this project.

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

### Why This Matters for FeCIM Development

FeCIM systems span multiple computational domains:

- **Physics simulation** - Hysteresis, phase-field dynamics, polarization switching
- **Circuit analysis** - Crossbar IR drop, sneak paths, device non-idealities
- **Hardware design** - DAC/ADC peripherals, EDA tools for chip layout
- **Neural networks** - Quantization to 30 levels (demo baseline), hardware-aware training
- **Device characterization** - Lab equipment control, P-E measurement, drift analysis

Each domain has specialized tools. This documentation organizes them by category with practical examples and integration strategies for the FeCIM Lattice Tools project.

---

## Documentation Index & Navigation

### Quick Links by Task

| I need to... | Best documentation | Time investment |
|---|---|---|
| Simulate ferroelectric behavior | [Ferroelectric Simulation Tools](./ferroelectric-simulation-tools.md) | 1-2 hours |
| Design crossbar arrays | [Circuit Analysis Libraries](./circuit-analysis-libraries.md) | 2-4 hours |
| Verify circuits with SPICE | [Circuit Simulation Tools](./circuit-simulation-tools.md) | 1-2 hours |
| Design chip layout | [EDA Tools](./eda-tools.md) | 4-8 hours |
| Train neural networks on hardware | [Neural Network Hardware Mapping Tools](./nn-hardware-mapping-tools.md) | 2-4 hours |
| Analyze P-E loops | [Hysteresis Modeling Tools](./hysteresis-modeling-tools.md) | 1-2 hours |
| Visualize results | [Physics Visualization Tools](./physics-visualization-tools.md) | 1-2 hours |
| Model memristor devices | [Memristor and RRAM Tools](./memristor-rram-tools.md) | 2-3 hours |
| Run first-principles DFT | [Scientific Computing Tools](./scientific-computing-tools.md) | 4-8 hours |
| Control lab equipment | [Data Acquisition Tools](./data-acquisition-tools.md) | 1-2 hours |

---

## Complete Documentation Index Table

| Document | Category | Purpose | # Tools |
|----------|----------|---------|---------|
| **[ferroelectric-simulation-tools.md](./ferroelectric-simulation-tools.md)** | Physics | Phase-field simulation, domain dynamics, hysteresis | 8 tools |
| **[eda-tools.md](./eda-tools.md)** | EDA/Design | RTL-to-GDSII flow, chip design, physical verification | 9 tools |
| **[circuit-simulation-tools.md](./circuit-simulation-tools.md)** | Circuits | SPICE simulation, peripheral design, analog verification | 7 tools |
| **[circuit-analysis-libraries.md](./circuit-analysis-libraries.md)** | Analysis | Crossbar simulation, nodal analysis, symbolic computation | 7 tools |
| **[physics-visualization-tools.md](./physics-visualization-tools.md)** | Visualization | 2D/3D plotting, interactive visualization, animation | 10 tools |
| **[hysteresis-modeling-tools.md](./hysteresis-modeling-tools.md)** | Hysteresis | Preisach models, FORC analysis, loop fitting | 6 tools |
| **[memristor-rram-tools.md](./memristor-rram-tools.md)** | Devices | Memristor/RRAM simulation, crossbar frameworks | 5 tools |
| **[nn-hardware-mapping-tools.md](./nn-hardware-mapping-tools.md)** | AI/ML | Quantization, hardware-aware training, mapping | 6 tools |
| **[scientific-computing-tools.md](./scientific-computing-tools.md)** | Materials | DFT, FEM, materials databases | 8 tools |
| **[data-acquisition-tools.md](./data-acquisition-tools.md)** | Instrumentation | Lab equipment control, measurement automation | 6 tools |

**Total tools documented: 72 across 10 categories**

---

## Tool Count Summary

```
Ferroelectric Simulation          8 tools (FERRET, FerroX, pymatgen, Q-POP, Ferro, PFECAP, negativec, feram)
EDA & Chip Design                9 tools (Yosys, OpenROAD, Magic, Netgen, KLayout, OpenLane, SKY130, etc.)
Circuit Simulation & Analysis    14 tools (ngspice, Xyce, PySpice, QUCS-S, badcrossbar, CrossSim, MemTorch)
Visualization & Plotting         10 tools (Matplotlib, Plotly, VPython, PyVista, Mayavi, K3D, Napari, etc.)
Hysteresis & Device Modeling      6 tools (python-preisach, Preisachmodel, Ferro, hysteresis pkg)
Neural Network Tools              6 tools (Brevitas, AIHWKIT, MemTorch, IBM AIHWKIT, TensorFlow QAT, PyTorch)
Materials & Scientific          8 tools (FEniCSx, VASP, QUANTUM ESPRESSO, pymatgen, ASE, ABINIT)
Instrumentation & Measurement     6 tools (PyVISA, PyMeasure, QCoDeS, Lakeshore, AutoDSP)

TOTAL: 72+ open-source tools documented
```

---

## Getting Started: Your First Steps

### New to FeCIM Tools? Start Here

**1. Pick your use case:**
- Simulating hysteresis loops? → Start with [Ferroelectric Simulation Tools](./ferroelectric-simulation-tools.md)
- Designing crossbar circuits? → Start with [Circuit Analysis Libraries](./circuit-analysis-libraries.md)
- Training networks? → Start with [Neural Network Hardware Mapping Tools](./nn-hardware-mapping-tools.md)

**2. Recommended first tool per category:**
- Ferroelectric physics: **FerroX** (GPU-accelerated, HfO₂-specific, well-documented)
- Crossbar simulation: **badcrossbar** (Python, MIT license, IR drop + sneak paths)
- Circuit simulation: **ngspice** (free, standard, excellent for SPICE validation)
- Visualization: **Matplotlib** + **Hysteresis package** (publication-quality, ferroelectric-specific)
- Neural networks: **Brevitas** (arbitrary bit-width quantization, PyTorch-native)

**3. Integration with FeCIM Lattice Tools:**

```
┌─────────────────────────────────────────────────────────┐
│         Your FeCIM Tool Selection                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Module 1: Hysteresis                                   │
│  ├─ Primary: Go implementation (Preisach)              │
│  └─ Validation: badcrossbar + ngspice                   │
│                                                         │
│  Module 2: Crossbar                                     │
│  ├─ Primary: Go MVM simulation                         │
│  └─ Validation: CrossSim, badcrossbar                   │
│                                                         │
│  Module 3: MNIST                                        │
│  ├─ Primary: Go neural network                         │
│  └─ Training: Brevitas, AIHWKIT                        │
│                                                         │
│  Module 4: Circuits                                     │
│  ├─ Primary: Go behavioral models                      │
│  └─ Verification: ngspice, PySpice                      │
│                                                         │
│  Module 6: EDA                                          │
│  ├─ Tool reference: OpenLane, Yosys, Magic            │
│  └─ Integration: SKY130 PDK, Caravel harness           │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## How to Use This Documentation

### For Each Tool Category

Each documentation file follows this structure:

1. **Overview** - What problems this category solves
2. **Individual tools** - Full details for 5-10 tools including:
   - Repository/Download link
   - Installation instructions
   - Key features and use cases
   - Code examples (copyable)
   - Relevance to FeCIM
   - Strengths and limitations
3. **Comparison matrix** - Side-by-side feature comparison
4. **Integration workflows** - How to combine multiple tools
5. **Troubleshooting** - Common issues and solutions

### Reading Strategy

**If you have 30 minutes:**
- Read the overview section of your chosen category
- Skim the "Recommended" and "Best For" sections

**If you have 1-2 hours:**
- Read one complete tool section with code examples
- Review the comparison matrix
- Try installing the tool locally

**If you have a full day:**
- Work through 2-3 tools from a category
- Run the provided code examples
- Adapt examples to your specific problem

---

## Cross-Document Reference Guide

### Looking for specific tool names?

| Tool | Document | Section |
|------|----------|---------|
| **badcrossbar** | Circuit Analysis | 1.1 (Crossbar simulators) |
| **CrossSim** | Circuit Analysis | 1.2 (Crossbar simulators) |
| **FerroX** | Ferroelectric Simulation | Section 2 |
| **Yosys** | EDA Tools | Section 2.1 (Synthesis) |
| **OpenLane** | EDA Tools | Section 4.1 (Integration) |
| **ngspice** | Circuit Simulation | Section 1 |
| **Brevitas** | Neural Networks | Section 1.1 (Quantization) |
| **AIHWKIT** | Neural Networks | Section 2 (Hardware-aware) |
| **Hysteresis package** | Physics Visualization | Section 2 |
| **PyVista** | Physics Visualization | Section 5 (3D visualization) |

---

## Tool Selection Decision Trees

### "I need to simulate hysteresis. Which tool?"

```
START: Do you have experimental P-E loop data?
├─ YES → Use Hysteresis package for ANALYSIS
│        └─ (ferroelectric-simulation-tools.md, Ferro section)
│
└─ NO → Do you want real-time 3D visualization?
   ├─ YES → Use FerroX (GPU phase-field)
   │        └─ (ferroelectric-simulation-tools.md, section 2)
   │
   └─ NO → Do you need analytical (symbolic) solution?
      ├─ YES → Use Q-POP-Thermo or Landau theory
      │        └─ (ferroelectric-simulation-tools.md, section 5)
      │
      └─ NO → Use Go implementation in Module 1
              └─ Validate against badcrossbar + ngspice
```

### "I need to design a crossbar circuit. Which tool?"

```
START: What's your focus?
├─ Non-idealities (IR drop, sneak paths) → badcrossbar
│  └─ (circuit-analysis-libraries.md, section 1.1)
│
├─ Full neural network on hardware → CrossSim or MemTorch
│  └─ (circuit-analysis-libraries.md, section 1.2 & 1.3)
│
├─ Symbolic analysis (transfer function) → Lcapy
│  └─ (circuit-analysis-libraries.md, section 2.1)
│
└─ SPICE-level validation → ngspice or PySpice
   └─ (circuit-simulation-tools.md, section 1 or 4)
```

---

## Related FeCIM Project Documentation

Beyond open-source tools, see also:

| Document | Purpose |
|----------|---------|
| `CLAUDE.md` | Project overview, physics constants, accuracy policy |
| `docs/development/scriptReference.md` | Function lookup, error resolution, decision trees |
| `docs/development/TESTING.md` | Test framework, running tests before commit |
| `docs/development/GUI/FYNE_NOTES.md` | GUI framework notes for module updates |
| `docs/video-transcripts/COSM_2025_AI_Hardware_Breakthrough/` | COSM 2025 transcript (archival) |
| `docs/comparison/HONESTY_AUDIT.md` | Scientific accuracy verification for all claims |

---

## Installation Quick Start

### Install All Essential Tools (Ubuntu/Debian)

```bash
# Physics simulation and analysis
pip install badcrossbar CrossSim FerroX pymatgen Preisachmodel hysteresis

# Circuit simulation
sudo apt-get install ngspice
pip install PySpice QUCS-S

# Visualization
pip install matplotlib plotly pyvista vpython k3d napari

# Neural networks
pip install brevitas aihwkit tensorflow torch

# Data acquisition
pip install pyvisa pymeasure qcodes

# EDA tools (requires more setup - see eda-tools.md)
docker pull efabless/openlane:latest
```

**Total time: 30-60 minutes depending on internet speed and compilation**

For GPU acceleration (CUDA), see individual tool documentation.

---

## Key Statistics & Coverage

### By Maturity Level

| Level | Count | Examples |
|-------|-------|----------|
| **Production (widely used)** | 25+ | ngspice, AIHWKIT, Yosys, OpenLane, VASP |
| **Research (published)** | 30+ | FerroX, badcrossbar, FERRET, MemTorch |
| **Active (maintained)** | 15+ | K3D-Jupyter, PyVista, Brevitas |
| **Maintenance (legacy)** | 5+ | Ahkab, feram, some older frameworks |

### By License Type

| License | Count | Examples |
|---------|-------|----------|
| **MIT/BSD** | 28 | badcrossbar, Hysteresis, Plotly, VPython |
| **GPL/LGPL** | 18 | ngspice, FERRET, MemTorch, FEniCSx |
| **Apache** | 6 | OpenLane, Yosys, Brevitas |
| **Custom/Open** | 14 | CrossSim, FerroX, VASP, QUANTUM ESPRESSO |

### Geographic Distribution of Tools

- **US-based** (National Labs, universities): 35 tools (Sandia, Berkeley, MIT, Stanford)
- **EU-based** (Academic, industry): 20 tools (Imperial College, TU Delft, Fraunhofer)
- **Community** (GitHub open-source): 17 tools

---

## Troubleshooting & Support

### Installation issues?
- See individual tool documentation's "Installation" section
- Check `TESTING.md` for environment setup
- Review tool-specific GitHub issues

### Integration questions?
- See "Recommended Workflows" in each tool's documentation
- Study code examples in this documentation
- Check `docs/development/scriptReference.md` for module APIs

### Need help?
- Tool-specific GitHub issues (most active)
- FeCIM project discussions in repository
- See CONTRIBUTING.md for code/doc contributions

---

## Contributing to This Documentation

Found a new tool? Updated an existing one? Please contribute:

1. Add tool details to appropriate `.md` file
2. Update tool count in this README
3. Test code examples before submission
4. Follow existing format and style

See `docs/about/Contributing.md` for full guidelines.

---

## Citation & References

**If you use tools based on this documentation, please cite:**

- **FeCIM Project**: FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
- **Individual tools**: Cite original papers per documentation links

---

**Last Updated: January 27, 2026**
**Maintained by:** FeCIM Documentation Team
**Total Tools Documented:** ~72 (approx.)
**Documentation Pages:** ~10 guides (approx.)
