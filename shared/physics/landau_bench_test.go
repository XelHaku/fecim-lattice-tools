package physics

import (
	"testing"
)

func BenchmarkLKSolverStep(b *testing.B) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.SetState(0)

	// Representative operating point: near coercive field and sub-ns time step.
	E := 0.8 * mat.Ec
	dt := 1e-12

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Step(E, dt)
	}
}

func BenchmarkLKSolverStep_StiffImplicitPath(b *testing.B) {
	mat := DefaultHZO()
	s := NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.SetState(0)

	// Larger dt tends to trigger the implicit step path.
	E := 1.2 * mat.Ec
	dt := 5e-10

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Step(E, dt)
	}
}
