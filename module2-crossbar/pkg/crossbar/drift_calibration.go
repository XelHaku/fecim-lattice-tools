package crossbar

import "math"

// RetentionDatum is a measured retention point from literature.
type RetentionDatum struct {
	TimeS        float64 // elapsed time
	TemperatureK float64 // temperature
	Retention    float64 // fraction [0,1]
}

// DriftCalibration contains fitted drift coefficients.
type DriftCalibration struct {
	Coeff       float64 // loss prefactor at tref,Tref
	Exponent    float64 // power-law exponent
	ActivationE float64 // eV
	RMSE        float64
}

func retentionFromParams(t, tempK, coeff, exp, ea float64) float64 {
	if t <= 0 {
		return 1
	}
	const tref = 3600.0
	const Tref = 358.0
	const kBeV = 8.617333262145e-5
	accel := math.Exp((ea / kBeV) * (1.0/Tref - 1.0/tempK))
	loss := coeff * math.Pow(t/tref, exp) * accel
	r := 1.0 - loss
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

// CalibrateDriftToRetention performs a compact grid-search fit.
func CalibrateDriftToRetention(data []RetentionDatum) DriftCalibration {
	best := DriftCalibration{RMSE: math.Inf(1)}
	for coeff := 1e-4; coeff <= 0.05; coeff *= 1.25 {
		for exp := 0.01; exp <= 0.35; exp += 0.01 {
			for ea := 0.2; ea <= 1.1; ea += 0.02 {
				sse := 0.0
				for _, d := range data {
					pred := retentionFromParams(d.TimeS, d.TemperatureK, coeff, exp, ea)
					e := pred - d.Retention
					sse += e * e
				}
				rmse := math.Sqrt(sse / float64(len(data)))
				if rmse < best.RMSE {
					best = DriftCalibration{Coeff: coeff, Exponent: exp, ActivationE: ea, RMSE: rmse}
				}
			}
		}
	}
	return best
}
