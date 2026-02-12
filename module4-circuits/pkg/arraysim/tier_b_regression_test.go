package arraysim

import (
	"math"
	"math/rand"
	"testing"
)

func TestTierBDCRegression_AgreesWithDenseOracle(t *testing.T) {
	tests := []struct {
		name string
		rows int
		cols int
		seed int64
	}{
		{"2x2_seed1", 2, 2, 1},
		{"3x4_seed2", 3, 4, 2},
		{"4x3_seed3", 4, 3, 3},
		{"5x5_seed4", 5, 5, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(tt.seed))
			params := SolveParams{
				WLVoltages:  make([]float64, tt.rows),
				BLVoltages:  make([]float64, tt.cols),
				Conductance: make([][]float64, tt.rows),
				ActiveRows:  make([]bool, tt.rows),
				Wire: WireParams{
					RWordLine: 0.5 + rng.Float64()*1.5,
					RBitLine:  0.5 + rng.Float64()*1.5,
				},
			}
			for r := 0; r < tt.rows; r++ {
				params.WLVoltages[r] = -0.2 + 1.4*rng.Float64()
				params.ActiveRows[r] = rng.Float64() > 0.2
				params.Conductance[r] = make([]float64, tt.cols)
				for c := 0; c < tt.cols; c++ {
					params.Conductance[r][c] = (0.2 + 2.8*rng.Float64()) * 1e-6
				}
			}
			for c := 0; c < tt.cols; c++ {
				params.BLVoltages[c] = -0.1 + 0.7*rng.Float64()
			}

			want, err := referenceSolveDense(params)
			if err != nil {
				t.Fatalf("dense oracle: %v", err)
			}
			got, err := NewTierBSolver().SolveDC(params)
			if err != nil {
				t.Fatalf("tier-b: %v", err)
			}

			assertMaxDeltaGrid(t, "cellV", want.CellVoltages, got.CellVoltages, 2e-8)
			assertMaxDeltaGrid(t, "cellI", want.CellCurrents, got.CellCurrents, 2e-8)
			assertMaxDeltaGrid(t, "wl", want.WLNodes, got.WLNodes, 2e-8)
			assertMaxDeltaGrid(t, "bl", want.BLNodes, got.BLNodes, 2e-8)
		})
	}
}

func assertMaxDeltaGrid(t *testing.T, name string, a, b [][]float64, tol float64) {
	t.Helper()
	for r := range a {
		for c := range a[r] {
			d := math.Abs(a[r][c] - b[r][c])
			if d > tol {
				t.Fatalf("%s[%d][%d] delta=%g exceeds tol=%g", name, r, c, d, tol)
			}
		}
	}
}
