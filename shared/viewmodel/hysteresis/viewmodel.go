package hysteresis

import (
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct {
	state HysteresisState
}

func New() *Module {
	materials := physics.AllMaterials()
	defaultMat := "HZO Default"
	if len(materials) > 0 && materials[0] != nil {
		defaultMat = materials[0].Name
	}
	return &Module{
		state: HysteresisState{
			SelectedMaterial: defaultMat,
			Materials:        materials,
			FieldRange:       FieldRange{MinField: -3000, MaxField: 3000},
			Waveform:         "sine",
		},
	}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleHysteresis,
		Title:       "FeCIM Hysteresis Simulation",
		Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
