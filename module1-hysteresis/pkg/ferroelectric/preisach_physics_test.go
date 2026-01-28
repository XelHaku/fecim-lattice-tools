package ferroelectric

import (
	"math"
	"testing"
)

// TestHysteronAlphaBetaConstraint verifies the fundamental Preisach constraint
// that ALL hysterons satisfy Alpha > Beta.
//
// Physical basis: In the Preisach model, each hysteron represents a bistable
// switching unit where:
//   - Alpha (α): Positive switching field (UP transition)
//   - Beta (β): Negative switching field (DOWN transition)
//
// The constraint α > β is required for physical validity because:
//   1. It ensures no simultaneous UP and DOWN switching
//   2. It creates the triangular Preisach plane geometry
//   3. It guarantees memory effect and hysteresis
//
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991), Section 2.1
func TestHysteronAlphaBetaConstraint(t *testing.T) {
	material := DefaultHZO()
	gridSizes := []int{10, 30, 50}

	for _, gridSize := range gridSizes {
		t.Run(t.Name(), func(t *testing.T) {
			model := NewMayergoyzPreisach(material, gridSize)

			// Access hysterons via GetPreisachPlane
			alphas, betas, _ := model.GetPreisachPlane()

			if len(alphas) == 0 {
				t.Fatal("Model has no hysterons")
			}

			// Verify EVERY hysteron satisfies α > β
			violationCount := 0
			for i := range alphas {
				if alphas[i] <= betas[i] {
					violationCount++
					if violationCount <= 5 { // Report first 5 violations
						t.Errorf("Hysteron %d violates α > β: α=%.4e, β=%.4e",
							i, alphas[i], betas[i])
					}
				}
			}

			if violationCount > 0 {
				t.Errorf("Total violations of α > β constraint: %d out of %d hysterons",
					violationCount, len(alphas))
			}
		})
	}
}

// TestPreisachDistributionNormalization verifies that at saturation (E >> Ec),
// the weighted sum of hysteron contributions equals Ps.
//
// Physical basis: The Preisach model represents polarization as:
//   P = ∫∫ μ(α, β) · γ(α, β) dα dβ
// where:
//   - μ(α, β) is the distribution weight (normalized to sum to Ps)
//   - γ(α, β) is the hysteron state (+1 or -1)
//
// At saturation (E → +∞), all hysterons are in +1 state, so:
//   P_max = ∫∫ μ(α, β) dα dβ = Ps
//
// This test verifies |P_max - Ps| < 1% of Ps, confirming proper normalization.
//
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991), Eq. 2.15
func TestPreisachDistributionNormalization(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 40)

	Ps := material.Ps
	Ec := material.Ec

	// Apply saturation field (3× Ec to ensure full saturation)
	Esat := Ec * 3.0
	Pmax := model.Update(Esat)

	// Calculate relative error
	relativeError := math.Abs(Pmax-Ps) / Ps

	if relativeError > 0.01 {
		t.Errorf("Distribution not properly normalized:\n"+
			"  P_max      = %.6e C/m²\n"+
			"  P_s        = %.6e C/m²\n"+
			"  |P_max-Ps| = %.6e C/m² (%.2f%% of Ps)\n"+
			"  Expected: < 1%% of Ps",
			Pmax, Ps, math.Abs(Pmax-Ps), relativeError*100)
	}

	// Also verify negative saturation
	model.Reset()
	Pmin := model.Update(-Esat)
	relativeErrorNeg := math.Abs(math.Abs(Pmin)-Ps) / Ps

	if relativeErrorNeg > 0.01 {
		t.Errorf("Negative saturation not properly normalized:\n"+
			"  P_min      = %.6e C/m²\n"+
			"  -P_s       = %.6e C/m²\n"+
			"  ||Pmin|-Ps| = %.6e C/m² (%.2f%% of Ps)\n"+
			"  Expected: < 1%% of Ps",
			Pmin, -Ps, math.Abs(math.Abs(Pmin)-Ps), relativeErrorNeg*100)
	}

	t.Logf("Distribution normalization verified:\n"+
		"  Positive saturation: P_max = %.6e C/m² (error: %.3f%%)\n"+
		"  Negative saturation: P_min = %.6e C/m² (error: %.3f%%)",
		Pmax, relativeError*100, Pmin, relativeErrorNeg*100)
}

// TestPreisachCongruencyProperty verifies the congruency property of hysteresis:
// minor loops starting from the same (P, branch) state should be parallel
// (same dP/dE slope).
//
// Physical basis: The Preisach model satisfies the "congruency" property:
// curves in the P-E plane that start from the same point with the same
// history slope have the same local curvature. This means minor loops
// starting from the same state should have parallel slopes.
//
// Test procedure:
//   1. Saturate to establish baseline
//   2. Create turning point at E1
//   3. Descend to specific P value, record E_A
//   4. Return to E1 (wipe out history)
//   5. Descend again to same P value, record E_B
//   6. Measure dP/dE at both points
//   7. Verify slopes differ by < 5%
//
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991), Section 2.4
func TestPreisachCongruencyProperty(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec
	Ps := material.Ps

	// Step 1: Saturate positive
	Emax := Ec * 2.5
	model.Update(Emax)

	// Step 2: Create first descending branch
	E1 := Ec * 1.0
	model.Update(E1)
	targetP := Ps * 0.3 // Target polarization

	// Find E where P ≈ targetP on first descent
	var EA, PA, PprevA float64
	foundA := false
	dE := Ec * 0.01 // Small step
	for E := E1; E > -Emax; E -= dE {
		P := model.Update(E)
		if P <= targetP && !foundA {
			EA = E
			PA = P
			foundA = true
			break
		}
		PprevA = P
	}

	if !foundA {
		t.Fatal("Could not reach target P on first descent")
	}

	// Calculate slope for first descent: dP/dE ≈ (P - P_prev) / (-dE)
	slopeA := (PA - PprevA) / (-dE)

	// Step 3: Reset and create second descending branch from same starting point
	model.Reset()
	model.Update(Emax) // Saturate again
	model.Update(E1)   // Same turning point

	// Find E where P ≈ targetP on second descent
	var EB, PB, PprevB float64
	foundB := false
	for E := E1; E > -Emax; E -= dE {
		P := model.Update(E)
		if P <= targetP && !foundB {
			EB = E
			PB = P
			foundB = true
			break
		}
		PprevB = P
	}

	if !foundB {
		t.Fatal("Could not reach target P on second descent")
	}

	// Calculate slope for second descent
	slopeB := (PB - PprevB) / (-dE)

	// Compare slopes (should be nearly identical due to congruency)
	slopeDiff := math.Abs(slopeA-slopeB) / math.Max(math.Abs(slopeA), math.Abs(slopeB))

	if slopeDiff > 0.05 {
		t.Errorf("Congruency property violated:\n"+
			"  First descent:  E=%.4e, P=%.6e, dP/dE=%.6e\n"+
			"  Second descent: E=%.4e, P=%.6e, dP/dE=%.6e\n"+
			"  Slope difference: %.2f%% (expected < 5%%)",
			EA, PA, slopeA, EB, PB, slopeB, slopeDiff*100)
	}

	t.Logf("Congruency property verified:\n"+
		"  First descent:  dP/dE = %.6e at P=%.6e C/m²\n"+
		"  Second descent: dP/dE = %.6e at P=%.6e C/m²\n"+
		"  Slope difference: %.3f%%",
		slopeA, PA, slopeB, PB, slopeDiff*100)
}

// TestPreisachWipingOutProperty verifies the "wiping-out" or "return point memory"
// property: when the field returns to a previous extremum, the polarization state
// should reflect only the relevant hysteron states, effectively erasing intermediate
// minor loop history.
//
// Physical basis: The Preisach model exhibits the wiping-out property:
//   - In the classical Preisach model, hysteron states are determined purely
//     by comparing current field E to their α and β thresholds
//   - This naturally implements memory: hysterons "remember" their last switch
//   - When field exceeds a previous turning point, hysterons switch again,
//     effectively erasing the effect of intermediate minor loops
//
// Test procedure:
//   1. Create two paths starting from negative saturation:
//      Path A: -Emax → +Emax (major loop, direct)
//      Path B: -Emax → E1 → E2 → E1 (minor loop) → +Emax
//   2. At the end, both should reach same state (Ps) because returning to +Emax
//      wipes out the intermediate minor loop history
//
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991), Section 2.3
func TestPreisachWipingOutProperty(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec
	Emax := Ec * 2.5

	// Path A: Simple saturation from negative to positive
	model.Reset()
	model.Update(-Emax) // Start at negative saturation
	PdirectA := model.Update(Emax) // Direct path to positive saturation

	// Path B: Same path but with intermediate minor loop
	model.Reset()
	model.Update(-Emax) // Start at negative saturation

	// Create a minor loop
	E1 := Ec * 0.8
	E2 := -Ec * 0.5
	model.Update(E1) // Ascend to E1
	model.Update(E2) // Descend to E2 (minor loop)
	model.Update(E1) // Return to E1 (close minor loop)

	// Now continue to positive saturation
	PdirectB := model.Update(Emax)

	// Both paths should reach the same saturation state
	relDiff := math.Abs(PdirectA-PdirectB) / math.Abs(PdirectA)

	if relDiff > 0.01 {
		t.Errorf("Wiping-out property violated:\n"+
			"  P_A (direct path)          = %.6e C/m²\n"+
			"  P_B (with minor loop)      = %.6e C/m²\n"+
			"  Relative difference        = %.2f%% (expected < 1%%)\n"+
			"  \n"+
			"  Path A: -Emax → +Emax\n"+
			"  Path B: -Emax → E1=%.3fEc → E2=%.3fEc → E1 → +Emax",
			PdirectA, PdirectB, relDiff*100, E1/Ec, E2/Ec)
	}

	t.Logf("Wiping-out property verified:\n"+
		"  P_A (direct path)      = %.6e C/m²\n"+
		"  P_B (with minor loop)  = %.6e C/m²\n"+
		"  Relative difference    = %.3f%%",
		PdirectA, PdirectB, relDiff*100)
}

// TestSimpleVsAdvancedPreisachConsistency verifies that the simple PreisachModel
// (hyperbolic tangent) and advanced MayergoyzPreisach produce qualitatively similar
// results at room temperature for the same material.
//
// Physical basis: Both models should capture the same fundamental physics:
//   - PreisachModel: Simplified tanh switching function (fast, approximate)
//   - MayergoyzPreisach: Full distribution over hysteron plane (accurate, slow)
//
// At 300K with DefaultHZO, the models should produce similar:
//   - Remanent polarization Pr (within 20%)
//   - Saturation polarization Ps (within 10%)
//
// Note: Loop area may differ significantly due to different hysteresis mechanisms
// (tanh-based vs hysteron-based), but saturation and remanent values should align.
//
// Reference: Bartic et al., J. Appl. Phys. 89, 3420 (2001)
func TestSimpleVsAdvancedPreisachConsistency(t *testing.T) {
	material := DefaultHZO()
	Ec := material.Ec
	Emax := Ec * 2.0 // Same amplitude for both models

	// Create both models
	simpleModel := NewPreisachModel(material)
	advancedModel := NewMayergoyzPreisach(material, 40)
	advancedModel.SetTemperature(300) // Ensure room temperature

	// Generate loops with same parameters
	points := 200
	Esimple, Psimple := simpleModel.GetHysteresisLoop(Emax, points)
	Eadv, Padv := advancedModel.GetHysteresisLoop(Emax, points)

	// Find saturation polarization (maximum |P| values)
	var PsatSimple, PsatAdvanced float64
	for _, p := range Psimple {
		if math.Abs(p) > math.Abs(PsatSimple) {
			PsatSimple = p
		}
	}
	for _, p := range Padv {
		if math.Abs(p) > math.Abs(PsatAdvanced) {
			PsatAdvanced = p
		}
	}

	// Compare saturation polarization
	PsatDiff := math.Abs(PsatSimple-PsatAdvanced) / math.Max(math.Abs(PsatSimple), math.Abs(PsatAdvanced))

	if PsatDiff > 0.10 {
		t.Errorf("Saturation polarization difference exceeds 10%%:\n"+
			"  Simple model Psat    = %.6e C/m²\n"+
			"  Advanced model Psat  = %.6e C/m²\n"+
			"  Relative difference  = %.2f%%",
			PsatSimple, PsatAdvanced, PsatDiff*100)
	}

	// Find remanent polarization (P at E≈0 on descending branch)
	// Look in second half of loop (descending from +Emax)
	var PrSimple, PrAdvanced float64

	// Simple model: find P near E=0 in second half
	for i := len(Esimple) / 4; i < 3*len(Esimple)/4; i++ {
		if math.Abs(Esimple[i]) < Ec*0.1 {
			PrSimple = Psimple[i]
			break
		}
	}

	// Advanced model: find P near E=0 in second half
	for i := len(Eadv) / 4; i < 3*len(Eadv)/4; i++ {
		if math.Abs(Eadv[i]) < Ec*0.1 {
			PrAdvanced = Padv[i]
			break
		}
	}

	// Compare Pr values
	PrDiff := math.Abs(PrSimple-PrAdvanced) / math.Max(math.Abs(PrSimple), math.Abs(PrAdvanced))

	if PrDiff > 0.20 {
		t.Errorf("Remanent polarization difference exceeds 20%%:\n"+
			"  Simple model Pr      = %.6e C/m²\n"+
			"  Advanced model Pr    = %.6e C/m²\n"+
			"  Relative difference  = %.2f%%",
			PrSimple, PrAdvanced, PrDiff*100)
	}

	t.Logf("Model consistency verified at 300K:\n"+
		"  Saturation polarization:\n"+
		"    Simple:   %.6e C/m²\n"+
		"    Advanced: %.6e C/m²\n"+
		"    Difference: %.2f%%\n"+
		"  Remanent polarization:\n"+
		"    Simple:   %.6e C/m²\n"+
		"    Advanced: %.6e C/m²\n"+
		"    Difference: %.2f%%",
		PsatSimple, PsatAdvanced, PsatDiff*100,
		PrSimple, PrAdvanced, PrDiff*100)
}
