package core

import (
	"fecim-lattice-tools/shared/gpu"
	"fmt"
)

// GPUAccelerator holds the GPU network for inference acceleration.
var globalGPUAccelerator *gpu.GPUNetwork
var gpuInitAttempted bool

// InitGPU initializes the global GPU accelerator.
// Safe to call multiple times; only initializes once.
func InitGPU() {
	if gpuInitAttempted {
		return
	}
	gpuInitAttempted = true
	globalGPUAccelerator = gpu.NewGPUNetwork()
}

// IsGPUAvailable returns true if GPU acceleration is available.
func IsGPUAvailable() bool {
	InitGPU()
	return globalGPUAccelerator != nil && globalGPUAccelerator.IsAvailable()
}

// DestroyGPU releases GPU resources. Call when done with all networks.
func DestroyGPU() {
	if globalGPUAccelerator != nil {
		globalGPUAccelerator.Destroy()
		globalGPUAccelerator = nil
	}
	gpuInitAttempted = false
}

// SetUseGPU enables or disables GPU acceleration for this network.
func (net *DualModeNetwork) SetUseGPU(use bool) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.useGPU = use && IsGPUAvailable()
}

// UseGPU returns whether GPU acceleration is enabled for this network.
func (net *DualModeNetwork) UseGPU() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.useGPU
}

// forwardFPGPU performs GPU-accelerated forward pass.
// Converts float64 to float32, runs on GPU, converts back.
func (net *DualModeNetwork) forwardFPGPU(input []float64, weights [][]float64, bias []float64) ([]float64, error) {
	if globalGPUAccelerator == nil || !globalGPUAccelerator.IsAvailable() {
		return nil, fmt.Errorf("GPU not available")
	}

	// Convert to float32
	input32 := gpu.Float64ToFloat32(input)
	weights32 := gpu.Weights2DToFlat(weights)
	bias32 := gpu.Float64ToFloat32(bias)

	// Create layer weights
	layers := []gpu.LayerWeights{{
		Weights:    weights32,
		Bias:       bias32,
		Rows:       len(bias),
		Cols:       len(input),
		Activation: gpu.ActivationNone, // Activation handled separately in inference
	}}

	// Run GPU forward
	output32, err := globalGPUAccelerator.Forward(input32, layers)
	if err != nil {
		return nil, err
	}

	// Convert back to float64
	return gpu.Float32ToFloat64(output32), nil
}
