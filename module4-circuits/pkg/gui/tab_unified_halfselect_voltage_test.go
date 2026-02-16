package gui

import (
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	crossbar "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
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
		ca.deviceState.EnableHalfSelectVisualization(2, 3, 1.40)
		ca.updateHalfSelectVisualization()
	})

	waitFor(t, 500*time.Millisecond, "half-select active text", func() bool {
		return strings.Contains(ca.halfSelectIndicator.Text, "Column Write Active") &&
			strings.Contains(ca.halfSelectIndicator.Text, "Target: 1.40V") &&
			strings.Contains(ca.halfSelectIndicator.Text, "Col Disturb: 1.40V")
	})

	targetColor, ok := ca.getHalfSelectCellColor(2, 3)
	if !ok {
		t.Fatal("expected target cell to be highlighted")
	}
	if got, want := color.RGBAModel.Convert(targetColor).(color.RGBA), colorFullVoltage; got != want {
		t.Fatalf("target color: got %#v, want %#v", got, want)
	}

	// Same Row (col 4): Should be safe (0V) -> No highlight
	_, ok = ca.getHalfSelectCellColor(2, 4)
	if ok {
		t.Fatal("expected same-row neighbor (safe) to NOT be highlighted")
	}

	// Same Column (row 1): Should be disturbed (Full V) -> Highlighted
	halfColor, ok := ca.getHalfSelectCellColor(1, 3)
	if !ok {
		t.Fatal("expected same-column neighbor (disturbed) to be highlighted")
	}
	if got, want := color.RGBAModel.Convert(halfColor).(color.RGBA), colorHalfSelect; got != want {
		t.Fatalf("half-select color: got %#v, want %#v", got, want)
	}

	_, ok = ca.getHalfSelectCellColor(0, 0)
	if ok {
		t.Fatal("expected non-half-selected cell to have no overlay color")
	}
}

func TestPassiveDisclosureText_RowAndColumnHalfSelect(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.halfSelectIndicator == nil {
		t.Fatal("expected half-select indicator")
	}
	if got, want := ca.halfSelectIndicator.Text, "0T1R: Column Write (Full Disturb)"; got != want {
		t.Fatalf("passive disclosure mismatch: got %q, want %q", got, want)
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

	// Pre-init engine with high stress rate so a single pulse exceeds the
	// threshold and causes visible level changes in neighbors.
	config := crossbar.DefaultWriteDisturbConfig()
	config.Enable = true
	config.Architecture1T1R = false
	config.StressAccumulationRate = 2.0 // 1 pulse > threshold (1.0)
	ca.writeDisturbEngine = crossbar.NewWriteDisturbEngine(ca.arrayRows, ca.arrayCols, config)

	// Set weights away from midpoint so the toward-center stress shift
	// actually changes levels (default weights are midLevel, which won't shift).
	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = 0
		}
	}
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
