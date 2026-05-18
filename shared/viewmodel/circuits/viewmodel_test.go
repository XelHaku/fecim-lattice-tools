package circuits

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		"export_operation_log",
		"export_reference_specs",
		"export_reference_timing",
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

func TestReferenceSpecExportBuffersJSONArtifact(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: ActionExportReferenceSpecs}); err != nil {
		t.Fatalf("export reference specs: %v", err)
	}

	if m.state.ReferenceSpecExportJSON == "" {
		t.Fatal("expected buffered reference spec export JSON")
	}
	var payload ReferenceSpecExport
	if err := json.Unmarshal([]byte(m.state.ReferenceSpecExportJSON), &payload); err != nil {
		t.Fatalf("reference spec export JSON did not unmarshal: %v\n%s", err, m.state.ReferenceSpecExportJSON)
	}
	if payload.Schema != "fecim.circuits.reference_specs.v1" {
		t.Fatalf("schema = %q, want reference-spec schema", payload.Schema)
	}
	if payload.Module != "circuits" || payload.Rows != 8 || payload.Cols != 8 || payload.QuantLevels != DefaultQuantLevels {
		t.Fatalf("unexpected reference spec context: %+v", payload)
	}
	if payload.Cells != 64 || payload.DACCount != 8 || payload.TIACount != 8 || payload.ADCCount != 8 {
		t.Fatalf("unexpected component counts: %+v", payload)
	}
	if payload.DACCodes != 32 || payload.ADCCodes != 32 {
		t.Fatalf("unexpected converter codes: DAC %d ADC %d", payload.DACCodes, payload.ADCCodes)
	}
	if !near(payload.BitsPerCell, 4.9068905956, 1e-9) {
		t.Fatalf("bits/cell = %.10f, want log2(30)", payload.BitsPerCell)
	}
	if !near(payload.TotalPowerMW, 5.8, 1e-9) || !near(payload.LatencyNS, 76, 1e-9) {
		t.Fatalf("unexpected power/latency: %+v", payload)
	}
	if !near(payload.ThroughputGOPS, 0.8421052632, 1e-9) || !near(payload.EfficiencyGOPSW, 145.1905626134, 1e-9) {
		t.Fatalf("unexpected performance terms: %+v", payload)
	}
	if payload.Compliance != "OK: DAC/ADC cover 30 levels" {
		t.Fatalf("compliance = %q, want OK summary", payload.Compliance)
	}
	if !contains(payload.BoundaryNotice, "educational") {
		t.Fatalf("boundary notice = %q, want educational boundary", payload.BoundaryNotice)
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"reference_spec_export":       "buffered 64 cells",
		"reference_spec_export_path":  "artifact buffer",
		"reference_spec_export_bytes": fmt.Sprintf("%d bytes", len(m.state.ReferenceSpecExportJSON)),
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
}

func TestReferenceSpecExportWritesValidatedPath(t *testing.T) {
	m := New()
	path := filepath.Join(t.TempDir(), "circuits-reference-specs.json")
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionExportReferenceSpecs,
		Payload: map[string]string{"path": path},
	}); err != nil {
		t.Fatalf("export reference specs to path: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported reference specs: %v", err)
	}
	if string(raw) != m.state.ReferenceSpecExportJSON {
		t.Fatalf("file export differs from buffered reference spec artifact")
	}
	if got := metricValue(m.Snapshot(), "reference_spec_export"); got != "wrote 64 cells" {
		t.Fatalf("reference_spec_export metric = %q, want file-write status", got)
	}
	if got := metricValue(m.Snapshot(), "reference_spec_export_path"); got != filepath.Clean(path) {
		t.Fatalf("reference_spec_export_path = %q, want cleaned path", got)
	}
}

func TestReferenceSpecExportRejectsTraversalPath(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{
		ID:      ActionExportReferenceSpecs,
		Payload: map[string]string{"path": "../escape.json"},
	})
	if err == nil {
		t.Fatal("expected traversal export path to fail")
	}
	if !contains(err.Error(), "path traversal") {
		t.Fatalf("reference spec export path error = %v, want traversal rejection", err)
	}
	if m.state.ReferenceSpecExportJSON != "" {
		t.Fatal("reference spec export artifact should not be buffered after path validation failure")
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

func TestReferenceTimingExportBuffersJSONArtifact(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionSetOperationMode,
		Payload: map[string]string{"mode": OperationCompute},
	}); err != nil {
		t.Fatalf("set compute mode: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: ActionExportReferenceTiming}); err != nil {
		t.Fatalf("export reference timing: %v", err)
	}

	if m.state.ReferenceTimingExportJSON == "" {
		t.Fatal("expected buffered reference timing export JSON")
	}
	var payload ReferenceTimingExport
	if err := json.Unmarshal([]byte(m.state.ReferenceTimingExportJSON), &payload); err != nil {
		t.Fatalf("reference timing export JSON did not unmarshal: %v\n%s", err, m.state.ReferenceTimingExportJSON)
	}
	if payload.Schema != "fecim.circuits.reference_timing.v1" {
		t.Fatalf("schema = %q, want reference-timing schema", payload.Schema)
	}
	if payload.Module != "circuits" || payload.OperationMode != OperationCompute {
		t.Fatalf("unexpected timing export context: %+v", payload)
	}
	if payload.WriteTotalNS != 203 || payload.ReadTotalNS != 76 || payload.ComputeTotalNS != 76 {
		t.Fatalf("unexpected operation totals: %+v", payload)
	}
	if payload.ActiveOperation != "COMPUTE" || payload.ActiveTotalNS != 76 || payload.ActivePhases != "DAC 10 / Array 5 / TIA+ADC 61 ns" {
		t.Fatalf("unexpected active timing context: %+v", payload)
	}
	if len(payload.Operations) != 3 {
		t.Fatalf("operations len = %d, want 3", len(payload.Operations))
	}
	wantTotals := map[string]int{"WRITE": 203, "READ": 76, "COMPUTE": 76}
	wantPhaseCounts := map[string]int{"WRITE": 4, "READ": 4, "COMPUTE": 3}
	for _, op := range payload.Operations {
		if op.TotalNS != wantTotals[op.Operation] {
			t.Fatalf("%s total = %d, want %d", op.Operation, op.TotalNS, wantTotals[op.Operation])
		}
		if len(op.Phases) != wantPhaseCounts[op.Operation] {
			t.Fatalf("%s phase count = %d, want %d", op.Operation, len(op.Phases), wantPhaseCounts[op.Operation])
		}
	}
	if payload.Operations[0].Phases[1].Name != "Pump" || payload.Operations[0].Phases[1].DurationNS != 88 {
		t.Fatalf("write phase[1] = %+v, want pump phase", payload.Operations[0].Phases[1])
	}
	if payload.Operations[1].Phases[3].Name != "ADC" || payload.Operations[1].Phases[3].DurationNS != 50 {
		t.Fatalf("read phase[3] = %+v, want ADC phase", payload.Operations[1].Phases[3])
	}
	if payload.Operations[2].Phases[2].Name != "TIA+ADC" || payload.Operations[2].Phases[2].DurationNS != 61 {
		t.Fatalf("compute phase[2] = %+v, want TIA+ADC phase", payload.Operations[2].Phases[2])
	}
	if !contains(payload.BoundaryNotice, "educational") {
		t.Fatalf("boundary notice = %q, want educational boundary", payload.BoundaryNotice)
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"reference_timing_export":       "buffered 3 operations",
		"reference_timing_export_path":  "artifact buffer",
		"reference_timing_export_bytes": fmt.Sprintf("%d bytes", len(m.state.ReferenceTimingExportJSON)),
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
}

func TestReferenceTimingExportWritesValidatedPath(t *testing.T) {
	m := New()
	path := filepath.Join(t.TempDir(), "circuits-reference-timing.json")
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionExportReferenceTiming,
		Payload: map[string]string{"path": path},
	}); err != nil {
		t.Fatalf("export reference timing to path: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported reference timing: %v", err)
	}
	if string(raw) != m.state.ReferenceTimingExportJSON {
		t.Fatalf("file export differs from buffered reference timing artifact")
	}
	if got := metricValue(m.Snapshot(), "reference_timing_export"); got != "wrote 3 operations" {
		t.Fatalf("reference_timing_export metric = %q, want file-write status", got)
	}
	if got := metricValue(m.Snapshot(), "reference_timing_export_path"); got != filepath.Clean(path) {
		t.Fatalf("reference_timing_export_path = %q, want cleaned path", got)
	}
}

func TestReferenceTimingExportRejectsTraversalPath(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{
		ID:      ActionExportReferenceTiming,
		Payload: map[string]string{"path": "../escape.json"},
	})
	if err == nil {
		t.Fatal("expected traversal timing export path to fail")
	}
	if !contains(err.Error(), "path traversal") {
		t.Fatalf("reference timing export path error = %v, want traversal rejection", err)
	}
	if m.state.ReferenceTimingExportJSON != "" {
		t.Fatal("reference timing export artifact should not be buffered after path validation failure")
	}
}

func TestOperationLogRecordsControlsAndOperations(t *testing.T) {
	m := New()
	actions := []viewmodel.Action{
		{ID: ActionSetOperationMode, Payload: map[string]string{"mode": OperationWrite}},
		{ID: ActionSelectCell, Payload: map[string]string{"row": "2", "col": "3"}},
		{ID: ActionRunWrite},
	}
	for _, action := range actions {
		if err := m.ApplyAction(action); err != nil {
			t.Fatalf("ApplyAction(%s): %v", action.ID, err)
		}
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"operation_log_count":  "3 total / 3 shown",
		"operation_log_latest": "operation #3: WRITE level 15 to cell [2,3] using Preisach (Level-based)",
		"operation_log_recent": "control #1: Operation mode set to write | control #2: Selected cell [2,3] | operation #3: WRITE level 15 to cell [2,3] using Preisach (Level-based)",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
	body := sectionBody(s, "operation_log")
	if body == "" {
		t.Fatal("missing operation log section")
	}
	for _, want := range []string{
		"control #1: Operation mode set to write",
		"control #2: Selected cell [2,3]",
		"operation #3: WRITE level 15 to cell [2,3] using Preisach (Level-based)",
	} {
		if !contains(body, want) {
			t.Fatalf("operation log section missing %q in %q", want, body)
		}
	}
}

func TestOperationLogKeepsMostRecentEntries(t *testing.T) {
	m := New()
	for level := 0; level < 15; level++ {
		if err := m.ApplyAction(viewmodel.Action{
			ID:      ActionSetWriteTarget,
			Payload: map[string]string{"level": strconv.Itoa(level % 10)},
		}); err != nil {
			t.Fatalf("set write target %d: %v", level, err)
		}
	}

	s := m.Snapshot()
	if got := metricValue(s, "operation_log_count"); got != "15 total / 12 shown" {
		t.Fatalf("operation_log_count = %q, want retained count", got)
	}
	if got := metricValue(s, "operation_log_latest"); got != "control #15: Write target set to level 4" {
		t.Fatalf("operation_log_latest = %q, want newest retained entry", got)
	}
	body := sectionBody(s, "operation_log")
	if contains(body, "control #1:") {
		t.Fatalf("operation log retained old entry: %q", body)
	}
	if !contains(body, "control #15: Write target set to level 4") {
		t.Fatalf("operation log section missing newest entry: %q", body)
	}
}

func TestOperationLogExportBuffersJSONArtifact(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: ActionRunCompute}); err != nil {
		t.Fatalf("run compute: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: ActionExportOperationLog}); err != nil {
		t.Fatalf("export operation log: %v", err)
	}

	if m.state.OperationLogExportJSON == "" {
		t.Fatal("expected buffered export JSON")
	}
	var payload struct {
		Schema            string              `json:"schema"`
		Module            string              `json:"module"`
		OperationMode     string              `json:"operation_mode"`
		Architecture      string              `json:"architecture"`
		OperationLogTotal int                 `json:"operation_log_total"`
		ExportedEntries   int                 `json:"exported_entries"`
		Entries           []OperationLogEntry `json:"entries"`
		ComputeRun        *ComputeRunLog      `json:"compute_run"`
	}
	if err := json.Unmarshal([]byte(m.state.OperationLogExportJSON), &payload); err != nil {
		t.Fatalf("export JSON did not unmarshal: %v\n%s", err, m.state.OperationLogExportJSON)
	}
	if payload.Schema != "fecim.circuits.operation_log.v1" {
		t.Fatalf("schema = %q, want operation-log schema", payload.Schema)
	}
	if payload.Module != "circuits" || payload.OperationMode != OperationCompute || payload.Architecture != ArchitecturePassive {
		t.Fatalf("unexpected export context: %+v", payload)
	}
	if payload.OperationLogTotal != 1 || payload.ExportedEntries != 1 || len(payload.Entries) != 1 {
		t.Fatalf("unexpected export counts: %+v", payload)
	}
	if payload.Entries[0].Message != "COMPUTE on 8x8 0T1R (Passive) array" {
		t.Fatalf("export entry message = %q, want compute summary", payload.Entries[0].Message)
	}
	if payload.ComputeRun == nil {
		t.Fatal("expected compute-run payload when latest operation is compute")
	}
	if payload.ComputeRun.ExportedCells != 64 {
		t.Fatalf("compute exported cells = %d, want 64", payload.ComputeRun.ExportedCells)
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"operation_log_export":       "buffered 1 entries",
		"operation_log_export_path":  "artifact buffer",
		"operation_log_export_bytes": fmt.Sprintf("%d bytes", len(m.state.OperationLogExportJSON)),
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
}

func TestComputeRunLogCapturesDetailedMVMSummary(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionResizeArray,
		Payload: map[string]string{"rows": "2", "cols": "2"},
	}); err != nil {
		t.Fatalf("resize array: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: ActionRunCompute}); err != nil {
		t.Fatalf("run compute: %v", err)
	}

	log := m.state.ComputeRunLog
	if log == nil {
		t.Fatal("missing compute-run log")
	}
	if log.Schema != "fecim.circuits.compute_run.v1" {
		t.Fatalf("schema = %q, want compute-run schema", log.Schema)
	}
	if log.ArraySize != "2x2" || log.QuantLevels != DefaultQuantLevels || log.ExportedCells != 4 {
		t.Fatalf("unexpected compute-run header: %+v", log)
	}
	wantInput := []float64{0.2, 0.5}
	if len(log.InputVector) != len(wantInput) {
		t.Fatalf("input vector len = %d, want %d", len(log.InputVector), len(wantInput))
	}
	for i, want := range wantInput {
		if !near(log.InputVector[i], want, 1e-9) {
			t.Fatalf("input[%d] = %.9f, want %.9f", i, log.InputVector[i], want)
		}
	}
	wantWeights := [][]int{{0, 1}, {2, 3}}
	for r := range wantWeights {
		for c := range wantWeights[r] {
			if log.Weights[r][c] != wantWeights[r][c] {
				t.Fatalf("weight[%d][%d] = %d, want %d", r, c, log.Weights[r][c], wantWeights[r][c])
			}
		}
	}
	if len(log.RowResults) != 2 {
		t.Fatalf("row result count = %d, want 2", len(log.RowResults))
	}
	row0 := log.RowResults[0]
	if row0.Row != 0 || !row0.Active || row0.Saturated || row0.ADCLevel != 0 {
		t.Fatalf("unexpected row0 flags: %+v", row0)
	}
	if !near(row0.CurrentUA, 2.4068965517, 1e-9) {
		t.Fatalf("row0 current = %.10f, want 2.4068965517", row0.CurrentUA)
	}
	if !near(row0.TIAVoltage, 0.0240689655, 1e-9) {
		t.Fatalf("row0 TIA voltage = %.10f, want 0.0240689655", row0.TIAVoltage)
	}
	if len(row0.CellDetail) != 2 {
		t.Fatalf("row0 cell detail count = %d, want 2", len(row0.CellDetail))
	}
	cell := row0.CellDetail[1]
	if cell.Col != 1 || cell.Weight != 1 {
		t.Fatalf("unexpected row0 cell1 identity: %+v", cell)
	}
	if !near(cell.ConductanceUS, 4.4137931034, 1e-9) || !near(cell.VoltageV, 0.5, 1e-9) || !near(cell.CurrentUA, 2.2068965517, 1e-9) {
		t.Fatalf("unexpected row0 cell1 analog terms: %+v", cell)
	}
	row1 := log.RowResults[1]
	if row1.ADCLevel != 1 || row1.Saturated {
		t.Fatalf("unexpected row1 ADC/saturation: %+v", row1)
	}
	if !near(row1.CurrentUA, 7.1862068966, 1e-9) {
		t.Fatalf("row1 current = %.10f, want 7.1862068966", row1.CurrentUA)
	}

	s := m.Snapshot()
	wantMetrics := map[string]string{
		"compute_run":       "2x2 / 2 rows / 4 cells",
		"compute_run_peak":  "7.186 uA row 1",
		"compute_run_input": "0.200..0.500 V",
	}
	for id, want := range wantMetrics {
		if got := metricValue(s, id); got != want {
			t.Errorf("%s metric = %q, want %q", id, got, want)
		}
	}
	if !contains(sectionBody(s, "compute_run_log"), "Rows: 2. Cells: 4. Peak current: 7.186 uA at row 1.") {
		t.Fatalf("compute_run_log section missing summary: %q", sectionBody(s, "compute_run_log"))
	}
}

func TestOperationLogExportOmitsComputeRunWhenLatestOperationIsNotCompute(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: ActionRunCompute}); err != nil {
		t.Fatalf("run compute: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: ActionRunRead}); err != nil {
		t.Fatalf("run read: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: ActionExportOperationLog}); err != nil {
		t.Fatalf("export operation log: %v", err)
	}

	var payload OperationLogExport
	if err := json.Unmarshal([]byte(m.state.OperationLogExportJSON), &payload); err != nil {
		t.Fatalf("unmarshal export: %v", err)
	}
	if payload.ComputeRun != nil {
		t.Fatalf("compute-run payload exported for non-compute latest operation: %+v", payload.ComputeRun)
	}
}

func TestOperationLogExportWritesValidatedPath(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: ActionRunRead}); err != nil {
		t.Fatalf("run read: %v", err)
	}
	path := filepath.Join(t.TempDir(), "circuits-operation-log.json")
	if err := m.ApplyAction(viewmodel.Action{
		ID:      ActionExportOperationLog,
		Payload: map[string]string{"path": path},
	}); err != nil {
		t.Fatalf("export operation log to path: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	if string(raw) != m.state.OperationLogExportJSON {
		t.Fatalf("file export differs from buffered artifact")
	}
	if got := metricValue(m.Snapshot(), "operation_log_export"); got != "wrote 1 entries" {
		t.Fatalf("operation_log_export metric = %q, want file-write status", got)
	}
	if got := metricValue(m.Snapshot(), "operation_log_export_path"); got != filepath.Clean(path) {
		t.Fatalf("operation_log_export_path = %q, want cleaned path", got)
	}
}

func TestOperationLogExportRejectsTraversalPath(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: ActionRunRead}); err != nil {
		t.Fatalf("run read: %v", err)
	}
	err := m.ApplyAction(viewmodel.Action{
		ID:      ActionExportOperationLog,
		Payload: map[string]string{"path": "../escape.json"},
	})
	if err == nil {
		t.Fatal("expected traversal export path to fail")
	}
	if !contains(err.Error(), "path traversal") {
		t.Fatalf("export path error = %v, want traversal rejection", err)
	}
	if m.state.OperationLogExportJSON != "" {
		t.Fatal("export artifact should not be buffered after path validation failure")
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

func sectionBody(s viewmodel.ModuleSnapshot, id string) string {
	for _, section := range s.Sections {
		if section.ID == id {
			return section.Body
		}
	}
	return ""
}

func hasAction(s viewmodel.ModuleSnapshot, id string) bool {
	for _, action := range s.Actions {
		if action.ID == id {
			return true
		}
	}
	return false
}

func contains(value, needle string) bool {
	return strings.Contains(value, needle)
}

func near(got, want, tolerance float64) bool {
	if got > want {
		return got-want <= tolerance
	}
	return want-got <= tolerance
}
