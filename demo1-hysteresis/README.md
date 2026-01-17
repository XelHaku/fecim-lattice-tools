# Demo 1: P-E Hysteresis Curve Visualizer

**Complexity:** ⭐ Beginner (Graphics only)  
**Timeline:** 1-2 weeks  
**Status:** In Development

## Goal

Interactive visualization of ferroelectric P-E hysteresis curve with:
- Real-time curve plotting as voltage changes
- Material parameter comparison (HfO₂, ZrO₂, HZO superlattice)
- 30 discrete analog state demonstration
- Interactive voltage slider control

## Architecture

```
demo1-hysteresis/
├── cmd/hysteresis/main.go     # Entry point
├── pkg/
│   ├── ferroelectric/         # Physics (CPU-based)
│   │   ├── preisach.go        # ✅ Already implemented
│   │   └── material.go        # ✅ HZO parameters
│   ├── render/                # Vulkan graphics
│   │   └── render.go          # Graphics pipeline
│   └── simulation/            # Time-stepping
│       └── engine.go
└── shaders/
    ├── hysteresis.vert        # Vertex shader
    └── hysteresis.frag        # Fragment shader
```

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Physics engine | CPU (Go) | Preisach model already works |
| Vulkan compute | **Not needed** | Simple enough for CPU |
| Rendering | 2D line graphics | P-E curve + state bars |
| Windowing | GLFW | Standard approach |

## Physics Model

Uses the **Preisach hysteresis model** with:
- Hyperbolic tangent switching function
- Minor loop support via turning point history
- HZO material parameters (Ps=25μC/cm², Ec=1.0MV/cm)

## Visualization

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

## Implementation Phases

- [x] Phase 1: Core Physics - Preisach model ✅
- [ ] Phase 2: GLFW Window + Vulkan Surface
- [ ] Phase 3: Graphics Pipeline (line rendering)
- [ ] Phase 4: Interactive UI (voltage slider)

## Dependencies

```go
require (
    github.com/bbredesen/go-vk    // Vulkan bindings
    github.com/go-gl/glfw/v3.3/glfw // Windowing
    gonum.org/v1/gonum            // Math library
)
```

## Run

```bash
cd demo1-hysteresis
go run cmd/hysteresis/main.go
# or with headless mode
go run cmd/hysteresis/main.go --headless
```
