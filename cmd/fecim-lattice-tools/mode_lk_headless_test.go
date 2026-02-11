package main

import "testing"

func TestHeadlessLKRun_CompletesISPP(t *testing.T) {
	// The headless LK mode should finish the full 5-target ISPP sequence without
	// producing NaN/Inf states.
	t.Setenv("FECIM_MATERIAL", "literature_superlattice")
	t.Setenv("FECIM_RANGE_FRAC", "1")

	if err := runMode("hysteresis", "lk"); err != nil {
		t.Fatalf("headless lk run failed: %v", err)
	}
}
