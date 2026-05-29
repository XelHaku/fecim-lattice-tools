package circuits

import (
	"fecim-lattice-tools/shared/mathutil"
	"fecim-lattice-tools/shared/physics"
)

type isppSimulationWorkflow struct {
	state CircuitsState
}

func newISPPSimulationWorkflow(state CircuitsState) isppSimulationWorkflow {
	return isppSimulationWorkflow{state: state}
}

func (w isppSimulationWorkflow) compute() CircuitsState {
	mat := physics.DefaultHZO()
	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	ctrl := physics.NewWriteController(solver, mat)

	numLevels := w.state.QuantLevels
	if numLevels <= 0 {
		numLevels = DefaultQuantLevels
		w.state.QuantLevels = numLevels
	}
	w.state.ISPPAttempts = make([]int, numLevels)
	w.state.ISPPConverged = make([]bool, numLevels)

	successCount := 0
	totalAttempts := 0
	for level := 0; level < numLevels; level++ {
		targetG := mathutil.LerpByIndex(level, numLevels, mat.Gmin, mat.Gmax)
		attempts, success, _ := ctrl.WriteTarget(targetG)
		w.state.ISPPAttempts[level] = attempts
		w.state.ISPPConverged[level] = success
		if success {
			successCount++
		}
		totalAttempts += attempts
	}

	w.state.ISPPTotalAttempts = totalAttempts
	w.state.ISPPConvergedCount = successCount
	if totalAttempts > 0 {
		w.state.ISPPAvgAttempts = float64(totalAttempts) / float64(numLevels)
	}
	w.state.ISPPExecuted = true
	return w.state
}
