package gui

import "testing"

func TestValidArraySizesIncludes128(t *testing.T) {
	for _, size := range ValidArraySizes {
		if size == 128 {
			return
		}
	}
	t.Fatal("expected ValidArraySizes to include 128")
}

func TestResizeArraySupports128x128(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	ca.resizeArray(128, 128)
	if ca.arrayRows != 128 || ca.arrayCols != 128 {
		t.Fatalf("array dimensions mismatch: got %dx%d", ca.arrayRows, ca.arrayCols)
	}
	if got := ca.deviceState.GetDACVoltage(127); got != 0 {
		t.Fatalf("expected DAC voltage at col 127 initialized to 0, got %.6f", got)
	}
}
