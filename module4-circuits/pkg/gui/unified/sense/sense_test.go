package sense

import "testing"

func TestADCCodePercent(t *testing.T) {
	if got := ADCCodePercent(0, 32); got != 0 {
		t.Fatalf("code0 got=%v want=0", got)
	}
	if got := ADCCodePercent(31, 32); got != 100 {
		t.Fatalf("fullscale got=%v want=100", got)
	}
}

func TestClampADCRange(t *testing.T) {
	vmin, vmax := ClampADCRange(-1.0, 2.0, 0.0, 1.0, 0.1)
	if vmin < 0.0 || vmax > 1.0 {
		t.Fatalf("range out of bounds: [%f,%f]", vmin, vmax)
	}
	if vmax-vmin < 0.1 {
		t.Fatalf("span too small: %f", vmax-vmin)
	}
}

func TestClampNearZero(t *testing.T) {
	if got := ClampNearZero(1e-10, 1e-6); got != 0 {
		t.Fatalf("near-zero not clamped: %g", got)
	}
	if got := ClampNearZero(1e-3, 1e-6); got == 0 {
		t.Fatalf("non-zero incorrectly clamped")
	}
}
