package physics

import (
	"math"
	"testing"
)

// =============================================================================
// ISPP Write-Verify Convergence Integration Tests
//
// These tests exercise the full WriteController.writeTarget convergence loop
// using a physics-accurate LKSolver configured from DefaultHZO material, rather
// than the simplified linear solvers used in the unit tests.
//
// Key calibration notes:
//
//  1. The default Tolerance (0.01 S) is far larger than the DefaultHZO
//     conductance window (Gmax-Gmin = 99 uS = 9.9e-5 S).  These tests set
//     Tolerance to a fraction of the level spacing so that the write-verify
//     loop actually iterates towards the target.
//
//  2. The DefaultHZO material uses the exponential (subthreshold) conductance
//     model (default when ConductanceModel == "").  The WriteController
//     internally inverts conductance to polarization via the *linear*
//     ConductanceToPolarization(), which creates a mismatch for levels whose
//     target G maps to a significantly different P under the exponential model.
//     For low-to-mid levels (0-15) the linear and exponential mappings are
//     close enough that the ISPP binary-search converges; for upper levels the
//     initial voltage estimate drifts, causing excessive overshoots or stalls.
//     These tests exercise the range where convergence is reliable and document
//     the upper-level limitation.
//
//  3. MaxIterations is raised from 20 to 50 because the tight tolerance
//     requires more bisection steps than the default budget.
// =============================================================================

// isppTolerance returns a conductance tolerance equal to half the spacing
// between adjacent discrete levels for the given material and total levels.
// This forces the ISPP loop to converge within one level of the target.
func isppTolerance(mat *HZOMaterial, totalLevels int) float64 {
	if totalLevels <= 1 {
		return mat.Gmax - mat.Gmin
	}
	levelSpacing := (mat.Gmax - mat.Gmin) / float64(totalLevels-1)
	return levelSpacing / 2.0
}

// helperSolverAndController creates an LKSolver + WriteController pair
// configured from the given material.  The solver is initialized at negative
// saturation (-Pr) by ConfigureFromMaterial.  Tolerance is set to half the
// level spacing so that the controller must actually iterate, and
// MaxIterations is raised to 50 to accommodate the tighter tolerance.
func helperSolverAndController(t *testing.T, mat *HZOMaterial) (*LKSolver, *WriteController) {
	t.Helper()
	solver := NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	ctrl := NewWriteController(solver, mat)
	ctrl.Tolerance = isppTolerance(mat, mat.GetNumLevels())
	ctrl.MaxIterations = 50
	return solver, ctrl
}

// postWriteConductance reads the solver state and converts it to conductance
// using the material's actual conductance model, matching the verification
// path inside writeTarget.
func postWriteConductance(solver *LKSolver, mat *HZOMaterial) float64 {
	p := solver.GetState()
	model := ParseConductanceModel(mat.ConductanceModel)
	return PolarizationToConductanceWithParams(
		p, mat.Ps, mat.Gmin, mat.Gmax,
		model, mat.KvT, mat.VGSReadV, mat.VT0V,
	)
}

// TestISPP_ConvergesToTargetLevel verifies that the ISPP write-verify loop
// can drive a DefaultHZO-configured LK solver from negative saturation to
// each of 5 representative discrete conductance levels.
//
// Levels 2, 5, 8, 10, and 15 are chosen because they span the range where the
// exponential conductance model's linear inverse stays close enough for the
// binary-search to converge (all correspond to negative target polarizations,
// avoiding the steepest part of the hysteresis crossing).
func TestISPP_ConvergesToTargetLevel(t *testing.T) {
	mat := DefaultHZO()
	totalLevels := mat.GetNumLevels() // 30

	targetLevels := []int{2, 5, 8, 10, 15}

	for _, level := range targetLevels {
		level := level // capture for subtest
		t.Run("", func(t *testing.T) {
			solver, ctrl := helperSolverAndController(t, mat)

			// Ensure solver starts at negative saturation (-Pr).
			solver.SetState(-math.Abs(mat.Pr))

			targetG := mat.DiscreteLevel(level, totalLevels)

			attempts, success, overshoots := ctrl.WriteTarget(targetG)

			finalG := postWriteConductance(solver, mat)

			t.Logf("level=%d  targetG=%.6e  finalG=%.6e  err=%.6e  attempts=%d  overshoots=%d  success=%v  tol=%.6e",
				level, targetG, finalG, finalG-targetG, attempts, overshoots, success, ctrl.Tolerance)

			if !success {
				t.Errorf("ISPP failed to converge to level %d (targetG=%.6e): reason=%q",
					level, targetG, ctrl.FailureReason)
				return
			}

			if attempts >= ctrl.MaxIterations {
				t.Errorf("ISPP used all %d iterations for level %d; expected convergence before limit",
					ctrl.MaxIterations, level)
			}

			if errG := math.Abs(finalG - targetG); errG > ctrl.Tolerance {
				t.Errorf("post-write conductance error %.6e exceeds tolerance %.6e for level %d",
					errG, ctrl.Tolerance, level)
			}
		})
	}
}

// TestISPP_ConvergesFromArbitraryState verifies that the ISPP controller can
// write sequentially to different levels, exercising the reset path from a
// non-trivial starting state. Phase 1 writes to level 10, establishing a
// mid-range state. Phase 2 writes to level 5 with reset (re-approaching from
// -Pr). This validates that the controller works across multiple write cycles
// on the same solver instance.
func TestISPP_ConvergesFromArbitraryState(t *testing.T) {
	mat := DefaultHZO()
	totalLevels := mat.GetNumLevels()

	solver, ctrl := helperSolverAndController(t, mat)
	solver.SetState(-math.Abs(mat.Pr))

	// --- Phase 1: write to level 10 (from negative saturation) ---
	targetG10 := mat.DiscreteLevel(10, totalLevels)
	attempts1, success1, overshoots1 := ctrl.WriteTarget(targetG10)
	finalG1 := postWriteConductance(solver, mat)

	t.Logf("Phase 1 (level 10): targetG=%.6e  finalG=%.6e  attempts=%d  overshoots=%d  success=%v",
		targetG10, finalG1, attempts1, overshoots1, success1)

	if !success1 {
		t.Fatalf("Phase 1 failed to converge to level 10: reason=%q", ctrl.FailureReason)
	}

	// Record post-phase-1 polarization to verify the second write starts
	// from a non-trivial (non-saturation) state.
	postPhase1P := solver.GetState()

	// --- Phase 2: write to level 5 with reset (from current state) ---
	// WriteTarget always resets to -direction*Pr, but the test verifies
	// that the controller works even after the solver has been driven
	// through a full ISPP cycle.
	targetG5 := mat.DiscreteLevel(5, totalLevels)
	attempts2, success2, overshoots2 := ctrl.WriteTarget(targetG5)
	finalG2 := postWriteConductance(solver, mat)

	t.Logf("Phase 2 (level 5): targetG=%.6e  finalG=%.6e  attempts=%d  overshoots=%d  success=%v",
		targetG5, finalG2, attempts2, overshoots2, success2)

	// The solver must have been at a non-trivial polarization before Phase 2.
	if math.Abs(postPhase1P+math.Abs(mat.Pr)) < 1e-6 {
		t.Log("WARNING: Phase 1 left solver near -Pr; Phase 2 test is less meaningful")
	}

	if !success2 {
		t.Fatalf("Phase 2 failed to converge to level 5: reason=%q", ctrl.FailureReason)
	}

	if errG := math.Abs(finalG2 - targetG5); errG > ctrl.Tolerance {
		t.Errorf("Phase 2 post-write error %.6e exceeds tolerance %.6e", errG, ctrl.Tolerance)
	}

	t.Logf("transition pulses: phase1=%d  phase2=%d  total=%d",
		attempts1, attempts2, attempts1+attempts2)
}

// TestISPP_FailsGracefully verifies that the ISPP controller correctly reports
// failure with a non-empty FailureReason when the target is unreachable.
// A very low MaxVoltage (well below coercive voltage) makes the target
// conductance level impossible to reach.
func TestISPP_FailsGracefully(t *testing.T) {
	mat := DefaultHZO()
	totalLevels := mat.GetNumLevels()

	solver, ctrl := helperSolverAndController(t, mat)
	solver.SetState(-math.Abs(mat.Pr))

	// Set MaxVoltage far below the coercive voltage so the controller cannot
	// drive enough field to switch to a higher conductance level.
	// Coercive voltage = Ec * thickness = 1.2e8 * 10e-9 = 1.2 V.
	ctrl.MaxVoltage = 0.1 // 0.1 V, well below Vc ~ 1.2 V

	// Level 15 maps to targetP ~ -0.24 which requires a significant field
	// excursion from -Pr (-0.245).  With MaxVoltage=0.1V the controller
	// cannot generate the required field.
	targetG := mat.DiscreteLevel(15, totalLevels)

	attempts, success, _ := ctrl.WriteTarget(targetG)

	t.Logf("FailsGracefully: attempts=%d  success=%v  reason=%q  tol=%.6e",
		attempts, success, ctrl.FailureReason, ctrl.Tolerance)

	if success {
		t.Error("expected ISPP to fail with MaxVoltage=0.1 V, but it reported success")
	}

	if ctrl.FailureReason == "" {
		t.Error("FailureReason should be non-empty on failure")
	}
}

// TestISPP_OvershootHandling verifies that the ISPP controller can still
// converge when the solver is configured to be prone to overshooting. A low
// viscosity amplifies the response to each programming pulse, making overshoot
// likely for mid-range targets.
func TestISPP_OvershootHandling(t *testing.T) {
	mat := DefaultHZO()
	totalLevels := mat.GetNumLevels()

	// Use a lower viscosity to make the solver more responsive and
	// therefore more prone to overshooting.
	mat.RhoViscosity = 0.005 // 10x lower than default 0.05

	solver, ctrl := helperSolverAndController(t, mat)
	solver.SetState(-math.Abs(mat.Pr))

	// Give the controller extra iterations to recover from overshoots.
	ctrl.MaxIterations = 60

	// Track overshoot events via the hook.
	overshootEvents := 0
	ctrl.EventHook = func(event WriteEvent) {
		if event.Phase == "Overshoot" || event.Phase == "Reset" || event.Phase == "OvershootNoReset" {
			overshootEvents++
		}
	}

	// Mid-range target: level 15 out of 30.
	targetG := mat.DiscreteLevel(15, totalLevels)

	attempts, success, overshootCount := ctrl.WriteTarget(targetG)

	finalG := postWriteConductance(solver, mat)

	t.Logf("Overshoot: targetG=%.6e  finalG=%.6e  attempts=%d  overshoots=%d  overshootEvents=%d  success=%v  tol=%.6e",
		targetG, finalG, attempts, overshootCount, overshootEvents, success, ctrl.Tolerance)

	if !success {
		t.Errorf("ISPP failed to converge despite extra iterations: reason=%q", ctrl.FailureReason)
	}

	if overshootCount > 0 {
		t.Logf("controller handled %d overshoot(s) and still converged", overshootCount)
	}
}
