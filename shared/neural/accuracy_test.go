package neural

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestMNISTAccuracyBounds_FPvsCIM(t *testing.T) {
	net := NewDualModeNetwork(784, 30, 10)
	net.Config.NumLevels = 30
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NoiseLevel = 0.0 // deterministic for a stable regression bound

	initializeDeterministicMNISTStyleWeights(net)
	net.RequantizeWeights()

	samples := generateSyntheticDigitPatterns(100)
	if len(samples) != 100 {
		t.Fatalf("dataset size=%d, want 100", len(samples))
	}

	fpCorrect := 0
	cimCorrect := 0
	for _, s := range samples {
		res := net.Infer(s.input)
		if res == nil {
			t.Fatal("Infer returned nil")
		}
		if res.FPPrediction == s.label {
			fpCorrect++
		}
		if res.CIMPrediction == s.label {
			cimCorrect++
		}
	}

	fpAcc := 100.0 * float64(fpCorrect) / float64(len(samples))
	cimAcc := 100.0 * float64(cimCorrect) / float64(len(samples))
	penalty := math.Abs(fpAcc - cimAcc)

	t.Logf("MNIST synthetic accuracy: FP=%.2f%% (%d/%d), CIM=%.2f%% (%d/%d), |Δ|=%.2f%%",
		fpAcc, fpCorrect, len(samples),
		cimAcc, cimCorrect, len(samples),
		penalty,
	)

	if fpAcc <= 80.0 {
		t.Fatalf("FP accuracy sanity failed: FP=%.2f%%, want >80%%", fpAcc)
	}
	if penalty > 15.0 {
		t.Fatalf("CIM penalty too high: FP=%.2f%%, CIM=%.2f%%, |Δ|=%.2f%%, want <=15%%", fpAcc, cimAcc, penalty)
	}
}

type labeledPattern struct {
	label int
	input []float64
}

func initializeDeterministicMNISTStyleWeights(net *DualModeNetwork) {
	// Build ten deterministic "stroke regions" across the 784 input pixels.
	regions := make([][]int, 10)
	for i := 0; i < net.InputSize; i++ {
		regions[i%10] = append(regions[i%10], i)
	}

	// 30 hidden neurons = 3 detectors per digit class.
	for d := 0; d < 10; d++ {
		for k := 0; k < 3; k++ {
			h := d*3 + k
			gain := 1.0 + 0.05*float64(k-1) // 0.95, 1.00, 1.05

			pos := gain / float64(len(regions[d]))
			neg := -0.25 / float64(net.InputSize-len(regions[d]))
			for i := 0; i < net.InputSize; i++ {
				net.FPWeights1[h][i] = neg
			}
			for _, idx := range regions[d] {
				net.FPWeights1[h][idx] = pos
			}
			net.FPBias1[h] = -0.18
		}
	}

	// Output layer: each class aggregates its 3 class detectors and mildly suppresses others.
	for d := 0; d < net.OutputSize; d++ {
		for h := 0; h < net.HiddenSize; h++ {
			net.FPWeights2[d][h] = -0.08
		}
		for k := 0; k < 3; k++ {
			net.FPWeights2[d][d*3+k] = 0.90
		}
		net.FPBias2[d] = 0.0
	}
}

func generateSyntheticDigitPatterns(total int) []labeledPattern {
	rng := rand.New(rand.NewSource(20260212))
	out := make([]labeledPattern, 0, total)

	for i := 0; i < total; i++ {
		label := i % 10
		x := make([]float64, 784)
		for p := 0; p < 784; p++ {
			if p%10 == label {
				x[p] = 0.82 + 0.16*rng.Float64() // strong class-correlated stroke pixels
			} else {
				x[p] = 0.02 + 0.20*rng.Float64() // background / off-stroke pixels
			}

			// occasional bit flips to avoid a perfectly trivial dataset
			if rng.Float64() < 0.01 {
				x[p] = 1.0 - x[p]
			}
		}

		// hard clip to [0,1] for valid DAC domain
		for p := range x {
			x[p] = math.Max(0.0, math.Min(1.0, x[p]))
		}

		out = append(out, labeledPattern{label: label, input: x})
	}

	if len(out) != total {
		panic(fmt.Sprintf("generated %d patterns, want %d", len(out), total))
	}
	return out
}
