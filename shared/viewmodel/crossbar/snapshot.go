package crossbar

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state CrossbarState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "rows", Label: "Rows", Value: fmt.Sprintf("%d", state.Rows), Unit: "cells"},
		{ID: "cols", Label: "Columns", Value: fmt.Sprintf("%d", state.Cols), Unit: "cells"},
		{ID: "ir_drop", Label: "IR Drop", Value: fmt.Sprintf("%.1f%%", state.IRDrop*100)},
		{ID: "drift", Label: "Drift Factor", Value: fmt.Sprintf("%.2f", state.DriftFactor)},
	}

	sections := []viewmodel.Section{
		{
			ID:       "ir_drop",
			Title:    "IR Drop",
			Body:     fmt.Sprintf("Voltage drop across array: %.1f%% of supply. Modeled via Ohm's law across wire parasitics at each row/column intersection.", state.IRDrop*100),
			Category: "research",
		},
		{
			ID:       "sneak",
			Title:    "Sneak Paths",
			Body:     "Sneak path currents modeled via Kirchhoff current law at each column node. Unselected cells contribute parasitic current through their finite conductance.",
			Category: "research",
		},
		{
			ID:       "mvm",
			Title:    "MVM Operation",
			Body:     fmt.Sprintf("%d×%d matrix × %d-element input vector → %d-element output vector. I_out = G × V_in (Ohm's law + Kirchhoff current law).", state.Rows, state.Cols, state.Cols, state.Rows),
			Category: "design",
		},
	}
	sections = append(sections, viewmodel.Section{
		ID:       "edu_ohm",
		Title:    "How Crossbars Compute",
		Body:     "Each cell stores conductance G. When voltage V is applied to a row, current I = G × V flows (Ohm's law). All column currents sum at the output (Kirchhoff's current law), computing I = G × V in a single analog step — no digital multiply-accumulate needed.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID:       "edu_irdrop",
		Title:    "IR Drop Explained",
		Body:     "The metal wires connecting cells have resistance. As current flows through them, voltage drops along the wire path, reducing the effective voltage at distant cells. This is IR drop — a key non-ideality that limits array size and accuracy.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID:       "research_drift",
		Title:    "Conductance Drift",
		Body:     "Conductance values drift over time due to relaxation effects in ferroelectric domains. Drift is temperature- and time-dependent. Literature models: power-law drift G(t) = G₀·(t/t₀)^(−ν). Typical ν ≈ 0.01–0.05 for HZO.",
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID:       "design_array",
		Title:    "Array Design Tradeoffs",
		Body:     fmt.Sprintf("Larger arrays increase parallelism but worsen IR drop. Recommended: start at 8×8, validate non-idealities, then scale. Quantization: %d discrete levels of %d configurable. Cross-reference: Module 1 for material → conductance mapping; Module 6 for array → layout export.", state.Cols*state.Rows, 30),
		Category: "design",
	})

	actions := []viewmodel.Action{
		{ID: "resize", Label: "Resize Array", Kind: viewmodel.ActionCommand},
		{ID: "run_mvm", Label: "Run MVM", Kind: viewmodel.ActionCommand},
		{ID: "toggle_ir", Label: "Toggle IR Drop", Kind: viewmodel.ActionToggle},
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID:             viewmodel.ModuleCrossbar,
			Title:          "FeCIM Crossbar Array Visualization",
			Description:    "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.",
			Status:         viewmodel.StatusFunctional,
			BoundaryNotice: "SIMULATION OUTPUT — Not measured device data. Conductance quantization (30 levels) is an educational baseline, not a validated device parameter. IR drop and sneak paths use idealized parasitic models.",
		},
		Metrics:  metrics,
		Sections: sections,
		Actions:  actions,
		Plots:    buildHeatmapPlots(state),
	}
}

func buildHeatmapPlots(state CrossbarState) []viewmodel.PlotData {
	if len(state.Conductances) == 0 || len(state.Conductances[0]) == 0 {
		return nil
	}
	return []viewmodel.PlotData{{
		ID:     "conductance_matrix",
		Title:  fmt.Sprintf("%d×%d Conductance Matrix", state.Rows, state.Cols),
		XLabel: "Column",
		YLabel: "Row",
		Series: []viewmodel.PlotSeries{{
			Name:   "G (µS)",
			Points: flattenConductances(state.Conductances),
		}},
	}, {
		ID:     "mvm_result",
		Title:  "MVM Output Vector",
		XLabel: "Row Index",
		YLabel: "I_out (µA)",
		Series: []viewmodel.PlotSeries{{
			Name:   "output",
			Points: vectorToPoints(state.OutputVector),
		}},
	}}
}

func flattenConductances(g [][]float64) []viewmodel.PlotPoint {
	pts := make([]viewmodel.PlotPoint, 0, len(g)*len(g[0]))
	for i, row := range g {
		for j, val := range row {
			pts = append(pts, viewmodel.PlotPoint{X: float64(j), Y: float64(i), V: val})
		}
	}
	return pts
}

func vectorToPoints(v []float64) []viewmodel.PlotPoint {
	pts := make([]viewmodel.PlotPoint, len(v))
	for i, val := range v {
		pts[i] = viewmodel.PlotPoint{X: float64(i), Y: val}
	}
	return pts
}
