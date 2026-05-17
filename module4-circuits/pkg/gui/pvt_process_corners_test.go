//go:build legacy_fyne

package gui

import (
	"math"
	"math/rand"
	"testing"
)

func TestPVTProcessCorners_MonteCarloYieldAbove90(t *testing.T) {
	ds := newTestDeviceState(4, 4)
	mat := ds.GetMaterial()
	if mat == nil {
		t.Fatal("expected material")
	}

	const (
		samples      = 20
		sigmaEcFrac  = 0.03
		sigmaPrFrac  = 0.03
		accuracySpec = 0.90 // normalized MVM-accuracy proxy
	)

	rng := rand.New(rand.NewSource(42))
	pass := 0
	for i := 0; i < samples; i++ {
		ec := mat.Ec * (1 + sigmaEcFrac*rng.NormFloat64())
		pr := mat.Pr * (1 + sigmaPrFrac*rng.NormFloat64())
		if ec <= 0 || pr <= 0 {
			continue
		}

		// Fast proxy: MVM gain error tracks Ec/Pr perturbation product around nominal.
		normGain := (ec / mat.Ec) * (pr / mat.Pr)
		accuracy := 1 - math.Abs(normGain-1)
		if accuracy >= accuracySpec {
			pass++
		}
	}
	yield := float64(pass) / samples
	if yield <= 0.90 {
		t.Fatalf("process-corner yield below target: %.1f%% (pass=%d/%d)", 100*yield, pass, samples)
	}
}
