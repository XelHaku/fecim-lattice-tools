package ferroelectric

import (
	"math"
	"testing"
)

// TestHelper: verifyUniformSpacing checks that states are uniformly spaced within tolerance.
// Returns true if max deviation from ideal spacing is within tolerancePct of Ps.
func verifyUniformSpacing(states []float64, Ps float64, tolerancePct float64) (bool, float64) {
	if len(states) < 2 {
		return true, 0
	}

	// Calculate ideal spacing
	idealSpacing := 2 * Ps / float64(len(states)-1)

	// Check spacing between all adjacent states
	maxDeviationPct := 0.0
	for i := 0; i < len(states)-1; i++ {
		actualSpacing := states[i+1] - states[i]
		deviation := math.Abs(actualSpacing - idealSpacing)
		deviationPct := (deviation / Ps) * 100
		if deviationPct > maxDeviationPct {
			maxDeviationPct = deviationPct
		}
	}

	return maxDeviationPct <= tolerancePct, maxDeviationPct
}

// Test30LevelQuantizationFullRange verifies all 30 discrete states span -Ps to +Ps uniformly.
func Test30LevelQuantizationFullRange(t *testing.T) {
	material := FeCIMMaterial()
	Ps := material.Ps

	t.Run("PreisachModel", func(t *testing.T) {
		model := NewPreisachModel(material)
		states := model.DiscreteStates(30)

		// Verify we got 30 states
		if len(states) != 30 {
			t.Errorf("Expected 30 states, got %d", len(states))
		}

		// Level 0: P ~ -Ps (within 5%)
		if math.Abs(states[0]+Ps)/Ps > 0.05 {
			t.Errorf("Level 0 polarization: got %.3e, expected ~%.3e (%.2f%% error)", states[0], -Ps, math.Abs(states[0]+Ps)/Ps*100)
		}

		// Level 15 (middle): P ~ 0 (within 5% of Ps)
		if math.Abs(states[15])/Ps > 0.05 {
			t.Errorf("Level 15 polarization: got %.3e, expected ~0 (off by %.2f%% of Ps)", states[15], math.Abs(states[15])/Ps*100)
		}

		// Level 29: P ~ +Ps (within 5%)
		if math.Abs(states[29]-Ps)/Ps > 0.05 {
			t.Errorf("Level 29 polarization: got %.3e, expected ~%.3e (%.2f%% error)", states[29], Ps, math.Abs(states[29]-Ps)/Ps*100)
		}

		// Verify uniform spacing within 1%
		uniform, maxDev := verifyUniformSpacing(states, Ps, 1.0)
		if !uniform {
			t.Errorf("Spacing not uniform: max deviation %.2f%% of Ps (tolerance: 1%%)", maxDev)
		}
	})

	t.Run("MayergoyzPreisach", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 50) // 50x50 grid
		discreteStates := model.DiscreteStates(30)

		// Verify we got 30 states
		if len(discreteStates) != 30 {
			t.Errorf("Expected 30 states, got %d", len(discreteStates))
		}

		// Extract polarization values for spacing check
		pValues := make([]float64, len(discreteStates))
		for i, s := range discreteStates {
			pValues[i] = s.Polarization
		}

		// Level 0: P ~ -Ps (within 5%)
		if math.Abs(discreteStates[0].Polarization+Ps)/Ps > 0.05 {
			t.Errorf("Level 0 polarization: got %.3e, expected ~%.3e", discreteStates[0].Polarization, -Ps)
		}

		// Level 15 (middle): P ~ 0 (within 5% of Ps)
		if math.Abs(discreteStates[15].Polarization)/Ps > 0.05 {
			t.Errorf("Level 15 polarization: got %.3e, expected ~0", discreteStates[15].Polarization)
		}

		// Level 29: P ~ +Ps (within 5%)
		if math.Abs(discreteStates[29].Polarization-Ps)/Ps > 0.05 {
			t.Errorf("Level 29 polarization: got %.3e, expected ~%.3e", discreteStates[29].Polarization, Ps)
		}

		// Verify uniform spacing within 1%
		uniform, maxDev := verifyUniformSpacing(pValues, Ps, 1.0)
		if !uniform {
			t.Errorf("Spacing not uniform: max deviation %.2f%% of Ps (tolerance: 1%%)", maxDev)
		}
	})
}

// TestStateSeparationMargin verifies adjacent states are sufficiently separated.
func TestStateSeparationMargin(t *testing.T) {
	material := FeCIMMaterial()
	Ps := material.Ps
	minSeparation := 0.02 * Ps // 2% of total range

	t.Run("PreisachModel", func(t *testing.T) {
		model := NewPreisachModel(material)
		states := model.DiscreteStates(30)

		for i := 0; i < len(states)-1; i++ {
			separation := states[i+1] - states[i]
			if separation < minSeparation {
				t.Errorf("Insufficient separation between states %d and %d: %.3e < %.3e (%.2f%% of Ps)",
					i, i+1, separation, minSeparation, (separation/Ps)*100)
			}
		}
	})

	t.Run("MayergoyzPreisach", func(t *testing.T) {
		model := NewMayergoyzPreisach(material, 50)
		discreteStates := model.DiscreteStates(30)

		for i := 0; i < len(discreteStates)-1; i++ {
			separation := discreteStates[i+1].Polarization - discreteStates[i].Polarization
			if separation < minSeparation {
				t.Errorf("Insufficient separation between states %d and %d: %.3e < %.3e",
					i, i+1, separation, minSeparation)
			}
		}
	})
}

// TestConductanceRangeMapping verifies 1-100 µS conductance range maps correctly to 30 states.
func TestConductanceRangeMapping(t *testing.T) {
	material := FeCIMMaterial()
	model := NewMayergoyzPreisach(material, 50)
	states := model.DiscreteStates(30)

	// Check state 0 (minimum)
	Gmin := states[0].Conductance
	expectedGmin := 1e-6 // 1 µS
	// Allow ±50% for FeTFT variation
	if Gmin < expectedGmin*0.5 || Gmin > expectedGmin*1.5 {
		t.Errorf("State 0 conductance %.3e µS outside expected range [%.3e, %.3e] µS (±50%%)",
			Gmin*1e6, expectedGmin*0.5*1e6, expectedGmin*1.5*1e6)
	}

	// Check state 29 (maximum)
	Gmax := states[29].Conductance
	expectedGmax := 100e-6 // 100 µS
	// Allow ±50% for FeTFT variation
	if Gmax < expectedGmax*0.5 || Gmax > expectedGmax*1.5 {
		t.Errorf("State 29 conductance %.3e µS outside expected range [%.3e, %.3e] µS (±50%%)",
			Gmax*1e6, expectedGmax*0.5*1e6, expectedGmax*1.5*1e6)
	}

	// Check on/off ratio > 10
	ratio := Gmax / Gmin
	if ratio < 10.0 {
		t.Errorf("Gmax/Gmin ratio %.2f < 10 (insufficient on/off ratio)", ratio)
	}

	// Verify monotonic increase
	for i := 0; i < len(states)-1; i++ {
		if states[i+1].Conductance <= states[i].Conductance {
			t.Errorf("Non-monotonic conductance: state %d (%.3e µS) >= state %d (%.3e µS)",
				i, states[i].Conductance*1e6, i+1, states[i+1].Conductance*1e6)
		}
	}

	// Report actual values for reference
	t.Logf("State 0: G = %.3f µS", Gmin*1e6)
	t.Logf("State 29: G = %.3f µS", Gmax*1e6)
	t.Logf("On/Off ratio: %.2f", ratio)
}

// TestVariableLevelCount tests quantization with different level counts.
func TestVariableLevelCount(t *testing.T) {
	testCases := []struct {
		name   string
		levels int
		desc   string
	}{
		{"8-level", 8, "MLC flash equivalent"},
		{"16-level", 16, "TLC/QLC flash"},
		{"32-level", 32, "Oh et al. 2017"},
		{"64-level", 64, "Superlattice target"},
		{"140-level", 140, "Song et al. 2024 FTJ"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			material := FeCIMMaterial()
			Ps := material.Ps

			// Test with PreisachModel
			t.Run("PreisachModel", func(t *testing.T) {
				model := NewPreisachModel(material)
				states := model.DiscreteStates(tc.levels)

				if len(states) != tc.levels {
					t.Errorf("Expected %d states, got %d", tc.levels, len(states))
				}

				// Verify endpoint coverage
				if math.Abs(states[0]+Ps)/Ps > 0.05 {
					t.Errorf("First state: got %.3e, expected ~%.3e", states[0], -Ps)
				}
				if math.Abs(states[tc.levels-1]-Ps)/Ps > 0.05 {
					t.Errorf("Last state: got %.3e, expected ~%.3e", states[tc.levels-1], Ps)
				}

				// Verify uniform spacing within 1%
				uniform, maxDev := verifyUniformSpacing(states, Ps, 1.0)
				if !uniform {
					t.Errorf("Spacing not uniform: max deviation %.2f%% (tolerance: 1%%)", maxDev)
				}

				t.Logf("%s (%d levels): spacing %.2f%% of Ps per level", tc.desc, tc.levels, 100.0/float64(tc.levels-1))
			})

			// Test with MayergoyzPreisach
			t.Run("MayergoyzPreisach", func(t *testing.T) {
				model := NewMayergoyzPreisach(material, 50)
				discreteStates := model.DiscreteStates(tc.levels)

				if len(discreteStates) != tc.levels {
					t.Errorf("Expected %d states, got %d", tc.levels, len(discreteStates))
				}

				// Extract polarization values
				pValues := make([]float64, len(discreteStates))
				for i, s := range discreteStates {
					pValues[i] = s.Polarization
				}

				// Verify endpoint coverage
				if math.Abs(discreteStates[0].Polarization+Ps)/Ps > 0.05 {
					t.Errorf("First state: got %.3e, expected ~%.3e", discreteStates[0].Polarization, -Ps)
				}
				if math.Abs(discreteStates[tc.levels-1].Polarization-Ps)/Ps > 0.05 {
					t.Errorf("Last state: got %.3e, expected ~%.3e", discreteStates[tc.levels-1].Polarization, Ps)
				}

				// Verify uniform spacing within 1%
				uniform, maxDev := verifyUniformSpacing(pValues, Ps, 1.0)
				if !uniform {
					t.Errorf("Spacing not uniform: max deviation %.2f%% (tolerance: 1%%)", maxDev)
				}
			})
		})
	}
}

// TestNormalizedPolarizationRange verifies normalized polarization is in [-1, +1].
func TestNormalizedPolarizationRange(t *testing.T) {
	material := FeCIMMaterial()
	model := NewMayergoyzPreisach(material, 50)
	states := model.DiscreteStates(30)

	for i, s := range states {
		if s.NormalizedP < -1.0 || s.NormalizedP > 1.0 {
			t.Errorf("State %d normalized polarization %.3f outside [-1, +1] range", i, s.NormalizedP)
		}
	}

	// Check endpoints are close to ±1
	if math.Abs(states[0].NormalizedP+1.0) > 0.05 {
		t.Errorf("State 0 normalized P: got %.3f, expected ~-1.0", states[0].NormalizedP)
	}
	if math.Abs(states[29].NormalizedP-1.0) > 0.05 {
		t.Errorf("State 29 normalized P: got %.3f, expected ~+1.0", states[29].NormalizedP)
	}
}

// TestLevelIndexing verifies state levels are correctly indexed 0 to N-1.
func TestLevelIndexing(t *testing.T) {
	material := FeCIMMaterial()
	model := NewMayergoyzPreisach(material, 50)

	testCases := []int{8, 16, 30, 32, 64, 140}

	for _, N := range testCases {
		t.Run(string(rune(N)), func(t *testing.T) {
			states := model.DiscreteStates(N)

			for i, s := range states {
				if s.Level != i {
					t.Errorf("State index mismatch: position %d has Level=%d, expected %d", i, s.Level, i)
				}
			}
		})
	}
}

// TestConductanceMonotonicity verifies conductance increases monotonically with state level.
func TestConductanceMonotonicity(t *testing.T) {
	materials := []*HZOMaterial{
		FeCIMMaterial(),
		DefaultHZO(),
		LiteratureSuperlattice(),
		HZOStandard32(),
	}

	for _, material := range materials {
		t.Run(material.Name, func(t *testing.T) {
			model := NewMayergoyzPreisach(material, 50)
			states := model.DiscreteStates(30)

			for i := 0; i < len(states)-1; i++ {
				if states[i+1].Conductance <= states[i].Conductance {
					t.Errorf("Non-monotonic: G[%d]=%.3e >= G[%d]=%.3e",
						i, states[i].Conductance, i+1, states[i+1].Conductance)
				}

				// Also check polarization monotonicity
				if states[i+1].Polarization <= states[i].Polarization {
					t.Errorf("Non-monotonic polarization: P[%d]=%.3e >= P[%d]=%.3e",
						i, states[i].Polarization, i+1, states[i+1].Polarization)
				}
			}
		})
	}
}

// TestStateReproducibility verifies that DiscreteStates returns consistent results.
func TestStateReproducibility(t *testing.T) {
	material := FeCIMMaterial()
	model1 := NewPreisachModel(material)
	model2 := NewPreisachModel(material)

	states1 := model1.DiscreteStates(30)
	states2 := model2.DiscreteStates(30)

	for i := range states1 {
		if math.Abs(states1[i]-states2[i]) > 1e-15 {
			t.Errorf("States not reproducible: state %d differs: %.3e vs %.3e", i, states1[i], states2[i])
		}
	}
}

// BenchmarkDiscreteStates30 benchmarks 30-level quantization performance.
func BenchmarkDiscreteStates30(b *testing.B) {
	material := FeCIMMaterial()

	b.Run("PreisachModel", func(b *testing.B) {
		model := NewPreisachModel(material)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.DiscreteStates(30)
		}
	})

	b.Run("MayergoyzPreisach", func(b *testing.B) {
		model := NewMayergoyzPreisach(material, 50)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.DiscreteStates(30)
		}
	})
}

// BenchmarkDiscreteStates140 benchmarks 140-level quantization (FTJ case).
func BenchmarkDiscreteStates140(b *testing.B) {
	material := HZOFJT140()

	b.Run("PreisachModel", func(b *testing.B) {
		model := NewPreisachModel(material)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.DiscreteStates(140)
		}
	})

	b.Run("MayergoyzPreisach", func(b *testing.B) {
		model := NewMayergoyzPreisach(material, 50)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.DiscreteStates(140)
		}
	})
}
