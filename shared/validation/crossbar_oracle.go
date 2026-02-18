// Package validation provides oracle implementations for validating crossbar simulations.
//
// This file contains a pure-Go exact nodal-analysis (KCL) solver for passive
// crossbar arrays, inspired by badcrossbar's exact circuit solver (UCL, MIT licence).
// It serves as a ground-truth reference for validating the behavioural MVM
// approximation in shared/crossbar.
package validation

import (
	"fmt"
	"math"
)

// CrossbarResult holds every output quantity from the exact KCL solver.
type CrossbarResult struct {
	// RowCurrents is the net current absorbed by the virtual-ground sense
	// node at the LEFT end of each row (A). This is the "output" of an MVM.
	RowCurrents []float64

	// CellCurrents[i][j] is the current flowing from column j to row i through
	// the cell conductance G[i][j] (A).  Positive = flowing into the row wire.
	CellCurrents [][]float64

	// ColNodeV[j][i] is the voltage on column wire j at the junction with row i.
	ColNodeV [][]float64

	// RowNodeV[i][j] is the voltage on row wire i at the junction with column j.
	RowNodeV [][]float64
}

// ExactMVM solves the passive crossbar MVM exactly using KCL / nodal analysis.
//
// Electrical model (matches badcrossbar's single-ended drive convention):
//
//   - N rows, M columns
//   - Column j is driven at its TOP by colVoltages[j].
//   - Row i is GROUNDED at its LEFT end (virtual-ground current sensing).
//   - Wire resistance wireR (Ω) per cell-pitch along both row and column wires.
//     Set wireR = 0 for ideal wires; the exact result then equals Σ G[i][j]·V[j].
//   - Bottom of each column and right end of each row are open (no connection).
//
// Node layout (N=2 rows, M=2 cols illustrated):
//
//	    V_in[0]        V_in[1]
//	      │              │
//	  col[0][0]      col[1][0]    ← row 0 junction nodes on columns
//	   G[0][0]         G[0][1]    ← cells connecting cols to rows
//	─── row[0][0] ─── row[0][1]  ← row 0 wire  (row[0][0]=0V grounded)
//	      │              │
//	  col[0][1]      col[1][1]    ← row 1 junction nodes on columns
//	   G[1][0]         G[1][1]
//	─── row[1][0] ─── row[1][1]  ← row 1 wire  (row[1][0]=0V grounded)
//
// The result RowCurrents[i] = total current consumed by the sensing resistor on
// row i, which equals Σ_j G[i][j]·V_col_j for wireR=0.
func ExactMVM(conductances [][]float64, colVoltages []float64, wireR float64) (*CrossbarResult, error) {
	N := len(conductances)
	if N == 0 {
		return nil, fmt.Errorf("crossbar_oracle: empty conductances")
	}
	M := len(conductances[0])
	if M == 0 || len(colVoltages) != M {
		return nil, fmt.Errorf("crossbar_oracle: dimension mismatch (rows=%d, cols=%d, inputs=%d)", N, M, len(colVoltages))
	}

	// ── Shortcut for ideal wires ────────────────────────────────────────────
	// When wireR == 0 every column is at its driven voltage everywhere and
	// every row is at 0 V everywhere.  The exact output is trivially:
	//   RowCurrents[i] = Σ_j G[i][j] · colVoltages[j]
	if wireR == 0 {
		return idealMVM(conductances, colVoltages, N, M), nil
	}

	// ── General case: nodal analysis via Gaussian elimination ───────────────
	//
	// Free (unknown) nodes:
	//   col_node[j][i]  j=0..M-1, i=1..N-1    → M·(N-1) unknowns
	//   row_node[i][j]  i=0..N-1, j=1..M-1    → N·(M-1) unknowns
	//
	// Index mapping:
	//   col_node[j][i] → idx = j*(N-1) + (i-1)
	//   row_node[i][j] → idx = M*(N-1) + i*(M-1) + (j-1)

	nCol := M * (N - 1) // number of unknown column-wire nodes
	nRow := N * (M - 1) // number of unknown row-wire nodes
	nUnknowns := nCol + nRow

	// Build the system matrix A and RHS vector b: A·x = b
	A := make([][]float64, nUnknowns)
	b := make([]float64, nUnknowns)
	for i := range A {
		A[i] = make([]float64, nUnknowns)
	}

	colIdx := func(j, i int) int { return j*(N-1) + (i - 1) }
	rowIdx := func(i, j int) int { return nCol + i*(M-1) + (j - 1) }
	gWire := 1.0 / wireR // wire conductance per segment

	// ── KCL at each unknown col_node[j][i] (i=1..N-1) ─────────────────────
	for j := 0; j < M; j++ {
		for i := 1; i < N; i++ {
			eq := colIdx(j, i)

			// Wire above (to col_node[j][i-1]):
			if i == 1 {
				// col_node[j][0] = colVoltages[j]  → known, move to RHS
				A[eq][eq] += gWire
				b[eq] += gWire * colVoltages[j]
			} else {
				A[eq][eq] += gWire
				A[eq][colIdx(j, i-1)] -= gWire
			}

			// Wire below (to col_node[j][i+1], only if not the bottom-most node):
			if i < N-1 {
				A[eq][eq] += gWire
				A[eq][colIdx(j, i+1)] -= gWire
			}
			// (If i == N-1 the bottom end is open — no wire conductance term.)

			// Cell (i,j) to row_node[i][j]:
			G := conductances[i][j]
			A[eq][eq] += G
			if j == 0 {
				// row_node[i][0] = 0V  → known; contributes G·0 = 0 to RHS
				// (nothing to add)
			} else {
				A[eq][rowIdx(i, j)] -= G
			}
		}
	}

	// ── KCL at each unknown row_node[i][j] (j=1..M-1) ─────────────────────
	for i := 0; i < N; i++ {
		for j := 1; j < M; j++ {
			eq := rowIdx(i, j)

			// Wire left (to row_node[i][j-1]):
			if j == 1 {
				// row_node[i][0] = 0V  → known
				A[eq][eq] += gWire
				// b[eq] += gWire * 0  (no-op)
			} else {
				A[eq][eq] += gWire
				A[eq][rowIdx(i, j-1)] -= gWire
			}

			// Wire right (to row_node[i][j+1], only if not the right-most node):
			if j < M-1 {
				A[eq][eq] += gWire
				A[eq][rowIdx(i, j+1)] -= gWire
			}
			// (If j == M-1 the right end is open — no wire conductance term.)

			// Cell (i,j) to col_node[j][i]:
			G := conductances[i][j]
			A[eq][eq] += G
			if i == 0 {
				// col_node[j][0] = colVoltages[j]  → known
				A[eq][eq] += 0 // already added G above; subtract connection to known node:
				// correction: col_node[j][0] = colVoltages[j]
				b[eq] += G * colVoltages[j]
			} else {
				A[eq][colIdx(j, i)] -= G
			}
		}
	}

	// ── Solve A·x = b using Gaussian elimination with partial pivoting ──────
	x, err := gaussianEliminate(A, b, nUnknowns)
	if err != nil {
		return nil, fmt.Errorf("crossbar_oracle: %w", err)
	}

	// ── Reconstruct full voltage arrays ────────────────────────────────────
	// colV[j][i] = voltage at col_node[j][i]
	colV := make([][]float64, M)
	for j := 0; j < M; j++ {
		colV[j] = make([]float64, N)
		colV[j][0] = colVoltages[j] // driven top node
		for i := 1; i < N; i++ {
			colV[j][i] = x[colIdx(j, i)]
		}
	}

	// rowV[i][j] = voltage at row_node[i][j]
	rowV := make([][]float64, N)
	for i := 0; i < N; i++ {
		rowV[i] = make([]float64, M)
		rowV[i][0] = 0 // grounded left node
		for j := 1; j < M; j++ {
			rowV[i][j] = x[rowIdx(i, j)]
		}
	}

	// ── Compute cell currents ───────────────────────────────────────────────
	cellI := make([][]float64, N)
	for i := 0; i < N; i++ {
		cellI[i] = make([]float64, M)
		for j := 0; j < M; j++ {
			cellI[i][j] = conductances[i][j] * (colV[j][i] - rowV[i][j])
		}
	}

	// ── Compute row output currents ─────────────────────────────────────────
	// I_out[i] = current absorbed by the grounded node row_node[i][0]:
	//   = (cell i,j=0 → row_node[i][0])  +  (wire from row_node[i][1])
	rowI := make([]float64, N)
	for i := 0; i < N; i++ {
		// From cell (i, j=0): G[i][0]·(colV[0][i] - 0)
		rowI[i] += conductances[i][0] * colV[0][i]
		// From wire (row_node[i][1] → row_node[i][0]) if M > 1:
		if M > 1 {
			rowI[i] += gWire * rowV[i][1]
		}
	}

	return &CrossbarResult{
		RowCurrents:  rowI,
		CellCurrents: cellI,
		ColNodeV:     colV,
		RowNodeV:     rowV,
	}, nil
}

// idealMVM computes the exact output for zero wire resistance.
func idealMVM(conductances [][]float64, colVoltages []float64, N, M int) *CrossbarResult {
	rowI := make([]float64, N)
	cellI := make([][]float64, N)
	colV := make([][]float64, M)
	rowV := make([][]float64, N)

	for j := 0; j < M; j++ {
		colV[j] = make([]float64, N)
		for i := 0; i < N; i++ {
			colV[j][i] = colVoltages[j]
		}
	}
	for i := 0; i < N; i++ {
		rowV[i] = make([]float64, M) // all zeros
		cellI[i] = make([]float64, M)
		for j := 0; j < M; j++ {
			cellI[i][j] = conductances[i][j] * colVoltages[j]
			rowI[i] += cellI[i][j]
		}
	}
	return &CrossbarResult{
		RowCurrents:  rowI,
		CellCurrents: cellI,
		ColNodeV:     colV,
		RowNodeV:     rowV,
	}
}

// gaussianEliminate solves A·x = b in-place using partial pivoting.
// Returns the solution vector x or an error if the matrix is singular.
func gaussianEliminate(A [][]float64, b []float64, n int) ([]float64, error) {
	// Augmented matrix [A | b]
	aug := make([][]float64, n)
	for i := range aug {
		aug[i] = make([]float64, n+1)
		copy(aug[i], A[i])
		aug[i][n] = b[i]
	}

	for col := 0; col < n; col++ {
		// Partial pivoting: find row with largest absolute value in this column
		maxVal := math.Abs(aug[col][col])
		maxRow := col
		for row := col + 1; row < n; row++ {
			if v := math.Abs(aug[row][col]); v > maxVal {
				maxVal = v
				maxRow = row
			}
		}
		if maxVal < 1e-18 {
			return nil, fmt.Errorf("singular or near-singular matrix at column %d", col)
		}
		aug[col], aug[maxRow] = aug[maxRow], aug[col]

		// Eliminate below
		pivot := aug[col][col]
		for row := col + 1; row < n; row++ {
			factor := aug[row][col] / pivot
			for k := col; k <= n; k++ {
				aug[row][k] -= factor * aug[col][k]
			}
		}
	}

	// Back-substitution
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		x[i] = aug[i][n]
		for j := i + 1; j < n; j++ {
			x[i] -= aug[i][j] * x[j]
		}
		x[i] /= aug[i][i]
	}
	return x, nil
}
