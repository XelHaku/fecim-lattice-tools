package physics

import (
	"math"
	"math/rand"
	"testing"
)

func fuzzFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func fuzzRange(r *rand.Rand, lo, hi float64) float64 {
	return lo + (hi-lo)*r.Float64()
}

func TestFuzz_PreisachRandomUpdates_NoNaNInf(t *testing.T) {
	r := rand.New(rand.NewSource(20260212))

	const satMVcm = 5.0
	satE := satMVcm * 1e8 // V/m
	ps := NewPreisachStack(satE, simpleEverett{sat: satE})

	for i := 0; i < 1000; i++ {
		eMVcm := fuzzRange(r, -5.0, 5.0)
		e := eMVcm * 1e8 // V/m
		p := ps.Update(e)
		if !fuzzFinite(p) {
			t.Fatalf("preisach update %d produced non-finite polarization: E=%g MV/cm P=%g", i, eMVcm, p)
		}
	}
}

func TestFuzz_LKSolverRandomTrajectories_Stability(t *testing.T) {
	baseSeed := int64(20260213)

	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(baseSeed + int64(i)*7919))

		s := NewLKSolver()
		s.Beta = -fuzzRange(r, 1e8, 8e8)   // J m^5 / C^4 (negative)
		s.Gamma = fuzzRange(r, 5e9, 3e10)  // J m^9 / C^6 (positive)
		s.Rho = fuzzRange(r, 1e-3, 0.2)    // Ohm·m
		s.Q12 = fuzzRange(r, -0.05, -0.01) // m^4/C^2
		s.Stress = fuzzRange(r, 0.2e9, 2e9)
		s.K_dep = fuzzRange(r, 5e7, 5e8)
		s.SeriesResistance = fuzzRange(r, 5, 150)
		s.Thickness = fuzzRange(r, 6e-9, 20e-9)
		s.Area = fuzzRange(r, 25e-9*25e-9, 120e-9*120e-9)
		s.CurieTemp = fuzzRange(r, 650, 800)
		s.CurieConst = fuzzRange(r, 1e5, 3e5)
		s.ActivationField = fuzzRange(r, 8e8, 2.5e9)
		s.TauInf = fuzzRange(r, 1e-13, 1e-10)
		s.PMax = fuzzRange(r, 0.2, 0.7)
		s.P = fuzzRange(r, -0.8*s.PMax, 0.8*s.PMax)
		s.Temperature = fuzzRange(r, 260, 420)
		s.UseNLS = r.Intn(2) == 0
		s.EnableNoise = r.Intn(4) == 0

		for step := 0; step < 200; step++ {
			e := fuzzRange(r, -3e8, 3e8)
			dt := fuzzRange(r, 1e-13, 2e-9)
			p := s.Step(e, dt)
			if !fuzzFinite(p) || !fuzzFinite(s.GetState()) || !fuzzFinite(s.Time) {
				t.Fatalf("trajectory %d step %d unstable: E=%g dt=%g P=%g state=%g time=%g", i, step, e, dt, p, s.GetState(), s.Time)
			}

			limit := 1.2*s.PMax + 1e-12
			if s.GetState() < -limit || s.GetState() > limit {
				t.Fatalf("trajectory %d step %d exceeded clamp: P=%g limit=%g", i, step, s.GetState(), limit)
			}
		}
	}
}

func TestFuzz_ISPPRandomSequences_ConvergeOrCleanFailure(t *testing.T) {
	r := rand.New(rand.NewSource(20260214))

	for i := 0; i < 50; i++ {
		mat := DefaultHZO()
		mat.Ps = fuzzRange(r, 0.22, 0.45)
		mat.Pr = fuzzRange(r, 0.18, math.Min(0.40, mat.Ps-0.01))
		mat.Ec = fuzzRange(r, 0.7e8, 1.6e8)
		mat.Thickness = fuzzRange(r, 8e-9, 16e-9)
		mat.Tau = fuzzRange(r, 1e-9, 20e-9)
		mat.K_dep = fuzzRange(r, 1e8, 5e8)

		solver := NewLKSolver()
		solver.ConfigureFromMaterial(mat)
		solver.SetState(fuzzRange(r, -0.9*mat.Pr, 0.9*mat.Pr))

		ispp := NewAdaptiveISPP(solver, mat)
		ispp.MaxIterations = 8 + r.Intn(9)              // 8..16
		ispp.TargetTolerance = fuzzRange(r, 0.01, 0.04) // 1..4% of Ps
		ispp.MaxVoltage = fuzzRange(r, 1.5, 4.0)
		ispp.MinVoltage = -ispp.MaxVoltage

		targetP := fuzzRange(r, -0.9*mat.Ps, 0.9*mat.Ps)
		p, iters, converged := ispp.BinarySearchWrite(targetP)

		if !fuzzFinite(p) {
			t.Fatalf("sequence %d returned non-finite P: target=%g P=%g", i, targetP, p)
		}
		if iters < 0 || iters > ispp.MaxIterations {
			t.Fatalf("sequence %d invalid iteration count: got=%d max=%d", i, iters, ispp.MaxIterations)
		}

		tolP := ispp.TargetTolerance * math.Abs(mat.Ps)
		err := math.Abs(p - targetP)
		if converged {
			if err > tolP+1e-12 {
				t.Fatalf("sequence %d marked converged outside tolerance: err=%g tol=%g", i, err, tolP)
			}
		} else {
			if iters != ispp.MaxIterations {
				t.Fatalf("sequence %d failed uncleanly: converged=false but iters=%d (max=%d)", i, iters, ispp.MaxIterations)
			}
		}
	}
}
