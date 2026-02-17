package crossbar

import "math"

// TemperatureProfile controls which temperature-dependent effects are applied.
//
// Units:
//   - AmbientK: Kelvin
//
// Notes:
//   - Wire resistance temperature scaling is handled by IR-drop code paths.
//   - Additional effects (conductance window, noise, drift) are gated behind this
//     profile so existing simulations remain unchanged unless explicitly enabled.
//   - When nil, callers get legacy behavior (wire resistance only).
//
// M2-P2 completed: this profile enables temperature scalings beyond wire
// resistance while keeping legacy behavior available when profile=nil.
//
// Rationale: a profile object keeps MVMOptions reasonably small while allowing
// coherent defaults and future extensions (e.g., per-line thermal gradients).

type TemperatureProfile struct {
	// Enable turns on additional temperature-dependent effects beyond wire resistance.
	Enable bool

	// ApplyConductanceWindow scales the effective conductance window (Gmin/Gmax)
	// as a function of temperature.
	ApplyConductanceWindow bool

	// ApplyVariationNoise scales device/process variation magnitude with
	// temperature (thermal noise ~ sqrt(T)).
	ApplyVariationNoise bool

	// ApplyDrift applies a simplified drift-with-temperature model during MVM.
	// Drift is usually simulated separately over time; this is an optional fast
	// approximation for read-time drift impact.
	ApplyDrift bool
}

// DefaultTemperatureProfile returns a conservative profile that enables all
// additional temperature-dependent effects.
func DefaultTemperatureProfile() *TemperatureProfile {
	return &TemperatureProfile{
		Enable:                 true,
		ApplyConductanceWindow: true,
		ApplyVariationNoise:    true,
		ApplyDrift:             true,
	}
}

// SpatialTemperatureProfile models a 2D spatial temperature distribution across
// an array.
//
// This is intentionally separate from TemperatureProfile (which gates which
// temperature-dependent *effects* are enabled during MVM). The spatial profile
// is used for validation and future per-cell thermal effects.
//
// Temperatures are in Kelvin.
//
// Phase 5 tests (M2-TMP-02) validate hotspot and linear-gradient profiles.
// These profiles are not yet wired into MVM per-cell conductance; they are a
// building block for future work.

type SpatialTemperatureProfile struct {
	Rows     int
	Cols     int
	AmbientK float64
	MapK     [][]float64
}

// TemperatureAt returns the temperature at (row,col). Out-of-range indexes
// return AmbientK.
func (p *SpatialTemperatureProfile) TemperatureAt(row, col int) float64 {
	if p == nil {
		return 300
	}
	if row < 0 || col < 0 || row >= p.Rows || col >= p.Cols {
		return p.AmbientK
	}
	return p.MapK[row][col]
}

// NewHotspotTemperatureProfile creates a simple symmetric hotspot profile where
// the center cell is hottest and temperature decays toward the edges.
//
// deltaK is the peak temperature rise above ambient at the center.
func NewHotspotTemperatureProfile(rows, cols int, ambientK, deltaK float64) *SpatialTemperatureProfile {
	if rows <= 0 {
		rows = 1
	}
	if cols <= 0 {
		cols = 1
	}
	if ambientK <= 0 {
		ambientK = 300
	}
	if deltaK < 0 {
		deltaK = 0
	}

	m := make([][]float64, rows)
	centerR := float64(rows-1) / 2.0
	centerC := float64(cols-1) / 2.0
	maxDist := centerR
	if centerC > maxDist {
		maxDist = centerC
	}
	if maxDist <= 0 {
		maxDist = 1
	}

	for r := 0; r < rows; r++ {
		m[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			dr := float64(r) - centerR
			dc := float64(c) - centerC
			// Normalize distance to [0,1] using Chebyshev distance for grid symmetry.
			d := math.Max(math.Abs(dr), math.Abs(dc)) / maxDist
			if d > 1 {
				d = 1
			}
			m[r][c] = ambientK + deltaK*(1.0-d)
		}
	}

	return &SpatialTemperatureProfile{Rows: rows, Cols: cols, AmbientK: ambientK, MapK: m}
}

// NewGradientTemperatureProfile creates a linear temperature gradient along the
// chosen axis.
//
// axis: "x" (columns) or "y" (rows). If axis is unknown, defaults to "x".
func NewGradientTemperatureProfile(rows, cols int, startK, endK float64, axis string) *SpatialTemperatureProfile {
	if rows <= 0 {
		rows = 1
	}
	if cols <= 0 {
		cols = 1
	}
	if startK <= 0 {
		startK = 300
	}
	if endK <= 0 {
		endK = startK
	}

	m := make([][]float64, rows)

	denX := float64(cols - 1)
	if denX <= 0 {
		denX = 1
	}
	denY := float64(rows - 1)
	if denY <= 0 {
		denY = 1
	}

	for r := 0; r < rows; r++ {
		m[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			var frac float64
			switch axis {
			case "y", "Y":
				frac = float64(r) / denY
			default:
				frac = float64(c) / denX
			}
			m[r][c] = startK + frac*(endK-startK)
		}
	}

	// AmbientK is defined as the start point for reference.
	return &SpatialTemperatureProfile{Rows: rows, Cols: cols, AmbientK: startK, MapK: m}
}
