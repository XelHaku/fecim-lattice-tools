// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

// HZOMaterial contains material parameters for Hafnium-Zirconium Oxide.
// Values are based on literature measurements for HfO2-ZrO2 superlattices.
type HZOMaterial struct {
	// Polarization parameters (C/m²)
	Pr float64 // Remanent polarization
	Ps float64 // Saturation polarization

	// Field parameters (V/m)
	Ec float64 // Coercive field

	// Dielectric properties
	Epsilon float64 // Relative permittivity

	// Film properties
	Thickness float64 // Film thickness (m)

	// Dynamics
	Tau float64 // Switching time constant (s)
}

// DefaultHZO returns material parameters for typical HZO thin films.
// Based on literature values from CURRICULUM.md references.
func DefaultHZO() *HZOMaterial {
	return &HZOMaterial{
		Pr:        25e-2,  // 25 μC/cm² = 0.25 C/m²
		Ps:        30e-2,  // 30 μC/cm² = 0.30 C/m²
		Ec:        1.2e8,  // 1.2 MV/cm = 1.2e8 V/m
		Epsilon:   30,     // Relative permittivity
		Thickness: 10e-9,  // 10 nm
		Tau:       1e-9,   // 1 ns switching time
	}
}

// OptimizedHZO returns parameters for optimized superlattice HZO.
// Higher Pr, lower Ec due to superlattice strain engineering.
func OptimizedHZO() *HZOMaterial {
	return &HZOMaterial{
		Pr:        35e-2,  // 35 μC/cm² (superlattice enhanced)
		Ps:        40e-2,  // 40 μC/cm²
		Ec:        1.0e8,  // 1.0 MV/cm (reduced by FE-AFE competition)
		Epsilon:   35,     // Higher permittivity
		Thickness: 10e-9,  // 10 nm
		Tau:       0.5e-9, // 0.5 ns (faster switching)
	}
}

// CoerciveVoltage returns the coercive voltage for a given thickness.
func (m *HZOMaterial) CoerciveVoltage() float64 {
	return m.Ec * m.Thickness
}

// RemanentCharge returns the remanent charge per unit area.
func (m *HZOMaterial) RemanentCharge() float64 {
	return m.Pr
}
