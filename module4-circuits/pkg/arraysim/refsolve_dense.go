package arraysim

import (
	"errors"
	"math"
)

// referenceSolveDense computes a DC nodal solution using a dense linear solve.
//
// It is intended as a correctness oracle and regression harness for small arrays.
// The network model is a resistive wordline/bitline grid with a conductance at each
// intersection.
//
// Boundary conditions (aligned with project SPICE deck conventions):
//   - Each row WL is driven from the left through one segment resistance RWordLine
//     (Thevenin drive point at c=0).
//   - Each column BL is driven from the top through one segment resistance RBitLine
//     (Thevenin drive point at r=0).
//   - Opposite ends are left open (no explicit termination resistor).
//   - Wordline segments connect adjacent columns; bitline segments connect adjacent rows.
func referenceSolveDense(params SolveParams) (DCResult, error) {
	rows := len(params.Conductance)
	if rows == 0 {
		return DCResult{SolveResult: SolveResult{}}, nil
	}

	cols := len(params.BLVoltages)
	if cols == 0 {
		for _, row := range params.Conductance {
			if len(row) > cols {
				cols = len(row)
			}
		}
	}
	if cols == 0 {
		return DCResult{}, errors.New("arraysim: no columns available")
	}

	geom := params.Geometry.WithDefaults()
	wire := params.Wire.WithDefaults(geom)
	if wire.RWordLine <= 0 || wire.RBitLine <= 0 {
		return DCResult{}, errors.New("arraysim: invalid wire parameters")
	}

	rowActive := func(r int) bool {
		if params.ActiveRows == nil {
			return true
		}
		if r < 0 || r >= len(params.ActiveRows) {
			return false
		}
		return params.ActiveRows[r]
	}

	// Unknown vector x is [WLNodes | BLNodes], each block is rows*cols.
	n := 2 * rows * cols
	A := make([][]float64, n)
	for i := range A {
		A[i] = make([]float64, n)
	}
	b := make([]float64, n)

	idxW := func(r, c int) int { return r*cols + c }
	idxB := func(r, c int) int { return rows*cols + r*cols + c }

	gWL := 1.0 / wire.RWordLine
	gBL := 1.0 / wire.RBitLine

	addConductance := func(i, j int, g float64) {
		if g == 0 {
			return
		}
		A[i][i] += g
		A[j][j] += g
		A[i][j] -= g
		A[j][i] -= g
	}

	addToSource := func(i int, g float64, vsrc float64) {
		if g == 0 {
			return
		}
		A[i][i] += g
		b[i] += g * vsrc
	}

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			w := idxW(r, c)
			bl := idxB(r, c)

			// Skip inactive rows by disconnecting their cells and WL source.
			active := rowActive(r)

			// Wordline neighbor segments (add each segment once).
			if c > 0 {
				addConductance(w, idxW(r, c-1), gWL)
			}
			if c == 0 {
				wlV := 0.0
				if r < len(params.WLVoltages) {
					wlV = params.WLVoltages[r]
				}
				// Even if the row is inactive, the line itself is still held at its driver voltage;
				// only the cell conductances are gated off.
				addToSource(w, gWL, wlV)
			}

			// Bitline neighbor segments (add each segment once).
			if r > 0 {
				addConductance(bl, idxB(r-1, c), gBL)
			}
			if r == 0 {
				blV := 0.0
				if c < len(params.BLVoltages) {
					blV = params.BLVoltages[c]
				}
				addToSource(bl, gBL, blV)
			}

			// Cell conductance.
			gcell := 0.0
			if r < len(params.Conductance) && c < len(params.Conductance[r]) {
				gcell = params.Conductance[r][c]
			}
			if !active || !selectorEnabled(params, r, c) {
				gcell = 0
			}
			addConductance(w, bl, gcell)
		}
	}

	x, err := gaussianElimSolve(A, b)
	if err != nil {
		return DCResult{}, err
	}

	wlNodes := make([][]float64, rows)
	blNodes := make([][]float64, rows)
	cellVoltages := make([][]float64, rows)
	cellCurrents := make([][]float64, rows)
	rowCurrents := make([]float64, rows)
	colCurrents := make([]float64, cols)

	for r := 0; r < rows; r++ {
		wlNodes[r] = make([]float64, cols)
		blNodes[r] = make([]float64, cols)
		cellVoltages[r] = make([]float64, cols)
		cellCurrents[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			vw := x[idxW(r, c)]
			vb := x[idxB(r, c)]
			wlNodes[r][c] = vw
			blNodes[r][c] = vb

			gcell := 0.0
			if r < len(params.Conductance) && c < len(params.Conductance[r]) {
				gcell = params.Conductance[r][c]
			}
			if !rowActive(r) || !selectorEnabled(params, r, c) {
				gcell = 0
			}

			v := vw - vb
			i := gcell * v
			cellVoltages[r][c] = v
			cellCurrents[r][c] = i
			rowCurrents[r] += i
			colCurrents[c] += i
		}
	}

	return DCResult{
		SolveResult: SolveResult{
			CellVoltages: cellVoltages,
			CellCurrents: cellCurrents,
			RowCurrents:  rowCurrents,
			ColCurrents:  colCurrents,
		},
		WLNodes: wlNodes,
		BLNodes: blNodes,
	}, nil
}

// gaussianElimSolve solves A x = b using naive Gaussian elimination with partial pivoting.
//
// This is intended only for small systems (e.g., <= 128 unknowns).
func gaussianElimSolve(A [][]float64, b []float64) ([]float64, error) {
	n := len(A)
	if n == 0 {
		return []float64{}, nil
	}
	if len(b) != n {
		return nil, errors.New("arraysim: dimension mismatch")
	}
	for i := range A {
		if len(A[i]) != n {
			return nil, errors.New("arraysim: matrix must be square")
		}
	}

	// Build augmented matrix.
	M := make([][]float64, n)
	for i := 0; i < n; i++ {
		row := make([]float64, n+1)
		copy(row, A[i])
		row[n] = b[i]
		M[i] = row
	}

	for col := 0; col < n; col++ {
		// Pivot.
		piv := col
		maxAbs := math.Abs(M[col][col])
		for r := col + 1; r < n; r++ {
			v := math.Abs(M[r][col])
			if v > maxAbs {
				maxAbs = v
				piv = r
			}
		}
		if maxAbs == 0 {
			return nil, errors.New("arraysim: singular matrix")
		}
		if piv != col {
			M[col], M[piv] = M[piv], M[col]
		}

		// Eliminate below.
		pivotVal := M[col][col]
		for r := col + 1; r < n; r++ {
			f := M[r][col] / pivotVal
			if f == 0 {
				continue
			}
			M[r][col] = 0
			for c := col + 1; c < n+1; c++ {
				M[r][c] -= f * M[col][c]
			}
		}
	}

	// Back substitution.
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		sum := M[i][n]
		for c := i + 1; c < n; c++ {
			sum -= M[i][c] * x[c]
		}
		if M[i][i] == 0 {
			return nil, errors.New("arraysim: singular matrix")
		}
		x[i] = sum / M[i][i]
	}
	return x, nil
}
