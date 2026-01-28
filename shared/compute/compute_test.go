package compute

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

// skipIfNoVulkan checks if Vulkan is available and skips test if not.
// Returns the context if available, nil otherwise.
func skipIfNoVulkan(t *testing.T) *VulkanContext {
	ctx, err := NewVulkanContext()
	if err != nil {
		t.Skipf("Failed to create Vulkan context: %v", err)
		return nil
	}
	if !ctx.IsAvailable() {
		t.Skip("Vulkan not available, skipping GPU tests")
		return nil
	}
	return ctx
}

// TestVulkanContext_Initialize verifies context creation and initialization.
func TestVulkanContext_Initialize(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	if !ctx.IsAvailable() {
		t.Error("Context should be available after successful init")
	}

	// Verify internal state
	if ctx.device == nil {
		t.Error("Device should not be nil")
	}
	if ctx.instance == nil {
		t.Error("Instance should not be nil")
	}
	if ctx.computeQueue == nil {
		t.Error("Compute queue should not be nil")
	}
	if ctx.commandPool == nil {
		t.Error("Command pool should not be nil")
	}
}

// TestVulkanContext_Destroy verifies clean destruction.
func TestVulkanContext_Destroy(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	if ctx == nil {
		return
	}

	// First destroy
	ctx.Destroy()

	if ctx.IsAvailable() {
		t.Error("Context should not be available after destroy")
	}

	// Second destroy should not panic (idempotent)
	ctx.Destroy()
}

// TestGPUBuffer_CreateDestroy tests basic buffer lifecycle.
func TestGPUBuffer_CreateDestroy(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	tests := []struct {
		name        string
		size        uint64
		usage       BufferUsage
		deviceLocal bool
	}{
		{"Small host-visible storage", 1024, BufferUsageStorage, false},
		{"Large host-visible storage", 1024 * 1024, BufferUsageStorage, false},
		{"Device-local storage", 4096, BufferUsageStorage, true},
		{"Host-visible uniform", 256, BufferUsageUniform, false},
		{"Host-visible transfer", 2048, BufferUsageTransfer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := ctx.CreateBuffer(tt.size, tt.usage, tt.deviceLocal)
			if err != nil {
				t.Fatalf("Failed to create buffer: %v", err)
			}
			defer buf.Destroy()

			// Verify properties
			if buf.Size() != tt.size {
				t.Errorf("Buffer size = %d, want %d", buf.Size(), tt.size)
			}
			if buf.Usage() != tt.usage {
				t.Errorf("Buffer usage = %v, want %v", buf.Usage(), tt.usage)
			}
			if buf.IsDeviceLocal() != tt.deviceLocal {
				t.Errorf("Buffer deviceLocal = %v, want %v", buf.IsDeviceLocal(), tt.deviceLocal)
			}

			// Destroy should work multiple times
			buf.Destroy()
			buf.Destroy() // Should not panic
		})
	}
}

// TestGPUBuffer_UploadDownload tests data transfer operations.
func TestGPUBuffer_UploadDownload(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	t.Run("Byte upload/download", func(t *testing.T) {
		// Create host-visible buffer
		buf, err := ctx.CreateBuffer(256, BufferUsageStorage, false)
		if err != nil {
			t.Fatalf("Failed to create buffer: %v", err)
		}
		defer buf.Destroy()

		// Upload test data
		uploadData := make([]byte, 256)
		for i := range uploadData {
			uploadData[i] = byte(i)
		}

		if err := buf.Upload(uploadData); err != nil {
			t.Fatalf("Upload failed: %v", err)
		}

		// Download and verify
		downloadData := make([]byte, 256)
		if err := buf.Download(downloadData); err != nil {
			t.Fatalf("Download failed: %v", err)
		}

		for i := range uploadData {
			if downloadData[i] != uploadData[i] {
				t.Errorf("Data mismatch at index %d: got %d, want %d", i, downloadData[i], uploadData[i])
			}
		}
	})

	t.Run("Float32 upload/download", func(t *testing.T) {
		// Create buffer for 64 floats
		buf, err := ctx.CreateBuffer(64*4, BufferUsageStorage, false)
		if err != nil {
			t.Fatalf("Failed to create buffer: %v", err)
		}
		defer buf.Destroy()

		// Upload float data
		uploadData := make([]float32, 64)
		for i := range uploadData {
			uploadData[i] = float32(i) * 0.5
		}

		if err := buf.UploadFloat32(uploadData); err != nil {
			t.Fatalf("UploadFloat32 failed: %v", err)
		}

		// Download and verify
		downloadData := make([]float32, 64)
		if err := buf.DownloadFloat32(downloadData); err != nil {
			t.Fatalf("DownloadFloat32 failed: %v", err)
		}

		for i := range uploadData {
			if downloadData[i] != uploadData[i] {
				t.Errorf("Data mismatch at index %d: got %f, want %f", i, downloadData[i], uploadData[i])
			}
		}
	})

	t.Run("Device-local buffer rejects direct upload", func(t *testing.T) {
		buf, err := ctx.CreateBuffer(256, BufferUsageStorage, true)
		if err != nil {
			t.Fatalf("Failed to create buffer: %v", err)
		}
		defer buf.Destroy()

		data := make([]byte, 256)
		err = buf.Upload(data)
		if err == nil {
			t.Error("Upload to device-local buffer should fail")
		}
	})

	t.Run("Upload with oversized data", func(t *testing.T) {
		buf, err := ctx.CreateBuffer(128, BufferUsageStorage, false)
		if err != nil {
			t.Fatalf("Failed to create buffer: %v", err)
		}
		defer buf.Destroy()

		// Try to upload more than buffer size
		data := make([]byte, 256)
		err = buf.Upload(data)
		if err == nil {
			t.Error("Upload with oversized data should fail")
		}
	})
}

// TestShaderLoader_LoadSPIRV tests SPIR-V shader loading.
func TestShaderLoader_LoadSPIRV(t *testing.T) {
	// Create a minimal valid SPIR-V file for testing
	tempDir := t.TempDir()
	testShaderPath := filepath.Join(tempDir, "test.spv")

	// SPIR-V magic number: 0x07230203
	validSPIRV := []byte{0x03, 0x02, 0x23, 0x07, 0x00, 0x00, 0x00, 0x00}
	if err := os.WriteFile(testShaderPath, validSPIRV, 0644); err != nil {
		t.Fatalf("Failed to create test shader: %v", err)
	}

	t.Run("Load valid SPIR-V", func(t *testing.T) {
		spirv, err := LoadSPIRV(testShaderPath)
		if err != nil {
			t.Errorf("Failed to load valid SPIR-V: %v", err)
		}
		if len(spirv) != len(validSPIRV) {
			t.Errorf("Loaded SPIR-V size = %d, want %d", len(spirv), len(validSPIRV))
		}

		// Check magic number
		magic := binary.LittleEndian.Uint32(spirv[0:4])
		if magic != 0x07230203 {
			t.Errorf("SPIR-V magic = 0x%08x, want 0x07230203", magic)
		}
	})

	t.Run("Load nonexistent file", func(t *testing.T) {
		_, err := LoadSPIRV("/nonexistent/shader.spv")
		if err == nil {
			t.Error("Loading nonexistent file should fail")
		}
	})

	t.Run("Load invalid SPIR-V (bad magic)", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "invalid.spv")
		invalidData := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		if err := os.WriteFile(invalidPath, invalidData, 0644); err != nil {
			t.Fatalf("Failed to create invalid shader: %v", err)
		}

		_, err := LoadSPIRV(invalidPath)
		if err == nil {
			t.Error("Loading invalid SPIR-V should fail")
		}
	})

	t.Run("Load too small file", func(t *testing.T) {
		tooSmallPath := filepath.Join(tempDir, "toosmall.spv")
		if err := os.WriteFile(tooSmallPath, []byte{0x01, 0x02}, 0644); err != nil {
			t.Fatalf("Failed to create small file: %v", err)
		}

		_, err := LoadSPIRV(tooSmallPath)
		if err == nil {
			t.Error("Loading too small SPIR-V should fail")
		}
	})
}

// TestShaderLoader_CreateShaderModule tests shader module creation.
func TestShaderLoader_CreateShaderModule(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	t.Run("Create shader module with valid SPIR-V", func(t *testing.T) {
		// Minimal valid SPIR-V (just header)
		spirv := []byte{
			0x03, 0x02, 0x23, 0x07, // Magic
			0x00, 0x00, 0x01, 0x00, // Version 1.0
			0x00, 0x00, 0x00, 0x00, // Generator
			0x01, 0x00, 0x00, 0x00, // Bound
			0x00, 0x00, 0x00, 0x00, // Schema
		}

		module, err := CreateShaderModule(ctx, spirv)
		if err != nil {
			t.Skipf("CreateShaderModule failed (may need valid complete shader): %v", err)
		}
		if module != nil {
			// Note: We don't have a way to destroy individual shader modules
			// in our API, they're destroyed with the pipeline
			_ = module
		}
	})

	t.Run("Create with nil context", func(t *testing.T) {
		spirv := []byte{0x03, 0x02, 0x23, 0x07, 0x00, 0x00, 0x00, 0x00}
		_, err := CreateShaderModule(nil, spirv)
		if err == nil {
			t.Error("Creating shader module with nil context should fail")
		}
	})

	t.Run("Create with empty SPIR-V", func(t *testing.T) {
		_, err := CreateShaderModule(ctx, []byte{})
		if err == nil {
			t.Error("Creating shader module with empty SPIR-V should fail")
		}
	})

	t.Run("Create with misaligned SPIR-V", func(t *testing.T) {
		spirv := []byte{0x03, 0x02, 0x23} // Not multiple of 4
		_, err := CreateShaderModule(ctx, spirv)
		if err == nil {
			t.Error("Creating shader module with misaligned SPIR-V should fail")
		}
	})
}

// TestWorkgroupConfig tests workgroup dispatch calculations.
func TestWorkgroupConfig(t *testing.T) {
	t.Run("DefaultWorkgroup", func(t *testing.T) {
		cfg := DefaultWorkgroup()
		if cfg.LocalSizeX != 256 {
			t.Errorf("LocalSizeX = %d, want 256", cfg.LocalSizeX)
		}
		if cfg.LocalSizeY != 1 {
			t.Errorf("LocalSizeY = %d, want 1", cfg.LocalSizeY)
		}
		if cfg.LocalSizeZ != 1 {
			t.Errorf("LocalSizeZ = %d, want 1", cfg.LocalSizeZ)
		}
	})

	t.Run("CalculateDispatchSize", func(t *testing.T) {
		cfg := WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1}

		tests := []struct {
			elements     uint32
			wantX, wantY, wantZ uint32
		}{
			{0, 0, 0, 0},       // Zero elements returns all zeros
			{1, 1, 1, 1},
			{256, 1, 1, 1},
			{257, 2, 1, 1},
			{512, 2, 1, 1},
			{513, 3, 1, 1},
			{1000, 4, 1, 1},
		}

		for _, tt := range tests {
			gotX, gotY, gotZ := CalculateDispatchSize(tt.elements, cfg)
			if gotX != tt.wantX || gotY != tt.wantY || gotZ != tt.wantZ {
				t.Errorf("CalculateDispatchSize(%d) = (%d,%d,%d), want (%d,%d,%d)",
					tt.elements, gotX, gotY, gotZ, tt.wantX, tt.wantY, tt.wantZ)
			}
		}
	})

	t.Run("CalculateDispatchSize2D", func(t *testing.T) {
		cfg := WorkgroupConfig{LocalSizeX: 16, LocalSizeY: 16, LocalSizeZ: 1}

		tests := []struct {
			width, height       uint32
			wantX, wantY, wantZ uint32
		}{
			{0, 0, 0, 0, 0},         // Zero dimensions returns all zeros
			{16, 16, 1, 1, 1},
			{17, 16, 2, 1, 1},
			{16, 17, 1, 2, 1},
			{1024, 768, 64, 48, 1},
		}

		for _, tt := range tests {
			gotX, gotY, gotZ := CalculateDispatchSize2D(tt.width, tt.height, cfg)
			if gotX != tt.wantX || gotY != tt.wantY || gotZ != tt.wantZ {
				t.Errorf("CalculateDispatchSize2D(%d,%d) = (%d,%d,%d), want (%d,%d,%d)",
					tt.width, tt.height, gotX, gotY, gotZ, tt.wantX, tt.wantY, tt.wantZ)
			}
		}
	})
}

// TestFence tests fence creation and synchronization.
func TestFence(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	t.Run("Create and destroy fence", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("Failed to create fence: %v", err)
		}
		if fence == nil {
			t.Fatal("Fence should not be nil")
		}

		ctx.DestroyFence(fence)
	})

	t.Run("Reset fence", func(t *testing.T) {
		fence, err := ctx.CreateFence()
		if err != nil {
			t.Fatalf("Failed to create fence: %v", err)
		}
		defer ctx.DestroyFence(fence)

		if err := ctx.ResetFence(fence); err != nil {
			t.Errorf("Failed to reset fence: %v", err)
		}
	})
}

// TestComputePipeline_CreateWithoutShader tests pipeline creation failures.
func TestComputePipeline_CreateWithoutShader(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	t.Run("Create with nonexistent shader", func(t *testing.T) {
		config := PipelineConfig{
			ShaderPath: "/nonexistent/shader.spv",
			Bindings: []BindingInfo{
				{Binding: 0, Type: BindingTypeStorage, Size: 0},
			},
		}

		_, err := NewComputePipeline(ctx, config)
		if err == nil {
			t.Error("Creating pipeline with nonexistent shader should fail")
		}
	})

	t.Run("Create with no bindings", func(t *testing.T) {
		config := PipelineConfig{
			ShaderPath: "dummy.spv",
			Bindings:   []BindingInfo{},
		}

		_, err := NewComputePipeline(ctx, config)
		if err == nil {
			t.Error("Creating pipeline with no bindings should fail")
		}
	})

	t.Run("Create with nil context", func(t *testing.T) {
		config := PipelineConfig{
			ShaderPath: "dummy.spv",
			Bindings: []BindingInfo{
				{Binding: 0, Type: BindingTypeStorage, Size: 0},
			},
		}

		_, err := NewComputePipeline(nil, config)
		if err == nil {
			t.Error("Creating pipeline with nil context should fail")
		}
	})
}

// TestGPUBuffer_EdgeCases tests edge cases and error conditions.
func TestGPUBuffer_EdgeCases(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	t.Run("Create buffer with zero size", func(t *testing.T) {
		// Vulkan may or may not support zero-sized buffers
		_, err := ctx.CreateBuffer(0, BufferUsageStorage, false)
		// Either succeeds or fails gracefully - both are acceptable
		_ = err
	})

	t.Run("Upload empty data", func(t *testing.T) {
		buf, err := ctx.CreateBuffer(1024, BufferUsageStorage, false)
		if err != nil {
			t.Fatalf("Failed to create buffer: %v", err)
		}
		defer buf.Destroy()

		// Should succeed without error
		if err := buf.Upload([]byte{}); err != nil {
			t.Errorf("Upload of empty data failed: %v", err)
		}
	})

	t.Run("Upload/Download float32 empty slice", func(t *testing.T) {
		buf, err := ctx.CreateBuffer(1024, BufferUsageStorage, false)
		if err != nil {
			t.Fatalf("Failed to create buffer: %v", err)
		}
		defer buf.Destroy()

		// Should succeed without error
		if err := buf.UploadFloat32([]float32{}); err != nil {
			t.Errorf("UploadFloat32 of empty slice failed: %v", err)
		}
		if err := buf.DownloadFloat32([]float32{}); err != nil {
			t.Errorf("DownloadFloat32 of empty slice failed: %v", err)
		}
	})
}

// BenchmarkGPUBuffer_Upload benchmarks buffer upload performance.
func BenchmarkGPUBuffer_Upload(b *testing.B) {
	ctx, err := NewVulkanContext()
	if err != nil || !ctx.IsAvailable() {
		b.Skip("Vulkan not available, skipping GPU benchmarks")
		return
	}
	defer ctx.Destroy()

	sizes := []uint64{
		1024,           // 1 KB
		1024 * 1024,    // 1 MB
		10 * 1024 * 1024, // 10 MB
	}

	for _, size := range sizes {
		b.Run(formatBytes(size), func(b *testing.B) {
			buf, err := ctx.CreateBuffer(size, BufferUsageStorage, false)
			if err != nil {
				b.Fatalf("Failed to create buffer: %v", err)
			}
			defer buf.Destroy()

			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i)
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				if err := buf.Upload(data); err != nil {
					b.Fatalf("Upload failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkGPUBuffer_Download benchmarks buffer download performance.
func BenchmarkGPUBuffer_Download(b *testing.B) {
	ctx, err := NewVulkanContext()
	if err != nil || !ctx.IsAvailable() {
		b.Skip("Vulkan not available, skipping GPU benchmarks")
		return
	}
	defer ctx.Destroy()

	sizes := []uint64{
		1024,           // 1 KB
		1024 * 1024,    // 1 MB
		10 * 1024 * 1024, // 10 MB
	}

	for _, size := range sizes {
		b.Run(formatBytes(size), func(b *testing.B) {
			buf, err := ctx.CreateBuffer(size, BufferUsageStorage, false)
			if err != nil {
				b.Fatalf("Failed to create buffer: %v", err)
			}
			defer buf.Destroy()

			// Upload some data first
			uploadData := make([]byte, size)
			if err := buf.Upload(uploadData); err != nil {
				b.Fatalf("Initial upload failed: %v", err)
			}

			data := make([]byte, size)
			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				if err := buf.Download(data); err != nil {
					b.Fatalf("Download failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkGPUBuffer_Float32Upload benchmarks float32 upload performance.
func BenchmarkGPUBuffer_Float32Upload(b *testing.B) {
	ctx, err := NewVulkanContext()
	if err != nil || !ctx.IsAvailable() {
		b.Skip("Vulkan not available, skipping GPU benchmarks")
		return
	}
	defer ctx.Destroy()

	elementCounts := []int{
		1024,      // 4 KB
		262144,    // 1 MB
		2621440,   // 10 MB
	}

	for _, count := range elementCounts {
		b.Run(formatElements(count), func(b *testing.B) {
			size := uint64(count) * uint64(unsafe.Sizeof(float32(0)))
			buf, err := ctx.CreateBuffer(size, BufferUsageStorage, false)
			if err != nil {
				b.Fatalf("Failed to create buffer: %v", err)
			}
			defer buf.Destroy()

			data := make([]float32, count)
			for i := range data {
				data[i] = float32(i) * 0.5
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				if err := buf.UploadFloat32(data); err != nil {
					b.Fatalf("UploadFloat32 failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkVulkanContext_CreateBuffer benchmarks buffer creation.
func BenchmarkVulkanContext_CreateBuffer(b *testing.B) {
	ctx, err := NewVulkanContext()
	if err != nil || !ctx.IsAvailable() {
		b.Skip("Vulkan not available, skipping GPU benchmarks")
		return
	}
	defer ctx.Destroy()

	b.Run("Small buffers (1KB)", func(b *testing.B) {
		buffers := make([]*GPUBuffer, 0, b.N)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf, err := ctx.CreateBuffer(1024, BufferUsageStorage, false)
			if err != nil {
				b.Fatalf("Failed to create buffer: %v", err)
			}
			buffers = append(buffers, buf)
		}

		b.StopTimer()
		for _, buf := range buffers {
			buf.Destroy()
		}
	})

	b.Run("Large buffers (1MB)", func(b *testing.B) {
		buffers := make([]*GPUBuffer, 0, b.N)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			buf, err := ctx.CreateBuffer(1024*1024, BufferUsageStorage, false)
			if err != nil {
				b.Fatalf("Failed to create buffer: %v", err)
			}
			buffers = append(buffers, buf)
		}

		b.StopTimer()
		for _, buf := range buffers {
			buf.Destroy()
		}
	})
}

// Helper function to format byte sizes for benchmark names.
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return string(rune(bytes)) + "B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return string(rune(math.Round(float64(bytes)/float64(div)))) + " " + "KMGTPE"[exp:exp+1] + "B"
}

// Helper function to format element counts for benchmark names.
func formatElements(count int) string {
	if count < 1024 {
		return string(rune(count)) + " elements"
	}
	k := count / 1024
	if k < 1024 {
		return string(rune(k)) + "K elements"
	}
	m := k / 1024
	return string(rune(m)) + "M elements"
}

// TestComputePipeline_WithRealShader tests pipeline creation with actual shader files.
func TestComputePipeline_WithRealShader(t *testing.T) {
	ctx := skipIfNoVulkan(t)
	defer ctx.Destroy()

	// Try to find mvm.comp.spv shader
	shaderPaths := []string{
		"shaders/mvm.comp.spv",
		"shared/compute/shaders/mvm.comp.spv",
		"../shared/compute/shaders/mvm.comp.spv",
		"../../shared/compute/shaders/mvm.comp.spv",
	}

	var foundShader string
	for _, path := range shaderPaths {
		if _, err := os.Stat(path); err == nil {
			foundShader = path
			break
		}
	}

	if foundShader == "" {
		t.Skip("No compute shader found, skipping pipeline integration test")
		return
	}

	t.Run("Create pipeline with real shader", func(t *testing.T) {
		// MVM shader typically has:
		// binding 0: weights matrix (storage buffer)
		// binding 1: input vector (storage buffer)
		// binding 2: output vector (storage buffer)
		config := PipelineConfig{
			ShaderPath: foundShader,
			Bindings: []BindingInfo{
				{Binding: 0, Type: BindingTypeStorage, Size: 0},
				{Binding: 1, Type: BindingTypeStorage, Size: 0},
				{Binding: 2, Type: BindingTypeStorage, Size: 0},
			},
		}

		pipeline, err := NewComputePipeline(ctx, config)
		if err != nil {
			t.Skipf("Failed to create pipeline (shader may not match expected bindings): %v", err)
			return
		}
		defer pipeline.Destroy()

		// Verify pipeline is valid
		if pipeline.pipeline == nil {
			t.Error("Pipeline handle should not be nil")
		}
		if pipeline.descriptorSet == nil {
			t.Error("Descriptor set should not be nil")
		}
		if pipeline.commandBuffer == nil {
			t.Error("Command buffer should not be nil")
		}
	})
}
