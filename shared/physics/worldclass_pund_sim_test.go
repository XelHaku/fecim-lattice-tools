package physics

import (
	"math"
	"testing"
)

func TestRunPUNDSimulation_Basic(t *testing.T) {
	// Use a simple Preisach stack
	ec := 3e7 // ~30 MV/m
	ps := 0.25
	stack := NewPreisachStack(ec, &TanhEverett{Ec: ec, Ps: ps, Delta: ec * 0.3})

	amplitude := 5 * ec
	pulseWidth := 100e-9   // 100 ns
	sampleInterval := 5e-9 // 5 ns
	area := 1e-12          // 1 µm²

	result, traces, err := RunPUNDSimulation(stack, amplitude, pulseWidth, sampleInterval, area)
	if err != nil {
		t.Fatalf("RunPUNDSimulation failed: %v", err)
	}
	// Each trace should have the right number of samples
	expectedSamples := int(math.Ceil(pulseWidth/sampleInterval)) + 1
	for i, trace := range traces {
		if len(trace) != expectedSamples {
			t.Errorf("trace[%d]: expected %d samples, got %d", i, expectedSamples, len(trace))
		}
	}
	// Switching charge should be non-zero (we drove well above coercive field)
	if math.Abs(result.SwitchingPositive_C) < 1e-20 {
		t.Errorf("SwitchingPositive_C should be non-zero, got %g", result.SwitchingPositive_C)
	}
}

func TestRunPUNDSimulation_InvalidInputs(t *testing.T) {
	ec := 3e7
	ps := 0.25
	stack := NewPreisachStack(ec, &TanhEverett{Ec: ec, Ps: ps, Delta: ec * 0.3})
	_, _, err := RunPUNDSimulation(stack, -1, 100e-9, 5e-9, 1e-12)
	if err == nil {
		t.Error("expected error for negative amplitude")
	}
}
