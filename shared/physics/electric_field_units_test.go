package physics

import "testing"

func TestElectricFieldConversions_MVcm_Vm(t *testing.T) {
	// Table-driven values that should round-trip exactly in floating point
	// because they are simple powers of ten.
	tests := []struct {
		name string
		mvcm float64
		vpm  float64
	}{
		{"zero", 0, 0},
		{"one", 1.0, 1e8},
		{"half", 0.5, 5e7},
		{"tenth", 0.1, 1e7},
		{"negative", -2.5, -2.5e8},
		{"alscn", 5.0, 5e8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MVPerCmToVPerM(tt.mvcm); got != tt.vpm {
				t.Fatalf("MVPerCmToVPerM(%v) = %g, want %g", tt.mvcm, got, tt.vpm)
			}
			if got := VPerMToMVPerCm(tt.vpm); got != tt.mvcm {
				t.Fatalf("VPerMToMVPerCm(%g) = %v, want %v", tt.vpm, got, tt.mvcm)
			}
		})
	}
}

func TestVPerMPerMVPerCmConstant(t *testing.T) {
	if VPerMPerMVPerCm != 1e8 {
		t.Fatalf("VPerMPerMVPerCm = %g, want 1e8", VPerMPerMVPerCm)
	}
}
