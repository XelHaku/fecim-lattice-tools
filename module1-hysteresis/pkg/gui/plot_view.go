//go:build legacy_fyne

package gui

import "math"

// PlotViewMode controls how the P–E plot is derived from the recorded history.
// It is presentation-only: it does not change the simulation, controller logic,
// or CSV logging.
type PlotViewMode int

const (
	PlotViewFullHistory PlotViewMode = iota
	PlotViewLastCycle
)

// trimTail keeps at most max points, preserving the newest samples.
func trimTail(hE, hP []float64, max int) ([]float64, []float64) {
	if max <= 0 || len(hE) != len(hP) {
		return hE, hP
	}
	if len(hE) <= max {
		return hE, hP
	}
	start := len(hE) - max
	return hE[start:], hP[start:]
}

// lastCompleteCycle tries to extract the last complete drive cycle from a
// time-ordered history by slicing between two successive cycle markers.
//
// Heuristic:
//  1. Prefer last two local maxima of E (peak-to-peak).
//  2. Fallback: rising zero-crossing to rising zero-crossing.
//
// If it cannot infer a stable cycle, it returns the input unchanged.
func lastCompleteCycle(hE, hP []float64) ([]float64, []float64) {
	if len(hE) < 10 || len(hE) != len(hP) {
		return hE, hP
	}

	// 1) Peak-to-peak: last two local maxima.
	maxima := make([]int, 0, 8)
	for i := 1; i < len(hE)-1; i++ {
		if hE[i] >= hE[i-1] && hE[i] >= hE[i+1] {
			maxima = append(maxima, i)
		}
	}
	if len(maxima) >= 2 {
		i2 := maxima[len(maxima)-1]
		i1 := maxima[len(maxima)-2]
		if i2-i1 >= 5 {
			return hE[i1:i2], hP[i1:i2]
		}
	}

	// 2) Rising zero-crossings.
	zc := make([]int, 0, 8)
	for i := 1; i < len(hE); i++ {
		if hE[i-1] <= 0 && hE[i] > 0 {
			zc = append(zc, i)
		}
	}
	if len(zc) >= 2 {
		i2 := zc[len(zc)-1]
		i1 := zc[len(zc)-2]
		if i2-i1 >= 5 {
			return hE[i1:i2], hP[i1:i2]
		}
	}

	// If E is basically constant/degenerate, do nothing.
	minE, maxE := hE[0], hE[0]
	for _, e := range hE[1:] {
		if e < minE {
			minE = e
		}
		if e > maxE {
			maxE = e
		}
	}
	if math.Abs(maxE-minE) < 1e-12 {
		return hE, hP
	}

	return hE, hP
}
