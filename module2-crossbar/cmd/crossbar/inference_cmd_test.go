package crossbarcli

import (
	"math"
	"math/rand"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

func TestInferenceDeterministicMVM_SmallArray(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0, // keep deterministic (no device noise)
		ADCBits:    6,
		DACBits:    8,
	}
	array, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	rng := rand.New(rand.NewSource(42))
	programRandomWeights(array, rng)

	input := make([]float64, cfg.Cols)
	for i := range input {
		input[i] = rng.Float64()
	}

	got, err := array.MVM(input)
	if err != nil {
		t.Fatalf("MVM: %v", err)
	}

	// Golden values for seed=42, 4x4 array, noise=0.
	want := []float64{0.206349206349, 0.666666666667, 0.428571428571, 0.301587301587}
	if len(got) != len(want) {
		t.Fatalf("len(output)=%d want %d", len(got), len(want))
	}

	const tol = 1e-9
	for i := range want {
		if math.Abs(got[i]-want[i]) > tol {
			t.Fatalf("output[%d]=%.12f want %.12f", i, got[i], want[i])
		}
	}
}
