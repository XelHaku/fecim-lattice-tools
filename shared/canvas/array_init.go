// Package utils provides common utility functions for FeCIM tools
package utils

import "math/rand"

// InitMatrix2D creates a 2D matrix of integers initialized to a constant value.
func InitMatrix2D(rows, cols int, value int) [][]int {
	if rows <= 0 || cols <= 0 {
		return nil
	}
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
		for j := range matrix[i] {
			matrix[i][j] = value
		}
	}
	return matrix
}

// InitMatrix2DFloat64 creates a 2D matrix of float64 initialized to a constant value.
func InitMatrix2DFloat64(rows, cols int, value float64) [][]float64 {
	if rows <= 0 || cols <= 0 {
		return nil
	}
	matrix := make([][]float64, rows)
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			matrix[i][j] = value
		}
	}
	return matrix
}

// InitMatrixRandom creates a 2D matrix of integers with random values in [0, max).
func InitMatrixRandom(rows, cols, max int) [][]int {
	if rows <= 0 || cols <= 0 || max <= 0 {
		return nil
	}
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
		for j := range matrix[i] {
			matrix[i][j] = rand.Intn(max)
		}
	}
	return matrix
}

// InitMatrixRandomFloat64 creates a 2D matrix of float64 with random values in [0, 1).
func InitMatrixRandomFloat64(rows, cols int) [][]float64 {
	if rows <= 0 || cols <= 0 {
		return nil
	}
	matrix := make([][]float64, rows)
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			matrix[i][j] = rand.Float64()
		}
	}
	return matrix
}

// InitMatrixRandomRange creates a 2D matrix of float64 with random values in [min, max).
func InitMatrixRandomRange(rows, cols int, min, max float64) [][]float64 {
	if rows <= 0 || cols <= 0 {
		return nil
	}
	matrix := make([][]float64, rows)
	scale := max - min
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			matrix[i][j] = min + rand.Float64()*scale
		}
	}
	return matrix
}

// InitMatrixFunc creates a 2D matrix where each element is computed by fn(row, col).
func InitMatrixFunc(rows, cols int, fn func(row, col int) int) [][]int {
	if rows <= 0 || cols <= 0 || fn == nil {
		return nil
	}
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
		for j := range matrix[i] {
			matrix[i][j] = fn(i, j)
		}
	}
	return matrix
}

// InitMatrixFuncFloat64 creates a 2D float64 matrix where each element is computed by fn(row, col).
func InitMatrixFuncFloat64(rows, cols int, fn func(row, col int) float64) [][]float64 {
	if rows <= 0 || cols <= 0 || fn == nil {
		return nil
	}
	matrix := make([][]float64, rows)
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			matrix[i][j] = fn(i, j)
		}
	}
	return matrix
}

// InitVector creates a 1D slice of integers initialized to a constant value.
func InitVector(size int, value int) []int {
	if size <= 0 {
		return nil
	}
	vec := make([]int, size)
	for i := range vec {
		vec[i] = value
	}
	return vec
}

// InitVectorFloat64 creates a 1D slice of float64 initialized to a constant value.
func InitVectorFloat64(size int, value float64) []float64 {
	if size <= 0 {
		return nil
	}
	vec := make([]float64, size)
	for i := range vec {
		vec[i] = value
	}
	return vec
}

// InitVectorFunc creates a 1D slice where each element is computed by fn(index).
func InitVectorFunc(size int, fn func(i int) int) []int {
	if size <= 0 || fn == nil {
		return nil
	}
	vec := make([]int, size)
	for i := range vec {
		vec[i] = fn(i)
	}
	return vec
}

// InitVectorFuncFloat64 creates a 1D float64 slice where each element is computed by fn(index).
func InitVectorFuncFloat64(size int, fn func(i int) float64) []float64 {
	if size <= 0 || fn == nil {
		return nil
	}
	vec := make([]float64, size)
	for i := range vec {
		vec[i] = fn(i)
	}
	return vec
}

// CopyMatrix2D creates a deep copy of a 2D integer matrix.
func CopyMatrix2D(src [][]int) [][]int {
	if src == nil {
		return nil
	}
	dst := make([][]int, len(src))
	for i := range src {
		dst[i] = make([]int, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

// CopyMatrix2DFloat64 creates a deep copy of a 2D float64 matrix.
func CopyMatrix2DFloat64(src [][]float64) [][]float64 {
	if src == nil {
		return nil
	}
	dst := make([][]float64, len(src))
	for i := range src {
		dst[i] = make([]float64, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}
