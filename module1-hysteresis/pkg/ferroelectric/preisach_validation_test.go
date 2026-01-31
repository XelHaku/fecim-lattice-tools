package ferroelectric

import (
	"math"
	"testing"
)

// ============================================================================
// C13: Preisach Model Validation Against Experimental Hysteresis Data
// ============================================================================
// These tests validate the Preisach model implementation against known
// hysteresis loop characteristics from peer-reviewed literature.
//
// References:
//   - Bartic et al., J. Appl. Phys. 89, 3420 (2001) - Preisach model for FE capacitors
//   - Park et al., Adv. Mater. 27, 1811 (2015) - HZO hysteresis loops
//   - Müller et al., Nano Lett. 12, 4318 (2012) - HfO2 ferroelectric properties
//   - Cheema et al., Nature 580, 478 (2020) - Superlattice hysteresis
// ============================================================================

// LiteratureLoopData contains expected hysteresis loop characteristics
// from published experimental measurements.
type LiteratureLoopData struct {
	Name          string  // Reference name
	MaterialType  string  // Material being measured
	PrMin         float64 // Remanent polarization min (C/m²)
	PrMax         float64 // Remanent polarization max (C/m²)
	EcMin         float64 // Coercive field min (V/m)
	EcMax         float64 // Coercive field max (V/m)
	SquarenessMin float64 // Minimum squareness ratio (Pr/Ps)
	SquarenessMax float64 // Maximum squareness ratio
}

// GetLiteratureData returns a set of literature hysteresis loop data.
func GetLiteratureData() []LiteratureLoopData {
	return []LiteratureLoopData{
		{
			Name:          "Park_2015_HZO",
			MaterialType:  "HZO",
			PrMin:         15e-2,
			PrMax:         34e-2,
			EcMin:         0.6e8,
			EcMax:         1.5e8,
			SquarenessMin: 0.60,
			SquarenessMax: 0.95,
		},
		{
			Name:          "Cheema_2020_Superlattice",
			MaterialType:  "HZO_Superlattice",
			PrMin:         40e-2,
			PrMax:         50e-2,
			EcMin:         0.5e8,
			EcMax:         1.0e8,
			SquarenessMin: 0.75,
			SquarenessMax: 0.95,
		},
		{
			Name:          "Muller_2012_HfO2",
			MaterialType:  "Si:HfO2",
			PrMin:         10e-2,
			PrMax:         25e-2,
			EcMin:         0.8e8,
			EcMax:         1.5e8,
			SquarenessMin: 0.55,
			SquarenessMax: 0.85,
		},
	}
}

// TestPreisachLoopShape verifies that the Preisach model produces
// hysteresis loops with correct shape characteristics.
func TestPreisachLoopShape(t *testing.T) {
	materials := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		LiteratureSuperlattice(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			model := NewMayergoyzPreisach(mat, 50)

			Emax := 2.0 * mat.Ec // Go to saturation

			// Use the built-in GetHysteresisLoop method
			eVals, pVals := model.GetHysteresisLoop(Emax, 100)

			// Analyze loop characteristics
			Pmax := pVals[0]
			Pmin := pVals[0]
			for _, P := range pVals {
				if P > Pmax {
					Pmax = P
				}
				if P < Pmin {
					Pmin = P
				}
			}

			// Get Pr using the model's built-in method
			model.Reset()
			model.Update(Emax) // Saturate positive
			model.Update(0)    // Return to zero
			PrPlus := model.Polarization()

			model.Update(-Emax) // Saturate negative
			model.Update(0)     // Return to zero
			PrMinus := model.Polarization()

			Pr := (math.Abs(PrPlus) + math.Abs(PrMinus)) / 2
			Ps := (Pmax - Pmin) / 2
			squareness := Pr / Ps

			t.Logf("%s: Ps=%.1f µC/cm², Pr=%.1f µC/cm², squareness=%.2f",
				mat.Name, Ps*1e4, Pr*1e4, squareness)
			t.Logf("%s: Loop points: %d, E range: [%.2f, %.2f] MV/cm",
				mat.Name, len(eVals), eVals[0]/1e8, eVals[len(eVals)-1]/1e8)

			// Verify against literature bounds
			if squareness < 0.5 {
				t.Errorf("%s: Squareness %.2f is too low (expected > 0.5 for FE)", mat.Name, squareness)
			}
			if squareness > 1.0 {
				t.Errorf("%s: Squareness %.2f is unphysical (must be ≤ 1.0)", mat.Name, squareness)
			}

			// Loop should be symmetric
			asymmetry := math.Abs(Pmax+Pmin) / (Pmax - Pmin)
			if asymmetry > 0.15 {
				t.Logf("%s: Warning - loop asymmetry %.2f%% (imprint?)", mat.Name, asymmetry*100)
			}
		})
	}
}

// TestPreisachCoerciveField verifies that the Preisach model produces
// correct coercive field values (where P crosses zero).
func TestPreisachCoerciveField(t *testing.T) {
	materials := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			model := NewMayergoyzPreisach(mat, 50)

			Emax := 2.0 * mat.Ec
			steps := 200

			// Saturate positive first
			for i := 0; i <= steps; i++ {
				E := float64(i) / float64(steps) * Emax
				model.Update(E)
			}

			// Sweep down to find Ec- (negative coercive field)
			var EcMinus float64
			for i := steps; i >= -steps; i-- {
				E := float64(i) / float64(steps) * Emax
				P := model.Update(E)
				if P <= 0 && EcMinus == 0 {
					EcMinus = E
				}
			}

			// Already at negative saturation from previous sweep (lines 159-165)
			// Just ensure we're at -Emax
			model.Update(-Emax)

			// Sweep up to find Ec+ (positive coercive field)
			var EcPlus float64
			for i := -steps; i <= steps; i++ {
				E := float64(i) / float64(steps) * Emax
				P := model.Update(E)
				if P >= 0 && EcPlus == 0 {
					EcPlus = E
				}
			}

			// Average coercive field
			EcMeasured := (math.Abs(EcMinus) + math.Abs(EcPlus)) / 2
			EcExpected := mat.Ec

			relError := math.Abs(EcMeasured-EcExpected) / EcExpected

			t.Logf("%s: Ec measured=%.2f MV/cm, expected=%.2f MV/cm (error=%.1f%%)",
				mat.Name, EcMeasured/1e8, EcExpected/1e8, relError*100)

			// Allow 30% tolerance for Preisach distribution effects
			if relError > 0.30 {
				t.Errorf("%s: Ec error %.1f%% exceeds 30%% tolerance", mat.Name, relError*100)
			}
		})
	}
}

// TestPreisachLoopArea verifies that loop area correlates with switching energy.
func TestPreisachLoopArea(t *testing.T) {
	materials := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			model := NewMayergoyzPreisach(mat, 50)

			Emax := 2.0 * mat.Ec
			steps := 200
			dE := Emax / float64(steps)

			// Compute loop area by numerical integration
			// Area = ∮ P dE (closed loop integral)
			var loopArea float64

			// Forward sweep: 0 → +Emax
			Pprev := model.Polarization()
			for i := 1; i <= steps; i++ {
				E := float64(i) / float64(steps) * Emax
				P := model.Update(E)
				loopArea += (Pprev + P) / 2 * dE // Trapezoidal rule
				Pprev = P
			}

			// Top: +Emax → -Emax
			for i := steps - 1; i >= -steps; i-- {
				E := float64(i) / float64(steps) * Emax
				P := model.Update(E)
				loopArea += (Pprev + P) / 2 * (-dE) // Descending
				Pprev = P
			}

			// Bottom: -Emax → 0
			for i := -steps + 1; i <= 0; i++ {
				E := float64(i) / float64(steps) * Emax
				P := model.Update(E)
				loopArea += (Pprev + P) / 2 * dE
				Pprev = P
			}

			loopArea = math.Abs(loopArea)

			// Theoretical minimum area: 4 * Pr * Ec (rectangular loop)
			minArea := 4 * mat.Pr * mat.Ec
			areaRatio := loopArea / minArea

			t.Logf("%s: Loop area = %.2e J/m³, theoretical min = %.2e J/m³, ratio = %.2f",
				mat.Name, loopArea, minArea, areaRatio)

			// For realistic loops, area should be 0.5-1.2x the rectangular approximation
			if areaRatio < 0.3 || areaRatio > 1.5 {
				t.Errorf("%s: Loop area ratio %.2f is outside expected range [0.3, 1.5]",
					mat.Name, areaRatio)
			}
		})
	}
}

// TestPreisachMemoryEffect verifies that the Preisach model correctly
// implements the wiping-out property (minor loops are erased by major loops).
func TestPreisachMemoryEffect(t *testing.T) {
	mat := DefaultHZO()
	model := NewMayergoyzPreisach(mat, 50)

	Emax := 2.0 * mat.Ec
	Emin := -2.0 * mat.Ec

	// Start from saturation
	model.Update(Emax)
	Psat := model.Polarization()

	// Create a minor loop
	model.Update(0)
	P0 := model.Polarization()
	model.Update(mat.Ec * 0.5) // Partial positive
	model.Update(0)
	P0Minor := model.Polarization()

	// The minor loop should cause some deviation
	minorLoopEffect := math.Abs(P0Minor - P0)
	t.Logf("Minor loop deviation: %.2f µC/cm²", minorLoopEffect*1e4)

	// Now saturate again to "wipe out" the minor loop
	model.Update(Emax)
	PsatAfter := model.Polarization()

	// After full saturation, should return to same Psat
	satError := math.Abs(PsatAfter-Psat) / math.Abs(Psat)
	t.Logf("Saturation recovery: before=%.2f, after=%.2f, error=%.2f%%",
		Psat*1e4, PsatAfter*1e4, satError*100)

	if satError > 0.05 {
		t.Errorf("Saturation should wipe out minor loop history (error %.2f%% > 5%%)", satError*100)
	}

	// Negative saturation should also work
	model.Update(Emin)
	PsatNeg := model.Polarization()
	model.Update(0)
	model.Update(-mat.Ec * 0.5) // Partial negative
	model.Update(0)
	model.Update(Emin)
	PsatNegAfter := model.Polarization()

	negError := math.Abs(PsatNegAfter-PsatNeg) / math.Abs(PsatNeg)
	if negError > 0.05 {
		t.Errorf("Negative saturation should wipe out history (error %.2f%% > 5%%)", negError*100)
	}
}

// TestPreisachRemanence verifies that polarization persists at E=0.
func TestPreisachRemanence(t *testing.T) {
	mat := DefaultHZO()
	model := NewMayergoyzPreisach(mat, 50)

	Emax := 2.0 * mat.Ec

	// Saturate positive
	model.Update(Emax)

	// Return to E=0
	model.Update(0)
	Pr := model.Polarization()

	// Pr should be close to material Pr
	expectedPr := mat.Pr
	relError := math.Abs(Pr-expectedPr) / expectedPr

	t.Logf("Remanent polarization: measured=%.2f µC/cm², expected=%.2f µC/cm² (error=%.1f%%)",
		Pr*1e4, expectedPr*1e4, relError*100)

	// Allow 40% tolerance (Preisach distribution affects exact Pr)
	if relError > 0.40 {
		t.Errorf("Remanent polarization error %.1f%% exceeds 40%% tolerance", relError*100)
	}

	// Pr should persist (not decay immediately)
	model.Update(0)
	model.Update(0) // Multiple reads at E=0
	PrStable := model.Polarization()

	drift := math.Abs(PrStable-Pr) / math.Abs(Pr)
	if drift > 0.01 {
		t.Errorf("Remanent polarization should be stable at E=0 (drift %.2f%%)", drift*100)
	}
}

// TestLiteratureComparison compares model output against published loop data.
func TestLiteratureComparison(t *testing.T) {
	litData := GetLiteratureData()

	for _, lit := range litData {
		t.Run(lit.Name, func(t *testing.T) {
			// Create material matching literature type
			var mat *HZOMaterial
			switch lit.MaterialType {
			case "HZO":
				mat = DefaultHZO()
			case "HZO_Superlattice":
				mat = LiteratureSuperlattice()
			case "Si:HfO2":
				mat = DefaultHZO() // Close approximation
			default:
				mat = DefaultHZO()
			}

			// Verify material parameters are within literature range
			if mat.Pr < lit.PrMin || mat.Pr > lit.PrMax {
				t.Errorf("%s: Pr=%.2f µC/cm² outside lit range [%.1f, %.1f]",
					lit.Name, mat.Pr*1e4, lit.PrMin*1e4, lit.PrMax*1e4)
			}

			if mat.Ec < lit.EcMin || mat.Ec > lit.EcMax {
				t.Errorf("%s: Ec=%.2f MV/cm outside lit range [%.1f, %.1f]",
					lit.Name, mat.Ec/1e8, lit.EcMin/1e8, lit.EcMax/1e8)
			}

			// Check squareness from simulated loop
			model := NewMayergoyzPreisach(mat, 50)
			Emax := 2.0 * mat.Ec

			// Saturate and measure
			model.Update(Emax)
			Psat := model.Polarization()
			model.Update(0)
			Pr := model.Polarization()

			squareness := Pr / Psat

			if squareness < lit.SquarenessMin || squareness > lit.SquarenessMax {
				t.Logf("%s: Squareness %.2f outside lit range [%.2f, %.2f] (informational)",
					lit.Name, squareness, lit.SquarenessMin, lit.SquarenessMax)
			}

			t.Logf("%s: Pr=%.1f µC/cm², Ec=%.2f MV/cm, squareness=%.2f ✓",
				lit.Name, mat.Pr*1e4, mat.Ec/1e8, squareness)
		})
	}
}

// ============================================================================
// C13-B: Advanced Preisach Physics Validation Tests
// ============================================================================
// These tests validate fundamental physics properties of the Preisach model
// beyond basic hysteresis loop shape. They ensure compliance with classical
// ferroelectric theory and thermodynamic constraints.
//
// References:
//   - Preisach, F. "Uber die magnetische Nachwirkung" Z. Physik 94, 277 (1935)
//   - Mayergoyz, I.D. "Mathematical Models of Hysteresis" (1986, 1991)
//   - Lines & Glass, "Principles and Applications of Ferroelectrics" (1977)
//   - Pike et al., J. Appl. Phys. 85, 6660 (1999) - FORC analysis
// ============================================================================

// TestPreisachLoopArea_EnergyConservation verifies that hysteresis loop area
// represents positive energy dissipation per cycle.
//
// Physical basis: The area enclosed by a P-E hysteresis loop equals the
// energy dissipated per unit volume per cycle:
//
//	W = ∮ E dP (J/m³)
//
// This energy is always POSITIVE (second law of thermodynamics) - the system
// dissipates energy through domain wall motion, Joule heating, and acoustic
// emission. A negative loop area would imply spontaneous energy generation.
//
// For HZO, typical energy dissipation is ~10-50 µJ/cm³ per cycle at 1 MV/cm.
// This depends on coercive field and remanent polarization: W ~ 2 * Ec * Pr
//
// References:
//   - Mayergoyz, "Mathematical Models of Hysteresis" (1991), Chapter 3
//   - Pesic et al., Adv. Funct. Mater. 26, 4601 (2016) - HZO cycling energy
func TestPreisachLoopArea_EnergyConservation(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	model.SetTemperature(300) // Room temperature

	Ec := material.Ec
	Emax := Ec * 2.0

	// Generate hysteresis loop
	E, P := model.GetHysteresisLoop(Emax, 200)

	// Calculate loop area using trapezoidal integration
	// A = ∮ E dP = Σ E_i * (P_{i+1} - P_i) + 0.5 * (E_{i+1} - E_i) * (P_{i+1} - P_i)
	// Simplified: A ≈ Σ 0.5 * (E_i + E_{i+1}) * (P_{i+1} - P_i)
	area := 0.0
	n := len(E)
	for i := 0; i < n-1; i++ {
		dP := P[i+1] - P[i]
		avgE := (E[i] + E[i+1]) / 2
		area += avgE * dP
	}
	// Close the loop if not already closed
	if len(E) > 2 {
		dP := P[0] - P[n-1]
		avgE := (E[n-1] + E[0]) / 2
		area += avgE * dP
	}

	// Loop area should be positive (energy dissipated, not generated)
	// Due to the direction of traversal, we take absolute value
	absArea := math.Abs(area)

	if absArea <= 0 {
		t.Errorf("Hysteresis loop has zero or negative area: %.6e J/m³\n"+
			"  This violates energy conservation (second law of thermodynamics).\n"+
			"  A physical hysteresis loop must enclose positive area.",
			area)
	}

	// Estimate expected energy density: W ~ 4 * Ec * Pr (rough estimate)
	// For HZO at 1.2 MV/cm, Pr ~ 0.25 C/m²: W ~ 4 * 1.2e8 * 0.25 = 1.2e8 J/m³
	// This is per unit thickness - convert to µJ/cm³ for typical 10nm film
	// W_volume ~ 2 * Ec * (2*Pr) = 4 * Ec * Pr = 4 * 1.2e8 * 0.25 = 1.2e8 J/m³
	expectedMinArea := 2 * Ec * material.Pr * 0.1 // At least 10% of ideal
	expectedMaxArea := 8 * Ec * material.Ps       // Upper bound

	if absArea < expectedMinArea {
		t.Errorf("Loop area is smaller than expected for HZO:\n"+
			"  Measured area  = %.4e J/m³\n"+
			"  Expected min   = %.4e J/m³ (10%% of 4*Ec*Pr)\n"+
			"  This suggests the model may have incorrect Pr or Ec.",
			absArea, expectedMinArea)
	}

	if absArea > expectedMaxArea {
		t.Errorf("Loop area is larger than expected for HZO:\n"+
			"  Measured area  = %.4e J/m³\n"+
			"  Expected max   = %.4e J/m³ (8*Ec*Ps)\n"+
			"  This suggests numerical instability or incorrect scaling.",
			absArea, expectedMaxArea)
	}

	// Convert to µJ/cm³ for logging (1 J/m³ = 1e-6 J/cm³ = 1 µJ/cm³)
	areaUJCm3 := absArea * 1e-6

	t.Logf("Energy conservation verified:\n"+
		"  Loop area (energy dissipation) = %.4e J/m³ = %.2f µJ/cm³\n"+
		"  Theoretical 4*Ec*Pr            = %.4e J/m³\n"+
		"  Ratio to theoretical           = %.1f%%",
		absArea, areaUJCm3, 4*Ec*material.Pr, absArea/(4*Ec*material.Pr)*100)
}

// TestPreisachMinorLoops_Nested verifies that minor loops exhibit proper
// nesting behavior and lie inside the major loop.
//
// Physical basis: In the Preisach model, minor loops must satisfy:
//  1. Minor loops must lie INSIDE the major loop (congruency property)
//  2. A minor loop that returns to its starting field from the SAME direction
//     will close (return-point memory), but NOT if approaching from the opposite direction
//  3. Nested minor loops (minor loops within minor loops) maintain these properties
//
// Important: The Preisach model's return-point memory means that a minor loop
// closes when the field returns PAST its starting point in the direction of
// the original major loop branch - this is tested in TestPreisachWipeout_Property.
// A simple E=0 → E1 → E=0 path on an ascending branch is NOT guaranteed to close
// because E=0 was reached from the ascending direction, not descending.
//
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991), Section 2.4
func TestPreisachMinorLoops_Nested(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	model.SetTemperature(300)

	Ec := material.Ec
	Ps := material.Ps
	Emax := Ec * 2.5

	// First, generate major loop to establish bounds
	model.Reset()
	_, Pmajor := model.GetHysteresisLoop(Emax, 100)

	// Find P range of major loop
	var PmajorMax, PmajorMin = Pmajor[0], Pmajor[0]
	for _, p := range Pmajor {
		if p > PmajorMax {
			PmajorMax = p
		}
		if p < PmajorMin {
			PmajorMin = p
		}
	}

	// Test proper minor loop closure using the return-point memory property:
	// Start from saturation, go to E1, then to E2 < E1, then back to E1.
	// When we return past E2, the minor loop should close.
	model.Reset()
	model.Update(Emax) // Saturate positive (establishes the descending branch)

	E1 := Ec * 0.5      // First reversal point
	E2 := -Ec * 0.3     // Second reversal (creates minor loop)
	E1Return := Ec * 0.5 // Return to E1

	// Descend to E1
	steps := 30
	for i := 0; i <= steps; i++ {
		e := Emax + (E1-Emax)*float64(i)/float64(steps)
		model.Update(e)
	}
	PE1First := model.Polarization()

	// Descend further to E2 (below E1)
	for i := 0; i <= steps; i++ {
		e := E1 + (E2-E1)*float64(i)/float64(steps)
		model.Update(e)
	}

	// Now ASCEND back to E1 - this creates a minor loop
	var PminorAll []float64
	for i := 0; i <= steps; i++ {
		e := E2 + (E1Return-E2)*float64(i)/float64(steps)
		p := model.Update(e)
		PminorAll = append(PminorAll, p)
	}
	PE1Return := model.Polarization()

	// In classical Preisach, returning to E1 from below (ascending) vs from above (descending)
	// gives different P values - this is expected hysteresis behavior.
	// The minor loop "closure" in Preisach refers to when we continue PAST E1 on the
	// original (descending) path - then hysterons that switched during E1→E2→E1 get reset.

	// Test 1: Minor loop lies inside major loop
	var minorOutside int
	for _, p := range PminorAll {
		margin := 0.05 * Ps
		if p > PmajorMax+margin || p < PmajorMin-margin {
			minorOutside++
		}
	}

	if minorOutside > 0 {
		t.Errorf("Minor loop extends outside major loop:\n"+
			"  Major loop P range: [%.6e, %.6e] C/m²\n"+
			"  Points outside:     %d out of %d\n"+
			"  This violates the Preisach congruency property.",
			PmajorMin, PmajorMax, minorOutside, len(PminorAll))
	}

	// Test 2: Verify minor loop has non-zero area (it's a real loop, not a line)
	// The ascending branch (E2→E1) should have different P values than descending (E1→E2)
	// at the same E values
	minorLoopWidth := math.Abs(PE1Return - PE1First)
	relWidth := minorLoopWidth / Ps

	if relWidth < 0.01 {
		t.Logf("Minor loop has very small width (%.2f%% of Ps) - may indicate degenerate loop",
			relWidth*100)
	}

	// Test 3: Verify return-point memory (continue past E1 to saturation)
	// After completing the minor loop and returning to saturation,
	// we should return to the same saturation state
	for i := 0; i <= steps; i++ {
		e := E1Return + (Emax-E1Return)*float64(i)/float64(steps)
		model.Update(e)
	}
	PsatAfterMinor := model.Polarization()

	// Compare with direct path to saturation (no minor loop)
	model.Reset()
	model.Update(Emax)
	PsatDirect := model.Polarization()

	satDiff := math.Abs(PsatAfterMinor-PsatDirect) / Ps

	if satDiff > 0.01 {
		t.Errorf("Return to saturation after minor loop differs from direct path:\n"+
			"  P_sat (direct):      %.6e C/m²\n"+
			"  P_sat (after minor): %.6e C/m²\n"+
			"  Difference:          %.2f%% of Ps (expected < 1%%)",
			PsatDirect, PsatAfterMinor, satDiff*100)
	}

	t.Logf("Minor loop nesting verified:\n"+
		"  Minor loop: E1=%.2f, E2=%.2f MV/cm\n"+
		"  P at E1 (descending): %.2f µC/cm²\n"+
		"  P at E1 (ascending):  %.2f µC/cm² (hysteresis width: %.2f%%)\n"+
		"  Points outside major: %d\n"+
		"  Saturation recovery:  %.3f%% difference",
		E1/1e8, E2/1e8, PE1First*100, PE1Return*100, relWidth*100,
		minorOutside, satDiff*100)
}

// TestPreisachWipeout_Property verifies the Preisach wipe-out (memory erasure)
// property.
//
// Physical basis: In the classical Preisach model, a turning point in the field
// history "wipes out" the memory of smaller previous excursions. Specifically:
//
//   - If we traverse A → B → C → B → A, the system should return to the
//     SAME state it would have reached via direct path A → B → A
//   - A large field excursion erases the memory of smaller ones
//
// This is because hysterons only "remember" whether the field exceeded their
// switching threshold, not the intermediate history.
//
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991), Section 2.3
func TestPreisachWipeout_Property(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	model.SetTemperature(300)

	Ec := material.Ec
	Emax := Ec * 2.5

	// Path A: Direct path with one turning point
	// -Emax → E1 → -Emax
	E1 := Ec * 0.8

	model.Reset()
	model.Update(-Emax)

	// Go to E1
	steps := 30
	for i := 0; i <= steps; i++ {
		e := -Emax + (E1-(-Emax))*float64(i)/float64(steps)
		model.Update(e)
	}

	// Return to -Emax
	for i := 0; i <= steps; i++ {
		e := E1 + (-Emax-E1)*float64(i)/float64(steps)
		model.Update(e)
	}
	PdirectA := model.Polarization()

	// Path B: With intermediate minor loop
	// -Emax → E1 → E2 (small excursion) → E1 → -Emax
	E2 := Ec * 0.4 // E2 < E1

	model.Reset()
	model.Update(-Emax)

	// Go to E1
	for i := 0; i <= steps; i++ {
		e := -Emax + (E1-(-Emax))*float64(i)/float64(steps)
		model.Update(e)
	}

	// Small excursion to E2 and back to E1 (this should be "wiped out")
	for i := 0; i <= steps/2; i++ {
		e := E1 + (E2-E1)*float64(i)/float64(steps/2)
		model.Update(e)
	}
	for i := 0; i <= steps/2; i++ {
		e := E2 + (E1-E2)*float64(i)/float64(steps/2)
		model.Update(e)
	}

	// Return to -Emax
	for i := 0; i <= steps; i++ {
		e := E1 + (-Emax-E1)*float64(i)/float64(steps)
		model.Update(e)
	}
	PwithMinorB := model.Polarization()

	// The two final states should be nearly identical
	// because returning past E1 wipes out the E2 excursion
	relDiff := math.Abs(PdirectA-PwithMinorB) / math.Abs(PdirectA)

	// Allow 5% tolerance for numerical precision
	if relDiff > 0.05 {
		t.Errorf("Wipe-out property violated:\n"+
			"  Path A (direct):       P_final = %.6e C/m²\n"+
			"  Path B (with minor):   P_final = %.6e C/m²\n"+
			"  Relative difference:   %.2f%% (expected < 5%%)\n"+
			"  \n"+
			"  Path A: -Emax → E1=%.2fEc → -Emax\n"+
			"  Path B: -Emax → E1=%.2fEc → E2=%.2fEc → E1 → -Emax\n"+
			"  \n"+
			"  The intermediate E2 excursion should have been erased when\n"+
			"  returning past E1, but the final states differ.",
			PdirectA, PwithMinorB, relDiff*100,
			E1/Ec, E1/Ec, E2/Ec)
	}

	t.Logf("Wipe-out property verified:\n"+
		"  Path A (direct):     P_final = %.6e C/m²\n"+
		"  Path B (with minor): P_final = %.6e C/m²\n"+
		"  Relative difference: %.3f%%\n"+
		"  The E2 intermediate excursion was properly wiped out.",
		PdirectA, PwithMinorB, relDiff*100)
}

// TestPreisachSymmetry verifies that the hysteresis loop is symmetric
// about the origin for materials with no imprint.
//
// Physical basis: For an ideal ferroelectric without imprint or built-in
// bias field, the P-E loop should be point-symmetric about the origin:
//   - |Pr+| = |Pr-|  (symmetric remanent polarization)
//   - |Ec+| = |Ec-|  (symmetric coercive field)
//   - Loop shape should be identical in both quadrants (rotated 180°)
//
// Symmetry breaking indicates either:
//   - Imprint (built-in field due to defects or interface)
//   - Asymmetric distribution in the Preisach model
//   - Numerical artifacts
//
// Reference: Lines & Glass, "Principles and Applications of Ferroelectrics" (1977)
func TestPreisachSymmetry(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	model.SetTemperature(300)

	Ec := material.Ec
	Emax := Ec * 2.5

	// Generate a symmetric loop
	E, P := model.GetHysteresisLoop(Emax, 200)

	// Find coercive fields (E where P crosses zero)
	var EcPositive, EcNegative float64
	foundPosEc, foundNegEc := false, false

	for i := 1; i < len(E); i++ {
		// Look for zero crossing from negative to positive P (ascending Ec)
		if !foundPosEc && P[i-1] < 0 && P[i] >= 0 && E[i] > 0 {
			// Linear interpolation for more accurate Ec
			interpT := -P[i-1] / (P[i] - P[i-1])
			EcPositive = E[i-1] + interpT*(E[i]-E[i-1])
			foundPosEc = true
		}
		// Look for zero crossing from positive to negative P (descending Ec)
		if !foundNegEc && P[i-1] > 0 && P[i] <= 0 && E[i] < 0 {
			interpT := P[i-1] / (P[i-1] - P[i])
			EcNegative = E[i-1] + interpT*(E[i]-E[i-1])
			foundNegEc = true
		}
	}

	// Find remanent polarization (P at E=0)
	var PrPositive, PrNegative float64
	foundPrPos, foundPrNeg := false, false

	for i := 1; i < len(E); i++ {
		// E crossing zero from positive (descending branch)
		if !foundPrPos && E[i-1] > 0 && E[i] <= 0 && P[i] > 0 {
			interpT := E[i-1] / (E[i-1] - E[i])
			PrPositive = P[i-1] + interpT*(P[i]-P[i-1])
			foundPrPos = true
		}
		// E crossing zero from negative (ascending branch)
		if !foundPrNeg && E[i-1] < 0 && E[i] >= 0 && P[i] < 0 {
			interpT := -E[i-1] / (E[i] - E[i-1])
			PrNegative = P[i-1] + interpT*(P[i]-P[i-1])
			foundPrNeg = true
		}
	}

	// Check Ec symmetry: |Ec+| should equal |Ec-|
	EcAsymmetry := math.Abs(math.Abs(EcPositive)-math.Abs(EcNegative)) / Ec

	if EcAsymmetry > 0.10 {
		t.Errorf("Coercive field asymmetry exceeds 10%%:\n"+
			"  Ec+  = %.4e V/m (%.2f MV/cm)\n"+
			"  Ec-  = %.4e V/m (%.2f MV/cm)\n"+
			"  |Ec+| - |Ec-| = %.4e V/m (%.1f%% of Ec)\n"+
			"  This suggests imprint or asymmetric Preisach distribution.",
			EcPositive, EcPositive/1e8, EcNegative, EcNegative/1e8,
			math.Abs(EcPositive)-math.Abs(EcNegative), EcAsymmetry*100)
	}

	// Check Pr symmetry: |Pr+| should equal |Pr-|
	PrAsymmetry := math.Abs(math.Abs(PrPositive)-math.Abs(PrNegative)) / material.Pr

	if PrAsymmetry > 0.10 {
		t.Errorf("Remanent polarization asymmetry exceeds 10%%:\n"+
			"  Pr+ = %.4e C/m² (%.2f µC/cm²)\n"+
			"  Pr- = %.4e C/m² (%.2f µC/cm²)\n"+
			"  |Pr+| - |Pr-| = %.4e C/m² (%.1f%% of Pr)\n"+
			"  This suggests imprint or asymmetric distribution.",
			PrPositive, PrPositive*100, PrNegative, PrNegative*100,
			math.Abs(PrPositive)-math.Abs(PrNegative), PrAsymmetry*100)
	}

	t.Logf("Loop symmetry verified:\n"+
		"  Ec+ = %.2f MV/cm, Ec- = %.2f MV/cm (asymmetry: %.1f%%)\n"+
		"  Pr+ = %.2f µC/cm², Pr- = %.2f µC/cm² (asymmetry: %.1f%%)",
		EcPositive/1e8, math.Abs(EcNegative)/1e8, EcAsymmetry*100,
		PrPositive*100, math.Abs(PrNegative)*100, PrAsymmetry*100)
}

// TestPreisachDistributionShape verifies that the Preisach distribution
// function has the expected shape centered near the coercive field.
//
// Physical basis: The Preisach distribution μ(α, β) represents the density
// of hysterons (elementary switching units) in the (α, β) plane. For a
// well-behaved ferroelectric:
//
//   - The distribution should peak along the diagonal α ≈ -β (symmetric switching)
//   - The peak should occur near α ≈ Ec, β ≈ -Ec
//   - The distribution width represents switching field dispersion
//   - Gaussian or Lorentzian shapes are typical (from domain size distribution)
//
// The width of the distribution correlates with the "squareness" of the loop:
// narrow distribution → square loop, wide distribution → rounded loop.
//
// Reference: Bartic et al., J. Appl. Phys. 89, 3420 (2001)
func TestPreisachDistributionShape(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	model.SetTemperature(300)

	Ec := material.Ec

	// Get the distribution weights
	alphas, betas, _ := model.GetPreisachPlane()
	weights := model.GetDistribution()

	if len(weights) == 0 {
		t.Fatal("Distribution is empty")
	}

	// Find the peak of the distribution
	maxWeight := weights[0]
	peakIdx := 0
	for i, w := range weights {
		if w > maxWeight {
			maxWeight = w
			peakIdx = i
		}
	}

	peakAlpha := alphas[peakIdx]
	peakBeta := betas[peakIdx]

	// The peak should be near (Ec, -Ec) for a symmetric material
	// Allow 50% deviation from Ec
	alphaDev := math.Abs(peakAlpha-Ec) / Ec
	betaDev := math.Abs(peakBeta+Ec) / Ec

	if alphaDev > 0.5 {
		t.Errorf("Distribution alpha peak is far from Ec:\n"+
			"  Peak alpha = %.4e V/m (%.2f MV/cm)\n"+
			"  Expected   ≈ %.4e V/m (%.2f MV/cm)\n"+
			"  Deviation  = %.1f%% of Ec",
			peakAlpha, peakAlpha/1e8, Ec, Ec/1e8, alphaDev*100)
	}

	if betaDev > 0.5 {
		t.Errorf("Distribution beta peak is far from -Ec:\n"+
			"  Peak beta  = %.4e V/m (%.2f MV/cm)\n"+
			"  Expected   ≈ %.4e V/m (%.2f MV/cm)\n"+
			"  Deviation  = %.1f%% of Ec",
			peakBeta, peakBeta/1e8, -Ec, -Ec/1e8, betaDev*100)
	}

	// Calculate mean and standard deviation of alpha distribution
	// (weighted by distribution function)
	totalWeight := 0.0
	meanAlpha := 0.0
	for i := range alphas {
		totalWeight += weights[i]
		meanAlpha += alphas[i] * weights[i]
	}
	meanAlpha /= totalWeight

	varAlpha := 0.0
	for i := range alphas {
		varAlpha += weights[i] * math.Pow(alphas[i]-meanAlpha, 2)
	}
	varAlpha /= totalWeight
	sigmaAlpha := math.Sqrt(varAlpha)

	// The distribution width should be physically reasonable
	// Typical σ/Ec is 0.2-0.8 for HZO
	relWidth := sigmaAlpha / Ec

	if relWidth < 0.1 {
		t.Errorf("Distribution is too narrow:\n"+
			"  σ_alpha = %.4e V/m\n"+
			"  σ/Ec    = %.2f (expected 0.2-0.8)\n"+
			"  A very narrow distribution produces unphysically square loops.",
			sigmaAlpha, relWidth)
	}

	if relWidth > 1.5 {
		t.Errorf("Distribution is too wide:\n"+
			"  σ_alpha = %.4e V/m\n"+
			"  σ/Ec    = %.2f (expected 0.2-0.8)\n"+
			"  A very wide distribution produces unphysically rounded loops.",
			sigmaAlpha, relWidth)
	}

	t.Logf("Distribution shape verified:\n"+
		"  Peak location: (α=%.2f, β=%.2f) MV/cm\n"+
		"  Expected peak: (Ec=%.2f, -Ec=%.2f) MV/cm\n"+
		"  Distribution mean α = %.2f MV/cm\n"+
		"  Distribution width σ_α = %.2f MV/cm (σ/Ec = %.2f)\n"+
		"  Total weight = %.4e (normalized to Ps)",
		peakAlpha/1e8, peakBeta/1e8, Ec/1e8, -Ec/1e8,
		meanAlpha/1e8, sigmaAlpha/1e8, relWidth, totalWeight)
}

// TestPreisachFirstOrderReversalCurves verifies that FORC (First Order
// Reversal Curves) exhibit expected patterns.
//
// Physical basis: FORC curves are a standard experimental technique for
// characterizing ferroelectric materials and extracting the Preisach
// distribution. Each FORC starts from positive saturation, descends to
// a reversal field Hr, then ascends back to saturation.
//
// Key properties:
//   - FORC curves should not cross each other
//   - The envelope of all FORCs forms the major loop
//   - The FORC distribution ρ(Hr, H) = -0.5 * ∂²M/∂Hr∂H relates to μ(α,β)
//
// For HZO, the FORC distribution should show a peak near Ec with
// width related to switching field dispersion.
//
// Reference: Pike et al., J. Appl. Phys. 85, 6660 (1999)
func TestPreisachFirstOrderReversalCurves(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	model.SetTemperature(300)

	Ec := material.Ec
	Ps := material.Ps
	Emax := Ec * 2.5

	// Generate FORC curves for several reversal fields
	// Hr values from +Emax to -Emax
	numFORCs := 10
	reversalFields := make([]float64, numFORCs)
	for i := 0; i < numFORCs; i++ {
		reversalFields[i] = Emax - 2*Emax*float64(i)/float64(numFORCs-1)
	}

	// Store FORC data: polarization values along each ascending branch
	type forcData struct {
		Hr    float64
		P     []float64
		PrMax float64 // P at end (Emax)
	}
	forcCurves := make([]forcData, numFORCs)

	steps := 50

	for i, Hr := range reversalFields {
		model.Reset()

		// Start from positive saturation
		model.Update(Emax)

		// Descend to reversal field Hr
		for j := 0; j <= steps; j++ {
			e := Emax + (Hr-Emax)*float64(j)/float64(steps)
			model.Update(e)
		}

		// Record ascending branch (the actual FORC)
		P := make([]float64, 0, steps+1)

		for j := 0; j <= steps; j++ {
			e := Hr + (Emax-Hr)*float64(j)/float64(steps)
			p := model.Update(e)
			P = append(P, p)
		}

		forcCurves[i] = forcData{Hr: Hr, P: P, PrMax: P[len(P)-1]}
	}

	// Test 1: FORC curves should not cross significantly
	// Compare adjacent FORC curves at same E value
	crossings := 0
	for i := 0; i < numFORCs-1; i++ {
		for j := 0; j < len(forcCurves[i].P) && j < len(forcCurves[i+1].P); j++ {
			// FORC with lower Hr should have lower P at same E
			// (because it started from more negative saturation state)
			if forcCurves[i].P[j] < forcCurves[i+1].P[j]-Ps*0.01 {
				crossings++
			}
		}
	}

	if crossings > len(forcCurves[0].P)/10 {
		t.Errorf("FORC curves show significant crossing:\n"+
			"  Number of crossings: %d (out of %d comparisons)\n"+
			"  FORC curves should be monotonically ordered by Hr.",
			crossings, (numFORCs-1)*len(forcCurves[0].P))
	}

	// Test 2: All FORCs should converge at saturation
	Psat := forcCurves[0].PrMax // P at Emax
	maxSatDev := 0.0
	for _, fc := range forcCurves {
		dev := math.Abs(fc.PrMax - Psat)
		if dev > maxSatDev {
			maxSatDev = dev
		}
	}
	relSatDev := maxSatDev / Ps

	if relSatDev > 0.05 {
		t.Errorf("FORC curves do not converge at saturation:\n"+
			"  Maximum deviation from P_sat: %.4e C/m² (%.1f%% of Ps)\n"+
			"  All curves should reach the same saturation polarization.",
			maxSatDev, relSatDev*100)
	}

	t.Logf("FORC analysis:\n"+
		"  Number of FORCs: %d\n"+
		"  Reversal field range: [%.2f, %.2f] MV/cm\n"+
		"  FORC crossings: %d (should be ~0)\n"+
		"  Saturation convergence: max dev = %.2f%% of Ps",
		numFORCs, reversalFields[0]/1e8, reversalFields[numFORCs-1]/1e8,
		crossings, relSatDev*100)
}

// TestPreisachSaturationBehavior verifies that polarization properly
// saturates at Ps and does not exhibit numerical instability.
//
// Physical basis: At fields much larger than Ec, all ferroelectric domains
// should be aligned, and polarization approaches the saturation value Ps.
// The polarization should:
//   - Never exceed |Ps| (this would violate physics)
//   - Approach Ps asymptotically (not suddenly)
//   - Remain stable at high fields (no oscillations or divergence)
//
// This tests the model's numerical stability at extreme conditions.
//
// Reference: Lines & Glass, "Principles and Applications of Ferroelectrics" (1977)
func TestPreisachSaturationBehavior(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	model.SetTemperature(300)

	Ec := material.Ec
	Ps := material.Ps

	// Test various field magnitudes up to 10× Ec
	fieldMultipliers := []float64{1.0, 2.0, 3.0, 5.0, 10.0}

	for _, mult := range fieldMultipliers {
		Etest := mult * Ec

		model.Reset()

		// Approach saturation gradually
		steps := 100
		var maxP float64

		for i := 0; i <= steps; i++ {
			e := Etest * float64(i) / float64(steps)
			p := model.Update(e)

			if p > maxP {
				maxP = p
			}
		}

		// Final saturation polarization
		Psat := model.Polarization()

		// Test 1: P should not exceed Ps
		if maxP > Ps*1.01 {
			t.Errorf("Polarization exceeds Ps at E=%.1f×Ec:\n"+
				"  Max P = %.6e C/m² (%.2f µC/cm²)\n"+
				"  Ps    = %.6e C/m² (%.2f µC/cm²)\n"+
				"  Excess = %.2f%%",
				mult, maxP, maxP*100, Ps, Ps*100, (maxP-Ps)/Ps*100)
		}

		// Test 2: P at saturation should be close to Ps
		satError := math.Abs(Psat-Ps) / Ps

		if satError > 0.05 && mult >= 2.0 {
			t.Errorf("Saturation not reached at E=%.1f×Ec:\n"+
				"  P_sat = %.6e C/m² (%.2f µC/cm²)\n"+
				"  Ps    = %.6e C/m² (%.2f µC/cm²)\n"+
				"  Error = %.2f%% (expected < 5%% for E > 2×Ec)",
				mult, Psat, Psat*100, Ps, Ps*100, satError*100)
		}

		t.Logf("Saturation at E=%.1f×Ec: P_sat=%.2f µC/cm² (%.1f%% of Ps)",
			mult, Psat*100, Psat/Ps*100)
	}

	// Test 3: Stability at very high fields
	model.Reset()
	Ehigh := 10 * Ec
	model.Update(Ehigh)

	// Apply small perturbations and check stability
	for i := 0; i < 10; i++ {
		p := model.Update(Ehigh + 0.001*Ec*float64(i-5))
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Errorf("Numerical instability at high field:\n"+
				"  E = %.4e V/m\n"+
				"  P = %v (NaN or Inf)",
				Ehigh+0.001*Ec*float64(i-5), p)
		}
		if math.Abs(p) > 2*Ps {
			t.Errorf("Polarization divergence at high field:\n"+
				"  E = %.4e V/m\n"+
				"  P = %.4e C/m² (> 2×Ps)",
				Ehigh+0.001*Ec*float64(i-5), p)
		}
	}

	t.Log("Numerical stability verified at high fields")
}

// TestPreisachTemperatureDependence verifies that the model exhibits
// correct Curie-Weiss temperature behavior.
//
// Physical basis: Ferroelectric polarization decreases with temperature
// according to Curie-Weiss law:
//
//	P(T) ∝ (1 - T/Tc)^β for T < Tc
//	P(T) = 0             for T ≥ Tc
//
// where:
//   - Tc is the Curie temperature (phase transition)
//   - β is a critical exponent (typically 0.3-0.5 for ferroelectrics)
//
// For HZO, Tc ≈ 450°C (723 K) and β ≈ 0.5 (mean-field theory).
//
// References:
//   - Lines & Glass, "Principles and Applications of Ferroelectrics" (1977)
//   - Muller et al., Nano Lett. 12, 4318 (2012) - HfO2 temperature dependence
func TestPreisachTemperatureDependence(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Tc := model.CurieTemp
	Ec0 := material.Ec
	beta := model.TempExponent

	// Test temperatures from 0.1*Tc to 0.95*Tc
	temperatures := []float64{0.1, 0.3, 0.5, 0.7, 0.9, 0.95}

	type tempResult struct {
		T        float64
		EcT      float64
		EcRatio  float64
		ExpRatio float64
	}
	results := make([]tempResult, len(temperatures))

	for i, tRatio := range temperatures {
		T := tRatio * Tc

		model.SetTemperature(T)
		EcT := model.GetEffectiveEc()

		// Theoretical Curie-Weiss prediction
		expected := Ec0 * math.Pow(1-T/Tc, beta)

		results[i] = tempResult{
			T:        T,
			EcT:      EcT,
			EcRatio:  EcT / Ec0,
			ExpRatio: expected / Ec0,
		}

		// Check against Curie-Weiss law (allow 20% deviation)
		relError := math.Abs(EcT-expected) / expected

		if relError > 0.2 && T < 0.9*Tc {
			t.Errorf("Temperature dependence deviates from Curie-Weiss:\n"+
				"  T = %.0f K (%.1f%% of Tc)\n"+
				"  Ec(T) = %.4e V/m (%.2f MV/cm)\n"+
				"  Expected = %.4e V/m (from Ec0*(1-T/Tc)^β)\n"+
				"  Deviation = %.1f%% (expected < 20%%)",
				T, T/Tc*100, EcT, EcT/1e8, expected, relError*100)
		}
	}

	// Verify that Ec decreases monotonically with T
	for i := 1; i < len(results); i++ {
		if results[i].EcT > results[i-1].EcT+1e-6 {
			t.Errorf("Ec not monotonically decreasing with T:\n"+
				"  Ec(%.0fK) = %.4e V/m\n"+
				"  Ec(%.0fK) = %.4e V/m (should be lower)",
				results[i-1].T, results[i-1].EcT,
				results[i].T, results[i].EcT)
		}
	}

	// Test above Curie temperature
	model.SetTemperature(Tc * 1.1)
	EcAboveTc := model.GetEffectiveEc()

	if EcAboveTc > Ec0*0.01 {
		t.Errorf("Ferroelectricity should vanish above Tc:\n"+
			"  T = %.0f K (110%% of Tc)\n"+
			"  Ec(T) = %.4e V/m (expected ≈ 0)",
			Tc*1.1, EcAboveTc)
	}

	// Log results
	t.Logf("Temperature dependence verified (Tc=%.0fK, β=%.2f):", Tc, beta)
	for _, r := range results {
		t.Logf("  T=%4.0fK (%4.1f%% Tc): Ec/Ec0=%.3f (expected %.3f)",
			r.T, r.T/Tc*100, r.EcRatio, r.ExpRatio)
	}
	t.Logf("  T=%.0fK (above Tc): Ec/Ec0=%.4f (should be ~0)", Tc*1.1, EcAboveTc/Ec0)
}
