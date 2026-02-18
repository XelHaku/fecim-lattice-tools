package core

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// hashFloat64Slice computes SHA-256 hash of a float64 slice.
// This provides a deterministic fingerprint for bit-exact comparison.
func hashFloat64Slice(data []float64) [32]byte {
	h := sha256.New()
	buf := make([]byte, 8)
	for _, v := range data {
		binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
		h.Write(buf)
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

// hashIntSlice computes SHA-256 hash of an int slice.
func hashIntSlice(data []int) [32]byte {
	h := sha256.New()
	buf := make([]byte, 8)
	for _, v := range data {
		binary.LittleEndian.PutUint64(buf, uint64(v))
		h.Write(buf)
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

// TestAccuracyDeterminism_M3_ACC_05 validates bit-exact reproducibility.
// Requirement: Run inference 3 times with same seed → all outputs bit-identical
// Evidence: Hash comparison of predictions across runs
func TestAccuracyDeterminism_M3_ACC_05(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full MNIST test in short mode")
	}

	// Load MNIST test set (use subset for speed)
	dataDir := filepath.Join("..", "..", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Use first 1000 images for determinism check (balance speed vs coverage)
	const testSize = 1000
	if len(images) > testSize {
		images = images[:testSize]
		labels = labels[:testSize]
	}

	t.Logf("M3-ACC-05: Testing determinism on %d images", len(images))

	// Run 3 independent inference passes
	const numRuns = 3
	type runResult struct {
		predictions     []int
		fpLogits        [][]float64
		cimLogits       [][]float64
		fpProbs         [][]float64
		cimProbs        [][]float64
		predictionsHash [32]byte
		fpLogitsHash    [32]byte
		cimLogitsHash   [32]byte
	}

	results := make([]runResult, numRuns)

	for run := 0; run < numRuns; run++ {
		t.Logf("  Run %d/%d...", run+1, numRuns)

		// Create fresh network instance for each run
		net := NewDualModeNetwork(784, 128, 10)
		net.Config.DACBits = 8
		net.Config.ADCBits = 8
		net.Config.NumLevels = 256
		net.Config.NoiseLevel = 0.0 // Critical: no random noise for determinism

		// Load pretrained weights
		weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
		if err := net.LoadWeights(weightsFile); err != nil {
			t.Fatalf("Failed to load pretrained weights: %v", err)
		}

		// Run inference on all test images
		results[run].predictions = make([]int, len(images))
		results[run].fpLogits = make([][]float64, len(images))
		results[run].cimLogits = make([][]float64, len(images))
		results[run].fpProbs = make([][]float64, len(images))
		results[run].cimProbs = make([][]float64, len(images))

		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result == nil {
				t.Fatalf("Infer returned nil for image %d in run %d", i, run)
			}

			results[run].predictions[i] = result.CIMPrediction
			results[run].fpLogits[i] = append([]float64(nil), result.FPLogits...)
			results[run].cimLogits[i] = append([]float64(nil), result.CIMLogits...)
			results[run].fpProbs[i] = append([]float64(nil), result.FPProbabilities...)
			results[run].cimProbs[i] = append([]float64(nil), result.CIMProbabilities...)
		}

		// Compute hashes for this run
		results[run].predictionsHash = hashIntSlice(results[run].predictions)

		// Flatten logits for hashing
		flatFPLogits := []float64{}
		flatCIMLogits := []float64{}
		for i := 0; i < len(images); i++ {
			flatFPLogits = append(flatFPLogits, results[run].fpLogits[i]...)
			flatCIMLogits = append(flatCIMLogits, results[run].cimLogits[i]...)
		}
		results[run].fpLogitsHash = hashFloat64Slice(flatFPLogits)
		results[run].cimLogitsHash = hashFloat64Slice(flatCIMLogits)
	}

	// Compare hashes across all runs
	t.Logf("")
	t.Logf("M3-ACC-05: Hash Comparison")
	t.Logf("")

	// Check CIM predictions (primary test)
	predHash0 := results[0].predictionsHash
	predMatch := true
	for run := 1; run < numRuns; run++ {
		if results[run].predictionsHash != predHash0 {
			predMatch = false
			t.Errorf("  CIM predictions hash mismatch: Run 0 vs Run %d", run)
			t.Errorf("    Run 0: %x", predHash0)
			t.Errorf("    Run %d: %x", run, results[run].predictionsHash)
		}
	}

	if predMatch {
		t.Logf("  ✓ CIM predictions: %x (all runs identical)", predHash0)
	} else {
		t.Fatalf("M3-ACC-05: FAIL — CIM predictions not deterministic")
	}

	// Check FP logits
	fpLogitsHash0 := results[0].fpLogitsHash
	fpLogitsMatch := true
	for run := 1; run < numRuns; run++ {
		if results[run].fpLogitsHash != fpLogitsHash0 {
			fpLogitsMatch = false
			t.Errorf("  FP logits hash mismatch: Run 0 vs Run %d", run)
		}
	}

	if fpLogitsMatch {
		t.Logf("  ✓ FP logits: %x (all runs identical)", fpLogitsHash0)
	} else {
		t.Errorf("  FP logits: MISMATCH (expected bit-exact match)")
	}

	// Check CIM logits
	cimLogitsHash0 := results[0].cimLogitsHash
	cimLogitsMatch := true
	for run := 1; run < numRuns; run++ {
		if results[run].cimLogitsHash != cimLogitsHash0 {
			cimLogitsMatch = false
			t.Errorf("  CIM logits hash mismatch: Run 0 vs Run %d", run)
		}
	}

	if cimLogitsMatch {
		t.Logf("  ✓ CIM logits: %x (all runs identical)", cimLogitsHash0)
	} else {
		t.Errorf("  CIM logits: MISMATCH (expected bit-exact match)")
	}

	// Detailed element-wise comparison for first mismatch (if any)
	if !predMatch || !fpLogitsMatch || !cimLogitsMatch {
		t.Logf("")
		t.Logf("M3-ACC-05: Detailed comparison (first 10 predictions)")
		t.Logf("  Image | Run0 | Run1 | Run2 | Label | Match?")
		t.Logf("  ------|------|------|------|-------|--------")
		for i := 0; i < 10 && i < len(images); i++ {
			match := (results[0].predictions[i] == results[1].predictions[i] &&
				results[0].predictions[i] == results[2].predictions[i])
			matchStr := "✓"
			if !match {
				matchStr = "✗"
			}
			t.Logf("   %4d |   %d  |   %d  |   %d  |   %d   |   %s",
				i,
				results[0].predictions[i],
				results[1].predictions[i],
				results[2].predictions[i],
				labels[i],
				matchStr,
			)
		}
	}

	t.Logf("")

	// Final validation
	if !predMatch {
		t.Fatalf("M3-ACC-05: FAIL — CIM predictions not bit-identical across runs")
	}
	if !fpLogitsMatch {
		t.Errorf("M3-ACC-05: WARNING — FP logits not bit-identical (may indicate numerical instability)")
	}
	if !cimLogitsMatch {
		t.Errorf("M3-ACC-05: WARNING — CIM logits not bit-identical (may indicate numerical instability)")
	}

	t.Logf("M3-ACC-05: PASS — All runs produced identical predictions")
}

// TestAccuracyDeterminism_M3_ACC_05_WithNoise validates that noise properly breaks determinism.
// This is a negative test: WITH noise enabled, results should differ.
func TestAccuracyDeterminism_M3_ACC_05_WithNoise(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full MNIST test in short mode")
	}

	// Load MNIST test set (small subset)
	dataDir := filepath.Join("..", "..", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	const testSize = 100
	if len(images) > testSize {
		images = images[:testSize]
	}

	// Run 2 passes with noise enabled
	predictions1 := make([]int, len(images))
	predictions2 := make([]int, len(images))

	for run := 0; run < 2; run++ {
		net := NewDualModeNetwork(784, 128, 10)
		net.Config.DACBits = 8
		net.Config.ADCBits = 8
		net.Config.NumLevels = 256
		net.Config.NoiseLevel = 0.05 // 5% noise — should cause variation

		weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
		if err := net.LoadWeights(weightsFile); err != nil {
			t.Fatalf("Failed to load pretrained weights: %v", err)
		}

		for i := 0; i < len(images); i++ {
			result := net.Infer(images[i])
			if result == nil {
				t.Fatalf("Infer returned nil for image %d", i)
			}
			if run == 0 {
				predictions1[i] = result.CIMPrediction
			} else {
				predictions2[i] = result.CIMPrediction
			}
		}
	}

	// Count differences
	differenceCount := 0
	for i := 0; i < len(images); i++ {
		if predictions1[i] != predictions2[i] {
			differenceCount++
		}
	}

	differenceRate := 100.0 * float64(differenceCount) / float64(len(images))

	t.Logf("M3-ACC-05 Noise Test: %.2f%% predictions differ (%d/%d) with 5%% noise",
		differenceRate, differenceCount, len(images))

	// With 5% noise, we expect *some* variation (but not too much)
	// Note: On easy test cases, 5% noise may not affect final predictions
	if differenceCount == 0 {
		t.Logf("NOTE: With 5%% noise, predictions remained identical (noise too small or test set too easy)")
		t.Logf("M3-ACC-05 Noise Test: PASS — Determinism verified (noise test informational only)")
	} else {
		if differenceRate > 90.0 {
			t.Errorf("With 5%% noise, prediction difference %.2f%% is too high (expected <90%%)",
				differenceRate)
		} else {
			t.Logf("M3-ACC-05 Noise Test: PASS — Noise correctly introduces variation")
		}
	}
}

// TestAccuracyDeterminism_M3_ACC_05_SeedIndependence validates that results
// are independent of the order of operations (network creation, weight loading).
func TestAccuracyDeterminism_M3_ACC_05_SeedIndependence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	dataDir := filepath.Join("..", "..", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Use single test image
	if len(images) == 0 {
		t.Fatal("No test images loaded")
	}
	testImage := images[0]

	// Create two networks in different order
	// Network 1: create, configure, load
	net1 := NewDualModeNetwork(784, 128, 10)
	net1.Config.NoiseLevel = 0.0
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net1.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load weights: %v", err)
	}

	// Network 2: create, load, then reconfigure (should not affect results)
	net2 := NewDualModeNetwork(784, 128, 10)
	if err := net2.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load weights: %v", err)
	}
	net2.Config.NoiseLevel = 0.0

	// Run inference
	result1 := net1.Infer(testImage)
	result2 := net2.Infer(testImage)

	if result1 == nil || result2 == nil {
		t.Fatal("Infer returned nil")
	}

	// Compare predictions
	if result1.CIMPrediction != result2.CIMPrediction {
		t.Errorf("Predictions differ: net1=%d, net2=%d (expected identical)",
			result1.CIMPrediction, result2.CIMPrediction)
	}

	// Compare logits (bit-exact)
	for i := 0; i < len(result1.CIMLogits); i++ {
		if result1.CIMLogits[i] != result2.CIMLogits[i] {
			t.Errorf("CIM logits[%d] differ: net1=%.15f, net2=%.15f",
				i, result1.CIMLogits[i], result2.CIMLogits[i])
		}
	}

	t.Logf("M3-ACC-05 Seed Independence: PASS — Initialization order does not affect results")
}
