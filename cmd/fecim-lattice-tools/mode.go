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

	mat := physics.DefaultHZO()

	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false

	Emax := mat.Ec * 1.2
	dt := 1e-9

	log.Info("Landau-Khalatnikov diagnostic sweep starting")
	for _, E := range []float64{-Emax, -0.5 * Emax, 0, 0.5 * Emax, Emax} {
		solver.Step(E, dt)
	}

	log.Info("ISPP write-verify sequence starting")
	writer := physics.NewWriteController(solver, mat)
	writer.MaxIterations = 12
	writer.Tolerance = 1e-6
	writer.MaxVoltage = 1.0
	writer.PulseWidth = mat.Tau * 0.1
	targetG := mat.Gmin + 0.85*(mat.Gmax-mat.Gmin)

	attempts, success, overshoots := writer.WriteTarget(targetG)
	log.Info("ISPP summary: attempts=%d success=%v overshoots=%d", attempts, success, overshoots)

	return nil
}
