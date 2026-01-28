// Package compute provides workgroup dispatch helpers for GPU compute operations.
package compute

import (
	"fmt"

	vk "github.com/vulkan-go/vulkan"
)

// WorkgroupConfig holds workgroup dimensions for compute shader dispatch.
type WorkgroupConfig struct {
	LocalSizeX uint32 // Matches shader local_size_x
	LocalSizeY uint32 // Usually 1 for 1D work
	LocalSizeZ uint32 // Usually 1 for 1D work
}

// DefaultWorkgroup returns config for standard 256-wide workgroups.
// This matches the common shader pattern: layout(local_size_x = 256) in;
func DefaultWorkgroup() WorkgroupConfig {
	return WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1}
}

// CalculateDispatchSize computes number of workgroups needed
// to process 'totalElements' with given workgroup config.
// Returns (groupCountX, groupCountY, groupCountZ).
//
// Example:
//   totalElements = 1000, LocalSizeX = 256
//   Returns (4, 1, 1) - 4 workgroups of 256 = 1024 threads
func CalculateDispatchSize(totalElements uint32, config WorkgroupConfig) (uint32, uint32, uint32) {
	if totalElements == 0 {
		return 0, 0, 0
	}

	// Calculate groups needed, rounding up
	groupsX := (totalElements + config.LocalSizeX - 1) / config.LocalSizeX
	return groupsX, 1, 1
}

// CalculateDispatchSize2D computes workgroups for 2D workloads (like matrices).
// Returns (groupCountX, groupCountY, groupCountZ).
//
// Example:
//   width = 1024, height = 768, LocalSizeX = 16, LocalSizeY = 16
//   Returns (64, 48, 1) - 64×48 = 3072 workgroups
func CalculateDispatchSize2D(width, height uint32, config WorkgroupConfig) (uint32, uint32, uint32) {
	if width == 0 || height == 0 {
		return 0, 0, 0
	}

	groupsX := (width + config.LocalSizeX - 1) / config.LocalSizeX
	groupsY := (height + config.LocalSizeY - 1) / config.LocalSizeY
	return groupsX, groupsY, 1
}

// CreateFence creates a Vulkan fence for GPU synchronization.
// The fence is created in unsignaled state.
func (c *VulkanContext) CreateFence() (vk.Fence, error) {
	if !c.available {
		return nil, fmt.Errorf("Vulkan context not available")
	}

	fenceInfo := vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: 0, // Unsignaled state
	}

	var fence vk.Fence
	if result := vk.CreateFence(c.device, &fenceInfo, nil, &fence); result != vk.Success {
		return nil, fmt.Errorf("failed to create fence: %d", result)
	}

	return fence, nil
}

// WaitForFence waits for a fence to become signaled.
// timeout is in nanoseconds. Use vk.MaxUint64 for infinite timeout.
func (c *VulkanContext) WaitForFence(fence vk.Fence, timeout uint64) error {
	if !c.available {
		return fmt.Errorf("Vulkan context not available")
	}

	result := vk.WaitForFences(c.device, 1, []vk.Fence{fence}, vk.True, timeout)
	if result == vk.Timeout {
		return fmt.Errorf("fence wait timeout")
	}
	if result != vk.Success {
		return fmt.Errorf("fence wait failed: %d", result)
	}

	return nil
}

// ResetFence resets a fence to unsignaled state.
// Required before reusing the fence for another command buffer.
func (c *VulkanContext) ResetFence(fence vk.Fence) error {
	if !c.available {
		return fmt.Errorf("Vulkan context not available")
	}

	if result := vk.ResetFences(c.device, 1, []vk.Fence{fence}); result != vk.Success {
		return fmt.Errorf("failed to reset fence: %d", result)
	}

	return nil
}

// DestroyFence destroys a Vulkan fence.
func (c *VulkanContext) DestroyFence(fence vk.Fence) {
	if c.available && c.device != nil && fence != nil {
		vk.DestroyFence(c.device, fence, nil)
	}
}
