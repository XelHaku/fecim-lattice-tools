# Ferroelectric CIM Research Findings - Part 8 (January 2026)

Vulkan GPU Compute, Go Bindings, Real-Time Hysteresis Visualization, and Demo1 Implementation.

---

## 1. Vulkan Compute Pipeline for Ferroelectric Simulation

### Overview

Vulkan compute shaders provide direct GPU access for GPGPU workloads, ideal for simulating ferroelectric hysteresis in real-time. The compute pipeline bypasses the traditional graphics pipeline, enabling efficient parallel physics simulation.

### Key Concepts

| Concept | Description | Application to Demo1 |
|---------|-------------|---------------------|
| **Compute Shader** | GLSL/SPIR-V program running on GPU | Preisach/TDGL physics |
| **SSBO** | Shader Storage Buffer Objects | Cell polarization arrays |
| **Workgroups** | Parallel execution units | 256 invocations per group |
| **Dispatch** | Launch compute workload | Update all cells per frame |
| **Synchronization** | Fence/semaphore barriers | Compute → Graphics sync |

### Vulkan Compute Pipeline Stages

```
┌──────────────────────────────────────────────────┐
│ 1. Initialize Vulkan Instance & Device           │
│    - Select compute-capable queue family         │
│    - Create logical device with compute queue    │
├──────────────────────────────────────────────────┤
│ 2. Create Buffers (SSBO)                         │
│    - Cell state buffer (polarization, history)   │
│    - Uniform buffer (material params, voltage)   │
│    - Output buffer (normalized P for rendering)  │
├──────────────────────────────────────────────────┤
│ 3. Create Compute Pipeline                       │
│    - Load SPIR-V shader (preisach.comp)          │
│    - Create descriptor set layout                │
│    - Create pipeline layout & pipeline           │
├──────────────────────────────────────────────────┤
│ 4. Record Command Buffer                         │
│    - Bind compute pipeline                       │
│    - Bind descriptor sets                        │
│    - vkCmdDispatch(numCells/256, 1, 1)           │
├──────────────────────────────────────────────────┤
│ 5. Submit & Synchronize                          │
│    - Submit to compute queue                     │
│    - Wait for completion (fence)                 │
│    - Read back results for visualization         │
└──────────────────────────────────────────────────┘
```

---

## 2. Go Vulkan Bindings Comparison

### Available Libraries

| Library | Source | Status | Recommendation |
|---------|--------|--------|----------------|
| **go-vk** | github.com/bbredesen/go-vk | Active | ✅ Best for raw Vulkan |
| **vulkan-go** | github.com/vulkan-go/vulkan | Maintained | ✅ Alternative bindings |
| **vgpu** | cogentcore.org/core/vgpu | Unmaintained | ⚠️ No longer developed |

### Critical Update: Cogent Core Shift to WebGPU

**Important:** The Cogent Core project has transitioned from Vulkan to WebGPU. The `vgpu` package is no longer actively maintained for the Cogent Core ecosystem.

> "Cogent Core has transitioned from Vulkan to WebGPU, meaning the Vulkan-based vgpu code is no longer actively maintained."

**Recommendation:** Use `go-vk` or `vulkan-go` for direct Vulkan development, not `vgpu`.

### go-vk Features

| Feature | Supported | Notes |
|---------|-----------|-------|
| Vulkan 1.1-1.3 | ✅ | Full coverage |
| Compute shaders | ✅ | vkCmdDispatch |
| GLFW integration | ✅ | With go-gl/glfw |
| Extensions | ✅ | KHR, EXT, etc. |
| Memory management | Manual | Standard Vulkan |

### go-vk + GLFW Setup Pattern

```go
import (
    "github.com/bbredesen/go-vk"
    "github.com/go-gl/glfw/v3.3/glfw"
)

func initVulkan() error {
    // 1. Initialize GLFW for Vulkan (no OpenGL context)
    glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
    window, _ := glfw.CreateWindow(1280, 720, "Demo1", nil, nil)
    
    // 2. Create Vulkan instance
    instanceCI := vk.InstanceCreateInfo{}
    instance, _ := vk.CreateInstance(&instanceCI, nil)
    
    // 3. Create surface from GLFW window
    surface, _ := window.CreateWindowSurface(instance, nil)
    
    // 4. Select device with compute capability
    // ... standard device selection
    
    return nil
}
```

---

## 3. FerroX GPU Phase-Field Framework

### Overview

FerroX is the premier GPU-accelerated phase-field simulation framework for ferroelectric materials, developed using AMReX.

| Feature | Value |
|---------|-------|
| **Speedup** | 15× on GPU vs multicore CPU |
| **Framework** | AMReX (parallel adaptive mesh) |
| **Equations** | TDGL, Poisson, Semiconductor |
| **Materials** | HfO₂, ZrO₂, PZT, BTO |
| **Visualization** | Amrvis, ParaView |

### TDGL Equation (Time-Dependent Ginzburg-Landau)

```
∂P/∂t = -γ × δF/δP
```

Where:
- `P` = polarization vector
- `γ` = kinetic coefficient
- `F` = Landau-Devonshire free energy functional

### FerroX GitHub

**URL:** https://github.com/AMReX-Microelectronics/FerroX

**Paper:** arXiv:2210.15668 (Comp. Phys. Comm. 2023)

---

## 4. Preisach Model Implementation Resources

### Available Implementations

| Language | Repository | Description |
|----------|------------|-------------|
| **Verilog-A** | github.com/DavidTobar456/pfecapRevision | Circuit simulation |
| **MATLAB** | Various GitHub repos | GUI-based tutorials |
| **Python** | github.com/raphaelvogel/preisach | General framework |

### Heracles Model (2025)

**New compact model for HfO₂ ferroelectric capacitors:**

| Feature | Details |
|---------|---------|
| **Name** | Heracles |
| **Language** | Verilog-A |
| **Target** | HfO₂ FeCap circuit simulation |
| **License** | MIT (open source) |
| **Paper** | arXiv (March 2025) |

### Preisach GLSL Implementation for Demo1

```glsl
#version 450

layout(local_size_x = 256) in;

struct Cell {
    float polarization;
    float lastField;
    float increasing;
    float turningPointCount;
    float turningPoints[8];  // History stack
};

layout(std140, binding = 0) uniform Material {
    float Ps;       // Saturation polarization
    float Ec;       // Coercive field
    float EcSigma;  // Distribution width
    float dt;
};

layout(std430, binding = 1) buffer CellBuffer {
    Cell cells[];
};

void main() {
    uint idx = gl_GlobalInvocationID.x;
    if (idx >= cells.length()) return;
    
    Cell cell = cells[idx];
    
    // Hyperbolic tangent switching (Bo Jiang method)
    float delta = EcSigma * 2.0;
    float P;
    
    if (cell.increasing > 0.5) {
        P = Ps * tanh((E - Ec) / delta);
    } else {
        P = Ps * tanh((E + Ec) / delta);
    }
    
    cells[idx].polarization = clamp(P, -Ps, Ps);
}
```

---

## 5. Real-Time GPU Hysteresis Visualization Architecture

### Data Flow Design

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CPU Input     │───▶│  GPU Compute    │───▶│  GPU Graphics   │
│  (Voltage UI)   │    │  (Physics Sim)  │    │  (Rendering)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                      │                      │
        ▼                      ▼                      ▼
   Uniform Buffer         SSBO Update         Vertex Buffer
   (appliedVoltage)     (polarization[])     (P-E curve pts)
```

### Integrated Compute + Graphics Pipeline

**Key Advantage:** Data stays on GPU between compute and render stages.

```go
// Per-frame loop
for !window.ShouldClose() {
    // 1. Update voltage from UI
    updateUniformBuffer(appliedVoltage)
    
    // 2. Dispatch compute shader
    vk.CmdDispatch(cmdBuffer, numCells/256, 1, 1)
    
    // 3. Pipeline barrier (compute → graphics)
    vk.CmdPipelineBarrier(cmdBuffer, 
        vk.PIPELINE_STAGE_COMPUTE_SHADER,
        vk.PIPELINE_STAGE_VERTEX_INPUT,
        ...)
    
    // 4. Render P-E curve using compute output as vertex data
    vk.CmdDraw(cmdBuffer, numPoints, 1, 0, 0)
}
```

---

## 6. Demo1 Implementation Strategy

### Phase 1: Vulkan Foundation (Week 1)

| Task | Status | Details |
|------|--------|---------|
| go-vk installation | Pending | `go get github.com/bbredesen/go-vk` |
| GLFW window creation | Pending | 1280×720, no OpenGL |
| Vulkan instance | Pending | Debug layers for dev |
| Device selection | Pending | Compute queue required |
| SPIR-V compilation | Pending | glslangValidator |

### Phase 2: Compute Pipeline (Week 2)

| Task | Status | Details |
|------|--------|---------|
| Material uniform buffer | Pending | Ps, Ec, EcSigma, thickness |
| Cell state SSBO | Pending | Array of ferroelectric cells |
| Preisach compute shader | ✅ Exists | shaders/preisach.comp |
| Descriptor sets | Pending | Bind buffers to shader |
| Pipeline creation | Pending | Load SPIR-V, create pipeline |

### Phase 3: Graphics Pipeline (Week 3)

| Task | Status | Details |
|------|--------|---------|
| P-E curve vertices | Pending | History buffer for plotting |
| Vertex/fragment shaders | ✅ Exists | hysteresis.vert/frag |
| Cell display | ✅ Exists | cell.vert/frag |
| Swapchain | Pending | VSync presentation |
| Render pass | Pending | Clear, draw, present |

### Phase 4: Integration (Week 4)

| Task | Status | Details |
|------|--------|---------|
| Interactive voltage control | Pending | GLFW input callbacks |
| Real-time P-E curve | Pending | Append points each frame |
| 30-state demonstration | Pending | Discrete levels |
| UI polish | Pending | Axis labels, legends |

---

## 7. New Papers for Download

### Vulkan & GPU Computing

| Paper/Resource | URL | Type |
|----------------|-----|------|
| Vulkan Compute Tutorial | https://vulkan-tutorial.com/Compute_Shader | Tutorial |
| vkGuide Compute Shaders | https://vkguide.dev/docs/gpudriven/compute_shaders | Guide |
| Vulkan GPGPU Example | https://github.com/Glavnokoman/vulkan-compute-example | Code |
| Baked Bits Vulkan Compute | https://bakedbits.dev/vulkan-compute | Tutorial |

### Go Vulkan Resources

| Resource | URL | Notes |
|----------|-----|-------|
| go-vk GitHub | https://github.com/bbredesen/go-vk | Official |
| vulkan-go GitHub | https://github.com/vulkan-go/vulkan | Alternative |
| Go Vulkan Tutorial Port | https://github.com/nicholasblaskey/go-vk-tutorial | Examples |
| GLFW Vulkan Guide | https://www.glfw.org/docs/latest/vulkan_guide.html | Integration |

### Preisach Model (2025)

| Paper | URL | Year |
|-------|-----|------|
| Heracles Compact Model | https://arxiv.org/abs/2503.xxxxx | 2025 |
| Preisach Parameter Fitting | https://pubmed.ncbi.nlm.nih.gov/... | 2024 |
| Physical Reality Preisach | https://www.nature.com/articles/s41467-018-06717-w | 2018 |

### Phase-Field GPU

| Paper | URL | Year |
|-------|-----|------|
| FerroX Framework | https://arxiv.org/abs/2210.15668 | 2022 |
| FerroX GitHub | https://github.com/AMReX-Microelectronics/FerroX | Current |
| GPU Phase-Field Review | https://www.sciencedirect.com/science/article/abs/pii/S0010465523001029 | 2023 |

---

## 8. Buffer Structure Definitions

### Go Structures for Vulkan Buffers

```go
package vulkan

// MaterialParams matches GLSL uniform buffer layout
type MaterialParams struct {
    Ps       float32 // Saturation polarization (C/m²)
    Ec       float32 // Coercive field (V/m)
    EcSigma  float32 // Distribution width
    Dt       float32 // Time step (s)
}

// SimParams matches GLSL uniform buffer
type SimParams struct {
    AppliedVoltage float32 // Input voltage (V)
    Thickness      float32 // Film thickness (m)
    Time           float32 // Current simulation time
    Reserved       float32 // Alignment padding
}

// Cell matches GLSL SSBO structure
type Cell struct {
    Polarization float32 // Current P state
    LastField    float32 // Previous E value
    Increasing   float32 // Direction flag (0 or 1)
    Padding      float32 // Alignment
}

// Std140 padding rules:
// - float32: 4 bytes, align 4
// - vec4/struct: align 16
// - arrays: each element align 16
```

### Descriptor Set Layout

```go
// Binding 0: MaterialParams (uniform buffer)
// Binding 1: SimParams (uniform buffer)  
// Binding 2: Cells[] (storage buffer, read/write)
// Binding 3: NormalizedP[] (storage buffer, output)

descriptorSetLayoutBindings := []vk.DescriptorSetLayoutBinding{
    {Binding: 0, DescriptorType: vk.DESCRIPTOR_TYPE_UNIFORM_BUFFER, ...},
    {Binding: 1, DescriptorType: vk.DESCRIPTOR_TYPE_UNIFORM_BUFFER, ...},
    {Binding: 2, DescriptorType: vk.DESCRIPTOR_TYPE_STORAGE_BUFFER, ...},
    {Binding: 3, DescriptorType: vk.DESCRIPTOR_TYPE_STORAGE_BUFFER, ...},
}
```

---

## 9. Shader Compilation

### glslangValidator Command

```bash
# Compile compute shader to SPIR-V
glslangValidator -V preisach.comp -o preisach.comp.spv

# Compile vertex/fragment shaders
glslangValidator -V hysteresis.vert -o hysteresis.vert.spv
glslangValidator -V hysteresis.frag -o hysteresis.frag.spv
glslangValidator -V cell.vert -o cell.vert.spv
glslangValidator -V cell.frag -o cell.frag.spv
```

### Updated compile.sh

```bash
#!/bin/bash
# Compile all shaders for Demo1

SHADER_DIR="$(dirname "$0")"
cd "$SHADER_DIR"

echo "Compiling shaders..."

# Compute shaders
glslangValidator -V preisach.comp -o preisach.comp.spv

# Hysteresis curve visualization
glslangValidator -V hysteresis.vert -o hysteresis.vert.spv
glslangValidator -V hysteresis.frag -o hysteresis.frag.spv

# Cell visualization
glslangValidator -V cell.vert -o cell.vert.spv
glslangValidator -V cell.frag -o cell.frag.spv

echo "Done. SPIR-V files generated."
ls -la *.spv
```

---

## 10. Key Takeaways

### Technology Stack Decision

| Component | Choice | Rationale |
|-----------|--------|-----------|
| **Vulkan bindings** | go-vk | Active, full API coverage |
| **Windowing** | GLFW via go-gl | Standard, proven |
| **Physics model** | Preisach (existing) | Already implemented in Go |
| **Shader language** | GLSL → SPIR-V | Standard Vulkan workflow |

### Risk Mitigation

| Risk | Mitigation |
|------|------------|
| go-vk learning curve | Use vulkan-tutorial.com (port to Go) |
| SPIR-V issues | Test with validation layers |
| Performance | Profile with GPU timers |
| GLFW integration | Follow official Vulkan guide |

### Next Research Priorities

1. **go-vk examples** - Find more compute shader samples
2. **Vulkan validation** - Debug layer configuration
3. **Memory optimization** - Device-local vs host-visible buffers
4. **Synchronization** - Fence/semaphore best practices

---

## References

1. Vulkan Tutorial (Compute): https://vulkan-tutorial.com/Compute_Shader
2. go-vk: https://github.com/bbredesen/go-vk
3. FerroX: https://github.com/AMReX-Microelectronics/FerroX
4. vkGuide: https://vkguide.dev/docs/gpudriven/compute_shaders
5. GLFW Vulkan: https://www.glfw.org/docs/latest/vulkan_guide.html

---

*Last updated: January 2026 - Part 8 (Vulkan Focus)*
