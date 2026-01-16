// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"math"
)

// PreisachModel implements the Preisach hysteresis model for ferroelectrics.
// Based on Bartic et al. (2001) "Preisach model for the simulation of
// ferroelectric capacitors" and Bo Jiang's hyperbolic tangent method.
type PreisachModel struct {
	material *HZOMaterial

	// Distribution parameters
	EcMean  float64 // Mean coercive field
	EcSigma float64 // Coercive field distribution width
	EuMean  float64 // Mean interaction field
	EuSigma float64 // Interaction field distribution width

	// History tracking (LIFO stack for turning points)
	turningPoints []float64
	lastE         float64
	increasing    bool

	// Current state
	polarization float64
}

// NewPreisachModel creates a new Preisach model with the given material.
func NewPreisachModel(material *HZOMaterial) *PreisachModel {
	return &PreisachModel{
		material:      material,
		EcMean:        material.Ec,
		EcSigma:       material.Ec * 0.25, // 25% distribution width
		EuMean:        0,
		EuSigma:       material.Ec * 0.4,
		turningPoints: make([]float64, 0, 100),
		polarization:  0,
	}
}

// Reset clears the history and sets polarization to zero.
func (p *PreisachModel) Reset() {
	p.turningPoints = p.turningPoints[:0]
	p.polarization = 0
	p.lastE = 0
}

// Update applies a new electric field and returns the resulting polarization.
// The field E should be in V/m.
func (p *PreisachModel) Update(E float64) float64 {
	// Determine direction
	increasing := E > p.lastE

	// Check for turning point (direction change)
	if len(p.turningPoints) > 0 && increasing != p.increasing {
		p.addTurningPoint(p.lastE)
	}

	// Calculate polarization using hyperbolic tangent model
	// This captures the S-shaped switching characteristic
	p.polarization = p.calculatePolarization(E)

	// Update state
	p.lastE = E
	p.increasing = increasing

	return p.polarization
}

// calculatePolarization computes P(E) using the Preisach distribution.
func (p *PreisachModel) calculatePolarization(E float64) float64 {
	// Hyperbolic tangent switching function (Bo Jiang method)
	// P = Ps * tanh((E - Ec_eff) / delta)
	// where delta controls the switching sharpness

	Ps := p.material.Ps
	Ec := p.EcMean
	delta := p.EcSigma * 2 // Transition width

	// Calculate effective coercive field based on history
	EcEff := p.effectiveCoerciveField()

	// Base switching function
	var P float64
	if p.increasing {
		// Ascending branch
		P = Ps * math.Tanh((E-EcEff)/delta)
	} else {
		// Descending branch
		P = Ps * math.Tanh((E+EcEff)/delta)
	}

	// Apply history correction for minor loops
	P = p.applyHistoryCorrection(P, E)

	return P
}

// effectiveCoerciveField returns Ec modified by the Preisach distribution.
func (p *PreisachModel) effectiveCoerciveField() float64 {
	// In a full Preisach model, this would integrate over the distribution
	// For simplicity, we use the mean with small random variation
	return p.EcMean
}

// addTurningPoint records a reversal in the field sweep direction.
func (p *PreisachModel) addTurningPoint(E float64) {
	// Implement memory wipe-out: a turning point erases smaller previous ones
	for len(p.turningPoints) > 0 {
		last := p.turningPoints[len(p.turningPoints)-1]
		if (p.increasing && E > last) || (!p.increasing && E < last) {
			// Wipe out the smaller turning point
			p.turningPoints = p.turningPoints[:len(p.turningPoints)-1]
		} else {
			break
		}
	}
	p.turningPoints = append(p.turningPoints, E)
}

// applyHistoryCorrection adjusts P based on the turning point history.
func (p *PreisachModel) applyHistoryCorrection(P, E float64) float64 {
	if len(p.turningPoints) == 0 {
		return P
	}

	// For minor loops, interpolate between major loop branches
	// This is a simplified implementation
	Ps := p.material.Ps

	// Calculate the "closure" of minor loops
	for i := len(p.turningPoints) - 1; i >= 0; i-- {
		tp := p.turningPoints[i]
		if p.increasing {
			if E >= tp {
				// Close the minor loop
				p.turningPoints = p.turningPoints[:i]
			}
		} else {
			if E <= tp {
				p.turningPoints = p.turningPoints[:i]
			}
		}
	}

	// Clamp to saturation
	if P > Ps {
		P = Ps
	} else if P < -Ps {
		P = -Ps
	}

	return P
}

// Polarization returns the current polarization state.
func (p *PreisachModel) Polarization() float64 {
	return p.polarization
}

// NormalizedPolarization returns polarization as fraction of Ps (-1 to +1).
func (p *PreisachModel) NormalizedPolarization() float64 {
	return p.polarization / p.material.Ps
}

// GetHysteresisLoop generates a full P-E hysteresis loop.
// Returns slices of E and P values for plotting.
func (p *PreisachModel) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64) {
	p.Reset()

	E := make([]float64, 0, points*4)
	P := make([]float64, 0, points*4)

	// Sweep from 0 to +Emax
	for i := 0; i <= points; i++ {
		e := Emax * float64(i) / float64(points)
		pol := p.Update(e)
		E = append(E, e)
		P = append(P, pol)
	}

	// Sweep from +Emax to -Emax
	for i := 0; i <= points*2; i++ {
		e := Emax - 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		P = append(P, pol)
	}

	// Sweep from -Emax back to +Emax
	for i := 0; i <= points*2; i++ {
		e := -Emax + 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		P = append(P, pol)
	}

	return E, P
}

// DiscreteStates returns polarization values for N discrete analog states.
// This demonstrates the 30-state capability of IronLattice.
func (p *PreisachModel) DiscreteStates(N int) []float64 {
	states := make([]float64, N)
	Ps := p.material.Ps

	for i := 0; i < N; i++ {
		// Linear spacing from -Ps to +Ps
		states[i] = -Ps + 2*Ps*float64(i)/float64(N-1)
	}

	return states
}
