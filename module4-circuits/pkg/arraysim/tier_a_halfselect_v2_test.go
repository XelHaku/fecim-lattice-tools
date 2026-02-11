package arraysim

import (
	"math"
	"testing"
)

func TestTierA_V2HalfSelectVoltageDistribution_NoIRDrop(t *testing.T) {
	tests := []struct {
		name      string
		writeV    float64
		wlSelected float64
		blSelected float64
		wantTarget float64
		wantHalf   float64
	}{
		{
			name:       "SET_positive",
			writeV:     1.0,
			wlSelected: +0.5,
			blSelected: -0.5,
			wantTarget: +1.0,
			wantHalf:   +0.5,
		},
		{
			name:       "ERASE_negative",
			writeV:     1.0,
			wlSelected: -0.5,
			blSelected: +0.5,
			wantTarget: -1.0,
			wantHalf:   -0.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			solver := NewTierASolver()
			rows, cols := 3, 3

			conductance := make([][]float64, rows)
			for r := 0; r < rows; r++ {
				conductance[r] = make([]float64, cols)
				for c := 0; c < cols; c++ {
					conductance[r][c] = 1e-3
				}
			}

			wl := make([]float64, rows)
			bl := make([]float64, cols)
			selectedRow, selectedCol := 1, 1
			wl[selectedRow] = tc.wlSelected
			bl[selectedCol] = tc.blSelected

			res, err := solver.Solve(SolveParams{
				WLVoltages:  wl,
				BLVoltages:  bl,
				Conductance: conductance,
				ActiveRows:  []bool{true, true, true},
				Wire:        WireParams{RWordLine: 1e-18, RBitLine: 1e-18},
			})
			if err != nil {
				t.Fatalf("Solve returned error: %v", err)
			}

			eps := 1e-6
			get := func(r, c int) float64 { return res.CellVoltages[r][c] }

			// Target cell: full write voltage
			if math.Abs(get(selectedRow, selectedCol)-tc.wantTarget) > eps {
				t.Fatalf("target Vcell: got %.9f, want %.9f", get(selectedRow, selectedCol), tc.wantTarget)
			}

			// Half-selected: same row (different columns)
			for c := 0; c < cols; c++ {
				if c == selectedCol {
					continue
				}
				if math.Abs(get(selectedRow, c)-tc.wantHalf) > eps {
					t.Fatalf("half-select same row cell (%d,%d): got %.9f, want %.9f", selectedRow, c, get(selectedRow, c), tc.wantHalf)
				}
			}

			// Half-selected: same column (different rows)
			for r := 0; r < rows; r++ {
				if r == selectedRow {
					continue
				}
				if math.Abs(get(r, selectedCol)-tc.wantHalf) > eps {
					t.Fatalf("half-select same col cell (%d,%d): got %.9f, want %.9f", r, selectedCol, get(r, selectedCol), tc.wantHalf)
				}
			}

			// Diagonal/unselected
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					if r == selectedRow || c == selectedCol {
						continue
					}
					if math.Abs(get(r, c)-0) > eps {
						t.Fatalf("unselected cell (%d,%d): got %.9f, want 0", r, c, get(r, c))
					}
				}
			}
		})
	}
}
