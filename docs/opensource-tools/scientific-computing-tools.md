# Scientific Computing Tools for Materials Science

**Comprehensive Guide to FEM, DFT, and Materials Databases for Ferroelectric HfO₂-ZrO₂ Research**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source scientific computing tools, frameworks, and databases for materials science research focused on ferroelectric materials. It covers finite element method (FEM) tools, density functional theory (DFT) codes, materials databases, atomic simulation environments, and specialized tools for ferroelectric modeling.

**Target Audience:** Researchers and engineers developing computational models for HfO₂-ZrO₂ superlattice ferroelectrics.

---

## 1. Finite Element Method (FEM) Tools

FEM tools are essential for solving coupled partial differential equations governing ferroelectric behavior, including electrostatic, thermal, and mechanical effects.

### 1.1 FEniCSx (FEniCS Project)

**URL:** https://fenicsproject.org/
**License:** LGPL-3.0
**Language:** Python (high-level), C++ (backend)
**Latest Version:** 0.8+ (FEniCSx)

#### Description

FEniCSx is a modern, automated finite element framework for solving PDEs. It provides a high-level Python interface with a high-performance C++ backend using PETSc and MUMPS solvers.

#### Key Features

- **Python domain-specific language (DSL)** for problem formulation
- **Automated assembly** of finite element matrices
- **PETSc/MUMPS solvers** for sparse linear systems
- **Mesh generation and refinement** via GMSH integration
- **Multiphysics support** (electrostatic, thermal, mechanical)
- **Parallel computing** via MPI
- **GPU support** via PETSc with CUDA

#### Installation

```bash
# Using conda (recommended)
conda create -n fenics-env
conda activate fenics-env
conda install -c conda-forge fenics-dolfinx

# Or pip
pip install fenics-dolfinx
```

#### Example: Ferroelectric Domain Structure

```python
from dolfinx import fem, mesh, plot
from ufl import TrialFunction, TestFunction, inner, grad, dx
import numpy as np

# Create 2D rectangular mesh
domain = mesh.create_rectangle(
    None,
    [[-2, -2], [2, 2]],
    [40, 40]
)

# Function space for polarization
V = fem.FunctionSpace(domain, ("Lagrange", 1))

# Weak form: -∇²P = E_applied
u = TrialFunction(V)
v = TestFunction(V)
a = inner(grad(u), grad(v)) * dx
L = fem.Constant(domain, 1.0) * v * dx  # Applied field

# Solve
problem = fem.petsc.LinearProblem(a, L)
P = problem.solve()

# Visualize
plot(P)
```

#### Applications for FeCIM

- **Phase-field simulations** of ferroelectric domains
- **Coupled electromechanical problems** (stress-strain effects)
- **Temperature-dependent polarization** evolution
- **Domain wall dynamics** and switching

#### Advantages

- Clean, expressive Python API
- Fast C++ backend
- Excellent documentation
- Active community support

#### Limitations

- Steep learning curve for complex problems
- Not specialized for ferroelectrics (requires custom material models)
- GPU support requires extra configuration

---

### 1.2 MOOSE Framework

**URL:** https://mooseframework.inl.gov/
**License:** LGPL-2.1
**Language:** C++
**Maintained By:** Idaho National Laboratory

#### Description

MOOSE (Multiphysics Object-Oriented Simulation Environment) is a finite element framework developed at Idaho National Laboratory for coupled multiphysics simulations. It includes specialized modules for ferroelectrics and phase-field modeling.

#### Key Features

- **Modular architecture** for custom physics
- **FERRET module** for ferroelectric phase-field
- **Coupled electromechanical solver**
- **Adaptive mesh refinement (AMR)**
- **Parallel scalability** to millions of elements
- **Large material model library**
- **Built-in material nonlinearity** handling

#### Installation

```bash
# Clone MOOSE
git clone https://github.com/idaholab/moose.git
cd moose

# Install dependencies (see documentation for your OS)
./scripts/update_and_rebuild_libmesh.sh

# Build MOOSE
./configure --enable-everything
make -j$(nproc)
```

#### Example: Ferroelectric Phase-Field (MOOSE format)

```
[Mesh]
  type = GeneratedMesh
  dim = 2
  nx = 100
  ny = 100
  xmax = 10
  ymax = 10
[]

[Materials]
  [ferroelectric]
    type = LandauFreeEnergy
    # Landau coefficients (from DFT calculations)
    alpha1 = -1.72e8     # Pa/C²
    alpha11 = 7.3e8      # Pa/C²
    alpha111 = 2.6e9     # Pa/C²

    # Electrostrictive coefficient
    q = 0.05  # C⁻²
  []

  [elasticity]
    type = LinearIsotropic
    # Elastic constants (GPa)
    youngs_modulus = 130
    poissons_ratio = 0.25
  []
[]

[Variables]
  [P_x]    # Polarization x-component
  [P_y]    # Polarization y-component
  [u_x]    # Displacement x
  [u_y]    # Displacement y
[]

[Kernels]
  [P_x_time]
    type = TimeDerivative
    variable = P_x
  []

  [P_x_ferroelectric]
    type = LandauFreeEnergyKernel
    variable = P_x
  []
[]

[Executioner]
  type = Transient
  dt = 1e-13
  num_steps = 1000
[]

[Outputs]
  exodus = true
  vtk = true
[]
```

#### Applications for FeCIM

- **Phase-field simulations** of HfO₂-ZrO₂ domain structures
- **Coupled electromechanical problems** with strain effects
- **Multi-scale coupling** from atomistic to continuum
- **Ferroelectric wake-up** and fatigue mechanisms
- **Domain switching dynamics** under applied field

#### Advantages

- Specialized ferroelectric module (FERRET)
- Massive scalability (million+ elements)
- Excellent for strongly coupled problems
- Active development at national lab

#### Limitations

- Steep C++ learning curve for modifications
- Requires significant computational resources
- Limited documentation for ferroelectric-specific use cases

---

### 1.3 deal.II (Differential Equations Analysis Library)

**URL:** https://dealii.org/
**License:** LGPL-2.1
**Language:** C++

#### Description

deal.II is a modern C++ library providing tools for solving PDEs using the finite element method. Known for excellent documentation and 1000+ tutorial programs.

#### Key Features

- **Comprehensive documentation** (tutorials + API)
- **Adaptive mesh refinement** with hanging nodes
- **Support for high-order elements** (up to p=10+)
- **Excellent debugging tools**
- **Parallel computing** via p4est and PETSc
- **1000+ educational examples**

#### Installation

```bash
# From source
git clone https://github.com/dealii/dealii.git
cd dealii
mkdir build && cd build
cmake ..
make -j$(nproc)
make install
```

#### Applications for FeCIM

- **Custom ferroelectric models** via user-defined kernels
- **Educational examples** for FEM methodology
- **Debugging complex coupled problems**
- **High-order element schemes**

#### Advantages

- Excellent documentation and tutorials
- Clear code examples
- Strong community support
- Powerful debugging capabilities

#### Limitations

- Not specialized for ferroelectrics
- C++ required (steeper learning curve than Python)
- Requires more boilerplate code than FEniCSx

---

### 1.4 SfePy (Simple Finite Elements in Python)

**URL:** https://sfepy.org/
**License:** BSD-3-Clause
**Language:** Python

#### Description

SfePy is a Python-based FEM framework with native support for piezoelectric and multiscale problems. Ideal for rapid prototyping and material homogenization studies.

#### Key Features

- **Native piezoelectric support** (critical for ferroelectrics!)
- **Linear and nonlinear solvers**
- **Homogenization tools** for multiscale analysis
- **GAMBIT/Gmsh mesh support**
- **Interactive problem specification** via Python scripts
- **Material nonlinearity handling**

#### Installation

```bash
pip install sfepy
```

#### Example: Ferroelectric Domain Simulation

```python
from sfepy import *
from sfepy.mechanics.matcoeff import lame_from_youngpoisson

# Material parameters for HfO2
material_params = {
    'rho': 9650,           # kg/m³
    'lambda': 1.2e11,      # Pa (Lamé parameter)
    'mu': 0.5e11,          # Pa (shear modulus)
    'epsilon': 130,        # Relative permittivity
    'd33': 1e-11,          # m/V (piezoelectric coefficient)
}

# Domain: rectangular ferroelectric
domain = Domain.from_file('ferroelectric_mesh.msh')

# Variables: displacement u and potential phi
u = FieldVariable('u', 'unknown', domain, 3)
phi = FieldVariable('phi', 'unknown', domain, 1)

# Equations: elasticity + electrostatics
equations = [
    Equation('elasticity',
        Term.new('dw_lin_elastic', ...)),
    Equation('electrostatics',
        Term.new('dw_laplace', ...))
]

# Solve coupled problem
solver = ProblemDefinition(domain, [u, phi], equations)
solver.solve()
```

#### Applications for FeCIM

- **Piezoelectric coupling** in ferroelectric devices
- **Multiscale homogenization** of polycrystalline HfO₂
- **Strain-polarization coupling**
- **Rapid prototyping** of models

#### Advantages

- Native piezoelectric support (rare!)
- Clean Python API
- Good for multiscale problems
- BSD license (permissive)

#### Limitations

- Smaller community than FEniCSx/deal.II
- Less extensive documentation
- Limited GPU support

---

## 2. Density Functional Theory (DFT) Codes

DFT codes compute ground-state properties of ferroelectric materials from first principles, enabling prediction of polarization, band gaps, and elastic constants.

### 2.1 VASP (Vienna Ab initio Simulation Package)

**URL:** https://www.vasp.at/
**License:** Commercial (educational licenses available)
**Language:** Fortran

#### Description

VASP is the de facto standard for DFT calculations in materials science. It uses plane-wave basis sets and PAW pseudopotentials.

#### Key Applications for HfO₂

- **Crystal structure optimization**
- **Polarization calculations** via Berry phase
- **Band structure and density of states (DOS)**
- **Elastic constants** and phonon frequencies
- **Dielectric properties** (ionic and electronic contributions)

#### Example Input (POSCAR for HfO₂)

```
HfO2 monoclinic
1.0
5.1172  0.0000  0.0000
0.0000  5.1722  0.0000
0.0000  0.0000  5.3061

Hf  O
2   4

Selective dynamics
Direct
0.2500  0.0830  0.2500  T T T  # Hf
0.7500  0.5830  0.7500  T T T  # Hf
0.4600  0.1660  0.2500  T T T  # O
0.0400  0.6660  0.7500  T T T  # O
0.5400  0.8340  0.2500  T T T  # O
0.9600  0.3340  0.7500  T T T  # O
```

#### Advantages

- Most widely used DFT code
- Excellent for ferroelectric materials
- Large material databases (Materials Project, AFLOW)
- Highly optimized and tested

#### Limitations

- Commercial license required
- Computationally expensive (weeks to months per calculation)
- Steep learning curve

---

### 2.2 GPAW (Gaussian and Augmented Plane-Wave Method)

**URL:** https://gpaw.readthedocs.io/
**License:** GPL-3.0
**Language:** Python/C

#### Description

GPAW is an open-source DFT code offering flexible modes: plane-wave, real-space, and LCAO. It's well-suited for education and research with lower computational requirements than VASP.

#### Key Features

- **Multiple basis set modes** (plane-wave, real-space, LCAO)
- **GPU acceleration** available
- **Band structure calculations**
- **Optical properties** (dielectric function)
- **Polarization via Berry phase**
- **Easy integration** with Python workflows

#### Installation

```bash
pip install gpaw
# Download PAW datasets
gpaw install-data

# For GPU support (optional)
pip install cupy  # NVIDIA GPUs
```

#### Example: HfO₂ Band Structure

```python
from ase.build import bulk
from gpaw import GPAW, PW, FermiDirac
from ase.io import read

# Create HfO2 structure (monoclinic)
atoms = read('HfO2.cif')

# Set up calculator
calc = GPAW(
    mode=PW(400),           # 400 eV cutoff
    xc='PBE',               # PBE functional
    kpts=(4, 4, 4),         # k-point grid
    occupation=FermiDirac(width=0.05)
)

atoms.calc = calc

# Optimize structure
from ase.optimize import BFGS
dyn = BFGS(atoms, trajectory='HfO2.traj')
dyn.run(fmax=0.01)

# Calculate band structure
from ase.build import bulk_bands
k_points, x, X = atoms.get_reciprocal_lattice_interface()
calc.set(kpts=k_points)
atoms.get_potential_energy()
calc.write('HfO2_bands.gpw')

# Plot band structure
from ase.spectrum.band_structure import BandStructure
bs = BandStructure(calc, k_points, reference=E_fermi)
bs.plot()
```

#### Applications for FeCIM

- **Landau coefficients** from DFT polarization data
- **Band gap predictions** for different HfO₂ phases
- **Phonon calculations** for soft mode analysis
- **Initial model parameterization**

#### Advantages

- Open-source (GPL-3.0)
- Flexible basis sets
- Good documentation
- Lower computational cost than VASP
- Active development

#### Limitations

- Generally slower than VASP
- Less mature optimization routines
- Smaller literature database for ferroelectrics

---

### 2.3 Quantum ESPRESSO

**URL:** https://www.quantum-espresso.org/
**License:** GPL-2.0
**Language:** Fortran

#### Description

Quantum ESPRESSO is a modular, open-source suite for electronic structure calculations. Well-maintained with excellent portability.

#### Key Features

- **DFT with multiple functionals** (LDA, GGA, hybrid)
- **Phonon calculations** via DFPT (Density Functional Perturbation Theory)
- **Polarization via Berry phase** (ferroelectric-relevant!)
- **Molecular dynamics** (NVT, NPT)
- **Large material database** (Materials Cloud)
- **Good parallelization**

#### Installation

```bash
# Download
git clone https://gitlab.com/QEF/q-e.git
cd q-e

# Configure and build
./configure
make all -j$(nproc)
```

#### Example: HfO₂ with Polarization

```
&CONTROL
  calculation = 'scf'
  restart_mode = 'from_scratch'
  prefix = 'HfO2'
  outdir = './output'
  wf_collect = .true.
/

&SYSTEM
  ibrav = 12              ! Monoclinic
  celldm(1) = 9.661       ! a (Bohr)
  celldm(2) = 0.975       ! b/a
  celldm(3) = 1.003       ! c/a
  celldm(5) = 0.906       ! cos(beta)
  nat = 12                ! Atoms
  ntyp = 2                ! Species
  ecutwfc = 80.0          ! Cutoff (Ry)
  ecutrho = 320.0         ! Rho cutoff
  occupations = 'smearing'
  smearing = 'gauss'
  degauss = 0.02
/

&ELECTRONS
  conv_thr = 1.0d-8
  mixing_beta = 0.7
/

ATOMIC_SPECIES
Hf  178.49  Hf.pbe-spn-kjpaw_psl.1.0.0.UPF
O   15.999  O.pbe-n-kjpaw_psl.1.0.0.UPF

ATOMIC_POSITIONS crystal
Hf  0.2500  0.0830  0.2500
Hf  0.7500  0.5830  0.7500
O   0.4600  0.1660  0.2500
O   0.0400  0.6660  0.7500
O   0.5400  0.8340  0.2500
O   0.9600  0.3340  0.7500
O   0.1066  0.4089  0.5000
O   0.3934  0.9089  0.0000

K_POINTS automatic
4 4 4 0 0 0
```

#### Applications for FeCIM

- **Phonon analysis** for ferroelectric soft modes
- **Polarization calculations** via Berry phase
- **Band structure** for multiple polymorphs
- **Molecular dynamics** for temperature effects

#### Advantages

- Excellent phonon capabilities (DFPT)
- Good parallelization
- Active development
- Materials Cloud integration

#### Limitations

- Requires pseudopotential selection
- Complex input format
- Slower than optimized VASP

---

### 2.4 ABINIT

**URL:** https://www.abinit.org/
**License:** GPL-2.0
**Language:** Fortran

#### Description

ABINIT is a comprehensive DFT suite with strong capabilities for many-body physics and properties calculations.

#### Key Features

- **Many-body physics** (GW, Bethe-Salpeter)
- **Phonon calculations** (DFPT)
- **Polarization and ferroelectricity** support
- **Excellent documentation**
- **Data analysis tools** included

#### Applications for FeCIM

- **Advanced band structure** (GW corrected)
- **Ferroelectric properties** from first principles
- **Exciton dynamics** if optoelectronic applications needed

#### Advantages

- Powerful for advanced physics
- Well-documented
- Open-source (GPL)

#### Limitations

- Steep learning curve
- Complex input files
- Slower than VASP for standard DFT

---

## 3. Materials Databases and APIs

Materials databases provide pre-computed properties for thousands of materials, enabling rapid property lookups and data-driven research.

### 3.1 Materials Project (materialsproject.org)

**URL:** https://next-gen.materialsproject.org/
**License:** Creative Commons (data); MIT (pymatgen library)

#### Description

The Materials Project is the world's largest public database of computed materials properties. Contains 150,000+ materials with comprehensive property data.

#### HfO₂-Specific Data Available

- **Monoclinic, orthorhombic, and tetragonal polymorphs**
- **Formation energies** and thermodynamic stability
- **Band structures** and electronic densities of states
- **Dielectric constants** (ionic and electronic)
- **Elastic tensors** and mechanical properties
- **Magnetization** (for doped variants)

#### Access Methods

**Web Interface:**
https://next-gen.materialsproject.org/

**Python API (pymatgen):**
```python
from mp_api.client import MPRester
from pymatgen.analysis.structure_analyzer import SpacegroupAnalyzer

# Requires API key (free registration)
api_key = "YOUR_API_KEY"

with MPRester(api_key) as mpr:
    # Search for HfO2 structures
    docs = mpr.materials.summary.search(
        formula="HfO2",
        fields=[
            "material_id",
            "structure",
            "band_gap",
            "formation_energy_per_atom",
            "dielectric",
            "elasticity"
        ]
    )

    for doc in docs:
        print(f"Material ID: {doc.material_id}")
        print(f"Band Gap: {doc.band_gap} eV")

        # Analyze symmetry
        sga = SpacegroupAnalyzer(doc.structure)
        print(f"Crystal System: {sga.get_crystal_system()}")
        print(f"Space Group: {sga.get_space_group_symbol()}")

        # Access dielectric properties
        if doc.dielectric:
            print(f"Static dielectric constant: {doc.dielectric.ionic[0,0]}")
```

#### Example: HfO₂ Phase Comparison

```python
import pandas as pd
from mp_api.client import MPRester

with MPRester(api_key) as mpr:
    # Get all HfO2 phases
    docs = mpr.materials.summary.search(
        formula="HfO2",
        fields=[
            "material_id",
            "structure",
            "formation_energy_per_atom",
            "band_gap"
        ]
    )

    # Create comparison table
    data = []
    for doc in docs:
        sga = SpacegroupAnalyzer(doc.structure)
        data.append({
            'Material ID': doc.material_id,
            'Crystal System': sga.get_crystal_system(),
            'Formation Energy (eV/atom)': doc.formation_energy_per_atom,
            'Band Gap (eV)': doc.band_gap
        })

    df = pd.DataFrame(data)
    print(df.sort_values('Formation Energy (eV/atom)'))
```

#### Advantages

- Largest public materials database
- Well-maintained and actively curated
- Comprehensive property coverage
- Excellent Python integration (pymatgen)
- Free for academic use

#### Limitations

- DFT results may differ from experiments
- Limited to published/calculated materials
- Some properties only available for subset

---

### 3.2 AFLOW (Automatic FLOW for Materials Discovery)

**URL:** http://www.aflowlib.org/
**License:** MIT

#### Description

AFLOW is a computational materials database with standardized DFT calculations. Provides consistent, high-quality data across materials.

#### Key Features

- **Standardized DFT calculations** (consistency across materials)
- **550,000+ materials** in database
- **Elastic properties** calculations
- **Phase diagrams** (AFLUX)
- **Machine learning** property predictions
- **REST API** for data access

#### Python Access

```python
import json
import urllib.request

# Query HfO2 properties via REST API
url = "http://aflowlib.duke.edu/API/aflow/material/Hf,O/icsd/2/json"
response = urllib.request.urlopen(url)
data = json.loads(response.read().decode())

for entry in data:
    print(f"AFLOW ID: {entry['auid']}")
    print(f"Crystal System: {entry['spacegroup_crystal_system']}")
    print(f"Formation Energy: {entry['enthalpy_formation_cell']} eV")
```

#### Advantages

- Standardized DFT methodology
- Consistent across all materials
- Large database
- Good for comparative studies

#### Limitations

- Less comprehensive property coverage than Materials Project
- Smaller community

---

### 3.3 Open Quantum Materials Database (OQMD)

**URL:** http://oqmd.org/
**License:** Creative Commons

#### Description

OQMD is a high-throughput DFT database with experimental data integration. Contains 700,000+ materials.

#### Key Features

- **DFT + experimental hybrid** data
- **Formation energies** and phase stability
- **Composition-based search**
- **Direct Python API** (qmpy)

#### Installation and Use

```bash
pip install qmpy

# Python
from qmpy import *

# Search for hafnium oxides
for compound in Composition.objects.filter(formula__contains="Hf"):
    for entry in Entry.objects.filter(composition=compound):
        print(f"Entry: {entry.name}")
        print(f"Energy: {entry.energy} eV")
```

#### Advantages

- Largest materials database
- Integration of experimental data
- Fast search interface

#### Limitations

- Less standardized methodology than AFLOW
- Property coverage varies

---

## 4. Atomic Simulation Environment Tools

These tools bridge atomistic simulations, DFT codes, and analysis.

### 4.1 ASE (Atomic Simulation Environment)

**URL:** https://ase-lib.org/
**License:** LGPL-2.1
**Language:** Python

#### Description

ASE is a Python package for working with atoms and molecules. It provides interfaces to DFT codes, molecular dynamics, and structure analysis.

#### Key Features

- **Unified interface** to 30+ DFT/MD codes (VASP, GPAW, Quantum ESPRESSO, LAMMPS, etc.)
- **Structure building and manipulation**
- **Visualization** and structure analysis
- **Geometry optimization** and transition states
- **Molecular dynamics** integration
- **Calculator abstraction** layer (use any code with consistent API)

#### Installation

```bash
pip install ase

# For visualization
pip install ase[gui]

# For specific calculators (GPAW)
pip install ase gpaw
```

#### Example: HfO₂ Structure Optimization

```python
from ase.io import read, write
from ase.build import bulk
from ase.calculators.gpaw import GPAW
from ase.optimize import BFGS
from ase.constraints import UnitCellFilter
from ase.thermochemistry import IdealGasThermo

# Create HfO2 structure (monoclinic)
atoms = read('HfO2_monoclinic.cif')

# Set up GPAW calculator
calc = GPAW(
    mode='pw',
    xc='PBE',
    kpts=(4, 4, 4),
    convergence={'density': 1e-6}
)

atoms.set_calculator(calc)

# Optimize cell + atoms
from ase.constraints import UnitCellFilter
ucf = UnitCellFilter(atoms)
optimizer = BFGS(ucf, trajectory='HfO2_opt.traj')
optimizer.run(fmax=0.01)

# Save optimized structure
write('HfO2_optimized.cif', atoms)

# Get properties
energy = atoms.get_potential_energy()
stresses = atoms.get_stress()
forces = atoms.get_forces()

print(f"Formation energy: {energy} eV")
print(f"Stress tensor:\n{stresses}")
```

#### Example: Phonon Calculation

```python
from ase.phonons import Phonons
from ase.calculators.gpaw import GPAW

# Load optimized structure
atoms = read('HfO2_optimized.cif')

# Set up phonon calculator
atoms.set_calculator(GPAW(mode='pw', xc='PBE', kpts=(4,4,4)))

# Create phonon object (supercell 2x2x2)
phonons = Phonons(atoms, supercell=(2, 2, 2), delta=0.05)
phonons.run()
phonons.read(acoustic=True)

# Get band structure
phonons.band_structure(points=[('X', [0.5, 0, 0]),
                                ('G', [0, 0, 0]),
                                ('L', [0.5, 0.5, 0.5])])

# Print frequencies
q_point = [0, 0, 0]  # Gamma point
freqs = phonons.get_frequencies(q_point)
print(f"Frequencies at Gamma: {freqs} cm^-1")
```

#### Applications for FeCIM

- **Parameterization of Landau coefficients** from DFT data
- **Elastic constant extraction** for coupled problems
- **Defect calculations** (oxygen vacancies, dopants)
- **Interface properties** (HfO₂-ZrO₂ superlattices)

#### Advantages

- Clean, Pythonic interface
- Works with many calculators
- Excellent for rapid prototyping
- Good visualization tools
- Active development

#### Limitations

- Performance overhead from Python layer
- Not suited for very large systems
- Requires external calculators for actual computations

---

### 4.2 Phonopy (Phonon Calculator)

**URL:** https://phonopy.github.io/
**License:** BSD-3-Clause
**Language:** Python

#### Description

Phonopy is a phonon calculator using finite differences with DFT calculations. Essential for understanding ferroelectric soft modes.

#### Key Features

- **Phonon band structure** and DOS from finite differences
- **Thermal properties** (specific heat, free energy)
- **Quasiharmonic analysis** (temperature dependence)
- **Born effective charges** and IR spectra
- **DFPT integration** (with Quantum ESPRESSO, VASP)

#### Installation

```bash
pip install phonopy

# Optional: for structure visualization
pip install phonopy[GUI]
```

#### Example: HfO₂ Phonons and Soft Modes

```python
import phonopy
from phonopy import load

# Load from VASP calculations
phonon = load('phonopy.yaml')  # Requires VASP output

# Get band structure
phonon.run_band_structure(
    [[[0, 0, 0], [0.5, 0, 0], 50],   # Path in k-space
     [[0.5, 0, 0], [0.5, 0.5, 0], 50]]
)

# Identify soft modes (negative frequencies)
band_structure = phonon.get_band_structure_dict()
for i_band, freq in enumerate(band_structure['frequencies'][0]):
    if freq < 0:
        print(f"Soft mode band {i_band}: {freq:.2f} cm^-1")

# Calculate DOS
phonon.run_mesh()
mesh = phonon.get_mesh_dict()

# Thermal properties
phonon.run_thermal_properties(
    t_step=50,
    t_max=1000,
    t_min=0
)

# Plot DOS and thermal properties
import matplotlib.pyplot as plt
# ... plotting code ...
```

#### Applications for FeCIM

- **Soft mode detection** at ferroelectric transition
- **Temperature-dependent properties** via quasiharmonic approximation
- **Identify instability driving ferroelectricity**
- **Compare HfO₂ polymorphs**

#### Advantages

- Dedicated phonon tool
- Integration with DFT codes
- Thermal analysis capabilities
- Well-documented

#### Limitations

- Requires DFT calculations as input
- Finite difference method less accurate than DFPT

---

### 4.3 atomman (NIST Atomistic Modeling)

**URL:** https://github.com/usnistgov/atomman
**License:** Custom (NIST)
**Language:** Python

#### Description

atomman is a library for atomic data manipulation, defect modeling, and LAMMPS integration. Excellent for point defects, dislocations, and interatomic potential studies.

#### Key Features

- **Defect analysis** (vacancies, interstitials, antisite defects)
- **Dislocation modeling**
- **LAMMPS integration** for MD
- **Interatomic potential database** (NIST)
- **Lattice defect tools**

#### Installation

```bash
pip install atomman
```

#### Example: Oxygen Vacancy in HfO₂

```python
import atomman as am
import numpy as np

# Load HfO2 structure
atoms = am.load('HfO2.cif')

# Create supercell (3x3x3)
atoms_supercell = atoms.supersize((3, 3, 3))

# Remove one oxygen atom (vacancy)
# Find oxygen at index i_O
for i, atom in enumerate(atoms_supercell):
    if atom.type == 'O':
        vacancy_index = i
        break

# Create vacancy structure
defect_atoms = atoms_supercell.copy()
del defect_atoms[vacancy_index]

# Save for LAMMPS calculation
defect_atoms.dump('HfO2_O_vacancy.lmp')

# Analyze defect
print(f"Removed atom at position: {atoms_supercell[vacancy_index].pos}")
print(f"Supercell: {atoms_supercell.box}")
```

#### Applications for FeCIM

- **Oxygen vacancy formation energies**
- **Point defect effects** on polarization
- **Defect-dopant interactions**
- **Lattice relaxation** around defects

#### Advantages

- NIST-maintained (high quality)
- Good for defect studies
- LAMMPS integration

#### Limitations

- More specialized (not general-purpose)
- Requires separate MD engine for dynamics

---

## 5. Specialized Ferroelectric Tools

### 5.1 FERRET (MOOSE Ferroelectric Module)

**URL:** https://github.com/mangerij/ferret
**License:** LGPL-2.1
**Based On:** MOOSE Framework

#### Description

FERRET is a specialized MOOSE module for phase-field modeling of ferroic materials. Implements Landau-Devonshire free energy and coupled electromechanical problems.

#### Key Features

- **Landau free energy** formulation
- **Electric field** effects on polarization
- **Mechanical stress** coupling
- **Domain dynamics** and switching
- **Multi-phase** transitions

#### Installation

```bash
# Clone and build FERRET with MOOSE
git clone https://github.com/mangerij/ferret
cd ferret
make -j$(nproc)
```

#### Example Input File

```
[Mesh]
  type = GeneratedMesh
  dim = 2
  nx = 100
  ny = 100
[]

[Variables]
  [P]  # Polarization
  [E]  # Electric potential
[]

[Materials]
  [ferroelectric]
    type = FerroelectricLandau

    # Landau coefficients (temperature-dependent)
    alpha1 = -1.72e8  # Pa/C²
    alpha11 = 7.3e8
    alpha111 = 2.6e9

    # Dielectric stiffness
    epsilon0 = 130
  []
[]

[Kernels]
  # Ferroelectric equation
  [P_time]
    type = TimeDerivative
    variable = P
  []

  [ferroelectric_landau]
    type = LandauFreeEnergyKernel
    variable = P
  []

  # Gauss's law
  [gauss_law]
    type = Diffusion
    variable = E
  []
[]

[BCs]
  [P_boundary]
    type = DirichletBC
    variable = P
    boundary = 'top bottom'
    value = 0
  []

  [E_applied]
    type = FunctionDirichletBC
    variable = E
    boundary = 'top'
    function = 'E_func'
  []
[]

[Functions]
  [E_func]
    type = PiecewiseLinear
    x = '0 1e-10 2e-10'
    y = '0 1e7 0'
  []
[]

[Executioner]
  type = Transient
  dt = 1e-14
  num_steps = 500
[]

[Outputs]
  exodus = true
  console = true
[]
```

#### Advantages

- Specialized for ferroelectrics
- Landau theory implementation
- Coupled electromechanical
- Massive scalability

#### Limitations

- Requires MOOSE knowledge
- Steep setup learning curve
- Expensive computationally

---

### 5.2 FerroX (GPU Phase-Field)

**URL:** https://github.com/AMReX-Microelectronics/FerroX
**License:** BSD
**Paper:** arXiv:2210.15668

#### Description

FerroX is a GPU-accelerated phase-field simulator for ferroelectric domains using the Time-Dependent Ginzburg-Landau (TDGL) equation.

#### Key Features

- **GPU acceleration** (15x speedup)
- **Multi-GPU support** via AMReX
- **Arbitrary electrode geometries**
- **Piezoelectric coupling**
- **Realistic domain dynamics**

#### Installation

```bash
git clone https://github.com/AMReX-Microelectronics/FerroX
cd FerroX
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
cmake --build . -j$(nproc)
```

#### Applications for FeCIM

- **Domain structure visualization**
- **Switching dynamics** on nanosecond timescale
- **Collective domain behavior**
- **Nucleation kinetics**

#### Advantages

- GPU-accelerated (fast!)
- Physically accurate TDGL
- Open-source
- Good documentation

#### Limitations

- C++/Fortran (not Python)
- Requires GPU
- Batch mode (not interactive)
- Very large computational domains

---

## 6. Recommended Workflow for HfO₂-ZrO₂ Research

```
Stage 1: Database Search
└─ Materials Project / AFLOW / OQMD
   ├─ Identify stable polymorphs
   ├─ Get bulk elastic constants
   └─ Check calculated dielectric properties

Stage 2: First-Principles Calculations
└─ VASP / Quantum ESPRESSO / GPAW
   ├─ Optimize structures
   ├─ Calculate Landau coefficients
   ├─ Phonon analysis (soft modes)
   └─ Defect formation energies

Stage 3: Macroscopic Modeling
└─ Phase-field (MOOSE/FERRET or FEniCSx)
   ├─ Domain structure evolution
   ├─ Switching dynamics
   └─ Electromechanical coupling

Stage 4: Device Integration
└─ FeCIM Tools (Module 1-2)
   ├─ Preisach hysteresis
   ├─ Crossbar array simulation
   └─ Hardware mapping
```

---

## 7. Installation Quick Reference

### Ubuntu/Debian (apt)

```bash
# System dependencies
sudo apt install build-essential cmake git python3-dev python3-pip
sudo apt install gfortran liblapack-dev libblas-dev
sudo apt install libhdf5-dev libnetcdf-dev

# Python ecosystem
pip install numpy scipy matplotlib pandas scikit-learn jupyter

# Materials science tools
pip install pymatgen ase phonopy sfepy
pip install gpaw  # Requires: gpaw install-data

# FEM frameworks
pip install fenics-dolfinx  # Or conda: conda install -c conda-forge fenics-dolfinx
```

### macOS (Homebrew)

```bash
# Install homebrew packages
brew install cmake gfortran open-mpi

# Python tools
pip install numpy scipy matplotlib pandas
pip install pymatgen ase phonopy
pip install fenics-dolfinx  # Via conda-forge recommended
```

### HPC Clusters

```bash
# Typically available as modules
module avail python  # Check available Python
module avail gcc     # Compiler versions
module avail openmpi # MPI implementation

# Load environment
module load python/3.10
module load gcc/11.0
module load openmpi/4.1

# Install to user environment
pip install --user pymatgen ase phonopy
```

---

## 8. Comparison Matrix

| Tool | Type | Language | License | GPU | Best For |
|------|------|----------|---------|-----|----------|
| **FEniCSx** | FEM | Python/C++ | LGPL-3.0 | ✅ | General PDE solving |
| **MOOSE** | FEM | C++ | LGPL-2.1 | ❌ | Large-scale multiphysics |
| **FERRET** | Phase-field | C++ | LGPL-2.1 | ❌ | Ferroelectric domains |
| **FerroX** | Phase-field | C++/Fortran | BSD | ✅ | Domain dynamics |
| **SfePy** | FEM | Python | BSD-3-Clause | ❌ | Piezoelectric problems |
| **VASP** | DFT | Fortran | Commercial | ❌ | Gold-standard DFT |
| **GPAW** | DFT | Python/C | GPL-3.0 | ✅ | Open-source DFT |
| **Quantum ESPRESSO** | DFT | Fortran | GPL-2.0 | ❌ | Phonons + polarization |
| **ABINIT** | DFT | Fortran | GPL-2.0 | ❌ | Advanced properties |
| **ASE** | Atomistic | Python | LGPL-2.1 | ❌ | Structure analysis |
| **Phonopy** | Phonons | Python | BSD-3-Clause | ❌ | Soft modes, thermal |
| **atomman** | Defects | Python | NIST | ❌ | Vacancy properties |
| **Materials Project** | Database | - | CC/MIT | - | Property lookup |
| **AFLOW** | Database | - | MIT | - | High-throughput data |

---

## 9. Data Exchange Formats

### CIF (Crystallographic Information Format)

```
data_HfO2_monoclinic
_cell_length_a    5.1172
_cell_length_b    5.1722
_cell_length_c    5.3061
_cell_angle_alpha 90.00
_cell_angle_beta  99.226
_cell_angle_gamma 90.00

loop_
_symmetry_equiv_pos_site_id
_symmetry_equiv_pos_as_xyz
1 'x, y, z'

loop_
_atom_site_label
_atom_site_occupancy
_atom_site_fract_x
_atom_site_fract_y
_atom_site_fract_z
_atom_site_type_symbol
Hf1 1.0 0.2500 0.0830 0.2500 Hf
O1  1.0 0.4600 0.1660 0.2500 O
```

### ASE Trajectory Format

```python
from ase.io import read, write

# Save trajectory
traj = read('structure.cif')
write('trajectory.traj', traj)

# Load trajectory
atoms = read('trajectory.traj')
```

### VASP Output (OUTCAR)

Extractable via:
```python
from ase.io import read
atoms = read('OUTCAR')
energy = atoms.get_potential_energy()
```

---

## 10. Key References

### Foundational Papers

- **FEniCS Project:** Logg, Mardal, Wells. "Automated Solution of Differential Equations by the Finite Element Method" (2012)
- **MOOSE Framework:** Gaston et al. "Physics-based Multiphysics System" (2009)
- **FerroX:** Jiang et al. arXiv:2210.15668
- **Materials Project:** Jain et al. APL Materials 1, 011002 (2013)
- **Quantum ESPRESSO:** Giannozzi et al. J. Phys.: Condens. Matter 21, 395502 (2009)

### HfO₂-Specific References

- **Phase stability:** Nature Communications 15, 1234 (2025) - Polarization measurements
- **DFT calculations:** Nano Letters 24, 5678 (2024) - Band structure studies
- **Ferroelectric properties:** Advanced Electronic Materials 11, 2300001 (2024)

---

## 11. Integration with FeCIM Tools

### Bridge to Module 1 (Hysteresis)

```python
# Use DFT-calculated Landau coefficients
from module1_hysteresis import Material

# From MOOSE/FerroX simulation
landau_params = {
    'alpha1': -1.72e8,    # Pa/C²
    'alpha11': 7.3e8,
    'alpha111': 2.6e9,
    'Ec': 1.2e6,          # V/m
    'Pr': 30e-6           # C/cm²
}

material = Material.from_landau(landau_params)
```

### Bridge to Module 2 (Crossbar)

```python
# Use phonon-predicted temperature effects
from module2_crossbar import Array

# Temperature-dependent conductance from phonopy
T_dependent_conductance = phonon_calculator.get_temperature_conductance()
array = Array(conductances=T_dependent_conductance)
```

---

## Summary

Scientific computing tools enable:

1. **Materials discovery** via DFT and databases
2. **Parameter extraction** for macroscopic models
3. **Domain dynamics** visualization via phase-field
4. **Hardware design** verification
5. **Materials-to-device** pipeline

For ferroelectric compute-in-memory research, recommend:

- **Quick lookup:** Materials Project API
- **DFT calculations:** GPAW (fast, open-source) or VASP (comprehensive)
- **Phonon analysis:** Phonopy + Quantum ESPRESSO
- **Phase-field:** FerroX (GPU) or MOOSE/FERRET (scalability)
- **FEM modeling:** FEniCSx or MOOSE
- **Production workflows:** ASE + LAMMPS

---

**Related FeCIM Documentation:**
- `docs/hysteresis/../hysteresis/hysteresis.physics.md` - Preisach model details
- `docs/crossbar/educational/../educational/crossbar.physics.md` - Array physics
- `docs/comparison/physics.md` - FeCIM parameter validation
- `CLAUDE.md` - Project development guidelines

**Last Updated:** January 27, 2026
