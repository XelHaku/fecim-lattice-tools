package neural

import (
	"math"
	"testing"
)

func TestEnergyPerMACJ_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		levels     int
		expectedFJ float64
	}{
		{"levels<=0 clamps to 1 bit", 0, 10.0},
		{"1 level clamps to 1 bit", 1, 10.0},
		{"2 levels = 1 bit", 2, 10.0},
		{"8 levels = 3 bits", 8, 30.0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotFJ := EnergyPerMACJ(tc.levels) / 1e-15
			if math.Abs(gotFJ-tc.expectedFJ) > 1e-9 {
				t.Fatalf("EnergyPerMACJ(%d) = %.12g fJ, want %.12g fJ", tc.levels, gotFJ, tc.expectedFJ)
			}
		})
	}
}

func TestEstimateInferenceEnergyJ_MNIST30Levels(t *testing.T) {
	t.Parallel()

	cfg := DefaultNetworkConfig()
	cfg.NumLevels = 30
	est := EstimateInferenceEnergyJ(cfg, 784, 128, 10)

	if est.TotalMACs != 101632 {
		t.Fatalf("TotalMACs = %d, want 101632", est.TotalMACs)
	}
	if est.TotalJ <= 0 {
		t.Fatalf("TotalJ = %g, want > 0", est.TotalJ)
	}

	// Expected: compute = MACs * (10 fJ/bit * log2(30)), overhead = DAC+ADC.
	expectedComputeJ := float64(101632) * EnergyPerBitPerMACJ * math.Log2(30)
	expectedOverheadJ := float64(784)*EnergyPerDACJ + float64(128+10)*EnergyPerADCJ
	expectedTotalJ := expectedComputeJ + expectedOverheadJ

	// Keep a small relative tolerance (floating math).
	if math.Abs(est.TotalJ-expectedTotalJ) > expectedTotalJ*1e-12 {
		t.Fatalf("TotalJ = %.18g, want %.18g", est.TotalJ, expectedTotalJ)
	}
}
