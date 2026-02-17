# Module 2: Crossbar - Open-Source Tools

**Navigation:** [← Module 2 Index](./README.md) | [Physics](./physics.md) | [Features](./features.md) | [Architecture](./architecture.md)

---

## Evidence Status

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

---

## Overview

This document catalogs open-source tools for crossbar array simulation, neural network hardware mapping, and analog computing. It covers how this module integrates with external tools and provides guidance for extending functionality.

---

## Dependencies Used by This Module

### Core Dependencies

| Tool | Purpose | License |
|------|---------|---------|
| **Go toolchain** | Build/runtime | BSD-style |
| **Fyne** | GUI framework | BSD-3-Clause |

### Optional External Tools (Referenced)

| Tool | Purpose | Integration |
|------|---------|-------------|
| **CrossSim** (Sandia) | Crossbar validation | Test reference |
| **BadCrossbar** | IR drop validation | Test reference |

---

## Crossbar Simulation Tools

### 1. CrossSim (Sandia National Labs)

**URL:** https://github.com/sandialabs/cross-sim

**Description:** Crossbar array simulator with device non-idealities for analog neural networks.

**Features:**
- Device variation modeling
- Noise injection
- Programming non-idealities
- Inference accuracy analysis

**Integration with this module:**
```python
# Export from this module
array.ExportNumPy("weights.npy")

# Load in CrossSim
import numpy as np
import simulator

weights = np.load("weights.npy")
core = simulator.CrossbarCore(weights)
output = core.vmm(input_vector)
```

**Use Case:** Validate MVM results against established simulator

---

### 2. BadCrossbar (Python)

**URL:** https://github.com/joksas/badcrossbar

**Description:** Python package for computing currents and voltages in passive crossbar arrays.

**Features:**
- IR drop calculation
- Parasitic effects
- Arbitrary applied voltages
- Visualization tools

**Example:**
```python
import badcrossbar

# Define resistances (inverse of conductance)
resistances = 1.0 / crossbar_conductances

# Compute with applied voltages
solution = badcrossbar.solve(resistances, voltages)
currents = solution.output_currents
```

**Comparison:** Our IR drop model is simplified but faster; BadCrossbar provides full circuit simulation.

---

### 3. NeuroSim (Georgia Tech)

**URL:** https://github.com/neurosim

**Description:** Device-to-architecture framework for neuromorphic computing benchmarks.

**Features:**
- FeFET, RRAM, PCM device models
- Crossbar peripheral circuits (DAC, ADC, S&H)
- Multi-level cell support
- Energy and area estimation

**Integration:**
```cpp
// Import weights from this module
#include "neurosim.h"

// Map to NeuroSim device
Device fefet;
fefet.conductance_levels = 30;
fefet.conductance_min = 10e-6;  // 10 µS
fefet.conductance_max = 100e-6; // 100 µS

Array array(128, 128, fefet);
array.loadWeights("weights.csv");
```

**Use Case:** Architecture-level energy and area estimates

---

### 4. MNSIM (Python)

**URL:** https://github.com/thu-nics/MNSIM

**Description:** Multi-level neural network simulator for computing-in-memory.

**Features:**
- Multi-bit cells
- Hybrid precision
- Hardware-aware training
- Latency modeling

---

### 5. AIHWKIT (IBM)

**URL:** https://github.com/IBM/aihwkit

**Description:** PyTorch toolkit for analog AI hardware simulation.

**Features:**
- Analog weight devices
- Drift and noise modeling
- In-situ training
- Hardware-aware optimization

**Integration:**
```python
from aihwkit.nn import AnalogLinear

# Define analog layer with FeFET-like device
linear = AnalogLinear(in_features, out_features, bias=True,
                     rpu_config=MyFeFETConfig())

# Train with hardware effects
loss.backward()
optimizer.step()
```

**Use Case:** Hardware-aware neural network training in PyTorch

---

## Circuit Simulation

### 1. ngspice + Verilog-A

**URL:** https://ngspice.sourceforge.io/

**Integration:**
```bash
# Export SPICE netlist from this module
array.ExportSPICE("crossbar.sp")

# Simulate with ngspice
ngspice crossbar.sp
```

**Netlist Example:**
```spice
* Crossbar array netlist
.subckt crossbar_cell wl bl gnd
R_cell wl bl {1/conductance}
.ends

.subckt crossbar_array
X_0_0 wl[0] bl[0] gnd crossbar_cell conductance=50u
X_0_1 wl[0] bl[1] gnd crossbar_cell conductance=75u
...
.ends

* Transient analysis
.tran 1n 1u
.print tran v(bl[0]) v(bl[1])
```

---

### 2. Xyce (Sandia)

**URL:** https://github.com/Xyce/Xyce

**Description:** High-performance parallel circuit simulator.

**Advantages:**
- Better convergence for large arrays
- Parallel simulation
- Advanced analysis (noise, sensitivity)

---

## Neural Network Mapping Tools

### 1. DNNWeaver

**URL:** https://github.com/georgia-tech-synergy-lab/dnnweaver

**Description:** Tool for mapping DNNs to custom accelerators.

**Use Case:** Map trained models to crossbar-based architectures

---

### 2. SCALE-Sim

**URL:** https://github.com/ARM-software/SCALE-Sim

**Description:** Systolic array simulator for DNNs.

**Comparison:** SCALE-Sim for digital systolic arrays; this module for analog crossbars

---

## Data Analysis Tools

### 1. NumPy/SciPy (Python)

**Use Case:** Post-processing exported data

```python
import numpy as np
import matplotlib.pyplot as plt

# Load conductance matrix
G = np.load('conductances.npy')

# Analyze distribution
plt.hist(G.flatten(), bins=30)
plt.xlabel('Conductance (µS)')
plt.ylabel('Count')
plt.show()

# Compute statistics
mean_G = G.mean()
std_G = G.std()
print(f"Mean: {mean_G:.2f}, Std: {std_G:.2f}")
```

---

### 2. Pandas (Python)

**Use Case:** CSV data analysis

```python
import pandas as pd

# Load weight matrix CSV
df = pd.read_csv('weights.csv')

# Group by level
level_stats = df.groupby('level').agg({
    'conductance_uS': ['mean', 'std', 'count']
})

print(level_stats)
```

---

## Visualization Tools

### 1. Matplotlib/Seaborn (Python)

```python
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

# Load heatmap data
G = np.load('conductances.npy')

# Visualize as heatmap
sns.heatmap(G, cmap='viridis', cbar_kws={'label': 'Conductance (µS)'})
plt.title('Crossbar Conductance Heatmap')
plt.xlabel('Column')
plt.ylabel('Row')
plt.show()
```

---

### 2. This Module's GUI (Fyne)

**Built-in visualization:**
- Conductance heatmap
- IR drop voltage map
- Sneak current visualization
- Real-time MVM animation
- Error metrics dashboard

**No external tools needed for basic visualization**

---

## Integration Guide

### Exporting from This Module

#### 1. NumPy Export (for Python)

```go
// Export weights
weights := network.GetWeights()
weights.ExportNumPy("weights.npy")

// Export MVM results
result, _ := array.MVMWithNonIdealities(input, opts)
// Save to JSON for analysis
```

#### 2. CSV Export (for spreadsheets)

```go
// Export conductance matrix
array.ExportWeightsCSV("conductances.csv")

// Format: row, col, level, conductance_norm, conductance_uS
```

#### 3. JSON Export (for reporting)

```go
// Export full analysis
array.ExportAnalysisJSON("analysis.json", mvmResult)

// Includes: array stats, non-ideality metrics, energy
```

#### 4. SPICE Export (for circuit sim)

```go
// Generate SPICE netlist
array.ExportSPICE("crossbar.sp")

// Includes: subcircuits, resistance values, analysis commands
```

---

### Importing to External Tools

#### CrossSim

```python
import numpy as np
import simulator  # CrossSim

# Load from this module
G = np.load('conductances.npy')

# Create CrossSim array
core = simulator.CrossbarCore(G)
output = core.vmm(input_vector)
```

#### NeuroSim

```cpp
#include "formula.h"
#include "Array.h"

// Load CSV
Array array(rows, cols);
array.LoadConductances("conductances.csv");

// Run simulation
array.CalculateLatency();
array.CalculateEnergy();
```

#### AIHWKIT

```python
from aihwkit.nn import AnalogLinear
import torch

# Load weights
weights_np = np.load('weights.npy')
weights_torch = torch.from_numpy(weights_np)

# Create analog layer
layer = AnalogLinear(in_features, out_features)
layer.weight.data = weights_torch
```

---

## Tool Comparison Matrix

| Tool | Purpose | Language | License | Use Case |
|------|---------|----------|---------|----------|
| **This Module** | Full crossbar sim | Go | Open | Education, visualization |
| **CrossSim** | Array-level | Python | Open | Validation reference |
| **BadCrossbar** | IR drop analysis | Python | MIT | Physics verification |
| **NeuroSim** | Architecture estimation | C++ | Open | Energy/area benchmarks |
| **MNSIM** | Multi-level NN | Python | Open | Multi-bit optimization |
| **AIHWKIT** | Analog training | Python | Apache 2.0 | PyTorch integration |
| **ngspice** | Circuit sim | C | BSD | Detailed circuit analysis |
| **Xyce** | Parallel circuit sim | C++ | GPL | Large-scale simulation |

---

## Best Practices

### 1. Model Validation

**Always cross-validate with multiple tools:**

1. Export weights from this module
2. Run same MVM in CrossSim
3. Compare outputs (should match within 1-2%)
4. If divergence > 5%, investigate:
   - Non-ideality configurations
   - Quantization differences
   - Conductance model differences

### 2. Energy Estimation

**Use NeuroSim for architecture-level estimates:**

1. Export array size and conductance distribution
2. NeuroSim provides detailed peripheral circuits
3. Compare with this module's simplified energy model
4. This module: educational baseline
5. NeuroSim: production-accurate estimates

### 3. Circuit-Level Verification

**Use ngspice/Xyce for critical paths:**

1. Export SPICE netlist
2. Add realistic parasitics (C_line, L_wire)
3. Run transient analysis
4. Verify timing, settling behavior
5. Feed back worst-case delays to this module

---

## Example Workflows

### Workflow 1: Python Analysis Pipeline

```python
# 1. Load data from this module
import numpy as np
import pandas as pd

weights = np.load('weights.npy')
results_df = pd.read_csv('mvm_results.csv')

# 2. Compute statistics
accuracy_loss = results_df['accuracy_loss'].mean()
energy_efficiency = results_df['energy_efficiency'].mean()

# 3. Visualize
import matplotlib.pyplot as plt
plt.figure(figsize=(10, 6))
plt.subplot(1, 2, 1)
plt.hist(weights.flatten(), bins=30)
plt.title('Weight Distribution')

plt.subplot(1, 2, 2)
plt.scatter(results_df['rmse'], results_df['accuracy_loss'])
plt.xlabel('RMSE')
plt.ylabel('Accuracy Loss (%)')
plt.title('Error vs Accuracy Impact')
plt.tight_layout()
plt.show()
```

### Workflow 2: CrossSim Validation

```python
import numpy as np
import simulator  # CrossSim

# Load from this module
G_ours = np.load('conductances.npy')
input_vec = np.load('input.npy')
output_ours = np.load('output_ours.npy')

# Run in CrossSim
core = simulator.CrossbarCore(G_ours)
output_crosssim = core.vmm(input_vec)

# Compare
rmse = np.sqrt(np.mean((output_ours - output_crosssim)**2))
print(f"RMSE between simulators: {rmse:.6f}")
```

### Workflow 3: SPICE Detailed Analysis

```bash
# 1. Export from this module
./fecim-lattice-tools crossbar --export-spice crossbar.sp

# 2. Add parasitics and analysis
cat >> crossbar.sp <<EOF
* Add wire parasitics
.param C_line=1f
.param L_wire=0.1n

* Transient analysis
.tran 10p 100n

* Measure settling time
.measure tran tsettle WHEN v(bl[0])=0.99*v_final CROSS=LAST
EOF

# 3. Run ngspice
ngspice -b crossbar.sp -o results.log

# 4. Extract metrics
grep "tsettle" results.log
```

---

## Research Opportunities

### 1. Device Calibration

**Goal:** Replace assumed parameters with measured data

**Tools needed:**
- Probe station data
- Parameter extraction scripts
- This module's lookup table conductance model

**Workflow:**
1. Measure G(V) curves from real FeFET devices
2. Extract 30-point conductance table
3. Load into `Config.ConductanceTable`
4. Compare simulation vs measurement

### 2. Advanced Drift Modeling

**Goal:** Validate drift coefficients with long-term retention tests

**Tools needed:**
- Temperature-controlled chamber
- This module's drift simulator
- Time-series analysis tools

**Workflow:**
1. Program array to known states
2. Bake at elevated temperature
3. Periodically read back conductances
4. Fit drift model: `G(t) = G₀ × (t/t₀)^(-ν)`
5. Update `drift.go` with measured ν

### 3. Multi-Tool Workflow

**Goal:** Complete device-to-system analysis

**Pipeline:**
```
This Module (array behavior)
    ↓ Export weights.npy
NeuroSim (energy/area)
    ↓ Export architecture.json
System-level power model
    ↓ Export netlist
SPICE (circuit verification)
```

---

## Documentation References

### Internal
- `docs/development/SCRIPT_REFERENCE.md#demo-2-crossbar`
- `docs/crossbar/reference/ARCHITECTURE.md`
- `docs/crossbar/reference/API.md`

### External
- CrossSim documentation: https://cross-sim.sandia.gov/
- NeuroSim tutorial: https://github.com/neurosim/DNN_NeuroSim_V1.3
- AIHWKIT docs: https://aihwkit.readthedocs.io/

---

## Summary

### Best Tools by Use Case

| Use Case | Recommended Tool | Why |
|----------|------------------|-----|
| **Educational demo** | This module | Built-in GUI, real-time |
| **Python prototyping** | CrossSim | Mature, well-tested |
| **Circuit verification** | ngspice | Industry standard |
| **Energy estimation** | NeuroSim | Comprehensive peripherals |
| **PyTorch training** | AIHWKIT | Native PyTorch integration |
| **IR drop analysis** | BadCrossbar | Specialized, accurate |

---

## See Also

- **[Features](./features.md)** - Module capabilities
- **[Physics](./physics.md)** - Simulation equations
- **[Architecture](./architecture.md)** - Code structure
- **Module 3 Tools** - MNIST neural network integration

---

**Last Updated:** 2026-02-16
