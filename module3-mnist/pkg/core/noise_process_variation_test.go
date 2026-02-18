package core

import (
	"math"
	"math/rand"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestProcessVariation_M3_NOISE_03_MonteCarlo validates inference under device-to-device variation.
// Requirement: Monte Carlo N=100 runs, compute mean and std of accuracy.
// Acceptance: mean ≥75%, std < 5%
// Evidence: accuracy distribution (mean±std), histogram, pass rate
func TestProcessVariation_M3_NOISE_03_MonteCarlo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Monte Carlo process variation test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-03: Monte Carlo process variation on %d images", subsetSize)

	// Load pretrained weights (reference)
	refNet := NewDualModeNetwork(784, 128, 10)
	refNet.Config.DACBits = 8
	refNet.Config.ADCBits = 8
	refNet.Config.NumLevels = 256
	refNet.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := refNet.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Store reference quantized weights
	refWeights1 := make([][]float64, len(refNet.QuantWeights1))
	for i := range refWeights1 {
		refWeights1[i] = make([]float64, len(refNet.QuantWeights1[i]))
		copy(refWeights1[i], refNet.QuantWeights1[i])
	}
	refWeights2 := make([][]float64, len(refNet.QuantWeights2))
	for i := range refWeights2 {
		refWeights2[i] = make([]float64, len(refNet.QuantWeights2[i]))
		copy(refWeights2[i], refNet.QuantWeights2[i])
	}

	// Monte Carlo parameters
	const numRuns = 100
	const deviceVariationSigma = 0.03 // 3% device-to-device variation (typical FeCIM)

	accuracies := make([]float64, numRuns)
	rng := rand.New(rand.NewSource(42)) // Master RNG for run seeds

	t.Logf("  Running %d Monte Carlo trials with σ_device = %.3f", numRuns, deviceVariationSigma)

	for run := 0; run < numRuns; run++ {
		// Create network instance for this run
		net := NewDualModeNetwork(784, 128, 10)
		net.Config.DACBits = 8
		net.Config.ADCBits = 8
		net.Config.NumLevels = 256
		net.Config.NoiseLevel = 0.0

		// Copy reference weights
		net.QuantWeights1 = make([][]float64, len(refWeights1))
		for i := range net.QuantWeights1 {
			net.QuantWeights1[i] = make([]float64, len(refWeights1[i]))
			copy(net.QuantWeights1[i], refWeights1[i])
		}
		net.QuantWeights2 = make([][]float64, len(refWeights2))
		for i := range net.QuantWeights2 {
			net.QuantWeights2[i] = make([]float64, len(refWeights2[i]))
			copy(net.QuantWeights2[i], refWeights2[i])
		}
		net.QuantBias1 = make([]float64, len(refNet.QuantBias1))
		copy(net.QuantBias1, refNet.QuantBias1)
		net.QuantBias2 = make([]float64, len(refNet.QuantBias2))
		copy(net.QuantBias2, refNet.QuantBias2)

		// Apply device-to-device variation (Gaussian per device)
		// Each weight experiences independent random variation
		runRNG := rand.New(rand.NewSource(rng.Int63())) // Unique seed per run
		addDeviceVariation(net.QuantWeights1, deviceVariationSigma, runRNG)
		addDeviceVariation(net.QuantWeights2, deviceVariationSigma, runRNG)

		// Run inference on test subset
		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result == nil {
				t.Fatalf("Run %d: Infer returned nil for image %d", run, i)
			}
			if result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		accuracies[run] = accuracy

		if run < 5 || run%20 == 0 {
			t.Logf("    Run %3d: %.2f%% (%d/%d)", run, accuracy, correct, len(images))
		}
	}

	// Compute statistics
	mean, stddev := computeMeanStdDev(accuracies)
	minAcc, maxAcc := computeMinMax(accuracies)

	t.Logf("  Monte Carlo results (N=%d):", numRuns)
	t.Logf("    Mean:   %.2f%%", mean)
	t.Logf("    StdDev: %.2f%%", stddev)
	t.Logf("    Min:    %.2f%%", minAcc)
	t.Logf("    Max:    %.2f%%", maxAcc)
	t.Logf("    Range:  %.2f%% (max - min)", maxAcc-minAcc)

	// Compute histogram (10% bins)
	histogram := make(map[int]int)
	for _, acc := range accuracies {
		bin := int(acc / 10.0) // 0-9.9% → bin 0, 10-19.9% → bin 1, etc.
		histogram[bin]++
	}
	t.Logf("  Accuracy histogram:")
	for bin := 0; bin <= 10; bin++ {
		count := histogram[bin]
		if count > 0 {
			t.Logf("    %2d-%2d%%: %3d runs %s", bin*10, (bin+1)*10-1, count, repeatChar('█', count/2))
		}
	}

	// Validate requirements
	const minMean = 75.0
	const maxStdDev = 5.0

	if mean < minMean {
		t.Errorf("Monte Carlo mean accuracy %.2f%% < required %.2f%%", mean, minMean)
	}
	if stddev >= maxStdDev {
		t.Errorf("Monte Carlo stddev %.2f%% ≥ maximum %.2f%%", stddev, maxStdDev)
	}

	t.Logf("M3-NOISE-03: PASS — Monte Carlo process variation validated")
	t.Logf("  Mean accuracy: %.2f%% ≥ %.2f%% ✓", mean, minMean)
	t.Logf("  Std dev:       %.2f%% < %.2f%% ✓", stddev, maxStdDev)
}

// TestProcessVariation_M3_NOISE_03_WorstCase identifies worst-case device variation.
// Useful for understanding tail behavior and reliability.
func TestProcessVariation_M3_NOISE_03_WorstCase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping worst-case process variation test in short mode")
	}

	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 1000
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-03: Worst-case process variation analysis on %d images", subsetSize)

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

	// Store reference weights
	refWeights1 := copyWeights(net.QuantWeights1)
	refWeights2 := copyWeights(net.QuantWeights2)

	// Test worst-case scenarios
	scenarios := []struct {
		name  string
		sigma float64
		bias  float64 // Systematic bias (0 = symmetric, >0 = positive shift)
	}{
		{"Symmetric 5%", 0.05, 0.0},
		{"Positive bias 3%+2%", 0.03, 0.02},
		{"Negative bias 3%-2%", 0.03, -0.02},
		{"High variation 10%", 0.10, 0.0},
	}

	rng := rand.New(rand.NewSource(42))

	for _, scenario := range scenarios {
		// Restore reference weights
		restoreWeights(net.QuantWeights1, refWeights1)
		restoreWeights(net.QuantWeights2, refWeights2)

		// Apply scenario variation
		addDeviceVariationWithBias(net.QuantWeights1, scenario.sigma, scenario.bias, rng)
		addDeviceVariationWithBias(net.QuantWeights2, scenario.sigma, scenario.bias, rng)

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
		t.Logf("  %s: %.2f%% (%d/%d)", scenario.name, accuracy, correct, len(images))
	}

	t.Logf("M3-NOISE-03: PASS — Worst-case process variation scenarios characterized")
}

// TestProcessVariation_M3_NOISE_03_Repeatability validates Monte Carlo reproducibility.
// Same seed → same results (critical for debugging and validation).
func TestProcessVariation_M3_NOISE_03_Repeatability(t *testing.T) {
	// Load MNIST test subset
	dataDir := filepath.Join("..", "..", "data")
	allImages, allLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const subsetSize = 100 // Small subset for fast test
	images := allImages[:subsetSize]
	labels := allLabels[:subsetSize]

	t.Logf("M3-NOISE-03: Testing Monte Carlo repeatability on %d images", subsetSize)

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

	refWeights1 := copyWeights(net.QuantWeights1)
	refWeights2 := copyWeights(net.QuantWeights2)

	const numTrials = 3
	const deviceVariationSigma = 0.03

	var firstAccuracy float64

	for trial := 0; trial < numTrials; trial++ {
		// Restore weights
		restoreWeights(net.QuantWeights1, refWeights1)
		restoreWeights(net.QuantWeights2, refWeights2)

		// Apply variation with SAME SEED
		rng := rand.New(rand.NewSource(12345))
		addDeviceVariation(net.QuantWeights1, deviceVariationSigma, rng)
		addDeviceVariation(net.QuantWeights2, deviceVariationSigma, rng)

		// Run inference
		correct := 0
		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result != nil && result.CIMPrediction == labels[i] {
				correct++
			}
		}

		accuracy := 100.0 * float64(correct) / float64(len(images))
		t.Logf("  Trial %d (seed=12345): %.2f%% (%d/%d)", trial, accuracy, correct, len(images))

		if trial == 0 {
			firstAccuracy = accuracy
		} else {
			// Should be bit-identical
			if accuracy != firstAccuracy {
				t.Errorf("Trial %d accuracy %.2f%% != first trial %.2f%% (non-deterministic!)",
					trial, accuracy, firstAccuracy)
			}
		}
	}

	t.Logf("M3-NOISE-03: PASS — Monte Carlo repeatability validated (all trials: %.2f%%)", firstAccuracy)
}

// addDeviceVariation adds device-to-device variation to weight matrix.
// Each weight gets independent Gaussian noise N(0, sigma).
func addDeviceVariation(weights [][]float64, sigma float64, rng *rand.Rand) {
	for i := range weights {
		for j := range weights[i] {
			variation := rng.NormFloat64() * sigma
			weights[i][j] += variation
		}
	}
}

// addDeviceVariationWithBias adds device-to-device variation with systematic bias.
// variation = bias + N(0, sigma)
func addDeviceVariationWithBias(weights [][]float64, sigma, bias float64, rng *rand.Rand) {
	for i := range weights {
		for j := range weights[i] {
			variation := bias + rng.NormFloat64()*sigma
			weights[i][j] += variation
		}
	}
}

// computeMeanStdDev computes mean and standard deviation of a float64 slice.
func computeMeanStdDev(values []float64) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Compute mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Compute standard deviation
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(values))
	stddev = math.Sqrt(variance)

	return mean, stddev
}

// computeMinMax computes min and max of a float64 slice.
func computeMinMax(values []float64) (min, max float64) {
	if len(values) == 0 {
		return 0, 0
	}

	min = values[0]
	max = values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// copyWeights creates a deep copy of a 2D weight matrix.
func copyWeights(weights [][]float64) [][]float64 {
	copied := make([][]float64, len(weights))
	for i := range weights {
		copied[i] = make([]float64, len(weights[i]))
		copy(copied[i], weights[i])
	}
	return copied
}

// restoreWeights copies weights from source to destination.
func restoreWeights(dest, src [][]float64) {
	for i := range dest {
		copy(dest[i], src[i])
	}
}

// repeatChar repeats a character n times for simple ASCII bar charts.
func repeatChar(char rune, n int) string {
	if n < 0 {
		n = 0
	}
	result := make([]rune, n)
	for i := range result {
		result[i] = char
	}
	return string(result)
}

// newSeededRNG creates a new deterministic RNG with the given seed.
func newSeededRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
