package gui

import (
	"math"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// TestUnifiedTabSliders tests slider widgets for valid configuration
func TestUnifiedTabSliders(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Enter WRITE mode to access write level slider
	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})

	// Test write level slider
	if ca.mfuxWriteLevelSlider == nil {
		t.Fatal("expected write level slider in write mode")
	}

	// Check slider bounds
	if ca.mfuxWriteLevelSlider.Min != 0 {
		t.Fatalf("write slider min: got %.0f, want 0", ca.mfuxWriteLevelSlider.Min)
	}
	expectedMax := float64(ca.quantLevels - 1)
	if ca.mfuxWriteLevelSlider.Max != expectedMax {
		t.Fatalf("write slider max: got %.0f, want %.0f", ca.mfuxWriteLevelSlider.Max, expectedMax)
	}
	if ca.mfuxWriteLevelSlider.Step != 1 {
		t.Fatalf("write slider step: got %.0f, want 1", ca.mfuxWriteLevelSlider.Step)
	}

	// Test programmatic value changes
	testLevel := float64(ca.quantLevels / 4)
	ca.mfuxWriteLevelSlider.SetValue(testLevel)
	if math.Abs(ca.mfuxWriteLevelSlider.Value-testLevel) > 0.1 {
		t.Fatalf("write slider value: got %.0f, want %.0f", ca.mfuxWriteLevelSlider.Value, testLevel)
	}

	// Test zoom slider
	if ca.zoomSlider == nil {
		t.Fatal("expected zoom slider")
	}
	if ca.zoomSlider.Min != 0.5 {
		t.Fatalf("zoom slider min: got %.2f, want 0.5", ca.zoomSlider.Min)
	}
	if ca.zoomSlider.Max != 3.0 {
		t.Fatalf("zoom slider max: got %.2f, want 3.0", ca.zoomSlider.Max)
	}
	if ca.zoomSlider.Step != 0.1 {
		t.Fatalf("zoom slider step: got %.2f, want 0.1", ca.zoomSlider.Step)
	}

	// Test zoom value change
	ca.zoomSlider.SetValue(2.0)
	waitFor(t, 200*time.Millisecond, "zoom update", func() bool {
		ca.mu.RLock()
		defer ca.mu.RUnlock()
		return math.Abs(ca.zoomLevel-2.0) < 0.01
	})
}

// TestUnifiedTabVoltageControls tests voltage-related controls
func TestUnifiedTabVoltageControls(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Test initial voltage state (READ mode default)
	selectedCol := ca.deviceState.GetSelectedCol()
	voltage := ca.deviceState.GetDACVoltage(selectedCol)
	if voltage <= 0 {
		t.Fatal("expected non-zero read voltage in READ mode")
	}

	// Enter WRITE mode
	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})

	// Verify WRITE mode has zero voltage until explicit program
	for col := 0; col < ca.arrayCols; col++ {
		v := ca.deviceState.GetDACVoltage(col)
		if v != 0 {
			t.Fatalf("expected 0V in WRITE mode before program, got %.3fV on col %d", v, col)
		}
	}

	// Test voltage range modes
	writeRange := ca.deviceState.GetWriteRange()
	if writeRange.Min >= writeRange.Max {
		t.Fatalf("invalid write range: min=%.2f, max=%.2f", writeRange.Min, writeRange.Max)
	}

	readRange := ca.deviceState.GetReadRange()
	if readRange.Min >= readRange.Max {
		t.Fatalf("invalid read range: min=%.2f, max=%.2f", readRange.Min, readRange.Max)
	}

	// Write range should exceed read range for programming
	if writeRange.Max <= readRange.Max {
		t.Fatalf("write range max (%.2f) should exceed read range max (%.2f)", writeRange.Max, readRange.Max)
	}
}

// TestComparisonTab tests the comparison tab rendering
func TestComparisonTab(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Check comparison tab components exist
	if ca.compArchCanvas == nil {
		t.Fatal("expected architecture comparison canvas")
	}
	if ca.compTimingCanvas == nil {
		t.Fatal("expected timing comparison canvas")
	}
	if ca.compEnergyCanvas == nil {
		t.Fatal("expected energy comparison canvas")
	}
	if ca.compStatusLabel == nil {
		t.Fatal("expected comparison status label")
	}

	// Test comparison actions
	ca.onRunComparison()
	if ca.compStatusLabel.Text == "" {
		t.Fatal("expected status message after running comparison")
	}

	// Test scale up functionality - initialize if needed
	if ca.compArraySize == 0 {
		ca.compArraySize = 8 // Default size
	}
	initialSize := ca.compArraySize
	ca.onScaleUpComparison()
	// Should cycle through sizes: 8 -> 16 -> 32 -> 64 -> 8
	if ca.compArraySize == initialSize {
		// Could be cycling back to same size, just verify it's a valid size
		validSizes := []int{8, 16, 32, 64}
		found := false
		for _, s := range validSizes {
			if ca.compArraySize == s {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected valid array size, got %d", ca.compArraySize)
		}
	}
}

// TestReferenceSpecsTab tests reference specs tab rendering
func TestReferenceSpecsTab(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Check specs components exist
	if ca.specArraySizeSelect == nil {
		t.Fatal("expected array size selector in specs")
	}
	if ca.specQuantLevelSelect == nil {
		t.Fatal("expected quantization level selector")
	}
	if ca.specDACBitsSelect == nil {
		t.Fatal("expected DAC bits selector")
	}
	if ca.specADCBitsSelect == nil {
		t.Fatal("expected ADC bits selector")
	}
	if ca.specTIAGainSelect == nil {
		t.Fatal("expected TIA gain selector")
	}
	if ca.specStatusLabel == nil {
		t.Fatal("expected specs status label")
	}

	// Test selector interactions
	ca.specArraySizeSelect.SetSelected("16")
	if ca.specArraySizeSelect.Selected != "16" {
		t.Fatal("failed to set array size selector")
	}

	ca.specDACBitsSelect.SetSelected("10")
	if ca.specDACBitsSelect.Selected != "10" {
		t.Fatal("failed to set DAC bits selector")
	}

	// Test comparison action
	ca.onCompareToGPU()
	if ca.specStatusLabel.Text == "" {
		t.Fatal("expected status message after GPU comparison")
	}
}

// TestReferenceTimingTab tests reference timing tab rendering
func TestReferenceTimingTab(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Check timing components exist
	if ca.timingOpSelect == nil {
		t.Fatal("expected timing operation selector")
	}
	if ca.timingWriteCanvas == nil {
		t.Fatal("expected write timing canvas")
	}
	if ca.timingReadCanvas == nil {
		t.Fatal("expected read timing canvas")
	}
	if ca.timingComputeCanvas == nil {
		t.Fatal("expected compute timing canvas")
	}
	if ca.timingStatusLabel == nil {
		t.Fatal("expected timing status label")
	}

	// Test operation selector
	operations := []string{"WRITE", "READ", "COMPUTE"}
	for _, op := range operations {
		ca.timingOpSelect.SetSelected(op)
		if ca.timingOpSelect.Selected != op {
			t.Fatalf("failed to set timing operation to %s", op)
		}
	}

	// Test refresh
	ca.refreshTimingDiagrams()
	// Should not panic
}

// TestUnifiedTabDeviceStateTransitions tests full lifecycle state transitions
func TestUnifiedTabDeviceStateTransitions(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Step 1: READ mode - sense a cell
	if ca.deviceState.GetOperationMode() != OpModeRead {
		t.Fatal("expected initial READ mode")
	}
	ca.onUnifiedRead()
	if ca.operationsStatusLabel.Text == "" {
		t.Fatal("expected status message after read")
	}

	// Step 2: WRITE mode - prepare to program
	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})
	if !ca.actionWriteCellBtn.Disabled() {
		// Program button should be enabled in WRITE mode
		targetLevel := ca.quantLevels / 4
		sharedwidgets.SafeDo(func() {
			if ca.mfuxWriteLevelSlider != nil {
				ca.mfuxWriteLevelSlider.SetValue(float64(targetLevel))
			}
		})
		// Note: We don't actually trigger ISPP here as it's a long-running goroutine
	}

	// Step 3: COMPUTE mode - setup input vector
	test.Tap(ca.modeComputeBtn)
	waitFor(t, 500*time.Millisecond, "compute mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeCompute
	})

	// Randomize inputs
	test.Tap(ca.computeRandomBtn)
	ca.mu.RLock()
	hasNonZero := false
	for _, v := range ca.inputVector {
		if v != 0 {
			hasNonZero = true
			break
		}
	}
	ca.mu.RUnlock()
	if !hasNonZero {
		t.Fatal("expected at least one non-zero input after randomize")
	}

	// Step 4: Run MVM
	test.Tap(ca.actionComputeBtn)
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		t.Fatal("mode should remain COMPUTE after MVM")
	}

	// Step 5: Clear inputs
	test.Tap(ca.computeClearBtn)
	ca.mu.RLock()
	for i, v := range ca.inputVector {
		if v != 0 {
			t.Fatalf("input[%d] not cleared: got %d", i, v)
		}
	}
	ca.mu.RUnlock()

	// Step 6: Back to READ mode
	test.Tap(ca.modeReadBtn)
	waitFor(t, 500*time.Millisecond, "read mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeRead
	})
}

// TestUnifiedTabArrayOperations tests array manipulation operations
func TestUnifiedTabArrayOperations(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Test array randomize
	test.Tap(ca.actionRandomArrayBtn)
	waitFor(t, 200*time.Millisecond, "randomize complete", func() bool {
		return ca.hasUndoHistory
	})

	if !ca.hasUndoHistory {
		t.Fatal("expected undo history after randomize")
	}
	if ca.undoHistoryBtn.Disabled() {
		t.Fatal("expected undo button enabled after randomize")
	}

	// Verify randomization changed values
	ca.mu.RLock()
	hasVariety := false
	firstVal := -1
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			if firstVal == -1 {
				firstVal = ca.arrayWeights[r][c]
			} else if ca.arrayWeights[r][c] != firstVal {
				hasVariety = true
				break
			}
		}
		if hasVariety {
			break
		}
	}
	ca.mu.RUnlock()
	if !hasVariety {
		t.Fatal("expected varied levels after randomize")
	}

	// Test undo
	test.Tap(ca.undoHistoryBtn)
	waitFor(t, 200*time.Millisecond, "undo disabled", func() bool {
		return ca.undoHistoryBtn.Disabled()
	})
	if ca.hasUndoHistory {
		t.Fatal("expected no undo history after undo")
	}

	// Test reset array
	midLevel := ca.quantLevels / 2
	test.Tap(ca.actionResetArrayBtn)
	waitFor(t, 200*time.Millisecond, "reset complete", func() bool {
		ca.mu.RLock()
		defer ca.mu.RUnlock()
		if len(ca.arrayWeights) == 0 || len(ca.arrayWeights[0]) == 0 {
			return false
		}
		return ca.arrayWeights[0][0] == midLevel
	})

	// Verify all cells reset to mid-level
	ca.mu.RLock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			if ca.arrayWeights[r][c] != midLevel {
				t.Fatalf("cell [%d,%d]: got %d, want %d after reset", r, c, ca.arrayWeights[r][c], midLevel)
			}
		}
	}
	ca.mu.RUnlock()

	// Test zoom fit
	ca.zoomSlider.SetValue(2.5)
	test.Tap(ca.actionFitBtn)
	waitFor(t, 200*time.Millisecond, "zoom fit", func() bool {
		return math.Abs(ca.zoomSlider.Value-1.0) < 0.01
	})
}

// TestUnifiedTabArchitectureToggle tests architecture switching
func TestUnifiedTabArchitectureToggle(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Test 1T1R
	test.Tap(ca.arch1T1RBtn)
	waitFor(t, 500*time.Millisecond, "1T1R architecture", func() bool {
		return ca.architecture == sharedwidgets.Architecture1T1R && !ca.deviceState.IsPassiveMode()
	})

	// Test 2T1R
	test.Tap(ca.arch2T1RBtn)
	waitFor(t, 500*time.Millisecond, "2T1R architecture", func() bool {
		return ca.architecture == sharedwidgets.Architecture2T1R && !ca.deviceState.IsPassiveMode()
	})

	// Test Passive (0T1R)
	test.Tap(ca.archPassiveBtn)
	waitFor(t, 500*time.Millisecond, "passive architecture", func() bool {
		return ca.architecture == sharedwidgets.Architecture0T1R && ca.deviceState.IsPassiveMode()
	})

	// Verify passive mode has all WLs active
	if !ca.deviceState.IsPassiveMode() {
		t.Fatal("expected passive mode after selecting 0T1R")
	}
}

// TestUnifiedTabMaterialSelector tests material selection
func TestUnifiedTabMaterialSelector(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	if ca.materialBtn == nil {
		t.Fatal("expected material button")
	}

	initialMaterial := ca.deviceState.GetMaterialName()
	if initialMaterial == "" {
		t.Fatal("expected initial material name")
	}

	// Material picker requires window interaction - just verify button exists and has text
	if ca.materialBtn.Text == "" {
		t.Fatal("expected material button to have text")
	}
}

// TestUnifiedTabArrayResize tests dynamic array resizing
func TestUnifiedTabArrayResize(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	initialRows := ca.arrayRows
	initialCols := ca.arrayCols

	// Resize to smaller array
	newSize := 2
	ca.resizeArray(newSize, newSize)

	if ca.arrayRows != newSize || ca.arrayCols != newSize {
		t.Fatalf("resize failed: got %dx%d, want %dx%d", ca.arrayRows, ca.arrayCols, newSize, newSize)
	}

	// Verify input/output vectors resized
	ca.mu.RLock()
	if len(ca.inputVector) != newSize {
		t.Fatalf("input vector size: got %d, want %d", len(ca.inputVector), newSize)
	}
	if len(ca.outputVector) != newSize {
		t.Fatalf("output vector size: got %d, want %d", len(ca.outputVector), newSize)
	}
	ca.mu.RUnlock()

	// Resize back to original
	ca.resizeArray(initialRows, initialCols)
	if ca.arrayRows != initialRows || ca.arrayCols != initialCols {
		t.Fatalf("resize back failed: got %dx%d, want %dx%d", ca.arrayRows, ca.arrayCols, initialRows, initialCols)
	}
}

// TestUnifiedTabCellSelection tests cell selection and info display
func TestUnifiedTabCellSelection(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Select a cell
	testRow, testCol := 2, 3
	ca.onUnifiedCellTapped(testRow, testCol)

	if ca.deviceState.GetSelectedRow() != testRow {
		t.Fatalf("selected row: got %d, want %d", ca.deviceState.GetSelectedRow(), testRow)
	}
	if ca.deviceState.GetSelectedCol() != testCol {
		t.Fatalf("selected col: got %d, want %d", ca.deviceState.GetSelectedCol(), testCol)
	}

	// Verify cell info label updated
	if ca.sharedCellInfoLabel.Text == "" {
		t.Fatal("expected cell info label to have text after selection")
	}
}

func TestUnifiedTabCellInfoSignedToggle(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}
	if ca.sharedCellDisplayToggle == nil {
		t.Fatal("expected Show V / Show I toggle")
	}

	ca.onUnifiedCellTapped(0, 0)
	ca.updateCellInfo()

	if !strings.Contains(ca.sharedCellInfoLabel.Text, "V_cell (V):") {
		t.Fatalf("expected explicit voltage label in default mode, got %q", ca.sharedCellInfoLabel.Text)
	}
	if strings.Contains(ca.sharedCellInfoLabel.Text, "I_cell (µA):") {
		t.Fatalf("did not expect current label in voltage mode, got %q", ca.sharedCellInfoLabel.Text)
	}
	if !(strings.Contains(ca.sharedCellInfoLabel.Text, "V_TIA (mV):") || strings.Contains(ca.sharedCellInfoLabel.Text, "V_TIA (V):")) || !strings.Contains(ca.sharedCellInfoLabel.Text, "ADC Code:") {
		t.Fatalf("expected TIA/ADC labels, got %q", ca.sharedCellInfoLabel.Text)
	}

	test.Tap(ca.sharedCellDisplayToggle)
	ca.updateCellInfo()
	if !strings.Contains(ca.sharedCellInfoLabel.Text, "I_cell (µA):") {
		t.Fatalf("expected current label after toggle, got %q", ca.sharedCellInfoLabel.Text)
	}
	if strings.Contains(ca.sharedCellInfoLabel.Text, "V_cell (V):") {
		t.Fatalf("did not expect voltage label in current mode, got %q", ca.sharedCellInfoLabel.Text)
	}
}

// TestUnifiedTabSensePanel tests sense panel visibility and updates
func TestUnifiedTabSensePanel(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Sense panel should be visible in READ mode
	test.Tap(ca.modeReadBtn)
	waitFor(t, 500*time.Millisecond, "read mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeRead
	})

	ca.updateSensePanel()

	// Verify sense panel components exist
	if ca.senseCurrentLabel == nil {
		t.Fatal("expected sense current label")
	}
	if ca.senseVoltageLabel == nil {
		t.Fatal("expected sense voltage label")
	}
	if ca.senseCodeLabel == nil {
		t.Fatal("expected sense code label")
	}
	if ca.senseSaturationLabel == nil {
		t.Fatal("expected sense saturation label")
	}
	if ca.sensePresetSelect == nil {
		t.Fatal("expected sense preset selector")
	}

	// Verify improved label formatting
	if !strings.Contains(ca.senseVoltageLabel.Text, "TIA out") {
		t.Fatalf("sense voltage label should show TIA context, got %q", ca.senseVoltageLabel.Text)
	}
	if !strings.Contains(ca.senseCodeLabel.Text, "Code") {
		t.Fatalf("sense code label should show Code prefix, got %q", ca.senseCodeLabel.Text)
	}

	// Test measurement preset application
	ca.applySensePreset("High Sensitivity")
	if math.Abs(ca.tiaGain-50.0) > 0.1 {
		t.Fatalf("preset TIA gain: got %.1f, want 50.0", ca.tiaGain)
	}

	// Manual edit should flip preset to Custom
	ca.applySenseRf("20.0")
	if math.Abs(ca.tiaGain-20.0) > 0.1 {
		t.Fatalf("TIA gain: got %.1f, want 20.0", ca.tiaGain)
	}
	if ca.sensePresetSelect.Selected != customSensePresetName {
		t.Fatalf("preset after manual edit: got %q, want %q", ca.sensePresetSelect.Selected, customSensePresetName)
	}
}

// TestUnifiedTabSensePanelLayoutAndPresets validates layout wiring and all presets.
func TestUnifiedTabSensePanelLayoutAndPresets(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}
	if ca.sensePresetSelect == nil {
		t.Fatal("expected preset selector")
	}

	// Default preset should be Balanced
	if ca.sensePresetSelect.Selected != "Balanced" {
		t.Fatalf("default preset: got %q, want Balanced", ca.sensePresetSelect.Selected)
	}

	// Apply each preset and verify parameters
	presets := []struct {
		name string
		rf   float64
		vmax float64
	}{
		{"Balanced", 10.0, 1.00},
		{"High Sensitivity", 50.0, 1.00},
		{"Wide Current Range", 5.0, 1.20},
		{"Low-Current Focus", 100.0, 0.90},
	}
	for _, p := range presets {
		ca.applySensePreset(p.name)
		if math.Abs(ca.tiaGain-p.rf) > 0.1 {
			t.Fatalf("preset %s: Rf got %.1f, want %.1f", p.name, ca.tiaGain, p.rf)
		}
		if math.Abs(ca.adc.VrefHigh-p.vmax) > 1e-6 {
			t.Fatalf("preset %s: Vmax got %.2f, want %.2f", p.name, ca.adc.VrefHigh, p.vmax)
		}
	}
}

// TestUnifiedTabCouplingMode tests fidelity tier selector wiring.
func TestUnifiedTabCouplingMode(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}
	if ca.couplingTierSelect == nil {
		t.Fatal("expected fidelity tier selector")
	}

	cases := []struct {
		label string
		mode  arraysim.CouplingMode
	}{
		{label: "Ideal", mode: arraysim.CouplingIdeal},
		{label: "Tier-A", mode: arraysim.CouplingTierA},
		{label: "Tier-B", mode: arraysim.CouplingTierB},
	}
	for _, tc := range cases {
		ca.couplingTierSelect.SetSelected(tc.label)
		waitFor(t, 200*time.Millisecond, "coupling mode set", func() bool {
			return ca.deviceState.GetCouplingMode() == tc.mode
		})
	}
}

// TestUnifiedTabISPPEngine tests ISPP engine selector
func TestUnifiedTabISPPEngine(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Enter WRITE mode to access ISPP engine selector
	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})

	if ca.isppEngineSelect == nil {
		t.Fatal("expected ISPP engine selector in write mode")
	}

	// Test engine selection wiring to device state
	cases := []struct {
		label    string
		expected ISPPEngine
	}{
		{label: "Preisach (Level-based)", expected: ISPPEngineLevel},
		{label: "Landau-Khalatnikov (Physics ODE)", expected: ISPPEngineLK},
	}
	for _, tc := range cases {
		ca.isppEngineSelect.SetSelected(tc.label)
		if ca.isppEngineSelect.Selected != tc.label {
			t.Fatalf("failed to set ISPP engine selector to %s", tc.label)
		}
		waitFor(t, 300*time.Millisecond, "device state ISPP engine sync", func() bool {
			return ca.deviceState.GetISPPEngine() == tc.expected
		})
	}
}

// TestUnifiedTabWriteTargetLabel tests write target label updates
func TestUnifiedTabWriteTargetLabel(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Enter WRITE mode
	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})

	if ca.mfuxWriteTargetLabel == nil {
		t.Fatal("expected write target label in write mode")
	}

	// Select a cell and verify label updates
	testRow, testCol := 1, 2
	ca.onUnifiedCellTapped(testRow, testCol)

	waitFor(t, 200*time.Millisecond, "target label updated", func() bool {
		return ca.mfuxWriteTargetLabel.Text != ""
	})
}

// TestUnifiedTabComputeInputsResize tests compute panel after array resize
func TestUnifiedTabComputeInputsResize(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Enter COMPUTE mode
	test.Tap(ca.modeComputeBtn)
	waitFor(t, 500*time.Millisecond, "compute mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeCompute
	})

	initialEntries := len(ca.mfuxInputVectorEntry)

	// Resize array
	newSize := 2
	ca.resizeArray(newSize, newSize)

	// Verify input entries rebuilt
	if len(ca.mfuxInputVectorEntry) != newSize {
		t.Fatalf("input entries: got %d, want %d after resize", len(ca.mfuxInputVectorEntry), newSize)
	}

	// Resize back
	ca.resizeArray(initialEntries, initialEntries)
	if len(ca.mfuxInputVectorEntry) != initialEntries {
		t.Fatalf("input entries: got %d, want %d after resize back", len(ca.mfuxInputVectorEntry), initialEntries)
	}
}

// TestUnifiedTabADCBitsSelector tests ADC resolution changes
func TestUnifiedTabADCBitsSelector(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// Test ADC bits changes
	bits := []int{5, 6, 7, 8}
	for _, b := range bits {
		ca.deviceState.SetADCBits(b)
		levels := ca.deviceState.GetADCLevels()
		expectedLevels := 1 << b
		if levels != expectedLevels {
			t.Fatalf("ADC levels for %d bits: got %d, want %d", b, levels, expectedLevels)
		}
	}
}

// TestComparisonTabCanvasRender tests comparison tab canvas rendering
func TestComparisonTabCanvasRender(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Verify canvases can render without panic
	if ca.compArchCanvas != nil {
		w, h := 400, 200
		img := ca.drawCompArch(w, h)
		if img == nil {
			t.Fatal("expected architecture comparison image")
		}
		bounds := img.Bounds()
		if bounds.Dx() != w || bounds.Dy() != h {
			t.Fatalf("arch canvas size: got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), w, h)
		}
	}

	if ca.compTimingCanvas != nil {
		w, h := 400, 150
		img := ca.drawCompTiming(w, h)
		if img == nil {
			t.Fatal("expected timing comparison image")
		}
		bounds := img.Bounds()
		if bounds.Dx() != w || bounds.Dy() != h {
			t.Fatalf("timing canvas size: got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), w, h)
		}
	}

	if ca.compEnergyCanvas != nil {
		w, h := 400, 200
		img := ca.drawCompEnergy(w, h)
		if img == nil {
			t.Fatal("expected energy comparison image")
		}
		bounds := img.Bounds()
		if bounds.Dx() != w || bounds.Dy() != h {
			t.Fatalf("energy canvas size: got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), w, h)
		}
	}
}

// TestReferenceTimingCanvasRender tests timing diagram rendering
func TestReferenceTimingCanvasRender(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Test write timing diagram
	if ca.timingWriteCanvas != nil {
		w, h := 600, 200
		img := ca.drawTimingWrite(w, h)
		if img == nil {
			t.Fatal("expected write timing image")
		}
		bounds := img.Bounds()
		if bounds.Dx() != w || bounds.Dy() != h {
			t.Fatalf("write timing size: got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), w, h)
		}
	}

	// Test read timing diagram
	if ca.timingReadCanvas != nil {
		w, h := 600, 180
		img := ca.drawTimingRead(w, h)
		if img == nil {
			t.Fatal("expected read timing image")
		}
		bounds := img.Bounds()
		if bounds.Dx() != w || bounds.Dy() != h {
			t.Fatalf("read timing size: got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), w, h)
		}
	}

	// Test compute timing diagram
	if ca.timingComputeCanvas != nil {
		w, h := 600, 200
		img := ca.drawTimingCompute(w, h)
		if img == nil {
			t.Fatal("expected compute timing image")
		}
		bounds := img.Bounds()
		if bounds.Dx() != w || bounds.Dy() != h {
			t.Fatalf("compute timing size: got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), w, h)
		}
	}
}

// TestReferenceSpecsSummaryUpdate tests specs summary recalculation
func TestReferenceSpecsSummaryUpdate(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	if ca.specArraySizeSelect == nil {
		t.Fatal("expected array size selector")
	}

	// Change array size and verify summary updates
	sizes := []string{"8", "16", "32", "64"}
	for _, size := range sizes {
		ca.specArraySizeSelect.SetSelected(size)
		ca.updateSpecSummary()
		// Should not panic and should update internal state
	}
}

// TestComparisonAnimateSteps tests comparison animation
func TestComparisonAnimateSteps(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	// Trigger animation (runs in goroutine, we just verify no panic)
	ca.onAnimateComparison()

	// Give goroutine time to start
	time.Sleep(100 * time.Millisecond)

	// Should update status label
	if ca.compStatusLabel.Text == "" {
		t.Fatal("expected status message during animation")
	}
}

// TestTimingAnimateOperations tests timing diagram animation
func TestTimingAnimateOperations(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	if ca.timingOpSelect == nil {
		t.Fatal("expected timing operation selector")
	}

	// Test animation for each operation
	operations := []string{"WRITE", "READ", "COMPUTE"}
	for _, op := range operations {
		ca.timingOpSelect.SetSelected(op)
		ca.onAnimateTiming()

		// Give goroutine time to start
		time.Sleep(50 * time.Millisecond)

		// Should update status
		if ca.timingStatusLabel.Text == "" {
			t.Fatal("expected status message during timing animation")
		}
	}
}
