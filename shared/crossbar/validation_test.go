package crossbar

import (
	"fecim-lattice-tools/shared/physics"
	"math"
	"testing"
)

// TestConstantsMatchDocumentation verifies FEATURES.md values
func TestConstantsMatchDocumentation(t *testing.T) {
	t.Run("GMin matches 10uS", func(t *testing.T) {
		expected := 10e-6
		if math.Abs(GMin-expected) > 1e-12 {
			t.Errorf("GMin = %e, want %e (10 uS)", GMin, expected)
		}
	})

	t.Run("GMax matches 100uS", func(t *testing.T) {
		expected := 100e-6
		if math.Abs(GMax-expected) > 1e-12 {
			t.Errorf("GMax = %e, want %e (100 uS)", GMax, expected)
		}
	})

	t.Run("DefaultQuantizationLevels matches 30", func(t *testing.T) {
		expected := 30
		if DefaultQuantizationLevels != expected {
			t.Errorf("DefaultQuantizationLevels = %d, want %d", DefaultQuantizationLevels, expected)
		}
	})

	t.Run("DefaultWireResistance matches 2.5 ohm", func(t *testing.T) {
		params := DefaultWireParams()
		expected := 2.5
		if math.Abs(params.RwordLine-expected) > 1e-10 {
			t.Errorf("RwordLine = %f, want %f", params.RwordLine, expected)
		}
		if math.Abs(params.RbitLine-expected) > 1e-10 {
			t.Errorf("RbitLine = %f, want %f", params.RbitLine, expected)
		}
	})

	t.Run("GRatio matches 10:1", func(t *testing.T) {
		expected := 10.0
		actual := GMax / GMin
		if math.Abs(actual-expected) > 1e-10 {
			t.Errorf("GRatio = %f, want %f", actual, expected)
		}
	})
}

// TestConductanceRangeIntegrity verifies conductance window properties
func TestConductanceRangeIntegrity(t *testing.T) {
	t.Run("GMax > GMin", func(t *testing.T) {
		if GMax <= GMin {
			t.Errorf("GMax (%e) must be > GMin (%e)", GMax, GMin)
		}
	})

	t.Run("Conductance window is positive", func(t *testing.T) {
		window := GMax - GMin
		if window <= 0 {
			t.Errorf("Conductance window (%e) must be positive", window)
		}
	})

	t.Run("Arithmetic midpoint calculation", func(t *testing.T) {
		expected := 55e-6 // (10 + 100) / 2
		actual := (GMin + GMax) / 2
		if math.Abs(actual-expected) > 1e-12 {
			t.Errorf("Arithmetic midpoint = %e, want %e", actual, expected)
		}
	})

	t.Run("Geometric midpoint calculation", func(t *testing.T) {
		expected := math.Sqrt(GMin * GMax) // ~31.62 uS
		actual := math.Sqrt(10e-6 * 100e-6)
		if math.Abs(actual-expected) > 1e-12 {
			t.Errorf("Geometric midpoint = %e, want %e", actual, expected)
		}
		t.Logf("Geometric midpoint: %.4f uS", expected*1e6)
	})
}

// TestQuantizationLevelsProperties verifies 30-level quantization
func TestQuantizationLevelsProperties(t *testing.T) {
	t.Run("Level count is 30", func(t *testing.T) {
		if DefaultQuantizationLevels != 30 {
			t.Errorf("DefaultQuantizationLevels = %d, want 30", DefaultQuantizationLevels)
		}
	})

	t.Run("Level spacing is uniform", func(t *testing.T) {
		expectedSpacing := 1.0 / float64(DefaultQuantizationLevels-1) // 1/29

		for i := 0; i < DefaultQuantizationLevels-1; i++ {
			level_i := float64(i) / float64(DefaultQuantizationLevels-1)
			level_next := float64(i+1) / float64(DefaultQuantizationLevels-1)
			spacing := level_next - level_i

			if math.Abs(spacing-expectedSpacing) > 1e-14 {
				t.Errorf("Level %d-%d spacing = %.15f, want %.15f", i, i+1, spacing, expectedSpacing)
			}
		}
	})

	t.Run("Bits per cell approximation", func(t *testing.T) {
		// 30 levels = log2(30) = ~4.9 bits
		bitsPerCell := math.Log2(float64(DefaultQuantizationLevels))
		if bitsPerCell < 4.5 || bitsPerCell > 5.0 {
			t.Errorf("Bits per cell = %.2f, expected ~4.9", bitsPerCell)
		}
		t.Logf("Bits per cell: %.3f", bitsPerCell)
	})
}

// TestQuantizationFunctionConsistency verifies quantization behavior
func TestQuantizationFunctionConsistency(t *testing.T) {
	t.Run("QuantizeToLevels wraps physics function", func(t *testing.T) {
		testValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
		for _, val := range testValues {
			result := QuantizeToLevels(val)
			expected := physics.QuantizeTo30Levels(val)
			if result != expected {
				t.Errorf("QuantizeToLevels(%f) = %f, want %f", val, result, expected)
			}
		}
	})

	t.Run("GetLevel wraps physics function", func(t *testing.T) {
		testValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
		for _, val := range testValues {
			result := GetLevel(val)
			expected := physics.GetLevelFor30(val)
			if result != expected {
				t.Errorf("GetLevel(%f) = %d, want %d", val, result, expected)
			}
		}
	})

	t.Run("GetLevel returns valid range", func(t *testing.T) {
		testValues := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
		for _, val := range testValues {
			level := GetLevel(val)
			if level < 0 || level >= DefaultQuantizationLevels {
				t.Errorf("GetLevel(%f) = %d, expected [0, %d)", val, level, DefaultQuantizationLevels)
			}
		}
	})
}

// TestPhysicsConstantConsistency verifies shared/physics aliases
func TestPhysicsConstantConsistency(t *testing.T) {
	t.Run("GMin aliases physics.GMin", func(t *testing.T) {
		if GMin != physics.GMin {
			t.Errorf("GMin = %e, physics.GMin = %e, should be identical", GMin, physics.GMin)
		}
	})

	t.Run("GMax aliases physics.GMax", func(t *testing.T) {
		if GMax != physics.GMax {
			t.Errorf("GMax = %e, physics.GMax = %e, should be identical", GMax, physics.GMax)
		}
	})

	t.Run("DefaultQuantizationLevels aliases physics.DefaultLevels", func(t *testing.T) {
		if DefaultQuantizationLevels != physics.DefaultLevels {
			t.Errorf("DefaultQuantizationLevels = %d, physics.DefaultLevels = %d, should be identical",
				DefaultQuantizationLevels, physics.DefaultLevels)
		}
	})
}

// TestConductanceModelConstants verifies model enum consistency
func TestConductanceModelConstants(t *testing.T) {
	t.Run("Model constants alias physics package", func(t *testing.T) {
		if ConductanceLinear != physics.ConductanceLinear {
			t.Error("ConductanceLinear constant mismatch with physics package")
		}
		if ConductanceExponential != physics.ConductanceExponential {
			t.Error("ConductanceExponential constant mismatch with physics package")
		}
		if ConductanceLookup != physics.ConductanceLookup {
			t.Error("ConductanceLookup constant mismatch with physics package")
		}
	})
}

// TestWireParamsDefaults verifies wire parameter physical realism
func TestWireParamsDefaults(t *testing.T) {
	params := DefaultWireParams()

	t.Run("Word line resistance is positive", func(t *testing.T) {
		if params.RwordLine <= 0 {
			t.Errorf("RwordLine = %f, must be positive", params.RwordLine)
		}
	})

	t.Run("Bit line resistance is positive", func(t *testing.T) {
		if params.RbitLine <= 0 {
			t.Errorf("RbitLine = %f, must be positive", params.RbitLine)
		}
	})

	t.Run("Contact resistance is positive", func(t *testing.T) {
		if params.Rcontact <= 0 {
			t.Errorf("Rcontact = %f, must be positive", params.Rcontact)
		}
	})

	t.Run("Contact resistance > wire resistance", func(t *testing.T) {
		// Contact resistance should be significantly higher than wire resistance
		if params.Rcontact <= params.RwordLine {
			t.Errorf("Rcontact (%f) should be > RwordLine (%f)", params.Rcontact, params.RwordLine)
		}
	})
}
