package crossbar

import (
	"math"
	"testing"
)

// M2-TMP-01: Verify temperature-dependent Arrhenius scaling through the MVM
// temperature integration.
//
// The current MVM temperature path applies Arrhenius scaling via
// TemperatureEffects.AdjustedDriftRate() when opts.EnableDrift is true AND a
// TemperatureProfile with ApplyDrift is enabled.
//
// We back out the implied rate ratio from a 1x1 MVM result and fit
// ln(rateRatio) vs 1/T to estimate Ea; require Ea > 0.
func TestM2_TMP_01_TemperatureArrheniusConductanceShift(t *testing.T) {
	cfg := &Config{Rows: 1, Cols: 1, NoiseLevel: 0, ADCBits: 16, DACBits: 16}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}
	if err := arr.ProgramWeight(0, 0, 1.0); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}

	prof := &TemperatureProfile{Enable: true, ApplyDrift: true}
	input := []float64{1.0}

	temps := []float64{300, 350, 400, 450}
	invT := make([]float64, 0, len(temps))
	lnRateRatio := make([]float64, 0, len(temps))

	base := FeFETDriftCoefficients.Assumed
	if base <= 0 {
		t.Fatalf("unexpected base drift coefficient: %v", base)
	}

	for _, T := range temps {
		res, err := arr.MVMWithNonIdealities(input, &MVMOptions{
			EnableIRDrop:       false,
			EnableSneakPaths:   false,
			EnableVariation:    false,
			EnableDrift:        true,
			Temperature:        T,
			TemperatureProfile: prof,
		})
		if err != nil {
			t.Fatalf("MVMWithNonIdealities @%.0fK: %v", T, err)
		}

		// For 1x1 with weight=1, input=1:
		//   G' = 0.5 + (1-0.5)*(1-effective) = 1 - 0.5*effective
		// => effective = 2*(1-G')
		gPrime := res.ActualOutput[0]
		effective := 2.0 * (1.0 - gPrime)
		rateRatio := effective / base // equals te.AdjustedDriftRate(1.0)

		if rateRatio <= 0 || math.IsNaN(rateRatio) || math.IsInf(rateRatio, 0) {
			t.Fatalf("invalid inferred rateRatio at %.0fK: g'=%v effective=%v rateRatio=%v", T, gPrime, effective, rateRatio)
		}

		invT = append(invT, 1.0/T)
		lnRateRatio = append(lnRateRatio, math.Log(rateRatio))
	}

	b, a := linearFit(invT, lnRateRatio)

	const kB_eV = 8.617e-5 // eV/K
	Ea := -b * kB_eV

	t.Logf("Arrhenius fit: ln(rateRatio)=a+b*(1/T); a=%.6g b=%.6g => Ea=%.4f eV", a, b, Ea)

	if !(Ea > 0) {
		t.Fatalf("expected Ea > 0, got %.6f eV (b=%.6g)", Ea, b)
	}
	if Ea < 0.2 || Ea > 1.5 {
		t.Fatalf("Ea out of expected range [0.2,1.5] eV: Ea=%.4f eV", Ea)
	}
}

func linearFit(x, y []float64) (b, a float64) {
	if len(x) != len(y) || len(x) < 2 {
		return 0, 0
	}
	var sx, sy, sxx, sxy float64
	n := float64(len(x))
	for i := range x {
		sx += x[i]
		sy += y[i]
		sxx += x[i] * x[i]
		sxy += x[i] * y[i]
	}
	den := n*sxx - sx*sx
	if den == 0 {
		return 0, 0
	}
	b = (n*sxy - sx*sy) / den
	a = (sy - b*sx) / n
	return b, a
}
