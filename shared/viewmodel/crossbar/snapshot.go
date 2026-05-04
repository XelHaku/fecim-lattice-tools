package crossbar

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state CrossbarState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "rows", Label: "Rows", Value: fmt.Sprintf("%d", state.Rows)},
		{ID: "cols", Label: "Columns", Value: fmt.Sprintf("%d", state.Cols)},
		{ID: "ir_drop", Label: "IR Drop", Value: fmt.Sprintf("%.1f%%", state.IRDrop*100)},
		{ID: "drift", Label: "Drift Factor", Value: fmt.Sprintf("%.2f", state.DriftFactor)},
	}

	sections := []viewmodel.Section{
		{
			ID:    "ir_drop",
			Title: "IR Drop",
			Body:  fmt.Sprintf("Voltage drop across array: %.1f%% of supply. Modeled via Ohm's law across wire parasitics at each row/column intersection.", state.IRDrop*100),
		},
		{
			ID:    "sneak",
			Title: "Sneak Paths",
			Body:  "Sneak path currents modeled via Kirchhoff current law at each column node. Unselected cells contribute parasitic current through their finite conductance.",
		},
		{
			ID:    "mvm",
			Title: "MVM Operation",
			Body:  fmt.Sprintf("%d×%d matrix × %d-element input vector → %d-element output vector. I_out = G × V_in (Ohm's law + Kirchhoff current law).", state.Rows, state.Cols, state.Cols, state.Rows),
		},
	}

	actions := []viewmodel.Action{
		{ID: "resize", Label: "Resize Array", Kind: viewmodel.ActionCommand},
		{ID: "run_mvm", Label: "Run MVM", Kind: viewmodel.ActionCommand},
		{ID: "toggle_ir", Label: "Toggle IR Drop", Kind: viewmodel.ActionToggle},
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID:          viewmodel.ModuleCrossbar,
			Title:       "FeCIM Crossbar Array Visualization",
			Description: "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.",
			Status:      viewmodel.StatusFunctional,
		},
		Metrics:  metrics,
		Sections: sections,
		Actions:  actions,
	}
}
