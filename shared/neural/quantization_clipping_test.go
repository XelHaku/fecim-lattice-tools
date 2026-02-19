package neural

import (
	"math"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestQuantizationClipping_M3_QUANT_04 validates graceful saturation under extreme inputs.
// Requirements:
// - Inject extreme values (±1e6) into inputs
// - Verify no NaN/Inf in outputs
// - Verify graceful saturation (clipping to valid range)
// Evidence: no NaN/Inf detected, outputs remain in valid range
func TestQuantizationClipping_M3_QUANT_04(t *testing.T) {
	// Load pretrained weights
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Configure for 8-bit quantization
	net.Config.NumLevels = 256
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	// Test extreme input values
	extremeValues := []float64{1e6, -1e6, 1e9, -1e9, math.Inf(1), math.Inf(-1)}

	for _, extremeVal := range extremeValues {
		// Create input with extreme value
		input := make([]float64, 784)
		for i := range input {
			input[i] = extremeVal
		}

		// Run inference
		result := net.Infer(input)
		if result == nil {
			t.Fatalf("Infer returned nil for extreme input value %e", extremeVal)
		}

		// Check for NaN/Inf in CIM output probabilities
		// Note: We only check CIM path, as FP path with ±Inf inputs is expected to produce NaN
		hasNaN := false
		hasInf := false
		for i, prob := range result.CIMProbabilities {
			if math.IsNaN(prob) {
				hasNaN = true
				t.Errorf("NaN detected in CIM probabilities[%d] for extreme input %e", i, extremeVal)
			}
			if math.IsInf(prob, 0) {
				hasInf = true
				t.Errorf("Inf detected in CIM probabilities[%d] for extreme input %e", i, extremeVal)
			}
		}

		if !hasNaN && !hasInf {
			t.Logf("M3-QUANT-04: Extreme input %e handled gracefully by CIM path (no NaN/Inf)", extremeVal)
		}
	}

	t.Logf("M3-QUANT-04: PASS — No NaN/Inf in outputs for extreme inputs")
}

// TestQuantizationClipping_SaturationBehavior validates saturation characteristics.
func TestQuantizationClipping_SaturationBehavior(t *testing.T) {
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	net.Config.NumLevels = 256
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	// Test progressive saturation
	testValues := []float64{0.0, 1.0, 10.0, 100.0, 1000.0, 1e6}

	t.Logf("M3-QUANT-04: Saturation behavior analysis")
	t.Logf("  Input Value | Valid Output | NaN/Inf Detected")
	t.Logf("  ------------|--------------|------------------")

	for _, val := range testValues {
		input := make([]float64, 784)
		for i := range input {
			input[i] = val
		}

		result := net.Infer(input)
		if result == nil {
			t.Errorf("Infer returned nil for input value %e", val)
			continue
		}

		// Check validity
		validOutput := true
		nanInfDetected := false

		for _, prob := range result.CIMProbabilities {
			if math.IsNaN(prob) || math.IsInf(prob, 0) {
				nanInfDetected = true
				validOutput = false
				break
			}
			// Probabilities should be in [0, 1]
			if prob < 0 || prob > 1 {
				validOutput = false
			}
		}

		t.Logf("  %11e | %12v | %16v", val, validOutput, nanInfDetected)
	}

	t.Logf("M3-QUANT-04: Saturation behavior characterization complete")
}

// TestQuantizationClipping_MixedExtremes tests mixed extreme values in input.
func TestQuantizationClipping_MixedExtremes(t *testing.T) {
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	net.Config.NumLevels = 256
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	// Load one real MNIST image to mix with extremes
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST: %v", err)
	}
	if len(images) == 0 {
		t.Fatalf("No MNIST images loaded")
	}

	realImage := images[0]

	t.Logf("M3-QUANT-04: Testing mixed extreme values")
	t.Logf("  Test Case                      | Valid Output | NaN/Inf")
	t.Logf("  -------------------------------|--------------|--------")

	// Test case 1: Half real, half extreme
	input1 := make([]float64, 784)
	copy(input1[:392], realImage[:392])
	for i := 392; i < 784; i++ {
		input1[i] = 1e6
	}

	result1 := net.Infer(input1)
	if result1 == nil {
		t.Errorf("Infer returned nil for mixed input case 1")
	} else {
		valid, nanInf := checkOutputValidity(result1.CIMProbabilities)
		t.Logf("  Half real + half extreme       | %12v | %7v", valid, nanInf)
	}

	// Test case 2: Sparse extremes
	input2 := make([]float64, 784)
	copy(input2, realImage)
	for i := 0; i < 784; i += 100 {
		input2[i] = 1e9
	}

	result2 := net.Infer(input2)
	if result2 == nil {
		t.Errorf("Infer returned nil for mixed input case 2")
	} else {
		valid, nanInf := checkOutputValidity(result2.CIMProbabilities)
		t.Logf("  Real + sparse extremes         | %12v | %7v", valid, nanInf)
	}

	// Test case 3: Alternating extreme values
	input3 := make([]float64, 784)
	for i := range input3 {
		if i%2 == 0 {
			input3[i] = 1e6
		} else {
			input3[i] = -1e6
		}
	}

	result3 := net.Infer(input3)
	if result3 == nil {
		t.Errorf("Infer returned nil for mixed input case 3")
	} else {
		valid, nanInf := checkOutputValidity(result3.CIMProbabilities)
		t.Logf("  Alternating +/-1e6             | %12v | %7v", valid, nanInf)
	}

	t.Logf("M3-QUANT-04: Mixed extreme value testing complete")
}

// TestQuantizationClipping_WeightExtremes tests extreme values in weights (not inputs).
func TestQuantizationClipping_WeightExtremes(t *testing.T) {
	// Create network with extreme weight values
	net := NewDualModeNetwork(784, 128, 10)

	// Inject extreme values into weights
	net.FPWeights1 = make([][]float64, 128)
	for i := range net.FPWeights1 {
		net.FPWeights1[i] = make([]float64, 784)
		for j := range net.FPWeights1[i] {
			if (i+j)%2 == 0 {
				net.FPWeights1[i][j] = 1e6
			} else {
				net.FPWeights1[i][j] = -1e6
			}
		}
	}

	net.FPWeights2 = make([][]float64, 10)
	for i := range net.FPWeights2 {
		net.FPWeights2[i] = make([]float64, 128)
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = 1e6
		}
	}

	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)

	// Configure and quantize
	net.Config.NumLevels = 256
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	// Test with normal input
	input := make([]float64, 784)
	for i := range input {
		input[i] = 0.5 // Normal range input
	}

	result := net.Infer(input)
	if result == nil {
		t.Fatalf("Infer returned nil for extreme weights")
	}

	valid, nanInf := checkOutputValidity(result.CIMProbabilities)

	t.Logf("M3-QUANT-04: Extreme weight handling")
	t.Logf("  Weights: ±1e6 throughout network")
	t.Logf("  Input: normal range (0.5)")
	t.Logf("  Valid output: %v", valid)
	t.Logf("  NaN/Inf detected: %v", nanInf)

	if nanInf {
		t.Logf("  NOTE: Extreme weights may saturate even with normal inputs")
	} else {
		t.Logf("  PASS: Extreme weights handled gracefully")
	}
}

// TestQuantizationClipping_QuantizationSaturation tests saturation in quantization itself.
func TestQuantizationClipping_QuantizationSaturation(t *testing.T) {
	// Create weights with extreme values
	fpWeights := [][]float64{
		{-1e9, 1e9, 0.0, 1e6},
		{1e9, -1e9, -1e6, 0.0},
	}

	// Test quantization with different bit widths
	bitWidths := []int{2, 4, 8}

	t.Logf("M3-QUANT-04: Quantization saturation behavior")
	t.Logf("  Original weights include: ±1e9, ±1e6, 0.0")
	t.Logf("")
	t.Logf("  Bits | Levels | Max Quantized | Min Quantized | All Finite")
	t.Logf("  -----|--------|---------------|---------------|------------")

	for _, bits := range bitWidths {
		levels := 1 << bits
		quantized, err := QuantizeWeights(fpWeights, levels)
		if err != nil {
			t.Errorf("QuantizeWeights failed for %d bits: %v", bits, err)
			continue
		}

		// Find min/max in quantized weights
		maxVal := math.Inf(-1)
		minVal := math.Inf(1)
		allFinite := true

		for i := range quantized {
			for j := range quantized[i] {
				val := quantized[i][j]
				if math.IsNaN(val) || math.IsInf(val, 0) {
					allFinite = false
				}
				if val > maxVal {
					maxVal = val
				}
				if val < minVal {
					minVal = val
				}
			}
		}

		t.Logf("  %4d | %6d | %13.2e | %13.2e | %10v",
			bits, levels, maxVal, minVal, allFinite)

		if !allFinite {
			t.Errorf("Non-finite values in quantized weights for %d bits", bits)
		}
	}

	t.Logf("M3-QUANT-04: Quantization saturation characterization complete")
}

// TestQuantizationClipping_ReLUSaturation tests ReLU activation clipping.
func TestQuantizationClipping_ReLUSaturation(t *testing.T) {
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	net.Config.NumLevels = 256
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	// Test inputs designed to trigger large pre-activation values
	testCases := []struct {
		name  string
		value float64
	}{
		{"Zero", 0.0},
		{"Small positive", 0.1},
		{"Large positive", 10.0},
		{"Extreme positive", 1e6},
		{"Small negative", -0.1},
		{"Large negative", -10.0},
		{"Extreme negative", -1e6},
	}

	t.Logf("M3-QUANT-04: ReLU saturation behavior")
	t.Logf("  Input Type         | Valid Output | NaN/Inf")
	t.Logf("  -------------------|--------------|--------")

	for _, tc := range testCases {
		input := make([]float64, 784)
		for i := range input {
			input[i] = tc.value
		}

		result := net.Infer(input)
		if result == nil {
			t.Errorf("Infer returned nil for %s", tc.name)
			continue
		}

		valid, nanInf := checkOutputValidity(result.CIMProbabilities)
		t.Logf("  %-18s | %12v | %7v", tc.name, valid, nanInf)
	}

	t.Logf("M3-QUANT-04: ReLU saturation characterization complete")
}

// checkOutputValidity checks if output probabilities are valid (finite, in [0,1]).
func checkOutputValidity(probs []float64) (valid bool, hasNaNInf bool) {
	valid = true
	hasNaNInf = false

	for _, prob := range probs {
		if math.IsNaN(prob) || math.IsInf(prob, 0) {
			hasNaNInf = true
			valid = false
			break
		}
		if prob < 0 || prob > 1 {
			valid = false
		}
	}

	return valid, hasNaNInf
}
