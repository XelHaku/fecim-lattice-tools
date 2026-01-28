# Shared Vulkan Compute Infrastructure Plan

## Overview

Create a unified Vulkan compute infrastructure in `shared/compute/` that serves both module2-crossbar and module4-circuits, eliminating code duplication and enabling GPU acceleration across the FeCIM toolchain.

## Current State Analysis

### What Exists

| Module | Vulkan Runtime | Compute Shaders | Status |
|--------|----------------|-----------------|--------|
| module1-hysteresis | Full (1473 lines) | preisach.comp | Production - graphics + compute |
| module2-crossbar | **NONE** | mvm.comp, activation.comp | Shaders orphaned - CPU fallback used |
| module4-circuits | **NONE** | **NONE** | No GPU code |

### Key Finding: Orphaned Shaders
The `module2-crossbar/shaders/mvm.comp` (203 lines) implements full MVM physics with:
- Kirchhoff's Current Law: I_j = Sum_i(V_i * G_ij)
- DAC/ADC quantization
- IR drop modeling
- Conductance drift
- Device-to-device variation
- Read noise (thermal + shot)

**But there's no Go code to call it** - the `CPUReference` struct in `pkg/crossbar/reference.go` is used instead.

### Existing Pattern (module1-hysteresis/pkg/render/vulkan.go)

```go
type VulkanRenderer struct {
    instance       vk.Instance
    device         vk.Device
    commandPool    vk.CommandPool
    commandBuffers []vk.CommandBuffer
    // ... full graphics pipeline
}
```

Key methods:
- `createInstance()` - Vulkan instance with extensions
- `createLogicalDevice()` - Device with compute queue
- `createCommandPool()` - Command pool for queue family
- Shader loading via `os.ReadFile()` for SPIR-V

## Proposed Architecture

### Directory Structure

```
shared/
├── compute/
│   ├── context.go          # VulkanContext - shared instance/device management
│   ├── compute_pipeline.go # ComputePipeline - shader execution abstraction
│   ├── buffer.go           # GPU buffer management (staging, device-local)
│   ├── shader_loader.go    # SPIR-V loading utilities
│   ├── dispatcher.go       # Workgroup dispatch helpers
│   └── compute_test.go     # Integration tests
└── physics/
    └── (existing quantization.go, conductance.go)
```

### Core Interfaces

```go
// shared/compute/context.go

// VulkanContext manages shared Vulkan resources
type VulkanContext struct {
    instance       vk.Instance
    physicalDevice vk.PhysicalDevice
    device         vk.Device
    computeQueue   vk.Queue
    computeFamily  uint32
    memoryProps    vk.PhysicalDeviceMemoryProperties
    commandPool    vk.CommandPool
}

func NewVulkanContext() (*VulkanContext, error)
func (c *VulkanContext) CreateComputePipeline(shaderPath string) (*ComputePipeline, error)
func (c *VulkanContext) Destroy()
```

```go
// shared/compute/compute_pipeline.go

// ComputePipeline represents a compiled compute shader
type ComputePipeline struct {
    ctx            *VulkanContext
    pipeline       vk.Pipeline
    pipelineLayout vk.PipelineLayout
    descriptorPool vk.DescriptorPool
    descriptorSets []vk.DescriptorSet
    shaderModule   vk.ShaderModule
}

func (p *ComputePipeline) BindBuffer(binding uint32, buffer *GPUBuffer) error
func (p *ComputePipeline) SetUniform(binding uint32, data []byte) error
func (p *ComputePipeline) Dispatch(groupCountX, groupCountY, groupCountZ uint32) error
func (p *ComputePipeline) Destroy()
```

```go
// shared/compute/buffer.go

// GPUBuffer wraps Vulkan buffer with memory management
type GPUBuffer struct {
    buffer       vk.Buffer
    memory       vk.DeviceMemory
    size         vk.DeviceSize
    usage        vk.BufferUsageFlags
    deviceLocal  bool
}

func (c *VulkanContext) CreateBuffer(size uint64, usage BufferUsage) (*GPUBuffer, error)
func (b *GPUBuffer) Upload(data []byte) error
func (b *GPUBuffer) Download(data []byte) error
func (b *GPUBuffer) Destroy()
```

## Implementation Tasks

### Phase 1: Core Infrastructure (shared/compute/)

1. **context.go** - VulkanContext
   - Initialize Vulkan without GLFW (headless compute)
   - **Key:** Use `vk.SetDefaultGetInstanceProcAddr()` instead of GLFW-based init
   - Find compute-capable queue family (VK_QUEUE_COMPUTE_BIT)
   - Create logical device with compute queue
   - Command pool for compute operations
   - Memory property caching

2. **buffer.go** - GPUBuffer
   - Device-local buffers (fast GPU access)
   - Host-visible staging buffers
   - Upload/download with staging
   - Automatic memory type selection

3. **shader_loader.go** - SPIR-V Utilities
   - Load .comp.spv files
   - Create shader modules
   - Validate shader bindings

4. **compute_pipeline.go** - ComputePipeline
   - Descriptor set layout from shader reflection
   - Pipeline layout creation
   - Pipeline compilation
   - Descriptor pool and set allocation
   - Buffer binding
   - Command buffer recording and submission

5. **dispatcher.go** - Dispatch Helpers
   - Calculate optimal workgroup sizes
   - Handle array size edge cases
   - Synchronization primitives

### Phase 2: Module Integration

#### module2-crossbar Integration

1. **pkg/crossbar/gpu_mvm.go** (new)
   ```go
   type GPUAccelerator struct {
       ctx      *compute.VulkanContext
       mvmPipe  *compute.ComputePipeline
       actPipe  *compute.ComputePipeline
   }

   func NewGPUAccelerator() (*GPUAccelerator, error)
   func (g *GPUAccelerator) MVM(conductances []float32, inputs []float32) ([]float32, error)
   func (g *GPUAccelerator) Activation(data []float32, actType ActivationType) error
   ```

2. Update `pkg/crossbar/array.go`:
   - Add `UseGPU bool` to `ArrayConfig` struct
   - Lazy-initialize GPUAccelerator on first MVM call when enabled
   - Fallback to CPU when Vulkan unavailable or initialization fails
   - Batch operations for efficiency

3. Move shaders:
   - `module2-crossbar/shaders/mvm.comp` → `shared/compute/shaders/mvm.comp`
   - `module2-crossbar/shaders/activation.comp` → `shared/compute/shaders/activation.comp`

#### module4-circuits Integration

1. **shaders/dac.comp** (new) - DAC conversion with INL/DNL
   ```glsl
   // Parallel DAC conversion for batch processing
   layout(local_size_x = 256) in;
   // ... INL/DNL modeling
   ```

2. **shaders/adc.comp** (new) - ADC conversion
   ```glsl
   // Flash ADC parallel comparators
   // SAR ADC step simulation
   ```

3. **shaders/tia.comp** (new) - TIA signal processing
   ```glsl
   // Transimpedance amplifier model
   // Current-to-voltage conversion
   ```

4. **pkg/peripherals/gpu_peripherals.go** (new)
   - GPU-accelerated batch peripheral simulation

### Phase 3: Shader Consolidation

Create `shared/compute/shaders/`:
```
shared/compute/shaders/
├── compile.sh           # Universal compilation script
├── mvm.comp             # From module2-crossbar
├── activation.comp      # From module2-crossbar
├── dac.comp             # New for module4
├── adc.comp             # New for module4
├── tia.comp             # New for module4
└── quantize.comp        # 30-level quantization (shared)
```

## File Changes Summary

### New Files

| File | Lines (est) | Purpose |
|------|-------------|---------|
| shared/compute/context.go | 250 | Vulkan instance/device management |
| shared/compute/buffer.go | 200 | GPU buffer abstraction |
| shared/compute/shader_loader.go | 80 | SPIR-V loading |
| shared/compute/compute_pipeline.go | 350 | Pipeline management |
| shared/compute/dispatcher.go | 100 | Dispatch helpers |
| shared/compute/compute_test.go | 150 | Tests |
| shared/compute/shaders/compile.sh | 40 | Shader build script |
| shared/compute/shaders/quantize.comp | 50 | Standalone 30-level quantization for batch ops (mvm.comp has inline version) |
| module2-crossbar/pkg/crossbar/gpu_mvm.go | 200 | MVM GPU integration |
| module4-circuits/shaders/dac.comp | 80 | DAC shader |
| module4-circuits/shaders/adc.comp | 100 | ADC shader |
| module4-circuits/shaders/tia.comp | 60 | TIA shader |
| module4-circuits/pkg/peripherals/gpu_peripherals.go | 150 | Circuits GPU integration |

### Modified Files

| File | Changes |
|------|---------|
| module2-crossbar/pkg/crossbar/array.go | Add GPU accelerator option |
| module4-circuits/pkg/peripherals/analysis.go | Use GPU for batch operations |
| go.mod | Ensure vulkan-go dependency |

### Moved Files

| From | To |
|------|-----|
| module2-crossbar/shaders/mvm.comp | shared/compute/shaders/mvm.comp |
| module2-crossbar/shaders/activation.comp | shared/compute/shaders/activation.comp |
| module2-crossbar/shaders/*.spv | shared/compute/shaders/*.spv |

## Verification Steps

1. **Unit Tests**
   - `go test ./shared/compute/...` - Context creation, buffer ops
   - Mock Vulkan for CI (or skip GPU tests)

2. **Integration Tests**
   - MVM correctness: GPU vs CPU reference
   - Shader compilation: All .comp → .spv
   - Memory leak detection

3. **Performance Benchmarks**
   - `BenchmarkMVM_CPU` vs `BenchmarkMVM_GPU`
   - Target: 10-100x speedup for 256x256+ arrays

4. **Build Verification**
   - `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools`
   - Run without Vulkan (graceful fallback)

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| No GPU available | CPU fallback in all GPU code paths |
| Vulkan init fails | Graceful error handling, fallback mode |
| Shader compilation fails | Pre-compiled .spv files committed |
| Memory exhaustion | Buffer pooling, explicit limits |
| Cross-platform issues | Test on Linux (primary), document others |

## Dependencies

- `github.com/vulkan-go/vulkan` (already in go.mod)
- `glslc` from Vulkan SDK (for shader compilation)
- No new Go dependencies required

## Acceptance Criteria

1. [ ] `shared/compute/` package compiles and tests pass
2. [ ] VulkanContext initializes without GLFW (headless)
3. [ ] module2-crossbar MVM uses GPU when available
4. [ ] module4-circuits DAC/ADC shaders work
5. [ ] Existing tests still pass (`go test ./...`)
6. [ ] CPU fallback works when Vulkan unavailable
7. [ ] No orphaned shaders - all are integrated
8. [ ] Documentation updated (CLAUDE.md references)
