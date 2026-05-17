//go:build legacy_fyne

package gui

import "testing"

// applyHalfSelectDisturb is a no-op for passive 0T mode with DAC-only column drive:
//   - Same-column cells see the full write voltage → handled by applyColumnWrite
//   - Same-row cells see 0V (unselected BLs grounded) → no disturb
// These tests verify the no-op contract and delegate column-write coverage to
// TestPassiveBehavior_0T1R_ColumnWriteAffectsEntireColumn.

func TestApplyHalfSelectDisturb_NoOpForPassiveDACOnlyMode(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca.deviceState == nil {
		t.Fatal("expected device state")
	}
	ca.deviceState.SetPassiveMode(true)

	targetRow, targetCol := 1, 1
	writeV := ca.deviceState.GetWriteRange().Max
	if writeV <= 0 {
		writeV = 1.8
	}
	ca.deviceState.ApplyHalfSelectWrite(targetRow, targetCol, writeV)

	changes := ca.applyHalfSelectDisturb(targetRow, targetCol)
	if changes != 0 {
		t.Fatalf("applyHalfSelectDisturb must return 0 in passive DAC-only mode, got %d", changes)
	}

	// writeDisturbEngine must not be initialized (DAC-only drive doesn't use it).
	if ca.writeDisturbEngine != nil {
		t.Error("writeDisturbEngine must not be initialized in passive DAC-only mode")
	}
}

func TestApplyHalfSelectDisturb_8x8_NoStressInDACOnlyMode(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca.deviceState == nil {
		t.Fatal("expected device state")
	}
	ca.resizeArray(8, 8)
	ca.deviceState.SetPassiveMode(true)

	targetRow, targetCol := 2, 3
	writeV := ca.deviceState.GetWriteRange().Max
	if writeV <= 0 {
		writeV = 1.8
	}
	ca.deviceState.ApplyHalfSelectWrite(targetRow, targetCol, writeV)
	ca.applyHalfSelectDisturb(targetRow, targetCol)

	// No stress engine should be active — DAC-only drive has no V/2 half-select.
	if ca.writeDisturbEngine != nil {
		t.Error("writeDisturbEngine must not be initialized in passive DAC-only mode")
	}
}
