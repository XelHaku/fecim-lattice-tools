//go:build legacy_fyne

package gui

import (
	"strings"
	"testing"
)

func TestOverlayRegression_NoPhantomCellsAndLabeledDualValues(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	sizes := []int{4, 8, 16}

	for _, n := range sizes {
		ca.resizeArray(n, n)
		ca.readOverlaySelect.SetSelected("Vcell")
		_ = ca.drawUnifiedArray(1000, 780)

		drawRows, drawCols := ca.overlayDrawableDims(ca.arrayRows, ca.arrayCols, ca.arrayWeights)
		if got, want := drawRows*drawCols, n*n; got != want {
			t.Fatalf("%dx%d overlay cell count mismatch: got %d, want %d", n, n, got, want)
		}

		for r := 0; r < n; r++ {
			for c := 0; c < n; c++ {
				level := ca.arrayWeights[r][c]
				v := ca.deviceState.GetEffectiveCellVoltage(r, c)
				info := ca.overlayCellInfo(level, ca.quantLevels, v, "Vcell")
				if info.TopLabel == "" || info.BottomLabel == "" {
					t.Fatalf("%dx%d [%d,%d] expected two labels, got top=%q bottom=%q", n, n, r, c, info.TopLabel, info.BottomLabel)
				}
				if info.TopLabel == info.BottomLabel {
					t.Fatalf("%dx%d [%d,%d] top and bottom labels must differ: %q", n, n, r, c, info.TopLabel)
				}
				if !strings.HasPrefix(info.TopLabel, "L:") {
					t.Fatalf("%dx%d [%d,%d] top label missing level prefix: %q", n, n, r, c, info.TopLabel)
				}
				if !strings.HasPrefix(info.BottomLabel, "V:") {
					t.Fatalf("%dx%d [%d,%d] bottom label missing V prefix: %q", n, n, r, c, info.BottomLabel)
				}
			}
		}
	}
}

func TestOverlayRegression_SelectedHighlightAlignment(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	ca.resizeArray(8, 8)
	ca.readOverlaySelect.SetSelected("Icell")

	selectedRow, selectedCol := 3, 5
	ca.deviceState.SetSelectedCell(selectedRow, selectedCol)
	_ = ca.drawUnifiedArray(1000, 780)

	cellRect := cellRectFor(selectedRow, selectedCol, ca.sharedArrayOffsetX, ca.sharedArrayOffsetY, ca.sharedArrayCellSize)
	highlightRect := selectedHighlightRectFor(selectedRow, selectedCol, ca.sharedArrayOffsetX, ca.sharedArrayOffsetY, ca.sharedArrayCellSize)

	if !highlightRect.In(cellRect.Inset(-4)) {
		t.Fatalf("highlight should align to selected cell (+/-4 px): cell=%v highlight=%v", cellRect, highlightRect)
	}

	for _, rc := range [][2]int{{selectedRow - 1, selectedCol}, {selectedRow + 1, selectedCol}, {selectedRow, selectedCol - 1}, {selectedRow, selectedCol + 1}} {
		r, c := rc[0], rc[1]
		if r < 0 || c < 0 || r >= ca.arrayRows || c >= ca.arrayCols {
			continue
		}
		neighbor := cellRectFor(r, c, ca.sharedArrayOffsetX, ca.sharedArrayOffsetY, ca.sharedArrayCellSize)
		if highlightRect.Min == neighbor.Min {
			t.Fatalf("highlight snapped to neighbor cell [%d,%d] instead of [%d,%d]", r, c, selectedRow, selectedCol)
		}
	}
}
