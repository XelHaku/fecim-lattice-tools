package utils

import (
	"math/rand"
	"testing"
)

func TestInitMatrix2D(t *testing.T) {
	m := InitMatrix2D(3, 4, 5)

	if len(m) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(m))
	}
	if len(m[0]) != 4 {
		t.Errorf("Expected 4 cols, got %d", len(m[0]))
	}

	for i := range m {
		for j := range m[i] {
			if m[i][j] != 5 {
				t.Errorf("Expected value 5 at [%d][%d], got %d", i, j, m[i][j])
			}
		}
	}
}

func TestInitMatrix2DInvalidSize(t *testing.T) {
	if InitMatrix2D(0, 4, 5) != nil {
		t.Error("Should return nil for 0 rows")
	}
	if InitMatrix2D(3, 0, 5) != nil {
		t.Error("Should return nil for 0 cols")
	}
	if InitMatrix2D(-1, 4, 5) != nil {
		t.Error("Should return nil for negative rows")
	}
}

func TestInitMatrix2DFloat64(t *testing.T) {
	m := InitMatrix2DFloat64(2, 3, 1.5)

	if len(m) != 2 || len(m[0]) != 3 {
		t.Error("Wrong dimensions")
	}

	for i := range m {
		for j := range m[i] {
			if m[i][j] != 1.5 {
				t.Errorf("Expected 1.5 at [%d][%d], got %f", i, j, m[i][j])
			}
		}
	}
}

func TestInitMatrixRandom(t *testing.T) {
	rand.Seed(42)
	m := InitMatrixRandom(10, 10, 30)

	if len(m) != 10 || len(m[0]) != 10 {
		t.Error("Wrong dimensions")
	}

	// Check all values are in range [0, 30)
	for i := range m {
		for j := range m[i] {
			if m[i][j] < 0 || m[i][j] >= 30 {
				t.Errorf("Value %d at [%d][%d] out of range [0, 30)", m[i][j], i, j)
			}
		}
	}
}

func TestInitMatrixRandomFloat64(t *testing.T) {
	rand.Seed(42)
	m := InitMatrixRandomFloat64(5, 5)

	for i := range m {
		for j := range m[i] {
			if m[i][j] < 0 || m[i][j] >= 1 {
				t.Errorf("Value %f at [%d][%d] out of range [0, 1)", m[i][j], i, j)
			}
		}
	}
}

func TestInitMatrixRandomRange(t *testing.T) {
	rand.Seed(42)
	m := InitMatrixRandomRange(5, 5, -1.0, 1.0)

	for i := range m {
		for j := range m[i] {
			if m[i][j] < -1.0 || m[i][j] >= 1.0 {
				t.Errorf("Value %f at [%d][%d] out of range [-1, 1)", m[i][j], i, j)
			}
		}
	}
}

func TestInitMatrixFunc(t *testing.T) {
	// Create identity-like matrix (1 on diagonal, 0 elsewhere)
	m := InitMatrixFunc(3, 3, func(row, col int) int {
		if row == col {
			return 1
		}
		return 0
	})

	expected := [][]int{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	for i := range m {
		for j := range m[i] {
			if m[i][j] != expected[i][j] {
				t.Errorf("Expected %d at [%d][%d], got %d", expected[i][j], i, j, m[i][j])
			}
		}
	}
}

func TestInitMatrixFuncFloat64(t *testing.T) {
	// Create matrix where each element is row + col
	m := InitMatrixFuncFloat64(2, 3, func(row, col int) float64 {
		return float64(row + col)
	})

	if m[0][0] != 0 || m[0][2] != 2 || m[1][1] != 2 || m[1][2] != 3 {
		t.Error("Matrix values not computed correctly")
	}
}

func TestInitMatrixFuncNilFn(t *testing.T) {
	if InitMatrixFunc(3, 3, nil) != nil {
		t.Error("Should return nil for nil function")
	}
	if InitMatrixFuncFloat64(3, 3, nil) != nil {
		t.Error("Should return nil for nil function")
	}
}

func TestInitVector(t *testing.T) {
	v := InitVector(5, 42)

	if len(v) != 5 {
		t.Errorf("Expected length 5, got %d", len(v))
	}

	for i, val := range v {
		if val != 42 {
			t.Errorf("Expected 42 at index %d, got %d", i, val)
		}
	}
}

func TestInitVectorFloat64(t *testing.T) {
	v := InitVectorFloat64(3, 3.14)

	if len(v) != 3 {
		t.Errorf("Expected length 3, got %d", len(v))
	}

	for i, val := range v {
		if val != 3.14 {
			t.Errorf("Expected 3.14 at index %d, got %f", i, val)
		}
	}
}

func TestInitVectorFunc(t *testing.T) {
	v := InitVectorFunc(5, func(i int) int {
		return i * i
	})

	expected := []int{0, 1, 4, 9, 16}
	for i, val := range v {
		if val != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, val)
		}
	}
}

func TestInitVectorFuncFloat64(t *testing.T) {
	v := InitVectorFuncFloat64(3, func(i int) float64 {
		return float64(i) * 0.5
	})

	expected := []float64{0, 0.5, 1.0}
	for i, val := range v {
		if val != expected[i] {
			t.Errorf("Expected %f at index %d, got %f", expected[i], i, val)
		}
	}
}

func TestInitVectorInvalidSize(t *testing.T) {
	if InitVector(0, 5) != nil {
		t.Error("Should return nil for size 0")
	}
	if InitVectorFloat64(-1, 5) != nil {
		t.Error("Should return nil for negative size")
	}
	if InitVectorFunc(0, func(i int) int { return i }) != nil {
		t.Error("Should return nil for size 0")
	}
}

func TestCopyMatrix2D(t *testing.T) {
	src := [][]int{
		{1, 2, 3},
		{4, 5, 6},
	}

	dst := CopyMatrix2D(src)

	// Verify copy
	for i := range src {
		for j := range src[i] {
			if dst[i][j] != src[i][j] {
				t.Errorf("Copy mismatch at [%d][%d]", i, j)
			}
		}
	}

	// Verify deep copy (modifying dst shouldn't affect src)
	dst[0][0] = 999
	if src[0][0] == 999 {
		t.Error("Copy is not deep - modifying dst affected src")
	}
}

func TestCopyMatrix2DFloat64(t *testing.T) {
	src := [][]float64{
		{1.1, 2.2},
		{3.3, 4.4},
	}

	dst := CopyMatrix2DFloat64(src)

	// Verify copy
	for i := range src {
		for j := range src[i] {
			if dst[i][j] != src[i][j] {
				t.Errorf("Copy mismatch at [%d][%d]", i, j)
			}
		}
	}

	// Verify deep copy
	dst[0][0] = 999.9
	if src[0][0] == 999.9 {
		t.Error("Copy is not deep")
	}
}

func TestCopyMatrix2DNil(t *testing.T) {
	if CopyMatrix2D(nil) != nil {
		t.Error("Should return nil for nil input")
	}
	if CopyMatrix2DFloat64(nil) != nil {
		t.Error("Should return nil for nil input")
	}
}
