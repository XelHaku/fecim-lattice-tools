package hysteresis

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state HysteresisState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "material", Label: "Material", Value: state.SelectedMaterial},
		{ID: "field_min", Label: "Min Field", Value: fmt.Sprintf("%.0f kV/cm", state.FieldRange.MinField)},
		{ID: "field_max", Label: "Max Field", Value: fmt.Sprintf("%.0f kV/cm", state.FieldRange.MaxField)},
		{ID: "waveform", Label: "Waveform", Value: state.Waveform},
	}

	sections := []viewmodel.Section{}
	for _, mat := range state.Materials {
		if mat == nil {
			continue
		}
		sections = append(sections, viewmodel.Section{
			ID:    "material_" + mat.Name,
			Title: mat.Name,
			Body:  materialSummary(mat),
		})
	}

	actions := []viewmodel.Action{
		{ID: EventSelectMaterial, Label: "Change Material", Kind: viewmodel.ActionSelect},
		{ID: EventSetFieldRange, Label: "Set Field Range", Kind: viewmodel.ActionCommand},
		{ID: EventToggleSimulation, Label: "Run/Pause", Kind: viewmodel.ActionToggle},
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID:          viewmodel.ModuleHysteresis,
			Title:       "FeCIM Hysteresis Simulation",
			Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
			Status:      viewmodel.StatusFunctional,
		},
		Metrics:  metrics,
		Sections: sections,
		Actions:  actions,
	}
}
