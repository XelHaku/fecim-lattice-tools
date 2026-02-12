package arraysim

import (
	"math"
	"testing"
)

func TestSelectorMasks_ReadVsWritePath(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{1.0, 1.0},
		BLVoltages: []float64{0.0, 0.0},
		Conductance: [][]float64{
			{10e-6, 10e-6},
			{10e-6, 10e-6},
		},
		Wire: WireParams{RWordLine: 0.2, RBitLine: 0.2},
		ReadMask: [][]bool{
			{true, false},
			{false, false},
		},
		WriteMask: [][]bool{
			{false, false},
			{false, true},
		},
	}

	// Read path: only (0,0) enabled.
	params.SelectorMode = SelectorRead
	readRes, err := NewTierBSolver().Solve(params)
	if err != nil {
		t.Fatalf("read solve: %v", err)
	}
	if !(readRes.CellCurrents[0][0] > 0) {
		t.Fatalf("read path expected cell (0,0) current > 0, got %.3g", readRes.CellCurrents[0][0])
	}
	if math.Abs(readRes.CellCurrents[1][1]) > 1e-15 {
		t.Fatalf("read path expected cell (1,1) off, got %.3g", readRes.CellCurrents[1][1])
	}

	// Write path: only (1,1) enabled.
	params.SelectorMode = SelectorWrite
	writeRes, err := NewTierBSolver().Solve(params)
	if err != nil {
		t.Fatalf("write solve: %v", err)
	}
	if !(writeRes.CellCurrents[1][1] > 0) {
		t.Fatalf("write path expected cell (1,1) current > 0, got %.3g", writeRes.CellCurrents[1][1])
	}
	if math.Abs(writeRes.CellCurrents[0][0]) > 1e-15 {
		t.Fatalf("write path expected cell (0,0) off, got %.3g", writeRes.CellCurrents[0][0])
	}
}
