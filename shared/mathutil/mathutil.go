// Package mathutil provides shared math utility functions for FeCIM tools.
package mathutil

import "fmt"

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

// LerpByIndex linearly interpolates index in [0,count-1] onto [min,max].
func LerpByIndex(index, count int, min, max float64) float64 {
	if count <= 1 {
		return min
	}
	return min + (max-min)*float64(index)/float64(count-1)
}

// Dot returns the dot product of equal-length vectors.
func Dot(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("dot product length mismatch: %d != %d", len(a), len(b))
	}
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum, nil
}

// MatrixVectorMultiply returns matrix×vector for a dense row-major matrix.
func MatrixVectorMultiply(matrix [][]float64, vector []float64) ([]float64, error) {
	out := make([]float64, len(matrix))
	for row, values := range matrix {
		if len(values) != len(vector) {
			return nil, fmt.Errorf("matrix row %d length mismatch: %d != %d", row, len(values), len(vector))
		}
		sum, err := Dot(values, vector)
		if err != nil {
			return nil, err
		}
		out[row] = sum
	}
	return out, nil
}
