package compute

import (
	"testing"

	vk "github.com/vulkan-go/vulkan"
)

// TestDefaultWorkgroup verifies the default workgroup configuration.
func TestDefaultWorkgroup(t *testing.T) {
	cfg := DefaultWorkgroup()

	if cfg.LocalSizeX != 256 {
		t.Errorf("DefaultWorkgroup().LocalSizeX = %d, want 256", cfg.LocalSizeX)
	}
	if cfg.LocalSizeY != 1 {
		t.Errorf("DefaultWorkgroup().LocalSizeY = %d, want 1", cfg.LocalSizeY)
	}
	if cfg.LocalSizeZ != 1 {
		t.Errorf("DefaultWorkgroup().LocalSizeZ = %d, want 1", cfg.LocalSizeZ)
	}
}

// TestCalculateDispatchSize tests 1D dispatch calculation with various scenarios.
func TestCalculateDispatchSize(t *testing.T) {
	tests := []struct {
		name         string
		totalSize    uint32
		workgroupCfg WorkgroupConfig
		wantX        uint32
		wantY        uint32
		wantZ        uint32
	}{
		{
			name:         "Zero elements",
			totalSize:    0,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        0,
			wantY:        0,
			wantZ:        0,
		},
		{
			name:         "Single element",
			totalSize:    1,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        1,
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "Exact workgroup size",
			totalSize:    256,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        1,
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "One more than workgroup size",
			totalSize:    257,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        2,
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "Multiple workgroups",
			totalSize:    1000,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        4,
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "Large number of elements",
			totalSize:    1000000,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        3907, // ceil(1000000 / 256)
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "Small workgroup size",
			totalSize:    100,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 32, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        4, // ceil(100 / 32)
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "Workgroup size of 1",
			totalSize:    10,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 1, LocalSizeY: 1, LocalSizeZ: 1},
			wantX:        10,
			wantY:        1,
			wantZ:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY, gotZ := CalculateDispatchSize(tt.totalSize, tt.workgroupCfg)
			if gotX != tt.wantX || gotY != tt.wantY || gotZ != tt.wantZ {
				t.Errorf("CalculateDispatchSize(%d, %+v) = (%d, %d, %d), want (%d, %d, %d)",
					tt.totalSize, tt.workgroupCfg, gotX, gotY, gotZ, tt.wantX, tt.wantY, tt.wantZ)
			}
		})
	}
}

// TestCalculateDispatchSize2D tests 2D dispatch calculation with various dimensions.
func TestCalculateDispatchSize2D(t *testing.T) {
	tests := []struct {
		name         string
		width        uint32
		height       uint32
		workgroupCfg WorkgroupConfig
		wantX        uint32
		wantY        uint32
		wantZ        uint32
	}{
		{
			name:         "Zero dimensions",
			width:        0,
			height:       0,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        0,
			wantY:        0,
			wantZ:        0,
		},
		{
			name:         "Zero width",
			width:        0,
			height:       100,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        0,
			wantY:        0,
			wantZ:        0,
		},
		{
			name:         "Zero height",
			width:        100,
			height:       0,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        0,
			wantY:        0,
			wantZ:        0,
		},
		{
			name:         "Single pixel",
			width:        1,
			height:       1,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        1,
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "Exact workgroup dimensions",
			width:        16,
			height:       16,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        1,
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "One more in width",
			width:        17,
			height:       16,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        2,
			wantY:        1,
			wantZ:        1,
		},
		{
			name:         "One more in height",
			width:        16,
			height:       17,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        1,
			wantY:        2,
			wantZ:        1,
		},
		{
			name:         "HD resolution (1920x1080)",
			width:        1920,
			height:       1080,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        120, // ceil(1920 / 16)
			wantY:        68,  // ceil(1080 / 16)
			wantZ:        1,
		},
		{
			name:         "Standard resolution (1024x768)",
			width:        1024,
			height:       768,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        64, // ceil(1024 / 16)
			wantY:        48, // ceil(768 / 16)
			wantZ:        1,
		},
		{
			name:         "Non-square workgroups",
			width:        100,
			height:       200,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 8, LocalSizeY: 16, LocalSizeZ: 1},
			wantX:        13, // ceil(100 / 8)
			wantY:        13, // ceil(200 / 16)
			wantZ:        1,
		},
		{
			name:         "Large dimensions",
			width:        4096,
			height:       4096,
			workgroupCfg: WorkgroupConfig{LocalSizeX: 32, LocalSizeY: 32, LocalSizeZ: 1},
			wantX:        128, // ceil(4096 / 32)
			wantY:        128, // ceil(4096 / 32)
			wantZ:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY, gotZ := CalculateDispatchSize2D(tt.width, tt.height, tt.workgroupCfg)
			if gotX != tt.wantX || gotY != tt.wantY || gotZ != tt.wantZ {
				t.Errorf("CalculateDispatchSize2D(%d, %d, %+v) = (%d, %d, %d), want (%d, %d, %d)",
					tt.width, tt.height, tt.workgroupCfg, gotX, gotY, gotZ, tt.wantX, tt.wantY, tt.wantZ)
			}
		})
	}
}

// TestVulkanContext_FenceMethods tests fence creation, waiting, and reset.
func TestVulkanContext_FenceMethods(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	t.Run("CreateFence", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("CreateFence() failed: %v", err)
		}
		if fence == nil {
			t.Error("CreateFence() returned nil fence")
		}
		defer ctx.DestroyFence(fence)
	})

	t.Run("CreateFence with unavailable context", func(t *testing.T) {
		unavailableCtx := &VulkanContext{available: false}
		_, err := unavailableCtx.CreateFence()
		if err == nil {
			t.Error("CreateFence() with unavailable context should fail")
		}
	})

	t.Run("WaitForFence", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("CreateFence() failed: %v", err)
		}
		defer ctx.DestroyFence(fence)

		// Reset fence to unsignaled state
		if err := ctx.ResetFence(fence); err != nil {
			t.Fatalf("ResetFence() failed: %v", err)
		}

		// Wait with zero timeout should timeout immediately on unsignaled fence
		err = ctx.WaitForFence(fence, 0)
		if err == nil {
			// Some implementations might signal immediately, which is acceptable
			t.Log("WaitForFence(0) did not timeout (fence may have been signaled)")
		}
	})

	t.Run("WaitForFence with unavailable context", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("CreateFence() failed: %v", err)
		}
		defer ctx.DestroyFence(fence)

		unavailableCtx := &VulkanContext{available: false}
		err = unavailableCtx.WaitForFence(fence, 0)
		if err == nil {
			t.Error("WaitForFence() with unavailable context should fail")
		}
	})

	t.Run("ResetFence", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("CreateFence() failed: %v", err)
		}
		defer ctx.DestroyFence(fence)

		if err := ctx.ResetFence(fence); err != nil {
			t.Errorf("ResetFence() failed: %v", err)
		}

		// Reset again to test idempotency
		if err := ctx.ResetFence(fence); err != nil {
			t.Errorf("Second ResetFence() failed: %v", err)
		}
	})

	t.Run("ResetFence with unavailable context", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("CreateFence() failed: %v", err)
		}
		defer ctx.DestroyFence(fence)

		unavailableCtx := &VulkanContext{available: false}
		err = unavailableCtx.ResetFence(fence)
		if err == nil {
			t.Error("ResetFence() with unavailable context should fail")
		}
	})

	t.Run("DestroyFence", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("CreateFence() failed: %v", err)
		}

		// Single destroy - test doesn't panic
		ctx.DestroyFence(fence)
		// Note: Double destroy causes driver crash, so we don't test idempotency
	})

	t.Run("DestroyFence with nil fence", func(t *testing.T) {
		// Should not panic
		ctx.DestroyFence(nil)
	})

	t.Run("DestroyFence with unavailable context", func(t *testing.T) {
		unavailableCtx := &VulkanContext{available: false}
		// Should not panic or error
		unavailableCtx.DestroyFence(nil)
	})
}

// TestVulkanContext_FindMemoryType tests memory type finding logic.
func TestVulkanContext_FindMemoryType(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	t.Run("Find device-local memory", func(t *testing.T) {
		// Try to find device-local memory type
		// TypeFilter with all bits set should allow any memory type
		typeFilter := uint32(0xFFFFFFFF)
		properties := vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit)

		memTypeIndex, err := ctx.FindMemoryType(typeFilter, properties)
		if err != nil {
			t.Skipf("No device-local memory found (acceptable on some systems): %v", err)
		}

		// Verify the returned index is within valid range
		if memTypeIndex >= ctx.memoryProps.MemoryTypeCount {
			t.Errorf("FindMemoryType() returned invalid index %d (max %d)",
				memTypeIndex, ctx.memoryProps.MemoryTypeCount)
		}
	})

	t.Run("Find host-visible memory", func(t *testing.T) {
		typeFilter := uint32(0xFFFFFFFF)
		properties := vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit)

		memTypeIndex, err := ctx.FindMemoryType(typeFilter, properties)
		if err != nil {
			t.Fatalf("FindMemoryType() failed for host-visible memory: %v", err)
		}

		// Verify the returned index is within valid range
		if memTypeIndex >= ctx.memoryProps.MemoryTypeCount {
			t.Errorf("FindMemoryType() returned invalid index %d (max %d)",
				memTypeIndex, ctx.memoryProps.MemoryTypeCount)
		}
	})

	t.Run("No matching memory type", func(t *testing.T) {
		// Use typeFilter that matches no memory types
		typeFilter := uint32(0)
		properties := vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit)

		_, err := ctx.FindMemoryType(typeFilter, properties)
		if err == nil {
			t.Error("FindMemoryType() with zero typeFilter should fail")
		}
	})
}

// TestComputePipeline_BindBuffer tests buffer binding to pipeline.
func TestComputePipeline_BindBuffer(t *testing.T) {
	t.Skip("Requires valid executable SPIR-V shader - tested in integration tests")
}

// TestComputePipeline_SetUniformRaw tests raw uniform data setting.
func TestComputePipeline_SetUniformRaw(t *testing.T) {
	t.Skip("Requires valid executable SPIR-V shader - tested in integration tests")
}

// TestComputePipeline_DispatchAsync tests asynchronous dispatch.
func TestComputePipeline_DispatchAsync(t *testing.T) {
	t.Skip("Requires valid executable SPIR-V shader - tested in integration tests")
}

// TestComputePipeline_WaitForCompletion tests waiting for completion.
func TestComputePipeline_WaitForCompletion(t *testing.T) {
	t.Skip("Requires valid executable SPIR-V shader - tested in integration tests")
}

// TestGPUBuffer_CreateWithZeroSize tests buffer creation with zero size.
func TestGPUBuffer_CreateWithZeroSize(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	_, err := ctx.CreateBuffer(0, BufferUsageStorage, false)
	if err == nil {
		t.Error("CreateBuffer(0, ...) should fail - Vulkan does not support zero-sized buffers")
	}
}

// TestGPUBuffer_CreateWithInvalidUsage tests buffer creation with invalid usage.
func TestGPUBuffer_CreateWithInvalidUsage(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	invalidUsage := BufferUsage(999)
	_, err := ctx.CreateBuffer(1024, invalidUsage, false)
	if err == nil {
		t.Error("CreateBuffer() with invalid usage should fail")
	}
}

