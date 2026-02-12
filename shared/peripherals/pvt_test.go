package peripherals

import "testing"

func TestEffectiveINLDNL_TemperatureAndCornerScaling(t *testing.T) {
	inlFastCold, dnlFastCold := EffectiveINLDNL(0.5, 0.25, 250, CornerFast)
	inlTyp, dnlTyp := EffectiveINLDNL(0.5, 0.25, 300, CornerTypical)
	inlSlowHot, dnlSlowHot := EffectiveINLDNL(0.5, 0.25, 400, CornerSlow)

	if !(inlFastCold < inlTyp && inlTyp < inlSlowHot) {
		t.Fatalf("INL scaling order failed: fast/cold=%.4f typ=%.4f slow/hot=%.4f", inlFastCold, inlTyp, inlSlowHot)
	}
	if !(dnlFastCold < dnlTyp && dnlTyp < dnlSlowHot) {
		t.Fatalf("DNL scaling order failed: fast/cold=%.4f typ=%.4f slow/hot=%.4f", dnlFastCold, dnlTyp, dnlSlowHot)
	}
}

func TestAnalyzeProcessCorners_Order(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	corners := AnalyzeProcessCorners(dac, adc, 300)

	if corners == nil || corners.Fast == nil || corners.Typical == nil || corners.Slow == nil {
		t.Fatal("expected all process-corner analyses to be populated")
	}

	fast := absFloat(corners.Fast.DAC.MaxINL)
	typical := absFloat(corners.Typical.DAC.MaxINL)
	slow := absFloat(corners.Slow.DAC.MaxINL)
	if !(fast <= typical && typical <= slow) {
		t.Fatalf("expected |MaxINL| fast<=typical<=slow, got %.4f <= %.4f <= %.4f", fast, typical, slow)
	}
}

func TestConvertWithCondition_NominalCompat(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()

	for _, code := range []int{0, 7, 15, 31} {
		vNominal := dac.ConvertWithNonlinearity(code)
		vCondition := dac.ConvertWithCondition(code, 300, CornerTypical)
		if absFloat(vNominal-vCondition) > 1e-12 {
			t.Fatalf("DAC nominal mismatch at code %d: %.12f vs %.12f", code, vNominal, vCondition)
		}
	}

	for _, vin := range []float64{0.1, 0.3, 0.7, 0.95} {
		lNominal := adc.ConvertWithNonlinearity(vin)
		lCondition := adc.ConvertWithCondition(vin, 300, CornerTypical)
		if lNominal != lCondition {
			t.Fatalf("ADC nominal mismatch at vin=%.3f: %d vs %d", vin, lNominal, lCondition)
		}
	}
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
