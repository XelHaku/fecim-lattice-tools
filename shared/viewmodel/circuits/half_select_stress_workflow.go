package circuits

import "math"

type halfSelectStressWorkflow struct {
	state CircuitsState
}

func newHalfSelectStressWorkflow(state CircuitsState) halfSelectStressWorkflow {
	return halfSelectStressWorkflow{state: state}
}

func (w halfSelectStressWorkflow) compute() CircuitsState {
	w.state.HalfSelectState = HalfSelectStateInactive
	w.state.HalfSelectCells = 0
	w.state.DisturbVoltage = 0
	w.state.StressPerPulse = 0
	w.state.StressCyclesToLevel = 0

	if w.state.OperationMode != OperationWrite {
		return w.state
	}

	switch w.state.Architecture {
	case ArchitecturePassive:
		w.state.HalfSelectState = HalfSelectStateColumnWriteActive
		w.state.HalfSelectCells = maxInt(0, w.state.Rows-1)
		w.state.DisturbVoltage = DefaultDisturbVoltage
		w.state.StressPerPulse = PassiveStressPerPulse
	case Architecture1T1R:
		w.state.HalfSelectState = HalfSelectStateAttenuated
		w.state.HalfSelectCells = maxInt(0, w.state.Rows-1)
		w.state.DisturbVoltage = DefaultDisturbVoltage / OneTOneRStressAttenuation
		w.state.StressPerPulse = PassiveStressPerPulse / OneTOneRStressAttenuation
	case Architecture2T1R:
		w.state.HalfSelectState = HalfSelectStateIsolated
		return w.state
	default:
		return w.state
	}
	if w.state.StressPerPulse > 0 {
		w.state.StressCyclesToLevel = int(math.Ceil(1.0 / w.state.StressPerPulse))
	}
	return w.state
}
