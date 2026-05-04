package crossbar

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New(8, 8)
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDescriptorHasCorrectID(t *testing.T) {
	m := New(4, 4)
	d := m.Descriptor()
	if d.ID != viewmodel.ModuleCrossbar {
		t.Errorf("Descriptor().ID = %v, want %v", d.ID, viewmodel.ModuleCrossbar)
	}
}

func TestSnapshotContainsArrayDimensions(t *testing.T) {
	m := New(16, 32)
	s := m.Snapshot()
	foundRows, foundCols := false, false
	for _, metric := range s.Metrics {
		if metric.ID == "rows" && metric.Value == "16" {
			foundRows = true
		}
		if metric.ID == "cols" && metric.Value == "32" {
			foundCols = true
		}
	}
	if !foundRows {
		t.Error("rows metric missing or wrong")
	}
	if !foundCols {
		t.Error("cols metric missing or wrong")
	}
}

func TestSnapshotSectionsContainNonIdealities(t *testing.T) {
	m := New(4, 4)
	s := m.Snapshot()
	ids := map[string]bool{}
	for _, section := range s.Sections {
		ids[section.ID] = true
	}
	for _, want := range []string{"ir_drop", "sneak", "mvm"} {
		if !ids[want] {
			t.Errorf("missing section: %s", want)
		}
	}
}

func TestApplyActionUnsupported(t *testing.T) {
	m := New(2, 2)
	err := m.ApplyAction(viewmodel.Action{ID: "test", Kind: viewmodel.ActionCommand})
	if err != viewmodel.ErrUnsupportedAction {
		t.Errorf("ApplyAction error = %v, want %v", err, viewmodel.ErrUnsupportedAction)
	}
}

func TestConductancesInitialized(t *testing.T) {
	m := New(4, 4)
	s := m.Snapshot()
	if s.Descriptor.Status != viewmodel.StatusFunctional {
		t.Errorf("Status = %v, want %v", s.Descriptor.Status, viewmodel.StatusFunctional)
	}
}
