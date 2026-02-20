<!-- Category: Tools | Module: module2-crossbar | Reading time: ~6 min -->
# Module 2 Open-Source Tools: Crossbar Simulation Ecosystem

> A guide to open-source tools for crossbar array simulation, neural network
> hardware mapping, and analog computing -- and how they integrate with this module.

---

## Tools Used by This Module

| Tool | Purpose | License |
|------|---------|---------|
| Go toolchain | Build and runtime | BSD-style |
| Fyne | GUI rendering | BSD-3-Clause |

---

## Crossbar Simulation Tools

### CrossSim (Sandia National Labs)

Crossbar array simulator with device non-idealities for analog neural
networks. Device variation modeling, noise injection, inference accuracy
analysis.

**Integration**: Export weights from this module as NumPy arrays, load in
CrossSim, compare MVM outputs. Useful for validating results against an
established reference.

### BadCrossbar (Python)

Computes currents and voltages in passive crossbar arrays with full IR drop
analysis. More detailed circuit-level IR drop than this module's simplified
analytical model.

**Use case**: Verify this module's IR drop analysis against a more detailed
circuit solver.

### NeuroSim (Georgia Tech)

Device-to-architecture benchmark for neuromorphic computing. Provides
detailed peripheral circuit models (DAC, ADC, sense amplifiers) and
energy/area estimation.

**Integration**: Export array configuration and conductance distributions,
use NeuroSim for production-accurate energy and area estimates.

### MNSIM

Multi-level neural network simulator for computing-in-memory. Supports
multi-bit cells and hybrid precision.

### AIHWKIT (IBM)

PyTorch toolkit for analog AI hardware simulation. Analog weight devices
with drift, noise, and hardware-aware training.

**Integration**: Export weights from this module, define custom FeFET device
models in AIHWKIT, train neural networks with hardware-aware optimization.

---

## Circuit Simulation

### ngspice + Verilog-A

Open-source SPICE simulator. Export SPICE netlists from this module for
detailed circuit-level transient analysis, settling time verification,
and parasitic effects.

### Xyce (Sandia)

High-performance parallel circuit simulator. Better convergence than ngspice
for large crossbar arrays and stiff problems.

---

## Data Analysis Tools

For post-processing exported data from this module:

- **NumPy/SciPy**: Weight distribution analysis, statistical characterization
- **Pandas**: CSV data analysis (conductance tables, MVM results)
- **Matplotlib/Seaborn**: Heatmaps, scatter plots, distribution histograms

---

## Tool Comparison Matrix

| Tool | Purpose | Language | Use Case |
|------|---------|----------|----------|
| This Module | Full crossbar sim + GUI | Go | Education, visualization |
| CrossSim | Array-level accuracy | Python | Validation reference |
| BadCrossbar | IR drop analysis | Python | Physics verification |
| NeuroSim | Architecture estimation | C++ | Energy/area benchmarks |
| AIHWKIT | Analog training | Python | PyTorch integration |
| ngspice | Circuit simulation | C | Detailed circuit analysis |
| Xyce | Parallel circuit sim | C++ | Large-scale simulation |

---

## Recommended Workflows

**Validation**: Export weights --> run same MVM in CrossSim --> compare
outputs (should match within 1-2%).

**Energy estimation**: Export configuration --> NeuroSim for detailed
peripheral circuits --> compare with this module's simplified energy model.

**Circuit verification**: Export SPICE netlist --> add realistic parasitics
--> run transient analysis in ngspice --> verify timing and settling.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
