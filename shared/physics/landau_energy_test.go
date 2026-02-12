package physics

import (
	"math"
	"sort"
	"testing"
)

type extremum struct {
	P      float64
	stable bool
}

func landauExtrema(alpha, beta, gamma float64) []extremum {
	const eps = 1e-12
	roots := []float64{0}

	a := 6.0 * gamma
	b := 4.0 * beta
	c := 2.0 * alpha
	if math.Abs(a) > eps {
		disc := b*b - 4*a*c
		if disc >= 0 {
			sqrtDisc := math.Sqrt(disc)
			x1 := (-b + sqrtDisc) / (2 * a)
			x2 := (-b - sqrtDisc) / (2 * a)
			for _, x := range []float64{x1, x2} {
				if x > eps {
					p := math.Sqrt(x)
					roots = append(roots, -p, p)
				}
			}
		}
	}

	sort.Float64s(roots)
	uniq := make([]float64, 0, len(roots))
	for _, r := range roots {
		if len(uniq) == 0 || math.Abs(r-uniq[len(uniq)-1]) > 1e-10 {
			uniq = append(uniq, r)
		}
	}

	extrema := make([]extremum, 0, len(uniq))
	for _, p := range uniq {
		p2 := p * p
		d2 := (2 * alpha) + (12 * beta * p2) + (30 * gamma * p2 * p2)
		extrema = append(extrema, extremum{P: p, stable: d2 > 0})
	}
	return extrema
}

func countEquilibria(alpha, beta, gamma, E, pMax float64) int {
	f := func(p float64) float64 {
		p2 := p * p
		p3 := p2 * p
		p5 := p3 * p2
		return (2*alpha*p + 4*beta*p3 + 6*gamma*p5) - E
	}

	const n = 8000
	const eps = 1e-9
	count := 0
	prevP := -pMax
	prevF := f(prevP)
	if math.Abs(prevF) < eps {
		count++
	}
	for i := 1; i <= n; i++ {
		p := -pMax + 2*pMax*float64(i)/float64(n)
		curF := f(p)
		if math.Abs(curF) < eps {
			count++
		} else if prevF*curF < 0 {
			count++
		}
		prevP = p
		_ = prevP
		prevF = curF
	}
	return count
}

func switchingDissipation(s *LKSolver, E, dt float64, steps int) (diss float64) {
	rhoEff := s.effectiveRho()
	for i := 0; i < steps; i++ {
		rate := s.dPdT(s.Time, s.P, E, 0, rhoEff)
		diss += rhoEff * rate * rate * dt
		s.Step(E, dt)
	}
	return diss
}

func timeToReach(target, E, dt float64, steps int, s *LKSolver) float64 {
	for i := 0; i < steps; i++ {
		if s.P >= target {
			return float64(i) * dt
		}
		s.Step(E, dt)
	}
	return math.Inf(1)
}

func TestLandauEnergyLandscape_DoubleWellExtremaAtZeroField(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.K_dep = 0

	extrema := landauExtrema(s.Alpha, s.Beta, s.Gamma)
	if len(extrema) != 3 {
		t.Fatalf("expected exactly 3 extrema at E=0 (2 minima + 1 maximum), got %d", len(extrema))
	}

	stable := 0
	unstable := 0
	for _, ex := range extrema {
		if ex.stable {
			stable++
		} else {
			unstable++
		}
	}
	if stable != 2 || unstable != 1 {
		t.Fatalf("expected 2 stable minima and 1 unstable maximum, got stable=%d unstable=%d", stable, unstable)
	}

	if math.Abs(extrema[1].P) > 1e-9 {
		t.Fatalf("expected middle extremum near P=0, got P=%.6e", extrema[1].P)
	}
}

func TestLandauEnergyLandscape_BarrierVanishesAboveCoerciveField(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.K_dep = 0

	pRange := 1.5 * s.PMax
	below := countEquilibria(s.Alpha, s.Beta, s.Gamma, 0.95*mat.Ec, pRange)
	above := countEquilibria(s.Alpha, s.Beta, s.Gamma, 1.05*mat.Ec, pRange)

	if below < 3 {
		t.Fatalf("expected bistable region below Ec to keep barrier (>=3 equilibria), got %d", below)
	}
	if above != 1 {
		t.Fatalf("expected single equilibrium above Ec (barrier vanished), got %d", above)
	}
}

func TestLandauEnergyLandscape_SwitchingDissipationIsPositive(t *testing.T) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.K_dep = 0
	s.UseNLS = false
	s.EnableNoise = false
	s.UseEffectiveViscosity = false
	s.SeriesResistance = 0
	s.SetState(-math.Abs(mat.Pr))

	E := 1.2 * mat.Ec
	diss := switchingDissipation(s, E, 2e-11, 8000)
	if !(diss > 0) {
		t.Fatalf("expected positive dissipation during switching, got %.6e", diss)
	}
	if s.P <= 0 {
		t.Fatalf("expected switched polarization to positive branch, got P=%.6e", s.P)
	}
}

func TestLandauEnergyLandscape_ViscosityAffectsSpeedNotFinalState(t *testing.T) {
	mat := DefaultHZO()
	makeSolver := func(rho float64) *LKSolver {
		s := NewLKSolver()
		s.ConfigureFromMaterial(mat)
		s.K_dep = 0
		s.UseNLS = false
		s.EnableNoise = false
		s.UseEffectiveViscosity = false
		s.SeriesResistance = 0
		s.Rho = rho
		s.SetState(-math.Abs(mat.Pr))
		return s
	}

	fast := makeSolver(0.02)
	slow := makeSolver(0.20)
	E := 1.15 * mat.Ec
	dt := 2e-11
	steps := 12000
	target := 0.9 * math.Abs(mat.Pr)

	tFast := timeToReach(target, E, dt, steps, fast)
	tSlow := timeToReach(target, E, dt, steps, slow)
	if !(tSlow > tFast) {
		t.Fatalf("expected higher viscosity to switch slower: tFast=%.3e s, tSlow=%.3e s", tFast, tSlow)
	}

	for i := 0; i < steps; i++ {
		fast.Step(E, dt)
		slow.Step(E, dt)
	}

	if math.Abs(fast.P-slow.P) > 0.01 {
		t.Fatalf("expected similar final states despite viscosity change: fast=%.6e slow=%.6e", fast.P, slow.P)
	}
	if fast.P <= 0 || slow.P <= 0 {
		t.Fatalf("expected both cases to end in positive branch: fast=%.6e slow=%.6e", fast.P, slow.P)
	}
}
