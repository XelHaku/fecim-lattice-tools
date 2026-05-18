package mnist

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
func TestSnapshotContainsAccuracy(t *testing.T) {
	s := New().Snapshot()
	found := false
	for _, m := range s.Metrics {
		if m.ID == "accuracy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("no accuracy metric")
	}
}
