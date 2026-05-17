//go:build legacy_fyne

package gui

import "testing"

func TestUnifiedStaticLayer_RedrawOnlyOnResize(t *testing.T) {
	ca := NewCircuitsApp()
	ca.deviceState = NewDeviceState(ca.arrayRows, ca.arrayCols, ca.tia, ca.adc)

	_ = ca.drawUnifiedArray(900, 600)
	first := ca.unifiedStaticRedrawCount
	if first != 1 {
		t.Fatalf("expected first static redraw count 1, got %d", first)
	}

	_ = ca.drawUnifiedArray(900, 600)
	if ca.unifiedStaticRedrawCount != first {
		t.Fatalf("expected static layer cache hit at same size, got redraw count %d", ca.unifiedStaticRedrawCount)
	}

	_ = ca.drawUnifiedArray(901, 600)
	if ca.unifiedStaticRedrawCount != first+1 {
		t.Fatalf("expected redraw on resize, got %d", ca.unifiedStaticRedrawCount)
	}
}
