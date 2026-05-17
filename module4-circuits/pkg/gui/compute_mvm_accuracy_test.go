//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

func TestComputeMVMAccuracy_IdealAndTierA(t *testing.T) {
	patterns := []struct {
		name string
		g    [][]int
		v    []float64
	}{
		{
			name: "diagonal-dominant",
			g: [][]int{
				{29, 2, 1, 0, 0, 0, 1, 2},
				{2, 28, 2, 1, 0, 1, 2, 1},
				{1, 2, 27, 2, 1, 2, 1, 0},
				{0, 1, 2, 26, 2, 1, 0, 1},
				{0, 0, 1, 2, 25, 2, 1, 0},
				{0, 1, 2, 1, 2, 24, 2, 1},
				{1, 2, 1, 0, 1, 2, 23, 2},
				{2, 1, 0, 1, 0, 1, 2, 22},
			},
			v: []float64{0.06, 0.11, 0.03, 0.09, 0.12, 0.04, 0.07, 0.05},
		},
		{
			name: "striped-pattern",
			g: [][]int{
				{5, 20, 5, 20, 5, 20, 5, 20},
				{20, 5, 20, 5, 20, 5, 20, 5},
				{5, 20, 5, 20, 5, 20, 5, 20},
				{20, 5, 20, 5, 20, 5, 20, 5},
				{5, 20, 5, 20, 5, 20, 5, 20},
				{20, 5, 20, 5, 20, 5, 20, 5},
				{5, 20, 5, 20, 5, 20, 5, 20},
				{20, 5, 20, 5, 20, 5, 20, 5},
			},
			v: []float64{0.08, 0.05, 0.12, 0.03, 0.10, 0.06, 0.09, 0.04},
		},
		{
			name: "graded-ramp",
			g: [][]int{
				{0, 3, 6, 9, 12, 15, 18, 21},
				{2, 5, 8, 11, 14, 17, 20, 23},
				{4, 7, 10, 13, 16, 19, 22, 25},
				{1, 4, 7, 10, 13, 16, 19, 22},
				{3, 6, 9, 12, 15, 18, 21, 24},
				{5, 8, 11, 14, 17, 20, 23, 26},
				{7, 10, 13, 16, 19, 22, 25, 28},
				{9, 12, 15, 18, 21, 24, 27, 29},
			},
			v: []float64{0.02, 0.04, 0.06, 0.08, 0.10, 0.12, 0.09, 0.07},
		},
	}

	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			expected := referenceRowCurrentsUA(tc.g, tc.v, 30)

			ideal := runComputeRowCurrents(tc.g, tc.v, arraysim.CouplingIdeal)
			assertRelativeRowError(t, ideal, expected, 1e-3)

			tierA := runComputeRowCurrents(tc.g, tc.v, arraysim.CouplingTierA)
			assertRelativeRowError(t, tierA, expected, 5e-2)
		})
	}
}

func runComputeRowCurrents(weights [][]int, inputs []float64, mode arraysim.CouplingMode) []float64 {
	ds := newTestDeviceState(8, 8)
	ds.SetOperationMode(OpModeCompute)
	ds.SetWLAll()
	ds.SetCouplingMode(mode)
	for c, v := range inputs {
		ds.SetDACVoltage(c, v)
	}
	ds.Compute(weights, 30)
	out := make([]float64, 8)
	for r := 0; r < 8; r++ {
		out[r] = ds.GetRowCurrent(r)
	}
	return out
}

func referenceRowCurrentsUA(weights [][]int, inputs []float64, quantLevels int) []float64 {
	ds := newTestDeviceState(8, 8)
	mat := ds.GetMaterial()
	y := make([]float64, 8)
	for r := 0; r < 8; r++ {
		totalUA := 0.0
		for c := 0; c < 8; c++ {
			gS := mat.DiscreteLevel(weights[r][c], quantLevels)
			totalUA += gS * 1e6 * inputs[c]
		}
		y[r] = totalUA
	}
	return y
}

func assertRelativeRowError(t *testing.T, actual, expected []float64, maxRel float64) {
	t.Helper()
	for i := range expected {
		den := math.Abs(expected[i])
		if den < 1e-12 {
			den = 1e-12
		}
		rel := math.Abs(actual[i]-expected[i]) / den
		if rel > maxRel {
			t.Fatalf("row %d relative error %.6f exceeds %.6f (actual=%.6f uA expected=%.6f uA)", i, rel, maxRel, actual[i], expected[i])
		}
	}
}
