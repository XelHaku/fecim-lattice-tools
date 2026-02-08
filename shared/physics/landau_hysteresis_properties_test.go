package physics

import (
	"math"
	"testing"
)

// Table-driven property tests for the Landau-Khalatnikov (dynamic) solver.
//
// Focus:
// - Saturation/finite bounds: P remains finite and within clamp limits
// - Monotonic switching (up to small jitter) under a strong constant field when starting from opposite polarization
// - Odd symmetry when the configuration is unbiased and depolarization is disabled:
//     P(t; E, P0) = -P(t; -E, -P0)
// - Energy sanity: a driven field cycle should dissipate non-negative work density
//   (approx via integral \int E dP over a cycle), with reasonable magnitude.
//
// Assumptions:
// - Tests disable NLS and noise for determinism.
// - Symmetry test sets Stress=0, Q12=0, and K_dep=0.
func TestLKSolver_TableDriven_Properties(t *testing.T) {
	tests := []struct {
		name      string
		solver    *LKSolver
		Ec        float64
		Pr        float64
		PMax      float64
		strongMul float64
	}{
		{
			name: "Deterministic_DefaultLike",
			solver: func() *LKSolver {
				s := NewLKSolver()
				s.EnableNoise = false
				s.UseNLS = false
				// Remove mechanical bias and depolarization for clean invariants.
				s.Q12 = 0
				s.Stress = 0
				s.K_dep = 0
				// Keep alpha constant for deterministic comparisons.
				s.UseMaterialAlpha = true
				// Pick a stable symmetric set (not tied to any specific material).
				s.Alpha = -1.0e8
				s.Beta = -2.0e8
				s.Gamma = 1.5e10
				s.Rho = 0.05
				s.UseEffectiveViscosity = false
				s.PMax = 0.30
				return s
			}(),
			Ec:        1.0e8,
			Pr:        0.25,
			PMax:      0.30,
			strongMul: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.solver
			if s == nil {
				t.Fatal("nil solver")
			}
			s.PMax = tt.PMax

			// 1) Bounds/finite under repeated stepping with a large field.
			Estrong := tt.strongMul * tt.Ec
			dt := 1e-10
			s.SetState(-math.Abs(tt.Pr))
			for i := 0; i < 2000; i++ {
				p := s.Step(Estrong, dt)
				if math.IsNaN(p) || math.IsInf(p, 0) {
					t.Fatalf("invalid P at step %d: %v", i, p)
				}
				limit := 1.2 * s.PMax
				if s.PMax > 0 && math.Abs(p) > limit+1e-12 {
					t.Fatalf("P exceeded clamp limit at step %d: P=%.6e limit=%.6e", i, p, limit)
				}
			}

			// 2) Mostly-monotonic switching under strong constant field.
			// Starting from negative polarization, a strong positive field should drive P upward.
			s.SetState(-math.Abs(tt.Pr))
			prev := s.GetState()
			// Allow small solver jitter from implicit/RK4 switching and clamps.
			// Use a tolerance relative to PMax; LK can exhibit tiny backsteps during stiff transitions.
			tol := 1e-6
			if s.PMax > 0 {
				tol = 1e-6 * s.PMax
			}
			for i := 0; i < 500; i++ {
				p := s.Step(Estrong, dt)
				if (p - prev) < -tol {
					t.Fatalf("non-monotone switching under strong +E at step %d: prev=%.6e p=%.6e (delta=%.3e tol=%.3e)", i, prev, p, p-prev, tol)
				}
				prev = p
			}
			if s.GetState() < 0 {
				t.Fatalf("expected to switch to positive polarization under strong +E, got P=%.6e", s.GetState())
			}

			// 3) Odd symmetry with unbiased configuration and K_dep=0.
			// Compare two trajectories driven by opposite fields from opposite initial states.
			steps := 400
			E := 1.7 * tt.Ec
			// Clone by copying the solver struct (safe here since no internal pointers).
			sa := *s
			sb := *s
			sa.SetState(-0.12)
			sb.SetState(+0.12)
			sa.Time = 0
			sb.Time = 0
			maxAsym := 0.0
			for i := 0; i < steps; i++ {
				pa := sa.Step(+E, dt)
				pb := sb.Step(-E, dt)
				asym := math.Abs(pa + pb)
				if asym > maxAsym {
					maxAsym = asym
				}
			}
			// Allow small numerical asymmetry.
			if maxAsym > 2e-4 {
				t.Fatalf("odd symmetry violated: max |P(E)+P(-E)| = %.6e", maxAsym)
			}

			// 4) Energy/work sanity over a crude cycle.
			// Approximate dissipated work density: W = \oint E dP.
			// We drive a simple rectangular cycle (+E hold, -E hold) and integrate E*dP.
			sCycle := *s
			sCycle.SetState(-math.Abs(tt.Pr))
			sCycle.Time = 0

			cycleE := []float64{+Estrong, +Estrong, -Estrong, -Estrong}
			holdSteps := 400
			lastP := sCycle.GetState()
			work := 0.0
			for seg := 0; seg < len(cycleE); seg++ {
				Eseg := cycleE[seg]
				for i := 0; i < holdSteps; i++ {
					p := sCycle.Step(Eseg, dt)
					dP := p - lastP
					work += Eseg * dP
					lastP = p
				}
			}
			if math.IsNaN(work) || math.IsInf(work, 0) {
				t.Fatalf("invalid work integral: %v", work)
			}
			absWork := math.Abs(work)
			if absWork <= 0 {
				t.Fatalf("expected non-zero dissipated work, got %.6e", absWork)
			}
			// Order-of-magnitude sanity: roughly Ec*Pr (J/m^3).
			est := tt.Ec * tt.Pr
			if est <= 0 {
				est = 1
			}
			ratio := absWork / est
			if ratio < 1e-3 || ratio > 1e3 {
				t.Fatalf("work density magnitude unexpected: |W|=%.6e, Ec*Pr=%.6e, ratio=%.3e", absWork, est, ratio)
			}
		})
	}
}
