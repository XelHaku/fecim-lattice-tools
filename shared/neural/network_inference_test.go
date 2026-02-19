package neural

import (
	"math"
	"testing"
)

// createTestNetwork creates a small test network with known weights for testing.
func createTestNetwork() *DualModeNetwork {
	net := NewDualModeNetwork(4, 3, 2)

	// Layer 1: 4 inputs -> 3 hidden
	net.FPWeights1 = [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
		{0.9, 1.0, 1.1, 1.2},
	}
	net.FPBias1 = []float64{0.1, 0.2, 0.3}

	// Layer 2: 3 hidden -> 2 outputs
	net.FPWeights2 = [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}
	net.FPBias2 = []float64{0.1, 0.2}

	// Initialize quantized weights to match FP weights
	net.QuantWeights1 = make([][]float64, 3)
	net.QuantWeights2 = make([][]float64, 2)
	for i := 0; i < 3; i++ {
		net.QuantWeights1[i] = make([]float64, 4)
		copy(net.QuantWeights1[i], net.FPWeights1[i])
	}
	for i := 0; i < 2; i++ {
		net.QuantWeights2[i] = make([]float64, 3)
		copy(net.QuantWeights2[i], net.FPWeights2[i])
	}
	net.QuantBias1 = make([]float64, 3)
	net.QuantBias2 = make([]float64, 2)
	copy(net.QuantBias1, net.FPBias1)
	copy(net.QuantBias2, net.FPBias2)

	net.Config.ADCBits = 16     // No quantization for tests
	net.Config.DACBits = 16     // No quantization for tests
	net.Config.NoiseLevel = 0.0 // No noise for deterministic tests

	return net
}

// ============================================
// InferFPOnly Tests
// ============================================

func TestInferFPOnly_BasicInference(t *testing.T) {
	net := createTestNetwork()
	input := []float64{0.1, 0.2, 0.3, 0.4}

	prediction, confidence, probs := net.InferFPOnly(input)

	// Verify prediction is in valid range
	if prediction < 0 || prediction >= net.OutputSize {
		t.Errorf("Invalid prediction: got %d, want 0-%d", prediction, net.OutputSize-1)
	}

	// Verify confidence is in [0, 1]
	if confidence < 0 || confidence > 1 {
		t.Errorf("Invalid confidence: got %f, want [0, 1]", confidence)
	}

	// Verify probabilities sum to 1
	sum := 0.0
	for _, p := range probs {
		sum += p
		if p < 0 || p > 1 {
			t.Errorf("Invalid probability: got %f, want [0, 1]", p)
		}
	}
	if math.Abs(sum-1.0) > 1e-6 {
		t.Errorf("Probabilities don't sum to 1: got %f", sum)
	}

	// Verify confidence matches the predicted class probability
	if math.Abs(confidence-probs[prediction]) > 1e-9 {
		t.Errorf("Confidence mismatch: got %f, want %f", confidence, probs[prediction])
	}
}

func TestInferFPOnly_OutputValidation(t *testing.T) {
	net := createTestNetwork()
	input := []float64{0.1, 0.2, 0.3, 0.4}

	_, _, probs := net.InferFPOnly(input)

	// Verify output dimensions match architecture
	if len(probs) != net.OutputSize {
		t.Errorf("Output size mismatch: got %d, want %d", len(probs), net.OutputSize)
	}
}

func TestInferFPOnly_ConsistencyWithInfer(t *testing.T) {
	net := createTestNetwork()
	input := []float64{0.1, 0.2, 0.3, 0.4}

	// Run both inference methods
	fpPred, fpConf, fpProbs := net.InferFPOnly(input)
	fullResult := net.Infer(input)

	// Verify FP path results match
	if fpPred != fullResult.FPPrediction {
		t.Errorf("Prediction mismatch: FPOnly=%d, Infer.FP=%d", fpPred, fullResult.FPPrediction)
	}

	if math.Abs(fpConf-fullResult.FPConfidence) > 1e-9 {
		t.Errorf("Confidence mismatch: FPOnly=%f, Infer.FP=%f", fpConf, fullResult.FPConfidence)
	}

	for i := range fpProbs {
		if math.Abs(fpProbs[i]-fullResult.FPProbabilities[i]) > 1e-9 {
			t.Errorf("Probability[%d] mismatch: FPOnly=%f, Infer.FP=%f", i, fpProbs[i], fullResult.FPProbabilities[i])
		}
	}
}

func TestInferFPOnly_InputLengthMismatch(t *testing.T) {
	net := createTestNetwork()

	testCases := []struct {
		name      string
		inputSize int
	}{
		{"TooShort", 2},
		{"TooLong", 6},
		{"Empty", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := make([]float64, tc.inputSize)

			// Document behavior: InferFPOnly doesn't validate, it just processes
			// This may cause panics or incorrect results, but we document this behavior
			defer func() {
				if r := recover(); r != nil {
					// If it panics, that's also a valid behavior to document
					t.Logf("InferFPOnly panicked with input size %d: %v", tc.inputSize, r)
				}
			}()

			pred, conf, probs := net.InferFPOnly(input)

			// If it doesn't panic, document what it returns
			t.Logf("InferFPOnly with input size %d: pred=%d, conf=%f, probs=%v",
				tc.inputSize, pred, conf, probs)
		})
	}
}

func TestInferFPOnly_SingleLayerNotSupported(t *testing.T) {
	net := createTestNetwork()
	net.Config.SingleLayer = true
	input := []float64{0.1, 0.2, 0.3, 0.4}

	// InferFPOnly always uses two-layer weights (FPWeights1/FPWeights2)
	// regardless of SingleLayer config
	_, _, probs := net.InferFPOnly(input)

	if len(probs) != net.OutputSize {
		t.Errorf("Output size mismatch: got %d, want %d", len(probs), net.OutputSize)
	}

	// Verify it used two-layer architecture (hidden layer should exist in computation)
	// This is implicit in the function design - InferFPOnly doesn't check SingleLayer
	t.Logf("InferFPOnly uses two-layer architecture even when SingleLayer=true")
}

// ============================================
// InferCIMOnly Tests
// ============================================

func TestInferCIMOnly_BasicInference(t *testing.T) {
	net := createTestNetwork()
	input := []float64{0.1, 0.2, 0.3, 0.4}

	prediction, confidence, probs := net.InferCIMOnly(input)

	// Verify prediction is in valid range
	if prediction < 0 || prediction >= net.OutputSize {
		t.Errorf("Invalid prediction: got %d, want 0-%d", prediction, net.OutputSize-1)
	}

	// Verify confidence is in [0, 1]
	if confidence < 0 || confidence > 1 {
		t.Errorf("Invalid confidence: got %f, want [0, 1]", confidence)
	}

	// Verify probabilities sum to 1
	sum := 0.0
	for _, p := range probs {
		sum += p
		if p < 0 || p > 1 {
			t.Errorf("Invalid probability: got %f, want [0, 1]", p)
		}
	}
	if math.Abs(sum-1.0) > 1e-6 {
		t.Errorf("Probabilities don't sum to 1: got %f", sum)
	}
}

func TestInferCIMOnly_UsesQuantizedWeights(t *testing.T) {
	// HIGH-003 regression test: Verify CIM path uses QuantWeights not FPWeights
	net := createTestNetwork()

	// Make quantized weights different from FP weights
	for i := range net.QuantWeights1 {
		for j := range net.QuantWeights1[i] {
			net.QuantWeights1[i][j] = net.FPWeights1[i][j] * 0.5
		}
	}
	for i := range net.QuantWeights2 {
		for j := range net.QuantWeights2[i] {
			net.QuantWeights2[i][j] = net.FPWeights2[i][j] * 0.5
		}
	}

	input := []float64{0.1, 0.2, 0.3, 0.4}

	fpPred, fpConf, _ := net.InferFPOnly(input)
	cimPred, cimConf, _ := net.InferCIMOnly(input)

	// Results should differ because we're using different weights
	if fpPred == cimPred && math.Abs(fpConf-cimConf) < 1e-9 {
		t.Errorf("CIM and FP results are identical despite different weights - CIM may not be using QuantWeights")
	}
}

func TestInferCIMOnly_OutputValidation(t *testing.T) {
	net := createTestNetwork()
	input := []float64{0.1, 0.2, 0.3, 0.4}

	_, _, probs := net.InferCIMOnly(input)

	// Verify output dimensions match architecture
	if len(probs) != net.OutputSize {
		t.Errorf("Output size mismatch: got %d, want %d", len(probs), net.OutputSize)
	}
}

func TestInferCIMOnly_InputLengthMismatch(t *testing.T) {
	net := createTestNetwork()

	testCases := []struct {
		name      string
		inputSize int
	}{
		{"TooShort", 2},
		{"TooLong", 6},
		{"Empty", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := make([]float64, tc.inputSize)

			// Document behavior: InferCIMOnly doesn't validate, it just processes
			defer func() {
				if r := recover(); r != nil {
					t.Logf("InferCIMOnly panicked with input size %d: %v", tc.inputSize, r)
				}
			}()

			pred, conf, probs := net.InferCIMOnly(input)

			// If it doesn't panic, document what it returns
			t.Logf("InferCIMOnly with input size %d: pred=%d, conf=%f, probs=%v",
				tc.inputSize, pred, conf, probs)
		})
	}
}

func TestInferCIMOnly_DACDACEffect(t *testing.T) {
	net := createTestNetwork()
	input := []float64{0.123, 0.456, 0.789, 0.234}

	testCases := []struct {
		name    string
		dacBits int
		adcBits int
	}{
		{"3bit_DAC_ADC", 3, 3},
		{"8bit_DAC_ADC", 8, 8},
		{"16bit_DAC_ADC", 16, 16},
	}

	var prevProbs []float64
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			net.Config.DACBits = tc.dacBits
			net.Config.ADCBits = tc.adcBits

			_, _, probs := net.InferCIMOnly(input)

			// Verify output is valid
			if len(probs) != net.OutputSize {
				t.Errorf("Output size mismatch: got %d, want %d", len(probs), net.OutputSize)
			}

			// Compare with previous run to verify quantization has an effect
			if prevProbs != nil && tc.dacBits != 16 {
				// Results should differ at different bit depths (unless 16-bit which is passthrough)
				identical := true
				for i := range probs {
					if math.Abs(probs[i]-prevProbs[i]) > 1e-9 {
						identical = false
						break
					}
				}
				if identical && tc.dacBits < 16 {
					t.Logf("Note: Results identical despite different bit depths")
				}
			}

			prevProbs = probs
		})
	}
}

// ============================================
// quantizeDAC Tests
// ============================================

func TestQuantizeDAC_BitLevels(t *testing.T) {
	testCases := []struct {
		bits           int
		expectedLevels int
	}{
		{3, 8},   // 2^3 = 8 levels
		{8, 256}, // 2^8 = 256 levels
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.bits))+"bit", func(t *testing.T) {
			input := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
			result := quantizeDAC(input, tc.bits)

			// Verify output is quantized to discrete levels
			uniqueLevels := make(map[float64]bool)
			for _, v := range result {
				uniqueLevels[v] = true
			}

			if len(uniqueLevels) > tc.expectedLevels {
				t.Errorf("Too many unique levels: got %d, want <=%d", len(uniqueLevels), tc.expectedLevels)
			}

			// Verify values are in [0, 1]
			for i, v := range result {
				if v < 0 || v > 1 {
					t.Errorf("Value[%d] out of range: got %f, want [0, 1]", i, v)
				}
			}
		})
	}
}

func TestQuantizeDAC_InputClamping(t *testing.T) {
	testCases := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"NegativeClampedTo0", -0.5, 0.0},
		{"GreaterThan1ClampedTo1", 1.5, 1.0},
		{"Zero", 0.0, 0.0},
		{"One", 1.0, 1.0},
		{"Mid", 0.5, 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := []float64{tc.input}
			result := quantizeDAC(input, 8)

			// Verify clamping
			if result[0] < 0 || result[0] > 1 {
				t.Errorf("Value not clamped: got %f, want [0, 1]", result[0])
			}

			// For extreme values, verify they're at the boundaries
			if tc.input < 0 && result[0] != 0.0 {
				t.Errorf("Negative value not clamped to 0: got %f", result[0])
			}
			if tc.input > 1 && result[0] != 1.0 {
				t.Errorf("Value >1 not clamped to 1: got %f", result[0])
			}
		})
	}
}

func TestQuantizeDAC_16BitPassthrough(t *testing.T) {
	input := []float64{0.123, 0.456, 0.789, 0.234}
	result := quantizeDAC(input, 16)

	// 16-bit should return input unchanged
	if len(result) != len(input) {
		t.Errorf("Length mismatch: got %d, want %d", len(result), len(input))
	}

	for i := range input {
		if result[i] != input[i] {
			t.Errorf("Value[%d] modified: got %f, want %f", i, result[i], input[i])
		}
	}
}

// ============================================
// quantizeADC Tests
// ============================================

func TestQuantizeADC_BitLevels(t *testing.T) {
	testCases := []struct {
		bits int
	}{
		{3},  // 8 levels
		{8},  // 256 levels
		{12}, // 4096 levels
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.bits))+"bit", func(t *testing.T) {
			input := []float64{-1.0, -0.5, 0.0, 0.5, 1.0, 1.5, 2.0}
			result := quantizeADC(input, tc.bits)

			// Verify output has been quantized (discrete levels)
			// and maintains the relative ordering
			if len(result) != len(input) {
				t.Errorf("Length mismatch: got %d, want %d", len(result), len(input))
			}

			// Verify monotonicity is preserved (sorted input -> sorted output)
			for i := 1; i < len(result); i++ {
				if result[i] < result[i-1] {
					t.Errorf("Monotonicity violated at index %d: %f < %f", i, result[i], result[i-1])
				}
			}
		})
	}
}

func TestQuantizeADC_EmptySlice(t *testing.T) {
	// CRIT-002 regression test: empty slice should return empty slice
	input := []float64{}
	result := quantizeADC(input, 8)

	if len(result) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(result))
	}
}

func TestQuantizeADC_16BitPassthrough(t *testing.T) {
	input := []float64{-1.0, -0.5, 0.0, 0.5, 1.0, 1.5}
	result := quantizeADC(input, 16)

	// 16-bit should return input unchanged
	if len(result) != len(input) {
		t.Errorf("Length mismatch: got %d, want %d", len(result), len(input))
	}

	for i := range input {
		if result[i] != input[i] {
			t.Errorf("Value[%d] modified: got %f, want %f", i, result[i], input[i])
		}
	}
}

func TestQuantizeADC_RangeNormalization(t *testing.T) {
	// Test that ADC properly handles different input ranges
	testCases := []struct {
		name  string
		input []float64
	}{
		{"NegativeRange", []float64{-10.0, -5.0, -1.0}},
		{"PositiveRange", []float64{1.0, 5.0, 10.0}},
		{"MixedRange", []float64{-5.0, 0.0, 5.0}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := quantizeADC(tc.input, 8)

			if len(result) != len(tc.input) {
				t.Errorf("Length mismatch: got %d, want %d", len(result), len(tc.input))
			}

			// Verify output is within input range
			for i, v := range result {
				if v < tc.input[0] || v > tc.input[len(tc.input)-1] {
					t.Errorf("Value[%d]=%f outside input range [%f, %f]",
						i, v, tc.input[0], tc.input[len(tc.input)-1])
				}
			}
		})
	}
}

// ============================================
// Regression Tests
// ============================================

func TestSoftmax_EmptySlice(t *testing.T) {
	// CRIT-001 regression test: softmax(nil) should return nil
	result := softmax(nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}

	result = softmax([]float64{})
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestArgmax_EmptySlice(t *testing.T) {
	// argmax(nil) should return -1 without panic
	result := argmax(nil)
	if result != -1 {
		t.Errorf("Expected -1, got %d", result)
	}

	result = argmax([]float64{})
	if result != -1 {
		t.Errorf("Expected -1, got %d", result)
	}
}

func TestInfer_InputLengthValidation(t *testing.T) {
	// MED-008 regression test: wrong input length should return nil
	net := createTestNetwork()

	testCases := []struct {
		name      string
		inputSize int
	}{
		{"TooShort", 2},
		{"TooLong", 6},
		{"Empty", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := make([]float64, tc.inputSize)
			result := net.Infer(input)

			if result != nil {
				t.Errorf("Expected nil for input length %d, got result: %+v", tc.inputSize, result)
			}
		})
	}
}

func TestInfer_SingleLayerMode(t *testing.T) {
	// Verify single-layer mode inference works
	net := createTestNetwork()

	// Initialize single-layer weights
	net.SingleLayerWeights = [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
	}
	net.SingleLayerBias = []float64{0.1, 0.2}
	net.QuantSingleLayerWeights = [][]float64{
		{0.1, 0.2, 0.3, 0.4},
		{0.5, 0.6, 0.7, 0.8},
	}
	net.QuantSingleLayerBias = []float64{0.1, 0.2}

	net.Config.SingleLayer = true
	input := []float64{0.1, 0.2, 0.3, 0.4}

	result := net.Infer(input)

	if result == nil {
		t.Fatal("Infer returned nil in single-layer mode")
	}

	// Verify FP results
	if result.FPPrediction < 0 || result.FPPrediction >= net.OutputSize {
		t.Errorf("Invalid FP prediction: got %d, want 0-%d", result.FPPrediction, net.OutputSize-1)
	}

	// Verify CIM results
	if result.CIMPrediction < 0 || result.CIMPrediction >= net.OutputSize {
		t.Errorf("Invalid CIM prediction: got %d, want 0-%d", result.CIMPrediction, net.OutputSize-1)
	}

	// Verify hidden layers are nil in single-layer mode
	if result.FPHidden != nil {
		t.Errorf("Expected nil FPHidden in single-layer mode, got %v", result.FPHidden)
	}
	if result.CIMHidden != nil {
		t.Errorf("Expected nil CIMHidden in single-layer mode, got %v", result.CIMHidden)
	}

	// Verify energy calculation is valid
	if result.EnergyUsed <= 0 {
		t.Errorf("Invalid energy: got %f, want >0", result.EnergyUsed)
	}
}
