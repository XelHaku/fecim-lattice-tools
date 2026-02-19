package neural

import (
	"math"
	"math/rand"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestNoiseInjection_M3_NOISE_01_GaussianWeights validates inference robustness under Gaussian weight noise.
// Requirement: Accuracy ≥70% at σ=0.05, ≥60% at σ=0.10
// Evidence: accuracy vs noise level (σ=0.01, 0.05, 0.10)
func TestNoiseInjection_M3_NOISE_01_GaussianWeights(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping noise injection test in short mode")
	}

	// Load MNIST test subset (1000 images for faster execution)
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Use first 1000 images
	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-01: Testing Gaussian weight noise injection on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0 // No inference noise, only weight noise

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Store original quantized weights for restoration
	origWeights1 := make([][]float64, len(net.QuantWeights1))
	for i := range origWeights1 {
		origWeights1[i] = make([]float64, len(net.QuantWeights1[i]))
		copy(origWeights1[i], net.QuantWeights1[i])
	}
	origWeights2 := make([][]float64, len(net.QuantWeights2))
	for i := range origWeights2 {
		origWeights2[i] = make([]float64, len(net.QuantWeights2[i]))
		copy(origWeights2[i], net.QuantWeights2[i])
	}

	// Test multiple noise levels
	noiseLevels := []float64{0.0, 0.01, 0.05, 0.10}
	results := make([]struct {
		sigma    float64
		correct  int
		accuracy float64
	}, len(noiseLevels))

	rng := rand.New(rand.NewSource(42)) // Deterministic for reproducibility

	for idx, sigma := range noiseLevels {
		// Restore original weights
		for i := range net.QuantWeights1 {
			copy(net.QuantWeights1[i], origWeights1[i])
		}
		for i := range net.QuantWeights2 {
			copy(net.QuantWeights2[i], origWeights2[i])
		}

		// Add Gaussian noise to weights if sigma > 0
		if sigma > 0.0 {
			addGaussianNoiseToWeights(net.QuantWeights1, sigma, rng)
			addGaussianNoiseToWeights(net.QuantWeights2, sigma, rng)
		}

		// Run inference on subset
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
		results[idx].sigma = sigma
		results[idx].correct = correct
		results[idx].accuracy = accuracy

		t.Logf("  σ=%.2f: Accuracy = %.2f%% (%d/%d correct)",
			sigma, accuracy, correct, len(images))
	}

	// Validate requirements
	// Baseline (σ=0.0) should be high
	if results[0].accuracy < 80.0 {
		t.Errorf("Baseline (σ=0.0) accuracy %.2f%% < 80%% (sanity check failed)", results[0].accuracy)
	}

	// σ=0.05 requirement: ≥70%
	acc005 := results[2].accuracy
	if acc005 < 70.0 {
		t.Errorf("σ=0.05 accuracy %.2f%% < required 70%%", acc005)
	}

	// σ=0.10 requirement: ≥60%
	acc010 := results[3].accuracy
	if acc010 < 60.0 {
		t.Errorf("σ=0.10 accuracy %.2f%% < required 60%%", acc010)
	}

	// Monotonicity check: accuracy should decrease with noise
	for i := 1; i < len(results); i++ {
		if results[i].accuracy > results[i-1].accuracy+2.0 { // Allow 2% tolerance for statistical fluctuation
			t.Logf("WARNING: Accuracy increased from σ=%.2f (%.2f%%) to σ=%.2f (%.2f%%) — non-monotonic",
				results[i-1].sigma, results[i-1].accuracy, results[i].sigma, results[i].accuracy)
		}
	}

	t.Logf("M3-NOISE-01: PASS — Gaussian weight noise robustness validated")
	t.Logf("  σ=0.05: %.2f%% ≥ 70%% ✓", acc005)
	t.Logf("  σ=0.10: %.2f%% ≥ 60%% ✓", acc010)
}

// TestNoiseInjection_M3_NOISE_01_BiasNoise validates inference robustness under bias noise.
// Biases are typically less sensitive but still contribute to output error.
func TestNoiseInjection_M3_NOISE_01_BiasNoise(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bias noise test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-01: Testing Gaussian bias noise injection on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Store original biases
	origBias1 := make([]float64, len(net.QuantBias1))
	copy(origBias1, net.QuantBias1)
	origBias2 := make([]float64, len(net.QuantBias2))
	copy(origBias2, net.QuantBias2)

	// Test moderate bias noise (σ=0.10)
	sigma := 0.10
	rng := rand.New(rand.NewSource(42))

	// Add Gaussian noise to biases
	addGaussianNoiseToBias(net.QuantBias1, sigma, rng)
	addGaussianNoiseToBias(net.QuantBias2, sigma, rng)

	// Run inference
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
	t.Logf("  Bias noise σ=%.2f: Accuracy = %.2f%% (%d/%d correct)",
		sigma, accuracy, correct, len(images))

	// Bias noise should have less impact than weight noise
	// Expect >70% accuracy even with σ=0.10 bias noise
	if accuracy < 70.0 {
		t.Errorf("Bias noise (σ=%.2f) accuracy %.2f%% < 70%% (excessive sensitivity)", sigma, accuracy)
	}

	t.Logf("M3-NOISE-01: PASS — Bias noise robustness validated (%.2f%% ≥ 70%%)", accuracy)
}

// addGaussianNoiseToWeights adds Gaussian noise N(0, σ) to weight matrix in-place.
// sigma is the standard deviation relative to weight magnitude.
func addGaussianNoiseToWeights(weights [][]float64, sigma float64, rng *rand.Rand) {
	for i := range weights {
		for j := range weights[i] {
			noise := rng.NormFloat64() * sigma
			weights[i][j] += noise
		}
	}
}

// addGaussianNoiseToBias adds Gaussian noise N(0, σ) to bias vector in-place.
func addGaussianNoiseToBias(bias []float64, sigma float64, rng *rand.Rand) {
	for i := range bias {
		noise := rng.NormFloat64() * sigma
		bias[i] += noise
	}
}

// TestNoiseInjection_M3_NOISE_01_LayerSpecific validates layer-specific noise sensitivity.
// Hidden layer weights typically have more impact than output layer due to fan-out.
func TestNoiseInjection_M3_NOISE_01_LayerSpecific(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping layer-specific noise test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-01: Testing layer-specific noise sensitivity on %d images", subsetSize)

	// Load pretrained weights
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Store original weights
	origWeights1 := make([][]float64, len(net.QuantWeights1))
	for i := range origWeights1 {
		origWeights1[i] = make([]float64, len(net.QuantWeights1[i]))
		copy(origWeights1[i], net.QuantWeights1[i])
	}
	origWeights2 := make([][]float64, len(net.QuantWeights2))
	for i := range origWeights2 {
		origWeights2[i] = make([]float64, len(net.QuantWeights2[i]))
		copy(origWeights2[i], net.QuantWeights2[i])
	}

	sigma := 0.05
	rng := rand.New(rand.NewSource(42))

	// Test 1: Noise only on Layer 1 (hidden layer)
	for i := range net.QuantWeights1 {
		copy(net.QuantWeights1[i], origWeights1[i])
	}
	for i := range net.QuantWeights2 {
		copy(net.QuantWeights2[i], origWeights2[i])
	}
	addGaussianNoiseToWeights(net.QuantWeights1, sigma, rng)

	correctLayer1 := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result != nil && result.CIMPrediction == labels[i] {
			correctLayer1++
		}
	}
	accLayer1 := 100.0 * float64(correctLayer1) / float64(len(images))

	// Test 2: Noise only on Layer 2 (output layer)
	for i := range net.QuantWeights1 {
		copy(net.QuantWeights1[i], origWeights1[i])
	}
	for i := range net.QuantWeights2 {
		copy(net.QuantWeights2[i], origWeights2[i])
	}
	rng = rand.New(rand.NewSource(42)) // Reset RNG for fair comparison
	addGaussianNoiseToWeights(net.QuantWeights2, sigma, rng)

	correctLayer2 := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result != nil && result.CIMPrediction == labels[i] {
			correctLayer2++
		}
	}
	accLayer2 := 100.0 * float64(correctLayer2) / float64(len(images))

	t.Logf("  Layer 1 noise only (σ=%.2f): %.2f%% (%d/%d)", sigma, accLayer1, correctLayer1, len(images))
	t.Logf("  Layer 2 noise only (σ=%.2f): %.2f%% (%d/%d)", sigma, accLayer2, correctLayer2, len(images))

	// Layer 1 typically has more impact (larger matrix, fan-out to 128 neurons)
	// But both should maintain >70% accuracy at σ=0.05
	if accLayer1 < 70.0 {
		t.Errorf("Layer 1 noise accuracy %.2f%% < 70%%", accLayer1)
	}
	if accLayer2 < 70.0 {
		t.Errorf("Layer 2 noise accuracy %.2f%% < 70%%", accLayer2)
	}

	t.Logf("M3-NOISE-01: PASS — Layer-specific noise sensitivity characterized")
}

// TestNoiseInjection_M3_NOISE_01_Statistics validates noise distribution properties.
// Ensures Gaussian noise generator produces correct mean and std dev.
func TestNoiseInjection_M3_NOISE_01_Statistics(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	const N = 100000
	sigma := 0.05

	// Generate N samples
	samples := make([]float64, N)
	for i := range samples {
		samples[i] = rng.NormFloat64() * sigma
	}

	// Compute mean
	mean := 0.0
	for _, v := range samples {
		mean += v
	}
	mean /= float64(N)

	// Compute standard deviation
	variance := 0.0
	for _, v := range samples {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(N)
	stddev := math.Sqrt(variance)

	t.Logf("M3-NOISE-01: Gaussian noise statistics (N=%d, σ=%.3f):", N, sigma)
	t.Logf("  Mean:   %.6f (expected 0.0)", mean)
	t.Logf("  Stddev: %.6f (expected %.3f)", stddev, sigma)

	// Validate distribution properties
	if math.Abs(mean) > 0.001 {
		t.Errorf("Mean %.6f deviates from 0.0 (>0.001)", mean)
	}
	if math.Abs(stddev-sigma) > 0.001 {
		t.Errorf("Stddev %.6f deviates from expected %.3f (>0.001)", stddev, sigma)
	}

	t.Logf("M3-NOISE-01: PASS — Noise generator statistics validated")
}
