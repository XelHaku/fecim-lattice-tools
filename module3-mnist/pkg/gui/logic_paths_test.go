package gui

import (
	"image"
	"image/color"
	"testing"
)

func TestMaxWithConfidence(t *testing.T) {
	idx, val := maxWithConfidence([]float64{0.1, 0.4, 0.3})
	if idx != 1 || val != 0.4 {
		t.Fatalf("expected (1,0.4), got (%d,%.3f)", idx, val)
	}
	idx, val = maxWithConfidence([]float64{-1, -0.2})
	if idx != 0 || val != 0 {
		t.Fatalf("expected default zero baseline behavior, got (%d,%.3f)", idx, val)
	}
}

func TestComparisonCard_GetSecondBest(t *testing.T) {
	cc := &ComparisonCard{}
	idx, val := cc.getSecondBest([]float64{0.1, 0.7, 0.4, 0.2})
	if idx != 2 || val != 0.4 {
		t.Fatalf("expected second best (2,0.4), got (%d,%.3f)", idx, val)
	}
	idx, val = cc.getSecondBest([]float64{0.9})
	if idx != -1 || val != 0 {
		t.Fatalf("expected short-slice guard, got (%d,%.3f)", idx, val)
	}
}

func TestWeightComparisonWidget_StatsAndColors(t *testing.T) {
	w := &WeightComparisonWidget{}
	w.fpWeights = [][]float64{{-1, 0}, {0.5, 1}}
	w.quantWeights = [][]float64{{-0.8, -0.1}, {0.4, 0.6}}
	w.calculateStats()
	if w.meanError <= 0 || w.maxError <= 0 {
		t.Fatalf("expected positive errors, got mean=%.4f max=%.4f", w.meanError, w.maxError)
	}

	b := w.blueWhiteRedColor(0.0)
	m := w.blueWhiteRedColor(0.5)
	r := w.blueWhiteRedColor(1.0)
	if b.B != 255 || m.R != 255 || r.G != 0 {
		t.Fatalf("unexpected blue-white-red mapping: b=%v m=%v r=%v", b, m, r)
	}

	e0 := w.errorColor(0.0)
	e1 := w.errorColor(1.0)
	if e0.R != 0 || e1.R != 255 || e1.G != 0 {
		t.Fatalf("unexpected error color mapping: e0=%v e1=%v", e0, e1)
	}
}

func TestComparisonCard_DrawHelpers(t *testing.T) {
	cc := &ComparisonCard{}
	img := image.NewRGBA(image.Rect(0, 0, 30, 30))
	cc.drawScaledDigit(img, 2, 2, "5", color.RGBA{255, 0, 0, 255}, 2)
	cc.drawCircle(img, 15, 15, 3, color.RGBA{0, 255, 0, 255})

	// ensure pixels got painted (not all zero)
	var painted bool
	for y := 0; y < 30 && !painted; y++ {
		for x := 0; x < 30; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r != 0 || g != 0 || b != 0 {
				painted = true
				break
			}
		}
	}
	if !painted {
		t.Fatal("expected helper draw functions to paint pixels")
	}
}
