// Package arraysim provides approximate array coupling solvers for module4-circuits.
package arraysim

import sharedphysics "fecim-lattice-tools/shared/physics"

// CouplingMode selects the array coupling model.
type CouplingMode int

const (
	// CouplingIdeal disables coupling; assumes ideal voltages.
	CouplingIdeal CouplingMode = iota
	// CouplingTierA enables the Tier A approximate IR-drop / half-select model.
	CouplingTierA
	// CouplingTierB enables the Tier B DC nodal (MNA/KCL) model.
	//
	// Tier B is not selected by default; callers must explicitly opt in.
	CouplingTierB
)

// CellGeometry captures array-relevant geometry (SI units).
//
// Core ferroelectric film dimensions (thickness/area) are centralized in
// shared/physics.CellGeometry and referenced here to avoid duplication.
type CellGeometry struct {
	PitchX           float64                    // Cell pitch in X (m)
	PitchY           float64                    // Cell pitch in Y (m)
	Film             sharedphysics.CellGeometry // Ferroelectric film geometry (thickness/area)
	WireWidth        float64                    // Word/bitline width (m)
	WireThickness    float64                    // Word/bitline thickness (m)
	MetalResistivity float64                    // Conductor resistivity (ohm*m)
}

// DefaultCellGeometry returns conservative defaults aligned with module6 layout assumptions.
func DefaultCellGeometry() CellGeometry {
	return CellGeometry{
		PitchX:           0.46e-6, // 0.46 um (SKY130 pitch)
		PitchY:           2.72e-6, // 2.72 um (SKY130 row height)
		Film:             sharedphysics.DefaultCellGeometry(),
		WireWidth:        0.08e-6, // 80 nm line width (tighter default)
		WireThickness:    0.16e-6, // 160 nm metal thickness (tighter default)
		MetalResistivity: 2.2e-8,  // Copper-like resistivity
	}
}

// WithDefaults fills missing geometry fields with defaults.
func (g CellGeometry) WithDefaults() CellGeometry {
	defaults := DefaultCellGeometry()
	if g.PitchX <= 0 {
		g.PitchX = defaults.PitchX
	}
	if g.PitchY <= 0 {
		g.PitchY = defaults.PitchY
	}
	g.Film = g.Film.WithDefaults()
	if g.WireWidth <= 0 {
		g.WireWidth = defaults.WireWidth
	}
	if g.WireThickness <= 0 {
		g.WireThickness = defaults.WireThickness
	}
	if g.MetalResistivity <= 0 {
		g.MetalResistivity = defaults.MetalResistivity
	}
	return g
}

// WireParams defines effective per-segment resistance for array wires (ohms per cell).
type WireParams struct {
	RWordLine float64 // Wordline resistance per cell pitch (ohm)
	RBitLine  float64 // Bitline resistance per cell pitch (ohm)
}

// WithDefaults fills missing wire parameters using geometry-derived estimates.
func (w WireParams) WithDefaults(geom CellGeometry) WireParams {
	geom = geom.WithDefaults()
	if w.RWordLine <= 0 || w.RBitLine <= 0 {
		crossSection := geom.WireWidth * geom.WireThickness
		if crossSection <= 0 {
			def := DefaultCellGeometry()
			crossSection = def.WireWidth * def.WireThickness
		}
		resistPerMeter := geom.MetalResistivity / crossSection
		if w.RWordLine <= 0 {
			w.RWordLine = resistPerMeter * geom.PitchX
		}
		if w.RBitLine <= 0 {
			w.RBitLine = resistPerMeter * geom.PitchY
		}
	}
	return w
}

// SolveParams collects inputs for a coupling solve.
type SolveParams struct {
	WLVoltages  []float64   // Row (WL) driver voltages (V)
	BLVoltages  []float64   // Column (BL) driver voltages (V)
	Conductance [][]float64 // Cell conductances (S)
	ActiveRows  []bool      // Optional row enable mask (nil = all active)
	Geometry    CellGeometry
	Wire        WireParams
}

// SolveResult captures Tier A solver outputs.
type SolveResult struct {
	CellVoltages [][]float64 // Vcell per cell (V)
	CellCurrents [][]float64 // Icell per cell (A)
	RowCurrents  []float64   // Summed row currents (A)
	ColCurrents  []float64   // Summed column currents (A)
}

// Engine defines a coupling solver that can be swapped for DC nodal/MNA later.
type Engine interface {
	Solve(params SolveParams) (SolveResult, error)
}
