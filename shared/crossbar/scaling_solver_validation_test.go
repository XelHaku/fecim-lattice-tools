package crossbar

import (
	"math"
	"testing"
)

// M2-SCL-04: Solver iterations vs Rp.
//
// Test-plan units are specified in ohms. The parasitic solver expects a normalized
// Rp = Rp/Rmin (see SetParasitics docs), so we convert using a fixed Rmin proxy.
func TestM2SCL04_SolverIterations_IncreaseWithRp(t *testing.T) {
	if testing.Short() {
		t.Skip("solver iteration sweep can exceed iteration budgets; skip in -short")
	}
	t.Parallel()

	const n = 32

	// Plan sweep (Ω): 0.1Ω → 50Ω.
	rpsOhm := []float64{0.1, 0.2, 0.5, 1, 2, 5, 10}

	const rMinOhm = 1000.0
	rps := make([]float64, 0, len(rpsOhm))
	for _, rp := range rpsOhm {
		rps = append(rps, rp/rMinOhm)
	}

	cfg := &SORConfig{
		MaxIterations: 10000,
		Tolerance:     1e-6,
		OmegaInitial:  1.0,
		OmegaMin:      0.01,
		OmegaDecay:    0.95,
		AdaptiveOmega: true,
	}

	s, err := NewOptimizedParasiticSolver(n, n, cfg)
	if err != nil {
		t.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}
	s.SetConductances(makeUniformConductanceMatrix(n, n, 0.2))

	input := make([]float64, n)
	for i := range input {
		input[i] = 0.5
	}

	iters := make([]int, 0, len(rps))
	for i, rp := range rps {
		s.SetParasitics(rp, rp)
		_, it, solveErr := s.SolveMVMFast(input)
		if solveErr != nil {
			t.Fatalf("SolveMVMFast(rp=%.3gΩ, normalized=%.3g): %v", rpsOhm[i], rp, solveErr)
		}
		iters = append(iters, it)
		t.Logf("rp=%.3gΩ (normalized=%.3g) iters=%d", rpsOhm[i], rp, it)
	}

	const allowDip = 2
	for i := 1; i < len(iters); i++ {
		if iters[i]+allowDip < iters[i-1] {
			t.Fatalf("iterations decreased with increasing Rp: rp=%.3gΩ iters=%d -> rp=%.3gΩ iters=%d (allowDip=%d)",
				rpsOhm[i-1], iters[i-1], rpsOhm[i], iters[i], allowDip)
		}
	}

	if iters[len(iters)-1] <= iters[0]+5 {
		t.Fatalf("expected higher iteration count at high Rp, got rp=%.3gΩ iters=%d and rp=%.3gΩ iters=%d",
			rpsOhm[0], iters[0], rpsOhm[len(rpsOhm)-1], iters[len(iters)-1])
	}

	// Positive correlation check (log scale of normalized Rp).
	var (
		sumX, sumY, sumXY, sumXX, sumYY float64
		m                               = float64(len(rps))
	)
	for i := range rps {
		x := math.Log10(rps[i])
		y := float64(iters[i])
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
		sumYY += y * y
	}
	cov := sumXY - (sumX*sumY)/m
	varX := sumXX - (sumX*sumX)/m
	varY := sumYY - (sumY*sumY)/m
	corr := cov / math.Sqrt(varX*varY+1e-15)
	if corr < 0.7 {
		t.Fatalf("expected strong positive correlation between log10(Rp_norm) and iterations; corr=%.3f", corr)
	}
	t.Logf("corr(log10(Rp_norm), iterations)=%.3f", corr)
}
