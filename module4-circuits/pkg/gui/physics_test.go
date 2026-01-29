// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains physics validation tests for device_state.go
package gui

import (
	"math"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/peripherals"
)

// =============================================================================
// PHYSICS VALIDATION TESTS
// These tests verify that the implemented physics calculations match
// documented equations from VOLTAGE_RULES.md and scientific literature.
// =============================================================================

// TestMVMEquation verifies the fundamental MVM equation: I_row = Σ(G(row,col) × V(col))
// Per VOLTAGE_RULES.md Section 3.3 and PHYSICS.md
func TestMVMEquation(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(4, 4, tia, adc)
	ds.SetWLAll() // All rows active for MVM

	// Create test weight matrix - each row has identical weights for easy verification
	weights := make([][]int, 4)
	testLevels := []int{0, 10, 20, 29} // Different weight levels per row
	for r := range weights {
		weights[r] = make([]int, 4)
		for c := range weights[r] {
			weights[r][c] = testLevels[r]
		}
	}

	// Set uniform input voltage on all columns (0.5V each)
	testVoltage := 0.5
	for c := 0; c < 4; c++ {
		ds.SetDACVoltage(c, testVoltage)
	}

	// Compute
	quantLevels := 30
	ds.Compute(weights, quantLevels)

	// Verify MVM calculation for each row
	mat := ds.GetMaterial()
	for r := 0; r < 4; r++ {
		// Calculate expected current manually: I = Σ(G × V)
		// Since all columns have same weight and same voltage:
		// I = 4 × G(level) × V
		level := testLevels[r]
		conductanceS := mat.DiscreteLevel(level, quantLevels)
		conductanceUS := conductanceS * 1e6                  // Convert to µS
		expectedCurrent := 4 * conductanceUS * testVoltage   // 4 columns × G × V

		actualCurrent := ds.GetRowCurrent(r)

		// Allow 1% tolerance for floating point
		tolerance := 0.01 * expectedCurrent
		if tolerance < 0.001 {
			tolerance = 0.001 // Minimum tolerance
		}

		diff := math.Abs(actualCurrent - expectedCurrent)
		if diff > tolerance {
			t.Errorf("Row %d MVM mismatch: expected %.6f µA, got %.6f µA (diff %.6f)",
				r, expectedCurrent, actualCurrent, diff)
		}
	}
}

// TestConductanceFromWeight verifies G = material.DiscreteLevel(weight, quantLevels)
// Per material.go FeFET physics model
func TestConductanceFromWeight(t *testing.T) {
	mat := sharedphysics.FeCIMMaterial()
	quantLevels := 30

	// Test boundary conditions
	// DiscreteLevel formula: G = Gmin + (Gmax-Gmin) * (normalizedP + 1) / 2
	// where normalizedP = -1 + 2*level/(totalLevels-1)
	// At level 0: normalizedP = -1, G = Gmin
	// At level 29: normalizedP = +1, G = Gmax
	// At level 14: normalizedP = -1 + 2*14/29 = -0.0345, G ≈ Gmin + 0.483*(Gmax-Gmin)
	normalizedP14 := -1.0 + 2.0*14.0/29.0
	expectedMid := mat.Gmin + (mat.Gmax-mat.Gmin)*(normalizedP14+1)/2

	testCases := []struct {
		level    int
		expected float64 // Expected conductance in S
		desc     string
	}{
		{0, mat.Gmin, "Level 0 should give Gmin (HRS)"},
		{29, mat.Gmax, "Level 29 should give Gmax (LRS)"},
		{14, expectedMid, "Level 14 should match formula"}, // Not exact midpoint due to formula
	}

	for _, tc := range testCases {
		actual := mat.DiscreteLevel(tc.level, quantLevels)
		tolerance := 0.01 * tc.expected // 1% tolerance
		if tolerance < 1e-9 {
			tolerance = 1e-9
		}

		diff := math.Abs(actual - tc.expected)
		if diff > tolerance {
			t.Errorf("%s: expected %.2e S, got %.2e S (diff %.2e)",
				tc.desc, tc.expected, actual, diff)
		}
	}

	// Verify monotonicity: higher level = higher conductance
	prevG := 0.0
	for level := 0; level < quantLevels; level++ {
		G := mat.DiscreteLevel(level, quantLevels)
		if level > 0 && G < prevG {
			t.Errorf("Conductance not monotonic at level %d: G[%d]=%.2e < G[%d]=%.2e",
				level, level, G, level-1, prevG)
		}
		prevG = G
	}
}

// TestTIAConversion verifies V_out = I × Gain with saturation
// Per shared/peripherals/tia.go
func TestTIAConversion(t *testing.T) {
	tia := peripherals.DefaultTIA()

	testCases := []struct {
		currentA float64 // Input current in Amperes
		expected float64 // Expected output voltage
		desc     string
	}{
		{0, 0, "Zero current should give zero voltage"},
		{10e-6, 0.1, "10µA × 10kΩ = 0.1V"},
		{50e-6, 0.5, "50µA × 10kΩ = 0.5V"},
		{100e-6, 1.0, "100µA × 10kΩ = 1.0V (at saturation)"},
		{150e-6, 1.0, "150µA should saturate at MaxOutputVoltage (1.0V)"},
		{-10e-6, 0, "Negative current should clamp to 0V"},
	}

	for _, tc := range testCases {
		actual := tia.Convert(tc.currentA)
		tolerance := 0.001 // 1mV tolerance

		diff := math.Abs(actual - tc.expected)
		if diff > tolerance {
			t.Errorf("%s: expected %.4f V, got %.4f V", tc.desc, tc.expected, actual)
		}
	}
}

// TestADCQuantization verifies level = (V / Vref) × (2^bits - 1)
// Per shared/peripherals/adc.go
func TestADCQuantization(t *testing.T) {
	adc := peripherals.DefaultADC()

	// Verify 5-bit ADC gives 32 levels (0-31)
	if adc.Levels() != 32 {
		t.Errorf("5-bit ADC should have 32 levels, got %d", adc.Levels())
	}

	testCases := []struct {
		voltage  float64
		expected int
		desc     string
	}{
		{0.0, 0, "0V should give level 0"},
		{0.5, 16, "0.5V should give level 16 (midpoint of 0-31)"},
		{1.0, 31, "1.0V should give level 31 (max)"},
		{-0.1, 0, "Negative voltage should clamp to level 0"},
		{1.5, 31, "Voltage > Vref should clamp to max level"},
	}

	for _, tc := range testCases {
		actual := adc.Convert(tc.voltage)
		if actual != tc.expected {
			t.Errorf("%s: expected level %d, got %d", tc.desc, tc.expected, actual)
		}
	}

	// Verify resolution
	expectedResolution := 1.0 / 31.0 // (VrefHigh - VrefLow) / (levels - 1)
	actualResolution := adc.Resolution()
	if math.Abs(actualResolution-expectedResolution) > 1e-10 {
		t.Errorf("ADC resolution: expected %.6f V/LSB, got %.6f V/LSB",
			expectedResolution, actualResolution)
	}
}

// TestCoerciveVoltageDerivation verifies Vc = Ec × thickness
// Per VOLTAGE_RULES.md Section 2 "Coercive Voltage (Vc)"
func TestCoerciveVoltageDerivation(t *testing.T) {
	testMaterials := []*sharedphysics.HZOMaterial{
		sharedphysics.DefaultHZO(),
		sharedphysics.FeCIMMaterial(),
		sharedphysics.LiteratureSuperlattice(),
	}

	for _, mat := range testMaterials {
		// Vc = Ec × thickness
		expectedVc := mat.Ec * mat.Thickness
		actualVc := mat.CoerciveVoltage()

		if math.Abs(actualVc-expectedVc) > 1e-15 {
			t.Errorf("%s: Vc calculation mismatch. Ec=%.2e V/m, t=%.2e m, expected Vc=%.4f V, got %.4f V",
				mat.Name, mat.Ec, mat.Thickness, expectedVc, actualVc)
		}

		t.Logf("%s: Ec=%.2f MV/cm, t=%.1f nm, Vc=%.3f V",
			mat.Name, mat.Ec/1e8, mat.Thickness*1e9, actualVc)
	}
}

// TestVoltageRulesCompliance verifies voltage constraints from VOLTAGE_RULES.md
// - Read voltage < Vc (non-destructive)
// - Write voltage > Vc (for switching)
// - Half-select voltage V/2 < Vc (minimize disturb)
func TestVoltageRulesCompliance(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)

	// Get derived voltage ranges
	readRange := ds.GetReadRange()
	writeRange := ds.GetWriteRange()
	mat := ds.GetMaterial()
	Vc := mat.CoerciveVoltage()

	t.Logf("Material: %s, Vc=%.3f V", mat.Name, Vc)
	t.Logf("Read range: %.3f - %.3f V", readRange.Min, readRange.Max)
	t.Logf("Write range: %.3f - %.3f V", writeRange.Min, writeRange.Max)

	// Rule 1: Read voltage < Vc (non-destructive read)
	// readRange.Max should be derived from calibParams.FieldMinRatio * Vc
	if readRange.Max >= Vc {
		t.Errorf("VIOLATION: Read voltage max (%.3f V) >= Vc (%.3f V) - reads would disturb data",
			readRange.Max, Vc)
	}

	// Rule 2: Write voltage >= Vc (ensure switching)
	if writeRange.Min < Vc {
		t.Errorf("VIOLATION: Write voltage min (%.3f V) < Vc (%.3f V) - writes may not switch",
			writeRange.Min, Vc)
	}

	// Rule 3: Half-select V/2 < Vc (minimize disturb in passive mode)
	ds.SetPassiveMode(true) // Enable 0T1R passive mode
	halfSelectV := ds.GetHalfSelectVoltage()

	if halfSelectV >= Vc {
		t.Errorf("VIOLATION: Half-select voltage (%.3f V) >= Vc (%.3f V) - would disturb half-selected cells",
			halfSelectV, Vc)
	}

	// Log the safety margins
	readMargin := (Vc - readRange.Max) / Vc * 100
	halfSelectMargin := (Vc - halfSelectV) / Vc * 100
	t.Logf("Read safety margin: %.1f%% below Vc", readMargin)
	t.Logf("Half-select safety margin: %.1f%% below Vc", halfSelectMargin)
}

// TestHalfSelectVoltageScheme verifies V/2 half-select biasing for passive 0T1R
// Per VOLTAGE_RULES.md Section 3.2 and 6.1
func TestHalfSelectVoltageScheme(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(4, 4, tia, adc)
	ds.SetPassiveMode(true) // Enable passive 0T1R mode

	// Apply V/2 half-select for writing to cell (1, 2)
	targetRow, targetCol := 1, 2
	writeVoltage := 1.5 // Full write voltage

	ds.ApplyHalfSelectWrite(targetRow, targetCol, writeVoltage)

	halfV := writeVoltage / 2.0

	// Verify WL voltages
	for r := 0; r < 4; r++ {
		wlV := ds.GetWLVoltage(r)
		if r == targetRow {
			// Selected row should have +V/2
			if math.Abs(wlV-halfV) > 0.001 {
				t.Errorf("Selected WL[%d] should be +%.3f V, got %.3f V", r, halfV, wlV)
			}
		} else {
			// Unselected rows should be 0V
			if math.Abs(wlV) > 0.001 {
				t.Errorf("Unselected WL[%d] should be 0V, got %.3f V", r, wlV)
			}
		}
	}

	// Verify BL (DAC) voltages
	for c := 0; c < 4; c++ {
		blV := ds.GetDACVoltage(c)
		if c == targetCol {
			// Selected column should have -V/2
			if math.Abs(blV-(-halfV)) > 0.001 {
				t.Errorf("Selected BL[%d] should be -%.3f V, got %.3f V", c, halfV, blV)
			}
		} else {
			// Unselected columns should be 0V
			if math.Abs(blV) > 0.001 {
				t.Errorf("Unselected BL[%d] should be 0V, got %.3f V", c, blV)
			}
		}
	}

	// Verify effective voltages on cells
	// Target cell: ΔV = WL - BL = +V/2 - (-V/2) = V_write
	targetDeltaV := ds.GetWLVoltage(targetRow) - ds.GetDACVoltage(targetCol)
	if math.Abs(targetDeltaV-writeVoltage) > 0.001 {
		t.Errorf("Target cell ΔV should be %.3f V, got %.3f V", writeVoltage, targetDeltaV)
	}

	// Half-selected cells (same row): ΔV = WL - BL = +V/2 - 0 = +V/2
	for c := 0; c < 4; c++ {
		if c != targetCol {
			halfSelectDeltaV := ds.GetWLVoltage(targetRow) - ds.GetDACVoltage(c)
			if math.Abs(halfSelectDeltaV-halfV) > 0.001 {
				t.Errorf("Half-selected cell (%d,%d) ΔV should be %.3f V, got %.3f V",
					targetRow, c, halfV, halfSelectDeltaV)
			}
		}
	}

	// Half-selected cells (same column): ΔV = WL - BL = 0 - (-V/2) = +V/2
	for r := 0; r < 4; r++ {
		if r != targetRow {
			halfSelectDeltaV := ds.GetWLVoltage(r) - ds.GetDACVoltage(targetCol)
			if math.Abs(halfSelectDeltaV-halfV) > 0.001 {
				t.Errorf("Half-selected cell (%d,%d) ΔV should be %.3f V, got %.3f V",
					r, targetCol, halfV, halfSelectDeltaV)
			}
		}
	}

	// Unselected (diagonal) cells: ΔV = 0 - 0 = 0
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if r != targetRow && c != targetCol {
				unselectedDeltaV := ds.GetWLVoltage(r) - ds.GetDACVoltage(c)
				if math.Abs(unselectedDeltaV) > 0.001 {
					t.Errorf("Unselected cell (%d,%d) ΔV should be 0V, got %.3f V",
						r, c, unselectedDeltaV)
				}
			}
		}
	}
}

// TestTransistorIsolationForNonPassive verifies 1T1R/2T1R don't use V/2 scheme
// Per VOLTAGE_RULES.md Section 4 and 5 - transistor isolation replaces V/2
func TestTransistorIsolationForNonPassive(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(4, 4, tia, adc)

	// Non-passive mode (1T1R/2T1R) - should NOT use V/2
	ds.SetPassiveMode(false)

	writeVoltage := 1.5
	targetCol := 2

	// In non-passive mode, ApplyHalfSelectWrite should just set full voltage
	ds.ApplyHalfSelectWrite(1, targetCol, writeVoltage)

	// Selected column should have full write voltage (not V/2)
	blV := ds.GetDACVoltage(targetCol)
	if math.Abs(blV-writeVoltage) > 0.001 {
		t.Errorf("Non-passive mode: Selected BL should have full voltage %.3f V, got %.3f V",
			writeVoltage, blV)
	}

	// WL voltages should not be modified (transistor gate handles isolation)
	// In 1T1R, WL controls transistor gate, not V/2 biasing
	if ds.IsUsingHalfSelect() {
		t.Error("Non-passive mode should NOT use half-select scheme")
	}
}

// TestMVMOutputRangeMatchesADC verifies MVM output stays within ADC input range
// Per VOLTAGE_RULES.md "Safety Checks" - MVM range = ADC range
func TestMVMOutputRangeMatchesADC(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)
	ds.SetWLAll()

	// Create worst-case weight matrix (all max)
	weights := make([][]int, 8)
	for r := range weights {
		weights[r] = make([]int, 8)
		for c := range weights[r] {
			weights[r][c] = 29 // Max weight
		}
	}

	// Apply max input voltage on all columns
	readRange := ds.GetReadRange()
	for c := 0; c < 8; c++ {
		ds.SetDACVoltage(c, readRange.Max)
	}

	// Compute
	ds.Compute(weights, 30)

	// Verify all TIA outputs are within ADC range
	for r := 0; r < 8; r++ {
		voltage := ds.GetRowVoltage(r)

		if voltage < adc.VrefLow {
			t.Errorf("Row %d voltage %.4f V below ADC VrefLow %.4f V", r, voltage, adc.VrefLow)
		}

		// TIA should auto-scale or saturate to keep within ADC range
		if voltage > tia.MaxOutputVoltage {
			t.Errorf("Row %d voltage %.4f V exceeds TIA MaxOutputVoltage %.4f V",
				r, voltage, tia.MaxOutputVoltage)
		}

		level := ds.GetRowLevel(r)
		if level < 0 || level >= adc.Levels() {
			t.Errorf("Row %d ADC level %d out of valid range [0, %d]", r, level, adc.Levels()-1)
		}
	}
}

// TestConductanceRange verifies conductance bounds from material
// Per material.go: Gmin = 1µS (HRS), Gmax = 100µS (LRS), ratio ~100x
func TestConductanceRange(t *testing.T) {
	mat := sharedphysics.FeCIMMaterial()

	// Verify Gmin and Gmax are set
	if mat.Gmin <= 0 {
		t.Errorf("Gmin should be positive, got %.2e S", mat.Gmin)
	}
	if mat.Gmax <= 0 {
		t.Errorf("Gmax should be positive, got %.2e S", mat.Gmax)
	}

	// Verify on/off ratio
	onOffRatio := mat.Gmax / mat.Gmin
	expectedRatio := 100.0 // Per FeCIM spec
	tolerance := 10.0      // Allow some variation

	if math.Abs(onOffRatio-expectedRatio) > tolerance {
		t.Errorf("On/off ratio should be ~%.0f, got %.1f", expectedRatio, onOffRatio)
	}

	t.Logf("FeCIM conductance: Gmin=%.2f µS, Gmax=%.2f µS, ratio=%.0f:1",
		mat.Gmin*1e6, mat.Gmax*1e6, onOffRatio)
}

// TestVoltageCalibrationPerLevel verifies voltage-level mapping
// Per VOLTAGE_RULES.md Section 3.2.1 "Per-Level Voltage Calibration"
func TestVoltageCalibrationPerLevel(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)

	// Initialize voltage calibration
	ds.InitVoltageCalibration()

	writeRange := ds.GetWriteRange()

	// Test that each level has unique calibrated voltage
	for level := 0; level < 30; level++ {
		ascV := ds.GetVoltageForLevel(level, true)
		descV := ds.GetVoltageForLevel(level, false)

		// Voltage should be within write range
		if ascV < writeRange.Min || ascV > writeRange.Max {
			t.Errorf("Level %d ascending voltage %.3f V outside write range [%.3f, %.3f]",
				level, ascV, writeRange.Min, writeRange.Max)
		}

		if descV < writeRange.Min || descV > writeRange.Max {
			t.Errorf("Level %d descending voltage %.3f V outside write range [%.3f, %.3f]",
				level, descV, writeRange.Min, writeRange.Max)
		}
	}

	// Verify voltage is monotonic with level (for ascending branch)
	prevV := 0.0
	for level := 0; level < 30; level++ {
		v := ds.GetVoltageForLevel(level, true)
		if level > 0 && v < prevV {
			t.Errorf("Voltage not monotonic at level %d: V[%d]=%.4f < V[%d]=%.4f",
				level, level, v, level-1, prevV)
		}
		prevV = v
	}
}

// TestISPPStateMachine verifies ISPP (Incremental Step Pulse Programming) logic
// Per VOLTAGE_RULES.md Section 3.2.1 "Program-Verify Loop (ISPP)"
func TestISPPStateMachine(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)

	// Start ISPP for a cell
	targetRow, targetCol := 3, 5
	targetLevel := 20
	currentLevel := 10

	ds.StartISPP(targetRow, targetCol, targetLevel, currentLevel)

	status := ds.GetISPPStatus()
	if !status.Active {
		t.Error("ISPP should be active after StartISPP")
	}
	if status.TargetLevel != targetLevel {
		t.Errorf("ISPP target level should be %d, got %d", targetLevel, status.TargetLevel)
	}
	if status.CurrentLevel != currentLevel {
		t.Errorf("ISPP current level should be %d, got %d", currentLevel, status.CurrentLevel)
	}
	if status.Direction != DirectionAscending {
		t.Errorf("ISPP direction should be Ascending (going from %d to %d)", currentLevel, targetLevel)
	}

	// Simulate iteration: target reached
	result := ds.ISPPIterate(targetLevel)
	if result != ISPPResultVerified {
		t.Errorf("ISPP should return Verified when target is reached, got %v", result)
	}

	status = ds.GetISPPStatus()
	if status.Active {
		t.Error("ISPP should be inactive after verification")
	}
	if !status.Success {
		t.Error("ISPP should report success")
	}
}

// TestISPPOvershoot verifies overshoot detection and handling
// Per VOLTAGE_RULES.md Section 3.2.1 "Overshoot handling"
func TestISPPOvershoot(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)

	// Test ascending overshoot: target=15, but we write to 18
	ds.StartISPP(0, 0, 15, 10)
	result := ds.ISPPIterate(18) // Overshoot target

	if result != ISPPResultOvershoot {
		t.Errorf("Should detect overshoot when writing to 18 with target 15, got %v", result)
	}

	// Test descending overshoot: target=10, but we write to 7
	ds.StartISPP(0, 0, 10, 20) // Descending direction
	status := ds.GetISPPStatus()
	if status.Direction != DirectionDescending {
		t.Errorf("Direction should be Descending when going from 20 to 10")
	}

	result = ds.ISPPIterate(7) // Overshoot target (went past 10)
	if result != ISPPResultOvershoot {
		t.Errorf("Should detect descending overshoot when writing to 7 with target 10, got %v", result)
	}
}

// TestWritePhaseSequence verifies 4-phase write sequence
// Per VOLTAGE_RULES.md Section 3.2.1 "4-Phase Write Sequence"
func TestWritePhaseSequence(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)

	// Start write sequence
	ds.StartWriteSequence(2, 3, 25)

	info := ds.GetWritePhaseInfo()
	if !info.Active {
		t.Error("Write sequence should be active")
	}
	if info.Phase != PhaseReset {
		t.Errorf("Initial phase should be PhaseReset, got %v", info.Phase)
	}
	if info.TargetLevel != 25 {
		t.Errorf("Target level should be 25, got %d", info.TargetLevel)
	}

	// Advance through phases
	expectedPhases := []WritePhase{PhaseHold1, PhaseWrite, PhaseHold2, PhaseIdle}
	for _, expected := range expectedPhases {
		complete := ds.AdvanceWritePhase()
		info = ds.GetWritePhaseInfo()

		if info.Phase != expected {
			t.Errorf("Expected phase %v, got %v", expected, info.Phase)
		}

		if expected == PhaseIdle {
			if !complete {
				t.Error("Sequence should be complete at PhaseIdle")
			}
			if info.Active {
				t.Error("Sequence should be inactive after completion")
			}
		} else {
			if complete {
				t.Errorf("Sequence should not be complete at phase %v", expected)
			}
		}
	}
}

// TestHysteresisDirectionTracking verifies hysteresis state tracking
// Per VOLTAGE_RULES.md Section 3.2.1 "Hysteresis Path-Dependence"
func TestHysteresisDirectionTracking(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)

	row, col := 2, 4

	// Initial state should be unknown
	dir := ds.GetLastHysteresisDirection(row, col)
	if dir != DirectionUnknown {
		t.Errorf("Initial direction should be Unknown, got %v", dir)
	}

	// Record ascending write (from 5 to 15)
	ds.RecordWrite(row, col, 5)
	ds.RecordWrite(row, col, 15)
	dir = ds.GetLastHysteresisDirection(row, col)
	if dir != DirectionAscending {
		t.Errorf("Direction after going 5→15 should be Ascending, got %v", dir)
	}

	// Record descending write (from 15 to 8)
	ds.RecordWrite(row, col, 8)
	dir = ds.GetLastHysteresisDirection(row, col)
	if dir != DirectionDescending {
		t.Errorf("Direction after going 15→8 should be Descending, got %v", dir)
	}

	// Determine write direction for new target
	detectedDir := ds.GetWriteDirection(row, col, 8, 20)
	if detectedDir != DirectionAscending {
		t.Errorf("Direction for 8→20 should be Ascending, got %v", detectedDir)
	}

	detectedDir = ds.GetWriteDirection(row, col, 20, 5)
	if detectedDir != DirectionDescending {
		t.Errorf("Direction for 20→5 should be Descending, got %v", detectedDir)
	}
}

// TestADCBitResolution verifies configurable ADC bits (5-8)
// Per device_state.go SetADCBits
func TestADCBitResolution(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(8, 8, tia, adc)

	testCases := []struct {
		bits     int
		expected int // Expected number of levels
	}{
		{5, 32},
		{6, 64},
		{7, 128},
		{8, 256},
	}

	for _, tc := range testCases {
		ds.SetADCBits(tc.bits)

		actualBits := ds.GetADCBits()
		if actualBits != tc.bits {
			t.Errorf("SetADCBits(%d): GetADCBits returned %d", tc.bits, actualBits)
		}

		actualLevels := ds.GetADCLevels()
		if actualLevels != tc.expected {
			t.Errorf("SetADCBits(%d): expected %d levels, got %d",
				tc.bits, tc.expected, actualLevels)
		}
	}

	// Verify invalid values are rejected
	ds.SetADCBits(4) // Below minimum
	if ds.GetADCBits() == 4 {
		t.Error("SetADCBits should reject bits < 5")
	}

	ds.SetADCBits(9) // Above maximum
	if ds.GetADCBits() == 9 {
		t.Error("SetADCBits should reject bits > 8")
	}
}

// TestPassiveModeWLEnforcement verifies all WLs are always on in passive mode
// Per device_state.go SetPassiveMode
func TestPassiveModeWLEnforcement(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(4, 4, tia, adc)

	// Enable passive mode
	ds.SetPassiveMode(true)

	// All WLs should be active
	for r := 0; r < 4; r++ {
		if !ds.IsRowActive(r) {
			t.Errorf("Passive mode: Row %d should be active", r)
		}
	}

	// Try to set single WL - should be ignored in passive mode
	ds.SetWLSingle(2)

	// All WLs should still be active
	for r := 0; r < 4; r++ {
		if !ds.IsRowActive(r) {
			t.Errorf("Passive mode after SetWLSingle: Row %d should still be active", r)
		}
	}

	if ds.GetWLMode() != WLAll {
		t.Errorf("Passive mode should force WLAll mode, got %v", ds.GetWLMode())
	}
}

// TestEndToEndMVMComputation tests complete MVM pipeline
// Input vector → DAC → Crossbar (I=G×V) → TIA → ADC → Output levels
func TestEndToEndMVMComputation(t *testing.T) {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	ds := NewDeviceState(4, 4, tia, adc)
	ds.SetWLAll()
	ds.SetOperationMode(OpModeCompute)

	// Create a known weight pattern
	// Row 0: all weights = 0 (minimum conductance)
	// Row 1: all weights = 15 (mid conductance)
	// Row 2: all weights = 29 (max conductance)
	// Row 3: weights = {0, 10, 20, 29} (mixed)
	weights := [][]int{
		{0, 0, 0, 0},
		{15, 15, 15, 15},
		{29, 29, 29, 29},
		{0, 10, 20, 29},
	}

	// Input vector: uniform 0.5V
	inputVoltage := 0.5
	for c := 0; c < 4; c++ {
		ds.SetDACVoltage(c, inputVoltage)
	}

	// Compute
	ds.Compute(weights, 30)

	// Verify output ordering (row 0 < row 1 < row 2)
	current0 := ds.GetRowCurrent(0)
	current1 := ds.GetRowCurrent(1)
	current2 := ds.GetRowCurrent(2)

	if current0 >= current1 {
		t.Errorf("Current ordering wrong: row0 (%.4f) should be < row1 (%.4f)",
			current0, current1)
	}
	if current1 >= current2 {
		t.Errorf("Current ordering wrong: row1 (%.4f) should be < row2 (%.4f)",
			current1, current2)
	}

	// Verify all outputs have valid ADC levels
	for r := 0; r < 4; r++ {
		level := ds.GetRowLevel(r)
		if level < 0 || level > 31 {
			t.Errorf("Row %d: ADC level %d out of range [0, 31]", r, level)
		}

		t.Logf("Row %d: I=%.4f µA, V=%.4f V, Level=%d",
			r, ds.GetRowCurrent(r), ds.GetRowVoltage(r), level)
	}
}
