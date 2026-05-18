package literature

// Literature-backed validation for the worldclass wake-up model.
//
// Published HZO wake-up data shows that remanent polarization (Pr) increases
// by 20-50% during the first 10^3 to 10^5 bipolar cycling pulses, then
// saturates or begins to fatigue.
//
// References:
//   - Pesic et al., Adv. Funct. Mater. 26, 4601 (2016),
//     doi:10.1002/adfm.201603182
//   - Park et al., Adv. Mater. 27, 1811 (2015),
//     doi:10.1002/adma.201404531

import (
	"math"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

// defaultWakeUpConfig returns a WakeUpModelConfig with parameters representative
// of HZO wake-up behavior from Pesic et al. 2016 (doi:10.1002/adfm.201603182).
func defaultWakeUpConfig() physics.WakeUpModelConfig {
	return physics.WakeUpModelConfig{
		PrInitial_Cm2:      0.15, // 15 uC/cm^2 initial (pre-wake-up, lower end)
		WakeUpGainFraction: 0.40, // 40% Pr gain during wake-up (literature: 20-50%)
		WakeUpTauCycles:    1e3,  // characteristic wake-up at ~10^3 cycles
		FatigueOnsetCycles: 1e6,  // fatigue begins after ~10^6 cycles
		FatigueTauCycles:   1e7,  // fatigue decay time constant
	}
}

// TestWakeUp_PrIncrease_10k validates that Pr increases after 10^4 cycles
// of bipolar cycling, consistent with HZO wake-up literature.
//
// Literature basis: Pesic et al., Adv. Funct. Mater. 2016
// (doi:10.1002/adfm.201603182) demonstrate Pr increase from ~15 to ~25 uC/cm^2
// during 10^4 cycles of bipolar cycling in Si-doped HfO2.
func TestWakeUp_PrIncrease_10k(t *testing.T) {
	cfg := defaultWakeUpConfig()

	pr0, err := physics.WakeUpPolarization(0, cfg)
	if err != nil {
		t.Fatalf("WakeUpPolarization at cycle 0 failed: %v", err)
	}

	pr10k, err := physics.WakeUpPolarization(1e4, cfg)
	if err != nil {
		t.Fatalf("WakeUpPolarization at 10^4 cycles failed: %v", err)
	}

	t.Logf("Pr at cycle 0: %.4f C/m^2, Pr at 10^4 cycles: %.4f C/m^2", pr0, pr10k)
	t.Logf("Wake-up gain: %.1f%%", (pr10k-pr0)/pr0*100)

	if pr10k <= pr0 {
		t.Errorf("Pr should increase after 10^4 cycles: Pr(0)=%.4f, Pr(10^4)=%.4f", pr0, pr10k)
	}
}

// TestWakeUp_PlateauBehavior validates that Pr saturates (plateaus) and does
// not increase indefinitely, consistent with domain-pinning saturation.
//
// Literature basis: Pesic et al. 2016 (doi:10.1002/adfm.201603182) show
// that wake-up saturates after ~10^4-10^5 cycles before fatigue onset.
func TestWakeUp_PlateauBehavior(t *testing.T) {
	cfg := defaultWakeUpConfig()

	// Sample at 10^4, 10^5, and 5*10^5 (all before fatigue onset at 10^6)
	cycles := []float64{1e4, 1e5, 5e5}
	prs := make([]float64, len(cycles))

	for i, c := range cycles {
		pr, err := physics.WakeUpPolarization(c, cfg)
		if err != nil {
			t.Fatalf("WakeUpPolarization at %.0e cycles failed: %v", c, err)
		}
		prs[i] = pr
		t.Logf("Pr at %.0e cycles: %.4f C/m^2", c, pr)
	}

	// Pr at 10^5 and 5*10^5 should be very close (plateau), both above 10^4
	relativeDiff := math.Abs(prs[2]-prs[1]) / prs[1]
	t.Logf("Relative Pr change between 10^5 and 5*10^5 cycles: %.4f", relativeDiff)

	if relativeDiff > 0.05 {
		t.Errorf("Pr should plateau (< 5%% change): Pr(10^5)=%.4f, Pr(5*10^5)=%.4f, diff=%.2f%%",
			prs[1], prs[2], relativeDiff*100)
	}
}

// TestWakeUp_GainMagnitude validates that the wake-up Pr gain falls within
// the literature-reported range of 10-80% of the initial Pr.
//
// Literature basis:
//   - Pesic et al. 2016 (doi:10.1002/adfm.201603182): ~40-60% gain
//   - Park et al. 2015 (doi:10.1002/adma.201404531): ~20-50% gain
//   - Broader literature: 10-80% range depending on film quality/composition
func TestWakeUp_GainMagnitude(t *testing.T) {
	cfg := defaultWakeUpConfig()

	pr0, err := physics.WakeUpPolarization(0, cfg)
	if err != nil {
		t.Fatalf("WakeUpPolarization at cycle 0 failed: %v", err)
	}

	// Measure at peak wake-up (before fatigue, well past tau)
	// At ~10*tau = 10^4 cycles, wake-up is ~99.99% complete
	prPeak, err := physics.WakeUpPolarization(1e5, cfg)
	if err != nil {
		t.Fatalf("WakeUpPolarization at 10^5 cycles failed: %v", err)
	}

	gainFrac := (prPeak - pr0) / pr0
	t.Logf("Wake-up gain fraction: %.4f (%.1f%%), Pr_initial=%.4f, Pr_peak=%.4f",
		gainFrac, gainFrac*100, pr0, prPeak)

	// Literature-backed range: 10-80% gain
	if gainFrac < 0.10 || gainFrac > 0.80 {
		t.Errorf("Wake-up gain = %.1f%%, want [10%%, 80%%] per Pesic 2016 / Park 2015", gainFrac*100)
	}
}

// TestWakeUp_FatigueDecay validates that Pr eventually decreases after
// fatigue onset, consistent with defect-generation-driven degradation.
//
// Literature basis: Pesic et al. 2016 (doi:10.1002/adfm.201603182) show
// fatigue onset after ~10^6 cycles in Si-doped HfO2 capacitors.
func TestWakeUp_FatigueDecay(t *testing.T) {
	cfg := defaultWakeUpConfig()

	// Pr at fatigue onset
	prOnset, err := physics.WakeUpPolarization(cfg.FatigueOnsetCycles, cfg)
	if err != nil {
		t.Fatalf("WakeUpPolarization at fatigue onset failed: %v", err)
	}

	// Pr well past fatigue onset
	prFatigued, err := physics.WakeUpPolarization(cfg.FatigueOnsetCycles+3*cfg.FatigueTauCycles, cfg)
	if err != nil {
		t.Fatalf("WakeUpPolarization at fatigue failed: %v", err)
	}

	t.Logf("Pr at fatigue onset (%.0e): %.4f C/m^2", cfg.FatigueOnsetCycles, prOnset)
	t.Logf("Pr at fatigue (%.0e): %.4f C/m^2",
		cfg.FatigueOnsetCycles+3*cfg.FatigueTauCycles, prFatigued)

	if prFatigued >= prOnset {
		t.Errorf("Pr should decrease after fatigue onset: Pr(onset)=%.4f, Pr(fatigue)=%.4f",
			prOnset, prFatigued)
	}
}

// TestWakeUp_MonotonicBeforeFatigue validates that Pr increases monotonically
// during the wake-up phase (before fatigue onset).
//
// Reference: Pesic et al., Adv. Funct. Mater. 2016, doi:10.1002/adfm.201603182
func TestWakeUp_MonotonicBeforeFatigue(t *testing.T) {
	cfg := defaultWakeUpConfig()

	// Sample logarithmically from 1 to FatigueOnsetCycles/2
	nPoints := 20
	maxCycles := cfg.FatigueOnsetCycles / 2

	var prevPr float64
	for i := 0; i < nPoints; i++ {
		frac := float64(i) / float64(nPoints-1)
		cycles := math.Pow(10, frac*math.Log10(maxCycles))

		pr, err := physics.WakeUpPolarization(cycles, cfg)
		if err != nil {
			t.Fatalf("WakeUpPolarization at %.1f cycles failed: %v", cycles, err)
		}

		if i > 0 && pr < prevPr-1e-12 {
			t.Errorf("non-monotonic wake-up at %.1f cycles: Pr=%.6f < prev=%.6f",
				cycles, pr, prevPr)
		}
		prevPr = pr
	}
}
