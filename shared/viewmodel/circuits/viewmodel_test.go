package circuits

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
	if New().Descriptor().ID != viewmodel.ModuleCircuits {
		t.Error("wrong ID")
	}
}

func TestSnapshotContainsCircuitBlocks(t *testing.T) {
	s := New().Snapshot()
	ids := map[string]bool{}
	for _, m := range s.Metrics {
		ids[m.ID] = true
	}
	for _, id := range []string{"adc", "dac", "tia", "charge_pump", "ispp"} {
		if !ids[id] {
			t.Errorf("missing: %s", id)
		}
	}
}

func TestSnapshotContainsUnifiedCircuitControls(t *testing.T) {
	s := New().Snapshot()

	for _, id := range []string{
		"array",
		"mode",
		"architecture",
		"selected_cell",
		"write_target",
		"coupling",
		"ispp_engine",
		"adc",
		"dac",
		"tia",
	} {
		if metricValue(s, id) == "" {
			t.Errorf("missing metric: %s", id)
		}
	}

	if !hasSection(s, "unified_controls") {
		t.Fatal("missing unified controls section")
	}

	for _, id := range []string{
		"resize_array",
		"set_operation_mode",
		"set_architecture",
		"select_cell",
		"set_write_target",
		"set_dac_bits",
		"set_adc_bits",
		"set_tia_gain",
		"set_coupling_tier",
		"set_ispp_engine",
		"run_read",
		"run_write",
		"run_compute",
		"toggle_ispp",
	} {
		if !hasAction(s, id) {
			t.Errorf("missing action: %s", id)
		}
	}
}

func TestApplyActionUpdatesUnifiedCircuitControls(t *testing.T) {
	m := New()

	actions := []viewmodel.Action{
		{ID: "resize_array", Payload: map[string]string{"rows": "32", "cols": "32"}},
		{ID: "set_operation_mode", Payload: map[string]string{"mode": "write"}},
		{ID: "set_architecture", Payload: map[string]string{"architecture": "1T1R (Transistor)"}},
		{ID: "select_cell", Payload: map[string]string{"row": "31", "col": "30"}},
		{ID: "set_write_target", Payload: map[string]string{"level": "17"}},
		{ID: "set_dac_bits", Payload: map[string]string{"bits": "8"}},
		{ID: "set_adc_bits", Payload: map[string]string{"bits": "7"}},
		{ID: "set_tia_gain", Payload: map[string]string{"gain_ohm": "50000"}},
		{ID: "set_coupling_tier", Payload: map[string]string{"tier": "Tier-B"}},
		{ID: "set_ispp_engine", Payload: map[string]string{"engine": "Landau-Khalatnikov (Physics ODE)"}},
	}
	for _, action := range actions {
		if err := m.ApplyAction(action); err != nil {
			t.Fatalf("ApplyAction(%s): %v", action.ID, err)
		}
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"array":         "32x32",
		"mode":          "WRITE",
		"architecture":  "1T1R (Transistor)",
		"selected_cell": "[31,30]",
		"write_target":  "17/29",
		"adc":           "7-bit SAR",
		"dac":           "8-bit R-2R",
		"tia":           "50 kΩ",
		"coupling":      "Tier-B",
		"ispp_engine":   "Landau-Khalatnikov (Physics ODE)",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
}

func metricValue(s viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range s.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}

func hasSection(s viewmodel.ModuleSnapshot, id string) bool {
	for _, section := range s.Sections {
		if section.ID == id {
			return true
		}
	}
	return false
}

func hasAction(s viewmodel.ModuleSnapshot, id string) bool {
	for _, action := range s.Actions {
		if action.ID == id {
			return true
		}
	}
	return false
}
