package physics

import (
	"math"
	"testing"
)

func TestConductanceConstants(t *testing.T) {
	// Verify constants match expected values
	if GMin != 1e-6 {
		t.Errorf("GMin = %e, want 1e-6", GMin)
	}
	if GMax != 100e-6 {
		t.Errorf("GMax = %e, want 100e-6", GMax)
	}
	if math.Abs(GRatio-100.0) > 1e-10 {
		t.Errorf("GRatio = %v, want 100.0", GRatio)
	}
}

func TestConductanceModelString(t *testing.T) {
	tests := []struct {
		model    ConductanceModel
		expected string
	}{
		{ConductanceLinear, "Linear (Educational/Simplified)"},
		{ConductanceExponential, "Exponential (Recommended)"},
		{ConductanceLookup, "Lookup (Calibrated)"},
		{ConductanceModel(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.model.String(); got != tt.expected {
			t.Errorf("ConductanceModel(%d).String() = %q, want %q", tt.model, got, tt.expected)
		}
	}
}

func TestNormalizedToPhysicalLinear(t *testing.T) {
	tests := []struct {
		gNorm    float64
		expected float64
	}{
		{0.0, GMin},
		{1.0, GMax},
		{0.5, (GMin + GMax) / 2}, // arithmetic mean
	}

	for _, tt := range tests {
		result := NormalizedToPhysical(tt.gNorm, ConductanceLinear)
		if math.Abs(result-tt.expected) > 1e-15 {
			t.Errorf("NormalizedToPhysical(%v, Linear) = %e, want %e", tt.gNorm, result, tt.expected)
		}
	}
}

func TestNormalizedToPhysicalExponential(t *testing.T) {
	// Test boundaries
	if result := NormalizedToPhysical(0.0, ConductanceExponential); math.Abs(result-GMin) > 1e-15 {
		t.Errorf("Exponential at 0.0 = %e, want %e", result, GMin)
	}
	if result := NormalizedToPhysical(1.0, ConductanceExponential); math.Abs(result-GMax) > 1e-15 {
		t.Errorf("Exponential at 1.0 = %e, want %e", result, GMax)
	}

	// Test midpoint is geometric mean
	geometricMean := math.Sqrt(GMin * GMax)
	result := NormalizedToPhysical(0.5, ConductanceExponential)
	if math.Abs(result-geometricMean) > 1e-15 {
		t.Errorf("Exponential at 0.5 = %e, want geometric mean %e", result, geometricMean)
	}
}

func TestNormalizedToPhysicalLookup(t *testing.T) {
	// Create a simple lookup table
	table := []float64{10e-6, 30e-6, 50e-6, 70e-6, 100e-6}

	result := NormalizedToPhysicalRange(0.0, ConductanceLookup, GMin, GMax, table)
	if math.Abs(result-10e-6) > 1e-15 {
		t.Errorf("Lookup at 0.0 = %e, want 10e-6", result)
	}

	result = NormalizedToPhysicalRange(1.0, ConductanceLookup, GMin, GMax, table)
	if math.Abs(result-100e-6) > 1e-15 {
		t.Errorf("Lookup at 1.0 = %e, want 100e-6", result)
	}

	result = NormalizedToPhysicalRange(0.5, ConductanceLookup, GMin, GMax, table)
	if math.Abs(result-50e-6) > 1e-15 {
		t.Errorf("Lookup at 0.5 = %e, want 50e-6", result)
	}
}

func TestNormalizedToPhysicalClamping(t *testing.T) {
	// Test clamping for out-of-range values
	result := NormalizedToPhysical(-0.5, ConductanceLinear)
	if math.Abs(result-GMin) > 1e-15 {
		t.Errorf("Linear at -0.5 should clamp to GMin, got %e", result)
	}

	result = NormalizedToPhysical(1.5, ConductanceLinear)
	if math.Abs(result-GMax) > 1e-15 {
		t.Errorf("Linear at 1.5 should clamp to GMax, got %e", result)
	}
}

func TestPhysicalToNormalized(t *testing.T) {
	tests := []struct {
		gPhys    float64
		expected float64
	}{
		{GMin, 0.0},
		{GMax, 1.0},
		{(GMin + GMax) / 2, 0.5},
		{0.5e-6, 0.0}, // below GMin, clamped
		{200e-6, 1.0}, // above GMax, clamped
	}

	for _, tt := range tests {
		result := PhysicalToNormalized(tt.gPhys)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("PhysicalToNormalized(%e) = %v, want %v", tt.gPhys, result, tt.expected)
		}
	}
}

func TestConductanceToLevel(t *testing.T) {
	// GMin should be level 0
	if level := ConductanceToLevel(GMin, 30); level != 0 {
		t.Errorf("ConductanceToLevel(GMin, 30) = %d, want 0", level)
	}

	// GMax should be level 29
	if level := ConductanceToLevel(GMax, 30); level != 29 {
		t.Errorf("ConductanceToLevel(GMax, 30) = %d, want 29", level)
	}
}

func TestLevelToConductance(t *testing.T) {
	// Level 0 should be GMin
	g := LevelToConductance(0, 30, ConductanceLinear)
	if math.Abs(g-GMin) > 1e-15 {
		t.Errorf("LevelToConductance(0, 30, Linear) = %e, want %e", g, GMin)
	}

	// Level 29 should be GMax
	g = LevelToConductance(29, 30, ConductanceLinear)
	if math.Abs(g-GMax) > 1e-15 {
		t.Errorf("LevelToConductance(29, 30, Linear) = %e, want %e", g, GMax)
	}
}

func TestConductanceWindow(t *testing.T) {
	expected := GMax - GMin
	if window := ConductanceWindow(); math.Abs(window-expected) > 1e-15 {
		t.Errorf("ConductanceWindow() = %e, want %e", window, expected)
	}
}

func TestConductanceMidpoint(t *testing.T) {
	expected := (GMax + GMin) / 2
	if mid := ConductanceMidpoint(); math.Abs(mid-expected) > 1e-15 {
		t.Errorf("ConductanceMidpoint() = %e, want %e", mid, expected)
	}
}

func TestConductanceGeometricMean(t *testing.T) {
	expected := math.Sqrt(GMin * GMax)
	if gm := ConductanceGeometricMean(); math.Abs(gm-expected) > 1e-15 {
		t.Errorf("ConductanceGeometricMean() = %e, want %e", gm, expected)
	}
}
