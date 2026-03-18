package physics

import (
	"fmt"
	"math"
	"testing"
)

// TestConductanceToPolarizationWithParams_RoundTrip verifies that the model-aware
// inverse ConductanceToPolarizationWithParams correctly inverts
// PolarizationToConductanceWithParams for each conductance model.
func TestConductanceToPolarizationWithParams_RoundTrip(t *testing.T) {
	const (
		Ps   = 0.25
		Gmin = 10e-6
		Gmax = 100e-6
		N    = 30
	)

	cases := []struct {
		name    string
		model   ConductanceModel
		kvT     float64
		vgsRead float64
		vt0     float64
	}{
		{"Linear", ConductanceLinear, 0, 0, 0},
		{"Subthreshold_kvT0", ConductanceSubthreshold, 0, 0, 0},
		{"Subthreshold_kvT0.2", ConductanceSubthreshold, 0.2, 0, 0},
		{"Saturation_kvT0.2", ConductanceSaturation, 0.2, 1.0, 0.5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < N; i++ {
				P := -Ps + 2*Ps*float64(i)/float64(N-1)

				G := PolarizationToConductanceWithParams(P, Ps, Gmin, Gmax, tc.model, tc.kvT, tc.vgsRead, tc.vt0)
				Precovered := ConductanceToPolarizationWithParams(G, Ps, Gmin, Gmax, tc.model, tc.kvT, tc.vgsRead, tc.vt0)

				tol := 1e-6 * Ps
				if math.Abs(P-Precovered) > tol {
					t.Errorf("model=%s i=%d: P=%.15g -> G=%.15g -> Precovered=%.15g, diff=%.3e (tol=%.3e)",
						tc.name, i, P, G, Precovered, math.Abs(P-Precovered), tol)
				}
			}
		})
	}
}

// TestConductanceToPolarizationWithParams_EdgeCases checks boundary and
// degenerate inputs for the model-aware inverse.
func TestConductanceToPolarizationWithParams_EdgeCases(t *testing.T) {
	const (
		Ps   = 0.25
		Gmin = 10e-6
		Gmax = 100e-6
	)

	t.Run("Gmin_returns_neg_Ps", func(t *testing.T) {
		P := ConductanceToPolarizationWithParams(Gmin, Ps, Gmin, Gmax, ConductanceLinear, 0, 0, 0)
		if P != -Ps {
			t.Errorf("expected -Ps=%g, got %g", -Ps, P)
		}
	})

	t.Run("Gmax_returns_pos_Ps", func(t *testing.T) {
		P := ConductanceToPolarizationWithParams(Gmax, Ps, Gmin, Gmax, ConductanceLinear, 0, 0, 0)
		if P != Ps {
			t.Errorf("expected Ps=%g, got %g", Ps, P)
		}
	})

	t.Run("below_Gmin_clamps", func(t *testing.T) {
		P := ConductanceToPolarizationWithParams(Gmin/2, Ps, Gmin, Gmax, ConductanceSubthreshold, 0, 0, 0)
		if P != -Ps {
			t.Errorf("expected -Ps=%g, got %g", -Ps, P)
		}
	})

	t.Run("above_Gmax_clamps", func(t *testing.T) {
		P := ConductanceToPolarizationWithParams(Gmax*2, Ps, Gmin, Gmax, ConductanceSaturation, 0.2, 1.0, 0.5)
		if P != Ps {
			t.Errorf("expected Ps=%g, got %g", Ps, P)
		}
	})

	t.Run("Ps_zero_returns_zero", func(t *testing.T) {
		P := ConductanceToPolarizationWithParams(50e-6, 0, Gmin, Gmax, ConductanceLinear, 0, 0, 0)
		if P != 0 {
			t.Errorf("expected 0, got %g", P)
		}
	})

	t.Run("Gmin_eq_Gmax_returns_zero", func(t *testing.T) {
		P := ConductanceToPolarizationWithParams(50e-6, Ps, Gmin, Gmin, ConductanceLinear, 0, 0, 0)
		if P != 0 {
			t.Errorf("expected 0, got %g", P)
		}
	})
}

// TestConductanceToPolarizationWithParams_BackwardCompat verifies that the linear
// model path produces identical results to the legacy ConductanceToPolarization.
func TestConductanceToPolarizationWithParams_BackwardCompat(t *testing.T) {
	const (
		Ps   = 0.25
		Gmin = 10e-6
		Gmax = 100e-6
	)

	for i := 0; i < 30; i++ {
		G := Gmin + (Gmax-Gmin)*float64(i)/29.0

		legacy := ConductanceToPolarization(G, Gmin, Gmax, Ps)
		newFn := ConductanceToPolarizationWithParams(G, Ps, Gmin, Gmax, ConductanceLinear, 0, 0, 0)

		if math.Abs(legacy-newFn) > 1e-15 {
			t.Errorf("i=%d G=%g: legacy=%g new=%g diff=%e",
				i, G, legacy, newFn, math.Abs(legacy-newFn))
		}
	}
}

// TestConductanceToPolarizationWithParams_Monotonic verifies the inverse is
// monotonically increasing for all models.
func TestConductanceToPolarizationWithParams_Monotonic(t *testing.T) {
	const (
		Ps   = 0.25
		Gmin = 10e-6
		Gmax = 100e-6
		N    = 100
	)

	cases := []struct {
		name    string
		model   ConductanceModel
		kvT     float64
		vgsRead float64
		vt0     float64
	}{
		{"Linear", ConductanceLinear, 0, 0, 0},
		{"Subthreshold_kvT0", ConductanceSubthreshold, 0, 0, 0},
		{"Subthreshold_kvT0.2", ConductanceSubthreshold, 0.2, 0, 0},
		{"Saturation", ConductanceSaturation, 0.2, 1.0, 0.5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prevP := -Ps - 1 // sentinel below any valid P
			for i := 0; i <= N; i++ {
				G := Gmin + (Gmax-Gmin)*float64(i)/float64(N)
				P := ConductanceToPolarizationWithParams(G, Ps, Gmin, Gmax, tc.model, tc.kvT, tc.vgsRead, tc.vt0)
				if P < prevP {
					t.Errorf("non-monotonic at i=%d: P(%g)=%g < prevP=%g",
						i, G, P, prevP)
				}
				prevP = P
			}
		})
	}
}

// BenchmarkConductanceToPolarizationWithParams_Bisection measures the cost of
// bisection-based inversion used by subthreshold (kvT>0) and saturation models.
func BenchmarkConductanceToPolarizationWithParams_Bisection(b *testing.B) {
	const (
		Ps   = 0.25
		Gmin = 10e-6
		Gmax = 100e-6
	)
	G := (Gmin + Gmax) / 2

	cases := []struct {
		name  string
		model ConductanceModel
		kvT   float64
	}{
		{"Linear", ConductanceLinear, 0},
		{"Subthreshold_log", ConductanceSubthreshold, 0},
		{"Subthreshold_bisect", ConductanceSubthreshold, 0.2},
		{"Saturation_bisect", ConductanceSaturation, 0.2},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ConductanceToPolarizationWithParams(G, Ps, Gmin, Gmax, tc.model, tc.kvT, 1.0, 0.5)
			}
		})
	}
}

// Prevent compiler from optimizing away benchmark results.
var _ = fmt.Sprintf
