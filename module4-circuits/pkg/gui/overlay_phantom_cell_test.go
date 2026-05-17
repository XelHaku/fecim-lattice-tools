//go:build legacy_fyne

package gui

import "testing"

func TestOverlayMode_DrawableCellCount8x8_NoPhantomCell(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	ca.resizeArray(8, 8)
	ca.readOverlaySelect.SetSelected("Vcell")

	_ = ca.drawUnifiedArray(1000, 780)

	drawRows, drawCols := ca.overlayDrawableDims(ca.arrayRows, ca.arrayCols, ca.arrayWeights)
	if got, want := drawRows*drawCols, 64; got != want {
		t.Fatalf("overlay drawable cell count mismatch: got %d, want %d", got, want)
	}
}
