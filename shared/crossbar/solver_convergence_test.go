package crossbar

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func maxAbsDiffVec(a, b []float64) float64 {
	max := 0.0
	for i := range a {
		d := math.Abs(a[i] - b[i])
		if d > max {
			max = d
		}
	}
	return max
}

func maxAbsDiffMat(a, b [][]float64) float64 {
	max := 0.0
	for i := range a {
		for j := range a[i] {
			d := math.Abs(a[i][j] - b[i][j])
			if d > max {
				max = d
			}
		}
	}
	return max
}

func TestSolverConvergence_M2KCH03_SORIterationsSublinearVsN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping convergence scaling analysis in -short (sensitive to solver parameters)")
	}
	sizes := []int{4, 8, 16, 32, 64}

	const (
		gUniform = 50e-6
		rpSeg    = 2.5
		vIn      = 0.2
	)

	iters := make([]int, 0, len(sizes))
	for _, n := range sizes {
		cfg := DefaultSORConfig()
		cfg.MaxIterations = 2000

		s, err := NewParasiticSolver(n, n, cfg)
		if err != nil {
			t.Fatalf("NewParasiticSolver(N=%d): %v", n, err)
		}
		s.SetParasitics(rpSeg, rpSeg)
		s.SetConductances(makeUniformG(n, n, gUniform))

		res, err := s.SolveMVM(makeApplied(n, vIn))
		if err != nil && !res.Converged {
			t.Fatalf("SolveMVM(N=%d): %v (iters=%d maxErr=%g)", n, err, res.Iterations, res.MaxError)
		}
		iters = append(iters, res.Iterations)
		t.Logf("N=%d iters=%d converged=%v maxVerr=%.3e V omegaFinal=%.3f", n, res.Iterations, res.Converged, res.MaxError, res.FinalOmega)
	}

	var sumSlope float64
	var maxSlope float64
	for k := 1; k < len(sizes); k++ {
		n1, n2 := float64(sizes[k-1]), float64(sizes[k])
		i1, i2 := float64(iters[k-1]), float64(iters[k])
		slope := math.Log(i2/i1) / math.Log(n2/n1)
		sumSlope += slope
		if slope > maxSlope {
			maxSlope = slope
		}
	}
	avgSlope := sumSlope / float64(len(sizes)-1)
	// Accept if average slope < 0.75 (clearly sublinear) even if one step is linear
	if avgSlope >= 0.75 {
		t.Fatalf("iterations scale too close to linear: avg log-slope=%.3f (want < 0.75); max=%.3f iters=%v sizes=%v", avgSlope, maxSlope, iters, sizes)
	}
	t.Logf("avg log-slope(iterations vs N)=%.3f, max=%.3f (avg < 0.75 PASS)", avgSlope, maxSlope)
}

func TestSolverConvergence_M2KCH04_OptimizedMatchesStandard(t *testing.T) {
	const (
		n        = 32
		rpSeg    = 2.5
		vIn      = 0.2
		diffTol  = 1e-8
		gMinRand = 10e-6
		gMaxRand = 100e-6
	)

	rng := rand.New(rand.NewSource(3))
	g := makeRandomG(n, n, gMinRand, gMaxRand, rng)
	applied := makeApplied(n, vIn)

	cfg := DefaultSORConfig()
	cfg.MaxIterations = 2000
	cfg.Tolerance = 1e-10

	std, err := NewParasiticSolver(n, n, cfg)
	if err != nil {
		t.Fatalf("NewParasiticSolver: %v", err)
	}
	std.SetParasitics(rpSeg, rpSeg)
	std.SetConductances(g)

	opt, err := NewOptimizedParasiticSolver(n, n, cfg)
	if err != nil {
		t.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}
	opt.SetParasitics(rpSeg, rpSeg)
	opt.SetConductances(g)

	rStd, errStd := std.SolveMVM(applied)
	if errStd != nil && !rStd.Converged {
		t.Fatalf("standard SolveMVM: %v (iters=%d maxErr=%g)", errStd, rStd.Iterations, rStd.MaxError)
	}
	rOpt, errOpt := opt.SolveMVM(applied)
	if errOpt != nil && !rOpt.Converged {
		t.Fatalf("optimized SolveMVM: %v (iters=%d maxErr=%g)", errOpt, rOpt.Iterations, rOpt.MaxError)
	}

	outDiff := maxAbsDiffVec(rStd.OutputCurrents, rOpt.OutputCurrents)
	vDiff := maxAbsDiffMat(rStd.DeviceVoltages, rOpt.DeviceVoltages)
	iDiff := maxAbsDiffMat(rStd.DeviceCurrents, rOpt.DeviceCurrents)
	maxDiff := math.Max(outDiff, math.Max(vDiff, iDiff))

	t.Logf("solver agreement: maxDiff=%.3e (out=%.3e, V=%.3e, I=%.3e); iters std=%d opt=%d", maxDiff, outDiff, vDiff, iDiff, rStd.Iterations, rOpt.Iterations)
	if maxDiff >= diffTol {
		t.Fatalf("optimized solver differs from standard: maxDiff=%.3e (want < %.1e)", maxDiff, diffTol)
	}
}

func TestSolverConvergence_M2KCH03_OmegaSweepConverges(t *testing.T) {
	const (
		n        = 32
		gUniform = 50e-6
		rpSeg    = 2.5
		vIn      = 0.2
	)

	omegas := []float64{0.5, 1.0, 1.5, 1.8}
	for _, omega := range omegas {
		t.Run(fmt.Sprintf("omega=%.1f", omega), func(t *testing.T) {
			cfg := DefaultSORConfig()
			cfg.OmegaInitial = omega
			cfg.MaxIterations = 3000
			cfg.Tolerance = 1e-6
			cfg.AdaptiveOmega = true

			s, err := NewParasiticSolver(n, n, cfg)
			if err != nil {
				t.Fatalf("NewParasiticSolver: %v", err)
			}
			s.SetParasitics(rpSeg, rpSeg)
			s.SetConductances(makeUniformG(n, n, gUniform))

			res, err := s.SolveMVM(makeApplied(n, vIn))
			if err != nil && !res.Converged {
				t.Fatalf("SolveMVM(omega=%.1f): %v (iters=%d maxErr=%g finalOmega=%.3f)", omega, err, res.Iterations, res.MaxError, res.FinalOmega)
			}
			if !res.Converged {
				t.Fatalf("did not converge for omega=%.1f (iters=%d maxErr=%g)", omega, res.Iterations, res.MaxError)
			}
			t.Logf("converged=%v iters=%d maxVerr=%.3e V finalOmega=%.3f", res.Converged, res.Iterations, res.MaxError, res.FinalOmega)
		})
	}
}
