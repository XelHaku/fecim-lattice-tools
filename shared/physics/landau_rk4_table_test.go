package physics

import (
	"math"
	"testing"
)

func TestLKSolverStep_RK4_TableDriven(t *testing.T) {
	t.Parallel()

	t.Run("LinearConstantDerivative_RK4IsExact", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			P0    float64
			E     float64
			rho   float64
			dt    float64
			steps int
		}{
			{name: "PositiveField", P0: 0.1, E: 3.4, rho: 2.0, dt: 5e-9, steps: 50},
			{name: "NegativeField", P0: -1.0, E: -9.0, rho: 0.25, dt: 1e-12, steps: 1000},
			{name: "MixedSign", P0: 2.0, E: -1.0, rho: 5.0, dt: 1e-6, steps: 3},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				// Arrange a strictly linear ODE: dP/dt = E/rho.
				s := NewLKSolver()
				s.UseNLS = false
				s.EnableNoise = false
				s.UseEffectiveViscosity = false
				s.UseMaterialAlpha = true
				s.Alpha = 0
				s.Beta = 0
				s.Gamma = 0
				s.K_dep = 0
				s.PMax = 0 // disable clamping
				s.Rho = tt.rho
				s.SetState(tt.P0)

				for i := 0; i < tt.steps; i++ {
					got := s.Step(tt.E, tt.dt)
					if math.IsNaN(got) || math.IsInf(got, 0) {
						t.Fatalf("invalid P at step %d: %g", i, got)
					}
				}

				expected := tt.P0 + float64(tt.steps)*tt.dt*(tt.E/tt.rho)
				got := s.GetState()

				// This case should be exact to floating-point rounding.
				if diff := math.Abs(got - expected); diff > 1e-12*math.Max(1, math.Abs(expected)) {
					t.Fatalf("unexpected RK4 result: got=%0.15g expected=%0.15g diff=%g", got, expected, diff)
				}
			})
		}
	})

	t.Run("LinearRelaxation_MatchesAnalyticSolution", func(t *testing.T) {
		t.Parallel()

		// ODE when Beta=Gamma=Kdep=0: dP/dt = (E - 2*Alpha*P)/rho.
		// Analytic: P(t) = E/(2α) + (P0 - E/(2α)) * exp(-(2α/ρ)t)
		tests := []struct {
			name  string
			P0    float64
			E     float64
			alpha float64
			rho   float64
			dt    float64
			steps int
			relTol float64
		}{
			{name: "Mild", P0: 0.3, E: 1.2, alpha: 2.0, rho: 3.0, dt: 1e-3, steps: 1000, relTol: 3e-6},
			{name: "Stiffer", P0: -0.2, E: 5.0, alpha: 150.0, rho: 1.0, dt: 1e-5, steps: 5000, relTol: 1e-5},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				s := NewLKSolver()
				s.UseNLS = false
				s.EnableNoise = false
				s.UseEffectiveViscosity = false
				s.UseMaterialAlpha = true
				s.Alpha = tt.alpha
				s.Beta = 0
				s.Gamma = 0
				s.K_dep = 0
				s.PMax = 0
				s.Rho = tt.rho
				s.SetState(tt.P0)

				T := float64(tt.steps) * tt.dt
				for i := 0; i < tt.steps; i++ {
					s.Step(tt.E, tt.dt)
				}

				pEq := tt.E / (2 * tt.alpha)
				expected := pEq + (tt.P0-pEq)*math.Exp(-(2*tt.alpha/tt.rho)*T)
				got := s.GetState()

				absTol := tt.relTol * math.Max(1, math.Abs(expected))
				if diff := math.Abs(got - expected); diff > absTol {
					t.Fatalf("analytic mismatch: got=%g expected=%g diff=%g absTol=%g", got, expected, diff, absTol)
				}
			})
		}
	})

	t.Run("ImplicitBranch_BackwardEulerForLinearODE", func(t *testing.T) {
		t.Parallel()

		// This is a *regression* test for the stiff-step fallback.
		// Configure linear ODE and choose dt to force stiffness > threshold.
		const (
			P0    = 0.25
			E     = 1.5
			alpha = 1.0e9
			rho   = 1.0
			dt    = 1.0e-9
		)

		s := NewLKSolver()
		s.UseNLS = false
		s.EnableNoise = false
		s.UseEffectiveViscosity = false
		s.UseMaterialAlpha = true
		s.Alpha = alpha
		s.Beta = 0
		s.Gamma = 0
		s.K_dep = 0
		s.PMax = 0
		s.Rho = rho
		s.SetState(P0)

		got := s.Step(E, dt)
		if math.IsNaN(got) || math.IsInf(got, 0) {
			t.Fatalf("invalid P after stiff step: %g", got)
		}

		lambda := (2 * alpha) / rho
		expected := (P0 + dt*(E/rho)) / (1 + dt*lambda)

		// The implicit Newton solve should converge to the backward-Euler fixed point.
		if diff := math.Abs(got - expected); diff > 1e-12*math.Max(1, math.Abs(expected)) {
			t.Fatalf("backward-euler mismatch: got=%0.15g expected=%0.15g diff=%g", got, expected, diff)
		}
	})

	t.Run("NaNState_IsRecoveredAndClamped", func(t *testing.T) {
		t.Parallel()

		s := NewLKSolver()
		s.UseNLS = false
		s.EnableNoise = false
		s.UseEffectiveViscosity = false
		s.UseMaterialAlpha = true
		s.Alpha = 0
		s.Beta = 0
		s.Gamma = 0
		s.K_dep = 0
		s.PMax = 0.30
		s.Rho = 1.0
		s.SetState(math.NaN())

		got := s.Step(0, 1e-12)
		if math.IsNaN(got) || math.IsInf(got, 0) {
			t.Fatalf("expected recovered finite P, got %g", got)
		}
		if got != -math.Abs(s.PMax) {
			t.Fatalf("expected NaN recovery to reset P to -PMax: got=%g PMax=%g", got, s.PMax)
		}
		if math.Abs(got) > 1.2*s.PMax {
			t.Fatalf("expected clamped state within 1.2*PMax, got=%g PMax=%g", got, s.PMax)
		}
	})
}

func TestLKSolverStep_ZeroDT_IsNoOp(t *testing.T) {
	t.Parallel()

	s := NewLKSolver()
	s.UseNLS = false
	s.EnableNoise = false
	s.UseEffectiveViscosity = false
	s.UseMaterialAlpha = true
	s.Alpha = 0
	s.Beta = 0
	s.Gamma = 0
	s.K_dep = 0
	s.PMax = 0
	s.Rho = 1
	s.SetState(0.123)
	s.Time = 7

	got := s.Step(99, 0)
	if got != 0.123 {
		t.Fatalf("dt=0 should not change state: got=%g want=%g", got, 0.123)
	}
	if s.Time != 7 {
		t.Fatalf("dt=0 should not change time: got=%v want=%v", s.Time, 7)
	}
}
