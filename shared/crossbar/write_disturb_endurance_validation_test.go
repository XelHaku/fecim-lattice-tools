package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

// M2-WRD-04: Endurance degradation curve: Pr (proxy via conductance window) degrades with cycles.
func TestM2WRD04_Endurance_PrDegradesWithCycles(t *testing.T) {
	rand.Seed(6)

	endCfg := &EnduranceConfig{Enabled: true, FatigueThreshold: 1_000, FailureThreshold: 100_000}
	cfg := &Config{Rows: 1, Cols: 1, NoiseLevel: 0, ADCBits: 8, DACBits: 8, ConductanceModel: ConductanceLinear, Endurance: endCfg}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	row, col := 0, 0
	if err := arr.ProgramWeight(row, col, 1.0); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}

	gMid := (GMin + GMax) / 2
	measureWindow := func() float64 {
		g := arr.GetPhysicalConductanceForCell(row, col)
		return math.Abs(g - gMid)
	}

	w0 := measureWindow()
	cycleTo := func(target int64) {
		for arr.cells[row][col].SwitchingCount < target {
			_ = arr.ProgramWeight(row, col, 1.0)
		}
	}

	cycleTo(1_000)
	w1 := measureWindow()
	cycleTo(10_000)
	w2 := measureWindow()
	cycleTo(100_000)
	w3 := measureWindow()

	if !(w1 <= w0) {
		t.Fatalf("expected window not to increase at 1e3 cycles: w0=%g w1=%g", w0, w1)
	}
	if !(w2 <= w1) {
		t.Fatalf("expected window non-increasing at 1e4 cycles: w1=%g w2=%g", w1, w2)
	}
	if !(w3 <= w2) {
		t.Fatalf("expected window non-increasing at 1e5 cycles: w2=%g w3=%g", w2, w3)
	}
	if !(w3 < 0.9*w0) {
		t.Fatalf("expected >=10%% degradation by 1e5 cycles: w0=%g w3=%g", w0, w3)
	}

	t.Logf("M2-WRD-04 conductance window |G-Gmid| shrinks with cycles: 1=%g, 1e3=%g, 1e4=%g, 1e5=%g", w0, w1, w2, w3)
}
