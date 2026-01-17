# Demo 3: GPU Phase-Field Domain Simulator

**Complexity:** ⭐⭐⭐ Advanced (Heavy Compute)  
**Timeline:** 3-4 weeks  
**Status:** Not Started

## Goal

Research-grade GPU-accelerated simulation of ferroelectric domain evolution using Time-Dependent Ginzburg-Landau (TDGL) equations. Inspired by the **FerroX framework** (arXiv:2210.15668).

Features:
- Real-time domain switching visualization
- 3D polarization field evolution
- GPU-accelerated TDGL solver
- Interactive parameter sweeps (temperature, field, composition)

## Physics Background

### Time-Dependent Ginzburg-Landau (TDGL)

The TDGL equation governs polarization dynamics:

```
∂P/∂t = -L * δF/δP

where F = ∫[α·P² + β·P⁴ + γ·P⁶ + κ·|∇P|² - E·P] dV
```

| Parameter | Symbol | HZO Value | Unit |
|-----------|--------|-----------|------|
| Landau α | α | 1.72×10⁶ | Vm/C |
| Landau β | β | -2.5×10⁹ | Vm⁵/C³ |
| Landau γ | γ | 1.5×10¹¹ | Vm⁹/C⁵ |
| Gradient κ | κ | ~10⁻⁹ | Vm³/C |
| Kinetic L | L | ~10⁻³ | m³/VsC |

### Domain Structure

```
┌───────────────────────────────────────┐
│  3D Domain Visualization              │
│                                       │
│      ┌───────────────┐                │
│     /▓▓▓▓▓░░░░░░▓▓▓▓/│   ▓ = P↑      │
│    /▓▓▓▓░░░░░░░░▓▓▓/ │   ░ = P↓      │
│   /░░░░░░░░░░░░░░░/  │               │
│  /░░░░▓▓▓▓▓▓░░░░░/   │  Domain walls │
│ /▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓/    │  = boundaries │
│└───────────────┘     │               │
│                                       │
│  Time: 0.5 ns   Step: 5000           │
└───────────────────────────────────────┘
```

## Architecture

```
demo3-phasefield/
├── cmd/phasefield/main.go     # Entry point
├── pkg/
│   ├── physics/               # TDGL equations
│   │   ├── tdgl.go            # Time integration
│   │   ├── landau.go          # Free energy
│   │   └── poisson.go         # Electrostatics (optional)
│   ├── vulkan/                # GPU compute
│   │   ├── compute.go         # Compute pipeline
│   │   ├── buffer3d.go        # 3D storage buffers
│   │   └── sync.go            # Synchronization
│   └── render/                # 3D visualization
│       ├── volume.go          # Volume rendering
│       └── isosurface.go      # Marching cubes (optional)
└── shaders/
    ├── tdgl.comp              # TDGL time step
    ├── gradient.comp          # ∇P calculation
    ├── volume.vert            # Volume rendering
    └── volume.frag            # Ray marching
```

## Vulkan Compute Architecture

### TDGL Compute Shader

```glsl
// tdgl.comp - Single time step
#version 450

layout(local_size_x = 8, local_size_y = 8, local_size_z = 8) in;

layout(set = 0, binding = 0) buffer Polarization { float P[]; };
layout(set = 0, binding = 1) buffer Gradient { vec3 gradP[]; };

layout(push_constant) uniform Params {
    float dt;       // Time step
    float alpha;    // Landau α
    float beta;     // Landau β  
    float gamma;    // Landau γ
    float kappa;    // Gradient coefficient
    float E_ext;    // External field
};

void main() {
    uvec3 pos = gl_GlobalInvocationID;
    uint idx = pos.z * Ny * Nx + pos.y * Nx + pos.x;
    
    float p = P[idx];
    vec3 grad = gradP[idx];
    
    // Free energy derivative: dF/dP
    float dF_dP = 2.0*alpha*p + 4.0*beta*p*p*p + 6.0*gamma*p*p*p*p*p
                  - kappa * laplacian(pos) - E_ext;
    
    // TDGL time evolution
    P[idx] = p - dt * L * dF_dP;
}
```

### GPU Memory Layout

| Buffer | Size | Description |
|--------|------|-------------|
| `P` | Nx × Ny × Nz × 4B | Polarization field |
| `gradP` | Nx × Ny × Nz × 12B | Gradient vectors |
| `scratch` | Nx × Ny × Nz × 4B | Temporary storage |

For 128³ grid: ~32 MB GPU memory

## Implementation Phases

- [ ] Phase 1: Vulkan 3D buffer management
- [ ] Phase 2: TDGL compute shader implementation
- [ ] Phase 3: Gradient/Laplacian calculation
- [ ] Phase 4: Volume rendering pipeline
- [ ] Phase 5: Interactive parameter control
- [ ] Phase 6: Performance optimization

## Key References

1. **FerroX** (arXiv:2210.15668) - GPU phase-field framework for ferroelectrics
2. **Physical Reality of Preisach Model** (Nature 2018) - Domain physics
3. **TDGL Algorithms** - Numerical methods for time integration

## Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| Grid size | 128³ - 256³ | Memory limited |
| Frame rate | 30+ FPS | Including visualization |
| Time step | ~1 ps | Numerical stability |
| GPU utilization | >80% | Bandwidth bound |

## Dependencies

```go
require (
    github.com/bbredesen/go-vk
    github.com/go-gl/glfw/v3.3/glfw
    gonum.org/v1/gonum
)
```

## Run

```bash
cd demo3-phasefield
go run cmd/phasefield/main.go
```

## Future Extensions

- Poisson equation solver (electrostatics coupling)
- Multi-component order parameter (HZO composition)
- Domain wall dynamics analysis
- Export to VTK for ParaView visualization
