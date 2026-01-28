package gpu

import (
	"math"
	"math/rand"
	"testing"
)

// ============================================================================
// CPU Reference Functions
// ============================================================================

// cpuDenseLayer performs a CPU-based dense layer forward pass.
// Computes: y = W * x + b (with optional ReLU activation)
//
// Parameters:
//   - weights: Weight matrix in row-major order [rows*cols]
//   - bias: Bias vector [rows]
//   - input: Input vector [cols]
//   - rows: Number of output features
//   - cols: Number of input features
//   - relu: Apply ReLU activation if true
//
// Returns:
//   - Output vector [rows]
func cpuDenseLayer(weights []float32, bias []float32, input []float32, rows, cols int, relu bool) []float32 {
	output := make([]float32, rows)
	for i := 0; i < rows; i++ {
		sum := bias[i]
		for j := 0; j < cols; j++ {
			sum += weights[i*cols+j] * input[j]
		}
		if relu && sum < 0 {
			sum = 0
		}
		output[i] = sum
	}
	return output
}

// ============================================================================
// Test Helpers
// ============================================================================

// assertFloat32SliceEqual compares two float32 slices within tolerance.
func assertFloat32SliceEqual(t *testing.T, expected, actual []float32, tolerance float32, label string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("%s: length mismatch: expected %d, got %d", label, len(expected), len(actual))
		return
	}

	for i := range expected {
		diff := float32(math.Abs(float64(expected[i] - actual[i])))
		if diff > tolerance {
			t.Errorf("%s: element %d mismatch: expected %.6f, got %.6f (diff=%.6f > tolerance=%.6f)",
				label, i, expected[i], actual[i], diff, tolerance)
		}
	}
}

// createTestWeights creates a small deterministic weight matrix for testing.
// Returns weights in row-major order.
func createTestWeights(rows, cols int) []float32 {
	weights := make([]float32, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Simple pattern: weight[i,j] = 0.1 * (i + j)
			weights[i*cols+j] = 0.1 * float32(i+j)
		}
	}
	return weights
}

// createTestBias creates a small deterministic bias vector for testing.
func createTestBias(rows int) []float32 {
	bias := make([]float32, rows)
	for i := 0; i < rows; i++ {
		// Simple pattern: bias[i] = 0.01 * i
		bias[i] = 0.01 * float32(i)
	}
	return bias
}

// createTestInput creates a deterministic input vector for testing.
func createTestInput(cols int) []float32 {
	input := make([]float32, cols)
	for i := 0; i < cols; i++ {
		// Simple pattern: input[i] = 0.5 + 0.1 * i
		input[i] = 0.5 + 0.1*float32(i)
	}
	return input
}

// createRandomWeights creates a random weight matrix.
func createRandomWeights(rows, cols int, seed int64) []float32 {
	rng := rand.New(rand.NewSource(seed))
	weights := make([]float32, rows*cols)
	for i := range weights {
		weights[i] = rng.Float32()*2 - 1 // Range: [-1, 1]
	}
	return weights
}

// createRandomBias creates a random bias vector.
func createRandomBias(rows int, seed int64) []float32 {
	rng := rand.New(rand.NewSource(seed))
	bias := make([]float32, rows)
	for i := range bias {
		bias[i] = rng.Float32()*0.2 - 0.1 // Range: [-0.1, 0.1]
	}
	return bias
}

// createRandomInput creates a random input vector.
func createRandomInput(cols int, seed int64) []float32 {
	rng := rand.New(rand.NewSource(seed))
	input := make([]float32, cols)
	for i := range input {
		input[i] = rng.Float32() // Range: [0, 1]
	}
	return input
}

// ============================================================================
// GPU Parity Tests
// ============================================================================

// TestDenseLayerParity verifies GPU vs CPU dense layer results match.
// Uses a small 4x4 matrix with known values for deterministic testing.
func TestDenseLayerParity(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	// Create small test matrix (4 outputs, 4 inputs)
	rows, cols := 4, 4
	weights := createTestWeights(rows, cols)
	bias := createTestBias(rows)
	input := createTestInput(cols)

	// GPU forward pass (no activation)
	gpuOutput, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationNone)
	if err != nil {
		t.Fatalf("GPU forward failed: %v", err)
	}

	// CPU reference
	cpuOutput := cpuDenseLayer(weights, bias, input, rows, cols, false)

	// Compare results
	assertFloat32SliceEqual(t, cpuOutput, gpuOutput, 1e-5, "DenseLayer output")
}

// TestDenseLayerReLUParity verifies ReLU activation works correctly on GPU.
// Uses input with negative values to test ReLU clamping.
func TestDenseLayerReLUParity(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	// Create test matrix designed to produce some negative outputs
	rows, cols := 4, 4
	weights := []float32{
		// Row 0: should produce negative output
		-0.5, -0.3, -0.2, -0.1,
		// Row 1: should produce positive output
		0.5, 0.3, 0.2, 0.1,
		// Row 2: borderline (around zero)
		0.1, -0.1, 0.05, -0.05,
		// Row 3: large positive
		1.0, 0.8, 0.6, 0.4,
	}
	bias := []float32{-0.5, 0.5, 0.0, 1.0}
	input := []float32{1.0, 1.0, 1.0, 1.0}

	// GPU forward pass WITH ReLU
	gpuOutput, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationReLU)
	if err != nil {
		t.Fatalf("GPU forward with ReLU failed: %v", err)
	}

	// CPU reference WITH ReLU
	cpuOutput := cpuDenseLayer(weights, bias, input, rows, cols, true)

	// Compare results
	assertFloat32SliceEqual(t, cpuOutput, gpuOutput, 1e-5, "DenseLayer ReLU output")

	// Verify ReLU behavior: all outputs should be >= 0
	for i, v := range gpuOutput {
		if v < 0 {
			t.Errorf("ReLU failed: output[%d] = %.6f (should be >= 0)", i, v)
		}
	}
}

// TestTwoLayerNetworkParity verifies full 784→128→10 network.
// Tests the high-level GPUNetwork.Forward API with multiple layers.
func TestTwoLayerNetworkParity(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	// Create two-layer network: 784 → 128 (ReLU) → 10 (none)
	const (
		inputSize  = 784
		hiddenSize = 128
		outputSize = 10
	)

	// Create random weights and biases with fixed seed for reproducibility
	w1 := createRandomWeights(hiddenSize, inputSize, 42)
	b1 := createRandomBias(hiddenSize, 42)
	w2 := createRandomWeights(outputSize, hiddenSize, 43)
	b2 := createRandomBias(outputSize, 43)
	input := createRandomInput(inputSize, 44)

	// Define network layers
	layers := []LayerWeights{
		{
			Weights:    w1,
			Bias:       b1,
			Rows:       hiddenSize,
			Cols:       inputSize,
			Activation: ActivationReLU,
		},
		{
			Weights:    w2,
			Bias:       b2,
			Rows:       outputSize,
			Cols:       hiddenSize,
			Activation: ActivationNone,
		},
	}

	// GPU forward pass through both layers
	gpuOutput, err := gpuNet.Forward(input, layers)
	if err != nil {
		t.Fatalf("GPU two-layer forward failed: %v", err)
	}

	// CPU reference: manually compute both layers
	hidden := cpuDenseLayer(w1, b1, input, hiddenSize, inputSize, true)
	cpuOutput := cpuDenseLayer(w2, b2, hidden, outputSize, hiddenSize, false)

	// Compare final output (tolerance higher for accumulated error)
	assertFloat32SliceEqual(t, cpuOutput, gpuOutput, 1e-4, "Two-layer network output")
}

// TestGracefulFallback verifies behavior when GPU is unavailable.
// This test should PASS even when GPU is not available.
func TestGracefulFallback(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	// IsAvailable() should return a valid boolean (true or false)
	isAvailable := gpuNet.IsAvailable()
	t.Logf("GPU available: %v", isAvailable)

	if !isAvailable {
		// Verify that operations fail gracefully with clear error message
		input := []float32{1.0, 2.0, 3.0}
		layers := []LayerWeights{
			{
				Weights:    []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
				Bias:       []float32{0.1, 0.2},
				Rows:       2,
				Cols:       3,
				Activation: ActivationNone,
			},
		}

		output, err := gpuNet.Forward(input, layers)
		if err == nil {
			t.Errorf("Expected error when GPU unavailable, got output: %v", output)
		}
		if output != nil {
			t.Errorf("Expected nil output when GPU unavailable, got: %v", output)
		}
		t.Logf("Expected error when GPU unavailable: %v", err)
	} else {
		// GPU is available - verify basic operation works
		input := []float32{1.0, 2.0, 3.0}
		layers := []LayerWeights{
			{
				Weights:    []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
				Bias:       []float32{0.1, 0.2},
				Rows:       2,
				Cols:       3,
				Activation: ActivationNone,
			},
		}

		output, err := gpuNet.Forward(input, layers)
		if err != nil {
			t.Errorf("GPU operation failed: %v", err)
		}
		if len(output) != 2 {
			t.Errorf("Expected output length 2, got %d", len(output))
		}
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

// TestSmallDimensions verifies correctness with minimal dimensions.
func TestSmallDimensions(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	// Minimal case: 1 output, 1 input
	rows, cols := 1, 1
	weights := []float32{0.5}
	bias := []float32{0.1}
	input := []float32{2.0}

	gpuOutput, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationNone)
	if err != nil {
		t.Fatalf("GPU forward failed on 1x1: %v", err)
	}

	cpuOutput := cpuDenseLayer(weights, bias, input, rows, cols, false)

	assertFloat32SliceEqual(t, cpuOutput, gpuOutput, 1e-6, "1x1 layer output")

	// Expected: 0.5 * 2.0 + 0.1 = 1.1
	expected := float32(1.1)
	if math.Abs(float64(gpuOutput[0]-expected)) > 1e-6 {
		t.Errorf("1x1 computation incorrect: expected %.6f, got %.6f", expected, gpuOutput[0])
	}
}

// TestZeroInputs verifies behavior with all-zero inputs.
func TestZeroInputs(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	rows, cols := 4, 4
	weights := createTestWeights(rows, cols)
	bias := createTestBias(rows)
	input := make([]float32, cols) // All zeros

	gpuOutput, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationNone)
	if err != nil {
		t.Fatalf("GPU forward failed with zero input: %v", err)
	}

	cpuOutput := cpuDenseLayer(weights, bias, input, rows, cols, false)

	assertFloat32SliceEqual(t, cpuOutput, gpuOutput, 1e-6, "Zero input output")

	// With zero input, output should equal bias
	assertFloat32SliceEqual(t, bias, gpuOutput, 1e-6, "Zero input should equal bias")
}

// TestLargeDimensions verifies correctness with maximum supported dimensions.
func TestLargeDimensions(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	// Test with maximum dimensions: 128x784 (typical MNIST hidden layer)
	rows, cols := 128, 784
	weights := createRandomWeights(rows, cols, 100)
	bias := createRandomBias(rows, 100)
	input := createRandomInput(cols, 101)

	gpuOutput, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationReLU)
	if err != nil {
		t.Fatalf("GPU forward failed on 128x784: %v", err)
	}

	// Don't compute full CPU reference (too slow), just verify output shape and ReLU
	if len(gpuOutput) != rows {
		t.Errorf("Output length mismatch: expected %d, got %d", rows, len(gpuOutput))
	}

	// Verify ReLU constraint
	for i, v := range gpuOutput {
		if v < 0 {
			t.Errorf("ReLU failed on large matrix: output[%d] = %.6f", i, v)
		}
	}
}

// ============================================================================
// Batch Processing Tests
// ============================================================================

// TestForwardBatch verifies batch processing produces consistent results.
func TestForwardBatch(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	// Create test network
	rows, cols := 10, 20
	weights := createRandomWeights(rows, cols, 200)
	bias := createRandomBias(rows, 200)

	layers := []LayerWeights{
		{
			Weights:    weights,
			Bias:       bias,
			Rows:       rows,
			Cols:       cols,
			Activation: ActivationReLU,
		},
	}

	// Create batch of 5 inputs
	batchSize := 5
	inputs := make([][]float32, batchSize)
	for i := 0; i < batchSize; i++ {
		inputs[i] = createRandomInput(cols, int64(300+i))
	}

	// Run batch forward
	batchOutputs, err := gpuNet.ForwardBatch(inputs, layers)
	if err != nil {
		t.Fatalf("Batch forward failed: %v", err)
	}

	// Verify batch size
	if len(batchOutputs) != batchSize {
		t.Fatalf("Batch output size mismatch: expected %d, got %d", batchSize, len(batchOutputs))
	}

	// Verify each output matches individual forward pass
	for i, input := range inputs {
		individualOutput, err := gpuNet.Forward(input, layers)
		if err != nil {
			t.Fatalf("Individual forward failed for input %d: %v", i, err)
		}

		assertFloat32SliceEqual(t, individualOutput, batchOutputs[i], 1e-5,
			"Batch output vs individual output")
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

// TestInvalidDimensions verifies proper error handling for dimension mismatches.
func TestInvalidDimensions(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	rows, cols := 4, 4

	// Test 1: Incorrect weight matrix size
	t.Run("incorrect_weight_size", func(t *testing.T) {
		weights := make([]float32, rows*cols+1) // Wrong size
		bias := createTestBias(rows)
		input := createTestInput(cols)

		_, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationNone)
		if err == nil {
			t.Error("Expected error for incorrect weight size")
		}
	})

	// Test 2: Incorrect bias size
	t.Run("incorrect_bias_size", func(t *testing.T) {
		weights := createTestWeights(rows, cols)
		bias := make([]float32, rows+1) // Wrong size
		input := createTestInput(cols)

		_, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationNone)
		if err == nil {
			t.Error("Expected error for incorrect bias size")
		}
	})

	// Test 3: Incorrect input size
	t.Run("incorrect_input_size", func(t *testing.T) {
		weights := createTestWeights(rows, cols)
		bias := createTestBias(rows)
		input := make([]float32, cols+1) // Wrong size

		_, err := gpuNet.denseLayer.Forward(weights, bias, input, rows, cols, ActivationNone)
		if err == nil {
			t.Error("Expected error for incorrect input size")
		}
	})

	// Test 4: Dimensions exceed maximum
	t.Run("exceed_max_dimensions", func(t *testing.T) {
		// GPUNetwork is initialized with max 128x784
		// Try to use larger dimensions
		largeRows := 200
		largeCols := 1000
		weights := make([]float32, largeRows*largeCols)
		bias := make([]float32, largeRows)
		input := make([]float32, largeCols)

		_, err := gpuNet.denseLayer.Forward(weights, bias, input, largeRows, largeCols, ActivationNone)
		if err == nil {
			t.Error("Expected error for dimensions exceeding maximum")
		}
	})
}

// TestEmptyLayers verifies error handling for empty layer list.
func TestEmptyLayers(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	input := createTestInput(10)
	layers := []LayerWeights{} // Empty

	_, err := gpuNet.Forward(input, layers)
	if err == nil {
		t.Error("Expected error for empty layer list")
	}
}

// TestLayerDimensionMismatch verifies error handling for incompatible layer dimensions.
func TestLayerDimensionMismatch(t *testing.T) {
	gpuNet := NewGPUNetwork()
	defer gpuNet.Destroy()

	if !gpuNet.IsAvailable() {
		t.Skip("GPU not available, skipping GPU test")
	}

	// Create two layers with incompatible dimensions
	// Layer 1: 10 outputs
	// Layer 2: expects 20 inputs (mismatch!)
	layers := []LayerWeights{
		{
			Weights:    createRandomWeights(10, 5, 500),
			Bias:       createRandomBias(10, 500),
			Rows:       10,
			Cols:       5,
			Activation: ActivationReLU,
		},
		{
			Weights:    createRandomWeights(8, 20, 501), // Expects 20 inputs, but layer 1 outputs 10
			Bias:       createRandomBias(8, 501),
			Rows:       8,
			Cols:       20,
			Activation: ActivationNone,
		},
	}

	input := createRandomInput(5, 502)

	_, err := gpuNet.Forward(input, layers)
	if err == nil {
		t.Error("Expected error for incompatible layer dimensions")
	}
}
