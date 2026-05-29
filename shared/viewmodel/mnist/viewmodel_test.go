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
func TestApplyActionSweepLevelsRejectsInvalidIntegerPayload(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{ID: "sweep_levels", Payload: map[string]string{"levels": "many"}})
	if err == nil {
		t.Fatal("expected invalid levels payload to fail")
	}
	if m.state.NumLevels != 30 {
		t.Fatalf("NumLevels changed after invalid payload: %d", m.state.NumLevels)
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
