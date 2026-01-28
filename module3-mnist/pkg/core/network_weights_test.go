package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ===========================
// Test Helpers
// ===========================

// createTestWeightsJSON creates a temporary JSON weights file for testing.
func createTestWeightsJSON(t *testing.T, tempDir string, filename string, wf *WeightsFile) string {
	t.Helper()

	filepath := filepath.Join(tempDir, filename)
	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal weights: %v", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		t.Fatalf("Failed to write weights file: %v", err)
	}

	return filepath
}

// createValidTestWeights creates a valid WeightsFile structure for testing.
func createValidTestWeights() *WeightsFile {
	// Create simple 784->128->10 network weights
	layer1 := make([][]float64, 128)
	for i := range layer1 {
		layer1[i] = make([]float64, 784)
		for j := range layer1[i] {
			layer1[i][j] = 0.01 * float64(i+j) // Simple pattern
		}
	}

	layer2 := make([][]float64, 10)
	for i := range layer2 {
		layer2[i] = make([]float64, 128)
		for j := range layer2[i] {
			layer2[i][j] = 0.02 * float64(i*j) // Simple pattern
		}
	}

	biases1 := make([]float64, 128)
	for i := range biases1 {
		biases1[i] = 0.1 * float64(i)
	}

	biases2 := make([]float64, 10)
	for i := range biases2 {
		biases2[i] = 0.2 * float64(i)
	}

	return &WeightsFile{
		Layer1Weights: layer1,
		Layer2Weights: layer2,
		Biases1:       biases1,
		Biases2:       biases2,
	}
}

// createSingleLayerTestWeights creates weights with single-layer architecture.
func createSingleLayerTestWeights() *WeightsFile {
	wf := createValidTestWeights()

	// Add single-layer weights (784->10)
	singleLayer := make([][]float64, 10)
	for i := range singleLayer {
		singleLayer[i] = make([]float64, 784)
		for j := range singleLayer[i] {
			singleLayer[i][j] = 0.005 * float64(i-j)
		}
	}

	singleBias := make([]float64, 10)
	for i := range singleBias {
		singleBias[i] = 0.3 * float64(i)
	}

	wf.SingleLayerWeights = singleLayer
	wf.SingleLayerBias = singleBias

	return wf
}

// ===========================
// 1. LoadWeights tests
// ===========================

func TestLoadWeights_ValidJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Create valid test weights
	wf := createValidTestWeights()
	filepath := createTestWeightsJSON(t, tempDir, "weights.json", wf)

	// Create network and load weights
	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(filepath)

	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify network architecture matches loaded weights
	if net.InputSize != 784 {
		t.Errorf("Expected InputSize=784, got %d", net.InputSize)
	}
	if net.HiddenSize != 128 {
		t.Errorf("Expected HiddenSize=128, got %d", net.HiddenSize)
	}
	if net.OutputSize != 10 {
		t.Errorf("Expected OutputSize=10, got %d", net.OutputSize)
	}

	// Verify weights were loaded correctly
	if len(net.FPWeights1) != 128 {
		t.Errorf("Expected 128 hidden neurons, got %d", len(net.FPWeights1))
	}
	if len(net.FPWeights2) != 10 {
		t.Errorf("Expected 10 output neurons, got %d", len(net.FPWeights2))
	}

	// Verify a sample weight value (no scale/offset, so should be identical)
	expectedW1 := wf.Layer1Weights[0][0]
	actualW1 := net.FPWeights1[0][0]
	if actualW1 != expectedW1 {
		t.Errorf("Weight mismatch: expected %f, got %f", expectedW1, actualW1)
	}

	// Verify biases
	if len(net.FPBias1) != 128 {
		t.Errorf("Expected 128 bias values for layer 1, got %d", len(net.FPBias1))
	}
	if len(net.FPBias2) != 10 {
		t.Errorf("Expected 10 bias values for layer 2, got %d", len(net.FPBias2))
	}

	// Verify quantized weights were initialized
	if len(net.QuantWeights1) != 128 {
		t.Errorf("Expected 128 quantized weights for layer 1, got %d", len(net.QuantWeights1))
	}
	if len(net.QuantWeights2) != 10 {
		t.Errorf("Expected 10 quantized weights for layer 2, got %d", len(net.QuantWeights2))
	}
}

func TestLoadWeights_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	filepath := filepath.Join(tempDir, "invalid.json")

	// Write malformed JSON
	invalidJSON := []byte(`{"layer1_weights": [[[invalid]]]`)
	if err := os.WriteFile(filepath, invalidJSON, 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(filepath)

	if err == nil {
		t.Error("Expected error for malformed JSON, got nil")
	}
}

func TestLoadWeights_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "nonexistent.json")

	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(nonExistentPath)

	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}

	// Verify error is file-not-found related
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got: %v", err)
	}
}

func TestLoadWeights_WithScaleOffset(t *testing.T) {
	tempDir := t.TempDir()

	wf := createValidTestWeights()

	// Add scale and offset for weight reconstruction
	// Simulates weights saved as: quantized = (fp - offset) / scale
	// So reconstruction: fp = quantized * scale + offset
	wf.L1Scale = 0.001
	wf.L1Offset = -0.5
	wf.L2Scale = 0.002
	wf.L2Offset = -1.0

	// Modify weights to be "quantized" values
	for i := range wf.Layer1Weights {
		for j := range wf.Layer1Weights[i] {
			// Store quantized value: (fp - offset) / scale
			fp := 0.01 * float64(i+j)
			wf.Layer1Weights[i][j] = (fp - wf.L1Offset) / wf.L1Scale
		}
	}

	filepath := createTestWeightsJSON(t, tempDir, "scaled_weights.json", wf)

	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(filepath)

	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify weights were correctly reconstructed: fp = quantized * scale + offset
	expectedFP := 0.01 * float64(0+0) // i=0, j=0
	quantized := (expectedFP - wf.L1Offset) / wf.L1Scale
	reconstructedFP := quantized*wf.L1Scale + wf.L1Offset

	actualFP := net.FPWeights1[0][0]

	// Should match within floating-point precision
	if abs(actualFP-expectedFP) > 1e-9 {
		t.Errorf("Weight reconstruction failed: expected %f, got %f (diff: %e)",
			expectedFP, actualFP, abs(actualFP-expectedFP))
	}

	// Verify reconstruction formula worked correctly
	if abs(actualFP-reconstructedFP) > 1e-9 {
		t.Errorf("Reconstruction formula mismatch: expected %f, got %f",
			reconstructedFP, actualFP)
	}
}

func TestLoadWeights_WithoutScaleOffset(t *testing.T) {
	tempDir := t.TempDir()

	wf := createValidTestWeights()
	// Explicitly set scale to 0 (or leave it at default 0)
	wf.L1Scale = 0
	wf.L2Scale = 0

	filepath := createTestWeightsJSON(t, tempDir, "direct_weights.json", wf)

	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(filepath)

	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify weights were used directly (no scaling/offset applied)
	expectedW1 := wf.Layer1Weights[0][0]
	actualW1 := net.FPWeights1[0][0]

	if actualW1 != expectedW1 {
		t.Errorf("Weight should be used directly: expected %f, got %f", expectedW1, actualW1)
	}

	// Verify layer 2 as well
	expectedW2 := wf.Layer2Weights[0][0]
	actualW2 := net.FPWeights2[0][0]

	if actualW2 != expectedW2 {
		t.Errorf("Layer2 weight should be used directly: expected %f, got %f", expectedW2, actualW2)
	}
}

func TestLoadWeights_SingleLayerWeights(t *testing.T) {
	tempDir := t.TempDir()

	wf := createSingleLayerTestWeights()
	filepath := createTestWeightsJSON(t, tempDir, "single_layer.json", wf)

	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(filepath)

	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify single-layer weights were loaded
	if len(net.SingleLayerWeights) != 10 {
		t.Errorf("Expected 10 single-layer output neurons, got %d", len(net.SingleLayerWeights))
	}

	if len(net.SingleLayerWeights[0]) != 784 {
		t.Errorf("Expected 784 single-layer inputs, got %d", len(net.SingleLayerWeights[0]))
	}

	// Verify single-layer weights match loaded values
	expectedSingleW := wf.SingleLayerWeights[0][0]
	actualSingleW := net.SingleLayerWeights[0][0]

	if actualSingleW != expectedSingleW {
		t.Errorf("Single-layer weight mismatch: expected %f, got %f", expectedSingleW, actualSingleW)
	}

	// Verify single-layer bias
	if len(net.SingleLayerBias) != 10 {
		t.Errorf("Expected 10 single-layer bias values, got %d", len(net.SingleLayerBias))
	}

	expectedBias := wf.SingleLayerBias[0]
	actualBias := net.SingleLayerBias[0]

	if actualBias != expectedBias {
		t.Errorf("Single-layer bias mismatch: expected %f, got %f", expectedBias, actualBias)
	}

	// Verify quantized single-layer weights were initialized
	if len(net.QuantSingleLayerWeights) != 10 {
		t.Errorf("Expected 10 quantized single-layer weights, got %d", len(net.QuantSingleLayerWeights))
	}
}

func TestLoadWeights_PerLayerQuantLevels(t *testing.T) {
	tempDir := t.TempDir()

	wf := createValidTestWeights()

	// Set per-layer quantization levels
	wf.Layer1QuantLevels = 20
	wf.Layer2QuantLevels = 10

	filepath := createTestWeightsJSON(t, tempDir, "per_layer_quant.json", wf)

	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(filepath)

	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify per-layer quantization levels were set correctly
	if net.Config.Layer1Levels != 20 {
		t.Errorf("Expected Layer1Levels=20, got %d", net.Config.Layer1Levels)
	}

	if net.Config.Layer2Levels != 10 {
		t.Errorf("Expected Layer2Levels=10, got %d", net.Config.Layer2Levels)
	}
}

func TestLoadWeights_LegacyQuantLevels(t *testing.T) {
	tempDir := t.TempDir()

	wf := createValidTestWeights()

	// Set legacy uniform quantization level
	wf.QuantLevels = 15
	// Do NOT set per-layer levels (test backward compatibility)

	filepath := createTestWeightsJSON(t, tempDir, "legacy_quant.json", wf)

	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeights(filepath)

	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify legacy quantization level applied to both layers
	if net.Config.Layer1Levels != 15 {
		t.Errorf("Expected Layer1Levels=15 (from legacy quant_levels), got %d", net.Config.Layer1Levels)
	}

	if net.Config.Layer2Levels != 15 {
		t.Errorf("Expected Layer2Levels=15 (from legacy quant_levels), got %d", net.Config.Layer2Levels)
	}
}

// ===========================
// 2. LoadWeightsForLevel tests
// ===========================

func TestLoadWeightsForLevel_ExactMatch(t *testing.T) {
	tempDir := t.TempDir()

	// Create weight files for exact matches
	for _, level := range []int{10, 20, 29, 31} {
		wf := createValidTestWeights()
		wf.QuantLevels = level
		filename := fmt.Sprintf("pretrained_weights_%d.json", level)
		createTestWeightsJSON(t, tempDir, filename, wf)
	}

	// Test exact match for each level
	testCases := []int{10, 20, 29, 31}

	for _, level := range testCases {
		t.Run(fmt.Sprintf("Level%d", level), func(t *testing.T) {
			net := NewDualModeNetwork(784, 128, 10)
			err := net.LoadWeightsForLevel(tempDir, level)

			if err != nil {
				t.Errorf("LoadWeightsForLevel(%d) failed: %v", level, err)
			}

			// Verify the correct level was loaded
			if net.Config.Layer1Levels != level {
				t.Errorf("Expected config to load level %d, got %d", level, net.Config.Layer1Levels)
			}
		})
	}
}

func TestLoadWeightsForLevel_FallbackToNearest(t *testing.T) {
	tempDir := t.TempDir()

	// Create only 10 and 20 level weights
	wf10 := createValidTestWeights()
	wf10.QuantLevels = 10
	createTestWeightsJSON(t, tempDir, "pretrained_weights_10.json", wf10)

	wf20 := createValidTestWeights()
	wf20.QuantLevels = 20
	createTestWeightsJSON(t, tempDir, "pretrained_weights_20.json", wf20)

	// Also create default weights as final fallback
	wfDefault := createValidTestWeights()
	wfDefault.QuantLevels = 30
	createTestWeightsJSON(t, tempDir, "pretrained_weights.json", wfDefault)

	// Request level 15 - should fall back to nearest (10 or 20)
	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeightsForLevel(tempDir, 15)

	if err != nil {
		t.Fatalf("LoadWeightsForLevel(15) failed: %v", err)
	}

	// Should use either 10 or 20 (both are distance 5 from 15)
	// GetBestMatchingWeightsLevel will pick the first one in the list with minimum distance
	loadedLevel := net.Config.Layer1Levels
	if loadedLevel != 10 && loadedLevel != 20 {
		t.Errorf("Expected fallback to 10 or 20 for level 15, got %d", loadedLevel)
	}
}

func TestLoadWeightsForLevel_DefaultFallback(t *testing.T) {
	tempDir := t.TempDir()

	// Create only default weights (no level-specific files)
	wfDefault := createValidTestWeights()
	wfDefault.QuantLevels = 30
	createTestWeightsJSON(t, tempDir, "pretrained_weights.json", wfDefault)

	// Request level 25 - should fall back to default since no level-specific files exist
	net := NewDualModeNetwork(784, 128, 10)
	err := net.LoadWeightsForLevel(tempDir, 25)

	if err != nil {
		t.Fatalf("LoadWeightsForLevel(25) failed: %v", err)
	}

	// Should have loaded default 30-level weights
	if net.Config.Layer1Levels != 30 {
		t.Errorf("Expected fallback to default 30 levels, got %d", net.Config.Layer1Levels)
	}
}

// ===========================
// 3. GetBestMatchingWeightsLevel tests
// ===========================

// createTestWeightsDir creates a temp directory with mock weight files for testing.
func createTestWeightsDir(t *testing.T, levels []int) string {
	t.Helper()
	dir := t.TempDir()

	for _, l := range levels {
		var filename string
		if l == 30 {
			filename = "pretrained_weights.json"
		} else {
			filename = fmt.Sprintf("pretrained_weights_%d.json", l)
		}
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create test weight file: %v", err)
		}
	}

	return dir
}

func TestScanAvailableQATLevels(t *testing.T) {
	// Test with multiple levels
	t.Run("MultipleLevels", func(t *testing.T) {
		dir := createTestWeightsDir(t, []int{10, 20, 30})
		levels := ScanAvailableQATLevels(dir)
		expected := []int{10, 20, 30}
		if len(levels) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, levels)
		}
		for i, l := range levels {
			if l != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], l)
			}
		}
	})

	// Test with only default 30-level weights
	t.Run("OnlyDefault30", func(t *testing.T) {
		dir := createTestWeightsDir(t, []int{30})
		levels := ScanAvailableQATLevels(dir)
		if len(levels) != 1 || levels[0] != 30 {
			t.Errorf("Expected [30], got %v", levels)
		}
	})

	// Test with empty directory (should return [30] as fallback)
	t.Run("EmptyDir", func(t *testing.T) {
		dir := t.TempDir()
		levels := ScanAvailableQATLevels(dir)
		if len(levels) != 1 || levels[0] != 30 {
			t.Errorf("Expected [30] fallback, got %v", levels)
		}
	})
}

func TestGetBestMatchingWeightsLevel_ExactMatch(t *testing.T) {
	dir := createTestWeightsDir(t, []int{10, 20, 29, 30, 31})

	testCases := []struct {
		target   int
		expected int
	}{
		{10, 10},
		{20, 20},
		{29, 29},
		{30, 30},
		{31, 31},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Target%d", tc.target), func(t *testing.T) {
			result := GetBestMatchingWeightsLevel(dir, tc.target)
			if result != tc.expected {
				t.Errorf("GetBestMatchingWeightsLevel(%d): expected %d, got %d",
					tc.target, tc.expected, result)
			}
		})
	}
}

func TestGetBestMatchingWeightsLevel_NearestLevel(t *testing.T) {
	dir := createTestWeightsDir(t, []int{10, 20, 29, 30, 31})

	testCases := []struct {
		target   int
		expected int
		desc     string
	}{
		{5, 10, "Below minimum"},
		{15, 10, "Midpoint 10-20, should pick first (10)"},
		{18, 20, "Closer to 20"},
		{12, 10, "Closer to 10"},
		{25, 29, "Closer to 29 than 20"},
		{28, 29, "Very close to 29"},
		{32, 31, "Above maximum"},
		{50, 31, "Far above maximum"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := GetBestMatchingWeightsLevel(dir, tc.target)
			if result != tc.expected {
				t.Errorf("GetBestMatchingWeightsLevel(%d): expected %d, got %d (%s)",
					tc.target, tc.expected, result, tc.desc)
			}
		})
	}
}

func TestGetBestMatchingWeightsLevel_AllCases(t *testing.T) {
	dir := createTestWeightsDir(t, []int{10, 20, 29, 30, 31})

	// Table-driven test covering all edge cases
	testCases := []struct {
		target   int
		expected int
	}{
		// Exact matches
		{10, 10},
		{20, 20},
		{29, 29},
		{30, 30},
		{31, 31},

		// Below minimum
		{1, 10},
		{5, 10},
		{9, 10},

		// Between 10 and 20
		{11, 10},
		{14, 10},
		{15, 10}, // Equidistant: picks first (10)
		{16, 20},
		{19, 20},

		// Between 20 and 29
		{21, 20},
		{24, 20},
		{25, 29}, // Equidistant: picks first in list (20 comes before 29, but 29-25=4, 25-20=5)
		{27, 29},
		{28, 29},

		// Between 29 and 30
		{29, 29}, // Exact
		{30, 30}, // Exact

		// Between 30 and 31
		{30, 30}, // Exact
		{31, 31}, // Exact

		// Above maximum
		{32, 31},
		{35, 31},
		{100, 31},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Level%d", tc.target), func(t *testing.T) {
			result := GetBestMatchingWeightsLevel(dir, tc.target)
			if result != tc.expected {
				t.Errorf("Target %d: expected %d, got %d", tc.target, tc.expected, result)
			}
		})
	}
}

// ===========================
// 4. GetWeightsFilename tests
// ===========================

func TestGetWeightsFilename_AvailableLevels(t *testing.T) {
	// Create temp dir with actual weight files
	dir := createTestWeightsDir(t, []int{10, 20, 29, 30, 31})

	testCases := []struct {
		level    int
		expected string
	}{
		{10, filepath.Join(dir, "pretrained_weights_10.json")},
		{20, filepath.Join(dir, "pretrained_weights_20.json")},
		{29, filepath.Join(dir, "pretrained_weights_29.json")},
		{31, filepath.Join(dir, "pretrained_weights_31.json")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Level%d", tc.level), func(t *testing.T) {
			result := GetWeightsFilename(dir, tc.level)
			if result != tc.expected {
				t.Errorf("GetWeightsFilename(%d): expected %s, got %s",
					tc.level, tc.expected, result)
			}
		})
	}
}

func TestGetWeightsFilename_Default30(t *testing.T) {
	// Create temp dir with only default 30-level weights
	dir := createTestWeightsDir(t, []int{30})
	expected := filepath.Join(dir, "pretrained_weights.json")

	// Level 30 should return default filename (not pretrained_weights_30.json)
	result := GetWeightsFilename(dir, 30)
	if result != expected {
		t.Errorf("GetWeightsFilename(30): expected %s, got %s", expected, result)
	}

	// Level without weights file should fall back to default
	result = GetWeightsFilename(dir, 15)
	if result != expected {
		t.Errorf("GetWeightsFilename(15): expected %s (fallback), got %s", expected, result)
	}
}

// ===========================
// 5. Getter tests
// ===========================

func TestGetFPWeights_ReturnsCopy(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Set a known value in FP weights
	net.FPWeights1[0][0] = 1.234
	originalValue := net.FPWeights1[0][0]

	// Get weights via getter
	w1, _, _, _ := net.GetFPWeights()

	// Modify returned slice
	w1[0][0] = 9.999

	// Verify original network weights are unchanged
	if net.FPWeights1[0][0] != originalValue {
		t.Errorf("GetFPWeights should return a copy: network weights were modified")
	}

	if net.FPWeights1[0][0] == 9.999 {
		t.Errorf("GetFPWeights returned reference, not copy: modification affected network")
	}
}

func TestGetFPWeights_NotAffectedByModification(t *testing.T) {
	tempDir := t.TempDir()

	wf := createValidTestWeights()
	filepath := createTestWeightsJSON(t, tempDir, "weights.json", wf)

	net := NewDualModeNetwork(784, 128, 10)
	if err := net.LoadWeights(filepath); err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Get original values
	originalW1 := net.FPWeights1[5][10]
	originalW2 := net.FPWeights2[3][7]
	originalB1 := net.FPBias1[15]
	originalB2 := net.FPBias2[4]

	// Get weights via getter
	w1, w2, b1, b2 := net.GetFPWeights()

	// Modify all returned slices
	w1[5][10] = 999.0
	w2[3][7] = 888.0
	b1[15] = 777.0
	b2[4] = 666.0

	// Verify network weights are completely unchanged
	if net.FPWeights1[5][10] != originalW1 {
		t.Errorf("FPWeights1 was modified: expected %f, got %f", originalW1, net.FPWeights1[5][10])
	}
	if net.FPWeights2[3][7] != originalW2 {
		t.Errorf("FPWeights2 was modified: expected %f, got %f", originalW2, net.FPWeights2[3][7])
	}
	if net.FPBias1[15] != originalB1 {
		t.Errorf("FPBias1 was modified: expected %f, got %f", originalB1, net.FPBias1[15])
	}
	if net.FPBias2[4] != originalB2 {
		t.Errorf("FPBias2 was modified: expected %f, got %f", originalB2, net.FPBias2[4])
	}
}

func TestGetQuantWeights_ReturnsCopy(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// Set a known value in quantized weights
	net.QuantWeights1[0][0] = 0.567
	originalValue := net.QuantWeights1[0][0]

	// Get weights via getter
	w1, _, _, _ := net.GetQuantWeights()

	// Modify returned slice
	w1[0][0] = 7.777

	// Verify original network weights are unchanged
	if net.QuantWeights1[0][0] != originalValue {
		t.Errorf("GetQuantWeights should return a copy: network weights were modified")
	}

	if net.QuantWeights1[0][0] == 7.777 {
		t.Errorf("GetQuantWeights returned reference, not copy: modification affected network")
	}
}

func TestGetQuantWeights_MatchesRequantized(t *testing.T) {
	tempDir := t.TempDir()

	wf := createValidTestWeights()
	filepath := createTestWeightsJSON(t, tempDir, "weights.json", wf)

	net := NewDualModeNetwork(784, 128, 10)
	if err := net.LoadWeights(filepath); err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Get quantized weights
	qw1Before, qw2Before, qb1Before, qb2Before := net.GetQuantWeights()

	// Change quantization level and requantize
	net.Config.NumLevels = 20
	net.RequantizeWeights()

	// Get new quantized weights
	qw1After, qw2After, qb1After, qb2After := net.GetQuantWeights()

	// Verify weights changed after requantization
	// (They should be different because quantization levels changed)
	changed := false
	for i := 0; i < len(qw1Before); i++ {
		for j := 0; j < len(qw1Before[i]); j++ {
			if qw1Before[i][j] != qw1After[i][j] {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Error("Quantized weights should have changed after requantization to different level")
	}

	// Verify getter returns correctly requantized values
	// by checking that they match network's internal quantized weights
	for i := 0; i < len(qw1After); i++ {
		for j := 0; j < len(qw1After[i]); j++ {
			if qw1After[i][j] != net.QuantWeights1[i][j] {
				t.Errorf("GetQuantWeights mismatch at [%d][%d]: getter=%f, internal=%f",
					i, j, qw1After[i][j], net.QuantWeights1[i][j])
			}
		}
	}

	// Verify biases as well
	for i := 0; i < len(qb1After); i++ {
		if qb1After[i] != net.QuantBias1[i] {
			t.Errorf("GetQuantWeights bias1 mismatch at [%d]: getter=%f, internal=%f",
				i, qb1After[i], net.QuantBias1[i])
		}
	}

	for i := 0; i < len(qb2After); i++ {
		if qb2After[i] != net.QuantBias2[i] {
			t.Errorf("GetQuantWeights bias2 mismatch at [%d]: getter=%f, internal=%f",
				i, qb2After[i], net.QuantBias2[i])
		}
	}

	// Suppress unused variable warnings
	_, _, _, _ = qw2Before, qw2After, qb1Before, qb2Before
}

// ===========================
// Helper functions
// ===========================

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
