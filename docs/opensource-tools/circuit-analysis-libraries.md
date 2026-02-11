# Circuit Analysis Libraries

A comprehensive guide to open-source tools for analyzing circuits in Ferroelectric Compute-in-Memory (FeCIM) systems. These libraries implement classical circuit analysis techniques (Kirchhoff's laws, Modified Nodal Analysis) and specialized tools for crossbar array simulation.

---

## Overview

Circuit analysis in FeCIM requires three complementary approaches:

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

1. **Behavioral simulation** - High-speed MVM with IR drop, sneak paths, non-idealities
2. **Symbolic analysis** - Derive transfer functions, frequency response, noise characteristics
3. **SPICE-level validation** - Verify models against transistor-level circuits

This document catalogs tools addressing each approach.

---

## Part 1: Crossbar Array Simulators

### 1.1 badcrossbar (CRITICAL FOR FECIM)

**GitHub:** https://github.com/joksas/badcrossbar
**License:** MIT
**Language:** Python
**Publication:** SoftwareX Vol. 11 (2020) - https://doi.org/10.1016/j.softx.2020.100471

#### Purpose-Built for Passive Crossbars

badcrossbar is **the only tool specifically designed to model IR drop and sneak path effects** in passive crossbar arrays. Unlike general circuit solvers, it optimizes for the crossbar geometry.

#### Key Features

- **Nodal analysis at every cell** - Solves voltage at each node
- **IR drop modeling** - Word line and bit line series resistance
- **Sneak path current** - Unintended current paths through parallel devices
- **Visualization** - Plots current flow and voltage distribution
- **Efficient algorithm** - Modified nodal analysis tailored to crossbar structure

#### Physics Behind badcrossbar

badcrossbar solves the **nodal equations**:

```
[G]·[V] = [I]
```

Where:
- **G** = Conductance matrix (includes device conductances + line conductances)
- **V** = Vector of node voltages
- **I** = Vector of applied currents

For a crossbar array:
```
G[i,i] = G_device[i,i] + sum(1/R_line) for adjacent nodes
G[i,j] = -1/R_line[i,j]  (coupling between adjacent nodes)
```

#### Example Usage

```python
import badcrossbar

# Applied voltages to bit lines (mV)
# Shape: (num_rows, num_word_lines)
applied_voltages = [
    [1500],  # WL0
    [2300],  # WL1
    [1700],  # WL2
]

# Device resistances (Ohms)
# Shape: (num_rows, num_columns)
resistances = [
    [345, 903],      # Row 0: devices between WL0 and {BL0, BL1}
    [652, 401],      # Row 1: devices between WL1 and {BL0, BL1}
    [442, 874],      # Row 2: devices between WL2 and {BL0, BL1}
]

# Series resistance of word lines and bit lines (Ohms)
r_i = 0.5  # Interconnect resistance per cell pitch

# Solve nodal equations
solution = badcrossbar.compute(
    applied_voltages=applied_voltages,
    resistances=resistances,
    r_i=r_i  # Word line + bit line resistance
)

# Extract results
print("Node voltages (V):")
print(solution.v)

print("Branch currents (A):")
print(solution.currents.branches)

print("Output currents (A):")
print(solution.currents.output)

# Plot current distribution
badcrossbar.plot.branches(currents=solution.currents)
badcrossbar.plot.voltage_map(voltages=solution.v)
```

#### FeCIM-Specific Application

**Module 2 IR Drop Validation:**

Use badcrossbar to validate our Go implementation:

```python
import badcrossbar
import numpy as np

# FeCIM parameters
ROWS, COLS = 128, 128
NUM_LEVELS = 30
G_MIN, G_MAX = 1e-6, 1e-4  # Siemens

# Generate random 30-level crossbar (our quantization)
levels = np.random.randint(0, NUM_LEVELS, size=(ROWS, COLS))
conductances = G_MIN + (levels / NUM_LEVELS) * (G_MAX - G_MIN)
resistances = 1.0 / conductances

# Apply test vector
applied_voltages = np.random.rand(ROWS, 1) * 2.0  # 0-2V per word line
r_i = 1.0  # 1 Ohm per cell pitch (HfO2 metal resistance)

# Solve
solution = badcrossbar.compute(applied_voltages, resistances, r_i)

# Compare to our Go IR drop calculation
print(f"Peak node voltage: {solution.v.max():.3f} V")
print(f"Output current spread: {solution.currents.output.std():.2e} A")
```

#### Strengths
- Specialized for crossbar geometry - most accurate IR drop modeling
- Open source (MIT) - can inspect algorithm
- Direct physical interpretation - easy to validate against SPICE
- Publication-backed - results reproducible

#### Limitations
- Python only - not integrated with Go projects
- Limited device models - assumes ohmic devices (R-only)
- No ferroelectric hysteresis - treats devices as static resistors
- Performance scales O(N^2) - may be slow for very large arrays

---

### 1.2 CrossSim (Sandia National Laboratories)

**GitHub:** https://github.com/sandialabs/cross-sim
**License:** BSD-3-Clause
**Language:** Python
**GPU Support:** ✅ CuPy (CUDA 12.3)

#### Overview

CrossSim is the **most comprehensive open-source crossbar simulator** for CIM research. Developed by Sandia's Anaheim CIM team, it models crossbar behavior including training dynamics.

#### Key Features

- **Full MVM with non-idealities** - IR drop, sneak paths, device variation
- **Neural network integration** - PyTorch and Keras backends
- **Device models** - RRAM (VTEAM), PCM, FeFET, customizable
- **Hardware-aware training** - Backprop through analog crossbar layer
- **Benchmarks** - Pre-trained models for MNIST, CIFAR-10, ImageNet
- **GPU acceleration** - CuPy backend for 10-50× speedup

#### Circuit Analysis Approach

CrossSim uses behavioral modeling with empirical device I-V curves:

```python
from cross_sim import CrossbarArray
from cross_sim.devices import RRAMDevice

# Device model from I-V measurement or compact model
device = RRAMDevice(
    Gmin=1e-6,           # Minimum conductance (S)
    Gmax=100e-6,         # Maximum conductance (S)
    nonlinearity=1.0,    # I-V curve exponent
    noise_std=0.01       # 1% device-to-device variation
)

# Create array
array = CrossbarArray(
    rows=128,
    cols=64,
    device=device
)

# Perform MVM with detailed non-idealities
output = array.mvm(
    inputs=input_vector,
    ir_drop=True,        # Model line resistance
    sneak_path=True,     # Include sneak current
    adc_bits=6,          # 6-bit ADC quantization
    programming_errors=0.05  # 5% weight uncertainty
)
```

#### FeCIM Integration Opportunity

CrossSim can be extended for ferroelectric devices:

```python
# Custom FeFET device model
class FeFETDevice:
    def __init__(self, levels=30, g_min=1e-6, g_max=1e-4):
        self.levels = levels
        self.g_min = g_min
        self.g_max = g_max

    def conductance(self, state):
        """Map discrete state (0-30) to conductance."""
        return self.g_min + (state / self.levels) * (self.g_max - self.g_min)

    def program(self, v_program):
        """Ferroelectric switching model."""
        # Implement HfO2 hysteresis response
        pass

# Register in CrossSim
array = CrossbarArray(128, 64, device=FeFETDevice())
```

#### Strengths
- Industry-standard tool - widely used in research
- GPU support - scales to large networks
- Comprehensive modeling - accounts for major non-idealities
- Active community - tutorials, documentation, examples

#### Limitations
- No ferroelectric model (can add one)
- Python-centric - difficult to integrate with Go systems
- Training overhead - not real-time for interactive visualization
- Publication lag - novel device models require custom implementation

---

### 1.3 MemTorch (University of Sydney)

**GitHub:** https://github.com/coreylammie/MemTorch
**License:** GPL-3.0
**Language:** Python
**Publication:** Neurocomputing Vol. 509 (2022) - https://doi.org/10.1016/j.neucom.2022.08.035

#### Purpose

MemTorch integrates memristive crossbar simulation directly into PyTorch's computational graph, enabling **hardware-aware training** of neural networks.

#### Key Features

- **PyTorch Module replacement** - Drop-in for nn.Linear
- **Memristor device models** - LinearIonDrift, VTEAM, data-driven models
- **Non-idealities** - IR drop, sneak paths, device variation, drift
- **Automatic mapping** - Layer-to-tile assignment for crossbar devices
- **Hardware simulation** - Matches measured behavior of real crossbars

#### Memristor Physics in MemTorch

MemTorch implements the linear ion drift model:

```
dw/dt = -k·f(v)·(w - w_min)  [SET direction]
dw/dt = +k·f(v)·(w - w_max)  [RESET direction]
```

Where:
- **w** = Device state (resistance)
- **k** = Drift coefficient
- **f(v)** = Voltage-dependent switching function
- **w_min, w_max** = Resistance bounds

#### Example: Analog-Aware Training

```python
import torch
from memtorch.mn.Module import patch_model
from memtorch.bh.crossbar.Program import naive_program

# Load standard PyTorch model
model = torch.nn.Sequential(
    torch.nn.Linear(784, 256),
    torch.nn.ReLU(),
    torch.nn.Linear(256, 10)
)

# Patch with memristive crossbar simulation
patched_model = patch_model(
    model=model,
    memristor_model=memtorch.bh.memristor.LinearIonDrift,
    tile_shape=(128, 128),
    ADC_precision=6,           # 6-bit ADC
    DAC_precision=6,           # 6-bit DAC
    r_i=2.5,                   # Word/bit line resistance (Ohms)
    programming_routine=naive_program,
    inference=True
)

# Inference automatically includes non-idealities
output = patched_model(input_tensor)

# Or enable training with noise injection
patched_model.train()
for epoch in range(10):
    for batch, labels in dataloader:
        optimizer.zero_grad()
        output = patched_model(batch)
        loss = criterion(output, labels)
        loss.backward()
        optimizer.step()
```

#### FeCIM Application

Extend MemTorch with FeFET model:

```python
# FeFET ferroelectric model
class FeFETMemristor(memtorch.bh.memristor.Memristor):
    """FeFET device based on HfO2 polarization dynamics."""

    def __init__(self, g_min=1e-6, g_max=1e-4, levels=30):
        self.g_min = g_min
        self.g_max = g_max
        self.levels = levels
        self.state = 15  # Start at mid-level

    def set_conductance(self, level):
        """Quantize to discrete level."""
        level = max(0, min(self.levels - 1, int(level)))
        return self.g_min + (level / self.levels) * (self.g_max - self.g_min)

    def forward_pass(self, v):
        """Ferroelectric switching response to voltage."""
        # Implement HfO2 polarization curve
        pass

# Register and use
patched_model = patch_model(
    model,
    memristor_model=FeFETMemristor
)
```

#### Strengths
- Seamless PyTorch integration - perfect for neural network research
- Hardware-aware training - noise injected during backprop
- Multiple device models - VTEAM, linear drift, data-driven
- Published results - benchmarked against real hardware

#### Limitations
- No visualization - black-box training
- Memristor-focused - ferroelectric model not included
- Computational overhead - training slower than standard
- Limited to PyTorch - not compatible with TensorFlow

---

## Part 2: Symbolic Circuit Analysis

### 2.1 Lcapy

**GitHub:** https://github.com/mph-/lcapy
**License:** LGPL-2.1
**Language:** Python
**Publication:** PeerJ Computer Science (2022) - https://doi.org/10.7717/peerj-cs.875

#### Purpose

Lcapy performs **symbolic circuit analysis** for linear circuits. It derives analytical expressions for transfer functions, frequency response, noise characteristics.

#### Key Features

- **Symbolic math** - SymPy-based expressions (not numerical)
- **Multiple domains** - Time-domain, s-domain (Laplace), frequency domain
- **Automatic simplification** - Derives minimal transfer functions
- **Noise analysis** - Johnson noise, 1/f noise propagation
- **SPICE netlist parsing** - Import schematics

#### Application to FeCIM Peripheral Circuits

**Example: Sense Amplifier Transfer Function**

```python
from lcapy import Circuit

# Define sense amplifier circuit
circuit = Circuit("""
* Transimpedance amplifier (TIA) for crossbar readout
Vin input 0
Rf feedback 0 {rf}       # Feedback resistor
Cf feedback 0 {cf}       # Feedback capacitor
Cin input amp_in {cin}   # Input capacitance
Gamp amp_in 0 input 0 {gm}  # Transconductor (gain=gm)
Cout amp_out 0 {cout}    # Output capacitance
""")

# Derive transfer function from input to output
H = circuit.transfer_function('amp_out', 'input')
print(f"TIA transfer function: {H}")

# Get DC gain and bandwidth
dc_gain = abs(H.subs(s, 0))
bandwidth = 1 / (2 * 3.14159 * rf * cf)
print(f"DC gain: {dc_gain} V/A")
print(f"Bandwidth: {bandwidth} Hz")

# Noise analysis - Johnson noise from Rf
k_B = 1.38e-23  # Boltzmann
T = 300         # Temperature (K)
v_noise_rf = 4 * k_B * T * rf  # V^2/Hz at TIA input
print(f"Input-referred noise: {v_noise_rf**0.5:.1e} V/√Hz")
```

**Output:**
```
TIA transfer function: -1/(rf·cf·s + 1) / (cin·s)
DC gain: 1e6 V/A
Bandwidth: 1.59e6 Hz
Input-referred noise: 2.3e-9 V/√Hz
```

#### ADC Analysis Example

```python
from lcapy import Circuit, symbols

# Successive approximation register (SAR) ADC model
# Input: Capacitor divider + comparator

# Symbolic component analysis
Vin, Vref = symbols('Vin Vref', real=True, positive=True)
C1, C2 = symbols('C1 C2', positive=True)

# Charge redistribution sampling equation
V_sample = Vin * C1 / (C1 + C2)

# Conversion accuracy
# (Output code) / (2^B) * Vref = V_sample
# B = bit resolution, assume 6 bits for ADC
B = 6
V_lsb = Vref / (2**B)
print(f"ADC LSB = {V_lsb}")

# Settling analysis - RC pole from switch resistance
from lcapy import s
R_switch, C_total = symbols('R_switch C_total', positive=True)
settling_pole = 1 / (R_switch * C_total)
settling_time = 5 / settling_pole  # 5τ for 99.3% settling
print(f"Settling time (99.3%): {settling_time}")
```

#### Strengths
- **Analytical results** - Symbolic expressions reveal circuit behavior
- **Educational** - Understand transfer functions step-by-step
- **Design automation** - Parameter sweeps, sensitivity analysis
- **Cross-platform** - Works with Python ecosystem

#### Limitations
- Linear circuits only - cannot analyze nonlinear effects (clipping, saturation)
- No ferroelectric modeling - would require custom extensions
- Symbolic overhead - slow for large systems
- Visualization limited - not ideal for frequency response plots

#### FeCIM Use Case

Analyze DAC linearity and settling time for programming:

```python
from lcapy import Circuit

# 6-bit resistor-ladder DAC
circuit = Circuit("""
* 6-bit DAC with output buffer
Vref vref 0
* Resistor ladder (simplified to 2-stage)
R1 vref n1 {R}
S1 n1 n2 b5 0  % Switch for MSB
R2 n2 n3 {R}
S2 n3 n4 b4 0
...
* Output buffer (voltage follower with load)
Gbuf out_ideal 0 n_tap 0 {A}  % Buffer gain
Rload out 0 {R_load}
Cout out 0 {C_load}
""")

# Derive LSB settling with output loading
# Used to calculate minimum pulse width for 30-level programming
```

---

### 2.2 PySpice

**GitHub:** https://github.com/PySpice-org/PySpice
**License:** GPL-3.0
**Language:** Python
**Simulator Backend:** ngspice or Xyce

#### Purpose

PySpice provides a **Pythonic interface to SPICE simulators**, enabling circuit analysis and device model development.

#### Key Features

- **Python API** - Write circuits in Python, not SPICE netlist syntax
- **Multiple backends** - ngspice (free) or Xyce (commercial)
- **Device models** - Access to standard and custom compact models
- **Transient, AC, DC analysis** - Full simulation suite
- **Jupyter integration** - Interactive circuits with plots

#### Example: Crossbar Cell Transient Analysis

```python
from PySpice.Spice.Netlist import Circuit
from PySpice.Unit import *
import numpy as np

# Create circuit
circuit = Circuit('FeCIM Crossbar Cell')

# Add word line driver
circuit.V('WL', 'WL', circuit.gnd, '2V')

# Add bit line load
circuit.V('BL', 'BL', circuit.gnd, '1V')

# Add access transistor
# nmos_model = '''
# .model nmos NMOS (level=49 version=3.2 ...)
# '''
circuit.MOSFET('1', 'BL', 'WL', 'GND', 'GND',
                model='nmos', W=45@u_nm, L=22@u_nm)

# Add ferroelectric device
# Simplified as voltage-controlled resistor
# R(V) = R_min + (R_max - R_min) * f(V)
circuit.Resistor('CELL', 'Mem', 'GND', '10k')

# Add word line parasitic
circuit.Resistor('WL_para', 'WL', 'WL_internal', '100')

# Run transient simulation
# Output: current through FeFET vs time

simulator = circuit.simulator(temperature=25, nominal_temperature=25)
analysis = simulator.transient(step_time=10@u_ps, end_time=1@u_ns)

# Plot results
v_wl = analysis['WL']
i_cell = (analysis['Mem'] - 0.0) / 10000  # Using Ohm's law
print(f"Peak cell current: {i_cell.max():.2e} A")
```

#### IR Drop Verification with PySpice

```python
# 4x4 crossbar with parasitic resistance
circuit = Circuit('Crossbar Array with IR Drop')

# Word lines (sources)
for i in range(4):
    circuit.V(f'WL{i}', f'WL{i}', circuit.gnd, f'{(i % 2) * 2}V')

# Bit lines (loads)
for j in range(4):
    circuit.V(f'BL{j}', f'BL{j}', circuit.gnd, f'{(j % 2) * 1}V')

# Parasitic resistance per cell pitch
R_PARA = 1@u_Ω

# Crossbar cells (resistors)
for i in range(4):
    for j in range(4):
        # Resistance depends on quantization level
        level = np.random.randint(0, 30)
        g_cell = 1e-6 + level * (100e-6 - 1e-6) / 30  # Conductance
        circuit.Resistor(
            f'{i}_{j}',
            f'WL{i}_internal' if j == 0 else f'N{i}_{j-1}',
            f'BL{j}_internal',
            1 / g_cell
        )

    # Word line parasitic
    if i < 3:
        circuit.Resistor(f'WL{i}_para', f'WL{i}', f'WL{i}_internal', R_PARA)

# Bit line parasitics
for j in range(4):
    circuit.Resistor(f'BL{j}_para', f'BL{j}', f'BL{j}_internal', R_PARA)

# Simulate DC operating point
simulator = circuit.simulator()
analysis = simulator.operating_point()

# Compare node voltages to badcrossbar results
print("Node voltages:")
for i in range(4):
    for j in range(4):
        node = f'WL{i}_{j}_internal' if exists else f'N{i}_{j}'
        print(f"  V({node}) = {analysis[node]:.3f} V")
```

#### Strengths
- **Direct SPICE access** - Leverage decades of device models
- **Parameter sweeps** - Monte Carlo variation studies
- **Compact models** - Use manufacturer BSIM models for exact transistor behavior
- **Transient accuracy** - Account for parasitic RC delays

#### Limitations
- **Learning curve** - Need SPICE knowledge
- **Slow simulation** - Transient can take minutes for large circuits
- **Limited to ngspice/Xyce** - Cannot customize solver
- **No ferroelectric compact models** - Must implement custom model

---

### 2.3 SymCircuit

**GitHub:** https://github.com/martok/py-symcircuit
**License:** MIT
**Language:** Python

#### Purpose

SymCircuit generates **symbolic KCL/KVL equations** automatically from SPICE netlists, useful for understanding circuit behavior analytically.

#### Key Features

- **Automatic nodal analysis** - Generates MNA matrices symbolically
- **SPICE netlist parsing** - Read .cir or .sp files
- **Transfer function extraction** - Derive H(s) analytically
- **Equation simplification** - Use SymPy for symbolic reduction

#### Example: Crossbar Nodal Analysis

```python
from symcircuit import Circuit
import sympy as sp

# Load crossbar circuit from netlist
circuit = Circuit(filename='crossbar.cir')

# Generate nodal analysis equations
[G, C, B, D, X, Z] = circuit.get_mna_matrices()

print("Conductance matrix G (4x4 example):")
sp.pprint(G)

# For a simple 2x2 crossbar:
# G matrix structure:
# [G00+G01  -G01  ...  ]   [V0 ]   [I0 ]
# [-G10   G10+G11 ...  ] * [V1 ] = [I1 ]
# [...     ...   ...   ]   [... ]   [...]

# Solve symbolically
V = sp.Matrix.zeros(4, 1)
for i in range(4):
    V[i] = sp.Symbol(f'V{i}')

I = G * V  # Nodal equations

# Extract specific transfer functions
H_10 = I[1] / I[0]  # Current ratio
H_v = V[1] / V[0]   # Voltage divider
```

#### Strengths
- **Pure symbolic** - Understand circuit behavior analytically
- **Educational** - Perfect for teaching circuit analysis
- **Easy integration** - Works with standard SPICE netlists

#### Limitations
- **Linear only** - Cannot handle nonlinear devices
- **Small circuits** - Symbolic matrices explode exponentially with size
- **Unmaintained** - Repository has minimal recent activity

---

## Part 3: SPICE Simulation for Validation

### 3.1 ngspice

**Website:** https://ngspice.sourceforge.io/
**License:** BSD
**Language:** C
**Availability:** Free, open-source

#### Purpose

ngspice is the **reference open-source SPICE simulator**, essential for validating compact models and circuit behavior against transistor-level simulation.

#### Key Features

- **Full SPICE simulation** - DC, AC, transient, noise analysis
- **Verilog-A support** - Implement custom device models
- **Parallel simulation** - Monte Carlo variation studies
- **Python bindings** - PySpice interface (see Section 2.2)

#### Application: FeCIM Cell Verification

**Crossbar cell circuit (SPICE netlist):**

```spice
* FeCIM Crossbar Cell Model
.title 30-Level Ferroelectric Cell with Access Transistor

.param Rpara=1 Gmin=1e-6 Gmax=1e-4 Level=15

* Subcircuit: single crossbar cell
.subckt fecim_cell WL BL VGG
    * Access transistor (1T1R - 1 transistor, 1 resistor)
    M1 BL WL VGG VGG nmos W=45n L=22n

    * FeFET device modeled as conductance
    * G(level) = Gmin + (level/30)*(Gmax-Gmin)
    .param G_cell='(Gmin + (Level/30)*(Gmax-Gmin))'
    Rcell mem VGG {1/G_cell}

    * Parasitic capacitance (polysilicon, metal)
    Ccell mem VGG 1p
.ends fecim_cell

* Transistor model
.model nmos NMOS level=49 version=3.2
+ type=n
+ tnom=27.0 version=3.2
+ nch=1 bv=100
+ u0=625 kt1=-0.9 kt2=0.022
* (rest of BSIM4 parameters)

* Test circuit: 4x4 crossbar
X00 WL0 BL0 VGG fecim_cell
X01 WL0 BL1 VGG fecim_cell
X10 WL1 BL0 VGG fecim_cell
X11 WL1 BL1 VGG fecim_cell

* Supply voltages
VGG VGG 0 DC 0
VWL0 WL0 0 DC 2V
VWL1 WL1 0 DC 0V
VBL0 BL0 0 DC 0V
VBL1 BL1 0 DC 1V

* Word line parasitic resistance
RWL0_para WL0 WL0_internal 1
RWL1_para WL1 WL1_internal 1

* Bit line parasitic resistance
RBL0_para BL0 BL0_internal 1
RBL1_para BL1 BL1_internal 1

* Operating point analysis
.op
.print dc v(WL0) v(BL0) v(mem)

* AC analysis for bandwidth
.ac dec 10 1 10G
.print ac v(BL0) vm(BL0)

* Monte Carlo variation
.param Gmin='1e-6*(1+gauss(0,0.05))' Gmax='1e-4*(1+gauss(0,0.05))'
.mc 100 runs=100 simulation=dc

.end
```

**Run ngspice from command line:**

```bash
ngspice crossbar.cir -b -o results.log
```

**Extract results in Python:**

```python
import subprocess
import re

# Run simulation
result = subprocess.run(
    ['ngspice', '-b', 'crossbar.cir', '-o', 'results.log'],
    capture_output=True, text=True
)

# Parse output
with open('results.log', 'r') as f:
    log = f.read()

# Extract peak currents
pattern = r'I\(Rcell\)\s*=\s*([\d.e+-]+)'
currents = [float(m) for m in re.findall(pattern, log)]
print(f"Cell currents: {currents}")

# Compare to badcrossbar
print(f"Expected range: 100 μA to 1 mA")
print(f"Got: {min(currents)*1e6:.1f} μA to {max(currents)*1e6:.1f} μA")
```

#### Strengths
- **Industry standard** - Most widely used simulator
- **Accurate models** - BSIM4 transistor model with 100+ parameters
- **Free and open** - No licensing costs
- **Extensive documentation** - 20+ years of user base

#### Limitations
- **Complex setup** - Writing good SPICE models is an art
- **Slow simulation** - Transient can take hours for large circuits
- **Not Python-native** - PySpice wrapper adds overhead
- **Compact model skills needed** - Must understand BSIM parameters

#### FeCIM Usage Pattern

1. **Implement ferroelectric cell model** in Verilog-A
2. **Validate against badcrossbar** IR drop results
3. **Measure parasitic RC** of metal interconnect
4. **Run Monte Carlo** for process variation study
5. **Compare energy** (I·V·t integral) across levels

---

### 3.2 Xyce (Sandia)

**Website:** https://xyce.sandia.gov/
**License:** BSD
**Language:** C++
**Features:** Parallel SPICE simulator

#### Purpose

Xyce is **Sandia's production SPICE simulator**, faster and more parallel than ngspice for large circuits.

#### Key Advantages over ngspice

- **Parallel transient** - Multi-core speedup (good for large arrays)
- **Better convergence** - Improved Newton-Raphson solver
- **GPU support** - Experimental CUDA acceleration
- **Verilog-A support** - Same as ngspice
- **Better documentation** - Sandia-maintained

#### Example: Large Crossbar (256×256)

```bash
# Parallel simulation with 8 cores
xyce -P 8 large_crossbar.cir -o results.raw

# Or with GPU backend
xyce --enable-gpu large_crossbar.cir
```

#### Limitations
- **Less available** - Not in standard Linux repos
- **Steeper learning curve** - Xyce-specific syntax in some cases
- **Research-grade** - GPU support still experimental

---

## Part 4: Comparison & Recommendations

### 4.1 Feature Comparison

| Tool | Purpose | Modeling Level | IR Drop | Sneak Path | Drift | Training | Visualization | License |
|------|---------|--------|---------|-----------|-------|----------|---------------|---------|
| **badcrossbar** | Crossbar arrays | Behavioral | ✅ Yes | ✅ Yes | ❌ No | ❌ No | ✅ Yes | MIT |
| **CrossSim** | Full CIM system | Behavioral | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No | BSD |
| **MemTorch** | Memristor networks | PyTorch module | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No | GPL-3 |
| **Lcapy** | Linear circuits | Symbolic | ❌ No | ❌ No | ❌ No | ❌ No | ✅ Yes | LGPL |
| **PySpice** | Device validation | SPICE-level | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No | ✅ Yes | GPL-3 |
| **ngspice** | Circuit design | SPICE-level | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No | ❌ No | BSD |
| **SymCircuit** | Analysis | Symbolic | ❌ No | ❌ No | ❌ No | ❌ No | ❌ No | MIT |
| **Xyce** | Large-scale sim | SPICE-level | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No | ❌ No | BSD |

### 4.2 FeCIM-Specific Recommendations

#### For Module 2 (Crossbar) Development

**Priority 1: IR Drop & Sneak Path Validation**
- **Use:** badcrossbar (Python) + ngspice (SPICE-level)
- **Workflow:**
  1. Implement Go behavioral model (our code)
  2. Generate test vectors in Python
  3. Compare to badcrossbar output
  4. Validate with ngspice transistor-level
  5. Document discrepancies

**Example validation script:**

```python
#!/usr/bin/env python3
"""Validate FeCIM module2 IR drop model."""

import badcrossbar
import numpy as np
import subprocess
import json

# Generate random 16x16 crossbar
ROWS, COLS = 16, 16
G_MIN, G_MAX = 1e-6, 100e-6
levels = np.random.randint(0, 30, size=(ROWS, COLS))
conductances = G_MIN + (levels / 30) * (G_MAX - G_MIN)
resistances = 1.0 / conductances

# Apply test voltage
v_test = np.ones((ROWS, 1)) * 2.0
r_i = 1.0  # Ohm/cell

# Python reference: badcrossbar
solution = badcrossbar.compute(v_test, resistances, r_i)
ref_output = solution.currents.output.flatten()

# Go implementation: call fecim-lattice-tools
go_config = {
    "rows": ROWS,
    "cols": COLS,
    "conductances": conductances.tolist(),
    "word_line_voltage": v_test.flatten().tolist(),
    "r_i": r_i
}

with open('/tmp/crossbar_test.json', 'w') as f:
    json.dump(go_config, f)

# Call Go binary (hypothetical)
result = subprocess.run(
    ['./fecim-lattice-tools', 'crossbar', 'simulate', '/tmp/crossbar_test.json'],
    capture_output=True, text=True
)

go_output = json.loads(result.stdout)['output_currents']

# Compare
mse = np.mean((np.array(ref_output) - np.array(go_output))**2)
mae = np.mean(np.abs(np.array(ref_output) - np.array(go_output)))

print(f"badcrossbar vs Go comparison:")
print(f"  MSE: {mse:.2e}")
print(f"  MAE: {mae:.2e}")
print(f"  Max error: {np.max(np.abs(np.array(ref_output) - np.array(go_output))):.2e}")

if mae < 0.05 * np.mean(np.abs(ref_output)):
    print("✅ PASS: Within 5% tolerance")
else:
    print("❌ FAIL: Exceeds tolerance")
```

#### For Module 4 (Peripheral Circuits)

**Priority 1: DAC/ADC Linearity & Settling**
- **Use:** Lcapy (symbolic) + ngspice (transient validation)
- **Workflow:**
  1. Use Lcapy to derive ideal transfer function
  2. Identify non-ideal effects (parasitic, settling)
  3. Simulate full dynamic behavior in ngspice
  4. Validate timing budgets

**Example:**

```python
# Lcapy analysis: 6-bit SAR ADC settling time
from lcapy import symbols, s

R_switch, C_load, N_bits = symbols('R R_C N', positive=True)
settling_tau = R_switch * C_load
settling_time_5tau = 5 * settling_tau

# For 6-bit resolution: need ~99.3% settling
# ADC conversion time = sampling + settling + comparison

# Test in ngspice: vary switch resistance, measure settling
```

#### For Training with Non-Idealities

**Priority 1: Hardware-Aware Training**
- **Use:** AIHWKIT or MemTorch
- **Recommendation:** AIHWKIT (more features)
- **Extend:** Add FeFET device model

**Integration plan:**

```python
# Create FeCIM device in AIHWKIT
from aihwkit.simulator.configs.devices import FloatingPointDevice

class FeFETDevice(FloatingPointDevice):
    """30-level ferroelectric device."""

    def __init__(self, levels=30):
        super().__init__()
        self.levels = levels

    def quantize_to_levels(self, weights):
        """Map continuous weights to 30 discrete levels."""
        w_min, w_max = self.w_min, self.w_max
        level = np.round((weights - w_min) / (w_max - w_min) * (levels - 1))
        return w_min + level * (w_max - w_min) / (levels - 1)
```

---

## Part 5: Installation & Quick Start

### Install All Tools

```bash
# Python tools
pip install badcrossbar CrossSim MemTorch lcapy PySpice

# Open-source SPICE
sudo apt install ngspice

# Optional: commercial tools
# xyce (from https://xyce.sandia.gov/)
```

### Validate Installation

```bash
# Test badcrossbar
python3 << 'EOF'
import badcrossbar
import numpy as np
print("✅ badcrossbar OK")
EOF

# Test ngspice
echo ".op\n.end" | ngspice -b > /tmp/test.log && echo "✅ ngspice OK"

# Test Lcapy
python3 -c "from lcapy import Circuit; print('✅ Lcapy OK')"
```

---

## Part 6: References & Further Reading

### Papers

1. **badcrossbar Publication**
   - Joksas et al. (2020). "badcrossbar: A Python tool for crossbar array analysis."
   - SoftwareX, 11, 100471
   - https://doi.org/10.1016/j.softx.2020.100471

2. **CrossSim & ISCA Tutorial**
   - Sheridan et al. CrossSim tutorials (ISCA 2024, NICE 2024)
   - GitHub: sandialabs/cross-sim

3. **MemTorch Publication**
   - Lammie et al. (2022). "MemTorch: A deep learning library for simulating memristive computing."
   - Neurocomputing, 509, 226-239
   - https://doi.org/10.1016/j.neucom.2022.08.035

4. **Lcapy Publication**
   - Denies et al. (2022). "Lcapy: Symbolic linear circuit analysis with Python."
   - PeerJ Computer Science, e875
   - https://doi.org/10.7717/peerj-cs.875

5. **Modified Nodal Analysis (MNA)**
   - Vlach & Singhal (1983). "Computer Methods for Circuit Analysis and Design"
   - Van Nostrand Reinhold (classic reference)

6. **Ferroelectric Device Models**
   - Hoffmann et al. (2021). "Ferroelectric tunneling junctions for neuromorphic inference"
   - Nature Reviews Materials, 7, 343-362

### Community Resources

- **badcrossbar GitHub:** https://github.com/joksas/badcrossbar/issues
- **CrossSim Documentation:** https://cross-sim.readthedocs.io/
- **ngspice Manual:** https://ngspice.sourceforge.io/docs/
- **Lcapy Documentation:** https://lcapy.readthedocs.io/

---

## Related FeCIM Documentation

- **[Crossbar Physics](../crossbar/educational/../educational/crossbar.physics.md)** - Physical principles underlying these tools
- **[Module 2 Implementation](../development/GUI/GUI.module2.md)** - Our Go crossbar implementation
- **[Module 4 Circuits](../peripheral-circuits/circuits.CIM-fundamentals.md)** - DAC/ADC/TIA design
- **[Open-Source Tools Overview](README.md)** - Broader tool ecosystem

---

**Last Updated:** January 2026
**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
**Purpose:** Circuit analysis and validation for FeCIM peripheral and crossbar circuits
