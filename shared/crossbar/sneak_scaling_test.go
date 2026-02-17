package crossbar

import (
	"math"
	"testing"
)

func TestSneakScaling_RatioIncreasesWithArraySize_0T1R(t *testing.T) {
	sizes := []int{4, 8, 16, 32}
	const (
		V = 1.0
		G = 50e-6
	)

	prev := -1.0
	for _, n := range sizes {
		sp := NewSneakPathAnalyzer(n, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				sp.SetConductance(i, j, G)
			}
		}
		sp.AnalyzeTarget(0, 0, V)
		r := sp.TotalSneakRatio

		want := float64((n-1)*(n-1)) / 3.0
		if e := math.Abs(r-want) / math.Max(math.Abs(want), 1e-15); e >= 1e-12 {
			t.Fatalf("N=%d sneak ratio mismatch: got %.12g want %.12g relErr=%.3g", n, r, want, e)
		}
		t.Logf("N=%d sneakRatio=%.6g (analytic %.6g)", n, r, want)

		if prev >= 0 && r <= prev {
			t.Fatalf("expected sneak ratio to increase with N: N=%d ratio=%.12g <= prev=%.12g", n, r, prev)
		}
		prev = r
	}
}

func TestSneakScaling_1T1RSelectorSuppressesSneak(t *testing.T) {
	const (
		V      = 1.0
		G      = 50e-6
		onOff  = 1e9
		absMax = 1e-6
	)

	for _, n := range []int{4, 8, 16, 32} {
		sp := NewSneakPathAnalyzer(n, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				sp.SetConductance(i, j, G)
			}
		}

		st0 := sp.AnalyzeWithMitigation(0, 0, V, SneakMitigation{})
		st1 := sp.AnalyzeWithMitigation(0, 0, V, SneakMitigation{UseSelector: true, SelectorOnOff: onOff})

		if st1.SneakRatio >= st0.SneakRatio {
			t.Fatalf("N=%d expected selector to reduce sneak ratio: 0T1R=%.12g 1T1R=%.12g", n, st0.SneakRatio, st1.SneakRatio)
		}
		if st1.SneakRatio > absMax {
			t.Fatalf("N=%d expected ~0 sneak ratio with selector: got %.12g > %.12g", n, st1.SneakRatio, absMax)
		}
		t.Logf("N=%d 0T1R_ratio=%.6g 1T1R_ratio=%.6g (suppression %.3g×)", n, st0.SneakRatio, st1.SneakRatio, st0.SneakRatio/math.Max(st1.SneakRatio, 1e-30))
	}
}
