//go:build !js
// +build !js

// Package crossbar provides GPU-accelerated Matrix-Vector Multiplication for crossbar arrays.
package crossbar

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"unsafe"

	"fecim-lattice-tools/shared/compute"
)

// GPUAccelerator provides GPU-accelerated MVM operations for crossbar arrays.
type GPUAccelerator struct {
	ctx         *compute.VulkanContext
	mvmPipeline *compute.ComputePipeline

	// Reusable buffers (sized for common operations)
	conductanceBuffer *compute.GPUBuffer
	inputBuffer       *compute.GPUBuffer
	outputBuffer      *compute.GPUBuffer
	variationBuffer   *compute.GPUBuffer

	maxRows int
	maxCols int
}

// CrossbarParams matches the shader's uniform buffer layout (std140).
// Must match mvm.comp binding 0 EXACTLY.
//
// std140 layout rules:
// - int32/float32: 4-byte alignment, 4-byte size
// - All members must follow std140 padding rules
type CrossbarParams struct {
	Rows           int32   // Number of rows (offset 0)
	Cols           int32   // Number of columns (offset 4)
	NoiseLevel     float32 // Device noise level (offset 8)
	ADCBits        int32   // ADC resolution (offset 12)
	DACBits        int32   // DAC resolution (offset 16)
	Time           float32 // Simulation time (offset 20)
	WireResistance float32 // Line resistance (offset 24)
	DriftCoeff     float32 // Drift coefficient (offset 28)
	// Total size: 32 bytes
}

// NewGPUAccelerator creates a GPU accelerator for MVM operations.
// If GPU is not available, returns (nil, nil) as a fallback indicator.
//
// Parameters:
//   - maxRows: Maximum number of rows in crossbar (input vector size)
//   - maxCols: Maximum number of columns in crossbar (output vector size)
func NewGPUAccelerator(maxRows, maxCols int) (*GPUAccelerator, error) {
	if maxRows <= 0 || maxCols <= 0 {
		return nil, fmt.Errorf("invalid dimensions: rows=%d, cols=%d", maxRows, maxCols)
	}

	// Find repository root for absolute shader path
	repoRoot, err := findRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Create Vulkan context
	ctx, err := compute.NewVulkanContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create Vulkan context: %w", err)
	}

	// Check if GPU is available
	if !ctx.IsAvailable() {
		// Not an error - just means GPU compute is not available
		// Caller should fall back to CPU implementation
		ctx.Destroy()
		return nil, nil
	}

	g := &GPUAccelerator{
		ctx:     ctx,
		maxRows: maxRows,
		maxCols: maxCols,
	}

	// Create MVM compute pipeline
	if err := g.createMVMPipeline(repoRoot); err != nil {
		g.Destroy()
		return nil, fmt.Errorf("failed to create MVM pipeline: %w", err)
	}

	// Pre-allocate buffers for maximum array size
	if err := g.allocateBuffers(); err != nil {
		g.Destroy()
		return nil, fmt.Errorf("failed to allocate buffers: %w", err)
	}

	return g, nil
}

// createMVMPipeline creates the compute pipeline for mvm.comp shader.
func (g *GPUAccelerator) createMVMPipeline(repoRoot string) error {
	// The mvm.comp shader expects:
	// - binding 0: uniform CrossbarParams (std140)
	// - binding 1: storage ConductanceBuffer (G matrix, row-major)
	// - binding 2: storage InputVoltageBuffer (V vector)
	// - binding 3: storage OutputCurrentBuffer (I vector)
	// - binding 4: storage VariationBuffer (device-to-device variation)
	config := compute.PipelineConfig{
		ShaderPath: filepath.Join(repoRoot, "shared/compute/shaders/mvm.comp.spv"),
		Bindings: []compute.BindingInfo{
			{Binding: 0, Type: compute.BindingTypeUniform, Size: uint64(unsafe.Sizeof(CrossbarParams{}))}, // 32 bytes
			{Binding: 1, Type: compute.BindingTypeStorage, Size: 0},                                       // ConductanceBuffer (dynamic)
			{Binding: 2, Type: compute.BindingTypeStorage, Size: 0},                                       // InputVoltageBuffer (dynamic)
			{Binding: 3, Type: compute.BindingTypeStorage, Size: 0},                                       // OutputCurrentBuffer (dynamic)
			{Binding: 4, Type: compute.BindingTypeStorage, Size: 0},                                       // VariationBuffer (dynamic)
		},
	}

	pipeline, err := compute.NewComputePipeline(g.ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create compute pipeline: %w", err)
	}

	g.mvmPipeline = pipeline
	return nil
}

// allocateBuffers pre-allocates GPU buffers for maximum array size.
func (g *GPUAccelerator) allocateBuffers() error {
	// Calculate buffer sizes
	conductanceSize := uint64(g.maxRows * g.maxCols * 4) // float32 = 4 bytes
	inputSize := uint64(g.maxRows * 4)
	outputSize := uint64(g.maxCols * 4)
	variationSize := conductanceSize // Same size as conductance matrix

	// Create host-visible buffers for easy CPU upload/download
	// These buffers are mapped persistently for zero-copy access
	var err error

	g.conductanceBuffer, err = g.ctx.CreateBuffer(conductanceSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create conductance buffer: %w", err)
	}

	g.inputBuffer, err = g.ctx.CreateBuffer(inputSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create input buffer: %w", err)
	}

	g.outputBuffer, err = g.ctx.CreateBuffer(outputSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create output buffer: %w", err)
	}

	g.variationBuffer, err = g.ctx.CreateBuffer(variationSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create variation buffer: %w", err)
	}

	// Bind buffers to pipeline descriptor set
	if err := g.mvmPipeline.BindBuffer(1, g.conductanceBuffer); err != nil {
		return fmt.Errorf("failed to bind conductance buffer: %w", err)
	}
	if err := g.mvmPipeline.BindBuffer(2, g.inputBuffer); err != nil {
		return fmt.Errorf("failed to bind input buffer: %w", err)
	}
	if err := g.mvmPipeline.BindBuffer(3, g.outputBuffer); err != nil {
		return fmt.Errorf("failed to bind output buffer: %w", err)
	}
	if err := g.mvmPipeline.BindBuffer(4, g.variationBuffer); err != nil {
		return fmt.Errorf("failed to bind variation buffer: %w", err)
	}

	return nil
}

// IsAvailable returns true if GPU compute is available and initialized.
func (g *GPUAccelerator) IsAvailable() bool {
	return g != nil && g.ctx != nil && g.ctx.IsAvailable()
}

// MVM performs GPU-accelerated Matrix-Vector Multiplication with non-idealities.
//
// Computes: I_out = G * V_in
// where:
//   - G is the conductance matrix (rows × cols)
//   - V_in is the input voltage vector (rows)
//   - I_out is the output current vector (cols)
//
// Parameters:
//   - conductances: Conductance matrix in row-major order (G[row*cols + col])
//   - inputs: Input voltage vector (normalized to [0, 1])
//   - params: Crossbar simulation parameters
//
// Returns:
//   - Output current vector (cols elements)
//   - Error if any
func (g *GPUAccelerator) MVM(conductances []float32, inputs []float32, params CrossbarParams) ([]float32, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("GPU not available")
	}

	// Validate dimensions
	rows := int(params.Rows)
	cols := int(params.Cols)

	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("invalid dimensions: rows=%d, cols=%d", rows, cols)
	}
	if rows > g.maxRows || cols > g.maxCols {
		return nil, fmt.Errorf("dimensions exceed max: rows=%d (max %d), cols=%d (max %d)",
			rows, g.maxRows, cols, g.maxCols)
	}
	if len(conductances) != rows*cols {
		return nil, fmt.Errorf("conductance matrix size mismatch: expected %d, got %d", rows*cols, len(conductances))
	}
	if len(inputs) != rows {
		return nil, fmt.Errorf("input vector size mismatch: expected %d, got %d", rows, len(inputs))
	}

	// Upload conductances to GPU
	if err := g.conductanceBuffer.UploadFloat32(conductances); err != nil {
		return nil, fmt.Errorf("failed to upload conductances: %w", err)
	}

	// Upload inputs to GPU
	if err := g.inputBuffer.UploadFloat32(inputs); err != nil {
		return nil, fmt.Errorf("failed to upload inputs: %w", err)
	}

	// Initialize variation buffer (if not already set)
	// For now, use unit variation (no device-to-device variation)
	// In a real implementation, this would be pre-programmed during weight loading
	variation := make([]float32, rows*cols)
	for i := range variation {
		variation[i] = 1.0 // No variation
	}
	if err := g.variationBuffer.UploadFloat32(variation); err != nil {
		return nil, fmt.Errorf("failed to upload variation data: %w", err)
	}

	// Set uniform parameters (std140 layout)
	paramBytes, err := encodeParams(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode parameters: %w", err)
	}
	if err := g.mvmPipeline.SetUniformRaw(0, paramBytes); err != nil {
		return nil, fmt.Errorf("failed to set uniform parameters: %w", err)
	}

	// Dispatch compute shader
	// Each thread computes one output column (one element of I_out)
	// workgroup size = 256 (local_size_x in shader)
	workgroupSize := uint32(256)
	numWorkgroups := (uint32(cols) + workgroupSize - 1) / workgroupSize

	if err := g.mvmPipeline.Dispatch(numWorkgroups, 1, 1); err != nil {
		return nil, fmt.Errorf("compute dispatch failed: %w", err)
	}

	// Download results from GPU
	outputs := make([]float32, cols)
	if err := g.outputBuffer.DownloadFloat32(outputs); err != nil {
		return nil, fmt.Errorf("failed to download outputs: %w", err)
	}

	return outputs, nil
}

// SetDeviceVariation allows setting custom device-to-device variation factors.
// variation must have rows*cols elements.
func (g *GPUAccelerator) SetDeviceVariation(variation []float32) error {
	if !g.IsAvailable() {
		return fmt.Errorf("GPU not available")
	}

	expectedSize := g.maxRows * g.maxCols
	if len(variation) != expectedSize {
		return fmt.Errorf("variation size mismatch: expected %d, got %d", expectedSize, len(variation))
	}

	return g.variationBuffer.UploadFloat32(variation)
}

// Destroy releases all GPU resources.
// Must be called when the accelerator is no longer needed.
func (g *GPUAccelerator) Destroy() {
	if g.conductanceBuffer != nil {
		g.conductanceBuffer.Destroy()
		g.conductanceBuffer = nil
	}
	if g.inputBuffer != nil {
		g.inputBuffer.Destroy()
		g.inputBuffer = nil
	}
	if g.outputBuffer != nil {
		g.outputBuffer.Destroy()
		g.outputBuffer = nil
	}
	if g.variationBuffer != nil {
		g.variationBuffer.Destroy()
		g.variationBuffer = nil
	}
	if g.mvmPipeline != nil {
		g.mvmPipeline.Destroy()
		g.mvmPipeline = nil
	}
	if g.ctx != nil {
		g.ctx.Destroy()
		g.ctx = nil
	}
}

// findRepoRoot walks up the directory tree to find the repository root (go.mod location).
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up directory tree looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find repository root (no go.mod found)")
}

// encodeParams encodes CrossbarParams into a byte slice using std140 layout.
// std140 requires 4-byte alignment for all scalar types.
func encodeParams(params CrossbarParams) ([]byte, error) {
	// Allocate 32 bytes for the struct (8 fields × 4 bytes each)
	buf := make([]byte, 32)

	// std140 layout (little-endian):
	// offset 0: rows (int32)
	// offset 4: cols (int32)
	// offset 8: noiseLevel (float32)
	// offset 12: adcBits (int32)
	// offset 16: dacBits (int32)
	// offset 20: time (float32)
	// offset 24: wireResistance (float32)
	// offset 28: driftCoeff (float32)

	binary.LittleEndian.PutUint32(buf[0:4], uint32(params.Rows))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(params.Cols))
	binary.LittleEndian.PutUint32(buf[8:12], math.Float32bits(params.NoiseLevel))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(params.ADCBits))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(params.DACBits))
	binary.LittleEndian.PutUint32(buf[20:24], math.Float32bits(params.Time))
	binary.LittleEndian.PutUint32(buf[24:28], math.Float32bits(params.WireResistance))
	binary.LittleEndian.PutUint32(buf[28:32], math.Float32bits(params.DriftCoeff))

	return buf, nil
}
