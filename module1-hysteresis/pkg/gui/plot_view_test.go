package gui

import "testing"

func TestTrimTail(t *testing.T) {
	hE := []float64{1, 2, 3, 4, 5}
	hP := []float64{10, 20, 30, 40, 50}
	te, tp := trimTail(hE, hP, 3)
	if len(te) != 3 || te[0] != 3 || tp[0] != 30 {
		t.Fatalf("unexpected trim: %v %v", te, tp)
	}
}

func TestLastCompleteCycle_ZeroCrossingFallback(t *testing.T) {
	// E crosses zero rising twice.
	hE := []float64{-1, -0.5, 0.2, 1, 0.2, -0.2, -1, -0.2, 0.1, 1, 0.1}
	hP := make([]float64, len(hE))
	for i := range hP {
		hP[i] = float64(i)
	}
	e2, p2 := lastCompleteCycle(hE, hP)
	if len(e2) == len(hE) {
		t.Fatalf("expected slicing, got full input")
	}
	if len(e2) < 5 {
		t.Fatalf("cycle too small: %d", len(e2))
	}
	// Ensure indices correspond to the last complete rising zero crossing window.
	if p2[0] <= 2 { // first crossing is at index 2
		t.Fatalf("expected later slice, got p2[0]=%v", p2[0])
	}
}
