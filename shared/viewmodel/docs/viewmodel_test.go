package docs

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
func TestSnapshotHasSections(t *testing.T) {
	s := New().Snapshot()
	if len(s.Sections) < 2 {
		t.Error("too few sections")
	}
}
