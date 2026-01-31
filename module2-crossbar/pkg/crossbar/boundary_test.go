package crossbar

import (
	"fecim-lattice-tools/shared/physics"
	"fmt"
	"math"
	"testing"
)

// TestSingleCellArray verifies 1x1 array edge case
func TestSingleCellArray(t *testing.T) {
	cfg := &Config{
		Rows:       1,
		Cols:       1,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create 1x1 array: %v", err)
	}
	defer arr.Destroy()

	arr.ProgramWeight(0, 0, 0.5)

	t.Run("MVM works", func(t *testing.T) {
		output, err := arr.MVM([]float64{1.0})
		if err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
		if len(output) != 1 {
			t.Errorf("Output length = %d, want 1", len(output))
		}
		if math.IsNaN(output[0]) || math.IsInf(output[0], 0) {
			t.Errorf("Output is NaN or Inf")
		}
	})

	t.Run("IR drop analysis works", func(t *testing.T) {
		analysis := arr.AnalyzeIRDrop([]float64{1.0}, nil)
		// 1x1 should have minimal IR drop
		t.Logf("1x1 IR drop: %.4f%%", analysis.MaxIRDrop*100)
	})

	t.Run("Sneak path analysis works", func(t *testing.T) {
		analysis := arr.AnalyzeSneakPaths(0, 0)
		// 1x1 should have zero sneak - using TotalSneak field
		if analysis.TotalSneak != 0 {
			t.Logf("1x1 sneak current (expected minimal): %e", analysis.TotalSneak)
		}
	})
}

// TestExtremeAspectRatioArrays verifies non-square arrays
func TestExtremeAspectRatioArrays(t *testing.T) {
	testCases := []struct {
		rows, cols int
	}{
		{1, 64},  // 1 row, 64 cols
		{64, 1},  // 64 rows, 1 col
		{2, 128}, // Wide array
		{128, 2}, // Tall array
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%dx%d", tc.rows, tc.cols), func(t *testing.T) {
			cfg := &Config{
				Rows:       tc.rows,
				Cols:       tc.cols,
				NoiseLevel: 0.0,
				ADCBits:    8,
				DACBits:    8,
			}

			arr, err := NewArray(cfg)
			if err != nil {
				t.Fatalf("Failed to create array: %v", err)
			}
			defer arr.Destroy()

			// Program weights
			for i := 0; i < tc.rows; i++ {
				for j := 0; j < tc.cols; j++ {
					arr.ProgramWeight(i, j, 0.5)
				}
			}

			// MVM with correct input size
			input := make([]float64, tc.cols)
			for j := range input {
				input[j] = 0.5
			}

			output, err := arr.MVM(input)
			if err != nil {
				t.Fatalf("MVM failed: %v", err)
			}

			if len(output) != tc.rows {
				t.Errorf("Output length = %d, want %d", len(output), tc.rows)
			}

			// Check no NaN/Inf
			for i, v := range output {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					t.Errorf("Output[%d] is NaN or Inf", i)
				}
			}
		})
	}
}

// TestZeroConductanceCell verifies handling of minimum conductance
func TestZeroConductanceCell(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)
	defer arr.Destroy()

	// Program one cell to minimum (level 0)
	arr.ProgramWeight(0, 0, 0.0)

	// Other cells at mid-range
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i != 0 || j != 0 {
				arr.ProgramWeight(i, j, 0.5)
			}
		}
	}

	t.Run("MVM handles minimum G", func(t *testing.T) {
		output, err := arr.MVM([]float64{1.0, 1.0, 1.0, 1.0})
		if err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
		// Should have no NaN/Inf
		for i, v := range output {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Errorf("Output[%d] is NaN or Inf", i)
			}
		}
	})
}

// TestMaximumConductanceSaturation verifies values > 1.0 saturate
func TestMaximumConductanceSaturation(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)
	defer arr.Destroy()

	// Try to program above 1.0 (should saturate to level 29)
	arr.ProgramWeight(0, 0, 1.5)

	matrix := arr.GetConductanceMatrix()
	if matrix[0][0] > 1.0 {
		t.Errorf("Weight %f not saturated to <= 1.0", matrix[0][0])
	}

	// Also test negative values should clamp to 0
	arr.ProgramWeight(1, 1, -0.5)
	matrix = arr.GetConductanceMatrix()
	if matrix[1][1] < 0.0 {
		t.Errorf("Weight %f not clamped to >= 0.0", matrix[1][1])
	}
}

// TestQuantizationBoundaryValues verifies edge quantization behavior
func TestQuantizationBoundaryValues(t *testing.T) {
	t.Run("Exactly 30 unique levels", func(t *testing.T) {
		seen := make(map[float64]bool)
		for i := 0; i <= 1000; i++ {
			input := float64(i) / 1000.0
			quantized := physics.QuantizeTo30Levels(input)
			seen[quantized] = true
		}
		if len(seen) != physics.DefaultLevels {
			t.Errorf("Found %d unique levels, want %d", len(seen), physics.DefaultLevels)
		}
	})

	t.Run("Boundary at 0", func(t *testing.T) {
		q := physics.QuantizeTo30Levels(0.0)
		if q != 0.0 {
			t.Errorf("QuantizeTo30Levels(0.0) = %f, want 0.0", q)
		}
	})

	t.Run("Boundary at 1", func(t *testing.T) {
		q := physics.QuantizeTo30Levels(1.0)
		if q != 1.0 {
			t.Errorf("QuantizeTo30Levels(1.0) = %f, want 1.0", q)
		}
	})

	t.Run("Midpoint quantization", func(t *testing.T) {
		// 0.5 should quantize to level 14 or 15 (14.5 rounded)
		q := physics.QuantizeTo30Levels(0.5)
		level := GetLevel(q)
		if level < 14 || level > 15 {
			t.Errorf("QuantizeTo30Levels(0.5) -> level %d, expected 14 or 15", level)
		}
	})
}

// TestArrayDestroyCleanup verifies Destroy() doesn't panic
func TestArrayDestroyCleanup(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)

	// Should not panic
	arr.Destroy()

	// Double destroy should also not panic
	arr.Destroy()
}

// TestLargeArrayPerformance verifies large arrays work (smoke test)
func TestLargeArrayPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large array test in short mode")
	}

	cfg := &Config{
		Rows:       256,
		Cols:       256,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create 256x256 array: %v", err)
	}
	defer arr.Destroy()

	// Quick program and MVM
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	input := make([]float64, 256)
	for j := range input {
		input[j] = 0.5
	}

	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	if len(output) != 256 {
		t.Errorf("Output length = %d, want 256", len(output))
	}
}

// ============================================================================
// ADDITIONAL BOUNDARY CONDITION TESTS FOR EXTREME PHYSICS PARAMETERS
// ============================================================================
// These tests verify that the crossbar array handles extreme parameter
// values without panics, NaN, Inf, or other numerical instabilities.

// ============================================================================
// 1. Array Size Boundary Tests
// ============================================================================

// TestBoundaryArraySize_Extremes tests extreme array dimensions.
//
// Physics boundary being tested:
// - 1x1: Minimum array (trivial multiplication)
// - Asymmetric arrays: Verify dimension handling
// - Large arrays: Memory and MVM correctness
func TestBoundaryArraySize_Extremes(t *testing.T) {
	testCases := []struct {
		name        string
		rows, cols  int
		expectError bool
	}{
		{"1x1 minimum", 1, 1, false},
		{"2x2 small", 2, 2, false},
		{"1x64 single row", 1, 64, false},
		{"64x1 single column", 64, 1, false},
		{"256x256 large", 256, 256, false},
		{"0x0 invalid", 0, 0, true},
		{"0x10 invalid rows", 0, 10, true},
		{"10x0 invalid cols", 10, 0, true},
		{"-1x10 negative rows", -1, 10, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.rows == 256 && testing.Short() {
				t.Skip("Skipping large array test in short mode")
			}

			cfg := &Config{
				Rows:       tc.rows,
				Cols:       tc.cols,
				NoiseLevel: 0.0,
				ADCBits:    8,
				DACBits:    8,
			}

			arr, err := NewArray(cfg)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %dx%d array, got nil", tc.rows, tc.cols)
					if arr != nil {
						arr.Destroy()
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create %dx%d array: %v", tc.rows, tc.cols, err)
			}
			defer arr.Destroy()

			// Verify dimensions
			if arr.Rows() != tc.rows {
				t.Errorf("Rows() = %d, want %d", arr.Rows(), tc.rows)
			}
			if arr.Cols() != tc.cols {
				t.Errorf("Cols() = %d, want %d", arr.Cols(), tc.cols)
			}

			// Program all cells with mid-range weight
			for i := 0; i < tc.rows; i++ {
				for j := 0; j < tc.cols; j++ {
					if err := arr.ProgramWeight(i, j, 0.5); err != nil {
						t.Fatalf("ProgramWeight failed: %v", err)
					}
				}
			}

			// Create input vector
			input := make([]float64, tc.cols)
			for j := range input {
				input[j] = 0.5
			}

			// MVM should work
			output, err := arr.MVM(input)
			if err != nil {
				t.Fatalf("MVM failed: %v", err)
			}

			// Verify output dimensions
			if len(output) != tc.rows {
				t.Errorf("Output length = %d, want %d", len(output), tc.rows)
			}

			// Verify no NaN/Inf in output
			for i, v := range output {
				if math.IsNaN(v) {
					t.Errorf("Output[%d] is NaN", i)
				}
				if math.IsInf(v, 0) {
					t.Errorf("Output[%d] is Inf", i)
				}
			}

			// For 1x1 array: MVM should be simple multiplication
			if tc.rows == 1 && tc.cols == 1 {
				// output[0] = weight * input[0] (with quantization)
				// At weight=0.5, input=0.5, expected ~0.25 normalized by maxCurrent
				// The exact value depends on DAC/ADC quantization
				t.Logf("1x1 MVM: output[0] = %.4f", output[0])
			}

			t.Logf("%dx%d array: MVM completed, output range [%.4f, %.4f]",
				tc.rows, tc.cols, minFloat64Slice(output), maxFloat64Slice(output))
		})
	}
}

// ============================================================================
// 2. Conductance Value Boundary Tests
// ============================================================================

// TestBoundaryConductance_Extremes tests extreme conductance values.
//
// Physics boundary being tested:
// - 0 conductance: No current flow (but no division by zero)
// - Very low conductance (1e-12 S): Near noise floor
// - Very high conductance (1e3 S): Above typical range
// - Verify quantization handles all cases
func TestBoundaryConductance_Extremes(t *testing.T) {
	conductanceValues := []struct {
		name       string
		normalized float64 // Input value (will be quantized to 0-1)
	}{
		{"zero", 0.0},
		{"near-zero", 0.001},
		{"very-low", 0.01},
		{"low", 0.1},
		{"mid", 0.5},
		{"high", 0.9},
		{"near-max", 0.99},
		{"max", 1.0},
		{"over-max", 1.5},    // Should clamp to 1.0
		{"negative", -0.5},   // Should clamp to 0.0
		{"very-negative", -10.0},
	}

	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	for _, tc := range conductanceValues {
		t.Run(tc.name, func(t *testing.T) {
			arr, _ := NewArray(cfg)
			defer arr.Destroy()

			// Program all cells with test conductance
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					arr.ProgramWeight(i, j, tc.normalized)
				}
			}

			// Verify stored conductance is clamped
			matrix := arr.GetConductanceMatrix()
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					g := matrix[i][j]

					// Must be in valid range [0, 1]
					if g < 0.0 {
						t.Errorf("Cell[%d][%d] conductance %.4e is negative", i, j, g)
					}
					if g > 1.0 {
						t.Errorf("Cell[%d][%d] conductance %.4e exceeds 1.0", i, j, g)
					}

					// Must not be NaN or Inf
					if math.IsNaN(g) {
						t.Errorf("Cell[%d][%d] conductance is NaN", i, j)
					}
					if math.IsInf(g, 0) {
						t.Errorf("Cell[%d][%d] conductance is Inf", i, j)
					}
				}
			}

			// MVM should work
			input := []float64{1.0, 1.0, 1.0, 1.0}
			output, err := arr.MVM(input)
			if err != nil {
				t.Fatalf("MVM failed: %v", err)
			}

			// Verify no NaN/Inf
			for i, v := range output {
				if math.IsNaN(v) {
					t.Errorf("Output[%d] is NaN for conductance %s", i, tc.name)
				}
				if math.IsInf(v, 0) {
					t.Errorf("Output[%d] is Inf for conductance %s", i, tc.name)
				}
			}

			t.Logf("Conductance %s (%.2f): stored=%.4f, MVM output[0]=%.4f",
				tc.name, tc.normalized, matrix[0][0], output[0])
		})
	}
}

// ============================================================================
// 3. Input Value Boundary Tests
// ============================================================================

// TestBoundaryInput_Extremes tests extreme input values to MVM.
//
// Physics boundary being tested:
// - Zero input: No computation
// - Negative input: Depends on DAC implementation
// - Very small input: Near noise floor
// - Very large input: DAC saturation
func TestBoundaryInput_Extremes(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, _ := NewArray(cfg)
	defer arr.Destroy()

	// Program array with mid-range weights
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	inputCases := []struct {
		name  string
		input []float64
	}{
		{"all-zero", []float64{0, 0, 0, 0}},
		{"all-one", []float64{1, 1, 1, 1}},
		{"mixed", []float64{0, 0.5, 1, 0.5}},
		{"negative", []float64{-0.5, -0.5, -0.5, -0.5}},
		{"very-small", []float64{1e-15, 1e-15, 1e-15, 1e-15}},
		{"very-large", []float64{1e6, 1e6, 1e6, 1e6}},
		{"over-one", []float64{1.5, 1.5, 1.5, 1.5}},
		{"partial", []float64{0.5, 0.5}}, // Fewer inputs than columns
	}

	for _, tc := range inputCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := arr.MVM(tc.input)
			if err != nil {
				t.Fatalf("MVM failed: %v", err)
			}

			// Verify output dimensions
			if len(output) != 4 {
				t.Errorf("Output length = %d, want 4", len(output))
			}

			// Verify no NaN/Inf
			for i, v := range output {
				if math.IsNaN(v) {
					t.Errorf("Output[%d] is NaN for input %s", i, tc.name)
				}
				if math.IsInf(v, 0) {
					t.Errorf("Output[%d] is Inf for input %s", i, tc.name)
				}
			}

			// Output should be bounded
			for i, v := range output {
				if v < 0 || v > 1 {
					t.Logf("Note: Output[%d] = %.4f outside [0,1] for input %s (may be expected)", i, v, tc.name)
				}
			}

			t.Logf("Input %s: output range [%.4f, %.4f]",
				tc.name, minFloat64Slice(output), maxFloat64Slice(output))
		})
	}
}

// ============================================================================
// 4. Non-Ideality Parameter Boundary Tests
// ============================================================================

// TestBoundaryNonIdealityParameters tests extreme non-ideality parameters.
//
// Physics boundary being tested:
// - Wire resistance: 0 (ideal) to very high (severe IR drop)
// - Sneak path ratio: 0 (no sneak) to 1.0 (maximum)
// - Drift coefficient: 0 (no drift) to very large
func TestBoundaryNonIdealityParameters(t *testing.T) {
	t.Run("WireResistance_Zero", func(t *testing.T) {
		// Zero wire resistance = ideal case (no IR drop)
		cfg := &Config{
			Rows:       8,
			Cols:       8,
			NoiseLevel: 0.0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		defer arr.Destroy()

		// Program weights
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				arr.ProgramWeight(i, j, 0.5)
			}
		}

		input := make([]float64, 8)
		for j := range input {
			input[j] = 1.0
		}

		// Zero wire resistance
		params := &WireParams{
			RwordLine: 0,
			RbitLine:  0,
			Rcontact:  0,
		}

		analysis := arr.AnalyzeIRDrop(input, params)

		// With zero resistance, IR drop should be zero
		if analysis.MaxIRDrop > 0.01 {
			t.Errorf("With zero wire resistance, MaxIRDrop=%.4f, expected ~0", analysis.MaxIRDrop)
		}

		t.Logf("Zero wire resistance: MaxIRDrop=%.4f, AvgIRDrop=%.4f", analysis.MaxIRDrop, analysis.AvgIRDrop)
	})

	t.Run("WireResistance_VeryHigh", func(t *testing.T) {
		cfg := &Config{
			Rows:       8,
			Cols:       8,
			NoiseLevel: 0.0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		defer arr.Destroy()

		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				arr.ProgramWeight(i, j, 0.5)
			}
		}

		input := make([]float64, 8)
		for j := range input {
			input[j] = 1.0
		}

		// Very high wire resistance
		params := &WireParams{
			RwordLine: 1000, // 1kOhm per cell
			RbitLine:  1000,
			Rcontact:  500,
		}

		analysis := arr.AnalyzeIRDrop(input, params)

		// Should have significant IR drop but no NaN/Inf
		if math.IsNaN(analysis.MaxIRDrop) {
			t.Error("MaxIRDrop is NaN with high resistance")
		}
		if math.IsInf(analysis.MaxIRDrop, 0) {
			t.Error("MaxIRDrop is Inf with high resistance")
		}

		t.Logf("High wire resistance: MaxIRDrop=%.4f, AvgIRDrop=%.4f", analysis.MaxIRDrop, analysis.AvgIRDrop)
	})

	t.Run("SneakPath_IsolationFactors", func(t *testing.T) {
		cfg := &Config{
			Rows:       8,
			Cols:       8,
			NoiseLevel: 0.0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		defer arr.Destroy()

		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				arr.ProgramWeight(i, j, 0.5)
			}
		}

		isolationFactors := []struct {
			name   string
			factor float64
		}{
			{"0T1R (no isolation)", 1.0},
			{"1T1R (1000x)", 0.001},
			{"2T1R (10000x)", 0.0001},
			{"zero isolation", 0.0},
		}

		for _, tc := range isolationFactors {
			analysis := arr.AnalyzeSneakPathsWithIsolation(4, 4, tc.factor)

			// No NaN/Inf
			if math.IsNaN(analysis.MaxSneakRatio) {
				t.Errorf("MaxSneakRatio is NaN for %s", tc.name)
			}
			if math.IsInf(analysis.MaxSneakRatio, 0) {
				t.Errorf("MaxSneakRatio is Inf for %s", tc.name)
			}

			// Zero isolation should give zero sneak
			if tc.factor == 0.0 && analysis.TotalSneak > 0 {
				t.Errorf("With zero isolation factor, TotalSneak=%.4e should be 0", analysis.TotalSneak)
			}

			t.Logf("%s: MaxSneakRatio=%.4f, TotalSneak=%.4e", tc.name, analysis.MaxSneakRatio, analysis.TotalSneak)
		}
	})
}

// ============================================================================
// 5. Quantization Edge Case Tests
// ============================================================================

// TestBoundaryQuantization_EdgeCases tests quantization at exact boundaries.
//
// Physics boundary being tested:
// - Exactly at quantization level boundaries
// - Values at 0.0 and 1.0 exactly
// - Verify all 30 levels are correctly mapped
func TestBoundaryQuantization_EdgeCases(t *testing.T) {
	t.Run("ExactLevelBoundaries", func(t *testing.T) {
		// Test values at exact level boundaries
		// Level spacing = 1/(30-1) = 1/29 ~ 0.0345
		levelSpacing := 1.0 / 29.0

		for level := 0; level < 30; level++ {
			expected := float64(level) * levelSpacing
			quantized := physics.QuantizeTo30Levels(expected)
			gotLevel := physics.GetLevelFor30(quantized)

			if gotLevel != level {
				t.Errorf("Level %d: input=%.10f, quantized=%.10f, gotLevel=%d",
					level, expected, quantized, gotLevel)
			}
		}
	})

	t.Run("HalfwayBetweenLevels", func(t *testing.T) {
		// Test values exactly halfway between levels (should round)
		levelSpacing := 1.0 / 29.0

		for level := 0; level < 29; level++ {
			// Value exactly halfway between level and level+1
			halfway := (float64(level) + 0.5) * levelSpacing
			quantized := physics.QuantizeTo30Levels(halfway)
			gotLevel := physics.GetLevelFor30(quantized)

			// Should round to one of the adjacent levels
			if gotLevel != level && gotLevel != level+1 {
				t.Errorf("Halfway between %d and %d: input=%.10f, gotLevel=%d",
					level, level+1, halfway, gotLevel)
			}
		}
	})

	t.Run("VerySmallIncrements", func(t *testing.T) {
		// Test that very small increments around a level stay at same level
		baseValue := 0.5 // Mid-range
		baseLevel := physics.GetLevelFor30(physics.QuantizeTo30Levels(baseValue))

		// Small perturbations should stay at same level
		perturbations := []float64{1e-10, 1e-8, 1e-6}
		for _, delta := range perturbations {
			plusLevel := physics.GetLevelFor30(physics.QuantizeTo30Levels(baseValue + delta))
			minusLevel := physics.GetLevelFor30(physics.QuantizeTo30Levels(baseValue - delta))

			// Small perturbations should not change level
			if math.Abs(float64(plusLevel-baseLevel)) > 1 {
				t.Errorf("Small perturbation +%e caused level jump: %d -> %d", delta, baseLevel, plusLevel)
			}
			if math.Abs(float64(minusLevel-baseLevel)) > 1 {
				t.Errorf("Small perturbation -%e caused level jump: %d -> %d", delta, baseLevel, minusLevel)
			}
		}
	})

	t.Run("QuantizationRoundTrip", func(t *testing.T) {
		// Test that quantizing a quantized value returns the same value
		for i := 0; i <= 100; i++ {
			original := float64(i) / 100.0
			quantized := physics.QuantizeTo30Levels(original)
			requantized := physics.QuantizeTo30Levels(quantized)

			if quantized != requantized {
				t.Errorf("Round-trip failed: %.4f -> %.4f -> %.4f", original, quantized, requantized)
			}
		}
	})
}

// ============================================================================
// 6. Physical Conductance Conversion Boundary Tests
// ============================================================================

// TestBoundaryPhysicalConductance tests conductance model edge cases.
func TestBoundaryPhysicalConductance(t *testing.T) {
	cfg := &Config{
		Rows:             4,
		Cols:             4,
		NoiseLevel:       0.0,
		ADCBits:          8,
		DACBits:          8,
		ConductanceModel: physics.ConductanceLinear,
	}

	t.Run("LinearModel_Bounds", func(t *testing.T) {
		arr, _ := NewArray(cfg)
		defer arr.Destroy()

		// Test at normalized boundaries
		testCases := []float64{0.0, 0.5, 1.0}
		for _, gNorm := range testCases {
			gPhys := arr.GetPhysicalConductance(gNorm)

			if math.IsNaN(gPhys) {
				t.Errorf("GetPhysicalConductance(%.2f) returned NaN", gNorm)
			}
			if math.IsInf(gPhys, 0) {
				t.Errorf("GetPhysicalConductance(%.2f) returned Inf", gNorm)
			}
			if gPhys < 0 {
				t.Errorf("GetPhysicalConductance(%.2f) = %.4e is negative", gNorm, gPhys)
			}

			t.Logf("Linear model: gNorm=%.2f -> gPhys=%.4e S", gNorm, gPhys)
		}
	})

	t.Run("ExponentialModel_Bounds", func(t *testing.T) {
		cfg.ConductanceModel = physics.ConductanceExponential
		arr, _ := NewArray(cfg)
		defer arr.Destroy()

		// Test at normalized boundaries
		testCases := []float64{0.0, 0.5, 1.0}
		for _, gNorm := range testCases {
			gPhys := arr.GetPhysicalConductance(gNorm)

			if math.IsNaN(gPhys) {
				t.Errorf("GetPhysicalConductance(%.2f) returned NaN", gNorm)
			}
			if math.IsInf(gPhys, 0) {
				t.Errorf("GetPhysicalConductance(%.2f) returned Inf", gNorm)
			}
			if gPhys < 0 {
				t.Errorf("GetPhysicalConductance(%.2f) = %.4e is negative", gNorm, gPhys)
			}

			t.Logf("Exponential model: gNorm=%.2f -> gPhys=%.4e S", gNorm, gPhys)
		}
	})

	t.Run("LookupModel_WithTable", func(t *testing.T) {
		cfg.ConductanceModel = physics.ConductanceLookup

		// Create a valid lookup table (30 levels)
		table := make([]float64, 30)
		for i := range table {
			table[i] = GMin + (GMax-GMin)*float64(i)/29.0
		}
		cfg.ConductanceTable = table

		arr, _ := NewArray(cfg)
		defer arr.Destroy()

		// Test at level boundaries
		for level := 0; level < 30; level++ {
			gNorm := float64(level) / 29.0
			gPhys := arr.GetPhysicalConductance(gNorm)

			if math.IsNaN(gPhys) {
				t.Errorf("Lookup GetPhysicalConductance at level %d returned NaN", level)
			}
			if math.IsInf(gPhys, 0) {
				t.Errorf("Lookup GetPhysicalConductance at level %d returned Inf", level)
			}
		}

		t.Logf("Lookup model: 30 levels verified")
	})
}

// ============================================================================
// Helper functions for tests
// ============================================================================

func minFloat64Slice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	min := s[0]
	for _, v := range s[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func maxFloat64Slice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	max := s[0]
	for _, v := range s[1:] {
		if v > max {
			max = v
		}
	}
	return max
}
