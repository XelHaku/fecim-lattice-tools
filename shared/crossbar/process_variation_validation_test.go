package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

// M2-PV-01: Validate per-cell variation factor statistics for configured NoiseLevel.
func TestM2_PV_01_ProcessVariationStatisticsMatchNoiseLevel(t *testing.T) {
	rand.Seed(12345)

	sigma := 0.05
	cfg := &Config{Rows: 64, Cols: 64, NoiseLevel: sigma, ADCBits: 12, DACBits: 12}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	N := cfg.Rows * cfg.Cols
	deltas := make([]float64, 0, N)
	for i := 0; i < cfg.Rows; i++ {
		for j := 0; j < cfg.Cols; j++ {
			deltas = append(deltas, arr.GetProcessVariationFactor(i, j)-1.0)
		}
	}

	m, s := meanStd(deltas)
	seMean := s / math.Sqrt(float64(N))
	bound := 3.0 * seMean

	t.Logf("variation delta stats: N=%d mean=%.6g std=%.6g (target std=%.6g); 3*SE=%.6g", N, m, s, sigma, bound)

	if math.Abs(m) > bound {
		t.Fatalf("mean not near 0: mean=%.6g exceeds 3*SE=%.6g", m, bound)
	}
	if rel := math.Abs(s-sigma) / sigma; rel > 0.20 {
		t.Fatalf("std not within 20%% of configured sigma: got=%.6g want=%.6g relErr=%.3f", s, sigma, rel)
	}
}

func meanStd(x []float64) (mean, std float64) {
	if len(x) == 0 {
		return 0, 0
	}
	for _, v := range x {
		mean += v
	}
	mean /= float64(len(x))
	if len(x) < 2 {
		return mean, 0
	}
	var ss float64
	for _, v := range x {
		d := v - mean
		ss += d * d
	}
	std = math.Sqrt(ss / float64(len(x)-1))
	return mean, std
}
