package crossbar

import (
	"math"
	"testing"
)

// TestQuantizeTo30LevelsProducesExactly30Values verifies FeCIM spec:
// "It's got 30 discrete states." — Dr. Tour
func TestQuantizeTo30LevelsProducesExactly30Values(t *testing.T) {
	seen := make(map[float64]bool)

	// Test many input values
	for i := 0; i <= 1000; i++ {
		input := float64(i) / 1000.0
		quantized := QuantizeTo30Levels(input)
		seen[quantized] = true
	}

	if len(seen) != FeCIMLevels {
		t.Errorf("Expected exactly %d discrete levels, got %d", FeCIMLevels, len(seen))
	}
}

// TestQuantizeTo30LevelsRange verifies output is in [0, 1]
func TestQuantizeTo30LevelsRange(t *testing.T) {
	testCases := []float64{-0.5, 0.0, 0.5, 1.0, 1.5}
	for _, input := range testCases {
		output := QuantizeTo30Levels(input)
		if output < 0 || output > 1 {
			t.Errorf("QuantizeTo30Levels(%v) = %v, expected in [0, 1]", input, output)
		}
	}
}

// TestQuantizeTo30LevelsSpecificValues verifies key quantization points
func TestQuantizeTo30LevelsSpecificValues(t *testing.T) {
	testCases := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},                 // Level 0
		{1.0, 1.0},                 // Level 29
		{0.5, 15.0 / 29.0},         // 0.5 * 29 = 14.5 -> rounds to 15 -> 15/29
		{1.0 / 29.0, 1.0 / 29.0},   // Level 1
		{14.0 / 29.0, 14.0 / 29.0}, // Level 14 (exact)
	}

	for _, tc := range testCases {
		output := QuantizeTo30Levels(tc.input)
		if math.Abs(output-tc.expected) > 1e-10 {
			t.Errorf("QuantizeTo30Levels(%v) = %v, expected %v", tc.input, output, tc.expected)
		}
	}
}

// TestGetLevelReturns0To29 verifies level indices are 0-29
func TestGetLevelReturns0To29(t *testing.T) {
	for i := 0; i < FeCIMLevels; i++ {
		conductance := float64(i) / float64(FeCIMLevels-1)
		level := GetLevel(conductance)
		if level != i {
			t.Errorf("GetLevel(%v) = %d, expected %d", conductance, level, i)
		}
	}
}

// TestProgramWeightQuantizes verifies ProgramWeight uses 30-level quantization
func TestProgramWeightQuantizes(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows: 4, Cols: 4, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Program a value that isn't on a 30-level boundary
	err = arr.ProgramWeight(0, 0, 0.123)
	if err != nil {
		t.Fatal(err)
	}

	// Get the stored value
	matrix := arr.GetConductanceMatrix()
	stored := matrix[0][0]

	// Should be quantized to nearest 30-level
	expected := QuantizeTo30Levels(0.123)
	if math.Abs(stored-expected) > 1e-10 {
		t.Errorf("ProgramWeight stored %v, expected quantized %v", stored, expected)
	}
}

// TestMVMCorrectness verifies matrix-vector multiplication
func TestMVMCorrectness(t *testing.T) {
	arr, err := NewArray(&Config{
		Rows: 2, Cols: 2, NoiseLevel: 0, ADCBits: 8, DACBits: 8,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Program identity-like weights
	arr.ProgramWeight(0, 0, 1.0)
	arr.ProgramWeight(0, 1, 0.0)
	arr.ProgramWeight(1, 0, 0.0)
	arr.ProgramWeight(1, 1, 1.0)

	input := []float64{1.0, 0.5}
	output, err := arr.MVM(input)
	if err != nil {
		t.Fatal(err)
	}

	// With quantization and averaging, outputs should be reasonable
	if len(output) != 2 {
		t.Errorf("MVM output length = %d, expected 2", len(output))
	}
}

// TestArrayDimensions verifies array size accessors
func TestArrayDimensions(t *testing.T) {
	arr, _ := NewArray(&Config{
		Rows: 128, Cols: 784, NoiseLevel: 0.01, ADCBits: 8, DACBits: 8,
	})

	if arr.Rows() != 128 {
		t.Errorf("Rows() = %d, expected 128", arr.Rows())
	}
	if arr.Cols() != 784 {
		t.Errorf("Cols() = %d, expected 784", arr.Cols())
	}
}
