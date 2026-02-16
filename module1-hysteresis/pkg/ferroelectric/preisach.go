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

// TanhEverett is a compatibility alias for shared Preisach Everett adapter.
type TanhEverett = physics.TanhEverett

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

	// Kinetics
	nls *physics.NLSKinetics
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
		nls:         physics.NewNLSKinetics(),
	}
	model.updateReversibleParams()

	// Scale Ea based on Ec (empirical heuristic: Ea approx 10-20 * Ec)
	// We'll trust the material Ec to be a proxy for the energy barrier.
	// But NLSKinetics has a default Ea=8e8 (8 MV/cm).
	// Let's adjust it if Ec is significantly different from typical 1 MV/cm.
	if material.Ec > 0 {
		model.nls.Ea = material.Ec * 10.0
	}

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
// This assumes quasi-static conditions (infinite time/instant switching).
func (p *PreisachModel) Update(E float64) float64 {
	return p.TimeStep(E, 1.0) // 1 second is effectively infinite for typical NLS
}

// TimeStep applies a constant electric field E for duration dt (seconds).
// Returns the resulting polarization after relaxation.
func (p *PreisachModel) TimeStep(E, dt float64) float64 {
	// 1. Calculate Static Equilibrium (Infinite Retention target)
	// This updates the Preisach stack history to the new field level immediately,
	// effectively determining WHERE the system would go if given infinite time.
	// This handles the "wipe-out" logic and minor loop nesting for the target state.

	// We need the *previous* state P_old before moving the stack?
	// Actually, the stack represents the Magnetic/Electric field history.
	// We need to track the "dynamic" polarization separately if it lags the stack.
	// For now, to keep it simple and stateless (w.r.t dynamic lag vars),
	// we will calculate the static target, then assume we started from
	// the *current* P value (calculated from previous history + previous lag?).
	//
	// Problem: The standard Preisach implementation here is stateful only in turning points.
	// It doesn't store a "current P" that might be lagging behind the turning points.
	// If we update the stack, we change the "static" P immediately.

	// Approximaton:
	// The stack represents the "Driving Force" history.
	// P_eq is the result of `ps.stack.Update(E)`.
	// Use NLS to relax towards P_eq.

	// 1. Get current static state (before update)
	// P_old_static := p.Polarization()

	// Ideally we would carry a p.currentP_dynamic state.
	// But for ISPP (step-and-hold), we usually start from a relaxed state.
	// Let's assume P_start is the *previous* stack state (at LastE).
	P_start := p.Polarization()

	// 2. Update Stack to new Field E -> P_target
	Pirrev_target := p.stack.Update(E)
	P_target := Pirrev_target + p.reversiblePolarization(E)

	// 3. Relax P_start -> P_target
	if p.nls == nil {
		p.nls = physics.NewNLSKinetics()
	}
	P_final := p.nls.Relax(P_start, P_target, E, dt)

	if logging.IsVerbose(logging.VerbosityTrace) {
		log.Calculation("TimeStep", map[string]interface{}{
			"E": E, "dt": dt, "P_start": P_start, "P_target": P_target, "P_final": P_final,
		}, P_final)
	}
	return P_final
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
	p.updateEffectiveParameters()
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

// updateEffectiveParameters recalculates Ec and Ps based on Curie-Weiss temperature
// scaling and electrostriction stress coupling.
func (p *PreisachModel) updateEffectiveParameters() {
	if p.material == nil || p.everett == nil {
		return
	}

	const epsilon0 = 8.854e-12

	newEc := p.material.Ec
	newPs := p.material.Ps + p.material.TempCoeffPr*(p.Temperature-300.0)

	if p.material.CurieConst != 0 {
		alphaT := (p.Temperature - p.material.CurieTemp) / (2 * epsilon0 * p.material.CurieConst)
		alphaRef := (300.0 - p.material.CurieTemp) / (2 * epsilon0 * p.material.CurieConst)

		if alphaRef != 0 {
			q12 := p.material.Q12
			if q12 == 0 {
				q12 = -0.026 // HZO default from LK material params
			}
			alphaStress := alphaT - 2*q12*p.Stress*1e9 // alpha(σ) = alpha(T) - 2*Q12*σ
			ecRatio := math.Pow(math.Abs(alphaStress/alphaRef), 1.5)
			newEc = p.material.Ec * ecRatio
		}
	}

	if newEc < 1e5 {
		newEc = 1e5
	}
	if newPs < 1e-6 {
		newPs = 1e-6
	}

	p.everett.Ec = newEc
	p.effectivePs = newPs
	p.updateReversibleParams()
}
