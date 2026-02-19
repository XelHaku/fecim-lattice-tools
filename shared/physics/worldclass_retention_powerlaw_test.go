package physics

import (
	"math"
	"testing"
)

func TestSimulateRetentionPowerLaw_ZeroBeta(t *testing.T) {
	times := []float64{1, 10, 100}
	pts, err := SimulateRetentionPowerLaw(0.25, 1.0, 0.0, times)
	if err != nil {
		t.Fatal(err)
	}
	for _, pt := range pts {
		if math.Abs(pt.Polarization_Cm-0.25) > 1e-12 {
			t.Errorf("beta=0 should not decay: got %g", pt.Polarization_Cm)
		}
	}
}

func TestSimulateRetentionPowerLaw_Decay(t *testing.T) {
	// At t=t0 P=P0, at t=10*t0 P=P0*10^(-beta)
	P0, t0, beta := 0.25, 1.0, 0.03
	times := []float64{t0, 10 * t0, 100 * t0}
	pts, err := SimulateRetentionPowerLaw(P0, t0, beta, times)
	if err != nil {
		t.Fatal(err)
	}
	// t=t0: P should equal P0
	if math.Abs(pts[0].Polarization_Cm-P0) > 1e-12 {
		t.Errorf("at t=t0 expected P0=%g, got %g", P0, pts[0].Polarization_Cm)
	}
	// t=10*t0: P = P0 * 10^(-0.03)
	expected := P0 * math.Pow(10.0, -beta)
	if math.Abs(pts[1].Polarization_Cm-expected) > 1e-10 {
		t.Errorf("at t=10*t0 expected %g, got %g", expected, pts[1].Polarization_Cm)
	}
	// Monotonic decrease
	if pts[1].Polarization_Cm >= pts[0].Polarization_Cm {
		t.Error("polarization should decrease over time")
	}
}

func TestSimulateRetentionPowerLaw_InvalidInputs(t *testing.T) {
	_, err := SimulateRetentionPowerLaw(0.25, -1, 0.03, []float64{1.0})
	if err == nil {
		t.Error("expected error for negative t0")
	}
	_, err = SimulateRetentionPowerLaw(0.25, 1.0, -0.1, []float64{1.0})
	if err == nil {
		t.Error("expected error for negative beta")
	}
}
