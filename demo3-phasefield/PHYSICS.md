# Phase-Field Simulation Physics: From Absolute Basics

Start here if you've never studied phase-field methods or domain dynamics before.

---

## Part 1: The Problem We're Solving

### What Happens INSIDE the Ferroelectric?

Demo 1 showed the P-E hysteresis loop (the macroscopic behavior).
Demo 2 showed crossbar arrays (the device level).
Demo 3 goes DEEPER—into the material itself to see **domains**.

```
Zoom levels:

Demo 1: One cell's P-E curve     Demo 2: Array of cells     Demo 3: Inside one grain
┌──────────────────┐             ┌─────────────────┐         ┌─────────────────┐
│       P          │             │ ● ● ● ● ● ● ● │         │▓▓▓░░░░░▓▓▓▓▓▓▓│
│       │  ╭──╮    │             │ ● ● ● ● ● ● ● │         │▓▓░░░░░░░░▓▓▓▓▓│
│       │ ╱    ╲   │             │ ● ● ● ● ● ● ● │         │░░░░░░░░░░░░░▓▓│
│       ├─●────●─→E│             │ ● ● ● ● ● ● ● │         │░░░░▓▓▓▓░░░░░░░│
│        ╲    ╱    │             └─────────────────┘         └─────────────────┘
│         ╰──╯     │              Conductance grid          Polarization domains
└──────────────────┘                                        (▓=P↑, ░=P↓)
```

### What are Domains?

A ferroelectric crystal doesn't have uniform polarization. It breaks into **domains** with different polarization directions:

```
Real ferroelectric crystal:

┌────────────────────────────────────┐
│▓▓▓▓▓▓▓│░░░░░░░│▓▓▓▓▓│░░░░░░░░░░░░│
│▓▓▓▓▓▓▓│░░░░░░░│▓▓▓▓▓│░░░░░░░░░░░░│  ← Domain 1 (P points UP ▓)
│▓▓▓▓▓▓▓│░░░░░░░│▓▓▓▓▓│░░░░░░░░░░░░│  ← Domain 2 (P points DOWN ░)
│▓▓▓▓▓▓▓│░░░░░░░│▓▓▓▓▓│░░░░░░░░░░░░│
└────────────────────────────────────┘
         ↑           ↑
    Domain walls (boundaries between domains)
```

**Why domains exist:** They minimize the total energy of the system. A single large domain creates large electric fields at the surface; multiple smaller domains cancel these out.

### Why Simulate Domains?

Understanding domains helps with:
- **Device design:** How fast can domains switch?
- **Reliability:** Where do domains nucleate (start)?
- **Optimization:** What geometry minimizes switching voltage?
- **Debugging:** Why does this device fail after N cycles?

---

## Part 2: What is Phase-Field?

### The Continuous Order Parameter

Instead of tracking individual atoms, we describe the material with a **field** that varies continuously in space:

```
Discrete (atomic):                Continuous (phase-field):

⊕ ⊖ ⊕ ⊖ ⊕ ⊖ ⊕ ⊖ ⊕ ⊖            P(x,y,z) = smooth function
↑ ↓ ↑ ↓ ↑ ↓ ↑ ↓ ↑ ↓
Each atom tracked                 One number at each point
(10²³ atoms = impossible!)        (manageable on GPU!)
```

**The phase-field variable:** P(x,y,z,t) = polarization at position (x,y,z) at time t

- P > 0: Polarization points UP (one phase)
- P < 0: Polarization points DOWN (other phase)
- P ≈ 0: Domain wall (transition region)

### The Domain Wall

The transition between domains isn't sharp—it has a finite width:

```
Polarization profile across a domain wall:

    P
    ↑
+Ps ├────────╮
    │        │╲
    │        │ ╲
  0 ├────────┼──╲────────────
    │        │   ╲
    │        │    ╲
-Ps ├────────╯     ╰─────────
    └──────────────────────→ x
             ↑
        Domain wall
        (width ~ 1-5 nm)
```

---

## Part 3: Energy Minimization

### The Free Energy Functional

The material wants to minimize its total energy. We write this as:

```
F = ∫∫∫ f(P, ∇P, E) dV

Total energy = sum over all space of local energy density
```

The local energy has three parts:

### 1. Landau Energy (Bulk)

Energy due to the polarization value itself:

```
f_Landau = α·P² + β·P⁴ + γ·P⁶

    Energy
      ↑
      │    ╱╲      ╱╲
      │   ╱  ╲    ╱  ╲
      │  ╱    ╲  ╱    ╲
      │ ╱      ╲╱      ╲
      ├─●──────────────●─→ P
      │-Ps     0      +Ps
          ↑            ↑
     Energy minima (stable states)
```

**Key insight:** Two stable states at ±Ps (the memory!), with an energy barrier in between.

| Coefficient | Name | Meaning |
|-------------|------|---------|
| α | Linear | Temperature-dependent; changes sign at Curie point |
| β | Quartic | Usually negative (creates double well) |
| γ | Sixth-order | Stabilizes the potential |

### 2. Gradient Energy (Domain Walls)

Energy cost of having polarization CHANGE in space:

```
f_gradient = κ·|∇P|²

Where ∇P = (∂P/∂x, ∂P/∂y, ∂P/∂z)
```

**Meaning:** Domain walls cost energy! The material "wants" uniform polarization, but other forces create domains anyway.

**κ (kappa):** Gradient coefficient. Larger κ = thicker, more energetic domain walls.

### 3. Electric Energy (External Field)

Energy from applied electric field:

```
f_electric = -E·P

Negative because: field WANTS to align polarization with it
```

**Meaning:** Applied field tilts the energy landscape, making one state more favorable.

---

## Part 4: Time Evolution - TDGL

### Time-Dependent Ginzburg-Landau Equation

How does polarization evolve over time? The material relaxes toward lower energy:

```
∂P/∂t = -L · δF/δP

Rate of change = -L × (energy gradient)
```

**In plain English:** Polarization changes in the direction that reduces total energy, at rate proportional to L.

### Expanding the Equation

The full TDGL equation for ferroelectrics:

```
∂P/∂t = -L · [2αP + 4βP³ + 6γP⁵ - κ∇²P - E]

         ↑         ↑            ↑      ↑
      Landau    Higher       Gradient  Electric
      terms     order Landau  term     field
```

| Term | Physical Meaning |
|------|------------------|
| 2αP | Restoring force toward equilibrium |
| 4βP³ | Non-linear term creating bistability |
| 6γP⁵ | Saturation (prevents runaway) |
| -κ∇²P | Smoothing (penalizes sharp gradients) |
| -E | External field driving |

### L - The Kinetic Coefficient

L controls how fast the system evolves:
- Large L: Fast dynamics, quick equilibration
- Small L: Slow dynamics, viscous relaxation

For HfO₂: L ~ 10⁻³ m³/(V·s·C)

---

## Part 5: Numerical Solution

### Discretization

We can't solve TDGL analytically, so we discretize:

```
Continuous:                      Discrete (3D grid):
                                 
P(x, y, z, t)                    P[i][j][k][n]
                                 
x ∈ [0, Lx]                      i = 0, 1, 2, ... Nx-1
y ∈ [0, Ly]                      j = 0, 1, 2, ... Ny-1
z ∈ [0, Lz]                      k = 0, 1, 2, ... Nz-1
t ∈ [0, T]                       n = 0, 1, 2, ... Nt-1
```

**Grid spacing:** dx, dy, dz (typically 0.5-1 nm)
**Time step:** dt (typically ~picoseconds)

### Finite Difference for Laplacian

The ∇²P term (second derivative) in discrete form:

```
∇²P ≈ (P[i+1] - 2P[i] + P[i-1]) / dx²
                    +
      (P[j+1] - 2P[j] + P[j-1]) / dy²
                    +
      (P[k+1] - 2P[k] + P[k-1]) / dz²
```

**In words:** Laplacian = How different is this cell from its neighbors?

### Time Stepping (Euler Method)

Simplest approach - explicit Euler:

```
P_new = P_old + dt · (-L · δF/δP)

Each time step:
1. Calculate δF/δP at all grid points
2. Update P everywhere
3. Repeat
```

**GPU parallelism:** Each grid point can be updated independently → perfect for Vulkan compute shaders!

---

## Part 6: What the Simulation Shows

### Domain Nucleation

When field is applied, domains don't switch everywhere at once. They **nucleate** (start) at defects:

```
t = 0 ns              t = 0.1 ns            t = 0.5 ns
┌────────────────┐    ┌────────────────┐    ┌────────────────┐
│▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│    │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│    │░░░░░░░░▓▓▓▓▓▓▓▓│
│▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│    │▓▓▓▓▓▓░░▓▓▓▓▓▓▓▓│    │░░░░░░░░░░░▓▓▓▓▓│
│▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│ →  │▓▓▓▓░░░░░░▓▓▓▓▓▓│ →  │░░░░░░░░░░░░░░░░│
│▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│    │▓▓▓▓▓░░░▓▓▓▓▓▓▓▓│    │░░░░░░░░░░░░▓▓▓▓│
└────────────────┘    └────────────────┘    └────────────────┘
All P↑                 New domain nucleates  Domain grows, walls move
                       at defect site
```

### Domain Wall Motion

Once nucleated, domain walls move:

```
Wall velocity v ∝ (E - E_threshold)

Slow field:                    Fast field:
wall creeps                    wall zips across
│░░░░│▓▓▓▓│ → │░░░░░│▓▓▓│     │░░░░│▓▓▓▓│ → │░░░░░░░░│
     ↑ slow                          ↑ fast
```

### What Parameters Control

| Parameter | Effect |
|-----------|--------|
| α (Landau) | Changes with temperature; α < 0 = ferroelectric |
| β (Landau) | Depth of double well |
| κ (gradient) | Domain wall width and energy |
| E (field) | Drives switching |
| L (kinetic) | Speed of dynamics |

---

## Part 7: GPU Acceleration

### Why GPU?

Each grid point's update is independent:

```
CPU (sequential):           GPU (parallel):
                           
Point 1 →                  Point 1 ──┐
Point 2 →                  Point 2 ──┼──→ All at once!
Point 3 →                  Point 3 ──┤
   ⋮ (one at a time)       Point 4 ──┘
Point N →                     ⋮
                           Point N ──┘
```

For a 128×128×128 grid:
- CPU: 2 million sequential operations
- GPU: 2 million parallel operations → 1000× faster!

### Vulkan Compute Shader

The core TDGL update as a compute shader:

```glsl
// tdgl.comp
void main() {
    ivec3 pos = ivec3(gl_GlobalInvocationID);
    float P = polGrid[pos];
    
    // Calculate Laplacian (neighbor average)
    float lap = laplacian(pos);
    
    // TDGL right-hand side
    float dFdP = 2*alpha*P + 4*beta*P*P*P + 6*gamma*P*P*P*P*P
                 - kappa*lap - E_ext;
    
    // Euler time step
    polGrid_new[pos] = P - dt * L * dFdP;
}
```

**GPU dispatch:** One thread per grid point, 8×8×8 workgroups.

---

## Part 8: HZO Parameters

Typical values for HfO₂-ZrO₂:

| Parameter | Symbol | Value | Unit |
|-----------|--------|-------|------|
| Landau α | α | 1.72×10⁶ (at 300K) | V·m/C |
| Landau β | β | -2.5×10⁹ | V·m⁵/C³ |
| Landau γ | γ | 1.5×10¹¹ | V·m⁹/C⁵ |
| Gradient κ | κ | ~10⁻⁹ | V·m³/C |
| Kinetic L | L | ~10⁻³ | m³/(V·s·C) |
| Curie temp | Tc | ~400°C | K |

These parameters come from first-principles calculations and experimental fitting.

---

## Summary Table

| Term | Plain English |
|------|---------------|
| **Phase-field** | Describe material with smooth field P(x,y,z) instead of tracking atoms |
| **Domain** | Region with uniform polarization direction |
| **Domain wall** | Boundary between domains; has finite width (~1-5 nm) |
| **Free energy F** | Total energy the system tries to minimize |
| **Landau energy** | Energy from polarization value itself (creates two stable states) |
| **Gradient energy** | Energy cost of polarization changing in space |
| **TDGL** | Time-Dependent Ginzburg-Landau equation (how P evolves) |
| **∂P/∂t = -L·δF/δP** | "Move toward lower energy at rate L" |
| **∇²P** | Laplacian = "difference from neighbors" |
| **Nucleation** | New domain starts at a defect |
| **Domain wall motion** | Existing walls move when field applied |

---

## What Demo 3 Visualizes

1. **3D Domain Structure** - See regions of P↑ and P↓ in real-time
2. **Domain Wall Dynamics** - Watch walls move as field changes
3. **Nucleation Events** - See where new domains form
4. **Parameter Sweeps** - Change α, β, E and see effect on domains
5. **Time Evolution** - Animate the switching process

This is research-grade simulation—the same physics used by FerroX and academic groups!
