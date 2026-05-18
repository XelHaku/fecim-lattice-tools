package eda

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state EDAState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "design", Label: "Design", Value: state.DesignName},
		{ID: "process", Label: "Process Node", Value: state.ProcessNode},
		{ID: "array", Label: "Array Size", Value: fmt.Sprintf("%d×%d", state.ArrayRows, state.ArrayCols)},
		{ID: "cells", Label: "Total Cells", Value: fmt.Sprintf("%d", state.TotalCells)},
		{ID: "area", Label: "Area", Value: fmt.Sprintf("%.3f mm²", state.AreaMM2)},
		{ID: "power", Label: "Power", Value: fmt.Sprintf("%.1f mW", state.PowerMW)},
		{ID: "spice", Label: "SPICE", Value: "ready"},
		{ID: "verilog", Label: "Verilog", Value: "ready"},
		{ID: "liberty", Label: "Liberty", Value: "ready"},
		{ID: "def", Label: "DEF", Value: "ready"},
		{ID: "lef", Label: "LEF", Value: "ready"},
	}
	sections := []viewmodel.Section{
		{ID: "design_stats", Title: "Design Statistics", Body: fmt.Sprintf("Generated %d×%d compute-mode design on %s.\nArea: %.3f mm²  |  Power: %.1f mW  |  Cells: %d", state.ArrayRows, state.ArrayCols, state.ProcessNode, state.AreaMM2, state.PowerMW, state.TotalCells), Category: "design"},
	}
	if state.SPICESnippet != "" {
		sections = append(sections, viewmodel.Section{
			ID:       "spice_content",
			Title:    "SPICE Netlist (first 15 lines)",
			Body:     state.SPICESnippet,
			Category: "research",
		})
	}
	if state.VerilogSnippet != "" {
		sections = append(sections, viewmodel.Section{
			ID:       "verilog_content",
			Title:    "Verilog Module (first 15 lines)",
			Body:     state.VerilogSnippet,
			Category: "research",
		})
	}
	if state.DEFSnippet != "" {
		sections = append(sections, viewmodel.Section{
			ID:       "def_content",
			Title:    "DEF Layout (first 15 lines)",
			Body:     state.DEFSnippet,
			Category: "design",
		})
	}
	if state.LEFSnippet != "" {
		sections = append(sections, viewmodel.Section{
			ID:       "lef_content",
			Title:    "LEF Macro (first 15 lines)",
			Body:     state.LEFSnippet,
			Category: "design",
		})
	}
	sections = append(sections, viewmodel.Section{
		ID: "edu_spice", Title: "What is SPICE?",
		Body:     "SPICE (Simulation Program with Integrated Circuit Emphasis) describes circuits as netlists — text files listing components and how they connect. SPICE simulators solve Kirchhoff's laws numerically. Our export generates a netlist with FeFET compact models.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "edu_flow", Title: "Design Flow",
		Body:     fmt.Sprintf("RTL → Synthesis → Floorplan → Place → Route → DRC/LVS → GDSII. Our %s flow targets OpenLane for open-source tapeout. SPICE for circuit simulation, Verilog for digital control, Liberty for timing closure.", state.ProcessNode),
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "research_validation", Title: "Validation Status",
		Body:     "Generated SPICE netlists validated with ngspice roundtrip tests. Netlist syntax checked against Verilog-A compact model. Liberty corners verified (TT/FF/SS). All exports are educational baselines — not validated against silicon measurements.",
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID: "design_workflow", Title: "Design Composition",
		Body:     "Design wizard: Start with material selection (Module 1) → Array configuration (Module 2) → Peripheral specs (Module 4) → Export (here). Use parameter sweep to generate netlist families across voltage corners. All formats exportable in one step.",
		Category: "design",
	})
	actions := []viewmodel.Action{
		{ID: "generate_all", Label: "Export All Formats", Kind: viewmodel.ActionCommand},
		{ID: "generate_spice", Label: "Generate SPICE", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleEDA, Title: "FeCIM EDA Design Suite",
			Description:    "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
			Status:         viewmodel.StatusFunctional,
			BoundaryNotice: "EDUCATIONAL EDA — Not a production chip design tool. Generated netlists and layouts are educational examples. Does not use proprietary foundry PDKs.",
		},
		Metrics: metrics, Sections: sections, Actions: actions,
	}
}
