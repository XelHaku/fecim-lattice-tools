package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

// M2-PV-02: Monte Carlo MVM spread should increase with NoiseLevel.
func TestM2_PV_02_MVMSigmaIncreasesWithNoiseLevel(t *testing.T) {
	rows, cols := 8, 8
	input := make([]float64, cols)
	for j := range input {
		input[j] = 0.7
	}

	weights := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		weights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			weights[i][j] = 0.6 + 0.2*float64((i+j)%2)
		}
	}

	sigmas := []float64{0.01, 0.05, 0.10}
	avgStd := make([]float64, len(sigmas))

	const runs = 100
	for idx, sigma := range sigmas {
		outputs := make([][]float64, runs)
		for r := 0; r < runs; r++ {
			rand.Seed(int64(1000 + 17*r))
			cfg := &Config{Rows: rows, Cols: cols, NoiseLevel: sigma, ADCBits: 12, DACBits: 12}
			arr, err := NewArray(cfg)
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
			outputs[r] = append([]float64(nil), res.ActualOutput...)
		}

		var sumStd float64
		for i := 0; i < rows; i++ {
			col := make([]float64, runs)
			for r := 0; r < runs; r++ {
				col[r] = outputs[r][i]
			}
			_, s := meanStd(col)
			sumStd += s
		}
		avgStd[idx] = sumStd / float64(rows)
		t.Logf("sigma=%.3f => avg element std=%.6g", sigma, avgStd[idx])
	}

	if !(avgStd[0] < avgStd[1] && avgStd[1] < avgStd[2]) {
		t.Fatalf("expected monotonic increase in output spread: std(0.01)=%.6g std(0.05)=%.6g std(0.10)=%.6g", avgStd[0], avgStd[1], avgStd[2])
	}
	if avgStd[2] <= 0 {
		t.Fatalf("expected nonzero spread at sigma=0.10")
	}
}

func TestM2_PV_02_MVMMomentsAreFinite(t *testing.T) {
	rand.Seed(2026)
	cfg := &Config{Rows: 4, Cols: 4, NoiseLevel: 0.10, ADCBits: 10, DACBits: 10}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			_ = arr.ProgramWeight(i, j, 0.5)
		}
	}
	input := []float64{0.3, 0.7, 0.9, 0.2}
	res, err := arr.MVMWithNonIdealities(input, &MVMOptions{EnableVariation: true, Temperature: 300})
	if err != nil {
		t.Fatalf("MVMWithNonIdealities: %v", err)
	}
	for i, v := range res.ActualOutput {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("output[%d] is NaN/Inf: %v", i, v)
		}
	}
}
