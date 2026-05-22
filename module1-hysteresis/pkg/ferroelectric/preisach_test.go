package ferroelectric

import (
	"math"
	"testing"
)

// TestTanhEverett_Calculate_Positive tests that large positive alpha and large negative beta
// yield a result near Ps.
func TestTanhEverett_Calculate_Positive(t *testing.T) {
	t.Run("HighFieldSaturation", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 2e5,
		}
		// alpha >> Ec, beta << -Ec
		alpha := ev.Ec * 10
		beta := -ev.Ec * 10
		result := ev.Calculate(alpha, beta)
		// Should be close to Ps
		if math.Abs(result-ev.Ps) > 0.01*ev.Ps {
			t.Errorf("Expected result near Ps=%f, got %f", ev.Ps, result)
		}
	})
}

// TestTanhEverett_Calculate_Zero tests behavior when alpha and beta are both zero.
func TestTanhEverett_Calculate_Zero(t *testing.T) {
	t.Run("BothZero", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 2e5,
		}
		result := ev.Calculate(0, 0)
		// Should be some intermediate value, not zero and not Ps
		if result < 0 || result > ev.Ps {
			t.Errorf("Expected result in [0, Ps], got %f", result)
		}
	})
}

// TestTanhEverett_Calculate_Symmetric tests symmetric field behavior.
func TestTanhEverett_Calculate_Symmetric(t *testing.T) {
	t.Run("SymmetricFields", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 2e5,
		}
		// Test several symmetric cases
		fields := []float64{0.5e6, 1.0e6, 2.0e6, 5.0e6}
		for _, E := range fields {
			result := ev.Calculate(E, -E)
			// Should be positive and within bounds
			if result < 0 || result > ev.Ps {
				t.Errorf("Calculate(%f, %f) = %f, expected in [0, Ps=%f]", E, -E, result, ev.Ps)
			}
		}
	})
}

// TestTanhEverett_Calculate_Monotonic tests that increasing alpha increases the result.
func TestTanhEverett_Calculate_Monotonic(t *testing.T) {
	t.Run("IncreasingAlpha", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 2e5,
		}
		beta := -2e6
		prev := 0.0
		for i := -10; i <= 10; i++ {
			alpha := float64(i) * 5e5
			result := ev.Calculate(alpha, beta)
			if i > -10 && result < prev-1e-9 { // Allow tiny numerical error
				t.Errorf("Non-monotonic: Calculate(%f, %f) = %f < prev %f", alpha, beta, result, prev)
			}
			prev = result
		}
	})
}

// TestTanhEverett_Calculate_Clamping tests that results are always in [0, Ps].
func TestTanhEverett_Calculate_Clamping(t *testing.T) {
	t.Run("BoundsChecking", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 2e5,
		}
		// Test many combinations
		testCases := []struct {
			alpha, beta float64
		}{
			{10e6, -10e6},
			{-10e6, 10e6},
			{0, 0},
			{1e6, -1e6},
			{5e6, 5e6}, // Same value
			{-5e6, -5e6},
		}
		for _, tc := range testCases {
			result := ev.Calculate(tc.alpha, tc.beta)
			if result < 0 || result > ev.Ps {
				t.Errorf("Calculate(%f, %f) = %f, expected in [0, %f]", tc.alpha, tc.beta, result, ev.Ps)
			}
		}
	})
}

// TestTanhEverett_Calculate_SmallDelta tests that small Delta gives sharp transition.
func TestTanhEverett_Calculate_SmallDelta(t *testing.T) {
	t.Run("SharpTransition", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 1e4, // Very small Delta
		}
		// At fields far from Ec, should be near extremes
		highResult := ev.Calculate(ev.Ec*3, -ev.Ec*3)
		lowResult := ev.Calculate(-ev.Ec*3, ev.Ec*3)

		// High field should give near Ps
		if highResult < 0.9*ev.Ps {
			t.Errorf("Small Delta high field: expected near Ps, got %f", highResult)
		}
		// Low field (reversed) should give near zero due to clamping
		if lowResult > 0.1*ev.Ps {
			t.Errorf("Small Delta low field: expected near 0, got %f", lowResult)
		}
	})
}

// TestTanhEverett_Calculate_LargeDelta tests that large Delta gives gradual transition.
func TestTanhEverett_Calculate_LargeDelta(t *testing.T) {
	t.Run("GradualTransition", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 1e6, // Large Delta (same as Ec)
		}
		// Results should vary more smoothly across field range
		results := make([]float64, 5)
		for i := 0; i < 5; i++ {
			alpha := -2e6 + float64(i)*1e6
			beta := -3e6
			results[i] = ev.Calculate(alpha, beta)
		}
		// Check for smooth progression (no large jumps)
		for i := 1; i < len(results); i++ {
			diff := math.Abs(results[i] - results[i-1])
			if diff > 0.5*ev.Ps {
				t.Errorf("Large jump in large Delta case: %f -> %f", results[i-1], results[i])
			}
		}
	})
}

// TestTanhEverett_Calculate_AlphaBetaEqual tests behavior when alpha == beta.
func TestTanhEverett_Calculate_AlphaBetaEqual(t *testing.T) {
	t.Run("EqualFields", func(t *testing.T) {
		ev := &TanhEverett{
			Ps:    0.3,
			Ec:    1e6,
			Delta: 2e5,
		}
		// When alpha == beta, the triangle area should be zero or very small
		testFields := []float64{0, 1e6, -1e6, 5e6}
		for _, E := range testFields {
			result := ev.Calculate(E, E)
			// Should be very small (clamped to 0 if negative)
			if result > 0.05*ev.Ps {
				t.Errorf("Calculate(%f, %f) = %f, expected near 0", E, E, result)
			}
		}
	})
}

// TestTuneDeltaForPr_ValidInput tests that tuneDeltaForPr returns positive Delta.
func TestTuneDeltaForPr_ValidInput(t *testing.T) {
	t.Run("ReturnsPositiveDelta", func(t *testing.T) {
		ec := 1e6
		satE := 5e6
		psIrrev := 0.25
		targetPr := 0.15

		delta := tuneDeltaForPr(ec, satE, psIrrev, targetPr)
		if delta <= 0 {
			t.Errorf("Expected positive Delta, got %f", delta)
		}
	})
}

// TestTuneDeltaForPr_ZeroEc tests behavior when Ec is zero.
func TestTuneDeltaForPr_ZeroEc(t *testing.T) {
	t.Run("ZeroEc", func(t *testing.T) {
		delta := tuneDeltaForPr(0, 5e6, 0.25, 0.15)
		if delta != 0 {
			t.Errorf("Expected 0 for zero Ec, got %f", delta)
		}
	})

	t.Run("NegativeEc", func(t *testing.T) {
		delta := tuneDeltaForPr(-1e6, 5e6, 0.25, 0.15)
		if delta != 0 {
			t.Errorf("Expected 0 for negative Ec, got %f", delta)
		}
	})
}

// TestTuneDeltaForPr_ZeroPr tests behavior when Pr is zero.
func TestTuneDeltaForPr_ZeroPr(t *testing.T) {
	t.Run("ZeroTargetPr", func(t *testing.T) {
		ec := 1e6
		delta := tuneDeltaForPr(ec, 5e6, 0.25, 0)
		expected := ec * 0.25
		if math.Abs(delta-expected) > 1e-6 {
			t.Errorf("Expected default %f for zero Pr, got %f", expected, delta)
		}
	})

	t.Run("ZeroPsIrrev", func(t *testing.T) {
		ec := 1e6
		delta := tuneDeltaForPr(ec, 5e6, 0, 0.15)
		expected := ec * 0.25
		if math.Abs(delta-expected) > 1e-6 {
			t.Errorf("Expected default %f for zero PsIrrev, got %f", expected, delta)
		}
	})
}

// TestTuneDeltaForPr_HighRatio tests that high Pr/Ps ratio gives smaller Delta.
// Note: Smaller Delta (narrower distribution) concentrates hysterons near Ec,
// leading to higher remanent polarization.
func TestTuneDeltaForPr_HighRatio(t *testing.T) {
	t.Run("HighPrPsRatio", func(t *testing.T) {
		ec := 1e6
		satE := 5e6
		psIrrev := 0.30

		// High ratio (Pr close to Ps)
		highPr := 0.28
		deltaHigh := tuneDeltaForPr(ec, satE, psIrrev, highPr)

		// Medium ratio
		medPr := 0.15
		deltaMed := tuneDeltaForPr(ec, satE, psIrrev, medPr)

		// Higher ratio should give smaller Delta (narrower distribution)
		if deltaHigh >= deltaMed {
			t.Errorf("Expected smaller Delta for higher Pr/Ps ratio: deltaHigh=%f >= deltaMed=%f", deltaHigh, deltaMed)
		}
	})
}

// TestTuneDeltaForPr_LowRatio tests that low Pr/Ps ratio gives larger Delta.
// Note: Larger Delta (broader distribution) spreads hysterons more,
// leading to lower remanent polarization.
func TestTuneDeltaForPr_LowRatio(t *testing.T) {
	t.Run("LowPrPsRatio", func(t *testing.T) {
		ec := 1e6
		satE := 5e6
		psIrrev := 0.30

		// Low ratio
		lowPr := 0.05
		deltaLow := tuneDeltaForPr(ec, satE, psIrrev, lowPr)

		// Medium ratio
		medPr := 0.20
		deltaMed := tuneDeltaForPr(ec, satE, psIrrev, medPr)

		// Lower ratio should give larger Delta (broader distribution)
		if deltaLow <= deltaMed {
			t.Errorf("Expected larger Delta for lower Pr/Ps ratio: deltaLow=%f <= deltaMed=%f", deltaLow, deltaMed)
		}
	})
}

// TestPreisachModel_NewPreisachModel tests model creation.
func TestPreisachModel_NewPreisachModel(t *testing.T) {
	t.Run("CreateModel", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		if model == nil {
			t.Fatal("Expected non-nil model")
		}
		if model.material != material {
			t.Error("Model material not set correctly")
		}
		if model.stack == nil {
			t.Error("Model stack not initialized")
		}
		if model.everett == nil {
			t.Error("Model everett not initialized")
		}
		if model.Temperature != 300.0 {
			t.Errorf("Expected default temperature 300K, got %f", model.Temperature)
		}
		if model.Stress != 1.0 {
			t.Errorf("Expected default stress 1 GPa, got %f", model.Stress)
		}
	})
}

func TestPreisachModel_NewPreisachModelRejectsNilMaterial(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected nil material to be rejected without panic, got panic: %v", r)
		}
	}()

	model := NewPreisachModel(nil)
	if model != nil {
		t.Fatalf("expected nil model for nil material, got %#v", model)
	}
}

// TestPreisachModel_DiscreteStates tests discrete state generation.
func TestPreisachModel_DiscreteStates(t *testing.T) {
	t.Run("GenerateStates", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		n := 30
		states := model.DiscreteStates(n)

		if len(states) != n {
			t.Errorf("Expected %d states, got %d", n, len(states))
		}

		// Check bounds
		if states[0].Polarization > -material.Ps+1e-6 {
			t.Errorf("First state should be near -Ps, got %f", states[0].Polarization)
		}
		if states[n-1].Polarization < material.Ps-1e-6 {
			t.Errorf("Last state should be near +Ps, got %f", states[n-1].Polarization)
		}

		// Check monotonic increase
		for i := 1; i < n; i++ {
			if states[i].Polarization <= states[i-1].Polarization {
				t.Errorf("States not monotonic at index %d", i)
			}
		}

		// Check levels are 1-indexed
		if states[0].Level != 1 {
			t.Errorf("Expected first level to be 1, got %d", states[0].Level)
		}
		if states[n-1].Level != n {
			t.Errorf("Expected last level to be %d, got %d", n, states[n-1].Level)
		}
	})
}

// TestPreisachModel_Update tests field update and polarization calculation.
func TestPreisachModel_Update(t *testing.T) {
	t.Run("BasicUpdate", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Apply saturation field
		E := material.Ec * 3
		P := model.Update(E)

		// Should get some polarization
		if math.Abs(P) > material.Ps*1.5 {
			t.Errorf("Polarization %f exceeds reasonable bounds for Ps=%f", P, material.Ps)
		}
	})

	t.Run("HysteresisLoop", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		Emax := material.Ec * 4

		// Reset to negative saturation
		model.Update(-Emax)
		P_neg := model.Polarization()

		// Go to positive saturation
		model.Update(Emax)
		P_pos := model.Polarization()

		// Positive should be higher than negative
		if P_pos <= P_neg {
			t.Errorf("Expected P_pos > P_neg, got P_pos=%f, P_neg=%f", P_pos, P_neg)
		}

		// Return to zero field
		P_rem := model.Update(0)

		// Remanent should be between zero and saturation
		if P_rem < 0 || P_rem > P_pos {
			t.Errorf("Remanent polarization %f out of expected range [0, %f]", P_rem, P_pos)
		}
	})
}

// TestPreisachModel_Reset tests model reset functionality.
func TestPreisachModel_Reset(t *testing.T) {
	t.Run("ResetClearsHistory", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Apply some field history
		model.Update(material.Ec * 2)
		model.Update(0)

		// Reset
		model.Reset()

		// After reset, stack should be reinitialized
		if model.stack == nil {
			t.Error("Stack should not be nil after reset")
		}

		// Polarization at zero field after reset should be at negative saturation state
		P_after_reset := model.Update(-material.Ec * 5)
		// This should be near -Ps
		if P_after_reset > 0 {
			t.Errorf("After reset and negative saturation, expected negative P, got %f", P_after_reset)
		}
	})
}

// TestPreisachModel_Polarization tests polarization query without mutation.
func TestPreisachModel_Polarization(t *testing.T) {
	t.Run("QueryPolarization", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Set a state
		E := material.Ec * 2
		P_update := model.Update(E)
		P_query := model.Polarization()

		// Should match
		if math.Abs(P_update-P_query) > 1e-9 {
			t.Errorf("Polarization query mismatch: Update=%f, Query=%f", P_update, P_query)
		}
	})
}

// TestPreisachModel_NormalizedPolarization tests normalized polarization.
func TestPreisachModel_NormalizedPolarization(t *testing.T) {
	t.Run("Normalized", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// At saturation
		model.Update(material.Ec * 5)
		P_norm := model.NormalizedPolarization()

		// Should be close to 1
		if P_norm < 0.5 || P_norm > 1.5 {
			t.Errorf("Expected normalized P near 1, got %f", P_norm)
		}
	})
}

// TestPreisachModel_GetHysteresisLoop tests loop generation.
func TestPreisachModel_GetHysteresisLoop(t *testing.T) {
	t.Run("GenerateLoop", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		Emax := material.Ec * 4
		points := 50
		E, P := model.GetHysteresisLoop(Emax, points)

		// Should have 4*points entries (ascending + descending)
		expectedLen := points * 4
		if len(E) < expectedLen {
			t.Errorf("Expected at least %d points, got %d", expectedLen, len(E))
		}
		if len(E) != len(P) {
			t.Errorf("E and P arrays have different lengths: %d vs %d", len(E), len(P))
		}

		// Check field range
		maxE := 0.0
		minE := 0.0
		for _, e := range E {
			if e > maxE {
				maxE = e
			}
			if e < minE {
				minE = e
			}
		}
		if math.Abs(maxE-Emax) > 1e-6 {
			t.Errorf("Expected max field %f, got %f", Emax, maxE)
		}
		if math.Abs(minE+Emax) > 1e-6 {
			t.Errorf("Expected min field %f, got %f", -Emax, minE)
		}
	})
}

// TestPreisachModel_GetHysteresisLoopRejectsNonPhysicalInputs verifies invalid loop requests return no data and preserve model state.
func TestPreisachModel_GetHysteresisLoopRejectsNonPhysicalInputs(t *testing.T) {
	material := DefaultHZO()
	tests := []struct {
		name   string
		Emax   float64
		points int
	}{
		{name: "zero field", Emax: 0, points: 50},
		{name: "negative field", Emax: -material.Ec, points: 50},
		{name: "nan field", Emax: math.NaN(), points: 50},
		{name: "infinite field", Emax: math.Inf(1), points: 50},
		{name: "zero points", Emax: material.Ec, points: 0},
		{name: "negative points", Emax: material.Ec, points: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewPreisachModel(material)
			initialP := model.Update(2 * material.Ec)

			E, P := model.GetHysteresisLoop(tt.Emax, tt.points)

			if len(E) != 0 || len(P) != 0 {
				t.Fatalf("expected no loop data for Emax=%.3e V/m points=%d, got E=%d P=%d", tt.Emax, tt.points, len(E), len(P))
			}
			if got := model.Polarization(); math.Abs(got-initialP) > 1e-12 {
				t.Fatalf("invalid loop request changed polarization: got %.12e C/m² want %.12e C/m²", got, initialP)
			}
		})
	}
}

// TestPreisachModel_SetTemperature tests temperature effects.
func TestPreisachModel_SetTemperature(t *testing.T) {
	t.Run("TemperatureScaling", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Get Ec at room temperature
		Ec_300K := model.GetEffectiveEc()

		// Increase temperature
		model.SetTemperature(400.0)
		Ec_400K := model.GetEffectiveEc()

		// Ec should change (typically decreases with temperature for ferroelectrics)
		if Ec_400K == Ec_300K {
			t.Errorf("Temperature change did not affect Ec")
		}

		// Check it's still positive
		if Ec_400K <= 0 {
			t.Errorf("Effective Ec should be positive, got %f", Ec_400K)
		}
	})

	t.Run("LowTemperature", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Very low temperature
		model.SetTemperature(100.0)
		Ec_low := model.GetEffectiveEc()

		// Should still be positive (clamped to minimum)
		if Ec_low < 1e5 {
			t.Errorf("Ec at low temp should be clamped to minimum 1e5, got %f", Ec_low)
		}
	})
}

// TestPreisachModel_SetStress tests stress effects.
func TestPreisachModel_SetStress(t *testing.T) {
	t.Run("StressScaling", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Get Ec at default stress (1 GPa)
		Ec_1GPa := model.GetEffectiveEc()

		// Increase stress
		model.SetStress(2.0)
		Ec_2GPa := model.GetEffectiveEc()

		// Ec should change with stress due to electrostriction coupling.
		if math.Abs(Ec_2GPa-Ec_1GPa) < math.Abs(Ec_1GPa)*1e-6 {
			t.Errorf("Expected Ec to change with stress: Ec_1GPa=%f, Ec_2GPa=%f", Ec_1GPa, Ec_2GPa)
		}
	})

	t.Run("ZeroStress", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		model.SetStress(0)
		Ec_0 := model.GetEffectiveEc()

		// Should still be valid
		if Ec_0 <= 0 {
			t.Errorf("Ec at zero stress should be positive, got %f", Ec_0)
		}
	})
}

// TestPreisachModel_ReversiblePolarization tests the reversible component.
func TestPreisachModel_ReversiblePolarization(t *testing.T) {
	t.Run("ReversibleContribution", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		// Reversible component should be zero at zero field
		P_rev_zero := model.reversiblePolarization(0)
		if math.Abs(P_rev_zero) > 1e-9 {
			t.Errorf("Expected zero reversible polarization at E=0, got %f", P_rev_zero)
		}

		// Should be non-zero at high field (if material has permittivity)
		P_rev_high := model.reversiblePolarization(material.Ec * 3)
		// Can be zero if no permittivity set, but should be symmetric
		P_rev_neg := model.reversiblePolarization(-material.Ec * 3)

		if math.Abs(P_rev_high+P_rev_neg) > 1e-9 {
			t.Errorf("Reversible polarization not symmetric: +E=%f, -E=%f", P_rev_high, P_rev_neg)
		}
	})
}

// TestPreisachModel_GetEffectiveEc tests effective coercive field getter.
func TestPreisachModel_GetEffectiveEc(t *testing.T) {
	t.Run("GetEffectiveEc", func(t *testing.T) {
		material := DefaultHZO()
		model := NewPreisachModel(material)

		Ec := model.GetEffectiveEc()

		// Should be positive and reasonable
		if Ec <= 0 {
			t.Errorf("Effective Ec should be positive, got %f", Ec)
		}
		if Ec < material.Ec*0.5 || Ec > material.Ec*2.0 {
			t.Errorf("Effective Ec %f seems far from material Ec %f", Ec, material.Ec)
		}
	})
}
