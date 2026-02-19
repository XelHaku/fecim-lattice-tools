// Package physics provides shared physics utilities for FeCIM simulations.
package physics

import (
	"math"
	"strings"
)

// Conductance range constants (physical units in Siemens)
const (
	// GMin is the minimum conductance (OFF state) - 10 µS
	GMin = 10e-6

	// GMax is the maximum conductance (ON state) - 100 µS
	GMax = 100e-6

	// GRatio is the ON/OFF conductance ratio (10:1)
	GRatio = GMax / GMin
)

// ConductanceModel specifies the G(V) relationship model for FeFET devices.
type ConductanceModel int

const (
	// ConductanceLinear uses linear interpolation: G = Gmin + gNorm*(Gmax-Gmin)
	// Simple model, less accurate at extremes.
	ConductanceLinear ConductanceModel = iota

	// ConductanceExponential uses exponential scaling: G = Gmin * exp(ln(Gmax/Gmin) * gNorm)
	// Legacy name retained for compatibility; equivalent to ConductanceSubthreshold.
	ConductanceExponential

	// ConductanceLookup uses a pre-defined calibration table.
	// Most accurate when calibration data is available.
	ConductanceLookup

	// ConductanceSaturation approximates above-threshold IDS ∝ (VGS - VT)^2 behavior.
	ConductanceSaturation
)

// ConductanceSubthreshold is an alias for the exponential model.
const ConductanceSubthreshold = ConductanceExponential

// String returns the string representation of the conductance model.
func (m ConductanceModel) String() string {
	switch m {
	case ConductanceLinear:
		return "Linear (Educational/Simplified)"
	case ConductanceExponential:
		return "Exponential (Recommended)"
	case ConductanceLookup:
		return "Lookup (Calibrated)"
	case ConductanceSaturation:
		return "Saturation"
	default:
		return "Unknown"
	}
}

// ParseConductanceModel parses config strings with exponential default.
// Default changed from Linear to Exponential per 2024-2025 literature consensus
// (Nature Comms 2018: linear model fundamentally wrong for ferroelectrics).
func ParseConductanceModel(model string) ConductanceModel {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "linear":
		return ConductanceLinear // Educational/Simplified mode
	case "saturation":
		return ConductanceSaturation
	case "lookup":
		return ConductanceLookup
	case "subthreshold", "exponential":
		fallthrough
	default:
		return ConductanceExponential // Default: physics-accurate model
	}
}

// NormalizedToPhysical converts normalized conductance [0,1] to physical units (Siemens).
// Uses the specified model type for interpolation.
//
// Level 0  → GMin (10 µS)
// Level 29 → GMax (100 µS)
//
// For exponential model, midpoint (level 15) = geometric mean ≈ 31.6 µS
func NormalizedToPhysical(gNorm float64, model ConductanceModel) float64 {
	return NormalizedToPhysicalRange(gNorm, model, GMin, GMax, nil)
}

// NormalizedToPhysicalRange converts normalized conductance to physical units
// with custom min/max range and optional lookup table.
func NormalizedToPhysicalRange(gNorm float64, model ConductanceModel, gMin, gMax float64, lookupTable []float64) float64 {
	// Clamp to valid range
	if gNorm < 0 {
		gNorm = 0
	}
	if gNorm > 1 {
		gNorm = 1
	}

	switch model {
	case ConductanceExponential:
		// G = Gmin * exp(ln(Gmax/Gmin) * gNorm)
		// This gives exponential scaling where:
		//   gNorm=0   → Gmin
		//   gNorm=0.5 → sqrt(Gmin*Gmax) = geometric mean
		//   gNorm=1   → Gmax
		ratio := gMax / gMin
		return gMin * math.Exp(math.Log(ratio)*gNorm)

	case ConductanceLookup:
		// Use calibration table if available
		if len(lookupTable) >= 2 {
			level := int(math.Round(gNorm * float64(len(lookupTable)-1)))
			if level < 0 {
				level = 0
			}
			if level >= len(lookupTable) {
				level = len(lookupTable) - 1
			}
			return lookupTable[level]
		}
		// Fall back to linear if no table
		fallthrough

	case ConductanceLinear:
		fallthrough
	default:
		// Linear interpolation
		return gMin + gNorm*(gMax-gMin)
	}
}

// PhysicalToNormalized converts physical conductance (Siemens) to normalized [0,1].
// Uses linear mapping by default.
func PhysicalToNormalized(gPhys float64) float64 {
	return PhysicalToNormalizedRange(gPhys, GMin, GMax)
}

// PhysicalToNormalizedRange converts physical conductance to normalized with custom range.
func PhysicalToNormalizedRange(gPhys, gMin, gMax float64) float64 {
	if gPhys <= gMin {
		return 0.0
	}
	if gPhys >= gMax {
		return 1.0
	}
	return (gPhys - gMin) / (gMax - gMin)
}

// PhysicalToNormalizedModel converts physical conductance to normalized [0,1]
// using the specified conductance model as the inverse function.
//
// For the exponential model: gNorm = ln(G/Gmin) / ln(Gmax/Gmin)
// This is the exact inverse of NormalizedToPhysicalRange for ConductanceExponential.
//
// For linear and other models: gNorm = (G-Gmin) / (Gmax-Gmin)
func PhysicalToNormalizedModel(gPhys, gMin, gMax float64, model ConductanceModel) float64 {
	if gPhys <= gMin {
		return 0.0
	}
	if gPhys >= gMax {
		return 1.0
	}
	switch model {
	case ConductanceExponential:
		if gMin <= 0 {
			// Positive Gmin required for log-space; fall back to linear.
			return (gPhys - gMin) / (gMax - gMin)
		}
		logRatio := math.Log(gMax / gMin)
		if logRatio < 1e-12 {
			return 0.5
		}
		return math.Log(gPhys/gMin) / logRatio
	default:
		return (gPhys - gMin) / (gMax - gMin)
	}
}

// ConductanceToLevel converts physical conductance to discrete level (0-29).
func ConductanceToLevel(gPhys float64, levels int) int {
	gNorm := PhysicalToNormalized(gPhys)
	return GetLevel(gNorm, levels)
}

// LevelToConductance converts discrete level to physical conductance.
func LevelToConductance(level, levels int, model ConductanceModel) float64 {
	gNorm := NormalizeFromLevel(level, levels)
	return NormalizedToPhysical(gNorm, model)
}

// ConductanceWindow returns the total conductance window (GMax - GMin).
func ConductanceWindow() float64 {
	return GMax - GMin
}

// ConductanceMidpoint returns the arithmetic midpoint conductance.
func ConductanceMidpoint() float64 {
	return (GMax + GMin) / 2
}

// ConductanceGeometricMean returns the geometric mean (for exponential model midpoint).
func ConductanceGeometricMean() float64 {
	return math.Sqrt(GMin * GMax)
}
