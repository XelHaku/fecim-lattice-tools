// Package core provides physics and calculation tests for FeCIM MNIST demo.
// These tests verify the physical accuracy of the 30-level ferroelectric model.
package core

import (
	"math"
	"testing"
)

// =============================================================================
// PHYSICS TESTS: 30-Level Ferroelectric Quantization
// =============================================================================

// TestFeCIM30LevelPhysics verifies that quantization matches FeCIM hardware specs.
// Dr. Tour's research: HZO ferroelectric achieves 30 stable polarization states.
func TestFeCIM30LevelPhysics(t *testing.T) {
	// Test case: full range weights
	weights := [][]float64{
		{-1.0, -0.9, -0.8, -0.7, -0.6, -0.5, -0.4, -0.3, -0.2, -0.1},
		{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9},
		{1.0, -0.95, -0.85, 0.95, 0.85, -0.75, 0.75, -0.65, 0.65, 0.0},
	}

	quantized, err := QuantizeWeights(weights, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Count distinct levels
	distinct := make(map[float64]bool)
	for i := range quantized {
		for j := range quantized[i] {
			distinct[quantized[i][j]] = true
		}
	}

	// FeCIM constraint: exactly 30 levels maximum
	if len(distinct) > 30 {
		t.Errorf("PHYSICS ERROR: FeCIM can only store 30 levels, got %d", len(distinct))
	}

	t.Logf("FeCIM quantization: %d distinct levels used (max 30)", len(distinct))
}

// TestLevelSpacingUniformity verifies uniform level spacing (linear DAC).
// FeCIM uses linear mapping: conductance proportional to polarization.
func TestLevelSpacingUniformity(t *testing.T) {
	// Create weights spanning full range
	weights := make([][]float64, 1)
	weights[0] = make([]float64, 100)
	for i := 0; i < 100; i++ {
		weights[0][i] = -1.0 + 2.0*float64(i)/99.0
	}

	quantized, err := QuantizeWeights(weights, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Collect distinct values
	valueSet := make(map[float64]bool)
	for _, v := range quantized[0] {
		valueSet[v] = true
	}

	// Convert to sorted slice
	values := make([]float64, 0, len(valueSet))
	for v := range valueSet {
		values = append(values, v)
	}

	// Sort values
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	if len(values) < 2 {
		t.Skip("Not enough distinct values to check spacing")
	}

	// Check spacing uniformity (should be constant for linear mapping)
	expectedSpacing := (values[len(values)-1] - values[0]) / float64(len(values)-1)
	maxSpacingError := 0.0

	for i := 1; i < len(values); i++ {
		spacing := values[i] - values[i-1]
		spacingError := math.Abs(spacing - expectedSpacing)
		if spacingError > maxSpacingError {
			maxSpacingError = spacingError
		}
	}

	// Spacing should be uniform within 1% tolerance
	if maxSpacingError > expectedSpacing*0.01 {
		t.Errorf("PHYSICS ERROR: Non-uniform level spacing (error: %.4f%%)", maxSpacingError/expectedSpacing*100)
	}

	t.Logf("Level spacing uniformity verified: max error %.4f%%", maxSpacingError/expectedSpacing*100)
}

// TestQuantizationSymmetry verifies symmetric quantization around zero.
// FeCIM uses bipolar weights: both positive and negative polarization.
func TestQuantizationSymmetry(t *testing.T) {
	// Create symmetric weights
	weights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
	}

	quantized, err := QuantizeWeights(weights, 30)
	if err != nil {
		t.Fatalf("QuantizeWeights failed: %v", err)
	}

	// Check symmetry: q(-x) should approximately equal -q(x)
	tolerance := 0.001
	for i := 0; i < len(quantized[0])/2; i++ {
		negIdx := i
		posIdx := len(quantized[0]) - 1 - i
		sum := quantized[0][negIdx] + quantized[0][posIdx]
		if math.Abs(sum) > tolerance {
			t.Errorf("PHYSICS ERROR: Asymmetric quantization at indices %d,%d: sum=%.6f", negIdx, posIdx, sum)
		}
	}
}

// TestQuantizationCliff validates the "quantization cliff" phenomenon.
// Dr. Tour's insight: binary (2-level) weights destroy network accuracy.
func TestQuantizationCliff(t *testing.T) {
	// Create varied weights
	weights := [][]float64{
		{-0.8, -0.6, -0.4, -0.2, 0.0, 0.2, 0.4, 0.6, 0.8},
	}

	testCases := []struct {
		levels         int
		maxExpectedMSE float64
		description    string
	}{
		{2, 0.5, "Binary (quantization cliff)"},
		{4, 0.15, "2-bit"},
		{8, 0.05, "3-bit"},
		{16, 0.02, "4-bit"},
		{30, 0.005, "FeCIM (30 levels)"},
	}

	for _, tc := range testCases {
		quantized, err := QuantizeWeights(weights, tc.levels)
		if err != nil {
			t.Fatalf("QuantizeWeights failed: %v", err)
		}
		stats := ComputeQuantizationStats(weights, quantized)

		if stats.MSE > tc.maxExpectedMSE {
			t.Errorf("%s: MSE %.4f exceeds expected max %.4f", tc.description, stats.MSE, tc.maxExpectedMSE)
		}

		t.Logf("%s (%d levels): MSE=%.6f, PSNR=%.1fdB", tc.description, tc.levels, stats.MSE, stats.PSNR)
	}
}

// =============================================================================
// CALCULATION TESTS: Neural Network Math
// =============================================================================

// TestSoftmaxNumericalStability verifies softmax handles extreme values.
func TestSoftmaxNumericalStability(t *testing.T) {
	testCases := []struct {
		name   string
		logits []float64
	}{
		{"Normal", []float64{1.0, 2.0, 3.0}},
		{"Large positive", []float64{100.0, 200.0, 300.0}},
		{"Large negative", []float64{-100.0, -200.0, -300.0}},
		{"Mixed extreme", []float64{-500.0, 0.0, 500.0}},
		{"All same", []float64{5.0, 5.0, 5.0, 5.0}},
		{"Near zero", []float64{0.001, 0.002, 0.003}},
	}

	for _, tc := range testCases {
		probs := softmax(tc.logits)

		// Check no NaN or Inf
		for i, p := range probs {
			if math.IsNaN(p) || math.IsInf(p, 0) {
				t.Errorf("%s: NaN/Inf at index %d", tc.name, i)
			}
		}

		// Check probabilities sum to 1
		sum := 0.0
		for _, p := range probs {
			sum += p
		}
		if math.Abs(sum-1.0) > 1e-6 {
			t.Errorf("%s: probabilities sum to %.10f, expected 1.0", tc.name, sum)
		}

		// Check all probabilities are in [0, 1]
		for i, p := range probs {
			if p < 0 || p > 1 {
				t.Errorf("%s: probability[%d]=%.6f out of range [0,1]", tc.name, i, p)
			}
		}
	}
}

// TestReLUCorrectness verifies ReLU activation function.
func TestReLUCorrectness(t *testing.T) {
	input := []float64{-10.0, -1.0, -0.001, 0.0, 0.001, 1.0, 10.0}
	expected := []float64{0.0, 0.0, 0.0, 0.0, 0.001, 1.0, 10.0}

	output := relu(input)

	for i := range output {
		if math.Abs(output[i]-expected[i]) > 1e-10 {
			t.Errorf("ReLU(%f) = %f, expected %f", input[i], output[i], expected[i])
		}
	}
}

// TestKLDivergenceProperties verifies KL divergence mathematical properties.
func TestKLDivergenceProperties(t *testing.T) {
	// Property 1: KL(P||P) = 0
	p := []float64{0.25, 0.25, 0.25, 0.25}
	kl := klDivergence(p, p)
	if math.Abs(kl) > 1e-6 {
		t.Errorf("KL(P||P) should be ~0, got %f", kl)
	}

	// Property 2: KL(P||Q) >= 0 (Gibbs' inequality)
	q := []float64{0.1, 0.2, 0.3, 0.4}
	kl = klDivergence(p, q)
	if kl < -1e-10 {
		t.Errorf("KL divergence should be non-negative, got %f", kl)
	}

	// Property 3: KL(P||Q) != KL(Q||P) in general (asymmetry)
	klPQ := klDivergence(p, q)
	klQP := klDivergence(q, p)
	if math.Abs(klPQ-klQP) < 1e-10 {
		t.Log("Warning: KL divergence happened to be symmetric for this test case")
	}

	t.Logf("KL(P||Q)=%.6f, KL(Q||P)=%.6f", klPQ, klQP)
}

// TestArgmaxCorrectness verifies argmax function.
func TestArgmaxCorrectness(t *testing.T) {
	testCases := []struct {
		input    []float64
		expected int
	}{
		{[]float64{1.0, 2.0, 3.0}, 2},
		{[]float64{3.0, 2.0, 1.0}, 0},
		{[]float64{1.0, 3.0, 2.0}, 1},
		{[]float64{5.0}, 0},
		{[]float64{0.0, 0.0, 0.0}, 0}, // First max for ties
	}

	for _, tc := range testCases {
		result := argmax(tc.input)
		if result != tc.expected {
			t.Errorf("argmax(%v) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

// =============================================================================
// ENERGY CALCULATION TESTS
// =============================================================================

// TestEnergyCalculation verifies FeCIM energy estimates.
// Verified claim: 25-100× more efficient than NAND (Samsung Nature 2025).
func TestEnergyCalculation(t *testing.T) {
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
	net.RequantizeWeights()

	input := make([]float64, 784)
	for i := range input {
		input[i] = 0.5
	}

	result := net.Infer(input)

	// Energy should be positive
	if result.EnergyUsed <= 0 {
		t.Error("Energy should be positive")
	}

	// MNIST network has: 784*128 + 128*10 = 101,632 MACs
	// FeCIM energy: ~50 fJ/MAC = 5.08 µJ total
	expectedMACs := 784*128 + 128*10
	expectedEnergy := float64(expectedMACs) * 50e-15 * 1e6 // Convert to µJ

	// Allow 2x tolerance for estimation
	if result.EnergyUsed < expectedEnergy*0.5 || result.EnergyUsed > expectedEnergy*2.0 {
		t.Errorf("Energy %.2f µJ outside expected range [%.2f, %.2f] µJ",
			result.EnergyUsed, expectedEnergy*0.5, expectedEnergy*2.0)
	}

	t.Logf("FeCIM energy: %.2f µJ for %d MACs (%.2f fJ/MAC)",
		result.EnergyUsed, expectedMACs, result.EnergyUsed*1e6/float64(expectedMACs))
}

// TestEnergyScalesWithLevels verifies that energy changes with quantization levels.
// More levels = more bits = higher energy per MAC.
func TestEnergyScalesWithLevels(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

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

	input := make([]float64, 784)
	for i := range input {
		input[i] = 0.5
	}

	// Test energy at different levels
	levelsToTest := []int{2, 8, 16, 30}
	var prevEnergy float64

	for _, levels := range levelsToTest {
		net.SetNumLevels(levels)
		result := net.Infer(input)

		t.Logf("Levels=%2d: Energy=%.4f µJ (%.2f fJ/MAC)",
			levels, result.EnergyUsed, result.EnergyUsed*1e6/101632)

		// Energy should increase with more levels
		if prevEnergy > 0 && result.EnergyUsed <= prevEnergy {
			t.Errorf("Energy should increase with more levels: %d levels (%.4f µJ) <= previous (%.4f µJ)",
				levels, result.EnergyUsed, prevEnergy)
		}
		prevEnergy = result.EnergyUsed
	}

	// Verify 30-level is about 5x more energy than 2-level
	// (log2(30) ≈ 4.9 bits vs log2(2) = 1 bit)
	net.SetNumLevels(2)
	result2 := net.Infer(input)
	net.SetNumLevels(30)
	result30 := net.Infer(input)

	ratio := result30.EnergyUsed / result2.EnergyUsed
	expectedRatio := 4.9 // log2(30) / log2(2)

	if ratio < expectedRatio*0.8 || ratio > expectedRatio*1.2 {
		t.Errorf("Energy ratio 30/2 levels = %.2f, expected ~%.2f", ratio, expectedRatio)
	}

	t.Logf("Energy ratio (30 levels / 2 levels) = %.2fx (expected ~%.1fx)", ratio, expectedRatio)
}

// =============================================================================
// DUAL-MODE INFERENCE TESTS
// =============================================================================

// TestDualModeAgreementVsLevels tests how FP/CIM agreement varies with quantization.
func TestDualModeAgreementVsLevels(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize deterministic weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i+j)%10-5) / 10.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i+j)%10-5) / 10.0
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)

	input := make([]float64, 16)
	for i := range input {
		input[i] = float64(i%2) * 0.5
	}

	// Test different quantization levels
	levels := []int{2, 4, 8, 16, 30}

	for _, numLevels := range levels {
		net.Config.NumLevels = numLevels
		net.Config.NoiseLevel = 0.0 // No noise for deterministic test
		net.RequantizeWeights()

		result := net.Infer(input)

		t.Logf("Levels=%2d: FP=%d(%.1f%%), CIM=%d(%.1f%%), Agree=%v, KL=%.4f",
			numLevels,
			result.FPPrediction, result.FPConfidence*100,
			result.CIMPrediction, result.CIMConfidence*100,
			result.Agree, result.Disagreement)
	}
}

// TestNoiseImpactOnAccuracy tests how noise affects CIM inference.
func TestNoiseImpactOnAccuracy(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i*3+j*7)%20-10) / 20.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i*5+j*3)%20-10) / 20.0
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)

	input := make([]float64, 16)
	for i := range input {
		input[i] = 0.5
	}

	// Test different noise levels
	noiseLevels := []float64{0.0, 0.01, 0.05, 0.10, 0.15, 0.20}

	for _, noise := range noiseLevels {
		net.Config.NumLevels = 30
		net.Config.NoiseLevel = noise
		net.RequantizeWeights()

		// Run multiple times to get average disagreement
		totalKL := 0.0
		agreements := 0
		runs := 10

		for r := 0; r < runs; r++ {
			result := net.Infer(input)
			totalKL += result.Disagreement
			if result.Agree {
				agreements++
			}
		}

		avgKL := totalKL / float64(runs)
		agreeRate := float64(agreements) / float64(runs) * 100

		t.Logf("Noise=%.2f: Agreement=%.0f%%, AvgKL=%.4f", noise, agreeRate, avgKL)
	}
}

// TestADCBitsImpact tests how ADC resolution affects inference.
func TestADCBitsImpact(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i+j)%10-5) / 10.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i+j)%10-5) / 10.0
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)

	input := make([]float64, 16)
	for i := range input {
		input[i] = 0.5
	}

	// Test different ADC resolutions
	adcBits := []int{3, 4, 5, 6, 8, 10, 12, 16}

	net.Config.NumLevels = 30
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	for _, bits := range adcBits {
		net.Config.ADCBits = bits

		result := net.Infer(input)

		t.Logf("ADC=%2d bits: FP=%d, CIM=%d, Agree=%v, KL=%.6f",
			bits, result.FPPrediction, result.CIMPrediction, result.Agree, result.Disagreement)
	}
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkQuantize30Levels(b *testing.B) {
	// MNIST layer 1: 784x128
	weights := make([][]float64, 128)
	for i := range weights {
		weights[i] = make([]float64, 784)
		for j := range weights[i] {
			weights[i][j] = float64((i*j)%1000-500) / 1000.0
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = QuantizeWeights(weights, 30) // levels=30 is always valid
	}
}

// =============================================================================
// PER-LAYER PTQ TESTS
// =============================================================================

// TestPerLayerQuantization verifies that per-layer quantization works correctly.
func TestPerLayerQuantization(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i+j)%10-5) / 10.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i*2+j*3)%10-5) / 10.0
		}
	}
	net.FPBias1 = make([]float64, 8)
	net.FPBias2 = make([]float64, 4)

	// Test uniform quantization first
	net.SetNumLevels(30)
	net.SetPerLayerQuant(false)
	net.RequantizeWeights()

	stats1Uniform, stats2Uniform := net.GetQuantizationStats()
	t.Logf("Uniform 30-level: L1 distinct=%d, L2 distinct=%d",
		stats1Uniform.NumDistinct, stats2Uniform.NumDistinct)

	// Test per-layer quantization with different levels
	net.SetPerLayerLevels(30, 10) // Layer1: 30 levels, Layer2: 10 levels

	stats1PerLayer, stats2PerLayer := net.GetQuantizationStats()
	t.Logf("Per-layer (30,10): L1 distinct=%d, L2 distinct=%d",
		stats1PerLayer.NumDistinct, stats2PerLayer.NumDistinct)

	// Layer 2 should have fewer distinct values with fewer levels
	if stats2PerLayer.NumDistinct > 10 {
		t.Errorf("Layer 2 should have at most 10 distinct values, got %d", stats2PerLayer.NumDistinct)
	}

	// Layer 1 should still use up to 30 levels
	if stats1PerLayer.NumDistinct > 30 {
		t.Errorf("Layer 1 should have at most 30 distinct values, got %d", stats1PerLayer.NumDistinct)
	}
}

// TestPerLayerQuantizationEnergyCalculation verifies energy calculation with per-layer PTQ.
func TestPerLayerQuantizationEnergyCalculation(t *testing.T) {
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

	input := make([]float64, 784)
	for i := range input {
		input[i] = 0.5
	}

	// Test uniform 30-level quantization
	net.SetNumLevels(30)
	net.SetPerLayerQuant(false)
	net.RequantizeWeights()
	resultUniform := net.Infer(input)

	// Test per-layer with different levels
	net.SetPerLayerLevels(30, 8) // Layer1: 30 levels (4.9 bits), Layer2: 8 levels (3 bits)
	resultPerLayer := net.Infer(input)

	// Per-layer should use less energy (layer 2 uses fewer bits)
	t.Logf("Uniform (30,30): Energy=%.4f µJ", resultUniform.EnergyUsed)
	t.Logf("Per-layer (30,8): Energy=%.4f µJ", resultPerLayer.EnergyUsed)

	// Layer 2 has 128*10 = 1280 MACs out of 101632 total (1.3%)
	// So energy difference should be small but measurable
	if resultPerLayer.EnergyUsed >= resultUniform.EnergyUsed {
		t.Errorf("Per-layer (30,8) should use less energy than uniform (30,30)")
	}
}

// TestPerLayerQuantizationConfig verifies the config getters/setters work correctly.
func TestPerLayerQuantizationConfig(t *testing.T) {
	net := NewDualModeNetwork(16, 8, 4)

	// Default should not have per-layer enabled
	if net.IsPerLayerQuant() {
		t.Error("PerLayerQuant should be disabled by default")
	}

	// Enable per-layer and set levels
	net.SetPerLayerLevels(20, 15)

	enabled, l1, l2 := net.GetPerLayerQuantInfo()
	if !enabled {
		t.Error("PerLayerQuant should be enabled after SetPerLayerLevels")
	}
	if l1 != 20 {
		t.Errorf("Layer1Levels should be 20, got %d", l1)
	}
	if l2 != 15 {
		t.Errorf("Layer2Levels should be 15, got %d", l2)
	}

	// Test individual setters
	net.SetLayer1Levels(25)
	if net.GetLayer1Levels() != 25 {
		t.Errorf("Layer1Levels should be 25 after SetLayer1Levels")
	}

	net.SetLayer2Levels(10)
	if net.GetLayer2Levels() != 10 {
		t.Errorf("Layer2Levels should be 10 after SetLayer2Levels")
	}

	// Test clamping
	net.SetLayer1Levels(100) // Should clamp to MaxDemoLevels
	if net.GetLayer1Levels() > MaxDemoLevels {
		t.Errorf("Layer1Levels should be clamped to max %d", MaxDemoLevels)
	}

	net.SetLayer2Levels(1) // Should clamp to 2
	if net.GetLayer2Levels() < 2 {
		t.Error("Layer2Levels should be clamped to min 2")
	}
}

func BenchmarkDualModeInference(b *testing.B) {
	net := NewDualModeNetwork(784, 128, 10)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i+j)%100-50) / 100.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i+j)%100-50) / 100.0
		}
	}
	net.RequantizeWeights()

	input := make([]float64, 784)
	for i := range input {
		input[i] = float64(i%256) / 255.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		net.Infer(input)
	}
}
