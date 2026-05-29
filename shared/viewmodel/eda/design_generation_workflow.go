package eda

import (
	"fecim-lattice-tools/module6-eda/pkg/compiler"
	edaconfig "fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

type designGenerationWorkflow struct {
	rows int
	cols int
}

func newDesignGenerationWorkflow(rows, cols int) designGenerationWorkflow {
	return designGenerationWorkflow{rows: rows, cols: cols}
}

func (w designGenerationWorkflow) buildState() EDAState {
	cfg := compiler.NewComputeConfig(w.rows, w.cols)
	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		design = compiler.GenerateBlank(cfg)
	}

	state := EDAState{
		DesignName:    "fecim_crossbar_8x8",
		ProcessNode:   "sky130",
		ArrayRows:     w.rows,
		ArrayCols:     w.cols,
		ExportFormats: []string{"spice", "verilog", "liberty", "def", "lef"},
	}

	if design != nil {
		state.TotalCells = design.Stats.TotalCells
		state.AreaMM2 = design.Stats.AreaMM2
		state.PowerMW = design.Stats.PowerMW
		state.SPICESnippet = truncateContent(export.GenerateSPICE(design, 1.8), 15)
		state.VerilogSnippet = truncateContent(export.GenerateVerilogWithDefaults(design), 15)
		state.DEFSnippet = truncateContent(export.GenerateDEFWithDefaults(design), 15)
		state.LEFSnippet = truncateContent(export.GenerateLEF(edaconfig.DefaultCellConfig()), 15)
	}
	return state
}
