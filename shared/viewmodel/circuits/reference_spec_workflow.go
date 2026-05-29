package circuits

import "math"

type referenceSpecWorkflow struct {
	state CircuitsState
}

func newReferenceSpecWorkflow(state CircuitsState) referenceSpecWorkflow {
	return referenceSpecWorkflow{state: state}
}

func (w referenceSpecWorkflow) compute() CircuitsState {
	quantLevels := w.state.QuantLevels
	if quantLevels <= 0 {
		quantLevels = DefaultQuantLevels
	}
	rows := maxInt(1, w.state.Rows)
	cols := maxInt(1, w.state.Cols)
	cells := rows * cols
	dacCodes := 1 << w.state.DACResolution
	adcCodes := 1 << w.state.ADCResolution

	const (
		arrayPowerMW       = 0.1
		controlPowerMW     = 0.5
		dacPowerPerColMW   = 0.1
		tiaPowerPerRowMW   = 0.05
		adcPowerPerRowMW   = 0.5
		referenceLatencyNS = 76.0
	)
	totalPowerMW := arrayPowerMW + controlPowerMW +
		dacPowerPerColMW*float64(cols) +
		tiaPowerPerRowMW*float64(rows) +
		adcPowerPerRowMW*float64(rows)
	throughputGOPS := float64(cells) / referenceLatencyNS
	efficiencyGOPSW := 0.0
	if totalPowerMW > 0 {
		efficiencyGOPSW = throughputGOPS * 1000 / totalPowerMW
	}

	w.state.SpecCells = cells
	w.state.SpecBitsPerCell = math.Log2(float64(quantLevels))
	w.state.SpecDACCount = cols
	w.state.SpecTIACount = rows
	w.state.SpecADCCount = rows
	w.state.SpecDACCodes = dacCodes
	w.state.SpecADCCodes = adcCodes
	w.state.SpecTotalPowerMW = totalPowerMW
	w.state.SpecLatencyNS = referenceLatencyNS
	w.state.SpecThroughputGOPS = throughputGOPS
	w.state.SpecEfficiencyGOPSW = efficiencyGOPSW
	w.state.SpecCompliance = referenceSpecCompliance(dacCodes, adcCodes, quantLevels)
	return w.state
}
