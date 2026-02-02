package main

import (
	"fmt"

	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
)

func runMode(mode string) error {
	switch mode {
	case "hysteresis":
		return runHysteresisMode()
	default:
		return fmt.Errorf("unknown mode %q (expected: hysteresis)", mode)
	}
}

func runHysteresisMode() error {
	log := logging.NewLogger("hysteresis-mode")

	mat := physics.FeCIMMaterial()

	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false
	solver.UpdateParams()

	log.Info("LK config: Beta=%.3e Gamma=%.3e Rho=%.3e K_dep=%.3e Q12=%.3e Stress=%.2f GPa SeriesR=%.1f Ohm Thickness=%.2e m Area=%.2e m^2 CurieTemp=%.1fK CurieConst=%.2e UseEffVisc=%v UseNLS=%v Noise=%v",
		solver.Beta, solver.Gamma, solver.Rho, solver.K_dep, solver.Q12, solver.Stress/1e9, solver.SeriesResistance, solver.Thickness, solver.Area,
		solver.CurieTemp, solver.CurieConst, solver.UseEffectiveViscosity, solver.UseNLS, solver.EnableNoise)
	log.Info("LK alpha(T,σ)=%.3e at T=%.1fK", solver.Alpha, solver.Temperature)

	Emax := mat.Ec * 1.2
	dt := 1e-12

	log.Info("Landau-Khalatnikov diagnostic sweep starting")
	for _, E := range []float64{-Emax, -0.5 * Emax, 0, 0.5 * Emax, Emax} {
		solver.Step(E, dt)
	}

	log.Info("ISPP write-verify sequence starting")
	writer := physics.NewWriteController(solver, mat)
	writer.MaxIterations = 15
	writer.Tolerance = 1e-6
	writer.MaxVoltage = mat.Ec * mat.Thickness * 2.5
	writer.PulseWidth = mat.Tau

	gWindow := mat.Gmax - mat.Gmin
	targets := []float64{
		mat.Gmin + 0.55*gWindow,
		mat.Gmin + 0.75*gWindow,
		mat.Gmin + 0.90*gWindow,
	}

	for i, targetG := range targets {
		reset := i == 0
		attempts, success, overshoots := writer.WriteTargetWithReset(targetG, reset)
		finalP := solver.GetState()
		finalG := physics.PolarizationToConductance(finalP, mat.Ps, mat.Gmin, mat.Gmax)
		log.Info("ISPP step %d: targetG=%.3e reset=%v attempts=%d success=%v overshoots=%d finalP=%.3e finalG=%.3e",
			i+1, targetG, reset, attempts, success, overshoots, finalP, finalG)
	}

	return nil
}
