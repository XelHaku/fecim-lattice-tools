package physics

import (
	"fmt"
	"math"
)

// WakeUpModelConfig captures wake-up then fatigue behavior of remanent polarization.
type WakeUpModelConfig struct {
	PrInitial_Cm2      float64
	WakeUpGainFraction float64 // Fractional gain above initial Pr at full wake-up
	WakeUpTauCycles    float64 // Characteristic wake-up cycle count
	FatigueOnsetCycles float64 // Fatigue starts after this cycle count
	FatigueTauCycles   float64 // Characteristic fatigue decay after onset
}

// DefaultWakeUpModelConfig returns a WakeUpModelConfig with literature-derived
// defaults for 10 nm HZO. The wakeup parameters are motivated by the phase
// transition from tetragonal to ferroelectric orthorhombic during early cycling.
//
// References:
//   - Zhou et al., ACS Applied Materials & Interfaces 12, 2020
//   - Pesic et al., Advanced Functional Materials 26, 2016
//   - PMC11789571 (2025): Pr rises from ~12 to ~28 uC/cm^2 during wake-up
func DefaultWakeUpModelConfig(prInitial float64) WakeUpModelConfig {
	return WakeUpModelConfig{
		PrInitial_Cm2:      prInitial,
		WakeUpGainFraction: 0.30, // 30% increase during wake-up (conservative vs 2.3x observed)
		WakeUpTauCycles:    200,  // Characteristic wake-up time constant
		FatigueOnsetCycles: 1e4,  // Fatigue onset after ~10^4 cycles
		FatigueTauCycles:   5e5,  // Gradual fatigue decay
	}
}

// WakeUpPolarization returns Pr at the given cycle count.
func WakeUpPolarization(cycles float64, cfg WakeUpModelConfig) (float64, error) {
	if cycles < 0 {
		return 0, fmt.Errorf("cycles must be non-negative, got %g", cycles)
	}
	if cfg.PrInitial_Cm2 <= 0 || cfg.WakeUpTauCycles <= 0 || cfg.FatigueTauCycles <= 0 {
		return 0, fmt.Errorf("invalid config: PrInitial=%g wakeTau=%g fatigueTau=%g", cfg.PrInitial_Cm2, cfg.WakeUpTauCycles, cfg.FatigueTauCycles)
	}
	wake := 1.0 + cfg.WakeUpGainFraction*(1.0-math.Exp(-cycles/cfg.WakeUpTauCycles))
	fatigue := 1.0
	if cycles > cfg.FatigueOnsetCycles {
		fatigue = math.Exp(-(cycles - cfg.FatigueOnsetCycles) / cfg.FatigueTauCycles)
	}
	return cfg.PrInitial_Cm2 * wake * fatigue, nil
}

// NonMonotonicEnduranceConfig parameterizes the three-phase Pr(N) trajectory
// observed in HZO ferroelectrics: wake-up rise, stable plateau, fatigue decline.
//
// The model uses the product form recommended by the physics realism audit:
//
//	Pr(N) = Pr0 * wakeup(N) * fatigue(N)
//	wakeup(N) = [1 - exp(-N / N_wakeup)]^alpha_w       (rising phase)
//	fatigue(N) = 1 - beta_f * (N / N_fatigue)^gamma_f   (declining phase, clamped >= 0)
//
// This captures:
//   - Phase 1 (N < ~N_wakeup): Pr increases as tetragonal -> orthorhombic conversion
//   - Phase 2 (N ~ N_wakeup to ~N_fatigue): Pr at plateau
//   - Phase 3 (N > N_fatigue): Pr decreases as orthorhombic -> monoclinic conversion
//
// References:
//   - Zhou et al., ACS Applied Materials & Interfaces 12, 2020
//   - Pesic et al., Advanced Functional Materials 26, 2016
//   - PMC11789571 (2025): 90-degree UCDW -> 180-degree DW transition
//   - Physics Realism Audit Addendum 2026-02, Section 3.2
type NonMonotonicEnduranceConfig struct {
	Pr0      float64 // Pristine Pr (post-wakeup peak, C/m^2)
	NWakeup  float64 // Wake-up characteristic cycle count (~10^3)
	AlphaW   float64 // Wake-up stretching exponent (0 < alpha_w <= 1)
	NFatigue float64 // Fatigue characteristic cycle count (~10^8)
	BetaF    float64 // Maximum fatigue fractional loss (0 < beta_f < 1)
	GammaF   float64 // Fatigue stretching exponent
}

// DefaultNonMonotonicEnduranceConfig returns literature-derived defaults for
// 10 nm HZO. These are simulation defaults for education, not calibrated to
// a specific device.
//
// The pristine polarization (prPeak) represents the post-wakeup maximum Pr.
// At N=0 the model returns a lower value representing the virgin state before
// wake-up has completed.
func DefaultNonMonotonicEnduranceConfig(prPeak float64) NonMonotonicEnduranceConfig {
	return NonMonotonicEnduranceConfig{
		Pr0:      prPeak,
		NWakeup:  1e3, // Wake-up completes around 10^3 cycles
		AlphaW:   0.3, // Stretched rise; shallow initial slope
		NFatigue: 1e8, // Fatigue onset at ~10^8 cycles
		BetaF:    0.5, // 50% maximum Pr loss at extreme cycling
		GammaF:   0.2, // Gradual fatigue (stretched power law)
	}
}

// NonMonotonicEndurance returns Pr at cycle count N using the three-phase
// wakeup-plateau-fatigue model. Returns an error if the configuration is
// invalid (non-positive parameters).
//
// The returned polarization is always non-negative and follows the
// non-monotonic trajectory observed in HZO endurance experiments.
func NonMonotonicEndurance(N float64, cfg NonMonotonicEnduranceConfig) (float64, error) {
	if N < 0 {
		return 0, fmt.Errorf("cycle count must be non-negative, got %g", N)
	}
	if cfg.Pr0 <= 0 {
		return 0, fmt.Errorf("Pr0 must be positive, got %g", cfg.Pr0)
	}
	if cfg.NWakeup <= 0 {
		return 0, fmt.Errorf("NWakeup must be positive, got %g", cfg.NWakeup)
	}
	if cfg.NFatigue <= 0 {
		return 0, fmt.Errorf("NFatigue must be positive, got %g", cfg.NFatigue)
	}
	if cfg.AlphaW <= 0 {
		return 0, fmt.Errorf("AlphaW must be positive, got %g", cfg.AlphaW)
	}
	if cfg.BetaF < 0 || cfg.BetaF >= 1.0 {
		return 0, fmt.Errorf("BetaF must be in [0, 1), got %g", cfg.BetaF)
	}
	if cfg.GammaF <= 0 {
		return 0, fmt.Errorf("GammaF must be positive, got %g", cfg.GammaF)
	}

	// Wake-up term: rises from 0 toward 1 as N increases.
	// At N=0 this is 0; at N >> NWakeup this approaches 1.
	wakeup := math.Pow(1.0-math.Exp(-N/cfg.NWakeup), cfg.AlphaW)

	// Fatigue term: starts at 1 and decreases toward (1 - BetaF) for large N.
	fatigue := 1.0 - cfg.BetaF*math.Pow(N/cfg.NFatigue, cfg.GammaF)
	if fatigue < 0 {
		fatigue = 0
	}

	return cfg.Pr0 * wakeup * fatigue, nil
}
