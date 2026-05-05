package circuits

import (
	"math"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct{ state CircuitsState }

func New() *Module {
	m := &Module{state: CircuitsState{
		ADCResolution: 5, DACResolution: 5, TIAGain: 1e4,
		ChargePumpStages: 4, SupplyVoltage: 1.8, ISPPEnabled: true,
	}}
	m.runISPPSimulation()
	return m
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleCircuits,
		Title:       "FeCIM Peripheral Circuits Visualizer",
		Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }

func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "run_read":
		return nil
	case "run_write":
		m.runISPPSimulation()
		return nil
	case "toggle_ispp":
		m.state.ISPPEnabled = !m.state.ISPPEnabled
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) runISPPSimulation() {
	mat := physics.DefaultHZO()
	solver := physics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)

	ctrl := physics.NewWriteController(solver, mat)

	numLevels := 30
	m.state.ISPPAttempts = make([]int, numLevels)
	m.state.ISPPConverged = make([]bool, numLevels)

	successCount := 0
	totalAttempts := 0
	for level := 0; level < numLevels; level++ {
		targetG := mat.Gmin + (mat.Gmax-mat.Gmin)*float64(level)/float64(numLevels-1)
		attempts, success, _ := ctrl.WriteTarget(targetG)
		m.state.ISPPAttempts[level] = attempts
		m.state.ISPPConverged[level] = success
		if success {
			successCount++
		}
		totalAttempts += attempts
	}

	m.state.ISPPTotalAttempts = totalAttempts
	m.state.ISPPConvergedCount = successCount
	if totalAttempts > 0 {
		m.state.ISPPAvgAttempts = float64(totalAttempts) / float64(numLevels)
	}
	m.state.ISPPExecuted = true

	m.computePVTCorners()
}

func (m *Module) computePVTCorners() {
	vref := m.state.SupplyVoltage
	bits := m.state.ADCResolution
	lsb := vref / float64(int(1)<<bits)

	enobForINL := func(inlLSB float64) float64 {
		return math.Max(float64(bits)-math.Log2(inlLSB+1.0), 1.0)
	}
	m.state.ENOBtt = enobForINL(0.5)
	m.state.ENOBff = enobForINL(0.5 * 0.80)
	m.state.ENOBss = enobForINL(0.5 * 1.25)
	m.state.ADCNoiseLSB = math.Sqrt(lsb * lsb / 12.0)
	m.state.SNRdB = 6.02*float64(bits) + 1.76

	_ = lsb
	_ = vref
}
