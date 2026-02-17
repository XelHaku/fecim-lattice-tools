package crossbar

import (
	"math/rand"
	"testing"
)

// M2-DFT-05: Multi-level drift: program 8 levels; after 1e4s drift, ordering preserved (no crossings).
func TestM2DFT05_DriftLevelIntegrity_NoCrossingsAfterLongTime(t *testing.T) {
	rand.Seed(3)

	levels := 8
	sim := NewDriftSimulator(1, levels, levels)
	sim.DriftNoiseSigma = 0 // deterministic ordering
	sim.DriftCoeff = 0.02
	sim.DriftExponent = 0.05

	for c := 0; c < levels; c++ {
		sim.SetConductanceLevel(0, c, c)
	}

	sim.SimulateTimeStep(1e4)

	prev := sim.Conductances[0][0]
	for c := 1; c < levels; c++ {
		g := sim.Conductances[0][c]
		if !(g > prev) {
			t.Fatalf("level ordering crossed at col=%d: g[%d]=%g <= g[%d]=%g", c, c, g, c-1, prev)
		}
		prev = g
	}

	t.Logf("M2-DFT-05 ordering preserved after 1e4s drift: g0=%g ... g7=%g", sim.Conductances[0][0], sim.Conductances[0][levels-1])
}
