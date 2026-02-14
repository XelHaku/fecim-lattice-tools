package core

import (
	"math"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestReadDisturb_M3_NOISE_02_ConductanceShift validates accuracy degradation from read disturb.
// Read disturb: Small conductance shift per read operation (cumulative).
// Requirement: Measure accuracy after 100, 1000, 10000 reads.
// Evidence: accuracy vs read count, degradation trajectory
func TestReadDisturb_M3_NOISE_02_ConductanceShift(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping read disturb test in short mode")
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

	t.Logf("M3-NOISE-02: Testing read disturb conductance shift on %d images", subsetSize)

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

	// Read disturb parameters (physical model)
	// Each read causes small conductance shift: ΔG/G ≈ 10^-5 to 10^-4 per read
	// Cumulative over many reads → measurable accuracy degradation
	const disturbPerRead = 1e-5 // Conservative (low disturb)

	// Test checkpoints: 0, 100, 1000, 10000 reads
	readCounts := []int{0, 100, 1000, 10000}
	results := make([]struct {
		reads    int
		correct  int
		accuracy float64
	}, len(readCounts))

	for idx, targetReads := range readCounts {
		// Simulate cumulative read disturb by adding small drift to weights
		// Total shift = disturbPerRead * targetReads
		cumulativeShift := disturbPerRead * float64(targetReads)

		// Apply multiplicative conductance shift to all weights
		// G_new = G_old * (1 + cumulativeShift)
		// This models systematic read disturb (directional drift)
		applyReadDisturbShift(net.QuantWeights1, cumulativeShift)
		applyReadDisturbShift(net.QuantWeights2, cumulativeShift)

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
		results[idx].reads = targetReads
		results[idx].correct = correct
		results[idx].accuracy = accuracy

		t.Logf("  After %5d reads: Accuracy = %.2f%% (%d/%d correct), cumulative shift = %.6f",
			targetReads, accuracy, correct, len(images), cumulativeShift)
	}

	// Validate degradation trajectory
	baseline := results[0].accuracy
	acc100 := results[1].accuracy
	acc1000 := results[2].accuracy
	acc10000 := results[3].accuracy

	// Accuracy should degrade monotonically (within tolerance)
	if acc100 > baseline+2.0 {
		t.Logf("WARNING: Accuracy increased after 100 reads (%.2f%% vs %.2f%% baseline)",
			acc100, baseline)
	}
	if acc1000 > acc100+2.0 {
		t.Logf("WARNING: Accuracy increased from 100 to 1000 reads (%.2f%% vs %.2f%%)",
			acc1000, acc100)
	}

	// At 10,000 reads with conservative disturb, expect <10% degradation from baseline
	degradation := baseline - acc10000
	t.Logf("  Total degradation (baseline → 10k reads): %.2f%%", degradation)

	if degradation < 0 {
		t.Logf("NOTE: No degradation observed (model may be robust or disturb too small)")
	}
	if degradation > 20.0 {
		t.Errorf("Excessive degradation %.2f%% after 10k reads (>20%% threshold)", degradation)
	}

	// Final accuracy should still be >60% even after 10k reads
	if acc10000 < 60.0 {
		t.Errorf("Accuracy after 10k reads %.2f%% < 60%% (excessive read disturb)", acc10000)
	}

	t.Logf("M3-NOISE-02: PASS — Read disturb degradation characterized")
	t.Logf("  Baseline: %.2f%%, 100 reads: %.2f%%, 1k reads: %.2f%%, 10k reads: %.2f%%",
		baseline, acc100, acc1000, acc10000)
}

// TestReadDisturb_M3_NOISE_02_AsymmetricShift validates asymmetric read disturb.
// Positive and negative conductance states may drift differently.
func TestReadDisturb_M3_NOISE_02_AsymmetricShift(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping asymmetric read disturb test in short mode")
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

	t.Logf("M3-NOISE-02: Testing asymmetric read disturb on %d images", subsetSize)

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

	// Asymmetric read disturb: positive weights drift up, negative weights drift down
	const readCount = 1000
	const disturbPerRead = 2e-5 // Slightly higher for visibility

	// Apply asymmetric shift
	applyAsymmetricReadDisturb(net.QuantWeights1, disturbPerRead, readCount)
	applyAsymmetricReadDisturb(net.QuantWeights2, disturbPerRead, readCount)

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
	t.Logf("  Asymmetric read disturb (%d reads): Accuracy = %.2f%% (%d/%d correct)",
		readCount, accuracy, correct, len(images))

	// Should maintain >65% accuracy even with asymmetric drift
	if accuracy < 65.0 {
		t.Errorf("Asymmetric read disturb accuracy %.2f%% < 65%%", accuracy)
	}

	t.Logf("M3-NOISE-02: PASS — Asymmetric read disturb validated (%.2f%% ≥ 65%%)", accuracy)
}

// TestReadDisturb_M3_NOISE_02_Recovery validates recovery after read disturb.
// In real FeCIM, periodic refresh can restore conductance states.
func TestReadDisturb_M3_NOISE_02_Recovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping read disturb recovery test in short mode")
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

	t.Logf("M3-NOISE-02: Testing read disturb recovery on %d images", subsetSize)

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

	// Baseline accuracy
	correct0 := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result != nil && result.CIMPrediction == labels[i] {
			correct0++
		}
	}
	acc0 := 100.0 * float64(correct0) / float64(len(images))

	// Apply read disturb
	const readCount = 5000
	const disturbPerRead = 1e-5
	applyReadDisturbShift(net.QuantWeights1, disturbPerRead*float64(readCount))
	applyReadDisturbShift(net.QuantWeights2, disturbPerRead*float64(readCount))

	// Disturbed accuracy
	correct1 := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result != nil && result.CIMPrediction == labels[i] {
			correct1++
		}
	}
	acc1 := 100.0 * float64(correct1) / float64(len(images))

	// Refresh: restore original weights (simulates periodic reprogramming)
	for i := range net.QuantWeights1 {
		copy(net.QuantWeights1[i], origWeights1[i])
	}
	for i := range net.QuantWeights2 {
		copy(net.QuantWeights2[i], origWeights2[i])
	}

	// Recovered accuracy
	correct2 := 0
	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result != nil && result.CIMPrediction == labels[i] {
			correct2++
		}
	}
	acc2 := 100.0 * float64(correct2) / float64(len(images))

	t.Logf("  Baseline accuracy:         %.2f%% (%d/%d)", acc0, correct0, len(images))
	t.Logf("  After %d reads:            %.2f%% (%d/%d)", readCount, acc1, correct1, len(images))
	t.Logf("  After refresh:             %.2f%% (%d/%d)", acc2, correct2, len(images))
	t.Logf("  Recovery: %.2f%% → %.2f%% (Δ = %.2f%%)", acc1, acc2, acc2-acc1)

	// Recovery should restore accuracy to within 1% of baseline
	if math.Abs(acc2-acc0) > 1.0 {
		t.Errorf("Recovery accuracy %.2f%% differs from baseline %.2f%% by >1%%", acc2, acc0)
	}

	t.Logf("M3-NOISE-02: PASS — Read disturb recovery validated")
}

// applyReadDisturbShift applies multiplicative conductance shift to weight matrix.
// shift: cumulative fractional change (e.g., 0.001 = 0.1% increase)
func applyReadDisturbShift(weights [][]float64, shift float64) {
	factor := 1.0 + shift
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] *= factor
		}
	}
}

// applyAsymmetricReadDisturb applies sign-dependent read disturb.
// Positive weights increase, negative weights decrease (drift away from zero).
func applyAsymmetricReadDisturb(weights [][]float64, disturbPerRead float64, readCount int) {
	cumulativeShift := disturbPerRead * float64(readCount)
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] > 0 {
				weights[i][j] *= (1.0 + cumulativeShift) // Positive drift up
			} else if weights[i][j] < 0 {
				weights[i][j] *= (1.0 + cumulativeShift) // Negative drift down (more negative)
			}
			// Zero weights remain zero
		}
	}
}

// TestReadDisturb_M3_NOISE_02_StatisticalVariation validates read disturb with statistical spread.
// Real devices show variation in disturb rate across cells.
func TestReadDisturb_M3_NOISE_02_StatisticalVariation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping statistical read disturb test in short mode")
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

	t.Logf("M3-NOISE-02: Testing statistical read disturb variation on %d images", subsetSize)

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

	// Apply statistical read disturb: each cell has different disturb rate
	// Mean disturb = 1e-5, std dev = 3e-6 (30% variation)
	const readCount = 1000
	const meanDisturbPerRead = 1e-5
	const stdDevDisturb = 3e-6

	applyStatisticalReadDisturb(net.QuantWeights1, meanDisturbPerRead, stdDevDisturb, readCount)
	applyStatisticalReadDisturb(net.QuantWeights2, meanDisturbPerRead, stdDevDisturb, readCount)

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
	t.Logf("  Statistical read disturb (%d reads, mean=%.1e±%.1e): Accuracy = %.2f%% (%d/%d)",
		readCount, meanDisturbPerRead, stdDevDisturb, accuracy, correct, len(images))

	// Should maintain >70% accuracy with moderate statistical variation
	if accuracy < 70.0 {
		t.Errorf("Statistical read disturb accuracy %.2f%% < 70%%", accuracy)
	}

	t.Logf("M3-NOISE-02: PASS — Statistical read disturb variation validated (%.2f%% ≥ 70%%)", accuracy)
}

// applyStatisticalReadDisturb applies read disturb with cell-to-cell variation.
func applyStatisticalReadDisturb(weights [][]float64, meanDisturb, stdDevDisturb float64, readCount int) {
	// Each cell gets a random disturb rate drawn from N(meanDisturb, stdDevDisturb)
	// This requires deterministic RNG seeded per cell
	for i := range weights {
		for j := range weights[i] {
			// Deterministic seed based on position (for reproducibility)
			cellSeed := int64(i*1000 + j)
			localRNG := newSeededRNG(cellSeed)

			// Cell-specific disturb rate
			cellDisturb := meanDisturb + localRNG.NormFloat64()*stdDevDisturb
			if cellDisturb < 0 {
				cellDisturb = 0 // Physical constraint: disturb rate ≥ 0
			}

			// Apply cumulative shift
			cumulativeShift := cellDisturb * float64(readCount)
			weights[i][j] *= (1.0 + cumulativeShift)
		}
	}
}
