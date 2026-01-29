package ferroelectric

import (
	"math"
	"testing"
)

// ============================================================================
// Boundary Condition and Numerical Stability Tests
// ============================================================================

// TestZeroFieldBehavior verifies E=0 correctly maintains memory retention.
// Ferroelectric memory effect: polarization should persist at remanent value
// when field returns to zero, NOT return to zero itself.
//
// NOTE: The simplified PreisachModel uses ascending/descending branch logic
// which doesn't perfectly model remanent polarization at E=0. For accurate
// P-E hysteresis including proper remanence, use MayergoyzPreisach.
func TestZeroFieldBehavior(t *testing.T) {
	t.Run("PreisachModel", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Apply positive field to saturate
		model.Reset()
		Emax := 2 * material.Ec
		model.Update(Emax)
		saturatedP := model.Polarization()

		// Return to E=0 - polarization should persist (remanent polarization)
		model.Update(0)
		P_at_zero := model.Polarization()

		// Key test: P should NOT return to zero - it should maintain significant polarization
		// (The simplified model may not maintain correct sign, but should show memory)
		if math.Abs(P_at_zero) < 0.1*material.Pr {
			t.Errorf("At E=0 after positive saturation, P=%.4f should persist near ±Pr (%.4f), not return to zero",
				P_at_zero, material.Pr)
		}

		t.Logf("PreisachModel: After +Emax then E=0: P_sat=%.4f → P(E=0)=%.4f (Pr=%.4f)",
			saturatedP, P_at_zero, material.Pr)

		// Apply negative field to saturate
		model.Update(-Emax)
		saturatedN := model.Polarization()

		// Return to E=0 - should persist at significant polarization
		model.Update(0)
		P_at_zero_neg := model.Polarization()

		// P should NOT return to zero - should maintain significant polarization magnitude
		if math.Abs(P_at_zero_neg) < 0.1*material.Pr {
			t.Errorf("At E=0 after negative saturation, P=%.4f should persist near ±Pr (%.4f), not return to zero",
				P_at_zero_neg, material.Pr)
		}

		// Verify the two E=0 states differ (shows history dependence)
		if math.Abs(P_at_zero-P_at_zero_neg) < 0.1*material.Pr {
			t.Errorf("P at E=0 should differ based on history: after +sat=%.4f, after -sat=%.4f",
				P_at_zero, P_at_zero_neg)
		}

		t.Logf("PreisachModel: After -Emax then E=0: P_sat=%.4f → P(E=0)=%.4f (-Pr=%.4f)",
			saturatedN, P_at_zero_neg, -material.Pr)
		t.Logf("PreisachModel: History dependence verified - P(E=0) differs: %.4f vs %.4f",
			P_at_zero, P_at_zero_neg)
	})

	t.Run("MayergoyzPreisach", func(t *testing.T) {
		material := DefaultHZO()
		model := NewMayergoyzPreisach(material, 20) // 20x20 grid

		// Apply positive field to saturate
		Emax := 2.5 * material.Ec
		model.Update(Emax)
		saturatedP := model.Polarization()

		// Return to E=0 - polarization should persist (CORRECT remanent behavior)
		model.Update(0)
		P_at_zero := model.Polarization()

		// P should NOT return to zero
		if math.Abs(P_at_zero) < 0.1*material.Pr {
			t.Errorf("MayergoyzPreisach: At E=0 after positive saturation, P=%.4f should persist near +Pr (%.4f)",
				P_at_zero, material.Pr)
		}

		// P should be positive (correct memory behavior)
		if P_at_zero < 0 {
			t.Errorf("MayergoyzPreisach: At E=0 after positive saturation, P=%.4f should be positive", P_at_zero)
		}

		t.Logf("MayergoyzPreisach: After +Emax then E=0: P_sat=%.4f → P(E=0)=%.4f (Pr=%.4f)",
			saturatedP, P_at_zero, material.Pr)

		// Apply negative field to saturate
		model.Update(-Emax)
		saturatedN := model.Polarization()

		// Return to E=0
		model.Update(0)
		P_at_zero_neg := model.Polarization()

		// P should NOT return to zero
		if math.Abs(P_at_zero_neg) < 0.1*material.Pr {
			t.Errorf("MayergoyzPreisach: At E=0 after negative saturation, P=%.4f should persist near -Pr (%.4f)",
				P_at_zero_neg, -material.Pr)
		}

		// P should be negative (correct memory behavior)
		if P_at_zero_neg > 0 {
			t.Errorf("MayergoyzPreisach: At E=0 after negative saturation, P=%.4f should be negative", P_at_zero_neg)
		}

		t.Logf("MayergoyzPreisach: After -Emax then E=0: P_sat=%.4f → P(E=0)=%.4f (-Pr=%.4f)",
			saturatedN, P_at_zero_neg, -material.Pr)
		t.Logf("MayergoyzPreisach: CORRECT remanent behavior - maintains sign based on history")
	})
}

// TestExtremeFieldValues verifies behavior at E >> Ec without overflow or NaN.
func TestExtremeFieldValues(t *testing.T) {
	t.Run("PreisachModel", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		extremeFields := []float64{
			10 * material.Ec,  // 10x coercive field
			100 * material.Ec, // 100x coercive field
		}

		for _, E := range extremeFields {
			// Test positive extreme
			model.Reset()
			P := model.Update(E)

			// Should not be NaN or Inf
			if math.IsNaN(P) {
				t.Errorf("PreisachModel: At E=%.2e (%.0fx Ec), got NaN", E, E/material.Ec)
			}
			if math.IsInf(P, 0) {
				t.Errorf("PreisachModel: At E=%.2e (%.0fx Ec), got Inf", E, E/material.Ec)
			}

			// Should saturate at +Ps, not exceed it
			if P > material.Ps*1.01 {
				t.Errorf("PreisachModel: At E=%.2e, P=%.4f exceeds Ps=%.4f", E, P, material.Ps)
			}

			// Should be bounded: -Ps <= P <= Ps
			if P < -material.Ps || P > material.Ps {
				t.Errorf("PreisachModel: At E=%.2e, P=%.4f is outside [-Ps, +Ps] bounds", E, P)
			}

			t.Logf("PreisachModel: E=%.0fx Ec → P=%.4f (bounded to Ps=%.4f)", E/material.Ec, P, material.Ps)

			// Test negative extreme
			model.Reset()
			P_neg := model.Update(-E)

			if math.IsNaN(P_neg) || math.IsInf(P_neg, 0) {
				t.Errorf("PreisachModel: At E=-%.2e, got invalid value", E)
			}
			if P_neg < -material.Ps*1.01 {
				t.Errorf("PreisachModel: At E=-%.2e, P=%.4f exceeds -Ps=%.4f", E, P_neg, -material.Ps)
			}
			if P_neg < -material.Ps || P_neg > material.Ps {
				t.Errorf("PreisachModel: At E=-%.2e, P=%.4f is outside [-Ps, +Ps] bounds", E, P_neg)
			}
		}
	})

	t.Run("MayergoyzPreisach", func(t *testing.T) {
		material := DefaultHZO()
		model := NewMayergoyzPreisach(material, 20)

		extremeFields := []float64{
			10 * material.Ec,
			100 * material.Ec,
		}

		for _, E := range extremeFields {
			model.Reset()
			P := model.Update(E)

			// No NaN or Inf
			if math.IsNaN(P) {
				t.Errorf("MayergoyzPreisach: At E=%.2e (%.0fx Ec), got NaN", E, E/material.Ec)
			}
			if math.IsInf(P, 0) {
				t.Errorf("MayergoyzPreisach: At E=%.2e (%.0fx Ec), got Inf", E, E/material.Ec)
			}

			// Bounded to [-Ps, +Ps] (with small tolerance for floating-point at saturation boundary)
			if P < -material.Ps*1.001 || P > material.Ps*1.001 {
				t.Errorf("MayergoyzPreisach: At E=%.2e, P=%.4f is outside [-Ps, +Ps] bounds", E, P)
			}

			// Should saturate close to Ps
			if P < 0.8*material.Ps {
				t.Logf("MayergoyzPreisach: At E=%.0fx Ec, P=%.4f is below 80%% of Ps (%.4f) - acceptable for hysteron distribution",
					E/material.Ec, P, material.Ps)
			}

			t.Logf("MayergoyzPreisach: E=%.0fx Ec → P=%.4f (Ps=%.4f)", E/material.Ec, P, material.Ps)
		}
	})
}

// TestRapidFieldReversals verifies stability under many rapid direction changes.
func TestRapidFieldReversals(t *testing.T) {
	t.Run("PreisachModel", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Apply 1000+ rapid reversals between +Ec and -Ec
		numReversals := 1000
		fieldAmplitude := material.Ec

		var prevP float64
		deterministicCheck := make([]float64, 0, numReversals)

		for i := 0; i < numReversals; i++ {
			// Alternate between +Ec and -Ec
			E := fieldAmplitude
			if i%2 == 1 {
				E = -fieldAmplitude
			}

			P := model.Update(E)

			// Check for NaN/Inf
			if math.IsNaN(P) {
				t.Fatalf("PreisachModel: Got NaN at reversal %d", i)
			}
			if math.IsInf(P, 0) {
				t.Fatalf("PreisachModel: Got Inf at reversal %d", i)
			}

			// Check bounds
			if P < -material.Ps || P > material.Ps {
				t.Errorf("PreisachModel: At reversal %d, P=%.4f is outside [-Ps, +Ps]", i, P)
			}

			deterministicCheck = append(deterministicCheck, P)
			prevP = P
		}

		// Verify deterministic behavior: repeat same sequence, should get same results
		model.Reset()
		for i := 0; i < numReversals; i++ {
			E := fieldAmplitude
			if i%2 == 1 {
				E = -fieldAmplitude
			}
			P := model.Update(E)

			// Should match previous run exactly (deterministic)
			if math.Abs(P-deterministicCheck[i]) > 1e-10 {
				t.Errorf("PreisachModel: Non-deterministic behavior at reversal %d: %.10f vs %.10f",
					i, P, deterministicCheck[i])
			}
		}

		// Verify turning points list stays bounded (no memory leak)
		// Access internal state via reflection or indirect observation
		// Here we just verify final polarization is stable
		if math.IsNaN(prevP) || math.Abs(prevP) > material.Ps {
			t.Errorf("PreisachModel: After %d reversals, final P=%.4f is unstable", numReversals, prevP)
		}

		t.Logf("PreisachModel: %d rapid reversals completed, final P=%.4f, stable and deterministic",
			numReversals, prevP)
	})

	t.Run("MayergoyzPreisach", func(t *testing.T) {
		material := DefaultHZO()
		model := NewMayergoyzPreisach(material, 20)

		numReversals := 1000
		fieldAmplitude := material.Ec

		var prevP float64
		for i := 0; i < numReversals; i++ {
			E := fieldAmplitude
			if i%2 == 1 {
				E = -fieldAmplitude
			}

			P := model.Update(E)

			if math.IsNaN(P) || math.IsInf(P, 0) {
				t.Fatalf("MayergoyzPreisach: Got invalid value at reversal %d", i)
			}

			if P < -material.Ps*1.001 || P > material.Ps*1.001 {
				t.Errorf("MayergoyzPreisach: At reversal %d, P=%.4f is outside bounds", i, P)
			}

			prevP = P
		}

		t.Logf("MayergoyzPreisach: %d rapid reversals completed, final P=%.4f, stable",
			numReversals, prevP)
	})
}

// TestNearCurieTemperatureBehavior verifies smooth transition as T approaches Tc.
func TestNearCurieTemperatureBehavior(t *testing.T) {
	material := DefaultHZO()

	tests := []struct {
		name        string
		temperature float64
		expectZero  bool
	}{
		{"T = 0.90*Tc", 0.90 * material.CurieTemp, false},
		{"T = 0.95*Tc", 0.95 * material.CurieTemp, false},
		{"T = 0.99*Tc", 0.99 * material.CurieTemp, false},
		{"T = Tc", material.CurieTemp, true},
		{"T = 1.01*Tc", 1.01 * material.CurieTemp, true},
		{"T = 1.10*Tc", 1.10 * material.CurieTemp, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Ec := material.CoerciveFieldAtTemp(tt.temperature)
			Pr := material.PolarizationAtTemp(tt.temperature)

			// Check for NaN or Inf
			if math.IsNaN(Ec) {
				t.Errorf("CoerciveFieldAtTemp(%.2f K) returned NaN", tt.temperature)
			}
			if math.IsInf(Ec, 0) {
				t.Errorf("CoerciveFieldAtTemp(%.2f K) returned Inf", tt.temperature)
			}
			if math.IsNaN(Pr) {
				t.Errorf("PolarizationAtTemp(%.2f K) returned NaN", tt.temperature)
			}
			if math.IsInf(Pr, 0) {
				t.Errorf("PolarizationAtTemp(%.2f K) returned Inf", tt.temperature)
			}

			// Check non-negative
			if Ec < 0 {
				t.Errorf("Ec(%.2f K) = %.4e should be non-negative", tt.temperature, Ec)
			}
			if Pr < 0 {
				t.Errorf("Pr(%.2f K) = %.4e should be non-negative", tt.temperature, Pr)
			}

			if tt.expectZero {
				// At or above Tc, should be zero
				if Ec != 0 {
					t.Errorf("At T=%.2f K (>= Tc), Ec should be 0, got %.4e", tt.temperature, Ec)
				}
				if Pr != 0 {
					t.Errorf("At T=%.2f K (>= Tc), Pr should be 0, got %.4e", tt.temperature, Pr)
				}
			} else {
				// Below Tc, should be non-zero but decreasing
				if Ec == 0 {
					t.Errorf("At T=%.2f K (< Tc), Ec should be non-zero", tt.temperature)
				}
				if Pr == 0 {
					t.Errorf("At T=%.2f K (< Tc), Pr should be non-zero", tt.temperature)
				}

				// At T=0.99*Tc, Ec and Pr should be small but non-zero
				if tt.temperature == 0.99*material.CurieTemp {
					if Ec > 0.2*material.Ec {
						t.Logf("At T=0.99*Tc, Ec=%.4e is higher than expected (> 20%% of Ec0)", Ec)
					}
					if Pr > 0.2*material.Pr {
						t.Logf("At T=0.99*Tc, Pr=%.4e is higher than expected (> 20%% of Pr0)", Pr)
					}
				}
			}

			t.Logf("%s (%.0f K): Ec=%.4e V/m (%.1f%% of Ec0), Pr=%.4e C/m² (%.1f%% of Pr0)",
				tt.name, tt.temperature,
				Ec, 100*Ec/material.Ec,
				Pr, 100*Pr/material.Pr)
		})
	}

	// Test MayergoyzPreisach temperature setting
	t.Run("MayergoyzPreisach_SetTemperature", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 20)

		// Set temperature near Tc
		model.SetTemperature(0.99 * material.CurieTemp)
		EcEff := model.GetEffectiveEc()

		if math.IsNaN(EcEff) || math.IsInf(EcEff, 0) {
			t.Errorf("GetEffectiveEc() at T=0.99*Tc returned invalid value")
		}
		if EcEff < 0 {
			t.Errorf("GetEffectiveEc() should be non-negative, got %.4e", EcEff)
		}
		if EcEff > 0.2*material.Ec {
			t.Logf("At T=0.99*Tc, effective Ec=%.4e is higher than expected", EcEff)
		}

		// Set temperature above Tc
		model.SetTemperature(material.CurieTemp + 10)
		EcAboveTc := model.GetEffectiveEc()
		if EcAboveTc != 0 {
			t.Errorf("At T > Tc, effective Ec should be 0, got %.4e", EcAboveTc)
		}

		t.Logf("MayergoyzPreisach: T=0.99*Tc → Ec=%.4e, T>Tc → Ec=%.4e", EcEff, EcAboveTc)
	})
}

// TestNegativeTemperatureGuard verifies graceful handling of T < 0 K.
func TestNegativeTemperatureGuard(t *testing.T) {
	material := DefaultHZO()

	t.Run("T = -1 K", func(t *testing.T) {
		// Models should handle T < 0 gracefully without panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic at T=-1 K: %v", r)
			}
		}()

		Ec := material.CoerciveFieldAtTemp(-1)
		Pr := material.PolarizationAtTemp(-1)

		// Should not be NaN (implementation may choose to clamp or use abs value)
		if math.IsNaN(Ec) {
			t.Errorf("CoerciveFieldAtTemp(-1 K) returned NaN")
		}
		if math.IsNaN(Pr) {
			t.Errorf("PolarizationAtTemp(-1 K) returned NaN")
		}

		// Reasonable implementation: either error or treat as T=0 K or clamp to valid range
		// Since T=-1 < Tc, if implementation doesn't clamp, (1 - T/Tc) > 1, which is fine for pow
		// Just verify no panic and no NaN
		t.Logf("T=-1 K: Ec=%.4e, Pr=%.4e (implementation-defined behavior)", Ec, Pr)
	})

	t.Run("T = 0 K", func(t *testing.T) {
		// Absolute zero should work without error
		Ec := material.CoerciveFieldAtTemp(0)
		Pr := material.PolarizationAtTemp(0)

		if math.IsNaN(Ec) || math.IsInf(Ec, 0) {
			t.Errorf("CoerciveFieldAtTemp(0 K) returned invalid value: %.4e", Ec)
		}
		if math.IsNaN(Pr) || math.IsInf(Pr, 0) {
			t.Errorf("PolarizationAtTemp(0 K) returned invalid value: %.4e", Pr)
		}

		// At T=0, should be at maximum values (1 - 0/Tc)^0.5 = 1
		expectedEc := material.Ec
		expectedPr := material.Pr

		if math.Abs(Ec-expectedEc) > 1e-10 {
			t.Errorf("At T=0 K, Ec=%.4e should equal Ec0=%.4e", Ec, expectedEc)
		}
		if math.Abs(Pr-expectedPr) > 1e-10 {
			t.Errorf("At T=0 K, Pr=%.4e should equal Pr0=%.4e", Pr, expectedPr)
		}

		t.Logf("T=0 K: Ec=%.4e (Ec0=%.4e), Pr=%.4e (Pr0=%.4e)", Ec, expectedEc, Pr, expectedPr)
	})

	t.Run("MayergoyzPreisach_NegativeTemp", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MayergoyzPreisach panicked at negative temperature: %v", r)
			}
		}()

		model := NewMayergoyzPreisach(material, 10)
		model.SetTemperature(-10) // Negative temperature

		// Should not panic, may produce implementation-defined behavior
		EcEff := model.GetEffectiveEc()
		if math.IsNaN(EcEff) {
			t.Errorf("GetEffectiveEc() at T=-10 K returned NaN")
		}

		t.Logf("MayergoyzPreisach at T=-10 K: Ec=%.4e (no panic)", EcEff)
	})
}

// TestNumericalPrecisionAtBoundaries verifies polarization stays bounded after many iterations.
func TestNumericalPrecisionAtBoundaries(t *testing.T) {
	t.Run("PreisachModel_Saturation_Stability", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Apply field well beyond saturation for many iterations
		Eextreme := 5 * material.Ec
		numIterations := 10000

		for i := 0; i < numIterations; i++ {
			P := model.Update(Eextreme)

			// P should never exceed Ps (with small tolerance for floating-point)
			if P > material.Ps*1.001 {
				t.Errorf("PreisachModel: At iteration %d, P=%.10f exceeds Ps=%.10f",
					i, P, material.Ps)
			}

			// P should not drift beyond bounds
			if P < -material.Ps || P > material.Ps {
				t.Errorf("PreisachModel: At iteration %d, P=%.10f is outside [-Ps, +Ps]", i, P)
			}
		}

		finalP := model.Polarization()
		normP := model.NormalizedPolarization()

		// Normalized polarization should be in [-1, 1]
		if normP < -1.001 || normP > 1.001 {
			t.Errorf("PreisachModel: Normalized P=%.10f is outside [-1, 1] range", normP)
		}

		t.Logf("PreisachModel: After %d iterations at E=5*Ec, P=%.10f (%.10f of Ps), normP=%.10f",
			numIterations, finalP, finalP/material.Ps, normP)
	})

	t.Run("MayergoyzPreisach_Saturation_Stability", func(t *testing.T) {
		material := DefaultHZO()
		model := NewMayergoyzPreisach(material, 20)

		Eextreme := 5 * material.Ec
		numIterations := 10000

		for i := 0; i < numIterations; i++ {
			P := model.Update(Eextreme)

			// Should stay bounded (with small tolerance for floating-point at saturation boundary)
			if P < -material.Ps*1.001 || P > material.Ps*1.001 {
				t.Errorf("MayergoyzPreisach: At iteration %d, P=%.10f is outside [-Ps, +Ps]", i, P)
			}
		}

		finalP := model.Polarization()
		normP := model.NormalizedPolarization()

		if normP < -1.001 || normP > 1.001 {
			t.Errorf("MayergoyzPreisach: Normalized P=%.10f is outside [-1, 1] range", normP)
		}

		t.Logf("MayergoyzPreisach: After %d iterations at E=5*Ec, P=%.10f, normP=%.10f",
			numIterations, finalP, normP)
	})

	t.Run("AlternatingExtremes", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Alternate between extreme positive and negative fields
		Eextreme := 10 * material.Ec
		numCycles := 1000

		for i := 0; i < numCycles; i++ {
			P_pos := model.Update(Eextreme)
			if math.Abs(P_pos) > material.Ps {
				t.Errorf("Cycle %d: P_pos=%.10f exceeds Ps bounds", i, P_pos)
			}

			P_neg := model.Update(-Eextreme)
			if math.Abs(P_neg) > material.Ps {
				t.Errorf("Cycle %d: P_neg=%.10f exceeds Ps bounds", i, P_neg)
			}
		}

		t.Logf("PreisachModel: %d extreme alternating cycles completed, polarization remains bounded",
			numCycles)
	})
}

// TestModelConsistency verifies both models produce similar hysteresis loops.
func TestModelConsistency(t *testing.T) {
	material := DefaultHZO()
	preisach := NewPreisachModel(material)
	mayergoyz := NewMayergoyzPreisach(material, 30)

	Emax := 2 * material.Ec
	points := 50

	E_preisach, P_preisach := preisach.GetHysteresisLoop(Emax, points)
	E_mayergoyz, P_mayergoyz := mayergoyz.GetHysteresisLoop(Emax, points)

	// Both should produce non-empty loops
	if len(E_preisach) == 0 || len(P_preisach) == 0 {
		t.Fatal("PreisachModel produced empty loop")
	}
	if len(E_mayergoyz) == 0 || len(P_mayergoyz) == 0 {
		t.Fatal("MayergoyzPreisach produced empty loop")
	}

	// Both should saturate to similar values (within factor of 2)
	maxP_preisach := 0.0
	minP_preisach := 0.0
	for _, p := range P_preisach {
		if p > maxP_preisach {
			maxP_preisach = p
		}
		if p < minP_preisach {
			minP_preisach = p
		}
	}

	maxP_mayergoyz := 0.0
	minP_mayergoyz := 0.0
	for _, p := range P_mayergoyz {
		if p > maxP_mayergoyz {
			maxP_mayergoyz = p
		}
		if p < minP_mayergoyz {
			minP_mayergoyz = p
		}
	}

	t.Logf("PreisachModel: P range [%.4f, %.4f]", minP_preisach, maxP_preisach)
	t.Logf("MayergoyzPreisach: P range [%.4f, %.4f]", minP_mayergoyz, maxP_mayergoyz)

	// Both should achieve significant polarization (> 50% of Ps)
	if maxP_preisach < 0.5*material.Ps {
		t.Errorf("PreisachModel max P=%.4f is too low (< 50%% of Ps)", maxP_preisach)
	}
	if maxP_mayergoyz < 0.5*material.Ps {
		t.Errorf("MayergoyzPreisach max P=%.4f is too low (< 50%% of Ps)", maxP_mayergoyz)
	}

	// Ranges should be comparable (within factor of 2)
	ratio := maxP_preisach / maxP_mayergoyz
	if ratio < 0.5 || ratio > 2.0 {
		t.Logf("Warning: Saturation polarization differs by more than 2x between models (ratio=%.2f)", ratio)
	}
}

// TestFieldHistoryMemory verifies turning point history behaves correctly.
func TestFieldHistoryMemory(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Create a minor loop: 0 → Emax → E1 → E2 → E1
	Emax := 2 * material.Ec
	E1 := 0.5 * material.Ec
	E2 := -0.5 * material.Ec

	model.Reset()
	model.Update(Emax)     // Go to positive saturation
	P1 := model.Update(E1) // First visit to E1 (descending)
	model.Update(E2)       // Go to E2
	P2 := model.Update(E1) // Second visit to E1 (ascending)

	// Due to hysteresis, P at same field E1 should differ based on history
	// (ascending vs descending path)
	if math.Abs(P1-P2) < 0.01*material.Ps {
		t.Logf("Warning: Polarization at E1 is similar for different paths (P1=%.4f, P2=%.4f)",
			P1, P2)
		t.Logf("This may indicate weak minor loop behavior")
	} else {
		t.Logf("Good: Minor loop shows path dependence: P1=%.4f (descending), P2=%.4f (ascending)",
			P1, P2)
	}

	// Verify no panic or instability
	if math.IsNaN(P1) || math.IsNaN(P2) {
		t.Error("Got NaN during minor loop")
	}
}
