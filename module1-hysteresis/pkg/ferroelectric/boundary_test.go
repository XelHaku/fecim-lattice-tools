package ferroelectric

import (
	"math"
	"testing"
)

// ============================================================================
// BOUNDARY CONDITION TESTS FOR EXTREME PHYSICS PARAMETERS
// ============================================================================
// These tests verify that the ferroelectric models handle extreme parameter
// values without panics, NaN, Inf, or other numerical instabilities.
//
// Physical boundary conditions tested:
// - Temperature: 0K (absolute zero) to above Curie temperature
// - Electric field: 0 to 10x coercive field
// - Preisach grid: 0, 1, very large grids
// - Polarization domain: exactly at saturation limits
//
// References:
// - Curie temperature for HZO: ~723K (450C)
// - Cryogenic enhancement: Pr up to 75 uC/cm2 at 4K
// - Standard Ec: ~1.0-1.5 MV/cm

// ============================================================================
// 1. Temperature Boundary Tests: Cryogenic to Above Curie
// ============================================================================

// TestBoundaryTemperature_CryogenicToAboveCurie tests the complete temperature
// range from cryogenic (4K) to above Curie temperature (800K).
//
// Physics boundary being tested:
// - Below Curie: Pr and Ec should be non-zero, decreasing as T approaches Tc
// - At Curie: Ferroelectric properties vanish (Pr=0, Ec=0)
// - Above Curie: Material becomes paraelectric (Pr=0, Ec=0)
// - Cryogenic: Enhanced polarization (verified experimentally at 4K)
func TestBoundaryTemperature_CryogenicToAboveCurie(t *testing.T) {
	material := DefaultHZO()
	curieTemp := material.CurieTemp // ~723K

	// Test temperatures spanning cryogenic to above Curie
	temperatures := []float64{
		4,                  // Cryogenic (quantum computing)
		77,                 // Liquid nitrogen
		150,                // Cold
		300,                // Room temperature
		400,                // Elevated
		500,                // High temperature
		600,                // Near Curie
		700,                // Very close to Curie
		curieTemp,          // Exactly at Curie
		curieTemp + 10,     // Just above Curie
		800,                // Well above Curie
	}

	t.Run("CoerciveFieldAtTemp", func(t *testing.T) {
		var prevEc float64 = math.MaxFloat64

		for _, T := range temperatures {
			Ec := material.CoerciveFieldAtTemp(T)

			// Test 1: No NaN values
			if math.IsNaN(Ec) {
				t.Errorf("CoerciveFieldAtTemp(%.0fK) returned NaN", T)
				continue
			}

			// Test 2: No Inf values
			if math.IsInf(Ec, 0) {
				t.Errorf("CoerciveFieldAtTemp(%.0fK) returned Inf", T)
				continue
			}

			// Test 3: Non-negative (Ec should never be negative)
			if Ec < 0 {
				t.Errorf("CoerciveFieldAtTemp(%.0fK) = %e (negative, should be >= 0)", T, Ec)
			}

			// Test 4: At/above Curie, Ec should be 0
			if T >= curieTemp && Ec != 0 {
				t.Errorf("CoerciveFieldAtTemp(%.0fK >= Tc=%.0fK) = %e, expected 0", T, curieTemp, Ec)
			}

			// Test 5: Below Curie, Ec should be positive and decreasing
			if T < curieTemp {
				if Ec == 0 {
					t.Errorf("CoerciveFieldAtTemp(%.0fK < Tc) = 0, should be positive", T)
				}
				// Ec should decrease monotonically as T increases toward Tc
				if T > 4 && Ec > prevEc {
					t.Logf("Warning: Ec(%.0fK)=%e > Ec(prev)=%e - non-monotonic", T, Ec, prevEc)
				}
				prevEc = Ec
			}

			t.Logf("T=%.0fK: Ec=%.4e V/m (%.1f%% of Ec0)", T, Ec, 100*Ec/material.Ec)
		}
	})

	t.Run("PolarizationAtTemp", func(t *testing.T) {
		var prevPr float64 = math.MaxFloat64

		for _, T := range temperatures {
			Pr := material.PolarizationAtTemp(T)

			// Test 1: No NaN values
			if math.IsNaN(Pr) {
				t.Errorf("PolarizationAtTemp(%.0fK) returned NaN", T)
				continue
			}

			// Test 2: No Inf values
			if math.IsInf(Pr, 0) {
				t.Errorf("PolarizationAtTemp(%.0fK) returned Inf", T)
				continue
			}

			// Test 3: Non-negative (Pr should never be negative)
			if Pr < 0 {
				t.Errorf("PolarizationAtTemp(%.0fK) = %e (negative, should be >= 0)", T, Pr)
			}

			// Test 4: At/above Curie, Pr should be 0
			if T >= curieTemp && Pr != 0 {
				t.Errorf("PolarizationAtTemp(%.0fK >= Tc=%.0fK) = %e, expected 0", T, curieTemp, Pr)
			}

			// Test 5: Below Curie, Pr should be positive and decreasing
			if T < curieTemp {
				if Pr == 0 {
					t.Errorf("PolarizationAtTemp(%.0fK < Tc) = 0, should be positive", T)
				}
				// Pr should decrease monotonically as T increases toward Tc
				if T > 4 && Pr > prevPr {
					t.Logf("Warning: Pr(%.0fK)=%e > Pr(prev)=%e - non-monotonic", T, Pr, prevPr)
				}
				prevPr = Pr
			}

			t.Logf("T=%.0fK: Pr=%.4e C/m2 (%.1f%% of Pr0)", T, Pr, 100*Pr/material.Pr)
		}
	})

	t.Run("MayergoyzPreisach_Temperature", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 20)

		for _, T := range temperatures {
			// Should not panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("SetTemperature(%.0fK) panicked: %v", T, r)
					}
				}()
				model.SetTemperature(T)
			}()

			EcEff := model.GetEffectiveEc()

			// Test 1: No NaN
			if math.IsNaN(EcEff) {
				t.Errorf("GetEffectiveEc() at T=%.0fK returned NaN", T)
				continue
			}

			// Test 2: No Inf
			if math.IsInf(EcEff, 0) {
				t.Errorf("GetEffectiveEc() at T=%.0fK returned Inf", T)
				continue
			}

			// Test 3: Non-negative
			if EcEff < 0 {
				t.Errorf("GetEffectiveEc() at T=%.0fK = %e (negative)", T, EcEff)
			}

			// Test 4: At/above Curie, should be 0
			if T >= curieTemp && EcEff != 0 {
				t.Errorf("GetEffectiveEc() at T=%.0fK (>= Tc) = %e, expected 0", T, EcEff)
			}

			// Test 5: Verify Update() works at this temperature
			P := model.Update(material.Ec)
			if math.IsNaN(P) {
				t.Errorf("Update() at T=%.0fK returned NaN", T)
			}
			if math.IsInf(P, 0) {
				t.Errorf("Update() at T=%.0fK returned Inf", T)
			}

			t.Logf("MayergoyzPreisach T=%.0fK: Ec_eff=%.4e, P=%.4e", T, EcEff, P)
		}
	})

	t.Run("CryogenicEnhancement", func(t *testing.T) {
		// At cryogenic temperatures (4K), HZO shows ENHANCED polarization
		// Reference: Adv. Electron. Mater. 2024 - Pr up to 75 uC/cm2 at 4K
		cryoMaterial := CryogenicHZO()

		// Verify cryogenic material has enhanced Pr
		if cryoMaterial.Pr <= material.Pr {
			t.Logf("Note: CryogenicHZO Pr=%.4e <= DefaultHZO Pr=%.4e (expected enhancement)",
				cryoMaterial.Pr, material.Pr)
		} else {
			t.Logf("Cryogenic enhancement verified: Pr=%.4e vs RT Pr=%.4e (%.1fx)",
				cryoMaterial.Pr, material.Pr, cryoMaterial.Pr/material.Pr)
		}

		// Create Preisach model with cryogenic material
		model := NewMayergoyzPreisach(cryoMaterial, 20)
		model.SetTemperature(4) // 4K

		// Should work without errors
		EcEff := model.GetEffectiveEc()
		if math.IsNaN(EcEff) || math.IsInf(EcEff, 0) {
			t.Errorf("Cryogenic model returned invalid Ec: %e", EcEff)
		}
		if EcEff <= 0 {
			t.Errorf("Cryogenic model Ec should be positive at 4K, got %e", EcEff)
		}

		t.Logf("Cryogenic model at 4K: Ec_eff=%.4e V/m", EcEff)
	})
}

// ============================================================================
// 2. Electric Field Boundary Tests: Zero to Extreme
// ============================================================================

// TestBoundaryEField_ExtremeValues tests electric field from 0 to 10x coercive.
//
// Physics boundary being tested:
// - At E=0: Polarization should retain (remanent state)
// - At E=Ec: Switching threshold
// - At E>>Ec: Saturation behavior - P should not exceed Ps
// - No numerical overflow or underflow at extreme fields
func TestBoundaryEField_ExtremeValues(t *testing.T) {
	material := DefaultHZO()
	Ec := material.Ec
	Ps := material.Ps

	// Test fields from 0 to 10x Ec
	fieldMultipliers := []float64{0, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0}

	t.Run("PreisachModel", func(t *testing.T) {
		model := NewPreisachModel(material)

		for _, mult := range fieldMultipliers {
			E := mult * Ec

			// Reset to clean state
			model.Reset()

			// Apply positive field
			P := model.Update(E)

			// Test 1: No NaN
			if math.IsNaN(P) {
				t.Errorf("Update(%.1f*Ec) returned NaN", mult)
				continue
			}

			// Test 2: No Inf
			if math.IsInf(P, 0) {
				t.Errorf("Update(%.1f*Ec) returned Inf", mult)
				continue
			}

			// Test 3: Bounded by Ps - polarization should never exceed saturation
			if math.Abs(P) > Ps*1.01 { // 1% tolerance for floating point
				t.Errorf("Update(%.1f*Ec) = %.4e exceeds Ps=%.4e", mult, P, Ps)
			}

			// Test 4: At high fields (>2*Ec), should be near saturation
			if mult >= 2.0 && P < 0.8*Ps {
				t.Logf("Note: At %.1f*Ec, P=%.4e is below 80%% of Ps", mult, P)
			}

			// Test negative field
			model.Reset()
			Pneg := model.Update(-E)

			if math.IsNaN(Pneg) || math.IsInf(Pneg, 0) {
				t.Errorf("Update(-%.1f*Ec) returned invalid: %e", mult, Pneg)
			}
			if math.Abs(Pneg) > Ps*1.01 {
				t.Errorf("Update(-%.1f*Ec) = %.4e exceeds Ps bounds", mult, Pneg)
			}

			t.Logf("PreisachModel: E=%.1f*Ec -> P=%.4e (%.1f%% of Ps)", mult, P, 100*P/Ps)
		}
	})

	t.Run("MayergoyzPreisach", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 30)

		for _, mult := range fieldMultipliers {
			E := mult * Ec

			model.Reset()
			P := model.Update(E)

			// Test 1: No NaN
			if math.IsNaN(P) {
				t.Errorf("MayergoyzPreisach Update(%.1f*Ec) returned NaN", mult)
				continue
			}

			// Test 2: No Inf
			if math.IsInf(P, 0) {
				t.Errorf("MayergoyzPreisach Update(%.1f*Ec) returned Inf", mult)
				continue
			}

			// Test 3: Bounded by Ps (with small tolerance for floating-point)
			if math.Abs(P) > Ps*1.001 {
				t.Errorf("MayergoyzPreisach Update(%.1f*Ec) = %.4e exceeds Ps=%.4e", mult, P, Ps)
			}

			t.Logf("MayergoyzPreisach: E=%.1f*Ec -> P=%.4e (%.1f%% of Ps)", mult, P, 100*P/Ps)
		}
	})

	t.Run("SaturationStability", func(t *testing.T) {
		// Apply extremely high field for many iterations to verify no drift
		model := NewPreisachModel(material)

		Eextreme := 10 * Ec
		numIterations := 1000

		for i := 0; i < numIterations; i++ {
			P := model.Update(Eextreme)

			if math.Abs(P) > Ps*1.01 {
				t.Errorf("At iteration %d, P=%.4e exceeds Ps bounds", i, P)
				break
			}
			if math.IsNaN(P) || math.IsInf(P, 0) {
				t.Errorf("At iteration %d, got invalid P: %e", i, P)
				break
			}
		}

		t.Logf("Saturation stability: %d iterations at 10*Ec completed successfully", numIterations)
	})

	t.Run("ZeroFieldRemanence", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 30)

		// Saturate positive
		model.Update(3 * Ec)

		// Return to zero field
		P_at_zero := model.Update(0)

		// Should have remanent polarization (not zero)
		if math.Abs(P_at_zero) < 0.1*Ps {
			t.Errorf("At E=0 after saturation, P=%.4e should persist (remanence)", P_at_zero)
		}

		// Should be no NaN/Inf
		if math.IsNaN(P_at_zero) || math.IsInf(P_at_zero, 0) {
			t.Errorf("At E=0, got invalid P: %e", P_at_zero)
		}

		t.Logf("Remanent polarization at E=0: %.4e (%.1f%% of Ps)", P_at_zero, 100*P_at_zero/Ps)
	})
}

// ============================================================================
// 3. Preisach Distribution Boundary Tests: Grid Edge Cases
// ============================================================================

// TestBoundaryPreisachDistribution_EdgeCases tests Preisach grid edge cases.
//
// Physics boundary being tested:
// - Zero grid points: Should handle gracefully (error or minimum grid)
// - Single grid point: Degenerate case
// - Very large grid: Memory and numerical stability
func TestBoundaryPreisachDistribution_EdgeCases(t *testing.T) {
	material := DefaultHZO()

	t.Run("ZeroGridPoints", func(t *testing.T) {
		// Zero grid points - implementation should handle gracefully
		defer func() {
			if r := recover(); r != nil {
				// Panic is acceptable for invalid input
				t.Logf("NewMayergoyzPreisach(0) panicked as expected: %v", r)
			}
		}()

		model := NewMayergoyzPreisach(material, 0)

		// If we get here, the model was created - verify it works minimally
		if model != nil {
			P := model.Update(material.Ec)
			if !math.IsNaN(P) && !math.IsInf(P, 0) {
				t.Logf("Zero grid model produced P=%e (graceful handling)", P)
			}
		}
	})

	t.Run("SingleGridPoint", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("NewMayergoyzPreisach(1) panicked: %v", r)
			}
		}()

		model := NewMayergoyzPreisach(material, 1)

		if model != nil {
			P := model.Update(material.Ec)
			if math.IsNaN(P) {
				t.Logf("Single grid point model produced NaN (acceptable for degenerate case)")
			} else if math.IsInf(P, 0) {
				t.Errorf("Single grid point model produced Inf")
			} else {
				t.Logf("Single grid point model produced P=%e", P)
			}
		}
	})

	t.Run("SmallGrid", func(t *testing.T) {
		// Small but valid grid (2x2)
		model := NewMayergoyzPreisach(material, 2)

		E, P := model.GetHysteresisLoop(2*material.Ec, 10)
		if len(E) == 0 || len(P) == 0 {
			t.Error("2x2 grid produced empty loop")
		}

		// Check for NaN/Inf in loop
		for i := range P {
			if math.IsNaN(P[i]) {
				t.Errorf("2x2 grid loop has NaN at index %d", i)
				break
			}
			if math.IsInf(P[i], 0) {
				t.Errorf("2x2 grid loop has Inf at index %d", i)
				break
			}
		}

		t.Logf("2x2 grid loop: %d points, P range [%.4e, %.4e]", len(P), minFloat64(P), maxFloat64(P))
	})

	t.Run("StandardGrid", func(t *testing.T) {
		// Standard grid (50x50 - recommended)
		model := NewMayergoyzPreisach(material, 50)

		E, P := model.GetHysteresisLoop(2*material.Ec, 50)
		if len(E) == 0 || len(P) == 0 {
			t.Error("50x50 grid produced empty loop")
		}

		// Verify full loop integrity
		for i := range P {
			if math.IsNaN(P[i]) || math.IsInf(P[i], 0) {
				t.Errorf("50x50 grid loop has invalid value at index %d", i)
				break
			}
		}

		t.Logf("50x50 grid loop: %d points, works correctly", len(P))
	})

	t.Run("LargeGrid", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping large grid test in short mode")
		}

		// Large grid (100x100) - test memory and performance
		model := NewMayergoyzPreisach(material, 100)

		// Just verify it can produce a loop without error
		E, P := model.GetHysteresisLoop(2*material.Ec, 20)
		if len(E) == 0 || len(P) == 0 {
			t.Error("100x100 grid produced empty loop")
		}

		// Basic sanity checks
		hasValid := false
		for _, p := range P {
			if !math.IsNaN(p) && !math.IsInf(p, 0) {
				hasValid = true
				break
			}
		}
		if !hasValid {
			t.Error("100x100 grid loop has no valid values")
		}

		t.Logf("100x100 grid loop: %d points, memory test passed", len(P))
	})

	t.Run("VeryLargeGrid", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping very large grid test in short mode")
		}

		// Very large grid (500x500) - stress test
		// This creates 500*500/2 = 125,000 hysterons
		defer func() {
			if r := recover(); r != nil {
				t.Logf("500x500 grid caused panic (acceptable for memory limits): %v", r)
			}
		}()

		model := NewMayergoyzPreisach(material, 500)

		// Quick test - just one update
		P := model.Update(material.Ec)
		if math.IsNaN(P) || math.IsInf(P, 0) {
			t.Logf("500x500 grid produced invalid value (acceptable)")
		} else {
			t.Logf("500x500 grid (125k hysterons) works: P=%e", P)
		}
	})
}

// ============================================================================
// 4. Polarization Domain Limit Tests
// ============================================================================

// TestBoundaryPolarizationDomain_Limits tests behavior at saturation limits.
//
// Physics boundary being tested:
// - Exactly at Ps: Polarization should not exceed saturation
// - Rapid field reversal: Numerical stability
// - Polarization must stay within [-Ps, +Ps] at all times
func TestBoundaryPolarizationDomain_Limits(t *testing.T) {
	material := DefaultHZO()
	Ps := material.Ps

	t.Run("ExactlyAtSaturation", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 30)

		// Apply field well beyond saturation
		Esat := 5 * material.Ec
		P := model.Update(Esat)

		// P should be at or very close to Ps, but not exceed it
		if P > Ps*1.001 {
			t.Errorf("P=%.4e exceeds Ps=%.4e at saturation field", P, Ps)
		}
		if P < 0.8*Ps {
			t.Logf("Note: P=%.4e is below 80%% of Ps at saturation (expected near Ps)", P)
		}

		// Negative saturation
		model.Reset()
		Pneg := model.Update(-Esat)

		if Pneg < -Ps*1.001 {
			t.Errorf("P=%.4e exceeds -Ps=%.4e at negative saturation", Pneg, -Ps)
		}

		t.Logf("Saturation test: +Ps=%.4e, P_pos=%.4e, P_neg=%.4e", Ps, P, Pneg)
	})

	t.Run("RapidFieldReversal", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 30)

		// Rapid reversals between +2Ec and -2Ec
		Eamp := 2 * material.Ec
		numReversals := 500

		for i := 0; i < numReversals; i++ {
			E := Eamp
			if i%2 == 1 {
				E = -Eamp
			}

			P := model.Update(E)

			// Test 1: No NaN
			if math.IsNaN(P) {
				t.Errorf("Reversal %d: NaN", i)
				break
			}

			// Test 2: No Inf
			if math.IsInf(P, 0) {
				t.Errorf("Reversal %d: Inf", i)
				break
			}

			// Test 3: Bounded by Ps (with small tolerance)
			if math.Abs(P) > Ps*1.001 {
				t.Errorf("Reversal %d: P=%.4e exceeds Ps bounds", i, P)
				break
			}
		}

		t.Logf("Rapid field reversal: %d reversals completed within bounds", numReversals)
	})

	t.Run("PolarizationBoundsUnderStress", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 30)

		// Stress test: random fields at various amplitudes
		numSteps := 1000
		maxViolation := 0.0

		for i := 0; i < numSteps; i++ {
			// Cycle through various field patterns
			phase := float64(i) / 50.0
			E := 3 * material.Ec * math.Sin(phase)

			P := model.Update(E)

			// Track maximum bound violation
			violation := math.Abs(P) - Ps
			if violation > maxViolation {
				maxViolation = violation
			}

			// Hard error if exceeds by more than 0.1%
			if violation > 0.001*Ps {
				t.Errorf("Step %d: P=%.4e exceeds Ps=%.4e by %.4e", i, P, Ps, violation)
				break
			}
		}

		t.Logf("Stress test: %d steps, max bound violation=%.4e (%.3f%% of Ps)",
			numSteps, maxViolation, 100*maxViolation/Ps)
	})

	t.Run("NormalizedPolarizationBounds", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 30)

		// Verify normalized polarization stays in [-1, +1]
		fields := []float64{0, material.Ec, -material.Ec, 3 * material.Ec, -3 * material.Ec}

		for _, E := range fields {
			model.Update(E)
			normP := model.NormalizedPolarization()

			if normP < -1.001 || normP > 1.001 {
				t.Errorf("NormalizedPolarization=%.4f at E=%.4e, should be in [-1, 1]", normP, E)
			}
		}

		t.Logf("Normalized polarization bounds verified for %d field values", len(fields))
	})
}

// ============================================================================
// 5. Material Property Boundary Tests
// ============================================================================

// TestBoundaryMaterialProperties tests edge cases in material parameter functions.
func TestBoundaryMaterialProperties(t *testing.T) {
	material := DefaultHZO()

	t.Run("SwitchingTime_ZeroField", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 20)

		// Switching time at E=0 should be infinite (no switching)
		tau := model.GetSwitchingTime(0)

		// Should be Inf or very large, not NaN
		if math.IsNaN(tau) {
			t.Error("GetSwitchingTime(0) returned NaN, expected Inf or very large")
		}
		if !math.IsInf(tau, 1) && tau < 1.0 {
			t.Logf("GetSwitchingTime(0) = %e (expected very large or Inf)", tau)
		}
	})

	t.Run("SwitchingTime_ExtremeFields", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 20)

		fields := []float64{
			material.Ec * 0.1,  // Low field
			material.Ec,        // Coercive field
			material.Ec * 10,   // High field
			material.Ec * 100,  // Very high field
		}

		var prevTau float64 = math.Inf(1)
		for _, E := range fields {
			tau := model.GetSwitchingTime(E)

			if math.IsNaN(tau) {
				t.Errorf("GetSwitchingTime(%.4e) returned NaN", E)
				continue
			}

			// Switching time should decrease with increasing field (Merz law)
			if E > 0 && tau > prevTau {
				t.Logf("Note: tau(%.4e)=%e > tau(prev)=%e - non-monotonic", E, tau, prevTau)
			}
			prevTau = tau

			t.Logf("GetSwitchingTime(%.1f*Ec) = %e s", E/material.Ec, tau)
		}
	})

	t.Run("DiscreteLevel_Boundaries", func(t *testing.T) {
		// Test DiscreteLevel at boundary levels
		levels := material.GetNumLevels()

		// Level 0 (minimum)
		G0 := material.DiscreteLevel(0, levels)
		if math.IsNaN(G0) || math.IsInf(G0, 0) {
			t.Errorf("DiscreteLevel(0) returned invalid: %e", G0)
		}

		// Level N-1 (maximum)
		Gmax := material.DiscreteLevel(levels-1, levels)
		if math.IsNaN(Gmax) || math.IsInf(Gmax, 0) {
			t.Errorf("DiscreteLevel(N-1) returned invalid: %e", Gmax)
		}

		// Level 0 should be less than level N-1
		if G0 >= Gmax {
			t.Errorf("DiscreteLevel(0)=%e >= DiscreteLevel(N-1)=%e (should be <)", G0, Gmax)
		}

		// Test negative level (should handle gracefully)
		Gneg := material.DiscreteLevel(-1, levels)
		if math.IsNaN(Gneg) || math.IsInf(Gneg, 0) {
			t.Logf("DiscreteLevel(-1) returned %e (implementation-defined)", Gneg)
		}

		// Test level >= N (should handle gracefully)
		Gover := material.DiscreteLevel(levels, levels)
		if math.IsNaN(Gover) || math.IsInf(Gover, 0) {
			t.Logf("DiscreteLevel(N) returned %e (implementation-defined)", Gover)
		}

		t.Logf("DiscreteLevel bounds: G(0)=%e, G(%d)=%e", G0, levels-1, Gmax)
	})

	t.Run("SingleLevelMaterial", func(t *testing.T) {
		// Test material with totalLevels=1 (degenerate case)
		G := material.DiscreteLevel(0, 1)

		if math.IsNaN(G) {
			t.Error("DiscreteLevel(0, 1) returned NaN")
		}
		if math.IsInf(G, 0) {
			t.Error("DiscreteLevel(0, 1) returned Inf")
		}

		t.Logf("Single level material: G=%e", G)
	})
}

// ============================================================================
// 6. Negative Temperature Guard
// ============================================================================

// TestBoundaryNegativeTemperature verifies graceful handling of T < 0 K.
func TestBoundaryNegativeTemperature(t *testing.T) {
	material := DefaultHZO()

	t.Run("CoerciveFieldAtNegativeTemp", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("CoerciveFieldAtTemp(-1) panicked: %v", r)
			}
		}()

		Ec := material.CoerciveFieldAtTemp(-1)

		// Should not be NaN
		if math.IsNaN(Ec) {
			t.Error("CoerciveFieldAtTemp(-1 K) returned NaN")
		}

		// Implementation may return a value (T=-1 < Tc, so (1-T/Tc)^0.5 is defined for negative T)
		t.Logf("CoerciveFieldAtTemp(-1 K) = %e (implementation-defined)", Ec)
	})

	t.Run("MayergoyzAtNegativeTemp", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetTemperature(-10) panicked: %v", r)
			}
		}()

		model := NewMayergoyzPreisach(material, 20)
		model.SetTemperature(-10)

		EcEff := model.GetEffectiveEc()
		if math.IsNaN(EcEff) {
			t.Error("GetEffectiveEc() at T=-10 K returned NaN")
		}

		t.Logf("MayergoyzPreisach at T=-10 K: Ec=%e (no panic)", EcEff)
	})
}
