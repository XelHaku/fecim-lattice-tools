# Ferroelectric Simulation Tools

**Comprehensive Open-Source Tools for Material Modeling, Polarization Dynamics, and Device Simulation**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source tools for ferroelectric material simulation, polarization switching, domain dynamics, and device modeling. These tools directly support HfO₂-ZrO₂ superlattice research and FeCIM development at various computational scales.

### Key Distinctions by Scale

```
┌─────────────────────────────────────────────────────────────┐
│              Computational Hierarchy                         │
├─────────────────────────────────────────────────────────────┤
│  MACRO:       Hysteresis loops, Preisach models (μs)        │
│  MESO:        Phase-field, domain structure (ns-ms)         │
│  DEVICE:      FeFET/FTJ circuits, SPICE (ps-ns)            │
│  ARRAY:       Crossbar, neural networks (ns-μs)            │
│  SYSTEM:      Full chips, architecture sims (ms+)          │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. FERRET (MOOSE Framework)

**Repository:** https://github.com/mangerij/ferret
**License:** LGPL (MOOSE-based)
**Language:** C++ (Fortran optional)
**GPU Support:** ❌ CPU only
**Development Status:** Active

### Description

FERRET is a phase-field simulation framework built on MOOSE (Multiphysics Object-Oriented Simulation Environment). It solves coupled electro-mechanical problems in ferroelectric materials using Landau-Devonshire thermodynamics.

### Key Features

- **Multi-physics coupling:** Electrostatics, mechanics, ferroelectricity
- **Free energy models:** Landau-Devonshire phenomenological approach
- **Domain wall dynamics:** Direct visualization of domain evolution
- **P-E hysteresis:** Full loop generation with domain structure
- **Strain effects:** Electro-mechanical coupling for thin films
- **Material libraries:** BaTiO₃, PbTiO₃ templates (HfO₂ via custom parameters)

### Relevant to HfO₂-ZrO₂

FERRET can model HfO₂ with custom Landau coefficients. The key parameters needed:

```
α₁ = -1.72 × 10⁸  (J/m³/C²)    # Temperature-dependent
α₁₁ = 7.3 × 10⁸   (J/m³/C⁴)
α₁₁₁ = 2.6 × 10⁹  (J/m³/C⁶)
q₁₁ = 0.5          (m⁵/C²)      # Electrostrictive coupling
```

These coefficients can be extracted from first-principles calculations or fitted to experimental P-E curves.

### Installation

```bash
# 1. Install MOOSE dependencies
cd /tmp
git clone https://github.com/idaholab/moose
cd moose
./scripts/configure_environment.sh

# 2. Clone FERRET
git clone https://github.com/mangerij/ferret
cd ferret

# 3. Build
make -j 4

# 4. Verify
./ferret-opt --version
```

### Example: HfO₂ Thin Film Simulation

**Input file:** `hfo2_thinfilm.i`

```
[Materials]
  [./ferroelectric]
    type = LandauFreeEnergy
    alpha1 = -1.72e8    # Pr = 20 µC/cm²
    alpha11 = 7.3e8
    alpha111 = 2.6e9
    alpha111_sq = 0
    C_v = 3.0e5         # Heat capacity
    rho = 9700          # Density kg/m³
  [../]

  [./strain]
    type = LinearElasticMaterial
    block = 'ferroelectric'
    disp_x = disp_x
    disp_y = disp_y
    disp_z = disp_z
    elasticity_tensor = elasticity_tensor
  [../]
[]

[Kernels]
  [./ferroelectric_eq]
    type = FerroelectricBase
    variable = polar_x
    component = 0
  [../]
[]
```

**Run simulation:**

```bash
mpirun -n 4 ferret-opt -i hfo2_thinfilm.i
```

### Outputs

- VTK files (compatible with ParaView)
- CSV data for post-processing
- Domain structure animations
- P-E curve plots

### Comparison with Other Tools

| vs FerroX | FERRET | FerroX |
|-----------|---------|---------|
| Speed | Slower (CPU) | Faster (GPU 15×) |
| Model | Landau (phenomenological) | TDGL (field-based) |
| Accuracy | ✅ Well-validated | ✅ Physics-based |
| Scaling | OK (hundreds of cells) | Excellent (millions of cells) |
| Learning curve | Steep (MOOSE framework) | Moderate |

### Resources

- **MOOSE Documentation:** https://mooseframework.inl.gov/
- **FERRET Github Pages:** https://mangerij.github.io/ferret/
- **Paper:** Computer Physics Communications (search FERRET ferroelectric)

---

## 2. FerroX (GPU-Accelerated)

**Repository:** https://github.com/AMReX-Microelectronics/FerroX
**License:** BSD-3-Clause
**Language:** C++ (AMReX framework)
**GPU Support:** ✅ NVIDIA CUDA (15× speedup)
**Development Status:** Active

### Description

FerroX is a massively-parallel, GPU-accelerated phase-field simulator for ferroelectric devices. It solves the Time-Dependent Ginzburg-Landau (TDGL) equation using AMReX (Adaptive Mesh Refinement EXtensible Library).

### Key Features

- **TDGL equation:** Modern field-based polarization dynamics
- **GPU acceleration:** 15× speedup vs CPU
- **Arbitrary geometries:** Complex electrode configurations
- **Adaptive mesh refinement:** Focus computation on domain walls
- **Piezoelectric coupling:** Stress-polarization feedback
- **Device-compatible:** MFM/MFIS capacitors, NC-FETs

### Directly Applicable to FeCIM

FerroX has been used to model:

- **3D HfO₂ devices:** 10×10×2 nm³ structures
- **Negative capacitance FETs:** Domain wall dynamics
- **Ferroelectric tunnel junctions (FTJ):** Tunneling with polarization
- **Domain nucleation:** Understanding wake-up and fatigue

### Installation

**Prerequisites:**
```bash
# CUDA 11.2+ or 12.x
# C++17 compiler
# CMake 3.17+
```

**Build:**
```bash
git clone https://github.com/AMReX-Microelectronics/FerroX
cd FerroX
mkdir build && cd build

# With GPU support
cmake -DAMReX_GPU_BACKEND=CUDA -DAMReX_CUDA_ARCH=8.0 ..
make -j 4

# Verify
./Exec/FerroX --help
```

### Example: 3D HfO₂ Domain Structure

**Input file:** `inputs_hfo2_3d`

```
# Spatial resolution
amr.n_cell = 128 128 32        # 128×128×32 cells
amr.max_level = 1

# Material parameters (HfO₂-like)
material.eps_r = 25            # Relative permittivity
material.Pr = 20e-6            # Remanent polarization (C/m²)
material.Ec = 1.0e8            # Coercive field (V/m)
material.tau = 1e-9            # Switching time (s)

# Boundary conditions
boundary.type = 1              # Voltage boundary
boundary.voltage = 1.0         # Applied voltage (V)

# Time stepping
time.dt = 1e-12                # 1 ps time step
time.nsteps = 10000            # 10 ns total

# Output
plot_int = 100
plot_file = hfo2_3d
```

**Run:**
```bash
mpirun -n 4 ./Exec/FerroX inputs_hfo2_3d
```

### Outputs

- **HDF5 format:** Polarization field snapshots
- **ParaView compatible:** 3D domain visualization
- **Statistics file:** Total polarization, energy evolution
- **Domain wall velocity:** Direct measurement

### Performance Metrics

| System Size | CPU (1 core) | GPU (1 V100) | Speedup |
|------------|--------------|--------------|---------|
| 64³ cells | 45 s/step | 3 s/step | 15× |
| 256³ cells | 720 s/step | 48 s/step | 15× |
| 512³ cells | >1h/step | 300 s/step | 12× |

### Why FerroX for FeCIM?

1. **HfO₂ focus:** Designed for hafnium oxide ferroelectrics
2. **Device-realistic:** 3D structures matching actual layouts
3. **Physical accuracy:** TDGL based on first-principles
4. **Open-source:** Full access to source code
5. **Published validation:** Peer-reviewed benchmarks

### Resources

- **Paper:** "FerroX: A GPU-accelerated, 3D Phase-Field Simulation Framework..." (Computer Physics Communications 2023)
- **arXiv:** https://arxiv.org/abs/2210.15668
- **AMReX Documentation:** https://amrex-codes.github.io/

---

## 3. feram

**Repository:** https://loto.sourceforge.net/feram/
**License:** GPL-2.0
**Language:** Fortran 95/2003
**GPU Support:** ❌ CPU only
**Development Status:** Maintenance

### Description

feram is a fast Molecular Dynamics (MD) simulator for bulk and thin-film ferroelectrics. It uses a first-principles-based effective Hamiltonian approach for rapid sampling of phase space.

### Key Features

- **Effective Hamiltonian:** First-principles-derived, fast MD
- **Bulk ferroelectrics:** BaTiO₃, PbTiO₃, PbZrO₃
- **Thin films:** Simulates interface effects and clamping strain
- **Phase diagrams:** Temperature-composition-strain phase maps
- **Domain structure:** Low computational cost

### Limitations for HfO₂

- **No HfO₂ model:** Designed for ABO₃ perovskites
- **Not orthorhombic:** HfO₂ typically orthorhombic, not tetragonal
- **Workaround:** Could parameterize custom Hamiltonian (advanced)

### Installation

```bash
# Requires gfortran, BLAS/LAPACK
sudo apt-get install gfortran libblas-dev liblapack-dev

# Download from SourceForge
tar -xzf feram-v1.1.tar.gz
cd feram-v1.1

# Build
./configure --prefix=$HOME/feram
make
make install
```

### Relevance to FeCIM

**Limited direct applicability** but useful for:

- Comparison with perovskite-based ferroelectrics
- Understanding effective Hamiltonian methodology
- Rapid phase diagram generation for material screening

**Recommendation:** Use FerroX or FERRET for HfO₂-specific modeling.

---

## 4. pymatgen (Ferroelectricity Module)

**Repository:** https://github.com/materialsproject/pymatgen
**License:** MIT
**Language:** Python
**GPU Support:** ❌ CPU only (DFT runs offloaded)
**Development Status:** Active

### Description

pymatgen is a powerful Python library for materials analysis. Its ferroelectricity module computes polarization from first-principles calculations using Berry phase theory.

### Key Features

- **Berry phase polarization:** From VASP/ABINIT output
- **EnergyTrend:** Spontaneous polarization from E-field scans
- **Born charges:** Ionic and electronic contributions
- **Materials Project integration:** Access 100k+ calculated materials
- **Workflow automation:** Batch DFT calculations

### HfO₂ Applications

1. **Verify Landau coefficients:** Calculate from DFT structures
2. **Screen dopants:** La-doped, Y-doped HfO₂
3. **Interfacial effects:** HfO₂/SiO₂ heterostructures
4. **Berry phase calculations:** First-principles polarization

### Installation

```bash
pip install pymatgen
# Optional: DFT requirements
pip install pymatgen[dev]
```

### Example: Berry Phase Polarization

```python
from pymatgen.io.vasp import Vasprun
from pymatgen.analysis.ferroelectricity import Polarization

# Load VASP output
vasprun = Vasprun('vasprun.xml')

# Calculate polarization
pol = Polarization(vasprun.ionic_steps)

# Get spontaneous polarization
P_ferroelectric = pol.get_polarization_from_vasp_run(vasprun)
print(f"Polarization: {P_ferroelectric} (C/m²)")

# Phase analysis
pol_diff = pol.get_polarization()  # In electronic units
print(f"Polarization change: {pol_diff}")
```

### Workflow: Screening HfO₂ Dopants

```python
from pymatgen.core import Structure
from pymatgen.analysis.ferroelectricity import Polarization

# Generate HfO₂ supercell
hfo2 = Structure.from_spacegroup('P212121', [6.07, 5.87, 5.00], ['Hf', 'O', 'O'])
supercell = hfo2.make_supercell([2, 2, 2])

# Dope with lanthanides (La, Sm, Gd)
for dopant in ['La', 'Sm', 'Gd']:
    doped = supercell.copy()
    doped[0] = dopant  # Replace first Hf atom

    # Submit to DFT (via atomate/custodian)
    # ... [run VASP]

    # Analyze polarization (post-DFT)
    vasprun = Vasprun(f'{dopant}/vasprun.xml')
    pol = Polarization(vasprun.ionic_steps)
    print(f"{dopant}-doped: Pr = {pol.get_polarization()}")
```

### Why Use pymatgen?

1. **Open access to 100k+ materials data**
2. **Berry phase is most accurate** for polar materials
3. **Interfacial effects:** HfO₂/SiO₂ models
4. **Extensible:** Custom analysis plugins
5. **Published:** Peer-reviewed methodology

### Resources

- **Documentation:** https://pymatgen.org/
- **Ferroelectricity module:** https://pymatgen.org/pymatgen.analysis.ferroelectricity.html
- **Materials Project:** https://materialsproject.org/

---

## 5. Q-POP-Thermo

**Repository:** Mendeley Data (not standard GitHub)
**License:** MIT
**Language:** Python
**GPU Support:** ❌ CPU only
**Development Status:** Research release

### Description

Q-POP-Thermo solves Landau-Ginzburg-Devonshire (LGD) thermodynamics for ferroelectric materials. It generates phase diagrams as functions of temperature, electric field, and strain.

### Key Features

- **LGD free energy minimization:** Equilibrium polarization states
- **Phase diagrams:** T-E, T-σ maps
- **Monodomain approximation:** Single-crystal or thin-film limit
- **Strain effects:** Epitaxial clamping, pressure
- **Temperature dependence:** Full Landau-Devonshire formalism

### Relevant to HfO₂

Can compute phase diagrams for HfO₂ if Landau coefficients are provided:

- **α(T) = a(T - T₀):** Temperature dependence
- **β, γ:** Higher-order coefficients
- **Material-specific strain:** From literature or DFT

### Installation

```bash
# Download from Mendeley
# https://data.mendeley.com/datasets/wd6228g9ww/1

pip install numpy scipy matplotlib
```

### Example: HfO₂ Phase Diagram

```python
import numpy as np
from qpopthermo import LGD_Ferroelectric

# HfO₂ parameters (from literature)
material = {
    'a': 1.72e8,      # α₁ coefficient (J/m³/C²)
    'b': 7.3e7,       # α₁₁ coefficient
    'c': 2.6e8,       # α₁₁₁ coefficient
    'T0': 200,        # Curie temperature (K) - typical for HfO₂
    'q11': 0.5,       # Electrostrictive coupling
}

# Generate phase diagram
T_range = np.linspace(100, 400, 100)  # Temperature (K)
E_range = np.linspace(0, 2e8, 100)    # E-field (V/m)

phase_diagram = np.zeros((len(T_range), len(E_range)))

for i, T in enumerate(T_range):
    for j, E in enumerate(E_range):
        lgd = LGD_Ferroelectric(material, T=T, E_field=E)
        phase_diagram[i, j] = lgd.equilibrium_polarization()

# Plot
import matplotlib.pyplot as plt
plt.contourf(E_range/1e8, T_range, phase_diagram)
plt.xlabel('E-field (MV/cm)')
plt.ylabel('Temperature (K)')
plt.colorbar(label='Polarization (C/m²)')
plt.title('HfO₂ Phase Diagram')
plt.show()
```

### Outputs

- Phase diagrams (T-E, T-σ, T-x for alloys)
- Spontaneous polarization vs. temperature
- Domain reorientation boundaries
- Dielectric constant tensor

### Limitations

- **Monodomain only:** No domain wall dynamics
- **Equilibrium:** No time-dependent behavior
- **Manual coefficient input:** Requires literature data
- **No GUI:** Scripting only

### Resources

- **Publication:** Computer Physics Communications 2022
- **Dataset:** https://data.mendeley.com/datasets/wd6228g9ww/1

---

## 6. Ferro (Python Package)

**Repository:** https://github.com/JAnderson419/Ferro
**License:** CC BY-NC-SA (Non-commercial)
**Language:** Python
**GPU Support:** ❌ CPU only
**Development Status:** Active

### Description

Ferro is a Python package for ferroelectric data analysis, hysteresis loop processing, and Preisach model implementation. It's designed for experimentalists measuring P-E curves.

### Key Features

- **HysteresisData class:** Load/process P-E loops from instruments
- **PUND measurement support:** Polarization-Up-Nucleation-Down
- **Preisach model:** Forward and inverse implementations
- **Loop analysis:** Area, coercivity, remanence extraction
- **Data visualization:** Publication-ready plots

### HfO₂ Applications

1. **Parse measurement files:** From commercial PE loop tracers
2. **Extract coercivity/remanence:** From raw data
3. **Preisach fitting:** Match experimental loops
4. **Statistical analysis:** Batch processing of multiple samples

### Installation

```bash
git clone https://github.com/JAnderson419/Ferro
cd Ferro
pip install -e .
```

### Example: Processing Experimental Data

```python
from ferro import HysteresisData, PreisachModel

# Load P-E loop from CSV
loop = HysteresisData('hfo2_sample_001.csv')

# Extract parameters
print(f"Remanent polarization (Pr): {loop.remanence():.2f} µC/cm²")
print(f"Coercive field (Ec): {loop.coercivity():.2f} MV/cm")
print(f"Loop area (dissipated energy): {loop.loop_area():.2e} J/m³")

# Fit Preisach model
preisach = PreisachModel()
params = preisach.fit(loop)

# Generate synthetic loop
E_synth = np.linspace(-2, 2, 1000)
P_synth = preisach.predict(E_synth)

# Plot comparison
plt.plot(loop.E, loop.P, 'o-', label='Experimental')
plt.plot(E_synth, P_synth, '-', label='Preisach fit')
plt.xlabel('E-field (MV/cm)')
plt.ylabel('Polarization (µC/cm²)')
plt.legend()
plt.show()
```

### Batch Processing

```python
import glob
from ferro import HysteresisData

# Analyze multiple samples
files = glob.glob('data/hfo2_*.csv')

results = []
for f in files:
    loop = HysteresisData(f)
    results.append({
        'sample': f,
        'Pr': loop.remanence(),
        'Ec': loop.coercivity(),
        'Loss': loop.loop_area()
    })

# Export summary
import pandas as pd
df = pd.DataFrame(results)
df.to_csv('hfo2_analysis_summary.csv')
print(df.describe())
```

### Limitations

- **Non-commercial license:** Restricts use in proprietary tools
- **Analysis-focused:** Not primarily for simulation
- **No domain modeling:** Preisach macroscopic level only

### Resources

- **GitHub:** https://github.com/JAnderson419/Ferro
- **Documentation:** In-repo README and examples

---

## 7. PFECAP (Verilog-A Model)

**Repository:** https://github.com/supadupaplex/pfecap
**License:** GPL-2.0
**Language:** Verilog-A
**Simulator Support:** SPICE (ngspice, Xyce, Cadence), LTspice
**Development Status:** Maintenance

### Description

PFECAP is a circuit-level Preisach ferroelectric capacitor model in Verilog-A. It enables SPICE simulation of ferroelectric circuits with realistic hysteresis and material properties.

### Key Features

- **Preisach implementation:** Circuit-compatible
- **HfZrO₂ parameters:** Pre-configured for ferroelectric capacitors
- **Memory effects:** Hysteresis and retention
- **Compact model:** Fast simulation (O(n) hysterons)
- **Temperature dependence:** Ec(T), Pr(T) scaling

### Direct Application to FeCIM

Perfect for:

1. **1T1C memory cells:** Ferroelectric capacitor simulation
2. **1T1R arrays:** 1 transistor + 1 resistor configs (ReRAM-like)
3. **Peripheral circuits:** Sense amplifiers, write drivers
4. **System-level simulation:** Full chip SPICE netlists

### Installation

**Step 1: Get Verilog-A file**
```bash
git clone https://github.com/supadupaplex/pfecap
cd pfecap
```

**Step 2: Compile for ngspice**
```bash
# Install OpenVAF (Verilog-A compiler)
cargo install openvaf

# Compile to OSDI plugin
openvaf pfecap.va -o pfecap.osdi
```

**Step 3: Use in SPICE**

### Example: 1T1C Ferroelectric Memory Cell

```spice
* 1T1C cell simulation with PFECAP

.include pfecap.osdi

* Transistor
M1 node_d node_g 0 0 nmos W=100n L=10n

* Ferroelectric capacitor (PFECAP model)
X_fc node_d node_plate fc_capacitor

* PFECAP device definition
.subckt fc_capacitor p1 p2
  * Device parameters for HfO₂ (30 µC/cm²)
  B1 p1 p2 pfecap W=1u L=1u Pr=30e-6 Ec=1.0e8
.ends

* Transient simulation: Write and read
.tran 0 100u 0 1u

* Pulse voltage to gate (write)
Vgate node_g 0 PULSE(0 2 1u 1n 1n 10u 20u)

* Plate voltage (ground during write)
Vplate node_plate 0 DC 0

.print v(node_d) v(node_plate) i(Vgate)
.end
```

**Run:**
```bash
ngspice -b 1t1c_cell.cir -o 1t1c_cell.log
```

### PFECAP Parameters

| Parameter | Default | Unit | Description |
|-----------|---------|------|-------------|
| `Pr` | 20e-6 | C/m² | Remanent polarization |
| `Ps` | 25e-6 | C/m² | Saturation polarization |
| `Ec` | 1.0e8 | V/m | Coercive field |
| `tau` | 1e-9 | s | Switching time constant |
| `n_hysterons` | 200 | - | Number of hysterons |
| `sigma` | 0.2 | - | Hysteron distribution width |

### HfO₂ Preset Configuration

```spice
* HfO₂ ferroelectric capacitor (30 levels, conference claim; COSM 2025)
.model hfo2_fc pfecap (
+   Pr=20e-6         W=10e-9 L=10e-9
+   Ps=25e-6
+   Ec=1.0e8
+   tau=1e-9
+   n_hysterons=400
+   alpha=1.0
+)
```

### Why PFECAP for FeCIM?

1. **Ferroelectric-specific:** Not generic MIM capacitor
2. **Circuit-compatible:** Integrates with existing SPICE
3. **Memory effects:** Proper hysteresis modeling
4. **Fast:** 100× faster than phase-field models
5. **Material-tunable:** Easy parameter changes

### Workflow: From Device to Circuit

```
┌─────────────────────────────────────────┐
│ 1. Extract HfO₂ params from literature  │
├─────────────────────────────────────────┤
│ 2. Tune PFECAP Pr, Ec, tau             │
├─────────────────────────────────────────┤
│ 3. Validate against P-E measurements    │
├─────────────────────────────────────────┤
│ 4. Design 1T1C cell netlist            │
├─────────────────────────────────────────┤
│ 5. Simulate transient/AC response      │
├─────────────────────────────────────────┤
│ 6. Export results for NeuroSim         │
└─────────────────────────────────────────┘
```

### Resources

- **GitHub:** https://github.com/supadupaplex/pfecap
- **ngspice manual:** http://ngspice.sourceforge.net/
- **OpenVAF:** https://github.com/openvaf/OpenVAF
- **SPICE primer:** "The SPICE Book" (Nagel & Pederson)

---

## 8. negativec

**Repository:** https://github.com/ferroelectrics/negativec
**License:** GPL-3.0
**Language:** Python + C (wrapping FERRET/ngspice)
**GPU Support:** ❌ CPU (via FERRET backend)
**Development Status:** Research

### Description

negativec is a toolkit for simulating negative capacitance (NC) effects in ferroelectric materials and devices. It integrates phase-field simulations (FERRET) with circuit models to study NC-FETs and NC heterostructures.

### Key Features

- **Negative capacitance calculation:** From polarization dynamics
- **S-curve visualization:** Current-voltage in steep-slope regime
- **Domain wall evolution:** Real-time snapshots during NC operation
- **FeFET + dielectric stacks:** Coupled simulation
- **Energy analysis:** Thermodynamic cost of NC operation

### Relevant to FeCIM

Negative capacitance is important for:

1. **Lower switching voltage:** Reduce power for ferroelectric devices
2. **Steep subthreshold slope:** Below 60 mV/dec for low-power logic
3. **Interface engineering:** HfO₂/dielectric stacks
4. **Beyond-CMOS computing:** NC-based in-memory compute

### Installation

```bash
git clone https://github.com/ferroelectrics/negativec
cd negativec

# Requires FERRET and Python
pip install numpy scipy matplotlib pandas

# Build C extension (optional)
python setup.py build_ext --inplace
```

### Example: NC-FET S-Curve

```python
from negativec import NCFETStack, FerroelectricLayer, DielectricLayer

# Build stack: Al₂O₃ (dielectric) + HfO₂ (ferroelectric) + Si
stack = NCFETStack([
    DielectricLayer(material='Al2O3', thickness=2e-9, k=9.0),
    FerroelectricLayer(material='HfO2', thickness=10e-9, Pr=20e-6, Ec=1e8),
])

# Sweep gate voltage
Vg = np.linspace(0, 2, 100)  # 0-2 V

# Compute channel charge (Id-like)
Qinv = []
for vg in Vg:
    Q = stack.inversion_charge(vg)
    Qinv.append(Q)

# Plot S-curve
plt.figure()
plt.plot(Vg, Qinv, 'b-', linewidth=2)
plt.xlabel('Gate Voltage (V)')
plt.ylabel('Inversion Charge (C/m²)')
plt.title('NC-FET S-Curve (Steep Slope)')
plt.yscale('log')
plt.grid(True, alpha=0.3)
plt.show()

# Extract subthreshold slope (mV/dec)
dvg = Vg[1] - Vg[0]
dQ_dvg = np.gradient(Qinv, dvg)
SS = np.min(dQ_dvg)  # Steepest slope
print(f"Subthreshold slope: {SS:.1f} mV/dec")
```

### Comparison: With vs Without NC

| Metric | Standard FET | NC-FET | Improvement |
|--------|--------------|--------|-------------|
| Subthreshold slope (mV/dec) | 60 | 15-30 | 2-4× steeper |
| Turn-on voltage (V) | 0.5-1.0 | 0.2-0.4 | 50-60% lower |
| Off-state leakage | mA/μm | pA/μm | 1000× lower |
| Power consumption | 1× | 0.1-0.3× | 3-10× lower |

### Domain Wall Dynamics Visualization

```python
# Export frame-by-frame polarization snapshots
from negativec import FerroelectricSimulation

sim = FerroelectricSimulation(stack)

# Apply gate voltage step
sim.apply_voltage_step(0, 1.5, duration=1e-9, n_steps=100)

# Capture polarization field
for i in range(100):
    P_field = sim.get_polarization_field(step=i)

    # Save as image
    plt.figure()
    plt.imshow(P_field, cmap='RdBu', vmin=-30, vmax=30)
    plt.colorbar(label='Polarization (µC/cm²)')
    plt.title(f'Domain Evolution: t={i*10}ps')
    plt.savefig(f'domain_frame_{i:03d}.png', dpi=100)
    plt.close()

# Create animation
import subprocess
subprocess.run([
    'ffmpeg', '-framerate', '10',
    '-i', 'domain_frame_%03d.png',
    '-c:v', 'libx264', '-pix_fmt', 'yuv420p',
    'domain_evolution.mp4'
])
```

### Why negativec Matters

1. **Emerging technology:** NC-FETs not mainstream yet
2. **Fundamental physics:** Understanding domain dynamics
3. **Design optimization:** Parameter screening for best SS
4. **Integration:** Path to 3D ferroelectric logic

### Resources

- **GitHub:** https://github.com/ferroelectrics/negativec
- **Paper:** "Negative Capacitance in Ferroelectrics" (published in Nature Electronics, etc.)
- **FERRET integration:** See FERRET documentation

---

## Comparison Matrix: All Tools

| Tool | Scale | Physics Model | Speed | License | HfO₂ Support | Best For |
|------|-------|---------------|-------|---------|--------------|----------|
| **FERRET** | Meso | Landau | Slow (CPU) | LGPL | Via params | Domain structure |
| **FerroX** | Meso | TDGL | Fast (GPU) | BSD-3 | Native | 3D device simulation |
| **feram** | Macro | Eff. Hamiltonian | Fast | GPL | No | ABO₃ perovskites |
| **pymatgen** | DFT | Berry phase | Varies | MIT | Via DFT | Material screening |
| **Q-POP-Thermo** | Macro | LGD | Fast | MIT | Via params | Phase diagrams |
| **Ferro** | Analysis | Preisach | Fast | CC BY-NC-SA | Via data | Experimental analysis |
| **PFECAP** | Device | Preisach | Very fast | GPL-2.0 | Yes | SPICE circuits |
| **negativec** | Device | FERRET+TDGL | Slow | GPL-3.0 | Native | NC dynamics |

---

## Recommended Workflow for FeCIM

### Scenario 1: Material Discovery (New Ferroelectric)

```
Step 1: DFT Screening
├─ Use pymatgen + VASP
├─ Calculate Pr, Ec, dielectric
└─ Filter high-performing candidates

Step 2: Phase Diagram Mapping
├─ Use Q-POP-Thermo
├─ Generate T-E, T-σ diagrams
└─ Identify stability window

Step 3: Device Simulation (Proof-of-Concept)
├─ Use FerroX (3D domain structure)
├─ Validate P-E curves
└─ Estimate switching dynamics

Step 4: Circuit Integration
├─ Extract to PFECAP parameters
├─ Simulate 1T1C cells
└─ Estimate memory density, power
```

### Scenario 2: Circuit Design (FeCIM 1T1R Array)

```
Step 1: Device Characterization
├─ Use Ferro (parse P-E measurements)
├─ Extract Pr, Ec, tau
└─ Validate Preisach fit

Step 2: Compact Model
├─ Use PFECAP (circuit model)
├─ Tune parameters to match device
└─ Validate against SPICE

Step 3: Array Simulation
├─ Build 32×32 crossbar netlist
├─ Include non-idealities (IR drop, sneak paths)
├─ Simulate read/write cycles
└─ Estimate endurance, retention

Step 4: System Integration
├─ Export to NeuroSim / CrossSim
├─ Simulate neural network accuracy
└─ Compare with RRAM/PCM alternatives
```

### Scenario 3: Physical Validation (Lab Test)

```
Step 1: Measure P-E Loop
├─ Use experimental tracer
├─ Collect raw CSV/HDF5
└─ Extract Pr, Ec, coercivity

Step 2: Preisach Fitting
├─ Use Ferro package
├─ Fit to experimental data
└─ Extract hysteron distribution

Step 3: Compare with Simulation
├─ Run PFECAP with fitted parameters
├─ Plot simulation vs. measurement
└─ Quantify accuracy (RMSE, etc.)

Step 4: Optimize Device
├─ Identify deviations (fatigue, wake-up)
├─ Run FerroX for domain analysis
└─ Propose design improvements
```

---

## Quick Installation Guide

### Minimal Stack (For Learning)

```bash
# Python-only (no compilation needed)
pip install pymatgen numpy scipy matplotlib

# Ferro (experimental analysis)
git clone https://github.com/JAnderson419/Ferro
cd Ferro && pip install -e .

# Run demo
python -c "from ferro import HysteresisData; print('Ferro ready!')"
```

**Time to ready:** 5 minutes

### Full Stack (For Production)

```bash
# 1. Install system dependencies
sudo apt-get install -y \
  build-essential cmake git \
  gfortran libblas-dev liblapack-dev \
  libopenmpi-dev openmpi-bin

# 2. FerroX (GPU optional)
git clone https://github.com/AMReX-Microelectronics/FerroX
cd FerroX && mkdir build && cd build
cmake -DAMReX_GPU_BACKEND=CUDA -DAMReX_CUDA_ARCH=8.0 ..
make -j 4
cd ../..

# 3. FERRET + MOOSE (complex, ~30 min)
# See MOOSE documentation: https://mooseframework.inl.gov/

# 4. OpenVAF (for Verilog-A compilation)
cargo install openvaf

# 5. PFECAP
git clone https://github.com/supadupaplex/pfecap

# 6. Python packages
pip install pymatgen numpy scipy matplotlib pandas

echo "Full stack installed!"
```

**Time to ready:** 30-60 minutes (depending on GPU drivers)

---

## Integration with FeCIM Project

### Where These Tools Fit

```
┌─────────────────────────────────────────────────────────────┐
│  FeCIM Lattice Tools (Unified Go Visualizer)                │
├─────────────────────────────────────────────────────────────┤
│  Module 1: Hysteresis                                        │
│  ├─ Current: Preisach model (Go implementation)             │
│  └─ Integration: Import P-E from FerroX / pymatgen          │
├─────────────────────────────────────────────────────────────┤
│  Module 2: Crossbar                                         │
│  ├─ Current: MVM + drift/sneak paths                        │
│  └─ Integration: Conductance from PFECAP SPICE             │
├─────────────────────────────────────────────────────────────┤
│  Module 3: MNIST                                            │
│  ├─ Current: Neural network simulation                      │
│  └─ Integration: Device models from CrossSim / AIHWKIT     │
└─────────────────────────────────────────────────────────────┘
        ↓
  (Export/Import Interface)
        ↓
┌─────────────────────────────────────────────────────────────┐
│  External Tools Ecosystem                                    │
├─────────────────────────────────────────────────────────────┤
│  DFT:          pymatgen + VASP                              │
│  Phase-field:  FerroX or FERRET                             │
│  Devices:      PFECAP or negativec                          │
│  Arrays:       CrossSim or AIHWKIT                          │
│  Analysis:     Ferro or custom Python                       │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow Example: From DFT to Visualization

```
pymatgen (extract DFT Pr, Ec)
    ↓
Q-POP-Thermo (generate phase diagram)
    ↓
FerroX (simulate domain structure)
    ↓
Export P-E curves (JSON/CSV)
    ↓
Import to FeCIM Module 1 (Go visualizer)
    ↓
Display real-time hysteresis loop
    ↓
Module 2 uses conductance for crossbar sim
```

### Proposed Export Formats

**JSON (Material Parameters):**
```json
{
  "material": "HfO2",
  "Pr": 20e-6,
  "Ps": 25e-6,
  "Ec": 1.0e8,
  "tau": 1e-9,
  "alpha1": -1.72e8,
  "source": "FerroX simulation"
}
```

**CSV (P-E Loop):**
```
E_field_MV_cm, Polarization_uC_cm2
0.0, 0.0
0.1, 2.3
...
-2.0, -18.5
```

---

## Troubleshooting & Tips

### Common Issues

| Problem | Solution |
|---------|----------|
| FerroX build fails on GPU | Check CUDA arch: `nvidia-smi --query-gpu=compute_cap --format=csv,noheader` |
| FERRET requires MOOSE | See https://mooseframework.inl.gov/getting_started/ |
| PFECAP not compiling | Ensure OpenVAF installed: `openvaf --version` |
| pymatgen DFT too slow | Pre-compute and cache results; use batch mode |
| SPICE simulation diverges | Reduce time step or add `gmin` parameter |

### Performance Tips

1. **FerroX GPU:** Use `amr.max_level=1` for coarse grids first
2. **FERRET:** Parallelize with `mpirun -n 8`
3. **pymatgen:** Use `Materials Project API` to skip redundant DFT
4. **PFECAP:** Pre-compile Verilog-A to `.osdi` once

### Validation Checklists

**Before publishing results:**

- [ ] P-E curves validated against literature
- [ ] Landau coefficients from peer-reviewed source
- [ ] Domain structure matches microscopy (if available)
- [ ] Switching time constants reasonable (~ns for HfO₂)
- [ ] Temperature dependence physically sensible
- [ ] Endurance fatigue model calibrated

---

## References & Further Reading

### Primary Papers

1. **FerroX:** "FerroX: A GPU-accelerated, 3D Phase-Field Simulation Framework for Modeling Ferroelectric Devices" (Computer Physics Communications 2023)
   - arXiv: https://arxiv.org/abs/2210.15668

2. **FERRET:** "Phase-field modeling of ferroic materials using MOOSE" (In FERRET documentation)
   - GitHub: https://mangerij.github.io/ferret/

3. **Q-POP-Thermo:** "Landau-Ginzburg-Devonshire Thermodynamics Solver" (Computer Physics Communications 2022)

4. **PFECAP:** "Circuit-level Preisach Ferroelectric Capacitor Model" (Various SPICE tool documentation)

5. **IBM AIHWKIT:** "Using the IBM Analog In-Memory Hardware Acceleration Kit for Neural Network Training and Inference" (arXiv:2307.09357)

### Review Articles

- "Analog In-Memory Computing: From GPU to Photonics" (arXiv, 2024)
- "Ferroelectric Materials for Computing-in-Memory" (IEEE JSSC, 2023)
- "Domain Wall Engineering in Ferroelectrics" (Nature Reviews Materials, 2023)

### Online Courses & Tutorials

- MOOSE Framework Tutorials: https://mooseframework.inl.gov/
- pymatgen Documentation: https://pymatgen.org/
- SPICE Simulation Basics: ngspice manual + tutorials
- FerroX GitHub Wiki: https://github.com/AMReX-Microelectronics/FerroX/wiki

### Key Literature for HfO₂

- **Pr & Ec values:** Nature Communications 2025 (HZO measurements)
- **3D integration:** CEA-Leti December 2024 (22nm BEOL)
- **Endurance:** IEEE IRPS 2022 (10⁹-10¹² cycles)
- **Cryogenic operation:** IEEE/Frontiers 2024 (5K operation)

---

## Contributing to These Tools

### For Researchers

- **FERRET/FerroX:** Submit domain structure images, new material models
- **pymatgen:** Contribute DFT-calculated materials data
- **Q-POP-Thermo:** Extend to new ferroelectric systems
- **negativec:** Add new stack geometries, materials

### For Developers

- **GitHub Issues:** Report bugs, request features
- **Pull Requests:** Contribute code improvements
- **Documentation:** Fix typos, clarify concepts
- **Performance:** Optimize critical sections (especially GPU code)

---

## Summary Table: Quick Reference

| Need | Recommended Tool | Why |
|------|------------------|-----|
| Visualize P-E curves | This project (Module 1) | Real-time, interactive |
| Screen new materials | pymatgen + VASP | Berry phase accuracy |
| Domain wall dynamics | FerroX | GPU-accelerated, physics-based |
| SPICE circuit design | PFECAP | Fast, circuit-compatible |
| Negative capacitance | negativec | Specialized physics |
| Analyze measurement | Ferro | Quick parameter extraction |
| Temperature effects | Q-POP-Thermo | LGD free energy |
| Publication plots | Any tool | Export to matplotlib |

---

*This document is part of the FeCIM Lattice Tools project. Last verified: January 2026. For updates, see the project repository: https://github.com/fecim/lattice-tools*
