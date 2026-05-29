package mnist

type quantizationSweepWorkflow struct {
	state MNISTState
}

func newQuantizationSweepWorkflow(state MNISTState) quantizationSweepWorkflow {
	return quantizationSweepWorkflow{state: state}
}

func (w quantizationSweepWorkflow) compute() MNISTState {
	levels := []int{2, 4, 8, 16, 32, 64, 128}
	w.state.SweepLevels = levels
	w.state.SweepAccuracy = make([]float64, len(levels))
	for i, level := range levels {
		w.state.SweepAccuracy[i] = accuracyForLevels(level)
	}
	return w.state
}
