# Demo1 Vulkan Implementation Guide

Comprehensive guide for implementing GPU-accelerated ferroelectric hysteresis visualization using Vulkan.

---

## Overview

Demo1 (Hysteresis Visualizer) simulates and visualizes P-E hysteresis curves for HfO₂-ZrO₂ superlattice ferroelectric materials. This guide details the Vulkan implementation architecture.

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Demo1 Application                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  UI Layer   │───▶│  Physics    │───▶│    Visualization    │  │
│  │  (GLFW)     │    │  Engine     │    │    (Vulkan)         │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│         │                  │                      │             │
│         ▼                  ▼                      ▼             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │   Input     │    │  Compute    │    │   Graphics          │  │
│  │  Handling   │    │  Pipeline   │    │   Pipeline          │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 1. Prerequisites

### System Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| GPU | Vulkan 1.1+ | Vulkan 1.3 |
| Go | 1.21+ | 1.22+ |
| Vulkan SDK | 1.3.x | Latest |
| glslangValidator | Included in SDK | Latest |

### Go Dependencies

```bash
# Vulkan bindings
go get github.com/bbredesen/go-vk

# GLFW for windowing
go get github.com/go-gl/glfw/v3.3/glfw

# Math library
go get gonum.org/v1/gonum
```

### Vulkan SDK Installation (Linux)

```bash
# Ubuntu/Debian
wget -qO- https://packages.lunarg.com/lunarg-signing-key-pub.asc | sudo tee /etc/apt/trusted.gpg.d/lunarg.asc
sudo wget -qO /etc/apt/sources.list.d/lunarg-vulkan-jammy.list https://packages.lunarg.com/vulkan/lunarg-vulkan-jammy.list
sudo apt update
sudo apt install vulkan-sdk

# Verify installation
vulkaninfo --summary
```

---

## 2. Project Structure

```
demo1-hysteresis/
├── cmd/
│   └── hysteresis/
│       └── main.go                 # Entry point
├── pkg/
│   ├── ferroelectric/
│   │   ├── material.go             # ✅ HZO parameters
│   │   └── preisach.go             # ✅ CPU Preisach model
│   ├── render/
│   │   ├── render.go               # ✅ Renderer interface
│   │   └── plot.go                 # ✅ Plot helpers
│   ├── simulation/
│   │   └── engine.go               # ✅ Simulation coordinator
│   └── vulkan/                     # 🆕 NEW PACKAGE
│       ├── instance.go             # Vulkan initialization
│       ├── device.go               # Device & queue selection
│       ├── buffer.go               # Buffer management
│       ├── pipeline_compute.go     # Compute pipeline
│       ├── pipeline_graphics.go    # Graphics pipeline
│       ├── command.go              # Command buffer recording
│       ├── swapchain.go            # Presentation
│       ├── sync.go                 # Fences & semaphores
│       └── window.go               # GLFW integration
└── shaders/
    ├── preisach.comp               # ✅ Compute shader
    ├── hysteresis.vert             # ✅ P-E curve vertex
    ├── hysteresis.frag             # ✅ P-E curve fragment
    ├── cell.vert                   # ✅ Cell vertex
    ├── cell.frag                   # ✅ Cell fragment
    └── compile.sh                  # ✅ SPIR-V compilation
```

---

## 3. Vulkan Initialization

### pkg/vulkan/instance.go

```go
package vulkan

import (
    vk "github.com/bbredesen/go-vk"
    "github.com/go-gl/glfw/v3.3/glfw"
)

type Instance struct {
    handle           vk.Instance
    debugMessenger   vk.DebugUtilsMessengerEXT
    validationLayers []string
}

func NewInstance(appName string, enableValidation bool) (*Instance, error) {
    // Check validation layer support
    layers := []string{}
    if enableValidation {
        layers = append(layers, "VK_LAYER_KHRONOS_validation")
    }
    
    // Get GLFW required extensions
    glfwExts := glfw.GetRequiredInstanceExtensions()
    extensions := append(glfwExts, vk.KHR_SURFACE_EXTENSION_NAME)
    
    if enableValidation {
        extensions = append(extensions, vk.EXT_DEBUG_UTILS_EXTENSION_NAME)
    }
    
    // Create instance
    appInfo := vk.ApplicationInfo{
        SType:              vk.STRUCTURE_TYPE_APPLICATION_INFO,
        PApplicationName:   appName,
        ApplicationVersion: vk.MAKE_VERSION(1, 0, 0),
        PEngineName:        "Ferroelectric CIM",
        EngineVersion:      vk.MAKE_VERSION(1, 0, 0),
        ApiVersion:         vk.API_VERSION_1_3,
    }
    
    createInfo := vk.InstanceCreateInfo{
        SType:                   vk.STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
        PApplicationInfo:        &appInfo,
        EnabledExtensionCount:   uint32(len(extensions)),
        PpEnabledExtensionNames: extensions,
        EnabledLayerCount:       uint32(len(layers)),
        PpEnabledLayerNames:     layers,
    }
    
    instance, err := vk.CreateInstance(&createInfo, nil)
    if err != nil {
        return nil, err
    }
    
    return &Instance{
        handle:           instance,
        validationLayers: layers,
    }, nil
}

func (i *Instance) Destroy() {
    if i.debugMessenger != vk.DebugUtilsMessengerEXT(vk.NULL_HANDLE) {
        vk.DestroyDebugUtilsMessengerEXT(i.handle, i.debugMessenger, nil)
    }
    vk.DestroyInstance(i.handle, nil)
}
```

---

## 4. Device Selection

### pkg/vulkan/device.go

```go
package vulkan

import vk "github.com/bbredesen/go-vk"

type QueueFamilyIndices struct {
    Graphics uint32
    Compute  uint32
    Present  uint32
}

type Device struct {
    Physical       vk.PhysicalDevice
    Logical        vk.Device
    QueueFamilies  QueueFamilyIndices
    GraphicsQueue  vk.Queue
    ComputeQueue   vk.Queue
    PresentQueue   vk.Queue
}

func SelectPhysicalDevice(instance vk.Instance, surface vk.SurfaceKHR) (*Device, error) {
    devices, _ := vk.EnumeratePhysicalDevices(instance)
    
    for _, device := range devices {
        if isDeviceSuitable(device, surface) {
            return createLogicalDevice(device, surface)
        }
    }
    
    return nil, fmt.Errorf("no suitable GPU found")
}

func isDeviceSuitable(device vk.PhysicalDevice, surface vk.SurfaceKHR) bool {
    // Check for compute queue support
    queueFamilies, _ := vk.GetPhysicalDeviceQueueFamilyProperties(device)
    
    hasCompute := false
    hasGraphics := false
    hasPresent := false
    
    for i, family := range queueFamilies {
        if family.QueueFlags&vk.QUEUE_COMPUTE_BIT != 0 {
            hasCompute = true
        }
        if family.QueueFlags&vk.QUEUE_GRAPHICS_BIT != 0 {
            hasGraphics = true
        }
        
        supported, _ := vk.GetPhysicalDeviceSurfaceSupportKHR(device, uint32(i), surface)
        if supported {
            hasPresent = true
        }
    }
    
    return hasCompute && hasGraphics && hasPresent
}

func createLogicalDevice(physical vk.PhysicalDevice, surface vk.SurfaceKHR) (*Device, error) {
    indices := findQueueFamilies(physical, surface)
    
    // Create queues
    uniqueFamilies := map[uint32]struct{}{
        indices.Graphics: {},
        indices.Compute:  {},
        indices.Present:  {},
    }
    
    var queueCreateInfos []vk.DeviceQueueCreateInfo
    queuePriority := float32(1.0)
    
    for family := range uniqueFamilies {
        queueCreateInfos = append(queueCreateInfos, vk.DeviceQueueCreateInfo{
            SType:            vk.STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
            QueueFamilyIndex: family,
            QueueCount:       1,
            PQueuePriorities: &queuePriority,
        })
    }
    
    // Device features (enable what we need)
    features := vk.PhysicalDeviceFeatures{}
    
    // Required extensions
    extensions := []string{vk.KHR_SWAPCHAIN_EXTENSION_NAME}
    
    createInfo := vk.DeviceCreateInfo{
        SType:                   vk.STRUCTURE_TYPE_DEVICE_CREATE_INFO,
        QueueCreateInfoCount:    uint32(len(queueCreateInfos)),
        PQueueCreateInfos:       queueCreateInfos,
        PEnabledFeatures:        &features,
        EnabledExtensionCount:   uint32(len(extensions)),
        PpEnabledExtensionNames: extensions,
    }
    
    logical, err := vk.CreateDevice(physical, &createInfo, nil)
    if err != nil {
        return nil, err
    }
    
    device := &Device{
        Physical:      physical,
        Logical:       logical,
        QueueFamilies: indices,
    }
    
    // Get queue handles
    device.GraphicsQueue = vk.GetDeviceQueue(logical, indices.Graphics, 0)
    device.ComputeQueue = vk.GetDeviceQueue(logical, indices.Compute, 0)
    device.PresentQueue = vk.GetDeviceQueue(logical, indices.Present, 0)
    
    return device, nil
}
```

---

## 5. Buffer Management

### pkg/vulkan/buffer.go

```go
package vulkan

import vk "github.com/bbredesen/go-vk"

type Buffer struct {
    Handle vk.Buffer
    Memory vk.DeviceMemory
    Size   vk.DeviceSize
}

// MaterialParams - uniform buffer for ferroelectric parameters
type MaterialParams struct {
    Ps      float32 // Saturation polarization (C/m²)
    Ec      float32 // Coercive field (V/m)
    EcSigma float32 // Distribution width
    Dt      float32 // Time step
}

// SimParams - uniform buffer for simulation state
type SimParams struct {
    AppliedVoltage float32
    Thickness      float32
    Time           float32
    Reserved       float32
}

// Cell - storage buffer element
type Cell struct {
    Polarization float32
    LastField    float32
    Increasing   float32
    Padding      float32
}

func CreateBuffer(
    device *Device,
    size vk.DeviceSize,
    usage vk.BufferUsageFlags,
    properties vk.MemoryPropertyFlags,
) (*Buffer, error) {
    
    bufferInfo := vk.BufferCreateInfo{
        SType:       vk.STRUCTURE_TYPE_BUFFER_CREATE_INFO,
        Size:        size,
        Usage:       usage,
        SharingMode: vk.SHARING_MODE_EXCLUSIVE,
    }
    
    buffer, err := vk.CreateBuffer(device.Logical, &bufferInfo, nil)
    if err != nil {
        return nil, err
    }
    
    // Get memory requirements
    memRequirements := vk.GetBufferMemoryRequirements(device.Logical, buffer)
    
    // Find suitable memory type
    memTypeIndex := findMemoryType(device.Physical, memRequirements.MemoryTypeBits, properties)
    
    allocInfo := vk.MemoryAllocateInfo{
        SType:           vk.STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
        AllocationSize:  memRequirements.Size,
        MemoryTypeIndex: memTypeIndex,
    }
    
    memory, err := vk.AllocateMemory(device.Logical, &allocInfo, nil)
    if err != nil {
        vk.DestroyBuffer(device.Logical, buffer, nil)
        return nil, err
    }
    
    vk.BindBufferMemory(device.Logical, buffer, memory, 0)
    
    return &Buffer{
        Handle: buffer,
        Memory: memory,
        Size:   size,
    }, nil
}

// CreateUniformBuffer creates a host-visible uniform buffer
func CreateUniformBuffer(device *Device, size int) (*Buffer, error) {
    return CreateBuffer(
        device,
        vk.DeviceSize(size),
        vk.BUFFER_USAGE_UNIFORM_BUFFER_BIT,
        vk.MEMORY_PROPERTY_HOST_VISIBLE_BIT|vk.MEMORY_PROPERTY_HOST_COHERENT_BIT,
    )
}

// CreateStorageBuffer creates a device-local storage buffer
func CreateStorageBuffer(device *Device, size int) (*Buffer, error) {
    return CreateBuffer(
        device,
        vk.DeviceSize(size),
        vk.BUFFER_USAGE_STORAGE_BUFFER_BIT|vk.BUFFER_USAGE_TRANSFER_DST_BIT,
        vk.MEMORY_PROPERTY_DEVICE_LOCAL_BIT,
    )
}

// Upload data to a host-visible buffer
func (b *Buffer) Upload(device vk.Device, data []byte) error {
    ptr, err := vk.MapMemory(device, b.Memory, 0, b.Size, 0)
    if err != nil {
        return err
    }
    
    copy(ptr, data)
    vk.UnmapMemory(device, b.Memory)
    
    return nil
}
```

---

## 6. Compute Pipeline

### pkg/vulkan/pipeline_compute.go

```go
package vulkan

import (
    "io/ioutil"
    vk "github.com/bbredesen/go-vk"
)

type ComputePipeline struct {
    Pipeline          vk.Pipeline
    Layout            vk.PipelineLayout
    DescriptorSetLayout vk.DescriptorSetLayout
    DescriptorPool    vk.DescriptorPool
    DescriptorSet     vk.DescriptorSet
}

func CreateComputePipeline(
    device *Device,
    shaderPath string,
    bindings []vk.DescriptorSetLayoutBinding,
) (*ComputePipeline, error) {
    
    // Create descriptor set layout
    layoutInfo := vk.DescriptorSetLayoutCreateInfo{
        SType:        vk.STRUCTURE_TYPE_DESCRIPTOR_SET_LAYOUT_CREATE_INFO,
        BindingCount: uint32(len(bindings)),
        PBindings:    bindings,
    }
    
    descriptorSetLayout, err := vk.CreateDescriptorSetLayout(device.Logical, &layoutInfo, nil)
    if err != nil {
        return nil, err
    }
    
    // Create pipeline layout
    pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
        SType:          vk.STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO,
        SetLayoutCount: 1,
        PSetLayouts:    []vk.DescriptorSetLayout{descriptorSetLayout},
    }
    
    pipelineLayout, err := vk.CreatePipelineLayout(device.Logical, &pipelineLayoutInfo, nil)
    if err != nil {
        return nil, err
    }
    
    // Load SPIR-V shader
    shaderCode, err := ioutil.ReadFile(shaderPath)
    if err != nil {
        return nil, err
    }
    
    shaderModuleInfo := vk.ShaderModuleCreateInfo{
        SType:    vk.STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO,
        CodeSize: uint64(len(shaderCode)),
        PCode:    shaderCode,
    }
    
    shaderModule, err := vk.CreateShaderModule(device.Logical, &shaderModuleInfo, nil)
    if err != nil {
        return nil, err
    }
    defer vk.DestroyShaderModule(device.Logical, shaderModule, nil)
    
    // Create compute pipeline
    shaderStageInfo := vk.PipelineShaderStageCreateInfo{
        SType:  vk.STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
        Stage:  vk.SHADER_STAGE_COMPUTE_BIT,
        Module: shaderModule,
        PName:  "main",
    }
    
    pipelineInfo := vk.ComputePipelineCreateInfo{
        SType:  vk.STRUCTURE_TYPE_COMPUTE_PIPELINE_CREATE_INFO,
        Stage:  shaderStageInfo,
        Layout: pipelineLayout,
    }
    
    pipelines, err := vk.CreateComputePipelines(
        device.Logical,
        vk.PipelineCache(vk.NULL_HANDLE),
        []vk.ComputePipelineCreateInfo{pipelineInfo},
        nil,
    )
    if err != nil {
        return nil, err
    }
    
    return &ComputePipeline{
        Pipeline:          pipelines[0],
        Layout:            pipelineLayout,
        DescriptorSetLayout: descriptorSetLayout,
    }, nil
}

// Preisach-specific pipeline setup
func CreatePreisachPipeline(device *Device, shaderDir string) (*ComputePipeline, error) {
    bindings := []vk.DescriptorSetLayoutBinding{
        // Binding 0: MaterialParams (uniform)
        {
            Binding:         0,
            DescriptorType:  vk.DESCRIPTOR_TYPE_UNIFORM_BUFFER,
            DescriptorCount: 1,
            StageFlags:      vk.SHADER_STAGE_COMPUTE_BIT,
        },
        // Binding 1: SimParams (uniform)
        {
            Binding:         1,
            DescriptorType:  vk.DESCRIPTOR_TYPE_UNIFORM_BUFFER,
            DescriptorCount: 1,
            StageFlags:      vk.SHADER_STAGE_COMPUTE_BIT,
        },
        // Binding 2: CellBuffer (storage, read/write)
        {
            Binding:         2,
            DescriptorType:  vk.DESCRIPTOR_TYPE_STORAGE_BUFFER,
            DescriptorCount: 1,
            StageFlags:      vk.SHADER_STAGE_COMPUTE_BIT,
        },
        // Binding 3: NormalizedP output (storage)
        {
            Binding:         3,
            DescriptorType:  vk.DESCRIPTOR_TYPE_STORAGE_BUFFER,
            DescriptorCount: 1,
            StageFlags:      vk.SHADER_STAGE_COMPUTE_BIT,
        },
    }
    
    return CreateComputePipeline(device, shaderDir+"/preisach.comp.spv", bindings)
}
```

---

## 7. Command Recording

### pkg/vulkan/command.go

```go
package vulkan

import vk "github.com/bbredesen/go-vk"

type CommandPool struct {
    Handle vk.CommandPool
}

func CreateCommandPool(device *Device, queueFamily uint32) (*CommandPool, error) {
    poolInfo := vk.CommandPoolCreateInfo{
        SType:            vk.STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO,
        QueueFamilyIndex: queueFamily,
        Flags:            vk.COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT,
    }
    
    pool, err := vk.CreateCommandPool(device.Logical, &poolInfo, nil)
    if err != nil {
        return nil, err
    }
    
    return &CommandPool{Handle: pool}, nil
}

func AllocateCommandBuffer(device *Device, pool *CommandPool) (vk.CommandBuffer, error) {
    allocInfo := vk.CommandBufferAllocateInfo{
        SType:              vk.STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
        CommandPool:        pool.Handle,
        Level:              vk.COMMAND_BUFFER_LEVEL_PRIMARY,
        CommandBufferCount: 1,
    }
    
    buffers, err := vk.AllocateCommandBuffers(device.Logical, &allocInfo)
    if err != nil {
        return vk.CommandBuffer(vk.NULL_HANDLE), err
    }
    
    return buffers[0], nil
}

// RecordComputeDispatch records commands for physics simulation
func RecordComputeDispatch(
    cmdBuffer vk.CommandBuffer,
    pipeline *ComputePipeline,
    numCells uint32,
) error {
    
    beginInfo := vk.CommandBufferBeginInfo{
        SType: vk.STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO,
    }
    
    if err := vk.BeginCommandBuffer(cmdBuffer, &beginInfo); err != nil {
        return err
    }
    
    // Bind compute pipeline
    vk.CmdBindPipeline(cmdBuffer, vk.PIPELINE_BIND_POINT_COMPUTE, pipeline.Pipeline)
    
    // Bind descriptor set
    vk.CmdBindDescriptorSets(
        cmdBuffer,
        vk.PIPELINE_BIND_POINT_COMPUTE,
        pipeline.Layout,
        0,
        []vk.DescriptorSet{pipeline.DescriptorSet},
        nil,
    )
    
    // Dispatch compute work
    // Each workgroup has 256 invocations (defined in shader)
    workgroupCount := (numCells + 255) / 256
    vk.CmdDispatch(cmdBuffer, workgroupCount, 1, 1)
    
    return vk.EndCommandBuffer(cmdBuffer)
}

// RecordComputeToGraphicsBarrier ensures compute writes complete before graphics reads
func RecordComputeToGraphicsBarrier(cmdBuffer vk.CommandBuffer, buffer *Buffer) {
    barrier := vk.BufferMemoryBarrier{
        SType:               vk.STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
        SrcAccessMask:       vk.ACCESS_SHADER_WRITE_BIT,
        DstAccessMask:       vk.ACCESS_VERTEX_ATTRIBUTE_READ_BIT,
        SrcQueueFamilyIndex: vk.QUEUE_FAMILY_IGNORED,
        DstQueueFamilyIndex: vk.QUEUE_FAMILY_IGNORED,
        Buffer:              buffer.Handle,
        Offset:              0,
        Size:                vk.WHOLE_SIZE,
    }
    
    vk.CmdPipelineBarrier(
        cmdBuffer,
        vk.PIPELINE_STAGE_COMPUTE_SHADER_BIT,
        vk.PIPELINE_STAGE_VERTEX_INPUT_BIT,
        0,
        nil,
        []vk.BufferMemoryBarrier{barrier},
        nil,
    )
}
```

---

## 8. Main Loop Integration

### Updated cmd/hysteresis/main.go (Vulkan mode)

```go
package main

import (
    "log"
    
    "multilayer-ferroelectric-cim-visualizer/demo1-hysteresis/pkg/ferroelectric"
    "multilayer-ferroelectric-cim-visualizer/demo1-hysteresis/pkg/vulkan"
    
    "github.com/go-gl/glfw/v3.3/glfw"
)

func runVulkanMode(material *ferroelectric.HZOMaterial) error {
    // Initialize GLFW
    if err := glfw.Init(); err != nil {
        return err
    }
    defer glfw.Terminate()
    
    // Create window (no OpenGL context)
    glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
    glfw.WindowHint(glfw.Resizable, glfw.False)
    
    window, err := glfw.CreateWindow(1280, 720, "Ferroelectric CIM Hysteresis Demo", nil, nil)
    if err != nil {
        return err
    }
    defer window.Destroy()
    
    // Initialize Vulkan
    instance, err := vulkan.NewInstance("Ferroelectric CIM Demo1", true)
    if err != nil {
        return err
    }
    defer instance.Destroy()
    
    // Create surface
    surface, err := window.CreateWindowSurface(instance.Handle(), nil)
    if err != nil {
        return err
    }
    
    // Select device
    device, err := vulkan.SelectPhysicalDevice(instance.Handle(), surface)
    if err != nil {
        return err
    }
    defer device.Destroy()
    
    // Create compute pipeline
    computePipeline, err := vulkan.CreatePreisachPipeline(device, "shaders")
    if err != nil {
        return err
    }
    defer computePipeline.Destroy()
    
    // Create buffers
    numCells := uint32(1024)
    
    materialBuffer, _ := vulkan.CreateUniformBuffer(device, 16) // MaterialParams
    simBuffer, _ := vulkan.CreateUniformBuffer(device, 16)      // SimParams
    cellBuffer, _ := vulkan.CreateStorageBuffer(device, int(numCells)*16)
    outputBuffer, _ := vulkan.CreateStorageBuffer(device, int(numCells)*4)
    
    // Upload material parameters
    materialParams := vulkan.MaterialParams{
        Ps:      float32(material.Ps),
        Ec:      float32(material.Ec),
        EcSigma: float32(material.Ec * 0.25),
        Dt:      1e-9,
    }
    materialBuffer.Upload(device.Logical, structToBytes(materialParams))
    
    // Create command pool and buffer
    cmdPool, _ := vulkan.CreateCommandPool(device, device.QueueFamilies.Compute)
    cmdBuffer, _ := vulkan.AllocateCommandBuffer(device, cmdPool)
    
    // Main loop
    appliedVoltage := float32(0)
    for !window.ShouldClose() {
        glfw.PollEvents()
        
        // Handle input (voltage control)
        if window.GetKey(glfw.KeyUp) == glfw.Press {
            appliedVoltage += 0.1
        }
        if window.GetKey(glfw.KeyDown) == glfw.Press {
            appliedVoltage -= 0.1
        }
        
        // Update simulation parameters
        simParams := vulkan.SimParams{
            AppliedVoltage: appliedVoltage,
            Thickness:      float32(material.Thickness),
            Time:           0,
        }
        simBuffer.Upload(device.Logical, structToBytes(simParams))
        
        // Record and submit compute commands
        vulkan.RecordComputeDispatch(cmdBuffer, computePipeline, numCells)
        submitCompute(device, cmdBuffer)
        
        // Render frame
        // ... graphics pipeline rendering ...
    }
    
    return nil
}
```

---

## 9. Implementation Checklist

### Core Vulkan Infrastructure

- [ ] `pkg/vulkan/instance.go` - Vulkan instance with validation
- [ ] `pkg/vulkan/device.go` - Device selection with compute support
- [ ] `pkg/vulkan/buffer.go` - Buffer creation and memory management
- [ ] `pkg/vulkan/sync.go` - Fences and semaphores
- [ ] `pkg/vulkan/command.go` - Command pool and buffer utilities

### Compute Pipeline

- [ ] `pkg/vulkan/pipeline_compute.go` - Preisach compute pipeline
- [ ] `pkg/vulkan/descriptor.go` - Descriptor set management
- [ ] Update `shaders/preisach.comp` - Add SPIR-V compatibility
- [ ] `shaders/compile.sh` - Generate .spv files

### Graphics Pipeline

- [ ] `pkg/vulkan/swapchain.go` - Swapchain management
- [ ] `pkg/vulkan/pipeline_graphics.go` - P-E curve rendering
- [ ] `pkg/vulkan/renderpass.go` - Render pass setup
- [ ] Framebuffers for swapchain images

### Integration

- [ ] `pkg/vulkan/window.go` - GLFW window wrapper
- [ ] Update `cmd/hysteresis/main.go` - Vulkan mode
- [ ] Input handling for voltage control
- [ ] Real-time P-E curve plotting

---

## 10. Performance Optimization

### Memory Strategy

| Buffer Type | Memory Location | Access Pattern |
|-------------|-----------------|----------------|
| MaterialParams | Host-visible | Update rarely |
| SimParams | Host-visible | Update every frame |
| CellBuffer | Device-local | GPU read/write |
| OutputBuffer | Device-local | GPU write, transfer for plotting |

### Workgroup Sizing

```glsl
// In preisach.comp
layout(local_size_x = 256, local_size_y = 1, local_size_z = 1) in;

// Optimal for most GPUs:
// - 256 invocations per workgroup
// - Multiple of warp/wavefront size (32/64)
// - Good occupancy for compute-bound shaders
```

### Synchronization Tips

1. Use `vkWaitForFences` instead of `vkDeviceWaitIdle`
2. Batch compute dispatches when possible
3. Use pipeline barriers for compute → graphics sync
4. Consider async compute on separate queue

---

## References

1. Vulkan Tutorial: https://vulkan-tutorial.com/Compute_Shader
2. go-vk Documentation: https://github.com/bbredesen/go-vk
3. GLFW Vulkan Guide: https://www.glfw.org/docs/latest/vulkan_guide.html
4. Vulkan Compute Best Practices: https://vkguide.dev/docs/gpudriven/compute_shaders

---

*Last updated: January 2026*
