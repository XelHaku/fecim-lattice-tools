// Package compute provides headless Vulkan compute context for GPU-accelerated operations.
package compute

import (
	"fmt"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// BindingType describes what kind of data a binding holds.
type BindingType uint32

const (
	// BindingTypeUniform is for uniform buffers (std140 layout).
	BindingTypeUniform BindingType = iota
	// BindingTypeStorage is for storage buffers (std430 layout).
	BindingTypeStorage
)

// BindingInfo describes a single shader binding.
type BindingInfo struct {
	Binding uint32      // Binding index in shader
	Type    BindingType // Uniform or storage buffer
	Size    uint64      // For uniform buffers, size of the struct
}

// PipelineConfig holds configuration for creating a compute pipeline.
type PipelineConfig struct {
	ShaderPath string        // Path to .comp.spv file
	Bindings   []BindingInfo // Descriptor bindings (must match shader)
}

// ComputePipeline represents a compiled compute shader ready for execution.
type ComputePipeline struct {
	ctx              *VulkanContext
	shaderModule     vk.ShaderModule
	descriptorLayout vk.DescriptorSetLayout
	pipelineLayout   vk.PipelineLayout
	pipeline         vk.Pipeline
	descriptorPool   vk.DescriptorPool
	descriptorSet    vk.DescriptorSet
	commandBuffer    vk.CommandBuffer
	fence            vk.Fence
	bindings         []BindingInfo

	// Internal uniform buffers (created on-demand for SetUniformRaw)
	uniformBuffers map[uint32]*uniformBuffer
}

// uniformBuffer holds internal uniform buffer state.
type uniformBuffer struct {
	buffer vk.Buffer
	memory vk.DeviceMemory
	size   uint64
}

// NewComputePipeline creates a new compute pipeline from configuration.
// The shader is loaded from disk and compiled. Descriptor sets are allocated.
func NewComputePipeline(ctx *VulkanContext, config PipelineConfig) (*ComputePipeline, error) {
	if ctx == nil {
		return nil, fmt.Errorf("VulkanContext is nil")
	}
	if !ctx.IsAvailable() {
		return nil, fmt.Errorf("VulkanContext is not available")
	}
	if len(config.Bindings) == 0 {
		return nil, fmt.Errorf("at least one binding is required")
	}

	p := &ComputePipeline{
		ctx:            ctx,
		bindings:       config.Bindings,
		uniformBuffers: make(map[uint32]*uniformBuffer),
	}

	// Load and create shader module
	spirvCode, err := LoadSPIRV(config.ShaderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load shader: %w", err)
	}

	p.shaderModule, err = CreateShaderModule(ctx, spirvCode)
	if err != nil {
		return nil, fmt.Errorf("failed to create shader module: %w", err)
	}

	// Create descriptor set layout
	if err := p.createDescriptorSetLayout(); err != nil {
		p.Destroy()
		return nil, fmt.Errorf("failed to create descriptor set layout: %w", err)
	}

	// Create pipeline layout
	if err := p.createPipelineLayout(); err != nil {
		p.Destroy()
		return nil, fmt.Errorf("failed to create pipeline layout: %w", err)
	}

	// Create compute pipeline
	if err := p.createComputePipeline(); err != nil {
		p.Destroy()
		return nil, fmt.Errorf("failed to create compute pipeline: %w", err)
	}

	// Create descriptor pool and allocate set
	if err := p.createDescriptorPoolAndSet(); err != nil {
		p.Destroy()
		return nil, fmt.Errorf("failed to create descriptor pool/set: %w", err)
	}

	// Allocate command buffer
	if err := p.allocateCommandBuffer(); err != nil {
		p.Destroy()
		return nil, fmt.Errorf("failed to allocate command buffer: %w", err)
	}

	// Create fence for synchronization
	if err := p.createFence(); err != nil {
		p.Destroy()
		return nil, fmt.Errorf("failed to create fence: %w", err)
	}

	return p, nil
}

// createDescriptorSetLayout creates the descriptor set layout from binding info.
func (p *ComputePipeline) createDescriptorSetLayout() error {
	layoutBindings := make([]vk.DescriptorSetLayoutBinding, len(p.bindings))

	for i, binding := range p.bindings {
		var descriptorType vk.DescriptorType
		switch binding.Type {
		case BindingTypeUniform:
			descriptorType = vk.DescriptorTypeUniformBuffer
		case BindingTypeStorage:
			descriptorType = vk.DescriptorTypeStorageBuffer
		default:
			return fmt.Errorf("unknown binding type: %d", binding.Type)
		}

		layoutBindings[i] = vk.DescriptorSetLayoutBinding{
			Binding:         binding.Binding,
			DescriptorType:  descriptorType,
			DescriptorCount: 1,
			StageFlags:      vk.ShaderStageFlags(vk.ShaderStageComputeBit),
		}
	}

	createInfo := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: uint32(len(layoutBindings)),
		PBindings:    layoutBindings,
	}

	var layout vk.DescriptorSetLayout
	result := vk.CreateDescriptorSetLayout(p.ctx.device, &createInfo, nil, &layout)
	if result != vk.Success {
		return fmt.Errorf("vkCreateDescriptorSetLayout failed: %d", result)
	}

	p.descriptorLayout = layout
	return nil
}

// createPipelineLayout creates the pipeline layout with the descriptor set layout.
func (p *ComputePipeline) createPipelineLayout() error {
	createInfo := vk.PipelineLayoutCreateInfo{
		SType:          vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: 1,
		PSetLayouts:    []vk.DescriptorSetLayout{p.descriptorLayout},
	}

	var layout vk.PipelineLayout
	result := vk.CreatePipelineLayout(p.ctx.device, &createInfo, nil, &layout)
	if result != vk.Success {
		return fmt.Errorf("vkCreatePipelineLayout failed: %d", result)
	}

	p.pipelineLayout = layout
	return nil
}

// createComputePipeline creates the compute pipeline with the shader.
func (p *ComputePipeline) createComputePipeline() error {
	shaderStage := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageComputeBit,
		Module: p.shaderModule,
		PName:  safeString("main"),
	}

	createInfo := vk.ComputePipelineCreateInfo{
		SType:  vk.StructureTypeComputePipelineCreateInfo,
		Stage:  shaderStage,
		Layout: p.pipelineLayout,
	}

	pipelines := make([]vk.Pipeline, 1)
	result := vk.CreateComputePipelines(p.ctx.device, vk.NullPipelineCache, 1,
		[]vk.ComputePipelineCreateInfo{createInfo}, nil, pipelines)
	if result != vk.Success {
		return fmt.Errorf("vkCreateComputePipelines failed: %d", result)
	}

	p.pipeline = pipelines[0]
	return nil
}

// createDescriptorPoolAndSet creates the descriptor pool and allocates a set.
func (p *ComputePipeline) createDescriptorPoolAndSet() error {
	// Count descriptor types
	uniformCount := uint32(0)
	storageCount := uint32(0)
	for _, binding := range p.bindings {
		switch binding.Type {
		case BindingTypeUniform:
			uniformCount++
		case BindingTypeStorage:
			storageCount++
		}
	}

	// Create pool sizes
	var poolSizes []vk.DescriptorPoolSize
	if uniformCount > 0 {
		poolSizes = append(poolSizes, vk.DescriptorPoolSize{
			Type:            vk.DescriptorTypeUniformBuffer,
			DescriptorCount: uniformCount,
		})
	}
	if storageCount > 0 {
		poolSizes = append(poolSizes, vk.DescriptorPoolSize{
			Type:            vk.DescriptorTypeStorageBuffer,
			DescriptorCount: storageCount,
		})
	}

	poolInfo := vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		MaxSets:       1,
		PoolSizeCount: uint32(len(poolSizes)),
		PPoolSizes:    poolSizes,
	}

	var pool vk.DescriptorPool
	result := vk.CreateDescriptorPool(p.ctx.device, &poolInfo, nil, &pool)
	if result != vk.Success {
		return fmt.Errorf("vkCreateDescriptorPool failed: %d", result)
	}
	p.descriptorPool = pool

	// Allocate descriptor set
	allocInfo := vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     p.descriptorPool,
		DescriptorSetCount: 1,
		PSetLayouts:        []vk.DescriptorSetLayout{p.descriptorLayout},
	}

	var set vk.DescriptorSet
	result = vk.AllocateDescriptorSets(p.ctx.device, &allocInfo, &set)
	if result != vk.Success {
		return fmt.Errorf("vkAllocateDescriptorSets failed: %d", result)
	}
	p.descriptorSet = set

	return nil
}

// allocateCommandBuffer allocates a command buffer from the context's pool.
func (p *ComputePipeline) allocateCommandBuffer() error {
	allocInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        p.ctx.commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}

	buffers := make([]vk.CommandBuffer, 1)
	result := vk.AllocateCommandBuffers(p.ctx.device, &allocInfo, buffers)
	if result != vk.Success {
		return fmt.Errorf("vkAllocateCommandBuffers failed: %d", result)
	}

	p.commandBuffer = buffers[0]
	return nil
}

// createFence creates a fence for GPU synchronization.
func (p *ComputePipeline) createFence() error {
	fenceInfo := vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		// Start unsignaled - will be signaled after first dispatch
	}

	var fence vk.Fence
	result := vk.CreateFence(p.ctx.device, &fenceInfo, nil, &fence)
	if result != vk.Success {
		return fmt.Errorf("vkCreateFence failed: %d", result)
	}

	p.fence = fence
	return nil
}

// BindBuffer binds a GPU buffer to a descriptor binding.
// The binding must exist in the pipeline configuration.
func (p *ComputePipeline) BindBuffer(binding uint32, buffer *GPUBuffer) error {
	// Validate binding exists
	bindingInfo := p.findBinding(binding)
	if bindingInfo == nil {
		return fmt.Errorf("binding %d not found in pipeline config", binding)
	}

	var descriptorType vk.DescriptorType
	switch bindingInfo.Type {
	case BindingTypeUniform:
		descriptorType = vk.DescriptorTypeUniformBuffer
	case BindingTypeStorage:
		descriptorType = vk.DescriptorTypeStorageBuffer
	}

	bufferInfo := vk.DescriptorBufferInfo{
		Buffer: buffer.VkBuffer(),
		Offset: 0,
		Range:  vk.DeviceSize(buffer.Size()),
	}

	writeDescriptor := vk.WriteDescriptorSet{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          p.descriptorSet,
		DstBinding:      binding,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  descriptorType,
		PBufferInfo:     []vk.DescriptorBufferInfo{bufferInfo},
	}

	vk.UpdateDescriptorSets(p.ctx.device, 1, []vk.WriteDescriptorSet{writeDescriptor}, 0, nil)
	return nil
}

// SetUniformRaw writes raw data to a uniform buffer binding.
// Creates or updates an internal uniform buffer for this binding.
func (p *ComputePipeline) SetUniformRaw(binding uint32, data []byte) error {
	// Validate binding exists and is uniform type
	bindingInfo := p.findBinding(binding)
	if bindingInfo == nil {
		return fmt.Errorf("binding %d not found in pipeline config", binding)
	}
	if bindingInfo.Type != BindingTypeUniform {
		return fmt.Errorf("binding %d is not a uniform buffer", binding)
	}

	// Check data size
	if bindingInfo.Size > 0 && uint64(len(data)) != bindingInfo.Size {
		return fmt.Errorf("data size %d does not match expected size %d for binding %d",
			len(data), bindingInfo.Size, binding)
	}

	// Get or create uniform buffer
	ub, exists := p.uniformBuffers[binding]
	if !exists || ub.size < uint64(len(data)) {
		// Destroy old buffer if exists
		if exists {
			p.destroyUniformBuffer(ub)
		}

		// Create new buffer
		var err error
		ub, err = p.createUniformBuffer(uint64(len(data)))
		if err != nil {
			return fmt.Errorf("failed to create uniform buffer: %w", err)
		}
		p.uniformBuffers[binding] = ub

		// Bind to descriptor set
		if err := p.bindUniformBuffer(binding, ub); err != nil {
			return fmt.Errorf("failed to bind uniform buffer: %w", err)
		}
	}

	// Upload data to buffer
	if err := p.uploadToUniformBuffer(ub, data); err != nil {
		return fmt.Errorf("failed to upload uniform data: %w", err)
	}

	return nil
}

// createUniformBuffer creates an internal uniform buffer.
func (p *ComputePipeline) createUniformBuffer(size uint64) (*uniformBuffer, error) {
	bufferInfo := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        vk.DeviceSize(size),
		Usage:       vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit),
		SharingMode: vk.SharingModeExclusive,
	}

	var buffer vk.Buffer
	result := vk.CreateBuffer(p.ctx.device, &bufferInfo, nil, &buffer)
	if result != vk.Success {
		return nil, fmt.Errorf("vkCreateBuffer failed: %d", result)
	}

	// Get memory requirements
	var memRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(p.ctx.device, buffer, &memRequirements)
	memRequirements.Deref()

	// Find suitable memory type (host visible and coherent for easy updates)
	memTypeIndex, err := p.ctx.FindMemoryType(memRequirements.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		vk.DestroyBuffer(p.ctx.device, buffer, nil)
		return nil, err
	}

	// Allocate memory
	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memRequirements.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var memory vk.DeviceMemory
	result = vk.AllocateMemory(p.ctx.device, &allocInfo, nil, &memory)
	if result != vk.Success {
		vk.DestroyBuffer(p.ctx.device, buffer, nil)
		return nil, fmt.Errorf("vkAllocateMemory failed: %d", result)
	}

	// Bind memory to buffer
	result = vk.BindBufferMemory(p.ctx.device, buffer, memory, 0)
	if result != vk.Success {
		vk.FreeMemory(p.ctx.device, memory, nil)
		vk.DestroyBuffer(p.ctx.device, buffer, nil)
		return nil, fmt.Errorf("vkBindBufferMemory failed: %d", result)
	}

	return &uniformBuffer{
		buffer: buffer,
		memory: memory,
		size:   size,
	}, nil
}

// bindUniformBuffer binds an internal uniform buffer to the descriptor set.
func (p *ComputePipeline) bindUniformBuffer(binding uint32, ub *uniformBuffer) error {
	bufferInfo := vk.DescriptorBufferInfo{
		Buffer: ub.buffer,
		Offset: 0,
		Range:  vk.DeviceSize(ub.size),
	}

	writeDescriptor := vk.WriteDescriptorSet{
		SType:           vk.StructureTypeWriteDescriptorSet,
		DstSet:          p.descriptorSet,
		DstBinding:      binding,
		DstArrayElement: 0,
		DescriptorCount: 1,
		DescriptorType:  vk.DescriptorTypeUniformBuffer,
		PBufferInfo:     []vk.DescriptorBufferInfo{bufferInfo},
	}

	vk.UpdateDescriptorSets(p.ctx.device, 1, []vk.WriteDescriptorSet{writeDescriptor}, 0, nil)
	return nil
}

// uploadToUniformBuffer copies data to the uniform buffer.
func (p *ComputePipeline) uploadToUniformBuffer(ub *uniformBuffer, data []byte) error {
	var mapped unsafe.Pointer
	result := vk.MapMemory(p.ctx.device, ub.memory, 0, vk.DeviceSize(len(data)), 0, &mapped)
	if result != vk.Success {
		return fmt.Errorf("vkMapMemory failed: %d", result)
	}

	// Copy data
	dst := (*[1 << 30]byte)(mapped)[:len(data):len(data)]
	copy(dst, data)

	vk.UnmapMemory(p.ctx.device, ub.memory)
	return nil
}

// destroyUniformBuffer cleans up a uniform buffer.
func (p *ComputePipeline) destroyUniformBuffer(ub *uniformBuffer) {
	if ub.buffer != nil {
		vk.DestroyBuffer(p.ctx.device, ub.buffer, nil)
	}
	if ub.memory != nil {
		vk.FreeMemory(p.ctx.device, ub.memory, nil)
	}
}

// findBinding looks up a binding by index.
func (p *ComputePipeline) findBinding(binding uint32) *BindingInfo {
	for i := range p.bindings {
		if p.bindings[i].Binding == binding {
			return &p.bindings[i]
		}
	}
	return nil
}

// Dispatch executes the compute shader with the specified workgroup counts.
// This method blocks until the GPU completes execution.
func (p *ComputePipeline) Dispatch(groupCountX, groupCountY, groupCountZ uint32) error {
	// Reset command buffer
	result := vk.ResetCommandBuffer(p.commandBuffer, 0)
	if result != vk.Success {
		return fmt.Errorf("vkResetCommandBuffer failed: %d", result)
	}

	// Begin command buffer
	beginInfo := vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	}

	result = vk.BeginCommandBuffer(p.commandBuffer, &beginInfo)
	if result != vk.Success {
		return fmt.Errorf("vkBeginCommandBuffer failed: %d", result)
	}

	// Bind pipeline and descriptor set
	vk.CmdBindPipeline(p.commandBuffer, vk.PipelineBindPointCompute, p.pipeline)
	vk.CmdBindDescriptorSets(p.commandBuffer, vk.PipelineBindPointCompute,
		p.pipelineLayout, 0, 1, []vk.DescriptorSet{p.descriptorSet}, 0, nil)

	// Dispatch compute work
	vk.CmdDispatch(p.commandBuffer, groupCountX, groupCountY, groupCountZ)

	// End command buffer
	result = vk.EndCommandBuffer(p.commandBuffer)
	if result != vk.Success {
		return fmt.Errorf("vkEndCommandBuffer failed: %d", result)
	}

	// Reset fence before submit
	fences := []vk.Fence{p.fence}
	vk.ResetFences(p.ctx.device, 1, fences)

	// Submit to compute queue
	submitInfo := vk.SubmitInfo{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{p.commandBuffer},
	}

	result = vk.QueueSubmit(p.ctx.computeQueue, 1, []vk.SubmitInfo{submitInfo}, p.fence)
	if result != vk.Success {
		return fmt.Errorf("vkQueueSubmit failed: %d", result)
	}

	// Wait for completion
	result = vk.WaitForFences(p.ctx.device, 1, fences, vk.True, ^uint64(0))
	if result != vk.Success {
		return fmt.Errorf("vkWaitForFences failed: %d", result)
	}

	return nil
}

// DispatchAsync executes the compute shader without waiting for completion.
// Use WaitForCompletion to synchronize later.
func (p *ComputePipeline) DispatchAsync(groupCountX, groupCountY, groupCountZ uint32) error {
	// Reset command buffer
	result := vk.ResetCommandBuffer(p.commandBuffer, 0)
	if result != vk.Success {
		return fmt.Errorf("vkResetCommandBuffer failed: %d", result)
	}

	// Begin command buffer
	beginInfo := vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	}

	result = vk.BeginCommandBuffer(p.commandBuffer, &beginInfo)
	if result != vk.Success {
		return fmt.Errorf("vkBeginCommandBuffer failed: %d", result)
	}

	// Bind pipeline and descriptor set
	vk.CmdBindPipeline(p.commandBuffer, vk.PipelineBindPointCompute, p.pipeline)
	vk.CmdBindDescriptorSets(p.commandBuffer, vk.PipelineBindPointCompute,
		p.pipelineLayout, 0, 1, []vk.DescriptorSet{p.descriptorSet}, 0, nil)

	// Dispatch compute work
	vk.CmdDispatch(p.commandBuffer, groupCountX, groupCountY, groupCountZ)

	// End command buffer
	result = vk.EndCommandBuffer(p.commandBuffer)
	if result != vk.Success {
		return fmt.Errorf("vkEndCommandBuffer failed: %d", result)
	}

	// Reset fence before submit
	fences := []vk.Fence{p.fence}
	vk.ResetFences(p.ctx.device, 1, fences)

	// Submit to compute queue
	submitInfo := vk.SubmitInfo{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{p.commandBuffer},
	}

	result = vk.QueueSubmit(p.ctx.computeQueue, 1, []vk.SubmitInfo{submitInfo}, p.fence)
	if result != vk.Success {
		return fmt.Errorf("vkQueueSubmit failed: %d", result)
	}

	return nil
}

// WaitForCompletion waits for an async dispatch to complete.
func (p *ComputePipeline) WaitForCompletion() error {
	fences := []vk.Fence{p.fence}
	result := vk.WaitForFences(p.ctx.device, 1, fences, vk.True, ^uint64(0))
	if result != vk.Success {
		return fmt.Errorf("vkWaitForFences failed: %d", result)
	}
	return nil
}

// Destroy releases all Vulkan resources associated with this pipeline.
// Must be called before destroying the VulkanContext.
func (p *ComputePipeline) Destroy() {
	if p.ctx == nil || p.ctx.device == nil {
		return
	}

	// Wait for any pending work
	vk.DeviceWaitIdle(p.ctx.device)

	// Destroy uniform buffers
	for _, ub := range p.uniformBuffers {
		p.destroyUniformBuffer(ub)
	}
	p.uniformBuffers = nil

	// Destroy fence
	if p.fence != nil {
		vk.DestroyFence(p.ctx.device, p.fence, nil)
		p.fence = nil
	}

	// Command buffer is freed when pool is destroyed (owned by context)

	// Destroy descriptor pool (frees descriptor sets automatically)
	if p.descriptorPool != nil {
		vk.DestroyDescriptorPool(p.ctx.device, p.descriptorPool, nil)
		p.descriptorPool = nil
		p.descriptorSet = nil
	}

	// Destroy pipeline
	if p.pipeline != nil {
		vk.DestroyPipeline(p.ctx.device, p.pipeline, nil)
		p.pipeline = nil
	}

	// Destroy pipeline layout
	if p.pipelineLayout != nil {
		vk.DestroyPipelineLayout(p.ctx.device, p.pipelineLayout, nil)
		p.pipelineLayout = nil
	}

	// Destroy descriptor set layout
	if p.descriptorLayout != nil {
		vk.DestroyDescriptorSetLayout(p.ctx.device, p.descriptorLayout, nil)
		p.descriptorLayout = nil
	}

	// Destroy shader module
	if p.shaderModule != nil {
		vk.DestroyShaderModule(p.ctx.device, p.shaderModule, nil)
		p.shaderModule = nil
	}
}
