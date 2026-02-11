package crossbar

import (
	"math"
	"testing"
)

// ============================================================================
// TemperatureProfile Tests
// ============================================================================

func TestDefaultTemperatureProfile(t *testing.T) {
	profile := DefaultTemperatureProfile()

	if profile == nil {
		t.Fatal("DefaultTemperatureProfile returned nil")
	}

	if !profile.Enable {
		t.Error("Expected Enable to be true")
	}
	if !profile.ApplyConductanceWindow {
		t.Error("Expected ApplyConductanceWindow to be true")
	}
	if !profile.ApplyVariationNoise {
		t.Error("Expected ApplyVariationNoise to be true")
	}
	if !profile.ApplyDrift {
		t.Error("Expected ApplyDrift to be true")
	}
}

func TestTemperatureProfile_NilIsLegacy(t *testing.T) {
	// Create array and options with nil TemperatureProfile
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize with simple pattern
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	opts := &MVMOptions{
		EnableIRDrop:       true,
		EnableSneakPaths:   false,
		EnableVariation:    false,
		EnableDrift:        false,
		Temperature:        350.0, // Non-default temperature
		TemperatureProfile: nil,   // Legacy behavior
	}

	input := []float64{1.0, 1.0, 1.0, 1.0}
	result, err := arr.MVMWithNonIdealities(input, opts)
	if err != nil {
		t.Fatalf("MVM with nil profile failed: %v", err)
	}

	// With nil profile, only wire resistance scaling should apply
	// Result should succeed without additional temperature effects
	if result == nil {
		t.Error("Expected valid result with nil TemperatureProfile")
	}
	if len(result.ActualOutput) != 4 {
		t.Errorf("Expected 4 outputs, got %d", len(result.ActualOutput))
	}
}

// ============================================================================
// scaleAroundMidpoint01 Tests
// ============================================================================

func TestScaleAroundMidpoint01_Identity(t *testing.T) {
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, x := range testCases {
		result := scaleAroundMidpoint01(x, 1.0)
		if math.Abs(result-x) > 1e-10 {
			t.Errorf("Identity scaling failed: scaleAroundMidpoint01(%.2f, 1.0) = %.6f, want %.6f",
				x, result, x)
		}
	}
}

func TestScaleAroundMidpoint01_Midpoint(t *testing.T) {
	factors := []float64{0.0, 0.5, 1.0, 2.0, 10.0}

	for _, factor := range factors {
		result := scaleAroundMidpoint01(0.5, factor)
		if math.Abs(result-0.5) > 1e-10 {
			t.Errorf("Midpoint scaling failed: scaleAroundMidpoint01(0.5, %.1f) = %.6f, want 0.5",
				factor, result)
		}
	}
}

func TestScaleAroundMidpoint01_Widen(t *testing.T) {
	// factor > 1 should move values away from 0.5
	testCases := []struct {
		x      float64
		factor float64
	}{
		{0.3, 2.0}, // Below midpoint: should move toward 0
		{0.7, 2.0}, // Above midpoint: should move toward 1
		{0.4, 1.5},
		{0.6, 1.5},
	}

	for _, tc := range testCases {
		result := scaleAroundMidpoint01(tc.x, tc.factor)
		distance := math.Abs(result - 0.5)
		originalDistance := math.Abs(tc.x - 0.5)

		if distance <= originalDistance {
			t.Errorf("Widen failed: scaleAroundMidpoint01(%.1f, %.1f) = %.6f, "+
				"distance %.6f not greater than original %.6f",
				tc.x, tc.factor, result, distance, originalDistance)
		}
	}
}

func TestScaleAroundMidpoint01_Narrow(t *testing.T) {
	// factor < 1 should move values toward 0.5
	testCases := []struct {
		x      float64
		factor float64
	}{
		{0.2, 0.5}, // Below midpoint: should move toward 0.5
		{0.8, 0.5}, // Above midpoint: should move toward 0.5
		{0.1, 0.3},
		{0.9, 0.3},
	}

	for _, tc := range testCases {
		result := scaleAroundMidpoint01(tc.x, tc.factor)
		distance := math.Abs(result - 0.5)
		originalDistance := math.Abs(tc.x - 0.5)

		if distance >= originalDistance {
			t.Errorf("Narrow failed: scaleAroundMidpoint01(%.1f, %.1f) = %.6f, "+
				"distance %.6f not less than original %.6f",
				tc.x, tc.factor, result, distance, originalDistance)
		}
	}
}

func TestScaleAroundMidpoint01_Clamping(t *testing.T) {
	testCases := []struct {
		x      float64
		factor float64
	}{
		{0.0, 2.0},   // Should stay at 0
		{1.0, 2.0},   // Should stay at 1
		{0.1, 10.0},  // Large factor, should clamp to 0
		{0.9, 10.0},  // Large factor, should clamp to 1
		{-0.5, 1.0},  // Negative input, should clamp to range first
		{1.5, 1.0},   // Out of range input, should clamp to range first
	}

	for _, tc := range testCases {
		result := scaleAroundMidpoint01(tc.x, tc.factor)
		if result < 0.0 || result > 1.0 {
			t.Errorf("Clamping failed: scaleAroundMidpoint01(%.1f, %.1f) = %.6f, "+
				"expected value in [0, 1]", tc.x, tc.factor, result)
		}
	}
}

func TestScaleAroundMidpoint01_ZeroFactor(t *testing.T) {
	testCases := []float64{0.0, 0.3, 0.5, 0.7, 1.0}

	for _, x := range testCases {
		result := scaleAroundMidpoint01(x, 0.0)
		if math.Abs(result-0.5) > 1e-10 {
			t.Errorf("Zero factor failed: scaleAroundMidpoint01(%.1f, 0.0) = %.6f, want 0.5",
				x, result)
		}
	}
}

// ============================================================================
// MVMOptions Tests
// ============================================================================

func TestDefaultMVMOptions(t *testing.T) {
	opts := DefaultMVMOptions()

	if opts == nil {
		t.Fatal("DefaultMVMOptions returned nil")
	}

	if !opts.EnableIRDrop {
		t.Error("Expected EnableIRDrop to be true")
	}
	if !opts.EnableSneakPaths {
		t.Error("Expected EnableSneakPaths to be true")
	}
	if !opts.EnableVariation {
		t.Error("Expected EnableVariation to be true")
	}
	if opts.EnableDrift {
		t.Error("Expected EnableDrift to be false (drift simulated separately)")
	}
	if opts.Temperature != 300.0 {
		t.Errorf("Expected Temperature = 300.0K, got %.1f", opts.Temperature)
	}
	if opts.Architecture != "0T1R" {
		t.Errorf("Expected Architecture = '0T1R', got '%s'", opts.Architecture)
	}
	if opts.TemperatureProfile != nil {
		t.Error("Expected TemperatureProfile to be nil (legacy behavior)")
	}
}

func TestMVMOptions_Is1T1R(t *testing.T) {
	testCases := []struct {
		arch     string
		expected bool
	}{
		{"1T1R", true},
		{"0T1R", false},
		{"2T1R", false},
		{"1T1R-Active", true},
		{"Transistor", true},
		{"", false},
		{"Passive", false},
	}

	for _, tc := range testCases {
		opts := &MVMOptions{Architecture: tc.arch}
		result := opts.Is1T1R()
		if result != tc.expected {
			t.Errorf("Is1T1R('%s') = %v, want %v", tc.arch, result, tc.expected)
		}
	}
}

func TestMVMOptions_Is1T1R_Nil(t *testing.T) {
	var opts *MVMOptions
	if opts.Is1T1R() {
		t.Error("Nil MVMOptions should return false for Is1T1R()")
	}
}

func TestMVMOptions_Is2T1R(t *testing.T) {
	testCases := []struct {
		arch     string
		expected bool
	}{
		{"2T1R", true},
		{"1T1R", false},
		{"0T1R", false},
		{"2T1R-Advanced", true},
		{"Dual", true},
		{"", false},
		{"Passive", false},
	}

	for _, tc := range testCases {
		opts := &MVMOptions{Architecture: tc.arch}
		result := opts.Is2T1R()
		if result != tc.expected {
			t.Errorf("Is2T1R('%s') = %v, want %v", tc.arch, result, tc.expected)
		}
	}
}

func TestMVMOptions_Is2T1R_Nil(t *testing.T) {
	var opts *MVMOptions
	if opts.Is2T1R() {
		t.Error("Nil MVMOptions should return false for Is2T1R()")
	}
}

func TestMVMOptions_HasTransistorIsolation(t *testing.T) {
	testCases := []struct {
		arch     string
		expected bool
	}{
		{"1T1R", true},
		{"2T1R", true},
		{"0T1R", false},
		{"Passive", false},
		{"", false},
	}

	for _, tc := range testCases {
		opts := &MVMOptions{Architecture: tc.arch}
		result := opts.HasTransistorIsolation()
		if result != tc.expected {
			t.Errorf("HasTransistorIsolation('%s') = %v, want %v",
				tc.arch, result, tc.expected)
		}
	}
}

func TestMVMOptions_HasTransistorIsolation_Nil(t *testing.T) {
	var opts *MVMOptions
	if opts.HasTransistorIsolation() {
		t.Error("Nil MVMOptions should return false for HasTransistorIsolation()")
	}
}

// ============================================================================
// MVMWithNonIdealities Tests
// ============================================================================

func TestMVMWithNonIdealities_NilOpts(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize with simple pattern
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	input := []float64{1.0, 1.0, 1.0, 1.0}
	result, err := arr.MVMWithNonIdealities(input, nil)

	if err != nil {
		t.Fatalf("MVMWithNonIdealities with nil opts failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(result.IdealOutput) != 4 {
		t.Errorf("Expected 4 ideal outputs, got %d", len(result.IdealOutput))
	}
	if len(result.ActualOutput) != 4 {
		t.Errorf("Expected 4 actual outputs, got %d", len(result.ActualOutput))
	}
}

func TestMVMWithNonIdealities_InputTooLarge(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	input := []float64{1.0, 1.0, 1.0, 1.0, 1.0} // 5 elements for 4 columns
	_, err = arr.MVMWithNonIdealities(input, nil)

	if err == nil {
		t.Error("Expected error for input size exceeding columns")
	}
}

func TestMVMWithNonIdealities_BasicResult(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize with known pattern
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			weight := float64(i+j) / 6.0 // Values 0 to 1
			arr.ProgramWeight(i, j, weight)
		}
	}

	opts := &MVMOptions{
		EnableIRDrop:     true,
		EnableSneakPaths: true,
		EnableVariation:  false,
		EnableDrift:      false,
		Temperature:      300.0,
		Architecture:     "0T1R",
	}

	input := []float64{1.0, 0.5, 0.5, 1.0}
	result, err := arr.MVMWithNonIdealities(input, opts)

	if err != nil {
		t.Fatalf("MVMWithNonIdealities failed: %v", err)
	}

	// Verify result structure
	if len(result.IdealOutput) != 4 {
		t.Errorf("Expected 4 ideal outputs, got %d", len(result.IdealOutput))
	}
	if len(result.ActualOutput) != 4 {
		t.Errorf("Expected 4 actual outputs, got %d", len(result.ActualOutput))
	}

	// Verify MAC operations
	expectedMACs := 4 * 4
	if result.MACOperations != expectedMACs {
		t.Errorf("Expected %d MAC operations, got %d", expectedMACs, result.MACOperations)
	}
}

func TestMVMWithNonIdealities_EnergyMetrics(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    6,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	result, err := arr.MVMWithNonIdealities(input, nil)
	if err != nil {
		t.Fatalf("MVMWithNonIdealities failed: %v", err)
	}

	// Verify energy fields are positive
	if result.ArrayEnergy <= 0 {
		t.Errorf("Expected positive ArrayEnergy, got %.6f", result.ArrayEnergy)
	}
	if result.ADCEnergy <= 0 {
		t.Errorf("Expected positive ADCEnergy, got %.6f", result.ADCEnergy)
	}
	if result.DACEnergy <= 0 {
		t.Errorf("Expected positive DACEnergy, got %.6f", result.DACEnergy)
	}
	if result.TotalEnergy <= 0 {
		t.Errorf("Expected positive TotalEnergy, got %.6f", result.TotalEnergy)
	}

	// Verify total energy is sum of components
	expectedTotal := result.ArrayEnergy + result.ADCEnergy + result.DACEnergy
	if math.Abs(result.TotalEnergy-expectedTotal) > 1e-6 {
		t.Errorf("TotalEnergy mismatch: got %.6f, expected %.6f",
			result.TotalEnergy, expectedTotal)
	}

	// Verify GPU comparison
	if result.GPUEquivalentEnergy <= 0 {
		t.Errorf("Expected positive GPUEquivalentEnergy, got %.6f",
			result.GPUEquivalentEnergy)
	}
	if result.EnergyEfficiency <= 0 {
		t.Errorf("Expected positive EnergyEfficiency, got %.6f",
			result.EnergyEfficiency)
	}
}

func TestMVMWithNonIdealities_NoNonIdealities(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    10, // High precision ADC
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize with uniform weights
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	opts := &MVMOptions{
		EnableIRDrop:     false,
		EnableSneakPaths: false,
		EnableVariation:  false,
		EnableDrift:      false,
		Temperature:      300.0,
		Architecture:     "0T1R",
	}

	input := []float64{1.0, 1.0, 1.0, 1.0}
	result, err := arr.MVMWithNonIdealities(input, opts)

	if err != nil {
		t.Fatalf("MVMWithNonIdealities failed: %v", err)
	}

	// With no non-idealities and high ADC precision,
	// ActualOutput should be very close to IdealOutput
	for i := range result.IdealOutput {
		diff := math.Abs(result.IdealOutput[i] - result.ActualOutput[i])
		// Allow for small ADC quantization error
		if diff > 0.01 {
			t.Errorf("Output[%d]: ideal=%.6f, actual=%.6f, diff=%.6f too large",
				i, result.IdealOutput[i], result.ActualOutput[i], diff)
		}
	}

	// Error metrics should be very small
	if result.RMSE > 0.01 {
		t.Errorf("Expected small RMSE without non-idealities, got %.6f", result.RMSE)
	}
}

func TestMVMWithNonIdealities_ArchitectureComparison(t *testing.T) {
	cfg := &Config{
		Rows:       16,
		Cols:       16,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}

	architectures := []string{"0T1R", "1T1R", "2T1R"}
	results := make([]*MVMResult, len(architectures))

	input := make([]float64, 16)
	for i := range input {
		input[i] = 1.0
	}

	for idx, arch := range architectures {
		arr, err := NewArray(cfg)
		if err != nil {
			t.Fatalf("NewArray failed: %v", err)
		}
		// Initialize with same pattern
		for i := 0; i < 16; i++ {
			for j := 0; j < 16; j++ {
				arr.ProgramWeight(i, j, 0.6)
			}
		}

		opts := &MVMOptions{
			EnableIRDrop:     true,
			EnableSneakPaths: true,
			EnableVariation:  false,
			EnableDrift:      false,
			Temperature:      300.0,
			Architecture:     arch,
		}

		result, err := arr.MVMWithNonIdealities(input, opts)
		if err != nil {
			t.Fatalf("MVMWithNonIdealities for %s failed: %v", arch, err)
		}
		results[idx] = result
	}

	// 0T1R should have highest errors due to sneak paths
	// 1T1R should have medium errors
	// 2T1R should have lowest errors
	if results[0].RMSE < results[1].RMSE {
		t.Logf("Note: 0T1R RMSE (%.6f) unexpectedly lower than 1T1R (%.6f)",
			results[0].RMSE, results[1].RMSE)
	}
	if results[1].RMSE < results[2].RMSE {
		t.Logf("Note: 1T1R RMSE (%.6f) unexpectedly lower than 2T1R (%.6f)",
			results[1].RMSE, results[2].RMSE)
	}

	t.Logf("Architecture comparison:")
	for idx, arch := range architectures {
		t.Logf("  %s: RMSE=%.6f, MaxError=%.6f", arch,
			results[idx].RMSE, results[idx].MaxError)
	}
}

func TestMVMWithNonIdealities_TemperatureEffect(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}

	temperatures := []float64{250.0, 300.0, 350.0, 400.0}
	results := make([]*MVMResult, len(temperatures))

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	for idx, temp := range temperatures {
		arr, err := NewArray(cfg)
		if err != nil {
			t.Fatalf("NewArray failed: %v", err)
		}
		// Initialize with same pattern
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				arr.ProgramWeight(i, j, 0.5)
			}
		}

		opts := &MVMOptions{
			EnableIRDrop:     true,
			EnableSneakPaths: false,
			EnableVariation:  false,
			EnableDrift:      false,
			Temperature:      temp,
			Architecture:     "0T1R",
		}

		result, err := arr.MVMWithNonIdealities(input, opts)
		if err != nil {
			t.Fatalf("MVMWithNonIdealities at %.1fK failed: %v", temp, err)
		}
		results[idx] = result
	}

	t.Logf("Temperature effect on IR drop:")
	for idx, temp := range temperatures {
		irMax := 0.0
		if results[idx].IRDropAnalysis != nil {
			irMax = results[idx].IRDropAnalysis.MaxIRDrop
		}
		t.Logf("  %.1fK: MaxIRDrop=%.6f", temp, irMax)
	}
}

func TestMVMWithNonIdealities_WithTemperatureProfile(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	profile := DefaultTemperatureProfile()
	opts := &MVMOptions{
		EnableIRDrop:       true,
		EnableSneakPaths:   false,
		EnableVariation:    true,
		EnableDrift:        true,
		Temperature:        350.0,
		Architecture:       "0T1R",
		TemperatureProfile: profile,
	}

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	result, err := arr.MVMWithNonIdealities(input, opts)
	if err != nil {
		t.Fatalf("MVMWithNonIdealities with temperature profile failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Result should include temperature-dependent effects
	if len(result.ActualOutput) != 8 {
		t.Errorf("Expected 8 outputs, got %d", len(result.ActualOutput))
	}
}

// ============================================================================
// Additional Function Tests
// ============================================================================

func TestComputeFullMVMSneak(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize with varying conductances
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			arr.ProgramWeight(i, j, float64(i+j)/14.0)
		}
	}

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	t.Run("0T1R", func(t *testing.T) {
		opts := &MVMOptions{Architecture: "0T1R"}
		sneakPerRow := arr.ComputeFullMVMSneak(input, opts)

		if len(sneakPerRow) != 8 {
			t.Errorf("Expected 8 sneak values, got %d", len(sneakPerRow))
		}

		// For 0T1R, should have non-zero sneak paths
		hasNonZero := false
		for i, sneak := range sneakPerRow {
			if sneak > 0 {
				hasNonZero = true
			}
			if sneak < 0 {
				t.Errorf("Row %d has negative sneak current: %.6f", i, sneak)
			}
		}
		if !hasNonZero {
			t.Error("Expected non-zero sneak paths for 0T1R")
		}
	})

	t.Run("1T1R", func(t *testing.T) {
		opts := &MVMOptions{Architecture: "1T1R"}
		sneakPerRow := arr.ComputeFullMVMSneak(input, opts)

		if len(sneakPerRow) != 8 {
			t.Errorf("Expected 8 sneak values, got %d", len(sneakPerRow))
		}

		// For 1T1R, sneak should be negligible (all zeros)
		for i, sneak := range sneakPerRow {
			if sneak != 0 {
				t.Errorf("Row %d has non-zero sneak for 1T1R: %.6f", i, sneak)
			}
		}
	})

	t.Run("2T1R", func(t *testing.T) {
		opts := &MVMOptions{Architecture: "2T1R"}
		sneakPerRow := arr.ComputeFullMVMSneak(input, opts)

		if len(sneakPerRow) != 8 {
			t.Errorf("Expected 8 sneak values, got %d", len(sneakPerRow))
		}

		// For 2T1R, sneak should be negligible (all zeros)
		for i, sneak := range sneakPerRow {
			if sneak != 0 {
				t.Errorf("Row %d has non-zero sneak for 2T1R: %.6f", i, sneak)
			}
		}
	})
}

func TestAnalyzeSneakContributions(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize with pattern that will create sneak paths
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			arr.ProgramWeight(i, j, 0.8) // High conductance
		}
	}

	input := []float64{1.0, 1.0, 1.0, 1.0}
	contributions := arr.AnalyzeSneakContributions(1, input, 5)

	// Should return up to 5 contributions
	if len(contributions) > 5 {
		t.Errorf("Expected max 5 contributions, got %d", len(contributions))
	}

	// Each contribution should have valid fields
	for i, contrib := range contributions {
		if contrib.PathG < 0 {
			t.Errorf("Contribution[%d]: negative PathG %.6f", i, contrib.PathG)
		}
		if contrib.PathCurrent < 0 {
			t.Errorf("Contribution[%d]: negative PathCurrent %.6f", i, contrib.PathCurrent)
		}
		if contrib.SourceRow < 0 || contrib.SourceRow >= 4 {
			t.Errorf("Contribution[%d]: invalid SourceRow %d", i, contrib.SourceRow)
		}
	}

	// Contributions should be sorted by current (descending)
	for i := 0; i < len(contributions)-1; i++ {
		if contributions[i].PathCurrent < contributions[i+1].PathCurrent {
			t.Errorf("Contributions not sorted: [%d]=%.6f < [%d]=%.6f",
				i, contributions[i].PathCurrent,
				i+1, contributions[i+1].PathCurrent)
		}
	}
}

func TestComputeAccuracyDegradation(t *testing.T) {
	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.05,
		ADCBits:    6,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Initialize
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	baselineAccuracy := 95.0
	degradation, err := arr.ComputeAccuracyDegradation(input, baselineAccuracy)

	if err != nil {
		t.Fatalf("ComputeAccuracyDegradation failed: %v", err)
	}

	if degradation.BaselineAccuracy != baselineAccuracy {
		t.Errorf("BaselineAccuracy mismatch: got %.2f, want %.2f",
			degradation.BaselineAccuracy, baselineAccuracy)
	}

	// Should have 4 degradation steps
	expectedSteps := 4
	if len(degradation.Degradations) != expectedSteps {
		t.Errorf("Expected %d degradation steps, got %d",
			expectedSteps, len(degradation.Degradations))
	}

	// Verify step names
	expectedSources := []string{
		"ADC/DAC Quantization",
		"IR Drop",
		"Device Variation",
		"Sneak Paths",
	}
	for i, expected := range expectedSources {
		if i >= len(degradation.Degradations) {
			break
		}
		if degradation.Degradations[i].Source != expected {
			t.Errorf("Step[%d]: expected source '%s', got '%s'",
				i, expected, degradation.Degradations[i].Source)
		}
	}

	// Final accuracy should be <= baseline
	if degradation.FinalAccuracy > baselineAccuracy {
		t.Errorf("FinalAccuracy (%.2f) > BaselineAccuracy (%.2f)",
			degradation.FinalAccuracy, baselineAccuracy)
	}

	t.Logf("Accuracy degradation:")
	t.Logf("  Baseline: %.2f%%", degradation.BaselineAccuracy)
	for _, step := range degradation.Degradations {
		t.Logf("  %s: -%.2f%% → %.2f%%", step.Source, step.Loss, step.AccuracyNow)
	}
	t.Logf("  Final: %.2f%%", degradation.FinalAccuracy)
}
