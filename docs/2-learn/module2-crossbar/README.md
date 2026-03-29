# Module 2: Crossbar Array Simulation

**Navigation:** [← Back to Learn](../README.md) | [ELI5](./eli5.md) | [Physics](./physics.md) | [Features](./features.md) | [Architecture](./architecture.md)

---

## Overview

Module 2 implements a physics-accurate crossbar array simulator for analog matrix-vector multiplication (MVM) in compute-in-memory (CIM) systems. It models 30-level discrete conductance states with comprehensive non-ideality simulation including IR drop, sneak paths, device variation, drift, and temperature effects.

**Key Concept:** A crossbar array is a grid of programmable resistors (or conductors) that naturally performs matrix multiplication in hardware. Apply voltages to rows, measure currents from columns—physics does the math in parallel.

---

## Quick Links

### For Beginners
- **[ELI5 Explanation](./eli5.md)** - Start here for crossbar array intuition

### For Developers
- **[Physics Reference](./physics.md)** - MVM equations, non-idealities, conductance models
- **[Features](./features.md)** - What the module does, workflows, extension points
- **[Architecture](./architecture.md)** - Code structure, types, data flow (comprehensive)

### For Researchers
- **[Open-Source Tools](./tools.md)** - Integration with CrossSim, NeuroSim, SPICE

---

## Module Contents

```
module2-crossbar/
├── pkg/crossbar/         # Core array simulation (4798 lines)
│   ├── array.go          # Array definition, MVM, quantization
│   ├── nonidealities.go  # Enhanced MVM with all non-idealities
│   ├── irdrop.go         # Wire resistance and IR drop
│   ├── sneakpath.go      # Parasitic current paths
│   ├── drift.go          # Conductance drift over time
│   ├── temperature.go    # Temperature-dependent physics
│   ├── enhanced.go       # Differential arrays, write-verify
│   └── reference.go      # Reference implementations
├── pkg/network/          # Multi-layer neural networks
├── pkg/training/         # Hardware-aware backpropagation
├── pkg/weights/          # Weight serialization (JSON/NumPy/CSV)
└── pkg/gui/              # Fyne visualization
```

---

## Quick Start

### GUI Mode (Default)
```bash
fecim-lattice-tools crossbar
```

### Headless MVM
```bash
fecim-lattice-tools --mode crossbar --size 64x64 --noise 0.05
```

### Python Integration
```python
# Export weights to NumPy format
# Load in Python for analysis
import numpy as np
weights = np.load('crossbar_weights.npy')
```

---

## What You'll Learn

1. **Crossbar Basics**
   - How voltage × conductance = current implements multiplication
   - Why Kirchhoff's current law = summation (accumulate)
   - MVM in hardware vs software

2. **Non-Idealities**
   - IR drop from wire resistance
   - Sneak paths through unselected cells
   - Device-to-device variation
   - Conductance drift over time
   - Temperature effects

3. **Architectures**
   - **0T1R (passive):** Simple but sneak-prone
   - **1T1R (1 transistor):** Sneak isolation, standard
   - **2T1R (2 transistors):** Full isolation, premium

4. **Conductance Models**
   - Linear (simple, fast)
   - Exponential (physics-accurate)
   - Lookup table (measurement-calibrated)

---

## Key Features

- **Analog MVM/VMM** with 30-level quantization (simulation baseline)
- **Comprehensive non-idealities:**
  - IR drop (wire resistance)
  - Sneak paths (0T1R vs 1T1R)
  - Process variation (device noise, gradients, edge effects)
  - Drift (time-dependent degradation)
  - Temperature (77K to 400K)
- **Differential arrays** for signed weights [-1, +1]
- **Write-verify programming** with convergence tracking
- **Multi-layer networks** with hardware-aware training
- **GPU MVM acceleration** (optional, via config)
- **Extensive visualization:** Heatmaps, IR drop maps, sneak current analysis
- **Export formats:** JSON, CSV, NumPy, SPICE netlists

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | Plain-language crossbar intro | Beginners |
| [physics.md](./physics.md) | MVM equations, non-idealities | Developers, researchers |
| [features.md](./features.md) | Feature list, workflows | Developers |
| [architecture.md](./architecture.md) | Code structure, types, data flow | Developers |
| [tools.md](./tools.md) | Open-source ecosystem | Researchers |

---

## Core Concepts

### Matrix-Vector Multiply (MVM)

```
Input vector:  v = [v₁, v₂, v₃]
Weight matrix: G = [[g₁₁, g₁₂, g₁₃],
                    [g₂₁, g₂₂, g₂₃]]

Output: I = G × v

Hardware implementation:
- Apply v to columns (word lines)
- Measure currents from rows (bit lines)
- Each cell contributes: I_ij = G_ij × V_j
- Each row sums automatically (Kirchhoff's law)
```

### 30 Discrete Levels

Conductance is quantized to 30 states:

```
G(level) = G_min + (level / 29) × (G_max - G_min)

where:
  level ∈ {0, 1, 2, ..., 29}
  G_min = 10 µS (OFF state)
  G_max = 100 µS (ON state)
```

This provides ~4.9 bits per cell (log₂(30) ≈ 4.9).

---

## Architecture Comparison

| Architecture | Transistors/Cell | Sneak Paths | IR Drop | Density | Use Case |
|--------------|------------------|-------------|---------|---------|----------|
| **0T1R** | 0 (passive) | High (full sneak) | High | Very high | Research |
| **1T1R** | 1 | Low (1000× isolation) | Medium | Moderate | Production |
| **2T1R** | 2 | None | Low | Lower | Premium applications |

See [architecture.md](./architecture.md) for detailed voltage schemes.

---

## Evidence Status

- **Demonstrated:** Repository structure, code paths, and simulation behaviors are verifiable from source/tests
- **Modeled:** Equations, defaults, and performance estimates are simulator models unless explicitly tied to cited measured data
- **Aspirational:** Production-scale or silicon-parity claims are roadmap intent and must not be reported as demonstrated results

See [Physics Audit](../../4-research/honesty-audit.md) for complete accuracy policy.

---

## Related Modules

- **[Module 1: Hysteresis](../module1-hysteresis/README.md)** - Provides conductance-polarization relationship
- **[Module 3: MNIST](../module3-mnist/README.md)** - Uses crossbar arrays for neural network inference
- **[Module 4: Circuits](../module4-circuits/README.md)** - DAC/ADC peripherals for crossbar I/O
- **[Module 6: EDA](../module6-eda/README.md)** - Layout generation for crossbar arrays

---

## Integration Examples

### With Module 1 (Hysteresis)
```go
// Map polarization level to conductance
level := hysteresis.GetDiscreteLevel()  // 0-29
conductance := crossbar.LevelToConductance(level, Gmin, Gmax)
crossbar.ProgramWeight(row, col, conductance)
```

### With Module 3 (MNIST)
```go
// Use crossbar for neural network layer
layer := network.NewLayer(inputSize, outputSize)
layer.LoadWeights(mnistWeights)
output := layer.Forward(input)  // Uses crossbar.MVM internally
```

### With Python (NumPy export)
```go
// Export for external analysis
weights.ExportNumPy("weights.npy")
```

```python
import numpy as np
weights = np.load("weights.npy")
# Analyze in Python
```

---

## Performance Characteristics

### Computational Complexity

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Basic MVM | O(rows × cols) | Linear combination |
| MVM + IR drop | O(rows × cols) | Analytical model |
| Sneak paths (≤32×32) | O(rows² × cols²) | Full enumeration |
| Sneak paths (>32×32) | O(rows × cols) | Simplified model |

### Energy Efficiency

Typical 128×128 array, 6-bit ADC:
- **FeCIM Total:** ~0.67 pJ per MVM
- **GPU Equivalent:** ~163 pJ per MVM
- **Efficiency:** ~240× better than GPU

*Note: These are model-based estimates, not measured hardware.*

---

## Testing

```bash
# Run all crossbar tests
go test ./module2-crossbar/pkg/crossbar

# Test specific functionality
go test -run TestArrayMVM ./module2-crossbar/pkg/crossbar
go test -run TestIRDrop ./module2-crossbar/pkg/crossbar
go test -run TestSneakPaths ./module2-crossbar/pkg/crossbar

# With coverage
go test -cover ./module2-crossbar/pkg/crossbar
```

---

## Source Code

- **GitHub:** [module2-crossbar/](../../../module2-crossbar/)
- **Core array:** `pkg/crossbar/array.go`
- **Non-idealities:** `pkg/crossbar/nonidealities.go`
- **GUI:** `pkg/gui/app.go`
- **Tests:** `pkg/crossbar/*_test.go`

---

## Next Steps

- **New to crossbars?** → Start with [ELI5](./eli5.md)
- **Need equations?** → See [Physics](./physics.md)
- **Want to extend?** → Read [Architecture](./architecture.md)
- **Researching?** → Check [Tools](./tools.md)

---

**Last Updated:** 2026-02-16
**Maintainer:** FeCIM Lattice Tools Project
