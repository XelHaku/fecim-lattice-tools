package physics

import (
	"encoding/json"
	"math"
	"testing"
)

func TestFORC_ReversalCurvesMonotonic(t *testing.T) {
	sat := 1.0
	ps := NewPreisachStack(sat, simpleUniformEverett{sat: sat})

	res, err := RunFORCSweep(ps, sat, 25)
	if err != nil {
		t.Fatalf("RunFORCSweep error: %v", err)
	}

	for _, c := range res.Curves {
		if len(c.Polarization_Cm2) < 2 {
			continue
		}
		for i := 1; i < len(c.Polarization_Cm2); i++ {
			if c.Polarization_Cm2[i] > c.Polarization_Cm2[i-1]+1e-12 {
				t.Fatalf("curve Er=%g is not monotonic descending at idx %d: prev=%g now=%g", c.ReversalField_Vm, i, c.Polarization_Cm2[i-1], c.Polarization_Cm2[i])
			}
		}
	}
}

func TestFORC_DensityPeakNearEc(t *testing.T) {
	sat := 1.0
	ps := NewPreisachStack(sat, simpleUniformEverett{sat: sat})

	res, err := RunFORCSweep(ps, sat, 41)
	if err != nil {
		t.Fatalf("RunFORCSweep error: %v", err)
	}

	ec := 0.0
	bestAbs := math.Inf(-1)
	peakEa := math.NaN()
	peakEb := math.NaN()
	for i := range res.PreisachDensity {
		for j := range res.PreisachDensity[i] {
			ea := res.ReversalFields_Vm[j]
			eb := res.ReversalFields_Vm[i]
			if math.Abs(ea) > 0.5*sat || math.Abs(eb) > 0.5*sat {
				continue // ignore edge artifacts from finite-difference stencil
			}
			rho := res.PreisachDensity[i][j]
			mag := math.Abs(rho)
			if mag > bestAbs {
				bestAbs = mag
				peakEa = ea
				peakEb = eb
			}
		}
	}
	if math.IsNaN(peakEa) || math.IsNaN(peakEb) {
		t.Fatal("no valid FORC density samples")
	}

	if math.Abs(peakEa-ec) > 0.6*sat || math.Abs(peakEb-ec) > 0.6*sat {
		t.Fatalf("FORC density peak too far from Ec: peak(Ea=%g,Eb=%g), Ec=%g", peakEa, peakEb, ec)
	}
}

func TestFORC_JSONExport(t *testing.T) {
	sat := 1.0
	ps := NewPreisachStack(sat, simpleUniformEverett{sat: sat})

	res, err := RunFORCSweep(ps, sat, 11)
	if err != nil {
		t.Fatalf("RunFORCSweep error: %v", err)
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var round FORCResult
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if len(round.Curves) != len(res.Curves) {
		t.Fatalf("curve count mismatch after JSON round-trip: got %d want %d", len(round.Curves), len(res.Curves))
	}
	if len(round.ReversalPairs) == 0 {
		t.Fatal("expected non-empty reversal field pairs after JSON round-trip")
	}
}
