# Module 1: Hysteresis - Open-Source Tools

**Navigation:** [← Module 1 Index](./README.md) | [Physics](./physics.md) | [Features](./features.md)

---

## Evidence Status

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

---

## Overview

This document catalogs open-source tools for ferroelectric hysteresis modeling, domain dynamics, and related simulations. It covers how this module integrates with external tools and provides guidance for extending functionality.

---

## Table of Contents

1. [Dependencies Used by This Module](#dependencies-used-by-this-module)
2. [Hysteresis Modeling Libraries](#hysteresis-modeling-libraries)
3. [Phase-Field Simulation Tools](#phase-field-simulation-tools)
4. [Device/Circuit Simulation](#devicecircuit-simulation)
5. [Architecture-Level Simulators](#architecture-level-simulators)
6. [Visualization Tools](#visualization-tools)
7. [Integration Guide](#integration-guide)
8. [Tool Comparison Matrix](#tool-comparison-matrix)

---

## Dependencies Used by This Module

### Core Dependencies

| Tool | Purpose | License |
|------|---------|---------|
| **Go toolchain** | Build/runtime for simulator | BSD-style |
| **Fyne** | GUI rendering (default mode) | BSD-3-Clause |
| **Bubble Tea** | TUI mode (terminal UI) | MIT |
| **Vulkan-go** | Optional GPU renderer | MIT |

### Installation

```bash
# Go (required)
# See: https://go.dev/doc/install

# Fyne dependencies (Linux)
sudo apt-get install libgl1-mesa-dev xorg-dev

# Vulkan SDK (optional, for GPU renderer)
# See: https://vulkan.lunarg.com/sdk/home
```

### Build

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

---

## Hysteresis Modeling Libraries

### 1. PyPreisach (Python)

**URL:** https://github.com/peterdjackson/PyPreisach (archived)

**Description:** Python implementation of classical Preisach hysteresis model.

**Features:**
- First-order reversal curves (FORCs)
- Everett function computation
- Preisach plane visualization

**Example:**
```python
from pypreisach import PreisachModel

model = PreisachModel(alpha_mean=1.0, beta_mean=-1.0)
E = np.linspace(-2, 2, 100)
P = [model.update(e) for e in E]
```

**Comparison to this module:**
- **This module:** Tanh-based Everett approximation in Go
- **PyPreisach:** Explicit hysteron grid in Python
- **Speed:** This module faster for real-time visualization
- **Accuracy:** PyPreisach more flexible for FORC calibration

---

### 2. hystereloop (Python)

**URL:** https://github.com/BioMag/hystereloop

**Description:** Analyzes hysteresis loops from experimental data.

**Features:**
- Loop area calculation
- Coercivity and remanence extraction
- FORC diagram generation

**Use Case:** Data analysis rather than simulation. Complement this module with experimental data processing.

---

### 3. Custom Preisach Implementation (Python)

**Minimal example for prototyping:**

```python
import numpy as np

class SimplePreisach:
    def __init__(self, Ec=1.0, Ps=1.0, n_hysterons=200, sigma=0.2):
        self.Ec = Ec
        self.Ps = Ps
        self.hysterons = []

        for _ in range(n_hysterons):
            alpha = np.random.normal(Ec, sigma * Ec)
            beta = np.random.normal(-Ec, sigma * Ec)
            if alpha > beta:
                self.hysterons.append({
                    'alpha': alpha,
                    'beta': beta,
                    'state': -1,
                    'weight': 1.0 / n_hysterons
                })

    def update(self, E):
        for h in self.hysterons:
            if E >= h['alpha']:
                h['state'] = +1
            elif E <= h['beta']:
                h['state'] = -1

        P = self.Ps * sum(h['weight'] * h['state'] for h in self.hysterons)
        return P
```

**Use Case:** Quick prototyping, educational demonstrations.

---

## Phase-Field Simulation Tools

### 1. FerroX (GPU-Accelerated)

**URL:** https://github.com/AMReX-Microelectronics/FerroX
**Paper:** arXiv:2210.15668

**Description:** GPU-accelerated phase-field simulator for ferroelectric domain dynamics.

**Features:**
- Time-Dependent Ginzburg-Landau (TDGL) solver
- Multi-GPU support via AMReX
- Arbitrary electrode geometries
- Piezoelectric coupling

**Installation:**
```bash
git clone https://github.com/AMReX-Microelectronics/FerroX
cd FerroX
cmake -S . -B build
cmake --build build
```

**Integration with this module:**
- Export Landau coefficients from this module's materials
- Use FerroX for detailed domain structure
- Import domain statistics to calibrate Preisach distribution

**Use Case:** Large-scale domain simulations (micron-scale), NOT real-time visualization.

---

### 2. FERRET (MOOSE Framework)

**URL:** https://github.com/mangerij/ferret
**Paper:** Computer Physics Communications

**Description:** Phase-field modeling of ferroic materials using MOOSE.

**Features:**
- Landau-Devonshire free energy
- Coupled electro-mechanical problems
- Domain wall dynamics
- Strain effects

**Installation:**
```bash
# Requires MOOSE framework
# See https://mooseframework.inl.gov/getting_started/
git clone https://github.com/mangerij/ferret
cd ferret && make -j4
```

**Input Example:**
```
[Materials]
  [./ferroelectric]
    type = LandauFreeEnergy
    alpha1 = -1.72e8
    alpha11 = 7.3e8
    alpha111 = 2.6e9
  [../]
[]
```

**Integration:** Export material parameters from this module → FERRET input file.

---

### 3. PRISMS-PF

**URL:** https://github.com/prisms-center/phaseField

**Description:** General-purpose phase-field code, adaptable for ferroelectrics.

**Features:**
- High-performance C++
- Adaptive mesh refinement
- Parallelization

---

## Device/Circuit Simulation

### 1. ngspice + Verilog-A

**URL:** https://ngspice.sourceforge.io/

**Description:** Open-source SPICE with Verilog-A support via OpenVAF.

**FeFET Model Workflow:**

```bash
# 1. Get Verilog-A model (e.g., from Purdue Compact Models)
# 2. Compile with OpenVAF
openvaf fefet.va -o fefet.osdi

# 3. Use in ngspice
```

**SPICE Netlist:**
```spice
.model fefet_n fefet osdi="fefet.osdi"
+ Pr=25u Ps=30u Ec=1.2Meg tau=1n
```

**Integration with this module:**
1. Export material parameters (Pr, Ps, Ec, τ)
2. Generate Verilog-A parameter file
3. Simulate FeFET circuits in ngspice

**Compact Model Sources:**
- Purdue University: https://nanohub.org/resources/fefetcompact
- University of Oulu thesis (2025): Cadence-compatible FeCap model

---

### 2. OpenVAF (Verilog-A Compiler)

**URL:** https://github.com/openvaf/OpenVAF

**Description:** Open-source Verilog-A compiler → OSDI plugins for ngspice.

**Installation:**
```bash
cargo install openvaf
```

**Usage:**
```bash
openvaf ferroelectric_model.va -o ferroelectric_model.osdi
```

---

### 3. Xyce (Sandia National Labs)

**URL:** https://github.com/Xyce/Xyce

**Description:** High-performance parallel circuit simulator with Verilog-A support.

**Advantages over ngspice:**
- Better convergence for stiff problems
- Parallel simulation
- Noise and sensitivity analysis

---

## Architecture-Level Simulators

### 1. NeuroSim (Georgia Tech)

**URL:** https://github.com/neurosim

**Description:** Device-to-architecture benchmark for neuromorphic computing.

**Features:**
- FeFET device models
- Crossbar array simulation
- Energy/area estimation
- Neural network accuracy

**Integration:**
```cpp
// Simplified FeFET model in NeuroSim
double conductance = G_min + (G_max - G_min) * polarization_level / 29.0;
```

**Export from this module:**
1. Discrete levels (0-29) → conductance mapping
2. Gmin, Gmax from Module 2
3. Energy per operation from material parameters

---

### 2. CrossSim (Sandia)

**URL:** https://github.com/sandialabs/cross-sim

**Description:** Crossbar array simulator with device non-idealities.

**Features:**
- Device variation
- Noise models
- Programming non-idealities

**Integration:** Export conductance functions based on P-E model.

---

### 3. AIHWKIT (IBM)

**URL:** https://github.com/IBM/aihwkit

**Description:** PyTorch extension for analog AI hardware simulation.

**Features:**
- Device-level noise
- Drift modeling
- Training with hardware awareness

**Integration:** Define custom device from this module's hysteresis curves.

---

## Visualization Tools

### 1. This Project's GUI (Fyne)

**Location:** `module1-hysteresis/pkg/gui/`

**Features:**
- Real-time P-E loop visualization
- Interactive E-field control
- 30-level state display
- Write/Read mode indicator
- Multiple waveform modes

---

### 2. ParaView (3D Visualization)

**URL:** https://www.paraview.org/

**Use Case:** Visualize phase-field simulation results from FerroX/FERRET.

**Workflow:**
1. Export domain structure from FerroX (VTK format)
2. Open in ParaView
3. Visualize domain walls, polarization vectors

---

### 3. KLayout (Layout Viewer)

**URL:** https://www.klayout.de/

**Use Case:** View ferroelectric device layouts, integrate with Module 6 (EDA).

---

## Integration Guide

### Exporting Material Parameters

```go
// From module1-hysteresis/pkg/ferroelectric/material.go
material := ferroelectric.DefaultSuperlattice()

// JSON export
json, _ := json.MarshalIndent(material, "", "  ")
ioutil.WriteFile("material_params.json", json, 0644)
```

### Importing to Python (PyPreisach)

```python
import json
import numpy as np

# Load parameters
with open('material_params.json') as f:
    params = json.load(f)

# Create PyPreisach model
from pypreisach import PreisachModel
model = PreisachModel(
    alpha_mean=params['Ec'],
    beta_mean=-params['Ec'],
    sigma_alpha=params['Ec'] * 0.2,
    sigma_beta=params['Ec'] * 0.2
)
```

### Exporting to SPICE (via Verilog-A)

```go
// Generate Verilog-A parameters
fmt.Printf(".model fefet fefet\n")
fmt.Printf("+ Pr=%.2eu Ps=%.2eu\n", material.Pr*1e6, material.Ps*1e6)
fmt.Printf("+ Ec=%.2eMeg\n", material.Ec/1e6)
fmt.Printf("+ tau=%.2en\n", material.Tau*1e9)
```

### Importing FORC Data

```go
// Future enhancement: load experimental FORC
func LoadFORC(filename string) PreisachDistribution {
    // Parse FORC file
    // Extract μ(α,β) distribution
    // Return calibrated Preisach model
}
```

---

## Tool Comparison Matrix

| Tool | Level | Hysteresis Model | Speed | Open Source | Use Case |
|------|-------|------------------|-------|-------------|----------|
| **This Module (Go)** | Macro | Preisach | Real-time | ✅ Yes | Education, visualization |
| **PyPreisach** | Macro | Preisach | Fast | ✅ Yes | Prototyping, FORC analysis |
| **FerroX** | Micro | TDGL | Slow (GPU) | ✅ Yes | Domain structure |
| **FERRET** | Micro | Landau | Slow | ✅ Yes | Coupled physics |
| **ngspice + VA** | Device | Custom | Medium | ✅ Yes | Circuit simulation |
| **NeuroSim** | Array | Simplified | Fast | ✅ Yes | Architecture estimation |
| **CrossSim** | Array | Custom | Fast | ✅ Yes | Array-level accuracy |
| **AIHWKIT** | Array | Custom | Fast | ✅ Yes | ML training |
| **Comsol** | Micro | Landau/TDGL | Slow | ❌ Commercial | Detailed FEA |
| **Sentaurus** | Device | Physics-based | Slow | ❌ Commercial | TCAD |

---

## Learning Path

### For Beginners

1. **Run this module** (`fecim-lattice-tools hysteresis`)
2. **Read [physics.md](./physics.md)** for equations
3. **Implement SimplePreisach** (Python example above)
4. **Compare loop shapes** with different parameters

### For Intermediate Users

1. **Study `preisach_advanced.go`** implementation
2. **Add temperature slider** to GUI
3. **Export Preisach distribution** visualization
4. **Connect to ngspice** via SPICE export

### For Advanced Users

1. **Build Verilog-A FeFET model** using OpenVAF
2. **Couple with FerroX** for domain structure
3. **Integrate with NeuroSim** for array-level
4. **Contribute improvements** to open-source tools

---

## Key Resources

### Documentation
- **[Physics Reference](./physics.md)** - Equations and implementation
- **[Features](./features.md)** - What the module does
- **FerroX documentation:** GitHub wiki
- **MOOSE tutorials:** https://mooseframework.inl.gov/

### Academic Papers (Open Access)
- FerroX paper: arXiv:2210.15668
- NeuroSim papers: Various arXiv
- Preisach modeling: See `docs/4-research/papers/`

### Online Courses
- "Ferroelectric Materials" (Coursera/edX)
- MOOSE framework tutorials
- SPICE simulation basics

---

## Export Formats Supported

| Format | File | Use Case |
|--------|------|----------|
| **JSON** | `material.json` | General interchange |
| **CSV** | `loop_data.csv` | Spreadsheet analysis |
| **SPICE** | `netlist.sp` | Circuit simulation |
| **Verilog-A** | `params.va` | Compact modeling |

---

## Integration Examples

### With Module 2 (Crossbar)

```go
// Export conductance mapping
level := hysteresis.GetDiscreteLevel()
conductance := crossbar.LevelToConductance(level, Gmin, Gmax)
```

### With Module 6 (EDA)

```go
// Export cell parameters for layout
cellParams := eda.CellParameters{
    Ec:        material.Ec,
    Thickness: material.Thickness,
    Area:      material.Area,
}
```

### With NeuroSim

```cpp
// Import discrete levels
#include "fecim_export.h"
double weight = fecim_level / 29.0;  // Normalize
```

---

## Summary

### Best Tools by Use Case

| Use Case | Recommended Tool | Why |
|----------|------------------|-----|
| **Educational visualization** | This module | Real-time, intuitive |
| **Python prototyping** | PyPreisach | Simple, flexible |
| **SPICE circuit sim** | ngspice + OpenVAF | Industry standard |
| **Domain structure** | FerroX | GPU-accelerated |
| **Array estimation** | NeuroSim | Well-validated |
| **ML training** | AIHWKIT | PyTorch integration |

---

**See Also:**
- **[Features](./features.md)** - Module capabilities
- **[Physics](./physics.md)** - Model equations
- **Module 2 Tools** - Crossbar simulation ecosystem

---

**Last Updated:** 2026-02-16
