package crossbar

import "fecim-lattice-tools/shared/mathutil"

type mvmWorkflow struct {
	state CrossbarState
}

func newMVMWorkflow(state CrossbarState) mvmWorkflow {
	return mvmWorkflow{state: state}
}

func (w mvmWorkflow) compute() CrossbarState {
	output, err := mathutil.MatrixVectorMultiply(w.state.Conductances, w.state.InputVector)
	if err != nil {
		return w.state
	}
	copy(w.state.OutputVector, output)
	return w.state
}
