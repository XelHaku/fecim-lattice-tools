package crossbar

// LIT-P2-07: Non-linear I-V curves for FeFET subthreshold region.
//
// In FeFET crossbar arrays the cells are biased in the subthreshold regime
// during read to maximise the on/off current ratio sensitivity to polarization
// state. In this regime the FET drain current saturates at small V_DS rather
// than following Ohm's Law (I = G × V):
//
//	I_DS = G × V_sat × (1 - exp(-V_DS / V_sat))   (subthreshold saturation)
//
// where V_sat = n × k_B × T / q (one subthreshold slope factor × thermal voltage).
// At room temperature V_sat ≈ 1.2 × 26 mV = 31 mV.
//
// Consequences for CIM:
//   - Linear model (I = G × V) overestimates current for V_DS > V_sat (~60 mV).
//   - Effective weight contribution becomes input-voltage-dependent (non-ideal MAC).
//   - High-voltage inputs (V >> V_sat) contribute the same saturation current
//     regardless of exact V, reducing dynamic range.
//
// Reference: Nature Comms 2018, IEEE IEDM 2020 FeFET array papers.

import (
	"fmt"
	"math"
	"sync/atomic"
)

const (
	kBoltzmann = 1.380649e-23  // Boltzmann constant (J/K)
	qElectron  = 1.602176634e-19 // Elementary charge (C)
)

// FeFETIVParams holds parameters for the FeFET subthreshold I-V model.
type FeFETIVParams struct {
	// SubthSlope is the subthreshold slope factor n (dimensionless, 1.0–2.0).
	// Ideal MOSFET: n = 1.0. Typical 28 nm CMOS FeFET: n ≈ 1.2.
	SubthSlope float64

	// TempK is the operating temperature (K). Default 300 K (27 °C).
	TempK float64
}

// DefaultFeFETIVParams returns FeFET subthreshold I-V parameters for a 28 nm
// CMOS-integrated FeFET at room temperature.
func DefaultFeFETIVParams() *FeFETIVParams {
	return &FeFETIVParams{
		SubthSlope: 1.2, // typical 28 nm FeFET
		TempK:      300, // room temperature
	}
}

// ThermalVoltage returns V_T = k_B × T / q (Volts).
func (p *FeFETIVParams) ThermalVoltage() float64 {
	return kBoltzmann * p.TempK / qElectron
}

// VSat returns the saturation voltage V_sat = n × V_T (Volts).
// Current saturates when V_DS ≳ V_sat (~31 mV at 300 K, n=1.2).
func (p *FeFETIVParams) VSat() float64 {
	return p.SubthSlope * p.ThermalVoltage()
}

// Current returns the subthreshold drain current for a single FeFET cell.
//
//	I = G × V_sat × (1 - exp(-|V_DS| / V_sat)) × sign(V_DS)
//
// Asymptotic behaviour:
//   - V_DS << V_sat: I ≈ G × V_DS  (Ohmic, linear regime)
//   - V_DS >> V_sat: I → G × V_sat (saturated)
func (p *FeFETIVParams) Current(gPhys, vDS float64) float64 {
	if math.Abs(vDS) < 1e-15 {
		return 0
	}
	vSat := p.VSat()
	sign := 1.0
	v := vDS
	if v < 0 {
		sign = -1.0
		v = -v
	}
	return sign * gPhys * vSat * (1 - math.Exp(-v/vSat))
}

// LinearityError returns the relative error of the Ohmic approximation (G×V)
// vs the non-linear model at the given V_DS:
//
//	err = (I_ohmic - I_nonlinear) / I_nonlinear
//
// Positive values indicate the linear model overestimates current.
func (p *FeFETIVParams) LinearityError(vDS float64) float64 {
	v := math.Abs(vDS)
	if v < 1e-15 {
		return 0
	}
	vSat := p.VSat()
	iNonlinear := vSat * (1 - math.Exp(-v/vSat))
	iOhmic := v
	return (iOhmic - iNonlinear) / iNonlinear
}

// MVMNonLinear performs a matrix-vector multiply using the FeFET subthreshold
// I-V model instead of the standard Ohmic (G × V) model.
//
// The output is normalised to [0,1] using the maximum possible non-linear
// current per input column (I_max = V_sat at unit conductance and unit input).
//
//	output[row] = (Σ_col I_nonlinear(G[row][col], V_dac[col])) / maxCurrent
//
// Returns an error if the array is in FeCAP mode (use MVMCharge instead) or
// if the input length does not match the number of columns.
func (a *Array) MVMNonLinear(input []float64, iv *FeFETIVParams) ([]float64, error) {
	if a.config.CellType == CellTypeFeCAP {
		return nil, fmt.Errorf("MVMNonLinear requires CellTypeFeFET; use MVMCharge for FeCAP")
	}
	if len(input) != a.config.Cols {
		return nil, fmt.Errorf("input length %d != cols %d", len(input), a.config.Cols)
	}
	if iv == nil {
		iv = DefaultFeFETIVParams()
	}

	// Normalisation: max current when all G=1, all V=1.
	// I_max_per_col = iv.Current(1.0, 1.0) = V_sat × (1-exp(-1/V_sat)) ≈ V_sat
	iMaxPerCol := iv.Current(1.0, 1.0)
	if iMaxPerCol <= 0 {
		return nil, fmt.Errorf("FeFETIVParams.Current(1,1) ≤ 0 — check SubthSlope/TempK")
	}
	maxCurrent := float64(a.config.Cols) * iMaxPerCol

	output := make([]float64, a.config.Rows)
	for r := 0; r < a.config.Rows; r++ {
		var sum float64
		for c, v := range input {
			vDAC := a.quantizeDAC(v)
			gPhys := a.cells[r][c].Conductance * a.GetProcessVariationFactor(r, c)
			sum += iv.Current(gPhys, vDAC)
		}
		output[r] = a.quantizeADC(sum / maxCurrent)
		atomic.AddInt64(&a.totalReads, 1)
	}

	getLog().Calculation("MVMNonLinear", map[string]interface{}{
		"rows": a.config.Rows, "cols": a.config.Cols,
		"v_sat_mV": iv.VSat() * 1000,
	}, output)
	return output, nil
}
