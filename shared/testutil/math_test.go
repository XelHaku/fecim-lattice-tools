package testutil

import (
	"math"
	"testing"
)

func TestRelErr(t *testing.T) {
	tests := []struct {
		name string
		got  float64
		want float64
	}{
		{"exact match", 1.0, 1.0},
		{"zero want", 0.5, 0.0},
		{"nonzero want", 1.0, 2.0},
		{"both zero", 0.0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RelErr(tt.got, tt.want)
			if math.IsNaN(err) || err < 0 {
				t.Errorf("RelErr(%v, %v) = %v, want non-negative number", tt.got, tt.want, err)
			}
		})
	}
}

func TestAbsFloat64(t *testing.T) {
	if got := AbsFloat64(-3.5); got != 3.5 {
		t.Errorf("AbsFloat64(-3.5) = %v, want 3.5", got)
	}
	if got := AbsFloat64(3.5); got != 3.5 {
		t.Errorf("AbsFloat64(3.5) = %v, want 3.5", got)
	}
	if got := AbsFloat64(0); got != 0 {
		t.Errorf("AbsFloat64(0) = %v, want 0", got)
	}
}

func TestAbsInt(t *testing.T) {
	if got := AbsInt(-5); got != 5 {
		t.Errorf("AbsInt(-5) = %v, want 5", got)
	}
	if got := AbsInt(5); got != 5 {
		t.Errorf("AbsInt(5) = %v, want 5", got)
	}
	if got := AbsInt(0); got != 0 {
		t.Errorf("AbsInt(0) = %v, want 0", got)
	}
}

func TestClamp01(t *testing.T) {
	if got := Clamp01(-0.5); got != 0 {
		t.Errorf("Clamp01(-0.5) = %v, want 0", got)
	}
	if got := Clamp01(0.5); got != 0.5 {
		t.Errorf("Clamp01(0.5) = %v, want 0.5", got)
	}
	if got := Clamp01(1.5); got != 1 {
		t.Errorf("Clamp01(1.5) = %v, want 1", got)
	}
}
