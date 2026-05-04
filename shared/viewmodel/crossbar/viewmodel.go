package crossbar

import (
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct {
	state CrossbarState
}

func New(rows, cols int) *Module {
	state := CrossbarState{Rows: rows, Cols: cols}
	state.Conductances = make([][]float64, rows)
	state.InputVector = make([]float64, cols)
	state.OutputVector = make([]float64, rows)
	for i := range state.Conductances {
		state.Conductances[i] = make([]float64, cols)
		for j := range state.Conductances[i] {
			state.Conductances[i][j] = physics.QuantizeTo30Levels(50.0)
		}
	}
	for j := range state.InputVector {
		state.InputVector[j] = 1.0
	}
	return &Module{state: state}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleCrossbar,
		Title:       "FeCIM Crossbar Array Visualization",
		Description: "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
