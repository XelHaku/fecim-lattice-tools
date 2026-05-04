package hysteresis

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDescriptorHasCorrectID(t *testing.T) {
	m := New()
	d := m.Descriptor()
	if d.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Descriptor().ID = %v, want %v", d.ID, viewmodel.ModuleHysteresis)
	}
	if d.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor().Status = %v, want %v", d.Status, viewmodel.StatusFunctional)
	}
}

func TestSnapshotContainsMetrics(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Metrics) == 0 {
		t.Error("Snapshot().Metrics is empty, expected material metrics")
	}
	if s.Descriptor.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Snapshot().Descriptor.ID = %v", s.Descriptor.ID)
	}
}

func TestSnapshotHasSections(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Sections) == 0 {
		t.Error("Snapshot().Sections is empty")
	}
}

func TestApplyActionUnsupported(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{ID: "run_simulation", Kind: viewmodel.ActionCommand})
	if err != viewmodel.ErrUnsupportedAction {
		t.Errorf("ApplyAction error = %v, want %v", err, viewmodel.ErrUnsupportedAction)
	}
}

func TestMaterialSummary(t *testing.T) {
	summary := materialSummary(nil)
	if summary != "N/A" {
		t.Errorf("materialSummary(nil) = %q, want N/A", summary)
	}
}

func TestStateDefaults(t *testing.T) {
	m := New()
	s := m.Snapshot()
	var matMetric *viewmodel.Metric
	for i := range s.Metrics {
		if s.Metrics[i].ID == "material" {
			matMetric = &s.Metrics[i]
			break
		}
	}
	if matMetric == nil {
		t.Fatal("no material metric found")
	}
	if matMetric.Value == "" {
		t.Error("material metric value is empty")
	}
}
