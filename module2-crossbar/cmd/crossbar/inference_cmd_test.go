package crossbarcli

import (
	"bytes"
	"math"
	"math/rand"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

func TestRunInferenceReportsFlagErrorToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runInference([]string{"-definitely-not-a-flag"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("runInference error = nil, want invalid flag error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout length = %d, want 0; stdout=%q", stdout.Len(), stdout.String())
	}
	text := stderr.String()
	if !strings.Contains(text, "flag provided but not defined: -definitely-not-a-flag") {
		t.Fatalf("stderr = %q, want invalid flag context", text)
	}
	if !strings.Contains(text, "Error:") {
		t.Fatalf("stderr = %q, want error prefix", text)
	}
	if !strings.Contains(text, "FeCIM Crossbar Inference") {
		t.Fatalf("stderr = %q, want usage", text)
	}
}

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
