package compact


// FeFET is a compact model for a ferroelectric field-effect transistor.
//
// It embeds a *FeCap to track the ferroelectric polarisation history, and
// adds simple MOSFET square-law drain-current equations with a polarisation-
// dependent threshold voltage. This follows the standard FeFET compact model
// convention:
//
//	Vt(P) = Vt0 - γ × P
//
// where P is the FE polarisation charge density (C/m²) stored in the embedded
// FeCap. A positive P (upward polarisation) lowers Vt, increasing on-current.
//
// Reference: Si et al., "FeFET compact model with history-dependent threshold",
// IEEE TED 2019 (conceptual basis).
type FeFET struct {
	*FeCap

	// Gamma is the body factor coupling FE polarisation to threshold voltage (m²/C).
	// Typical value: ~0.1 V·cm²/µC = 0.1e-2 V·m²/C.
	Gamma float64

	// Vt0 is the nominal threshold voltage at zero polarisation (V).
	// Typical nFET: ~0.3–0.5 V.
	Vt0 float64

	// K is the MOSFET transconductance parameter (A/V²):
	//   K = (µ_n × Cox × W) / (2 × L)
	// Typical value for a small FeFET: ~50 µA/V² = 50e-6 A/V².
	K float64
}

// NewFeFET creates a FeFET model wrapping the given FeCap, with the supplied
// body factor gamma (m²/C), zero-polarisation threshold voltage vt0 (V), and
// transconductance parameter k (A/V²).
func NewFeFET(cap *FeCap, gamma, vt0, k float64) *FeFET {
	return &FeFET{FeCap: cap, Gamma: gamma, Vt0: vt0, K: k}
}

// ThresholdV returns the effective threshold voltage (V) for the given
// ferroelectric polarisation charge density polarization (C/m²):
//
//	Vt = Vt0 - Gamma × polarization
func (f *FeFET) ThresholdV(polarization float64) float64 {
	return f.Vt0 - f.Gamma*polarization
}

// Polarization returns the current polarisation charge density (C/m²) from
// the embedded FeCap at the supplied gate voltage vg (V).
//
// This is a convenience wrapper that evaluates TotalCharge at vg on the
// current direction branch of the FeCap.
func (f *FeFET) Polarization(vg float64) float64 {
	return f.FeCap.TotalCharge(vg)
}

// DrainCurrentA returns the drain current (A) using the simple square-law
// MOSFET model with the polarisation-dependent threshold voltage.
//
// The threshold voltage is evaluated at the current FeCap polarisation state
// (TotalCharge at vgs on the current branch):
//
//	Vt = ThresholdV(Polarization(vgs))
//
// Regions:
//   - Off  (vgs ≤ Vt):          Id = 0
//   - Linear  (vds < vgs - Vt): Id = K × (2×(vgs-Vt)×vds - vds²)
//   - Saturation (vds ≥ vgs-Vt): Id = K × (vgs - Vt)²
//
// Sub-threshold and short-channel effects are not modelled.
func (f *FeFET) DrainCurrentA(vgs, vds float64) float64 {
	p := f.Polarization(vgs)
	vt := f.ThresholdV(p)
	vov := vgs - vt
	if vov <= 0 {
		return 0
	}
	if vds < vov {
		// Linear region
		return f.K * (2*vov*vds - vds*vds)
	}
	// Saturation region
	return f.K * vov * vov
}

