package physics

import (
	"math"
	"testing"
)

func TestQuantizeToLevels(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		levels   int
		expected float64
	}{
		{"zero", 0.0, 30, 0.0},
		{"one", 1.0, 30, 1.0},
		{"mid", 0.5, 30, 0.5172413793103448}, // level 15/29
		{"quarter", 0.25, 30, 0.2413793103448276}, // level 7/29
		{"clamp_negative", -0.5, 30, 0.0},
		{"clamp_above", 1.5, 30, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := QuantizeToLevels(tt.value, tt.levels)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("QuantizeToLevels(%v, %d) = %v, want %v", tt.value, tt.levels, result, tt.expected)
			}
		})
	}
}

func TestQuantizeTo30Levels(t *testing.T) {
	// Test that quantization produces exactly 30 unique levels
	levels := make(map[float64]bool)
	for i := 0; i < 1000; i++ {
		v := float64(i) / 999.0
		q := QuantizeTo30Levels(v)
		levels[q] = true
	}
	if len(levels) != 30 {
		t.Errorf("QuantizeTo30Levels produced %d unique levels, want 30", len(levels))
	}
}

func TestGetLevel(t *testing.T) {
	tests := []struct {
		conductance float64
		levels      int
		expected    int
	}{
		{0.0, 30, 0},
		{1.0, 30, 29},
		{0.5, 30, 15},
		{0.0, 32, 0},
		{1.0, 32, 31},
	}

	for _, tt := range tests {
		result := GetLevel(tt.conductance, tt.levels)
		if result != tt.expected {
			t.Errorf("GetLevel(%v, %d) = %d, want %d", tt.conductance, tt.levels, result, tt.expected)
		}
	}
}

func TestGetLevelFor30(t *testing.T) {
	if GetLevelFor30(0.0) != 0 {
		t.Error("GetLevelFor30(0.0) should be 0")
	}
	if GetLevelFor30(1.0) != 29 {
		t.Error("GetLevelFor30(1.0) should be 29")
	}
}

func TestNormalizeFromLevel(t *testing.T) {
	tests := []struct {
		level    int
		levels   int
		expected float64
	}{
		{0, 30, 0.0},
		{29, 30, 1.0},
		{15, 30, 15.0 / 29.0},
		{-1, 30, 0.0},   // clamped
		{100, 30, 1.0},  // clamped
	}

	for _, tt := range tests {
		result := NormalizeFromLevel(tt.level, tt.levels)
		if math.Abs(result-tt.expected) > 1e-10 {
			t.Errorf("NormalizeFromLevel(%d, %d) = %v, want %v", tt.level, tt.levels, result, tt.expected)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that level -> normalized -> level is identity
	for level := 0; level < 30; level++ {
		norm := NormalizeFromLevel30(level)
		back := GetLevelFor30(norm)
		if back != level {
			t.Errorf("Round trip failed: level %d -> norm %v -> level %d", level, norm, back)
		}
	}
}

func TestLevelSpacing(t *testing.T) {
	spacing := LevelSpacing(30)
	expected := 1.0 / 29.0
	if math.Abs(spacing-expected) > 1e-10 {
		t.Errorf("LevelSpacing(30) = %v, want %v", spacing, expected)
	}
}

func TestQuantizationError(t *testing.T) {
	err := QuantizationError(30)
	expected := 1.0 / 58.0 // half of level spacing
	if math.Abs(err-expected) > 1e-10 {
		t.Errorf("QuantizationError(30) = %v, want %v", err, expected)
	}
}

func TestDefaultLevelsConstant(t *testing.T) {
	if DefaultLevels != 30 {
		t.Errorf("DefaultLevels = %d, want 30", DefaultLevels)
	}
}
