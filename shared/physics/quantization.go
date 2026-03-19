// Package physics provides shared physics utilities for FeCIM simulations.
package physics

import (
	"math"

	"fecim-lattice-tools/shared/mathutil"
)

// FeCIM quantization constants
const (
	// DefaultLevels is the baseline number of discrete analog states used in this demo.
	// Conference claim (COSM 2025), pending peer review: "It's got 30 discrete states."
	DefaultLevels = 30

	// BitsPerCell is the effective bits per cell: log2(30) ≈ 4.91
	BitsPerCell = 4.91
)

// QuantizeToLevels quantizes a normalized value [0,1] to exactly N discrete levels.
// Returns a normalized value that maps exactly to one of the N levels.
//
// Example:
//
//	QuantizeToLevels(0.5, 30)  // Returns 0.5 (level 15)
//	QuantizeToLevels(0.33, 30) // Returns ~0.345 (level 10)
func QuantizeToLevels(value float64, levels int) float64 {
	value = mathutil.Clamp01(value)
	// Quantize to levels (0 to N-1)
	level := math.Round(value * float64(levels-1))
	return level / float64(levels-1)
}

// QuantizeTo30Levels quantizes a value to exactly 30 discrete FeCIM levels.
// This is the standard quantization for FeCIM devices.
//
// Example:
//
//	QuantizeTo30Levels(0.5)  // Returns 0.5 (level 15)
//	QuantizeTo30Levels(0.0)  // Returns 0.0 (level 0)
//	QuantizeTo30Levels(1.0)  // Returns 1.0 (level 29)
func QuantizeTo30Levels(value float64) float64 {
	return QuantizeToLevels(value, DefaultLevels)
}

// GetLevel returns the discrete level index (0 to N-1) for a normalized conductance value.
//
// Example:
//
//	GetLevel(0.0, 30)  // Returns 0
//	GetLevel(0.5, 30)  // Returns 15
//	GetLevel(1.0, 30)  // Returns 29
func GetLevel(conductance float64, levels int) int {
	return int(math.Round(conductance * float64(levels-1)))
}

// GetLevelFor30 returns the discrete level (0-29) for a normalized conductance value.
// This is the standard function for FeCIM devices with 30 levels.
func GetLevelFor30(conductance float64) int {
	return GetLevel(conductance, DefaultLevels)
}

// NormalizeFromLevel converts a discrete level index to normalized value [0,1].
//
// Example:
//
//	NormalizeFromLevel(0, 30)  // Returns 0.0
//	NormalizeFromLevel(15, 30) // Returns ~0.517
//	NormalizeFromLevel(29, 30) // Returns 1.0
func NormalizeFromLevel(level, levels int) float64 {
	if level < 0 {
		level = 0
	}
	if level >= levels {
		level = levels - 1
	}
	return float64(level) / float64(levels-1)
}

// NormalizeFromLevel30 converts a level (0-29) to normalized value for FeCIM.
func NormalizeFromLevel30(level int) float64 {
	return NormalizeFromLevel(level, DefaultLevels)
}

// LevelSpacing returns the normalized spacing between adjacent levels.
//
// Example:
//
//	LevelSpacing(30) // Returns ~0.0345 (1/29)
func LevelSpacing(levels int) float64 {
	if levels <= 1 {
		return 1.0
	}
	return 1.0 / float64(levels-1)
}

// QuantizationError returns the maximum quantization error for N levels.
// This is half the level spacing (worst case: value falls exactly between levels).
func QuantizationError(levels int) float64 {
	return LevelSpacing(levels) / 2.0
}
