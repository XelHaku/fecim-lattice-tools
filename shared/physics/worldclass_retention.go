package physics

import (
	"fmt"
	"math"
)

// RetentionPoint stores polarization at a hold time after removing the field.
type RetentionPoint struct {
	TimeS           float64 // Hold time after field removal
	Polarization_Cm float64 // C/m^2
}

// GenerateLogTimeSweep returns logarithmically spaced hold times from tMin to tMax.
func GenerateLogTimeSweep(tMinS, tMaxS float64, points int) ([]float64, error) {
	if tMinS <= 0 || tMaxS <= 0 || tMaxS <= tMinS {
		return nil, fmt.Errorf("invalid time range: tMin=%g tMax=%g", tMinS, tMaxS)
	}
	if points < 2 {
		return nil, fmt.Errorf("need at least 2 points, got %d", points)
	}
	out := make([]float64, points)
	logMin := math.Log10(tMinS)
	logMax := math.Log10(tMaxS)
	for i := 0; i < points; i++ {
		f := float64(i) / float64(points-1)
		out[i] = math.Pow(10, logMin+f*(logMax-logMin))
	}
	return out, nil
}

// SimulateRetentionPowerLaw models P(t) = P0 * (t/t0)^(-beta) for t >= t0.
// For t < t0 the value is clamped to P0.
// beta is the retention decay exponent (typical 0.01-0.05 for HZO).
func SimulateRetentionPowerLaw(P0_Cm, t0S, beta float64, timesS []float64) ([]RetentionPoint, error) {
	if t0S <= 0 {
		return nil, fmt.Errorf("t0 must be positive, got %g", t0S)
	}
	if beta < 0 {
		return nil, fmt.Errorf("beta must be non-negative, got %g", beta)
	}
	out := make([]RetentionPoint, len(timesS))
	for i, t := range timesS {
		if t < 0 {
			return nil, fmt.Errorf("negative time at index %d: %g", i, t)
		}
		p := P0_Cm
		if t >= t0S && beta > 0 {
			p = P0_Cm * math.Pow(t/t0S, -beta)
		}
		out[i] = RetentionPoint{TimeS: t, Polarization_Cm: p}
	}
	return out, nil
}

// SimulateRetentionExponential models P(t)=Pinf+(P0-Pinf)*exp(-t/tau).
func SimulateRetentionExponential(initialP_Cm, asymptoticP_Cm, tauS float64, timesS []float64) ([]RetentionPoint, error) {
	if tauS <= 0 {
		return nil, fmt.Errorf("tau must be positive, got %g", tauS)
	}
	out := make([]RetentionPoint, len(timesS))
	for i, t := range timesS {
		if t < 0 {
			return nil, fmt.Errorf("negative time at index %d: %g", i, t)
		}
		p := asymptoticP_Cm + (initialP_Cm-asymptoticP_Cm)*math.Exp(-t/tauS)
		out[i] = RetentionPoint{TimeS: t, Polarization_Cm: p}
	}
	return out, nil
}
