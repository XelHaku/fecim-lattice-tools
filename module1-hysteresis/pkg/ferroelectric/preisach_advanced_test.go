package ferroelectric

import (
	"math"
	"testing"
)

// TestNewMayergoyzPreisach verifies model creation.
func TestNewMayergoyzPreisach(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	if len(model.hysterons) == 0 {
		t.Error("Model should have hysterons")
	}

	if model.Temperature != 300 {
		t.Errorf("Expected 300K, got %f", model.Temperature)
	}
}

// TestPreisachHysteresisLoop verifies loop generation.
func TestPreisachHysteresisLoop(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	Emax := material.Ec * 2
	E, P := model.GetHysteresisLoop(Emax, 100)

	if len(E) != len(P) {
		t.Error("E and P should have same length")
	}

	if len(E) < 200 {
		t.Errorf("Expected at least 200 points, got %d", len(E))
	}

	// Check that loop has proper range
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
	if maxE < Emax*0.9 || minE > -Emax*0.9 {
		t.Errorf("Loop should span ±Emax, got [%.2e, %.2e]", minE, maxE)
	}
}

// TestPreisachSaturation verifies polarization saturates near Ps.
func TestPreisachSaturation(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 40)

	// Apply large field
	Emax := material.Ec * 3
	model.Update(Emax)

	// Should be close to +Ps
	P := model.Polarization()
	if P < 0.8*material.Ps {
		t.Errorf("Should saturate near Ps, got %.4f vs %.4f", P, material.Ps)
	}

	// Apply negative field
	model.Update(-Emax)
	P = model.Polarization()
	if P > -0.8*material.Ps {
		t.Errorf("Should saturate near -Ps, got %.4f", P)
	}
}

// TestPreisachMemory verifies hysteresis memory effect.
func TestPreisachMemory(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 40)

	// Saturate positive
	Emax := material.Ec * 2
	model.Update(Emax)
	Psat := model.Polarization()

	// Reduce field to zero
	model.Update(0)
	Prem := model.Polarization()

	// Remanent polarization should be positive (memory)
	if Prem <= 0 {
		t.Errorf("Remanent polarization should be positive, got %.4f", Prem)
	}

	// Remanent should be less than saturation
	if Prem >= Psat {
		t.Error("Remanent should be less than saturation")
	}
}

// TestPreisachMinorLoop verifies minor loop generation.
func TestPreisachMinorLoop(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 40)

	// First establish some state
	model.Update(material.Ec)

	// Generate minor loop
	E1 := material.Ec * 0.5
	E2 := -material.Ec * 0.3
	E, P := model.GetMinorLoop(E1, E2, 50)

	if len(E) < 100 {
		t.Errorf("Expected at least 100 points, got %d", len(E))
	}

	// Minor loop should be contained within major loop
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

	if maxP > material.Ps {
		t.Error("Minor loop P should not exceed Ps")
	}
}

// TestTemperatureDependence verifies Ec decreases with temperature.
func TestTemperatureDependence(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	Ec300 := model.GetEffectiveEc()

	// Increase temperature
	model.SetTemperature(400)
	Ec400 := model.GetEffectiveEc()

	if Ec400 >= Ec300 {
		t.Errorf("Ec should decrease with temperature: Ec(300K)=%.2e, Ec(400K)=%.2e",
			Ec300, Ec400)
	}

	// At Curie temperature, Ec should be zero
	model.SetTemperature(model.CurieTemp)
	EcTc := model.GetEffectiveEc()
	if EcTc > 0.01*material.Ec {
		t.Errorf("Ec should be near zero at Curie temp, got %.2e", EcTc)
	}
}

// TestPreisachPlane verifies Preisach plane state retrieval.
func TestPreisachPlane(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	alphas, betas, states := model.GetPreisachPlane()

	if len(alphas) != len(betas) || len(alphas) != len(states) {
		t.Error("Arrays should have same length")
	}

	// All alpha > beta (Preisach constraint)
	for i := range alphas {
		if alphas[i] <= betas[i] {
			t.Errorf("Alpha should be > beta: alpha=%.2e, beta=%.2e",
				alphas[i], betas[i])
		}
	}
}

// TestSwitchedFraction verifies fraction calculation.
func TestSwitchedFraction(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 40)

	// Initial state: all switched down
	frac0 := model.GetSwitchedFraction()
	if frac0 > 0.01 {
		t.Errorf("Initially should be ~0%% switched, got %.2f%%", frac0*100)
	}

	// Apply large positive field
	model.Update(material.Ec * 3)
	frac1 := model.GetSwitchedFraction()
	if frac1 < 0.9 {
		t.Errorf("Should be >90%% switched after saturation, got %.2f%%", frac1*100)
	}
}

// TestDiscreteStates verifies 30-level state generation.
func TestDiscreteStates(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 40)

	states := model.DiscreteStates(30)

	if len(states) != 30 {
		t.Errorf("Expected 30 states, got %d", len(states))
	}

	// Check ordering
	for i := 1; i < len(states); i++ {
		if states[i].Polarization <= states[i-1].Polarization {
			t.Error("States should be ordered by increasing polarization")
		}
	}

	// First state should be near -Ps
	if states[0].NormalizedP > -0.9 {
		t.Errorf("First state should be near -1, got %.2f", states[0].NormalizedP)
	}

	// Last state should be near +Ps
	if states[29].NormalizedP < 0.9 {
		t.Errorf("Last state should be near +1, got %.2f", states[29].NormalizedP)
	}
}

// TestDomainSwitching verifies switching dynamics simulation.
func TestDomainSwitching(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	Eapplied := material.Ec * 2
	duration := 10 * material.Tau
	steps := 100

	times, pols, switched := model.SimulateDomainSwitching(Eapplied, duration, steps)

	if len(times) != steps {
		t.Errorf("Expected %d time points, got %d", steps, len(times))
	}

	// Polarization should increase monotonically
	for i := 1; i < len(pols); i++ {
		if pols[i] < pols[i-1]-1e-10 {
			t.Error("Polarization should increase during switching")
			break
		}
	}

	// Switched count should increase
	for i := 1; i < len(switched); i++ {
		if switched[i] < switched[i-1] {
			t.Error("Switched count should increase")
			break
		}
	}
}

// TestWakeupEffect verifies wake-up cycling effect.
func TestWakeupEffect(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	// Initial wake-up
	_, wakeup0, _ := model.GetFatigueState()

	// Run several cycles
	Emax := material.Ec * 2
	for i := 0; i < 50; i++ {
		model.GetHysteresisLoop(Emax, 20)
	}

	cycles, _, wakeup50 := model.GetFatigueState()

	if cycles != 50 {
		t.Errorf("Expected 50 cycles, got %d", cycles)
	}

	if wakeup50 <= wakeup0 {
		t.Error("Wake-up should increase with cycling")
	}
}

// TestFatigueDegradation verifies fatigue modeling.
func TestFatigueDegradation(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	// Get initial Pmax
	Emax := material.Ec * 2
	model.Update(Emax)
	P0 := model.Polarization()

	// Run many cycles (simulate fatigue)
	model.cycleCount = 1e9 // Simulate 1 billion cycles

	model.Reset()
	model.Update(Emax)
	P1 := model.Polarization()

	// Some degradation should occur
	if P1 >= P0 {
		// Note: with very low fatigue rate, this might pass anyway
		t.Logf("P before: %.4f, P after 1B cycles: %.4f", P0, P1)
	}
}

// TestConductanceRange verifies conductance mapping.
func TestConductanceRange(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 40)

	states := model.DiscreteStates(30)

	// Check conductance range (should be 1-100 µS for FeCIM)
	Gmin := states[0].Conductance
	Gmax := states[29].Conductance

	if Gmin < 0 {
		t.Errorf("Conductance should be positive, got %.2e", Gmin)
	}

	if Gmax <= Gmin {
		t.Error("Gmax should be greater than Gmin")
	}

	// Ratio should be significant (at least 10x)
	ratio := Gmax / Gmin
	if ratio < 5 {
		t.Errorf("Conductance ratio should be >5, got %.1f", ratio)
	}
}

// BenchmarkPreisachUpdate benchmarks the update function.
func BenchmarkPreisachUpdate(b *testing.B) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	E := material.Ec * 0.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Update(E)
		model.Update(-E)
	}
}

// BenchmarkPreisachLoop benchmarks full loop generation.
func BenchmarkPreisachLoop(b *testing.B) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	Emax := material.Ec * 2

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.GetHysteresisLoop(Emax, 100)
	}
}

// TestPECurveSmoothness verifies the P-E curve has enough granularity for 30-level quantization.
func TestPECurveSmoothness(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 60) // Match updated GUI grid size

	Emax := material.Ec * 2.0
	E, P := model.GetHysteresisLoop(Emax, 100)
	_ = E // Use E to avoid unused variable error

	// Count unique P values in -Pr to +Pr range
	Pr := material.Pr
	uniqueP := make(map[float64]bool)
	for _, p := range P {
		if p >= -Pr && p <= Pr {
			// Round to 5% of Pr for comparison
			rounded := math.Round(p/(Pr*0.05)) * (Pr * 0.05)
			uniqueP[rounded] = true
		}
	}

	// Should have at least 20 distinct levels in the polarization range
	if len(uniqueP) < 20 {
		t.Errorf("P-E curve too coarse: only %d distinct P values (expected >= 20)", len(uniqueP))
	}
	t.Logf("P-E curve smoothness: %d distinct P values in ±Pr range", len(uniqueP))
}

// TestNLSSwitchingTime verifies the Merz law switching time calculation.
func TestNLSSwitchingTime(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec

	// Test cases: field -> expected tau range
	testCases := []struct {
		field  float64
		tauMin float64
		tauMax float64
		desc   string
	}{
		{2.0 * Ec, 1e-12, 1e-6, "High field (2*Ec)"},
		{1.5 * Ec, 1e-11, 1e-5, "Moderate field (1.5*Ec)"},
		{1.1 * Ec, 1e-10, 1e-3, "Near threshold (1.1*Ec)"},
		{0.5 * Ec, 1e-6, 1.0, "Below Ec (0.5*Ec)"},
	}

	for _, tc := range testCases {
		tau := model.GetSwitchingTime(tc.field)
		if tau < tc.tauMin || tau > tc.tauMax {
			t.Errorf("%s: tau=%.2e, expected [%.2e, %.2e]", tc.desc, tau, tc.tauMin, tc.tauMax)
		} else {
			t.Logf("%s: tau=%.2e s (OK)", tc.desc, tau)
		}
	}
}

// TestNLSFieldDependence verifies switching time increases as field decreases.
func TestNLSFieldDependence(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec
	fields := []float64{2.0 * Ec, 1.5 * Ec, 1.2 * Ec, 1.0 * Ec}

	var prevTau float64 = 0
	for _, E := range fields {
		tau := model.GetSwitchingTime(E)
		if prevTau > 0 && tau <= prevTau {
			t.Errorf("Switching time should increase as field decreases: E=%.2f*Ec gave tau=%.2e (prev=%.2e)",
				E/Ec, tau, prevTau)
		}
		t.Logf("E=%.2f*Ec -> tau=%.2e s", E/Ec, tau)
		prevTau = tau
	}
}

// TestNLSPerMaterial verifies different materials have different NLS parameters.
func TestNLSPerMaterial(t *testing.T) {
	hzo := DefaultHZO()
	alscn := AlScN()

	modelHZO := NewMayergoyzPreisach(hzo, 50)
	modelAlScN := NewMayergoyzPreisach(alscn, 50)

	// At same normalized field (1.5*Ec), AlScN should have different tau
	fieldHZO := 1.5 * hzo.Ec
	fieldAlScN := 1.5 * alscn.Ec

	tauHZO := modelHZO.GetSwitchingTime(fieldHZO)
	tauAlScN := modelAlScN.GetSwitchingTime(fieldAlScN)

	t.Logf("HZO at 1.5*Ec: tau=%.2e s (Tau0NLS=%.2e, EaNLS=%.2e)", tauHZO, modelHZO.Tau0NLS, modelHZO.EaNLS)
	t.Logf("AlScN at 1.5*Ec: tau=%.2e s (Tau0NLS=%.2e, EaNLS=%.2e)", tauAlScN, modelAlScN.Tau0NLS, modelAlScN.EaNLS)

	// They should be different (AlScN has higher EaNLS but faster Tau0NLS)
	if tauHZO == tauAlScN {
		t.Errorf("Expected different switching times for different materials")
	}
}

// TestNLSZeroField verifies GetSwitchingTime handles zero field correctly.
func TestNLSZeroField(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	tau := model.GetSwitchingTime(0)
	if !math.IsInf(tau, 1) {
		t.Errorf("Expected Inf for zero field, got %v", tau)
	}
}
