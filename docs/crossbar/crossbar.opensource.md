# Open-Source Crossbar and CIM Simulation Tools

**A Comprehensive Guide to Available Tools for Crossbar Array and Compute-in-Memory Simulation**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source tools, libraries, and frameworks for simulating crossbar arrays, compute-in-memory (CIM) systems, non-idealities (IR drop, sneak paths, device variation), and neural network mapping. It covers tools from academic research, industry consortia, and the open-source community.

---

## 1. Crossbar Array Simulators

### 1.1 CrossSim (Sandia National Laboratories)

**URL:** https://github.com/sandialabs/cross-sim

**Description:** CrossSim is the most comprehensive open-source crossbar simulator. Developed by Sandia National Labs, it provides physics-accurate simulation of analog crossbar arrays for neural network inference.

**Features:**
- Full MVM/VMM simulation with non-idealities
- IR drop modeling with resistive network solver
- Sneak path analysis for passive arrays
- Device-to-device and cycle-to-cycle variation
- ADC/DAC quantization effects
- Support for RRAM, PCM, and custom device models
- Neural network layer mapping

**Installation:**
```bash
git clone https://github.com/sandialabs/cross-sim.git
cd cross-sim
pip install -e .
```

**Example Usage:**
```python
from cross_sim import CrossbarArray
from cross_sim.devices import IdealDevice, RRAMDevice

# Create crossbar with RRAM devices
array = CrossbarArray(
    rows=128,
    cols=64,
    device=RRAMDevice(
        Gmin=1e-6,      # Minimum conductance (S)
        Gmax=100e-6,    # Maximum conductance (S)
        sigma=0.05      # Device variation (5%)
    )
)

# Program weights
array.program_weights(weight_matrix)

# Perform MVM with non-idealities
output = array.mvm(input_vector, ir_drop=True, sneak_path=True)
```

**Relevance to FeCIM:** CrossSim's architecture directly inspired our Go implementation. Key differences:
- CrossSim: Python, research-focused, slower
- FeCIM: Go, visualization-focused, real-time

**Limitations:** No native FeFET device model (must create custom). Slow for large arrays (>256×256).

---

### 1.2 NeuroSim (Georgia Tech / ASU)

**URL:** https://github.com/neurosim/DNN_NeuroSim_V2.1

**Description:** NeuroSim is a device-circuit-architecture co-simulator for deep neural networks on neuromorphic hardware. It focuses on energy, latency, and area estimation.

**Features:**
- Complete neural network simulation (not just single array)
- CMOS peripheral circuit models (ADC, DAC, sense amplifiers)
- Energy/latency/area estimation for full chip
- Support for SRAM, RRAM, PCM, FeFET
- Technology node scaling (22nm to 7nm)
- Non-ideality aware training support

**Installation:**
```bash
git clone https://github.com/neurosim/DNN_NeuroSim_V2.1.git
cd DNN_NeuroSim_V2.1
make
```

**Example (C++ API):**
```cpp
// Configure technology node
Technology tech(22);  // 22nm

// Create FeFET device
FeFET fefet(
    30,      // Number of levels
    1e-6,    // Gmin
    100e-6,  // Gmax
    1e10     // Endurance
);

// Create crossbar array
CrossbarArray array(128, 64, fefet);

// Run simulation
SimResult result = array.run_inference(input);
printf("Energy: %.2f pJ, Latency: %.2f ns\n",
       result.energy, result.latency);
```

**Relevance to FeCIM:** NeuroSim provides the most realistic peripheral circuit models. Useful for validating our energy estimates.

**Limitations:** C++/Python mixed, complex setup, focused on benchmarking rather than visualization.

---

### 1.3 IBM Analog Hardware Acceleration Kit (AIHWKIT)

**URL:** https://github.com/IBM/aihwkit

**Description:** IBM's production-grade library for simulating and training neural networks on analog crossbar hardware. Based on their RPU (Resistive Processing Unit) research.

**Features:**
- PyTorch integration for seamless neural network training
- Analog-aware training (noise injection, quantization-aware)
- Multiple device models (capacitor, resistive, ideal)
- Tile-based crossbar abstraction
- Hardware-in-the-loop support (for IBM AIU chips)
- Extensive documentation and tutorials

**Installation:**
```bash
pip install aihwkit
# Or with CUDA support
pip install aihwkit-cuda
```

**Example Usage:**
```python
import torch
from aihwkit.nn import AnalogLinear
from aihwkit.simulator.configs import SingleRPUConfig
from aihwkit.simulator.configs.devices import ConstantStepDevice

# Define analog device (30 levels = 5 bits)
rpu_config = SingleRPUConfig(device=ConstantStepDevice())
rpu_config.device.w_min = -1.0
rpu_config.device.w_max = 1.0
rpu_config.device.w_min_dtod = 0.05  # 5% variation

# Create analog layer (replaces nn.Linear)
analog_layer = AnalogLinear(784, 128, bias=True, rpu_config=rpu_config)

# Training works normally
optimizer = torch.optim.SGD(analog_layer.parameters(), lr=0.1)
output = analog_layer(input_tensor)
loss = criterion(output, target)
loss.backward()
optimizer.step()
```

**Relevance to FeCIM:** AIHWKIT is the gold standard for analog-aware training. FeCIM should adopt similar noise injection for training scripts.

**Limitations:** No direct visualization. Device models tuned for IBM hardware.

---

### 1.4 MemTorch (University of Sydney)

**URL:** https://github.com/coreylammie/MemTorch

**Description:** A simulation framework for memristive deep learning systems. Integrates with PyTorch for hardware-aware neural network training.

**Features:**
- Memristor device models (linear, nonlinear, data-driven)
- Crossbar tile mapping with line resistance
- IR drop and sneak path simulation
- ADC/DAC quantization
- Noise and variation injection
- PyTorch Module replacement

**Installation:**
```bash
pip install memtorch
```

**Example Usage:**
```python
import torch
import memtorch
from memtorch.mn.Module import patch_model
from memtorch.map.Parameter import naive_map
from memtorch.bh.crossbar.Program import naive_program

# Load trained PyTorch model
model = torch.load('model.pt')

# Patch model with memristive simulation
patched_model = patch_model(
    model,
    memristor_model=memtorch.bh.memristor.LinearIonDrift,
    tile_shape=(128, 128),
    ADC_precision=6,
    DAC_precision=6,
    programming_routine=naive_program,
    mapping_routine=naive_map,
    line_resistance=2.5  # Ohms per cell pitch
)

# Inference with non-idealities
output = patched_model(input)
```

**Relevance to FeCIM:** MemTorch's line resistance model matches our IR drop implementation. Good reference for validation.

**Limitations:** Memristor-focused (not ferroelectric). Some models are overly simplified.

---

### 1.5 TxSim (NCSU)

**URL:** https://github.com/RuohanRen/TxSim

**Description:** Transient crossbar simulator with detailed SPICE-level accuracy for device physics.

**Features:**
- SPICE-accurate transient simulation
- Parasitic extraction for metal lines
- Device compact models (Verilog-A compatible)
- Write pulse optimization
- Multi-level programming simulation

**Installation:**
```bash
git clone https://github.com/RuohanRen/TxSim
cd TxSim
pip install -r requirements.txt
```

**Relevance to FeCIM:** TxSim provides more accurate transient behavior than our steady-state model. Useful for write-verify simulation.

**Limitations:** Much slower than behavioral simulation. Overkill for inference-only analysis.

---

## 2. Neural Network Mapping Tools

### 2.1 MNSIM (Tsinghua University)

**URL:** https://github.com/thu-nics/MNSIM

**Description:** Maps neural networks to crossbar arrays with optimization for accuracy under non-idealities.

**Features:**
- Automatic layer-to-tile mapping
- Weight quantization optimization
- Non-ideality-aware mapping
- Energy and latency estimation

**Example:**
```python
from mnsim import NetworkMapper

mapper = NetworkMapper(
    tile_size=(128, 128),
    adc_bits=6,
    weight_bits=5
)

mapping = mapper.map_network(pytorch_model)
print(f"Required tiles: {mapping.num_tiles}")
print(f"Estimated energy: {mapping.energy_pj} pJ")
```

**Relevance to FeCIM:** Useful for mapping larger networks to multiple crossbar tiles.

---

### 2.2 DNN+NeuroSim Framework

**URL:** https://github.com/neurosim/DNN_NeuroSim

**Description:** End-to-end framework for mapping DNNs to neuromorphic hardware with benchmark results.

**Features:**
- Full DNN training with hardware constraints
- Automatic crossbar mapping
- Benchmark suite (MNIST, CIFAR-10, ImageNet)
- Comparative analysis across technologies

---

### 2.3 PIMcomp (Peking University)

**URL:** Not publicly available (paper-only)

**Description:** Compiler for Processing-in-Memory systems. Maps computations to crossbar arrays.

**Key Concepts:**
- Dataflow optimization for CIM
- Tiling strategies for large matrices
- Pipelining across multiple arrays

**Note:** Listed in PAPERS_NEEDED.md - paper at ResearchGate.

---

### 2.4 COMPASS (MIT)

**URL:** Available via paper authors

**Description:** Crossbar compiler for DNNs with automatic optimization.

**Paper:** compass_crossbar_compiler_2025.pdf in project repository

**Features:**
- Graph-level optimization
- Kernel fusion for crossbar
- Memory hierarchy optimization

---

## 3. Device Modeling Tools

### 3.1 VTEAM (Technion)

**URL:** https://github.com/technion-csl/vteam

**Description:** Verilog-A memristor model with threshold-based switching.

**Features:**
- Threshold-adaptive memristor model
- SPICE-compatible
- Multi-level state support

**Example (Verilog-A):**
```verilog
module vteam(p, n);
    inout p, n;
    electrical p, n;

    parameter real Ron = 1e3;      // Low resistance
    parameter real Roff = 1e6;     // High resistance
    parameter real vt = 0.5;       // Threshold voltage
    parameter real k = 1e-12;      // Switching rate

    real x;  // State variable

    analog begin
        x = idt((V(p,n) > vt) ? k*(1-x) :
               (V(p,n) < -vt) ? -k*x : 0, x0);
        I(p,n) = V(p,n) / (Ron*x + Roff*(1-x));
    end
endmodule
```

**Relevance to FeCIM:** Can be adapted for FeFET threshold switching.

---

### 3.2 JART VCM (Jülich Research Centre)

**URL:** https://www.fz-juelich.de/ias/jart

**Description:** Physics-based valence change memristor model.

**Features:**
- Kinetic Monte Carlo simulation
- Temperature-dependent switching
- Filament formation model

---

### 3.3 FerroX (Purdue/Berkeley)

**URL:** https://github.com/AMReX-Microelectronics/FerroX

**Description:** Multi-scale ferroelectric device simulator (see hysteresis.opensource.md).

**Relevance to Crossbar:** Can generate device I-V curves for crossbar integration.

---

## 4. Circuit Simulation Tools

### 4.1 ngspice

**URL:** https://ngspice.sourceforge.io/

**Description:** Open-source SPICE simulator. Essential for crossbar peripheral circuit design.

**Features:**
- Full SPICE simulation
- Verilog-A device support
- Python bindings (PySpice)
- Parallel simulation

**Example (Crossbar Cell):**
```spice
* Single crossbar cell with access transistor
.subckt cell WL BL M
    M1 BL WL mem 0 nmos W=45n L=22n
    Rm mem M {Rcell}
.ends cell

* 4x4 crossbar
X00 WL0 BL0 M00 cell Rcell=10k
X01 WL0 BL1 M01 cell Rcell=50k
...

* Apply input voltages
V_BL0 BL0 0 1.0
V_BL1 BL1 0 0.5
V_WL0 WL0 0 1.0
V_WL1 WL1 0 0.0

.tran 1n 100n
.end
```

**Relevance to FeCIM:** Validates our behavioral IR drop model.

---

### 4.2 Xschem + ngspice

**URL:** https://xschem.sourceforge.io/

**Description:** Schematic capture tool that integrates with ngspice.

**Features:**
- GUI schematic editor
- Automatic netlist generation
- Parameter sweeps
- Integration with open PDKs

---

### 4.3 OpenVAF

**URL:** https://github.com/openvaf/openvaf

**Description:** Open-source Verilog-A compiler for device models.

**Features:**
- Fast Verilog-A compilation
- Compatible with ngspice
- Supports complex device models

**Relevance to FeCIM:** Enables custom FeFET compact models.

---

## 5. Visualization and Analysis Tools

### 5.1 Our FeCIM Visualizer (This Project)

**URL:** Local project

**Description:** Go-based real-time crossbar visualization.

**Features:**
- Real-time MVM animation
- IR drop heatmaps
- Sneak path visualization
- 30-level conductance display
- Interactive controls

**Key Files:**
```
module2-crossbar/
├── pkg/crossbar/
│   ├── array.go         # Core MVM implementation
│   ├── irdrop.go        # IR drop simulator
│   ├── sneakpath.go     # Sneak path analyzer
│   └── drift.go         # Conductance drift model
└── pkg/gui/
    └── app.go           # Fyne visualization
```

---

### 5.2 Wafer Maps (Python)

**URL:** Various packages

**Description:** Heatmap visualization for crossbar conductance.

**Example:**
```python
import matplotlib.pyplot as plt
import numpy as np

def plot_crossbar(conductances, title="Crossbar Conductances"):
    fig, ax = plt.subplots(figsize=(10, 8))
    im = ax.imshow(conductances, cmap='viridis', aspect='auto')
    ax.set_xlabel('Column (Bit Line)')
    ax.set_ylabel('Row (Word Line)')
    plt.colorbar(im, label='Conductance (S)')
    plt.title(title)
    plt.show()
```

---

### 5.3 NetworkX for Sneak Path Analysis

**URL:** https://networkx.org/

**Description:** Graph analysis library useful for sneak path enumeration.

**Example:**
```python
import networkx as nx

def find_sneak_paths(rows, cols, target_row, target_col):
    G = nx.DiGraph()

    # Build crossbar graph
    for i in range(rows):
        for j in range(cols):
            G.add_node(f"cell_{i}_{j}")

    # Add edges (current can flow through adjacent cells)
    # ...

    # Find all paths from input to output that bypass target
    sneak_paths = nx.all_simple_paths(G, source, target)
    return list(sneak_paths)
```

---

## 6. Benchmarking and Datasets

### 6.1 MLPerf Tiny

**URL:** https://mlcommons.org/en/inference-tiny/

**Description:** Benchmark suite for edge AI inference.

**Relevance:** Standard benchmark for CIM accuracy comparison.

---

### 6.2 MNIST / CIFAR-10

**Standard Datasets:**
- MNIST: 28×28 grayscale digits
- CIFAR-10: 32×32 color images

**Our Usage:**
```go
// From module3-mnist/pkg/mnist/loader.go
type MNISTLoader struct {
    TrainImages [][]float64
    TrainLabels []int
    TestImages  [][]float64
    TestLabels  []int
}
```

---

## 7. Open PDKs for Physical Design

### 7.1 SkyWater SKY130

**URL:** https://github.com/google/skywater-pdk

**Description:** Open-source 130nm PDK from SkyWater via Google.

**Relevance:** Enable open-source tape-outs of CIM designs.

---

### 7.2 IHP SG13G2

**URL:** https://github.com/IHP-GmbH/IHP-Open-PDK

**Description:** 130nm BiCMOS PDK with ferroelectric options (research).

---

### 7.3 GlobalFoundries GF180MCU

**URL:** https://github.com/google/gf180mcu-pdk

**Description:** 180nm MCU-focused PDK.

---

## 8. Comparison Table

| Tool | Language | Focus | Non-Idealities | Visualization | Training |
|------|----------|-------|----------------|---------------|----------|
| **CrossSim** | Python | Research | ✅ All | Basic | ❌ |
| **NeuroSim** | C++/Python | Benchmark | ✅ All | ❌ | ❌ |
| **AIHWKIT** | Python | Training | ✅ Device | ❌ | ✅ |
| **MemTorch** | Python | PyTorch | ✅ IR, variation | ❌ | ✅ |
| **TxSim** | Python | Transient | ✅ SPICE-level | ❌ | ❌ |
| **FeCIM (Ours)** | Go | Visualization | ✅ IR, sneak, drift | ✅ Real-time | ❌ |

---

## 9. Integration Recommendations for FeCIM

### 9.1 Recommended Tool Stack

| Task | Tool | Rationale |
|------|------|-----------|
| Training | AIHWKIT | Best noise-aware training |
| Validation | CrossSim | Most comprehensive non-idealities |
| Benchmarking | NeuroSim | Industry-standard metrics |
| Visualization | FeCIM (ours) | Real-time, educational |
| SPICE validation | ngspice | Circuit-level accuracy |

### 9.2 Data Exchange Formats

```json
// Recommended weight exchange format (JSON)
{
  "format": "fecim_weights_v1",
  "array_size": [128, 64],
  "levels": 30,
  "conductances": [
    [0.034, 0.967, 0.500, ...],
    [0.133, 0.800, 0.233, ...],
    ...
  ],
  "metadata": {
    "trained_with": "aihwkit",
    "accuracy": 0.87,
    "dataset": "mnist"
  }
}
```

### 9.3 Future Integration Plans

1. **Import CrossSim device models** - Use their calibrated RRAM/PCM models
2. **Export to AIHWKIT** - Enable hardware-aware retraining
3. **Benchmark against NeuroSim** - Validate energy estimates
4. **ngspice validation** - Verify IR drop model accuracy

---

## 10. Installation Quick Reference

```bash
# CrossSim
git clone https://github.com/sandialabs/cross-sim && pip install -e cross-sim

# AIHWKIT
pip install aihwkit

# MemTorch
pip install memtorch

# NeuroSim
git clone https://github.com/neurosim/DNN_NeuroSim_V2.1 && cd DNN_NeuroSim_V2.1 && make

# ngspice (Ubuntu)
sudo apt install ngspice

# Our FeCIM Visualizer
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer
```

---

## 11. Community Resources

### 11.1 Forums and Discussions

- **Memristor Forum:** https://memristor.org/
- **IEEE CIM Community:** IEEE Circuits and Systems Society
- **GitHub Discussions:** Each tool's repository

### 11.2 Conferences

- **IEDM:** International Electron Devices Meeting (device-level)
- **ISSCC:** International Solid-State Circuits Conference (circuit-level)
- **DAC:** Design Automation Conference (architecture-level)
- **MLSys:** Machine Learning and Systems (software-level)

### 11.3 Key Research Groups

| Group | Institution | Focus |
|-------|-------------|-------|
| Sandia CIM | Sandia National Labs | CrossSim, system design |
| Georgia Tech NVM | Georgia Tech | NeuroSim, benchmarking |
| IBM Research | IBM | AIHWKIT, production CIM |
| Purdue CNSR | Purdue | Device physics |
| Stanford NVMLab | Stanford | Non-volatile memory |

---

## Appendix: Tool Feature Matrix

| Feature | CrossSim | NeuroSim | AIHWKIT | MemTorch | FeCIM |
|---------|----------|----------|---------|----------|-------|
| MVM simulation | ✅ | ✅ | ✅ | ✅ | ✅ |
| IR drop | ✅ | ✅ | ❌ | ✅ | ✅ |
| Sneak paths | ✅ | ❌ | ❌ | ❌ | ✅ |
| Device variation | ✅ | ✅ | ✅ | ✅ | ✅ |
| Drift | ❌ | ✅ | ✅ | ✅ | ✅ |
| ADC/DAC | ✅ | ✅ | ✅ | ✅ | ✅ |
| FeFET model | ❌ | ✅ | ❌ | ❌ | ✅ |
| 30 levels | ✅ | ✅ | ✅ | ✅ | ✅ |
| PyTorch integration | ❌ | ❌ | ✅ | ✅ | ❌ |
| Real-time GUI | ❌ | ❌ | ❌ | ❌ | ✅ |
| Energy estimation | ✅ | ✅ | ✅ | ❌ | ❌ |
| Area estimation | ❌ | ✅ | ❌ | ❌ | ❌ |

---

## Related Documentation

- **[Crossbar Physics](crossbar.physics.md)** - Physics fundamentals
- **[Demo Guide](crossbar.demo.md)** - Run the visualization
- **[ELI5 Explanation](crossbar.ELI5.md)** - Simple analogies
- **[Research Papers](crossbar.research.md)** - Academic references

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
**Purpose:** Compare open-source CIM simulation tools and libraries
