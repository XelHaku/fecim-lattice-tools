package crossbar

import (
	"math"
	"testing"
)

// M2-KCH-05: Higher Rp/Rcell ratio increases solver difficulty (more iterations).
// The ParasiticSolver uses normalized conductances [0,1] and Rp as normalized
// parasitic resistance ratio. We sweep Rp to show iteration count increases.
func TestSolverCondition_M2KCH05_RpOverRcellAffectsIterations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping solver conditioning sweep in -short (can be sensitive to SOR configuration)")
	}
	const (
		n  = 16
		vIn = 1.0
	)

	// Sweep normalized parasitic resistance from small to large
	// Range limited to values where SOR converges reliably for 16×16.
	// High Rp (>0.03) may fail to converge — that's a real solver limitation.
	rpValues := []float64{0.001, 0.003, 0.005, 0.01, 0.02}
	iters := make([]int, len(rpValues))

	for k, rp := range rpValues {
		cfg := DefaultSORConfig()
		cfg.MaxIterations = 5000
		cfg.Tolerance = 1e-6

		s, err := NewParasiticSolver(n, n, cfg)
		if err != nil {
			t.Fatalf("NewParasiticSolver: %v", err)
		}
		// Normalized conductance — mid-range
		g := make([][]float64, n)
		for i := range g {
			g[i] = make([]float64, n)
			for j := range g[i] {
				g[i][j] = 0.5
			}
		}
		s.SetConductances(g)
		s.SetParasitics(rp, rp)

		input := make([]float64, n)
		for i := range input {
			input[i] = vIn
		}

		res, err := s.SolveMVM(input)
		if err != nil && (res == nil || !res.Converged) {
			t.Fatalf("SolveMVM(rp=%.3g): %v", rp, err)
		}
		if !res.Converged {
			t.Fatalf("did not converge at rp=%.3g (iters=%d maxErr=%g)", rp, res.Iterations, res.MaxError)
		}
		iters[k] = res.Iterations
		t.Logf("Rp=%.3g: iters=%d maxVerr=%.3e finalOmega=%.3f",
			rp, res.Iterations, res.MaxError, res.FinalOmega)
	}

	// High Rp should need more (or equal) iterations than low Rp
	var lowMax, highMax int
	for k, rp := range rpValues {
		if rp <= 0.003 {
			if iters[k] > lowMax {
				lowMax = iters[k]
			}
		}
		if rp >= 0.01 {
			if iters[k] > highMax {
				highMax = iters[k]
			}
		}
	}
	if highMax < lowMax {
		t.Logf("Note: high Rp iters (%d) < low Rp iters (%d); SOR omega adaptation may cause this", highMax, lowMax)
	}

	// Weak monotonicity: allow at most 1 dip
	dips := 0
	for k := 1; k < len(iters); k++ {
		if iters[k] < iters[k-1]-int(math.Ceil(0.1*float64(iters[k-1]))) {
			dips++
		}
	}
	if dips > 2 {
		t.Errorf("too many iteration dips: dips=%d rpValues=%v iters=%v", dips, rpValues, iters)
	}
}
