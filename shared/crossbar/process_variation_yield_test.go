package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

// M2-PV-03: Yield (fraction of outputs with <10% error vs ideal) should decrease
// as process variation sigma increases.
func TestM2_PV_03_YieldDecreasesWithNoiseLevel(t *testing.T) {
	rows, cols := 8, 8
	input := make([]float64, cols)
	for j := range input {
		input[j] = 1.0
	}

	weights := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		weights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			weights[i][j] = 0.85
		}
	}

	// Ideal (no variation) baseline.
	rand.Seed(1)
	idealArr, err := NewArray(&Config{Rows: rows, Cols: cols, NoiseLevel: 0, ADCBits: 12, DACBits: 12})
	if err != nil {
		t.Fatalf("NewArray(ideal): %v", err)
	}
	if err := idealArr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix: %v", err)
	}
	idealRes, err := idealArr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: false, Temperature: 300})
	if err != nil {
		t.Fatalf("ideal MVM: %v", err)
	}

	threshold := 0.10 // 10% relative error per element
	sigmas := []float64{0.01, 0.10, 0.30}
	yields := make([]float64, len(sigmas))

	const runs = 200
	for idx, sigma := range sigmas {
		pass := 0
		for r := 0; r < runs; r++ {
			rand.Seed(int64(9000 + 31*r))
			arr, err := NewArray(&Config{Rows: rows, Cols: cols, NoiseLevel: sigma, ADCBits: 12, DACBits: 12})
			if err != nil {
				t.Fatalf("NewArray(sigma=%.3f): %v", sigma, err)
			}
			if err := arr.ProgramWeightMatrix(weights); err != nil {
				t.Fatalf("ProgramWeightMatrix: %v", err)
			}

			res, err := arr.MVMWithNonIdealities(input, &MVMOptions{EnableIRDrop: false, EnableSneakPaths: false, EnableVariation: true, Temperature: 300})
			if err != nil {
				t.Fatalf("MVMWithNonIdealities: %v", err)
			}

			if withinRelError(res.ActualOutput, idealRes.ActualOutput, threshold) {
				pass++
			}
		}
		yields[idx] = float64(pass) / float64(runs)
		t.Logf("yield @ sigma=%.3f: %.3f (%d/%d)", sigma, yields[idx], pass, runs)
	}

	if !(yields[0] > yields[1] && yields[1] > yields[2]) {
		t.Fatalf("expected yield to strictly decrease with sigma: y(0.01)=%.3f y(0.10)=%.3f y(0.30)=%.3f", yields[0], yields[1], yields[2])
	}

	// Sanity: avoid degenerate all-0 or all-1 yields.
	if yields[0] < 0.5 {
		t.Fatalf("unexpectedly low yield at sigma=0.01: %.3f", yields[0])
	}
	if yields[2] > 0.50 {
		t.Fatalf("unexpectedly high yield at sigma=0.30: %.3f", yields[2])
	}
}

func withinRelError(got, want []float64, thresh float64) bool {
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	if n == 0 {
		return true
	}
	const eps = 1e-12
	for i := 0; i < n; i++ {
		den := math.Abs(want[i])
		if den < eps {
			den = eps
		}
		rel := math.Abs(got[i]-want[i]) / den
		if rel >= thresh {
			return false
		}
	}
	return true
}
