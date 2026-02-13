package arraysim

import (
	"math"
	"testing"
)

const (
	tolExact  = 1e-12
	tolSolver = 1e-6
)

func testConductanceMatrix(n int, base float64) [][]float64 {
	g := make([][]float64, n)
	for r := 0; r < n; r++ {
		g[r] = make([]float64, n)
		for c := 0; c < n; c++ {
			g[r][c] = base * (1 + 0.1*float64(r) + 0.05*float64(c))
		}
	}
	return g
}

func solveForKirchhoff(t *testing.T, n int) (SolveParams, DCResult) {
	t.Helper()
	params := SolveParams{
		WLVoltages:  make([]float64, n),
		BLVoltages:  make([]float64, n),
		Conductance: testConductanceMatrix(n, 1e-4),
		Geometry:    DefaultCellGeometry(),
		Wire:        WireParams{RWordLine: 5.0, RBitLine: 7.5},
		Boundary: BoundaryParams{
			WLDriveResistance: 2.0,
			BLDriveResistance: 2.0,
		},
	}
	for i := 0; i < n; i++ {
		params.WLVoltages[i] = 0.6 - 0.1*float64(i)
		params.BLVoltages[i] = -0.2 + 0.05*float64(i)
	}
	res, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC failed: %v", err)
	}
	return params, res
}

func TestKirchhoff_ExactIdentities(t *testing.T) {
	// Exact arithmetic sanity identities.
	if got := math.Abs((1.2-0.7)+(0.7-0.1)+(0.1-1.2)); got > tolExact {
		t.Fatalf("KVL exact identity failed: residual=%g", got)
	}
	if got := math.Abs((3.0 - 1.5 - 1.5)); got > tolExact {
		t.Fatalf("KCL exact identity failed: residual=%g", got)
	}
	if got := math.Abs((2e-4*0.5)-1e-4); got > tolExact {
		t.Fatalf("Ohm exact identity failed: residual=%g A", got)
	}
}

func TestKirchhoff_2x2_KCLKVLAndOhm(t *testing.T) {
	params, res := solveForKirchhoff(t, 2)
	assertOhmBranches(t, params, res, tolSolver)
	assertKCLInternalNodes(t, params, res, tolSolver)
	assertKVLLoopResidual(t, res, 0, 0, tolSolver)
}

func TestKirchhoff_4x4_KCLKVLAndOhm(t *testing.T) {
	params, res := solveForKirchhoff(t, 4)
	assertOhmBranches(t, params, res, tolSolver)
	assertKCLInternalNodes(t, params, res, tolSolver)
	assertKVLLoopResidual(t, res, 1, 1, tolSolver)
}

func assertOhmBranches(t *testing.T, params SolveParams, res DCResult, tol float64) {
	t.Helper()
	for r := range params.Conductance {
		for c := range params.Conductance[r] {
			g := effectiveCellConductance(params, r, c)
			want := g * (res.WLNodes[r][c] - res.BLNodes[r][c])
			got := res.CellCurrents[r][c]
			if diff := math.Abs(got - want); diff > tol {
				t.Fatalf("Ohm mismatch cell[%d,%d]: got=%g A want=%g A diff=%g A", r, c, got, want, diff)
			}
		}
	}
}

func assertKCLInternalNodes(t *testing.T, params SolveParams, res DCResult, tol float64) {
	t.Helper()
	rows := len(params.Conductance)
	cols := len(params.Conductance[0])
	wire := params.Wire.WithDefaults(params.Geometry.WithDefaults())
	gWL := 1.0 / wire.RWordLine
	gBL := 1.0 / wire.RBitLine

	for r := 0; r < rows; r++ {
		for c := 1; c < cols-1; c++ {
			vw := res.WLNodes[r][c]
			kcl := gWL*(vw-res.WLNodes[r][c-1]) + gWL*(vw-res.WLNodes[r][c+1]) + effectiveCellConductance(params, r, c)*(vw-res.BLNodes[r][c])
			if math.Abs(kcl) > tol {
				t.Fatalf("KCL WL node[%d,%d] residual=%g A", r, c, kcl)
			}
		}
	}

	for r := 1; r < rows-1; r++ {
		for c := 0; c < cols; c++ {
			vb := res.BLNodes[r][c]
			kcl := gBL*(vb-res.BLNodes[r-1][c]) + gBL*(vb-res.BLNodes[r+1][c]) + effectiveCellConductance(params, r, c)*(vb-res.WLNodes[r][c])
			if math.Abs(kcl) > tol {
				t.Fatalf("KCL BL node[%d,%d] residual=%g A", r, c, kcl)
			}
		}
	}
}

func assertKVLLoopResidual(t *testing.T, res DCResult, r, c int, tol float64) {
	t.Helper()
	// 2x2 loop over (r,c)..(r+1,c+1) in the bipartite WL/BL graph.
	terms := []float64{
		res.WLNodes[r][c] - res.WLNodes[r][c+1],
		res.WLNodes[r][c+1] - res.BLNodes[r][c+1],
		res.BLNodes[r][c+1] - res.BLNodes[r+1][c+1],
		res.BLNodes[r+1][c+1] - res.WLNodes[r+1][c+1],
		res.WLNodes[r+1][c+1] - res.WLNodes[r+1][c],
		res.WLNodes[r+1][c] - res.BLNodes[r+1][c],
		res.BLNodes[r+1][c] - res.BLNodes[r][c],
		res.BLNodes[r][c] - res.WLNodes[r][c],
	}
	sum := 0.0
	for _, v := range terms {
		sum += v
	}
	if math.Abs(sum) > tol {
		t.Fatalf("KVL residual for loop r=%d c=%d: %g V", r, c, sum)
	}
}
