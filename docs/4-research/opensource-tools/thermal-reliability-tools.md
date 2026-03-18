# Thermal Simulation & Reliability Tools

**Open-Source Tools for Thermal Management, Endurance Testing, and Reliability Analysis**

*Last Updated: March 2026*

---

## Overview

FeCIM devices face critical thermal and reliability challenges: ferroelectric fatigue (10^9-10^12 cycles), wake-up effects, retention loss at elevated temperatures, and BEOL thermal budgets. This document catalogs open-source tools for thermal simulation, reliability modeling, and endurance characterization.

### Why This Matters for FeCIM

- **Automotive (AEC-Q100 Grade 0):** Operating range -40 to +150 C
- **Cryogenic CIM:** Potential 4K operation for quantum-adjacent computing
- **Fatigue/endurance:** HZO cycling endurance degrades above 10^9 cycles
- **Retention:** Polarization retention time depends exponentially on temperature
- **BEOL integration:** Maximum thermal budget constraints for HfO2 crystallization

---

## Thermal Simulation Tools

### 1. Elmer FEM

**Open-source multiphysics FEM solver with thermal analysis capabilities.**

- **Website:** https://www.elmerfem.org/
- **GitHub:** https://github.com/ElmerCSC/elmerfem
- **License:** GPL-2.0
- **Status:** Active (maintained by CSC Finland)

#### Key Features
- Heat equation solver (steady-state and transient)
- Coupled electro-thermal simulation
- Structured/unstructured mesh support
- MPI parallelization for HPC
- ElmerGUI for interactive setup
- Python/MATLAB interface for scripting

#### FeCIM Use Cases
- Crossbar array thermal profile during MVM operations
- Joule heating in word/bit lines under IR drop
- Thermal budget analysis for BEOL integration
- Temperature-dependent polarization modeling

#### Installation
```bash
sudo apt-get install elmer
# or from source:
git clone https://github.com/ElmerCSC/elmerfem.git
cd elmerfem && mkdir build && cd build
cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local
make -j$(nproc) && sudo make install
```

#### Example: Crossbar Thermal Profile
```
! Simple 2D thermal model for 32x32 crossbar
Header
  Mesh DB "." "crossbar_mesh"
End

Simulation
  Coordinate System = Cartesian
  Simulation Type = Steady State
  Output Intervals = 1
End

Body 1
  Equation = 1
  Material = 1
  Body Force = 1
End

Material 1
  Heat Conductivity = 148.0  ! Silicon W/(m*K)
  Density = 2330.0
  Heat Capacity = 700.0
End

Body Force 1
  Heat Source = Variable Coordinate 1, Coordinate 2
    Real Procedure "HeatSource" "GetSource"
End

Equation 1
  Active Solvers(1) = 1
End

Solver 1
  Equation = Heat Equation
  Variable = Temperature
  Procedure = "HeatSolve" "HeatSolver"
End

Boundary Condition 1
  Target Boundaries(1) = 1
  Temperature = 300.0  ! Substrate at room temperature
End
```

---

### 2. OpenFOAM (Thermal Transport)

**CFD/thermal transport solver for chip-level thermal management.**

- **Website:** https://www.openfoam.com/
- **GitHub:** https://github.com/OpenFOAM/OpenFOAM-dev
- **License:** GPL-3.0
- **Status:** Active (maintained by OpenFOAM Foundation + ESI Group)

#### Key Features
- Conjugate heat transfer (solid + fluid)
- Transient thermal analysis
- Multi-region thermal coupling
- Custom boundary conditions (heat flux, convection)
- Massively parallel (MPI)

#### FeCIM Use Cases
- Package-level thermal analysis with cooling
- 3D IC thermal management (stacked crossbar arrays)
- Hotspot identification during sustained MVM operations
- Thermal coupling between analog and digital blocks

#### Limitations
- Steep learning curve (C++ based)
- Overkill for simple crossbar thermal models
- Better suited for package/system level than device level

---

### 3. HotSpot (UVA)

**Chip-level thermal modeling tool specifically designed for IC analysis.**

- **Website:** https://lava.cs.virginia.edu/HotSpot/
- **GitHub:** https://github.com/uvahotspot/HotSpot
- **License:** BSD
- **Status:** Maintained

#### Key Features
- Compact thermal model (fast execution)
- Floorplan-aware thermal simulation
- Transient and steady-state analysis
- Integration with power simulators
- Grid-based and block-based models

#### FeCIM Use Cases
- Chip-level thermal map for FeCIM + digital logic
- Power density estimation for crossbar arrays
- Thermal-aware floorplanning for Module 6 EDA
- Temperature gradient analysis across die

#### Installation
```bash
git clone https://github.com/uvahotspot/HotSpot.git
cd HotSpot && make
```

---

### 4. CACTI (HP Labs)

**Cache/memory energy, area, and timing model — useful for CIM energy comparison.**

- **GitHub:** https://github.com/HewlettPackard/cacti
- **License:** BSD-3
- **Status:** Maintained

#### Key Features
- SRAM/DRAM/NVM energy models
- Area estimation for memory arrays
- Timing (access latency) modeling
- Technology scaling (7nm to 180nm)
- Leakage power estimation

#### FeCIM Use Cases
- Energy comparison: FeCIM crossbar vs conventional SRAM/DRAM
- Area estimation for peripheral circuits (ADC/DAC)
- Technology scaling projections for FeCIM roadmap
- Module 5 comparison metrics validation

#### Limitations
- No ferroelectric device model (must add custom)
- Optimized for standard memory, not CIM operations
- MVM energy not directly modeled

---

## Reliability & Endurance Tools

### 5. SPICE Monte Carlo (ngspice built-in)

**Statistical variation analysis using Monte Carlo sweeps in ngspice.**

Already documented in [circuit-simulation-tools.md](./circuit-simulation-tools.md), but specifically for reliability:

```spice
* Monte Carlo for FeCIM cell variation
.param Pr_nom=20u
.param Pr_sigma=2u
.param Pr={gauss(Pr_nom, Pr_sigma, 3)}

* Run 1000 Monte Carlo iterations
.control
  let N = 1000
  let i = 0
  while i < N
    alter @Pr = gauss(20u, 2u)
    tran 1n 100n
    let i = i + 1
  end
.endc
```

#### FeCIM Use Cases
- Process variation impact on crossbar accuracy
- Threshold voltage (Vt) distribution for FeFET arrays
- Remnant polarization (Pr) spread across devices
- Worst-case analysis for ADC precision requirements

---

### 6. Device Aging Models (BTI/HCI)

**No single open-source tool dominates this space. Key approaches:**

| Model | Type | Tool | Applicability |
|-------|------|------|--------------|
| NBTI | Transistor aging | ngspice + aging models | Peripheral transistor lifetime |
| HCI | Hot carrier injection | Xyce (built-in) | Sense amplifier degradation |
| TDDB | Dielectric breakdown | Custom SPICE models | FE capacitor lifetime |
| Fatigue | Cycling endurance | FerroX (partial) | HZO polarization loss |

### FeCIM-Specific Reliability Concerns

| Failure Mode | Mechanism | Simulation Approach |
|--------------|-----------|-------------------|
| **Wake-up** | Oxygen vacancy redistribution | FerroX phase-field |
| **Fatigue** | Domain pinning, crack formation | Preisach with cycling model |
| **Imprint** | Asymmetric internal field | L-K with time-dependent bias |
| **Retention** | Depolarization field | Thermal activation model |
| **Endurance** | Dielectric degradation | TDDB + cycling stress |

---

### 7. CoMeT — Integrated Core-Memory Thermal Simulator

**First thermal simulator that models core AND memory together — critical for CIM where compute happens inside memory.**

- **GitHub:** https://github.com/marg-tools/CoMeT
- **License:** MIT
- **Status:** Active

#### Key Features
- Integrated core AND memory interval thermal simulation
- Supports 2D, 2.5D, and 3D processor-memory systems
- Only ~5% simulation overhead vs core-only tools
- Thermal visualization (video output)
- DTM (Dynamic Thermal Management) policy support
- Built-in floorplan generator

#### FeCIM Use Cases
- **3D stacked FeCIM** inter-tier thermal coupling
- CIM-specific thermal modeling (compute inside memory — not possible with HotSpot alone)
- DTM policy evaluation for FeCIM arrays under sustained workloads
- Comparison of 2D vs 3D CIM thermal profiles

#### Why CoMeT Matters for CIM
Traditional thermal tools model processor and memory separately. In CIM, compute happens inside memory — there is no separation. CoMeT is the first tool to model this correctly.

```
Traditional (HotSpot):        CIM-aware (CoMeT):
  Core thermal ──┐              Core + Memory thermal
  Memory thermal ─┘ (separate)  (integrated, coupled)
```

---

### 8. 3D-ICE 4.0 (EPFL, November 2025)

**Chip-level thermal simulator with GDS parsing and anisotropic materials — ideal for 3D-stacked FeCIM.**

- **GitHub:** https://github.com/esl-epfl/3d-ice
- **Paper:** arXiv [2512.05823](https://arxiv.org/abs/2512.05823)
- **License:** Open source (EPFL)
- **Status:** Active (major v4.0 release November 2025)

#### Key Features
- Anisotropic material support (distinct x/y/z thermal conductivities)
- Direct GDS layout parsing for thermal analysis
- Customizable non-uniform thermal modeling
- Multiple heat sink models
- Parallel acceleration
- Support for 2.5D/3D heterogeneous chiplet systems
- Microchannel liquid cooling emulation

#### FeCIM Use Cases
- **HZO thin films** have anisotropic thermal properties — 3D-ICE can model this natively
- **3D-stacked FeCIM architectures** need inter-tier thermal analysis
- Direct GDS import from Module 6 EDA output for thermal verification
- Chiplet-based CIM designs with heterogeneous thermal profiles

#### Installation
```bash
git clone https://github.com/esl-epfl/3d-ice.git
cd 3d-ice && make
```

#### Comparison with HotSpot

| Feature | HotSpot | 3D-ICE 4.0 |
|---------|---------|-----------|
| Anisotropic materials | No | Yes |
| GDS import | No | Yes |
| 3D IC support | Basic | Native |
| Liquid cooling | No | Microchannel |
| Speed | Fast | Fast |
| Best for | 2D floorplans | 3D chiplets |

---

## Comparison Matrix

| Tool | Type | License | GPU | FeCIM Relevance | Best For |
|------|------|---------|-----|----------------|----------|
| **Elmer FEM** | Thermal FEM | GPL-2 | No | HIGH | Crossbar thermal profiles |
| **OpenFOAM** | CFD/Thermal | GPL-3 | Yes | MEDIUM | Package-level thermal |
| **HotSpot** | IC Thermal | BSD | No | HIGH | Chip floorplan thermal (2D) |
| **CoMeT** | Core+Memory | MIT | No | HIGH | CIM-specific thermal (core+memory coupled) |
| **3D-ICE 4.0** | IC Thermal | Open | Yes | HIGH | 3D chiplet thermal (GDS import) |
| **CACTI** | Memory model | BSD-3 | No | MEDIUM | Energy/area comparison |
| **ngspice MC** | Statistical | BSD-3 | No | HIGH | Device variation analysis |

---

## Integration with FeCIM Modules

### Module 2 (Crossbar) — Thermal Profile
```
Power map from MVM simulation
    ↓
HotSpot or Elmer FEM thermal solve
    ↓
Temperature-dependent conductance adjustment
    ↓
Re-run MVM with thermal corrections
```

### Module 4 (Circuits) — Reliability Budgeting
```
Circuit operating point (ngspice)
    ↓
Monte Carlo variation analysis
    ↓
Worst-case timing/accuracy extraction
    ↓
Design margin allocation
```

### Module 5 (Comparison) — Energy Benchmarking
```
CACTI baseline (SRAM energy at target node)
    ↓
FeCIM crossbar energy model
    ↓
Normalized comparison (pJ/MAC)
    ↓
Technology scaling projection
```

---

**Last Updated:** March 2026
**Category:** Thermal Simulation & Reliability
**Tools Documented:** 8
