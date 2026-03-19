// Package mathutil provides shared math utility functions for FeCIM tools.
package mathutil

// Clamp restricts v to the range [min, max].
func Clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Clamp01 restricts v to the range [0, 1].
func Clamp01(v float64) float64 {
	return Clamp(v, 0, 1)
}

// ClampInt restricts v to the range [min, max].
func ClampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ClampPositive returns max(0, v).
func ClampPositive(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

// AbsInt returns the absolute value of an integer.
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
