package design_test

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"
	"fecim-lattice-tools/shared/viewmodel/design"
	edavm "fecim-lattice-tools/shared/viewmodel/eda"
	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
)

func TestCompositionSnapshot(t *testing.T) {
	c := &design.Composition{}
	ds := c.Snapshot()
	if ds.Summary == "" {
		t.Error("Snapshot().Summary is empty")
	}
}

func TestCompositionExportWithoutEDA(t *testing.T) {
	c := &design.Composition{}
	err := c.ExportDesign()
	if err == nil {
		t.Error("ExportDesign without EDA should return error")
	}
}

func TestDesignSnapshotDefaults(t *testing.T) {
	c := &design.Composition{
		Hysteresis: viewmodel.NewStaticModule(viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleHysteresis,
		}, []viewmodel.Section{}),
	}
	ds := c.Snapshot()
	if ds.Material != "" {
		t.Error("empty hysteresis should produce empty material")
	}
}

func TestCompositionSnapshotUsesRealModuleTypedDesignState(t *testing.T) {
	hysteresis := hysteresisvm.New()
	crossbar := crossbarvm.New(12, 10)
	circuits := circuitsvm.New()
	eda := edavm.New()
	for name, port := range map[string]viewmodel.ModulePort{
		"hysteresis": hysteresis,
		"crossbar":   crossbar,
		"circuits":   circuits,
		"eda":        eda,
	} {
		if _, ok := port.(design.DesignStateProvider); !ok {
			t.Fatalf("%s module does not implement DesignStateProvider", name)
		}
	}

	c := &design.Composition{Hysteresis: hysteresis, Crossbar: crossbar, Circuits: circuits, EDA: eda}
	ds := c.Snapshot()
	if ds.Material == "" || ds.ArrayRows != 12 || ds.ArrayCols != 10 || ds.ADCResolution == 0 || ds.DACResolution == 0 || ds.ProcessNode == "" || ds.DesignName == "" {
		t.Fatalf("DesignSnapshot from real modules = %+v, want typed module values", ds)
	}
}

func TestCompositionSnapshotUsesTypedDesignStateBeforeDisplayMetrics(t *testing.T) {
	c := &design.Composition{
		Hysteresis: typedDesignModule{
			id:      viewmodel.ModuleHysteresis,
			metrics: []viewmodel.Metric{{ID: "material", Value: "display-material"}},
			state:   design.ModuleDesignState{Material: "typed-material"},
		},
		Crossbar: typedDesignModule{
			id:      viewmodel.ModuleCrossbar,
			metrics: []viewmodel.Metric{{ID: "rows", Value: "not-an-int"}, {ID: "cols", Value: "also-bad"}},
			state:   design.ModuleDesignState{ArrayRows: 32, ArrayCols: 16},
		},
		Circuits: typedDesignModule{
			id:      viewmodel.ModuleCircuits,
			metrics: []viewmodel.Metric{{ID: "adc", Value: "visual ADC label"}, {ID: "dac", Value: "visual DAC label"}},
			state:   design.ModuleDesignState{ADCResolution: 7, DACResolution: 8},
		},
		EDA: typedDesignModule{
			id:      viewmodel.ModuleEDA,
			metrics: []viewmodel.Metric{{ID: "process", Value: "display-process"}, {ID: "design", Value: "display-design"}},
			state:   design.ModuleDesignState{ProcessNode: "typed-process", DesignName: "typed-design"},
		},
	}

	ds := c.Snapshot()
	if ds.Material != "typed-material" || ds.ArrayRows != 32 || ds.ArrayCols != 16 || ds.ADCResolution != 7 || ds.DACResolution != 8 || ds.ProcessNode != "typed-process" || ds.DesignName != "typed-design" {
		t.Fatalf("DesignSnapshot = %+v, want typed design state values", ds)
	}
}

type typedDesignModule struct {
	id      viewmodel.ModuleID
	metrics []viewmodel.Metric
	state   design.ModuleDesignState
}

func (m typedDesignModule) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{ID: m.id}
}
func (m typedDesignModule) Snapshot() viewmodel.ModuleSnapshot {
	return viewmodel.ModuleSnapshot{Descriptor: m.Descriptor(), Metrics: m.metrics}
}
func (m typedDesignModule) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m typedDesignModule) Start()                             {}
func (m typedDesignModule) Stop()                              {}
func (m typedDesignModule) DesignState() design.ModuleDesignState {
	return m.state
}
