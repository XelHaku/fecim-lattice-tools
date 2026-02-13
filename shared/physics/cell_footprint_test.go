package physics

import "testing"

func TestCalculateFootprint_KnownArchitectures(t *testing.T) {
	tests := []struct {
		arch         string
		minF2, maxF2 float64
	}{
		{"0T1R", 4, 4},
		{"1T1R", 6, 12},
		{"2T1R", 12, 20},
		{"SRAM", 120, 150},
	}

	for _, tc := range tests {
		fp := CalculateFootprint(tc.arch, 28)
		if fp.F <= 0 {
			t.Fatalf("%s: F must be positive", tc.arch)
		}
		f2 := fp.F * fp.F
		areaF2 := fp.TotalArea / f2
		if areaF2 < tc.minF2 || areaF2 > tc.maxF2 {
			t.Fatalf("%s: area=%.2f F^2 out of [%.2f, %.2f]", tc.arch, areaF2, tc.minF2, tc.maxF2)
		}
		if fp.Density <= 0 {
			t.Fatalf("%s: density must be positive", tc.arch)
		}
	}
}

func TestCalculateFootprint_DensityOrdering(t *testing.T) {
	fp0 := CalculateFootprint("0T1R", 28)
	fp1 := CalculateFootprint("1T1R", 28)
	fp2 := CalculateFootprint("2T1R", 28)
	fps := CalculateFootprint("SRAM", 28)

	if !(fp0.Density > fp1.Density && fp1.Density > fp2.Density && fp2.Density > fps.Density) {
		t.Fatalf("unexpected density ordering: 0T1R=%e 1T1R=%e 2T1R=%e SRAM=%e", fp0.Density, fp1.Density, fp2.Density, fps.Density)
	}
}
