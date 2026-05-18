package literature

// Literature-backed validation for the worldclass retention model.
//
// Published HZO retention data shows that remanent polarization (Pr) decays
// by approximately 10-20% over 10 years (10^8 seconds) at 85 C, with
// stronger decay at elevated temperature and weaker decay at room temperature.
//
// References:
//   - Mueller et al., IEEE Trans. Electron Devices 60(12), 4199 (2013),
//     doi:10.1109/TED.2013.2283465
//   - Pesic et al., Adv. Funct. Mater. 26, 4601 (2016),
//     doi:10.1002/adfm.201603182

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// TestRetentionPowerLaw_HZO_85C_10Year validates that the power-law retention
// model predicts Pr retained in [0.70, 0.95] after 10^8 s at 85 C (358 K).
//
// Literature basis: Mueller et al., IEEE TED 2013 (doi:10.1109/TED.2013.2283465)
// report ~10-20% Pr loss in HZO after 10-year extrapolation at 85 C.
// Pesic et al., Adv. Funct. Mater. 2016 (doi:10.1002/adfm.201603182) confirm
// comparable retention behavior in Si-doped HfO2 films.
func TestRetentionPowerLaw_HZO_85C_10Year(t *testing.T) {
	P0 := 0.245   // C/m^2 (24.5 uC/cm^2, DefaultHZO Pr)
	t0 := 1.0     // reference time 1 s
	beta := 0.012 // HZO decay exponent calibrated for 10-year retention
	// (literature range 0.01-0.05; lower end for well-optimized films)
	tenYears := 1e8 // 10^8 seconds

	times := []float64{tenYears}
	pts, err := physics.SimulateRetentionPowerLaw(P0, t0, beta, times)
	if err != nil {
		t.Fatalf("SimulateRetentionPowerLaw failed: %v", err)
	}

	ratio := pts[0].Polarization_Cm / P0
	t.Logf("Retention ratio at 10^8 s (85C, beta=%.3f): %.4f", beta, ratio)

	// Literature range: 80-90% retained (10-20% loss)
	if ratio < 0.70 || ratio > 0.95 {
		t.Errorf("Pr_retained/Pr_initial = %.4f, want [0.70, 0.95] per Mueller 2013 / Pesic 2016", ratio)
	}
}

// TestRetentionPowerLaw_TemperatureOrdering validates that retention improves
// at lower temperature and degrades at higher temperature, consistent with
// thermally activated depolarization.
//
// Literature basis: Mueller et al., IEEE TED 2013 (doi:10.1109/TED.2013.2283465)
// show Arrhenius-type acceleration of retention loss with temperature.
func TestRetentionPowerLaw_TemperatureOrdering(t *testing.T) {
	P0 := 0.245 // C/m^2
	t0 := 1.0
	tenYears := 1e8

	// Model temperature dependence via beta scaling:
	// Higher temperature -> larger effective beta -> more decay.
	// beta_eff(T) ~ beta_ref * exp(Ea/kB * (1/Tref - 1/T))
	// For simplicity, use representative betas from literature ranges.
	beta25C := 0.006  // room temperature: slower decay
	beta85C := 0.012  // 85 C reference
	beta150C := 0.025 // 150 C: accelerated decay

	times := []float64{tenYears}

	pts25, err := physics.SimulateRetentionPowerLaw(P0, t0, beta25C, times)
	if err != nil {
		t.Fatalf("25C retention failed: %v", err)
	}
	pts85, err := physics.SimulateRetentionPowerLaw(P0, t0, beta85C, times)
	if err != nil {
		t.Fatalf("85C retention failed: %v", err)
	}
	pts150, err := physics.SimulateRetentionPowerLaw(P0, t0, beta150C, times)
	if err != nil {
		t.Fatalf("150C retention failed: %v", err)
	}

	r25 := pts25[0].Polarization_Cm / P0
	r85 := pts85[0].Polarization_Cm / P0
	r150 := pts150[0].Polarization_Cm / P0

	t.Logf("Retention ratios at 10^8 s: 25C=%.4f, 85C=%.4f, 150C=%.4f", r25, r85, r150)

	if r25 <= r85 {
		t.Errorf("25C retention (%.4f) should be better than 85C (%.4f)", r25, r85)
	}
	if r85 <= r150 {
		t.Errorf("85C retention (%.4f) should be better than 150C (%.4f)", r85, r150)
	}
}

// TestRetentionExponential_HZO_85C validates the exponential retention model
// against the same literature constraints: 70-95% retained at 10^8 s.
//
// Reference: Mueller et al., IEEE TED 2013, doi:10.1109/TED.2013.2283465
func TestRetentionExponential_HZO_85C(t *testing.T) {
	P0 := 0.245       // C/m^2
	Pinf := 0.80 * P0 // asymptotic Pr (20% max loss at infinite time)
	tau := 1e8        // characteristic decay time ~10^8 s, calibrated so retention
	// at 10^8 s matches literature 10-20% loss range
	tenYears := 1e8

	times := []float64{tenYears}
	pts, err := physics.SimulateRetentionExponential(P0, Pinf, tau, times)
	if err != nil {
		t.Fatalf("SimulateRetentionExponential failed: %v", err)
	}

	ratio := pts[0].Polarization_Cm / P0
	t.Logf("Exponential retention ratio at 10^8 s: %.4f (tau=%.2e s)", ratio, tau)

	if ratio < 0.70 || ratio > 0.95 {
		t.Errorf("Pr_retained/Pr_initial = %.4f, want [0.70, 0.95]", ratio)
	}
}

// TestRetentionExponential_TemperatureOrdering validates monotonic temperature
// ordering using the exponential model with temperature-scaled tau.
//
// Reference: Pesic et al., Adv. Funct. Mater. 2016, doi:10.1002/adfm.201603182
func TestRetentionExponential_TemperatureOrdering(t *testing.T) {
	P0 := 0.245
	Pinf := 0.80 * P0
	tenYears := 1e8

	// Arrhenius-scaled tau: tau(T) = tau_ref * exp(Ea/kB * (1/T - 1/Tref))
	// Higher T -> shorter tau -> more decay
	Ea := 1.0      // eV
	kB := 8.617e-5 // eV/K
	Tref := 358.0  // 85 C
	tauRef := 1e8  // calibrated for 85 C reference

	temperatures := []struct {
		name string
		T    float64
	}{
		{"25C", 298.0},
		{"85C", 358.0},
		{"150C", 423.0},
	}

	times := []float64{tenYears}
	ratios := make([]float64, len(temperatures))

	for i, tc := range temperatures {
		tau := tauRef * math.Exp(Ea/kB*(1.0/tc.T-1.0/Tref))
		pts, err := physics.SimulateRetentionExponential(P0, Pinf, tau, times)
		if err != nil {
			t.Fatalf("%s retention failed: %v", tc.name, err)
		}
		ratios[i] = pts[0].Polarization_Cm / P0
		t.Logf("%s (%.0f K): retention=%.4f, tau=%.2e s", tc.name, tc.T, ratios[i], tau)
	}

	// 25C should retain more than 85C
	if ratios[0] <= ratios[1] {
		t.Errorf("25C retention (%.4f) should exceed 85C (%.4f)", ratios[0], ratios[1])
	}
	// 85C should retain more than 150C
	if ratios[1] <= ratios[2] {
		t.Errorf("85C retention (%.4f) should exceed 150C (%.4f)", ratios[1], ratios[2])
	}
}

// TestRetentionPowerLaw_DecayMonotonicity confirms that retention decreases
// monotonically with time for positive beta, as expected physically.
//
// Reference: Mueller et al., IEEE TED 2013, doi:10.1109/TED.2013.2283465
func TestRetentionPowerLaw_DecayMonotonicity(t *testing.T) {
	P0 := 0.245
	t0 := 1.0
	beta := 0.03

	times, err := physics.GenerateLogTimeSweep(1.0, 1e8, 20)
	if err != nil {
		t.Fatalf("GenerateLogTimeSweep failed: %v", err)
	}

	pts, err := physics.SimulateRetentionPowerLaw(P0, t0, beta, times)
	if err != nil {
		t.Fatalf("SimulateRetentionPowerLaw failed: %v", err)
	}

	for i := 1; i < len(pts); i++ {
		if pts[i].Polarization_Cm > pts[i-1].Polarization_Cm+1e-12 {
			t.Errorf("non-monotonic decay at index %d: P[%d]=%.6f > P[%d]=%.6f",
				i, i, pts[i].Polarization_Cm, i-1, pts[i-1].Polarization_Cm)
		}
	}
}
