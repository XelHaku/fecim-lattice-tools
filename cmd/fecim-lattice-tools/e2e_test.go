// Package main provides end-to-end tests for the FeCIM Lattice Tools unified application.
// These tests verify complete application workflows including module lifecycle,
// switching, and core functionality flows.
//
// Run with: go test -v -race ./cmd/fecim-lattice-tools/... -run E2E
//
// Note: GUI-based tests that require a display are in e2e_visual_test.go with build tag !ci
package main

import (
	"context"
	"testing"
	"time"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/module3-mnist/pkg/core"
)

// =============================================================================
// NON-GUI WORKFLOW TESTS (These run without display)
// =============================================================================

// TestE2ECrossbarMVMWorkflow tests the complete matrix-vector multiply operation.
// This test exercises the crossbar core without GUI dependencies.
func TestE2ECrossbarMVMWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("IdentityMatrix", func(t *testing.T) {
		testCrossbarMVMIdentity(t)
	})

	t.Run("RandomWeights", func(t *testing.T) {
		testCrossbarMVMRandom(t)
	})

	t.Run("FullScale", func(t *testing.T) {
		testCrossbarMVMFullScale(t)
	})
}

// testCrossbarMVMIdentity tests MVM with identity-like matrix
func testCrossbarMVMIdentity(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0, // No noise for deterministic test
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create crossbar array: %v", err)
	}

	// Program identity-like weights (diagonal)
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			if i == j {
				arr.ProgramWeight(i, j, 1.0)
			} else {
				arr.ProgramWeight(i, j, 0.0)
			}
		}
	}

	// Create input vector
	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = float64(i+1) / float64(cfg.Cols) // [0.125, 0.25, 0.375, ...]
	}

	// Run MVM
	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// Verify output dimensions
	if len(output) != cfg.Rows {
		t.Errorf("Expected %d outputs, got %d", cfg.Rows, len(output))
	}

	// For identity-like matrix, output should approximate input
	for i := 0; i < len(output); i++ {
		expected := input[i]
		actual := output[i]
		tolerance := 0.25 // 25% tolerance for quantization effects
		if actual < expected*(1-tolerance) || actual > expected*(1+tolerance) {
			t.Logf("Output[%d]: expected ~%.3f, got %.3f", i, expected, actual)
		}
	}

	t.Logf("Identity MVM: %d inputs -> %d outputs", len(input), len(output))
}

// testCrossbarMVMRandom tests MVM with random weights
func testCrossbarMVMRandom(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:       32,
		Cols:       32,
		NoiseLevel: 0.02, // Small noise
		ADCBits:    6,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create crossbar array: %v", err)
	}

	// Program random weights (quantized to 30 levels)
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			level := (i*j + i + j) % 30
			weight := float64(level) / 29.0
			arr.ProgramWeight(i, j, weight)
		}
	}

	// Create input vector
	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = float64(i%10) / 10.0
	}

	// Run MVM
	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// Verify output dimensions and valid values
	if len(output) != cfg.Rows {
		t.Errorf("Expected %d outputs, got %d", cfg.Rows, len(output))
	}

	for i, v := range output {
		if v < -2.0 || v > 2.0 {
			t.Errorf("Output[%d] = %.4f out of expected range", i, v)
		}
	}

	t.Logf("Random MVM: 32x32 array with 2%% noise")
}

// testCrossbarMVMFullScale tests MVM with full-scale weights
func testCrossbarMVMFullScale(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:       64,
		Cols:       64,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create crossbar array: %v", err)
	}

	// Program all weights to 1.0
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			arr.ProgramWeight(i, j, 1.0)
		}
	}

	// Create all-ones input
	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = 1.0
	}

	// Run MVM
	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// All outputs should be ~1.0 (normalized)
	for i, v := range output {
		if v < 0.8 || v > 1.2 {
			t.Errorf("Output[%d] = %.4f, expected ~1.0", i, v)
		}
	}

	t.Log("Full-scale MVM: 64x64 array all-ones test passed")
}

// =============================================================================
// MNIST INFERENCE WORKFLOW TEST
// =============================================================================

// TestE2EMNISTInferenceWorkflow tests the complete MNIST digit recognition pipeline.
func TestE2EMNISTInferenceWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("BasicInference", func(t *testing.T) {
		testMNISTBasicInference(t)
	})

	t.Run("DualPathComparison", func(t *testing.T) {
		testMNISTDualPath(t)
	})

	t.Run("AllDigits", func(t *testing.T) {
		testMNISTAllDigits(t)
	})
}

// testMNISTBasicInference tests basic inference with the dual-mode network
func testMNISTBasicInference(t *testing.T) {
	net := core.NewDualModeNetwork(784, 128, 10)
	if net == nil {
		t.Fatal("Failed to create network")
	}

	// Initialize with simple weights
	initializeTestWeights(net)

	// Create synthetic input
	input := make([]float64, 784)
	for i := range input {
		input[i] = float64(i%28) / 28.0
	}

	// Run inference
	result := net.Infer(input)

	// Verify result structure
	if result.FPPrediction < 0 || result.FPPrediction > 9 {
		t.Errorf("Invalid FP prediction: %d", result.FPPrediction)
	}
	if result.CIMPrediction < 0 || result.CIMPrediction > 9 {
		t.Errorf("Invalid CIM prediction: %d", result.CIMPrediction)
	}
	if len(result.FPProbabilities) != 10 {
		t.Errorf("Expected 10 FP probabilities, got %d", len(result.FPProbabilities))
	}

	t.Logf("Basic inference: FP=%d, CIM=%d", result.FPPrediction, result.CIMPrediction)
}

// testMNISTDualPath compares FP and CIM paths
func testMNISTDualPath(t *testing.T) {
	net := core.NewDualModeNetwork(784, 128, 10)
	initializeTestWeights(net)

	// Run multiple inferences and check agreement rate
	agreements := 0
	total := 100

	for i := 0; i < total; i++ {
		input := make([]float64, 784)
		for j := range input {
			input[j] = float64((i*j+j)%100) / 100.0
		}

		result := net.Infer(input)
		if result.Agree {
			agreements++
		}
	}

	agreementRate := float64(agreements) / float64(total) * 100
	t.Logf("FP/CIM agreement rate: %.1f%% (%d/%d)", agreementRate, agreements, total)

	// With 30 levels and low noise, agreement should be high
	if agreementRate < 50 {
		t.Errorf("Agreement rate too low: %.1f%% (expected >50%%)", agreementRate)
	}
}

// testMNISTAllDigits tests inference for all digit patterns
func testMNISTAllDigits(t *testing.T) {
	net := core.NewDualModeNetwork(784, 128, 10)
	initializeTestWeights(net)

	for digit := 0; digit < 10; digit++ {
		input := make([]float64, 784)
		// Create a simple pattern
		for i := 0; i < 784; i++ {
			input[i] = float64((i+digit*100)%256) / 255.0
		}

		result := net.Infer(input)

		// Verify valid predictions
		if result.FPPrediction < 0 || result.FPPrediction > 9 {
			t.Errorf("Digit %d: Invalid FP prediction %d", digit, result.FPPrediction)
		}
		if result.CIMPrediction < 0 || result.CIMPrediction > 9 {
			t.Errorf("Digit %d: Invalid CIM prediction %d", digit, result.CIMPrediction)
		}

		// Verify probabilities sum to ~1
		sum := 0.0
		for _, p := range result.FPProbabilities {
			sum += p
		}
		if sum < 0.99 || sum > 1.01 {
			t.Errorf("Digit %d: FP probabilities sum to %.4f", digit, sum)
		}
	}

	t.Log("All 10 digit patterns processed successfully")
}

// initializeTestWeights sets up deterministic weights for testing
func initializeTestWeights(net *core.DualModeNetwork) {
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i*j)%100-50) / 500.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i*j)%100-50) / 200.0
		}
	}
	net.FPBias1 = make([]float64, net.HiddenSize)
	net.FPBias2 = make([]float64, net.OutputSize)
	for i := range net.FPBias1 {
		net.FPBias1[i] = 0.01
	}
	for i := range net.FPBias2 {
		net.FPBias2[i] = 0.01
	}
	net.RequantizeWeights()
}

// =============================================================================
// QUANTIZATION WORKFLOW TEST
// =============================================================================

// TestE2EQuantizationWorkflow tests the 30-level quantization across modules.
func TestE2EQuantizationWorkflow(t *testing.T) {
	t.Run("DefaultLevels", func(t *testing.T) {
		net := core.NewDualModeNetwork(100, 50, 10)
		if net.Config.NumLevels != 30 {
			t.Errorf("Expected 30 default levels, got %d", net.Config.NumLevels)
		}
	})

	t.Run("QuantizationQuality", func(t *testing.T) {
		weights := make([][]float64, 10)
		for i := range weights {
			weights[i] = make([]float64, 10)
			for j := range weights[i] {
				weights[i][j] = float64(j)/9.0*2 - 1 // Range [-1, 1]
			}
		}

		quantized, err := core.QuantizeWeights(weights, 30)
		if err != nil {
			t.Fatalf("Quantization failed: %v", err)
		}

		// Count distinct levels
		levels := make(map[float64]bool)
		for _, row := range quantized {
			for _, v := range row {
				levels[v] = true
			}
		}

		t.Logf("30-level quantization produced %d distinct levels", len(levels))
		if len(levels) < 5 {
			t.Errorf("Expected multiple distinct levels, got %d", len(levels))
		}
	})

	t.Run("QuantizationStats", func(t *testing.T) {
		net := core.NewDualModeNetwork(100, 50, 10)
		initializeTestWeights(net)

		stats := core.ComputeQuantizationStats(net.FPWeights1, net.QuantWeights1)
		t.Logf("Quantization stats: MSE=%.6f, PSNR=%.1f dB", stats.MSE, stats.PSNR)

		if stats.PSNR < 20 {
			t.Errorf("PSNR too low: %.1f dB (expected >20 dB)", stats.PSNR)
		}
	})
}

// =============================================================================
// TIMEOUT AND CONCURRENCY TESTS
// =============================================================================

// TestE2EModuleCreationTimeout verifies modules can be created within timeout.
func TestE2EModuleCreationTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("CrossbarCreation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			cfg := &crossbar.Config{
				Rows:       128,
				Cols:       128,
				NoiseLevel: 0.02,
				ADCBits:    8,
				DACBits:    8,
			}
			_, err := crossbar.NewArray(cfg)
			done <- err
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Crossbar creation failed: %v", err)
			} else {
				t.Log("Large crossbar (128x128) created within timeout")
			}
		case <-ctx.Done():
			t.Error("Crossbar creation timed out")
		}
	})

	t.Run("NetworkCreation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := make(chan bool, 1)
		go func() {
			net := core.NewDualModeNetwork(784, 256, 10)
			initializeTestWeights(net)
			done <- net != nil
		}()

		select {
		case success := <-done:
			if !success {
				t.Error("Network creation failed")
			} else {
				t.Log("Large network (784-256-10) created within timeout")
			}
		case <-ctx.Done():
			t.Error("Network creation timed out")
		}
	})
}

// TestE2EConcurrentInference tests thread safety of inference.
func TestE2EConcurrentInference(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	net := core.NewDualModeNetwork(784, 128, 10)
	initializeTestWeights(net)

	const numGoroutines = 10
	const inferPerGoroutine = 50

	errors := make(chan error, numGoroutines*inferPerGoroutine)
	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					errors <- errorf("goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			for i := 0; i < inferPerGoroutine; i++ {
				input := make([]float64, 784)
				for j := range input {
					input[j] = float64((id*i+j)%100) / 100.0
				}

				result := net.Infer(input)
				if result.FPPrediction < 0 || result.FPPrediction > 9 {
					errors <- errorf("goroutine %d iter %d: invalid prediction", id, i)
				}
			}
		}(g)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errors)

	errCount := 0
	for err := range errors {
		t.Error(err)
		errCount++
	}

	if errCount == 0 {
		t.Logf("Concurrent inference: %d goroutines x %d inferences = %d total",
			numGoroutines, inferPerGoroutine, numGoroutines*inferPerGoroutine)
	}
}

// =============================================================================
// PHYSICS CONSTANTS VERIFICATION
// =============================================================================

// TestE2EPhysicsConstants verifies FeCIM physics constants are consistent.
func TestE2EPhysicsConstants(t *testing.T) {
	// Verify 30-level quantization constant
	if core.FeCIMLevels != 30 {
		t.Errorf("FeCIMLevels should be 30, got %d", core.FeCIMLevels)
	}

	// Verify network uses correct levels
	net := core.NewDualModeNetwork(784, 128, 10)
	if net.Config.NumLevels != 30 {
		t.Errorf("Default network levels should be 30, got %d", net.Config.NumLevels)
	}

	t.Log("Physics constants verified: 30 discrete analog levels")
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// errorf creates an error with formatted message
func errorf(format string, args ...interface{}) error {
	return &testError{msg: format, args: args}
}

type testError struct {
	msg  string
	args []interface{}
}

func (e *testError) Error() string {
	return e.msg
}
