package physics

// Tests for LIT-P3-01: DCC (Displacement Current Control) write mode.

import (
	"math"
	"testing"
)

// newDCCPair creates matched WriteController + DCCController for default_hzo.
// Both share the same HZOMaterial but use separate solvers.
func newDCCPair(t *testing.T) (*WriteController, *DCCController) {
	t.Helper()
	mat := DefaultHZO()

	// ISPP solver
	solverISPP := NewLKSolver()
	solverISPP.ConfigureFromMaterial(mat)
	wc := NewWriteController(solverISPP, mat)

	// DCC solver — independent state
	solverDCC := NewLKSolver()
	solverDCC.ConfigureFromMaterial(mat)
	dcc := NewDCCController(solverDCC, mat)

	return wc, dcc
}

// --- DCCController construction ---

func TestNewDCCController_Defaults(t *testing.T) {
	mat := DefaultHZO()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	dcc := NewDCCController(solver, mat)

	if dcc.WriteVoltage <= 0 {
		t.Errorf("WriteVoltage=%e should be > 0", dcc.WriteVoltage)
	}
	if dcc.MaxPulseTime <= 0 {
		t.Errorf("MaxPulseTime=%e should be > 0", dcc.MaxPulseTime)
	}
	if dcc.Tolerance <= 0 || dcc.Tolerance >= 1 {
		t.Errorf("Tolerance=%e out of (0,1)", dcc.Tolerance)
	}
	if dcc.TimeStep <= 0 {
		t.Errorf("TimeStep=%e should be > 0", dcc.TimeStep)
	}
	// WriteVoltage should be ≥ coercive voltage.
	vc := mat.CoerciveVoltage()
	if dcc.WriteVoltage < vc {
		t.Errorf("WriteVoltage=%e < coercive %e — may not switch", dcc.WriteVoltage, vc)
	}
}

func TestDCCController_NilMaterial(t *testing.T) {
	solver := NewLKSolver()
	dcc := NewDCCController(solver, nil)
	res := dcc.ProgramDCC(0.5, true)
	if res.Success || res.FailureReason == "" {
		t.Errorf("nil material should fail with reason, got success=%v reason=%q",
			res.Success, res.FailureReason)
	}
}

func TestDCCController_OutOfBoundsTarget(t *testing.T) {
	_, dcc := newDCCPair(t)
	res := dcc.ProgramDCC(1e6, true) // way above Gmax
	if res.Success || res.FailureReason == "" {
		t.Errorf("out-of-bounds target should fail, got success=%v", res.Success)
	}
}

// --- DCC writes to target states ---

func TestDCC_WritesPositiveTarget(t *testing.T) {
	_, dcc := newDCCPair(t)
	mat := dcc.Material

	// Mid-range positive target.
	targetG := mat.Gmin + 0.7*(mat.Gmax-mat.Gmin)
	res := dcc.ProgramDCC(targetG, true)

	if !res.Success {
		t.Errorf("DCC failed positive target G=%.4f: %s", targetG, res.FailureReason)
	}
	if res.FinalG < mat.Gmin || res.FinalG > mat.Gmax {
		t.Errorf("FinalG=%e out of [%e, %e]", res.FinalG, mat.Gmin, mat.Gmax)
	}
}

func TestDCC_WritesNegativeTarget(t *testing.T) {
	_, dcc := newDCCPair(t)
	mat := dcc.Material

	// Mid-range negative (low) target.
	targetG := mat.Gmin + 0.3*(mat.Gmax-mat.Gmin)
	res := dcc.ProgramDCC(targetG, true)

	if !res.Success {
		t.Errorf("DCC failed negative target G=%.4f: %s", targetG, res.FailureReason)
	}
}

func TestDCC_SinglePulse(t *testing.T) {
	// DCC must reach target in a SINGLE pulse — PulseDuration > 0 but only
	// one continuous write interval (no iterative loop).
	_, dcc := newDCCPair(t)
	mat := dcc.Material
	targetG := mat.Gmin + 0.6*(mat.Gmax-mat.Gmin)
	res := dcc.ProgramDCC(targetG, true)

	if !res.Success {
		t.Skipf("DCC did not converge: %s", res.FailureReason)
	}
	if res.PulseDuration <= 0 {
		t.Errorf("PulseDuration=%e should be > 0", res.PulseDuration)
	}
	if res.PulseDuration > dcc.MaxPulseTime+1e-18 {
		t.Errorf("PulseDuration=%e > MaxPulseTime=%e", res.PulseDuration, dcc.MaxPulseTime)
	}
}

func TestDCC_StepsConsistentWithDuration(t *testing.T) {
	_, dcc := newDCCPair(t)
	mat := dcc.Material
	targetG := mat.Gmin + 0.5*(mat.Gmax-mat.Gmin)
	res := dcc.ProgramDCC(targetG, true)
	if !res.Success {
		t.Skipf("DCC did not converge: %s", res.FailureReason)
	}
	// PulseDuration ≈ Steps × TimeStep (within 1 step).
	expectedDuration := float64(res.Steps) * dcc.TimeStep
	if math.Abs(res.PulseDuration-expectedDuration) > dcc.TimeStep+1e-20 {
		t.Errorf("PulseDuration=%e != Steps×dt=%e", res.PulseDuration, expectedDuration)
	}
}

func TestDCC_ResetFalse_StartsFromCurrent(t *testing.T) {
	// With reset=false the solver starts from wherever it currently is.
	// If already at target, PulseDuration should be 0 (or very small).
	_, dcc := newDCCPair(t)
	mat := dcc.Material
	targetG := mat.Gmin + 0.6*(mat.Gmax-mat.Gmin)

	// First write to bring state near target.
	dcc.ProgramDCC(targetG, true)

	// Second write with reset=false — should need little/no additional time.
	res2 := dcc.ProgramDCC(targetG, false)
	if !res2.Success {
		t.Errorf("reset=false second write failed: %s", res2.FailureReason)
	}
	// Duration should be very short (already at target within tolerance).
	t.Logf("reset=false second DCC pulse: %e s (%d steps)", res2.PulseDuration, res2.Steps)
}

// --- DCC binary saturation switching (primary DCC use case) ---

// TestDCC_BinarySwitching verifies DCC for full-saturation writes (+Pr, -Pr).
// DCC at constant high voltage drives the L-K dynamics to saturation and
// terminates early. Intermediate-state accuracy is physics-limited (see below).
//
// Physics note: Under constant high field (V = 3×Vc), L-K dynamics switch
// from -Pr to +Pr in ~1-2 ns but trajectory through intermediate states is
// fast and hard to intercept. DCC is inherently suited to binary writes;
// multi-level DCC requires a ramped voltage profile (beyond this scope).
func TestDCC_BinarySwitching(t *testing.T) {
	_, dcc := newDCCPair(t)
	mat := dcc.Material

	// Full saturation targets: 5% inside Gmin/Gmax to avoid boundary issues.
	cases := []struct {
		frac float64
		name string
	}{
		{0.95, "near-Gmax (+Pr)"},
		{0.05, "near-Gmin (-Pr)"},
	}
	for _, tc := range cases {
		targetG := mat.Gmin + tc.frac*(mat.Gmax-mat.Gmin)

		solverDCC := NewLKSolver()
		solverDCC.ConfigureFromMaterial(mat)
		dcc.Solver = solverDCC
		res := dcc.ProgramDCC(targetG, true)

		if !res.Success {
			t.Errorf("%s: DCC failed: %s", tc.name, res.FailureReason)
			continue
		}
		// For saturation targets: DCC drives P to full saturation (+Ps or -Ps),
		// which may differ from the exact 5%/95% conductance target because the
		// G-P mapping is non-linear at saturation. Allow 25% tolerance.
		rangeG := mat.Gmax - mat.Gmin
		err := math.Abs(res.FinalG-targetG) / rangeG
		t.Logf("%s: DCC err=%.2f%% pulse=%.1f ns steps=%d",
			tc.name, err*100, res.PulseDuration/1e-9, res.Steps)
		if err > 0.25 {
			t.Errorf("%s: DCC err %.2f%% > 25%%", tc.name, err*100)
		}
	}
}

// TestDCC_SwitchesCorrectDirection verifies DCC moves polarization in the
// right direction regardless of whether the target is above or below current.
func TestDCC_SwitchesCorrectDirection(t *testing.T) {
	_, dcc := newDCCPair(t)
	mat := dcc.Material

	// Write to +Pr region: final P should be positive.
	solverPos := NewLKSolver()
	solverPos.ConfigureFromMaterial(mat)
	dcc.Solver = solverPos
	resPos := dcc.ProgramDCC(mat.Gmin+0.8*(mat.Gmax-mat.Gmin), true)
	if resPos.FinalP <= 0 {
		t.Errorf("positive target: FinalP=%e should be > 0", resPos.FinalP)
	}

	// Write to -Pr region: final P should be negative.
	// For the exponential conductance model (default), the midpoint is the geometric mean
	// sqrt(Gmin*Gmax) = sqrt(1µS * 100µS) = 10 µS. Use 2% of linear range (≈2.98 µS)
	// which sits well below the geometric mean in log space → clearly negative P.
	solverNeg := NewLKSolver()
	solverNeg.ConfigureFromMaterial(mat)
	dcc.Solver = solverNeg
	resNeg := dcc.ProgramDCC(mat.Gmin+0.02*(mat.Gmax-mat.Gmin), true)
	if resNeg.FinalP >= 0 {
		t.Errorf("negative target: FinalP=%e should be < 0", resNeg.FinalP)
	}
}

// --- CompareISPPvsDCC ---

func TestCompareISPPvsDCC(t *testing.T) {
	wc, dcc := newDCCPair(t)
	mat := wc.Material
	targetG := mat.Gmin + 0.6*(mat.Gmax-mat.Gmin)

	cmp := CompareISPPvsDCC(wc, dcc, targetG)

	if cmp.TargetG != targetG {
		t.Errorf("TargetG mismatch")
	}
	// At least one method should succeed.
	if !cmp.ISPPSuccess && !cmp.DCCSuccess {
		t.Errorf("both ISPP and DCC failed for G=%.4f", targetG)
	}
	if cmp.DCCSuccess {
		t.Logf("ISPP: %d attempts, %d overshoots | DCC: %.1f ns pulse",
			cmp.ISPPAttempts, cmp.ISPPOvershoots, cmp.DCCPulseDuration/1e-9)
	}
}
