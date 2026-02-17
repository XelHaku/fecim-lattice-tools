package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

// M2-TMP-03: For the same array + input, MVM accuracy relative to an ideal
// baseline should be worse at higher temperature.
func TestM2_TMP_03_MVMErrorIncreasesFrom300KTo400K(t *testing.T) {
	rand.Seed(7)
	cfg := &Config{Rows: 16, Cols: 16, NoiseLevel: 0.03, ADCBits: 12, DACBits: 12}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			_ = arr.ProgramWeight(i, j, 0.8)
		}
	}
	input := make([]float64, cfg.Cols)
	for j := range input {
		input[j] = 0.9
	}

	ideal, err := arr.MVMWithNonIdealities(input, &MVMOptions{
		EnableIRDrop:     false,
		EnableSneakPaths: false,
		EnableVariation:  false,
		EnableDrift:      false,
		Temperature:      300,
	})
	if err != nil {
		t.Fatalf("ideal MVM: %v", err)
	}

	prof := &TemperatureProfile{Enable: true, ApplyConductanceWindow: true, ApplyDrift: true}
	res300, err := arr.MVMWithNonIdealities(input, &MVMOptions{
		EnableIRDrop:       false,
		EnableSneakPaths:   false,
		EnableVariation:    false,
		EnableDrift:        true,
		Temperature:        300,
		TemperatureProfile: prof,
	})
	if err != nil {
		t.Fatalf("MVM 300K: %v", err)
	}
	res400, err := arr.MVMWithNonIdealities(input, &MVMOptions{
		EnableIRDrop:       false,
		EnableSneakPaths:   false,
		EnableVariation:    false,
		EnableDrift:        true,
		Temperature:        400,
		TemperatureProfile: prof,
	})
	if err != nil {
		t.Fatalf("MVM 400K: %v", err)
	}

	err300 := rmse(res300.ActualOutput, ideal.ActualOutput)
	err400 := rmse(res400.ActualOutput, ideal.ActualOutput)
	if !(err400 > err300) {
		t.Fatalf("expected higher-T error > lower-T error; err300=%.6g err400=%.6g", err300, err400)
	}
	if err400-err300 < 1e-6 {
		t.Fatalf("expected measurable error increase; err300=%.6g err400=%.6g", err300, err400)
	}

	t.Logf("RMSE vs ideal: 300K=%.6g 400K=%.6g (delta=%.6g)", err300, err400, err400-err300)
}

func rmse(a, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}
	var s float64
	for i := 0; i < n; i++ {
		d := a[i] - b[i]
		s += d * d
	}
	return math.Sqrt(s / float64(n))
}
