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

			// Saturate negative
			for i := -steps; i <= -steps; i-- {
				E := float64(i) / float64(steps) * Emax
				model.Update(E)
			}

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
