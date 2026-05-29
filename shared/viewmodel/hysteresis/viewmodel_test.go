package hysteresis

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	m := New()
	d := m.Descriptor()
	if d.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Descriptor().ID = %v, want %v", d.ID, viewmodel.ModuleHysteresis)
	}
	if d.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor().Status = %v, want %v", d.Status, viewmodel.StatusFunctional)
	}
}

func TestSnapshotContainsMetrics(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Metrics) == 0 {
		t.Error("Snapshot().Metrics is empty, expected material metrics")
	}
	if s.Descriptor.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Snapshot().Descriptor.ID = %v", s.Descriptor.ID)
	}
}

func TestSnapshotHasSections(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Sections) == 0 {
		t.Error("Snapshot().Sections is empty")
	}
}

func TestApplyActionUnsupported(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{ID: "unknown", Kind: viewmodel.ActionCommand})
	if err != viewmodel.ErrUnsupportedAction {
		t.Errorf("ApplyAction error = %v, want %v", err, viewmodel.ErrUnsupportedAction)
	}
}

func TestApplyActionSelectMaterial(t *testing.T) {
	m := New()
	matName := m.Descriptor().Title
	_ = matName
	s := m.Snapshot()
	var firstMaterial string
	for _, m := range s.Metrics {
		if m.ID == "material" {
			firstMaterial = m.Value
			break
		}
	}
	if firstMaterial == "" {
		t.Fatal("no initial material found")
	}
	err := m.ApplyAction(viewmodel.Action{
		ID:      EventSelectMaterial,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"material": firstMaterial},
	})
	if err != nil {
		t.Fatalf("ApplyAction select_material(%q): %v", firstMaterial, err)
	}
}

func TestApplyActionToggleSimulation(t *testing.T) {
	m := New()
	m.ApplyAction(viewmodel.Action{ID: EventToggleSimulation, Kind: viewmodel.ActionToggle})
	m.ApplyAction(viewmodel.Action{ID: EventToggleSimulation, Kind: viewmodel.ActionToggle})
}

func TestSnapshotExposesWaveformAndCSVExportActions(t *testing.T) {
	m := New()
	s := m.Snapshot()
	actions := snapshotActionsByID(s)

	if _, ok := actions[EventSetWaveform]; !ok {
		t.Fatalf("snapshot actions missing %q", EventSetWaveform)
	}
	if _, ok := actions[EventExportCSV]; !ok {
		t.Fatalf("snapshot actions missing %q", EventExportCSV)
	}
}

func TestSnapshotExposesPUNDAndFORCActions(t *testing.T) {
	m := New()
	actions := snapshotActionsByID(m.Snapshot())

	if _, ok := actions["run_pund"]; !ok {
		t.Fatalf("snapshot actions missing run_pund")
	}
	if _, ok := actions["run_forc"]; !ok {
		t.Fatalf("snapshot actions missing run_forc")
	}
	for _, id := range []string{"export_pund_csv", "export_forc_sweep_csv", "export_forc_matrix_csv", "export_forc_metadata_json"} {
		if _, ok := actions[id]; !ok {
			t.Fatalf("snapshot actions missing %s", id)
		}
	}
}

func TestApplyActionRunLevelCalibrationPublishesFreshSummary(t *testing.T) {
	m := New()
	before := m.Snapshot()
	if got := snapshotMetricValue(before, "level_calibration_state"); got != "not calibrated" {
		t.Fatalf("initial level_calibration_state = %q, want not calibrated", got)
	}
	actions := snapshotActionsByID(before)
	if _, ok := actions[EventRunLevelCalibration]; !ok {
		t.Fatalf("snapshot actions missing %q", EventRunLevelCalibration)
	}
	if _, ok := actions[EventExportLevelCalibration]; !ok {
		t.Fatalf("snapshot actions missing %q", EventExportLevelCalibration)
	}

	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction run level calibration: %v", err)
	}

	after := m.Snapshot()
	if got := snapshotMetricValue(after, "level_calibration_state"); got != "fresh" {
		t.Fatalf("level_calibration_state after run = %q, want fresh", got)
	}
	if got := snapshotMetricValue(after, "level_calibration_method"); !strings.Contains(got, "LK quasi-static") {
		t.Fatalf("level_calibration_method = %q, want LK quasi-static method", got)
	}
	if got := snapshotMetricValue(after, "level_calibration_entries"); !strings.Contains(got, "ascending") || !strings.Contains(got, "descending") {
		t.Fatalf("level_calibration_entries = %q, want ascending/descending entry counts", got)
	}
	if got := snapshotMetricValue(after, "level_calibration_monotonicity"); got != "ascending/descending monotonic" {
		t.Fatalf("level_calibration_monotonicity = %q, want ascending/descending monotonic", got)
	}
	section := snapshotSectionBody(after, "level_calibration_summary")
	for _, want := range []string{"LK quasi-static", "Target range", "SIMULATION OUTPUT", "not measured"} {
		if !strings.Contains(section, want) {
			t.Fatalf("level calibration summary missing %q: %q", want, section)
		}
	}
}

func TestSnapshotLevelCalibrationDetailIncludesRepresentativeRows(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}

	detail := snapshotSectionBody(m.Snapshot(), "level_calibration_detail")
	for _, want := range []string{"Representative rows", "level 0", "level 15", "level 29", "ascending", "descending", "not measured"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("level calibration detail missing %q: %q", want, detail)
		}
	}
}

func TestApplyActionExportLevelCalibrationBuffersFreshJSONArtifact(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}

	if err := m.ApplyAction(viewmodel.Action{ID: EventExportLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("export level calibration: %v", err)
	}

	snapshot := m.Snapshot()
	if got := snapshotMetricValue(snapshot, "level_calibration_export_path"); got != "artifact buffer" {
		t.Fatalf("level_calibration_export_path = %q, want artifact buffer", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_export"); !strings.Contains(got, "buffered") {
		t.Fatalf("level_calibration_export = %q, want buffered status", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_export_bytes"); got == "" || got == "0 bytes" {
		t.Fatalf("level_calibration_export_bytes = %q, want non-zero bytes", got)
	}

	content := levelCalibrationExportContent(t, m)
	var payload map[string]any
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		t.Fatalf("level calibration export JSON did not unmarshal: %v\n%s", err, content)
	}
	if got := payload["artifact_type"]; got != "level_calibration" {
		t.Fatalf("artifact_type = %#v, want level_calibration", got)
	}
	if got := payload["status"]; got != "fresh" {
		t.Fatalf("status = %#v, want fresh", got)
	}
	if got := payload["boundary_notice"]; !strings.Contains(got.(string), "not measured") {
		t.Fatalf("boundary_notice = %#v, want simulation boundary", got)
	}
	levels, ok := payload["levels"].([]any)
	if !ok || len(levels) != 30 {
		t.Fatalf("levels length = %d ok=%v, want 30", len(levels), ok)
	}
	first, ok := levels[0].(map[string]any)
	if !ok {
		t.Fatalf("first level row type = %T, want object", levels[0])
	}
	for _, key := range []string{"level_index", "normalized_target", "ascending_field_vm", "descending_field_vm"} {
		if _, ok := first[key]; !ok {
			t.Fatalf("first level row missing %q: %#v", key, first)
		}
	}
}

func TestApplyActionExportLevelCalibrationRequiresFreshResult(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventExportLevelCalibration, Kind: viewmodel.ActionCommand}); err == nil || !strings.Contains(err.Error(), "run level calibration") {
		t.Fatalf("export before calibration error = %v, want run-first error", err)
	}

	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: EventSetLevelCalibrationLevelCount, Kind: viewmodel.ActionCommand, Payload: map[string]string{"level_count": "16"}}); err != nil {
		t.Fatalf("set level count: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: EventExportLevelCalibration, Kind: viewmodel.ActionCommand}); err == nil || !strings.Contains(err.Error(), "rerun level calibration") {
		t.Fatalf("export stale calibration error = %v, want rerun error", err)
	}
	if got := snapshotMetricValue(m.Snapshot(), "level_calibration_export"); got != "" {
		t.Fatalf("level_calibration_export after rejected stale export = %q, want empty", got)
	}
}

func TestApplyActionExportLevelCalibrationWritesValidatedPath(t *testing.T) {
	m := New()
	path := filepath.Join(t.TempDir(), "level-calibration.json")
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}
	if err := m.ApplyAction(viewmodel.Action{ID: EventExportLevelCalibration, Kind: viewmodel.ActionCommand, Payload: map[string]string{"path": path}}); err != nil {
		t.Fatalf("export level calibration path: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read level calibration export: %v", err)
	}
	if string(raw) != m.state.LevelCalibration.ExportContent {
		t.Fatalf("file content does not match buffered export content")
	}
	if got := snapshotMetricValue(m.Snapshot(), "level_calibration_export_path"); got != path {
		t.Fatalf("level_calibration_export_path = %q, want %q", got, path)
	}
}

func TestApplyActionSetLevelCalibrationLevelCountMarksFreshResultStale(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}

	if err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetLevelCalibrationLevelCount,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"level_count": "16"},
	}); err != nil {
		t.Fatalf("set level calibration level count: %v", err)
	}

	snapshot := m.Snapshot()
	if got := snapshotMetricValue(snapshot, "level_calibration_state"); got != "stale" {
		t.Fatalf("level_calibration_state after level count change = %q, want stale", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_inputs"); !strings.Contains(got, "16 levels") {
		t.Fatalf("level_calibration_inputs = %q, want 16 levels", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_entries"); !strings.Contains(got, "30 ascending") {
		t.Fatalf("stale result entries = %q, want prior 30-level summary preserved", got)
	}
}

func TestApplyActionSetLevelCalibrationTargetRangeMarksFreshResultStale(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}

	if err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetLevelCalibrationTargetRange,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"target_range": "0.70"},
	}); err != nil {
		t.Fatalf("set level calibration target range: %v", err)
	}

	snapshot := m.Snapshot()
	if got := snapshotMetricValue(snapshot, "level_calibration_state"); got != "stale" {
		t.Fatalf("level_calibration_state after target range change = %q, want stale", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_inputs"); !strings.Contains(got, "target 70%") {
		t.Fatalf("level_calibration_inputs = %q, want target 70%%", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_entries"); !strings.Contains(got, "30 ascending") {
		t.Fatalf("stale result entries = %q, want prior summary preserved", got)
	}
}

func TestApplyActionSetLevelCalibrationTemperatureMarksFreshResultStale(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}

	if err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetLevelCalibrationTemperature,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"temperature_k": "350"},
	}); err != nil {
		t.Fatalf("set level calibration temperature: %v", err)
	}

	snapshot := m.Snapshot()
	if got := snapshotMetricValue(snapshot, "level_calibration_state"); got != "stale" {
		t.Fatalf("level_calibration_state after temperature change = %q, want stale", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_inputs"); !strings.Contains(got, "350 K") {
		t.Fatalf("level_calibration_inputs = %q, want 350 K", got)
	}
	if got := snapshotMetricValue(snapshot, "level_calibration_entries"); !strings.Contains(got, "30 ascending") {
		t.Fatalf("stale result entries = %q, want prior summary preserved", got)
	}
}

func TestApplyActionSelectMaterialResetsCalibrationDefaultsAndPreservesTemperature(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}
	for _, action := range []viewmodel.Action{
		{ID: EventSetLevelCalibrationLevelCount, Kind: viewmodel.ActionCommand, Payload: map[string]string{"level_count": "16"}},
		{ID: EventSetLevelCalibrationTargetRange, Kind: viewmodel.ActionCommand, Payload: map[string]string{"target_range": "0.70"}},
		{ID: EventSetLevelCalibrationTemperature, Kind: viewmodel.ActionCommand, Payload: map[string]string{"temperature_k": "350"}},
	} {
		if err := m.ApplyAction(action); err != nil {
			t.Fatalf("ApplyAction(%s): %v", action.ID, err)
		}
	}

	newMaterial := secondSnapshotMaterialName(t, m.Snapshot(), snapshotMetricValue(m.Snapshot(), "material"))
	if err := m.ApplyAction(viewmodel.Action{
		ID:      EventSelectMaterial,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"material": newMaterial},
	}); err != nil {
		t.Fatalf("select material %q: %v", newMaterial, err)
	}

	snapshot := m.Snapshot()
	if got := snapshotMetricValue(snapshot, "level_calibration_state"); got != "stale" {
		t.Fatalf("level_calibration_state after material change = %q, want stale because prior result no longer matches", got)
	}
	inputs := snapshotMetricValue(snapshot, "level_calibration_inputs")
	if !strings.Contains(inputs, "350 K") {
		t.Fatalf("level_calibration_inputs = %q, want preserved 350 K", inputs)
	}
	if strings.Contains(inputs, "16 levels") || strings.Contains(inputs, "target 70%") {
		t.Fatalf("level_calibration_inputs = %q, want material-default level count and target range", inputs)
	}
}

func TestApplyActionSetWaveformUpdatesSnapshotAndLoop(t *testing.T) {
	m := New()
	before := m.Snapshot()
	beforeFirst := before.Plots[0].Series[0].Points[0]

	err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetWaveform,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"waveform": "triangle"},
	})
	if err != nil {
		t.Fatalf("ApplyAction set_waveform: %v", err)
	}

	after := m.Snapshot()
	if got := snapshotMetricValue(after, "waveform"); got != "triangle" {
		t.Fatalf("waveform metric = %q, want triangle", got)
	}
	afterFirst := after.Plots[0].Series[0].Points[0]
	if afterFirst == beforeFirst {
		t.Fatal("first P-E loop point did not change after selecting triangle waveform")
	}
}

func TestApplyActionSetFieldRangeRejectsInvalidFloatPayload(t *testing.T) {
	m := New()
	before := m.state.FieldRange

	err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetFieldRange,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"min": "low", "max": "2.5"},
	})
	if err == nil {
		t.Fatal("expected invalid field range payload to fail")
	}
	if m.state.FieldRange != before {
		t.Fatalf("field range changed after invalid payload: %+v, want %+v", m.state.FieldRange, before)
	}
}

func TestApplyActionSetWaveformRejectsUnknownWaveform(t *testing.T) {
	m := New()

	err := m.ApplyAction(viewmodel.Action{
		ID:      EventSetWaveform,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"waveform": "sawtooth"},
	})
	if err == nil {
		t.Fatal("ApplyAction accepted unknown waveform")
	}
	if got := snapshotMetricValue(m.Snapshot(), "waveform"); got != "sine" {
		t.Fatalf("waveform metric after invalid action = %q, want sine", got)
	}
}

func TestApplyActionExportCSVBuffersLoopData(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: EventExportCSV, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction export_csv: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "csv_export_path"); got != "artifact buffer" {
		t.Fatalf("csv_export_path = %q, want artifact buffer", got)
	}
	if got := snapshotMetricValue(s, "csv_export"); !strings.Contains(got, "buffered 200 points") {
		t.Fatalf("csv_export = %q, want buffered point count", got)
	}
	if got := snapshotMetricValue(s, "csv_export_bytes"); got == "" || got == "0 bytes" {
		t.Fatalf("csv_export_bytes = %q, want non-zero bytes", got)
	}
}

func TestApplyActionExportCSVWritesValidatedPath(t *testing.T) {
	m := New()
	path := filepath.Join(t.TempDir(), "exports", "pe-loop.csv")

	if err := m.ApplyAction(viewmodel.Action{
		ID:      EventExportCSV,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"path": path},
	}); err != nil {
		t.Fatalf("ApplyAction export_csv path: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported CSV: %v", err)
	}
	text := string(content)
	if !strings.HasPrefix(text, "Index,E_field_kV_cm,Polarization_uC_cm2\n") {
		t.Fatalf("CSV header = %q", strings.SplitN(text, "\n", 2)[0])
	}
	if got := strings.Count(text, "\n"); got != 201 {
		t.Fatalf("CSV newline count = %d, want 201", got)
	}
	if got := snapshotMetricValue(m.Snapshot(), "csv_export_path"); got != path {
		t.Fatalf("csv_export_path = %q, want %q", got, path)
	}
}

func TestApplyActionExportCSVRejectsTraversalPath(t *testing.T) {
	m := New()

	err := m.ApplyAction(viewmodel.Action{
		ID:      EventExportCSV,
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"path": "../escape.csv"},
	})
	if err == nil {
		t.Fatal("ApplyAction accepted traversal export path")
	}
}

func TestApplyActionRunPUNDPopulatesDiagnosticSummary(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: "run_pund", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction run_pund: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "pund_switching_positive"); got == "" || got == "0.000e+00 C" {
		t.Fatalf("pund_switching_positive = %q, want non-zero charge", got)
	}
	if got := snapshotMetricValue(s, "pund_switching_negative"); got == "" || got == "0.000e+00 C" {
		t.Fatalf("pund_switching_negative = %q, want non-zero charge", got)
	}
	if got := snapshotMetricValue(s, "pund_switching_ratio"); got == "" {
		t.Fatal("missing pund_switching_ratio metric")
	}
	section := snapshotSectionBody(s, "diagnostic_pund")
	if !strings.Contains(section, "QP=") || !strings.Contains(section, "Switching ratio") {
		t.Fatalf("PUND diagnostic section = %q, want charge and ratio summary", section)
	}
}

func TestApplyActionRunPUNDExposesCurrentWaveformPlot(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: "run_pund", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction run_pund: %v", err)
	}

	plot := snapshotPlotByID(m.Snapshot(), "pund_current_waveforms")
	if plot.ID == "" {
		t.Fatal("missing pund_current_waveforms plot")
	}
	if plot.XLabel != "Time (ns)" || plot.YLabel != "Current (A)" {
		t.Fatalf("PUND plot labels = %q/%q, want Time (ns)/Current (A)", plot.XLabel, plot.YLabel)
	}
	if len(plot.Series) != 4 {
		t.Fatalf("PUND plot series count = %d, want P/U/N/D", len(plot.Series))
	}
	for _, wantName := range []string{"P", "U", "N", "D"} {
		series := plotSeriesByName(plot, wantName)
		if len(series.Points) < 2 {
			t.Fatalf("PUND %s series point count = %d, want at least 2", wantName, len(series.Points))
		}
		if series.Points[1].X <= series.Points[0].X {
			t.Fatalf("PUND %s series time did not increase: %+v", wantName, series.Points[:2])
		}
	}
}

func TestApplyActionRunFORCPopulatesDensitySummary(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{
		ID:      "run_forc",
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"reversals": "13"},
	}); err != nil {
		t.Fatalf("ApplyAction run_forc: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "forc_curves"); got != "13" {
		t.Fatalf("forc_curves = %q, want 13", got)
	}
	if got := snapshotMetricValue(s, "forc_density_peak"); got == "" || got == "0.000e+00" {
		t.Fatalf("forc_density_peak = %q, want non-zero density peak", got)
	}
	section := snapshotSectionBody(s, "diagnostic_forc")
	if !strings.Contains(section, "peak_density=") || !strings.Contains(section, "density_range=") {
		t.Fatalf("FORC diagnostic section = %q, want density summary", section)
	}
}

func TestApplyActionRunFORCExposesDensityHeatmapPlot(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{
		ID:      "run_forc",
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"reversals": "13"},
	}); err != nil {
		t.Fatalf("ApplyAction run_forc: %v", err)
	}

	plot := snapshotPlotByID(m.Snapshot(), "forc_density_heatmap")
	if plot.ID == "" {
		t.Fatal("missing forc_density_heatmap plot")
	}
	if plot.XLabel != "Ea (V/m)" || plot.YLabel != "Eb (V/m)" {
		t.Fatalf("FORC plot labels = %q/%q, want Ea (V/m)/Eb (V/m)", plot.XLabel, plot.YLabel)
	}
	if len(plot.Series) != 1 {
		t.Fatalf("FORC plot series count = %d, want one density series", len(plot.Series))
	}
	points := plot.Series[0].Points
	if len(points) != 13*13 {
		t.Fatalf("FORC density point count = %d, want 169", len(points))
	}
	foundDensity := false
	for _, point := range points {
		if point.V != 0 {
			foundDensity = true
			break
		}
	}
	if !foundDensity {
		t.Fatal("FORC density heatmap plot has no non-zero density values")
	}
}

func TestApplyActionExportPUNDCSVBuffersSummary(t *testing.T) {
	m := New()

	if err := m.ApplyAction(viewmodel.Action{ID: "export_pund_csv", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("ApplyAction export_pund_csv: %v", err)
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "pund_export_path"); got != "artifact buffer" {
		t.Fatalf("pund_export_path = %q, want artifact buffer", got)
	}
	if got := snapshotMetricValue(s, "pund_export"); !strings.Contains(got, "buffered") {
		t.Fatalf("pund_export = %q, want buffered status", got)
	}
	if got := snapshotMetricValue(s, "pund_export_bytes"); got == "" || got == "0 bytes" {
		t.Fatalf("pund_export_bytes = %q, want non-zero bytes", got)
	}
}

func TestApplyActionExportFORCArtifactsWritesValidatedPaths(t *testing.T) {
	m := New()
	dir := t.TempDir()
	sweepPath := filepath.Join(dir, "forc-sweep.csv")
	matrixPath := filepath.Join(dir, "forc-matrix.csv")
	metaPath := filepath.Join(dir, "forc-meta.json")

	for _, tc := range []struct {
		id     string
		path   string
		header string
	}{
		{"export_forc_sweep_csv", sweepPath, "reversal_field_vm,applied_field_vm,polarization_cm2\n"},
		{"export_forc_matrix_csv", matrixPath, "Ea_Vm,Eb_Vm,density\n"},
		{"export_forc_metadata_json", metaPath, "{\n"},
	} {
		if err := m.ApplyAction(viewmodel.Action{
			ID:      tc.id,
			Kind:    viewmodel.ActionCommand,
			Payload: map[string]string{"path": tc.path, "reversals": "9"},
		}); err != nil {
			t.Fatalf("ApplyAction %s: %v", tc.id, err)
		}
		content, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.path, err)
		}
		if !strings.HasPrefix(string(content), tc.header) {
			t.Fatalf("%s header = %q, want prefix %q", tc.id, strings.SplitN(string(content), "\n", 2)[0], strings.TrimSpace(tc.header))
		}
	}

	s := m.Snapshot()
	if got := snapshotMetricValue(s, "forc_sweep_export_path"); got != sweepPath {
		t.Fatalf("forc_sweep_export_path = %q, want %q", got, sweepPath)
	}
	if got := snapshotMetricValue(s, "forc_matrix_export_path"); got != matrixPath {
		t.Fatalf("forc_matrix_export_path = %q, want %q", got, matrixPath)
	}
	if got := snapshotMetricValue(s, "forc_metadata_export_path"); got != metaPath {
		t.Fatalf("forc_metadata_export_path = %q, want %q", got, metaPath)
	}
}

func TestMaterialSummary(t *testing.T) {
	summary := materialSummary(nil)
	if summary != "N/A" {
		t.Errorf("materialSummary(nil) = %q, want N/A", summary)
	}
}

func TestStateDefaults(t *testing.T) {
	m := New()
	s := m.Snapshot()
	var matMetric *viewmodel.Metric
	for i := range s.Metrics {
		if s.Metrics[i].ID == "material" {
			matMetric = &s.Metrics[i]
			break
		}
	}
	if matMetric == nil {
		t.Fatal("no material metric found")
	}
	if matMetric.Value == "" {
		t.Error("material metric value is empty")
	}
}

func snapshotActionsByID(snapshot viewmodel.ModuleSnapshot) map[string]viewmodel.Action {
	actions := make(map[string]viewmodel.Action, len(snapshot.Actions))
	for _, action := range snapshot.Actions {
		actions[action.ID] = action
	}
	return actions
}

func secondSnapshotMaterialName(t *testing.T, snapshot viewmodel.ModuleSnapshot, current string) string {
	t.Helper()
	for _, section := range snapshot.Sections {
		if section.Category == "research" && strings.HasPrefix(section.ID, "material_") && section.Title != current {
			return section.Title
		}
	}
	t.Fatalf("no second material found in snapshot sections; current=%q", current)
	return ""
}

func levelCalibrationExportContent(t *testing.T, m *Module) string {
	t.Helper()
	if m.state.LevelCalibration.ExportContent == "" {
		t.Fatal("level calibration export content is empty")
	}
	return m.state.LevelCalibration.ExportContent
}

func snapshotMetricValue(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range snapshot.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}

func snapshotSectionBody(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, section := range snapshot.Sections {
		if section.ID == id {
			return section.Body
		}
	}
	return ""
}

func snapshotPlotByID(snapshot viewmodel.ModuleSnapshot, id string) viewmodel.PlotData {
	for _, plot := range snapshot.Plots {
		if plot.ID == id {
			return plot
		}
	}
	return viewmodel.PlotData{}
}

func plotSeriesByName(plot viewmodel.PlotData, name string) viewmodel.PlotSeries {
	for _, series := range plot.Series {
		if series.Name == name {
			return series
		}
	}
	return viewmodel.PlotSeries{}
}
