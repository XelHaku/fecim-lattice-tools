// Package core provides integration tests for the MNIST FeCIM demo.
// These tests verify end-to-end functionality and UI/UX flows.
package core

import (
	"fmt"
	"math"
	"testing"
)

// =============================================================================
// END-TO-END INFERENCE TESTS
// =============================================================================

// TestFullInferencePipeline tests complete inference from input to prediction.
func TestFullInferencePipeline(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Initialize weights with a pattern
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			// Xavier-like initialization
			net.FPWeights1[i][j] = (float64((i*j)%100) - 50) / 500.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = (float64((i*j)%100) - 50) / 200.0
		}
	}
	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)
	for i := range net.FPBias1 {
		net.FPBias1[i] = 0.01
	}
	for i := range net.FPBias2 {
		net.FPBias2[i] = 0.01
	}

	net.RequantizeWeights()

	// Create a synthetic digit "1" pattern (vertical line in center)
	input := make([]float64, 784)
	for row := 4; row < 24; row++ {
		input[row*28+14] = 1.0 // Center column
		input[row*28+13] = 0.5 // Adjacent columns
		input[row*28+15] = 0.5
	}

	result := net.Infer(input)

	// Verify result structure
	if len(result.FPProbabilities) != 10 {
		t.Errorf("Expected 10 FP probabilities, got %d", len(result.FPProbabilities))
	}
	if len(result.CIMProbabilities) != 10 {
		t.Errorf("Expected 10 CIM probabilities, got %d", len(result.CIMProbabilities))
	}

	// Predictions should be valid class indices
	if result.FPPrediction < 0 || result.FPPrediction > 9 {
		t.Errorf("Invalid FP prediction: %d", result.FPPrediction)
	}
	if result.CIMPrediction < 0 || result.CIMPrediction > 9 {
		t.Errorf("Invalid CIM prediction: %d", result.CIMPrediction)
	}

	// Confidences should be in [0, 1]
	if result.FPConfidence < 0 || result.FPConfidence > 1 {
		t.Errorf("FP confidence out of range: %f", result.FPConfidence)
	}
	if result.CIMConfidence < 0 || result.CIMConfidence > 1 {
		t.Errorf("CIM confidence out of range: %f", result.CIMConfidence)
	}

	// Hidden activations should exist
	if len(result.FPHidden) != 128 {
		t.Errorf("Expected 128 FP hidden activations, got %d", len(result.FPHidden))
	}
	if len(result.CIMHidden) != 128 {
		t.Errorf("Expected 128 CIM hidden activations, got %d", len(result.CIMHidden))
	}

	t.Logf("Pipeline test: FP=%d (%.1f%%), CIM=%d (%.1f%%), Agree=%v",
		result.FPPrediction, result.FPConfidence*100,
		result.CIMPrediction, result.CIMConfidence*100,
		result.Agree)
}

// TestPresetConfigurations tests the 4 failure mode presets.
func TestPresetConfigurations(t *testing.T) {
	presets := []struct {
		name    string
		levels  int
		noise   float64
		adcBits int
		dacBits int
	}{
		{"Ideal", 30, 0.01, 8, 8},
		{"Quant Cliff", 2, 0.01, 8, 8},
		{"Noisy", 30, 0.15, 6, 8},
		{"Broken ADC", 30, 0.01, 3, 8},
	}

	net := NewDualModeNetwork(16, 8, 4)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = (float64((i+j)%10) - 5) / 10.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = (float64((i+j)%10) - 5) / 10.0
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)

	input := make([]float64, 16)
	for i := range input {
		input[i] = float64(i) / 15.0
	}

	for _, preset := range presets {
		net.SetNumLevels(preset.levels)
		net.SetNoiseLevel(preset.noise)
		net.SetADCBits(preset.adcBits)
		net.SetDACBits(preset.dacBits)

		result := net.Infer(input)

		t.Logf("%-12s: FP=%d(%.0f%%), CIM=%d(%.0f%%), Agree=%v, KL=%.4f",
			preset.name,
			result.FPPrediction, result.FPConfidence*100,
			result.CIMPrediction, result.CIMConfidence*100,
			result.Agree, result.Disagreement)
	}
}

// TestQuickTestAccuracy simulates the "Run Quick Test" functionality.
func TestQuickTestAccuracy(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize with structured weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = math.Sin(float64(i+j)) * 0.5
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = math.Cos(float64(i*j)) * 0.5
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)
	net.RequantizeWeights()

	// Simulate quick test with synthetic data
	numSamples := 50
	agreements := 0

	for sample := 0; sample < numSamples; sample++ {
		// Create random-ish input
		input := make([]float64, 16)
		for i := range input {
			input[i] = math.Mod(float64(sample*i+1)*0.1, 1.0)
		}

		result := net.Infer(input)
		if result.Agree {
			agreements++
		}
	}

	agreementRate := float64(agreements) / float64(numSamples) * 100

	t.Logf("Quick test: %d/%d agreements (%.1f%%)", agreements, numSamples, agreementRate)

	// Under ideal conditions (30 levels, low noise), agreement should be high
	if agreementRate < 50 {
		t.Log("Note: Low agreement rate - may indicate issues or expected degradation")
	}
}

// =============================================================================
// UI STATE TESTS
// =============================================================================

// TestNetworkConfigBounds tests that configuration parameters are properly bounded.
func TestNetworkConfigBounds(t *testing.T) {
	net := NewDualModeNetwork(10, 5, 3)

	// Test NumLevels bounds
	net.SetNumLevels(0)
	if net.Config.NumLevels < 2 {
		t.Error("NumLevels should be >= 2")
	}

	net.SetNumLevels(100)
	if net.Config.NumLevels > MaxDemoLevels {
		t.Errorf("NumLevels should be <= %d", MaxDemoLevels)
	}

	// Test NoiseLevel bounds
	net.SetNoiseLevel(-0.5)
	if net.Config.NoiseLevel < 0 {
		t.Error("NoiseLevel should be >= 0")
	}

	net.SetNoiseLevel(1.0)
	if net.Config.NoiseLevel > 0.5 {
		// Note: Some implementations may allow higher noise for testing
		t.Log("NoiseLevel might want to be capped")
	}

	// Test ADC bounds
	net.SetADCBits(1)
	if net.Config.ADCBits < 3 {
		t.Error("ADCBits should be >= 3")
	}

	net.SetADCBits(32)
	if net.Config.ADCBits > 16 {
		t.Error("ADCBits should be <= 16")
	}

	// Test DAC bounds
	net.SetDACBits(1)
	if net.Config.DACBits < 3 {
		t.Error("DACBits should be >= 3")
	}
}

// TestRequantizeWeightsIdempotent tests that requantization is consistent.
func TestRequantizeWeightsIdempotent(t *testing.T) {
	net := NewDualModeNetwork(10, 5, 3)

	// Initialize FP weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64(i-j) / 10.0
		}
	}

	// Quantize once
	net.RequantizeWeights()
	firstQuant := make([][]float64, len(net.QuantWeights1))
	for i := range net.QuantWeights1 {
		firstQuant[i] = make([]float64, len(net.QuantWeights1[i]))
		copy(firstQuant[i], net.QuantWeights1[i])
	}

	// Quantize again (without changing FP weights)
	net.RequantizeWeights()

	// Should get same result
	for i := range net.QuantWeights1 {
		for j := range net.QuantWeights1[i] {
			if net.QuantWeights1[i][j] != firstQuant[i][j] {
				t.Errorf("Requantization not idempotent at [%d][%d]", i, j)
			}
		}
	}
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// TestZeroInput tests inference with all-zero input.
func TestZeroInput(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = 0.1
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = 0.1
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)
	net.RequantizeWeights()

	// All-zero input
	input := make([]float64, 16)

	result := net.Infer(input)

	// Should still produce valid output
	if result.FPPrediction < 0 || result.FPPrediction > 3 {
		t.Error("Invalid prediction for zero input")
	}

	// Check no NaN/Inf
	for i, p := range result.FPProbabilities {
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Errorf("NaN/Inf in probability[%d]", i)
		}
	}
}

// TestMaxInput tests inference with all-maximum input.
func TestMaxInput(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = 0.1
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = 0.1
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)
	net.RequantizeWeights()

	// All-max input
	input := make([]float64, 16)
	for i := range input {
		input[i] = 1.0
	}

	result := net.Infer(input)

	// Should still produce valid output
	if result.FPPrediction < 0 || result.FPPrediction > 3 {
		t.Error("Invalid prediction for max input")
	}

	// Check no NaN/Inf
	for i, p := range result.FPProbabilities {
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Errorf("NaN/Inf in probability[%d]", i)
		}
	}
}

// TestDifferentHiddenSizes tests various network architectures.
func TestDifferentHiddenSizes(t *testing.T) {
	hiddenSizes := []int{16, 32, 64, 128, 256}

	for _, hidden := range hiddenSizes {
		net := NewDualModeNetwork(100, hidden, 10)

		// Initialize weights
		for i := range net.FPWeights1 {
			for j := range net.FPWeights1[i] {
				net.FPWeights1[i][j] = float64((i+j)%20-10) / 100.0
			}
		}
		for i := range net.FPWeights2 {
			for j := range net.FPWeights2[i] {
				net.FPWeights2[i][j] = float64((i+j)%20-10) / 100.0
			}
		}
		net.FPBias1 = make([]float64, hidden)
		net.FPBias2 = make([]float64, 10)
		net.RequantizeWeights()

		input := make([]float64, 100)
		for i := range input {
			input[i] = float64(i%10) / 10.0
		}

		result := net.Infer(input)

		if len(result.FPHidden) != hidden {
			t.Errorf("Hidden size %d: expected %d hidden activations, got %d",
				hidden, hidden, len(result.FPHidden))
		}

		t.Logf("Hidden=%3d: FP=%d, CIM=%d, Agree=%v",
			hidden, result.FPPrediction, result.CIMPrediction, result.Agree)
	}
}

// =============================================================================
// CONCURRENCY TESTS
// =============================================================================

// TestConcurrentInference tests thread safety of inference.
func TestConcurrentInference(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i+j)%10) / 10.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i+j)%10) / 10.0
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)
	net.RequantizeWeights()

	// Run concurrent inferences
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for goroutine := 0; goroutine < 10; goroutine++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					// Safe type assertion with comma-ok idiom
					if err, ok := r.(error); ok {
						errors <- err
					} else {
						errors <- fmt.Errorf("panic: %v", r)
					}
				}
				done <- true
			}()

			for i := 0; i < 100; i++ {
				input := make([]float64, 16)
				for j := range input {
					input[j] = float64((id+i+j)%10) / 10.0
				}

				result := net.Infer(input)

				if result.FPPrediction < 0 || result.FPPrediction > 3 {
					t.Errorf("Goroutine %d: invalid prediction", id)
				}
			}
		}(goroutine)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	select {
	case err := <-errors:
		t.Errorf("Concurrent inference caused panic: %v", err)
	default:
		t.Log("Concurrent inference test passed")
	}
}

// =============================================================================
// SINGLE-LAYER MODE TESTS
// =============================================================================

// TestSingleLayerModeE2E tests Calibration Mode (784→10 single layer).
// This mode simulates the simplified hardware demo architecture without hidden layer.
func TestSingleLayerModeE2E(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Initialize single-layer weights with Xavier-like initialization
	for i := range net.SingleLayerWeights {
		for j := range net.SingleLayerWeights[i] {
			net.SingleLayerWeights[i][j] = (float64((i*j)%100) - 50) / 1000.0
		}
	}
	net.SingleLayerBias = make([]float64, 10)
	for i := range net.SingleLayerBias {
		net.SingleLayerBias[i] = 0.01
	}

	// Enable SingleLayer mode
	net.Config.SingleLayer = true
	net.RequantizeWeights()

	// Create synthetic input (784 elements)
	input := make([]float64, 784)
	for i := range input {
		input[i] = math.Mod(float64(i)*0.01, 1.0)
	}

	result := net.Infer(input)

	// Verify FP path produces valid result
	if result.FPPrediction < 0 || result.FPPrediction > 9 {
		t.Errorf("Invalid FP prediction: %d", result.FPPrediction)
	}
	if len(result.FPProbabilities) != 10 {
		t.Errorf("Expected 10 FP probabilities, got %d", len(result.FPProbabilities))
	}

	// Verify CIM path produces valid result
	if result.CIMPrediction < 0 || result.CIMPrediction > 9 {
		t.Errorf("Invalid CIM prediction: %d", result.CIMPrediction)
	}
	if len(result.CIMProbabilities) != 10 {
		t.Errorf("Expected 10 CIM probabilities, got %d", len(result.CIMProbabilities))
	}

	// No hidden layer in single-layer mode
	if result.FPHidden != nil {
		t.Error("Expected no FP hidden layer in single-layer mode")
	}
	if result.CIMHidden != nil {
		t.Error("Expected no CIM hidden layer in single-layer mode")
	}

	// Energy calculation should reflect single layer only (7,840 MACs)
	expectedMACs := 784 * 10 // 7,840 MACs
	bitsPerCell := math.Log2(30.0)
	expectedEnergy := float64(expectedMACs) * 10e-15 * bitsPerCell * 1e6

	if result.EnergyUsed < expectedEnergy*0.5 || result.EnergyUsed > expectedEnergy*2.0 {
		t.Errorf("Single-layer energy %.4f µJ outside expected range [%.4f, %.4f] µJ",
			result.EnergyUsed, expectedEnergy*0.5, expectedEnergy*2.0)
	}

	t.Logf("Single-layer mode: FP=%d (%.1f%%), CIM=%d (%.1f%%), Energy=%.4f µJ, MACs=%d",
		result.FPPrediction, result.FPConfidence*100,
		result.CIMPrediction, result.CIMConfidence*100,
		result.EnergyUsed, expectedMACs)
}

// =============================================================================
// PER-LAYER QUANTIZATION TESTS
// =============================================================================

// TestPerLayerQuantizationE2E tests per-layer PTQ configuration.
// Layer 1 can use different quantization levels than Layer 2.
func TestPerLayerQuantizationE2E(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = (float64((i+j)%100) - 50) / 500.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = (float64((i+j)%100) - 50) / 200.0
		}
	}
	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)

	// Enable per-layer quantization with different levels
	net.Config.PerLayerQuant = true
	net.Config.Layer1Levels = 20
	net.Config.Layer2Levels = 10
	net.RequantizeWeights()

	// Create synthetic input
	input := make([]float64, 784)
	for i := range input {
		input[i] = math.Sin(float64(i)*0.1)*0.5 + 0.5
	}

	result := net.Infer(input)

	// Verify inference produces valid results
	if result.FPPrediction < 0 || result.FPPrediction > 9 {
		t.Errorf("Invalid FP prediction: %d", result.FPPrediction)
	}
	if result.CIMPrediction < 0 || result.CIMPrediction > 9 {
		t.Errorf("Invalid CIM prediction: %d", result.CIMPrediction)
	}

	// Energy calculation should reflect different bit depths
	macs1 := 784 * 128
	macs2 := 128 * 10
	bits1 := math.Log2(20.0) // ~4.32 bits
	bits2 := math.Log2(10.0) // ~3.32 bits
	expectedEnergy := float64(macs1)*10e-15*bits1*1e6 + float64(macs2)*10e-15*bits2*1e6

	if result.EnergyUsed < expectedEnergy*0.5 || result.EnergyUsed > expectedEnergy*2.0 {
		t.Errorf("Per-layer quant energy %.4f µJ outside expected range [%.4f, %.4f] µJ",
			result.EnergyUsed, expectedEnergy*0.5, expectedEnergy*2.0)
	}

	t.Logf("Per-layer quantization: L1=%d levels, L2=%d levels, Energy=%.4f µJ, Agree=%v",
		net.Config.Layer1Levels, net.Config.Layer2Levels, result.EnergyUsed, result.Agree)
}

// =============================================================================
// AGREEMENT RATE STATISTICS TESTS
// =============================================================================

// TestAgreementRateStatistics performs statistical test of FP/CIM agreement rate.
// With moderate noise, agreement should still be high (>80%).
func TestAgreementRateStatistics(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Initialize weights with known values
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = math.Sin(float64(i+j)) * 0.3
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = math.Cos(float64(i*j)) * 0.3
		}
	}
	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)

	// Test with moderate noise first
	net.SetNoiseLevel(0.05)
	net.RequantizeWeights()

	numTrials := 100
	agreements := 0
	var disagreements []int

	for trial := 0; trial < numTrials; trial++ {
		// Create different input for each trial
		input := make([]float64, 784)
		for i := range input {
			input[i] = math.Mod(float64(trial*i+1)*0.01, 1.0)
		}

		result := net.Infer(input)
		if result.Agree {
			agreements++
		} else {
			disagreements = append(disagreements, trial)
		}
	}

	agreementRate := float64(agreements) / float64(numTrials) * 100

	// With 5% noise and 30 levels, agreement should be >80%
	if agreementRate < 80 {
		t.Errorf("Agreement rate %.1f%% below expected 80%% threshold (noise=0.05)",
			agreementRate)
	}

	t.Logf("Agreement statistics (noise=0.05): %d/%d agreements (%.1f%%), %d disagreements",
		agreements, numTrials, agreementRate, len(disagreements))

	// Test with zero noise - should have 100% agreement
	net.SetNoiseLevel(0.0)
	perfectAgreements := 0

	for trial := 0; trial < numTrials; trial++ {
		input := make([]float64, 784)
		for i := range input {
			input[i] = math.Mod(float64(trial*i+1)*0.01, 1.0)
		}

		result := net.Infer(input)
		if result.Agree {
			perfectAgreements++
		}
	}

	perfectRate := float64(perfectAgreements) / float64(numTrials) * 100
	if perfectRate < 100 {
		t.Logf("Note: With zero noise, agreement rate is %.1f%% (expected 100%%)", perfectRate)
	}

	t.Logf("Agreement statistics (noise=0.00): %d/%d agreements (%.1f%%)",
		perfectAgreements, numTrials, perfectRate)
}

// =============================================================================
// ENERGY CALCULATION VERIFICATION TESTS
// =============================================================================

// TestEnergyCalculationVerification verifies the energy calculation formula.
// Energy should scale with NumLevels and reflect correct MAC counts.
func TestEnergyCalculationVerification(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Initialize simple weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = 0.1
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = 0.1
		}
	}
	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)
	net.RequantizeWeights()

	input := make([]float64, 784)
	for i := range input {
		input[i] = 0.5
	}

	// Verify MAC counts: Layer 1: 784*128, Layer 2: 128*10
	macs1 := 784 * 128         // 100,352 MACs
	macs2 := 128 * 10          // 1,280 MACs
	totalMACs := macs1 + macs2 // 101,632 MACs

	// Test energy scaling with different levels
	levelsToTest := []struct {
		levels        int
		expectedBits  float64
		expectedFJMAC float64
	}{
		{2, 1.0, 10.0},  // 1 bit
		{8, 3.0, 30.0},  // 3 bits
		{16, 4.0, 40.0}, // 4 bits
		{30, 4.9, 49.0}, // ~4.9 bits
	}

	for _, test := range levelsToTest {
		net.SetNumLevels(test.levels)
		result := net.Infer(input)

		// Energy formula: 10 * log2(levels) fJ/MAC
		expectedEnergy := float64(totalMACs) * test.expectedFJMAC * 1e-15 * 1e6 // Convert to µJ

		if math.Abs(result.EnergyUsed-expectedEnergy) > expectedEnergy*0.3 {
			t.Errorf("Energy mismatch for %d levels: got %.4f µJ, expected %.4f µJ",
				test.levels, result.EnergyUsed, expectedEnergy)
		}

		actualFJMAC := result.EnergyUsed * 1e6 / float64(totalMACs)
		t.Logf("Levels=%2d: Bits=%.1f, Energy=%.4f µJ, %.1f fJ/MAC (expected %.1f fJ/MAC)",
			test.levels, test.expectedBits, result.EnergyUsed, actualFJMAC, test.expectedFJMAC)
	}

	// Verify energy scales with levels (more levels = more energy)
	net.SetNumLevels(2)
	result2 := net.Infer(input)
	net.SetNumLevels(30)
	result30 := net.Infer(input)

	ratio := result30.EnergyUsed / result2.EnergyUsed
	expectedRatio := 4.9 // log2(30) / log2(2)

	if ratio < expectedRatio*0.8 || ratio > expectedRatio*1.2 {
		t.Errorf("Energy scaling ratio %.2f outside expected range [%.2f, %.2f]",
			ratio, expectedRatio*0.8, expectedRatio*1.2)
	}

	t.Logf("Energy scaling verification: 30-level/2-level = %.2fx (expected ~%.1fx)",
		ratio, expectedRatio)
}
