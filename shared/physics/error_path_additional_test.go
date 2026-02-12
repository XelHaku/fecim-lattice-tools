package physics

import (
	"math"
	"testing"
)

type linearEverett struct{}

func (linearEverett) Calculate(alpha, beta float64) float64 {
	if alpha < beta {
		return 0
	}
	return alpha - beta
}

func TestPreisachStack_HandlesInvalidInputsGracefully(t *testing.T) {
	ps := NewPreisachStack(1.0, linearEverett{})
	_ = ps.Update(0.25)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Preisach.Update panicked on invalid input: %v", r)
		}
	}()

	_ = ps.Update(math.NaN())
	_ = ps.Update(math.Inf(1))
}

func TestLKSolver_InvalidMaterialParamsHandledGracefully(t *testing.T) {
	base := NewLKSolver()
	baseThickness := base.Thickness
	baseArea := base.Area

	invalid := DefaultHZO()
	invalid.Thickness = -10e-9 // invalid; should be ignored
	invalid.Area = -1          // invalid; should be ignored
	invalid.Ec = 0             // invalid for Ec-derived scaling path

	base.ConfigureFromMaterial(invalid)

	if base.Thickness != baseThickness {
		t.Fatalf("Thickness changed from valid default to invalid value: got %e want %e", base.Thickness, baseThickness)
	}
	if base.Area != baseArea {
		t.Fatalf("Area changed from valid default to invalid value: got %e want %e", base.Area, baseArea)
	}

	p := base.Step(1e8, 1e-11)
	if math.IsNaN(p) || math.IsInf(p, 0) {
		t.Fatalf("Step after invalid material config returned non-finite polarization: %v", p)
	}
}

func TestISPPCalculator_EdgeTargetsAndAlreadyAtTarget(t *testing.T) {
	calc := NewISPPCalculator(1.0, 30)

	t.Run("target zero", func(t *testing.T) {
		got := calc.CheckResult(0, 0, DirectionAscending, 0)
		if got != ISPPSuccess {
			t.Fatalf("CheckResult(0->0) = %v, want %v", got, ISPPSuccess)
		}
	})

	t.Run("target max", func(t *testing.T) {
		maxLevel := calc.NumLevels - 1
		got := calc.CheckResult(maxLevel, maxLevel, DirectionAscending, 1)
		if got != ISPPSuccess {
			t.Fatalf("CheckResult(max->max) = %v, want %v", got, ISPPSuccess)
		}
	})

	t.Run("already at target with unknown direction", func(t *testing.T) {
		got := calc.CheckResult(12, 12, DirectionUnknown, 3)
		if got != ISPPSuccess {
			t.Fatalf("CheckResult(already-at-target) = %v, want %v", got, ISPPSuccess)
		}
	})
}
