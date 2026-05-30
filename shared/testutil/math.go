// Package testutil provides shared test helpers for FeCIM simulation tests.
//
// Test files across multiple packages independently define the same small
// utility functions (relative-error, absolute-value, mock constructors).
// This package consolidates those helpers to reduce duplication and ensure
// consistent error/boundary behavior across test suites.
package testutil

import "math"

// RelErr computes the relative error |got - want| / |want|.
// When want == 0, returns |got| as the absolute error.
func RelErr(got, want float64) float64 {
	if want == 0 {
		return math.Abs(got)
	}
	return math.Abs(got-want) / math.Abs(want)
}

// AbsFloat64 returns the absolute value of x.
func AbsFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// AbsInt returns the absolute value of x.
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Clamp01 restricts v to [0, 1].
func Clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
