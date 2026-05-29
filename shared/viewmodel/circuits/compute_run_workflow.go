package circuits

import (
	"fmt"
	"math"

	"fecim-lattice-tools/shared/mathutil"
)

type computeRunWorkflow struct {
	state CircuitsState
}

func newComputeRunWorkflow(state CircuitsState) computeRunWorkflow {
	return computeRunWorkflow{state: state}
}

func (w computeRunWorkflow) buildLog() *ComputeRunLog {
	rows := maxInt(1, w.state.Rows)
	cols := maxInt(1, w.state.Cols)
	quantLevels := w.state.QuantLevels
	if quantLevels <= 1 {
		quantLevels = DefaultQuantLevels
	}
	input := deterministicComputeInputVector(cols)
	weights := make([][]int, rows)
	conductances := make([][]float64, rows)
	rowResults := make([]ComputeRowResult, rows)
	for r := 0; r < rows; r++ {
		weights[r] = make([]int, cols)
		conductances[r] = make([]float64, cols)
		cells := make([]ComputeCellContribution, cols)
		for c := 0; c < cols; c++ {
			weight := (r*cols + c) % quantLevels
			conductanceUS := conductanceForLevelUS(weight, quantLevels)
			currentUA := conductanceUS * input[c]
			weights[r][c] = weight
			conductances[r][c] = conductanceUS
			cells[c] = ComputeCellContribution{
				Col:           c,
				Weight:        weight,
				ConductanceUS: conductanceUS,
				VoltageV:      input[c],
				CurrentUA:     currentUA,
			}
		}
		rowCurrentUA, _ := mathutil.Dot(conductances[r], input)
		tiaVoltage := rowCurrentUA * 1e-6 * w.state.TIAGain
		senseVoltage := mathutil.Clamp(tiaVoltage, 0, w.state.SupplyVoltage)
		adcMax := float64((int(1) << uint(w.state.ADCResolution)) - 1)
		adcLevel := 0
		if w.state.SupplyVoltage > 0 && adcMax > 0 {
			adcLevel = int(math.Round(senseVoltage / w.state.SupplyVoltage * adcMax))
		}
		rowResults[r] = ComputeRowResult{
			Row:        r,
			Active:     true,
			CurrentUA:  rowCurrentUA,
			TIAVoltage: tiaVoltage,
			ADCLevel:   adcLevel,
			Saturated:  tiaVoltage > w.state.SupplyVoltage,
			CellDetail: cells,
		}
	}
	return &ComputeRunLog{
		Schema:        "fecim.circuits.compute_run.v1",
		ArraySize:     fmt.Sprintf("%dx%d", rows, cols),
		Material:      "HZO default educational preset",
		QuantLevels:   quantLevels,
		Architecture:  w.state.Architecture,
		CouplingTier:  w.state.CouplingTier,
		InputVector:   input,
		Weights:       weights,
		Conductances:  conductances,
		RowResults:    rowResults,
		ExportedCells: rows * cols,
	}
}
