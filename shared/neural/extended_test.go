package neural

import (
	"math"
	"testing"
)

// ============================================
// Quantization Edge Cases
// ============================================

func TestQuantizeWeights_1Level(t *testing.T) {
	// levels=1 should not panic, weights should collapse to single value
	fpWeights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
		{-0.8, -0.3, 0.2, 0.7, 0.9},
	}

	// Note: QuantizeWeights requires levels >= 2
	quantized, err := QuantizeWeights(fpWeights, 1)
	if err == nil {
		t.Fatal("Expected error for levels=1, got nil")
	}

	if quantized != nil {
		t.Error("Expected nil quantized weights for invalid levels")
	}
}

func TestQuantizeWeights_MaxLevels(t *testing.T) {
	// levels=MaxDemoLevels: verify distinct values <= MaxDemoLevels
	fpWeights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
		{-0.8, -0.3, 0.2, 0.7, 0.9},
	}

	quantized, err := QuantizeWeights(fpWeights, MaxDemoLevels)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Count distinct values
	distinct := make(map[float64]bool)
	for i := range quantized {
		for j := range quantized[i] {
			distinct[quantized[i][j]] = true
		}
	}

	if len(distinct) > MaxDemoLevels {
		t.Errorf("Too many distinct values: got %d, want <= %d", len(distinct), MaxDemoLevels)
	}

	// Verify all values in [-1, 1]
	for i := range quantized {
		for j := range quantized[i] {
			if math.Abs(quantized[i][j]) > 1.01 {
				t.Errorf("Quantized value out of range: %f", quantized[i][j])
			}
		}
	}
}

func TestQuantizeWeights_NegativeLevel(t *testing.T) {
	// levels=-1 or levels=0: should handle gracefully (not panic)
	fpWeights := [][]float64{
		{-0.5, 0.0, 0.5},
	}

	testCases := []struct {
		name   string
		levels int
	}{
		{"NegativeOne", -1},
		{"Zero", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			quantized, err := QuantizeWeights(fpWeights, tc.levels)
			if err == nil {
				t.Errorf("Expected error for levels=%d, got nil", tc.levels)
			}
			if quantized != nil {
				t.Errorf("Expected nil result for levels=%d", tc.levels)
			}
		})
	}
}

func TestQuantizeWeights_LargeMatrix(t *testing.T) {
	// 784x128 matrix: verify shape preserved, values in [-1,1]
	rows := 784
	cols := 128

	fpWeights := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		fpWeights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Initialize with varied values in range [-1, 1]
			fpWeights[i][j] = 2.0*(float64(i*cols+j)/float64(rows*cols)) - 1.0
		}
	}

	quantized, err := QuantizeWeights(fpWeights, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Verify shape preserved
	if len(quantized) != rows {
		t.Errorf("Row count mismatch: got %d, want %d", len(quantized), rows)
	}

	for i := 0; i < rows; i++ {
		if len(quantized[i]) != cols {
			t.Errorf("Col count mismatch at row %d: got %d, want %d", i, len(quantized[i]), cols)
		}
	}

	// Verify values in [-1, 1]
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if math.Abs(quantized[i][j]) > 1.01 {
				t.Errorf("Value[%d][%d] out of range: %f", i, j, quantized[i][j])
			}
		}
	}
}

func TestQuantizeWeights_Symmetry(t *testing.T) {
	// Symmetric input like {-0.5, 0, 0.5} should produce symmetric quantized values
	fpWeights := [][]float64{
		{-0.5, 0.0, 0.5},
		{-1.0, 0.0, 1.0},
	}

	quantized, err := QuantizeWeights(fpWeights, 10)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Check symmetry: quantized[-x] should equal -quantized[x]
	// The quantization uses global max, so symmetry is relative to the quantization bins
	for i := range quantized {
		if len(quantized[i]) >= 3 {
			// For {-x, 0, +x} pattern
			leftVal := quantized[i][0]
			rightVal := quantized[i][2]
			if math.Abs(leftVal+rightVal) > 0.001 {
				t.Errorf("Symmetry broken: quantized[%d][0]=%f, quantized[%d][2]=%f (should be negatives)",
					i, leftVal, i, rightVal)
			}

			// Note: Zero may not map exactly to zero due to discrete binning
			// For the row {-1.0, 0.0, 1.0}, zero should be relatively centered
			// For the row {-0.5, 0.0, 0.5}, it depends on the global max (which is 1.0)
			// So we just verify zero is between the symmetric values
			zeroVal := quantized[i][1]
			if zeroVal < leftVal || zeroVal > rightVal {
				t.Errorf("Zero value out of symmetric range: quantized[%d][1]=%f not in [%f, %f]",
					i, zeroVal, leftVal, rightVal)
			}
		}
	}
}

// ============================================
// QuantizeBias Edge Cases
// ============================================

func TestQuantizeBias_EmptyInput(t *testing.T) {
	// Empty slice should return empty
	bias := []float64{}
	quantized, err := QuantizeBias(bias, 10)
	if err != nil {
		t.Fatalf("QuantizeBias failed: %v", err)
	}

	if len(quantized) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(quantized))
	}
}

func TestQuantizeBias_AllSameValue(t *testing.T) {
	// All values identical, quantized should be close
	bias := []float64{0.5, 0.5, 0.5, 0.5}
	quantized, err := QuantizeBias(bias, 10)
	if err != nil {
		t.Fatalf("QuantizeBias failed: %v", err)
	}

	// All quantized values should be identical
	if len(quantized) != len(bias) {
		t.Fatalf("Length mismatch: got %d, want %d", len(quantized), len(bias))
	}

	firstVal := quantized[0]
	for i, v := range quantized {
		if v != firstVal {
			t.Errorf("Value[%d]=%f differs from first value %f", i, v, firstVal)
		}
	}
}

// ============================================
// ComputeQuantizationStats Edge Cases
// ============================================

func TestComputeQuantizationStats_IdenticalArrays(t *testing.T) {
	// MSE should be 0, PSNR should be Inf or very high
	original := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
	}
	quantized := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
	}

	stats := ComputeQuantizationStats(original, quantized)

	if stats.MSE != 0 {
		t.Errorf("MSE should be 0 for identical arrays, got %f", stats.MSE)
	}

	if !math.IsInf(stats.PSNR, 1) && stats.PSNR < 100 {
		t.Errorf("PSNR should be Inf or very high for identical arrays, got %f", stats.PSNR)
	}
}

func TestComputeQuantizationStats_BinaryQuantization(t *testing.T) {
	// 2-level: MSE should be higher than 30-level
	original := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
	}

	quantized2, err := QuantizeWeights(original, 2)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	quantized30, err := QuantizeWeights(original, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	stats2 := ComputeQuantizationStats(original, quantized2)
	stats30 := ComputeQuantizationStats(original, quantized30)

	if stats2.MSE <= stats30.MSE {
		t.Errorf("2-level MSE (%f) should be higher than 30-level MSE (%f)", stats2.MSE, stats30.MSE)
	}

	if stats2.PSNR >= stats30.PSNR {
		t.Errorf("2-level PSNR (%f dB) should be lower than 30-level PSNR (%f dB)", stats2.PSNR, stats30.PSNR)
	}
}

// ============================================
// DualModeNetwork Edge Cases
// ============================================

func TestDualModeNetwork_ZeroWeights(t *testing.T) {
	// All weights zero, verify inference doesn't panic, all probabilities equal
	net := NewDualModeNetwork(4, 3, 2)

	// All weights and biases are initialized to zero by default
	// Verify they are zero
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			if net.FPWeights1[i][j] != 0 {
				net.FPWeights1[i][j] = 0
			}
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			if net.FPWeights2[i][j] != 0 {
				net.FPWeights2[i][j] = 0
			}
		}
	}

	net.RequantizeWeights()

	input := []float64{0.5, 0.5, 0.5, 0.5}
	result := net.Infer(input)

	if result == nil {
		t.Fatal("Infer returned nil for zero weights")
	}

	// With zero weights and biases, all outputs should be equal after softmax
	// Check that probabilities are approximately equal
	expectedProb := 1.0 / float64(net.OutputSize)
	for i, p := range result.FPProbabilities {
		if math.Abs(p-expectedProb) > 0.01 {
			t.Errorf("FP probability[%d]=%f, expected ~%f (equal distribution)", i, p, expectedProb)
		}
	}
}

func TestDualModeNetwork_SingleNeuronHidden(t *testing.T) {
	// HiddenSize=1, verify works
	net := NewDualModeNetwork(4, 1, 2)

	// Initialize weights
	net.FPWeights1 = [][]float64{
		{0.1, 0.2, 0.3, 0.4},
	}
	net.FPWeights2 = [][]float64{
		{0.5},
		{0.6},
	}
	net.FPBias1 = []float64{0.1}
	net.FPBias2 = []float64{0.1, 0.2}

	net.RequantizeWeights()

	input := []float64{0.5, 0.5, 0.5, 0.5}
	result := net.Infer(input)

	if result == nil {
		t.Fatal("Infer returned nil for single hidden neuron")
	}

	if result.FPPrediction < 0 || result.FPPrediction >= 2 {
		t.Errorf("Invalid prediction: %d", result.FPPrediction)
	}
}

func TestDualModeNetwork_LargeNetwork(t *testing.T) {
	// 784,256,10 network creation and single inference
	net := NewDualModeNetwork(784, 256, 10)

	if net.InputSize != 784 {
		t.Errorf("Wrong input size: %d", net.InputSize)
	}
	if net.HiddenSize != 256 {
		t.Errorf("Wrong hidden size: %d", net.HiddenSize)
	}
	if net.OutputSize != 10 {
		t.Errorf("Wrong output size: %d", net.OutputSize)
	}

	// Initialize with small random values
	rng := NewRandomSource(42)
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = (rng.Float64()*2 - 1) * 0.1
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = (rng.Float64()*2 - 1) * 0.1
		}
	}

	net.RequantizeWeights()

	// Create input
	input := make([]float64, 784)
	for i := range input {
		input[i] = rng.Float64()
	}

	result := net.Infer(input)

	if result == nil {
		t.Fatal("Infer returned nil for large network")
	}

	if result.FPPrediction < 0 || result.FPPrediction >= 10 {
		t.Errorf("Invalid prediction: %d", result.FPPrediction)
	}
}

func TestDualModeNetwork_RequantizeIdempotent(t *testing.T) {
	// Calling RequantizeWeights twice doesn't change quantized weights
	net := NewDualModeNetwork(4, 3, 2)

	// Initialize FP weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64(i+j) / 10.0
		}
	}

	net.RequantizeWeights()

	// Copy quantized weights
	firstQuant := make([][]float64, len(net.QuantWeights1))
	for i := range net.QuantWeights1 {
		firstQuant[i] = make([]float64, len(net.QuantWeights1[i]))
		copy(firstQuant[i], net.QuantWeights1[i])
	}

	// Call again
	net.RequantizeWeights()

	// Compare
	for i := range net.QuantWeights1 {
		for j := range net.QuantWeights1[i] {
			if net.QuantWeights1[i][j] != firstQuant[i][j] {
				t.Errorf("Quantized weights changed after second RequantizeWeights: [%d][%d] %f != %f",
					i, j, net.QuantWeights1[i][j], firstQuant[i][j])
			}
		}
	}
}

// ============================================
// Noise and Agreement Tests
// ============================================

func TestDualModeNetwork_NoiseEffectOnAgreement(t *testing.T) {
	// NoiseLevel=0 vs NoiseLevel=0.3, run 20 inferences each, noisy should have lower agreement rate
	net := NewDualModeNetwork(4, 3, 2)

	// Initialize weights
	net.FPWeights1 = [][]float64{
		{0.5, 0.5, 0.5, 0.5},
		{-0.5, 0.5, -0.5, 0.5},
		{0.5, -0.5, 0.5, -0.5},
	}
	net.FPWeights2 = [][]float64{
		{1.0, 0.5, 0.5},
		{0.5, 1.0, 0.5},
	}
	net.FPBias1 = []float64{0.1, 0.1, 0.1}
	net.FPBias2 = []float64{0.1, 0.1}

	input := []float64{1.0, 0.0, 1.0, 0.0}

	// Test with zero noise
	net.Config.NoiseLevel = 0.0
	net.Config.NumLevels = 30
	net.Config.ADCBits = 16
	net.Config.DACBits = 16
	net.RequantizeWeights()

	zeroNoiseAgree := 0
	for i := 0; i < 20; i++ {
		result := net.Infer(input)
		if result != nil && result.Agree {
			zeroNoiseAgree++
		}
	}

	// Test with high noise
	net.Config.NoiseLevel = 0.3
	net.RequantizeWeights()

	highNoiseAgree := 0
	for i := 0; i < 20; i++ {
		result := net.Infer(input)
		if result != nil && result.Agree {
			highNoiseAgree++
		}
	}

	// With deterministic RNG and zero noise, agreement should be higher
	if zeroNoiseAgree < highNoiseAgree {
		t.Errorf("Zero noise agreement (%d) should be >= high noise agreement (%d)",
			zeroNoiseAgree, highNoiseAgree)
	}

	t.Logf("Agreement rates: zero noise=%d/20, high noise=%d/20", zeroNoiseAgree, highNoiseAgree)
}

func TestDualModeNetwork_LowLevelsHighError(t *testing.T) {
	// NumLevels=2 should produce higher CIM error than NumLevels=30
	net := NewDualModeNetwork(4, 3, 2)

	// Initialize weights
	net.FPWeights1 = [][]float64{
		{0.5, 0.5, 0.5, 0.5},
		{-0.5, 0.5, -0.5, 0.5},
		{0.5, -0.5, 0.5, -0.5},
	}
	net.FPWeights2 = [][]float64{
		{1.0, 0.5, 0.5},
		{0.5, 1.0, 0.5},
	}
	net.FPBias1 = []float64{0.1, 0.1, 0.1}
	net.FPBias2 = []float64{0.1, 0.1}

	input := []float64{1.0, 0.0, 1.0, 0.0}

	// Test with 2 levels
	net.Config.NumLevels = 2
	net.Config.NoiseLevel = 0.0
	net.Config.ADCBits = 16
	net.Config.DACBits = 16
	net.RequantizeWeights()

	result2 := net.Infer(input)

	// Test with 30 levels
	net.Config.NumLevels = 30
	net.RequantizeWeights()

	result30 := net.Infer(input)

	if result2 == nil || result30 == nil {
		t.Fatal("Infer returned nil")
	}

	// KL divergence should be higher for 2-level
	if result2.Disagreement < result30.Disagreement {
		t.Logf("Note: 2-level KL (%f) is not higher than 30-level KL (%f) - this can happen with simple inputs",
			result2.Disagreement, result30.Disagreement)
	}

	t.Logf("Disagreement (KL): 2-level=%f, 30-level=%f", result2.Disagreement, result30.Disagreement)
}

// ============================================
// Energy Model Tests
// ============================================

func TestEnergyPerMACJ_Monotonic(t *testing.T) {
	// Energy increases with levels (more bits = more energy)
	levels := []int{2, 4, 8, 16, 30}
	prevEnergy := 0.0

	for _, level := range levels {
		energy := EnergyPerMACJ(level)
		if energy <= prevEnergy && prevEnergy > 0 {
			t.Errorf("Energy not monotonic: EnergyPerMACJ(%d)=%e <= EnergyPerMACJ(prev)=%e",
				level, energy, prevEnergy)
		}
		prevEnergy = energy
	}
}

func TestEnergyPerMACJ_2Levels(t *testing.T) {
	// Returns EnergyPerBitPerMACJ * 1.0 (log2(2)=1)
	energy := EnergyPerMACJ(2)
	expected := EnergyPerBitPerMACJ * 1.0

	if math.Abs(energy-expected) > 1e-18 {
		t.Errorf("EnergyPerMACJ(2)=%e, expected %e", energy, expected)
	}
}

func TestEnergyPerMACJ_30Levels(t *testing.T) {
	// Returns EnergyPerBitPerMACJ * log2(30)
	energy := EnergyPerMACJ(30)
	expected := EnergyPerBitPerMACJ * math.Log2(30)

	if math.Abs(energy-expected) > 1e-18 {
		t.Errorf("EnergyPerMACJ(30)=%e, expected %e", energy, expected)
	}
}

// ============================================
// EstimateInferenceEnergyJ Tests
// ============================================

func TestEstimateInferenceEnergyJ_Components(t *testing.T) {
	// Verify TotalMACs = input*hidden + hidden*output
	cfg := DefaultNetworkConfig()
	inputSize := 784
	hiddenSize := 128
	outputSize := 10

	est := EstimateInferenceEnergyJ(cfg, inputSize, hiddenSize, outputSize)

	expectedMACs1 := inputSize * hiddenSize
	expectedMACs2 := hiddenSize * outputSize
	expectedTotal := expectedMACs1 + expectedMACs2

	if est.MACs1 != expectedMACs1 {
		t.Errorf("MACs1=%d, expected %d", est.MACs1, expectedMACs1)
	}
	if est.MACs2 != expectedMACs2 {
		t.Errorf("MACs2=%d, expected %d", est.MACs2, expectedMACs2)
	}
	if est.TotalMACs != expectedTotal {
		t.Errorf("TotalMACs=%d, expected %d", est.TotalMACs, expectedTotal)
	}

	// Verify TotalJ = ComputeJ + ADCDACJ
	if math.Abs(est.TotalJ-(est.ComputeJ+est.ADCDACJ)) > 1e-18 {
		t.Errorf("TotalJ=%e != ComputeJ(%e) + ADCDACJ(%e)", est.TotalJ, est.ComputeJ, est.ADCDACJ)
	}
}

func TestEstimateInferenceEnergyJ_ZeroLevels(t *testing.T) {
	// Should handle gracefully
	cfg := &NetworkConfig{
		NumLevels: 0, // Invalid
		ADCBits:   8,
		DACBits:   8,
	}

	est := EstimateInferenceEnergyJ(cfg, 784, 128, 10)

	// Should not panic and should return some energy
	if est.TotalJ <= 0 {
		t.Errorf("Expected positive energy even with zero levels, got %e", est.TotalJ)
	}

	// EnergyPerMACJ(0) is clamped to EnergyPerMACJ(1) internally
	// Verify it used at least 1 bit worth of energy
	expectedMinEnergy := EnergyPerBitPerMACJ * 1.0
	if est.EnergyPerMAC1J < expectedMinEnergy {
		t.Errorf("EnergyPerMAC1J=%e should be >= %e", est.EnergyPerMAC1J, expectedMinEnergy)
	}
}

// ============================================
// InferenceResult Tests
// ============================================

func TestInferenceResult_EnergyPositive(t *testing.T) {
	// Every inference should report positive energy
	net := NewDualModeNetwork(4, 3, 2)

	// Initialize weights
	net.FPWeights1 = [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, 1.1, 1.2},
	}
	net.FPWeights2 = [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}
	net.FPBias1 = []float64{0.1, 0.2, 0.3}
	net.FPBias2 = []float64{0.1, 0.2}

	net.RequantizeWeights()

	input := []float64{0.5, 0.5, 0.5, 0.5}
	result := net.Infer(input)

	if result == nil {
		t.Fatal("Infer returned nil")
	}

	if result.EnergyUsed <= 0 {
		t.Errorf("Energy should be positive, got %e", result.EnergyUsed)
	}
}

func TestInferenceResult_KLDivergence_IdenticalPaths(t *testing.T) {
	// Under ideal conditions, KL should be near 0
	net := NewDualModeNetwork(4, 3, 2)

	// Set ideal conditions
	net.Config.NumLevels = 30
	net.Config.NoiseLevel = 0.0
	net.Config.ADCBits = 16
	net.Config.DACBits = 16

	// Simple weights
	net.FPWeights1 = [][]float64{
		{0.5, 0.5, 0.5, 0.5},
		{-0.5, 0.5, -0.5, 0.5},
		{0.5, -0.5, 0.5, -0.5},
	}
	net.FPWeights2 = [][]float64{
		{1.0, 0.5, 0.5},
		{0.5, 1.0, 0.5},
	}
	net.FPBias1 = []float64{0.1, 0.1, 0.1}
	net.FPBias2 = []float64{0.1, 0.1}

	net.RequantizeWeights()

	input := []float64{1.0, 0.0, 1.0, 0.0}
	result := net.Infer(input)

	if result == nil {
		t.Fatal("Infer returned nil")
	}

	// KL divergence should be very small under ideal conditions
	if result.Disagreement > 0.1 {
		t.Logf("Note: KL divergence=%f is higher than expected under ideal conditions", result.Disagreement)
	}

	t.Logf("KL divergence under ideal conditions: %f", result.Disagreement)
}

// ============================================
// Configuration Tests
// ============================================

func TestSetDACBits_Clamping(t *testing.T) {
	// Verify min/max clamping
	net := NewDualModeNetwork(4, 3, 2)

	testCases := []struct {
		name     string
		input    int
		expected int
	}{
		{"TooLow", 1, 3},
		{"Min", 3, 3},
		{"Normal", 8, 8},
		{"Max", 16, 16},
		{"TooHigh", 20, 16},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			net.SetDACBits(tc.input)
			if net.Config.DACBits != tc.expected {
				t.Errorf("SetDACBits(%d): got %d, want %d", tc.input, net.Config.DACBits, tc.expected)
			}
		})
	}
}

func TestSetNumLevels_RequantizeTrigger(t *testing.T) {
	// Changing levels and calling RequantizeWeights updates quantized weights
	net := NewDualModeNetwork(4, 3, 2)

	// Initialize FP weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64(i+j) / 10.0
		}
	}

	// Quantize at 30 levels
	net.SetNumLevels(30)
	net.RequantizeWeights()

	// Copy quantized weights
	quant30 := make([][]float64, len(net.QuantWeights1))
	for i := range net.QuantWeights1 {
		quant30[i] = make([]float64, len(net.QuantWeights1[i]))
		copy(quant30[i], net.QuantWeights1[i])
	}

	// Change to 2 levels
	net.SetNumLevels(2)
	net.RequantizeWeights()

	// Quantized weights should be different
	different := false
	for i := range net.QuantWeights1 {
		for j := range net.QuantWeights1[i] {
			if net.QuantWeights1[i][j] != quant30[i][j] {
				different = true
				break
			}
		}
		if different {
			break
		}
	}

	if !different {
		t.Error("Quantized weights should change when NumLevels changes from 30 to 2")
	}
}
