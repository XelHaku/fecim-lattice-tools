package circuits

type referenceTimingWorkflow struct {
	state CircuitsState
}

func newReferenceTimingWorkflow(state CircuitsState) referenceTimingWorkflow {
	return referenceTimingWorkflow{state: state}
}

func (w referenceTimingWorkflow) compute() CircuitsState {
	const (
		writeTotalNS   = 203
		readTotalNS    = 76
		computeTotalNS = 76
	)
	w.state.TimingWriteTotalNS = writeTotalNS
	w.state.TimingReadTotalNS = readTotalNS
	w.state.TimingComputeTotalNS = computeTotalNS
	w.state.TimingWaveforms = referenceTimingWaveforms()

	switch timingOperationValue(w.state) {
	case "WRITE":
		w.state.TimingActiveOp = "WRITE"
		w.state.TimingActiveTotalNS = writeTotalNS
		w.state.TimingActivePhases = "DAC 10 / Pump 88 / Pulse 100 / Array 5 ns"
	case "COMPUTE":
		w.state.TimingActiveOp = "COMPUTE"
		w.state.TimingActiveTotalNS = computeTotalNS
		w.state.TimingActivePhases = "DAC 10 / Array 5 / TIA+ADC 61 ns"
	default:
		w.state.TimingActiveOp = "READ"
		w.state.TimingActiveTotalNS = readTotalNS
		w.state.TimingActivePhases = "DAC 10 / Array 5 / TIA 11 / ADC 50 ns"
	}
	return w.state
}
