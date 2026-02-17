// Package compact provides device-level compact models for ferroelectric capacitors (FeCaps)
// and ferroelectric field-effect transistors (FeFETs).
//
// Inspired by PFECAP (Verilog-A, Preisach FeCap + Newton solver) and DNN+NeuroSim,
// this package enables circuit-simulator-compatible device models in Go.
// The FeCap model can generate charge–voltage curves with realistic hysteresis,
// history-dependent minor loops, and transient response.
//
// References:
//   - PFECAP Verilog-A model: https://github.com/nctu-ms07/PFECAP
//   - J.-P. Doré et al., "PFECAP: Compact model for ferroelectric capacitors"
package compact

import (
	"errors"
	"math"
)

// ── FeCap model ───────────────────────────────────────────────────────────────

// FeCapParams holds the physical parameters for one ferroelectric capacitor.
// All parameters should be calibrated against measured C-V or P-V curves.
type FeCapParams struct {
	// Thickness of the ferroelectric film (m). Typical: 10e-9 (10 nm HZO).
	ThicknessFE float64

	// Coercive voltage across the FE film (V). Often ≈ Ec × ThicknessFE.
	CoerciveV float64

	// Relative permittivity of the FE layer (linear dielectric component).
	// Typical HZO: ε_r ≈ 25–35.
	EpsFEr float64

	// Saturation polarisation charge density (C/m²). Typical HZO: ~25 µC/cm² = 0.25 C/m².
	Qs float64

	// Switching steepness parameter (1/V). Larger a → sharper switching.
	// Typical value: a = 1 / (0.1 × CoerciveV).
	A float64

	// Series resistance (Ω). Set to 0 for ideal capacitor.
	Rho float64
}

// DefaultFeCapHZO returns a default parameter set calibrated for HfO₂-based FeCap.
// These are approximate simulation defaults — calibrate against measured data for
// quantitative claims.
func DefaultFeCapHZO() FeCapParams {
	tFE := 10e-9   // 10 nm HZO
	Ec := 1.0e6    // 1 MV/m coercive field
	Vc := Ec * tFE // ~1 V coercive voltage
	return FeCapParams{
		ThicknessFE: tFE,
		CoerciveV:   Vc,
		EpsFEr:      30.0,
		Qs:          0.25,                // 25 µC/cm²
		A:           1.0 / (0.1 * Vc),   // steep switching
		Rho:         0.0,
	}
}

// ── Switching function ────────────────────────────────────────────────────────

// SwitchingFunction returns the polarisation charge (C/m²) for a given applied
// voltage V and polarisation direction dir (+1 or -1).
//
// Implements PFECAP's F(V, dir):
//
//	F(V, dir) = Qs × tanh(a × (V - dir × Vc))
//
// where Vc = p.CoerciveV and a = p.A.
func (p *FeCapParams) SwitchingFunction(V, dir float64) float64 {
	return p.Qs * math.Tanh(p.A*(V-dir*p.CoerciveV))
}

// ── Linear dielectric contribution ───────────────────────────────────────────

const epsilon0 = 8.854187817e-12 // F/m

// LinearCharge returns the linear dielectric charge (C/m²) for voltage V.
//
//	Q_lin = ε₀ × εᵣ × V / tFE
func (p *FeCapParams) LinearCharge(V float64) float64 {
	return epsilon0 * p.EpsFEr * V / p.ThicknessFE
}

// ── History-tracking FeCap instance ──────────────────────────────────────────

// FeCap is a stateful ferroelectric capacitor instance with Preisach-style history.
//
// Turning-point arrays (aV, bV) track voltage reversal events as ascending/descending
// branch pairs — matching PFECAP's a_V[]/b_V[] history management.
// Call Reset() to erase history (fresh erase from a known state before use).
type FeCap struct {
	p   FeCapParams
	aV  []float64 // ascending turning-point voltages (branch starts)
	bV  []float64 // descending turning-point voltages (branch starts)
	dir float64   // current polarisation direction: +1 or -1
}

// NewFeCap creates a new FeCap instance starting in the negative saturation state
// (fully erased toward negative polarisation).
func NewFeCap(p FeCapParams) *FeCap {
	return &FeCap{
		p:   p,
		aV:  nil,
		bV:  nil,
		dir: -1.0,
	}
}

// Reset clears the voltage history, returning the device to a fresh state.
// The direction (dir) is preserved — call Reset(-1) or set dir manually after
// Reset to put the device in a known erased state.
func (fc *FeCap) Reset() {
	fc.aV = nil
	fc.bV = nil
}

// TotalCharge returns the instantaneous total charge density (C/m²) at voltage V,
// including both the ferroelectric polarisation and the linear dielectric term.
//
// For the current direction branch:
//
//	Q(V) = F(V, dir) + ε₀ × εᵣ × V / tFE
func (fc *FeCap) TotalCharge(V float64) float64 {
	return fc.p.SwitchingFunction(V, fc.dir) + fc.p.LinearCharge(V)
}

// dTotalCharge returns dQ/dV at voltage V for Newton-solver convergence.
//
//	dQ/dV = Qs × a × sech²(a × (V - dir × Vc))  +  ε₀ × εᵣ / tFE
func (fc *FeCap) dTotalCharge(V float64) float64 {
	arg := fc.p.A * (V - fc.dir*fc.p.CoerciveV)
	sech2 := 1.0 / math.Cosh(arg)
	sech2 *= sech2
	return fc.p.Qs*fc.p.A*sech2 + epsilon0*fc.p.EpsFEr/fc.p.ThicknessFE
}

// SolveVFE finds the FE voltage V given a target total charge Qtarget (C/m²)
// using Newton iteration. Returns an error if the solver fails to converge.
//
// This is equivalent to PFECAP's Newton solver for Q(V) = Qtarget.
func (fc *FeCap) SolveVFE(Qtarget float64) (float64, error) {
	const maxIter = 50
	const tol = 1e-10

	V := fc.dir * fc.p.CoerciveV // initial guess: near coercive voltage
	for i := 0; i < maxIter; i++ {
		f := fc.TotalCharge(V) - Qtarget
		df := fc.dTotalCharge(V)
		if math.Abs(df) < 1e-30 {
			return 0, errors.New("fecap: Newton solver stalled (zero derivative)")
		}
		dV := -f / df
		V += dV
		if math.Abs(dV) < tol {
			return V, nil
		}
	}
	return 0, errors.New("fecap: Newton solver did not converge")
}

// UpdateDirection records a voltage reversal if V crossed the current branch tip,
// appending to the turning-point history (matching PFECAP's append logic).
// Returns the new direction (+1 ascending, -1 descending).
func (fc *FeCap) UpdateDirection(V float64) {
	if fc.dir < 0 && len(fc.aV) > 0 && V > fc.aV[len(fc.aV)-1] {
		// Reversal: was descending, now ascending
		fc.dir = +1
		fc.bV = append(fc.bV, V)
	} else if fc.dir > 0 && len(fc.bV) > 0 && V < fc.bV[len(fc.bV)-1] {
		// Reversal: was ascending, now descending
		fc.dir = -1
		fc.aV = append(fc.aV, V)
	}
}

// PEPoint is a single point on a polarisation–electric-field (P-E) or Q-V curve.
type PEPoint struct {
	V float64 // applied voltage (V)
	Q float64 // total charge density (C/m²)
}

// SweepPELoop generates a full P-E hysteresis loop by sweeping voltage from
// -Vmax to +Vmax and back, returning nPoints per half-cycle.
// The device history is reset before the sweep.
func (fc *FeCap) SweepPELoop(Vmax float64, nPoints int) []PEPoint {
	fc.Reset()
	fc.dir = -1.0

	points := make([]PEPoint, 0, 2*nPoints)

	// Negative-to-positive sweep
	for i := 0; i < nPoints; i++ {
		V := -Vmax + (2*Vmax)*float64(i)/float64(nPoints-1)
		points = append(points, PEPoint{V: V, Q: fc.TotalCharge(V)})
	}

	// Positive-to-negative sweep (return branch)
	fc.dir = +1.0
	for i := nPoints - 1; i >= 0; i-- {
		V := -Vmax + (2*Vmax)*float64(i)/float64(nPoints-1)
		points = append(points, PEPoint{V: V, Q: fc.TotalCharge(V)})
	}

	return points
}

// Capacitance returns the small-signal capacitance density (F/m²) at voltage V.
// Equals dQ/dV evaluated at the current direction branch.
func (fc *FeCap) Capacitance(V float64) float64 {
	return fc.dTotalCharge(V)
}
