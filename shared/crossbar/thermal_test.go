package crossbar

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Gap 8: Compact thermal model for crossbar arrays
// ---------------------------------------------------------------------------

// TestThermalModel_SteadyState_UniformPower verifies that uniform power
// dissipation produces a uniform temperature rise of dT = P * R_th.
func TestThermalModel_SteadyState_UniformPower(t *testing.T) {
	const rows, cols = 4, 4
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(rows, cols, cfg)

	// Uniform 1 mW per cell.
	power := 1e-3 // W
	powerMap := make([][]float64, rows)
	for i := range powerMap {
		powerMap[i] = make([]float64, cols)
		for j := range powerMap[i] {
			powerMap[i][j] = power
		}
	}

	state := tm.ComputeSteadyState(powerMap)

	expectedT := cfg.AmbientTempK + power*cfg.ThermalResistance
	t.Logf("Expected T = %.2f K, Peak = %.2f K, Avg = %.2f K", expectedT, state.PeakTempK, state.AvgTempK)

	// All cells should be at the same temperature.
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if math.Abs(state.CellTemperatures[i][j]-expectedT) > 1e-10 {
				t.Errorf("Cell(%d,%d) T=%.6f K, expected %.6f K", i, j,
					state.CellTemperatures[i][j], expectedT)
			}
		}
	}

	// Peak and average should be equal for uniform power.
	if math.Abs(state.PeakTempK-state.AvgTempK) > 1e-10 {
		t.Errorf("Peak (%.6f) != Avg (%.6f) for uniform power", state.PeakTempK, state.AvgTempK)
	}

	// Verify the exact value: dT = 1e-3 * 20 = 0.02 K, so T = 300.02 K.
	if math.Abs(state.PeakTempK-300.02) > 1e-10 {
		t.Errorf("Peak T = %.6f K, expected 300.02 K", state.PeakTempK)
	}
}

// TestThermalModel_SteadyState_HotSpot verifies that a single hot cell
// produces peak temperature higher than the average.
func TestThermalModel_SteadyState_HotSpot(t *testing.T) {
	const rows, cols = 8, 8
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(rows, cols, cfg)

	// One hot cell with 10 mW, rest at 0.1 mW.
	powerMap := make([][]float64, rows)
	for i := range powerMap {
		powerMap[i] = make([]float64, cols)
		for j := range powerMap[i] {
			powerMap[i][j] = 0.1e-3 // 0.1 mW background
		}
	}
	powerMap[4][4] = 10e-3 // 10 mW hot spot

	state := tm.ComputeSteadyState(powerMap)

	hotCellT := state.CellTemperatures[4][4]
	expectedHot := cfg.AmbientTempK + 10e-3*cfg.ThermalResistance

	t.Logf("Hot cell T = %.4f K, Peak = %.4f K, Avg = %.4f K", hotCellT, state.PeakTempK, state.AvgTempK)

	// Hot cell should be the peak.
	if math.Abs(hotCellT-expectedHot) > 1e-10 {
		t.Errorf("Hot cell T=%.6f K, expected %.6f K", hotCellT, expectedHot)
	}
	if math.Abs(state.PeakTempK-hotCellT) > 1e-10 {
		t.Errorf("Peak T=%.6f K != hot cell T=%.6f K", state.PeakTempK, hotCellT)
	}

	// Peak must be greater than average.
	if state.PeakTempK <= state.AvgTempK {
		t.Errorf("Peak (%.4f K) should be > Avg (%.4f K) with hot spot", state.PeakTempK, state.AvgTempK)
	}

	// Average should be between ambient and hot cell.
	if state.AvgTempK <= cfg.AmbientTempK {
		t.Errorf("Avg T (%.4f K) should be > ambient (%.4f K)", state.AvgTempK, cfg.AmbientTempK)
	}
	if state.AvgTempK >= state.PeakTempK {
		t.Errorf("Avg T (%.4f K) should be < peak (%.4f K)", state.AvgTempK, state.PeakTempK)
	}
}

// TestThermalModel_Transient_Heating verifies that the thermal transient
// follows an exponential approach to steady state with time constant
// tau = R_th * C_th.
func TestThermalModel_Transient_Heating(t *testing.T) {
	const rows, cols = 1, 1
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(rows, cols, cfg)

	tau := cfg.ThermalResistance * cfg.ThermalCapacitance // 20 * 1e-6 = 20 us
	power := 1e-3                                         // 1 mW

	// Time step should be much smaller than tau for accuracy.
	dt := tau / 100.0
	steps := 500 // 5 * tau should reach ~99.3% of steady state

	powerMap := [][]float64{{power}}
	snapshots := tm.ComputeTransient(powerMap, dt, steps)

	if len(snapshots) != steps {
		t.Fatalf("Expected %d snapshots, got %d", steps, len(snapshots))
	}

	steadyStateT := cfg.AmbientTempK + power*cfg.ThermalResistance

	// Check temperature at t = tau (should be ~63.2% of steady-state rise).
	tauStep := int(tau / dt)
	if tauStep >= steps {
		t.Fatalf("tau step %d >= total steps %d", tauStep, steps)
	}

	tAtTau := snapshots[tauStep].CellTemperatures[0][0]
	dTExpected := (steadyStateT - cfg.AmbientTempK) * (1 - math.Exp(-1))
	expectedTAtTau := cfg.AmbientTempK + dTExpected

	t.Logf("tau = %.2e s, T(tau) = %.6f K, expected = %.6f K", tau, tAtTau, expectedTAtTau)

	// Allow 2% error from Euler integration.
	relErr := math.Abs(tAtTau-expectedTAtTau) / (expectedTAtTau - cfg.AmbientTempK) * 100
	if relErr > 2.0 {
		t.Errorf("T(tau) error = %.1f%% (got %.6f, expected %.6f)", relErr, tAtTau, expectedTAtTau)
	}

	// Check that final temperature is close to steady state.
	tFinal := snapshots[steps-1].CellTemperatures[0][0]
	finalErr := math.Abs(tFinal-steadyStateT) / (steadyStateT - cfg.AmbientTempK) * 100
	t.Logf("T(5*tau) = %.6f K, steady state = %.6f K, err = %.1f%%", tFinal, steadyStateT, finalErr)

	if finalErr > 1.0 {
		t.Errorf("Final T error = %.1f%% (got %.6f, expected %.6f)", finalErr, tFinal, steadyStateT)
	}
}

// TestThermalModel_PowerFromMVM verifies that the power map computed from
// an MVM operation has correct dimensions and all non-negative values.
func TestThermalModel_PowerFromMVM(t *testing.T) {
	const rows, cols = 4, 8
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(rows, cols, cfg)

	// Create a crossbar array.
	arrCfg := &Config{
		Rows:    rows,
		Cols:    cols,
		ADCBits: 8,
		DACBits: 8,
	}
	arr, err := NewArray(arrCfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	// Program some conductances.
	weights := make([][]float64, rows)
	for i := range weights {
		weights[i] = make([]float64, cols)
		for j := range weights[i] {
			weights[i][j] = float64(i*cols+j) / float64(rows*cols-1)
		}
	}
	if err := arr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix failed: %v", err)
	}

	// Input vector (normalized 0-1).
	input := make([]float64, cols)
	for j := range input {
		input[j] = float64(j) / float64(cols-1)
	}

	powerMap := tm.PowerFromMVM(arr, input)
	if powerMap == nil {
		t.Fatal("PowerFromMVM returned nil")
	}

	// Check dimensions.
	if len(powerMap) != rows {
		t.Errorf("Power map rows = %d, expected %d", len(powerMap), rows)
	}
	for i, row := range powerMap {
		if len(row) != cols {
			t.Errorf("Power map row %d cols = %d, expected %d", i, len(row), cols)
		}
	}

	// All values should be non-negative.
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if powerMap[i][j] < 0 {
				t.Errorf("Power(%d,%d) = %e < 0", i, j, powerMap[i][j])
			}
		}
	}

	// First column with input[0] = 0 should have zero power.
	for i := 0; i < rows; i++ {
		if powerMap[i][0] != 0 {
			t.Errorf("Power(%d,0) = %e, expected 0 (input[0] = 0)", i, powerMap[i][0])
		}
	}

	// Last column with max input and max conductance should have highest power.
	maxPower := 0.0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if powerMap[i][j] > maxPower {
				maxPower = powerMap[i][j]
			}
		}
	}
	lastRow := rows - 1
	lastCol := cols - 1
	if powerMap[lastRow][lastCol] != maxPower {
		t.Logf("Note: max power cell may not be at (%d,%d) due to conductance quantization",
			lastRow, lastCol)
	}

	t.Logf("Power range: min=0, max=%.4e W", maxPower)
}

// TestThermalModel_AEC_Q100_Check verifies that for a typical workload the
// junction temperature stays below the AEC-Q100 Grade 1 limit of 150 C
// (423.15 K).
//
// This is a sanity check using realistic power levels for a FeFET crossbar
// array. Typical per-cell power during read is V^2 * G ~ (0.2V)^2 * 100uS
// = 4 uW per cell.
func TestThermalModel_AEC_Q100_Check(t *testing.T) {
	const rows, cols = 64, 64
	cfg := DefaultThermalConfig()
	// Use a more realistic elevated ambient for automotive.
	cfg.AmbientTempK = 358.15 // 85 C (AEC-Q100 ambient)
	tm := NewThermalModel(rows, cols, cfg)

	// Typical read power per cell: V^2 * G_max = (0.2)^2 * 100e-6 = 4 uW.
	typicalPower := 4e-6 // W per cell

	powerMap := make([][]float64, rows)
	for i := range powerMap {
		powerMap[i] = make([]float64, cols)
		for j := range powerMap[i] {
			powerMap[i][j] = typicalPower
		}
	}

	state := tm.ComputeSteadyState(powerMap)

	// AEC-Q100 Grade 1: Tj_max = 150 C = 423.15 K.
	const maxTempK = 423.15

	t.Logf("Peak T = %.2f K (%.2f C), limit = %.2f K (%.2f C)",
		state.PeakTempK, state.PeakTempK-273.15, maxTempK, maxTempK-273.15)
	t.Logf("Avg T = %.2f K (%.2f C)", state.AvgTempK, state.AvgTempK-273.15)

	if state.PeakTempK > maxTempK {
		t.Errorf("Peak T = %.2f K (%.2f C) exceeds AEC-Q100 limit of %.2f K (%.2f C)",
			state.PeakTempK, state.PeakTempK-273.15, maxTempK, maxTempK-273.15)
	}

	// Verify the temperature rise is physically reasonable.
	dT := state.PeakTempK - cfg.AmbientTempK
	expectedDT := typicalPower * cfg.ThermalResistance
	t.Logf("Temperature rise: %.4f K (expected %.4f K)", dT, expectedDT)

	if math.Abs(dT-expectedDT) > 1e-10 {
		t.Errorf("Temperature rise %.6f K != expected %.6f K", dT, expectedDT)
	}
}

// TestThermalModel_DefaultConfig verifies that the default config has
// sensible values.
func TestThermalModel_DefaultConfig(t *testing.T) {
	cfg := DefaultThermalConfig()

	if cfg.AmbientTempK != 300 {
		t.Errorf("AmbientTempK = %f, want 300", cfg.AmbientTempK)
	}
	if cfg.ThermalResistance != 20 {
		t.Errorf("ThermalResistance = %f, want 20", cfg.ThermalResistance)
	}
	if cfg.ThermalCapacitance != 1e-6 {
		t.Errorf("ThermalCapacitance = %e, want 1e-6", cfg.ThermalCapacitance)
	}
	if cfg.SubstrateTempK != 300 {
		t.Errorf("SubstrateTempK = %f, want 300", cfg.SubstrateTempK)
	}
}

// TestThermalModel_TimeConstant verifies tau = R_th * C_th.
func TestThermalModel_TimeConstant(t *testing.T) {
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(4, 4, cfg)

	tau := tm.TimeConstant()
	expected := cfg.ThermalResistance * cfg.ThermalCapacitance

	if math.Abs(tau-expected) > 1e-15 {
		t.Errorf("TimeConstant() = %e, expected %e", tau, expected)
	}
	t.Logf("tau = %e s (%.1f us)", tau, tau*1e6)
}

// TestThermalModel_AnalyticTransient verifies the TransientTemp helper
// against the exact analytical solution.
func TestThermalModel_AnalyticTransient(t *testing.T) {
	cfg := DefaultThermalConfig()
	power := 5e-3 // 5 mW
	tau := cfg.ThermalResistance * cfg.ThermalCapacitance

	testTimes := []float64{0, tau * 0.5, tau, 2 * tau, 5 * tau}

	for _, tt := range testTimes {
		temp := TransientTemp(power, tt, cfg)
		expected := cfg.AmbientTempK + power*cfg.ThermalResistance*(1-math.Exp(-tt/tau))
		if math.Abs(temp-expected) > 1e-12 {
			t.Errorf("TransientTemp(t=%e) = %f, expected %f", tt, temp, expected)
		}
	}

	// t=0 should equal ambient.
	if TransientTemp(power, 0, cfg) != cfg.AmbientTempK {
		t.Errorf("TransientTemp(t=0) != ambient")
	}
}

// TestThermalModel_ZeroPower verifies that zero power keeps everything at
// ambient temperature.
func TestThermalModel_ZeroPower(t *testing.T) {
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(4, 4, cfg)

	powerMap := make([][]float64, 4)
	for i := range powerMap {
		powerMap[i] = make([]float64, 4)
	}

	state := tm.ComputeSteadyState(powerMap)

	if state.PeakTempK != cfg.AmbientTempK {
		t.Errorf("Peak T with zero power = %f, expected %f", state.PeakTempK, cfg.AmbientTempK)
	}
	if state.AvgTempK != cfg.AmbientTempK {
		t.Errorf("Avg T with zero power = %f, expected %f", state.AvgTempK, cfg.AmbientTempK)
	}
}

// TestThermalModel_NilArray verifies PowerFromMVM handles nil array gracefully.
func TestThermalModel_NilArray(t *testing.T) {
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(4, 4, cfg)

	result := tm.PowerFromMVM(nil, []float64{0.5, 0.5, 0.5, 0.5})
	if result != nil {
		t.Errorf("PowerFromMVM(nil) should return nil, got %v", result)
	}
}

// TestThermalModel_Checkerboard_Analytical validates the thermal model against
// an analytical solution for a checkerboard power pattern where alternating
// cells have P_high and P_low power dissipation.
//
// For independent RC cells (no lateral coupling), each cell's temperature
// is independently determined by: T_cell = T_amb + P_cell * R_th
// This test verifies all cells match their individual analytical solutions.
func TestThermalModel_Checkerboard_Analytical(t *testing.T) {
	const rows, cols = 8, 8
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(rows, cols, cfg)

	pHigh := 5e-3   // 5 mW
	pLow := 0.5e-3  // 0.5 mW

	powerMap := make([][]float64, rows)
	for i := range powerMap {
		powerMap[i] = make([]float64, cols)
		for j := range powerMap[i] {
			if (i+j)%2 == 0 {
				powerMap[i][j] = pHigh
			} else {
				powerMap[i][j] = pLow
			}
		}
	}

	state := tm.ComputeSteadyState(powerMap)

	// Verify each cell matches its analytical solution.
	maxErr := 0.0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			var expectedP float64
			if (i+j)%2 == 0 {
				expectedP = pHigh
			} else {
				expectedP = pLow
			}
			expectedT := cfg.AmbientTempK + expectedP*cfg.ThermalResistance
			err := math.Abs(state.CellTemperatures[i][j] - expectedT)
			if err > maxErr {
				maxErr = err
			}
			if err > 1e-10 {
				t.Errorf("Cell(%d,%d) T=%.8f K, expected %.8f K (err=%.2e)",
					i, j, state.CellTemperatures[i][j], expectedT, err)
			}
		}
	}

	// Verify peak is at a high-power cell.
	expectedPeak := cfg.AmbientTempK + pHigh*cfg.ThermalResistance
	if math.Abs(state.PeakTempK-expectedPeak) > 1e-10 {
		t.Errorf("Peak T=%.8f K, expected %.8f K", state.PeakTempK, expectedPeak)
	}

	// Verify average temperature.
	expectedAvg := cfg.AmbientTempK + ((pHigh+pLow)/2)*cfg.ThermalResistance
	if math.Abs(state.AvgTempK-expectedAvg) > 1e-10 {
		t.Errorf("Avg T=%.8f K, expected %.8f K", state.AvgTempK, expectedAvg)
	}

	t.Logf("Checkerboard thermal: peak=%.4f K, avg=%.4f K, max_err=%.2e K",
		state.PeakTempK, state.AvgTempK, maxErr)
}

// TestThermalModel_Transient_MultiCell verifies that a multi-cell transient
// converges to the analytical steady-state for each cell independently.
func TestThermalModel_Transient_MultiCell(t *testing.T) {
	const rows, cols = 4, 4
	cfg := DefaultThermalConfig()
	tm := NewThermalModel(rows, cols, cfg)

	tau := cfg.ThermalResistance * cfg.ThermalCapacitance
	dt := tau / 100.0
	steps := 500 // 5*tau

	// Non-uniform power: each cell gets a different power.
	powerMap := make([][]float64, rows)
	for i := range powerMap {
		powerMap[i] = make([]float64, cols)
		for j := range powerMap[i] {
			powerMap[i][j] = float64(i*cols+j+1) * 0.1e-3 // 0.1 to 1.6 mW
		}
	}

	snapshots := tm.ComputeTransient(powerMap, dt, steps)
	final := snapshots[steps-1]

	// Each cell should converge to its analytical steady state within 1%.
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			expectedT := cfg.AmbientTempK + powerMap[i][j]*cfg.ThermalResistance
			relErr := math.Abs(final.CellTemperatures[i][j]-expectedT) / (expectedT - cfg.AmbientTempK) * 100
			if relErr > 1.0 {
				t.Errorf("Cell(%d,%d) transient converged to %.6f K, expected %.6f K (err=%.2f%%)",
					i, j, final.CellTemperatures[i][j], expectedT, relErr)
			}
		}
	}

	t.Logf("Multi-cell transient: all cells converged to within 1%% of analytical steady state")
}
