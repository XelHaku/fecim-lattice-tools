<!-- Category: Tools | Module: module1-hysteresis | Reading time: ~7 min -->
# Module 1 Open-Source Tools: Ferroelectric Simulation Ecosystem

> A guide to open-source tools for ferroelectric hysteresis modeling,
> phase-field simulation, and device/circuit integration.

---

## Tools Used by This Module

| Tool | Purpose | License |
|------|---------|---------|
| Go toolchain | Build and runtime | BSD-style |
| Fyne | GUI rendering | BSD-3-Clause |
| Bubble Tea | Terminal UI mode | MIT |

---

## Hysteresis Modeling Libraries

### PyPreisach (Python)

Python implementation of the classical Preisach hysteresis model.

- First-order reversal curves (FORCs)
- Everett function computation
- Preisach plane visualization

**Comparison to this module**: PyPreisach uses explicit hysteron grids and is
more flexible for FORC calibration. This module uses a tanh-based Everett
approximation in Go, which is faster for real-time visualization but less
configurable.

### hystereloop (Python)

Analyzes hysteresis loops from experimental data: loop area, coercivity,
remanence extraction, FORC diagrams. Useful for processing measured data
alongside this module's simulations.

---

## Phase-Field Simulation Tools

### FerroX (GPU-Accelerated)

GPU-accelerated phase-field simulator for ferroelectric domain dynamics using
Time-Dependent Ginzburg-Landau (TDGL). Multi-GPU via AMReX. Suitable for
large-scale domain structure simulations (micron-scale), not real-time
visualization.

**Integration**: Export Landau coefficients from this module's materials,
use FerroX for detailed domain structure, import statistics back to calibrate
Preisach distribution.

### FERRET (MOOSE Framework)

Phase-field modeling of ferroic materials: Landau-Devonshire free energy,
coupled electro-mechanical problems, domain wall dynamics, strain effects.
Requires the MOOSE framework.

### PRISMS-PF

General-purpose phase-field code (C++) with adaptive mesh refinement,
adaptable for ferroelectrics.

---

## Device and Circuit Simulation

### ngspice + Verilog-A

Open-source SPICE simulator. Using OpenVAF to compile Verilog-A compact
models (e.g., FeFET models from Purdue), you can simulate ferroelectric
device circuits.

**Integration**: Export material parameters (Pr, Ps, Ec, tau) from this
module and generate Verilog-A parameter files for circuit simulation.

### OpenVAF

Open-source Verilog-A compiler that produces OSDI plugins for ngspice.
Bridges the gap between compact model descriptions and circuit simulation.

### Xyce (Sandia National Labs)

High-performance parallel circuit simulator with Verilog-A support. Better
convergence than ngspice for stiff problems and large circuits.

---

## Architecture-Level Simulators

### NeuroSim (Georgia Tech)

Device-to-architecture benchmark for neuromorphic computing: FeFET device
models, crossbar array simulation, energy/area estimation, neural network
accuracy analysis.

**Integration**: Export discrete levels (0-29) as conductance mappings,
import energy estimates for system-level analysis.

### CrossSim (Sandia)

Crossbar array simulator with device non-idealities. Useful for validating
this module's MVM results against an established reference.

### AIHWKIT (IBM)

PyTorch extension for analog AI hardware simulation: device-level noise,
drift modeling, hardware-aware training. Define custom devices from this
module's hysteresis curves.

---

## Tool Comparison Matrix

| Tool | Level | Hysteresis Model | Speed | Use Case |
|------|-------|-----------------|-------|----------|
| This Module (Go) | Macro | Preisach + L-K | Real-time | Education, visualization |
| PyPreisach | Macro | Preisach | Fast | Prototyping, FORC analysis |
| FerroX | Micro | TDGL | Slow (GPU) | Domain structure |
| FERRET | Micro | Landau | Slow | Coupled physics |
| ngspice + VA | Device | Custom | Medium | Circuit simulation |
| NeuroSim | Array | Simplified | Fast | Architecture estimation |
| CrossSim | Array | Custom | Fast | Array-level accuracy |
| AIHWKIT | Array | Custom | Fast | ML training |

---

## Learning Path

**Beginner**: Run this module, read PHYSICS.md, implement a simple Preisach
model in Python, compare loop shapes with different parameters.

**Intermediate**: Study the advanced Preisach implementation, export Preisach
distributions, connect to ngspice via SPICE export.

**Advanced**: Build a Verilog-A FeFET model using OpenVAF, couple with FerroX
for domain structure, integrate with NeuroSim for array-level analysis.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
