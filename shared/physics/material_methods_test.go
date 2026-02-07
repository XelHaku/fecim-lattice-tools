package physics

import (
	"math"
	"testing"
)

const tolerance = 1e-6

// relativeError returns the relative error between actual and expected values.
func relativeError(actual, expected float64) float64 {
	if expected == 0 {
		return math.Abs(actual)
	}
	return math.Abs((actual - expected) / expected)
}

// assertClose checks if two float64 values are within relative tolerance.
func assertClose(t *testing.T, actual, expected float64, name string) {
	t.Helper()
	if relErr := relativeError(actual, expected); relErr > tolerance {
		t.Errorf("%s: got %e, want %e (relative error: %e)", name, actual, expected, relErr)
	}
}

// TestAllPresets_NonNilAndValidFields verifies all presets have valid parameters.
func TestAllPresets_NonNilAndValidFields(t *testing.T) {
	materials := AllMaterials()
	if len(materials) == 0 {
		t.Fatal("AllMaterials() returned empty slice")
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			if mat.Name == "" {
				t.Error("Name is empty")
			}
			if mat.Ps <= 0 {
				t.Errorf("Ps must be > 0, got %e", mat.Ps)
			}
			if mat.Ec <= 0 {
				t.Errorf("Ec must be > 0, got %e", mat.Ec)
			}
			if mat.Thickness <= 0 {
				t.Errorf("Thickness must be > 0, got %e", mat.Thickness)
			}
			if mat.NumLevels <= 0 {
				t.Errorf("NumLevels must be > 0, got %d", mat.NumLevels)
			}
			if mat.Pr <= 0 {
				t.Errorf("Pr must be > 0, got %e", mat.Pr)
			}
			if mat.Area <= 0 {
				t.Errorf("Area must be > 0, got %e", mat.Area)
			}
			// CurieTemp can be 0 for materials loaded from YAML config that don't specify it
			// Only validate if it's set
			if mat.CurieTemp < 0 {
				t.Errorf("CurieTemp must be >= 0, got %e", mat.CurieTemp)
			}
		})
	}
}

// TestPreset_DefaultHZO_Values spot-checks DefaultHZO preset values.
func TestPreset_DefaultHZO_Values(t *testing.T) {
	mat := DefaultHZO()

	if mat.Name != "HZO (Si-doped)" {
		t.Errorf("Name = %q, want %q", mat.Name, "HZO (Si-doped)")
	}
	assertClose(t, mat.Pr, 25e-2, "Pr")
	assertClose(t, mat.Ps, 30e-2, "Ps")
	assertClose(t, mat.Ec, 1.2e8, "Ec")
	assertClose(t, mat.Thickness, 10e-9, "Thickness")
	if mat.NumLevels != 30 {
		t.Errorf("NumLevels = %d, want 30", mat.NumLevels)
	}
}

// TestPreset_FeCIMMaterial_Values spot-checks FeCIMMaterial preset.
func TestPreset_FeCIMMaterial_Values(t *testing.T) {
	mat := FeCIMMaterial()

	if mat.Name != "FeCIM HZO" {
		t.Errorf("Name = %q, want %q", mat.Name, "FeCIM HZO")
	}
	assertClose(t, mat.EnduranceCycles, 1e9, "EnduranceCycles")
	assertClose(t, mat.RetentionTime, 1e7, "RetentionTime")
	if mat.NumLevels != 30 {
		t.Errorf("NumLevels = %d, want 30", mat.NumLevels)
	}
}

// TestPreset_FeCIMMaterialTarget_IsExtension verifies FeCIMMaterialTarget extends FeCIMMaterial.
func TestPreset_FeCIMMaterialTarget_IsExtension(t *testing.T) {
	mat := FeCIMMaterialTarget()

	if mat.Name != "FeCIM HZO (TARGET - NOT DEMONSTRATED)" {
		t.Errorf("Name = %q, unexpected", mat.Name)
	}
	assertClose(t, mat.EnduranceCycles, 1e12, "EnduranceCycles")
	assertClose(t, mat.RetentionTime, 3.15e9, "RetentionTime")
	assertClose(t, mat.Tau, 1e-9, "Tau")
}

// TestPreset_CryogenicHZO_EnhancedPolarization verifies cryogenic enhancement.
func TestPreset_CryogenicHZO_EnhancedPolarization(t *testing.T) {
	mat := CryogenicHZO()

	assertClose(t, mat.Pr, 75e-2, "Pr (enhanced at 4K)")
	if mat.NumLevels != 48 {
		t.Errorf("NumLevels = %d, want 48", mat.NumLevels)
	}
}

// TestPreset_AlScN_HighEcHighPr verifies AlScN characteristics.
func TestPreset_AlScN_HighEcHighPr(t *testing.T) {
	mat := AlScN()

	assertClose(t, mat.Pr, 120e-2, "Pr")
	assertClose(t, mat.Ec, 5.0e8, "Ec")
	if mat.NumLevels != 12 {
		t.Errorf("NumLevels = %d, want 12", mat.NumLevels)
	}
}

// TestGetNumLevels_Default tests default level fallback.
func TestGetNumLevels_Default(t *testing.T) {
	mat := &HZOMaterial{NumLevels: 0}
	levels := mat.GetNumLevels()
	if levels != 30 {
		t.Errorf("GetNumLevels() = %d, want 30 (DefaultLevels)", levels)
	}
}

// TestGetNumLevels_Custom tests custom level specification.
func TestGetNumLevels_Custom(t *testing.T) {
	mat := &HZOMaterial{NumLevels: 64}
	levels := mat.GetNumLevels()
	if levels != 64 {
		t.Errorf("GetNumLevels() = %d, want 64", levels)
	}
}

// TestCoerciveVoltage tests coercive voltage calculation.
func TestCoerciveVoltage(t *testing.T) {
	mat := DefaultHZO()
	voltage := mat.CoerciveVoltage()
	expected := 1.2e8 * 10e-9 // 1.2 V
	assertClose(t, voltage, expected, "CoerciveVoltage")
}

// TestRemanentCharge tests remanent charge retrieval.
func TestRemanentCharge(t *testing.T) {
	mat := DefaultHZO()
	charge := mat.RemanentCharge()
	if charge != mat.Pr {
		t.Errorf("RemanentCharge() = %e, want %e", charge, mat.Pr)
	}
}

// TestCapacitance tests capacitance calculation.
func TestCapacitance(t *testing.T) {
	mat := DefaultHZO()
	cap := mat.Capacitance()
	epsilon0 := 8.854e-12
	expected := epsilon0 * mat.Epsilon * mat.Area / mat.Thickness
	assertClose(t, cap, expected, "Capacitance")
}

// TestSwitchingEnergy tests switching energy formula.
func TestSwitchingEnergy(t *testing.T) {
	mat := DefaultHZO()
	energy := mat.SwitchingEnergy()
	expected := 2 * mat.Pr * mat.Ec * mat.Area * mat.Thickness
	assertClose(t, energy, expected, "SwitchingEnergy")
}

// TestSwitchingTime_RoomTemp tests switching time at room temperature.
func TestSwitchingTime_RoomTemp(t *testing.T) {
	mat := DefaultHZO()
	T := 300.0 // K
	switchTime := mat.SwitchingTime(T)

	// Should be greater than Tau0 due to exponential activation
	if switchTime <= mat.Tau0 {
		t.Errorf("SwitchingTime(%f) = %e, expected > Tau0 (%e)", T, switchTime, mat.Tau0)
	}

	// Verify formula: τ(T) = τ0 * exp(Ea / kT)
	kB := 8.617e-5
	expected := mat.Tau0 * math.Exp(mat.Ea/(kB*T))
	assertClose(t, switchTime, expected, "SwitchingTime at 300K")
}

// TestSwitchingTime_HighTemp tests faster switching at higher temperature.
func TestSwitchingTime_HighTemp(t *testing.T) {
	mat := DefaultHZO()
	lowT := 300.0
	highT := 400.0

	timeLowT := mat.SwitchingTime(lowT)
	timeHighT := mat.SwitchingTime(highT)

	if timeHighT >= timeLowT {
		t.Errorf("SwitchingTime should decrease with temperature: %e (400K) >= %e (300K)", timeHighT, timeLowT)
	}
}

// TestCoerciveFieldAtTemp_BelowCurie tests Ec below Curie temperature.
func TestCoerciveFieldAtTemp_BelowCurie(t *testing.T) {
	mat := DefaultHZO()
	T := 300.0 // K, below 723 K
	Ec := mat.CoerciveFieldAtTemp(T)
	expected := mat.Ec * math.Pow(1-T/mat.CurieTemp, 0.5)
	assertClose(t, Ec, expected, "CoerciveFieldAtTemp at 300K")
}

// TestCoerciveFieldAtTemp_AtCurie tests Ec at Curie temperature.
func TestCoerciveFieldAtTemp_AtCurie(t *testing.T) {
	mat := DefaultHZO()
	Ec := mat.CoerciveFieldAtTemp(mat.CurieTemp)
	if Ec != 0 {
		t.Errorf("CoerciveFieldAtTemp(Tc) = %e, want 0", Ec)
	}
}

// TestCoerciveFieldAtTemp_AboveCurie tests Ec above Curie temperature.
func TestCoerciveFieldAtTemp_AboveCurie(t *testing.T) {
	mat := DefaultHZO()
	T := mat.CurieTemp + 100
	Ec := mat.CoerciveFieldAtTemp(T)
	if Ec != 0 {
		t.Errorf("CoerciveFieldAtTemp(T > Tc) = %e, want 0", Ec)
	}
}

// TestPolarizationAtTemp_BelowCurie tests Pr below Curie temperature.
func TestPolarizationAtTemp_BelowCurie(t *testing.T) {
	mat := DefaultHZO()
	T := 300.0 // K
	Pr := mat.PolarizationAtTemp(T)
	expected := mat.Pr * math.Pow(1-T/mat.CurieTemp, 0.5)
	assertClose(t, Pr, expected, "PolarizationAtTemp at 300K")
}

// TestPolarizationAtTemp_AtCurie tests Pr at Curie temperature.
func TestPolarizationAtTemp_AtCurie(t *testing.T) {
	mat := DefaultHZO()
	Pr := mat.PolarizationAtTemp(mat.CurieTemp)
	if Pr != 0 {
		t.Errorf("PolarizationAtTemp(Tc) = %e, want 0", Pr)
	}
}

// TestEnduranceAtCycles_Fresh tests endurance at zero cycles.
func TestEnduranceAtCycles_Fresh(t *testing.T) {
	mat := DefaultHZO()
	Pr := mat.EnduranceAtCycles(0)
	// At N=0, exp(0) = 1, so should return full Pr
	assertClose(t, Pr, mat.Pr, "EnduranceAtCycles at N=0")
}

// TestEnduranceAtCycles_Degraded tests endurance at rated cycles.
func TestEnduranceAtCycles_Degraded(t *testing.T) {
	mat := DefaultHZO()
	N := mat.EnduranceCycles
	Pr := mat.EnduranceAtCycles(N)

	// At N = EnduranceCycles, exp(-1^0.3) ≈ exp(-1) ≈ 0.368
	beta := 0.3
	expected := mat.Pr * math.Exp(-math.Pow(1.0, beta))
	assertClose(t, Pr, expected, "EnduranceAtCycles at EnduranceCycles")
}

// TestEnduranceAtCycles_VeryHigh tests endurance at very high cycles.
func TestEnduranceAtCycles_VeryHigh(t *testing.T) {
	mat := DefaultHZO()
	N := mat.EnduranceCycles * 100
	Pr := mat.EnduranceAtCycles(N)

	// Should be small (< 5% of original Pr due to stretched exponential)
	if Pr > 0.05*mat.Pr {
		t.Errorf("EnduranceAtCycles(%e) = %e, expected < 5%% of Pr (%e)", N, Pr, 0.05*mat.Pr)
	}
}

// TestRetentionAtTime_Fresh tests retention at t=0.
func TestRetentionAtTime_Fresh(t *testing.T) {
	mat := DefaultHZO()
	Tref := 358.0 // 85°C
	Pr := mat.RetentionAtTime(0, Tref)
	assertClose(t, Pr, mat.Pr, "RetentionAtTime at t=0")
}

// TestRetentionAtTime_EndOfLife tests retention at end of life.
func TestRetentionAtTime_EndOfLife(t *testing.T) {
	mat := DefaultHZO()
	Tref := 358.0 // 85°C reference temperature
	time := mat.RetentionTime
	Pr := mat.RetentionAtTime(time, Tref)

	// Should return 0.9 * Pr (10% loss at end of life)
	expected := 0.9 * mat.Pr
	assertClose(t, Pr, expected, "RetentionAtTime at end of life")
}

// TestDiscreteLevel_Boundaries tests boundary conditions for discrete levels.
func TestDiscreteLevel_Boundaries(t *testing.T) {
	mat := DefaultHZO()
	totalLevels := 30

	// Level 0 should give Gmin
	G0 := mat.DiscreteLevel(0, totalLevels)
	assertClose(t, G0, mat.Gmin, "DiscreteLevel(0)")

	// Level N-1 should give Gmax
	GMax := mat.DiscreteLevel(totalLevels-1, totalLevels)
	assertClose(t, GMax, mat.Gmax, "DiscreteLevel(N-1)")
}

// TestDiscreteLevel_Monotonic tests that levels are strictly increasing.
func TestDiscreteLevel_Monotonic(t *testing.T) {
	mat := DefaultHZO()
	totalLevels := 30

	prevG := mat.DiscreteLevel(0, totalLevels)
	for level := 1; level < totalLevels; level++ {
		G := mat.DiscreteLevel(level, totalLevels)
		if G <= prevG {
			t.Errorf("DiscreteLevel not monotonic: level %d (%e) <= level %d (%e)",
				level, G, level-1, prevG)
		}
		prevG = G
	}
}

// TestDiscreteLevel_SingleLevel tests degenerate case with single level.
func TestDiscreteLevel_SingleLevel(t *testing.T) {
	mat := DefaultHZO()
	G := mat.DiscreteLevel(0, 1)
	expected := (mat.Gmin + mat.Gmax) / 2
	assertClose(t, G, expected, "DiscreteLevel with totalLevels=1")
}

// TestDiscreteLevel_ZeroGminGmax tests fallback when Gmin/Gmax are zero.
func TestDiscreteLevel_ZeroGminGmax(t *testing.T) {
	mat := &HZOMaterial{
		Gmin: 0,
		Gmax: 0,
	}

	G0 := mat.DiscreteLevel(0, 30)
	G29 := mat.DiscreteLevel(29, 30)

	// Should use fallback values: Gmin=1e-6, Gmax=100e-6
	expectedGmin := 1e-6
	expectedGmax := 100e-6

	assertClose(t, G0, expectedGmin, "DiscreteLevel(0) with zero Gmin/Gmax")
	assertClose(t, G29, expectedGmax, "DiscreteLevel(29) with zero Gmin/Gmax")
}

// TestDiscreteLevel_MidLevel tests midpoint level calculation.
func TestDiscreteLevel_MidLevel(t *testing.T) {
	mat := DefaultHZO()
	totalLevels := 30
	midLevel := (totalLevels - 1) / 2 // Level 14 for 30 levels

	G := mat.DiscreteLevel(midLevel, totalLevels)
	expected := (mat.Gmin + mat.Gmax) / 2

	// For 30 levels, mid level is 14 (0-indexed), giving normalizedP = -1 + 2*14/29 = -0.034
	// This is close to 0, so G should be close to (Gmin+Gmax)/2, but not exact
	// Allow 10% tolerance since the exact midpoint isn't guaranteed for even counts
	relErr := relativeError(G, expected)
	if relErr > 0.1 {
		t.Errorf("DiscreteLevel at midpoint: got %e, want %e (relative error: %e > 0.1)", G, expected, relErr)
	}
}

// TestAllMethods_TableDriven runs all methods on various presets.
func TestAllMethods_TableDriven(t *testing.T) {
	presets := []struct {
		name     string
		material *HZOMaterial
	}{
		{"DefaultHZO", DefaultHZO()},
		{"FeCIMMaterial", FeCIMMaterial()},
		{"LiteratureSuperlattice", LiteratureSuperlattice()},
		{"CryogenicHZO", CryogenicHZO()},
		{"AlScN", AlScN()},
	}

	for _, preset := range presets {
		t.Run(preset.name, func(t *testing.T) {
			mat := preset.material

			// Test all methods don't panic and return reasonable values
			t.Run("GetNumLevels", func(t *testing.T) {
				levels := mat.GetNumLevels()
				if levels <= 0 {
					t.Errorf("GetNumLevels() = %d, must be > 0", levels)
				}
			})

			t.Run("CoerciveVoltage", func(t *testing.T) {
				V := mat.CoerciveVoltage()
				if V <= 0 {
					t.Errorf("CoerciveVoltage() = %e, must be > 0", V)
				}
			})

			t.Run("RemanentCharge", func(t *testing.T) {
				Q := mat.RemanentCharge()
				if Q <= 0 {
					t.Errorf("RemanentCharge() = %e, must be > 0", Q)
				}
			})

			t.Run("Capacitance", func(t *testing.T) {
				C := mat.Capacitance()
				if C <= 0 {
					t.Errorf("Capacitance() = %e, must be > 0", C)
				}
			})

			t.Run("SwitchingEnergy", func(t *testing.T) {
				E := mat.SwitchingEnergy()
				if E <= 0 {
					t.Errorf("SwitchingEnergy() = %e, must be > 0", E)
				}
			})

			t.Run("SwitchingTime", func(t *testing.T) {
				tau := mat.SwitchingTime(300)
				if tau <= 0 {
					t.Errorf("SwitchingTime(300K) = %e, must be > 0", tau)
				}
			})

			t.Run("CoerciveFieldAtTemp", func(t *testing.T) {
				Ec := mat.CoerciveFieldAtTemp(300)
				if Ec <= 0 || Ec > mat.Ec {
					t.Errorf("CoerciveFieldAtTemp(300K) = %e, must be in (0, %e]", Ec, mat.Ec)
				}
			})

			t.Run("PolarizationAtTemp", func(t *testing.T) {
				Pr := mat.PolarizationAtTemp(300)
				if Pr <= 0 || Pr > mat.Pr {
					t.Errorf("PolarizationAtTemp(300K) = %e, must be in (0, %e]", Pr, mat.Pr)
				}
			})

			t.Run("EnduranceAtCycles", func(t *testing.T) {
				Pr := mat.EnduranceAtCycles(0)
				if Pr <= 0 || Pr > mat.Pr {
					t.Errorf("EnduranceAtCycles(0) = %e, must be in (0, %e]", Pr, mat.Pr)
				}
			})

			t.Run("RetentionAtTime", func(t *testing.T) {
				Pr := mat.RetentionAtTime(0, 358)
				if Pr <= 0 || Pr > mat.Pr {
					t.Errorf("RetentionAtTime(0, 358K) = %e, must be in (0, %e]", Pr, mat.Pr)
				}
			})

			t.Run("DiscreteLevel", func(t *testing.T) {
				levels := mat.GetNumLevels()
				G := mat.DiscreteLevel(0, levels)
				if G <= 0 {
					t.Errorf("DiscreteLevel(0, %d) = %e, must be > 0", levels, G)
				}
			})
		})
	}
}

// TestPreset_HZOStandard32_Values verifies HZOStandard32 preset.
func TestPreset_HZOStandard32_Values(t *testing.T) {
	mat := HZOStandard32()

	if mat.Name != "HZO Standard (32 states)" {
		t.Errorf("Name = %q, unexpected", mat.Name)
	}
	if mat.NumLevels != 32 {
		t.Errorf("NumLevels = %d, want 32", mat.NumLevels)
	}
	assertClose(t, mat.Pr, 20e-2, "Pr")
}

// TestPreset_HZOFJT140_Values verifies HZOFJT140 preset.
func TestPreset_HZOFJT140_Values(t *testing.T) {
	mat := HZOFJT140()

	if mat.Name != "HZO FTJ (140 states)" {
		t.Errorf("Name = %q, unexpected", mat.Name)
	}
	if mat.NumLevels != 140 {
		t.Errorf("NumLevels = %d, want 140", mat.NumLevels)
	}
	assertClose(t, mat.Thickness, 4.5e-9, "Thickness (ultrathin)")
}

// TestPreset_HZOCustom14_Values verifies HZOCustom14 preset.
func TestPreset_HZOCustom14_Values(t *testing.T) {
	mat := HZOCustom14()

	if mat.Name != "HZO Custom (14 states)" {
		t.Errorf("Name = %q, unexpected", mat.Name)
	}
	if mat.NumLevels != 14 {
		t.Errorf("NumLevels = %d, want 14", mat.NumLevels)
	}
}
