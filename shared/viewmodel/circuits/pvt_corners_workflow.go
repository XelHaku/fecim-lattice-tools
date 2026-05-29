package circuits

import (
	"math"

	"fecim-lattice-tools/shared/physics"
)

type pvtCornersWorkflow struct {
	state CircuitsState
}

func newPVTCornersWorkflow(state CircuitsState) pvtCornersWorkflow {
	return pvtCornersWorkflow{state: state}
}

func (w pvtCornersWorkflow) compute() CircuitsState {
	mat := physics.DefaultHZO()
	vref := w.state.SupplyVoltage
	bits := w.state.ADCResolution
	lsb := vref / float64(int(1)<<bits)

	enobForINL := func(inlLSB float64) float64 {
		return math.Max(float64(bits)-math.Log2(inlLSB+1.0), 1.0)
	}
	w.state.ENOBtt = enobForINL(0.5)
	w.state.ENOBff = enobForINL(0.5 * 0.80)
	w.state.ENOBss = enobForINL(0.5 * 1.25)
	w.state.ADCNoiseLSB = math.Sqrt(lsb * lsb / 12.0)
	w.state.SNRdB = 6.02*float64(bits) + 1.76
	w.state.PVTTemperatureSweep = pvtTemperatureSweepStatus(mat)
	w.state.PVTProcessYield, w.state.PVTPassSamples, w.state.PVTSamples = pvtProcessYield(mat)
	w.state.PVTENOBNoiseCeiling, w.state.PVTENOBCeilingBits = pvtNoiseLimitedENOBCeiling(w.state.TIAGain)

	return w.state
}
