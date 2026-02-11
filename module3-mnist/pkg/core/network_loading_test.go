package core

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// TestNetworkArchitecture_Dimensions verifies network dimensions are set correctly.
func TestNetworkArchitecture_Dimensions(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	if net.InputSize != 784 {
		t.Errorf("InputSize = %d, want 784", net.InputSize)
	}
	if net.HiddenSize != 128 {
		t.Errorf("HiddenSize = %d, want 128", net.HiddenSize)
	}
	if net.OutputSize != 10 {
		t.Errorf("OutputSize = %d, want 10", net.OutputSize)
	}
}

// TestNetworkArchitecture_WeightShapes verifies weight matrix shapes.
func TestNetworkArchitecture_WeightShapes(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	// FPWeights1 should be [HiddenSize][InputSize]
	if len(net.FPWeights1) != 128 {
		t.Errorf("FPWeights1 rows = %d, want 128", len(net.FPWeights1))
	}
	if len(net.FPWeights1) > 0 && len(net.FPWeights1[0]) != 784 {
		t.Errorf("FPWeights1 cols = %d, want 784", len(net.FPWeights1[0]))
	}

	// FPWeights2 should be [OutputSize][HiddenSize]
	if len(net.FPWeights2) != 10 {
		t.Errorf("FPWeights2 rows = %d, want 10", len(net.FPWeights2))
	}
	if len(net.FPWeights2) > 0 && len(net.FPWeights2[0]) != 128 {
		t.Errorf("FPWeights2 cols = %d, want 128", len(net.FPWeights2[0]))
	}

	// QuantWeights1 should match FPWeights1 shape
	if len(net.QuantWeights1) != 128 {
		t.Errorf("QuantWeights1 rows = %d, want 128", len(net.QuantWeights1))
	}
	if len(net.QuantWeights1) > 0 && len(net.QuantWeights1[0]) != 784 {
		t.Errorf("QuantWeights1 cols = %d, want 784", len(net.QuantWeights1[0]))
	}

	// QuantWeights2 should match FPWeights2 shape
	if len(net.QuantWeights2) != 10 {
		t.Errorf("QuantWeights2 rows = %d, want 10", len(net.QuantWeights2))
	}
	if len(net.QuantWeights2) > 0 && len(net.QuantWeights2[0]) != 128 {
		t.Errorf("QuantWeights2 cols = %d, want 128", len(net.QuantWeights2[0]))
	}

	// Bias shapes
	if len(net.FPBias1) != 128 {
		t.Errorf("FPBias1 length = %d, want 128", len(net.FPBias1))
	}
	if len(net.FPBias2) != 10 {
		t.Errorf("FPBias2 length = %d, want 10", len(net.FPBias2))
	}

	// Single-layer weights should be [OutputSize][InputSize]
	if len(net.SingleLayerWeights) != 10 {
		t.Errorf("SingleLayerWeights rows = %d, want 10", len(net.SingleLayerWeights))
	}
	if len(net.SingleLayerWeights) > 0 && len(net.SingleLayerWeights[0]) != 784 {
		t.Errorf("SingleLayerWeights cols = %d, want 784", len(net.SingleLayerWeights[0]))
	}
}

// TestNetworkArchitecture_ConfigDefaults verifies default configuration.
func TestNetworkArchitecture_ConfigDefaults(t *testing.T) {
	net := NewDualModeNetwork(784, 128, 10)

	if net.Config == nil {
		t.Fatal("Config is nil")
	}

	if net.Config.NumLevels != 30 {
		t.Errorf("NumLevels = %d, want 30", net.Config.NumLevels)
	}
	if net.Config.ADCBits != 8 {
		t.Errorf("ADCBits = %d, want 8", net.Config.ADCBits)
	}
	if net.Config.DACBits != 8 {
		t.Errorf("DACBits = %d, want 8", net.Config.DACBits)
	}
	if net.Config.NoiseLevel != 0.01 {
		t.Errorf("NoiseLevel = %f, want 0.01", net.Config.NoiseLevel)
	}
	if net.Config.PerLayerQuant != false {
		t.Errorf("PerLayerQuant = %v, want false", net.Config.PerLayerQuant)
	}
	if net.Config.Layer1Levels != 30 {
		t.Errorf("Layer1Levels = %d, want 30", net.Config.Layer1Levels)
	}
	if net.Config.Layer2Levels != 30 {
		t.Errorf("Layer2Levels = %d, want 30", net.Config.Layer2Levels)
	}
}

// TestScanAvailableQATLevels_SkipsPTQFile verifies PTQ file is correctly ignored.
func TestScanAvailableQATLevels_SkipsPTQFile(t *testing.T) {
	// Create temporary directory with test weight files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"pretrained_weights.json",      // 30 levels (default)
		"pretrained_weights_20.json",   // 20 levels
		"pretrained_weights_ptq.json",  // Should be ignored (PTQ file)
		"other_file.json",              // Should be ignored (wrong pattern)
	}

	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	levels := ScanAvailableQATLevels(tmpDir)

	// Should find only 30, 20 (sorted), PTQ file should be ignored
	expected := []int{20, 30}
	if len(levels) != len(expected) {
		t.Errorf("ScanAvailableQATLevels found %d levels, want %d (PTQ should be ignored)", len(levels), len(expected))
	}

	for i, want := range expected {
		if i >= len(levels) || levels[i] != want {
			t.Errorf("levels[%d] = %d, want %d", i, levels[i], want)
		}
	}
}

// TestGetBestMatchingWeightsLevel_TieBreaking verifies tie-breaking behavior.
func TestGetBestMatchingWeightsLevel_TieBreaking(t *testing.T) {
	// Create temporary directory with test weight files
	tmpDir := t.TempDir()

	testFiles := []string{
		"pretrained_weights.json",    // 30 levels
		"pretrained_weights_20.json", // 20 levels
	}

	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test case where distance is equal: target=25 is equidistant from 20 and 30
	got := GetBestMatchingWeightsLevel(tmpDir, 25)

	// Should return one of them (implementation-defined, but should be consistent)
	if got != 20 && got != 30 {
		t.Errorf("GetBestMatchingWeightsLevel(25) = %d, want 20 or 30", got)
	}

	// Verify distance is indeed equal
	dist20 := math.Abs(float64(25 - 20))
	dist30 := math.Abs(float64(25 - 30))
	if dist20 != dist30 {
		t.Errorf("Test setup error: distances should be equal, got %f and %f", dist20, dist30)
	}
}

// TestLoadWeights_UpdatesDimensions verifies LoadWeights updates network dimensions.
func TestLoadWeights_UpdatesDimensions(t *testing.T) {
	// Create temporary weights file with different dimensions
	tmpDir := t.TempDir()
	weightsPath := filepath.Join(tmpDir, "test_weights.json")

	// Create weights with 2 inputs, 3 hidden, 2 outputs
	weightsJSON := `{
		"layer1_weights": [[0.1, 0.2], [0.3, 0.4], [0.5, 0.6]],
		"layer2_weights": [[0.7, 0.8, 0.9], [1.0, 1.1, 1.2]],
		"biases1": [0.1, 0.2, 0.3],
		"biases2": [0.4, 0.5],
		"l1_scale": 1.0,
		"l1_offset": 0.0,
		"l2_scale": 1.0,
		"l2_offset": 0.0
	}`

	if err := os.WriteFile(weightsPath, []byte(weightsJSON), 0644); err != nil {
		t.Fatalf("Failed to create test weights file: %v", err)
	}

	// Create network with different dimensions
	net := NewDualModeNetwork(100, 50, 5)

	// Load weights should update dimensions
	err := net.LoadWeights(weightsPath)
	if err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify dimensions were updated to match loaded weights
	if net.InputSize != 2 {
		t.Errorf("After LoadWeights, InputSize = %d, want 2", net.InputSize)
	}
	if net.HiddenSize != 3 {
		t.Errorf("After LoadWeights, HiddenSize = %d, want 3", net.HiddenSize)
	}
	if net.OutputSize != 2 {
		t.Errorf("After LoadWeights, OutputSize = %d, want 2", net.OutputSize)
	}

	// Verify weight values were loaded correctly
	if math.Abs(net.FPWeights1[0][0]-0.1) > 1e-6 {
		t.Errorf("FPWeights1[0][0] = %f, want 0.1", net.FPWeights1[0][0])
	}
	if math.Abs(net.FPWeights1[2][1]-0.6) > 1e-6 {
		t.Errorf("FPWeights1[2][1] = %f, want 0.6", net.FPWeights1[2][1])
	}
}
