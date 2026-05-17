//go:build legacy_fyne

package gui

import "testing"

func TestPassiveBehavior_0T1R_ColumnWriteAffectsEntireColumn(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil || ca.deviceState == nil {
		t.Fatal("expected circuits app with device state")
	}

	ca.resizeArray(8, 8)
	ca.deviceState.SetPassiveMode(true)
	ca.deviceState.SetOperationMode(OpModeWrite)

	targetRow, targetCol := 2, 3
	ca.deviceState.SetSelectedCell(targetRow, targetCol)

	// Set all cells to level 0 so any write-induced change is detectable.
	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = 0
		}
	}
	ca.mu.Unlock()

	writeV := ca.deviceState.GetWriteRange().Max
	if writeV <= 0 {
		writeV = 1.8
	}

	// In passive 0T mode with DAC-only column drive, the selected BL is driven to
	// writeV while all WLs are grounded. Every cell in the column sees the full
	// write voltage and must have its state updated.
	ca.applyColumnWrite(targetRow, targetCol, writeV)

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	// Same-column cells (r != targetRow) must have their levels changed from 0.
	anyChanged := false
	for r := 0; r < 8; r++ {
		if r == targetRow {
			continue
		}
		if ca.arrayWeights[r][targetCol] != 0 {
			anyChanged = true
		}
	}
	if !anyChanged {
		t.Fatal("expected at least one same-column cell to change level after column write")
	}

	// applyColumnWrite must NOT touch the selected cell (ISPP loop owns it).
	if ca.arrayWeights[targetRow][targetCol] != 0 {
		t.Fatalf("selected cell [%d,%d] must not be modified by applyColumnWrite; got level %d",
			targetRow, targetCol, ca.arrayWeights[targetRow][targetCol])
	}

	// Same-row cells (c != targetCol) see 0V in DAC-only drive — they must not change.
	for c := 0; c < 8; c++ {
		if c == targetCol {
			continue
		}
		if ca.arrayWeights[targetRow][c] != 0 {
			t.Fatalf("same-row cell [%d,%d] must not be modified (0V exposure); got level %d",
				targetRow, c, ca.arrayWeights[targetRow][c])
		}
	}
}
