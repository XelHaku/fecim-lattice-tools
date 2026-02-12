package gui

import (
	"math"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

func setupUnifiedTestApp(t *testing.T) (*EmbeddedCircuitsApp, fyne.App, fyne.Window) {
	t.Helper()
	app := test.NewApp()
	win := test.NewWindow(nil)
	embedded := NewEmbeddedCircuitsApp()
	sharedwidgets.WithUILock(func() {
		content := embedded.BuildContent(app, win)
		win.SetContent(content)
	})
	return embedded, app, win
}

func waitFor(t *testing.T, timeout time.Duration, label string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", label)
}

func TestUnifiedModeButtonsAndArchitecture(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	if got := ca.deviceState.GetOperationMode(); got != OpModeRead {
		t.Fatalf("default mode: got %v, want %v", got, OpModeRead)
	}

	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})

	test.Tap(ca.modeComputeBtn)
	waitFor(t, 500*time.Millisecond, "compute mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeCompute
	})

	test.Tap(ca.modeReadBtn)
	waitFor(t, 500*time.Millisecond, "read mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeRead
	})

	test.Tap(ca.arch1T1RBtn)
	waitFor(t, 500*time.Millisecond, "1T1R architecture", func() bool {
		return ca.architecture == sharedwidgets.Architecture1T1R && !ca.deviceState.IsPassiveMode()
	})

	test.Tap(ca.arch2T1RBtn)
	waitFor(t, 500*time.Millisecond, "2T1R architecture", func() bool {
		return ca.architecture == sharedwidgets.Architecture2T1R && !ca.deviceState.IsPassiveMode()
	})

	test.Tap(ca.archPassiveBtn)
	waitFor(t, 500*time.Millisecond, "passive architecture", func() bool {
		return ca.architecture == sharedwidgets.Architecture0T1R && ca.deviceState.IsPassiveMode()
	})
}

func TestUnifiedActionButtons(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	// READ mode should hide both Program Cell and MVM buttons.
	if ca.actionWriteCellBtn.Visible() {
		t.Fatal("program button should be hidden in READ mode")
	}
	if ca.actionComputeBtn.Visible() {
		t.Fatal("MVM button should be hidden in READ mode")
	}

	// Enter WRITE mode and ensure Program Cell shows/enables while MVM stays hidden.
	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})
	waitFor(t, 500*time.Millisecond, "program button enabled", func() bool {
		return ca.actionWriteCellBtn != nil && ca.actionWriteCellBtn.Visible() && !ca.actionWriteCellBtn.Disabled()
	})
	if ca.actionComputeBtn.Visible() {
		t.Fatal("MVM button should remain hidden in WRITE mode")
	}

	// Make sure target level differs from current to start ISPP.
	midLevel := ca.quantLevels / 2
	sharedwidgets.SafeDo(func() {
		if ca.mfuxWriteLevelSlider != nil {
			ca.mfuxWriteLevelSlider.SetValue(float64(midLevel + 1))
		}
	})

	test.Tap(ca.actionWriteCellBtn)
	waitFor(t, 500*time.Millisecond, "ISPP active", func() bool {
		return ca.deviceState.GetISPPStatus().Active
	})
	ca.deviceState.CancelISPP()

	// Random array should enable undo history.
	test.Tap(ca.actionRandomArrayBtn)
	if !ca.hasUndoHistory {
		t.Fatal("expected undo history after random array")
	}
	waitFor(t, 500*time.Millisecond, "undo enabled", func() bool {
		return ca.undoHistoryBtn != nil && !ca.undoHistoryBtn.Disabled()
	})

	// Undo should clear history and disable button.
	test.Tap(ca.undoHistoryBtn)
	waitFor(t, 500*time.Millisecond, "undo disabled", func() bool {
		return ca.undoHistoryBtn.Disabled()
	})

	// Reset array should restore mid-level.
	ca.mu.Lock()
	if len(ca.arrayWeights) > 0 && len(ca.arrayWeights[0]) > 0 {
		ca.arrayWeights[0][0] = 0
	}
	ca.mu.Unlock()
	test.Tap(ca.actionResetArrayBtn)
	ca.mu.RLock()
	resetLevel := ca.arrayWeights[0][0]
	ca.mu.RUnlock()
	if resetLevel != midLevel {
		t.Fatalf("reset level: got %d, want %d", resetLevel, midLevel)
	}

	// Zoom fit button should restore 1.0 zoom.
	ca.zoomSlider.SetValue(1.5)
	test.Tap(ca.actionFitBtn)
	if math.Abs(ca.zoomSlider.Value-1.0) > 0.001 {
		t.Fatalf("zoom fit: got %.3f, want 1.000", ca.zoomSlider.Value)
	}

	// Enter COMPUTE mode to exercise inputs and MVM.
	test.Tap(ca.modeComputeBtn)
	waitFor(t, 500*time.Millisecond, "compute mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeCompute
	})
	waitFor(t, 500*time.Millisecond, "compute button enabled", func() bool {
		return ca.actionComputeBtn != nil && ca.actionComputeBtn.Visible() && !ca.actionComputeBtn.Disabled()
	})
	if ca.actionWriteCellBtn.Visible() {
		t.Fatal("program button should be hidden in COMPUTE mode")
	}

	// Randomize inputs and verify values update.
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = 0
	}
	ca.mu.Unlock()
	test.Tap(ca.computeRandomBtn)
	ca.mu.RLock()
	changed := false
	for i, v := range ca.inputVector {
		if v < 0 || v > 255 {
			t.Fatalf("input[%d] out of range: %d", i, v)
		}
		if v != 0 {
			changed = true
		}
	}
	ca.mu.RUnlock()
	if !changed {
		t.Fatal("expected at least one non-zero input after randomize")
	}

	// Clear inputs should zero the vector.
	test.Tap(ca.computeClearBtn)
	ca.mu.RLock()
	for i, v := range ca.inputVector {
		if v != 0 {
			t.Fatalf("input[%d] not cleared: got %d", i, v)
		}
	}
	ca.mu.RUnlock()

	// MVM button should run without panic and keep compute mode.
	test.Tap(ca.actionComputeBtn)
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		t.Fatalf("mode after MVM: got %v, want %v", ca.deviceState.GetOperationMode(), OpModeCompute)
	}
	if ca.operationsStatusLabel != nil && !strings.HasPrefix(ca.operationsStatusLabel.Text, "MVM") {
		t.Fatalf("expected MVM status message, got %q", ca.operationsStatusLabel.Text)
	}
}
