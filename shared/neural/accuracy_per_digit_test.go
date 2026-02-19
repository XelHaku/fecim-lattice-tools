package neural

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// perDigitStats holds accuracy metrics for a single digit class.
type perDigitStats struct {
	Digit      int
	TotalCount int
	Correct    int
	Accuracy   float64
}

// TestAccuracyPerDigit_M3_ACC_03_AllDigits validates per-digit accuracy.
// Requirement: All digits 0-9 must achieve ≥70% accuracy individually
// Evidence: table format with digit | correct | total | accuracy%
func TestAccuracyPerDigit_M3_ACC_03_AllDigits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full MNIST test in short mode")
	}

	// Load full MNIST test set (10,000 images)
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	if len(images) != 10000 {
		t.Fatalf("Expected 10,000 test images, got %d", len(images))
	}

	// Load pretrained weights with 8-bit quantization (CIM mode)
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Initialize per-digit statistics (digits 0-9)
	stats := make([]perDigitStats, 10)
	for i := 0; i < 10; i++ {
		stats[i].Digit = i
	}

	// Run inference and accumulate per-digit results
	for i := 0; i < len(images); i++ {
		trueLabel := labels[i]
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}

		stats[trueLabel].TotalCount++
		if result.CIMPrediction == trueLabel {
			stats[trueLabel].Correct++
		}
	}

	// Compute accuracy percentages
	for i := 0; i < 10; i++ {
		if stats[i].TotalCount > 0 {
			stats[i].Accuracy = 100.0 * float64(stats[i].Correct) / float64(stats[i].TotalCount)
		}
	}

	// Format and log table
	t.Logf("M3-ACC-03: Per-Digit Accuracy (8-bit CIM)")
	t.Logf("")
	t.Logf("  Digit | Correct | Total | Accuracy%%")
	t.Logf("  ------|---------|-------|----------")
	for i := 0; i < 10; i++ {
		t.Logf("    %d   |  %4d   | %4d  |  %6.2f%%",
			stats[i].Digit,
			stats[i].Correct,
			stats[i].TotalCount,
			stats[i].Accuracy,
		)
	}
	t.Logf("")

	// Validate requirement: all digits ≥70% accuracy
	const minPerDigitAccuracy = 70.0
	failedDigits := []string{}

	for i := 0; i < 10; i++ {
		if stats[i].Accuracy < minPerDigitAccuracy {
			failedDigits = append(failedDigits,
				fmt.Sprintf("digit %d: %.2f%%", stats[i].Digit, stats[i].Accuracy))
		}
	}

	if len(failedDigits) > 0 {
		t.Fatalf("M3-ACC-03: FAIL — %d digits below %.2f%% threshold:\n  %s",
			len(failedDigits),
			minPerDigitAccuracy,
			strings.Join(failedDigits, "\n  "))
	}

	// Find min/max for summary
	minAcc := stats[0].Accuracy
	maxAcc := stats[0].Accuracy
	minDigit := 0
	maxDigit := 0

	for i := 1; i < 10; i++ {
		if stats[i].Accuracy < minAcc {
			minAcc = stats[i].Accuracy
			minDigit = i
		}
		if stats[i].Accuracy > maxAcc {
			maxAcc = stats[i].Accuracy
			maxDigit = i
		}
	}

	// Compute overall accuracy from per-digit stats
	totalCorrect := 0
	totalSamples := 0
	for i := 0; i < 10; i++ {
		totalCorrect += stats[i].Correct
		totalSamples += stats[i].TotalCount
	}
	overallAccuracy := 100.0 * float64(totalCorrect) / float64(totalSamples)

	t.Logf("M3-ACC-03: Summary")
	t.Logf("  Overall:     %.2f%% (%d/%d)", overallAccuracy, totalCorrect, totalSamples)
	t.Logf("  Best digit:  %d (%.2f%%)", maxDigit, maxAcc)
	t.Logf("  Worst digit: %d (%.2f%%)", minDigit, minAcc)
	t.Logf("  Range:       %.2f%% (max - min)", maxAcc-minAcc)
	t.Logf("")
	t.Logf("M3-ACC-03: PASS — All digits meet ≥%.2f%% requirement", minPerDigitAccuracy)
}

// TestAccuracyPerDigit_M3_ACC_03_FP32 validates per-digit FP32 accuracy.
// This provides a reference for comparison with CIM quantized results.
func TestAccuracyPerDigit_M3_ACC_03_FP32(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full MNIST test in short mode")
	}

	// Load full MNIST test set
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Load pretrained weights (FP32 mode)
	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Initialize per-digit statistics
	stats := make([]perDigitStats, 10)
	for i := 0; i < 10; i++ {
		stats[i].Digit = i
	}

	// Run FP32 inference
	for i := 0; i < len(images); i++ {
		trueLabel := labels[i]
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}

		stats[trueLabel].TotalCount++
		if result.FPPrediction == trueLabel {
			stats[trueLabel].Correct++
		}
	}

	// Compute accuracy percentages
	for i := 0; i < 10; i++ {
		if stats[i].TotalCount > 0 {
			stats[i].Accuracy = 100.0 * float64(stats[i].Correct) / float64(stats[i].TotalCount)
		}
	}

	// Format and log table
	t.Logf("M3-ACC-03: Per-Digit Accuracy (FP32 Reference)")
	t.Logf("")
	t.Logf("  Digit | Correct | Total | Accuracy%%")
	t.Logf("  ------|---------|-------|----------")
	for i := 0; i < 10; i++ {
		t.Logf("    %d   |  %4d   | %4d  |  %6.2f%%",
			stats[i].Digit,
			stats[i].Correct,
			stats[i].TotalCount,
			stats[i].Accuracy,
		)
	}
	t.Logf("")

	// Find min/max
	minAcc := stats[0].Accuracy
	maxAcc := stats[0].Accuracy
	minDigit := 0
	maxDigit := 0

	for i := 1; i < 10; i++ {
		if stats[i].Accuracy < minAcc {
			minAcc = stats[i].Accuracy
			minDigit = i
		}
		if stats[i].Accuracy > maxAcc {
			maxAcc = stats[i].Accuracy
			maxDigit = i
		}
	}

	totalCorrect := 0
	totalSamples := 0
	for i := 0; i < 10; i++ {
		totalCorrect += stats[i].Correct
		totalSamples += stats[i].TotalCount
	}
	overallAccuracy := 100.0 * float64(totalCorrect) / float64(totalSamples)

	t.Logf("M3-ACC-03 FP32: Summary")
	t.Logf("  Overall:     %.2f%% (%d/%d)", overallAccuracy, totalCorrect, totalSamples)
	t.Logf("  Best digit:  %d (%.2f%%)", maxDigit, maxAcc)
	t.Logf("  Worst digit: %d (%.2f%%)", minDigit, minAcc)
	t.Logf("  Range:       %.2f%% (max - min)", maxAcc-minAcc)

	// Log any digits below 70% for FP32 (informational, not a hard requirement)
	const referenceAccuracy = 70.0
	for i := 0; i < 10; i++ {
		if stats[i].Accuracy < referenceAccuracy {
			t.Logf("NOTE: FP32 digit %d accuracy %.2f%% < %.2f%% (informational)",
				i, stats[i].Accuracy, referenceAccuracy)
		}
	}

	// The main requirement is overall accuracy ≥85%, not per-digit
	if overallAccuracy < 85.0 {
		t.Errorf("FP32 overall accuracy %.2f%% < 85.0%%", overallAccuracy)
	}
}
