package crossbar

import "math/rand"

// makeUniformG creates an NxM conductance matrix with uniform value g.
func makeUniformG(rows, cols int, g float64) [][]float64 {
	m := make([][]float64, rows)
	for i := range m {
		m[i] = make([]float64, cols)
		for j := range m[i] {
			m[i][j] = g
		}
	}
	return m
}

// makeRandomG creates an NxM conductance matrix with random values in [lo, hi].
func makeRandomG(rows, cols int, lo, hi float64, rng *rand.Rand) [][]float64 {
	m := make([][]float64, rows)
	for i := range m {
		m[i] = make([]float64, cols)
		for j := range m[i] {
			m[i][j] = lo + (hi-lo)*rng.Float64()
		}
	}
	return m
}

// makeApplied creates a uniform applied voltage vector of length n.
func makeApplied(n int, v float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = v
	}
	return out
}
