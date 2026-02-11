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
	Stack     string  // Optional process/material stack label (e.g. "TiN/HZO/TiN")
}

// DefaultCellGeometry returns conservative defaults aligned with FeCIMMaterial.
func DefaultCellGeometry() CellGeometry {
	m := FeCIMMaterial()
	return CellGeometry{Thickness: m.Thickness, Area: m.Area, Stack: "TiN/HZO/TiN"}
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
	if g.Stack == "" {
		g.Stack = def.Stack
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

// ConductanceFromConductivity computes film conductance from conductivity using
// G = sigma * area / thickness.
func (g CellGeometry) ConductanceFromConductivity(conductivitySPerM float64) float64 {
	if conductivitySPerM <= 0 || g.Area <= 0 || g.Thickness <= 0 {
		return 0
	}
	return conductivitySPerM * g.Area / g.Thickness
}

// ResistanceFromResistivity computes film resistance from resistivity using
// R = rho * thickness / area.
func (g CellGeometry) ResistanceFromResistivity(resistivityOhmM float64) float64 {
	if resistivityOhmM <= 0 || g.Area <= 0 || g.Thickness <= 0 {
		return 0
	}
	return resistivityOhmM * g.Thickness / g.Area
}

// ConductanceScale returns the relative geometry scaling factor (A/t) / (Aref/tref).
// Returns 1 when either geometry is invalid to avoid exploding callers.
func (g CellGeometry) ConductanceScale(reference CellGeometry) float64 {
	if g.Area <= 0 || g.Thickness <= 0 || reference.Area <= 0 || reference.Thickness <= 0 {
		return 1
	}
	return (g.Area / g.Thickness) / (reference.Area / reference.Thickness)
}

// GeometryFromMaterial extracts canonical cell geometry from material parameters.
func GeometryFromMaterial(mat *HZOMaterial) CellGeometry {
	if mat == nil {
		return DefaultCellGeometry()
	}
	g := CellGeometry{Thickness: mat.Thickness, Area: mat.Area, Stack: mat.Name}
	return g.WithDefaults()
}
