package crossbar

import (
	"math"
	"testing"
)

// M2-DFT-04: Round-trip: feed known retention data → CalibrateDriftToRetention → recover coefficients.
func TestM2DFT04_CalibrateDriftToRetention_RoundTripRecoversParams(t *testing.T) {
	trueCoeff := 0.008
	trueExp := 0.06
	trueEa := 0.86

	data := []RetentionDatum{
		{TimeS: 1e2, TemperatureK: 300, Retention: retentionFromParams(1e2, 300, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e4, TemperatureK: 300, Retention: retentionFromParams(1e4, 300, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e6, TemperatureK: 300, Retention: retentionFromParams(1e6, 300, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e2, TemperatureK: 358, Retention: retentionFromParams(1e2, 358, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e4, TemperatureK: 358, Retention: retentionFromParams(1e4, 358, trueCoeff, trueExp, trueEa)},
		{TimeS: 1e6, TemperatureK: 358, Retention: retentionFromParams(1e6, 358, trueCoeff, trueExp, trueEa)},
	}

	fit := CalibrateDriftToRetention(data)

	// Grid-search resolution is coarse, so use loose tolerances.
	// Coeff grid is multiplicative, exponent and Ea are linear steps.
	relErr := func(a, b float64) float64 {
		if b == 0 {
			return math.Inf(1)
		}
		return math.Abs(a-b) / math.Abs(b)
	}

	if relErr(fit.Coeff, trueCoeff) > 0.35 {
		t.Fatalf("coeff not recovered: got=%g want=%g relErr=%0.3f", fit.Coeff, trueCoeff, relErr(fit.Coeff, trueCoeff))
	}
	if math.Abs(fit.Exponent-trueExp) > 0.02 {
		t.Fatalf("exponent not recovered: got=%0.4f want=%0.4f", fit.Exponent, trueExp)
	}
	if math.Abs(fit.ActivationE-trueEa) > 0.08 {
		t.Fatalf("Ea not recovered: got=%0.4f want=%0.4f", fit.ActivationE, trueEa)
	}
	// With coarse grid-search, best-fit RMSE will not be exactly zero even for noiseless data.
	if fit.RMSE > 1e-3 {
		t.Fatalf("RMSE too large on synthetic data: got %g", fit.RMSE)
	}

	t.Logf("M2-DFT-04 recovered: coeff=%g (true=%g) exp=%0.3f (true=%0.3f) Ea=%0.3feV (true=%0.3f) rmse=%g", fit.Coeff, trueCoeff, fit.Exponent, trueExp, fit.ActivationE, trueEa, fit.RMSE)
}
