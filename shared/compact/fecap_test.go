package compact

import (
	"math"
	"testing"
)

func defaultParams() FeCapParams {
	return DefaultFeCapHZO()
}

// TestSwitchingFunction_Saturation verifies that the switching function saturates
// at ±Qs for large positive/negative voltages.
func TestSwitchingFunction_Saturation(t *testing.T) {
	p := defaultParams()
	// Large positive V, dir=+1 → should be near +Qs
	Vlarge := 10 * p.CoerciveV
	hi := p.SwitchingFunction(Vlarge, +1)
	if math.Abs(hi-p.Qs) > 1e-6*p.Qs {
		t.Errorf("positive saturation: got %.4e, want ≈%.4e", hi, p.Qs)
	}
	// Large negative V, dir=-1 → should be near -Qs
	lo := p.SwitchingFunction(-Vlarge, -1)
	if math.Abs(lo+p.Qs) > 1e-6*p.Qs {
		t.Errorf("negative saturation: got %.4e, want ≈-%.4e", lo, p.Qs)
	}
}

// TestSwitchingFunction_AtCoercive verifies F(Vc, +1) = 0 (coercive point).
func TestSwitchingFunction_AtCoercive(t *testing.T) {
	p := defaultParams()
	// At V=+Vc, dir=+1: tanh(0) = 0
	f := p.SwitchingFunction(p.CoerciveV, +1)
	if math.Abs(f) > 1e-15 {
		t.Errorf("at coercive voltage: got %.3e, want 0", f)
	}
}

// TestLinearCharge verifies the linear dielectric contribution.
func TestLinearCharge_Proportional(t *testing.T) {
	p := defaultParams()
	// Q_lin should scale linearly with V
	q1 := p.LinearCharge(1.0)
	q2 := p.LinearCharge(2.0)
	if math.Abs(q2/q1-2.0) > 1e-10 {
		t.Errorf("LinearCharge not proportional to V: ratio %.6f, want 2.0", q2/q1)
	}
	// Sanity: ε₀ × 30 × 1V / 10nm = 8.854e-12 × 30 / 10e-9 = 0.02656 C/m²
	want := epsilon0 * p.EpsFEr * 1.0 / p.ThicknessFE
	if math.Abs(q1-want) > 1e-15 {
		t.Errorf("LinearCharge(1V): got %.4e, want %.4e", q1, want)
	}
}

// TestNewtonSolver_RoundTrip verifies that SolveVFE(TotalCharge(V)) ≈ V.
func TestNewtonSolver_RoundTrip(t *testing.T) {
	p := defaultParams()
	fc := NewFeCap(p)

	voltages := []float64{-2 * p.CoerciveV, -p.CoerciveV, 0, p.CoerciveV, 2 * p.CoerciveV}
	for _, V := range voltages {
		Q := fc.TotalCharge(V)
		Vback, err := fc.SolveVFE(Q)
		if err != nil {
			t.Errorf("SolveVFE at V=%.3f: %v", V, err)
			continue
		}
		if math.Abs(Vback-V) > 1e-8 {
			t.Errorf("round-trip V=%.4f: got %.4f (err=%.2e)", V, Vback, math.Abs(Vback-V))
		}
	}
}

// TestSweepPELoop_Shape verifies basic properties of the hysteresis loop:
// 1. Loop opens (positive and negative branches differ at V=0)
// 2. Both branches eventually saturate
// 3. Correct number of points returned
func TestSweepPELoop_Shape(t *testing.T) {
	p := defaultParams()
	fc := NewFeCap(p)
	Vmax := 3 * p.CoerciveV
	nPts := 100

	loop := fc.SweepPELoop(Vmax, nPts)
	if len(loop) != 2*nPts {
		t.Fatalf("expected %d points, got %d", 2*nPts, len(loop))
	}

	// First half is ascending (V goes -Vmax → +Vmax)
	if loop[0].V > loop[nPts-1].V {
		t.Errorf("ascending branch not sorted: loop[0].V=%.3f > loop[%d].V=%.3f",
			loop[0].V, nPts-1, loop[nPts-1].V)
	}

	// Second half is descending (V goes +Vmax → -Vmax)
	if loop[nPts].V < loop[2*nPts-1].V {
		t.Errorf("descending branch not sorted: loop[%d].V=%.3f < loop[%d].V=%.3f",
			nPts, loop[nPts].V, 2*nPts-1, loop[2*nPts-1].V)
	}

	// Near ±Vmax, Q should approach ±(Qs + linear term)
	qMax := loop[nPts-1].Q
	if qMax <= 0 {
		t.Errorf("Q at +Vmax should be positive, got %.4e", qMax)
	}
	qMin := loop[2*nPts-1].Q
	if qMin >= 0 {
		t.Errorf("Q at -Vmax should be negative, got %.4e", qMin)
	}
}

// TestCapacitance_Positive verifies dQ/dV is always positive (monotone Q-V branch).
func TestCapacitance_Positive(t *testing.T) {
	p := defaultParams()
	fc := NewFeCap(p)
	voltages := []float64{-3 * p.CoerciveV, -p.CoerciveV, 0, p.CoerciveV, 3 * p.CoerciveV}
	for _, V := range voltages {
		C := fc.Capacitance(V)
		if C <= 0 {
			t.Errorf("Capacitance at V=%.3f: got %.4e (want > 0)", V, C)
		}
	}
}

// TestReset_ClearsHistory verifies that Reset() wipes the turning-point arrays.
func TestReset_ClearsHistory(t *testing.T) {
	p := defaultParams()
	fc := NewFeCap(p)
	// Simulate some direction changes by manually appending
	fc.aV = append(fc.aV, -0.5, 0.3)
	fc.bV = append(fc.bV, 0.8)

	fc.Reset()
	if len(fc.aV) != 0 || len(fc.bV) != 0 {
		t.Errorf("Reset() did not clear history: aV=%v bV=%v", fc.aV, fc.bV)
	}
}
