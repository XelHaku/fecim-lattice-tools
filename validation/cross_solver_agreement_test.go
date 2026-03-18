package validation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

// TestCrossSolverAgreement_PrEc verifies that the Landau-Khalatnikov (LK)
// thermodynamic solver and the Preisach phenomenological engine agree on
// remanent polarization (Pr) and coercive field (Ec) within documented bounds
// when driven through a full P-E hysteresis loop using the same HZO material.
//
// Physics context:
//   - LK is a time-domain, thermodynamic model: polarization evolves via
//     rho_eff * dP/dt = E - K_dep*P - dG/dP. It produces a continuous
//     trajectory through the Landau double-well landscape.
//   - Preisach is a rate-independent, phenomenological model: polarization is
//     a function of field history only, evaluated via the Everett integral
//     over a distribution of hysterons with thresholds (alpha, beta).
//   - The two models encode fundamentally different physics. Pr and Ec
//     agreement within 15-20% is physically meaningful evidence that both
//     engines are correctly calibrated to the same material.
//
// Tolerances:
//   - |Pr_LK - Pr_Preisach| / Pr_LK < 0.15 (15%)
//   - |Ec_LK - Ec_Preisach| / Ec_LK < 0.50 (50% — see note below)
//
// These bounds reflect expected model divergence, not measurement noise.
// The Preisach model includes a reversible (dielectric) contribution and
// NLS kinetics relaxation that shift its effective Pr/Ec relative to the
// pure Landau free-energy surface. LK's depolarization field (K_dep) also
// slants the loop in ways the Preisach Everett function does not replicate
// identically.
func TestCrossSolverAgreement_PrEc(t *testing.T) {
	mat := physics.DefaultHZO()
	if mat == nil {
		t.Fatal("DefaultHZO() returned nil")
	}

	// ----- Preisach engine: quasi-static hysteresis loop -----
	preisachModel := ferroelectric.NewPreisachModel(mat)
	if preisachModel == nil {
		t.Fatal("NewPreisachModel() returned nil")
	}

	eMax := mat.Ec * 2.5
	const loopPoints = 801

	ePreisach, pPreisach := preisachModel.GetHysteresisLoop(eMax, loopPoints)
	if len(ePreisach) < 10 || len(ePreisach) != len(pPreisach) {
		t.Fatalf("Preisach loop data invalid: len(E)=%d len(P)=%d", len(ePreisach), len(pPreisach))
	}

	prPreisachCandidates := interpolateYAtXZero(ePreisach, pPreisach)
	if len(prPreisachCandidates) < 2 {
		t.Fatalf("Preisach: could not extract Pr (E=0 crossings=%d)", len(prPreisachCandidates))
	}
	prPreisach := meanAbs(prPreisachCandidates)

	ecPreisachCandidates := interpolateXAtYZero(ePreisach, pPreisach)
	if len(ecPreisachCandidates) < 2 {
		t.Fatalf("Preisach: could not extract Ec (P=0 crossings=%d)", len(ecPreisachCandidates))
	}
	ecPreisach := meanAbs(ecPreisachCandidates)

	// ----- LK engine: time-domain hysteresis loop -----
	//
	// The LK solver is a time-domain integrator. To produce a quasi-static
	// hysteresis loop comparable to the Preisach output, we sweep the applied
	// field slowly enough that the polarization reaches near-equilibrium at
	// each field step.
	//
	// NLS and noise are disabled so the loop shape is governed purely by the
	// Landau free-energy landscape and depolarization.
	lk := physics.NewLKSolver()
	lk.ConfigureFromMaterial(mat)
	lk.UseNLS = false
	lk.EnableNoise = false
	// Reset to negative saturation for a clean sweep start.
	lk.SetState(-mat.Ps)

	// Sweep parameters: the field ramp rate and per-point dwell must be slow
	// enough for the RK4 integrator to track the Landau equilibrium. We use
	// many sub-steps per field point to achieve quasi-static conditions.
	const (
		lkFieldSteps = 400  // field points per half-sweep
		lkSubSteps   = 200  // integration sub-steps per field point
		lkDt         = 1e-9 // timestep per sub-step (1 ns)
	)

	eLK := make([]float64, 0, lkFieldSteps*4+2)
	pLK := make([]float64, 0, lkFieldSteps*4+2)

	// Drive to negative saturation first.
	for s := 0; s < lkSubSteps*5; s++ {
		lk.Step(-eMax, lkDt)
	}

	// Ascending: -Emax to +Emax
	for i := 0; i <= lkFieldSteps*2; i++ {
		e := -eMax + 2.0*eMax*float64(i)/float64(lkFieldSteps*2)
		for s := 0; s < lkSubSteps; s++ {
			lk.Step(e, lkDt)
		}
		eLK = append(eLK, e)
		pLK = append(pLK, lk.P)
	}

	// Descending: +Emax to -Emax
	for i := 1; i <= lkFieldSteps*2; i++ {
		e := eMax - 2.0*eMax*float64(i)/float64(lkFieldSteps*2)
		for s := 0; s < lkSubSteps; s++ {
			lk.Step(e, lkDt)
		}
		eLK = append(eLK, e)
		pLK = append(pLK, lk.P)
	}

	if len(eLK) < 10 {
		t.Fatalf("LK loop data too short: len=%d", len(eLK))
	}

	prLKCandidates := interpolateYAtXZero(eLK, pLK)
	if len(prLKCandidates) < 2 {
		t.Fatalf("LK: could not extract Pr (E=0 crossings=%d, loop len=%d)", len(prLKCandidates), len(eLK))
	}
	prLK := meanAbs(prLKCandidates)

	ecLKCandidates := interpolateXAtYZero(eLK, pLK)
	if len(ecLKCandidates) < 2 {
		// If P never crosses zero, the LK loop may be offset or unsaturated.
		// Log diagnostics and skip Ec assertion with an explanation.
		pMin, pMax := pLK[0], pLK[0]
		for _, p := range pLK {
			if p < pMin {
				pMin = p
			}
			if p > pMax {
				pMax = p
			}
		}
		t.Fatalf("LK: could not extract Ec (P=0 crossings=%d); Prange=[%.4e, %.4e] C/m²",
			len(ecLKCandidates), pMin, pMax)
	}
	ecLK := meanAbs(ecLKCandidates)

	// ----- Log all values for research documentation -----
	t.Logf("=== Cross-Solver Agreement: DefaultHZO ===")
	t.Logf("Material: %s", mat.Name)
	t.Logf("Material Pr=%.4f C/m² (%.2f µC/cm²), Ec=%.4e V/m (%.3f MV/cm)",
		mat.Pr, mat.Pr*100, mat.Ec, mat.Ec/1e8)
	t.Logf("")
	t.Logf("LK engine:")
	t.Logf("  Pr = %.4f C/m² (%.2f µC/cm²)", prLK, prLK*100)
	t.Logf("  Ec = %.4e V/m (%.3f MV/cm)", ecLK, ecLK/1e8)
	t.Logf("")
	t.Logf("Preisach engine:")
	t.Logf("  Pr = %.4f C/m² (%.2f µC/cm²)", prPreisach, prPreisach*100)
	t.Logf("  Ec = %.4e V/m (%.3f MV/cm)", ecPreisach, ecPreisach/1e8)
	t.Logf("")

	// ----- Assertions -----
	// Guard against degenerate zero values that would cause division-by-zero.
	if prLK < 1e-6 {
		t.Fatalf("LK Pr is near zero (%.4e C/m²); loop may not have saturated", prLK)
	}
	if ecLK < 1e3 {
		t.Fatalf("LK Ec is near zero (%.4e V/m); loop may not have formed", ecLK)
	}
	if prPreisach < 1e-6 {
		t.Fatalf("Preisach Pr is near zero (%.4e C/m²)", prPreisach)
	}
	if ecPreisach < 1e3 {
		t.Fatalf("Preisach Ec is near zero (%.4e V/m)", ecPreisach)
	}

	prRelDiff := math.Abs(prLK-prPreisach) / prLK
	ecRelDiff := math.Abs(ecLK-ecPreisach) / ecLK

	t.Logf("Relative differences:")
	t.Logf("  |Pr_LK - Pr_Preisach| / Pr_LK = %.4f (%.1f%%)", prRelDiff, prRelDiff*100)
	t.Logf("  |Ec_LK - Ec_Preisach| / Ec_LK = %.4f (%.1f%%)", ecRelDiff, ecRelDiff*100)

	const (
		prTolerance = 0.15 // 15% — models use fundamentally different physics
		ecTolerance = 0.50 // 50% — LK continuous switching starts before Ec (lower apparent Ec)
		// while Preisach sharp hysteron thresholds align with calibrated Ec.
		// This 30-50% Ec discrepancy is physically expected between thermodynamic
		// and phenomenological models. Pr agreement is the more meaningful metric.
	)

	if prRelDiff > prTolerance {
		t.Errorf("Pr disagreement exceeds %.0f%% tolerance: LK=%.4f Preisach=%.4f (diff=%.1f%%)",
			prTolerance*100, prLK, prPreisach, prRelDiff*100)
	}
	if ecRelDiff > ecTolerance {
		t.Errorf("Ec disagreement exceeds %.0f%% tolerance: LK=%.4e Preisach=%.4e (diff=%.1f%%)",
			ecTolerance*100, ecLK, ecPreisach, ecRelDiff*100)
	}

	if !t.Failed() {
		t.Logf("PASS: LK and Preisach engines agree within tolerances (Pr<%.0f%%, Ec<%.0f%%)",
			prTolerance*100, ecTolerance*100)
	}
}
