// Package arraysim provides approximate array coupling solvers for module4-circuits.
package arraysim

// CouplingMode selects the array coupling model.
type CouplingMode int

const (
	// CouplingIdeal disables coupling; assumes ideal voltages.
	CouplingIdeal CouplingMode = iota
	// CouplingTierA enables the Tier A approximate IR-drop / half-select model.
	CouplingTierA
)

// CellGeometry captures physical cell dimensions and wire geometry (SI units).
type CellGeometry struct {
	PitchX           float64 // Cell pitch in X (m)
	PitchY           float64 // Cell pitch in Y (m)
	Thickness        float64 // Ferroelectric thickness (m)
	ActiveArea       float64 // Active cell area (m^2)
	WireWidth        float64 // Word/bitline width (m)
	WireThickness    float64 // Word/bitline thickness (m)
	MetalResistivity float64 // Conductor resistivity (ohm*m)
}

// DefaultCellGeometry returns conservative defaults aligned with module6 layout assumptions.
func DefaultCellGeometry() CellGeometry {
	return CellGeometry{
		PitchX:           0.46e-6, // 0.46 um (SKY130 pitch)
		PitchY:           2.72e-6, // 2.72 um (SKY130 row height)
		Thickness:        10e-9,   // 10 nm HZO film
		ActiveArea:       1e-14,   // 0.01 um^2 active area
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
	if g.Thickness <= 0 {
		g.Thickness = defaults.Thickness
	}
	if g.ActiveArea <= 0 {
		g.ActiveArea = defaults.ActiveArea
	}
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
			crossSection = DefaultCellGeometry().WireWidth * DefaultCellGeometry().WireThickness
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
