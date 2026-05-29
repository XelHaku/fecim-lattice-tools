package gogpuapp

import (
	"errors"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestFullAppE2ENavigateAndDispatchRepresentativeActions(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)

	cases := []struct {
		module   viewmodel.ModuleID
		action   viewmodel.Action
		readOnly bool
	}{
		{module: viewmodel.ModuleHysteresis, action: viewmodel.Action{ID: "toggle_simulation", Kind: viewmodel.ActionToggle}},
		{module: viewmodel.ModuleCrossbar, action: viewmodel.Action{ID: "run_mvm", Kind: viewmodel.ActionCommand}},
		{module: viewmodel.ModuleMNIST, action: viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "16"}}},
		{module: viewmodel.ModuleCircuits, action: viewmodel.Action{ID: "run_read", Kind: viewmodel.ActionCommand}},
		{module: viewmodel.ModuleComparison, action: viewmodel.Action{ID: "unsupported_e2e_probe", Kind: viewmodel.ActionCommand}, readOnly: true},
		{module: viewmodel.ModuleEDA, action: viewmodel.Action{ID: "generate_spice", Kind: viewmodel.ActionCommand}},
		{module: viewmodel.ModuleDocs, action: viewmodel.Action{ID: "start_curriculum", Kind: viewmodel.ActionCommand}},
	}

	for _, tc := range cases {
		t.Run(string(tc.module), func(t *testing.T) {
			if !model.SelectModule(tc.module) {
				t.Fatalf("SelectModule(%s) returned false", tc.module)
			}
			before := model.ActivePort().Snapshot()
			if before.Descriptor.ID != tc.module {
				t.Fatalf("active snapshot descriptor = %s, want %s", before.Descriptor.ID, tc.module)
			}
			if !tc.readOnly && !snapshotHasAction(before, tc.action.ID) {
				t.Fatalf("module %s snapshot does not expose representative action %q", tc.module, tc.action.ID)
			}

			err := model.DispatchAction(tc.action)
			if tc.readOnly {
				if !errors.Is(err, viewmodel.ErrUnsupportedAction) {
					t.Fatalf("read-only module dispatch error = %v, want ErrUnsupportedAction", err)
				}
			} else if err != nil {
				t.Fatalf("DispatchAction(%s/%s) returned error: %v", tc.module, tc.action.ID, err)
			}

			after := model.ActivePort().Snapshot()
			if after.Descriptor.ID != tc.module {
				t.Fatalf("dispatch changed active module to %s, want %s", after.Descriptor.ID, tc.module)
			}
			if len(after.Sections) == 0 {
				t.Fatalf("module %s has no sections after dispatch", tc.module)
			}
		})
	}
}

func TestFullAppE2ENavigationPreservesPerModuleState(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)

	if !model.SelectModule(viewmodel.ModuleMNIST) {
		t.Fatal("SelectModule(mnist) returned false")
	}
	if err := model.DispatchAction(viewmodel.Action{ID: "sweep_levels", Kind: viewmodel.ActionCommand, Payload: map[string]string{"levels": "16"}}); err != nil {
		t.Fatalf("DispatchAction(mnist/sweep_levels): %v", err)
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "levels"); got != "16 levels" {
		t.Fatalf("mnist levels metric after dispatch = %q, want 16 levels", got)
	}

	for _, id := range []viewmodel.ModuleID{
		viewmodel.ModuleHysteresis,
		viewmodel.ModuleCrossbar,
		viewmodel.ModuleCircuits,
		viewmodel.ModuleComparison,
		viewmodel.ModuleEDA,
		viewmodel.ModuleDocs,
	} {
		if !model.SelectModule(id) {
			t.Fatalf("SelectModule(%s) returned false", id)
		}
		if got := model.ActivePort().Snapshot().Descriptor.ID; got != id {
			t.Fatalf("active descriptor = %s, want %s", got, id)
		}
	}

	if !model.SelectModule(viewmodel.ModuleMNIST) {
		t.Fatal("SelectModule(mnist) returned false on return")
	}
	if got := appModelE2EMetricValue(model.ActivePort().Snapshot(), "levels"); got != "16 levels" {
		t.Fatalf("mnist levels metric after round-trip navigation = %q, want preserved 16 levels", got)
	}
}

func TestFullAppE2EAllModulesPublishTrustBoundariesAndContent(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)

	for _, descriptor := range viewmodel.KnownDescriptors() {
		t.Run(string(descriptor.ID), func(t *testing.T) {
			if !model.SelectModule(descriptor.ID) {
				t.Fatalf("SelectModule(%s) returned false", descriptor.ID)
			}
			snapshot := model.ActivePort().Snapshot()
			if snapshot.Descriptor.ID != descriptor.ID {
				t.Fatalf("snapshot descriptor = %s, want %s", snapshot.Descriptor.ID, descriptor.ID)
			}
			if snapshot.Descriptor.BoundaryNotice == "" {
				t.Fatalf("module %s omitted boundary notice", descriptor.ID)
			}
			if len(snapshot.Metrics) == 0 {
				t.Fatalf("module %s omitted user-visible metrics", descriptor.ID)
			}
			if len(snapshot.Sections) == 0 {
				t.Fatalf("module %s omitted user-visible sections", descriptor.ID)
			}
		})
	}
}

func TestFullAppE2EStartsAndStopsAllRegisteredModules(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)
	wantDescriptors := viewmodel.KnownDescriptors()
	if len(model.Ports) != len(wantDescriptors) {
		t.Fatalf("ports = %d, want %d known descriptors", len(model.Ports), len(wantDescriptors))
	}

	model.StartAllModules()
	for i, port := range model.Ports {
		if got, want := port.Snapshot().Descriptor.ID, wantDescriptors[i].ID; got != want {
			t.Fatalf("after StartAllModules port[%d] descriptor = %s, want %s", i, got, want)
		}
	}

	model.StopAllModules()
	for i, port := range model.Ports {
		if got, want := port.Snapshot().Descriptor.ID, wantDescriptors[i].ID; got != want {
			t.Fatalf("after StopAllModules port[%d] descriptor = %s, want %s", i, got, want)
		}
	}
}

func snapshotHasAction(snapshot viewmodel.ModuleSnapshot, id string) bool {
	for _, action := range snapshot.Actions {
		if action.ID == id {
			return true
		}
	}
	return false
}

func appModelE2EMetricValue(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range snapshot.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}
