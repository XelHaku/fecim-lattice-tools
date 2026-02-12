package gui

import (
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

func TestUnifiedHalfSelectVisualization_ShowsVoltageAndCellColors(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	waitFor(t, 500*time.Millisecond, "half-select indicator", func() bool {
		return ca.halfSelectIndicator != nil
	})

	sharedwidgets.WithUILock(func() {
		ca.deviceState.SetOperationMode(OpModeWrite)
		ca.deviceState.EnableHalfSelectVisualization(2, 3, 1.40)
		ca.updateHalfSelectVisualization()
	})

	waitFor(t, 500*time.Millisecond, "half-select active text", func() bool {
		return strings.Contains(ca.halfSelectIndicator.Text, "V/2 Bias Active") &&
			strings.Contains(ca.halfSelectIndicator.Text, "Full: 1.40V") &&
			strings.Contains(ca.halfSelectIndicator.Text, "Half: 0.70V")
	})

	_, ok := ca.getHalfSelectCellColor(2, 3)
	if ok {
		t.Fatal("expected target cell to have no V/2 overlay")
	}

	halfColor, ok := ca.getHalfSelectCellColor(2, 4)
	if !ok {
		t.Fatal("expected half-selected neighbor to be highlighted")
	}
	if got, want := color.RGBAModel.Convert(halfColor).(color.RGBA), colorHalfSelect; got != want {
		t.Fatalf("half-select color: got %#v, want %#v", got, want)
	}

	_, ok = ca.getHalfSelectCellColor(0, 0)
	if ok {
		t.Fatal("expected non-half-selected cell to have no overlay color")
	}

	ca.deviceState.SetOperationMode(OpModeRead)
	_, ok = ca.getHalfSelectCellColor(2, 4)
	if ok {
		t.Fatal("expected V/2 overlay to be hidden outside WRITE mode")
	}
}

func TestUnifiedDisturbAndDACDisplay_ReportChangesAndDACCode(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	ca.deviceState.SetPassiveMode(true)
	ca.deviceState.SetOperationMode(OpModeWrite)

	row, col := 1, 1
	ca.deviceState.SetSelectedCell(row, col)

	ca.mu.Lock()
	before := make([][]int, len(ca.arrayWeights))
	for r := range ca.arrayWeights {
		before[r] = append([]int(nil), ca.arrayWeights[r]...)
	}
	ca.mu.Unlock()

	// Intentionally large pulse to force visible disturb in neighboring cells.
	ca.deviceState.ApplyHalfSelectWrite(row, col, 8.0)
	changes := ca.applyHalfSelectDisturb(row, col)
	if changes <= 0 {
		t.Fatalf("expected disturb changes > 0, got %d", changes)
	}

	ca.mu.RLock()
	diffCount := 0
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			if before[r][c] != ca.arrayWeights[r][c] {
				diffCount++
			}
		}
	}
	ca.mu.RUnlock()

	if diffCount != changes {
		t.Fatalf("disturb report mismatch: reported %d changes, observed %d", changes, diffCount)
	}

	ca.deviceState.StartISPP(row, col, ca.quantLevels-1, 0)
	ca.updateISPPUI()
	waitFor(t, 500*time.Millisecond, "ISPP DAC text", func() bool {
		return ca.operationsStatusLabel != nil && strings.Contains(ca.operationsStatusLabel.Text, "DAC")
	})

	status := ca.operationsStatusLabel.Text
	if !strings.Contains(status, "V=") {
		t.Fatalf("expected applied voltage in status, got %q", status)
	}

	applied, code := ca.deviceState.DACWriteVoltage(ca.deviceState.GetISPPStatus().Voltage)
	if code < 0 {
		t.Fatalf("expected valid DAC code, got %d", code)
	}
	if math.Abs(applied) <= 0 {
		t.Fatalf("expected non-zero applied DAC voltage, got %.6f", applied)
	}
}
