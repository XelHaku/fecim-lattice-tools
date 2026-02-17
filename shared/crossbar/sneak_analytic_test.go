package crossbar

import (
	"math"
	"testing"
)

func TestSneakAnalytic_2x2Target00(t *testing.T) {
	sp := NewSneakPathAnalyzer(2, 2)
	const (
		V = 1.0
		G = 50e-6
	)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			sp.SetConductance(i, j, G)
		}
	}

	sp.AnalyzeTarget(0, 0, V)
	st := sp.GetStats(V)

	// For 2x2 there is exactly one 3-cell path: (0,1)->(1,1)->(1,0).
	wantSneak := V * (G / 3.0)
	wantTarget := V * G
	wantRatio := wantSneak / wantTarget // = 1/3

	if st.NumSneakPaths != 1 {
		t.Fatalf("expected 1 sneak path; got %d", st.NumSneakPaths)
	}
	if e := math.Abs(st.TotalSneakCurrent-wantSneak) / math.Max(math.Abs(wantSneak), 1e-15); e >= 1e-12 {
		t.Fatalf("total sneak current mismatch: got %.12g want %.12g relErr=%.3g", st.TotalSneakCurrent, wantSneak, e)
	}
	if e := math.Abs(st.TargetCurrent-wantTarget) / math.Max(math.Abs(wantTarget), 1e-15); e >= 1e-12 {
		t.Fatalf("target current mismatch: got %.12g want %.12g relErr=%.3g", st.TargetCurrent, wantTarget, e)
	}
	if e := math.Abs(st.SneakRatio-wantRatio) / math.Max(math.Abs(wantRatio), 1e-15); e >= 1e-12 {
		t.Fatalf("sneak ratio mismatch: got %.12g want %.12g relErr=%.3g", st.SneakRatio, wantRatio, e)
	}
	if math.Abs(st.WorstSneakPath-wantSneak) > 1e-18 {
		t.Fatalf("worst sneak mismatch: got %.12g want %.12g", st.WorstSneakPath, wantSneak)
	}
}
