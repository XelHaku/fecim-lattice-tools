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

func TestSnapshotContainsHalfSelectStressState(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetOperationMode,
		Payload: map[string]string{"mode": OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSelectCell,
		Payload: map[string]string{"row": "3", "col": "4"},
	}); err != nil {
		t.Fatalf("select cell: %v", err)
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"half_select_state":    "column-write active",
		"half_select_cells":    "7 same-column cells",
		"disturb_voltage":      "1.40 V",
		"stress_budget":        "400 pulses/level",
		"stress_per_pulse":     "0.002500 level/pulse",
		"stress_selected_cell": "[3,4]",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
	if !hasSection(s, "half_select_stress") {
		t.Fatal("missing half-select stress section")
	}
}

func TestHalfSelectStressBudgetFollowsArchitecture(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetOperationMode,
		Payload: map[string]string{"mode": OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}

	passive := m.Snapshot()
	if got := metricValue(passive, "stress_budget"); got != "400 pulses/level" {
		t.Fatalf("passive stress budget = %q, want 400 pulses/level", got)
	}

	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetArchitecture,
		Payload: map[string]string{"architecture": Architecture1T1R},
	}); err != nil {
		t.Fatalf("set 1T1R: %v", err)
	}
	oneT := m.Snapshot()
	if got := metricValue(oneT, "half_select_state"); got != "attenuated residual" {
		t.Fatalf("1T1R state = %q, want attenuated residual", got)
	}
	if got := metricValue(oneT, "stress_budget"); got != "8000 pulses/level" {
		t.Fatalf("1T1R stress budget = %q, want 8000 pulses/level", got)
	}

	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetArchitecture,
		Payload: map[string]string{"architecture": Architecture2T1R},
	}); err != nil {
		t.Fatalf("set 2T1R: %v", err)
	}
	twoT := m.Snapshot()
	if got := metricValue(twoT, "half_select_state"); got != "isolated" {
		t.Fatalf("2T1R state = %q, want isolated", got)
	}
	if got := metricValue(twoT, "half_select_cells"); got != "0 cells" {
		t.Fatalf("2T1R cells = %q, want 0 cells", got)
	}
}

func TestSnapshotContainsPVTInvestigationSummaries(t *testing.T) {
	s := New().Snapshot()

	wantMetrics := map[string]string{
		"pvt_temperature_sweep": "pass -40/25/85/125 C",
		"pvt_process_yield":     "100.0% (20/20)",
		"pvt_corner_enob":       "FF 4.51 / TT 4.42 / SS 4.30 bits",
		"pvt_noise_ceiling":     "13.61 bits at 16-bit ADC",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
	if !hasSection(s, "pvt_investigations") {
		t.Fatal("missing PVT investigations section")
	}
}

func TestPVTInvestigationSummaryFollowsADCBits(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetADCBits,
		Payload: map[string]string{"bits": "7"},
	}); err != nil {
		t.Fatalf("set ADC bits: %v", err)
	}

	if got := metricValue(m.Snapshot(), "pvt_corner_enob"); got != "FF 6.51 / TT 6.42 / SS 6.30 bits" {
		t.Fatalf("pvt corner ENOB = %q, want 7-bit summary", got)
	}
}

func TestSnapshotContainsReferenceSpecSummaries(t *testing.T) {
	s := New().Snapshot()

	wantMetrics := map[string]string{
		"spec_storage":       "64 cells / 4.91 bits/cell",
		"spec_components":    "DAC 8 / TIA 8 / ADC 8",
		"spec_power_latency": "5.8 mW / 76 ns",
		"spec_throughput":    "0.84 GOPS / 145 GOPS/W",
		"spec_resolution":    "DAC 32 / ADC 32 / 30 levels",
		"spec_compliance":    "OK: DAC/ADC cover 30 levels",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
	if !hasSection(s, "reference_specs") {
		t.Fatal("missing reference specs section")
	}
}

func TestReferenceSpecSummaryFollowsArrayAndDACBits(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionResizeArray,
		Payload: map[string]string{"rows": "32", "cols": "32"},
	}); err != nil {
		t.Fatalf("resize array: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetDACBits,
		Payload: map[string]string{"bits": "4"},
	}); err != nil {
		t.Fatalf("set DAC bits: %v", err)
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"spec_storage":       "1024 cells / 4.91 bits/cell",
		"spec_components":    "DAC 32 / TIA 32 / ADC 32",
		"spec_power_latency": "21.4 mW / 76 ns",
		"spec_throughput":    "13.47 GOPS / 630 GOPS/W",
		"spec_resolution":    "DAC 16 / ADC 32 / 30 levels",
		"spec_compliance":    "CHECK: DAC 16 codes < 30 levels",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
}

func TestSnapshotContainsReferenceTimingSummaries(t *testing.T) {
	s := New().Snapshot()

	wantMetrics := map[string]string{
		"timing_write":         "203 ns total",
		"timing_read":          "76 ns total",
		"timing_compute":       "76 ns total",
		"timing_active":        "READ 76 ns total",
		"timing_active_phases": "DAC 10 / Array 5 / TIA 11 / ADC 50 ns",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
	if !hasSection(s, "reference_timing") {
		t.Fatal("missing reference timing section")
	}
}

func TestReferenceTimingSummaryFollowsOperationMode(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetOperationMode,
		Payload: map[string]string{"mode": OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}
	write := m.Snapshot()
	if got := metricValue(write, "timing_active"); got != "WRITE 203 ns total" {
		t.Fatalf("write timing active = %q, want WRITE total", got)
	}
	if got := metricValue(write, "timing_active_phases"); got != "DAC 10 / Pump 88 / Pulse 100 / Array 5 ns" {
		t.Fatalf("write timing phases = %q, want write phase summary", got)
	}

	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetOperationMode,
		Payload: map[string]string{"mode": OperationCompute},
	}); err != nil {
		t.Fatalf("set compute mode: %v", err)
	}
	compute := m.Snapshot()
	if got := metricValue(compute, "timing_active"); got != "COMPUTE 76 ns total" {
		t.Fatalf("compute timing active = %q, want COMPUTE total", got)
	}
	if got := metricValue(compute, "timing_active_phases"); got != "DAC 10 / Array 5 / TIA+ADC 61 ns" {
		t.Fatalf("compute timing phases = %q, want compute phase summary", got)
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
