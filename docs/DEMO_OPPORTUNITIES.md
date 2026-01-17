# IronLattice Demo Opportunities Analysis

Analysis of research to identify Vulkan/Go visualization/simulation tool opportunities. Demos ordered by complexity (easiest first).

---

## Key Research Insights

### From Downloaded Papers

| Paper | Key Opportunity | Vulkan/Go Fit |
|-------|-----------------|---------------|
| **FerroX** (arXiv:2210.15668) | GPU phase-field simulation, TDGL solver | ⭐⭐⭐ Perfect for compute shaders |
| **Physical Reality of Preisach** (Nature 2018) | Domain switching visualization | ⭐⭐⭐ 2D/3D hysteresis plotting |
| **Multi-Level FeFET Crossbar** (Nature 2023) | Crossbar array visualization | ⭐⭐⭐ Matrix visualization |
| **Sneak Path in Self-Rectifying Arrays** (Frontiers 2022) | Crossbar non-idealities animation | ⭐⭐⭐ Current flow visualization |
| **COMPASS Compiler** (arXiv 2025) | Weight mapping visualization | ⭐⭐ Algorithm animation |
| **HCiM: ADC-Less Hybrid CIM** (arXiv 2024) | Mixed CIM architecture demo | ⭐⭐ Architecture comparison |

### Visualization Opportunities by Category

```
┌────────────────────────────────────────────────────────────────┐
│                    VISUALIZATION OPPORTUNITIES                  │
├──────────────────┬─────────────────┬───────────────────────────┤
│     EASIEST      │     MEDIUM      │        ADVANCED           │
├──────────────────┼─────────────────┼───────────────────────────┤
│ P-E Curve Plot   │ Crossbar MVM    │ GPU Phase-Field (TDGL)    │
│ Domain Heatmap   │ Sneak Path Anim │ Neural Network Training   │
│ Material Compare │ Weight Mapping  │ 3D Domain Evolution       │
│ State Histogram  │ ADC/DAC Quant   │ Compiler Dataflow         │
└──────────────────┴─────────────────┴───────────────────────────┘
```

---

## Proposed Demo Redefinition

### Demo 1: Hysteresis P-E Curve Visualizer (EASIEST)

**Complexity:** ⭐ Beginner-friendly  
**Timeline:** 1-2 weeks  
**Vulkan Use:** Graphics pipeline only (2D line plotting)

#### Features
1. **Real-time P-E curve plotting** - Draw hysteresis loop as voltage changes
2. **Material selector** - Compare HfO₂, ZrO₂, HZO superlattice parameters
3. **Discrete state display** - Show 30 analog states as bars/histogram
4. **Interactive voltage slider** - Mouse/keyboard control of applied field

#### Technical Approach
- **No compute shaders needed** - CPU Preisach model (already implemented!)
- Simple vertex/fragment shaders for line drawing
- GLFW window with basic UI
- Color gradient for polarization state

#### Why This is Easiest
- Preisach model already works in Go (`preisach.go`)
- 2D graphics only (no 3D complexity)
- No GPU physics simulation
- Simple data (100-200 points for curve)

```
┌─────────────────────────────────────────────────┐
│  P-E Hysteresis Curve        [HZO Superlattice] │
│                                                 │
│       P                                         │
│       ↑     ╭──────╮                           │
│       │    ╱        ╲                          │
│       │   /          \                         │
│  ─────┼──●────────────●───→ E                  │
│       │   \          /                         │
│       │    ╲        ╱                          │
│       │     ╰──────╯                           │
│                                                 │
│  Voltage: ████████░░░░░░░  [+2.5V]             │
│  State:   14/30  [▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░]  │
└─────────────────────────────────────────────────┘
```

---

### Demo 2: Crossbar Array Visualizer (MEDIUM)

**Complexity:** ⭐⭐ Intermediate  
**Timeline:** 2-3 weeks  
**Vulkan Use:** Compute + Graphics (parallel MVM calculation)

#### Features
1. **Matrix-vector multiply animation** - Watch currents flow through crossbar
2. **Non-ideality toggles** - IR drop, sneak paths, device variation
3. **Weight programming** - Click cells to set conductance values
4. **Input pulse animation** - Watch voltages propagate row-by-row

#### Technical Approach
- **Compute shader for MVM** - Parallel O(1) multiply-accumulate
- 2D grid visualization (8×8 or 16×16 array)
- Color = conductance (blue→red gradient)
- Arrows/lines for current flow animation

#### Why This is Medium
- Requires compute shader for parallel MVM
- More complex state management
- Animation timing for current flow
- Multiple interacting parameters

```
┌─────────────────────────────────────────────────┐
│  8×8 Crossbar Array           [Non-idealities] │
│                                                 │
│     V₀ V₁ V₂ V₃ V₄ V₅ V₆ V₇                    │
│    ┌──┬──┬──┬──┬──┬──┬──┬──┐                   │
│ I₀ │▓▓│░░│▓░│░▓│▓▓│░░│▓░│░▓│→ Σ               │
│ I₁ │░▓│▓░│▓▓│░░│░▓│▓░│░░│▓▓│→ Σ               │
│ I₂ │▓░│░░│░▓│▓▓│░░│▓▓│▓░│░▓│→ Σ               │
│ ...│  │  │  │  │  │  │  │  │                   │
│    └──┴──┴──┴──┴──┴──┴──┴──┘                   │
│                                                 │
│  [✓] IR Drop  [✓] Sneak Path  [ ] Variation    │
│  Input: [1,0,1,1,0,1,0,1]  Output: [0.7,0.3,..]│
└─────────────────────────────────────────────────┘
```

---

### Demo 3: GPU Phase-Field Domain Simulator (ADVANCED)

**Complexity:** ⭐⭐⭐ Expert level  
**Timeline:** 3-4 weeks  
**Vulkan Use:** Heavy compute (TDGL solver on GPU)

#### Features
1. **Real-time domain evolution** - Watch ferroelectric domains form/switch
2. **GPU-accelerated TDGL** - Time-dependent Ginzburg-Landau solver
3. **3D polarization field** - Volume rendering of P vector field
4. **Parameter sweeps** - Animate temperature, field, composition changes

#### Technical Approach
- **Compute shaders for TDGL** - FerroX-inspired approach (arXiv:2210.15668)
- 3D texture for polarization field (128³ or 256³)
- Volume rendering or isosurface extraction
- Real-time time evolution (100+ FPS target)

#### Why This is Advanced
- Complex physics (coupled PDEs)
- 3D data structures and rendering
- Numerical stability considerations
- Large memory bandwidth requirements

```
┌─────────────────────────────────────────────────┐
│  GPU Phase-Field Simulator     [T=300K, E=1MV] │
│                                                 │
│         ┌───────────────┐                       │
│        /▓▓▓▓▓░░░░░░▓▓▓▓/│                      │
│       /▓▓▓▓░░░░░░░░▓▓▓/ │  ▓ = P↑ domain      │
│      /░░░░░░░░░░░░░░░/  │  ░ = P↓ domain      │
│     /░░░░▓▓▓▓▓▓░░░░░/   │                      │
│    /▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓/    │  Time: 0.5 ns       │
│   └───────────────┘     │  Step: 5000         │
│                                                 │
│  α: 1.72e6  β: -2.5e9  γ: 1.5e11  E: ██░░ MV  │
│  [Play] [Pause] [Step] [Reset]                 │
└─────────────────────────────────────────────────┘
```

---

## Comparison with Original Demos

| Aspect | Original Demo 1 | New Demo 1 |
|--------|-----------------|------------|
| **Focus** | Vulkan compute shaders | 2D plotting with existing physics |
| **Complexity** | High (Vulkan setup) | Low (minimal Vulkan) |
| **Physics** | GPU Preisach | CPU Preisach (already works) |
| **Timeline** | 4 weeks | 1-2 weeks |

| Aspect | Original Demo 2 | New Demo 2 |
|--------|-----------------|------------|
| **Focus** | MVM inference | Same + visualization |
| **Vulkan** | Compute shaders | First real compute work |
| **Visual** | Minimal | Full crossbar animation |

| Aspect | Original Demo 3 | New Demo 3 |
|--------|-----------------|------------|
| **Focus** | MNIST | GPU physics simulation |
| **Vulkan** | Network inference | Phase-field TDGL |
| **Impact** | Accuracy demo | Scientific visualization |

---

## Recommended Implementation Order

### Phase 1: Demo 1 Foundation (Week 1-2)
1. GLFW window setup with Vulkan surface
2. Basic graphics pipeline (vertex/fragment only)
3. Line drawing utilities
4. CPU Preisach model integration
5. Real-time P-E curve plotting

### Phase 2: Demo 2 Crossbar (Week 3-4)
1. Add Vulkan compute pipeline
2. Implement MVM in compute shader
3. Grid visualization
4. Current flow animation
5. Non-ideality parameters

### Phase 3: Demo 3 TDGL (Week 5-8)
1. Study FerroX paper in detail
2. Implement 3D storage buffers
3. TDGL compute shader
4. Volume rendering
5. Interactive parameter control

---

## Technology Stack Decision

| Component | Choice | Rationale |
|-----------|--------|-----------|
| **Vulkan bindings** | `go-vk` | Active, full API coverage |
| **Windowing** | GLFW via `go-gl` | Standard, proven |
| **Physics model** | Existing Go code | Preisach already implemented |
| **Shader language** | GLSL → SPIR-V | Standard Vulkan workflow |
| **Math library** | gonum | Already in go.mod |

---

## Key Takeaways

1. **Demo 1 should NOT require compute shaders** - Use existing CPU physics
2. **Demo 2 introduces compute** - Natural progression
3. **Demo 3 is research-grade** - FerroX-level GPU simulation
4. **All demos share** - GLFW window, Vulkan instance, graphics pipeline basics

This progression lets you:
- Get something working quickly (Demo 1 in 1-2 weeks)
- Learn Vulkan incrementally
- Build on success rather than tackling everything at once
