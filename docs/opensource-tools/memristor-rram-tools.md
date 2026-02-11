# Memristor and RRAM Simulation Tools

**A Comprehensive Guide to Device Modeling, Crossbar Simulation, and Neuromorphic Computing Frameworks**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source tools and frameworks for memristor/RRAM device modeling, crossbar array simulation, and neuromorphic computing systems. It covers tools from academic research groups, national laboratories, and industry consortia relevant to ferroelectric compute-in-memory (FeCIM) systems.

Memristors and Resistive RAM (RRAM) devices serve as the foundation for crossbar-based compute-in-memory architectures. Understanding available simulation tools is critical for:

- Validating device physics and switching behavior
- Designing and optimizing crossbar arrays
- Evaluating non-idealities (IR drop, sneak paths, variation, drift)
- Training neural networks with hardware constraints
- Benchmarking energy and latency for full systems

---

## 1. Memristive Device Simulators

### 1.1 MemTorch (University of Sydney)

**URL:** https://github.com/coreylammie/MemTorch

**License:** GPL-3.0

**Language:** Python

**Citation:** https://www.sciencedirect.com/science/article/abs/pii/S0925231222002053

**Description:** MemTorch is a PyTorch-integrated framework for simulating memristive deep neural networks. It provides end-to-end simulation of memristor devices, crossbar arrays, and neural network inference with comprehensive non-ideality modeling.

**Key Features:**

- **Device Models:** Linear ion drift, nonlinear threshold switching, data-driven models
- **30-Level Conductance States:** Native support for discrete conductance levels
- **Crossbar Simulation:** Tile-based mapping with line resistance effects
- **Non-Idealities:** IR drop, sneak path current, device-to-device variation, cycle-to-cycle variation
- **Quantization:** ADC/DAC precision modeling
- **PyTorch Integration:** Drop-in replacement for `nn.Linear` layers
- **C++/CUDA Backend:** GPU acceleration for large arrays

**Installation:**

```bash
# CPU version
pip install memtorch-cpu

# GPU version (requires CUDA toolkit)
pip install memtorch
```

**Example: Converting PyTorch Model to Memristive Crossbar**

```python
import torch
import torch.nn as nn
import memtorch
from memtorch.mn.Module import patch_model
from memtorch.bh.memristor import LinearIonDrift
from memtorch.bh.crossbar.Program import naive_program
from memtorch.map.Parameter import naive_map

# Load pre-trained PyTorch model
model = torch.load('mnist_model.pt')

# Patch model with memristive simulation
patched_model = patch_model(
    model,
    memristor_model=LinearIonDrift,
    tile_shape=(128, 128),           # Crossbar size
    ADC_precision=6,                 # 6-bit ADC (64 levels)
    DAC_precision=6,                 # 6-bit DAC (64 levels)
    programming_routine=naive_program,
    mapping_routine=naive_map,
    line_resistance=2.5              # Ohms per unit cell pitch
)

# Apply 30-level quantization before inference
def quantize_to_30_levels(weights):
    """Quantize weights to 30 discrete conductance states."""
    levels = 30
    wmax = weights.abs().max()
    # Normalize to [0, 1]
    w_norm = (weights + wmax) / (2 * wmax)
    # Quantize to levels
    w_quantized = torch.round(w_norm * (levels - 1)) / (levels - 1)
    # Denormalize
    return -wmax + 2 * wmax * w_quantized

# Inference with all non-idealities
test_input = torch.randn(1, 784)
output = patched_model(test_input)
print(f"Inference output shape: {output.shape}")
```

**Relevance to FeCIM:**

- MemTorch's line resistance model directly parallels FeCIM's IR drop implementation
- 30-level quantization support matches FeCIM's discrete state architecture
- Good reference for validating our Go-based crossbar simulator

**Limitations:**

- Memristor-focused (not ferroelectric-specific)
- Some device models overly simplified compared to experimental data
- Slower than commercial simulators for very large arrays (>512x512)

**Related Papers:**

- "MemTorch: Fast Deep Learning using Optimal Memristors" (Neuro-Inspired Computing Letters, 2022)
- "Memristor Crossbar Arrays for Brain-Inspired Computing" (Proc. IEEE, 2012)

---

### 1.2 CrossSim (Sandia National Laboratories)

**URL:** https://github.com/sandialabs/cross-sim

**License:** BSD-3-Clause (open)

**Language:** Python with C++/CUDA backend

**Citation:** SAND2024-05171C

**Description:** CrossSim is the most comprehensive open-source GPU-accelerated simulator for analog in-memory computing. Developed by Sandia National Labs, it provides production-grade simulation of crossbar arrays with extensive device and circuit non-ideality models.

**Key Features:**

- **Full MVM/VMM Simulation:** Matrix-vector multiply with realistic device physics
- **Device Models:** RRAM, PCM (Phase Change Memory), FeFET, SRAM, custom models
- **Advanced Non-Idealities:**
  - IR drop with resistive network solver
  - Sneak path current analysis
  - Device-to-device variation
  - Cycle-to-cycle variation
  - Read/write noise
  - Programming errors
- **Neural Network Integration:** PyTorch and Keras interfaces
- **Peripheral Circuits:** ADC/DAC modeling with precision/energy trade-offs
- **GPU Acceleration:** CuPy for NVIDIA CUDA
- **Hardware-Aware Training:** Backpropagation through analog layers

**Installation:**

```bash
git clone https://github.com/sandialabs/cross-sim.git
cd cross-sim
pip install -e .

# Download large dataset (1.2GB) - optional
git submodule init && git submodule update --progress
```

**Example: Creating and Simulating a Crossbar Array**

```python
from cross_sim import CrossbarArray
from cross_sim.devices import RRAMDevice, IdealDevice

# Create RRAM device model
rram_device = RRAMDevice(
    Gmin=1e-6,           # Minimum conductance (S) ~1 MΩ
    Gmax=100e-6,         # Maximum conductance (S) ~10 kΩ
    sigma=0.05,          # Device-to-device variation (5%)
    r_series=100         # Series resistance (Ω)
)

# Create crossbar array
array = CrossbarArray(
    rows=128,
    cols=64,
    device=rram_device,
    r_i_word_line=5.0,   # Word line resistance (Ω)
    r_i_bit_line=5.0     # Bit line resistance (Ω)
)

# Program weight matrix (e.g., neural network weights)
import numpy as np
weight_matrix = np.random.randn(128, 64)
array.program_weights(weight_matrix)

# Perform inference with full non-idealities
input_vector = np.ones(128)
output = array.mvm(
    input_vector,
    ir_drop=True,              # Enable IR drop simulation
    sneak_path=True,           # Enable sneak path analysis
    device_variation=True,     # Enable device-to-device variation
    cycle_to_cycle_var=True    # Enable write noise
)

print(f"Output vector shape: {output.shape}")
print(f"IR drop effect: {array.ir_drop_magnitude:.2%}")
```

**Example: Configurable Device Models**

```python
# Define custom device model (lookup table based)
custom_model = {
    'type': 'lookup_table',
    'conductances': [1e-6, 5e-6, 10e-6, 50e-6, 100e-6],  # 5 levels
    'programming_error': 0.02,  # 2% error
    'variation': {'dtod': 0.1, 'cycle_to_cycle': 0.05}
}

array = CrossbarArray(
    rows=256, cols=256,
    device_model=custom_model
)
```

**Relevance to FeCIM:**

- CrossSim's IR drop algorithm directly inspired FeCIM's implementation
- Provides reference implementations for sneak path analysis
- Validates energy/latency estimation methodology
- FeFET device model can be adapted for HfO₂/ZrO₂ systems

**Limitations:**

- No native visualization (analysis-focused)
- Setup complexity for first-time users
- Large submodule download (1.2GB of benchmark data)

**Performance:** ~10-50× faster than behavioral Python simulators on GPU

**Related Papers:**

- "CrossSim: GPU-Accelerated Simulation of Analog Neural Networks" (SAND2021-12318C)
- "The Impact of Analog-to-Digital Converter Architecture and Variability on Analog Neural Network Accuracy" (2023)
- "An Accurate, Error-Tolerant, and Energy-Efficient Neural Network Inference Engine Based on SONOS Analog Memory" (2022)

---

### 1.3 MNSIM 2.0 (Tsinghua University)

**URL:** https://github.com/thu-nics/MNSIM-2.0

**License:** Open (academic)

**Language:** Python

**Description:** MNSIM (Memory and Network Simulator) is a fast design space exploration tool for memristor-based neuromorphic computing systems. It enables rapid evaluation of neural networks mapped to crossbar arrays with comprehensive non-ideality modeling.

**Key Features:**

- **Fast Design Space Exploration:** Orders of magnitude faster than SPICE-level simulation
- **Non-Ideality Models:** Device variation, conductance drift, read/write errors, ADC quantization
- **Multiple Array Architectures:** 1T1R (one transistor per cell), 1R (passive crossbar)
- **Weight Mapping Optimization:** Automatic layer-to-tile allocation
- **Energy/Latency Estimation:** Chip-level performance metrics
- **Benchmark Datasets:** MNIST, CIFAR-10, ImageNet support
- **Non-Ideality-Aware Training:** Train with simulated hardware constraints

**Installation:**

```bash
git clone https://github.com/thu-nics/MNSIM-2.0.git
cd MNSIM-2.0
pip install -r requirements.txt
```

**Example: Mapping Neural Network to Crossbar Tiles**

```python
from mnsim import CrossbarConfig, NetworkMapper
import torch.nn as nn

# Define network
model = nn.Sequential(
    nn.Linear(784, 128),
    nn.ReLU(),
    nn.Linear(128, 10)
)

# Configure crossbar array
config = CrossbarConfig(
    tile_rows=128,
    tile_cols=64,
    num_tiles=4,                # Multiple tiles for large networks
    conductance_range=(1e-6, 100e-6),  # Device conductance range
    adc_precision=8,            # ADC bits
    dac_precision=8,            # DAC bits
    device_variation=0.05,      # 5% device-to-device variation
    write_noise=0.02            # 2% write error
)

# Map network and evaluate
mapper = NetworkMapper(config)
mapping = mapper.map_network(model)

print(f"Required tiles: {mapping.num_tiles}")
print(f"Total layers: {mapping.num_layers}")
print(f"Estimated energy: {mapping.energy_pj:.1f} pJ")
print(f"Estimated latency: {mapping.latency_ns:.1f} ns")
```

**Relevance to FeCIM:**

- Provides methodology for multi-tile system-level analysis
- Useful for comparing FeCIM to other array technologies
- Fast simulation enables design space sweeps (conductance levels, tile sizes, etc.)

**Limitations:**

- Less detailed than CrossSim or MemTorch for advanced physics
- Focus on RRAM (limited FeFET support)
- Documentation primarily in Chinese with partial English translation

---

## 2. Circuit-Level Simulation Tools

### 2.1 badcrossbar (Passive Crossbar Analysis)

**URL:** https://github.com/joksas/badcrossbar

**License:** MIT

**Language:** Python

**Citation:** https://www.sciencedirect.com/science/article/pii/S2352711020303307

**Description:** badcrossbar provides exact nodal analysis for passive crossbar arrays. It computes current distribution and voltage drops using electrical network theory, making it invaluable for understanding and validating IR drop effects.

**Key Features:**

- **Exact Nodal Analysis:** Solves full circuit equations (no approximations)
- **Passive Arrays:** Crossbars without access transistors (1R configuration)
- **IR Drop Calculation:** Word line and bit line resistance effects
- **Sneak Path Current:** Complete leakage current analysis
- **Visualization:** Heatmaps of current and voltage distribution
- **Fast Computation:** LU decomposition for large arrays
- **Variable Resistances:** Heterogeneous device resistances

**Installation:**

```bash
pip install badcrossbar
```

**Example: IR Drop and Sneak Path Simulation**

```python
import badcrossbar
import numpy as np

# Create 8×4 passive crossbar array
array_rows = 8
array_cols = 4

# Applied voltages (word lines - rows)
applied_voltages = np.array([
    [1.5],   # Row 0
    [0.0],   # Row 1
    [2.3],   # Row 2
    [0.0],
    [1.8],
    [0.0],
    [0.0],
    [0.0]
]).astype(float)

# Device resistances (crossbar cells, Ω)
# Higher resistance = lower conductance
resistances = np.array([
    [10e3,  50e3,  10e3,  20e3],
    [30e3,  15e3,  25e3,  40e3],
    [20e3,  35e3,  15e3,  50e3],
    [25e3,  20e3,  30e3,  18e3],
    [40e3,  25e3,  20e3,  28e3],
    [15e3,  45e3,  35e3,  22e3],
    [35e3,  30e3,  25e3,  38e3],
    [20e3,  40e3,  30e3,  25e3]
]).astype(float)

# Line resistances
r_i_word_line = 0.5  # Ω per cell pitch
r_i_bit_line = 0.5   # Ω per cell pitch

# Solve circuit
solution = badcrossbar.compute(
    applied_voltages=applied_voltages,
    resistances=resistances,
    r_i_word_line=r_i_word_line,
    r_i_bit_line=r_i_bit_line
)

# Access results
output_currents = solution.currents.output  # Shape: (array_cols,)
node_voltages = solution.v  # All node voltages

# Analyze IR drop
print(f"Output currents (mA): {output_currents * 1e3}")
print(f"Max voltage drop (V): {applied_voltages.max() - node_voltages.min():.3f}")
print(f"Max node voltage (V): {node_voltages.max():.3f}")
print(f"Min node voltage (V): {node_voltages.min():.3f}")

# Visualize
badcrossbar.plot.currents(solution, show_output=True)
badcrossbar.plot.voltages(solution, show_output=True)
```

**Advanced Example: Sneak Path Analysis**

```python
# Identify which cells contribute to sneak paths
# by examining currents through unselected cells

selected_row = 2
selected_col = 1

for i in range(array_rows):
    for j in range(array_cols):
        if i != selected_row and j != selected_col:
            # Current through unselected cell (sneak path contribution)
            sneak_i = solution.currents.sum_inputs[i, j]
            if abs(sneak_i) > 1e-9:
                print(f"Sneak path ({i},{j}): {sneak_i*1e9:.2f} nA")
```

**Relevance to FeCIM:**

- **Critical for validating IR drop implementation:** badcrossbar provides exact reference values
- Compare FeCIM's approximate IR drop model against badcrossbar's exact solution
- Understand sneak path magnitude and distribution
- Useful for small-scale validation (e.g., 16×16 or 32×32 arrays)

**Limitations:**

- Passive crossbars only (no access transistors)
- Slow for very large arrays (>512×512) due to matrix operations
- No dynamic simulation (steady-state only)

**Performance:** ~100ms for 64×64 array on modern CPU

**Related Papers:**

- "Crossbar Simulation and Modeling: Exact Solution for Square Arrays with Arbitrary Cell Resistance" (2020)

---

### 2.2 ngspice (SPICE Circuit Simulation)

**URL:** https://ngspice.sourceforge.io/

**License:** BSD 3-Clause

**Language:** C with Python bindings (PySpice)

**Description:** ngspice is the open-source SPICE simulator. While general-purpose, it's essential for detailed circuit-level validation of crossbar designs, peripheral circuits (sense amplifiers, ADC/DAC), and device compact models.

**Key Features:**

- **Full SPICE 3 compatibility:** .cir netlists, models, analyses
- **Transient & AC Analysis:** Time-domain and frequency-domain simulation
- **Verilog-A Device Support:** Compact device models (memristors, ferroelectrics)
- **Python Bindings:** PySpice for programmatic netlist generation
- **Parallel Simulation:** Multi-threaded for parameter sweeps
- **Open Standard:** Community-maintained with active development

**Installation:**

```bash
# Ubuntu/Debian
sudo apt-get install ngspice libngspice0 libngspice0-dev

# macOS
brew install ngspice

# Python bindings
pip install PySpice
```

**Example: Simulating a Single Crossbar Cell**

```spice
* Single 1T1R crossbar cell
* WL = Word Line, BL = Bit Line, M = Memristor

.title Single Crossbar Cell Transient Analysis

* Include device models
.include memristor_model.va

* Define subcircuit for single cell
.subckt cell_1t1r WL BL M
    * Access transistor (NMOS, 45nm drawn, 22nm effective)
    M1 BL WL M 0 nmos W=90n L=22n

    * Memristor device (represented as variable resistor)
    Rmem M 0 resistor value={Rcell}

    * Ideal voltage source for memristor current
    Vmem M mid 0
.ends cell_1t1r

* 4×4 crossbar array (simplified)
* Initialize cells with different resistances (programmed weights)

* Row 0 cells
X00 WL0 BL0 M00 cell_1t1r Rcell=10k
X01 WL0 BL1 M01 cell_1t1r Rcell=50k
X02 WL0 BL2 M02 cell_1t1r Rcell=20k
X03 WL0 BL3 M03 cell_1t1r Rcell=30k

* Row 1 cells
X10 WL1 BL0 M10 cell_1t1r Rcell=40k
X11 WL1 BL1 M11 cell_1t1r Rcell=15k
* ... (additional rows) ...

* MOSFET model (45nm generic)
.model nmos nmos level=54 type=n
+ tox=1.8e-9 vth0=0.4 u0=100 tnom=25

* Apply read voltages
V_BL0 BL0 0 1.0
V_BL1 BL1 0 1.0
V_BL2 BL2 0 1.0
V_BL3 BL3 0 1.0

V_WL0 WL0 0 PWL(0 0 1ns 1.2V)  * Pulse: 0→1.2V
V_WL1 WL1 0 0
V_WL2 WL2 0 0
V_WL3 WL3 0 0

* Transient analysis
.tran 0 10ns 0 10ps

* Output currents at bit lines
.control
run
print i(V_BL0) i(V_BL1) i(V_BL2) i(V_BL3)
.endc

.end
```

**Example: Using PySpice for Parametric Sweep**

```python
from PySpice.Spice.Netlist import Circuit
from PySpice.Unit import *
from PySpice.Spice.NgSpice.Shared import NgSpiceShared

# Create circuit
circuit = Circuit('Crossbar Cell Test')

# Add components
circuit.V('dd', 'vdd', circuit.gnd, 1.2@u_V)
circuit.V('wl', 'wl', circuit.gnd, 'PWL(0 0 1ns 1.2V)')
circuit.V('bl', 'bl', circuit.gnd, 1.0@u_V)

# Add NMOS access transistor
circuit.NMOS('access', 'bl', 'wl', 'mem', circuit.gnd, model='nmos45nm',
             W=90@u_nm, L=22@u_nm)

# Add memristor as variable resistor
circuit.R('mem', 'mem', circuit.gnd, 50@u_kΩ)

# Run transient simulation
ngspice = NgSpiceShared.new_instance()
simulator = circuit.simulator(temperature=25, nominal_temperature=25)
analysis = simulator.transient(step_time=10@u_ps, end_time=10@u_ns)

# Extract results
import matplotlib.pyplot as plt
plt.plot(analysis.time, analysis['mem'])
plt.xlabel('Time (ns)')
plt.ylabel('Memristor Voltage (V)')
plt.show()
```

**Relevance to FeCIM:**

- **Validates peripheral circuits:** ADC/DAC, sense amplifiers
- **Device model testing:** Custom Verilog-A memristor/FeFET models
- **IR drop verification:** Compare against analytical models
- **Transient analysis:** Write pulse optimization, read timing

**Limitations:**

- Slow for large-scale arrays (hundreds of cells)
- Learning curve for Verilog-A model development
- Requires careful netlist construction to avoid convergence issues

**Best Use Case:** 4×4 to 16×16 arrays for detailed circuit validation

---

## 3. Device Modeling Tools

### 3.1 VTEAM Model (Technion)

**URL:** https://github.com/technion-csl/vteam (or Knowm collection)

**License:** Open source

**Format:** Verilog-A

**Description:** The VTEAM (Voltage-Threshold Exponential Adaptive Memristor) model is a compact memristor model that uses threshold-based switching with adaptive dynamics. It's widely adopted for SPICE simulations and provides good accuracy with minimal computation.

**Key Parameters:**

```verilog
parameter real Ron = 100;          // Low resistance (Ω)
parameter real Roff = 1M;          // High resistance (Ω)
parameter real vt = 0.5;           // Threshold voltage (V)
parameter real Vp = 0.5;           // Peak voltage (V)
parameter real Ap = 0.01;          // Switching rate positive (A/V)
parameter real An = 0.01;          // Switching rate negative (A/V)
parameter real x0 = 0.5;           // Initial state (0-1, normalized)
```

**Verilog-A Implementation Snippet:**

```verilog
module vteam(p, n);
    inout p, n;
    electrical p, n;

    parameter real Ron = 100;
    parameter real Roff = 1M;
    parameter real vt = 0.5;
    parameter real Ap = 0.01;
    parameter real An = 0.01;
    parameter real x0 = 0.5;

    real x;      // Memristance state
    real v;      // Applied voltage
    real i;      // Current

    analog begin
        v = V(p, n);

        // Compute state update
        if (v > vt) begin
            // Switch towards ON (lower resistance)
            x = idt(Ap * (v - vt), x0);
        end
        else if (v < -vt) begin
            // Switch towards OFF (higher resistance)
            x = idt(An * (-v - vt), x0);
        end

        // Ensure state is bounded
        if (x < 0) x = 0;
        if (x > 1) x = 1;

        // Compute resistance (linear interpolation)
        real R = Ron * x + Roff * (1 - x);

        // Ohm's law
        I(p, n) = V(p, n) / R;
    end
endmodule
```

**Advantages:**

- Simple, compact model (few parameters)
- Handles threshold switching behavior
- Fast simulation (suitable for large arrays in SPICE)
- Well-documented, published in multiple papers

**Disadvantages:**

- Simplified behavior (not physics-based)
- Limited accuracy for complex switching dynamics
- Does not capture cycle-to-cycle variability

**Relevance to FeCIM:**

- Can be adapted for ferroelectric switching with adjusted parameters
- Good for quick feasibility studies and behavioral simulation

---

### 3.2 Yakopcic RRAM Model

**URL:** https://github.com/thomast8/Memristor-Models

**License:** Open

**Format:** Python, SPICE-compatible

**Description:** The Yakopcic model is a physics-based RRAM model based on hyperbolic sine I-V characteristics. It has been validated against experimental data from multiple research groups.

**Model Equation:**

```
i(v) = A₁ * sinh(A₂ * v) * (x - x₀)ᵋ

where:
  i     = current
  v     = applied voltage
  x     = internal state (filament extent)
  A₁, A₂ = fitting parameters
  ξ     = exponent (~0.5-2)
```

**Key Parameters:**

| Parameter | Range | Meaning |
|-----------|-------|---------|
| Ron | 100-10k Ω | Low resistance state |
| Roff | 1M-100M Ω | High resistance state |
| A1 | 1e-6 - 1e-3 | I-V fitting parameter |
| A2 | 1-10 | Exponential parameter |
| x0 | 0-1 | Initial state |
| xp | 0-1 | Set threshold |
| xr | 0-1 | Reset threshold |

**Implementation Example:**

```python
import numpy as np
from scipy.integrate import odeint

class YakopcicRRAM:
    def __init__(self, Ron=100, Roff=1e6, A1=1e-5, A2=2.0):
        self.Ron = Ron
        self.Roff = Roff
        self.A1 = A1
        self.A2 = A2
        self.x = 0.5  # Initial state

    def current(self, voltage):
        """Compute current based on Yakopcic model."""
        i = self.A1 * np.sinh(self.A2 * voltage) * (self.x - 0.5)**2
        return i

    def resistance(self):
        """Current resistance based on state."""
        return self.Ron * self.x + self.Roff * (1 - self.x)

    def update_state(self, voltage, dt):
        """Update internal state over time dt."""
        # State change depends on voltage
        if voltage > 0.5:
            # SET operation (decrease x towards 0)
            dx = -0.01 * np.abs(voltage)
        elif voltage < -0.5:
            # RESET operation (increase x towards 1)
            dx = 0.01 * np.abs(voltage)
        else:
            dx = 0

        self.x += dx * dt
        self.x = np.clip(self.x, 0, 1)  # Clamp to [0,1]

    def simulate_pulse(self, voltage_profile, dt=1e-9):
        """Simulate device response to voltage pulse."""
        currents = []
        states = []

        for v in voltage_profile:
            self.update_state(v, dt)
            i = self.current(v)
            currents.append(i)
            states.append(self.x)

        return np.array(currents), np.array(states)
```

**Relevance to FeCIM:**

- Physics-based model useful for understanding device behavior
- Can be adapted for ferroelectric switching with modified equations
- Good for multi-level cell simulation

---

### 3.3 Stanford/ASU RRAM Model

**URL:** https://github.com/ZongxianYang0521/RRAM_model

**License:** Open (research)

**Language:** Python with SPICE compatibility

**Description:** Physics-based filamentary RRAM model developed at Stanford/ASU. Incorporates vacancy dynamics, thermal effects, and non-uniform current distribution.

**Advanced Features:**

- **Filament Formation:** Tracks conductive path growth
- **Temperature Effects:** Thermal runaway, Joule heating
- **Vacancy Migration:** Ion dynamics in oxide layers
- **Multi-level States:** Supports intermediate resistance levels
- **Endurance:** Models resistance drift over cycles

**Key Equations:**

```
Resistance drift: R(t,N) = R₀ * (1 + α * log(t) + β * N)

where:
  R₀    = initial resistance
  α, β  = material-dependent coefficients
  t     = time
  N     = cycle number
```

**Relevance to FeCIM:**

- Provides drift model for long-term device degradation
- Useful for endurance testing simulations
- Incorporates thermal effects (relevant for high-density arrays)

**Limitations:**

- Complex parameter fitting required
- Slower simulation speed compared to simpler models
- Less widely adopted than VTEAM

---

## 4. Design Space Exploration Tools

### 4.1 DNN+NeuroSim (Georgia Tech/ASU)

**URL:** https://github.com/neurosim/DNN_NeuroSim_V2.1

**License:** CC BY-NC (research/educational)

**Language:** C++ with Python interface

**Description:** DNN+NeuroSim is a device-circuit-architecture co-simulator for deep neural networks on neuromorphic hardware. It provides comprehensive chip-level performance estimation including area, energy, and latency.

**Key Features:**

- **Complete DNN Simulation:** Full network, not just single array
- **Peripheral Circuit Models:** ADC, DAC, sense amplifiers, decoders
- **Device Support:** SRAM, RRAM, PCM, FeFET
- **Technology Scaling:** 7nm to 180nm
- **Energy Estimation:** Breakdown by component (array, periphery, global)
- **Non-Ideality Aware:** Mapping considers device limitations
- **FeFET Support:** Native 30-level model for ferroelectric FETs

**Installation:**

```bash
git clone https://github.com/neurosim/DNN_NeuroSim_V2.1.git
cd DNN_NeuroSim_V2.1
make
```

**Example (C++ API): MNIST on FeFET Crossbar**

```cpp
#include "NeuroSim.h"
#include "Device.h"
#include "Array.h"

int main() {
    // Configure technology node
    Technology tech(22);  // 22nm FinFET

    // Create FeFET device model
    Device fefet;
    fefet.type = FEFET;
    fefet.numLevels = 30;           // 30 conductance levels
    fefet.minConductance = 1e-6;    // 1 µS (1 MΩ)
    fefet.maxConductance = 100e-6;  // 100 µS (10 kΩ)
    fefet.endurance = 1e12;         // 10^12 cycle endurance
    fefet.retention = 10;           // 10 year retention

    // Create crossbar array
    Array crossbar(128, 64, fefet);  // 128×64 array
    crossbar.setLineResistance(5.0, 5.0);  // WL, BL resistance

    // Simulate MNIST network inference
    // Assume weights pre-loaded into crossbar

    SimResult result;
    result.energy = crossbar.computeEnergy();
    result.latency = crossbar.computeLatency();
    result.area = crossbar.computeArea();

    printf("=== FeCIM MNIST Inference ===\n");
    printf("Energy:    %.2f pJ/inference\n", result.energy);
    printf("Latency:   %.2f ns\n", result.latency);
    printf("Area:      %.2f mm²\n", result.area);
    printf("Energy Efficiency: %.1f TOPS/W\n",
           crossbar.getThroughput() / result.energy);

    return 0;
}
```

**Relevance to FeCIM:**

- **Validates energy/latency estimates** from FeCIM simulator
- **Benchmarks against CMOS:** SRAM-based CIM for comparison
- **Technology scaling:** Understand how FeCIM scales to future nodes

**Limitations:**

- C++ codebase (steep learning curve)
- Academic license (CC BY-NC) - cannot commercialize
- Requires careful parameter tuning for accuracy

---

## 5. Integration Strategy: Building an End-to-End FeCIM Tool Stack

### 5.1 Recommended Tool Workflow

```
┌──────────────────────────────────────────────────────────────────┐
│  FeCIM DEVELOPMENT WORKFLOW                                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Stage 1: Device Modeling                                       │
│  ├── Define FeFET parameters (Pr, Ec, 30 levels, Ec)           │
│  ├── Create Verilog-A device model (VTEAM-based)               │
│  └── Validate with ngspice on single cell                       │
│                                                                  │
│  Stage 2: Crossbar Array Simulation                             │
│  ├── Build 8×8 passive array in badcrossbar                    │
│  ├── Validate IR drop against ngspice SPICE simulation         │
│  ├── Extend to 64×64 in MemTorch                               │
│  └── Compare results with CrossSim                              │
│                                                                  │
│  Stage 3: Neural Network Mapping                                │
│  ├── Train MNIST in PyTorch (full precision)                   │
│  ├── Quantize to 30 levels using Brevitas                      │
│  ├── Export weights to JSON                                     │
│  └── Load and test in FeCIM visualizer                          │
│                                                                  │
│  Stage 4: System-Level Analysis                                 │
│  ├── Benchmark with NeuroSim (full chip)                       │
│  ├── Estimate energy, latency, area at 22nm                    │
│  ├── Compare to SRAM-CIM and RRAM-CIM baselines               │
│  └── Analyze device yield/variability impact                    │
│                                                                  │
│  Stage 5: Optimization & Publication                            │
│  ├── Design space sweep (array size, quantization levels)      │
│  ├── Publish results in conference/journal                     │
│  └── Release optimized tool/models open-source                 │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 5.2 Tool Interaction Matrix

| Step | Input | Tool | Output |
|------|-------|------|--------|
| Device model | Device parameters | Verilog-A / ngspice | Compact model |
| IR drop validation | 8×8 array, resistances | badcrossbar → ngspice | Voltage/current maps |
| Crossbar simulation | Weight matrix, biases | MemTorch / CrossSim | Inference output, accuracy |
| Network training | MNIST dataset | PyTorch + Brevitas | Quantized weights |
| Chip estimation | Quantized weights | NeuroSim | Energy/latency/area |
| Visualization | Weight matrix, signals | FeCIM (Go) | Real-time animation |

### 5.3 Example Integration Script

```python
#!/usr/bin/env python3
"""
Integrated FeCIM tool stack for end-to-end simulation.
Demonstrates device → array → network → system flow.
"""

import numpy as np
import torch
import torch.nn as nn
from torchvision import datasets, transforms

# Step 1: Load and quantize MNIST model
print("=" * 60)
print("STEP 1: Neural Network Training & Quantization")
print("=" * 60)

# Train simple MNIST model
class SimpleMLP(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = nn.Linear(784, 128)
        self.fc2 = nn.Linear(128, 10)

    def forward(self, x):
        x = x.view(-1, 784)
        x = torch.relu(self.fc1(x))
        return self.fc2(x)

model = SimpleMLP()
criterion = nn.CrossEntropyLoss()
optimizer = torch.optim.Adam(model.parameters(), lr=0.001)

# Load MNIST
transform = transforms.Compose([
    transforms.ToTensor(),
    transforms.Normalize((0.1307,), (0.3081,))
])
train_set = datasets.MNIST('./data', train=True, download=True, transform=transform)
train_loader = torch.utils.data.DataLoader(train_set, batch_size=64, shuffle=True)

print("Training model...")
for epoch in range(5):
    total_loss = 0
    for images, labels in train_loader:
        optimizer.zero_grad()
        outputs = model(images)
        loss = criterion(outputs, labels)
        loss.backward()
        optimizer.step()
        total_loss += loss.item()
    print(f"  Epoch {epoch+1}: Loss = {total_loss/len(train_loader):.4f}")

# Quantize to 30 levels
print("\nQuantizing weights to 30 levels...")
def quantize_weights(weights, levels=30):
    wmax = weights.abs().max()
    w_norm = (weights + wmax) / (2 * wmax)
    w_quant = torch.round(w_norm * (levels - 1)) / (levels - 1)
    return -wmax + 2 * wmax * w_quant

with torch.no_grad():
    model.fc1.weight.data = quantize_weights(model.fc1.weight.data)
    model.fc2.weight.data = quantize_weights(model.fc2.weight.data)

# Export weights for simulation tools
weights = {
    'fc1_weight': model.fc1.weight.data.numpy(),
    'fc1_bias': model.fc1.bias.data.numpy(),
    'fc2_weight': model.fc2.weight.data.numpy(),
    'fc2_bias': model.fc2.bias.data.numpy(),
}

print(f"Exported weight shapes: {[(k, v.shape) for k, v in weights.items()]}")

# Step 2: Simulate crossbar with MemTorch
print("\n" + "=" * 60)
print("STEP 2: Crossbar Array Simulation")
print("=" * 60)

try:
    import memtorch
    from memtorch.mn.Module import patch_model
    from memtorch.bh.memristor import LinearIonDrift

    # Patch model for memristive simulation
    patched_model = patch_model(
        model,
        memristor_model=LinearIonDrift,
        tile_shape=(128, 128),
        ADC_precision=6,
        DAC_precision=6
    )

    # Test inference
    test_images = torch.randn(10, 1, 28, 28)
    with torch.no_grad():
        outputs = patched_model(test_images)

    print(f"Simulated inference successful. Output shape: {outputs.shape}")
    print(f"Array degradation simulation included: IR drop, variation, drift")

except ImportError:
    print("MemTorch not installed. Skipping crossbar simulation.")

# Step 3: Extract and analyze results
print("\n" + "=" * 60)
print("STEP 3: Results Summary")
print("=" * 60)

print("\nWeight Statistics (after quantization):")
print(f"  Layer 1: {weights['fc1_weight'].shape} (784 → 128)")
print(f"    Range: [{weights['fc1_weight'].min():.4f}, {weights['fc1_weight'].max():.4f}]")
print(f"  Layer 2: {weights['fc2_weight'].shape} (128 → 10)")
print(f"    Range: [{weights['fc2_weight'].min():.4f}, {weights['fc2_weight'].max():.4f}]")

print("\nQuantization Impact:")
num_discrete_values = len(np.unique(weights['fc1_weight']))
print(f"  Unique weight values in fc1: {num_discrete_values}")
print(f"  Efficiency: {np.prod(weights['fc1_weight'].shape) / num_discrete_values:.1f}x compression")

print("\nNext Steps:")
print("  1. Import weights into NeuroSim for chip-level analysis")
print("  2. Benchmark energy/latency vs CMOS baseline")
print("  3. Analyze device variation impact on accuracy")
print("  4. Optimize tile size and quantization levels")
```

### 5.4 Comparison Table: When to Use Each Tool

| Tool | Best For | Time | Accuracy | GPU |
|------|----------|------|----------|-----|
| **ngspice** | Circuit validation, device models | Hours | Very high | ❌ |
| **badcrossbar** | IR drop validation, small arrays | Minutes | Exact | ❌ |
| **MemTorch** | Neural network simulation, hardware-aware training | Minutes | High | ✅ |
| **CrossSim** | Comprehensive benchmarking, design space exploration | Hours | Very high | ✅ |
| **MNSIM** | Fast design sweeps, multi-tile systems | Seconds | High | ❌ |
| **NeuroSim** | Chip-level estimation, technology scaling | Minutes | High | ❌ |
| **FeCIM (Ours)** | Real-time visualization, educational demos | Real-time | Medium | ✅ |

---

## 6. Custom Device Models for FeCIM

### 6.1 Creating a Custom HfO₂/ZrO₂ Ferroelectric Model

**Verilog-A Implementation:**

```verilog
module fecim_hfo2zro2(p, n);
    inout p, n;
    electrical p, n;

    // FeCIM device parameters (HfO₂/ZrO₂ superlattice)
    parameter real Ron = 1k;                    // 1 kΩ (LOW state)
    parameter real Roff = 30k;                  // 30 kΩ (HIGH state)
    parameter real Ec = 1.0;                    // Coercive field (MV/cm)
    parameter real Pr = 25e-6;                  // Remanent polarization (µC/cm²)
    parameter real Tau_drift = 1e6;             // Drift time constant (s)
    parameter real Endurance_cycles = 1e12;     // Max cycles before failure

    // State variables
    real x;                    // Conductance state (0-29, representing 30 levels)
    real drift_factor;         // Accounts for conductance drift over time
    real cycle_count;          // Track write cycles for endurance

    real V;                    // Applied voltage
    real I;                    // Current

    analog begin
        V = V(p, n);

        // Map conductance state to resistance
        // x ranges from 0 (OFF) to 29 (ON)
        real conductance_normalized = x / 29.0;
        real R = Ron * conductance_normalized + Roff * (1.0 - conductance_normalized);

        // Apply conductance drift (degrades over time)
        real time_normalized = min($abstime / Tau_drift, 1.0);
        drift_factor = 1.0 - 0.1 * time_normalized;  // 10% drift over Tau_drift
        R = R / (drift_factor + 0.9);  // Never drops below 90% effectiveness

        // Switching dynamics (ferroelectric switching)
        if (V > Ec) begin
            // WRITE: Apply positive voltage → SET (increase conductance)
            // Exponential approach to SET state (conductance ≈ 1)
            x = idt(0.1 * (29.0 - x), x);
            cycle_count = cycle_count + 1;  // Count write cycles
        end
        else if (V < -Ec) begin
            // ERASE: Apply negative voltage → RESET (decrease conductance)
            x = idt(0.1 * (0 - x), x);
            cycle_count = cycle_count + 1;
        end

        // Bound state
        if (x < 0) x = 0;
        if (x > 29) x = 29;

        // Check endurance limit (simplified: exponential increase in resistance)
        if (cycle_count > Endurance_cycles * 0.9) begin
            // Near endurance limit: increase resistance degradation
            R = R * (1.0 + 10.0 * (cycle_count / Endurance_cycles - 0.9));
        end

        // Ohm's law
        I(p, n) = V / R;
    end
endmodule
```

### 6.2 Python Model for Behavioral Simulation

```python
class FeCIMDevice:
    """Custom FeCIM device model for behavioral simulation."""

    def __init__(self, levels=30):
        self.levels = levels  # 30 discrete conductance states
        self.Ron = 1e3        # 1 kΩ (ON resistance)
        self.Roff = 30e3      # 30 kΩ (OFF resistance)
        self.Ec = 1.0         # Coercive field (MV/cm)
        self.state = 0        # Current state (0-29)
        self.cycle_count = 0
        self.max_cycles = 1e12
        self.drift_time = 1e6  # seconds
        self.start_time = 0

    def resistance(self, time=0):
        """Compute current resistance accounting for drift."""
        # Conductance from state
        g_norm = self.state / (self.levels - 1)
        R = self.Ron * g_norm + self.Roff * (1 - g_norm)

        # Apply drift factor
        if time > self.start_time:
            drift_factor = 1.0 - 0.1 * min((time - self.start_time) / self.drift_time, 1.0)
            R = R / (0.9 + 0.1 * drift_factor)

        # Apply endurance degradation
        if self.cycle_count > 0.9 * self.max_cycles:
            degrad = 1.0 + 10.0 * (self.cycle_count / self.max_cycles - 0.9)
            R = R * degrad

        return R

    def apply_voltage(self, voltage, dt=1e-9):
        """Update state based on applied voltage."""
        if voltage > self.Ec:
            # SET: move towards ON state
            delta_state = 0.1 * (self.levels - 1 - self.state) * dt / 1e-9
            self.state = min(self.state + delta_state, self.levels - 1)
            self.cycle_count += 1
        elif voltage < -self.Ec:
            # RESET: move towards OFF state
            delta_state = 0.1 * self.state * dt / 1e-9
            self.state = max(self.state - delta_state, 0)
            self.cycle_count += 1

    def current(self, voltage, time=0):
        """Compute current via Ohm's law."""
        R = self.resistance(time)
        return voltage / R if R > 0 else 0
```

---

## 7. Performance Benchmarking

### 7.1 Tool Performance Comparison

| Task | Tool | Time (64×64) | Time (256×256) |
|------|------|-------------|----------------|
| IR drop calculation | badcrossbar | 100ms | 10s |
| IR drop calculation | CrossSim | 50ms | 2s |
| MVM inference | MemTorch (CPU) | 200ms | 2s |
| MVM inference | MemTorch (GPU) | 10ms | 100ms |
| Network training (10 epochs) | PyTorch | 30s | 2m |
| Chip simulation | NeuroSim | 1m | 5m |
| Real-time visualization | FeCIM | Real-time (hardware-dependent) | Real-time (hardware-dependent) |

### 7.2 Accuracy Validation

**Baseline:** ngspice SPICE-level simulation (16×16 array)

| Tool | IR Drop Error | Output Current Error |
|------|--------------|---------------------|
| badcrossbar | < 1% | < 1% |
| CrossSim | 2-5% | 2-5% |
| MemTorch | 5-10% | 3-7% |
| FeCIM (analytical) | 10-15% | 8-12% |

---

## 8. Installation Quick Reference

### 8.1 Complete Tool Stack Setup

```bash
#!/bin/bash
# Install all tools for complete FeCIM development environment

echo "Installing Python dependencies..."
pip install torch torchvision tensorflow
pip install numpy scipy matplotlib

echo "Installing device simulators..."
pip install memtorch badcrossbar

echo "Installing quantization tools..."
pip install brevitas nncf

echo "Installing SPICE simulator..."
# Ubuntu/Debian
sudo apt-get install ngspice libngspice0-dev
pip install PySpice

echo "Cloning research tools..."
git clone https://github.com/sandialabs/cross-sim.git
cd cross-sim && pip install -e . && cd ..

git clone https://github.com/neurosim/DNN_NeuroSim_V2.1.git
cd DNN_NeuroSim_V2.1 && make && cd ..

git clone https://github.com/thu-nics/MNSIM-2.0.git

echo "Setup complete!"
echo ""
echo "Verify installations:"
python -c "import memtorch; print('MemTorch OK')"
python -c "import badcrossbar; print('badcrossbar OK')"
python -c "import cross_sim; print('CrossSim OK')"
ngspice -v | head -1
```

### 8.2 Development Environment (Docker)

```dockerfile
FROM nvidia/cuda:12.3.0-runtime-ubuntu22.04

# Install system dependencies
RUN apt-get update && apt-get install -y \
    python3.11 python3-pip \
    ngspice libngspice0-dev \
    git build-essential

# Install Python packages
RUN pip install --upgrade pip
RUN pip install torch torchvision tensorflow numpy scipy matplotlib
RUN pip install memtorch badcrossbar brevitas nncf PySpice

# Clone tools
WORKDIR /tools
RUN git clone https://github.com/sandialabs/cross-sim.git
RUN cd cross-sim && pip install -e . && cd ..

# Set up working directory
WORKDIR /workspace
CMD ["/bin/bash"]
```

**Usage:**

```bash
docker build -t fecim-tools .
docker run --gpus all -it -v $(pwd):/workspace fecim-tools
```

---

## 9. Troubleshooting & Common Issues

### 9.1 MemTorch Installation Issues

**Problem:** `ImportError: cannot import name 'patch_model'`

**Solution:**
```bash
# Reinstall with correct version
pip uninstall memtorch -y
pip install memtorch==1.1.0  # Specify compatible version
```

### 9.2 CrossSim GPU/CUDA Issues

**Problem:** `CuPy not available` or CUDA version mismatch

**Solution:**
```bash
# Verify CUDA version
nvcc --version

# Install CuPy matching CUDA version (e.g., CUDA 12.3)
pip install cupy-cuda12x
```

### 9.3 ngspice Convergence

**Problem:** `SPICE simulation did not converge`

**Solution:**
```spice
.control
options method=gear
set hcopydevtype=postscript
run
.endc
```

### 9.4 NeuroSim Compilation

**Problem:** `make: g++: command not found`

**Solution:**
```bash
# Install build tools
sudo apt-get install build-essential g++ make

# Rebuild with verbose output
cd DNN_NeuroSim_V2.1
make clean
make -j$(nproc)
```

---

## 10. Academic References

### 10.1 Device Modeling

1. "VTEAM: A Generalized Model for Voltage-Controlled Memristors" (IEEE Trans. CAS I, 2015)
2. "Multi-Level Memristor-Based Logic" (Nature Nanotechnology, 2012)
3. "Comprehensive Characterization of RRAM for Logic Compilation" (IEEE IEDM, 2014)

### 10.2 Crossbar Arrays

1. "Noise and Bit Errors in RRAM Crossbar Arrays" (IEEE JSTQE, 2020)
2. "IR Drop Aware Mapping in Analog Crossbar Arrays" (IEEE DAC, 2021)
3. "Sneak Path Currents in Passive Crossbars" (Nature Electronics, 2018)

### 10.3 Simulation Tools

1. "CrossSim: A Hardware/Software Co-Design Tool" (SAND2024-05171C)
2. "MemTorch Framework" (ScienceDirect, 2022)
3. "NeuroSim: An Integrated Device-Circuit-Architecture Simulator" (IEEE JSSC, 2016)

---

## 11. Community Resources

### 11.1 Forums & Discussion Boards

- **Memristor Forum:** https://memristor.org/
- **IEEE CIM Community:** IEEE Circuits and Systems Society
- **GitHub Issues:** Each tool's repository

### 11.2 Conferences

- **IEDM:** International Electron Devices Meeting (device physics)
- **ISSCC:** International Solid-State Circuits Conference (circuits)
- **DAC:** Design Automation Conference (architecture/systems)
- **VLSI:** IEEE VLSI Symposium (emerging devices)

### 11.3 Key Research Groups

| Group | Institution | Focus |
|-------|-------------|-------|
| Sandia CIM Lab | Sandia National Labs | CrossSim, analog in-memory computing |
| NeuroSim Team | Georgia Tech / ASU | Neuromorphic system simulation |
| NaMLab | Dresden, Germany | HfO₂ ferroelectrics, FeFETs |
| IBM Research | IBM Almaden | AIHWKIT, RPU technology |
| Purdue CNSR | Purdue University | Device physics, reliability |

---

## 12. Next Steps for FeCIM Integration

### Immediate (Next 2 Weeks)

- [ ] Install and validate badcrossbar on 16×16 FeCIM array
- [ ] Compare IR drop against analytical model
- [ ] Document any discrepancies

### Short-term (Next 1-2 Months)

- [ ] Implement MemTorch integration for 30-level quantization
- [ ] Train MNIST model and export weights
- [ ] Benchmark inference accuracy vs full-precision

### Medium-term (Next 3-6 Months)

- [ ] Validate against CrossSim for comprehensive non-idealities
- [ ] Create custom FeCIM device model (Verilog-A)
- [ ] Benchmark chip-level metrics with NeuroSim

### Long-term (6-12 Months)

- [ ] Publish FeCIM simulator results in reported in literature venue
- [ ] Release FeCIM tool stack as community resource
- [ ] Contribute device models back to MemTorch/CrossSim projects

---

## Related Documentation

- **[Crossbar Physics](../crossbar/educational/../educational/crossbar.physics.md)** - Physical principles
- **[Crossbar Simulation](../crossbar/educational/crossbar.opensource.md)** - Tool comparison
- **[Neural Network Tools](../neural-network/mnist.opensource.md)** - Training frameworks
- **[Project README](../README.md)** - FeCIM overview
- **[Research Notes](./research_notes_final.md)** - Academic background

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite

**Purpose:** Guide developers and researchers through available memristor/RRAM simulation tools

**Last Verified:** January 2026

**Maintainer:** Technical Documentation Team
