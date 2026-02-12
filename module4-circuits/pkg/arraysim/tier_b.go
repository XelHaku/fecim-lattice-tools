package arraysim

import (
	"errors"
	"fmt"
	"math"
)

// DCEngine extends Engine with access to the full DC nodal solution (internal line node voltages).
//
// Tier A provides a fast approximation; Tier B solves the full resistive WL/BL network
// using a sparse iterative solver (PCG with Jacobi preconditioning).
type DCEngine interface {
	Engine
	SolveDC(params SolveParams) (DCResult, error)
}

// DCResult contains the standard per-cell outputs plus the internal line node voltages.
//
// WLNodes and BLNodes are sized [rows][cols] and correspond to the wordline/bitline
// node at each cell intersection.
type DCResult struct {
	SolveResult
	WLNodes [][]float64
	BLNodes [][]float64
}

// TierBSolver provides a scalable Tier B DC nodal solver.
type TierBSolver struct {
	// MaxIter bounds solver iterations. If <=0, default is used.
	MaxIter int
	// RelativeTolerance is the relative residual tolerance. If <=0, default is used.
	RelativeTolerance float64
	// AbsoluteTolerance is the absolute residual tolerance. If <=0, default is used.
	AbsoluteTolerance float64
}

// NewTierBSolver returns a Tier B solver instance.
func NewTierBSolver() *TierBSolver {
	return &TierBSolver{
		MaxIter:           4000,
		RelativeTolerance: 1e-8,
		AbsoluteTolerance: 1e-12,
	}
}

func (t *TierBSolver) maxIter() int {
	if t == nil || t.MaxIter <= 0 {
		return 4000
	}
	return t.MaxIter
}

func (t *TierBSolver) relTol() float64 {
	if t == nil || t.RelativeTolerance <= 0 {
		return 1e-8
	}
	return t.RelativeTolerance
}

func (t *TierBSolver) absTol() float64 {
	if t == nil || t.AbsoluteTolerance <= 0 {
		return 1e-12
	}
	return t.AbsoluteTolerance
}

// Solve satisfies Engine by returning only the per-cell outputs.
func (t *TierBSolver) Solve(params SolveParams) (SolveResult, error) {
	res, err := t.SolveDC(params)
	if err != nil {
		return SolveResult{}, err
	}
	return res.SolveResult, nil
}

// SolveDC returns the full DC solution including internal line node voltages.
func (t *TierBSolver) SolveDC(params SolveParams) (DCResult, error) {
	rows := len(params.Conductance)
	cols := len(params.BLVoltages)
	if cols == 0 {
		for _, row := range params.Conductance {
			if len(row) > cols {
				cols = len(row)
			}
		}
	}
	if rows == 0 || cols == 0 {
		return DCResult{SolveResult: SolveResult{}}, nil
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

	n := 2 * rows * cols
	idxW := func(r, c int) int { return r*cols + c }
	idxB := func(r, c int) int { return rows*cols + r*cols + c }
	gWL := 1.0 / wire.RWordLine
	gBL := 1.0 / wire.RBitLine

	diag := make([]float64, n)
	b := make([]float64, n)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			w := idxW(r, c)
			bl := idxB(r, c)

			if c > 0 {
				diag[w] += gWL
			}
			if c < cols-1 {
				diag[w] += gWL
			}
			if c == 0 {
				wlV := 0.0
				if r < len(params.WLVoltages) {
					wlV = params.WLVoltages[r]
				}
				diag[w] += gWL
				b[w] += gWL * wlV
			}

			if r > 0 {
				diag[bl] += gBL
			}
			if r < rows-1 {
				diag[bl] += gBL
			}
			if r == 0 {
				blV := 0.0
				if c < len(params.BLVoltages) {
					blV = params.BLVoltages[c]
				}
				diag[bl] += gBL
				b[bl] += gBL * blV
			}

			gcell := 0.0
			if r < len(params.Conductance) && c < len(params.Conductance[r]) {
				gcell = params.Conductance[r][c]
			}
			if !rowActive(r) || !selectorEnabled(params, r, c) {
				gcell = 0
			}
			diag[w] += gcell
			diag[bl] += gcell
		}
	}

	applyA := func(x, y []float64) {
		for i := range y {
			y[i] = diag[i] * x[i]
		}
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				w := idxW(r, c)
				bl := idxB(r, c)
				if c > 0 {
					left := idxW(r, c-1)
					y[w] -= gWL * x[left]
				}
				if c < cols-1 {
					right := idxW(r, c+1)
					y[w] -= gWL * x[right]
				}
				if r > 0 {
					up := idxB(r-1, c)
					y[bl] -= gBL * x[up]
				}
				if r < rows-1 {
					down := idxB(r+1, c)
					y[bl] -= gBL * x[down]
				}
				gcell := 0.0
				if r < len(params.Conductance) && c < len(params.Conductance[r]) {
					gcell = params.Conductance[r][c]
				}
				if !rowActive(r) || !selectorEnabled(params, r, c) {
					gcell = 0
				}
				if gcell != 0 {
					y[w] -= gcell * x[bl]
					y[bl] -= gcell * x[w]
				}
			}
		}
	}

	x, err := pcgSolve(applyA, diag, b, t.maxIter(), t.relTol(), t.absTol())
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

func pcgSolve(applyA func(x, y []float64), diag, b []float64, maxIter int, relTol, absTol float64) ([]float64, error) {
	n := len(b)
	if n == 0 {
		return []float64{}, nil
	}
	if len(diag) != n {
		return nil, errors.New("arraysim: pcg dimension mismatch")
	}
	for i := range diag {
		if !(diag[i] > 0) {
			return nil, fmt.Errorf("arraysim: non-positive diagonal at %d", i)
		}
	}

	x := make([]float64, n)
	r := make([]float64, n)
	z := make([]float64, n)
	p := make([]float64, n)
	Ap := make([]float64, n)

	copy(r, b) // x=0 => r=b
	bnorm := math.Sqrt(dot(b, b))
	if bnorm == 0 {
		return x, nil
	}
	tol := math.Max(absTol, relTol*bnorm)

	for i := 0; i < n; i++ {
		z[i] = r[i] / diag[i]
		p[i] = z[i]
	}
	rzOld := dot(r, z)
	if math.Sqrt(dot(r, r)) <= tol {
		return x, nil
	}

	for it := 0; it < maxIter; it++ {
		applyA(p, Ap)
		den := dot(p, Ap)
		if den == 0 {
			return nil, errors.New("arraysim: pcg breakdown (zero denominator)")
		}
		alpha := rzOld / den

		for i := 0; i < n; i++ {
			x[i] += alpha * p[i]
			r[i] -= alpha * Ap[i]
		}

		rnorm := math.Sqrt(dot(r, r))
		if rnorm <= tol {
			return x, nil
		}

		for i := 0; i < n; i++ {
			z[i] = r[i] / diag[i]
		}
		rzNew := dot(r, z)
		beta := rzNew / rzOld
		for i := 0; i < n; i++ {
			p[i] = z[i] + beta*p[i]
		}
		rzOld = rzNew
	}

	return nil, fmt.Errorf("arraysim: pcg did not converge within %d iterations", maxIter)
}

func dot(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}
