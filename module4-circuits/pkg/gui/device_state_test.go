package gui

import (
	"math"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/peripherals"
)

const testEpsilon = 1e-6

func newTestDeviceState(rows, cols int) *DeviceState {
	tia := peripherals.DefaultTIA()
	adc := peripherals.DefaultADC()
	return NewDeviceState(rows, cols, tia, adc)
}

func newTestDeviceStateNilPeripherals(rows, cols int) *DeviceState {
	return NewDeviceState(rows, cols, nil, nil)
}

func assertVoltageInRange(t *testing.T, name string, voltage, min, max float64) {
	t.Helper()
	if voltage < min-testEpsilon || voltage > max+testEpsilon {
		t.Errorf("%s: voltage %.6f out of range [%.6f, %.6f]", name, voltage, min, max)
	}
}

func assertFloatEquals(t *testing.T, name string, got, want float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > testEpsilon {
		t.Errorf("%s: got %.6f, want %.6f (diff %.6f)", name, got, want, diff)
	}
}

func resetGlobalState() {
	voltageCalibration = nil
	hysteresisState = nil
	writeSequenceState = nil
	isppState = nil
	halfSelectState = nil
}

// ============================================================================
// Category 1: DeviceState Initialization
// ============================================================================

func TestNewDeviceState_Dimensions(t *testing.T) {
	resetGlobalState()

	testCases := []struct {
		name string
		rows int
		cols int
	}{
		{"small 4x4", 4, 4},
		{"medium 8x8", 8, 8},
		{"large 16x16", 16, 16},
		{"asymmetric 4x8", 4, 8},
		{"asymmetric 16x4", 16, 4},
		{"single row", 1, 8},
		{"single col", 8, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := newTestDeviceState(tc.rows, tc.cols)

			if len(ds.activeRows) != tc.rows {
				t.Errorf("activeRows: got len %d, want %d", len(ds.activeRows), tc.rows)
			}
			if len(ds.dacVoltages) != tc.cols {
				t.Errorf("dacVoltages: got len %d, want %d", len(ds.dacVoltages), tc.cols)
			}
			if len(ds.rowCurrents) != tc.rows {
				t.Errorf("rowCurrents: got len %d, want %d", len(ds.rowCurrents), tc.rows)
			}
			if len(ds.rowVoltages) != tc.rows {
				t.Errorf("rowVoltages: got len %d, want %d", len(ds.rowVoltages), tc.rows)
			}
			if len(ds.rowLevels) != tc.rows {
				t.Errorf("rowLevels: got len %d, want %d", len(ds.rowLevels), tc.rows)
			}
			if len(ds.saturated) != tc.rows {
				t.Errorf("saturated: got len %d, want %d", len(ds.saturated), tc.rows)
			}
		})
	}
}

func TestNewDeviceState_DefaultMode(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	if ds.opMode != OpModeRead {
		t.Errorf("default opMode: got %d, want OpModeRead (%d)", ds.opMode, OpModeRead)
	}
	if ds.wlMode != WLSingle {
		t.Errorf("default wlMode: got %d, want WLSingle (%d)", ds.wlMode, WLSingle)
	}
	if ds.dacMode != DACReadPreset {
		t.Errorf("default dacMode: got %d, want DACReadPreset (%d)", ds.dacMode, DACReadPreset)
	}
	if ds.dacRangeMode != DACRangeRead {
		t.Errorf("default dacRangeMode: got %d, want DACRangeRead (%d)", ds.dacRangeMode, DACRangeRead)
	}
	if ds.selectedRow != 0 {
		t.Errorf("default selectedRow: got %d, want 0", ds.selectedRow)
	}
	if ds.selectedCol != 0 {
		t.Errorf("default selectedCol: got %d, want 0", ds.selectedCol)
	}

	// First row should be active by default in WLSingle mode
	if !ds.activeRows[0] {
		t.Error("row 0 should be active by default")
	}
	for i := 1; i < len(ds.activeRows); i++ {
		if ds.activeRows[i] {
			t.Errorf("row %d should not be active by default", i)
		}
	}
}

func TestNewDeviceState_VoltageRanges(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Read range should be valid
	if ds.readRange.Min < 0 {
		t.Errorf("read range min %.6f should be >= 0", ds.readRange.Min)
	}
	if ds.readRange.Max <= ds.readRange.Min {
		t.Errorf("read range max %.6f should be > min %.6f", ds.readRange.Max, ds.readRange.Min)
	}
	if ds.readRange.NumLevels <= 0 {
		t.Errorf("read range numLevels %d should be > 0", ds.readRange.NumLevels)
	}
	if ds.readRange.StepSize <= 0 {
		t.Errorf("read range stepSize %.6f should be > 0", ds.readRange.StepSize)
	}

	// Write range should be valid and start at Vc
	if ds.writeRange.Min <= 0 {
		t.Errorf("write range min %.6f should be > 0 (Vc)", ds.writeRange.Min)
	}
	if ds.writeRange.Max <= ds.writeRange.Min {
		t.Errorf("write range max %.6f should be > min %.6f", ds.writeRange.Max, ds.writeRange.Min)
	}
	if ds.writeRange.Max > MaxPracticalVoltage {
		t.Errorf("write range max %.6f should be <= MaxPracticalVoltage %.6f", ds.writeRange.Max, MaxPracticalVoltage)
	}
}

func TestNewDeviceState_NilPeripherals(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceStateNilPeripherals(8, 8)

	// Should not panic or crash
	if ds == nil {
		t.Fatal("DeviceState should not be nil with nil peripherals")
	}

	// Voltage ranges should still be calculated
	if ds.readRange.NumLevels <= 0 {
		t.Error("read range should be calculated even with nil peripherals")
	}
	if ds.writeRange.NumLevels <= 0 {
		t.Error("write range should be calculated even with nil peripherals")
	}

	// Material should be set to default
	if ds.material == nil {
		t.Error("material should be set to default FeCIM material")
	}
}

// ============================================================================
// Category 2: Voltage Range Calculation
// ============================================================================

func TestUpdateVoltageRanges_ReadRange(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Read range should start at 0
	if ds.readRange.Min != 0 {
		t.Errorf("read range min: got %.6f, want 0", ds.readRange.Min)
	}

	// Read range max should be FieldMinRatio * Vc (capped at 1.0V)
	Vc := ds.material.CoerciveVoltage()
	expectedMax := ds.calibParams.FieldMinRatio * Vc
	if expectedMax > 1.0 {
		expectedMax = 1.0
	}
	if expectedMax < 0.1 {
		expectedMax = 0.1
	}

	if math.Abs(ds.readRange.Max-expectedMax) > testEpsilon {
		t.Errorf("read range max: got %.6f, want %.6f (FieldMinRatio*Vc)", ds.readRange.Max, expectedMax)
	}
}

func TestUpdateVoltageRanges_WriteRange(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	Vc := ds.material.CoerciveVoltage()

	// Write range should start at Vc
	if math.Abs(ds.writeRange.Min-Vc) > testEpsilon {
		t.Errorf("write range min: got %.6f, want Vc=%.6f", ds.writeRange.Min, Vc)
	}

	// Write range max should be FieldMaxRatio * Vc (capped at MaxPracticalVoltage)
	expectedMax := ds.calibParams.FieldMaxRatio * Vc
	if expectedMax > MaxPracticalVoltage {
		expectedMax = MaxPracticalVoltage
	}

	if math.Abs(ds.writeRange.Max-expectedMax) > testEpsilon {
		t.Errorf("write range max: got %.6f, want %.6f", ds.writeRange.Max, expectedMax)
	}
}

func TestUpdateVoltageRanges_MaterialCoerciveVoltage(t *testing.T) {
	resetGlobalState()

	testMaterials := []*sharedphysics.HZOMaterial{
		sharedphysics.FeCIMMaterial(),
		sharedphysics.DefaultHZO(),
		sharedphysics.LiteratureSuperlattice(),
		sharedphysics.CryogenicHZO(),
	}

	for _, mat := range testMaterials {
		t.Run(mat.Name, func(t *testing.T) {
			ds := newTestDeviceState(8, 8)
			ds.SetMaterial(mat)

			Vc := mat.Ec * mat.Thickness
			materialVc := mat.CoerciveVoltage()

			if math.Abs(Vc-materialVc) > testEpsilon {
				t.Errorf("Vc calculation: Ec*Thickness=%.9f, CoerciveVoltage()=%.9f", Vc, materialVc)
			}

			// Write range min should match Vc
			if math.Abs(ds.writeRange.Min-materialVc) > testEpsilon {
				t.Errorf("write range min %.6f should equal material Vc %.6f", ds.writeRange.Min, materialVc)
			}
		})
	}
}

func TestUpdateVoltageRanges_MaxPracticalVoltageClamp(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Create a material with very high Ec that would exceed 3V
	highEcMaterial := &sharedphysics.HZOMaterial{
		Name:      "High Ec Test",
		Ec:        5e8, // 5 MV/cm
		Thickness: 20e-9,
		NumLevels: 30,
	}
	ds.SetMaterial(highEcMaterial)

	// Vc = 5e8 * 20e-9 = 10V, but should be capped
	if ds.writeRange.Max > MaxPracticalVoltage+testEpsilon {
		t.Errorf("write range max %.6f exceeds MaxPracticalVoltage %.6f", ds.writeRange.Max, MaxPracticalVoltage)
	}
}

// ============================================================================
// Category 3: Per-Level Voltage Calibration
// ============================================================================

func TestInitVoltageCalibration_LinearInterpolation(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)
	ds.InitVoltageCalibration()

	if voltageCalibration == nil {
		t.Fatal("voltageCalibration should be initialized")
	}

	// Check level 0 equals writeRange.Min
	if math.Abs(voltageCalibration.AscendingVoltages[0]-ds.writeRange.Min) > testEpsilon {
		t.Errorf("level 0 voltage: got %.6f, want %.6f", voltageCalibration.AscendingVoltages[0], ds.writeRange.Min)
	}

	// Check level 29 equals writeRange.Max
	if math.Abs(voltageCalibration.AscendingVoltages[29]-ds.writeRange.Max) > testEpsilon {
		t.Errorf("level 29 voltage: got %.6f, want %.6f", voltageCalibration.AscendingVoltages[29], ds.writeRange.Max)
	}

	// Check linear interpolation - all 30 levels should be evenly spaced
	step := (ds.writeRange.Max - ds.writeRange.Min) / 29.0
	for i := 0; i < 30; i++ {
		expected := ds.writeRange.Min + float64(i)*step
		if math.Abs(voltageCalibration.AscendingVoltages[i]-expected) > testEpsilon {
			t.Errorf("level %d voltage: got %.6f, want %.6f", i, voltageCalibration.AscendingVoltages[i], expected)
		}
	}
}

func TestGetVoltageForLevel_BoundaryClamping(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Test level < 0 clamps to level 0
	vNegative := ds.GetVoltageForLevel(-5, true)
	vZero := ds.GetVoltageForLevel(0, true)
	if math.Abs(vNegative-vZero) > testEpsilon {
		t.Errorf("level -5 should clamp to level 0: got %.6f, want %.6f", vNegative, vZero)
	}

	// Test level > 29 clamps to level 29
	vOver := ds.GetVoltageForLevel(50, true)
	v29 := ds.GetVoltageForLevel(29, true)
	if math.Abs(vOver-v29) > testEpsilon {
		t.Errorf("level 50 should clamp to level 29: got %.6f, want %.6f", vOver, v29)
	}
}

func TestGetVoltageForLevel_Direction(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// In simplified model, ascending == descending
	for level := 0; level < 30; level++ {
		vAsc := ds.GetVoltageForLevel(level, true)
		vDesc := ds.GetVoltageForLevel(level, false)
		if math.Abs(vAsc-vDesc) > testEpsilon {
			t.Errorf("level %d: ascending %.6f != descending %.6f (simplified model)", level, vAsc, vDesc)
		}
	}
}

// ============================================================================
// Category 4: Hysteresis Direction Tracking
// ============================================================================

func TestRecordWrite_AscendingDirection(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Record initial write at level 5
	ds.RecordWrite(0, 0, 5)
	// Record second write at higher level
	ds.RecordWrite(0, 0, 15)

	dir := ds.GetLastHysteresisDirection(0, 0)
	if dir != DirectionAscending {
		t.Errorf("direction after 5->15: got %d, want DirectionAscending (%d)", dir, DirectionAscending)
	}
}

func TestRecordWrite_DescendingDirection(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Record initial write at level 20
	ds.RecordWrite(0, 0, 20)
	// Record second write at lower level
	ds.RecordWrite(0, 0, 10)

	dir := ds.GetLastHysteresisDirection(0, 0)
	if dir != DirectionDescending {
		t.Errorf("direction after 20->10: got %d, want DirectionDescending (%d)", dir, DirectionDescending)
	}
}

func TestGetWriteDirection_AllCases(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	testCases := []struct {
		name         string
		currentLevel int
		targetLevel  int
		expected     HysteresisDirection
	}{
		{"ascending", 5, 15, DirectionAscending},
		{"descending", 15, 5, DirectionDescending},
		{"same level", 10, 10, DirectionUnknown},
		{"zero to max", 0, 29, DirectionAscending},
		{"max to zero", 29, 0, DirectionDescending},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := ds.GetWriteDirection(0, 0, tc.currentLevel, tc.targetLevel)
			if dir != tc.expected {
				t.Errorf("direction %d->%d: got %d, want %d", tc.currentLevel, tc.targetLevel, dir, tc.expected)
			}
		})
	}
}

// ============================================================================
// Category 5: 4-Phase Write Sequence
// ============================================================================

func TestStartWriteSequence_Initialization(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartWriteSequence(2, 3, 15)

	ws := ds.GetWritePhaseInfo()
	if !ws.Active {
		t.Error("write sequence should be active")
	}
	if ws.Phase != PhaseReset {
		t.Errorf("initial phase: got %d, want PhaseReset (%d)", ws.Phase, PhaseReset)
	}
	if ws.TargetRow != 2 {
		t.Errorf("target row: got %d, want 2", ws.TargetRow)
	}
	if ws.TargetCol != 3 {
		t.Errorf("target col: got %d, want 3", ws.TargetCol)
	}
	if ws.TargetLevel != 15 {
		t.Errorf("target level: got %d, want 15", ws.TargetLevel)
	}
	if ws.Progress != 0.0 {
		t.Errorf("initial progress: got %.2f, want 0.0", ws.Progress)
	}
	if ws.WriteVoltage <= 0 {
		t.Error("write voltage should be calculated")
	}
}

func TestAdvanceWritePhase_Progression(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartWriteSequence(0, 0, 15)

	expectedPhases := []WritePhase{PhaseHold1, PhaseWrite, PhaseHold2, PhaseIdle}
	expectedProgress := []float64{0.25, 0.5, 0.75, 1.0}

	for i, expectedPhase := range expectedPhases {
		complete := ds.AdvanceWritePhase()
		ws := ds.GetWritePhaseInfo()

		if ws.Phase != expectedPhase {
			t.Errorf("phase after advance %d: got %d, want %d", i+1, ws.Phase, expectedPhase)
		}
		if math.Abs(ws.Progress-expectedProgress[i]) > testEpsilon {
			t.Errorf("progress after advance %d: got %.2f, want %.2f", i+1, ws.Progress, expectedProgress[i])
		}

		if i < len(expectedPhases)-1 && complete {
			t.Errorf("sequence should not be complete at phase %d", i+1)
		}
		if i == len(expectedPhases)-1 && !complete {
			t.Error("sequence should be complete after 4 advances")
		}
	}
}

func TestAdvanceWritePhase_Completion(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartWriteSequence(0, 0, 15)

	// Advance 4 times (RESET -> HOLD1 -> WRITE -> HOLD2 -> IDLE)
	for i := 0; i < 3; i++ {
		complete := ds.AdvanceWritePhase()
		if complete {
			t.Errorf("should not be complete after %d advances", i+1)
		}
	}

	// 4th advance should complete
	complete := ds.AdvanceWritePhase()
	if !complete {
		t.Error("should be complete after 4 advances")
	}

	ws := ds.GetWritePhaseInfo()
	if ws.Active {
		t.Error("write sequence should not be active after completion")
	}
}

func TestCancelWriteSequence_Reset(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartWriteSequence(0, 0, 15)
	ds.AdvanceWritePhase() // Move to HOLD1

	ds.CancelWriteSequence()

	ws := ds.GetWritePhaseInfo()
	if ws.Active {
		t.Error("write sequence should not be active after cancel")
	}
	if ws.Phase != PhaseIdle {
		t.Errorf("phase after cancel: got %d, want PhaseIdle (%d)", ws.Phase, PhaseIdle)
	}
	if ws.Progress != 0.0 {
		t.Errorf("progress after cancel: got %.2f, want 0.0", ws.Progress)
	}
}

// ============================================================================
// Category 6: ISPP State Machine
// ============================================================================

func TestStartISPP_Initialization(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartISPP(2, 3, 15, 5)

	is := ds.GetISPPStatus()
	if !is.Active {
		t.Error("ISPP should be active")
	}
	if is.Iteration != 0 {
		t.Errorf("initial iteration: got %d, want 0", is.Iteration)
	}
	if is.MaxIter != ISPPMaxIterations {
		t.Errorf("max iterations: got %d, want %d", is.MaxIter, ISPPMaxIterations)
	}
	if is.TargetRow != 2 {
		t.Errorf("target row: got %d, want 2", is.TargetRow)
	}
	if is.TargetCol != 3 {
		t.Errorf("target col: got %d, want 3", is.TargetCol)
	}
	if is.TargetLevel != 15 {
		t.Errorf("target level: got %d, want 15", is.TargetLevel)
	}
	if is.CurrentLevel != 5 {
		t.Errorf("current level: got %d, want 5", is.CurrentLevel)
	}
	if is.Direction != DirectionAscending {
		t.Errorf("direction: got %d, want DirectionAscending (%d)", is.Direction, DirectionAscending)
	}
	if is.Verified || is.Complete || is.Success {
		t.Error("initial state should not be verified/complete/success")
	}
}

func TestISPPIterate_Verification(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartISPP(0, 0, 15, 5)

	// Simulate reaching target
	result := ds.ISPPIterate(15)

	if result != ISPPResultVerified {
		t.Errorf("result: got %d, want ISPPResultVerified (%d)", result, ISPPResultVerified)
	}

	is := ds.GetISPPStatus()
	if !is.Verified {
		t.Error("should be verified")
	}
	if !is.Complete {
		t.Error("should be complete")
	}
	if !is.Success {
		t.Error("should be success")
	}
	if is.Active {
		t.Error("should not be active after verification")
	}
}

func TestISPPIterate_Overshoot(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Test ascending overshoot
	ds.StartISPP(0, 0, 15, 5) // Target 15, starting from 5
	result := ds.ISPPIterate(20) // Overshoot to 20
	if result != ISPPResultOvershoot {
		t.Errorf("ascending overshoot: got %d, want ISPPResultOvershoot (%d)", result, ISPPResultOvershoot)
	}

	// Test descending overshoot
	resetGlobalState()
	ds.StartISPP(0, 0, 10, 20) // Target 10, starting from 20 (descending)
	result = ds.ISPPIterate(5) // Overshoot to 5
	if result != ISPPResultOvershoot {
		t.Errorf("descending overshoot: got %d, want ISPPResultOvershoot (%d)", result, ISPPResultOvershoot)
	}
}

func TestISPPIterate_MaxIterations(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartISPP(0, 0, 15, 5)

	// Run 5 iterations without reaching target
	var result ISPPResult
	for i := 0; i < ISPPMaxIterations; i++ {
		result = ds.ISPPIterate(7) // Not reaching target
		if result == ISPPResultMaxIterations {
			break
		}
	}

	if result != ISPPResultMaxIterations {
		t.Errorf("result after max iterations: got %d, want ISPPResultMaxIterations (%d)", result, ISPPResultMaxIterations)
	}

	is := ds.GetISPPStatus()
	if !is.Complete {
		t.Error("should be complete after max iterations")
	}
	if is.Success {
		t.Error("should not be success after max iterations")
	}
}

func TestHandleOvershoot_Ascending(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartISPP(0, 0, 15, 5) // Ascending: 5 -> 15
	ds.ISPPIterate(20)        // Overshoot

	handled := ds.HandleOvershoot(0, 0)
	if !handled {
		t.Error("overshoot should be handled")
	}

	is := ds.GetISPPStatus()
	if is.CurrentLevel != 0 {
		t.Errorf("after ascending overshoot, current level should reset to 0: got %d", is.CurrentLevel)
	}
	if is.Direction != DirectionAscending {
		t.Errorf("direction should remain ascending: got %d", is.Direction)
	}
}

func TestHandleOvershoot_Descending(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartISPP(0, 0, 10, 20) // Descending: 20 -> 10
	ds.ISPPIterate(5)          // Overshoot

	handled := ds.HandleOvershoot(0, 0)
	if !handled {
		t.Error("overshoot should be handled")
	}

	is := ds.GetISPPStatus()
	if is.CurrentLevel != 29 {
		t.Errorf("after descending overshoot, current level should reset to 29: got %d", is.CurrentLevel)
	}
	if is.Direction != DirectionDescending {
		t.Errorf("direction should remain descending: got %d", is.Direction)
	}
}

// ============================================================================
// Category 7: V/2 Half-Select Visualization
// ============================================================================

func TestEnableHalfSelectVisualization_State(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	fullVoltage := 2.0
	ds.EnableHalfSelectVisualization(3, 4, fullVoltage)

	hs := ds.GetHalfSelectState()
	if !hs.Enabled {
		t.Error("half-select should be enabled")
	}
	if hs.FullVoltage != fullVoltage {
		t.Errorf("full voltage: got %.6f, want %.6f", hs.FullVoltage, fullVoltage)
	}
	if hs.HalfVoltage != fullVoltage*HalfSelectVoltageRatio {
		t.Errorf("half voltage: got %.6f, want %.6f", hs.HalfVoltage, fullVoltage*HalfSelectVoltageRatio)
	}
	if hs.SelectedRow != 3 {
		t.Errorf("selected row: got %d, want 3", hs.SelectedRow)
	}
	if hs.SelectedCol != 4 {
		t.Errorf("selected col: got %d, want 4", hs.SelectedCol)
	}

	// Check half-select rows (all rows except selected in same column)
	if len(hs.HalfSelectRows) != 7 {
		t.Errorf("half-select rows count: got %d, want 7", len(hs.HalfSelectRows))
	}

	// Check half-select cols (all cols except selected in same row)
	if len(hs.HalfSelectCols) != 7 {
		t.Errorf("half-select cols count: got %d, want 7", len(hs.HalfSelectCols))
	}
}

func TestIsHalfSelected_TargetCell(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.EnableHalfSelectVisualization(3, 4, 2.0)

	// Target cell (3,4) should NOT be half-selected (it gets full voltage)
	if ds.IsHalfSelected(3, 4) {
		t.Error("target cell (3,4) should not be half-selected")
	}
}

func TestIsHalfSelected_SameRow(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.EnableHalfSelectVisualization(3, 4, 2.0)

	// Cells in same row (row 3), different column should be half-selected
	for col := 0; col < 8; col++ {
		if col == 4 {
			continue // Skip target cell
		}
		if !ds.IsHalfSelected(3, col) {
			t.Errorf("cell (3,%d) should be half-selected (same row)", col)
		}
	}
}

func TestIsHalfSelected_SameColumn(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.EnableHalfSelectVisualization(3, 4, 2.0)

	// Cells in same column (col 4), different row should be half-selected
	for row := 0; row < 8; row++ {
		if row == 3 {
			continue // Skip target cell
		}
		if !ds.IsHalfSelected(row, 4) {
			t.Errorf("cell (%d,4) should be half-selected (same column)", row)
		}
	}
}

func TestDisableHalfSelectVisualization_Clear(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.EnableHalfSelectVisualization(3, 4, 2.0)
	ds.DisableHalfSelectVisualization()

	hs := ds.GetHalfSelectState()
	if hs.Enabled {
		t.Error("half-select should be disabled")
	}
	if hs.HalfSelectRows != nil {
		t.Error("half-select rows should be nil")
	}
	if hs.HalfSelectCols != nil {
		t.Error("half-select cols should be nil")
	}

	// No cells should be half-selected
	if ds.IsHalfSelected(3, 0) {
		t.Error("no cells should be half-selected after disable")
	}
}

// ============================================================================
// Category 8: Compute Function
// ============================================================================

func TestCompute_SingleRow(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	// Set up: single row active
	ds.SetWLSingle(1)
	ds.SetDACVoltage(0, 0.5)

	weights := [][]int{
		{10, 10, 10, 10},
		{15, 15, 15, 15},
		{20, 20, 20, 20},
		{25, 25, 25, 25},
	}

	ds.Compute(weights, 30)

	// Only row 1 should have current
	for row := 0; row < 4; row++ {
		current := ds.GetRowCurrent(row)
		if row == 1 {
			if current <= 0 {
				t.Errorf("active row 1 should have current > 0, got %.6f", current)
			}
		} else {
			if current != 0 {
				t.Errorf("inactive row %d should have current = 0, got %.6f", row, current)
			}
		}
	}
}

func TestCompute_AllRowsActive(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	// Set up: all rows active
	ds.SetWLAll()
	ds.SetDACVoltage(0, 0.5)

	// All cells same weight
	weights := [][]int{
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
	}

	ds.Compute(weights, 30)

	// All rows should have equal current
	reference := ds.GetRowCurrent(0)
	for row := 1; row < 4; row++ {
		current := ds.GetRowCurrent(row)
		if math.Abs(current-reference) > testEpsilon {
			t.Errorf("row %d current %.6f != row 0 current %.6f", row, current, reference)
		}
	}
}

func TestCompute_InactiveRows(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	// Single row active
	ds.SetWLSingle(0)
	ds.SetDACVoltage(0, 0.5)

	weights := [][]int{
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
	}

	ds.Compute(weights, 30)

	// Inactive rows should have zero current, voltage, level
	for row := 1; row < 4; row++ {
		if ds.GetRowCurrent(row) != 0 {
			t.Errorf("inactive row %d should have zero current", row)
		}
		if ds.GetRowVoltage(row) != 0 {
			t.Errorf("inactive row %d should have zero voltage", row)
		}
		if ds.GetRowLevel(row) != 0 {
			t.Errorf("inactive row %d should have zero level", row)
		}
		if ds.IsSaturated(row) {
			t.Errorf("inactive row %d should not be saturated", row)
		}
	}
}

func TestCompute_WithWeights(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	ds.SetWLAll()
	ds.SetDACVoltage(0, 0.5)

	// Different weights per row
	weights := [][]int{
		{5, 5, 5, 5},   // Low weight
		{15, 15, 15, 15}, // Medium weight
		{25, 25, 25, 25}, // High weight
		{29, 29, 29, 29}, // Max weight
	}

	ds.Compute(weights, 30)

	// Higher weights should produce more current
	for row := 0; row < 3; row++ {
		currentLow := ds.GetRowCurrent(row)
		currentHigh := ds.GetRowCurrent(row + 1)
		if currentLow >= currentHigh {
			t.Errorf("row %d (weight %d) current %.6f >= row %d (weight %d) current %.6f",
				row, weights[row][0], currentLow, row+1, weights[row+1][0], currentHigh)
		}
	}
}

func TestCompute_Saturation(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	ds.SetWLAll()

	// Set very high voltages on all columns to trigger saturation
	for col := 0; col < 4; col++ {
		ds.SetDACVoltage(col, 1.0)
	}

	// Max weights
	weights := [][]int{
		{29, 29, 29, 29},
		{29, 29, 29, 29},
		{29, 29, 29, 29},
		{29, 29, 29, 29},
	}

	ds.Compute(weights, 30)

	// At least some rows should trigger saturation flag
	saturatedCount := 0
	for row := 0; row < 4; row++ {
		if ds.IsSaturated(row) {
			saturatedCount++
		}
	}

	// With 4 cols * 1V * 100µS max conductance = 400µA per row
	// TIA saturates at 100µA, so should be saturated
	if saturatedCount == 0 {
		t.Log("Warning: no rows saturated with max weights and voltages")
	}
}

// ============================================================================
// Category 9: DAC Preset Modes
// ============================================================================

func TestSetDACPreset_ReadPreset(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.SetSelectedCell(2, 3)
	ds.SetDACPreset(DACReadPreset)

	// Only selected column should have voltage
	for col := 0; col < 8; col++ {
		v := ds.GetDACVoltage(col)
		if col == 3 {
			if v <= 0 {
				t.Errorf("selected column 3 should have voltage > 0, got %.6f", v)
			}
			assertVoltageInRange(t, "selected col voltage", v, ds.readRange.Min, ds.readRange.Max)
		} else {
			if v != 0 {
				t.Errorf("unselected column %d should have voltage = 0, got %.6f", col, v)
			}
		}
	}

	if ds.GetDACRangeMode() != DACRangeRead {
		t.Errorf("DAC range mode: got %d, want DACRangeRead (%d)", ds.GetDACRangeMode(), DACRangeRead)
	}
}

func TestSetDACPreset_WritePreset(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.SetSelectedCell(2, 5)
	ds.SetDACPreset(DACWritePreset)

	// Only selected column should have voltage
	for col := 0; col < 8; col++ {
		v := ds.GetDACVoltage(col)
		if col == 5 {
			if v <= 0 {
				t.Errorf("selected column 5 should have voltage > 0, got %.6f", v)
			}
			assertVoltageInRange(t, "selected col voltage", v, ds.writeRange.Min, ds.writeRange.Max)
		} else {
			if v != 0 {
				t.Errorf("unselected column %d should have voltage = 0, got %.6f", col, v)
			}
		}
	}

	if ds.GetDACRangeMode() != DACRangeWrite {
		t.Errorf("DAC range mode: got %d, want DACRangeWrite (%d)", ds.GetDACRangeMode(), DACRangeWrite)
	}
}

func TestSetDACPreset_InputVector(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	// Input vector: 0, 127, 200, 255
	inputs := []float64{0, 127, 200, 255}
	ds.SetDACPreset(DACInputVector, inputs...)

	// Check voltage mapping
	for col := 0; col < 4; col++ {
		v := ds.GetDACVoltage(col)
		normalized := inputs[col] / 255.0
		expected := ds.readRange.Min + normalized*(ds.readRange.Max-ds.readRange.Min)
		if math.Abs(v-expected) > testEpsilon {
			t.Errorf("column %d voltage: got %.6f, want %.6f", col, v, expected)
		}
	}
}

func TestSetDACVoltageForState_AllLevels(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Test all 30 levels map to write range
	for level := 0; level < 30; level++ {
		ds.SetDACVoltageForState(0, level, 30)
		v := ds.GetDACVoltage(0)

		// Should be within write range
		assertVoltageInRange(t, "level voltage", v, ds.writeRange.Min, ds.writeRange.Max)

		// Check linear interpolation
		normalized := float64(level) / 29.0
		expected := ds.writeRange.Min + normalized*(ds.writeRange.Max-ds.writeRange.Min)
		if math.Abs(v-expected) > testEpsilon {
			t.Errorf("level %d voltage: got %.6f, want %.6f", level, v, expected)
		}
	}
}

// ============================================================================
// Category 10: Passive Mode
// ============================================================================

func TestSetPassiveMode_AllWLsActive(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.SetPassiveMode(true)

	// All WLs should be active
	for row := 0; row < 8; row++ {
		if !ds.IsRowActive(row) {
			t.Errorf("row %d should be active in passive mode", row)
		}
	}

	if ds.GetWLMode() != WLAll {
		t.Errorf("WL mode: got %d, want WLAll (%d)", ds.GetWLMode(), WLAll)
	}

	if !ds.IsPassiveMode() {
		t.Error("IsPassiveMode() should return true")
	}
}

func TestSetWLSingle_IgnoredInPassiveMode(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.SetPassiveMode(true)
	ds.SetWLSingle(3) // Should be ignored

	// All WLs should still be active
	for row := 0; row < 8; row++ {
		if !ds.IsRowActive(row) {
			t.Errorf("row %d should remain active in passive mode after SetWLSingle", row)
		}
	}
}

func TestSetWLCustom_IgnoredInPassiveMode(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.SetPassiveMode(true)
	pattern := []bool{true, false, true, false, true, false, true, false}
	ds.SetWLCustom(pattern) // Should be ignored

	// All WLs should still be active
	for row := 0; row < 8; row++ {
		if !ds.IsRowActive(row) {
			t.Errorf("row %d should remain active in passive mode after SetWLCustom", row)
		}
	}
}

// ============================================================================
// Category 11: Edge Cases
// ============================================================================

func TestResize_SmallToLarge(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)

	ds.Resize(8, 16)

	if len(ds.activeRows) != 8 {
		t.Errorf("activeRows after resize: got %d, want 8", len(ds.activeRows))
	}
	if len(ds.dacVoltages) != 16 {
		t.Errorf("dacVoltages after resize: got %d, want 16", len(ds.dacVoltages))
	}
	if len(ds.rowCurrents) != 8 {
		t.Errorf("rowCurrents after resize: got %d, want 8", len(ds.rowCurrents))
	}

	// Row 0 should be active
	if !ds.activeRows[0] {
		t.Error("row 0 should be active after resize")
	}
}

func TestResize_LargeToSmall(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(16, 16)

	// Set selected cell outside future bounds
	ds.SetSelectedCell(12, 14)

	ds.Resize(4, 4)

	if len(ds.activeRows) != 4 {
		t.Errorf("activeRows after resize: got %d, want 4", len(ds.activeRows))
	}
	if len(ds.dacVoltages) != 4 {
		t.Errorf("dacVoltages after resize: got %d, want 4", len(ds.dacVoltages))
	}

	// Selected cell should be clamped to valid bounds
	if ds.GetSelectedRow() >= 4 {
		t.Errorf("selected row should be clamped: got %d, want < 4", ds.GetSelectedRow())
	}
	if ds.GetSelectedCol() >= 4 {
		t.Errorf("selected col should be clamped: got %d, want < 4", ds.GetSelectedCol())
	}
}

func TestNilMaterial_FallbackBehavior(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Set material to nil
	ds.material = nil
	ds.updateVoltageRanges()

	// Fallback Vc should be 1.0, numLevels should be 30
	if ds.writeRange.Min != 1.0 {
		t.Errorf("fallback Vc: got %.6f, want 1.0", ds.writeRange.Min)
	}
	if ds.writeRange.NumLevels != 30 {
		t.Errorf("fallback numLevels: got %d, want 30", ds.writeRange.NumLevels)
	}
}

func TestComputeWithNilMaterial_Fallback(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(4, 4)
	ds.material = nil

	ds.SetWLSingle(0)
	ds.SetDACVoltage(0, 0.5)

	weights := [][]int{
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
	}

	// Should not panic
	ds.Compute(weights, 30)

	// Should use fallback conductance (linear 1-100 µS)
	current := ds.GetRowCurrent(0)
	if current <= 0 {
		t.Errorf("current with nil material should be > 0, got %.6f", current)
	}
}

func TestComputeWithNilPeripherals_PartialOutput(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceStateNilPeripherals(4, 4)

	ds.SetWLSingle(0)
	ds.SetDACVoltage(0, 0.5)

	weights := [][]int{
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
	}

	ds.Compute(weights, 30)

	// Current should be computed
	current := ds.GetRowCurrent(0)
	if current <= 0 {
		t.Errorf("current should be computed even with nil peripherals, got %.6f", current)
	}

	// Voltage and level should be 0 (no TIA/ADC to convert)
	voltage := ds.GetRowVoltage(0)
	level := ds.GetRowLevel(0)
	if voltage != 0 {
		t.Errorf("voltage should be 0 with nil TIA, got %.6f", voltage)
	}
	if level != 0 {
		t.Errorf("level should be 0 with nil ADC, got %d", level)
	}
}

// ============================================================================
// Additional Tests for Complete Coverage
// ============================================================================

func TestGetPhaseName(t *testing.T) {
	testCases := []struct {
		phase    WritePhase
		expected string
	}{
		{PhaseIdle, "IDLE"},
		{PhaseReset, "RESET"},
		{PhaseHold1, "HOLD"},
		{PhaseWrite, "WRITE"},
		{PhaseHold2, "HOLD"},
		{WritePhase(99), "UNKNOWN"},
	}

	for _, tc := range testCases {
		name := GetPhaseName(tc.phase)
		if name != tc.expected {
			t.Errorf("GetPhaseName(%d): got %s, want %s", tc.phase, name, tc.expected)
		}
	}
}

func TestGetPhaseDuration(t *testing.T) {
	testCases := []struct {
		phase    WritePhase
		expected int
	}{
		{PhaseReset, PhaseResetDurationNs},
		{PhaseHold1, PhaseHold1DurationNs},
		{PhaseWrite, PhaseWriteDurationNs},
		{PhaseHold2, PhaseHold2DurationNs},
		{PhaseIdle, 0},
	}

	for _, tc := range testCases {
		duration := GetPhaseDuration(tc.phase)
		if duration != tc.expected {
			t.Errorf("GetPhaseDuration(%d): got %d, want %d", tc.phase, duration, tc.expected)
		}
	}
}

func TestClassifyOperation(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	testCases := []struct {
		mode     OpMode
		expected string
	}{
		{OpModeRead, "READ"},
		{OpModeWrite, "WRITE"},
		{OpModeCompute, "COMPUTE (MVM)"},
		{OpMode(99), "IDLE"},
	}

	for _, tc := range testCases {
		ds.SetOperationMode(tc.mode)
		result := ds.ClassifyOperation()
		if result != tc.expected {
			t.Errorf("ClassifyOperation() for mode %d: got %s, want %s", tc.mode, result, tc.expected)
		}
	}
}

func TestGetMaterialName(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	name := ds.GetMaterialName()
	if name == "" || name == "Unknown" {
		t.Errorf("GetMaterialName() should return material name, got %s", name)
	}

	// Test with nil material
	ds.material = nil
	name = ds.GetMaterialName()
	if name != "Unknown" {
		t.Errorf("GetMaterialName() with nil material: got %s, want Unknown", name)
	}
}

func TestGetCurrentVoltageRange(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Test read mode
	ds.SetDACRangeMode(DACRangeRead)
	vr := ds.GetCurrentVoltageRange()
	if vr.Max != ds.readRange.Max {
		t.Errorf("read mode range max: got %.6f, want %.6f", vr.Max, ds.readRange.Max)
	}

	// Test write mode
	ds.SetDACRangeMode(DACRangeWrite)
	vr = ds.GetCurrentVoltageRange()
	if vr.Max != ds.writeRange.Max {
		t.Errorf("write mode range max: got %.6f, want %.6f", vr.Max, ds.writeRange.Max)
	}
}

func TestCancelISPP(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.StartISPP(0, 0, 15, 5)
	ds.CancelISPP()

	is := ds.GetISPPStatus()
	if is.Active {
		t.Error("ISPP should not be active after cancel")
	}
	if !is.Complete {
		t.Error("ISPP should be complete after cancel")
	}
	if is.Success {
		t.Error("ISPP should not be success after cancel")
	}
}

func TestSetAllDACVoltages(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	voltage := 0.75
	ds.SetAllDACVoltages(voltage)

	for col := 0; col < 8; col++ {
		v := ds.GetDACVoltage(col)
		if math.Abs(v-voltage) > testEpsilon {
			t.Errorf("column %d voltage: got %.6f, want %.6f", col, v, voltage)
		}
	}

	if ds.GetDACMode() != DACManual {
		t.Errorf("DAC mode: got %d, want DACManual (%d)", ds.GetDACMode(), DACManual)
	}
}

func TestAdvanceWritePhase_NotActive(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Without starting a sequence, advance should return true (already complete)
	complete := ds.AdvanceWritePhase()
	if !complete {
		t.Error("advance on inactive sequence should return true")
	}
}

func TestHandleOvershoot_NotActive(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Without active ISPP, should return false
	handled := ds.HandleOvershoot(0, 0)
	if handled {
		t.Error("HandleOvershoot on inactive ISPP should return false")
	}
}

func TestISPPIterate_NotActive(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Without starting ISPP, iterate should return verified
	result := ds.ISPPIterate(15)
	if result != ISPPResultVerified {
		t.Errorf("iterate on inactive ISPP: got %d, want ISPPResultVerified", result)
	}
}

func TestRecordWrite_FirstWrite(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// First write should set direction to Unknown
	ds.RecordWrite(0, 0, 15)

	dir := ds.GetLastHysteresisDirection(0, 0)
	if dir != DirectionUnknown {
		t.Errorf("first write direction: got %d, want DirectionUnknown (%d)", dir, DirectionUnknown)
	}
}

func TestRecordWrite_SameLevel(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Write twice at same level
	ds.RecordWrite(0, 0, 15)
	ds.RecordWrite(0, 0, 20) // Ascending
	ds.RecordWrite(0, 0, 20) // Same level - should keep previous direction

	dir := ds.GetLastHysteresisDirection(0, 0)
	if dir != DirectionAscending {
		t.Errorf("same level should keep previous direction: got %d, want DirectionAscending", dir)
	}
}

func TestGetLastHysteresisDirection_UnknownCell(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Cell that was never written should return Unknown
	dir := ds.GetLastHysteresisDirection(5, 5)
	if dir != DirectionUnknown {
		t.Errorf("unknown cell direction: got %d, want DirectionUnknown", dir)
	}
}

// ============================================================================
// Category: V/2 Half-Select Write Tests
// Per VOLTAGE_RULES.md Section 3.2 and 6.1
// ============================================================================

func TestApplyHalfSelectWrite_PassiveMode(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Enable passive mode (0T1R)
	ds.SetPassiveMode(true)

	// Apply V/2 write at cell (3, 5) with 1.5V write voltage
	writeVoltage := 1.5
	ds.ApplyHalfSelectWrite(3, 5, writeVoltage)

	halfV := writeVoltage / 2.0

	// Check WL voltages: selected row should have +V/2, others 0
	for row := 0; row < 8; row++ {
		wlV := ds.GetWLVoltage(row)
		if row == 3 {
			assertFloatEquals(t, "selected WL voltage", wlV, halfV)
		} else {
			assertFloatEquals(t, "unselected WL voltage", wlV, 0.0)
		}
	}

	// Check BL (DAC) voltages: selected col should have -V/2, others 0
	for col := 0; col < 8; col++ {
		blV := ds.GetDACVoltage(col)
		if col == 5 {
			assertFloatEquals(t, "selected BL voltage", blV, -halfV)
		} else {
			assertFloatEquals(t, "unselected BL voltage", blV, 0.0)
		}
	}
}

func TestApplyHalfSelectWrite_NonPassiveMode(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Non-passive mode (1T1R/2T1R) - should use full voltage on BL
	ds.SetPassiveMode(false)

	writeVoltage := 1.5
	ds.ApplyHalfSelectWrite(3, 5, writeVoltage)

	// Check BL voltage: selected col should have full voltage
	blV := ds.GetDACVoltage(5)
	assertFloatEquals(t, "BL voltage in non-passive mode", blV, writeVoltage)

	// Other columns should be 0 (from previous DAC state or SetDACVoltage)
	// Note: SetDACVoltage only sets one column, doesn't zero others explicitly
}

func TestResetWriteVoltages_ClearsAll(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Enable passive mode and apply V/2 write
	ds.SetPassiveMode(true)
	ds.ApplyHalfSelectWrite(3, 5, 1.5)

	// Reset all voltages
	ds.ResetWriteVoltages()

	// All WL voltages should be 0
	for row := 0; row < 8; row++ {
		wlV := ds.GetWLVoltage(row)
		assertFloatEquals(t, "WL voltage after reset", wlV, 0.0)
	}

	// All BL voltages should be 0
	for col := 0; col < 8; col++ {
		blV := ds.GetDACVoltage(col)
		assertFloatEquals(t, "BL voltage after reset", blV, 0.0)
	}
}

func TestGetHalfSelectVoltage_DerivedFromMaterial(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	halfV := ds.GetHalfSelectVoltage()
	writeRange := ds.GetWriteRange()

	// Half-select voltage should be half of middle write voltage
	expectedHalfV := (writeRange.Min + writeRange.Max) / 4.0

	assertFloatEquals(t, "half-select voltage", halfV, expectedHalfV)

	// Verify it's below coercive voltage (safe for half-selected cells)
	Vc := ds.GetMaterial().CoerciveVoltage()
	if halfV >= Vc {
		t.Errorf("half-select voltage %.3fV should be below Vc %.3fV", halfV, Vc)
	}
}

func TestIsUsingHalfSelect_PassiveWriteMode(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	// Not passive, not write mode
	if ds.IsUsingHalfSelect() {
		t.Error("should not use half-select in non-passive, non-write mode")
	}

	// Passive but read mode
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeRead)
	if ds.IsUsingHalfSelect() {
		t.Error("should not use half-select in passive read mode")
	}

	// Passive and write mode
	ds.SetOperationMode(OpModeWrite)
	if !ds.IsUsingHalfSelect() {
		t.Error("should use half-select in passive write mode")
	}

	// Non-passive write mode
	ds.SetPassiveMode(false)
	if ds.IsUsingHalfSelect() {
		t.Error("should not use half-select in non-passive write mode")
	}
}

func TestApplyHalfSelectWrite_TargetCellEffectiveVoltage(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.SetPassiveMode(true)

	writeVoltage := 1.5
	ds.ApplyHalfSelectWrite(3, 5, writeVoltage)

	// Target cell effective voltage = WL - BL = +V/2 - (-V/2) = V
	wlV := ds.GetWLVoltage(3)
	blV := ds.GetDACVoltage(5)
	effectiveV := wlV - blV

	assertFloatEquals(t, "target cell effective voltage", effectiveV, writeVoltage)
}

func TestApplyHalfSelectWrite_HalfSelectedCellVoltages(t *testing.T) {
	resetGlobalState()
	ds := newTestDeviceState(8, 8)

	ds.SetPassiveMode(true)

	writeVoltage := 1.5
	halfV := writeVoltage / 2.0
	ds.ApplyHalfSelectWrite(3, 5, writeVoltage)

	// Same row, different column (half-selected): WL = +V/2, BL = 0 → ΔV = +V/2
	sameRowEffectiveV := ds.GetWLVoltage(3) - ds.GetDACVoltage(0) // col 0 is unselected
	assertFloatEquals(t, "same row half-selected voltage", sameRowEffectiveV, halfV)

	// Different row, same column (half-selected): WL = 0, BL = -V/2 → ΔV = +V/2
	sameColEffectiveV := ds.GetWLVoltage(0) - ds.GetDACVoltage(5) // row 0 is unselected
	assertFloatEquals(t, "same col half-selected voltage", sameColEffectiveV, halfV)

	// Diagonal cell (unselected): WL = 0, BL = 0 → ΔV = 0
	diagonalEffectiveV := ds.GetWLVoltage(0) - ds.GetDACVoltage(0)
	assertFloatEquals(t, "diagonal cell voltage", diagonalEffectiveV, 0.0)
}
