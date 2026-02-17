package crossbar

import (
	"math"
	"testing"
)

func relErr(got, want float64) float64 {
	d := math.Abs(got - want)
	s := math.Max(math.Abs(want), 1e-15)
	return d / s
}

func solveLinearSystemGaussian(A [][]float64, b []float64) ([]float64, bool) {
	n := len(b)
	if len(A) != n {
		return nil, false
	}
	for i := 0; i < n; i++ {
		if len(A[i]) != n {
			return nil, false
		}
	}

	M := make([][]float64, n)
	for i := 0; i < n; i++ {
		row := make([]float64, n+1)
		copy(row, A[i])
		row[n] = b[i]
		M[i] = row
	}

	for k := 0; k < n; k++ {
		pivot := k
		maxAbs := math.Abs(M[k][k])
		for i := k + 1; i < n; i++ {
			v := math.Abs(M[i][k])
			if v > maxAbs {
				maxAbs = v
				pivot = i
			}
		}
		if maxAbs == 0 {
			return nil, false
		}
		if pivot != k {
			M[k], M[pivot] = M[pivot], M[k]
		}

		for i := k + 1; i < n; i++ {
			f := M[i][k] / M[k][k]
			if f == 0 {
				continue
			}
			for j := k; j < n+1; j++ {
				M[i][j] -= f * M[k][j]
			}
		}
	}

	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		sum := M[i][n]
		for j := i + 1; j < n; j++ {
			sum -= M[i][j] * x[j]
		}
		if M[i][i] == 0 {
			return nil, false
		}
		x[i] = sum / M[i][i]
	}
	return x, true
}

func analyticIRDropSolve(rows, cols int, vin []float64, vout []float64, g [][]float64, rRow, rCol float64) (vr, vc [][]float64, ok bool) {
	n := 2 * rows * cols
	A := make([][]float64, n)
	b := make([]float64, n)
	for i := range A {
		A[i] = make([]float64, n)
	}

	idxR := func(i, j int) int { return i*cols + j }
	idxC := func(i, j int) int { return rows*cols + i*cols + j }

	gRow := 0.0
	if rRow > 0 {
		gRow = 1 / rRow
	}
	gCol := 0.0
	if rCol > 0 {
		gCol = 1 / rCol
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			r := idxR(i, j)
			gCell := g[i][j]
			A[r][r] += gCell
			A[r][idxC(i, j)] += -gCell
			if j == 0 {
				A[r][r] += gRow
				b[r] += gRow * vin[i]
			} else {
				A[r][r] += gRow
				A[r][idxR(i, j-1)] += -gRow
			}
			if j < cols-1 {
				A[r][r] += gRow
				A[r][idxR(i, j+1)] += -gRow
			}

			c := idxC(i, j)
			A[c][c] += gCell
			A[c][idxR(i, j)] += -gCell
			if i == rows-1 {
				A[c][c] += gCol
				b[c] += gCol * vout[j]
			} else {
				A[c][c] += gCol
				A[c][idxC(i+1, j)] += -gCol
			}
			if i > 0 {
				A[c][c] += gCol
				A[c][idxC(i-1, j)] += -gCol
			}
		}
	}

	x, ok := solveLinearSystemGaussian(A, b)
	if !ok {
		return nil, nil, false
	}
	vr = make([][]float64, rows)
	vc = make([][]float64, rows)
	for i := 0; i < rows; i++ {
		vr[i] = make([]float64, cols)
		vc[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			vr[i][j] = x[idxR(i, j)]
			vc[i][j] = x[idxC(i, j)]
		}
	}
	return vr, vc, true
}

func TestIRDropAnalytic_1x1(t *testing.T) {
	if testing.Short() {
		t.Skip("IR-drop analytic checks are long/strict; skip in -short")
	}
	const (
		vin  = 1.0
		g    = 50e-6
		rRow = 2.5
		rCol = 2.5
	)
	rSum := rRow + rCol
	I := vin * g / (1.0 + g*rSum)
	vCellWant := vin - I*rSum

	sim := NewIRDropSimulator(1, 1)
	sim.RowResist = rRow
	sim.ColResist = rCol
	sim.SetInputVoltage(0, vin)
	sim.VoltageOut[0] = 0
	sim.SetConductance(0, 0, g)
	sim.Simulate(200)

	vCellGot := sim.RowVoltages[0][0] - sim.ColVoltages[0][0]
	// IR drop simulator uses iterative solver — allow 1% tolerance
	if e := relErr(vCellGot, vCellWant); e >= 1e-2 {
		t.Fatalf("V_cell mismatch: got %.12g want %.12g relErr=%.3g", vCellGot, vCellWant, e)
	}

	Igot := sim.CellCurrents[0][0]
	if e := relErr(Igot, I); e >= 1e-2 {
		t.Fatalf("I mismatch: got %.12g want %.12g relErr=%.3g", Igot, I, e)
	}
}

func TestIRDropAnalytic_2x2UniformHandSolve(t *testing.T) {
	if testing.Short() {
		t.Skip("IR-drop analytic checks are long/strict; skip in -short")
	}
	rows, cols := 2, 2
	vin := []float64{1.0, 0.5}
	vout := []float64{0.0, 0.0}
	const (
		g    = 50e-6
		rRow = 2.5
		rCol = 2.5
	)

	G := make([][]float64, rows)
	for i := range G {
		G[i] = make([]float64, cols)
		for j := range G[i] {
			G[i][j] = g
		}
	}

	vrWant, vcWant, ok := analyticIRDropSolve(rows, cols, vin, vout, G, rRow, rCol)
	if !ok {
		t.Fatalf("analytic solver failed")
	}

	sim := NewIRDropSimulator(rows, cols)
	sim.RowResist = rRow
	sim.ColResist = rCol
	sim.SetAllInputs(vin)
	for j := 0; j < cols; j++ {
		sim.VoltageOut[j] = vout[j]
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			sim.SetConductance(i, j, g)
		}
	}
	sim.Simulate(5000)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			vrGot := sim.RowVoltages[i][j]
			vcGot := sim.ColVoltages[i][j]
			if e := relErr(vrGot, vrWant[i][j]); e >= 1e-2 {
				t.Fatalf("vr[%d][%d] mismatch: got %.12g want %.12g relErr=%.3g", i, j, vrGot, vrWant[i][j], e)
			}
			if e := relErr(vcGot, vcWant[i][j]); e >= 1e-2 {
				t.Fatalf("vc[%d][%d] mismatch: got %.12g want %.12g relErr=%.3g", i, j, vcGot, vcWant[i][j], e)
			}

			vCellWant := vrWant[i][j] - vcWant[i][j]
			vCellGot := vrGot - vcGot
			if e := relErr(vCellGot, vCellWant); e >= 1e-2 {
				t.Fatalf("vCell[%d][%d] mismatch: got %.12g want %.12g relErr=%.3g", i, j, vCellGot, vCellWant, e)
			}

			Iwant := vCellWant * g
			Igot := sim.CellCurrents[i][j]
			if e := relErr(Igot, Iwant); e >= 1e-2 {
				t.Fatalf("I[%d][%d] mismatch: got %.12g want %.12g relErr=%.3g", i, j, Igot, Iwant, e)
			}
		}
	}
}
