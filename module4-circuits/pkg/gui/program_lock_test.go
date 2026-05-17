//go:build legacy_fyne

package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestProgramLockDisablesControls(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})

	ca.setProgrammingActive(true)
	defer ca.setProgrammingActive(false)

	if ca.actionWriteCellBtn == nil || !ca.actionWriteCellBtn.Disabled() {
		t.Fatal("program button should be disabled while programmingActive")
	}
	if ca.mfuxWriteLevelSlider == nil || !ca.mfuxWriteLevelSlider.Disabled() {
		t.Fatal("write level slider should be disabled while programmingActive")
	}
	if ca.archPassiveBtn == nil || !ca.archPassiveBtn.Disabled() {
		t.Fatal("architecture toggle should be disabled while programmingActive")
	}
	if ca.operationsStatusLabel == nil || !strings.Contains(ca.operationsStatusLabel.Text, "PROGRAMMING — controls locked") {
		t.Fatalf("expected lock status text, got %q", ca.operationsStatusLabel.Text)
	}
}

func TestISPPLabelUsesPreisachWording(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}
	if ca.isppEngineSelect == nil {
		t.Fatal("expected ISPP engine selector")
	}

	labels := strings.Join(ca.isppEngineSelect.Options, " | ")
	if !strings.Contains(labels, "Preisach") {
		t.Fatalf("expected Preisach wording in ISPP selector, got %q", labels)
	}
	if strings.Contains(labels, "Fast (Level)") {
		t.Fatalf("unexpected legacy Fast wording in ISPP selector: %q", labels)
	}
	if !strings.Contains(labels, "Landau-Khalatnikov (Physics ODE)") {
		t.Fatalf("expected Landau-Khalatnikov wording in ISPP selector, got %q", labels)
	}
	if strings.Contains(labels, "L-K (Physics)") {
		t.Fatalf("unexpected legacy L-K wording in ISPP selector: %q", labels)
	}
}
