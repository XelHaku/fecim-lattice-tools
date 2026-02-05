package physics

// CellGeometry captures the physical dimensions of a single ferroelectric cell (SI units).
//
// This type is the single source of truth for geometry used in field/charge conversions.
// Higher-level modules (e.g. array wire models) may extend this with layout-specific
// parameters, but should not duplicate these core fields.
//
// Conventions:
//   - Thickness is the ferroelectric film thickness (m)
//   - Area is the electrically-active ferroelectric area (m^2)
//
// Note: Pitch/metal geometry live outside this type because they are layout/technology
// dependent rather than intrinsic to the ferroelectric film.
type CellGeometry struct {
	Thickness float64 // Ferroelectric thickness (m)
	Area      float64 // Active area (m^2)
}

// DefaultCellGeometry returns conservative defaults aligned with FeCIMMaterial.
func DefaultCellGeometry() CellGeometry {
	m := FeCIMMaterial()
	return CellGeometry{Thickness: m.Thickness, Area: m.Area}
}

// WithDefaults fills missing geometry fields with defaults.
func (g CellGeometry) WithDefaults() CellGeometry {
	def := DefaultCellGeometry()
	if g.Thickness <= 0 {
		g.Thickness = def.Thickness
	}
	if g.Area <= 0 {
		g.Area = def.Area
	}
	return g
}

// ElectricField converts an applied voltage to electric field using E = V / t.
func (g CellGeometry) ElectricField(voltageV float64) float64 {
	if g.Thickness <= 0 {
		return 0
	}
	return voltageV / g.Thickness
}

// ChargeFromPolarization converts polarization (C/m^2) to total charge (C) via Q = P*A.
func (g CellGeometry) ChargeFromPolarization(polarizationCPerM2 float64) float64 {
	if g.Area <= 0 {
		return 0
	}
	return polarizationCPerM2 * g.Area
}

// CurrentFromDPdt converts dP/dt (C/m^2/s) to total displacement current (A) via I = A*dP/dt.
func (g CellGeometry) CurrentFromDPdt(dPdtCPerM2PerS float64) float64 {
	if g.Area <= 0 {
		return 0
	}
	return dPdtCPerM2PerS * g.Area
}
