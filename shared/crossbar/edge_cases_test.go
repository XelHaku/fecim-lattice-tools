// Package crossbar implements ferroelectric crossbar array simulation.
// This file contains edge case tests for numerical robustness.
package crossbar

import (
	"math"
	"testing"
)

// =============================================================================
// ZERO ARRAY TESTS
// =============================================================================

func TestNewArray_ZeroDimensions(t *testing.T) {
	tests := []struct {
		name    string
		rows    int
		cols    int
		wantErr bool
	}{
		{"zero rows", 0, 10, true},
		{"zero cols", 10, 0, true},
		{"both zero", 0, 0, true},
		{"negative rows", -1, 10, true},
		{"negative cols", 10, -1, true},
		{"both negative", -5, -5, true},
		{"valid 1x1", 1, 1, false},
		{"valid standard", 128, 128, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Rows:       tt.rows,
				Cols:       tt.cols,
				NoiseLevel: 0.01,
				ADCBits:    8,
				DACBits:    8,
			}

			arr, err := NewArray(cfg)

			if tt.wantErr && err == nil {
				t.Errorf("NewArray should error for %dx%d", tt.rows, tt.cols)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("NewArray failed for valid %dx%d: %v", tt.rows, tt.cols, err)
			}
			if !tt.wantErr && arr == nil {
				t.Error("NewArray returned nil array without error")
			}
		})
	}
}

func TestNewParasiticSolver_ZeroDimensions(t *testing.T) {
	tests := []struct {
		name    string
		rows    int
		cols    int
		wantErr bool
	}{
		{"zero rows", 0, 10, true},
		{"zero cols", 10, 0, true},
		{"both zero", 0, 0, true},
		{"negative rows", -1, 10, true},
		{"negative cols", 10, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewParasiticSolver(tt.rows, tt.cols, nil)

			if tt.wantErr && err == nil {
				t.Errorf("NewParasiticSolver should error for %dx%d", tt.rows, tt.cols)
			}
		})
	}
}

// =============================================================================
// NaN INPUT TESTS
// =============================================================================

func TestParasiticSolver_SetConductances_NaN(t *testing.T) {
	solver, err := NewParasiticSolver(4, 4, nil)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	// Create conductance matrix with NaN values
	g := make([][]float64, 4)
	for i := range g {
		g[i] = make([]float64, 4)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	g[2][2] = math.NaN() // Inject NaN

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetConductances panicked on NaN: %v", r)
		}
	}()

	solver.SetConductances(g)
	t.Log("SetConductances accepted NaN input (may propagate to output)")
}

func TestParasiticSolver_SolveMVM_NaNInput(t *testing.T) {
	solver, err := NewParasiticSolver(4, 4, nil)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	g := make([][]float64, 4)
	for i := range g {
		g[i] = make([]float64, 4)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.1, 0.1)

	// Input with NaN
	input := []float64{1.0, math.NaN(), 1.0, 1.0}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SolveMVM panicked on NaN input: %v", r)
		}
	}()

	result, err := solver.SolveMVM(input)
	if err != nil {
		t.Logf("SolveMVM returned error on NaN input: %v", err)
	}
	if result != nil {
		t.Logf("SolveMVM result with NaN input: converged=%v, iterations=%d",
			result.Converged, result.Iterations)
	}
}

func TestParasiticSolver_SetParasitics_NaN(t *testing.T) {
	solver, err := NewParasiticSolver(4, 4, nil)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetParasitics panicked on NaN: %v", r)
		}
	}()

	solver.SetParasitics(math.NaN(), 0.1)
	t.Log("SetParasitics accepted NaN wordline resistance")

	solver.SetParasitics(0.1, math.NaN())
	t.Log("SetParasitics accepted NaN bitline resistance")
}

// =============================================================================
// EXTREME TEMPERATURE TESTS
// =============================================================================

func TestTemperatureEffects_ExtremeTemperatures(t *testing.T) {
	tests := []struct {
		name  string
		tempK float64
	}{
		{"absolute zero", 0},
		{"near absolute zero", 0.001},
		{"1K", 1},
		{"liquid helium 4K", 4},
		{"liquid nitrogen 77K", 77},
		{"room temperature 300K", 300},
		{"industrial 358K", 358},
		{"automotive 423K", 423},
		{"extreme high 1000K", 1000},
		{"beyond Curie 2000K", 2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewTemperatureEffects panicked at %fK: %v", tt.tempK, r)
				}
			}()

			te := NewTemperatureEffects(tt.tempK)

			// Test wire resistance adjustment
			R := te.AdjustedWireResistance(100.0)
			if math.IsNaN(R) {
				t.Errorf("AdjustedWireResistance returned NaN at %fK", tt.tempK)
			}
			t.Logf("Wire resistance at %fK: %f Ω (base 100Ω)", tt.tempK, R)

			// Test conductance range adjustment
			gMin, gMax := te.AdjustedConductanceRange(GMin, GMax)
			if math.IsNaN(gMin) || math.IsNaN(gMax) {
				t.Errorf("AdjustedConductanceRange returned NaN at %fK", tt.tempK)
			}
			t.Logf("Conductance range at %fK: [%e, %e] S", tt.tempK, gMin, gMax)

			// Test drift rate adjustment (skip 0K to avoid division by zero)
			if tt.tempK > 0 {
				drift := te.AdjustedDriftRate(0.01)
				if math.IsNaN(drift) {
					t.Errorf("AdjustedDriftRate returned NaN at %fK", tt.tempK)
				}
				t.Logf("Drift rate at %fK: %e (base 0.01)", tt.tempK, drift)
			}

			// Test retention adjustment (skip 0K)
			if tt.tempK > 0 {
				ret := te.AdjustedRetention()
				if math.IsNaN(ret) || ret < 0 {
					t.Errorf("AdjustedRetention returned invalid value at %fK: %e", tt.tempK, ret)
				}
				t.Logf("Retention factor at %fK: %e", tt.tempK, ret)
			}

			// Test noise adjustment
			noise := te.AdjustedNoise()
			if math.IsNaN(noise) {
				t.Errorf("AdjustedNoise returned NaN at %fK", tt.tempK)
			}
			t.Logf("Noise factor at %fK: %f", tt.tempK, noise)
		})
	}
}

func TestTemperatureEffects_NegativeTemperature(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewTemperatureEffects panicked on negative temp: %v", r)
		}
	}()

	te := NewTemperatureEffects(-100)

	// Should clamp to 300K or handle gracefully
	if te.AmbientK < 0 {
		t.Errorf("TemperatureEffects accepted negative temperature: %f", te.AmbientK)
	}
	t.Logf("TemperatureEffects(-100) → AmbientK = %f", te.AmbientK)
}

func TestRetentionCurve_ExtremeTemperatures(t *testing.T) {
	curve := DefaultRetentionCurve()

	tests := []struct {
		name  string
		tempK float64
	}{
		{"near absolute zero 1K", 1},
		{"4K", 4},
		{"77K", 77},
		{"room temp 300K", 300},
		{"industrial 358K", 358},
		{"automotive 423K", 423},
		{"extreme 1000K", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RetentionAt panicked at %fK: %v", tt.tempK, r)
				}
			}()

			ret := curve.RetentionAt(tt.tempK)

			if math.IsNaN(ret) {
				t.Errorf("RetentionAt returned NaN at %fK", tt.tempK)
			}
			if ret < 0 {
				t.Errorf("RetentionAt returned negative at %fK: %e", tt.tempK, ret)
			}

			years := curve.RetentionYearsAt(tt.tempK)
			t.Logf("Retention at %fK: %e seconds (%.2f years)", tt.tempK, ret, years)
		})
	}
}

func TestThermalPhysicsModel_ExtremeTemperatures(t *testing.T) {
	model := DefaultHZOThermalModel()

	tests := []struct {
		name  string
		tempK float64
	}{
		{"near absolute zero", 1},
		{"cryogenic", 77},
		{"room temp", 300},
		{"at Curie temp", model.CurieTempK},
		{"above Curie", model.CurieTempK + 100},
		{"extreme high", 1500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Thermal model panicked at %fK: %v", tt.tempK, r)
				}
			}()

			Pr := model.PrAtTemperature(tt.tempK)
			Ec := model.EcAtTemperature(tt.tempK)
			levels := model.EffectiveLevelsAtTemperature(tt.tempK, 30)

			if math.IsNaN(Pr) {
				t.Errorf("PrAtTemperature returned NaN at %fK", tt.tempK)
			}
			if math.IsNaN(Ec) {
				t.Errorf("EcAtTemperature returned NaN at %fK", tt.tempK)
			}

			t.Logf("At %fK: Pr=%e, Ec=%e, levels=%d", tt.tempK, Pr, Ec, levels)
		})
	}
}

// =============================================================================
// MAXIMUM ARRAY SIZE TESTS
// =============================================================================

func TestNewArray_LargeArrays(t *testing.T) {
	tests := []struct {
		name    string
		rows    int
		cols    int
		wantErr bool
	}{
		{"128x128 standard", 128, 128, false},
		{"256x256", 256, 256, false},
		{"512x512", 512, 512, false},
		{"1024x1024", 1024, 1024, false},
		// Larger arrays may OOM, so we don't test them by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Rows:       tt.rows,
				Cols:       tt.cols,
				NoiseLevel: 0.01,
				ADCBits:    8,
				DACBits:    8,
			}

			arr, err := NewArray(cfg)

			if tt.wantErr && err == nil {
				t.Errorf("NewArray should error for %dx%d", tt.rows, tt.cols)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("NewArray failed for %dx%d: %v", tt.rows, tt.cols, err)
			}
			if arr != nil {
				t.Logf("Created %dx%d array successfully", tt.rows, tt.cols)
			}
		})
	}
}

func TestNewParasiticSolver_LargeArrays(t *testing.T) {
	tests := []struct {
		name string
		rows int
		cols int
	}{
		{"128x128", 128, 128},
		{"256x256", 256, 256},
		{"512x512", 512, 512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solver, err := NewParasiticSolver(tt.rows, tt.cols, nil)
			if err != nil {
				t.Errorf("NewParasiticSolver failed for %dx%d: %v", tt.rows, tt.cols, err)
			}
			if solver != nil {
				t.Logf("Created %dx%d solver successfully", tt.rows, tt.cols)
			}
		})
	}
}

// =============================================================================
// EMPTY CONFIG TESTS
// =============================================================================

func TestNewArray_EmptyConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewArray panicked on sparse config: %v", r)
		}
	}()

	// Minimal config - only required fields
	cfg := &Config{
		Rows: 8,
		Cols: 8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Logf("NewArray with minimal config: %v", err)
	}
	if arr != nil {
		t.Log("Created array with minimal config")
	}
}

func TestNewArray_NilSubconfigs(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewArray panicked with nil subconfigs: %v", r)
		}
	}()

	cfg := &Config{
		Rows:             8,
		Cols:             8,
		NoiseLevel:       0.01,
		ADCBits:          6,
		DACBits:          4,
		Endurance:        nil, // Explicitly nil
		ProcessVariation: nil,
		HalfSelect:       nil,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Errorf("NewArray with nil subconfigs failed: %v", err)
	}
	if arr != nil {
		t.Log("Created array with nil subconfigs")
	}
}

func TestNewParasiticSolver_NilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewParasiticSolver panicked with nil config: %v", r)
		}
	}()

	solver, err := NewParasiticSolver(8, 8, nil)
	if err != nil {
		t.Errorf("NewParasiticSolver with nil config failed: %v", err)
	}
	if solver != nil {
		t.Log("Created solver with nil config (uses defaults)")
	}
}

func TestDefaultConfigs(t *testing.T) {
	t.Run("DefaultEnduranceConfig", func(t *testing.T) {
		cfg := DefaultEnduranceConfig()
		if cfg == nil {
			t.Error("DefaultEnduranceConfig returned nil")
		}
		if cfg.FailureThreshold <= 0 {
			t.Error("FailureThreshold should be positive")
		}
	})

	t.Run("DefaultProcessVariationConfig", func(t *testing.T) {
		cfg := DefaultProcessVariationConfig()
		if cfg == nil {
			t.Error("DefaultProcessVariationConfig returned nil")
		}
		if cfg.DeviceSigma < 0 {
			t.Error("DeviceSigma should be non-negative")
		}
	})

	t.Run("DefaultHalfSelectConfig", func(t *testing.T) {
		cfg := DefaultHalfSelectConfig()
		if cfg == nil {
			t.Error("DefaultHalfSelectConfig returned nil")
		}
		if cfg.DisturbRate < 0 {
			t.Error("DisturbRate should be non-negative")
		}
	})
}

// =============================================================================
// SOR SOLVER EDGE CASES
// =============================================================================

func TestSORSolver_ZeroParasitics(t *testing.T) {
	solver, err := NewParasiticSolver(8, 8, nil)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	g := make([][]float64, 8)
	for i := range g {
		g[i] = make([]float64, 8)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0, 0) // Zero parasitics

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	result, err := solver.SolveMVM(input)
	if err != nil {
		t.Errorf("SolveMVM with zero parasitics failed: %v", err)
	}
	if result != nil && !result.Converged {
		t.Error("Should converge immediately with zero parasitics")
	}
}

func TestSORSolver_ExtremeParasitics(t *testing.T) {
	tests := []struct {
		name string
		Rw   float64
		Rb   float64
	}{
		{"very low", 1e-10, 1e-10},
		{"low", 0.01, 0.01},
		{"medium", 0.5, 0.5},
		{"high", 2.0, 2.0},
		{"very high", 10.0, 10.0},
		{"extreme", 100.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SORConfig{
				MaxIterations: 200,
				Tolerance:     1e-4,
				OmegaInitial:  0.6,
				OmegaMin:      0.01,
				OmegaDecay:    0.95,
				AdaptiveOmega: true,
			}

			solver, _ := NewParasiticSolver(16, 16, config)

			g := make([][]float64, 16)
			for i := range g {
				g[i] = make([]float64, 16)
				for j := range g[i] {
					g[i][j] = 0.5
				}
			}
			solver.SetConductances(g)
			solver.SetParasitics(tt.Rw, tt.Rb)

			input := make([]float64, 16)
			for i := range input {
				input[i] = 1.0
			}

			result, _ := solver.SolveMVMWithFallback(input)

			if result == nil {
				t.Error("SolveMVMWithFallback returned nil result")
			} else {
				t.Logf("Rp=%f: converged=%v, iterations=%d, maxErr=%e",
					tt.Rw, result.Converged, result.Iterations, result.MaxError)
			}
		})
	}
}

func TestSORSolver_ZeroConductances(t *testing.T) {
	solver, _ := NewParasiticSolver(4, 4, nil)

	g := make([][]float64, 4)
	for i := range g {
		g[i] = make([]float64, 4)
		// All zeros
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.1, 0.1)

	input := []float64{1.0, 1.0, 1.0, 1.0}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SolveMVM panicked with zero conductances: %v", r)
		}
	}()

	result, err := solver.SolveMVM(input)
	if err != nil {
		t.Logf("SolveMVM with zero conductances: %v", err)
	}
	if result != nil {
		t.Logf("Result with zero conductances: %+v", result)
	}
}

func TestSORSolver_ZeroInput(t *testing.T) {
	solver, _ := NewParasiticSolver(4, 4, nil)

	g := make([][]float64, 4)
	for i := range g {
		g[i] = make([]float64, 4)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.1, 0.1)

	// Zero input voltages
	input := []float64{0, 0, 0, 0}

	result, err := solver.SolveMVM(input)
	if err != nil {
		t.Errorf("SolveMVM with zero input failed: %v", err)
	}

	// Output should be zero or near zero
	if result != nil {
		for i, out := range result.OutputCurrents {
			if math.Abs(out) > 1e-10 {
				t.Errorf("Output[%d] = %e, expected ~0", i, out)
			}
		}
	}
}

func TestSORSolver_EmptyInput(t *testing.T) {
	solver, _ := NewParasiticSolver(4, 4, nil)

	g := make([][]float64, 4)
	for i := range g {
		g[i] = make([]float64, 4)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("SolveMVM panicked on empty input: %v (may be expected)", r)
		}
	}()

	// Empty slice
	_, err := solver.SolveMVM([]float64{})
	if err != nil {
		t.Logf("SolveMVM with empty input: %v", err)
	}

	// Nil slice
	_, err = solver.SolveMVM(nil)
	if err != nil {
		t.Logf("SolveMVM with nil input: %v", err)
	}
}

func TestSORSolver_WrongSizeInput(t *testing.T) {
	solver, _ := NewParasiticSolver(4, 4, nil)

	g := make([][]float64, 4)
	for i := range g {
		g[i] = make([]float64, 4)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)

	// Wrong size input (too small)
	_, err := solver.SolveMVM([]float64{1.0, 1.0})
	if err == nil {
		t.Error("SolveMVM should error with wrong-size input")
	} else {
		t.Logf("SolveMVM with wrong size: %v", err)
	}

	// Wrong size input (too large)
	_, err = solver.SolveMVM([]float64{1, 1, 1, 1, 1, 1, 1, 1})
	if err == nil {
		t.Error("SolveMVM should error with oversized input")
	}
}

// =============================================================================
// CONDUCTANCE EDGE CASES
// =============================================================================

func TestConductanceConstants(t *testing.T) {
	// Verify conductance constants are sensible
	if GMin <= 0 {
		t.Error("GMin should be positive")
	}
	if GMax <= GMin {
		t.Error("GMax should be greater than GMin")
	}
	// GRatio is defined in physics package
	expectedRatio := GMax / GMin
	if expectedRatio <= 1 {
		t.Errorf("GMax/GMin ratio (%e) should be > 1", expectedRatio)
	}
	t.Logf("Conductance range: GMin=%e, GMax=%e, ratio=%e", GMin, GMax, expectedRatio)
}

func TestComputeAccuracyLoss(t *testing.T) {
	tests := []struct {
		name   string
		ideal  []float64
		actual []float64
	}{
		{"identical", []float64{1, 2, 3}, []float64{1, 2, 3}},
		{"small diff", []float64{1, 2, 3}, []float64{1.01, 2.01, 3.01}},
		{"large diff", []float64{1, 2, 3}, []float64{2, 4, 6}},
		{"empty", []float64{}, []float64{}},
		{"single element", []float64{1}, []float64{1.5}},
		{"with zeros", []float64{0, 1, 0}, []float64{0, 1, 0}},
		{"with NaN", []float64{1, math.NaN(), 3}, []float64{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ComputeAccuracyLoss panicked: %v", r)
				}
			}()

			loss := ComputeAccuracyLoss(tt.ideal, tt.actual)
			t.Logf("AccuracyLoss(%v, %v) = %e", tt.ideal, tt.actual, loss)
		})
	}
}

func TestComputeAccuracyLoss_MismatchedLengths(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("ComputeAccuracyLoss panicked on mismatched lengths: %v", r)
		}
	}()

	loss := ComputeAccuracyLoss([]float64{1, 2, 3}, []float64{1, 2})
	if !math.IsNaN(loss) && !math.IsInf(loss, 0) {
		t.Logf("ComputeAccuracyLoss with mismatched lengths = %e", loss)
	}
}

// =============================================================================
// ARRAY OPERATIONS EDGE CASES
// =============================================================================

func TestArray_ProgramWeight_EdgeCases(t *testing.T) {
	cfg := &Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}

	tests := []struct {
		name      string
		row       int
		col       int
		weight    float64
		expectErr bool
	}{
		{"valid write", 0, 0, 0.5, false},
		{"write zero", 1, 1, 0.0, false},
		{"write one", 2, 2, 1.0, false},
		{"out of bounds row", 10, 0, 0.5, true},
		{"out of bounds col", 0, 10, 0.5, true},
		{"negative row", -1, 0, 0.5, true},
		{"negative col", 0, -1, 0.5, true},
		{"NaN weight", 0, 0, math.NaN(), false}, // May or may not error
		{"Inf weight", 0, 0, math.Inf(1), false},
		{"negative weight", 0, 0, -0.5, false}, // Should clamp or error
		{"weight > 1", 0, 0, 1.5, false},       // Should clamp or error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectErr {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			err := arr.ProgramWeight(tt.row, tt.col, tt.weight)
			if tt.expectErr && err == nil {
				t.Logf("ProgramWeight(%d, %d, %f) should have errored", tt.row, tt.col, tt.weight)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("ProgramWeight(%d, %d, %f) unexpected error: %v", tt.row, tt.col, tt.weight, err)
			}
		})
	}
}
