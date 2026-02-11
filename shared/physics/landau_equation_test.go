package physics

import (
	"math"
	"testing"
)

func TestLKSolver_dPdT_EquationSignAndTerms_TableDriven(t *testing.T) {
	tests := []struct {
		name   string
		solver LKSolver
		P      float64
		E      float64
		noise  float64
		rhoEff float64
	}{
		{
			name: "reference case with positive depolarization penalty",
			solver: LKSolver{Alpha: 1.2, Beta: -0.5, Gamma: 0.25, K_dep: 2.0},
			P:      0.3,
			E:      5.0,
			noise:  -0.2,
			rhoEff: 1.5,
		},
		{
			name: "negative polarization increases effective field via -k_dep*P",
			solver: LKSolver{Alpha: 0.8, Beta: -0.2, Gamma: 0.05, K_dep: 3.0},
			P:      -0.4,
			E:      2.0,
			noise:  0.0,
			rhoEff: 0.9,
		},
		{
			name: "zero depolarization reduces to classic LK drive",
			solver: LKSolver{Alpha: 1.0, Beta: -0.3, Gamma: 0.07, K_dep: 0.0},
			P:      0.2,
			E:      -1.0,
			noise:  0.1,
			rhoEff: 1.3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Edep := tc.solver.K_dep * tc.P
			Eeff := tc.E - Edep
			dGdP := (2 * tc.solver.Alpha * tc.P) + (4 * tc.solver.Beta * math.Pow(tc.P, 3)) + (6 * tc.solver.Gamma * math.Pow(tc.P, 5))
			expected := (Eeff + tc.noise - dGdP) / tc.rhoEff

			got := tc.solver.dPdT(0, tc.P, tc.E, tc.noise, tc.rhoEff)
			if math.Abs(got-expected) > 1e-12 {
				t.Fatalf("dPdT mismatch: got %.12f, expected %.12f", got, expected)
			}
		})
	}
}

func TestLKSolver_effectiveRho_SignAndUnits_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		solver   LKSolver
		expected float64
	}{
		{
			name: "enabled uses rho + (R*A/d)",
			solver: LKSolver{
				Rho:                   0.1,
				UseEffectiveViscosity: true,
				SeriesResistance:      50,
				Area:                  2,
				Thickness:             5,
			},
			expected: 0.1 + (50*2)/5,
		},
		{
			name: "disabled keeps intrinsic rho only",
			solver: LKSolver{
				Rho:                   0.25,
				UseEffectiveViscosity: false,
				SeriesResistance:      50,
				Area:                  2,
				Thickness:             5,
			},
			expected: 0.25,
		},
		{
			name: "physical units example (ohm*meter contribution)",
			solver: LKSolver{
				Rho:                   0.05,
				UseEffectiveViscosity: true,
				SeriesResistance:      50.0,
				Area:                  45e-9 * 45e-9,
				Thickness:             10e-9,
			},
			expected: 0.05 + (50.0*(45e-9*45e-9))/(10e-9),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.solver.effectiveRho()
			if math.Abs(got-tc.expected) > 1e-12 {
				t.Fatalf("effectiveRho mismatch: got %.12f, expected %.12f", got, tc.expected)
			}
		})
	}
}

func TestLKSolver_effectiveRho_Disabled(t *testing.T) {
	s := &LKSolver{
		Rho:                   0.25,
		UseEffectiveViscosity: false,
		SeriesResistance:      50,
		Area:                  2,
		Thickness:             5,
	}

	got := s.effectiveRho()
	if math.Abs(got-s.Rho) > 1e-12 {
		t.Fatalf("expected rho without series contribution: got %.12f, expected %.12f", got, s.Rho)
	}
}

func TestLKSolver_SetState_Clamp(t *testing.T) {
	s := NewLKSolver()
	s.PMax = 0.5
	s.SetState(10)
	if math.Abs(s.GetState()) > s.PMax*1.2+1e-12 {
		t.Fatalf("expected polarization to be clamped within limit: got %.4f, limit %.4f", s.GetState(), s.PMax*1.2)
	}
}

func TestLKSolver_SetState_IgnoresInvalid(t *testing.T) {
	s := NewLKSolver()
	s.SetState(0.1)
	s.SetState(math.NaN())
	if math.IsNaN(s.GetState()) {
		t.Fatal("solver should ignore invalid polarization values")
	}
}

func TestLKSolver_SwitchesUnderStrongField(t *testing.T) {
	mat := LiteratureSuperlattice()
	if mat == nil {
		t.Fatal("expected material config")
	}
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.Temperature = 300
	s.EnableNoise = false
	s.UseNLS = false
	if !s.UseMaterialAlpha {
		s.UpdateParams()
	}
	s.SetState(0)

	dt := 1e-4
	E := 2.5 * mat.Ec
	start := s.GetState()
	for i := 0; i < 2000; i++ {
		s.Step(E, dt)
	}
	end := s.GetState()
	if math.Abs(end) <= math.Abs(start)+0.1 {
		t.Fatalf("expected polarization to move under strong field: start=%.4f end=%.4f", start, end)
	}
}
