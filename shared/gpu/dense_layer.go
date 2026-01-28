// Package gpu provides GPU-accelerated dense layer operations using Vulkan compute.
package gpu

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"fecim-lattice-tools/shared/compute"
)

// DenseLayerGPU provides GPU-accelerated dense layer forward pass.
// This is a simplified, single-layer inference suitable for real-time GUI updates.
type DenseLayerGPU struct {
	ctx      *compute.VulkanContext
	pipeline *compute.ComputePipeline

	// Pre-allocated buffers
	weightBuffer *compute.GPUBuffer
	biasBuffer   *compute.GPUBuffer
	inputBuffer  *compute.GPUBuffer
	outputBuffer *compute.GPUBuffer

	maxRows int // Output features
	maxCols int // Input features
}

// NewDenseLayerGPU creates a GPU-accelerated dense layer using an existing Vulkan context.
//
// Parameters:
//   - ctx: Existing VulkanContext to use
//   - maxRows: Maximum output features (rows in weight matrix)
//   - maxCols: Maximum input features (columns in weight matrix)
func NewDenseLayerGPU(ctx *compute.VulkanContext, maxRows, maxCols int) (*DenseLayerGPU, error) {
	if ctx == nil {
		return nil, fmt.Errorf("VulkanContext is nil")
	}
	if !ctx.IsAvailable() {
		return nil, fmt.Errorf("VulkanContext is not available")
	}
	if maxRows <= 0 || maxCols <= 0 {
		return nil, fmt.Errorf("invalid dimensions: rows=%d, cols=%d", maxRows, maxCols)
	}

	// Find repository root for absolute shader path
	repoRoot, err := findRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	layer := &DenseLayerGPU{
		ctx:     ctx,
		maxRows: maxRows,
		maxCols: maxCols,
	}

	// Create compute pipeline
	if err := layer.createPipeline(repoRoot); err != nil {
		layer.Destroy()
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Pre-allocate buffers for maximum dimensions
	if err := layer.allocateBuffers(); err != nil {
		layer.Destroy()
		return nil, fmt.Errorf("failed to allocate buffers: %w", err)
	}

	return layer, nil
}

// createPipeline creates the compute pipeline for dense_mvm.comp shader.
func (d *DenseLayerGPU) createPipeline(repoRoot string) error {
	// The dense_mvm.comp shader expects:
	// - binding 0: uniform LayerParams (std140)
	// - binding 1: storage WeightBuffer (W matrix, row-major)
	// - binding 2: storage BiasBuffer (b vector)
	// - binding 3: storage InputBuffer (x vector)
	// - binding 4: storage OutputBuffer (y vector)
	config := compute.PipelineConfig{
		ShaderPath: filepath.Join(repoRoot, "shared/gpu/shaders/dense_mvm.comp.spv"),
		Bindings: []compute.BindingInfo{
			{Binding: 0, Type: compute.BindingTypeUniform, Size: uint64(unsafe.Sizeof(LayerParams{}))}, // 16 bytes
			{Binding: 1, Type: compute.BindingTypeStorage, Size: 0},                                    // WeightBuffer (dynamic)
			{Binding: 2, Type: compute.BindingTypeStorage, Size: 0},                                    // BiasBuffer (dynamic)
			{Binding: 3, Type: compute.BindingTypeStorage, Size: 0},                                    // InputBuffer (dynamic)
			{Binding: 4, Type: compute.BindingTypeStorage, Size: 0},                                    // OutputBuffer (dynamic)
		},
	}

	pipeline, err := compute.NewComputePipeline(d.ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create compute pipeline: %w", err)
	}

	d.pipeline = pipeline
	return nil
}

// allocateBuffers pre-allocates GPU buffers for maximum layer size.
func (d *DenseLayerGPU) allocateBuffers() error {
	// Calculate buffer sizes
	weightSize := uint64(d.maxRows * d.maxCols * 4) // float32 = 4 bytes
	biasSize := uint64(d.maxRows * 4)
	inputSize := uint64(d.maxCols * 4)
	outputSize := uint64(d.maxRows * 4)

	// Create host-visible buffers for easy CPU upload/download
	var err error

	d.weightBuffer, err = d.ctx.CreateBuffer(weightSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create weight buffer: %w", err)
	}

	d.biasBuffer, err = d.ctx.CreateBuffer(biasSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create bias buffer: %w", err)
	}

	d.inputBuffer, err = d.ctx.CreateBuffer(inputSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create input buffer: %w", err)
	}

	d.outputBuffer, err = d.ctx.CreateBuffer(outputSize, compute.BufferUsageStorage, false)
	if err != nil {
		return fmt.Errorf("failed to create output buffer: %w", err)
	}

	// Bind buffers to pipeline descriptor set
	if err := d.pipeline.BindBuffer(1, d.weightBuffer); err != nil {
		return fmt.Errorf("failed to bind weight buffer: %w", err)
	}
	if err := d.pipeline.BindBuffer(2, d.biasBuffer); err != nil {
		return fmt.Errorf("failed to bind bias buffer: %w", err)
	}
	if err := d.pipeline.BindBuffer(3, d.inputBuffer); err != nil {
		return fmt.Errorf("failed to bind input buffer: %w", err)
	}
	if err := d.pipeline.BindBuffer(4, d.outputBuffer); err != nil {
		return fmt.Errorf("failed to bind output buffer: %w", err)
	}

	return nil
}

// IsAvailable returns true if GPU compute is available and initialized.
func (d *DenseLayerGPU) IsAvailable() bool {
	return d != nil && d.ctx != nil && d.ctx.IsAvailable()
}

// Forward performs a GPU-accelerated forward pass through the dense layer.
//
// Computes: y = W * x + b (with optional activation)
// where:
//   - W is the weight matrix (rows × cols)
//   - x is the input vector (cols)
//   - b is the bias vector (rows)
//   - y is the output vector (rows)
//
// Parameters:
//   - weights: Weight matrix in row-major order (W[row*cols + col])
//   - bias: Bias vector (rows elements)
//   - input: Input vector (cols elements)
//   - rows: Number of output features (rows in weight matrix)
//   - cols: Number of input features (columns in weight matrix)
//   - activation: Activation function to apply
//
// Returns:
//   - Output vector (rows elements)
//   - Error if any
func (d *DenseLayerGPU) Forward(weights, bias, input []float32, rows, cols int, activation ActivationType) ([]float32, error) {
	if !d.IsAvailable() {
		return nil, fmt.Errorf("GPU not available")
	}

	// Validate dimensions
	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("invalid dimensions: rows=%d, cols=%d", rows, cols)
	}
	if rows > d.maxRows || cols > d.maxCols {
		return nil, fmt.Errorf("dimensions exceed max: rows=%d (max %d), cols=%d (max %d)",
			rows, d.maxRows, cols, d.maxCols)
	}
	if len(weights) != rows*cols {
		return nil, fmt.Errorf("weight matrix size mismatch: expected %d, got %d", rows*cols, len(weights))
	}
	if len(bias) != rows {
		return nil, fmt.Errorf("bias vector size mismatch: expected %d, got %d", rows, len(bias))
	}
	if len(input) != cols {
		return nil, fmt.Errorf("input vector size mismatch: expected %d, got %d", cols, len(input))
	}

	// Upload weights to GPU
	if err := d.weightBuffer.UploadFloat32(weights); err != nil {
		return nil, fmt.Errorf("failed to upload weights: %w", err)
	}

	// Upload bias to GPU
	if err := d.biasBuffer.UploadFloat32(bias); err != nil {
		return nil, fmt.Errorf("failed to upload bias: %w", err)
	}

	// Upload input to GPU
	if err := d.inputBuffer.UploadFloat32(input); err != nil {
		return nil, fmt.Errorf("failed to upload input: %w", err)
	}

	// Set uniform parameters (std140 layout)
	params := LayerParams{
		Rows:       int32(rows),
		Cols:       int32(cols),
		Activation: int32(activation),
		Padding:    0,
	}

	paramBytes := EncodeLayerParams(params)
	if err := d.pipeline.SetUniformRaw(0, paramBytes); err != nil {
		return nil, fmt.Errorf("failed to set uniform parameters: %w", err)
	}

	// Dispatch compute shader
	// Each thread computes one output element
	// workgroup size = 256 (local_size_x in shader)
	workgroupSize := uint32(256)
	numWorkgroups := (uint32(rows) + workgroupSize - 1) / workgroupSize

	if err := d.pipeline.Dispatch(numWorkgroups, 1, 1); err != nil {
		return nil, fmt.Errorf("compute dispatch failed: %w", err)
	}

	// Download results from GPU
	output := make([]float32, rows)
	if err := d.outputBuffer.DownloadFloat32(output); err != nil {
		return nil, fmt.Errorf("failed to download output: %w", err)
	}

	return output, nil
}

// Destroy releases GPU resources owned by this layer.
// Does NOT destroy the VulkanContext (it's owned by the caller).
// Must be called when the layer is no longer needed.
func (d *DenseLayerGPU) Destroy() {
	if d.weightBuffer != nil {
		d.weightBuffer.Destroy()
		d.weightBuffer = nil
	}
	if d.biasBuffer != nil {
		d.biasBuffer.Destroy()
		d.biasBuffer = nil
	}
	if d.inputBuffer != nil {
		d.inputBuffer.Destroy()
		d.inputBuffer = nil
	}
	if d.outputBuffer != nil {
		d.outputBuffer.Destroy()
		d.outputBuffer = nil
	}
	if d.pipeline != nil {
		d.pipeline.Destroy()
		d.pipeline = nil
	}
	// Note: ctx is owned by caller, not destroyed here
	d.ctx = nil
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
