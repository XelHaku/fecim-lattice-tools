package physics_test

import (
	"math"
	"testing"

	crossbar "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/shared/physics"
)

// Published HZO temperature coefficients used in this repository baseline:
//
//	dEc/dT ≈ -2.0e5 V/m/K
//	dPr/dT ≈ -5.0e-5 C/m²/K
//
// These are encoded in DefaultHZO() from the HZO literature calibration.
func TestTemperatureDependence_PreisachEc_DecreasesWithTemperature(t *testing.T) {
	mat := physics.DefaultHZO()
	temps := []float64{300, 350, 400}

	Ec300 := mat.Ec + mat.TempCoeffEc*(temps[0]-300)
	Ec350 := mat.Ec + mat.TempCoeffEc*(temps[1]-300)
	Ec400 := mat.Ec + mat.TempCoeffEc*(temps[2]-300)

	if !(Ec300 > Ec350 && Ec350 > Ec400) {
		t.Fatalf("expected Ec to decrease with temperature: Ec300=%.3e, Ec350=%.3e, Ec400=%.3e", Ec300, Ec350, Ec400)
	}

	// Validate coefficient is in a published HZO range (~1.5e5 to 2.5e5 V/m/K magnitude).
	if math.Abs(mat.TempCoeffEc) < 1.5e5 || math.Abs(mat.TempCoeffEc) > 2.5e5 {
		t.Fatalf("HZO TempCoeffEc out of expected literature range: got %.3e V/m/K", mat.TempCoeffEc)
	}
}

func TestTemperatureDependence_PreisachPr_DecreasesWithTemperature(t *testing.T) {
	mat := physics.DefaultHZO()
	temps := []float64{300, 350, 400}

	Pr300 := mat.Pr + mat.TempCoeffPr*(temps[0]-300)
	Pr350 := mat.Pr + mat.TempCoeffPr*(temps[1]-300)
	Pr400 := mat.Pr + mat.TempCoeffPr*(temps[2]-300)

	if !(Pr300 > Pr350 && Pr350 > Pr400) {
		t.Fatalf("expected Pr to decrease with temperature: Pr300=%.6f, Pr350=%.6f, Pr400=%.6f C/m²", Pr300, Pr350, Pr400)
	}

	// Validate coefficient is in a published HZO range (~3e-5 to 7e-5 C/m²/K magnitude).
	if math.Abs(mat.TempCoeffPr) < 3e-5 || math.Abs(mat.TempCoeffPr) > 7e-5 {
		t.Fatalf("HZO TempCoeffPr out of expected literature range: got %.3e C/m²/K", mat.TempCoeffPr)
	}
}

func TestTemperatureDependence_LKSolverSwitchingFasterAtHigherTemperature(t *testing.T) {
	mat := physics.DefaultHZO()

	switchTime := func(tempK float64) float64 {
		s := physics.NewLKSolver()
		s.ConfigureFromMaterial(mat)
		s.UseMaterialAlpha = false // ensure alpha(T) coupling is active
		s.UseNLS = false           // deterministic dynamics for this test
		s.EnableNoise = false
		s.Temperature = tempK
		s.UpdateParams()
		s.SetState(-mat.Pr)

		const (
			EField   = 1.0e8 // V/m, near Ec so temperature dependence is observable
			dt       = 1e-11 // 10 ps
			maxSteps = 100000
		)

		for i := 1; i <= maxSteps; i++ {
			p := s.Step(EField, dt)
			if p >= 0 {
				return float64(i) * dt
			}
		}

		t.Fatalf("did not switch within %.3e s at T=%.1fK", float64(maxSteps)*dt, tempK)
		return math.NaN()
	}

	t300 := switchTime(300)
	t400 := switchTime(400)

	if !(t400 < t300) {
		t.Fatalf("expected faster LK switching at higher temperature: t300=%.3e s, t400=%.3e s", t300, t400)
	}
}

func TestTemperatureDependence_CrossbarWireResistanceIncreasesWithTemperature(t *testing.T) {
	const r0 = 10.0 // ohms at 300K baseline

	r300 := crossbar.NewTemperatureEffects(300).AdjustedWireResistance(r0)
	r350 := crossbar.NewTemperatureEffects(350).AdjustedWireResistance(r0)
	r400 := crossbar.NewTemperatureEffects(400).AdjustedWireResistance(r0)

	if !(r300 < r350 && r350 < r400) {
		t.Fatalf("expected wire resistance to increase with temperature: R300=%.6fΩ, R350=%.6fΩ, R400=%.6fΩ", r300, r350, r400)
	}

	// Copper TCR check: R(T)=R0*(1+0.00393*(T-300K)).
	expected400 := r0 * (1 + 0.00393*(400-300))
	if math.Abs(r400-expected400) > 1e-9 {
		t.Fatalf("wire resistance coefficient mismatch at 400K: got %.9fΩ, expected %.9fΩ", r400, expected400)
	}
}
