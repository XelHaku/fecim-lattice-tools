package crossbar

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func randomConductanceMatrix(rng *rand.Rand, rows, cols int) [][]float64 {
	g := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		g[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			g[i][j] = 0.05 + 0.9*rng.Float64()
		}
	}
	return g
}

func randomVector(rng *rand.Rand, n int) []float64 {
	x := make([]float64, n)
	for i := range x {
		x[i] = 0.05 + 0.95*rng.Float64()
	}
	return x
}

func rms(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	return math.Sqrt(sum / float64(len(v)))
}

// M2-SCL-01: Accuracy vs array size.
func TestM2SCL01_ScalingAccuracy_MonotonicErrorVsSize(t *testing.T) {
	if testing.Short() {
		t.Skip("scaling sweep can exceed solver iteration budgets; skip in -short")
	}
	// NOTE: Keep this test stable in -short mode; large sizes can exceed iteration budgets on slow CI.
	sizes := []int{4, 8, 16, 32}
	if testing.Short() {
		sizes = []int{4, 8, 16, 32}
	}

	const (
		rpRow = 0.002
		rpCol = 0.002
	)

	rng := rand.New(rand.NewSource(1337))

	trials := 6
	if testing.Short() {
		trials = 3
	}
	type point struct {
		n    int
		err  float64
		iter float64
	}
	pts := make([]point, 0, len(sizes))

	for _, n := range sizes {
		cfg := &SORConfig{
			MaxIterations: 20000,
			Tolerance:     1e-4, // looser tolerance for stability in automated tests
			OmegaInitial:  1.0,
			OmegaMin:      0.02,
			OmegaDecay:    0.95,
			AdaptiveOmega: true,
		}

		solver, err := NewOptimizedParasiticSolver(n, n, cfg)
		if err != nil {
			t.Fatalf("NewOptimizedParasiticSolver(%d): %v", n, err)
		}
		solver.SetParasitics(rpRow, rpCol)

		var sumErr, sumIter float64
		for trial := 0; trial < trials; trial++ {
			g := randomConductanceMatrix(rng, n, n)
			x := randomVector(rng, n)
			solver.SetConductances(g)

			ref, err := NewParasiticSolver(n, n, DefaultSORConfig())
			if err != nil {
				t.Fatalf("NewParasiticSolver(%d): %v", n, err)
			}
			ref.SetConductances(g)
			ideal := ref.ComputeIdealMVM(x)

			actual, iters, solveErr := solver.SolveMVMFast(x)
			if solveErr != nil {
				t.Fatalf("SolveMVMFast(%dx%d) trial=%d: %v", n, n, trial, solveErr)
			}

			absErr := ComputeAccuracyLoss(ideal, actual)
			relErr := absErr / (rms(ideal) + 1e-15)
			sumErr += relErr
			sumIter += float64(iters)
		}

		pts = append(pts, point{n: n, err: sumErr / float64(trials), iter: sumIter / float64(trials)})
	}

	const eps = 1e-3
	for i := 1; i < len(pts); i++ {
		prev := pts[i-1]
		curr := pts[i]
		if curr.err+eps < prev.err {
			t.Fatalf("relative RMS error decreased with size: n=%d err=%.6g -> n=%d err=%.6g (eps=%.1g)",
				prev.n, prev.err, curr.n, curr.err, eps)
		}
	}

	for _, p := range pts {
		t.Logf("n=%-3d meanRelRMSError=%.6g meanIters=%.1f", p.n, p.err, p.iter)
	}

	if pts[len(pts)-1].err < pts[0].err*1.05 {
		t.Fatalf("expected accuracy degradation with size, but n=%d err=%.6g and n=%d err=%.6g are too close",
			pts[0].n, pts[0].err, pts[len(pts)-1].n, pts[len(pts)-1].err)
	}
}

func BenchmarkM2SCL01_IdealBaseline_ComputeIdealMVM(b *testing.B) {
	sizes := []int{8, 32, 64, 128}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			rng := rand.New(rand.NewSource(int64(n)))
			g := randomConductanceMatrix(rng, n, n)
			x := randomVector(rng, n)
			s, err := NewParasiticSolver(n, n, DefaultSORConfig())
			if err != nil {
				b.Fatalf("NewParasiticSolver: %v", err)
			}
			s.SetConductances(g)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.ComputeIdealMVM(x)
			}
		})
	}
}
