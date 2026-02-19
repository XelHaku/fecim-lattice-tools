package neural

import (
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestQuantizationActivationOnly_M3_QUANT_03 validates activation-only quantization impact.
// Requirements:
// - Quantize activations to N-bit (2, 4, 8), keep weights at 8-bit
// - Measure accuracy degradation from activation quantization alone
// Evidence: exact accuracy percentages per activation bit width
func TestQuantizationActivationOnly_M3_QUANT_03(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping activation-only quantization test in short mode")
	}

	// Load MNIST test set
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Use 1000-image subset
	const testSubsetSize = 1000
	if len(images) < testSubsetSize {
		t.Fatalf("Expected at least %d test images, got %d", testSubsetSize, len(images))
	}
	images = images[:testSubsetSize]
	labels = labels[:testSubsetSize]

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Test activation bit widths: 2, 4, 8 (weights fixed at high precision)
	activationBits := []int{2, 4, 8}

	t.Logf("M3-QUANT-03: Activation-Only Quantization (weights fixed at 8-bit)")
	t.Logf("  Activation Bits | DAC Bits | ADC Bits | Accuracy")
	t.Logf("  ----------------|----------|----------|----------")

	for _, bits := range activationBits {
		// Keep weights at high precision (8-bit quantization = 256 levels)
		net.Config.NumLevels = 256 // 8-bit weights

		// Quantize activations via DAC (input) and ADC (output)
		net.Config.DACBits = bits // Input activation quantization
		net.Config.ADCBits = bits // Output activation quantization
		net.Config.NoiseLevel = 0.0

		net.RequantizeWeights()

		// Run inference
		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result == nil {
				t.Fatalf("Infer returned nil for image %d (activation_bits=%d)", i, bits)
			}
			if result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		t.Logf("  %15d | %8d | %8d | %7.2f%% (%d/%d)",
			bits, bits, bits, accuracy, correct, len(images))
	}

	t.Logf("M3-QUANT-03: Activation-only quantization characterization complete")
}

// TestQuantizationActivationOnly_InputOutput validates separate input/output quantization.
// This tests if DAC (input) vs ADC (output) quantization have different impacts.
func TestQuantizationActivationOnly_InputOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping input/output quantization test in short mode")
	}

	// Load MNIST test set
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Use 1000-image subset
	const testSubsetSize = 1000
	if len(images) < testSubsetSize {
		t.Fatalf("Expected at least %d test images, got %d", testSubsetSize, len(images))
	}
	images = images[:testSubsetSize]
	labels = labels[:testSubsetSize]

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Keep weights at 8-bit
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	t.Logf("M3-QUANT-03 Input/Output: Testing DAC vs ADC quantization sensitivity")
	t.Logf("  DAC Bits | ADC Bits | Accuracy  | Configuration")
	t.Logf("  ---------|----------|-----------|------------------")

	// Test configurations: vary DAC and ADC independently
	configs := []struct {
		dacBits int
		adcBits int
		desc    string
	}{
		{8, 8, "Baseline"},
		{4, 8, "DAC quantized"},
		{8, 4, "ADC quantized"},
		{4, 4, "Both quantized"},
		{2, 8, "DAC aggressive"},
		{8, 2, "ADC aggressive"},
		{2, 2, "Both aggressive"},
	}

	for _, cfg := range configs {
		net.Config.DACBits = cfg.dacBits
		net.Config.ADCBits = cfg.adcBits
		net.RequantizeWeights()

		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result == nil {
				t.Fatalf("Infer returned nil for image %d", i)
			}
			if result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		t.Logf("  %8d | %8d | %7.2f%% | %s",
			cfg.dacBits, cfg.adcBits, accuracy, cfg.desc)
	}

	t.Logf("M3-QUANT-03 Input/Output: Characterization complete")
}

// TestQuantizationActivationOnly_Sensitivity analyzes activation quantization impact.
func TestQuantizationActivationOnly_Sensitivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping activation quantization sensitivity test in short mode")
	}

	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const testSubsetSize = 1000
	if len(images) < testSubsetSize {
		t.Fatalf("Expected at least %d test images, got %d", testSubsetSize, len(images))
	}
	images = images[:testSubsetSize]
	labels = labels[:testSubsetSize]

	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Measure baseline with 8-bit activations
	net.Config.NumLevels = 256 // 8-bit weights
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	baselineCorrect := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result.CIMPrediction == labels[i] {
			baselineCorrect++
		}
	}
	baselineAccuracy := 100.0 * float64(baselineCorrect) / float64(len(images))

	// Test 4-bit activation quantization
	net.Config.DACBits = 4
	net.Config.ADCBits = 4
	net.RequantizeWeights()

	quantCorrect := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result.CIMPrediction == labels[i] {
			quantCorrect++
		}
	}
	quantAccuracy := 100.0 * float64(quantCorrect) / float64(len(images))

	degradation := baselineAccuracy - quantAccuracy

	t.Logf("M3-QUANT-03 Sensitivity Analysis:")
	t.Logf("  Baseline (8-bit activations): %.2f%% (%d/%d)",
		baselineAccuracy, baselineCorrect, len(images))
	t.Logf("  4-bit activation quantization: %.2f%% (%d/%d)",
		quantAccuracy, quantCorrect, len(images))
	t.Logf("  Accuracy degradation:          %.2f%%", degradation)

	if degradation > 30.0 {
		t.Logf("  WARNING: High sensitivity to activation quantization (>30%% drop)")
	} else if degradation > 15.0 {
		t.Logf("  MODERATE: Medium sensitivity to activation quantization (15-30%% drop)")
	} else {
		t.Logf("  GOOD: Low sensitivity to activation quantization (<15%% drop)")
	}
}

// TestQuantizationActivationOnly_CompareWeightActivation compares weight vs activation impact.
func TestQuantizationActivationOnly_CompareWeightActivation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping weight vs activation comparison test in short mode")
	}

	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const testSubsetSize = 1000
	if len(images) < testSubsetSize {
		t.Fatalf("Expected at least %d test images, got %d", testSubsetSize, len(images))
	}
	images = images[:testSubsetSize]
	labels = labels[:testSubsetSize]

	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	t.Logf("M3-QUANT-03 Comparison: Weight vs Activation Quantization Impact")
	t.Logf("  Configuration              | Accuracy  | Degradation")
	t.Logf("  ---------------------------|-----------|------------")

	// Baseline: 8-bit everywhere
	net.Config.NumLevels = 256
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0
	net.RequantizeWeights()

	baselineCorrect := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result.CIMPrediction == labels[i] {
			baselineCorrect++
		}
	}
	baselineAccuracy := 100.0 * float64(baselineCorrect) / float64(len(images))
	t.Logf("  8-bit weights + activations | %7.2f%% |    --", baselineAccuracy)

	// 4-bit weights only
	net.Config.NumLevels = 16 // 4-bit
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.RequantizeWeights()

	weightCorrect := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result.CIMPrediction == labels[i] {
			weightCorrect++
		}
	}
	weightAccuracy := 100.0 * float64(weightCorrect) / float64(len(images))
	weightDegradation := baselineAccuracy - weightAccuracy
	t.Logf("  4-bit weights, 8-bit acts   | %7.2f%% | %7.2f%%",
		weightAccuracy, weightDegradation)

	// 4-bit activations only
	net.Config.NumLevels = 256 // 8-bit weights
	net.Config.DACBits = 4
	net.Config.ADCBits = 4
	net.RequantizeWeights()

	actCorrect := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result.CIMPrediction == labels[i] {
			actCorrect++
		}
	}
	actAccuracy := 100.0 * float64(actCorrect) / float64(len(images))
	actDegradation := baselineAccuracy - actAccuracy
	t.Logf("  8-bit weights, 4-bit acts   | %7.2f%% | %7.2f%%",
		actAccuracy, actDegradation)

	// Both 4-bit
	net.Config.NumLevels = 16
	net.Config.DACBits = 4
	net.Config.ADCBits = 4
	net.RequantizeWeights()

	bothCorrect := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result.CIMPrediction == labels[i] {
			bothCorrect++
		}
	}
	bothAccuracy := 100.0 * float64(bothCorrect) / float64(len(images))
	bothDegradation := baselineAccuracy - bothAccuracy
	t.Logf("  4-bit weights + activations | %7.2f%% | %7.2f%%",
		bothAccuracy, bothDegradation)

	t.Logf("")
	if weightDegradation > actDegradation {
		t.Logf("  Result: Weights more sensitive than activations (%.2f%% vs %.2f%%)",
			weightDegradation, actDegradation)
	} else if actDegradation > weightDegradation {
		t.Logf("  Result: Activations more sensitive than weights (%.2f%% vs %.2f%%)",
			actDegradation, weightDegradation)
	} else {
		t.Logf("  Result: Similar sensitivity (%.2f%% vs %.2f%%)",
			weightDegradation, actDegradation)
	}
}
