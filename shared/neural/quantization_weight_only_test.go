package neural

import (
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestQuantizationWeightOnly_M3_QUANT_02 validates weight-only quantization impact.
// Requirements:
// - Quantize weights to N-bit (2, 4, 8), keep activations at 8-bit
// - Measure accuracy degradation from weight quantization alone
// Evidence: exact accuracy percentages per weight bit width
func TestQuantizationWeightOnly_M3_QUANT_02(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping weight-only quantization test in short mode")
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

	// Test weight bit widths: 2, 4, 8 (activations fixed at 8-bit)
	weightBits := []int{2, 4, 8}

	t.Logf("M3-QUANT-02: Weight-Only Quantization (activations fixed at 8-bit)")
	t.Logf("  Weight Bits | Weight Levels | Accuracy")
	t.Logf("  ------------|---------------|----------")

	for _, bits := range weightBits {
		// Configure weight quantization
		weightLevels := 1 << bits // 2^bits levels for weights
		net.Config.NumLevels = weightLevels

		// Keep DAC/ADC at 8-bit (activations remain high precision)
		// DAC controls input quantization, ADC controls output quantization
		net.Config.DACBits = 8
		net.Config.ADCBits = 8
		net.Config.NoiseLevel = 0.0

		net.RequantizeWeights()

		// Run inference
		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result == nil {
				t.Fatalf("Infer returned nil for image %d (weight_bits=%d)", i, bits)
			}
			if result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		t.Logf("  %11d | %13d | %7.2f%% (%d/%d)",
			bits, weightLevels, accuracy, correct, len(images))
	}

	t.Logf("M3-QUANT-02: Weight-only quantization characterization complete")
}

// TestQuantizationWeightOnly_PerLayer validates per-layer weight quantization.
// This tests if different layers have different sensitivity to quantization.
func TestQuantizationWeightOnly_PerLayer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping per-layer weight quantization test in short mode")
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

	// Enable per-layer quantization
	net.Config.PerLayerQuant = true
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0

	t.Logf("M3-QUANT-02 Per-Layer: Testing layer quantization sensitivity")
	t.Logf("  Layer1 Bits | Layer2 Bits | Accuracy")
	t.Logf("  ------------|-------------|----------")

	// Test configurations: vary each layer independently
	configs := []struct {
		layer1Bits int
		layer2Bits int
	}{
		{8, 8}, // Baseline
		{4, 8}, // Layer 1 quantized
		{8, 4}, // Layer 2 quantized
		{4, 4}, // Both quantized
		{2, 8}, // Layer 1 aggressive
		{8, 2}, // Layer 2 aggressive
	}

	for _, cfg := range configs {
		net.Config.Layer1Levels = 1 << cfg.layer1Bits
		net.Config.Layer2Levels = 1 << cfg.layer2Bits
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
		t.Logf("  %11d | %11d | %7.2f%% (%d/%d)",
			cfg.layer1Bits, cfg.layer2Bits, accuracy, correct, len(images))
	}

	t.Logf("M3-QUANT-02 Per-Layer: Characterization complete")
}

// TestQuantizationWeightOnly_Sensitivity analyzes which layer is more sensitive.
func TestQuantizationWeightOnly_Sensitivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping weight quantization sensitivity test in short mode")
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

	// Measure baseline with 8-bit everywhere
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

	// Test 4-bit weight quantization
	net.Config.NumLevels = 16 // 4-bit
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

	t.Logf("M3-QUANT-02 Sensitivity Analysis:")
	t.Logf("  Baseline (8-bit weights):     %.2f%% (%d/%d)",
		baselineAccuracy, baselineCorrect, len(images))
	t.Logf("  4-bit weight quantization:    %.2f%% (%d/%d)",
		quantAccuracy, quantCorrect, len(images))
	t.Logf("  Accuracy degradation:         %.2f%%", degradation)

	if degradation > 30.0 {
		t.Logf("  WARNING: High sensitivity to weight quantization (>30%% drop)")
	} else if degradation > 15.0 {
		t.Logf("  MODERATE: Medium sensitivity to weight quantization (15-30%% drop)")
	} else {
		t.Logf("  GOOD: Low sensitivity to weight quantization (<15%% drop)")
	}
}
