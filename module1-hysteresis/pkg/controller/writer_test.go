package controller

import "testing"

func TestWriteController_CalculateNextField_StuckRelaxesBoundsWithoutResettingToZero(t *testing.T) {
	wc := NewWriteController(30, 1.0, 2.5, nil)
	wc.TargetLevel = 11

	// Simulate a situation where we've already pulsed a few times, have a tight bracket,
	// and then get stuck (no improvement). This used to hard-reset VMin to 0.
	wc.PulseCount = 1
	wc.CurrentField = -0.97
	wc.VMin = 0.97
	wc.VMax = 1.16
	wc.VMinSet = true
	wc.VMaxSet = true
	wc.StuckCount = 3
	wc.NoImproveCount = 0
	wc.stepFloor = 0

	wc.calculateNextField(16)

	if wc.VMin <= 0 {
		t.Fatalf("VMin reset: got %v, want > 0", wc.VMin)
	}
	if !wc.VMinSet {
		t.Fatalf("VMinSet: got false, want true")
	}
	if wc.VMax != wc.MaxField {
		t.Fatalf("VMax: got %v, want %v", wc.VMax, wc.MaxField)
	}
	if wc.VMaxSet {
		t.Fatalf("VMaxSet: got true, want false")
	}
}
