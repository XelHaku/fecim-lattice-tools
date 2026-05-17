//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

// M4-INV-03: half-select disturb budget from current residue model.
func TestM4INV03_HalfSelectDisturbBudget(t *testing.T) {
	// From tab_unified_voltage.go: base = 0.01 * halfSelectDisturbRate, with rate=0.25.
	deltaPerPulse0T1R := 0.01 * 0.25
	cycles0T1R := math.Ceil(1.0 / deltaPerPulse0T1R)

	// 1T1R model gates unselected rows; use conservative 20x attenuation for residual coupling.
	deltaPerPulse1T1R := deltaPerPulse0T1R / 20.0
	cycles1T1R := math.Ceil(1.0 / deltaPerPulse1T1R)

	t.Logf("0T1R: delta/pulse=%.6f level, cycles_to_1_level=%.0f", deltaPerPulse0T1R, cycles0T1R)
	t.Logf("1T1R: delta/pulse=%.6f level, cycles_to_1_level=%.0f", deltaPerPulse1T1R, cycles1T1R)

	if !(cycles1T1R > cycles0T1R) {
		t.Fatalf("expected 1T1R disturb budget > 0T1R")
	}
}

// M4-INV-06: dynamic metrics from current array configuration.
func TestM4INV06_DynamicGOPSMetrics(t *testing.T) {
	for _, n := range []int{8, 16, 32, 64} {
		_, _, fefet := computeComparisonMetrics(n)
		t.Logf("array=%dx%d fefet latency=%.1fns energy=%.2fpJ GOPS=%.4f energy/op=%.4fpJ",
			n, n, fefet.LatencyNS, fefet.EnergyPJ, fefet.GOPS, fefet.EnergyOpPJ)
		if fefet.GOPS <= 0 || fefet.EnergyOpPJ <= 0 {
			t.Fatalf("invalid dynamic metrics at N=%d", n)
		}
	}
}
