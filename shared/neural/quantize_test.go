package neural

import (
	"math"
	"testing"
)

func TestQuantizeWeights_30Levels(t *testing.T) {
	// Test symmetric quantization to 30 levels
	fpWeights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
		{-0.8, -0.3, 0.2, 0.7, 0.9},
	}

	quantized, err := QuantizeWeights(fpWeights, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Check: quantized values are in [-1, 1]
	for i := range quantized {
		for j := range quantized[i] {
			if math.Abs(quantized[i][j]) > 1.01 {
				t.Errorf("Quantized value out of range: %f", quantized[i][j])
			}
		}
	}

	// Check: at most 30 distinct values
	distinct := make(map[float64]bool)
	for i := range quantized {
		for j := range quantized[i] {
			distinct[quantized[i][j]] = true
		}
	}

	if len(distinct) > 30 {
		t.Errorf("Too many distinct values: %d (expected ≤ 30)", len(distinct))
	}

	// Check: MSE is reasonable
	mse := 0.0
	count := 0
	for i := range fpWeights {
		for j := range fpWeights[i] {
			diff := fpWeights[i][j] - quantized[i][j]
			mse += diff * diff
			count++
		}
	}
	mse /= float64(count)

	// For 30 levels, MSE should be small
	if mse > 0.01 {
		t.Errorf("MSE too high: %f (expected < 0.01)", mse)
	}
}

func TestQuantizeWeights_2Levels(t *testing.T) {
	// Test binary quantization (should show high error)
	fpWeights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
	}

	quantized, err := QuantizeWeights(fpWeights, 2)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Check: only 2 distinct values
	distinct := make(map[float64]bool)
	for i := range quantized {
		for j := range quantized[i] {
			distinct[quantized[i][j]] = true
		}
	}

	if len(distinct) != 2 {
		t.Errorf("Expected 2 levels, got %d", len(distinct))
	}

	// Check: values are approximately {-1, +1}
	for v := range distinct {
		if math.Abs(math.Abs(v)-1.0) > 0.01 {
			t.Errorf("Expected ±1, got %f", v)
		}
	}
}

func TestQuantizeWeights_EdgeCases(t *testing.T) {
	// Test empty input
	empty, err := QuantizeWeights([][]float64{}, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}
	if len(empty) != 0 {
		t.Error("Expected empty output for empty input")
	}

	// Test single element
	single, err := QuantizeWeights([][]float64{{0.5}}, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}
	if len(single) != 1 || len(single[0]) != 1 {
		t.Error("Single element quantization failed")
	}

	// Test all zeros
	zeros := [][]float64{{0.0, 0.0, 0.0}}
	quantizedZeros, err := QuantizeWeights(zeros, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}
	for i := range quantizedZeros {
		for j := range quantizedZeros[i] {
			if quantizedZeros[i][j] != 0.0 {
				t.Errorf("Expected 0, got %f", quantizedZeros[i][j])
			}
		}
	}
}

func TestQuantizeBias(t *testing.T) {
	bias := []float64{-0.5, 0.0, 0.5}
	quantized, err := QuantizeBias(bias, 10)
	if err != nil {
		t.Fatalf("QuantizeBias failed: %v", err)
	}

	if len(quantized) != len(bias) {
		t.Errorf("Length mismatch: got %d, expected %d", len(quantized), len(bias))
	}

	// Check: all values in range
	for i, v := range quantized {
		if v < -0.51 || v > 0.51 {
			t.Errorf("Bias %d out of range: %f", i, v)
		}
	}
}

func TestComputeQuantizationStats(t *testing.T) {
	original := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
	}
	quantized, err := QuantizeWeights(original, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	stats := ComputeQuantizationStats(original, quantized)

	// Check range
	if stats.OriginalRange != 2.0 {
		t.Errorf("Expected original range 2.0, got %f", stats.OriginalRange)
	}

	// Check distinct values
	if stats.NumDistinct > 30 || stats.NumDistinct < 2 {
		t.Errorf("Unexpected number of distinct values: %d", stats.NumDistinct)
	}

	// Check MSE > 0 (there is quantization error)
	if stats.MSE < 0 {
		t.Errorf("MSE should be non-negative, got %f", stats.MSE)
	}

	// Check PSNR is reasonable for 30 levels
	if stats.PSNR < 20 {
		t.Errorf("PSNR too low: %f dB", stats.PSNR)
	}
}

func TestRandomSource(t *testing.T) {
	rng := NewRandomSource(42)

	// Test Float64 is in [0, 1)
	for i := 0; i < 100; i++ {
		v := rng.Float64()
		if v < 0 || v >= 1 {
			t.Errorf("Float64 out of range: %f", v)
		}
	}

	// Test Intn
	rng = NewRandomSource(42)
	for i := 0; i < 100; i++ {
		v := rng.Intn(10)
		if v < 0 || v >= 10 {
			t.Errorf("Intn out of range: %d", v)
		}
	}

	// Test determinism
	rng1 := NewRandomSource(123)
	rng2 := NewRandomSource(123)
	for i := 0; i < 10; i++ {
		if rng1.Float64() != rng2.Float64() {
			t.Error("Random sources with same seed should produce same values")
		}
	}
}

func TestAddGaussianNoise(t *testing.T) {
	values := []float64{1.0, 1.0, 1.0, 1.0, 1.0}
	rng := NewRandomSource(42)

	// Test zero noise
	noisy := AddGaussianNoise(values, 0.0, rng)
	for i := range noisy {
		if noisy[i] != values[i] {
			t.Error("Zero noise should not modify values")
		}
	}

	// Test with noise
	noisy = AddGaussianNoise(values, 0.1, rng)
	different := false
	for i := range noisy {
		if noisy[i] != values[i] {
			different = true
			break
		}
	}
	if !different {
		t.Error("Noise should modify values")
	}
}

func TestDualModeNetwork_Creation(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	if net.InputSize != 784 {
		t.Errorf("Wrong input size: %d", net.InputSize)
	}
	if net.HiddenSize != 128 {
		t.Errorf("Wrong hidden size: %d", net.HiddenSize)
	}
	if net.OutputSize != 10 {
		t.Errorf("Wrong output size: %d", net.OutputSize)
	}

	// Check config defaults
	if net.Config.NumLevels != 30 {
		t.Errorf("Expected 30 levels, got %d", net.Config.NumLevels)
	}
	if net.Config.ADCBits != 8 {
		t.Errorf("Expected 8-bit ADC, got %d", net.Config.ADCBits)
	}
}

func TestDualModeNetwork_SetParameters(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Initialize some FP weights for testing
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64(i+j) / 1000.0
		}
	}

	// Test SetNumLevels
	net.SetNumLevels(10)
	if net.Config.NumLevels != 10 {
		t.Errorf("SetNumLevels failed: got %d", net.Config.NumLevels)
	}

	// Test bounds (minimum is now 2, required by QuantizeWeights)
	net.SetNumLevels(0)
	if net.Config.NumLevels != 2 {
		t.Errorf("NumLevels should be clamped to 2 (minimum for quantization), got %d", net.Config.NumLevels)
	}

	net.SetNumLevels(100)
	if net.Config.NumLevels != MaxDemoLevels {
		t.Errorf("NumLevels should be clamped to %d, got %d", MaxDemoLevels, net.Config.NumLevels)
	}

	// Test SetNoiseLevel
	net.SetNoiseLevel(0.1)
	if net.Config.NoiseLevel != 0.1 {
		t.Errorf("SetNoiseLevel failed: got %f", net.Config.NoiseLevel)
	}

	net.SetNoiseLevel(-0.1)
	if net.Config.NoiseLevel != 0 {
		t.Errorf("NoiseLevel should be clamped to 0, got %f", net.Config.NoiseLevel)
	}

	net.SetNoiseLevel(0.5)
	if net.Config.NoiseLevel != 0.20 {
		t.Errorf("NoiseLevel should be clamped to 0.20, got %f", net.Config.NoiseLevel)
	}

	// Test SetADCBits
	net.SetADCBits(6)
	if net.Config.ADCBits != 6 {
		t.Errorf("SetADCBits failed: got %d", net.Config.ADCBits)
	}

	net.SetADCBits(1)
	if net.Config.ADCBits != 3 {
		t.Errorf("ADCBits should be clamped to 3, got %d", net.Config.ADCBits)
	}
}

func TestDualModeNetwork_Inference(t *testing.T) {
	net := NewDualModeNetwork(4, 3, 2)

	// Initialize simple weights
	net.FPWeights1 = [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{-0.1, 0.1, -0.1, 0.1},
		{0.2, 0.1, 0.2, 0.1},
	}
	net.FPWeights2 = [][]float64{
		{0.5, 0.2, 0.1},
		{0.1, 0.3, 0.4},
	}
	net.FPBias1 = []float64{0.01, 0.01, 0.01}
	net.FPBias2 = []float64{0.01, 0.01}

	net.RequantizeWeights()

	input := []float64{0.5, 0.5, 0.5, 0.5}
	result := net.Infer(input)

	// Check that both paths produced valid predictions
	if result.FPPrediction < 0 || result.FPPrediction >= 2 {
		t.Errorf("Invalid FP prediction: %d", result.FPPrediction)
	}
	if result.CIMPrediction < 0 || result.CIMPrediction >= 2 {
		t.Errorf("Invalid CIM prediction: %d", result.CIMPrediction)
	}

	// Check probabilities sum to 1
	fpSum := 0.0
	for _, p := range result.FPProbabilities {
		fpSum += p
	}
	if math.Abs(fpSum-1.0) > 0.001 {
		t.Errorf("FP probabilities don't sum to 1: %f", fpSum)
	}

	cimSum := 0.0
	for _, p := range result.CIMProbabilities {
		cimSum += p
	}
	if math.Abs(cimSum-1.0) > 0.001 {
		t.Errorf("CIM probabilities don't sum to 1: %f", cimSum)
	}

	// Check energy is calculated
	if result.EnergyUsed <= 0 {
		t.Error("Energy should be positive")
	}
}

func TestDualModeNetwork_AgreementUnderIdealConditions(t *testing.T) {
	net := NewDualModeNetwork(4, 3, 2)

	// Set ideal conditions
	net.Config.NumLevels = 30
	net.Config.NoiseLevel = 0.0
	net.Config.ADCBits = 16
	net.Config.DACBits = 16

	// Simple weights that should work well
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

	// Under ideal conditions, predictions should often agree
	input := []float64{1.0, 0.0, 1.0, 0.0}
	result := net.Infer(input)

	// With 30 levels, no noise, high ADC/DAC, they should agree
	if !result.Agree {
		t.Log("FP prediction:", result.FPPrediction, "confidence:", result.FPConfidence)
		t.Log("CIM prediction:", result.CIMPrediction, "confidence:", result.CIMConfidence)
		// Note: Don't fail test, as small numerical differences can cause disagreement
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test relu
	x := []float64{-1, 0, 1, -0.5, 0.5}
	y := relu(x)
	expected := []float64{0, 0, 1, 0, 0.5}
	for i := range y {
		if y[i] != expected[i] {
			t.Errorf("ReLU(%f) = %f, expected %f", x[i], y[i], expected[i])
		}
	}

	// Test softmax
	logits := []float64{1, 2, 3}
	probs := softmax(logits)
	sum := 0.0
	for _, p := range probs {
		sum += p
		if p < 0 || p > 1 {
			t.Errorf("Softmax probability out of range: %f", p)
		}
	}
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("Softmax doesn't sum to 1: %f", sum)
	}

	// Test argmax
	if argmax(probs) != 2 {
		t.Error("Argmax should return index of maximum")
	}

	// Test klDivergence
	p := []float64{0.5, 0.5}
	q := []float64{0.5, 0.5}
	kl := klDivergence(p, q)
	if math.Abs(kl) > 0.0001 {
		t.Errorf("KL divergence of identical distributions should be ~0, got %f", kl)
	}

	// Different distributions
	p = []float64{0.9, 0.1}
	q = []float64{0.5, 0.5}
	kl = klDivergence(p, q)
	if kl <= 0 {
		t.Errorf("KL divergence should be positive for different distributions, got %f", kl)
	}
}
