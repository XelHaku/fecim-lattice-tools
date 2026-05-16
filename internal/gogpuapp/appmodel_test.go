package gogpuapp

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestDefaultAppSpecNamesReleasedShell(t *testing.T) {
	spec := DefaultAppSpec()

	if spec.Title != "FeCIM Lattice Tools" {
		t.Fatalf("Title = %q, want FeCIM Lattice Tools", spec.Title)
	}
	if spec.Command != "fecim-lattice-tools" {
		t.Fatalf("Command = %q, want fecim-lattice-tools", spec.Command)
	}
	if spec.Width != 1400 {
		t.Fatalf("Width = %d, want 1400", spec.Width)
	}
	if spec.Height != 900 {
		t.Fatalf("Height = %d, want 900", spec.Height)
	}
}

func TestAppModelSelectsRequestedModule(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleCrossbar)

	if model.ActiveModuleID != viewmodel.ModuleCrossbar {
		t.Fatalf("ActiveModuleID = %q, want %q", model.ActiveModuleID, viewmodel.ModuleCrossbar)
	}
	if got := model.ActivePort().Descriptor().ID; got != viewmodel.ModuleCrossbar {
		t.Fatalf("ActivePort descriptor ID = %q, want %q", got, viewmodel.ModuleCrossbar)
	}
}

func TestBuildPlaceholderPortsCoversAllKnownDescriptors(t *testing.T) {
	ports := BuildPlaceholderPorts()
	descriptors := viewmodel.KnownDescriptors()

	if len(ports) != len(descriptors) {
		t.Fatalf("port count = %d, want %d", len(ports), len(descriptors))
	}

	for i, port := range ports {
		got := port.Descriptor()
		want := descriptors[i]
		if got.ID != want.ID {
			t.Fatalf("port[%d] descriptor.ID = %#v, want %#v", i, got.ID, want.ID)
		}
		if got.Title != want.Title {
			t.Fatalf("port[%d] descriptor.Title = %#v, want %#v", i, got.Title, want.Title)
		}
		if got.Status == "" {
			t.Fatalf("port[%d] descriptor.Status is empty", i)
		}
		snapshot := port.Snapshot()
		if snapshot.Descriptor.ID != want.ID {
			t.Fatalf("port[%d] snapshot descriptor.ID = %#v, want %#v", i, snapshot.Descriptor.ID, want.ID)
		}
		if len(snapshot.Sections) == 0 {
			t.Fatalf("port[%d] snapshot has no sections", i)
		}
		if err := port.ApplyAction(viewmodel.Action{ID: "unknown"}); err == nil {
			t.Fatalf("port[%d] ApplyAction for unknown action returned nil error", i)
		}
	}
}

func TestBuildPlaceholderPorts_ComparisonIsFunctional(t *testing.T) {
	ports := BuildPlaceholderPorts()
	var got viewmodel.ModulePort
	for _, p := range ports {
		if p.Descriptor().ID == viewmodel.ModuleComparison {
			got = p
			break
		}
	}
	if got == nil {
		t.Fatal("no port found for ModuleComparison")
	}
	if got.Descriptor().Status != viewmodel.StatusFunctional {
		t.Errorf("comparison port Status = %q, want %q (no longer placeholder)",
			got.Descriptor().Status, viewmodel.StatusFunctional)
	}
	snap := got.Snapshot()
	if len(snap.Sections) < 3 {
		t.Errorf("comparison snapshot has %d sections, want >= 3 (one per canonical architecture)", len(snap.Sections))
	}
	if len(snap.Metrics) == 0 {
		t.Error("comparison snapshot has no metrics")
	}
}
