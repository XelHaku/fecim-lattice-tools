# Open-Source Hysteresis Simulation Tools

**A Comprehensive Guide to Available Tools for Ferroelectric Hysteresis Modeling**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source tools, libraries, and frameworks for simulating ferroelectric hysteresis, Preisach models, domain dynamics, and related phenomena. It covers tools from academic research, industry, and the open-source community.

---

## 1. Hysteresis Modeling Libraries

### 1.1 Python Libraries

#### PyPreisach
**URL:** https://github.com/peterdjackson/PyPreisach (archived)

**Description:** Python implementation of the classical Preisach hysteresis model.

**Features:**
- First-order reversal curves (FORCs)
- Everett function computation
- Visualization of Preisach plane

**Installation:**
```bash
pip install pypreisach  # If available on PyPI
# Or clone and install
git clone https://github.com/peterdjackson/PyPreisach
cd PyPreisach && pip install -e .
```

**Example:**
```python
from pypreisach import PreisachModel

model = PreisachModel(alpha_mean=1.0, beta_mean=-1.0)
E = np.linspace(-2, 2, 100)
P = [model.update(e) for e in E]
```

**Limitations:** Basic implementation, no temperature dependence.

---

#### hystereloop
**URL:** https://github.com/BioMag/hystereloop

**Description:** Python library for analyzing hysteresis loops from experimental data.

**Features:**
- Loop area calculation
- Coercivity and remanence extraction
- FORC diagram generation

**Use Case:** Data analysis rather than simulation.

---

#### PyFerroelectric (Custom/Academic)
**Status:** Multiple implementations in academic papers

**Typical Implementation:**
```python
import numpy as np

class PreisachHysteresis:
    def __init__(self, Ec=1e8, Ps=0.25, sigma=0.2):
        self.Ec = Ec
        self.Ps = Ps
        self.sigma = sigma
        self.hysterons = self._init_hysterons(400)

    def _init_hysterons(self, n):
        """Create hysteron distribution."""
        hysterons = []
        for _ in range(n):
            alpha = np.random.normal(self.Ec, self.sigma * self.Ec)
            beta = np.random.normal(-self.Ec, self.sigma * self.Ec)
            if alpha > beta:
                hysterons.append({'alpha': alpha, 'beta': beta, 'state': -1})
        return hysterons

    def update(self, E):
        """Apply field and return polarization."""
        for h in self.hysterons:
            if E >= h['alpha']:
                h['state'] = +1
            elif E <= h['beta']:
                h['state'] = -1
        return self.Ps * sum(h['state'] for h in self.hysterons) / len(self.hysterons)
```

---

### 1.2 MATLAB/Octave

#### Jiles-Atherton Model (Magnetic, Adaptable)
**URL:** https://www.mathworks.com/matlabcentral/fileexchange/

**Description:** While designed for magnetic hysteresis, adaptable for ferroelectrics.

**Modification for Ferroelectrics:**
- Replace B→H with P→E
- Adjust parameters for ferroelectric materials

---

#### FORC Analysis Toolbox
**URL:** Academic distribution (contact researchers)

**Description:** First-Order Reversal Curve analysis in MATLAB.

**Use Case:** Extracting Preisach distributions from measured data.

---

### 1.3 C/C++ Libraries

#### This Project's Implementation (Go)
**Location:** `module1-hysteresis/pkg/ferroelectric/`

**Features:**
- Full Mayergoyz Preisach model (`preisach_advanced.go`)
- Basic tanh-based Preisach (`preisach.go`)
- HZO material parameters (`material.go`)
- Temperature dependence
- Wake-up and fatigue modeling
- 30-level discretization
- Real-time GUI visualization

**Example (Go):**
```go
import "module1-hysteresis/pkg/ferroelectric"

material := ferroelectric.DefaultHZO()
model := ferroelectric.NewMayergoyzPreisach(material, 30)

// Generate hysteresis loop
E, P := model.GetHysteresisLoop(2*material.Ec, 100)
```

---

## 2. Phase-Field Simulation Tools

### 2.1 FerroX (GPU-Accelerated)

**URL:** https://github.com/AMReX-Microelectronics/FerroX
**Paper:** arXiv:2210.15668

**Description:** GPU-accelerated, massively-parallel phase-field simulator for ferroelectric domain dynamics.

**Features:**
- Time-Dependent Ginzburg-Landau (TDGL) solver
- Multi-GPU support via AMReX
- Arbitrary electrode geometries
- Piezoelectric coupling

**Installation:**
```bash
git clone https://github.com/AMReX-Microelectronics/FerroX
cd FerroX
# Requires AMReX framework
cmake -S . -B build
cmake --build build
```

**Use Case:** Large-scale domain structure simulations, NOT real-time visualization.

---

### 2.2 FERRET (MOOSE Framework)

**URL:** https://github.com/mangerij/ferret
**Paper:** Computer Physics Communications

**Description:** Phase-field modeling of ferroic materials using the MOOSE framework.

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
cd ferret
make -j4
```

**Input Example (MOOSE format):**
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

---

### 2.3 PRISMS-PF (Phase-Field)

**URL:** https://github.com/prisms-center/phaseField

**Description:** General-purpose phase-field code, adaptable for ferroelectrics.

**Features:**
- High-performance C++
- Adaptive mesh refinement
- Parallelization

---

## 3. Device/Circuit Simulation Tools

### 3.1 ngspice + Verilog-A (FeFET Models)

**URL:** https://ngspice.sourceforge.io/

**Description:** Open-source SPICE simulator with Verilog-A support via OpenVAF.

**FeFET Model Workflow:**

1. **Get Verilog-A model** (from academic papers or Purdue Compact Models)
2. **Compile with OpenVAF:**
```bash
openvaf fefet.va -o fefet.osdi
```
3. **Use in ngspice:**
```spice
.model fefet_n fefet osdi="fefet.osdi"
+ Pr=25u Ps=30u Ec=1.2Meg tau=1n
```

**Compact Model Sources:**
- **Purdue University:** https://nanohub.org/resources/fefetcompact
- **University of Oulu thesis (2025):** Cadence-compatible FeCap model
- **Academic papers:** Multiple Verilog-A implementations

---

### 3.2 OpenVAF (Verilog-A Compiler)

**URL:** https://github.com/openvaf/OpenVAF

**Description:** Open-source Verilog-A compiler producing OSDI plugins for ngspice.

**Installation:**
```bash
cargo install openvaf
```

**Compilation:**
```bash
openvaf ferroelectric_model.va -o ferroelectric_model.osdi
```

---

### 3.3 Xyce (Sandia National Labs)

**URL:** https://github.com/Xyce/Xyce

**Description:** High-performance parallel circuit simulator with Verilog-A support.

**Advantages over ngspice:**
- Better convergence for stiff problems
- Parallel simulation
- Noise and sensitivity analysis

---

## 4. Architecture-Level Simulators

### 4.1 NeuroSim (Georgia Tech)

**URL:** https://github.com/neurosim

**Description:** Device-to-architecture benchmark framework for neuromorphic computing.

**Features:**
- FeFET device models
- Crossbar array simulation
- Energy/area estimation
- Neural network accuracy

**Hysteresis Relevance:** Uses simplified conductance models, not full P-E curves.

**Device Model:**
```cpp
// Simplified FeFET model in NeuroSim
double conductance = G_min + (G_max - G_min) * polarization_level / 29.0;
```

---

### 4.2 CrossSim (Sandia)

**URL:** https://github.com/sandialabs/cross-sim

**Description:** Crossbar array simulator with device non-idealities.

**Features:**
- Device variation
- Noise models
- Programming non-idealities

**Hysteresis Relevance:** Can integrate custom conductance functions based on P-E models.

---

### 4.3 AIHWKIT (IBM)

**URL:** https://github.com/IBM/aihwkit

**Description:** PyTorch extension for analog AI hardware simulation.

**Features:**
- Device-level noise
- Drift modeling
- Training with hardware awareness

**Hysteresis Relevance:** Supports custom device models.

---

## 5. Visualization and Analysis Tools

### 5.1 This Project's GUI (Fyne)

**Location:** `module1-hysteresis/pkg/gui/gui.go`

**Features:**
- Real-time P-E loop visualization
- Interactive E-field control
- 30-level state display
- Write/Read mode indicator
- Multiple waveform modes
- Material comparison

**Screenshot (ASCII representation):**
```
┌────────────────────────────────────────────────────────────┐
│  P (µC/cm²)                                                │
│   40 ┼    ╭──────╮                                         │
│  +Pr ┼────╯      │           ● Current point               │
│   20 ┼           │                                         │
│    0 ┼───────────┼─→ E (MV/cm)                            │
│  -20 ┼           │                                         │
│  -Pr ┼────╮      │                                         │
│  -40 ┼    ╰──────╯                                         │
│        -2  -Ec  0  +Ec  2                                  │
└────────────────────────────────────────────────────────────┘
```

---

### 5.2 ParaView (3D Visualization)

**URL:** https://www.paraview.org/

**Use Case:** Visualizing phase-field simulation results (domain structures).

---

### 5.3 KLayout (Layout Viewer)

**URL:** https://www.klayout.de/

**Use Case:** Viewing ferroelectric device layouts, integrating with EDA flow.

---

## 6. First-Principles Tools

### 6.1 VASP/Quantum ESPRESSO/ABINIT

**Description:** Density functional theory (DFT) codes for calculating ferroelectric properties from first principles.

**Use Case:**
- Compute Landau coefficients
- Predict polarization values
- Study phase stability

**Limitation:** Very slow, not for real-time simulation.

---

### 6.2 Phonopy (Phonon Calculations)

**URL:** https://github.com/phonopy/phonopy

**Use Case:** Soft mode analysis for ferroelectric phase transitions.

---

## 7. Tool Comparison Matrix

| Tool | Level | Hysteresis Model | Speed | Open Source |
|------|-------|------------------|-------|-------------|
| **This Project (Go)** | Macro | Preisach | 60 FPS | Yes |
| PyPreisach | Macro | Preisach | Fast | Yes |
| FerroX | Micro | TDGL | Slow (GPU) | Yes |
| FERRET | Micro | Landau | Slow | Yes |
| ngspice + Verilog-A | Device | Custom | Medium | Yes |
| NeuroSim | Array | Simplified | Fast | Yes |
| CrossSim | Array | Simplified | Fast | Yes |
| AIHWKIT | Array | Custom | Fast | Yes |
| Comsol | Micro | Landau/TDGL | Slow | **No** |
| Sentaurus | Device | Physics-based | Slow | **No** |

---

## 8. Integration with FeCIM Project

### 8.1 Current Stack

```
┌─────────────────────────────────────────────────────────────┐
│                    FeCIM Visualizer                          │
├─────────────────────────────────────────────────────────────┤
│  Module 1: Hysteresis                                        │
│  ├── preisach_advanced.go (Mayergoyz Preisach)              │
│  ├── preisach.go (tanh-based)                                │
│  ├── material.go (HZO parameters)                            │
│  └── gui.go (Fyne visualization)                             │
├─────────────────────────────────────────────────────────────┤
│  Module 2: Crossbar (uses conductance from hysteresis)      │
├─────────────────────────────────────────────────────────────┤
│  Module 3: MNIST (uses quantized states)                     │
└─────────────────────────────────────────────────────────────┘
```

### 8.2 Potential Integrations

| External Tool | Integration Point | Benefit |
|---------------|-------------------|---------|
| ngspice + OpenVAF | Export cell netlist | SPICE simulation |
| NeuroSim | Export array config | Energy estimation |
| FerroX | Import domain data | Detailed physics |
| CrossSim | Export weights | Array accuracy |

### 8.3 Export Formats Supported

| Format | File | Use |
|--------|------|-----|
| JSON | Cell parameters | General interchange |
| CSV | Weight matrices | Spreadsheet analysis |
| SPICE | Netlist | Circuit simulation |

---

## 9. Creating Your Own Hysteresis Model

### 9.1 Minimal Preisach Implementation (Python)

```python
import numpy as np
import matplotlib.pyplot as plt

class SimplePreisach:
    """Minimal Preisach model for educational purposes."""

    def __init__(self, Ec=1.0, Ps=1.0, n_hysterons=200, sigma=0.2):
        self.Ec = Ec
        self.Ps = Ps

        # Initialize hysterons with Gaussian distribution
        self.hysterons = []
        for _ in range(n_hysterons):
            alpha = np.random.normal(Ec, sigma * Ec)
            beta = np.random.normal(-Ec, sigma * Ec)
            if alpha > beta:  # Valid hysteron condition
                self.hysterons.append({
                    'alpha': alpha,
                    'beta': beta,
                    'state': -1,
                    'weight': 1.0 / n_hysterons
                })

    def update(self, E):
        """Apply electric field and return polarization."""
        for h in self.hysterons:
            if E >= h['alpha']:
                h['state'] = +1
            elif E <= h['beta']:
                h['state'] = -1
            # Otherwise: state unchanged (memory!)

        # Sum weighted contributions
        P = self.Ps * sum(h['weight'] * h['state'] for h in self.hysterons)
        return P

    def hysteresis_loop(self, Emax, n_points=100):
        """Generate full hysteresis loop."""
        E_values = []
        P_values = []

        # Reset
        for h in self.hysterons:
            h['state'] = -1

        # 0 → +Emax
        for E in np.linspace(0, Emax, n_points):
            E_values.append(E)
            P_values.append(self.update(E))

        # +Emax → -Emax
        for E in np.linspace(Emax, -Emax, n_points * 2):
            E_values.append(E)
            P_values.append(self.update(E))

        # -Emax → +Emax
        for E in np.linspace(-Emax, Emax, n_points * 2):
            E_values.append(E)
            P_values.append(self.update(E))

        return np.array(E_values), np.array(P_values)


# Demo
if __name__ == "__main__":
    model = SimplePreisach(Ec=1.0, Ps=1.0, n_hysterons=500)
    E, P = model.hysteresis_loop(Emax=2.0)

    plt.figure(figsize=(8, 6))
    plt.plot(E, P, 'b-', linewidth=0.5)
    plt.xlabel('Electric Field E (normalized)')
    plt.ylabel('Polarization P (normalized)')
    plt.title('Preisach Hysteresis Loop')
    plt.axhline(0, color='gray', linestyle='--', linewidth=0.5)
    plt.axvline(0, color='gray', linestyle='--', linewidth=0.5)
    plt.grid(True, alpha=0.3)
    plt.show()
```

### 9.2 Adding Temperature Dependence

```python
def temperature_corrected_Ec(self, Ec0, T, Tc=723):
    """Ec(T) = Ec0 * (1 - T/Tc)^0.5"""
    if T >= Tc:
        return 0
    return Ec0 * np.sqrt(1 - T / Tc)
```

### 9.3 Adding KAI Switching Dynamics

```python
def switching_progress(self, t, tau=1e-9, n=2.0):
    """KAI model: P(t) = 1 - exp(-(t/tau)^n)"""
    return 1 - np.exp(-(t / tau) ** n)
```

---

## 10. Recommended Learning Path

### For Beginners

1. **Run this project's demo** (`module1-hysteresis`)
2. **Read [hysteresis.physics.md](hysteresis.physics.md)** for equations
3. **Implement SimplePreisach** above in Python
4. **Compare loop shapes** with different parameters

### For Intermediate Users

1. **Study preisach_advanced.go** implementation
2. **Add temperature slider** to GUI
3. **Export Preisach distribution** visualization
4. **Connect to ngspice** via SPICE export

### For Advanced Users

1. **Build Verilog-A FeFET model** using OpenVAF
2. **Couple with FerroX** for domain structure
3. **Integrate with NeuroSim** for array-level
4. **Contribute improvements** to open-source tools

---

## 11. Key Resources

### Documentation

- **This project's PHYSICS.md:** [hysteresis.physics.md](hysteresis.physics.md)
- **This project's ELI5:** [hysteresis.ELI5.md](hysteresis.ELI5.md)
- **FerroX documentation:** GitHub wiki
- **MOOSE tutorials:** https://mooseframework.inl.gov/

### Academic Papers (Open Access)

- FerroX paper: arXiv:2210.15668
- NeuroSim papers: Various arXiv
- Preisach modeling: See hysteresis.research.md

### Online Courses

- "Ferroelectric Materials" (Coursera/edX)
- MOOSE framework tutorials
- SPICE simulation basics

---

## 12. Summary

### Best Tools by Use Case

| Use Case | Recommended Tool | Reason |
|----------|------------------|--------|
| **Educational visualization** | This project | Real-time, intuitive |
| **Python prototyping** | Custom Preisach | Simple, flexible |
| **SPICE circuit simulation** | ngspice + OpenVAF | Industry standard |
| **Domain structure** | FerroX | GPU-accelerated |
| **Array-level estimation** | NeuroSim | Well-validated |
| **Training with hardware awareness** | AIHWKIT | PyTorch integration |

### The Open-Source Hysteresis Stack

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │ FeCIM Viz │  │ NeuroSim  │  │ AIHWKIT   │               │
│  │ (This)    │  │ (Array)   │  │ (ML)      │               │
│  └───────────┘  └───────────┘  └───────────┘               │
├─────────────────────────────────────────────────────────────┤
│                     Model Layer                              │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │ Preisach  │  │ Landau    │  │ TDGL      │               │
│  │ (Macro)   │  │ (Thermo)  │  │ (Phase)   │               │
│  └───────────┘  └───────────┘  └───────────┘               │
├─────────────────────────────────────────────────────────────┤
│                     Simulation Layer                         │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │ ngspice   │  │ FerroX    │  │ FERRET    │               │
│  │ (SPICE)   │  │ (GPU PF)  │  │ (MOOSE)   │               │
│  └───────────┘  └───────────┘  └───────────┘               │
├─────────────────────────────────────────────────────────────┤
│                     Foundation Layer                         │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │ OpenVAF   │  │ AMReX     │  │ PETSc     │               │
│  │ (Verilog) │  │ (MPI)     │  │ (Linear)  │               │
│  └───────────┘  └───────────┘  └───────────┘               │
└─────────────────────────────────────────────────────────────┘
```

---

*This document is part of the FeCIM Visualizer project. For the research meta-study, see [hysteresis.research.md](hysteresis.research.md). For beginner explanations, see [hysteresis.ELI5.md](hysteresis.ELI5.md). For deep physics, see [hysteresis.physics.md](hysteresis.physics.md). For demo docs, see [hysteresis.demo.md](hysteresis.demo.md).*
