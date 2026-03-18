package crossbar

import (
	"encoding/json"
	"math"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// TestExportToMemTorch_Conversion verifies the R = 1/G relationship holds
// for the exported MemTorch parameters.
func TestExportToMemTorch_Conversion(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.05,
	}

	params := ExportToMemTorch(cfg)

	// R_ON = 1/G_MAX, R_OFF = 1/G_MIN
	expectedROn := 1.0 / physics.GMax
	expectedROff := 1.0 / physics.GMin

	tolerance := 1e-6

	if math.Abs(params.ROn-expectedROn) > tolerance {
		t.Errorf("R_ON: expected %.6f, got %.6f (should be 1/GMax)", expectedROn, params.ROn)
	}
	if math.Abs(params.ROff-expectedROff) > tolerance {
		t.Errorf("R_OFF: expected %.6f, got %.6f (should be 1/GMin)", expectedROff, params.ROff)
	}

	// R_ON should be less than R_OFF (low resistance = high conductance)
	if params.ROn >= params.ROff {
		t.Errorf("R_ON (%.2f) should be less than R_OFF (%.2f)", params.ROn, params.ROff)
	}

	// RON/ROFF ratio should equal GMax/GMin ratio
	rRatio := params.ROff / params.ROn
	gRatio := physics.GMax / physics.GMin
	if math.Abs(rRatio-gRatio) > tolerance {
		t.Errorf("R_OFF/R_ON ratio (%.2f) should equal G_MAX/G_MIN ratio (%.2f)", rRatio, gRatio)
	}

	// Sigma values should be positive when noise > 0
	if params.ROnSigma <= 0 {
		t.Error("R_ON sigma should be positive when NoiseLevel > 0")
	}
	if params.ROffSigma <= 0 {
		t.Error("R_OFF sigma should be positive when NoiseLevel > 0")
	}

	// R_OFF sigma should be larger than R_ON sigma (error propagation: sigma_R ~ 1/G^2)
	// Since G_MIN < G_MAX, R_OFF (1/G_MIN) has larger sigma
	if params.ROffSigma <= params.ROnSigma {
		t.Errorf("R_OFF sigma (%.6f) should be larger than R_ON sigma (%.6f) due to error propagation",
			params.ROffSigma, params.ROnSigma)
	}

	// Programming noise should match config noise level
	if math.Abs(params.PNoiseStd-cfg.NoiseLevel) > tolerance {
		t.Errorf("PNoiseStd: expected %.4f, got %.4f", cfg.NoiseLevel, params.PNoiseStd)
	}
}

// TestExportToMemTorch_ZeroNoise verifies that zero noise produces zero sigma.
func TestExportToMemTorch_ZeroNoise(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
	}

	params := ExportToMemTorch(cfg)

	if params.ROnSigma != 0 {
		t.Errorf("R_ON sigma should be 0 for zero noise, got: %f", params.ROnSigma)
	}
	if params.ROffSigma != 0 {
		t.Errorf("R_OFF sigma should be 0 for zero noise, got: %f", params.ROffSigma)
	}
	if params.PNoiseStd != 0 {
		t.Errorf("PNoiseStd should be 0 for zero noise, got: %f", params.PNoiseStd)
	}
}

// TestImportFromMemTorch_RoundTrip verifies that export followed by import
// recovers the original configuration parameters.
func TestImportFromMemTorch_RoundTrip(t *testing.T) {
	noises := []float64{0.0, 0.01, 0.05, 0.1, 0.2}

	for _, noise := range noises {
		cfg := &Config{
			Rows:       8,
			Cols:       8,
			NoiseLevel: noise,
		}

		exported := ExportToMemTorch(cfg)
		imported := ImportFromMemTorch(exported)

		// NoiseLevel should survive the round-trip
		tolerance := 1e-10
		if math.Abs(imported.NoiseLevel-cfg.NoiseLevel) > tolerance {
			t.Errorf("NoiseLevel %.4f did not survive round-trip: got %.4f (delta=%.2e)",
				cfg.NoiseLevel, imported.NoiseLevel, math.Abs(imported.NoiseLevel-cfg.NoiseLevel))
		}
	}
}

// TestImportFromMemTorch_InvalidParams verifies handling of invalid parameters.
func TestImportFromMemTorch_InvalidParams(t *testing.T) {
	// Zero resistance should return zero config
	params := MemTorchDeviceParams{
		ROn:  0,
		ROff: 0,
	}

	cfg := ImportFromMemTorch(params)
	if cfg.NoiseLevel != 0 {
		t.Errorf("Zero resistance should give zero noise, got: %f", cfg.NoiseLevel)
	}

	// Negative resistance should return zero config
	params = MemTorchDeviceParams{
		ROn:  -1000,
		ROff: -100000,
	}

	cfg = ImportFromMemTorch(params)
	if cfg.NoiseLevel != 0 {
		t.Errorf("Negative resistance should give zero noise, got: %f", cfg.NoiseLevel)
	}
}

// TestImportFromMemTorch_KnownValues verifies import from known MemTorch values.
func TestImportFromMemTorch_KnownValues(t *testing.T) {
	// 10 kOhm ON, 100 kOhm OFF => G_MAX = 100 µS, G_MIN = 10 µS
	params := MemTorchDeviceParams{
		ROn:       10000,  // 10 kOhm
		ROff:      100000, // 100 kOhm
		ROnSigma:  0,
		ROffSigma: 0,
		PNoiseStd: 0,
	}

	cfg := ImportFromMemTorch(params)

	// The imported config should reconstruct a valid array config
	if cfg.NoiseLevel != 0 {
		t.Errorf("Zero sigma should give zero NoiseLevel, got: %f", cfg.NoiseLevel)
	}
}

// TestExportWeightsAsMemTorchJSON_ValidJSON verifies the output is valid JSON
// with weights preserved.
func TestExportWeightsAsMemTorchJSON_ValidJSON(t *testing.T) {
	weights := [][]float64{
		{0.0, 0.25, 0.5},
		{0.75, 1.0, 0.3},
	}

	params := MemTorchDeviceParams{
		ROn:       10000,
		ROff:      100000,
		ROnSigma:  100,
		ROffSigma: 10000,
		PNoiseStd: 0.05,
	}

	data, err := ExportWeightsAsMemTorchJSON(weights, params)
	if err != nil {
		t.Fatalf("ExportWeightsAsMemTorchJSON failed: %v", err)
	}

	// Must be valid JSON
	var parsed MemTorchWeightExport
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Dimensions must match
	if parsed.Rows != 2 {
		t.Errorf("Expected 2 rows, got %d", parsed.Rows)
	}
	if parsed.Cols != 3 {
		t.Errorf("Expected 3 cols, got %d", parsed.Cols)
	}

	// Weights must be preserved
	if len(parsed.Weights) != 2 {
		t.Fatalf("Expected 2 weight rows, got %d", len(parsed.Weights))
	}
	for i, row := range weights {
		if len(parsed.Weights[i]) != len(row) {
			t.Errorf("Row %d: expected %d cols, got %d", i, len(row), len(parsed.Weights[i]))
			continue
		}
		for j, w := range row {
			if math.Abs(parsed.Weights[i][j]-w) > 1e-15 {
				t.Errorf("Weight[%d][%d]: expected %f, got %f", i, j, w, parsed.Weights[i][j])
			}
		}
	}

	// Device params must be preserved
	if math.Abs(parsed.DeviceParams.ROn-params.ROn) > 1e-6 {
		t.Errorf("R_ON not preserved: expected %f, got %f", params.ROn, parsed.DeviceParams.ROn)
	}
	if math.Abs(parsed.DeviceParams.ROff-params.ROff) > 1e-6 {
		t.Errorf("R_OFF not preserved: expected %f, got %f", params.ROff, parsed.DeviceParams.ROff)
	}
}

// TestExportWeightsAsMemTorchJSON_EmptyWeights verifies error on empty input.
func TestExportWeightsAsMemTorchJSON_EmptyWeights(t *testing.T) {
	params := MemTorchDeviceParams{ROn: 10000, ROff: 100000}

	_, err := ExportWeightsAsMemTorchJSON([][]float64{}, params)
	if err == nil {
		t.Error("Expected error for empty weight matrix")
	}
}

// TestExportWeightsAsMemTorchJSON_JaggedMatrix verifies error on jagged matrix.
func TestExportWeightsAsMemTorchJSON_JaggedMatrix(t *testing.T) {
	params := MemTorchDeviceParams{ROn: 10000, ROff: 100000}
	weights := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5}, // Different number of cols
	}

	_, err := ExportWeightsAsMemTorchJSON(weights, params)
	if err == nil {
		t.Error("Expected error for jagged weight matrix")
	}
}

// TestExportWeightsAsMemTorchJSON_OutOfRange verifies error on out-of-range weights.
func TestExportWeightsAsMemTorchJSON_OutOfRange(t *testing.T) {
	params := MemTorchDeviceParams{ROn: 10000, ROff: 100000}

	testCases := []struct {
		name    string
		weights [][]float64
	}{
		{"negative weight", [][]float64{{-0.1, 0.5}}},
		{"weight > 1", [][]float64{{0.5, 1.1}}},
		{"NaN weight", [][]float64{{math.NaN(), 0.5}}},
		{"Inf weight", [][]float64{{math.Inf(1), 0.5}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ExportWeightsAsMemTorchJSON(tc.weights, params)
			if err == nil {
				t.Errorf("Expected error for %s", tc.name)
			}
		})
	}
}

// TestExportWeightsAsMemTorchJSON_LargeMatrix verifies export of a larger matrix.
func TestExportWeightsAsMemTorchJSON_LargeMatrix(t *testing.T) {
	params := ExportToMemTorch(&Config{NoiseLevel: 0.05})

	// 32x32 weight matrix
	rows, cols := 32, 32
	weights := make([][]float64, rows)
	for i := range weights {
		weights[i] = make([]float64, cols)
		for j := range weights[i] {
			weights[i][j] = float64(i*cols+j) / float64(rows*cols)
		}
	}

	data, err := ExportWeightsAsMemTorchJSON(weights, params)
	if err != nil {
		t.Fatalf("Failed to export 32x32 matrix: %v", err)
	}

	// Verify valid JSON
	var parsed MemTorchWeightExport
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Large matrix output is not valid JSON: %v", err)
	}

	if parsed.Rows != rows || parsed.Cols != cols {
		t.Errorf("Dimensions mismatch: expected %dx%d, got %dx%d",
			rows, cols, parsed.Rows, parsed.Cols)
	}
}
