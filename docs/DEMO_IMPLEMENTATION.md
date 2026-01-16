# Demo Implementation Roadmap

## Overview

This document outlines the implementation strategy for the three IronLattice visualization demos.

---

## Demo 1: Ferroelectric Hysteresis Visualizer

### Goal
Interactive visualization of a single ferroelectric memory cell showing P-E hysteresis curve and domain switching.

### Physics Models

#### Option A: Preisach Model (Recommended for Demo 1)
- **Pros:** Computationally efficient, captures minor loops, well-documented
- **Cons:** Phenomenological (no domain visualization)
- **Implementation:**
  - Hyperbolic tangent basis function (Bo Jiang method)
  - LIFO structure for turning points
  - 2D distribution of coercive voltages

#### Option B: Landau-Khalatnikov (Full Physics)
- **Pros:** First-principles, domain dynamics
- **Cons:** Computationally expensive, requires GPU
- **Implementation:**
  - Implicit finite difference scheme
  - Iterative solver
  - 2D/3D grid discretization

### Architecture

```
demo1-hysteresis/
├── cmd/
│   └── hysteresis/
│       └── main.go           # Entry point
├── pkg/
│   ├── ferroelectric/
│   │   ├── preisach.go       # Preisach model
│   │   ├── landau.go         # L-K equations
│   │   └── material.go       # HZO parameters
│   ├── simulation/
│   │   ├── engine.go         # Time-stepping
│   │   └── state.go          # Polarization state
│   └── vulkan/
│       ├── renderer.go       # Graphics pipeline
│       ├── compute.go        # Compute shaders
│       └── window.go         # GLFW integration
└── shaders/
    ├── hysteresis.comp       # Compute shader
    ├── cell.vert             # Vertex shader
    ├── cell.frag             # Fragment shader
    └── compile.sh
```

### Implementation Steps

1. **Phase 1: Core Physics (Week 1)**
   - [ ] Implement Preisach model in Go
   - [ ] Define HZO material parameters (Pᵣ, Eᶜ)
   - [ ] Unit tests for hysteresis loop generation

2. **Phase 2: Vulkan Setup (Week 2)**
   - [ ] Initialize Vulkan via go-vk or vgpu
   - [ ] Create window with GLFW
   - [ ] Setup graphics pipeline for cell rendering

3. **Phase 3: Visualization (Week 3)**
   - [ ] Real-time P-E curve plotting
   - [ ] Cell color = polarization state
   - [ ] Interactive voltage control (slider)

4. **Phase 4: Polish (Week 4)**
   - [ ] 30 discrete state demonstration
   - [ ] Minor loop visualization
   - [ ] UI improvements

### Key Parameters (HZO)

| Parameter | Symbol | Value | Unit |
|-----------|--------|-------|------|
| Remanent Polarization | Pᵣ | 25 | μC/cm² |
| Coercive Field | Eᶜ | 1.5 | MV/cm |
| Film Thickness | t | 10 | nm |
| Saturation Polarization | Pₛ | 30 | μC/cm² |

### Preisach Model Pseudocode

```go
type PreisachModel struct {
    alpha, beta float64  // Distribution parameters
    turningPoints []float64  // LIFO stack
    Ps float64  // Saturation polarization
}

func (p *PreisachModel) Polarization(E float64) float64 {
    // Hyperbolic tangent switching function
    P := p.Ps * math.Tanh((E - p.Ec) / p.delta)

    // Apply history-dependent correction
    P = p.applyHistory(P, E)
    return P
}
```

---

## Demo 2: Crossbar Array MVM

### Goal
Visualize Matrix-Vector Multiplication in a ferroelectric crossbar array.

### MNIST Accuracy Benchmarks (Literature)

| Architecture | Accuracy | Source |
|--------------|----------|--------|
| 24×24 FE Memristor | 98.78% | [ScienceDirect 2025](https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963) |
| Multi-Level FeFET 28nm | 96.6% | [Nature Comms 2023](https://www.nature.com/articles/s41467-023-42110-y) |
| Ferroelectric Tunnel Junction | 92% | [SemiEngineering 2024](https://semiengineering.com/ferroelectric-tunnel-junctions-in-crossbar-array-analog-in-memory-compute-accelerators/) |
| IronLattice Target | 87% | Dr. Tour's presentation |

### Non-Idealities to Model

| Effect | Impact | Mitigation |
|--------|--------|------------|
| **IR Drop** | Voltage variation along wires | Compensation circuits |
| **Sneak Path** | Current leakage in passive arrays | Self-rectifying devices |
| **Conductance Variation** | D2D and C2C variability | Noise-aware training |
| **ADC Quantization** | Limited precision | Low-precision algorithms |

### Architecture

```
demo2-crossbar/
├── cmd/
│   └── crossbar/
│       └── main.go
├── pkg/
│   ├── array/
│   │   ├── crossbar.go       # Array structure
│   │   ├── cell.go           # Individual FeFET
│   │   └── wire.go           # Wire resistance
│   ├── compute/
│   │   ├── mvm.go            # Matrix-vector multiply
│   │   └── nonideal.go       # Non-ideality models
│   └── vulkan/
│       └── ...
└── shaders/
    ├── mvm.comp              # MVM compute shader
    └── array.frag            # Array visualization
```

### MVM Visualization Concept

```
Input Vector (Voltage)
    V₁   V₂   V₃
    ↓    ↓    ↓
   ┌────┬────┬────┐
   │G₁₁ │G₁₂ │G₁₃ │→ I₁
   ├────┼────┼────┤
   │G₂₁ │G₂₂ │G₂₃ │→ I₂
   ├────┼────┼────┤
   │G₃₁ │G₃₂ │G₃₃ │→ I₃
   └────┴────┴────┘
          ↓
   Output Vector (Current)

Color: Conductance value (analog weight)
Animation: Current flow during MVM
```

### Implementation Steps

1. **Phase 1: Array Model**
   - [ ] Define crossbar topology
   - [ ] Implement Ohm's Law multiplication
   - [ ] Implement Kirchhoff's summation

2. **Phase 2: Non-Idealities**
   - [ ] Add wire resistance (IR drop)
   - [ ] Conductance variation model
   - [ ] Sneak path simulation (optional)

3. **Phase 3: Visualization**
   - [ ] 2D array rendering
   - [ ] Current flow animation
   - [ ] Input/output vector display

4. **Phase 4: Demo Integration**
   - [ ] Load pre-trained MNIST weights
   - [ ] Interactive digit input
   - [ ] Classification output

---

## Demo 3: MNIST on CIM

### Goal
Handwritten digit recognition on simulated IronLattice hardware.

### Architecture

```
demo3-mnist/
├── cmd/
│   └── mnist/
│       └── main.go
├── pkg/
│   ├── network/
│   │   ├── layer.go          # Neural network layers
│   │   ├── weights.go        # Weight loading
│   │   └── inference.go      # Forward pass
│   ├── cim/
│   │   ├── accelerator.go    # CIM simulation
│   │   └── quantization.go   # INT4/INT8 support
│   └── vulkan/
│       └── ...
├── models/
│   └── mnist_weights.bin     # Pre-trained weights
└── shaders/
    └── inference.comp        # Inference compute shader
```

### Neural Network Architecture

```
Input: 28×28 = 784 pixels
       ↓
FC Layer 1: 784 → 256 (ReLU)
       ↓
FC Layer 2: 256 → 64 (ReLU)
       ↓
FC Layer 3: 64 → 10 (Softmax)
       ↓
Output: 10 classes (digits 0-9)
```

### Noise-Aware Training

To achieve 87% accuracy matching IronLattice benchmarks:

```python
# During training (PyTorch)
def add_hardware_noise(weights, sigma=0.05):
    noise = torch.randn_like(weights) * sigma
    return weights + noise

# Apply during forward pass
noisy_weights = add_hardware_noise(layer.weight)
output = F.linear(input, noisy_weights)
```

### Implementation Steps

1. **Phase 1: Network Definition**
   - [ ] Define FC layers
   - [ ] Implement forward pass
   - [ ] Weight quantization (INT8)

2. **Phase 2: CIM Simulation**
   - [ ] Map weights to crossbar arrays
   - [ ] Simulate analog MVM
   - [ ] Add realistic noise

3. **Phase 3: Visualization**
   - [ ] Draw input digit
   - [ ] Show layer activations
   - [ ] Display confidence scores

4. **Phase 4: Interactive Demo**
   - [ ] Mouse/touch digit input
   - [ ] Real-time inference
   - [ ] Accuracy statistics

---

## Shared Components

### Go + Vulkan Setup

```go
// Using vgpu (recommended)
import "cogentcore.org/core/vgpu"

func initVulkan() *vgpu.GPU {
    gpu := vgpu.NewGPU()
    gpu.Config("IronLattice Visualizer")
    return gpu
}
```

### GLSL Shader Template

```glsl
#version 450

layout(local_size_x = 256) in;

layout(std140, binding = 0) buffer StateBuffer {
    float polarization[];
};

layout(std140, binding = 1) uniform Params {
    float voltage;
    float dt;
};

void main() {
    uint idx = gl_GlobalInvocationID.x;

    // Physics update
    float P = polarization[idx];
    float dP = computePolarizationChange(P, voltage, dt);
    polarization[idx] = P + dP;
}
```

---

## Dependencies

| Package | Purpose | Installation |
|---------|---------|-------------|
| go-vk | Vulkan bindings | `go get github.com/bbredesen/go-vk` |
| vgpu | High-level Vulkan | `go get cogentcore.org/core/vgpu` |
| glfw | Window management | `go get github.com/go-gl/glfw/v3.3/glfw` |
| gonum | Math operations | `go get gonum.org/v1/gonum` |

---

## Timeline

| Phase | Demo 1 | Demo 2 | Demo 3 |
|-------|--------|--------|--------|
| Core Physics | ✓ | | |
| Vulkan Setup | ✓ | | |
| Basic Visualization | ✓ | | |
| Polish | ✓ | | |
| Array Model | | ✓ | |
| Non-Idealities | | ✓ | |
| MVM Visualization | | ✓ | |
| Network Implementation | | | ✓ |
| CIM Simulation | | | ✓ |
| Interactive Demo | | | ✓ |

---

## References

1. [Preisach Model Paper](https://juser.fz-juelich.de/record/36264/files/4398.pdf)
2. [Vulkan Compute Tutorial](https://vulkan-tutorial.com/Compute_Shader)
3. [FerroX GPU Framework](https://arxiv.org/abs/2210.15668)
4. [Multi-Level FeFET CIM](https://www.nature.com/articles/s41467-023-42110-y)
5. [Crossbar Non-Idealities](https://link.springer.com/article/10.1007/s11432-024-4240-x)
