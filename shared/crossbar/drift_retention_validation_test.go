package crossbar

import (
	"math"
	"testing"
)

// M2-DFT-03: Use CalibrateDriftToRetention with synthetic Arrhenius retention data.
// Extrapolate to 10yr@85°C. Activation energy Ea should be in [0.5, 1.5] eV.
func TestM2DFT03_RetentionCalibration_ArrheniusEaInRange(t *testing.T) {
	// Synthetic ground-truth parameters (inside calibrator search bounds).
	trueCoeff := 0.01
	trueExp := 0.05
	trueEa := 0.9

	// Build synthetic dataset across T and time.
	data := []RetentionDatum{
		{TimeS: 1e3, TemperatureK: 300, Retention: retentionFromParams(1e3, 300, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e5, TemperatureK: 300, Retention: retentionFromParams(1e5, 300, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e3, TemperatureK: 358, Retention: retentionFromParams(1e3, 358, trueCoeff, trueExp, trueEa)}, // 85C
		{TimeS: 1e5, TemperatureK: 358, Retention: retentionFromParams(1e5, 358, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e7, TemperatureK: 358, Retention: retentionFromParams(1e7, 358, trueCoeff, trueExp, trueEa)},
	}

	fit := CalibrateDriftToRetention(data)
	if fit.ActivationE < 0.5 || fit.ActivationE > 1.5 {
		t.Fatalf("Ea out of expected range: Ea=%0.4feV (expected [0.5, 1.5])", fit.ActivationE)
	}

	// Extrapolate to 10 years at 85C.
	tenYearSeconds := 10 * 365.25 * 24 * 3600
	ret10y := retentionFromParams(float64(tenYearSeconds), 358, fit.Coeff, fit.Exponent, fit.ActivationE)
	if math.IsNaN(ret10y) || math.IsInf(ret10y, 0) {
		t.Fatalf("invalid 10yr retention prediction: %v", ret10y)
	}
	if ret10y <= 0 {
		t.Fatalf("expected positive 10yr retention at 85C, got %g", ret10y)
	}

	t.Logf("M2-DFT-03 fit: coeff=%g exp=%0.3f Ea=%0.3feV rmse=%g; 10yr@85C retention=%0.6f", fit.Coeff, fit.Exponent, fit.ActivationE, fit.RMSE, ret10y)
}
