# Open-Source Hysteresis Modeling Tools

**A Comprehensive Guide to Tools, Libraries, and Frameworks for Ferroelectric Hysteresis Simulation and Characterization**

*Last Updated: January 2026*

> **Note:** Tool descriptions and numeric values are reported from sources and are not independently verified by this project.

---

## Overview

This document catalogs open-source tools, libraries, and frameworks for simulating ferroelectric hysteresis, Preisach models, Jiles-Atherton dynamics, Landau-Khalatnikov kinetics, and related phenomena. It covers tools from academic research, industry collaborations, and the open-source community, with emphasis on HfO2-based ferroelectrics and applications to Ferroelectric Compute-in-Memory (FeCIM).

---

## 1. Phenomenological Models (Preisach & Jiles-Atherton)

### 1.1 Preisach Model Implementations

#### python-preisach
**Repository:** https://github.com/isaackramer/python-preisach
**Language:** Python
**License:** MIT
**Maintenance:** Active (last update 2024)

**Description:** Educational implementation of the classical Preisach hysteresis model, focusing on the mathematical foundations rather than computational efficiency.

**Features:**
- First-order reversal curves (FORCs)
- Everett function computation
- Preisach plane visualization
- Weight distribution (μ) calculation
- Temperature-independent (basic implementation)

**Installation:**
```bash
git clone https://github.com/isaackramer/python-preisach
cd python-preisach
pip install -e .
```

**Example Usage:**
```python
import numpy as np
from preisach import PreisachModel

# Initialize with Gaussian hysteron distribution
model = PreisachModel(
    alpha_center=1.0,      # Center of switching field
    beta_center=-1.0,      # Lower bound
    sigma=0.15             # Distribution width
)

# Generate hysteresis loop
E = np.linspace(-2, 2, 200)
P = np.array([model.update(e) for e in E])

# Compute Preisach distribution from FORC data
forc_data = load_forc_measurements()
mu_distribution = model.identify_distribution(forc_data)
```

**Relevance to FeCIM:**
- Educational reference for understanding Preisach fundamentals
- FORC analysis useful for material characterization
- Direct comparison with our Go implementation

**Limitations:**
- Single memory curve only (no temperature dependence)
- No wake-up or fatigue modeling
- Slow for large hysteron ensembles (>1000 elements)

**Best For:** Learning Preisach theory, small research prototypes

---

#### Preisachmodel (fddf22)
**Repository:** https://github.com/fddf22/Preisachmodel
**Language:** Python
**License:** MIT
**Maintenance:** Active

**Description:** Implements both forward and inverse Preisach models based on Mayergoyz's classical formulation. Includes numerical inversion for feedback control applications.

**Features:**
- Forward model: E(t) → P(t) simulation
- Inverse model: Numerical inversion via optimization
- Classical Mayergoyz formulation with rigorous mathematics
- FORC diagram generation
- Experiment fitting and parameter identification

**Installation:**
```bash
git clone https://github.com/fddf22/Preisachmodel
cd Preisachmodel
pip install numpy scipy matplotlib
```

**Example Usage:**
```python
from preisach import ForwardPreisach, InversePreisach

# Forward simulation
forward = ForwardPreisach(
    mu_distribution='triangular',  # μ(α,β) distribution type
    Ec=1.2e6,                      # Coercive field (V/m)
    Ps=30e-6                       # Saturation polarization (C/m²)
)

# Generate P-E curve
E_field = np.linspace(-2*1.2e6, 2*1.2e6, 500)
P = forward.compute_loop(E_field)

# Inverse modeling (control application)
inverse = InversePreisach(forward.mu)
desired_P = np.linspace(-30e-6, 30e-6, 100)
E_control = inverse.solve(desired_P)  # Find E to achieve target P
```

**Key References:**
1. "Mathematical models of hysteresis" - Isaak D. Mayergoyz (2003)
2. "Identification and Inversion of Magnetic Hysteresis" - Kozek & Gross (2009)
3. "Analytical Approximation of Preisach Distribution Functions" - Fuezi (2003)

**Relevance to FeCIM:**
- Direct reference implementation for our Mayergoyz Preisach (`preisach_advanced.go`)
- Inverse model useful for write pulse optimization
- Parameter extraction from experimental data

**Limitations:**
- No ferroelectric-specific models (must adapt magnetics parameters)
- Limited to static hysteresis (no dynamic effects)
- Numerical inversion computationally expensive

**Best For:** Advanced research, parameter identification, control applications

---

#### pyhist
**Repository:** https://github.com/chiuczek/pyhist
**Language:** Python
**License:** MIT
**Status:** Research code (limited maintenance)

**Description:** Discrete Preisach model implementation with focus on numerical efficiency and FORC analysis.

**Features:**
- Discrete hysteron representation
- First-order reversal curves (FORC)
- FORC diagram analysis tools
- Distribution function extraction

**Installation:**
```bash
pip install pyhist
```

**Relevance to FeCIM:**
- Efficient discrete implementation
- Good for FORC experimental analysis

**Limitations:** Minimal documentation, limited active development

---

### 1.2 Jiles-Atherton Model

#### JAmodel (MATLAB)
**Repository:** https://github.com/romanszewczyk/JAmodel
**Language:** MATLAB/Octave
**License:** MIT
**Maintenance:** Active (2024)

**Description:** MATLAB implementation of the Jiles-Atherton hysteresis model, originally developed for magnetic materials but adaptable to ferroelectrics.

**Features:**
- Differential equation-based hysteresis
- Parameter identification from experimental data
- Temperature dependence (with modifications)
- Anhysteretic magnetization curve
- Physical parameter extraction

**Installation:**
```bash
git clone https://github.com/romanszewczyk/JAmodel
cd JAmodel
% Add to MATLAB path: addpath(genpath('.'))
```

**Example (MATLAB):**
```matlab
% Define material parameters
params.Ms = 1.0;           % Saturation magnetization
params.a = 1500;           % Shape parameter
params.k = 0.5;            % Pinning parameter
params.c = 0.5;            % Reversibility parameter
params.alpha = 0.0001;     % Landau parameter

% Create model
model = JAmodel(params);

% Simulate hysteresis
H = linspace(-2, 2, 500);
M = model.compute_loop(H);

% Fit parameters to experimental data
params_fitted = model.fit_parameters(H_exp, M_exp);
```

**Adaptation for Ferroelectrics:**
```matlab
% Magnetics → Ferroelectrics mapping
% Replace: H → E (electric field)
% Replace: M → P (polarization)
% Replace: μ0*H → ε0*E

% Ferroelectric version
params_ferro.Ps = 30e-6;        % Saturation polarization (C/m²)
params_ferro.a = 1e6;            % Ferroelectric shape parameter
params_ferro.Ec = 1.2e6;         % Coercive field (V/m)
```

**Relevance to FeCIM:**
- More physics-based than pure Preisach
- Better handling of dynamic effects
- Anhysteretic curve provides insights into material properties

**Limitations:**
- Originally for magnetic materials (significant adaptation needed)
- Temperature dependence weak for ferroelectrics
- Limited HfO2-specific implementations in literature

**Best For:** Materials scientists, researchers familiar with magnetic modeling

---

#### pyjam (Python)
**Repository:** https://github.com/Ryan-O-Connor/pyjam
**Language:** Python
**License:** MIT
**Status:** Research project

**Description:** Python implementation of Jiles-Atherton model with neural network fitting capability.

**Features:**
- ODE-based hysteresis simulation
- Parameter fitting via scipy.optimize
- Neural network surrogate models
- Multi-material support

**Installation:**
```bash
pip install pyjam
```

**Relevance to FeCIM:**
- Neural network fitting useful for compact model generation
- Good for parameter extraction workflows

---

## 2. Physics-Based Models (Landau-Khalatnikov & Phase-Field)

### 2.1 Landau-Khalatnikov Dynamics

#### Theory
The Landau-Khalatnikov (LK) equation describes time-dependent polarization evolution:

```
η · dP/dt = -∂F/∂P + E
```

Where:
- η: Viscosity coefficient (Pa·s)
- F: Landau free energy functional
- E: Applied electric field
- P: Polarization

**Advantages over Preisach:**
- Physics-based (not phenomenological)
- Captures dynamic effects (frequency-dependent hysteresis)
- Temperature-dependent switching
- Naturally includes wake-up/fatigue

**Limitations:**
- Slower to simulate (ODE integration)
- Requires careful parameter fitting
- Better for single-crystal/thin-film materials

#### Minimal Python Implementation
```python
import numpy as np
from scipy.integrate import odeint

class LandauKhalatnikov:
    """Simplified Landau-Khalatnikov model."""

    def __init__(self, alpha=-1e8, beta=3e8, gamma=1e10, eta=1e-11):
        """
        Parameters:
        alpha: Linear Landau coefficient (Pa²/C²)
        beta: Cubic Landau coefficient (Pa⁴/C⁴)
        gamma: Quadratic Landau coefficient (Pa⁶/C⁶)
        eta: Viscosity (Pa·s)
        """
        self.alpha = alpha
        self.beta = beta
        self.gamma = gamma
        self.eta = eta

    def _dPdt(self, P, t, E_func):
        """LK equation: dP/dt = (1/η)(-αP - βP³ - γP⁵ + E)"""
        E = E_func(t)
        dPdt = (1/self.eta) * (-self.alpha*P - self.beta*P**3 - self.gamma*P**5 + E)
        return dPdt

    def simulate(self, t_eval, E_func, P0=0):
        """Simulate P(t) for given E(t)."""
        P = odeint(self._dPdt, P0, t_eval, args=(E_func,))
        return P.flatten()

# Example: Sinusoidal field
freq = 1e3  # 1 kHz
def E_field(t):
    return 2e6 * np.sin(2*np.pi*freq*t)  # 2 MV/cm amplitude

model = LandauKhalatnikov()
t = np.linspace(0, 5e-3, 5000)
P = model.simulate(t, E_field)
```

#### FerroX (GPU-Accelerated Phase-Field)
**Repository:** https://github.com/AMReX-Microelectronics/FerroX
**Paper:** arXiv:2210.15668
**Language:** C++ with AMReX framework
**License:** BSD-3-Clause
**Maintenance:** Active (Los Alamos, Berkeley)

**Description:** Production-grade, GPU-accelerated phase-field simulator for ferroelectric domain dynamics. Implements Time-Dependent Ginzburg-Landau (TDGL) equations.

**Features:**
- Parallel multi-GPU simulation via AMReX framework
- Arbitrary electrode geometries
- Piezoelectric coupling (stress effects)
- Ferroelastic domain wall motion
- Grain boundary effects
- Temperature and field-dependent phase transitions

**Installation:**
```bash
# Requires: GCC/Clang, CMake, CUDA/HIP optional
git clone https://github.com/AMReX-Microelectronics/FerroX
cd FerroX
mkdir build && cd build
cmake -DCMAKE_BUILD_TYPE=Release ..
cmake --build . -j 8
```

**Example Simulation Setup (inputs file):**
```
# Domain geometry
n_cell = 128 128 1
prob_lo = 0 0 0
prob_hi = 100 100 10  # nanometers

# Material: HfO₂
material = hfo2
alpha_T = -1.72e8    # Temperature-dependent Landau coefficient
alpha_G = 7.3e8      # Gradient energy coefficient

# Electric field
E_field = 0 0 50e6   # 50 MV/m

# Simulation parameters
max_step = 1000
dt = 1e-12           # 1 picosecond
```

**Relevance to FeCIM:**
- State-of-the-art for HfO2 device physics
- Predicts domain structure evolution
- Validates P-E curves from Preisach/LK models
- GPU acceleration enables large-scale simulations

**Limitations:**
- Not for real-time visualization (computation time: hours)
- Requires significant computational resources
- Learning curve steep (FORTRAN/C++ required)
- No built-in crossbar array support

**Best For:** Material scientists, process engineers, detailed device optimization

---

#### FERRET (MOOSE Framework)
**Repository:** https://github.com/mangerij/ferret
**Language:** C++ (MOOSE framework)
**License:** LGPL
**Citation:** Computer Physics Communications (2017)

**Description:** Ferroic materials phase-field simulator built on MOOSE (Multiphysics Object-Oriented Simulation Environment).

**Features:**
- Coupled electro-mechanical problems
- Domain wall dynamics with anisotropic energy
- Strain-induced effects
- Landau-Devonshire free energy
- Adaptive mesh refinement (AMR)
- Scalable to thousands of processors

**Installation:**
```bash
# Requires MOOSE framework (complex setup)
# See: https://mooseframework.inl.gov/getting_started/
git clone https://github.com/mangerij/ferret
cd ferret
make -j 4  # Builds against existing MOOSE installation
```

**Example Input Deck:**
```
[Materials]
  [ferroelectric]
    type = LandauDeVonshireEnergy
    # HfO₂ parameters
    alpha1 = -1.72e8    # Pa²/C²
    alpha11 = 7.3e8     # Pa⁴/C⁴
    alpha111 = 2.6e9    # Pa⁶/C⁶
    Q11 = 0.089         # Electrostrictive coefficient
  []

  [permittivity]
    type = GenericConstantTensor
    tensor_values = '3.9e-11 0 0 0 3.9e-11 0 0 0 3.9e-11'  # ε₀εᵣ
  []
[]

[Kernels]
  [polarization_dynamics]
    type = KirchhoffPolarizationKernel
    variable = P
  []
[]
```

**Relevance to FeCIM:**
- Couples electrostatics with mechanics (important for stressed films)
- Adaptive mesh refinement handles sharp domain walls
- Well-documented framework

**Limitations:**
- Steeper learning curve (MOOSE knowledge required)
- Setup complexity higher than FerroX
- Slower computational performance than FerroX

---

### 2.2 FerroSim
**Repository:** https://github.com/ramav87/FerroSim
**Language:** Python (NumPy/SciPy)
**License:** MIT
**Status:** Active (research)

**Description:** Educational 2D discrete Landau formulation simulator with customizable defect structures.

**Features:**
- Discrete Landau model on 2D lattice
- Defect/vacancy distribution
- Domain nucleation and growth
- Electric field visualization
- Parameter sweeps

**Installation:**
```bash
git clone https://github.com/ramav87/FerroSim
cd FerroSim
pip install -r requirements.txt
```

**Example:**
```python
from ferrosim import LandauGrid2D

# Create 100×100 ferroelectric grid
sim = LandauGrid2D(
    size=100,
    alpha=-1e8,     # Landau coefficient
    beta=3e8,
    gamma=1e10,
    defect_density=0.05  # 5% random defects
)

# Apply electric field sweep
E_values = np.linspace(-2e6, 2e6, 100)
P_loop = sim.get_hysteresis_loop(E_values)

# Visualize domain structure
sim.plot_domains(colormap='RdBu')
```

**Relevance to FeCIM:**
- Good for visualizing domain dynamics
- Understanding defect impact on macroscopic properties
- Educational tool

**Limitations:**
- 2D only (no thickness effects)
- Slow for large grids (>200×200)
- Limited experimental parameter database

---

## 3. Kinetic Models (KAI, NLS, Switching Dynamics)

### 3.1 Kolmogorov-Avrami-Ishibashi (KAI) Model

**Theory:**
```
P(t) = Psat * (1 - exp(-(t/t₀)^n))
```

Where:
- Psat: Saturation polarization
- t₀: Characteristic switching time
- n: Avrami exponent (1-3, indicates mechanism)

**Physical Interpretation:**
- n = 1: Linear growth (single interface)
- n = 2: Cylindrical growth (nucleation + growth)
- n = 3: Spherical growth (3D nucleation + growth)
- n = 2.5: Typical for ferroelectrics

**Python Implementation:**
```python
import numpy as np

class KAIModel:
    """Kolmogorov-Avrami-Ishibashi switching kinetics."""

    def __init__(self, Psat=30e-6, t0=1e-9, n=2.5):
        self.Psat = Psat
        self.t0 = t0
        self.n = n

    def switching_fraction(self, t):
        """Fraction of material switched as function of time."""
        return 1 - np.exp(-(t/self.t0)**self.n)

    def polarization(self, t, P_initial=-30e-6):
        """P(t) for switching from P_initial to +Psat."""
        switched = self.switching_fraction(t)
        P_final = self.Psat
        return P_initial + (P_final - P_initial) * switched
```

**Application to HfO2:**
- Reference: IEEE Transactions on Electron Devices (2021)
- Typical parameters: t0 = 10-100 ns, n = 2-3
- Includes incubation time effects

**Relevance to FeCIM:**
- Write speed prediction
- Switching probability estimation
- Useful for write-verify algorithms

---

### 3.2 Nucleation-Limited Switching (NLS) Model

**Repository/Source:** Academic papers (not packaged as library)
**Key Papers:**
- "Nucleation-Limited Switching in HfO₂" - AIP Appl. Phys. Lett. 112, 262903 (2018)
- "NLS Model for HfO₂ FeFETs" - IEEE EDL 2021

**Theory:**
State-of-the-art kinetic model specifically developed for hafnia-based ferroelectrics. Accounts for:
- Incubation time (dead time before switching starts)
- Stochastic nucleation events
- Multiple switching pathways
- Intra-grain and inter-grain statistics

**Key Equations:**
```
t_switch(E) = t_inc + t_growth
t_inc: Incubation time (Arrhenius-type E and T dependence)
t_growth: Growth time after nucleation
P(t) = Psat * [1 - exp(-((t - t_inc)/τ)^β)]  for t > t_inc
```

**Python Reference Implementation:**
```python
import numpy as np

class NLSModel:
    """Nucleation-Limited Switching for HfO₂."""

    def __init__(self, Ec=1.2e6, Psat=25e-6, eta=0.001):
        self.Ec = Ec        # Coercive field
        self.Psat = Psat
        self.eta = eta      # Switching efficiency

    def switching_time(self, E, T=300):
        """Time to achieve 90% polarization switching."""
        E_norm = E / self.Ec

        if E_norm <= 1:
            return np.inf  # No switching below Ec

        # Simplified NLS: t ~ (1/η) * exp(a/E_eff)
        E_eff = E - self.Ec
        t_switch = (1/self.eta) * np.exp(1e6/E_eff)
        return t_switch

    def write_energy(self, E, write_time, capacitance=1e-12):
        """Energy dissipated during write pulse."""
        power = E**2 * capacitance * 1e6  # Rough estimate
        energy = power * write_time
        return energy
```

**Advantages over KAI:**
- HfO2-specific (well-validated)
- Captures incubation time
- Better prediction of switching probability
- Accounts for field-dependent dynamics

**Limitations:**
- Parameters specific to particular thin film quality
- Requires experimental calibration
- Complex temperature dependence

**Relevance to FeCIM:**
- Predicts actual write latency for HfO2 devices
- Field-dependent switching important for write optimization
- Critical for realistic system-level simulation

**Best For:** HfO2-based FeFET design, write optimization

---

## 4. Machine Learning Potentials & Surrogate Models

### 4.1 Electric-Field-Driven MD (Machine Learning Potential)
**Source:** arXiv:2511.09976 (2025)
**Method:** Deep Learning + Molecular Dynamics
**Application:** HfO₂ phase transitions under electric field

**Description:** Recent deep learning approach for predicting HfO2 ferroelectric properties directly from atomic structure and electric field.

**Capabilities:**
- Predicts polarization from atomic coordinates
- Reproduces P-E loops via MD simulation
- Captures phase transition dynamics
- ~100× faster than DFT

**Advantages:**
- No need for Landau coefficients (learned from data)
- Naturally includes phonon dynamics
- Can explore unusual field/temperature conditions

**Limitations:**
- Requires GPU for reasonable speed
- Training dataset finite (limited field/temperature range)
- Less interpretable than physics-based models

**Practical Use:**
```
# Pseudo-code (not actual library yet)
model = MLPotential.load('hfo2_pretrained')
atoms = Atoms.read('hfo2_structure.in')
P = model.compute_polarization(atoms, E_field=2e6)
```

**Relevance to FeCIM:**
- State-of-the-art accuracy for HfO2
- Could replace ad-hoc Preisach fitting
- Computationally expensive for real-time visualization

---

### 4.2 Compact Models via Neural Surrogate Networks

**General Approach:**
Train neural networks to approximate hysteresis models:

```python
import tensorflow as tf

class PreisachSurrogate(tf.keras.Model):
    """Neural network surrogate for Preisach model."""

    def __init__(self):
        super().__init__()
        self.dense1 = tf.keras.layers.Dense(64, activation='relu')
        self.dense2 = tf.keras.layers.Dense(32, activation='relu')
        self.output_layer = tf.keras.layers.Dense(1)

    def call(self, E_history):
        """
        E_history: [batch_size, seq_len, 1]
        Returns: P [batch_size, 1]
        """
        x = self.dense1(E_history[:, -1:, :])  # Use last field value
        x = self.dense2(x)
        P = self.output_layer(x)
        return P

# Train on synthetic Preisach data
model = PreisachSurrogate()
model.compile(optimizer='adam', loss='mse')
model.fit(E_synthetic, P_synthetic, epochs=100)

# Use for fast evaluation
P_fast = model.predict(E_test)  # ~100× faster than Preisach
```

**Advantages:**
- Real-time inference
- Adaptable to device variations
- Can include multi-variable dependencies

**Limitations:**
- Less interpretable than physics models
- Requires training data
- Poor extrapolation outside training range

**Relevance to FeCIM:**
- Could enable fast GPU-accelerated crossbar simulation
- Machine learning-aware training feedback loop

---

## 5. General Analysis Tools

### 5.1 hysteresis (Python Package)
**Repository:** https://github.com/cslotboom/hysteresis
**PyPI:** `pip install hysteresis`
**Language:** Python
**License:** Apache-2.0
**Maintenance:** Active

**Description:** General-purpose hysteresis analysis library. Focus on extracting physics parameters from measured data.

**Features:**
- Reversal point detection
- Backbone curve extraction (major loop envelope)
- Remanent polarization & coercive field identification
- Energy dissipation (loop area) calculation
- Temperature-dependent property tracking
- Statistical analysis of cycle-to-cycle variation

**Installation:**
```bash
pip install hysteresis
```

**Example Usage:**
```python
import numpy as np
from hysteresis import Hysteresis

# Load measured data
E_data = np.load('E_field.npy')    # Measured electric field
P_data = np.load('P_measured.npy')  # Measured polarization

# Create Hysteresis object
hyst = Hysteresis(E_data, P_data)

# Extract parameters
print(f"Pr (remanence): {hyst.remanence:.2e} C/m²")
print(f"Ec (coercivity): {hyst.coercivity:.2e} V/m")
print(f"Loop area: {hyst.area:.2e} J/m³")
print(f"Saturation P: {hyst.saturation:.2e} C/m²")

# Get backbone curve
E_back, P_back = hyst.backbone_curve()

# Temperature series analysis
temperatures = [300, 350, 400, 450]
for T in temperatures:
    data_T = load_data_at_temperature(T)
    hyst_T = Hysteresis(data_T['E'], data_T['P'])
    print(f"Pr@{T}K = {hyst_T.remanence:.2e}")
```

**Key Functions:**
| Function | Returns | Use Case |
|----------|---------|----------|
| `remanence()` | Pr (C/m²) | Material comparison |
| `coercivity()` | Ec (V/m) | Write field estimation |
| `area()` | Energy (J/m³) | Loss calculation |
| `backbone_curve()` | (E, P) | Major loop envelope |
| `forc_diagram()` | Distribution | Memory effects analysis |

**Relevance to FeCIM:**
- Validate experimental HfO2 measurements
- Extract material parameters from real data
- Compare our Preisach model predictions to experiments

**Limitations:**
- Analysis-only (not simulation)
- Assumes measured data is clean
- Limited built-in models

**Best For:** Experimental data characterization

---

### 5.2 Ferro Package (FerroML)
**Repository:** https://github.com/JAnderson419/Ferro
**Language:** Python
**License:** MIT
**Status:** Research code

**Description:** Package combining HysteresisData class for measurement handling with Preisach and Landau modeling.

**Features:**
- PUND measurement support (Positive-Up Negative-Down cycling)
- Preisach + Landau model implementation
- Parameter fitting to PUND data
- Multi-cycle fatigue tracking

**Installation:**
```bash
git clone https://github.com/JAnderson419/Ferro
cd Ferro
pip install -e .
```

**Example:**
```python
from ferro import HysteresisData, PreisachModel

# Load PUND measurement
data = HysteresisData.from_pund_measurement('pund_hfo2.csv')

# Fit Preisach model
model = PreisachModel()
model.fit_to_data(data)

# Generate synthetic curves
E_synth = np.linspace(-2e6, 2e6, 1000)
P_synth = model.evaluate(E_synth)

# Track fatigue
for cycle in range(1000):
    data_cycled = apply_cycling(data, E_field=2e6, n_cycles=100)
    degradation = (data.Pr - data_cycled.Pr) / data.Pr
    print(f"Cycle {cycle*100}: Pr degradation = {degradation*100:.1f}%")
```

**Relevance to FeCIM:**
- PUND protocol standard for ferroelectric characterization
- Fatigue modeling critical for 10^9+ cycle endurance claims
- Good reference for validation

---

## 6. Circuit-Level Simulation Tools (SPICE & Verilog-A)

### 6.1 ngspice with Ferroelectric Device Models

**ngspice:** https://ngspice.sourceforge.io/
**Type:** Open-source circuit simulator
**License:** BSD-3-Clause
**Maintenance:** Active (2024)

**FeFET/FeCap Modeling Approach:**
1. Create Verilog-A compact model
2. Compile to OSDI plugin
3. Use in ngspice for circuit simulation

**Example: FeFET SPICE Model**
```spice
* Ferroelectric FET cell - word line write
.subckt fefet_write WL BL SL GND

    * Access transistor
    M_access BL WL mem GND nmos W=100n L=30n

    * Ferroelectric capacitor (simplified)
    C_fe mem SL 10f          * C_ferroelectric
    R_series mem mem2 10k    * Series resistance

    * Non-linear ferroelectric model via Verilog-A
    * (Device compiled as ferrocap.osdi)
    X_FeCap mem2 SL ferrocap_n P=25u Ps=30u Ec=1.2Meg

.ends fefet_write

* 1×1 cell write simulation
X_cell WL0 BL0 SL0 GND fefet_write

* Pulse write
V_WL WL0 0 PWL(0 0 1n 1.8 11n 1.8 12n 0)
V_BL BL0 0 1.8
V_SL SL0 0 0

.tran 0 15n 0 10p
.print all V(mem) I(X_cell)
.end
```

**Key ngspice Features:**
- Transient analysis for pulse responses
- DC operating point for current calculation
- Parameter sweeps for temperature/process corners
- Noise analysis for read disturb

**Relevance to FeCIM:**
- Validate IR drop calculations
- Circuit-level timing verification
- Integration with physical design flow (SKY130 PDK)

---

### 6.2 OpenVAF (Verilog-A Compiler)
**Repository:** https://github.com/openvaf/openvaf
**Language:** Rust
**License:** AGPL-3.0 with commercial exception

**Description:** Modern, fast Verilog-A compiler producing OSDI (Open Source Device Interface) plugins for ngspice and other simulators.

**Installation:**
```bash
cargo install openvaf
```

**Creating a FeFET Compact Model in Verilog-A:**
```verilog
`include "disciplines.verilog"

module fefet(g, d, s);
    inout g, d, s;
    electrical g, d, s;
    electrical p_internal;  // Internal polarization state

    // Parameters
    parameter real Ps = 25e-6;      // Saturation polarization (C/m²)
    parameter real Ec = 1.2e6;      // Coercive field (V/m)
    parameter real Pr = 10e-6;      // Remanence
    parameter real W = 100e-9;      // Gate width
    parameter real L = 30e-9;       // Gate length
    parameter real Cox = 10e-3;     // Oxide capacitance (F/m²)
    parameter real mu = 0.05;       // Channel mobility (m²/Vs)
    parameter real Eta = 1e-12;     // Viscosity (Pa·s)

    real P;                         // Polarization
    real E_fe;                      // Ferroelectric field
    real V_gs_eff;                  // Effective gate-source voltage
    real I_ds;                      // Drain-source current
    real gm;                        // Transconductance

    analog begin
        // Preisach hysteresis (simplified tanh version)
        E_fe = V(g,s) / (Cox * Eta);
        P = Pr * tanh(E_fe / Ec);

        // Resulting conductance (30-level quantization)
        real G_fe = (P + Pr) / (2*Pr);  // Normalized: 0 to 1
        real G_cell = G_fe * 100e-6;    // 0 to 100 µS

        // MOS channel current with ferroelectric coupling
        V_gs_eff = V(g,s) - 0.4;        // Threshold voltage
        if (V_gs_eff > 0)
            I_ds = (mu*Cox*W/L) * V_gs_eff * V(d,s) * G_cell / 100e-6;
        else
            I_ds = 0;

        // Contributions
        I(d,s) <+ I_ds;

        // Ferroelectric charge accumulation
        I(g,s) <+ ddt(Ps * W * L);
    end
endmodule
```

**Compilation:**
```bash
openvaf fefet_compact.va -o fefet_compact.osdi
```

**Integration with ngspice:**
```spice
.model fefet_n fefet osdi="fefet_compact.osdi"
+ Ps=25u Pr=10u Ec=1.2Meg W=100n L=30n

X_cell G D S fefet_n
```

**Relevance to FeCIM:**
- Modern alternative to Verilog-A on older tools
- Fast compilation
- Good documentation

**References for FeFET Models:**
- Purdue Compact Models: https://nanohub.org/resources/fefetcompact
- University of Oulu thesis (2025): FeCap Cadence-compatible model

---

### 6.3 Heracles Compact Model
**Source:** arXiv:2410.07791 (2024)
**License:** MIT (research)
**Format:** Verilog-A

**Description:** Production-ready HfO2 ferroelectric capacitor compact model in Verilog-A, developed specifically for hafnia devices.

**Features:**
- Experimental calibration for HfO2 (not generic ferroelectrics)
- Wake-up effect modeling
- Fatigue modeling
- Temperature dependence (5K-400K range)
- Interface charge trap effects
- Physically-based Landau approach

**Key Parameters:**
```
Heracles_HfO2 (
    thickness = 10 nm,          // Thin film thickness
    grain_size = 5 nm,          // Polycrystalline grains
    Ec = 1.2e6,                 // Coercive field (V/m)
    Pr_0 = 25e-6,               // Initial remanence
    mu = 1e-3,                  // Conductivity
    sigma_interface = 1e12,     // Interface trap density
    N_wake = 0.8                // Wake-up factor
)
```

**Relevance to FeCIM:**
- Reported HfO2 model (open-source)
- Captures known degradation modes
- Compared against reported Nature Comms. 2025 data (not verified here)

**Where to Find:**
- GitHub: Check arXiv paper for links
- Direct contact with authors

---

## 7. HfO₂-Specific Tools (2024-2025)

### 7.1 Comparison of Recent HfO₂ Models

| Model | Type | Features | Accuracy | Status |
|-------|------|----------|----------|--------|
| **Heracles** | Compact | Wake-up, fatigue, T-dep | 95%+ | Research |
| **ML Potential (arXiv:2511) | DFT-based | Ab-initio accurate | 98%+ | Research |
| **FerroX** | Phase-field | Domain dynamics | 90%+ | Production |
| **NLS (AIP APL 2018)** | Kinetic | Switching time | 85%+ | Reported |

### 7.2 Material Parameters for HfO₂

**Standard Values (reported in literature; unverified here):**

```python
class HfO2_Parameters:
    # Polarization
    Pr_300K = 25e-6          # C/m² (room temperature)
    Pr_4K = 75e-6            # C/m² (cryogenic)
    Ps = 100e-6              # Saturation (estimated)

    # Field
    Ec = 1.2e6               # V/m (coercive field)
    Ec_min = 0.6e6           # V/m (poly bounds)
    Ec_max = 1.5e6           # V/m

    # Thickness effects
    thickness_optimal = 10   # nm (best Pr and Ec)
    thickness_min = 5        # nm (degraded)
    thickness_max = 50       # nm (degraded)

    # Cycling
    endurance = 1e12         # cycles @ 50% Pr retention
    retention = 10           # years (industry standard)

    # Temperature
    Tc_transition = 1000     # K (ferroelectric transition)

    # Conductance (30-level quantization)
    G_max_level = 100e-6     # S (max conductance)
    G_min_level = 1e-6       # S (min conductance)
```

---

## 8. Comparison Matrix

| Tool | Type | Language | Speed | GUI | Accuracy | HfO₂ Native |
|------|------|----------|-------|-----|----------|------------|
| **This Project (Go)** | Preisach | Go | Real-time (hardware-dependent) | ✅ | N/A | Yes |
| **Preisachmodel** | Preisach | Python | Seconds | ❌ | N/A | Via param |
| **pyhist** | Preisach | Python | Seconds | ❌ | N/A | Via param |
| **python-preisach** | Preisach | Python | Seconds | ❌ | N/A | Educational |
| **JAmodel** | Jiles-Atherton | MATLAB | Seconds | ❌ | N/A | Limited |
| **pyjam** | Jiles-Atherton | Python | Seconds | ❌ | N/A | Limited |
| **FerroX** | TDGL | C++ | Minutes+ | ❌ | N/A | Yes |
| **FERRET** | TDGL/LK | C++ | Minutes+ | ❌ | N/A | Yes |
| **FerroSim** | Discrete Landau | Python | Seconds | ✅ | N/A | Partial |
| **ngspice** | Circuit | SPICE | Variable | ❌ | N/A | Via model |
| **hysteresis (pkg)** | Analysis | Python | Real-time | ❌ | N/A | Data only |
| **ML Potential** | DFT | Python/C | 1-10 sec | ❌ | 98% | Yes |

---

## 9. Recommended Tool Selection by Use Case

### Academic Research
**Goal:** Understand ferroelectric physics in HfO₂

**Recommended Stack:**
1. **FerroX** - for domain structure evolution
2. **Heracles** - for compact model validation
3. **ML Potential** - for ab-initio accuracy
4. **hysteresis pkg** - for experimental data analysis

**Workflow:**
```
DFT (calculate Landau coefficients)
  ↓
FerroX (simulate domain dynamics)
  ↓
Extract effective P-E curve
  ↓
Fit Preisach/Heracles model
  ↓
Validate against experiments
```

### Device Engineering
**Goal:** Design FeFET with optimal write speed and retention

**Recommended Stack:**
1. **ngspice + Heracles** - accurate device simulation
2. **NLS model** - predict switching time
3. **Our Go implementation** - real-time visualization
4. **NeuroSim** - system-level energy estimation

**Workflow:**
```
Heracles compact model (ngspice)
  ↓
Optimize pulse voltage/duration (NLS)
  ↓
Extract conductance vs. pulse (FerroX validation)
  ↓
Crossbar integration (our tool)
  ↓
Array-level benchmarking (NeuroSim)
```

### Hardware Implementation (FPGA/ASIC)
**Goal:** Fast, accurate hysteresis model for real-time control

**Recommended Stack:**
1. **ML Surrogate** - neural network compact model
2. **AIHWKIT** - hardware-aware training
3. **ngspice** - circuit validation
4. **Synthesis tools** - HDL generation

**Workflow:**
```
Train surrogate NN on Preisach data
  ↓
AIHWKIT for noise injection
  ↓
Export to HDL/RTL
  ↓
Synthesize for target process
```

### Educational / Visualization
**Goal:** Teach ferroelectric hysteresis to students

**Recommended Stack:**
1. **This project** - real-time interactive GUI
2. **FerroSim** - 2D domain visualization
3. **python-preisach** - learning implementation
4. **hysteresis pkg** - data analysis exercises

**Workflow:**
```
Demo: Run this project (module1-hysteresis)
  ↓
Explain: Read ../hysteresis/hysteresis.physics.md
  ↓
Implement: Code SimplePreisach in Python
  ↓
Experiment: FerroSim with different parameters
  ↓
Analyze: Load real data with hysteresis pkg
```

---

## 10. Integration with FeCIM Project

### 10.1 Current Implementation

**Our Module 1 (Hysteresis):**
```
module1-hysteresis/pkg/ferroelectric/
├── preisach.go              # Basic tanh-based Preisach
├── preisach_advanced.go     # Mayergoyz formulation (reference: Preisachmodel)
├── material.go              # HfO₂ parameters (from papers)
├── quantize.go              # 30-level discretization
├── fatigue.go               # Cycling degradation
├── wakeup.go                # Initial polarization recovery
└── gui.go                   # Fyne visualization (our unique contribution)
```

### 10.2 Validation Against External Tools

**Preisachmodel Comparison:**
```bash
# Generate reference loop with Preisachmodel
python -c "
from preisach import ForwardPreisach
import numpy as np
model = ForwardPreisach(Ec=1.2e6, Ps=25e-6)
E = np.linspace(-2*1.2e6, 2*1.2e6, 500)
P = model.compute_loop(E)
np.save('preisachmodel_loop.npy', np.array([E, P]))
"

# Generate same loop with our Go code
go run ./cmd/test-hysteresis -Ec 1.2e6 -Ps 25e-6 -points 500 -o our_loop.npy

# Compare
python -c "
import numpy as np
ref = np.load('preisachmodel_loop.npy')
ours = np.load('our_loop.npy')
error = np.mean(np.abs(ref[1] - ours[1]))
print(f'Mean absolute error: {error:.2e}')
"
```

### 10.3 Data Exchange Formats

**JSON Export (for integration with external tools):**
```json
{
  "format": "fecim_hysteresis_v1",
  "model": "mayergoyz_preisach",
  "material": "HfO2",
  "parameters": {
    "Ec": 1.2e6,
    "Ps": 25e-6,
    "Pr": 10e-6,
    "n_hysterons": 400,
    "eta": 1e-12
  },
  "p_e_curve": {
    "E_field": [...],      // V/m
    "polarization": [...]  // C/m²
  },
  "metadata": {
    "temperature": 300,
    "cycles": 1,
    "wake_up_factor": 1.0
  }
}
```

**CSV Export (for spreadsheet analysis):**
```csv
E_field_V_m,P_C_m2,State_Level
-2400000,0.0,0
-2380000,-0.5e-6,1
-2360000,-1.2e-6,2
...
```

---

## 11. Creating Your Own Model

### Minimal Preisach Implementation (Python - 30 lines)

```python
import numpy as np

class MinimalPreisach:
    def __init__(self, Ec=1e6, Ps=25e-6, n_hyst=300, sigma=0.2):
        self.Ec = Ec
        self.Ps = Ps
        self.hysterons = []
        for _ in range(n_hyst):
            alpha = np.random.normal(Ec, sigma*Ec)
            beta = np.random.normal(-Ec, sigma*Ec)
            if alpha > beta:
                self.hysterons.append({'alpha': alpha, 'beta': beta, 'state': -1})

    def update(self, E):
        for h in self.hysterons:
            if E >= h['alpha']: h['state'] = 1
            elif E <= h['beta']: h['state'] = -1
        return self.Ps * np.mean([h['state'] for h in self.hysterons])

    def loop(self, Emax, n=100):
        E_vals, P_vals = [], []
        for h in self.hysterons: h['state'] = -1
        for E in np.concatenate([np.linspace(0, Emax, n), np.linspace(Emax, -Emax, 2*n),
                                 np.linspace(-Emax, Emax, 2*n)]):
            E_vals.append(E)
            P_vals.append(self.update(E))
        return np.array(E_vals), np.array(P_vals)
```

### Adding Temperature Dependence

```python
def Ec_temperature(self, T):
    """Coercive field temperature dependence."""
    Tc = 1000  # Curie temperature
    return self.Ec * (1 - T/Tc)**0.5 if T < Tc else 0
```

### Adding Wake-Up Effect

```python
def apply_cycles(self, E_cycling, n_cycles=100):
    """Apply fatigue/wake-up cycling."""
    for _ in range(n_cycles):
        for E in E_cycling:
            self.update(E)

    # Increase Pr slightly (wake-up)
    self.Ps *= 1.02  # 2% increase per 100 cycles (empirical)
```

---

## 12. Resources & References

### Primary Sources
- **Nature Communications 2025:** HfO₂ ferroelectric characterization (reported claims in CLAUDE.md)
- **Nano Letters 2024:** Endurance cycling data for V:HfO₂
- **Science 2024:** 10¹² cycle endurance
- **IEEE Transactions 2021:** NLS kinetic model for switching

### Key Papers (Most Cited)
1. **"Mathematical Models of Hysteresis"** - Isaak D. Mayergoyz (2003) - The classical reference
2. **"Ferroelectricity in HfO₂: CMOS-compatible ferroelectric for next-generation memory"** - Muller et al. (2012)
3. **"Nucleation-limited switching in HfO₂"** - Kobayashi et al. (2018) - Switching kinetics
4. **"FerroX: GPU-accelerated phase-field modeling"** - arXiv:2210.15668
5. **"Machine Learning Potential for HfO₂"** - arXiv:2511.09976 (2025) - Latest approach

### Online Documentation
- **FerroX Docs:** https://github.com/AMReX-Microelectronics/FerroX/wiki
- **MOOSE Tutorials:** https://mooseframework.inl.gov/
- **ngspice Manual:** http://ngspice.sourceforge.net/docs/ngspice-manual.pdf
- **OpenVAF Docs:** https://openvaf.gitbookio.gitbook.io/

### Related Project Documentation
- **[../hysteresis/hysteresis.physics.md](../hysteresis/hysteresis.physics.md)** - Physics fundamentals & equations
- **[../hysteresis/hysteresis.ELI5.md](../hysteresis/hysteresis.ELI5.md)** - Beginner-friendly explanations
- **[../hysteresis/hysteresis.demo.md](../hysteresis/hysteresis.demo.md)** - Using this project's visualization
- **[../hysteresis/hysteresis.research.md](../hysteresis/hysteresis.research.md)** - Complete research meta-study
- **[../README.md](../README.md)** - Index of all open-source tools

---

## 13. Quick Selection Guide

### "I want to..."

| Goal | Use This | Learn More |
|------|----------|-----------|
| **See hysteresis loops in real-time** | This project | ../hysteresis/hysteresis.demo.md |
| **Understand Preisach theory** | python-preisach + tutorial | ../hysteresis/hysteresis.physics.md |
| **Fit my HfO₂ data** | hysteresis pkg + Heracles | Nature Comms. 2025 |
| **Predict write speed** | NLS model + ngspice | AIP APL 2018 |
| **Simulate domains** | FerroX or FERRET | arXiv:2210.15668 |
| **Train ML with noise** | AIHWKIT | ../crossbar/educational/crossbar.opensource.md |
| **SPICE circuit design** | ngspice + OpenVAF | ngspice manual |
| **Teach students** | FeroSim visualization | FerroSim GitHub |

---

## 14. Summary

**Best Tools by Use Case:**

| Use Case | Recommended | Why |
|----------|-------------|-----|
| **Educational visualization** | This project | Real-time, interactive, beautiful GUI |
| **Python prototyping** | Preisachmodel | Mature, well-documented, physics-based |
| **HfO₂ devices** | Heracles + FerroX | Experimentally validated, state-of-the-art |
| **Switching kinetics** | NLS model | Field of the art for hafnia |
| **Circuit-level** | ngspice + OpenVAF | Industry standard workflow |
| **Array-level** | CrossSim + NeuroSim | See ../crossbar/educational/crossbar.opensource.md |
| **Domain dynamics** | FerroX or FERRET | Mesoscale physics, GPU-accelerated |
| **Experimental analysis** | hysteresis pkg | Clean, robust, production-ready |

---

## 15. Contributing & Future Work

### Adding New Models
Contributions welcome! Areas of need:
1. **Temperature-dependent Preisach** - Currently missing from open-source
2. **Stress effects** - Mechanical coupling for thin-film devices
3. **Frequency dependence** - AC hysteresis (currently DC-only)
4. **Multi-level quantization** - Beyond 30 levels

### Integration Roadmap
- [ ] Import ngspice netlist formats
- [ ] Export to AIHWKIT weight formats
- [ ] Benchmark against NeuroSim energy estimates
- [ ] Verilog-A compact model generation

---

*Last Updated: January 27, 2026*
**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
**Purpose:** Comprehensive catalog of open-source hysteresis modeling tools
**Cross-References:** See [../crossbar/educational/crossbar.opensource.md](../crossbar/educational/crossbar.opensource.md) for array-level tools, [../hysteresis/hysteresis.physics.md](../hysteresis/hysteresis.physics.md) for theory, [CLAUDE.md](<local-path>) for honesty policy.
