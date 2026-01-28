// Package compute provides headless Vulkan compute context for GPU-accelerated operations.
package compute

import (
	"fmt"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// BufferUsage defines how a buffer will be used
type BufferUsage uint32

const (
	BufferUsageStorage  BufferUsage = iota // SSBO for compute shaders
	BufferUsageUniform                     // Uniform buffer
	BufferUsageTransfer                    // Staging buffer for upload/download
)

// GPUBuffer wraps Vulkan buffer with memory management
type GPUBuffer struct {
	ctx         *VulkanContext
	buffer      vk.Buffer
	memory      vk.DeviceMemory
	size        vk.DeviceSize
	usage       BufferUsage
	deviceLocal bool
	mapped      unsafe.Pointer // For host-visible buffers
}

// CreateBuffer creates a GPU buffer with appropriate usage flags and memory type.
//
// Parameters:
//   - size: Buffer size in bytes
//   - usage: How the buffer will be used (Storage, Uniform, or Transfer)
//   - deviceLocal: If true, allocates in device-local memory (faster GPU access, no CPU access)
//     If false, allocates in host-visible memory (slower GPU access, but CPU can read/write)
//
// For compute workloads:
//   - Use deviceLocal=true for input/output buffers that stay on GPU
//   - Use deviceLocal=false for staging buffers that need CPU upload/download
func (c *VulkanContext) CreateBuffer(size uint64, usage BufferUsage, deviceLocal bool) (*GPUBuffer, error) {
	if !c.available {
		return nil, fmt.Errorf("Vulkan context not available")
	}

	buf := &GPUBuffer{
		ctx:         c,
		size:        vk.DeviceSize(size),
		usage:       usage,
		deviceLocal: deviceLocal,
	}

	// Determine Vulkan buffer usage flags
	var vkUsage vk.BufferUsageFlags
	switch usage {
	case BufferUsageStorage:
		vkUsage = vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit | vk.BufferUsageTransferSrcBit | vk.BufferUsageTransferDstBit)
	case BufferUsageUniform:
		vkUsage = vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit | vk.BufferUsageTransferDstBit)
	case BufferUsageTransfer:
		vkUsage = vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit | vk.BufferUsageTransferDstBit)
	default:
		return nil, fmt.Errorf("invalid buffer usage: %d", usage)
	}

	// Create buffer
	bufferInfo := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        buf.size,
		Usage:       vkUsage,
		SharingMode: vk.SharingModeExclusive,
	}

	var buffer vk.Buffer
	if result := vk.CreateBuffer(c.device, &bufferInfo, nil, &buffer); result != vk.Success {
		return nil, fmt.Errorf("failed to create buffer: %d", result)
	}
	buf.buffer = buffer

	// Get memory requirements
	var memRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(c.device, buf.buffer, &memRequirements)
	memRequirements.Deref()

	// Determine memory property flags
	var memProperties vk.MemoryPropertyFlags
	if deviceLocal {
		memProperties = vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit)
	} else {
		// Host-visible and coherent for CPU access
		memProperties = vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit)
	}

	// Find suitable memory type
	memTypeIndex, err := c.FindMemoryType(memRequirements.MemoryTypeBits, memProperties)
	if err != nil {
		vk.DestroyBuffer(c.device, buf.buffer, nil)
		return nil, fmt.Errorf("failed to find suitable memory type: %w", err)
	}

	// Allocate memory
	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memRequirements.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var memory vk.DeviceMemory
	if result := vk.AllocateMemory(c.device, &allocInfo, nil, &memory); result != vk.Success {
		vk.DestroyBuffer(c.device, buf.buffer, nil)
		return nil, fmt.Errorf("failed to allocate buffer memory: %d", result)
	}
	buf.memory = memory

	// Bind buffer memory
	if result := vk.BindBufferMemory(c.device, buf.buffer, buf.memory, 0); result != vk.Success {
		vk.FreeMemory(c.device, buf.memory, nil)
		vk.DestroyBuffer(c.device, buf.buffer, nil)
		return nil, fmt.Errorf("failed to bind buffer memory: %d", result)
	}

	// For host-visible buffers, map memory persistently
	if !deviceLocal {
		var data unsafe.Pointer
		if result := vk.MapMemory(c.device, buf.memory, 0, buf.size, 0, &data); result != vk.Success {
			vk.FreeMemory(c.device, buf.memory, nil)
			vk.DestroyBuffer(c.device, buf.buffer, nil)
			return nil, fmt.Errorf("failed to map buffer memory: %d", result)
		}
		buf.mapped = data
	}

	return buf, nil
}

// Upload copies data from CPU to GPU buffer.
// Only works for host-visible buffers (deviceLocal=false).
// For device-local buffers, use a staging buffer with CopyBuffer.
func (b *GPUBuffer) Upload(data []byte) error {
	if b.deviceLocal {
		return fmt.Errorf("cannot upload directly to device-local buffer; use staging buffer")
	}

	if b.mapped == nil {
		return fmt.Errorf("buffer not mapped")
	}

	if uint64(len(data)) > uint64(b.size) {
		return fmt.Errorf("data size (%d bytes) exceeds buffer size (%d bytes)", len(data), b.size)
	}

	// Copy data using unsafe.Slice for efficient memory copy
	dst := unsafe.Slice((*byte)(b.mapped), b.size)
	copy(dst, data)

	return nil
}

// Download copies data from GPU buffer to CPU.
// Only works for host-visible buffers (deviceLocal=false).
// For device-local buffers, use a staging buffer with CopyBuffer.
func (b *GPUBuffer) Download(data []byte) error {
	if b.deviceLocal {
		return fmt.Errorf("cannot download directly from device-local buffer; use staging buffer")
	}

	if b.mapped == nil {
		return fmt.Errorf("buffer not mapped")
	}

	if uint64(len(data)) > uint64(b.size) {
		return fmt.Errorf("data size (%d bytes) exceeds buffer size (%d bytes)", len(data), b.size)
	}

	// Copy data using unsafe.Slice for efficient memory copy
	src := unsafe.Slice((*byte)(b.mapped), b.size)
	copy(data, src)

	return nil
}

// UploadFloat32 is a convenience wrapper for uploading float32 slices.
func (b *GPUBuffer) UploadFloat32(data []float32) error {
	if len(data) == 0 {
		return nil
	}

	// Convert float32 slice to byte slice using unsafe
	byteSize := len(data) * int(unsafe.Sizeof(float32(0)))
	if uint64(byteSize) > uint64(b.size) {
		return fmt.Errorf("data size (%d bytes) exceeds buffer size (%d bytes)", byteSize, b.size)
	}

	byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), byteSize)
	return b.Upload(byteSlice)
}

// DownloadFloat32 is a convenience wrapper for downloading float32 slices.
func (b *GPUBuffer) DownloadFloat32(data []float32) error {
	if len(data) == 0 {
		return nil
	}

	// Convert float32 slice to byte slice using unsafe
	byteSize := len(data) * int(unsafe.Sizeof(float32(0)))
	if uint64(byteSize) > uint64(b.size) {
		return fmt.Errorf("data size (%d bytes) exceeds buffer size (%d bytes)", byteSize, b.size)
	}

	byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), byteSize)
	return b.Download(byteSlice)
}

// Destroy releases the buffer and its memory.
// Must be called when the buffer is no longer needed.
func (b *GPUBuffer) Destroy() {
	if b.ctx.device == nil {
		return
	}

	// Unmap memory if it was mapped
	if b.mapped != nil {
		vk.UnmapMemory(b.ctx.device, b.memory)
		b.mapped = nil
	}

	// Free memory
	if b.memory != nil {
		vk.FreeMemory(b.ctx.device, b.memory, nil)
		b.memory = nil
	}

	// Destroy buffer
	if b.buffer != nil {
		vk.DestroyBuffer(b.ctx.device, b.buffer, nil)
		b.buffer = nil
	}
}

// Size returns the buffer size in bytes.
func (b *GPUBuffer) Size() uint64 {
	return uint64(b.size)
}

// IsDeviceLocal returns true if the buffer is allocated in device-local memory.
func (b *GPUBuffer) IsDeviceLocal() bool {
	return b.deviceLocal
}

// Usage returns the buffer usage type.
func (b *GPUBuffer) Usage() BufferUsage {
	return b.usage
}

// VkBuffer returns the underlying Vulkan buffer handle.
// This is useful for advanced operations like descriptor sets or buffer copies.
func (b *GPUBuffer) VkBuffer() vk.Buffer {
	return b.buffer
}
