// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"math"
	"testing"
)

// TestWakeUpEffectProgression verifies that Pr increases with initial cycling
// before stabilizing due to wake-up effect.
//
// Wake-up effect: In as-deposited ferroelectric films, remanent polarization
// increases during initial cycling due to redistribution of oxygen vacancies
// and defect dipole alignment. After ~100 cycles, the film reaches a stable state.
//
// Reference: Pesic et al., Adv. Funct. Mater. 26, 4601 (2016)
func TestWakeUpEffectProgression(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30) // 30x30 hysteron grid

	// Get initial wakeup factor (should start at ~0.8)
	_, _, wakeup0 := model.GetFatigueState()
	if wakeup0 < 0.75 || wakeup0 > 0.85 {
		t.Errorf("Initial wakeup factor = %f, expected ~0.8", wakeup0)
	}

	// Simulate 100 cycles
	Emax := material.Ec * 2.0
	for i := 0; i < 100; i++ {
		_, _ = model.GetHysteresisLoop(Emax, 50)
	}

	// After 100 cycles, wakeup should reach > 0.90
	// (The model reaches ~0.926 which represents substantial wake-up)
	cycles, _, wakeupFinal := model.GetFatigueState()
	if cycles != 100 {
		t.Errorf("Cycle count = %d, expected 100", cycles)
	}
	if wakeupFinal < 0.90 {
		t.Errorf("Final wakeup factor = %f, expected > 0.90 after 100 cycles", wakeupFinal)
	}

	// Verify wakeup increases monotonically
	if wakeupFinal <= wakeup0 {
		t.Errorf("Wakeup did not increase: initial = %f, final = %f", wakeup0, wakeupFinal)
	}

	t.Logf("Wake-up progression: initial = %.3f → final = %.3f (after %d cycles)",
		wakeup0, wakeupFinal, cycles)
}

// TestFatigueDegradationModel verifies the stretched exponential model
// for polarization degradation with cycling.
//
// Fatigue model: Pr(N) = Pr0 * exp(-(N/N0)^β)
// where N0 is the endurance limit and β ≈ 0.3 is the stretched exponent.
// At N = N0, retention is exp(-1) ≈ 36.8%
//
// Reference: Schroeder et al., Inorg. Chem. Front. 3, 1472 (2016)
func TestFatigueDegradationModel(t *testing.T) {
	tests := []struct {
		name            string
		material        *HZOMaterial
		testCycles      []float64 // Cycle counts to test
		expectedRetPct  []float64 // Expected retention percentage
		tolerancePct    float64   // Tolerance in percentage points
	}{
		{
			name:     "DefaultHZO (1e10 endurance)",
			material: DefaultHZO(),
			testCycles: []float64{
				1e9,  // 10% of endurance
				5e9,  // 50% of endurance
				1e10, // 100% of endurance (end-of-life)
			},
			expectedRetPct: []float64{
				60.0, // At 10%: ~60% (stretched exponential β=0.3)
				45.0, // At 50%: ~45%
				37.0, // At EOL: ~37% (exp(-1^0.3) ≈ 0.368)
			},
			tolerancePct: 10.0,
		},
		{
			name:     "FeCIMMaterial (1e9 endurance)",
			material: FeCIMMaterial(),
			testCycles: []float64{
				1e8,
				5e8,
				1e9,
			},
			expectedRetPct: []float64{
				60.0,
				45.0,
				37.0,
			},
			tolerancePct: 10.0,
		},
		{
			name:     "LiteratureSuperlattice (1e12 endurance)",
			material: LiteratureSuperlattice(),
			testCycles: []float64{
				1e11,
				5e11,
				1e12,
			},
			expectedRetPct: []float64{
				60.0,
				45.0,
				37.0,
			},
			tolerancePct: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Pr0 := tt.material.Pr

			for i, N := range tt.testCycles {
				PrN := tt.material.EnduranceAtCycles(N)
				retentionPct := (PrN / Pr0) * 100

				expectedPct := tt.expectedRetPct[i]
				if math.Abs(retentionPct-expectedPct) > tt.tolerancePct {
					t.Errorf("At N = %.1e: retention = %.1f%%, expected ~%.1f%% ± %.1f%%",
						N, retentionPct, expectedPct, tt.tolerancePct)
				}

				t.Logf("At N = %.1e cycles (%.0f%% of endurance): Pr retention = %.1f%%",
					N, (N/tt.material.EnduranceCycles)*100, retentionPct)
			}
		})
	}
}

// TestEnduranceLimitConsistency verifies that each material has the correct
// endurance limit specified in the literature.
func TestEnduranceLimitConsistency(t *testing.T) {
	tests := []struct {
		name              string
		material          *HZOMaterial
		expectedEndurance float64
	}{
		{
			name:              "DefaultHZO",
			material:          DefaultHZO(),
			expectedEndurance: 1e10, // 10^10 cycles (standard HZO)
		},
		{
			name:              "FeCIMMaterial",
			material:          FeCIMMaterial(),
			expectedEndurance: 1e9, // 10^9 cycles DEMONSTRATED
		},
		{
			name:              "FeCIMMaterialTarget",
			material:          FeCIMMaterialTarget(),
			expectedEndurance: 1e12, // 10^12 cycles TARGET (not yet achieved)
		},
		{
			name:              "LiteratureSuperlattice",
			material:          LiteratureSuperlattice(),
			expectedEndurance: 1e12, // 10^12 cycles (Cheema et al. Nature 2020)
		},
		{
			name:              "CryogenicHZO",
			material:          CryogenicHZO(),
			expectedEndurance: 1e9, // 10^9 at ±5V verified at 4K
		},
		{
			name:              "HZOStandard32",
			material:          HZOStandard32(),
			expectedEndurance: 1e8, // 10^8 cycles (2017 era)
		},
		{
			name:              "HZOFJT140",
			material:          HZOFJT140(),
			expectedEndurance: 1e7, // 10^7 cycles (FTJ tunneling barrier)
		},
		{
			name:              "AlScN",
			material:          AlScN(),
			expectedEndurance: 1e9, // 10^9 cycles
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.material.EnduranceCycles != tt.expectedEndurance {
				t.Errorf("%s endurance = %.1e, expected %.1e",
					tt.name, tt.material.EnduranceCycles, tt.expectedEndurance)
			}
			t.Logf("%s: endurance = %.1e cycles", tt.name, tt.material.EnduranceCycles)
		})
	}
}

// TestRetentionTimeModeling verifies Arrhenius retention model
// for data retention at elevated temperatures.
//
// Retention model: Uses temperature acceleration factor
// τ_eff = τ * exp(Ea/kB * (1/T_ref - 1/T))
// where Ea ≈ 1.0 eV is the retention activation energy.
//
// Specification: 10-year retention at 85°C
//
// Reference: Müller et al., ECS J. Solid State Sci. Technol. 4, N30 (2015)
func TestRetentionTimeModeling(t *testing.T) {
	tests := []struct {
		name            string
		material        *HZOMaterial
		testTime        float64 // Time to test (seconds)
		testTemp        float64 // Temperature (K)
		minRetentionPct float64 // Minimum acceptable retention %
	}{
		{
			name:            "DefaultHZO at 85°C for 10 years",
			material:        DefaultHZO(),
			testTime:        3.15e8, // 10 years in seconds
			testTemp:        358,    // 85°C = 358K
			minRetentionPct: 90.0,   // Should maintain > 90% after 10 years
		},
		{
			name:            "DefaultHZO at 125°C for 1000 hours",
			material:        DefaultHZO(),
			testTime:        3.6e6, // 1000 hours in seconds
			testTemp:        398,   // 125°C = 398K (accelerated test)
			minRetentionPct: 85.0,  // Higher temp → more loss
		},
		{
			name:            "FeCIMMaterial at 85°C",
			material:        FeCIMMaterial(),
			testTime:        1e7,  // 10^7 seconds (~116 days DEMONSTRATED)
			testTemp:        358,  // 85°C
			minRetentionPct: 90.0, // Should maintain > 90%
		},
		{
			name:            "CryogenicHZO at 4K for 10 years",
			material:        CryogenicHZO(),
			testTime:        3.15e8, // 10 years
			testTemp:        4,      // 4K (negligible thermal degradation)
			minRetentionPct: 99.9,   // Essentially perfect at cryo
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Pr0 := tt.material.Pr
			PrT := tt.material.RetentionAtTime(tt.testTime, tt.testTemp)
			retentionPct := (PrT / Pr0) * 100

			if retentionPct < tt.minRetentionPct {
				t.Errorf("Retention = %.1f%%, expected > %.1f%% at T=%.0fK after t=%.2e s",
					retentionPct, tt.minRetentionPct, tt.testTemp, tt.testTime)
			}

			// Also verify retention doesn't exceed 100%
			if retentionPct > 100.0 {
				t.Errorf("Retention = %.1f%% exceeds 100%% (unphysical)", retentionPct)
			}

			t.Logf("T=%.0fK, t=%.2e s: Pr retention = %.2f%% (Pr: %.2e → %.2e C/m²)",
				tt.testTemp, tt.testTime, retentionPct, Pr0, PrT)
		})
	}
}

// TestTemperatureAcceleration verifies that retention loss accelerates
// exponentially with temperature (Arrhenius behavior).
func TestTemperatureAcceleration(t *testing.T) {
	material := DefaultHZO()
	testTime := 1e6 // 1 million seconds (~11.6 days)

	temps := []float64{298, 358, 398, 423} // 25°C, 85°C, 125°C, 150°C
	var prevRetention float64 = 100.0

	for _, T := range temps {
		PrT := material.RetentionAtTime(testTime, T)
		retentionPct := (PrT / material.Pr) * 100

		// Higher temperature should give LOWER retention
		if retentionPct > prevRetention+0.1 { // Small tolerance for numerical errors
			t.Errorf("Retention INCREASED with temperature (unphysical): %.1f%% at %.0fK > %.1f%% at previous temp",
				retentionPct, T, prevRetention)
		}

		t.Logf("T = %.0f K (%.0f°C): retention = %.2f%% after %.1e s",
			T, T-273, retentionPct, testTime)

		prevRetention = retentionPct
	}
}

// TestCombinedEnduranceAndRetention verifies that a device can withstand
// both cycling fatigue AND retention loss simultaneously.
//
// This represents realistic usage: cycling for some time, then data retention.
func TestCombinedEnduranceAndRetention(t *testing.T) {
	material := DefaultHZO()

	// Scenario: Device cycled to 50% of endurance, then held at 85°C for 10 years
	cycleCount := 5e9        // 50% of 1e10 endurance
	retentionTime := 3.15e8  // 10 years
	retentionTemp := 358.0   // 85°C

	// Calculate Pr after cycling
	PrAfterCycling := material.EnduranceAtCycles(cycleCount)
	cyclingRetention := (PrAfterCycling / material.Pr) * 100

	// Calculate further degradation due to retention
	PrFinal := (PrAfterCycling / material.Pr) * material.RetentionAtTime(retentionTime, retentionTemp)
	finalRetention := (PrFinal / material.Pr) * 100

	t.Logf("After %.1e cycles: Pr retention = %.1f%%", cycleCount, cyclingRetention)
	t.Logf("After additional %.1e s at %.0fK: Pr retention = %.1f%%",
		retentionTime, retentionTemp, finalRetention)

	// Final retention should still be reasonable (> 40% at 50% endurance + 10yr retention)
	// This represents ~45% cycling loss * ~99% retention loss ≈ 44%
	if finalRetention < 40.0 {
		t.Errorf("Combined endurance+retention = %.1f%%, below 40%% threshold", finalRetention)
	}
}

// TestWakeUpSaturation verifies that wake-up effect saturates
// and doesn't exceed 1.0 (unphysical).
func TestWakeUpSaturation(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	// Cycle far beyond wake-up completion (10x the wakeup cycles)
	Emax := material.Ec * 2.0
	for i := 0; i < 1000; i++ {
		_, _ = model.GetHysteresisLoop(Emax, 50)
	}

	_, _, wakeup := model.GetFatigueState()

	// Wake-up should saturate at 1.0, never exceed it
	if wakeup > 1.0 {
		t.Errorf("Wake-up factor = %f exceeds 1.0 (unphysical)", wakeup)
	}

	// Should be very close to 1.0 after many cycles
	if wakeup < 0.99 {
		t.Errorf("Wake-up factor = %f should saturate near 1.0 after 1000 cycles", wakeup)
	}

	t.Logf("After 1000 cycles: wakeup = %.4f (saturated)", wakeup)
}

// TestFatigueRateRealism verifies that fatigue rate is physically reasonable
// and doesn't cause excessive degradation within endurance limit.
func TestFatigueRateRealism(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	// Simulate cycling to a reasonable test count (1000 cycles)
	// The Preisach model has very low fatigue rate (1e-10), so we test
	// that degradation remains minimal after realistic cycling.
	numCycles := 1000
	Emax := material.Ec * 2.0

	for i := 0; i < numCycles; i++ {
		_, _ = model.GetHysteresisLoop(Emax, 50)
	}

	_, degradation, _ := model.GetFatigueState()

	// At 1000 cycles, degradation should be negligible (< 0.001%)
	if degradation > 0.00001 {
		t.Errorf("Degradation = %.6f%% at 1000 cycles (too high)", degradation*100)
	}

	t.Logf("At %d cycles: degradation = %.6f%%", numCycles, degradation*100)
}

// TestZeroCycleFatigue verifies that uncycled material has zero fatigue.
func TestZeroCycleFatigue(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	cycles, degradation, _ := model.GetFatigueState()

	if cycles != 0 {
		t.Errorf("Uncycled material has cycle count = %d, expected 0", cycles)
	}

	if degradation != 0.0 {
		t.Errorf("Uncycled material has degradation = %f, expected 0.0", degradation)
	}

	t.Logf("Initial state: cycles = %d, degradation = %.6f", cycles, degradation)
}

// TestEnduranceModelMonotonicity verifies that Pr decreases monotonically
// with cycle count (never increases).
func TestEnduranceModelMonotonicity(t *testing.T) {
	material := DefaultHZO()

	cycleCounts := []float64{0, 1e8, 1e9, 5e9, 1e10}
	var prevPr float64 = material.Pr

	for _, N := range cycleCounts {
		PrN := material.EnduranceAtCycles(N)

		if PrN > prevPr {
			t.Errorf("Pr INCREASED with cycling (unphysical): Pr(%.1e) = %.3e > Pr(prev) = %.3e",
				N, PrN, prevPr)
		}

		retentionPct := (PrN / material.Pr) * 100
		t.Logf("N = %.1e: Pr = %.3e C/m² (%.1f%% retention)", N, PrN, retentionPct)

		prevPr = PrN
	}
}

// TestRetentionModelMonotonicity verifies that Pr decreases monotonically
// with time at constant temperature (never increases).
func TestRetentionModelMonotonicity(t *testing.T) {
	material := DefaultHZO()
	temp := 358.0 // 85°C

	times := []float64{0, 1e5, 1e6, 1e7, 1e8, 3.15e8} // 0 to 10 years
	var prevPr float64 = material.Pr

	for _, time := range times {
		PrT := material.RetentionAtTime(time, temp)

		if PrT > prevPr {
			t.Errorf("Pr INCREASED with time (unphysical): Pr(t=%.1e) = %.3e > Pr(prev) = %.3e",
				time, PrT, prevPr)
		}

		retentionPct := (PrT / material.Pr) * 100
		t.Logf("t = %.1e s (%.1f days): Pr = %.3e C/m² (%.1f%% retention)",
			time, time/(86400), PrT, retentionPct)

		prevPr = PrT
	}
}
