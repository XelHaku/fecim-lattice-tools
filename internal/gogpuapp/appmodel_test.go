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

func TestBuildAppPortsCoversAllKnownDescriptors(t *testing.T) {
	ports := BuildAppPorts()
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

func TestBuildAppPortsAllReleasedModulesAreFunctional(t *testing.T) {
	ports := BuildAppPorts()
	if len(ports) == 0 {
		t.Fatal("BuildAppPorts returned no ports")
	}
	for _, p := range ports {
		desc := p.Descriptor()
		if desc.Status != viewmodel.StatusFunctional {
			t.Errorf("port %s Status = %q, want %q", desc.ID, desc.Status, viewmodel.StatusFunctional)
		}
		snap := p.Snapshot()
		if snap.Descriptor.Status != viewmodel.StatusFunctional {
			t.Errorf("port %s snapshot Status = %q, want %q", desc.ID, snap.Descriptor.Status, viewmodel.StatusFunctional)
		}
	}
}
