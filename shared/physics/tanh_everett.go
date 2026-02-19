package physics

import "math"

// TanhEverett implements EverettFunction using the product form of the
// Preisach Everett integral for sech^2-distributed hysterons.
//
// The Everett function F(alpha,beta) is the integral of the Preisach density
// mu(alpha,beta) over the triangle {(a,b) : beta <= b <= a <= alpha} in the
// Preisach half-plane. For the factorizable sech^2 density:
//
//	mu(a,b) = (Ps / 4*Delta^2) * sech^2((a - Ec)/Delta) * sech^2((b + Ec)/Delta)
//
// the exact integral yields the product form:
//
//	F(alpha,beta) = [1 + tanh((alpha - Ec)/Delta)] * [1 - tanh((beta + Ec)/Delta)] * Ps/4
//
// This product form is guaranteed non-negative for all (alpha,beta), unlike the
// older difference form [tanh(.) - tanh(.)] which can go negative for minor
// loops within the coercive gap (|alpha - beta| < 2*Ec). See MEMORY.md entry
// "Preisach Everett Zero-Clamp Teleportation" for the bug this form fixed.
//
// Units: Ps, Ec in SI (C/m^2, V/m). Delta controls distribution width (V/m).
type TanhEverett struct {
	Ps    float64 // Irreversible saturation polarization (C/m^2)
	Ec    float64 // Coercive field (V/m)
	Delta float64 // Distribution width (V/m); controls loop squareness
}

// Calculate returns the Everett integral F(alpha, beta) for the half-plane
// region defined by switching thresholds alpha (ascending) and beta (descending).
// The result is in C/m^2 and is guaranteed non-negative by the product form.
func (t *TanhEverett) Calculate(alpha, beta float64) float64 {
	ascCDF := 1.0 + math.Tanh((alpha-t.Ec)/t.Delta)
	descSurv := 1.0 - math.Tanh((beta+t.Ec)/t.Delta)

	val := ascCDF * descSurv * t.Ps * 0.25
	if val > t.Ps {
		return t.Ps
	}
	return val
}
