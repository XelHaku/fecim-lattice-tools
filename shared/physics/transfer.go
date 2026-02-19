package physics

import "math"

// PolarizationToConductance maps polarization to conductance using the linear model by default
// for backward compatibility.
//
// Limitation note: linear mapping tends to over-estimate level separability versus real
// FeFET subthreshold behavior, where IDS varies exponentially with VT shift.
func PolarizationToConductance(P, Ps, Gmin, Gmax float64) float64 {
	return PolarizationToConductanceWithParams(P, Ps, Gmin, Gmax, ConductanceLinear, 0, 0, 0)
}

// PolarizationToConductanceModel maps polarization to conductance using the requested model.
func PolarizationToConductanceModel(P, Ps, Gmin, Gmax float64, model ConductanceModel) float64 {
	return PolarizationToConductanceWithParams(P, Ps, Gmin, Gmax, model, 0, 0, 0)
}

// PolarizationToConductanceWithParams maps polarization to conductance using model-specific
// physical parameters.
func PolarizationToConductanceWithParams(P, Ps, Gmin, Gmax float64, model ConductanceModel, kvT, vgsRead, vt0 float64) float64 {
	if Ps == 0 {
		return (Gmin + Gmax) / 2
	}
	normalizedP := P / Ps
	if normalizedP < -1 {
		normalizedP = -1
	}
	if normalizedP > 1 {
		normalizedP = 1
	}
	if Gmax <= Gmin {
		return Gmin
	}

	switch model {
	case ConductanceSubthreshold:
		if kvT <= 0 {
			// Fallback to log-space interpolation preserving Gmax/Gmin window.
			n := (normalizedP + 1.0) / 2.0
			logGmin := math.Log(Gmin)
			logGmax := math.Log(Gmax)
			return math.Exp(logGmin + (logGmax-logGmin)*n)
		}
		nFactor := 1.3
		vThermal := 0.026
		deltaVt := kvT * normalizedP
		iRaw := math.Exp(deltaVt / (nFactor * vThermal))
		iMin := math.Exp((-kvT) / (nFactor * vThermal))
		iMax := math.Exp((kvT) / (nFactor * vThermal))
		n := (iRaw - iMin) / (iMax - iMin)
		if n < 0 {
			n = 0
		}
		if n > 1 {
			n = 1
		}
		return Gmin + (Gmax-Gmin)*n
	case ConductanceSaturation:
		if kvT <= 0 {
			kvT = 0.2
		}
		if vgsRead <= 0 {
			vgsRead = 1.0
		}
		if vt0 <= 0 {
			vt0 = 0.5
		}
		deltaVt := kvT * normalizedP
		vov := vgsRead - (vt0 - deltaVt)
		vovMin := vgsRead - (vt0 + kvT)
		vovMax := vgsRead - (vt0 - kvT)
		iRaw := math.Max(vov, 0)
		iMin := math.Max(vovMin, 0)
		iMax := math.Max(vovMax, 0)
		iRaw *= iRaw
		iMin *= iMin
		iMax *= iMax
		if iMax <= iMin {
			return Gmin + (Gmax-Gmin)*(normalizedP+1)/2
		}
		n := (iRaw - iMin) / (iMax - iMin)
		if n < 0 {
			n = 0
		}
		if n > 1 {
			n = 1
		}
		return Gmin + (Gmax-Gmin)*n
	default:
		return Gmin + (Gmax-Gmin)*(normalizedP+1)/2
	}
}

func ConductanceToPolarization(G, Gmin, Gmax, Ps float64) float64 {
	if Gmax == Gmin {
		return 0
	}

	normalizedG := (G - Gmin) / (Gmax - Gmin)
	normalizedP := 2*normalizedG - 1

	return normalizedP * Ps
}

// ConductanceToPolarizationModel maps conductance to polarization using the
// specified conductance model as the inverse function.
//
// For the exponential model (ConductanceExponential/ConductanceSubthreshold),
// uses the physically correct log-space inverse:
//
//	gNorm = ln(G/Gmin) / ln(Gmax/Gmin)
//	P/Ps  = 2*gNorm - 1
//
// This is the exact inverse of PolarizationToConductanceWithParams for the
// exponential case. Using the linear inverse for an exponential-mapped G
// introduces level errors of 5–10 discrete steps at moderate conductances.
//
// For linear and other models, falls back to the linear inverse (same as
// the original ConductanceToPolarization).
func ConductanceToPolarizationModel(G, Gmin, Gmax, Ps float64, model ConductanceModel) float64 {
	if Gmax == Gmin {
		return 0
	}
	if G < Gmin {
		G = Gmin
	}
	if G > Gmax {
		G = Gmax
	}

	var gNorm float64
	switch model {
	case ConductanceExponential:
		if Gmin <= 0 {
			// Gmin must be positive for log-space inverse; fall back to linear.
			gNorm = (G - Gmin) / (Gmax - Gmin)
		} else {
			logRatio := math.Log(Gmax / Gmin)
			if logRatio < 1e-12 {
				gNorm = 0.5
			} else {
				gNorm = math.Log(G/Gmin) / logRatio
			}
		}
	default:
		gNorm = (G - Gmin) / (Gmax - Gmin)
	}

	if gNorm < 0 {
		gNorm = 0
	}
	if gNorm > 1 {
		gNorm = 1
	}
	return (2*gNorm - 1) * Ps
}
