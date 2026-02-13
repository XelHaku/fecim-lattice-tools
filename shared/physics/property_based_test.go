package physics_test

import (
	"math"
	"math/rand"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

func adcCodeFromNormalized(x float64, bits int) int {
	if bits <= 0 {
		bits = 1
	}
	if math.IsNaN(x) || math.IsInf(x, 0) {
		x = 0
	}
	if x < 0 {
		x = 0
	}
	if x > 1 {
		x = 1
	}
	maxCode := (1 << bits) - 1
	code := int(math.Round(x * float64(maxCode)))
	if code < 0 {
		code = 0
	}
	if code > maxCode {
		code = maxCode
	}
	return code
}

func TestProperty_PolarizationBounded(t *testing.T) {
	r := rand.New(rand.NewSource(20260213))
	mat := physics.DefaultHZO()
	s := physics.NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false
	s.EnableNoise = false
	s.SetState(0)

	for i := 0; i < 100; i++ {
		E := (-5.0 + 10.0*r.Float64()) * mat.Ec
		dt := 1e-13 + r.Float64()*(5e-11-1e-13)
		P := s.Step(E, dt)
		bound := mat.Ps * 1.1
		if math.Abs(P) > bound+1e-12 {
			t.Fatalf("sample %d: |P| exceeds bound: E=%.3e dt=%.3e P=%.6e bound=%.6e", i, E, dt, math.Abs(P), bound)
		}
	}
}

func TestProperty_EnergyPositive(t *testing.T) {
	mat := physics.DefaultHZO()
	s := physics.NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false
	s.EnableNoise = false
	s.SetState(-math.Abs(mat.Pr))

	eMax := 3.0 * mat.Ec
	const (
		nPtsHalf      = 401
		stepsPerPoint = 200
		dt            = 2e-12
	)

	fields := make([]float64, 0, 2*nPtsHalf)
	pols := make([]float64, 0, 2*nPtsHalf)

	for i := 0; i < nPtsHalf; i++ {
		E := -eMax + (2*eMax*float64(i))/float64(nPtsHalf-1)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		fields = append(fields, E)
		pols = append(pols, s.GetState())
	}
	for i := 0; i < nPtsHalf; i++ {
		E := eMax - (2*eMax*float64(i))/float64(nPtsHalf-1)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		fields = append(fields, E)
		pols = append(pols, s.GetState())
	}

	energy := 0.0
	for i := 1; i < len(fields); i++ {
		dP := pols[i] - pols[i-1]
		Eavg := 0.5 * (fields[i] + fields[i-1])
		energy += Eavg * dP
	}
	if energy < -1e-9 {
		t.Fatalf("loop energy must be non-negative: integral(E·dP)=%.6e", energy)
	}
}

func TestProperty_ConductanceMonotonic(t *testing.T) {
	mat := physics.DefaultHZO()
	const n = 100
	prev := math.Inf(-1)
	for i := 0; i < n; i++ {
		frac := float64(i) / float64(n-1)
		P := -mat.Ps + 2*mat.Ps*frac
		G := physics.PolarizationToConductance(P, mat.Ps, mat.Gmin, mat.Gmax)
		if G <= prev {
			t.Fatalf("conductance not strictly increasing at i=%d: prev=%.6e current=%.6e", i, prev, G)
		}
		prev = G
	}
}

func TestProperty_ADCCodeInRange(t *testing.T) {
	r := rand.New(rand.NewSource(20260214))
	for i := 0; i < 100; i++ {
		bits := 4 + r.Intn(7) // 4..10
		x := -2.0 + 5.0*r.Float64()
		code := adcCodeFromNormalized(x, bits)
		maxCode := (1 << bits) - 1
		if code < 0 || code > maxCode {
			t.Fatalf("sample %d: code out of range: bits=%d x=%.6f code=%d max=%d", i, bits, x, code, maxCode)
		}
	}
}

func TestProperty_SenseChainMonotonic(t *testing.T) {
	r := rand.New(rand.NewSource(20260215))
	for i := 0; i < 100; i++ {
		bits := 4 + r.Intn(5) // 4..8
		maxCode := (1 << bits) - 1
		prev := -1
		for codeTarget := 0; codeTarget <= maxCode; codeTarget++ {
			norm := float64(codeTarget) / float64(maxCode)
			G := physics.GMin + norm*(physics.GMax-physics.GMin)
			normBack := physics.PhysicalToNormalized(G)
			code := adcCodeFromNormalized(normBack, bits)
			if code <= prev {
				t.Fatalf("sample %d: ADC code not strictly increasing at target=%d: prev=%d current=%d", i, codeTarget, prev, code)
			}
			prev = code
		}
	}
}
