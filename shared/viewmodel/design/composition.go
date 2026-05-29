package design

import (
	"fmt"

	"fecim-lattice-tools/shared/viewmodel"
)

// ModuleDesignState contains typed cross-module design values. Fields are
// optional; zero values mean the contributing module does not provide that part.
type ModuleDesignState struct {
	Material      string
	ArrayRows     int
	ArrayCols     int
	ADCResolution int
	DACResolution int
	ProcessNode   string
	DesignName    string
}

// DesignStateProvider is implemented by ModulePort adapters that can contribute
// typed design state without requiring Composition to parse display metrics.
type DesignStateProvider interface {
	DesignState() ModuleDesignState
}

// Composition holds references to all active module ports and provides
// cross-module design aggregation.
type Composition struct {
	Hysteresis viewmodel.ModulePort
	Crossbar   viewmodel.ModulePort
	Circuits   viewmodel.ModulePort
	EDA        viewmodel.ModulePort
}

// DesignSnapshot aggregates state across the design pipeline:
// Material → Array → Circuits → Export.
type DesignSnapshot struct {
	Material      string
	ArrayRows     int
	ArrayCols     int
	ADCResolution int
	DACResolution int
	ProcessNode   string
	DesignName    string
	Summary       string
}

// Snapshot computes a unified design state from all modules.
func (c *Composition) Snapshot() DesignSnapshot {
	ds := DesignSnapshot{}

	if c.Hysteresis != nil {
		state := designStateFrom(c.Hysteresis)
		ds.Material = state.Material
	}
	if c.Crossbar != nil {
		state := designStateFrom(c.Crossbar)
		ds.ArrayRows = state.ArrayRows
		ds.ArrayCols = state.ArrayCols
	}
	if c.Circuits != nil {
		state := designStateFrom(c.Circuits)
		ds.ADCResolution = state.ADCResolution
		ds.DACResolution = state.DACResolution
	}
	if c.EDA != nil {
		state := designStateFrom(c.EDA)
		ds.ProcessNode = state.ProcessNode
		ds.DesignName = state.DesignName
	}

	ds.Summary = fmt.Sprintf("Design: %s | %s × %d×%d (%d-bit ADC/%d-bit DAC) @ %s",
		ds.DesignName, ds.Material, ds.ArrayRows, ds.ArrayCols,
		ds.ADCResolution, ds.DACResolution, ds.ProcessNode)
	return ds
}

func designStateFrom(port viewmodel.ModulePort) ModuleDesignState {
	if provider, ok := port.(DesignStateProvider); ok {
		return provider.DesignState()
	}
	state := ModuleDesignState{}
	for _, m := range port.Snapshot().Metrics {
		switch m.ID {
		case "material":
			state.Material = m.Value
		case "rows":
			fmt.Sscanf(m.Value, "%d", &state.ArrayRows)
		case "cols":
			fmt.Sscanf(m.Value, "%d", &state.ArrayCols)
		case "adc":
			fmt.Sscanf(m.Value, "%d-bit", &state.ADCResolution)
		case "dac":
			fmt.Sscanf(m.Value, "%d-bit", &state.DACResolution)
		case "process":
			state.ProcessNode = m.Value
		case "design":
			state.DesignName = m.Value
		}
	}
	return state
}

// ExportDesign triggers export across the EDA module.
func (c *Composition) ExportDesign() error {
	if c.EDA == nil {
		return fmt.Errorf("design: EDA module not connected")
	}
	return c.EDA.ApplyAction(viewmodel.Action{
		ID:   "generate_all",
		Kind: viewmodel.ActionCommand,
	})
}
