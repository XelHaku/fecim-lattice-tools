package crossbar

import (
	"fmt"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
	"fecim-lattice-tools/shared/viewmodel/design"
)

type Module struct {
	state CrossbarState
}

func New(rows, cols int) *Module {
	m := &Module{state: CrossbarState{Rows: rows, Cols: cols}}
	m.reallocate(rows, cols)
	return m
}

func (m *Module) reallocate(rows, cols int) {
	m.state.Rows = rows
	m.state.Cols = cols
	m.state.Conductances = make([][]float64, rows)
	m.state.InputVector = make([]float64, cols)
	m.state.OutputVector = make([]float64, rows)
	for i := range rows {
		m.state.Conductances[i] = make([]float64, cols)
		for j := range cols {
			m.state.Conductances[i][j] = physics.QuantizeTo30Levels(50.0 * float64(1+((i*cols+j)%3)))
		}
	}
	for j := range cols {
		m.state.InputVector[j] = 1.0
	}
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

func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "resize":
		rows, err := viewmodel.OptionalPayloadInt(action.Payload, "rows", m.state.Rows)
		if err != nil {
			return fmt.Errorf("crossbar: %w", err)
		}
		cols, err := viewmodel.OptionalPayloadInt(action.Payload, "cols", m.state.Cols)
		if err != nil {
			return fmt.Errorf("crossbar: %w", err)
		}
		if rows > 0 && cols > 0 && rows <= 128 && cols <= 128 {
			m.reallocate(rows, cols)
			return nil
		}
		return fmt.Errorf("crossbar: invalid dimensions %d×%d", rows, cols)
	case "run_mvm":
		m.runMVM()
		return nil
	case "toggle_ir":
		m.state.SneakPaths = !m.state.SneakPaths
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}

func (m *Module) runMVM() {
	m.state = newMVMWorkflow(m.state).compute()
}

func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) DesignState() design.ModuleDesignState {
	return design.ModuleDesignState{ArrayRows: m.state.Rows, ArrayCols: m.state.Cols}
}
