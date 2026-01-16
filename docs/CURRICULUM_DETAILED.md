# IronLattice Curriculum - Detailed Technical Guide

This document provides in-depth technical details for each curriculum area, with implementation guidance for the visualization demos.

---

## Area 1: Solid-State Physics - HfO₂/ZrO₂ Ferroelectrics

### 1.1 Crystal Structure and Phase Transitions

| Phase | Space Group | Structure | Electrical Property | Stability |
|-------|-------------|-----------|---------------------|-----------|
| Monoclinic (m) | P2₁/c | Centrosymmetric | Dielectric | Stable (RT) |
| Tetragonal (t) | P4₂/nmc | High-symmetry | Antiferroelectric | High T / doped |
| Cubic (c) | Fm3̄m | Fluorite | Paraelectric | Very high T |
| **Orthorhombic (o)** | **Pca2₁** | **Non-centrosymmetric** | **Ferroelectric** | **Metastable** |

### 1.2 Stabilization Mechanisms

**Grain Size Effect:**
```
Surface Energy: γ_ortho < γ_mono
Total Gibbs Free Energy: G = G_bulk + G_surface
For nanoscale grains (<10nm): G_ortho < G_mono
```

**Zr Doping:**
- ZrO₂ promotes tetragonal → stabilizes orthorhombic on cooling
- Optimal Hf:Zr ratio ≈ 1:1 for maximum Pᵣ

**Mechanical Confinement:**
- TiN capping prevents volume expansion during annealing
- Stress inhibits monoclinic transformation

### 1.3 Simulation Parameters

```go
// Crystal parameters for HZO
type CrystalParams struct {
    LatticeA     float64 = 5.01e-10  // m (a-axis)
    LatticeB     float64 = 5.24e-10  // m (b-axis)
    LatticeC     float64 = 5.07e-10  // m (c-axis)
    UnitCellVol  float64 = 1.33e-28  // m³
    DipoleMoment float64 = 0.5e-29   // C·m per unit cell
}
```

---

## Area 2: Ferroelectric Hysteresis Models

### 2.1 Landau-Khalatnikov Equation

The time-dependent polarization dynamics:

```
γ · dP/dt = -∂F/∂P + E_ext
```

Where:
- `γ` = viscosity coefficient (damping)
- `F` = Landau free energy functional
- `E_ext` = external electric field

**Free Energy (6th order expansion):**
```
F = α·P² + β·P⁴ + γ·P⁶ - E·P
```

**Implementation (Go):**
```go
func LandauKhalatnikov(P, E, dt float64, params *LKParams) float64 {
    // dF/dP = 2αP + 4βP³ + 6γP⁵ - E
    dFdP := 2*params.Alpha*P + 4*params.Beta*math.Pow(P,3) +
            6*params.Gamma*math.Pow(P,5) - E

    // Euler integration
    dPdt := -dFdP / params.Viscosity
    return P + dPdt*dt
}
```

**Parameters for HZO:**
| Parameter | Value | Unit |
|-----------|-------|------|
| α (T > Tc) | > 0 | J·m/C² |
| α (T < Tc) | < 0 | J·m/C² |
| β | < 0 (1st order) | J·m⁵/C⁴ |
| γ | > 0 | J·m⁹/C⁶ |
| viscosity | 10⁻⁸ - 10⁻⁶ | Ω·m |

### 2.2 Preisach Model

Decomposes hysteresis into elementary hysterons:

```
P(E) = ∫∫ μ(α,β) · γ_αβ[E] dα dβ
```

Where:
- `μ(α,β)` = Preisach distribution function
- `γ_αβ` = elementary hysteron (rectangular loop)
- `α` = switching-up threshold
- `β` = switching-down threshold

**Gaussian Distribution:**
```go
func PreisachDistribution(alpha, beta float64, params *PreisachParams) float64 {
    Ec := params.EcMean
    sigma := params.EcSigma

    // Lorentzian distribution centered at coercive field
    dist := 1.0 / (math.Pi * sigma * (1 + math.Pow((alpha-Ec)/sigma, 2)))
    return dist * params.Ps
}
```

**Advantages:**
- Captures minor loops and history dependence
- Computationally efficient for circuit simulation
- 5 parameters describe full behavior

### 2.3 Phase-Field (TDGL) Simulation

For domain visualization, solve the Time-Dependent Ginzburg-Landau equation:

```
∂P/∂t = -L · δF/δP

F = ∫[f_landau(P) + κ/2·|∇P|² + f_elastic + f_electrostatic] dV
```

**Gradient Energy (Domain Wall Cost):**
```
f_gradient = κ/2 · (∂P/∂x)² + (∂P/∂y)² + (∂P/∂z)²)
```

**Finite Difference Scheme:**
```go
func TDGL_Step(P [][]float64, E float64, dx, dt float64, params *TDGLParams) {
    L := params.Kinetic
    kappa := params.Gradient

    for i := 1; i < len(P)-1; i++ {
        for j := 1; j < len(P[0])-1; j++ {
            // Landau term
            dF_landau := LandauDerivative(P[i][j], E, params)

            // Gradient term (Laplacian)
            laplacian := (P[i+1][j] + P[i-1][j] + P[i][j+1] + P[i][j-1] - 4*P[i][j]) / (dx*dx)
            dF_gradient := -kappa * laplacian

            // Update
            P[i][j] += -L * (dF_landau + dF_gradient) * dt
        }
    }
}
```

---

## Area 3: Crossbar Array Architecture

### 3.1 Matrix-Vector Multiplication

**Physical Implementation:**
```
           V₁    V₂    V₃    V₄   (Input Voltages)
            ↓     ↓     ↓     ↓
        ┌──[G₁₁]─[G₁₂]─[G₁₃]─[G₁₄]──→ I₁ = Σ(Vⱼ·G₁ⱼ)
        │
BL₁ ────┤──[G₂₁]─[G₂₂]─[G₂₃]─[G₂₄]──→ I₂ = Σ(Vⱼ·G₂ⱼ)
        │
BL₂ ────┤──[G₃₁]─[G₃₂]─[G₃₃]─[G₃₄]──→ I₃ = Σ(Vⱼ·G₃ⱼ)
        │
BL₃ ────┴──[G₄₁]─[G₄₂]─[G₄₃]─[G₄₄]──→ I₄ = Σ(Vⱼ·G₄ⱼ)
```

**Operations:**
1. **Ohm's Law (Multiplication):** I_ij = V_i × G_ij
2. **Kirchhoff's Law (Summation):** I_out,j = Σᵢ I_ij

### 3.2 Non-Idealities

| Non-Ideality | Cause | Impact | Mitigation |
|--------------|-------|--------|------------|
| **Sneak Path** | Current through unselected cells | Read errors | Self-rectifying devices, selectors |
| **IR Drop** | Wire resistance | Accuracy degradation | Larger wires, smaller arrays |
| **D2D Variation** | Manufacturing | Weight errors | Calibration, noise-aware training |
| **Stuck-at-Fault** | Device defects | Dead cells | Redundancy, remapping |

**Sneak Path Current:**
```
For NxN array without selector:
I_sneak ∝ N² × G_off × V_read

With self-rectifying ratio R = I_on/I_off:
Max array size ≈ √(R × read_margin)
```

**IR Drop Model:**
```go
func ComputeIRDrop(array [][]float64, wireR float64, V_in []float64) []float64 {
    N := len(array)
    V_effective := make([][]float64, N)

    // Wire resistance per segment
    for i := 0; i < N; i++ {
        V_drop := 0.0
        for j := 0; j < N; j++ {
            // Current through this segment affects downstream cells
            I_segment := (V_in[j] - V_drop) * array[i][j]
            V_drop += I_segment * wireR
            V_effective[i][j] = V_in[j] - V_drop
        }
    }
    return computeOutputCurrents(array, V_effective)
}
```

### 3.3 Weight Mapping

**Differential Pair for Signed Weights:**
```
W_ij = G⁺_ij - G⁻_ij

Where:
- G⁺ stores positive component
- G⁻ stores negative component (reference)

For weight W in range [-W_max, +W_max]:
G⁺ = G_mid + W × ΔG/2
G⁻ = G_mid - W × ΔG/2
```

**Mapping Algorithm:**
```go
func MapWeights(W [][]float64, Gmin, Gmax float64) (Gplus, Gminus [][]float64) {
    Gmid := (Gmax + Gmin) / 2
    deltaG := Gmax - Gmin

    // Find weight range
    Wmax := findAbsMax(W)
    scale := deltaG / (2 * Wmax)

    for i := range W {
        for j := range W[i] {
            Gplus[i][j] = Gmid + W[i][j]*scale
            Gminus[i][j] = Gmid - W[i][j]*scale
        }
    }
    return
}
```

---

## Area 4: ADC/DAC Design for CIM

### 4.1 Precision-Energy Tradeoff

**ADC Energy Scaling:**
```
E_ADC ∝ 2^ENOB × f_sample
```

Where ENOB = Effective Number of Bits

**Typical CIM Requirements:**
| Precision | Bits | Energy (pJ) | Area (μm²) |
|-----------|------|-------------|------------|
| Low | 4 | 0.1-1 | 100-500 |
| Medium | 6-8 | 1-10 | 500-2000 |
| High | 10+ | 10-100 | 2000+ |

### 4.2 ADC-Less Approaches

**Binary/Ternary Quantization:**
- Eliminate ADC entirely
- Trade precision for energy
- Use noise-aware training to maintain accuracy

**HCiM Architecture:**
```
Analog CIM → Binary Quantizer → Digital Accumulator
        ↓
   (No ADC needed)
```

**Energy Savings:** Up to 28% vs 7-bit ADC baseline

### 4.3 Adaptive Quantization

**Non-uniform Thresholds:**
```go
func AdaptiveQuantize(analog float64, thresholds []float64) int {
    for i, thresh := range thresholds {
        if analog < thresh {
            return i
        }
    }
    return len(thresholds)
}

// Thresholds optimized for output distribution
// Rather than uniform spacing
```

---

## Area 5: Neural Network Training

### 5.1 Noise-Aware Training (IBM AIHWKit)

**Inject Hardware Noise During Training:**
```python
from aihwkit.nn import AnalogLinear
from aihwkit.simulator.configs import InferenceRPUConfig

config = InferenceRPUConfig()
config.forward.noise_management = "AbsMax"
config.forward.out_noise = 0.06  # 6% output noise

model = AnalogLinear(784, 128, rpu_config=config)
```

**Noise Sources Modeled:**
- Programming noise (write variability)
- Read noise (thermal, shot noise)
- Drift (PCM: resistance drift over time)
- ADC quantization noise

### 5.2 Weight Update Linearity

**Requirements for On-Chip Training:**
| Metric | Requirement | Best Achieved |
|--------|-------------|---------------|
| Linearity (α) | Close to 1 | 0.45-0.73 |
| Symmetry | α_p ≈ α_d | 0.89 |
| States | ≥32 (5-bit) | >128 |
| Gmax/Gmin | >10 | >10⁵ |

**Superlattice Advantage:**
FE/DE/FE superlattice structure moderates domain switching, providing:
- Smoother conductance transitions
- Better linearity than solid-solution HZO
- Reduced abrupt switching

### 5.3 Quantization-Aware Training

```python
# Quantize weights to match hardware precision
def quantize_weights(W, bits=4):
    Wmax = torch.max(torch.abs(W))
    scale = (2**(bits-1) - 1) / Wmax
    W_q = torch.round(W * scale) / scale
    return W_q

# Use straight-through estimator for gradients
class QuantizedLinear(nn.Module):
    def forward(self, x):
        W_q = quantize_weights(self.weight, self.bits)
        return F.linear(x, W_q, self.bias)
```

---

## Area 6: GPU Simulation with Vulkan

### 6.1 Compute Shader Architecture

**GLSL Compute Shader for TDGL:**
```glsl
#version 450

layout(local_size_x = 16, local_size_y = 16) in;

layout(binding = 0) buffer PolarizationIn {
    float P_in[];
};

layout(binding = 1) buffer PolarizationOut {
    float P_out[];
};

layout(push_constant) uniform Params {
    float E;      // External field
    float dt;     // Time step
    float dx;     // Spatial step
    float alpha;  // Landau coefficient
    float kappa;  // Gradient coefficient
    float L;      // Kinetic coefficient
    int width;
    int height;
};

void main() {
    ivec2 pos = ivec2(gl_GlobalInvocationID.xy);
    if (pos.x >= width || pos.y >= height) return;

    int idx = pos.y * width + pos.x;
    float P = P_in[idx];

    // Landau derivative: dF/dP = 2αP + 4βP³ - E
    float dF_landau = 2.0*alpha*P + 4.0*alpha*P*P*P - E;

    // Laplacian (5-point stencil)
    float laplacian = 0.0;
    if (pos.x > 0) laplacian += P_in[idx-1];
    if (pos.x < width-1) laplacian += P_in[idx+1];
    if (pos.y > 0) laplacian += P_in[idx-width];
    if (pos.y < height-1) laplacian += P_in[idx+width];
    laplacian = (laplacian - 4.0*P) / (dx*dx);

    float dF_gradient = -kappa * laplacian;

    // TDGL update
    P_out[idx] = P - L * (dF_landau + dF_gradient) * dt;
}
```

### 6.2 Vulkan Pipeline Setup (Go)

```go
// Using go-vk or vulkan-go bindings
func CreateComputePipeline(device vk.Device, shader []byte) vk.Pipeline {
    // Create shader module
    shaderModule := createShaderModule(device, shader)

    // Pipeline layout with push constants
    pushConstantRange := vk.PushConstantRange{
        StageFlags: vk.ShaderStageComputeBit,
        Offset:     0,
        Size:       32, // Params struct size
    }

    layoutInfo := vk.PipelineLayoutCreateInfo{
        PushConstantRangeCount: 1,
        PPushConstantRanges:    &pushConstantRange,
    }

    // Create compute pipeline
    pipelineInfo := vk.ComputePipelineCreateInfo{
        Stage: vk.PipelineShaderStageCreateInfo{
            Stage:  vk.ShaderStageComputeBit,
            Module: shaderModule,
            PName:  "main",
        },
        Layout: pipelineLayout,
    }

    return vk.CreateComputePipeline(device, pipelineInfo)
}
```

### 6.3 Memory Management

**Double Buffering for Time Stepping:**
```go
type SimulationBuffers struct {
    PolarizationA vk.Buffer  // Current state
    PolarizationB vk.Buffer  // Next state
    current       int        // 0 or 1
}

func (s *SimulationBuffers) Swap() {
    s.current = 1 - s.current
}

func (s *SimulationBuffers) Input() vk.Buffer {
    if s.current == 0 {
        return s.PolarizationA
    }
    return s.PolarizationB
}
```

---

## Area 7: Visualization Techniques

### 7.1 Domain Coloring

**Polarization to Color Mapping:**
```go
func PolarizationToColor(P, Pmax float64) (r, g, b float32) {
    // Normalize to [-1, 1]
    normalized := P / Pmax

    // Divergent colormap (blue-white-red)
    if normalized > 0 {
        r = float32(normalized)
        g = float32(1 - normalized)
        b = float32(1 - normalized)
    } else {
        r = float32(1 + normalized)
        g = float32(1 + normalized)
        b = float32(-normalized)
    }
    return
}
```

### 7.2 Hysteresis Curve Rendering

**Real-Time P-E Curve:**
```go
type HysteresisTracer struct {
    Points []vec2
    MaxPoints int
}

func (h *HysteresisTracer) AddPoint(E, P float64) {
    h.Points = append(h.Points, vec2{float32(E), float32(P)})
    if len(h.Points) > h.MaxPoints {
        h.Points = h.Points[1:]
    }
}

func (h *HysteresisTracer) Render(renderer *Renderer) {
    // Draw as line strip with fading alpha
    for i, pt := range h.Points {
        alpha := float32(i) / float32(len(h.Points))
        renderer.DrawPoint(pt.x, pt.y, alpha)
    }
}
```

### 7.3 Domain Wall Visualization

**Isosurface Extraction (P = 0):**
```go
// Marching squares for 2D domain wall lines
func ExtractDomainWalls(P [][]float64, threshold float64) []Line {
    var walls []Line

    for i := 0; i < len(P)-1; i++ {
        for j := 0; j < len(P[0])-1; j++ {
            // Check if threshold crosses this cell
            case := 0
            if P[i][j] > threshold { case |= 1 }
            if P[i+1][j] > threshold { case |= 2 }
            if P[i+1][j+1] > threshold { case |= 4 }
            if P[i][j+1] > threshold { case |= 8 }

            // Generate line segments based on case
            walls = append(walls, marchingSquaresLookup[case]...)
        }
    }
    return walls
}
```

---

## Area 8: Demo Implementations

### Demo 1: Hysteresis Visualizer

**Architecture:**
```
┌─────────────────────────────────────────────────────────┐
│                    Main Window                          │
├─────────────────────┬───────────────────────────────────┤
│                     │                                   │
│   Ferroelectric     │      P-E Hysteresis Curve        │
│   Cell Display      │                                   │
│   (Color = P)       │      ┌────────────────────────┐  │
│                     │      │         P              │  │
│   ┌───────────────┐ │      │         ↑    +Pᵣ      │  │
│   │               │ │      │    ─────┼────╮        │  │
│   │   Domain      │ │      │    ─────┼────┼──→ E   │  │
│   │   Structure   │ │      │         ╰────╯        │  │
│   │               │ │      │              -Pᵣ      │  │
│   └───────────────┘ │      └────────────────────────┘  │
│                     │                                   │
├─────────────────────┴───────────────────────────────────┤
│  [◀ -V] ═══════════●═══════════ [+V ▶]  V = 0.0 V     │
│                                                         │
│  Model: [Preisach ▼]  Speed: [1x ▼]  [Reset] [Export]  │
└─────────────────────────────────────────────────────────┘
```

**Features:**
- Interactive voltage slider
- Real-time P-E curve tracing
- Domain coloring visualization
- Model selection (Preisach, L-K, TDGL)
- 30 discrete state demonstration

### Demo 2: Crossbar MVM

**Architecture:**
```
┌─────────────────────────────────────────────────────────┐
│                  Crossbar Array (8×8)                   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Input Vector V                                         │
│  [0.5][0.8][0.2][0.9][0.1][0.7][0.4][0.6]              │
│    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓               │
│  ┌────────────────────────────────────────┐             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₁ │             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₂ │             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₃ │             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₄ │             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₅ │             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₆ │             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₇ │             │
│  │ [G] [G] [G] [G] [G] [G] [G] [G] ──→ I₈ │             │
│  └────────────────────────────────────────┘             │
│                                                         │
│  Output Current: [2.1][1.8][3.2][0.9][2.5][1.4][2.8][1.1]│
│                                                         │
│  [Load Weights] [Random Input] [Show IR Drop] [Animate] │
└─────────────────────────────────────────────────────────┘
```

### Demo 3: MNIST Inference

**Network Architecture:**
```
Input (28×28=784) → FC Layer (128) → ReLU → FC Layer (10) → Softmax
                         ↓                        ↓
                   Crossbar 1              Crossbar 2
                   (784×128)               (128×10)
```

**Target Metrics:**
- Accuracy: 87% (matching IronLattice claims)
- Weight precision: 4-5 bits
- Noise injection: Gaussian σ = 0.05

---

## References

### Simulation Frameworks
1. FerroX: https://github.com/AMReX-Microelectronics/FerroX
2. IBM AIHWKit: https://github.com/IBM/aihwkit
3. Preisach Verilog-A: https://github.com/DavidTobar456/pfecapRevision

### Key Papers
1. BEOL Superlattice FeFET: https://ieeexplore.ieee.org/document/9691825/
2. Multi-level FeFET Crossbar: https://www.nature.com/articles/s41467-023-42110-y
3. HCiM ADC-Less CIM: https://arxiv.org/abs/2403.13577
4. Landau-Khalatnikov: https://www.sciencedirect.com/science/article/abs/pii/S1007570420303543
5. Preisach Model: https://www.nature.com/articles/s41467-018-06717-w

### Vulkan Resources
1. Vulkan Tutorial Compute: https://vulkan-tutorial.com/Compute_Shader
2. go-vk: https://github.com/bbredesen/go-vk
3. Datoviz (Vulkan visualization): https://datoviz.org/

---

*Last updated: January 2026*
