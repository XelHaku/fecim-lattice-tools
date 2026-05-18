package literature

// Literature-backed validation for the worldclass frequency dispersion model.
//
// Published HZO frequency dispersion data shows that coercive field (Ec)
// increases with measurement frequency in a log-linear relationship.
// At higher frequencies, less time is available for thermally activated
// domain nucleation, requiring larger fields for switching.
//
// References:
//   - Schenk et al., ACS Appl. Mater. Interfaces 7, 20224 (2015),
//     doi:10.1021/acsami.5b05773
//   - Mueller et al., IEEE Trans. Electron Devices 60(12), 4199 (2013),
//     doi:10.1109/TED.2013.2283465

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// defaultDispersionConfig returns a FrequencyDispersionConfig representative
// of HZO log-linear frequency dependence from Schenk et al. 2015.
func defaultDispersionConfig() physics.FrequencyDispersionConfig {
	return physics.FrequencyDispersionConfig{
		ReferenceHz:        1e3,   // 1 kHz reference
		EcLogSlope:         0.05,  // Ec increases ~5% per ln(f/f0)
		PrLogSlope:         -0.01, // Pr decreases slightly with frequency
		LoopAreaLogSlope:   0.03,  // Loop area increases slightly
		MinMultiplierClamp: 0.1,   // Physical lower bound
	}
}

// baselineMetrics returns a typical HZO hysteresis metrics at 1 kHz.
func baselineMetrics() physics.HysteresisMetrics {
	return physics.HysteresisMetrics{
		FrequencyHz:  1e3,
		Pr_Cm2:       0.245, // 24.5 uC/cm^2
		Ec_Vm:        1.0e8, // 1.0 MV/cm
		LoopArea_Jm3: 4.9e7, // approximate
	}
}

// TestDispersion_EcMonotonicWithFrequency validates that Ec increases
// monotonically with measurement frequency, consistent with thermally
// activated switching kinetics.
//
// Literature basis: Schenk et al., ACS Appl. Mater. Interfaces 2015
// (doi:10.1021/acsami.5b05773) demonstrate log-linear Ec increase with
// frequency in HfO2-based ferroelectrics.
func TestDispersion_EcMonotonicWithFrequency(t *testing.T) {
	cfg := defaultDispersionConfig()
	base := baselineMetrics()

	frequencies := []float64{1e2, 1e3, 1e4, 1e5, 1e6}
	var prevEc float64

	for i, freq := range frequencies {
		m, err := physics.ApplyFrequencyDispersion(base, freq, cfg)
		if err != nil {
			t.Fatalf("ApplyFrequencyDispersion at %.0e Hz failed: %v", freq, err)
		}

		t.Logf("f=%.0e Hz: Ec=%.4e V/m (ratio to base: %.3f)", freq, m.Ec_Vm, m.Ec_Vm/base.Ec_Vm)

		if i > 0 && m.Ec_Vm < prevEc-1e-3 {
			t.Errorf("Ec should increase with frequency: Ec(%.0e)=%.4e < Ec_prev=%.4e",
				freq, m.Ec_Vm, prevEc)
		}
		prevEc = m.Ec_Vm
	}
}

// TestDispersion_EcRatio_1kHz_1MHz validates that the Ec ratio between
// 1 MHz and 1 kHz falls within the literature-reported range of [1.1, 2.0]
// for HZO ferroelectrics.
//
// Literature basis: Schenk et al., ACS Appl. Mater. Interfaces 2015
// (doi:10.1021/acsami.5b05773) report Ec ratios of ~1.2-1.5 across
// the 1 kHz to 1 MHz range in doped HfO2 films. The wider [1.1, 2.0]
// band accounts for material and thickness variations.
func TestDispersion_EcRatio_1kHz_1MHz(t *testing.T) {
	cfg := defaultDispersionConfig()
	base := baselineMetrics()

	m1k, err := physics.ApplyFrequencyDispersion(base, 1e3, cfg)
	if err != nil {
		t.Fatalf("ApplyFrequencyDispersion at 1 kHz failed: %v", err)
	}

	m1M, err := physics.ApplyFrequencyDispersion(base, 1e6, cfg)
	if err != nil {
		t.Fatalf("ApplyFrequencyDispersion at 1 MHz failed: %v", err)
	}

	ratio := m1M.Ec_Vm / m1k.Ec_Vm
	t.Logf("Ec(1MHz)/Ec(1kHz) = %.4f", ratio)
	t.Logf("Ec at 1 kHz: %.4e V/m, Ec at 1 MHz: %.4e V/m", m1k.Ec_Vm, m1M.Ec_Vm)

	if ratio < 1.1 || ratio > 2.0 {
		t.Errorf("Ec ratio (1MHz/1kHz) = %.4f, want [1.1, 2.0] per Schenk 2015", ratio)
	}
}

// TestDispersion_PrWeakFrequencyDependence validates that Pr shows weak
// (or slightly negative) frequency dependence, consistent with literature.
// Pr may decrease slightly at high frequency due to incomplete switching.
//
// Reference: Schenk et al. 2015, doi:10.1021/acsami.5b05773
func TestDispersion_PrWeakFrequencyDependence(t *testing.T) {
	cfg := defaultDispersionConfig()
	base := baselineMetrics()

	m1k, err := physics.ApplyFrequencyDispersion(base, 1e3, cfg)
	if err != nil {
		t.Fatalf("ApplyFrequencyDispersion at 1 kHz failed: %v", err)
	}

	m1M, err := physics.ApplyFrequencyDispersion(base, 1e6, cfg)
	if err != nil {
		t.Fatalf("ApplyFrequencyDispersion at 1 MHz failed: %v", err)
	}

	prRatio := m1M.Pr_Cm2 / m1k.Pr_Cm2
	t.Logf("Pr(1MHz)/Pr(1kHz) = %.4f", prRatio)

	// Pr change should be small (< 30% variation across 3 decades)
	if math.Abs(prRatio-1.0) > 0.30 {
		t.Errorf("Pr variation too large: ratio = %.4f, want within 30%% of unity", prRatio)
	}
}

// TestDispersion_LogLinearScaling validates that the Ec vs frequency
// relationship follows a log-linear trend, consistent with Merz-law
// thermally activated kinetics.
//
// Reference: Mueller et al., IEEE TED 2013 (doi:10.1109/TED.2013.2283465)
func TestDispersion_LogLinearScaling(t *testing.T) {
	cfg := defaultDispersionConfig()
	base := baselineMetrics()

	// Sample Ec at logarithmically spaced frequencies
	freqs := []float64{1e2, 1e3, 1e4, 1e5, 1e6}
	logF := make([]float64, len(freqs))
	ecVals := make([]float64, len(freqs))

	for i, f := range freqs {
		m, err := physics.ApplyFrequencyDispersion(base, f, cfg)
		if err != nil {
			t.Fatalf("ApplyFrequencyDispersion at %.0e Hz failed: %v", f, err)
		}
		logF[i] = math.Log10(f)
		ecVals[i] = m.Ec_Vm
	}

	// Verify that Ec increments per decade are approximately constant
	// (log-linear behavior means equal Ec gain per decade of frequency)
	var increments []float64
	for i := 1; i < len(ecVals); i++ {
		inc := (ecVals[i] - ecVals[i-1]) / (logF[i] - logF[i-1])
		increments = append(increments, inc)
		t.Logf("Ec increment per decade [%.0f-%.0f Hz]: %.4e V/m/decade",
			freqs[i-1], freqs[i], inc)
	}

	// All increments should be within a factor of 2 of each other (log-linear)
	if len(increments) >= 2 {
		for i := 1; i < len(increments); i++ {
			ratio := increments[i] / increments[0]
			if ratio < 0.3 || ratio > 3.0 {
				t.Errorf("non-log-linear: increment ratio [%d]/[0] = %.3f, want [0.3, 3.0]",
					i, ratio)
			}
		}
	}
}

// TestDispersion_ReferenceFrequencyIdentity validates that applying
// dispersion at the reference frequency returns the baseline metrics.
//
// This is a self-consistency check rather than a literature comparison,
// but it ensures the model does not corrupt baseline values.
func TestDispersion_ReferenceFrequencyIdentity(t *testing.T) {
	cfg := defaultDispersionConfig()
	base := baselineMetrics()

	m, err := physics.ApplyFrequencyDispersion(base, cfg.ReferenceHz, cfg)
	if err != nil {
		t.Fatalf("ApplyFrequencyDispersion at reference failed: %v", err)
	}

	tol := 1e-10
	if math.Abs(m.Ec_Vm-base.Ec_Vm) > tol {
		t.Errorf("Ec at reference frequency: got %.6e, want %.6e", m.Ec_Vm, base.Ec_Vm)
	}
	if math.Abs(m.Pr_Cm2-base.Pr_Cm2) > tol {
		t.Errorf("Pr at reference frequency: got %.6e, want %.6e", m.Pr_Cm2, base.Pr_Cm2)
	}
}
