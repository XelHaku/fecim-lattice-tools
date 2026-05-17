//go:build legacy_fyne

package gui

import "testing"

func TestBuildDesignSpaceSweep_CountAndFields(t *testing.T) {
	points := BuildDesignSpaceSweep([]int{8, 16}, []int{4, 6}, []string{"FeFET", "RRAM"})
	if got, want := len(points), 8; got != want {
		t.Fatalf("len(points) = %d, want %d", got, want)
	}
	for _, p := range points {
		if p.ArraySize <= 0 || p.ADCBits <= 0 {
			t.Fatalf("invalid sweep point: %+v", p)
		}
		if p.LatencyNS <= 0 || p.EnergyPJ <= 0 {
			t.Fatalf("non-positive metrics: %+v", p)
		}
	}
}

func TestRunProcessVariationMonteCarlo_Basic(t *testing.T) {
	stats := RunProcessVariationMonteCarlo(1e-6, 0.1, 1000, 42)
	if stats.Mean <= 0 {
		t.Fatalf("mean should be > 0, got %g", stats.Mean)
	}
	if stats.StdDev <= 0 {
		t.Fatalf("stddev should be > 0, got %g", stats.StdDev)
	}
	if stats.Min < 0 {
		t.Fatalf("min should be >= 0, got %g", stats.Min)
	}
	if stats.Max < stats.Min {
		t.Fatalf("max should be >= min, got min=%g max=%g", stats.Min, stats.Max)
	}
}
