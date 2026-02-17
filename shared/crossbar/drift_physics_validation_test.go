package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

func TestM2DFT01_DriftDirection_RelaxesTowardMean(t *testing.T) {
	rand.Seed(1)

	sim := NewDriftSimulator(1, 2, 8)
	sim.DriftNoiseSigma = 0
	sim.DriftCoeff = 0.05
	sim.DriftExponent = 0.05

	sim.Conductances[0][0] = sim.GMax
	sim.InitialConds[0][0] = sim.GMax
	sim.Conductances[0][1] = sim.GMin
	sim.InitialConds[0][1] = sim.GMin

	high0 := sim.Conductances[0][0]
	low0 := sim.Conductances[0][1]

	sim.SimulateTimeStep(10)

	high1 := sim.Conductances[0][0]
	low1 := sim.Conductances[0][1]

	if !(high1 < high0) {
		t.Fatalf("high-G cell should decrease toward mean: before=%g after=%g", high0, high1)
	}
	if !(low1 > low0) {
		t.Fatalf("low-G cell should increase toward mean: before=%g after=%g", low0, low1)
	}
}

func linregSlope(xs, ys []float64) float64 {
	if len(xs) != len(ys) || len(xs) < 2 {
		return math.NaN()
	}
	n := float64(len(xs))
	sx, sy, sxx, sxy := 0.0, 0.0, 0.0, 0.0
	for i := range xs {
		x := xs[i]
		y := ys[i]
		sx += x
		sy += y
		sxx += x * x
		sxy += x * y
	}
	den := n*sxx - sx*sx
	if den == 0 {
		return math.NaN()
	}
	return (n*sxy - sx*sy) / den
}

func TestM2DFT02_DriftPowerLawExponentInRange(t *testing.T) {
	rand.Seed(2)

	sim := NewDriftSimulator(1, 2, 8)
	sim.DriftNoiseSigma = 0
	sim.DriftCoeff = 1e-3
	sim.DriftExponent = 0.05

	sim.Conductances[0][0] = sim.GMax
	sim.InitialConds[0][0] = sim.GMax
	sim.Conductances[0][1] = sim.GMin
	sim.InitialConds[0][1] = sim.GMin

	times := []float64{1, 10, 100, 1000}
	logt := make([]float64, 0, len(times))
	logd := make([]float64, 0, len(times))

	sim.Reset()
	prevT := 0.0
	for _, tt := range times {
		sim.SimulateTimeStep(tt - prevT)
		prevT = tt

		dg := math.Abs(sim.Conductances[0][0] - sim.InitialConds[0][0])
		if dg <= 0 {
			t.Fatalf("expected positive drift magnitude at t=%gs", tt)
		}
		logt = append(logt, math.Log(tt))
		logd = append(logd, math.Log(dg))
	}

	nu := linregSlope(logt, logd)
	if nu < 0.01 || nu > 0.1 {
		t.Fatalf("drift exponent nu out of range: nu=%0.6f (expected [0.01, 0.1])", nu)
	}
	if math.IsNaN(nu) || math.IsInf(nu, 0) {
		t.Fatalf("invalid fitted slope nu=%v", nu)
	}

	t.Logf("M2-DFT-02 fitted drift exponent nu=%0.6f", nu)
}
