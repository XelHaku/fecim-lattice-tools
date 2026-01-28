// Package gpu provides GPU-accelerated neural network operations using Vulkan compute shaders.
package gpu

import (
	"fmt"

	"fecim-lattice-tools/shared/compute"
)

// GPUNetwork provides high-level GPU-accelerated neural network inference.
// It wraps a VulkanContext and DenseLayerGPU for multi-layer forward passes.
type GPUNetwork struct {
	ctx        *compute.VulkanContext
	denseLayer *DenseLayerGPU
	available  bool
}

// NewGPUNetwork creates a new GPU-accelerated neural network.
// Returns a non-nil GPUNetwork even if GPU is unavailable (for graceful fallback).
// Check IsAvailable() to determine if GPU acceleration is active.
//
// The network pre-allocates buffers for maximum expected dimensions:
// - Max input features: 784 (MNIST image size)
// - Max output features: 128 (typical hidden layer size)
func NewGPUNetwork() *GPUNetwork {
	net := &GPUNetwork{
		available: false,
	}

	// Attempt to create Vulkan context
	ctx, err := compute.NewVulkanContext()
	if err != nil {
		// Not an error - GPU just not available
		return net
	}

	if !ctx.IsAvailable() {
		// Vulkan context created but GPU not available
		ctx.Destroy()
		return net
	}

	net.ctx = ctx

	// Create dense layer with max dimensions (128 outputs, 784 inputs)
	// This is sufficient for MNIST-scale networks
	denseLayer, err := NewDenseLayerGPU(ctx, 128, 784)
	if err != nil {
		// Failed to create dense layer - cleanup and return unavailable
		ctx.Destroy()
		return net
	}

	net.denseLayer = denseLayer
	net.available = true

	return net
}

// IsAvailable returns true if GPU acceleration is available and initialized.
func (g *GPUNetwork) IsAvailable() bool {
	return g.available
}

// Forward performs a single forward pass through multiple layers.
// Processes layers sequentially, applying weights, bias, and activation for each layer.
//
// Parameters:
//   - input: Input vector (float32 slice)
//   - layers: Slice of LayerWeights defining the network architecture
//
// Returns:
//   - Output vector from the final layer
//   - Error if any layer fails
//
// Example:
//
//	layers := []gpu.LayerWeights{
//	    {Weights: w1, Bias: b1, Rows: 128, Cols: 784, Activation: ActivationReLU},
//	    {Weights: w2, Bias: b2, Rows: 10, Cols: 128, Activation: ActivationSoftmax},
//	}
//	output, err := net.Forward(input, layers)
func (g *GPUNetwork) Forward(input []float32, layers []LayerWeights) ([]float32, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("GPU not available")
	}

	if len(layers) == 0 {
		return nil, fmt.Errorf("no layers provided")
	}

	// Process layers sequentially
	currentInput := input
	for i, layer := range layers {
		// Validate dimensions
		if len(currentInput) != layer.Cols {
			return nil, fmt.Errorf("layer %d: input size %d does not match layer cols %d",
				i, len(currentInput), layer.Cols)
		}

		// Execute layer forward pass
		output, err := g.denseLayer.Forward(layer.Weights, layer.Bias, currentInput, layer.Rows, layer.Cols, layer.Activation)
		if err != nil {
			return nil, fmt.Errorf("layer %d forward failed: %w", i, err)
		}

		// Output becomes input for next layer
		currentInput = output
	}

	return currentInput, nil
}

// ForwardBatch performs forward pass on multiple inputs in batch.
// This is the primary method for achieving GPU speedup - processing multiple samples
// in parallel rather than one at a time.
//
// Parameters:
//   - inputs: Slice of input vectors
//   - layers: Network architecture (same for all inputs)
//
// Returns:
//   - Slice of output vectors (one per input)
//   - Error if any input fails
//
// Note: Current implementation processes inputs sequentially.
// Future optimization: Batch GPU kernels for true parallel processing.
func (g *GPUNetwork) ForwardBatch(inputs [][]float32, layers []LayerWeights) ([][]float32, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("GPU not available")
	}

	if len(inputs) == 0 {
		return nil, fmt.Errorf("no inputs provided")
	}

	outputs := make([][]float32, len(inputs))
	for i, input := range inputs {
		output, err := g.Forward(input, layers)
		if err != nil {
			return nil, fmt.Errorf("batch input %d failed: %w", i, err)
		}
		outputs[i] = output
	}

	return outputs, nil
}

// Destroy releases all GPU resources.
// Must be called when the network is no longer needed to prevent resource leaks.
func (g *GPUNetwork) Destroy() {
	if g.denseLayer != nil {
		g.denseLayer.Destroy()
		g.denseLayer = nil
	}
	if g.ctx != nil {
		g.ctx.Destroy()
		g.ctx = nil
	}
	g.available = false
}
