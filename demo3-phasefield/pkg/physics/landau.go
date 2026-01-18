// Package physics provides TDGL phase-field physics models for ferroelectrics.
//
// Physics references:
// - Landau-Khalatnikov: Sivasubramanian & Widom, arXiv:cond-mat/0108189v1 (2001)
// - Preisach model: Urbanavičiūtė et al., Nature Communications 9:4409 (2018)
package physics

import (
	"math"
)

// LandauPotential computes the Landau free energy density f(P).
//
// The 2-4-6 Landau expansion:
//   f(P) = α·P² + β·P⁴ + γ·P⁶
//
// This is Eq. (14) from Sivasubramanian & Widom (2001) in generalized form:
//   U(Q) = (Qs²/8C)[1 - (Q/Qs)²]²
//
// For ferroelectrics with first-order transitions (like HZO), β < 0 and γ > 0
// creates the characteristic double-well potential with minima at ±Ps.
type LandauPotential struct {
	Alpha float64 // Quadratic coefficient
	Beta  float64 // Quartic coefficient
	Gamma float64 // Sixth-order coefficient
}

// NewLandauPotential creates a new Landau potential with given coefficients.
func NewLandauPotential(alpha, beta, gamma float64) *LandauPotential {
	return &LandauPotential{
		Alpha: alpha,
		Beta:  beta,
		Gamma: gamma,
	}
}

// FromMaterial creates a Landau potential from material parameters at given T.
func FromMaterial(m *HZOMaterial, T float64) *LandauPotential {
	return &LandauPotential{
		Alpha: m.AlphaTemperature(T),
		Beta:  m.Beta,
		Gamma: m.Gamma,
	}
}

// Energy computes the Landau free energy density at polarization P.
// f(P) = α·P² + β·P⁴ + γ·P⁶
func (lp *LandauPotential) Energy(P float64) float64 {
	P2 := P * P
	P4 := P2 * P2
	P6 := P4 * P2
	return lp.Alpha*P2 + lp.Beta*P4 + lp.Gamma*P6
}

// Derivative computes df/dP, the thermodynamic electric field.
//
// df/dP = 2α·P + 4β·P³ + 6γ·P⁵
//
// This appears in the Landau-Khalatnikov equation (Eq. 3 from Sivasubramanian & Widom):
//   E = (∂U/∂P)_S + ρ(dP/dt)
//
// At equilibrium (dP/dt = 0), E = df/dP.
func (lp *LandauPotential) Derivative(P float64) float64 {
	P2 := P * P
	P3 := P2 * P
	P5 := P3 * P2
	return 2*lp.Alpha*P + 4*lp.Beta*P3 + 6*lp.Gamma*P5
}

// SecondDerivative computes d²f/dP².
// d²f/dP² = 2α + 12β·P² + 30γ·P⁴
func (lp *LandauPotential) SecondDerivative(P float64) float64 {
	P2 := P * P
	P4 := P2 * P2
	return 2*lp.Alpha + 12*lp.Beta*P2 + 30*lp.Gamma*P4
}

// Minima finds the equilibrium polarization values (where df/dP = 0).
// Returns P=0 and ±P_s for T < Tc.
func (lp *LandauPotential) Minima() []float64 {
	// df/dP = 2αP + 4βP³ + 6γP⁵ = 0
	// P(2α + 4βP² + 6γP⁴) = 0
	// Solutions: P = 0, or 2α + 4βP² + 6γP⁴ = 0

	minima := []float64{0} // P=0 always a solution

	// For non-trivial solutions: 6γP⁴ + 4βP² + 2α = 0
	// Let u = P², then: 6γu² + 4βu + 2α = 0
	// u = (-4β ± sqrt(16β² - 48γα)) / (12γ)

	if lp.Gamma == 0 {
		// Simplified case: 4βP² + 2α = 0 → P² = -α/(2β)
		if lp.Beta != 0 {
			u := -lp.Alpha / (2 * lp.Beta)
			if u > 0 {
				Ps := math.Sqrt(u)
				minima = append(minima, Ps, -Ps)
			}
		}
		return minima
	}

	// Full sixth-order case
	discriminant := 16*lp.Beta*lp.Beta - 48*lp.Gamma*lp.Alpha

	if discriminant >= 0 {
		sqrtD := math.Sqrt(discriminant)
		u1 := (-4*lp.Beta + sqrtD) / (12 * lp.Gamma)
		u2 := (-4*lp.Beta - sqrtD) / (12 * lp.Gamma)

		if u1 > 0 {
			Ps := math.Sqrt(u1)
			minima = append(minima, Ps, -Ps)
		}
		if u2 > 0 && u2 != u1 {
			Ps := math.Sqrt(u2)
			minima = append(minima, Ps, -Ps)
		}
	}

	return minima
}

// SpontaneousPolarization returns |P_s|, the stable spontaneous polarization.
// For T < Tc, this is the non-zero minimum.
func (lp *LandauPotential) SpontaneousPolarization() float64 {
	minima := lp.Minima()

	// Find the minimum with lowest energy (excluding P=0 if other minima exist)
	var bestP float64
	bestE := math.Inf(1)

	for _, P := range minima {
		E := lp.Energy(P)
		if E < bestE && (P != 0 || len(minima) == 1) {
			bestE = E
			bestP = P
		}
	}

	return math.Abs(bestP)
}

// IsAboveTc returns true if the system is in the paraelectric phase (only P=0 is stable).
func (lp *LandauPotential) IsAboveTc() bool {
	// If α > 0 and β > 0, only P=0 is stable
	// For HZO, β < 0 and γ > 0, so check if α is large enough
	minima := lp.Minima()

	// Check if P=0 is the only minimum
	if len(minima) == 1 && minima[0] == 0 {
		return true
	}

	// Check if P=0 has the lowest energy
	E0 := lp.Energy(0)
	for _, P := range minima {
		if P != 0 && lp.Energy(P) < E0 {
			return false
		}
	}

	return true
}

// BarrierHeight computes the energy barrier between ±P_s states.
// This is the energy at P=0 relative to the minima.
func (lp *LandauPotential) BarrierHeight() float64 {
	Ps := lp.SpontaneousPolarization()
	if Ps == 0 {
		return 0 // No barrier in paraelectric phase
	}

	E0 := lp.Energy(0)
	EPs := lp.Energy(Ps)

	return E0 - EPs
}

// EnergyLandscape generates (P, f(P)) pairs for plotting.
func (lp *LandauPotential) EnergyLandscape(Pmax float64, nPoints int) ([]float64, []float64) {
	P := make([]float64, nPoints)
	f := make([]float64, nPoints)

	dp := 2 * Pmax / float64(nPoints-1)
	for i := 0; i < nPoints; i++ {
		P[i] = -Pmax + float64(i)*dp
		f[i] = lp.Energy(P[i])
	}

	return P, f
}

// CoerciveField estimates the coercive field from Landau theory.
//
// From Urbanavičiūtė et al. (2018), Eq. (2):
//   E_c ≈ w_b/P_r - k_B·T·ln(ν₀·t·ln(2)⁻¹)/(P_r·V*)
//
// For pure Landau theory without thermal activation:
//   E_c ≈ |dF/dP|_max along the switching path
//
// The inflection point P ≈ Ps/√3 gives the maximum barrier.
func (lp *LandauPotential) CoerciveField() float64 {
	Ps := lp.SpontaneousPolarization()
	if Ps == 0 {
		return 0
	}

	// For 2-4-6 potential, the inflection point is approximately at P = Ps/√3
	Pinflect := Ps / math.Sqrt(3)

	// The derivative at this point gives the coercive field
	return math.Abs(lp.Derivative(Pinflect))
}
