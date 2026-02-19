package neural

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// confusionMatrix represents a 10×10 matrix for MNIST digit classification.
// Entry [i][j] = count of true label i predicted as j
type confusionMatrix [10][10]int

// TestAccuracyConfusionMatrix_M3_ACC_04 generates and validates the confusion matrix.
// Requirement: No class should have >50% misclassification to any other single class
// Evidence: Full 10×10 matrix in readable format + validation results
func TestAccuracyConfusionMatrix_M3_ACC_04(t *testing.T) {
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

	// Initialize confusion matrix
	var cm confusionMatrix

	// Run inference and populate confusion matrix
	for i := 0; i < len(images); i++ {
		trueLabel := labels[i]
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}
		predicted := result.CIMPrediction

		// Sanity checks
		if trueLabel < 0 || trueLabel > 9 {
			t.Fatalf("Invalid true label %d for image %d", trueLabel, i)
		}
		if predicted < 0 || predicted > 9 {
			t.Fatalf("Invalid prediction %d for image %d", predicted, i)
		}

		cm[trueLabel][predicted]++
	}

	// Log confusion matrix in readable format
	t.Logf("M3-ACC-04: Confusion Matrix (8-bit CIM)")
	t.Logf("")
	t.Logf("True\\Pred |   0    1    2    3    4    5    6    7    8    9")
	t.Logf("----------|--------------------------------------------------")

	for trueDigit := 0; trueDigit < 10; trueDigit++ {
		line := fmt.Sprintf("    %d     |", trueDigit)
		for predDigit := 0; predDigit < 10; predDigit++ {
			line += fmt.Sprintf(" %4d", cm[trueDigit][predDigit])
		}
		t.Logf("%s", line)
	}
	t.Logf("")

	// Validate requirement: no class should have >50% misclassification to any other single class
	violations := []string{}
	const maxMisclassificationRate = 50.0

	for trueDigit := 0; trueDigit < 10; trueDigit++ {
		// Calculate total samples for this true digit
		totalForDigit := 0
		for predDigit := 0; predDigit < 10; predDigit++ {
			totalForDigit += cm[trueDigit][predDigit]
		}

		if totalForDigit == 0 {
			t.Fatalf("No samples found for digit %d (data corruption?)", trueDigit)
		}

		// Check each off-diagonal element (misclassifications)
		for predDigit := 0; predDigit < 10; predDigit++ {
			if predDigit == trueDigit {
				continue // Skip diagonal (correct classifications)
			}

			misclassCount := cm[trueDigit][predDigit]
			misclassRate := 100.0 * float64(misclassCount) / float64(totalForDigit)

			if misclassRate > maxMisclassificationRate {
				violations = append(violations,
					fmt.Sprintf("True %d → Predicted %d: %d/%d (%.2f%%)",
						trueDigit, predDigit, misclassCount, totalForDigit, misclassRate))
			}
		}
	}

	if len(violations) > 0 {
		t.Fatalf("M3-ACC-04: FAIL — %d violations of ≤%.2f%% misclassification:\n  %s",
			len(violations),
			maxMisclassificationRate,
			strings.Join(violations, "\n  "))
	}

	// Compute and log additional metrics
	totalSamples := 0
	totalCorrect := 0
	for trueDigit := 0; trueDigit < 10; trueDigit++ {
		for predDigit := 0; predDigit < 10; predDigit++ {
			totalSamples += cm[trueDigit][predDigit]
			if trueDigit == predDigit {
				totalCorrect += cm[trueDigit][predDigit]
			}
		}
	}

	overallAccuracy := 100.0 * float64(totalCorrect) / float64(totalSamples)

	t.Logf("M3-ACC-04: Summary")
	t.Logf("  Overall accuracy: %.2f%% (%d/%d)", overallAccuracy, totalCorrect, totalSamples)
	t.Logf("  Max single-class misclassification: ≤%.2f%%", maxMisclassificationRate)
	t.Logf("")
	t.Logf("M3-ACC-04: PASS — No class exceeds %.2f%% misclassification to any single class",
		maxMisclassificationRate)
}

// TestAccuracyConfusionMatrix_M3_ACC_04_Analysis provides detailed analysis
// of the confusion matrix, including most common confusions.
func TestAccuracyConfusionMatrix_M3_ACC_04_Analysis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full MNIST test in short mode")
	}

	// Load MNIST test set
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Load pretrained weights (CIM mode)
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Build confusion matrix
	var cm confusionMatrix
	for i := 0; i < len(images); i++ {
		trueLabel := labels[i]
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}
		cm[trueLabel][result.CIMPrediction]++
	}

	// Find top misclassifications (most common errors)
	type confusion struct {
		trueDigit int
		predDigit int
		count     int
		rate      float64
	}

	confusions := []confusion{}

	for trueDigit := 0; trueDigit < 10; trueDigit++ {
		totalForDigit := 0
		for predDigit := 0; predDigit < 10; predDigit++ {
			totalForDigit += cm[trueDigit][predDigit]
		}

		for predDigit := 0; predDigit < 10; predDigit++ {
			if predDigit == trueDigit {
				continue
			}
			if cm[trueDigit][predDigit] > 0 {
				rate := 100.0 * float64(cm[trueDigit][predDigit]) / float64(totalForDigit)
				confusions = append(confusions, confusion{
					trueDigit: trueDigit,
					predDigit: predDigit,
					count:     cm[trueDigit][predDigit],
					rate:      rate,
				})
			}
		}
	}

	// Sort by count (descending)
	for i := 0; i < len(confusions); i++ {
		for j := i + 1; j < len(confusions); j++ {
			if confusions[j].count > confusions[i].count {
				confusions[i], confusions[j] = confusions[j], confusions[i]
			}
		}
	}

	// Log top 10 misclassifications
	t.Logf("M3-ACC-04 Analysis: Top 10 Misclassifications")
	t.Logf("")
	t.Logf("  Rank | True → Pred | Count | Rate")
	t.Logf("  -----|-------------|-------|-------")

	topN := 10
	if len(confusions) < topN {
		topN = len(confusions)
	}

	for i := 0; i < topN; i++ {
		c := confusions[i]
		t.Logf("   %2d  |  %d → %d      |  %4d | %5.2f%%",
			i+1, c.trueDigit, c.predDigit, c.count, c.rate)
	}
	t.Logf("")

	// Report per-digit precision and recall
	t.Logf("M3-ACC-04 Analysis: Per-Digit Precision & Recall")
	t.Logf("")
	t.Logf("  Digit | Precision | Recall")
	t.Logf("  ------|-----------|--------")

	for digit := 0; digit < 10; digit++ {
		// Precision: of all predictions for this digit, how many were correct?
		predictedAsDigit := 0
		for trueDigit := 0; trueDigit < 10; trueDigit++ {
			predictedAsDigit += cm[trueDigit][digit]
		}
		precision := 0.0
		if predictedAsDigit > 0 {
			precision = 100.0 * float64(cm[digit][digit]) / float64(predictedAsDigit)
		}

		// Recall: of all true instances of this digit, how many were predicted correctly?
		trueInstances := 0
		for predDigit := 0; predDigit < 10; predDigit++ {
			trueInstances += cm[digit][predDigit]
		}
		recall := 0.0
		if trueInstances > 0 {
			recall = 100.0 * float64(cm[digit][digit]) / float64(trueInstances)
		}

		t.Logf("    %d   |   %6.2f%% |  %6.2f%%", digit, precision, recall)
	}
	t.Logf("")
}
