package crossbar

import (
	"math"
	"testing"
)

// M2-TMP-02: Validate spatial temperature profiles (hotspot + gradient).
func TestM2_TMP_02_TemperatureProfileHotspotCenterHotterThanEdge(t *testing.T) {
	rows, cols := 9, 9
	ambient := 300.0
	delta := 20.0
	p := NewHotspotTemperatureProfile(rows, cols, ambient, delta)

	center := p.TemperatureAt(rows/2, cols/2)
	edges := []float64{
		p.TemperatureAt(0, 0),
		p.TemperatureAt(0, cols-1),
		p.TemperatureAt(rows-1, 0),
		p.TemperatureAt(rows-1, cols-1),
	}

	minEdge := edges[0]
	for _, e := range edges[1:] {
		if e < minEdge {
			minEdge = e
		}
	}

	if center <= minEdge {
		t.Fatalf("expected hotspot center > edge; center=%.2fK minEdge=%.2fK", center, minEdge)
	}
	if center < ambient+delta-1e-9 {
		t.Fatalf("expected center ~= ambient+delta; center=%.3fK ambient+delta=%.3fK", center, ambient+delta)
	}

	// Acceptance from plan: center hotter than edge by > 5K.
	if center-minEdge <= 5.0 {
		t.Fatalf("expected center-edge delta > 5K; center=%.2fK edge=%.2fK delta=%.2fK", center, minEdge, center-minEdge)
	}
	t.Logf("hotspot: center=%.2fK minEdge=%.2fK delta=%.2fK", center, minEdge, center-minEdge)
}

func TestM2_TMP_02_TemperatureProfileGradientLinearAlongAxis(t *testing.T) {
	rows, cols := 5, 11
	startK := 300.0
	endK := 410.0
	p := NewGradientTemperatureProfile(rows, cols, startK, endK, "x")

	if got := p.TemperatureAt(0, 0); math.Abs(got-startK) > 1e-12 {
		t.Fatalf("expected startK at col0; got %.6f want %.6f", got, startK)
	}
	if got := p.TemperatureAt(0, cols-1); math.Abs(got-endK) > 1e-12 {
		t.Fatalf("expected endK at last col; got %.6f want %.6f", got, endK)
	}

	r := rows / 2
	step := (endK - startK) / float64(cols-1)
	for c := 0; c < cols; c++ {
		want := startK + float64(c)*step
		got := p.TemperatureAt(r, c)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("nonlinear gradient at (r=%d,c=%d): got %.6f want %.6f (step=%.6f)", r, c, got, want, step)
		}
	}

	if math.Abs(p.TemperatureAt(0, cols/2)-p.TemperatureAt(rows-1, cols/2)) > 1e-12 {
		t.Fatalf("expected x-gradient to be uniform across rows")
	}

	t.Logf("gradient-x: start=%.1fK end=%.1fK step=%.3fK/col", startK, endK, step)
}
