//go:build legacy_fyne

package gui

import (
	"math"
	"math/rand"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

func TestReadMarginBER_AdjacentLevelsExceedThreeSigma(t *testing.T) {
	ds := newTestDeviceState(1, 1)
	mat := ds.GetMaterial()
	// Use linear conductance model for uniform level spacing in this read margin test
	mat.ConductanceModel = "linear"
	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{Rf: 10e3, Vref: 0.0, Vmin: 0.0, Vmax: 1.0},
		ADC: arraysim.ADCConfig{Bits: 5, Vmin: 0.0, Vmax: 1.0},
	}
	const (
		levels    = 30
		readV     = 0.2
		samples   = 50
		minMargin = 3.0
	)

	worst := math.Inf(1)
	worstPair := -1
	rng := rand.New(rand.NewSource(7))

	for i := 0; i < levels-1; i++ {
		g0 := mat.DiscreteLevel(i, levels)
		g1 := mat.DiscreteLevel(i+1, levels)

		m0, s0 := sampleReadVoutStats(sense, g0*readV, samples, rng)
		m1, s1 := sampleReadVoutStats(sense, g1*readV, samples, rng)

		deltaV := math.Abs(m1 - m0)
		sigma := math.Sqrt((s0*s0 + s1*s1) / 2)
		margin := math.Inf(1)
		if sigma > 0 {
			margin = deltaV / (2 * sigma)
		}
		if margin < worst {
			worst = margin
			worstPair = i
		}
		if margin <= minMargin {
			t.Fatalf("adjacent levels (%d,%d) insufficient read margin: ΔV=%.6e V sigma=%.6e V margin=%.3f", i, i+1, deltaV, sigma, margin)
		}
	}

	t.Logf("worst-case read margin %.3f sigma at level pair (%d,%d)", worst, worstPair, worstPair+1)
}

func sampleReadVoutStats(sense arraysim.SenseChain, currentA float64, n int, rng *rand.Rand) (mean, std float64) {
	vals := make([]float64, n)
	sum := 0.0
	for i := 0; i < n; i++ {
		vals[i] = sense.ConvertCurrentWithNoise(currentA, rng).Vout
		sum += vals[i]
	}
	mean = sum / float64(n)
	if n < 2 {
		return mean, 0
	}
	var ss float64
	for i := range vals {
		d := vals[i] - mean
		ss += d * d
	}
	std = math.Sqrt(ss / float64(n-1))
	return mean, std
}
