package physics

import (
	"math"
	"testing"
)

// Extracted from module4-circuits/pkg/gui/tab_unified_voltage.go
// (legacy deterministic disturb accumulator used by UI hooks):
//
//	base := 0.01 * halfSelectDisturbRate
//	halfSelectDisturbRate = 0.25
//	disturb residue += sign(targetVCell) * base
//
// Drift by ±1 level happens when |residue| >= 1.0.
const (
	halfSelectDisturbRateFromGUI = 0.25
	halfSelectResidueStep        = 0.01 * halfSelectDisturbRateFromGUI // 0.0025 per half-select write event
)

func disturbEventsToAdjacentLevel() int {
	if halfSelectResidueStep <= 0 {
		return math.MaxInt
	}
	return int(math.Ceil(1.0 / halfSelectResidueStep))
}

func expectedRandomWritesForOneVictim(eventsToDrift, n int) int {
	// For an N x N array, excluding self-writes, a cell is half-selected when
	// writes target same row XOR same column:
	// P(half-select) = 2*(N-1)/(N^2-1) = 2/(N+1)
	p := (2.0 * float64(n-1)) / float64(n*n-1)
	return int(math.Round(float64(eventsToDrift) / p))
}

func TestHalfSelectDisturbBudgetPerLevel(t *testing.T) {
	events := disturbEventsToAdjacentLevel()
	if events != 400 {
		t.Fatalf("expected 400 events for 1-level drift with residue step %.6f, got %d", halfSelectResidueStep, events)
	}

	for level := 1; level <= 29; level++ {
		// In the current simplified residue model, budget is level-independent.
		// (No level/material dependence is encoded in the residue threshold path.)
		if events != 400 {
			t.Fatalf("level %d: expected 400 disturb events, got %d", level, events)
		}
		t.Logf("start_level=%d disturb_events_to_drift=%d note=simplified residue-counter model", level, events)
	}
}

func TestHalfSelectDisturbBudgetRandomWriteScaling(t *testing.T) {
	events := disturbEventsToAdjacentLevel()

	cases := []struct {
		n            int
		wantWrites   int
		probHalfSel  float64
		policyWrites int
	}{
		{n: 8, wantWrites: 1800, probHalfSel: 2.0 / 9.0, policyWrites: 900},
		{n: 32, wantWrites: 6600, probHalfSel: 2.0 / 33.0, policyWrites: 3300},
		{n: 64, wantWrites: 13000, probHalfSel: 2.0 / 65.0, policyWrites: 6500},
	}

	for _, tc := range cases {
		got := expectedRandomWritesForOneVictim(events, tc.n)
		if got != tc.wantWrites {
			t.Fatalf("N=%d: expected %d writes, got %d", tc.n, tc.wantWrites, got)
		}
		t.Logf("N=%dx%d P_half_select=%.6f writes_to_%d_events=%d suggested_refresh_interval=%d",
			tc.n, tc.n, tc.probHalfSel, events, got, tc.policyWrites)
	}
}
