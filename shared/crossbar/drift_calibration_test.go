package crossbar

import "testing"

func TestCalibrateDriftToPublishedRetentionData(t *testing.T) {
	// Representative published-style targets for HZO FeFET retention:
	// ~99% at 85C after 10 years and ~90% at 150C after 10 years.
	tenYears := 10.0 * 365.25 * 24 * 3600
	data := []RetentionDatum{
		{TimeS: 3600, TemperatureK: 358, Retention: 0.999},
		{TimeS: 1e6, TemperatureK: 358, Retention: 0.997},
		{TimeS: tenYears, TemperatureK: 358, Retention: 0.99},
		{TimeS: tenYears, TemperatureK: 423, Retention: 0.90},
	}
	fit := CalibrateDriftToRetention(data)
	t.Logf("drift fit coeff=%.4g exp=%.3f Ea=%.3f eV RMSE=%.4f", fit.Coeff, fit.Exponent, fit.ActivationE, fit.RMSE)
	if fit.RMSE > 0.02 {
		t.Fatalf("retention fit RMSE=%.4f > 0.02", fit.RMSE)
	}
}
