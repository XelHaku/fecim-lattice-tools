package eda

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	edaconfig "fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct{ state EDAState }

func New() *Module {
	rows, cols := 8, 8
	cfg := compiler.NewComputeConfig(rows, cols)
	design, err := compiler.GenerateDesign(cfg)
	if err != nil {
		design = compiler.GenerateBlank(cfg)
	}

	m := &Module{state: EDAState{
		DesignName:  "fecim_crossbar_8x8",
		ProcessNode: "sky130",
		ArrayRows:   rows, ArrayCols: cols,
		ExportFormats: []string{"spice", "verilog", "liberty", "def", "lef"},
	}}

	if design != nil {
		m.state.TotalCells = design.Stats.TotalCells
		m.state.AreaMM2 = design.Stats.AreaMM2
		m.state.PowerMW = design.Stats.PowerMW
		m.state.SPICESnippet = truncateContent(export.GenerateSPICE(design, 1.8), 15)
		m.state.VerilogSnippet = truncateContent(export.GenerateVerilogWithDefaults(design), 15)
		m.state.DEFSnippet = truncateContent(export.GenerateDEFWithDefaults(design), 15)
		m.state.LEFSnippet = truncateContent(export.GenerateLEF(edaconfig.DefaultCellConfig()), 15)
	}
	return m
}

func truncateContent(s string, maxLines int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n... (%d more lines)", len(lines)-maxLines)
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleEDA,
		Title:       "FeCIM EDA Design Suite",
		Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
		Status:      viewmodel.StatusFunctional,
	}
}
func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case "generate_all":
		return nil
	case "generate_spice":
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}
func (m *Module) Start() {}
func (m *Module) Stop()  {}
