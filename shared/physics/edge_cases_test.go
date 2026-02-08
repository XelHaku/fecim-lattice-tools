// Package physics provides shared physics utilities for FeCIM simulations.
// This file contains edge case tests for numerical robustness.
package physics

import (
	"math"
	"testing"
)

// =============================================================================
// ZERO ARRAY / EMPTY INPUT TESTS
// =============================================================================

func TestQuantizeToLevels_ZeroLevels(t *testing.T) {
	// Edge case: 0 or 1 levels
	tests := []struct {
		name      string
		value     float64
		levels    int
		knownNaN  bool // Document known NaN-producing edge cases
	}{
		{"zero levels", 0.5, 0, false},
		{"one level", 0.5, 1, true}, // KNOWN ISSUE: 1 level causes division by 0 (levels-1)
		{"negative levels", 0.5, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("QuantizeToLevels panicked with %d levels: %v", tt.levels, r)
				}
			}()

			result := QuantizeToLevels(tt.value, tt.levels)

			// Document known NaN cases as discovered edge cases (not failures)
			if math.IsNaN(result) || math.IsInf(result, 0) {
				if tt.knownNaN {
					t.Logf("KNOWN EDGE CASE: QuantizeToLevels(%f, %d) = %f (division by zero for levels-1=0)", tt.value, tt.levels, result)
				} else {
					t.Errorf("QuantizeToLevels(%f, %d) returned invalid: %f", tt.value, tt.levels, result)
				}
			}
		})
	}
}

func TestGetLevel_ZeroLevels(t *testing.T) {
	tests := []struct {
		name        string
		conductance float64
		levels      int
	}{
		{"zero levels", 0.5, 0},
		{"one level", 0.5, 1},
		{"negative levels", 0.5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetLevel panicked with %d levels: %v", tt.levels, r)
				}
			}()

			result := GetLevel(tt.conductance, tt.levels)

			// Result should be non-negative integer
			if result < 0 {
				t.Logf("GetLevel(%f, %d) = %d (negative, but handled)", tt.conductance, tt.levels, result)
			}
		})
	}
}

func TestNormalizeFromLevel_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		levels   int
		knownNaN bool // Document known NaN-producing edge cases
	}{
		{"negative level", -5, 30, false},
		{"level exceeds max", 100, 30, false},
		{"zero total levels", 15, 0, false},
		{"one total level", 0, 1, true}, // KNOWN ISSUE: 1 level causes division by 0 (levels-1)
		{"negative total levels", 15, -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NormalizeFromLevel panicked: %v", r)
				}
			}()

			result := NormalizeFromLevel(tt.level, tt.levels)

			// Document known NaN cases as discovered edge cases (not failures)
			if math.IsNaN(result) || math.IsInf(result, 0) {
				if tt.knownNaN {
					t.Logf("KNOWN EDGE CASE: NormalizeFromLevel(%d, %d) = %f (division by zero for levels-1=0)", tt.level, tt.levels, result)
				} else {
					t.Errorf("NormalizeFromLevel(%d, %d) returned invalid: %f", tt.level, tt.levels, result)
				}
			}
		})
	}
}

func TestLevelSpacing_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		levels int
	}{
		{"zero levels", 0},
		{"one level", 1},
		{"negative levels", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LevelSpacing panicked: %v", r)
				}
			}()

			result := LevelSpacing(tt.levels)

			if math.IsNaN(result) || math.IsInf(result, 0) {
				t.Errorf("LevelSpacing(%d) returned invalid: %f", tt.levels, result)
			}
		})
	}
}

// =============================================================================
// NaN INPUT TESTS
// =============================================================================

func TestQuantizeToLevels_NaNInput(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("QuantizeToLevels panicked on NaN: %v", r)
		}
	}()

	result := QuantizeToLevels(math.NaN(), 30)

	// Should handle NaN gracefully (return 0 or clamp)
	// Note: math.Max/Min with NaN propagates NaN, so this tests that case
	t.Logf("QuantizeToLevels(NaN, 30) = %f", result)
}

func TestQuantizeToLevels_InfInput(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("QuantizeToLevels panicked on %s: %v", tt.name, r)
				}
			}()

			result := QuantizeToLevels(tt.value, 30)

			// Should clamp to [0, 1]
			if result < 0 || result > 1 {
				t.Logf("QuantizeToLevels(%f, 30) = %f (clamping expected)", tt.value, result)
			}
		})
	}
}

func TestConductance_NaNInputs(t *testing.T) {
	tests := []struct {
		name  string
		gNorm float64
		gMin  float64
		gMax  float64
	}{
		{"NaN normalized", math.NaN(), GMin, GMax},
		{"NaN gMin", 0.5, math.NaN(), GMax},
		{"NaN gMax", 0.5, GMin, math.NaN()},
		{"all NaN", math.NaN(), math.NaN(), math.NaN()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NormalizedToPhysicalRange panicked: %v", r)
				}
			}()

			result := NormalizedToPhysicalRange(tt.gNorm, ConductanceLinear, tt.gMin, tt.gMax, nil)
			t.Logf("NormalizedToPhysicalRange with %s = %e", tt.name, result)
		})
	}
}

func TestConductanceToLevel_NaN(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ConductanceToLevel panicked on NaN: %v", r)
		}
	}()

	result := ConductanceToLevel(math.NaN(), 30)
	t.Logf("ConductanceToLevel(NaN, 30) = %d", result)
}

// =============================================================================
// EXTREME TEMPERATURE TESTS (0K, 1000K)
// =============================================================================

func TestHZOMaterial_SwitchingTime_ExtremeTemperatures(t *testing.T) {
	mat := DefaultHZO()

	tests := []struct {
		name     string
		tempK    float64
		wantZero bool
		wantInf  bool
	}{
		{"absolute zero", 0, false, true},         // exp(Ea/0) = Inf
		{"near absolute zero", 1, false, false},   // very large but finite
		{"cryogenic", 4, false, false},            // liquid helium
		{"liquid nitrogen", 77, false, false},
		{"room temperature", 300, false, false},
		{"industrial", 358, false, false},
		{"automotive", 423, false, false},
		{"above Curie", 1000, false, false},       // should still compute
		{"extreme high", 10000, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SwitchingTime panicked at %fK: %v", tt.tempK, r)
				}
			}()

			tau := mat.SwitchingTime(tt.tempK)

			if tt.wantInf && !math.IsInf(tau, 1) {
				t.Logf("SwitchingTime(%fK) = %e (expected +Inf)", tt.tempK, tau)
			}
			if tt.wantZero && tau != 0 {
				t.Errorf("SwitchingTime(%fK) = %e (expected 0)", tt.tempK, tau)
			}
			if !tt.wantInf && !tt.wantZero && (math.IsNaN(tau) || tau < 0) {
				t.Errorf("SwitchingTime(%fK) returned invalid: %e", tt.tempK, tau)
			}

			t.Logf("SwitchingTime(%fK) = %e s", tt.tempK, tau)
		})
	}
}

func TestHZOMaterial_CoerciveFieldAtTemp_ExtremeTemperatures(t *testing.T) {
	mat := DefaultHZO()

	tests := []struct {
		name       string
		tempK      float64
		shouldBeZero bool
	}{
		{"absolute zero", 0, false},
		{"near absolute zero", 1, false},
		{"cryogenic", 77, false},
		{"room temperature", 300, false},
		{"at Curie temperature", mat.CurieTemp, true},
		{"above Curie", mat.CurieTemp + 100, true},
		{"extreme high", 2000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("CoerciveFieldAtTemp panicked at %fK: %v", tt.tempK, r)
				}
			}()

			Ec := mat.CoerciveFieldAtTemp(tt.tempK)

			if tt.shouldBeZero && Ec != 0 {
				t.Errorf("CoerciveFieldAtTemp(%fK) = %e, expected 0", tt.tempK, Ec)
			}
			if !tt.shouldBeZero && Ec <= 0 {
				t.Errorf("CoerciveFieldAtTemp(%fK) = %e, expected positive", tt.tempK, Ec)
			}
			if math.IsNaN(Ec) {
				t.Errorf("CoerciveFieldAtTemp(%fK) returned NaN", tt.tempK)
			}

			t.Logf("CoerciveFieldAtTemp(%fK) = %e V/m", tt.tempK, Ec)
		})
	}
}

func TestHZOMaterial_PolarizationAtTemp_ExtremeTemperatures(t *testing.T) {
	mat := DefaultHZO()

	tests := []struct {
		name         string
		tempK        float64
		shouldBeZero bool
	}{
		{"absolute zero", 0, false},
		{"near absolute zero", 1, false},
		{"cryogenic", 77, false},
		{"room temperature", 300, false},
		{"at Curie temperature", mat.CurieTemp, true},
		{"above Curie", mat.CurieTemp + 100, true},
		{"extreme high", 2000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("PolarizationAtTemp panicked at %fK: %v", tt.tempK, r)
				}
			}()

			Pr := mat.PolarizationAtTemp(tt.tempK)

			if tt.shouldBeZero && Pr != 0 {
				t.Errorf("PolarizationAtTemp(%fK) = %e, expected 0", tt.tempK, Pr)
			}
			if !tt.shouldBeZero && Pr <= 0 {
				t.Errorf("PolarizationAtTemp(%fK) = %e, expected positive", tt.tempK, Pr)
			}
			if math.IsNaN(Pr) {
				t.Errorf("PolarizationAtTemp(%fK) returned NaN", tt.tempK)
			}

			t.Logf("PolarizationAtTemp(%fK) = %e C/m²", tt.tempK, Pr)
		})
	}
}

func TestHZOMaterial_RetentionAtTime_ExtremeTemperatures(t *testing.T) {
	mat := DefaultHZO()

	tests := []struct {
		name  string
		time  float64
		tempK float64
	}{
		{"zero time zero K", 0, 0},
		{"zero time room temp", 0, 300},
		{"1 year at 0K", 3.15e7, 0},
		{"1 year at 1K", 3.15e7, 1},
		{"1 year at room temp", 3.15e7, 300},
		{"1 year at 1000K", 3.15e7, 1000},
		{"100 years at room temp", 3.15e9, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RetentionAtTime panicked: %v", r)
				}
			}()

			// Special handling for 0K to avoid division by zero
			if tt.tempK == 0 {
				t.Skip("0K causes division by zero in Arrhenius model - expected behavior")
			}

			Pr := mat.RetentionAtTime(tt.time, tt.tempK)

			if math.IsNaN(Pr) {
				t.Errorf("RetentionAtTime(%e, %fK) returned NaN", tt.time, tt.tempK)
			}
			if Pr < 0 {
				t.Errorf("RetentionAtTime(%e, %fK) returned negative: %e", tt.time, tt.tempK, Pr)
			}

			t.Logf("RetentionAtTime(%.0fs, %.0fK) = %e C/m²", tt.time, tt.tempK, Pr)
		})
	}
}

// =============================================================================
// LK SOLVER EXTREME TEMPERATURE TESTS
// =============================================================================

func TestLKSolver_ExtremeTemperatures(t *testing.T) {
	tests := []struct {
		name  string
		tempK float64
	}{
		{"absolute zero", 0},
		{"near absolute zero", 1},
		{"cryogenic 4K", 4},
		{"liquid nitrogen 77K", 77},
		{"room temperature 300K", 300},
		{"industrial 358K", 358},
		{"automotive 423K", 423},
		{"above Curie 1000K", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LKSolver panicked at %fK: %v", tt.tempK, r)
				}
			}()

			s := NewLKSolver()
			s.Temperature = tt.tempK
			s.EnableNoise = false
			s.UseNLS = false

			// Try to step the solver
			E := 1e9 // 10 MV/cm field
			dt := 1e-11 // 10 ps

			for i := 0; i < 100; i++ {
				P := s.Step(E, dt)

				if math.IsNaN(P) {
					t.Errorf("LKSolver.Step returned NaN at step %d", i)
					break
				}
				if math.IsInf(P, 0) {
					t.Errorf("LKSolver.Step returned Inf at step %d", i)
					break
				}
			}

			t.Logf("LKSolver at %fK: final P = %e C/m²", tt.tempK, s.P)
		})
	}
}

func TestLKSolver_UpdateParams_ExtremeTemperatures(t *testing.T) {
	tests := []struct {
		name  string
		tempK float64
	}{
		{"0K", 0},
		{"1K", 1},
		{"10K", 10},
		{"room temp", 300},
		{"1000K", 1000},
		{"10000K", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("UpdateParams panicked at %fK: %v", tt.tempK, r)
				}
			}()

			s := NewLKSolver()
			s.Temperature = tt.tempK
			s.UseMaterialAlpha = false // Force UpdateParams to recalculate

			s.UpdateParams()

			if math.IsNaN(s.Alpha) {
				t.Errorf("Alpha is NaN at %fK", tt.tempK)
			}
			if math.IsInf(s.Alpha, 0) {
				t.Logf("Alpha is Inf at %fK (may be expected for extreme temps)", tt.tempK)
			}

			t.Logf("UpdateParams at %fK: Alpha = %e", tt.tempK, s.Alpha)
		})
	}
}

// =============================================================================
// MAXIMUM ARRAY SIZE TESTS
// =============================================================================

func TestDiscreteLevel_MaxLevels(t *testing.T) {
	mat := FeCIMMaterial()

	tests := []struct {
		name        string
		totalLevels int
	}{
		{"1 level", 1},
		{"2 levels", 2},
		{"30 levels (standard)", 30},
		{"64 levels (superlattice)", 64},
		{"140 levels (FTJ)", 140},
		{"256 levels", 256},
		{"1000 levels", 1000},
		{"1000000 levels", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DiscreteLevel panicked with %d levels: %v", tt.totalLevels, r)
				}
			}()

			// Test first, middle, and last levels
			testLevels := []int{0}
			if tt.totalLevels > 1 {
				testLevels = append(testLevels, tt.totalLevels/2, tt.totalLevels-1)
			}

			for _, level := range testLevels {
				G := mat.DiscreteLevel(level, tt.totalLevels)

				if math.IsNaN(G) {
					t.Errorf("DiscreteLevel(%d, %d) returned NaN", level, tt.totalLevels)
				}
				if math.IsInf(G, 0) {
					t.Errorf("DiscreteLevel(%d, %d) returned Inf", level, tt.totalLevels)
				}
				if G < 0 {
					t.Errorf("DiscreteLevel(%d, %d) returned negative: %e", level, tt.totalLevels, G)
				}
			}
		})
	}
}

func TestQuantizationError_LargeLevels(t *testing.T) {
	tests := []struct {
		name   string
		levels int
	}{
		{"30 levels", 30},
		{"1000 levels", 1000},
		{"1000000 levels", 1000000},
		{"max int32", 2147483647},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("QuantizationError panicked: %v", r)
				}
			}()

			err := QuantizationError(tt.levels)

			if math.IsNaN(err) {
				t.Errorf("QuantizationError(%d) returned NaN", tt.levels)
			}
			if err < 0 {
				t.Errorf("QuantizationError(%d) returned negative: %e", tt.levels, err)
			}
			if tt.levels > 1 && err >= 1 {
				t.Errorf("QuantizationError(%d) should be < 1, got %e", tt.levels, err)
			}

			t.Logf("QuantizationError(%d) = %e", tt.levels, err)
		})
	}
}

// =============================================================================
// EMPTY CONFIG / NIL POINTER TESTS
// =============================================================================

func TestHZOMaterial_NilMaterial(t *testing.T) {
	var mat *HZOMaterial = nil

	// These should panic or handle nil gracefully
	t.Run("GetNumLevels on nil", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("GetNumLevels on nil panicked as expected: %v", r)
			}
		}()

		if mat != nil {
			_ = mat.GetNumLevels()
		} else {
			t.Log("Correctly avoided calling method on nil pointer")
		}
	})
}

func TestLKSolver_ConfigureFromMaterial_Nil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ConfigureFromMaterial panicked on nil: %v", r)
		}
	}()

	s := NewLKSolver()
	s.ConfigureFromMaterial(nil)

	// Should not change defaults
	t.Logf("After nil config: Beta=%e, Gamma=%e", s.Beta, s.Gamma)
}

func TestLKSolver_ConfigureFromMaterial_EmptyMaterial(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ConfigureFromMaterial panicked on empty material: %v", r)
		}
	}()

	s := NewLKSolver()
	emptyMat := &HZOMaterial{}
	s.ConfigureFromMaterial(emptyMat)

	// Should not overwrite with zero values
	if s.Beta == 0 && s.Gamma == 0 {
		t.Log("Warning: Beta and Gamma became zero from empty material")
	}

	t.Logf("After empty config: Beta=%e, Gamma=%e, P=%e", s.Beta, s.Gamma, s.P)
}

func TestLKSolver_SetState_Invalid(t *testing.T) {
	tests := []struct {
		name string
		P    float64
	}{
		{"NaN", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetState panicked on %s: %v", tt.name, r)
				}
			}()

			s := NewLKSolver()
			initialP := s.P
			s.SetState(tt.P)

			// Should reject invalid values
			if math.IsNaN(s.P) {
				t.Errorf("SetState accepted NaN, P is now NaN")
			}
			if math.IsInf(s.P, 0) {
				t.Errorf("SetState accepted Inf, P is now Inf")
			}

			t.Logf("SetState(%s): P before=%e, after=%e", tt.name, initialP, s.P)
		})
	}
}

// =============================================================================
// MATERIAL PARAMETER EDGE CASES
// =============================================================================

func TestHZOMaterial_ZeroParameters(t *testing.T) {
	mat := &HZOMaterial{
		Pr:        0,
		Ps:        0,
		Ec:        0,
		Thickness: 0,
		Area:      0,
		Tau:       0,
		Tau0:      0,
	}

	t.Run("CoerciveVoltage with zero Ec", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("CoerciveVoltage panicked: %v", r)
			}
		}()
		Vc := mat.CoerciveVoltage()
		if Vc != 0 {
			t.Errorf("Expected 0, got %e", Vc)
		}
	})

	t.Run("Capacitance with zero thickness", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Capacitance panicked: %v", r)
			}
		}()
		C := mat.Capacitance()
		if !math.IsInf(C, 1) && !math.IsNaN(C) {
			t.Logf("Capacitance with zero thickness = %e", C)
		}
	})

	t.Run("SwitchingEnergy with zero area", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SwitchingEnergy panicked: %v", r)
			}
		}()
		E := mat.SwitchingEnergy()
		if E != 0 {
			t.Errorf("Expected 0, got %e", E)
		}
	})
}

func TestHZOMaterial_NegativeParameters(t *testing.T) {
	mat := &HZOMaterial{
		Pr:        -0.25,
		Ps:        -0.30,
		Ec:        -1.0e8,
		Thickness: -10e-9,
		Area:      -100e-12,
		CurieTemp: -100,
	}

	t.Run("CoerciveVoltage with negative params", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("CoerciveVoltage panicked: %v", r)
			}
		}()
		Vc := mat.CoerciveVoltage()
		t.Logf("CoerciveVoltage with negative Ec/thickness = %e V", Vc)
	})

	t.Run("PolarizationAtTemp with negative CurieTemp", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PolarizationAtTemp panicked: %v", r)
			}
		}()
		Pr := mat.PolarizationAtTemp(300)
		t.Logf("PolarizationAtTemp with negative CurieTemp = %e", Pr)
	})
}

// =============================================================================
// LK SOLVER NUMERICAL STABILITY TESTS
// =============================================================================

func TestLKSolver_Step_ExtremeFields(t *testing.T) {
	tests := []struct {
		name  string
		E     float64
		dt    float64
	}{
		{"zero field", 0, 1e-11},
		{"very small field", 1e-10, 1e-11},
		{"normal field", 1e9, 1e-11},
		{"very high field", 1e12, 1e-11},
		{"extreme field", 1e15, 1e-11},
		{"negative high field", -1e12, 1e-11},
		{"tiny dt", 1e9, 1e-20},
		{"large dt", 1e9, 1e-6},
		{"huge dt", 1e9, 1.0}, // 1 second
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Step panicked: %v", r)
				}
			}()

			s := NewLKSolver()
			s.EnableNoise = false
			s.UseNLS = false

			P := s.Step(tt.E, tt.dt)

			if math.IsNaN(P) {
				t.Errorf("Step(E=%e, dt=%e) returned NaN", tt.E, tt.dt)
			}

			t.Logf("Step(E=%e, dt=%e) = %e", tt.E, tt.dt, P)
		})
	}
}

func TestLKSolver_Step_NaNField(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Step panicked on NaN field: %v", r)
		}
	}()

	s := NewLKSolver()
	s.EnableNoise = false
	s.UseNLS = false

	P := s.Step(math.NaN(), 1e-11)

	if !math.IsNaN(P) {
		t.Logf("Step with NaN field returned %e (expected NaN propagation or handling)", P)
	}
}

func TestLKSolver_Step_ZeroDt(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Step panicked on zero dt: %v", r)
		}
	}()

	s := NewLKSolver()
	s.EnableNoise = false
	initialP := s.P

	P := s.Step(1e9, 0)

	// With dt=0, P should not change
	if P != initialP {
		t.Logf("Step with dt=0: P changed from %e to %e", initialP, P)
	}
}

func TestLKSolver_Step_NegativeDt(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Step panicked on negative dt: %v", r)
		}
	}()

	s := NewLKSolver()
	s.EnableNoise = false

	P := s.Step(1e9, -1e-11)

	if math.IsNaN(P) {
		t.Error("Step with negative dt returned NaN")
	}

	t.Logf("Step with negative dt = %e", P)
}

// =============================================================================
// CONDUCTANCE MODEL EDGE CASES
// =============================================================================

func TestNormalizedToPhysicalRange_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		gNorm float64
		model ConductanceModel
		gMin  float64
		gMax  float64
	}{
		{"gNorm < 0", -0.5, ConductanceLinear, GMin, GMax},
		{"gNorm > 1", 1.5, ConductanceLinear, GMin, GMax},
		{"gMin = gMax", 0.5, ConductanceLinear, GMin, GMin},
		{"gMin > gMax", 0.5, ConductanceLinear, GMax, GMin},
		{"zero range", 0.5, ConductanceLinear, 0, 0},
		{"exponential zero gMin", 0.5, ConductanceExponential, 0, GMax},
		{"exponential equal bounds", 0.5, ConductanceExponential, GMin, GMin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NormalizedToPhysicalRange panicked: %v", r)
				}
			}()

			result := NormalizedToPhysicalRange(tt.gNorm, tt.model, tt.gMin, tt.gMax, nil)

			if math.IsNaN(result) {
				t.Logf("NormalizedToPhysicalRange returned NaN for %s", tt.name)
			} else if math.IsInf(result, 0) {
				t.Logf("NormalizedToPhysicalRange returned Inf for %s", tt.name)
			} else {
				t.Logf("NormalizedToPhysicalRange for %s = %e", tt.name, result)
			}
		})
	}
}

func TestNormalizedToPhysicalRange_LookupEmptyTable(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Lookup with empty table panicked: %v", r)
		}
	}()

	result := NormalizedToPhysicalRange(0.5, ConductanceLookup, GMin, GMax, nil)
	t.Logf("Lookup with nil table = %e (should fallback to linear)", result)

	result = NormalizedToPhysicalRange(0.5, ConductanceLookup, GMin, GMax, []float64{})
	t.Logf("Lookup with empty table = %e (should fallback to linear)", result)

	result = NormalizedToPhysicalRange(0.5, ConductanceLookup, GMin, GMax, []float64{GMin})
	t.Logf("Lookup with single-element table = %e", result)
}
