package crossbar

import (
	"math"
	"testing"
)

// TestGPUMVMParityWithCPU verifies GPU MVM produces identical results to CPU MVM.
func TestGPUMVMParityWithCPU(t *testing.T) {
	// Use non-square matrix to catch row/col confusion
	cfg := &Config{
		Rows:       4,
		Cols:       8, // Non-square: 4 outputs, 8 inputs
		NoiseLevel: 0, // Zero noise for deterministic comparison
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer arr.Destroy()

	// Check if GPU is available
	arr.initGPU()
	gpuAvailable := arr.gpuAccelerator != nil && arr.gpuAccelerator.IsAvailable()

	// Program known weights (pattern that reveals transpose errors)
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			// Weight = row*0.1 + col*0.01 (unique per cell)
			w := float64(i)*0.1 + float64(j)*0.01
			arr.ProgramWeight(i, j, w)
		}
	}

	// Test input (8 elements for 8 columns)
	input := []float64{1.0, 0.5, 0.8, 0.3, 0.6, 0.9, 0.2, 0.7}

	// Get CPU result
	outputCPU, err := arr.mvmCPU(input)
	if err != nil {
		t.Fatalf("CPU MVM failed: %v", err)
	}

	// Verify CPU output has correct length
	if len(outputCPU) != cfg.Rows {
		t.Fatalf("CPU output length wrong: got %d, expected %d", len(outputCPU), cfg.Rows)
	}

	if !gpuAvailable {
		t.Skip("GPU not available, skipping GPU parity test")
	}

	// Get GPU result
	outputGPU, err := arr.mvmGPU(input)
	if err != nil {
		t.Fatalf("GPU MVM failed: %v", err)
	}

	// Verify GPU output has correct length
	if len(outputGPU) != cfg.Rows {
		t.Fatalf("GPU output length wrong: got %d, expected %d", len(outputGPU), cfg.Rows)
	}

	// Compare results
	tolerance := 1e-4 // GPU uses float32, allow some precision loss
	for i := range outputGPU {
		diff := math.Abs(outputGPU[i] - outputCPU[i])
		if diff > tolerance {
			t.Errorf("Output[%d] mismatch: GPU=%f, CPU=%f, diff=%f",
				i, outputGPU[i], outputCPU[i], diff)
		}
	}
}

// TestGPUMVMIdentityMatrix verifies GPU MVM with identity-like weights.
func TestGPUMVMIdentityMatrix(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer arr.Destroy()

	arr.initGPU()
	if arr.gpuAccelerator == nil || !arr.gpuAccelerator.IsAvailable() {
		t.Skip("GPU not available")
	}

	// Program diagonal weights (identity-like)
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			if i == j {
				arr.ProgramWeight(i, j, 1.0)
			} else {
				arr.ProgramWeight(i, j, 0.0)
			}
		}
	}

	input := []float64{1.0, 0.5, 0.25, 0.75}

	outputGPU, err := arr.mvmGPU(input)
	if err != nil {
		t.Fatalf("GPU MVM failed: %v", err)
	}

	outputCPU, err := arr.mvmCPU(input)
	if err != nil {
		t.Fatalf("CPU MVM failed: %v", err)
	}

	// With identity matrix, output should approximate input (after quantization)
	tolerance := 1e-4
	for i := range outputGPU {
		diff := math.Abs(outputGPU[i] - outputCPU[i])
		if diff > tolerance {
			t.Errorf("Identity test Output[%d]: GPU=%f, CPU=%f, diff=%f",
				i, outputGPU[i], outputCPU[i], diff)
		}
	}
}

// TestGPUFallbackToCPU verifies graceful fallback when GPU unavailable.
func TestGPUFallbackToCPU(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer arr.Destroy()

	// Program some weights
	arr.ProgramWeight(0, 0, 1.0)
	arr.ProgramWeight(1, 1, 1.0)

	// MVM should work regardless of GPU availability
	input := []float64{1.0, 0.5, 0.3, 0.8}
	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	if len(output) != cfg.Rows {
		t.Errorf("Output length wrong: got %d, expected %d", len(output), cfg.Rows)
	}
}

// TestGPUNeuralNetworkLayer verifies GPU works for typical neural network use.
func TestGPUNeuralNetworkLayer(t *testing.T) {
	// Simulate MNIST first layer: 784 -> 128
	cfg := &Config{
		Rows:       128,
		Cols:       784, // 128 outputs (hidden), 784 inputs (pixels)
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer arr.Destroy()

	arr.initGPU()
	gpuAvailable := arr.gpuAccelerator != nil && arr.gpuAccelerator.IsAvailable()

	// Program deterministic weights
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			w := float64((i*cfg.Cols+j)%30) / 29.0 // Deterministic "random"
			arr.ProgramWeight(i, j, w)
		}
	}

	// Simulate 784-pixel input
	input := make([]float64, 784)
	for i := range input {
		input[i] = float64(i%10) / 10.0
	}

	// Get CPU result
	outputCPU, err := arr.mvmCPU(input)
	if err != nil {
		t.Fatalf("CPU MVM failed: %v", err)
	}

	if len(outputCPU) != 128 {
		t.Fatalf("CPU output wrong size: got %d, expected 128", len(outputCPU))
	}

	if !gpuAvailable {
		t.Skip("GPU not available")
	}

	// Get GPU result
	outputGPU, err := arr.mvmGPU(input)
	if err != nil {
		t.Fatalf("GPU MVM failed: %v", err)
	}

	if len(outputGPU) != 128 {
		t.Fatalf("GPU output wrong size: got %d, expected 128", len(outputGPU))
	}

	// Compare
	tolerance := 1e-4
	maxDiff := 0.0
	for i := range outputGPU {
		diff := math.Abs(outputGPU[i] - outputCPU[i])
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff > tolerance {
			t.Errorf("Layer output[%d]: GPU=%f, CPU=%f, diff=%f", i, outputGPU[i], outputCPU[i], diff)
		}
	}
	t.Logf("Max difference: %e", maxDiff)
}
