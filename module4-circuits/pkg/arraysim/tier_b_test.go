package arraysim

import (
	"math"
	"testing"
)

func TestNewTierBSolver_Defaults(t *testing.T) {
	s := NewTierBSolver()
	if s == nil {
		t.Fatal("nil solver")
	}
	if s.MaxIter != 4000 {
		t.Fatalf("MaxIter=%d want 4000", s.MaxIter)
	}
	if s.RelativeTolerance != 1e-8 {
		t.Fatalf("RelativeTolerance=%g want 1e-8", s.RelativeTolerance)
	}
	if s.AbsoluteTolerance != 1e-12 {
		t.Fatalf("AbsoluteTolerance=%g want 1e-12", s.AbsoluteTolerance)
	}
}

func TestTierBSolver_SolveDC_EmptyInput(t *testing.T) {
	s := NewTierBSolver()
	res, err := s.SolveDC(SolveParams{})
	if err != nil {
		t.Fatalf("SolveDC error: %v", err)
	}
	if len(res.CellVoltages) != 0 {
		t.Fatalf("CellVoltages len=%d want 0", len(res.CellVoltages))
	}
}

func TestTierBSolver_SolveDC_MatchesDenseReference_4x4(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{1.0, 0.9, 0.4, 0.0},
		BLVoltages: []float64{0.0, 0.1, 0.2, 0.3},
		Conductance: [][]float64{
			{1e-6, 2e-6, 0.5e-6, 3e-6},
			{1.5e-6, 0.8e-6, 2.2e-6, 1.1e-6},
			{3.0e-6, 1.2e-6, 0.9e-6, 2.8e-6},
			{0.6e-6, 1.7e-6, 2.4e-6, 0.4e-6},
		},
		ActiveRows: []bool{true, true, false, true},
		Wire:       WireParams{RWordLine: 1.1, RBitLine: 1.3},
	}

	want, err := referenceSolveDense(params)
	if err != nil {
		t.Fatalf("dense solve: %v", err)
	}
	got, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("tier-b solve: %v", err)
	}

	const tol = 2e-8
	compareGrid := func(name string, a, b [][]float64) {
		t.Helper()
		if len(a) != len(b) {
			t.Fatalf("%s rows: %d vs %d", name, len(a), len(b))
		}
		for r := range a {
			if len(a[r]) != len(b[r]) {
				t.Fatalf("%s cols row %d: %d vs %d", name, r, len(a[r]), len(b[r]))
			}
			for c := range a[r] {
				if math.Abs(a[r][c]-b[r][c]) > tol {
					t.Fatalf("%s[%d][%d]=%.12g want %.12g (|Δ|=%.3g)", name, r, c, b[r][c], a[r][c], math.Abs(a[r][c]-b[r][c]))
				}
			}
		}
	}
	compareVec := func(name string, a, b []float64) {
		t.Helper()
		if len(a) != len(b) {
			t.Fatalf("%s len: %d vs %d", name, len(a), len(b))
		}
		for i := range a {
			if math.Abs(a[i]-b[i]) > tol {
				t.Fatalf("%s[%d]=%.12g want %.12g", name, i, b[i], a[i])
			}
		}
	}

	compareGrid("WLNodes", want.WLNodes, got.WLNodes)
	compareGrid("BLNodes", want.BLNodes, got.BLNodes)
	compareGrid("CellVoltages", want.CellVoltages, got.CellVoltages)
	compareGrid("CellCurrents", want.CellCurrents, got.CellCurrents)
	compareVec("RowCurrents", want.RowCurrents, got.RowCurrents)
	compareVec("ColCurrents", want.ColCurrents, got.ColCurrents)
}

func TestTierBSolver_SolveDC_LargeArrayConverges(t *testing.T) {
	rows, cols := 64, 64
	cond := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		cond[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			cond[r][c] = 0.8e-6 + float64((r+c)%7)*0.2e-6
		}
	}
	params := SolveParams{
		WLVoltages:  make([]float64, rows),
		BLVoltages:  make([]float64, cols),
		Conductance: cond,
	}
	for r := 0; r < rows; r++ {
		params.WLVoltages[r] = 1.0 - 0.6*float64(r)/float64(rows-1)
	}
	for c := 0; c < cols; c++ {
		params.BLVoltages[c] = 0.2 * float64(c) / float64(cols-1)
	}

	res, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC large array: %v", err)
	}
	if len(res.WLNodes) != rows || len(res.WLNodes[0]) != cols {
		t.Fatalf("WLNodes size mismatch")
	}
	if len(res.BLNodes) != rows || len(res.BLNodes[0]) != cols {
		t.Fatalf("BLNodes size mismatch")
	}
}

func TestTierBSolver_BoundaryConvention_ZeroLoadTracksDrivers(t *testing.T) {
	params := SolveParams{
		WLVoltages: []float64{1.0, 0.7, 0.4},
		BLVoltages: []float64{0.0, 0.2, 0.3, 0.5},
		Conductance: [][]float64{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		Wire: WireParams{RWordLine: 0.5, RBitLine: 0.5},
	}
	res, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC: %v", err)
	}
	for r := range res.WLNodes {
		for c := range res.WLNodes[r] {
			if math.Abs(res.WLNodes[r][c]-params.WLVoltages[r]) > 1e-10 {
				t.Fatalf("WL node mismatch r=%d c=%d got=%.12g want=%.12g", r, c, res.WLNodes[r][c], params.WLVoltages[r])
			}
			if math.Abs(res.BLNodes[r][c]-params.BLVoltages[c]) > 1e-10 {
				t.Fatalf("BL node mismatch r=%d c=%d got=%.12g want=%.12g", r, c, res.BLNodes[r][c], params.BLVoltages[c])
			}
		}
	}
}

func TestTierBSolver_Interface(t *testing.T) {
	var _ DCEngine = (*TierBSolver)(nil)
	var _ Engine = (*TierBSolver)(nil)
}
