// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"math"

	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
)

// Package-level logger
var log *logging.Logger

func init() {
	log = logging.NewLogger("preisach")
}

// TanhEverett implements the EverettFunction interface using Bo Jiang's Tanh model.
// This preserves the existing "S-shape" characteristic of the simulation.
type TanhEverett struct {
	Ps    float64
	Ec    float64
	Delta float64 // Distribution width
}

// Calculate returns the polarization contribution for a hysteron switching region.
// Note: The shared/physics.PreisachStack uses this to sum triangular areas.
// For the Tanh model, we treat the major loop function P_major(E) as the
// integral of the Everett function along the line beta = -E_sat.
// Simplification: We return P_major(alpha) - P_major(beta) ??
//
// Actually, the shared stack implementation computes:
// P = -Ps + 2 * Sum( Everett(Max, Min) )
// If we want to match the analytic Tanh model:
// P(E) = Ps * Tanh((E - Ec)/Delta)
//
// We need to map the "Everett(Max, Min)" call to something that reconstructs this.
// A common approximation for independent hysterons (like Tanh) is:
// Everett(alpha, beta) = (F(alpha) - F(beta)) / 2 * Ps
// where F is the cumulative distribution function (the major loop normalized to 0..1).
//
// Let's implement this factorization.
func (t *TanhEverett) Calculate(alpha, beta float64) float64 {
	// Product-form Everett function for the factorized sech² Preisach density:
	//
	//   µ(α,β) = (Ps / 4Δ²) · sech²((α − Ec)/Δ) · sech²((β + Ec)/Δ)
	//
	// The Everett integral over the triangle {α' ≤ α, β' ≥ β} gives:
	//
	//   E(α,β) = (Ps/4) · [1 + tanh((α − Ec)/Δ)] · [1 − tanh((β + Ec)/Δ)]
	//
	// Both factors lie in [0, 2], so the product is always non-negative.
	// This preserves the Preisach wipe-out continuity property that the
	// previous factorized-difference form violated (it went negative for
	// minor loops within the coercive gap, requiring a hard zero-clamp
	// that caused polarization teleportation during ISPP programming).
	//
	// The major loop shape and remanent polarization are identical to the
	// difference form because for Esat >> Ec the limit factors collapse
	// to the same expression.
	ascCDF := 1.0 + math.Tanh((alpha-t.Ec)/t.Delta)   // CDF of ascending distribution ∈ [0, 2]
	descSurv := 1.0 - math.Tanh((beta+t.Ec)/t.Delta)  // Survival of descending distribution ∈ [0, 2]

	val := ascCDF * descSurv * t.Ps * 0.25
	if val > t.Ps {
		return t.Ps // Safety clamp for floating-point edge cases
	}
	return val
}

// tuneDeltaForPr estimates a Tanh Everett distribution width (Delta) so that
// the remanent polarization after a full saturation-and-return matches targetPr.
// This keeps the Preisach loop consistent with the material Pr/Ps ratio.
func tuneDeltaForPr(ec, saturationE, psIrrev, targetPr float64) float64 {
	if ec <= 0 {
		return 0
	}
	if psIrrev <= 0 || targetPr <= 0 {
		return ec * 0.25
	}

	satE := math.Abs(saturationE)
	if satE <= 0 {
		satE = ec * 5.0
	}

	targetRatio := targetPr / psIrrev
	if targetRatio <= 0 {
		return ec * 2.0
	}
	if targetRatio > 0.999 {
		targetRatio = 0.999
	}
	if targetRatio < 0.01 {
		targetRatio = 0.01
	}

	ratioFor := func(delta float64) float64 {
		if delta <= 0 {
			return 0
		}
		everett := &TanhEverett{
			Ps:    psIrrev,
			Ec:    ec,
			Delta: delta,
		}
		stack := physics.NewPreisachStack(satE, everett)
		stack.Update(satE)
		pr := stack.Update(0)
		if psIrrev == 0 {
			return 0
		}
		ratio := pr / psIrrev
		if ratio < 0 {
			return 0
		}
		if ratio > 1 {
			return 1
		}
		return ratio
	}

	lo := ec * 0.05
	hi := ec * 2.0
	rLo := ratioFor(lo)
	rHi := ratioFor(hi)
	if rLo < rHi {
		lo, hi = hi, lo
		rLo, rHi = rHi, rLo
	}

	// Expand search bounds if needed to bracket target.
	for rLo < targetRatio && lo > ec*1e-6 {
		lo *= 0.5
		rLo = ratioFor(lo)
	}
	for rHi > targetRatio && hi < ec*10.0 {
		hi *= 1.5
		rHi = ratioFor(hi)
	}

	if targetRatio >= rLo {
		return lo
	}
	if targetRatio <= rHi {
		return hi
	}

	for i := 0; i < 32; i++ {
		mid := 0.5 * (lo + hi)
		rMid := ratioFor(mid)
		if rMid > targetRatio {
			lo = mid
		} else {
			hi = mid
		}
	}
	return 0.5 * (lo + hi)
}

// PreisachModel implements the Preisach hysteresis model for ferroelectrics.
// Wraps the shared/physics.PreisachStack engine.
// PreisachModel implements the Preisach hysteresis model for ferroelectrics.
// Wraps the shared/physics.PreisachStack engine.
type PreisachModel struct {
	material *HZOMaterial
	stack    *physics.PreisachStack
	everett  *TanhEverett

	Temperature float64
	Stress      float64 // GPa

	// Reversible (nonlinear) contribution derived from permittivity and Ec.
	reversibleChi  float64 // C/(V*m)
	reversiblePSat float64 // Saturating reversible polarization (C/m^2)

	effectivePs float64 // Temperature/stress-adjusted total Ps
}

// NewPreisachModel creates a new Preisach model with the given material.
func NewPreisachModel(material *HZOMaterial) *PreisachModel {
	log.Input("NewPreisachModel", map[string]interface{}{
		"material_name": material.Name,
		"Ec":            material.Ec,
		"Ps":            material.Ps,
	})

	// Configure Everett function based on material
	everett := &TanhEverett{
		Ps:    material.Ps,
		Ec:    material.Ec,
		Delta: material.Ec * 0.25, // Initial guess; tuned to match Pr in updateReversibleParams
	}

	// E_saturation should be > Ec. typically 3-5x Ec.
	E_sat := material.Ec * 5.0

	model := &PreisachModel{
		material:    material,
		stack:       physics.NewPreisachStack(E_sat, everett),
		everett:     everett,
		Temperature: 300.0,
		Stress:      1.0, // Default 1 GPa
		effectivePs: material.Ps,
	}
	model.updateReversibleParams()
	return model
}

// DiscreteState represents a single programmable state.
type DiscreteState struct {
	Level        int
	Polarization float64
	NormalizedP  float64
	Voltage      float64
	Conductance  float64
}

// DiscreteStates returns the polarization values for n evenly spaced discrete states.
// This is a helper for testing and visualization.
// DiscreteStates returns the discrete states for n evenly spaced levels.
func (p *PreisachModel) DiscreteStates(n int) []DiscreteState {
	states := make([]DiscreteState, n)
	step := 2.0 * p.material.Ps / float64(n-1)
	for i := 0; i < n; i++ {
		pol := -p.material.Ps + float64(i)*step
		states[i] = DiscreteState{
			Level:        i + 1,
			Polarization: pol,
			NormalizedP:  pol / p.material.Ps,
			Voltage:      0, // Placeholder
			Conductance:  0, // Placeholder
		}
	}
	return states
}

// Reset clears the history and sets the model to negative saturation
// (including the reversible dielectric contribution at -E_sat).
func (p *PreisachModel) Reset() {
	// Re-initialize stack
	E_sat := p.material.Ec * 5.0
	everett := p.stack.Everett
	p.stack = physics.NewPreisachStack(E_sat, everett)
}

// Update applies a new electric field and returns the resulting polarization.
func (p *PreisachModel) Update(E float64) float64 {
	if logging.IsVerbose(logging.VerbosityTrace) {
		log.Input("Update", map[string]interface{}{"E": E})
	}

	Pirrev := p.stack.Update(E)
	P := Pirrev + p.reversiblePolarization(E)

	if logging.IsVerbose(logging.VerbosityTrace) {
		log.Calculation("Update", map[string]interface{}{"E": E}, P)
	}
	return P
}

// Polarization returns the current polarization state.
func (p *PreisachModel) Polarization() float64 {
	// Compute polarization at current field without mutating history.
	Pirrev := p.stack.ComputePolarization(p.stack.LastE)
	return Pirrev + p.reversiblePolarization(p.stack.LastE)
}

// NormalizedPolarization returns polarization as fraction of Ps.
func (p *PreisachModel) NormalizedPolarization() float64 {
	denom := p.effectivePs
	if denom == 0 {
		denom = p.material.Ps
	}
	if denom == 0 {
		return 0
	}
	return p.Polarization() / denom
}

// reversiblePolarization returns the nonlinear reversible contribution.
// The small-signal slope matches the material permittivity.
func (p *PreisachModel) reversiblePolarization(E float64) float64 {
	if p.reversiblePSat == 0 || p.everett == nil || p.everett.Ec <= 0 {
		return 0
	}
	return p.reversiblePSat * math.Tanh(E/p.everett.Ec)
}

func (p *PreisachModel) updateReversibleParams() {
	if p.material == nil || p.everett == nil {
		return
	}

	// Linear susceptibility from permittivity (low frequency preferred).
	const epsilon0 = 8.854e-12
	if p.material.EpsilonLF > 1 {
		p.reversibleChi = epsilon0 * (p.material.EpsilonLF - 1.0)
	} else if p.material.Epsilon > 1 {
		p.reversibleChi = epsilon0 * (p.material.Epsilon - 1.0)
	} else {
		p.reversibleChi = 0
	}

	if p.reversibleChi > 0 && p.everett.Ec > 0 {
		p.reversiblePSat = p.reversibleChi * p.everett.Ec
	} else {
		p.reversiblePSat = 0
	}

	// Split saturation into irreversible + reversible components so total Ps is preserved.
	totalPs := p.effectivePs
	if totalPs == 0 {
		totalPs = p.material.Ps
	}
	psIrrev := totalPs - p.reversiblePSat
	if psIrrev < 0 {
		psIrrev = 0
		p.reversiblePSat = totalPs
	}
	p.everett.Ps = psIrrev

	// Tune Delta so that remanent polarization matches material Pr.
	satE := p.everett.Ec * 5.0
	if p.stack != nil && p.stack.SaturationE > 0 {
		satE = p.stack.SaturationE
	}
	p.everett.Delta = tuneDeltaForPr(p.everett.Ec, satE, p.everett.Ps, p.material.Pr)
}

// GetHysteresisLoop generates a full P-E hysteresis loop.
func (p *PreisachModel) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64) {
	// Temporarily reset stack to generate loop
	// Ideally we clone, but for GUI generation we typically want a fresh loop
	p.Reset()

	E := make([]float64, 0, points*4)
	PVal := make([]float64, 0, points*4) // renamed to avoid collision with P() method

	// Saturation start
	p.Update(-Emax)

	// Ascending
	for i := 0; i <= points*2; i++ {
		e := -Emax + 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		PVal = append(PVal, pol)
	}

	// Descending
	for i := 1; i <= points*2; i++ {
		e := Emax - 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		PVal = append(PVal, pol)
	}

	return E, PVal
}

// SetTemperature updates the simulation temperature and scales material parameters.
func (p *PreisachModel) SetTemperature(tempK float64) {
	p.Temperature = tempK

	// Scale Ec and Ps based on linear temperature coefficients
	// Ec(T) = Ec0 + Coeff * (T - 300)
	// Note: We use 300K as reference, assuming params are defined at RT

	deltaT := tempK - 300.0

	// Retrieve coefficients from material config (HZOMaterial needs these fields exposed)
	// Assuming HZOMaterial has TempCoeffEc and TempCoeffPr

	newEc := p.material.Ec + p.material.TempCoeffEc*deltaT
	newPs := p.material.Ps + p.material.TempCoeffPr*deltaT

	// Safety clamps
	if newEc < 1e5 {
		newEc = 1e5
	} // Minimum 0.001 MV/cm
	if newPs < 1e-6 {
		newPs = 1e-6
	} // Minimum polarization

	// Update Everett function
	p.everett.Ec = newEc
	p.effectivePs = newPs
	p.updateReversibleParams()
}

// GetEffectiveEc returns the current temperature-scaled Coercive Field.
func (p *PreisachModel) GetEffectiveEc() float64 {
	return p.everett.Ec
}

// SetStress updates the mechanical stress and scales Ec accordingly.
// Stress is in GPa.
// Scaling Logic: Ec ~ sqrt(|Alpha|)
// Alpha = AlphaT - 2*Q12*Stress
func (p *PreisachModel) SetStress(stressGPa float64) {
	p.Stress = stressGPa

	// Recalculate everything (Temperature and Stress)
	p.updateEffectiveParameters()
}

// updateEffectiveParameters recalculates Ec and Ps based on T and Stress
func (p *PreisachModel) updateEffectiveParameters() {
	// Base parameters at 300K, 1GPa (if calibrated there) or 0GPa?
	// Let's assume material.Ec is at 300K and defined Stress (usually 1GPa for HZO).
	// For simplicity, we use linear temp scaling + alpha-based stress scaling relative to baseline.

	// 1. Temperature Scaling
	deltaT := p.Temperature - 300.0
	ec_T := p.material.Ec + p.material.TempCoeffEc*deltaT
	ps_T := p.material.Ps + p.material.TempCoeffPr*deltaT

	// 2. Stress Scaling
	// Relative change in Alpha
	// Alpha_ref = Alpha(300K, 1GPa)
	// Alpha_new = Alpha(T, Stress)

	// Need material Landau coefficients. If not available, assume default.
	// HZO defaults:
	// Q12 ~ -0.026
	// Alpha_0 ~ -1e9 (at 300K)

	// If material struct doesn't have Q12, we can't do accurate stress scaling.
	// We will skip stress scaling if Q12 is 0 or missing, to avoid breaking logic.

	// Assuming linear stress effect on Ec for now if we lack full Landau params:
	// Ec increases with tensile stress (negative Q12 makes alpha more negative).
	// Sensitivity: dEc/dStress approx 0.1 MV/cm per GPa?

	// Let's implement a simplified linear sensitivity based on literature if coefficients missing
	// dEc/dSigma ~ +5% per GPa for HZO

	stressFactor := 1.0 + 0.05*(p.Stress-1.0) // Relative to 1GPa baseline

	newEc := ec_T * stressFactor
	newPs := ps_T // Ps is less sensitive to stress in first order

	// Safety
	if newEc < 1e5 {
		newEc = 1e5
	}
	if newPs < 1e-6 {
		newPs = 1e-6
	}

	// Update Everett
	p.everett.Ec = newEc
	p.effectivePs = newPs
	p.updateReversibleParams()
}
