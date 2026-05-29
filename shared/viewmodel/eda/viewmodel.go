package eda

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/shared/viewmodel"
	"fecim-lattice-tools/shared/viewmodel/design"
)

type Module struct{ state EDAState }

func New() *Module {
	return &Module{state: newDesignGenerationWorkflow(8, 8).buildState()}
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

func (m *Module) DesignState() design.ModuleDesignState {
	return design.ModuleDesignState{ProcessNode: m.state.ProcessNode, DesignName: m.state.DesignName}
}
