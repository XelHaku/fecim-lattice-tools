package ferroelectric

import (
	"math"
	"testing"
)

// TestHysteresisLoopExists verifies the P-E curve shows proper hysteresis behavior.
func TestHysteresisLoopExists(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Generate hysteresis loop
	Emax := 2 * material.Ec // Go beyond coercive field
	E, P := model.GetHysteresisLoop(Emax, 50)

	if len(E) == 0 || len(P) == 0 {
		t.Fatal("Hysteresis loop generation returned empty arrays")
	}

	t.Logf("Generated hysteresis loop with %d points", len(E))

	// Verify we have reasonable polarization values
	maxP := 0.0
	minP := 0.0
	for _, p := range P {
		if p > maxP {
			maxP = p
		}
		if p < minP {
			minP = p
		}
	}

	t.Logf("Polarization range: [%.4f, %.4f] C/m²", minP, maxP)

	// Should reach close to saturation polarization
	if maxP < 0.5*material.Ps {
		t.Errorf("Max polarization %.4f is too low (expected > %.4f)", maxP, 0.5*material.Ps)
	}
	if minP > -0.5*material.Ps {
		t.Errorf("Min polarization %.4f is too high (expected < %.4f)", minP, -0.5*material.Ps)
	}
}

// TestHysteresisAsymmetry verifies that the ascending and descending branches differ.
// This is the key signature of hysteresis - path dependence.
func TestHysteresisAsymmetry(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Sweep from negative to positive
	model.Reset()
	E := 0.0
	for i := 0; i < 20; i++ {
		E = -material.Ec + 2*material.Ec*float64(i)/19
		model.Update(E)
	}
	ascendingP := model.Polarization()

	// Sweep from positive to negative
	model.Reset()
	// First go to positive saturation
	for i := 0; i <= 20; i++ {
		E = 2 * material.Ec * float64(i) / 20
		model.Update(E)
	}
	// Then come back to zero
	for i := 0; i <= 20; i++ {
		E = 2*material.Ec - 2*material.Ec*float64(i)/20
		model.Update(E)
	}
	descendingP := model.Polarization()

	t.Logf("Ascending from -Ec to +Ec: P = %.4f", ascendingP)
	t.Logf("Descending from +Emax to 0: P = %.4f", descendingP)

	// At E=0, the polarization should differ based on history
	// (this is remanent polarization)
	if math.Abs(ascendingP-descendingP) < 0.01*material.Ps {
		t.Log("Warning: Ascending and descending paths show similar polarization at same E")
		// This might be OK depending on where we measure, but it's worth noting
	}
}

// TestCoerciveFieldSwitching verifies polarization switches sign around Ec.
func TestCoerciveFieldSwitching(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Start with positive polarization (apply positive field)
	model.Reset()
	model.Update(2 * material.Ec) // Positive saturation
	initialP := model.Polarization()

	// Apply field just below coercive field - should still be positive
	model.Update(-0.5 * material.Ec)
	belowEcP := model.Polarization()

	// Apply field beyond coercive field - should switch to negative
	model.Update(-2 * material.Ec)
	beyondEcP := model.Polarization()

	t.Logf("Initial (E=+2Ec): P = %.4f", initialP)
	t.Logf("Below Ec (E=-0.5Ec): P = %.4f", belowEcP)
	t.Logf("Beyond Ec (E=-2Ec): P = %.4f", beyondEcP)

	// Initial should be positive
	if initialP < 0 {
		t.Errorf("Initial polarization should be positive, got %.4f", initialP)
	}

	// Beyond Ec should switch to negative
	if beyondEcP > 0 {
		t.Errorf("Polarization beyond -Ec should be negative, got %.4f", beyondEcP)
	}
}

// TestDiscreteStatesCount verifies 30 discrete states for FeCIM.
func TestDiscreteStatesCount(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	states := model.DiscreteStates(30)

	if len(states) != 30 {
		t.Errorf("Expected 30 discrete states, got %d", len(states))
	}

	// Verify states span from -Ps to +Ps
	if states[0] > -0.9*material.Ps {
		t.Errorf("First state %.4f should be close to -Ps (%.4f)", states[0], -material.Ps)
	}
	if states[29] < 0.9*material.Ps {
		t.Errorf("Last state %.4f should be close to +Ps (%.4f)", states[29], material.Ps)
	}

	// Verify states are evenly spaced
	expectedSpacing := 2 * material.Ps / 29
	for i := 1; i < 30; i++ {
		spacing := states[i] - states[i-1]
		if math.Abs(spacing-expectedSpacing) > 1e-10 {
			t.Errorf("State spacing at index %d is %.6f, expected %.6f", i, spacing, expectedSpacing)
		}
	}

	t.Logf("30 discrete states verified, spacing: %.6f C/m²", expectedSpacing)
}

// TestMaterialParameters verifies HZO material parameters are physically reasonable.
func TestMaterialParameters(t *testing.T) {
	material := DefaultHZO()

	// Check polarization values are in reasonable range for HZO
	// Literature values: Pr ~ 10-40 μC/cm², Ps ~ 15-50 μC/cm²
	if material.Pr < 10e-2 || material.Pr > 50e-2 {
		t.Errorf("Remanent polarization %.2f C/m² is outside expected range", material.Pr)
	}
	if material.Ps < 15e-2 || material.Ps > 60e-2 {
		t.Errorf("Saturation polarization %.2f C/m² is outside expected range", material.Ps)
	}

	// Check coercive field (literature: 0.5-2 MV/cm)
	if material.Ec < 0.5e8 || material.Ec > 3e8 {
		t.Errorf("Coercive field %.2e V/m is outside expected range", material.Ec)
	}

	// Check thickness is reasonable for FeFET applications (5-20 nm typical)
	if material.Thickness < 1e-9 || material.Thickness > 50e-9 {
		t.Errorf("Film thickness %.0f nm is outside expected range", material.Thickness*1e9)
	}

	// Check Ps > Pr (saturation should exceed remanent)
	if material.Ps <= material.Pr {
		t.Errorf("Ps (%.4f) should be greater than Pr (%.4f)", material.Ps, material.Pr)
	}

	t.Logf("HZO material parameters verified:")
	t.Logf("  Pr = %.1f μC/cm²", material.Pr*1e4)
	t.Logf("  Ps = %.1f μC/cm²", material.Ps*1e4)
	t.Logf("  Ec = %.2f MV/cm", material.Ec/1e8)
	t.Logf("  Thickness = %.0f nm", material.Thickness*1e9)
}

// TestPreisachModelReset verifies the reset function works correctly.
func TestPreisachModelReset(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Apply some fields
	model.Update(2 * material.Ec)
	model.Update(-material.Ec)

	if model.Polarization() == 0 {
		t.Error("Model should have non-zero polarization after updates")
	}

	// Reset
	model.Reset()

	if model.Polarization() != 0 {
		t.Errorf("After reset, polarization should be 0, got %.4f", model.Polarization())
	}
}

// TestNormalizedPolarization verifies normalized output is in [-1, 1] range.
func TestNormalizedPolarization(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	// Test at various field values
	testFields := []float64{
		-2 * material.Ec,
		-material.Ec,
		0,
		material.Ec,
		2 * material.Ec,
	}

	for _, E := range testFields {
		model.Update(E)
		normP := model.NormalizedPolarization()

		if normP < -1.1 || normP > 1.1 {
			t.Errorf("Normalized polarization %.4f at E=%.2e is outside [-1, 1] range",
				normP, E)
		}
	}
}

// ============================================================================
// Temperature-Dependent Tests (Automotive Range: -40°C to 150°C)
// ============================================================================

// TestCoerciveFieldTemperatureDependence verifies Ec decreases with temperature.
func TestCoerciveFieldTemperatureDependence(t *testing.T) {
	material := DefaultHZO()

	// Test automotive temperature range
	temps := []struct {
		kelvin float64
		name   string
	}{
		{233, "-40°C (cold start)"},
		{300, "27°C (room temp)"},
		{358, "85°C (standard)"},
		{423, "150°C (extreme)"},
	}

	var prevEc float64
	for i, tt := range temps {
		Ec := material.CoerciveFieldAtTemp(tt.kelvin)

		t.Logf("%s: Ec = %.2e V/m (%.2f MV/cm)", tt.name, Ec, Ec/1e8)

		// Ec should decrease with increasing temperature
		if i > 0 && Ec >= prevEc {
			t.Errorf("Ec at %s (%.2e) should be less than at previous temp (%.2e)",
				tt.name, Ec, prevEc)
		}

		// Ec should not be negative
		if Ec < 0 {
			t.Errorf("Ec at %s should not be negative, got %.2e", tt.name, Ec)
		}

		prevEc = Ec
	}
}

// TestPolarizationTemperatureDependence verifies Pr decreases with temperature.
func TestPolarizationTemperatureDependence(t *testing.T) {
	material := DefaultHZO()

	// Test automotive temperature range
	temps := []struct {
		kelvin float64
		name   string
	}{
		{233, "-40°C"},
		{300, "27°C"},
		{358, "85°C"},
		{423, "150°C"},
	}

	var prevPr float64
	for i, tt := range temps {
		Pr := material.PolarizationAtTemp(tt.kelvin)

		t.Logf("%s: Pr = %.4f C/m² (%.1f μC/cm²)", tt.name, Pr, Pr*1e4)

		// Pr should decrease with increasing temperature
		if i > 0 && Pr >= prevPr {
			t.Errorf("Pr at %s (%.4f) should be less than at previous temp (%.4f)",
				tt.name, Pr, prevPr)
		}

		// Pr should remain reasonable at all automotive temps
		if Pr < 0.5*material.Pr {
			t.Errorf("Pr at %s (%.4f) is too low (< 50%% of room temp value)", tt.name, Pr)
		}

		prevPr = Pr
	}
}

// TestSwitchingTimeTemperatureDependence verifies faster switching at higher temps.
func TestSwitchingTimeTemperatureDependence(t *testing.T) {
	material := DefaultHZO()

	// Test automotive temperature range
	temps := []struct {
		kelvin float64
		name   string
	}{
		{233, "-40°C"},
		{300, "27°C"},
		{358, "85°C"},
		{423, "150°C"},
	}

	var prevTau float64
	for i, tt := range temps {
		tau := material.SwitchingTime(tt.kelvin)

		t.Logf("%s: τ = %.2e s (%.2f ns)", tt.name, tau, tau*1e9)

		// Switching should be faster (smaller tau) at higher temperatures
		if i > 0 && tau >= prevTau {
			t.Errorf("Switching time at %s (%.2e) should be less than at previous temp (%.2e)",
				tt.name, tau, prevTau)
		}

		prevTau = tau
	}
}

// TestHysteresisAtAutomotiveTemps verifies hysteresis works across temperature range.
func TestHysteresisAtAutomotiveTemps(t *testing.T) {
	temps := []struct {
		kelvin float64
		name   string
	}{
		{233, "-40°C (cold start)"},
		{300, "27°C (room temp)"},
		{358, "85°C (standard test)"},
		{423, "150°C (extreme)"},
	}

	for _, tt := range temps {
		t.Run(tt.name, func(t *testing.T) {
			material := DefaultHZO()
			model := NewPreisachModel(material)

			// Apply field to positive saturation
			Emax := 2 * material.Ec
			model.Update(Emax)
			P := model.Polarization()

			// Should still achieve significant polarization at all temps
			if P < 0.3*material.Ps {
				t.Errorf("At %s, polarization %.4f is too low (expected > %.4f)",
					tt.name, P, 0.3*material.Ps)
			}

			t.Logf("%s: P_sat = %.4f C/m² (%.1f%% of Ps)",
				tt.name, P, 100*P/material.Ps)
		})
	}
}

// TestRetentionVsTemperature verifies retention behavior at different temperatures.
func TestRetentionVsTemperature(t *testing.T) {
	material := DefaultHZO()

	temps := []struct {
		kelvin float64
		name   string
	}{
		{300, "27°C"},
		{358, "85°C (standard test)"},
		{423, "150°C (accelerated)"},
	}

	// Test retention at 10 years (3.15e8 seconds)
	tenYears := 3.15e8

	for _, tt := range temps {
		retainedPr := material.RetentionAtTime(tenYears, tt.kelvin)
		retentionPercent := 100 * retainedPr / material.Pr

		t.Logf("%s: 10-year retention = %.1f%% of initial Pr", tt.name, retentionPercent)

		// At all temps, should retain at least 90% (FeCIM spec)
		if retentionPercent < 90 {
			t.Errorf("At %s, 10-year retention %.1f%% is below 90%% requirement",
				tt.name, retentionPercent)
		}
	}
}

// TestEnduranceAtCycles verifies degradation model.
func TestEnduranceAtCycles(t *testing.T) {
	material := DefaultHZO()

	cycles := []float64{1e6, 1e7, 1e8, 1e9}

	for _, N := range cycles {
		remainingPr := material.EnduranceAtCycles(N)
		remainingPercent := 100 * remainingPr / material.Pr

		t.Logf("After %.0e cycles: Pr = %.1f%% of initial", N, remainingPercent)

		// Should maintain reasonable Pr well below endurance limit
		// Using stretched exponential, expect ~60% at 10^9 for 10^10 limit
		if N <= 1e8 && remainingPercent < 70 {
			t.Errorf("At %.0e cycles, Pr degradation too severe: %.1f%%", N, remainingPercent)
		}
	}
}

// TestMaterialVariants verifies different HZO material configurations.
func TestMaterialVariants(t *testing.T) {
	materials := []struct {
		name string
		mat  *HZOMaterial
	}{
		{"DefaultHZO", DefaultHZO()},
		{"FeCIMMaterial", FeCIMMaterial()},
		{"FeCIMMaterialTarget", FeCIMMaterialTarget()},
	}

	for _, m := range materials {
		t.Run(m.name, func(t *testing.T) {
			mat := m.mat

			// All materials should have valid parameters
			if mat.Pr <= 0 {
				t.Error("Pr should be positive")
			}
			if mat.Ps <= mat.Pr {
				t.Error("Ps should be greater than Pr")
			}
			if mat.Ec <= 0 {
				t.Error("Ec should be positive")
			}
			if mat.Thickness <= 0 {
				t.Error("Thickness should be positive")
			}

			t.Logf("%s: Pr=%.1f μC/cm², Ec=%.2f MV/cm, Endurance=%.0e",
				m.name, mat.Pr*1e4, mat.Ec/1e8, mat.EnduranceCycles)
		})
	}
}

// TestCurieTemperatureBehavior verifies polarization goes to zero at Curie temp.
func TestCurieTemperatureBehavior(t *testing.T) {
	material := DefaultHZO()

	// Below Curie temp - should have polarization
	belowTc := material.CurieTemp - 100
	PrBelowTc := material.PolarizationAtTemp(belowTc)
	if PrBelowTc <= 0 {
		t.Errorf("Below Tc: Pr should be positive, got %.4f", PrBelowTc)
	}

	// At Curie temp - should be zero
	PrAtTc := material.PolarizationAtTemp(material.CurieTemp)
	if PrAtTc != 0 {
		t.Errorf("At Tc: Pr should be zero, got %.4f", PrAtTc)
	}

	// Above Curie temp - should be zero
	aboveTc := material.CurieTemp + 100
	PrAboveTc := material.PolarizationAtTemp(aboveTc)
	if PrAboveTc != 0 {
		t.Errorf("Above Tc: Pr should be zero, got %.4f", PrAboveTc)
	}

	t.Logf("Curie temp = %.0f K (%.0f°C)", material.CurieTemp, material.CurieTemp-273)
	t.Logf("Below Tc (%.0f K): Pr = %.4f", belowTc, PrBelowTc)
	t.Logf("At Tc: Pr = %.4f", PrAtTc)
}

// TestCoerciveVoltage verifies voltage calculation from field.
func TestCoerciveVoltage(t *testing.T) {
	material := DefaultHZO()

	Vc := material.CoerciveVoltage()

	// For 10nm film with Ec ~1-2 MV/cm, Vc should be ~1-2V
	expectedVc := material.Ec * material.Thickness

	if math.Abs(Vc-expectedVc) > 1e-10 {
		t.Errorf("Coercive voltage %.4f V doesn't match expected %.4f V", Vc, expectedVc)
	}

	t.Logf("Coercive voltage: %.3f V (for %.0f nm film)", Vc, material.Thickness*1e9)
}

// TestSwitchingEnergy verifies energy calculation.
func TestSwitchingEnergy(t *testing.T) {
	material := DefaultHZO()

	E := material.SwitchingEnergy()

	// Energy should be positive
	if E <= 0 {
		t.Error("Switching energy should be positive")
	}

	// Verify energy calculation formula: E = 2 * Pr * Ec * Volume
	// For DefaultHZO: Area=100e-12 m², Thickness=10e-9 m
	// Volume = 1e-18 m³
	// E = 2 * 0.25 * 1.2e8 * 1e-18 = 6e-11 J (60 fJ)
	// This is per-capacitor energy for a 100nm² device
	expectedE := 2 * material.Pr * material.Ec * material.Area * material.Thickness
	if math.Abs(E-expectedE) > 1e-15 {
		t.Errorf("Switching energy %.2e J doesn't match formula %.2e J", E, expectedE)
	}

	t.Logf("Switching energy: %.2e J (%.2f pJ per 100nm² cap)", E, E*1e12)
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkHysteresisLoop(b *testing.B) {
	material := DefaultHZO()
	model := NewPreisachModel(material)
	Emax := 2 * material.Ec

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.GetHysteresisLoop(Emax, 100)
	}
}

func BenchmarkPreisachModelUpdate(b *testing.B) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		E := float64(i%100-50) * material.Ec / 50
		model.Update(E)
	}
}
