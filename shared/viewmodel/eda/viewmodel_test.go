package eda

import (
	"fecim-lattice-tools/shared/viewmodel"
	"testing"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}
func TestDescriptorHasCorrectID(t *testing.T) {
	m := New()
	if m.Descriptor().ID != viewmodel.ModuleEDA {
		t.Errorf("ID = %v", m.Descriptor().ID)
	}
}
func TestSnapshotContainsExportFormats(t *testing.T) {
	m := New()
	s := m.Snapshot()
	ids := map[string]bool{}
	for _, m := range s.Metrics {
		ids[m.ID] = true
	}
	for _, id := range []string{"spice", "verilog", "liberty", "def", "lef"} {
		if !ids[id] {
			t.Errorf("missing: %s", id)
		}
	}
}
