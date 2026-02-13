package arraysim

import (
	"math"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestCurrentValidation_SingleCell_Exact50uA(t *testing.T) {
	g := 1e-4
	v := 0.5
	i := g * v
	want := 50e-6
	if math.Abs(i-want) > tolExact {
		t.Fatalf("single-cell current mismatch: got=%g A want=%g A", i, want)
	}
}

func TestCurrentValidation_2x2_AnalyticParallelRows(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{0.5, 0.25},
		BLVoltages: []float64{0, 0},
		Conductance: [][]float64{
			{1e-4, 2e-4},
			{3e-4, 4e-4},
		},
		Wire: WireParams{RWordLine: 1e-9, RBitLine: 1e-9},
		Boundary: BoundaryParams{WLDriveResistance: 1e-9, BLDriveResistance: 1e-9},
	}
	res, err := NewTierBSolver().Solve(params)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}
	wantRow0 := 0.5 * (1e-4 + 2e-4)
	wantRow1 := 0.25 * (3e-4 + 4e-4)
	if diff := math.Abs(res.RowCurrents[0] - wantRow0); diff > 1e-6 {
		t.Fatalf("row0 current mismatch: got=%g A want=%g A", res.RowCurrents[0], wantRow0)
	}
	if diff := math.Abs(res.RowCurrents[1] - wantRow1); diff > 1e-6 {
		t.Fatalf("row1 current mismatch: got=%g A want=%g A", res.RowCurrents[1], wantRow1)
	}
}

func TestCurrentValidation_4x4_KCLAndBounds(t *testing.T) {
	params, res := solveForKirchhoff(t, 4)
	assertKCLInternalNodes(t, params, res, tolSolver)
	for r := range res.RowCurrents {
		if math.Abs(res.RowCurrents[r]) > 5e-3 {
			t.Fatalf("row current out of bound row=%d current=%g A", r, res.RowCurrents[r])
		}
	}
}

func TestCurrentValidation_WriteSequence_MonotonicConductance(t *testing.T) {
	ispp := sharedphysics.NewISPPCalculator(1.0, 30)
	g := 20e-6
	step := 5e-6
	targetLevel := 20
	currentLevel := 10
	for pulse := 1; pulse <= 3; pulse++ {
		voltage := ispp.CalculateNextVoltage(0.7, sharedphysics.DirectionAscending)
		if voltage <= 0 {
			t.Fatalf("pulse %d invalid voltage=%g V", pulse, voltage)
		}
		newG := g + step
		if newG < g {
			t.Fatalf("non-monotonic conductance at pulse %d: prev=%g S new=%g S", pulse, g, newG)
		}
		g = newG
		currentLevel++
		_ = ispp.CheckResult(currentLevel, targetLevel, sharedphysics.DirectionAscending, pulse)
	}
}

func TestCurrentValidation_OperationInvariants(t *testing.T) {
	base := [][]float64{{1e-4, 2e-4}, {3e-4, 4e-4}}
	clone := func(src [][]float64) [][]float64 {
		d := make([][]float64, len(src))
		for i := range src {
			d[i] = append([]float64(nil), src[i]...)
		}
		return d
	}
	readState := clone(base)
	computeState := clone(base)
	writeState := clone(base)

	readParams := SolveParams{WLVoltages: []float64{0.3, 0.0}, BLVoltages: []float64{0, 0}, Conductance: readState}
	_, _ = NewTierASolver().Solve(readParams)
	if !same2D(readState, base, 0) {
		t.Fatal("READ mutated cell state")
	}

	computeParams := SolveParams{WLVoltages: []float64{0.4, 0.4}, BLVoltages: []float64{0, 0}, Conductance: computeState}
	_, _ = NewTierASolver().Solve(computeParams)
	if !same2D(computeState, base, 0) {
		t.Fatal("COMPUTE mutated cell state")
	}

	targetR, targetC := 0, 1
	writeState[targetR][targetC] += 1e-5
	for r := range writeState {
		for c := range writeState[r] {
			if r == targetR && c == targetC {
				if writeState[r][c] == base[r][c] {
					t.Fatal("WRITE did not mutate targeted cell")
				}
				continue
			}
			if writeState[r][c] != base[r][c] {
				t.Fatalf("WRITE mutated non-target cell [%d,%d]", r, c)
			}
		}
	}
}

func same2D(a, b [][]float64, tol float64) bool {
	if len(a) != len(b) {
		return false
	}
	for r := range a {
		if len(a[r]) != len(b[r]) {
			return false
		}
		for c := range a[r] {
			if math.Abs(a[r][c]-b[r][c]) > tol {
				return false
			}
		}
	}
	return true
}
